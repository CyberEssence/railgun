package collector

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessInfo struct {
	Timestamp time.Time `json:"timestamp"`
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Type      string    `json:"type"` // "process_info"

	Processes   []ProcessDetail `json:"processes"`
	TotalCount  int             `json:"total_count"`
	ZombieCount int             `json:"zombie_count"`
	ThreadCount int             `json:"thread_count"`
}

type ProcessDetail struct {
	PID           int32    `json:"pid"`
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	User          string   `json:"user"`
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryPercent float32  `json:"memory_percent"`
	Status        []string `json:"status"`
	CreateTime    int64    `json:"create_time"`
	Threads       int32    `json:"threads"`
}

type ProcessCollector struct {
	hostID   string
	hostname string
}

func NewProcessCollector(hostID, hostname string) *ProcessCollector {
	return &ProcessCollector{
		hostID:   hostID,
		hostname: hostname,
	}
}

func (p *ProcessCollector) Name() string  { return "process" }
func (p *ProcessCollector) Enabled() bool { return true }

func (p *ProcessCollector) Collect() (interface{}, error) {
	return CollectProcessInfo(p.hostID, p.hostname)
}

func CollectProcessInfo(hostID, hostname string) (*ProcessInfo, error) {
	info := &ProcessInfo{
		Timestamp: time.Now().UTC(),
		HostID:    hostID,
		Hostname:  hostname,
		Type:      "process_info",
	}

	// Получаем список процессов
	processes, _ := process.Processes()
	info.TotalCount = len(processes)

	// Ограничиваем количество процессов для отправки (например, топ 50 по памяти)
	maxProcesses := 50
	if len(processes) > maxProcesses {
		processes = processes[:maxProcesses]
	}

	for _, proc := range processes {
		detail, err := getProcessDetail(proc)
		if err == nil {
			info.Processes = append(info.Processes, detail)

			if slices.Contains(detail.Status, "Z") {
				info.ZombieCount++
			}
			info.ThreadCount += int(detail.Threads)
		}
	}

	return info, nil
}

func getProcessDetail(proc *process.Process) (ProcessDetail, error) {
	detail := ProcessDetail{
		PID: proc.Pid,
	}

	// Имя процесса
	name, err := proc.Name()
	if err == nil {
		detail.Name = name
	}

	// Команда
	cmdline, err := proc.Cmdline()
	if err == nil && cmdline != "" {
		detail.Command = cmdline
	} else {
		// Если команда пустая, используем имя
		detail.Command = detail.Name
	}

	// Пользователь
	username, err := proc.Username()
	if err == nil {
		detail.User = username
	}

	// CPU и память
	cpuPercent, _ := proc.CPUPercent()
	detail.CPUPercent = cpuPercent

	memPercent, _ := proc.MemoryPercent()
	detail.MemoryPercent = memPercent

	// Статус
	status, err := proc.Status()
	if err == nil {
		detail.Status = status
	}

	// Время создания
	createTime, _ := proc.CreateTime()
	detail.CreateTime = createTime

	// Количество потоков
	threads, _ := proc.NumThreads()
	detail.Threads = threads

	return detail, nil
}

func (p *ProcessInfo) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
