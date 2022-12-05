package xproto_test

import (
	"bytes"
	"fmt"
	"log"
	"math"

	"github.com/gojekfarm/xtools/xproto"
)

func ExampleNewEncoder() {
	var buf bytes.Buffer

	xproto.NewEncoder(&buf)
}

func ExampleEncoder_Encode() {
	var buf bytes.Buffer

	// xproto.TestMessage is a protobuf message
	tp := &xproto.TestMessage{
		KeyValues: []*xproto.TestMessage_Tuple{{Key: "KEY1", Value: "VALUE1"}},
		Decimal:   3.145145,
		Long:      math.MaxInt64,
		Valid:     true,
		Payload:   []byte("some_random_bytes"),
	}

	if err := xproto.NewEncoder(&buf).Encode(tp); err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf.Bytes())
	// Output: [10 14 10 4 75 69 89 49 18 6 86 65 76 85 69 49 17 97 108 33 200 65 41 9 64 24 255 255 255 255 255 255 255 255 127 40 1 50 17 115 111 109 101 95 114 97 110 100 111 109 95 98 121 116 101 115]
}

func ExampleNewDecoder() {
	var buf bytes.Buffer

	xproto.NewDecoder(&buf)
}

func ExampleDecoder_Decode() {
	var buf bytes.Buffer
	_, _ = buf.Write([]byte{10, 14, 10, 4, 75, 69, 89, 49, 18, 6, 86, 65, 76, 85, 69, 49, 17, 97, 108, 33, 200, 65, 41, 9, 64, 24, 255, 255, 255, 255, 255, 255, 255, 255, 127, 40, 1, 50, 17, 115, 111, 109, 101, 95, 114, 97, 110, 100, 111, 109, 95, 98, 121, 116, 101, 115})

	// xproto.TestMessage is a protobuf message
	var tp xproto.TestMessage

	if err := xproto.NewDecoder(&buf).Decode(&tp); err != nil {
		log.Fatal(err)
	}
}
