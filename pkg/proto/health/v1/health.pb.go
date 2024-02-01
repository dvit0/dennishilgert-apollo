// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: apollo/proto/health/v1/health.proto

package health

import (
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

type HealthStatus int32

const (
	HealthStatus_HEALTH_STATUS_UNSPECIFIED HealthStatus = 0
	HealthStatus_HEALTHY                   HealthStatus = 1
	HealthStatus_UNHEALTHY                 HealthStatus = 2
)

// Enum value maps for HealthStatus.
var (
	HealthStatus_name = map[int32]string{
		0: "HEALTH_STATUS_UNSPECIFIED",
		1: "HEALTHY",
		2: "UNHEALTHY",
	}
	HealthStatus_value = map[string]int32{
		"HEALTH_STATUS_UNSPECIFIED": 0,
		"HEALTHY":                   1,
		"UNHEALTHY":                 2,
	}
)

func (x HealthStatus) Enum() *HealthStatus {
	p := new(HealthStatus)
	*p = x
	return p
}

func (x HealthStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HealthStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_apollo_proto_health_v1_health_proto_enumTypes[0].Descriptor()
}

func (HealthStatus) Type() protoreflect.EnumType {
	return &file_apollo_proto_health_v1_health_proto_enumTypes[0]
}

func (x HealthStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HealthStatus.Descriptor instead.
func (HealthStatus) EnumDescriptor() ([]byte, []int) {
	return file_apollo_proto_health_v1_health_proto_rawDescGZIP(), []int{0}
}

type HealthStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HealthStatusRequest) Reset() {
	*x = HealthStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_health_v1_health_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthStatusRequest) ProtoMessage() {}

func (x *HealthStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_health_v1_health_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthStatusRequest.ProtoReflect.Descriptor instead.
func (*HealthStatusRequest) Descriptor() ([]byte, []int) {
	return file_apollo_proto_health_v1_health_proto_rawDescGZIP(), []int{0}
}

type HealthStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status HealthStatus `protobuf:"varint,1,opt,name=status,proto3,enum=apollo.proto.health.v1.HealthStatus" json:"status,omitempty"`
}

func (x *HealthStatusResponse) Reset() {
	*x = HealthStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apollo_proto_health_v1_health_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthStatusResponse) ProtoMessage() {}

func (x *HealthStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_apollo_proto_health_v1_health_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthStatusResponse.ProtoReflect.Descriptor instead.
func (*HealthStatusResponse) Descriptor() ([]byte, []int) {
	return file_apollo_proto_health_v1_health_proto_rawDescGZIP(), []int{1}
}

func (x *HealthStatusResponse) GetStatus() HealthStatus {
	if x != nil {
		return x.Status
	}
	return HealthStatus_HEALTH_STATUS_UNSPECIFIED
}

var File_apollo_proto_health_v1_health_proto protoreflect.FileDescriptor

var file_apollo_proto_health_v1_health_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x68,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x22, 0x15, 0x0a,
	0x13, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x54, 0x0a, 0x14, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3c, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x61,
	0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x68, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0x49, 0x0a, 0x0c, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x19, 0x48, 0x45,
	0x41, 0x4c, 0x54, 0x48, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x48, 0x45, 0x41,
	0x4c, 0x54, 0x48, 0x59, 0x10, 0x01, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x48, 0x45, 0x41, 0x4c,
	0x54, 0x48, 0x59, 0x10, 0x02, 0x32, 0x6f, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12,
	0x65, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2b, 0x2e, 0x61, 0x70, 0x6f, 0x6c,
	0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e,
	0x76, 0x31, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x76, 0x31, 0x2e,
	0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x6e, 0x6e, 0x69, 0x73, 0x68, 0x69, 0x6c, 0x67, 0x65,
	0x72, 0x74, 0x2f, 0x61, 0x70, 0x6f, 0x6c, 0x6c, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2f, 0x76, 0x31, 0x3b, 0x68, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apollo_proto_health_v1_health_proto_rawDescOnce sync.Once
	file_apollo_proto_health_v1_health_proto_rawDescData = file_apollo_proto_health_v1_health_proto_rawDesc
)

func file_apollo_proto_health_v1_health_proto_rawDescGZIP() []byte {
	file_apollo_proto_health_v1_health_proto_rawDescOnce.Do(func() {
		file_apollo_proto_health_v1_health_proto_rawDescData = protoimpl.X.CompressGZIP(file_apollo_proto_health_v1_health_proto_rawDescData)
	})
	return file_apollo_proto_health_v1_health_proto_rawDescData
}

var file_apollo_proto_health_v1_health_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_apollo_proto_health_v1_health_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_apollo_proto_health_v1_health_proto_goTypes = []interface{}{
	(HealthStatus)(0),            // 0: apollo.proto.health.v1.HealthStatus
	(*HealthStatusRequest)(nil),  // 1: apollo.proto.health.v1.HealthStatusRequest
	(*HealthStatusResponse)(nil), // 2: apollo.proto.health.v1.HealthStatusResponse
}
var file_apollo_proto_health_v1_health_proto_depIdxs = []int32{
	0, // 0: apollo.proto.health.v1.HealthStatusResponse.status:type_name -> apollo.proto.health.v1.HealthStatus
	1, // 1: apollo.proto.health.v1.Health.Status:input_type -> apollo.proto.health.v1.HealthStatusRequest
	2, // 2: apollo.proto.health.v1.Health.Status:output_type -> apollo.proto.health.v1.HealthStatusResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_apollo_proto_health_v1_health_proto_init() }
func file_apollo_proto_health_v1_health_proto_init() {
	if File_apollo_proto_health_v1_health_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apollo_proto_health_v1_health_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthStatusRequest); i {
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
		file_apollo_proto_health_v1_health_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthStatusResponse); i {
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
			RawDescriptor: file_apollo_proto_health_v1_health_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_apollo_proto_health_v1_health_proto_goTypes,
		DependencyIndexes: file_apollo_proto_health_v1_health_proto_depIdxs,
		EnumInfos:         file_apollo_proto_health_v1_health_proto_enumTypes,
		MessageInfos:      file_apollo_proto_health_v1_health_proto_msgTypes,
	}.Build()
	File_apollo_proto_health_v1_health_proto = out.File
	file_apollo_proto_health_v1_health_proto_rawDesc = nil
	file_apollo_proto_health_v1_health_proto_goTypes = nil
	file_apollo_proto_health_v1_health_proto_depIdxs = nil
}
