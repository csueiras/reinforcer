package types

import (
	"go/types"
	"golang.org/x/tools/go/packages"
)

// ErrType is the types.Type for the error interface
var ErrType types.Type

// ContextType is the types.Type for the context.Context interface
var ContextType *types.Interface

func init() {
	errType := types.NewInterfaceType([]*types.Func{
		types.NewFunc(0, nil, "Error",
			types.NewSignature(
				nil,
				types.NewTuple(),
				types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.String])),
				false,
			),
		),
	}, nil)
	errType.Complete()
	ErrType = types.NewNamed(types.NewTypeName(0, nil, "error", nil), errType, nil)

	// Load the type definition for the Context type
	ctxPkg, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax | packages.NeedTypesInfo,
	}, "context")
	if err != nil {
		panic(err)
	}
	ContextType = ctxPkg[0].Types.
		Scope().
		Lookup("Context").
		Type().(*types.Named).
		Underlying().
		(*types.Interface)
}

// IsErrorType determines if the given type implements the Error interface
func IsErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.Implements(t, ErrType.Underlying().(*types.Interface))
}

// IsContextType determines if the given type is context.Context
func IsContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	if t.String() == "context.Context" {
		return true
	}
	return types.Implements(t, ContextType)
}
