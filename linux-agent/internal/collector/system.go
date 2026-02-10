package collector

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	Timestamp time.Time `json:"timestamp"`
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Type      string    `json:"type"` // "system_info"

	// Основная информация
	OS           string `json:"os"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platform_name"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Uptime       uint64 `json:"uptime"`

	// CPU
	CPUCount    int       `json:"cpu_count"`
	CPUUsage    float64   `json:"cpu_usage"`
	LoadAverage []float64 `json:"load_average"`

	// Память
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryPercent float64 `json:"memory_percent"`
	SwapTotal     uint64  `json:"swap_total"`
	SwapUsed      uint64  `json:"swap_used"`

	// Диски
	Disks []DiskInfo `json:"disks"`

	// Дополнительно
	Users     []UserInfo `json:"users"`
	Processes int        `json:"processes"`
	OpenFiles uint64     `json:"open_files"`
	BootTime  uint64     `json:"boot_time"`
	Timezone  string     `json:"timezone"`
}

type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	FSType     string  `json:"fs_type"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	Percent    float64 `json:"percent"`
}

type UserInfo struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
}

func CollectSystemInfo(hostID, hostname string) (*SystemInfo, error) {
	info := &SystemInfo{
		Timestamp: time.Now().UTC(),
		HostID:    hostID,
		Hostname:  hostname,
		Type:      "system_info",
	}

	// Информация о хосте
	hostInfo, _ := host.Info()
	if hostInfo != nil {
		info.OS = hostInfo.OS
		info.Platform = hostInfo.Platform
		info.PlatformName = hostInfo.PlatformFamily
		info.Kernel = hostInfo.KernelVersion
		info.Architecture = hostInfo.KernelArch
		info.Uptime = hostInfo.Uptime
		info.BootTime = hostInfo.BootTime
	}

	info.Architecture = runtime.GOARCH

	// CPU
	info.CPUCount = runtime.NumCPU()
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		info.CPUUsage = cpuPercent[0]
	}

	loadAvg, _ := load.Avg()
	if loadAvg != nil {
		info.LoadAverage = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	}

	// Память
	memStat, _ := mem.VirtualMemory()
	if memStat != nil {
		info.MemoryTotal = memStat.Total
		info.MemoryUsed = memStat.Used
		info.MemoryPercent = memStat.UsedPercent
	}

	swapStat, _ := mem.SwapMemory()
	if swapStat != nil {
		info.SwapTotal = swapStat.Total
		info.SwapUsed = swapStat.Used
	}

	// Диски
	partitions, _ := disk.Partitions(true)
	for _, part := range partitions {
		usage, err := disk.Usage(part.Mountpoint)
		if err == nil {
			info.Disks = append(info.Disks, DiskInfo{
				Device:     part.Device,
				Mountpoint: part.Mountpoint,
				FSType:     part.Fstype,
				Total:      usage.Total,
				Used:       usage.Used,
				Free:       usage.Free,
				Percent:    usage.UsedPercent,
			})
		}
	}

	// Пользователи
	info.collectUsers()

	// Количество процессов
	info.collectProcessCount()

	return info, nil
}

func (s *SystemInfo) collectUsers() {
	// Чтение who или w
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("w", "-h")
		output, err = cmd.Output()
		if err != nil {
			return
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 {
			userInfo := UserInfo{
				User: parts[0],
			}

			// Определяем терминал и хост
			if len(parts) > 1 {
				userInfo.Terminal = parts[1]
			}
			if len(parts) > 2 {
				userInfo.Host = parts[2]
			}
			if len(parts) > 3 {
				userInfo.Started = parts[3]
			}

			s.Users = append(s.Users, userInfo)
		}
	}
}

func (s *SystemInfo) collectProcessCount() {
	// Подсчет процессов через ps
	cmd := exec.Command("ps", "-e", "--no-headers")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		s.Processes = len(lines) - 1 // минус пустая строка
	}

	// Количество открытых файлов для текущего пользователя
	currentUser := os.Getenv("USER")
	if currentUser != "" {
		cmd = exec.Command("lsof", "-u", currentUser)
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			s.OpenFiles = uint64(len(lines) - 1)
		}
	}
}

func (s *SystemInfo) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}
