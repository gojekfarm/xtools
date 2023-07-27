package xload

import (
	"context"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

var (
	// ErrNotPointer is returned when the given config is not a pointer.
	ErrNotPointer = errors.New("xload: config must be a pointer")
	// ErrNotStruct is returned when the given config is not a struct.
	ErrNotStruct = errors.New("xload: config must be a struct")
	// ErrUnknownTagOption is returned when an unknown tag option is used.
	ErrUnknownTagOption = errors.New("xload: unknown tag option")
	// ErrRequired is returned when a required field is missing.
	ErrRequired = errors.New("xload: missing required value")
	// ErrUnknownFieldType is returned when the field type is not supported.
	ErrUnknownFieldType = errors.New("xload: unknown field type")
	// ErrInvalidMapValue is returned when the map value is invalid.
	ErrInvalidMapValue = errors.New("xload: invalid map value")
	// ErrMissingKey is returned when the key is missing from the tag.
	ErrMissingKey = errors.New("xload: missing key")
	// ErrInvalidPrefix is returned when the prefix option is used on a non-struct field.
	ErrInvalidPrefix = errors.New("xload: prefix is only valid on struct types")
)

const (
	optRequired = "required"
	optPrefix   = "prefix="
)

// LoadEnv loads values from OS environment using default options.
func LoadEnv(ctx context.Context, v any) error {
	return Load(ctx, v)
}

// Load loads values into the given struct using the given options.
func Load(ctx context.Context, v any, opts ...Option) error {
	o := newOptions(opts...)

	err := process(ctx, v, o.key, o.loader)
	if err != nil {
		return err
	}

	return nil
}

//nolint:funlen,nestif
func process(ctx context.Context, obj any, tagKey string, loader Loader) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	value := v.Elem()
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := value.Type()

	for i := 0; i < typ.NumField(); i++ {
		fTyp := typ.Field(i)
		fVal := value.Field(i)

		// skip unexported fields
		if !fVal.CanSet() {
			continue
		}

		tag := fTyp.Tag.Get(tagKey)

		if tag == "-" {
			continue
		}

		meta, err := parseField(tag)
		if err != nil {
			return err
		}

		// handle pointers to structs
		isNilStructPtr := false
		setNilStructPtr := func(v reflect.Value) {
			original := value.Field(i)

			if isNilStructPtr {
				empty := reflect.New(original.Type().Elem()).Interface()

				if !reflect.DeepEqual(empty, v.Interface()) {
					original.Set(v)
				}
			}
		}

		// initialise pointer to structs
		for fVal.Kind() == reflect.Ptr {
			if fVal.IsNil() && fVal.Type().Elem().Kind() != reflect.Struct {
				break
			}

			if fVal.IsNil() {
				isNilStructPtr = true
				fVal = reflect.New(fVal.Type().Elem())
			}

			fVal = fVal.Elem()
		}

		// handle structs
		if fVal.Kind() == reflect.Struct {
			for fVal.CanAddr() {
				fVal = fVal.Addr()
			}

			// if the struct has a key, load it
			// and set the value to the struct
			if meta.name != "" {
				val, err := loader.Load(ctx, meta.name)
				if err != nil {
					return err
				}

				if val == "" && meta.required {
					return ErrRequired
				}

				if ok, err := decode(fVal, val); ok {
					if err != nil {
						return err
					}

					setNilStructPtr(fVal)

					continue
				}
			}

			pld := loader
			if meta.prefix != "" {
				pld = PrefixLoader(meta.prefix, loader)
			}

			err := process(ctx, fVal.Interface(), tagKey, pld)
			if err != nil {
				return err
			}

			setNilStructPtr(fVal)

			continue
		}

		if meta.prefix != "" {
			return ErrInvalidPrefix
		}

		// lookup value
		val, err := loader.Load(ctx, meta.name)
		if err != nil {
			return err
		}

		if val == "" && meta.required {
			return ErrRequired
		}

		// set value
		err = setVal(fVal, val)
		if err != nil {
			return err
		}
	}

	return nil
}

type field struct {
	name     string
	prefix   string
	required bool
}

