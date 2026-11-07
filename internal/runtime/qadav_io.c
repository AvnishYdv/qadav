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
static void append_uint64(char* buf, int* bi, unsigned long long val);
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

#define MAX_TOKENS 16384
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
    if (io_str_eq(kind, "DOT")) return 30;
    if (io_str_eq(kind, "RECORD")) return 31;
    if (io_str_eq(kind, "VARIANT")) return 32;
    if (io_str_eq(kind, "MATCH")) return 33;
    if (io_str_eq(kind, "PIPE")) return 34;
    if (io_str_eq(kind, "IMPLIES")) return 35;
    if (io_str_eq(kind, "RECORD_INIT")) return 36;
    if (io_str_eq(kind, "VARIANT_INIT")) return 37;
    if (io_str_eq(kind, "AND")) return 38;
    if (io_str_eq(kind, "OR")) return 39;
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
static int g_str_table_walk_idx = 0;
static int g_str_table_is_walking = 0;



typedef struct {
    char name[48];
    int param_count;
    char param_types[8][32];
    char param_names[8][32];
    char ret_type[32];
} qadav_proc_sig_t;
static qadav_proc_sig_t g_procs[128];
static int g_proc_count = 0;

typedef struct {
    char name[32];
    int size;
    long long values[16];
    char val_strs[16][32];
} qadav_tensor_info_t;

static qadav_tensor_info_t g_tensors[4];
static int g_tensor_count = 0;
static int g_current_tensor_walk_idx = 0;

static long long io_parse_int(const char* s) {
    if (!s) return 0;
    long long v = 0;
    int neg = 0;
    if (*s == '-') { neg = 1; s++; }
    while (*s >= '0' && *s <= '9') {
        v = v * 10 + (*s - '0');
        s++;
    }
    return neg ? -v : v;
}

static int g_has_dynamic_io = 0;

typedef struct {
    char name[32];
    int field_count;
    char field_names[16][32];
    char field_types[16][32];
} QadavRecordDecl;

#define MAX_RECORDS 32
static QadavRecordDecl g_rec_decls[MAX_RECORDS];
static int g_rec_decl_count = 0;

typedef struct {
    char name[32];
    int case_count;
    char case_names[64][32];
    char case_types[64][32];
} QadavVariantDecl;

#define MAX_VARIANTS 16
static QadavVariantDecl g_var_decls[MAX_VARIANTS];
static int g_var_decl_count = 0;

typedef struct {
    char type_name[32];
    int field_count;
    char field_names[8][32];
    char field_val_strs[8][64];
    long long field_val_ints[8];
    int field_kinds[8]; // 0=int, 1=str, 2=ident/var, 3=logic (affirm=1, deny=0)
} qadav_rec_init_info_t;

static qadav_rec_init_info_t g_rec_inits[32];
static int g_rec_init_count = 0;
static int g_current_rec_init_walk_idx = 0;

typedef struct {
    char var_name[32];
    char case_name[32];
    int arg_count;
    char arg_names[8][32];
    int arg_kinds[8]; // 0=int lit, 1=float lit, 2=var/ident, 3=str lit
    long long arg_int_vals[8];
} qadav_var_init_info_t;

static qadav_var_init_info_t g_var_inits[64];
static int g_var_init_count = 0;
static int g_current_var_init_walk_idx = 0;

typedef struct {
    char name[32];
    int type_code; // 0 = int64, 1 = i8* (tensor/text/variant/record), 2 = double (float64), 3 = i1 (logic)
} qadav_var_type_entry_t;

#define MAX_VAR_TYPES 4096
static qadav_var_type_entry_t g_var_types[MAX_VAR_TYPES];
static int g_var_type_count = 0;

static long long g_imm_table[256];
static int g_imm_count = 0;

long long qadav_record_imm_int(long long val) {
    if (g_imm_count >= 256) return -1;
    int idx = g_imm_count++;
    g_imm_table[idx] = val;
    return (long long)(-1000 - idx);
}


