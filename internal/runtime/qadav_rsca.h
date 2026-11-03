#ifndef QADAV_RSCA_H
#define QADAV_RSCA_H

// Invention #23: Recursive Self-Compilation Attestation (RSCA)
// Provides 144-byte cryptographic attestation preimage per bootstrap stage,
// verifying stage0 (bootstrap), stage1 (self-host-1), stage2 (self-host-2)
// bit-for-bit equivalence and decoherence triggering on corruption.

#define QADAV_RSCA_MAX_STAGES 8
#define QADAV_RSCA_PREIMAGE_SIZE 144

typedef struct {
    unsigned long long stage;             // 8 bytes LE
    unsigned long long timestamp;         // 8 bytes LE (rdtsc)
    unsigned char      source_hash[32];   // SHA3-256 of source .ydv files
    unsigned char      compiler_hash[32]; // SHA3-256 of compiler binary used
    unsigned char      output_hash[32];   // SHA3-256 of generated output binary
    unsigned char      prev_stage_hash[32];// compound hash of preceding stage
    unsigned char      record_hash[32];   // SHA3-256(144-byte preimage)
} qadav_rsca_record;

typedef struct {
    qadav_rsca_record records[QADAV_RSCA_MAX_STAGES];
    int               n_records;
    int               is_verified;
} qadav_rsca_table;

// RSCA lifecycle functions
int  qadav_rsca_record_stage(unsigned long long stage,
                             const unsigned char source_hash[32],
                             const unsigned char compiler_hash[32],
                             const unsigned char output_hash[32]);
int  qadav_rsca_verify_record(int stage_idx);
int  qadav_rsca_compare_outputs(int stage_a, int stage_b);
int  qadav_rsca_attest_chain(void);
int  qadav_rsca_corrupt_record(int stage_idx, int field_type);
int  qadav_rsca_get_record_count(void);
void qadav_rsca_reset(void);

#endif
