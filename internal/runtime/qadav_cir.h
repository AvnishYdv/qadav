#ifndef QADAV_CIR_H
#define QADAV_CIR_H

#include "qadav_tensor.h"

typedef signed char int8_t;

typedef struct qadav_confidence {
    double  value;
    double  score;
    int8_t  epistemic;   /* 0=known, 1=unknown, 2=explore */
    char    provenance[64];
} qadav_confidence;

qadav_confidence* qadav_confidence_create(double value, double score, int8_t state);
void              qadav_confidence_free(qadav_confidence* c);
void              qadav_confidence_print(qadav_confidence* c);

/* Propagation */
qadav_confidence* qadav_confidence_add(qadav_confidence* a, qadav_confidence* b);
qadav_confidence* qadav_confidence_mul(qadav_confidence* a, qadav_confidence* b);
qadav_confidence* qadav_confidence_min(qadav_confidence* a, qadav_confidence* b);

/* Assertion */
int qadav_confidence_assert(qadav_confidence* c, double threshold);

#endif
