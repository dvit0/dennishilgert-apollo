package scoring

import (
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

type ScoringResult struct {
	InstanceUuid string
	InstanceType registrypb.InstanceType
	Score        float64
}

// CalculateScore calculates the score of a service instance based on its metrics.
func CalculateScore(metrics *registrypb.ServiceInstanceMetrics) float64 {
	return float64((metrics.CpuUsage + metrics.MemoryUsage) / 2)
}
