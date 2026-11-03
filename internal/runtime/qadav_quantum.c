/*
 * qadav_quantum.c — Zero-libc quantum state vector simulation
 *
 * NO <math.h>, <stdio.h>, <stdlib.h>, <string.h>, <complex.h>
 * Memory: qadav_alloc / qadav_free  (from qadav_tensor.h)
 * I/O:    qadav_emit_int / qadav_emit_float / qadav_emit_text
 *
 * PRNG: xorshift64 — suitable for quantum measurement simulation ONLY.
 *       NOT cryptographically secure.
 *
 * Supports up to QADAV_MAX_QUBITS = 20 qubits per process.
 */

#include "qadav_quantum.h"
#include "qadav_tensor.h"   /* qadav_alloc / qadav_free */

/* ── Forward declarations for emit functions ─────────────────────────── */
extern void qadav_emit_int(long long v);
extern void qadav_emit_float(double v);
extern void qadav_emit_text(const char *s);

/* ── Newton-Raphson sqrt (no <math.h>) ──────────────────────────────── */
static double qadav_q_sqrt(double x) {
    if (x <= 0.0) return 0.0;
    double g = x > 1.0 ? x * 0.5 : 1.0;
    int i;
    for (i = 0; i < 60; i++) {
        double ng = 0.5 * (g + x / g);
        double diff = ng - g;
        if (diff < 0.0) diff = -diff;
        if (diff < 1e-15) { g = ng; break; }
        g = ng;
    }
    return g;
}

/* ── xorshift64 PRNG — NOT cryptographic ─────────────────────────────── */
static unsigned long long prng_state = 0ULL;

static unsigned long long prng_next(void) {
    if (prng_state == 0ULL) {
        prng_state = __builtin_readcyclecounter();
        if (prng_state == 0ULL) prng_state = 0x123456789ABCDEFULL;
    }
    unsigned long long x = prng_state;
    x ^= x << 13;
    x ^= x >> 7;
    x ^= x << 17;
    prng_state = x;
    return x;
}

/* Returns a double in [0.0, 1.0) */
static double prng_double(void) {
    unsigned long long r = prng_next();
    /* Use top 53 bits for double mantissa */
    return (double)(r >> 11) * (1.0 / (double)(1ULL << 53));
}

/* ── Complex arithmetic ──────────────────────────────────────────────── */
static qadav_complex qc_mul(qadav_complex a, qadav_complex b) {
    qadav_complex c;
    c.real = a.real * b.real - a.imag * b.imag;
    c.imag = a.real * b.imag + a.imag * b.real;
    return c;
}

static qadav_complex qc_add(qadav_complex a, qadav_complex b) {
    qadav_complex c;
    c.real = a.real + b.real;
    c.imag = a.imag + b.imag;
    return c;
}

static qadav_complex qc_scale(qadav_complex a, double s) {
    qadav_complex c;
    c.real = a.real * s;
    c.imag = a.imag * s;
    return c;
}

static double qc_mag2(qadav_complex a) {
    return a.real * a.real + a.imag * a.imag;
}

/* 1/sqrt(2) */
static double SQRT2_INV = 0.7071067811865475;

/* ── Lifecycle ───────────────────────────────────────────────────────── */

qadav_qstate *qadav_qstate_create(int max_qubits) {
    if (max_qubits <= 0) max_qubits = 1;
    if (max_qubits > QADAV_MAX_QUBITS) max_qubits = QADAV_MAX_QUBITS;

    qadav_qstate *qs = (qadav_qstate *)qadav_alloc(sizeof(qadav_qstate));
    if (!qs) return (void*)0;

    long long n_states = 1LL << max_qubits;
    qs->n_qubits   = 0;
    qs->n_states   = n_states;
    qs->capacity   = max_qubits;
    qs->amplitudes = (qadav_complex *)qadav_alloc((unsigned long long)n_states * sizeof(qadav_complex));
    if (!qs->amplitudes) {
        qadav_free(qs);
        return (void*)0;
    }

    /* Initialise to |0...0⟩ */
    long long i;
    for (i = 0; i < n_states; i++) {
        qs->amplitudes[i].real = 0.0;
        qs->amplitudes[i].imag = 0.0;
    }
    qs->amplitudes[0].real = 1.0; /* |0⟩ */
    return qs;
}

static qadav_qstate *global_qs = (void*)0;

qadav_qstate *qadav_get_qstate(void) {
    if (!global_qs) {
        global_qs = qadav_qstate_create(QADAV_MAX_QUBITS);
    }
    return global_qs;
}

void qadav_qstate_free(qadav_qstate *qs) {
    if (!qs) return;
    if (qs->amplitudes) qadav_free(qs->amplitudes);
    qadav_free(qs);
}

int qadav_qubit_alloc(qadav_qstate *qs) {
    if (!qs) return -1;
    if (qs->n_qubits >= qs->capacity) return -1;
    int qid = qs->n_qubits;
    qs->n_qubits++;
    return qid;
}

/* ── Gate helpers ────────────────────────────────────────────────────── */

/*
 * Apply a 2x2 unitary to qubit qid across all basis states.
 * u00 u01
 * u10 u11
 */
