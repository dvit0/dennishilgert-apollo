package microvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func firecrackerConfig(params VmParameters) firecracker.Config {
	socket := socketPath(params.VmID.String())

	return firecracker.Config{
		SocketPath:      socket,
		KernelImagePath: params.KernelImagePath,
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
				PathOnHost:   firecracker.String(params.RootDrivePath),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(true),
			},
			{
				DriveID:      firecracker.String("2"),
				PathOnHost:   firecracker.String(params.CodeDrivePath),
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
			VcpuCount:       firecracker.Int64(params.VCpuCount),
			MemSizeMib:      firecracker.Int64(params.MemSizeMib),
			Smt:             firecracker.Bool(params.Multithreading),
			TrackDirtyPages: firecracker.Bool(false),
		},
		VMID: params.VmID.String(),
	}
}

func socketPath(vmID string) string {
	filename := strings.Join([]string{".firecracker.sock", strconv.Itoa(os.Getpid()), vmID}, "-")
	dir := os.TempDir()

	return filepath.Join(dir, filename)
}
