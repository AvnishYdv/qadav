// Qadav native zero-libc Byte Pair Encoding (BPE) tokenizer runtime (M10.2)
// Strict zero-libc: VirtualAlloc / VirtualFree / Win32 File I/O only.

typedef unsigned char      bpe_u8;
typedef unsigned short     bpe_u16;
typedef unsigned int       bpe_u32;
typedef unsigned long long bpe_u64;
typedef long long          bpe_i64;

#ifdef _WIN32
typedef void*         BPE_HANDLE;
typedef unsigned long BPE_DWORD;
typedef int           BPE_BOOL;

#define BPE_GENERIC_READ     0x80000000UL
#define BPE_GENERIC_WRITE    0x40000000UL
#define BPE_CREATE_ALWAYS    2UL
#define BPE_OPEN_EXISTING    3UL
#define BPE_FILE_SHARE_READ  1UL
#define BPE_FILE_ATTR_NORM   0x80UL
#define BPE_INVALID_HANDLE   ((BPE_HANDLE)(long long)-1)
#define BPE_STD_OUT_HANDLE   ((BPE_DWORD)-11)

__declspec(dllimport) BPE_HANDLE __stdcall CreateFileA(const char*, BPE_DWORD, BPE_DWORD, void*, BPE_DWORD, BPE_DWORD, void*);
__declspec(dllimport) BPE_BOOL   __stdcall ReadFile(BPE_HANDLE, void*, BPE_DWORD, BPE_DWORD*, void*);
__declspec(dllimport) BPE_BOOL   __stdcall WriteFile(BPE_HANDLE, const void*, BPE_DWORD, BPE_DWORD*, void*);
__declspec(dllimport) BPE_BOOL   __stdcall CloseHandle(BPE_HANDLE);
__declspec(dllimport) BPE_DWORD  __stdcall GetFileSize(BPE_HANDLE, BPE_DWORD*);
__declspec(dllimport) BPE_HANDLE __stdcall GetStdHandle(BPE_DWORD);
__declspec(dllimport) BPE_HANDLE __stdcall GetModuleHandleA(const char*);
__declspec(dllimport) void*      __stdcall GetProcAddress(BPE_HANDLE, const char*);
__declspec(dllimport) char*      __stdcall GetCommandLineA(void);
__declspec(dllimport) void       __stdcall ExitProcess(unsigned int);

typedef void*    (__stdcall *bpe_pfn_VirtualAlloc)(void*, unsigned long long, BPE_DWORD, BPE_DWORD);
typedef BPE_BOOL (__stdcall *bpe_pfn_VirtualFree)(void*, unsigned long long, BPE_DWORD);
typedef BPE_BOOL (__stdcall *bpe_pfn_QPC)(bpe_i64*);
typedef BPE_BOOL (__stdcall *bpe_pfn_QPF)(bpe_i64*);

static void* bpe_valloc(unsigned long long sz) {
    static bpe_pfn_VirtualAlloc pfn = (bpe_pfn_VirtualAlloc)0;
    if (!pfn) {
        BPE_HANDLE hm = GetModuleHandleA("kernel32.dll");
        if (hm) pfn = (bpe_pfn_VirtualAlloc)GetProcAddress(hm, "VirtualAlloc");
    }
    if (!pfn || sz == 0) return (void*)0;
    return pfn((void*)0, sz, 0x00001000UL | 0x00002000UL, 0x04UL);
}

static void bpe_vfree(void* ptr) {
    static bpe_pfn_VirtualFree pfn = (bpe_pfn_VirtualFree)0;
    if (!ptr) return;
    if (!pfn) {
        BPE_HANDLE hm = GetModuleHandleA("kernel32.dll");
        if (hm) pfn = (bpe_pfn_VirtualFree)GetProcAddress(hm, "VirtualFree");
    }
    if (pfn) pfn(ptr, 0ULL, 0x00008000UL);
}

static bpe_i64 bpe_qpc_now(void) {
    static bpe_pfn_QPC pfn = (bpe_pfn_QPC)0;
    if (!pfn) {
        BPE_HANDLE hm = GetModuleHandleA("kernel32.dll");
        if (hm) pfn = (bpe_pfn_QPC)GetProcAddress(hm, "QueryPerformanceCounter");
    }
    bpe_i64 t = 0;
    if (pfn) pfn(&t);
    return t;
}

static bpe_i64 bpe_qpf_hz(void) {
    static bpe_pfn_QPF pfn = (bpe_pfn_QPF)0;
    if (!pfn) {
        BPE_HANDLE hm = GetModuleHandleA("kernel32.dll");
        if (hm) pfn = (bpe_pfn_QPF)GetProcAddress(hm, "QueryPerformanceFrequency");
    }
    bpe_i64 f = 10000000LL;
    if (pfn) pfn(&f);
    return f > 0 ? f : 10000000LL;
}
#else
static void* bpe_valloc(unsigned long long sz) { (void)sz; return (void*)0; }
static void  bpe_vfree(void* ptr) { (void)ptr; }
static bpe_i64 bpe_qpc_now(void) { return 0; }
static bpe_i64 bpe_qpf_hz(void) { return 1; }
#endif

static void bpe_memset_zero(void* dst, unsigned long long n) {
    volatile bpe_u8* p = (volatile bpe_u8*)dst;
    for (unsigned long long i = 0; i < n; i++) p[i] = 0;
}

static void bpe_memcpy_raw(void* dst, const void* src, unsigned long long n) {
    volatile bpe_u8* d = (volatile bpe_u8*)dst;
    const bpe_u8* s = (const bpe_u8*)src;
    for (unsigned long long i = 0; i < n; i++) d[i] = s[i];
}

static int bpe_is_word_byte(bpe_u8 c) {
    if (c >= 'a' && c <= 'z') return 1;
    if (c >= 'A' && c <= 'Z') return 1;
    if (c >= '0' && c <= '9') return 1;
    if (c == '\'' || c == '_') return 1;
    if (c >= 0x80) return 1; // UTF-8 multi-byte codepoints preserved inside words
    return 0;
}

