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

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	agentpb "github.com/dennishilgert/apollo/internal/pkg/proto/agent/v1"
	healthpb "github.com/dennishilgert/apollo/internal/pkg/proto/health/v1"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var log = logger.NewLogger("apollo.manager.runner")

type RunnerState int

const (
	RunnerStateCreated RunnerState = iota
	RunnerStateReady
	RunnerStateReserved
	RunnerStateBusy
	RunnerStateShutdown
)

func (r RunnerState) String() string {
	switch r {
	case RunnerStateCreated:
		return "CREATED"
	case RunnerStateReady:
		return "READY"
	case RunnerStateReserved:
		return "RESERVED"
	case RunnerStateBusy:
		return "BUSY"
	case RunnerStateShutdown:
		return "SHUTDOWN"
	}
	return "UNKNOWN"
}

type RunnerInstance interface {
	Config() Config
	State() RunnerState
	SetState(state RunnerState)
	CreateAndStart(ctx context.Context) error
	ShutdownAndDestroy(ctx context.Context) error
	Invoke(ctx context.Context, request *agentpb.InvokeRequest) (*agentpb.InvokeResponse, error)
	Health(ctx context.Context) (*healthpb.HealthStatus, error)
	AgentReady(err error)
	Ready(ctx context.Context) error
	IsRunning() bool
	ConnectionAlive() bool
	Idle() time.Duration
}

type runnerInstance struct {
	cfg        *Config
	ctx        context.Context
	ctxCancel  context.CancelFunc
	machine    *firecracker.Machine
	ip         net.IP
	lastUsed   time.Time
	clientConn *grpc.ClientConn
	state      atomic.Value
	stdout     *os.File
	stderr     *os.File
	readyCh    chan error
}

// NewInstance creates a new runner instance.
func NewInstance(ctx context.Context, cfg *Config) (RunnerInstance, error) {
	log.Debugf("validating machine configuration for machine with id: %s", cfg.RunnerUuid)
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("machine configuration validation failed: %w", err)
	}

	fnCtx, fnCtxCancel := context.WithCancel(ctx)

	return &runnerInstance{
		cfg:       cfg,
		ctx:       fnCtx,
		ctxCancel: fnCtxCancel,
		lastUsed:  time.Now(),
		readyCh:   make(chan error, 1),
	}, nil
}

// Config returns the configuration of the runner.
func (r *runnerInstance) Config() Config {
	return *r.cfg
}

// State returns the state of the runner.
func (r *runnerInstance) State() RunnerState {
	return r.state.Load().(RunnerState)
}

// SetState sets the state of the runner.
func (r *runnerInstance) SetState(state RunnerState) {
	r.state.Store(state)
}

// CreateAndStart creates and starts a new firecracker machine.
func (r *runnerInstance) CreateAndStart(ctx context.Context) error {
	log = log.WithFields(map[string]any{"runner": r.cfg.RunnerUuid})

	fcCfg := firecrackerConfig(*r.cfg)

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.LogrusEntry()),
	}

	// Stdout will be directed to this file.
	stdout, err := os.OpenFile(r.cfg.StdOutFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open stdout file: %w", err)
	}
	r.stdout = stdout

	// Stderr will be directed to this file.
	stderr, err := os.OpenFile(r.cfg.StdErrFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open stderr file: %w", err)
	}
	r.stderr = stderr

	// If the jailer is used, the final command will be built in firecracker.NewMachine(...).
	// If not, the final command is built here.
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(r.cfg.FirecrackerBinaryPath).
			WithSocketPath(fcCfg.SocketPath).
			WithStdout(stdout).
			WithStderr(stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}
	log.Debugf("creating new firecracker machine with id: %s", fcCfg.VMID)
	machine, err := firecracker.NewMachine(r.ctx, fcCfg, machineOpts...)
	if err != nil {
		return fmt.Errorf("failed to create firecracker machine: %w", err)
	}
	log.Debugf("starting firecracker machine with id: %s", fcCfg.VMID)
	if err := machine.Start(r.ctx); err != nil {
		return fmt.Errorf("failed to start firecracker machine: %w", err)
	}
	r.machine = machine
	r.ip = machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP

	r.SetState(RunnerStateCreated)

	return nil
}

// ShutdownAndDestroy shuts a firecracker machine down and destroys it afterwards.
func (r *runnerInstance) ShutdownAndDestroy(parentCtx context.Context) error {
	log.Debugf("shutting down runner: %s", r.cfg.RunnerUuid)

	// If the runner is already in the shutdown state, return immediately.
	if r.State() == RunnerStateShutdown {
		log.Debugf("runner already in shutdown state: %s", r.cfg.RunnerUuid)
		return nil
	}

	// Close logging files and connection after the machine has been shut down.
	defer func() {
		if r.stdout != nil {
			r.stdout.Close()
		}
		if r.stderr != nil {
			r.stderr.Close()
		}
		if r.clientConn != nil {
			r.clientConn.Close()
		}

		close(r.readyCh)

		r.ctxCancel()
	}()

	r.SetState(RunnerStateShutdown)

	if err := r.machine.Shutdown(parentCtx); err != nil {
		return fmt.Errorf("failed to shut down runner: %w", err)
	}
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	// Check every 500 ms if the machine is still running.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !r.IsRunning() {
				log.Debugf("runner has been shut down gracefully: %s", r.cfg.RunnerUuid)
				return nil
			}
		case <-ctx.Done():
			log.Debugf("graceful stop timeout reached - force stopping runner: %s", r.cfg.RunnerUuid)

			if err := r.machine.StopVMM(); err != nil {
				return fmt.Errorf("failed to stop runner forcefully: %w", err)
			} else {
				log.Warnf("runner has been stopped forcefully: %s", r.cfg.RunnerUuid)
				return nil
			}
		}
	}
}

