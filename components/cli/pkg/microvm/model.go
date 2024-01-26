package firecracker

import (
	"context"
	"net"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type firecrackerInstance struct {
	VmCtx       context.Context
	VmCtxCancel context.CancelFunc
	VmID        string
	Machine     *firecracker.Machine
	IP          net.IP
}

type firecrackerParameters struct {
	VmID            string
	KernelImagePath string
	RootDrivePath   string
	CodeDrivePath   string
	VCpuCount       int64
	MemSizeMib      int64
	Multithreading  bool
}
