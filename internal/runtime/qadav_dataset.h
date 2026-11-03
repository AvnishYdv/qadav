#ifndef QADAV_DATASET_H
#define QADAV_DATASET_H

#include "qadav_tensor.h"

typedef struct qadav_dataset {
    int          n_tensors;
    char**       names;
    void**       tensors;   /* qadav_tensor* array */
    long long    n_samples;
} qadav_dataset;

qadav_dataset* qadav_dataset_load(const char* path);
void           qadav_dataset_free(qadav_dataset* d);
qadav_tensor*  qadav_dataset_get(qadav_dataset* d, const char* name);
void           qadav_dataset_normalize_minmax(qadav_dataset* d);
void           qadav_dataset_normalize_zscore(qadav_dataset* d);
void           qadav_dataset_shuffle(qadav_dataset* d);   /* Fisher-Yates + xorshift64 */
void           qadav_dataset_print(qadav_dataset* d);

#endif

