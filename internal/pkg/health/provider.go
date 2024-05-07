package health

import "sync/atomic"

type ProviderOptions struct {
	Targets int
}

type Provider interface {
	WithCallback(callback func()) Provider
	Ready()
	Healthy() bool
}

type healthStatusProvider struct {
	targets      int32
	targetsReady atomic.Int32
	healthy      atomic.Bool
	callback     func()
}

// NewHealthStatusProvider creates a new Provider.
func NewHealthStatusProvider(opts ProviderOptions) Provider {
	return &healthStatusProvider{
		targets: int32(opts.Targets),
	}
}

// WithCallback adds a callback function which is called when all targets are ready.
func (h *healthStatusProvider) WithCallback(callback func()) Provider {
	h.callback = callback
	return h
}

// Ready tells the health status provider that a target is ready.
func (h *healthStatusProvider) Ready() {
	h.targetsReady.Add(1)
	if h.targetsReady.Load() >= h.targets {
		h.healthy.Store(true)
		// As all targets are ready, call the callback function if set.
		if h.callback != nil {
			h.callback()
		}
	}
}

// Healthy returns if all targets are ready.
func (h *healthStatusProvider) Healthy() bool {
	return h.healthy.Load()
}
