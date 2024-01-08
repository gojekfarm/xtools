package xload

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/gotidy/ptr"
)

type House struct {
	Name    string `env:"NAME"`
	Address Address
	Living  Room  `env:",prefix=LIVING_"`
	Bedroom *Room `env:",prefix=BEDROOM_"`
	Plot    Plot  `env:"PLOT"`
}

type Villa struct {
	House

	Floors int `env:"FLOORS"`
}

type Address struct {
	Street    string   `env:"STREET"`
	City      string   `env:"CITY"`
	Longitute *float64 `env:"LONGITUTE"`
	Latitude  *float64 `env:"LATITUDE"`
}

type Room struct {
	Name   string `env:"NAME" json:"name,omitempty"`
	Width  int    `env:"WIDTH" json:"width,omitempty"`
	Length int    `env:"LENGTH" json:"length,omitempty"`
}

type Plot struct {
	Width  int    `json:"width,omitempty"`
	Length int    `json:"length,omitempty"`
	Number string `json:"number,omitempty"`
}

func (p *Plot) UnmarshalJSON(b []byte) error {
	type Alias Plot

	var a Alias
	err := json.Unmarshal(b, &a)
	if err != nil {
		return err
	}

	*p = Plot(a)

	return nil
}

type Plots []Plot

func (p *Plots) Decode(s string) error {
	plots := []Plot{}
	err := json.Unmarshal([]byte(s), &plots)
	if err != nil {
		return err
	}

	*p = plots

	return nil
}

