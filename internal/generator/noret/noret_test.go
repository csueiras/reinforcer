package noret_test

import (
	"bytes"
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/csueiras/reinforcer/internal/generator/noret"
	"github.com/stretchr/testify/require"
	"go/token"
	"go/types"
	"testing"
)

func TestNoReturn_Statement(t *testing.T) {
	ctxVar := types.NewVar(token.NoPos, nil, "ctx", method.ContextType)

	tests := []struct {
		name       string
		methodName string
		signature  *types.Signature
		want       string
		wantErr    bool
	}{
		{
			name:       "MyFunction()",
			methodName: "MyFunction",
			signature:  types.NewSignature(nil, types.NewTuple(), types.NewTuple(), false),
			want: `func (r *resilient) MyFunction() {
	err := r.run(context.Background(), ParentMethods.MyFunction, func(_ context.Context) error {
		r.delegate.MyFunction()
		return nil
	})
	if err != nil {
		panic(err)
	}
}`,
			wantErr: false,
		},
		{
			name:       "MyFunction(ctx context.Context, arg1 string)",
			methodName: "MyFunction",
			signature: types.NewSignature(nil, types.NewTuple(
				ctxVar,
				types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
			), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])), false),
			want: `func (r *resilient) MyFunction(ctx context.Context, arg1 string) {
	err := r.run(ctx, ParentMethods.MyFunction, func(ctx context.Context) error {
		r.delegate.MyFunction(ctx, arg1)
		return nil
	})
	if err != nil {
		panic(err)
	}
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := method.ParseMethod(nil, "Parent", tt.methodName, tt.signature)
			require.NoError(t, err)
			ret := noret.NewNoReturn(m, "resilient", "r")
			buf := &bytes.Buffer{}
			s, err := ret.Statement()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NotNil(t, s)
				require.NoError(t, err)
				renderErr := s.Render(buf)
				require.NoError(t, renderErr)

				got := buf.String()
				//fmt.Print(got)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
