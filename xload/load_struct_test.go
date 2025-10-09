package xload

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"

	xloadtype "github.com/gojekfarm/xtools/xload/type"
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
			wantErr: errContains(&ErrInvalidPrefix{field: "Name", kind: strKind}),
			loader:  MapLoader{},
		},
		{
			name: "struct with key and prefix",
			input: &struct {
				Address Address `env:"ADDRESS,prefix=CLUSTER"`
			}{},
			wantErr: errContains(&ErrInvalidPrefixAndKey{field: "Address", key: "ADDRESS"}),
			loader:  MapLoader{},
		},

		// key collision
		{
			name: "key collision",
			input: &struct {
				Address1 Address  `env:",prefix=ADDRESS_"`
				Address2 *Address `env:",prefix=ADDRESS_"`
			}{},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				tErr := &ErrCollision{}
				assert.ErrorAs(t, err, &tErr)
				return assert.ElementsMatch(t, tErr.Keys(), []string{
					"ADDRESS_CITY",
					"ADDRESS_LATITUDE",
					"ADDRESS_LONGITUTE",
					"ADDRESS_STREET",
				})
			},
			loader: MapLoader{
				"ADDRESS_STREET":    "street1",
				"ADDRESS_CITY":      "city1",
				"ADDRESS_LONGITUTE": "1.1",
				"ADDRESS_LATITUDE":  "-2.2",
			},
		},
		{
			name: "key collision with detection disabled",
			opts: []Option{SkipCollisionDetection},
			input: &struct {
				Address1 Address  `env:",prefix=ADDRESS_"`
				Address2 *Address `env:",prefix=ADDRESS_"`
			}{},
			want: &struct {
				Address1 Address
				Address2 *Address
			}{
				Address{
					Street:    "street1",
					City:      "city1",
					Longitute: ptr.Float64(1.1),
					Latitude:  ptr.Float64(-2.2),
				},
				&Address{
					Street:    "street1",
					City:      "city1",
					Longitute: ptr.Float64(1.1),
					Latitude:  ptr.Float64(-2.2),
				},
			},
			loader: MapLoader{
				"ADDRESS_STREET":    "street1",
				"ADDRESS_CITY":      "city1",
				"ADDRESS_LONGITUTE": "1.1",
				"ADDRESS_LATITUDE":  "-2.2",
			},
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
			loader:  MapLoader{"TIME": "invalid"},
			wantErr: errContains(errors.New("cannot parse")),
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
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				errDec := new(ErrDecode)
				return assert.ErrorAs(t, err, &errDec, i...) && assert.Equal(t, "invalid", errDec.Value())
			},
			loader: MapLoader{"PLOT": `invalid`},
		},
		{
			name: "json: array(not-struct) invalid",
			input: &struct {
				Plots Plots `env:"PLOTS"`
			}{},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				errDec := new(ErrDecode)
				return assert.ErrorAs(t, err, &errDec, i...) && assert.Equal(t, "invalid", errDec.Value())
			},
			loader: MapLoader{"PLOTS": `invalid`},
		},
		{
			name: "json: loader error",
			input: &struct {
				Plot Plot `env:"PLOT"`
			}{},
			wantErr: errContains(errors.New("loader error")),
			loader: LoaderFunc(func(ctx context.Context, key string) (string, error) {
				return "", errors.New("loader error")
			}),
		},
		{
			name: "json: empty required value",
			input: &struct {
				Plot Plot `env:"PLOT,required"`
			}{},
			wantErr: errContains(&ErrRequired{key: "PLOT"}),
			loader:  MapLoader{},
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_keyCollisions(t *testing.T) {
	t.Run("MissingNestedNillableKeyRegression", func(t *testing.T) {
		type ServerConfig struct {
			HTTP *xloadtype.Listener `env:"ADDRESS"`
		}
		type Config struct {
			Server ServerConfig `env:",prefix=SERVER_"`
		}

		cfg1 := new(Config)
		assert.NoError(t, Load(context.TODO(), cfg1))
		assert.Nil(t, cfg1.Server.HTTP)

		cfg2 := new(Config)
		assert.NoError(t, Load(context.TODO(), cfg2, MapLoader{"SERVER_ADDRESS": "127.0.0.1:80"}))
		assert.NotNil(t, cfg2.Server.HTTP)
		assert.Equal(t, net.IPv4(127, 0, 0, 1), cfg2.Server.HTTP.IP)
		assert.Equal(t, 80, cfg2.Server.HTTP.Port)
	})
}
