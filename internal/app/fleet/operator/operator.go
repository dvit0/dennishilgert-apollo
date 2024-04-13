package operator

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/initializer"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator/pool"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	taskRunner "github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	agentpb "github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	fleetpb "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/google/uuid"
)

var log = logger.NewLogger("apollo.manager.operator")

type Options struct {
	WorkerUuid                string
	AgentApiPort              int
	MessagingBootstrapServers string
	OsArch                    utils.OsArch
	FirecrackerBinaryPath     string
	WatchdogCheckInterval     time.Duration
	WatchdogWorkerCount       int
}

type RunnerOperator interface {
	Init(ctx context.Context) error
	Runner(functionUuid string, runnerUuid string) (runner.RunnerInstance, error)
	AvailableRunner(request *fleetpb.AvailableRunnerRequest) (*fleetpb.AvailableRunnerResponse, error)
	ProvisionRunner(request *fleetpb.ProvisionRunnerRequest) (*fleetpb.ProvisionRunnerResponse, error)
	InvokeFunction(ctx context.Context, request *fleetpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error)
}

type runnerOperator struct {
	workerUuid                string
	osArch                    utils.OsArch
	firecrackerBinaryPath     string
	agentApiPort              int
	messagingBootstrapServers string
	runnerTeardownCh          chan runner.RunnerInstance
	runnerPool                pool.RunnerPool
	runnerPoolWatchdog        pool.RunnerPoolWatchdog
	runnerInitializer         initializer.RunnerInitializer
	appCtx                    context.Context
	runnerCtx                 context.Context
	runnerCtxCancel           context.CancelFunc
}

// NewRunnerOperator creates a new Operator.
func NewRunnerOperator(ctx context.Context, runnerInitializer initializer.RunnerInitializer, opts Options) (RunnerOperator, error) {
	runnerTeardownCh := make(chan runner.RunnerInstance)

	runnerPool := pool.NewRunnerPool()
	runnerPoolWatchdog := pool.NewRunnerPoolWatchdog(
		runnerPool,
		runnerTeardownCh,
		pool.WatchdogOptions{
			CheckInterval: opts.WatchdogCheckInterval,
			WorkerCount:   opts.WatchdogWorkerCount,
		},
	)

	runnerCtx, runnerCtxCancel := context.WithCancel(context.Background())

	return &runnerOperator{
		workerUuid:                opts.WorkerUuid,
		osArch:                    opts.OsArch,
		firecrackerBinaryPath:     opts.FirecrackerBinaryPath,
		agentApiPort:              opts.AgentApiPort,
		messagingBootstrapServers: opts.MessagingBootstrapServers,
		runnerTeardownCh:          runnerTeardownCh,
		runnerPool:                runnerPool,
		runnerPoolWatchdog:        runnerPoolWatchdog,
		runnerInitializer:         runnerInitializer,
		appCtx:                    ctx,
		runnerCtx:                 runnerCtx,
		runnerCtxCancel:           runnerCtxCancel,
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
				case <-v.appCtx.Done():
					log.Info("tearing down all runners")

					// Cancel runner context after 1 minute.
					// This gives the opportunity to shutdown all runners gracefully.
					time.AfterFunc(1*time.Minute, func() {
						log.Warnf("runner teardown timeout reached - killing left over runners")
						v.runnerCtxCancel()
					})

					v.TeardownRunners(v.runnerCtx)
					return nil
				case runnerInstance := <-v.runnerTeardownCh:
					if err := v.TeardownRunner(v.runnerCtx, runnerInstance); err != nil {
						log.Errorf("failed to tear down runner: %s - reason: %v", runnerInstance.Config().RunnerUuid, err)
					}
				}
			}
		},
	)
	return runnerManager.Run(ctx)
}

func (r *runnerOperator) Runner(functionUuid string, runnerUuid string) (runner.RunnerInstance, error) {
	return r.runnerPool.Get(functionUuid, runnerUuid)
}

func (r *runnerOperator) AvailableRunner(request *fleetpb.AvailableRunnerRequest) (*fleetpb.AvailableRunnerResponse, error) {
	instance, err := r.runnerPool.AvailableRunner(request.FunctionUuid)
	if err != nil {
		return nil, err
	}
	instance.SetState(runner.RunnerStateReserved)
	response := &fleetpb.AvailableRunnerResponse{
		RunnerUuid: instance.Config().RunnerUuid,
	}
	return response, nil
}

