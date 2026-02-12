package collector

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	gopsnet "github.com/shirou/gopsutil/v3/net"
)

type NetworkInfo struct {
	Timestamp time.Time `json:"timestamp"`
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Type      string    `json:"type"` // "network_info"

	Interfaces     []InterfaceInfo  `json:"interfaces"`
	Connections    []ConnectionInfo `json:"connections"`
	ListeningPorts []PortInfo       `json:"listening_ports"`
	Bandwidth      BandwidthInfo    `json:"bandwidth"`
	Firewall       FirewallInfo     `json:"firewall"`
	DNS            DNSInfo          `json:"dns"`
}

type InterfaceInfo struct {
	Name  string   `json:"name"`
	MAC   string   `json:"mac"`
	IPs   []string `json:"ips"`
	MTU   int      `json:"mtu"`
	Speed uint64   `json:"speed"`
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

type FirewallInfo struct {
	Rules   []FirewallRule `json:"rules"`
	Enabled bool           `json:"enabled"`
	Type    string         `json:"type"` // iptables, nftables, firewalld
}

type FirewallRule struct {
	Chain       string `json:"chain"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Action      string `json:"action"`
}

type DNSInfo struct {
	Servers []string `json:"servers"`
	Domain  string   `json:"domain"`
	Search  []string `json:"search"`
}

func CollectNetworkInfo(hostID, hostname string) (*NetworkInfo, error) {
	info := &NetworkInfo{
		Timestamp: time.Now().UTC(),
		HostID:    hostID,
		Hostname:  hostname,
		Type:      "network_info",
	}

	// Сетевые интерфейсы
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		ips := []string{}
		for _, addr := range addrs {
			ips = append(ips, addr.String())
		}

		// Конвертация MAC в строку
		mac := iface.HardwareAddr.String()

		// Конвертация флагов в строки
		flags := []string{}
		if iface.Flags&net.FlagUp != 0 {
			flags = append(flags, "UP")
		}
		if iface.Flags&net.FlagBroadcast != 0 {
			flags = append(flags, "BROADCAST")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			flags = append(flags, "LOOPBACK")
		}
		if iface.Flags&net.FlagPointToPoint != 0 {
			flags = append(flags, "POINTTOPOINT")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			flags = append(flags, "MULTICAST")
		}

		info.Interfaces = append(info.Interfaces, InterfaceInfo{
			Name:  iface.Name,
			MAC:   mac,
			IPs:   ips,
			MTU:   iface.MTU,
			Flags: flags,
		})
	}

	// Активные соединения через gopsutil
	conns, _ := gopsnet.Connections("all")
	for _, conn := range conns {
		info.Connections = append(info.Connections, ConnectionInfo{
			LocalAddr:  conn.Laddr.IP,
			LocalPort:  conn.Laddr.Port,
			RemoteAddr: conn.Raddr.IP,
			RemotePort: conn.Raddr.Port,
			Status:     conn.Status,
			PID:        conn.Pid,
			Process:    getProcessName(conn.Pid),
		})
	}

	// Слушающие порты
	info.collectListeningPorts()

	// Статистика сети через gopsutil
	stats, _ := gopsnet.IOCounters(true)
	for _, stat := range stats {
		info.Bandwidth.BytesSent += stat.BytesSent
		info.Bandwidth.BytesRecv += stat.BytesRecv
		info.Bandwidth.PacketsSent += stat.PacketsSent
		info.Bandwidth.PacketsRecv += stat.PacketsRecv
		info.Bandwidth.ErrorsIn += stat.Errin
		info.Bandwidth.ErrorsOut += stat.Errout
		info.Bandwidth.DropsIn += stat.Dropin
		info.Bandwidth.DropsOut += stat.Dropout
	}

	// Firewall rules
	info.collectFirewallInfo()

	// DNS информация
	info.collectDNSInfo()

	return info, nil
}

func (n *NetworkInfo) collectListeningPorts() {
	// Используем ss для получения слушающих портов
	cmd := exec.Command("ss", "-tulnp")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 6 {
			// Формат: Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process
			localAddr := parts[4]
			localParts := strings.Split(localAddr, ":")
			if len(localParts) == 2 {
				portStr := localParts[1]
				port, err := strconv.ParseInt(portStr, 10, 32)
				if err == nil {
					portInfo := PortInfo{
						Port:     int32(port),
						Protocol: parts[0],
					}

					// Парсим процесс из последнего столбца
					if len(parts) > 6 {
						processInfo := parts[6]
						// users:(("nginx",pid=123,fd=3))
						if strings.Contains(processInfo, "pid=") {
							pidStart := strings.Index(processInfo, "pid=") + 4
							pidEnd := strings.Index(processInfo[pidStart:], ",")
							if pidEnd == -1 {
								pidEnd = strings.Index(processInfo[pidStart:], ")")
							}
							if pidEnd > 0 {
								pidStr := processInfo[pidStart : pidStart+pidEnd]
								pid, _ := strconv.ParseInt(pidStr, 10, 32)
								portInfo.PID = int32(pid)

								// Имя процесса
								if strings.Contains(processInfo, "\"") {
									nameStart := strings.Index(processInfo, "\"") + 1
									nameEnd := strings.Index(processInfo[nameStart:], "\"")
									if nameEnd > 0 {
										portInfo.Process = processInfo[nameStart : nameStart+nameEnd]
									}
								}
							}
						}
					}

					n.ListeningPorts = append(n.ListeningPorts, portInfo)
				}
			}
		}
	}
}

func (n *NetworkInfo) collectFirewallInfo() {
	// Пробуем разные фаерволы
	firewallTypes := []struct {
		cmd    string
		args   []string
		parser func(string) []FirewallRule
	}{
		{"iptables", []string{"-L", "-n", "-v"}, parseIptablesOutput},
		{"nft", []string{"list", "ruleset"}, parseNftablesOutput},
		{"firewall-cmd", []string{"--list-all"}, parseFirewalldOutput},
	}

	for _, fw := range firewallTypes {
		cmd := exec.Command(fw.cmd, fw.args...)
		output, err := cmd.Output()
		if err == nil {
			n.Firewall.Enabled = true
			n.Firewall.Type = fw.cmd
			n.Firewall.Rules = fw.parser(string(output))
			break
		}
	}
}

func (n *NetworkInfo) collectDNSInfo() {
	// Чтение resolv.conf
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				n.DNS.Servers = append(n.DNS.Servers, parts[1])
			}
		} else if strings.HasPrefix(line, "domain") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				n.DNS.Domain = parts[1]
			}
		} else if strings.HasPrefix(line, "search") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				n.DNS.Search = parts[1:]
			}
		}
	}
}

func parseIptablesOutput(output string) []FirewallRule {
	// Упрощенный парсинг iptables
	rules := []FirewallRule{}
	lines := strings.Split(output, "\n")

	var currentChain string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Chain") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				currentChain = parts[1]
			}
		} else if strings.HasPrefix(line, "[") || line == "" {
			continue
		} else if currentChain != "" {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				rules = append(rules, FirewallRule{
					Chain:       currentChain,
					Protocol:    parts[3],
					Source:      parts[7],
					Destination: parts[8],
					Action:      parts[1],
				})
			}
		}
	}

	return rules
}

func parseNftablesOutput(output string) []FirewallRule {
	// Базовая обработка nftables
	return []FirewallRule{}
}

func parseFirewalldOutput(output string) []FirewallRule {
	// Базовая обработка firewalld
	return []FirewallRule{}
}

func getProcessName(pid int32) string {
	if pid == 0 {
		return "kernel"
	}

	// Чтение /proc/[pid]/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := os.ReadFile(commPath)
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(data))
}

func (n *NetworkInfo) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}
