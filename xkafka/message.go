package xkafka

import (
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// AckFunc are callback funtions triggered for every ack.
type AckFunc func(m *Message)

// Status is an enum for state of Message.
type Status int

// Status enums.
const (
	Unassigned Status = iota
	Success
	Fail
	Skip
)

// String returns a string status value.
func (s Status) String() string {
	return [...]string{"UNASSIGNED", "SUCCESS", "FAIL", "SKIP"}[s]
}

// Message holds the Kafka message data and manages the
// lifecycle of the message.
type Message struct {
	ID        string
	Topic     string
	Partition int32
	Group     string
	Key       []byte
	Value     []byte
	Timestamp time.Time
	Status    Status
	ErrMsg    string

	headers      map[string][]byte
	ackCallbacks []AckFunc
	mutex        sync.Mutex
	err          error
}

// newMessage creates a new message from a kafka message.
func newMessage(group string, raw *kafka.Message) *Message {
	return &Message{
		Topic:     *raw.TopicPartition.Topic,
		Partition: raw.TopicPartition.Partition,
		Group:     group,
		Key:       raw.Key,
		Value:     raw.Value,
		Timestamp: raw.Timestamp,
		Status:    Unassigned,
		headers:   mapHeaders(raw.Headers),
	}
}

// AddCallback adds the callback func to the call stack.
func (m *Message) AddCallback(fn AckFunc) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ackCallbacks = append(m.ackCallbacks, fn)
}

func (m *Message) triggerCallbacks() {
	for i := len(m.ackCallbacks) - 1; i >= 0; i-- {
		m.ackCallbacks[i](m)
	}
}

// AckSuccess marks the message as successfully processed.
// Overrides any existing ack status.
func (m *Message) AckSuccess() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Status = Success

	m.triggerCallbacks()

	return true
}

// AckSkip marks the message as skipped.
// Overrides any existing ack status.
func (m *Message) AckSkip() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Status = Skip

	m.triggerCallbacks()

	return true
}

// AckFail marks the message as failed out and stores the error.
// Error overrides any existing ack status.
func (m *Message) AckFail(err error) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Status = Fail
	m.ErrMsg = err.Error()
	m.err = err

	m.triggerCallbacks()

	return true
}

// Err returns the underlying error that cause message fail.
// DESIGN: Intentionally called Err to avoid confusion with Error().
func (m *Message) Err() error {
	return m.err
}

// SetHeader stores the key and value of the header field of the message.
func (m *Message) SetHeader(key string, value []byte) {
	m.headers[key] = value
}

// Headers returns a map to access the key and value of the header field of the message.
func (m *Message) Headers() map[string][]byte {
	return m.headers
}

// Header returns the value for the given key of the header field of the message.
func (m *Message) Header(key string) []byte {
	return m.headers[key]
}

// asKafkaMessage returns the message as a kafka.Message.
func (m *Message) asKafkaMessage() *kafka.Message {
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &m.Topic,
			Partition: m.Partition,
		},
		Key:       m.Key,
		Value:     m.Value,
		Timestamp: m.Timestamp,
		Opaque:    m,
	}

	if m.headers != nil {
		km.Headers = make([]kafka.Header, 0, len(m.headers))

		for k, v := range m.headers {
			km.Headers = append(km.Headers, kafka.Header{
				Key:   k,
				Value: v,
			})
		}
	}

	return km
}

func mapHeaders(headers []kafka.Header) map[string][]byte {
	m := make(map[string][]byte, len(headers))

	for _, h := range headers {
		m[h.Key] = h.Value
	}

	return m
}
