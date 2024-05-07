package metrics

import (
	"fmt"
	"math"

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/internal/pkg/proto/registry/v1"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var log = logger.NewLogger("apollo.metrics")

type MetricsService interface {
	ServiceInstanceMetrics() *registrypb.ServiceInstanceMetrics
	WorkerInstanceMetrics() *registrypb.WorkerInstanceMetrics
}

type metricsService struct {
	runnerPoolMetricsFunc func() *registrypb.RunnerPoolMetrics
}

// NewMetricsService creates a new MetricsService instance.
func NewMetricsService(runnerPoolMetricsFunc func() *registrypb.RunnerPoolMetrics) MetricsService {
	return &metricsService{
		runnerPoolMetricsFunc: runnerPoolMetricsFunc,
	}
}

// ServiceInstanceMetrics returns the current metrics of the service instance.
func (m *metricsService) ServiceInstanceMetrics() *registrypb.ServiceInstanceMetrics {
	cpuUsage, err := m.cpuUsage()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	memUsage, err := m.memoryUsage()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	return &registrypb.ServiceInstanceMetrics{
		CpuUsage:    int32(cpuUsage),
		MemoryUsage: int32(memUsage),
	}
}

// WorkerInstanceMetrics returns the current metrics of the worker instance.
func (m *metricsService) WorkerInstanceMetrics() *registrypb.WorkerInstanceMetrics {
	cpuUsage, err := m.cpuUsage()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	memUsage, err := m.memoryUsage()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	memTotal, err := m.memoryTotal()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	// TODO: Implement light, medium, and heavy functions load.
	return &registrypb.WorkerInstanceMetrics{
		CpuUsage:          int32(cpuUsage),
		MemoryUsage:       int32(memUsage),
		MemoryTotal:       int32(memTotal),
		RunnerPoolMetrics: m.runnerPoolMetricsFunc(),
	}
}

// cpuUsage returns the current CPU usage in percent.
func (m *metricsService) cpuUsage() (int, error) {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get cpu usage: %w", err)
	}
	if len(percent) > 0 {
		return int(math.Round(percent[0])), nil
	}
	return 0, nil
}

// memoryTotal returns the total memory in megabytes.
func (m *metricsService) memoryTotal() (int, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory total: %w", err)
	}
	return int(stat.Total / (1 << 20)), nil

}

// memoryUsage returns the current memory usage in percent.
func (m *metricsService) memoryUsage() (int, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory usage: %w", err)
	}
	return int(math.Round(stat.UsedPercent)), nil
}
