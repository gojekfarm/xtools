//go:generate mockery --name KafkaConsumer --structname MockKafkaConsumer --filename consumer_mock_test.go --outpkg xkafka_test --output .
//go:generate mockery --name KafkaProducer --structname MockKafkaProducer --filename producer_mock_test.go --outpkg xkafka_test --output .

package xkafka
