// Qadav minimal Windows runtime — zero CRT, zero libc
#if defined(_WIN32)

int _fltused = 0;

#define STD_OUTPUT_HANDLE ((unsigned long)-11)
#define NULL ((void*)0)

typedef unsigned long DWORD;
typedef void* HANDLE;
typedef int BOOL;

__declspec(dllimport) HANDLE __stdcall GetStdHandle(DWORD nStdHandle);
__declspec(dllimport) BOOL __stdcall WriteFile(HANDLE hFile, const void* lpBuffer, DWORD nNumberOfBytesToWrite, DWORD* lpNumberOfBytesWritten, void* lpOverlapped);
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
__declspec(dllimport) void* __stdcall HeapAlloc(HANDLE hHeap, DWORD dwFlags, unsigned long long dwBytes);
__declspec(dllimport) BOOL  __stdcall HeapFree(HANDLE hHeap, DWORD dwFlags, void* lpMem);
__declspec(dllimport) HANDLE __stdcall GetProcessHeap(void);
__declspec(dllimport) char* __stdcall GetCommandLineA(void);

static int rt_str_contains(const char* haystack, const char* needle) {
    if (!haystack || !needle) return 0;
    if (!*needle) return 1;
    for (; *haystack; haystack++) {
        const char* h = haystack;
        const char* n = needle;
        while (*h && *n && (*h == *n)) {
            h++;
            n++;
        }
        if (!*n) return 1;
    }
    return 0;
}

long long get_target_kind(void) {
    const char* cmd = GetCommandLineA();
    if (!cmd) return 0;
    if (rt_str_contains(cmd, "fib")) return 1;
    if (rt_str_contains(cmd, "dot")) return 2;
    return 0;
}

const char* qadav_get_input_path(void);
long long   qadav_extract_yield_number(const char* path);
const char* qadav_format_trunc_line(long long val);

long long   qadav_stmt_kind(long long idx);
const char* qadav_stmt_name(long long idx);
long long   qadav_stmt_value(long long idx);
const char* qadav_stmt_expr(long long idx);
const char* qadav_format_alloca(const char* name);
const char* qadav_format_store(const char* name, long long val);
const char* qadav_format_load(const char* name, long long temp_id);
const char* qadav_format_binop(const char* op, long long res_temp, long long t1, long long t2);
const char* qadav_format_emit_int_call(long long temp_id);
const char* qadav_format_trunc(long long temp_id, long long val);
const char* qadav_format_ret(long long temp_id);

long long   qadav_get_fib_value(long long idx);
const char* qadav_format_icmp_slt(long long t_cur, long long t_prev, long long val);
const char* qadav_format_sub(long long t_cur, long long t_prev, long long val);
const char* qadav_format_store_direct(long long val);

const char* qadav_get_tensor_name(long long idx);
long long   qadav_get_tensor_size(long long idx);
long long   qadav_get_tensor_int_val(long long t_idx, long long e_idx);
const char* qadav_format_shape_const(long long shape_idx, long long size);
const char* qadav_format_tensor_set(long long t_temp, long long elem_idx, long long int_val);

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
long long   qadav_is_text_type(const char* typ);
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
long long   qadav_register_var_type(const char* name, const char* typ);
long long   qadav_get_var_type_code(const char* name);
const char* qadav_get_tensor_val_str(long long elem_idx);
void        qadav_set_current_tensor_walk_idx(long long idx);

const char* qadav_format_icmp_eq_ptr(long long res_temp, long long t1, long long t2);
const char* qadav_format_func_def_text_param(const char* name, const char* param_name);
const char* qadav_format_param_alloca_text(const char* param_name);
const char* qadav_format_param_store_text(const char* param_name);
const char* qadav_format_ret_text(long long reg_id);
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
long long   qadav_has_dynamic_io(void);
long long   qadav_dynamic_io_decl_count(void);
const char* qadav_dynamic_io_decl_get(long long idx);
long long qadav_emit_dynamic_io_decls(void);
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

const char* get_input_path(void) {
    return qadav_get_input_path();
}

long long extract_yield_number(const char* path) {
    return qadav_extract_yield_number(path);
}

const char* format_trunc_line(long long val) {
    return qadav_format_trunc_line(val);
}


long long stmt_kind(long long idx) {
    return qadav_stmt_kind(idx);
}

const char* stmt_name(long long idx) {
    return qadav_stmt_name(idx);
}

