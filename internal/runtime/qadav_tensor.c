// Qadav tensor runtime — portable implementation (zero libc)
#include "qadav_tensor.h"

static int int_to_str_local(long long v, char* buf) {
    int i = 0;
    if (v == 0) {
        buf[i++] = '0';
        return i;
    }
    int is_neg = 0;
    unsigned long long uv;
    if (v < 0) {
        is_neg = 1;
        uv = (unsigned long long)(-v);
    } else {
        uv = (unsigned long long)v;
    }
    char temp[32];
    int t = 0;
    while (uv > 0) {
        temp[t++] = '0' + (char)(uv % 10);
        uv /= 10;
    }
    if (is_neg) {
        buf[i++] = '-';
    }
    while (t > 0) {
        buf[i++] = temp[--t];
    }
    return i;
}

static void print_str(const char* s) {
    if (!s) return;
    unsigned long long len = 0;
    while (s[len]) len++;
    qadav_write_raw(s, len);
}

static void print_i64_raw(long long v) {
    char buf[32];
    int n = int_to_str_local(v, buf);
    qadav_write_raw(buf, n);
}

static void print_f64_raw(double v) {
    char buf[64];
    int idx = 0;
    if (v < 0.0) {
        buf[idx++] = '-';
        v = -v;
    }
    long long ipart = (long long)v;
    double frac = v - (double)ipart;
    long long fpart = (long long)(frac * 1000000.0 + 0.5);
    if (fpart >= 1000000) {
        ipart += 1;
        fpart -= 1000000;
    }
    idx += int_to_str_local(ipart, buf + idx);
    buf[idx++] = '.';
    char fbuf[16];
    int fn = int_to_str_local(fpart, fbuf);
    for (int p = 0; p < 6 - fn; p++) {
        buf[idx++] = '0';
    }
    for (int p = 0; p < fn; p++) {
        buf[idx++] = fbuf[p];
    }
    qadav_write_raw(buf, idx);
}

qadav_tensor* qadav_tensor_create(int elem_kind, int rank, const int* dims) {
    qadav_tensor* t = (qadav_tensor*)qadav_alloc(sizeof(qadav_tensor));
    if (!t) return (void*)0;

    t->rank = rank;
    t->elem_kind = elem_kind;
    t->elem_size = 8;

    int total = 1;
    for (int i = 0; i < rank && i < 4; i++) {
        t->shape[i] = dims[i];
        total *= dims[i];
    }
    if (rank == 0) {
        total = 1;
    }
    t->total_size = total;
    t->data = qadav_alloc((unsigned long long)total * 8);
    return t;
}

void qadav_tensor_free(qadav_tensor* t) {
    if (!t) return;
    if (t->data) {
        qadav_free(t->data);
    }
    qadav_free(t);
}

void qadav_tensor_set_i64(qadav_tensor* t, int idx, long long val) {
    if (!t || idx < 0 || idx >= t->total_size) return;
    if (t->elem_kind == QADAV_ELEM_FLOAT64) {
        ((double*)t->data)[idx] = (double)val;
    } else {
        ((long long*)t->data)[idx] = val;
    }
}

long long qadav_tensor_get_i64(const qadav_tensor* t, int idx) {
    if (!t || idx < 0 || idx >= t->total_size) return 0;
    if (t->elem_kind == QADAV_ELEM_FLOAT64) {
        return (long long)(((double*)t->data)[idx]);
    }
    return ((long long*)t->data)[idx];
}

void qadav_tensor_set_f64(qadav_tensor* t, int idx, double val) {
    if (!t || idx < 0 || idx >= t->total_size) return;
    if (t->elem_kind == QADAV_ELEM_INT64) {
        ((long long*)t->data)[idx] = (long long)val;
    } else {
        ((double*)t->data)[idx] = val;
    }
}

double qadav_tensor_get_f64(const qadav_tensor* t, int idx) {
    if (!t || idx < 0 || idx >= t->total_size) return 0.0;
    if (t->elem_kind == QADAV_ELEM_INT64) {
        return (double)(((long long*)t->data)[idx]);
    }
    return ((double*)t->data)[idx];
}

