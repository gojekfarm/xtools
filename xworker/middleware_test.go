package xworker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdapter_UseJobMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		mwf     func(*mockCounter) []JobMiddlewareFunc
		wantLen int
	}{
		{
			name: "NoMiddlewares",
			mwf: func(m *mockCounter) []JobMiddlewareFunc {
				return []JobMiddlewareFunc{}
			},
		},
		{
			name: "SingleMiddleware",
			mwf: func(m *mockCounter) []JobMiddlewareFunc {
				m.On("Inc").Return().Once()
				return []JobMiddlewareFunc{
					func(next Handler) Handler {
						return HandlerFunc(func(ctx context.Context, j *Job) error {
							m.Inc()
							return next.Handle(ctx, j)
						})
					},
				}
			},
			wantLen: 1,
		},
		{
			name: "MultipleMiddlewares",
			mwf: func(m *mockCounter) []JobMiddlewareFunc {
				m.On("Inc").Return().Times(3)
				mwf := JobMiddlewareFunc(func(next Handler) Handler {
					return HandlerFunc(func(ctx context.Context, j *Job) error {
						m.Inc()
						return next.Handle(ctx, j)
					})
				})
				return []JobMiddlewareFunc{
					mwf, mwf, mwf,
				}
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{}
			mc := newMockCounter(t)

			for _, f := range tt.mwf(mc) {
				a.UseJobMiddleware(f)
			}

			h := a.wrapHandlerWithMiddlewares(HandlerFunc(func(ctx context.Context, j *Job) error {
				return nil
			}))

			assert.NoError(t, h.Handle(context.Background(), &Job{}))

			mc.AssertExpectations(t)
			assert.Len(t, a.handlerMiddlewares, tt.wantLen)
		})
	}
}

func TestAdapter_UseEnqueueMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		mwf     func(*mockCounter) []EnqueueMiddlewareFunc
		wantLen int
	}{

		{
			name: "NoMiddlewares",
			mwf: func(m *mockCounter) []EnqueueMiddlewareFunc {
				return []EnqueueMiddlewareFunc{}
			},
		},
		{
			name: "SingleMiddleware",
			mwf: func(m *mockCounter) []EnqueueMiddlewareFunc {
				m.On("Inc").Return().Once()
				return []EnqueueMiddlewareFunc{
					func(next Enqueuer) Enqueuer {
						return EnqueuerFunc(func(ctx context.Context, j *Job, opt ...Option) (*EnqueueResult, error) {
							m.Inc()
							return next.Enqueue(ctx, j, opt...)
						})
					},
				}
			},
			wantLen: 1,
		},
		{
			name: "MultipleMiddlewares",
			mwf: func(m *mockCounter) []EnqueueMiddlewareFunc {
				m.On("Inc").Return().Times(3)
				mwf := EnqueueMiddlewareFunc(func(next Enqueuer) Enqueuer {
					return EnqueuerFunc(func(ctx context.Context, j *Job, opt ...Option) (*EnqueueResult, error) {
						m.Inc()
						return next.Enqueue(ctx, j, opt...)
					})
				})
				return []EnqueueMiddlewareFunc{
					mwf, mwf, mwf,
				}
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := newMockFulfiller(t)
			mf.On("Enqueue", mock.Anything, mock.AnythingOfType("*xworker.Job"), []Option(nil)).
				Return(&EnqueueResult{}, nil)
			mf.On("Start").Return(nil)
			mf.On("Stop").Return(nil)

			a := &Adapter{
				fulfiller: mf,
			}
			a.wrappedEnqueuer = a.enqueuer()
			mc := newMockCounter(t)

			for _, f := range tt.mwf(mc) {
				a.UseEnqueueMiddleware(f)
			}

			_ = mf.Start()

			_, err := a.Enqueue(context.Background(), &Job{})
			assert.NoError(t, err)

			_ = mf.Stop()

			mc.AssertExpectations(t)
			mf.AssertExpectations(t)
			assert.Len(t, a.enqueueMiddlewares, tt.wantLen)
		})
	}
}

type mockCounter struct {
	mock.Mock
}

func (m *mockCounter) Inc() {
	m.Called()
}

func newMockCounter(t *testing.T) *mockCounter {
	m := &mockCounter{}
	m.Test(t)
	return m
}
