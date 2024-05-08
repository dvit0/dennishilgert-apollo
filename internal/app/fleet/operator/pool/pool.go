package pool

import (
	"fmt"
	"sync"

	"github.com/dennishilgert/apollo/internal/app/fleet/operator/runner"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.pool")

type RunnerPool interface {
	Pool() *map[string]map[string]runner.RunnerInstance
	DeepCopy() map[string]map[string]runner.RunnerInstance
	Lock()
	Unlock()
	Add(instance runner.RunnerInstance) error
	Get(functionIdentifier string, runnerUuid string) (runner.RunnerInstance, error)
	GetByFunction(functionIdentifier string) (map[string]runner.RunnerInstance, error)
	Remove(functionIdentifier string, runnerUuid string)
	AvailableRunner(functionIdentifier string) (runner.RunnerInstance, error)
}

type runnerPool struct {
	pool map[string]map[string]runner.RunnerInstance
	lock sync.Mutex
}

// NewRunnerPool creates a new instance of runnerPool.
func NewRunnerPool() RunnerPool {
	return &runnerPool{
		pool: make(map[string]map[string]runner.RunnerInstance),
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
func (r *runnerPool) Pool() *map[string]map[string]runner.RunnerInstance {
	return &r.pool
}

// DeepCopy creates a deep copy of the pool.
func (r *runnerPool) DeepCopy() map[string]map[string]runner.RunnerInstance {
	r.lock.Lock()
	defer r.lock.Unlock()

	copyPool := make(map[string]map[string]runner.RunnerInstance)

	for functionId, runners := range r.pool {
		copyRunners := make(map[string]runner.RunnerInstance)
		for runnerId, runnerInstance := range runners {
			copyRunners[runnerId] = runnerInstance
		}
		copyPool[functionId] = copyRunners
	}

	return copyPool
}

// Add adds a runner instance to the pool.
func (r *runnerPool) Add(instance runner.RunnerInstance) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	functionIdentifier := instance.Config().FunctionIdentifier
	runnerUuid := instance.Config().RunnerUuid

	// Create inner map if it is non existent.
	if r.pool[functionIdentifier] == nil {
		r.pool[functionIdentifier] = make(map[string]runner.RunnerInstance)
	}
	if r.pool[functionIdentifier][runnerUuid] != nil {
		return fmt.Errorf("pool already contains a runner instance with the given uuid: %s", runnerUuid)
	}
	r.pool[functionIdentifier][runnerUuid] = instance
	return nil
}

// Get returns a runner instance by its uuid from the pool.
func (r *runnerPool) Get(functionIdentifier string, runnerUuid string) (runner.RunnerInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	instance := r.pool[functionIdentifier][runnerUuid]
	if instance == nil {
		return nil, fmt.Errorf("requested runner does not exist: %s", runnerUuid)
	}
	return instance, nil
}

// GetByFunction returns all runner instances for a given function identifier.
func (r *runnerPool) GetByFunction(functionIdentifier string) (map[string]runner.RunnerInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	runners := r.pool[functionIdentifier]
	if runners == nil {
		return nil, fmt.Errorf("no runners for given function identifier exist: %s", functionIdentifier)
	}
	return runners, nil
}

// Remove removes a runner instance from the pool.
func (r *runnerPool) Remove(functionIdentifier string, runnerUuid string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.pool[functionIdentifier], runnerUuid)
}

// AvailableRunner returns a available runner instance from the pool.
func (r *runnerPool) AvailableRunner(functionIdentifier string) (runner.RunnerInstance, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	runners := r.pool[functionIdentifier]
	if runners == nil || len(runners) < 1 {
		log.Debugf("pool does not contain runner instances for function identifier: %s", functionIdentifier)
		return nil, fmt.Errorf("pool does not contain runner instances for function identifier: %s", functionIdentifier)
	}
	for _, runnerInstance := range runners {
		if runnerInstance.State() == runner.RunnerStateReady {
			return runnerInstance, nil
		}
	}
	return nil, fmt.Errorf("pool does not contain available runner instance function identifier: %s", functionIdentifier)
}
