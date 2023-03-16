package kfq

import "github.com/confluentinc/confluent-kafka-go/kafka"

// IsErrorRetriable is a function which returns true if the error is retriable.
type IsErrorRetriable func(err error) bool

func (fn IsErrorRetriable) apply(q *Queue) { q.isErrorRetriable = fn }

func isErrRetriable(err error) bool {
	ke, ok := err.(kafka.Error)
	if !ok {
		return false
	}

	switch ke.Code() {
	// DESIGN: Message timeout is retryable for our usecases.
	case kafka.ErrMsgTimedOut:
		return true
	// DESIGN: These are transient errors from Kafka and can be retried.
	// https://github.com/segmentio/kafka-go/blob/master/error.go#L119
	case kafka.ErrInvalidMsg,
		kafka.ErrUnknownTopicOrPart,
		kafka.ErrLeaderNotAvailable,
		kafka.ErrNotLeaderForPartition,
		kafka.ErrRequestTimedOut,
		kafka.ErrNetworkException,
		kafka.ErrCoordinatorLoadInProgress,
		kafka.ErrCoordinatorNotAvailable,
		kafka.ErrNotCoordinator,
		kafka.ErrNotEnoughReplicas,
		kafka.ErrNotEnoughReplicasAfterAppend,
		kafka.ErrNotController,
		kafka.ErrKafkaStorageError,
		kafka.ErrFetchSessionIDNotFound,
		kafka.ErrInvalidFetchSessionEpoch,
		kafka.ErrListenerNotFound,
		kafka.ErrFencedLeaderEpoch,
		kafka.ErrUnknownLeaderEpoch,
		kafka.ErrOffsetNotAvailable,
		kafka.ErrPreferredLeaderNotAvailable,
		kafka.ErrEligibleLeadersNotAvailable,
		kafka.ErrElectionNotNeeded,
		kafka.ErrNoReassignmentInProgress,
		kafka.ErrGroupSubscribedToTopic,
		kafka.ErrUnstableOffsetCommit:
		return true
	default:
		return false
	}
}
