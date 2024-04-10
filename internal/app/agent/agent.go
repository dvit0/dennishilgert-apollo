package agent

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/agent/api"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/health"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/pkg/proto/messages/v1"
)

var log = logger.NewLogger("apollo.agent")

// Agent is an Apollo application that runs inside a Firecracker Micro VM.
type Agent interface {
	Run(ctx context.Context) error
}

// Options contains the options for `NewAgent`.
type Options struct {
	ManagerUuid       string
	FunctionUuid      string
	RunnerUuid        string
	RuntimeHandler    string
	RuntimeBinaryPath string
	RuntimeBinaryArgs []string

	ApiPort                   int
	MessagingBootstrapServers string
}

type agent struct {
	managerUuid       string
	functionUuid      string
	runnerUuid        string
	runtimeHandler    string
	runtimeBinaryPath string
	runtimeBinaryArgs []string

	apiServer         api.Server
	messagingProducer producer.MessagingProducer
}

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
		return nil, fmt.Errorf("error while creating messaging producer: %v", err)
	}

	return &agent{
		managerUuid:       opts.ManagerUuid,
		functionUuid:      opts.FunctionUuid,
		runnerUuid:        opts.RunnerUuid,
		runtimeHandler:    opts.RuntimeHandler,
		runtimeBinaryPath: opts.RuntimeBinaryPath,
		runtimeBinaryArgs: opts.RuntimeBinaryArgs,
		apiServer:         apiServer,
		messagingProducer: messagingProducer,
	}, nil
}

func (a *agent) Run(ctx context.Context) error {
	log.Info("apollo agent is starting")

	healthStatusProvider := health.NewHealthStatusProvider(health.ProviderOptions{
		Targets: 1,
	})
	readyCallback := func() {
		log.Infof("signalizing agent ready")
		a.messagingProducer.Publish(ctx, naming.MessagingManagerRelatedAgentReadyTopic(a.managerUuid), messages.RunnerAgentReadyMessage{
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
				return fmt.Errorf("failed to start api server: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := a.apiServer.Ready(ctx); err != nil {
				return fmt.Errorf("api server did not become ready in time: %v", err)
			}
			healthStatusProvider.Ready()
			log.Info("api server started")

			// Wait for the main context to be done.
			<-ctx.Done()
			return nil
		},
	)

	return runner.Run(ctx)
}
