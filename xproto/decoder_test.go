package xproto

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestDecoder_Decode(t *testing.T) {
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

	d, err := proto.Marshal(validMsg)
	if err != nil {
		t.Errorf("marshal error: %s", err)
	}

	bd, err := json.Marshal(validMsg)
	if err != nil {
		t.Errorf("marshal error: %s", err)
	}

	type fields struct {
		r io.Reader
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
			fields: fields{r: bytes.NewReader(d)},
			args:   args{v: new(TestMessage)},
		},
		{
			name:    "InvalidProto",
			fields:  fields{r: bytes.NewReader(d)},
			args:    args{v: 1},
			wantErr: true,
		},
		{
			name:    "InvalidBytes",
			fields:  fields{r: bytes.NewReader(bd)},
			args:    args{v: new(TestMessage)},
			wantErr: true,
		},
		{
			name:    "BadReader",
			fields:  fields{r: &badReader{}},
			args:    args{v: new(TestMessage)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewDecoder(tt.fields.r).Decode(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type badReader struct {
}

func (b *badReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func BenchmarkDecoder(b *testing.B) {
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

	d, err := proto.Marshal(validMsg)
	if err != nil {
		b.Errorf("marshal error: %s", err)
	}

	r := bytes.NewReader(d)

	for i := 0; i < b.N; i++ {
		var decMsg TestMessage
		if err := NewDecoder(r).Decode(&decMsg); err != nil {
			b.Errorf("decode error: %s", err)
		}
	}
}
