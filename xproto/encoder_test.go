package xproto

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
)

func TestEncoder_Encode(t *testing.T) {
	anyPb, _ := anypb.New(&TestMessage_Tuple{})
	validMsg := &TestMessage{
		KeyValues: []*TestMessage_Tuple{{Key: "KEY1", Value: "VALUE1"}},
		Decimal:   3.145145,
		Long:      math.MaxInt64,
		Str:       "string to test",
		Valid:     true,
		Payload:   []byte("some_random_bytes"),
		Objects:   []*anypb.Any{anyPb},
	}

	type fields struct {
		w io.Writer
	}
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Success",
			fields: fields{w: &bytes.Buffer{}},
			args:   args{v: validMsg},
		},
		{
			name:    "NonProtoMessage",
			fields:  fields{w: &bytes.Buffer{}},
			args:    args{v: 1},
			wantErr: true,
		},
		{
			name:    "BadWriter",
			fields:  fields{w: &badWriter{}},
			args:    args{v: validMsg},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewEncoder(tt.fields.w).Encode(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type badWriter struct {
}

func (b *badWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func BenchmarkEncoder(b *testing.B) {
	anyPb, _ := anypb.New(&TestMessage_Tuple{})
	validMsg := &TestMessage{
		KeyValues: []*TestMessage_Tuple{{Key: "KEY1", Value: "VALUE1"}},
		Decimal:   3.145145,
		Long:      math.MaxInt64,
		Str:       "string to test",
		Valid:     true,
		Payload:   []byte("some_random_bytes"),
		Objects:   []*anypb.Any{anyPb},
	}

	buf := bytes.Buffer{}

	for i := 0; i < b.N; i++ {
		buf.Reset()

		if err := NewEncoder(&buf).Encode(validMsg); err != nil {
			b.Errorf("encode error: %s", err)
		}
	}
}