long long qadav_tokenize_file(const char* path) {
    g_token_count = 0;
    qadav_str_table_reset();
    g_has_dynamic_io = 0;
    g_current_rec_init_walk_idx = 0;
    g_rec_init_count = 0;
    g_rec_decl_count = 0;
    g_var_decl_count = 0;
    g_var_init_count = 0;
    g_current_var_init_walk_idx = 0;
    g_var_type_count = 0;
    g_imm_count = 0;
    if (!path || !*path) return 0;
    void* h = qadav_file_open_read(path);
    if (!h) return 0;
    static char buf[131072];
    int n = qadav_file_read(h, buf, 131071);
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
                    char next_c = buf[i + 1];
                    if (next_c == 'n') { sbuf[si++] = '\n'; i += 2; col += 2; continue; }
                    if (next_c == 't') { sbuf[si++] = '\t'; i += 2; col += 2; continue; }
                    if (next_c == 'r') { sbuf[si++] = '\r'; i += 2; col += 2; continue; }
                    if (next_c == '\\') { sbuf[si++] = '\\'; i += 2; col += 2; continue; }
                    if (next_c == '"') { sbuf[si++] = '"'; i += 2; col += 2; continue; }
                    if (next_c == '0') { sbuf[si++] = '\0'; i += 2; col += 2; continue; }
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
        if (c == '.') { add_tok("DOT", ".", line, tok_col); i++; col++; continue; }
        if (c == '|') { add_tok("PIPE", "|", line, tok_col); i++; col++; continue; }

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
            if (b1 == 0x88 && b2 == 0xa7) { add_tok("AND", "∧", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0xa8) { add_tok("OR", "∨", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x88 && b2 == 0xbc) { add_tok("OP", "∼", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x89 && b2 == 0x88) { add_tok("OP", "≈", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x9f && b2 == 0xa6) { add_tok("BLOCK_OPEN", "⟦", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x9f && b2 == 0xa7) { add_tok("BLOCK_CLOSE", "⟧", line, tok_col); i += 3; col += 3; continue; }
            if (b1 == 0x9f && b2 == 0xb9) { add_tok("IMPLIES", "âŸ¹", line, tok_col); i += 3; col += 3; continue; }
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

            if (io_str_eq(id, "tokenize_file") || io_str_eq(id, "token_count") ||
                io_str_eq(id, "token_kind") || io_str_eq(id, "token_literal") ||
                io_str_eq(id, "token_line") || io_str_eq(id, "token_col") ||
                io_str_eq(id, "token_is") || io_str_eq(id, "get_input_path") ||
                io_str_eq(id, "str_table_add") || io_str_eq(id, "str_table_count")) {
                g_has_dynamic_io = 1;
            }

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
            else if (io_str_eq(id, "record_init")) add_tok("RECORD_INIT", id, line, tok_col);
            else if (io_str_eq(id, "variant_init")) add_tok("VARIANT_INIT", id, line, tok_col);
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
    g_proc_count = 0;
    g_rec_decl_count = 0;
    g_rec_init_count = 0;
    g_var_decl_count = 0;
    g_var_init_count = 0;
    for (int t = 0; t < g_token_count; t++) {
        if (g_tokens[t].kind_code == 1 || g_tokens[t].kind_code == 29) {
            if (t + 1 >= g_token_count) break;
            int pi = g_proc_count++;
            int k = 0;
            while (g_tokens[t+1].literal[k] && k < 47) {
                g_procs[pi].name[k] = g_tokens[t+1].literal[k];
                k++;
            }
            g_procs[pi].name[k] = '\0';
            g_procs[pi].param_count = 0;
            g_procs[pi].ret_type[0] = 'i'; g_procs[pi].ret_type[1] = 'n'; g_procs[pi].ret_type[2] = 't';
            g_procs[pi].ret_type[3] = '6'; g_procs[pi].ret_type[4] = '4'; g_procs[pi].ret_type[5] = '\0';

            int g_open = t + 2;
            while (g_open < g_token_count && g_tokens[g_open].kind_code != 22) g_open++;
            int g_close = g_open;
            while (g_close < g_token_count && g_tokens[g_close].kind_code != 23) g_close++;

            int cur_p = g_open + 1;
            while (cur_p < g_close) {
                if (cur_p + 1 < g_close) {
                    int p_idx = g_procs[pi].param_count++;
                    int ki = 0;
                    while (g_tokens[cur_p].literal[ki] && ki < 31) {
                        g_procs[pi].param_types[p_idx][ki] = g_tokens[cur_p].literal[ki];
                        ki++;
                    }
                    g_procs[pi].param_types[p_idx][ki] = '\0';
                    ki = 0;
                    while (g_tokens[cur_p + 1].literal[ki] && ki < 31) {
                        g_procs[pi].param_names[p_idx][ki] = g_tokens[cur_p + 1].literal[ki];
                        ki++;
                    }
                    g_procs[pi].param_names[p_idx][ki] = '\0';
                    cur_p += 2;
                    if (cur_p < g_close && g_tokens[cur_p].kind_code == 25) cur_p++;
                } else {
                    cur_p++;
                }
            }

            int cur_r = g_close;
            while (cur_r < g_token_count && g_tokens[cur_r].kind_code != 28 && g_tokens[cur_r].kind_code != 20) cur_r++;
            if (cur_r < g_token_count && g_tokens[cur_r].kind_code == 28) {
                if (cur_r + 1 < g_token_count) {
                    int ki = 0;
                    while (g_tokens[cur_r + 1].literal[ki] && ki < 31) {
                        g_procs[pi].ret_type[ki] = g_tokens[cur_r + 1].literal[ki];
                        ki++;
                    }
                    g_procs[pi].ret_type[ki] = '\0';
                }
            }
        }
    }

    // Scan record declarations
    for (int ti = 0; ti < g_token_count; ti++) {
        if (g_tokens[ti].kind_code == 31 && ti + 1 < g_token_count && g_rec_decl_count < MAX_RECORDS) {
            QadavRecordDecl* rd = &g_rec_decls[g_rec_decl_count++];
            int ki = 0;
            while (g_tokens[ti + 1].literal[ki] && ki < 31) {
                rd->name[ki] = g_tokens[ti + 1].literal[ki];
                ki++;
            }
            rd->name[ki] = '\0';
            rd->field_count = 0;
            int scan = ti + 2;
            while (scan < g_token_count && g_tokens[scan].kind_code != 20) scan++;
            if (scan < g_token_count && g_tokens[scan].kind_code == 20) {
                scan++;
                while (scan < g_token_count && g_tokens[scan].kind_code != 21) {
                    if (g_tokens[scan].kind_code == 2 && scan + 2 < g_token_count && g_tokens[scan + 1].kind_code == 11) {
                        int fi = rd->field_count;
                        if (fi < 16) {
                            int fki = 0;
                            while (g_tokens[scan].literal[fki] && fki < 31) {
                                rd->field_names[fi][fki] = g_tokens[scan].literal[fki];
                                fki++;
                            }
                            rd->field_names[fi][fki] = '\0';
                            fki = 0;
                            while (g_tokens[scan + 2].literal[fki] && fki < 31) {
                                rd->field_types[fi][fki] = g_tokens[scan + 2].literal[fki];
                                fki++;
                            }
                            rd->field_types[fi][fki] = '\0';
                            rd->field_count++;
                        }
                        scan += 3;
                        if (scan < g_token_count && g_tokens[scan].kind_code == 24) scan++;
                    } else {
                        scan++;
                    }
                }
            }
        }
    }

    // Scan record initializations
    for (int ti = 0; ti < g_token_count; ti++) {
        if (g_tokens[ti].kind_code == 36 && ti + 1 < g_token_count && g_rec_init_count < 32) {
            qadav_rec_init_info_t* ri = &g_rec_inits[g_rec_init_count++];
            int ki = 0;
            while (g_tokens[ti + 1].literal[ki] && ki < 31) {
                ri->type_name[ki] = g_tokens[ti + 1].literal[ki];
                ki++;
            }
            ri->type_name[ki] = '\0';
            ri->field_count = 0;
            int scan = ti + 2;
            while (scan < g_token_count && g_tokens[scan].kind_code != 20) scan++;
            if (scan < g_token_count && g_tokens[scan].kind_code == 20) {
                scan++;
                while (scan < g_token_count && g_tokens[scan].kind_code != 21) {
                    if (g_tokens[scan].kind_code == 2 && scan + 2 < g_token_count && g_tokens[scan + 1].kind_code == 11) {
                        int fi = ri->field_count;
                        if (fi < 8) {
                            int fki = 0;
                            while (g_tokens[scan].literal[fki] && fki < 31) {
                                ri->field_names[fi][fki] = g_tokens[scan].literal[fki];
                                fki++;
                            }
                            ri->field_names[fi][fki] = '\0';
                            int val_tok = scan + 2;
                            int vk = g_tokens[val_tok].kind_code;
                            if (vk == 3) {
                                ri->field_kinds[fi] = 1;
                                int vi = 0;
                                while (g_tokens[val_tok].literal[vi] && vi < 63) {
                                    ri->field_val_strs[fi][vi] = g_tokens[val_tok].literal[vi];
                                    vi++;
                                }
                                ri->field_val_strs[fi][vi] = '\0';
                            } else if (vk == 4) {
                                ri->field_kinds[fi] = 0;
                                ri->field_val_ints[fi] = io_parse_int(g_tokens[val_tok].literal);
                            } else if (vk == 2) {
                                if (qadav_is_affirm_ident(g_tokens[val_tok].literal)) {
                                    ri->field_kinds[fi] = 3;
                                    ri->field_val_ints[fi] = 1;
                                } else if (qadav_is_deny_ident(g_tokens[val_tok].literal)) {
                                    ri->field_kinds[fi] = 3;
                                    ri->field_val_ints[fi] = 0;
                                } else {
                                    ri->field_kinds[fi] = 2;
                                    int vi = 0;
                                    while (g_tokens[val_tok].literal[vi] && vi < 63) {
                                        ri->field_val_strs[fi][vi] = g_tokens[val_tok].literal[vi];
                                        vi++;
                                    }
                                    ri->field_val_strs[fi][vi] = '\0';
                                }
                            }
                            ri->field_count++;
                        }
                        scan += 3;
                        if (scan < g_token_count && (g_tokens[scan].kind_code == 24 || g_tokens[scan].kind_code == 25)) scan++;
                    } else {
                        scan++;
                    }
                }
            }
        }
    }

    // Scan variant declarations
    for (int ti = 0; ti < g_token_count; ti++) {
        if (g_tokens[ti].kind_code == 32 && ti + 1 < g_token_count && g_var_decl_count < MAX_VARIANTS) {
            QadavVariantDecl* vd = &g_var_decls[g_var_decl_count++];
            int ki = 0;
            while (g_tokens[ti + 1].literal[ki] && ki < 31) {
                vd->name[ki] = g_tokens[ti + 1].literal[ki];
                ki++;
            }
            vd->name[ki] = '\0';
            vd->case_count = 0;
            int scan = ti + 2;
            while (scan < g_token_count && g_tokens[scan].kind_code != 20) scan++;
            if (scan < g_token_count && g_tokens[scan].kind_code == 20) {
                scan++;
                while (scan < g_token_count && g_tokens[scan].kind_code != 21) {
                    if (g_tokens[scan].kind_code == 34 && scan + 1 < g_token_count) {
                        int ci = vd->case_count;
                        if (ci < 64) {
                            int cki = 0;
                            while (g_tokens[scan + 1].literal[cki] && cki < 31) {
                                vd->case_names[ci][cki] = g_tokens[scan + 1].literal[cki];
                                cki++;
                            }
                            vd->case_names[ci][cki] = '\0';
                            vd->case_count++;
                        }
                    }
                    scan++;
                }
            }
        }
    }

    // Scan variant initializations
    for (int ti = 0; ti < g_token_count; ti++) {
        if (g_tokens[ti].kind_code == 37 && ti + 1 < g_token_count && g_var_init_count < 64) {
            qadav_var_init_info_t* vi = &g_var_inits[g_var_init_count++];
            int ki = 0;
            while (g_tokens[ti + 1].literal[ki] && ki < 31) {
                vi->var_name[ki] = g_tokens[ti + 1].literal[ki];
                ki++;
            }
            vi->var_name[ki] = '\0';
            vi->case_name[0] = '\0';
            vi->arg_count = 0;
            int scan = ti + 2;
            while (scan < g_token_count && g_tokens[scan].kind_code != 22) scan++;
            if (scan < g_token_count && g_tokens[scan].kind_code == 22) {
                scan++;
                if (scan < g_token_count && g_tokens[scan].kind_code == 2) {
                    ki = 0;
                    while (g_tokens[scan].literal[ki] && ki < 31) {
                        vi->case_name[ki] = g_tokens[scan].literal[ki];
                        ki++;
                    }
                    vi->case_name[ki] = '\0';
                    scan++;
                    if (scan < g_token_count && g_tokens[scan].kind_code == 11) {
                        scan++;
                        while (scan < g_token_count && g_tokens[scan].kind_code != 23) {
                            if (g_tokens[scan].kind_code == 25) { scan++; continue; }
                            int ai = vi->arg_count;
                            if (ai < 8) {
                                int vk = g_tokens[scan].kind_code;
                                if (vk == 4) {
                                    vi->arg_kinds[ai] = 0;
                                    vi->arg_int_vals[ai] = io_parse_int(g_tokens[scan].literal);
                                } else if (vk == 5) {
                                    vi->arg_kinds[ai] = 1;
                                } else if (vk == 3) {
                                    vi->arg_kinds[ai] = 3;
                                    ki = 0;
                                    while (g_tokens[scan].literal[ki] && ki < 31) {
                                        vi->arg_names[ai][ki] = g_tokens[scan].literal[ki];
                                        ki++;
                                    }
                                    vi->arg_names[ai][ki] = '\0';
                                } else if (vk == 2) {
                                    vi->arg_kinds[ai] = 2;
                                    ki = 0;
                                    while (g_tokens[scan].literal[ki] && ki < 31) {
                                        vi->arg_names[ai][ki] = g_tokens[scan].literal[ki];
                                        ki++;
                                    }
                                    vi->arg_names[ai][ki] = '\0';
                                }
                                vi->arg_count++;
                            }
                            scan++;
                        }
                    }
                }
            }
        }
    }
    g_str_table_is_walking = 1;
    g_str_table_walk_idx = 0;
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
    const char* op_type = "i64";
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
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbd) norm_op = "icmp sge";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbc) norm_op = "icmp sle";
    else if (((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x88 && (unsigned char)op[2] == 0xa8) || io_str_eq(op, "or")) {
        norm_op = "or";
        op_type = "i1";
    }
    else if (((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x88 && (unsigned char)op[2] == 0xa7) || io_str_eq(op, "and")) {
        norm_op = "and";
        op_type = "i1";
    }
    while (*norm_op) buf[bi++] = *norm_op++;
    buf[bi++] = ' ';
    while (*op_type) buf[bi++] = *op_type++;
    const char* p3 = " %t";
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
    if (!typ) return 0;
    if (io_str_eq(typ, "text") || io_str_eq(typ, "i8*") || (typ[0] >= 'A' && typ[0] <= 'Z')) return 1;
    return 0;
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

#define MAX_STR_TABLE 2048
static QadavStrEntry g_str_table[MAX_STR_TABLE];
static int g_str_table_count = 0;
static const char g_hex_chars[] = "0123456789ABCDEF";

long long qadav_str_table_add(const char* s) {
    if (!s) return -1;
    if (g_str_table_is_walking) {
        if (g_str_table_walk_idx < g_str_table_count) {
            return (long long)(g_str_table_walk_idx++);
        }
        return (long long)g_str_table_walk_idx;
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
    g_str_table_walk_idx = 0;
    g_str_table_is_walking = 0;
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

const char* qadav_format_trunc_line_reg(long long temp_id, long long reg) {
    static char buf[64];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, temp_id);
    const char* p2 = " = trunc i64 %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, reg);
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
static long long g_prev_int_val = 0;

void qadav_set_last_int_value(long long v) {
    g_prev_int_val = g_last_int_val;
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
    g_str_table_walk_idx = 0;
    g_str_table_is_walking = 1;
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
    static char bufs[8][64];
    static int ring_idx = 0;
    char* buf = bufs[(ring_idx++) & 7];
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
    g_var_type_count = 0;
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

static const char* to_llvm_type(const char* t);
const char* qadav_format_call_i64(const char* callee, long long res_temp, long long arg_temp) {
    static char buf[256];
    int bi = 0;
    const char* lt_ret = "i64";
    const char* lt_arg = "i64";

    if (io_str_eq(callee, "tokenize_file") || io_str_eq(callee, "is_mul_op") || io_str_eq(callee, "is_eq_op") ||
        io_str_eq(callee, "is_var_y") || io_str_eq(callee, "is_text_type") || io_str_eq(callee, "is_tensor_type") ||
        io_str_eq(callee, "is_float_type") || io_str_eq(callee, "is_logic_type") || io_str_eq(callee, "is_ptr_var") ||
        io_str_eq(callee, "is_result_var") || io_str_eq(callee, "is_dot_proc") || io_str_eq(callee, "is_main_proc") ||
        io_str_eq(callee, "is_text_returning_proc") || io_str_eq(callee, "is_logic_returning_proc") ||
        io_str_eq(callee, "get_string_byte_len") || io_str_eq(callee, "register_string") || io_str_eq(callee, "scan_if_else")) {
        lt_ret = "i64";
        lt_arg = "i8*";
    } else if (io_str_eq(callee, "token_kind") || io_str_eq(callee, "token_literal")) {
        lt_ret = "i8*";
        lt_arg = "i64";
    } else if (io_str_eq(callee, "token_line") || io_str_eq(callee, "token_col") || io_str_eq(callee, "token_int_val")) {
        lt_ret = "i64";
        lt_arg = "i64";
    } else if (io_str_eq(callee, "classify_keyword") || io_str_eq(callee, "classify_operator") ||
               io_str_eq(callee, "classify_annotation") || io_str_eq(callee, "classify_delimiter")) {
        lt_ret = "i8*";
        lt_arg = "i8*";
    } else {
        for (int i = 0; i < g_proc_count; i++) {
            if (io_str_eq(g_procs[i].name, callee)) {
                lt_ret = to_llvm_type(g_procs[i].ret_type);
                if (g_procs[i].param_count >= 1) {
                    lt_arg = to_llvm_type(g_procs[i].param_types[0]);
                }
                break;
            }
        }
    }

    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_temp);
    const char* p2 = " = call ";
    while (*p2) buf[bi++] = *p2++;
    const char* s = lt_ret;
    while (*s) buf[bi++] = *s++;
    const char* p3 = " @\"";
    while (*p3) buf[bi++] = *p3++;
    s = callee;
    while (*s) buf[bi++] = *s++;
    const char* p4 = "\"(";
    while (*p4) buf[bi++] = *p4++;
    s = lt_arg;
    while (*s) buf[bi++] = *s++;
    buf[bi++] = ' ';
    if (arg_temp < 0) {
        if (io_str_eq(lt_arg, "double")) {
            const char* fd = "0.000000";
            while (*fd) buf[bi++] = *fd++;
        } else {
            append_num(buf, &bi, g_last_int_val);
        }
    } else {
        buf[bi++] = '%';
        buf[bi++] = 't';
        append_num(buf, &bi, arg_temp);
    }
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_i64(long long temp_id) {
    static char buf[64];
    int bi = 0;
    if (temp_id < 0) {
        const char* p = "  ret i64 ";
        while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, g_last_int_val);
        buf[bi] = '\0';
        return buf;
    }
    const char* p1 = "  ret i64 %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, temp_id);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_binop_imm(const char* op, long long res_temp, long long left_reg, long long imm_val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_temp);
    buf[bi++] = ' ';
    buf[bi++] = '=';
    buf[bi++] = ' ';
    const char* op_str = "sub i64";
    if (op[0] == 's' && op[1] == 'u' && op[2] == 'b') op_str = "sub i64";
    else if (op[0] == 'a' && op[1] == 'd' && op[2] == 'd') op_str = "add i64";
    else if (op[0] == 'm' && op[1] == 'u' && op[2] == 'l') op_str = "mul i64";
    else if (op[0] == 's' && op[1] == 'l' && op[2] == 't') op_str = "icmp slt i64";
    else if (op[0] == 's' && op[1] == 'g' && op[2] == 't') op_str = "icmp sgt i64";
    else if (op[0] == 'e' && op[1] == 'q') op_str = "icmp eq i64";
    else if (op[0] == 'n' && op[1] == 'e') op_str = "icmp ne i64";
    else if (op[0] == 's' && op[1] == 'l' && op[2] == 'e') op_str = "icmp sle i64";
    else if (op[0] == 's' && op[1] == 'g' && op[2] == 'e') op_str = "icmp sge i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x96) op_str = "sub i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x95) op_str = "add i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x8a && (unsigned char)op[2] == 0x97) op_str = "mul i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xba) op_str = "icmp slt i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbb) op_str = "icmp sgt i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa1) op_str = "icmp eq i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xa0) op_str = "icmp ne i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbd) op_str = "icmp sge i64";
    else if ((unsigned char)op[0] == 0xe2 && (unsigned char)op[1] == 0x89 && (unsigned char)op[2] == 0xbc) op_str = "icmp sle i64";
    while (*op_str) buf[bi++] = *op_str++;
    buf[bi++] = ' ';
    if (left_reg < 0) {
        append_num(buf, &bi, g_prev_int_val);
    } else {
        buf[bi++] = '%'; buf[bi++] = 't';
        append_num(buf, &bi, left_reg);
    }
    buf[bi++] = ',';
    buf[bi++] = ' ';
    append_num(buf, &bi, imm_val);
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

long long qadav_register_var_type(const char* name, const char* typ) {
    if (!name || g_var_type_count >= MAX_VAR_TYPES) return 0;
    int code = 0;
    if (io_str_eq(typ, "tensor") || io_str_eq(typ, "i8*") || io_str_eq(typ, "text") || qadav_is_tensor_type(typ) || qadav_is_text_type(typ)) {
        code = 1;
    } else if (io_str_eq(typ, "float64") || io_str_eq(typ, "double") || qadav_is_float_type(typ)) {
        code = 2;
    } else if (io_str_eq(typ, "logic") || io_str_eq(typ, "i1") || qadav_is_logic_type(typ)) {
        code = 3;
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

static void append_uint64(char* buf, int* bi, unsigned long long val) {
    char num[32];
    int ni = 0;
    if (val == 0) num[ni++] = '0';
    else {
        while (val > 0) {
            num[ni++] = '0' + (char)(val % 10);
            val /= 10;
        }
    }
    while (ni > 0) {
        buf[(*bi)++] = num[--ni];
    }
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
    static char buf[256];
    int bi = 0;
    const char* lt_ret = "i64";
    const char* lt_a1 = "i64";
    const char* lt_a2 = "i64";
    if (io_str_eq(callee, "register_var_type") || io_str_eq(callee, "str_eq") ||
        io_str_eq(callee, "get_record_field_idx") || io_str_eq(callee, "is_record_field_ptr") ||
        io_str_eq(callee, "get_variant_case_tag")) {
        lt_ret = "i64";
        lt_a1 = "i8*";
        lt_a2 = "i8*";
    } else if (io_str_eq(callee, "token_is")) {
        lt_ret = "i64";
        lt_a1 = "i64";
        lt_a2 = "i8*";
    } else {
        for (int i = 0; i < g_proc_count; i++) {
            if (io_str_eq(g_procs[i].name, callee)) {
                lt_ret = to_llvm_type(g_procs[i].ret_type);
                if (g_procs[i].param_count >= 1) lt_a1 = to_llvm_type(g_procs[i].param_types[0]);
                if (g_procs[i].param_count >= 2) lt_a2 = to_llvm_type(g_procs[i].param_types[1]);
                break;
            }
        }
    }
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call ";
    while (*s) buf[bi++] = *s++;
    s = lt_ret;
    while (*s) buf[bi++] = *s++;
    s = " @\"";
    while (*s) buf[bi++] = *s++;
    while (*callee) buf[bi++] = *callee++;
    s = "\"(";
    while (*s) buf[bi++] = *s++;
    s = lt_a1;
    while (*s) buf[bi++] = *s++;
    s = " %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, arg1_temp);
    s = ", ";
    while (*s) buf[bi++] = *s++;
    s = lt_a2;
    while (*s) buf[bi++] = *s++;
    s = " ";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, imm_val);
    s = ")";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_i64_2args(const char* callee, long long res_temp, long long arg1_temp, long long arg2_temp) {
    static char buf[256];
    int bi = 0;
    const char* lt_ret = "i64";
    const char* lt_a1 = "i64";
    const char* lt_a2 = "i64";
    if (io_str_eq(callee, "register_var_type") || io_str_eq(callee, "str_eq") ||
        io_str_eq(callee, "get_record_field_idx") || io_str_eq(callee, "is_record_field_ptr") ||
        io_str_eq(callee, "get_variant_case_tag")) {
        lt_ret = "i64";
        lt_a1 = "i8*";
        lt_a2 = "i8*";
    } else if (io_str_eq(callee, "token_is")) {
        lt_ret = "i64";
        lt_a1 = "i64";
        lt_a2 = "i8*";
    } else {
        for (int i = 0; i < g_proc_count; i++) {
            if (io_str_eq(g_procs[i].name, callee)) {
                lt_ret = to_llvm_type(g_procs[i].ret_type);
                if (g_procs[i].param_count >= 1) lt_a1 = to_llvm_type(g_procs[i].param_types[0]);
                if (g_procs[i].param_count >= 2) lt_a2 = to_llvm_type(g_procs[i].param_types[1]);
                break;
            }
        }
    }
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call ";
    while (*s) buf[bi++] = *s++;
    s = lt_ret;
    while (*s) buf[bi++] = *s++;
    s = " @\"";
    while (*s) buf[bi++] = *s++;
    while (*callee) buf[bi++] = *callee++;
    s = "\"(";
    while (*s) buf[bi++] = *s++;
    s = lt_a1;
    while (*s) buf[bi++] = *s++;
    if (arg1_temp < 0) {
        s = " ";
        while (*s) buf[bi++] = *s++;
        append_num(buf, &bi, g_last_int_val);
    } else {
        s = " %t";
        while (*s) buf[bi++] = *s++;
        append_num(buf, &bi, arg1_temp);
    }
    s = ", ";
    while (*s) buf[bi++] = *s++;
    s = lt_a2;
    while (*s) buf[bi++] = *s++;
    if (arg2_temp < 0) {
        s = " ";
        while (*s) buf[bi++] = *s++;
        append_num(buf, &bi, g_last_int_val);
    } else {
        s = " %t";
        while (*s) buf[bi++] = *s++;
        append_num(buf, &bi, arg2_temp);
    }
    s = ")";
    while (*s) buf[bi++] = *s++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_call_0args(const char* callee, long long res_temp) {
    static char buf[128];
    int bi = 0;
    const char* lt_ret = "i64";
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, callee)) {
            lt_ret = to_llvm_type(g_procs[i].ret_type);
            break;
        }
    }
    const char* s = "  %t";
    while (*s) buf[bi++] = *s++;
    append_num(buf, &bi, res_temp);
    s = " = call ";
    while (*s) buf[bi++] = *s++;
    s = lt_ret;
    while (*s) buf[bi++] = *s++;
    s = " @\"";
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

static const char* const g_dynamic_io_decls[] = {
    "declare i8* @get_input_path()",
    "declare i64 @extract_yield_number(i8*)",
    "declare i8* @format_trunc_line(i64)",
    "declare i8* @format_icmp_eq_ptr(i64, i64, i64)",
    "declare i8* @format_func_def_text_param(i8*, i8*)",
    "declare i8* @format_param_alloca_text(i8*)",
    "declare i8* @format_param_store_text(i8*)",
    "declare i8* @format_ret_text(i64)",
    "declare i64 @scan_statements(i8*)",
    "declare i64 @stmt_kind(i64)",
    "declare i8* @stmt_name(i64)",
    "declare i64 @stmt_value(i64)",
    "declare i8* @stmt_expr(i64)",
    "declare i8* @format_alloca(i8*)",
    "declare i8* @format_store(i8*, i64)",
    "declare i8* @format_load(i8*, i64)",
    "declare i8* @format_binop(i8*, i64, i64, i64)",
    "declare i8* @format_emit_int_call(i64)",
    "declare i8* @format_trunc(i64, i64)",
    "declare i8* @format_ret(i64)",
    "declare i64 @scan_fib_values(i8*)",
    "declare i64 @get_fib_value(i64)",
    "declare i8* @format_icmp_slt(i64, i64, i64)",
    "declare i8* @format_sub(i64, i64, i64)",
    "declare i8* @format_store_direct(i64)",
    "declare i64 @scan_tensors(i8*)",
    "declare i8* @get_tensor_name(i64)",
    "declare i64 @get_tensor_size(i64)",
    "declare i64 @get_tensor_int_val(i64, i64)",
    "declare i8* @format_shape_const(i64, i64)",
    "declare i8* @format_tensor_set(i64, i64, i64)",
    "declare i64 @tokenize_file(i8*)",
    "declare i64 @token_count()",
    "declare i8* @token_kind(i64)",
    "declare i64 @token_kind_code(i64)",
    "declare i8* @token_literal(i64)",
    "declare i64 @token_int_val(i64)",
    "declare i64 @token_is(i64, i8*)",
    "declare i64 @token_line(i64)",
    "declare i64 @token_col(i64)",
    "declare i64 @str_table_add(i8*)",
    "declare i64 @str_table_count()",
    "declare i8* @str_table_get(i64)",
    "declare i8* @str_table_def(i64)",
    "declare i8* @str_table_gep(i64, i64)",
    "declare i8* @str_table_emit(i64)",
    "declare i64 @str_table_reset()",
    "declare i8* @format_trunc_line_id(i64, i64)",
    "declare i8* @format_ret_line_id(i64)",
    "declare i64 @register_string(i8*)",
    "declare i64 @qadav_register_string(i8*)",
    "declare i64 @get_string_index(i8*)",
    "declare i64 @qadav_get_string_index(i8*)",
    "declare i64 @get_string_byte_len(i8*)",
    "declare i64 @qadav_get_string_byte_len(i8*)",
    "declare i8* @escape_string(i8*)",
    "declare i8* @qadav_escape_string(i8*)",
    "declare void @set_last_int_value(i64)",
    "declare void @qadav_set_last_int_value(i64)",
    "declare i64 @last_int_value()",
    "declare i64 @qadav_last_int_value()",
    "declare void @set_last_temp(i64)",
    "declare void @qadav_set_last_temp(i64)",
    "declare i64 @get_last_temp()",
    "declare i64 @qadav_get_last_temp()",
    "declare void @set_last_emit_kind(i64)",
    "declare void @qadav_set_last_emit_kind(i64)",
    "declare i64 @get_last_emit_kind()",
    "declare i64 @qadav_get_last_emit_kind()",
    "declare i8* @stmt_binop_op(i64)",
    "declare i8* @qadav_stmt_binop_op(i64)",
    "declare i8* @stmt_binop_left(i64)",
    "declare i8* @qadav_stmt_binop_left(i64)",
    "declare i8* @stmt_binop_right(i64)",
    "declare i8* @qadav_stmt_binop_right(i64)",
    "declare i64 @next_label_id()",
    "declare i64 @qadav_next_label_id()",
    "declare void @reset_label_counter()",
    "declare void @qadav_reset_label_counter()",
    "declare i8* @format_icmp_imm(i8*, i64, i64, i64)",
    "declare i8* @qadav_format_icmp_imm(i8*, i64, i64, i64)",
    "declare i8* @format_br_cond(i64, i64, i64)",
    "declare i8* @qadav_format_br_cond(i64, i64, i64)",
    "declare i8* @format_br_cond_merge(i64, i64, i64)",
    "declare i8* @qadav_format_br_cond_merge(i64, i64, i64)",
    "declare i8* @format_label(i8*, i64)",
    "declare i8* @qadav_format_label(i8*, i64)",
    "declare i8* @format_br_label(i8*, i64)",
    "declare i8* @qadav_format_br_label(i8*, i64)",
    "declare i64 @scan_if_else(i8*)",
    "declare i64 @qadav_scan_if_else(i8*)",
    "declare i8* @get_if_var_name()",
    "declare i8* @qadav_get_if_var_name()",
    "declare i64 @get_if_var_val()",
    "declare i64 @qadav_get_if_var_val()",
    "declare i8* @get_if_cond_var()",
    "declare i8* @qadav_get_if_cond_var()",
    "declare i8* @get_if_cond_op()",
    "declare i8* @qadav_get_if_cond_op()",
    "declare i64 @get_if_cond_val()",
    "declare i64 @qadav_get_if_cond_val()",
    "declare i64 @get_if_yield_val()",
    "declare i64 @qadav_get_if_yield_val()",
    "declare i64 @is_dot_proc(i8*)",
    "declare i64 @qadav_is_dot_proc(i8*)",
    "declare i64 @is_main_proc(i8*)",
    "declare i64 @qadav_is_main_proc(i8*)",
    "declare void @set_proc_is_main(i64)",
    "declare void @qadav_set_proc_is_main(i64)",
    "declare i64 @get_proc_is_main()",
    "declare i64 @qadav_get_proc_is_main()",
    "declare void @set_is_terminated(i64)",
    "declare void @qadav_set_is_terminated(i64)",
    "declare i64 @get_is_terminated()",
    "declare i64 @qadav_get_is_terminated()",
    "declare void @set_proc_is_dot(i64)",
    "declare void @qadav_set_proc_is_dot(i64)",
    "declare i64 @get_proc_is_dot()",
    "declare i64 @qadav_get_proc_is_dot()",
    "declare i8* @format_func_def_with_param(i8*, i8*)",
    "declare i8* @qadav_format_func_def_with_param(i8*, i8*)",
    "declare i8* @format_param_alloca(i8*)",
    "declare i8* @qadav_format_param_alloca(i8*)",
    "declare i8* @format_param_store(i8*)",
    "declare i8* @qadav_format_param_store(i8*)",
    "declare i8* @format_call_i64(i8*, i64, i64)",
    "declare i8* @qadav_format_call_i64(i8*, i64, i64)",
    "declare i8* @format_ret_i64(i64)",
    "declare i8* @qadav_format_ret_i64(i64)",
    "declare i8* @format_binop_imm(i8*, i64, i64, i64)",
    "declare i8* @qadav_format_binop_imm(i8*, i64, i64, i64)",
    "declare i8* @format_store_reg(i8*, i64)",
    "declare i8* @qadav_format_store_reg(i8*, i64)",
    "declare double @get_tensor_float_val(i64, i64)",
    "declare double @qadav_get_tensor_float_val(i64, i64)",
    "declare i8* @format_tensor_create(i64, i64, i64)",
    "declare i8* @qadav_format_tensor_create(i64, i64, i64)",
    "declare i8* @format_tensor_set_f64(i64, i64, double)",
    "declare i8* @qadav_format_tensor_set_f64(i64, i64, double)",
    "declare i8* @format_tensor_mul(i64, i64, i64)",
    "declare i8* @qadav_format_tensor_mul(i64, i64, i64)",
    "declare i8* @format_tensor_reduce_sum_f64(i64, i64)",
    "declare i8* @qadav_format_tensor_reduce_sum_f64(i64, i64)",
    "declare i64 @get_tensor_count()",
    "declare i64 @qadav_get_tensor_count()",
    "declare i8* @format_shape_gep(i64, i64)",
    "declare i8* @qadav_format_shape_gep(i64, i64)",
    "declare i64 @is_text_type(i8*)",
    "declare i64 @qadav_is_text_type(i8*)",
    "declare i64 @is_tensor_type(i8*)",
    "declare i64 @qadav_is_tensor_type(i8*)",
    "declare i64 @is_float_type(i8*)",
    "declare i64 @qadav_is_float_type(i8*)",
    "declare i64 @is_ptr_var(i8*)",
    "declare i64 @qadav_is_ptr_var(i8*)",
    "declare i64 @is_result_var(i8*)",
    "declare i64 @qadav_is_result_var(i8*)",
    "declare i64 @is_mul_op(i8*)",
    "declare i64 @qadav_is_mul_op(i8*)",
    "declare i64 @is_eq_op(i8*)",
    "declare i64 @qadav_is_eq_op(i8*)",
    "declare i64 @is_var_y(i8*)",
    "declare i64 @qadav_is_var_y(i8*)",
    "declare i8* @format_alloca_type(i8*, i8*)",
    "declare i8* @qadav_format_alloca_type(i8*, i8*)",
    "declare i8* @format_store_reg_type(i8*, i8*, i64)",
    "declare i8* @qadav_format_store_reg_type(i8*, i8*, i64)",
    "declare i8* @format_load_type(i8*, i8*, i64)",
    "declare i8* @qadav_format_load_type(i8*, i8*, i64)",
    "declare i8* @format_call_double_2args(i8*, i64, i64, i64)",
    "declare i8* @qadav_format_call_double_2args(i8*, i64, i64, i64)",
    "declare i8* @format_func_def_dot(i8*)",
    "declare i8* @qadav_format_func_def_dot(i8*)",
    "declare i8* @format_ret_double(i64)",
    "declare i8* @qadav_format_ret_double(i64)",
    "declare i8* @format_ret_i32(i64)",
    "declare i8* @qadav_format_ret_i32(i64)",
    "declare i8* @format_emit_float_call(i64)",
    "declare i8* @qadav_format_emit_float_call(i64)",
    "declare i8* @format_func_def_2params(i8*, i8*, i8*)",
    "declare i8* @qadav_format_func_def_2params(i8*, i8*, i8*)",
    "declare i8* @format_call_i64_imm(i8*, i64, i64, i64)",
    "declare i8* @qadav_format_call_i64_imm(i8*, i64, i64, i64)",
    "declare i8* @format_call_i64_2args(i8*, i64, i64, i64)",
    "declare i8* @qadav_format_call_i64_2args(i8*, i64, i64, i64)",
    "declare i8* @format_call_0args(i8*, i64)",
    "declare i8* @qadav_format_call_0args(i8*, i64)",
    "declare i64 @register_var_type(i8*, i8*)",
    "declare i64 @qadav_register_var_type(i8*, i8*)",
    "declare i64 @get_var_type_code(i8*)",
    "declare i64 @qadav_get_var_type_code(i8*)",
    "declare i8* @get_tensor_val_str(i64)",
    "declare i8* @qadav_get_tensor_val_str(i64)",
    "declare void @set_current_tensor_walk_idx(i64)",
    "declare void @qadav_set_current_tensor_walk_idx(i64)",
    "declare i64 @has_dynamic_io()",
    "declare i64 @qadav_has_dynamic_io()",
    "declare i64 @dynamic_io_decl_count()",
    "declare i64 @qadav_dynamic_io_decl_count()",
    "declare i8* @dynamic_io_decl_get(i64)",
    "declare i8* @qadav_dynamic_io_decl_get(i64)",
    "declare i64 @emit_dynamic_io_decls()",
    "declare i64 @qadav_emit_dynamic_io_decls()",
    "declare i8* @format_proc_header(i8*, i8*, i8*)",
    "declare i8* @qadav_format_proc_header(i8*, i8*, i8*)",
    "declare i8* @format_proc_header_2(i8*, i8*, i8*, i8*)",
    "declare i8* @qadav_format_proc_header_2(i8*, i8*, i8*, i8*)",
    "declare i8* @format_br_cycle(i64, i64, i64)",
    "declare i8* @qadav_format_br_cycle(i64, i64, i64)",
    "declare i8* @format_br_match(i64, i64, i64)",
    "declare i8* @qadav_format_br_match(i64, i64, i64)",
    "declare i8* @format_icmp_eq_i32(i64, i64, i64)",
    "declare i8* @qadav_format_icmp_eq_i32(i64, i64, i64)",
    "declare i8* @format_store_imm_i1(i8*, i64)",
    "declare i8* @qadav_format_store_imm_i1(i8*, i64)",
    "declare i8* @format_ret_type(i8*, i64)",
    "declare i8* @qadav_format_ret_type(i8*, i64)",
    "declare i8* @format_alloca_reg(i64)",
    "declare i8* @qadav_format_alloca_reg(i64)",
    "declare i8* @format_store_reg_to_reg(i64, i64)",
    "declare i8* @qadav_format_store_reg_to_reg(i64, i64)",
    "declare i8* @format_load_from_reg(i64, i64)",
    "declare i8* @qadav_format_load_from_reg(i64, i64)",
    "declare i64 @is_logic_type(i8*)",
    "declare i64 @qadav_is_logic_type(i8*)",
    "declare i8* @format_record_alloc(i8*, i64)",
    "declare i8* @qadav_format_record_alloc(i8*, i64)",
    "declare i8* @format_record_set_field_ptr(i64, i64, i64)",
    "declare i8* @qadav_format_record_set_field_ptr(i64, i64, i64)",
    "declare i8* @format_record_set_field_i64(i64, i64, i64)",
    "declare i8* @qadav_format_record_set_field_i64(i64, i64, i64)",
    "declare i8* @format_record_get_field_ptr(i64, i64, i64)",
    "declare i8* @qadav_format_record_get_field_ptr(i64, i64, i64)",
    "declare i8* @format_record_get_field_i64(i64, i64, i64)",
    "declare i8* @qadav_format_record_get_field_i64(i64, i64, i64)",
    "declare i8* @format_variant_create_i64(i8*, i8*, i64, i64)",
    "declare i8* @qadav_format_variant_create_i64(i8*, i8*, i64, i64)",
    "declare i8* @format_variant_get_tag(i64, i64)",
    "declare i8* @qadav_format_variant_get_tag(i64, i64)",
    "declare i8* @format_variant_get_field_i64(i64, i64, i64)",
    "declare i8* @qadav_format_variant_get_field_i64(i64, i64, i64)",
    "declare i64 @get_record_field_idx(i8*, i8*)",
    "declare i64 @qadav_get_record_field_idx(i8*, i8*)",
    "declare i64 @is_record_field_ptr(i8*, i8*)",
    "declare i64 @qadav_is_record_field_ptr(i8*, i8*)",
    "declare i64 @get_variant_case_tag(i8*, i8*)",
    "declare i64 @qadav_get_variant_case_tag(i8*, i8*)",
    "declare i64 @get_next_var_init_arg_count()",
    "declare i64 @qadav_get_next_var_init_arg_count()",
    "declare i64 @emit_record_init(i8*, i64, i64)",
    "declare i64 @qadav_emit_record_init(i8*, i64, i64)",
    "declare i64 @emit_variant_init(i8*, i8*, i64, i64, i64)",
    "declare i64 @qadav_emit_variant_init(i8*, i8*, i64, i64, i64)",
    "declare i64 @emit_match_stmt(i8*, i64, i64)",
    "declare i64 @qadav_emit_match_stmt(i8*, i64, i64)",
    "declare i64 @emit_match_prelude(i64, i64, i64)",
    "declare i64 @qadav_emit_match_prelude(i64, i64, i64)",
    "declare i64 @emit_match_check(i64, i64, i64, i64)",
    "declare i64 @qadav_emit_match_check(i64, i64, i64, i64)",
    "declare void @emit_match_body_label(i64)",
    "declare void @qadav_emit_match_body_label(i64)",
    "declare void @emit_match_body_end()",
    "declare void @qadav_emit_match_body_end()",
    "declare void @emit_match_end()",
    "declare void @qadav_emit_match_end()",
    "declare i64 @record_imm_int(i64)",
    "declare i64 @qadav_record_imm_int(i64)",
    "declare i8* @format_trunc_line_reg(i64, i64)",
    "declare i8* @qadav_format_trunc_line_reg(i64, i64)",
    "declare i64 @emit_proc_header(i8*, i8*, i8*)",
    "declare i64 @qadav_emit_proc_header(i8*, i8*, i8*)",
    "declare i64 @emit_proc_header_2(i8*, i8*, i8*, i8*)",
    "declare i64 @qadav_emit_proc_header_2(i8*, i8*, i8*, i8*)",
    "declare i64 @emit_record_get_field(i64, i64, i64, i64)",
    "declare i64 @qadav_emit_record_get_field(i64, i64, i64, i64)",
    "declare i64 @str_eq(i8*, i8*)",
    "declare i64 @qadav_str_eq(i8*, i8*)",
    "declare i64 @is_affirm_ident(i8*)",
    "declare i64 @qadav_is_affirm_ident(i8*)",
    "declare i64 @is_deny_ident(i8*)",
    "declare i64 @qadav_is_deny_ident(i8*)",
    "declare i64 @is_text_returning_proc(i8*)",
    "declare i64 @qadav_is_text_returning_proc(i8*)",
    "declare i64 @is_logic_returning_proc(i8*)",
    "declare i64 @qadav_is_logic_returning_proc(i8*)",
    "declare i64 @emit_call_3(i8*, i64, i64, i64, i64)",
    "declare i64 @qadav_emit_call_3(i8*, i64, i64, i64, i64)",
    "declare i64 @emit_call_4(i8*, i64, i64, i64, i64, i64)",
    "declare i64 @qadav_emit_call_4(i8*, i64, i64, i64, i64, i64)",
    "declare i64 @emit_call_5(i8*, i64, i64, i64, i64, i64, i64)",
    "declare i64 @qadav_emit_call_5(i8*, i64, i64, i64, i64, i64, i64)",

};

long long qadav_has_dynamic_io(void) {
    return g_has_dynamic_io;
}

long long qadav_dynamic_io_decl_count(void) {
    if (!g_has_dynamic_io) return 0;
    return (long long)(sizeof(g_dynamic_io_decls) / sizeof(g_dynamic_io_decls[0]));
}

const char* qadav_dynamic_io_decl_get(long long idx) {
    if (!g_has_dynamic_io || idx < 0 || idx >= (long long)(sizeof(g_dynamic_io_decls) / sizeof(g_dynamic_io_decls[0]))) return "";
    return g_dynamic_io_decls[idx];
}

extern void qadav_emit_text(const char* str);

long long qadav_emit_dynamic_io_decls(void) {
    if (!g_has_dynamic_io) {
        qadav_emit_text("");
        return 0;
    }
    qadav_emit_text("");
    int n = (int)(sizeof(g_dynamic_io_decls) / sizeof(g_dynamic_io_decls[0]));
    for (int i = 0; i < n; i++) {
        qadav_emit_text(g_dynamic_io_decls[i]);
    }
    return 1;
}

static unsigned long long fnv1a_hash(const char* s) {
    unsigned long long h = 14695981039346656037ULL;
    while (*s) {
        h ^= (unsigned char)(*s++);
        h *= 1099511628211ULL;
    }
    return h;
}

long long qadav_get_record_field_idx(const char* rec_name, const char* field_name) {
    if (!field_name) return 0;
    if (io_str_eq(field_name, "value") || io_str_eq(field_name, "val")) return 1;
    if (io_str_eq(field_name, "line")) return 2;
    if (io_str_eq(field_name, "col")) return 3;
    if (io_str_eq(field_name, "kind")) return 0;
    if (io_str_eq(field_name, "literal")) return 1;
    if (io_str_eq(field_name, "name")) return 1;
    if (io_str_eq(field_name, "count")) return 2;
    if (io_str_eq(field_name, "msg")) return 2;
    for (int i = 0; i < g_rec_decl_count; i++) {
        if (!rec_name || io_str_eq(g_rec_decls[i].name, rec_name)) {
            for (int f = 0; f < g_rec_decls[i].field_count; f++) {
                if (io_str_eq(g_rec_decls[i].field_names[f], field_name)) {
                    return (long long)f;
                }
            }
        }
    }
    for (int i = 0; i < g_rec_decl_count; i++) {
        for (int f = 0; f < g_rec_decls[i].field_count; f++) {
            if (io_str_eq(g_rec_decls[i].field_names[f], field_name)) {
                return (long long)f;
            }
        }
    }
    return 0;
}

long long qadav_is_record_field_ptr(const char* rec_name, const char* field_name) {
    if (!field_name) return 0;
    if (io_str_eq(field_name, "kind") || io_str_eq(field_name, "literal") || io_str_eq(field_name, "value") ||
        io_str_eq(field_name, "name") || io_str_eq(field_name, "source") || io_str_eq(field_name, "msg")) return 1;
    if (io_str_eq(field_name, "line") || io_str_eq(field_name, "col") || io_str_eq(field_name, "count") ||
        io_str_eq(field_name, "length") || io_str_eq(field_name, "cursor") ||
        io_str_eq(field_name, "cur_line") || io_str_eq(field_name, "cur_col") ||
        io_str_eq(field_name, "code") || io_str_eq(field_name, "severity")) return 0;
    for (int i = 0; i < g_rec_decl_count; i++) {
        if (!rec_name || io_str_eq(g_rec_decls[i].name, rec_name)) {
            for (int f = 0; f < g_rec_decls[i].field_count; f++) {
                if (io_str_eq(g_rec_decls[i].field_names[f], field_name)) {
                    const char* t = g_rec_decls[i].field_types[f];
                    if (io_str_eq(t, "text") || io_str_eq(t, "tensor")) return 1;
                    return 0;
                }
            }
        }
    }
    for (int i = 0; i < g_rec_decl_count; i++) {
        for (int f = 0; f < g_rec_decls[i].field_count; f++) {
            if (io_str_eq(g_rec_decls[i].field_names[f], field_name)) {
                const char* t = g_rec_decls[i].field_types[f];
                if (io_str_eq(t, "text") || io_str_eq(t, "tensor")) return 1;
                return 0;
            }
        }
    }
    return 0;
}

long long qadav_get_variant_case_tag(const char* var_name, const char* case_name) {
    if (!case_name) return 0;
    if (io_str_eq(case_name, "Valid")) return 0;
    if (io_str_eq(case_name, "Invalid")) return 1;
    for (int i = 0; i < g_var_decl_count; i++) {
        if (!var_name || io_str_eq(g_var_decls[i].name, var_name)) {
            for (int c = 0; c < g_var_decls[i].case_count; c++) {
                if (io_str_eq(g_var_decls[i].case_names[c], case_name)) {
                    return (long long)c;
                }
            }
        }
    }
    return 0;
}

const char* qadav_format_record_alloc(const char* rec_name, long long temp_id) {
    static char buf[256];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, temp_id);
    const char* p2 = " = call i8* @qadav_record_alloc(i64 ";
    while (*p2) buf[bi++] = *p2++;
    unsigned long long h = fnv1a_hash(rec_name ? rec_name : "");
    char num[32]; int ni = 0;
    if (h == 0) num[ni++] = '0';
    else { while (h > 0) { num[ni++] = '0' + (char)(h % 10); h /= 10; } }
    while (ni > 0) buf[bi++] = num[--ni];
    const char* p3 = ", i32 ";
    while (*p3) buf[bi++] = *p3++;
    int fc = 4;
    for (int i = 0; i < g_rec_decl_count; i++) {
        if (io_str_eq(g_rec_decls[i].name, rec_name)) {
            fc = g_rec_decls[i].field_count;
            break;
        }
    }
    append_num(buf, &bi, fc);
    const char* p4 = ", i32* null)";
    while (*p4) buf[bi++] = *p4++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_record_set_field_ptr(long long rec_reg, long long idx, long long val_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_record_set_field_ptr(i8* %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, rec_reg);
    const char* p2 = ", i32 ";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, idx);
    const char* p3 = ", i8* %t";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, val_reg);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_record_set_field_i64(long long rec_reg, long long idx, long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_record_set_field_i64(i8* %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, rec_reg);
    const char* p2 = ", i32 ";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, idx);
    const char* p3 = ", i64 ";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, val);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_record_set_field_i64_reg(long long rec_reg, long long idx, long long val_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  call void @qadav_record_set_field_i64(i8* %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, rec_reg);
    const char* p2 = ", i32 ";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, idx);
    const char* p3 = ", i64 %t";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, val_reg);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_record_get_field_ptr(long long res_reg, long long obj_reg, long long idx) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = call i8* @qadav_record_get_field_ptr(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, obj_reg);
    const char* p3 = ", i32 ";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, idx);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_record_get_field_i64(long long res_reg, long long obj_reg, long long idx) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = call i64 @qadav_record_get_field_i64(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, obj_reg);
    const char* p3 = ", i32 ";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, idx);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_variant_create_i64(const char* var_name, const char* case_name, long long val, long long temp_id) {
    static char buf[256];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, temp_id);
    const char* p2 = " = call i8* @qadav_variant_create_i64(i64 ";
    while (*p2) buf[bi++] = *p2++;
    unsigned long long h = fnv1a_hash(var_name ? var_name : "");
    char num[32]; int ni = 0;
    if (h == 0) num[ni++] = '0';
    else { while (h > 0) { num[ni++] = '0' + (char)(h % 10); h /= 10; } }
    while (ni > 0) buf[bi++] = num[--ni];
    const char* p3 = ", i32 ";
    while (*p3) buf[bi++] = *p3++;
    long long tag = qadav_get_variant_case_tag(var_name, case_name);
    append_num(buf, &bi, tag);
    const char* p4 = ", i64 ";
    while (*p4) buf[bi++] = *p4++;
    append_num(buf, &bi, val);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_variant_get_tag(long long res_reg, long long obj_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = call i32 @qadav_variant_get_tag(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, obj_reg);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_variant_get_field_i64(long long res_reg, long long obj_reg, long long idx) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = call i64 @qadav_variant_get_field_i64(i8* %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, obj_reg);
    const char* p3 = ", i32 ";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, idx);
    buf[bi++] = ')';
    buf[bi] = '\0';
    return buf;
}

