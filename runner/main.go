package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

var random = rand.New(rand.NewSource(1))

func randomString(length int) string {
	const allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = allowed[random.Intn(len(allowed))]
	}
	return string(b)
}

func main() {
	log.Print("Starting vm ...")
	ctx, cancel := context.WithCancel(context.Background())
	bootVm(ctx, "../vmlinux-5.10.204", "../ubuntu-22.04.ext4", "../firecracker")
	defer cancel()

	installSignalHandlers()
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
		SocketPath:      filepath.Join(string(os.PathSeparator), "tmp", "fc", vmId+".sock"),
		KernelImagePath: kernel,
		//KernelArgs:      "console=tty0 reboot=k panic=1 acpi=off pci=off nomodules init=/usr/bin/tini-static -p SIGINT -p SIGTERM -- \"/usr/bin/agent.sh\"",
		KernelArgs: "console=tty0 reboot=k panic=1 acpi=off pci=off i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd nomodules",
		Drives:     devices,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:       firecracker.Int64(2),
			MemSizeMib:      firecracker.Int64(1024),
			Smt:             firecracker.Bool(true),
			TrackDirtyPages: *firecracker.Bool(false),
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
		vmmCancel()
		os.Exit(1)
	}

	if err := machine.Start(vmmCtx); err != nil {
		vmmCancel()
		fmt.Printf("Failed to start the vm: %+v\n", err)
		os.Exit(1)
	}

	log.Print("Machine was started")
}

func installSignalHandlers() {
	go func() {
		// Clear some default handlers installed by the firecracker SDK:
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				log.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				os.Exit(0)
			case s == syscall.SIGQUIT:
				log.Printf("Caught signal: %s, forcing shutdown", s.String())
				os.Exit(0)
			}
		}
	}()
}
