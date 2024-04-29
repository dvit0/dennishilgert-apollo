package metrics

import (
	"fmt"
	"math"

	"github.com/dennishilgert/apollo/pkg/logger"
	registrypb "github.com/dennishilgert/apollo/pkg/proto/registry/v1"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var log = logger.NewLogger("apollo.metrics")

type MetricsService interface {
	ServiceInstanceMetrics() *registrypb.ServiceInstanceMetrics
	WorkerInstanceMetrics() *registrypb.WorkerInstanceMetrics
}

type metricsService struct{}

// NewMetricsService creates a new MetricsService instance.
func NewMetricsService() MetricsService {
	return &metricsService{}
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
	diskUsage, err := m.diskUsage()
	if err != nil {
		log.Warnf("failed to build service instance metrics: %v", err)
		return nil
	}
	// TODO: Implement light, medium, and heavy functions load.
	return &registrypb.WorkerInstanceMetrics{
		CpuUsage:            int32(cpuUsage),
		MemoryUsage:         int32(memUsage),
		StorageUsage:        int32(diskUsage),
		LightFunctionsLoad:  0,
		MediumFunctionsLoad: 0,
		HeavyFunctionsLoad:  0,
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

// memoryUsage returns the current memory usage in percent.
func (m *metricsService) memoryUsage() (int, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory usage: %w", err)
	}
	return int(math.Round(stat.UsedPercent)), nil
}

// diskUsage returns the current disk usage in percent.
func (m *metricsService) diskUsage() (int, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return 0, fmt.Errorf("failed to get disk usage: %w", err)
	}
	return int(math.Round(stat.UsedPercent)), nil
}
