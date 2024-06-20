package xworker

import (
	"bytes"
	"io"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_Read(t *testing.T) {
	type obj struct {
		FieldOne string `json:"field_one"`
		FieldTwo int    `json:"field_two"`
	}

	tests := []struct {
		name    string
		payload interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Success",
			payload: obj{
				FieldOne: "value-one",
				FieldTwo: 1,
			},
			want: `{"field_one":"value-one","field_two":1}
`,
		},
		{
			name:    "EncodeError",
			payload: math.Inf(1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := Job{
				Payload:     tt.payload,
				encoderFunc: DefaultPayloadEncoderFunc,
			}

			var buf bytes.Buffer
			_, err := buf.ReadFrom(&j)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestJob_Read_LargeValue(t *testing.T) {
	type testPayload struct {
		Int    int    `json:"int"`
		String string `json:"string"`
		Bytes  []byte `json:"bytes"`
	}

	randomBytes := make([]byte, 1000)
	rand.Read(randomBytes)

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	randSeq := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		return string(b)
	}

	tp := &testPayload{
		Int:    100000,
		String: randSeq(100),
		Bytes:  randomBytes,
	}
	j := &Job{
		Payload:     tp,
		encoderFunc: DefaultPayloadEncoderFunc,
	}

	var buf bytes.Buffer
	n, err := buf.ReadFrom(j)
	assert.NoError(t, err)

	assert.Equal(t, 1474, int(n))
	assert.Equal(t, int(n), buf.Len())
}

func TestJob_Write(t *testing.T) {
	j := &Job{}
	n, err := j.Write([]byte(`{"field_one":"value-one","field_two":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 39, n)
	assert.Equal(t, `{"field_one":"value-one","field_two":1}`, j.buf.String())
}

func TestJob_DecodePayload(t *testing.T) {
	type obj struct {
		FieldOne string `json:"field_one"`
		FieldTwo int    `json:"field_two"`
	}

	j := &Job{decoderFunc: DefaultPayloadDecoderFunc}
	n, err := j.Write([]byte(`{"field_one":"value-one","field_two":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 39, n)

	var got obj
	assert.NoError(t, j.DecodePayload(&got))

	assert.Equal(t, obj{
		FieldOne: "value-one",
		FieldTwo: 1,
	}, got)
}

func TestJob_DecodePayload_MultipleTimes(t *testing.T) {
	type obj struct {
		FieldOne string `json:"field_one"`
		FieldTwo int    `json:"field_two"`
	}

	j := &Job{decoderFunc: DefaultPayloadDecoderFunc}
	n, err := j.Write([]byte(`{"field_one":"value-one","field_two":1}`))
	assert.NoError(t, err)
	assert.Equal(t, 39, n)

	var got1 obj
	assert.NoError(t, j.DecodePayload(&got1))

	assert.Equal(t, obj{
		FieldOne: "value-one",
		FieldTwo: 1,
	}, got1)

	var got2 obj
	assert.NoError(t, j.DecodePayload(&got2))

	assert.Equal(t, obj{
		FieldOne: "value-one",
		FieldTwo: 1,
	}, got2)
}

func TestNewJobWithDecoder(t *testing.T) {
	f := func(r io.Reader) PayloadDecoder { return nil }
	assert.NotNil(t, NewJobWithDecoder(f).decoderFunc)
}
