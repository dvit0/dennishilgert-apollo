package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.pool")

type RunnerPool interface {
	Pool() *map[string]map[string]*runner.RunnerInstance
	Lock()
	Unlock()
	Add(instance runner.RunnerInstance) error
	Get(functionUuid string, runnerUuid string) (*runner.RunnerInstance, error)
	Remove(functionUuid string, runnerUuid string)
	RemoveAndDestroy(functionUuid string, runnerUuid string) error
	AvailableRunner(functionUuid string) (*runner.RunnerInstance, error)
}

type runnerPool struct {
	pool map[string]map[string]*runner.RunnerInstance
	lock sync.Mutex
}

// NewRunnerPool returns a new instance of runnerPool.
func NewRunnerPool() RunnerPool {
	return &runnerPool{
		pool: make(map[string]map[string]*runner.RunnerInstance),
	}
}

// Lock locks the runner pool.
func (r *runnerPool) Lock() {
	r.lock.Lock()
}

// Unlock unlocks the runner pool.
func (r *runnerPool) Unlock() {
	r.lock.Unlock()
}

// Pool returns the pool with runners inside.
func (r *runnerPool) Pool() *map[string]map[string]*runner.RunnerInstance {
	return &r.pool
}

// Add adds a runner instance to the pool.
func (r *runnerPool) Add(instance runner.RunnerInstance) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.pool[instance.Cfg.FunctionUuid][instance.Cfg.RunnerUuid] != nil {
		return fmt.Errorf("pool already contains a runner instance with the given id: %s", instance.Cfg.RunnerUuid)
	}
	r.pool[instance.Cfg.FunctionUuid][instance.Cfg.RunnerUuid] = &instance
	return nil
}

// Get returns a runner instance by its uuid from the pool.
func (r *runnerPool) Get(functionUuid string, runnerUuid string) (*runner.RunnerInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	instance := r.pool[functionUuid][runnerUuid]
	if instance == nil {
		return nil, fmt.Errorf("requested runner does not exist: %s", runnerUuid)
	}
	return instance, nil
}

// Remove removes a runner instance from the pool.
func (r *runnerPool) Remove(functionUuid string, runnerUuid string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.pool[functionUuid], runnerUuid)
}

// RemoveAndDestroy removes a runner instance from the pool and destroys it afterwards.
func (r *runnerPool) RemoveAndDestroy(functionUuid string, runnerUuid string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	target := r.pool[functionUuid][runnerUuid]
	if target == nil {
		log.Errorf("failed to find runner in pool: %s", runnerUuid)
		return fmt.Errorf("failed to find runner in pool: %s", runnerUuid)
	}
	r.Remove(functionUuid, runnerUuid)
	if err := target.ShutdownAndDestroy(context.Background()); err != nil {
		log.Errorf("error while shutting down runner: %s", runnerUuid)
		return err
	}
	return nil
}

// AvailableRunner returns a available runner instance from the pool.
func (r *runnerPool) AvailableRunner(functionUuid string) (*runner.RunnerInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	vms := r.pool[functionUuid]
	if vms == nil || len(vms) < 1 {
		log.Debugf("pool does not contain runner instances for function uuid: %s", functionUuid)
		return nil, nil
	}
	for _, runnerInstance := range vms {
		if runnerInstance.State() == runner.RunnerStateReady {
			return runnerInstance, nil
		}
	}
	return nil, fmt.Errorf("pool does not contain available runner instance function uuid: %s", functionUuid)
}
