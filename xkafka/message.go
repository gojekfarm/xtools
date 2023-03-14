package xkafka

import (
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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

	headers      *sync.Map
	ackCallbacks []AckFunc
	mutex        sync.Mutex
	err          error
}

// NewMessage creates a new message from a kafka message.
func NewMessage(group string, raw *kafka.Message) *Message {
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
	m.headers.Store(key, value)
}

// Headers returns a map to access the key and value of the header field of the message.
func (m *Message) Headers() map[string][]byte {
	res := make(map[string][]byte)

	m.headers.Range(func(key, val interface{}) bool {
		k, ok1 := key.(string)
		v, ok2 := val.([]byte)

		if ok1 && ok2 {
			res[k] = v

			return true
		}

		return false
	})

	return res
}

// Header returns the value for the given key of the header field of the message.
func (m *Message) Header(key string) []byte {
	v, _ := m.headers.Load(key)
	b, _ := v.([]byte)

	return b
}

func mapHeaders(headers []kafka.Header) *sync.Map {
	res := &sync.Map{}

	for _, s := range headers {
		res.Store(s.Key, s.Value)
	}

	return res
}
