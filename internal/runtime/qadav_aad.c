#include "qadav_aad.h"
#include "qadav_sha3.h"
#include "qadav_qera.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
int   qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len);
void  qadav_qeeh_trigger_decoherence(void* boundary);

static void aad_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void aad_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static int aad_popcount_byte(unsigned char b) {
    int c = 0;
    while (b) {
        c += (b & 1);
        b >>= 1;
    }
    return c;
}

static void aad_pack_u64(unsigned char* buf, unsigned long long val) {
    for (int i = 0; i < 8; i++) {
        buf[i] = (unsigned char)((val >> (i * 8)) & 0xFF);
    }
}

static void aad_pack_u32(unsigned char* buf, unsigned int val) {
    buf[0] = (unsigned char)(val & 0xFF);
    buf[1] = (unsigned char)((val >> 8) & 0xFF);
    buf[2] = (unsigned char)((val >> 16) & 0xFF);
    buf[3] = (unsigned char)((val >> 24) & 0xFF);
}

static void aad_compute_cert_hash(const aad_training_cert* cert, unsigned char out[32]) {
    // model_hash (32) || train_set_hash (32) || train_size (8) || max_syn_weight (8) || max_corr_dist (8) = 88 bytes
    unsigned char preimage[88];
    aad_memcpy(preimage + 0, cert->model_hash, 32);
    aad_memcpy(preimage + 32, cert->train_set_hash, 32);
    aad_pack_u64(preimage + 64, cert->train_size);
    aad_pack_u64(preimage + 72, cert->max_syn_weight);
    aad_pack_u64(preimage + 80, cert->max_corr_dist);
    qadav_sha3_256(preimage, 88, out);
}

static void aad_compute_envelope_hash(const aad_decode_envelope* env, unsigned char out[32]) {
    // step_index (8) || syndrome_hash (32) || syndrome_weight (4) || n_corrections (4) ||
    // correction_distance (4) || confidence (8) || training_cert_hash (32) = 92 bytes
    unsigned char preimage[92];
    aad_pack_u64(preimage + 0, env->step_index);
    aad_memcpy(preimage + 8, env->syndrome_hash, 32);
    aad_pack_u32(preimage + 40, env->syndrome_weight);
    aad_pack_u32(preimage + 44, env->n_corrections);
    aad_pack_u32(preimage + 48, env->correction_distance);
    aad_memcpy(preimage + 52, &env->confidence, 8);
    aad_memcpy(preimage + 60, env->training_cert_hash, 32);
    qadav_sha3_256(preimage, 92, out);
}

aad_training_cert* aad_build_training_cert(
    const unsigned char* model_bytes, unsigned long long model_len,
    const unsigned char* train_bytes, unsigned long long train_len,
    unsigned long long n_train,
    unsigned long long max_syndrome_weight,
    unsigned long long max_correction_distance
) {
    if (!model_bytes || model_len == 0 || !train_bytes || train_len == 0) {
        return 0;
    }
    if (n_train > QADAV_AAD_MAX_TRAIN_SIZE ||
        max_syndrome_weight > QADAV_AAD_MAX_SYN_WEIGHT ||
        max_correction_distance > QADAV_AAD_MAX_CORR_DIST) {
        return 0;
    }

    aad_training_cert* cert = (aad_training_cert*)qadav_alloc(sizeof(aad_training_cert));
    if (!cert) return 0;
    aad_memset(cert, 0, sizeof(aad_training_cert));

    qadav_sha3_256(model_bytes, model_len, cert->model_hash);
    qadav_sha3_256(train_bytes, train_len, cert->train_set_hash);
    cert->train_size = n_train;
    cert->max_syn_weight = max_syndrome_weight;
    cert->max_corr_dist = max_correction_distance;

    aad_compute_cert_hash(cert, cert->cert_hash);
    return cert;
}

