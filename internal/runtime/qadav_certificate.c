#include "qadav_certificate.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

extern void* qadav_alloc(unsigned long long size);
extern void  qadav_free(void* ptr);

static unsigned long long g_cert_create_count = 0;
static unsigned long long g_cert_free_count = 0;

unsigned long long qadav_cert_get_create_count(void) {
    return g_cert_create_count;
}

unsigned long long qadav_cert_get_free_count(void) {
    return g_cert_free_count;
}

void qadav_certificate_free(qadav_reasoning_chain* chain) {
    if (!chain) return;
    if (chain->certs) {
        g_cert_free_count += (unsigned long long)chain->n_steps;
        qadav_free(chain->certs);
        chain->certs = 0;
        chain->n_certs = 0;
    }
}

static void cert_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void cert_memset(void* dst, int v, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)v;
    }
}

int qadav_certificate_attach(
    qadav_reasoning_chain* chain,
    unsigned long long step_index,
    unsigned char kind,
    const unsigned char* witness, unsigned int witness_len,
    double confidence
) {
    if (!chain) return -1;
    if (kind < QADAV_CERT_DEDUCTIVE || kind > QADAV_CERT_VERIFICATION) return -1;
    if (witness_len < 1 || witness_len > QADAV_CERT_WITNESS_MAX || !witness) return -2;
    if (confidence < 0.0 || confidence > 1.0) return -3;
    if (step_index >= (unsigned long long)chain->n_steps) return -4;

    qadav_reasoning_step* step = &chain->steps[step_index];

    if (!chain->certs) {
        unsigned long long cap = (unsigned long long)chain->capacity;
        if (cap == 0) cap = 100;
        chain->certs = qadav_alloc(cap * sizeof(qadav_reason_certificate));
        if (!chain->certs) return -1;
        cert_memset(chain->certs, 0, cap * sizeof(qadav_reason_certificate));
        chain->n_certs = (int)cap;
    }

    qadav_reason_certificate* cert = &((qadav_reason_certificate*)chain->certs)[step_index];
    cert->kind = kind;
    cert_memcpy(cert->premise_hash, step->input_hash, 32);
    cert_memcpy(cert->conclusion_hash, step->output_hash, 32);
    cert->witness_len = witness_len;
    cert_memcpy(cert->witness, witness, witness_len);
    cert->confidence = confidence;

    // cert_id = SHA3-256(kind (1) || premise_hash (32) || conclusion_hash (32) || witness (witness_len) || witness_len (4) || confidence (8))
    unsigned char id_preimage[77 + QADAV_CERT_WITNESS_MAX];
    id_preimage[0] = kind;
    cert_memcpy(id_preimage + 1, cert->premise_hash, 32);
    cert_memcpy(id_preimage + 33, cert->conclusion_hash, 32);
    cert_memcpy(id_preimage + 65, cert->witness, witness_len);
    cert_memcpy(id_preimage + 65 + witness_len, &witness_len, 4);
    cert_memcpy(id_preimage + 69 + witness_len, &confidence, 8);

    unsigned long long id_len = 77 + (unsigned long long)witness_len;
    qadav_sha3_256(id_preimage, id_len, cert->cert_id);

    // commitment = SHA3-256(cert_id (32) || id_preimage)
    unsigned char commit_preimage[32 + 77 + QADAV_CERT_WITNESS_MAX];
    cert_memcpy(commit_preimage, cert->cert_id, 32);
    cert_memcpy(commit_preimage + 32, id_preimage, id_len);

    qadav_sha3_256(commit_preimage, 32 + id_len, cert->commitment);

    g_cert_create_count++;
    return 0;
}

