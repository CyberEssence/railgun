package executor

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"linux-agent/internal/core/ports"
)

type FirewallExecutor struct{}

func NewFirewallExecutor() ports.IsolationExecutor {
	return &FirewallExecutor{}
}

// Isolate принимает IP адрес сервера для исключения из блокировки
func (e *FirewallExecutor) Isolate(serverIP string) error {
	log.Printf("Executing ISOLATION commands (iptables). Whitelisting IP: %s", serverIP)

	// Внимание: Требует прав root

	// Устанавливаем политики DROP
	if err := runCmd("iptables", "-P", "INPUT", "DROP"); err != nil {
		return err
	}
	if err := runCmd("iptables", "-P", "OUTPUT", "DROP"); err != nil {
		return err
	}
	if err := runCmd("iptables", "-P", "FORWARD", "DROP"); err != nil {
		return err
	}

	// Разрешаем Loopback
	if err := runCmd("iptables", "-A", "INPUT", "-i", "lo", "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := runCmd("iptables", "-A", "OUTPUT", "-o", "lo", "-j", "ACCEPT"); err != nil {
		return err
	}

	// Разрешаем доступ к SIEM серверу
	if serverIP != "" {
		// Разрешаем исходящие соединения к серверу
		if err := runCmd("iptables", "-A", "OUTPUT", "-d", serverIP, "-j", "ACCEPT"); err != nil {
			return err
		}
		// Разрешаем входящие ответы от сервера
		if err := runCmd("iptables", "-A", "INPUT", "-s", serverIP, "-j", "ACCEPT"); err != nil {
			return err
		}
	}

	// Разрешаем установленные соединения (важно для TCP handshake)
	if err := runCmd("iptables", "-A", "INPUT", "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT"); err != nil {
		return err
	}

	log.Println("Isolation applied successfully.")
	return nil
}

func (e *FirewallExecutor) Unisolate() error {
	log.Println("Executing UNISOLATION commands (flushing rules)...")

	// Сброс политик в ACCEPT
	runCmd("iptables", "-P", "INPUT", "ACCEPT")
	runCmd("iptables", "-P", "OUTPUT", "ACCEPT")
	runCmd("iptables", "-P", "FORWARD", "ACCEPT")

	// Примечание: Здесь мы не делаем iptables -F, чтобы не сбросить другие правила,
	// но для полной разблокировки достаточно переключить политики.

	log.Println("Unisolation applied (policies set to ACCEPT).")
	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cmd '%s %s' failed: %v, output: %s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}