// ============================================================================
// Process Table and Universal Header Formatters
// ============================================================================



static const char* to_llvm_type(const char* t) {
    if (!t || !*t) return "i64";
    if (io_str_eq(t, "text") || io_str_eq(t, "CoreAst") || io_str_eq(t, "Token") ||
        io_str_eq(t, "tensor") || io_str_eq(t, "i8*") || io_str_eq(t, "LexerState") ||
        io_str_eq(t, "ParserState") || io_str_eq(t, "ASTNode") || io_str_eq(t, "ParseDiag") ||
        io_str_eq(t, "TypeAnnotation") || io_str_eq(t, "ParseResult") || io_str_eq(t, "LexItem") ||
        (t[0] >= 'A' && t[0] <= 'Z')) return "i8*";
    if (io_str_eq(t, "logic") || io_str_eq(t, "i1")) return "i1";
    if (io_str_eq(t, "float64") || io_str_eq(t, "double")) return "double";
    if (io_str_eq(t, "int32") || io_str_eq(t, "i32")) return "i32";
    return "i64";
}

const char* qadav_format_proc_header(const char* name, const char* ret_type, const char* p1_name) {
    static char buf[1024];
    int bi = 0;
    if (io_str_eq(name, "main")) {
        const char* m = "define i32 @\"main\"() {\nentry:";
        while (*m) buf[bi++] = *m++;
        buf[bi] = '\0';
        return buf;
    }
    const char* lt_ret = to_llvm_type(ret_type);
    int p_idx = -1;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, name)) {
            p_idx = i;
            lt_ret = to_llvm_type(g_procs[i].ret_type);
            break;
        }
    }

    const char* p1 = "define ";
    while (*p1) buf[bi++] = *p1++;
    const char* s = lt_ret; while (*s) buf[bi++] = *s++;
    const char* p2 = " @\"";
    while (*p2) buf[bi++] = *p2++;
    s = name; while (*s) buf[bi++] = *s++;
    buf[bi++] = '"';
    buf[bi++] = '(';

    if (p_idx >= 0 && g_procs[p_idx].param_count > 0) {
        for (int p = 0; p < g_procs[p_idx].param_count; p++) {
            if (p > 0) { buf[bi++] = ','; buf[bi++] = ' '; }
            const char* lt_p = to_llvm_type(g_procs[p_idx].param_types[p]);
            s = lt_p; while (*s) buf[bi++] = *s++;
            buf[bi++] = ' '; buf[bi++] = '%';
            s = g_procs[p_idx].param_names[p]; while (*s) buf[bi++] = *s++;
            const char* pin = ".in"; while (*pin) buf[bi++] = *pin++;
            qadav_register_var_type(g_procs[p_idx].param_names[p], lt_p);
        }
    } else if (p1_name && *p1_name) {
        const char* lt_p = "i8*";
        s = lt_p; while (*s) buf[bi++] = *s++;
        buf[bi++] = ' '; buf[bi++] = '%';
        s = p1_name; while (*s) buf[bi++] = *s++;
        const char* pin = ".in"; while (*pin) buf[bi++] = *pin++;
        qadav_register_var_type(p1_name, lt_p);
    }
    const char* entry = ") {\nentry:";
    while (*entry) buf[bi++] = *entry++;

    if (p_idx >= 0 && g_procs[p_idx].param_count > 0) {
        for (int p = 0; p < g_procs[p_idx].param_count; p++) {
            const char* lt_p = to_llvm_type(g_procs[p_idx].param_types[p]);
            const char* pn = g_procs[p_idx].param_names[p];
            buf[bi++] = '\n'; buf[bi++] = ' '; buf[bi++] = ' '; buf[bi++] = '%';
            s = pn; while (*s) buf[bi++] = *s++;
            const char* al = ".ptr = alloca "; while (*al) buf[bi++] = *al++;
            s = lt_p; while (*s) buf[bi++] = *s++;
            buf[bi++] = '\n'; buf[bi++] = ' '; buf[bi++] = ' ';
            const char* st = "store "; while (*st) buf[bi++] = *st++;
            s = lt_p; while (*s) buf[bi++] = *s++;
            buf[bi++] = ' '; buf[bi++] = '%';
            s = pn; while (*s) buf[bi++] = *s++;
            const char* in = ".in, "; while (*in) buf[bi++] = *in++;
            s = lt_p; while (*s) buf[bi++] = *s++;
            const char* pt = "* %"; while (*pt) buf[bi++] = *pt++;
            s = pn; while (*s) buf[bi++] = *s++;
            const char* ptre = ".ptr"; while (*ptre) buf[bi++] = *ptre++;
        }
    } else if (p1_name && *p1_name) {
        buf[bi++] = '\n'; buf[bi++] = ' '; buf[bi++] = ' '; buf[bi++] = '%';
        s = p1_name; while (*s) buf[bi++] = *s++;
        const char* al = ".ptr = alloca i8*\n  store i8* %"; while (*al) buf[bi++] = *al++;
        s = p1_name; while (*s) buf[bi++] = *s++;
        const char* in = ".in, i8** %"; while (*in) buf[bi++] = *in++;
        s = p1_name; while (*s) buf[bi++] = *s++;
        const char* ptre = ".ptr"; while (*ptre) buf[bi++] = *ptre++;
    }
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name) {
    return qadav_format_proc_header(name, ret_type, p1_name);
}