int aad_wrap_decode(
    const aad_training_cert* cert,
    const qecd_decode_result* decode,
    const unsigned char* raw_syndrome,
    int syndrome_len,
    aad_decode_envelope* out_env
) {
    if (!cert || !decode || !out_env) return -1;

    // Calculate syndrome Hamming weight
    unsigned int weight = 0;
    if (raw_syndrome && syndrome_len > 0) {
        for (int i = 0; i < syndrome_len; i++) {
            weight += (unsigned int)aad_popcount_byte(raw_syndrome[i]);
        }
    }

    // Operating Envelope Bounds Enforcement:
    // If input syndrome weight > declared max_syn_weight -> abstain (return -3, NO decoherence)
    if (weight > (unsigned int)cert->max_syn_weight) {
        return -3;
    }

    // If correction distance > declared max_corr_dist -> abstain (return -3, NO decoherence)
    if ((unsigned int)decode->n_corrections > (unsigned int)cert->max_corr_dist) {
        return -3;
    }

    aad_memset(out_env, 0, sizeof(aad_decode_envelope));
    out_env->step_index = 1;

    if (raw_syndrome && syndrome_len > 0) {
        qadav_sha3_256(raw_syndrome, (unsigned long long)syndrome_len, out_env->syndrome_hash);
    } else {
        aad_memset(out_env->syndrome_hash, 0, 32);
    }

    out_env->syndrome_weight = weight;
    out_env->n_corrections = (unsigned int)decode->n_corrections;
    out_env->correction_distance = (unsigned int)decode->n_corrections;

    double conf = 1.0;
    if (cert->max_syn_weight > 0) {
        conf = 1.0 - ((double)weight / (double)(cert->max_syn_weight + 1));
    }
    if (conf < 0.0) conf = 0.0;
    if (conf > 1.0) conf = 1.0;
    out_env->confidence = conf;

    aad_memcpy(out_env->training_cert_hash, cert->cert_hash, 32);
    aad_compute_envelope_hash(out_env, out_env->envelope_hash);

    return 0;
}

int aad_verify_envelope(const aad_decode_envelope* env, const aad_training_cert* cert) {
    if (!env) return 0;

    unsigned char expected_env_hash[32];
    aad_compute_envelope_hash(env, expected_env_hash);
    if (qadav_sha3_ct_compare(env->envelope_hash, expected_env_hash, 32) != 0) {
        return 0;
    }

    if (cert) {
        if (qadav_sha3_ct_compare(env->training_cert_hash, cert->cert_hash, 32) != 0) {
            return 0;
        }
        if (env->syndrome_weight > (unsigned int)cert->max_syn_weight ||
            env->n_corrections > (unsigned int)cert->max_corr_dist) {
            return 0;
        }
    }

    return 1;
}

int aad_attest_envelope(aad_decode_envelope* env, qadav_reasoning_chain* chain) {
    if (!env) return 0;

    if (!aad_verify_envelope(env, 0)) {
        // Active Quantum Decoherence Nullification
        qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
        ExitProcess(0xDEADBEEF);
#endif
        return 0;
    }

    if (chain) {
        int r = qadav_qera_reason(chain, 6 /* QADAV_OP_AAD_DECODE */,
                                  env->syndrome_hash, 32,
                                  env->envelope_hash, 32,
                                  env->training_cert_hash, 32,
                                  env->confidence, 0);
        if (r != 0) return 0;
    }

    return 1;
}

int aad_meta_fit(aad_meta_state* state, const aad_training_cert* cert) {
    if (!state || !cert) return -1;
    aad_memset(state, 0, sizeof(aad_meta_state));
    state->base_cert = *cert;
    state->adaptation_steps = 0;
    aad_memcpy(state->adaptation_chain, cert->cert_hash, 32);
    return 0;
}

int aad_meta_adapt(aad_meta_state* state, const aad_decode_envelope* env) {
    if (!state || !env) return -1;

    // Verify envelope against base cert: if invalid -> return -2 (no decoherence at adapt)
    if (!aad_verify_envelope(env, &state->base_cert)) {
        return -2;
    }

    // Cumulative lineage update: chain = SHA3(prev_chain || env->envelope_hash)
    unsigned char chain_in[64];
    aad_memcpy(chain_in, state->adaptation_chain, 32);
    aad_memcpy(chain_in + 32, env->envelope_hash, 32);
    qadav_sha3_256(chain_in, 64, state->adaptation_chain);

    state->last_envelope = *env;
    state->adaptation_steps++;
    return 0;
}

int aad_meta_verify(const aad_meta_state* state) {
    if (!state) return 0;

    // Verify base cert
    unsigned char expected_cert_hash[32];
    aad_compute_cert_hash(&state->base_cert, expected_cert_hash);
    if (qadav_sha3_ct_compare(state->base_cert.cert_hash, expected_cert_hash, 32) != 0) {
        return 0;
    }

    // If adapt occurred, verify last envelope
    if (state->adaptation_steps > 0) {
        unsigned char expected_env_hash[32];
        aad_compute_envelope_hash(&state->last_envelope, expected_env_hash);
        if (qadav_sha3_ct_compare(state->last_envelope.envelope_hash, expected_env_hash, 32) != 0) {
            return 0;
        }
    }

    return 1;
}

int aad_corrupt_envelope(aad_decode_envelope* env, int offset) {
    if (!env) return -1;
    unsigned int off = (unsigned int)(offset >= 0 ? offset : -offset);
    ((unsigned char*)env)[off % sizeof(aad_decode_envelope)] ^= 0xFF;
    return 0;
}

void aad_free_training_cert(aad_training_cert* cert) {
    if (cert) {
        qadav_free(cert);
    }
}

void aad_free_meta_state(aad_meta_state* state) {
    if (state) {
        qadav_free(state);
    }
}
