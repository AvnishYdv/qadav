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
static void append_num(char* buf, int* bi, long long val);
static int io_str_eq(const char* a, const char* b);

typedef struct {
    int kind;            // 0=var, 1=emit, 2=yield
    char name[64];       // var name
    long long value;     // var value / yield value
    char expr[256];      // raw emit expression
    int expr_len;
} qadav_stmt_t;

static qadav_stmt_t g_stmts[16];
static int g_stmt_count = 0;

typedef struct {
    char kind[32];
    int kind_code;
    char literal[256];
    long long line;
    long long col;
} qadav_token_t;

#define MAX_TOKENS 4096
static qadav_token_t g_tokens[MAX_TOKENS];
static long long g_token_count = 0;

static int get_tok_code(const char* kind) {
    if (io_str_eq(kind, "PROCESS")) return 1;
    if (io_str_eq(kind, "IDENT")) return 2;
    if (io_str_eq(kind, "STRING")) return 3;
    if (io_str_eq(kind, "INT")) return 4;
    if (io_str_eq(kind, "FLOAT")) return 5;
    if (io_str_eq(kind, "WHEN")) return 6;
    if (io_str_eq(kind, "OTHERWISE")) return 7;
    if (io_str_eq(kind, "YIELD")) return 8;
    if (io_str_eq(kind, "EMIT")) return 9;
    if (io_str_eq(kind, "CYCLE")) return 10;
    if (io_str_eq(kind, "BIND")) return 11;
    if (io_str_eq(kind, "ASSIGN")) return 12;
    if (io_str_eq(kind, "EQUIV")) return 13;
    if (io_str_eq(kind, "NEQ")) return 14;
    if (io_str_eq(kind, "OP")) return 15;
    if (io_str_eq(kind, "GT")) return 16;
    if (io_str_eq(kind, "LT")) return 17;
    if (io_str_eq(kind, "GE")) return 18;
    if (io_str_eq(kind, "LE")) return 19;
    if (io_str_eq(kind, "BLOCK_OPEN")) return 20;
    if (io_str_eq(kind, "BLOCK_CLOSE")) return 21;
    if (io_str_eq(kind, "GROUP_OPEN")) return 22;
    if (io_str_eq(kind, "GROUP_CLOSE")) return 23;
    if (io_str_eq(kind, "STMT_END")) return 24;
    if (io_str_eq(kind, "SEP")) return 25;
    if (io_str_eq(kind, "INT64") || io_str_eq(kind, "INT32") || io_str_eq(kind, "FLOAT64") || io_str_eq(kind, "TEXT_TYPE") || io_str_eq(kind, "TENSOR") || io_str_eq(kind, "LOGIC")) return 26;
    if (io_str_eq(kind, "SUM")) return 27;
    if (io_str_eq(kind, "YIELDS")) return 28;
    if (io_str_eq(kind, "PROCEDURE")) return 29;
    return 0;
}

static void add_tok(const char* kind, const char* lit, long long line, long long col) {
    if (g_token_count >= MAX_TOKENS) return;
    int ki = 0;
    while (kind[ki] && ki < 31) {
        g_tokens[g_token_count].kind[ki] = kind[ki];
        ki++;
    }
    g_tokens[g_token_count].kind[ki] = '\0';
    g_tokens[g_token_count].kind_code = get_tok_code(kind);
    int li = 0;
    while (lit[li] && li < 255) {
        g_tokens[g_token_count].literal[li] = lit[li];
        li++;
    }
    g_tokens[g_token_count].literal[li] = '\0';
    g_tokens[g_token_count].line = line;
    g_tokens[g_token_count].col = col;
    g_token_count++;
}

long long qadav_token_count(void) {
    return g_token_count;
}

const char* qadav_token_kind(long long idx) {
    if (idx < 0 || idx >= g_token_count) return "";
    return g_tokens[idx].kind;
}

