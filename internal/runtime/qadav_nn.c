// Qadav neural network runtime implementation (zero libc)
#include "qadav_nn.h"

// ---------------------------------------------------------------------
// Zero-Libc Math Primitives (Newton / Taylor / Range-Reduction)
// ---------------------------------------------------------------------

double qadav_sqrt(double x) {
    if (x <= 0.0) return 0.0;
    double y = x;
    if (y > 1.0) y = x * 0.5;
    for (int i = 0; i < 15; i++) {
        y = 0.5 * (y + x / y);
    }
    return y;
}

double qadav_exp(double x) {
    if (x < -50.0) return 0.0;
    if (x > 50.0) x = 50.0;

    // Range reduction: x = k * ln2 + r, with k integer, r in [-ln2/2, ln2/2]
    const double LN2 = 0.693147180559945309417;
    const double INV_LN2 = 1.442695040888963407360;

    int k = (int)(x * INV_LN2 + (x >= 0.0 ? 0.5 : -0.5));
    double r = x - (double)k * LN2;

    // Taylor series on reduced interval: exp(r) = 1 + r + r^2/2! + ...
    double term = 1.0;
    double sum = 1.0;
    for (int i = 1; i <= 14; i++) {
        term *= r / (double)i;
        sum += term;
    }

    // Multiply by 2^k using IEEE-754 double exponent manipulation
    if (k >= -1022 && k <= 1023) {
        union {
            double d;
            unsigned long long u;
        } val;
        val.u = ((unsigned long long)(k + 1023)) << 52;
        return sum * val.d;
    }

    double scale = 1.0;
    if (k > 0) {
        while (k > 0) { scale *= 2.0; k--; }
    } else {
        while (k < 0) { scale *= 0.5; k++; }
    }
    return sum * scale;
}

double qadav_log(double x) {
    if (x <= 1e-15) return -34.53877639491; // log(1e-15)

    // Range reduction: x = m * 2^e with m in [1.0, 2.0)
    int e = 0;
    while (x >= 2.0) { x *= 0.5; e++; }
    while (x < 1.0)  { x *= 2.0; e--; }

    // ln(m) using z = (m - 1)/(m + 1), ln(m) = 2*(z + z^3/3 + z^5/5 + ...)
    double z = (x - 1.0) / (x + 1.0);
    double z2 = z * z;
    double term = z;
    double sum = z;
    for (int i = 3; i <= 17; i += 2) {
        term *= z2;
        sum += term / (double)i;
    }
    const double LN2 = 0.693147180559945309417;
    return 2.0 * sum + (double)e * LN2;
}

// ---------------------------------------------------------------------
// Deterministic Pseudorandom Number Generator (Xorshift32)
// ---------------------------------------------------------------------

static unsigned int g_qadav_seed = 133742;

static double qadav_random_uniform(void) {
    g_qadav_seed ^= (g_qadav_seed << 13);
    g_qadav_seed ^= (g_qadav_seed >> 17);
    g_qadav_seed ^= (g_qadav_seed << 5);
    return (double)(g_qadav_seed & 0x7FFFFFFF) / (double)0x7FFFFFFF;
}

// ---------------------------------------------------------------------
// Network Creation and Destruction
// ---------------------------------------------------------------------

