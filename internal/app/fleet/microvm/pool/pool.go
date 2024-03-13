package pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/dennishilgert/apollo/internal/app/fleet/microvm/machine"
	"github.com/dennishilgert/apollo/pkg/logger"
)

var log = logger.NewLogger("apollo.manager.pool")

type Pool interface {
	Pool() *map[string]map[string]*machine.MachineInstance
	Lock()
	Unlock()
	Add(instance machine.MachineInstance) error
	Remove(fnId string, vmId string)
	RemoveAndDestroy(fnId string, vmId string) error
	AvailableVm(fnId string) (*machine.MachineInstance, error)
}

type vmPool struct {
	pool map[string]map[string]*machine.MachineInstance
	lock sync.Mutex
}

// NewVmPool returns a new instance of vmPool.
func NewVmPool() Pool {
	return &vmPool{
		pool: make(map[string]map[string]*machine.MachineInstance),
	}
}

// Lock locks the vm pool.
func (v *vmPool) Lock() {
	v.lock.Lock()
}

// Unlock unlocks the vm pool.
func (v *vmPool) Unlock() {
	v.lock.Unlock()
}

// Pool returns the pool with vms inside.
func (m *vmPool) Pool() *map[string]map[string]*machine.MachineInstance {
	return &m.pool
}

// Add adds a machine instance to the pool.
func (m *vmPool) Add(instance machine.MachineInstance) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.pool[instance.Cfg.FnId][instance.Cfg.VmId] != nil {
		return fmt.Errorf("pool already contains a machine instance with the given id: %s", instance.Cfg.VmId)
	}
	m.pool[instance.Cfg.FnId][instance.Cfg.VmId] = &instance
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
		log.Debugf("pool does not contain available machine instance for function id: %s", fnId)
		return nil, nil
	}
	for _, vmInstance := range vms {
		if vmInstance.State() == machine.VmStateReady {
			return vmInstance, nil
		}
	}
	return nil, fmt.Errorf("pool contains machine instances with a state other than available")
}
