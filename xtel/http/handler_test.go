package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewHandler creates a new http.Handler, by sending buffered data to the client.
func TestNewHandler(t *testing.T) {
	h := NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Implements(t, (*http.Flusher)(nil), w)
		}), "test_handler",
	)

	r, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	require.NoError(t, err)

	h.ServeHTTP(httptest.NewRecorder(), r)

	// To check the body when it is nil
	nh := NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				_, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
			}
		}), "test_handler",
	)
	nr, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	require.NoError(t, err)

	nrr := httptest.NewRecorder()
	rb := nrr.Result()
	defer func() {
		assert.NoError(t, rb.Body.Close())
	}()

	nh.ServeHTTP(nrr, nr)
	assert.Equal(t, 200, rb.StatusCode)
}
