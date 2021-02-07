package utils

import (
	"fmt"
	"github.com/csueiras/reinforcer/internal/loader"
	. "github.com/dave/jennifer/jen"
	"go/types"
)

type named interface {
	Name() string
}

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

	_, iface, err := loader.Loader().Load("context", "Context")
	if err != nil {
		panic(err)
	}
	ContextType = iface
}

// IsErrorType determines if the given type implements the Error interface
func IsErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	return types.Implements(t, ErrType.Underlying().(*types.Interface))
}

func IsContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	if t.String() == "context.Context" {
		return true
	}
	return types.Implements(t, ContextType)
}

func VariadicToType(t types.Type) (Code, error) {
	sliceType, ok := t.(*types.Slice)
	if !ok {
		return nil, fmt.Errorf("expected type to be *types.Slice, got=%T", t)
	}
	sliceElemType, err := ToType(sliceType.Elem(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert slice's type; error=%w", err)
	}
	return Op("...").Add(sliceElemType), nil
}

func ToType(t types.Type, variadic bool) (Code, error) {
	if variadic {
		return VariadicToType(t)
	}

	switch v := t.(type) {
	case *types.Basic:
		return Id(v.Name()), nil
	case *types.Named:
		typeName := v.Obj()
		if _, ok := v.Underlying().(*types.Interface); ok {
			if typeName.Pkg() != nil {
				return Qual(
					typeName.Pkg().Path(),
					typeName.Name(),
				), nil
			}
			return Id(typeName.Name()), nil
		}
		return Qual(
			typeName.Pkg().Path(),
			typeName.Name(),
		), nil
	case *types.Pointer:
		rt, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return Op("*").Add(rt), nil
	case *types.Interface:
		if v.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}
		return Id("interface{}"), nil
	case *types.Slice:
		elemType, err := ToType(v.Elem(), false)
		if err != nil {
			return nil, err
		}
		return Index().Add(elemType), nil
	case named:
		return Id(v.Name()), nil
	default:
		return nil, fmt.Errorf("type not hanled: %T", v)
	}
}
