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
		scoringResult := scoring.CalculateScore(&message)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		m.cacheClient.UpdateScore(ctx, scoringResult)
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
