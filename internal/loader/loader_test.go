package loader_test

import (
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"

	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("Loads type from targeted file", func(t *testing.T) {
		exported := packagestest.Export(t, packagestest.GOPATH, []packagestest.Module{
			{
				Name: "github.com/csueiras",
				Files: map[string]interface{}{
					"fake/fake.go": `package fake

import "context"

type Service interface {
	GetUserID(ctx context.Context, userID string) (string, error)
}
`,
				},
			},
			{
				Name: "github.com/csueiras",
				Files: map[string]interface{}{
					"fake/other.go": `package fake

import "context"

type OtherService interface {
	GetSomeOtherUserID(ctx context.Context, userID string) (string, error)
}
`,
				},
			},
		})
		defer exported.Cleanup()

		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadAll(exported.File("github.com/csueiras", "fake/fake.go"), loader.FileLoadMode)
		require.NoError(t, err)
		require.Equal(t, 1, len(results))
		svc, ok := results["Service"]
		require.True(t, ok)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.Equal(t, "Service", svc.Name)
		require.Equal(t, 1, len(svc.Methods))
		require.Equal(t, "GetUserID", svc.Methods[0].Name)
	})

	t.Run("Load Interface", func(t *testing.T) {

		exported := packagestest.Export(t, packagestest.GOPATH, []packagestest.Module{
			{
				Name: "github.com/csueiras",
				Files: map[string]interface{}{
					"fake/fake.go": `package fake

import "context"

type Service interface {
	GetUserID(ctx context.Context, userID string) (string, error)
}
`,
				},
			},
		})
		defer exported.Cleanup()

		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		svc, err := l.LoadOne("github.com/csueiras/fake", "Service", loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.Equal(t, "Service", svc.Name)
		require.Equal(t, 1, len(svc.Methods))
		require.Equal(t, "GetUserID", svc.Methods[0].Name)
	})

	t.Run("Load Struct", func(t *testing.T) {
		exported := packagestest.Export(t, packagestest.GOPATH, []packagestest.Module{{
			Name: "github.com/csueiras",
			Files: map[string]interface{}{
				"fake/fake.go": `package fake

import "context"

type service struct {	
}

func (s *service) GetUserID(ctx context.Context, userID string) (string, error) {
	return "My User", nil
}
`,
			}}})
		defer exported.Cleanup()

		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		svc, err := l.LoadOne("github.com/csueiras/fake", "service", loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.Equal(t, "service", svc.Name)
		require.Equal(t, 1, len(svc.Methods))
		require.Equal(t, "GetUserID", svc.Methods[0].Name)
	})
}

func TestLoadMatched(t *testing.T) {
	exported := packagestest.Export(t, packagestest.GOPATH, []packagestest.Module{{
		Name: "github.com/csueiras",
		Files: map[string]interface{}{
			"fake/fake.go": `package fake

import "context"

type UserService interface {
	GetUserID(ctx context.Context, userID string) (string, error)
}

type HelloWorldService interface {
	Hello(ctx context.Context, name string) error
}

type unexportedService interface {
	HelloWorld()
}

type StructWithNoMethods struct {
	SomeField string
}
`,
		}}})
	defer exported.Cleanup()

	t.Run("RegEx", func(t *testing.T) {
		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{".*Service"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 3, len(results))

		require.Equal(t, "UserService", results["UserService"].Name)
		require.Equal(t, 1, len(results["UserService"].Methods))
		require.Equal(t, "GetUserID", results["UserService"].Methods[0].Name)

		require.Equal(t, "HelloWorldService", results["HelloWorldService"].Name)
		require.Equal(t, 1, len(results["HelloWorldService"].Methods))
		require.Equal(t, "Hello", results["HelloWorldService"].Methods[0].Name)

		require.Equal(t, "unexportedService", results["unexportedService"].Name)
		require.Equal(t, 1, len(results["unexportedService"].Methods))
		require.Equal(t, "HelloWorld", results["unexportedService"].Methods[0].Name)
	})

	t.Run("Multiple RegEx Expressions", func(t *testing.T) {
		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{"User.*", "Hello.*Service"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 2, len(results))

		require.Equal(t, "UserService", results["UserService"].Name)
		require.Equal(t, 1, len(results["UserService"].Methods))
		require.Equal(t, "GetUserID", results["UserService"].Methods[0].Name)

		require.Equal(t, "HelloWorldService", results["HelloWorldService"].Name)
		require.Equal(t, 1, len(results["HelloWorldService"].Methods))
		require.Equal(t, "Hello", results["HelloWorldService"].Methods[0].Name)
	})

	t.Run("Exact Match", func(t *testing.T) {
		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{"HelloWorldService"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 1, len(results))
		require.Equal(t, "HelloWorldService", results["HelloWorldService"].Name)
		require.Equal(t, 1, len(results["HelloWorldService"].Methods))
		require.Equal(t, "Hello", results["HelloWorldService"].Methods[0].Name)
	})

	t.Run("Exact Match: No Match", func(t *testing.T) {
		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{"Hello"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 0, len(results))
	})

	t.Run("Multiple Exact Matches", func(t *testing.T) {
		l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
			exported.Config.Mode = cfg.Mode
			return packages.Load(exported.Config, patterns...)
		})

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{"UserService", "HelloWorldService", "StructWithNoMethods"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 2, len(results))

		require.Equal(t, "UserService", results["UserService"].Name)
		require.Equal(t, 1, len(results["UserService"].Methods))
		require.Equal(t, "GetUserID", results["UserService"].Methods[0].Name)

		require.Equal(t, "HelloWorldService", results["HelloWorldService"].Name)
		require.Equal(t, 1, len(results["HelloWorldService"].Methods))
		require.Equal(t, "Hello", results["HelloWorldService"].Methods[0].Name)
	})
}