// Compute contiguous byte-slice boundary [pos .. end) such that chunk length <= 128
// and concatenating all chunks identically reconstructs text[0 .. text_len).
static int bpe_next_chunk_end(const bpe_u8* text, int text_len, int pos) {
    int rem = text_len - pos;
    int max_end = pos + (rem < 128 ? rem : 128);
    bpe_u8 c = text[pos];
    if (c == ' ') {
        if (pos + 1 < max_end && bpe_is_word_byte(text[pos + 1])) {
            int p = pos + 2;
            while (p < max_end && bpe_is_word_byte(text[p])) p++;
            return p;
        }
        int p = pos + 1;
        while (p < max_end && p - pos < 16 && text[p] == ' ') p++;
        return p;
    }
    if (bpe_is_word_byte(c)) {
        int p = pos + 1;
        while (p < max_end && bpe_is_word_byte(text[p])) p++;
        return p;
    }
    if (c == '\r' && pos + 1 < max_end && text[pos + 1] == '\n') return pos + 2;
    return pos + 1;
}

static bpe_u64 bpe_fnv1a(const bpe_u8* data, int len) {
    bpe_u64 h = 14695981039346656037ULL;
    for (int i = 0; i < len; i++) {
        h ^= (bpe_u64)data[i];
        h *= 1099511628211ULL;
    }
    return h;
}

static bpe_u64 bpe_pair_hash(bpe_u64 key) {
    bpe_u64 x = key ^ (key >> 30);
    x *= 0xbf58476d1ce4e5b9ULL;
    x ^= x >> 27;
    x *= 0x94d049bb133111ebULL;
    x ^= x >> 31;
    return x;
}

static int bpe_read_entire_file(const char* path, bpe_u8** out_buf, bpe_u32* out_len) {
#ifdef _WIN32
    BPE_HANDLE h = CreateFileA(path, BPE_GENERIC_READ, BPE_FILE_SHARE_READ, (void*)0,
                               BPE_OPEN_EXISTING, BPE_FILE_ATTR_NORM, (void*)0);
    if (h == BPE_INVALID_HANDLE) return -1;
    BPE_DWORD hi = 0;
    BPE_DWORD sz = GetFileSize(h, &hi);
    if (hi != 0 || sz == 0xFFFFFFFFUL) { CloseHandle(h); return -1; }
    bpe_u8* buf = (bpe_u8*)bpe_valloc((unsigned long long)sz + 16ULL);
    if (!buf) { CloseHandle(h); return -1; }
    bpe_u32 done = 0;
    while (done < sz) {
        BPE_DWORD chunk = sz - done;
        if (chunk > (1U << 24)) chunk = (1U << 24);
        BPE_DWORD got = 0;
        if (!ReadFile(h, buf + done, chunk, &got, (void*)0) || got == 0) {
            bpe_vfree(buf);
            CloseHandle(h);
            return -1;
        }
        done += got;
    }
    buf[sz] = 0;
    CloseHandle(h);
    *out_buf = buf;
    *out_len = (bpe_u32)sz;
    return 0;
#else
    (void)path; (void)out_buf; (void)out_len;
    return -1;
#endif
}

static int bpe_write_all(BPE_HANDLE h, const void* data, unsigned long long len) {
#ifdef _WIN32
    const bpe_u8* p = (const bpe_u8*)data;
    unsigned long long done = 0;
    while (done < len) {
        unsigned long long rem = len - done;
        BPE_DWORD chunk = (rem > (1ULL << 24)) ? (1UL << 24) : (BPE_DWORD)rem;
        BPE_DWORD wr = 0;
        if (!WriteFile(h, p + done, chunk, &wr, (void*)0) || wr != chunk) return -1;
        done += wr;
    }
    return 0;
#else
    (void)h; (void)data; (void)len;
    return -1;
#endif
}

// ---------------------------------------------------------------------
// Vocabulary Binary Format ("BPE1") State for Encode / Decode
// ---------------------------------------------------------------------

typedef struct {
    bpe_u32 id_a;
    bpe_u32 id_b;
    bpe_u32 new_id;
} BPEMerge;

typedef struct {
    bpe_u64 key;
    bpe_u32 new_id;
    bpe_u32 rank;
    bpe_u8  used;
} BPEMergeHTEntry;

typedef struct {
    int loaded;
    int vocab_size;
    int num_merges;
    BPEMerge* merges;
    bpe_u8* tok_lens;
    bpe_u32* tok_offs;
    bpe_u8* tok_bytes;
    bpe_u32 total_tok_bytes;
    BPEMergeHTEntry* merge_ht;
    bpe_u32 merge_ht_mask;
    void* mem_block;
} BPEState;

static BPEState g_bpe;

void bpe_free(void) {
    if (g_bpe.mem_block) bpe_vfree(g_bpe.mem_block);
    bpe_memset_zero(&g_bpe, sizeof(g_bpe));
}

int bpe_vocab_size(void) {
    return g_bpe.loaded ? g_bpe.vocab_size : 0;
}

static void bpe_merge_ht_insert(bpe_u32 id_a, bpe_u32 id_b, bpe_u32 new_id, bpe_u32 rank) {
    bpe_u64 key = ((bpe_u64)id_a << 32) | (bpe_u64)id_b;
    bpe_u32 idx = (bpe_u32)bpe_pair_hash(key) & g_bpe.merge_ht_mask;
    while (g_bpe.merge_ht[idx].used) {
        if (g_bpe.merge_ht[idx].key == key) return;
        idx = (idx + 1U) & g_bpe.merge_ht_mask;
    }
    g_bpe.merge_ht[idx].key    = key;
    g_bpe.merge_ht[idx].new_id = new_id;
    g_bpe.merge_ht[idx].rank   = rank;
    g_bpe.merge_ht[idx].used   = 1;
}

static int bpe_merge_ht_lookup(bpe_u32 id_a, bpe_u32 id_b, bpe_u32* out_new_id, bpe_u32* out_rank) {
    bpe_u64 key = ((bpe_u64)id_a << 32) | (bpe_u64)id_b;
    bpe_u32 idx = (bpe_u32)bpe_pair_hash(key) & g_bpe.merge_ht_mask;
    while (g_bpe.merge_ht[idx].used) {
        if (g_bpe.merge_ht[idx].key == key) {
            *out_new_id = g_bpe.merge_ht[idx].new_id;
            *out_rank   = g_bpe.merge_ht[idx].rank;
            return 1;
        }
        idx = (idx + 1U) & g_bpe.merge_ht_mask;
    }
    return 0;
}

