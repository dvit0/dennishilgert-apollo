package microvm

import (
	"context"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	log "github.com/sirupsen/logrus"
)

var (
	FirecrackerBinaryPath = "../firecracker"
)

func CreateAndStart(ctx context.Context, opLogger hclog.Logger, params VmParameters) (*Vm, error) {
	vmID := uuid.New()
	params.VmID = vmID

	fcCfg := firecrackerConfig(params)

	vmCtx, vmCtxCancel := context.WithCancel(ctx)

	vmLogger := log.New()
	vmLogger.SetLevel(log.DebugLevel)

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(vmLogger)),
	}

	// if the jailer is used, the final command will be built in NewMachine()
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(FirecrackerBinaryPath).
			WithSocketPath(fcCfg.SocketPath).
			WithStdin(os.Stdin).
			WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}

	opLogger.Debug("creating new machine", "vm-id", vmID)
	machine, err := firecracker.NewMachine(vmCtx, fcCfg, machineOpts...)
	if err != nil {
		return nil, err
	}

	opLogger.Debug("starting machine", "vm-id", vmID)
	if err := machine.Start(vmCtx); err != nil {
		return nil, err
	}

	return &Vm{
		VmID:        vmID,
		VmCtx:       vmCtx,
		VmCtxCancel: vmCtxCancel,
		Machine:     machine,
		IP:          machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP,
	}, nil
}
