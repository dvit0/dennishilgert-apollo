package registry

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/metrics"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger("apollo.registry")

type Options struct {
	Address           string
	HeartbeatInterval time.Duration
	InstanceUuid      string
	InstanceType      registrypb.InstanceType
}

type ServiceRegistryClient interface {
	AcquireLease(ctx context.Context, request *registrypb.AcquireLeaseRequest) error
	ReleaseLease(ctx context.Context) error
	AvailableServiceInstance(ctx context.Context, instanceType registrypb.InstanceType) (*registrypb.ServiceInstance, error)
	SendHeartbeat(ctx context.Context) error
	EstablishConnection(ctx context.Context) error
	CloseConnection()
}

type serviceRegistryClient struct {
	address           string
	heartbeatInterval time.Duration
	instanceUuid      string
	instanceType      registrypb.InstanceType
	metricsService    metrics.MetricsService
	messagingProducer producer.MessagingProducer
	clientConn        *grpc.ClientConn
}

// NewServiceRegistryClient creates a new ServiceRegistryClient instance.
func NewServiceRegistryClient(metricsService metrics.MetricsService, messagingProducer producer.MessagingProducer, opts Options) ServiceRegistryClient {
	return &serviceRegistryClient{
		address:           opts.Address,
		heartbeatInterval: opts.HeartbeatInterval,
		instanceUuid:      opts.InstanceUuid,
		instanceType:      opts.InstanceType,
		metricsService:    metricsService,
		messagingProducer: messagingProducer,
	}
}

// AcquireLease acquires a lease for the instance.
func (s *serviceRegistryClient) AcquireLease(ctx context.Context, request *registrypb.AcquireLeaseRequest) error {
	if !s.connectionAlive() {
		return fmt.Errorf("connection dead - failed to aquire lease")
	}
	apiClient := registrypb.NewServiceRegistryClient(s.clientConn)
	_, err := apiClient.AcquireLease(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to acquire lease: %w", err)
	}
	return nil
}

// ReleaseLease releases the lease for the instance.
func (s *serviceRegistryClient) ReleaseLease(ctx context.Context) error {
	if !s.connectionAlive() {
		return fmt.Errorf("connection dead - failed to release lease")
	}
	apiClient := registrypb.NewServiceRegistryClient(s.clientConn)
	_, err := apiClient.ReleaseLease(ctx, &registrypb.ReleaseLeaseRequest{
		InstanceUuid: s.instanceUuid,
		InstanceType: s.instanceType,
	})
	if err != nil {
		return fmt.Errorf("failed to release lease: %w", err)
	}
	return nil
}

// AvailableServiceInstance returns an available service instance of the specified type.
func (s *serviceRegistryClient) AvailableServiceInstance(ctx context.Context, instanceType registrypb.InstanceType) (*registrypb.ServiceInstance, error) {
	if !s.connectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to get available service instance")
	}
	apiClient := registrypb.NewServiceRegistryClient(s.clientConn)
	response, err := apiClient.AvailableServiceInstance(ctx, &registrypb.AvailableInstanceRequest{
		InstanceType: instanceType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance: %w", err)
	}
	return response.Instance, nil
}

// SendHeartbeat sends a heartbeat message in the specified interval via messaging.
func (s *serviceRegistryClient) SendHeartbeat(ctx context.Context) error {
	ticker := time.NewTicker(s.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var heartbeatMessage messagespb.InstanceHeartbeatMessage
			if s.instanceType == registrypb.InstanceType_FLEET_MANAGER {
				heartbeatMessage = messagespb.InstanceHeartbeatMessage{
					InstanceUuid: s.instanceUuid,
					InstanceType: s.instanceType,
					Metrics: &messagespb.InstanceHeartbeatMessage_WorkerInstanceMetrics{
						WorkerInstanceMetrics: s.metricsService.WorkerInstanceMetrics(),
					},
				}
			} else {
				heartbeatMessage = messagespb.InstanceHeartbeatMessage{
					InstanceUuid: s.instanceUuid,
					InstanceType: s.instanceType,
					Metrics: &messagespb.InstanceHeartbeatMessage_ServiceInstanceMetrics{
						ServiceInstanceMetrics: s.metricsService.ServiceInstanceMetrics(),
					},
				}
			}
			s.messagingProducer.Publish(ctx, naming.MessagingInstanceHeartbeatTopic, &heartbeatMessage)
		}
	}
}

// EstablishConnection establishes a connection to the service registry.
func (s *serviceRegistryClient) EstablishConnection(ctx context.Context) error {
	log.Info("establishing a connection to the service registry")

	if s.clientConn != nil {
		if s.connectionAlive() {
			log.Debug("connection alive - no need to establish a new connection to the service registry")
			return nil
		}
		// Try to close the existing connection and ignore the possible error.
		s.clientConn.Close()
	}

	log.Debugf("connecting to service registry with address: %s", s.address)

	const retrySeconds = 3     // trying to connect for a period of 3 seconds
	const retriesPerSecond = 2 // trying to connect 2 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, s.address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			s.clientConn = conn
			log.Info("connection to service registry established")
			return nil
		} else {
			if conn != nil {
				conn.Close()
			}
			log.Errorf("failed to establish connection to service registry - reason: %v", err)
		}
		// Wait before retrying, but stop if context is done.
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done before connection to servie registry could be established: %w", ctx.Err())
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return fmt.Errorf("failed to establish connection to service registry after %d seconds", retrySeconds)
}

// CloseConnection closes the grpc client connection.
func (s *serviceRegistryClient) CloseConnection() {
	if s.clientConn != nil {
		s.clientConn.Close()
	}
}

// connectionAlive returns if a grpc client connection is still alive.
func (s *serviceRegistryClient) connectionAlive() bool {
	if s.clientConn == nil {
		return false
	}
	return (s.clientConn.GetState() == connectivity.Ready) || (s.clientConn.GetState() == connectivity.Idle)
}
