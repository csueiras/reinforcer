//nolint

package testpkg

import "context"

type service struct {
}

func (s *service) unexportedOperation(arg string) (string, error) {
	return arg, nil
}

func (s *service) GetUserByID(_ context.Context, _ string) (string, error) {
	return "Christian", nil
}

type anotherService struct{}

func (a anotherService) DoOperation() {
}
