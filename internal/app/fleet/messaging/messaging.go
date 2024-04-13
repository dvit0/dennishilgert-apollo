package messaging

import (
	"context"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging"
)

var log = logger.NewLogger("apollo.manager.messaging")

// CreateRelatedTopic creates the related topic for the worker.
func CreateRelatedTopic(ctx context.Context, bootstrapServers string, workerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		log.Errorf("failed to get default kafka admin client")
		return err
	}
	topic := naming.MessagingWorkerRelatedAgentReadyTopic(workerUuid)
	_, err = messaging.CreateTopic(ctx, adminClient, log, topic)
	if err != nil {
		log.Errorf("failed to create manager related topic: %s", topic)
		return err
	}
	return nil
}

// DeleteRelatedTopic deletes the related topic for the worker.
func DeleteRelatedTopic(ctx context.Context, bootstrapServers string, workerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		log.Errorf("failed to get default kafka admin client")
		return err
	}
	topic := naming.MessagingWorkerRelatedAgentReadyTopic(workerUuid)
	if err := messaging.DeleteTopic(ctx, adminClient, log, topic); err != nil {
		log.Errorf("failed to delete manager related topic: %v", err)
		return err
	}
	return nil
}
