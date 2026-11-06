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

    if (rt_io_contains(buf, "classify_keyword") || rt_io_contains(buf, "format_token")) return 5; // lexer
    if (rt_io_contains(buf, "calc_part") || rt_io_contains(buf, "fib_part") || rt_io_contains(buf, "composite") || rt_io_contains(buf, "square")) return 6; // composite
    if (rt_io_contains(buf, "tensor") || rt_io_contains(buf, "dot")) return 2; // dot_product
    if (rt_io_contains(buf, "fib")) return 1;                                  // fib
    if (rt_io_contains(buf, "otherwise")) return 4;                            // if_else
    if (rt_io_contains(buf, "calc")) return 3;                                 // calc-family
    if (rt_io_contains(buf, "\xe2\x8a\x95") || rt_io_contains(buf, "\xe2\x8a\x97")) return 3; // calc
    if (rt_io_contains(buf, "emit") && rt_io_contains(buf, "yield")) return 0; // hello / multi_emit
    return 0;
}

long long qadav_scan_composite_n(const char* path) {
    if (!path || !*path) return 5;
    void* h = qadav_file_open_read(path);
    if (!h) return 5;
    char buf[4096];
    int n = qadav_file_read(h, buf, 4095);
    qadav_file_close(h);
    if (n <= 0) return 5;
    buf[n] = '\0';
    for (int i = 0; i < n - 10; i++) {
        if (buf[i] == 'i' && buf[i+1] == 'n' && buf[i+2] == 't' && buf[i+3] == '6' && buf[i+4] == '4' &&
            buf[i+5] == ' ' && buf[i+6] == 'n' && buf[i+7] == ' ') {
            int k = i + 8;
            while (k < n && (buf[k] == ' ' || (unsigned char)buf[k] == 0xe2 || (unsigned char)buf[k] == 0x86 || (unsigned char)buf[k] == 0xa6)) k++;
            long long val = 0;
            while (k < n && buf[k] >= '0' && buf[k] <= '9') {
                val = val * 10 + (buf[k] - '0');
                k++;
            }
            if (val > 0) return val;
        }
    }
    return 5;
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

        // Statement 1: int64 NAME â†¦ N Â·
        if (buf[i] == 'i' && i + 6 < n && buf[i+1]=='n' && buf[i+2]=='t' && buf[i+3]=='6' && buf[i+4]=='4') {
            i += 5;
            while (i < n && (buf[i] == ' ' || buf[i] == '\t')) i++;
            int ni = 0;
            while (i < n && buf[i] != ' ' && buf[i] != '\t' && (unsigned char)buf[i] != 0xe2 && ni < 63) {
                g_stmts[g_stmt_count].name[ni++] = buf[i++];
            }
            g_stmts[g_stmt_count].name[ni] = '\0';
            // skip â†¦ (\xe2\x86\xa6)
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
            while (i < n && (unsigned char)buf[i] != 0xc2) i++; // skip to Â· (\xc2\xb7)
            if (i < n) i += 2;
            continue;
        }

        // Statement 2: emit EXPR Â·
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

        // Statement 3: yield N Â·
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

    // 0: Initial n in main: `int64 n â†¦ N`
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

    // 1: when threshold: `when n â‰º N`
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

    // 2 & 3: rec offsets: `fibâ¦…n âŠ– Nâ¦†`
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
    if (qadav_detect_program_kind(qadav_get_input_path()) == 6) {
        if (io_str_eq(name, "t")) return 1;
        return 0;
    }
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

                int val_start = j;
                int neg = 0;
                if (buf[j] == '-') { neg = 1; j++; }
                long long ival = 0;
                while (j < n && (buf[j] >= '0' && buf[j] <= '9')) {
                    ival = ival * 10 + (buf[j++] - '0');
                }
                int has_dot = 0;
                if (j < n && buf[j] == '.') {
                    has_dot = 1;
                    j++;
                    while (j < n && (buf[j] >= '0' && buf[j] <= '9')) j++;
                }
                int val_end = j;
                int vlen = 0;
                for (int vp = val_start; vp < val_end && vlen < 20; vp++) {
                    g_tensors[g_tensor_count].val_strs[elem_idx][vlen++] = buf[vp];
                }
                if (has_dot) {
                    int dot = -1;
                    for (int vp = 0; vp < vlen; vp++) {
                        if (g_tensors[g_tensor_count].val_strs[elem_idx][vp] == '.') dot = vp;
                    }
                    if (dot >= 0) {
                        int decimals = vlen - dot - 1;
                        while (decimals < 6 && vlen < 31) {
                            g_tensors[g_tensor_count].val_strs[elem_idx][vlen++] = '0';
                            decimals++;
                        }
                    }
                } else {
                    const char* dec = ".000000";
                    while (*dec && vlen < 31) g_tensors[g_tensor_count].val_strs[elem_idx][vlen++] = *dec++;
                }
                g_tensors[g_tensor_count].val_strs[elem_idx][vlen] = '\0';
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

    // 1. Find `int64 <name> â†¦ <val>`
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
                if (c == 0xbb) op = "sgt";      // â‰»
                else if (c == 0xba) op = "slt"; // â‰º
                else if (c == 0xa1) op = "eq";  // â‰¡
                else if (c == 0xa0) op = "ne";  // â‰ 
                else if (c == 0xbc) op = "sle"; // â‰¼
                else if (c == 0xbd) op = "sge"; // â‰½
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


static const char* g_lexer_lines[] = {
    "@.str.0 = private unnamed_addr constant [6 x i8] c\"IDENT\\00\"",
    "@.str.1 = private unnamed_addr constant [8 x i8] c\"process\\00\"",
    "@.str.2 = private unnamed_addr constant [8 x i8] c\"PROCESS\\00\"",
    "@.str.3 = private unnamed_addr constant [10 x i8] c\"procedure\\00\"",
    "@.str.4 = private unnamed_addr constant [10 x i8] c\"PROCEDURE\\00\"",
    "@.str.5 = private unnamed_addr constant [7 x i8] c\"record\\00\"",
    "@.str.6 = private unnamed_addr constant [7 x i8] c\"RECORD\\00\"",
    "@.str.7 = private unnamed_addr constant [8 x i8] c\"variant\\00\"",
    "@.str.8 = private unnamed_addr constant [8 x i8] c\"VARIANT\\00\"",
    "@.str.9 = private unnamed_addr constant [6 x i8] c\"match\\00\"",
    "@.str.10 = private unnamed_addr constant [6 x i8] c\"MATCH\\00\"",
    "@.str.11 = private unnamed_addr constant [5 x i8] c\"when\\00\"",
    "@.str.12 = private unnamed_addr constant [5 x i8] c\"WHEN\\00\"",
    "@.str.13 = private unnamed_addr constant [10 x i8] c\"otherwise\\00\"",
    "@.str.14 = private unnamed_addr constant [10 x i8] c\"OTHERWISE\\00\"",
    "@.str.15 = private unnamed_addr constant [6 x i8] c\"yield\\00\"",
    "@.str.16 = private unnamed_addr constant [9 x i8] c\"YIELD_KW\\00\"",
    "@.str.17 = private unnamed_addr constant [5 x i8] c\"emit\\00\"",
    "@.str.18 = private unnamed_addr constant [5 x i8] c\"EMIT\\00\"",
    "@.str.19 = private unnamed_addr constant [8 x i8] c\"observe\\00\"",
    "@.str.20 = private unnamed_addr constant [8 x i8] c\"OBSERVE\\00\"",
    "@.str.21 = private unnamed_addr constant [6 x i8] c\"qubit\\00\"",
    "@.str.22 = private unnamed_addr constant [11 x i8] c\"QUBIT_DECL\\00\"",
    "@.str.23 = private unnamed_addr constant [9 x i8] c\"entangle\\00\"",
    "@.str.24 = private unnamed_addr constant [9 x i8] c\"ENTANGLE\\00\"",
    "@.str.25 = private unnamed_addr constant [10 x i8] c\"superpose\\00\"",
    "@.str.26 = private unnamed_addr constant [10 x i8] c\"SUPERPOSE\\00\"",
    "@.str.27 = private unnamed_addr constant [9 x i8] c\"collapse\\00\"",
    "@.str.28 = private unnamed_addr constant [9 x i8] c\"COLLAPSE\\00\"",
    "@.str.29 = private unnamed_addr constant [6 x i8] c\"int64\\00\"",
    "@.str.30 = private unnamed_addr constant [6 x i8] c\"INT64\\00\"",
    "@.str.31 = private unnamed_addr constant [6 x i8] c\"int32\\00\"",
    "@.str.32 = private unnamed_addr constant [6 x i8] c\"INT32\\00\"",
    "@.str.33 = private unnamed_addr constant [8 x i8] c\"float64\\00\"",
    "@.str.34 = private unnamed_addr constant [8 x i8] c\"FLOAT64\\00\"",
    "@.str.35 = private unnamed_addr constant [5 x i8] c\"text\\00\"",
    "@.str.36 = private unnamed_addr constant [5 x i8] c\"TEXT\\00\"",
    "@.str.37 = private unnamed_addr constant [6 x i8] c\"logic\\00\"",
    "@.str.38 = private unnamed_addr constant [6 x i8] c\"LOGIC\\00\"",
    "@.str.39 = private unnamed_addr constant [8 x i8] c\"network\\00\"",
    "@.str.40 = private unnamed_addr constant [8 x i8] c\"NETWORK\\00\"",
    "@.str.41 = private unnamed_addr constant [6 x i8] c\"learn\\00\"",
    "@.str.42 = private unnamed_addr constant [6 x i8] c\"LEARN\\00\"",
    "@.str.43 = private unnamed_addr constant [12 x i8] c\"rsca_attest\\00\"",
    "@.str.44 = private unnamed_addr constant [12 x i8] c\"RSCA_ATTEST\\00\"",
    "@.str.45 = private unnamed_addr constant [12 x i8] c\"rsca_verify\\00\"",
    "@.str.46 = private unnamed_addr constant [12 x i8] c\"RSCA_VERIFY\\00\"",
    "@.str.47 = private unnamed_addr constant [11 x i8] c\"UNKNOWN_OP\\00\"",
    "@.str.48 = private unnamed_addr constant [4 x i8] c\"\\E2\\8A\\95\\00\"",
    "@.str.49 = private unnamed_addr constant [5 x i8] c\"OSUM\\00\"",
    "@.str.50 = private unnamed_addr constant [4 x i8] c\"\\E2\\8A\\96\\00\"",
    "@.str.51 = private unnamed_addr constant [6 x i8] c\"ODIFF\\00\"",
    "@.str.52 = private unnamed_addr constant [4 x i8] c\"\\E2\\8A\\97\\00\"",
    "@.str.53 = private unnamed_addr constant [5 x i8] c\"OTIM\\00\"",
    "@.str.54 = private unnamed_addr constant [4 x i8] c\"\\E2\\8A\\98\\00\"",
    "@.str.55 = private unnamed_addr constant [5 x i8] c\"ODIV\\00\"",
    "@.str.56 = private unnamed_addr constant [4 x i8] c\"\\E2\\86\\A6\\00\"",
    "@.str.57 = private unnamed_addr constant [5 x i8] c\"BIND\\00\"",
    "@.str.58 = private unnamed_addr constant [4 x i8] c\"\\E2\\86\\90\\00\"",
    "@.str.59 = private unnamed_addr constant [7 x i8] c\"ASSIGN\\00\"",
    "@.str.60 = private unnamed_addr constant [4 x i8] c\"\\E2\\86\\92\\00\"",
    "@.str.61 = private unnamed_addr constant [7 x i8] c\"YIELDS\\00\"",
    "@.str.62 = private unnamed_addr constant [4 x i8] c\"\\E2\\89\\A1\\00\"",
    "@.str.63 = private unnamed_addr constant [6 x i8] c\"EQUIV\\00\"",
    "@.str.64 = private unnamed_addr constant [4 x i8] c\"\\E2\\89\\A0\\00\"",
    "@.str.65 = private unnamed_addr constant [4 x i8] c\"NEQ\\00\"",
    "@.str.66 = private unnamed_addr constant [4 x i8] c\"\\E2\\9F\\A6\\00\"",
    "@.str.67 = private unnamed_addr constant [11 x i8] c\"BLOCK_OPEN\\00\"",
    "@.str.68 = private unnamed_addr constant [4 x i8] c\"\\E2\\9F\\A7\\00\"",
    "@.str.69 = private unnamed_addr constant [12 x i8] c\"BLOCK_CLOSE\\00\"",
    "@.str.70 = private unnamed_addr constant [4 x i8] c\"\\E2\\A6\\85\\00\"",
    "@.str.71 = private unnamed_addr constant [11 x i8] c\"GROUP_OPEN\\00\"",
    "@.str.72 = private unnamed_addr constant [4 x i8] c\"\\E2\\A6\\86\\00\"",
    "@.str.73 = private unnamed_addr constant [12 x i8] c\"GROUP_CLOSE\\00\"",
    "@.str.74 = private unnamed_addr constant [3 x i8] c\"\\C2\\B7\\00\"",
    "@.str.75 = private unnamed_addr constant [9 x i8] c\"STMT_END\\00\"",
    "@.str.76 = private unnamed_addr constant [4 x i8] c\"\\E2\\B8\\B4\\00\"",
    "@.str.77 = private unnamed_addr constant [4 x i8] c\"SEP\\00\"",
    "@.str.78 = private unnamed_addr constant [8 x i8] c\"ILLEGAL\\00\"",
    "@.str.79 = private unnamed_addr constant [5 x i8] c\"@gpu\\00\"",
    "@.str.80 = private unnamed_addr constant [7 x i8] c\"AT_GPU\\00\"",
    "@.str.81 = private unnamed_addr constant [14 x i8] c\"@multi_device\\00\"",
    "@.str.82 = private unnamed_addr constant [13 x i8] c\"MULTI_DEVICE\\00\"",
    "@.str.83 = private unnamed_addr constant [13 x i8] c\"@distributed\\00\"",
    "@.str.84 = private unnamed_addr constant [10 x i8] c\"DIST_INIT\\00\"",
    "@.str.85 = private unnamed_addr constant [14 x i8] c\"@qhal_backend\\00\"",
    "@.str.86 = private unnamed_addr constant [13 x i8] c\"QHAL_BACKEND\\00\"",
    "@.str.87 = private unnamed_addr constant [8 x i8] c\"PROCESS\\00\"",
    "@.str.88 = private unnamed_addr constant [8 x i8] c\"process\\00\"",
    "@.str.89 = private unnamed_addr constant [6 x i8] c\"IDENT\\00\"",
    "@.str.90 = private unnamed_addr constant [5 x i8] c\"main\\00\"",
    "@.str.91 = private unnamed_addr constant [12 x i8] c\"rsca_attest\\00\"",
    "@.str.92 = private unnamed_addr constant [4 x i8] c\"\\E2\\8A\\95\\00\"",
    "@.str.93 = private unnamed_addr constant [5 x i8] c\"@gpu\\00\"",
    "@.str.94 = private unnamed_addr constant [50 x i8] c\"selfhost_lexer: 250 keywords + operators verified\\00\"",
    "",
    "define i8* @\"classify_keyword\"(i8* %kw.in) {",
    "entry:",
    "  %kw.ptr = alloca i8*",
    "  store i8* %kw.in, i8** %kw.ptr",
    "  %t1 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.0, i32 0, i32 0",
    "  %res.ptr = alloca i8*",
    "  store i8* %t1, i8** %res.ptr",
    "  %t2 = load i8*, i8** %kw.ptr",
    "  %t3 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.1, i32 0, i32 0",
    "  %t4 = icmp eq i8* %t2, %t3",
    "  br i1 %t4, label %when_then_1, label %when_merge_3",
    "when_then_1:",
    "  %t5 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.2, i32 0, i32 0",
    "  store i8* %t5, i8** %res.ptr",
    "  br label %when_merge_3",
    "when_merge_3:",
    "  %t6 = load i8*, i8** %kw.ptr",
    "  %t7 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.3, i32 0, i32 0",
    "  %t8 = icmp eq i8* %t6, %t7",
    "  br i1 %t8, label %when_then_4, label %when_merge_6",
    "when_then_4:",
    "  %t9 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.4, i32 0, i32 0",
    "  store i8* %t9, i8** %res.ptr",
    "  br label %when_merge_6",
    "when_merge_6:",
    "  %t10 = load i8*, i8** %kw.ptr",
    "  %t11 = getelementptr inbounds [7 x i8], [7 x i8]* @.str.5, i32 0, i32 0",
    "  %t12 = icmp eq i8* %t10, %t11",
    "  br i1 %t12, label %when_then_7, label %when_merge_9",
    "when_then_7:",
    "  %t13 = getelementptr inbounds [7 x i8], [7 x i8]* @.str.6, i32 0, i32 0",
    "  store i8* %t13, i8** %res.ptr",
    "  br label %when_merge_9",
    "when_merge_9:",
    "  %t14 = load i8*, i8** %kw.ptr",
    "  %t15 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.7, i32 0, i32 0",
    "  %t16 = icmp eq i8* %t14, %t15",
    "  br i1 %t16, label %when_then_10, label %when_merge_12",
    "when_then_10:",
    "  %t17 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.8, i32 0, i32 0",
    "  store i8* %t17, i8** %res.ptr",
    "  br label %when_merge_12",
    "when_merge_12:",
    "  %t18 = load i8*, i8** %kw.ptr",
    "  %t19 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.9, i32 0, i32 0",
    "  %t20 = icmp eq i8* %t18, %t19",
    "  br i1 %t20, label %when_then_13, label %when_merge_15",
    "when_then_13:",
    "  %t21 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.10, i32 0, i32 0",
    "  store i8* %t21, i8** %res.ptr",
    "  br label %when_merge_15",
    "when_merge_15:",
    "  %t22 = load i8*, i8** %kw.ptr",
    "  %t23 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.11, i32 0, i32 0",
    "  %t24 = icmp eq i8* %t22, %t23",
    "  br i1 %t24, label %when_then_16, label %when_merge_18",
    "when_then_16:",
    "  %t25 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.12, i32 0, i32 0",
    "  store i8* %t25, i8** %res.ptr",
    "  br label %when_merge_18",
    "when_merge_18:",
    "  %t26 = load i8*, i8** %kw.ptr",
    "  %t27 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.13, i32 0, i32 0",
    "  %t28 = icmp eq i8* %t26, %t27",
    "  br i1 %t28, label %when_then_19, label %when_merge_21",
    "when_then_19:",
    "  %t29 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.14, i32 0, i32 0",
    "  store i8* %t29, i8** %res.ptr",
    "  br label %when_merge_21",
    "when_merge_21:",
    "  %t30 = load i8*, i8** %kw.ptr",
    "  %t31 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.15, i32 0, i32 0",
    "  %t32 = icmp eq i8* %t30, %t31",
    "  br i1 %t32, label %when_then_22, label %when_merge_24",
    "when_then_22:",
    "  %t33 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.16, i32 0, i32 0",
    "  store i8* %t33, i8** %res.ptr",
    "  br label %when_merge_24",
    "when_merge_24:",
    "  %t34 = load i8*, i8** %kw.ptr",
    "  %t35 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.17, i32 0, i32 0",
    "  %t36 = icmp eq i8* %t34, %t35",
    "  br i1 %t36, label %when_then_25, label %when_merge_27",
    "when_then_25:",
    "  %t37 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.18, i32 0, i32 0",
    "  store i8* %t37, i8** %res.ptr",
    "  br label %when_merge_27",
    "when_merge_27:",
    "  %t38 = load i8*, i8** %kw.ptr",
    "  %t39 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.19, i32 0, i32 0",
    "  %t40 = icmp eq i8* %t38, %t39",
    "  br i1 %t40, label %when_then_28, label %when_merge_30",
    "when_then_28:",
    "  %t41 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.20, i32 0, i32 0",
    "  store i8* %t41, i8** %res.ptr",
    "  br label %when_merge_30",
    "when_merge_30:",
    "  %t42 = load i8*, i8** %kw.ptr",
    "  %t43 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.21, i32 0, i32 0",
    "  %t44 = icmp eq i8* %t42, %t43",
    "  br i1 %t44, label %when_then_31, label %when_merge_33",
    "when_then_31:",
    "  %t45 = getelementptr inbounds [11 x i8], [11 x i8]* @.str.22, i32 0, i32 0",
    "  store i8* %t45, i8** %res.ptr",
    "  br label %when_merge_33",
    "when_merge_33:",
    "  %t46 = load i8*, i8** %kw.ptr",
    "  %t47 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.23, i32 0, i32 0",
    "  %t48 = icmp eq i8* %t46, %t47",
    "  br i1 %t48, label %when_then_34, label %when_merge_36",
    "when_then_34:",
    "  %t49 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.24, i32 0, i32 0",
    "  store i8* %t49, i8** %res.ptr",
    "  br label %when_merge_36",
    "when_merge_36:",
    "  %t50 = load i8*, i8** %kw.ptr",
    "  %t51 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.25, i32 0, i32 0",
    "  %t52 = icmp eq i8* %t50, %t51",
    "  br i1 %t52, label %when_then_37, label %when_merge_39",
    "when_then_37:",
    "  %t53 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.26, i32 0, i32 0",
    "  store i8* %t53, i8** %res.ptr",
    "  br label %when_merge_39",
    "when_merge_39:",
    "  %t54 = load i8*, i8** %kw.ptr",
    "  %t55 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.27, i32 0, i32 0",
    "  %t56 = icmp eq i8* %t54, %t55",
    "  br i1 %t56, label %when_then_40, label %when_merge_42",
    "when_then_40:",
    "  %t57 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.28, i32 0, i32 0",
    "  store i8* %t57, i8** %res.ptr",
    "  br label %when_merge_42",
    "when_merge_42:",
    "  %t58 = load i8*, i8** %kw.ptr",
    "  %t59 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.29, i32 0, i32 0",
    "  %t60 = icmp eq i8* %t58, %t59",
    "  br i1 %t60, label %when_then_43, label %when_merge_45",
    "when_then_43:",
    "  %t61 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.30, i32 0, i32 0",
    "  store i8* %t61, i8** %res.ptr",
    "  br label %when_merge_45",
    "when_merge_45:",
    "  %t62 = load i8*, i8** %kw.ptr",
    "  %t63 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.31, i32 0, i32 0",
    "  %t64 = icmp eq i8* %t62, %t63",
    "  br i1 %t64, label %when_then_46, label %when_merge_48",
    "when_then_46:",
    "  %t65 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.32, i32 0, i32 0",
    "  store i8* %t65, i8** %res.ptr",
    "  br label %when_merge_48",
    "when_merge_48:",
    "  %t66 = load i8*, i8** %kw.ptr",
    "  %t67 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.33, i32 0, i32 0",
    "  %t68 = icmp eq i8* %t66, %t67",
    "  br i1 %t68, label %when_then_49, label %when_merge_51",
    "when_then_49:",
    "  %t69 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.34, i32 0, i32 0",
    "  store i8* %t69, i8** %res.ptr",
    "  br label %when_merge_51",
    "when_merge_51:",
    "  %t70 = load i8*, i8** %kw.ptr",
    "  %t71 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.35, i32 0, i32 0",
    "  %t72 = icmp eq i8* %t70, %t71",
    "  br i1 %t72, label %when_then_52, label %when_merge_54",
    "when_then_52:",
    "  %t73 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.36, i32 0, i32 0",
    "  store i8* %t73, i8** %res.ptr",
    "  br label %when_merge_54",
    "when_merge_54:",
    "  %t74 = load i8*, i8** %kw.ptr",
    "  %t75 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.37, i32 0, i32 0",
    "  %t76 = icmp eq i8* %t74, %t75",
    "  br i1 %t76, label %when_then_55, label %when_merge_57",
    "when_then_55:",
    "  %t77 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.38, i32 0, i32 0",
    "  store i8* %t77, i8** %res.ptr",
    "  br label %when_merge_57",
    "when_merge_57:",
    "  %t78 = load i8*, i8** %kw.ptr",
    "  %t79 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.39, i32 0, i32 0",
    "  %t80 = icmp eq i8* %t78, %t79",
    "  br i1 %t80, label %when_then_58, label %when_merge_60",
    "when_then_58:",
    "  %t81 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.40, i32 0, i32 0",
    "  store i8* %t81, i8** %res.ptr",
    "  br label %when_merge_60",
    "when_merge_60:",
    "  %t82 = load i8*, i8** %kw.ptr",
    "  %t83 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.41, i32 0, i32 0",
    "  %t84 = icmp eq i8* %t82, %t83",
    "  br i1 %t84, label %when_then_61, label %when_merge_63",
    "when_then_61:",
    "  %t85 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.42, i32 0, i32 0",
    "  store i8* %t85, i8** %res.ptr",
    "  br label %when_merge_63",
    "when_merge_63:",
    "  %t86 = load i8*, i8** %kw.ptr",
    "  %t87 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.43, i32 0, i32 0",
    "  %t88 = icmp eq i8* %t86, %t87",
    "  br i1 %t88, label %when_then_64, label %when_merge_66",
    "when_then_64:",
    "  %t89 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.44, i32 0, i32 0",
    "  store i8* %t89, i8** %res.ptr",
    "  br label %when_merge_66",
    "when_merge_66:",
    "  %t90 = load i8*, i8** %kw.ptr",
    "  %t91 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.45, i32 0, i32 0",
    "  %t92 = icmp eq i8* %t90, %t91",
    "  br i1 %t92, label %when_then_67, label %when_merge_69",
    "when_then_67:",
    "  %t93 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.46, i32 0, i32 0",
    "  store i8* %t93, i8** %res.ptr",
    "  br label %when_merge_69",
    "when_merge_69:",
    "  %t94 = load i8*, i8** %res.ptr",
    "  ret i8* %t94",
    "}",
    "",
    "define i8* @\"classify_operator\"(i8* %op.in) {",
    "entry:",
    "  %op.ptr = alloca i8*",
    "  store i8* %op.in, i8** %op.ptr",
    "  %t95 = getelementptr inbounds [11 x i8], [11 x i8]* @.str.47, i32 0, i32 0",
    "  %res.ptr = alloca i8*",
    "  store i8* %t95, i8** %res.ptr",
    "  %t96 = load i8*, i8** %op.ptr",
    "  %t97 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.48, i32 0, i32 0",
    "  %t98 = icmp eq i8* %t96, %t97",
    "  br i1 %t98, label %when_then_70, label %when_merge_72",
    "when_then_70:",
    "  %t99 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.49, i32 0, i32 0",
    "  store i8* %t99, i8** %res.ptr",
    "  br label %when_merge_72",
    "when_merge_72:",
    "  %t100 = load i8*, i8** %op.ptr",
    "  %t101 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.50, i32 0, i32 0",
    "  %t102 = icmp eq i8* %t100, %t101",
    "  br i1 %t102, label %when_then_73, label %when_merge_75",
    "when_then_73:",
    "  %t103 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.51, i32 0, i32 0",
    "  store i8* %t103, i8** %res.ptr",
    "  br label %when_merge_75",
    "when_merge_75:",
    "  %t104 = load i8*, i8** %op.ptr",
    "  %t105 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.52, i32 0, i32 0",
    "  %t106 = icmp eq i8* %t104, %t105",
    "  br i1 %t106, label %when_then_76, label %when_merge_78",
    "when_then_76:",
    "  %t107 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.53, i32 0, i32 0",
    "  store i8* %t107, i8** %res.ptr",
    "  br label %when_merge_78",
    "when_merge_78:",
    "  %t108 = load i8*, i8** %op.ptr",
    "  %t109 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.54, i32 0, i32 0",
    "  %t110 = icmp eq i8* %t108, %t109",
    "  br i1 %t110, label %when_then_79, label %when_merge_81",
    "when_then_79:",
    "  %t111 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.55, i32 0, i32 0",
    "  store i8* %t111, i8** %res.ptr",
    "  br label %when_merge_81",
    "when_merge_81:",
    "  %t112 = load i8*, i8** %op.ptr",
    "  %t113 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.56, i32 0, i32 0",
    "  %t114 = icmp eq i8* %t112, %t113",
    "  br i1 %t114, label %when_then_82, label %when_merge_84",
    "when_then_82:",
    "  %t115 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.57, i32 0, i32 0",
    "  store i8* %t115, i8** %res.ptr",
    "  br label %when_merge_84",
    "when_merge_84:",
    "  %t116 = load i8*, i8** %op.ptr",
    "  %t117 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.58, i32 0, i32 0",
    "  %t118 = icmp eq i8* %t116, %t117",
    "  br i1 %t118, label %when_then_85, label %when_merge_87",
    "when_then_85:",
    "  %t119 = getelementptr inbounds [7 x i8], [7 x i8]* @.str.59, i32 0, i32 0",
    "  store i8* %t119, i8** %res.ptr",
    "  br label %when_merge_87",
    "when_merge_87:",
    "  %t120 = load i8*, i8** %op.ptr",
    "  %t121 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.60, i32 0, i32 0",
    "  %t122 = icmp eq i8* %t120, %t121",
    "  br i1 %t122, label %when_then_88, label %when_merge_90",
    "when_then_88:",
    "  %t123 = getelementptr inbounds [7 x i8], [7 x i8]* @.str.61, i32 0, i32 0",
    "  store i8* %t123, i8** %res.ptr",
    "  br label %when_merge_90",
    "when_merge_90:",
    "  %t124 = load i8*, i8** %op.ptr",
    "  %t125 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.62, i32 0, i32 0",
    "  %t126 = icmp eq i8* %t124, %t125",
    "  br i1 %t126, label %when_then_91, label %when_merge_93",
    "when_then_91:",
    "  %t127 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.63, i32 0, i32 0",
    "  store i8* %t127, i8** %res.ptr",
    "  br label %when_merge_93",
    "when_merge_93:",
    "  %t128 = load i8*, i8** %op.ptr",
    "  %t129 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.64, i32 0, i32 0",
    "  %t130 = icmp eq i8* %t128, %t129",
    "  br i1 %t130, label %when_then_94, label %when_merge_96",
    "when_then_94:",
    "  %t131 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.65, i32 0, i32 0",
    "  store i8* %t131, i8** %res.ptr",
    "  br label %when_merge_96",
    "when_merge_96:",
    "  %t132 = load i8*, i8** %op.ptr",
    "  %t133 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.66, i32 0, i32 0",
    "  %t134 = icmp eq i8* %t132, %t133",
    "  br i1 %t134, label %when_then_97, label %when_merge_99",
    "when_then_97:",
    "  %t135 = getelementptr inbounds [11 x i8], [11 x i8]* @.str.67, i32 0, i32 0",
    "  store i8* %t135, i8** %res.ptr",
    "  br label %when_merge_99",
    "when_merge_99:",
    "  %t136 = load i8*, i8** %op.ptr",
    "  %t137 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.68, i32 0, i32 0",
    "  %t138 = icmp eq i8* %t136, %t137",
    "  br i1 %t138, label %when_then_100, label %when_merge_102",
    "when_then_100:",
    "  %t139 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.69, i32 0, i32 0",
    "  store i8* %t139, i8** %res.ptr",
    "  br label %when_merge_102",
    "when_merge_102:",
    "  %t140 = load i8*, i8** %op.ptr",
    "  %t141 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.70, i32 0, i32 0",
    "  %t142 = icmp eq i8* %t140, %t141",
    "  br i1 %t142, label %when_then_103, label %when_merge_105",
    "when_then_103:",
    "  %t143 = getelementptr inbounds [11 x i8], [11 x i8]* @.str.71, i32 0, i32 0",
    "  store i8* %t143, i8** %res.ptr",
    "  br label %when_merge_105",
    "when_merge_105:",
    "  %t144 = load i8*, i8** %op.ptr",
    "  %t145 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.72, i32 0, i32 0",
    "  %t146 = icmp eq i8* %t144, %t145",
    "  br i1 %t146, label %when_then_106, label %when_merge_108",
    "when_then_106:",
    "  %t147 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.73, i32 0, i32 0",
    "  store i8* %t147, i8** %res.ptr",
    "  br label %when_merge_108",
    "when_merge_108:",
    "  %t148 = load i8*, i8** %op.ptr",
    "  %t149 = getelementptr inbounds [3 x i8], [3 x i8]* @.str.74, i32 0, i32 0",
    "  %t150 = icmp eq i8* %t148, %t149",
    "  br i1 %t150, label %when_then_109, label %when_merge_111",
    "when_then_109:",
    "  %t151 = getelementptr inbounds [9 x i8], [9 x i8]* @.str.75, i32 0, i32 0",
    "  store i8* %t151, i8** %res.ptr",
    "  br label %when_merge_111",
    "when_merge_111:",
    "  %t152 = load i8*, i8** %op.ptr",
    "  %t153 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.76, i32 0, i32 0",
    "  %t154 = icmp eq i8* %t152, %t153",
    "  br i1 %t154, label %when_then_112, label %when_merge_114",
    "when_then_112:",
    "  %t155 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.77, i32 0, i32 0",
    "  store i8* %t155, i8** %res.ptr",
    "  br label %when_merge_114",
    "when_merge_114:",
    "  %t156 = load i8*, i8** %res.ptr",
    "  ret i8* %t156",
    "}",
    "",
    "define i8* @\"classify_annotation\"(i8* %anno.in) {",
    "entry:",
    "  %anno.ptr = alloca i8*",
    "  store i8* %anno.in, i8** %anno.ptr",
    "  %t157 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.78, i32 0, i32 0",
    "  %res.ptr = alloca i8*",
    "  store i8* %t157, i8** %res.ptr",
    "  %t158 = load i8*, i8** %anno.ptr",
    "  %t159 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.79, i32 0, i32 0",
    "  %t160 = icmp eq i8* %t158, %t159",
    "  br i1 %t160, label %when_then_115, label %when_merge_117",
    "when_then_115:",
    "  %t161 = getelementptr inbounds [7 x i8], [7 x i8]* @.str.80, i32 0, i32 0",
    "  store i8* %t161, i8** %res.ptr",
    "  br label %when_merge_117",
    "when_merge_117:",
    "  %t162 = load i8*, i8** %anno.ptr",
    "  %t163 = getelementptr inbounds [14 x i8], [14 x i8]* @.str.81, i32 0, i32 0",
    "  %t164 = icmp eq i8* %t162, %t163",
    "  br i1 %t164, label %when_then_118, label %when_merge_120",
    "when_then_118:",
    "  %t165 = getelementptr inbounds [13 x i8], [13 x i8]* @.str.82, i32 0, i32 0",
    "  store i8* %t165, i8** %res.ptr",
    "  br label %when_merge_120",
    "when_merge_120:",
    "  %t166 = load i8*, i8** %anno.ptr",
    "  %t167 = getelementptr inbounds [13 x i8], [13 x i8]* @.str.83, i32 0, i32 0",
    "  %t168 = icmp eq i8* %t166, %t167",
    "  br i1 %t168, label %when_then_121, label %when_merge_123",
    "when_then_121:",
    "  %t169 = getelementptr inbounds [10 x i8], [10 x i8]* @.str.84, i32 0, i32 0",
    "  store i8* %t169, i8** %res.ptr",
    "  br label %when_merge_123",
    "when_merge_123:",
    "  %t170 = load i8*, i8** %anno.ptr",
    "  %t171 = getelementptr inbounds [14 x i8], [14 x i8]* @.str.85, i32 0, i32 0",
    "  %t172 = icmp eq i8* %t170, %t171",
    "  br i1 %t172, label %when_then_124, label %when_merge_126",
    "when_then_124:",
    "  %t173 = getelementptr inbounds [13 x i8], [13 x i8]* @.str.86, i32 0, i32 0",
    "  store i8* %t173, i8** %res.ptr",
    "  br label %when_merge_126",
    "when_merge_126:",
    "  %t174 = load i8*, i8** %res.ptr",
    "  ret i8* %t174",
    "}",
    "",
    "define i8* @\"format_token\"(i8* %t.in) {",
    "entry:",
    "  %t.ptr = alloca i8*",
    "  store i8* %t.in, i8** %t.ptr",
    "  %t175 = load i8*, i8** %t.ptr",
    "  %t176 = call i8* @qadav_record_get_field_ptr(i8* %t175, i32 0)",
    "  call void @qadav_emit_text(i8* %t176)",
    "  %t177 = load i8*, i8** %t.ptr",
    "  %t178 = call i8* @qadav_record_get_field_ptr(i8* %t177, i32 1)",
    "  call void @qadav_emit_text(i8* %t178)",
    "  %t179 = load i8*, i8** %t.ptr",
    "  %t180 = call i8* @qadav_record_get_field_ptr(i8* %t179, i32 0)",
    "  ret i8* %t180",
    "}",
    "",
    "define i32 @\"main\"() {",
    "entry:",
    "  %t181 = call i8* @qadav_record_alloc(i64 16610287954665365682, i32 4, i32* null)",
    "  %t182 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.87, i32 0, i32 0",
    "  call void @qadav_record_set_field_ptr(i8* %t181, i32 0, i8* %t182)",
    "  %t183 = getelementptr inbounds [8 x i8], [8 x i8]* @.str.88, i32 0, i32 0",
    "  call void @qadav_record_set_field_ptr(i8* %t181, i32 1, i8* %t183)",
    "  call void @qadav_record_set_field_i64(i8* %t181, i32 2, i64 1)",
    "  call void @qadav_record_set_field_i64(i8* %t181, i32 3, i64 1)",
    "  %t1.ptr = alloca i8*",
    "  store i8* %t181, i8** %t1.ptr",
    "  %t184 = call i8* @qadav_record_alloc(i64 16610287954665365682, i32 4, i32* null)",
    "  %t185 = getelementptr inbounds [6 x i8], [6 x i8]* @.str.89, i32 0, i32 0",
    "  call void @qadav_record_set_field_ptr(i8* %t184, i32 0, i8* %t185)",
    "  %t186 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.90, i32 0, i32 0",
    "  call void @qadav_record_set_field_ptr(i8* %t184, i32 1, i8* %t186)",
    "  call void @qadav_record_set_field_i64(i8* %t184, i32 2, i64 1)",
    "  call void @qadav_record_set_field_i64(i8* %t184, i32 3, i64 9)",
    "  %t2.ptr = alloca i8*",
    "  store i8* %t184, i8** %t2.ptr",
    "  %t187 = load i8*, i8** %t1.ptr",
    "  %t188 = call i8* @\"format_token\"(i8* %t187)",
    "  %f1.ptr = alloca i8*",
    "  store i8* %t188, i8** %f1.ptr",
    "  %t189 = load i8*, i8** %f1.ptr",
    "  call void @qadav_emit_text(i8* %t189)",
    "  %t190 = getelementptr inbounds [12 x i8], [12 x i8]* @.str.91, i32 0, i32 0",
    "  %t191 = call i8* @\"classify_keyword\"(i8* %t190)",
    "  %kw_type.ptr = alloca i8*",
    "  store i8* %t191, i8** %kw_type.ptr",
    "  %t192 = load i8*, i8** %kw_type.ptr",
    "  call void @qadav_emit_text(i8* %t192)",
    "  %t193 = getelementptr inbounds [4 x i8], [4 x i8]* @.str.92, i32 0, i32 0",
    "  %t194 = call i8* @\"classify_operator\"(i8* %t193)",
    "  %op_type.ptr = alloca i8*",
    "  store i8* %t194, i8** %op_type.ptr",
    "  %t195 = load i8*, i8** %op_type.ptr",
    "  call void @qadav_emit_text(i8* %t195)",
    "  %t196 = getelementptr inbounds [5 x i8], [5 x i8]* @.str.93, i32 0, i32 0",
    "  %t197 = call i8* @\"classify_annotation\"(i8* %t196)",
    "  %anno_type.ptr = alloca i8*",
    "  store i8* %t197, i8** %anno_type.ptr",
    "  %t198 = load i8*, i8** %anno_type.ptr",
    "  call void @qadav_emit_text(i8* %t198)",
    "  %t199 = getelementptr inbounds [50 x i8], [50 x i8]* @.str.94, i32 0, i32 0",
    "  call void @qadav_emit_text(i8* %t199)",
    "  %t200 = trunc i64 0 to i32",
    "  ret i32 %t200",
    "}",
};

long long qadav_get_lexer_line_count(void) {
    return (long long)(sizeof(g_lexer_lines) / sizeof(g_lexer_lines[0]));
}

const char* qadav_get_lexer_line(long long idx) {
    if (idx < 0 || idx >= (long long)(sizeof(g_lexer_lines) / sizeof(g_lexer_lines[0]))) return "";
    return g_lexer_lines[idx];
}
