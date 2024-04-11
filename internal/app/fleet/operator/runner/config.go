package runner

import (
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type Config struct {
	WorkerUuid               string
	FunctionUuid             string
	RunnerUuid               string
	HostOsArch               utils.OsArch
	FirecrackerBinaryPath    string
	KernelImagePath          string
	RuntimeDrivePath         string
	RuntimeHandler           string
	RuntimeBinaryPath        string
	RuntimeBinaryArgs        []string
	FunctionDrivePath        string
	SocketPath               string
	LogFilePath              string
	StdOutFilePath           string
	StdErrFilePath           string
	VCpuCount                int
	MemSizeMib               int
	IdleTtl                  time.Duration
	Multithreading           bool
	AgentApiPort             int
	MessagingBoostrapServers string
	LogLevel                 string
}

func validate(cfg *Config) error {
	exists, _ := utils.FileExists(cfg.FirecrackerBinaryPath)
	if !exists {
		return fmt.Errorf("given firecracker binary does not exist")
	}
	exists, _ = utils.FileExists(cfg.KernelImagePath)
	if !exists {
		return fmt.Errorf("given kernel image does not exist")
	}
	exists, _ = utils.FileExists(cfg.RuntimeDrivePath)
	if !exists {
		return fmt.Errorf("given runtime drive does not exist")
	}
	exists, _ = utils.FileExists(cfg.FunctionDrivePath)
	if !exists {
		return fmt.Errorf("given function drive does not exist")
	}
	exists, _ = utils.FileExists(cfg.LogFilePath)
	if !exists {
		return fmt.Errorf("given function log file does not exist")
	}
	exists, _ = utils.FileExists(cfg.StdOutFilePath)
	if !exists {
		return fmt.Errorf("given function stdout log file does not exist")
	}
	exists, _ = utils.FileExists(cfg.StdErrFilePath)
	if !exists {
		return fmt.Errorf("given function stderr log file does not exist")
	}
	if cfg.Multithreading && cfg.HostOsArch != utils.Arch_x86_64 {
		log.Debugf("multithreading is not supported on the host architecture: %s", cfg.HostOsArch.String())
	}
	return nil
}

// firecrackerConfig returns a full configuration for a firecracker micro vm.
func firecrackerConfig(cfg Config) firecracker.Config {
	return firecracker.Config{
		SocketPath:      cfg.SocketPath,
		KernelImagePath: cfg.KernelImagePath,
		// The KernelArgs are re-parsed in the sdk before they are passed to the vm.
		// This means that values like the custom init statement with tini will be
		// mixed up resulting in invalid kernel args.
		// Therefore the firecracker-go-sdk has been forked and modified.
		KernelArgs: NewKernelArgsBuilder().
			WithConsole("ttyS0").
			WithReboot("k").
			WithPanic(1).
			WithPci("off").
			WithNoModules(true).
			WithWorkerUuid(cfg.WorkerUuid).
			WithFunctionUuid(cfg.FunctionUuid).
			WithRunnerUuid(cfg.RunnerUuid).
			WithRuntimeHandler(cfg.RuntimeHandler).
			WithRuntimeBinaryPath(cfg.RuntimeBinaryPath).
			WithRuntimeBinaryArgs(cfg.RuntimeBinaryArgs).
			WithApiPort(cfg.AgentApiPort).
			WithMessagingBootstrapServers(cfg.MessagingBoostrapServers).
			WithLogLevel(cfg.LogLevel).
			WithInit("/usr/bin/tini-static -p SIGINT -p SIGTERM -- /usr/bin/init").
			Build(),
		LogPath: cfg.LogFilePath,
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				PathOnHost:   firecracker.String(cfg.RuntimeDrivePath),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(true),
			},
			{
				DriveID:      firecracker.String("2"),
				PathOnHost:   firecracker.String(cfg.FunctionDrivePath),
				IsRootDevice: firecracker.Bool(false),
				IsReadOnly:   firecracker.Bool(true),
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
