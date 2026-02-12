package models

import (
	"time"
)

// MetricBatch - основной контейнер для метрик
type MetricBatch struct {
	HostID    string                 `json:"host_id"`
	Hostname  string                 `json:"hostname"`
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// SystemInfo - системная информация
type SystemInfo struct {
	Timestamp     time.Time  `json:"timestamp"`
	HostID        string     `json:"host_id"`
	Hostname      string     `json:"hostname"`
	Type          string     `json:"type"`
	OS            string     `json:"os"`
	Platform      string     `json:"platform"`
	Kernel        string     `json:"kernel"`
	Architecture  string     `json:"architecture"`
	Uptime        uint64     `json:"uptime"`
	CPUCount      int        `json:"cpu_count"`
	CPUUsage      float64    `json:"cpu_usage"`
	LoadAverage   []float64  `json:"load_average"`
	MemoryTotal   uint64     `json:"memory_total"`
	MemoryUsed    uint64     `json:"memory_used"`
	MemoryPercent float64    `json:"memory_percent"`
	SwapTotal     uint64     `json:"swap_total"`
	SwapUsed      uint64     `json:"swap_used"`
	Disks         []DiskInfo `json:"disks"`
	Users         []UserInfo `json:"users"`
	Processes     int        `json:"processes"`
	BootTime      uint64     `json:"boot_time"`
}

// NetworkInfo - сетевая информация
type NetworkInfo struct {
	Timestamp      time.Time        `json:"timestamp"`
	HostID         string           `json:"host_id"`
	Hostname       string           `json:"hostname"`
	Type           string           `json:"type"`
	Interfaces     []InterfaceInfo  `json:"interfaces"`
	Connections    []ConnectionInfo `json:"connections"`
	ListeningPorts []PortInfo       `json:"listening_ports"`
	Bandwidth      BandwidthInfo    `json:"bandwidth"`
	DNS            DNSInfo          `json:"dns"`
}

// ProcessInfo - информация о процессах
type ProcessInfo struct {
	Timestamp   time.Time       `json:"timestamp"`
	HostID      string          `json:"host_id"`
	Hostname    string          `json:"hostname"`
	Type        string          `json:"type"`
	Processes   []ProcessDetail `json:"processes"`
	TotalCount  int             `json:"total_count"`
	ZombieCount int             `json:"zombie_count"`
	ThreadCount int             `json:"thread_count"`
}

// SecurityInfo - информация о безопасности
type SecurityInfo struct {
	Timestamp    time.Time      `json:"timestamp"`
	HostID       string         `json:"host_id"`
	Hostname     string         `json:"hostname"`
	Type         string         `json:"type"`
	FailedLogins []FailedLogin  `json:"failed_logins"`
	LastLogins   []LastLogin    `json:"last_logins"`
	SudoCommands []SudoCommand  `json:"sudo_commands"`
	OpenPorts    []OpenPort     `json:"open_ports"`
	Users        []UserSecurity `json:"users"`
	Updates      []SystemUpdate `json:"updates"`
}

// Вспомогательные структуры
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

type InterfaceInfo struct {
	Name  string   `json:"name"`
	MAC   string   `json:"mac"`
	IPs   []string `json:"ips"`
	MTU   int      `json:"mtu"`
	Flags []string `json:"flags"`
}

type ConnectionInfo struct {
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint32 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
	PID        int32  `json:"pid"`
	Process    string `json:"process"`
}

type PortInfo struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process"`
	PID      int32  `json:"pid"`
}

type BandwidthInfo struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
}

type DNSInfo struct {
	Servers []string `json:"servers"`
	Domain  string   `json:"domain"`
	Search  []string `json:"search"`
}

type ProcessDetail struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Command       string  `json:"command"`
	User          string  `json:"user"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	Status        string  `json:"status"`
	CreateTime    int64   `json:"create_time"`
	Threads       int32   `json:"threads"`
}

type FailedLogin struct {
	User      string `json:"user"`
	From      string `json:"from"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}

type LastLogin struct {
	User string `json:"user"`
	From string `json:"from"`
	When string `json:"when"`
}

type SudoCommand struct {
	User    string `json:"user"`
	Command string `json:"command"`
	Time    string `json:"time"`
}

type OpenPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	State    string `json:"state"`
}

type UserSecurity struct {
	User         string `json:"user"`
	LastLogin    string `json:"last_login"`
	FailedLogins int    `json:"failed_logins"`
	HasSudo      bool   `json:"has_sudo"`
	Locked       bool   `json:"locked"`
}

type SystemUpdate struct {
	Package   string `json:"package"`
	Current   string `json:"current"`
	Available string `json:"available"`
	Security  bool   `json:"security"`
}
