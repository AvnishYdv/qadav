#include "qadav_scoped.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void *qadav_alloc(unsigned long long sz);
void qadav_free(void *p);
void qadav_write_raw(const char *s, unsigned long long len);
int qadav_sha3_ct_compare(const unsigned char *a, const unsigned char *b,
                          int len);

static void scoped_memcpy(void *dst, const void *src, unsigned long long n) {
  volatile unsigned char *d = (volatile unsigned char *)dst;
  const volatile unsigned char *s = (const volatile unsigned char *)src;
  for (unsigned long long i = 0; i < n; i++) {
    d[i] = s[i];
  }
}

static void scoped_memset(void *dst, int v, unsigned long long n) {
  volatile unsigned char *d = (volatile unsigned char *)dst;
  for (unsigned long long i = 0; i < n; i++) {
    d[i] = (unsigned char)v;
  }
}

static unsigned long long scoped_strlen(const char *s) {
  if (!s)
    return 0;
  unsigned long long len = 0;
  while (s[len])
    len++;
  return len;
}

static int scoped_streq(const char *a, const char *b) {
  if (!a || !b)
    return 0;
  while (*a && *b && *a == *b) {
    a++;
    b++;
  }
  return (*a == *b);
}

// GF(256) arithmetic for Shamir reconstruction
static unsigned char gf256_mul_scoped(unsigned char a, unsigned char b) {
  unsigned char p = 0;
  for (int i = 0; i < 8; i++) {
    if (b & 1)
      p ^= a;
    int hi = (a & 0x80);
    a <<= 1;
    if (hi)
      a ^= 0x1B;
    b >>= 1;
  }
  return p;
}

static unsigned char gf256_inv_scoped(unsigned char a) {
  if (a == 0)
    return 0;
  unsigned char res = 1;
  unsigned char base = a;
  unsigned char exp = 254; // a^254 in GF(2^8)
  while (exp > 0) {
    if (exp & 1)
      res = gf256_mul_scoped(res, base);
    base = gf256_mul_scoped(base, base);
    exp >>= 1;
  }
  return res;
}

qadav_scoped_token *qadav_scoped_issue(const char *name, void *owner_boundary,
                                       const unsigned char recipient_hash[32],
                                       const char *const *allowed_keys,
                                       int num_keys,
                                       unsigned long long expiry_seconds) {
  qadav_scoped_token *tok =
      (qadav_scoped_token *)qadav_alloc(sizeof(qadav_scoped_token));
  if (!tok)
    return 0;
  scoped_memset(tok, 0, sizeof(qadav_scoped_token));

  if (name) {
    unsigned long long nl = scoped_strlen(name);
    if (nl >= 64)
      nl = 63;
    scoped_memcpy(tok->name, name, nl);
    tok->name[nl] = '\0';
  }

  if (owner_boundary) {
    qadav_qeeh_boundary_hash(owner_boundary, tok->boundary_hash);
  }

  if (recipient_hash) {
    scoped_memcpy(tok->recipient_hash, recipient_hash, 32);
  }

  tok->num_keys =
      (num_keys > QADAV_SCOPED_MAX_KEYS) ? QADAV_SCOPED_MAX_KEYS : num_keys;
  unsigned char leaves[QADAV_SCOPED_MAX_KEYS * 32];
  scoped_memset(leaves, 0, sizeof(leaves));

  for (int i = 0; i < tok->num_keys; i++) {
    if (allowed_keys && allowed_keys[i]) {
      unsigned long long kl = scoped_strlen(allowed_keys[i]);
      if (kl >= QADAV_SCOPED_KEY_LEN)
        kl = QADAV_SCOPED_KEY_LEN - 1;
      scoped_memcpy(tok->allowed_keys[i], allowed_keys[i], kl);
      tok->allowed_keys[i][kl] = '\0';

      // Hash leaf: SHA3(key || index)
      unsigned char leaf_input[QADAV_SCOPED_KEY_LEN + 4];
      scoped_memcpy(leaf_input, tok->allowed_keys[i], kl);
      leaf_input[kl] = (unsigned char)(i & 0xFF);
      qadav_sha3_256(leaf_input, kl + 1, leaves + i * 32);
    }
  }

  // Merkle root of allowed keys
  if (tok->num_keys > 0) {
    qadav_sha3_256(leaves, tok->num_keys * 32, tok->keys_root);
  }

  tok->issued_at = 1000;
  tok->expiry = tok->issued_at + expiry_seconds;
  tok->revoked = 0;

  // Token ID: SHA3(boundary_hash (32) || keys_root (32) || recipient_hash (32)
  // || expiry (8))
  unsigned char id_preimage[104];
  scoped_memcpy(id_preimage, tok->boundary_hash, 32);
  scoped_memcpy(id_preimage + 32, tok->keys_root, 32);
  scoped_memcpy(id_preimage + 64, tok->recipient_hash, 32);
  scoped_memcpy(id_preimage + 96, &tok->expiry, 8);
  qadav_sha3_256(id_preimage, 104, tok->token_id);

  return tok;
}

int qadav_scoped_grant(qadav_scoped_token *token,
                       const unsigned char recipient_hash[32]) {
  if (!token)
    return -1;
  if (recipient_hash) {
    scoped_memcpy(token->recipient_hash, recipient_hash, 32);
  }
  return 0;
}

