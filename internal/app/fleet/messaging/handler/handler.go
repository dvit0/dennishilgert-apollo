package handler

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/fleet/operator"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/messages/v1"
)

var log = logger.NewLogger("apollo.manager.messaging.handler")

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers       map[string]func(msg *kafka.Message)
	lock           sync.Mutex
	runnerOperator operator.RunnerOperator
}

func NewMessagingHandler(runnerOperator operator.RunnerOperator) MessagingHandler {
	return &messagingHandler{
		handlers:       map[string]func(msg *kafka.Message){},
		runnerOperator: runnerOperator,
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handle MessagingFunctionInitializationTopic messages
	m.add(naming.MessagingFunctionInitializationTopic, func(msg *kafka.Message) {
		var message messages.FunctionInitializationMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}
		log.Infof("NOT IMPLEMENTED: handling message in topic: %s - value: %v", *msg.TopicPartition.Topic, &message)
	})

	// Handle MessagingRunnerAgentReadyTopic messages
	m.add(naming.MessagingRunnerAgentReadyTopic, func(msg *kafka.Message) {
		var message messages.RunnerAgentReadyMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}
		instance, err := m.runnerOperator.Runner(message.FunctionUuid, message.RunnerUuid)
		if err != nil {
			log.Warnf("requested runner instance does not exist in pool: %s", message.RunnerUuid)
		}
		var mErr error
		if !message.Success {
			mErr = fmt.Errorf(message.Reason)
		}
		instance.AgentReady(mErr)
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
