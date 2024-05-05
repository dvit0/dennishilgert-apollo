package placement

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/dennishilgert/apollo/internal/app/worker/evaluation"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	"github.com/dennishilgert/apollo/pkg/cache"
	"github.com/dennishilgert/apollo/pkg/logger"
	fleetpb "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	workerpb "github.com/dennishilgert/apollo/pkg/proto/worker/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger("apollo.worker.placement")

type Options struct{}

type PlacementService interface {
	AllocateFunctionInitialization(ctx context.Context, request *workerpb.InitializeFunctionRequest) error
	FindAvailableRunner(ctx context.Context, request *workerpb.AllocateInvocationRequest) (*fleetpb.AvailableRunnerResponse, error)
	AllocateRunnerProvisioning(ctx context.Context, request *workerpb.AllocateInvocationRequest) (*fleetpb.ProvisionRunnerResponse, error)
	AddInitializedFunction(ctx context.Context, workerUuid string, functionIdentifier string) error
	RemoveInitializedFunction(ctx context.Context, workerUuid string, functionIdentifier string) error
	Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, *registrypb.WorkerInstanceMetrics, error)
	WorkersByArchitecture(ctx context.Context, architecture string) ([]string, error)
	WorkersByFunction(ctx context.Context, functionIdentifier string) ([]string, error)
}

type placementService struct {
	cacheClient cache.CacheClient
}

func NewPlacementService(cacheClient cache.CacheClient, opts Options) PlacementService {
	return &placementService{
		cacheClient: cacheClient,
	}
}

func (p *placementService) AllocateFunctionInitialization(ctx context.Context, request *workerpb.InitializeFunctionRequest) error {
	workers, err := p.WorkersByArchitecture(ctx, request.Kernel.Architecture)
	if err != nil {
		return fmt.Errorf("failed to get workers by architecture: %w", err)
	}
	if len(workers) == 0 {
		return fmt.Errorf("no workers found for architecture")
	}
	workerEvaluations := make([]evaluation.EvaluationResult, 0)
	for _, worker := range workers {
		workerInstance, workerMetrics, err := p.Worker(ctx, worker)
		if err != nil {
			return fmt.Errorf("failed to get worker instance: %w", err)
		}
		eval := evaluation.EvaluateFunctionInitialization(workerInstance, workerMetrics)
		workerEvaluations = append(workerEvaluations, *eval)
	}
	workerInstance, err := evaluation.SelectWorker(workerEvaluations)
	if err != nil {
		return fmt.Errorf("failed to select worker: %w", err)
	}

	clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", workerInstance.Host, workerInstance.Port))
	if err != nil {
		return fmt.Errorf("failed to establish connection to worker instance: %w", err)
	}
	defer clientConn.Close()

	apiClient := fleetpb.NewFleetManagerClient(clientConn)
	transformedReq := &fleetpb.InitializeFunctionRequest{
		Function: request.Function,
		Kernel:   request.Kernel,
		Runtime:  request.Runtime,
	}
	_, err = apiClient.InitializeFunction(ctx, transformedReq)
	if err != nil {
		return fmt.Errorf("failed to initialize function on worker instance: %w", err)
	}
	return nil
}

func (p *placementService) FindAvailableRunner(ctx context.Context, request *workerpb.AllocateInvocationRequest) (*fleetpb.AvailableRunnerResponse, error) {
	functionIdentifier := naming.FunctionIdentifier(request.Function.Uuid, request.Function.Version)
	workers, err := p.WorkersByFunction(ctx, functionIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers by function: %w", err)
	}
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers found for function")
	}
	for _, worker := range workers {
		workerInstance, _, err := p.Worker(ctx, worker)
		if err != nil {
			return nil, fmt.Errorf("failed to get worker instance: %w", err)
		}

		clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", workerInstance.Host, workerInstance.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to establish connection to worker instance: %w", err)
		}
		defer clientConn.Close()

		apiClient := fleetpb.NewFleetManagerClient(clientConn)
		transformedReq := &fleetpb.AvailableRunnerRequest{
			Function: request.Function,
		}
		res, err := apiClient.AvailableRunner(ctx, transformedReq)
		if err != nil {
			continue
		}
		return res, nil
	}

	return nil, fmt.Errorf("no available runner found")
}

