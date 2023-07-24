package xconfig

import (
	"reflect"
	"strings"
)

const (
	optRequired = "required"
	optPrefix   = "prefix="
)

// Node is an internal representation of a single config value.
type Node struct {
	name     string
	required bool
	prefix   string

	field reflect.Value
}

// SetVal sets the value of the given node.
func (n *Node) SetVal(val string) error {
	if val == "" && n.required {
		return ErrRequired
	}

	return setVal(n.field, val)
}

// parse returns a flattened list of all nodes
// by parsing the tags of the given struct recursively.
// Private fields are ignored.
func parse(s any, o *options) ([]*Node, error) {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return nil, ErrNotPointer
	}

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	t := e.Type()

	nodes := make([]*Node, 0)

	for i := 0; i < t.NumField(); i++ {
		ef := e.Field(i)
		tf := t.Field(i)
		tag := tf.Tag.Get(o.key)

		if !ef.CanSet() {
			continue
		}

		// resolve the pointer recursively
		for ef.Kind() == reflect.Ptr {
			if ef.IsNil() {
				// stop if the underlying value is not a struct
				if ef.Type().Elem().Kind() != reflect.Struct {
					break
				}

				// create a new struct and set it to the pointer
				ef = reflect.New(ef.Type().Elem()).Elem()
			} else {
				ef = ef.Elem()
			}
		}

		node, err := newNodeFromTag(tag)
		if err != nil {
			return nil, err
		}

		node.field = ef

		if ef.Kind() == reflect.Struct {
			for ef.CanAddr() {
				ef = ef.Addr()
			}

			child, err := parse(ef.Interface(), o)
			if err != nil {
				return nil, err
			}

			for _, c := range child {
				c.name = node.prefix + c.name

				nodes = append(nodes, c)
			}
		} else {
			nodes = append(nodes, node)
		}

	}

	return nodes, nil
}

func newNodeFromTag(tag string) (*Node, error) {
	parts := strings.Split(tag, ",")
	key, tagOpts := strings.TrimSpace(parts[0]), parts[1:]

	node := &Node{
		name: key,
	}

	for _, opt := range tagOpts {
		opt = strings.TrimSpace(opt)

		switch {
		case opt == optRequired:
			if key == "" {
				return nil, ErrMissingKey
			}

			node.required = true
		case strings.HasPrefix(opt, optPrefix):
			node.prefix = strings.TrimPrefix(opt, optPrefix)
		default:
			return nil, ErrUnknownTagOption
		}

	}

	return node, nil
}
