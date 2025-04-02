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
#define API_HOST "http://backend:8080/api"
#define AGENT_VERSION "1.0.0"

// Глобальные переменные
char host_id[64] = {0};
char hostname[MAX_COMPUTERNAME_LENGTH + 1] = {0};
char ip_address[46] = {0};  // IPv6 может быть длиннее IPv4
DWORD name_size = sizeof(hostname);

// Функция для отправки данных на сервер
size_t write_callback(void *contents, size_t size, size_t nmemb, void *userp) {
    return size * nmemb;  // Игнорируем ответ
}

int send_data_to_api(const char *endpoint, const char *json_data) {
    CURL *curl;
    CURLcode res;
    char url[512] = {0};
    
    struct curl_slist *headers = NULL;
    headers = curl_slist_append(headers, "Content-Type: application/json");
    
    snprintf(url, sizeof(url), "%s/%s", API_HOST, endpoint);
    
    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url);
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, json_data);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 5L);
        
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

// Получение информации о хосте
int initialize_host_info() {
    // Получение имени компьютера
    if (!GetComputerNameA(hostname, &name_size)) {
        fprintf(stderr, "Failed to get hostname. Error: %lu\n", GetLastError());
        return 0;
    }
    
    // Получение IP-адреса
    DWORD dwRetVal = 0;
    PIP_ADAPTER_ADDRESSES pAddresses = NULL;
    ULONG outBufLen = 15000;  // Рекомендуемый размер
    
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
    
    // Генерация host_id (в реальной системе должен быть постоянным)
    sprintf(host_id, "%s-%08x", hostname, rand());
    
    return 1;
}

// Регистрация хоста в SIEM
int register_host() {
    json_t *root;
    char *json_data;
    
    root = json_object();
    
    json_object_set_new(root, "id", json_string(host_id));
    json_object_set_new(root, "hostname", json_string(hostname));
    json_object_set_new(root, "ip_address", json_string(ip_address));
    json_object_set_new(root, "os_version", json_string("Windows 10"));
    json_object_set_new(root, "status", json_string("active"));
    json_object_set_new(root, "last_seen", json_string("2023-10-30T10:00:00Z"));
    
    json_data = json_dumps(root, JSON_COMPACT);
    
    int result = send_data_to_api("hosts", json_data);
    
    json_decref(root);
    free(json_data);
    
    return result;
}

// Мониторинг сетевого трафика
int collect_network_traffic() {
    // Это упрощенная версия - в реальности нужно использовать WinPCap/Npcap или ETW
    // для действительного мониторинга сетевого трафика
    
    MIB_TCPSTATS tcpStats;
    DWORD dwRetVal = GetTcpStatistics(&tcpStats);
    
    if (dwRetVal == NO_ERROR) {
        json_t *root = json_object();
        
        json_object_set_new(root, "host_id", json_string(host_id));
        json_object_set_new(root, "timestamp", json_string("2023-10-30T10:00:00Z"));
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
        
        int result = send_data_to_api("traffic", json_data);
        
        json_decref(root);
        free(json_data);
        
        return result;
    } else {
        fprintf(stderr, "GetTcpStatistics failed with error: %d\n", dwRetVal);
        return 0;
    }
}

// Сбор артефактов из реестра
int collect_registry_artifacts() {
    HKEY hKey;
    DWORD dwType;
    char data[BUFFER_SIZE];
    DWORD dataSize = BUFFER_SIZE;
    
    // Открытие ключа реестра для примера
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
    
    // Чтение значения ключа
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
    
    // Отправка данных
    json_t *root = json_object();
    json_t *metadata = json_object();
    
    json_object_set_new(root, "host_id", json_string(host_id));
    json_object_set_new(root, "timestamp", json_string("2023-10-30T10:00:00Z"));
    json_object_set_new(root, "type", json_string("registry"));
    json_object_set_new(root, "path", json_string("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\ProductName"));
    json_object_set_new(root, "value", json_string(data));
    json_object_set_new(root, "size", json_integer(dataSize));
    json_object_set_new(root, "owner", json_string("SYSTEM"));
    
    json_object_set_new(metadata, "registry_type", json_string("REG_SZ"));
    json_object_set_new(metadata, "last_modified", json_string("2023-10-30T09:00:00Z"));
    
    json_object_set_new(root, "metadata", metadata);
    
    char *json_data = json_dumps(root, JSON_COMPACT);
    
    int result = send_data_to_api("artifacts", json_data);
    
    json_decref(root);
    free(json_data);
    
    return result;
}