int qadav_certificate_verify(const qadav_reason_certificate* cert) {
    if (!cert) return -1;

    // 1. Kind validation: {1, 2, 3, 4}
    if (cert->kind < QADAV_CERT_DEDUCTIVE || cert->kind > QADAV_CERT_VERIFICATION) return -1;

    // 2. Witness length validation: [1, 512]
    if (cert->witness_len < 1 || cert->witness_len > QADAV_CERT_WITNESS_MAX) return -2;

    // 3. Confidence range validation: [0.0, 1.0]
    if (cert->confidence < 0.0 || cert->confidence > 1.0) return -3;

    // 4. Check premise_hash and conclusion_hash non-zero (F11 pattern)
    int p_zero = 1, c_zero = 1;
    for (int i = 0; i < 32; i++) {
        if (cert->premise_hash[i] != 0) p_zero = 0;
        if (cert->conclusion_hash[i] != 0) c_zero = 0;
    }
    if (p_zero || c_zero) return -4;

    // 5. Recompute cert_id and compare in constant time
    unsigned char id_preimage[77 + QADAV_CERT_WITNESS_MAX];
    id_preimage[0] = cert->kind;
    cert_memcpy(id_preimage + 1, cert->premise_hash, 32);
    cert_memcpy(id_preimage + 33, cert->conclusion_hash, 32);
    cert_memcpy(id_preimage + 65, cert->witness, cert->witness_len);
    cert_memcpy(id_preimage + 65 + cert->witness_len, &cert->witness_len, 4);
    cert_memcpy(id_preimage + 69 + cert->witness_len, &cert->confidence, 8);

    unsigned long long id_len = 77 + (unsigned long long)cert->witness_len;
    unsigned char expected_id[32];
    qadav_sha3_256(id_preimage, id_len, expected_id);

    if (qadav_sha3_ct_compare(expected_id, cert->cert_id, 32) != 0) {
        return -5;
    }

    // 6. Recompute commitment and compare in constant time
    unsigned char commit_preimage[32 + 77 + QADAV_CERT_WITNESS_MAX];
    cert_memcpy(commit_preimage, cert->cert_id, 32);
    cert_memcpy(commit_preimage + 32, id_preimage, id_len);

    unsigned char expected_commitment[32];
    qadav_sha3_256(commit_preimage, 32 + id_len, expected_commitment);

    if (qadav_sha3_ct_compare(expected_commitment, cert->commitment, 32) != 0) {
        return -6;
    }

    return 0; // Structurally valid
}

int qadav_certificate_verify_chain(qadav_reasoning_chain* chain) {
    if (!chain) return -1;

    // Verify underlying QERA cryptographic integrity (returns 1 on success)
    if (!qadav_qera_verify(chain, 0)) return -9;

    if (!chain->certs) return 0; // Pure QERA chain without certificates, verified

    qadav_reason_certificate* certs = (qadav_reason_certificate*)chain->certs;
    for (int i = 0; i < chain->n_steps; i++) {
        int vr = qadav_certificate_verify(&certs[i]);
        if (vr != 0) {
#ifdef _WIN32
            qadav_qeeh_trigger_decoherence(0);
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return vr;
        }

        if (qadav_sha3_ct_compare(certs[i].premise_hash, chain->steps[i].input_hash, 32) != 0) {
#ifdef _WIN32
            qadav_qeeh_trigger_decoherence(0);
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -7;
        }
        if (qadav_sha3_ct_compare(certs[i].conclusion_hash, chain->steps[i].output_hash, 32) != 0) {
#ifdef _WIN32
            qadav_qeeh_trigger_decoherence(0);
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -8;
        }
    }

    return 0;
}

int qadav_certificate_corrupt(qadav_reasoning_chain* chain, int cert_idx, int at_idx) {
    if (!chain || !chain->certs || cert_idx < 0 || cert_idx >= chain->n_steps) return -1;
    qadav_reason_certificate* certs = (qadav_reason_certificate*)chain->certs;
    certs[cert_idx].witness[at_idx % QADAV_CERT_WITNESS_MAX] ^= 0xFF;
    return 0;
}
