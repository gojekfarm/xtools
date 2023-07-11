package xconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//nolint:funlen
func setFieldValue(v string, ef reflect.Value, opts *tagOptions) error {
	// Handle pointers and uninitialized pointers.
	for ef.Type().Kind() == reflect.Ptr {
		if ef.IsNil() {
			ef.Set(reflect.New(ef.Type().Elem()))
		}

		ef = ef.Elem()
	}

	tf := ef.Type()
	tk := tf.Kind()

	// Handle existing decoders.
	if ok, err := decode(v, ef); ok {
		return err
	}

	// We don't check if the value is empty earlier, because the user might want
	// to define a custom decoder and treat the empty variable as a special case.
	// However, if we got this far, none of the remaining parsers will succeed, so
	// bail out now.
	if v == "" {
		return nil
	}

	switch tk {
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}

		ef.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(v, tf.Bits())
		if err != nil {
			return err
		}

		ef.SetFloat(f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		i, err := strconv.ParseInt(v, 0, tf.Bits())
		if err != nil {
			return err
		}

		ef.SetInt(i)
	case reflect.Int64:
		// Special case time.Duration values.
		if tf.PkgPath() == "time" && tf.Name() == "Duration" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return err
			}

			ef.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(v, 0, tf.Bits())
			if err != nil {
				return err
			}

			ef.SetInt(i)
		}

	// Strings
	case reflect.String:
		ef.SetString(v)

	// Unsigned integers
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(v, 0, tf.Bits())
		if err != nil {
			return err
		}

		ef.SetUint(i)

	case reflect.Interface:
		return fmt.Errorf("cannot decode into interfaces")

	// Maps
	case reflect.Map:
		vals := strings.Split(v, opts.delimiter)
		mp := reflect.MakeMapWithSize(tf, len(vals))

		for _, val := range vals {
			pair := strings.SplitN(val, opts.separator, 2)
			if len(pair) < 2 {
				return fmt.Errorf("%s: %w", val, ErrInvalidMapItem)
			}

			mKey, mVal := strings.TrimSpace(pair[0]), strings.TrimSpace(pair[1])

			k := reflect.New(tf.Key()).Elem()
			if err := setFieldValue(mKey, k, opts); err != nil {
				return fmt.Errorf("%s: %w", mKey, err)
			}

			v := reflect.New(tf.Elem()).Elem()
			if err := setFieldValue(mVal, v, opts); err != nil {
				return fmt.Errorf("%s: %w", mVal, err)
			}

			mp.SetMapIndex(k, v)
		}

		ef.Set(mp)

	// Slices
	case reflect.Slice:
		// Special case: []byte
		if tf.Elem().Kind() == reflect.Uint8 {
			ef.Set(reflect.ValueOf([]byte(v)))
		} else {
			vals := strings.Split(v, opts.delimiter)
			s := reflect.MakeSlice(tf, len(vals), len(vals))
			for i, val := range vals {
				val = strings.TrimSpace(val)
				if err := setFieldValue(val, s.Index(i), opts); err != nil {
					return fmt.Errorf("%s: %w", val, err)
				}
			}

			ef.Set(s)
		}
	}

	return nil
}
