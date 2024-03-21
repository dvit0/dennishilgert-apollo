package messaging

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/handler"
	"github.com/dennishilgert/apollo/internal/app/fleet/preparer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/concurrency/worker"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.messaging")

type Options struct {
	BootstrapServers string
	WorkerCount      int
}

type MessagingService interface {
	Start(ctx context.Context) error
}

type messagingService struct {
	runnerPreparer preparer.RunnerPreparer
	worker         worker.WorkerManager
	consumer       *kafka.Consumer
	handlers       map[string]func(msg *kafka.Message)
}

func NewMessagingService(runnerPreparer preparer.RunnerPreparer, opts Options) (MessagingService, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
		"group.id":          "fleet_manager_group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging consumer: %v", err)
	}

	messagingHandler := handler.NewMessagingHandler(runnerPreparer)
	messagingHandler.RegisterAll()

	return &messagingService{
		runnerPreparer: runnerPreparer,
		worker:         worker.NewWorkerManager(opts.WorkerCount),
		consumer:       consumer,
		handlers:       messagingHandler.Handlers(),
	}, nil
}

func (m *messagingService) Start(ctx context.Context) error {
	subscribedTopics := []string{
		naming.MessagingFunctionInitializationTopic,
	}
	if sErr := m.consumer.SubscribeTopics(subscribedTopics, nil); sErr != nil {
		return fmt.Errorf("failed to subscribe to topics: %v", sErr)
	}

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			if err := m.worker.Run(ctx); err != nil {
				return fmt.Errorf("failed to run worker manager: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			defer func() {
				if err := m.consumer.Close(); err != nil {
					log.Errorf("failed to close message consumer: %v", err)
				}
			}()

			for {
				select {
				case <-ctx.Done():
					// context cancelled, return the reason
					log.Info("stopping messaging service")
					return ctx.Err()
				default:
					// Poll for a message.
					e := m.consumer.Poll(100)
					switch event := e.(type) {
					case *kafka.Message:
						// Handle received message.
						handler := m.handlers[*event.TopicPartition.Topic]
						if handler == nil {
							return fmt.Errorf("failed to find handler for topic: %s", *event.TopicPartition.Topic)
						}

						// Executing a task resulting from a message can take a lot of computing time (like initializing a function).
						// To avoid blocking the receive routine, each message is handled by a worker.
						task := worker.NewTask[struct{}](func(ctx context.Context) (struct{}, error) {
							handler(event)
							return struct{}{}, nil
						})
						m.worker.Add(task)
					case kafka.PartitionEOF:
						log.Debugf("no more kafka messages to read at the moment")
						// There are no more messages to read at the moment.
					case kafka.Error:
						return fmt.Errorf("error while polling for kafka messages: %v", event)
					default:
						log.Debugf("ignored kafka event: %v", e)
					}
				}
			}
		},
	)

	return runner.Run(ctx)
}