func (p *placementService) AllocateRunnerProvisioning(ctx context.Context, request *workerpb.AllocateInvocationRequest) (*fleetpb.ProvisionRunnerResponse, error) {
	runnerHeaviness := evaluation.EvaluateRunnerHeaviness(int(request.Machine.VcpuCores), int(request.Machine.MemoryLimit))
	functionIdentifier := naming.FunctionIdentifier(request.Function.Uuid, request.Function.Version)
	workers, err := p.WorkersByFunction(ctx, functionIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get workers by function: %w", err)
	}
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers found for function")
	}
	workerEvaluations := make([]evaluation.EvaluationResult, 0)
	for _, worker := range workers {
		workerInstance, workerMetrics, err := p.Worker(ctx, worker)
		if err != nil {
			return nil, fmt.Errorf("failed to get worker instance: %w", err)
		}
		eval := evaluation.EvaluateRunnerProvisioning(runnerHeaviness, workerInstance, workerMetrics)
		workerEvaluations = append(workerEvaluations, *eval)
	}
	workerInstance, err := evaluation.SelectWorker(workerEvaluations)
	if err != nil {
		return nil, fmt.Errorf("failed to select worker: %w", err)
	}

	clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", workerInstance.Host, workerInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to worker instance: %w", err)
	}
	defer clientConn.Close()

	apiClient := fleetpb.NewFleetManagerClient(clientConn)
	transformedReq := &fleetpb.ProvisionRunnerRequest{
		Function: request.Function,
		Kernel:   request.Kernel,
		Runtime:  request.Runtime,
		Machine:  request.Machine,
	}
	res, err := apiClient.ProvisionRunner(ctx, transformedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to provision runner on worker instance: %w", err)
	}

	return res, nil
}

func (p *placementService) AddInitializedFunction(ctx context.Context, workerUuid string, functionIdentifier string) error {
	workerInstance, _, err := p.Worker(ctx, workerUuid)
	if err != nil {
		return fmt.Errorf("failed to add initialized function: %w", err)
	}
	initializedFunctions := append(workerInstance.InitializedFunctions, functionIdentifier)
	initializedFunctionsBytes, err := json.Marshal(initializedFunctions)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized functions: %w", err)
	}
	if err := p.cacheClient.Client().HSet(ctx, naming.CacheWorkerInstanceKeyName(workerUuid), map[string]interface{}{
		"functions": string(initializedFunctionsBytes),
	}).Err(); err != nil {
		return fmt.Errorf("failed to set initialized functions: %w", err)
	}
	if err := p.cacheClient.Client().SAdd(ctx, naming.CacheFunctionSetKey(functionIdentifier), workerUuid).Err(); err != nil {
		return fmt.Errorf("failed to add worker instance to function set: %w", err)
	}
	return nil
}

func (p *placementService) RemoveInitializedFunction(ctx context.Context, workerUuid string, functionIdentifier string) error {
	workerInstance, _, err := p.Worker(ctx, workerUuid)
	if err != nil {
		return fmt.Errorf("failed to remove initialized function: %w", err)
	}
	initializedFunctions := make([]string, 0)
	for _, initializedFunction := range workerInstance.InitializedFunctions {
		if initializedFunction != functionIdentifier {
			initializedFunctions = append(initializedFunctions, initializedFunction)
		}
	}
	initializedFunctionsBytes, err := json.Marshal(initializedFunctions)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized functions: %w", err)
	}
	if err := p.cacheClient.Client().HSet(ctx, naming.CacheWorkerInstanceKeyName(workerUuid), map[string]interface{}{
		"functions": string(initializedFunctionsBytes),
	}).Err(); err != nil {
		return fmt.Errorf("failed to set initialized functions: %w", err)
	}
	if err := p.cacheClient.Client().SRem(ctx, naming.CacheFunctionSetKey(functionIdentifier), workerUuid).Err(); err != nil {
		return fmt.Errorf("failed to remove worker instance from function set: %w", err)
	}
	return nil
}

// Worker returns the worker instance and its metrics from the cache.
func (p *placementService) Worker(ctx context.Context, workerUuid string) (*registrypb.WorkerInstance, *registrypb.WorkerInstanceMetrics, error) {
	key := naming.CacheWorkerInstanceKeyName(workerUuid)
	values, err := p.cacheClient.Client().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worker instance from cache: %w", err)
	}
	if len(values) == 0 {
		return nil, nil, fmt.Errorf("worker instance not found in cache")
	}
	initializedFunctions := make([]string, 0)
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
func (p *placementService) WorkersByFunction(ctx context.Context, functionIdentifier string) ([]string, error) {
	members, err := p.cacheClient.Client().SMembers(ctx, naming.CacheFunctionSetKey(functionIdentifier)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker instances by runtime: %w", err)
	}
	return members, nil
}

func establishConnection(ctx context.Context, address string) (*grpc.ClientConn, error) {
	const retrySeconds = 3     // trying to connect for a period of 3 seconds
	const retriesPerSecond = 2 // trying to connect 2 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			return conn, nil
		} else {
			if conn != nil {
				conn.Close()
			}
			log.Errorf("failed to establish connection to service registry - reason: %v", err)
		}
		// Wait before retrying, but stop if context is done.
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context done before connection to servie registry could be established: %w", ctx.Err())
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return nil, fmt.Errorf("failed to establish connection to service registry after %d seconds", retrySeconds)
}
