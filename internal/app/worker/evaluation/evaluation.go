package evaluation

import (
	"fmt"

	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

type RunnerHeaviness int

const (
	RunnerHeavinessLight RunnerHeaviness = iota
	RunnerHeavinessMedium
	RunnerHeavinessHeavy
	RunnerHeavinessSuperHeavy
)

func (f RunnerHeaviness) String() string {
	switch f {
	case RunnerHeavinessLight:
		return "LIGHT"
	case RunnerHeavinessMedium:
		return "MEDIUM"
	case RunnerHeavinessHeavy:
		return "HEAVY"
	case RunnerHeavinessSuperHeavy:
		return "SUPER_HEAVY"
	}
	return "UNKNOWN"
}

type UsageRestriction int

const (
	UsageRestrictionUnrestricted UsageRestriction = iota
	UsageRestrictionException
	UsageRestrictionForbidden
)

type EvaluationResult struct {
	WorkerInstance *registrypb.WorkerInstance
	Restriction    UsageRestriction
	Score          int
}

func EvaluateFunctionInitialization(workerInstance *registrypb.WorkerInstance, metrics *registrypb.WorkerInstanceMetrics) *EvaluationResult {
	result := &EvaluationResult{
		WorkerInstance: workerInstance,
		Score:          100 - int(metrics.CpuUsage+metrics.MemoryUsage)/2,
	}
	if (metrics.CpuUsage > 90) || (metrics.MemoryUsage > 90) {
		result.Restriction = UsageRestrictionForbidden
	} else if (metrics.CpuUsage > 80) || (metrics.MemoryUsage > 80) {
		result.Restriction = UsageRestrictionException
	} else {
		result.Restriction = UsageRestrictionUnrestricted
	}
	return result
}

func EvaluateRunnerProvisioning(heaviness RunnerHeaviness, workerInstance *registrypb.WorkerInstance, metrics *registrypb.WorkerInstanceMetrics) *EvaluationResult {
	result := &EvaluationResult{
		WorkerInstance: workerInstance,
	}
	var score int
	switch heaviness {
	case RunnerHeavinessLight:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.LightRunnersCount), 128, 0.6)
	case RunnerHeavinessMedium:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.MediumRunnersCount), 256, 0.2)
	case RunnerHeavinessHeavy:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.HeavyRunnersCount), 512, 0.1)
	case RunnerHeavinessSuperHeavy:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.SuperHeavyRunnersCount), 1024, 0.1)
	}
	result.Score = score
	result.Restriction = evaluateUsageRestriction(score)
	return result
}

func SelectWorker(workerEvaluations []EvaluationResult) (*registrypb.WorkerInstance, error) {
	var selectedWorker *registrypb.WorkerInstance
	selectedWorker = selectMostFreeCapacityWorker(workerEvaluations, UsageRestrictionUnrestricted)
	if selectedWorker != nil {
		return selectedWorker, nil
	}
	selectedWorker = selectMostFreeCapacityWorker(workerEvaluations, UsageRestrictionException)
	if selectedWorker != nil {
		return selectedWorker, nil
	}
	return nil, fmt.Errorf("no worker with free capacity available")
}

func selectMostFreeCapacityWorker(workerEvaluations []EvaluationResult, restriction UsageRestriction) *registrypb.WorkerInstance {
	var selectedWorker *registrypb.WorkerInstance
	var maxScore int
	for _, evaluation := range workerEvaluations {
		if evaluation.Restriction != restriction {
			continue
		}
		if evaluation.Score > maxScore {
			maxScore = evaluation.Score
			selectedWorker = evaluation.WorkerInstance
		}
	}
	return selectedWorker
}

func EvaluateRunnerHeaviness(cpuCores int, memoryLimit int) RunnerHeaviness {
	if cpuCores <= 1 && memoryLimit <= 128 {
		return RunnerHeavinessLight
	} else if cpuCores <= 2 && memoryLimit <= 256 {
		return RunnerHeavinessMedium
	} else if cpuCores <= 2 && memoryLimit <= 512 {
		return RunnerHeavinessHeavy
	}
	return RunnerHeavinessSuperHeavy
}

func evaluateUsageRestriction(score int) UsageRestriction {
	if score > 90 {
		return UsageRestrictionUnrestricted
	} else if score > 80 {
		return UsageRestrictionException
	}
	return UsageRestrictionForbidden

}

func calculateScore(totalMemory int, runnerCount int, memoryPerRunner int, densityFactor float32) int {
	maxMemory := float32(totalMemory) * densityFactor
	usedMemory := runnerCount * memoryPerRunner
	return 100 - int((float32(usedMemory)/maxMemory)*100)
}
