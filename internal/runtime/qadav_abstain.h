#ifndef QADAV_ABSTAIN_H
#define QADAV_ABSTAIN_H

#include "qadav_qera.h"
#include "qadav_certificate.h"

#define QADAV_ABSTAIN_RETRY    1
#define QADAV_ABSTAIN_ESCALATE 2
#define QADAV_ABSTAIN_REFUSE   3

typedef struct qadav_abstain_policy {
    double  threshold;       // 0.0-1.0
    int     mode;            // RETRY/ESCALATE/REFUSE
    int     max_retries;     // only for RETRY mode
} qadav_abstain_policy;

// Lifecycle management
qadav_abstain_policy* qadav_abstain_policy_create(double threshold, int mode, int max_retries);
void qadav_abstain_policy_free(qadav_abstain_policy* p);
unsigned long long qadav_abstain_get_create_count(void);
unsigned long long qadav_abstain_get_free_count(void);

// Compute aggregated confidence (variance-penalized geometric composite)
double qadav_abstain_compute_confidence(qadav_reasoning_chain* chain);

// Check if chain confidence < threshold
// Returns 1 if should abstain, 0 if OK
int qadav_abstain_check(
    qadav_reasoning_chain* chain,
    qadav_abstain_policy* policy,
    int current_retry_count
);

// Execute abstention action
// Returns: mode (RETRY/ESCALATE/REFUSE) or 0 if no abstention
int qadav_abstain_execute(
    qadav_reasoning_chain* chain,
    qadav_abstain_policy* policy
);

#endif
