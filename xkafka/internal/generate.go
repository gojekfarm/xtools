//go:generate mockery --name ConsumerClient --structname MockConsumerClient --filename consumer_mock_test.go --outpkg xkafka --output ../
//go:generate mockery --name ProducerClient --structname MockProducerClient --filename producer_mock_test.go --outpkg xkafka --output ../

package internal
