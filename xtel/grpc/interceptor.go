package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryClientInterceptor is a gRPC client-side interceptor that provides OpenTelemetry monitoring for Unary RPCs.
func UnaryClientInterceptor(ctx context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return defaultTracer.UnaryClientInterceptor(ctx, method, req, reply, cc, invoker, opts...)
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides OpenTelemetry monitoring for Unary RPCs.
func UnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return defaultTracer.UnaryServerInterceptor(ctx, req, info, handler)
}

// UnaryClientInterceptor is implementation for UnaryClientInterceptor function.
func (t *Tracer) UnaryClientInterceptor(
	ctx context.Context,
	method string,
	req,
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return t.uci(ctx, method, req, reply, cc, invoker, opts...)
}

// UnaryServerInterceptor is implementation for UnaryServerInterceptor function.
func (t *Tracer) UnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	return t.usi(ctx, req, info, handler)
}
