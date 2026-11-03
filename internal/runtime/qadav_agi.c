// Qadav AGI Native Types Runtime (Episodic Memory, Semantic Graph, Procedure) — Zero-Libc
#include "qadav_agi.h"

static int agi_strlen(const char* s) {
    int len = 0;
    while (s && s[len]) len++;
    return len;
}

static int agi_strcmp(const char* a, const char* b) {
    if (!a && !b) return 0;
    if (!a) return -1;
    if (!b) return 1;
    while (*a && (*a == *b)) {
        a++;
        b++;
    }
    return *(const unsigned char*)a - *(const unsigned char*)b;
}

static void agi_strcpy(char* dst, const char* src, int max_len) {
    int i = 0;
    while (src && src[i] && i < max_len - 1) {
        dst[i] = src[i];
        i++;
    }
    dst[i] = '\0';
}

static unsigned long long agi_fnv1a(const char* s) {
    unsigned long long hash = 14695981039346656037ULL;
    while (s && *s) {
        hash ^= (unsigned char)(*s);
        hash *= 1099511628211ULL;
        s++;
    }
    return hash;
}

static int agi_int_to_str(long long v, char* buf) {
    int i = 0;
    if (v == 0) {
        buf[i++] = '0';
        return i;
    }
    int is_neg = 0;
    unsigned long long uv;
    if (v < 0) {
        is_neg = 1;
        uv = (unsigned long long)(-v);
    } else {
        uv = (unsigned long long)v;
    }
    char temp[32];
    int t = 0;
    while (uv > 0) {
        temp[t++] = '0' + (char)(uv % 10);
        uv /= 10;
    }
    if (is_neg) {
        buf[i++] = '-';
    }
    while (t > 0) {
        buf[i++] = temp[--t];
    }
    return i;
}

/* ========================================================================= */
/* Episodic Memory Implementation                                            */
/* ========================================================================= */

qadav_memory* qadav_memory_create(long long capacity) {
    qadav_memory* m = (qadav_memory*)qadav_alloc(sizeof(qadav_memory));
    if (!m) return 0;
    if (capacity <= 0) capacity = 64;
    m->capacity = capacity;
    m->count = 0;
    m->num_buckets = 64;
    m->buckets = (qadav_mem_entry**)qadav_alloc(sizeof(qadav_mem_entry*) * m->num_buckets);
    if (!m->buckets) {
        qadav_free(m);
        return 0;
    }
    for (int i = 0; i < m->num_buckets; i++) {
        m->buckets[i] = 0;
    }
    return m;
}

void qadav_memory_free(qadav_memory* m) {
    if (!m) return;
    if (m->buckets) {
        for (int i = 0; i < m->num_buckets; i++) {
            qadav_mem_entry* cur = m->buckets[i];
            while (cur) {
                qadav_mem_entry* next = cur->next;
                qadav_free(cur);
                cur = next;
            }
        }
        qadav_free(m->buckets);
    }
    qadav_free(m);
}

void qadav_memory_remember(qadav_memory* m, const char* key, long long val) {
    if (!m || !m->buckets || !key) return;
    unsigned long long h = agi_fnv1a(key);
    int bidx = (int)(h % (unsigned long long)m->num_buckets);

    qadav_mem_entry* cur = m->buckets[bidx];
    while (cur) {
        if (agi_strcmp(cur->key, key) == 0) {
            cur->val = val;
            return;
        }
        cur = cur->next;
    }

    qadav_mem_entry* entry = (qadav_mem_entry*)qadav_alloc(sizeof(qadav_mem_entry));
    if (!entry) return;
    agi_strcpy(entry->key, key, 64);
    entry->val = val;
    entry->next = m->buckets[bidx];
    m->buckets[bidx] = entry;
    m->count++;
}

long long qadav_memory_recall(qadav_memory* m, const char* key) {
    if (!m || !m->buckets || !key) return 0;
    unsigned long long h = agi_fnv1a(key);
    int bidx = (int)(h % (unsigned long long)m->num_buckets);

    qadav_mem_entry* cur = m->buckets[bidx];
    while (cur) {
        if (agi_strcmp(cur->key, key) == 0) {
            return cur->val;
        }
        cur = cur->next;
    }
    return 0;
}

void qadav_memory_print(qadav_memory* m) {
    if (!m) {
        const char* s = "memory⟨null⟩\n";
        qadav_write_raw(s, agi_strlen(s));
        return;
    }
    const char* prefix = "memory\xe2\x9f\xa8" "count=";
    qadav_write_raw(prefix, agi_strlen(prefix));

    char buf[32];
    int n = agi_int_to_str(m->count, buf);
    qadav_write_raw(buf, n);

    const char* capPfx = ", capacity=";
    qadav_write_raw(capPfx, agi_strlen(capPfx));

    n = agi_int_to_str(m->capacity, buf);
    qadav_write_raw(buf, n);

    const char* suffix = "\xe2\x9f\xa9\n";
    qadav_write_raw(suffix, agi_strlen(suffix));
}

/* ========================================================================= */
/* Semantic Association Graph Implementation                                 */
/* ========================================================================= */

static int graph_find_node(qadav_graph* g, const char* name) {
    if (!g || !name) return -1;
    for (int i = 0; i < g->num_nodes; i++) {
        if (agi_strcmp(g->nodes[i].name, name) == 0) {
            return i;
        }
    }
    return -1;
}

