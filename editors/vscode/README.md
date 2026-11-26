# Qadav Language Support for VS Code & Open VSX

Official Visual Studio Code, VSCodium, and Open VSX extension for the **Qadav (`.ydv`)** quantum-native, zero-libc programming language.

## Features

- **Rich Syntax Highlighting:**
  - Full keyword, delimiter, type, and identifier highlighting.
  - Native Unicode operator support: `↦` (bind), `←` (assign), `⊕` (add), `⊖` (sub), `⊗` (mul), `⊘` (div), `≡` (eq), `≠` (neq), `∧` (and), `∨` (or), `·` (statement terminator).
  - Native brackets: `⟦ ⟧` (blocks), `⦅ ⦆` (parameter lists), `⟬ ⟭` (tensor dimensions).
  - Native AVX2 SIMD intrinsics: `@vec256`, `vec4f64`, `vec8f32`, `vec4_fma`, `vec4_hsum`, etc.
  - Quantum operators and Dirac/stabilizer notation.
- **Snippets Suite:**
  - Production snippets for `proc`, `proc_main`, `vec256`, `when`, `cycle`, `match`, `record`, `variant`, `stab`, `qera`, `barat`, and `pkg`.
- **Integrated Tooling Commands:**
  - `Qadav: Build Current File`: Compiles active `.ydv` to standalone PE32+ executable using zero external tools.
  - `Qadav: Run Current File`: Builds and executes active file in output channel.
  - `Qadav: Typecheck File`: Fast AST validation and static type checking.
  - `Qadav: Run Doctor Diagnostics`: Runs compiler and system health checks.
- **Native LSP Integration:**
  - Built-in stdio client connecting directly to `bin\qyn.exe lsp`.

## Extension Settings

This extension contributes the following settings:

* `qadav.compilerPath`: Absolute or workspace-relative path to `bin/qyn.exe`.
* `qadav.packageManagerPath`: Path to `bin/qpk.exe`.
* `qadav.enableLsp`: Enable native Language Server Protocol (LSP).

## License

Apache-2.0
