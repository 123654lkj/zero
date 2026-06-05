package chunk

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/123654lkj/zero/go/opcode"
	"github.com/123654lkj/zero/go/value"
)

func TestNewChunk(t *testing.T) {
	c := NewChunk()
	if c == nil {
		t.Fatal("NewChunk returned nil")
	}
	if len(c.Code) != 0 {
		t.Errorf("Code length = %d, want 0", len(c.Code))
	}
	if len(c.Constants) != 0 {
		t.Errorf("Constants length = %d, want 0", len(c.Constants))
	}
	if len(c.Names) != 0 {
		t.Errorf("Names length = %d, want 0", len(c.Names))
	}
	if len(c.Lines) != 0 {
		t.Errorf("Lines length = %d, want 0", len(c.Lines))
	}
	if len(c.Functions) != 0 {
		t.Errorf("Functions length = %d, want 0", len(c.Functions))
	}
}

func TestLen(t *testing.T) {
	c := NewChunk()
	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
	c.WriteByte(byte(opcode.OP_NOP), 1)
	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
	c.WriteByte(byte(opcode.OP_POP), 1)
	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}
}

func TestWriteByte(t *testing.T) {
	c := NewChunk()
	c.WriteByte(byte(opcode.OP_NOP), 10)
	c.WriteByte(byte(opcode.OP_POP), 11)
	c.WriteByte(byte(opcode.OP_DUP), 12)

	if len(c.Code) != 3 {
		t.Fatalf("Code length = %d, want 3", len(c.Code))
	}
	if c.Code[0] != byte(opcode.OP_NOP) {
		t.Errorf("Code[0] = 0x%02X, want OP_NOP", c.Code[0])
	}
	if c.Code[1] != byte(opcode.OP_POP) {
		t.Errorf("Code[1] = 0x%02X, want OP_POP", c.Code[1])
	}
	if c.Code[2] != byte(opcode.OP_DUP) {
		t.Errorf("Code[2] = 0x%02X, want OP_DUP", c.Code[2])
	}

	if len(c.Lines) != 3 {
		t.Fatalf("Lines length = %d, want 3", len(c.Lines))
	}
	if c.Lines[0] != 10 || c.Lines[1] != 11 || c.Lines[2] != 12 {
		t.Errorf("Lines = %v, want [10 11 12]", c.Lines)
	}
}

func TestWriteByteWithOperand(t *testing.T) {
	c := NewChunk()
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 3, 5)

	if len(c.Code) != 2 {
		t.Fatalf("Code length = %d, want 2", len(c.Code))
	}
	if c.Code[0] != byte(opcode.OP_CALL) {
		t.Errorf("Code[0] = 0x%02X, want OP_CALL", c.Code[0])
	}
	if c.Code[1] != 3 {
		t.Errorf("Code[1] = %d, want 3", c.Code[1])
	}
	if len(c.Lines) != 2 {
		t.Fatalf("Lines length = %d, want 2", len(c.Lines))
	}
	if c.Lines[0] != 5 || c.Lines[1] != 5 {
		t.Errorf("Lines = %v, want [5 5]", c.Lines)
	}
}

func TestWriteWordWithOperand(t *testing.T) {
	c := NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 256, 7)

	if len(c.Code) != 3 {
		t.Fatalf("Code length = %d, want 3", len(c.Code))
	}
	if c.Code[0] != byte(opcode.OP_JMP) {
		t.Errorf("Code[0] = 0x%02X, want OP_JMP", c.Code[0])
	}
	// 256 = 0x0100, big-endian: 0x01, 0x00
	if c.Code[1] != 0x01 {
		t.Errorf("Code[1] = 0x%02X, want 0x01", c.Code[1])
	}
	if c.Code[2] != 0x00 {
		t.Errorf("Code[2] = 0x%02X, want 0x00", c.Code[2])
	}
	if len(c.Lines) != 3 {
		t.Fatalf("Lines length = %d, want 3", len(c.Lines))
	}
	for i, l := range c.Lines {
		if l != 7 {
			t.Errorf("Lines[%d] = %d, want 7", i, l)
		}
	}
}

func TestWriteWordWithOperandMaxValue(t *testing.T) {
	c := NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 65535, 1)

	if c.Code[1] != 0xFF {
		t.Errorf("Code[1] = 0x%02X, want 0xFF", c.Code[1])
	}
	if c.Code[2] != 0xFF {
		t.Errorf("Code[2] = 0x%02X, want 0xFF", c.Code[2])
	}
}

