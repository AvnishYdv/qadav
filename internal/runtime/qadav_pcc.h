#ifndef QADAV_PCC_H
#define QADAV_PCC_H

// Invention #25: Proof-Carrying Codegen (PCC, Patent Pending)
// Cryptographic attestation certificate proving that emitted LLVM IR was produced
// directly from a canonical AST via a verified codegen execution path.
// Contains ast_hash[32], codegen_path_hash[32], and ir_hash[32] bound into
// a tamper-evident certificate with fail-stop decoherence (0xDEADBEEF).

typedef struct {
    unsigned char      ast_hash[32];          // SHA3-256 of canonical AST structure
    unsigned char      codegen_path_hash[32]; // SHA3-256 of executed codegen function chain
    unsigned char      ir_hash[32];           // SHA3-256 of emitted LLVM IR
    unsigned long long timestamp;             // Compilation timestamp / rdtsc
    unsigned char      cert_hash[32];         // SHA3-256 of 104-byte certificate preimage
    int                is_verified;           // 1 if active, 0 otherwise
} qadav_pcc_certificate;

// PCC runtime API
void qadav_pcc_init(void);
int  qadav_pcc_attest(const unsigned char ast_hash[32],
                      const unsigned char codegen_path_hash[32],
                      const unsigned char ir_hash[32]);
int  qadav_pcc_verify(void);
int  qadav_pcc_corrupt(int field_type);
int  qadav_pcc_get_status(void);
void qadav_pcc_reset(void);

#endif

