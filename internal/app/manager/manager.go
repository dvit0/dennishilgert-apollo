package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/dennishilgert/apollo/internal/app/manager/microvm"
	"github.com/dennishilgert/apollo/pkg/concurrency/runner"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager")

type Options struct {
	FirecrackerBinaryPath string
	VmHealthCheckInterval time.Duration
}

type Manager interface {
	Run(ctx context.Context) error
}

type manager struct {
	vmOperator microvm.Operator
}

func NewManager(ctx context.Context, opts Options) (Manager, error) {
	return &manager{
		vmOperator: microvm.NewVmOperator(ctx, microvm.Options{
			OsArch:                utils.DetectArchitecture(),
			FirecrackerBinaryPath: opts.FirecrackerBinaryPath,
		}),
	}, nil
}

func (m *manager) Run(ctx context.Context) error {
	log.Info("apollo manager is starting")

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			log.Info("initializing vm operator")
			if err := m.vmOperator.Init(ctx); err != nil {
				return fmt.Errorf("failed to initialize vm operator: %v", err)
			}
			return nil
		},
	)
	return runner.Run(ctx)
}
