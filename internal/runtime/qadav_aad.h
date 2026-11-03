#ifndef QADAV_AAD_H
#define QADAV_AAD_H

#include "qadav_qecdecoder.h"
#include "qadav_sea.h"
#include "qadav_eea.h"

#define QADAV_AAD_MAX_TRAIN_SIZE 1024
#define QADAV_AAD_MAX_SYN_WEIGHT 128
#define QADAV_AAD_MAX_CORR_DIST  16

typedef struct aad_training_cert {
    unsigned char      model_hash[32];      // hash of QEC decoder model
    unsigned char      train_set_hash[32];  // hash of training syndromes
    unsigned long long train_size;          // <= 1024
    unsigned long long max_syn_weight;      // <= 128
    unsigned long long max_corr_dist;       // <= 16
    unsigned char      cert_hash[32];       // SHA3(all above)
} aad_training_cert;

typedef struct aad_decode_envelope {
    unsigned long long step_index;
    unsigned char      syndrome_hash[32];
    unsigned int       syndrome_weight;
    unsigned int       n_corrections;
    unsigned int       correction_distance;
    double             confidence;
    unsigned char      training_cert_hash[32];
    unsigned char      envelope_hash[32];
} aad_decode_envelope;

typedef struct aad_meta_state {
    aad_training_cert   base_cert;
    aad_decode_envelope last_envelope;
    unsigned long long  adaptation_steps;
    unsigned char       adaptation_chain[32];   // SEA-style lineage
} aad_meta_state;

aad_training_cert* aad_build_training_cert(
    const unsigned char* model_bytes, unsigned long long model_len,
    const unsigned char* train_bytes, unsigned long long train_len,
    unsigned long long n_train,
    unsigned long long max_syndrome_weight,
    unsigned long long max_correction_distance
);

int aad_wrap_decode(
    const aad_training_cert* cert,
    const qecd_decode_result* decode,
    const unsigned char* raw_syndrome,
    int syndrome_len,
    aad_decode_envelope* out_env
);

int aad_verify_envelope(const aad_decode_envelope* env, const aad_training_cert* cert);
int aad_attest_envelope(aad_decode_envelope* env, qadav_reasoning_chain* chain);
int aad_meta_fit(aad_meta_state* state, const aad_training_cert* cert);
int aad_meta_adapt(aad_meta_state* state, const aad_decode_envelope* env);
int aad_meta_verify(const aad_meta_state* state);
int aad_corrupt_envelope(aad_decode_envelope* env, int offset);

void aad_free_training_cert(aad_training_cert* cert);
void aad_free_meta_state(aad_meta_state* state);

#endif
