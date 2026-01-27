package xkafka

import (
	"errors"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
)

func TestNewBatch(t *testing.T) {
	batch := NewBatch()

	assert.NotEmpty(t, batch.ID)
	assert.Empty(t, batch.Messages)
	assert.Zero(t, batch.Status)
}

func TestBatch_AckSuccess(t *testing.T) {
	batch := NewBatch()
	batch.AckSuccess()

	assert.Equal(t, Success, batch.Status)
}

func TestBatch_AckFail(t *testing.T) {
	batch := NewBatch()
	testErr := errors.New("test error")

	err := batch.AckFail(testErr)

	assert.Equal(t, Fail, batch.Status)
	assert.Equal(t, testErr, batch.Err())
	assert.Equal(t, testErr, err)
}

func TestBatch_AckSkip(t *testing.T) {
	batch := NewBatch()
	batch.AckSkip()

	assert.Equal(t, Skip, batch.Status)
}

func TestBatch_OffsetMethods(t *testing.T) {
	tests := []struct {
		name            string
		messages        []*Message
		wantMaxOffset   int64
		wantGroupOffset []kafka.TopicPartition
	}{
		{
			name:            "empty batch",
			messages:        []*Message{},
			wantMaxOffset:   0,
			wantGroupOffset: []kafka.TopicPartition{},
		},
		{
			name: "single topic-partition",
			messages: []*Message{
				{Topic: "topic1", Partition: 0, Offset: 5},
				{Topic: "topic1", Partition: 0, Offset: 10},
			},
			wantMaxOffset: 10,
			wantGroupOffset: []kafka.TopicPartition{
				{
					Topic:     strPtr("topic1"),
					Partition: 0,
					Offset:    kafka.Offset(10),
				},
			},
		},
		{
			name: "multiple topic-partitions",
			messages: []*Message{
				{Topic: "topic1", Partition: 0, Offset: 5},
				{Topic: "topic1", Partition: 1, Offset: 10},
				{Topic: "topic2", Partition: 0, Offset: 15},
			},
			wantMaxOffset: 15,
			wantGroupOffset: []kafka.TopicPartition{
				{
					Topic:     strPtr("topic1"),
					Partition: 0,
					Offset:    kafka.Offset(5),
				},
				{
					Topic:     strPtr("topic1"),
					Partition: 1,
					Offset:    kafka.Offset(10),
				},
				{
					Topic:     strPtr("topic2"),
					Partition: 0,
					Offset:    kafka.Offset(15),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := NewBatch()
			batch.Messages = tt.messages

			t.Run("MaxOffset", func(t *testing.T) {
				gotMaxOffset := batch.MaxOffset()
				assert.Equal(t, tt.wantMaxOffset, gotMaxOffset)
			})

			t.Run("GroupMaxOffset", func(t *testing.T) {
				gotGroupOffset := batch.GroupMaxOffset()
				assert.ElementsMatch(t, tt.wantGroupOffset, gotGroupOffset)
			})
		})
	}
}

func strPtr(s string) *string {
	return &s
}
