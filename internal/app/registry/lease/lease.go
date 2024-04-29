package lease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dennishilgert/apollo/internal/app/registry/scoring"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/cache"
	"github.com/dennishilgert/apollo/pkg/logger"
	messagespb "github.com/dennishilgert/apollo/pkg/proto/messages/v1"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

var log = logger.NewLogger("apollo.registry.lease")

type Options struct {
	LeaseTimeout time.Duration
}

type LeaseService interface {
	Listen(ctx context.Context) error
	AquireLease(ctx context.Context, request *registrypb.AcquireLeaseRequest) error
	RenewLease(ctx context.Context, request *messagespb.InstanceHeartbeatMessage) error
	ReleaseLease(ctx context.Context, instanceUuid string, instanceType string) error
	AvailableServiceInstance(ctx context.Context, instanceType registrypb.InstanceType) (*registrypb.ServiceInstance, error)
}

type leaseService struct {
	leaseTimeout time.Duration
	cacheClient  cache.CacheClient
}

// NewLeaseService creates a new LeaseService instance.
func NewLeaseService(cacheClient cache.CacheClient, opts Options) LeaseService {
	return &leaseService{
		leaseTimeout: opts.LeaseTimeout,
		cacheClient:  cacheClient,
	}
}

// Listen listens for expired keys in the cache and releases the lease for the corresponding instance.
func (l *leaseService) Listen(ctx context.Context) error {
	pubsub := l.cacheClient.Client().PSubscribe(ctx, "__keyevent@0__:expired")

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down lease listener")
			pubsub.Close()
			return nil
		case msg := <-pubsub.Channel():
			if msg == nil {
				return nil
			}
			if !naming.CacheIsLeaseKey(msg.Payload) {
				continue
			}
			log.Warnf("key in cache expired: %s", msg.Payload)

			rawKey := naming.CacheStripLeaseDeclaration(msg.Payload)
			instanceUuid := naming.CacheExtractInstanceUuid(rawKey)
			var instanceType string
			if naming.CacheIsWorkerInstanceLease(rawKey) {
				instanceType = registrypb.InstanceType_FLEET_MANAGER.String()
			} else if naming.CacheIsServiceInstanceLease(rawKey) {
				instanceType = naming.CacheExtractServiceInstanceType(rawKey)
			}
			if err := l.ReleaseLease(ctx, instanceUuid, instanceType); err != nil {
				log.Errorf("failed to release lease: %v", err)
			}
			log.Debugf("lease released for instance: %s", instanceUuid)
		}
	}
}

// AquireLease acquires a lease for the given instance.
func (l *leaseService) AquireLease(ctx context.Context, request *registrypb.AcquireLeaseRequest) error {
	switch instance := request.Instance.(type) {
	case *registrypb.AcquireLeaseRequest_WorkerInstance:
		log.Debugf("acquiring lease for worker instance: %s", instance.WorkerInstance.WorkerUuid)

		if err := l.addWorkerInstance(ctx, instance.WorkerInstance, request.GetWorkerInstanceMetrics()); err != nil {
			return fmt.Errorf("failed to add worker instance to cache: %w", err)
		}
		return nil
	case *registrypb.AcquireLeaseRequest_ServiceInstance:
		log.Debugf("acquiring lease for service instance: %s", instance.ServiceInstance.InstanceUuid)

		if err := l.addServiceInstance(ctx, instance.ServiceInstance); err != nil {
			return fmt.Errorf("failed to add service instance to cache: %w", err)
		}
		metrics := request.Metrics.(*registrypb.AcquireLeaseRequest_ServiceInstanceMetrics)
		score := scoring.CalculateScore(metrics.ServiceInstanceMetrics)
		if err := l.setServiceInstanceScore(ctx, instance.ServiceInstance.InstanceUuid, instance.ServiceInstance.InstanceType.String(), score); err != nil {
			return fmt.Errorf("failed to add score for service instance: %w", err)
		}
		return nil
	default:
		return errors.New("unknown instance type")
	}
}