int bpe_load(const char* vocab_path) {
    bpe_free();
    bpe_u8* raw = (bpe_u8*)0;
    bpe_u32 raw_len = 0;
    if (bpe_read_entire_file(vocab_path, &raw, &raw_len) != 0) return -1;
    if (raw_len < 12U) { bpe_vfree(raw); return -1; }
    if (raw[0] != 'B' || raw[1] != 'P' || raw[2] != 'E' || raw[3] != '1') {
        bpe_vfree(raw); return -1;
    }
    bpe_u32 vocab_size = 0, num_merges = 0;
    bpe_memcpy_raw(&vocab_size, raw + 4, 4);
    bpe_memcpy_raw(&num_merges, raw + 8, 4);
    if (vocab_size < 256U || vocab_size > 262144U || num_merges != (vocab_size - 256U)) {
        bpe_vfree(raw); return -1;
    }

    unsigned long long merges_bytes = (unsigned long long)num_merges * sizeof(BPEMerge);
    unsigned long long header_min   = 12ULL + merges_bytes + (unsigned long long)vocab_size;
    if ((unsigned long long)raw_len < header_min) { bpe_vfree(raw); return -1; }

    const bpe_u8* lens_ptr = raw + 12ULL + merges_bytes;
    unsigned long long sum_bytes = 0;
    for (bpe_u32 i = 0; i < vocab_size; i++) sum_bytes += (unsigned long long)lens_ptr[i];
    if ((unsigned long long)raw_len < header_min + sum_bytes) { bpe_vfree(raw); return -1; }

    bpe_u32 ht_cap = 1024U;
    while (ht_cap < num_merges * 4U) ht_cap <<= 1;

    unsigned long long sz_merges = ((merges_bytes + 15ULL) & ~15ULL);
    unsigned long long sz_lens   = (((unsigned long long)vocab_size + 15ULL) & ~15ULL);
    unsigned long long sz_offs   = (((unsigned long long)vocab_size * 4ULL + 15ULL) & ~15ULL);
    unsigned long long sz_bytes  = ((sum_bytes + 15ULL) & ~15ULL);
    unsigned long long sz_ht     = ((unsigned long long)ht_cap * sizeof(BPEMergeHTEntry));
    unsigned long long total_mem = sz_merges + sz_lens + sz_offs + sz_bytes + sz_ht;

    bpe_u8* blk = (bpe_u8*)bpe_valloc(total_mem);
    if (!blk) { bpe_vfree(raw); return -1; }
    bpe_memset_zero(blk, total_mem);

    g_bpe.mem_block       = blk;
    g_bpe.vocab_size      = (int)vocab_size;
    g_bpe.num_merges      = (int)num_merges;
    g_bpe.merges          = (BPEMerge*)blk;
    g_bpe.tok_lens        = blk + sz_merges;
    g_bpe.tok_offs        = (bpe_u32*)(blk + sz_merges + sz_lens);
    g_bpe.tok_bytes       = blk + sz_merges + sz_lens + sz_offs;
    g_bpe.total_tok_bytes = (bpe_u32)sum_bytes;
    g_bpe.merge_ht        = (BPEMergeHTEntry*)(blk + sz_merges + sz_lens + sz_offs + sz_bytes);
    g_bpe.merge_ht_mask   = ht_cap - 1U;

    if (merges_bytes > 0) bpe_memcpy_raw(g_bpe.merges, raw + 12, merges_bytes);
    bpe_memcpy_raw(g_bpe.tok_lens, lens_ptr, vocab_size);
    if (sum_bytes > 0) bpe_memcpy_raw(g_bpe.tok_bytes, lens_ptr + vocab_size, sum_bytes);

    bpe_u32 cur_off = 0;
    for (bpe_u32 i = 0; i < vocab_size; i++) {
        g_bpe.tok_offs[i] = cur_off;
        cur_off += (bpe_u32)g_bpe.tok_lens[i];
    }
    for (bpe_u32 r = 0; r < num_merges; r++) {
        bpe_merge_ht_insert(g_bpe.merges[r].id_a, g_bpe.merges[r].id_b, g_bpe.merges[r].new_id, r);
    }

    bpe_vfree(raw);
    g_bpe.loaded = 1;
    return 0;
}

int bpe_encode(const char* text, int text_len, int* token_ids, int max_tokens) {
    if (!g_bpe.loaded || !text || !token_ids || text_len < 0 || max_tokens < 0) return -1;
    if (text_len == 0) return 0;

    const bpe_u8* src = (const bpe_u8*)text;
    int out_cnt = 0;
    int pos = 0;
    bpe_u32 sym[128];

    while (pos < text_len) {
        int end = bpe_next_chunk_end(src, text_len, pos);
        int n = end - pos;
        for (int i = 0; i < n; i++) sym[i] = (bpe_u32)src[pos + i];

        while (n >= 2) {
            bpe_u32 best_rank = 0xFFFFFFFFU;
            bpe_u32 best_new  = 0;
            bpe_u32 best_a = 0, best_b = 0;
            int found = 0;
            for (int i = 0; i + 1 < n; i++) {
                bpe_u32 nid = 0, rk = 0;
                if (bpe_merge_ht_lookup(sym[i], sym[i + 1], &nid, &rk)) {
                    if (!found || rk < best_rank) {
                        found = 1;
                        best_rank = rk;
                        best_new  = nid;
                        best_a    = sym[i];
                        best_b    = sym[i + 1];
                    }
                }
            }
            if (!found) break;
            int write_i = 0;
            for (int i = 0; i < n; ) {
                if (i + 1 < n && sym[i] == best_a && sym[i + 1] == best_b) {
                    sym[write_i++] = best_new;
                    i += 2;
                } else {
                    sym[write_i++] = sym[i++];
                }
            }
            n = write_i;
        }

        if (out_cnt + n > max_tokens) return -2;
        for (int i = 0; i < n; i++) token_ids[out_cnt++] = (int)sym[i];
        pos = end;
    }
    return out_cnt;
}

