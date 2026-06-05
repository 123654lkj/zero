# Zero Architecture

## Pipeline

\\\
.zero source  →  Lexer (tokens)  →  Parser (AST)  →  Compiler (.zbc)  →  VM (result)
\\\

## Bootstrap Compiler (Go)

### go/value/ — Value System

Every Zero value is an 8-byte struct:
- 1 byte: type tag (Nil, Bool, Int, Float, String, Array, Map, Closure, Native, Pattern, Tagged, Stream, Image, IO, Table)
- 7 bytes: payload (embedded for small values, pointer for heap values)

Small ints (<2^31) and short strings (<7 bytes) are embedded directly. Zero heap allocation for common cases.

### go/opcode/ — Instruction Set

39 instructions in groups with reserved ranges for future expansion.

Reserved ranges:
\\\
0x00–0x1F  Stack operations (PUSH, POP, DUP)
0x20–0x3F  Variables (LOAD, STORE)
0x40–0x5F  Arithmetic/Logic
0x60–0x7F  Jumps
0x80–0x9F  Functions
0xA0–0xBF  Data structures
0xC0–0xDF  Concurrency (SPAWN, SEND, RECEIVE)
0xE0–0xFF  Zero features (IO, IMAGE, QUERY, AI)
\\\

Phase 0 implements 0x00–0x9F (24 instructions).

### go/chunk/ — Bytecode Container

Chunk holds: instruction stream, constant pool, global names, line mapping, function metadata.

Format: each instruction is 1-3 bytes: [opcode: 1] [operand_hi: 0-1] [operand_lo: 0-1]

### go/compiler/ — Three-Stage Pipeline

1. Lexer — Character-by-character tokenizer, 27 token types, skips comments/whitespace
2. Parser — Recursive descent with precedence climbing for expressions
3. Compiler — AST walk generating bytecode, scope resolution (local/global/upvalue)

### go/vm/ — Bytecode Interpreter

Stack-based VM. Call frames with return address, base pointer, locals array. Native function bridge via interface dispatch.

---

## Self-Hosted Compiler

Same three-stage pipeline, written in Zero. Bootstrap compiler compiles it, then it compiles everything else.

---

## Verification

At each phase, the test suite runs on both compilers and bytecode output must match.