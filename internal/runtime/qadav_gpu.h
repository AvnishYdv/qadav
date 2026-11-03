#ifndef QADAV_GPU_H
#define QADAV_GPU_H

#include "qadav_sha3.h"
#include "qadav_tensor.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct qadav_gpu_state {
    int                is_initialized;
    int                is_d3d_available;
    unsigned char      device_uuid[32];
    void*              d3d_module;
    void*              d3d_device;
    void*              d3d_context;
    unsigned long long dispatch_count;
    unsigned long long fallback_count;
} qadav_gpu_state;

int  qadav_gpu_init(void);
int  qadav_gpu_is_available(void);
void qadav_gpu_get_device_uuid(unsigned char out_uuid[32]);
int  qadav_gpu_matmul(const double* A, const double* B, double* C, int M, int K, int N);
int  qadav_gpu_softmax(const double* in, double* out, int rows, int cols);
int  qadav_gpu_layernorm(const double* in, double* out, int rows, int cols, double eps);
unsigned long long qadav_gpu_get_dispatch_count(void);
unsigned long long qadav_gpu_get_fallback_count(void);
void qadav_gpu_reset(void);
void qadav_gpu_shutdown(void);

qadav_tensor* qadav_gpu_tensor_matmul(const qadav_tensor* a, const qadav_tensor* b);
qadav_tensor* qadav_gpu_tensor_softmax(const qadav_tensor* a);
qadav_tensor* qadav_gpu_tensor_layernorm(const qadav_tensor* a, double eps);
int qadav_gpu_matmul_tensor_raw(qadav_tensor* a, qadav_tensor* b, qadav_tensor* dest, int m, int k, int n);
int qadav_gpu_softmax_tensor_raw(qadav_tensor* in, qadav_tensor* dest, int rows, int cols);
int qadav_gpu_layernorm_tensor_raw(qadav_tensor* in, qadav_tensor* dest, int rows, int cols, double eps);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_GPU_H */
