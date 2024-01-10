package kernel

import (
	"apollo/agent/pkg/kernel/ip"
	"strings"

	"github.com/hashicorp/go-hclog"
)

type DefaultKernelArgs struct {
	Ip *ip.NetConfig
}

func ReadCmdLine(logger hclog.Logger) (*DefaultKernelArgs, error) {
	kernelArgs := DefaultKernelArgs{}
	// cmdLine, err := os.ReadFile("/proc/cmdline")
	// if err != nil {
	// 	logger.Error("failed to read kernel command line", "reason", err)
	// 	return nil, err
	// }

	// use this line only for local development
	cmdLine := "ip=10.0.0.2::10.0.0.1:255.255.255.0::eth0:off:1.1.1.1::"
	logger.Debug("kernel command line", "cmdline", cmdLine)

	kernelArgsString := string(cmdLine)
	for _, arg := range strings.Fields(kernelArgsString) {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		switch key {
		case "ip":
			ipArg, err := ip.ParseFromString(value)
			if err != nil {
				return nil, err
			}
			kernelArgs.Ip = ipArg
		}
	}

	return &kernelArgs, nil
}
