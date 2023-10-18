package grpc_test

import (
	"google.golang.org/grpc"

	grpctrace "github.com/gojekfarm/xtools/xtel/grpc"
)

func ExampleUnaryServerInterceptor() {
	grpc.NewServer(grpc.ChainUnaryInterceptor(grpctrace.UnaryServerInterceptor))
}

func ExampleUnaryClientInterceptor() {
	conn, err := grpc.Dial("fake:connection", grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
}
