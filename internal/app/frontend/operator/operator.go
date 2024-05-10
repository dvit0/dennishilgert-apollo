package operator

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/dennishilgert/apollo/internal/app/frontend/db"
	"github.com/dennishilgert/apollo/internal/app/frontend/models"
	"github.com/dennishilgert/apollo/internal/pkg/clients"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/messaging/producer"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	fleetpb "github.com/dennishilgert/apollo/internal/pkg/proto/fleet/v1"
	frontendpb "github.com/dennishilgert/apollo/internal/pkg/proto/frontend/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
	packpb "github.com/dennishilgert/apollo/internal/pkg/proto/pack/v1"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	workerpb "github.com/dennishilgert/apollo/internal/pkg/proto/worker/v1"
	"github.com/dennishilgert/apollo/internal/pkg/registry"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger("apollo.frontend.operator")

type FrontendOperator interface {
	InitializeFunction(ctx context.Context, functionUuid string, functionVersion string) error
	DeinitializeFunction(ctx context.Context, function *fleetpb.FunctionSpecs)
	InvokeFunction(ctx context.Context, request *frontendpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error)
	AddKernel(ctx context.Context, request *frontendpb.AddKernelRequest) error
	ListKernels(ctx context.Context, request *frontendpb.ListKernelsRequest) (*frontendpb.ListKernelsResponse, error)
	RemoveKernel(ctx context.Context, request *frontendpb.RemoveKernelRequest) error
	AddRuntime(ctx context.Context, request *frontendpb.AddRuntimeRequest) error
	ListRuntimes(ctx context.Context, request *frontendpb.ListRuntimesRequest) (*frontendpb.ListRuntimesResponse, error)
	RemoveRuntime(ctx context.Context, request *frontendpb.RemoveRuntimeRequest) error
	CreateFunction(ctx context.Context, request *frontendpb.CreateFunctionRequest) (*frontendpb.CreateFunctionResponse, error)
	GetFunction(ctx context.Context, request *frontendpb.GetFunctionRequest) (*frontendpb.GetFunctionResponse, error)
	ListFunctions(ctx context.Context, request *frontendpb.ListFunctionsRequest) (*frontendpb.ListFunctionsResponse, error)
	FunctionCodeUploadUrl(ctx context.Context, request *frontendpb.FunctionCodeUploadUrlRequest) (*frontendpb.FunctionCodeUploadUrlResponse, error)
	UpdateFunctionRuntime(ctx context.Context, request *frontendpb.UpdateFunctionRuntimeRequest) error
	UpdateFunctionResources(ctx context.Context, request *frontendpb.UpdateFunctionResourcesRequest) error
	UpdateFunctionVersion(ctx context.Context, functionSpecs *fleetpb.FunctionSpecs) error
	UpdateFunctionStatus(ctx context.Context, functionUuid string, status frontendpb.FunctionStatus) error
	DeleteFunction(ctx context.Context, request *frontendpb.DeleteFunctionRequest) error
}

type frontendOperator struct {
	databaseClient        db.DatabaseClient
	messagingProducer     producer.MessagingProducer
	serviceRegistryClient registry.ServiceRegistryClient
}

func NewFrontendOperator(databaseClient db.DatabaseClient, messagingProducer producer.MessagingProducer, serviceRegistryClient registry.ServiceRegistryClient) FrontendOperator {
	return &frontendOperator{
		databaseClient:        databaseClient,
		messagingProducer:     messagingProducer,
		serviceRegistryClient: serviceRegistryClient,
	}
}