static void apply_1q_gate(qadav_qstate *qs, int qid,
                           qadav_complex u00, qadav_complex u01,
                           qadav_complex u10, qadav_complex u11) {
    long long bit = 1LL << qid;
    long long n   = qs->n_states;
    long long i;
    for (i = 0; i < n; i++) {
        if (i & bit) continue; /* process pairs: i (|0⟩), i|bit (|1⟩) */
        long long j = i | bit;
        qadav_complex a0 = qs->amplitudes[i];
        qadav_complex a1 = qs->amplitudes[j];
        qs->amplitudes[i] = qc_add(qc_mul(u00, a0), qc_mul(u01, a1));
        qs->amplitudes[j] = qc_add(qc_mul(u10, a0), qc_mul(u11, a1));
    }
}

/* ── Single-qubit gates ─────────────────────────────────────────────── */

void qadav_gate_h(qadav_qstate *qs, int qid) {
    qadav_complex h = { SQRT2_INV, 0.0 };
    qadav_complex hm = { -SQRT2_INV, 0.0 };
    apply_1q_gate(qs, qid, h, h, h, hm);
}

void qadav_gate_x(qadav_qstate *qs, int qid) {
    qadav_complex zero = {0.0, 0.0};
    qadav_complex one  = {1.0, 0.0};
    apply_1q_gate(qs, qid, zero, one, one, zero);
}

void qadav_gate_y(qadav_qstate *qs, int qid) {
    qadav_complex zero = {0.0, 0.0};
    qadav_complex pi   = {0.0, 1.0};   /* +i */
    qadav_complex mi   = {0.0, -1.0};  /* -i */
    apply_1q_gate(qs, qid, zero, mi, pi, zero);
}

void qadav_gate_z(qadav_qstate *qs, int qid) {
    qadav_complex one = {1.0, 0.0};
    qadav_complex mone = {-1.0, 0.0};
    qadav_complex zero = {0.0, 0.0};
    apply_1q_gate(qs, qid, one, zero, zero, mone);
}

void qadav_gate_s(qadav_qstate *qs, int qid) {
    /* S = diag(1, i) */
    qadav_complex one  = {1.0, 0.0};
    qadav_complex zero = {0.0, 0.0};
    qadav_complex pi   = {0.0, 1.0};
    apply_1q_gate(qs, qid, one, zero, zero, pi);
}

void qadav_gate_t(qadav_qstate *qs, int qid) {
    /* T = diag(1, e^{ipi/4}) = diag(1, (1+i)/sqrt(2)) */
    qadav_complex one  = {1.0, 0.0};
    qadav_complex zero = {0.0, 0.0};
    qadav_complex t11  = { SQRT2_INV, SQRT2_INV }; /* (1+i)/sqrt(2) */
    apply_1q_gate(qs, qid, one, zero, zero, t11);
}

/* ── Two-qubit CNOT gate ────────────────────────────────────────────── */

void qadav_gate_cnot(qadav_qstate *qs, int control, int target) {
    long long ctrl_bit   = 1LL << control;
    long long target_bit = 1LL << target;
    long long n = qs->n_states;
    long long i;
    for (i = 0; i < n; i++) {
        /* Flip target only when control is |1⟩ and target is |0⟩ */
        if ((i & ctrl_bit) && !(i & target_bit)) {
            long long j = i | target_bit;
            qadav_complex tmp = qs->amplitudes[i];
            qs->amplitudes[i] = qs->amplitudes[j];
            qs->amplitudes[j] = tmp;
        }
    }
}

/* ── Measurement ────────────────────────────────────────────────────── */

double qadav_probability(qadav_qstate *qs, int qid) {
    if (!qs) return 0.0;
    long long bit = 1LL << qid;
    long long n   = qs->n_states;
    double prob1 = 0.0;
    long long i;
    for (i = 0; i < n; i++) {
        if (i & bit) {
            prob1 += qc_mag2(qs->amplitudes[i]);
        }
    }
    return prob1;
}

int qadav_measure(qadav_qstate *qs, int qid) {
    if (!qs) return 0;

    double prob1 = qadav_probability(qs, qid);
    double r     = prng_double();
    int    result = (r < prob1) ? 1 : 0;

    /* Collapse: zero out states inconsistent with measurement */
    long long bit = 1LL << qid;
    long long n   = qs->n_states;
    long long i;
    for (i = 0; i < n; i++) {
        int has_bit = (i & bit) ? 1 : 0;
        if (has_bit != result) {
            qs->amplitudes[i].real = 0.0;
            qs->amplitudes[i].imag = 0.0;
        }
    }

    /* Renormalise */
    double norm2 = 0.0;
    for (i = 0; i < n; i++) {
        norm2 += qc_mag2(qs->amplitudes[i]);
    }
    if (norm2 > 1e-15) {
        double inv_norm = 1.0 / qadav_q_sqrt(norm2);
        for (i = 0; i < n; i++) {
            qs->amplitudes[i] = qc_scale(qs->amplitudes[i], inv_norm);
        }
    }

    return result;
}

/* ── Debug print ────────────────────────────────────────────────────── */

void qadav_qstate_print(qadav_qstate *qs) {
    if (!qs) return;
    long long n = qs->n_states;
    long long i;
    for (i = 0; i < n; i++) {
        double mag2 = qc_mag2(qs->amplitudes[i]);
        if (mag2 > 1e-10) {
            qadav_emit_int(i);
            qadav_emit_float(qs->amplitudes[i].real);
            qadav_emit_float(qs->amplitudes[i].imag);
        }
    }
}

