#ifndef QADAV_QECDECODER_H
#define QADAV_QECDECODER_H

#include "qadav_qec.h"
#include "qadav_sha3.h"

#define QADAV_QECD_MAX_QUBITS 32
#define QADAV_QECD_MAX_SYNDROMES 512
#define QADAV_QECD_MAX_CORRECTIONS 64

typedef enum {
    QECD_SURFACE_D3 = 3,
    QECD_SURFACE_D5 = 5,
    QECD_SURFACE_D7 = 7
} qecd_code_distance;

typedef struct qecd_stabilizer {
    unsigned int   x_pauli_mask;    // bitmask of X stabilizers
    unsigned int   z_pauli_mask;    // bitmask of Z stabilizers
    int            is_x_type;       // 1 = X stab, 0 = Z stab
} qecd_stabilizer;

typedef struct qecd_correction {
    unsigned int   qubit_index;
    unsigned char  pauli;           // 1=X, 2=Z, 3=Y
} qecd_correction;

typedef struct qecd_decode_result {
    int              code_distance;
    int              n_qubits;
    int              n_syndromes;
    unsigned char    raw_syndrome[QADAV_QECD_MAX_SYNDROMES];
    int              n_corrections;
    qecd_correction  corrections[QADAV_QECD_MAX_CORRECTIONS];
    unsigned char    decode_attestation[32];
} qecd_decode_result;

qecd_decode_result* qecd_decode(
    int distance,
    int n_qubits,
    const unsigned char* syndrome_bytes,
    int syndrome_len
);

int qecd_verify_decode_attestation(const qecd_decode_result* result);
int qecd_apply_corrections(qecd_decode_result* result, unsigned char* state_vector, int state_len);
void qecd_free(qecd_decode_result* result);

#endif
