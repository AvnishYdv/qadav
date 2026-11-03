#include "qadav_transpile.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static unsigned long long tr_append_str(char* buf, unsigned long long buf_len, unsigned long long pos, const char* s) {
    if (!s) return pos;
    while (*s) {
        if (pos + 1 < buf_len && buf) {
            buf[pos] = *s;
        }
        pos++;
        s++;
    }
    return pos;
}

static unsigned long long tr_append_uint(char* buf, unsigned long long buf_len, unsigned long long pos, unsigned int val) {
    char tmp[16];
    int len = 0;
    if (val == 0) {
        tmp[len++] = '0';
    } else {
        unsigned int v = val;
        while (v > 0) {
            tmp[len++] = (char)('0' + (v % 10));
            v /= 10;
        }
    }
    for (int i = len - 1; i >= 0; i--) {
        if (pos + 1 < buf_len && buf) {
            buf[pos] = tmp[i];
        }
        pos++;
    }
    return pos;
}

int transpile_supported_targets(char* out_buf, unsigned long long buf_len) {
    const char* list = "qiskit,cirq,ionq,qasm3";
    unsigned long long pos = tr_append_str(out_buf, buf_len, 0, list);
    if (out_buf && buf_len > 0) {
        unsigned long long end = (pos < buf_len) ? pos : (buf_len - 1);
        out_buf[end] = '\0';
    }
    return 0;
}

