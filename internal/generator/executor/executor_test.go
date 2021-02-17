package executor_test

import (
	"github.com/csueiras/reinforcer/internal/generator/executor"
	"github.com/csueiras/reinforcer/internal/generator/executor/mocks"
	"github.com/stretchr/testify/require"
	"go/token"
	"go/types"
	"testing"
)

func TestExecutor_Execute(t *testing.T) {
	l := &mocks.Loader{}
	l.On("LoadMatched", "github.com/csueiras/reinforcer/pkg/testpkg", []string{"MyService"}).Return(
		map[string]*types.Interface{
			"LockService": createTestInterfaceType(),
		}, nil,
	)

	exec := executor.New(l)
	got, err := exec.Execute(&executor.Parameters{
		Sources:               []string{"github.com/csueiras/reinforcer/pkg/testpkg"},
		Targets:               []string{"MyService"},
		OutPkg:                "testpkg",
		IgnoreNoReturnMethods: false,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 1, len(got.Files))
	require.Equal(t, "LockService", got.Files[0].TypeName)
}

func createTestInterfaceType() *types.Interface {
	nullary := types.NewSignature(nil, nil, nil, false) // func()
	methods := []*types.Func{
		types.NewFunc(token.NoPos, nil, "Lock", nullary),
		types.NewFunc(token.NoPos, nil, "Unlock", nullary),
	}
	return types.NewInterfaceType(methods, nil).Complete()
}
