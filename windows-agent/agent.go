package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	BUFFER_SIZE    = 4096
	API_HOST       = "http://localhost:8080" // Change to your server address
	AGENT_VERSION  = "1.0.0"
	MAX_COMP_NAME  = 15
	computernameEx = 5                     // ComputerNamePhysicalDnsFullyQualified
	AUTH_TOKEN     = "valid_token_example" // Используем токен из примера
)

// Global variables
var (
	hostID    string
	hostname  string
	ipAddress string
)

// Structs for Windows API
type MIB_TCPSTATS struct {
	DwRtoAlgorithm uint32
	DwRtoMin       uint32
	DwRtoMax       uint32
	DwMaxConn      uint32
	DwActiveOpens  uint32
	DwPassiveOpens uint32
	DwAttemptFails uint32
	DwEstabResets  uint32
	DwCurrEstab    uint32
	DwInSegs       uint32
	DwOutSegs      uint32
	DwRetransSegs  uint32
	DwInErrs       uint32
	DwOutRsts      uint32
	DwNumConns     uint32
}

// Windows API functions
var (
	modIphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetTcpStatistics   = modIphlpapi.NewProc("GetTcpStatistics")
	procGetComputerNameExW = modKernel32.NewProc("GetComputerNameExW")
)

// Function to send data to the API server
func sendDataToAPI(endpoint string, data interface{}) bool {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return false
	}

	url := API_HOST + endpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", AUTH_TOKEN) // Добавляем токен авторизации

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Printf("Sending data to %s: %s\n", url, string(jsonData))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("API request failed. Status: %d, Response: %s\n", resp.StatusCode, string(body))
	return false
}

// GET запрос к API
func getDataFromAPI(endpoint string) ([]byte, error) {
	url := API_HOST + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", AUTH_TOKEN) // Добавляем токен авторизации

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Printf("Sending GET request to %s\n", url)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Initialize host information
func initializeHostInfo() bool {
	// Get hostname
	var size uint32 = MAX_COMP_NAME + 1
	var buf [MAX_COMP_NAME + 1]uint16
	ret, _, _ := procGetComputerNameExW.Call(
		uintptr(computernameEx),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)))

	if ret == 0 {
		fmt.Printf("Failed to get hostname. Error: %d\n", windows.GetLastError())
		return false
	}

	hostname = windows.UTF16ToString(buf[:])

	// Get IP address
	cmd := exec.Command("powershell", "-Command", "Get-NetIPAddress | Where-Object {$_.AddressFamily -eq 'IPv4' -and $_.PrefixOrigin -ne 'WellKnown'} | Select-Object -First 1 -ExpandProperty IPAddress")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get IP address: %v\n", err)
		return false
	}

	ipAddress = strings.TrimSpace(string(output))
	if ipAddress == "" {
		ipAddress = "127.0.0.1" // Default to localhost if no IP found
	}

	// Generate host_id
	rand.Seed(time.Now().UnixNano())
	hostID = fmt.Sprintf("%s-%08x", hostname, rand.Uint32())

	return true
}

// Register host with the SIEM server
func registerHost() bool {
	currentTime := time.Now().UTC().Format(time.RFC3339)

	hostData := map[string]string{
		"id":          hostID,
		"hostname":    hostname,
		"ip_address":  ipAddress,
		"os_version":  "Windows 10",
		"status":      "active",
		"last_seen":   currentTime,
		"description": "Windows Agent Host",
	}

	return sendDataToAPI("/api/v1/hosts/", hostData)
}

// Collect network traffic statistics
func collectNetworkTraffic() bool {
	var tcpStats MIB_TCPSTATS
	ret, _, _ := procGetTcpStatistics.Call(uintptr(unsafe.Pointer(&tcpStats)))

	if ret != 0 {
		fmt.Printf("GetTcpStatistics failed with error: %d\n", ret)
		return false
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)

	trafficData := map[string]interface{}{
		"host_id":      hostID,
		"timestamp":    currentTime,
		"src_ip":       ipAddress,
		"dst_ip":       "8.8.8.8",
		"src_port":     12345,
		"dst_port":     80,
		"protocol":     "TCP",
		"bytes_sent":   tcpStats.DwOutSegs * 1460,
		"bytes_recv":   tcpStats.DwInSegs * 1460,
		"packets_sent": tcpStats.DwOutSegs,
		"packets_recv": tcpStats.DwInSegs,
		"duration":     10.5,
	}

	return sendDataToAPI("/api/v1/traffic/", trafficData)
}

