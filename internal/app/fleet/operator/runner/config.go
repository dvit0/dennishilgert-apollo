package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func validate(cfg *Config) error {
	if utils.FileExists(cfg.FirecrackerBinaryPath) == nil {
		return fmt.Errorf("file given for the firecracker binary does not exist")
	}
	if utils.FileExists(cfg.KernelImagePath) == nil {
		return fmt.Errorf("file given for the kernel image does not exist")
	}
	if utils.FileExists(cfg.RootDrivePath) == nil {
		return fmt.Errorf("file for the root drive does not exist")
	}
	if utils.FileExists(cfg.CodeDrivePath) == nil {
		return fmt.Errorf("file for the code drive does not exist")
	}
	if cfg.Multithreading && cfg.HostOsArch != utils.Arch_x86_64 {
		log.Warnf("multithreading is not supported on the host architecture: %s", cfg.HostOsArch.String())
	}
	return nil
}

// firecrackerConfig returns a full configuration for a firecracker micro vm.
func firecrackerConfig(cfg Config) firecracker.Config {
	socket := socketPath(cfg.RunnerUuid)

	return firecracker.Config{
		SocketPath:      socket,
		KernelImagePath: cfg.KernelImagePath,
		// The KernelArgs are re-parsed in the sdk before they are passed to the vm.
		// This means that values like the custom init statement with tini will be
		// mixed up resulting in invalid kernel args.
		// Therefore the firecracker-go-sdk has been forked and modified which needs
		// to be used for this.
		KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off nomodules init=/usr/bin/tini-static -p SIGINT -p SIGTERM -- /usr/bin/init",
		LogPath:    fmt.Sprintf("%s.log", socket),
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				PathOnHost:   firecracker.String(cfg.RootDrivePath),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(true),
			},
			{
				DriveID:      firecracker.String("2"),
				PathOnHost:   firecracker.String(cfg.CodeDrivePath),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(false),
			},
		},
		NetworkInterfaces: []firecracker.NetworkInterface{{
			CNIConfiguration: &firecracker.CNIConfiguration{
				NetworkName: "fcnet",
				IfName:      "veth0",
				VMIfName:    "eth0",
				BinPath:     []string{"/opt/cni/bin"},
				ConfDir:     "/etc/cni/conf.d",
				CacheDir:    "/var/lib/cni",
			},
		}},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:       firecracker.Int64(int64(cfg.VCpuCount)),
			MemSizeMib:      firecracker.Int64(int64(cfg.MemSizeMib)),
			Smt:             firecracker.Bool(cfg.Multithreading),
			TrackDirtyPages: firecracker.Bool(false),
		},
		VMID: cfg.RunnerUuid,
	}
}

func socketPath(runnerUuid string) string {
	filename := strings.Join([]string{".firecracker.sock", strconv.Itoa(os.Getpid()), runnerUuid}, "-")
	dir := os.TempDir()

	return filepath.Join(dir, filename)
}
