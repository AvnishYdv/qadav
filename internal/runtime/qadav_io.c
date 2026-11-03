#include "qadav_io.h"

static int qadav_path_is_safe(const char* path) {
    if (!path) return 0;
    for (int i = 0; path[i] != '\0'; i++) {
        if (path[i] == '.' && path[i+1] == '.') {
            char after = path[i+2];
            if (after == '/' || after == '\\' || after == '\0') {
                return 0; /* Path traversal rejected */
            }
        }
    }
    return 1;
}

#ifdef _WIN32

#define NULL ((void*)0)
#define INVALID_HANDLE_VALUE ((void*)(long long)-1)
#define GENERIC_READ (0x80000000L)
#define GENERIC_WRITE (0x40000000L)
#define FILE_SHARE_READ 0x00000001
#define CREATE_ALWAYS 2
#define OPEN_EXISTING 3
#define FILE_ATTRIBUTE_NORMAL 0x00000080

typedef unsigned long DWORD;
typedef void* HANDLE;
typedef int BOOL;

__declspec(dllimport) HANDLE __stdcall CreateFileA(const char* lpFileName, DWORD dwDesiredAccess, DWORD dwShareMode, void* lpSecurityAttributes, DWORD dwCreationDisposition, DWORD dwFlagsAndAttributes, HANDLE hTemplateFile);
__declspec(dllimport) BOOL __stdcall ReadFile(HANDLE hFile, void* lpBuffer, DWORD nNumberOfBytesToRead, DWORD* lpNumberOfBytesRead, void* lpOverlapped);
__declspec(dllimport) BOOL __stdcall WriteFile(HANDLE hFile, const void* lpBuffer, DWORD nNumberOfBytesToWrite, DWORD* lpNumberOfBytesWritten, void* lpOverlapped);
__declspec(dllimport) BOOL __stdcall CloseHandle(HANDLE hObject);

void* qadav_file_open_read(const char* path) {
    if (!qadav_path_is_safe(path)) return NULL;
    HANDLE h = CreateFileA(path, GENERIC_READ, FILE_SHARE_READ, NULL,
                           OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    return (h == INVALID_HANDLE_VALUE) ? NULL : (void*)h;
}

void* qadav_file_open_write(const char* path) {
    if (!qadav_path_is_safe(path)) return NULL;
    HANDLE h = CreateFileA(path, GENERIC_WRITE, 0, NULL,
                           CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    return (h == INVALID_HANDLE_VALUE) ? NULL : (void*)h;
}

int qadav_file_read(void* h, void* buf, unsigned long long len) {
    if (!h) return -1;
    DWORD bytesRead = 0;
    BOOL ok = ReadFile((HANDLE)h, buf, (DWORD)len, &bytesRead, NULL);
    return ok ? (int)bytesRead : -1;
}

int qadav_file_write(void* h, const void* buf, unsigned long long len) {
    if (!h) return -1;
    DWORD bytesWritten = 0;
    BOOL ok = WriteFile((HANDLE)h, buf, (DWORD)len, &bytesWritten, NULL);
    return ok ? (int)bytesWritten : -1;
}

void qadav_file_close(void* h) {
    if (h && h != INVALID_HANDLE_VALUE) {
        CloseHandle((HANDLE)h);
    }
}
#else
void* qadav_file_open_read(const char* path) {
    if (!qadav_path_is_safe(path)) return (void*)-1;
    long fd;
    __asm__ volatile (
        "syscall"
        : "=a"(fd)
        : "a"(2), "D"(path), "S"(0), "d"(0)
        : "rcx", "r11", "memory"
    );
    return (void*)fd;
}

void* qadav_file_open_write(const char* path) {
    if (!qadav_path_is_safe(path)) return (void*)-1;
    long fd;
    __asm__ volatile (
        "syscall"
        : "=a"(fd)
        : "a"(2), "D"(path), "S"(577), "d"(0644)
        : "rcx", "r11", "memory"
    );
    return (void*)fd;
}

int qadav_file_read(void* h, void* buf, unsigned long long len) {
    long fd = (long)h;
    if (fd < 0) return -1;
    long ret;
    __asm__ volatile (
        "syscall"
        : "=a"(ret)
        : "a"(0), "D"(fd), "S"(buf), "d"(len)
        : "rcx", "r11", "memory"
    );
    return (int)ret;
}

int qadav_file_write(void* h, const void* buf, unsigned long long len) {
    long fd = (long)h;
    if (fd < 0) return -1;
    long ret;
    __asm__ volatile (
        "syscall"
        : "=a"(ret)
        : "a"(1), "D"(fd), "S"(buf), "d"(len)
        : "rcx", "r11", "memory"
    );
    return (int)ret;
}

void qadav_file_close(void* h) {
    long fd = (long)h;
    if (fd >= 0) {
        long ret;
        __asm__ volatile (
            "syscall"
            : "=a"(ret)
            : "a"(3), "D"(fd)
            : "rcx", "r11", "memory"
        );
    }
}
#endif
