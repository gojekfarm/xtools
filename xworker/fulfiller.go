package xworker

// Fulfiller is an interface that defines what all behaviour is required from a worker implementation.
type Fulfiller interface {
	Enqueuer
	PeriodicEnqueuer
	Registerer
	Start() error
	Stop() error
}
