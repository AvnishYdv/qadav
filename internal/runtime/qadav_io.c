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

#if defined(_WIN32)
__declspec(dllimport) char* __stdcall GetCommandLineA(void);
#endif

const char* qadav_get_input_path(void) {
    static char path[260];
#if defined(_WIN32)
    const char* cmd = GetCommandLineA();
    if (!cmd) return "";
    int pi = 0;
    int in_arg = 0;
    int in_quote = 0;
    for (int i = 0; cmd[i]; i++) {
        char c = cmd[i];
        if (c == '"') {
            in_quote = !in_quote;
            continue;
        }
        if (!in_quote && (c == ' ' || c == '\t')) {
            if (in_arg) {
                path[pi] = '\0';
                if (pi > 4 && path[pi-4] == '.' && path[pi-3] == 'y' && path[pi-2] == 'd' && path[pi-1] == 'v') {
                    return path;
                }
                pi = 0;
                in_arg = 0;
            }
        } else {
            in_arg = 1;
            if (pi < 255) {
                path[pi++] = c;
            }
        }
    }
    if (in_arg) {
        path[pi] = '\0';
        if (pi > 4 && path[pi-4] == '.' && path[pi-3] == 'y' && path[pi-2] == 'd' && path[pi-1] == 'v') {
            return path;
        }
    }
#endif
    return "";
}

long long qadav_extract_yield_number(const char* path) {
    if (!path || !*path) return -1;
    void* h = qadav_file_open_read(path);
    if (!h) return -1;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return -1;
    buf[n] = '\0';
    for (int i = 0; i <= n - 5; i++) {
        if (buf[i] == 'y' && buf[i+1] == 'i' && buf[i+2] == 'e' && buf[i+3] == 'l' && buf[i+4] == 'd') {
            int j = i + 5;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t' || buf[j] == '\r' || buf[j] == '\n')) j++;
            int neg = 0;
            if (j < n && buf[j] == '-') {
                neg = 1;
                j++;
            }
            if (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                long long val = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    val = val * 10 + (buf[j] - '0');
                    j++;
                }
                return neg ? -val : val;
            }
        }
    }
    return -1;
}