int transpile_circuit(
    const qhal_circuit* in,
    transpile_target target,
    const transpile_opts* opts,
    char* out_buf,
    unsigned long long buf_len,
    unsigned long long* out_written,
    int* out_op_count
) {
    if (!in || !out_buf || buf_len == 0) {
        if (out_written) *out_written = 0;
        if (out_op_count) *out_op_count = 0;
        return -1;
    }

    if (target < TRANSPILE_TARGET_QISKIT || target > TRANSPILE_TARGET_QASM3) {
        if (out_written) *out_written = 0;
        if (out_op_count) *out_op_count = 0;
        return -1;
    }

    // Copy ops to local working buffer
    qhal_op local_ops[QADAV_QHAL_MAX_OPS];
    int n_ops = in->n_ops;
    if (n_ops > QADAV_QHAL_MAX_OPS) n_ops = QADAV_QHAL_MAX_OPS;

    for (int i = 0; i < n_ops; i++) {
        local_ops[i] = in->ops[i];
    }

    // Optimization: cancel adjacent identical self-inverse gates
    if (opts && opts->optimize_adjacent) {
        int changed = 1;
        while (changed) {
            changed = 0;
            for (int i = 0; i < n_ops - 1; i++) {
                if (local_ops[i].kind == local_ops[i + 1].kind) {
                    int can_cancel = 0;
                    if (local_ops[i].kind == QHAL_OP_H ||
                        local_ops[i].kind == QHAL_OP_X ||
                        local_ops[i].kind == QHAL_OP_Y ||
                        local_ops[i].kind == QHAL_OP_Z) {
                        if (local_ops[i].q1 == local_ops[i + 1].q1) {
                            can_cancel = 1;
                        }
                    } else if (local_ops[i].kind == QHAL_OP_CNOT) {
                        if (local_ops[i].q1 == local_ops[i + 1].q1 &&
                            local_ops[i].q2 == local_ops[i + 1].q2) {
                            can_cancel = 1;
                        }
                    }
                    if (can_cancel) {
                        for (int j = i; j < n_ops - 2; j++) {
                            local_ops[j] = local_ops[j + 2];
                        }
                        n_ops -= 2;
                        changed = 1;
                        break;
                    }
                }
            }
        }
    }

    // Validate supported gates
    for (int i = 0; i < n_ops; i++) {
        if (local_ops[i].kind > QHAL_OP_MEASURE) {
            if (out_written) *out_written = 0;
            if (out_op_count) *out_op_count = 0;
            return -2; // Unsupported gate
        }
    }

    unsigned long long pos = 0;
    unsigned int n_qubits = (unsigned int)in->n_qubits;
    if (n_qubits == 0) n_qubits = 1;

    if (target == TRANSPILE_TARGET_QISKIT) {
        pos = tr_append_str(out_buf, buf_len, pos, "from qiskit import QuantumCircuit\nqc = QuantumCircuit(");
        pos = tr_append_uint(out_buf, buf_len, pos, n_qubits);
        pos = tr_append_str(out_buf, buf_len, pos, ", ");
        pos = tr_append_uint(out_buf, buf_len, pos, n_qubits);
        pos = tr_append_str(out_buf, buf_len, pos, ")\n");

        for (int i = 0; i < n_ops; i++) {
            switch (local_ops[i].kind) {
                case QHAL_OP_H:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.h(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
                case QHAL_OP_X:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.x(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
                case QHAL_OP_Y:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.y(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
                case QHAL_OP_Z:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.z(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
                case QHAL_OP_CNOT:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.cx(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ", ");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q2);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
                case QHAL_OP_MEASURE:
                    pos = tr_append_str(out_buf, buf_len, pos, "qc.measure(");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ", ");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ")\n");
                    break;
            }
        }
    } else if (target == TRANSPILE_TARGET_CIRQ) {
        pos = tr_append_str(out_buf, buf_len, pos, "import cirq\nq = cirq.LineQubit.range(");
        pos = tr_append_uint(out_buf, buf_len, pos, n_qubits);
        pos = tr_append_str(out_buf, buf_len, pos, ")\ncircuit = cirq.Circuit([");

        for (int i = 0; i < n_ops; i++) {
            if (i > 0) {
                pos = tr_append_str(out_buf, buf_len, pos, ", ");
            }
            switch (local_ops[i].kind) {
                case QHAL_OP_H:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.H(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
                case QHAL_OP_X:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.X(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
                case QHAL_OP_Y:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.Y(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
                case QHAL_OP_Z:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.Z(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
                case QHAL_OP_CNOT:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.CNOT(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "], q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q2);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
                case QHAL_OP_MEASURE:
                    pos = tr_append_str(out_buf, buf_len, pos, "cirq.measure(q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "])");
                    break;
            }
        }
        pos = tr_append_str(out_buf, buf_len, pos, "])\n");
    } else if (target == TRANSPILE_TARGET_IONQ) {
        pos = tr_append_str(out_buf, buf_len, pos, "{\"gateset\":\"qis\",\"qubits\":");
        pos = tr_append_uint(out_buf, buf_len, pos, n_qubits);
        pos = tr_append_str(out_buf, buf_len, pos, ",\"circuit\":[");

        for (int i = 0; i < n_ops; i++) {
            if (i > 0) {
                pos = tr_append_str(out_buf, buf_len, pos, ",");
            }
            switch (local_ops[i].kind) {
                case QHAL_OP_H:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"h\",\"targets\":[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "]}");
                    break;
                case QHAL_OP_X:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"x\",\"targets\":[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "]}");
                    break;
                case QHAL_OP_Y:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"y\",\"targets\":[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "]}");
                    break;
                case QHAL_OP_Z:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"z\",\"targets\":[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "]}");
                    break;
                case QHAL_OP_CNOT:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"cnot\",\"control\":");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, ",\"target\":");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q2);
                    pos = tr_append_str(out_buf, buf_len, pos, "}");
                    break;
                case QHAL_OP_MEASURE:
                    pos = tr_append_str(out_buf, buf_len, pos, "{\"gate\":\"measure\",\"targets\":[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "]}");
                    break;
            }
        }
        pos = tr_append_str(out_buf, buf_len, pos, "]}\n");
    } else if (target == TRANSPILE_TARGET_QASM3) {
        pos = tr_append_str(out_buf, buf_len, pos, "OPENQASM 3.0;\ninclude \"stdgates.inc\";\nqubit[");
        pos = tr_append_uint(out_buf, buf_len, pos, n_qubits);
        pos = tr_append_str(out_buf, buf_len, pos, "] q;\n");

        for (int i = 0; i < n_ops; i++) {
            switch (local_ops[i].kind) {
                case QHAL_OP_H:
                    pos = tr_append_str(out_buf, buf_len, pos, "h q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
                case QHAL_OP_X:
                    pos = tr_append_str(out_buf, buf_len, pos, "x q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
                case QHAL_OP_Y:
                    pos = tr_append_str(out_buf, buf_len, pos, "y q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
                case QHAL_OP_Z:
                    pos = tr_append_str(out_buf, buf_len, pos, "z q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
                case QHAL_OP_CNOT:
                    pos = tr_append_str(out_buf, buf_len, pos, "cnot q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "], q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q2);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
                case QHAL_OP_MEASURE:
                    pos = tr_append_str(out_buf, buf_len, pos, "measure q[");
                    pos = tr_append_uint(out_buf, buf_len, pos, local_ops[i].q1);
                    pos = tr_append_str(out_buf, buf_len, pos, "];\n");
                    break;
            }
        }
    }

    if (pos + 1 > buf_len) {
        if (out_written) *out_written = pos + 1;
        if (out_op_count) *out_op_count = n_ops;
        return -3; // Buffer too small
    }

    out_buf[pos] = '\0';
    if (out_written) *out_written = pos;
    if (out_op_count) *out_op_count = n_ops;
    return 0;
}