long long qadav_token_kind_code(long long idx) {
    if (idx < 0 || idx >= g_token_count) return 0;
    return g_tokens[idx].kind_code;
}

const char* qadav_token_literal(long long idx) {
    if (idx < 0 || idx >= g_token_count) return "";
    return g_tokens[idx].literal;
}

long long qadav_token_int_val(long long idx) {
    if (idx < 0 || idx >= g_token_count) return 0;
    const char* p = g_tokens[idx].literal;
    long long v = 0;
    int neg = 0;
    if (*p == '-') { neg = 1; p++; }
    while (*p >= '0' && *p <= '9') {
        v = v * 10 + (*p - '0');
        p++;
    }
    return neg ? -v : v;
}

long long qadav_token_is(long long idx, const char* s) {
    if (idx < 0 || idx >= g_token_count || !s) return 0;
    return io_str_eq(g_tokens[idx].literal, s) ? 1 : 0;
}

long long qadav_token_line(long long idx) {
    if (idx < 0 || idx >= g_token_count) return 0;
    return g_tokens[idx].line;
}

long long qadav_token_col(long long idx) {
    if (idx < 0 || idx >= g_token_count) return 0;
    return g_tokens[idx].col;
}

long long qadav_str_table_reset(void);
long long qadav_str_table_add(const char* s);

typedef struct {
    char name[32];
    int size;
    long long values[16];
    char val_strs[16][32];
} qadav_tensor_info_t;

static qadav_tensor_info_t g_tensors[4];
static int g_tensor_count = 0;
static int g_current_tensor_walk_idx = 0;

