package http

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func TestChildSpanFromGlobalTracer(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)
	otel.SetTracerProvider(provider)

	router := mux.NewRouter()
	router.Use(MuxMiddleware("foobar"))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		span := oteltrace.SpanFromContext(r.Context())
		mockTracer, ok := span.TracerProvider().(oteltrace.Tracer)
		require.True(t, ok)
		assert.Equal(t, "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux", mockTracer)
		w.WriteHeader(http.StatusOK)
	})
}

func TestChildSpanFromCustomTracer(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)

	router := mux.NewRouter()
	router.Use(MuxMiddleware("foobar", WithTracerProvider(provider)))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		span := oteltrace.SpanFromContext(r.Context())
		assert.NotNil(t, span)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
}

func TestChildSpanNames(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)

	router := mux.NewRouter()
	router.Use(MuxMiddleware("foobar", WithTracerProvider(provider)))
	router.HandleFunc("/user/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.HandleFunc("/book/{title}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(([]byte)("ok"))
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	spans := sr.Ended()
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, "/user/{id:[0-9]+}", span.Name())
	assert.Equal(t, oteltrace.SpanKindServer, span.SpanKind())

	r = httptest.NewRequest("GET", "/book/foo", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	spans = sr.Ended()
	require.Len(t, spans, 2)
	span = spans[1]
	assert.Equal(t, "/book/{title}", span.Name())
	assert.Equal(t, oteltrace.SpanKindServer, span.SpanKind())
}

func TestGetSpanNotInstrumented(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		span := oteltrace.SpanFromContext(r.Context())
		ok := !span.SpanContext().IsValid()
		assert.True(t, ok)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
}

func TestPropagationWithGlobalPropagators(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)

	r := httptest.NewRequest("GET", "/user/123", nil)

	ctx, pspan := provider.Tracer(tracerName).Start(context.Background(), "test")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	router := mux.NewRouter()
	router.Use(MuxMiddleware("foobar", WithTracerProvider(provider)))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		span := oteltrace.SpanFromContext(r.Context())
		assert.Equal(t, pspan.SpanContext().TraceID, span.SpanContext().TraceID)
		assert.Equal(t, pspan.SpanContext().SpanID, span.SpanContext().SpanID())
		w.WriteHeader(http.StatusOK)
	})

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
}

type testResponseWriter struct {
	writer http.ResponseWriter
}

func (rw *testResponseWriter) Header() http.Header {
	return rw.writer.Header()
}
func (rw *testResponseWriter) Write(b []byte) (int, error) {
	return rw.writer.Write(b)
}
func (rw *testResponseWriter) WriteHeader(statusCode int) {
	rw.writer.WriteHeader(statusCode)
}

func (rw *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (rw *testResponseWriter) Push(target string, opts *http.PushOptions) error {
	return nil
}

func (rw *testResponseWriter) Flush() {
}

func (rw *testResponseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func TestResponseWriterInterfaces(t *testing.T) {
	// make sure the recordingResponseWriter preserves interfaces implemented by the wrapped writer
	provider := oteltrace.NewNoopTracerProvider()
	tmp := otel.GetTextMapPropagator()
	wlf := defaultPathWhitelistFunc

	router := mux.NewRouter()
	router.Use(MuxMiddleware("foobar", WithTracerProvider(provider), WithTextMapPropagator(tmp), wlf))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		assert.Implements(t, (*http.Hijacker)(nil), w)
		assert.Implements(t, (*http.Pusher)(nil), w)
		assert.Implements(t, (*http.Flusher)(nil), w)
		assert.Implements(t, (*io.ReaderFrom)(nil), w)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := &testResponseWriter{
		writer: httptest.NewRecorder(),
	}

	router.ServeHTTP(w, r)
}

func TestWithPathWhitelistFunc(t *testing.T) {
	test := MuxMiddleware("service-a", ignoreRoutes([]string{
		"/login",
	}))

	assert.NotNil(t, test)
}

func ignoreRoutes(in []string) PathWhitelistFunc {
	return func(r *http.Request) bool {
		spanName := ""

		route := mux.CurrentRoute(r)
		if route != nil {
			var err error

			spanName, err = route.GetPathTemplate()
			if err != nil {
				spanName, err = route.GetPathRegexp()
				if err != nil {
					spanName = ""
				}
			}
		}

		for _, s := range in {
			if strings.EqualFold(spanName, s) {
				return true
			}
		}

		return false
	}
}

func TestMuxMiddleware_IgnorePaths(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)
	otel.SetTracerProvider(provider)

	h := MuxMiddleware("test-service", PathWhitelistFunc(func(*http.Request) bool { return false })).
		Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := oteltrace.SpanFromContext(r.Context())
			assert.False(t, span.IsRecording())
			w.WriteHeader(http.StatusOK)
		}))

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
}
