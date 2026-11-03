#ifndef QADAV_VARIANT_H
#define QADAV_VARIANT_H

typedef struct {
    unsigned long long variant_type_hash;
    unsigned int       tag;
    unsigned int       payload_len;
    void*              payload;
} qadav_variant;

// Variant allocation and matching (zero libc)
qadav_variant*     qadav_variant_create(unsigned long long variant_type_hash, unsigned int tag, unsigned int payload_len, const void* payload_data);
qadav_variant*     qadav_variant_create_i64(unsigned long long variant_type_hash, unsigned int tag, unsigned long long val);
int                qadav_variant_get_tag(const qadav_variant* v);
void*              qadav_variant_get_payload(const qadav_variant* v);
unsigned long long qadav_variant_get_payload_i64(const qadav_variant* v);
int                qadav_variant_match(const qadav_variant* v, unsigned int expected_tag);
void               qadav_variant_free(qadav_variant* v);

#endif
