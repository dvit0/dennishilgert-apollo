package messaging

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging"
)

var log = logger.NewLogger("apollo.manager.messaging")

// CreateRelatedTopic creates the related topic for the worker.
func CreateRelatedTopic(ctx context.Context, bootstrapServers string, workerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		return fmt.Errorf("failed to get default kafka admin client: %w", err)
	}
	topic := naming.MessagingWorkerRelatedAgentReadyTopic(workerUuid)
	_, err = messaging.CreateTopic(ctx, adminClient, log, topic)
	if err != nil {
		return fmt.Errorf("failed to create manager related topic: %w", err)
	}
	return nil
}

// DeleteRelatedTopic deletes the related topic for the worker.
func DeleteRelatedTopic(ctx context.Context, bootstrapServers string, workerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		return fmt.Errorf("failed to get default kafka admin client: %w", err)
	}
	topic := naming.MessagingWorkerRelatedAgentReadyTopic(workerUuid)
	if err := messaging.DeleteTopic(ctx, adminClient, log, topic); err != nil {
		return fmt.Errorf("failed to delete manager related topic: %w", err)
	}
	return nil
}