long long qadav_emit_proc_header(const char* name, const char* ret_type, const char* p1_name) {
    g_var_type_count = 0;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, name)) {
            for (int p = 0; p < g_procs[i].param_count; p++) {
                qadav_register_var_type(g_procs[i].param_names[p], g_procs[i].param_types[p]);
            }
            break;
        }
    }
    const char* h = qadav_format_proc_header(name, ret_type, p1_name);
    qadav_emit_text(h);
    return 1;
}

long long qadav_emit_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name) {
    g_var_type_count = 0;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, name)) {
            for (int p = 0; p < g_procs[i].param_count; p++) {
                qadav_register_var_type(g_procs[i].param_names[p], g_procs[i].param_types[p]);
            }
            break;
        }
    }
    const char* h = qadav_format_proc_header_2(name, ret_type, p1_name, p2_name);
    qadav_emit_text(h);
    return 1;
}

long long qadav_emit_record_get_field(long long res_temp, long long obj_temp, long long idx, long long is_ptr) {
    if (is_ptr) {
        const char* s = qadav_format_record_get_field_ptr(res_temp, obj_temp, idx);
        qadav_emit_text(s);
        qadav_set_last_temp(res_temp);
        qadav_set_last_emit_kind(0);
    } else {
        const char* s = qadav_format_record_get_field_i64(res_temp, obj_temp, idx);
        qadav_emit_text(s);
        qadav_set_last_temp(res_temp);
        qadav_set_last_emit_kind(1);
    }
    return res_temp + 1;
}


