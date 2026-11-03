#include "qadav_sha3.h"

#define ROL64(a, n) (((a) << (n)) | ((a) >> (64 - (n))))

static const unsigned long long keccak_rndc[24] = {
    0x0000000000000001ULL, 0x0000000000008082ULL, 0x800000000000808aULL, 0x8000000080008000ULL,
    0x000000000000808bULL, 0x0000000080000001ULL, 0x8000000080008081ULL, 0x8000000000008009ULL,
    0x000000000000008aULL, 0x0000000000000088ULL, 0x0000000080008009ULL, 0x000000008000000aULL,
    0x000000008000808bULL, 0x800000000000008bULL, 0x8000000000008089ULL, 0x8000000000008003ULL,
    0x8000000000008002ULL, 0x8000000000000080ULL, 0x000000000000800aULL, 0x800000008000000aULL,
    0x8000000080008081ULL, 0x8000000000008080ULL, 0x0000000080000001ULL, 0x8000000080008008ULL
};

static void keccak_f1600(unsigned long long s[25]) {
    for (int r = 0; r < 24; r++) {
        // 1. Theta step (fully unrolled, 0 modulo operations)
        unsigned long long C0 = s[0] ^ s[5] ^ s[10] ^ s[15] ^ s[20];
        unsigned long long C1 = s[1] ^ s[6] ^ s[11] ^ s[16] ^ s[21];
        unsigned long long C2 = s[2] ^ s[7] ^ s[12] ^ s[17] ^ s[22];
        unsigned long long C3 = s[3] ^ s[8] ^ s[13] ^ s[18] ^ s[23];
        unsigned long long C4 = s[4] ^ s[9] ^ s[14] ^ s[19] ^ s[24];

        unsigned long long D0 = C4 ^ ROL64(C1, 1);
        unsigned long long D1 = C0 ^ ROL64(C2, 1);
        unsigned long long D2 = C1 ^ ROL64(C3, 1);
        unsigned long long D3 = C2 ^ ROL64(C4, 1);
        unsigned long long D4 = C3 ^ ROL64(C0, 1);

        unsigned long long B[25];
        unsigned long long A[25];

        A[0]  = s[0]  ^ D0; A[1]  = s[1]  ^ D1; A[2]  = s[2]  ^ D2; A[3]  = s[3]  ^ D3; A[4]  = s[4]  ^ D4;
        A[5]  = s[5]  ^ D0; A[6]  = s[6]  ^ D1; A[7]  = s[7]  ^ D2; A[8]  = s[8]  ^ D3; A[9]  = s[9]  ^ D4;
        A[10] = s[10] ^ D0; A[11] = s[11] ^ D1; A[12] = s[12] ^ D2; A[13] = s[13] ^ D3; A[14] = s[14] ^ D4;
        A[15] = s[15] ^ D0; A[16] = s[16] ^ D1; A[17] = s[17] ^ D2; A[18] = s[18] ^ D3; A[19] = s[19] ^ D4;
        A[20] = s[20] ^ D0; A[21] = s[21] ^ D1; A[22] = s[22] ^ D2; A[23] = s[23] ^ D3; A[24] = s[24] ^ D4;

        // 2. Rho & Pi step (direct register transfers with constant rotations)
        B[0]  = A[0];
        B[10] = ROL64(A[1], 1);
        B[7]  = ROL64(A[10], 3);
        B[11] = ROL64(A[7], 6);
        B[17] = ROL64(A[11], 10);
        B[18] = ROL64(A[17], 15);
        B[3]  = ROL64(A[18], 21);
        B[5]  = ROL64(A[3], 28);
        B[16] = ROL64(A[5], 36);
        B[8]  = ROL64(A[16], 45);
        B[21] = ROL64(A[8], 55);
        B[24] = ROL64(A[21], 2);
        B[4]  = ROL64(A[24], 14);
        B[15] = ROL64(A[4], 27);
        B[23] = ROL64(A[15], 41);
        B[19] = ROL64(A[23], 56);
        B[13] = ROL64(A[19], 8);
        B[12] = ROL64(A[13], 25);
        B[2]  = ROL64(A[12], 43);
        B[20] = ROL64(A[2], 62);
        B[14] = ROL64(A[20], 18);
        B[22] = ROL64(A[14], 39);
        B[9]  = ROL64(A[22], 61);
        B[6]  = ROL64(A[9], 20);
        B[1]  = ROL64(A[6], 44);

        // 3. Chi step (unrolled rows)
        s[0]  = B[0]  ^ ((~B[1])  & B[2]);
        s[1]  = B[1]  ^ ((~B[2])  & B[3]);
        s[2]  = B[2]  ^ ((~B[3])  & B[4]);
        s[3]  = B[3]  ^ ((~B[4])  & B[0]);
        s[4]  = B[4]  ^ ((~B[0])  & B[1]);

        s[5]  = B[5]  ^ ((~B[6])  & B[7]);
        s[6]  = B[6]  ^ ((~B[7])  & B[8]);
        s[7]  = B[7]  ^ ((~B[8])  & B[9]);
        s[8]  = B[8]  ^ ((~B[9])  & B[5]);
        s[9]  = B[9]  ^ ((~B[5])  & B[6]);

        s[10] = B[10] ^ ((~B[11]) & B[12]);
        s[11] = B[11] ^ ((~B[12]) & B[13]);
        s[12] = B[12] ^ ((~B[13]) & B[14]);
        s[13] = B[13] ^ ((~B[14]) & B[10]);
        s[14] = B[14] ^ ((~B[10]) & B[11]);

        s[15] = B[15] ^ ((~B[16]) & B[17]);
        s[16] = B[16] ^ ((~B[17]) & B[18]);
        s[17] = B[17] ^ ((~B[18]) & B[19]);
        s[18] = B[18] ^ ((~B[19]) & B[15]);
        s[19] = B[19] ^ ((~B[15]) & B[16]);

        s[20] = B[20] ^ ((~B[21]) & B[22]);
        s[21] = B[21] ^ ((~B[22]) & B[23]);
        s[22] = B[22] ^ ((~B[23]) & B[24]);
        s[23] = B[23] ^ ((~B[24]) & B[20]);
        s[24] = B[24] ^ ((~B[20]) & B[21]);

        // 4. Iota step
        s[0] ^= keccak_rndc[r];
    }
}

