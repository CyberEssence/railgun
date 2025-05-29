package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	API_HOST      = "http://localhost:8080"
	AGENT_VERSION = "1.2.0"
	AUTH_TOKEN    = "Bearer insert_your_token"
)

var (
	hostID    string
	hostname  string
	ipAddress string
)

type MemoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func initializeHostInfo() bool {
	// Получение имени хоста
	var buf [256]uint16
	size := uint32(len(buf))
	if err := windows.GetComputerNameEx(windows.ComputerNamePhysicalDnsFullyQualified, &buf[0], &size); err != nil {
		fmt.Printf("Failed to get hostname: %v\n", err)
		return false
	}
	hostname = windows.UTF16ToString(buf[:size])

	// Получение IP адреса
	cmd := exec.Command("powershell", "(Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.PrefixOrigin -ne 'WellKnown' } | Select-Object -First 1).IPAddress")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get IP address: %v\n", err)
		return false
	}
	ipAddress = strings.TrimSpace(string(output))
	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}

	// Генерация уникального ID хоста
	rand.Seed(time.Now().UnixNano())
	hostID = fmt.Sprintf("%s-%08x", hostname, rand.Uint32())

	return true
}

func sendRequest(method, endpoint string, data interface{}) error {
	url := API_HOST + endpoint
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", AUTH_TOKEN)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func getOSVersion() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Unknown"
	}
	defer key.Close()

	productName, _, _ := key.GetStringValue("ProductName")
	version, _, _ := key.GetStringValue("CurrentVersion")
	build, _, _ := key.GetStringValue("CurrentBuildNumber")

	return fmt.Sprintf("%s %s (Build %s)", productName, version, build)
}

func getTotalRAM() uint64 {
	var memStatus MemoryStatusEx
	memStatus.Length = uint32(unsafe.Sizeof(memStatus))
	ret, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(
		uintptr(unsafe.Pointer(&memStatus)),
	)
	if ret == 0 {
		return 0
	}
	return memStatus.TotalPhys / (1024 * 1024) // Возвращаем в МБ
}

