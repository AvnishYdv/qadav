#include "qadav_qeeh.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

static unsigned long long qadav_strlen(const char* s) {
    if (!s) return 0;
    const volatile char* p = (const volatile char*)s;
    unsigned long long len = 0;
    while (p[len] != '\0') len++;
    return len;
}

static int qadav_bytes_equal(const unsigned char* a, const unsigned char* b, int len) {
    if (!a || !b) return 0;
    int diff = 0;
    for (int i = 0; i < len; i++) {
        diff |= (a[i] ^ b[i]);
    }
    return diff == 0;
}

qadav_ethical_hamiltonian* qadav_qeeh_create(int num_rules, const double* penalties) {
    if (num_rules <= 0) num_rules = 1;

    qadav_ethical_hamiltonian* h = (qadav_ethical_hamiltonian*)qadav_alloc(sizeof(qadav_ethical_hamiltonian));
    if (!h) return 0;

    h->num_rules = num_rules;
    h->ground_energy = 0.0;
    h->penalties = (double*)qadav_alloc((unsigned long long)num_rules * sizeof(double));
    h->conditions = (double*)qadav_alloc((unsigned long long)num_rules * sizeof(double));

    for (int i = 0; i < num_rules; i++) {
        h->penalties[i] = (penalties && penalties[i] >= 0.0) ? penalties[i] : 1.0;
        h->conditions[i] = 1.0; // default: compliant
        if (i < QADAV_QEEH_MAX_RULES) {
            h->rule_names[i][0] = '\0';
        }
    }

    return h;
}

void qadav_qeeh_set_rule_name(qadav_ethical_hamiltonian* h, int idx, const char* name) {
    if (!h || idx < 0 || idx >= h->num_rules || idx >= QADAV_QEEH_MAX_RULES || !name) return;
    int k = 0;
    while (name[k] != '\0' && k < QADAV_QEEH_RULE_NAME_MAX - 1) {
        h->rule_names[idx][k] = name[k];
        k++;
    }
    h->rule_names[idx][k] = '\0';
}

void qadav_qeeh_set_rule_code(void* boundary, int idx, unsigned long long code) {
    if (!boundary || idx < 0 || idx >= QADAV_QEEH_MAX_RULES) return;
    qadav_ethical_hamiltonian* h = (qadav_ethical_hamiltonian*)boundary;
    int k = 0;
    while (code > 0 && k < QADAV_QEEH_RULE_NAME_MAX - 1) {
        h->rule_names[idx][k++] = 'a' + (char)(code % 26);
        code /= 26;
    }
    if (k == 0) h->rule_names[idx][k++] = 'x';
    h->rule_names[idx][k] = '\0';
}

int qadav_qeeh_boundary_hash(const void* boundary, unsigned char out[32]) {
    if (!out) return 0;
    if (!boundary) {
        for (int i = 0; i < 32; i++) out[i] = 0;
        return 0;
    }
    const qadav_ethical_hamiltonian* h = (const qadav_ethical_hamiltonian*)boundary;
    unsigned char buf[4096];
    unsigned long long offset = 0;

    // 1. num_rules (4 bytes Little Endian)
    buf[offset++] = (unsigned char)(h->num_rules & 0xFF);
    buf[offset++] = (unsigned char)((h->num_rules >> 8) & 0xFF);
    buf[offset++] = (unsigned char)((h->num_rules >> 16) & 0xFF);
    buf[offset++] = (unsigned char)((h->num_rules >> 24) & 0xFF);

    // 2. For each rule: rule_name (null-terminated) || penalty (8 bytes double LE)
    for (int i = 0; i < h->num_rules; i++) {
        const char* name = (i < QADAV_QEEH_MAX_RULES) ? h->rule_names[i] : "";
        if (!name || name[0] == '\0') {
            const char* def_p = "rule_";
            for (int k = 0; def_p[k]; k++) buf[offset++] = (unsigned char)def_p[k];
            buf[offset++] = (unsigned char)('0' + (i % 10));
            buf[offset++] = '\0';
        } else {
            int k = 0;
            while (name[k] != '\0' && k < QADAV_QEEH_RULE_NAME_MAX - 1) {
                buf[offset++] = (unsigned char)name[k++];
            }
            buf[offset++] = '\0';
        }

        double p = (h->penalties) ? h->penalties[i] : 1.0;
        unsigned char* p_bytes = (unsigned char*)&p;
        for (int k = 0; k < 8; k++) {
            buf[offset++] = p_bytes[k];
        }
    }

    qadav_sha3_256(buf, offset, out);
    return 1;
}

