# Zero Language Philosophy

## Zero doesn't add features. Zero removes problems.

Every programming language ever created solved some problems and introduced new ones. Zero is a systematic audit of 60 years of programming language design — absorb every lesson, fix every mistake.

---

## The Audit

### C
- **Learned:** Minimal runtime, no hidden costs, honest about hardware
- **Fixed:** Manual memory, pointer arithmetic, no module system

### C++
- **Learned:** Zero-cost abstraction, RAII
- **Fixed:** C compatibility debt, template hell, complexity explosion

### Java
- **Learned:** Bytecode VM cross-platform, GC, massive ecosystem
- **Fixed:** Verbosity, null pointers, checked exceptions

### Python
- **Learned:** Readability is a feature, batteries included, REPL
- **Fixed:** GIL, runtime type errors, slow

### JavaScript
- **Learned:** Event loop, first-class functions
- **Fixed:** Legacy design mistakes, callback hell

### Go
- **Learned:** Minimal spec, second-compilation, goroutines, single binary
- **Fixed:** nil, `if err != nil` ceremony, interface{} erasure

### Rust
- **Learned:** Algebraic types, pattern matching, traits, no null
- **Fixed:** Borrow checker complexity, compile speed

### Haskell
- **Learned:** Pure functions + explicit effects, type classes
- **Fixed:** Lazy-by-default (space leaks), Monad narrative

### Erlang
- **Learned:** Actor model, let-it-crash, hot reload
- **Fixed:** Syntax, CPU-bound performance

### Smalltalk
- **Learned:** Live programming, image persistence
- **Fixed:** Closed ecosystem

### Lisp
- **Learned:** Code is data, macros
- **Fixed:** Parenthesis tsunami

### SQL
- **Learned:** Declarative > imperative
- **Fixed:** ORM impedance mismatch

### Lua
- **Learned:** Table as universal structure
- **Fixed:** 1-indexing, no types

### TypeScript
- **Learned:** Gradual typing, type inference
- **Fixed:** Type gymnastics

### Prolog
- **Learned:** Pattern matching as execution model
- **Fixed:** No general-purpose usability

---

## The Seven Problems

1. **Code Rot** — Dependencies drift, APIs change. Zero makes version constraints language-level.
2. **Environment Dependency** — Bytecode binds to environment hash. Changed environment = refused or auto-adapted.
3. **AI Writes But Cannot Run** — `ai fn` is language-level. Compiler understands cost, latency, caching.
4. **Glue Code** — 80% of real work. Zero's `|>` is a compiler-optimized data pipeline.
5. **Failure Is Normal** — Blocked, timed out, malformed. Zero treats failure as first-class.
6. **Runtime Black Box** — Snapshot, restore, replay. Built into the language.
7. **Cognitive Tax** — Zero config. One file, one command.