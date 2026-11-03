#include "qadav_dataset.h"
#include "qadav_io.h"

/* Forward external helpers */
extern void qadav_emit_int(long long v);
extern void qadav_emit_float(double v);
extern void qadav_emit_string(const char* s);
extern void qadav_tensor_print_i64(const qadav_tensor* t);
extern void qadav_tensor_print_f64(const qadav_tensor* t);

/* xorshift64 PRNG for Fisher-Yates dataset shuffle */
static unsigned long long ds_prng_state = 0ULL;
static unsigned long long ds_prng_next(void) {
    if (ds_prng_state == 0ULL) {
        ds_prng_state = __builtin_readcyclecounter();
        if (ds_prng_state == 0ULL) ds_prng_state = 0x9e3779b97f4a7c15ULL;
    }
    unsigned long long x = ds_prng_state;
    x ^= x << 13;
    x ^= x >> 7;
    x ^= x << 17;
    ds_prng_state = x;
    return x;
}

/* Zero-libc Newton-Raphson sqrt */
static double ds_sqrt(double x) {
    if (x <= 0.0) return 0.0;
    double g = x > 1.0 ? x * 0.5 : 1.0;
    for (int i = 0; i < 20; i++) {
        g = 0.5 * (g + x / g);
    }
    return g;
}

static int ds_strcmp(const char* a, const char* b) {
    if (!a || !b) return -1;
    while (*a && (*a == *b)) {
        a++;
        b++;
    }
    return *(const unsigned char*)a - *(const unsigned char*)b;
}

qadav_dataset* qadav_dataset_load(const char* path) {
    void* h = qadav_file_open_read(path);
    if (!h) {
        qadav_write_raw("error: cannot open dataset file\n", 32);
        return (qadav_dataset*)0;
    }

    /* 32-byte Header */
    char magic[4];
    unsigned int n_tensors = 0;
    unsigned long long total_bytes = 0ULL;
    char reserved[16];

    if (qadav_file_read(h, magic, 4) != 4 ||
        qadav_file_read(h, &n_tensors, 4) != 4 ||
        qadav_file_read(h, &total_bytes, 8) != 8 ||
        qadav_file_read(h, reserved, 16) != 16) {
        qadav_write_raw("error: corrupted QDS header\n", 28);
        qadav_file_close(h);
        return (qadav_dataset*)0;
    }

    if (magic[0] != 'Q' || magic[1] != 'D' || magic[2] != 'S' || magic[3] != '1') {
        qadav_write_raw("error: invalid QDS magic\n", 25);
        qadav_file_close(h);
        return (qadav_dataset*)0;
    }

    /* Dataset bomb protections */
    if (total_bytes > 1073741824ULL) { /* > 1GB */
        qadav_write_raw("error: dataset too large: exceeds 1GB\n", 38);
        qadav_file_close(h);
        return (qadav_dataset*)0;
    }

    if (n_tensors > 100) {
        qadav_write_raw("error: too many tensors: exceeds 100\n", 37);
        qadav_file_close(h);
        return (qadav_dataset*)0;
    }

    qadav_dataset* d = (qadav_dataset*)qadav_alloc(sizeof(qadav_dataset));
    d->n_tensors = (int)n_tensors;
    d->names = (char**)qadav_alloc(sizeof(char*) * n_tensors);
    d->tensors = (void**)qadav_alloc(sizeof(void*) * n_tensors);
    d->n_samples = 0;

    for (unsigned int i = 0; i < n_tensors; i++) {
        unsigned int name_len = 0;
        if (qadav_file_read(h, &name_len, 4) != 4 || name_len > 256) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        char* name = (char*)qadav_alloc(name_len + 1);
        if (qadav_file_read(h, name, name_len) != (int)name_len) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }
        name[name_len] = '\0';
        d->names[i] = name;

        unsigned int rank = 0;
        if (qadav_file_read(h, &rank, 4) != 4 || rank > 8) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        int dims[8];
        for (unsigned int r = 0; r < rank; r++) {
            unsigned int dim_val = 0;
            if (qadav_file_read(h, &dim_val, 4) != 4) {
                qadav_file_close(h);
                return (qadav_dataset*)0;
            }
            dims[r] = (int)dim_val;
        }

        unsigned int elem_kind = 0;
        unsigned long long data_len = 0ULL;
        if (qadav_file_read(h, &elem_kind, 4) != 4 ||
            qadav_file_read(h, &data_len, 8) != 8) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        if (data_len > 524288000ULL) { /* 500MB */
            qadav_write_raw("error: tensor data exceeds 500MB\n", 33);
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        int tensor_elem_kind = (elem_kind == 1) ? QADAV_ELEM_FLOAT64 : QADAV_ELEM_INT64;
        qadav_tensor* t = qadav_tensor_create(tensor_elem_kind, (int)rank, dims);
        if (!t) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        if (qadav_file_read(h, t->data, data_len) != (int)data_len) {
            qadav_file_close(h);
            return (qadav_dataset*)0;
        }

        d->tensors[i] = t;
        if (i == 0 && rank >= 1) {
            d->n_samples = dims[0];
        }
    }

    qadav_file_close(h);
    return d;
}

