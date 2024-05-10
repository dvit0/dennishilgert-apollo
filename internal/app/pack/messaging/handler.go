package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/dennishilgert/apollo/internal/app/pack/operator"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	"github.com/dennishilgert/apollo/internal/pkg/storage"
)

var log = logger.NewLogger("apollo.messaging.handler")

type Options struct{}

type MessagingHandler interface {
	RegisterAll()
	Handlers() map[string]func(msg *kafka.Message)
}

type messagingHandler struct {
	handlers          map[string]func(msg *kafka.Message)
	lock              sync.Mutex
	packageOperator   operator.PackageOperator
	messagingProducer producer.MessagingProducer
}

// NewMessagingHandler creates a new MessagingHandler instance.
func NewMessagingHandler(packageOperator operator.PackageOperator, messagingProducer producer.MessagingProducer, opts Options) MessagingHandler {
	return &messagingHandler{
		handlers:          map[string]func(msg *kafka.Message){},
		packageOperator:   packageOperator,
		messagingProducer: messagingProducer,
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
		functionUuid, functionVersion, err := extractFunctionDetails(&message)
		if err != nil {
			log.Errorf("failed to extract function details: %v", err)
		}

		// Create a new context with a timeout of 3 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		m.messagingProducer.Publish(ctx, naming.MessagingFunctionStatusUpdateTopic, &messagespb.FunctionStatusUpdateMessage{
			Function: &fleetpb.FunctionSpecs{
				Uuid:    functionUuid,
				Version: functionVersion,
			},
			Status: frontendpb.FunctionStatus_PACKING,
			Reason: "function code uploaded",
		})

		m.packageOperator.BuildPackage(&fleetpb.FunctionSpecs{
			Uuid:    functionUuid,
			Version: functionVersion,
		})
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

func extractFunctionDetails(event *storage.MinioObjectUploadedEvent) (string, string, error) {
	pattern := fmt.Sprintf(
		`%s\/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})\/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})\.zip`,
		naming.StorageFunctionBucketName,
	)
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(event.Key)

	if len(matches) < 3 {
		return "", "", errors.New("input does not match the required format")
	}
	// matches[0] is the entire match, matches[1] is the function uuid, matches[2] is the function version.
	return matches[1], matches[2], nil
}
