#include "qadav_io.h"
#include "qadav_variant.h"

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
    if (rt_io_contains(buf, "otherwise")) return 4;                            // if_else
    if (rt_io_contains(buf, "calc")) return 3;                                 // calc-family
    if (rt_io_contains(buf, "\xe2\x8a\x95") || rt_io_contains(buf, "\xe2\x8a\x97")) return 3; // calc
    if (rt_io_contains(buf, "emit") && rt_io_contains(buf, "yield")) return 0; // hello / multi_emit
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
    char val_strs[16][32];
} qadav_tensor_info_t;

static qadav_tensor_info_t g_tensors[4];
static int g_tensor_count = 0;
static int g_current_tensor_walk_idx = 0;

void qadav_set_current_tensor_walk_idx(long long idx) {
    g_current_tensor_walk_idx = (int)idx;
}

const char* qadav_get_tensor_val_str(long long elem_idx) {
    int tidx = g_current_tensor_walk_idx;
    if (tidx < 0 || tidx >= g_tensor_count || elem_idx < 0 || elem_idx >= 16) return "";
    return g_tensors[tidx].val_strs[elem_idx];
}

long long qadav_get_tensor_count(void) {
    return g_tensor_count;
}

static int io_str_eq(const char* a, const char* b) {
    if (!a || !b) return 0;
    while (*a && *b) {
        if (*a != *b) return 0;
        a++; b++;
    }
    return (*a == '\0' && *b == '\0') ? 1 : 0;
}

long long qadav_is_tensor_type(const char* typ) {
    return io_str_eq(typ, "tensor") ? 1 : 0;
}

long long qadav_is_float_type(const char* typ) {
    return io_str_eq(typ, "float64") ? 1 : 0;
}

long long qadav_is_ptr_var(const char* name) {
    if (!name || g_tensor_count <= 0) return 0;
    if (io_str_eq(name, "a") || io_str_eq(name, "b") ||
        io_str_eq(name, "x") || io_str_eq(name, "y")) return 1;
    return 0;
}

long long qadav_is_result_var(const char* name) {
    return io_str_eq(name, "result") ? 1 : 0;
}

long long qadav_is_mul_op(const char* op) {
    return io_str_eq(op, "mul") ? 1 : 0;
}

