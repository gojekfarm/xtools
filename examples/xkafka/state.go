package main

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
)

type Tracker struct {
	expect        map[string]*xkafka.Message
	mu            sync.Mutex
	received      map[string]*xkafka.Message
	order         []string
	cancel        func()
	simulateError bool
}

func NewTracker(messages []*xkafka.Message, cancel func()) *Tracker {
	t := &Tracker{
		expect:   make(map[string]*xkafka.Message),
		received: make(map[string]*xkafka.Message),
		order:    make([]string, 0),
		cancel:   cancel,
	}

	for _, m := range messages {
		t.expect[string(m.Key)] = m
	}

	return t
}

func (t *Tracker) Ack(msg *xkafka.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.received[string(msg.Key)] = msg
	t.order = append(t.order, string(msg.Key))
}

func (t *Tracker) SimulateWork() error {
	<-time.After(time.Duration(rand.Int63n(200)) * time.Millisecond)

	t.mu.Lock()
	defer t.mu.Unlock()

	after := len(t.expect) / 3

	// simulate error after 1/3 of messages are received
	if len(t.received) >= after && !t.simulateError {
		t.simulateError = true

		return errors.New("simulated error")
	}

	return nil
}

func (t *Tracker) CancelIfDone() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.received) == len(t.expect) {
		slog.Info("[TRACKER] all messages received, cancelling context")

		t.cancel()
	}
}

func (t *Tracker) Summary() {
	t.mu.Lock()
	defer t.mu.Unlock()

	slog.Info("[TRACKER] summary",
		"received", len(t.received),
		"expected", len(t.expect),
	)

	for _, key := range t.order {
		msg := t.received[key]
		slog.Info("[TRACKER] message",
			"key", string(msg.Key),
			"offset", msg.Offset,
			"partition", msg.Partition,
			"topic", msg.Topic,
		)
	}
}