void qadav_sha3_256(const unsigned char* input, unsigned long long input_len,
                    unsigned char output[QADAV_SHA3_256_SIZE]) {
    unsigned long long s[25];
    for (int i = 0; i < 25; i++) s[i] = 0;

    unsigned long long pos = 0;
    while (input_len >= QADAV_SHA3_256_RATE) {
        for (int i = 0; i < (QADAV_SHA3_256_RATE / 8); i++) {
            unsigned long long word = 0;
            for (int b = 0; b < 8; b++) {
                word |= ((unsigned long long)input[pos + i * 8 + b]) << (8 * b);
            }
            s[i] ^= word;
        }
        keccak_f1600(s);
        pos += QADAV_SHA3_256_RATE;
        input_len -= QADAV_SHA3_256_RATE;
    }

    // Pad with 0x06 ... 0x80
    unsigned char block[QADAV_SHA3_256_RATE];
    for (int i = 0; i < QADAV_SHA3_256_RATE; i++) block[i] = 0;
    for (unsigned long long i = 0; i < input_len; i++) {
        block[i] = input[pos + i];
    }
    block[input_len] = 0x06;
    block[QADAV_SHA3_256_RATE - 1] |= 0x80;

    for (int i = 0; i < (QADAV_SHA3_256_RATE / 8); i++) {
        unsigned long long word = 0;
        for (int b = 0; b < 8; b++) {
            word |= ((unsigned long long)block[i * 8 + b]) << (8 * b);
        }
        s[i] ^= word;
    }
    keccak_f1600(s);

    // Extract first 32 bytes (4 64-bit words)
    for (int i = 0; i < 4; i++) {
        for (int b = 0; b < 8; b++) {
            output[i * 8 + b] = (unsigned char)((s[i] >> (8 * b)) & 0xFF);
        }
    }
}

int qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len) {
    unsigned char diff = 0;
    for (int i = 0; i < len; i++) {
        diff |= (a[i] ^ b[i]);
    }
    return diff == 0 ? 0 : 1;
}

int qadav_memcmp_const(const void* a, const void* b, unsigned long long len) {
    const unsigned char* pa = (const unsigned char*)a;
    const unsigned char* pb = (const unsigned char*)b;
    unsigned char diff = 0;
    for (unsigned long long i = 0; i < len; i++) {
        diff |= (pa[i] ^ pb[i]);
    }
    return diff == 0 ? 1 : 0;
}
