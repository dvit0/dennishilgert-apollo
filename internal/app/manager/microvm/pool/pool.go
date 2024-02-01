package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dennishilgert/apollo/internal/app/manager/microvm/machine"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.pool")

type Options struct {
	WatchdogCheckInterval time.Duration
	WatchdogWorkerCount   int
}

type Pool interface {
	Pool() *map[string]map[string]*machine.MachineInstance
	Add(instance machine.MachineInstance) error
	Remove(fnId string, vmId string)
	RemoveAndDestroy(fnId string, vmId string) error
	AvailableVm(fnId string) (*machine.MachineInstance, error)
}

type vmPool struct {
	watchdogCheckInterval time.Duration
	watchdogWorkerCount   int
	watchdog              Watchdog
	pool                  map[string]map[string]*machine.MachineInstance
	lock                  sync.Mutex
}

// NewVmPool returns a new instance of vmPool.
func NewVmPool(opts Options) Pool {
	return &vmPool{
		watchdogCheckInterval: opts.WatchdogCheckInterval,
		watchdogWorkerCount:   opts.WatchdogWorkerCount,
		pool:                  make(map[string]map[string]*machine.MachineInstance),
	}
}

// Init initializes a new pool by setting up its watchdog.
func (v *vmPool) Init() error {
	v.watchdog = NewPoolWatchdog(v)
	return nil
}

// Pool returns the pool with vms inside.
func (m *vmPool) Pool() *map[string]map[string]*machine.MachineInstance {
	return &m.pool
}

// Add adds a machine instance to the pool.
func (m *vmPool) Add(instance machine.MachineInstance) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.pool[instance.Cfg.FnId.String()][instance.Cfg.VmId.String()] != nil {
		return fmt.Errorf("pool already contains a machine instance with the given id: %s", instance.Cfg.VmId.String())
	}
	m.pool[instance.Cfg.FnId.String()][instance.Cfg.VmId.String()] = &instance
	return nil
}

// Remove removes a machine instance from the pool.
func (m *vmPool) Remove(fnId string, vmId string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.pool[fnId], vmId)
}

// RemoveAndDestroy removes a machine instance from the pool and destroys it afterwards.
func (m *vmPool) RemoveAndDestroy(fnId string, vmId string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	target := m.pool[fnId][vmId]
	if target == nil {
		log.Errorf("failed to find firecracker machine in pool: %s", vmId)
		return fmt.Errorf("failed to find firecracker machine in pool: %s", vmId)
	}
	m.Remove(fnId, vmId)
	if err := target.ShutdownAndDestroy(context.Background()); err != nil {
		log.Errorf("error while shutting down firecracker machine: %s", vmId)
		return err
	}
	return nil
}

// AvailableVm returns a available machine instance from the pool.
func (m *vmPool) AvailableVm(fnId string) (*machine.MachineInstance, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	vms := m.pool[fnId]
	if vms == nil || len(vms) < 1 {
		return nil, fmt.Errorf("pool does not contain available machine instance for function id: %s", fnId)
	}
	for _, vmInstance := range vms {
		if vmInstance.State() == machine.VmStateReady {
			return vmInstance, nil
		}
	}
	return nil, fmt.Errorf("pool contains machine instances with a state other than available")
}
