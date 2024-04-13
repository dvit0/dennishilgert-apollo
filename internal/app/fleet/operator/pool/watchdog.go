package pool

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	taskRunner "github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/concurrency/worker"
	healthpb "github.com/dennishilgert/apollo/pkg/proto/health/v1"
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
	teardownCh    chan runner.RunnerInstance
	errCh         chan error
}

// NewRunnerPoolWatchdog create a new Watchdog.
func NewRunnerPoolWatchdog(runnerPool RunnerPool, runnerTeardownCh chan runner.RunnerInstance, opts WatchdogOptions) RunnerPoolWatchdog {
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
					if err != nil {
						log.Errorf("error while checking the runners: %v", err)
					}
				}
			}
		},
	)
	return runner.Run(ctx)
}

// checkRunners checks the health status of the runners in the pool.
func (p *runnerPoolWatchdog) checkRunners() {
	// To iterate over the runner pool, we need to create a deep copy of the pool.
	// This is necessary because the pool is locked during the iteration and
	// we don't want to block the pool for a long time.
	runnerPool := p.runnerPool.DeepCopy()

	for _, runnersByFunction := range runnerPool {
		for _, runnerInstance := range runnersByFunction {
			if (runnerInstance.State() == runner.RunnerStateCreated) || (runnerInstance.State() == runner.RunnerStateShutdown) {
				log.Debugf("skipping runner health check because of state: %s", runnerInstance.State().String())
				continue
			}
			if runnerInstance.Idle() >= runnerInstance.Config().IdleTtl {
				log.Debugf("runner idle ttl has been reached: %s", runnerInstance.Config().RunnerUuid)
				p.teardownRunner(runnerInstance)
				continue
			}

			// building the task for the health check
			task := p.createHealthCheckTask(runnerInstance)

			// add the task to the worker queue
			p.worker.Add(task)
		}
	}
}

// createHealthCheckTask creates a health check task for the given runner instance.
func (p *runnerPoolWatchdog) createHealthCheckTask(runnerInstance runner.RunnerInstance) *worker.Task[bool] {
	task := worker.NewTask(func(ctx context.Context) (bool, error) {
		log.Debugf("checking health status of runner: %s", runnerInstance.Config().RunnerUuid)

		healthStatus, err := runnerInstance.Health(ctx)
		if err != nil {
			log.Errorf("error while checking health status of runner: %s - reason: %v", runnerInstance.Config().RunnerUuid, err)
			return false, err
		}
		if *healthStatus != healthpb.HealthStatus_HEALTHY {
			log.Warnf("runner not healthy: %s", runnerInstance.Config().RunnerUuid)
			return false, nil
		}
		return true, nil
	}, 5*time.Second)

	task.Callback(func(healthy bool, err error) {
		if !healthy || err != nil {
			if err != nil {
				p.errCh <- err
			}
			log.Debugf("runner is unhealthy - tear down: %s", runnerInstance.Config().RunnerUuid)
			p.teardownRunner(runnerInstance)
		}
	})
	return task
}

// teardownRunner removes the runner from the pool and sends the tear down signal to the runner operator.
func (p *runnerPoolWatchdog) teardownRunner(runnerInstance runner.RunnerInstance) {
	// Remove runner from pool to prevent further usage.
	p.runnerPool.Remove(runnerInstance.Config().FunctionUuid, runnerInstance.Config().RunnerUuid)

	// Send runner params to runner operator to tear down the runner.
	p.teardownCh <- runnerInstance
}
