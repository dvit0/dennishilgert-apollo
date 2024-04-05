package operator

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator/pool"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	taskRunner "github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/google/uuid"
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
	AvailableRunner(request *fleet.AvailableRunnerRequest) (*fleet.AvailableRunnerResponse, error)
	ProvisionRunner(ctx context.Context, request *fleet.ProvisionRunnerRequest) (*fleet.ProvisionRunnerResponse, error)
	InvokeFunction(ctx context.Context, request *fleet.InvokeFunctionRequest) (*fleet.InvokeFunctionResponse, error)
}

type runnerOperator struct {
	osArch                utils.OsArch
	firecrackerBinaryPath string
	agentApiPort          int
	runnerTeardownCh      chan runner.TeardownParams
	runnerPool            pool.RunnerPool
	runnerPoolWatchdog    pool.RunnerPoolWatchdog
	runnerInitializer     initializer.RunnerInitializer
}

// NewRunnerOperator creates a new Operator.
func NewRunnerOperator(runnerInitializer initializer.RunnerInitializer, opts Options) (RunnerOperator, error) {
	runnerTeardownCh := make(chan runner.TeardownParams)

	runnerPool := pool.NewRunnerPool()
	runnerPoolWatchdog := pool.NewRunnerPoolWatchdog(
		runnerPool,
		runnerTeardownCh,
		pool.WatchdogOptions{
			CheckInterval: opts.WatchdogCheckInterval,
			WorkerCount:   opts.WatchdogWorkerCount,
		},
	)

	return &runnerOperator{
		osArch:                opts.OsArch,
		firecrackerBinaryPath: opts.FirecrackerBinaryPath,
		agentApiPort:          opts.AgentApiPort,
		runnerTeardownCh:      runnerTeardownCh,
		runnerPool:            runnerPool,
		runnerPoolWatchdog:    runnerPoolWatchdog,
		runnerInitializer:     runnerInitializer,
	}, nil
}

// Init initializes the runner operator.
func (v *runnerOperator) Init(ctx context.Context) error {
	runnerManager := taskRunner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting runner pool watchdog")
			if err := v.runnerPoolWatchdog.Run(ctx); err != nil {
				log.Error("error while starting runner pool watchdog")
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			defer close(v.runnerTeardownCh)

			for {
				select {
				case <-ctx.Done():
					log.Info("tearing down all runners")
					// Use new context as the main context has already been cancelled.
					v.TeardownRunners(context.Background())
					return nil
				case params := <-v.runnerTeardownCh:
					if err := v.TeardownRunner(ctx, params.FunctionUuid, params.RunnerUuid); err != nil {
						return err
					}
				}
			}
		},
	)
	return runnerManager.Run(ctx)
}

func (r *runnerOperator) AvailableRunner(request *fleet.AvailableRunnerRequest) (*fleet.AvailableRunnerResponse, error) {
	instance, err := r.runnerPool.AvailableRunner(request.FunctionUuid)
	if err != nil {
		return nil, err
	}
	instance.SetState(runner.RunnerStateReserved)
	response := &fleet.AvailableRunnerResponse{
		RunnerUuid: instance.Config().RunnerUuid,
	}
	return response, nil
}