// Сбор артефактов процессов
int collect_process_artifacts() {
    HANDLE hProcessSnap;
    PROCESSENTRY32 pe32;
    
    hProcessSnap = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0);
    if (hProcessSnap == INVALID_HANDLE_VALUE) {
        fprintf(stderr, "CreateToolhelp32Snapshot failed. Error: %lu\n", GetLastError());
        return 0;
    }
    
    pe32.dwSize = sizeof(PROCESSENTRY32);
    
    if (!Process32First(hProcessSnap, &pe32)) {
        CloseHandle(hProcessSnap);
        fprintf(stderr, "Process32First failed. Error: %lu\n", GetLastError());
        return 0;
    }
    
    // Для примера берем только первый процесс
    json_t *root = json_object();
    json_t *metadata = json_object();
    
    char process_path[MAX_PATH] = {0};
    HANDLE hProcess = OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ, FALSE, pe32.th32ProcessID);
    if (hProcess) {
        if (GetModuleFileNameExA(hProcess, NULL, process_path, MAX_PATH) == 0) {
            strcpy(process_path, "Unknown");
        }
        CloseHandle(hProcess);
    } else {
        strcpy(process_path, "Access Denied");
    }
    
    json_object_set_new(root, "host_id", json_string(host_id));
    json_object_set_new(root, "timestamp", json_string("2023-10-30T10:00:00Z"));
    json_object_set_new(root, "type", json_string("process"));
    json_object_set_new(root, "path", json_string(process_path));
    json_object_set_new(root, "value", json_string(pe32.szExeFile));
    json_object_set_new(root, "size", json_integer(0));
    json_object_set_new(root, "owner", json_string("SYSTEM"));
    
    json_object_set_new(metadata, "pid", json_integer(pe32.th32ProcessID));
    json_object_set_new(metadata, "ppid", json_integer(pe32.th32ParentProcessID));
    json_object_set_new(metadata, "thread_count", json_integer(pe32.cntThreads));
    
    json_object_set_new(root, "metadata", metadata);
    
    char *json_data = json_dumps(root, JSON_COMPACT);
    
    int result = send_data_to_api("artifacts", json_data);
    
    json_decref(root);
    free(json_data);
    CloseHandle(hProcessSnap);
    
    return result;
}

// Сбор артефактов файловой системы
int collect_file_artifacts() {
    WIN32_FIND_DATAA findData;
    HANDLE hFind;
    char windowsDir[MAX_PATH];
    char searchPath[MAX_PATH];
    
    if (!GetWindowsDirectoryA(windowsDir, MAX_PATH)) {
        fprintf(stderr, "GetWindowsDirectory failed. Error: %lu\n", GetLastError());
        return 0;
    }
    
    // Поиск файлов в папке Windows\System32 (для примера)
    sprintf(searchPath, "%s\\System32\\*.dll", windowsDir);
    
    hFind = FindFirstFileA(searchPath, &findData);
    if (hFind == INVALID_HANDLE_VALUE) {
        fprintf(stderr, "FindFirstFile failed. Error: %lu\n", GetLastError());
        return 0;
    }
    
    // Для примера берем только первый файл
    char filePath[MAX_PATH];
    sprintf(filePath, "%s\\System32\\%s", windowsDir, findData.cFileName);
    
    json_t *root = json_object();
    json_t *metadata = json_object();
    
    json_object_set_new(root, "host_id", json_string(host_id));
    json_object_set_new(root, "timestamp", json_string("2023-10-30T10:00:00Z"));
    json_object_set_new(root, "type", json_string("file"));
    json_object_set_new(root, "path", json_string(filePath));
    json_object_set_new(root, "value", json_string(""));
    
    // Получение размера файла
    LARGE_INTEGER fileSize;
    fileSize.LowPart = findData.nFileSizeLow;
    fileSize.HighPart = findData.nFileSizeHigh;
    
    json_object_set_new(root, "size", json_integer(fileSize.QuadPart));
    json_object_set_new(root, "owner", json_string("SYSTEM"));
    json_object_set_new(root, "permissions", json_string("rw-r--r--"));
    
    // Метаданные файла
    FILETIME ft = findData.ftLastWriteTime;
    SYSTEMTIME st;
    FileTimeToSystemTime(&ft, &st);
    
    char timeStr[64];
    sprintf(timeStr, "%04d-%02d-%02dT%02d:%02d:%02dZ", 
            st.wYear, st.wMonth, st.wDay, st.wHour, st.wMinute, st.wSecond);
    
    json_object_set_new(metadata, "created", json_string(timeStr));
    json_object_set_new(metadata, "modified", json_string(timeStr));
    json_object_set_new(metadata, "attributes", json_integer(findData.dwFileAttributes));
    
    json_object_set_new(root, "metadata", metadata);
    
    char *json_data = json_dumps(root, JSON_COMPACT);
    
    int result = send_data_to_api("artifacts", json_data);
    
    json_decref(root);
    free(json_data);
    FindClose(hFind);
    
    return result;
}

int main(int argc, char *argv[]) {
    // Инициализация системы
    srand((unsigned int)time(NULL));  // Инициализация генератора случайных чисел
    curl_global_init(CURL_GLOBAL_ALL);  // Инициализация libcurl
    
    printf("Windows SIEM Agent v%s starting...\n", AGENT_VERSION);
    
    // Получение информации о хосте
    if (!initialize_host_info()) {
        fprintf(stderr, "Failed to initialize host info\n");
        return 1;
    }
    
    printf("Host identified as: %s (%s)\n", hostname, ip_address);
    printf("Host ID: %s\n", host_id);
    
    // Регистрация хоста в SIEM
    printf("Registering host with SIEM...\n");
    if (!register_host()) {
        fprintf(stderr, "Failed to register host\n");
    } else {
        printf("Host registered successfully\n");
    }
    
    // Сбор и отправка данных
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
    
    printf("Collecting process artifacts...\n");
    if (!collect_process_artifacts()) {
        fprintf(stderr, "Failed to collect process artifacts\n");
    } else {
        printf("Process artifacts sent successfully\n");
    }
    
    printf("Collecting file artifacts...\n");
    if (!collect_file_artifacts()) {
        fprintf(stderr, "Failed to collect file artifacts\n");
    } else {
        printf("File artifacts sent successfully\n");
    }
    
    // В реальной системе здесь был бы бесконечный цикл сбора данных с задержкой
    
    // Очистка ресурсов
    curl_global_cleanup();
    
    printf("Windows SIEM Agent finished\n");
    return 0;
}