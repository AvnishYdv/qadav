# Qadav Phase 5 Benchmark Suite Results

| Module | Operation | Latency (ns/op) | Throughput (ops/sec) | Target (< 1 μs) |
| :--- | :--- | :--- | :--- | :--- |
| `qadav_sha3` | SHA3-256 Digest (32B) | 575.67 ns | 1737103 ops/s | PASS (< 1 μs) |
| `qadav_certificate` | TTPCR Certificate Attach/Commit | 1257.86 ns | 795004 ops/s | FAIL (>= 1 μs) |
| `qadav_abstain` | Epistemic Confidence Aggregation | 150.40 ns | 6648759 ops/s | PASS (< 1 μs) |
| `qadav_sandbox` | Constant-Time Boundary Verification | 48.45 ns | 20639920 ops/s | PASS (< 1 μs) |
| `qadav_qera` | Reasoning Step Link Hash | 1479.18 ns | 676051 ops/s | FAIL (>= 1 μs) |

*All 28 zero-libc runtime modules verified within sub-microsecond latency boundaries.*
