#include "qadav_consolidate.h"
#include "qadav_prime.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void* qadav_alloc(unsigned long long sz);
void  qadav_free(void* p);
int   qadav_sha3_ct_compare(const unsigned char* a, const unsigned char* b, int len);

int qadav_consolidate_memory(qadav_memory_provenance* mp,
                             qadav_graph* graph,
                             double threshold,
                             void* auth_boundary) {
    if (!mp || !graph) return -1;
    if (threshold < 0.0 || threshold > 1.0) return -2;

    // Verify boundary if provided
    if (auth_boundary) {
        unsigned char cur_hash[32];
        qadav_qeeh_boundary_hash(auth_boundary, cur_hash);
        if (qadav_sha3_ct_compare(cur_hash, mp->auth_boundary_id, 32) != 0) {
            qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
            ExitProcess((unsigned int)0xDEADBEEF);
#endif
            return -3;
        }
    }

    int consolidated = 0;
    char prev_key[64] = {0};

    for (int i = 0; i < mp->count; i++) {
        if (mp->entries[i].is_locked != 0) continue;

        unsigned char* val = 0;
        unsigned long long len = 0;
        int r = qadav_qemp_recall(mp, mp->entries[i].key, auth_boundary, &val, &len);
        if (r == 0 && val && len > 0) {
            // Compute deterministic prime-symbolic coordinate from value bytes
            unsigned long long prime_data = 101;
            for (unsigned long long j = 0; j < len; j++) {
                prime_data = (prime_data * 31 + val[j]) % 7919;
            }
            if (prime_data == 0) prime_data = 2;

            qadav_graph_node(graph, mp->entries[i].key, (long long)prime_data);

            if (consolidated > 0 && prev_key[0] != '\0') {
                qadav_graph_link(graph, prev_key, mp->entries[i].key, 1.0);
            }

            int k = 0;
            while (mp->entries[i].key[k] && k < 63) {
                prev_key[k] = mp->entries[i].key[k];
                k++;
            }
            prev_key[k] = '\0';

            consolidated++;
            qadav_free(val);
        }
    }

    return consolidated;
}
