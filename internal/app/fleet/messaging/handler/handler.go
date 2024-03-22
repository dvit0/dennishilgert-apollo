package handler

import (
	"encoding/json"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/fleet/preparer"
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
	runnerPreparer preparer.RunnerPreparer
	handlers       map[string]func(msg *kafka.Message)
	lock           sync.Mutex
}

func NewMessagingHandler(runnerPreparer preparer.RunnerPreparer) MessagingHandler {
	return &messagingHandler{
		runnerPreparer: runnerPreparer,
		handlers:       map[string]func(msg *kafka.Message){},
	}
}

// RegisterAll registrates all handlers for the subscribed topics in the handler map.
func (m *messagingHandler) RegisterAll() {
	// Handling MessagingFunctionInitializationTopic messages
	m.add(naming.MessagingFunctionStatusUpdateTopic, func(msg *kafka.Message) {
		var message messages.FunctionInitialized
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Errorf("failed to unmarshal kafka message: %v", err)
		}
		log.Infof("NOT IMPLEMENTED: handling message in topic: %s - value: %v", *msg.TopicPartition.Topic, &message)
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
