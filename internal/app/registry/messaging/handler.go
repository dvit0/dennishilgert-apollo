package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/registry/lease"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
)

var log = logger.NewLogger("apollo.registry.messaging.handler")

type TempInstanceHeartbeatMessage struct {
	InstanceUuid string                  `json:"instance_uuid"`
	InstanceType registrypb.InstanceType `json:"instance_type"`
	Metrics      json.RawMessage         `json:"Metrics"`
}

type Options struct{}

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers     map[string]func(msg *kafka.Message)
	lock         sync.Mutex
	leaseService lease.LeaseService
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(leaseService lease.LeaseService, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:     map[string]func(msg *kafka.Message){},
		leaseService: leaseService,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionInitializationTopic messages
	m.add(naming.MessagingInstanceHeartbeatTopic, func(msg *kafka.Message) {
		var tempMessage TempInstanceHeartbeatMessage
		// Unmarshal the message into a temporary struct to determine the type of metrics.
		if err := json.Unmarshal(msg.Value, &tempMessage); err != nil {
			log.Errorf("failed to unmarshal kafka message into temp struct: %v", err)
			return
		}
		// Copy the instanceUuid and instanceType into the real message.
		message := &messagespb.InstanceHeartbeatMessage{
			InstanceUuid: tempMessage.InstanceUuid,
			InstanceType: tempMessage.InstanceType,
		}

		// Determine the type of metrics and unmarshal it into the correct type.
		if message.InstanceType == registrypb.InstanceType_FLEET_MANAGER {
			var workerInstanceMetrics messagespb.InstanceHeartbeatMessage_WorkerInstanceMetrics
			if err := json.Unmarshal(tempMessage.Metrics, &workerInstanceMetrics); err != nil {
				log.Errorf("failed to unmarshal worker instance metrics: %v", err)
				return
			}
			message.Metrics = &workerInstanceMetrics
		} else {
			var serviceInstanceMetrics messagespb.InstanceHeartbeatMessage_ServiceInstanceMetrics
			if err := json.Unmarshal(tempMessage.Metrics, &serviceInstanceMetrics); err != nil {
				log.Errorf("failed to unmarshal service instance metrics: %v", err)
				return
			}
			message.Metrics = &serviceInstanceMetrics
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := m.leaseService.RenewLease(ctx, message); err != nil {
			log.Errorf("failed to renew lease: %v", err)
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
