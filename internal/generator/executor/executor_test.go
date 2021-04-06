package executor_test

import (
	"github.com/csueiras/reinforcer/internal/generator/executor"
	"github.com/csueiras/reinforcer/internal/generator/executor/mocks"
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/csueiras/reinforcer/internal/loader"
	"github.com/stretchr/testify/require"
	"go/types"
	"testing"
)

func TestExecutor_Execute(t *testing.T) {
	t.Run("Loads types", func(t *testing.T) {
		l := &mocks.Loader{}
		l.On("LoadMatched", "./testpkg.go", []string{"MyService"}, loader.FileLoadMode).Return(
			map[string]*loader.Result{
				"LockService": {
					Name:    "LockService",
					Methods: createTestServiceMethods(),
				},
			}, nil,
		)

		exec := executor.New(l)
		got, err := exec.Execute(&executor.Parameters{
			Sources:               []string{"./testpkg.go"},
			Targets:               []string{"MyService"},
			OutPkg:                "testpkg",
			IgnoreNoReturnMethods: false,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, 1, len(got.Files))
		require.Equal(t, "LockService", got.Files[0].TypeName)
	})

	t.Run("Loads types from packages", func(t *testing.T) {
		l := &mocks.Loader{}
		l.On("LoadMatched", "github.com/csueiras/somelib", []string{"MyService"}, loader.PackageLoadMode).Return(
			map[string]*loader.Result{
				"LockService": {
					Name:    "LockService",
					Methods: createTestServiceMethods(),
				},
			}, nil,
		)

		exec := executor.New(l)
		got, err := exec.Execute(&executor.Parameters{
			SourcePackages:        []string{"github.com/csueiras/somelib"},
			Targets:               []string{"MyService"},
			OutPkg:                "testpkg",
			IgnoreNoReturnMethods: false,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, 1, len(got.Files))
		require.Equal(t, "LockService", got.Files[0].TypeName)
	})

	t.Run("No types found", func(t *testing.T) {
		l := &mocks.Loader{}
		l.On("LoadMatched", "./testpkg.go", []string{"MyService"}, loader.FileLoadMode).
			Return(map[string]*loader.Result{}, nil)

		exec := executor.New(l)
		got, err := exec.Execute(&executor.Parameters{
			Sources:               []string{"./testpkg.go"},
			Targets:               []string{"MyService"},
			OutPkg:                "testpkg",
			IgnoreNoReturnMethods: false,
		})
		require.EqualError(t, err, executor.ErrNoTargetableTypesFound.Error())
		require.Nil(t, got)
	})
}

func createTestServiceMethods() []*method.Method {
	nullary := types.NewSignature(nil, nil, nil, false) // func()
	return []*method.Method{
		method.MustParseMethod("Lock", nullary),
		method.MustParseMethod("Unlock", nullary),
	}
}
