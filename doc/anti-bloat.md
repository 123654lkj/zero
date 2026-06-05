# Zero Anti-Bloat Manifesto

**Zero is not a platform. Zero is a compiler and a VM. Nothing else.**

Every existing language suffers from the same disease: they grow. They add features, libraries, abstractions, runtimes, package managers, build tools, and configuration files until the original language is a tiny speck inside a mountain of scaffolding.

Zero does the opposite. Zero starts from zero and stays at zero.

---

## The Bloat Audit

### Where bloat comes from in existing languages

| Language | Bloat Source | Size | What Zero Does |
|----------|-------------|------|----------------|
| Python | Batteries-included stdlib | 1000+ modules | Ship nothing. Import on demand. |
| Java | JVM + classpath + build tools | 300MB+ | 10MB VM, no classpath |
| Go | Compiled-in runtime + net/http | 30MB binary | Strippable, modular runtime |
| Rust | Monomorphization + LLVM | 100MB compiler | Single-pass, no heavy opts |
| Node | node_modules | Infinite | No dependency system |
| C++ | Header parsing + templates | Slow compile | No headers, no templates |

### Zero's rule

**If it's not necessary for the language to speak to itself, it doesn't ship.**

---

## Concrete Zero Decisions

### 1. Compiler: Single Pass

Zero's Phase 0 compiler is a single forward pass over the AST. No optimization passes, no IR transformation, no inlining, no dead code elimination.

`
.zero  →  Lexer  →  Parser  →  Compiler  →  .zbc
         1 pass    1 pass    1 pass
`

Target: **< 5000 lines of Go, compiles in < 1 second.**

Compare:
- Go gc compiler: 500K+ lines
- Rustc: 1M+ lines
- V8 TurboFan: 300K+ lines

Compiler bloat is not sophistication. It is debt.

### 2. VM: Bare Minimum

No JIT. No tiered compilation. No garbage collector in Phase 0 (Go GC handles it). No profiling infrastructure.

The VM main loop is a single switch(opcode) dispatch:

`
loop:
    op = chunk.code[ip++]
    switch op:
        case ADD:  b = stack[sp--]; a = stack[sp--]; stack[++sp] = a + b
        case CALL: // push frame
        case RET:  // pop frame
        goto loop
`

Target: **< 1000 lines of Go, < 10KB binary footprint for VM logic.**

Compare:
- CPython ceval.c: 5000+ lines
- JVM interpreter: 50K+ lines

### 3. Value System: Zero Heap Allocation for Common Cases

Every Zero value is 8 bytes:
- 1 byte: type tag
- 7 bytes: payload

Small integers (< 2^31) are embedded directly in the payload. Short strings (< 7 bytes) are embedded directly. No heap allocation for the most common cases.

Result: **80%+ of value operations require zero memory allocation.**

Compare:
- Python: every int is a PyObject* (16+ bytes heap + pointer)
- Java: Integer vs int boxing/unboxing confusion
- JavaScript: every number is a heap object in older engines

### 4. Standard Library: Minimum Viable

Zero ships with exactly what a compiler needs to compile itself:

`
std/
├── io.zero        — read/write files, stdin/stdout
├── math.zero      — basic math operations
├── collections.zero — arrays, maps
└── format.zero    — string formatting
`

Everything else (http, browser, ai, crypto, json) is external. Not installed by default. Loaded explicitly when needed.

Target: **< 1000 lines of Zero code for the entire standard library.**

Compare:
- Python stdlib: hundreds of thousands of lines
- Go stdlib: megabytes of code

### 5. Build System: None

`
zero run file.zero       → compile + run
zero build file.zero     → compile to .zbc
zero repl                → interactive mode
`

That's it. There is no build system. There is no package manager. There is no dependency resolution. There are no configuration files.

Zero source files are self-contained. If you need something from another file:

`zero
import 
another.zero
`

The compiler resolves paths, not packages. No registries, no versions, no lock files.

Compare:
- Go: go.mod, go.sum, GOPATH, module proxies
- Rust: Cargo.toml, Cargo.lock, crates.io
- Node: package.json, node_modules (average 300+ packages for hello world)

### 6. Memory: Contiguous and Preallocated

The VM's stack, call frames, and constant pools are all preallocated contiguous arrays. No per-operation allocation.

`
VM {
    stack:  [4096]Value    // preallocated
    frames: [256]Frame     // preallocated
    globals: []Value       // grown only when new globals are defined
}
`

No malloc in the hot path. No hash table lookups in the hot path. No runtime type checks in the hot path (the opcode tells you the operation).

### 7. Compilation Speed: Instant

Phase 0 target for compiling a typical .zero file:

`
File size    Compile time
  100 lines    < 1ms
 1000 lines    < 5ms
10000 lines    < 50ms
`

This is achieved by:
- No IR construction
- No optimization passes
- No type checking in Phase 0 (deferred to Phase 1+)
- Bounded memory allocation (arena for AST, freed in one shot)

Compare:
- Rust: seconds to minutes
- Go: 100ms to seconds
- C++: seconds to hours

---

## The Test: Zero's Binary Size

The final Phase 2 Zero binary should be:

| Component | Target Size |
|-----------|------------|
| Compiler  | < 100KB |
| VM        | < 50KB |
| Runtime   | < 50KB |
| Stdlib    | < 20KB |
| Total (single binary) | **< 250KB** |

Compare:
- Go hello world: ~2MB
- Rust hello world: ~400KB (stripped)
- Python interpreter: ~15MB
- Node binary: ~40MB

---

## The Rule

Every line of code in Zero must justify its existence against this question:

**Does this help the language speak to itself?**

If yes, it belongs. If no, it does not ship. Not later. Not in a feature release. Never.

What Zero is not:
- Not an operating system
- Not a platform
- Not an ecosystem
- Not an IDE
- Not a build tool
- Not a package manager
- Not a container runtime

What Zero is:
- A compiler that reads .zero files and produces .zbc bytecode
- A VM that executes .zbc bytecode
- That is all.

---

## Summary

| Metric | Existing Languages | Zero Target |
|--------|-------------------|-------------|
| Compiler size | 500K - 1M+ lines | < 5000 lines |
| VM size | 5K - 50K+ lines | < 1000 lines |
| Stdlib | 1000+ modules | < 10 modules |
| Binary size | 2MB - 40MB | < 250KB |
| Compile time (1K lines) | 100ms - minutes | < 5ms |
| Build system | Global complexity | None |
| Heap allocations per op | Yes | No (small values inlined) |
| Dependencies | Hundreds | Zero |

---

**Zero starts at zero and stays at zero.**