qadav_network* qadav_network_create(int num_layers, const int* kinds, const int* in_sizes, const int* out_sizes) {
    if (num_layers <= 0) return (void*)0;

    qadav_network* net = (qadav_network*)qadav_alloc(sizeof(qadav_network));
    if (!net) return (void*)0;

    net->num_layers = num_layers;
    net->layers = (qadav_layer**)qadav_alloc(sizeof(qadav_layer*) * (unsigned long long)num_layers);
    if (!net->layers) {
        qadav_free(net);
        return (void*)0;
    }

    net->input_size = in_sizes[0];
    net->output_size = out_sizes[num_layers - 1];

    for (int i = 0; i < num_layers; i++) {
        qadav_layer* lay = (qadav_layer*)qadav_alloc(sizeof(qadav_layer));
        lay->kind = kinds[i];
        lay->in_size = in_sizes[i];
        lay->out_size = out_sizes[i];
        lay->last_input = (void*)0;
        lay->last_output = (void*)0;

        if (lay->kind == QADAV_LAYER_DENSE) {
            int w_dims[2] = { lay->out_size, lay->in_size };
            int b_dims[1] = { lay->out_size };

            lay->weights = qadav_tensor_create(QADAV_ELEM_FLOAT64, 2, w_dims);
            lay->biases  = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, b_dims);
            lay->grads_w = qadav_tensor_create(QADAV_ELEM_FLOAT64, 2, w_dims);
            lay->grads_b = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, b_dims);
            lay->m_w     = qadav_tensor_create(QADAV_ELEM_FLOAT64, 2, w_dims);
            lay->v_w     = qadav_tensor_create(QADAV_ELEM_FLOAT64, 2, w_dims);
            lay->m_b     = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, b_dims);
            lay->v_b     = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, b_dims);

            // Xavier / Glorot Initialization: scale = sqrt(6.0 / (fan_in + fan_out))
            double limit = qadav_sqrt(6.0 / (double)(lay->in_size + lay->out_size));
            int total_w = lay->out_size * lay->in_size;
            for (int w = 0; w < total_w; w++) {
                double val = (qadav_random_uniform() * 2.0 - 1.0) * limit;
                qadav_tensor_set_f64(lay->weights, w, val);
                qadav_tensor_set_f64(lay->grads_w, w, 0.0);
                qadav_tensor_set_f64(lay->m_w, w, 0.0);
                qadav_tensor_set_f64(lay->v_w, w, 0.0);
            }
            for (int b = 0; b < lay->out_size; b++) {
                qadav_tensor_set_f64(lay->biases, b, 0.0);
                qadav_tensor_set_f64(lay->grads_b, b, 0.0);
                qadav_tensor_set_f64(lay->m_b, b, 0.0);
                qadav_tensor_set_f64(lay->v_b, b, 0.0);
            }
        } else {
            lay->weights = (void*)0;
            lay->biases  = (void*)0;
            lay->grads_w = (void*)0;
            lay->grads_b = (void*)0;
            lay->m_w = (void*)0;
            lay->v_w = (void*)0;
            lay->m_b = (void*)0;
            lay->v_b = (void*)0;
        }

        net->layers[i] = lay;
    }

    return net;
}

void qadav_network_free(qadav_network* n) {
    if (!n) return;
    for (int i = 0; i < n->num_layers; i++) {
        qadav_layer* lay = n->layers[i];
        if (lay) {
            if (lay->weights) qadav_tensor_free(lay->weights);
            if (lay->biases)  qadav_tensor_free(lay->biases);
            if (lay->grads_w) qadav_tensor_free(lay->grads_w);
            if (lay->grads_b) qadav_tensor_free(lay->grads_b);
            if (lay->last_input)  qadav_tensor_free(lay->last_input);
            if (lay->last_output) qadav_tensor_free(lay->last_output);
            if (lay->m_w) qadav_tensor_free(lay->m_w);
            if (lay->v_w) qadav_tensor_free(lay->v_w);
            if (lay->m_b) qadav_tensor_free(lay->m_b);
            if (lay->v_b) qadav_tensor_free(lay->v_b);
            qadav_free(lay);
        }
    }
    qadav_free(n->layers);
    qadav_free(n);
}

// ---------------------------------------------------------------------
// Forward Pass
// ---------------------------------------------------------------------

static qadav_tensor* copy_tensor(const qadav_tensor* src) {
    if (!src) return (void*)0;
    qadav_tensor* dst = qadav_tensor_create(src->elem_kind, src->rank, src->shape);
    if (!dst) return (void*)0;
    for (int i = 0; i < src->total_size; i++) {
        qadav_tensor_set_f64(dst, i, qadav_tensor_get_f64(src, i));
    }
    return dst;
}