func TestAddConstant(t *testing.T) {
	c := NewChunk()

	idx0 := c.AddConstant(value.IntValue(42))
	if idx0 != 0 {
		t.Errorf("AddConstant returned %d, want 0", idx0)
	}

	idx1 := c.AddConstant(value.FloatValue(3.14))
	if idx1 != 1 {
		t.Errorf("AddConstant returned %d, want 1", idx1)
	}

	idx2 := c.AddConstant(value.StringValue("hello"))
	if idx2 != 2 {
		t.Errorf("AddConstant returned %d, want 2", idx2)
	}

	if len(c.Constants) != 3 {
		t.Errorf("Constants length = %d, want 3", len(c.Constants))
	}
}

func TestAddConstantNil(t *testing.T) {
	c := NewChunk()
	idx := c.AddConstant(value.NilValue())
	if idx != 0 {
		t.Errorf("AddConstant(nil) returned %d, want 0", idx)
	}
	if len(c.Constants) != 1 {
		t.Errorf("Constants length = %d, want 1", len(c.Constants))
	}
}

func TestAddConstantBool(t *testing.T) {
	c := NewChunk()
	idx := c.AddConstant(value.BoolValue(true))
	if idx != 0 {
		t.Errorf("AddConstant(bool) returned %d, want 0", idx)
	}
}

func TestAddName(t *testing.T) {
	c := NewChunk()

	idx0 := c.AddName("x")
	if idx0 != 0 {
		t.Errorf("AddName(\"x\") returned %d, want 0", idx0)
	}

	idx1 := c.AddName("y")
	if idx1 != 1 {
		t.Errorf("AddName(\"y\") returned %d, want 1", idx1)
	}

	if len(c.Names) != 2 {
		t.Errorf("Names length = %d, want 2", len(c.Names))
	}
}

func TestAddNameDedup(t *testing.T) {
	c := NewChunk()

	idx0 := c.AddName("x")
	idx1 := c.AddName("y")
	idx2 := c.AddName("x") // duplicate

	if idx0 != idx2 {
		t.Errorf("AddName(\"x\") returned different indices: %d vs %d", idx0, idx2)
	}
	if idx1 != 1 {
		t.Errorf("AddName(\"y\") = %d, want 1", idx1)
	}
	if len(c.Names) != 2 {
		t.Errorf("Names length = %d, want 2 (no duplicate)", len(c.Names))
	}
}

func TestAddNameMultipleDuplicates(t *testing.T) {
	c := NewChunk()

	idx0 := c.AddName("a")
	idx1 := c.AddName("a")
	idx2 := c.AddName("a")

	if idx0 != 0 || idx1 != 0 || idx2 != 0 {
		t.Errorf("Multiple AddName(\"a\") returned different indices: %d, %d, %d", idx0, idx1, idx2)
	}
	if len(c.Names) != 1 {
		t.Errorf("Names length = %d, want 1", len(c.Names))
	}
}

func TestAddFunction(t *testing.T) {
	c := NewChunk()

	idx0 := c.AddFunction(FunctionMeta{
		Name:  "main",
		Arity: 0,
		Start: 0,
		End:   10,
	})
	if idx0 != 0 {
		t.Errorf("AddFunction returned %d, want 0", idx0)
	}

	idx1 := c.AddFunction(FunctionMeta{
		Name:          "add",
		Arity:         2,
		UpvalueCount:  1,
		Start:         10,
		End:           20,
	})
	if idx1 != 1 {
		t.Errorf("AddFunction returned %d, want 1", idx1)
	}

	if len(c.Functions) != 2 {
		t.Errorf("Functions length = %d, want 2", len(c.Functions))
	}
	if c.Functions[0].Name != "main" {
		t.Errorf("Functions[0].Name = %q, want \"main\"", c.Functions[0].Name)
	}
	if c.Functions[1].Arity != 2 {
		t.Errorf("Functions[1].Arity = %d, want 2", c.Functions[1].Arity)
	}
	if c.Functions[1].UpvalueCount != 1 {
		t.Errorf("Functions[1].UpvalueCount = %d, want 1", c.Functions[1].UpvalueCount)
	}
}