// ProvisionRunner provisions a new runner with specified parameters.
func (v *runnerOperator) ProvisionRunner(ctx context.Context, request *fleet.ProvisionRunnerRequest) (*fleet.ProvisionRunnerResponse, error) {
	runnerUuid := uuid.New().String()

	if err := v.runnerInitializer.InitializeRunner(ctx, runnerUuid); err != nil {
		// We don't care if the runner storage removal throws an error as this is just for cleanup
		v.runnerInitializer.RemoveRunner(ctx, runnerUuid)
		return nil, err
	}

	multiThreading := false
	if v.osArch == utils.Arch_x86_64 {
		multiThreading = true
	}
	cfg := &runner.Config{
		FunctionUuid:          request.FunctionUuid,
		RunnerUuid:            runnerUuid,
		HostOsArch:            v.osArch,
		FirecrackerBinaryPath: v.firecrackerBinaryPath,
		KernelImagePath: strings.Join(
			[]string{
				naming.KernelStoragePath(v.runnerInitializer.DataPath(), request.Kernel.Name, request.Kernel.Version),
				naming.KernelFileName(request.Kernel.Name, request.Kernel.Version),
			},
			string(os.PathSeparator),
		),
		RuntimeDrivePath: strings.Join(
			[]string{
				naming.RuntimeStoragePath(v.runnerInitializer.DataPath(), request.Runtime.Name, request.Runtime.Version),
				naming.RuntimeImageFileName(request.Runtime.Name, request.Runtime.Version),
			},
			string(os.PathSeparator),
		),
		RuntimeHandler:    request.Runtime.Handler,
		RuntimeBinaryPath: request.Runtime.BinaryPath,
		RuntimeBinaryArgs: request.Runtime.BinaryArgs,
		FunctionDrivePath: strings.Join(
			[]string{
				naming.FunctionStoragePath(v.runnerInitializer.DataPath(), request.FunctionUuid),
				naming.FunctionImageFileName(request.FunctionUuid),
			},
			string(os.PathSeparator),
		),
		SocketPath: strings.Join(
			[]string{
				naming.RunnerStoragePath(v.runnerInitializer.DataPath(), runnerUuid),
				naming.RunnerSocketFileName(runnerUuid),
			},
			string(os.PathSeparator),
		),
		LogFilePath: strings.Join(
			[]string{
				naming.RunnerStoragePath(v.runnerInitializer.DataPath(), runnerUuid),
				naming.RunnerLogFileName(runnerUuid),
			},
			string(os.PathSeparator),
		),
		VCpuCount:      int(request.Machine.VcpuCores),
		MemSizeMib:     int(request.Machine.MemoryLimit),
		IdleTtl:        time.Duration(request.Machine.IdleTtl) * time.Minute,
		Multithreading: multiThreading,
		AgentApiPort:   v.agentApiPort,
	}

	// Create and start new runner instance.
	instance := runner.NewInstance(ctx, cfg)
	if err := instance.CreateAndStart(ctx); err != nil {
		return nil, err
	}

	// Waiting for the runner to become ready.
	if err := instance.Ready(ctx); err != nil {
		return nil, err
	}

	// Add runner to the runner pool.
	if err := v.runnerPool.Add(instance); err != nil {
		return nil, err
	}

	response := &fleet.ProvisionRunnerResponse{
		RunnerUuid: instance.Config().RunnerUuid,
	}
	return response, nil
}

// TeardownRunner tears down a specified runner.
func (r *runnerOperator) TeardownRunner(ctx context.Context, functionUuid string, runnerUuid string) error {
	instance, err := r.runnerPool.Get(functionUuid, runnerUuid)
	if err != nil {
		log.Errorf("failed to teardown runner: %s", runnerUuid)
		return err
	}
	if err := instance.ShutdownAndDestroy(ctx); err != nil {
		log.Errorf("error while shutting down runner: %s", runnerUuid)
		return err
	}
	if err := r.runnerInitializer.RemoveRunner(ctx, runnerUuid); err != nil {
		log.Errorf("failed to remove runner storage: %s", runnerUuid)
		return err
	}
	return nil
}

// TeardownRunners tears down all runners in the pool.
func (r *runnerOperator) TeardownRunners(ctx context.Context) {
	functionCount := len(*r.runnerPool.Pool())
	functionIndex := 1

	for functionUuid, runnersByFunction := range *r.runnerPool.Pool() {
		runnerCount := len(runnersByFunction)
		runnerIndex := 1

		log.Debugf("tearing down runners of function (%d/%d): %s", functionIndex, functionCount, functionUuid)

		for runnerUuid := range runnersByFunction {
			log.Debugf("tearing down runner (%d/%d): %s", runnerIndex, runnerCount, runnerUuid)
			r.runnerPool.Remove(functionUuid, runnerUuid)
			if err := r.TeardownRunner(ctx, functionUuid, runnerUuid); err != nil {
				log.Warnf("failed to tear down runner: %s", runnerUuid)
			}
			runnerIndex += 1
		}
		functionIndex += 1
	}
}

// InvokeFunction invokes the function inside of a specified runner.
func (r *runnerOperator) InvokeFunction(ctx context.Context, request *fleet.InvokeFunctionRequest) (*fleet.InvokeFunctionResponse, error) {
	instance, err := r.runnerPool.Get(request.FunctionUuid, request.RunnerUuid)
	if err != nil {
		return nil, err
	}
	invokeRequest := &agent.InvokeRequest{
		Context: &agent.ContextData{
			Runtime:        "mocked - will be removed in future",
			RuntimeVersion: "mocked - will be removed in future",
			RuntimeHandler: "mocked - will be removed in future",
			VCpuCores:      -1,
			MemoryLimit:    -1,
		},
		Event: &agent.EventData{
			Uuid: request.Event.Uuid,
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
	log.Infof("Billed duration for execution: %s, duration: %s", invokeResponse.EventUuid, invokeResponse.Duration)
	// TODO: send logs to logs service via messaging
	log.Infof("Logs for execution: %s, logs: %v", invokeResponse.EventUuid, invokeResponse.Logs)

	response := &fleet.InvokeFunctionResponse{
		EventUuid:     invokeResponse.EventUuid,
		Status:        invokeResponse.Status,
		StatusMessage: invokeResponse.StatusMessage,
		Data:          invokeResponse.Data,
	}
	return response, nil
}