qadav_tensor* qadav_tensor_add(const qadav_tensor* a, const qadav_tensor* b) {
    if (!a || !b) return (void*)0;
    int is_f64 = (a->elem_kind == QADAV_ELEM_FLOAT64 || b->elem_kind == QADAV_ELEM_FLOAT64);
    int elem_kind = is_f64 ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;
    qadav_tensor* res = qadav_tensor_create(elem_kind, a->rank, a->shape);
    if (!res) return (void*)0;

    for (int i = 0; i < a->total_size; i++) {
        if (is_f64) {
            double v = qadav_tensor_get_f64(a, i) + qadav_tensor_get_f64(b, i);
            qadav_tensor_set_f64(res, i, v);
        } else {
            long long v = qadav_tensor_get_i64(a, i) + qadav_tensor_get_i64(b, i);
            qadav_tensor_set_i64(res, i, v);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_sub(const qadav_tensor* a, const qadav_tensor* b) {
    if (!a || !b) return (void*)0;
    int is_f64 = (a->elem_kind == QADAV_ELEM_FLOAT64 || b->elem_kind == QADAV_ELEM_FLOAT64);
    int elem_kind = is_f64 ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;
    qadav_tensor* res = qadav_tensor_create(elem_kind, a->rank, a->shape);
    if (!res) return (void*)0;

    for (int i = 0; i < a->total_size; i++) {
        if (is_f64) {
            double v = qadav_tensor_get_f64(a, i) - qadav_tensor_get_f64(b, i);
            qadav_tensor_set_f64(res, i, v);
        } else {
            long long v = qadav_tensor_get_i64(a, i) - qadav_tensor_get_i64(b, i);
            qadav_tensor_set_i64(res, i, v);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_mul(const qadav_tensor* a, const qadav_tensor* b) {
    if (!a || !b) return (void*)0;
    int is_f64 = (a->elem_kind == QADAV_ELEM_FLOAT64 || b->elem_kind == QADAV_ELEM_FLOAT64);
    int elem_kind = is_f64 ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;
    qadav_tensor* res = qadav_tensor_create(elem_kind, a->rank, a->shape);
    if (!res) return (void*)0;

    for (int i = 0; i < a->total_size; i++) {
        if (is_f64) {
            double v = qadav_tensor_get_f64(a, i) * qadav_tensor_get_f64(b, i);
            qadav_tensor_set_f64(res, i, v);
        } else {
            long long v = qadav_tensor_get_i64(a, i) * qadav_tensor_get_i64(b, i);
            qadav_tensor_set_i64(res, i, v);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_div(const qadav_tensor* a, const qadav_tensor* b) {
    if (!a || !b) return (void*)0;
    int is_f64 = (a->elem_kind == QADAV_ELEM_FLOAT64 || b->elem_kind == QADAV_ELEM_FLOAT64);
    int elem_kind = is_f64 ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;
    qadav_tensor* res = qadav_tensor_create(elem_kind, a->rank, a->shape);
    if (!res) return (void*)0;

    for (int i = 0; i < a->total_size; i++) {
        if (is_f64) {
            double den = qadav_tensor_get_f64(b, i);
            double v = (den != 0.0) ? (qadav_tensor_get_f64(a, i) / den) : 0.0;
            qadav_tensor_set_f64(res, i, v);
        } else {
            long long den = qadav_tensor_get_i64(b, i);
            long long v = (den != 0) ? (qadav_tensor_get_i64(a, i) / den) : 0;
            qadav_tensor_set_i64(res, i, v);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_add_scalar_f64(const qadav_tensor* t, double s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(QADAV_ELEM_FLOAT64, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) + s);
    }
    return res;
}

qadav_tensor* qadav_tensor_mul_scalar_f64(const qadav_tensor* t, double s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(QADAV_ELEM_FLOAT64, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) * s);
    }
    return res;
}

qadav_tensor* qadav_tensor_sub_scalar_f64(const qadav_tensor* t, double s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(QADAV_ELEM_FLOAT64, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) - s);
    }
    return res;
}

qadav_tensor* qadav_tensor_div_scalar_f64(const qadav_tensor* t, double s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(QADAV_ELEM_FLOAT64, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        double v = (s != 0.0) ? (qadav_tensor_get_f64(t, i) / s) : 0.0;
        qadav_tensor_set_f64(res, i, v);
    }
    return res;
}

qadav_tensor* qadav_tensor_add_scalar_i64(const qadav_tensor* t, long long s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(t->elem_kind, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        if (t->elem_kind == QADAV_ELEM_FLOAT64) {
            qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) + (double)s);
        } else {
            qadav_tensor_set_i64(res, i, qadav_tensor_get_i64(t, i) + s);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_mul_scalar_i64(const qadav_tensor* t, long long s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(t->elem_kind, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        if (t->elem_kind == QADAV_ELEM_FLOAT64) {
            qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) * (double)s);
        } else {
            qadav_tensor_set_i64(res, i, qadav_tensor_get_i64(t, i) * s);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_sub_scalar_i64(const qadav_tensor* t, long long s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(t->elem_kind, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        if (t->elem_kind == QADAV_ELEM_FLOAT64) {
            qadav_tensor_set_f64(res, i, qadav_tensor_get_f64(t, i) - (double)s);
        } else {
            qadav_tensor_set_i64(res, i, qadav_tensor_get_i64(t, i) - s);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_div_scalar_i64(const qadav_tensor* t, long long s) {
    if (!t) return (void*)0;
    qadav_tensor* res = qadav_tensor_create(t->elem_kind, t->rank, t->shape);
    for (int i = 0; i < t->total_size; i++) {
        if (t->elem_kind == QADAV_ELEM_FLOAT64) {
            double v = (s != 0) ? (qadav_tensor_get_f64(t, i) / (double)s) : 0.0;
            qadav_tensor_set_f64(res, i, v);
        } else {
            long long v = (s != 0) ? (qadav_tensor_get_i64(t, i) / s) : 0;
            qadav_tensor_set_i64(res, i, v);
        }
    }
    return res;
}

qadav_tensor* qadav_tensor_matmul(const qadav_tensor* a, const qadav_tensor* b) {
    if (!a || !b) return (void*)0;
    int is_f64 = (a->elem_kind == QADAV_ELEM_FLOAT64 || b->elem_kind == QADAV_ELEM_FLOAT64);
    int elem_kind = is_f64 ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;

    if (a->rank == 2 && b->rank == 2) {
        int M = a->shape[0];
        int K = a->shape[1];
        int N = b->shape[1];
        int out_dims[2] = { M, N };
        qadav_tensor* res = qadav_tensor_create(elem_kind, 2, out_dims);
        if (!res) return (void*)0;

        for (int i = 0; i < M; i++) {
            for (int j = 0; j < N; j++) {
                if (is_f64) {
                    double sum = 0.0;
                    for (int k = 0; k < K; k++) {
                        sum += qadav_tensor_get_f64(a, i * K + k) * qadav_tensor_get_f64(b, k * N + j);
                    }
                    qadav_tensor_set_f64(res, i * N + j, sum);
                } else {
                    long long sum = 0;
                    for (int k = 0; k < K; k++) {
                        sum += qadav_tensor_get_i64(a, i * K + k) * qadav_tensor_get_i64(b, k * N + j);
                    }
                    qadav_tensor_set_i64(res, i * N + j, sum);
                }
            }
        }
        return res;
    }

    if (a->rank == 1 && b->rank == 2) {
        int K = a->shape[0];
        int N = b->shape[1];
        int out_dims[1] = { N };
        qadav_tensor* res = qadav_tensor_create(elem_kind, 1, out_dims);
        for (int j = 0; j < N; j++) {
            if (is_f64) {
                double sum = 0.0;
                for (int k = 0; k < K; k++) {
                    sum += qadav_tensor_get_f64(a, k) * qadav_tensor_get_f64(b, k * N + j);
                }
                qadav_tensor_set_f64(res, j, sum);
            } else {
                long long sum = 0;
                for (int k = 0; k < K; k++) {
                    sum += qadav_tensor_get_i64(a, k) * qadav_tensor_get_i64(b, k * N + j);
                }
                qadav_tensor_set_i64(res, j, sum);
            }
        }
        return res;
    }

    if (a->rank == 2 && b->rank == 1) {
        int M = a->shape[0];
        int K = a->shape[1];
        int out_dims[1] = { M };
        qadav_tensor* res = qadav_tensor_create(elem_kind, 1, out_dims);
        for (int i = 0; i < M; i++) {
            if (is_f64) {
                double sum = 0.0;
                for (int k = 0; k < K; k++) {
                    sum += qadav_tensor_get_f64(a, i * K + k) * qadav_tensor_get_f64(b, k);
                }
                qadav_tensor_set_f64(res, i, sum);
            } else {
                long long sum = 0;
                for (int k = 0; k < K; k++) {
                    sum += qadav_tensor_get_i64(a, i * K + k) * qadav_tensor_get_i64(b, k);
                }
                qadav_tensor_set_i64(res, i, sum);
            }
        }
        return res;
    }

    if (a->rank == 1 && b->rank == 1) {
        int K = a->shape[0];
        int out_dims[1] = { 1 };
        qadav_tensor* res = qadav_tensor_create(elem_kind, 1, out_dims);
        if (is_f64) {
            double sum = 0.0;
            for (int k = 0; k < K; k++) {
                sum += qadav_tensor_get_f64(a, k) * qadav_tensor_get_f64(b, k);
            }
            qadav_tensor_set_f64(res, 0, sum);
        } else {
            long long sum = 0;
            for (int k = 0; k < K; k++) {
                sum += qadav_tensor_get_i64(a, k) * qadav_tensor_get_i64(b, k);
            }
            qadav_tensor_set_i64(res, 0, sum);
        }
        return res;
    }

    return (void*)0;
}

long long qadav_tensor_reduce_sum_i64(const qadav_tensor* t) {
    if (!t) return 0;
    long long sum = 0;
    for (int i = 0; i < t->total_size; i++) {
        sum += qadav_tensor_get_i64(t, i);
    }
    return sum;
}

double qadav_tensor_reduce_sum_f64(const qadav_tensor* t) {
    if (!t) return 0.0;
    double sum = 0.0;
    for (int i = 0; i < t->total_size; i++) {
        sum += qadav_tensor_get_f64(t, i);
    }
    return sum;
}

long long qadav_tensor_reduce_prod_i64(const qadav_tensor* t) {
    if (!t || t->total_size == 0) return 0;
    long long prod = 1;
    for (int i = 0; i < t->total_size; i++) {
        prod *= qadav_tensor_get_i64(t, i);
    }
    return prod;
}

double qadav_tensor_reduce_prod_f64(const qadav_tensor* t) {
    if (!t || t->total_size == 0) return 0.0;
    double prod = 1.0;
    for (int i = 0; i < t->total_size; i++) {
        prod *= qadav_tensor_get_f64(t, i);
    }
    return prod;
}

qadav_tensor* qadav_tensor_transpose(const qadav_tensor* t) {
    if (!t || t->rank != 2) return (void*)0;
    int M = t->shape[0];
    int N = t->shape[1];
    int out_dims[2] = { N, M };
    qadav_tensor* res = qadav_tensor_create(t->elem_kind, 2, out_dims);
    if (!res) return (void*)0;

    for (int r = 0; r < M; r++) {
        for (int c = 0; c < N; c++) {
            if (t->elem_kind == QADAV_ELEM_FLOAT64) {
                double v = qadav_tensor_get_f64(t, r * N + c);
                qadav_tensor_set_f64(res, c * M + r, v);
            } else {
                long long v = qadav_tensor_get_i64(t, r * N + c);
                qadav_tensor_set_i64(res, c * M + r, v);
            }
        }
    }
    return res;
}

void qadav_tensor_print_i64(const qadav_tensor* t) {
    if (!t) return;
    print_str("tensor⟨int64⟩");
    if (t->rank > 0) {
        print_str("⟬");
        for (int i = 0; i < t->rank; i++) {
            if (i > 0) print_str("⸴ ");
            print_i64_raw(t->shape[i]);
        }
        print_str("⟭");
    }
    print_str(":\n");

    if (t->rank == 1) {
        print_str("⟬");
        for (int i = 0; i < t->total_size; i++) {
            if (i > 0) print_str("⸴ ");
            print_i64_raw(qadav_tensor_get_i64(t, i));
        }
        print_str("⟭\n");
    } else if (t->rank == 2) {
        int M = t->shape[0];
        int N = t->shape[1];
        print_str("⟬");
        for (int i = 0; i < M; i++) {
            if (i > 0) print_str("⸴ ");
            print_str("⟬");
            for (int j = 0; j < N; j++) {
                if (j > 0) print_str("⸴ ");
                print_i64_raw(qadav_tensor_get_i64(t, i * N + j));
            }
            print_str("⟭");
        }
        print_str("⟭\n");
    } else {
        print_i64_raw(qadav_tensor_get_i64(t, 0));
        print_str("\n");
    }
}

void qadav_tensor_print_f64(const qadav_tensor* t) {
    if (!t) return;
    print_str("tensor⟨float64⟩");
    if (t->rank > 0) {
        print_str("⟬");
        for (int i = 0; i < t->rank; i++) {
            if (i > 0) print_str("⸴ ");
            print_i64_raw(t->shape[i]);
        }
        print_str("⟭");
    }
    print_str(":\n");

    if (t->rank == 1) {
        print_str("⟬");
        for (int i = 0; i < t->total_size; i++) {
            if (i > 0) print_str("⸴ ");
            print_f64_raw(qadav_tensor_get_f64(t, i));
        }
        print_str("⟭\n");
    } else if (t->rank == 2) {
        int M = t->shape[0];
        int N = t->shape[1];
        print_str("⟬");
        for (int i = 0; i < M; i++) {
            if (i > 0) print_str("⸴ ");
            print_str("⟬");
            for (int j = 0; j < N; j++) {
                if (j > 0) print_str("⸴ ");
                print_f64_raw(qadav_tensor_get_f64(t, i * N + j));
            }
            print_str("⟭");
        }
        print_str("⟭\n");
    } else {
        print_f64_raw(qadav_tensor_get_f64(t, 0));
        print_str("\n");
    }
}

