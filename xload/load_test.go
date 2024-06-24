package xload

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name    string
	input   any
	want    any
	loader  Loader
	opts    []Option
	wantErr assert.ErrorAssertionFunc
}

func errContains(want error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgArgs ...interface{}) bool {
		return assert.ErrorContains(t, err, want.Error(), msgArgs...)
	}
}

func TestLoad_Default(t *testing.T) {
	cfg := &struct {
		Host string `env:"XLOAD_HOST"`
		Port int    `env:"XLOAD_PORT"`
	}{
		Port: 8080,
	}

	_ = os.Setenv("XLOAD_HOST", "localhost")
	// Port is intentionally not set using env var.

	err := Load(context.Background(), cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

func TestLoad_Errors(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		// not a pointer
		{
			name: "nil pointer",
			input: struct {
				Host string `env:"HOST"`
			}{},
			loader:  MapLoader{},
			wantErr: errContains(ErrNotPointer),
		},

		// not a struct
		{
			name:    "not a struct",
			input:   ptr.String(""),
			loader:  MapLoader{},
			wantErr: errContains(ErrNotStruct),
		},

		// private fields
		{
			name: "private fields",
			input: &struct {
				host string `env:"HOST"`
			}{
				host: "localhost",
			},
			want: &struct {
				host string `env:"HOST"`
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
				Host string `env:"-"`
			}{
				Host: "localhost",
			},
			want: &struct {
				Host string `env:"-"`
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
				Host string `env:"HOST"`
			}{},
			loader: LoaderFunc(func(ctx context.Context, k string) (string, error) {
				return "", errors.New("loader error")
			}),
			wantErr: errContains(errors.New("loader error")),
		},

		// unknown tag option
		{
			name: "unknown tag option",
			input: &struct {
				Host string `env:"HOST,unknown"`
			}{},
			loader:  MapLoader{},
			wantErr: errContains(&ErrUnknownTagOption{key: "HOST", opt: "unknown"}),
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_NativeTypes(t *testing.T) {
	t.Parallel()

	anyKind := reflect.TypeOf(new(any)).Elem().Kind()

	testcases := []testcase{
		// boolean value
		{
			name: "bool: true",
			input: &struct {
				Bool bool `env:"BOOL"`
			}{},
			want: &struct {
				Bool bool `env:"BOOL"`
			}{
				Bool: true,
			},
			loader: MapLoader{"BOOL": "true"},
		},
		{
			name: "bool: false",
			input: &struct {
				Bool bool `env:"BOOL"`
			}{},
			want: &struct {
				Bool bool `env:"BOOL"`
			}{
				Bool: false,
			},
			loader: MapLoader{"BOOL": "false"},
		},
		{
			name: "bool: invalid",
			input: &struct {
				Bool bool `env:"BOOL"`
			}{},
			loader:  MapLoader{"BOOL": "invalid"},
			wantErr: errContains(errors.New("invalid syntax")),
		},

		// integer values
		{
			name: "int, int8, int16, int32, int64",
			input: &struct {
				Int   int   `env:"INT"`
				Int8  int8  `env:"INT8"`
				Int16 int16 `env:"INT16"`
				Int32 int32 `env:"INT32"`
				Int64 int64 `env:"INT64"`
			}{},
			want: &struct {
				Int   int   `env:"INT"`
				Int8  int8  `env:"INT8"`
				Int16 int16 `env:"INT16"`
				Int32 int32 `env:"INT32"`
				Int64 int64 `env:"INT64"`
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
				Int int `env:"INT"`
			}{},
			loader: MapLoader{"INT": "invalid"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to int64`)
			},
		},

		// unsigned integer values
		{
			name: "uint, uint8, uint16, uint32, uint64",
			input: &struct {
				Uint   uint   `env:"UINT"`
				Uint8  uint8  `env:"UINT8"`
				Uint16 uint16 `env:"UINT16"`
				Uint32 uint32 `env:"UINT32"`
				Uint64 uint64 `env:"UINT64"`
			}{},
			want: &struct {
				Uint   uint   `env:"UINT"`
				Uint8  uint8  `env:"UINT8"`
				Uint16 uint16 `env:"UINT16"`
				Uint32 uint32 `env:"UINT32"`
				Uint64 uint64 `env:"UINT64"`
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
				Uint uint `env:"UINT"`
			}{},
			loader: MapLoader{"UINT": "invalid"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "UINT", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to uint64`)
			},
		},

		// floating-point values
		{
			name: "float32, float64",
			input: &struct {
				Float32 float32 `env:"FLOAT32"`
				Float64 float64 `env:"FLOAT64"`
			}{},
			want: &struct {
				Float32 float32 `env:"FLOAT32"`
				Float64 float64 `env:"FLOAT64"`
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
				Float float32 `env:"FLOAT"`
			}{},
			loader: MapLoader{"FLOAT": "invalid"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "FLOAT", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to float64`)
			},
		},

		// duration values
		{
			name: "duration",
			input: &struct {
				Duration    time.Duration  `env:"DURATION"`
				OptDuration *time.Duration `env:"OPT_DURATION"`
			}{},
			want: &struct {
				Duration    time.Duration  `env:"DURATION"`
				OptDuration *time.Duration `env:"OPT_DURATION"`
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
				Duration time.Duration `env:"DURATION"`
			}{},
			loader:  MapLoader{"DURATION": "invalid"},
			wantErr: errContains(errors.New("invalid duration")),
		},

		// string values
		{
			name: "string",
			input: &struct {
				String    string  `env:"STRING"`
				OptString *string `env:"OPT_STRING"`
			}{},
			want: &struct {
				String    string  `env:"STRING"`
				OptString *string `env:"OPT_STRING"`
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
				Bytes []byte `env:"BYTES"`
			}{},
			want: &struct {
				Bytes []byte `env:"BYTES"`
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
				StringSlice    []string  `env:"STRING_SLICE"`
				OptStringSlice []*string `env:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `env:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `env:"INT64_SLICE"`
				OptInt64Slice  []*int64  `env:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `env:"PTR_INT64_SLICE"`
			}{},
			want: &struct {
				StringSlice    []string  `env:"STRING_SLICE"`
				OptStringSlice []*string `env:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `env:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `env:"INT64_SLICE"`
				OptInt64Slice  []*int64  `env:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `env:"PTR_INT64_SLICE"`
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
				Int64Slice []int64 `env:"INT64_SLICE"`
			}{},
			loader: MapLoader{"INT64_SLICE": "invalid,2"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_SLICE", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to int64`)
			},
		},

		// map values
		{
			name: "map",
			input: &struct {
				StringMap    map[string]string  `env:"STRING_MAP"`
				PtrStringMap *map[string]string `env:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `env:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `env:"PTR_INT64_MAP"`
			}{},
			want: &struct {
				StringMap    map[string]string  `env:"STRING_MAP"`
				PtrStringMap *map[string]string `env:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `env:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `env:"PTR_INT64_MAP"`
			}{
				StringMap:    map[string]string{"key1": "value1", "key2": "value2"},
				PtrStringMap: &map[string]string{"key3": "value3", "key4": "value4"},
				Int64Map:     map[string]int64{"key5": 5, "key6": 6},
				PtrInt64Map:  &map[string]int64{"key7": 7, "key8": 8},
			},
			loader: MapLoader{
				"STRING_MAP":     "key1=value1,key2=value2",
				"PTR_STRING_MAP": "key3=value3,key4=value4",
				"INT64_MAP":      "key5=5,key6=6",
				"PTR_INT64_MAP":  "key7=7,key8=8",
			},
		},
		{
			name: "map: invalid separator",
			input: &struct {
				StringMap map[string]string `env:"STRING_MAP"`
			}{},
			loader:  MapLoader{"STRING_MAP": "key1::value1,key2::value2"},
			wantErr: errContains(&ErrInvalidMapValue{key: "STRING_MAP"}),
		},
		{
			name: "map: invalid value",
			input: &struct {
				Int64Map map[string]int64 `env:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1=1,key2=invalid"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_MAP", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to int64`)
			},
		},
		{
			name: "map: invalid key",
			input: &struct {
				Int64Map map[int]int64 `env:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1=1,key2=2"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_MAP", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "key1" of type string to int64`)
			},
		},

		// unknown key type
		{
			name: "unknown key type",
			input: &struct {
				Unknown interface{} `env:"UNKNOWN"`
			}{},
			loader:  MapLoader{"UNKNOWN": "1+2i"},
			wantErr: errContains(&ErrUnknownFieldType{field: "Unknown", key: "UNKNOWN", kind: anyKind}),
		},
		{
			name: "nested unknown key type",
			input: &struct {
				Nested struct {
					Unknown interface{} `env:"UNKNOWN"`
				} `env:",prefix=NESTED_"`
			}{},
			loader:  MapLoader{"NESTED_UNKNOWN": "1+2i"},
			wantErr: errContains(&ErrUnknownFieldType{field: "Unknown", key: "UNKNOWN", kind: anyKind}),
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_ArrayTypes(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "slice",
			input: &struct {
				StringSlice    []string  `env:"STRING_SLICE"`
				OptStringSlice []*string `env:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `env:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `env:"INT64_SLICE"`
				OptInt64Slice  []*int64  `env:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `env:"PTR_INT64_SLICE"`
			}{},
			want: &struct {
				StringSlice    []string  `env:"STRING_SLICE"`
				OptStringSlice []*string `env:"OPT_STRING_SLICE"`
				PtrStringSlice *[]string `env:"PTR_STRING_SLICE"`
				Int64Slice     []int64   `env:"INT64_SLICE"`
				OptInt64Slice  []*int64  `env:"OPT_INT64_SLICE"`
				PtrInt64Slice  *[]int64  `env:"PTR_INT64_SLICE"`
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
			name: "slice: custom delimiter",
			input: &struct {
				StringSlice []string `env:"STRING_SLICE,delimiter=;"`
			}{},
			want: &struct {
				StringSlice []string `env:"STRING_SLICE,delimiter=;"`
			}{
				StringSlice: []string{"string1", "string2", "string3"},
			},
			loader: MapLoader{"STRING_SLICE": "string1;string2;string3"},
		},
		{
			name: "slice: invalid value",
			input: &struct {
				Int64Slice []int64 `env:"INT64_SLICE"`
			}{},
			loader: MapLoader{"INT64_SLICE": "invalid,2"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_SLICE", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to int64`)
			},
		},
	}

	runTestcases(t, testcases)
}