qadav_tensor* qadav_network_forward(qadav_network* n, const qadav_tensor* input) {
    if (!n || !input) return (void*)0;

    qadav_tensor* curr = copy_tensor(input);

    for (int i = 0; i < n->num_layers; i++) {
        qadav_layer* lay = n->layers[i];

        if (lay->last_input) qadav_tensor_free(lay->last_input);
        lay->last_input = copy_tensor(curr);

        qadav_tensor* next_out = (void*)0;

        switch (lay->kind) {
        case QADAV_LAYER_DENSE: {
            int out_dims[1] = { lay->out_size };
            next_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);
            for (int r = 0; r < lay->out_size; r++) {
                double sum = qadav_tensor_get_f64(lay->biases, r);
                for (int c = 0; c < lay->in_size; c++) {
                    double w = qadav_tensor_get_f64(lay->weights, r * lay->in_size + c);
                    double x = qadav_tensor_get_f64(curr, c);
                    sum += w * x;
                }
                qadav_tensor_set_f64(next_out, r, sum);
            }
            break;
        }

        case QADAV_LAYER_RELU: {
            int out_dims[1] = { curr->total_size };
            next_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);
            for (int j = 0; j < curr->total_size; j++) {
                double val = qadav_tensor_get_f64(curr, j);
                qadav_tensor_set_f64(next_out, j, val > 0.0 ? val : 0.0);
            }
            break;
        }

        case QADAV_LAYER_SIGMOID: {
            int out_dims[1] = { curr->total_size };
            next_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);
            for (int j = 0; j < curr->total_size; j++) {
                double val = qadav_tensor_get_f64(curr, j);
                double sig = 1.0 / (1.0 + qadav_exp(-val));
                qadav_tensor_set_f64(next_out, j, sig);
            }
            break;
        }

        case QADAV_LAYER_SOFTMAX: {
            int out_dims[1] = { curr->total_size };
            next_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);
            double max_val = qadav_tensor_get_f64(curr, 0);
            for (int j = 1; j < curr->total_size; j++) {
                double v = qadav_tensor_get_f64(curr, j);
                if (v > max_val) max_val = v;
            }
            double sum_exp = 0.0;
            for (int j = 0; j < curr->total_size; j++) {
                double exp_v = qadav_exp(qadav_tensor_get_f64(curr, j) - max_val);
                qadav_tensor_set_f64(next_out, j, exp_v);
                sum_exp += exp_v;
            }
            if (sum_exp > 0.0) {
                for (int j = 0; j < curr->total_size; j++) {
                    double norm_v = qadav_tensor_get_f64(next_out, j) / sum_exp;
                    qadav_tensor_set_f64(next_out, j, norm_v);
                }
            }
            break;
        }

        case QADAV_LAYER_LAYERNORM: {
            int out_dims[1] = { curr->total_size };
            next_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);
            double mean = 0.0;
            for (int j = 0; j < curr->total_size; j++) {
                mean += qadav_tensor_get_f64(curr, j);
            }
            mean /= (double)curr->total_size;

            double var = 0.0;
            for (int j = 0; j < curr->total_size; j++) {
                double diff = qadav_tensor_get_f64(curr, j) - mean;
                var += diff * diff;
            }
            var /= (double)curr->total_size;
            double std_inv = 1.0 / qadav_sqrt(var + 1e-5);

            for (int j = 0; j < curr->total_size; j++) {
                double norm_val = (qadav_tensor_get_f64(curr, j) - mean) * std_inv;
                qadav_tensor_set_f64(next_out, j, norm_val);
            }
            break;
        }

        default:
            next_out = copy_tensor(curr);
            break;
        }

        if (lay->last_output) qadav_tensor_free(lay->last_output);
        lay->last_output = copy_tensor(next_out);

        qadav_tensor_free(curr);
        curr = next_out;
    }

    return curr;
}

// ---------------------------------------------------------------------
// Backward Pass (Reverse-Mode Automatic Differentiation)
// ---------------------------------------------------------------------

void qadav_network_backward(qadav_network* n, const qadav_tensor* input, const qadav_tensor* grad_output) {
    if (!n || !grad_output) return;

    qadav_tensor* d_out = copy_tensor(grad_output);

    for (int i = n->num_layers - 1; i >= 0; i--) {
        qadav_layer* lay = n->layers[i];
        qadav_tensor* d_in = (void*)0;

        switch (lay->kind) {
        case QADAV_LAYER_DENSE: {
            int in_dims[1] = { lay->in_size };
            d_in = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, in_dims);

            const qadav_tensor* in_t = lay->last_input ? lay->last_input : input;

            // Compute input gradient: d_in = W^T * d_out
            for (int c = 0; c < lay->in_size; c++) {
                double sum = 0.0;
                for (int r = 0; r < lay->out_size; r++) {
                    double w = qadav_tensor_get_f64(lay->weights, r * lay->in_size + c);
                    double g = qadav_tensor_get_f64(d_out, r);
                    sum += w * g;
                }
                qadav_tensor_set_f64(d_in, c, sum);
            }

            // Accumulate weight & bias gradients
            for (int r = 0; r < lay->out_size; r++) {
                double g = qadav_tensor_get_f64(d_out, r);
                double curr_gb = qadav_tensor_get_f64(lay->grads_b, r);
                qadav_tensor_set_f64(lay->grads_b, r, curr_gb + g);

                for (int c = 0; c < lay->in_size; c++) {
                    double x = qadav_tensor_get_f64(in_t, c);
                    int idx = r * lay->in_size + c;
                    double curr_gw = qadav_tensor_get_f64(lay->grads_w, idx);
                    qadav_tensor_set_f64(lay->grads_w, idx, curr_gw + g * x);
                }
            }
            break;
        }

        case QADAV_LAYER_RELU: {
            int in_dims[1] = { d_out->total_size };
            d_in = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, in_dims);
            for (int j = 0; j < d_out->total_size; j++) {
                double x = lay->last_input ? qadav_tensor_get_f64(lay->last_input, j) : 0.0;
                double g = qadav_tensor_get_f64(d_out, j);
                qadav_tensor_set_f64(d_in, j, x > 0.0 ? g : 0.0);
            }
            break;
        }

        case QADAV_LAYER_SIGMOID: {
            int in_dims[1] = { d_out->total_size };
            d_in = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, in_dims);
            for (int j = 0; j < d_out->total_size; j++) {
                double y = lay->last_output ? qadav_tensor_get_f64(lay->last_output, j) : 0.5;
                double g = qadav_tensor_get_f64(d_out, j);
                qadav_tensor_set_f64(d_in, j, g * y * (1.0 - y));
            }
            break;
        }

        case QADAV_LAYER_SOFTMAX:
        case QADAV_LAYER_LAYERNORM:
        default:
            d_in = copy_tensor(d_out);
            break;
        }

        qadav_tensor_free(d_out);
        d_out = d_in;
    }

    if (d_out) qadav_tensor_free(d_out);
}

