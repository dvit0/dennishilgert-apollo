package worker

import (
	"context"
	"errors"
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
func NewTask[T any](executor func(ctx context.Context) (T, error), timeout time.Duration) *Task[T] {
	return &Task[T]{
		timeout:  timeout,
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

type WorkerManager interface {
	Run(ctx context.Context) error
	Add(task TaskExecutor)
}

type workerManager struct {
	workerCount int
	taskCh      chan TaskExecutor
}

// NewWorkerManager create a new WorkerManager.
func NewWorkerManager(workerCount int) WorkerManager {
	return &workerManager{
		workerCount: workerCount,
		taskCh:      make(chan TaskExecutor),
	}
}

// Run runs a specified number of workers.
func (w *workerManager) Run(ctx context.Context) error {
	for i := 0; i < w.workerCount; i++ {
		// Run each worker in its own goroutine.
		go w.worker(ctx)
	}

	// Block until the context is cancelled.
	<-ctx.Done()

	close(w.taskCh)

	err := ctx.Err()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

// Add adds a task to the task channel.
func (w *workerManager) Add(task TaskExecutor) {
	w.taskCh <- task
}

// worker is a goroutine that executes tasks from the task channel.
func (w *workerManager) worker(parentCtx context.Context) {
	for {
		select {
		case <-parentCtx.Done():
			// Exit the goroutine when the context is cancelled.
			return
		case task, ok := <-w.taskCh:
			if !ok {
				// Exit the goroutine when the channel is closed.
				return
			}
			ctx, cancel := context.WithTimeout(parentCtx, task.Timeout())
			task.Execute(ctx)
			// Cancel task after the task is done.
			cancel()
		}
	}
}
