package microvm

import (
	"context"
	"time"

	"github.com/dennishilgert/apollo/internal/app/manager/microvm/pool"
	"github.com/dennishilgert/apollo/pkg/utils"
)

type Options struct {
	OsArch                utils.OsArch
	FirecrackerBinaryPath string
	VmHealthCheckInterval time.Duration
}

type Operator interface {
	Init(ctx context.Context) error
}

type vmOperator struct {
	pool pool.Pool
}

func NewVmOperator(ctx context.Context, opts Options) Operator {
	return &vmOperator{
		pool: pool.NewVmPool(opts.VmHealthCheckInterval),
	}
}

func (m *vmOperator) Init(ctx context.Context) error {
	return nil
}
