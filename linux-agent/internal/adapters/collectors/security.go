package collectors

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"linux-agent/internal/core/domain"
)

// SecurityCollector сборщик информации о безопасности
type SecurityCollector struct {
	BaseCollector
}

// NewSecurityCollector создает новый коллектор безопасности
func NewSecurityCollector(hostID, hostname string, enabled bool) *SecurityCollector {
	return &SecurityCollector{
		BaseCollector: BaseCollector{
			HostID:     hostID,
			Hostname:   hostname,
			NameVal:    "security",
			EnabledVal: enabled,
		},
	}
}

// Collect собирает информацию о безопасности
func (c *SecurityCollector) Collect(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info := &domain.SecurityInfo{
		BaseMetric: domain.BaseMetric{
			HostID:    c.HostID,
			Hostname:  c.Hostname,
			Timestamp: time.Now().UTC(),
			Type:      "security",
		},
	}

	// Неудачные логины
	c.collectFailedLogins(ctx, info)

	// Последние логины
	c.collectLastLogins(ctx, info)

	// Открытые порты
	c.collectOpenPorts(ctx, info)

	// Информация о пользователях
	c.collectUserSecurity(ctx, info)

	return info, nil
}

func (c *SecurityCollector) collectFailedLogins(ctx context.Context, info *domain.SecurityInfo) {
	// Проверяем разные лог-файлы
	logFiles := []string{"/var/log/auth.log", "/var/log/secure"}

	for _, logFile := range logFiles {
		if _, err := os.Stat(logFile); err == nil {
			cmd := exec.CommandContext(ctx, "tail", "-n", "50", logFile)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") {
					login := domain.FailedLogin{
						Service: "ssh",
					}

					parts := strings.Fields(line)
					for i, part := range parts {
						if part == "for" && i+1 < len(parts) {
							login.User = parts[i+1]
						}
						if part == "from" && i+1 < len(parts) {
							login.From = parts[i+1]
						}
					}

					if len(parts) >= 3 {
						login.Timestamp = parts[0] + " " + parts[1] + " " + parts[2]
					}

					info.FailedLogins = append(info.FailedLogins, login)
				}
			}
			break
		}
	}
}

func (c *SecurityCollector) collectLastLogins(ctx context.Context, info *domain.SecurityInfo) {
	cmd := exec.CommandContext(ctx, "last", "-n", "10")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[0] != "" && parts[0] != "wtmp" {
			info.LastLogins = append(info.LastLogins, domain.LastLogin{
				User: parts[0],
				From: parts[2],
				When: strings.Join(parts[3:], " "),
			})
		}
	}
}

func (c *SecurityCollector) collectOpenPorts(ctx context.Context, info *domain.SecurityInfo) {
	cmd := exec.CommandContext(ctx, "ss", "-tuln")
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
		if len(parts) >= 5 && parts[0] != "Netid" {
			local := parts[4]
			localParts := strings.Split(local, ":")
			if len(localParts) == 2 {
				port, err := strconv.Atoi(localParts[1])
				if err == nil {
					openPort := domain.OpenPort{
						Port:     port,
						Protocol: parts[0],
						State:    parts[1],
					}

					// Определяем сервис
					switch port {
					case 22:
						openPort.Service = "ssh"
					case 80:
						openPort.Service = "http"
					case 443:
						openPort.Service = "https"
					case 3306:
						openPort.Service = "mysql"
					case 5432:
						openPort.Service = "postgresql"
					default:
						openPort.Service = "unknown"
					}

					info.OpenPorts = append(info.OpenPorts, openPort)
				}
			}
		}
	}
}

func (c *SecurityCollector) collectUserSecurity(ctx context.Context, info *domain.SecurityInfo) {
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 7 {
			user := domain.UserSecurity{
				User: parts[0],
			}

			// Проверяем, заблокирован ли пользователь
			if strings.Contains(parts[1], "!") || strings.Contains(parts[1], "*") {
				user.Locked = true
			}

			info.Users = append(info.Users, user)
		}
	}
}
