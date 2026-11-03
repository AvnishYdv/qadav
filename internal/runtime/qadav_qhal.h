#ifndef QADAV_QHAL_H
#define QADAV_QHAL_H

#ifdef __cplusplus
extern "C" {
#endif

#define QADAV_QHAL_MAX_QUBITS 20
#define QADAV_QHAL_MAX_OPS 256

typedef enum {
    QHAL_BACKEND_LOCAL_SIM = 0,
    QHAL_BACKEND_IBMQ = 1,
    QHAL_BACKEND_IONQ = 2,
    QHAL_BACKEND_RIGETTI = 3
} qhal_backend_kind;

typedef enum {
    QHAL_OP_H = 0,
    QHAL_OP_X = 1,
    QHAL_OP_Y = 2,
    QHAL_OP_Z = 3,
    QHAL_OP_CNOT = 4,
    QHAL_OP_MEASURE = 5
} qhal_op_kind;

typedef struct qhal_op {
    unsigned char kind;
    unsigned int  q1;
    unsigned int  q2;  // for CNOT
} qhal_op;

typedef struct qhal_circuit {
    int          n_qubits;
    int          n_ops;
    qhal_op      ops[QADAV_QHAL_MAX_OPS];
    unsigned char circuit_hash[32];
    int          backend;
} qhal_circuit;

qhal_circuit* qhal_create(int n_qubits, int backend);
int  qhal_add_op(qhal_circuit* c, unsigned char kind, unsigned int q1, unsigned int q2);
int  qhal_export_qasm3(const qhal_circuit* c, char* out_buf, unsigned long long buf_len, unsigned long long* out_written);
int  qhal_hash_circuit(qhal_circuit* c);
void qhal_shutdown(qhal_circuit* c);

#ifdef __cplusplus
}
#endif

#endif /* QADAV_QHAL_H */

