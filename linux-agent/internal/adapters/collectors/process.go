package collectors

import (
	"context"
	"time"

	"linux-agent/internal/core/domain"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessCollector сборщик информации о процессах
type ProcessCollector struct {
	BaseCollector
}

// NewProcessCollector создает новый коллектор процессов
func NewProcessCollector(hostID, hostname string, enabled bool) *ProcessCollector {
	return &ProcessCollector{
		BaseCollector: BaseCollector{
			HostID:     hostID,
			Hostname:   hostname,
			NameVal:    "processes",
			EnabledVal: enabled,
		},
	}
}

// Collect собирает информацию о процессах
func (c *ProcessCollector) Collect(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info := &domain.ProcessInfo{
		BaseMetric: domain.BaseMetric{
			HostID:    c.HostID,
			Hostname:  c.Hostname,
			Timestamp: time.Now().UTC(),
			Type:      "processes",
		},
	}

	// Получаем список процессов
	processes, _ := process.Processes()
	info.TotalCount = len(processes)

	// Ограничиваем количество процессов для отправки
	maxProcesses := 50
	if len(processes) > maxProcesses {
		processes = processes[:maxProcesses]
	}

	for _, proc := range processes {
		p, err := c.getProcessDetail(ctx, proc)
		if err == nil {
			info.Processes = append(info.Processes, *p)

			if p.Status == "Z" {
				info.ZombieCount++
			}
			info.ThreadCount += int(p.Threads)
		}
	}

	return info, nil
}

func (c *ProcessCollector) getProcessDetail(ctx context.Context, proc *process.Process) (*domain.Process, error) {
	detail := &domain.Process{
		PID: proc.Pid,
	}

	// Имя процесса
	name, _ := proc.NameWithContext(ctx)
	detail.Name = name

	// Команда
	cmdline, _ := proc.CmdlineWithContext(ctx)
	detail.Command = cmdline

	// Пользователь
	username, _ := proc.UsernameWithContext(ctx)
	detail.User = username

	// CPU и память
	cpuPercent, _ := proc.CPUPercentWithContext(ctx)
	detail.CPUPercent = cpuPercent

	memPercent, _ := proc.MemoryPercentWithContext(ctx)
	detail.MemoryPercent = memPercent

	// Статус
	status, _ := proc.StatusWithContext(ctx)
	if len(status) > 0 {
		detail.Status = status[0]
	}

	// Время создания
	createTime, _ := proc.CreateTimeWithContext(ctx)
	detail.CreateTime = createTime

	// Количество потоков
	threads, _ := proc.NumThreadsWithContext(ctx)
	detail.Threads = int32(threads)

	return detail, nil
}
