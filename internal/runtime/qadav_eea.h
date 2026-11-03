#ifndef QADAV_EEA_H
#define QADAV_EEA_H

#include "qadav_sea.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_EEA_MAX_POLICY_NAME 32
#define QADAV_EEA_MAX_REASON 64
#define QADAV_EEA_PREIMAGE_LEN 187

typedef enum {
    POLICY_CONSTANT_RATE = 0,
    POLICY_DECAYING_RATE = 1,
    POLICY_ADAPTIVE = 2
} eea_policy_kind;

typedef struct eea_envelope {
    unsigned long long step_index;
    unsigned long long max_bit_flips;      // declared budget
    unsigned long long max_new_symbols;    // declared budget
    unsigned long long actual_bit_flips;   // measured
    unsigned long long actual_new_symbols; // measured
    unsigned char      policy_kind;        // 0/1/2
    char               policy_name[QADAV_EEA_MAX_POLICY_NAME];
    double             mutation_rate;      // 0.0-1.0
    unsigned char      from_state_hash[32];
    unsigned char      to_state_hash[32];
    unsigned char      proposal_hash[32];
    unsigned char      envelope_hash[32];
} eea_envelope;

typedef struct eea_context {
    eea_policy_kind    kind;
    double             mutation_rate;
    unsigned long long total_flips_budget;
    unsigned long long total_flips_used;
    int                current_step;
} eea_context;

eea_context* eea_create_policy(eea_policy_kind kind, double initial_rate, unsigned long long total_budget);

eea_envelope* eea_create_envelope(void);

int eea_compute_diff(const unsigned char* from_buf, unsigned long long from_len,
                     const unsigned char* to_buf, unsigned long long to_len,
                     unsigned long long* out_flips, unsigned long long* out_new_syms);

int eea_build_envelope(eea_context* ctx, const sea_proposal* proposal,
                       const unsigned char* from_buf, unsigned long long from_len,
                       const unsigned char* to_buf, unsigned long long to_len,
                       eea_envelope* out_env);

int eea_build_envelope_text(eea_context* ctx, const sea_proposal* proposal,
                            const char* from_text, const char* to_text,
                            unsigned long long max_flips, unsigned long long max_new_syms,
                            eea_envelope* out_env);

int eea_verify_envelope(const eea_envelope* env);

int eea_attest_envelope(const eea_envelope* env, qadav_reasoning_chain* chain);

int eea_corrupt(eea_envelope* env, int offset);  // --test only

void eea_free_envelope(eea_envelope* env);

void eea_free(eea_context* ctx);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_EEA_H */
