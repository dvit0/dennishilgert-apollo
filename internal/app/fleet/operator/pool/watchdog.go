package pool

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	taskRunner "github.com/dennishilgert/apollo/pkg/concurrency/runner"
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
	teardownCh    chan runner.TeardownParams
	errCh         chan error
}

// NewRunnerPoolWatchdog create a new Watchdog.
func NewRunnerPoolWatchdog(runnerPool RunnerPool, runnerTeardownCh chan runner.TeardownParams, opts WatchdogOptions) RunnerPoolWatchdog {
	return &runnerPoolWatchdog{
		runnerPool:    runnerPool,
		worker:        worker.NewWorkerManager(opts.WorkerCount),
		checkInterval: opts.CheckInterval,
		teardownCh:    runnerTeardownCh,
		errCh:         make(chan error),
	}
}

// Run runs the health checks in a specified time interval.
func (p *runnerPoolWatchdog) Run(ctx context.Context) error {
	runner := taskRunner.NewRunnerManager(
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
	runnerPool := *p.runnerPool.Pool()

	for _, runnersByFunction := range runnerPool {
		for _, target := range runnersByFunction {
			if target.Idle() >= target.Config().IdleTtl {
				log.Debugf("runner idle ttl has been reached: %s", target.Config().RunnerUuid)
				p.teardownRunner(target.Config().FunctionUuid, target.Config().RunnerUuid)
				continue
			}

			// building the task for the health check
			task := worker.NewTask[bool](func(ctx context.Context) (bool, error) {
				log.Debugf("checking health status of runner: %s", target.Config().RunnerUuid)

				healthStatus, err := target.Health(ctx)
				if err != nil {
					log.Errorf("error while checking health status of runner %s: %v", target.Config().RunnerUuid, err)
					return false, err
				}
				if *healthStatus != health.HealthStatus_HEALTHY {
					log.Warnf("runner not healthy: %s", target.Config().RunnerUuid)
					return false, nil
				}
				return true, nil
			})

			// add a callback to handle the result of the health check
			task.Callback(func(healthy bool, err error) {
				if !healthy || err != nil {
					if err != nil {
						p.errCh <- err
					}
					log.Debugf("runner is unhealthy: %s", target.Config().RunnerUuid)
					p.teardownRunner(target.Config().FunctionUuid, target.Config().RunnerUuid)
				}
			})

			// add the task to the worker queue
			p.worker.Add(task)
		}
	}
}

// teardownRunner removes the runner from the pool and sends the tear down signal to the runner operator.
func (p *runnerPoolWatchdog) teardownRunner(functionUuid string, runnerUuid string) {
	// remove runner from pool to prevent further usage
	p.runnerPool.Remove(functionUuid, runnerUuid)

	// send runner params to runner operator to tear down the runner
	p.teardownCh <- runner.TeardownParams{
		FunctionUuid: functionUuid,
		RunnerUuid:   runnerUuid,
	}
}
