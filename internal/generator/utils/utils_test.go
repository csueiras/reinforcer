package utils_test

import (
	"github.com/csueiras/reinforcer/internal/generator/utils"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestIsErrorType(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		inLookup string
		want     bool
	}{
		{
			name:     "Is Error",
			in:       `package testpkg; var err error`,
			inLookup: `err`,
			want:     true,
		},
		{
			name:     "Is Error: Not An Error",
			in:       `package testpkg; var someString string`,
			inLookup: `someString`,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			fast, err := parser.ParseFile(fset, "", tt.in, parser.AllErrors)
			if err != nil {
				panic(err)
			}
			conf := types.Config{
				Importer: importer.Default(),
			}
			pkg, err := conf.Check("testpkg", fset, []*ast.File{fast}, nil)
			require.NoError(t, err)

			e := pkg.Scope().Lookup(tt.inLookup)
			require.Equal(t, tt.want, utils.IsErrorType(e.Type()))
		})
	}
}