func (o *frontendOperator) InitializeFunction(ctx context.Context, functionUuid string, functionVersion string) error {
	function, err := o.databaseClient.GetFunction(functionUuid)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}
	runtime, err := o.databaseClient.GetRuntime(function.RuntimeName, function.RuntimeVersion, function.RuntimeArchitecture)
	if err != nil {
		return fmt.Errorf("failed to get runtime: %w", err)
	}
	kernel, err := o.databaseClient.GetKernel(runtime.KernelName, runtime.KernelVersion, runtime.KernelArchitecture)
	if err != nil {
		return fmt.Errorf("failed to get kernel: %w", err)
	}
	initializationRequest := &workerpb.InitializeFunctionRequest{
		Function: &fleetpb.FunctionSpecs{
			Uuid:    function.Uuid,
			Version: function.Version,
		},
		Kernel: &fleetpb.KernelSpecs{
			Name:         kernel.Name,
			Version:      kernel.Version,
			Architecture: kernel.Architecture,
		},
		Runtime: &fleetpb.RuntimeSpecs{
			Name:         runtime.Name,
			Version:      runtime.Version,
			Architecture: runtime.KernelArchitecture,
		},
		Machine: &fleetpb.MachineSpecs{
			MemoryLimit: function.MemoryLimit,
			VcpuCores:   function.VCpuCores,
			IdleTtl:     function.IdleTtl,
			LogLevel:    &function.LogLevel,
		},
	}
	serviceInstance, err := o.serviceRegistryClient.AvailableServiceInstance(ctx, registrypb.InstanceType_WORKER_MANAGER)
	if err != nil {
		return fmt.Errorf("failed to get available service instance: %w", err)
	}
	clientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))
	if err != nil {
		return fmt.Errorf("failed to establish connection: %w", err)
	}
	defer clientConn.Close()

	apiClient := workerpb.NewWorkerManagerClient(clientConn)
	_, err = apiClient.InitializeFunction(ctx, initializationRequest)
	if err != nil {
		return fmt.Errorf("failed to initialize function: %w", err)
	}
	return nil
}

func (o *frontendOperator) DeinitializeFunction(ctx context.Context, function *fleetpb.FunctionSpecs) {
	o.messagingProducer.Publish(ctx, naming.MessagingFunctionDeinitializationRequestsTopic, &messagespb.FunctionDeinitializationRequestMessage{
		Function: function,
	})
}

