#ifndef QADAV_PRIME_H
#define QADAV_PRIME_H

/*
 * qadav_prime.h — Prime-Symbolic Knowledge Engine
 * Zero-libc: no <math.h>, <stdio.h>, <stdlib.h>
 * I/O via qadav_emit_*. Memory via qadav_alloc/qadav_free.
 */

#define QADAV_PRIME_COUNT 10000

/* Check if n is prime using the precomputed table. */
int qadav_is_prime(long long n);

/*
 * Factorize n into prime factors.
 * Stores factors in out[0..max_out-1]. Returns count stored.
 */
int qadav_factorize(long long n, long long *out, int max_out);

/*
 * Compose concept: multiply an array of prime values.
 */
long long qadav_encode_concept(const long long *primes_in, int count);

/*
 * Print a concept's prime factorisation via qadav_emit_int (one factor per call).
 */
void qadav_print_concept(long long n);

#endif /* QADAV_PRIME_H */
