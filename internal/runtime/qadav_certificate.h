#ifndef QADAV_CERTIFICATE_H
#define QADAV_CERTIFICATE_H

#include "qadav_sha3.h"
#include "qadav_qera.h"
#include "qadav_qeeh.h"

#define QADAV_CERT_WITNESS_MAX 512
#define QADAV_CERT_MAX_STEPS   1024

#define QADAV_CERT_DEDUCTIVE    1
#define QADAV_CERT_INDUCTIVE    2
#define QADAV_CERT_ABDUCTIVE    3
#define QADAV_CERT_VERIFICATION 4

typedef struct qadav_reason_certificate {
    unsigned char  kind;
    unsigned char  cert_id[32];
    unsigned char  premise_hash[32];
    unsigned char  conclusion_hash[32];
    unsigned char  witness[QADAV_CERT_WITNESS_MAX];
    unsigned int   witness_len;
    double         confidence;
    unsigned char  commitment[32];
} qadav_reason_certificate;

// Lifecycle counters
unsigned long long qadav_cert_get_create_count(void);
unsigned long long qadav_cert_get_free_count(void);
void qadav_certificate_free(qadav_reasoning_chain* chain);

// Attach certificate to a QERA reasoning step
int qadav_certificate_attach(
    qadav_reasoning_chain* chain,
    unsigned long long step_index,
    unsigned char kind,
    const unsigned char* witness, unsigned int witness_len,
    double confidence
);

// Structural verify certificate (kind/witness/commitment checks)
int qadav_certificate_verify(const qadav_reason_certificate* cert);

// Verify whole chain (all certs valid + commitment links match)
int qadav_certificate_verify_chain(qadav_reasoning_chain* chain);

// Simulate certificate tamper (--test only)
int qadav_certificate_corrupt(qadav_reasoning_chain* chain, int cert_idx, int at_idx);

#endif