func TestLoad_MapTypes(t *testing.T) {
	t.Parallel()

	testcases := []testcase{
		{
			name: "map",
			input: &struct {
				StringMap    map[string]string  `env:"STRING_MAP"`
				PtrStringMap *map[string]string `env:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `env:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `env:"PTR_INT64_MAP"`
			}{},
			want: &struct {
				StringMap    map[string]string  `env:"STRING_MAP"`
				PtrStringMap *map[string]string `env:"PTR_STRING_MAP"`
				Int64Map     map[string]int64   `env:"INT64_MAP"`
				PtrInt64Map  *map[string]int64  `env:"PTR_INT64_MAP"`
			}{
				StringMap:    map[string]string{"key1": "value1", "key2": "value2"},
				PtrStringMap: &map[string]string{"key3": "value3", "key4": "value4"},
				Int64Map:     map[string]int64{"key5": 5, "key6": 6},
				PtrInt64Map:  &map[string]int64{"key7": 7, "key8": 8},
			},
			loader: MapLoader{
				"STRING_MAP":     "key1=value1,key2=value2",
				"PTR_STRING_MAP": "key3=value3,key4=value4",
				"INT64_MAP":      "key5=5,key6=6",
				"PTR_INT64_MAP":  "key7=7,key8=8",
			},
		},
		{
			name: "map: custom delimiter",
			input: &struct {
				StringMap map[string]string `env:"STRING_MAP,delimiter=;"`
			}{},
			want: &struct {
				StringMap map[string]string `env:"STRING_MAP,delimiter=;"`
			}{
				StringMap: map[string]string{"key1": "value1", "key2": "value2"},
			},
			loader: MapLoader{"STRING_MAP": "key1=value1;key2=value2"},
		},
		{
			name: "map: custom separator",
			input: &struct {
				StringMap map[string]string `env:"STRING_MAP,separator=::"`
			}{},
			want: &struct {
				StringMap map[string]string `env:"STRING_MAP,separator=::"`
			}{
				StringMap: map[string]string{"key1": "value1", "key2": "value2"},
			},
			loader: MapLoader{"STRING_MAP": "key1::value1,key2::value2"},
		},
		{
			name: "map: invalid separator",
			input: &struct {
				StringMap map[string]string `env:"STRING_MAP"`
			}{},
			loader:  MapLoader{"STRING_MAP": "key1::value1,key2::value2"},
			wantErr: errContains(&ErrInvalidMapValue{key: "STRING_MAP"}),
		},
		{
			name: "map: invalid value",
			input: &struct {
				Int64Map map[string]int64 `env:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1=1,key2=invalid"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_MAP", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "invalid" of type string to int64`)
			},
		},
		{
			name: "map: invalid key",
			input: &struct {
				Int64Map map[int]int64 `env:"INT64_MAP"`
			}{},
			loader: MapLoader{"INT64_MAP": "key1=1,key2=2"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				errCast := new(ErrCast)
				return assert.ErrorAs(t, err, &errCast) && assert.Equal(t, "INT64_MAP", errCast.key) &&
					assert.EqualError(t, errCast.Unwrap(), `unable to cast "key1" of type string to int64`)
			},
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
				Binary CustomBinary `env:"BINARY"`
			}{},
			want: &struct {
				Binary CustomBinary `env:"BINARY"`
			}{
				Binary: CustomBinary("binary-value"),
			},
			loader: MapLoader{"BINARY": "value"},
		},
		{
			name: "custom type: gob",
			input: &struct {
				Gob CustomGob `env:"GOB"`
			}{},
			want: &struct {
				Gob CustomGob `env:"GOB"`
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
				Name string `env:"NAME,required"`
			}{},
			want: &struct{ Name string }{
				Name: "app1",
			},
			loader: MapLoader{"NAME": "app1"},
		},
		{
			name: "required option: missing value",
			input: &struct {
				Name string `env:"NAME,required"`
			}{},
			wantErr: errContains(&ErrRequired{key: "NAME"}),
			loader:  MapLoader{},
		},
		{
			name: "required custom decoder",
			input: &struct {
				Name CustomGob `env:"NAME,required"`
			}{},
			wantErr: errContains(&ErrRequired{key: "NAME"}),
			loader:  MapLoader{},
		},
		{
			name: "required option: empty value",
			input: &struct {
				Name *string `env:"NAME,required"`
			}{},
			wantErr: errContains(&ErrRequired{key: "NAME"}),
			loader:  MapLoader{"NAME": ""},
		},
		{
			name: "missing key",
			input: &struct {
				Name string `env:",required"`
			}{},
			wantErr: errContains(ErrMissingKey),
			loader:  MapLoader{},
		},
	}

	runTestcases(t, testcases)
}

func runTestcases(t *testing.T, testcases []testcase) {
	t.Helper()

	for _, tc := range testcases {
		tc := tc

		t.Run("Load_"+tc.name, func(t *testing.T) {
			err := Load(context.Background(), tc.input, append(tc.opts, WithLoader(tc.loader))...)
			if tc.wantErr != nil {
				tc.wantErr(t, err)

				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.want, tc.input)
		})

		t.Run("LoadAsync_"+tc.name, func(t *testing.T) {
			err := Load(context.Background(), tc.input, append(tc.opts, Concurrency(5), WithLoader(tc.loader))...)
			if tc.wantErr != nil {
				tc.wantErr(t, err)

				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.want, tc.input)
		})
	}
}
