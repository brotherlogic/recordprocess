// Code generated by protoc-gen-go. DO NOT EDIT.
// source: recordprocess.proto

package recordprocessor

import (
	fmt "fmt"
	proto1 "github.com/brotherlogic/recordcollection/proto"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Scores struct {
	Scores               []*RecordScore `protobuf:"bytes,1,rep,name=scores,proto3" json:"scores,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Scores) Reset()         { *m = Scores{} }
func (m *Scores) String() string { return proto.CompactTextString(m) }
func (*Scores) ProtoMessage()    {}
func (*Scores) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{0}
}

func (m *Scores) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Scores.Unmarshal(m, b)
}
func (m *Scores) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Scores.Marshal(b, m, deterministic)
}
func (m *Scores) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Scores.Merge(m, src)
}
func (m *Scores) XXX_Size() int {
	return xxx_messageInfo_Scores.Size(m)
}
func (m *Scores) XXX_DiscardUnknown() {
	xxx_messageInfo_Scores.DiscardUnknown(m)
}

var xxx_messageInfo_Scores proto.InternalMessageInfo

func (m *Scores) GetScores() []*RecordScore {
	if m != nil {
		return m.Scores
	}
	return nil
}

type Config struct {
	LastRunTime          int64           `protobuf:"varint,1,opt,name=last_run_time,json=lastRunTime,proto3" json:"last_run_time,omitempty"`
	NextUpdateTime       map[int32]int64 `protobuf:"bytes,2,rep,name=next_update_time,json=nextUpdateTime,proto3" json:"next_update_time,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{1}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetLastRunTime() int64 {
	if m != nil {
		return m.LastRunTime
	}
	return 0
}

func (m *Config) GetNextUpdateTime() map[int32]int64 {
	if m != nil {
		return m.NextUpdateTime
	}
	return nil
}

type RecordScore struct {
	InstanceId           int32                           `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	Rating               int32                           `protobuf:"varint,2,opt,name=rating,proto3" json:"rating,omitempty"`
	Category             proto1.ReleaseMetadata_Category `protobuf:"varint,3,opt,name=category,proto3,enum=recordcollection.ReleaseMetadata_Category" json:"category,omitempty"`
	ScoreTime            int64                           `protobuf:"varint,4,opt,name=score_time,json=scoreTime,proto3" json:"score_time,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                        `json:"-"`
	XXX_unrecognized     []byte                          `json:"-"`
	XXX_sizecache        int32                           `json:"-"`
}

func (m *RecordScore) Reset()         { *m = RecordScore{} }
func (m *RecordScore) String() string { return proto.CompactTextString(m) }
func (*RecordScore) ProtoMessage()    {}
func (*RecordScore) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{2}
}

func (m *RecordScore) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RecordScore.Unmarshal(m, b)
}
func (m *RecordScore) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RecordScore.Marshal(b, m, deterministic)
}
func (m *RecordScore) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RecordScore.Merge(m, src)
}
func (m *RecordScore) XXX_Size() int {
	return xxx_messageInfo_RecordScore.Size(m)
}
func (m *RecordScore) XXX_DiscardUnknown() {
	xxx_messageInfo_RecordScore.DiscardUnknown(m)
}

var xxx_messageInfo_RecordScore proto.InternalMessageInfo

func (m *RecordScore) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

func (m *RecordScore) GetRating() int32 {
	if m != nil {
		return m.Rating
	}
	return 0
}

func (m *RecordScore) GetCategory() proto1.ReleaseMetadata_Category {
	if m != nil {
		return m.Category
	}
	return proto1.ReleaseMetadata_UNKNOWN
}

func (m *RecordScore) GetScoreTime() int64 {
	if m != nil {
		return m.ScoreTime
	}
	return 0
}

