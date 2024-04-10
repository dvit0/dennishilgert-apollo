package messaging

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var config = DefaultConfig()

func GetDefaultAdminClient(bootstrapServers string) (*kafka.AdminClient, error) {
	return kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
}

func CreateTopic(ctx context.Context, client *kafka.AdminClient, log logger.Logger, topic string) (kafka.TopicResult, error) {
	result, err := CreateTopics(ctx, client, log, []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	}})
	return result[0], err
}

func CreateTopics(ctx context.Context, client *kafka.AdminClient, log logger.Logger, topics []kafka.TopicSpecification) ([]kafka.TopicResult, error) {
	result, err := client.CreateTopics(
		ctx,
		topics,
		kafka.SetAdminOperationTimeout(config.AdminOperationTimeout),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}
