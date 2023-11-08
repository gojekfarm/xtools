package main

import (
	"sync"

	"github.com/gojekfarm/xtools/xkafka"
)

type state struct {
	generated []*xkafka.Message
	received  map[string]*xkafka.Message
	mu        sync.Mutex
	consumers []*xkafka.Consumer
}

var s = &state{
	received: make(map[string]*xkafka.Message),
}
