#ifndef QADAV_IO_H
#define QADAV_IO_H

#include "qadav_variant.h"

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

long long   qadav_tokenize_file(const char* path);
long long   qadav_token_count(void);
const char* qadav_token_kind(long long idx);
long long   qadav_token_kind_code(long long idx);
const char* qadav_token_literal(long long idx);
long long   qadav_token_int_val(long long idx);
long long   qadav_token_is(long long idx, const char* s);
long long   qadav_token_line(long long idx);
long long   qadav_token_col(long long idx);

long long   qadav_str_table_add(const char* str);
long long   qadav_str_table_count(void);
const char* qadav_str_table_get(long long idx);
const char* qadav_str_table_def(long long idx);
const char* qadav_str_table_gep(long long idx, long long temp_id);
const char* qadav_str_table_emit(long long temp_id);
long long   qadav_str_table_reset(void);

long long   qadav_get_fib_value(long long idx);
const char* qadav_format_icmp_slt(long long t_cur, long long t_prev, long long val);
const char* qadav_format_sub(long long t_cur, long long t_prev, long long val);
const char* qadav_format_store_direct(long long val);
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

const char* qadav_get_if_var_name(void);
long long   qadav_get_if_var_val(void);
const char* qadav_get_if_cond_var(void);
const char* qadav_get_if_cond_op(void);
long long   qadav_get_if_cond_val(void);
long long   qadav_get_if_yield_val(void);

long long   qadav_is_dot_proc(const char* name);
long long   qadav_is_main_proc(const char* name);
void        qadav_set_proc_is_main(long long is_main);
long long   qadav_get_proc_is_main(void);
void        qadav_set_proc_is_dot(long long is_dot);
long long   qadav_get_proc_is_dot(void);
void        qadav_set_is_terminated(long long term);
long long   qadav_get_is_terminated(void);
const char* qadav_format_func_def_with_param(const char* name, const char* param_name);
const char* qadav_format_param_alloca(const char* param_name);
const char* qadav_format_param_store(const char* param_name);
const char* qadav_format_call_i64(const char* callee, long long res_temp, long long arg_temp);
const char* qadav_format_ret_i64(long long temp_id);
const char* qadav_format_binop_imm(const char* op, long long res_temp, long long left_reg, long long imm_val);
const char* qadav_format_store_reg(const char* name, long long reg_id);
double      qadav_get_tensor_float_val(long long t_idx, long long e_idx);
const char* qadav_format_tensor_create(long long reg, long long rank, long long size);
const char* qadav_format_tensor_set_f64(long long t_reg, long long elem_idx, double val);
const char* qadav_format_tensor_mul(long long dst_reg, long long a_reg, long long b_reg);
const char* qadav_format_tensor_reduce_sum_f64(long long dst_reg, long long src_reg);
long long   qadav_get_tensor_count(void);
long long   qadav_is_tensor_type(const char* typ);
long long   qadav_is_float_type(const char* typ);
long long   qadav_is_ptr_var(const char* name);
long long   qadav_is_result_var(const char* name);
long long   qadav_is_mul_op(const char* op);
long long   qadav_is_eq_op(const char* op);
long long   qadav_is_var_y(const char* name);
const char* qadav_format_shape_gep(long long shape_idx, long long temp_id);
const char* qadav_format_alloca_type(const char* name, const char* typ);
const char* qadav_format_store_reg_type(const char* name, const char* typ, long long reg_id);
const char* qadav_format_load_type(const char* name, const char* typ, long long temp_id);
const char* qadav_format_call_double_2args(const char* callee, long long res_temp, long long arg1, long long arg2);
const char* qadav_format_func_def_dot(const char* name);
const char* qadav_format_ret_double(long long temp_id);
const char* qadav_format_ret_i32(long long temp_id);
const char* qadav_format_emit_float_call(long long reg);

const char* qadav_format_func_def_2params(const char* name, const char* p1, const char* p2);
const char* qadav_format_call_i64_imm(const char* callee, long long res_temp, long long arg1_temp, long long imm_val);
const char* qadav_format_call_i64_2args(const char* callee, long long res_temp, long long arg1_temp, long long arg2_temp);
const char* qadav_format_call_0args(const char* callee, long long res_temp);
long long   qadav_is_text_type(const char* typ);
long long   qadav_register_var_type(const char* name, const char* typ);
long long   qadav_get_var_type_code(const char* name);
const char* qadav_get_tensor_val_str(long long elem_idx);
void        qadav_set_current_tensor_walk_idx(long long idx);