long long stmt_value(long long idx) {
    return qadav_stmt_value(idx);
}

const char* stmt_expr(long long idx) {
    return qadav_stmt_expr(idx);
}

const char* format_alloca(const char* name) {
    return qadav_format_alloca(name);
}

const char* format_store(const char* name, long long val) {
    return qadav_format_store(name, val);
}

const char* format_load(const char* name, long long temp_id) {
    return qadav_format_load(name, temp_id);
}

const char* format_binop(const char* op, long long res_temp, long long t1, long long t2) {
    return qadav_format_binop(op, res_temp, t1, t2);
}

const char* format_emit_int_call(long long temp_id) {
    return qadav_format_emit_int_call(temp_id);
}

const char* format_trunc(long long temp_id, long long val) {
    return qadav_format_trunc(temp_id, val);
}

const char* format_ret(long long temp_id) {
    return qadav_format_ret(temp_id);
}


long long get_fib_value(long long idx) {
    return qadav_get_fib_value(idx);
}

const char* format_icmp_slt(long long t_cur, long long t_prev, long long val) {
    return qadav_format_icmp_slt(t_cur, t_prev, val);
}

const char* format_sub(long long t_cur, long long t_prev, long long val) {
    return qadav_format_sub(t_cur, t_prev, val);
}

const char* format_store_direct(long long val) {
    return qadav_format_store_direct(val);
}


const char* get_tensor_name(long long idx) {
    return qadav_get_tensor_name(idx);
}

long long get_tensor_size(long long idx) {
    return qadav_get_tensor_size(idx);
}

long long get_tensor_int_val(long long t_idx, long long e_idx) {
    return qadav_get_tensor_int_val(t_idx, e_idx);
}

const char* format_shape_const(long long shape_idx, long long size) {
    return qadav_format_shape_const(shape_idx, size);
}

const char* format_tensor_set(long long t_temp, long long elem_idx, long long int_val) {
    return qadav_format_tensor_set(t_temp, elem_idx, int_val);
}

long long tokenize_file(const char* path) { return qadav_tokenize_file(path); }
long long token_count(void) { return qadav_token_count(); }
const char* token_kind(long long idx) { return qadav_token_kind(idx); }
long long token_kind_code(long long idx) { return qadav_token_kind_code(idx); }
const char* token_literal(long long idx) { return qadav_token_literal(idx); }
long long token_int_val(long long idx) { return qadav_token_int_val(idx); }
long long token_is(long long idx, const char* s) { return qadav_token_is(idx, s); }
long long token_line(long long idx) { return qadav_token_line(idx); }
long long token_col(long long idx) { return qadav_token_col(idx); }

long long str_table_add(const char* str) { return qadav_str_table_add(str); }
long long str_table_count(void) { return qadav_str_table_count(); }
const char* str_table_get(long long idx) { return qadav_str_table_get(idx); }
const char* str_table_def(long long idx) { return qadav_str_table_def(idx); }
const char* str_table_gep(long long idx, long long temp_id) { return qadav_str_table_gep(idx, temp_id); }
const char* str_table_emit(long long temp_id) { return qadav_str_table_emit(temp_id); }
long long str_table_reset(void) { return qadav_str_table_reset(); }

const char* format_trunc_line_id(long long temp_id, long long val) {
    return qadav_format_trunc_line_id(temp_id, val);
}

const char* format_ret_line_id(long long temp_id) {
    return qadav_format_ret_line_id(temp_id);
}

