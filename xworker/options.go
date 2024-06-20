package xworker

import (
	"fmt"
	"time"
)

// OptionType defines the type of the Option.
type OptionType string

const (
	// UniqueOpt makes sure that the task is enqueued only once within a certain period.
	UniqueOpt OptionType = "UniqueOpt"
	// InOpt enqueues the task after a certain time.Duration.
	InOpt = "InOpt"
	// AtOpt enqueues the task at a certain time.Time.
	AtOpt = "AtOpt"
)

// Option can be used to add additional behaviour to Adapter.Enqueue.
type Option interface {
	Type() OptionType
	Value() interface{}
	String() string
}

// Unique will enqueue a job uniquely.
var Unique = enqueueUnique{}

// In enqueues a job in the scheduled job queue for execution, default uniqueness TTL is 24 hours.
type In time.Duration

// Type returns InOpt OptionType.
func (i In) Type() OptionType {
	return InOpt
}

// Value returns Option value.
func (i In) Value() interface{} {
	return time.Duration(i)
}

func (i In) String() string {
	return fmt.Sprintf("In(%s)", time.Duration(i))
}

// At enqueues a job in the scheduled job queue for execution at time specified.
type At time.Time

// Type returns AtOpt OptionType.
func (a At) Type() OptionType {
	return AtOpt
}

// Value returns Option value.
func (a At) Value() interface{} {
	return time.Time(a)
}

func (a At) String() string {
	return fmt.Sprintf("At(%s)", time.Time(a).Format(time.RFC3339))
}

type enqueueUnique struct{}

func (enqueueUnique) Type() OptionType {
	return UniqueOpt
}

func (enqueueUnique) Value() interface{} {
	return true
}

func (enqueueUnique) String() string {
	return "Unique"
}