func (o *frontendOperator) InvokeFunction(ctx context.Context, request *frontendpb.InvokeFunctionRequest) (*fleetpb.InvokeFunctionResponse, error) {
	httpTrigger, err := o.databaseClient.GetHttpTrigger(request.HttpTriggerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get http trigger: %w", err)
	}
	function, err := o.databaseClient.GetFunction(httpTrigger.FunctionUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}
	runtime, err := o.databaseClient.GetRuntime(function.RuntimeName, function.RuntimeVersion, function.RuntimeArchitecture)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime: %w", err)
	}
	kernel, err := o.databaseClient.GetKernel(runtime.KernelName, runtime.KernelVersion, runtime.KernelArchitecture)
	if err != nil {
		return nil, fmt.Errorf("failed to get kernel: %w", err)
	}

	// Merge environment variables from runtime and function.
	environment := append(runtime.Environment, function.Environment...)

	allocationRequest := &workerpb.AllocateInvocationRequest{
		Function: &fleetpb.FunctionSpecs{
			Uuid:    function.Uuid,
			Version: function.Version,
		},
		Kernel: &fleetpb.KernelSpecs{
			Name:         kernel.Name,
			Version:      kernel.Version,
			Architecture: kernel.Architecture,
		},
		Runtime: &fleetpb.RuntimeExecutionSpecs{
			Name:        runtime.Name,
			Version:     runtime.Version,
			Handler:     function.Handler,
			BinaryPath:  runtime.BinaryPath,
			BinaryArgs:  runtime.BinaryArgs,
			Environment: environment,
		},
		Machine: &fleetpb.MachineSpecs{
			MemoryLimit: function.MemoryLimit,
			VcpuCores:   function.VCpuCores,
			IdleTtl:     function.IdleTtl,
			Weight:      evaluateMachineWeight(function.MemoryLimit, function.VCpuCores),
			LogLevel:    &function.LogLevel,
		},
	}
	serviceInstance, err := o.serviceRegistryClient.AvailableServiceInstance(ctx, registrypb.InstanceType_WORKER_MANAGER)
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance: %w", err)
	}

	workerClientConn, err := establishConnection(ctx, fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}
	defer workerClientConn.Close()
	workerApiClient := workerpb.NewWorkerManagerClient(workerClientConn)
	allocationResponse, err := workerApiClient.AllocateInvocation(ctx, allocationRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate invocation: %w", err)
	}

	fleetClientConn, err := establishConnection(ctx, allocationResponse.WorkerNodeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}
	defer fleetClientConn.Close()
	fleetApiClient := fleetpb.NewFleetManagerClient(fleetClientConn)
	invocationResponse, err := fleetApiClient.InvokeFunction(ctx, &fleetpb.InvokeFunctionRequest{
		RunnerUuid: allocationResponse.RunnerUuid,
		Function: &fleetpb.FunctionSpecs{
			Uuid:    function.Uuid,
			Version: function.Version,
		},
		Event: &fleetpb.EventSpecs{
			Uuid:     uuid.NewString(),
			Type:     request.Event.Type,
			SourceIp: request.Event.SourceIp,
			Headers:  request.Event.Headers,
			Params:   request.Event.Params,
			Payload:  request.Event.Payload,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}
	return invocationResponse, nil
}

func (o *frontendOperator) AddKernel(ctx context.Context, request *frontendpb.AddKernelRequest) error {
	kernel := &models.Kernel{
		Name:         request.Kernel.Name,
		Version:      request.Kernel.Version,
		Architecture: request.Kernel.Architecture,
	}
	if err := o.databaseClient.CreateKernel(kernel); err != nil {
		return fmt.Errorf("failed to create kernel: %w", err)
	}
	return nil
}

func (o *frontendOperator) ListKernels(ctx context.Context, request *frontendpb.ListKernelsRequest) (*frontendpb.ListKernelsResponse, error) {
	kernels, err := o.databaseClient.ListKernels()
	if err != nil {
		return nil, fmt.Errorf("failed to list kernels: %w", err)
	}
	response := &frontendpb.ListKernelsResponse{}
	for _, kernel := range kernels {
		response.Kernels = append(response.Kernels, &frontendpb.KernelSpecs{
			Name:         kernel.Name,
			Version:      kernel.Version,
			Architecture: kernel.Architecture,
		})
	}
	return response, nil
}

func (o *frontendOperator) RemoveKernel(ctx context.Context, request *frontendpb.RemoveKernelRequest) error {
	if err := o.databaseClient.DeleteKernel(request.Name, request.Version, request.Architecture); err != nil {
		return fmt.Errorf("failed to remove kernel: %w", err)
	}
	return nil
}

func (o *frontendOperator) AddRuntime(ctx context.Context, request *frontendpb.AddRuntimeRequest) error {
	runtime := &models.Runtime{
		Name:               request.Runtime.Name,
		Version:            request.Runtime.Version,
		BinaryPath:         request.Runtime.BinaryPath,
		BinaryArgs:         request.Runtime.BinaryArgs,
		Environment:        request.Runtime.Environment,
		DisplayName:        request.Runtime.DisplayName,
		DefaultHandler:     request.Runtime.DefaultHandler,
		DefaultMemoryLimit: request.Runtime.DefaultMemoryLimit,
		DefaultVCpuCores:   request.Runtime.DefaultVCpuCores,
		KernelName:         request.Runtime.KernelName,
		KernelVersion:      request.Runtime.KernelVersion,
		KernelArchitecture: request.Runtime.KernelArchitecture,
	}
	if err := o.databaseClient.CreateRuntime(runtime); err != nil {
		return fmt.Errorf("failed to create runtime: %w", err)
	}
	return nil
}

func (o *frontendOperator) ListRuntimes(ctx context.Context, request *frontendpb.ListRuntimesRequest) (*frontendpb.ListRuntimesResponse, error) {
	runtimes, err := o.databaseClient.ListRuntimes()
	if err != nil {
		return nil, fmt.Errorf("failed to list runtimes: %w", err)
	}
	response := &frontendpb.ListRuntimesResponse{}
	for _, runtime := range runtimes {
		response.Runtimes = append(response.Runtimes, &frontendpb.RuntimeSpecs{
			Name:               runtime.Name,
			Version:            runtime.Version,
			BinaryPath:         runtime.BinaryPath,
			BinaryArgs:         runtime.BinaryArgs,
			Environment:        runtime.Environment,
			DisplayName:        runtime.DisplayName,
			DefaultHandler:     runtime.DefaultHandler,
			DefaultMemoryLimit: runtime.DefaultMemoryLimit,
			DefaultVCpuCores:   runtime.DefaultVCpuCores,
			KernelName:         runtime.KernelName,
			KernelVersion:      runtime.KernelVersion,
			KernelArchitecture: runtime.KernelArchitecture,
		})
	}
	return response, nil
}

func (o *frontendOperator) RemoveRuntime(ctx context.Context, request *frontendpb.RemoveRuntimeRequest) error {
	if err := o.databaseClient.DeleteRuntime(request.Name, request.Version, request.Architecture); err != nil {
		return fmt.Errorf("failed to remove runtime: %w", err)
	}
	return nil
}

func (o *frontendOperator) CreateFunction(ctx context.Context, request *frontendpb.CreateFunctionRequest) (*frontendpb.CreateFunctionResponse, error) {
	runtime, err := o.databaseClient.GetRuntime(request.RuntimeName, request.RuntimeVersion, request.RuntimeArchitecture)
	if err != nil {
		return nil, fmt.Errorf("failed to get runtime: %w", err)
	}
	function := &models.Function{
		Uuid:                uuid.NewString(),
		Name:                request.Name,
		Version:             naming.RuntimeInitialCodeDeclarator(),
		Handler:             runtime.DefaultHandler,
		MemoryLimit:         runtime.DefaultMemoryLimit,
		VCpuCores:           runtime.DefaultVCpuCores,
		Status:              frontendpb.FunctionStatus_CREATED,
		IdleTtl:             1,
		LogLevel:            log.LogLevel(),
		Environment:         []string{},
		RuntimeName:         request.RuntimeName,
		RuntimeVersion:      request.RuntimeVersion,
		RuntimeArchitecture: request.RuntimeArchitecture,
	}
	if err := o.databaseClient.CreateFunction(function); err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}
	trigger := &models.HttpTrigger{
		UrlId:        utils.RandomString(32, false, true),
		FunctionUuid: function.Uuid,
	}
	if err := o.databaseClient.CreateHttpTrigger(trigger); err != nil {
		return nil, fmt.Errorf("failed to create http trigger: %w", err)
	}
	if err := o.InitializeFunction(ctx, function.Uuid, function.Version); err != nil {
		return nil, fmt.Errorf("failed to initialize function: %w", err)
	}
	if err := o.UpdateFunctionStatus(ctx, function.Uuid, frontendpb.FunctionStatus_INITIALIZING); err != nil {
		return nil, fmt.Errorf("failed to update function status: %w", err)
	}
	return &frontendpb.CreateFunctionResponse{
		Uuid: function.Uuid,
	}, nil
}