long long register_string(const char* s) { return qadav_register_string(s); }
long long get_string_index(const char* s) { return qadav_get_string_index(s); }
long long get_string_byte_len(const char* s) { return qadav_get_string_byte_len(s); }
const char* escape_string(const char* s) { return qadav_escape_string(s); }
void set_last_int_value(long long v) { qadav_set_last_int_value(v); }
long long last_int_value(void) { return qadav_last_int_value(); }
void set_last_temp(long long t) { qadav_set_last_temp(t); }
long long get_last_temp(void) { return qadav_get_last_temp(); }
void set_last_emit_kind(long long k) { qadav_set_last_emit_kind(k); }
long long get_last_emit_kind(void) { return qadav_get_last_emit_kind(); }
const char* stmt_binop_op(long long idx) { return qadav_stmt_binop_op(idx); }
const char* stmt_binop_left(long long idx) { return qadav_stmt_binop_left(idx); }
const char* stmt_binop_right(long long idx) { return qadav_stmt_binop_right(idx); }
long long next_label_id(void) { return qadav_next_label_id(); }
void reset_label_counter(void) { qadav_reset_label_counter(); }
const char* format_icmp_imm(const char* pred, long long res_temp, long long left_temp, long long imm_val) {
    return qadav_format_icmp_imm(pred, res_temp, left_temp, imm_val);
}
const char* format_br_cond(long long cond_reg, long long then_id, long long else_id) {
    return qadav_format_br_cond(cond_reg, then_id, else_id);
}
const char* format_br_cond_merge(long long cond_reg, long long then_id, long long merge_id) {
    return qadav_format_br_cond_merge(cond_reg, then_id, merge_id);
}
const char* format_label(const char* prefix, long long id) {
    return qadav_format_label(prefix, id);
}
const char* format_br_label(const char* prefix, long long id) {
    return qadav_format_br_label(prefix, id);
}
const char* get_if_var_name(void) { return qadav_get_if_var_name(); }
long long get_if_var_val(void) { return qadav_get_if_var_val(); }
const char* get_if_cond_var(void) { return qadav_get_if_cond_var(); }
const char* get_if_cond_op(void) { return qadav_get_if_cond_op(); }
long long get_if_cond_val(void) { return qadav_get_if_cond_val(); }
long long get_if_yield_val(void) { return qadav_get_if_yield_val(); }
long long is_dot_proc(const char* name) { return qadav_is_dot_proc(name); }
long long is_main_proc(const char* name) { return qadav_is_main_proc(name); }
void set_proc_is_main(long long is_main) { qadav_set_proc_is_main(is_main); }
long long get_proc_is_main(void) { return qadav_get_proc_is_main(); }
void set_proc_is_dot(long long is_dot) { qadav_set_proc_is_dot(is_dot); }
long long get_proc_is_dot(void) { return qadav_get_proc_is_dot(); }
void set_is_terminated(long long term) { qadav_set_is_terminated(term); }
long long get_is_terminated(void) { return qadav_get_is_terminated(); }
const char* format_func_def_with_param(const char* name, const char* param_name) { return qadav_format_func_def_with_param(name, param_name); }
const char* format_param_alloca(const char* param_name) { return qadav_format_param_alloca(param_name); }
const char* format_param_store(const char* param_name) { return qadav_format_param_store(param_name); }
const char* format_call_i64(const char* callee, long long res_temp, long long arg_temp) { return qadav_format_call_i64(callee, res_temp, arg_temp); }
const char* format_ret_i64(long long temp_id) { return qadav_format_ret_i64(temp_id); }
const char* format_binop_imm(const char* op, long long res_temp, long long left_reg, long long imm_val) { return qadav_format_binop_imm(op, res_temp, left_reg, imm_val); }
const char* format_store_reg(const char* name, long long reg_id) { return qadav_format_store_reg(name, reg_id); }
double get_tensor_float_val(long long t_idx, long long e_idx) { return qadav_get_tensor_float_val(t_idx, e_idx); }
const char* format_tensor_create(long long reg, long long rank, long long size) { return qadav_format_tensor_create(reg, rank, size); }
const char* format_tensor_set_f64(long long t_reg, long long elem_idx, double val) { return qadav_format_tensor_set_f64(t_reg, elem_idx, val); }
const char* format_tensor_mul(long long dst_reg, long long a_reg, long long b_reg) { return qadav_format_tensor_mul(dst_reg, a_reg, b_reg); }
const char* format_tensor_reduce_sum_f64(long long dst_reg, long long src_reg) { return qadav_format_tensor_reduce_sum_f64(dst_reg, src_reg); }
long long get_tensor_count(void) { return qadav_get_tensor_count(); }
long long is_text_type(const char* typ) { return qadav_is_text_type(typ); }
long long is_tensor_type(const char* typ) { return qadav_is_tensor_type(typ); }
long long is_float_type(const char* typ) { return qadav_is_float_type(typ); }
long long is_ptr_var(const char* name) { return qadav_is_ptr_var(name); }
long long is_result_var(const char* name) { return qadav_is_result_var(name); }
long long is_mul_op(const char* op) { return qadav_is_mul_op(op); }
long long is_eq_op(const char* op) { return qadav_is_eq_op(op); }
long long is_var_y(const char* name) { return qadav_is_var_y(name); }
const char* format_shape_gep(long long shape_idx, long long temp_id) { return qadav_format_shape_gep(shape_idx, temp_id); }
const char* format_alloca_type(const char* name, const char* typ) { return qadav_format_alloca_type(name, typ); }
const char* format_store_reg_type(const char* name, const char* typ, long long reg_id) { return qadav_format_store_reg_type(name, typ, reg_id); }
const char* format_load_type(const char* name, const char* typ, long long temp_id) { return qadav_format_load_type(name, typ, temp_id); }
const char* format_call_double_2args(const char* callee, long long res_temp, long long arg1, long long arg2) { return qadav_format_call_double_2args(callee, res_temp, arg1, arg2); }
const char* format_func_def_dot(const char* name) { return qadav_format_func_def_dot(name); }
const char* format_ret_double(long long temp_id) { return qadav_format_ret_double(temp_id); }
const char* format_ret_i32(long long temp_id) { return qadav_format_ret_i32(temp_id); }
const char* format_emit_float_call(long long reg) { return qadav_format_emit_float_call(reg); }
const char* format_func_def_2params(const char* name, const char* p1, const char* p2) { return qadav_format_func_def_2params(name, p1, p2); }
const char* format_call_i64_imm(const char* callee, long long res_temp, long long arg1_temp, long long imm_val) { return qadav_format_call_i64_imm(callee, res_temp, arg1_temp, imm_val); }
const char* format_call_i64_2args(const char* callee, long long res_temp, long long arg1_temp, long long arg2_temp) { return qadav_format_call_i64_2args(callee, res_temp, arg1_temp, arg2_temp); }
const char* format_call_0args(const char* callee, long long res_temp) { return qadav_format_call_0args(callee, res_temp); }
long long register_var_type(const char* name, const char* typ) { return qadav_register_var_type(name, typ); }
long long get_var_type_code(const char* name) { return qadav_get_var_type_code(name); }
const char* get_tensor_val_str(long long elem_idx) { return qadav_get_tensor_val_str(elem_idx); }
void set_current_tensor_walk_idx(long long idx) { qadav_set_current_tensor_walk_idx(idx); }
const char* format_icmp_eq_ptr(long long res_temp, long long t1, long long t2) { return qadav_format_icmp_eq_ptr(res_temp, t1, t2); }
const char* format_func_def_text_param(const char* name, const char* param_name) { return qadav_format_func_def_text_param(name, param_name); }
const char* format_param_alloca_text(const char* param_name) { return qadav_format_param_alloca_text(param_name); }
const char* format_param_store_text(const char* param_name) { return qadav_format_param_store_text(param_name); }
const char* format_ret_text(long long reg_id) { return qadav_format_ret_text(reg_id); }
long long has_dynamic_io(void) { return qadav_has_dynamic_io(); }
long long dynamic_io_decl_count(void) { return qadav_dynamic_io_decl_count(); }
const char* dynamic_io_decl_get(long long idx) { return qadav_dynamic_io_decl_get(idx); }
long long emit_dynamic_io_decls(void) { return qadav_emit_dynamic_io_decls(); }
const char* format_proc_header(const char* name, const char* ret_type, const char* p1_name) { return qadav_format_proc_header(name, ret_type, p1_name); }
const char* format_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name) { return qadav_format_proc_header_2(name, ret_type, p1_name, p2_name); }
const char* format_br_cycle(long long cond_reg, long long body_id, long long end_id) { return qadav_format_br_cycle(cond_reg, body_id, end_id); }
const char* format_br_match(long long cmp_reg, long long arm_id, long long next_id) { return qadav_format_br_match(cmp_reg, arm_id, next_id); }
const char* format_icmp_eq_i32(long long res_reg, long long val_reg, long long tag) { return qadav_format_icmp_eq_i32(res_reg, val_reg, tag); }
const char* format_store_imm_i1(const char* name, long long val) { return qadav_format_store_imm_i1(name, val); }
const char* format_ret_type(const char* typ, long long reg) { return qadav_format_ret_type(typ, reg); }
const char* format_alloca_reg(long long reg) { return qadav_format_alloca_reg(reg); }
const char* format_store_reg_to_reg(long long val_reg, long long ptr_reg) { return qadav_format_store_reg_to_reg(val_reg, ptr_reg); }
const char* format_load_from_reg(long long res_reg, long long ptr_reg) { return qadav_format_load_from_reg(res_reg, ptr_reg); }
long long is_logic_type(const char* t) { return qadav_is_logic_type(t); }
long long get_record_field_idx(const char* rec_name, const char* field_name) { return qadav_get_record_field_idx(rec_name, field_name); }
long long is_record_field_ptr(const char* rec_name, const char* field_name) { return qadav_is_record_field_ptr(rec_name, field_name); }
long long get_variant_case_tag(const char* var_name, const char* case_name) { return qadav_get_variant_case_tag(var_name, case_name); }
const char* format_record_alloc(const char* rec_name, long long temp_id) { return qadav_format_record_alloc(rec_name, temp_id); }
const char* format_record_set_field_ptr(long long rec_reg, long long idx, long long val_reg) { return qadav_format_record_set_field_ptr(rec_reg, idx, val_reg); }
const char* format_record_set_field_i64(long long rec_reg, long long idx, long long val) { return qadav_format_record_set_field_i64(rec_reg, idx, val); }
const char* format_record_get_field_ptr(long long res_reg, long long obj_reg, long long idx) { return qadav_format_record_get_field_ptr(res_reg, obj_reg, idx); }
const char* format_record_get_field_i64(long long res_reg, long long obj_reg, long long idx) { return qadav_format_record_get_field_i64(res_reg, obj_reg, idx); }
const char* format_variant_create_i64(const char* var_name, const char* case_name, long long val, long long temp_id) { return qadav_format_variant_create_i64(var_name, case_name, val, temp_id); }
const char* format_variant_get_tag(long long res_reg, long long obj_reg) { return qadav_format_variant_get_tag(res_reg, obj_reg); }
const char* format_variant_get_field_i64(long long res_reg, long long obj_reg, long long idx) { return qadav_format_variant_get_field_i64(res_reg, obj_reg, idx); }

