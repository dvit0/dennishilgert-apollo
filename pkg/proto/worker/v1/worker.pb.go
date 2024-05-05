// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: apollo/proto/worker/v1/worker.proto

package workerpb

import (
	v1 "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	v11 "github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type InitializeFunctionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function *v1.FunctionSpecs `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	Kernel   *v1.KernelSpecs   `protobuf:"bytes,2,opt,name=kernel,proto3" json:"kernel,omitempty"`
	Runtime  *v1.RuntimeSpecs  `protobuf:"bytes,3,opt,name=runtime,proto3" json:"runtime,omitempty"`
	Machine  *v1.MachineSpecs  `protobuf:"bytes,4,opt,name=machine,proto3" json:"machine,omitempty"`
}

func (x *InitializeFunctionRequest) Reset() {
	*x = InitializeFunctionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitializeFunctionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitializeFunctionRequest) ProtoMessage() {}

func (x *InitializeFunctionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitializeFunctionRequest.ProtoReflect.Descriptor instead.
func (*InitializeFunctionRequest) Descriptor() ([]byte, []int) {
	return file_apollo_proto_worker_v1_worker_proto_rawDescGZIP(), []int{0}
}

func (x *InitializeFunctionRequest) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *InitializeFunctionRequest) GetKernel() *v1.KernelSpecs {
	if x != nil {
		return x.Kernel
	}
	return nil
}

func (x *InitializeFunctionRequest) GetRuntime() *v1.RuntimeSpecs {
	if x != nil {
		return x.Runtime
	}
	return nil
}

func (x *InitializeFunctionRequest) GetMachine() *v1.MachineSpecs {
	if x != nil {
		return x.Machine
	}
	return nil
}

type AllocateInvocationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function *v1.FunctionSpecs         `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	Kernel   *v1.KernelSpecs           `protobuf:"bytes,2,opt,name=kernel,proto3" json:"kernel,omitempty"`
	Runtime  *v1.RuntimeExecutionSpecs `protobuf:"bytes,3,opt,name=runtime,proto3" json:"runtime,omitempty"`
	Machine  *v1.MachineSpecs          `protobuf:"bytes,4,opt,name=machine,proto3" json:"machine,omitempty"`
}

func (x *AllocateInvocationRequest) Reset() {
	*x = AllocateInvocationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AllocateInvocationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllocateInvocationRequest) ProtoMessage() {}

func (x *AllocateInvocationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllocateInvocationRequest.ProtoReflect.Descriptor instead.
func (*AllocateInvocationRequest) Descriptor() ([]byte, []int) {
	return file_apollo_proto_worker_v1_worker_proto_rawDescGZIP(), []int{1}
}

func (x *AllocateInvocationRequest) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *AllocateInvocationRequest) GetKernel() *v1.KernelSpecs {
	if x != nil {
		return x.Kernel
	}
	return nil
}

func (x *AllocateInvocationRequest) GetRuntime() *v1.RuntimeExecutionSpecs {
	if x != nil {
		return x.Runtime
	}
	return nil
}

func (x *AllocateInvocationRequest) GetMachine() *v1.MachineSpecs {
	if x != nil {
		return x.Machine
	}
	return nil
}

type AllocateInvocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunnerUuid        string `protobuf:"bytes,1,opt,name=runner_uuid,json=runnerUuid,proto3" json:"runner_uuid,omitempty"`
	WorkerNodeAddress string `protobuf:"bytes,2,opt,name=worker_node_address,json=workerNodeAddress,proto3" json:"worker_node_address,omitempty"`
}

func (x *AllocateInvocationResponse) Reset() {
	*x = AllocateInvocationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AllocateInvocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AllocateInvocationResponse) ProtoMessage() {}

