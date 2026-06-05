# Zero Anti-Bloat Manifesto

**Zero is not a platform. Zero is a compiler and a VM. Nothing else.**

Every existing language suffers from the same disease: they grow. They add features, libraries, abstractions, runtimes, package managers, build tools, and configuration files until the original language is a tiny speck inside a mountain of scaffolding.

Zero does the opposite. Zero starts from zero and stays at zero.

---

## The Bloat Audit

| Language | Bloat Source | Size | What Zero Does |
|----------|-------------|------|----------------|
| Python | Batteries-included stdlib | 1000+ modules | Ship nothing. Import on demand. |
| Java | JVM + classpath + build tools | 300MB+ | 10MB VM, no classpath |
| Go | Compiled-in runtime + net/http | 30MB binary | Strippable, modular runtime |
| Rust | Monomorphization + LLVM | 100MB compiler | Single-pass, no heavy opts |
| Node | node_modules | Infinite | No dependency system |
| C++ | Header parsing + templates | Slow compile | No headers, no templates |

**Zero's rule: If it's not necessary for the language to speak to itself, it doesn't ship.**

---

## Concrete Decisions

### 1. Compiler: Single Pass

No optimization passes. No IR transformation. No inlining. No dead code elimination.

\\\
.zero  →  Lexer (1 pass)  →  Parser (1 pass)  →  Compiler (1 pass)  →  .zbc
\\\

Target: **< 5000 lines of Go, < 1 second compile.**

### 2. VM: One Switch Loop

No JIT. No tiered compilation. No GC in Phase 0. Profiling: zero.

\\\
loop:
    op = code[ip++]
    switch op:
        case ADD:  b = pop(); a = pop(); push(a + b)
        case CALL: push_frame()
        case RET:  pop_frame()
    goto loop
\\\

Target: **< 1000 lines of Go.**

### 3. Value: Zero Heap for Common Cases

Every Value is 8 bytes. Small ints (< 2^31) and short strings (< 7 bytes) inline. No allocation.

80%+ of operations require **zero memory allocation**.

### 4. Stdlib: Minimum Viable

\\\
std/
├── io.zero
├── math.zero
├── collections.zero
└── format.zero
\\\

Everything else (http, browser, ai, crypto, json) external. Loaded explicitly.

Target: **< 1000 lines of Zero.**

### 5. Build System: None

\\\
zero run file.zero       ← compile + run
zero build file.zero     ← compile to .zbc
zero repl                ← interactive
\\\

No go.mod. No Cargo.toml. No package.json. No node_modules. No lock files.

### 6. Memory: Preallocated

\\\
VM {
    stack:  [4096]Value
    frames: [256]Frame
}
\\\

No malloc in hot path. No hash table lookups in hot path.

### 7. Compile Speed: Instant

\\\
100 lines:   < 1ms
1000 lines:  < 5ms
10000 lines: < 50ms
\\\

---

## Binary Size Target

| Component | Target |
|-----------|--------|
| Compiler  | < 100KB |
| VM        | < 50KB |
| Runtime   | < 50KB |
| Stdlib    | < 20KB |

**Total Phase 2 binary: < 250KB.**

---

## The Rule

Every line of code in Zero must justify its existence against this question:

**Does this help the language speak to itself?**

If yes: include it.
If no: **it does not ship. Not later. Not in a feature release. Never.**

---

## Summary

| Metric | Existing | Zero |
|--------|----------|------|
| Compiler lines | 500K - 1M | < 5000 |
| VM lines | 5K - 50K | < 1000 |
| Stdlib modules | 1000+ | < 10 |
| Binary size | 2MB - 40MB | < 250KB |
| Compile 1K lines | 100ms - minutes | < 5ms |
| Build system | Complex | None |
| Heap alloc per op | Common | None (inlined) |
| Dependencies | Hundreds | Zero |

---

**Zero starts at zero and stays at zero.**