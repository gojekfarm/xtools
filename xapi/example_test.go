package xapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gojekfarm/xtools/xapi"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Language string `json:"-"`
}

func (user *CreateUserRequest) Validate() error {
	if user.Name == "" {
		return fmt.Errorf("name is required")
	}
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}
	return nil
}

func (user *CreateUserRequest) Extract(r *http.Request) error {
	user.Language = r.Header.Get("Language")
	return nil
}

type CreateUserResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Language string `json:"language"`
}

func (user *CreateUserResponse) StatusCode() int {
	return http.StatusCreated
}

func ExampleEndpoint_basic() {
	createUser := xapi.EndpointFunc[CreateUserRequest, CreateUserResponse](
		func(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
			// Simulate user creation logic
			return &CreateUserResponse{
				ID:       1,
				Name:     req.Name,
				Email:    req.Email,
				Language: req.Language,
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

type GetArticleRequest struct {
	ID string `json:"-"`
}

func (article *GetArticleRequest) Extract(r *http.Request) error {
	article.ID = r.PathValue("id")

	return nil
}

type GetArticleResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (article *GetArticleResponse) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/html")
	w.WriteHeader(http.StatusOK)

	_, _ = fmt.Fprintf(w, "<html><body><h1>%s</h1></body></html>", article.Title)

	return nil
}

func ExampleEndpoint_withCustomResponseWriter() {
	getArticle := xapi.EndpointFunc[GetArticleRequest, GetArticleResponse](
		func(ctx context.Context, req *GetArticleRequest) (*GetArticleResponse, error) {
			return &GetArticleResponse{
				ID:    req.ID,
				Title: "Article " + req.ID,
			}, nil
		},
	)

	endpoint := xapi.NewEndpoint(getArticle)

	http.Handle("/articles/{id}", endpoint.Handler())
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
