package xproto

import (
	"bytes"
	"errors"
	"io"
	"sync"

	"google.golang.org/protobuf/proto"
)

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// A Decoder reads and decodes proto values from an input stream.
type Decoder struct {
	r io.Reader
}

// Decode reads the proto-encoded value from its
// input and stores it in the value pointed to by v.
func (d *Decoder) Decode(v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return errors.New("value should be a proto.Message")
	}

	buf := bufPool.Get().(*bytes.Buffer) //nolint:errcheck
	defer bufPool.Put(buf)

	buf.Reset()

	if _, err := buf.ReadFrom(d.r); err != nil {
		return err
	}

	if err := proto.Unmarshal(buf.Bytes(), m); err != nil {
		return err
	}

	return nil
}