// RenewLease renews the lease for the given instance.
func (l *leaseService) RenewLease(ctx context.Context, request *messagespb.InstanceHeartbeatMessage) error {
	log.Debugf("renewing lease for instance: %s", request.InstanceUuid)

	var key string
	switch instanceMetrics := request.Metrics.(type) {
	case *messagespb.InstanceHeartbeatMessage_WorkerInstanceMetrics:
		key = naming.CacheWorkerInstanceKeyName(request.InstanceUuid)
		if err := l.setWokerInstanceMetrics(ctx, request.InstanceUuid, instanceMetrics.WorkerInstanceMetrics); err != nil {
			return fmt.Errorf("failed to set metrics for worker instance: %w", err)
		}
	case *messagespb.InstanceHeartbeatMessage_ServiceInstanceMetrics:
		key = naming.CacheServiceInstanceKeyName(request.InstanceUuid)
		score := scoring.CalculateScore(instanceMetrics.ServiceInstanceMetrics)
		if err := l.setServiceInstanceScore(ctx, request.InstanceUuid, request.InstanceType.String(), score); err != nil {
			return fmt.Errorf("failed to update score for service instance: %w", err)
		}
	default:
		return errors.New("unknown metrics type")
	}

	if err := l.cacheClient.Client().Expire(ctx, key, l.leaseTimeout).Err(); err != nil {
		return fmt.Errorf("failed to extend lease expiration time for instance: %w", err)
	}
	return nil
}

// ReleaseLease releases the lease for the given instance.
func (l *leaseService) ReleaseLease(ctx context.Context, instanceUuid string, instanceType string) error {
	log.Debugf("releasing lease for instance: %s", instanceUuid)

	switch instanceType {
	case registrypb.InstanceType_FLEET_MANAGER.String():
		if err := l.removeWorkerInstance(ctx, instanceUuid); err != nil {
			return fmt.Errorf("failed to remove worker instance from cache: %w", err)
		}
	default:
		if err := l.removeServiceInstance(ctx, instanceUuid); err != nil {
			return fmt.Errorf("failed to remove service instance from cache: %w", err)
		}
	}
	return nil
}

// AvailableServiceInstance returns the next available service instance by type.
func (l *leaseService) AvailableServiceInstance(ctx context.Context, instanceType registrypb.InstanceType) (*registrypb.ServiceInstance, error) {
	result, err := l.cacheClient.Client().ZRevRangeWithScores(ctx, instanceType.String(), 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get next instance by type: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no instances found for service type %s", instanceType)
	}
	instanceUuid := result[0].Member.(string)
	key := naming.CacheServiceInstanceKeyName(instanceUuid)
	values, err := l.cacheClient.Client().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance from cache: %w", err)
	}
	port, _ := strconv.Atoi(values["port"])
	var metadata map[string]interface{}
	json.Unmarshal([]byte(values["metadata"]), &metadata)

	// Convert metadata to *structpb.Struct
	metadataStruct, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to convert metadata to struct: %w", err)
	}

	return &registrypb.ServiceInstance{
		InstanceUuid: instanceUuid,
		InstanceType: instanceType,
		Host:         values["host"],
		Port:         int32(port),
		Metadata:     metadataStruct,
	}, nil
}

