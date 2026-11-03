// Qadav neural network runtime header (zero libc)
#ifndef QADAV_NN_H
#define QADAV_NN_H

#include "qadav_tensor.h"

// Layer kinds
#define QADAV_LAYER_DENSE      1
#define QADAV_LAYER_RELU       2
#define QADAV_LAYER_SIGMOID    3
#define QADAV_LAYER_SOFTMAX    4
#define QADAV_LAYER_LAYERNORM  5

// Loss kinds
#define QADAV_LOSS_MSE           1
#define QADAV_LOSS_CROSS_ENTROPY 2

// Optimizer kinds
#define QADAV_OPT_SGD  1
#define QADAV_OPT_ADAM 2

typedef struct qadav_layer {
    int           kind;
    int           in_size;
    int           out_size;
    qadav_tensor* weights;       // 2D tensor [out_size, in_size]
    qadav_tensor* biases;        // 1D tensor [out_size]
    qadav_tensor* grads_w;       // 2D gradient accumulator
    qadav_tensor* grads_b;       // 1D gradient accumulator
    qadav_tensor* last_input;    // cached for backpropagation
    qadav_tensor* last_output;   // cached for backpropagation
    // Adam momentum buffers
    qadav_tensor* m_w;
    qadav_tensor* v_w;
    qadav_tensor* m_b;
    qadav_tensor* v_b;
} qadav_layer;

typedef struct qadav_network {
    int           num_layers;
    qadav_layer** layers;
    int           input_size;
    int           output_size;
} qadav_network;

// Memory & Lifecycle
qadav_network* qadav_network_create(int num_layers, const int* kinds, const int* in_sizes, const int* out_sizes);
void           qadav_network_free(qadav_network* n);

// Forward Pass: input tensor -> output tensor
qadav_tensor*  qadav_network_forward(qadav_network* n, const qadav_tensor* input);

// Backward Pass: reverse-mode autodiff with gradient accumulation
void           qadav_network_backward(qadav_network* n, const qadav_tensor* input, const qadav_tensor* grad_output);

// Optimizer Steps
void           qadav_network_sgd_step(qadav_network* n, double lr);
void           qadav_network_adam_step(qadav_network* n, double lr, int step);

// Loss Functions
double         qadav_loss_mse(const qadav_tensor* pred, const qadav_tensor* target);
double         qadav_loss_cross_entropy(const qadav_tensor* pred, const qadav_tensor* target);

// Training Execution
double         qadav_train_step(qadav_network* n, const qadav_tensor* x, const qadav_tensor* y, int loss_kind, int opt_kind, double lr, int step);
void           qadav_train_epochs(qadav_network* n, const qadav_tensor* dataset_x, const qadav_tensor* dataset_y, int loss_kind, int opt_kind, double lr, int epochs);

// Zero-libc math functions (Newton / Taylor / CORDIC)
double         qadav_sqrt(double x);
double         qadav_exp(double x);
double         qadav_log(double x);

#endif // QADAV_NN_H

