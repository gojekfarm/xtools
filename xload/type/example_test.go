package xloadtype_test

import (
	"context"
	"fmt"
	"github.com/gojekfarm/xtools/xload"
	xloadtype "github.com/gojekfarm/xtools/xload/type"
)

var testValues = map[string]string{
	"LISTENER": "[::1]:8080",
	"ENDPOINT": "example.com:80",
}

var loader = xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
	return testValues[key], nil
})

func ExampleEndpoint() {
	type Server struct {
		Endpoint xloadtype.Endpoint `env:"ENDPOINT"`
	}

	var srv Server
	if err := xload.Load(context.Background(), &srv, loader); err != nil {
		panic(err)
	}

	fmt.Println(srv.Endpoint.String())

	// Output: example.com:80
}

func ExampleListener() {
	type Server struct {
		Listener xloadtype.Listener `env:"LISTENER"`
	}

	var srv Server
	if err := xload.Load(context.Background(), &srv, loader); err != nil {
		panic(err)
	}

	fmt.Println(srv.Listener.String())

	// Output: [::1]:8080
}
