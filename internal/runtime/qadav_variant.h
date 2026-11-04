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
qadav_variant*     qadav_variant_create_ptr(unsigned long long variant_type_hash, unsigned int tag, void* ptr);
qadav_variant*     qadav_variant_create_2(unsigned long long variant_type_hash, unsigned int tag, unsigned long long f0, unsigned long long f1);
int                qadav_variant_set_payload_ptr(qadav_variant* v, void* ptr);
int                qadav_variant_has_payload(const qadav_variant* v);
int                qadav_variant_get_tag(const qadav_variant* v);
void*              qadav_variant_get_payload(const qadav_variant* v);
void*              qadav_variant_get_payload_ptr(const qadav_variant* v);
void*              qadav_variant_get_payload_offset(const qadav_variant* v, unsigned int offset);
unsigned long long qadav_variant_get_payload_i64(const qadav_variant* v);
unsigned long long qadav_variant_get_field_i64(const qadav_variant* v, unsigned int idx);
void*              qadav_variant_get_field_ptr(const qadav_variant* v, unsigned int idx);
int                qadav_variant_match(const qadav_variant* v, unsigned int expected_tag);
void               qadav_variant_free(qadav_variant* v);

#endif
