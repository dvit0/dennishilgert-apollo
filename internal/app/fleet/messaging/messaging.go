package messaging

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/fleet/messaging/handler"
	"github.com/dennishilgert/apollo/internal/app/fleet/preparer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
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

	errCh := make(chan error, 1)
	go func() {
		if err := m.worker.Run(ctx); err != nil {
			errCh <- fmt.Errorf("failed to run worker manager: %v", err)
		}
	}()

	go func() {
		defer close(errCh) // Ensure channel is closed to avoid goroutine leak.

		for {
			select {
			case <-ctx.Done():
				log.Debug("stop listening for messages")
				errCh <- nil // Signal no error on graceful shutdown.
				return
			default:
				// Poll for a message.
				msg, err := m.consumer.ReadMessage(-1)
				if err != nil {
					switch errType := err.(type) {
					case kafka.Error:
						if errType.IsFatal() {
							errCh <- err
							return
						} else {
							// Log non fatal kafka errors and continue.
							log.Warnf("kafka error while reading message: %v", err)
							continue
						}
					default:
						errCh <- err
						return
					}
				}

				// Handle received message.
				handler := m.handlers[*msg.TopicPartition.Topic]
				if handler == nil {
					errCh <- fmt.Errorf("failed to find handler for topic: %s", *msg.TopicPartition.Topic)
					return
				}

				// Executing a task resulting from a message can take a lot of computing time (like initializing a function).
				// To avoid blocking the receive routine, each message is handled by a worker.
				task := worker.NewTask[struct{}](func(ctx context.Context) (struct{}, error) {
					handler(msg)
					return struct{}{}, nil
				})
				m.worker.Add(task)
			}
		}
	}()

	// Block until the context is done or an error occurs.
	var serveErr error
	select {
	case <-ctx.Done():
		log.Info("shutting down messaging service")
	case err := <-errCh:
		if err != nil {
			serveErr = err
			log.Errorf("error while reading message: %v", err)
		}
	}

	if cErr := m.consumer.Close(); cErr != nil && serveErr == nil {
		log.Errorf("failed to close message consumer: %v", cErr)
	}

	return serveErr
}
