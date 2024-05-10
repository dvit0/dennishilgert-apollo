package runner

import (
	"fmt"
	"strings"

	"github.com/dennishilgert/apollo/internal/pkg/utils"
)

type KernelArgsBuilder interface {
	Build() string
	WithConsole(console string) KernelArgsBuilder
	WithReboot(reboot string) KernelArgsBuilder
	WithPanic(panic int) KernelArgsBuilder
	WithPci(pci string) KernelArgsBuilder
	WithNoModules(noModules bool) KernelArgsBuilder
	WithInit(init string) KernelArgsBuilder
	WithWorkerUuid(workerUuid string) KernelArgsBuilder
	WithFunctionIdentifier(functionIdentifier string) KernelArgsBuilder
	WithRunnerUuid(runnerUuid string) KernelArgsBuilder
	WithRuntimeConfiguration(handler string, binaryPath string, binaryArgs []string, environment []string) KernelArgsBuilder
	WithApiPort(port int) KernelArgsBuilder
	WithMessagingBootstrapServers(servers string) KernelArgsBuilder
	WithLogLevel(logLevel string) KernelArgsBuilder
}

type kernelArgsBuilder struct {
	console                   string
	reboot                    string
	panic                     string
	pci                       string
	nomodules                 string
	init                      string
	workerUuid                string
	functionUuid              string
	runnerUuid                string
	runtimeConfiguration      string
	apiPort                   string
	messagingBootstrapServers string
	logLevel                  string
}

// NewKernelArgsBuilder creates a new instance of KernelArgsBuilder.
func NewKernelArgsBuilder() KernelArgsBuilder {
	return &kernelArgsBuilder{}
}

// Build builds the kernel arguments.
func (c *kernelArgsBuilder) Build() string {
	preBuilt := strings.Join([]string{
		c.console,
		c.reboot,
		c.panic,
		c.pci,
		c.nomodules,
		c.workerUuid,
		c.functionUuid,
		c.runnerUuid,
		c.runtimeConfiguration,
		c.apiPort,
		c.messagingBootstrapServers,
		c.logLevel,
		c.init,
	}, " ")
	return strings.Join(strings.Fields(preBuilt), " ")
}

// WithConsole sets the console argument.
func (c *kernelArgsBuilder) WithConsole(console string) KernelArgsBuilder {
	c.console = fmt.Sprintf("console=%s", console)
	return c
}

// WithReboot sets the reboot argument.
func (c *kernelArgsBuilder) WithReboot(reboot string) KernelArgsBuilder {
	c.reboot = fmt.Sprintf("reboot=%s", reboot)
	return c
}

// WithPanic sets the panic argument.
func (c *kernelArgsBuilder) WithPanic(panic int) KernelArgsBuilder {
	c.panic = fmt.Sprintf("panic=%d", panic)
	return c
}

// WithPci sets the pci argument.
func (c *kernelArgsBuilder) WithPci(pci string) KernelArgsBuilder {
	c.pci = fmt.Sprintf("pci=%s", pci)
	return c
}

// WithNoModules sets the nomodules argument.
func (c *kernelArgsBuilder) WithNoModules(noModules bool) KernelArgsBuilder {
	if noModules {
		c.nomodules = "nomodules"
	}
	return c
}

// WithInit sets the init argument.
func (c *kernelArgsBuilder) WithInit(init string) KernelArgsBuilder {
	c.init = fmt.Sprintf("init=%s", init)
	return c
}

// WithWorkerUuid sets the workerUuid argument.
func (c *kernelArgsBuilder) WithWorkerUuid(workerUuid string) KernelArgsBuilder {
	c.workerUuid = fmt.Sprintf("wkr-uuid=%s", workerUuid)
	return c
}

// WithFunctionIdentifier sets the functionIdentifier argument.
func (c *kernelArgsBuilder) WithFunctionIdentifier(functionIdentifier string) KernelArgsBuilder {
	c.functionUuid = fmt.Sprintf("fn-ident=%s", functionIdentifier)
	return c
}

// WithRunnerUuid sets the runnerUuid argument.
func (c *kernelArgsBuilder) WithRunnerUuid(runnerUuid string) KernelArgsBuilder {
	c.runnerUuid = fmt.Sprintf("rn-uuid=%s", runnerUuid)
	return c
}

// WithRuntimeConfiguration sets the runtime configuration arguments.
func (c *kernelArgsBuilder) WithRuntimeConfiguration(handler string, binaryPath string, binaryArgs []string, environment []string) KernelArgsBuilder {
	runtimeConfiguration := map[string]interface{}{
		"handler":     handler,
		"binaryPath":  binaryPath,
		"binaryArgs":  binaryArgs,
		"environment": environment,
	}
	serializedRuntimeConfiguration, err := utils.SerializeJson(runtimeConfiguration)
	if err != nil {
		log.Errorf("error while serializing runtime configuration: %v", err)
	}
	c.runtimeConfiguration = fmt.Sprintf("rt-cfg=%s", serializedRuntimeConfiguration)
	return c
}

// WithApiPort sets the apiPort argument.
func (c *kernelArgsBuilder) WithApiPort(port int) KernelArgsBuilder {
	c.apiPort = fmt.Sprintf("api-port=%d", port)
	return c
}

// WithMessagingBootstrapServers sets the messagingBootstrapServers argument.
func (c *kernelArgsBuilder) WithMessagingBootstrapServers(servers string) KernelArgsBuilder {
	c.messagingBootstrapServers = fmt.Sprintf("msg-srvs=%s", servers)
	return c
}

// WithLogLevel sets the logLevel argument.
func (c *kernelArgsBuilder) WithLogLevel(logLevel string) KernelArgsBuilder {
	c.logLevel = fmt.Sprintf("log-lvl=%s", logLevel)
	return c
}
