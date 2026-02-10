package collector

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SecurityInfo struct {
	Timestamp time.Time `json:"timestamp"`
	HostID    string    `json:"host_id"`
	Hostname  string    `json:"hostname"`
	Type      string    `json:"type"` // "security_info"

	FailedLogins []FailedLogin  `json:"failed_logins"`
	LastLogins   []LastLogin    `json:"last_logins"`
	SudoCommands []SudoCommand  `json:"sudo_commands"`
	OpenPorts    []OpenPort     `json:"open_ports"`
	Users        []UserSecurity `json:"users"`
	Updates      []SystemUpdate `json:"updates"`
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

type SecurityCollector struct {
	hostID   string
	hostname string
}

func NewSecurityCollector(hostID, hostname string) *SecurityCollector {
	return &SecurityCollector{
		hostID:   hostID,
		hostname: hostname,
	}
}

func (s *SecurityCollector) Name() string  { return "security" }
func (s *SecurityCollector) Enabled() bool { return true }

func (s *SecurityCollector) Collect() (interface{}, error) {
	return CollectSecurityInfo(s.hostID, s.hostname)
}

func CollectSecurityInfo(hostID, hostname string) (*SecurityInfo, error) {
	info := &SecurityInfo{
		Timestamp: time.Now().UTC(),
		HostID:    hostID,
		Hostname:  hostname,
		Type:      "security_info",
	}

	// Собираем данные о неудачных логинах
	info.collectFailedLogins()

	// Последние логины
	info.collectLastLogins()

	// Sudo команды
	info.collectSudoCommands()

	// Открытые порты
	info.collectOpenPorts()

	// Информация о пользователях
	info.collectUserSecurity()

	// Доступные обновления
	info.collectSystemUpdates()

	return info, nil
}

func (s *SecurityInfo) collectFailedLogins() {
	// Парсим /var/log/auth.log или /var/log/secure
	logFiles := []string{"/var/log/auth.log", "/var/log/secure"}

	for _, logFile := range logFiles {
		if _, err := os.Stat(logFile); err == nil {
			// Читаем последние 100 строк
			cmd := exec.Command("tail", "-n", "100", logFile)
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") {
					parts := strings.Fields(line)
					if len(parts) > 10 {
						login := FailedLogin{}

						// Ищем пользователя
						for i, part := range parts {
							if part == "for" && i+1 < len(parts) {
								login.User = parts[i+1]
							}
							if part == "from" && i+1 < len(parts) {
								login.From = parts[i+1]
							}
						}

						// Время
						if len(parts) >= 3 {
							login.Timestamp = parts[0] + " " + parts[1] + " " + parts[2]
						}

						login.Service = "ssh"
						s.FailedLogins = append(s.FailedLogins, login)
					}
				}
			}
			break
		}
	}
}

func (s *SecurityInfo) collectLastLogins() {
	// Команда last
	cmd := exec.Command("last", "-n", "10")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[0] != "" && parts[0] != "wtmp" {
			login := LastLogin{
				User: parts[0],
				From: parts[2],
				When: strings.Join(parts[3:], " "),
			}
			s.LastLogins = append(s.LastLogins, login)
		}
	}
}

func (s *SecurityInfo) collectSudoCommands() {
	// Парсим /var/log/auth.log на sudo команды
	logFile := "/var/log/auth.log"
	if _, err := os.Stat(logFile); err != nil {
		return
	}

	cmd := exec.Command("grep", "sudo", logFile)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if count >= 10 {
			break
		}

		if strings.Contains(line, "COMMAND=") {
			parts := strings.Fields(line)
			if len(parts) > 10 {
				sudo := SudoCommand{}

				// Ищем пользователя и команду
				for i, part := range parts {
					if part == "user=" && i+1 < len(parts) {
						sudo.User = strings.Trim(parts[i+1], ",")
					}
					if strings.HasPrefix(part, "COMMAND=") {
						sudo.Command = strings.TrimPrefix(part, "COMMAND=")
						sudo.Command = strings.Trim(sudo.Command, "\"")
					}
				}

				// Время
				if len(parts) >= 3 {
					sudo.Time = parts[0] + " " + parts[1] + " " + parts[2]
				}

				s.SudoCommands = append(s.SudoCommands, sudo)
				count++
			}
		}
	}
}

func (s *SecurityInfo) collectOpenPorts() {
	// Используем ss для проверки открытых портов
	cmd := exec.Command("ss", "-tuln")
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
					openPort := OpenPort{
						Port:     port,
						Protocol: parts[0],
						State:    parts[1],
					}

					// Определяем сервис
					if port == 22 {
						openPort.Service = "ssh"
					} else if port == 80 {
						openPort.Service = "http"
					} else if port == 443 {
						openPort.Service = "https"
					} else if port == 3306 {
						openPort.Service = "mysql"
					} else if port == 5432 {
						openPort.Service = "postgresql"
					} else {
						openPort.Service = "unknown"
					}

					s.OpenPorts = append(s.OpenPorts, openPort)
				}
			}
		}
	}
}

func (s *SecurityInfo) collectUserSecurity() {
	// Читаем /etc/passwd
	data, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 7 {
			user := UserSecurity{
				User: parts[0],
			}

			// Проверяем, заблокирован ли пользователь
			if strings.Contains(parts[1], "!") || strings.Contains(parts[1], "*") {
				user.Locked = true
			}

			// Проверяем, есть ли у пользователя sudo права
			cmd := exec.Command("grep", "-q", "^"+user.User+".*ALL$", "/etc/sudoers")
			if cmd.Run() == nil {
				user.HasSudo = true
			}

			s.Users = append(s.Users, user)
		}
	}
}

func (s *SecurityInfo) collectSystemUpdates() {
	// Проверяем доступные обновления (для apt/dnf/yum)
	packageManagers := []struct {
		cmd  string
		args []string
	}{
		{"apt", []string{"list", "--upgradable"}},
		{"dnf", []string{"check-update"}},
		{"yum", []string{"check-update"}},
	}

	for _, pm := range packageManagers {
		cmd := exec.Command(pm.cmd, pm.args...)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					update := SystemUpdate{
						Package:   parts[0],
						Current:   parts[1],
						Available: parts[2],
						Security:  strings.Contains(line, "security") || strings.Contains(line, "Security"),
					}
					s.Updates = append(s.Updates, update)
				}
			}
			break
		}
	}
}

func (s *SecurityInfo) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}
