package retryable_test

import (
	"bytes"
	"github.com/csueiras/reinforcer/internal/generator/method"
	"github.com/csueiras/reinforcer/internal/generator/retryable"
	"github.com/stretchr/testify/require"
	"go/token"
	"go/types"
	"testing"
)

func TestRetryable_Statement(t *testing.T) {
	errVar := types.NewVar(token.NoPos, nil, "", method.ErrType)
	ctxVar := types.NewVar(token.NoPos, nil, "ctx", method.ContextType)

	tests := []struct {
		name       string
		methodName string
		signature  *types.Signature
		want       string
		wantErr    bool
	}{
		{
			name:       "Function returns error",
			methodName: "MyFunction",
			signature:  types.NewSignature(nil, types.NewTuple(), types.NewTuple(errVar), false),
			want: `func (r *resilient) MyFunction() error {
	var nonRetryableErr error
	err := r.run(context.Background(), ParentMethods.MyFunction, func(_ context.Context) error {
		var err error
		err = r.delegate.MyFunction()
		if r.errorPredicate(ParentMethods.MyFunction, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return nonRetryableErr
	}
	return err
}`,
			wantErr: false,
		},
		{
			name:       "Function returns string and error",
			methodName: "MyFunction",
			signature:  types.NewSignature(nil, types.NewTuple(), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String]), errVar), false),
			want: `func (r *resilient) MyFunction() (string, error) {
	var nonRetryableErr error
	var r0 string
	err := r.run(context.Background(), ParentMethods.MyFunction, func(_ context.Context) error {
		var err error
		r0, err = r.delegate.MyFunction()
		if r.errorPredicate(ParentMethods.MyFunction, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}`,
			wantErr: false,
		},
		{
			name:       "Function passes arguments",
			methodName: "MyFunction",
			signature: types.NewSignature(nil, types.NewTuple(
				ctxVar,
				types.NewVar(token.NoPos, nil, "myArg", types.Typ[types.String]),
			), types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String]), errVar), false),
			want: `func (r *resilient) MyFunction(ctx context.Context, arg1 string) (string, error) {
	var nonRetryableErr error
	var r0 string
	err := r.run(ctx, ParentMethods.MyFunction, func(ctx context.Context) error {
		var err error
		r0, err = r.delegate.MyFunction(ctx, arg1)
		if r.errorPredicate(ParentMethods.MyFunction, err) {
			return err
		}
		nonRetryableErr = err
		return nil
	})
	if nonRetryableErr != nil {
		return r0, nonRetryableErr
	}
	return r0, err
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := method.ParseMethod(nil, "Parent", tt.methodName, tt.signature)
			require.NoError(t, err)
			ret := retryable.NewRetryable(m, "resilient", "r")
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
				require.Equal(t, tt.want, got)
			}
		})
	}

	t.Run("Function does not return error", func(t *testing.T) {
		require.Panics(t, func() {
			m, err := method.ParseMethod(nil, "Parent", "Fn", types.NewSignature(nil, types.NewTuple(), types.NewTuple(), false))
			require.NoError(t, err)
			retryable.NewRetryable(m, "resilient", "r")
		})
	})
}
