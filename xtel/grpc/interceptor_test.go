package grpc

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type TracerTestSuite struct {
	suite.Suite

	lis *bufconn.Listener
}

func TestTracerTestSuite(t *testing.T) {
	suite.Run(t, new(TracerTestSuite))
}

func (s *TracerTestSuite) setupServer(tr *Tracer, mockService func(*mock.Mock)) (
	context.CancelFunc, <-chan error,
) {
	s.lis = bufconn.Listen(bufSize)
	ctx, cancel := context.WithCancel(context.Background())
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(tr.UnaryServerInterceptor))

	mtss := newMockTestServiceServer(s.T())
	if mockService != nil {
		mockService(&mtss.Mock)
	}

	RegisterTestServiceServer(srv, mtss)

	errCh := make(chan error, 1)

	go func(ctx context.Context) {
		errCh <- func(ctx context.Context) error {
			errCh := make(chan error, 1)

			go func(errCh chan error) {
				if err := srv.Serve(s.lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					errCh <- err
				}
			}(errCh)

			select {
			case err := <-errCh:
				return err
			case <-ctx.Done():
			}

			srv.GracefulStop()

			return nil
		}(ctx)
	}(ctx)

	return func() {
		cancel()
		mtss.AssertExpectations(s.T())
	}, errCh
}

func (s *TracerTestSuite) TestDefaultTracer() {
	assert.Equal(s.T(), defaultTracer, DefaultTracer)
}

func (s *TracerTestSuite) TestNewTracer_UnaryClientInterceptor() {
	sr := tracetest.NewSpanRecorder()
	tr := NewTracer(
		WithTracerProvider(trace.NewTracerProvider(trace.WithSpanProcessor(sr))),
	)

	cancel, errCh := s.setupServer(tr, func(m *mock.Mock) {
		m.On("TestMethod").Return(&TestMethodResponse{}, nil).Once()
	})
	defer cancel()

	conn, err := grpc.Dial("bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return s.lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(tr.UnaryClientInterceptor),
	)
	s.NoError(err)

	defer func() {
		s.NoError(conn.Close())
	}()

	_, err = NewTestServiceClient(conn).TestMethod(context.Background(), &TestMethodRequest{Key: "test"})

	s.NoError(err)

	cancel()

	s.NoError(<-errCh)

	spans := sr.Ended()
	s.Len(spans, 2)

	for _, span := range spans {
		spanCtx := span.SpanContext()
		s.True(spanCtx.IsValid())
		s.True(spanCtx.HasSpanID())
		s.True(spanCtx.HasTraceID())

		s.ElementsMatch([]attribute.KeyValue{
			semconv.RPCSystemKey.String("grpc"),
			semconv.RPCServiceKey.String("TestService"),
			semconv.RPCMethodKey.String("TestMethod"),
			otelgrpc.GRPCStatusCodeKey.Int64(0),
		}, span.Attributes())
	}
}

type mockTestServiceServer struct {
	mock.Mock
	UnimplementedTestServiceServer
}

func newMockTestServiceServer(t *testing.T) *mockTestServiceServer {
	m := &mockTestServiceServer{}
	m.Test(t)
	return m
}

func (m *mockTestServiceServer) TestMethod(context.Context, *TestMethodRequest) (*TestMethodResponse, error) {
	args := m.Called()
	if tmr := args.Get(0); tmr != nil {
		return tmr.(*TestMethodResponse), args.Error(1)
	}

	return nil, args.Error(1)
}
