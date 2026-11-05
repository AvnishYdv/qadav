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

#endif

