//go:generate protoc --go_out=. test.proto
//go:generate mv test.pb.go pb_test.go

package xproto