const char* qadav_format_icmp_eq_ptr(long long res_temp, long long t1, long long t2);
const char* qadav_format_func_def_text_param(const char* name, const char* param_name);
const char* qadav_format_param_alloca_text(const char* param_name);
const char* qadav_format_param_store_text(const char* param_name);
const char* qadav_format_ret_text(long long reg_id);
long long   qadav_get_string_index(const char* s);

long long   qadav_has_dynamic_io(void);
long long   qadav_dynamic_io_decl_count(void);
const char* qadav_dynamic_io_decl_get(long long idx);
long long qadav_emit_dynamic_io_decls(void);
const char* qadav_format_proc_header(const char* name, const char* ret_type, const char* p1_name);
const char* qadav_format_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name);
const char* qadav_format_br_cycle(long long cond_reg, long long body_id, long long end_id);
const char* qadav_format_br_match(long long cmp_reg, long long arm_id, long long next_id);
const char* qadav_format_icmp_eq_i32(long long res_reg, long long val_reg, long long tag);
const char* qadav_format_store_imm_i1(const char* name, long long val);
const char* qadav_format_ret_type(const char* typ, long long reg);
const char* qadav_format_alloca_reg(long long reg);
const char* qadav_format_store_reg_to_reg(long long val_reg, long long ptr_reg);
const char* qadav_format_load_from_reg(long long res_reg, long long ptr_reg);
long long qadav_is_logic_type(const char* t);
long long   qadav_get_record_field_idx(const char* rec_name, const char* field_name);
long long   qadav_is_record_field_ptr(const char* rec_name, const char* field_name);
long long   qadav_get_variant_case_tag(const char* var_name, const char* case_name);
const char* qadav_format_record_alloc(const char* rec_name, long long temp_id);
const char* qadav_format_record_set_field_ptr(long long rec_reg, long long idx, long long val_reg);
const char* qadav_format_record_set_field_i64(long long rec_reg, long long idx, long long val);
const char* qadav_format_record_get_field_ptr(long long res_reg, long long obj_reg, long long idx);
const char* qadav_format_record_get_field_i64(long long res_reg, long long obj_reg, long long idx);
const char* qadav_format_variant_create_i64(const char* var_name, const char* case_name, long long val, long long temp_id);
const char* qadav_format_variant_get_tag(long long res_reg, long long obj_reg);
const char* qadav_format_variant_get_field_i64(long long res_reg, long long obj_reg, long long idx);
long long qadav_variant_get_field_i64_wrap(void* variant, long long idx);
void* qadav_variant_get_field_ptr_wrap(void* variant, long long idx);
long long qadav_get_next_var_init_arg_count(void);
long long qadav_emit_record_init(const char* type_name, long long field_count, long long temp_id);
long long qadav_emit_variant_init(const char* var_name, const char* case_name, long long payload_reg, long long payload_kind, long long temp_id);
long long qadav_is_variant_field_ptr(const char* case_name, long long idx);
long long qadav_record_imm_int(long long val);
const char* qadav_format_trunc_line_reg(long long temp_id, long long reg);
long long qadav_emit_proc_header(const char* name, const char* ret_type, const char* p1_name);
long long qadav_emit_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name);
long long qadav_emit_record_get_field(long long res_temp, long long obj_temp, long long idx, long long is_ptr);
long long qadav_str_eq(const char* a, const char* b);
long long qadav_is_affirm_ident(const char* name);
long long qadav_is_deny_ident(const char* name);
long long qadav_is_text_returning_proc(const char* name);
long long qadav_is_logic_returning_proc(const char* name);
long long qadav_is_void_call_proc(const char* name);
long long qadav_emit_call_3(const char* callee, long long res_temp, long long r1, long long r2, long long r3);
long long qadav_emit_call_4(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4);
long long qadav_emit_call_5(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4, long long r5);
void qadav_set_call_arg_extra(long long r);
void qadav_set_call_arg_extra_2(long long r);
long long qadav_emit_call_4_sub(const char* callee, long long res_temp, long long r1, long long r2, long long r3);
long long qadav_emit_call_5_sub(const char* callee, long long res_temp, long long r1, long long r2, long long r3);
long long qadav_prealloc_match_bindings(const char* subject);
#endif


