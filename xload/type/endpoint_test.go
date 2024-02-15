package xloadtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpoint_Decode(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *Endpoint
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			in:   "localhost:8080",
			want: &Endpoint{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid",
			in:   "localhost",
			want: &Endpoint{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "address localhost: missing port in address")
			},
		},
		{
			name: "invalid port",
			in:   "localhost:port",
			want: &Endpoint{Host: "localhost"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, `strconv.ParseInt: parsing "port": invalid syntax`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := new(Endpoint)
			tt.wantErr(t, e.Decode(tt.in))
			assert.Equal(t, tt.want, e)
		})
	}
}

func TestEndpoint_String(t *testing.T) {
	tests := []struct {
		name string
		in   *Endpoint
		want string
	}{
		{
			name: "valid",
			in:   &Endpoint{Host: "localhost", Port: 8080},
			want: "localhost:8080",
		},
		{
			name: "empty",
			in:   &Endpoint{},
			want: ":0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.String())
		})
	}
}
