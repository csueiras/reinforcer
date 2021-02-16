package main

import (
	"context"
	"fmt"
	"github.com/csueiras/reinforcer/example/client"
	"github.com/csueiras/reinforcer/example/client/reinforced"
	"github.com/csueiras/reinforcer/pkg/runner"
	"github.com/slok/goresilience/retry"
	"github.com/slok/goresilience/timeout"
	"time"
)

func main() {
	cl := client.NewClient()
	f := runner.NewFactory(
		timeout.NewMiddleware(timeout.Config{Timeout: 100 * time.Millisecond}),
		retry.NewMiddleware(retry.Config{
			Times: 10,
		}),
	)
	rCl := reinforced.NewClient(cl, f)
	for i := 0; i < 100; i++ {
		err := rCl.SayHello(context.Background(), "Christian")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
