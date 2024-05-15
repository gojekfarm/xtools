package xload

import (
	"context"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"reflect"

	"github.com/sourcegraph/conc/pool"
)

type loadAndSet func(context.Context, reflect.Value) error
type loadAndSetPointer func(context.Context, reflect.Value, reflect.Value, bool) error

func processConcurrently(ctx context.Context, v any, o *options) error {
	if !o.detectCollisions {
		return doProcessConcurrently(ctx, v, o)
	}

	syncKeyUsage := &collisionSyncMap{}
	ldr := o.loader
	o.loader = LoaderFunc(func(ctx context.Context, key string) (string, error) {
		v, err := ldr.Load(ctx, key)

		if err == nil {
			syncKeyUsage.add(key)
		}

		return v, err
	})

	if err := doProcessConcurrently(ctx, v, o); err != nil {
		return err
	}

	return syncKeyUsage.err()
}

func doProcessConcurrently(ctx context.Context, v any, opts *options) error {
	doneCh := make(chan struct{}, 1)
	defer close(doneCh)

	p := pool.New().WithContext(ctx).WithMaxGoroutines(opts.concurrency).WithCancelOnError()

	err := processAsync(p, opts, opts.loader, v, func() {
		doneCh <- struct{}{}
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return err
	}
}

//nolint:funlen,nestif
func processAsync(p *pool.ContextPool, o *options, loader Loader, obj any, cb func()) error {
	if cb != nil {
		defer cb()
	}

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

		tag := fTyp.Tag.Get(o.tagName)

		if tag == "-" {
			continue
		}

		meta, err := parseField(fTyp.Name, tag)
		if err != nil {
			return err
		}

		// handle pointers to structs
		isNilStructPtr := false

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
			if meta.key != "" && hasDecoder(fVal) {
				las := loadAndSetWithOriginal(loader, meta)

				original := value.Field(i)

				p.Go(func(ctx context.Context) error { return las(ctx, original, fVal, isNilStructPtr) })

				continue
			}

			pld := loader
			if meta.prefix != "" {
				pld = PrefixLoader(meta.prefix, loader)
			}

			err := processAsync(p, o, pld, fVal.Interface(), nil)
			if err != nil {
				return err
			}

			setNilStructPtr(value.Field(i), fVal, isNilStructPtr)

			continue
		}

		if meta.prefix != "" {
			return &ErrInvalidPrefix{field: fTyp.Name, kind: fVal.Kind()}
		}

		las := loadAndSetVal(loader, meta)

		p.Go(func(ctx context.Context) error { return las(ctx, fVal) })
	}

	if cb == nil {
		return nil
	}

	return p.Wait()
}

func setNilStructPtr(original reflect.Value, v reflect.Value, isNilStructPtr bool) {
	if isNilStructPtr {
		empty := reflect.New(original.Type().Elem()).Interface()

		if !reflect.DeepEqual(empty, v.Interface()) && original.CanSet() {
			original.Set(v)
		}
	}
}

func loadAndSetWithOriginal(loader Loader, meta *field) loadAndSetPointer {
	return func(ctx context.Context, original reflect.Value, fVal reflect.Value, isNilStructPtr bool) error {
		val, err := loader.Load(ctx, meta.key)
		if err != nil {
			return err
		}

		if val == "" && meta.required {
			return &ErrRequired{key: meta.key}
		}

		if ok, err := decode(fVal, val); ok {
			if err != nil {
				return err
			}

			setNilStructPtr(original, fVal, isNilStructPtr)
		}

		return nil
	}
}

func loadAndSetVal(loader Loader, meta *field) loadAndSet {
	return func(ctx context.Context, fVal reflect.Value) error {
		// lookup value
		val, err := loader.Load(ctx, meta.key)
		if err != nil {
			return err
		}

		if val == "" && meta.required {
			return &ErrRequired{key: meta.key}
		}

		// set value
		err = setVal(fVal, val, meta)
		if err != nil {
			return err
		}

		return nil
	}
}

func hasDecoder(field reflect.Value) bool {
	for field.CanAddr() {
		field = field.Addr()
	}

	if field.CanInterface() {
		switch field.Interface().(type) {
		case Decoder, encoding.TextUnmarshaler, json.Unmarshaler, encoding.BinaryUnmarshaler, gob.GobDecoder:
			return true
		}
	}

	return false
}