func getProcessList() []map[string]interface{} {
	cmd := exec.Command("powershell", "Get-Process | Select-Object Id,Name,Path,Company,CPU,WorkingSet,StartTime | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var processes []map[string]interface{}
	if err := json.Unmarshal(output, &processes); err != nil {
		return nil
	}

	return processes
}

func getNetworkConnections() []map[string]interface{} {
	cmd := exec.Command("powershell", "Get-NetTCPConnection | Where-Object { $_.State -eq 'Established' } | Select-Object LocalAddress,LocalPort,RemoteAddress,RemotePort,State,OwningProcess | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var connections []map[string]interface{}
	if err := json.Unmarshal(output, &connections); err != nil {
		return nil
	}

	return connections
}

func getEventLogs(logName string) []map[string]interface{} {
	cmd := exec.Command("powershell", fmt.Sprintf("Get-EventLog -LogName %s -Newest 50 | Select-Object TimeGenerated,EntryType,Source,Message | ConvertTo-Json", logName))
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var events []map[string]interface{}
	if err := json.Unmarshal(output, &events); err != nil {
		return nil
	}

	return events
}

func getRegistryKey(path string) map[string]interface{} {
	var root registry.Key
	var subkey string

	switch {
	case strings.HasPrefix(path, "HKLM\\"):
		root = registry.LOCAL_MACHINE
		subkey = strings.TrimPrefix(path, "HKLM\\")
	case strings.HasPrefix(path, "HKCU\\"):
		root = registry.CURRENT_USER
		subkey = strings.TrimPrefix(path, "HKCU\\")
	default:
		return nil
	}

	key, err := registry.OpenKey(root, subkey, registry.READ)
	if err != nil {
		return nil
	}
	defer key.Close()

	values, err := key.ReadValueNames(0)
	if err != nil {
		return nil
	}

	result := make(map[string]interface{})
	for _, value := range values {
		data, _, err := key.GetStringValue(value)
		if err == nil {
			result[value] = data
		}
	}

	return map[string]interface{}{
		"path":   path,
		"values": result,
	}
}

// Отправка данных на эндпоинты SIEM

func sendHostInfo() error {
	data := map[string]interface{}{
		"host_id":    hostID,
		"hostname":   hostname,
		"ip_address": ipAddress,
		"os":         "Windows",
		"os_version": getOSVersion(),
		"status":     "active",
		"last_seen":  time.Now().UTC().Format(time.RFC3339),
	}

	return sendRequest("POST", "/api/hosts", data)
}

func sendTrafficData() error {
	connections := getNetworkConnections()
	if connections == nil {
		return fmt.Errorf("failed to get network connections")
	}

	data := map[string]interface{}{
		"host_id":     hostID,
		"connections": connections,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	return sendRequest("POST", "/api/traffic", data)
}

func sendArtifacts() error {
	// Реестр
	registryArtifacts := []map[string]interface{}{
		getRegistryKey(`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`),
		getRegistryKey(`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`),
		getRegistryKey(`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`),
	}

	// Процессы
	processes := getProcessList()

	data := map[string]interface{}{
		"host_id":            hostID,
		"registry_artifacts": registryArtifacts,
		"processes":          processes,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	}

	return sendRequest("POST", "/api/artifacts", data)
}

func sendSystemLogs() error {
	logs := map[string]interface{}{
		"host_id":     hostID,
		"system":      getEventLogs("System"),
		"security":    getEventLogs("Security"),
		"application": getEventLogs("Application"),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	return sendRequest("POST", "/api/logs/system", logs)
}

func sendNetworkLogs() error {
	// Симулируем сетевые логи
	logs := []map[string]interface{}{
		{
			"source_ip":      ipAddress,
			"destination_ip": "8.8.8.8",
			"protocol":       "TCP",
			"port":           53,
			"action":         "allowed",
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	data := map[string]interface{}{
		"host_id": hostID,
		"logs":    logs,
	}

	return sendRequest("POST", "/api/logs/network", data)
}

func sendRealtimeAnalysis() error {
	data := map[string]interface{}{
		"host_id":       hostID,
		"processes":     getProcessList(),
		"connections":   getNetworkConnections(),
		"event_logs":    getEventLogs("Security"),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"analysis_type": "realtime",
	}

	return sendRequest("POST", "/api/ai/analyze", data)
}

func sendAttackPatternsRequest() error {
	data := map[string]interface{}{
		"host_id": hostID,
		"query": map[string]interface{}{
			"category": "all",
			"severity": "high",
		},
	}

	return sendRequest("GET", "/api/ai/patterns", data)
}

func sendAPTTimelineRequest() error {
	data := map[string]interface{}{
		"host_id": hostID,
		"from":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"to":      time.Now().Format(time.RFC3339),
	}

	return sendRequest("GET", "/api/ai/apt-timeline", data)
}

func sendFileScan(filePath string) error {
	// В реальной реализации нужно добавить отправку файла
	data := map[string]interface{}{
		"host_id":   hostID,
		"file_path": filePath,
		"status":    "queued",
	}

	return sendRequest("POST", "/api/integration/scan", data)
}

func main() {
	fmt.Printf("SIEM Agent v%s starting...\n", AGENT_VERSION)

	if !initializeHostInfo() {
		fmt.Println("Failed to initialize host info")
		return
	}

	fmt.Printf("Host ID: %s\n", hostID)
	fmt.Printf("Hostname: %s\n", hostname)
	fmt.Printf("IP Address: %s\n", ipAddress)

	// Отправка данных на все эндпоинты
	tasks := []struct {
		name string
		fn   func() error
	}{
		{"Sending host info", sendHostInfo},
		{"Sending traffic data", sendTrafficData},
		{"Sending artifacts", sendArtifacts},
		{"Sending system logs", sendSystemLogs},
		{"Sending network logs", sendNetworkLogs},
		{"Performing realtime analysis", sendRealtimeAnalysis},
		{"Requesting attack patterns", sendAttackPatternsRequest},
		{"Requesting APT timeline", sendAPTTimelineRequest},
	}

	for _, task := range tasks {
		fmt.Printf("%s... ", task.name)
		if err := task.fn(); err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Println("Success")
		}
	}

	fmt.Println("Agent finished")
}
