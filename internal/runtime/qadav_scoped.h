#ifndef QADAV_SCOPED_H
#define QADAV_SCOPED_H

#include "qadav_sha3.h"
#include "qadav_qeeh.h"
#include "qadav_qemp.h"

#define QADAV_SCOPED_MAX_KEYS 16
#define QADAV_SCOPED_KEY_LEN  64

// CB-SRCPP Token Structure (Invention #14 - Patent Pending)
// Preserves privacy: Contains boundary hash commitment without revealing rules.
typedef struct qadav_scoped_token {
    char               name[64];
    unsigned char      token_id[32];          // SHA3(boundary_hash || keys_root || recipient_hash || issued_at || expiry)
    unsigned char      boundary_hash[32];     // Owner's F7 boundary hash
    unsigned char      keys_root[32];         // Merkle root over allowed keys
    unsigned char      recipient_hash[32];    // Recipient public key / identifier hash
    unsigned long long issued_at;
    unsigned long long expiry;
    int                revoked;               // 0 = active, 1 = revoked
    int                num_keys;
    char               allowed_keys[QADAV_SCOPED_MAX_KEYS][QADAV_SCOPED_KEY_LEN];
} qadav_scoped_token;

// Issue a scoped access token
qadav_scoped_token* qadav_scoped_issue(const char* name, void* owner_boundary,
                                       const unsigned char recipient_hash[32],
                                       const char* const* allowed_keys, int num_keys,
                                       unsigned long long expiry_seconds);

// Grant token to recipient
int qadav_scoped_grant(qadav_scoped_token* token, const unsigned char recipient_hash[32]);

// Revoke token
int qadav_scoped_revoke(qadav_scoped_token* token);

// Verify scoped token
int qadav_scoped_verify(const qadav_scoped_token* token,
                        const unsigned char recipient_hash[32],
                        const char* requested_key,
                        unsigned long long current_time);

// Perform cross-boundary recall using privacy-preserving token
int qadav_scoped_recall(qadav_memory_provenance* mp,
                        const qadav_scoped_token* token,
                        const unsigned char recipient_hash[32],
                        const char* requested_key,
                        unsigned long long current_time,
                        void** out_value, unsigned long long* out_len);

// Free scoped token
void qadav_scoped_free(qadav_scoped_token* token);

#endif

