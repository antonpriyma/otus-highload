// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: api/dialog/grpc/v1/dialog.proto

package dialogs

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

// DialogsClient is the client API for Dialogs service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DialogsClient interface {
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
	GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error)
}

type dialogsClient struct {
	cc grpc.ClientConnInterface
}

func NewDialogsClient(cc grpc.ClientConnInterface) DialogsClient {
	return &dialogsClient{cc}
}

func (c *dialogsClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, "/Dialogs/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dialogsClient) GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error) {
	out := new(GetMessagesResponse)
	err := c.cc.Invoke(ctx, "/Dialogs/GetMessages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DialogsServer is the server API for Dialogs service.
// All implementations must embed UnimplementedDialogsServer
// for forward compatibility
type DialogsServer interface {
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error)
	mustEmbedUnimplementedDialogsServer()
}

// UnimplementedDialogsServer must be embedded to have forward compatible implementations.
type UnimplementedDialogsServer struct {
}

func (UnimplementedDialogsServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedDialogsServer) GetMessages(context.Context, *GetMessagesRequest) (*GetMessagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessages not implemented")
}
func (UnimplementedDialogsServer) mustEmbedUnimplementedDialogsServer() {}

// UnsafeDialogsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DialogsServer will
// result in compilation errors.
type UnsafeDialogsServer interface {
	mustEmbedUnimplementedDialogsServer()
}

func RegisterDialogsServer(s grpc.ServiceRegistrar, srv DialogsServer) {
	s.RegisterService(&Dialogs_ServiceDesc, srv)
}

func _Dialogs_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DialogsServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Dialogs/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DialogsServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Dialogs_GetMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMessagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DialogsServer).GetMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Dialogs/GetMessages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DialogsServer).GetMessages(ctx, req.(*GetMessagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Dialogs_ServiceDesc is the grpc.ServiceDesc for Dialogs service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Dialogs_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Dialogs",
	HandlerType: (*DialogsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _Dialogs_SendMessage_Handler,
		},
		{
			MethodName: "GetMessages",
			Handler:    _Dialogs_GetMessages_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/dialog/grpc/v1/dialog.proto",
}
