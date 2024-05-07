// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: apollo/proto/messages/v1/messages.proto

package messagespb

import (
	v1 "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	v11 "github.com/dennishilgert/apollo/pkg/proto/frontend/v1"
	v12 "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
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

type FunctionInitializationMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function   *v1.FunctionSpecs `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	WorkerUuid string            `protobuf:"bytes,2,opt,name=worker_uuid,json=workerUuid,proto3" json:"worker_uuid,omitempty"`
	Reason     string            `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Success    bool              `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *FunctionInitializationMessage) Reset() {
	*x = FunctionInitializationMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionInitializationMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionInitializationMessage) ProtoMessage() {}

func (x *FunctionInitializationMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionInitializationMessage.ProtoReflect.Descriptor instead.
func (*FunctionInitializationMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{0}
}

func (x *FunctionInitializationMessage) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *FunctionInitializationMessage) GetWorkerUuid() string {
	if x != nil {
		return x.WorkerUuid
	}
	return ""
}

func (x *FunctionInitializationMessage) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *FunctionInitializationMessage) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type FunctionDeinitializationMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function   *v1.FunctionSpecs `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	WorkerUuid string            `protobuf:"bytes,2,opt,name=worker_uuid,json=workerUuid,proto3" json:"worker_uuid,omitempty"`
	Reason     string            `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Success    bool              `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *FunctionDeinitializationMessage) Reset() {
	*x = FunctionDeinitializationMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionDeinitializationMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionDeinitializationMessage) ProtoMessage() {}

func (x *FunctionDeinitializationMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionDeinitializationMessage.ProtoReflect.Descriptor instead.
func (*FunctionDeinitializationMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{1}
}

func (x *FunctionDeinitializationMessage) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *FunctionDeinitializationMessage) GetWorkerUuid() string {
	if x != nil {
		return x.WorkerUuid
	}
	return ""
}

