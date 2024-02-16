package xloadtype

import (
	"fmt"
	"net"
	"strconv"
)

// Endpoint represents a network endpoint
// It can be used to represent a target host:port pair.
type Endpoint struct {
	Host string
	Port int
}

func (e *Endpoint) String() string { return fmt.Sprintf("%s:%d", e.Host, e.Port) }

func (e *Endpoint) Decode(v string) error {
	host, port, err := net.SplitHostPort(v)
	if err != nil {
		return err
	}

	e.Host = host

	p, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return err
	}

	e.Port = int(p)

	return nil
}
