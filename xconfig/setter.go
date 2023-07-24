package xconfig

import (
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

func setVal(field reflect.Value, val string) error {
	for field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		field = field.Elem()
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
