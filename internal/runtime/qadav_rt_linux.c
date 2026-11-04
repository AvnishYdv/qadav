// Qadav minimal Linux runtime — zero CRT, zero libc
#if !defined(_WIN32)

static long sys_write(int fd, const void* buf, unsigned long count) {
    long ret;
    __asm__ volatile (
        "syscall"
        : "=a"(ret)
        : "a"(1), "D"(fd), "S"(buf), "d"(count)
        : "rcx", "r11", "memory"
    );
    return ret;
}

static void sys_exit(int code) {
    __asm__ volatile (
        "syscall"
        :
        : "a"(60), "D"(code)
        : "rcx", "r11", "memory"
    );
    __builtin_unreachable();
}

void qadav_write_raw(const char* s, unsigned long long len) {
    sys_write(1, s, (unsigned long)len);
}

static void write_raw(const char* s, unsigned long len) {
    qadav_write_raw(s, len);
}

static void* qadav_mmap(unsigned long size) {
    void* ret;
    register long rax __asm__("rax") = 9;   // sys_mmap
    register long rdi __asm__("rdi") = 0;
    register long rsi __asm__("rsi") = size;
    register long rdx __asm__("rdx") = 3;    // PROT_READ|PROT_WRITE
    register long r10 __asm__("r10") = 0x22; // MAP_PRIVATE|MAP_ANONYMOUS
    register long r8  __asm__("r8")  = -1;
    register long r9  __asm__("r9")  = 0;
    __asm__ volatile("syscall" : "=r"(rax) : "r"(rax),"r"(rdi),"r"(rsi),"r"(rdx),"r"(r10),"r"(r8),"r"(r9) : "rcx","r11","memory");
    return (void*)rax;
}

void* qadav_alloc(unsigned long long size) {
    return qadav_mmap((unsigned long)size);
}

void qadav_free(void* p) {
    (void)p;
}

long long get_target_kind(void) {
    return 0;
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
    unsigned long len = 0;
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
    sys_exit(code);
}

// Forward declare user program main
int main(void);

// Custom entry point for Linux
void _start(void) {
    int ret = main();
    sys_exit(ret);
}

#endif

