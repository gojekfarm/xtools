Examples of how to use the xkafka package to publish and consume messages from a Kafka broker. Simulates different Kafka consumer use cases.

## Running Kafka

Start a Kafka broker using the provided `docker-compose.yml` file:

```bash
$ docker-compose up -d
```

## Scenarios

### Basic Consumer

The basic consumer reads messages from a Kafka topic and prints them to the console. It simulates a process crash by restarting the consumer after a random number of messages have been consumed.

```bash
go run *.go basic --partitions=2 --consumers=2 --messages=10
```

### Async Consumer

The async consumer reads messages concurrently using a configurable pool of goroutines.

```bash
go run *.go basic --partitions=2 --consumers=2 --messages=10 --concurrency=4
```

### Batch Consumer

The batch consumer reads messages in batches of a configurable size.

```bash
go run *.go batch --partitions=2 --consumers=2 --messages=20 --batch-size=3
```

### Async Batch Consumer

The async batch consumer processes batches concurrently using a configurable pool of goroutines.

```bash
go run *.go batch --partitions=2 --consumers=2 --messages=20 --batch-size=3 --concurrency=4
```
