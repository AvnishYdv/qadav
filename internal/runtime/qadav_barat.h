#ifndef QADAV_BARAT_H
#define QADAV_BARAT_H

#include "qadav_qera.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_BARAT_MAX_WORKERS 16
#define QADAV_BARAT_GRAD_MAX (1024 * 1024 * 16) // 16MB max gradient tensor

typedef struct qadav_barat_submission {
    int                worker_id;
    unsigned char      worker_uuid[16];
    unsigned char      digest[32];
    void*              data;
    unsigned long long size;
} qadav_barat_submission;

typedef struct qadav_barat_cert {
    unsigned long long step_index;
    unsigned long long timestamp;
    int                worker_count;
    unsigned int       inclusion_mask;
    double             trimmed_alpha; // default 0.20
    unsigned char      worker_uuids[QADAV_BARAT_MAX_WORKERS][16];
    unsigned char      gradient_digests[QADAV_BARAT_MAX_WORKERS][32];
    unsigned char      selection_proof[32];
    unsigned char      aggregated_hash[32];
    unsigned char      prev_record_hash[32];
    unsigned char      record_hash[32];
} qadav_barat_cert;

typedef struct qadav_barat_group {
    int                    group_id;
    int                    submission_count;
    qadav_barat_submission submissions[QADAV_BARAT_MAX_WORKERS];
    unsigned char          last_record_hash[32];
    unsigned long long     step_index;
} qadav_barat_group;

qadav_barat_group* qadav_barat_create_group(int group_id);
void qadav_barat_destroy_group(qadav_barat_group* grp);

int qadav_barat_submit(
    qadav_barat_group* grp,
    int worker_id,
    const void* grad_data,
    unsigned long long grad_size
);

int qadav_barat_aggregate(
    qadav_barat_group* grp,
    void* out_tensor,
    unsigned long long out_size,
    qadav_barat_cert* out_cert
);

int qadav_barat_verify(
    const qadav_barat_cert* cert,
    const void* aggregated_tensor,
    unsigned long long tensor_size
);

int qadav_barat_attest(
    qadav_reasoning_chain* chain,
    const qadav_barat_cert* cert,
    const void* aggregated_tensor,
    unsigned long long tensor_size
);

int qadav_barat_corrupt(
    qadav_barat_cert* cert,
    int offset
);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_BARAT_H */

