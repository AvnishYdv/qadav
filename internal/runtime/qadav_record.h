#ifndef QADAV_RECORD_H
#define QADAV_RECORD_H

typedef struct {
    unsigned long long type_hash; // FNV-1a or SHA3-256 hash of record type name
    unsigned int       n_fields;  // Number of fields in record
    unsigned int       total_len; // Total record size in bytes
} qadav_record_header;

// Record allocation and lifecycle (zero libc)
void*              qadav_record_alloc(unsigned long long type_hash, unsigned int n_fields, const unsigned int* field_sizes);
int                qadav_record_set_field(void* rec_ptr, unsigned int field_idx, const void* data, unsigned int size);
void*              qadav_record_get_field(void* rec_ptr, unsigned int field_idx);
unsigned long long qadav_record_get_field_i64(void* rec_ptr, unsigned int field_idx);
void               qadav_record_set_field_i64(void* rec_ptr, unsigned int field_idx, unsigned long long val);
void*              qadav_record_get_field_ptr(void* rec_ptr, unsigned int field_idx);
void               qadav_record_set_field_ptr(void* rec_ptr, unsigned int field_idx, void* ptr);
unsigned int       qadav_record_get_field_count(const void* rec_ptr);
unsigned long long qadav_record_get_type_hash(const void* rec_ptr);
void               qadav_record_free(void* rec_ptr);

#endif
