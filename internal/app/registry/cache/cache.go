package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dennishilgert/apollo/internal/app/registry/scoring"
	"github.com/dennishilgert/apollo/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

var log = logger.NewLogger("apollo.registry.cache")

type Options struct {
	Address           string
	Username          string
	Password          string
	Database          int
	ExpirationTimeout time.Duration
}

type CacheClient interface {
	AttemptLeaderElection(ctx context.Context, instanceType string) (bool, error)
	AddInstance(ctx context.Context, serviceInstance *registrypb.ServiceInstance) error
	UpdateScore(ctx context.Context, scoringResult scoring.ScoringResult) error
	AddWorker(ctx context.Context, workerInstance *registrypb.WorkerInstance) error
	Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, error)
	WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error)
	WorkersByRuntime(ctx context.Context, runtimeName, runtimeVersion string) ([]string, error)
	RemoveWorker(ctx context.Context, workerUuid string) error
}

type cacheClient struct {
	instanceUuid      string
	expirationTimeout time.Duration
	client            *redis.Client
}

// NewCachingClient creates a new caching client.
func NewCacheClient(instanceUuid string, opts Options) CacheClient {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Address,
		Username: opts.Username,
		Password: opts.Password,
		DB:       opts.Database,
	})

	return &cacheClient{
		instanceUuid:      instanceUuid,
		expirationTimeout: opts.ExpirationTimeout,
		client:            client,
	}
}

// Listen listens for cache events.
func (c *cacheClient) Listen(ctx context.Context) error {
	pubsub := c.client.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down cache listener")
			return ctx.Err()
		case msg := <-pubsub.Channel():
			if msg == nil {
				return nil
			}
			log.Warnf("key in cache expired: %s", msg.Payload)
		}
	}
}

// Close closes the caching client.
func (c *cacheClient) Close() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close caching client: %w", err)
		}
	}
	return nil
}

// AttemptLeaderElection attempts to elect the current instance as the leader.
func (c *cacheClient) AttemptLeaderElection(ctx context.Context, forGroup string) (bool, error) {
	return c.client.SetNX(ctx, forGroup, c.instanceUuid, 5*time.Second).Result()
}

// AddInstance adds a service instance to the cache.
func (c *cacheClient) AddInstance(ctx context.Context, serviceInstance *registrypb.ServiceInstance) error {
	key := fmt.Sprintf("instance:%s", serviceInstance.InstanceUuid)

	// Serialize metadata
	metadataBytes, err := json.Marshal(serviceInstance.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	metadata := string(metadataBytes)

	if err := c.client.HSet(ctx, key, map[string]interface{}{
		"serviceType": serviceInstance.ServiceType.String(),
		"host":        serviceInstance.Host,
		"port":        serviceInstance.Port,
		"metadata":    metadata,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add service instance to cache: %w", err)
	}
	if err := c.client.ExpireGT(ctx, key, c.expirationTimeout).Err(); err != nil {
		return fmt.Errorf("failed to set expiration time for service instance: %w", err)
	}
	return nil
}

// UpdateScore updates the score of a service instance.
func (c *cacheClient) UpdateScore(ctx context.Context, scoringResult scoring.ScoringResult) error {
	if err := c.client.ZAdd(ctx, scoringResult.ServiceType.String(), redis.Z{
		Score:  scoringResult.Score,
		Member: scoringResult.InstanceUuid,
	}).Err(); err != nil {
		return fmt.Errorf("failed to update score for service instance: %w", err)
	}
	return nil
}

// TopInstanceByType returns the top service instance by type.
func (c *cacheClient) TopInstanceByType(ctx context.Context, serviceType registrypb.ServiceType) (*registrypb.ServiceInstance, error) {
	result, err := c.client.ZRevRangeWithScores(ctx, serviceType.String(), 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get next instance by type: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no instances found for service type %s", serviceType.String())
	}

	instanceUuid := result[0].Member.(string)
	key := fmt.Sprintf("instance:%s", instanceUuid)
	values, err := c.client.HGetAll(ctx, key).Result()
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
		ServiceType:  serviceType,
		Host:         values["host"],
		Port:         int32(port),
		Metadata:     metadataStruct,
	}, nil
}

// RemoveInstance removes a service instance from the cache.
func (c *cacheClient) RemoveInstance(ctx context.Context, serviceType registrypb.ServiceType, instanceUuid string) error {
	key := fmt.Sprintf("instance:%s", instanceUuid)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove instance from cache: %w", err)
	}
	if err := c.client.ZRem(ctx, serviceType.String(), instanceUuid).Err(); err != nil {
		return fmt.Errorf("failed to remove instance from sorted set: %w", err)
	}
	return nil
}

