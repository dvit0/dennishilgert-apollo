package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/handler"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/concurrency/worker"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.messaging.consumer")

type Options struct {
	ManagerUuid      string
	BootstrapServers string
	WorkerCount      int
}

type MessagingConsumer interface {
	SetupDone()
	Start(ctx context.Context) error
}

type messagingConsumer struct {
	managerUuid string
	worker      worker.WorkerManager
	consumer    *kafka.Consumer
	handlers    map[string]func(msg *kafka.Message)
	setupDoneCh chan bool
}

func NewMessagingConsumer(runnerOperator operator.RunnerOperator, opts Options) (MessagingConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
		"group.id":          "fleet_manager_group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Error("failed to create messaging consumer")
		return nil, err
	}

	messagingHandler := handler.NewMessagingHandler(runnerOperator, handler.Options{
		ManagerUuid: opts.ManagerUuid,
	})
	messagingHandler.RegisterAll()

	return &messagingConsumer{
		managerUuid: opts.ManagerUuid,
		worker:      worker.NewWorkerManager(opts.WorkerCount),
		consumer:    consumer,
		handlers:    messagingHandler.Handlers(),
		setupDoneCh: make(chan bool, 1),
	}, nil
}

// SetupDone signalizes that the messaging setup is done.
func (m *messagingConsumer) SetupDone() {
	m.setupDoneCh <- true
}

// Start subscribes to the specified topics and starts listening for incoming messages.
func (m *messagingConsumer) Start(ctx context.Context) error {
	log.Info("waiting until messaging setup is done")
	<-m.setupDoneCh

	subscribedTopics := []string{
		naming.MessagingFunctionInitializationTopic,
		naming.MessagingManagerRelatedAgentReadyTopic(m.managerUuid),
	}
	if err := m.consumer.SubscribeTopics(subscribedTopics, nil); err != nil {
		log.Error("failed to subscribe to topics")
		return err
	}

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			if err := m.worker.Run(ctx); err != nil {
				log.Error("failed to run worker manager")
				return err
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
					log.Info("shutting down messaging consumer")
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
						}, 10*time.Second)
						m.worker.Add(task)
					case kafka.PartitionEOF:
						log.Debugf("no more kafka messages to read at the moment")
						// There are no more messages to read at the moment.
					case kafka.Error:
						log.Error("error while polling for kafka messages")
						return event
					default:
						if e != nil {
							log.Debugf("ignored kafka event: %v", e)
						}
					}
				}
			}
		},
	)
	return runner.Run(ctx)
}
