#ifndef QADAV_CONSOLIDATE_H
#define QADAV_CONSOLIDATE_H

#include "qadav_qemp.h"
#include "qadav_agi.h"

int qadav_consolidate_memory(qadav_memory_provenance* mp,
                             qadav_graph* graph,
                             double threshold,
                             void* auth_boundary);

#endif

