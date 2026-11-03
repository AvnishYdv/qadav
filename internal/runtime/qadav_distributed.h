#ifndef QADAV_DISTRIBUTED_H
#define QADAV_DISTRIBUTED_H

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_MAX_WORKERS 16
#define QADAV_GRAD_BUF_MAX (1024 * 1024 * 16)  // 16MB

typedef struct qadav_worker {
    int                id;
    unsigned char      grad_buf[QADAV_GRAD_BUF_MAX];
    unsigned long long grad_size;
    int                active;
    unsigned char      last_attestation[32];
} qadav_worker;

typedef struct qadav_dist_ctx {
    qadav_worker       workers[QADAV_MAX_WORKERS];
    int                worker_count;
    int                single_worker_mode;
    unsigned char      topo_hash[32];
    unsigned long long step_index;
} qadav_dist_ctx;

qadav_dist_ctx* qadav_dist_init(int n_workers);
int  qadav_dist_ring_allreduce(qadav_dist_ctx* ctx);
int  qadav_dist_pserver_aggregate(qadav_dist_ctx* ctx);
void qadav_dist_shutdown(qadav_dist_ctx* ctx);
unsigned long long qadav_dist_get_worker_count(void);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_DISTRIBUTED_H */
