#ifndef QADAV_CDAGA_H
#define QADAV_CDAGA_H

#include "qadav_qera.h"
#include "qadav_multigpu.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_CDAGA_MAX_DEVICES 8

typedef struct qadav_cdaga_record {
    unsigned long long step_index;
    unsigned long long timestamp;
    int                device_count;
    unsigned char      device_uuids[QADAV_CDAGA_MAX_DEVICES][16];
    unsigned char      gradient_digests[QADAV_CDAGA_MAX_DEVICES][32];
    unsigned char      topology_hash[32];
    unsigned char      aggregation_digest[32];
    unsigned char      record_hash[32];
} qadav_cdaga_record;

int qadav_cdaga_record_aggregation(
    qadav_multigpu_context* ctx,
    unsigned long long step_index,
    void** gradient_buffers,
    unsigned long long* gradient_sizes,
    int buffer_count,
    qadav_cdaga_record* out
);

int qadav_cdaga_record_aggregation_with_uuids(
    const unsigned char uuids[QADAV_CDAGA_MAX_DEVICES][16],
    const unsigned char topology_hash[32],
    unsigned long long step_index,
    void** gradient_buffers,
    unsigned long long* gradient_sizes,
    int buffer_count,
    qadav_cdaga_record* out
);

int qadav_cdaga_verify(
    const qadav_cdaga_record* record,
    const qadav_multigpu_context* ctx
);

int qadav_cdaga_attest(
    qadav_reasoning_chain* chain,
    const qadav_cdaga_record* record,
    const qadav_multigpu_context* ctx
);

int qadav_cdaga_corrupt(
    qadav_cdaga_record* record,
    int offset
);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_CDAGA_H */

