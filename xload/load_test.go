package xload

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name   string
	input  any
	want   any
	loader Loader
	err    error
}

func TestLoadEnv(t *testing.T) {
	cfg := &struct {
		Host string `config:"XLOAD_HOST"`
	}{}

	os.Setenv("XLOAD_HOST", "localhost")

	err := LoadEnv(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
}

func TestLoad_Errors(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		// not a pointer
		{
			name: "nil pointer",
			input: struct {
				Host string `config:"HOST"`
			}{},
			loader: MapLoader{},
			err:    ErrNotPointer,
		},

		// not a struct
		{
			name:   "not a struct",
			input:  ptr.String(""),
			loader: MapLoader{},
			err:    ErrNotStruct,
		},

		// private fields
		{
			name: "private fields",
			input: &struct {
				host string `config:"HOST"`
			}{
				host: "localhost",
			},
			want: &struct {
				host string `config:"HOST"`
			}{
				host: "localhost",
			},
			loader: MapLoader{
				"HOST": "192.0.0.1",
			},
		},

		// skip fields
		{
			name: "skip fields",
			input: &struct {
				Host string `config:"-"`
			}{
				Host: "localhost",
			},
			want: &struct {
				Host string `config:"-"`
			}{
				Host: "localhost",
			},
			loader: MapLoader{
				"HOST": "192.0.0.1",
			},
		},

		// loader error
		{
			name: "loader error",
			input: &struct {
				Host string `config:"HOST"`
			}{},
			loader: LoaderFunc(func(ctx context.Context, k string) (string, error) {
				return "", errors.New("loader error")
			}),
			err: errors.New("loader error"),
		},

		// unknown tag option
		{
			name: "unknown tag option",
			input: &struct {
				Host string `config:"HOST,unknown"`
			}{},
			loader: MapLoader{},
			err:    ErrUnknownTagOption,
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_NativeTypes(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		// boolean value
		{
			name: "bool: true",
			input: &struct {
				Bool bool `config:"BOOL"`
			}{},
			want: &struct {
				Bool bool `config:"BOOL"`
			}{
				Bool: true,
			},
			loader: MapLoader{"BOOL": "true"},
		},
		{
			name: "bool: false",
			input: &struct {
				Bool bool `config:"BOOL"`
			}{},
			want: &struct {
				Bool bool `config:"BOOL"`
			}{
				Bool: false,
			},
			loader: MapLoader{"BOOL": "false"},
		},
		{
			name: "bool: invalid",
			input: &struct {
				Bool bool `config:"BOOL"`
			}{},
			loader: MapLoader{"BOOL": "invalid"},
			err:    errors.New("invalid syntax"),
		},

		// integer values
		{
			name: "int, int8, int16, int32, int64",
			input: &struct {
				Int   int   `config:"INT"`
				Int8  int8  `config:"INT8"`
				Int16 int16 `config:"INT16"`
				Int32 int32 `config:"INT32"`
				Int64 int64 `config:"INT64"`
			}{},
			want: &struct {
				Int   int   `config:"INT"`
				Int8  int8  `config:"INT8"`
				Int16 int16 `config:"INT16"`
				Int32 int32 `config:"INT32"`
				Int64 int64 `config:"INT64"`
			}{
				Int:   1,
				Int8:  12,
				Int16: 123,
				Int32: 1234,
				Int64: 12345,
			},
			loader: MapLoader{
				"INT":   "1",
				"INT8":  "12",
				"INT16": "123",
				"INT32": "1234",
				"INT64": "12345",
			},
		},
		{
			name: "int: invalid",
			input: &struct {
				Int int `config:"INT"`
			}{},
			loader: MapLoader{"INT": "invalid"},
			err:    errors.New("unable to cast"),
		},

		// unsigned integer values
		{
			name: "uint, uint8, uint16, uint32, uint64",
			input: &struct {
				Uint   uint   `config:"UINT"`
				Uint8  uint8  `config:"UINT8"`
				Uint16 uint16 `config:"UINT16"`
				Uint32 uint32 `config:"UINT32"`
				Uint64 uint64 `config:"UINT64"`
			}{},
			want: &struct {
				Uint   uint   `config:"UINT"`
				Uint8  uint8  `config:"UINT8"`
				Uint16 uint16 `config:"UINT16"`
				Uint32 uint32 `config:"UINT32"`
				Uint64 uint64 `config:"UINT64"`
			}{
				Uint:   1,
				Uint8:  12,
				Uint16: 123,
				Uint32: 1234,
				Uint64: 12345,
			},
			loader: MapLoader{
				"UINT":   "1",
				"UINT8":  "12",
				"UINT16": "123",
				"UINT32": "1234",
				"UINT64": "12345",
			},
		},
		{
			name: "uint: invalid",
			input: &struct {
				Uint uint `config:"UINT"`
			}{},
			loader: MapLoader{"UINT": "invalid"},
			err:    errors.New("unable to cast"),
		},

		// floating-point values
		{
			name: "float32, float64",
			input: &struct {
				Float32 float32 `config:"FLOAT32"`
				Float64 float64 `config:"FLOAT64"`
			}{},
			want: &struct {
				Float32 float32 `config:"FLOAT32"`
				Float64 float64 `config:"FLOAT64"`
			}{
				Float32: 1.23,
				Float64: 1.2345,
			},
			loader: MapLoader{
				"FLOAT32": "1.23",
				"FLOAT64": "1.2345",
			},
		},
		{
			name: "float: invalid",
			input: &struct {
				Float float32 `config:"FLOAT"`
			}{},
			loader: MapLoader{"FLOAT": "invalid"},
			err:    errors.New("unable to cast"),
		},

		// duration values
		{
			name: "duration",
			input: &struct {
				Duration    time.Duration  `config:"DURATION"`
				OptDuration *time.Duration `config:"OPT_DURATION"`
			}{},
			want: &struct {
				Duration    time.Duration  `config:"DURATION"`
				OptDuration *time.Duration `config:"OPT_DURATION"`
			}{
				Duration:    10 * time.Second,
				OptDuration: ptr.Duration(10 * time.Second),
			},
			loader: MapLoader{
				"DURATION":     "10s",
				"OPT_DURATION": "10s",
			},
		},
		{
			name: "duration: invalid",
			input: &struct {
				Duration time.Duration `config:"DURATION"`
			}{},
			loader: MapLoader{"DURATION": "invalid"},
			err:    errors.New("invalid duration"),
		},

		// string values
		{
			name: "string",
			input: &struct {
				String    string  `config:"STRING"`
				OptString *string `config:"OPT_STRING"`
			}{},
			want: &struct {
				String    string  `config:"STRING"`
				OptString *string `config:"OPT_STRING"`
			}{
				String:    "string",
				OptString: ptr.String("string"),
			},
			loader: MapLoader{
				"STRING":     "string",
				"OPT_STRING": "string",
			},
		},

		// byte values
		{
			name: "byte array",
			input: &struct {
				Bytes []byte `config:"BYTES"`
			}{},
			want: &struct {
				Bytes []byte `config:"BYTES"`
			}{
				Bytes: []byte("bytes"),
			},
			loader: MapLoader{
				"BYTES": "bytes",
			},
		},

		// slice values
		{
			name: "slice",
			input: &struct {
				StringSlice    []string  `config:"STRING_SLICE"`
				OptStringSlice []*string `config:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `config:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `config:"INT64_SLICE"`
				OptInt64Slice  []*int64  `config:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `config:"PTR_INT64_SLICE"`
			}{},
			want: &struct {
				StringSlice    []string  `config:"STRING_SLICE"`
				OptStringSlice []*string `config:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `config:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `config:"INT64_SLICE"`
				OptInt64Slice  []*int64  `config:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `config:"PTR_INT64_SLICE"`
			}{
				StringSlice:    []string{"string1", "string2"},
				OptStringSlice: []*string{ptr.String("string3"), ptr.String("string4")},
				PtrStringSlice: &[]string{"string5", "string6"},
				Int64Slice:     []int64{1, 2},
				OptInt64Slice:  []*int64{ptr.Int64(3), ptr.Int64(4)},
				PtrInt64Slice:  &[]int64{5, 6},
			},
			loader: MapLoader{
				"STRING_SLICE":     "string1,string2",
				"OPT_STRING_SLICE": "string3,string4",
				"PTR_STRING_SLICE": "string5,string6",
				"INT64_SLICE":      "1,2",
				"OPT_INT64_SLICE":  "3,4",
				"PTR_INT64_SLICE":  "5,6",
			},
		},
		{
			name: "slice: invalid value",
			input: &struct {
				Int64Slice []int64 `config:"INT64_SLICE"`
			}{},
			loader: MapLoader{"INT64_SLICE": "invalid,2"},
			err:    errors.New("unable to cast"),
		},

		// map values
		{
			name: "map",
			input: &struct {
				StringMap    map[string]string  `config:"STRING_MAP"`
				PtrStringMap *map[string]string `config:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `config:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `config:"PTR_INT64_MAP"`
			}{},
			want: &struct {
				StringMap    map[string]string  `config:"STRING_MAP"`
				PtrStringMap *map[string]string `config:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `config:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `config:"PTR_INT64_MAP"`
			}{
				StringMap:    map[string]string{"key1": "value1", "key2": "value2"},
				PtrStringMap: &map[string]string{"key3": "value3", "key4": "value4"},
				Int64Map:     map[string]int64{"key5": 5, "key6": 6},
				PtrInt64Map:  &map[string]int64{"key7": 7, "key8": 8},
			},
			loader: MapLoader{
				"STRING_MAP":     "key1:value1,key2:value2",
				"PTR_STRING_MAP": "key3:value3,key4:value4",
				"INT64_MAP":      "key5:5,key6:6",
				"PTR_INT64_MAP":  "key7:7,key8:8",
			},
		},
		{
			name: "map: invalid delimiter",
			input: &struct {
				StringMap map[string]string `config:"STRING_MAP"`
			}{},
			loader: MapLoader{"STRING_MAP": "key1=value1,key2=value2"},
			err:    ErrInvalidMapValue,
		},
		{
			name: "map: invalid delimiter",
			input: &struct {
				Int64Map map[string]int64 `config:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1=1,key2=invalid"},
			err:    ErrInvalidMapValue,
		},
		{
			name: "map: invalid value",
			input: &struct {
				Int64Map map[string]int64 `config:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1:1,key2:invalid"},
			err:    errors.New("unable to cast"),
		},
		{
			name: "map: invalid key",
			input: &struct {
				Int64Map map[int]int64 `config:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1:1,key2:2"},
			err:    errors.New("unable to cast"),
		},
	}

	runTestcases(t, testcases)
}

type CustomBinary string

func (c *CustomBinary) UnmarshalBinary(data []byte) error {
	val := CustomBinary("binary-" + string(data))

	*c = val

	return nil
}

type CustomGob string

func (c *CustomGob) GobDecode(data []byte) error {
	val := CustomGob("gob-" + string(data))

	*c = val

	return nil
}

func TestLoad_CustomType(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "custom type: binary",
			input: &struct {
				Binary CustomBinary `config:"BINARY"`
			}{},
			want: &struct {
				Binary CustomBinary `config:"BINARY"`
			}{
				Binary: CustomBinary("binary-value"),
			},
			loader: MapLoader{"BINARY": "value"},
		},
		{
			name: "custom type: gob",
			input: &struct {
				Gob CustomGob `config:"GOB"`
			}{},
			want: &struct {
				Gob CustomGob `config:"GOB"`
			}{
				Gob: CustomGob("gob-value"),
			},
			loader: MapLoader{"GOB": "value"},
		},
	}

	runTestcases(t, testcases)
}

func TestOption_Required(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "required option",
			input: &struct {
				Name string `config:"NAME,required"`
			}{},
			want: &struct{ Name string }{
				Name: "app1",
			},
			loader: MapLoader{"NAME": "app1"},
		},
		{
			name: "required option: missing value",
			input: &struct {
				Name string `config:"NAME,required"`
			}{},
			err:    ErrRequired,
			loader: MapLoader{},
		},
		{
			name: "required option: empty value",
			input: &struct {
				Name *string `config:"NAME,required"`
			}{},
			err:    ErrRequired,
			loader: MapLoader{"NAME": ""},
		},
		{
			name: "missing key",
			input: &struct {
				Name string `config:",required"`
			}{},
			err:    ErrMissingKey,
			loader: MapLoader{},
		},
	}

	runTestcases(t, testcases)
}

func runTestcases(t *testing.T, testcases []testcase) {
	t.Helper()

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Load(context.Background(), tc.input, WithLoader(tc.loader))
			if tc.err != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.err.Error())

				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.want, tc.input)
		})
	}
}