func (o *frontendOperator) GetFunction(ctx context.Context, request *frontendpb.GetFunctionRequest) (*frontendpb.GetFunctionResponse, error) {
	function, err := o.databaseClient.GetFunction(request.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}
	triggerId, err := o.databaseClient.GetHttpTriggerByFunctionUuid(function.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get http trigger: %w", err)
	}
	return &frontendpb.GetFunctionResponse{
		Function: &frontendpb.FunctionSpecs{
			Uuid:                function.Uuid,
			Name:                function.Name,
			Version:             function.Version,
			Handler:             function.Handler,
			MemoryLimit:         function.MemoryLimit,
			VCpuCores:           function.VCpuCores,
			Status:              function.Status,
			IdleTtl:             function.IdleTtl,
			LogLevel:            function.LogLevel,
			Environment:         function.Environment,
			RuntimeName:         function.RuntimeName,
			RuntimeVersion:      function.RuntimeVersion,
			RuntimeArchitecture: function.RuntimeArchitecture,
			RuntimeDisplayName:  function.Runtime.DisplayName,
			HttpTriggerId:       triggerId.UrlId,
		},
	}, nil
}

func (o *frontendOperator) ListFunctions(ctx context.Context, request *frontendpb.ListFunctionsRequest) (*frontendpb.ListFunctionsResponse, error) {
	functions, err := o.databaseClient.ListFunctions()
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}
	response := &frontendpb.ListFunctionsResponse{}
	for _, function := range functions {
		triggerId, err := o.databaseClient.GetHttpTriggerByFunctionUuid(function.Uuid)
		if err != nil {
			return nil, fmt.Errorf("failed to get http trigger: %w", err)
		}
		response.Functions = append(response.Functions, &frontendpb.FunctionSpecs{
			Uuid:                function.Uuid,
			Name:                function.Name,
			Version:             function.Version,
			Handler:             function.Handler,
			MemoryLimit:         function.MemoryLimit,
			VCpuCores:           function.VCpuCores,
			Status:              function.Status,
			IdleTtl:             function.IdleTtl,
			LogLevel:            function.LogLevel,
			Environment:         function.Environment,
			RuntimeName:         function.RuntimeName,
			RuntimeVersion:      function.RuntimeVersion,
			RuntimeArchitecture: function.RuntimeArchitecture,
			RuntimeDisplayName:  function.Runtime.DisplayName,
			HttpTriggerId:       triggerId.UrlId,
		})
	}
	return response, nil
}

