package xload

import (
	"context"
	"reflect"

	"github.com/sourcegraph/conc/pool"
)

//nolint:funlen,nestif
func processAsync(ctx context.Context, obj any, o *options, loader Loader) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	value := v.Elem()
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := value.Type()

	p := pool.New().WithErrors().WithMaxGoroutines(o.concurrency)

	for i := 0; i < typ.NumField(); i++ {
		fTyp := typ.Field(i)
		fVal := value.Field(i)

		// skip unexported fields
		if !fVal.CanSet() {
			continue
		}

		tag := fTyp.Tag.Get(o.tagName)

		if tag == "-" {
			continue
		}

		meta, err := parseField(tag)
		if err != nil {
			return err
		}

		// handle pointers to structs
		isNilStructPtr := false
		setNilStructPtr := func(original reflect.Value, v reflect.Value, isNilStructPtr bool) {
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
			if meta.name != "" && hasDecoder(fVal) {
				loadAndSet := func(original reflect.Value, fVal reflect.Value, isNilStructPtr bool) error {
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

						setNilStructPtr(original, fVal, isNilStructPtr)
					}

					return nil
				}

				original := value.Field(i)

				p.Go(func() error {
					return loadAndSet(original, fVal, isNilStructPtr)
				})

				continue
			}

			pld := loader
			if meta.prefix != "" {
				pld = PrefixLoader(meta.prefix, loader)
			}

			err := processAsync(ctx, fVal.Interface(), o, pld)
			if err != nil {
				return err
			}

			setNilStructPtr(value.Field(i), fVal, isNilStructPtr)

			continue
		}

		if meta.prefix != "" {
			return ErrInvalidPrefix
		}

		loadAndSet := func(fVal reflect.Value) error {
			// lookup value
			val, err := loader.Load(ctx, meta.name)
			if err != nil {
				return err
			}

			if val == "" && meta.required {
				return ErrRequired
			}

			// set value
			err = setVal(fVal, val, meta)
			if err != nil {
				return err
			}

			return nil
		}

		p.Go(func() error {
			return loadAndSet(fVal)
		})
	}

	return p.Wait()
}
