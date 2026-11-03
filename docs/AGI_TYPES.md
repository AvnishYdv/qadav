# AGI Native Types Specification

**Language:** Qadav Native Language  
**Milestone:** Combined M3.1 + M3.2  
**Runtime:** Freestanding Zero-Libc Native Runtime  

---

## 1. Overview

Qadav introduces first-class native language types designed specifically for Artificial General Intelligence (AGI) cognitive workflows:

1. **`memory`**: Episodic memory buffer for persistent association and key-value recall.
2. **`graph`**: Semantic association graph for relational concept networks and BFS topological exploration.
3. **`procedure`**: First-class executable cognitive token representing reusable computational routines.

---

## 2. Episodic Memory (`memory`)

The `memory` type provides an associative memory buffer utilizing FNV-1a hashing with linked list bucket chaining.

### Syntax
```ydv
// Declaration with capacity
memory mem ⟬ 100 ⟭ ·

// Store an episodic entry
remember mem⟬"agent_state"⟭ ↦ 42 ·

// Retrieve an episodic entry
int64 val ↦ recall mem⟬"agent_state"⟭ ·
```

### Runtime Details (`qadav_agi.h` / `qadav_agi.c`)
- Hash function: 64-bit FNV-1a offset basis `14695981039346656037` with prime multiplier `1099511628211`.
- Chained bucket resolution allocated via Win32 `HeapAlloc` / Linux `sys_brk`.
- Output inspection format: `memory⟨count=N, capacity=M⟩`.

---

## 3. Semantic Association Graph (`graph`)

The `graph` type represents non-Euclidean knowledge graphs, ontology structures, and conceptual associations.

### Syntax
```ydv
// Graph declaration
graph g ·

// Node declaration with payload
node g ⟬ "conceptA" ↦ 101 ⟭ ·
node g ⟬ "conceptB" ↦ 102 ⟭ ·

// Weighted relational link
link g ⟬ "conceptA" ⸴ "conceptB" ⸴ 1.0 ⟭ ·

// Breadth-First Search trace returning tensor⟨int64⟩ of visited node IDs
tensor⟨int64⟩ visited ↦ trace g ⟬ "conceptA" ⸴ 2 ⟭ ·
```

### Runtime Details
- Node dictionary storing unique identifiers and `int64` concept payloads.
- Adjacency list storing outgoing directed edges with double-precision weights.
- Breadth-First Search traversal executing within native memory buffers without recursion or external graph libraries.

---

## 4. Executable Procedure (`procedure`)

The `procedure` type encapsulates reusable computational and cognitive routines into first-class callable tokens.

### Syntax
```ydv
// Procedure declaration
procedure add_two ⦅ int64 x ⦆ → int64 ⟦
    yield x ⊕ 2 ·
⟧

// Invocation via execute
process main ⦅ ⦆ → int64 ⟦
    int64 res ↦ execute add_two ⦅ 5 ⦆ ·
    emit res ·
    yield 0 ·
⟧
```

### Runtime Details
- Compiled to isolated LLVM functions (`@proc_<name>`) with direct calling conventions.
- Zero-libc procedure dispatching avoiding heap allocation when invoked within local process scopes.

