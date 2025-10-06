package xapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gojekfarm/xtools/xapi"
)

func ExampleEndpoint_basic() {
	type CreateUserRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type CreateUserResponse struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	createUser := xapi.EndpointFunc[CreateUserRequest, CreateUserResponse](
		func(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
			// Simulate user creation logic
			return &CreateUserResponse{
				ID:    1,
				Name:  req.Name,
				Email: req.Email,
			}, nil
		},
	)

	endpoint := xapi.NewEndpoint(createUser)

	http.Handle("/users", endpoint.Handler())
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExampleEndpoint_customErrorHandler() {
	type GetUserRequest struct {
		ID int `json:"id"`
	}

	type GetUserResponse struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	customErrorHandler := xapi.ErrorFunc(func(w http.ResponseWriter, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		errorResponse := map[string]string{
			"error": err.Error(),
		}

		json.NewEncoder(w).Encode(errorResponse)
	})

	getUser := xapi.EndpointFunc[GetUserRequest, GetUserResponse](
		func(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
			if req.ID <= 0 {
				return nil, fmt.Errorf("invalid user ID: %d", req.ID)
			}

			// Simulate user lookup
			return &GetUserResponse{
				ID:    req.ID,
				Name:  "John Doe",
				Email: "john@example.com",
			}, nil
		},
	)

	// Create endpoint with custom error handler
	endpoint := xapi.NewEndpoint(
		getUser,
		xapi.WithErrorHandler(customErrorHandler),
	)

	http.Handle("/users/", endpoint.Handler())
}

func ExampleEndpoint_withMiddleware() {
	type GetDataRequest struct {
		ID string `json:"id"`
	}

	type GetDataResponse struct {
		ID   string `json:"id"`
		Data string `json:"data"`
	}

	// Authentication middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Simulate token validation
			if token != "Bearer valid-token" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Rate limiting middleware
	rateLimitMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate rate limiting logic
			w.Header().Set("X-RateLimit-Limit", "100")
			w.Header().Set("X-RateLimit-Remaining", "99")

			next.ServeHTTP(w, r)
		})
	}

	getData := xapi.EndpointFunc[GetDataRequest, GetDataResponse](
		func(ctx context.Context, req *GetDataRequest) (*GetDataResponse, error) {
			requestID := ctx.Value("requestID")
			log.Printf("Processing request %s for ID: %s", requestID, req.ID)

			return &GetDataResponse{
				ID:   req.ID,
				Data: fmt.Sprintf("Data for %s", req.ID),
			}, nil
		},
	)

	endpoint := xapi.NewEndpoint(
		getData,
		xapi.WithMiddleware(
			xapi.MiddlewareFunc(rateLimitMiddleware),
			xapi.MiddlewareFunc(authMiddleware),
		),
	)

	http.Handle("/data", endpoint.Handler())
}
