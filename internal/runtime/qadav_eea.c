#include "qadav_eea.h"
#include "qadav_sha3.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void  qadav_qeeh_trigger_decoherence(void* boundary);

static void eea_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void eea_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static int eea_popcount_byte(unsigned char b) {
    int c = 0;
    while (b) {
        c += (b & 1);
        b >>= 1;
    }
    return c;
}

eea_context* eea_create_policy(eea_policy_kind kind, double initial_rate, unsigned long long total_budget) {
    eea_context* ctx = (eea_context*)qadav_alloc(sizeof(eea_context));
    if (!ctx) return 0;
    eea_memset(ctx, 0, sizeof(eea_context));

    if (kind < POLICY_CONSTANT_RATE || kind > POLICY_ADAPTIVE) {
        kind = POLICY_CONSTANT_RATE;
    }
    if (initial_rate <= 0.0 || initial_rate > 1.0) {
        initial_rate = 0.10;
    }
    if (total_budget == 0) {
        total_budget = 1000;
    }

    ctx->kind = kind;
    ctx->mutation_rate = initial_rate;
    ctx->total_flips_budget = total_budget;
    ctx->total_flips_used = 0;
    ctx->current_step = 0;
    return ctx;
}

eea_envelope* eea_create_envelope(void) {
    eea_envelope* env = (eea_envelope*)qadav_alloc(sizeof(eea_envelope));
    if (env) {
        eea_memset(env, 0, sizeof(eea_envelope));
    }
    return env;
}

void eea_free_envelope(eea_envelope* env) {
    if (env) {
        qadav_free(env);
    }
}

int eea_compute_diff(const unsigned char* from_buf, unsigned long long from_len,
                     const unsigned char* to_buf, unsigned long long to_len,
                     unsigned long long* out_flips, unsigned long long* out_new_syms) {
    if (!out_flips || !out_new_syms) return -1;

    unsigned long long flips = 0;
    unsigned long long min_len = (from_len < to_len) ? from_len : to_len;

    for (unsigned long long i = 0; i < min_len; i++) {
        unsigned char diff = from_buf[i] ^ to_buf[i];
        if (diff) {
            flips += (unsigned long long)eea_popcount_byte(diff);
        }
    }

    if (from_len < to_len) {
        for (unsigned long long i = min_len; i < to_len; i++) {
            flips += (unsigned long long)eea_popcount_byte(to_buf[i]);
        }
    } else if (to_len < from_len) {
        for (unsigned long long i = min_len; i < from_len; i++) {
            flips += (unsigned long long)eea_popcount_byte(from_buf[i]);
        }
    }

    // New symbol detection (bytes present in to_buf that were not in from_buf)
    unsigned char from_seen[256];
    eea_memset(from_seen, 0, 256);
    for (unsigned long long i = 0; i < from_len; i++) {
        from_seen[from_buf[i]] = 1;
    }

    unsigned char to_seen[256];
    eea_memset(to_seen, 0, 256);
    unsigned long long new_syms = 0;
    for (unsigned long long i = 0; i < to_len; i++) {
        unsigned char b = to_buf[i];
        if (!from_seen[b] && !to_seen[b]) {
            new_syms++;
            to_seen[b] = 1;
        }
    }

    *out_flips = flips;
    *out_new_syms = new_syms;
    return 0;
}

static void eea_pack_preimage(const eea_envelope* env, unsigned char preimage[QADAV_EEA_PREIMAGE_LEN]) {
    // Deterministic 187-byte packing (no struct padding / zero ASLR variance):
    // step_index (8) || max_flips (8) || max_new (8) || actual_flips (8) || actual_new (8) ||
    // policy_kind (1) || policy_name (32) || mutation_rate (8) || from_state (32) ||
    // to_state (32) || proposal_hash (32)
    eea_memset(preimage, 0, QADAV_EEA_PREIMAGE_LEN);
    eea_memcpy(preimage + 0, &env->step_index, 8);
    eea_memcpy(preimage + 8, &env->max_bit_flips, 8);
    eea_memcpy(preimage + 16, &env->max_new_symbols, 8);
    eea_memcpy(preimage + 24, &env->actual_bit_flips, 8);
    eea_memcpy(preimage + 32, &env->actual_new_symbols, 8);
    preimage[40] = env->policy_kind;
    eea_memcpy(preimage + 41, env->policy_name, 32);
    eea_memcpy(preimage + 73, &env->mutation_rate, 8);
    eea_memcpy(preimage + 81, env->from_state_hash, 32);
    eea_memcpy(preimage + 113, env->to_state_hash, 32);
    eea_memcpy(preimage + 145, env->proposal_hash, 32);
}