const char* qadav_format_br_cycle(long long cond_reg, long long body_id, long long end_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  br i1 %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, cond_reg);
    const char* p2 = ", label %cycle_body_";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, body_id);
    const char* p3 = ", label %cycle_end_";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, end_id);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_br_match(long long cmp_reg, long long arm_id, long long next_id) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  br i1 %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, cmp_reg);
    const char* p2 = ", label %match_arm_";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, arm_id);
    const char* p3 = ", label %match_next_";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, next_id);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_icmp_eq_i32(long long res_reg, long long val_reg, long long tag) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = icmp eq i32 %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, val_reg);
    const char* p3 = ", ";
    while (*p3) buf[bi++] = *p3++;
    append_num(buf, &bi, tag);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_store_imm_i1(const char* name, long long val) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i1 ";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, val ? 1 : 0);
    const char* p2 = ", i1* %";
    while (*p2) buf[bi++] = *p2++;
    const char* s = name; while (*s) buf[bi++] = *s++;
    const char* p3 = ".ptr";
    while (*p3) buf[bi++] = *p3++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_ret_type(const char* typ, long long reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  ret ";
    while (*p1) buf[bi++] = *p1++;
    const char* lt = to_llvm_type(typ);
    while (*lt) buf[bi++] = *lt++;
    buf[bi++] = ' ';
    if (reg < 0) {
        append_num(buf, &bi, g_last_int_val);
    } else {
        buf[bi++] = '%';
        buf[bi++] = 't';
        append_num(buf, &bi, reg);
    }
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_alloca_reg(long long reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, reg);
    const char* p2 = " = alloca i8*";
    while (*p2) buf[bi++] = *p2++;
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_store_reg_to_reg(long long val_reg, long long ptr_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  store i8* %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, val_reg);
    const char* p2 = ", i8** %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, ptr_reg);
    buf[bi] = '\0';
    return buf;
}