// ---------------------------------------------------------------------
// Optimizers
// ---------------------------------------------------------------------

void qadav_network_sgd_step(qadav_network* n, double lr) {
    if (!n) return;
    for (int i = 0; i < n->num_layers; i++) {
        qadav_layer* lay = n->layers[i];
        if (lay->kind == QADAV_LAYER_DENSE) {
            int total_w = lay->out_size * lay->in_size;
            for (int w = 0; w < total_w; w++) {
                double weight = qadav_tensor_get_f64(lay->weights, w);
                double grad   = qadav_tensor_get_f64(lay->grads_w, w);
                qadav_tensor_set_f64(lay->weights, w, weight - lr * grad);
                qadav_tensor_set_f64(lay->grads_w, w, 0.0); // Reset gradient
            }
            for (int b = 0; b < lay->out_size; b++) {
                double bias = qadav_tensor_get_f64(lay->biases, b);
                double grad = qadav_tensor_get_f64(lay->grads_b, b);
                qadav_tensor_set_f64(lay->biases, b, bias - lr * grad);
                qadav_tensor_set_f64(lay->grads_b, b, 0.0); // Reset gradient
            }
        }
    }
}

void qadav_network_adam_step(qadav_network* n, double lr, int step) {
    if (!n) return;
    const double beta1 = 0.9;
    const double beta2 = 0.999;
    const double eps = 1e-8;

    double beta1_t = 1.0;
    double beta2_t = 1.0;
    for (int s = 0; s < step; s++) {
        beta1_t *= beta1;
        beta2_t *= beta2;
    }
    double fix1 = 1.0 - beta1_t;
    double fix2 = 1.0 - beta2_t;
    if (fix1 < 1e-7) fix1 = 1e-7;
    if (fix2 < 1e-7) fix2 = 1e-7;

    for (int i = 0; i < n->num_layers; i++) {
        qadav_layer* lay = n->layers[i];
        if (lay->kind == QADAV_LAYER_DENSE) {
            int total_w = lay->out_size * lay->in_size;
            for (int w = 0; w < total_w; w++) {
                double g = qadav_tensor_get_f64(lay->grads_w, w);
                double m = beta1 * qadav_tensor_get_f64(lay->m_w, w) + (1.0 - beta1) * g;
                double v = beta2 * qadav_tensor_get_f64(lay->v_w, w) + (1.0 - beta2) * g * g;

                qadav_tensor_set_f64(lay->m_w, w, m);
                qadav_tensor_set_f64(lay->v_w, w, v);

                double m_hat = m / fix1;
                double v_hat = v / fix2;

                double weight = qadav_tensor_get_f64(lay->weights, w);
                qadav_tensor_set_f64(lay->weights, w, weight - lr * m_hat / (qadav_sqrt(v_hat) + eps));
                qadav_tensor_set_f64(lay->grads_w, w, 0.0);
            }
            for (int b = 0; b < lay->out_size; b++) {
                double g = qadav_tensor_get_f64(lay->grads_b, b);
                double m = beta1 * qadav_tensor_get_f64(lay->m_b, b) + (1.0 - beta1) * g;
                double v = beta2 * qadav_tensor_get_f64(lay->v_b, b) + (1.0 - beta2) * g * g;

                qadav_tensor_set_f64(lay->m_b, b, m);
                qadav_tensor_set_f64(lay->v_b, b, v);

                double m_hat = m / fix1;
                double v_hat = v / fix2;

                double bias = qadav_tensor_get_f64(lay->biases, b);
                qadav_tensor_set_f64(lay->biases, b, bias - lr * m_hat / (qadav_sqrt(v_hat) + eps));
                qadav_tensor_set_f64(lay->grads_b, b, 0.0);
            }
        }
    }
}