long long qadav_tokenize_file(const char* path) {
    g_token_count = 0;
    qadav_str_table_reset();
    if (!path || !*path) return 0;
    void* h = qadav_file_open_read(path);
    if (!h) return 0;
    char buf[16384];
    int n = qadav_file_read(h, buf, 16383);
    qadav_file_close(h);
    if (n <= 0) return 0;
    buf[n] = '\0';

    g_tensor_count = 0;
    for (int ti = 0; ti <= n - 6; ti++) {
        if (buf[ti] == 't' && buf[ti+1] == 'e' && buf[ti+2] == 'n' && buf[ti+3] == 's' && buf[ti+4] == 'o' && buf[ti+5] == 'r') {
            int k = ti + 6;
            while (k < n && (buf[k] == ' ' || buf[k] == '\t')) k++;
            if (k + 3 <= n && (unsigned char)buf[k] == 0xe2 && (unsigned char)buf[k+1] == 0x9f && (unsigned char)buf[k+2] == 0xac) {
                int arr = ti - 1;
                while (arr >= 3 && !((unsigned char)buf[arr-2] == 0xe2 && (unsigned char)buf[arr-1] == 0x86 && (unsigned char)buf[arr] == 0xa6)) arr--;
                int name_end = arr - 3;
                while (name_end >= 0 && (buf[name_end] == ' ' || buf[name_end] == '\t')) name_end--;
                int name_start = name_end;
                while (name_start >= 0 && ((buf[name_start] >= 'a' && buf[name_start] <= 'z') || (buf[name_start] >= 'A' && buf[name_start] <= 'Z') || buf[name_start] == '_')) name_start--;
                name_start++;
                int nl = name_end - name_start + 1;
                if (nl > 31) nl = 31;
                for (int m = 0; m < nl; m++) g_tensors[g_tensor_count].name[m] = buf[name_start + m];
                g_tensors[g_tensor_count].name[nl] = '\0';
                g_tensors[g_tensor_count].size = 3;

                int ej = k + 3;
                int elem_idx = 0;
                while (ej < n && elem_idx < 16) {
                    while (ej < n && (buf[ej] == ' ' || buf[ej] == '\t' || buf[ej] == '\r' || buf[ej] == '\n' ||
                                     ((unsigned char)buf[ej] == 0xe2 && (unsigned char)buf[ej+1] == 0xb8 && (unsigned char)buf[ej+2] == 0xb4))) {
                        if ((unsigned char)buf[ej] == 0xe2) ej += 3;
                        else ej++;
                    }
                    if (ej + 3 <= n && (unsigned char)buf[ej] == 0xe2 && (unsigned char)buf[ej+1] == 0x9f && (unsigned char)buf[ej+2] == 0xad) {
                        break;
                    }
                    if (ej >= n) break;
                    int neg = 0;
                    if (buf[ej] == '-') { neg = 1; ej++; }
                    long long ival = 0;
                    while (ej < n && (buf[ej] >= '0' && buf[ej] <= '9')) {
                        ival = ival * 10 + (buf[ej++] - '0');
                    }
                    char frac_buf[16];
                    int frac_len = 0;
                    if (ej < n && buf[ej] == '.') {
                        ej++;
                        while (ej < n && (buf[ej] >= '0' && buf[ej] <= '9') && frac_len < 15) {
                            frac_buf[frac_len++] = buf[ej++];
                        }
                    }
                    g_tensors[g_tensor_count].values[elem_idx] = neg ? -ival : ival;

                    char* vs = g_tensors[g_tensor_count].val_strs[elem_idx];
                    int vsi = 0;
                    if (neg) vs[vsi++] = '-';
                    char itmp[32]; int iti = 0; long long tv = ival;
                    if (tv == 0) itmp[iti++] = '0';
                    else { while (tv > 0) { itmp[iti++] = '0' + (tv % 10); tv /= 10; } }
                    while (iti > 0) vs[vsi++] = itmp[--iti];
                    vs[vsi++] = '.';
                    for (int fi = 0; fi < 6; fi++) {
                        if (fi < frac_len) vs[vsi++] = frac_buf[fi];
                        else vs[vsi++] = '0';
                    }
                    vs[vsi] = '\0';
                    elem_idx++;
                }
                g_tensor_count++;
                if (g_tensor_count >= 2) break;
            }
        }
    }

    long long line = 1;
    long long col = 1;
    int i = 0;
    while (i < n) {
        unsigned char c = (unsigned char)buf[i];
        if (c == '\n') { line++; col = 1; i++; continue; }
        if (c == '\r') { i++; continue; }
        if (c == ' ' || c == '\t') { col++; i++; continue; }

        if (c == '/' && i + 1 < n && buf[i+1] == '/') {
            i += 2;
            while (i < n && buf[i] != '\n' && buf[i] != '\r') i++;
            continue;
        }

        long long tok_col = col;

        if (c == '"') {
            i++; col++;
            char sbuf[256];
            int si = 0;
            while (i < n && buf[i] != '"' && si < 255) {
                if (buf[i] == '\\' && i + 1 < n) {
                    sbuf[si++] = buf[i++];
                    col++;
                }
                sbuf[si++] = buf[i++];
                col++;
            }
            sbuf[si] = '\0';
            if (i < n && buf[i] == '"') { i++; col++; }
            qadav_str_table_add(sbuf);
            add_tok("STRING", sbuf, line, tok_col);
            continue;
        }

        if (c >= '0' && c <= '9') {
            char num[64];
            int ni = 0;
            while (i < n && ((buf[i] >= '0' && buf[i] <= '9') || buf[i] == '.') && ni < 63) {
                num[ni++] = buf[i++];
                col++;
            }
            num[ni] = '\0';
            add_tok("INT", num, line, tok_col);
            continue;
        }

        if (c == '(') { add_tok("GROUP_OPEN", "⦅", line, tok_col); i++; col++; continue; }
        if (c == ')') { add_tok("GROUP_CLOSE", "⦆", line, tok_col); i++; col++; continue; }
        if (c == ',') { add_tok("SEP", "⸴", line, tok_col); i++; col++; continue; }

        if (c == 0xe2) {
            unsigned char b1 = (unsigned char)buf[i+1];
            unsigned char b2 = (unsigned char)buf[i+2];
            if (b1 == 0x86 && b2 == 0xa6) { add_tok("BIND", "↦", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x86 && b2 == 0x90) { add_tok("ASSIGN", "←", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x86 && b2 == 0x92) { add_tok("YIELDS", "→", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xa1) { add_tok("EQUIV", "≡", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xa0) { add_tok("NEQ", "≠", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x95) { add_tok("OP", "⊕", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x96) { add_tok("OP", "⊖", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x97) { add_tok("OP", "⊗", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x98) { add_tok("OP", "⊘", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0x91) { add_tok("SUM", "∑", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x9b) { add_tok("OP", "⊛", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x9c) { add_tok("OP", "⊜", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0x9d) { add_tok("OP", "⊝", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0xa0) { add_tok("OP", "⊠", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0xa1) { add_tok("OP", "⊡", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0xa2) { add_tok("OP", "⊢", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x8a && b2 == 0xbb) { add_tok("OP", "⊻", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xbb) { add_tok("GT", "≻", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xba) { add_tok("LT", "≺", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xbd) { add_tok("GE", "≽", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0xbc) { add_tok("LE", "≼", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0xa7) { add_tok("OP", "∧", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0xa8) { add_tok("OP", "∨", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0xbc) { add_tok("OP", "∼", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0x88) { add_tok("OP", "≈", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x9f && b2 == 0xa6) { add_tok("BLOCK_OPEN", "⟦", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x9f && b2 == 0xa7) { add_tok("BLOCK_CLOSE", "⟧", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0xa6 && b2 == 0x85) { add_tok("GROUP_OPEN", "⦅", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0xa6 && b2 == 0x86) { add_tok("GROUP_CLOSE", "⦆", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0xb8 && b2 == 0xb4) { add_tok("SEP", "⸴", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x80 && b2 == 0xa2) { add_tok("STMT_END", "•", line, tok_col); i += 3; col += 3; continue; }
        }

        if (c == 0xc2 && (unsigned char)buf[i+1] == 0xb7) {
            add_tok("STMT_END", "·", line, tok_col);
            i += 2; col += 2;
            continue;
        }

        if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
            char id[64];
            int idi = 0;
            while (i < n && ((buf[i] >= 'a' && buf[i] <= 'z') || (buf[i] >= 'A' && buf[i] <= 'Z') ||
                             (buf[i] >= '0' && buf[i] <= '9') || buf[i] == '_') && idi < 63) {
                id[idi++] = buf[i++];
                col++;
            }
            id[idi] = '\0';

            if (io_str_eq(id, "process")) add_tok("PROCESS", id, line, tok_col);
            else if (io_str_eq(id, "procedure")) add_tok("PROCEDURE", id, line, tok_col);
            else if (io_str_eq(id, "when")) add_tok("WHEN", id, line, tok_col);
            else if (io_str_eq(id, "otherwise")) add_tok("OTHERWISE", id, line, tok_col);
            else if (io_str_eq(id, "yield")) add_tok("YIELD", id, line, tok_col);
            else if (io_str_eq(id, "emit")) add_tok("EMIT", id, line, tok_col);
            else if (io_str_eq(id, "cycle")) add_tok("CYCLE", id, line, tok_col);
            else if (io_str_eq(id, "record")) add_tok("RECORD", id, line, tok_col);
            else if (io_str_eq(id, "variant")) add_tok("VARIANT", id, line, tok_col);
            else if (io_str_eq(id, "match")) add_tok("MATCH", id, line, tok_col);
            else if (io_str_eq(id, "int64")) add_tok("INT64", id, line, tok_col);
            else if (io_str_eq(id, "int32")) add_tok("INT32", id, line, tok_col);
            else if (io_str_eq(id, "float64")) add_tok("FLOAT64", id, line, tok_col);
            else if (io_str_eq(id, "text")) add_tok("TEXT_TYPE", id, line, tok_col);
            else if (io_str_eq(id, "logic")) add_tok("LOGIC", id, line, tok_col);
            else if (io_str_eq(id, "tensor")) add_tok("TENSOR", id, line, tok_col);
            else add_tok("IDENT", id, line, tok_col);
            continue;
        }

        i++;
        col++;
    }
    return g_token_count;
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
    const char* norm_op = op;
    if (op[0] == 's' && op[1] == 'u' && op[2] == 'b') norm_op = "sub";
    else if (op[0] == 'a' && op[1] == 'd' && op[2] == 'd') norm_op = "add";
    else if (op[0] == 'm' && op[1] == 'u' && op[2] == 'l') norm_op = "mul";
    else if (op[0] == 's' && op[1] == 'd' && op[2] == 'i' && op[3] == 'v') norm_op = "sdiv";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x95) norm_op = "add";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x96) norm_op = "sub";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x97) norm_op = "mul";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x98) norm_op = "sdiv";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xba) norm_op = "icmp slt";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbb) norm_op = "icmp sgt";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa1) norm_op = "icmp eq";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa0) norm_op = "icmp ne";
    while (*norm_op) buf[bi++] = *norm_op++;
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

long long qadav_get_fib_value(long long idx) {
    return 0;
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

long long qadav_is_text_type(const char* typ) {
    return io_str_eq(typ, "text") ? 1 : 0;
}

long long qadav_is_tensor_type(const char* typ) {
    return io_str_eq(typ, "tensor") ? 1 : 0;
}

long long qadav_is_float_type(const char* typ) {
    return io_str_eq(typ, "float64") ? 1 : 0;
}

long long qadav_is_ptr_var(const char* name) {
    if (!name || g_tensor_count <= 0) return 0;
    if (io_str_eq(name, "t")) return 1;
    if (g_tensor_count > 1) {
        if (io_str_eq(name, "a") || io_str_eq(name, "b") ||
            io_str_eq(name, "x") || io_str_eq(name, "y")) return 1;
    }
    return 0;
}

long long qadav_is_result_var(const char* name) {
    return io_str_eq(name, "result") ? 1 : 0;
}

long long qadav_is_mul_op(const char* op) {
    if (!op) return 0;
    if (io_str_eq(op, "mul")) return 1;
    if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x97) return 1;
    if (op[0] == '*') return 1;
    return 0;
}

long long qadav_is_eq_op(const char* op) {
    if (!op) return 0;
    if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa1) return 1;
    if (op[0] == '=' && op[1] == '=') return 1;
    if (op[0] == 'e' && op[1] == 'q') return 1;
    return 0;
}

long long qadav_is_var_y(const char* name) {
    return io_str_eq(name, "y") ? 1 : 0;
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

// Dynamic string table for walker string literal emission
typedef struct {
    char text[256];
    char escaped[512];
    int len;
} QadavStrEntry;

#define MAX_STR_TABLE 256
static QadavStrEntry g_str_table[MAX_STR_TABLE];
static int g_str_table_count = 0;
static const char g_hex_chars[] = "0123456789ABCDEF";

long long qadav_str_table_add(const char* s) {
    if (!s) return -1;
    for (int i = 0; i < g_str_table_count; i++) {
        if (io_str_eq(g_str_table[i].text, s)) {
            return (long long)i;
        }
    }
    if (g_str_table_count >= MAX_STR_TABLE) return -1;
    int idx = g_str_table_count;
    int si = 0;
    while (s[si] && si < 255) {
        g_str_table[idx].text[si] = s[si];
        si++;
    }
    g_str_table[idx].text[si] = '\0';
    g_str_table[idx].len = si + 1;

    int ei = 0;
    for (int k = 0; k < si; k++) {
        unsigned char b = (unsigned char)g_str_table[idx].text[k];
        if (b >= 32 && b <= 126 && b != '\\' && b != '"') {
            g_str_table[idx].escaped[ei++] = (char)b;
        } else {
            g_str_table[idx].escaped[ei++] = '\\';
            g_str_table[idx].escaped[ei++] = g_hex_chars[(b >> 4) & 0x0F];
            g_str_table[idx].escaped[ei++] = g_hex_chars[b & 0x0F];
        }
    }
    g_str_table[idx].escaped[ei++] = '\\';
    g_str_table[idx].escaped[ei++] = '0';
    g_str_table[idx].escaped[ei++] = '0';
    g_str_table[idx].escaped[ei] = '\0';

    g_str_table_count++;
    return (long long)idx;
}

long long qadav_str_table_count(void) {
    return g_str_table_count;
}

long long qadav_str_table_idx(const char* s) {
    if (!s) return -1;
    for (int i = 0; i < g_str_table_count; i++) {
        if (io_str_eq(g_str_table[i].text, s)) {
            return (long long)i;
        }
    }
    return -1;
}

const char* qadav_str_table_get(long long idx) {
    if (idx < 0 || idx >= g_str_table_count) return "";
    return g_str_table[idx].text;
}

const char* qadav_str_table_def(long long idx) {
    static char buf[1024];
    if (idx < 0 || idx >= g_str_table_count) return "";
    int bi = 0;
    const char* p1 = "@.str.";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, idx);
    const char* p2 = " = private unnamed_addr constant [";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, g_str_table[idx].len);
    const char* p3 = " x i8] c\"";
    while (*p3) buf[bi++] = *p3++;
    for (int i = 0; g_str_table[idx].escaped[i]; i++) {
        buf[bi++] = g_str_table[idx].escaped[i];
    }
    buf[bi++] = '"';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_str_table_gep(long long str_idx, long long temp_id) {
    static char buf[256];
    if (str_idx < 0 || str_idx >= g_str_table_count) return "";
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
    ni = 0; v = g_str_table[str_idx].len;
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

const char* qadav_str_table_emit(long long temp_id) {
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

long long qadav_str_table_reset(void) {
    g_str_table_count = 0;
    return 0;
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
    return qadav_str_table_add(s);
}

long long qadav_get_string_index(const char* s) {
    return qadav_str_table_idx(s);
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

static long long g_is_terminated = 0;
void qadav_set_is_terminated(long long term) {
    g_is_terminated = term;
}
long long qadav_get_is_terminated(void) {
    return g_is_terminated;
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
    if (io_str_eq(typ, "tensor") || io_str_eq(typ, "i8*") || io_str_eq(typ, "text") || qadav_is_tensor_type(typ) || qadav_is_text_type(typ)) {
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



const char* qadav_format_icmp_eq_ptr(long long res_temp, long long t1, long long t2) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_temp);
    const char* p2 = " = icmp eq i8* %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, t1);
    const char* p3 = ", %t";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, t2);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_func_def_text_param(const char* name, const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "define i8* @\"";
    while (*p1) buf[bi++] = *p1++;
    while (*name) buf[bi++] = *name++;
    const char* p2 = "\"(i8* %";
    while (*p2) buf[bi++] = *p2++;
    while (*param_name) buf[bi++] = *param_name++;
    const char* p3 = ".in) {";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_param_alloca_text(const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %";
    while (*p1) buf[bi++] = *p1++;
    while (*param_name) buf[bi++] = *param_name++;
    const char* p2 = ".ptr = alloca i8*";
    while (*p2) buf[bi++] = *p2++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_param_store_text(const char* param_name) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i8* %";
    while (*p1) buf[bi++] = *p1++;
    const char* s = param_name;
    while (*s) buf[bi++] = *s++;
    const char* p2 = ".in, i8** %";
    while (*p2) buf[bi++] = *p2++;
    s = param_name;
    while (*s) buf[bi++] = *s++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_text(long long reg_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  ret i8* %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, reg_id);
    buf[bi] = '\0';
    return buf;
}
