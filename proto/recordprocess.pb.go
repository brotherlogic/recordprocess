// Code generated by protoc-gen-go. DO NOT EDIT.
// source: recordprocess.proto

/*
Package recordprocessor is a generated protocol buffer package.

It is generated from these files:
	recordprocess.proto

It has these top-level messages:
	Scores
	RecordScore
	GetScoreRequest
	GetScoreResponse
*/
package recordprocessor

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import recordcollection "github.com/brotherlogic/recordcollection/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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
	Scores []*RecordScore `protobuf:"bytes,1,rep,name=scores" json:"scores,omitempty"`
}

func (m *Scores) Reset()                    { *m = Scores{} }
func (m *Scores) String() string            { return proto.CompactTextString(m) }
func (*Scores) ProtoMessage()               {}
func (*Scores) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Scores) GetScores() []*RecordScore {
	if m != nil {
		return m.Scores
	}
	return nil
}

type RecordScore struct {
	InstanceId int32                                     `protobuf:"varint,1,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
	Rating     int32                                     `protobuf:"varint,2,opt,name=rating" json:"rating,omitempty"`
	Category   recordcollection.ReleaseMetadata_Category `protobuf:"varint,3,opt,name=category,enum=recordcollection.ReleaseMetadata_Category" json:"category,omitempty"`
	ScoreTime  int64                                     `protobuf:"varint,4,opt,name=score_time,json=scoreTime" json:"score_time,omitempty"`
}

func (m *RecordScore) Reset()                    { *m = RecordScore{} }
func (m *RecordScore) String() string            { return proto.CompactTextString(m) }
func (*RecordScore) ProtoMessage()               {}
func (*RecordScore) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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

func (m *RecordScore) GetCategory() recordcollection.ReleaseMetadata_Category {
	if m != nil {
		return m.Category
	}
	return recordcollection.ReleaseMetadata_UNKNOWN
}

func (m *RecordScore) GetScoreTime() int64 {
	if m != nil {
		return m.ScoreTime
	}
	return 0
}

type GetScoreRequest struct {
	InstanceId int32 `protobuf:"varint,1,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *GetScoreRequest) Reset()                    { *m = GetScoreRequest{} }
func (m *GetScoreRequest) String() string            { return proto.CompactTextString(m) }
func (*GetScoreRequest) ProtoMessage()               {}
func (*GetScoreRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *GetScoreRequest) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

type GetScoreResponse struct {
	Scores []*RecordScore `protobuf:"bytes,1,rep,name=scores" json:"scores,omitempty"`
}

func (m *GetScoreResponse) Reset()                    { *m = GetScoreResponse{} }
func (m *GetScoreResponse) String() string            { return proto.CompactTextString(m) }
func (*GetScoreResponse) ProtoMessage()               {}
func (*GetScoreResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetScoreResponse) GetScores() []*RecordScore {
	if m != nil {
		return m.Scores
	}
	return nil
}

func init() {
	proto.RegisterType((*Scores)(nil), "recordprocessor.Scores")
	proto.RegisterType((*RecordScore)(nil), "recordprocessor.RecordScore")
	proto.RegisterType((*GetScoreRequest)(nil), "recordprocessor.GetScoreRequest")
	proto.RegisterType((*GetScoreResponse)(nil), "recordprocessor.GetScoreResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ScoreService service

type ScoreServiceClient interface {
	GetScore(ctx context.Context, in *GetScoreRequest, opts ...grpc.CallOption) (*GetScoreResponse, error)
}

type scoreServiceClient struct {
	cc *grpc.ClientConn
}

func NewScoreServiceClient(cc *grpc.ClientConn) ScoreServiceClient {
	return &scoreServiceClient{cc}
}

func (c *scoreServiceClient) GetScore(ctx context.Context, in *GetScoreRequest, opts ...grpc.CallOption) (*GetScoreResponse, error) {
	out := new(GetScoreResponse)
	err := grpc.Invoke(ctx, "/recordprocessor.ScoreService/GetScore", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ScoreService service

type ScoreServiceServer interface {
	GetScore(context.Context, *GetScoreRequest) (*GetScoreResponse, error)
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

var _ScoreService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "recordprocessor.ScoreService",
	HandlerType: (*ScoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetScore",
			Handler:    _ScoreService_GetScore_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "recordprocess.proto",
}

func init() { proto.RegisterFile("recordprocess.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 306 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x91, 0x4d, 0x4b, 0x33, 0x31,
	0x14, 0x85, 0xdf, 0x79, 0xab, 0x43, 0xbd, 0x15, 0x2b, 0x11, 0x64, 0x28, 0x8a, 0xe3, 0xac, 0x06,
	0x17, 0x29, 0x8c, 0xae, 0xdd, 0x88, 0x5f, 0x0b, 0x17, 0xa6, 0xee, 0x4b, 0x9a, 0xb9, 0x4c, 0x03,
	0xd3, 0xb9, 0x35, 0xb9, 0x15, 0xfc, 0x51, 0xfe, 0x47, 0x21, 0xd3, 0xfa, 0xd1, 0x82, 0x82, 0xbb,
	0xdc, 0x27, 0xe7, 0x9c, 0x9c, 0x24, 0x70, 0xe0, 0xd0, 0x90, 0x2b, 0xe7, 0x8e, 0x0c, 0x7a, 0x2f,
	0xe7, 0x8e, 0x98, 0x44, 0xff, 0x1b, 0x24, 0x37, 0xb8, 0xae, 0x2c, 0x4f, 0x17, 0x13, 0x69, 0x68,
	0x36, 0x9c, 0x38, 0xe2, 0x29, 0xba, 0x9a, 0x2a, 0x6b, 0x86, 0xad, 0xd0, 0x50, 0x5d, 0xa3, 0x61,
	0x4b, 0xcd, 0x30, 0x04, 0x6c, 0xe0, 0x36, 0x37, 0xbb, 0x84, 0x78, 0x64, 0xc8, 0xa1, 0x17, 0x17,
	0x10, 0xfb, 0xb0, 0x4a, 0xa2, 0xb4, 0x93, 0xf7, 0x8a, 0x23, 0xb9, 0x76, 0xa4, 0x54, 0x61, 0x0e,
	0x72, 0xb5, 0xd4, 0x66, 0x6f, 0x11, 0xf4, 0xbe, 0x70, 0x71, 0x02, 0x3d, 0xdb, 0x78, 0xd6, 0x8d,
	0xc1, 0xb1, 0x2d, 0x93, 0x28, 0x8d, 0xf2, 0x6d, 0x05, 0x2b, 0x74, 0x5f, 0x8a, 0x43, 0x88, 0x9d,
	0x66, 0xdb, 0x54, 0xc9, 0xff, 0xb0, 0xb7, 0x9c, 0xc4, 0x0d, 0x74, 0x8d, 0x66, 0xac, 0xc8, 0xbd,
	0x26, 0x9d, 0x34, 0xca, 0xf7, 0x8a, 0x33, 0xb9, 0xd1, 0x59, 0x61, 0x8d, 0xda, 0xe3, 0x03, 0xb2,
	0x2e, 0x35, 0x6b, 0x79, 0xb5, 0x74, 0xa8, 0x0f, 0xaf, 0x38, 0x06, 0x08, 0xd5, 0xc6, 0x6c, 0x67,
	0x98, 0x6c, 0xa5, 0x51, 0xde, 0x51, 0x3b, 0x81, 0x3c, 0xd9, 0x19, 0x66, 0x05, 0xf4, 0x6f, 0x91,
	0xdb, 0x3b, 0xe0, 0xf3, 0x02, 0x3d, 0xff, 0x5a, 0x39, 0xbb, 0x83, 0xfd, 0x4f, 0x8f, 0x9f, 0x53,
	0xe3, 0xf1, 0x6f, 0xaf, 0x55, 0x68, 0xd8, 0x0d, 0x60, 0x84, 0xee, 0xc5, 0x1a, 0x14, 0x8f, 0xd0,
	0x5d, 0x25, 0x8b, 0x74, 0x23, 0x61, 0xad, 0xe8, 0xe0, 0xf4, 0x07, 0x45, 0x5b, 0x2b, 0xfb, 0x37,
	0x89, 0xc3, 0xbf, 0x9e, 0xbf, 0x07, 0x00, 0x00, 0xff, 0xff, 0xac, 0x5d, 0x82, 0xf3, 0x46, 0x02,
	0x00, 0x00,
}
