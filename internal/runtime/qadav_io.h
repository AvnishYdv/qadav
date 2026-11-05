#ifndef QADAV_IO_H
#define QADAV_IO_H

// QDS file format (Qadav Dataset Format):
//   Header (32 bytes):
//     magic[4]       = "QDS1"
//     n_tensors[4]   = uint32 little-endian
//     total_bytes[8] = uint64
//     reserved[16]
//   For each tensor:
//     name_len[4]    = uint32
//     name[name_len]
//     rank[4]        = uint32
//     dims[rank*4]   = uint32 each
//     elem_kind[4]   = uint32 (0=int64, 1=float64)
//     data_len[8]    = uint64
//     data[data_len]

void* qadav_file_open_read(const char* path);
void* qadav_file_open_write(const char* path);
int   qadav_file_read(void* handle, void* buf, unsigned long long len);
int   qadav_file_write(void* handle, const void* buf, unsigned long long len);
void  qadav_file_close(void* handle);

const char* qadav_get_input_path(void);
long long   qadav_extract_yield_number(const char* path);
const char* qadav_format_trunc_line(long long val);

int         qadav_scan_fib_values(const char* path);
long long   qadav_get_fib_value(long long idx);
const char* qadav_format_icmp_slt(long long t_cur, long long t_prev, long long val);
const char* qadav_format_sub(long long t_cur, long long t_prev, long long val);
const char* qadav_format_store_direct(long long val);

long long   qadav_scan_emitted_strings(const char* path);
long long   qadav_get_emitted_string_count(void);
const char* qadav_format_emitted_str_def(long long idx);
const char* qadav_format_emitted_str_gep(long long str_idx, long long temp_id);
const char* qadav_format_emitted_str_call(long long temp_id);
const char* qadav_format_trunc_line_id(long long temp_id, long long val);
const char* qadav_format_ret_line_id(long long temp_id);

long long   qadav_register_string(const char* s);
long long   qadav_get_string_index(const char* s);
long long   qadav_get_string_byte_len(const char* s);
const char* qadav_escape_string(const char* s);
void        qadav_set_last_int_value(long long v);
long long   qadav_last_int_value(void);

void        qadav_set_last_temp(long long t);
long long   qadav_get_last_temp(void);
void        qadav_set_last_emit_kind(long long k);
long long   qadav_get_last_emit_kind(void);
const char* qadav_stmt_binop_op(long long idx);
const char* qadav_stmt_binop_left(long long idx);
const char* qadav_stmt_binop_right(long long idx);

long long   qadav_next_label_id(void);
void        qadav_reset_label_counter(void);
const char* qadav_format_icmp_imm(const char* pred, long long res_temp, long long left_temp, long long imm_val);
const char* qadav_format_br_cond(long long cond_reg, long long then_id, long long else_id);
const char* qadav_format_br_cond_merge(long long cond_reg, long long then_id, long long merge_id);
const char* qadav_format_label(const char* prefix, long long id);
const char* qadav_format_br_label(const char* prefix, long long id);

int         qadav_scan_if_else(const char* path);
const char* qadav_get_if_var_name(void);
long long   qadav_get_if_var_val(void);
const char* qadav_get_if_cond_var(void);
const char* qadav_get_if_cond_op(void);
long long   qadav_get_if_cond_val(void);
long long   qadav_get_if_yield_val(void);

long long   qadav_is_main_proc(const char* name);
void        qadav_set_proc_is_main(long long is_main);
long long   qadav_get_proc_is_main(void);
const char* qadav_format_func_def_with_param(const char* name, const char* param_name);
const char* qadav_format_param_alloca(const char* param_name);
const char* qadav_format_param_store(const char* param_name);
const char* qadav_format_call_i64(const char* callee, long long res_temp, long long arg_temp);
const char* qadav_format_ret_i64(long long temp_id);
const char* qadav_format_binop_imm(const char* op, long long res_temp, long long left_reg, long long imm_val);
const char* qadav_format_store_reg(const char* name, long long reg_id);

#endif