// AddWorker adds a worker node to the cache.
func (c *cacheClient) AddWorker(ctx context.Context, workerInstance *registrypb.WorkerInstance) error {
	runtimesJson, err := json.Marshal(workerInstance.InitializedRuntimes)
	if err != nil {
		return fmt.Errorf("failed to marshal runtimes: %w", err)
	}
	if err := c.client.HSet(ctx, workerInstance.WorkerUuid, map[string]interface{}{
		"architecture": workerInstance.Architecture,
		"host":         workerInstance.Host,
		"port":         workerInstance.Port,
		"runtimes":     string(runtimesJson),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add worker node to cache: %w", err)
	}
	if err := c.client.SAdd(ctx, fmt.Sprintf("arch:%s", workerInstance.Architecture), workerInstance.WorkerUuid).Err(); err != nil {
		return fmt.Errorf("failed to add worker node to architecture set: %w", err)
	}
	for _, runtime := range workerInstance.InitializedRuntimes {
		if err := c.client.SAdd(ctx, fmt.Sprintf("runtime:%s-%s", runtime.Name, runtime.Version), workerInstance.WorkerUuid).Err(); err != nil {
			return fmt.Errorf("failed to add worker node to runtime set: %w", err)
		}
	}
	return nil
}

// Worker returns a worker node from the cache.
func (c *cacheClient) Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, error) {
	values, err := c.client.HGetAll(ctx, workerUuid).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker node from cache: %w", err)
	}
	initializedRuntimes := make([]*registrypb.Runtime, 0)
	if err := json.Unmarshal([]byte(values["runtimes"]), &initializedRuntimes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal runtimes: %w", err)
	}
	port, _ := strconv.Atoi(values["port"])
	return &registrypb.WorkerInstance{
		WorkerUuid:          workerUuid,
		Architecture:        values["architecture"],
		Host:                values["host"],
		Port:                int32(port),
		InitializedRuntimes: initializedRuntimes,
	}, nil
}

// WorkersByArchitecture returns a list of worker nodes by architecture.
func (c *cacheClient) WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error) {
	result, err := c.client.SMembers(ctx, fmt.Sprintf("arch:%s", architecture)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker nodes by architecture: %w", err)
	}
	return result, nil
}

// WorkersByRuntime returns a list of worker nodes by runtime.
func (c *cacheClient) WorkersByRuntime(ctx context.Context, runtimeName, runtimeVersion string) ([]string, error) {
	result, err := c.client.SMembers(ctx, fmt.Sprintf("runtime:%s-%s", runtimeName, runtimeVersion)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker nodes by runtime: %w", err)
	}
	return result, nil
}

// RemoveWorker removes a worker node from the cache.
func (c *cacheClient) RemoveWorker(ctx context.Context, workerUuid string) error {
	result, err := c.client.HGetAll(ctx, workerUuid).Result()
	if err != nil {
		return fmt.Errorf("failed to get worker node from cache: %w", err)
	}
	keys := make([]string, 0, len(result))
	for key := range result {
		keys = append(keys, key)
	}
	if err := c.client.HDel(ctx, workerUuid, keys...).Err(); err != nil {
		return fmt.Errorf("failed to remove worker node from cache: %w", err)
	}
	if err := c.client.SRem(ctx, fmt.Sprintf("arch:%s", result["architecture"]), workerUuid).Err(); err != nil {
		return fmt.Errorf("failed to remove worker node from architecture set: %w", err)
	}
	initializedRuntimes := make([]*registrypb.Runtime, 0)
	if err := json.Unmarshal([]byte(result["runtimes"]), &initializedRuntimes); err != nil {
		return fmt.Errorf("failed to unmarshal runtimes: %w", err)
	}
	for _, runtime := range initializedRuntimes {
		if err := c.client.SRem(ctx, fmt.Sprintf("runtime:%s-%s", runtime.Name, runtime.Version), workerUuid).Err(); err != nil {
			return fmt.Errorf("failed to remove worker node from runtime set: %w", err)
		}
	}
	return nil
}