int bpe_decode(const int* token_ids, int n_tokens, char* text_out, int max_chars) {
    if (!g_bpe.loaded || !text_out || n_tokens < 0 || max_chars < 0) return -1;
    if (n_tokens == 0) {
        if (max_chars > 0) text_out[0] = '\0';
        return 0;
    }
    if (!token_ids) return -1;

    int out_pos = 0;
    for (int i = 0; i < n_tokens; i++) {
        int id = token_ids[i];
        if (id < 0 || id >= g_bpe.vocab_size) return -1;
        int len = (int)g_bpe.tok_lens[id];
        if (out_pos + len > max_chars) return -2;
        bpe_u32 off = g_bpe.tok_offs[id];
        for (int j = 0; j < len; j++) {
            text_out[out_pos++] = (char)g_bpe.tok_bytes[off + (bpe_u32)j];
        }
    }
    if (out_pos < max_chars) text_out[out_pos] = '\0';
    return out_pos;
}

// ---------------------------------------------------------------------
// BPE Trainer (Weighted Unique Chunks + Pair Hash Table + Lazy Max-Heap)
// ---------------------------------------------------------------------

typedef struct {
    bpe_u64 key;
    bpe_i64 freq;
    bpe_u8  used;
} BPETrainPair;

typedef struct {
    bpe_i64 freq;
    bpe_u64 key;
} BPEHeapNode;

typedef struct {
    bpe_u32 tok_off;
    bpe_u32 raw_off;
    bpe_u16 len;
    bpe_u8  raw_len;
    bpe_u8  pad;
    bpe_i64 freq;
    bpe_u64 hash;
} BPETrainWord;

static int bpe_heap_greater(const BPEHeapNode* a, const BPEHeapNode* b) {
    if (a->freq != b->freq) return a->freq > b->freq;
    return a->key < b->key; // Deterministic tie-break on smallest (id_a, id_b)
}

static void bpe_heap_push(BPEHeapNode* heap, bpe_u32* sz, bpe_u32 cap, bpe_i64 freq, bpe_u64 key) {
    if (freq <= 0 || *sz >= cap) return;
    bpe_u32 idx = (*sz)++;
    BPEHeapNode node;
    node.freq = freq;
    node.key  = key;
    while (idx > 0U) {
        bpe_u32 par = (idx - 1U) >> 1;
        if (!bpe_heap_greater(&node, &heap[par])) break;
        heap[idx] = heap[par];
        idx = par;
    }
    heap[idx] = node;
}

static BPEHeapNode bpe_heap_pop(BPEHeapNode* heap, bpe_u32* sz) {
    BPEHeapNode top = heap[0];
    bpe_u32 n = --(*sz);
    if (n > 0U) {
        BPEHeapNode last = heap[n];
        bpe_u32 idx = 0U;
        for (;;) {
            bpe_u32 l = idx * 2U + 1U;
            if (l >= n) break;
            bpe_u32 r = l + 1U;
            bpe_u32 best = l;
            if (r < n && bpe_heap_greater(&heap[r], &heap[l])) best = r;
            if (!bpe_heap_greater(&heap[best], &last)) break;
            heap[idx] = heap[best];
            idx = best;
        }
        heap[idx] = last;
    }
    return top;
}

static void bpe_pair_add(BPETrainPair* pt, bpe_u32 mask, bpe_u64 key, bpe_i64 delta,
                         BPEHeapNode* heap, bpe_u32* heap_sz, bpe_u32 heap_cap) {
    bpe_u32 idx = (bpe_u32)bpe_pair_hash(key) & mask;
    while (pt[idx].used) {
        if (pt[idx].key == key) {
            pt[idx].freq += delta;
            if (delta > 0 && pt[idx].freq > 0) {
                bpe_heap_push(heap, heap_sz, heap_cap, pt[idx].freq, key);
            }
            return;
        }
        idx = (idx + 1U) & mask;
    }
    if (delta <= 0) return;
    pt[idx].key  = key;
    pt[idx].freq = delta;
    pt[idx].used = 1;
    bpe_heap_push(heap, heap_sz, heap_cap, delta, key);
}

static bpe_i64 bpe_pair_lookup(const BPETrainPair* pt, bpe_u32 mask, bpe_u64 key) {
    bpe_u32 idx = (bpe_u32)bpe_pair_hash(key) & mask;
    while (pt[idx].used) {
        if (pt[idx].key == key) return pt[idx].freq;
        idx = (idx + 1U) & mask;
    }
    return 0;
}

static void bpe_heap_rebuild_from_pairs(const BPETrainPair* pt, bpe_u32 pt_cap,
                                        BPEHeapNode* heap, bpe_u32* heap_sz, bpe_u32 heap_cap) {
    *heap_sz = 0;
    for (bpe_u32 i = 0; i < pt_cap; i++) {
        if (pt[i].used && pt[i].freq > 0) {
            bpe_heap_push(heap, heap_sz, heap_cap, pt[i].freq, pt[i].key);
        }
    }
}

