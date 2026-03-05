package collectors

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"linux-agent/internal/core/domain"

	net "github.com/shirou/gopsutil/v3/net"

	netrun "net"
)

// NetworkCollector сборщик сетевой информации
type NetworkCollector struct {
	BaseCollector
	prevStats map[string]net.IOCountersStat
	mu        sync.Mutex
}

func NewNetworkCollector(hostID, hostname string, enabled bool) *NetworkCollector {
	return &NetworkCollector{
		BaseCollector: BaseCollector{
			HostID:     hostID,
			Hostname:   hostname,
			NameVal:    "network",
			EnabledVal: enabled,
		},
		prevStats: make(map[string]net.IOCountersStat),
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

	// Сетевые интерфейсы (добавлена обработка ошибок и контекста)
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		// Логируем ошибку, но продолжаем работу, если это возможно
		fmt.Printf("Error getting interfaces: %v\n", err)
	} else {
		for _, iface := range interfaces {
			addrs := make([]string, 0, len(iface.Addrs))
			for _, addr := range iface.Addrs {
				addrs = append(addrs, addr.Addr)
			}

			info.Interfaces = append(info.Interfaces, domain.Interface{
				Name:  iface.Name,
				MAC:   iface.HardwareAddr,
				IPs:   addrs,
				MTU:   iface.MTU,
				Flags: iface.Flags,
			})
		}
	}

	// Активные соединения (используем WithContext)
	conns, err := net.ConnectionsWithContext(ctx, "all")
	if err != nil {
		fmt.Printf("Error getting connections: %v\n", err)
	} else {
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
	}

	// Слушающие порты
	c.collectListeningPorts(ctx, info)

	// Статистика сети (Трафик)
	c.collectTrafficStats(ctx, info)

	// DNS информация
	c.collectDNSInfo(ctx, info)

	return info, nil
}

// collectTrafficStats вычисляет дельту трафика за интервал
func (c *NetworkCollector) collectTrafficStats(ctx context.Context, info *domain.NetworkInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		fmt.Printf("Error getting IO counters: %v\n", err)
		return
	}

	var totalBytesSentDelta uint64
	var totalBytesRecvDelta uint64
	var totalPacketsSentDelta uint64
	var totalPacketsRecvDelta uint64
	var totalErrInDelta uint64
	var totalErrOutDelta uint64
	var totalDropInDelta uint64
	var totalDropOutDelta uint64

	for _, stat := range stats {
		if stat.Name == "lo" {
			continue
		}

		prev, exists := c.prevStats[stat.Name]

		// Суммируем дельты
		if exists {
			totalBytesSentDelta += safeSub(stat.BytesSent, prev.BytesSent)
			totalBytesRecvDelta += safeSub(stat.BytesRecv, prev.BytesRecv)
			totalPacketsSentDelta += safeSub(stat.PacketsSent, prev.PacketsSent)
			totalPacketsRecvDelta += safeSub(stat.PacketsRecv, prev.PacketsRecv)
			totalErrInDelta += safeSub(stat.Errin, prev.Errin)
			totalErrOutDelta += safeSub(stat.Errout, prev.Errout)
			totalDropInDelta += safeSub(stat.Dropin, prev.Dropin)
			totalDropOutDelta += safeSub(stat.Dropout, prev.Dropout)
		}

		// Обновляем сохраненное состояние
		c.prevStats[stat.Name] = stat
	}

	// Записываем ТОЛЬКО дельту в структуру
	info.Bandwidth = domain.Bandwidth{
		BytesSent:   totalBytesSentDelta,
		BytesRecv:   totalBytesRecvDelta,
		PacketsSent: totalPacketsSentDelta,
		PacketsRecv: totalPacketsRecvDelta,
		ErrorsIn:    totalErrInDelta,
		ErrorsOut:   totalErrOutDelta,
		DropsIn:     totalDropInDelta,
		DropsOut:    totalDropOutDelta,
	}
}

// safeSub безопасное вычитание с защитой от отрицательных чисел при перезагрузке счетчиков
func safeSub(current, prev uint64) uint64 {
	if current >= prev {
		return current - prev
	}
	return current
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
		// Format: Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port
		if len(parts) >= 6 {
			localAddrRaw := parts[4]

			host, portStr, err := netrun.SplitHostPort(localAddrRaw)
			if err != nil {
				lastColon := strings.LastIndex(localAddrRaw, ":")
				if lastColon != -1 {
					host = localAddrRaw[:lastColon]
					portStr = localAddrRaw[lastColon+1:]
				}
			}

			port, err := strconv.ParseInt(portStr, 10, 32)
			if err == nil {
				info.ListeningPorts = append(info.ListeningPorts, domain.Port{
					Port:     int32(port),
					Protocol: parts[0],
					Address:  host,
				})
			}
		}
	}
}

func (c *NetworkCollector) collectDNSInfo(ctx context.Context, info *domain.NetworkInfo) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		info.DNS = domain.DNS{}
		return
	}
	defer file.Close()

	var servers []string
	var searchDomain string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				servers = append(servers, fields[1])
			}
		} else if strings.HasPrefix(line, "domain") || strings.HasPrefix(line, "search") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				searchDomain = fields[1]
			}
		}
	}

	info.DNS = domain.DNS{
		Servers: servers,
		Domain:  searchDomain,
	}
}
