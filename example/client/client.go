//go:generate reinforcer --debug --target=Client --target=SomeOtherClient --outputdir=./reinforced

package client

import (
	"context"
	"fmt"
	"github.com/csueiras/reinforcer/example/client/sub"
	"math/rand"
	"os"
)

// File is a type defined in the same package that we are using for targetting
type File struct {
	Name string
}

// Client is an example service interface that we will generate code for
type Client interface {
	SayHello(ctx context.Context, name string) error
	GenerateGreeting(ctx context.Context, name string) (string, error)
}

// SomeOtherClient is another example service interface that can be targeted
type SomeOtherClient interface {
	DoStuff() error
	SaveFile(myFile *File, osFile *os.File) error
	GetUser(ctx context.Context) (*sub.User, error)
	MethodWithChannel(myChan <- chan bool) error
	MethodWithWildcard(arg interface{})
}

// FakeClient is a Client implementation that will randomly fail
type FakeClient struct {
}

// SayHello is a method that will randomly return an error otherwise it will print a nice greeting
func (f *FakeClient) SayHello(ctx context.Context, name string) error {
	greeting, err := f.GenerateGreeting(ctx, name)
	if err != nil {
		return err
	}
	fmt.Print(greeting)
	return nil
}

// GenerateGreeting generates a string for a greeting, this will randomly return errors
func (f *FakeClient) GenerateGreeting(_ context.Context, name string) (string, error) {
	if rand.Int()%10 == 5 {
		return "", fmt.Errorf("random failure")
	}
	return fmt.Sprintf("Hello, %s!\n", name), nil
}

// NewClient is a ctor for fakeClient
func NewClient() *FakeClient {
	return &FakeClient{}
}
