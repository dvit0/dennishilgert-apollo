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
	"github.com/dennishilgert/apollo/pkg/proto/shared/v1"
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

type TeardownParams struct {
	FunctionUuid string
	RunnerUuid   string
}

type RunnerInstance interface {
	Config() Config
	State() RunnerState
	SetState(state RunnerState)
	CreateAndStart(ctx context.Context) error
	ShutdownAndDestroy(ctx context.Context) error
	Invoke(ctx context.Context, request *agent.InvokeRequest) (*agent.InvokeResponse, error)
	Health(ctx context.Context) (*health.HealthStatus, error)
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
}

// NewInstance creates a new RunnerInstance.
func NewInstance(ctx context.Context, cfg *Config) (RunnerInstance, error) {
	log.Debugf("validating machine configuration for machine with id: %s", cfg.RunnerUuid)
	if err := validate(cfg); err != nil {
		log.Error("failed to validate machine configuration")
		return nil, err
	}

	fnCtx, fnCtxCancel := context.WithCancel(ctx)

	return &runnerInstance{
		cfg:       cfg,
		ctx:       fnCtx,
		ctxCancel: fnCtxCancel,
		lastUsed:  time.Now(),
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
		panic(fmt.Errorf("failed to create stdout file: %v", err))
	}

	// Stderr will be directed to this file.
	stderr, err := os.OpenFile(r.cfg.StdErrFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to create stderr file: %v", err))
	}

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
		log.Error("failed to create a new firecracker machine")
		return err
	}
	log.Debugf("starting firecracker machine with id: %s", fcCfg.VMID)
	if err := machine.Start(r.ctx); err != nil {
		log.Error("failed to start firecracker machine")
		return err
	}
	r.machine = machine
	r.ip = machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP

	r.SetState(RunnerStateCreated)

	return nil
}

// ShutdownAndDestroy shuts a firecracker machine down and destroys it afterwards.
func (r *runnerInstance) ShutdownAndDestroy(parentCtx context.Context) error {
	log.Debugf("shutting down runner: %s", r.cfg.RunnerUuid)

	r.SetState(RunnerStateShutdown)

	if err := r.machine.Shutdown(parentCtx); err != nil {
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
			if !r.IsRunning() {
				log.Debugf("runner has been shut down gracefully: %s", r.cfg.RunnerUuid)
				return nil
			}
		case <-ctx.Done():
			log.Debugf("force stopping runner: %s", r.cfg.RunnerUuid)

			if err := r.machine.StopVMM(); err != nil {
				log.Error("failed to force stop the runner")
				return err
			} else {
				log.Warnf("runner has been stopped forcefully: %s", r.cfg.RunnerUuid)
				return nil
			}
		}
	}
}

// Initialize initializes and executes the runtime inside the runner.
func (r *runnerInstance) Initialize(ctx context.Context, request *agent.InitializeRuntimeRequest) (*shared.EmptyResponse, error) {
	if !r.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", r.cfg.RunnerUuid)
	}
	apiClient := agent.NewAgentClient(r.clientConn)
	initResponse, err := apiClient.Initialize(ctx, request)
	if err != nil {
		return nil, err
	}
	return initResponse, nil
}

// Invoke invokes the function inside the runner through the agent.
func (r *runnerInstance) Invoke(ctx context.Context, request *agent.InvokeRequest) (*agent.InvokeResponse, error) {
	if !r.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", r.cfg.RunnerUuid)
	}

	r.lastUsed = time.Now()

	apiClient := agent.NewAgentClient(r.clientConn)
	invokeResponse, err := apiClient.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}
	return invokeResponse, nil
}

// Health checks the health status of an agent and returns the result.
func (r *runnerInstance) Health(ctx context.Context) (*health.HealthStatus, error) {
	if !r.ConnectionAlive() {
		return nil, fmt.Errorf("connection dead - failed to connect to agent in runner: %s", r.cfg.RunnerUuid)
	}
	apiClient := health.NewHealthClient(r.clientConn)
	healthStatus, err := apiClient.Status(ctx, &health.HealthStatusRequest{})
	if err != nil {
		return nil, err
	}
	return &healthStatus.Status, nil
}

// Ready makes sure the runner is ready.
func (r *runnerInstance) Ready(ctx context.Context) error {
	if err := r.establishConnection(ctx); err != nil {
		return err
	}
	_, err := r.Initialize(ctx, &agent.InitializeRuntimeRequest{
		Handler:    r.cfg.RuntimeHandler,
		BinaryPath: r.cfg.RuntimeBinaryPath,
		BinaryArgs: r.cfg.RuntimeBinaryArgs,
	})
	if err != nil {
		log.Errorf("error while initializing runtime in runner: %s", r.cfg.RunnerUuid)
		return err
	}
	r.SetState(RunnerStateReady)
	return nil
}

// IsRunning returns if a firecracker machine is running.
func (r *runnerInstance) IsRunning() bool {
	// to check if the machine is running, try to establish a connection via the unix socket
	conn, err := net.Dial("unix", r.machine.Cfg.SocketPath)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// ConnectionAlive returns if a grpc client connection is still alive.
func (r *runnerInstance) ConnectionAlive() bool {
	return ((r.clientConn.GetState() == connectivity.Ready) || (r.clientConn.GetState() == connectivity.Idle))
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
		// try to close the existing conenction and ignore the possible error
		r.clientConn.Close()
	}

	addr := strings.Join([]string{r.ip.String(), fmt.Sprint(r.cfg.AgentApiPort)}, ":")
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
			r.clientConn = conn
			return nil
		} else {
			log.Debugf("failed to establish agent connection in runner: %s - reason: %v", r.cfg.RunnerUuid, err)
		}
		// wait before retrying, but stop if context is done
		select {
		case <-ctx.Done():
			log.Errorf("context done before connection could be established to agent in runner: %s", r.cfg.RunnerUuid)
			return ctx.Err() // send context cancellation error
		case <-time.After(time.Duration(math.Round(1000/retriesPerSecond)) * time.Millisecond): // retry delay
			continue
		}
	}
	return nil
}