type GetScoreRequest struct {
	InstanceId           int32    `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetScoreRequest) Reset()         { *m = GetScoreRequest{} }
func (m *GetScoreRequest) String() string { return proto.CompactTextString(m) }
func (*GetScoreRequest) ProtoMessage()    {}
func (*GetScoreRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{3}
}

func (m *GetScoreRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetScoreRequest.Unmarshal(m, b)
}
func (m *GetScoreRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetScoreRequest.Marshal(b, m, deterministic)
}
func (m *GetScoreRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetScoreRequest.Merge(m, src)
}
func (m *GetScoreRequest) XXX_Size() int {
	return xxx_messageInfo_GetScoreRequest.Size(m)
}
func (m *GetScoreRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetScoreRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetScoreRequest proto.InternalMessageInfo

func (m *GetScoreRequest) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

type GetScoreResponse struct {
	Scores               []*RecordScore `protobuf:"bytes,1,rep,name=scores,proto3" json:"scores,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *GetScoreResponse) Reset()         { *m = GetScoreResponse{} }
func (m *GetScoreResponse) String() string { return proto.CompactTextString(m) }
func (*GetScoreResponse) ProtoMessage()    {}
func (*GetScoreResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{4}
}

func (m *GetScoreResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetScoreResponse.Unmarshal(m, b)
}
func (m *GetScoreResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetScoreResponse.Marshal(b, m, deterministic)
}
func (m *GetScoreResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetScoreResponse.Merge(m, src)
}
func (m *GetScoreResponse) XXX_Size() int {
	return xxx_messageInfo_GetScoreResponse.Size(m)
}
func (m *GetScoreResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetScoreResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetScoreResponse proto.InternalMessageInfo

func (m *GetScoreResponse) GetScores() []*RecordScore {
	if m != nil {
		return m.Scores
	}
	return nil
}

type ForceRequest struct {
	InstanceId           int32    `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ForceRequest) Reset()         { *m = ForceRequest{} }
func (m *ForceRequest) String() string { return proto.CompactTextString(m) }
func (*ForceRequest) ProtoMessage()    {}
func (*ForceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{5}
}

func (m *ForceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForceRequest.Unmarshal(m, b)
}
func (m *ForceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForceRequest.Marshal(b, m, deterministic)
}
func (m *ForceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForceRequest.Merge(m, src)
}
func (m *ForceRequest) XXX_Size() int {
	return xxx_messageInfo_ForceRequest.Size(m)
}
func (m *ForceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ForceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ForceRequest proto.InternalMessageInfo

func (m *ForceRequest) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

type ForceResponse struct {
	Result               *proto1.Record `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ForceResponse) Reset()         { *m = ForceResponse{} }
func (m *ForceResponse) String() string { return proto.CompactTextString(m) }
func (*ForceResponse) ProtoMessage()    {}
func (*ForceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_af3f0d6f9a1758de, []int{6}
}

func (m *ForceResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForceResponse.Unmarshal(m, b)
}
func (m *ForceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForceResponse.Marshal(b, m, deterministic)
}
func (m *ForceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForceResponse.Merge(m, src)
}
func (m *ForceResponse) XXX_Size() int {
	return xxx_messageInfo_ForceResponse.Size(m)
}
func (m *ForceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ForceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ForceResponse proto.InternalMessageInfo

func (m *ForceResponse) GetResult() *proto1.Record {
	if m != nil {
		return m.Result
	}
	return nil
}

func init() {
	proto.RegisterType((*Scores)(nil), "recordprocessor.Scores")
	proto.RegisterType((*Config)(nil), "recordprocessor.Config")
	proto.RegisterMapType((map[int32]int64)(nil), "recordprocessor.Config.NextUpdateTimeEntry")
	proto.RegisterType((*RecordScore)(nil), "recordprocessor.RecordScore")
	proto.RegisterType((*GetScoreRequest)(nil), "recordprocessor.GetScoreRequest")
	proto.RegisterType((*GetScoreResponse)(nil), "recordprocessor.GetScoreResponse")
	proto.RegisterType((*ForceRequest)(nil), "recordprocessor.ForceRequest")
	proto.RegisterType((*ForceResponse)(nil), "recordprocessor.ForceResponse")
}

func init() { proto.RegisterFile("recordprocess.proto", fileDescriptor_af3f0d6f9a1758de) }

var fileDescriptor_af3f0d6f9a1758de = []byte{
	// 460 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0x5d, 0x6f, 0xd3, 0x30,
	0x14, 0x25, 0x2b, 0x8d, 0xc6, 0xcd, 0x3e, 0x2a, 0x0f, 0xa1, 0xaa, 0x62, 0x50, 0xf2, 0x54, 0x81,
	0x94, 0xa2, 0xc0, 0x03, 0xe2, 0x01, 0x69, 0x9a, 0x36, 0xc6, 0x03, 0x48, 0x78, 0xec, 0xb9, 0x72,
	0x9d, 0x4b, 0x66, 0x91, 0xda, 0xc1, 0xbe, 0x99, 0xd6, 0xdf, 0x84, 0xf8, 0x1b, 0xfc, 0x2e, 0x14,
	0x27, 0x85, 0x6d, 0x29, 0x1f, 0xda, 0x9b, 0x7d, 0x72, 0xee, 0x39, 0xe7, 0x1e, 0x2b, 0xb0, 0x67,
	0x51, 0x1a, 0x9b, 0x95, 0xd6, 0x48, 0x74, 0x2e, 0x29, 0xad, 0x21, 0xc3, 0x76, 0xaf, 0x81, 0xc6,
	0x8e, 0x8e, 0x72, 0x45, 0xe7, 0xd5, 0x3c, 0x91, 0x66, 0x31, 0x9d, 0x5b, 0x43, 0xe7, 0x68, 0x0b,
	0x93, 0x2b, 0x39, 0x6d, 0x88, 0xd2, 0x14, 0x05, 0x4a, 0x52, 0x46, 0x4f, 0xbd, 0x40, 0x07, 0x6e,
	0x74, 0xe3, 0x37, 0x10, 0x9e, 0x4a, 0x63, 0xd1, 0xb1, 0x97, 0x10, 0x3a, 0x7f, 0x1a, 0x06, 0xe3,
	0xde, 0x24, 0x4a, 0x1f, 0x26, 0x37, 0x2c, 0x13, 0xee, 0xef, 0x9e, 0xce, 0x5b, 0x6e, 0xfc, 0x23,
	0x80, 0xf0, 0xd0, 0xe8, 0xcf, 0x2a, 0x67, 0x31, 0x6c, 0x17, 0xc2, 0xd1, 0xcc, 0x56, 0x7a, 0x46,
	0x6a, 0x81, 0xc3, 0x60, 0x1c, 0x4c, 0x7a, 0x3c, 0xaa, 0x41, 0x5e, 0xe9, 0x4f, 0x6a, 0x81, 0xec,
	0x0c, 0x06, 0x1a, 0x2f, 0x69, 0x56, 0x95, 0x99, 0x20, 0x6c, 0x68, 0x1b, 0xde, 0xee, 0x59, 0xc7,
	0xae, 0x91, 0x4d, 0x3e, 0xe0, 0x25, 0x9d, 0x79, 0x7a, 0xad, 0x70, 0xa4, 0xc9, 0x2e, 0xf9, 0x8e,
	0xbe, 0x06, 0x8e, 0x0e, 0x60, 0x6f, 0x0d, 0x8d, 0x0d, 0xa0, 0xf7, 0x05, 0x97, 0x3e, 0x47, 0x9f,
	0xd7, 0x47, 0x76, 0x1f, 0xfa, 0x17, 0xa2, 0xa8, 0x6a, 0xd3, 0x3a, 0x5b, 0x73, 0x79, 0xbd, 0xf1,
	0x2a, 0x88, 0xbf, 0x07, 0x10, 0x5d, 0x59, 0x90, 0x3d, 0x86, 0x48, 0x69, 0x47, 0x42, 0x4b, 0x9c,
	0xa9, 0xac, 0xd5, 0x80, 0x15, 0xf4, 0x2e, 0x63, 0x0f, 0x20, 0xb4, 0x82, 0x94, 0xce, 0xbd, 0x56,
	0x9f, 0xb7, 0x37, 0x76, 0x0c, 0x9b, 0x52, 0x10, 0xe6, 0xc6, 0x2e, 0x87, 0xbd, 0x71, 0x30, 0xd9,
	0x49, 0x9f, 0x26, 0x9d, 0xf2, 0x39, 0x16, 0x28, 0x1c, 0xbe, 0x47, 0x12, 0x99, 0x20, 0x91, 0x1c,
	0xb6, 0x13, 0xfc, 0xd7, 0x2c, 0xdb, 0x07, 0xf0, 0x1d, 0x37, 0x25, 0xdd, 0xf5, 0x79, 0xef, 0x79,
	0xa4, 0x5e, 0x30, 0x4e, 0x61, 0xf7, 0x2d, 0x52, 0xf3, 0x18, 0xf8, 0xb5, 0x42, 0x47, 0xff, 0x8c,
	0x1c, 0x9f, 0xc0, 0xe0, 0xf7, 0x8c, 0x2b, 0x8d, 0x76, 0x78, 0xcb, 0x67, 0x9f, 0xc2, 0xd6, 0xb1,
	0xb1, 0xf2, 0xff, 0xad, 0x0f, 0x60, 0xbb, 0x1d, 0x68, 0x7d, 0x9f, 0x43, 0x68, 0xd1, 0x55, 0x05,
	0x79, 0x72, 0x94, 0x0e, 0xd7, 0x95, 0x54, 0x03, 0xbc, 0xe5, 0xa5, 0xdf, 0x02, 0xd8, 0xf2, 0x29,
	0x4e, 0xd1, 0x5e, 0x28, 0x89, 0xec, 0x23, 0x6c, 0xae, 0xd6, 0x61, 0xe3, 0x4e, 0xec, 0x1b, 0xed,
	0x8c, 0x9e, 0xfc, 0x85, 0xd1, 0x64, 0x8a, 0xef, 0xb0, 0x13, 0xe8, 0xfb, 0x98, 0x6c, 0xbf, 0xc3,
	0xbe, 0xba, 0xef, 0xe8, 0xd1, 0x9f, 0x3e, 0xaf, 0x94, 0xe6, 0xa1, 0xff, 0xbf, 0x5e, 0xfc, 0x0c,
	0x00, 0x00, 0xff, 0xff, 0x79, 0xf8, 0x85, 0xfd, 0xce, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ScoreServiceClient is the client API for ScoreService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ScoreServiceClient interface {
	GetScore(ctx context.Context, in *GetScoreRequest, opts ...grpc.CallOption) (*GetScoreResponse, error)
	Force(ctx context.Context, in *ForceRequest, opts ...grpc.CallOption) (*ForceResponse, error)
}

type scoreServiceClient struct {
	cc *grpc.ClientConn
}

func NewScoreServiceClient(cc *grpc.ClientConn) ScoreServiceClient {
	return &scoreServiceClient{cc}
}

func (c *scoreServiceClient) GetScore(ctx context.Context, in *GetScoreRequest, opts ...grpc.CallOption) (*GetScoreResponse, error) {
	out := new(GetScoreResponse)
	err := c.cc.Invoke(ctx, "/recordprocessor.ScoreService/GetScore", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scoreServiceClient) Force(ctx context.Context, in *ForceRequest, opts ...grpc.CallOption) (*ForceResponse, error) {
	out := new(ForceResponse)
	err := c.cc.Invoke(ctx, "/recordprocessor.ScoreService/Force", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ScoreServiceServer is the server API for ScoreService service.
type ScoreServiceServer interface {
	GetScore(context.Context, *GetScoreRequest) (*GetScoreResponse, error)
	Force(context.Context, *ForceRequest) (*ForceResponse, error)
}

func RegisterScoreServiceServer(s *grpc.Server, srv ScoreServiceServer) {
	s.RegisterService(&_ScoreService_serviceDesc, srv)
}

func _ScoreService_GetScore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetScoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScoreServiceServer).GetScore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordprocessor.ScoreService/GetScore",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScoreServiceServer).GetScore(ctx, req.(*GetScoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ScoreService_Force_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ForceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScoreServiceServer).Force(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordprocessor.ScoreService/Force",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScoreServiceServer).Force(ctx, req.(*ForceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ScoreService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "recordprocessor.ScoreService",
	HandlerType: (*ScoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetScore",
			Handler:    _ScoreService_GetScore_Handler,
		},
		{
			MethodName: "Force",
			Handler:    _ScoreService_Force_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "recordprocess.proto",
}