// Collect registry artifacts
func collectRegistryArtifacts() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Printf("Failed to open registry key: %v\n", err)
		return false
	}
	defer key.Close()

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		fmt.Printf("Failed to read registry value: %v\n", err)
		return false
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)

	artifactData := map[string]interface{}{
		"host_id":   hostID,
		"timestamp": currentTime,
		"type":      "registry",
		"path":      `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\ProductName`,
		"value":     productName,
		"size":      len(productName),
		"owner":     "SYSTEM",
		"metadata": map[string]string{
			"registry_type": "REG_SZ",
			"last_modified": currentTime,
		},
	}

	return sendDataToAPI("/api/v1/artifacts/", artifactData)
}

// Collect process artifacts
func collectProcessArtifacts() bool {
	cmd := exec.Command("powershell", "-Command", "Get-Process | Select-Object -First 5 | ForEach-Object { [PSCustomObject]@{Name=$_.Name; Path=$_.Path; ID=$_.Id; Memory=$_.WorkingSet64} } | ConvertTo-Json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get process info: %v\n", err)
		return false
	}

	var processes []map[string]interface{}
	if err := json.Unmarshal(output, &processes); err != nil {
		fmt.Printf("Failed to parse process JSON: %v\n", err)
		return false
	}

	success := true
	for _, process := range processes {
		currentTime := time.Now().UTC().Format(time.RFC3339)
		name := fmt.Sprintf("%v", process["Name"])
		path := fmt.Sprintf("%v", process["Path"])

		artifactData := map[string]interface{}{
			"host_id":   hostID,
			"timestamp": currentTime,
			"type":      "process",
			"path":      path,
			"value":     name,
			"size":      process["Memory"],
			"owner":     "SYSTEM",
			"metadata": map[string]interface{}{
				"process_id": process["ID"],
				"started_at": currentTime,
			},
		}

		if !sendDataToAPI("/api/v1/artifacts/", artifactData) {
			success = false
		}
	}

	return success
}

// Collect network logs
func collectNetworkLogs() bool {
	currentTime := time.Now().UTC().Format(time.RFC3339)

	// Симулируем сетевые логи
	logTypes := []string{"firewall", "ids", "proxy"}
	severities := []string{"info", "warning", "critical"}

	for i := 0; i < 3; i++ {
		logType := logTypes[rand.Intn(len(logTypes))]
		severity := severities[rand.Intn(len(severities))]

		logData := map[string]interface{}{
			"source_ip":      fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"destination_ip": fmt.Sprintf("10.0.0.%d", rand.Intn(254)+1),
			"protocol":       "TCP",
			"log_type":       logType,
			"raw_data":       fmt.Sprintf("Sample %s log entry", logType),
			"timestamp":      currentTime,
			"severity":       severity,
		}

		if !sendDataToAPI("/api/v1/logs/network", logData) {
			return false
		}
	}

	return true
}

// Collect system logs
func collectSystemLogs() bool {
	currentTime := time.Now().UTC().Format(time.RFC3339)

	// Симулируем системные логи
	logData := map[string]interface{}{
		"host_id":   hostID,
		"timestamp": currentTime,
		"type":      "system",
		"source":    "Windows Event Log",
		"severity":  "info",
		"data": map[string]interface{}{
			"event_id":   4624,
			"message":    "An account was successfully logged on",
			"user":       "SYSTEM",
			"process_id": 1234,
		},
	}

	return sendDataToAPI("/api/v1/logs/system", logData)
}

// Scan file with VirusTotal
func scanFileWithVirusTotal(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", filePath, err)
		return false
	}
	defer file.Close()

	// Создаем multipart форму для отправки файла
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		return false
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Failed to copy file content: %v\n", err)
		return false
	}
	writer.Close()

	// Отправляем запрос
	url := API_HOST + "/api/v1/integrations/scan-file"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return false
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", AUTH_TOKEN)

	client := &http.Client{
		Timeout: 60 * time.Second, // Увеличиваем таймаут для сканирования файла
	}

	fmt.Printf("Scanning file %s with VirusTotal...\n", filePath)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP request failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Scan result: %s\n", string(respBody))
		return true
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return false
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("Scan result: %s\n", string(respBody))
		return true
	}

	fmt.Printf("API request failed. Status: %d, Response: %s\n", resp.StatusCode, string(respBody))
	return false
}

// Get threat patterns
func getThreatPatterns() {
	data, err := getDataFromAPI("/api/v1/detection/patterns")
	if err != nil {
		fmt.Printf("Failed to get threat patterns: %v\n", err)
		return
	}

	fmt.Printf("Threat patterns: %s\n", string(data))
}

// Get traffic statistics
func getTrafficStats() {
	endpoint := fmt.Sprintf("/api/v1/traffic/stats/host/%s?from=%s&to=%s",
		hostID,
		time.Now().Add(-24*time.Hour).Format(time.RFC3339),
		time.Now().Format(time.RFC3339))

	data, err := getDataFromAPI(endpoint)
	if err != nil {
		fmt.Printf("Failed to get traffic stats: %v\n", err)
		return
	}

	fmt.Printf("Traffic stats: %s\n", string(data))
}

