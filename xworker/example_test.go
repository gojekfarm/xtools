package xworker_test

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gojekfarm/xtools/xworker"
)

func ExampleNewAdapter() {
	a, err := xworker.NewAdapter(xworker.AdapterOptions{
		/* Fulfiller: gocraft.New(grocraft.Options{}) */
	})

	if err != nil {
		panic(err)
	}

	_ = a.Start()
}

func ExampleAdapter_Enqueue() {
	type jobPayload struct {
		Key1 string `json:"key_1"`
		Key2 string `json:"key_2"`
	}

	a, err := xworker.NewAdapter(xworker.AdapterOptions{})
	if err != nil {
		panic(err)
	}

	er, err := a.Enqueue(context.Background(), &xworker.Job{
		Name:    "job-1",
		Payload: &jobPayload{Key1: "value1", Key2: "value2"},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(er.String())
}

func ExampleAdapter_RegisterHandlerWithOptions() {
	type jobPayload struct {
		Key1 string `json:"key_1"`
		Key2 string `json:"key_2"`
	}

	a, err := xworker.NewAdapter(xworker.AdapterOptions{})
	if err != nil {
		panic(err)
	}

	h := xworker.HandlerFunc(func(ctx context.Context, j *xworker.Job) error {
		var jp jobPayload

		if err := j.DecodePayload(&jp); err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", jp)

		return nil
	})

	if err := a.RegisterHandlerWithOptions(
		/* jobName */ "job-1",
		h,
		xworker.RegisterOptions{
			MaxRetries:     1,
			MaxConcurrency: 3,
		},
	); err != nil {
		panic(err)
	}
}

func ExampleAdapter_UseEnqueueMiddleware() {
	logMiddleware := xworker.EnqueueMiddlewareFunc(func(next xworker.Enqueuer) xworker.Enqueuer {
		return xworker.EnqueuerFunc(
			func(ctx context.Context, j *xworker.Job, opt ...xworker.Option) (*xworker.EnqueueResult, error) {
				log.Printf("job enqueued: %+v\n", j)
				return next.Enqueue(ctx, j, opt...)
			},
		)
	})

	a, err := xworker.NewAdapter(xworker.AdapterOptions{})
	if err != nil {
		panic(err)
	}

	a.UseEnqueueMiddleware(logMiddleware)
}

func ExampleAdapter_UseJobMiddleware() {
	logMiddleware := xworker.JobMiddlewareFunc(func(next xworker.Handler) xworker.Handler {
		return xworker.HandlerFunc(
			func(ctx context.Context, j *xworker.Job) error {
				log.Printf("handling job: %+v\n", j)
				return next.Handle(ctx, j)
			},
		)
	})

	a, err := xworker.NewAdapter(xworker.AdapterOptions{})
	if err != nil {
		panic(err)
	}

	a.UseJobMiddleware(logMiddleware)
}

func ExampleJob_DecodePayload() {
	type jobPayload struct {
		Key1 string `json:"key_1"`
		Key2 string `json:"key_2"`
	}

	_ = xworker.HandlerFunc(func(ctx context.Context, j *xworker.Job) error {
		var jp jobPayload

		if err := j.DecodePayload(&jp); err != nil {
			panic(err)
		}

		return nil
	})
}

func ExampleAdapter_EnqueuePeriodically() {
	type dbShardPayload struct {
		TableName string `json:"table_name"`
	}

	a, err := xworker.NewAdapter(xworker.AdapterOptions{})
	if err != nil {
		panic(err)
	}

	if err := a.EnqueuePeriodically("* * * * *", &xworker.Job{
		Name:    "periodic-job",
		Payload: &dbShardPayload{TableName: "order_history"},
	}, xworker.Unique); err != nil {
		panic(err)
	}
}
