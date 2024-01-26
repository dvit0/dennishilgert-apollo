package firecracker

import (
	"context"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/hashicorp/go-hclog"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

var (
	FirecrackerBinaryPath = "../firecracker"
)

func CreateAndStart(ctx context.Context, opLogger hclog.Logger, params firecrackerParameters) (*firecrackerInstance, error) {
	vmID := xid.New().String()
	params.VmID = vmID

	fcCfg := firecrackerConfig(params)

	vmCtx, vmCtxCancel := context.WithCancel(ctx)
	defer vmCtxCancel()

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

	machine, err := firecracker.NewMachine(vmCtx, fcCfg, machineOpts...)
	if err != nil {

	}

	if err := machine.Start(vmCtx); err != nil {

	}

	return &firecrackerInstance{
		VmID:        vmID,
		VmCtx:       vmCtx,
		VmCtxCancel: vmCtxCancel,
		Machine:     machine,
		IP:          machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP,
	}, nil
}
