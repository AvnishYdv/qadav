// Qadav tensor runtime — shared header (zero libc)
#ifndef QADAV_TENSOR_H
#define QADAV_TENSOR_H

// Element kinds (match Go types.Kind numeric values)
#define QADAV_ELEM_INT64    4    // types.KindInt64
#define QADAV_ELEM_FLOAT64  10   // types.KindFloat64

typedef struct qadav_tensor {
    void*   data;        // raw data pointer
    int     rank;        // 0 = scalar, 1 = vector, 2 = matrix
    int     shape[4];    // max rank 4 for MVP
    int     elem_kind;   // QADAV_ELEM_*
    int     elem_size;   // 8 for int64/float64
    int     total_size;  // total element count (product of shape)
} qadav_tensor;

// Memory allocation and low-level I/O implemented in qadav_rt.c / qadav_rt_linux.c
void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
void  qadav_write_raw(const char* s, unsigned long long len);

// Allocation & Lifetime
qadav_tensor* qadav_tensor_create(int elem_kind, int rank, const int* dims);
void          qadav_tensor_free(qadav_tensor* t);

// Element access
void          qadav_tensor_set_i64(qadav_tensor* t, int idx, long long val);
long long     qadav_tensor_get_i64(const qadav_tensor* t, int idx);
void          qadav_tensor_set_f64(qadav_tensor* t, int idx, double val);
double        qadav_tensor_get_f64(const qadav_tensor* t, int idx);

// Element-wise binary ops (returns new tensor)
qadav_tensor* qadav_tensor_add(const qadav_tensor* a, const qadav_tensor* b);
qadav_tensor* qadav_tensor_sub(const qadav_tensor* a, const qadav_tensor* b);
qadav_tensor* qadav_tensor_mul(const qadav_tensor* a, const qadav_tensor* b);
qadav_tensor* qadav_tensor_div(const qadav_tensor* a, const qadav_tensor* b);

// Scalar broadcast
qadav_tensor* qadav_tensor_add_scalar_f64(const qadav_tensor* t, double s);
qadav_tensor* qadav_tensor_mul_scalar_f64(const qadav_tensor* t, double s);
qadav_tensor* qadav_tensor_sub_scalar_f64(const qadav_tensor* t, double s);
qadav_tensor* qadav_tensor_div_scalar_f64(const qadav_tensor* t, double s);
qadav_tensor* qadav_tensor_add_scalar_i64(const qadav_tensor* t, long long s);
qadav_tensor* qadav_tensor_mul_scalar_i64(const qadav_tensor* t, long long s);
qadav_tensor* qadav_tensor_sub_scalar_i64(const qadav_tensor* t, long long s);
qadav_tensor* qadav_tensor_div_scalar_i64(const qadav_tensor* t, long long s);

// Matrix multiplication
qadav_tensor* qadav_tensor_matmul(const qadav_tensor* a, const qadav_tensor* b);

// Reductions
long long     qadav_tensor_reduce_sum_i64(const qadav_tensor* t);
double        qadav_tensor_reduce_sum_f64(const qadav_tensor* t);
long long     qadav_tensor_reduce_prod_i64(const qadav_tensor* t);
double        qadav_tensor_reduce_prod_f64(const qadav_tensor* t);

// Transpose (2D only)
qadav_tensor* qadav_tensor_transpose(const qadav_tensor* t);

// Print
void          qadav_tensor_print_i64(const qadav_tensor* t);
void          qadav_tensor_print_f64(const qadav_tensor* t);

#endif // QADAV_TENSOR_H

