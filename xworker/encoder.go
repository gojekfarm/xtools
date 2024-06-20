package xworker

import (
	"encoding/json"
	"io"
)

// PayloadEncoderFunc is used to create an PayloadEncoder from io.Writer.
type PayloadEncoderFunc func(io.Writer) PayloadEncoder

// PayloadEncoder helps in transforming job payload to bytes.
type PayloadEncoder interface {
	// Encode takes any object and encodes it into bytes.
	Encode(v interface{}) error
}

// DefaultPayloadEncoderFunc uses a json.NewEncoder to encode job payload.
func DefaultPayloadEncoderFunc(w io.Writer) PayloadEncoder {
	return json.NewEncoder(w)
}
