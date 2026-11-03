#ifndef QADAV_SELFCHECK_H
#define QADAV_SELFCHECK_H

#include "qadav_sha3.h"
#include "qadav_lamport.h"
#include "qadav_merkle.h"

// Embedded by linker or stored in .qsc sidecar file
typedef struct qadav_selfcheck_header {
    char          magic[8];        // "QSC1\0\0\0\0"
    unsigned char merkle_root[32]; // expected Merkle root
    unsigned char signature[QADAV_LAMPORT_SIG_SIZE]; // WOTS signature (1088 bytes)
    long long     sig_len;         // signature length (1088)
    long long     code_start;      // offset of code to check
    long long     code_size;       // size of code
    long long     block_size;      // block size used for Merkle
    int           n_leaves;        // number of leaves
    int           flags;           // reserved
} qadav_selfcheck_header;

// Runtime check: verify binary integrity against header & pubkey
// Returns 0 on success, 1 on tamper (triggers self-nullification)
int  qadav_selfcheck_verify(const qadav_selfcheck_header* hdr,
                            const unsigned char* code,
                            long long code_size,
                            const unsigned char* pubkey);

// Automatic startup check loading sidecar .qsc if present
// Returns 0 on success (or if no .qsc present). If tampered, calls nullify and exits.
int  qadav_selfcheck_run(void);

// Self-nullification: overwrite code pages with 0x00, then exit 0xDEADBEEF
__attribute__((noinline))
void qadav_selfcheck_nullify(void* code_start, long long code_size);

#endif
