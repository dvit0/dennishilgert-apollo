package machine

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/health/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var log = logger.NewLogger("apollo.manager.machine")

type VmState int32

const (
	VmStateAvailable VmState = iota
	VmStateBusy
)

type Config struct {
	FnId                  uuid.UUID
	VmId                  uuid.UUID
	HostOsArch            utils.OsArch
	FirecrackerBinaryPath string
	KernelImagePath       string
	RootDrivePath         string
	CodeDrivePath         string
	VCpuCount             int64
	MemSizeMib            int64
	Multithreading        bool
}

type Instance interface {
	State() VmState
	SetState(state VmState)
	CreateAndStart(ctx context.Context) error
	ShutdownAndDestroy(ctx context.Context) error
}

type MachineInstance struct {
	Cfg        *Config
	Ctx        context.Context
	CtxCancel  context.CancelFunc
	Machine    *firecracker.Machine
	Ip         net.IP
	ClientConn *grpc.ClientConn
	state      atomic.Value
}

func NewMachineInstance(ctx context.Context, cfg *Config) Instance {
	fnCtx, fnCtxCancel := context.WithCancel(ctx)
	return &MachineInstance{
		Cfg:       cfg,
		Ctx:       fnCtx,
		CtxCancel: fnCtxCancel,
	}
}

func (m *MachineInstance) State() VmState {
	return m.state.Load().(VmState)
}

func (m *MachineInstance) SetState(state VmState) {
	m.state.Store(state)
}

// CreateAndStart creates and starts a new firecracker machine.
func (m *MachineInstance) CreateAndStart(ctx context.Context) error {
	log = log.WithFields(map[string]any{"vm-id": m.Cfg.VmId.String()})

	log.Debugf("validating machine configuration for machine with id: %s", m.Cfg.VmId.String())
	if err := validate(m.Cfg); err != nil {
		log.Errorf("failed to validate machine configuration: %v", err)
		return err
	}
	fcCfg := firecrackerConfig(*m.Cfg)

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.LogrusEntry()),
	}
	// If the jailer is used, the final command will be built in firecracker.NewMachine(...).
	// If not, build the final command here.
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
		log.Errorf("failed to create a new firecracker machine: %v", err)
		return err
	}

	log.Debugf("starting firecracker machine with id: %s", fcCfg.VMID)
	if err := machine.Start(m.Ctx); err != nil {
		log.Errorf("failed to start firecracker machine: %v", err)
		return err
	}

	m.Machine = machine
	m.Ip = machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP
	m.state.Store(VmStateAvailable)

	return nil
}

func (m *MachineInstance) Ready() error {

	return nil
}

// ShutdownAndDestroy shuts a firecracker machine down and destroys it afterwards.
func (m *MachineInstance) ShutdownAndDestroy(ctx context.Context) error {
	log.Debugf("shutting down firecracker machine: %s", m.Cfg.VmId.String())
	if err := m.Machine.Shutdown(ctx); err != nil {
		return err
	}
	timeout := 3 * time.Second
	shutdownCtx, shutdownCtxCancel := context.WithTimeout(ctx, timeout)
	defer shutdownCtxCancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.IsRunning() {
				log.Debugf("firecracker machine has been shut down gracefully: %s", m.Cfg.VmId.String())
				return nil
			}
		case <-shutdownCtx.Done():
			if err := m.Machine.StopVMM(); err != nil {
				log.Errorf("failed to force stop the firecracker machine: %v", err)
				return err
			} else {
				log.Debugf("firecracker machine has been stopped forcefully: %s", m.Cfg.VmId.String())
				return nil
			}
		}
	}
}

// IsRunning returns if a firecracker machine is running.
func (m *MachineInstance) IsRunning() bool {
	// to check if the machine is running, try to establish a connection via the unix socket
	conn, err := net.Dial("unix", m.Machine.Cfg.SocketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (m *MachineInstance) EstablishConnection(ctx context.Context) error {
	addr := strings.Join([]string{m.Ip.String(), "50051"}, ":")
	keepAliveParams := keepalive.ClientParameters{
		Time:                10 * time.Second, // time after which a ping is sent if no activity; grpc defined min value is 10s
		Timeout:             3 * time.Second,  // time after which the connection is closed if no ack for ping
		PermitWithoutStream: true,             // allows pings even when there are no active streams
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		var conn *grpc.ClientConn
		var err error
		for i := 0; i < 25; i++ { // 25 retries, 5 retries per second = 5 second retrying
			conn, err = grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepAliveParams))
			if err == nil {
				m.ClientConn = conn
				return // success, no error to send
			}

			// wait before retrying, but stop if context is done
			select {
			case <-ctx.Done():
				errCh <- ctx.Err() // send context cancellation error
				return
			case <-time.After(200 * time.Millisecond): // retry delay
			}
		}

		// Send the last error after retries are exhausted
		errCh <- err
	}()

	return <-errCh // Wait for error or nil from the goroutine
}

// ConnectionAlive returns if a grpc client connection is still alive.
func (m *MachineInstance) ConnectionAlive() bool {
	return (m.ClientConn.GetState() == connectivity.Ready || m.ClientConn.GetState() == connectivity.Idle)
}

// Health checks the health status of an agent and returns the result.
func (m *MachineInstance) Health(ctx context.Context) (*health.HealthStatus, error) {
	if !m.ConnectionAlive() {
		return nil, fmt.Errorf("failed to check machine health - connection to agent not alive: %s", m.Cfg.VmId.String())
	}
	healthClient := health.NewHealthClient(m.ClientConn)
	healthStatus, err := healthClient.Status(ctx, &health.HealthStatusRequest{})
	if err != nil {
		return nil, err
	}
	return &healthStatus.Status, nil
}