int qadav_scoped_revoke(qadav_scoped_token *token) {
  if (!token)
    return -1;
  token->revoked = 1;
  return 0;
}

int qadav_scoped_verify(const qadav_scoped_token *token,
                        const unsigned char recipient_hash[32],
                        const char *requested_key,
                        unsigned long long current_time) {
  if (!token)
    return -1;

  // 1. Check revocation
  if (token->revoked) {
    return -2; // Revoked
  }

  // 2. Check expiration
  if (current_time > token->expiry) {
    return -3; // Expired
  }

  // 3. Constant-time recipient binding check
  if (recipient_hash) {
    if (qadav_sha3_ct_compare(recipient_hash, token->recipient_hash, 32) != 0) {
      return -4; // Recipient mismatch
    }
  }

  // 4. Check requested key in allowed keys
  if (!requested_key)
    return -5;
  int found = 0;
  for (int i = 0; i < token->num_keys; i++) {
    if (scoped_streq(token->allowed_keys[i], requested_key)) {
      found = 1;
      break;
    }
  }
  if (!found) {
    return -5; // Key not in allowed set
  }

  return 0;
}

int qadav_scoped_recall(qadav_memory_provenance *mp,
                        const qadav_scoped_token *token,
                        const unsigned char recipient_hash[32],
                        const char *requested_key,
                        unsigned long long current_time, void **out_value,
                        unsigned long long *out_len) {
  if (!mp || !token || !requested_key)
    return -1;

  // 1. Verify scoped token
  int v =
      qadav_scoped_verify(token, recipient_hash, requested_key, current_time);
  if (v != 0) {
    if (v == -2 || v == -4) {
      // Revoked token or recipient mismatch: trigger active decoherence
      qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
      ExitProcess((unsigned int)0xDEADBEEF);
#endif
    }
    return v;
  }

  // 2. Validate token was issued for this memory provenance vault
  if (qadav_sha3_ct_compare(token->boundary_hash, mp->auth_boundary_id, 32) !=
      0) {
    qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
    ExitProcess((unsigned int)0xDEADBEEF);
#endif
    return -6; // Boundary mismatch
  }

  // 3. Locate entry in memory provenance
  qadav_memory_entry *e = 0;
  for (int i = 0; i < mp->count; i++) {
    if (scoped_streq(mp->entries[i].key, requested_key)) {
      e = &mp->entries[i];
      break;
    }
  }
  if (!e || e->is_locked == 2)
    return -7; // Not found or wiped

  // 4. Verify share commitments using Shamir threshold shares (0, 1, 2)
  int k = QADAV_QEMP_THRESHOLD;
  for (int s = 0; s < k; s++) {
    if (!e->shares[s]) {
      qadav_qeeh_trigger_decoherence(0);
      return -8; // Tamper
    }
    unsigned char c_buf[48];
    unsigned char share_head[32];
    qadav_sha3_256(e->shares[s], e->len, share_head);
    scoped_memcpy(c_buf, share_head, 32);
    int slot = (int)(e - mp->entries);
    c_buf[32] = (unsigned char)(slot & 0xFF);
    c_buf[33] = (unsigned char)(s & 0xFF);
    unsigned char comp_c[32];
    qadav_sha3_256(c_buf, 34, comp_c);
    if (qadav_sha3_ct_compare(comp_c, e->commitments[s], 32) != 0) {
      qadav_qeeh_trigger_decoherence(0);
      return -8; // Tamper
    }
  }

  // 5. Lagrange interpolation over GF(256) at x = 0
  unsigned char basis[QADAV_QEMP_THRESHOLD];
  for (int i = 0; i < k; i++) {
    unsigned char xi = (unsigned char)(i + 1);
    unsigned char num = 1;
    unsigned char den = 1;
    for (int j = 0; j < k; j++) {
      if (i == j)
        continue;
      unsigned char xj = (unsigned char)(j + 1);
      num = gf256_mul_scoped(num, xj);
      den = gf256_mul_scoped(den, xj ^ xi);
    }
    basis[i] = gf256_mul_scoped(num, gf256_inv_scoped(den));
  }

  unsigned char *reconstructed = (unsigned char *)qadav_alloc(e->len + 1);
  if (!reconstructed)
    return -1;

  for (unsigned long long b = 0; b < e->len; b++) {
    unsigned char secret_byte = 0;
    for (int i = 0; i < k; i++) {
      secret_byte ^= gf256_mul_scoped(e->shares[i][b], basis[i]);
    }
    reconstructed[b] = secret_byte;
  }
  reconstructed[e->len] = '\0';

  // 6. Verify reconstructed SHA3-256 matches value_hash
  unsigned char rec_hash[32];
  qadav_sha3_256(reconstructed, e->len, rec_hash);
  if (qadav_sha3_ct_compare(rec_hash, e->value_hash, 32) != 0) {
    qadav_free(reconstructed);
    qadav_qeeh_trigger_decoherence(0);
    return -8; // Tamper
  }

  if (out_value) {
    *out_value = reconstructed;
  } else {
    qadav_free(reconstructed);
  }
  if (out_len) {
    *out_len = e->len;
  }
  return 0;
}

void qadav_scoped_free(qadav_scoped_token *token) {
  if (token) {
    scoped_memset(token, 0, sizeof(qadav_scoped_token));
    qadav_free(token);
  }
}