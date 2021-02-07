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
	t.Cleanup(exported.Cleanup)

	l := loader.NewLoader(func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
		exported.Config.Mode = cfg.Mode
		return packages.Load(exported.Config, patterns...)
	})

	pkg, svc, err := l.Load("github.com/csueiras/fake", "Service")
	require.NoError(t, err)
	require.NotNil(t, pkg)
	require.NotNil(t, svc)
	require.Equal(t, "interface{GetUserID(ctx context.Context, userID string) (string, error)}", svc.String())
}
