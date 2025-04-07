#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
#include <windows.h>
#include <ws2tcpip.h>
#include <iphlpapi.h>
#include <winreg.h>
#include <psapi.h>
#include <tlhelp32.h>
#include <curl/curl.h>
#include <jansson.h>

#pragma comment(lib, "iphlpapi.lib")
#pragma comment(lib, "ws2_32.lib")
#pragma comment(lib, "libcurl.lib")

#define BUFFER_SIZE 4096
#define API_HOST "http://localhost:8080"  // Измените на ваш адрес сервера
#define AGENT_VERSION "1.0.0"

// Глобальные переменные
char host_id[64] = {0};
char hostname[MAX_COMPUTERNAME_LENGTH + 1] = {0};
char ip_address[46] = {0};
DWORD name_size = sizeof(hostname);

// Функция для отправки данных на сервер
size_t write_callback(void *contents, size_t size, size_t nmemb, void *userp) {
    return size * nmemb;
}

int send_data_to_api(const char *endpoint, const char *json_data) {
    CURL *curl;
    CURLcode res;
    char url[512] = {0};
    
    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    headers = curl_slist_append(headers, "Accept: application/json");
    
    snprintf(url, sizeof(url), "%s%s", API_HOST, endpoint);
    
    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 5L);
        
        // Для отладки (можно убрать в продакшене)
        curl_easy_setopt(curl, CURLOPT_VERBOSE, 1L);
        
        res = curl_easy_perform(curl);
        
        curl_easy_cleanup(curl);
        curl_slist_free_all(headers);
        
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
            return 0;
        }
        
        return 1;
    }
    
    return 0;
}

int initialize_host_info() {
    // Получение имени компьютера
    if (!GetComputerNameA(hostname, &name_size)) {
        fprintf(stderr, "Failed to get hostname. Error: %lu\n", GetLastError());
        return 0;
    }
    
    // Получение IP-адреса
    DWORD dwRetVal = 0;
    PIP_ADAPTER_ADDRESSES pAddresses = NULL;
    ULONG outBufLen = 15000;
    
    pAddresses = (IP_ADAPTER_ADDRESSES *) malloc(outBufLen);
    if (pAddresses == NULL) {
        fprintf(stderr, "Memory allocation failed for IP_ADAPTER_ADDRESSES\n");
        return 0;
    }
    
    dwRetVal = GetAdaptersAddresses(AF_INET, 0, NULL, pAddresses, &outBufLen);
    if (dwRetVal == NO_ERROR) {
        PIP_ADAPTER_ADDRESSES pCurrAddresses = pAddresses;
        while (pCurrAddresses) {
            if (pCurrAddresses->OperStatus == IfOperStatusUp && pCurrAddresses->Ipv4Enabled) {
                PIP_ADAPTER_UNICAST_ADDRESS pUnicast = pCurrAddresses->FirstUnicastAddress;
                if (pUnicast != NULL) {
                    SOCKET_ADDRESS SocketAddress = pUnicast->Address;
                    SOCKADDR_IN *sockaddr_ipv4 = (SOCKADDR_IN *) SocketAddress.lpSockaddr;
                    inet_ntop(AF_INET, &sockaddr_ipv4->sin_addr, ip_address, sizeof(ip_address));
                    break;
                }
            }
            pCurrAddresses = pCurrAddresses->Next;
        }
    }
    
    free(pAddresses);
    
    // Генерация host_id
    sprintf(host_id, "%s-%08x", hostname, rand());
    
    return 1;
}

int register_host() {
    json_t *root;
    char *json_data;
    int result = 0;
    
    root = json_object();
    if (!root) {
        fprintf(stderr, "Failed to create JSON object\n");
        return 0;
    }
    
    json_object_set_new(root, "id", json_string(host_id));
    json_object_set_new(root, "hostname", json_string(hostname));
    json_object_set_new(root, "ip_address", json_string(ip_address));
    json_object_set_new(root, "os_version", json_string("Windows 10"));
    json_object_set_new(root, "status", json_string("active"));
    
    char timestamp[64];
    time_t now = time(NULL);
    strftime(timestamp, sizeof(timestamp), "%Y-%m-%dT%H:%M:%SZ", gmtime(&now));
    json_object_set_new(root, "last_seen", json_string(timestamp));
    
    json_data = json_dumps(root, JSON_COMPACT);
    if (!json_data) {
        fprintf(stderr, "Failed to serialize JSON\n");
        json_decref(root);
        return 0;
    }
    
    printf("Sending host data: %s\n", json_data);
    result = send_data_to_api("/api/hosts", json_data);
    
    json_decref(root);
    free(json_data);
    
    return result;
}

