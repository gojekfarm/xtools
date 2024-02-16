package xloadtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL_Decode(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *URL
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid",
			in:      "http://localhost:8080",
			want:    &URL{Scheme: "http", Host: "localhost:8080"},
			wantErr: assert.NoError,
		},
		{
			name: "schema missing",
			in:   "://localhost:8080",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, `parse "://localhost:8080": missing protocol scheme`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := new(URL)
			tt.wantErr(t, u.Decode(tt.in))

			if tt.want != nil {
				assert.Equal(t, tt.want, u)
			} else {
				assert.Equal(t, &URL{}, u)
			}
		})
	}
}

func TestURL_String(t *testing.T) {
	tests := []struct {
		name string
		in   *URL
		want string
	}{
		{
			name: "valid",
			in:   &URL{Scheme: "http", Host: "localhost:8080"},
			want: "http://localhost:8080",
		},
		{
			name: "empty",
			in:   &URL{},
			want: "",
		},
		{
			name: "valid with path",
			in:   &URL{Scheme: "http", Host: "localhost:8080", Path: "/path"},
			want: "http://localhost:8080/path",
		},
		{
			name: "valid with query",
			in:   &URL{Scheme: "http", Host: "localhost:8080", Path: "/path", RawQuery: "key=value"},
			want: "http://localhost:8080/path?key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.in.String())
		})
	}
}

func TestURL_Endpoint(t *testing.T) {
	tests := []struct {
		name    string
		u       URL
		want    *Endpoint
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid",
			u:       URL{Scheme: "http", Host: "localhost:8080"},
			want:    &Endpoint{Host: "localhost", Port: 8080},
			wantErr: assert.NoError,
		},
		{
			name: "invalid",
			u:    URL{Scheme: "http", Host: "localhost"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, "address localhost: missing port in address")
			},
		},
		{
			name: "invalid port",
			u:    URL{Scheme: "http", Host: "localhost:port"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, `strconv.ParseInt: parsing "port": invalid syntax`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.u.Endpoint()
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
