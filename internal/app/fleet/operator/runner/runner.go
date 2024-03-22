package runner

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/agent/v1"
	"github.com/dennishilgert/apollo/pkg/proto/health/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var log = logger.NewLogger("apollo.manager.runner")

type RunnerState int32

const (
	RunnerStateCreated RunnerState = iota
	RunnerStateReady
	RunnerStateReserved
	RunnerStateBusy
	RunnerStateShutdown
)

type Config struct {
	FunctionUuid          string
	RunnerUuid            string
	HostOsArch            utils.OsArch
	FirecrackerBinaryPath string
	KernelImagePath       string
	RootDrivePath         string
	CodeDrivePath         string
	VCpuCount             int
	MemSizeMib            int
	Multithreading        bool
	AgentApiPort          int
}

type Instance interface {
	Config() Config
	State() RunnerState
	SetState(state RunnerState)
	CreateAndStart(ctx context.Context) error
	ShutdownAndDestroy(ctx context.Context) error
}

type RunnerInstance struct {
	Cfg        *Config
	Ctx        context.Context
	CtxCancel  context.CancelFunc
	Machine    *firecracker.Machine
	Ip         net.IP
	ClientConn *grpc.ClientConn
	state      atomic.Value
}

// NewInstance create a new RunnerInstance.
func NewInstance(ctx context.Context, cfg *Config) Instance {
	fnCtx, fnCtxCancel := context.WithCancel(ctx)
	return &RunnerInstance{
		Cfg:       cfg,
		Ctx:       fnCtx,
		CtxCancel: fnCtxCancel,
	}
}

// Config returns the configuration of the runner.
func (m *RunnerInstance) Config() Config {
	return *m.Cfg
}

// State returns the state of the runner.
func (m *RunnerInstance) State() RunnerState {
	return m.state.Load().(RunnerState)
}

// SetState sets the state of the runner.
func (m *RunnerInstance) SetState(state RunnerState) {
	m.state.Store(state)
}

// CreateAndStart creates and starts a new firecracker machine.
func (m *RunnerInstance) CreateAndStart(ctx context.Context) error {
	log = log.WithFields(map[string]any{"runner": m.Cfg.RunnerUuid})

	log.Debugf("validating machine configuration for machine with id: %s", m.Cfg.RunnerUuid)
	if err := validate(m.Cfg); err != nil {
		log.Error("failed to validate machine configuration")
		return err
	}
	fcCfg := firecrackerConfig(*m.Cfg)

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.LogrusEntry()),
	}
	// If the jailer is used, the final command will be built in firecracker.NewMachine(...).
	// If not, the final command is built here.
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(m.Cfg.FirecrackerBinaryPath).
			WithSocketPath(fcCfg.SocketPath).
			WithStdin(os.Stdin).
			WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}
	log.Debugf("creating new firecracker machine with id: %s", fcCfg.VMID)
	machine, err := firecracker.NewMachine(m.Ctx, fcCfg, machineOpts...)
	if err != nil {
		log.Error("failed to create a new firecracker machine")
		return err
	}
	log.Debugf("starting firecracker machine with id: %s", fcCfg.VMID)
	if err := machine.Start(m.Ctx); err != nil {
		log.Error("failed to start firecracker machine")
		return err
	}
	m.Machine = machine
	m.Ip = machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP

	m.SetState(RunnerStateCreated)

	return nil
}

// ShutdownAndDestroy shuts a firecracker machine down and destroys it afterwards.
func (m *RunnerInstance) ShutdownAndDestroy(parentCtx context.Context) error {
	log.Debugf("shutting down runner: %s", m.Cfg.RunnerUuid)

	m.SetState(RunnerStateShutdown)

	if err := m.Machine.Shutdown(parentCtx); err != nil {
		return err
	}
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// check every 500ms if the machine is still running
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.IsRunning() {
				log.Debugf("runner has been shut down gracefully: %s", m.Cfg.RunnerUuid)
				return nil
			}
		case <-ctx.Done():
			log.Debugf("force stopping runner: %s", m.Cfg.RunnerUuid)

			if err := m.Machine.StopVMM(); err != nil {
				log.Error("failed to force stop the runner")
				return err
			} else {
				log.Warnf("runner has been stopped forcefully: %s", m.Cfg.RunnerUuid)
				return nil
			}
		}
	}
}

// Ready makes sure the runner is ready.
func (m *RunnerInstance) Ready(ctx context.Context) error {
	if err := m.establishConnection(ctx); err != nil {
		return err
	}

	m.SetState(RunnerStateReady)
	return nil
}

// IsRunning returns if a firecracker machine is running.
func (m *RunnerInstance) IsRunning() bool {
	// to check if the machine is running, try to establish a connection via the unix socket
	conn, err := net.Dial("unix", m.Machine.Cfg.SocketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ConnectionAlive returns if a grpc client connection is still alive.
func (m *RunnerInstance) ConnectionAlive() bool {
	return (m.ClientConn.GetState() == connectivity.Ready || m.ClientConn.GetState() == connectivity.Idle)
}

// Invoke invokes the function inside the runner through the agent.
func (m *RunnerInstance) Invoke(ctx context.Context, request *agent.InvokeRequest) (*agent.InvokeResponse, error) {
	if !m.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connecto to agent in runner: %s", m.Cfg.RunnerUuid)
	}
	apiClient := agent.NewAgentClient(m.ClientConn)
	invokeResponse, err := apiClient.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}
	return invokeResponse, nil
}

// Health checks the health status of an agent and returns the result.
func (m *RunnerInstance) Health(ctx context.Context) (*health.HealthStatus, error) {
	if !m.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", m.Cfg.RunnerUuid)
	}
	apiClient := health.NewHealthClient(m.ClientConn)
	healthStatus, err := apiClient.Status(ctx, &health.HealthStatusRequest{})
	if err != nil {
		return nil, err
	}
	return &healthStatus.Status, nil
}

// establishConnection establishs a connection to the agent api inside the runner.
func (m *RunnerInstance) establishConnection(ctx context.Context) error {
	log.Debugf("establishing a connection to agent in runner: %s", m.Cfg.RunnerUuid)

	if m.ClientConn != nil {
		if m.ConnectionAlive() {
			log.Infof("connection alive - no need to establish a new connection to agent in runner: %s", m.Cfg.RunnerUuid)
			return nil
		}
		// try to close the existing conenction and ignore the possible error
		m.ClientConn.Close()
	}

	addr := strings.Join([]string{m.Ip.String(), fmt.Sprint(m.Cfg.AgentApiPort)}, ":")
	log.Debugf("connecting to agent with address: %s", addr)

	// As there are frequent requests made to the agent api, a keep-alive connection for the whole vm time-to-live is used.
	// The connection is saved in the object of the vm and used for every request to the agent api.
	keepAliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second, // time after which a ping is sent if no activity; grpc defined min value is 10s
		Timeout:             3 * time.Second,  // time after which the connection is closed if no ack for ping
		PermitWithoutStream: true,             // allows pings even when there are no active streams
	}

	const retrySeconds = 5     // trying to connect for a period of 5 seconds
	const retriesPerSecond = 4 // trying to connect 5 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepAliveParams))
		if err == nil {
			m.ClientConn = conn
			return err
		}
		// wait before retrying, but stop if context is done
		select {
		case <-ctx.Done():
			log.Errorf("failed to establish a connection to the agent in runner: %s", m.Cfg.RunnerUuid)
			return ctx.Err() // send context cancellation error
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return nil
}
