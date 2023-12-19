Examples of how to use the xkafka package to read and write messages to a Kafka topic using `xkafka.Consumer` and `xkafka.Producer`.

## Running Kafka

Start a Kafka broker using the provided `docker-compose.yml` file:

```bash
$ docker-compose up -d
```

## Scenarios

### Sequential Consumer

This is the default behavior of the consumer. The consumer will process messages sequentially and commit the offset based on the `auto.commit.interval.ms` configuration.

```bash
go run *.go sequential
```

### Async Consumer

Async mode is enabled by setting `xkafka.Concurrency` to a value greater than 1. The consumer will process messages using a pool of Go routines.

```bash
go run *.go async
```
