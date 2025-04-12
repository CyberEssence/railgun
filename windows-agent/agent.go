package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	BUFFER_SIZE    = 4096
	API_HOST       = "http://localhost:8080" // Change to your server address
	AGENT_VERSION  = "1.0.0"
	MAX_COMP_NAME  = 15
	computernameEx = 5 // ComputerNamePhysicalDnsFullyQualified
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

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("API request failed. Status: %d, Response: %s\n", resp.StatusCode, string(body))
	return false
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
		"id":         hostID,
		"hostname":   hostname,
		"ip_address": ipAddress,
		"os_version": "Windows 10",
		"status":     "active",
		"last_seen":  currentTime,
	}

	return sendDataToAPI("/api/hosts", hostData)
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

	return sendDataToAPI("/api/traffic", trafficData)
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

	return sendDataToAPI("/api/artifacts", artifactData)
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

	fmt.Println("Windows SIEM Agent finished")
}
