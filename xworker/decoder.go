package xworker

import (
	"encoding/json"
	"io"
)

// PayloadDecoderFunc is used to create an PayloadDecoder from io.Reader.
type PayloadDecoderFunc func(io.Reader) PayloadDecoder

// PayloadDecoder helps to decode bytes into the desired object.
type PayloadDecoder interface {
	// Decode decodes message bytes into the passed object
	Decode(v interface{}) error
}

// DefaultPayloadDecoderFunc uses a json.NewDecoder to decode job payload.
func DefaultPayloadDecoderFunc(r io.Reader) PayloadDecoder {
	return json.NewDecoder(r)
}
