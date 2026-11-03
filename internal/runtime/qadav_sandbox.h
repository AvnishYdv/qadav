#ifndef QADAV_SANDBOX_H
#define QADAV_SANDBOX_H

#include "qadav_qeeh.h"
#include "qadav_qera.h"
#include "qadav_sha3.h"

#define QADAV_SANDBOX_MAX 8
#define QADAV_SANDBOX_NAME_MAX 64

typedef enum {
    QADAV_SANDBOX_CLEAN = 0,
    QADAV_SANDBOX_VIOLATED = 1,
    QADAV_SANDBOX_QUARANTINED = 2
} qadav_sandbox_state;

typedef struct qadav_sandbox {
    char                       name[QADAV_SANDBOX_NAME_MAX];
    qadav_ethical_hamiltonian* boundary;
    qadav_reasoning_chain*     local_chain;
    unsigned char              boundary_hash[32];
    unsigned char              violation_attestation[32]; // SHA3 of (name||created_at||violations||state)
    qadav_sandbox_state        state;
    unsigned long long         violations;
    unsigned long long         created_at;
} qadav_sandbox;

// Global sandbox registry
qadav_sandbox* qadav_sandbox_create(const char* name, void* boundary, int capacity);
int qadav_sandbox_run(qadav_sandbox* sb, const char* tool_name, const char* arg);
int qadav_sandbox_verify(const qadav_sandbox* sb, unsigned char out_attestation[32]);
void qadav_sandbox_quarantine(qadav_sandbox* sb);
void qadav_sandbox_free(qadav_sandbox* sb);
qadav_sandbox* qadav_sandbox_find(const char* name);
void qadav_sandbox_reset(void);
void qadav_sandbox_set_host_chain(void* chain);

// Lifecycle counters (F24 pattern)
unsigned long long qadav_sandbox_get_create_count(void);
unsigned long long qadav_sandbox_get_free_count(void);

// Text conversion helper for YDV runtime emission
const char* qadav_sandbox_verify_text(const char* name);

#endif

