
This example shows how to use the xkafka package to read and write messages to a Kafka topic using `xkafka.Consumer` and `xkafka.Producer`.

## Running xkafka

To run the example, first start a Kafka broker.:

```bash
$ docker run -p 2181:2181 -p 9092:9092 \
    --env ADVERTISED_HOST=kafka \
    --env ADVERTISED_PORT=9092 \
    --env AUTO_CREATE_TOPICS=true \
    spotify/kafka
```

Run the example:

```bash
$ go run main.go
```
