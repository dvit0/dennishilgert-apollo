package microvm

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/microvm/machine"
	"github.com/dennishilgert/apollo/internal/app/fleet/microvm/pool"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"github.com/dennishilgert/apollo/pkg/proto/manager/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.operator")

type Options struct {
	OsArch                utils.OsArch
	FirecrackerBinaryPath string
	WatchdogCheckInterval time.Duration
	WatchdogWorkerCount   int
	AgentApiPort          int
}

type Operator interface {
	Init(ctx context.Context) error
	ExecuteFunction(ctx context.Context, request *manager.ExecuteFunctionRequest) (*manager.ExecuteFunctionResponse, error)
}

type vmOperator struct {
	vmPool                pool.Pool
	vmPoolWatchdog        pool.Watchdog
	osArch                utils.OsArch
	firecrackerBinaryPath string
	agentApiPort          int
}

// NewVmOperator creates a new Operator.
func NewVmOperator(opts Options) (Operator, error) {
	vmPool := pool.NewVmPool()
	vmPoolWatchdog := pool.NewVmPoolWatchdog(vmPool, pool.WatchdogOptions{
		CheckInterval: opts.WatchdogCheckInterval,
		WorkerCount:   opts.WatchdogWorkerCount,
	})

	return &vmOperator{
		vmPool:                vmPool,
		vmPoolWatchdog:        vmPoolWatchdog,
		osArch:                opts.OsArch,
		firecrackerBinaryPath: opts.FirecrackerBinaryPath,
		agentApiPort:          opts.AgentApiPort,
	}, nil
}

// Init initializes the vm operator.
func (v *vmOperator) Init(ctx context.Context) error {
	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting vm pool watchdog")
			if err := v.vmPoolWatchdog.Run(ctx); err != nil {
				log.Errorf("error while running pool watchdog: %v", err)
				return err
			}
			return nil
		},
	)
	return runner.Run(ctx)
}

func (v *vmOperator) ExecuteFunction(ctx context.Context, request *manager.ExecuteFunctionRequest) (*manager.ExecuteFunctionResponse, error) {
	multiThreading := true
	if v.osArch != utils.Arch_x86_64 {
		multiThreading = false
	}
	cfg := &machine.Config{
		FnId:                  request.Function.Id,
		HostOsArch:            v.osArch,
		FirecrackerBinaryPath: v.firecrackerBinaryPath,
		KernelImagePath:       request.Function.KernelImagePath,
		RootDrivePath:         request.Function.RootDrivePath,
		CodeDrivePath:         request.Function.CodeDrivePath,
		VCpuCount:             int(request.Function.VcpuCores),
		MemSizeMib:            int(request.Function.MemoryLimit),
		Multithreading:        multiThreading,
		AgentApiPort:          v.agentApiPort,
	}
	// get available firecracker machine from pool or start a new one
	instance, err := v.getOrStartVm(ctx, cfg)
	if err != nil {
		return nil, err
	}
	invokeRequest := &agent.InvokeRequest{
		Config: &agent.ConfigData{
			RuntimeBinaryPath: request.Runtime.BinaryPath,
			RuntimeBinaryArgs: request.Runtime.BinaryArgs,
		},
		Context: &agent.ContextData{
			Runtime:        request.Runtime.Name,
			RuntimeVersion: request.Runtime.Version,
			RuntimeHandler: request.Runtime.Handler,
			VCpuCores:      request.Function.VcpuCores,
			MemoryLimit:    request.Function.MemoryLimit,
		},
		Event: &agent.EventData{
			Id:   request.Event.Id,
			Type: request.Event.Type,
			Data: request.Event.Data,
		},
	}
	// invoke the function code with the data of the event
	invokeResponse, err := instance.Invoke(ctx, invokeRequest)
	if err != nil {
		return nil, err
	}
	// TODO: send billed duration to user service via messaging
	log.Infof("Billed duration for execution: %s, duration: %s", invokeResponse.EventId, invokeResponse.Duration)
	// TODO: send logs to logs service via messaging
	log.Infof("Logs for execution: %s, logs: %v", invokeResponse.EventId, invokeResponse.Logs)

	executeResponse := &manager.ExecuteFunctionResponse{
		EventId:       invokeResponse.EventId,
		Status:        invokeResponse.Status,
		StatusMessage: invokeResponse.StatusMessage,
		Data:          invokeResponse.Data,
	}
	return executeResponse, nil
}

// getOrStart returns a available vm from the pool or starts a new one.
func (v *vmOperator) getOrStartVm(ctx context.Context, cfg *machine.Config) (*machine.MachineInstance, error) {
	instance, err := v.vmPool.AvailableVm(cfg.FnId)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		log.Debugf("no instance available for function id: %s, reason: %v", cfg.FnId, err)
		instance := machine.NewInstance(ctx, cfg)
		if err := instance.CreateAndStart(ctx); err != nil {
			log.Errorf("failed to create and start new firecracker machine: %s", instance.Config().VmId)
			return nil, err
		}
	}
	// update machine status in pool
	instance.SetState(machine.VmStateBusy)
	return instance, nil
}