// Invoke invokes the function inside the runner through the agent.
func (r *runnerInstance) Invoke(ctx context.Context, request *agentpb.InvokeRequest) (*agentpb.InvokeResponse, error) {
	if !r.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", r.cfg.RunnerUuid)
	}

	r.lastUsed = time.Now()

	apiClient := agentpb.NewAgentClient(r.clientConn)
	invokeResponse, err := apiClient.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}
	return invokeResponse, nil
}

// Health checks the health status of an agent and returns the result.
func (r *runnerInstance) Health(ctx context.Context) (*healthpb.HealthStatus, error) {
	if !r.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", r.cfg.RunnerUuid)
	}
	apiClient := healthpb.NewHealthClient(r.clientConn)
	healthStatus, err := apiClient.Status(ctx, &healthpb.HealthStatusRequest{})
	if err != nil {
		return nil, err
	}
	return &healthStatus.Status, nil
}

// AgentReady signalizes that the agent inside the runner is ready to handle requests.
func (r *runnerInstance) AgentReady(err error) {
	r.readyCh <- err
}

// Ready makes sure the runner is ready.
func (r *runnerInstance) Ready(ctx context.Context) error {
	// The agent inside the runner sends a message to the messaging service which will be received
	// by the messaging consumer of the fleet manager. The handler of this topic will then send the error
	// or nil to the readyCh of this runner instance. If an error is sent to the channel, it will be
	// handled here and the runner will be shut down.
	log.Infof("waiting for the agent inside the runner to become ready")

	waitCtx, waitCtxCancel := context.WithTimeout(ctx, 5*time.Second)
	defer waitCtxCancel()

	select {
	case <-waitCtx.Done():
		return fmt.Errorf("timeout while waiting for the agent in runner to become ready")
	case err := <-r.readyCh:
		if err != nil {
			return fmt.Errorf("agent inside runner failed to become ready: %w", err)
		}
	}

	// Establish the gRPC connection to the agent inside the runner.
	if err := r.establishConnection(ctx); err != nil {
		return err
	}

	r.SetState(RunnerStateReady)
	return nil
}

// IsRunning returns if a firecracker machine is running.
func (r *runnerInstance) IsRunning() bool {
	// To check if the machine is running, try to establish a connection via the unix socket.
	conn, err := net.DialTimeout("unix", r.machine.Cfg.SocketPath, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// ConnectionAlive returns if a grpc client connection is still alive.
func (r *runnerInstance) ConnectionAlive() bool {
	if r.clientConn == nil {
		return false
	}
	return (r.clientConn.GetState() == connectivity.Ready) || (r.clientConn.GetState() == connectivity.Idle)
}

// Idle returns for how long the instance is idling.
func (r *runnerInstance) Idle() time.Duration {
	return time.Since(r.lastUsed)
}

// establishConnection establishs a connection to the agent api inside the runner.
func (r *runnerInstance) establishConnection(ctx context.Context) error {
	log.Debugf("establishing a connection to agent in runner: %s", r.cfg.RunnerUuid)

	if r.clientConn != nil {
		if r.ConnectionAlive() {
			log.Infof("connection alive - no need to establish a new connection to agent in runner: %s", r.cfg.RunnerUuid)
			return nil
		}
		// Try to close the existing connection and ignore the possible error.
		r.clientConn.Close()
	}

	addr := strings.Join([]string{r.ip.String(), fmt.Sprint(r.cfg.AgentApiPort)}, ":")
	log.Debugf("connecting to agent with address: %s", addr)

	const retrySeconds = 3     // trying to connect for a period of 3 seconds
	const retriesPerSecond = 2 // trying to connect 2 times per second
	for i := 0; i < (retrySeconds * retriesPerSecond); i++ {
		conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err == nil {
			r.clientConn = conn
			log.Infof("connection to agent in runner established: %s", r.cfg.RunnerUuid)
			return nil
		} else {
			conn.Close()
			log.Errorf("failed to establish agent connection in runner: %s - reason: %v", r.cfg.RunnerUuid, err)
		}
		// Wait before retrying, but stop if context is done.
		select {
		case <-ctx.Done():
			log.Errorf("context done before connection could be established to agent in runner: %s", r.cfg.RunnerUuid)
			return ctx.Err() // Send context cancellation error.
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return fmt.Errorf("failed to establish connection to agent in runner: %s", r.cfg.RunnerUuid)
}
