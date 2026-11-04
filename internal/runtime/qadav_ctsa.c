#include "qadav_ctsa.h"
#include "qadav_sha3.h"
#include "qadav_qeeh.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static qadav_ctsa_certificate g_ctsa_cert = {0};
static int g_ctsa_initialized = 0;

static inline unsigned long long ctsa_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64)
    unsigned int lo, hi;
    __asm__ __volatile__ ("rdtsc" : "=a" (lo), "=d" (hi));
    return ((unsigned long long)hi << 32) | lo;
#else
    return 987654321ULL;
#endif
}

static void ctsa_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void ctsa_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void compute_ctsa_cert_hash(const qadav_ctsa_certificate* cert, unsigned char out_hash[32]) {
    // Hash preimage: ast_hash (32) || type_env_hash (32) || rule_chain_hash (32) ||
    // num_sigs (8) || num_fields (8) || timestamp (8)
    unsigned char preimage[120];
    ctsa_memcpy(preimage, cert->ast_hash, 32);
    ctsa_memcpy(preimage + 32, cert->type_env_hash, 32);
    ctsa_memcpy(preimage + 64, cert->rule_chain_hash, 32);
    ctsa_memcpy(preimage + 96, &cert->num_signatures, 8);
    ctsa_memcpy(preimage + 104, &cert->num_fields, 8);
    ctsa_memcpy(preimage + 112, &cert->timestamp, 8);

    qadav_sha3_256(preimage, 120, out_hash);
}

void qadav_ctsa_init(void) {
    if (!g_ctsa_initialized) {
        ctsa_memset(&g_ctsa_cert, 0, sizeof(g_ctsa_cert));
        g_ctsa_initialized = 1;
    }
}

int qadav_ctsa_attest(const unsigned char ast_hash[32],
                      const unsigned char type_env_hash[32],
                      const unsigned char rule_chain_hash[32],
                      unsigned long long num_sigs,
                      unsigned long long num_fields) {
    qadav_ctsa_init();
    ctsa_memcpy(g_ctsa_cert.ast_hash, ast_hash, 32);
    ctsa_memcpy(g_ctsa_cert.type_env_hash, type_env_hash, 32);
    ctsa_memcpy(g_ctsa_cert.rule_chain_hash, rule_chain_hash, 32);
    g_ctsa_cert.num_signatures = num_sigs;
    g_ctsa_cert.num_fields = num_fields;
    g_ctsa_cert.timestamp = ctsa_rdtsc();

    compute_ctsa_cert_hash(&g_ctsa_cert, g_ctsa_cert.cert_hash);
    g_ctsa_cert.is_verified = 1;
    return 1;
}

int qadav_ctsa_verify(void) {
    if (!g_ctsa_initialized || !g_ctsa_cert.is_verified) {
        return 0;
    }

    unsigned char expected[32];
    compute_ctsa_cert_hash(&g_ctsa_cert, expected);

    int diff = 0;
    for (int i = 0; i < 32; i++) {
        if (g_ctsa_cert.cert_hash[i] != expected[i]) {
            diff = 1;
            break;
        }
    }

    if (diff != 0) {
        // Tamper detected! Trigger decoherence and abort
#ifdef _WIN32
        ExitProcess(0xDEADBEEF);
#else
        __builtin_trap();
#endif
        return 0;
    }

    return 1;
}

int qadav_ctsa_corrupt(int field_type) {
    if (!g_ctsa_initialized) return 0;

    switch (field_type) {
        case 1: // corrupt ast_hash
            g_ctsa_cert.ast_hash[0] ^= 0xFF;
            break;
        case 2: // corrupt type_env_hash
            g_ctsa_cert.type_env_hash[0] ^= 0xFF;
            break;
        case 3: // corrupt rule_chain_hash
            g_ctsa_cert.rule_chain_hash[0] ^= 0xFF;
            break;
        case 4: // corrupt cert_hash
            g_ctsa_cert.cert_hash[0] ^= 0xFF;
            break;
        default:
            g_ctsa_cert.cert_hash[0] ^= 0xAA;
            break;
    }
    return 1;
}

int qadav_ctsa_get_status(void) {
    return g_ctsa_cert.is_verified;
}

void qadav_ctsa_reset(void) {
    ctsa_memset(&g_ctsa_cert, 0, sizeof(g_ctsa_cert));
    g_ctsa_initialized = 0;
}