func TestLineAt(t *testing.T) {
	c := NewChunk()
	c.WriteByte(byte(opcode.OP_NOP), 10)
	c.WriteByte(byte(opcode.OP_POP), 10)
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 5, 15)
	c.WriteByte(byte(opcode.OP_ADD), 20)

	// LineAt for each byte offset
	tests := []struct {
		offset int
		want   int
	}{
		{0, 10},
		{1, 10},
		{2, 15},
		{3, 15},
		{4, 15},
		{5, 20},
	}

	for _, tc := range tests {
		got := c.LineAt(tc.offset)
		if got != tc.want {
			t.Errorf("LineAt(%d) = %d, want %d", tc.offset, got, tc.want)
		}
	}
}

func TestLineAtOutOfBounds(t *testing.T) {
	c := NewChunk()
	c.WriteByte(byte(opcode.OP_NOP), 5)

	// Negative offset
	if got := c.LineAt(-1); got != -1 {
		t.Errorf("LineAt(-1) = %d, want -1", got)
	}

	// Beyond end
	if got := c.LineAt(100); got != -1 {
		t.Errorf("LineAt(100) = %d, want -1", got)
	}
}

func TestLineAtEmptyChunk(t *testing.T) {
	c := NewChunk()
	if got := c.LineAt(0); got != -1 {
		t.Errorf("LineAt(0) on empty chunk = %d, want -1", got)
	}
}

// captureStdout runs fn with os.Stdout redirected and returns what was printed.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestDisassembleEmpty(t *testing.T) {
	c := NewChunk()
	output := captureStdout(func() {
		c.Disassemble("test_module")
	})

	if !strings.Contains(output, "=== test_module ===") {
		t.Errorf("Disassembly missing header: %s", output)
	}
}

func TestDisassembleNoOperands(t *testing.T) {
	c := NewChunk()
	c.WriteByte(byte(opcode.OP_NOP), 1)
	c.WriteByte(byte(opcode.OP_POP), 1)
	c.WriteByte(byte(opcode.OP_RET), 1)

	output := captureStdout(func() {
		c.Disassemble("simple")
	})

	if !strings.Contains(output, "NOP") {
		t.Errorf("Disassembly missing NOP: %s", output)
	}
	if !strings.Contains(output, "POP") {
		t.Errorf("Disassembly missing POP: %s", output)
	}
	if !strings.Contains(output, "RET") {
		t.Errorf("Disassembly missing RET: %s", output)
	}
}

func TestDisassembleWithOneByteOperand(t *testing.T) {
	c := NewChunk()
	c.AddName("myVar")
	c.WriteByteWithOperand(byte(opcode.OP_LOAD_GLOBAL), 0, 5)

	output := captureStdout(func() {
		c.Disassemble("globals")
	})

	if !strings.Contains(output, "LOAD_GLOBAL") {
		t.Errorf("Disassembly missing LOAD_GLOBAL: %s", output)
	}
	if !strings.Contains(output, "myVar") {
		t.Errorf("Disassembly missing variable name annotation: %s", output)
	}
}

func TestDisassembleWithTwoByteOperand(t *testing.T) {
	c := NewChunk()
	c.AddConstant(value.IntValue(99))
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), 0, 10)

	output := captureStdout(func() {
		c.Disassemble("push_test")
	})

	if !strings.Contains(output, "PUSH") {
		t.Errorf("Disassembly missing PUSH: %s", output)
	}
	// PUSH constant should show the value annotation
	if !strings.Contains(output, "int") {
		t.Errorf("Disassembly missing constant annotation for PUSH: %s", output)
	}
}

func TestDisassembleJumpAnnotation(t *testing.T) {
	c := NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 42, 1)

	output := captureStdout(func() {
		c.Disassemble("jump_test")
	})

	if !strings.Contains(output, "JMP") {
		t.Errorf("Disassembly missing JMP: %s", output)
	}
	if !strings.Contains(output, "-> 42") {
		t.Errorf("Disassembly missing jump target annotation: %s", output)
	}
}

func TestDisassembleInstructionSequence(t *testing.T) {
	c := NewChunk()

	// Build a simple program:
	// PUSH 0 (constant index 0)
	// LOAD_GLOBAL 0 (name index 0)
	// CALL 2
	// POP
	// RET
	c.AddConstant(value.IntValue(42))
	c.AddName("x")

	c.WriteWordWithOperand(byte(opcode.OP_PUSH), 0, 1)
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), 0, 1)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 2, 1)
	c.WriteByte(byte(opcode.OP_POP), 1)
	c.WriteByte(byte(opcode.OP_RET), 1)

	output := captureStdout(func() {
		c.Disassemble("program")
	})

	// Check all opcodes are present
	for _, name := range []string{"PUSH", "LOAD_GLOBAL", "CALL", "POP", "RET"} {
		if !strings.Contains(output, name) {
			t.Errorf("Disassembly missing %s: %s", name, output)
		}
	}
}