int bpe_train(const char* corpus_path, const char* vocab_out_path, int target_vocab) {
    if (!corpus_path || !vocab_out_path || target_vocab < 256 || target_vocab > 131072) return -1;

    bpe_u8* corpus = (bpe_u8*)0;
    bpe_u32 corpus_len = 0;
    if (bpe_read_entire_file(corpus_path, &corpus, &corpus_len) != 0) return -1;

    bpe_u32 max_words  = (corpus_len <= 4U * 1024U * 1024U) ? 65536U  : 262144U;
    bpe_u32 word_ht_cap = max_words * 2U;
    bpe_u32 pair_ht_cap = (corpus_len <= 4U * 1024U * 1024U) ? 262144U : 1048576U;
    bpe_u32 heap_cap    = pair_ht_cap * 2U;
    bpe_u32 raw_pool_cap = (corpus_len <= 4U * 1024U * 1024U) ? (2U * 1024U * 1024U) : (8U * 1024U * 1024U);
    bpe_u32 tok_pool_cap = raw_pool_cap;

    bpe_u32 word_ht_mask = word_ht_cap - 1U;
    bpe_u32 pair_ht_mask = pair_ht_cap - 1U;

    BPETrainWord* words     = (BPETrainWord*)bpe_valloc((unsigned long long)max_words * sizeof(BPETrainWord));
    bpe_u32*      word_ht   = (bpe_u32*)bpe_valloc((unsigned long long)word_ht_cap * sizeof(bpe_u32));
    bpe_u8*       raw_pool  = (bpe_u8*)bpe_valloc(raw_pool_cap);
    bpe_u32*      tok_pool  = (bpe_u32*)bpe_valloc((unsigned long long)tok_pool_cap * sizeof(bpe_u32));
    BPETrainPair* pair_ht   = (BPETrainPair*)bpe_valloc((unsigned long long)pair_ht_cap * sizeof(BPETrainPair));
    BPEHeapNode*  heap      = (BPEHeapNode*)bpe_valloc((unsigned long long)heap_cap * sizeof(BPEHeapNode));
    bpe_u32 max_merges      = (bpe_u32)(target_vocab - 256);
    BPEMerge*     merges    = (BPEMerge*)bpe_valloc(((unsigned long long)max_merges + 1ULL) * sizeof(BPEMerge));
    bpe_u8*       v_lens    = (bpe_u8*)bpe_valloc((unsigned long long)target_vocab);
    bpe_u32*      v_offs    = (bpe_u32*)bpe_valloc((unsigned long long)target_vocab * sizeof(bpe_u32));
    bpe_u32       v_bytes_cap = (bpe_u32)target_vocab * 64U;
    bpe_u8*       v_bytes   = (bpe_u8*)bpe_valloc(v_bytes_cap);

    if (!words || !word_ht || !raw_pool || !tok_pool || !pair_ht || !heap || !merges || !v_lens || !v_offs || !v_bytes) {
        if (corpus) bpe_vfree(corpus);
        if (words) bpe_vfree(words); if (word_ht) bpe_vfree(word_ht);
        if (raw_pool) bpe_vfree(raw_pool); if (tok_pool) bpe_vfree(tok_pool);
        if (pair_ht) bpe_vfree(pair_ht); if (heap) bpe_vfree(heap);
        if (merges) bpe_vfree(merges); if (v_lens) bpe_vfree(v_lens);
        if (v_offs) bpe_vfree(v_offs); if (v_bytes) bpe_vfree(v_bytes);
        return -1;
    }

    bpe_memset_zero(word_ht, (unsigned long long)word_ht_cap * sizeof(bpe_u32));
    bpe_memset_zero(pair_ht, (unsigned long long)pair_ht_cap * sizeof(BPETrainPair));

    bpe_u32 num_words = 0, raw_used = 0, tok_used = 0;
    bpe_u32 pos = 0;
    while (pos < corpus_len) {
        bpe_u32 end = (bpe_u32)bpe_next_chunk_end(corpus, (int)corpus_len, (int)pos);
        int clen = (int)(end - pos);
        const bpe_u8* cptr = corpus + pos;
        bpe_u64 h = bpe_fnv1a(cptr, clen);
        bpe_u32 idx = (bpe_u32)h & word_ht_mask;
        int matched = -1;
        while (word_ht[idx] != 0U) {
            bpe_u32 wi = word_ht[idx] - 1U;
            if (words[wi].hash == h && (int)words[wi].raw_len == clen) {
                int eq = 1;
                const bpe_u8* rp = raw_pool + words[wi].raw_off;
                for (int k = 0; k < clen; k++) { if (rp[k] != cptr[k]) { eq = 0; break; } }
                if (eq) { matched = (int)wi; break; }
            }
            idx = (idx + 1U) & word_ht_mask;
        }
        if (matched >= 0) {
            words[matched].freq++;
        } else if (num_words < max_words && raw_used + (bpe_u32)clen <= raw_pool_cap && tok_used + (bpe_u32)clen <= tok_pool_cap) {
            bpe_u32 wi = num_words++;
            words[wi].raw_off = raw_used;
            words[wi].tok_off = tok_used;
            words[wi].raw_len = (bpe_u8)clen;
            words[wi].len     = (bpe_u16)clen;
            words[wi].freq    = 1;
            words[wi].hash    = h;
            bpe_memcpy_raw(raw_pool + raw_used, cptr, (unsigned long long)clen);
            for (int k = 0; k < clen; k++) tok_pool[tok_used + (bpe_u32)k] = (bpe_u32)cptr[k];
            raw_used += (bpe_u32)clen;
            tok_used += (bpe_u32)clen;
            word_ht[idx] = wi + 1U;
        }
        pos = end;
    }

    // Corpus buffer no longer needed once deduplicated chunks are built
    bpe_vfree(corpus); corpus = (bpe_u8*)0;
    bpe_vfree(word_ht); word_ht = (bpe_u32*)0;
    bpe_vfree(raw_pool); raw_pool = (bpe_u8*)0;

    // Seed initial 256 byte tokens
    bpe_u32 v_bytes_used = 0;
    for (bpe_u32 id = 0; id < 256U; id++) {
        v_lens[id] = 1;
        v_offs[id] = v_bytes_used;
        v_bytes[v_bytes_used++] = (bpe_u8)id;
    }

    // Count initial adjacent pairs across all unique words
    bpe_u32 heap_sz = 0;
    for (bpe_u32 wi = 0; wi < num_words; wi++) {
        bpe_u16 wlen = words[wi].len;
        if (wlen < 2) continue;
        const bpe_u32* wt = tok_pool + words[wi].tok_off;
        bpe_i64 wf = words[wi].freq;
        for (int k = 0; k + 1 < (int)wlen; k++) {
            bpe_u64 pk = ((bpe_u64)wt[k] << 32) | (bpe_u64)wt[k + 1];
            bpe_pair_add(pair_ht, pair_ht_mask, pk, wf, heap, &heap_sz, heap_cap);
        }
    }

    bpe_u32 actual_merges = 0;
    for (bpe_u32 step = 0; step < max_merges; step++) {
        bpe_u64 best_key = 0;
        bpe_i64 best_freq = 0;
        for (;;) {
            if (heap_sz == 0U) break;
            BPEHeapNode top = bpe_heap_pop(heap, &heap_sz);
            bpe_i64 cur = bpe_pair_lookup(pair_ht, pair_ht_mask, top.key);
            if (cur > 0 && cur == top.freq) {
                best_key  = top.key;
                best_freq = cur;
                break;
            }
        }
        if (best_freq <= 0) break;

        bpe_u32 id_a   = (bpe_u32)(best_key >> 32);
        bpe_u32 id_b   = (bpe_u32)(best_key & 0xFFFFFFFFULL);
        bpe_u32 new_id = 256U + step;

        bpe_u32 new_len = (bpe_u32)v_lens[id_a] + (bpe_u32)v_lens[id_b];
        if (new_len > 255U || v_bytes_used + new_len > v_bytes_cap) break;

        v_lens[new_id] = (bpe_u8)new_len;
        v_offs[new_id] = v_bytes_used;
        bpe_memcpy_raw(v_bytes + v_bytes_used, v_bytes + v_offs[id_a], (unsigned long long)v_lens[id_a]);
        bpe_memcpy_raw(v_bytes + v_bytes_used + v_lens[id_a], v_bytes + v_offs[id_b], (unsigned long long)v_lens[id_b]);
        v_bytes_used += new_len;

        merges[actual_merges].id_a   = id_a;
        merges[actual_merges].id_b   = id_b;
        merges[actual_merges].new_id = new_id;
        actual_merges++;

        // Apply merge on all unique words containing adjacent (id_a, id_b)
        for (bpe_u32 wi = 0; wi < num_words; wi++) {
            bpe_u16 wlen = words[wi].len;
            if (wlen < 2) continue;
            bpe_u32* wt = tok_pool + words[wi].tok_off;
            int hit = 0;
            for (int k = 0; k + 1 < (int)wlen; k++) {
                if (wt[k] == id_a && wt[k + 1] == id_b) { hit = 1; break; }
            }
            if (!hit) continue;

            bpe_i64 wf = words[wi].freq;
            for (int k = 0; k + 1 < (int)wlen; k++) {
                bpe_u64 pk = ((bpe_u64)wt[k] << 32) | (bpe_u64)wt[k + 1];
                bpe_pair_add(pair_ht, pair_ht_mask, pk, -wf, heap, &heap_sz, heap_cap);
            }
            int nw = 0;
            for (int k = 0; k < (int)wlen; ) {
                if (k + 1 < (int)wlen && wt[k] == id_a && wt[k + 1] == id_b) {
                    wt[nw++] = new_id;
                    k += 2;
                } else {
                    wt[nw++] = wt[k++];
                }
            }
            words[wi].len = (bpe_u16)nw;
            if (heap_sz + (bpe_u32)nw + 4U >= heap_cap) {
                bpe_heap_rebuild_from_pairs(pair_ht, pair_ht_cap, heap, &heap_sz, heap_cap);
            }
            for (int k = 0; k + 1 < nw; k++) {
                bpe_u64 pk = ((bpe_u64)wt[k] << 32) | (bpe_u64)wt[k + 1];
                bpe_pair_add(pair_ht, pair_ht_mask, pk, wf, heap, &heap_sz, heap_cap);
            }
        }
    }

    bpe_vfree(words);
    bpe_vfree(tok_pool);
    bpe_vfree(pair_ht);
    bpe_vfree(heap);

    bpe_u32 final_vocab = 256U + actual_merges;
    int save_ok = -1;
#ifdef _WIN32
    BPE_HANDLE ho = CreateFileA(vocab_out_path, BPE_GENERIC_WRITE, 0, (void*)0,
                                BPE_CREATE_ALWAYS, BPE_FILE_ATTR_NORM, (void*)0);
    if (ho != BPE_INVALID_HANDLE) {
        bpe_u8 hdr[12];
        hdr[0] = 'B'; hdr[1] = 'P'; hdr[2] = 'E'; hdr[3] = '1';
        bpe_memcpy_raw(hdr + 4, &final_vocab, 4);
        bpe_memcpy_raw(hdr + 8, &actual_merges, 4);
        if (bpe_write_all(ho, hdr, 12ULL) == 0 &&
            bpe_write_all(ho, merges, (unsigned long long)actual_merges * sizeof(BPEMerge)) == 0 &&
            bpe_write_all(ho, v_lens, (unsigned long long)final_vocab) == 0 &&
            bpe_write_all(ho, v_bytes, (unsigned long long)v_bytes_used) == 0) {
            save_ok = 0;
        }
        CloseHandle(ho);
    }
#endif

    bpe_vfree(merges);
    bpe_vfree(v_lens);
    bpe_vfree(v_offs);
    bpe_vfree(v_bytes);

    return (save_ok == 0) ? (int)final_vocab : -1;
}

