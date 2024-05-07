package messaging

import (
	"encoding/json"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/pack/operator"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	messagespb "github.com/dennishilgert/apollo/pkg/proto/messages/v1"
	"github.com/dennishilgert/apollo/pkg/storage"
)

var log = logger.NewLogger("apollo.messaging.handler")

type Options struct{}

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers        map[string]func(msg *kafka.Message)
	lock            sync.Mutex
	packageOperator operator.PackageOperator
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(packageOperator operator.PackageOperator, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:        map[string]func(msg *kafka.Message){},
		packageOperator: packageOperator,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionCodeUploadedTopic messages.
	m.add(naming.MessagingFunctionCodeUploadedTopic, func(msg *kafka.Message) {
		var message storage.MinioObjectUploadedEvent
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}
		log.Infof("received message: %s", message)
	})

	// Handle MessagingFunctionPackageCreationTopic messages.
	m.add(naming.MessagingFunctionPackageCreationTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionPackageCreationMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}

		m.packageOperator.BuildPackage(message.Function, message.Runtime)
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