func TestDisassembleLineNumbers(t *testing.T) {
	c := NewChunk()
	c.WriteByte(byte(opcode.OP_NOP), 10)
	c.WriteByte(byte(opcode.OP_NOP), 20)

	output := captureStdout(func() {
		c.Disassemble("lines")
	})

	if !strings.Contains(output, "10") {
		t.Errorf("Disassembly missing line 10: %s", output)
	}
	if !strings.Contains(output, "20") {
		t.Errorf("Disassembly missing line 20: %s", output)
	}
}

func TestMixedInstructionWidths(t *testing.T) {
	c := NewChunk()

	c.AddConstant(value.IntValue(1))
	c.AddName("g")

	// 1-byte: NOP
	c.WriteByte(byte(opcode.OP_NOP), 1)
	// 2-byte: CALL 3
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 3, 2)
	// 3-byte: JMP 100
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 100, 3)
	// 2-byte: LOAD_GLOBAL 0
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), 0, 4)
	// 1-byte: RET
	c.WriteByte(byte(opcode.OP_RET), 5)

	// Verify total code length: 1 + 2 + 3 + 3 + 1 = 10
	if len(c.Code) != 10 {
		t.Errorf("Code length = %d, want 10", len(c.Code))
	}

	// Verify lines array matches
	if len(c.Lines) != 10 {
		t.Errorf("Lines length = %d, want 10", len(c.Lines))
	}

	// Verify lineAt works correctly
	expected := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 5}
	for i, want := range expected {
		got := c.LineAt(i)
		if got != want {
			t.Errorf("LineAt(%d) = %d, want %d", i, got, want)
		}
	}
}

func TestChunkIntegration(t *testing.T) {
	// Simulate compiling a small program:
	// x = 42
	// print(x)
	c := NewChunk()

	// Define constants
	c.AddConstant(value.IntValue(42))
	c.AddConstant(value.StringValue("hello"))

	// Define names
	xIdx := c.AddName("x")
	printIdx := c.AddName("print")

	// Write instructions
	// PUSH 0 (push constant 42)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), 0, 1)
	// DEF_GLOBAL 0 (define global "x")
	c.WriteWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(xIdx), 1)
	// LOAD_GLOBAL 0 (load x)
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(xIdx), 2)
	// PUSH 1 (push constant "hello")
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), 1, 2)
	// LOAD_GLOBAL 1 (load print)
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(printIdx), 2)
	// CALL 2 (call print with 2 args)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 2, 2)
	// POP (discard result)
	c.WriteByte(byte(opcode.OP_POP), 2)
	// RET
	c.WriteByte(byte(opcode.OP_RET), 2)

	// Verify state
	// PUSH: 3, DEF_GLOBAL: 3, LOAD_GLOBAL: 3, PUSH: 3, LOAD_GLOBAL: 3, CALL: 2, POP: 1, RET: 1
	// Total: 3+3+3+3+3+2+1+1 = 19
	if c.Len() != 19 {
		t.Errorf("Code length = %d, want 19", c.Len())
	}

	if len(c.Constants) != 2 {
		t.Errorf("Constants = %d, want 2", len(c.Constants))
	}
	if len(c.Names) != 2 {
		t.Errorf("Names = %d, want 2", len(c.Names))
	}

	// Disassemble should work without panic
	output := captureStdout(func() {
		c.Disassemble("integration_test")
	})

	if !strings.Contains(output, "PUSH") {
		t.Errorf("Missing PUSH in disassembly")
	}
	if !strings.Contains(output, "DEF_GLOBAL") {
		t.Errorf("Missing DEF_GLOBAL in disassembly")
	}
	if !strings.Contains(output, "LOAD_GLOBAL") {
		t.Errorf("Missing LOAD_GLOBAL in disassembly")
	}
	if !strings.Contains(output, "CALL") {
		t.Errorf("Missing CALL in disassembly")
	}

	// Verify name annotations appear
	if !strings.Contains(output, "x") {
		t.Errorf("Missing 'x' annotation in disassembly: %s", output)
	}
	if !strings.Contains(output, "print") {
		t.Errorf("Missing 'print' annotation in disassembly: %s", output)
	}

	_ = fmt.Sprintf("") // avoid unused import
}
