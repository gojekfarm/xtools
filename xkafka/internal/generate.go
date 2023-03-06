//go:generate mockery --name ConsumerClient --structname MockConsumerClient --filename consumer_mock_test.go --outpkg xkafka_test --output ../
//go:generate mockery --name ProducerClient --structname MockProducerClient --filename producer_mock_test.go --outpkg xkafka_test --output ../

package internal