int qadav_qeeh_evaluate(qadav_ethical_hamiltonian* h, const double* conditions) {
    if (!h) return 0;

    double energy = 0.0;
    for (int i = 0; i < h->num_rules; i++) {
        double cond = (conditions != 0) ? conditions[i] : h->conditions[i];
        if (cond < 0.0) cond = 0.0;
        if (cond > 1.0) cond = 1.0;
        // Energy penalty: E = sum(penalty_i * (1.0 - condition_i))
        energy += h->penalties[i] * (1.0 - cond);
    }

    // Violation threshold: any positive energy perturbation above ground state triggers decoherence
    if (energy > h->ground_energy + 1e-9) {
        return 1; // Unethical violation
    }

    return 0; // Ethical
}

int qadav_qeeh_evaluate_action_kind(qadav_ethical_hamiltonian* h, qadav_action_kind kind) {
    if (!h) return 0;
    int is_unethical = (kind != QADAV_ACTION_PERMITTED) ? 1 : 0;
    double* conds = (double*)qadav_alloc((unsigned long long)h->num_rules * sizeof(double));
    if (!conds) return is_unethical;

    for (int i = 0; i < h->num_rules; i++) {
        conds[i] = is_unethical ? 0.0 : 1.0;
    }

    int res = qadav_qeeh_evaluate(h, conds);
    qadav_free(conds);
    return res;
}

int qadav_qeeh_evaluate_action_hash(qadav_ethical_hamiltonian* h, const unsigned char* action_hash) {
    if (!h || !action_hash) return 0;

    // Check against canonical prohibited action hashes (computed via SHA3-256)
    // 1. "harm_user"
    static const unsigned char HASH_HARM_USER[32] = {
        0xa7, 0x05, 0x22, 0xd8, 0xef, 0x26, 0x48, 0x51,
        0x5a, 0x3e, 0x56, 0xb8, 0x93, 0x0a, 0xc0, 0x98,
        0x5d, 0xec, 0x45, 0x22, 0xd0, 0xd3, 0xd9, 0xb8,
        0xc9, 0xee, 0x92, 0xe2, 0x54, 0x82, 0xf6, 0xbe
    };
    // 2. "leak_weights"
    static const unsigned char HASH_LEAK_WEIGHTS[32] = {
        0xc4, 0x8b, 0x9e, 0xd2, 0x77, 0x24, 0xb2, 0x30,
        0x10, 0x57, 0x5b, 0xa3, 0x1d, 0xc4, 0x89, 0x18,
        0x20, 0x4f, 0x76, 0xb5, 0x41, 0x14, 0xd3, 0x77,
        0x26, 0x78, 0xd9, 0x25, 0x47, 0x3c, 0xef, 0x0a
    };

    int is_unethical = 0;
    if (qadav_bytes_equal(action_hash, HASH_HARM_USER, 32) ||
        qadav_bytes_equal(action_hash, HASH_LEAK_WEIGHTS, 32)) {
        is_unethical = 1;
    }

    double* conds = (double*)qadav_alloc((unsigned long long)h->num_rules * sizeof(double));
    if (!conds) return is_unethical;

    for (int i = 0; i < h->num_rules; i++) {
        conds[i] = is_unethical ? 0.0 : 1.0;
    }

    int res = qadav_qeeh_evaluate(h, conds);
    qadav_free(conds);
    return res;
}

int qadav_qeeh_evaluate_action(qadav_ethical_hamiltonian* h, const char* action_str) {
    if (!h) return 0;
    if (!action_str) return 0;

    // Cryptographic SHA3-256 message digest of the action
    unsigned long long len = qadav_strlen(action_str);
    unsigned char action_hash[32];
    qadav_sha3_256((const unsigned char*)action_str, len, action_hash);

    // Also support exact action tag matching
    // "harm_user" or "leak_weights" or "unauthorized"
    if (len == 9 && action_str[0] == 'h' && action_str[1] == 'a' && action_str[2] == 'r' && action_str[3] == 'm') {
        return qadav_qeeh_evaluate_action_kind(h, QADAV_ACTION_HARMFUL);
    }
    if (len >= 4 && action_str[0] == 'l' && action_str[1] == 'e' && action_str[2] == 'a' && action_str[3] == 'k') {
        return qadav_qeeh_evaluate_action_kind(h, QADAV_ACTION_LEAK_WEIGHTS);
    }

    return qadav_qeeh_evaluate_action_hash(h, action_hash);
}

void qadav_qeeh_trigger_decoherence(void* qstate) {
    qadav_qstate* qs = (qadav_qstate*)qstate;
    if (!qs) {
        qs = qadav_get_qstate();
    }
    if (qs && qs->amplitudes) {
        for (long long i = 0; i < qs->n_states; i++) {
            qs->amplitudes[i].real = 0.0;
            qs->amplitudes[i].imag = 0.0;
        }
    }
}

void qadav_qeeh_free(qadav_ethical_hamiltonian* h) {
    if (!h) return;
    if (h->penalties) qadav_free(h->penalties);
    if (h->conditions) qadav_free(h->conditions);
    qadav_free(h);
}
