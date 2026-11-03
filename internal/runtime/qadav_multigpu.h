#ifndef QADAV_MULTIGPU_H
#define QADAV_MULTIGPU_H

#include "qadav_sha3.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_MAX_DEVICES 8
#define QADAV_MAX_PIPELINE_STAGES 16

typedef struct qadav_device_info {
    unsigned int       index;
    char               name[128];
    unsigned int       vendor_id;
    unsigned int       device_id;
    unsigned long long vram_bytes;
    unsigned char      uuid[16];
    int                pipeline_rank;
} qadav_device_info;

typedef struct qadav_multigpu_context {
    qadav_device_info  devices[QADAV_MAX_DEVICES];
    int                device_count;
    int                stage_count;
    int                single_device_mode;
    unsigned char      topology_hash[32];
    void*              stage_buffers[QADAV_MAX_PIPELINE_STAGES];
} qadav_multigpu_context;

qadav_multigpu_context* qadav_multigpu_init(void);
int  qadav_multigpu_build_pipeline(qadav_multigpu_context* ctx, int total_layers, int n_stages, int micro_batch);
int  qadav_multigpu_run_pipeline(qadav_multigpu_context* ctx, void* input, void* output);
void qadav_multigpu_shutdown(qadav_multigpu_context* ctx);
unsigned long long qadav_multigpu_get_device_count(void);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_MULTIGPU_H */

