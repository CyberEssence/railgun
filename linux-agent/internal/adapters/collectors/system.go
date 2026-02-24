package collectors

import (
	"context"
	"runtime"
	"time"

	"linux-agent/internal/core/domain"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemCollector сборщик системной информации
type SystemCollector struct {
	BaseCollector
}

// NewSystemCollector создает новый системный коллектор
func NewSystemCollector(hostID, hostname string, enabled bool) *SystemCollector {
	return &SystemCollector{
		BaseCollector: BaseCollector{
			HostID:     hostID,
			Hostname:   hostname,
			NameVal:    "system",
			EnabledVal: enabled,
		},
	}
}

// Collect собирает системную информацию
func (c *SystemCollector) Collect(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info := &domain.SystemInfo{
		BaseMetric: domain.BaseMetric{
			HostID:    c.HostID,
			Hostname:  c.Hostname,
			Timestamp: time.Now().UTC(),
			Type:      "system",
		},
	}

	// Информация о хосте
	hostInfo, _ := host.InfoWithContext(ctx)
	if hostInfo != nil {
		info.OS = hostInfo.OS
		info.Platform = hostInfo.Platform
		info.Kernel = hostInfo.KernelVersion
		info.Architecture = hostInfo.KernelArch
		info.Uptime = hostInfo.Uptime
	}

	info.Architecture = runtime.GOARCH

	// CPU
	info.CPUCount = runtime.NumCPU()
	cpuPercent, _ := cpu.PercentWithContext(ctx, 0, false)
	if len(cpuPercent) > 0 {
		info.CPUUsage = cpuPercent[0]
	}

	loadAvg, _ := load.AvgWithContext(ctx)
	if loadAvg != nil {
		info.LoadAverage = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	}

	// Память
	memStat, _ := mem.VirtualMemoryWithContext(ctx)
	if memStat != nil {
		info.MemoryTotal = memStat.Total
		info.MemoryUsed = memStat.Used
		info.MemoryFree = memStat.Free
	}

	swapStat, _ := mem.SwapMemoryWithContext(ctx)
	if swapStat != nil {
		info.SwapTotal = swapStat.Total
		info.SwapUsed = swapStat.Used
	}

	// Диски
	partitions, _ := disk.PartitionsWithContext(ctx, true)
	for _, part := range partitions {
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err == nil {
			info.Disks = append(info.Disks, domain.Disk{
				Device:     part.Device,
				Mountpoint: part.Mountpoint,
				FSType:     part.Fstype,
				Total:      usage.Total,
				Used:       usage.Used,
				Free:       usage.Free,
				Percent:    int64(usage.UsedPercent),
			})
		}
	}

	return info, nil
}
