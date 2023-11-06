package main

import (
	"github.com/gojekfarm/xtools/xkafka"
)

type state struct {
	generated []*xkafka.Message
	received  []*xkafka.Message

	consumers []*xkafka.Consumer
}

var s = &state{}
