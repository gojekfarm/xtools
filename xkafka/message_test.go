package xkafka

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

var topic = "test-topic"

func TestNewMessage(t *testing.T) {
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
		},
		Value:     []byte("value"),
		Key:       []byte("key"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		Headers: []kafka.Header{
			{
				Key:   "x-header",
				Value: []byte("header-value"),
			},
		},
	}

	expectMsg := &Message{
		Group:     "consumer-group-1",
		Topic:     topic,
		Partition: 1,
		Key:       []byte("key"),
		Value:     []byte("value"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		headers: map[string][]byte{
			"x-header": []byte("header-value"),
		},
	}

	m := NewMessage("consumer-group-1", km)
	assert.Equal(t, expectMsg.Group, m.Group)
	assert.Equal(t, expectMsg.Topic, m.Topic)
	assert.Equal(t, expectMsg.Partition, m.Partition)
	assert.Equal(t, expectMsg.Key, m.Key)
	assert.Equal(t, expectMsg.Value, m.Value)
	assert.Equal(t, expectMsg.Timestamp, m.Timestamp)
	assert.Equal(t, expectMsg.headers, m.headers)
}

func TestMessageAsKafkaMessage(t *testing.T) {
	expectKafka := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
		},
		Value:     []byte("value"),
		Key:       []byte("key"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		Headers: []kafka.Header{
			{
				Key:   "x-header",
				Value: []byte("header-value"),
			},
		},
	}

	msg := &Message{
		Group:     "consumer-group-1",
		Topic:     topic,
		Partition: 1,
		Key:       []byte("key"),
		Value:     []byte("value"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		headers: map[string][]byte{
			"x-header": []byte("header-value"),
		},
	}

	kmsg := msg.AsKafkaMessage()
	assert.EqualValues(t, expectKafka.TopicPartition, kmsg.TopicPartition)
	assert.Equal(t, expectKafka.Value, kmsg.Value)
	assert.Equal(t, expectKafka.Key, kmsg.Key)
	assert.Equal(t, expectKafka.Timestamp, kmsg.Timestamp)
	assert.Equal(t, expectKafka.Headers, kmsg.Headers)
	assert.NotNil(t, kmsg.Opaque)
}

func TestAck(t *testing.T) {
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
		},
		Value:     []byte("value"),
		Key:       []byte("key"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	t.Run("FirstAckSuccess", func(t *testing.T) {
		m := NewMessage("consumer-1", km)
		res := m.AckSuccess()

		assert.True(t, res)
		assert.Equal(t, Success, m.Status)
	})

	t.Run("FirstAckSkip", func(t *testing.T) {
		m := NewMessage("consumer-1", km)
		res := m.AckSkip()

		assert.True(t, res)
		assert.Equal(t, Skip, m.Status)
	})

	t.Run("FirstAckFail", func(t *testing.T) {
		m := NewMessage("consumer-1", km)
		err := errors.New("some-err")
		res := m.AckFail(err)

		assert.True(t, res)
		assert.Equal(t, Fail, m.Status)
		assert.EqualError(t, m.Err(), "some-err")
	})

	t.Run("SuccessAfterFailure", func(t *testing.T) {
		m := NewMessage("consumer-1", km)

		m.AckFail(errors.New("error"))
		res := m.AckSuccess()

		assert.True(t, res)
		assert.Equal(t, Success, m.Status)
	})

	t.Run("FailureAfterSuccess", func(t *testing.T) {
		m := NewMessage("consumer-1", km)
		m.AckSuccess()

		res := m.AckFail(errors.New("new-found-error"))

		assert.True(t, res)
		assert.Equal(t, Fail, m.Status)
		assert.EqualError(t, m.Err(), "new-found-error")
	})

	t.Run("SkipAfterSuccess", func(t *testing.T) {
		m := NewMessage("consumer-1", km)
		m.AckSuccess()

		res := m.AckSkip()

		assert.True(t, res)
	})
}

func TestCallbacks(t *testing.T) {
	t.Run("OnAckSuccess", func(t *testing.T) {
		var wg sync.WaitGroup
		ack1Called := false
		ack2Called := false

		wg.Add(2)

		onAck1 := func(m *Message) {
			ack1Called = true
			assert.EqualValues(t, Success, m.Status)
			wg.Done()
		}

		onAck2 := func(m *Message) {
			ack2Called = true
			assert.EqualValues(t, Success, m.Status)
			wg.Done()
		}

		m := Message{}
		m.AddCallback(onAck1)
		m.AddCallback(onAck2)

		go m.AckSuccess()

		wg.Wait()

		assert.True(t, ack1Called)
		assert.True(t, ack2Called)
	})

	t.Run("OnAckSkip", func(t *testing.T) {
		var wg sync.WaitGroup
		ack1Called := false
		ack2Called := false

		wg.Add(2)

		onAck1 := func(m *Message) {
			ack1Called = true
			assert.EqualValues(t, Skip, m.Status)
			wg.Done()
		}

		onAck2 := func(m *Message) {
			ack2Called = true
			assert.EqualValues(t, Skip, m.Status)
			wg.Done()
		}

		m := Message{}
		m.AddCallback(onAck1)
		m.AddCallback(onAck2)

		go m.AckSkip()

		wg.Wait()

		assert.True(t, ack1Called)
		assert.True(t, ack2Called)
	})

	t.Run("OnAckFail", func(t *testing.T) {
		var wg sync.WaitGroup
		ack1Called := false
		ack2Called := false
		err := errors.New("some error")

		wg.Add(2)

		onAck1 := func(m *Message) {
			ack1Called = true
			assert.EqualValues(t, Fail, m.Status)
			wg.Done()
		}

		onAck2 := func(m *Message) {
			ack2Called = true
			assert.EqualValues(t, Fail, m.Status)
			wg.Done()
		}

		m := Message{}
		m.AddCallback(onAck1)
		m.AddCallback(onAck2)

		go m.AckFail(err)

		wg.Wait()

		assert.True(t, ack1Called)
		assert.True(t, ack2Called)
	})
}

func TestHeaders(t *testing.T) {
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
		},
		Value:     []byte("value"),
		Key:       []byte("key"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
	}

	m := NewMessage("consumer-group-1", km)

	m.SetHeader("foo", []byte("bar"))
	assert.Equal(t, m.Headers(), map[string][]byte{"foo": []byte("bar")})
	assert.Equal(t, m.Header("foo"), []byte("bar"))
}
