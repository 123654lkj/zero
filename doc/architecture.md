# Zero Architecture

## Overview

```
             .zero source file
                   |
              +----+----+
              |  Lexer   |
              +----+----+
                   | tokens
              +----+----+
              |  Parser  |
              +----+----+
                   | AST
              +----+----+
              | Compiler |
              +----+----+
                   | bytecode (.zbc)
              +----+----+
              |    VM    |
              +----+----+
                   | result
```

## Bootstrap Compiler (Go)

### go/value/
The value system. Every Zero value is an 8-byte struct:
- 1 byte: type tag (Nil, Bool, Int, Float, String, Array, Map, Closure, Native, Pattern, Tagged, Stream, Image, IO, Table)
- 7 bytes: payload (embedded for small values, pointer for heap-allocated)

Small integers (<2^31) and short strings (<7 bytes) are embedded directly. Zero heap allocation for common cases.

### go/opcode/
39 instructions organized into groups with reserved ranges for future expansion. Each instruction is 1-3 bytes.

### go/chunk/
Bytecode container: instruction stream, constant pool, global names, line mapping, function metadata.

### go/compiler/
Three-stage pipeline:
1. Lexer — character-by-character tokenizer, 27 token types
2. Parser — recursive descent with precedence climbing
3. Compiler — AST walk generating bytecode, scope resolution (local/global/upvalue)

### go/vm/
Stack-based bytecode interpreter. Call frames with return addresses and local storage. Native function bridge for Go runtime functions.

## Self-Hosted Compiler (Zero)

The compiler is rewritten in Zero itself. Same three-stage pipeline, expressed in Zero syntax. Bootstrap compiler compiles self-hosted compiler, which then compiles everything else.

## Bytecode Format

Each instruction is 1-3 bytes:
```
[opcode: 1 byte] [operand_hi: 0-1 bytes] [operand_lo: 0-1 bytes]
```

Constants are stored in a per-chunk constant pool. Variables are resolved at compile time to indices.

## VM Design

- Stack-based execution
- Register window per call frame (locals, upvalues)
- Go garbage collection for heap values
- Native function bridge via interface dispatch