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
	WatchdogCheckInterval time.Duration
	WatchdogWorkerCount   int
	AgentApiPort          int
}

type Operator interface {
	Init(ctx context.Context) error
}

type vmOperator struct {
	pool pool.Pool
}

func NewVmOperator(ctx context.Context, opts Options) Operator {
	return &vmOperator{
		pool: pool.NewVmPool(pool.Options{
			WatchdogCheckInterval: opts.WatchdogCheckInterval,
			WatchdogWorkerCount:   opts.WatchdogWorkerCount,
		}),
	}
}

func (v *vmOperator) Init(ctx context.Context) error {
	return nil
}

func (v *vmOperator) GetOrStart(ctx context.Context) error {
	return nil
}
