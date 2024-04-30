package messaging

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var config = DefaultConfig()

// GetDefaultAdminClient creates a new kafka admin client with the default configuration.
func GetDefaultAdminClient(bootstrapServers string) (*kafka.AdminClient, error) {
	return kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
}

// CreateTopic creates a new topic in the kafka cluster.
func CreateTopic(ctx context.Context, client *kafka.AdminClient, log logger.Logger, topic string) (kafka.TopicResult, error) {
	result, err := CreateTopics(ctx, client, log, []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	}})
	if (result == nil) || (len(result) == 0) {
		return kafka.TopicResult{}, err
	}
	return result[0], err
}

// CreateTopics creates multiple topics in the kafka cluster at once.
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

// DeleteTopic deletes a topic from the kafka cluster.
func DeleteTopic(ctx context.Context, client *kafka.AdminClient, log logger.Logger, topic string) error {
	_, err := DeleteTopics(ctx, client, log, []string{topic})
	return err
}

// DeleteTopics deletes multiple topics from the kafka cluster at once.
func DeleteTopics(ctx context.Context, client *kafka.AdminClient, log logger.Logger, topics []string) ([]kafka.TopicResult, error) {
	result, err := client.DeleteTopics(
		ctx,
		topics,
		kafka.SetAdminOperationTimeout(config.AdminOperationTimeout),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}
