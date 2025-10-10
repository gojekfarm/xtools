package xapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpoint_Handler(t *testing.T) {
	t.Parallel()

	t.Run("BasicEndpoint", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return &BasicResponse{
					Message: "Hello " + req.Name,
					ID:      123,
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "World"})))
		req.Header.Set("Content-Type", "application/json")

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"message": "Hello World",
			"id": 123
		}`, rec.Body.String())
	})

	t.Run("WithValidation", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[ValidatedRequest, BasicResponse](
			func(ctx context.Context, req *ValidatedRequest) (*BasicResponse, error) {
				return &BasicResponse{
					Message: "Valid request",
					ID:      456,
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &ValidatedRequest{Name: ""}))) // Empty name should fail validation

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "name is required")
	})

	t.Run("WithExtraction", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[ExtractedRequest, BasicResponse](
			func(ctx context.Context, req *ExtractedRequest) (*BasicResponse, error) {
				return &BasicResponse{
					Message: fmt.Sprintf("Hello %s from %s", req.Name, req.Language),
					ID:      789,
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &ExtractedRequest{Name: "World"})))
		req.Header.Set("Language", "en-US")

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"message": "Hello World from en-US",
			"id": 789
		}`, rec.Body.String())
	})

	t.Run("WithCustomStatusCode", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, StatusResponse](
			func(ctx context.Context, req *BasicRequest) (*StatusResponse, error) {
				return &StatusResponse{
					Message: "Created successfully",
					ID:      999,
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Result().StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"message": "Created successfully",
			"id": 999
		}`, rec.Body.String())
	})

	t.Run("WithRawWriter", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, RawResponse](
			func(ctx context.Context, req *BasicRequest) (*RawResponse, error) {
				return &RawResponse{
					Content: fmt.Sprintf("<h1>Hello %s</h1>", req.Name),
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "World"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
		assert.Equal(t, "<h1>Hello World</h1>", rec.Body.String())
	})

	t.Run("HandlerError", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return nil, errors.New("handler error")
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "handler error")
	})

	t.Run("JSONDecodeError", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return &BasicResponse{Message: "Success"}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(
			http.MethodPost, "/test",
			strings.NewReader("invalid json"),
		)

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "invalid character")
	})

	t.Run("JSONEncodeError", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, InvalidJSONResponse](
			func(ctx context.Context, req *BasicRequest) (*InvalidJSONResponse, error) {
				return &InvalidJSONResponse{
					Channel: make(chan int), // Channels can't be JSON encoded
				}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		// With the improved implementation, JSON encoding errors are caught before headers are written
		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "json: unsupported type")
	})

	t.Run("ExtractionError", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[ExtractionErrorRequest, BasicResponse](
			func(ctx context.Context, req *ExtractionErrorRequest) (*BasicResponse, error) {
				return &BasicResponse{Message: "Success"}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &ExtractionErrorRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "extraction error")
	})

	t.Run("RawWriterError", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, RawWriterErrorResponse](
			func(ctx context.Context, req *BasicRequest) (*RawWriterErrorResponse, error) {
				return &RawWriterErrorResponse{}, nil
			},
		)

		endpoint := NewEndpoint(handler)
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "raw writer error")
	})
}

func TestEndpoint_WithMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("SingleMiddleware", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return &BasicResponse{Message: "Success"}, nil
			},
		)

		middleware := MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Middleware", "applied")
				next.ServeHTTP(w, r)
			})
		})

		endpoint := NewEndpoint(handler, WithMiddleware(middleware))
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Equal(t, "applied", rec.Header().Get("X-Middleware"))
	})

	t.Run("MultipleMiddleware", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return &BasicResponse{Message: "Success"}, nil
			},
		)

		middleware1 := MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Middleware-1", "first")
				next.ServeHTTP(w, r)
			})
		})

		middleware2 := MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Middleware-2", "second")
				next.ServeHTTP(w, r)
			})
		})

		endpoint := NewEndpoint(handler, WithMiddleware(middleware1, middleware2))
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.Equal(t, "first", rec.Header().Get("X-Middleware-1"))
		assert.Equal(t, "second", rec.Header().Get("X-Middleware-2"))
	})

	t.Run("MiddlewareOrder", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				return &BasicResponse{Message: "Success"}, nil
			},
		)

		order := []string{}

		// Middleware should be applied in reverse order (last added first)
		middleware1 := MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
				order = append(order, "1")
			})
		})

		middleware2 := MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
				order = append(order, "2")
			})
		})

		endpoint := NewEndpoint(handler, WithMiddleware(middleware1, middleware2))
		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		assert.EqualValues(t, []string{"2", "1"}, order)
	})
}

func TestEndpoint_WithCustomErrorHandler(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	handler := EndpointFunc[BasicRequest, BasicResponse](
		func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
			return nil, errors.New("custom error")
		},
	)

	customErrorHandler := ErrorFunc(func(w http.ResponseWriter, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Custom: " + err.Error(),
		})
	})

	endpoint := NewEndpoint(handler, WithErrorHandler(customErrorHandler))
	req := httptest.NewRequest(http.MethodPost, "/test",
		bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))

	endpoint.Handler().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Custom: custom error", response["error"])
}

func TestEndpoint_ContextCancellation(t *testing.T) {
	t.Parallel()

	t.Run("ContextCancelled", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()

		handler := EndpointFunc[BasicRequest, BasicResponse](
			func(ctx context.Context, req *BasicRequest) (*BasicResponse, error) {
				// Simulate context cancellation
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return &BasicResponse{Message: "Success"}, nil
				}
			},
		)

		endpoint := NewEndpoint(handler)

		// Create a request with a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := httptest.NewRequest(http.MethodPost, "/test",
			bytes.NewBuffer(mustMarshalJSON(t, &BasicRequest{Name: "Test"})))
		req = req.WithContext(ctx)

		endpoint.Handler().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "context canceled")
	})
}

// Test types and implementations

type BasicRequest struct {
	Name string `json:"name"`
}

type BasicResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
}

type ValidatedRequest struct {
	Name string `json:"name"`
}

func (r *ValidatedRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

type ExtractedRequest struct {
	Name     string `json:"name"`
	Language string `json:"-"`
}

func (r *ExtractedRequest) Extract(req *http.Request) error {
	r.Language = req.Header.Get("Language")
	return nil
}

type StatusResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
}

func (r *StatusResponse) StatusCode() int {
	return http.StatusCreated
}

type RawResponse struct {
	Content string `json:"-"`
}

func (r *RawResponse) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(r.Content))
	return err
}

type InvalidJSONResponse struct {
	Channel chan int `json:"channel"`
}

func (r *InvalidJSONResponse) MarshalJSON() ([]byte, error) {
	// Force an error during JSON marshaling
	return nil, errors.New("json: unsupported type")
}

type ExtractionErrorRequest struct {
	Name string `json:"name"`
}

func (r *ExtractionErrorRequest) Extract(req *http.Request) error {
	return errors.New("extraction error")
}

type RawWriterErrorResponse struct{}

func (r *RawWriterErrorResponse) Write(w http.ResponseWriter) error {
	return errors.New("raw writer error")
}

// Helper functions

func mustMarshalJSON(t *testing.T, v any) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}
