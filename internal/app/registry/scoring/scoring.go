package scoring

import (
	messagespb "github.com/dennishilgert/apollo/pkg/proto/messages/v1"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

type ScoringResult struct {
	InstanceUuid string
	ServiceType  registrypb.ServiceType
	Score        float64
}

// CalculateScore calculates the score of a service instance based on its metrics.
func CalculateScore(message *messagespb.InstanceHeartbeatMessage) ScoringResult {
	return ScoringResult{
		InstanceUuid: message.InstanceUuid,
		ServiceType:  message.ServiceType,
		Score:        float64(message.Metrics.CpuUsage+message.Metrics.MemoryUsage) / 2,
	}
}
