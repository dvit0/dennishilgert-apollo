package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

var random = rand.New(rand.NewSource(1))

func randomString(length int) string {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = allowed[random.Intn(len(allowed))]
	}
	return string(b)
}

func bootVm(ctx context.Context, kernel string, image string, firecrackerBinary string) {
	vmId := randomString(16)

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
		SocketPath:      filepath.Join("tmp", "fc", vmId+".sock"),
		KernelImagePath: kernel,
		//KernelArgs:      "console=tty0 reboot=k panic=1 acpi=off pci=off nomodules init=/usr/bin/tini-static -p SIGINT -p SIGTERM -- \"/usr/bin/agent.sh\"",
		KernelArgs: "console=tty0 reboot=k panic=1 acpi=off pci=off nomodules",
		Drives:     devices,
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
	log.Print("Starting vm ...")
	bootVm(context.Background(), "../vmlinux-5.10.204", "../ubuntu-22.04.ext4", "../firecracker")
}
