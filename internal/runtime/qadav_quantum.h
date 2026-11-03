#ifndef QADAV_QUANTUM_H
#define QADAV_QUANTUM_H

/*
 * qadav_quantum.h — Zero-libc quantum state vector simulation runtime
 * NO <math.h>, <stdio.h>, <stdlib.h>, <string.h>, <complex.h>
 * All memory via qadav_alloc/qadav_free. I/O via qadav_emit_*.
 *
 * SECURITY NOTE: xorshift64 PRNG is NOT cryptographically secure.
 * Use only for quantum measurement simulation. Phase 3 will add ChaCha20.
 *
 * Maximum supported qubits: 20 (2^20 amplitudes = ~32MB at 16 bytes each)
 */

#define QADAV_MAX_QUBITS 20

/* Complex number — no <complex.h> dependency */
typedef struct {
    double real;
    double imag;
} qadav_complex;

/* Quantum state vector */
typedef struct {
    int           n_qubits;   /* current number of allocated qubits */
    long long     n_states;   /* 2^n_qubits amplitude slots */
    qadav_complex *amplitudes; /* heap via qadav_alloc */
    int           capacity;   /* max qubits reserved at creation */
} qadav_qstate;

/* ── Lifecycle ──────────────────────────────────────────────────────── */

/* Create a state for up to max_qubits qubits, initialized to |0...0⟩. */
qadav_qstate *qadav_qstate_create(int max_qubits);

/* Get or initialize process-global quantum state. */
qadav_qstate *qadav_get_qstate(void);

/* Release state memory. */
void qadav_qstate_free(qadav_qstate *qs);

/* Allocate next qubit ID in the state (initialised to |0⟩). Returns qubit id. */
int qadav_qubit_alloc(qadav_qstate *qs);

/* ── Single-qubit gates ─────────────────────────────────────────────── */

void qadav_gate_h(qadav_qstate *qs, int qid);   /* Hadamard    */
void qadav_gate_x(qadav_qstate *qs, int qid);   /* Pauli-X     */
void qadav_gate_y(qadav_qstate *qs, int qid);   /* Pauli-Y     */
void qadav_gate_z(qadav_qstate *qs, int qid);   /* Pauli-Z     */
void qadav_gate_s(qadav_qstate *qs, int qid);   /* S phase     */
void qadav_gate_t(qadav_qstate *qs, int qid);   /* T phase     */

/* ── Two-qubit gates ────────────────────────────────────────────────── */

void qadav_gate_cnot(qadav_qstate *qs, int control, int target);

/* ── Measurement ────────────────────────────────────────────────────── */

/*
 * Measure qubit qid: returns 0 or 1.
 * Collapses the state and renormalises.
 * PRNG: xorshift64 — NOT cryptographic.
 */
int qadav_measure(qadav_qstate *qs, int qid);

/* Compute P(|1⟩) for qubit qid without collapsing state. */
double qadav_probability(qadav_qstate *qs, int qid);

/* Print state vector amplitudes (debug). */
void qadav_qstate_print(qadav_qstate *qs);

#endif /* QADAV_QUANTUM_H */