func (x *AllocateInvocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_worker_v1_worker_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AllocateInvocationResponse.ProtoReflect.Descriptor instead.
func (*AllocateInvocationResponse) Descriptor() ([]byte, []int) {
	return file_apollo_proto_worker_v1_worker_proto_rawDescGZIP(), []int{2}
}

func (x *AllocateInvocationResponse) GetRunnerUuid() string {
	if x != nil {
		return x.RunnerUuid
	}
	return ""
}

func (x *AllocateInvocationResponse) GetWorkerNodeAddress() string {
	if x != nil {
		return x.WorkerNodeAddress
	}
	return ""
}

var File_apollo_proto_worker_v1_worker_proto protoreflect.FileDescriptor

var file_apollo_proto_worker_v1_worker_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77,
	0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x1a, 0x23, 0x61,
	0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x68, 0x61, 0x72,
	0x65, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x21, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x97, 0x02, 0x0a, 0x19, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61,
	0x6c, 0x69, 0x7a, 0x65, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x40, 0x0a, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3a, 0x0a, 0x06, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65,
	0x72, 0x6e, 0x65, 0x6c, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x06, 0x6b, 0x65, 0x72, 0x6e, 0x65,
	0x6c, 0x12, 0x3d, 0x0a, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x12, 0x3d, 0x0a, 0x07, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x23, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e,
	0x65, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x07, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x22,
	0xa0, 0x02, 0x0a, 0x19, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x76, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x40, 0x0a,
	0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x24, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x3a, 0x0a, 0x06, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x22, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66,
	0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x53, 0x70,
	0x65, 0x63, 0x73, 0x52, 0x06, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x12, 0x46, 0x0a, 0x07, 0x72,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x61,
	0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x07, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x61, 0x63,
	0x68, 0x69, 0x6e, 0x65, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x07, 0x6d, 0x61, 0x63, 0x68, 0x69,
	0x6e, 0x65, 0x22, 0x6d, 0x0a, 0x1a, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x49, 0x6e,
	0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x75, 0x75, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x55, 0x75, 0x69,
	0x64, 0x12, 0x2e, 0x0a, 0x13, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x6e, 0x6f, 0x64, 0x65,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11,
	0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4e, 0x6f, 0x64, 0x65, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x32, 0x80, 0x02, 0x0a, 0x0d, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61,
	0x67, 0x65, 0x72, 0x12, 0x70, 0x0a, 0x12, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a,
	0x65, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x31, 0x2e, 0x61, 0x70, 0x6f, 0x6c,
	0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x61,
	0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x73, 0x68, 0x61, 0x72,
	0x65, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x7d, 0x0a, 0x12, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74,
	0x65, 0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x31, 0x2e, 0x61, 0x70,
	0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x76,
	0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32,
	0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x77, 0x6f,
	0x72, 0x6b, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65,
	0x49, 0x6e, 0x76, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x6e, 0x6e, 0x69, 0x73, 0x68, 0x69, 0x6c, 0x67, 0x65, 0x72, 0x74,
	0x2f, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x3b, 0x77, 0x6f, 0x72, 0x6b,
	0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apollo_proto_worker_v1_worker_proto_rawDescOnce sync.Once
	file_apollo_proto_worker_v1_worker_proto_rawDescData = file_apollo_proto_worker_v1_worker_proto_rawDesc
)

func file_apollo_proto_worker_v1_worker_proto_rawDescGZIP() []byte {
	file_apollo_proto_worker_v1_worker_proto_rawDescOnce.Do(func() {
		file_apollo_proto_worker_v1_worker_proto_rawDescData = protoimpl.X.CompressGZIP(file_apollo_proto_worker_v1_worker_proto_rawDescData)
	})
	return file_apollo_proto_worker_v1_worker_proto_rawDescData
}

