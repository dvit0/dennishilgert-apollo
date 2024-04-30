package placement

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/cache"
	"github.com/dennishilgert/apollo/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	workerpb "github.com/dennishilgert/apollo/pkg/proto/worker/v1"
)

var log = logger.NewLogger("apollo.worker.placement")

type FunctionHeaviness int

const (
	FunctionHeavinessLight FunctionHeaviness = iota
	FunctionHeavinessMedium
	FunctionHeavinessHeavy
)

func (f FunctionHeaviness) String() string {
	switch f {
	case FunctionHeavinessLight:
		return "LIGHT"
	case FunctionHeavinessMedium:
		return "MEDIUM"
	case FunctionHeavinessHeavy:
		return "HEAVY"
	}
	return "UNKNOWN"
}

type Options struct{}

type PlacementService interface {
	FunctionInitializationWorker(ctx context.Context, request *workerpb.InitializeFunctionRequest) (*registrypb.WorkerInstance, error)
	RunnerAllocationWorker(ctx context.Context, request *workerpb.AllocateRunnerRequest) (*registrypb.WorkerInstance, error)
	Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, *registrypb.WorkerInstanceMetrics, error)
	WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error)
	WorkersByFunction(ctx context.Context, functionUuid string) ([]string, error)
}

type placementService struct {
	cacheClient cache.CacheClient
}

func NewPlacementService(cacheClient cache.CacheClient, opts Options) PlacementService {
	return &placementService{
		cacheClient: cacheClient,
	}
}

// FunctionInitializationWorker returns a worker instance that matches the function requirements.
func (p *placementService) FunctionInitializationWorker(ctx context.Context, request *workerpb.InitializeFunctionRequest) (*registrypb.WorkerInstance, error) {
	functionHeaviness := p.DetermineFunctionHeaviness(int(request.Machine.VcpuCores), int(request.Machine.MemoryLimit))
	workers, err := p.WorkersByArchitecture(ctx, request.Machine.Architecture)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers by architecture: %w", err)
	}
	for _, worker := range workers {
		workerInstance, workerMetrics, err := p.Worker(ctx, worker)
		if err != nil {
			return nil, fmt.Errorf("failed to get worker instance: %w", err)
		}
		if ok := p.checkPlacement(workerMetrics, functionHeaviness); ok {
			return workerInstance, nil
		}
	}
	return nil, fmt.Errorf("no worker found for architecture")
}

func (p *placementService) RunnerAllocationWorker(ctx context.Context, request *workerpb.AllocateRunnerRequest) (*registrypb.WorkerInstance, error) {
	workers, err := p.WorkersByFunction(ctx, request.FunctionUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers by function: %w", err)
	}
	for _, worker := range workers {
		workerInstance, workerMetrics, err := p.Worker(ctx, worker)
		if err != nil {
			return nil, fmt.Errorf("failed to get worker instance: %w", err)
		}
		if ok := p.checkPlacement(workerMetrics, FunctionHeavinessLight); ok {
			return workerInstance, nil
		}
	}
	return nil, fmt.Errorf("no worker found for function")
}

// Worker returns the worker instance and its metrics from the cache.
func (p *placementService) Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, *registrypb.WorkerInstanceMetrics, error) {
	key := naming.CacheWorkerInstanceKeyName(workerUuid)
	values, err := p.cacheClient.Client().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worker instance from cache: %w", err)
	}
	initializedFunctions := make([]*registrypb.Function, 0)
	if err := json.Unmarshal([]byte(values["functions"]), &initializedFunctions); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal functions: %w", err)
	}
	var metrics *registrypb.WorkerInstanceMetrics
	if err := json.Unmarshal([]byte(values["metrics"]), &metrics); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}
	port, _ := strconv.Atoi(values["port"])
	return &registrypb.WorkerInstance{
		WorkerUuid:           workerUuid,
		Architecture:         values["architecture"],
		Host:                 values["host"],
		Port:                 int32(port),
		InitializedFunctions: initializedFunctions,
	}, metrics, nil
}

// WorkersByArchitecture returns a list of worker nodes by architecture.
func (p *placementService) WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error) {
	members, err := p.cacheClient.Client().SMembers(ctx, naming.CacheArchitectureSetKey(architecture)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instances by architecture: %w", err)
	}
	return members, nil
}

// WorkersByFunction returns a list of worker nodes by function.
func (p *placementService) WorkersByFunction(ctx context.Context, functionUuid string) ([]string, error) {
	members, err := p.cacheClient.Client().SMembers(ctx, naming.CacheFunctionSetKey(functionUuid)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instances by runtime: %w", err)
	}
	return members, nil
}

// DetermineFunctionHeaviness determines the heaviness of a function based on the number of CPU cores and memory limit.
func (p *placementService) DetermineFunctionHeaviness(cpuCores int, memoryLimit int) FunctionHeaviness {
	if cpuCores <= 1 && memoryLimit <= 128 {
		return FunctionHeavinessLight
	} else if cpuCores <= 2 && memoryLimit <= 256 {
		return FunctionHeavinessMedium
	}
	return FunctionHeavinessHeavy
}

// checkPlacement checks if the worker instance is suitable for the function placement.
func (p *placementService) checkPlacement(metrics *registrypb.WorkerInstanceMetrics, functionHeaviness FunctionHeaviness) bool {
	if (metrics.CpuUsage <= 80) && (metrics.MemoryUsage <= 80) && (metrics.StorageUsage <= 80) {
		// TODO: Currently only the machine metrics are checked.
		// In the future the load of the different function heavinesses should be considered as well.
		log.Infof("worker instance is suitable for function with heaviness %s", functionHeaviness.String())
		return true
	}
	return false
}