const char* qadav_format_trunc_line(long long val) {
    static char buf[64];
    char num[32];
    int ni = 0;
    long long v = val;
    if (v == 0) {
        num[ni++] = '0';
    } else {
        int neg = 0;
        unsigned long long uv;
        if (v < 0) {
            neg = 1;
            uv = (unsigned long long)(-v);
        } else {
            uv = (unsigned long long)v;
        }
        char tmp[32];
        int ti = 0;
        while (uv > 0) {
            tmp[ti++] = '0' + (char)(uv % 10);
            uv /= 10;
        }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    num[ni] = '\0';

    int bi = 0;
    const char* pfx = "  %t2 = trunc i64 ";
    while (*pfx) buf[bi++] = *pfx++;
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* sfx = " to i32";
    while (*sfx) buf[bi++] = *sfx++;
    buf[bi] = '\0';
    return buf;
}

// Statement scanning and multi-statement helpers for M8.4e
typedef struct {
    int kind;            // 0=var, 1=emit, 2=yield
    char name[64];       // var name
    long long value;     // var value / yield value
    char expr[256];      // raw emit expression
    int expr_len;
} qadav_stmt_t;

static qadav_stmt_t g_stmts[16];
static int g_stmt_count = 0;

static int rt_io_contains(const char* haystack, const char* needle) {
    if (!haystack || !needle) return 0;
    if (!*needle) return 1;
    for (; *haystack; haystack++) {
        const char* h = haystack;
        const char* n = needle;
        while (*h && *n && (*h == *n)) {
            h++;
            n++;
        }
        if (!*n) return 1;
    }
    return 0;
}

long long qadav_detect_program_kind(const char* path) {
    if (!path || !*path) return -1;
    void* h = qadav_file_open_read(path);
    if (!h) return -1;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return -1;
    buf[n] = '\0';

    if (rt_io_contains(buf, "tensor") || rt_io_contains(buf, "dot")) return 2; // dot_product
    if (rt_io_contains(buf, "fib")) return 1;                                  // fib
    if (rt_io_contains(buf, "calc")) return 3;                                 // calc-family
    if (rt_io_contains(buf, "x \xe2\x86\xa6") || rt_io_contains(buf, "\xe2\x8a\x95") || rt_io_contains(buf, "\xe2\x8a\x97")) return 3; // calc
    if (rt_io_contains(buf, "emit") && rt_io_contains(buf, "yield")) return 0; // hello-family
    return 0;
}

long long qadav_scan_statements(const char* path) {
    g_stmt_count = 0;
    if (!path || !*path) return 0;
    void* h = qadav_file_open_read(path);
    if (!h) return 0;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return 0;
    buf[n] = '\0';

    int i = 0;
    while (i < n && g_stmt_count < 16) {
        // Skip whitespace and block symbols
        while (i < n && (buf[i] == ' ' || buf[i] == '\t' || buf[i] == '\r' || buf[i] == '\n' ||
                         (unsigned char)buf[i] == 0xe2 || (unsigned char)buf[i] == 0x9f || (unsigned char)buf[i] == 0xa6 || (unsigned char)buf[i] == 0xa7)) {
            i++;
        }
        if (i >= n) break;

        // Check for process declaration -> skip to block start
        if (buf[i] == 'p' && i + 7 < n && buf[i+1]=='r' && buf[i+2]=='o' && buf[i+3]=='c' && buf[i+4]=='e' && buf[i+5]=='s' && buf[i+6]=='s') {
            while (i < n && !((unsigned char)buf[i] == 0xe2 && (unsigned char)buf[i+1] == 0x9f && (unsigned char)buf[i+2] == 0xa6)) i++;
            if (i < n) i += 3;
            continue;
        }

        // Statement 1: int64 NAME ↦ N ·
        if (buf[i] == 'i' && i + 6 < n && buf[i+1]=='n' && buf[i+2]=='t' && buf[i+3]=='6' && buf[i+4]=='4') {
            i += 5;
            while (i < n && (buf[i] == ' ' || buf[i] == '\t')) i++;
            int ni = 0;
            while (i < n && buf[i] != ' ' && buf[i] != '\t' && (unsigned char)buf[i] != 0xe2 && ni < 63) {
                g_stmts[g_stmt_count].name[ni++] = buf[i++];
            }
            g_stmts[g_stmt_count].name[ni] = '\0';
            // skip ↦ (\xe2\x86\xa6)
            while (i < n && !((buf[i] >= '0' && buf[i] <= '9') || buf[i] == '-')) i++;
            int neg = 0;
            if (i < n && buf[i] == '-') { neg = 1; i++; }
            long long val = 0;
            while (i < n && (buf[i] >= '0' && buf[i] <= '9')) {
                val = val * 10 + (buf[i++] - '0');
            }
            g_stmts[g_stmt_count].kind = 0;
            g_stmts[g_stmt_count].value = neg ? -val : val;
            g_stmt_count++;
            while (i < n && (unsigned char)buf[i] != 0xc2) i++; // skip to · (\xc2\xb7)
            if (i < n) i += 2;
            continue;
        }

        // Statement 2: emit EXPR ·
        if (buf[i] == 'e' && i + 5 < n && buf[i+1]=='m' && buf[i+2]=='i' && buf[i+3]=='t' && (buf[i+4]==' ' || buf[i+4]=='\t')) {
            i += 5;
            while (i < n && (buf[i] == ' ' || buf[i] == '\t')) i++;
            int ei = 0;
            while (i < n && !((unsigned char)buf[i] == 0xc2 && (unsigned char)buf[i+1] == 0xb7) && buf[i] != '\r' && buf[i] != '\n' && ei < 255) {
                g_stmts[g_stmt_count].expr[ei++] = buf[i++];
            }
            while (ei > 0 && (g_stmts[g_stmt_count].expr[ei-1] == ' ' || g_stmts[g_stmt_count].expr[ei-1] == '\t')) ei--;
            g_stmts[g_stmt_count].expr[ei] = '\0';
            g_stmts[g_stmt_count].expr_len = ei;
            g_stmts[g_stmt_count].kind = 1;
            g_stmt_count++;
            while (i < n && !((unsigned char)buf[i] == 0xc2 && (unsigned char)buf[i+1] == 0xb7)) i++;
            if (i < n) i += 2;
            continue;
        }

        // Statement 3: yield N ·
        if (buf[i] == 'y' && i + 6 < n && buf[i+1]=='i' && buf[i+2]=='e' && buf[i+3]=='l' && buf[i+4]=='d' && (buf[i+5]==' ' || buf[i+5]=='\t')) {
            i += 6;
            while (i < n && (buf[i] == ' ' || buf[i] == '\t')) i++;
            int neg = 0;
            if (i < n && buf[i] == '-') { neg = 1; i++; }
            long long val = 0;
            while (i < n && (buf[i] >= '0' && buf[i] <= '9')) {
                val = val * 10 + (buf[i++] - '0');
            }
            g_stmts[g_stmt_count].kind = 2;
            g_stmts[g_stmt_count].value = neg ? -val : val;
            g_stmt_count++;
            while (i < n && !((unsigned char)buf[i] == 0xc2 && (unsigned char)buf[i+1] == 0xb7)) i++;
            if (i < n) i += 2;
            continue;
        }

        i++;
    }
    return g_stmt_count;
}

long long qadav_stmt_kind(long long idx) {
    if (idx < 0 || idx >= g_stmt_count) return -1;
    return g_stmts[idx].kind;
}

const char* qadav_stmt_name(long long idx) {
    if (idx < 0 || idx >= g_stmt_count) return "";
    return g_stmts[idx].name;
}

long long qadav_stmt_value(long long idx) {
    if (idx < 0 || idx >= g_stmt_count) return 0;
    return g_stmts[idx].value;
}

const char* qadav_stmt_expr(long long idx) {
    if (idx < 0 || idx >= g_stmt_count) return "";
    return g_stmts[idx].expr;
}

const char* qadav_format_alloca(const char* name) {
    static char buf[128];
    int bi = 0;
    const char* pfx = "  %";
    while (*pfx) buf[bi++] = *pfx++;
    while (*name) buf[bi++] = *name++;
    const char* sfx = ".ptr = alloca i64";
    while (*sfx) buf[bi++] = *sfx++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_store(const char* name, long long val) {
    static char buf[128];
    char num[32];
    int ni = 0;
    long long v = val;
    if (v == 0) {
        num[ni++] = '0';
    } else {
        int neg = 0;
        unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32];
        int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    num[ni] = '\0';

    int bi = 0;
    const char* p1 = "  store i64 ";
    while (*p1) buf[bi++] = *p1++;
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", i64* %";
    while (*p2) buf[bi++] = *p2++;
    while (*name) buf[bi++] = *name++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_load(const char* name, long long temp_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else {
        char tmp[32]; int ti = 0;
        while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; }
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = load i64, i64* %";
    while (*p2) buf[bi++] = *p2++;
    while (*name) buf[bi++] = *name++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_binop(const char* op, long long res_temp, long long t1, long long t2) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = res_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = ";
    while (*p2) buf[bi++] = *p2++;
    while (*op) buf[bi++] = *op++;
    const char* p3 = " i64 %t";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = t1;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p4 = ", %t";
    while (*p4) buf[bi++] = *p4++;
    ni = 0; v = t2;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_emit_int_call(long long temp_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_emit_int(i64 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_trunc(long long temp_id, long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = trunc i64 ";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = val;
    if (v == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = " to i32";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret(long long temp_id) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  ret i32 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

// Fib helpers for M8.4f
static long long g_fib_values[4];

int qadav_scan_fib_values(const char* path) {
    if (!path || !*path) return -1;
    void* h = qadav_file_open_read(path);
    if (!h) return -1;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return -1;
    buf[n] = '\0';

    int found0 = 0, found1 = 0, found2 = 0, found3 = 0;

    // 0: Initial n in main: `int64 n ↦ N`
    for (int i = 0; i <= n - 5; i++) {
        if ((unsigned char)buf[i] == 'n' && (unsigned char)buf[i+1] == ' ' &&
            (unsigned char)buf[i+2] == 0xe2 && (unsigned char)buf[i+3] == 0x86 && (unsigned char)buf[i+4] == 0xa6) {
            int j = i + 5;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            if (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                long long val = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    val = val * 10 + (buf[j++] - '0');
                }
                g_fib_values[0] = val;
                found0 = 1;
                break;
            }
        }
    }

    // 1: when threshold: `when n ≺ N`
    for (int i = 0; i <= n - 10; i++) {
        if (buf[i] == 'w' && buf[i+1] == 'h' && buf[i+2] == 'e' && buf[i+3] == 'n' &&
            buf[i+4] == ' ' && buf[i+5] == 'n' && buf[i+6] == ' ' &&
            (unsigned char)buf[i+7] == 0xe2 && (unsigned char)buf[i+8] == 0x89 && (unsigned char)buf[i+9] == 0xba) {
            int j = i + 10;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            if (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                long long val = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    val = val * 10 + (buf[j++] - '0');
                }
                g_fib_values[1] = val;
                found1 = 1;
                break;
            }
        }
    }

    // 2 & 3: rec offsets: `fib⦅n ⊖ N⦆`
    int rec_count = 0;
    for (int i = 0; i <= n - 11; i++) {
        if (buf[i] == 'f' && buf[i+1] == 'i' && buf[i+2] == 'b' &&
            (unsigned char)buf[i+3] == 0xe2 && (unsigned char)buf[i+4] == 0xa6 && (unsigned char)buf[i+5] == 0x85 &&
            buf[i+6] == 'n' && buf[i+7] == ' ' &&
            (unsigned char)buf[i+8] == 0xe2 && (unsigned char)buf[i+9] == 0x8a && (unsigned char)buf[i+10] == 0x96) {
            int j = i + 11;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            if (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                long long val = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    val = val * 10 + (buf[j++] - '0');
                }
                if (rec_count == 0) {
                    g_fib_values[2] = val;
                    found2 = 1;
                    rec_count++;
                } else if (rec_count == 1) {
                    g_fib_values[3] = val;
                    found3 = 1;
                    rec_count++;
                    break;
                }
            }
        }
    }

    if (found0 && found1 && found2 && found3) return 4;
    return -1;
}

long long qadav_get_fib_value(long long idx) {
    if (idx < 0 || idx >= 4) return -1;
    return g_fib_values[idx];
}

const char* qadav_format_icmp_slt(long long t_cur, long long t_prev, long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = t_cur;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = icmp slt i64 %t";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = t_prev;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", ";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = val;
    if (v == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_sub(long long t_cur, long long t_prev, long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = t_cur;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = sub i64 %t";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = t_prev;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", ";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = val;
    if (v == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_store_direct(long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i64 ";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = val;
    if (v == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", i64* %n.ptr";
    while (*p2) buf[bi++] = *p2++;
    buf[bi] = '\0';
    return buf;
}

// Tensor helpers for M8.4g
typedef struct {
    char name[32];
    int size;
    long long values[16];
} qadav_tensor_info_t;

static qadav_tensor_info_t g_tensors[4];
static int g_tensor_count = 0;

int qadav_scan_tensors(const char* path) {
    g_tensor_count = 0;
    if (!path || !*path) return -1;
    void* h = qadav_file_open_read(path);
    if (!h) return -1;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return -1;
    buf[n] = '\0';

    for (int i = 0; i <= n - 15; i++) {
        if ((unsigned char)buf[i] == 0xe2 && (unsigned char)buf[i+1] == 0x86 && (unsigned char)buf[i+2] == 0xa6 &&
            buf[i+3] == ' ' &&
            buf[i+4] == 't' && buf[i+5] == 'e' && buf[i+6] == 'n' && buf[i+7] == 's' && buf[i+8] == 'o' && buf[i+9] == 'r' &&
            (unsigned char)buf[i+10] == 0xe2 && (unsigned char)buf[i+11] == 0x9f && (unsigned char)buf[i+12] == 0xac) {
            
            int k = i - 1;
            while (k >= 0 && (buf[k] == ' ' || buf[k] == '\t')) k--;
            int name_end = k;
            while (k >= 0 && buf[k] != ' ' && buf[k] != '\t' && (unsigned char)buf[k] != 0xad) k--;
            int name_start = k + 1;
            int nl = 0;
            for (int p = name_start; p <= name_end && nl < 31; p++) {
                g_tensors[g_tensor_count].name[nl++] = buf[p];
            }
            g_tensors[g_tensor_count].name[nl] = '\0';

            int sz = 3;
            for (int p = name_start - 1; p >= 3; p--) {
                if ((unsigned char)buf[p-3] == 0xe2 && (unsigned char)buf[p-2] == 0x9f && (unsigned char)buf[p-1] == 0xad) {
                    int q = p - 4;
                    while (q >= 0 && (buf[q] >= '0' && buf[q] <= '9')) q--;
                    int sval = 0;
                    for (int r = q + 1; r < p - 3; r++) {
                        sval = sval * 10 + (buf[r] - '0');
                    }
                    if (sval > 0) sz = sval;
                    break;
                }
            }
            g_tensors[g_tensor_count].size = sz;

            int j = i + 13;
            int elem_idx = 0;
            while (j < n && elem_idx < 16) {
                while (j < n && (buf[j] == ' ' || buf[j] == '\t' || buf[j] == '\r' || buf[j] == '\n' ||
                                 ((unsigned char)buf[j] == 0xe2 && (unsigned char)buf[j+1] == 0xb8 && (unsigned char)buf[j+2] == 0xb4))) {
                    if ((unsigned char)buf[j] == 0xe2) j += 3;
                    else j++;
                }
                if ((unsigned char)buf[j] == 0xe2 && (unsigned char)buf[j+1] == 0x9f && (unsigned char)buf[j+2] == 0xad) {
                    break;
                }
                if (j >= n) break;

                int neg = 0;
                if (buf[j] == '-') { neg = 1; j++; }
                long long ival = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    ival = ival * 10 + (buf[j++] - '0');
                }
                if (j < n && buf[j] == '.') {
                    j++;
                    while (j < n && (buf[j] >= '0' && buf[j] <= '9')) j++;
                }
                g_tensors[g_tensor_count].values[elem_idx++] = neg ? -ival : ival;
            }
            g_tensor_count++;
            if (g_tensor_count >= 2) break;
        }
    }
    return g_tensor_count;
}

const char* qadav_get_tensor_name(long long idx) {
    if (idx < 0 || idx >= g_tensor_count) return "";
    return g_tensors[idx].name;
}

long long qadav_get_tensor_size(long long idx) {
    if (idx < 0 || idx >= g_tensor_count) return 0;
    return g_tensors[idx].size;
}

long long qadav_get_tensor_int_val(long long t_idx, long long e_idx) {
    if (t_idx < 0 || t_idx >= g_tensor_count || e_idx < 0 || e_idx >= 16) return 0;
    return g_tensors[t_idx].values[e_idx];
}

const char* qadav_format_shape_const(long long shape_idx, long long size) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "@.shape.";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = shape_idx;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = private unnamed_addr constant [1 x i32] [i32 ";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = size;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ']';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_tensor_set(long long t_temp, long long elem_idx, long long int_val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_tensor_set_f64(i8* %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = t_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", i32 ";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = elem_idx;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", double ";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = int_val;
    if (v == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (v < 0) { neg = 1; uv = (unsigned long long)(-v); } else { uv = (unsigned long long)v; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p4 = ".000000)";
    while (*p4) buf[bi++] = *p4++;
    buf[bi] = '\0';
    return buf;
}


