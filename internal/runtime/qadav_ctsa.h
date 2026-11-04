#ifndef QADAV_CTSA_H
#define QADAV_CTSA_H

// Invention #24: Compile-Time Type Safety Attestation (CTSA)
// Cryptographic attestation certificate proving type safety preserved across compilation,
// containing AST hash, type environment hash, rule chain rolling hash, function signature hashes,
// checked record fields, and tamper-evident runtime decoherence.

#define QADAV_CTSA_MAX_SIGS 32
#define QADAV_CTSA_MAX_FIELDS 32

typedef struct {
    unsigned char  ast_hash[32];               // SHA3-256 of AST structure
    unsigned char  type_env_hash[32];          // SHA3-256 of type environment
    unsigned char  rule_chain_hash[32];        // SHA3-256 rolling hash of applied type rules
    unsigned long long num_signatures;         // Count of checked function signatures
    unsigned char  sig_hashes[QADAV_CTSA_MAX_SIGS][32]; // Hashes of checked function signatures
    unsigned long long num_fields;             // Count of checked record fields
    unsigned char  field_hashes[QADAV_CTSA_MAX_FIELDS][32]; // Hashes of checked record fields
    unsigned long long timestamp;              // Compile timestamp / rdtsc
    unsigned char  cert_hash[32];              // SHA3-256 of complete certificate
    int            is_verified;                // 1 if verified, 0 if unverified
} qadav_ctsa_certificate;

// CTSA runtime API
void qadav_ctsa_init(void);
int  qadav_ctsa_attest(const unsigned char ast_hash[32],
                       const unsigned char type_env_hash[32],
                       const unsigned char rule_chain_hash[32],
                       unsigned long long num_sigs,
                       unsigned long long num_fields);
int  qadav_ctsa_verify(void);
int  qadav_ctsa_corrupt(int field_type);
int  qadav_ctsa_get_status(void);
void qadav_ctsa_reset(void);

#endif