// ProvisionRunner provisions a new runner with specified parameters.
func (v *runnerOperator) ProvisionRunner(request *fleetpb.ProvisionRunnerRequest) (*fleetpb.ProvisionRunnerResponse, error) {
	runnerUuid := uuid.New().String()

	multiThreading := false
	if v.osArch == utils.Arch_x86_64 {
		multiThreading = true
	}
	logLevel := log.LogLevel()
	if request.Machine.LogLevel != nil {
		logLevel = *request.Machine.LogLevel
	}
	cfg := &runner.Config{
		WorkerUuid:            v.workerUuid,
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
				naming.RunnerSocketFileName(),
			},
			string(os.PathSeparator),
		),
		LogFilePath: strings.Join(
			[]string{
				naming.RunnerStoragePath(v.runnerInitializer.DataPath(), runnerUuid),
				naming.RunnerLogFileName(),
			},
			string(os.PathSeparator),
		),
		StdOutFilePath: strings.Join(
			[]string{
				naming.RunnerStoragePath(v.runnerInitializer.DataPath(), runnerUuid),
				naming.RunnerStdOutFileName(),
			},
			string(os.PathSeparator),
		),
		StdErrFilePath: strings.Join(
			[]string{
				naming.RunnerStoragePath(v.runnerInitializer.DataPath(), runnerUuid),
				naming.RunnerStdErrFileName(),
			},
			string(os.PathSeparator),
		),
		VCpuCount:                int(request.Machine.VcpuCores),
		MemSizeMib:               int(request.Machine.MemoryLimit),
		IdleTtl:                  time.Duration(request.Machine.IdleTtl) * time.Minute,
		Multithreading:           multiThreading,
		AgentApiPort:             v.agentApiPort,
		MessagingBoostrapServers: v.messagingBootstrapServers,
		LogLevel:                 logLevel,
	}

	//
	if err := v.runnerInitializer.InitializeRunner(v.runnerCtx, cfg); err != nil {
		// We don't care if the runner storage removal throws an error as this is just for cleanup
		v.runnerInitializer.RemoveRunner(v.runnerCtx, runnerUuid)
		return nil, err
	}

	// Create and start new runner instance.
	// It is important to use the app context here as the instance would be terminated
	// after the request is done.
	instance, err := runner.NewInstance(v.runnerCtx, cfg)
	if err != nil {
		return nil, err
	}
	if err := instance.CreateAndStart(v.runnerCtx); err != nil {
		if err := instance.ShutdownAndDestroy(v.runnerCtx); err != nil {
			log.Errorf("failed to shutdown runner instance: %v", err)
		}
		return nil, err
	}

	// Add runner to the runner pool.
	if err := v.runnerPool.Add(instance); err != nil {
		if err := instance.ShutdownAndDestroy(v.runnerCtx); err != nil {
			log.Errorf("failed to shutdown runner instance: %v", err)
		}
		return nil, err
	}

	// Waiting for the runner to become ready.
	if err := instance.Ready(v.runnerCtx); err != nil {
		// Remove runner from pool.
		v.runnerPool.Remove(request.FunctionUuid, runnerUuid)

		if err := instance.ShutdownAndDestroy(v.runnerCtx); err != nil {
			log.Errorf("failed to shutdown runner instance: %v", err)
		}
		return nil, err
	}

	response := &fleetpb.ProvisionRunnerResponse{
		RunnerUuid: instance.Config().RunnerUuid,
	}
	return response, nil
}

// TeardownRunner tears down a specified runner.
func (r *runnerOperator) TeardownRunner(ctx context.Context, runnerInstance runner.RunnerInstance) error {
	if err := runnerInstance.ShutdownAndDestroy(ctx); err != nil {
		return fmt.Errorf("failed to shutdown runner instance: %w", err)
	}
	// TODO: uncomment to enable runner directory cleanup after teardown.
	// if err := r.runnerInitializer.RemoveRunner(ctx, runnerInstance.Config().RunnerUuid); err != nil {
	// 	log.Errorf("failed to remove runner storage: %s", runnerInstance.Config().RunnerUuid)
	// 	return err
	// }
	return nil
}

// TeardownRunners tears down all runners in the pool.
func (r *runnerOperator) TeardownRunners(ctx context.Context) {
	functionCount := len(*r.runnerPool.Pool())
	functionIndex := 1

	for functionUuid, runnersByFunction := range *r.runnerPool.Pool() {
		runnerCount := len(runnersByFunction)
		runnerIndex := 1

		if runnerCount > 0 {
			log.Debugf("tearing down runners of function (%d/%d): %s", functionIndex, functionCount, functionUuid)
		}

		for runnerUuid, runnerInstance := range runnersByFunction {
			log.Debugf("tearing down runner (%d/%d): %s", runnerIndex, runnerCount, runnerUuid)
			r.runnerPool.Remove(functionUuid, runnerUuid)
			if err := r.TeardownRunner(ctx, runnerInstance); err != nil {
				log.Warnf("failed to tear down runner - manual cleanup needed: %s", runnerUuid)
			}
			runnerIndex += 1
		}
		functionIndex += 1
	}
}

// InvokeFunction invokes the function inside of a specified runner.
func (r *runnerOperator) InvokeFunction(ctx context.Context, request *fleetpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error) {
	instance, err := r.runnerPool.Get(request.FunctionUuid, request.RunnerUuid)
	if err != nil {
		return nil, fmt.Errorf("runner not found: %w", err)
	}
	invokeRequest := &agentpb.InvokeRequest{
		Context: &agentpb.ContextData{
			Runtime:        "mocked - will be removed in future",
			RuntimeVersion: "mocked - will be removed in future",
			RuntimeHandler: "mocked - will be removed in future",
			VCpuCores:      -1,
			MemoryLimit:    -1,
		},
		Event: &agentpb.EventData{
			Uuid: request.Event.Uuid,
			Type: request.Event.Type,
			Data: request.Event.Data,
		},
	}
	// invoke the function code with the data of the event
	invokeResponse, err := instance.Invoke(ctx, invokeRequest)
	if err != nil {
		return nil, fmt.Errorf("function invocation failed: %w", err)
	}
	// TODO: Send billed duration to user service and logs to the log collector via messaging.
	// IDEA: This can be done directly inside the runner. The connection to the messaging
	// service is already implemented, so why not use it.
	log.Infof("Billed duration for execution: %s, duration: %s", invokeResponse.EventUuid, invokeResponse.Duration)
	log.Infof("Logs for execution: %s, logs: %v", invokeResponse.EventUuid, invokeResponse.Logs)

	response := &fleetpb.InvokeFunctionResponse{
		EventUuid:     invokeResponse.EventUuid,
		Status:        invokeResponse.Status,
		StatusMessage: invokeResponse.StatusMessage,
		Data:          invokeResponse.Data,
	}
	return response, nil
}
