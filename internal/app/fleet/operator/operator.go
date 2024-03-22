package operator

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/pool"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	taskRunner "github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.operator")

type Options struct {
	AgentApiPort          int
	OsArch                utils.OsArch
	FirecrackerBinaryPath string
	WatchdogCheckInterval time.Duration
	WatchdogWorkerCount   int
}

type RunnerOperator interface {
	Init(ctx context.Context) error
	InvokeFunction(ctx context.Context, request *fleet.InvokeFunctionRequest) (*fleet.InvokeFunctionResponse, error)
}

type runnerOperator struct {
	runnerPool            pool.RunnerPool
	runnerPoolWatchdog    pool.RunnerPoolWatchdog
	osArch                utils.OsArch
	firecrackerBinaryPath string
	agentApiPort          int
}

// NewRunnerOperator creates a new Operator.
func NewRunnerOperator(opts Options) (RunnerOperator, error) {
	runnerPool := pool.NewRunnerPool()
	runnerPoolWatchdog := pool.NewRunnerPoolWatchdog(runnerPool, pool.WatchdogOptions{
		CheckInterval: opts.WatchdogCheckInterval,
		WorkerCount:   opts.WatchdogWorkerCount,
	})

	return &runnerOperator{
		runnerPool:            runnerPool,
		runnerPoolWatchdog:    runnerPoolWatchdog,
		osArch:                opts.OsArch,
		firecrackerBinaryPath: opts.FirecrackerBinaryPath,
		agentApiPort:          opts.AgentApiPort,
	}, nil
}

// Init initializes the vm operator.
func (v *runnerOperator) Init(ctx context.Context) error {
	runnerManager := taskRunner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting runner pool watchdog")
			if err := v.runnerPoolWatchdog.Run(ctx); err != nil {
				log.Errorf("error while starting runner pool watchdog: %v", err)
				return err
			}
			return nil
		},
	)
	return runnerManager.Run(ctx)
}

func (v *runnerOperator) InvokeFunction(ctx context.Context, request *fleet.InvokeFunctionRequest) (*fleet.InvokeFunctionResponse, error) {
	multiThreading := false
	if v.osArch == utils.Arch_x86_64 {
		multiThreading = true
	}
	cfg := &runner.Config{
		FunctionUuid:          request.Function.Id,
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

	executeResponse := &fleet.InvokeFunctionResponse{
		EventId:       invokeResponse.EventId,
		Status:        invokeResponse.Status,
		StatusMessage: invokeResponse.StatusMessage,
		Data:          invokeResponse.Data,
	}
	return executeResponse, nil
}

// getOrStart returns a available vm from the pool or starts a new one.
func (v *runnerOperator) getOrStartVm(ctx context.Context, cfg *runner.Config) (*runner.RunnerInstance, error) {
	instance, err := v.runnerPool.AvailableRunner(cfg.FunctionUuid)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		log.Debugf("no instance available for function id: %s, reason: %v", cfg.FunctionUuid, err)
		instance := runner.NewInstance(ctx, cfg)
		if err := instance.CreateAndStart(ctx); err != nil {
			log.Errorf("failed to create and start new firecracker machine: %s", instance.Config().RunnerUuid)
			return nil, err
		}
	}
	// update machine status in pool
	instance.SetState(runner.RunnerStateBusy)
	return instance, nil
}

func (r *runnerOperator) availableRunner(functionUuid string) (*runner.RunnerInstance, error) {
	instance, err := r.runnerPool.AvailableRunner(functionUuid)
	if err != nil {
		return nil, err
	}
	instance.SetState(runner.RunnerStateReserved)
	return instance, nil
}

func (r *runnerOperator) runner(functionUuid string, runnerUuid string) (*runner.RunnerInstance, error) {
	instance, err := r.runnerPool.Get(functionUuid, runnerUuid)
	if err != nil {
		return nil, err
	}
	return instance, err
}