func parseField(tag string) (*field, error) {
	parts := strings.Split(tag, ",")
	key, tagOpts := strings.TrimSpace(parts[0]), parts[1:]

	f := &field{
		name: key,
	}

	for _, opt := range tagOpts {
		opt = strings.TrimSpace(opt)

		switch {
		case opt == optRequired:
			if key == "" {
				return nil, ErrMissingKey
			}

			f.required = true
		case strings.HasPrefix(opt, optPrefix):
			f.prefix = strings.TrimPrefix(opt, optPrefix)
		default:
			return nil, ErrUnknownTagOption
		}
	}

	return f, nil
}

//nolint:funlen,nestif
func setVal(field reflect.Value, val string) error {
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		field = field.Elem()
	}

	if val == "" {
		return nil
	}

	dec, err := decode(field, val)
	if dec || err != nil {
		return err
	}

	ty := field.Type()
	kd := field.Kind()

	switch kd {
	case reflect.String:
		field.SetString(val)

	case reflect.Bool:
		b, err := cast.ToBoolE(val)
		if err != nil {
			return err
		}

		field.SetBool(b)

	case reflect.Int64:
		// special case for time.Duration
		if ty.String() == "time.Duration" {
			d, err := cast.ToDurationE(val)
			if err != nil {
				return err
			}

			field.SetInt(int64(d))

			return nil
		}

		i, err := cast.ToInt64E(val)
		if err != nil {
			return err
		}

		field.SetInt(i)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		i, err := cast.ToInt64E(val)
		if err != nil {
			return err
		}

		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := cast.ToUint64E(val)
		if err != nil {
			return err
		}

		field.SetUint(i)

	case reflect.Float32, reflect.Float64:
		f, err := cast.ToFloat64E(val)
		if err != nil {
			return err
		}

		field.SetFloat(f)

	case reflect.Map:
		vals := strings.Split(val, ",")
		m := reflect.MakeMapWithSize(ty, len(vals))

		for _, v := range vals {
			kv := strings.Split(v, ":")
			if len(kv) != 2 {
				return ErrInvalidMapValue
			}

			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])

			key := reflect.New(ty.Key()).Elem()

			err := setVal(key, k)
			if err != nil {
				return err
			}

			value := reflect.New(ty.Elem()).Elem()

			err = setVal(value, v)
			if err != nil {
				return err
			}

			m.SetMapIndex(key, value)
		}

		field.Set(m)

	case reflect.Slice:
		// special case for []byte
		if ty.Elem().Kind() == reflect.Uint8 {
			field.SetBytes([]byte(val))

			return nil
		}

		vals := strings.Split(val, ",")
		slice := reflect.MakeSlice(ty, len(vals), len(vals))

		for i, v := range vals {
			v = strings.TrimSpace(v)

			err := setVal(slice.Index(i), v)
			if err != nil {
				return err
			}
		}

		field.Set(slice)

	default:
		return ErrUnknownFieldType
	}

	return nil
}

// Decoder is the interface for custom decoders.
type Decoder interface {
	Decode(string) error
}

// decode decodes the given value using custom decoder if available.
// If not, it uses one of the default decoders:
// - encoding.TextUnmarshaler
// - json.Unmarshaler
// - encoding.BinaryUnmarshaler
// - encoding.GobDecoder
//
//nolint:nestif
func decode(field reflect.Value, val string) (bool, error) {
	for field.CanAddr() {
		field = field.Addr()
	}

	if field.CanInterface() {
		iface := field.Interface()

		if dec, ok := iface.(Decoder); ok {
			return true, dec.Decode(val)
		}

		if txt, ok := iface.(encoding.TextUnmarshaler); ok {
			return true, txt.UnmarshalText([]byte(val))
		}

		if json, ok := iface.(json.Unmarshaler); ok {
			return true, json.UnmarshalJSON([]byte(val))
		}

		if bin, ok := iface.(encoding.BinaryUnmarshaler); ok {
			return true, bin.UnmarshalBinary([]byte(val))
		}

		if gob, ok := iface.(gob.GobDecoder); ok {
			return true, gob.GobDecode([]byte(val))
		}
	}

	return false, nil
}