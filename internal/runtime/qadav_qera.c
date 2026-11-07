#include "qadav_qera.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);
void  qadav_write_raw(const char* s, unsigned long long len);
void  qadav_emit_int(long long v);

static void qera_memcpy(void* dst, const void* src, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    const volatile unsigned char* s = (const volatile unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qera_memset(void* dst, int v, unsigned long long n) {
    volatile unsigned char* d = (volatile unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

static unsigned long long qera_strlen(const char* s) {
    if (!s) return 0;
    unsigned long long len = 0;
    while (s[len]) len++;
    return len;
}

static void qera_print_hex8(const unsigned char* h) {
    const char hex[] = "0123456789abcdef";
    char buf[16];
    for (int i = 0; i < 8; i++) {
        buf[i * 2] = hex[(h[i] >> 4) & 0xF];
        buf[i * 2 + 1] = hex[h[i] & 0xF];
    }
    qadav_write_raw(buf, 16);
}

static void compute_step_commitment_bytes(const qadav_reasoning_step* step, unsigned char out[32]) {
    // Deterministic 185-byte packing (no struct padding / no ASLR):
    // step_index (8) || timestamp (8) || input_hash (32) || weights_hash (32) ||
    // ethics_hash (32) || output_hash (32) || operation (1) || confidence (8) || prev_commitment (32)
    unsigned char buf[185];
    qera_memcpy(buf, &step->step_index, 8);
    qera_memcpy(buf + 8, &step->timestamp, 8);
    qera_memcpy(buf + 16, step->input_hash, 32);
    qera_memcpy(buf + 48, step->weights_hash, 32);
    qera_memcpy(buf + 80, step->ethics_hash, 32);
    qera_memcpy(buf + 112, step->output_hash, 32);
    buf[144] = step->operation;
    qera_memcpy(buf + 145, &step->confidence, 8);
    qera_memcpy(buf + 153, step->prev_commitment, 32);

    qadav_sha3_256(buf, 185, out);
}

qadav_reasoning_chain* qadav_qera_create(const char* name, int capacity, void* auth_boundary) {
    qadav_reasoning_chain* chain = (qadav_reasoning_chain*)qadav_alloc(sizeof(qadav_reasoning_chain));
    if (!chain) return 0;
    qera_memset(chain, 0, sizeof(qadav_reasoning_chain));

    if (capacity <= 0 || capacity > QADAV_QERA_MAX_STEPS) {
        capacity = 100;
    }
    chain->capacity = capacity;
    chain->steps = (qadav_reasoning_step*)qadav_alloc((unsigned long long)capacity * sizeof(qadav_reasoning_step));
    if (!chain->steps) {
        qadav_free(chain);
        return 0;
    }
    qera_memset(chain->steps, 0, (unsigned long long)capacity * sizeof(qadav_reasoning_step));

    if (name && name[0] != '\0') {
        unsigned long long len = qera_strlen(name);
        if (len >= 64) len = 63;
        qera_memcpy(chain->chain_name, name, len);
        chain->chain_name[len] = '\0';
    } else {
        qera_memcpy(chain->chain_name, "clinical", 8);
        chain->chain_name[8] = '\0';
    }

    if (auth_boundary) {
        qadav_qeeh_boundary_hash(auth_boundary, chain->authorized_boundary_id);
    }

    return chain;
}

int qadav_qera_reason(qadav_reasoning_chain* chain,
                      int operation,
                      const unsigned char* input_data, unsigned long long input_len,
                      const unsigned char* output_data, unsigned long long output_len,
                      const unsigned char* weights_data, unsigned long long weights_len,
                      double confidence,
                      void* auth_boundary) {
    if (!chain || chain->n_steps >= chain->capacity) return -1;

    // Boundary authorization check
    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, chain->authorized_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -1;
        }
    }

    if (operation < 1 || operation > 6) return -2;
    if (confidence < 0.0 || confidence > 1.0) return -3;

    int idx = chain->n_steps;
    qadav_reasoning_step* step = &chain->steps[idx];

    step->step_index = (unsigned long long)idx;
    step->timestamp = (unsigned long long)idx;

    // F11: If input_len == 0 and input_data != NULL, compute strlen:
    if (input_data && input_len == 0) {
        input_len = qera_strlen((const char*)input_data);
    }
    if (output_data && output_len == 0) {
        output_len = qera_strlen((const char*)output_data);
    }

    // F11: Reject empty inputs (silent zeros नहीं)
    if (!input_data || input_len == 0) {
        return -5;  // Empty input error
    }
    if (!output_data || output_len == 0) {
        return -6;  // Empty output error
    }

    // F11: Verify hashes are computed
    qadav_sha3_256(input_data, input_len, step->input_hash);
    qadav_sha3_256(output_data, output_len, step->output_hash);

    // Sanity check: first 8 bytes of input_hash should NOT be all zero
    int first8_zero = 1;
    for (int i = 0; i < 8; i++) {
        if (step->input_hash[i] != 0) { first8_zero = 0; break; }
    }
    if (first8_zero) {
        return -7;  // Sanity check failed
    }

    if (weights_data && weights_len > 0) {
        qadav_sha3_256(weights_data, weights_len, step->weights_hash);
    } else {
        qera_memset(step->weights_hash, 0, 32);
    }

    qera_memcpy(step->ethics_hash, chain->authorized_boundary_id, 32);
    step->operation = (unsigned char)operation;
    step->confidence = confidence;

    if (idx == 0) {
        qera_memset(step->prev_commitment, 0, 32);
    } else {
        qera_memcpy(step->prev_commitment, chain->steps[idx - 1].step_commitment, 32);
    }

    compute_step_commitment_bytes(step, step->step_commitment);

    // Update rolling chain root
    unsigned char root_buf[64];
    qera_memcpy(root_buf, chain->chain_root, 32);
    qera_memcpy(root_buf + 32, step->step_commitment, 32);
    qadav_sha3_256(root_buf, 64, chain->chain_root);

    chain->n_steps++;
    return 0;
}

int qadav_qera_verify(qadav_reasoning_chain* chain, void* auth_boundary) {
    if (!chain) return 0;

    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, chain->authorized_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }
    }

    unsigned char local_root[32];
    qera_memset(local_root, 0, 32);

    for (int i = 0; i < chain->n_steps; i++) {
        qadav_reasoning_step* step = &chain->steps[i];

        // 1. Verify prev_commitment link
        if (i == 0) {
            for (int b = 0; b < 32; b++) {
                if (step->prev_commitment[b] != 0) {
                    qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
                    ExitProcess((unsigned int)0xDEADBEEF);
#endif
                    return 0;
                }
            }
        } else {
            if (qadav_sha3_ct_compare(step->prev_commitment, chain->steps[i - 1].step_commitment, 32) != 0) {
                qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
                ExitProcess((unsigned int)0xDEADBEEF);
#endif
                return 0;
            }
        }

        // 2. Recompute step_commitment
        unsigned char expected_commitment[32];
        compute_step_commitment_bytes(step, expected_commitment);
        if (qadav_sha3_ct_compare(expected_commitment, step->step_commitment, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return 0;
        }

        // 3. Update local rolling root
        unsigned char root_buf[64];
        qera_memcpy(root_buf, local_root, 32);
        qera_memcpy(root_buf + 32, step->step_commitment, 32);
        qadav_sha3_256(root_buf, 64, local_root);
    }

    // 4. Verify chain root matches
    if (qadav_sha3_ct_compare(local_root, chain->chain_root, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    chain->is_verified = 1;
    return 1;
}

static void qera_print_num(long long v) {
    if (v == 0) {
        qadav_write_raw("0", 1);
        return;
    }
    char tmp[32];
    char buf[32];
    int ti = 0, bi = 0;
    while (v > 0) {
        tmp[ti++] = '0' + (char)(v % 10);
        v /= 10;
    }
    while (ti > 0) buf[bi++] = tmp[--ti];
    qadav_write_raw(buf, bi);
}

int qadav_qera_audit(qadav_reasoning_chain* chain) {
    if (!chain) return -1;

    qadav_write_raw("--- QERA REASONING CHAIN AUDIT ---\n", 35);
    qadav_write_raw("Chain: ", 7);
    qadav_write_raw(chain->chain_name, qera_strlen(chain->chain_name));
    qadav_write_raw(" | Steps: ", 10);
    qera_print_num((long long)chain->n_steps);
    qadav_write_raw("\n", 1);

    const char* op_names[] = {"none", "deduce", "induce", "abduce", "verify", "evolve"};

    for (int i = 0; i < chain->n_steps; i++) {
        qadav_reasoning_step* s = &chain->steps[i];
        qadav_write_raw("  [", 3);
        qera_print_num((long long)i);
        qadav_write_raw("] op=", 5);
        int op = (int)s->operation;
        if (op >= 1 && op <= 5) {
            qadav_write_raw(op_names[op], qera_strlen(op_names[op]));
        } else {
            qadav_write_raw("unknown", 7);
        }

        qadav_write_raw(" in=", 4);
        qera_print_hex8(s->input_hash);
        qadav_write_raw(" out=", 5);
        qera_print_hex8(s->output_hash);
        qadav_write_raw(" commit=", 8);
        qera_print_hex8(s->step_commitment);
        qadav_write_raw("\n", 1);
    }

    qadav_write_raw("Chain Root: ", 12);
    qera_print_hex8(chain->chain_root);
    qadav_write_raw("\n", 1);
    qadav_write_raw("----------------------------------\n", 35);
    return 0;
}

int qadav_qera_corrupt_step(qadav_reasoning_chain* chain, int step_idx, int at_idx) {
    if (!chain || step_idx < 0 || step_idx >= chain->n_steps) return -1;
    chain->steps[step_idx].step_commitment[at_idx % 32] ^= 0xFF;
    return 0;
}

void qadav_qera_free(qadav_reasoning_chain* chain) {
    if (!chain) return;
    if (chain->steps) {
        qadav_free(chain->steps);
    }
    void qadav_certificate_free(qadav_reasoning_chain* chain);
    qadav_certificate_free(chain);
    qadav_free(chain);
}

