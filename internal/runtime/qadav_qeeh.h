#ifndef QADAV_QEEH_H
#define QADAV_QEEH_H

#include "qadav_tensor.h"
#include "qadav_quantum.h"
#include "qadav_sha3.h"

// Quantum-Entangled Ethical Hamiltonian (QEEH)
// Invention #6 — Enforces AGI ethical constraints through quantum decoherence.

typedef enum qadav_action_kind {
    QADAV_ACTION_PERMITTED    = 0,
    QADAV_ACTION_HARMFUL      = 1,
    QADAV_ACTION_UNAUTHORIZED = 2,
    QADAV_ACTION_LEAK_WEIGHTS = 3
} qadav_action_kind;

#define QADAV_QEEH_MAX_RULES 32
#define QADAV_QEEH_RULE_NAME_MAX 64

typedef struct qadav_ethical_hamiltonian {
    int     num_rules;
    double* penalties;      // rule penalties (0.0 to 1.0)
    double* conditions;     // current rule satisfaction (0.0 to 1.0)
    double  ground_energy;  // baseline ground state energy (0.0)
    char    rule_names[QADAV_QEEH_MAX_RULES][QADAV_QEEH_RULE_NAME_MAX];
} qadav_ethical_hamiltonian;

// Creates an ethical Hamiltonian boundary with given rule penalties
qadav_ethical_hamiltonian* qadav_qeeh_create(int num_rules, const double* penalties);

// Sets the name of a specific rule
void qadav_qeeh_set_rule_name(qadav_ethical_hamiltonian* h, int idx, const char* name);

// Computes a deterministic content-only SHA3-256 hash of the boundary (ASLR-free)
// Hashes ONLY: num_rules (4 bytes LE) || for each rule: rule_name (null-terminated) || penalty (8 bytes double LE)
int qadav_qeeh_boundary_hash(const void* boundary, unsigned char out[32]);

// Evaluates an action vector against the ethical Hamiltonian:
// Computes energy E = sum(penalty_i * (1.0 - condition_i)).
// Returns 0 if ethical (E <= ground_energy), or 1 if unethical (E > ground_energy).
int qadav_qeeh_evaluate(qadav_ethical_hamiltonian* h, const double* conditions);

// Evaluates a named action using cryptographic SHA3-256 message hashing and typed categorization:
// 0 = allowed, 1 = blocked (decoherence)
int qadav_qeeh_evaluate_action(qadav_ethical_hamiltonian* h, const char* action_str);

// Evaluates action by 32-byte cryptographic SHA3-256 message digest
int qadav_qeeh_evaluate_action_hash(qadav_ethical_hamiltonian* h, const unsigned char* action_hash);

// Evaluates action by typed category enum
int qadav_qeeh_evaluate_action_kind(qadav_ethical_hamiltonian* h, qadav_action_kind kind);

// Triggers quantum decoherence: physically zeroes out quantum state amplitudes
void qadav_qeeh_trigger_decoherence(void* qstate);

// Frees ethical Hamiltonian memory
void qadav_qeeh_free(qadav_ethical_hamiltonian* h);

#endif
