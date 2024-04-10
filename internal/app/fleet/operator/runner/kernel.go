package runner

import (
	"fmt"
	"strings"
)

type KernelArgsBuilder interface {
	Build() string
	WithConsole(console string) KernelArgsBuilder
	WithReboot(reboot string) KernelArgsBuilder
	WithPanic(panic int) KernelArgsBuilder
	WithPci(pci string) KernelArgsBuilder
	WithNoModules(noModules bool) KernelArgsBuilder
	WithInit(init string) KernelArgsBuilder
	WithFunctionUuid(functionUuid string) KernelArgsBuilder
	WithRunnerUuid(runnerUuid string) KernelArgsBuilder
	WithRuntimeHandler(handler string) KernelArgsBuilder
	WithRuntimeBinaryPath(binaryPath string) KernelArgsBuilder
	WithRuntimeBinaryArgs(binaryArgs []string) KernelArgsBuilder
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
	functionUuid              string
	runnerUuid                string
	runtimeHandler            string
	runtimeBinaryPath         string
	runtimeBinaryArgs         string
	apiPort                   string
	messagingBootstrapServers string
	logLevel                  string
}

func NewKernelArgsBuilder() KernelArgsBuilder {
	return &kernelArgsBuilder{}
}

func (c *kernelArgsBuilder) Build() string {
	preBuilt := strings.Join([]string{
		c.console,
		c.reboot,
		c.panic,
		c.pci,
		c.nomodules,
		c.functionUuid,
		c.runnerUuid,
		c.runtimeHandler,
		c.runtimeBinaryPath,
		c.runtimeBinaryArgs,
		c.init,
	}, " ")
	return strings.Join(strings.Fields(preBuilt), " ")
}

func (c *kernelArgsBuilder) WithConsole(console string) KernelArgsBuilder {
	c.console = fmt.Sprintf("console=%s", console)
	return c
}

func (c *kernelArgsBuilder) WithReboot(reboot string) KernelArgsBuilder {
	c.reboot = fmt.Sprintf("reboot=%s", reboot)
	return c
}

func (c *kernelArgsBuilder) WithPanic(panic int) KernelArgsBuilder {
	c.panic = fmt.Sprintf("panic=%d", panic)
	return c
}

func (c *kernelArgsBuilder) WithPci(pci string) KernelArgsBuilder {
	c.pci = fmt.Sprintf("pci=%s", pci)
	return c
}

func (c *kernelArgsBuilder) WithNoModules(noModules bool) KernelArgsBuilder {
	if noModules {
		c.nomodules = "nomodules"
	}
	return c
}

func (c *kernelArgsBuilder) WithInit(init string) KernelArgsBuilder {
	c.init = fmt.Sprintf("init=%s", init)
	return c
}

func (c *kernelArgsBuilder) WithFunctionUuid(functionUuid string) KernelArgsBuilder {
	c.functionUuid = fmt.Sprintf("fn-uuid=%s", functionUuid)
	return c
}

func (c *kernelArgsBuilder) WithRunnerUuid(runnerUuid string) KernelArgsBuilder {
	c.runnerUuid = fmt.Sprintf("rn-uuid=%s", runnerUuid)
	return c
}

func (c *kernelArgsBuilder) WithRuntimeHandler(handler string) KernelArgsBuilder {
	c.runtimeHandler = fmt.Sprintf("rt-hdlr=%s", handler)
	return c
}

func (c *kernelArgsBuilder) WithRuntimeBinaryPath(binaryPath string) KernelArgsBuilder {
	c.runtimeBinaryPath = fmt.Sprintf("rt-bin-path=%s", binaryPath)
	return c
}

func (c *kernelArgsBuilder) WithRuntimeBinaryArgs(binaryArgs []string) KernelArgsBuilder {
	c.runtimeBinaryArgs = fmt.Sprintf("rt-bin-args=\"%s\"", strings.Join(binaryArgs, " "))
	return c
}

func (c *kernelArgsBuilder) WithApiPort(port int) KernelArgsBuilder {
	c.apiPort = fmt.Sprintf("api-port=%d", port)
	return c
}

func (c *kernelArgsBuilder) WithMessagingBootstrapServers(servers string) KernelArgsBuilder {
	c.messagingBootstrapServers = fmt.Sprintf("msg-srvs=%s", servers)
	return c
}

func (c *kernelArgsBuilder) WithLogLevel(logLevel string) KernelArgsBuilder {
	c.logLevel = fmt.Sprintf("log-lvl=%s", logLevel)
	return c
}