qadav_graph* qadav_graph_create(void) {
    qadav_graph* g = (qadav_graph*)qadav_alloc(sizeof(qadav_graph));
    if (!g) return 0;
    g->capacity = 64;
    g->num_nodes = 0;
    g->nodes = (qadav_graph_node_entry*)qadav_alloc(sizeof(qadav_graph_node_entry) * g->capacity);
    if (!g->nodes) {
        qadav_free(g);
        return 0;
    }
    for (int i = 0; i < g->capacity; i++) {
        g->nodes[i].name[0] = '\0';
        g->nodes[i].data = 0;
        g->nodes[i].edges = 0;
    }
    return g;
}

void qadav_graph_free(qadav_graph* g) {
    if (!g) return;
    if (g->nodes) {
        for (int i = 0; i < g->num_nodes; i++) {
            qadav_graph_edge* cur = g->nodes[i].edges;
            while (cur) {
                qadav_graph_edge* next = cur->next;
                qadav_free(cur);
                cur = next;
            }
        }
        qadav_free(g->nodes);
    }
    qadav_free(g);
}

void qadav_graph_node(qadav_graph* g, const char* name, long long data) {
    if (!g || !g->nodes || !name) return;
    int idx = graph_find_node(g, name);
    if (idx >= 0) {
        g->nodes[idx].data = data;
        return;
    }
    if (g->num_nodes >= g->capacity) {
        return; // Fixed capacity for safety
    }
    idx = g->num_nodes++;
    agi_strcpy(g->nodes[idx].name, name, 64);
    g->nodes[idx].data = data;
    g->nodes[idx].edges = 0;
}

void qadav_graph_link(qadav_graph* g, const char* from, const char* to, double weight) {
    if (!g || !from || !to) return;
    int f_idx = graph_find_node(g, from);
    if (f_idx < 0) {
        qadav_graph_node(g, from, 0);
        f_idx = graph_find_node(g, from);
    }
    int t_idx = graph_find_node(g, to);
    if (t_idx < 0) {
        qadav_graph_node(g, to, 0);
        t_idx = graph_find_node(g, to);
    }
    if (f_idx < 0 || t_idx < 0) return;

    qadav_graph_edge* edge = (qadav_graph_edge*)qadav_alloc(sizeof(qadav_graph_edge));
    if (!edge) return;
    edge->to_idx = t_idx;
    edge->weight = weight;
    edge->next = g->nodes[f_idx].edges;
    g->nodes[f_idx].edges = edge;
}

qadav_tensor* qadav_graph_trace(qadav_graph* g, const char* start, long long depth) {
    if (!g || !start) {
        int dim = 0;
        return qadav_tensor_create(QADAV_ELEM_INT64, 1, &dim);
    }
    int s_idx = graph_find_node(g, start);
    if (s_idx < 0) {
        int dim = 0;
        return qadav_tensor_create(QADAV_ELEM_INT64, 1, &dim);
    }

    /* BFS queue and visited tracking */
    int visited[64];
    for (int i = 0; i < 64; i++) visited[i] = 0;

    int queue[64];
    int q_depth[64];
    int head = 0;
    int tail = 0;

    queue[tail] = s_idx;
    q_depth[tail] = 0;
    tail++;
    visited[s_idx] = 1;

    int out_nodes[64];
    int out_count = 0;

    while (head < tail) {
        int curr = queue[head];
        int d = q_depth[head];
        head++;

        out_nodes[out_count++] = curr;

        if (d < depth) {
            qadav_graph_edge* e = g->nodes[curr].edges;
            while (e) {
                int to = e->to_idx;
                if (!visited[to] && tail < 64) {
                    visited[to] = 1;
                    queue[tail] = to;
                    q_depth[tail] = d + 1;
                    tail++;
                }
                e = e->next;
            }
        }
    }

    qadav_tensor* t = qadav_tensor_create(QADAV_ELEM_INT64, 1, &out_count);
    for (int i = 0; i < out_count; i++) {
        long long val = g->nodes[out_nodes[i]].data;
        if (val == 0) {
            val = (long long)out_nodes[i];
        }
        qadav_tensor_set_i64(t, i, val);
    }
    return t;
}

void qadav_graph_print(qadav_graph* g) {
    if (!g) {
        const char* s = "graph⟨null⟩\n";
        qadav_write_raw(s, agi_strlen(s));
        return;
    }
    const char* prefix = "graph\xe2\x9f\xa8" "nodes=";
    qadav_write_raw(prefix, agi_strlen(prefix));

    char buf[32];
    int n = agi_int_to_str(g->num_nodes, buf);
    qadav_write_raw(buf, n);

    const char* suffix = "\xe2\x9f\xa9\n";
    qadav_write_raw(suffix, agi_strlen(suffix));
}

/* ========================================================================= */
/* Procedure Implementation                                                  */
/* ========================================================================= */

qadav_procedure* qadav_procedure_create(void* fn_ptr) {
    qadav_procedure* p = (qadav_procedure*)qadav_alloc(sizeof(qadav_procedure));
    if (!p) return 0;
    p->fn_ptr = fn_ptr;
    return p;
}

void qadav_procedure_free(qadav_procedure* p) {
    if (p) {
        qadav_free(p);
    }
}