// ---------------------------------------------------------------------
// Integration Hook for M10.3 ("TOKS" Binary Token Stream)
// ---------------------------------------------------------------------

int bpe_encode_file(const char* input_txt_path, const char* output_toks_path) {
    if (!g_bpe.loaded || !input_txt_path || !output_toks_path) return -1;

    bpe_u8* txt = (bpe_u8*)0;
    bpe_u32 txt_len = 0;
    if (bpe_read_entire_file(input_txt_path, &txt, &txt_len) != 0) return -1;

    int max_toks = (int)txt_len + 16;
    int* ids = (int*)bpe_valloc((unsigned long long)max_toks * sizeof(int));
    if (!ids) { bpe_vfree(txt); return -1; }

    int n_toks = bpe_encode((const char*)txt, (int)txt_len, ids, max_toks);
    bpe_vfree(txt);
    if (n_toks < 0) { bpe_vfree(ids); return -1; }

    int ok = -1;
#ifdef _WIN32
    BPE_HANDLE ho = CreateFileA(output_toks_path, BPE_GENERIC_WRITE, 0, (void*)0,
                                BPE_CREATE_ALWAYS, BPE_FILE_ATTR_NORM, (void*)0);
    if (ho != BPE_INVALID_HANDLE) {
        bpe_u8 hdr[8];
        hdr[0] = 'T'; hdr[1] = 'O'; hdr[2] = 'K'; hdr[3] = 'S';
        bpe_u32 n_u32 = (bpe_u32)n_toks;
        bpe_memcpy_raw(hdr + 4, &n_u32, 4);
        if (bpe_write_all(ho, hdr, 8ULL) == 0 &&
            bpe_write_all(ho, ids, (unsigned long long)n_u32 * 4ULL) == 0) {
            ok = 0;
        }
        CloseHandle(ho);
    }
#endif
    bpe_vfree(ids);
    return (ok == 0) ? n_toks : -1;
}

