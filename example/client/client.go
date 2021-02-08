//go:generate reinforcer --name=Client --outputdir=./reinforced

package client

import (
	"context"
	"fmt"
	"math/rand"
)

// Client is an example service interface that we will generate code for
type Client interface {
	SayHello(ctx context.Context, name string) error
}

// FakeClient is a Client implementation that will randomly fail
type FakeClient struct {
}

// SayHello is a method that will randomly return an error otherwise it will print a nice greeting
func (f *FakeClient) SayHello(_ context.Context, name string) error {
	if rand.Int()%10 == 5 {
		return fmt.Errorf("random failure")
	}
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

// NewClient is a ctor for fakeClient
func NewClient() *FakeClient {
	return &FakeClient{}
}
