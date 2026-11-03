// Qadav Confidence & Epistemic Runtime (CIR) — Zero-Libc
#include "qadav_cir.h"

#if defined(_WIN32)
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

static double cir_sqrt(double x) {
    if (x <= 0.0) return 0.0;
    double guess = (x > 1.0) ? x : 1.0;
    for (int i = 0; i < 20; i++) {
        guess = 0.5 * (guess + x / guess);
    }
    return guess;
}

static int cir_int_to_str(long long v, char* buf) {
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

static int cir_float_to_str(double v, char* buf) {
    int idx = 0;
    if (v < 0.0) {
        buf[idx++] = '-';
        v = -v;
    }
    long long ipart = (long long)v;
    double frac = v - (double)ipart;
    long long fpart = (long long)(frac * 1000000.0 + 0.5);
    if (fpart >= 1000000) {
        ipart += 1;
        fpart -= 1000000;
    }
    idx += cir_int_to_str(ipart, buf + idx);
    buf[idx++] = '.';
    char fbuf[16];
    int fn = cir_int_to_str(fpart, fbuf);
    for (int p = 0; p < 6 - fn; p++) {
        buf[idx++] = '0';
    }
    for (int p = 0; p < fn; p++) {
        buf[idx++] = fbuf[p];
    }
    buf[idx] = '\0';
    return idx;
}

static int cir_strlen(const char* s) {
    int len = 0;
    while (s && s[len]) len++;
    return len;
}

qadav_confidence* qadav_confidence_create(double value, double score, int8_t state) {
    qadav_confidence* c = (qadav_confidence*)qadav_alloc(sizeof(qadav_confidence));
    if (!c) return 0;
    c->value = value;
    if (score < 0.0) score = 0.0;
    if (score > 1.0) score = 1.0;
    c->score = score;
    c->epistemic = state;
    for (int i = 0; i < 64; i++) c->provenance[i] = 0;
    return c;
}

void qadav_confidence_free(qadav_confidence* c) {
    if (c) {
        qadav_free(c);
    }
}

void qadav_confidence_print(qadav_confidence* c) {
    if (!c) {
        const char* msg = "conf⟨null⟩\n";
        qadav_write_raw(msg, cir_strlen(msg));
        return;
    }

    const char* prefix = "conf\xe2\x9f\xa8value="; // "conf⟨value="
    qadav_write_raw(prefix, cir_strlen(prefix));

    char valBuf[64];
    int valLen = cir_float_to_str(c->value, valBuf);
    qadav_write_raw(valBuf, valLen);

    const char* mid1 = ", score=";
    qadav_write_raw(mid1, cir_strlen(mid1));

    char scoreBuf[64];
    int scoreLen = cir_float_to_str(c->score, scoreBuf);
    qadav_write_raw(scoreBuf, scoreLen);

    const char* mid2 = ", state=";
    qadav_write_raw(mid2, cir_strlen(mid2));

    const char* stateStr = "known";
    if (c->epistemic == 1) {
        stateStr = "unknown";
    } else if (c->epistemic == 2) {
        stateStr = "explore";
    }
    qadav_write_raw(stateStr, cir_strlen(stateStr));

    const char* suffix = "\xe2\x9f\xa9\n"; // "⟩\n"
    qadav_write_raw(suffix, cir_strlen(suffix));
}

qadav_confidence* qadav_confidence_add(qadav_confidence* a, qadav_confidence* b) {
    if (!a || !b) return 0;
    double val = a->value + b->value;
    int8_t state = 0;
    double score = 0.0;

    if (a->epistemic == 1 || b->epistemic == 1) {
        state = 1; // UNKNOWN propagates
        score = 0.0;
    } else {
        if (a->epistemic == 2 || b->epistemic == 2) {
            state = 2; // EXPLORE
        } else {
            state = 0; // KNOWN
        }
        score = cir_sqrt(a->score * b->score);
    }
    return qadav_confidence_create(val, score, state);
}

qadav_confidence* qadav_confidence_mul(qadav_confidence* a, qadav_confidence* b) {
    if (!a || !b) return 0;
    double val = a->value * b->value;
    int8_t state = 0;
    double score = 0.0;

    if (a->epistemic == 1 || b->epistemic == 1) {
        state = 1; // UNKNOWN propagates
        score = 0.0;
    } else {
        if (a->epistemic == 2 || b->epistemic == 2) {
            state = 2; // EXPLORE
        } else {
            state = 0; // KNOWN
        }
        score = a->score * b->score;
    }
    return qadav_confidence_create(val, score, state);
}

qadav_confidence* qadav_confidence_min(qadav_confidence* a, qadav_confidence* b) {
    if (!a || !b) return 0;
    double val = (a->value < b->value) ? a->value : b->value;
    int8_t state = (a->epistemic == 1 || b->epistemic == 1) ? 1 : ((a->epistemic == 2 || b->epistemic == 2) ? 2 : 0);
    double score = (state == 1) ? 0.0 : ((a->score < b->score) ? a->score : b->score);
    return qadav_confidence_create(val, score, state);
}

int qadav_confidence_assert(qadav_confidence* c, double threshold) {
    if (!c || c->epistemic == 1 || c->score < threshold) {
        const char* err1 = "Assertion failed: confidence score ";
        qadav_write_raw(err1, cir_strlen(err1));

        char scoreBuf[64];
        int sLen = cir_float_to_str(c ? c->score : 0.0, scoreBuf);
        qadav_write_raw(scoreBuf, sLen);

        const char* err2 = " < threshold ";
        qadav_write_raw(err2, cir_strlen(err2));

        char threshBuf[64];
        int tLen = cir_float_to_str(threshold, threshBuf);
        qadav_write_raw(threshBuf, tLen);

        const char* nl = "\n";
        qadav_write_raw(nl, cir_strlen(nl));

#if defined(_WIN32)
        ExitProcess(1);
#else
        __asm__ volatile("mov $60, %%rax\n\tmov $1, %%rdi\n\tsyscall" ::: "rax", "rdi");
#endif
        return 0;
    }
    return 1;
}

