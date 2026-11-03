#ifndef QADAV_AGI_H
#define QADAV_AGI_H

#include "qadav_tensor.h"

/* Episodic Memory */
typedef struct qadav_mem_entry {
    char                    key[64];
    long long               val;
    struct qadav_mem_entry* next;
} qadav_mem_entry;

typedef struct qadav_memory {
    long long         capacity;
    long long         count;
    long long         num_buckets;
    qadav_mem_entry** buckets;
} qadav_memory;

qadav_memory* qadav_memory_create(long long capacity);
void          qadav_memory_free(qadav_memory* m);
void          qadav_memory_remember(qadav_memory* m, const char* key, long long val);
long long     qadav_memory_recall(qadav_memory* m, const char* key);
void          qadav_memory_print(qadav_memory* m);

/* Semantic Association Graph */
typedef struct qadav_graph_edge {
    int                      to_idx;
    double                   weight;
    struct qadav_graph_edge* next;
} qadav_graph_edge;

typedef struct qadav_graph_node_entry {
    char              name[64];
    long long         data;
    qadav_graph_edge* edges;
} qadav_graph_node_entry;

typedef struct qadav_graph {
    int                     num_nodes;
    int                     capacity;
    qadav_graph_node_entry* nodes;
} qadav_graph;

qadav_graph*  qadav_graph_create(void);
void          qadav_graph_free(qadav_graph* g);
void          qadav_graph_node(qadav_graph* g, const char* name, long long data);
void          qadav_graph_link(qadav_graph* g, const char* from, const char* to, double weight);
qadav_tensor* qadav_graph_trace(qadav_graph* g, const char* start, long long depth);
void          qadav_graph_print(qadav_graph* g);

/* Executable Procedure */
typedef struct qadav_procedure {
    void* fn_ptr;
} qadav_procedure;

qadav_procedure* qadav_procedure_create(void* fn_ptr);
void             qadav_procedure_free(qadav_procedure* p);

#endif

