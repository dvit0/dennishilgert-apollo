// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.2
// source: apollo/proto/fleet/v1/fleet.proto

package fleetpb

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
	FleetManager_InitializeFunction_FullMethodName   = "/apollo.proto.fleet.v1.FleetManager/InitializeFunction"
	FleetManager_DeinitializeFunction_FullMethodName = "/apollo.proto.fleet.v1.FleetManager/DeinitializeFunction"
	FleetManager_AvailableRunner_FullMethodName      = "/apollo.proto.fleet.v1.FleetManager/AvailableRunner"
	FleetManager_ProvisionRunner_FullMethodName      = "/apollo.proto.fleet.v1.FleetManager/ProvisionRunner"
	FleetManager_InvokeFunction_FullMethodName       = "/apollo.proto.fleet.v1.FleetManager/InvokeFunction"
)

// FleetManagerClient is the client API for FleetManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FleetManagerClient interface {
	InitializeFunction(ctx context.Context, in *InitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error)
	DeinitializeFunction(ctx context.Context, in *DeinitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error)
	AvailableRunner(ctx context.Context, in *AvailableRunnerRequest, opts ...grpc.CallOption) (*AvailableRunnerResponse, error)
	ProvisionRunner(ctx context.Context, in *ProvisionRunnerRequest, opts ...grpc.CallOption) (*ProvisionRunnerResponse, error)
	InvokeFunction(ctx context.Context, in *InvokeFunctionRequest, opts ...grpc.CallOption) (*InvokeFunctionResponse, error)
}

type fleetManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewFleetManagerClient(cc grpc.ClientConnInterface) FleetManagerClient {
	return &fleetManagerClient{cc}
}

func (c *fleetManagerClient) InitializeFunction(ctx context.Context, in *InitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error) {
	out := new(v1.EmptyResponse)
	err := c.cc.Invoke(ctx, FleetManager_InitializeFunction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetManagerClient) DeinitializeFunction(ctx context.Context, in *DeinitializeFunctionRequest, opts ...grpc.CallOption) (*v1.EmptyResponse, error) {
	out := new(v1.EmptyResponse)
	err := c.cc.Invoke(ctx, FleetManager_DeinitializeFunction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetManagerClient) AvailableRunner(ctx context.Context, in *AvailableRunnerRequest, opts ...grpc.CallOption) (*AvailableRunnerResponse, error) {
	out := new(AvailableRunnerResponse)
	err := c.cc.Invoke(ctx, FleetManager_AvailableRunner_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetManagerClient) ProvisionRunner(ctx context.Context, in *ProvisionRunnerRequest, opts ...grpc.CallOption) (*ProvisionRunnerResponse, error) {
	out := new(ProvisionRunnerResponse)
	err := c.cc.Invoke(ctx, FleetManager_ProvisionRunner_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fleetManagerClient) InvokeFunction(ctx context.Context, in *InvokeFunctionRequest, opts ...grpc.CallOption) (*InvokeFunctionResponse, error) {
	out := new(InvokeFunctionResponse)
	err := c.cc.Invoke(ctx, FleetManager_InvokeFunction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FleetManagerServer is the server API for FleetManager service.
// All implementations should embed UnimplementedFleetManagerServer
// for forward compatibility
type FleetManagerServer interface {
	InitializeFunction(context.Context, *InitializeFunctionRequest) (*v1.EmptyResponse, error)
	DeinitializeFunction(context.Context, *DeinitializeFunctionRequest) (*v1.EmptyResponse, error)
	AvailableRunner(context.Context, *AvailableRunnerRequest) (*AvailableRunnerResponse, error)
	ProvisionRunner(context.Context, *ProvisionRunnerRequest) (*ProvisionRunnerResponse, error)
	InvokeFunction(context.Context, *InvokeFunctionRequest) (*InvokeFunctionResponse, error)
}

// UnimplementedFleetManagerServer should be embedded to have forward compatible implementations.
type UnimplementedFleetManagerServer struct {
}

func (UnimplementedFleetManagerServer) InitializeFunction(context.Context, *InitializeFunctionRequest) (*v1.EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitializeFunction not implemented")
}
func (UnimplementedFleetManagerServer) DeinitializeFunction(context.Context, *DeinitializeFunctionRequest) (*v1.EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeinitializeFunction not implemented")
}
func (UnimplementedFleetManagerServer) AvailableRunner(context.Context, *AvailableRunnerRequest) (*AvailableRunnerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AvailableRunner not implemented")
}
func (UnimplementedFleetManagerServer) ProvisionRunner(context.Context, *ProvisionRunnerRequest) (*ProvisionRunnerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProvisionRunner not implemented")
}
func (UnimplementedFleetManagerServer) InvokeFunction(context.Context, *InvokeFunctionRequest) (*InvokeFunctionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InvokeFunction not implemented")
}

// UnsafeFleetManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FleetManagerServer will
// result in compilation errors.
type UnsafeFleetManagerServer interface {
	mustEmbedUnimplementedFleetManagerServer()
}

func RegisterFleetManagerServer(s grpc.ServiceRegistrar, srv FleetManagerServer) {
	s.RegisterService(&FleetManager_ServiceDesc, srv)
}

func _FleetManager_InitializeFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitializeFunctionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetManagerServer).InitializeFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FleetManager_InitializeFunction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetManagerServer).InitializeFunction(ctx, req.(*InitializeFunctionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FleetManager_DeinitializeFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeinitializeFunctionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetManagerServer).DeinitializeFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FleetManager_DeinitializeFunction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetManagerServer).DeinitializeFunction(ctx, req.(*DeinitializeFunctionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FleetManager_AvailableRunner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AvailableRunnerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetManagerServer).AvailableRunner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FleetManager_AvailableRunner_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetManagerServer).AvailableRunner(ctx, req.(*AvailableRunnerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FleetManager_ProvisionRunner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProvisionRunnerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetManagerServer).ProvisionRunner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FleetManager_ProvisionRunner_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetManagerServer).ProvisionRunner(ctx, req.(*ProvisionRunnerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FleetManager_InvokeFunction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvokeFunctionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FleetManagerServer).InvokeFunction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FleetManager_InvokeFunction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FleetManagerServer).InvokeFunction(ctx, req.(*InvokeFunctionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FleetManager_ServiceDesc is the grpc.ServiceDesc for FleetManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FleetManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "apollo.proto.fleet.v1.FleetManager",
	HandlerType: (*FleetManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitializeFunction",
			Handler:    _FleetManager_InitializeFunction_Handler,
		},
		{
			MethodName: "DeinitializeFunction",
			Handler:    _FleetManager_DeinitializeFunction_Handler,
		},
		{
			MethodName: "AvailableRunner",
			Handler:    _FleetManager_AvailableRunner_Handler,
		},
		{
			MethodName: "ProvisionRunner",
			Handler:    _FleetManager_ProvisionRunner_Handler,
		},
		{
			MethodName: "InvokeFunction",
			Handler:    _FleetManager_InvokeFunction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "apollo/proto/fleet/v1/fleet.proto",
}
