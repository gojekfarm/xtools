package xconfig

import (
	"encoding"
	"encoding/gob"
	"encoding/json"
	"reflect"
)

//nolint:nestif
func decode(v string, ef reflect.Value) (imp bool, err error) {
	for ef.CanAddr() {
		ef = ef.Addr()
	}

	if ef.CanInterface() {
		iface := ef.Interface()

		// Custom decoder supersedes all other decoders
		if dec, ok := iface.(Decoder); ok {
			imp = true
			err = dec.Decode(v)

			return imp, err
		}

		if tu, ok := iface.(encoding.TextUnmarshaler); ok {
			imp = true
			if err = tu.UnmarshalText([]byte(v)); err == nil {
				return imp, nil
			}
		}

		if tu, ok := iface.(json.Unmarshaler); ok {
			imp = true
			if err = tu.UnmarshalJSON([]byte(v)); err == nil {
				return imp, nil
			}
		}

		if tu, ok := iface.(encoding.BinaryUnmarshaler); ok {
			imp = true
			if err = tu.UnmarshalBinary([]byte(v)); err == nil {
				return imp, nil
			}
		}

		if tu, ok := iface.(gob.GobDecoder); ok {
			imp = true
			if err = tu.GobDecode([]byte(v)); err == nil {
				return imp, nil
			}
		}
	}

	return imp, err
}
