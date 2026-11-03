#include "qadav_record.h"
#include "qadav_tensor.h"

// Memory layout of allocated record:
// [qadav_record_header] (16 bytes)
// [unsigned int offsets[n_fields]] (4 * n_fields bytes)
// [unsigned int sizes[n_fields]]   (4 * n_fields bytes)
// [payload data...]

static void rec_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static void rec_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

void* qadav_record_alloc(unsigned long long type_hash, unsigned int n_fields, const unsigned int* field_sizes) {
    unsigned int header_size = (unsigned int)sizeof(qadav_record_header);
    unsigned int meta_size = n_fields * (unsigned int)sizeof(unsigned int) * 2;
    unsigned int offset = header_size + meta_size;

    // Calculate total size
    for (unsigned int i = 0; i < n_fields; i++) {
        unsigned int sz = (field_sizes != (void*)0) ? field_sizes[i] : 8;
        if (sz == 0) sz = 8;
        // Align to 8 bytes
        if (sz % 8 != 0) {
            sz = ((sz + 7) / 8) * 8;
        }
        offset += sz;
    }

    void* mem = qadav_alloc(offset);
    if (!mem) return (void*)0;
    rec_memset(mem, 0, offset);

    qadav_record_header* hdr = (qadav_record_header*)mem;
    hdr->type_hash = type_hash;
    hdr->n_fields = n_fields;
    hdr->total_len = offset;

    unsigned int* offsets = (unsigned int*)((char*)mem + header_size);
    unsigned int* sizes = offsets + n_fields;

    unsigned int current_off = header_size + meta_size;
    for (unsigned int i = 0; i < n_fields; i++) {
        unsigned int sz = (field_sizes != (void*)0) ? field_sizes[i] : 8;
        if (sz == 0) sz = 8;
        if (sz % 8 != 0) {
            sz = ((sz + 7) / 8) * 8;
        }
        offsets[i] = current_off;
        sizes[i] = sz;
        current_off += sz;
    }

    return mem;
}

int qadav_record_set_field(void* rec_ptr, unsigned int field_idx, const void* data, unsigned int size) {
    if (!rec_ptr || !data) return -1;
    qadav_record_header* hdr = (qadav_record_header*)rec_ptr;
    if (field_idx >= hdr->n_fields) return -2;

    unsigned int* offsets = (unsigned int*)((char*)rec_ptr + sizeof(qadav_record_header));
    unsigned int* sizes = offsets + hdr->n_fields;

    unsigned int max_sz = sizes[field_idx];
    unsigned int copy_sz = size < max_sz ? size : max_sz;
    void* dst = (char*)rec_ptr + offsets[field_idx];
    rec_memcpy(dst, data, copy_sz);
    return 0;
}

void* qadav_record_get_field(void* rec_ptr, unsigned int field_idx) {
    if (!rec_ptr) return (void*)0;
    qadav_record_header* hdr = (qadav_record_header*)rec_ptr;
    if (field_idx >= hdr->n_fields) return (void*)0;

    unsigned int* offsets = (unsigned int*)((char*)rec_ptr + sizeof(qadav_record_header));
    return (void*)((char*)rec_ptr + offsets[field_idx]);
}

unsigned long long qadav_record_get_field_i64(void* rec_ptr, unsigned int field_idx) {
    void* f = qadav_record_get_field(rec_ptr, field_idx);
    if (!f) return 0ULL;
    return *(unsigned long long*)f;
}

void qadav_record_set_field_i64(void* rec_ptr, unsigned int field_idx, unsigned long long val) {
    qadav_record_set_field(rec_ptr, field_idx, &val, (unsigned int)sizeof(unsigned long long));
}

void* qadav_record_get_field_ptr(void* rec_ptr, unsigned int field_idx) {
    void* f = qadav_record_get_field(rec_ptr, field_idx);
    if (!f) return (void*)0;
    return *(void**)f;
}

void qadav_record_set_field_ptr(void* rec_ptr, unsigned int field_idx, void* ptr) {
    void* f = qadav_record_get_field(rec_ptr, field_idx);
    if (f) {
        *(void**)f = ptr;
    }
}

unsigned int qadav_record_get_field_count(const void* rec_ptr) {
    if (!rec_ptr) return 0;
    const qadav_record_header* hdr = (const qadav_record_header*)rec_ptr;
    return hdr->n_fields;
}

unsigned long long qadav_record_get_type_hash(const void* rec_ptr) {
    if (!rec_ptr) return 0ULL;
    const qadav_record_header* hdr = (const qadav_record_header*)rec_ptr;
    return hdr->type_hash;
}

void qadav_record_free(void* rec_ptr) {
    if (rec_ptr) {
        qadav_free(rec_ptr);
    }
}
