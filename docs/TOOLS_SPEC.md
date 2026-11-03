# Safe Tool Orchestration Specification
## Milestone M5.8 (Qadav Phase 5)

---

### 1. Abstract & Executive Summary

Autonomous AI agents requiring native execution capabilities present severe safety risks if allowed unrestricted or unmonitored computer access. **Safe Tool Orchestration (M5.8)** establishes a cryptographically governed, permission-bounded execution harness integrated directly into Qadav's Rust/C runtime and LLVM compiler.

Every tool registered in Qadav is bound to a Quantum-Entangled Ethical Hamiltonian (QEEH) boundary, strict pre-conditions, and post-conditions. Crucially, tool invocations are natively attested into Quantum-Entangled Reasoning Attestation (QERA — Invention #11) chains, providing mathematically verifiable, tamper-evident audit trails of every tool interaction.

---

### 2. Architectural Design

1. **Mandatory Ethical Gating:** Prior to executing any tool action, the runtime evaluates Hamiltonian penalty bounds against the tool identifier and invocation context.
2. **Deterministic Pre- and Post-Conditions:** Inputs and outputs are strictly validated against invariant constraints before and after execution.
3. **Automatic QERA Reasoning Step Logging:** When an audit chain is provided to `tool_call`, the runtime records a `QADAV_OP_VERIFY` reasoning step:
   $$\text{step\_commitment} = \text{SHA3-256}(\text{step\_id} \parallel \text{rdtsc} \parallel \text{tool\_name} \parallel \text{result\_hash} \parallel \text{boundary\_hash} \parallel \text{confidence} = 1.0)$$
4. **Active Decoherence on Unauthorized Invocations:** If an unauthorized tool execution or boundary breach is attempted, the runtime triggers active decoherence and aborts execution immediately.

---

### 3. Syntax & Grammar

In Qadav source files (`.ydv`):

#### 3.1 Tool Registration Declaration
```ydv
tool_register file_scanner with tool_auth pre ⟦ 1 ⟧ post ⟦ 1 ⟧ ·
```

#### 3.2 Tool Call Statement with QERA Audit
```ydv
tool_call file_scanner⦅"/var/log/kernel.log"⦆ audit audit_ledger ·
```

#### 3.3 Dynamic Audit Verification
```ydv
verify_chain audit_ledger ·
```

---

### 4. Zero-Libc C Runtime Interface

Defined in `internal/runtime/qadav_tools.h` and `internal/runtime/qadav_tools.c`:

```c
typedef struct qadav_tool_entry {
    char          name[64];
    unsigned char boundary_hash[32];
    void*         boundary;
    int           has_pre;
    int           has_post;
} qadav_tool_entry;

typedef struct qadav_tool_registry {
    int              count;
    qadav_tool_entry tools[QADAV_MAX_TOOLS];
} qadav_tool_registry;

int qadav_tool_register(const char* name, void* boundary, int has_pre, int has_post);
int qadav_tool_call(const char* name, void* caller_boundary, const char* arg, void* audit_chain);
void qadav_tools_reset(void);
```

---

### 5. Verification & Test Evidence

Safe Tool Orchestration is verified by:
- Runtime Test: `TestToolRegisterAndCall` in `internal/runtime/crypto_test.go`
- Static Type Checker: `internal/checker/tools_checker_test.go`
- End-to-End Execution: `examples/tool_demo.ydv` (Passes and prints `Tool orchestration with QERA attestation verified`)

---

### 6. Post-Condition Failure Handling (Fix F20)

In M5.9/M5.10 hardening, tool execution recovery semantics were formalized:
- **Graceful Failure (`qadav_tool_call` returns `-4`)**: When a tool post-condition check fails (`has_post == 2` or `"fail_post"` argument), the runtime detects invalid post-state and returns `-4`.
- **Audit Step Omission**: When post-conditions fail, the step is explicitly **not** appended to the QERA reasoning chain. This guarantees that unverified or corrupted tool outputs never acquire cryptographic attestation.
- **No Quantum Decoherence**: Unlike intentional cryptographic tampering or boundary violations, a failed tool post-condition does not trigger quantum decoherence or abort the process. This allows upstream cognitive orchestration to initiate calibrated retries, alternate tool selection, or principled abstention. Verified via `TestToolPostConditionFailure`.