var file_apollo_proto_worker_v1_worker_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_apollo_proto_worker_v1_worker_proto_goTypes = []interface{}{
	(*InitializeFunctionRequest)(nil),  // 0: apollo.proto.worker.v1.InitializeFunctionRequest
	(*AllocateInvocationRequest)(nil),  // 1: apollo.proto.worker.v1.AllocateInvocationRequest
	(*AllocateInvocationResponse)(nil), // 2: apollo.proto.worker.v1.AllocateInvocationResponse
	(*v1.FunctionSpecs)(nil),           // 3: apollo.proto.fleet.v1.FunctionSpecs
	(*v1.KernelSpecs)(nil),             // 4: apollo.proto.fleet.v1.KernelSpecs
	(*v1.RuntimeSpecs)(nil),            // 5: apollo.proto.fleet.v1.RuntimeSpecs
	(*v1.MachineSpecs)(nil),            // 6: apollo.proto.fleet.v1.MachineSpecs
	(*v1.RuntimeExecutionSpecs)(nil),   // 7: apollo.proto.fleet.v1.RuntimeExecutionSpecs
	(*v11.EmptyResponse)(nil),          // 8: apollo.proto.shared.v1.EmptyResponse
}
var file_apollo_proto_worker_v1_worker_proto_depIdxs = []int32{
	3,  // 0: apollo.proto.worker.v1.InitializeFunctionRequest.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	4,  // 1: apollo.proto.worker.v1.InitializeFunctionRequest.kernel:type_name -> apollo.proto.fleet.v1.KernelSpecs
	5,  // 2: apollo.proto.worker.v1.InitializeFunctionRequest.runtime:type_name -> apollo.proto.fleet.v1.RuntimeSpecs
	6,  // 3: apollo.proto.worker.v1.InitializeFunctionRequest.machine:type_name -> apollo.proto.fleet.v1.MachineSpecs
	3,  // 4: apollo.proto.worker.v1.AllocateInvocationRequest.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	4,  // 5: apollo.proto.worker.v1.AllocateInvocationRequest.kernel:type_name -> apollo.proto.fleet.v1.KernelSpecs
	7,  // 6: apollo.proto.worker.v1.AllocateInvocationRequest.runtime:type_name -> apollo.proto.fleet.v1.RuntimeExecutionSpecs
	6,  // 7: apollo.proto.worker.v1.AllocateInvocationRequest.machine:type_name -> apollo.proto.fleet.v1.MachineSpecs
	0,  // 8: apollo.proto.worker.v1.WorkerManager.InitializeFunction:input_type -> apollo.proto.worker.v1.InitializeFunctionRequest
	1,  // 9: apollo.proto.worker.v1.WorkerManager.AllocateInvocation:input_type -> apollo.proto.worker.v1.AllocateInvocationRequest
	8,  // 10: apollo.proto.worker.v1.WorkerManager.InitializeFunction:output_type -> apollo.proto.shared.v1.EmptyResponse
	2,  // 11: apollo.proto.worker.v1.WorkerManager.AllocateInvocation:output_type -> apollo.proto.worker.v1.AllocateInvocationResponse
	10, // [10:12] is the sub-list for method output_type
	8,  // [8:10] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_apollo_proto_worker_v1_worker_proto_init() }
func file_apollo_proto_worker_v1_worker_proto_init() {
	if File_apollo_proto_worker_v1_worker_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apollo_proto_worker_v1_worker_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitializeFunctionRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apollo_proto_worker_v1_worker_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AllocateInvocationRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_apollo_proto_worker_v1_worker_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AllocateInvocationResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_apollo_proto_worker_v1_worker_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_apollo_proto_worker_v1_worker_proto_goTypes,
		DependencyIndexes: file_apollo_proto_worker_v1_worker_proto_depIdxs,
		MessageInfos:      file_apollo_proto_worker_v1_worker_proto_msgTypes,
	}.Build()
	File_apollo_proto_worker_v1_worker_proto = out.File
	file_apollo_proto_worker_v1_worker_proto_rawDesc = nil
	file_apollo_proto_worker_v1_worker_proto_goTypes = nil
	file_apollo_proto_worker_v1_worker_proto_depIdxs = nil
}
