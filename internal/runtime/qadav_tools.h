#ifndef QADAV_TOOLS_H
#define QADAV_TOOLS_H

#include "qadav_sha3.h"
#include "qadav_qeeh.h"
#include "qadav_qera.h"

#define QADAV_MAX_TOOLS 32

typedef struct qadav_tool_entry {
    char          name[64];
    unsigned char boundary_hash[32];
    void*         boundary;
    int           has_pre;
    int           has_post;
} qadav_tool_entry;

typedef struct qadav_tool_registry {
    int              count;
    qadav_tool_entry tools[QADAV_MAX_TOOLS];
} qadav_tool_registry;

// Tool management & execution
int qadav_tool_register(const char* name, void* boundary, int has_pre, int has_post);
int qadav_tool_call(const char* name, void* caller_boundary, const char* arg, void* audit_chain);
void qadav_tools_reset(void);

#endif

