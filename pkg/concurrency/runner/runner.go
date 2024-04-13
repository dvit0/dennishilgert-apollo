package runner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var ErrManagerAlreadyStarted = errors.New("runner manager already started")

// Runner is a function that runs a task.
type Runner func(ctx context.Context) error

type RunnerManager interface {
	Add(runner ...Runner) error
	Run(ctx context.Context) error
}

// RunnerManager is a manager for runners. It runs all runners in parallel and
// waits for all runners to finish. If any runner returns, the RunnerManager
// will stop all other runners and return any error.
type runnerManager struct {
	runners []Runner
	lock    sync.Mutex
	running atomic.Bool
}

// NewRunnerManager creates a new RunnerManager.
func NewRunnerManager(runners ...Runner) RunnerManager {
	return &runnerManager{
		runners: runners,
	}
}

// Add adds a new runner to the RunnerManager.
func (r *runnerManager) Add(runner ...Runner) error {
	if r.running.Load() {
		return ErrManagerAlreadyStarted
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.runners = append(r.runners, runner...)
	return nil
}

// Run runs all runners in parallel and waits for all runners to finish. If any
// runner returns, the RunnerManager will stop all other runners and return any
// error.
func (r *runnerManager) Run(ctx context.Context) error {
	if !r.running.CompareAndSwap(false, true) {
		return ErrManagerAlreadyStarted
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	for _, runner := range r.runners {
		go func(runner Runner) {
			// Since the task returned, we need to cancel all other tasks.
			// This is a noop if the parent context is already cancelled, or another
			// task returned before this one.
			defer cancel()

			// Ignore context cancelled errors since errors from a runner manager
			// will likely determine the exit code of the program.
			// Context cancelled errors are also not really useful to the user in
			// this situation.
			err := runner(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
				return
			}
			errCh <- nil
		}(runner)
	}

	// Collect all errors.
	errObjs := make([]error, 0)
	for i := 0; i < len(r.runners); i++ {
		err := <-errCh
		if err != nil {
			errObjs = append(errObjs, err)
		}
	}

	close(errCh) // Ensure channel is closed to avoid goroutine leak.

	return errors.Join(errObjs...)
}
