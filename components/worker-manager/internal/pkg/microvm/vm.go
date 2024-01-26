package microvm

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/google/uuid"
)

type VmState int32

const (
	VmStateAvailable VmState = iota
	VmStateBusy
)

type Vm struct {
	VmCtx       context.Context
	VmCtxCancel context.CancelFunc
	VmID        uuid.UUID
	FnID        uuid.UUID
	Machine     *firecracker.Machine
	IP          net.IP
	state       atomic.Value
}

func (vm *Vm) State() VmState {
	return vm.state.Load().(VmState)
}

func (vm *Vm) SetState(state VmState) {
	vm.state.Store(state)
}

type VmParameters struct {
	VmID            uuid.UUID
	KernelImagePath string
	RootDrivePath   string
	CodeDrivePath   string
	VCpuCount       int64
	MemSizeMib      int64
	Multithreading  bool
}
