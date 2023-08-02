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
				A string `config:"A"`
				B string `config:"B"`
				C string `config:"C"`
				D string `config:"D"`
			}{},
			want: &struct {
				A string `config:"A"`
				B string `config:"B"`
				C string `config:"C"`
				D string `config:"D"`
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
				A string `config:"A"`
				B string `config:"B"`
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
