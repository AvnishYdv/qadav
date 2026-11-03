#ifndef QADAV_SEA_H
#define QADAV_SEA_H

#include "qadav_qera.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_SEA_MAX_HISTORY 64

typedef struct sea_proposal {
    unsigned long long step_index;
    unsigned long long timestamp;
    unsigned char from_state_hash[32];
    unsigned char to_state_hash[32];
    unsigned char reason_hash[32];
    unsigned char prev_proposal_hash[32];
    unsigned char proposal_hash[32];
} sea_proposal;

typedef struct sea_history {
    int            count;
    sea_proposal   proposals[QADAV_SEA_MAX_HISTORY];
    unsigned char  lineage_root[32];
} sea_history;

sea_history* sea_create(void);
int  sea_propose(sea_history* h, const unsigned char from[32], const unsigned char to[32], const unsigned char reason[32], sea_proposal** out);
int  sea_verify_proposal(const sea_proposal* p, const sea_proposal* prev);
int  sea_attest(sea_history* h, qadav_reasoning_chain* chain);
int  sea_verify_history(const sea_history* h);
int  sea_corrupt(sea_proposal* p, int offset);  // --test only
void sea_hash_text(const char* text, unsigned char out[32]);
void sea_free(sea_history* h);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_SEA_H */
