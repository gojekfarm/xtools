package gocraft

import "encoding/json"

type jobBlob struct {
	Args struct {
		Payload    []byte     `json:"payload"`
		PayloadRaw rawPayload `json:"payload_raw"`
	} `json:"args"`
}

type rawPayload []byte

var (
	_ json.Unmarshaler = (*rawPayload)(nil)
	_ json.Marshaler   = (*rawPayload)(nil)
)

func (p rawPayload) MarshalJSON() ([]byte, error) {
	return p, nil
}

func (p *rawPayload) UnmarshalJSON(data []byte) error {
	*p = data

	return nil
}

func (j *jobBlob) payload() []byte {
	if len(j.Args.Payload) > 0 {
		return j.Args.Payload
	}

	return j.Args.PayloadRaw
}
