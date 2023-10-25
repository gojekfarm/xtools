# End to End Testing for xkafka

This package contains end to end tests for xkafka. The tests are run against a real Kafka cluster and are not run in CI.

The tests simulate various scenarios to verify the correctness of the consumer and producer.

## Running the tests

Each scenario is defined in a separate file. To run a scenario, use the following command:

```bash
go run ./e2e/xkafka/*.go <scenario-name> [flags]
```

## Scenarios

### Sequential Consumer

This is the default behavior of the consumer. The consumer will process messages sequentially and commit the offset based on the `auto.commit.interval.ms` configuration.

```bash
go run ./e2e/xkafka/*.go sequential
```

### Sequential Consumer with Manual Commit

### Async Consumer

### Async Consumer with Retry Middleware
