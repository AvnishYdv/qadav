# Qadav Dataset Specification (QDS Binary Format v1)

## 1. Specification Overview

The Qadav Dataset (QDS) binary container is a high-performance, zero-libc, standalone format engineered for streaming tensor payloads into Qadav neural network and machine learning workflows without external parser dependencies.

## 2. Global Header Specification (32 Bytes Exact)

All integers are stored in Little-Endian byte order.

| Offset | Field | Type | Size | Description |
|---|---|---|---|---|
| `0x00` | `magic` | `char[4]` | 4 bytes | ASCII identifier: `"QDS1"` (`0x51`, `0x44`, `0x53`, `0x31`) |
| `0x04` | `n_tensors` | `uint32` | 4 bytes | Number of encapsulated tensors ($N \le 100$) |
| `0x08` | `total_bytes` | `uint64` | 8 bytes | Total payload size in bytes ($\le 1,073,741,824$ bytes / 1 GB) |
| `0x10` | `reserved` | `uint8[16]` | 16 bytes | Reserved padding bytes (all zeros `0x00`) |

## 3. Tensor Record Specification

Following the 32-byte header, exactly `n_tensors` records are sequentially stored:

| Field | Type | Size | Description |
|---|---|---|---|
| `name_len` | `uint32` | 4 bytes | Byte length of tensor identifier ($L \le 256$) |
| `name` | `char[name_len]` | $L$ bytes | UTF-8/ASCII identifier (e.g., `"features"`, `"labels"`) |
| `rank` | `uint32` | 4 bytes | Tensor dimension rank ($0 \le R \le 8$) |
| `shape` | `uint32[rank]` | $4 \times R$ bytes | Extents per axis |
| `elem_kind` | `uint32` | 4 bytes | `0`: `int64`, `1`: `float64` |
| `data_len` | `uint64` | 8 bytes | Raw payload bytes ($B \le 524,288,000$ bytes / 500 MB) |
| `payload` | `byte[data_len]` | $B$ bytes | Little-endian numeric values |

## 4. Dataset Bomb Protections

The zero-libc runtime enforces hard defensive bounds upon file load:
- Maximum total file payload: 1 GB (`total_bytes <= 1073741824`)
- Maximum encapsulated tensors: 100 (`n_tensors <= 100`)
- Maximum single tensor buffer: 500 MB (`data_len <= 524288000`)
- Maximum identifier length: 256 characters (`name_len <= 256`)
- Path traversal rejection: File paths containing `..` or leading `/` or illegal null bytes are rejected prior to calling OS open APIs.

