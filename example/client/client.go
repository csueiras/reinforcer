//go:generate reinforcer --name=Client --outputdir=./reinforced

package client

import (
	"context"
	"fmt"
	"math/rand"
)

type Client interface {
	SayHello(ctx context.Context, name string) error
}

type fakeClient struct {
}

func (f *fakeClient) SayHello(_ context.Context, name string) error {
	if rand.Int()%10 == 5 {
		return fmt.Errorf("random failure")
	}
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

func NewClient() *fakeClient {
	return &fakeClient{}
}
