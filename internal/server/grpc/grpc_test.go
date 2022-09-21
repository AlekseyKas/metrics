package grpc

import (
	"testing"
	"time"
)

func TestGRPC(t *testing.T) {
	go func() {
		srv := New()
		srv.Start()
	}()
	time.Sleep(2 * time.Second)
	cli()
}