func (x *FunctionDeinitializationMessage) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *FunctionDeinitializationMessage) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type FunctionPackageCreationMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function *v1.FunctionSpecs `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	Runtime  *v1.RuntimeSpecs  `protobuf:"bytes,2,opt,name=runtime,proto3" json:"runtime,omitempty"`
}

func (x *FunctionPackageCreationMessage) Reset() {
	*x = FunctionPackageCreationMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionPackageCreationMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionPackageCreationMessage) ProtoMessage() {}

func (x *FunctionPackageCreationMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionPackageCreationMessage.ProtoReflect.Descriptor instead.
func (*FunctionPackageCreationMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{2}
}

func (x *FunctionPackageCreationMessage) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *FunctionPackageCreationMessage) GetRuntime() *v1.RuntimeSpecs {
	if x != nil {
		return x.Runtime
	}
	return nil
}

type FunctionStatusUpdateMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Function *v1.FunctionSpecs  `protobuf:"bytes,1,opt,name=function,proto3" json:"function,omitempty"`
	Status   v11.FunctionStatus `protobuf:"varint,2,opt,name=status,proto3,enum=apollo.proto.frontend.v1.FunctionStatus" json:"status,omitempty"`
}

func (x *FunctionStatusUpdateMessage) Reset() {
	*x = FunctionStatusUpdateMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FunctionStatusUpdateMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FunctionStatusUpdateMessage) ProtoMessage() {}

func (x *FunctionStatusUpdateMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FunctionStatusUpdateMessage.ProtoReflect.Descriptor instead.
func (*FunctionStatusUpdateMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{3}
}

func (x *FunctionStatusUpdateMessage) GetFunction() *v1.FunctionSpecs {
	if x != nil {
		return x.Function
	}
	return nil
}

func (x *FunctionStatusUpdateMessage) GetStatus() v11.FunctionStatus {
	if x != nil {
		return x.Status
	}
	return v11.FunctionStatus(0)
}

type RunnerAgentReadyMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FunctionIdentifier string `protobuf:"bytes,1,opt,name=function_identifier,json=functionIdentifier,proto3" json:"function_identifier,omitempty"`
	RunnerUuid         string `protobuf:"bytes,2,opt,name=runner_uuid,json=runnerUuid,proto3" json:"runner_uuid,omitempty"`
	Reason             string `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Success            bool   `protobuf:"varint,4,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *RunnerAgentReadyMessage) Reset() {
	*x = RunnerAgentReadyMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunnerAgentReadyMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerAgentReadyMessage) ProtoMessage() {}

func (x *RunnerAgentReadyMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunnerAgentReadyMessage.ProtoReflect.Descriptor instead.
func (*RunnerAgentReadyMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{4}
}

func (x *RunnerAgentReadyMessage) GetFunctionIdentifier() string {
	if x != nil {
		return x.FunctionIdentifier
	}
	return ""
}

func (x *RunnerAgentReadyMessage) GetRunnerUuid() string {
	if x != nil {
		return x.RunnerUuid
	}
	return ""
}

func (x *RunnerAgentReadyMessage) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *RunnerAgentReadyMessage) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type InstanceHeartbeatMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceUuid string           `protobuf:"bytes,1,opt,name=instance_uuid,json=instanceUuid,proto3" json:"instance_uuid,omitempty"`
	InstanceType v12.InstanceType `protobuf:"varint,2,opt,name=instance_type,json=instanceType,proto3,enum=apollo.proto.registry.v1.InstanceType" json:"instance_type,omitempty"`
	// Types that are assignable to Metrics:
	//
	//	*InstanceHeartbeatMessage_ServiceInstanceMetrics
	//	*InstanceHeartbeatMessage_WorkerInstanceMetrics
	Metrics isInstanceHeartbeatMessage_Metrics `protobuf_oneof:"metrics"`
}

func (x *InstanceHeartbeatMessage) Reset() {
	*x = InstanceHeartbeatMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InstanceHeartbeatMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InstanceHeartbeatMessage) ProtoMessage() {}

func (x *InstanceHeartbeatMessage) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_messages_v1_messages_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InstanceHeartbeatMessage.ProtoReflect.Descriptor instead.
func (*InstanceHeartbeatMessage) Descriptor() ([]byte, []int) {
	return file_apollo_proto_messages_v1_messages_proto_rawDescGZIP(), []int{5}
}

func (x *InstanceHeartbeatMessage) GetInstanceUuid() string {
	if x != nil {
		return x.InstanceUuid
	}
	return ""
}

func (x *InstanceHeartbeatMessage) GetInstanceType() v12.InstanceType {
	if x != nil {
		return x.InstanceType
	}
	return v12.InstanceType(0)
}

func (m *InstanceHeartbeatMessage) GetMetrics() isInstanceHeartbeatMessage_Metrics {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func (x *InstanceHeartbeatMessage) GetServiceInstanceMetrics() *v12.ServiceInstanceMetrics {
	if x, ok := x.GetMetrics().(*InstanceHeartbeatMessage_ServiceInstanceMetrics); ok {
		return x.ServiceInstanceMetrics
	}
	return nil
}

func (x *InstanceHeartbeatMessage) GetWorkerInstanceMetrics() *v12.WorkerInstanceMetrics {
	if x, ok := x.GetMetrics().(*InstanceHeartbeatMessage_WorkerInstanceMetrics); ok {
		return x.WorkerInstanceMetrics
	}
	return nil
}

type isInstanceHeartbeatMessage_Metrics interface {
	isInstanceHeartbeatMessage_Metrics()
}

type InstanceHeartbeatMessage_ServiceInstanceMetrics struct {
	ServiceInstanceMetrics *v12.ServiceInstanceMetrics `protobuf:"bytes,3,opt,name=service_instance_metrics,json=serviceInstanceMetrics,proto3,oneof"`
}

type InstanceHeartbeatMessage_WorkerInstanceMetrics struct {
	WorkerInstanceMetrics *v12.WorkerInstanceMetrics `protobuf:"bytes,4,opt,name=worker_instance_metrics,json=workerInstanceMetrics,proto3,oneof"`
}

func (*InstanceHeartbeatMessage_ServiceInstanceMetrics) isInstanceHeartbeatMessage_Metrics() {}

func (*InstanceHeartbeatMessage_WorkerInstanceMetrics) isInstanceHeartbeatMessage_Metrics() {}

var File_apollo_proto_messages_v1_messages_proto protoreflect.FileDescriptor

var file_apollo_proto_messages_v1_messages_proto_rawDesc = []byte{
	0x0a, 0x27, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x61, 0x70, 0x6f, 0x6c, 0x6c,
	0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x2e, 0x76, 0x31, 0x1a, 0x27, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x27, 0x61, 0x70,
	0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x72, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x6c, 0x65,
	0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb4, 0x01, 0x0a, 0x1d, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x40, 0x0a, 0x08, 0x66, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x61,
	0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65,
	0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b,
	0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x55, 0x75, 0x69, 0x64, 0x12, 0x16, 0x0a,
	0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22,
	0xb6, 0x01, 0x0a, 0x1f, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x69, 0x6e,
	0x69, 0x74, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x40, 0x0a, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75,
	0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f,
	0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b,
	0x65, 0x72, 0x55, 0x75, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18,
	0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0xa1, 0x01, 0x0a, 0x1e, 0x46, 0x75, 0x6e,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x40, 0x0a, 0x08, 0x66,
	0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e,
	0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c, 0x65,
	0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70,
	0x65, 0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3d, 0x0a,
	0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23,
	0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c,
	0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x53, 0x70,
	0x65, 0x63, 0x73, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x22, 0xa1, 0x01, 0x0a,
	0x1b, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x40, 0x0a, 0x08,
	0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24,
	0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x6c,
	0x65, 0x65, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x70, 0x65, 0x63, 0x73, 0x52, 0x08, 0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x40,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28,
	0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x66, 0x72,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x75, 0x6e, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x22, 0x9d, 0x01, 0x0a, 0x17, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x41, 0x67, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x61, 0x64, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2f, 0x0a, 0x13,
	0x66, 0x75, 0x6e, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66,
	0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x66, 0x75, 0x6e, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x1f, 0x0a,
	0x0b, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x5f, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x55, 0x75, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x22, 0xf0, 0x02, 0x0a, 0x18, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x48, 0x65, 0x61,
	0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x23, 0x0a,
	0x0d, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x55, 0x75,
	0x69, 0x64, 0x12, 0x4b, 0x0a, 0x0d, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x61, 0x70, 0x6f, 0x6c,
	0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x0c, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x6c, 0x0a, 0x18, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x30, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x48, 0x00, 0x52, 0x16, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x69, 0x0a,
	0x17, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x5f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f,
	0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x65, 0x72,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x48,
	0x00, 0x52, 0x15, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x42, 0x09, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x64, 0x65, 0x6e, 0x6e, 0x69, 0x73, 0x68, 0x69, 0x6c, 0x67, 0x65, 0x72, 0x74, 0x2f,
	0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apollo_proto_messages_v1_messages_proto_rawDescOnce sync.Once
	file_apollo_proto_messages_v1_messages_proto_rawDescData = file_apollo_proto_messages_v1_messages_proto_rawDesc
)

func file_apollo_proto_messages_v1_messages_proto_rawDescGZIP() []byte {
	file_apollo_proto_messages_v1_messages_proto_rawDescOnce.Do(func() {
		file_apollo_proto_messages_v1_messages_proto_rawDescData = protoimpl.X.CompressGZIP(file_apollo_proto_messages_v1_messages_proto_rawDescData)
	})
	return file_apollo_proto_messages_v1_messages_proto_rawDescData
}

var file_apollo_proto_messages_v1_messages_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_apollo_proto_messages_v1_messages_proto_goTypes = []interface{}{
	(*FunctionInitializationMessage)(nil),   // 0: apollo.proto.messages.v1.FunctionInitializationMessage
	(*FunctionDeinitializationMessage)(nil), // 1: apollo.proto.messages.v1.FunctionDeinitializationMessage
	(*FunctionPackageCreationMessage)(nil),  // 2: apollo.proto.messages.v1.FunctionPackageCreationMessage
	(*FunctionStatusUpdateMessage)(nil),     // 3: apollo.proto.messages.v1.FunctionStatusUpdateMessage
	(*RunnerAgentReadyMessage)(nil),         // 4: apollo.proto.messages.v1.RunnerAgentReadyMessage
	(*InstanceHeartbeatMessage)(nil),        // 5: apollo.proto.messages.v1.InstanceHeartbeatMessage
	(*v1.FunctionSpecs)(nil),                // 6: apollo.proto.fleet.v1.FunctionSpecs
	(*v1.RuntimeSpecs)(nil),                 // 7: apollo.proto.fleet.v1.RuntimeSpecs
	(v11.FunctionStatus)(0),                 // 8: apollo.proto.frontend.v1.FunctionStatus
	(v12.InstanceType)(0),                   // 9: apollo.proto.registry.v1.InstanceType
	(*v12.ServiceInstanceMetrics)(nil),      // 10: apollo.proto.registry.v1.ServiceInstanceMetrics
	(*v12.WorkerInstanceMetrics)(nil),       // 11: apollo.proto.registry.v1.WorkerInstanceMetrics
}
var file_apollo_proto_messages_v1_messages_proto_depIdxs = []int32{
	6,  // 0: apollo.proto.messages.v1.FunctionInitializationMessage.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	6,  // 1: apollo.proto.messages.v1.FunctionDeinitializationMessage.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	6,  // 2: apollo.proto.messages.v1.FunctionPackageCreationMessage.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	7,  // 3: apollo.proto.messages.v1.FunctionPackageCreationMessage.runtime:type_name -> apollo.proto.fleet.v1.RuntimeSpecs
	6,  // 4: apollo.proto.messages.v1.FunctionStatusUpdateMessage.function:type_name -> apollo.proto.fleet.v1.FunctionSpecs
	8,  // 5: apollo.proto.messages.v1.FunctionStatusUpdateMessage.status:type_name -> apollo.proto.frontend.v1.FunctionStatus
	9,  // 6: apollo.proto.messages.v1.InstanceHeartbeatMessage.instance_type:type_name -> apollo.proto.registry.v1.InstanceType
	10, // 7: apollo.proto.messages.v1.InstanceHeartbeatMessage.service_instance_metrics:type_name -> apollo.proto.registry.v1.ServiceInstanceMetrics
	11, // 8: apollo.proto.messages.v1.InstanceHeartbeatMessage.worker_instance_metrics:type_name -> apollo.proto.registry.v1.WorkerInstanceMetrics
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_apollo_proto_messages_v1_messages_proto_init() }
func file_apollo_proto_messages_v1_messages_proto_init() {
	if File_apollo_proto_messages_v1_messages_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apollo_proto_messages_v1_messages_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionInitializationMessage); i {
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
		file_apollo_proto_messages_v1_messages_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionDeinitializationMessage); i {
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
		file_apollo_proto_messages_v1_messages_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionPackageCreationMessage); i {
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
		file_apollo_proto_messages_v1_messages_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FunctionStatusUpdateMessage); i {
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
		file_apollo_proto_messages_v1_messages_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunnerAgentReadyMessage); i {
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
		file_apollo_proto_messages_v1_messages_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InstanceHeartbeatMessage); i {
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
	file_apollo_proto_messages_v1_messages_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*InstanceHeartbeatMessage_ServiceInstanceMetrics)(nil),
		(*InstanceHeartbeatMessage_WorkerInstanceMetrics)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_apollo_proto_messages_v1_messages_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_apollo_proto_messages_v1_messages_proto_goTypes,
		DependencyIndexes: file_apollo_proto_messages_v1_messages_proto_depIdxs,
		MessageInfos:      file_apollo_proto_messages_v1_messages_proto_msgTypes,
	}.Build()
	File_apollo_proto_messages_v1_messages_proto = out.File
	file_apollo_proto_messages_v1_messages_proto_rawDesc = nil
	file_apollo_proto_messages_v1_messages_proto_goTypes = nil
	file_apollo_proto_messages_v1_messages_proto_depIdxs = nil
}
