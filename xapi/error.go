package xapi

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorHandler defines the interface for handling errors in HTTP responses.
type ErrorHandler interface {
	HandleError(w http.ResponseWriter, err error)
}

// ErrorFunc is a function type that implements ErrorHandler.
type ErrorFunc func(w http.ResponseWriter, err error)

// HandleError implements the ErrorHandler interface.
func (e ErrorFunc) HandleError(w http.ResponseWriter, err error) {
	e(w, err)
}

// DefaultErrorHandler provides default error handling for common JSON errors.
func DefaultErrorHandler(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, &json.SyntaxError{}):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, &json.UnmarshalTypeError{}):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, &json.InvalidUnmarshalError{}):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
