package xkafka

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/xtools/xkafka/internal"
)

type ConsumerSuite struct {
	suite.Suite
	kafka    *MockConsumerClient
	consumer *Consumer
	topics   []string
	brokers  []string
	messages []*kafka.Message
}

func TestConsumerSuite(t *testing.T) {
	suite.Run(t, &ConsumerSuite{})
}

func (s *ConsumerSuite) SetupTest() {
	s.kafka = &MockConsumerClient{}
	s.topics = []string{topic}
	s.brokers = []string{"localhost:9092"}

	consumer, err := NewConsumer(
		"consumer-id",
		Topics(s.topics),
		Brokers(s.brokers),
		mockConsumerFunc(s.kafka),
		MetadataTimeout(10*time.Second),
		PollTimeout(1*time.Second),
	)
	s.Require().NoError(err)
	s.Require().NotNil(consumer)

	s.consumer = consumer

	s.generateMessages()
}

func (s *ConsumerSuite) TestGetMetadata() {
	s.kafka.On("GetMetadata", (*string)(nil), false, 10000).Return(&kafka.Metadata{}, nil)

	metadata, err := s.consumer.GetMetadata()
	s.NoError(err)
	s.NotNil(metadata)

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestNoHandler() {
	ctx := context.Background()

	err := s.consumer.Start(ctx)
	s.Error(err)
	s.EqualError(err, ErrNoHandler)

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestHandleMessage() {
	km := s.messages[0]
	ctx, cancel := context.WithCancel(context.Background())
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		s.Equal(km.Key, msg.Key)
		s.Equal(km.Value, msg.Value)

		cancel()
		return nil
	})

	s.kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	s.kafka.On("ReadMessage", 1*time.Second).Return(km, nil)

	err := s.consumer.WithHandler(handler).Start(ctx)
	s.NoError(err)

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestHandleMessageWithErrors() {
	km := s.messages[0]
	ctx := context.Background()
	expect := errors.New("error in handler")

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		return expect
	})

	s.kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	s.kafka.On("ReadMessage", 1*time.Second).Return(km, nil)

	err := s.consumer.WithHandler(handler).Start(ctx)
	s.Error(err)
	s.EqualError(err, expect.Error())

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestKafkaReadTimeout() {
	km := s.messages[0]
	ctx, cancel := context.WithCancel(context.Background())
	counter := 0

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		counter++

		if counter > 1 {
			cancel()
		}

		return nil
	})
	expect := kafka.NewError(kafka.ErrTimedOut, "kafka: timed out", false)

	s.kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	s.kafka.On("ReadMessage", 1*time.Second).Return(km, nil).Once()
	s.kafka.On("ReadMessage", 1*time.Second).Return(nil, expect).Once()
	s.kafka.On("ReadMessage", 1*time.Second).Return(km, nil)

	err := s.consumer.WithHandler(handler).Start(ctx)
	s.NoError(err)

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestKafkaError() {
	ctx := context.Background()
	expect := kafka.NewError(kafka.ErrUnknown, "kafka: unknown error", false)

	s.kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	s.kafka.On("ReadMessage", 1*time.Second).Return(nil, expect).Once()

	err := s.consumer.WithHandler(noopHandler()).Start(ctx)
	s.Error(err)
	s.EqualError(err, expect.Error())

	s.kafka.AssertExpectations(s.T())
}

func (s *ConsumerSuite) TestMiddlewareExecutionOrder() {
	km := s.messages[0]
	ctx, cancel := context.WithCancel(context.Background())
	preExec := []int{}
	postExec := []int{}

	m1 := MiddlewareFunc(func(handler Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *Message) error {
			preExec = append(preExec, 1)

			err := handler.Handle(ctx, msg)

			postExec = append(postExec, 1)

			return err
		})
	})

	m2 := MiddlewareFunc(func(handler Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *Message) error {
			preExec = append(preExec, 2)

			err := handler.Handle(ctx, msg)

			postExec = append(postExec, 2)

			return err
		})
	})

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		cancel()
		return nil
	})

	s.kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	s.kafka.On("ReadMessage", 1*time.Second).Return(km, nil)

	s.consumer.Use(m1, m2)

	err := s.consumer.WithHandler(handler).Start(ctx)
	s.NoError(err)

	s.kafka.AssertExpectations(s.T())
	s.Equal(preExec, []int{1, 2})
	s.Equal(postExec, []int{2, 1})
}

func (s *ConsumerSuite) generateMessages() {
	for i := 0; i < 10; i++ {
		km := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: 1,
			},
			Key:       []byte(fmt.Sprintf("key-%d", i)),
			Value:     []byte(fmt.Sprintf("value-%d", i)),
			Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
		}

		s.messages = append(s.messages, km)
	}
}

func mockConsumerFunc(mock *MockConsumerClient) ConsumerFunc {
	return func(configMap *kafka.ConfigMap) (internal.ConsumerClient, error) {
		return mock, nil
	}
}

func noopHandler() Handler {
	return HandlerFunc(func(ctx context.Context, msg *Message) error {
		return nil
	})
}
