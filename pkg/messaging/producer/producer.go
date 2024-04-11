package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.messaging.producer")

type Options struct {
	BootstrapServers string
}

type MessagingProducer interface {
	Publish(ctx context.Context, topic string, message interface{})
	Close() error
}

type messagingProducer struct {
	producer *kafka.Producer
}

func NewMessagingProducer(ctx context.Context, opts Options) (MessagingProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": opts.BootstrapServers,
	})
	if err != nil {
		log.Error("failed to create messaging producer")
		return nil, err
	}

	// Start a goroutine to listen for delivery reports.
	// Handling the delivery reports prevent the producer flush
	// from reporting the sent messages as not sent.
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Infof("shutting down messaging producer")
				return
			case e := <-producer.Events():
				switch event := e.(type) {
				case *kafka.Message:
					if event.TopicPartition.Error != nil {
						log.Errorf("failed to deliver message: %v", event.TopicPartition.Error)
					} else {
						log.Debugf("successfully delivered message to topic %s [%d] at offset %v", *event.TopicPartition.Topic, event.TopicPartition.Partition, event.TopicPartition.Offset)
					}
				case kafka.Error:
					log.Errorf("failed to send kafka message: %v", event)
				default:
					if e != nil {
						log.Debugf("ignored kafka event: %v", e)
					}
				}
			}
		}
	}()

	return &messagingProducer{
		producer: producer,
	}, nil
}

func (m *messagingProducer) Publish(ctx context.Context, topic string, message interface{}) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Errorf("failed to marshal json message: %v", err)
		return
	}
	if err := m.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          jsonMessage,
	}, nil); err != nil {
		log.Errorf("failed to enqueue message to topic: %s - message: %v - error: %v", topic, message, err)
	}
	log.Debugf("enqueued message to topic: %s - message: %v", topic, message)
}

func (m *messagingProducer) Close() error {
	unsentMessages := m.producer.Flush(1000 * 5)
	m.producer.Close()
	if unsentMessages > 0 {
		return fmt.Errorf("failed to flush unsent messages: %d", unsentMessages)
	}
	return nil
}
