#ifndef QADAV_MERKLE_H
#define QADAV_MERKLE_H

#include "qadav_sha3.h"

#define QADAV_MERKLE_MAX_BLOCKS 4096

typedef struct qadav_merkle_tree {
    int            n_leaves;
    int            n_levels;
    unsigned char* hashes;        // flattened binary tree hashes
    long long*     block_offsets; // byte offset of each leaf block
    long long*     block_sizes;   // byte size of each leaf block
} qadav_merkle_tree;

// Build Merkle tree from code blocks
qadav_merkle_tree* qadav_merkle_build(const unsigned char* code_start,
                                      long long code_size,
                                      int block_size);

void qadav_merkle_free(qadav_merkle_tree* mt);

void qadav_merkle_root(qadav_merkle_tree* mt,
                       unsigned char out[QADAV_SHA3_256_SIZE]);

#endif

