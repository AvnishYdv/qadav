#ifndef QADAV_SHA3_H
#define QADAV_SHA3_H

#define QADAV_SHA3_256_SIZE 32
#define QADAV_SHA3_256_RATE 136

// SHA3-256 (FIPS 202) — zero-libc, no external crypto
void qadav_sha3_256(const unsigned char* input, unsigned long long input_len,
                    unsigned char output[QADAV_SHA3_256_SIZE]);

// Constant-time comparison (for signature verification)
// Returns 0 if equal, 1 if different
int  qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len);

// Returns 1 if equal, 0 if different
int  qadav_memcmp_const(const void* a, const void* b, unsigned long long len);

#endif
