package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/frontend/operator"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
)

var log = logger.NewLogger("apollo.messaging.handler")

type Options struct{}

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers         map[string]func(msg *kafka.Message)
	lock             sync.Mutex
	frontendOperator operator.FrontendOperator
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(frontendOperator operator.FrontendOperator, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:         map[string]func(msg *kafka.Message){},
		frontendOperator: frontendOperator,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionStatusUpdateTopic messages.
	m.add(naming.MessagingFunctionStatusUpdateTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionStatusUpdateMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := m.frontendOperator.UpdateFunctionStatus(ctx, message.Function.Uuid, message.Status); err != nil {
			log.Errorf("failed to update function status: %v", err)
		}

		switch message.Status {
		case frontendpb.FunctionStatus_PACKED:
			function, err := m.frontendOperator.GetFunction(ctx, &frontendpb.GetFunctionRequest{
				Uuid: message.Function.Uuid,
			})
			if err != nil {
				log.Errorf("failed to get function: %v", err)
				return
			}
			m.frontendOperator.DeinitializeFunction(ctx, &fleetpb.FunctionSpecs{
				Uuid:    function.Function.Uuid,
				Version: function.Function.Version,
			})
			if err := m.frontendOperator.UpdateFunctionVersion(ctx, message.Function); err != nil {
				log.Errorf("failed to update function version: %v", err)
				return
			}
			if err := m.frontendOperator.InitializeFunction(ctx, message.Function.Uuid, message.Function.Version); err != nil {
				log.Errorf("failed to initialize function: %v", err)
			}
			if err := m.frontendOperator.UpdateFunctionStatus(ctx, message.Function.Uuid, frontendpb.FunctionStatus_INITIALIZING); err != nil {
				log.Errorf("failed to update function status: %v", err)
			}
		case frontendpb.FunctionStatus_PACKING_FAILED:
			log.Errorf("function packing failed: %s", message.Reason)
		}
	})

	m.add(naming.MessagingFunctionInitializationResponsesTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionInitializationResponseMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if !message.Success {
			log.Errorf("function initialization failed: %s", message.Reason)
			if err := m.frontendOperator.UpdateFunctionStatus(ctx, message.Function.Uuid, frontendpb.FunctionStatus_INITIALIZAION_FAILED); err != nil {
				log.Errorf("failed to update function status: %v", err)
			}
			return
		}

		if err := m.frontendOperator.UpdateFunctionStatus(ctx, message.Function.Uuid, frontendpb.FunctionStatus_READY); err != nil {
			log.Errorf("failed to update function status: %v", err)
		}
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
