package agent

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/agent/api"
	"github.com/dennishilgert/apollo/internal/app/agent/runtime"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging/producer"
	messagespb "github.com/dennishilgert/apollo/pkg/proto/messages/v1"
)

var log = logger.NewLogger("apollo.agent")

// Options contains the options for `NewAgent`.
type Options struct {
	WorkerUuid        string
	FunctionUuid      string
	RunnerUuid        string
	RuntimeHandler    string
	RuntimeBinaryPath string
	RuntimeBinaryArgs []string

	ApiPort                   int
	MessagingBootstrapServers string
}

// Agent is an Apollo application that runs inside a Firecracker Micro VM.
type Agent interface {
	Run(ctx context.Context) error
}

type agent struct {
	workerUuid        string
	functionUuid      string
	runnerUuid        string
	runtimeHandler    string
	apiServer         api.Server
	messagingProducer producer.MessagingProducer
	persistentRuntime runtime.PersistentRuntime
}

// NewAgent creates a new Agent instance.
func NewAgent(ctx context.Context, opts Options) (Agent, error) {
	apiServer := api.NewApiServer(api.Options{
		Port: opts.ApiPort,
	})

	messagingProducer, err := producer.NewMessagingProducer(
		ctx,
		producer.Options{
			BootstrapServers: opts.MessagingBootstrapServers,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating messaging producer failed: %w", err)
	}

	persistentRuntime, err := runtime.NewPersistentRuntime(ctx, runtime.Config{
		BinaryPath: opts.RuntimeBinaryPath,
		BinaryArgs: opts.RuntimeBinaryArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing runtime failed: %w", err)
	}

	return &agent{
		workerUuid:        opts.WorkerUuid,
		functionUuid:      opts.FunctionUuid,
		runnerUuid:        opts.RunnerUuid,
		runtimeHandler:    opts.RuntimeHandler,
		apiServer:         apiServer,
		messagingProducer: messagingProducer,
		persistentRuntime: persistentRuntime,
	}, nil
}

// Run starts the agent.
func (a *agent) Run(ctx context.Context) error {
	log.Info("apollo agent is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 2,
	})
	readyCallback := func() {
		log.Infof("signalizing agent ready")
		a.messagingProducer.Publish(ctx, naming.MessagingWorkerRelatedAgentReadyTopic(a.workerUuid), messagespb.RunnerAgentReadyMessage{
			FunctionUuid: a.functionUuid,
			RunnerUuid:   a.runnerUuid,
			Reason:       "ok",
			Success:      true,
		})
	}
	healthStatusProvider.WithCallback(readyCallback)

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("starting api server")
			if err := a.apiServer.Run(ctx, healthStatusProvider); err != nil {
				return fmt.Errorf("starting api server failed: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := a.apiServer.Ready(ctx); err != nil {
				return fmt.Errorf("api server did not become ready in time: %w", err)
			}
			healthStatusProvider.Ready()
			log.Info("api server started")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
		func(ctx context.Context) error {
			log.Info("starting persistent runtime")
			if err := a.persistentRuntime.Start(a.runtimeHandler); err != nil {
				return fmt.Errorf("starting runtime failed: %w", err)
			}

			healthStatusProvider.Ready()
			log.Info("persistent runtime started")

			// Wait for the main context to be done.
			<-ctx.Done()

			if err := a.persistentRuntime.Tidy(); err != nil {
				return fmt.Errorf("tidying runtime failed: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			log.Info("setting up runtime crash recovery")
			const maxRecoveryCount = 3
			var recoveryCounter int

			for {
				if recoveryCounter >= maxRecoveryCount {
					return fmt.Errorf("max runtime recovery attempts reached")
				}

				err := a.persistentRuntime.Wait()
				if err != nil {
					log.Errorf("runtime finished with an error: %v", err)
				}

				// Directly check for context cancellation.
				if ctx.Err() != nil {
					return fmt.Errorf("context canceled, halting recovery process: %w", ctx.Err())
				} else {
					recoveryCounter++
					log.Warnf("runtime has finished unexpectedly - recovery %d of max %d", recoveryCounter, maxRecoveryCount)
					if err := a.recoverRuntime(ctx); err != nil {
						return fmt.Errorf("recovering runtime failed: %w", err)
					}
				}
			}
		},
	)

	return runner.Run(ctx)
}

// recoverRuntime recovers the runtime by creating a new persistent runtime instance.
func (a *agent) recoverRuntime(ctx context.Context) error {
	if err := a.persistentRuntime.Tidy(); err != nil {
		return fmt.Errorf("tidying runtime failed: %w", err)
	}

	runtimeConfig := a.persistentRuntime.Config()
	newPersistentRuntime, err := runtime.NewPersistentRuntime(ctx, runtimeConfig)
	if err != nil {
		return fmt.Errorf("initializing replacement runtime failed: %w", err)
	}
	if err := newPersistentRuntime.Start(a.runtimeHandler); err != nil {
		return fmt.Errorf("starting replacement runtime failed: %w", err)
	}
	a.persistentRuntime = newPersistentRuntime

	log.Info("runtime recovered")
	return nil
}
