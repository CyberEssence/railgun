package models

import (
	"encoding/json"
	"time"
)

// BaseMetric общая структура для всех метрик
type BaseMetric struct {
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// MetricBatch контейнер для метрик
type MetricBatch struct {
	HostID    string                 `json:"host_id"`
	Hostname  string                 `json:"hostname"`
	Timestamp time.Time              `json:"timestamp"`
	System    *SystemInfo            `json:"system,omitempty"`
	Network   *NetworkInfo           `json:"network,omitempty"`
	Processes *ProcessInfo           `json:"processes,omitempty"`
	Security  *SecurityInfo          `json:"security,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// SystemInfo системная информация
type SystemInfo struct {
	BaseMetric   `json:",inline"`
	OS           string `json:"os"`
	Platform     string `json:"platform"`
	PlatformName string `json:"platform_name"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Uptime       uint64 `json:"uptime"`
	BootTime     uint64 `json:"boot_time"`
	Timezone     string `json:"timezone"`

	// CPU
	CPUCount    int       `json:"cpu_count"`
	CPUUsage    float64   `json:"cpu_usage"`
	LoadAverage []float64 `json:"load_average"`

	// Память
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryFree    uint64  `json:"memory_free"`
	MemoryPercent float64 `json:"memory_percent"`
	SwapTotal     uint64  `json:"swap_total"`
	SwapUsed      uint64  `json:"swap_used"`
	SwapFree      uint64  `json:"swap_free"`

	// Диски
	Disks     []DiskInfo `json:"disks"`
	Users     []UserInfo `json:"users"`
	Processes int        `json:"processes"`
	OpenFiles uint64     `json:"open_files"`
}

// NetworkInfo сетевая информация
type NetworkInfo struct {
	BaseMetric     `json:",inline"`
	Interfaces     []InterfaceInfo  `json:"interfaces"`
	Connections    []ConnectionInfo `json:"connections"`
	ListeningPorts []PortInfo       `json:"listening_ports"`
	Bandwidth      BandwidthInfo    `json:"bandwidth"`
	Firewall       FirewallInfo     `json:"firewall"`
	DNS            DNSInfo          `json:"dns"`
}

// ProcessInfo информация о процессах
type ProcessInfo struct {
	BaseMetric    `json:",inline"`
	Processes     []ProcessDetail `json:"processes"`
	TotalCount    int             `json:"total_count"`
	ZombieCount   int             `json:"zombie_count"`
	ThreadCount   int             `json:"thread_count"`
	RunningCount  int             `json:"running_count"`
	SleepingCount int             `json:"sleeping_count"`
	StoppedCount  int             `json:"stopped_count"`
}

// SecurityInfo информация о безопасности
type SecurityInfo struct {
	BaseMetric      `json:",inline"`
	FailedLogins    []FailedLogin  `json:"failed_logins"`
	LastLogins      []LastLogin    `json:"last_logins"`
	SudoCommands    []SudoCommand  `json:"sudo_commands"`
	OpenPorts       []OpenPort     `json:"open_ports"`
	Users           []UserSecurity `json:"users"`
	Updates         []SystemUpdate `json:"updates"`
	SuspiciousProcs []string       `json:"suspicious_processes"`
	FileChanges     []FileChange   `json:"file_changes"`
}

// DiskInfo информация о диске
type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	FSType     string  `json:"fs_type"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	Percent    float64 `json:"percent"`
	InodesUsed uint64  `json:"inodes_used"`
	InodesFree uint64  `json:"inodes_free"`
}

// UserInfo информация о пользователе
type UserInfo struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  string `json:"started"`
	PID      int32  `json:"pid"`
}

// InterfaceInfo информация о сетевом интерфейсе
type InterfaceInfo struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac"`
	IPs       []string `json:"ips"`
	MTU       int      `json:"mtu"`
	Speed     uint64   `json:"speed"`
	Flags     []string `json:"flags"`
	RxBytes   uint64   `json:"rx_bytes"`
	TxBytes   uint64   `json:"tx_bytes"`
	RxPackets uint64   `json:"rx_packets"`
	TxPackets uint64   `json:"tx_packets"`
	RxErrors  uint64   `json:"rx_errors"`
	TxErrors  uint64   `json:"tx_errors"`
	RxDropped uint64   `json:"rx_dropped"`
	TxDropped uint64   `json:"tx_dropped"`
}

