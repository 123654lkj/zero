// Package opcode defines the Zero language instruction set.
//
// See doc.go for the full opcode layout and encoding specification.
package opcode

import "fmt"

// Opcode is a single byte identifying a VM instruction.
type Opcode byte

// ──────────────────────────────────────────────────────────────────────────────
// Stack operations  0x00–0x1F
// ──────────────────────────────────────────────────────────────────────────────
const (

	OP_NOP Opcode = iota // 0x00 — No operation
	OP_PUSH              // 0x01 — Push a constant onto the stack (2-byte index)
	OP_POP               // 0x02 — Pop the top value from the stack
	OP_DUP               // 0x03 — Duplicate the top stack value

	// 0x04–0x1F reserved for future stack operations (SWAP, ROT, etc.)
)

// ──────────────────────────────────────────────────────────────────────────────
// Variables  0x20–0x3F
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_LOAD_0    Opcode = 0x20 + iota // Load local variable 0
	OP_LOAD_1                          // Load local variable 1
	OP_LOAD_2                          // Load local variable 2
	OP_LOAD_3                          // Load local variable 3

	OP_STORE_0                         // Store to local variable 0
	OP_STORE_1                         // Store to local variable 1
	OP_STORE_2                         // Store to local variable 2
	OP_STORE_3                         // Store to local variable 3

	OP_LOAD_GLOBAL                     // 0x28 — Load a global variable (2-byte name index)
	OP_STORE_GLOBAL                    // 0x29 — Store to a global variable (2-byte name index)
	OP_DEF_GLOBAL                      // 0x2A — Define a new global variable (2-byte name index)

	// 0x2B–0x3F reserved for future variable operations
)

// ──────────────────────────────────────────────────────────────────────────────
// Arithmetic and logic  0x40–0x5F
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_ADD Opcode = 0x40 + iota // 0x40 — Add two values
	OP_SUB                       // 0x41 — Subtract (second from first)
	OP_MUL                       // 0x42 — Multiply two values
	OP_DIV                       // 0x43 — Divide (first by second)
	OP_MOD                       // 0x44 — Modulo (first mod second)
	OP_NEG                       // 0x45 — Negate a value (arithmetic minus)
	OP_NOT                       // 0x46 — Logical not
	OP_EQ                        // 0x47 — Equal
	OP_NEQ                       // 0x48 — Not equal
	OP_LT                        // 0x49 — Less than
	OP_GT                        // 0x4A — Greater than
	OP_LTE                       // 0x4B — Less than or equal
	OP_GTE                       // 0x4C — Greater than or equal
	OP_AND                       // 0x4D — Logical and
	OP_OR                        // 0x4E — Logical or

	// 0x4F–0x5F reserved for future arithmetic/logic ops (XOR, SHL, SHR, etc.)
)

// ──────────────────────────────────────────────────────────────────────────────
// Control flow / Jumps  0x60–0x7F
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_JMP     Opcode = 0x60 + iota // 0x60 — Unconditional jump (2-byte offset)
	OP_JMP_IF                       // 0x61 — Jump if top of stack is truthy (2-byte offset)
	OP_JMP_IFN                      // 0x62 — Jump if top of stack is falsy (2-byte offset)

	// 0x63–0x7F reserved for future control flow (LOOP, BREAK, SWITCH, etc.)
)

// OP_JMP_IF_NOT is an alias for OP_JMP_IFN for clarity in some contexts.
const OP_JMP_IF_NOT = OP_JMP_IFN

// ──────────────────────────────────────────────────────────────────────────────
// Functions  0x80–0x9F
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_CALL    Opcode = 0x80 + iota // 0x80 — Call a function (1-byte arg count)
	OP_RET                          // 0x81 — Return from current function
	OP_CLOSURE                      // 0x82 — Create a closure capturing the current scope

	// 0x83–0x9F reserved for future function-related ops (TAIL_CALL, DEFER, etc.)
)

// ──────────────────────────────────────────────────────────────────────────────
// Phase 0 special opcodes  0x90–0x9F
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_PRINT Opcode = 0x90 // Print top of stack to stdout
	OP_HALT  Opcode = 0x91 // Halt the VM
)