// ---------------------------------------------------------------------
// Loss Functions
// ---------------------------------------------------------------------

double qadav_loss_mse(const qadav_tensor* pred, const qadav_tensor* target) {
    if (!pred || !target) return 0.0;
    double sum = 0.0;
    int n = pred->total_size;
    for (int i = 0; i < n; i++) {
        double diff = qadav_tensor_get_f64(pred, i) - qadav_tensor_get_f64(target, i);
        sum += diff * diff;
    }
    return sum / (double)n;
}

double qadav_loss_cross_entropy(const qadav_tensor* pred, const qadav_tensor* target) {
    if (!pred || !target) return 0.0;
    double sum = 0.0;
    int n = pred->total_size;
    for (int i = 0; i < n; i++) {
        double p = qadav_tensor_get_f64(pred, i);
        double t = qadav_tensor_get_f64(target, i);
        sum -= t * qadav_log(p);
    }
    return sum;
}

// ---------------------------------------------------------------------
// End-to-End Training Routines
// ---------------------------------------------------------------------

double qadav_train_step(qadav_network* n, const qadav_tensor* x, const qadav_tensor* y, int loss_kind, int opt_kind, double lr, int step) {
    if (!n || !x || !y) return 0.0;

    // 1. Forward pass
    qadav_tensor* pred = qadav_network_forward(n, x);
    if (!pred) return 0.0;

    // 2. Compute loss and loss gradient
    int out_dims[1] = { pred->total_size };
    qadav_tensor* grad_out = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, out_dims);

    double loss_val = 0.0;
    if (loss_kind == QADAV_LOSS_CROSS_ENTROPY) {
        loss_val = qadav_loss_cross_entropy(pred, y);
        for (int i = 0; i < pred->total_size; i++) {
            double p = qadav_tensor_get_f64(pred, i);
            double t = qadav_tensor_get_f64(y, i);
            qadav_tensor_set_f64(grad_out, i, p - t);
        }
    } else { // QADAV_LOSS_MSE default
        loss_val = qadav_loss_mse(pred, y);
        double scale = 2.0 / (double)pred->total_size;
        for (int i = 0; i < pred->total_size; i++) {
            double diff = qadav_tensor_get_f64(pred, i) - qadav_tensor_get_f64(y, i);
            qadav_tensor_set_f64(grad_out, i, scale * diff);
        }
    }

    // 3. Backward pass
    qadav_network_backward(n, x, grad_out);

    // 4. Optimizer update step
    if (opt_kind == QADAV_OPT_ADAM) {
        qadav_network_adam_step(n, lr, step);
    } else {
        qadav_network_sgd_step(n, lr);
    }

    qadav_tensor_free(pred);
    qadav_tensor_free(grad_out);

    return loss_val;
}

void qadav_train_epochs(qadav_network* n, const qadav_tensor* dataset_x, const qadav_tensor* dataset_y, int loss_kind, int opt_kind, double lr, int epochs) {
    if (!n || !dataset_x || !dataset_y) return;

    int num_samples = 1;
    if (dataset_x->rank >= 2) {
        num_samples = dataset_x->shape[0];
    }

    int in_dim = n->input_size;
    int out_dim = n->output_size;

    int x_dims[1] = { in_dim };
    int y_dims[1] = { out_dim };

    qadav_tensor* x_sample = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, x_dims);
    qadav_tensor* y_sample = qadav_tensor_create(QADAV_ELEM_FLOAT64, 1, y_dims);

    int step = 1;
    for (int epoch = 0; epoch < epochs; epoch++) {
        for (int s = 0; s < num_samples; s++) {
            // Extract sample s
            for (int i = 0; i < in_dim; i++) {
                int idx = (num_samples > 1) ? (s * in_dim + i) : i;
                qadav_tensor_set_f64(x_sample, i, qadav_tensor_get_f64(dataset_x, idx));
            }
            for (int i = 0; i < out_dim; i++) {
                int idx = (num_samples > 1) ? (s * out_dim + i) : i;
                qadav_tensor_set_f64(y_sample, i, qadav_tensor_get_f64(dataset_y, idx));
            }

            qadav_train_step(n, x_sample, y_sample, loss_kind, opt_kind, lr, step);
            step++;
        }
    }

    qadav_tensor_free(x_sample);
    qadav_tensor_free(y_sample);
}

