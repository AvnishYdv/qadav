#include "qadav_merkle.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

qadav_merkle_tree* qadav_merkle_build(const unsigned char* code_start,
                                      long long code_size,
                                      int block_size) {
    if (block_size <= 0) {
        block_size = 4096;
    }

    int n_leaves = 1;
    if (code_size > 0) {
        n_leaves = (int)((code_size + block_size - 1) / block_size);
    }
    if (n_leaves > QADAV_MERKLE_MAX_BLOCKS) {
        n_leaves = QADAV_MERKLE_MAX_BLOCKS;
    }

    qadav_merkle_tree* mt = (qadav_merkle_tree*)qadav_alloc(sizeof(qadav_merkle_tree));
    if (!mt) return (void*)0;

    mt->n_leaves = n_leaves;
    mt->block_offsets = (long long*)qadav_alloc(n_leaves * sizeof(long long));
    mt->block_sizes = (long long*)qadav_alloc(n_leaves * sizeof(long long));

    // Calculate maximum nodes in binary tree (<= 4 * n_leaves + 16)
    int max_nodes = n_leaves * 4 + 16;
    mt->hashes = (unsigned char*)qadav_alloc((unsigned long long)max_nodes * QADAV_SHA3_256_SIZE);

    if (!mt->block_offsets || !mt->block_sizes || !mt->hashes) {
        qadav_merkle_free(mt);
        return (void*)0;
    }

    // 1. Build leaf hashes
    if (code_size <= 0 || !code_start) {
        mt->block_offsets[0] = 0;
        mt->block_sizes[0] = 0;
        qadav_sha3_256((const unsigned char*)"", 0, mt->hashes);
    } else {
        for (int i = 0; i < n_leaves; i++) {
            long long offset = (long long)i * block_size;
            long long size = block_size;
            if (offset + size > code_size) {
                size = code_size - offset;
            }
            if (size < 0) size = 0;

            mt->block_offsets[i] = offset;
            mt->block_sizes[i] = size;

            qadav_sha3_256(code_start + offset, (unsigned long long)size,
                           mt->hashes + (i * QADAV_SHA3_256_SIZE));
        }
    }

    // 2. Build upper tree levels
    int current_level_start = 0;
    int current_level_count = n_leaves;
    int levels = 1;
    int total_nodes = n_leaves;

    while (current_level_count > 1) {
        int next_level_count = (current_level_count + 1) / 2;
        int next_level_start = total_nodes;

        for (int i = 0; i < next_level_count; i++) {
            int left_idx = current_level_start + i * 2;
            int right_idx = left_idx + 1;
            if (right_idx >= current_level_start + current_level_count) {
                right_idx = left_idx; // Duplicate last node if odd
            }

            unsigned char pair[QADAV_SHA3_256_SIZE * 2];
            for (int b = 0; b < QADAV_SHA3_256_SIZE; b++) {
                pair[b] = mt->hashes[left_idx * QADAV_SHA3_256_SIZE + b];
                pair[QADAV_SHA3_256_SIZE + b] = mt->hashes[right_idx * QADAV_SHA3_256_SIZE + b];
            }

            qadav_sha3_256(pair, QADAV_SHA3_256_SIZE * 2,
                           mt->hashes + ((next_level_start + i) * QADAV_SHA3_256_SIZE));
        }

        current_level_start = next_level_start;
        current_level_count = next_level_count;
        total_nodes += next_level_count;
        levels++;
    }

    mt->n_levels = levels;
    return mt;
}

void qadav_merkle_root(qadav_merkle_tree* mt, unsigned char out[QADAV_SHA3_256_SIZE]) {
    if (!mt || !mt->hashes || mt->n_leaves <= 0) {
        for (int i = 0; i < QADAV_SHA3_256_SIZE; i++) out[i] = 0;
        return;
    }

    // The root is the last 32 bytes computed in the tree
    // Total nodes = sum of all levels
    // Let's find root by stepping levels
    int curr_start = 0;
    int curr_count = mt->n_leaves;
    int total = mt->n_leaves;
    while (curr_count > 1) {
        int next_count = (curr_count + 1) / 2;
        curr_start = total;
        curr_count = next_count;
        total += next_count;
    }

    unsigned char* root_ptr = mt->hashes + (curr_start * QADAV_SHA3_256_SIZE);
    for (int i = 0; i < QADAV_SHA3_256_SIZE; i++) {
        out[i] = root_ptr[i];
    }
}

void qadav_merkle_free(qadav_merkle_tree* mt) {
    if (!mt) return;
    if (mt->hashes) qadav_free(mt->hashes);
    if (mt->block_offsets) qadav_free(mt->block_offsets);
    if (mt->block_sizes) qadav_free(mt->block_sizes);
    qadav_free(mt);
}
