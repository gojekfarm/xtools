package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(lis); err != nil && err.Error() != "closed" {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

type grpcOpts struct {
	unaryInterceptors []grpc.UnaryServerInterceptor
}

func WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) func(o *grpcOpts) {
	return func(o *grpcOpts) {
		o.unaryInterceptors = interceptors
	}
}
func getSpanFromRecorder(sr *tracetest.SpanRecorder, name string) (trace.ReadOnlySpan, bool) {
	for _, s := range sr.Ended() {
		if s.Name() == name {
			return s, true
		}
	}
	return nil, false
}

type mockUICInvoker struct {
	ctx context.Context
}

func (mcuici *mockUICInvoker) invoker(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
	mcuici.ctx = ctx

	// if method contains error name, mock error return
	if strings.Contains(method, "error") {
		return status.Error(grpccodes.Internal, "internal error")
	}

	return nil
}

type mockProtoMessage struct{}

func (mm *mockProtoMessage) Reset() {
}

func (mm *mockProtoMessage) String() string {
	return "mock"
}

func (mm *mockProtoMessage) ProtoMessage() {
}

func TestUnaryClientInterceptor(t *testing.T) {
	conn, err := grpc.Dial("fake:connection", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to create client connection: %v", err)
	}
	defer func() {
		assert.NoError(t, conn.Close())
	}()

	vrr := UnaryClientInterceptor(
		context.Background(),
		"/package.v1/SomeMethod",
		nil,
		nil,
		conn,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)

	assert.NoError(t, vrr)

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	mwf := NewTracer(WithTracerProvider(tp))

	req := &mockProtoMessage{}
	reply := &mockProtoMessage{}
	uniInterceptorInvoker := &mockUICInvoker{}

	testcases := []struct {
		name             string
		method           string
		expectedSpanCode codes.Code
		expectedAttr     []attribute.KeyValue
		eventsAttr       []map[attribute.Key]attribute.Value
		wantErr          assert.ErrorAssertionFunc
	}{
		{
			name:   "serviceName/bar",
			method: "/serviceName/bar",
			expectedAttr: []attribute.KeyValue{
				semconv.RPCSystemKey.String("grpc"),
				semconv.RPCServiceKey.String("serviceName"),
				semconv.RPCMethodKey.String("bar"),
				otelgrpc.GRPCStatusCodeKey.Int64(0),
			},
			eventsAttr: []map[attribute.Key]attribute.Value{
				{
					otelgrpc.RPCMessageTypeKey:             attribute.StringValue("SENT"),
					otelgrpc.RPCMessageIDKey:               attribute.IntValue(1),
					otelgrpc.RPCMessageUncompressedSizeKey: attribute.IntValue(proto.Size(proto.Message(req))),
				},
				{
					otelgrpc.RPCMessageTypeKey:             attribute.StringValue("RECEIVED"),
					otelgrpc.RPCMessageIDKey:               attribute.IntValue(1),
					otelgrpc.RPCMessageUncompressedSizeKey: attribute.IntValue(proto.Size(proto.Message(reply))),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:   "invalidName",
			method: "invalidName",
			expectedAttr: []attribute.KeyValue{
				semconv.RPCSystemKey.String("grpc"),
				otelgrpc.GRPCStatusCodeKey.Int64(0),
			},
			eventsAttr: []map[attribute.Key]attribute.Value{
				{
					otelgrpc.RPCMessageTypeKey:             attribute.StringValue("SENT"),
					otelgrpc.RPCMessageIDKey:               attribute.IntValue(1),
					otelgrpc.RPCMessageUncompressedSizeKey: attribute.IntValue(proto.Size(proto.Message(req))),
				},
				{
					otelgrpc.RPCMessageTypeKey:             attribute.StringValue("RECEIVED"),
					otelgrpc.RPCMessageIDKey:               attribute.IntValue(1),
					otelgrpc.RPCMessageUncompressedSizeKey: attribute.IntValue(proto.Size(proto.Message(reply))),
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := mwf.UnaryClientInterceptor(context.Background(), tc.method, req, reply, conn, uniInterceptorInvoker.invoker)
			tc.wantErr(t, err)
			span, ok := getSpanFromRecorder(sr, tc.name)
			assert.True(t, ok, "missing span %q", tc.name)
			assert.Equal(t, tc.expectedSpanCode, span.Status().Code)
			assert.ElementsMatch(t, tc.expectedAttr, span.Attributes())
		})
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	mwf := NewTracer(WithTracerProvider(tp))
	uErr := status.Error(grpccodes.Internal, "INTERNAL_ERROR")

	// nolint:unparam
	p := func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, uErr
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(mwf.UnaryServerInterceptor))
	l := bufconn.Listen(bufSize)
	defer func() {
		assert.NoError(t, l.Close())
	}()

	go func() {
		if err := s.Serve(l); err != nil && err.Error() != "closed" {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "package.v1",
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "SomeMethod",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					return interceptor(ctx, nil, &grpc.UnaryServerInfo{
						FullMethod: "/package.v1/SomeMethod",
					}, p)
				},
			},
		},
	}, p)

	cc, err := grpc.Dial("bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, cc.Close())
	}()

	assert.Error(t, cc.Invoke(context.Background(), "/package.v1/SomeMethod", new(emptypb.Empty), new(emptypb.Empty)))

	span, _ := getSpanFromRecorder(sr, "package.v1/SomeMethod")
	assert.NotNil(t, span, "missing span %q", "package.v1/SomeMethod")

	val := span.SpanContext().IsValid()
	id := span.SpanContext().HasSpanID()
	assert.True(t, val)
	assert.True(t, id)

	assert.Equal(t, codes.Error, span.Status().Code)
	assert.Contains(t, uErr.Error(), span.Status().Description)

	assert.Len(t, span.Events(), 2)
	assert.ElementsMatch(t, []attribute.KeyValue{
		attribute.Key("message.type").String("SENT"),
		attribute.Key("message.id").Int(1),
	}, span.Events()[1].Attributes)
}

func TestWithUnaryInterceptor(t *testing.T) {
	wantInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, uErr error) {
		return nil, nil
	}

	o := grpcOpts{}
	WithUnaryInterceptor(wantInterceptor)(&o)
	require.Len(t, o.unaryInterceptors, 1)
	assert.Equal(t, fmt.Sprintf("%p", wantInterceptor), fmt.Sprintf("%p", o.unaryInterceptors[0]))
}
