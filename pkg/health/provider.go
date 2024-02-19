package health

import "sync/atomic"

type ProviderOptions struct {
	Targets int
}

type Provider interface {
	Ready()
	Healthy() bool
}

type healthStatusProvider struct {
	targets      int32
	targetsReady atomic.Int32
	healthy      atomic.Bool
}

// NewHealthStatusProvider creates a new Provider.
func NewHealthStatusProvider(opts ProviderOptions) Provider {
	return &healthStatusProvider{
		targets: int32(opts.Targets),
	}
}

// Ready tells the health status provider that a target is ready.
func (h *healthStatusProvider) Ready() {
	h.targetsReady.Add(1)
	if h.targetsReady.Load() >= h.targets {
		h.healthy.Store(true)
	}
}

// Healthy returns if all targets are ready.
func (h *healthStatusProvider) Healthy() bool {
	return h.healthy.Load()
}
