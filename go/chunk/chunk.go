// Package chunk provides the bytecode container for the Zero language compiler and VM.
//
// A Chunk holds a contiguous instruction stream, a constant pool, a global name table,
// per-byte line information (for error reporting), and function metadata.
//
// # Instruction Encoding
//
// Each instruction is 1–3 bytes:
//
//	[opcode: 1] [operand_hi: 0–1] [operand_lo: 0–1]
//
// The opcode package determines the width via Opcode.OperandCount().
package chunk

import (
	"fmt"

	"github.com/123654lkj/zero/go/opcode"
	"github.com/123654lkj/zero/go/value"
)

// FunctionMeta describes a function defined in this chunk.
type FunctionMeta struct {
	Name         string
	Arity        int // number of parameters
	UpvalueCount int
	Start        int // byte offset into Code where function body begins
	End          int // byte offset where function body ends
}

// AIFunctionMeta describes an AI function defined in this chunk.
type AIFunctionMeta struct {
	Name       string
	ParamNames []string // parameter names for prompt template substitution
	Prompt     string   // prompt template with {{param}} placeholders
	Model      string   // optional model specification (empty = default)
	Cache      bool     // whether to cache responses
	ReturnType string   // optional return type annotation
}

// Chunk is the bytecode container for a Zero module or function body.
type Chunk struct {
	Code        []byte          // raw instruction bytes
	Constants   []value.Value   // constant pool
	Names       []string        // global name table
	Lines       []int           // line number for each byte in Code
	Functions   []FunctionMeta  // function metadata
	AIFunctions []AIFunctionMeta // AI function metadata
}

// NewChunk creates an empty Chunk ready to accept instructions.
func NewChunk() *Chunk {
	return &Chunk{
		Code:      make([]byte, 0, 64),
		Constants: make([]value.Value, 0, 16),
		Names:     make([]string, 0, 16),
		Lines:     make([]int, 0, 64),
		Functions: make([]FunctionMeta, 0, 4),
	}
}

// Len returns the current length of the instruction stream.
func (c *Chunk) Len() int {
	return len(c.Code)
}

// WriteByte appends a single-byte (no-operand) instruction.
func (c *Chunk) WriteByte(op byte, line int) {
	c.Code = append(c.Code, op)
	c.Lines = append(c.Lines, line)
}

// WriteByteWithOperand appends a 2-byte instruction: opcode + 1-byte operand.
func (c *Chunk) WriteByteWithOperand(op byte, operand byte, line int) {
	c.Code = append(c.Code, op, operand)
	c.Lines = append(c.Lines, line, line)
}

// WriteWordWithOperand appends a 3-byte instruction: opcode + 2-byte (16-bit) operand.
// The operand is encoded in big-endian order (high byte first).
func (c *Chunk) WriteWordWithOperand(op byte, operand uint16, line int) {
	c.Code = append(c.Code, op, byte(operand>>8), byte(operand))
	c.Lines = append(c.Lines, line, line, line)
}

// AddConstant adds a value to the constant pool and returns its index.
func (c *Chunk) AddConstant(v value.Value) int {
	idx := len(c.Constants)
	c.Constants = append(c.Constants, v)
	return idx
}

// AddName adds a global name to the name table and returns its index.
// If the name already exists, the existing index is returned.
func (c *Chunk) AddName(name string) int {
	for i, n := range c.Names {
		if n == name {
			return i
		}
	}
	idx := len(c.Names)
	c.Names = append(c.Names, name)
	return idx
}

// AddFunction appends function metadata and returns its index.
func (c *Chunk) AddFunction(meta FunctionMeta) int {
	idx := len(c.Functions)
	c.Functions = append(c.Functions, meta)
	return idx
}

// AddAIFunction appends AI function metadata and returns its index.
func (c *Chunk) AddAIFunction(meta AIFunctionMeta) int {
	idx := len(c.AIFunctions)
	c.AIFunctions = append(c.AIFunctions, meta)
	return idx
}

// LineAt returns the source line number for the instruction at the given byte offset.
func (c *Chunk) LineAt(offset int) int {
	if offset >= 0 && offset < len(c.Lines) {
		return c.Lines[offset]
	}
	return -1
}

// Disassemble prints a human-readable disassembly of the chunk to stdout.
func (c *Chunk) Disassemble(name string) {
	fmt.Printf("=== %s ===\n", name)
	fmt.Printf("%4s  %6s  %-16s %s\n", "Off", "Line", "Opcode", "Operands")
	fmt.Println("----  ------  ----------------  --------")

	offset := 0
	for offset < len(c.Code) {
		offset = c.disassembleInstruction(offset)
	}
}

// disassembleInstruction disassembles a single instruction starting at offset
// and returns the offset of the next instruction.
func (c *Chunk) disassembleInstruction(offset int) int {
	op := c.Code[offset]
	line := -1
	if offset < len(c.Lines) {
		line = c.Lines[offset]
	}

	opc := opcode.Opcode(op)
	operandCount := opc.OperandCount()

	switch operandCount {
	case 0:
		fmt.Printf("%4d  %6d  %-16s\n", offset, line, opc.Name())
		return offset + 1
	case 1:
		operand := byte(0)
		if offset+1 < len(c.Code) {
			operand = c.Code[offset+1]
		}
		fmt.Printf("%4d  %6d  %-16s %d\n", offset, line, opc.Name(), operand)
		return offset + 2
	case 2:
		hi := byte(0)
		lo := byte(0)
		if offset+1 < len(c.Code) {
			hi = c.Code[offset+1]
		}
		if offset+2 < len(c.Code) {
			lo = c.Code[offset+2]
		}
		operand := uint16(hi)<<8 | uint16(lo)

		// Provide human-readable annotations for common operand types
		extra := ""
		switch {
		case opc == opcode.OP_LOAD_GLOBAL || opc == opcode.OP_STORE_GLOBAL || opc == opcode.OP_DEF_GLOBAL:
			if int(operand) < len(c.Names) {
				extra = fmt.Sprintf(" (%s)", c.Names[operand])
			}
		case opc == opcode.OP_PUSH:
			if int(operand) < len(c.Constants) {
				extra = fmt.Sprintf(" (%s)", c.Constants[operand])
			}
		case opc == opcode.OP_JMP || opc == opcode.OP_JMP_IF || opc == opcode.OP_JMP_IFN:
			extra = fmt.Sprintf(" -> %d", operand)
		}

		if extra != "" {
			fmt.Printf("%4d  %6d  %-16s %d%s\n", offset, line, opc.Name(), operand, extra)
		} else {
			fmt.Printf("%4d  %6d  %-16s %d\n", offset, line, opc.Name(), operand)
		}
		return offset + 3
	default:
		fmt.Printf("%4d  %6d  %-16s\n", offset, line, opc.Name())
		return offset + 1
	}
}
