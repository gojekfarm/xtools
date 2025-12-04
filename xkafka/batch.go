package xkafka

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/xid"
)

// Batch is a group of messages that are processed together.
type Batch struct {
	ID       string
	Messages []*Message
	Status   Status

	err  error
	lock sync.Mutex
}

// NewBatch creates a new Batch.
func NewBatch() *Batch {
	return &Batch{
		ID:       xid.New().String(),
		Messages: make([]*Message, 0),
	}
}

// AckSuccess marks the batch as successfully processed.
func (b *Batch) AckSuccess() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.Status = Success
}

// AckFail marks the batch as failed to process.
func (b *Batch) AckFail(err error) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.Status = Fail
	b.err = err

	return err
}

// AckSkip marks the batch as skipped.
func (b *Batch) AckSkip() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.Status = Skip
}

// Err returns the error that caused the batch to fail.
func (b *Batch) Err() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.err
}

// MaxOffset returns the maximum offset among the
// messages in the batch.
func (b *Batch) MaxOffset() int64 {
	b.lock.Lock()
	defer b.lock.Unlock()

	var max int64
	for _, m := range b.Messages {
		if m.Offset > max {
			max = m.Offset
		}
	}

	return max
}

// GroupMaxOffset returns the maximum offset for each
// topic-partition in the batch.
func (b *Batch) GroupMaxOffset() []kafka.TopicPartition {
	b.lock.Lock()
	defer b.lock.Unlock()

	offsets := make(map[string]map[int32]int64)
	for _, m := range b.Messages {
		if _, ok := offsets[m.Topic]; !ok {
			offsets[m.Topic] = map[int32]int64{
				m.Partition: m.Offset,
			}
		}

		if m.Offset > offsets[m.Topic][m.Partition] {
			offsets[m.Topic][m.Partition] = m.Offset
		}
	}

	var tps []kafka.TopicPartition

	for topic, partitions := range offsets {
		topic := topic

		for partition, offset := range partitions {
			tps = append(tps, kafka.TopicPartition{
				Topic:     &topic,
				Partition: partition,
				Offset:    kafka.Offset(offset),
			})
		}
	}

	return tps
}
