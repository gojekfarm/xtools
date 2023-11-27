// Package xkafka provides consumers & producers to work efficiently with Kafka.
// `Message` is the core data structure for xkafka.
// `xkafka` comes with implementations that support:
//   - HTTP-like handler for message processing
//   - Concurrent message consumer
//   - Batch message consumer with retries
//   - Middleware support for consumer & producer
//
// ## Consumer
// The processing mode is determined by the xkafka.Concurrency option. By default, the
// consumer is initialized with `enable.auto.offset.store=false`. The offset is "stored"
// after the message is processed. The offset is "committed" based on the
// `auto.commit.interval.ms` options.
//
// It is important to understand the difference between "store" and "commit". The offset is
// "stored" in the consumer's memory and is "committed" to Kafka. The offset is "stored"
// after the message is processed and the `message.Status` is Success or Skip. The stored
// offsets will be automatically committed, unless the `ManualCommit` option is enabled.
//
// ### Error Handling
// By default, xkafka.Consumer will stop processing, commit last stored offset,
// and exit if there is a Kafka error or if the handler returns an error.
//
// Errors can be handled by using one or more of the following options:
//
// Within the handler implementation
// Using error handling & retry middlewares
// through the catch all xkafka.ErrorHandler option
// xkafka.ErrorHandler is called for every error that is not handled by the handler or
// the middlewares. It is also called for errors returned by underlying Kafka client.
//
// ### Sequential Processing
// Sequential processing is the default mode. It is same as xkafka.Concurrency(1).
//
// ### Async Processing
// Async processing is enabled by setting xkafka.Concurrency to a value greater than 1.
// The consumer will use a pool of Go routines to process messages concurrently.
// Offsets are stored and committed in the order that the messages are received.
//
// ### Manual Commit
// By default, the consumer will automatically commit the offset based on the
// `auto.commit.interval.ms` option, asynchronously in the background.
//
// The consumer can be configured to commit the offset manually by setting
// `xkafka.EnableManualCommit` option to true. When ManualCommit is enabled,
// the consumer will synchronously commit the offset after each message is processed.
//
// NOTE: Enabling ManualCommit will add an overhead to each message. It is
// recommended to use ManualCommit only when necessary.
package xkafka
