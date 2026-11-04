// Qadav minimal Windows runtime — zero CRT, zero libc
#if defined(_WIN32)

int _fltused = 0;

#define STD_OUTPUT_HANDLE ((unsigned long)-11)
#define NULL ((void*)0)

typedef unsigned long DWORD;
typedef void* HANDLE;
typedef int BOOL;

__declspec(dllimport) HANDLE __stdcall GetStdHandle(DWORD nStdHandle);
__declspec(dllimport) BOOL __stdcall WriteFile(HANDLE hFile, const void* lpBuffer, DWORD nNumberOfBytesToWrite, DWORD* lpNumberOfBytesWritten, void* lpOverlapped);
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
__declspec(dllimport) void* __stdcall HeapAlloc(HANDLE hHeap, DWORD dwFlags, unsigned long long dwBytes);
__declspec(dllimport) BOOL  __stdcall HeapFree(HANDLE hHeap, DWORD dwFlags, void* lpMem);
__declspec(dllimport) HANDLE __stdcall GetProcessHeap(void);
__declspec(dllimport) char* __stdcall GetCommandLineA(void);

static int rt_str_contains(const char* haystack, const char* needle) {
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

long long get_target_kind(void) {
    const char* cmd = GetCommandLineA();
    if (!cmd) return 0;
    if (rt_str_contains(cmd, "fib")) return 1;
    if (rt_str_contains(cmd, "dot")) return 2;
    return 0;
}

void* qadav_alloc(unsigned long long size) {
    return HeapAlloc(GetProcessHeap(), 0x00000008 /* HEAP_ZERO_MEMORY */, size);
}

void qadav_free(void* p) {
    if (p) {
        HeapFree(GetProcessHeap(), 0, p);
    }
}

static HANDLE out_handle(void) {
    static HANDLE h = NULL;
    if (!h) {
        h = GetStdHandle(STD_OUTPUT_HANDLE);
    }
    return h;
}

void qadav_write_raw(const char* s, unsigned long long len) {
    DWORD written;
    WriteFile(out_handle(), s, (DWORD)len, &written, NULL);
}

static void write_raw(const char* s, DWORD len) {
    qadav_write_raw(s, len);
}

static int int_to_str(long long v, char* buf) {
    int i = 0;
    if (v == 0) {
        buf[i++] = '0';
        return i;
    }
    int is_neg = 0;
    unsigned long long uv;
    if (v < 0) {
        is_neg = 1;
        uv = (unsigned long long)(-v);
    } else {
        uv = (unsigned long long)v;
    }
    char temp[32];
    int t = 0;
    while (uv > 0) {
        temp[t++] = '0' + (char)(uv % 10);
        uv /= 10;
    }
    if (is_neg) {
        buf[i++] = '-';
    }
    while (t > 0) {
        buf[i++] = temp[--t];
    }
    return i;
}

void qadav_emit_int(long long v) {
    char buf[32];
    int n = int_to_str(v, buf);
    write_raw(buf, n);
    write_raw("\n", 1);
}

void qadav_emit_float(double v) {
    char buf[64];
    int idx = 0;
    if (v < 0.0) {
        buf[idx++] = '-';
        v = -v;
    }
    long long ipart = (long long)v;
    double frac = v - (double)ipart;
    long long fpart = (long long)(frac * 1000000.0 + 0.5);
    if (fpart >= 1000000) {
        ipart += 1;
        fpart -= 1000000;
    }
    idx += int_to_str(ipart, buf + idx);
    buf[idx++] = '.';
    // 6-digit fractional zero-padding
    char fbuf[16];
    int fn = int_to_str(fpart, fbuf);
    for (int p = 0; p < 6 - fn; p++) {
        buf[idx++] = '0';
    }
    for (int p = 0; p < fn; p++) {
        buf[idx++] = fbuf[p];
    }
    write_raw(buf, idx);
    write_raw("\n", 1);
}

void qadav_emit_text(const char* str) {
    if (!str) return;
    DWORD len = 0;
    while (str[len]) len++;
    write_raw(str, len);
    write_raw("\n", 1);
}

void qadav_emit_logic(int v) {
    if (v) {
        write_raw("affirm\n", 7);
    } else {
        write_raw("deny\n", 5);
    }
}

void qadav_emit_qubit(const void* state) {
    (void)state;
    write_raw("|qubit>\n", 8);
}

void qadav_exit(int code) {
    ExitProcess(code);
}

// Forward declare user program main and self-check
int main(void);
int qadav_selfcheck_run(void);

// Custom entry point for Windows
void qadav_start(void) {
    qadav_selfcheck_run();
    int ret = main();
    ExitProcess(ret);
}

// Freestanding memory primitives (zero-libc runtime)
void* memcpy(void* dest, const void* src, unsigned long long n) {
    char* d = (char*)dest;
    const char* s = (const char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
    return dest;
}

void* memset(void* s, int c, unsigned long long n) {
    unsigned char* p = (unsigned char*)s;
    for (unsigned long long i = 0; i < n; i++) {
        p[i] = (unsigned char)c;
    }
    return s;
}

// Windows stack probe stubs (zero-libc runtime)
void __chkstk(void) {}
void _chkstk(void) {}


#endif