// addServiceInstance adds a service instance to the cache.
func (l *leaseService) addServiceInstance(ctx context.Context, instance *registrypb.ServiceInstance) error {
	key := naming.CacheServiceInstanceKeyName(instance.InstanceUuid)

	// Serialize metadata.
	metadataBytes, err := json.Marshal(instance.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := l.cacheClient.Client().HSet(ctx, key, map[string]interface{}{
		"type":     instance.InstanceType.String(),
		"host":     instance.Host,
		"port":     instance.Port,
		"metadata": string(metadataBytes),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add service instance to cache: %w", err)
	}
	if err := l.cacheClient.Client().Set(ctx, naming.CacheAddLeaseDeclaration(key), "active", l.leaseTimeout).Err(); err != nil {
		return fmt.Errorf("failed to add service instance lease to cache: %w", err)
	}
	return nil
}

// removeServiceInstance removes a service instance from the cache.
func (l *leaseService) removeServiceInstance(ctx context.Context, instanceUuid string) error {
	key := naming.CacheServiceInstanceKeyName(instanceUuid)
	values, err := l.cacheClient.Client().HGetAll(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get service instance from cache: %w", err)
	}
	if err := l.cacheClient.Client().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove service instance from cache: %w", err)
	}
	if err := l.cacheClient.Client().Del(ctx, naming.CacheAddLeaseDeclaration(key)).Err(); err != nil {
		return fmt.Errorf("failed to remove service instance lease from cache: %w", err)
	}
	if err := l.cacheClient.Client().ZRem(ctx, values["type"], instanceUuid).Err(); err != nil {
		return fmt.Errorf("failed to remove service instance from sorted set: %w", err)
	}
	return nil
}

func (l *leaseService) setServiceInstanceScore(ctx context.Context, instanceUuid string, instanceType string, score float64) error {
	if err := l.cacheClient.Client().ZAdd(ctx, instanceType, redis.Z{
		Score:  score,
		Member: instanceUuid,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add service instance score: %w", err)
	}
	return nil
}

// addWorkerInstance adds a worker instance to the cache.
func (l *leaseService) addWorkerInstance(ctx context.Context, instance *registrypb.WorkerInstance, metrics *registrypb.WorkerInstanceMetrics) error {
	key := naming.CacheWorkerInstanceKeyName(instance.WorkerUuid)

	// Serialize initialized functions.
	functionsBytes, err := json.Marshal(instance.InitializedFunctions)
	if err != nil {
		return fmt.Errorf("failed to marshal functions: %w", err)
	}
	// Serialize metrics.
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	if err := l.cacheClient.Client().HSet(ctx, key, map[string]interface{}{
		"architecture": instance.Architecture,
		"host":         instance.Host,
		"port":         instance.Port,
		"functions":    string(functionsBytes),
		"metrics":      string(metricsBytes),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add worker instance to cache: %w", err)
	}
	if err := l.cacheClient.Client().Set(ctx, naming.CacheAddLeaseDeclaration(key), "active", l.leaseTimeout).Err(); err != nil {
		return fmt.Errorf("failed to add worker instance lease to cache: %w", err)
	}
	if err := l.cacheClient.Client().SAdd(ctx, naming.CacheArchitectureSetKey(instance.Architecture), instance.WorkerUuid).Err(); err != nil {
		return fmt.Errorf("failed to add worker instance to architecture set: %w", err)
	}
	for _, function := range instance.InitializedFunctions {
		if err := l.cacheClient.Client().SAdd(ctx, naming.CacheFunctionSetKey(function.Uuid), instance.WorkerUuid).Err(); err != nil {
			return fmt.Errorf("failed to add worker instance to function set: %w", err)
		}
	}
	return nil
}

// removeWorkerInstance removes a worker instance from the cache.
func (l *leaseService) removeWorkerInstance(ctx context.Context, workerUuid string) error {
	key := naming.CacheWorkerInstanceKeyName(workerUuid)
	values, err := l.cacheClient.Client().HGetAll(ctx, workerUuid).Result()
	if err != nil {
		return fmt.Errorf("failed to get worker instance from cache: %w", err)
	}
	if err := l.cacheClient.Client().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove worker instance from cache: %w", err)
	}
	if err := l.cacheClient.Client().Del(ctx, naming.CacheAddLeaseDeclaration(key)).Err(); err != nil {
		return fmt.Errorf("failed to remove worker instance lease from cache: %w", err)
	}
	if err := l.cacheClient.Client().SRem(ctx, naming.CacheArchitectureSetKey(values["architecture"]), workerUuid).Err(); err != nil {
		return fmt.Errorf("failed to remove worker instance from architecture set: %w", err)
	}
	initializedFunctions := make([]*registrypb.Function, 0)
	if err := json.Unmarshal([]byte(values["functions"]), &initializedFunctions); err != nil {
		return fmt.Errorf("failed to unmarshal functions: %w", err)
	}
	for _, function := range initializedFunctions {
		if err := l.cacheClient.Client().SRem(ctx, naming.CacheFunctionSetKey(function.Uuid), workerUuid).Err(); err != nil {
			return fmt.Errorf("failed to remove worker instance from function set: %w", err)
		}
	}
	return nil
}

func (l *leaseService) setWokerInstanceMetrics(ctx context.Context, workerUuid string, metrics *registrypb.WorkerInstanceMetrics) error {
	key := naming.CacheWorkerInstanceKeyName(workerUuid)

	// Serialize metrics.
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}
	if err := l.cacheClient.Client().HSet(ctx, key, map[string]interface{}{
		"metrics": string(metricsBytes),
	}).Err(); err != nil {
		return fmt.Errorf("failed to set worker instance metrics in cache: %w", err)
	}
	return nil
}
