package xkafka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gojekfarm/xtools/xkafka"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProducerSuite struct {
	suite.Suite
	kafka    *MockKafkaProducer
	producer *xkafka.Producer
	topic    string
	messages []*xkafka.Message
}

func TestProducerSuite(t *testing.T) {
	suite.Run(t, &ProducerSuite{})
}

func (s *ProducerSuite) SetupTest() {
	s.kafka = &MockKafkaProducer{}
	s.kafka.On("Events").Return(make(chan kafka.Event, 0))

	producer, err := xkafka.NewProducer(
		xkafka.Brokers([]string{"localhost:9092"}),
		mockProducerFunc(s.kafka),
		xkafka.ShutdownTimeout(1*time.Second),
	)
	s.Require().NoError(err)
	s.Require().NotNil(producer)

	s.producer = producer
	s.topic = topic
	s.generateMessages()
}

func (s *ProducerSuite) TestPublish() {
	msg := s.messages[0]
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	s.kafka.On("Events").Return(make(chan kafka.Event, 0))
	s.kafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			args.Get(1).(chan kafka.Event) <- km
		}()
	}).Return(nil)

	err := s.producer.Publish(context.Background(), msg)
	s.NoError(err)

	s.kafka.AssertExpectations(s.T())
}

func (s *ProducerSuite) TestPublishError() {
	msg := s.messages[0]
	expectErr := fmt.Errorf("kafka error")

	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	s.kafka.On("Events").Return(make(chan kafka.Event, 0))
	s.kafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			kmWithError := *km
			kmWithError.TopicPartition.Error = expectErr

			args.Get(1).(chan kafka.Event) <- &kmWithError
		}()
	}).Return(nil)

	err := s.producer.Publish(context.Background(), msg)
	s.Error(err)
	s.EqualError(err, expectErr.Error())

	s.kafka.AssertExpectations(s.T())
}

func (s *ProducerSuite) TestClose() {
	ctx := context.Background()
	s.kafka.On("Flush", 1000).Return(0)
	s.kafka.On("Close").Return()

	s.producer.Close(ctx)

	s.kafka.AssertExpectations(s.T())
}

func (s *ProducerSuite) generateMessages() {
	s.messages = make([]*xkafka.Message, 10)
	for i := 0; i < 10; i++ {
		s.messages[i] = &xkafka.Message{
			Topic: s.topic,
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("value-%d", i)),
		}
	}
}

func mockProducerFunc(mock *MockKafkaProducer) xkafka.ProducerFunc {
	return func(configMap *kafka.ConfigMap) (xkafka.KafkaProducer, error) {
		return mock, nil
	}
}
