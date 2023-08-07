package xload

import (
	"context"
	"errors"
	"testing"
)

func TestSerialLoader(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "serial loader",
			input: &struct {
				A string `env:"A"`
				B string `env:"B"`
				C string `env:"C"`
				D string `env:"D"`
			}{},
			want: &struct {
				A string `env:"A"`
				B string `env:"B"`
				C string `env:"C"`
				D string `env:"D"`
			}{
				A: "loader-1: 1",
				B: "loader-2: 2",
				C: "loader-2: 3",
				D: "",
			},
			loader: SerialLoader(
				MapLoader{"A": "loader-1: 1", "B": "loader-1: 2"},
				MapLoader{"B": "loader-2: 2", "C": "loader-2: 3"},
			),
		},
		{
			name: "serial loader: error",
			input: &struct {
				A string `env:"A"`
				B string `env:"B"`
			}{},
			loader: SerialLoader(
				MapLoader{"A": "loader-1: 1", "B": "loader-1: 2"},
				LoaderFunc(func(ctx context.Context, key string) (string, error) {
					return "", errors.New("error loading field")
				}),
			),
			err: errors.New("error loading field"),
		},
	}

	runTestcases(t, testcases)
}
