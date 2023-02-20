package xkafka_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
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
	}

	m := xkafka.NewMessage("consumer-group-1", km)

	assert.Equal(t, topic, m.Topic)
	assert.EqualValues(t, 1, m.Partition)
	assert.Equal(t, "consumer-group-1", m.Group)
	assert.Equal(t, []byte("key"), m.Key)
	assert.Equal(t, []byte("value"), m.Value)
	assert.Equal(t, xkafka.Unassigned, m.Status)
	assert.EqualValues(t, time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC), m.Timestamp)
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
		m := xkafka.NewMessage("consumer-1", km)
		res := m.AckSuccess()

		assert.True(t, res)
		assert.Equal(t, xkafka.Success, m.Status)
	})

	t.Run("FirstAckSkip", func(t *testing.T) {
		m := xkafka.NewMessage("consumer-1", km)
		res := m.AckSkip()

		assert.True(t, res)
		assert.Equal(t, xkafka.Skip, m.Status)
	})

	t.Run("FirstAckFail", func(t *testing.T) {
		m := xkafka.NewMessage("consumer-1", km)
		err := errors.New("some-err")
		res := m.AckFail(err)

		assert.True(t, res)
		assert.Equal(t, xkafka.Fail, m.Status)
		assert.EqualError(t, m.Err(), "some-err")
	})

	t.Run("SuccessAfterFailure", func(t *testing.T) {
		m := xkafka.NewMessage("consumer-1", km)

		m.AckFail(errors.New("error"))
		res := m.AckSuccess()

		assert.True(t, res)
		assert.Equal(t, xkafka.Success, m.Status)
	})

	t.Run("FailureAfterSuccess", func(t *testing.T) {
		m := xkafka.NewMessage("consumer-1", km)
		m.AckSuccess()

		res := m.AckFail(errors.New("new-found-error"))

		assert.True(t, res)
		assert.Equal(t, xkafka.Fail, m.Status)
		assert.EqualError(t, m.Err(), "new-found-error")
	})

	t.Run("SkipAfterSuccess", func(t *testing.T) {
		m := xkafka.NewMessage("consumer-1", km)
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

		onAck1 := func(m *xkafka.Message) {
			ack1Called = true
			assert.EqualValues(t, xkafka.Success, m.Status)
			wg.Done()
		}

		onAck2 := func(m *xkafka.Message) {
			ack2Called = true
			assert.EqualValues(t, xkafka.Success, m.Status)
			wg.Done()
		}

		m := xkafka.Message{}
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

		onAck1 := func(m *xkafka.Message) {
			ack1Called = true
			assert.EqualValues(t, xkafka.Skip, m.Status)
			wg.Done()
		}

		onAck2 := func(m *xkafka.Message) {
			ack2Called = true
			assert.EqualValues(t, xkafka.Skip, m.Status)
			wg.Done()
		}

		m := xkafka.Message{}
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

		onAck1 := func(m *xkafka.Message) {
			ack1Called = true
			assert.EqualValues(t, xkafka.Fail, m.Status)
			wg.Done()
		}

		onAck2 := func(m *xkafka.Message) {
			ack2Called = true
			assert.EqualValues(t, xkafka.Fail, m.Status)
			wg.Done()
		}

		m := xkafka.Message{}
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

	m := xkafka.NewMessage("consumer-group-1", km)

	m.SetHeader("foo", []byte("bar"))
	assert.Equal(t, m.Headers(), map[string][]byte{"foo": []byte("bar")})
	assert.Equal(t, m.Header("foo"), []byte("bar"))
}