// ──────────────────────────────────────────────────────────────────────────────
// Data structures  0xA0–0xBF  (Phase 1+)
// ──────────────────────────────────────────────────────────────────────────────
const (
	OP_ARRAY_NEW Opcode = 0xA0 + iota // 0xA0 — Create a new array
	OP_ARRAY_GET                       // 0xA1 — Get element from array by index
	OP_ARRAY_SET                       // 0xA2 — Set element in array by index
	OP_MAP_NEW                         // 0xA3 — Create a new map
	OP_MAP_GET                         // 0xA4 — Get value from map by key
	OP_MAP_SET                         // 0xA5 — Set value in map by key

	// 0xA6–0xBF reserved for future data structure ops (SET, TUPLE, RECORD, etc.)
)

// ──────────────────────────────────────────────────────────────────────────────
// Concurrency  0xC0–0xDF  (Phase 1+)
// ──────────────────────────────────────────────────────────────────────────────
//
// Planned opcodes (Phase 1+):
//
//	OP_SPAWN   — Launch a goroutine / async task
//	OP_SEND    — Send a value on a channel
//	OP_RECEIVE — Receive a value from a channel
//	OP_CHAN_NEW — Create a new channel
//	OP_JOIN    — Wait for a goroutine to finish
//	OP_SELECT  — Multiplex over multiple channels
//	0xC6–0xDF  — Reserved for MUTEX, LOCK, UNLOCK, etc.

// ──────────────────────────────────────────────────────────────────────────────
// Zero-specific features  0xE0–0xFF  (Phase 2+)
// ──────────────────────────────────────────────────────────────────────────────
//
// Planned opcodes (Phase 2+):
//
//	OP_IO_OPEN    — Open a file / resource
//	OP_IO_READ    — Read from a handle
//	OP_IO_WRITE   — Write to a handle
//	OP_IO_CLOSE   — Close a handle
//	OP_IMAGE_LOAD — Load an image
//	OP_IMAGE_SAVE — Save an image
//	OP_IMAGE_PROC — Process / transform an image
//	OP_QUERY_EXEC — Execute a structured query
//	OP_AI_INFER   — Run an AI inference call

// ──────────────────────────────────────────────────────────────────────────────
// Operand counts
// ──────────────────────────────────────────────────────────────────────────────

// operandCounts maps each implemented opcode to the number of operand bytes
// that follow it in the bytecode stream.
var operandCounts = map[Opcode]int{
	// Stack — 0 operands
	OP_NOP: 0,
	OP_POP: 0,
	OP_DUP: 0,

	// PUSH — 2-byte constant pool index
	OP_PUSH: 2,

	// Variables — LOAD/STORE locals use no operand (slot encoded in opcode)
	OP_LOAD_0:       0,
	OP_LOAD_1:       0,
	OP_LOAD_2:       0,
	OP_LOAD_3:       0,
	OP_STORE_0:      0,
	OP_STORE_1:      0,
	OP_STORE_2:      0,
	OP_STORE_3:      0,
	OP_LOAD_GLOBAL:  2, // 2-byte name index
	OP_STORE_GLOBAL: 2, // 2-byte name index
	OP_DEF_GLOBAL:   2, // 2-byte name index

	// Arithmetic / logic — 0 operands
	OP_ADD: 0,
	OP_SUB: 0,
	OP_MUL: 0,
	OP_DIV: 0,
	OP_MOD: 0,
	OP_NEG: 0,
	OP_NOT: 0,
	OP_EQ:  0,
	OP_NEQ: 0,
	OP_LT:  0,
	OP_GT:  0,
	OP_LTE: 0,
	OP_GTE: 0,
	OP_AND: 0,
	OP_OR:  0,

	// Jumps — 2-byte signed offset
	OP_JMP:      2,
	OP_JMP_IF:   2,
	OP_JMP_IFN:  2,

	// Functions
	OP_CALL:    1, // 1-byte argument count
	OP_RET:     0,
	OP_CLOSURE: 0,

	// Data structures
	OP_ARRAY_NEW: 0,
	OP_ARRAY_GET: 0,
	OP_ARRAY_SET: 0,
	OP_MAP_NEW:   0,
	OP_MAP_GET:   0,
	OP_MAP_SET:   0,

	// Phase 0 specials
	OP_PRINT: 0,
	OP_HALT:  0,
}

