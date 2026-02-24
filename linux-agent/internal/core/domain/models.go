package domain

import "time"

// BaseMetric общая структура для всех метрик
type BaseMetric struct {
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

// SystemInfo системная информация
type SystemInfo struct {
	BaseMetric
	OS           string    `json:"os"`
	Platform     string    `json:"platform"`
	Kernel       string    `json:"kernel"`
	Architecture string    `json:"architecture"`
	Uptime       uint64    `json:"uptime"`
	CPUCount     int       `json:"cpu_count"`
	CPUUsage     float64   `json:"cpu_usage"`
	LoadAverage  []float64 `json:"load_average"`
	MemoryTotal  uint64    `json:"memory_total"`
	MemoryUsed   uint64    `json:"memory_used"`
	MemoryFree   uint64    `json:"memory_free"`
	SwapTotal    uint64    `json:"swap_total"`
	SwapUsed     uint64    `json:"swap_used"`
	Disks        []Disk    `json:"disks"`
}

// NetworkInfo сетевая информация
type NetworkInfo struct {
	BaseMetric
	Interfaces     []Interface  `json:"interfaces"`
	Connections    []Connection `json:"connections"`
	ListeningPorts []Port       `json:"listening_ports"`
	Bandwidth      Bandwidth    `json:"bandwidth"`
	DNS            DNS          `json:"dns"`
	Firewall       Firewall     `json:"firewall"`
}

// ProcessInfo информация о процессах
type ProcessInfo struct {
	BaseMetric
	Processes   []Process `json:"processes"`
	TotalCount  int       `json:"total_count"`
	ZombieCount int       `json:"zombie_count"`
	ThreadCount int       `json:"thread_count"`
}

// SecurityInfo информация о безопасности
type SecurityInfo struct {
	BaseMetric
	FailedLogins []FailedLogin  `json:"failed_logins"`
	LastLogins   []LastLogin    `json:"last_logins"`
	SudoCommands []SudoCommand  `json:"sudo_commands"`
	OpenPorts    []OpenPort     `json:"open_ports"`
	Users        []UserSecurity `json:"users"`
	Updates      []SystemUpdate `json:"updates"`
}

// Disk информация о диске
type Disk struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	FSType     string `json:"fs_type"`
	Total      uint64 `json:"total"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
	Percent    int64  `json:"percent"`
}

// Interface сетевая интерфейс
type Interface struct {
	Name    string   `json:"name"`
	MAC     string   `json:"mac"`
	IPs     []string `json:"ips"`
	MTU     int      `json:"mtu"`
	Speed   uint64   `json:"speed"`
	Flags   []string `json:"flags"`
	RxBytes uint64   `json:"rx_bytes"`
	TxBytes uint64   `json:"tx_bytes"`
}

// Connection сетевое соединение
type Connection struct {
	LocalAddr  string `json:"local_addr"`
	LocalPort  uint32 `json:"local_port"`
	RemoteAddr string `json:"remote_addr"`
	RemotePort uint32 `json:"remote_port"`
	Status     string `json:"status"`
	PID        int32  `json:"pid"`
	Process    string `json:"process"`
}

// Port слушающий порт
type Port struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process"`
	PID      int32  `json:"pid"`
}

// Bandwidth пропускная способность
type Bandwidth struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
}

// DNS информация
type DNS struct {
	Servers []string `json:"servers"`
	Domain  string   `json:"domain"`
	Search  []string `json:"search"`
}

// Firewall информация
type Firewall struct {
	Enabled bool     `json:"enabled"`
	Type    string   `json:"type"`
	Rules   []string `json:"rules"`
}

// Process процесс
type Process struct {
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

// FailedLogin неудачный логин
type FailedLogin struct {
	User      string `json:"user"`
	From      string `json:"from"`
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
}

// LastLogin последний логин
type LastLogin struct {
	User string `json:"user"`
	From string `json:"from"`
	When string `json:"when"`
}

// SudoCommand sudo команда
type SudoCommand struct {
	User    string `json:"user"`
	Command string `json:"command"`
	Time    string `json:"time"`
}

// OpenPort открытый порт
type OpenPort struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	State    string `json:"state"`
}

// UserSecurity пользователь
type UserSecurity struct {
	User         string `json:"user"`
	LastLogin    string `json:"last_login"`
	FailedLogins int    `json:"failed_logins"`
	HasSudo      bool   `json:"has_sudo"`
	Locked       bool   `json:"locked"`
}

// SystemUpdate системное обновление
type SystemUpdate struct {
	Package   string `json:"package"`
	Current   string `json:"current"`
	Available string `json:"available"`
	Security  bool   `json:"security"`
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
