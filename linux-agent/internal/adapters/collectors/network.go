package collectors

import (
	"context"

	"os/exec"
	"strconv"
	"strings"
	"time"

	"linux-agent/internal/core/domain"

	net "github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector сборщик сетевой информации
type NetworkCollector struct {
	BaseCollector
}

// NewNetworkCollector создает новый сетевой коллектор
func NewNetworkCollector(hostID, hostname string, enabled bool) *NetworkCollector {
	return &NetworkCollector{
		BaseCollector: BaseCollector{
			HostID:     hostID,
			Hostname:   hostname,
			NameVal:    "network",
			EnabledVal: enabled,
		},
	}
}

// Collect собирает сетевую информацию
func (c *NetworkCollector) Collect(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info := &domain.NetworkInfo{
		BaseMetric: domain.BaseMetric{
			HostID:    c.HostID,
			Hostname:  c.Hostname,
			Timestamp: time.Now().UTC(),
			Type:      "network",
		},
	}

	// Сетевые интерфейсы
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		addrs := make([]string, len(iface.Addrs))
		for i, addr := range iface.Addrs {
			addrs[i] = addr.Addr
		}

		info.Interfaces = append(info.Interfaces, domain.Interface{
			Name:  iface.Name,
			MAC:   iface.HardwareAddr,
			IPs:   addrs,
			MTU:   iface.MTU,
			Flags: iface.Flags,
		})
	}

	// Активные соединения
	conns, _ := net.Connections("all")
	for _, conn := range conns {
		info.Connections = append(info.Connections, domain.Connection{
			LocalAddr:  conn.Laddr.IP,
			LocalPort:  conn.Laddr.Port,
			RemoteAddr: conn.Raddr.IP,
			RemotePort: conn.Raddr.Port,
			Status:     conn.Status,
			PID:        conn.Pid,
		})
	}

	// Слушающие порты
	c.collectListeningPorts(ctx, info)

	// Статистика сети
	stats, _ := net.IOCounters(true)
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

	// DNS информация
	c.collectDNSInfo(ctx, info)

	return info, nil
}

func (c *NetworkCollector) collectListeningPorts(ctx context.Context, info *domain.NetworkInfo) {
	cmd := exec.CommandContext(ctx, "ss", "-tulnp")
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
			localAddr := parts[4]
			localParts := strings.Split(localAddr, ":")
			if len(localParts) == 2 {
				portStr := localParts[1]
				port, err := strconv.ParseInt(portStr, 10, 32)
				if err == nil {
					info.ListeningPorts = append(info.ListeningPorts, domain.Port{
						Port:     int32(port),
						Protocol: parts[0],
					})
				}
			}
		}
	}
}

func (c *NetworkCollector) collectDNSInfo(ctx context.Context, info *domain.NetworkInfo) {
	// В реальном приложении здесь парсинг /etc/resolv.conf
	// Для простоты оставим заглушку
	info.DNS = domain.DNS{
		Servers: []string{"8.8.8.8", "1.1.1.1"},
		Domain:  "local",
	}
}
