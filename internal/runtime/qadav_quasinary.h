#ifndef QADAV_QUASINARY_H
#define QADAV_QUASINARY_H

#include "qadav_tensor.h"

int  qadav_quasi_add(long long a, long long b);   /* saturated */
int  qadav_quasi_mul(long long a, long long b);
int  qadav_quasi_neg(long long a);

qadav_tensor* qadav_quasi_grid_create(int rows, int cols);
long long     qadav_quasi_grid_get(qadav_tensor* g, int r, int c);
void          qadav_quasi_grid_set(qadav_tensor* g, int r, int c, long long v);

double qadav_quasi_energy_ising(qadav_tensor* grid);
double qadav_quasi_energy_majority(qadav_tensor* grid);

void   qadav_quasi_minimize(qadav_tensor* grid, int iterations);

void   qadav_quasi_print(long long v);
void   qadav_quasi_grid_print(qadav_tensor* grid);

#endif