func (o *frontendOperator) FunctionCodeUploadUrl(ctx context.Context, request *frontendpb.FunctionCodeUploadUrlRequest) (*frontendpb.FunctionCodeUploadUrlResponse, error) {
	function, err := o.databaseClient.GetFunction(request.FunctionUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get function: %w", err)
	}
	serviceInstance, err := o.serviceRegistryClient.AvailableServiceInstance(ctx, registrypb.InstanceType_PACKAGE_SERVICE)
	if err != nil {
		return nil, fmt.Errorf("failed to get available service instance: %w", err)
	}

	packageClient, err := clients.NewPackageServiceClient(ctx, fmt.Sprintf("%s:%d", serviceInstance.Host, serviceInstance.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create package service client: %w", err)
	}

	newFunctionVersion := uuid.NewString()
	res, err := packageClient.PresignedUploadUrl(ctx, &packpb.PresignedUploadUrlRequest{
		BucketName: naming.StorageFunctionBucketName,
		ObjectName: naming.FunctionCodeStorageName(function.Uuid, newFunctionVersion),
		ExpiresIn:  3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned upload url: %w", err)
	}

	url, err := url.Parse(res.RawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return &frontendpb.FunctionCodeUploadUrlResponse{
		PresignedUploadUrl: url.String(),
	}, nil
}

func (o *frontendOperator) UpdateFunctionRuntime(ctx context.Context, request *frontendpb.UpdateFunctionRuntimeRequest) error {
	function, err := o.databaseClient.GetFunction(request.Uuid)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}
	function.Handler = updateOrKeep(function.Handler, request.Handler).(string)
	function.Environment = updateOrKeep(function.Environment, request.Environment).([]string)

	function.RuntimeName = updateOrKeep(function.RuntimeName, request.RuntimeName).(string)
	function.RuntimeVersion = updateOrKeep(function.RuntimeVersion, request.RuntimeVersion).(string)
	function.RuntimeArchitecture = updateOrKeep(function.RuntimeArchitecture, request.RuntimeArchitecture).(string)
	if err := o.databaseClient.UpdateFunction(function); err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

func (o *frontendOperator) UpdateFunctionResources(ctx context.Context, request *frontendpb.UpdateFunctionResourcesRequest) error {
	function, err := o.databaseClient.GetFunction(request.Uuid)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}
	function.MemoryLimit = updateOrKeep(function.MemoryLimit, request.MemoryLimit).(int32)
	function.VCpuCores = updateOrKeep(function.VCpuCores, request.VCpuCores).(int32)
	if err := o.databaseClient.UpdateFunction(function); err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

func (o *frontendOperator) UpdateFunctionVersion(ctx context.Context, functionSpecs *fleetpb.FunctionSpecs) error {
	function, err := o.databaseClient.GetFunction(functionSpecs.Uuid)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}
	function.Version = functionSpecs.Version
	if err := o.databaseClient.UpdateFunction(function); err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

func (o *frontendOperator) UpdateFunctionStatus(ctx context.Context, functionUuid string, status frontendpb.FunctionStatus) error {
	function, err := o.databaseClient.GetFunction(functionUuid)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}
	function.Status = status
	if err := o.databaseClient.UpdateFunction(function); err != nil {
		return fmt.Errorf("failed to update function: %w", err)
	}
	return nil
}

func (o *frontendOperator) DeleteFunction(ctx context.Context, request *frontendpb.DeleteFunctionRequest) error {
	if err := o.databaseClient.DeleteFunction(request.Uuid); err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}
	return nil
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

func evaluateMachineWeight(memoryLimit int32, vCpuCores int32) fleetpb.MachineWeight {
	if vCpuCores <= 1 && memoryLimit <= 128 {
		return fleetpb.MachineWeight_LIGHT
	} else if vCpuCores <= 2 && memoryLimit <= 256 {
		return fleetpb.MachineWeight_MEDIUM
	} else if vCpuCores <= 2 && memoryLimit <= 512 {
		return fleetpb.MachineWeight_HEAVY
	}
	return fleetpb.MachineWeight_SUPER_HEAVY
}

func updateOrKeep(value interface{}, newValue interface{}) interface{} {
	if newValue != "" {
		return newValue
	}
	return value
}
