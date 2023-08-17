package xload

import (
	"context"
	"reflect"
	"sync"
)

// LoadAsync loads values concurrently by calling Loader in parallel.
// Number of goroutines can be controlled with xload.Concurrency.
func LoadAsync(ctx context.Context, v any, opts ...Option) error {
	o := newOptions(opts...)

	return processAsync(ctx, v, o.tagName, o.loader)
}

//nolint:funlen,nestif
func processAsync(ctx context.Context, obj any, tagKey string, loader Loader) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	value := v.Elem()
	if value.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	typ := value.Type()

	var wg sync.WaitGroup

	errors := make([]error, 0)

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
				loadAndSet := func(original reflect.Value, fVal reflect.Value, isNilStructPtr bool) {
					defer wg.Done()

					val, err := loader.Load(ctx, meta.name)
					if err != nil {
						errors = append(errors, err)

						return
					}

					if val == "" && meta.required {
						errors = append(errors, ErrRequired)

						return
					}

					if ok, err := decode(fVal, val); ok {
						if err != nil {
							errors = append(errors, err)

							return
						}

						setNilStructPtr(original, fVal, isNilStructPtr)
					}
				}

				wg.Add(1)

				go loadAndSet(value.Field(i), fVal, isNilStructPtr)

				continue
			}

			pld := loader
			if meta.prefix != "" {
				pld = PrefixLoader(meta.prefix, loader)
			}

			err := processAsync(ctx, fVal.Interface(), tagKey, pld)
			if err != nil {
				return err
			}

			setNilStructPtr(value.Field(i), fVal, isNilStructPtr)

			continue
		}

		if meta.prefix != "" {
			return ErrInvalidPrefix
		}

		loadAndSet := func(fVal reflect.Value) {
			defer wg.Done()

			// lookup value
			val, err := loader.Load(ctx, meta.name)
			if err != nil {
				errors = append(errors, err)

				return
			}

			if val == "" && meta.required {
				errors = append(errors, ErrRequired)

				return
			}

			// set value
			err = setVal(fVal, val, meta)
			if err != nil {
				errors = append(errors, err)

				return
			}
		}

		wg.Add(1)

		go loadAndSet(fVal)
	}

	wg.Wait()

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}
