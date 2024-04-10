package messaging

import (
	"context"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging"
)

var log = logger.NewLogger("apollo.manager.messaging")

func CreateRelatedTopic(ctx context.Context, bootstrapServers string, managerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		log.Errorf("failed to get default kafka admin client")
		return err
	}
	topic := naming.MessagingManagerRelatedAgentReadyTopic(managerUuid)
	_, err = messaging.CreateTopic(ctx, adminClient, log, topic)
	if err != nil {
		log.Errorf("failed to create manager related topic: %s", topic)
		return err
	}
	return nil
}

func DeleteRelatedTopic(ctx context.Context, bootstrapServers string, managerUuid string) error {
	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		log.Errorf("failed to get default kafka admin client")
		return err
	}
	topic := naming.MessagingManagerRelatedAgentReadyTopic(managerUuid)
	_, err = messaging.DeleteTopic(ctx, adminClient, log, topic)
	if err != nil {
		log.Errorf("failed to delete manager related topic: %s", topic)
		return err
	}
	return nil
}