// OperandCount returns the number of operand bytes that follow this opcode
// in the bytecode stream. The VM reads this many bytes as the instruction's
// operand and advances the program counter by 1 (opcode) + N (operand bytes).
func (o Opcode) OperandCount() int {
	if n, ok := operandCounts[o]; ok {
		return n
	}
	// Unknown / future opcode — conservatively assume 0 operands.
	return 0
}

// opcodeNames provides a human-readable name for every implemented opcode.
var opcodeNames = map[Opcode]string{
	OP_NOP:          "NOP",
	OP_PUSH:         "PUSH",
	OP_POP:          "POP",
	OP_DUP:          "DUP",
	OP_LOAD_0:       "LOAD_0",
	OP_LOAD_1:       "LOAD_1",
	OP_LOAD_2:       "LOAD_2",
	OP_LOAD_3:       "LOAD_3",
	OP_STORE_0:      "STORE_0",
	OP_STORE_1:      "STORE_1",
	OP_STORE_2:      "STORE_2",
	OP_STORE_3:      "STORE_3",
	OP_LOAD_GLOBAL:  "LOAD_GLOBAL",
	OP_STORE_GLOBAL: "STORE_GLOBAL",
	OP_DEF_GLOBAL:   "DEF_GLOBAL",
	OP_ADD:          "ADD",
	OP_SUB:          "SUB",
	OP_MUL:          "MUL",
	OP_DIV:          "DIV",
	OP_MOD:          "MOD",
	OP_NEG:          "NEG",
	OP_NOT:          "NOT",
	OP_EQ:           "EQ",
	OP_NEQ:          "NEQ",
	OP_LT:           "LT",
	OP_GT:           "GT",
	OP_LTE:          "LTE",
	OP_GTE:          "GTE",
	OP_AND:          "AND",
	OP_OR:           "OR",
	OP_JMP:          "JMP",
	OP_JMP_IF:       "JMP_IF",
	OP_JMP_IFN:      "JMP_IFN",
	OP_CALL:         "CALL",
	OP_RET:          "RET",
	OP_CLOSURE:      "CLOSURE",
	OP_ARRAY_NEW:    "ARRAY_NEW",
	OP_ARRAY_GET:    "ARRAY_GET",
	OP_ARRAY_SET:    "ARRAY_SET",
	OP_MAP_NEW:      "MAP_NEW",
	OP_MAP_GET:      "MAP_GET",
	OP_MAP_SET:      "MAP_SET",
	OP_PRINT:        "PRINT",
	OP_HALT:         "HALT",
}

// Name returns the human-readable name of the opcode (e.g. "ADD", "JMP_IF").
// If the opcode is not recognized, it returns "UNKNOWN(0xHH)".
func (o Opcode) Name() string {
	if name, ok := opcodeNames[o]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(0x%02X)", byte(o))
}

// String implements [fmt.Stringer] and returns a debug representation such
// as "ADD [0x40]".
func (o Opcode) String() string {
	if name, ok := opcodeNames[o]; ok {
		return fmt.Sprintf("%s [0x%02X]", name, byte(o))
	}
	return fmt.Sprintf("UNKNOWN [0x%02X]", byte(o))
}

// Byte returns the underlying numeric value of the opcode.
func (o Opcode) Byte() byte {
	return byte(o)
}

// IsValid reports whether o is a recognized opcode in the current phase.
func (o Opcode) IsValid() bool {
	_, ok := opcodeNames[o]
	return ok
}

// ──────────────────────────────────────────────────────────────────────────────
// Convenience helpers for programmatic access to load/store families.
// ──────────────────────────────────────────────────────────────────────────────

// LoadLocal returns the LOAD opcode for local slot n (0–3).
// Panics if n is out of range.
func LoadLocal(n int) Opcode {
	switch n {
	case 0:
		return OP_LOAD_0
	case 1:
		return OP_LOAD_1
	case 2:
		return OP_LOAD_2
	case 3:
		return OP_LOAD_3
	default:
		panic("opcode: LoadLocal: slot out of range 0–3")
	}
}

// StoreLocal returns the STORE opcode for local slot n (0–3).
// Panics if n is out of range.
func StoreLocal(n int) Opcode {
	switch n {
	case 0:
		return OP_STORE_0
	case 1:
		return OP_STORE_1
	case 2:
		return OP_STORE_2
	case 3:
		return OP_STORE_3
	default:
		panic("opcode: StoreLocal: slot out of range 0–3")
	}
}
