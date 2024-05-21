package messaging

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/logs/operator"
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
	handlers     map[string]func(msg *kafka.Message)
	lock         sync.Mutex
	logsOperator operator.LogsOperator
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(logsOperator operator.LogsOperator, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:     map[string]func(msg *kafka.Message){},
		logsOperator: logsOperator,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	m.add(naming.MessagingFunctionInvocationLogsTopic, func(msg *kafka.Message) {
		var message messagespb.FunctionInvocationLogsMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := m.logsOperator.StoreInvocationLogs(ctx, &message); err != nil {
			log.Errorf("failed to store invocation logs: %v", err)
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
