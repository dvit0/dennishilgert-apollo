package pool

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/concurrency/worker"
	"github.com/dennishilgert/apollo/pkg/proto/health/v1"
)

type WatchdogOptions struct {
	CheckInterval time.Duration
	WorkerCount   int
}

type RunnerPoolWatchdog interface {
	Run(ctx context.Context) error
}

type runnerPoolWatchdog struct {
	runnerPool    RunnerPool
	worker        worker.WorkerManager
	checkInterval time.Duration
	errCh         chan error
}

// NewRunnerPoolWatchdog create a new Watchdog.
func NewRunnerPoolWatchdog(runnerPool RunnerPool, opts WatchdogOptions) RunnerPoolWatchdog {
	return &runnerPoolWatchdog{
		runnerPool:    runnerPool,
		worker:        worker.NewWorkerManager(opts.WorkerCount),
		checkInterval: opts.CheckInterval,
		errCh:         make(chan error),
	}
}

// Run runs the health checks in a specified time interval.
func (p *runnerPoolWatchdog) Run(ctx context.Context) error {
	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			if err := p.worker.Run(ctx); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context) error {
			ticker := time.NewTicker(p.checkInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Info("shutting down pool watchdog")
					return ctx.Err()
				case <-ticker.C:
					// check the vms health
					p.checkRunners()
				case err := <-p.errCh:
					// log error occured while checking the health
					if err != nil {
						log.Errorf("error while running the watchdog: %v", err)
					}
				}
			}
		},
	)
	return runner.Run(ctx)
}

// checkRunners checks the health status of the runners in the pool.
func (p *runnerPoolWatchdog) checkRunners() {
	p.runnerPool.Lock()
	defer p.runnerPool.Unlock()

	for _, runnerPool := range *p.runnerPool.Pool() {
		for _, target := range runnerPool {

			// building the task for the health check
			task := worker.NewTask[bool](func(ctx context.Context) (bool, error) {
				log.Debugf("checking health status of runner: %s", target.Cfg.RunnerUuid)

				healthStatus, err := target.Health(ctx)
				if err != nil {
					log.Errorf("error while checking health status of runner %s: %v", target.Cfg.RunnerUuid, err)
					return false, err
				}
				if *healthStatus != health.HealthStatus_HEALTHY {
					log.Warnf("runner not healthy: %s", target.Cfg.RunnerUuid)
					return false, nil
				}
				return true, nil
			})

			// add a callback to handle the result of the health check
			task.Callback(func(healthy bool, err error) {
				if !healthy || err != nil {
					// destroy the runner if it's not healthy
					p.runnerPool.RemoveAndDestroy(target.Cfg.FunctionUuid, target.Cfg.RunnerUuid)
				}
				if err != nil {
					p.errCh <- err
				}
			})

			// add the task to the worker queue
			p.worker.Add(task)
		}
	}
}
