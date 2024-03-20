package messaging

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.messaging")

type Options struct {
	BootstrapServers string
}

type Service interface {
	Start(ctx context.Context) error
}

type messagingService struct {
	consumer *kafka.Consumer
}

func NewMessagingService(opts Options) (Service, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
		"group.id":          "fleet_manager_group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging consumer: %v", err)
	}

	return &messagingService{
		consumer: consumer,
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
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		for {
			select {
			case <-ctx.Done():
				log.Debug("stop listening for messages")
				errCh <- nil // signal no error on graceful shutdown
				return
			default:
				// Poll for a message
				msg, err := m.consumer.ReadMessage(-1)
				if err != nil {
					switch errType := err.(type) {
					case kafka.Error:
						// Kafka error that isn't a fatal error
						if errType.IsFatal() {
							errCh <- err
							return
						} else {
							// log non fatal kafka errors and continue
							log.Warnf("kafka error while reading message: %v", err)
							continue
						}
					default:
						errCh <- err
						return
					}
				}

				// handle received message
				log.Infof("received message on %s: %s", msg.TopicPartition, string(msg.Value))
			}
		}
	}()

	// block until the context is done or an error occurs
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
