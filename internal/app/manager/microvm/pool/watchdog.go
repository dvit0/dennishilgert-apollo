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

type Watchdog interface {
	Run(ctx context.Context) error
}

type vmPoolWatchdog struct {
	vmPool        Pool
	worker        *worker.WorkerManager
	checkInterval time.Duration
	errCh         chan error
}

// NewPoolWatchdog create a new Watchdog.
func NewVmPoolWatchdog(vmPool Pool, opts WatchdogOptions) Watchdog {
	return &vmPoolWatchdog{
		vmPool:        vmPool,
		worker:        worker.NewWorkerManager(opts.WorkerCount),
		checkInterval: opts.CheckInterval,
		errCh:         make(chan error),
	}
}

// Run runs the health checks in a specified time interval.
func (p *vmPoolWatchdog) Run(ctx context.Context) error {
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
					// context cancelled, return the reason
					log.Info("shutting down pool watchdog")
					return ctx.Err()
				case <-ticker.C:
					// check the vms health
					p.checkVms()
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

// checkVms checks the health status of the vms in the pool.
func (p *vmPoolWatchdog) checkVms() {
	p.vmPool.Lock()
	defer p.vmPool.Unlock()

	for _, fnPool := range *p.vmPool.Pool() {
		for _, target := range fnPool {

			// building the task for the health check
			task := worker.NewTask[bool](func(ctx context.Context) (bool, error) {
				log.Debugf("checking health status of machine: %s", target.Cfg.VmId)

				healthStatus, err := target.Health(ctx)
				if err != nil {
					log.Errorf("error while checking health status of firecracker machine %s: %v", target.Cfg.VmId, err)
					return false, err
				}
				if *healthStatus != health.HealthStatus_HEALTHY {
					log.Warnf("firecracker machine not healthy: %s", target.Cfg.VmId)
					return false, nil
				}
				return true, nil
			})

			// add a callback to handle the result of the health check
			task.Callback(func(healthy bool, err error) {
				if !healthy || err != nil {
					// destroy the firecracker machine if it's not healthy
					p.vmPool.RemoveAndDestroy(target.Cfg.FnId, target.Cfg.VmId)
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