func TestLoad_Structs(t *testing.T) {
	t.Parallel()

	strKind := reflect.TypeOf("").Kind()

	testcases := []testcase{
		{
			name: "nested struct: using prefix",
			input: &struct {
				Name       string `env:"NAME"`
				Living     Room   `env:",prefix=LIVING_"`
				FirstLevel struct {
					Bedroom Room `env:",prefix=BEDROOM_"`
				} `env:",prefix=FIRST_LEVEL_"`
			}{},
			want: &struct {
				Name       string `env:"NAME"`
				Living     Room   `env:",prefix=LIVING_"`
				FirstLevel struct {
					Bedroom Room `env:",prefix=BEDROOM_"`
				} `env:",prefix=FIRST_LEVEL_"`
			}{
				Name: "app1",
				Living: Room{
					Name:   "living",
					Width:  1,
					Length: 2,
				},
				FirstLevel: struct {
					Bedroom Room `env:",prefix=BEDROOM_"`
				}{
					Bedroom: Room{
						Name:   "bedroom",
						Width:  3,
						Length: 4,
					},
				},
			},
			loader: MapLoader{
				"NAME":                       "app1",
				"LIVING_NAME":                "living",
				"LIVING_WIDTH":               "1",
				"LIVING_LENGTH":              "2",
				"FIRST_LEVEL_BEDROOM_NAME":   "bedroom",
				"FIRST_LEVEL_BEDROOM_WIDTH":  "3",
				"FIRST_LEVEL_BEDROOM_LENGTH": "4",
			},
		},
		{
			name: "nested struct: without prefix",
			input: &struct {
				Name    string `env:"NAME"`
				Address Address
			}{},
			want: &struct {
				Name    string `env:"NAME"`
				Address Address
			}{
				Name: "house1",
				Address: Address{
					Street:    "street1",
					City:      "city1",
					Longitute: ptr.Float64(1.1),
					Latitude:  ptr.Float64(-2.2),
				},
			},
			loader: MapLoader{
				"NAME":      "house1",
				"STREET":    "street1",
				"CITY":      "city1",
				"LONGITUTE": "1.1",
				"LATITUDE":  "-2.2",
			},
		},
		{
			name:  "inheritance: using prefix",
			input: &Villa{},
			want: &Villa{
				House: House{
					Name: "villa1",
					Address: Address{
						Street:    "street1",
						City:      "city1",
						Longitute: ptr.Float64(1.1),
						Latitude:  ptr.Float64(-2.2),
					},
				},
				Floors: 2,
			},
			loader: MapLoader{
				"NAME":      "villa1",
				"STREET":    "street1",
				"CITY":      "city1",
				"LONGITUTE": "1.1",
				"LATITUDE":  "-2.2",
				"FLOORS":    "2",
			},
		},
		{
			name: "non-struct key with prefix",
			input: &struct {
				Name string `env:",prefix=CLUSTER"`
			}{},
			err:    &ErrInvalidPrefix{field: "Name", kind: strKind},
			loader: MapLoader{},
		},
		{
			name: "struct with key and prefix",
			input: &struct {
				Address Address `env:"ADDRESS,prefix=CLUSTER"`
			}{},
			err:    &ErrInvalidPrefixAndKey{field: "Address", key: "ADDRESS"},
			loader: MapLoader{},
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_Decoder(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		// time values
		{
			name: "time",
			input: &struct {
				Time    time.Time  `env:"TIME"`
				OptTime *time.Time `env:"OPT_TIME"`
			}{},
			want: &struct {
				Time    time.Time  `env:"TIME"`
				OptTime *time.Time `env:"OPT_TIME"`
			}{
				Time:    time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
				OptTime: ptr.Time(time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)),
			},
			loader: MapLoader{
				"TIME":     "2020-01-02T03:04:05.000000006Z",
				"OPT_TIME": "2020-01-02T03:04:05.000000006Z",
			},
		},
		{
			name: "time: invalid",
			input: &struct {
				Time time.Time `env:"TIME"`
			}{},
			loader: MapLoader{"TIME": "invalid"},
			err:    errors.New("cannot parse"),
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_JSON(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "json object as string",
			input: &struct {
				Plot Plot `env:"PLOT"`
			}{},
			want: &struct {
				Plot Plot `env:"PLOT"`
			}{
				Plot: Plot{
					Width:  100,
					Length: 200,
					Number: "plot1",
				},
			},
			loader: MapLoader{
				"PLOT": `{"width":100,"length":200,"number":"plot1"}`,
			},
		},
		{
			name: "override json default value",
			input: &struct {
				Plot Plot `env:"PLOT"`
			}{
				Plot: Plot{
					Width:  10,
					Length: 20,
					Number: "default",
				},
			},
			want: &struct {
				Plot Plot `env:"PLOT"`
			}{
				Plot: Plot{
					Width:  100,
					Length: 200,
					Number: "plot1",
				},
			},
			loader: MapLoader{
				"PLOT": `{"width":100,"length":200,"number":"plot1"}`,
			},
		},
		{
			name: "json array as string",
			input: &struct {
				Plots Plots `env:"PLOTS"`
			}{},
			want: &struct {
				Plots Plots `env:"PLOTS"`
			}{
				Plots: Plots{
					{
						Width:  100,
						Length: 200,
						Number: "plot1",
					},
					{
						Width:  300,
						Length: 400,
						Number: "plot2",
					},
				},
			},
			loader: MapLoader{
				"PLOTS": `[{"width":100,"length":200,"number":"plot1"},{"width":300,"length":400,"number":"plot2"}]`,
			},
		},
		{
			name: "json: invalid",
			input: &struct {
				Plot Plot `env:"PLOT"`
			}{},
			err:    errors.New("invalid character"),
			loader: MapLoader{"PLOT": `invalid`},
		},
		{
			name: "json: loader error",
			input: &struct {
				Plot Plot `env:"PLOT"`
			}{},
			err: errors.New("loader error"),
			loader: LoaderFunc(func(ctx context.Context, key string) (string, error) {
				return "", errors.New("loader error")
			}),
		},
		{
			name: "json: empty required value",
			input: &struct {
				Plot Plot `env:"PLOT,required"`
			}{},
			err:    &ErrRequired{key: "PLOT"},
			loader: MapLoader{},
		},
	}

	runTestcases(t, testcases)
}