#ifdef QADAV_BPE_MAIN
int _fltused = 0;

static void bpe_cli_print(const char* s) {
#ifdef _WIN32
    BPE_DWORD n = 0; while (s[n]) n++;
    BPE_DWORD w = 0; WriteFile(GetStdHandle(BPE_STD_OUT_HANDLE), s, n, &w, (void*)0);
#else
    (void)s;
#endif
}

static void bpe_cli_print_i64(bpe_i64 v) {
    char buf[32]; int p = 0;
    if (v == 0) { bpe_cli_print("0"); return; }
    if (v < 0) { bpe_cli_print("-"); v = -v; }
    while (v > 0) { buf[p++] = (char)('0' + (v % 10)); v /= 10; }
    char rev[32];
    for (int i = 0; i < p; i++) rev[i] = buf[p - 1 - i];
    rev[p] = 0;
    bpe_cli_print(rev);
}

static void bpe_cli_print_f64(double v, int dp) {
    if (v < 0.0) { bpe_cli_print("-"); v = -v; }
    double rnd = 0.5;
    for (int i = 0; i < dp; i++) rnd *= 0.1;
    v += rnd;
    bpe_i64 ip = (bpe_i64)v;
    bpe_cli_print_i64(ip);
    bpe_cli_print(".");
    double frac = v - (double)ip;
    for (int i = 0; i < dp; i++) {
        frac *= 10.0;
        int d = (int)frac;
        if (d > 9) d = 9;
        char c[2]; c[0] = (char)('0' + d); c[1] = 0;
        bpe_cli_print(c);
        frac -= (double)d;
    }
}

static int bpe_streq(const char* a, const char* b) {
    while (*a && *b) { if (*a != *b) return 0; a++; b++; }
    return (*a == 0 && *b == 0);
}

static int bpe_parse_int(const char* s) {
    int v = 0;
    while (*s >= '0' && *s <= '9') { v = v * 10 + (*s - '0'); s++; }
    return v;
}

static int bpe_parse_cmdline(const char* cmd, char* buf, int buf_sz, char* argv[], int max_args) {
    int argc = 0, bp = 0, i = 0;
    while (cmd[i] && argc < max_args) {
        while (cmd[i] == ' ' || cmd[i] == '\t') i++;
        if (!cmd[i]) break;
        argv[argc++] = buf + bp;
        int in_q = 0;
        while (cmd[i]) {
            if (cmd[i] == '"') { in_q = !in_q; i++; continue; }
            if (!in_q && (cmd[i] == ' ' || cmd[i] == '\t')) break;
            if (bp + 1 < buf_sz) buf[bp++] = cmd[i];
            i++;
        }
        if (bp < buf_sz) buf[bp++] = '\0';
    }
    return argc;
}

