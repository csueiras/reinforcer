package runner_test

import (
	"context"
	"github.com/csueiras/reinforcer/pkg/runner"
	"github.com/slok/goresilience"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFactory_GetRunner(t *testing.T) {
	mwCalled := 0
	mwCreated := 0
	f := runner.NewFactory(
		func(r goresilience.Runner) goresilience.Runner {
			mwCreated++
			return goresilience.RunnerFunc(func(ctx context.Context, f goresilience.Func) error {
				mwCalled++
				return r.Run(ctx, f)
			})
		},
	)

	require.NoError(t, f.GetRunner("Call1").Run(context.Background(), func(ctx context.Context) error {
		return nil
	}))
	require.NoError(t, f.GetRunner("Call2").Run(context.Background(), func(ctx context.Context) error {
		return nil
	}))
	require.NoError(t, f.GetRunner("Call1").Run(context.Background(), func(ctx context.Context) error {
		return nil
	}))
	require.NoError(t, f.GetRunner("Call1").Run(context.Background(), func(ctx context.Context) error {
		return nil
	}))
	require.Equal(t, 4, mwCalled)
	require.Equal(t, 2, mwCreated)
}
