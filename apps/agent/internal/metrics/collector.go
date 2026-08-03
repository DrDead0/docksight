// Package metrics samples host-level resource utilisation (CPU, memory).
package metrics

import (
	"context"
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// CPU is host processor utilisation for one sample window.
type CPU struct {
	UsagePercent float64     `json:"usagePercent"`
	Cores        int         `json:"cores"`
	LoadAvg      *[3]float64 `json:"loadAvg"`
}

// Memory is host physical memory utilisation.
type Memory struct {
	TotalBytes     uint64  `json:"totalBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsagePercent   float64 `json:"usagePercent"`
}

// Snapshot is one host metrics reading.
type Snapshot struct {
	CPU    CPU    `json:"cpu"`
	Memory Memory `json:"memory"`
}

// Collector samples host CPU and memory usage.
type Collector struct {
	cores int
}

// NewCollector creates a host metrics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Prime takes a throwaway CPU sample so the next Collect reports utilisation
// over the interval since this call rather than since boot.
func (c *Collector) Prime(ctx context.Context) {
	_, _ = cpu.PercentWithContext(ctx, 0, false)
}

// Collect returns current host usage. CPU percent is the delta since the
// previous Collect (or Prime) call, so call it on a fixed interval.
func (c *Collector) Collect(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot

	// An interval of 0 means "since the last call", which pairs with the
	// caller's ticker and avoids blocking here for a sampling window.
	percents, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return snapshot, fmt.Errorf("sample cpu: %w", err)
	}
	if len(percents) > 0 {
		snapshot.CPU.UsagePercent = clampPercent(percents[0])
	}

	snapshot.CPU.Cores = c.coreCount(ctx)

	// Load average is absent on some platforms. Windows has no real equivalent
	// and reports a flat zero, so an all-zero reading is treated as
	// "unavailable" rather than surfaced as a misleading 0.00.
	if avg, avgErr := load.AvgWithContext(ctx); avgErr == nil && avg != nil {
		if avg.Load1 > 0 || avg.Load5 > 0 || avg.Load15 > 0 {
			snapshot.CPU.LoadAvg = &[3]float64{
				round2(avg.Load1),
				round2(avg.Load5),
				round2(avg.Load15),
			}
		}
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return snapshot, fmt.Errorf("sample memory: %w", err)
	}
	snapshot.Memory = Memory{
		TotalBytes:     vm.Total,
		UsedBytes:      vm.Used,
		AvailableBytes: vm.Available,
		UsagePercent:   clampPercent(vm.UsedPercent),
	}

	return snapshot, nil
}

// coreCount caches the logical core count; it cannot change while we run.
func (c *Collector) coreCount(ctx context.Context) int {
	if c.cores > 0 {
		return c.cores
	}
	count, err := cpu.CountsWithContext(ctx, true)
	if err != nil || count <= 0 {
		return 0
	}
	c.cores = count
	return count
}

// clampPercent guards against NaN and out-of-range readings so the JSON payload
// is always a valid 0-100 number.
func clampPercent(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return round2(value)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
