package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/worker/placement"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
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
	placementService placement.PlacementService
	lock             sync.Mutex
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(placementService placement.PlacementService, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:         map[string]func(msg *kafka.Message){},
		placementService: placementService,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionInitializationResponsesTopic messages.
	m.add(naming.MessagingFunctionInitializationResponsesTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionInitializationResponseMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}
		functionIdentifier := naming.FunctionIdentifier(message.Function.Uuid, message.Function.Version)

		if !message.Success {
			log.Errorf("function initialization failed: %s", message.Reason)
			return
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := m.placementService.AddInitializedFunction(ctx, message.WorkerUuid, functionIdentifier); err != nil {
			log.Errorf("failed to add initialized function: %v", err)
			return
		}
		log.Infof("function %s initialized on worker %s", functionIdentifier, message.WorkerUuid)
	})

	// Handle MessagingFunctionDeinitializationRequestsTopic messages.
	m.add(naming.MessagingFunctionDeinitializationRequestsTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionDeinitializationRequestMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message into struct: %v", err)
			return
		}
		functionIdentifier := naming.FunctionIdentifier(message.Function.Uuid, message.Function.Version)

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := m.placementService.PerformFunctionDeinitialization(ctx, message.Function); err != nil {
			log.Errorf("failed to perform function deinitialization: %v", err)
			return
		}
		log.Infof("function %s is beeing deinitialized on all workers", functionIdentifier)
	})

	// Handle MessagingFunctionDeinitializationResponsesTopic messages.
	m.add(naming.MessagingFunctionDeinitializationResponsesTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionDeinitializationResponseMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message into struct: %v", err)
			return
		}
		functionIdentifier := naming.FunctionIdentifier(message.Function.Uuid, message.Function.Version)

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := m.placementService.RemoveInitializedFunction(ctx, message.WorkerUuid, functionIdentifier); err != nil {
			log.Errorf("failed to remove initialized function: %v", err)
		}
		log.Infof("function %s deinitialized on worker %s", functionIdentifier, message.WorkerUuid)
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
