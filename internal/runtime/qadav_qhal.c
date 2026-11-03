#include "qadav_qhal.h"
#include "qadav_sha3.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static void qhal_memset(void* dst, int val, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = (unsigned char)val;
    }
}

static void qhal_memcpy(void* dst, const void* src, unsigned long long n) {
    unsigned char* d = (unsigned char*)dst;
    const unsigned char* s = (const unsigned char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
}

static unsigned long long qhal_append_str(char* buf, unsigned long long buf_len, unsigned long long pos, const char* s) {
    if (!s) return pos;
    while (*s) {
        if (pos + 1 < buf_len) {
            buf[pos] = *s;
        }
        pos++;
        s++;
    }
    return pos;
}

static unsigned long long qhal_append_uint(char* buf, unsigned long long buf_len, unsigned long long pos, unsigned int v) {
    char digits[16];
    int len = 0;
    if (v == 0) {
        digits[len++] = '0';
    } else {
        unsigned int temp = v;
        while (temp > 0) {
            digits[len++] = (char)('0' + (temp % 10));
            temp /= 10;
        }
    }
    for (int i = len - 1; i >= 0; i--) {
        if (pos + 1 < buf_len) {
            buf[pos] = digits[i];
        }
        pos++;
    }
    return pos;
}

qhal_circuit* qhal_create(int n_qubits, int backend) {
    if (n_qubits < 1) n_qubits = 1;
    if (n_qubits > QADAV_QHAL_MAX_QUBITS) n_qubits = QADAV_QHAL_MAX_QUBITS;

    qhal_circuit* c = (qhal_circuit*)qadav_alloc(sizeof(qhal_circuit));
    if (!c) return 0;

    qhal_memset(c, 0, sizeof(qhal_circuit));
    c->n_qubits = n_qubits;
    c->n_ops = 0;
    c->backend = backend;
    return c;
}

int qhal_add_op(qhal_circuit* c, unsigned char kind, unsigned int q1, unsigned int q2) {
    if (!c) return -1;
    if (c->n_ops >= QADAV_QHAL_MAX_OPS) return -2;
    if (q1 >= (unsigned int)c->n_qubits) return -3;
    if (kind == QHAL_OP_CNOT && q2 >= (unsigned int)c->n_qubits) return -4;

    int idx = c->n_ops;
    c->ops[idx].kind = kind;
    c->ops[idx].q1 = q1;
    c->ops[idx].q2 = q2;
    c->n_ops++;
    return 0;
}

int qhal_hash_circuit(qhal_circuit* c) {
    if (!c) return -1;
    // Deterministic 9-byte packing per op (no struct padding):
    // Header (12 bytes): n_qubits (4) || n_ops (4) || backend (4)
    // Each op (9 bytes): kind (1) || q1 (4) || q2 (4)
    unsigned char preimage[12 + QADAV_QHAL_MAX_OPS * 9];
    qhal_memset(preimage, 0, sizeof(preimage));

    unsigned int* meta = (unsigned int*)preimage;
    meta[0] = (unsigned int)c->n_qubits;
    meta[1] = (unsigned int)c->n_ops;
    meta[2] = (unsigned int)c->backend;

    unsigned char* p = preimage + 12;
    for (int i = 0; i < c->n_ops; i++) {
        p[i * 9 + 0] = c->ops[i].kind;
        qhal_memcpy(p + i * 9 + 1, &c->ops[i].q1, 4);
        qhal_memcpy(p + i * 9 + 5, &c->ops[i].q2, 4);
    }

    unsigned long long total_len = 12 + (unsigned long long)c->n_ops * 9;
    qadav_sha3_256(preimage, total_len, (unsigned char*)c->circuit_hash);
    return 0;
}

int qhal_export_qasm3(const qhal_circuit* c, char* out_buf, unsigned long long buf_len, unsigned long long* out_written) {
    if (!c || !out_buf || buf_len == 0) {
        if (out_written) *out_written = 0;
        return -1;
    }

    unsigned long long pos = 0;

    pos = qhal_append_str(out_buf, buf_len, pos, "OPENQASM 3.0;\n");
    pos = qhal_append_str(out_buf, buf_len, pos, "include \"stdgates.inc\";\n");
    pos = qhal_append_str(out_buf, buf_len, pos, "qubit[");
    pos = qhal_append_uint(out_buf, buf_len, pos, (unsigned int)c->n_qubits);
    pos = qhal_append_str(out_buf, buf_len, pos, "] q;\n");

    for (int i = 0; i < c->n_ops; i++) {
        switch (c->ops[i].kind) {
            case QHAL_OP_H:
                pos = qhal_append_str(out_buf, buf_len, pos, "h q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            case QHAL_OP_X:
                pos = qhal_append_str(out_buf, buf_len, pos, "x q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            case QHAL_OP_Y:
                pos = qhal_append_str(out_buf, buf_len, pos, "y q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            case QHAL_OP_Z:
                pos = qhal_append_str(out_buf, buf_len, pos, "z q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            case QHAL_OP_CNOT:
                pos = qhal_append_str(out_buf, buf_len, pos, "cnot q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "], q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q2);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            case QHAL_OP_MEASURE:
                pos = qhal_append_str(out_buf, buf_len, pos, "c[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "] = measure q[");
                pos = qhal_append_uint(out_buf, buf_len, pos, c->ops[i].q1);
                pos = qhal_append_str(out_buf, buf_len, pos, "];\n");
                break;
            default:
                break;
        }
    }

    if (pos < buf_len) {
        out_buf[pos] = '\0';
    } else {
        out_buf[buf_len - 1] = '\0';
    }

    if (out_written) {
        *out_written = (pos < buf_len) ? pos : (buf_len - 1);
    }
    return 0;
}

void qhal_shutdown(qhal_circuit* c) {
    if (c) {
        qadav_free(c);
    }
}
