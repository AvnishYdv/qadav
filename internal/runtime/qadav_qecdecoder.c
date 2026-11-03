#include "qadav_qecdecoder.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);
int   qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len);

static void qecd_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void qecd_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void qecd_pack_u32(unsigned char* buf, unsigned int val) {
    buf[0] = (unsigned char)(val & 0xFF);
    buf[1] = (unsigned char)((val >> 8) & 0xFF);
    buf[2] = (unsigned char)((val >> 16) & 0xFF);
    buf[3] = (unsigned char)((val >> 24) & 0xFF);
}

static void qecd_compute_attestation(qecd_decode_result* res) {
    if (!res) return;

    // Preimage:
    // distance (4) || n_qubits (4) || n_syndromes (4) || raw_syndrome (n_syndromes) ||
    // n_corrections (4) || corrections (n_corrections * 5)
    unsigned char preimage[4 + 4 + 4 + QADAV_QECD_MAX_SYNDROMES + 4 + (QADAV_QECD_MAX_CORRECTIONS * 5)];
    unsigned long long offset = 0;

    qecd_pack_u32(preimage + offset, (unsigned int)res->code_distance);
    offset += 4;
    qecd_pack_u32(preimage + offset, (unsigned int)res->n_qubits);
    offset += 4;
    qecd_pack_u32(preimage + offset, (unsigned int)res->n_syndromes);
    offset += 4;

    if (res->n_syndromes > 0) {
        qecd_memcpy(preimage + offset, res->raw_syndrome, res->n_syndromes);
        offset += res->n_syndromes;
    }

    qecd_pack_u32(preimage + offset, (unsigned int)res->n_corrections);
    offset += 4;

    for (int i = 0; i < res->n_corrections; i++) {
        qecd_pack_u32(preimage + offset, res->corrections[i].qubit_index);
        offset += 4;
        preimage[offset] = res->corrections[i].pauli;
        offset += 1;
    }

    qadav_sha3_256(preimage, offset, res->decode_attestation);
}

qecd_decode_result* qecd_decode(
    int distance,
    int n_qubits,
    const unsigned char* syndrome_bytes,
    int syndrome_len
) {
    if (distance != QECD_SURFACE_D3 && distance != QECD_SURFACE_D5 && distance != QECD_SURFACE_D7) {
        return 0;
    }
    if (syndrome_len < 0 || syndrome_len > QADAV_QECD_MAX_SYNDROMES) {
        return 0;
    }

    if (n_qubits <= 0) {
        n_qubits = distance * distance;
    }
    if (n_qubits > QADAV_QECD_MAX_QUBITS) {
        n_qubits = QADAV_QECD_MAX_QUBITS;
    }

    qecd_decode_result* res = (qecd_decode_result*)qadav_alloc(sizeof(qecd_decode_result));
    if (!res) return 0;
    qecd_memset(res, 0, sizeof(qecd_decode_result));

    res->code_distance = distance;
    res->n_qubits = n_qubits;
    res->n_syndromes = syndrome_len;

    if (syndrome_bytes && syndrome_len > 0) {
        qecd_memcpy(res->raw_syndrome, syndrome_bytes, syndrome_len);

        // Greedy Min-Weight Decoding:
        // Parse bits from syndrome bytes.
        // Even bit positions -> X stabilizer violation -> Z error on nearest data qubit.
        // Odd bit positions  -> Z stabilizer violation -> X error on nearest data qubit.
        int total_bits = syndrome_len * 8;
        for (int b = 0; b < total_bits; b++) {
            int byte_idx = b / 8;
            int bit_idx = b % 8;
            if ((syndrome_bytes[byte_idx] >> bit_idx) & 1) {
                unsigned int target_qubit = (unsigned int)((b / 2) % n_qubits);
                unsigned char pauli_type = (b % 2 == 0) ? 1 : 2; // 1=X, 2=Z

                // Check if target_qubit already has a pending correction
                int found = -1;
                for (int c = 0; c < res->n_corrections; c++) {
                    if (res->corrections[c].qubit_index == target_qubit) {
                        found = c;
                        break;
                    }
                }

                if (found >= 0) {
                    unsigned char curr = res->corrections[found].pauli;
                    if (curr == pauli_type) {
                        // Idempotent cancellation: X+X = I, Z+Z = I
                        // Remove correction by shifting down
                        for (int k = found; k < res->n_corrections - 1; k++) {
                            res->corrections[k] = res->corrections[k + 1];
                        }
                        res->n_corrections--;
                    } else if ((curr == 1 && pauli_type == 2) || (curr == 2 && pauli_type == 1)) {
                        // X + Z = Y
                        res->corrections[found].pauli = 3;
                    } else if (curr == 3 && pauli_type == 1) {
                        // Y + X = Z
                        res->corrections[found].pauli = 2;
                    } else if (curr == 3 && pauli_type == 2) {
                        // Y + Z = X
                        res->corrections[found].pauli = 1;
                    }
                } else if (res->n_corrections < QADAV_QECD_MAX_CORRECTIONS) {
                    res->corrections[res->n_corrections].qubit_index = target_qubit;
                    res->corrections[res->n_corrections].pauli = pauli_type;
                    res->n_corrections++;
                }
            }
        }
    }

    qecd_compute_attestation(res);
    return res;
}

int qecd_verify_decode_attestation(const qecd_decode_result* result) {
    if (!result) return 0;

    qecd_decode_result copy = *result;
    unsigned char expected[32];
    qecd_compute_attestation(&copy);
    qecd_memcpy(expected, copy.decode_attestation, 32);

    return (qadav_sha3_ct_compare(result->decode_attestation, expected, 32) == 0);
}

int qecd_apply_corrections(qecd_decode_result* result, unsigned char* state_vector, int state_len) {
    if (!result || !state_vector || state_len <= 0) return -1;

    // Bounds check: require at least 2 bytes per qubit (amplitude + phase representation)
    if (state_len < result->n_qubits * 2) return -2;

    for (int i = 0; i < result->n_corrections; i++) {
        unsigned int q = result->corrections[i].qubit_index;
        if (q >= (unsigned int)result->n_qubits) return -2;
        if ((int)(q * 2 + 1) >= state_len) return -2;

        unsigned char p = result->corrections[i].pauli;
        if (p == 1) {
            // Pauli X: flip bit 0
            state_vector[q * 2] ^= 0x01;
        } else if (p == 2) {
            // Pauli Z: flip bit 1 (phase)
            state_vector[q * 2 + 1] ^= 0x01;
        } else if (p == 3) {
            // Pauli Y: flip both
            state_vector[q * 2] ^= 0x01;
            state_vector[q * 2 + 1] ^= 0x01;
        }
    }

    return 0;
}

void qecd_free(qecd_decode_result* result) {
    if (result) {
        qadav_free(result);
    }
}