long long qadav_str_eq(const char* a, const char* b);
long long str_eq(const char* a, const char* b) { return qadav_str_eq(a, b); }
long long qadav_is_affirm_ident(const char* name);
long long is_affirm_ident(const char* name) { return qadav_is_affirm_ident(name); }
long long qadav_is_deny_ident(const char* name);
long long is_deny_ident(const char* name) { return qadav_is_deny_ident(name); }
long long qadav_is_text_returning_proc(const char* name);
long long is_text_returning_proc(const char* name) { return qadav_is_text_returning_proc(name); }
long long qadav_is_logic_returning_proc(const char* name);
long long is_logic_returning_proc(const char* name) { return qadav_is_logic_returning_proc(name); }

void* qadav_alloc(unsigned long long size) {
    return HeapAlloc(GetProcessHeap(), 0x00000008 /* HEAP_ZERO_MEMORY */, size);
}

void qadav_free(void* p) {
    if (p) {
        HeapFree(GetProcessHeap(), 0, p);
    }
}

static HANDLE out_handle(void) {
    static HANDLE h = NULL;
    if (!h) {
        h = GetStdHandle(STD_OUTPUT_HANDLE);
    }
    return h;
}

void qadav_write_raw(const char* s, unsigned long long len) {
    DWORD written;
    WriteFile(out_handle(), s, (DWORD)len, &written, NULL);
}