// ConnectionInfo информация о сетевом соединении
type ConnectionInfo struct {
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint32 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
	PID        int32  `json:"pid"`
	Process    string `json:"process"`
	Protocol   string `json:"protocol"`
	UID        uint32 `json:"uid"`
}

// PortInfo информация о порте
type PortInfo struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process"`
	PID      int32  `json:"pid"`
	User     string `json:"user"`
}

// BandwidthInfo информация о пропускной способности
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

// FirewallInfo информация о файрволе
type FirewallInfo struct {
	Enabled bool     `json:"enabled"`
	Type    string   `json:"type"` // iptables, nftables, firewalld
	Rules   []string `json:"rules"`
}

// DNSInfo информация о DNS
type DNSInfo struct {
	Servers []string `json:"servers"`
	Domain  string   `json:"domain"`
	Search  []string `json:"search"`
}

// ProcessDetail детальная информация о процессе
type ProcessDetail struct {
	PID           int32   `json:"pid"`
	PPID          int32   `json:"ppid"`
	Name          string  `json:"name"`
	Command       string  `json:"command"`
	User          string  `json:"user"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	MemoryRSS     uint64  `json:"memory_rss"`
	MemoryVMS     uint64  `json:"memory_vms"`
	Status        string  `json:"status"`
	CreateTime    int64   `json:"create_time"`
	Threads       int32   `json:"threads"`
	Nice          int32   `json:"nice"`
	Priority      int32   `json:"priority"`
	Connections   int     `json:"connections"`
	OpenFiles     int32   `json:"open_files"`
}

// FailedLogin неудачная попытка входа
type FailedLogin struct {
	User      string    `json:"user"`
	From      string    `json:"from"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
}

// LastLogin последний вход
type LastLogin struct {
	User      string    `json:"user"`
	From      string    `json:"from"`
	Timestamp time.Time `json:"timestamp"`
	Terminal  string    `json:"terminal"`
	PID       int32     `json:"pid"`
}

// SudoCommand sudo команда
type SudoCommand struct {
	User    string    `json:"user"`
	Command string    `json:"command"`
	Time    time.Time `json:"time"`
	TTY     string    `json:"tty"`
	PWD     string    `json:"pwd"`
}

// OpenPort открытый порт
type OpenPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	State    string `json:"state"`
	Process  string `json:"process"`
	PID      int32  `json:"pid"`
}

// UserSecurity информация о пользователе для безопасности
type UserSecurity struct {
	User         string    `json:"user"`
	UID          uint32    `json:"uid"`
	GID          uint32    `json:"gid"`
	Home         string    `json:"home"`
	Shell        string    `json:"shell"`
	LastLogin    time.Time `json:"last_login"`
	FailedLogins int       `json:"failed_logins"`
	HasSudo      bool      `json:"has_sudo"`
	Locked       bool      `json:"locked"`
	PasswordAge  int       `json:"password_age"`
	Groups       []string  `json:"groups"`
}

// SystemUpdate системное обновление
type SystemUpdate struct {
	Package    string `json:"package"`
	Current    string `json:"current"`
	Available  string `json:"available"`
	Security   bool   `json:"security"`
	Repository string `json:"repository"`
}

// FileChange изменение файла
type FileChange struct {
	Path      string    `json:"path"`
	Action    string    `json:"action"` // created, modified, deleted
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	Mode      uint32    `json:"mode"`
	UID       uint32    `json:"uid"`
	GID       uint32    `json:"gid"`
	Hash      string    `json:"hash"`
}

// ToJSON конвертирует в JSON
func (m *MetricBatch) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// ToJSON конвертирует SystemInfo в JSON
func (s *SystemInfo) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ToJSON конвертирует NetworkInfo в JSON
func (n *NetworkInfo) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// ToJSON конвертирует ProcessInfo в JSON
func (p *ProcessInfo) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// ToJSON конвертирует SecurityInfo в JSON
func (s *SecurityInfo) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}
