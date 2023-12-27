package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func bootVm(ctx context.Context, kernel string, image string, firecrackerBinary string) {
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	devices := make([]models.Drive, 1)
	devices[0] = models.Drive{
		DriveID:      firecracker.String("1"),
		PathOnHost:   &image,
		IsRootDevice: firecracker.Bool(true),
		IsReadOnly:   firecracker.Bool(true),
	}
	fcCfg := firecracker.Config{
		LogLevel:        "debug",
		SocketPath:      "",
		KernelImagePath: kernel,
		KernelArgs:      "console=tty0 reboot=k panic=1 acpi=off pci=off nomodules init=/usr/bin/tini-static -p SIGINT -p SIGTERM -- \"/usr/bin/agent.sh\"",
		Drives:          devices,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:   firecracker.Int64(2),
			CPUTemplate: models.CPUTemplateC3,
			MemSizeMib:  firecracker.Int64(1024),
		},
		NetworkInterfaces: []firecracker.NetworkInterface{
			{
				CNIConfiguration: &firecracker.CNIConfiguration{
					NetworkName: "fcnet",
					IfName:      "eth0",
				},
			},
		},
	}
	machineOpts := []firecracker.Opt{}

	command := firecracker.VMCommandBuilder{}.
		WithBin(firecrackerBinary).
		WithSocketPath(fcCfg.SocketPath).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	machineOpts = append(machineOpts, firecracker.WithProcessRunner(command))

	machine, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		fmt.Printf("Failed to start the vm: %+v\n", err)
		os.Exit(1)
	}

	if err := machine.Start(vmmCtx); err != nil {
		fmt.Printf("Failed to start the vm: %+v\n", err)
		os.Exit(1)
	}
	defer machine.StopVMM()

	if err := machine.Wait(vmmCtx); err != nil {
		fmt.Printf("Failed to start the vm: %+v\n", err)
		os.Exit(1)
	}

	log.Print("Machine was started")
}

func main() {
	bootVm(context.Background(), "../vmlinux-5.10.204", "../", "../firecracker")
}