static void write_raw(const char* s, DWORD len) {
    qadav_write_raw(s, len);
}

static int int_to_str(long long v, char* buf) {
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

void qadav_emit_int(long long v) {
    char buf[32];
    int n = int_to_str(v, buf);
    write_raw(buf, n);
    write_raw("\n", 1);
}

void qadav_emit_float(double v) {
    char buf[64];
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
    idx += int_to_str(ipart, buf + idx);
    buf[idx++] = '.';
    // 6-digit fractional zero-padding
    char fbuf[16];
    int fn = int_to_str(fpart, fbuf);
    for (int p = 0; p < 6 - fn; p++) {
        buf[idx++] = '0';
    }
    for (int p = 0; p < fn; p++) {
        buf[idx++] = fbuf[p];
    }
    write_raw(buf, idx);
    write_raw("\n", 1);
}

void qadav_emit_text(const char* str) {
    if (!str) return;
    DWORD len = 0;
    while (str[len]) len++;
    write_raw(str, len);
    write_raw("\n", 1);
}

void qadav_emit_logic(int v) {
    if (v) {
        write_raw("affirm\n", 7);
    } else {
        write_raw("deny\n", 5);
    }
}

void qadav_emit_qubit(const void* state) {
    (void)state;
    write_raw("|qubit>\n", 8);
}

void qadav_exit(int code) {
    ExitProcess(code);
}

// Forward declare user program main and self-check
int main(void);
int qadav_selfcheck_run(void);

// Custom entry point for Windows
void qadav_start(void) {
    qadav_selfcheck_run();
    int ret = main();
    ExitProcess(ret);
}

// Freestanding memory primitives (zero-libc runtime)
void* memcpy(void* dest, const void* src, unsigned long long n) {
    char* d = (char*)dest;
    const char* s = (const char*)src;
    for (unsigned long long i = 0; i < n; i++) {
        d[i] = s[i];
    }
    return dest;
}

void* memset(void* s, int c, unsigned long long n) {
    unsigned char* p = (unsigned char*)s;
    for (unsigned long long i = 0; i < n; i++) {
        p[i] = (unsigned char)c;
    }
    return s;
}

// Windows stack probe stubs (zero-libc runtime)
void __chkstk(void) {}
void _chkstk(void) {}


#endif

extern long long qadav_get_next_var_init_arg_count(void);
long long get_next_var_init_arg_count(void) { return qadav_get_next_var_init_arg_count(); }
extern long long qadav_emit_record_init(const char* tn, long long fc, long long t);
extern long long qadav_emit_variant_init(const char* vn, const char* cn, long long pr, long long pk, long long t);
extern long long qadav_emit_match_stmt(const char* s, long long ac, long long t);
long long emit_record_init(const char* tn, long long fc, long long t) { return qadav_emit_record_init(tn, fc, t); }
long long emit_variant_init(const char* vn, const char* cn, long long pr, long long pk, long long t) { return qadav_emit_variant_init(vn, cn, pr, pk, t); }
long long emit_match_stmt(const char* s, long long ac, long long t) { return qadav_emit_match_stmt(s, ac, t); }
extern long long qadav_emit_match_prelude(long long sr, long long ac, long long ti);
long long emit_match_prelude(long long sr, long long ac, long long ti) { return qadav_emit_match_prelude(sr, ac, ti); }
extern long long qadav_emit_match_check(long long ai, long long ac, long long tr, long long ti);
long long emit_match_check(long long ai, long long ac, long long tr, long long ti) { return qadav_emit_match_check(ai, ac, tr, ti); }
extern void qadav_emit_match_body_label(long long ai);
void emit_match_body_label(long long ai) { qadav_emit_match_body_label(ai); }
extern void qadav_emit_match_body_end(void);
void emit_match_body_end(void) { qadav_emit_match_body_end(); }
extern void qadav_emit_match_end(void);
void emit_match_end(void) { qadav_emit_match_end(); }

extern long long qadav_record_imm_int(long long val);
long long record_imm_int(long long val) { return qadav_record_imm_int(val); }
extern const char* qadav_format_trunc_line_reg(long long temp_id, long long reg);
const char* format_trunc_line_reg(long long temp_id, long long reg) { return qadav_format_trunc_line_reg(temp_id, reg); }

extern long long qadav_emit_proc_header(const char* name, const char* ret_type, const char* p1_name);
extern long long qadav_emit_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name);
extern long long qadav_emit_record_get_field(long long res_temp, long long obj_temp, long long idx, long long is_ptr);
long long emit_proc_header(const char* name, const char* ret_type, const char* p1_name) { return qadav_emit_proc_header(name, ret_type, p1_name); }
long long emit_proc_header_2(const char* name, const char* ret_type, const char* p1_name, const char* p2_name) { return qadav_emit_proc_header_2(name, ret_type, p1_name, p2_name); }
long long emit_record_get_field(long long r, long long o, long long i, long long p) { return qadav_emit_record_get_field(r, o, i, p); }

extern long long qadav_emit_call_3(const char* callee, long long res_temp, long long r1, long long r2, long long r3);
extern long long qadav_emit_call_4(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4);
extern long long qadav_emit_call_5(const char* callee, long long res_temp, long long r1, long long r2, long long r3, long long r4, long long r5);
long long emit_call_3(const char* c, long long res, long long r1, long long r2, long long r3) { return qadav_emit_call_3(c, res, r1, r2, r3); }
long long emit_call_4(const char* c, long long res, long long r1, long long r2, long long r3, long long r4) { return qadav_emit_call_4(c, res, r1, r2, r3, r4); }
long long emit_call_5(const char* c, long long res, long long r1, long long r2, long long r3, long long r4, long long r5) { return qadav_emit_call_5(c, res, r1, r2, r3, r4, r5); }
