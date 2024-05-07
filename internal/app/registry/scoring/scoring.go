package scoring

import (
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
)

type ScoringResult struct {
	InstanceUuid string
	InstanceType registrypb.InstanceType
	Score        float64
}

// CalculateScore calculates the score of a service instance based on its metrics.
func CalculateScore(metrics *registrypb.ServiceInstanceMetrics) float64 {
	return 100 - float64((metrics.CpuUsage+metrics.MemoryUsage)/2)
}
