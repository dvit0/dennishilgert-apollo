package placement

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/cache"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

type Options struct{}

type PlacementService interface {
	Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, error)
	WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error)
	WorkersByRuntime(ctx context.Context, runtimeName string, runtimeVersion string) ([]string, error)
}

type placementService struct {
	cacheClient cache.CacheClient
}

func NewPlacementService(cacheClient cache.CacheClient, opts Options) PlacementService {
	return &placementService{
		cacheClient: cacheClient,
	}
}

// Worker returns a worker node from the cache.
func (p *placementService) Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, error) {
	values, err := p.cacheClient.Client().HGetAll(ctx, workerUuid).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instance from cache: %w", err)
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
func (p *placementService) WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error) {
	members, err := p.cacheClient.Client().SMembers(ctx, naming.CacheArchitectureSetKey(architecture)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instances by architecture: %w", err)
	}
	return members, nil
}

// WorkersByRuntime returns a list of worker nodes by runtime.
func (p *placementService) WorkersByRuntime(ctx context.Context, runtimeName string, runtimeVersion string) ([]string, error) {
	members, err := p.cacheClient.Client().SMembers(ctx, naming.CacheRuntimeSetKey(runtimeName, runtimeVersion)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instances by runtime: %w", err)
	}
	return members, nil
}
