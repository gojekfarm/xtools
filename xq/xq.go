// Package xq provides abstractions for custom queue implementations.
package xq

import (
	"context"
	"time"
)

// State is the state of a job in the queue.
type State int

// Possible states of a job.
const (
	StateAvailable State = iota
	StateInProgress
	StateScheduled
	StateDead
	StateDone
)

// Job defines the structure of a job in the queue.
type Job[T any] struct {
	ID          string
	Msg         *T
	State       State
	Err         error
	Attempts    int
	StartedAt   *time.Time
	ScheduledAt *time.Time
}

// Handler is an interface for handling messages.
type Handler[T any] interface {
	Handle(ctx context.Context, job *Job[T]) error
}

// HandlerFunc is a function that implements the Handler interface.
type HandlerFunc[T any] func(ctx context.Context, job *Job[T]) error

// Handle calls the HandlerFunc.
func (f HandlerFunc[T]) Handle(ctx context.Context, job *Job[T]) error {
	return f(ctx, job)
}
