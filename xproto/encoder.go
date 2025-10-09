package xproto

import (
	"errors"
	"io"

	"google.golang.org/protobuf/proto"
)

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// An Encoder writes proto values to an output stream.
type Encoder struct {
	w io.Writer
}

// Encode writes the proto encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return errors.New("value should be a proto.Message")
	}

	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	if _, err := e.w.Write(b); err != nil {
		return err
	}

	return nil
}
