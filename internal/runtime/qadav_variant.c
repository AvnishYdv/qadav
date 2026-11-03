#include "qadav_variant.h"
#include "qadav_tensor.h"

static void var_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void var_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

qadav_variant* qadav_variant_create(unsigned long long variant_type_hash, unsigned int tag, unsigned int payload_len, const void* payload_data) {
    qadav_variant* v = (qadav_variant*)qadav_alloc(sizeof(qadav_variant));
    if (!v) return (void*)0;
    var_memset(v, 0, sizeof(qadav_variant));

    v->variant_type_hash = variant_type_hash;
    v->tag = tag;
    v->payload_len = payload_len;

    if (payload_len > 0 && payload_data != (void*)0) {
        v->payload = qadav_alloc(payload_len);
        if (v->payload) {
            var_memcpy(v->payload, payload_data, payload_len);
        }
    } else {
        v->payload = (void*)0;
    }

    return v;
}

qadav_variant* qadav_variant_create_i64(unsigned long long variant_type_hash, unsigned int tag, unsigned long long val) {
    return qadav_variant_create(variant_type_hash, tag, sizeof(unsigned long long), &val);
}

int qadav_variant_get_tag(const qadav_variant* v) {
    if (!v) return -1;
    return (int)v->tag;
}

void* qadav_variant_get_payload(const qadav_variant* v) {
    if (!v) return (void*)0;
    return v->payload;
}

unsigned long long qadav_variant_get_payload_i64(const qadav_variant* v) {
    if (!v || !v->payload) return 0ULL;
    return *(unsigned long long*)v->payload;
}

int qadav_variant_match(const qadav_variant* v, unsigned int expected_tag) {
    if (!v) return 0;
    return (v->tag == expected_tag) ? 1 : 0;
}

void qadav_variant_free(qadav_variant* v) {
    if (!v) return;
    if (v->payload) {
        qadav_free(v->payload);
        v->payload = (void*)0;
    }
    qadav_free(v);
}
