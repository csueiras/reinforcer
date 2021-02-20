package loader_test

import (
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"

	"testing"
)

func TestLoad(t *testing.T) {
	exported := packagestest.Export(t, packagestest.GOPATH, []packagestest.Module{{
		Name: "github.com/csueiras",
		Files: map[string]interface{}{
			"fake/fake.go": `package fake

import "context"

type Service interface {
	GetUserID(ctx context.Context, userID string) (string, error)
}
`,
		}}})
	defer exported.Cleanup()

	l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
		exported.Config.Mode = cfg.Mode
		return packages.Load(exported.Config, patterns...)
	})

	svc, err := l.LoadOne("github.com/csueiras/fake", "Service", loader.PackageLoadMode)
	require.NoError(t, err)
	require.NotNil(t, svc)
	require.Equal(t, "interface{GetUserID(ctx context.Context, userID string) (string, error)}", svc.InterfaceType.String())
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
	ShouldNotBeSeen()
}

type NotAnInterface struct {
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
		require.Equal(t, 2, len(results))
		require.NotNil(t, results["UserService"])
		require.Equal(t, "interface{GetUserID(ctx context.Context, userID string) (string, error)}", results["UserService"].InterfaceType.String())
		require.NotNil(t, results["HelloWorldService"])
		require.Equal(t, "interface{Hello(ctx context.Context, name string) error}", results["HelloWorldService"].InterfaceType.String())
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
		require.NotNil(t, results["UserService"])
		require.Equal(t, "interface{GetUserID(ctx context.Context, userID string) (string, error)}", results["UserService"].InterfaceType.String())
		require.NotNil(t, results["HelloWorldService"])
		require.Equal(t, "interface{Hello(ctx context.Context, name string) error}", results["HelloWorldService"].InterfaceType.String())
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
		require.NotNil(t, results["HelloWorldService"])
		require.Equal(t, "interface{Hello(ctx context.Context, name string) error}", results["HelloWorldService"].InterfaceType.String())
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

		results, err := l.LoadMatched("github.com/csueiras/fake", []string{"UserService", "HelloWorldService", "NotAnInterface"}, loader.PackageLoadMode)
		require.NoError(t, err)
		require.NotNil(t, results)
		require.Equal(t, 2, len(results))
		require.NotNil(t, results["UserService"])
		require.Equal(t, "interface{GetUserID(ctx context.Context, userID string) (string, error)}", results["UserService"].InterfaceType.String())
		require.NotNil(t, results["HelloWorldService"])
		require.Equal(t, "interface{Hello(ctx context.Context, name string) error}", results["HelloWorldService"].InterfaceType.String())
	})
}
