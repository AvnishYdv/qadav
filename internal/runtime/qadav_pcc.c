#include "qadav_pcc.h"
#include "qadav_sha3.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static qadav_pcc_certificate g_pcc_cert = {0};
static int g_pcc_initialized = 0;

static inline unsigned long long pcc_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64)
    unsigned int lo, hi;
    __asm__ __volatile__ ("rdtsc" : "=a" (lo), "=d" (hi));
    return ((unsigned long long)hi << 32) | lo;
#else
    return 1357924680ULL;
#endif
}

static void pcc_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void pcc_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void compute_pcc_cert_hash(const qadav_pcc_certificate* cert, unsigned char out_hash[32]) {
    // Hash preimage: ast_hash (32) || codegen_path_hash (32) || ir_hash (32) || timestamp (8)
    unsigned char preimage[104];
    pcc_memcpy(preimage, cert->ast_hash, 32);
    pcc_memcpy(preimage + 32, cert->codegen_path_hash, 32);
    pcc_memcpy(preimage + 64, cert->ir_hash, 32);
    pcc_memcpy(preimage + 96, &cert->timestamp, 8);

    qadav_sha3_256(preimage, 104, out_hash);
}

void qadav_pcc_init(void) {
    if (!g_pcc_initialized) {
        pcc_memset(&g_pcc_cert, 0, sizeof(g_pcc_cert));
        g_pcc_initialized = 1;
    }
}

int qadav_pcc_attest(const unsigned char ast_hash[32],
                     const unsigned char codegen_path_hash[32],
                     const unsigned char ir_hash[32]) {
    qadav_pcc_init();
    pcc_memcpy(g_pcc_cert.ast_hash, ast_hash, 32);
    pcc_memcpy(g_pcc_cert.codegen_path_hash, codegen_path_hash, 32);
    pcc_memcpy(g_pcc_cert.ir_hash, ir_hash, 32);
    g_pcc_cert.timestamp = pcc_rdtsc();

    compute_pcc_cert_hash(&g_pcc_cert, g_pcc_cert.cert_hash);
    g_pcc_cert.is_verified = 1;
    return 1;
}

int qadav_pcc_verify(void) {
    if (!g_pcc_initialized || !g_pcc_cert.is_verified) {
        return 0;
    }

    unsigned char expected[32];
    compute_pcc_cert_hash(&g_pcc_cert, expected);

    int diff = 0;
    for (int i = 0; i < 32; i++) {
        if (g_pcc_cert.cert_hash[i] != expected[i]) {
            diff = 1;
            break;
        }
    }

    if (diff != 0) {
        // Tamper detected! Trigger quantum decoherence fail-stop
#ifdef _WIN32
        ExitProcess(0xDEADBEEF);
#else
        __builtin_trap();
#endif
        return 0;
    }

    return 1;
}

int qadav_pcc_corrupt(int field_type) {
    if (!g_pcc_initialized) {
        return 0;
    }

    switch (field_type) {
    case 1: // Corrupt ast_hash
        g_pcc_cert.ast_hash[0] ^= 0xFF;
        break;
    case 2: // Corrupt codegen_path_hash
        g_pcc_cert.codegen_path_hash[0] ^= 0xFF;
        break;
    case 3: // Corrupt ir_hash
        g_pcc_cert.ir_hash[0] ^= 0xFF;
        break;
    case 4: // Corrupt cert_hash
        g_pcc_cert.cert_hash[0] ^= 0xFF;
        break;
    default:
        g_pcc_cert.cert_hash[31] ^= 0xAA;
        break;
    }

    return 1;
}

int qadav_pcc_get_status(void) {
    return g_pcc_initialized && g_pcc_cert.is_verified;
}

void qadav_pcc_reset(void) {
    pcc_memset(&g_pcc_cert, 0, sizeof(g_pcc_cert));
    g_pcc_initialized = 0;
}

