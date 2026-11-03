#include "qadav_quasinary.h"

/* PRNG: xorshift64 seeded from CPU instruction cycle counter */
static unsigned long long quasi_prng_state = 0ULL;
static unsigned long long quasi_prng_next(void) {
    if (quasi_prng_state == 0ULL) {
        quasi_prng_state = __builtin_readcyclecounter();
        if (quasi_prng_state == 0ULL) quasi_prng_state = 0x517cc1b727220a95ULL;
    }
    unsigned long long x = quasi_prng_state;
    x ^= x << 13;
    x ^= x >> 7;
    x ^= x << 17;
    quasi_prng_state = x;
    return x;
}

static double quasi_uniform(void) {
    unsigned long long r = quasi_prng_next();
    return (double)(r & 0x1FFFFFFFFFFFFFULL) / (double)0x20000000000000ULL;
}

/* Zero-libc Taylor series exp(-x) for x >= 0 */
static double quasi_exp_neg(double x) {
    if (x < 0.0) return 1.0;
    if (x > 20.0) return 0.0;
    /* exp(-x) = 1 / exp(x) */
    double term = 1.0;
    double sum = 1.0;
    for (int i = 1; i < 25; i++) {
        term = term * x / (double)i;
        sum += term;
    }
    return 1.0 / sum;
}

int qadav_quasi_add(long long a, long long b) {
    long long r = a + b;
    if (r > 1) r = 1;
    if (r < -1) r = -1;
    return (int)r;
}

int qadav_quasi_mul(long long a, long long b) {
    return (int)(a * b);
}

int qadav_quasi_neg(long long a) {
    return (int)(-a);
}

qadav_tensor* qadav_quasi_grid_create(int rows, int cols) {
    int dims[2];
    dims[0] = rows;
    dims[1] = cols;
    return qadav_tensor_create(QADAV_ELEM_INT64, 2, dims);
}

long long qadav_quasi_grid_get(qadav_tensor* g, int r, int c) {
    if (!g || !g->data || r < 0 || r >= g->shape[0] || c < 0 || c >= g->shape[1]) return 0;
    long long* vals = (long long*)g->data;
    return vals[r * g->shape[1] + c];
}

void qadav_quasi_grid_set(qadav_tensor* g, int r, int c, long long v) {
    if (!g || !g->data || r < 0 || r >= g->shape[0] || c < 0 || c >= g->shape[1]) return;
    long long* vals = (long long*)g->data;
    if (v > 1) v = 1;
    if (v < -1) v = -1;
    vals[r * g->shape[1] + c] = v;
}

double qadav_quasi_energy_ising(qadav_tensor* grid) {
    if (!grid || !grid->data || grid->rank != 2) return 0.0;
    int R = grid->shape[0];
    int C = grid->shape[1];
    long long* s = (long long*)grid->data;
    double energy = 0.0;

    for (int r = 0; r < R; r++) {
        int r_down = (r + 1) % R;
        for (int c = 0; c < C; c++) {
            int c_right = (c + 1) % C;
            long long cur = s[r * C + c];
            long long right = s[r * C + c_right];
            long long down = s[r_down * C + c];
            /* E = -J * sum(s_i * s_j) with J = 1.0 */
            energy -= (double)(cur * right + cur * down);
        }
    }
    return energy;
}

double qadav_quasi_energy_majority(qadav_tensor* grid) {
    return qadav_quasi_energy_ising(grid);
}

void qadav_quasi_minimize(qadav_tensor* grid, int iterations) {
    if (!grid || !grid->data || grid->rank != 2) return;
    if (iterations <= 0) iterations = 1000;
    if (iterations > 100000) iterations = 100000; /* Security cap */

    int R = grid->shape[0];
    int C = grid->shape[1];
    long long* s = (long long*)grid->data;

    double T0 = 2.5;

    for (int iter = 0; iter < iterations; iter++) {
        double T = T0 * (1.0 - (double)iter / (double)iterations) + 0.001;

        int r = (int)(quasi_prng_next() % (unsigned long long)R);
        int c = (int)(quasi_prng_next() % (unsigned long long)C);

        int r_up = (r - 1 + R) % R;
        int r_down = (r + 1) % R;
        int c_left = (c - 1 + C) % C;
        int c_right = (c + 1) % C;

        long long cur = s[r * C + c];
        long long neighbor_sum = s[r_up * C + c] + s[r_down * C + c] +
                                 s[r * C + c_left] + s[r * C + c_right];

        /* Candidate flip: if +1 -> -1, if -1 -> +1, if 0 -> +1 or -1 */
        long long cand = -cur;
        if (cand == 0) {
            cand = (quasi_prng_next() & 1) ? 1 : -1;
        }

        /* delta_E = E_cand - E_cur = -J * (cand - cur) * neighbor_sum */
        double delta_E = -(double)(cand - cur) * (double)neighbor_sum;

        if (delta_E <= 0.0) {
            s[r * C + c] = cand; /* Always accept improvement */
        } else {
            double prob = quasi_exp_neg(delta_E / T);
            if (quasi_uniform() < prob) {
                s[r * C + c] = cand; /* Probabilistically accept thermal jump */
            }
        }
    }
}

void qadav_quasi_print(long long v) {
    if (v == -1) {
        qadav_write_raw("◄\n", 4);
    } else if (v == 1) {
        qadav_write_raw("►\n", 4);
    } else {
        qadav_write_raw("〇\n", 4);
    }
}

static void quasi_print_int(long long v) {
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

void qadav_quasi_grid_print(qadav_tensor* grid) {
    if (!grid || !grid->data || grid->rank != 2) {
        qadav_write_raw("grid: <nil>\n", 12);
        return;
    }
    int R = grid->shape[0];
    int C = grid->shape[1];
    long long* s = (long long*)grid->data;

    qadav_write_raw("tensor⟨quasi⟩⟬", 20);
    quasi_print_int(R);
    qadav_write_raw(" ⸴ ", 5);
    quasi_print_int(C);
    qadav_write_raw("⟭:\n⟬\n", 9);

    for (int r = 0; r < R; r++) {
        qadav_write_raw("  ⟬ ", 5);
        for (int c = 0; c < C; c++) {
            long long v = s[r * C + c];
            if (v == -1) {
                qadav_write_raw("◄", 3);
            } else if (v == 1) {
                qadav_write_raw("►", 3);
            } else {
                qadav_write_raw("〇", 3);
            }
            if (c < C - 1) {
                qadav_write_raw(" ⸴ ", 5);
            }
        }
        qadav_write_raw(" ⟭", 4);
        if (r < R - 1) {
            qadav_write_raw(" ⸴\n", 5);
        } else {
            qadav_write_raw("\n", 1);
        }
    }
    qadav_write_raw("⟭\n", 4);
}
