package main

import (
	"context"
	"fmt"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

func createAndStartVm(ctx context.Context) (*firecrackerInstance, error) {
	vmmId := xid.New().String()

	fcCfg, err := getFirecrackerConfig(vmmId)
	if err != nil {
		log.Errorf("Error: %s", err)
		return nil, err
	}

	logger := log.New()

	log.SetLevel(log.DebugLevel)
	logger.SetLevel(log.DebugLevel)

	vmmCtx, vmmCancel := context.WithCancel(ctx)

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(logger)),
	}

	firecrackerBinary := "../firecracker"

	// if the jailer is used, the final command will be built in NewMachine()
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(firecrackerBinary).
			WithSocketPath(fcCfg.SocketPath).
			WithStdin(os.Stdin).
			WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}

	log.Printf("Building machine %s ...", vmmId)
	machine, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		return nil, fmt.Errorf("Failed creating machine: %s", err)
	}

	log.Printf("Starting machine %s ...", vmmId)
	if err := machine.Start(vmmCtx); err != nil {
		return nil, fmt.Errorf("Failed to start machine: %v", err)
	}
	//defer func() {
	//	if err := machine.StopVMM(); err != nil {
	//		log.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
	//	}
	//}()

	//installSignalHandlers(vmmCtx, machine)

	log.Printf("Machine has been started successfully")
	return &firecrackerInstance{
		vmmCtx:    vmmCtx,
		vmmCancel: vmmCancel,
		vmmId:     vmmId,
		machine:   machine,
		ip:        machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP,
	}, nil
}
