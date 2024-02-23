package worker

import (
	"context"
	"time"
)

type TaskExecutor interface {
	Timeout() time.Duration
	Execute(ctx context.Context)
}

type Task[T any] struct {
	timeout  time.Duration
	executor func(ctx context.Context) (T, error)
	callback func(result T, err error)
}

// NewTask creates a new Task.
func NewTask[T any](executor func(ctx context.Context) (T, error)) *Task[T] {
	return &Task[T]{
		executor: executor,
	}
}

// Callback adds a callback function to the task.
func (t *Task[T]) Callback(callback func(result T, err error)) *Task[T] {
	t.callback = callback
	return t
}

// Timeout returns the timeout for the current task.
func (t *Task[T]) Timeout() time.Duration {
	return t.timeout
}

// Execute invokes the as executor definded function of the task and passes the result to the callback function of the task.
func (t *Task[T]) Execute(ctx context.Context) {
	result, err := t.executor(ctx)
	if t.callback != nil {
		t.callback(result, err)
	}
}

type WorkerManager struct {
	workerCount int
	taskCh      chan TaskExecutor
}

// NewWorkerManager create a new WorkerManager.
func NewWorkerManager(workerCount int) *WorkerManager {
	return &WorkerManager{
		workerCount: workerCount,
		taskCh:      make(chan TaskExecutor),
	}
}

// Run runs a specified number of workers.
func (w *WorkerManager) Run(ctx context.Context) error {
	for i := 0; i < w.workerCount; i++ {
		// run each worker in its own goroutine
		go w.worker(ctx)
	}

	// block until the context is cancelled
	<-ctx.Done()

	close(w.taskCh)
	return ctx.Err()
}

// Add adds a task to the task channel.
func (w *WorkerManager) Add(task TaskExecutor) {
	w.taskCh <- task
}

// worker starts a
func (w *WorkerManager) worker(parentCtx context.Context) {
	for {
		select {
		case <-parentCtx.Done():
			// exit the goroutine when the context is cancelled
			return
		case task, ok := <-w.taskCh:
			if !ok {
				// exit the goroutine when the channel is closed
				return
			}
			ctx, cancel := context.WithTimeout(parentCtx, task.Timeout())
			task.Execute(ctx)
			// cancel task after the task is done
			cancel()
		}
	}
}
