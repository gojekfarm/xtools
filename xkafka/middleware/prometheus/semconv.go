package prometheus

// This file contains the semantic conventions for the Prometheus metrics.

// Labels
const (
	LabelTopic     = "topic"
	LabelPartition = "partition"
	LabelGroup     = "group"
	LabelStatus    = "status"
)

// Metric names
const (
	MetricMessageLagSeconds      = "message_lag_seconds"
	MetricMessagesTotal          = "messages_total"
	MetricMessagesInFlight       = "messages_in_flight"
	MetricMessageDurationSeconds = "message_duration_seconds"
)

// Subsystem names
const (
	SubsystemConsumer = "kafka_consumer"
	SubsystemProducer = "kafka_producer"
)