int eea_build_envelope(eea_context* ctx, const sea_proposal* proposal,
                       const unsigned char* from_buf, unsigned long long from_len,
                       const unsigned char* to_buf, unsigned long long to_len,
                       eea_envelope* out_env) {
    if (!ctx || !out_env) return -1;

    unsigned long long flips = 0;
    unsigned long long new_syms = 0;
    int diff_res = eea_compute_diff(from_buf, from_len, to_buf, to_len, &flips, &new_syms);
    if (diff_res != 0) return -1;

    // Check declared bounds (if out_env has pre-declared budgets, enforce them)
    unsigned long long max_flips = out_env->max_bit_flips;
    unsigned long long max_syms = out_env->max_new_symbols;

    if (flips > max_flips || new_syms > max_syms) {
        return -2; // Bounded mutation violation (clean validation failure, no decoherence)
    }

    out_env->step_index = proposal ? proposal->step_index : (unsigned long long)(ctx->current_step + 1);
    out_env->max_bit_flips = max_flips;
    out_env->max_new_symbols = max_syms;
    out_env->actual_bit_flips = flips;
    out_env->actual_new_symbols = new_syms;
    out_env->policy_kind = (unsigned char)ctx->kind;

    eea_memset(out_env->policy_name, 0, QADAV_EEA_MAX_POLICY_NAME);
    const char* pname = "constant";
    if (ctx->kind == POLICY_DECAYING_RATE) pname = "decaying";
    else if (ctx->kind == POLICY_ADAPTIVE) pname = "adaptive";
    for (int i = 0; pname[i] && i < QADAV_EEA_MAX_POLICY_NAME - 1; i++) {
        out_env->policy_name[i] = pname[i];
    }

    out_env->mutation_rate = ctx->mutation_rate;

    if (proposal) {
        eea_memcpy(out_env->from_state_hash, proposal->from_state_hash, 32);
        eea_memcpy(out_env->to_state_hash, proposal->to_state_hash, 32);
        eea_memcpy(out_env->proposal_hash, proposal->proposal_hash, 32);
    } else {
        qadav_sha3_256(from_buf, from_len, out_env->from_state_hash);
        qadav_sha3_256(to_buf, to_len, out_env->to_state_hash);
        eea_memset(out_env->proposal_hash, 0, 32);
    }

    // Compute envelope_hash = SHA3-256(preimage)
    unsigned char preimage[QADAV_EEA_PREIMAGE_LEN];
    eea_pack_preimage(out_env, preimage);
    qadav_sha3_256(preimage, QADAV_EEA_PREIMAGE_LEN, out_env->envelope_hash);

    // Apply policy state updates
    ctx->total_flips_used += flips;
    if (ctx->kind == POLICY_DECAYING_RATE) {
        ctx->mutation_rate *= 0.95;
    }
    ctx->current_step++;
    return 0;
}

int eea_build_envelope_text(eea_context* ctx, const sea_proposal* proposal,
                            const char* from_text, const char* to_text,
                            unsigned long long max_flips, unsigned long long max_new_syms,
                            eea_envelope* out_env) {
    if (!out_env) return -1;
    if (!from_text) from_text = "";
    if (!to_text) to_text = "";
    unsigned long long flen = 0;
    while (from_text[flen]) flen++;
    unsigned long long tlen = 0;
    while (to_text[tlen]) tlen++;
    out_env->max_bit_flips = max_flips;
    out_env->max_new_symbols = max_new_syms;
    return eea_build_envelope(ctx, proposal, (const unsigned char*)from_text, flen, (const unsigned char*)to_text, tlen, out_env);
}

int eea_verify_envelope(const eea_envelope* env) {
    if (!env) return 0;

    // Check bounds
    if (env->actual_bit_flips > env->max_bit_flips ||
        env->actual_new_symbols > env->max_new_symbols) {
        return 0;
    }

    // Recompute envelope hash
    unsigned char preimage[QADAV_EEA_PREIMAGE_LEN];
    eea_pack_preimage(env, preimage);

    unsigned char expected[32];
    qadav_sha3_256(preimage, QADAV_EEA_PREIMAGE_LEN, expected);

    if (qadav_sha3_ct_compare(expected, env->envelope_hash, 32) != 0) {
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess((unsigned int)0xDEADBEEF);
#endif
        return 0;
    }

    return 1;
}

int eea_attest_envelope(const eea_envelope* env, qadav_reasoning_chain* chain) {
    if (!env) return 0;

    if (!eea_verify_envelope(env)) {
        return 0;
    }

    if (chain) {
        int r = qadav_qera_reason(
            chain,
            QADAV_OP_EVOLVE,
            env->from_state_hash, 32,
            env->to_state_hash, 32,
            env->envelope_hash, 32,
            1.0,
            0
        );
        if (r != 0) return 0;
    }

    return 1;
}

int eea_corrupt(eea_envelope* env, int offset) {
    if (!env) return -1;
    if (offset < 0 || offset >= (int)sizeof(eea_envelope)) {
        offset = 0;
    }
    unsigned char* b = (unsigned char*)env;
    b[offset] ^= 0xFF;
    return 0;
}

void eea_free(eea_context* ctx) {
    if (ctx) {
        qadav_free(ctx);
    }
}
