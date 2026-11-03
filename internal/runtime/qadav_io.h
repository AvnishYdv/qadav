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

#endif

