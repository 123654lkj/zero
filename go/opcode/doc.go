// Package opcode defines the instruction set for the Zero language virtual machine.
//
// # Opcode Layout
//
// Zero uses a compact bytecoded instruction set with opcodes allocated in 32-byte
// (0x20) ranges to allow future expansion within each category:
//
//	0x00–0x1F  Stack operations       (PUSH, POP, DUP, …)
//	0x20–0x3F  Variables              (LOAD, STORE, …)
//	0x40–0x5F  Arithmetic and Logic   (ADD, SUB, EQ, …)
//	0x60–0x7F  Control flow / Jumps   (JMP, JMP_IF, …)
//	0x80–0x9F  Functions              (CALL, RET, CLOSURE)
//	0xA0–0xBF  Data structures        (ARRAY, MAP, …)
//	0xC0–0xDF  Concurrency            (SPAWN, SEND, RECEIVE)  [Phase 1+]
//	0xE0–0xFF  Zero-specific features (IO, IMAGE, QUERY, AI)  [Phase 2+]
//
// # Phase 0
//
// Phase 0 implements opcodes in the range 0x00–0x9F, covering stack manipulation,
// local/global variables, arithmetic and logic, control flow, and function calls.
// Opcodes in the 0xA0–0xFF ranges are defined as reserved placeholders for future phases.
//
// # Operand Encoding
//
// Some opcodes carry a variable-length operand encoded in the bytes immediately
// following the opcode. Operand widths are defined per-opcode and can be queried
// via [Opcode.OperandCount]. The VM reads that many bytes and advances the program
// counter accordingly.
//
// # Extensions
//
// The 0xC0–0xDF range is reserved for concurrency primitives (goroutines, channels,
// synchronization). The 0xE0–0xFF range is reserved for Zero's unique features
// such as I/O bindings, image processing, structured queries, and AI integration.
package opcode
