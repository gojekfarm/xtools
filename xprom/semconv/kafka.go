package semconv

// System names.
const (
	SystemKafka = "kafka"
)

// Metric names.
// https://github.com/open-telemetry/semantic-conventions/blob/v1.26.0/docs/messaging/messaging-metrics.md
const (
	MessagingClientOperationDuration = "messaging_client_operation_duration"
	MessagingClientPublishedMessages = "messaging_client_published_messages"
	MessagingClientConsumedMessages  = "messaging_client_consumed_messages"
)

// Custom metrics.
const (
	MessagingInflightMessages = "messaging_inflight_messages"
)

// Labels.
// https://github.com/open-telemetry/semantic-conventions/blob/v1.26.0/docs/messaging/messaging-metrics.md
const (
	MessagingSystem                 = "messaging_system"
	MessagingOperationName          = "messaging_operation_name"
	MessagingDestinationName        = "messaging_destination_name"
	MessagingConsumerGroupName      = "messaging_consumer_group_name"
	MessagingDestinationPartitionID = "messaging_destination_partition_id"
)

// Kafka aliases.
const (
	MessagingKafkaTopic         = MessagingDestinationName
	MessagingKafkaPartition     = MessagingDestinationPartitionID
	MessagingKafkaConsumerGroup = MessagingConsumerGroupName
	MessagingKafkaMessageStatus = "messaging_kafka_message_status"
)

// Operation names.
const (
	OperationPublish = "publish"
	OperationConsume = "consume"
)