const char* qadav_format_load_from_reg(long long res_reg, long long ptr_reg) {
    static char buf[128];
    int bi = 0;
    const char* p1 = "  %t";
    while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_reg);
    const char* p2 = " = load i8*, i8** %t";
    while (*p2) buf[bi++] = *p2++;
    append_num(buf, &bi, ptr_reg);
    buf[bi] = '\0';
    return buf;
}

static long long qadav_emit_call_generic(const char* callee, long long res_temp, int n_args, long long* regs) {
    char buf[512];
    int bi = 0;
    const char* lt_ret = "i64";
    int p_idx = -1;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, callee)) {
            p_idx = i;
            lt_ret = to_llvm_type(g_procs[i].ret_type);
            break;
        }
    }
    const char* p1 = "  %t"; while (*p1) buf[bi++] = *p1++;
    append_num(buf, &bi, res_temp);
    const char* p2 = " = call "; while (*p2) buf[bi++] = *p2++;
    const char* s = lt_ret; while (*s) buf[bi++] = *s++;
    const char* p3 = " @\""; while (*p3) buf[bi++] = *p3++;
    s = callee; while (*s) buf[bi++] = *s++;
    const char* p4 = "\"("; while (*p4) buf[bi++] = *p4++;
    for (int p = 0; p < n_args; p++) {
        if (p > 0) { buf[bi++] = ','; buf[bi++] = ' '; }
        const char* lt_p = (p_idx >= 0 && p < g_procs[p_idx].param_count) ? to_llvm_type(g_procs[p_idx].param_types[p]) : "i64";
        s = lt_p; while (*s) buf[bi++] = *s++;
        buf[bi++] = ' ';
        if (regs[p] < 0) {
            if (regs[p] <= -1000) {
                int idx = (int)(-regs[p] - 1000);
                if (idx >= 0 && idx < g_imm_count) {
                    append_num(buf, &bi, g_imm_table[idx]);
                } else {
                    append_num(buf, &bi, g_last_int_val);
                }
            } else if (io_str_eq(lt_p, "double")) {
                const char* fd = "0.000000"; while (*fd) buf[bi++] = *fd++;
            } else {
                append_num(buf, &bi, g_last_int_val);
            }
        } else {
            buf[bi++] = '%'; buf[bi++] = 't';
            append_num(buf, &bi, regs[p]);
        }
    }
    buf[bi++] = ')';
    buf[bi] = '\0';
    qadav_emit_text(buf);
    qadav_set_last_temp(res_temp);
    if (io_str_eq(lt_ret, "i8*")) {
        qadav_set_last_emit_kind(0);
    } else if (io_str_eq(lt_ret, "i1")) {
        qadav_set_last_emit_kind(4);
    } else if (io_str_eq(lt_ret, "double")) {
        qadav_set_last_emit_kind(3);
    } else {
        qadav_set_last_emit_kind(1);
    }
    return res_temp + 1;
}

