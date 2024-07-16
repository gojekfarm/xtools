//go:generate mockery --name consumerClient --structname MockConsumerClient --filename consumer_mock_test.go --outpkg xkafka --output .
//go:generate mockery --name producerClient --structname MockProducerClient --filename producer_mock_test.go --outpkg xkafka --output .

package xkafka
