// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: net.proto

package net_pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Service_GetDocGraph_FullMethodName  = "/net.pb.Service/GetDocGraph"
	Service_PushDocGraph_FullMethodName = "/net.pb.Service/PushDocGraph"
	Service_GetLog_FullMethodName       = "/net.pb.Service/GetLog"
	Service_PushLog_FullMethodName      = "/net.pb.Service/PushLog"
	Service_GetHeadLog_FullMethodName   = "/net.pb.Service/GetHeadLog"
)

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceClient interface {
	// GetDocGraph from this peer.
	GetDocGraph(ctx context.Context, in *GetDocGraphRequest, opts ...grpc.CallOption) (*GetDocGraphReply, error)
	// PushDocGraph to this peer.
	PushDocGraph(ctx context.Context, in *PushDocGraphRequest, opts ...grpc.CallOption) (*PushDocGraphReply, error)
	// GetLog from this peer.
	GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (*GetLogReply, error)
	// PushLog to this peer.
	PushLog(ctx context.Context, in *PushLogRequest, opts ...grpc.CallOption) (*PushLogReply, error)
	// GetHeadLog from this peer
	GetHeadLog(ctx context.Context, in *GetHeadLogRequest, opts ...grpc.CallOption) (*GetHeadLogReply, error)
}

type serviceClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceClient(cc grpc.ClientConnInterface) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) GetDocGraph(ctx context.Context, in *GetDocGraphRequest, opts ...grpc.CallOption) (*GetDocGraphReply, error) {
	out := new(GetDocGraphReply)
	err := c.cc.Invoke(ctx, Service_GetDocGraph_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) PushDocGraph(ctx context.Context, in *PushDocGraphRequest, opts ...grpc.CallOption) (*PushDocGraphReply, error) {
	out := new(PushDocGraphReply)
	err := c.cc.Invoke(ctx, Service_PushDocGraph_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) GetLog(ctx context.Context, in *GetLogRequest, opts ...grpc.CallOption) (*GetLogReply, error) {
	out := new(GetLogReply)
	err := c.cc.Invoke(ctx, Service_GetLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) PushLog(ctx context.Context, in *PushLogRequest, opts ...grpc.CallOption) (*PushLogReply, error) {
	out := new(PushLogReply)
	err := c.cc.Invoke(ctx, Service_PushLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) GetHeadLog(ctx context.Context, in *GetHeadLogRequest, opts ...grpc.CallOption) (*GetHeadLogReply, error) {
	out := new(GetHeadLogReply)
	err := c.cc.Invoke(ctx, Service_GetHeadLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceServer is the server API for Service service.
// All implementations must embed UnimplementedServiceServer
// for forward compatibility
type ServiceServer interface {
	// GetDocGraph from this peer.
	GetDocGraph(context.Context, *GetDocGraphRequest) (*GetDocGraphReply, error)
	// PushDocGraph to this peer.
	PushDocGraph(context.Context, *PushDocGraphRequest) (*PushDocGraphReply, error)
	// GetLog from this peer.
	GetLog(context.Context, *GetLogRequest) (*GetLogReply, error)
	// PushLog to this peer.
	PushLog(context.Context, *PushLogRequest) (*PushLogReply, error)
	// GetHeadLog from this peer
	GetHeadLog(context.Context, *GetHeadLogRequest) (*GetHeadLogReply, error)
	mustEmbedUnimplementedServiceServer()
}

// UnimplementedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (UnimplementedServiceServer) GetDocGraph(context.Context, *GetDocGraphRequest) (*GetDocGraphReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDocGraph not implemented")
}
func (UnimplementedServiceServer) PushDocGraph(context.Context, *PushDocGraphRequest) (*PushDocGraphReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushDocGraph not implemented")
}
func (UnimplementedServiceServer) GetLog(context.Context, *GetLogRequest) (*GetLogReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLog not implemented")
}
func (UnimplementedServiceServer) PushLog(context.Context, *PushLogRequest) (*PushLogReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushLog not implemented")
}
func (UnimplementedServiceServer) GetHeadLog(context.Context, *GetHeadLogRequest) (*GetHeadLogReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHeadLog not implemented")
}
func (UnimplementedServiceServer) mustEmbedUnimplementedServiceServer() {}

// UnsafeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceServer will
// result in compilation errors.
type UnsafeServiceServer interface {
	mustEmbedUnimplementedServiceServer()
}

func RegisterServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&Service_ServiceDesc, srv)
}

func _Service_GetDocGraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDocGraphRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).GetDocGraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_GetDocGraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).GetDocGraph(ctx, req.(*GetDocGraphRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_PushDocGraph_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushDocGraphRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).PushDocGraph(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_PushDocGraph_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).PushDocGraph(ctx, req.(*PushDocGraphRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_GetLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).GetLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_GetLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).GetLog(ctx, req.(*GetLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_PushLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).PushLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_PushLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).PushLog(ctx, req.(*PushLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_GetHeadLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHeadLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).GetHeadLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_GetHeadLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).GetHeadLog(ctx, req.(*GetHeadLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Service_ServiceDesc is the grpc.ServiceDesc for Service service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Service_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "net.pb.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDocGraph",
			Handler:    _Service_GetDocGraph_Handler,
		},
		{
			MethodName: "PushDocGraph",
			Handler:    _Service_PushDocGraph_Handler,
		},
		{
			MethodName: "GetLog",
			Handler:    _Service_GetLog_Handler,
		},
		{
			MethodName: "PushLog",
			Handler:    _Service_PushLog_Handler,
		},
		{
			MethodName: "GetHeadLog",
			Handler:    _Service_GetHeadLog_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "net.proto",
}