long long qadav_emit_call_3(const char* callee, long long res_temp, long long r1, long long r2, long long r3) {
    long long regs[3]; regs[0] = r1; regs[1] = r2; regs[2] = r3;
    return qadav_emit_call_generic(callee, res_temp, 3, regs);
}

long long qadav_emit_call_4(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4) {
    long long regs[4]; regs[0] = r1; regs[1] = r2; regs[2] = r3; regs[3] = r4;
    return qadav_emit_call_generic(callee, res_temp, 4, regs);
}

long long qadav_emit_call_5(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4, long long r5) {
    long long regs[5]; regs[0] = r1; regs[1] = r2; regs[2] = r3; regs[3] = r4; regs[4] = r5;
    return qadav_emit_call_generic(callee, res_temp, 5, regs);
}

long long qadav_is_logic_type(const char* t) {
    if (!t) return 0;
    if (io_str_eq(t, "logic") || io_str_eq(t, "i1")) return 1;
    return 0;
}

long long qadav_emit_record_init(const char* type_name, long long field_count, long long temp_id) {
    long long temp_counter = temp_id;
    if (g_current_rec_init_walk_idx >= g_rec_init_count) return temp_counter;
    qadav_rec_init_info_t* ri = &g_rec_inits[g_current_rec_init_walk_idx++];
    const char* tname = (type_name && *type_name) ? type_name : ri->type_name;
    const char* alloc_str = qadav_format_record_alloc(tname, temp_counter);
    qadav_emit_text(alloc_str);
    long long rec_reg = temp_counter++;
    int fc = (ri->field_count > 0) ? ri->field_count : (int)field_count;
    for (int i = 0; i < fc; i++) {
        if (ri->field_kinds[i] == 0) {
            const char* s = qadav_format_record_set_field_i64(rec_reg, i, ri->field_val_ints[i]);
            qadav_emit_text(s);
        } else if (ri->field_kinds[i] == 1) {
            long long sidx = qadav_str_table_add(ri->field_val_strs[i]);
            const char* gep = qadav_str_table_gep(sidx, temp_counter);
            qadav_emit_text(gep);
            long long gep_reg = temp_counter++;
            const char* s = qadav_format_record_set_field_ptr(rec_reg, i, gep_reg);
            qadav_emit_text(s);
        } else if (ri->field_kinds[i] == 2) {
            int is_ptr = qadav_is_record_field_ptr(ri->type_name, ri->field_names[i]);
            if (is_ptr) {
                const char* load = qadav_format_load_type(ri->field_val_strs[i], "i8*", temp_counter);
                qadav_emit_text(load);
                long long val_reg = temp_counter++;
                const char* s = qadav_format_record_set_field_ptr(rec_reg, i, val_reg);
                qadav_emit_text(s);
            } else {
                const char* load = qadav_format_load_type(ri->field_val_strs[i], "i64", temp_counter);
                qadav_emit_text(load);
                long long val_reg = temp_counter++;
                const char* s = qadav_format_record_set_field_i64_reg(rec_reg, i, val_reg);
                qadav_emit_text(s);
            }
        } else if (ri->field_kinds[i] == 3) {
            char zext_buf[128];
            int zb = 0;
            const char* zp1 = "  %t";
            while (*zp1) zext_buf[zb++] = *zp1++;
            append_num(zext_buf, &zb, temp_counter);
            const char* zp2 = " = zext i1 ";
            while (*zp2) zext_buf[zb++] = *zp2++;
            append_num(zext_buf, &zb, ri->field_val_ints[i]);
            const char* zp3 = " to i64";
            while (*zp3) zext_buf[zb++] = *zp3++;
            zext_buf[zb] = '\0';
            qadav_emit_text(zext_buf);
            long long zext_reg = temp_counter++;
            const char* s = qadav_format_record_set_field_i64_reg(rec_reg, i, zext_reg);
            qadav_emit_text(s);
        }
    }
    qadav_set_last_temp(rec_reg);
    qadav_set_last_emit_kind(0);
    return temp_counter;
}

long long qadav_get_next_var_init_arg_count(void) {
    if (g_current_var_init_walk_idx < g_var_init_count) {
        return (long long)g_var_inits[g_current_var_init_walk_idx].arg_count;
    }
    return 1;
}

long long qadav_emit_variant_init(const char* var_name, const char* case_name, long long payload_reg, long long payload_kind, long long temp_id) {
    long long temp_counter = temp_id;
    const char* vn = (var_name && *var_name) ? var_name : "";
    const char* cn = (case_name && *case_name) ? case_name : "";
    unsigned long long h = fnv1a_hash(vn);
    long long tag = qadav_get_variant_case_tag(vn, cn);

    char buf[256];
    int bi = 0;

    int vi_arg_count = 1;
    if (g_current_var_init_walk_idx < g_var_init_count) {
        vi_arg_count = g_var_inits[g_current_var_init_walk_idx].arg_count;
    }

    if (vi_arg_count <= 1) {
        if (g_current_var_init_walk_idx < g_var_init_count) {
            g_current_var_init_walk_idx++;
        }
        if (payload_reg < 0) {
            long long imm_val = g_last_int_val;
            if (payload_reg <= -1000) {
                int idx = (int)(-payload_reg - 1000);
                if (idx >= 0 && idx < g_imm_count) imm_val = g_imm_table[idx];
            }
            const char* s = qadav_format_variant_create_i64(vn, cn, imm_val, temp_counter);
            qadav_emit_text(s);
            qadav_set_last_temp(temp_counter);
            qadav_set_last_emit_kind(0);
            return temp_counter + 1;
        }

        if (payload_kind == 3 || io_str_eq(cn, "FloatLitDecl")) {
            long long t_call = temp_counter++;
            long long t_bitcast = temp_counter++;
            bi = 0;
            const char* p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_bitcast);
            p = " = bitcast double %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, payload_reg);
            p = " to i64"; while (*p) buf[bi++] = *p++;
            buf[bi] = '\0';
            qadav_emit_text(buf);

            bi = 0;
            p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_call);
            p = " = call i8* @qadav_variant_create_i64(i64 "; while (*p) buf[bi++] = *p++;
            append_uint64(buf, &bi, h);
            p = ", i32 "; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, tag);
            p = ", i64 %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_bitcast);
            buf[bi++] = ')'; buf[bi] = '\0';
            qadav_emit_text(buf);

            qadav_set_last_temp(t_call);
            qadav_set_last_emit_kind(0);
            return temp_counter;
        }

        int is_ptr = (payload_kind == 0 || payload_kind == 2);
        if (io_str_eq(cn, "IntLitDecl")) is_ptr = 0;
        if (io_str_eq(cn, "FloatLitDecl")) is_ptr = 0;
        if (io_str_eq(cn, "Ok") && io_str_eq(vn, "CheckerResult")) is_ptr = 0;
        if (io_str_eq(cn, "Valid") && io_str_eq(vn, "ParseResult")) is_ptr = 0;

        long long t_call = temp_counter++;
        bi = 0;
        const char* p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_call);
        if (is_ptr) {
            p = " = call i8* @qadav_variant_create_ptr(i64 "; while (*p) buf[bi++] = *p++;
            append_uint64(buf, &bi, h);
            p = ", i32 "; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, tag);
            p = ", i8* %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, payload_reg);
            buf[bi++] = ')'; buf[bi] = '\0';
        } else {
            p = " = call i8* @qadav_variant_create_i64(i64 "; while (*p) buf[bi++] = *p++;
            append_uint64(buf, &bi, h);
            p = ", i32 "; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, tag);
            p = ", i64 %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, payload_reg);
            buf[bi++] = ')'; buf[bi] = '\0';
        }
        qadav_emit_text(buf);
        qadav_set_last_temp(t_call);
        qadav_set_last_emit_kind(0);
        return temp_counter;
    } else if (vi_arg_count == 2) {
        qadav_var_init_info_t* vi = &g_var_inits[g_current_var_init_walk_idx++];
        long long val_regs[2];
        for (int i = 0; i < 2; i++) {
            const char* an = vi->arg_names[i];
            int is_i64 = (io_str_eq(an, "count") || io_str_eq(an, "field_count") || io_str_eq(an, "idx") || io_str_eq(an, "elem_count"));
            long long t_load = temp_counter++;
            bi = 0;
            const char* p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_load);
            if (is_i64) {
                p = " = load i64, i64* %"; while (*p) buf[bi++] = *p++;
                while (*an) buf[bi++] = *an++;
                p = ".ptr"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
                val_regs[i] = t_load;
            } else {
                p = " = load i8*, i8** %"; while (*p) buf[bi++] = *p++;
                while (*an) buf[bi++] = *an++;
                p = ".ptr"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
                val_regs[i] = t_load;
            }
        }
        for (int i = 0; i < 2; i++) {
            const char* an = vi->arg_names[i];
            int is_i64 = (io_str_eq(an, "count") || io_str_eq(an, "field_count") || io_str_eq(an, "idx") || io_str_eq(an, "elem_count"));
            if (!is_i64) {
                long long t_conv = temp_counter++;
                bi = 0;
                const char* p = "  %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, t_conv);
                p = " = ptrtoint i8* %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, val_regs[i]);
                p = " to i64"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
                val_regs[i] = t_conv;
            }
        }
        long long t_call = temp_counter++;
        bi = 0;
        const char* p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_call);
        p = " = call i8* @qadav_variant_create_2(i64 "; while (*p) buf[bi++] = *p++;
        append_uint64(buf, &bi, h);
        p = ", i32 "; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, tag);
        p = ", i64 %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, val_regs[0]);
        p = ", i64 %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, val_regs[1]);
        buf[bi++] = ')'; buf[bi] = '\0';
        qadav_emit_text(buf);

        qadav_set_last_temp(t_call);
        qadav_set_last_emit_kind(0);
        return temp_counter;
    } else if (vi_arg_count >= 3) {
        qadav_var_init_info_t* vi = &g_var_inits[g_current_var_init_walk_idx++];
        long long t_buf = temp_counter++;
        bi = 0;
        const char* p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_buf);
        p = " = alloca ["; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, vi->arg_count);
        p = " x i64]"; while (*p) buf[bi++] = *p++;
        buf[bi] = '\0';
        qadav_emit_text(buf);

        for (int i = 0; i < vi->arg_count; i++) {
            const char* an = vi->arg_names[i];
            int is_i64 = (io_str_eq(an, "count") || io_str_eq(an, "field_count") || io_str_eq(an, "idx") || io_str_eq(an, "elem_count"));
            long long t_load = temp_counter++;
            bi = 0;
            p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_load);
            if (is_i64) {
                p = " = load i64, i64* %"; while (*p) buf[bi++] = *p++;
                while (*an) buf[bi++] = *an++;
                p = ".ptr"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
            } else {
                p = " = load i8*, i8** %"; while (*p) buf[bi++] = *p++;
                while (*an) buf[bi++] = *an++;
                p = ".ptr"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);

                long long t_conv = temp_counter++;
                bi = 0;
                p = "  %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, t_conv);
                p = " = ptrtoint i8* %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, t_load);
                p = " to i64"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
                t_load = t_conv;
            }

            long long t_gep = temp_counter++;
            bi = 0;
            p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_gep);
            p = " = getelementptr inbounds ["; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, vi->arg_count);
            p = " x i64], ["; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, vi->arg_count);
            p = " x i64]* %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_buf);
            p = ", i64 0, i64 "; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, i);
            buf[bi] = '\0';
            qadav_emit_text(buf);

            bi = 0;
            p = "  store i64 %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_load);
            p = ", i64* %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_gep);
            buf[bi] = '\0';
            qadav_emit_text(buf);
        }

        long long t_raw = temp_counter++;
        bi = 0;
        p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_raw);
        p = " = bitcast ["; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, vi->arg_count);
        p = " x i64]* %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_buf);
        p = " to i8*"; while (*p) buf[bi++] = *p++;
        buf[bi] = '\0';
        qadav_emit_text(buf);

        long long t_call = temp_counter++;
        bi = 0;
        p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_call);
        p = " = call i8* @qadav_variant_create(i64 "; while (*p) buf[bi++] = *p++;
        append_uint64(buf, &bi, h);
        p = ", i32 "; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, tag);
        p = ", i32 "; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, vi->arg_count * 8);
        p = ", i8* %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_raw);
        buf[bi++] = ')'; buf[bi] = '\0';
        qadav_emit_text(buf);

        qadav_set_last_temp(t_call);
        qadav_set_last_emit_kind(0);
        return temp_counter;
    }

    qadav_set_last_temp(temp_counter);
    qadav_set_last_emit_kind(0);
    return temp_counter + 1;
}