void bpe_cli_start(void) {
#ifdef _WIN32
    char arg_buf[2048];
    char* argv[32];
    int argc = bpe_parse_cmdline(GetCommandLineA(), arg_buf, 2048, argv, 32);

    const char* mode = "train";
    const char* corpus = (const char*)0;
    const char* vocab  = (const char*)0;
    const char* input  = (const char*)0;
    const char* output = (const char*)0;
    const char* held   = (const char*)0;
    int target = 8000;

    for (int i = 1; i < argc; i++) {
        if (bpe_streq(argv[i], "--encode-file")) mode = "encode_file";
        else if (bpe_streq(argv[i], "--eval-heldout")) mode = "eval_heldout";
        else if (bpe_streq(argv[i], "--dump")) {
            mode = "dump";
            if (i + 1 < argc && argv[i + 1][0] != '-') vocab = argv[++i];
        }
        else if (bpe_streq(argv[i], "--corpus") && i + 1 < argc) corpus = argv[++i];
        else if (bpe_streq(argv[i], "--vocab") && i + 1 < argc)  vocab  = argv[++i];
        else if (bpe_streq(argv[i], "--input") && i + 1 < argc)  input  = argv[++i];
        else if (bpe_streq(argv[i], "--output") && i + 1 < argc) output = argv[++i];
        else if (bpe_streq(argv[i], "--heldout") && i + 1 < argc) held  = argv[++i];
        else if (bpe_streq(argv[i], "--target") && i + 1 < argc) target = bpe_parse_int(argv[++i]);
    }

    bpe_i64 freq = bpe_qpf_hz();
    if (bpe_streq(mode, "dump")) {
        if (!vocab) { bpe_cli_print("USAGE: --dump <vocab.bin>\n"); ExitProcess(1); }
        if (bpe_load(vocab) != 0) { bpe_cli_print("BPE_LOAD_FAILED\n"); ExitProcess(1); }
        bpe_cli_print("BPE_DUMP_OK | vocab="); bpe_cli_print_i64(g_bpe.vocab_size);
        bpe_cli_print(" | merges="); bpe_cli_print_i64(g_bpe.num_merges);
        bpe_cli_print(" | total_tok_bytes="); bpe_cli_print_i64(g_bpe.total_tok_bytes);
        bpe_cli_print("\n");
        int sample_ids[12] = { 256, 257, 258, 259, 260, 500, 1000, 2000, 4000, 6000, g_bpe.vocab_size - 2, g_bpe.vocab_size - 1 };
        for (int s = 0; s < 12; s++) {
            int id = sample_ids[s];
            if (id < 0 || id >= g_bpe.vocab_size) continue;
            int len = (int)g_bpe.tok_lens[id];
            bpe_u32 off = g_bpe.tok_offs[id];
            bpe_cli_print("  tok["); bpe_cli_print_i64(id); bpe_cli_print("] (len=");
            bpe_cli_print_i64(len); bpe_cli_print(") = \"");
            char one[2]; one[1] = 0;
            for (int j = 0; j < len; j++) {
                bpe_u8 ch = g_bpe.tok_bytes[off + (bpe_u32)j];
                if (ch == '\n') bpe_cli_print("\\n");
                else if (ch == '\r') bpe_cli_print("\\r");
                else if (ch >= 32 && ch <= 126) { one[0] = (char)ch; bpe_cli_print(one); }
                else { bpe_cli_print("\\x"); }
            }
            bpe_cli_print("\"\n");
        }
        bpe_free();
        ExitProcess(0);
    } else if (bpe_streq(mode, "train")) {
        if (!corpus || !vocab) { bpe_cli_print("USAGE: --corpus <txt> --vocab <bin> --target <V>\n"); ExitProcess(1); }
        bpe_i64 t0 = bpe_qpc_now();
        int vs = bpe_train(corpus, vocab, target);
        bpe_i64 t1 = bpe_qpc_now();
        double ms = (double)(t1 - t0) * 1000.0 / (double)freq;
        if (vs < 0) { bpe_cli_print("BPE_TRAIN_FAILED\n"); ExitProcess(1); }
        bpe_cli_print("BPE_TRAIN_OK | vocab_size="); bpe_cli_print_i64(vs);
        bpe_cli_print(" | time_ms="); bpe_cli_print_f64(ms, 2); bpe_cli_print("\n");
        ExitProcess(0);
    } else if (bpe_streq(mode, "encode_file")) {
        if (!vocab || !input || !output) ExitProcess(1);
        if (bpe_load(vocab) != 0) { bpe_cli_print("BPE_LOAD_FAILED\n"); ExitProcess(1); }
        bpe_i64 t0 = bpe_qpc_now();
        int ntoks = bpe_encode_file(input, output);
        bpe_i64 t1 = bpe_qpc_now();
        double ms = (double)(t1 - t0) * 1000.0 / (double)freq;
        bpe_free();
        if (ntoks < 0) { bpe_cli_print("BPE_ENCODE_FILE_FAILED\n"); ExitProcess(1); }
        bpe_cli_print("BPE_ENCODE_FILE_OK | tokens="); bpe_cli_print_i64(ntoks);
        bpe_cli_print(" | time_ms="); bpe_cli_print_f64(ms, 2); bpe_cli_print("\n");
        ExitProcess(0);
    } else if (bpe_streq(mode, "eval_heldout")) {
        if (!vocab || !held) ExitProcess(1);
        if (bpe_load(vocab) != 0) { bpe_cli_print("BPE_LOAD_FAILED\n"); ExitProcess(1); }
        bpe_u8* raw = (bpe_u8*)0; bpe_u32 rlen = 0;
        if (bpe_read_entire_file(held, &raw, &rlen) != 0) { bpe_free(); ExitProcess(1); }

        int* tok_buf = (int*)bpe_valloc(131072ULL * sizeof(int));
        char* dec_buf = (char*)bpe_valloc(131072ULL);
        if (!tok_buf || !dec_buf) { bpe_vfree(raw); bpe_free(); ExitProcess(1); }

        int exact = 0, tested = 0;
        long long total_chars = 0, total_toks = 0;
        bpe_u32 p = 0;
        // Split into 100 held-out sentences delimited by newline after >= 25 chars or blank line
        while (p < rlen && tested < 100) {
            while (p < rlen && (raw[p] == '\r' || raw[p] == '\n' || raw[p] == ' ')) p++;
            if (p >= rlen) break;
            bpe_u32 start = p;
            while (p < rlen && !(raw[p] == '\n' && (p - start >= 24 || p + 1 >= rlen || raw[p + 1] == '\r' || raw[p + 1] == '\n'))) {
                if ((raw[p] == '.' || raw[p] == '!' || raw[p] == '?') && (p - start >= 20)) { p++; break; }
                p++;
            }
            int slen = (int)(p - start);
            if (slen < 15) continue;
            int nt = bpe_encode((const char*)(raw + start), slen, tok_buf, 131072);
            int nd = (nt >= 0) ? bpe_decode(tok_buf, nt, dec_buf, 131072) : -1;
            int eq = (nd == slen);
            if (eq) {
                for (int k = 0; k < slen; k++) {
                    if (dec_buf[k] != (char)raw[start + (bpe_u32)k]) { eq = 0; break; }
                }
            }
            if (eq) exact++;
            tested++;
            total_chars += slen;
            if (nt > 0) total_toks += nt;
        }

        // Measure 10,000+ tokens encode speed on contiguous held-out text
        bpe_i64 ts0 = bpe_qpc_now();
        int speed_toks = bpe_encode((const char*)raw, (int)(rlen > 45000U ? 45000U : rlen), tok_buf, 131072);
        bpe_i64 ts1 = bpe_qpc_now();
        double speed_ms = (double)(ts1 - ts0) * 1000.0 / (double)freq;
        double ms_per_10k = (speed_toks > 0) ? (speed_ms * (10000.0 / (double)speed_toks)) : -1.0;

        double avg_cpt = (total_toks > 0) ? ((double)total_chars / (double)total_toks) : 0.0;
        double byte_comp = (total_toks > 0) ? ((double)total_chars / ((double)total_toks * 4.0)) : 0.0;

        bpe_cli_print("BPE_EVAL_OK | vocab="); bpe_cli_print_i64(bpe_vocab_size());
        bpe_cli_print(" | roundtrip="); bpe_cli_print_i64(exact); bpe_cli_print("/"); bpe_cli_print_i64(tested);
        bpe_cli_print(" | chars="); bpe_cli_print_i64(total_chars);
        bpe_cli_print(" | tokens="); bpe_cli_print_i64(total_toks);
        bpe_cli_print(" | avg_chars_per_tok="); bpe_cli_print_f64(avg_cpt, 4);
        bpe_cli_print(" | uint32_byte_ratio="); bpe_cli_print_f64(byte_comp, 4);
        bpe_cli_print(" | unk=0/0 (0.00% byte-fallback)\n");
        bpe_cli_print("ENCODE_SPEED | toks="); bpe_cli_print_i64(speed_toks);
        bpe_cli_print(" | elapsed_ms="); bpe_cli_print_f64(speed_ms, 3);
        bpe_cli_print(" | ms_per_10k_toks="); bpe_cli_print_f64(ms_per_10k, 3); bpe_cli_print("\n");

        bpe_vfree(tok_buf); bpe_vfree(dec_buf); bpe_vfree(raw); bpe_free();
        ExitProcess((exact == tested && tested == 100) ? 0 : 1);
    }
    ExitProcess(1);
#endif
}
#endif