int collect_network_traffic() {
    MIB_TCPSTATS tcpStats;
    DWORD dwRetVal = GetTcpStatistics(&tcpStats);
    
    if (dwRetVal == NO_ERROR) {
        json_t *root = json_object();
        if (!root) {
            fprintf(stderr, "Failed to create JSON object\n");
            return 0;
        }
        
        char timestamp[64];
        time_t now = time(NULL);
        strftime(timestamp, sizeof(timestamp), "%Y-%m-%dT%H:%M:%SZ", gmtime(&now));
        
        json_object_set_new(root, "host_id", json_string(host_id));
        json_object_set_new(root, "timestamp", json_string(timestamp));
        json_object_set_new(root, "src_ip", json_string(ip_address));
        json_object_set_new(root, "dst_ip", json_string("8.8.8.8"));
        json_object_set_new(root, "src_port", json_integer(12345));
        json_object_set_new(root, "dst_port", json_integer(80));
        json_object_set_new(root, "protocol", json_string("TCP"));
        json_object_set_new(root, "bytes_sent", json_integer(tcpStats.dwOutSegs * 1460));
        json_object_set_new(root, "bytes_recv", json_integer(tcpStats.dwInSegs * 1460));
        json_object_set_new(root, "packets_sent", json_integer(tcpStats.dwOutSegs));
        json_object_set_new(root, "packets_recv", json_integer(tcpStats.dwInSegs));
        json_object_set_new(root, "duration", json_real(10.5));
        
        char *json_data = json_dumps(root, JSON_COMPACT);
        if (!json_data) {
            fprintf(stderr, "Failed to serialize JSON\n");
            json_decref(root);
            return 0;
        }
        
        printf("Sending traffic data: %s\n", json_data);
        int result = send_data_to_api("/api/traffic", json_data);
        
        json_decref(root);
        free(json_data);
        
        return result;
    } else {
        fprintf(stderr, "GetTcpStatistics failed with error: %d\n", dwRetVal);
        return 0;
    }
}

int collect_registry_artifacts() {
    HKEY hKey;
    DWORD dwType;
    char data[BUFFER_SIZE];
    DWORD dataSize = BUFFER_SIZE;
    
    LONG lResult = RegOpenKeyExA(
        HKEY_LOCAL_MACHINE,
        "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion",
        0,
        KEY_READ,
        &hKey);
    
    if (lResult != ERROR_SUCCESS) {
        fprintf(stderr, "Failed to open registry key. Error: %lu\n", GetLastError());
        return 0;
    }
    
    lResult = RegQueryValueExA(
        hKey,
        "ProductName",
        NULL,
        &dwType,
        (LPBYTE)data,
        &dataSize);
    
    RegCloseKey(hKey);
    
    if (lResult != ERROR_SUCCESS) {
        fprintf(stderr, "Failed to read registry value. Error: %lu\n", GetLastError());
        return 0;
    }
    
    json_t *root = json_object();
    if (!root) {
        fprintf(stderr, "Failed to create JSON object\n");
        return 0;
    }
    
    json_t *metadata = json_object();
    if (!metadata) {
        fprintf(stderr, "Failed to create metadata JSON object\n");
        json_decref(root);
        return 0;
    }
    
    char timestamp[64];
    time_t now = time(NULL);
    strftime(timestamp, sizeof(timestamp), "%Y-%m-%dT%H:%M:%SZ", gmtime(&now));
    
    json_object_set_new(root, "host_id", json_string(host_id));
    json_object_set_new(root, "timestamp", json_string(timestamp));
    json_object_set_new(root, "type", json_string("registry"));
    json_object_set_new(root, "path", json_string("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\ProductName"));
    json_object_set_new(root, "value", json_string(data));
    json_object_set_new(root, "size", json_integer(dataSize));
    json_object_set_new(root, "owner", json_string("SYSTEM"));
    
    json_object_set_new(metadata, "registry_type", json_string("REG_SZ"));
    json_object_set_new(metadata, "last_modified", json_string(timestamp));
    
    json_object_set_new(root, "metadata", metadata);
    
    char *json_data = json_dumps(root, JSON_COMPACT);
    if (!json_data) {
        fprintf(stderr, "Failed to serialize JSON\n");
        json_decref(root);
        return 0;
    }
    
    printf("Sending artifact data: %s\n", json_data);
    int result = send_data_to_api("/api/artifacts", json_data);
    
    json_decref(root);
    free(json_data);
    
    return result;
}

// Остальные функции (collect_process_artifacts, collect_file_artifacts) остаются аналогичными,
// но с добавлением проверок ошибок JSON и правильными эндпоинтами

int main(int argc, char *argv[]) {
    srand((unsigned int)time(NULL));
    curl_global_init(CURL_GLOBAL_ALL);
    
    printf("Windows SIEM Agent v%s starting...\n", AGENT_VERSION);
    
    if (!initialize_host_info()) {
        fprintf(stderr, "Failed to initialize host info\n");
        return 1;
    }
    
    printf("Host identified as: %s (%s)\n", hostname, ip_address);
    printf("Host ID: %s\n", host_id);
    
    printf("Registering host with SIEM...\n");
    if (!register_host()) {
        fprintf(stderr, "Failed to register host\n");
    } else {
        printf("Host registered successfully\n");
    }
    
    printf("Collecting network traffic data...\n");
    if (!collect_network_traffic()) {
        fprintf(stderr, "Failed to collect network traffic data\n");
    } else {
        printf("Network traffic data sent successfully\n");
    }
    
    printf("Collecting registry artifacts...\n");
    if (!collect_registry_artifacts()) {
        fprintf(stderr, "Failed to collect registry artifacts\n");
    } else {
        printf("Registry artifacts sent successfully\n");
    }
    
    curl_global_cleanup();
    printf("Windows SIEM Agent finished\n");
    return 0;
}
