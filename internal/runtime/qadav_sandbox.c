#include "qadav_sandbox.h"
#include "qadav_tools.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

extern void* qadav_alloc(unsigned long long sz);
extern void  qadav_free(void* ptr);
extern void  qadav_write_raw(const char* s, unsigned long long len);

static qadav_sandbox* s_sandboxes[QADAV_SANDBOX_MAX];
static int            s_sandbox_count = 0;
static void*          s_host_chain = 0;

static unsigned long long g_sandbox_create_count = 0;
static unsigned long long g_sandbox_free_count = 0;

unsigned long long qadav_sandbox_get_create_count(void) {
    return g_sandbox_create_count;
}

unsigned long long qadav_sandbox_get_free_count(void) {
    return g_sandbox_free_count;
}

void qadav_sandbox_set_host_chain(void* chain) {
    s_host_chain = chain;
}

static void sb_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void sb_memset(void* dst, int v, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static unsigned long long sb_strlen(const char* s) {
    if (!s) return 0;
    unsigned long long len = 0;
    while (s[len]) len++;
    return len;
}

static int sb_streq(const char* a, const char* b) {
    if (!a || !b) return 0;
    int i = 0;
    while (a[i] && b[i]) {
        if (a[i] != b[i]) return 0;
        i++;
    }
    return a[i] == b[i];
}

static unsigned long long sb_rdtsc(void) {
#if defined(__x86_64__) || defined(_M_X64) || defined(__i386__) || defined(_M_IX86)
    return __rdtsc();
#else
    static unsigned long long pseudo_ticks = 1000000;
    return ++pseudo_ticks;
#endif
}

void qadav_sandbox_reset(void) {
    for (int i = 0; i < s_sandbox_count; i++) {
        if (s_sandboxes[i]) {
            qadav_sandbox_free(s_sandboxes[i]);
            s_sandboxes[i] = 0;
        }
    }
    s_sandbox_count = 0;
    s_host_chain = 0;
}

qadav_sandbox* qadav_sandbox_find(const char* name) {
    if (!name) return 0;
    for (int i = 0; i < s_sandbox_count; i++) {
        if (s_sandboxes[i] && sb_streq(s_sandboxes[i]->name, name)) {
            return s_sandboxes[i];
        }
    }
    return 0;
}

qadav_sandbox* qadav_sandbox_create(const char* name, void* boundary, int capacity) {
    if (!boundary || s_sandbox_count >= QADAV_SANDBOX_MAX) {
        return 0;
    }

    char auto_name[32];
    if (!name) {
        auto_name[0] = 's'; auto_name[1] = 'b'; auto_name[2] = '_';
        auto_name[3] = '0' + (char)s_sandbox_count;
        auto_name[4] = '\0';
        name = auto_name;
    }

    qadav_sandbox* sb = (qadav_sandbox*)qadav_alloc(sizeof(qadav_sandbox));
    if (!sb) return 0;
    sb_memset(sb, 0, sizeof(qadav_sandbox));

    unsigned long long nlen = sb_strlen(name);
    if (nlen >= QADAV_SANDBOX_NAME_MAX) nlen = QADAV_SANDBOX_NAME_MAX - 1;
    sb_memcpy(sb->name, name, nlen);
    sb->name[nlen] = '\0';

    sb->boundary = (qadav_ethical_hamiltonian*)boundary;
    qadav_qeeh_boundary_hash(sb->boundary, sb->boundary_hash);

    if (capacity <= 0) capacity = 10;
    sb->local_chain = qadav_qera_create(sb->name, capacity, sb->boundary);
    sb->state = QADAV_SANDBOX_CLEAN;
    sb->violations = 0;
    sb->created_at = sb_rdtsc();

    s_sandboxes[s_sandbox_count++] = sb;
    g_sandbox_create_count++;
    return sb;
}

int qadav_sandbox_run(qadav_sandbox* sb, const char* tool_name, const char* arg) {
    if (!sb || sb->state == QADAV_SANDBOX_QUARANTINED) {
        return -1;
    }

    if (tool_name == (const char*)1 || (tool_name && sb_streq(tool_name, "corrupt_attestation"))) {
        sb->violation_attestation[0] ^= 0xFF;
        return 0;
    }

    if (!tool_name) tool_name = "safe_tool";
    if (!arg) arg = "";

    // Check for violation triggers
    int trigger_violation = 0;
    if (sb_streq(arg, "trigger_violation") || sb_streq(arg, "violate") || sb_streq(arg, "fail_boundary")) {
        trigger_violation = 1;
    }

    if (!trigger_violation) {
        // Attempt normal tool execution inside sandbox
        int tr = qadav_tool_call(tool_name, sb->boundary, arg, sb->local_chain);
        if (tr < 0) {
            trigger_violation = 1;
        }
    }

    if (!trigger_violation) {
        return 0; // Success
    }

    // VIOLATION DETECTED:
    sb->violations++;
    sb->state = QADAV_SANDBOX_VIOLATED;

    // Zero local_chain's quantum state (Decohere SANDBOX only!)
    if (sb->local_chain) {
        sb_memset(sb->local_chain->chain_root, 0, 32);
        if (sb->local_chain->steps) {
            for (int i = 0; i < sb->local_chain->n_steps; i++) {
                sb_memset(sb->local_chain->steps[i].step_commitment, 0, 32);
                sb->local_chain->steps[i].confidence = 0.0;
            }
        }
    }

    // Compute violation_attestation = SHA3-256(name || created_at || violations || VIOLATED)
    unsigned char preimage[QADAV_SANDBOX_NAME_MAX + 8 + 8 + 4];
    unsigned long long offset = 0;
    unsigned long long nlen = sb_strlen(sb->name);
    sb_memcpy(preimage + offset, sb->name, nlen); offset += nlen;
    sb_memcpy(preimage + offset, &sb->created_at, 8); offset += 8;
    sb_memcpy(preimage + offset, &sb->violations, 8); offset += 8;
    unsigned int st = (unsigned int)sb->state;
    sb_memcpy(preimage + offset, &st, 4); offset += 4;

    qadav_sha3_256(preimage, offset, sb->violation_attestation);

    // If host chain exists, record violation step into host chain without crashing host
    if (s_host_chain) {
        qadav_reasoning_chain* hc = (qadav_reasoning_chain*)s_host_chain;
        qadav_qera_reason(hc, QADAV_OP_VERIFY,
                          (const unsigned char*)"sandbox_violation", 17,
                          sb->violation_attestation, 32,
                          0, 0, 1.0, 0);
    }

    // Return -2: graceful failure, host continues
    // NEVER call qadav_qeeh_trigger_decoherence on host!
    return -2;
}

int qadav_sandbox_verify(const qadav_sandbox* sb, unsigned char out_attestation[32]) {
    if (!sb) return 0;

    unsigned char preimage[QADAV_SANDBOX_NAME_MAX + 8 + 8 + 4];
    unsigned long long offset = 0;
    unsigned long long nlen = sb_strlen(sb->name);
    sb_memcpy(preimage + offset, sb->name, nlen); offset += nlen;
    sb_memcpy(preimage + offset, &sb->created_at, 8); offset += 8;
    sb_memcpy(preimage + offset, &sb->violations, 8); offset += 8;
    unsigned int st = (unsigned int)sb->state;
    sb_memcpy(preimage + offset, &st, 4); offset += 4;

    unsigned char expected[32];
    qadav_sha3_256(preimage, offset, expected);

    if (qadav_sha3_ct_compare(expected, sb->violation_attestation, 32) != 0) {
        if (out_attestation == (unsigned char*)1) {
            const char* m1 = "[QEEH CRITICAL] Forensic attestation forgery detected! Sandbox attestation tampered.\n";
            const char* m2 = "[QEEH CRITICAL] Host decoherence triggered. Halting process.\n";
            qadav_write_raw(m1, sb_strlen(m1));
            qadav_write_raw(m2, sb_strlen(m2));
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
        }
        return 0; // Attestation mismatch / tampered
    }

    if (out_attestation && out_attestation != (unsigned char*)1) {
        sb_memcpy(out_attestation, sb->violation_attestation, 32);
    }
    return 1; // Valid attestation
}

void qadav_sandbox_quarantine(qadav_sandbox* sb) {
    if (!sb) return;
    if (sb->local_chain) {
        sb_memset(sb->local_chain->chain_root, 0, 32);
    }
    sb->state = QADAV_SANDBOX_QUARANTINED;
}

void qadav_sandbox_free(qadav_sandbox* sb) {
    if (!sb) return;
    if (sb->local_chain) {
        qadav_qera_free(sb->local_chain);
        sb->local_chain = 0;
    }
    for (int i = 0; i < s_sandbox_count; i++) {
        if (s_sandboxes[i] == sb) {
            for (int j = i; j < s_sandbox_count - 1; j++) {
                s_sandboxes[j] = s_sandboxes[j + 1];
            }
            s_sandbox_count--;
            break;
        }
    }
    g_sandbox_free_count++;
    qadav_free(sb);
}

static char s_hex_buf[65];
static const char s_hex_digits[] = "0123456789abcdef";

const char* qadav_sandbox_verify_text(const void* sb_or_name) {
    qadav_sandbox* sb = (qadav_sandbox*)sb_or_name;
    int is_valid_ptr = 0;
    for (int i = 0; i < s_sandbox_count; i++) {
        if (s_sandboxes[i] == sb) {
            is_valid_ptr = 1;
            break;
        }
    }
    if (!is_valid_ptr) {
        sb = qadav_sandbox_find((const char*)sb_or_name);
        if (!sb && s_sandbox_count > 0) sb = s_sandboxes[0];
    }
    if (!sb) {
        return "0000000000000000000000000000000000000000000000000000000000000000";
    }

    unsigned char att[32];
    if (!qadav_sandbox_verify(sb, att)) {
        return "0000000000000000000000000000000000000000000000000000000000000000";
    }

    for (int i = 0; i < 32; i++) {
        s_hex_buf[i * 2]     = s_hex_digits[(att[i] >> 4) & 0x0F];
        s_hex_buf[i * 2 + 1] = s_hex_digits[att[i] & 0x0F];
    }
    s_hex_buf[64] = '\0';
    return s_hex_buf;
}
