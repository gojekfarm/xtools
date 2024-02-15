package xloadtype

import (
	"fmt"
	"net"
	"strconv"
)

type Listener struct {
	IP   net.IP
	Port int
}

func (l *Listener) String() string {
	if l.IP == nil {
		return fmt.Sprintf(":%d", l.Port)
	}

	return net.JoinHostPort(l.IP.String(), strconv.Itoa(l.Port))
}

func (l *Listener) Decode(v string) error {
	host, port, err := net.SplitHostPort(v)
	if err != nil {
		return err
	}

	if host != "" {
		l.IP = net.ParseIP(host)

		if l.IP == nil {
			return net.InvalidAddrError("invalid IP address")
		}
	}

	p, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return err
	}

	l.Port = int(p)

	return err
}
