package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/registry/cache"
	"github.com/dennishilgert/apollo/internal/app/registry/scoring"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	messagespb "github.com/dennishilgert/apollo/pkg/proto/messages/v1"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

var log = logger.NewLogger("apollo.registry.messaging.handler")

type Options struct{}

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers    map[string]func(msg *kafka.Message)
	lock        sync.Mutex
	cacheClient cache.CacheClient
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(cacheClient cache.CacheClient, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:    map[string]func(msg *kafka.Message){},
		cacheClient: cacheClient,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionInitializationTopic messages
	m.add(naming.MessagingInstanceHeartbeatTopic, func(msg *kafka.Message) {
		var message messagespb.InstanceHeartbeatMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		var key string

		switch message.ServiceType {
		case registrypb.ServiceType_FLEET_MANAGER:
			key = naming.CacheWorkerInstanceKeyName(message.InstanceUuid)
		default:
			key = naming.CacheServiceInstanceKeyName(message.InstanceUuid)
			scoringResult := scoring.CalculateScore(&message)
			if err := m.cacheClient.UpdateScore(ctx, scoringResult); err != nil {
				log.Errorf("failed to update score: %v", err)
			}
		}

		if err := m.cacheClient.ExtendExpiration(ctx, key); err != nil {
			log.Errorf("failed to extend expiration: %v", err)
		}
		// Cancel context after the score has been updated.
		cancel()
	})
}

// Handlers returns a map containing all handlers specified by the corresponding topic.
func (m *messagingHandler) Handlers() map[string]func(msg *kafka.Message) {
	return m.handlers
}

// add adds a handler as value and the topic as key to the handler map.
func (m *messagingHandler) add(topic string, handler func(msg *kafka.Message)) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.handlers[topic] = handler
}
