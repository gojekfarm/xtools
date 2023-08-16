package xload

import (
	"context"
	"reflect"

	"github.com/sourcegraph/conc/pool"
)

// LoadAsync loads values concurrently by calling Loader in parallel.
// Number of goroutines can be controlled with xload.Concurrency.
func LoadAsync(ctx context.Context, v any, opts ...Option) error {
	o := newOptions(opts...)

	p := processor{
		pool: pool.New().WithErrors().WithMaxGoroutines(o.concurrency),
		opts: o,
	}

	err := p.run(ctx, v, o.loader)
	if err != nil {
		return err
	}

	return p.pool.Wait()
}

type processor struct {
	pool *pool.ErrorPool
	opts *options
}

func (p *processor) run(ctx context.Context, obj any, loader Loader) error {
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

		p.pool.Go(func() error {
			return p.processField(ctx, fTyp, fVal, loader)
		})
	}

	return nil
}

func (p *processor) processField(ctx context.Context, fTyp reflect.StructField, fVal reflect.Value, loader Loader) error {

	tag := fTyp.Tag.Get(p.opts.tagName)

	if tag == "-" {
		return nil
	}

	meta, err := parseField(tag)
	if err != nil {
		return err
	}

	// handle pointers to structs
	isNilStructPtr := false
	setNilStructPtr := func(v reflect.Value) {
		original := fVal

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

				return nil
			}
		}

		pld := loader
		if meta.prefix != "" {
			pld = PrefixLoader(meta.prefix, loader)
		}

		err := p.run(ctx, fVal.Interface(), pld)
		if err != nil {
			return err
		}

		setNilStructPtr(fVal)

		return nil
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
	err = setVal(fVal, val, meta)
	if err != nil {
		return err
	}

	return nil
}
