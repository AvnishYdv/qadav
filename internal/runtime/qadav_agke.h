#ifndef QADAV_AGKE_H
#define QADAV_AGKE_H

#include "qadav_sha3.h"

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_AGKE_MAX_RECORDS 256

typedef struct qadav_agke_record {
    unsigned long long dispatch_id;
    unsigned long long timestamp;
    char               kernel_name[64];
    unsigned char      kernel_hash[32];
    unsigned char      input_digest[32];
    unsigned char      output_digest[32];
    unsigned char      device_uuid[32];
    unsigned char      record_hash[32];
} qadav_agke_record;

typedef struct qadav_agke_table {
    int               n_records;
    qadav_agke_record records[QADAV_AGKE_MAX_RECORDS];
    int               is_attested;
    int               is_verified;
} qadav_agke_table;

int  qadav_agke_record_dispatch(const char* kernel_name,
                                const unsigned char* kernel_bytecode, unsigned long long bytecode_len,
                                const void* input_data, unsigned long long input_len,
                                const void* output_data, unsigned long long output_len);

int  qadav_agke_attest(void* reasoning_chain, void* auth_boundary);
int  qadav_agke_verify(void* reasoning_chain, void* auth_boundary);
int  qadav_agke_corrupt_record(int record_index, int field_type);
int  qadav_agke_get_count(void);
const qadav_agke_record* qadav_agke_get_record(int index);
void qadav_agke_reset(void);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_AGKE_H */

