package evaluation

import (
	"fmt"
	"sort"

	"github.com/dennishilgert/apollo/pkg/logger"
	fleetpb "github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
)

var log = logger.NewLogger("apollo.evaluation")

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
	} else if (metrics.CpuUsage > 70) || (metrics.MemoryUsage > 70) {
		result.Restriction = UsageRestrictionException
	} else {
		result.Restriction = UsageRestrictionUnrestricted
	}
	return result
}

func EvaluateRunnerProvisioning(machineWeight fleetpb.MachineWeight, workerInstance *registrypb.WorkerInstance, metrics *registrypb.WorkerInstanceMetrics) *EvaluationResult {
	result := &EvaluationResult{
		WorkerInstance: workerInstance,
	}
	var score int
	switch machineWeight {
	case fleetpb.MachineWeight_LIGHT:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.LightRunnersCount), 128, 0.6)
	case fleetpb.MachineWeight_MEDIUM:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.MediumRunnersCount), 256, 0.2)
	case fleetpb.MachineWeight_HEAVY:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.HeavyRunnersCount), 512, 0.1)
	case fleetpb.MachineWeight_SUPER_HEAVY:
		score = calculateScore(int(metrics.MemoryTotal), int(metrics.RunnerPoolMetrics.SuperHeavyRunnersCount), 1024, 0.1)
	}
	result.Score = score
	result.Restriction = evaluateUsageRestriction(score)

	log.Debugf("evaluated worker %s with machine weight %s and got score %d", workerInstance.WorkerUuid, machineWeight, score)
	return result
}

func SelectWorker(workerEvaluations []EvaluationResult) (*registrypb.WorkerInstance, error) {
	// Sort worker evaluations by descending score.
	sort.Slice(workerEvaluations, func(i, j int) bool {
		return workerEvaluations[i].Score > workerEvaluations[j].Score
	})

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

func SelectWorkers(workerEvaluations []EvaluationResult, n int) ([]*registrypb.WorkerInstance, error) {
	// Sort worker evaluations by descending score.
	sort.Slice(workerEvaluations, func(i, j int) bool {
		return workerEvaluations[i].Score > workerEvaluations[j].Score
	})
	if n > len(workerEvaluations) {
		n = len(workerEvaluations)
	}

	// Initialize an empty slice for the selected workers
	selectedWorkers := []*registrypb.WorkerInstance{}

	// First pass: select workers with unrestricted usage.
	for _, evaluation := range workerEvaluations {
		if evaluation.Restriction == UsageRestrictionUnrestricted {
			selectedWorkers = append(selectedWorkers, evaluation.WorkerInstance)
			if len(selectedWorkers) == n {
				return selectedWorkers, nil
			}
		}
	}

	// Second pass: select workers with exception usage if needed.
	for _, evaluation := range workerEvaluations {
		if evaluation.Restriction == UsageRestrictionException {
			selectedWorkers = append(selectedWorkers, evaluation.WorkerInstance)
			if len(selectedWorkers) == n {
				return selectedWorkers, nil
			}
		}
	}

	if len(selectedWorkers) == 0 {
		return nil, fmt.Errorf("no worker with free capacity available")
	}

	return selectedWorkers, nil
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

func evaluateUsageRestriction(score int) UsageRestriction {
	if score > 50 {
		return UsageRestrictionUnrestricted
	} else if score > 30 {
		return UsageRestrictionException
	}
	return UsageRestrictionForbidden

}

func calculateScore(totalMemory int, runnerCount int, memoryPerRunner int, densityFactor float32) int {
	maxMemory := float32(totalMemory) * densityFactor
	usedMemory := runnerCount * memoryPerRunner
	maxRunnerCount := int(maxMemory) / memoryPerRunner
	runnerCountFactor := (runnerCount / maxRunnerCount) * 10
	return 100 - int((float32(usedMemory)/maxMemory)*100) - runnerCountFactor
}