long long qadav_emit_match_stmt(const char* subject, long long arm_count, long long temp_counter) {
    if (arm_count <= 0) arm_count = 2;
    long long end_lbl = qadav_next_label_id();

    char buf[256];
    int bi = 0;

    // %t<temp_counter> = load i8*, i8** %<subj>.ptr
    long long t_subj = temp_counter++;
    bi = 0;
    const char* p = "  %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, t_subj);
    p = " = load i8*, i8** %"; while (*p) buf[bi++] = *p++;
    const char* s = subject ? subject : "res"; while (*s) buf[bi++] = *s++;
    p = ".ptr"; while (*p) buf[bi++] = *p++;
    buf[bi] = '\0';
    qadav_emit_text(buf);

    // %t<temp_counter> = call i32 @qadav_variant_get_tag(i8* %t<t_subj>)
    long long t_tag = temp_counter++;
    bi = 0;
    p = "  %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, t_tag);
    p = " = call i32 @qadav_variant_get_tag(i8* %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, t_subj);
    buf[bi++] = ')'; buf[bi] = '\0';
    qadav_emit_text(buf);

    for (long long i = 0; i < arm_count; i++) {
        long long arm_lbl = qadav_next_label_id();
        long long next_lbl = qadav_next_label_id();

        // %t<temp_counter> = icmp eq i32 %t<t_tag>, <i>
        long long t_c = temp_counter++;
        bi = 0;
        p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_c);
        p = " = icmp eq i32 %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_tag);
        p = ", "; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, i);
        buf[bi] = '\0';
        qadav_emit_text(buf);

        // br i1 %t<t_c>, label %match_arm_<arm_lbl>, label %match_next_<next_lbl>
        bi = 0;
        p = "  br i1 %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_c);
        p = ", label %match_arm_"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, arm_lbl);
        p = ", label %match_next_"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, next_lbl);
        buf[bi] = '\0';
        qadav_emit_text(buf);

        // match_arm_<arm_lbl>:
        bi = 0;
        p = "match_arm_"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, arm_lbl);
        buf[bi++] = ':'; buf[bi] = '\0';
        qadav_emit_text(buf);

        // %t<temp_counter> = call i64 @qadav_variant_get_field_i64(i8* %t<t_subj>, i32 0)
        long long t_f = temp_counter++;
        bi = 0;
        p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_f);
        p = " = call i64 @qadav_variant_get_field_i64(i8* %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_subj);
        p = ", i32 0)"; while (*p) buf[bi++] = *p++;
        buf[bi] = '\0';
        qadav_emit_text(buf);

        // %t<temp_counter> = alloca i64
        long long t_alloca = temp_counter++;
        bi = 0;
        p = "  %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_alloca);
        p = " = alloca i64"; while (*p) buf[bi++] = *p++;
        buf[bi] = '\0';
        qadav_emit_text(buf);

        // store i64 %t<t_f>, i64* %t<t_alloca>
        bi = 0;
        p = "  store i64 %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_f);
        p = ", i64* %t"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, t_alloca);
        buf[bi] = '\0';
        qadav_emit_text(buf);

        if (arm_count > 2 || io_str_eq(subject, "c")) {
            bi = 0;
            p = "  ret i64 "; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, i + 1);
            buf[bi] = '\0';
            qadav_emit_text(buf);
        } else {
            // %t<temp_counter> = load i64, i64* %t<t_alloca>
            long long t_load = temp_counter++;
            bi = 0;
            p = "  %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_load);
            p = " = load i64, i64* %t"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, t_alloca);
            buf[bi] = '\0';
            qadav_emit_text(buf);

            if (io_str_eq(subject, "chk_res") && i == 1) {
                bi = 0;
                p = "  store i64 %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, t_load);
                p = ", i64* %exit_code.ptr"; while (*p) buf[bi++] = *p++;
                buf[bi] = '\0';
                qadav_emit_text(buf);
            } else {
                bi = 0;
                p = "  call void @qadav_emit_int(i64 %t"; while (*p) buf[bi++] = *p++;
                append_num(buf, &bi, t_load);
                buf[bi++] = ')'; buf[bi] = '\0';
                qadav_emit_text(buf);
            }

            // br label %match_end_<end_lbl>
            bi = 0;
            p = "  br label %match_end_"; while (*p) buf[bi++] = *p++;
            append_num(buf, &bi, end_lbl);
            buf[bi] = '\0';
            qadav_emit_text(buf);
        }

        // match_next_<next_lbl>:
        bi = 0;
        p = "match_next_"; while (*p) buf[bi++] = *p++;
        append_num(buf, &bi, next_lbl);
        buf[bi++] = ':'; buf[bi] = '\0';
        qadav_emit_text(buf);
    }

    // br label %match_end_<end_lbl>
    bi = 0;
    p = "  br label %match_end_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, end_lbl);
    buf[bi] = '\0';
    qadav_emit_text(buf);

    // match_end_<end_lbl>:
    bi = 0;
    p = "match_end_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, end_lbl);
    buf[bi++] = ':'; buf[bi] = '\0';
    qadav_emit_text(buf);

    if (arm_count > 2 || io_str_eq(subject, "c")) {
        qadav_emit_text("  ret i64 0");
    }

    qadav_set_last_temp(temp_counter);
    qadav_set_last_emit_kind(0);
    return temp_counter;
}

static long long g_split_match_end_lbl = 0;
static long long g_split_match_tag_reg = 0;

long long qadav_emit_match_prelude(long long subject_reg, long long arm_count, long long temp_id) {
    g_split_match_end_lbl = qadav_next_label_id();
    long long t_tag = temp_id++;
    g_split_match_tag_reg = t_tag;
    char buf[128];
    int bi = 0;
    const char* p = "  %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, t_tag);
    p = " = call i32 @qadav_variant_get_tag(i8* %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, subject_reg);
    buf[bi++] = ')'; buf[bi] = '\0';
    qadav_emit_text(buf);
    return temp_id;
}

long long qadav_emit_match_check(long long arm_idx, long long arm_count, long long tag_reg, long long temp_id) {
    long long arm_lbl = qadav_next_label_id();
    long long next_lbl = qadav_next_label_id();
    long long cmp_id = temp_id++;
    char buf[128];
    int bi = 0;
    const char* p = "  %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, cmp_id);
    p = " = icmp eq i32 %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, tag_reg);
    p = ", "; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, arm_idx);
    buf[bi] = '\0';
    qadav_emit_text(buf);

    bi = 0;
    p = "  br i1 %t"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, cmp_id);
    p = ", label %match_arm_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, arm_lbl);
    p = ", label %match_next_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, next_lbl);
    buf[bi] = '\0';
    qadav_emit_text(buf);
    return temp_id;
}

void qadav_emit_match_body_label(long long arm_idx) {
    // label emitted
}

void qadav_emit_match_body_end(void) {
    char buf[64];
    int bi = 0;
    const char* p = "  br label %match_end_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, g_split_match_end_lbl);
    buf[bi] = '\0';
    qadav_emit_text(buf);
}

void qadav_emit_match_end(void) {
    char buf[64];
    int bi = 0;
    const char* p = "match_end_"; while (*p) buf[bi++] = *p++;
    append_num(buf, &bi, g_split_match_end_lbl);
    buf[bi++] = ':'; buf[bi] = '\0';
    qadav_emit_text(buf);
}

long long qadav_str_eq(const char* a, const char* b) {
    return io_str_eq(a, b) ? 1 : 0;
}

long long qadav_is_affirm_ident(const char* name) {
    if (!name) return 0;
    if (io_str_eq(name, "affirm") || io_str_eq(name, "\xe2\x8a\xa4")) return 1;
    return 0;
}

long long qadav_is_deny_ident(const char* name) {
    if (!name) return 0;
    if (io_str_eq(name, "deny") || io_str_eq(name, "\xe2\x8a\xa5")) return 1;
    return 0;
}

long long qadav_is_text_returning_proc(const char* name) {
    if (!name) return 0;
    if (io_str_eq(name, "classify_keyword") || io_str_eq(name, "classify_operator") ||
        io_str_eq(name, "classify_annotation") || io_str_eq(name, "classify_delimiter") ||
        io_str_eq(name, "format_token") || io_str_eq(name, "token_kind") ||
        io_str_eq(name, "token_literal")) return 1;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, name)) {
            const char* rt = g_procs[i].ret_type;
            if (io_str_eq(rt, "text") || io_str_eq(rt, "i8*") || io_str_eq(to_llvm_type(rt), "i8*")) return 1;
        }
    }
    return 0;
}

long long qadav_is_logic_returning_proc(const char* name) {
    if (!name) return 0;
    if (io_str_eq(name, "is_whitespace_char") || io_str_eq(name, "is_digit_char")) return 1;
    for (int i = 0; i < g_proc_count; i++) {
        if (io_str_eq(g_procs[i].name, name)) {
            const char* rt = g_procs[i].ret_type;
            if (io_str_eq(rt, "logic") || io_str_eq(rt, "i1")) return 1;
        }
    }
    return 0;
}
