package xconfig

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

const (
	optRequired  = "required"
	optDelimiter = "delimiter="
	optPrefix    = "prefix="
	optSeparator = "separator="

	defaultDelimiter = ","
	defaultSeparator = ":"
)

// Error is a custom error type for errors returned by envconfig.
type Error string

// Error implements error.
func (e Error) Error() string {
	return string(e)
}

// Error messages for various errors that can occur during configuration loading and parsing.
const (
	ErrInvalidEnvvarName  = Error("invalid environment variable name")
	ErrInvalidMapItem     = Error("invalid map item")
	ErrLoaderNil          = Error("lookuper cannot be nil")
	ErrMissingKey         = Error("missing key")
	ErrMissingRequired    = Error("missing required value")
	ErrNoInitNotPtr       = Error("field must be a pointer to have noinit")
	ErrNotPtr             = Error("input must be a pointer")
	ErrNotStruct          = Error("input must be a struct")
	ErrPrefixNotStruct    = Error("prefix is only valid on struct types")
	ErrPrivateField       = Error("cannot parse private fields")
	ErrRequiredAndDefault = Error("field cannot be required and have a default value")
	ErrUnknownOption      = Error("unknown option")
)

// Decoder is an interface that custom types/fields can implement to control how
// decoding takes place. For example:
//
//	type MyType string
//
//	func (mt MyType) Decode(val string) error {
//	    return "CUSTOM-"+val
//	}
type Decoder interface {
	Decode(val string) error
}

// Load loads the configuration from environment variables.
func Load(ctx context.Context, cfg any) error {
	return LoadWith(ctx, cfg, CustomLoader(OSLoader()))
}

// LoadWith loads the configuration using the given options.
func LoadWith(ctx context.Context, cfg any, opts ...Option) error {
	o := newOptions(opts...)

	return process(ctx, cfg, o, false)
}

//nolint:funlen,nestif,gocyclo
func process(ctx context.Context, cfg any, o *options, parentNoInit bool) error {
	if o.loader == nil {
		return ErrLoaderNil
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return ErrNotPtr
	}

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	t := e.Type()

	for i := 0; i < t.NumField(); i++ {
		ef := e.Field(i)
		tf := t.Field(i)
		tag := tf.Tag.Get(o.tag)

		if !ef.CanSet() {
			if tag != "" {
				// There's an "env" tag on a private field, we can't alter it, and it's
				// likely a mistake. Return an error so the user can handle.
				return fmt.Errorf("%s: %w", tf.Name, ErrPrivateField)
			}

			// Otherwise continue to the next field.
			continue
		}

		// Parse the key and options.
		key, tagOpts, err := parseTag(tag)
		if err != nil {
			return fmt.Errorf("%s: %w", tf.Name, err)
		}

		isNilStructPtr := false
		setNilStruct := func(v reflect.Value) {
			origin := e.Field(i)
			if isNilStructPtr {
				empty := reflect.New(origin.Type().Elem()).Interface()

				// If a struct (after traversal) equals to the empty value, it means
				// nothing was changed in any sub-fields. With the noinit opt, we skip
				// setting the empty value to the original struct pointer (keep it nil).
				if !reflect.DeepEqual(v.Interface(), empty) || !parentNoInit {
					origin.Set(v)
				}
			}
		}

		// Initialise pointer structs.
		pointerWasSet := false

		for ef.Kind() == reflect.Ptr {
			if ef.IsNil() {
				if ef.Type().Elem().Kind() != reflect.Struct {
					// This is a nil pointer to something that isn't a struct, like
					// *string. Move along.
					break
				}

				isNilStructPtr = true
				// Use an empty struct of the type so we can traverse.
				ef = reflect.New(ef.Type().Elem()).Elem()
			} else {
				pointerWasSet = true
				ef = ef.Elem()
			}
		}

		// Special case handle structs. This has to come after the value resolution in
		// case the struct has a custom decoder.
		if ef.Kind() == reflect.Struct {
			for ef.CanAddr() {
				ef = ef.Addr()
			}

			// Load the value, ignoring an error if the key isn't defined. This is
			// required for nested structs that don't declare their own `env` keys,
			// but have internal fields with an `env` defined.
			val, _, err := lookup(ctx, key, tagOpts, o.loader)
			if err != nil && !errors.Is(err, ErrMissingKey) {
				return fmt.Errorf("%s: %w", tf.Name, err)
			}

			if ok, err := decode(val, ef); ok {
				if err != nil {
					return err
				}

				setNilStruct(ef)

				continue
			}

			plu := o.clone()
			if tagOpts.prefix != "" {
				plu.loader = PrefixLoader(tagOpts.prefix, plu.loader)
			}

			if err := process(ctx, ef.Interface(), plu, parentNoInit); err != nil {
				return fmt.Errorf("%s: %w", tf.Name, err)
			}

			setNilStruct(ef)

			continue
		}

		// Stop processing if there's no env tag (this comes after nested parsing),
		// in case there's an env tag in an embedded struct.
		if tag == "" {
			continue
		}

		// It's invalid to have a prefix on a non-struct field.
		if tagOpts.prefix != "" {
			return ErrPrefixNotStruct
		}

		// The field already has a non-zero value and overwrite is false, do not
		// overwrite.
		if pointerWasSet || !ef.IsZero() {
			continue
		}

		val, found, err := lookup(ctx, key, tagOpts, o.loader)
		if err != nil {
			return fmt.Errorf("%s: %w", tf.Name, err)
		}

		// If the field already has a non-zero value and there was no value directly
		// specified, do not overwrite the existing field. We only want to overwrite
		// when the envvar was provided directly.
		if (pointerWasSet || !ef.IsZero()) && !found {
			continue
		}

		// Set value.
		if err := setFieldValue(val, ef, tagOpts); err != nil {
			return fmt.Errorf("%s(%q): %w", tf.Name, val, err)
		}
	}

	return nil
}

func parseTag(tag string) (string, *tagOptions, error) {
	if tag == "" {
		return "", nil, nil
	}

	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return "", nil, nil
	}

	opts := &tagOptions{
		prefix:    "",
		delimiter: defaultDelimiter,
		separator: defaultSeparator,
	}

	key := parts[0]
	tagOpts := parts[1:]

	for _, opt := range tagOpts {
		o := strings.TrimLeftFunc(opt, unicode.IsSpace)

		switch {
		case o == optRequired:
			opts.required = true
		case strings.HasPrefix(o, optPrefix):
			opts.prefix = strings.TrimPrefix(o, optPrefix)
		case strings.HasPrefix(o, optDelimiter):
			opts.delimiter = strings.TrimPrefix(o, optDelimiter)
		case strings.HasPrefix(o, optSeparator):
			opts.separator = strings.TrimPrefix(o, optSeparator)
		default:
			return "", opts, fmt.Errorf("xconfig: unknown tag option %q", o)
		}
	}

	return key, opts, nil
}

func lookup(ctx context.Context, key string, opts *tagOptions, l Loader) (string, bool, error) {
	if key == "" {
		// The struct has something like `config:",required"`, which is likely a
		// mistake. We could try to infer the config var from the field name, but that
		// feels too magical.
		return "", false, ErrMissingKey
	}

	// Load value.
	val, found := l.Load(ctx, key)
	if !found {
		if opts.required {
			return "", false, fmt.Errorf("%w: %s", ErrMissingRequired, key)
		}
	}

	return val, found, nil
}
