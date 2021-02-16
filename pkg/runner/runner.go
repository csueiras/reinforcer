package runner

import (
	"github.com/slok/goresilience"
	"sync"
)

// Factory of runners
type Factory struct {
	mu          sync.RWMutex
	runners     map[string]goresilience.Runner
	middlewares []goresilience.Middleware
}

// NewFactory creates an instance of a Runner factory that will create runners on demand if they don't exist otherwise
// return a singleton instance of a runner for each unique runner identifier.
func NewFactory(middlewares ...goresilience.Middleware) *Factory {
	return &Factory{
		runners:     make(map[string]goresilience.Runner),
		middlewares: middlewares,
	}
}

// GetRunner retrieves a runner with the given name, this is guaranteed to always return a Runner. This is thread-safe.
func (f *Factory) GetRunner(name string) goresilience.Runner {
	f.mu.RLock()
	if r, ok := f.runners[name]; ok {
		f.mu.RUnlock()
		return r
	}
	f.mu.RUnlock()

	// Obtain write lock as we need to mutate the underlying map
	f.mu.Lock()
	defer f.mu.Unlock()

	// The runner might've been created in between the two synchronized blocks
	if r, ok := f.runners[name]; ok {
		return r
	}
	runner := goresilience.RunnerChain(f.middlewares...)
	f.runners[name] = runner
	return runner
}
