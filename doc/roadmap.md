# Zero Development Roadmap

## Phase 0: Bootstrap (Current — estimated 4-6 weeks)

Goal: A working Zero implementation in Go that can compile and run Zero programs.

### Week 1-2: Foundation
- [ ] Value system (go/value/value.go)
- [ ] Opcode definitions (go/opcode/opcode.go)
- [ ] Chunk container (go/chunk/chunk.go)
- [ ] Basic VM (go/vm/vm.go)
- [ ] Built-in functions (go/vm/builtins.go)
- [ ] First test: `print("hello")` works

### Week 3-4: Compiler
- [ ] Lexer (go/compiler/lexer.go) — 27 token types
- [ ] Parser + AST (go/compiler/parser.go, ast.go)
- [ ] Compiler (go/compiler/compiler.go) — expressions, variables, functions
- [ ] REPL (go/repl/repl.go)
- [ ] Tests pass: arithmetic, if, loops, functions, recursion, arrays

### Week 5-6: Polish
- [ ] Error messages with source location
- [ ] Type checking pass
- [ ] Closure support
- [ ] Module system
- [ ] Standard library foundation (io, json, http)
- [ ] All Phase 0 tests green

## Phase 1: Self-Hosting (6-8 weeks)

Goal: Zero compiler rewritten in Zero, bootstrapped via Phase 0.

- [ ] Write lexer.zero
- [ ] Write parser.zero
- [ ] Write codegen.zero
- [ ] Bootstrap: Phase 0 compiles Phase 1
- [ ] Bytecode verification between phases
- [ ] Self-hosting test: compiler compiles itself

## Phase 2: Unique Features (8-12 weeks)

Goal: Zero's unique language-level features.

- [ ] ai fn construct and runtime
- [ ] Effect tracking system
- [ ] Image snapshots (serialize/restore VM state)
- [ ] Actor model (spawn/send/receive)
- [ ] Data pipeline optimizer
- [ ] Pattern matching engine
- [ ] Error-as-value with retry strategies

## Phase 3: Browser/AI Runtime (ongoing)

Goal: Zero powers real-world AI agents.

- [ ] std/browser — Playwright/rod integration via FFI
- [ ] std/ai — LLM integration via EchoBird
- [ ] std/web — HTTP with uTLS fingerprinting
- [ ] Guard system — captcha detection, cooling strategies
- [ ] Adapter system — JD, WeChat, Xiaohongshu, etc.