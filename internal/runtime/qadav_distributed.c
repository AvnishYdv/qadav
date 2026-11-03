#include "qadav_distributed.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static unsigned long long g_dist_worker_count = 0;

static void dist_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

qadav_dist_ctx* qadav_dist_init(int n_workers) {
    if (n_workers < 1) n_workers = 1;
    if (n_workers > QADAV_MAX_WORKERS) n_workers = QADAV_MAX_WORKERS;

    qadav_dist_ctx* ctx = (qadav_dist_ctx*)qadav_alloc(sizeof(qadav_dist_ctx));
    if (!ctx) return 0;

    dist_memset(ctx, 0, sizeof(qadav_dist_ctx));
    ctx->worker_count = n_workers;
    ctx->single_worker_mode = (n_workers == 1) ? 1 : 0;
    ctx->step_index = 0;
    g_dist_worker_count = (unsigned long long)n_workers;

    // Generate worker UUIDs and compute initial topology hash
    unsigned char uuids[QADAV_MAX_WORKERS][16];
    for (int i = 0; i < n_workers; i++) {
        ctx->workers[i].id = i;
        ctx->workers[i].active = 1;
        ctx->workers[i].grad_size = 0;

        uuids[i][0] = 0xAA;
        uuids[i][1] = 0x55;
        uuids[i][2] = (unsigned char)(i & 0xFF);
        for (int k = 3; k < 16; k++) {
            uuids[i][k] = (unsigned char)((i * 17 + k * 13) & 0xFF);
        }
    }

    // Topo hash = SHA3-256 over all worker UUIDs
    qadav_sha3_256((const unsigned char*)uuids, (unsigned long long)(n_workers * 16), ctx->topo_hash);

    return ctx;
}

int qadav_dist_ring_allreduce(qadav_dist_ctx* ctx) {
    if (!ctx || ctx->worker_count <= 0) return -1;

    // Single-worker mode is a graceful no-op
    if (ctx->single_worker_mode || ctx->worker_count == 1) {
        return 0;
    }

    int n = ctx->worker_count;
    unsigned long long size = ctx->workers[0].grad_size;
    if (size == 0) return 0;

    // Verify all active workers have consistent gradient sizes
    for (int i = 1; i < n; i++) {
        if (ctx->workers[i].active && ctx->workers[i].grad_size != size) {
            return -2;
        }
    }

    // High-precision float64 ring reduction if 8-byte aligned
    if (size % 8 == 0) {
        unsigned long long num_floats = size / 8;
        // In-process shared memory reduction: sum into temp buffer then distribute
        for (unsigned long long f = 0; f < num_floats; f++) {
            double sum = 0.0;
            int active_count = 0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    double* fptr = (double*)ctx->workers[w].grad_buf;
                    sum += fptr[f];
                    active_count++;
                }
            }
            double avg = (active_count > 0) ? (sum / (double)active_count) : 0.0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    double* fptr = (double*)ctx->workers[w].grad_buf;
                    fptr[f] = avg;
                }
            }
        }
    } else {
        // Byte-wise average
        for (unsigned long long b = 0; b < size; b++) {
            unsigned int sum = 0;
            int active_count = 0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    sum += (unsigned int)ctx->workers[w].grad_buf[b];
                    active_count++;
                }
            }
            unsigned char avg = (active_count > 0) ? (unsigned char)(sum / active_count) : 0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    ctx->workers[w].grad_buf[b] = avg;
                }
            }
        }
    }

    // Commit cryptographic attestation for each worker
    ctx->step_index++;
    for (int w = 0; w < n; w++) {
        if (ctx->workers[w].active) {
            qadav_sha3_256(ctx->workers[w].grad_buf, size, ctx->workers[w].last_attestation);
        }
    }

    return 0;
}

int qadav_dist_pserver_aggregate(qadav_dist_ctx* ctx) {
    if (!ctx || ctx->worker_count <= 0) return -1;
    if (ctx->single_worker_mode || ctx->worker_count == 1) {
        return 0;
    }

    int n = ctx->worker_count;
    unsigned long long size = ctx->workers[0].grad_size;
    if (size == 0) return 0;

    if (size % 8 == 0) {
        unsigned long long num_floats = size / 8;
        for (unsigned long long f = 0; f < num_floats; f++) {
            double sum = 0.0;
            int active_count = 0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    double* fptr = (double*)ctx->workers[w].grad_buf;
                    sum += fptr[f];
                    active_count++;
                }
            }
            double avg = (active_count > 0) ? (sum / (double)active_count) : 0.0;
            // Master parameter server resides in worker[0]
            double* master_ptr = (double*)ctx->workers[0].grad_buf;
            master_ptr[f] = avg;
        }
    } else {
        for (unsigned long long b = 0; b < size; b++) {
            unsigned int sum = 0;
            int active_count = 0;
            for (int w = 0; w < n; w++) {
                if (ctx->workers[w].active) {
                    sum += (unsigned int)ctx->workers[w].grad_buf[b];
                    active_count++;
                }
            }
            unsigned char avg = (active_count > 0) ? (unsigned char)(sum / active_count) : 0;
            ctx->workers[0].grad_buf[b] = avg;
        }
    }

    // Attest parameter server aggregated weights
    ctx->step_index++;
    qadav_sha3_256(ctx->workers[0].grad_buf, size, ctx->workers[0].last_attestation);

    return 0;
}

void qadav_dist_shutdown(qadav_dist_ctx* ctx) {
    if (ctx) {
        qadav_free(ctx);
    }
}

unsigned long long qadav_dist_get_worker_count(void) {
    return g_dist_worker_count;
}