// Get dashboard stats
func getDashboardStats() {
	data, err := getDataFromAPI("/api/v1/dashboard/stats")
	if err != nil {
		fmt.Printf("Failed to get dashboard stats: %v\n", err)
		return
	}

	fmt.Printf("Dashboard stats: %s\n", string(data))
}

// Perform realtime detection
func performRealtimeDetection() bool {
	// Собираем данные о трафике для анализа
	var tcpStats MIB_TCPSTATS
	ret, _, _ := procGetTcpStatistics.Call(uintptr(unsafe.Pointer(&tcpStats)))

	if ret != 0 {
		fmt.Printf("GetTcpStatistics failed with error: %d\n", ret)
		return false
	}

	currentTime := time.Now().UTC().Format(time.RFC3339)

	// Создаем несколько записей о трафике
	var trafficData []map[string]interface{}
	for i := 0; i < 3; i++ {
		trafficData = append(trafficData, map[string]interface{}{
			"host_id":      hostID,
			"timestamp":    currentTime,
			"src_ip":       ipAddress,
			"dst_ip":       fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"src_port":     rand.Intn(60000) + 1024,
			"dst_port":     rand.Intn(1000) + 1,
			"protocol":     "TCP",
			"bytes_sent":   rand.Intn(10000),
			"bytes_recv":   rand.Intn(10000),
			"packets_sent": rand.Intn(100),
			"packets_recv": rand.Intn(100),
			"duration":     float64(rand.Intn(1000)) / 100.0,
		})
	}

	detectionRequest := map[string]interface{}{
		"traffic_data":  trafficData,
		"analysis_type": "network",
		"time_window":   "3600s", // 1 час
		"timestamp":     currentTime,
	}

	return sendDataToAPI("/api/v1/detection/realtime", detectionRequest)
}

// Search artifacts
func searchArtifacts(query string) {
	endpoint := fmt.Sprintf("/api/v1/artifacts/search?q=%s", query)
	data, err := getDataFromAPI(endpoint)
	if err != nil {
		fmt.Printf("Failed to search artifacts: %v\n", err)
		return
	}

	fmt.Printf("Search results for '%s': %s\n", query, string(data))
}

func main() {
	fmt.Printf("Windows SIEM Agent v%s starting...\n", AGENT_VERSION)

	if !initializeHostInfo() {
		fmt.Println("Failed to initialize host info")
		return
	}

	fmt.Printf("Host identified as: %s (%s)\n", hostname, ipAddress)
	fmt.Printf("Host ID: %s\n", hostID)

	fmt.Println("Registering host with SIEM...")
	if !registerHost() {
		fmt.Println("Failed to register host")
	} else {
		fmt.Println("Host registered successfully")
	}

	fmt.Println("Collecting network traffic data...")
	if !collectNetworkTraffic() {
		fmt.Println("Failed to collect network traffic data")
	} else {
		fmt.Println("Network traffic data sent successfully")
	}

	fmt.Println("Collecting registry artifacts...")
	if !collectRegistryArtifacts() {
		fmt.Println("Failed to collect registry artifacts")
	} else {
		fmt.Println("Registry artifacts sent successfully")
	}

	fmt.Println("Collecting process artifacts...")
	if !collectProcessArtifacts() {
		fmt.Println("Failed to collect process artifacts")
	} else {
		fmt.Println("Process artifacts sent successfully")
	}

	fmt.Println("Collecting network logs...")
	if !collectNetworkLogs() {
		fmt.Println("Failed to collect network logs")
	} else {
		fmt.Println("Network logs sent successfully")
	}

	fmt.Println("Collecting system logs...")
	if !collectSystemLogs() {
		fmt.Println("Failed to collect system logs")
	} else {
		fmt.Println("System logs sent successfully")
	}

	fmt.Println("Performing realtime detection...")
	if !performRealtimeDetection() {
		fmt.Println("Failed to perform realtime detection")
	} else {
		fmt.Println("Realtime detection completed successfully")
	}

	fmt.Println("Getting threat patterns...")
	getThreatPatterns()

	fmt.Println("Getting traffic statistics...")
	getTrafficStats()

	fmt.Println("Getting dashboard stats...")
	getDashboardStats()

	fmt.Println("Searching for suspicious artifacts...")
	searchArtifacts("suspicious")

	// Проверяем наличие файла для сканирования
	exePath, _ := os.Executable()
	if _, err := os.Stat(exePath); err == nil {
		fmt.Println("Scanning executable with VirusTotal...")
		if !scanFileWithVirusTotal(exePath) {
			fmt.Println("Failed to scan file with VirusTotal")
		} else {
			fmt.Println("File scanned successfully")
		}
	}

	fmt.Println("Windows SIEM Agent finished")
}