void qadav_dataset_free(qadav_dataset* d) {
    if (!d) return;
    for (int i = 0; i < d->n_tensors; i++) {
        if (d->names[i]) qadav_free(d->names[i]);
        if (d->tensors[i]) qadav_tensor_free((qadav_tensor*)d->tensors[i]);
    }
    if (d->names) qadav_free(d->names);
    if (d->tensors) qadav_free(d->tensors);
    qadav_free(d);
}

qadav_tensor* qadav_dataset_get(qadav_dataset* d, const char* name) {
    if (!d || d->n_tensors == 0) return (qadav_tensor*)0;
    if (!name || name[0] == '\0') return (qadav_tensor*)d->tensors[0];
    for (int i = 0; i < d->n_tensors; i++) {
        if (ds_strcmp(d->names[i], name) == 0) {
            return (qadav_tensor*)d->tensors[i];
        }
    }
    return (qadav_tensor*)d->tensors[0];
}

void qadav_dataset_normalize_minmax(qadav_dataset* d) {
    if (!d) return;
    for (int t_idx = 0; t_idx < d->n_tensors; t_idx++) {
        qadav_tensor* t = (qadav_tensor*)d->tensors[t_idx];
        if (!t || t->elem_kind != QADAV_ELEM_FLOAT64) continue;
        double* vals = (double*)t->data;
        if (t->total_size <= 0) continue;

        double min_val = vals[0];
        double max_val = vals[0];
        for (long long i = 1; i < t->total_size; i++) {
            if (vals[i] < min_val) min_val = vals[i];
            if (vals[i] > max_val) max_val = vals[i];
        }

        double diff = max_val - min_val;
        if (diff > 1e-12) {
            for (long long i = 0; i < t->total_size; i++) {
                vals[i] = (vals[i] - min_val) / diff;
            }
        }
    }
}

void qadav_dataset_normalize_zscore(qadav_dataset* d) {
    if (!d) return;
    for (int t_idx = 0; t_idx < d->n_tensors; t_idx++) {
        qadav_tensor* t = (qadav_tensor*)d->tensors[t_idx];
        if (!t || t->elem_kind != QADAV_ELEM_FLOAT64) continue;
        double* vals = (double*)t->data;
        if (t->total_size <= 0) continue;

        double sum = 0.0;
        for (long long i = 0; i < t->total_size; i++) sum += vals[i];
        double mean = sum / (double)t->total_size;

        double var_sum = 0.0;
        for (long long i = 0; i < t->total_size; i++) {
            double diff = vals[i] - mean;
            var_sum += diff * diff;
        }
        double std_dev = ds_sqrt(var_sum / (double)t->total_size);
        if (std_dev > 1e-12) {
            for (long long i = 0; i < t->total_size; i++) {
                vals[i] = (vals[i] - mean) / std_dev;
            }
        }
    }
}

void qadav_dataset_shuffle(qadav_dataset* d) {
    if (!d || d->n_samples <= 1) return;
    long long n = d->n_samples;

    /* Fisher-Yates shuffle on sample indices */
    for (long long i = n - 1; i > 0; i--) {
        long long j = (long long)(ds_prng_next() % (unsigned long long)(i + 1));
        if (i == j) continue;

        /* Swap sample row i and sample row j across all tensors */
        for (int t_idx = 0; t_idx < d->n_tensors; t_idx++) {
            qadav_tensor* t = (qadav_tensor*)d->tensors[t_idx];
            if (!t || t->rank < 1 || t->shape[0] <= 0) continue;
            long long sample_stride = t->total_size / t->shape[0];
            long long elem_size = t->elem_size;

            char* row_i = (char*)t->data + (i * sample_stride * elem_size);
            char* row_j = (char*)t->data + (j * sample_stride * elem_size);

            for (long long k = 0; k < sample_stride * elem_size; k++) {
                char temp = row_i[k];
                row_i[k] = row_j[k];
                row_j[k] = temp;
            }
        }
    }
}

static void dataset_print_int(long long v) {
    char buf[32];
    int n = 0;
    if (v < 0) {
        qadav_write_raw("-", 1);
        v = -v;
    }
    char tmp[32];
    int t = 0;
    if (v == 0) tmp[t++] = '0';
    while (v > 0) {
        tmp[t++] = (char)('0' + (v % 10));
        v /= 10;
    }
    while (t > 0) {
        buf[n++] = tmp[--t];
    }
    qadav_write_raw(buf, n);
}

void qadav_dataset_print(qadav_dataset* d) {
    if (!d) {
        qadav_write_raw("dataset: <nil>\n", 15);
        return;
    }
    qadav_write_raw("dataset⟨n_tensors=", 20);
    dataset_print_int(d->n_tensors);
    qadav_write_raw(", n_samples=", 12);
    dataset_print_int(d->n_samples);
    qadav_write_raw("⟩:\n", 5);
    for (int i = 0; i < d->n_tensors; i++) {
        qadav_write_raw("  [", 3);
        if (d->names && d->names[i]) {
            int len = 0;
            while (d->names[i][len]) len++;
            qadav_write_raw(d->names[i], len);
        }
        qadav_write_raw("] ", 2);
        qadav_tensor* pt = (qadav_tensor*)d->tensors[i];
        if (pt) {
            if (pt->elem_kind == QADAV_ELEM_INT64) {
                qadav_tensor_print_i64(pt);
            } else {
                qadav_tensor_print_f64(pt);
            }
        }
    }
}