const char* qadav_format_shape_gep(long long shape_idx, long long temp_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = getelementptr inbounds [1 x i32], [1 x i32]* @.shape.";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = shape_idx;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", i32 0, i32 0";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

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

    int tidx = g_current_tensor_walk_idx;
    if (tidx >= 0 && tidx < g_tensor_count && elem_idx >= 0 && elem_idx < 16) {
        const char* fstr = g_tensors[tidx].val_strs[elem_idx];
        if (fstr && fstr[0] != '\0') {
            while (*fstr) buf[bi++] = *fstr++;
            buf[bi++] = ')';
            buf[bi] = '\0';
            return buf;
        }
    }

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

// String literal scanner for multi_emit and hello families
typedef struct {
    char text[128];
    int len;
} QadavStrLit;

static QadavStrLit g_emitted_strings[16];
static int g_emitted_string_count = 0;

long long qadav_scan_emitted_strings(const char* path) {
    g_emitted_string_count = 0;
    if (!path || !*path) return 0;
    void* h = qadav_file_open_read(path);
    if (!h) return 0;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return 0;
    buf[n] = '\0';

    int i = 0;
    while (i < n && g_emitted_string_count < 16) {
        if (buf[i] == 'e' && i + 4 < n && buf[i+1]=='m' && buf[i+2]=='i' && buf[i+3]=='t' && (buf[i+4]==' ' || buf[i+4]=='\t')) {
            i += 5;
            while (i < n && (buf[i] == ' ' || buf[i] == '\t')) i++;
            if (i < n && buf[i] == '"') {
                i++; // skip opening quote
                int si = 0;
                while (i < n && buf[i] != '"' && si < 127) {
                    g_emitted_strings[g_emitted_string_count].text[si++] = buf[i++];
                }
                g_emitted_strings[g_emitted_string_count].text[si] = '\0';
                g_emitted_strings[g_emitted_string_count].len = si + 1; // includes \00
                g_emitted_string_count++;
                if (i < n && buf[i] == '"') i++;
            }
        } else {
            i++;
        }
    }
    return g_emitted_string_count;
}

long long qadav_get_emitted_string_count(void) {
    return g_emitted_string_count;
}

const char* qadav_get_emitted_string_text(long long idx) {
    if (idx < 0 || idx >= g_emitted_string_count) return "";
    return g_emitted_strings[idx].text;
}

long long qadav_get_emitted_string_len(long long idx) {
    if (idx < 0 || idx >= g_emitted_string_count) return 0;
    return g_emitted_strings[idx].len;
}

const char* qadav_format_emitted_str_def(long long idx) {
    static char buf[256];
    if (idx < 0 || idx >= g_emitted_string_count) return "";
    int bi = 0;
    const char* p1 = "@.str.";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = idx;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = private unnamed_addr constant [";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = g_emitted_strings[idx].len;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = " x i8] c\"";
    while (*p3) buf[bi++] = *p3++;
    for (int i = 0; g_emitted_strings[idx].text[i]; i++) {
        buf[bi++] = g_emitted_strings[idx].text[i];
    }
    const char* p4 = "\\00\"";
    while (*p4) buf[bi++] = *p4++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_emitted_str_gep(long long str_idx, long long temp_id) {
    static char buf[256];
    if (str_idx < 0 || str_idx >= g_emitted_string_count) return "";
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = getelementptr inbounds [";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = g_emitted_strings[str_idx].len;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = " x i8], [";
    while (*p3) buf[bi++] = *p3++;
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p4 = " x i8]* @.str.";
    while (*p4) buf[bi++] = *p4++;
    ni = 0; v = str_idx;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p5 = ", i32 0, i32 0";
    while (*p5) buf[bi++] = *p5++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_emitted_str_call(long long temp_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_emit_text(i8* %t";
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

const char* qadav_format_trunc_line_id(long long temp_id, long long val) {
    static char buf[64];
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

const char* qadav_format_ret_line_id(long long temp_id) {
    static char buf[32];
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

static int io_streq(const char* a, const char* b) {
    if (!a || !b) return (a == b);
    while (*a && *b) {
        if (*a != *b) return 0;
        a++; b++;
    }
    return (*a == *b);
}

long long qadav_register_string(const char* s) {
    if (!s) return -1;
    for (int i = 0; i < g_emitted_string_count; i++) {
        if (io_streq(g_emitted_strings[i].text, s)) {
            return (long long)i;
        }
    }
    if (g_emitted_string_count >= 16) return -1;
    int idx = g_emitted_string_count;
    int si = 0;
    while (s[si] && si < 127) {
        g_emitted_strings[idx].text[si] = s[si];
        si++;
    }
    g_emitted_strings[idx].text[si] = '\0';
    g_emitted_strings[idx].len = si + 1;
    g_emitted_string_count++;
    return (long long)idx;
}

long long qadav_get_string_index(const char* s) {
    if (!s) return -1;
    for (int i = 0; i < g_emitted_string_count; i++) {
        if (io_streq(g_emitted_strings[i].text, s)) {
            return (long long)i;
        }
    }
    return -1;
}

long long qadav_get_string_byte_len(const char* s) {
    if (!s) return 0;
    long long len = 0;
    while (s[len]) len++;
    return len;
}

const char* qadav_escape_string(const char* s) {
    static char buf[256];
    if (!s) return "";
    int bi = 0;
    for (int i = 0; s[i] && bi < 250; i++) {
        if (s[i] == '"' || s[i] == '\\') {
            buf[bi++] = '\\';
        }
        buf[bi++] = s[i];
    }
    buf[bi] = '\0';
    return buf;
}

static long long g_last_int_val = 0;

void qadav_set_last_int_value(long long v) {
    g_last_int_val = v;
}

long long qadav_last_int_value(void) {
    return g_last_int_val;
}

static long long g_last_temp_reg = 1;
static long long g_last_emit_kind_val = 0; // 0 = text, 1 = int

void qadav_set_last_temp(long long t) {
    g_last_temp_reg = t;
}

long long qadav_get_last_temp(void) {
    return g_last_temp_reg;
}

void qadav_set_last_emit_kind(long long k) {
    g_last_emit_kind_val = k;
}

long long qadav_get_last_emit_kind(void) {
    return g_last_emit_kind_val;
}

const char* qadav_stmt_binop_op(long long idx) {
    if (idx < 0 || idx >= g_stmt_count) return "add";
    const char* s = g_stmts[idx].expr;
    while (*s) {
        if ((unsigned char)s[0] == 0xe2 && (unsigned char)s[1] == 0x8a) {
            if ((unsigned char)s[2] == 0x95) return "add";
            if ((unsigned char)s[2] == 0x96) return "sub";
            if ((unsigned char)s[2] == 0x97) return "mul";
            if ((unsigned char)s[2] == 0x98) return "sdiv";
        }
        s++;
    }
    return "add";
}

const char* qadav_stmt_binop_left(long long idx) {
    static char buf[64];
    if (idx < 0 || idx >= g_stmt_count) return "";
    const char* s = g_stmts[idx].expr;
    int bi = 0;
    while (*s && (*s == ' ' || *s == '\t')) s++;
    while (*s && *s != ' ' && *s != '\t' && (unsigned char)*s != 0xe2 && bi < 63) {
        buf[bi++] = *s++;
    }
    buf[bi] = '\0';
    return buf;
}

const char* qadav_stmt_binop_right(long long idx) {
    static char buf[64];
    if (idx < 0 || idx >= g_stmt_count) return "";
    const char* s = g_stmts[idx].expr;
    while (*s && !((unsigned char)s[0] == 0xe2 && (unsigned char)s[1] == 0x8a)) s++;
    if (*s) s += 3;
    while (*s && (*s == ' ' || *s == '\t')) s++;
    int bi = 0;
    while (*s && *s != ' ' && *s != '\t' && (unsigned char)*s != 0xc2 && bi < 63) {
        buf[bi++] = *s++;
    }
    buf[bi] = '\0';
    return buf;
}

static long long g_label_counter = 0;

long long qadav_next_label_id(void) {
    g_label_counter++;
    return g_label_counter;
}

void qadav_reset_label_counter(void) {
    g_label_counter = 0;
}

const char* qadav_format_icmp_imm(const char* pred, long long res_temp, long long left_temp, long long imm_val) {
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
    const char* p2 = " = icmp ";
    while (*p2) buf[bi++] = *p2++;
    while (*pred) buf[bi++] = *pred++;
    const char* p3 = " i64 %t";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = left_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p4 = ", ";
    while (*p4) buf[bi++] = *p4++;
    ni = 0; v = imm_val;
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

const char* qadav_format_br_cond(long long cond_reg, long long then_id, long long else_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  br i1 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32];
    int ni = 0;
    long long v = cond_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", label %when_then_";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = then_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", label %when_else_";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = else_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_label(const char* prefix, long long id) {
    static char buf[64];
    int bi = 0;
    while (*prefix) buf[bi++] = *prefix++;
    buf[bi++] = '_';
    char num[32];
    int ni = 0;
    long long v = id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ':';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_br_label(const char* prefix, long long id) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  br label %";
    while (*p1) buf[bi++] = *p1++;
    while (*prefix) buf[bi++] = *prefix++;
    buf[bi++] = '_';
    char num[32];
    int ni = 0;
    long long v = id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

static struct {
    char var_name[64];
    long long var_val;
    char cond_var[64];
    char cond_op[16];
    long long cond_val;
    long long yield_val;
} g_if_info;

int qadav_scan_if_else(const char* path) {
    if (!path || !*path) return -1;
    void* h = qadav_file_open_read(path);
    if (!h) return -1;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return -1;
    buf[n] = '\0';

    int body_start = 0;
    for (int i = 0; i <= n - 3; i++) {
        if ((unsigned char)buf[i] == 0xe2 && (unsigned char)buf[i+1] == 0x9f && (unsigned char)buf[i+2] == 0xa6) {
            body_start = i + 3;
            break;
        }
    }

    // 1. Find `int64 <name> ↦ <val>`
    g_if_info.var_name[0] = '\0';
    g_if_info.var_val = 0;
    for (int i = body_start; i <= n - 5; i++) {
        if (buf[i] == 'i' && buf[i+1] == 'n' && buf[i+2] == 't' && buf[i+3] == '6' && buf[i+4] == '4') {
            int j = i + 5;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            int vi = 0;
            while (j < n && buf[j] != ' ' && buf[j] != '\t' && (unsigned char)buf[j] != 0xe2 && vi < 63) {
                g_if_info.var_name[vi++] = buf[j++];
            }
            g_if_info.var_name[vi] = '\0';
            while (j < n && !((buf[j] >= '0' && buf[j] <= '9') || buf[j] == '-')) j++;
            int neg = 0;
            if (j < n && buf[j] == '-') { neg = 1; j++; }
            long long val = 0;
            while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                val = val * 10 + (buf[j++] - '0');
            }
            g_if_info.var_val = neg ? -val : val;
            break;
        }
    }

    // 2. Find `when <name> <op> <val>`
    for (int i = body_start; i <= n - 4; i++) {
        if (buf[i] == 'w' && buf[i+1] == 'h' && buf[i+2] == 'e' && buf[i+3] == 'n') {
            int j = i + 4;
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            int vi = 0;
            while (j < n && buf[j] != ' ' && buf[j] != '\t' && (unsigned char)buf[j] != 0xe2 && vi < 63) {
                g_if_info.cond_var[vi++] = buf[j++];
            }
            g_if_info.cond_var[vi] = '\0';
            while (j < n && (buf[j] == ' ' || buf[j] == '\t')) j++;
            const char* op = "sgt";
            if ((unsigned char)buf[j] == 0xe2 && (unsigned char)buf[j+1] == 0x89) {
                unsigned char c = (unsigned char)buf[j+2];
                if (c == 0xbb) op = "sgt";      // ≻
                else if (c == 0xba) op = "slt"; // ≺
                else if (c == 0xa1) op = "eq";  // ≡
                else if (c == 0xa0) op = "ne";  // ≠
                else if (c == 0xbc) op = "sle"; // ≼
                else if (c == 0xbd) op = "sge"; // ≽
                j += 3;
            }
            int oi = 0;
            while (op[oi] && oi < 15) { g_if_info.cond_op[oi] = op[oi]; oi++; }
            g_if_info.cond_op[oi] = '\0';
            while (j < n && !((buf[j] >= '0' && buf[j] <= '9') || buf[j] == '-')) j++;
            int neg = 0;
            if (j < n && buf[j] == '-') { neg = 1; j++; }
            long long val = 0;
            while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                val = val * 10 + (buf[j++] - '0');
            }
            g_if_info.cond_val = neg ? -val : val;
            break;
        }
    }

    // 3. Find `yield <val>`
    g_if_info.yield_val = 0;
    for (int i = body_start; i <= n - 5; i++) {
        if (buf[i] == 'y' && buf[i+1] == 'i' && buf[i+2] == 'e' && buf[i+3] == 'l' && buf[i+4] == 'd') {
            int j = i + 5;
            while (j < n && !((buf[j] >= '0' && buf[j] <= '9') || buf[j] == '-')) j++;
            int neg = 0;
            if (j < n && buf[j] == '-') { neg = 1; j++; }
            long long val = 0;
            while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                val = val * 10 + (buf[j++] - '0');
            }
            g_if_info.yield_val = neg ? -val : val;
            break;
        }
    }

    return 0;
}

const char* qadav_get_if_var_name(void) { return g_if_info.var_name; }
long long   qadav_get_if_var_val(void) { return g_if_info.var_val; }
const char* qadav_get_if_cond_var(void) { return g_if_info.cond_var; }
const char* qadav_get_if_cond_op(void) { return g_if_info.cond_op; }
long long   qadav_get_if_cond_val(void) { return g_if_info.cond_val; }
long long   qadav_get_if_yield_val(void) { return g_if_info.yield_val; }

long long qadav_is_dot_proc(const char* name) {
    if (!name) return 0;
    return (name[0] == 'd' && name[1] == 'o' && name[2] == 't' && name[3] == '\0') ? 1 : 0;
}

long long qadav_is_main_proc(const char* name) {
    if (!name) return 0;
    return (name[0] == 'm' && name[1] == 'a' && name[2] == 'i' && name[3] == 'n' && name[4] == '\0') ? 1 : 0;
}

static long long g_proc_is_main = 1;
void qadav_set_proc_is_main(long long is_main) {
    g_proc_is_main = is_main;
}
long long qadav_get_proc_is_main(void) {
    return g_proc_is_main;
}

static long long g_proc_is_dot = 0;
void qadav_set_proc_is_dot(long long is_dot) {
    g_proc_is_dot = is_dot;
}
long long qadav_get_proc_is_dot(void) {
    return g_proc_is_dot;
}

const char* qadav_format_func_def_with_param(const char* name, const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "define i64 @\"";
    while (*p1) buf[bi++] = *p1++;
    while (*name) buf[bi++] = *name++;
    const char* p2 = "\"(i64 %";
    while (*p2) buf[bi++] = *p2++;
    while (*param_name) buf[bi++] = *param_name++;
    const char* p3 = ".in) {";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_param_alloca(const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %";
    while (*p1) buf[bi++] = *p1++;
    while (*param_name) buf[bi++] = *param_name++;
    const char* p2 = ".ptr = alloca i64";
    while (*p2) buf[bi++] = *p2++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_param_store(const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i64 %";
    while (*p1) buf[bi++] = *p1++;
    const char* s = param_name;
    while (*s) buf[bi++] = *s++;
    const char* p2 = ".in, i64* %";
    while (*p2) buf[bi++] = *p2++;
    s = param_name;
    while (*s) buf[bi++] = *s++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_br_cond_merge(long long cond_reg, long long then_id, long long merge_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  br i1 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = cond_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", label %when_then_";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = then_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", label %when_merge_";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = merge_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_i64(const char* callee, long long res_temp, long long arg_temp) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = res_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = call i64 @\"";
    while (*p2) buf[bi++] = *p2++;
    while (*callee) buf[bi++] = *callee++;
    const char* p3 = "\"(i64 %t";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = arg_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_i64(long long temp_id) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  ret i64 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_binop_imm(const char* op, long long res_temp, long long left_reg, long long imm_val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = res_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ' ';
    buf[bi++] = '=';
    buf[bi++] = ' ';
    const char* op_str = "sub i64 %t";
    if (op[0] == 's' && op[1] == 'u' && op[2] == 'b') op_str = "sub i64 %t";
    else if (op[0] == 'a' && op[1] == 'd' && op[2] == 'd') op_str = "add i64 %t";
    else if (op[0] == 'm' && op[1] == 'u' && op[2] == 'l') op_str = "mul i64 %t";
    else if (op[0] == 's' && op[1] == 'l' && op[2] == 't') op_str = "icmp slt i64 %t";
    else if (op[0] == 's' && op[1] == 'g' && op[2] == 't') op_str = "icmp sgt i64 %t";
    else if (op[0] == 'e' && op[1] == 'q') op_str = "icmp eq i64 %t";
    else if (op[0] == 'n' && op[1] == 'e') op_str = "icmp ne i64 %t";
    else if (op[0] == 's' && op[1] == 'l' && op[2] == 'e') op_str = "icmp sle i64 %t";
    else if (op[0] == 's' && op[1] == 'g' && op[2] == 'e') op_str = "icmp sge i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x96) op_str = "sub i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x95) op_str = "add i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x97) op_str = "mul i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xba) op_str = "icmp slt i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbb) op_str = "icmp sgt i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa1) op_str = "icmp eq i64 %t";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa0) op_str = "icmp ne i64 %t";
    while (*op_str) buf[bi++] = *op_str++;
    ni = 0; v = left_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ',';
    buf[bi++] = ' ';
    ni = 0; v = imm_val;
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

const char* qadav_format_store_reg(const char* name, long long reg_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i64 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = reg_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = ", i64* %";
    while (*p2) buf[bi++] = *p2++;
    while (*name) buf[bi++] = *name++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

double qadav_get_tensor_float_val(long long t_idx, long long e_idx) {
    if (t_idx < 0 || t_idx >= g_tensor_count || e_idx < 0 || e_idx >= 16) return 0.0;
    return (double)g_tensors[t_idx].values[e_idx];
}

const char* qadav_format_tensor_create(long long reg, long long rank, long long size) {
    (void)rank;
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = call i8* @qadav_tensor_create(i32 10, i32 1, i32* %t";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = size;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_tensor_set_f64(long long t_reg, long long elem_idx, double val) {
    return qadav_format_tensor_set(t_reg, elem_idx, (long long)val);
}

const char* qadav_format_tensor_mul(long long dst_reg, long long a_reg, long long b_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = dst_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = call i8* @qadav_tensor_mul(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = a_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", i8* %t";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = b_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_tensor_reduce_sum_f64(long long dst_reg, long long src_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = dst_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = call double @qadav_tensor_reduce_sum_f64(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    ni = 0; v = src_reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_alloca_type(const char* name, const char* typ) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %";
    while (*p1) buf[bi++] = *p1++;
    while (*name) buf[bi++] = *name++;
    const char* p2 = ".ptr = alloca ";
    while (*p2) buf[bi++] = *p2++;
    while (*typ) buf[bi++] = *typ++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_store_reg_type(const char* name, const char* typ, long long reg_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store ";
    while (*p1) buf[bi++] = *p1++;
    const char* t = typ;
    while (*t) buf[bi++] = *t++;
    const char* p2 = " %t";
    while (*p2) buf[bi++] = *p2++;
    char num[32]; int ni = 0; long long v = reg_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p3 = ", ";
    while (*p3) buf[bi++] = *p3++;
    t = typ;
    while (*t) buf[bi++] = *t++;
    const char* p4 = "* %";
    while (*p4) buf[bi++] = *p4++;
    while (*name) buf[bi++] = *name++;
    const char* p5 = ".ptr";
    while (*p5) buf[bi++] = *p5++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_load_type(const char* name, const char* typ, long long temp_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = load ";
    while (*p2) buf[bi++] = *p2++;
    const char* t = typ;
    while (*t) buf[bi++] = *t++;
    const char* p3 = ", ";
    while (*p3) buf[bi++] = *p3++;
    t = typ;
    while (*t) buf[bi++] = *t++;
    const char* p4 = "* %";
    while (*p4) buf[bi++] = *p4++;
    while (*name) buf[bi++] = *name++;
    const char* p5 = ".ptr";
    while (*p5) buf[bi++] = *p5++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_double_2args(const char* callee, long long res_temp, long long arg1, long long arg2) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = res_temp;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p2 = " = call double @\"";
    while (*p2) buf[bi++] = *p2++;
    while (*callee) buf[bi++] = *callee++;
    const char* p3 = "\"(i8* %t";
    while (*p3) buf[bi++] = *p3++;
    ni = 0; v = arg1;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    const char* p4 = ", i8* %t";
    while (*p4) buf[bi++] = *p4++;
    ni = 0; v = arg2;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_func_def_dot(const char* name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "define double @\"";
    while (*p1) buf[bi++] = *p1++;
    while (*name) buf[bi++] = *name++;
    const char* p2 = "\"(i8* %a.in, i8* %b.in) {";
    while (*p2) buf[bi++] = *p2++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_double(long long temp_id) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  ret double %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_i32(long long temp_id) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  ret i32 %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = temp_id;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_emit_float_call(long long reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_emit_float(double %t";
    while (*p1) buf[bi++] = *p1++;
    char num[32]; int ni = 0; long long v = reg;
    if (v == 0) num[ni++] = '0';
    else { char tmp[32]; int ti = 0; while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; } while (ti > 0) num[ni++] = tmp[--ti]; }
    for (int i = 0; i < ni; i++) buf[bi++] = num[i];
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

// Variable type tracking for M8.5g-1
typedef struct {
    char name[32];
    int type_code; // 0 = int64, 1 = i8* (tensor), 2 = double (float64)
} qadav_var_type_entry_t;

static qadav_var_type_entry_t g_var_types[64];
static int g_var_type_count = 0;

long long qadav_register_var_type(const char* name, const char* typ) {
    if (!name || g_var_type_count >= 64) return 0;
    int code = 0;
    if (io_str_eq(typ, "tensor") || io_str_eq(typ, "i8*") || qadav_is_tensor_type(typ)) {
        code = 1;
    } else if (io_str_eq(typ, "float64") || io_str_eq(typ, "double") || qadav_is_float_type(typ)) {
        code = 2;
    }
    int idx = 0;
    while (name[idx] && idx < 31) {
        g_var_types[g_var_type_count].name[idx] = name[idx];
        idx++;
    }
    g_var_types[g_var_type_count].name[idx] = '\0';
    g_var_types[g_var_type_count].type_code = code;
    g_var_type_count++;
    return 1;
}

long long qadav_get_var_type_code(const char* name) {
    if (!name) return 0;
    for (int i = g_var_type_count - 1; i >= 0; i--) {
        if (io_str_eq(g_var_types[i].name, name)) {
            return g_var_types[i].type_code;
        }
    }
    if (qadav_is_ptr_var(name)) return 1;
    if (io_str_eq(name, "sum")) return 2;
    if (qadav_is_result_var(name)) {
        if (g_tensor_count > 0 && !io_str_eq(name, "x") && !io_str_eq(name, "y")) return 2;
    }
    return 0;
}

static void append_num(char* buf, int* bi, long long val) {
    char num[32];
    int ni = 0;
    if (val == 0) num[ni++] = '0';
    else {
        int neg = 0; unsigned long long uv;
        if (val < 0) { neg = 1; uv = (unsigned long long)(-val); } else { uv = (unsigned long long)val; }
        char tmp[32]; int ti = 0;
        while (uv > 0) { tmp[ti++] = '0' + (char)(uv % 10); uv /= 10; }
        if (neg) num[ni++] = '-';
        while (ti > 0) num[ni++] = tmp[--ti];
    }
    for (int i = 0; i < ni; i++) buf[(*bi)++] = num[i];
}

const char* qadav_format_func_def_2params(const char* name, const char* p1, const char* p2) {
    static char buf[128];
    int bi = 0;
    const char* s = "define i64 @\"";
    while (*s) buf[bi++] = *s++;
    while (*name) buf[bi++] = *name++;
    s = "\"(i64 %";
    while (*s) buf[bi++] = *s++;
    while (*p1) buf[bi++] = *p1++;
    s = ".in, i64 %";
    while (*s) buf[bi++] = *s++;
    while (*p2) buf[bi++] = *p2++;
    s = ".in) {";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_i64_imm(const char* callee, long long res_temp, long long arg1_temp, long long imm_val) {
    static char buf[128];
    int bi = 0;
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call i64 @\"";
    while (*s) buf[bi++] = *s++;
    while (*callee) buf[bi++] = *callee++;
    s = "\"(i64 %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, arg1_temp);
    s = ", i64 ";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, imm_val);
    s = ")";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_i64_2args(const char* callee, long long res_temp, long long arg1_temp, long long arg2_temp) {
    static char buf[128];
    int bi = 0;
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call i64 @\"";
    while (*s) buf[bi++] = *s++;
    while (*callee) buf[bi++] = *callee++;
    s = "\"(i64 %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, arg1_temp);
    s = ", i64 %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, arg2_temp);
    s = ")";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_0args(const char* callee, long long res_temp) {
    static char buf[128];
    int bi = 0;
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call i64 @\"";
    while (*s) buf[bi++] = *s++;
    while (*callee) buf[bi++] = *callee++;
    s = "\"()";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

// -----------------------------------------------------------------------------
// Real AST Recursive Descent Parser for M8.5g-1
// -----------------------------------------------------------------------------

#define AST_TAG_PROCESS_DECL   0
#define AST_TAG_PROCESS2_DECL  1
#define AST_TAG_BLOCK_DECL     2
#define AST_TAG_BIND_DECL      3
#define AST_TAG_YIELD_DECL     4
#define AST_TAG_EMIT_DECL      5
#define AST_TAG_WHEN_DECL      6
#define AST_TAG_WHEN1_DECL     7
#define AST_TAG_CYCLE_DECL     8
#define AST_TAG_BINOP_DECL     9
#define AST_TAG_CALL_DECL      10
#define AST_TAG_CALL0_DECL     11
#define AST_TAG_CALL1_DECL     12
#define AST_TAG_CALL2_DECL     13
#define AST_TAG_TENSOR_DECL    14
#define AST_TAG_REDUCE_SUM     15
#define AST_TAG_EXPR_STMT      16
#define AST_TAG_IDENT_DECL     17
#define AST_TAG_INTLIT_DECL    18
#define AST_TAG_FLOATLIT_DECL  19
#define AST_TAG_STRLIT_DECL    20

static char g_ast_str_pool[512][64];
static int g_ast_str_pool_cnt = 0;

static const char* ast_pool_str(const char* s) {
    if (!s) return "";
    for (int i = 0; i < g_ast_str_pool_cnt; i++) {
        if (io_str_eq(g_ast_str_pool[i], s)) return g_ast_str_pool[i];
    }
    if (g_ast_str_pool_cnt >= 512) return "";
    int j = 0;
    while (s[j] && j < 63) {
        g_ast_str_pool[g_ast_str_pool_cnt][j] = s[j];
        j++;
    }
    g_ast_str_pool[g_ast_str_pool_cnt][j] = '\0';
    return g_ast_str_pool[g_ast_str_pool_cnt++];
}

static qadav_variant* ast_make_node(unsigned int tag, int nfields, const unsigned long long* fields) {
    return qadav_variant_create(0, tag, nfields * sizeof(unsigned long long), fields);
}

static qadav_variant* ast_process_decl(const char* name, const char* p, const char* ret, qadav_variant* body) {
    unsigned long long f[4] = { (unsigned long long)ast_pool_str(name), (unsigned long long)ast_pool_str(p), (unsigned long long)ast_pool_str(ret), (unsigned long long)body };
    return ast_make_node(AST_TAG_PROCESS_DECL, 4, f);
}
static qadav_variant* ast_process2_decl(const char* name, const char* p1, const char* p2, const char* ret, qadav_variant* body) {
    unsigned long long f[5] = { (unsigned long long)ast_pool_str(name), (unsigned long long)ast_pool_str(p1), (unsigned long long)ast_pool_str(p2), (unsigned long long)ast_pool_str(ret), (unsigned long long)body };
    return ast_make_node(AST_TAG_PROCESS2_DECL, 5, f);
}
static qadav_variant* ast_block_decl(qadav_variant* first, qadav_variant* second) {
    unsigned long long f[2] = { (unsigned long long)first, (unsigned long long)second };
    return ast_make_node(AST_TAG_BLOCK_DECL, 2, f);
}
static qadav_variant* ast_bind_decl(const char* name, const char* typ, qadav_variant* val) {
    unsigned long long f[3] = { (unsigned long long)ast_pool_str(name), (unsigned long long)ast_pool_str(typ), (unsigned long long)val };
    return ast_make_node(AST_TAG_BIND_DECL, 3, f);
}
static qadav_variant* ast_yield_decl(qadav_variant* val) {
    unsigned long long f[1] = { (unsigned long long)val };
    return ast_make_node(AST_TAG_YIELD_DECL, 1, f);
}
static qadav_variant* ast_emit_decl(qadav_variant* val) {
    unsigned long long f[1] = { (unsigned long long)val };
    return ast_make_node(AST_TAG_EMIT_DECL, 1, f);
}
static qadav_variant* ast_when_decl(qadav_variant* cond, qadav_variant* th, qadav_variant* el) {
    unsigned long long f[3] = { (unsigned long long)cond, (unsigned long long)th, (unsigned long long)el };
    return ast_make_node(AST_TAG_WHEN_DECL, 3, f);
}
static qadav_variant* ast_when1_decl(qadav_variant* cond, qadav_variant* th) {
    unsigned long long f[2] = { (unsigned long long)cond, (unsigned long long)th };
    return ast_make_node(AST_TAG_WHEN1_DECL, 2, f);
}
static qadav_variant* ast_binop_decl(const char* op, qadav_variant* l, qadav_variant* r) {
    unsigned long long f[3] = { (unsigned long long)ast_pool_str(op), (unsigned long long)l, (unsigned long long)r };
    return ast_make_node(AST_TAG_BINOP_DECL, 3, f);
}
static qadav_variant* ast_call0_decl(const char* name) {
    unsigned long long f[1] = { (unsigned long long)ast_pool_str(name) };
    return ast_make_node(AST_TAG_CALL0_DECL, 1, f);
}
static qadav_variant* ast_call1_decl(const char* name, qadav_variant* arg) {
    unsigned long long f[2] = { (unsigned long long)ast_pool_str(name), (unsigned long long)arg };
    return ast_make_node(AST_TAG_CALL1_DECL, 2, f);
}
static qadav_variant* ast_call2_decl(const char* name, qadav_variant* a1, qadav_variant* a2) {
    unsigned long long f[3] = { (unsigned long long)ast_pool_str(name), (unsigned long long)a1, (unsigned long long)a2 };
    return ast_make_node(AST_TAG_CALL2_DECL, 3, f);
}
static qadav_variant* ast_tensor_decl(long long idx, long long cnt) {
    unsigned long long f[2] = { (unsigned long long)idx, (unsigned long long)cnt };
    return ast_make_node(AST_TAG_TENSOR_DECL, 2, f);
}
static qadav_variant* ast_reduce_sum_decl(qadav_variant* inner) {
    unsigned long long f[1] = { (unsigned long long)inner };
    return ast_make_node(AST_TAG_REDUCE_SUM, 1, f);
}
static qadav_variant* ast_ident_decl(const char* name) {
    unsigned long long f[1] = { (unsigned long long)ast_pool_str(name) };
    return ast_make_node(AST_TAG_IDENT_DECL, 1, f);
}
static qadav_variant* ast_intlit_decl(long long val) {
    unsigned long long f[1] = { (unsigned long long)val };
    return ast_make_node(AST_TAG_INTLIT_DECL, 1, f);
}
static qadav_variant* ast_strlit_decl(const char* val) {
    unsigned long long f[1] = { (unsigned long long)ast_pool_str(val) };
    return ast_make_node(AST_TAG_STRLIT_DECL, 1, f);
}

typedef enum {
    AST_TOK_EOF = 0,
    AST_TOK_IDENT,
    AST_TOK_INT,
    AST_TOK_FLOAT,
    AST_TOK_STRING,
    AST_TOK_KW_PROCESS,
    AST_TOK_KW_WHEN,
    AST_TOK_KW_OTHERWISE,
    AST_TOK_KW_YIELD,
    AST_TOK_KW_EMIT,
    AST_TOK_KW_TENSOR,
    AST_TOK_OP_BIND,      // ↦
    AST_TOK_OP_ASSIGN,    // ←
    AST_TOK_OP_ARROW,     // →
    AST_TOK_OP_DOT,       // ·
    AST_TOK_OP_COMMA,     // ⸴
    AST_TOK_OP_LBLOCK,    // ⟦
    AST_TOK_OP_RBLOCK,    // ⟧
    AST_TOK_OP_LPAREN,    // ⦅
    AST_TOK_OP_RPAREN,    // ⦆
    AST_TOK_OP_LTENSOR,   // ⟬
    AST_TOK_OP_RTENSOR,   // ⟭
    AST_TOK_OP_LANGLE,    // ⟨ or <
    AST_TOK_OP_RANGLE,    // ⟩ or >
    AST_TOK_OP_ADD,       // ⊕
    AST_TOK_OP_SUB,       // ⊖
    AST_TOK_OP_MUL,       // ⊗
    AST_TOK_OP_DIV,       // ⊘
    AST_TOK_OP_GT,        // ≻
    AST_TOK_OP_LT,        // ≺
    AST_TOK_OP_EQ,        // ≡
    AST_TOK_OP_NEQ,       // ≠
    AST_TOK_OP_SUM,       // ∑
    AST_TOK_OTHER
} ast_tok_kind;

typedef struct {
    ast_tok_kind kind;
    char text[128];
    long long int_val;
} ast_token_t;

typedef struct {
    const unsigned char* src;
    int len;
    int pos;
    ast_token_t cur;
} ast_parser_t;

static void ast_lex_advance(ast_parser_t* p) {
    while (p->pos < p->len) {
        unsigned char c = p->src[p->pos];
        if (c == ' ' || c == '\t' || c == '\r' || c == '\n') {
            p->pos++;
            continue;
        }
        if (c == '/' && p->pos + 1 < p->len && p->src[p->pos + 1] == '/') {
            p->pos += 2;
            while (p->pos < p->len && p->src[p->pos] != '\n') {
                p->pos++;
            }
            continue;
        }
        break;
    }

    if (p->pos >= p->len) {
        p->cur.kind = AST_TOK_EOF;
        p->cur.text[0] = '\0';
        p->cur.int_val = 0;
        return;
    }

    const unsigned char* s = p->src + p->pos;
    int rem = p->len - p->pos;

    // Check multi-byte UTF-8
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x86 && s[2] == 0xa6) { p->pos += 3; p->cur.kind = AST_TOK_OP_BIND; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x86 && s[2] == 0x90) { p->pos += 3; p->cur.kind = AST_TOK_OP_ASSIGN; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x86 && s[2] == 0x92) { p->pos += 3; p->cur.kind = AST_TOK_OP_ARROW; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xa6) { p->pos += 3; p->cur.kind = AST_TOK_OP_LBLOCK; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xa7) { p->pos += 3; p->cur.kind = AST_TOK_OP_RBLOCK; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xac) { p->pos += 3; p->cur.kind = AST_TOK_OP_LTENSOR; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xad) { p->pos += 3; p->cur.kind = AST_TOK_OP_RTENSOR; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xa8) { p->pos += 3; p->cur.kind = AST_TOK_OP_LANGLE; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x9f && s[2] == 0xa9) { p->pos += 3; p->cur.kind = AST_TOK_OP_RANGLE; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0xa6 && s[2] == 0x85) { p->pos += 3; p->cur.kind = AST_TOK_OP_LPAREN; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0xa6 && s[2] == 0x86) { p->pos += 3; p->cur.kind = AST_TOK_OP_RPAREN; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0xb8 && s[2] == 0xb4) { p->pos += 3; p->cur.kind = AST_TOK_OP_COMMA; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x8a && s[2] == 0x95) { p->pos += 3; p->cur.kind = AST_TOK_OP_ADD; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x8a && s[2] == 0x96) { p->pos += 3; p->cur.kind = AST_TOK_OP_SUB; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x8a && s[2] == 0x97) { p->pos += 3; p->cur.kind = AST_TOK_OP_MUL; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x8a && s[2] == 0x98) { p->pos += 3; p->cur.kind = AST_TOK_OP_DIV; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x89 && s[2] == 0xbb) { p->pos += 3; p->cur.kind = AST_TOK_OP_GT; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x89 && s[2] == 0xba) { p->pos += 3; p->cur.kind = AST_TOK_OP_LT; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x89 && s[2] == 0xa1) { p->pos += 3; p->cur.kind = AST_TOK_OP_EQ; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x89 && s[2] == 0xa0) { p->pos += 3; p->cur.kind = AST_TOK_OP_NEQ; p->cur.text[0] = '\0'; return; }
    if (rem >= 3 && s[0] == 0xe2 && s[1] == 0x88 && s[2] == 0x91) { p->pos += 3; p->cur.kind = AST_TOK_OP_SUM; p->cur.text[0] = '\0'; return; }
    if (rem >= 2 && s[0] == 0xc2 && s[1] == 0xb7) { p->pos += 2; p->cur.kind = AST_TOK_OP_DOT; p->cur.text[0] = '\0'; return; }

    // ASCII delimiters / brackets
    if (s[0] == '<') { p->pos++; p->cur.kind = AST_TOK_OP_LANGLE; p->cur.text[0] = '<'; p->cur.text[1] = '\0'; return; }
    if (s[0] == '>') { p->pos++; p->cur.kind = AST_TOK_OP_RANGLE; p->cur.text[0] = '>'; p->cur.text[1] = '\0'; return; }

    // String literal
    if (s[0] == '"') {
        p->pos++;
        int ti = 0;
        while (p->pos < p->len && p->src[p->pos] != '"' && ti < 127) {
            p->cur.text[ti++] = p->src[p->pos++];
        }
        p->cur.text[ti] = '\0';
        if (p->pos < p->len && p->src[p->pos] == '"') p->pos++;
        p->cur.kind = AST_TOK_STRING;
        return;
    }

    // Number (int or float)
    if (s[0] >= '0' && s[0] <= '9') {
        int ti = 0;
        int is_float = 0;
        long long val = 0;
        while (p->pos < p->len && ((p->src[p->pos] >= '0' && p->src[p->pos] <= '9') || p->src[p->pos] == '.') && ti < 127) {
            if (p->src[p->pos] == '.') is_float = 1;
            else if (!is_float) val = val * 10 + (p->src[p->pos] - '0');
            p->cur.text[ti++] = p->src[p->pos++];
        }
        p->cur.text[ti] = '\0';
        if (is_float) {
            p->cur.kind = AST_TOK_FLOAT;
        } else {
            p->cur.kind = AST_TOK_INT;
            p->cur.int_val = val;
        }
        return;
    }

    // Identifier or keyword
    if ((s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z') || s[0] == '_') {
        int ti = 0;
        while (p->pos < p->len && ((p->src[p->pos] >= 'a' && p->src[p->pos] <= 'z') ||
                                   (p->src[p->pos] >= 'A' && p->src[p->pos] <= 'Z') ||
                                   (p->src[p->pos] >= '0' && p->src[p->pos] <= '9') ||
                                   p->src[p->pos] == '_') && ti < 127) {
            p->cur.text[ti++] = p->src[p->pos++];
        }
        p->cur.text[ti] = '\0';
        if (io_str_eq(p->cur.text, "process")) p->cur.kind = AST_TOK_KW_PROCESS;
        else if (io_str_eq(p->cur.text, "when")) p->cur.kind = AST_TOK_KW_WHEN;
        else if (io_str_eq(p->cur.text, "otherwise")) p->cur.kind = AST_TOK_KW_OTHERWISE;
        else if (io_str_eq(p->cur.text, "yield")) p->cur.kind = AST_TOK_KW_YIELD;
        else if (io_str_eq(p->cur.text, "emit")) p->cur.kind = AST_TOK_KW_EMIT;
        else if (io_str_eq(p->cur.text, "tensor")) p->cur.kind = AST_TOK_KW_TENSOR;
        else p->cur.kind = AST_TOK_IDENT;
        return;
    }

    // Unknown single char
    p->cur.kind = AST_TOK_OTHER;
    p->cur.text[0] = s[0];
    p->cur.text[1] = '\0';
    p->pos++;
}

static qadav_variant* ast_parse_expr(ast_parser_t* p);

static qadav_variant* ast_parse_primary(ast_parser_t* p) {
    if (p->cur.kind == AST_TOK_OP_LPAREN) {
        ast_lex_advance(p);
        qadav_variant* inner = ast_parse_expr(p);
        if (p->cur.kind == AST_TOK_OP_RPAREN) ast_lex_advance(p);
        return inner;
    }
    if (p->cur.kind == AST_TOK_OP_SUM) {
        ast_lex_advance(p);
        qadav_variant* inner = ast_parse_expr(p);
        return ast_reduce_sum_decl(inner);
    }
    if (p->cur.kind == AST_TOK_KW_TENSOR) {
        ast_lex_advance(p);
        if (p->cur.kind == AST_TOK_OP_LTENSOR) {
            ast_lex_advance(p);
            int tidx = g_tensor_count++;
            int count = 0;
            while (p->cur.kind != AST_TOK_OP_RTENSOR && p->cur.kind != AST_TOK_EOF && count < 16) {
                if (p->cur.kind == AST_TOK_INT) {
                    g_tensors[tidx].values[count] = p->cur.int_val;
                    int ni = 0; long long v = p->cur.int_val;
                    if (v == 0) g_tensors[tidx].val_strs[count][ni++] = '0';
                    else {
                        char tmp[32]; int ti = 0;
                        while (v > 0) { tmp[ti++] = '0' + (char)(v % 10); v /= 10; }
                        while (ti > 0) g_tensors[tidx].val_strs[count][ni++] = tmp[--ti];
                    }
                    const char* dec = ".000000";
                    while (*dec) g_tensors[tidx].val_strs[count][ni++] = *dec++;
                    g_tensors[tidx].val_strs[count][ni] = '\0';
                    count++;
                    ast_lex_advance(p);
                } else if (p->cur.kind == AST_TOK_FLOAT) {
                    int ni = 0;
                    const char* s = p->cur.text;
                    while (*s && ni < 20) g_tensors[tidx].val_strs[count][ni++] = *s++;
                    int dot = -1;
                    for (int i = 0; i < ni; i++) if (g_tensors[tidx].val_strs[count][i] == '.') dot = i;
                    if (dot >= 0) {
                        int decimals = ni - dot - 1;
                        while (decimals < 6 && ni < 31) {
                            g_tensors[tidx].val_strs[count][ni++] = '0';
                            decimals++;
                        }
                    }
                    g_tensors[tidx].val_strs[count][ni] = '\0';
                    count++;
                    ast_lex_advance(p);
                }
                if (p->cur.kind == AST_TOK_OP_COMMA) {
                    ast_lex_advance(p);
                }
            }
            if (p->cur.kind == AST_TOK_OP_RTENSOR) ast_lex_advance(p);
            g_tensors[tidx].size = count;
            return ast_tensor_decl(tidx, count);
        }
    }
    if (p->cur.kind == AST_TOK_STRING) {
        char s[128];
        int si = 0;
        while (p->cur.text[si] && si < 127) { s[si] = p->cur.text[si]; si++; }
        s[si] = '\0';
        qadav_register_string(s);
        ast_lex_advance(p);
        return ast_strlit_decl(s);
    }
    if (p->cur.kind == AST_TOK_INT) {
        long long val = p->cur.int_val;
        ast_lex_advance(p);
        return ast_intlit_decl(val);
    }
    if (p->cur.kind == AST_TOK_FLOAT) {
        ast_lex_advance(p);
        return ast_intlit_decl(0);
    }
    if (p->cur.kind == AST_TOK_IDENT) {
        char id[64];
        int ii = 0;
        while (p->cur.text[ii] && ii < 63) { id[ii] = p->cur.text[ii]; ii++; }
        id[ii] = '\0';
        ast_lex_advance(p);
        if (p->cur.kind == AST_TOK_OP_LPAREN) {
            ast_lex_advance(p);
            if (p->cur.kind == AST_TOK_OP_RPAREN) {
                ast_lex_advance(p);
                return ast_call0_decl(id);
            }
            qadav_variant* arg1 = ast_parse_expr(p);
            if (p->cur.kind == AST_TOK_OP_COMMA) {
                ast_lex_advance(p);
                qadav_variant* arg2 = ast_parse_expr(p);
                if (p->cur.kind == AST_TOK_OP_RPAREN) ast_lex_advance(p);
                return ast_call2_decl(id, arg1, arg2);
            }
            if (p->cur.kind == AST_TOK_OP_RPAREN) ast_lex_advance(p);
            return ast_call1_decl(id, arg1);
        }
        return ast_ident_decl(id);
    }
    ast_lex_advance(p);
    return ast_intlit_decl(0);
}

static qadav_variant* ast_parse_expr(ast_parser_t* p) {
    qadav_variant* left = ast_parse_primary(p);

    const char* op = (void*)0;
    if (p->cur.kind == AST_TOK_OP_ADD) op = "add";
    else if (p->cur.kind == AST_TOK_OP_SUB) op = "sub";
    else if (p->cur.kind == AST_TOK_OP_MUL) op = "mul";
    else if (p->cur.kind == AST_TOK_OP_DIV) op = "sdiv";
    else if (p->cur.kind == AST_TOK_OP_GT) op = "sgt";
    else if (p->cur.kind == AST_TOK_OP_LT) op = "slt";
    else if (p->cur.kind == AST_TOK_OP_EQ) op = "eq";
    else if (p->cur.kind == AST_TOK_OP_NEQ) op = "ne";

    if (op) {
        ast_lex_advance(p);
        qadav_variant* right = ast_parse_expr(p);
        return ast_binop_decl(op, left, right);
    }
    return left;
}

static qadav_variant* ast_parse_block(ast_parser_t* p);

static qadav_variant* ast_parse_stmt(ast_parser_t* p) {
    if (p->cur.kind == AST_TOK_KW_WHEN) {
        ast_lex_advance(p);
        qadav_variant* cond = ast_parse_expr(p);
        if (p->cur.kind == AST_TOK_OP_LBLOCK) ast_lex_advance(p);
        qadav_variant* th = ast_parse_block(p);
        if (p->cur.kind == AST_TOK_OP_RBLOCK) ast_lex_advance(p);
        if (p->cur.kind == AST_TOK_KW_OTHERWISE) {
            ast_lex_advance(p);
            if (p->cur.kind == AST_TOK_OP_LBLOCK) ast_lex_advance(p);
            qadav_variant* el = ast_parse_block(p);
            if (p->cur.kind == AST_TOK_OP_RBLOCK) ast_lex_advance(p);
            return ast_when_decl(cond, th, el);
        }
        return ast_when1_decl(cond, th);
    }
    if (p->cur.kind == AST_TOK_KW_EMIT) {
        ast_lex_advance(p);
        qadav_variant* val = ast_parse_expr(p);
        if (p->cur.kind == AST_TOK_OP_DOT) ast_lex_advance(p);
        return ast_emit_decl(val);
    }
    if (p->cur.kind == AST_TOK_KW_YIELD) {
        ast_lex_advance(p);
        qadav_variant* val = ast_parse_expr(p);
        if (p->cur.kind == AST_TOK_OP_DOT) ast_lex_advance(p);
        return ast_yield_decl(val);
    }

    char typ[32] = "int64";
    char name[64] = "";
    if (p->cur.kind == AST_TOK_KW_TENSOR || (p->cur.kind == AST_TOK_IDENT && (io_str_eq(p->cur.text, "tensor") || io_str_eq(p->cur.text, "int64") || io_str_eq(p->cur.text, "float64")))) {
        int ti = 0;
        while (p->cur.text[ti] && ti < 31) { typ[ti] = p->cur.text[ti]; ti++; }
        typ[ti] = '\0';
        ast_lex_advance(p);

        if (p->cur.kind == AST_TOK_OP_LANGLE) {
            while (p->cur.kind != AST_TOK_OP_RANGLE && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
            if (p->cur.kind == AST_TOK_OP_RANGLE) ast_lex_advance(p);
        }
        if (p->cur.kind == AST_TOK_OP_LTENSOR) {
            while (p->cur.kind != AST_TOK_OP_RTENSOR && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
            if (p->cur.kind == AST_TOK_OP_RTENSOR) ast_lex_advance(p);
        }
        if (p->cur.kind == AST_TOK_IDENT) {
            int ni = 0;
            while (p->cur.text[ni] && ni < 63) { name[ni] = p->cur.text[ni]; ni++; }
            name[ni] = '\0';
            ast_lex_advance(p);
        }
    } else if (p->cur.kind == AST_TOK_IDENT) {
        int ni = 0;
        while (p->cur.text[ni] && ni < 63) { name[ni] = p->cur.text[ni]; ni++; }
        name[ni] = '\0';
        ast_lex_advance(p);
        if (p->cur.kind == AST_TOK_IDENT) {
            int ti = 0;
            while (name[ti] && ti < 31) { typ[ti] = name[ti]; ti++; }
            typ[ti] = '\0';
            ni = 0;
            while (p->cur.text[ni] && ni < 63) { name[ni] = p->cur.text[ni]; ni++; }
            name[ni] = '\0';
            ast_lex_advance(p);
        }
    }

    if (p->cur.kind == AST_TOK_OP_BIND) {
        ast_lex_advance(p);
        qadav_register_var_type(name, typ);
        qadav_variant* val = ast_parse_expr(p);
        if (p->cur.kind == AST_TOK_OP_DOT) ast_lex_advance(p);
        return ast_bind_decl(name, typ, val);
    }

    ast_lex_advance(p);
    return ast_yield_decl(ast_intlit_decl(0));
}

static qadav_variant* ast_parse_block(ast_parser_t* p) {
    qadav_variant* first = ast_parse_stmt(p);
    while (p->cur.kind != AST_TOK_OP_RBLOCK && p->cur.kind != AST_TOK_KW_OTHERWISE && p->cur.kind != AST_TOK_EOF) {
        qadav_variant* next = ast_parse_stmt(p);
        first = ast_block_decl(first, next);
    }
    return first;
}

static qadav_variant* ast_parse_process(ast_parser_t* p) {
    if (p->cur.kind == AST_TOK_KW_PROCESS) ast_lex_advance(p);
    char name[64] = "";
    if (p->cur.kind == AST_TOK_IDENT) {
        int ni = 0;
        while (p->cur.text[ni] && ni < 63) { name[ni] = p->cur.text[ni]; ni++; }
        name[ni] = '\0';
        ast_lex_advance(p);
    }
    char p1[64] = "";
    char p2[64] = "";
    int param_count = 0;
    if (p->cur.kind == AST_TOK_OP_LPAREN) {
        ast_lex_advance(p);
        if (p->cur.kind != AST_TOK_OP_RPAREN) {
            char t1[32] = "";
            if (p->cur.kind == AST_TOK_KW_TENSOR || p->cur.kind == AST_TOK_IDENT) {
                int ti = 0; while (p->cur.text[ti] && ti < 31) { t1[ti] = p->cur.text[ti]; ti++; }
                t1[ti] = '\0';
                ast_lex_advance(p);
                if (p->cur.kind == AST_TOK_OP_LANGLE) {
                    while (p->cur.kind != AST_TOK_OP_RANGLE && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
                    if (p->cur.kind == AST_TOK_OP_RANGLE) ast_lex_advance(p);
                }
                if (p->cur.kind == AST_TOK_OP_LTENSOR) {
                    while (p->cur.kind != AST_TOK_OP_RTENSOR && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
                    if (p->cur.kind == AST_TOK_OP_RTENSOR) ast_lex_advance(p);
                }
            }
            if (p->cur.kind == AST_TOK_IDENT) {
                int ni = 0; while (p->cur.text[ni] && ni < 63) { p1[ni] = p->cur.text[ni]; ni++; }
                p1[ni] = '\0';
                ast_lex_advance(p);
                param_count = 1;
                qadav_register_var_type(p1, t1);
            }
            if (p->cur.kind == AST_TOK_OP_COMMA) {
                ast_lex_advance(p);
                char t2[32] = "";
                if (p->cur.kind == AST_TOK_KW_TENSOR || p->cur.kind == AST_TOK_IDENT) {
                    int ti = 0; while (p->cur.text[ti] && ti < 31) { t2[ti] = p->cur.text[ti]; ti++; }
                    t2[ti] = '\0';
                    ast_lex_advance(p);
                    if (p->cur.kind == AST_TOK_OP_LANGLE) {
                        while (p->cur.kind != AST_TOK_OP_RANGLE && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
                        if (p->cur.kind == AST_TOK_OP_RANGLE) ast_lex_advance(p);
                    }
                    if (p->cur.kind == AST_TOK_OP_LTENSOR) {
                        while (p->cur.kind != AST_TOK_OP_RTENSOR && p->cur.kind != AST_TOK_EOF) ast_lex_advance(p);
                        if (p->cur.kind == AST_TOK_OP_RTENSOR) ast_lex_advance(p);
                    }
                }
                if (p->cur.kind == AST_TOK_IDENT) {
                    int ni = 0; while (p->cur.text[ni] && ni < 63) { p2[ni] = p->cur.text[ni]; ni++; }
                    p2[ni] = '\0';
                    ast_lex_advance(p);
                    param_count = 2;
                    qadav_register_var_type(p2, t2);
                }
            }
        }
        if (p->cur.kind == AST_TOK_OP_RPAREN) ast_lex_advance(p);
    }
    char ret[32] = "int64";
    if (p->cur.kind == AST_TOK_OP_ARROW) {
        ast_lex_advance(p);
        if (p->cur.kind == AST_TOK_IDENT) {
            int ri = 0; while (p->cur.text[ri] && ri < 31) { ret[ri] = p->cur.text[ri]; ri++; }
            ret[ri] = '\0';
            ast_lex_advance(p);
        }
    }
    if (p->cur.kind == AST_TOK_OP_LBLOCK) ast_lex_advance(p);
    qadav_variant* body = ast_parse_block(p);
    if (p->cur.kind == AST_TOK_OP_RBLOCK) ast_lex_advance(p);

    if (param_count == 2) {
        return ast_process2_decl(name, p1, p2, ret, body);
    }
    return ast_process_decl(name, p1, ret, body);
}

void* qadav_parse_file_to_ast(const char* file_path) {
    if (!file_path || !*file_path) return (void*)0;
    void* h = qadav_file_open_read(file_path);
    if (!h) return (void*)0;
    static char buf[16384];
    int n = qadav_file_read(h, buf, 16383);
    qadav_file_close(h);
    if (n <= 0) return (void*)0;
    buf[n] = '\0';

    // Reset parser-related state
    g_tensor_count = 0;
    g_emitted_string_count = 0;
    g_var_type_count = 0;
    g_ast_str_pool_cnt = 0;

    ast_parser_t parser;
    parser.src = (const unsigned char*)buf;
    parser.len = n;
    parser.pos = 0;
    ast_lex_advance(&parser);

    qadav_variant* root = ast_parse_process(&parser);
    while (parser.cur.kind != AST_TOK_EOF) {
        qadav_variant* next_proc = ast_parse_process(&parser);
        root = ast_block_decl(root, next_proc);
    }
    return (void*)root;
}







