#include "qadav_abstain.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

extern void* qadav_alloc(unsigned long long sz);
extern void  qadav_free(void* ptr);

static unsigned long long g_abstain_create_count = 0;
static unsigned long long g_abstain_free_count = 0;

unsigned long long qadav_abstain_get_create_count(void) {
    return g_abstain_create_count;
}

unsigned long long qadav_abstain_get_free_count(void) {
    return g_abstain_free_count;
}

qadav_abstain_policy* qadav_abstain_policy_create(double threshold, int mode, int max_retries) {
    qadav_abstain_policy* p = (qadav_abstain_policy*)qadav_alloc(sizeof(qadav_abstain_policy));
    if (!p) return 0;
    p->threshold = threshold;
    p->mode = mode;
    p->max_retries = max_retries;
    g_abstain_create_count++;
    return p;
}

void qadav_abstain_policy_free(qadav_abstain_policy* p) {
    if (!p) return;
    g_abstain_free_count++;
    qadav_free(p);
}

static double qadav_pow_int(double base, int exp) {
    double res = 1.0;
    for (int i = 0; i < exp; i++) {
        res *= base;
    }
    return res;
}

static double qadav_nth_root(double p, int n) {
    if (n <= 1) return p;
    if (p <= 0.0) return 0.0;
    if (p >= 1.0) return 1.0;
    double low = 0.0, high = 1.0, mid = 0.5;
    for (int iter = 0; iter < 45; iter++) {
        mid = (low + high) * 0.5;
        double val = qadav_pow_int(mid, n);
        if (val < p) {
            low = mid;
        } else {
            high = mid;
        }
    }
    return mid;
}

double qadav_abstain_compute_confidence(qadav_reasoning_chain* chain) {
    if (!chain || chain->n_steps <= 0) {
        return 1.0; // Empty chain = full confidence baseline
    }

    int n = chain->n_steps;
    double prod = 1.0;
    double sum = 0.0;

    for (int i = 0; i < n; i++) {
        double c = 1.0;
        if (chain->certs) {
            c = ((qadav_reason_certificate*)chain->certs)[i].confidence;
        } else if (chain->steps) {
            c = chain->steps[i].confidence;
        }
        prod *= c;
        sum += c;
    }

    double mean = sum / (double)n;
    double var_sum = 0.0;
    for (int i = 0; i < n; i++) {
        double c = 1.0;
        if (chain->certs) {
            c = ((qadav_reason_certificate*)chain->certs)[i].confidence;
        } else if (chain->steps) {
            c = chain->steps[i].confidence;
        }
        double diff = c - mean;
        var_sum += diff * diff;
    }
    double var = var_sum / (double)n;

    double geom = qadav_nth_root(prod, n);
    double penalty = 1.0 - (var / 4.0);
    if (penalty < 0.0) penalty = 0.0;

    return geom * penalty;
}

int qadav_abstain_check(
    qadav_reasoning_chain* chain,
    qadav_abstain_policy* policy,
    int current_retry_count
) {
    if (!policy || policy->mode == 0) {
        return 0; // No abstention policy configured
    }

    double conf = qadav_abstain_compute_confidence(chain);

    // If aggregated confidence meets or exceeds threshold, do not abstain
    if (conf >= policy->threshold) {
        return 0;
    }

    // If in RETRY mode, enforce maximum retry limit
    if (policy->mode == QADAV_ABSTAIN_RETRY) {
        if (policy->max_retries > 0 && current_retry_count >= policy->max_retries) {
            return 1; // Retry budget exhausted, must abstain
        }
    }

    return 1; // Confidence below threshold: signal abstention
}

extern void qadav_write_raw(const char* s, unsigned long long len);

static unsigned long long abstain_strlen(const char* s) {
    if (!s) return 0;
    unsigned long long len = 0;
    while (s[len]) len++;
    return len;
}

static void abstain_print_num(long long v) {
    if (v == 0) {
        qadav_write_raw("0", 1);
        return;
    }
    char tmp[32];
    char buf[32];
    int ti = 0, bi = 0;
    while (v > 0) {
        tmp[ti++] = '0' + (char)(v % 10);
        v /= 10;
    }
    while (ti > 0) buf[bi++] = tmp[--ti];
    qadav_write_raw(buf, bi);
}

static void abstain_print_float(double val) {
    if (val < 0.0) {
        qadav_write_raw("-", 1);
        val = -val;
    }
    long long int_part = (long long)val;
    double frac = val - (double)int_part;
    long long frac_part = (long long)(frac * 1000000.0 + 0.5);
    abstain_print_num(int_part);
    qadav_write_raw(".", 1);
    char buf[7];
    for (int i = 5; i >= 0; i--) {
        buf[i] = '0' + (char)(frac_part % 10);
        frac_part /= 10;
    }
    buf[6] = '\0';
    qadav_write_raw(buf, 6);
}

int qadav_abstain_execute(
    qadav_reasoning_chain* chain,
    qadav_abstain_policy* policy
) {
    qadav_abstain_policy default_policy;
    if (!policy) {
        default_policy.threshold = 0.80;
        default_policy.mode = QADAV_ABSTAIN_REFUSE;
        default_policy.max_retries = 0;
        policy = &default_policy;
    }

    double conf = qadav_abstain_compute_confidence(chain);
    int should_abstain = qadav_abstain_check(chain, policy, 0);

    qadav_write_raw("--- TTPCR CALIBRATED ABSTENTION ---\n", 36);
    qadav_write_raw("Chain: ", 7);
    if (chain && chain->chain_name[0]) {
        qadav_write_raw(chain->chain_name, abstain_strlen(chain->chain_name));
    } else {
        qadav_write_raw("uncertain_logic", 15);
    }
    qadav_write_raw(" | Threshold: ", 14);
    abstain_print_float(policy->threshold);
    qadav_write_raw(" | Calibrated: ", 15);
    abstain_print_float(conf);
    qadav_write_raw("\n", 1);

    if (should_abstain) {
        qadav_write_raw("Action: REFUSE (Confidence below safety threshold)\n", 51);
        qadav_write_raw("Status: GRACEFUL_ABSTENTION_SIGNALED\n", 37);
    } else {
        qadav_write_raw("Action: PROCEED (Confidence sufficient)\n", 40);
        qadav_write_raw("Status: PROCEED_AUTHORIZED\n", 27);
    }
    qadav_write_raw("-----------------------------------\n", 36);

    return should_abstain ? policy->mode : 0;
}
