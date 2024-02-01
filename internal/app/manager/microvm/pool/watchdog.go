package pool

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/pkg/concurrency/worker"
	"github.com/dennishilgert/apollo/pkg/proto/health/v1"
)

type Watchdog interface {
	Run(ctx context.Context) error
}

type poolWatchdog struct {
	vmPool *vmPool
	worker *worker.WorkerManager
	errCh  chan error
}

// NewPoolWatchdog create a new Watchdog.
func NewPoolWatchdog(vmPool *vmPool) Watchdog {
	return &poolWatchdog{
		vmPool: vmPool,
		worker: worker.NewWorkerManager(10),
		errCh:  make(chan error),
	}
}

// Run runs the health checks in a specified time interval.
func (p *poolWatchdog) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.vmPool.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkVms()
		case <-ctx.Done():
			return ctx.Err()
		case err := <-p.errCh:
			if err != nil {
				log.Errorf("error while running the watchdog: %v", err)
			}
		}
	}
}

// checkVms checks the health status of the vms in the pool.
func (p *poolWatchdog) checkVms() {
	p.vmPool.lock.Lock()
	defer p.vmPool.lock.Unlock()

	for _, fnPool := range *p.vmPool.Pool() {
		for _, target := range fnPool {

			// building the task for the health check
			task := worker.NewTask[bool](func(ctx context.Context) (bool, error) {
				log.Debugf("checking health status of machine: %s", target.Cfg.VmId.String())

				healthStatus, err := target.Health(ctx)
				if err != nil {
					log.Errorf("error while checking health status of firecracker machine %s: %v", target.Cfg.VmId.String(), err)
					return false, err
				}
				if *healthStatus != health.HealthStatus_HEALTHY {
					log.Warnf("firecracker machine not healthy: %s", target.Cfg.VmId.String())
					return false, nil
				}
				return true, nil
			})

			// add a callback to handle the result of the health check
			task.Callback(func(healthy bool, err error) {
				if !healthy || err != nil {
					// destroy the firecracker machine if it is not healthy
					p.vmPool.RemoveAndDestroy(target.Cfg.FnId.String(), target.Cfg.VmId.String())
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
