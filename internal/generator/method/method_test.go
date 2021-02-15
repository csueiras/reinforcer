package method_test

import (
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
	"go/token"
	"go/types"
	"testing"
)

func TestNewMethod(t *testing.T) {
	ctxVar := types.NewVar(token.NoPos, nil, "ctx", method.ContextType)
	zero := new(int)
	*zero = 0

	type args struct {
		name      string
		signature *types.Signature
	}
	tests := []struct {
		name string
		args args
		want *method.Method
	}{

		{
			name: "Fn()",
			args: args{
				name:      "Fn",
				signature: types.NewSignature(nil, types.NewTuple(), types.NewTuple(), false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            false,
				ParameterNames:        nil,
				ParametersNameAndType: nil,
			},
		},
		{
			name: "Fn(arg string)",
			args: args{
				name:      "Fn",
				signature: types.NewSignature(nil, types.NewTuple(types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String])), types.NewTuple(), false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            false,
				ParameterNames:        []string{"arg0"},
				ParametersNameAndType: []jen.Code{jen.Id("arg0").Add(jen.Id("string"))},
			},
		},
		{
			name: "Fn(args ...string)",
			args: args{
				name:      "Fn",
				signature: types.NewSignature(nil, types.NewTuple(types.NewVar(token.NoPos, nil, "args", types.NewSlice(types.Typ[types.String]))), types.NewTuple(), true),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            false,
				ParameterNames:        []string{"arg0"},
				ParametersNameAndType: []jen.Code{jen.Id("arg0").Add(jen.Op("...").Add(jen.Id("string")))},
			},
		},
		{
			name: "Fn(arg0 string, args ...string)",
			args: args{
				name: "Fn",
				signature: types.NewSignature(nil,
					types.NewTuple(types.NewVar(token.NoPos, nil, "arg0", types.Typ[types.String]), types.NewVar(token.NoPos, nil, "args", types.NewSlice(types.Typ[types.String]))), types.NewTuple(), true),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            false,
				HasVariadic:           true,
				ParameterNames:        []string{"arg0", "arg1"},
				ParametersNameAndType: []jen.Code{jen.Id("arg0").Add(jen.Id("string")), jen.Id("arg1").Add(jen.Op("...").Add(jen.Id("string")))},
			},
		},
		{
			name: "Fn(ctx context.Context, arg1 string)",
			args: args{
				name: "Fn",
				signature: types.NewSignature(nil, types.NewTuple(
					ctxVar,
					types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
				), types.NewTuple(), false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            true,
				ContextParameter:      zero,
				ReturnsError:          false,
				ParameterNames:        []string{"ctx", "arg1"},
				ParametersNameAndType: []jen.Code{jen.Id("ctx").Add(jen.Qual("context", "Context")), jen.Id("arg1").Add(jen.Id("string"))},
				ReturnTypes:           nil,
			},
		},
		{
			name: "Fn(ctx context.Context, arg1 string) error",
			args: args{
				name: "Fn",
				signature: types.NewSignature(nil, types.NewTuple(
					ctxVar,
					types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
				), types.NewTuple(types.NewVar(token.NoPos, nil, "", method.ErrType)), false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            true,
				ReturnsError:          true,
				ParameterNames:        []string{"ctx", "arg1"},
				ParametersNameAndType: []jen.Code{jen.Id("ctx").Add(jen.Qual("context", "Context")), jen.Id("arg1").Add(jen.Id("string"))},
				ReturnTypes:           []jen.Code{jen.Id("error")},
			},
		},
		{
			name: "Fn(ctx context.Context, arg1 string) (string, error)",
			args: args{
				name: "Fn",
				signature: types.NewSignature(nil, types.NewTuple(
					ctxVar,
					types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
				), types.NewTuple(
					types.NewVar(token.NoPos, nil, "", types.Typ[types.String]),
					types.NewVar(token.NoPos, nil, "", method.ErrType),
				), false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            true,
				ReturnsError:          true,
				ParameterNames:        []string{"ctx", "arg1"},
				ParametersNameAndType: []jen.Code{jen.Id("ctx").Add(jen.Qual("context", "Context")), jen.Id("arg1").Add(jen.Id("string"))},
				ReturnTypes:           []jen.Code{jen.Id("string"), jen.Id("error")},
			},
		},
		{
			name: "Fn(arg interface{}) interface{}",
			args: args{
				name: "Fn",
				signature: types.NewSignature(nil,
					types.NewTuple(types.NewVar(token.NoPos, nil, "arg", types.NewInterfaceType(nil, nil))),
					types.NewTuple(types.NewVar(token.NoPos, nil, "", types.NewInterfaceType(nil, nil))),
					false),
			},
			want: &method.Method{
				Name:                  "Fn",
				HasContext:            false,
				ParameterNames:        []string{"arg0"},
				ParametersNameAndType: []jen.Code{jen.Id("arg0").Add(jen.Id("interface{}"))},
				ReturnTypes:           []jen.Code{jen.Id("interface{}")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := method.ParseMethod(tt.args.name, tt.args.signature)
			require.NoError(t, err)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.HasContext, got.HasContext)
			require.Equal(t, tt.want.ReturnsError, got.ReturnsError)
			if tt.want.ContextParameter != nil {
				require.Equal(t, *tt.want.ContextParameter, *got.ContextParameter)
			}
			if tt.want.ReturnErrorIndex != nil {
				require.Equal(t, *tt.want.ReturnErrorIndex, *got.ReturnErrorIndex)
			}
			require.ElementsMatch(t, tt.want.ParameterNames, got.ParameterNames)
			require.ElementsMatch(t, tt.want.ParametersNameAndType, got.ParametersNameAndType)
			require.ElementsMatch(t, tt.want.ReturnTypes, got.ReturnTypes)
		})
	}
}
