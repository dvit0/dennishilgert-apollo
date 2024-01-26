package pool

import (
	"apollo/worker-manager/internal/pkg/microvm"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type VmPool struct {
	lock sync.Mutex

	vms map[uuid.UUID]map[uuid.UUID]*microvm.Vm
}

func NewVmPool() *VmPool {
	return &VmPool{
		vms: make(map[uuid.UUID]map[uuid.UUID]*microvm.Vm),
	}
}

func (p *VmPool) AddVm(vm *microvm.Vm) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.vms[vm.FnID][vm.VmID] != nil {
		return fmt.Errorf("pool already contains a vm with the id %s", vm.VmID)
	}
	p.vms[vm.FnID][vm.VmID] = vm
	return nil
}

func (p *VmPool) RemoveVm(fnID uuid.UUID, vmID uuid.UUID) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.vms[fnID], vmID)
}

func (p *VmPool) AvailableVm(fnID uuid.UUID) (*microvm.Vm, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	vms := p.vms[fnID]
	if vms == nil || len(vms) < 1 {
		return nil, fmt.Errorf("pool does not contain available vm for function id %s", fnID)
	}
	for _, vm := range vms {
		if vm.State() == microvm.VmStateAvailable {
			delete(p.vms[fnID], vm.VmID)
			return vm, nil
		}
	}
	return nil, fmt.Errorf("pool contains vms with a state other than available")
}

func (p *VmPool) AvailableVms(fnID uuid.UUID) ([]*microvm.Vm, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	vms := p.vms[fnID]
	if vms == nil || len(vms) < 1 {
		return nil, fmt.Errorf("pool does not contain available vm for function id %s", fnID)
	}
	var availableVms []*microvm.Vm
	for _, vm := range vms {
		if vm.State() == microvm.VmStateAvailable {
			availableVms = append(availableVms, vm)
		}
	}
	if len(availableVms) < 1 {
		return nil, fmt.Errorf("pool contains vms with a state other than available")
	}
	return availableVms, nil
}
