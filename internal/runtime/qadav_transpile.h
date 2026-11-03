#ifndef QADAV_TRANSPILE_H
#define QADAV_TRANSPILE_H

#include "qadav_qhal.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    TRANSPILE_TARGET_QISKIT = 0,
    TRANSPILE_TARGET_CIRQ   = 1,
    TRANSPILE_TARGET_IONQ   = 2,
    TRANSPILE_TARGET_QASM3  = 3
} transpile_target;

typedef struct transpile_opts {
    int     optimize_adjacent;      // 1 = cancel adjacent CNOT·CNOT, X·X, etc.
    int     decompose_u3;           // 1 = decompose to native gates
    int     reduce_depth;           // 1 = reorder to reduce critical path
    int     max_qubits;             // hardware limit
} transpile_opts;

int transpile_circuit(
    const qhal_circuit* in,
    transpile_target target,
    const transpile_opts* opts,
    char* out_buf,
    unsigned long long buf_len,
    unsigned long long* out_written,
    int* out_op_count
);

int transpile_supported_targets(char* out_buf, unsigned long long buf_len);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_TRANSPILE_H */
