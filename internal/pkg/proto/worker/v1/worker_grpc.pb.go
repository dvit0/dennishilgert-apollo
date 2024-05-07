// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.2
// source: apollo/proto/worker/v1/worker.proto

package workerpb

import (
	context "context"
	v1 "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	WorkerManager_InitializeFunction_FullMethodName = "/apollo.proto.worker.v1.WorkerManager/InitializeFunction"
	WorkerManager_AllocateInvocation_FullMethodName = "/apollo.proto.worker.v1.WorkerManager/AllocateInvocation"
)

// WorkerManagerClient is the client API for WorkerManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkerManagerClient interface {
	InitializeFunction(ctx context.Context, in *InitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error)
	AllocateInvocation(ctx context.Context, in *AllocateInvocationRequest, opts ...grpc.CallOption) (*AllocateInvocationResponse, error)
}

type workerManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerManagerClient(cc grpc.ClientConnInterface) WorkerManagerClient {
	return &workerManagerClient{cc}
}

func (c *workerManagerClient) InitializeFunction(ctx context.Context, in *InitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error) {
	out := new(v1.EmptyResponse)
	err := c.cc.Invoke(ctx, WorkerManager_InitializeFunction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerManagerClient) AllocateInvocation(ctx context.Context, in *AllocateInvocationRequest, opts ...grpc.CallOption) (*AllocateInvocationResponse, error) {
	out := new(AllocateInvocationResponse)
	err := c.cc.Invoke(ctx, WorkerManager_AllocateInvocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerManagerServer is the server API for WorkerManager service.
// All implementations should embed UnimplementedWorkerManagerServer
// for forward compatibility
type WorkerManagerServer interface {
	InitializeFunction(context.Context, *InitializeFunctionRequest) (*v1.EmptyResponse, error)
	AllocateInvocation(context.Context, *AllocateInvocationRequest) (*AllocateInvocationResponse, error)
}

// UnimplementedWorkerManagerServer should be embedded to have forward compatible implementations.
type UnimplementedWorkerManagerServer struct {
}

func (UnimplementedWorkerManagerServer) InitializeFunction(context.Context, *InitializeFunctionRequest) (*v1.EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitializeFunction not implemented")
}
func (UnimplementedWorkerManagerServer) AllocateInvocation(context.Context, *AllocateInvocationRequest) (*AllocateInvocationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AllocateInvocation not implemented")
}

// UnsafeWorkerManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerManagerServer will
// result in compilation errors.
type UnsafeWorkerManagerServer interface {
	mustEmbedUnimplementedWorkerManagerServer()
}

func RegisterWorkerManagerServer(s grpc.ServiceRegistrar, srv WorkerManagerServer) {
	s.RegisterService(&WorkerManager_ServiceDesc, srv)
}

func _WorkerManager_InitializeFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitializeFunctionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerManagerServer).InitializeFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerManager_InitializeFunction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerManagerServer).InitializeFunction(ctx, req.(*InitializeFunctionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerManager_AllocateInvocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AllocateInvocationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerManagerServer).AllocateInvocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerManager_AllocateInvocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerManagerServer).AllocateInvocation(ctx, req.(*AllocateInvocationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorkerManager_ServiceDesc is the grpc.ServiceDesc for WorkerManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkerManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "apollo.proto.worker.v1.WorkerManager",
	HandlerType: (*WorkerManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitializeFunction",
			Handler:    _WorkerManager_InitializeFunction_Handler,
		},
		{
			MethodName: "AllocateInvocation",
			Handler:    _WorkerManager_AllocateInvocation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apollo/proto/worker/v1/worker.proto",
}