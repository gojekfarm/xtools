package xloadtype

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListener_Decode(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *Listener
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid ipv4",
			in:   "127.0.0.1:8080",
			want: &Listener{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 8080,
			},
			wantErr: assert.NoError,
		},
		{
			name: "valid ipv6",
			in:   "[::1]:8080",
			want: &Listener{
				IP:   net.IPv6loopback,
				Port: 8080,
			},
			wantErr: assert.NoError,
		},
		{
			name: "missing port",
			in:   "localhost",
			want: &Listener{},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				return assert.EqualError(t, err, "address localhost: missing port in address")
			},
		},
		{
			name: "invalid port",
			in:   "127.0.0.1:port",
			want: &Listener{IP: net.IPv4(127, 0, 0, 1)},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				return assert.EqualError(t, err, `strconv.ParseInt: parsing "port": invalid syntax`)
			},
		},
		{
			name: "invalid ip",
			in:   "localhost:8080",
			want: &Listener{},
			wantErr: func(t assert.TestingT, err error, i ...any) bool {
				return assert.EqualError(t, err, "invalid IP address")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := new(Listener)
			tt.wantErr(t, l.Decode(tt.in))
			assert.Equal(t, tt.want, l)
		})
	}
}

func TestListener_String(t *testing.T) {
	tests := []struct {
		name string
		in   *Listener
		want string
	}{
		{
			name: "valid ipv4",
			in:   &Listener{IP: net.IPv4(127, 0, 0, 1), Port: 8080},
			want: "127.0.0.1:8080",
		},
		{
			name: "valid ipv6",
			in:   &Listener{IP: net.IPv6loopback, Port: 8080},
			want: "[::1]:8080",
		},
		{
			name: "empty ip",
			in: &Listener{
				Port: 8080,
			},
			want: ":8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.String())
		})
	}
}
