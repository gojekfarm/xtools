---
"xkafka": patch
---

Treat `kafka.ErrNoOffset` ("Local: No offset stored") from a manual `Commit()` as a benign no-op instead of a fatal error, for both `Consumer` and `BatchConsumer`. With `Concurrency > 1` and `ManualCommit(true)`, the poll loop can service a rebalance (`Unassign` clears the stored offsets) on another goroutine between `StoreOffsets` and `Commit`, leaving the store empty; the partition's new owner resumes from the last committed offset, so there is no data loss and the consumer should not stop.
