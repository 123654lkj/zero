package vm

import (
	"testing"

	"github.com/123654lkj/zero/go/chunk"
	"github.com/123654lkj/zero/go/opcode"
	"github.com/123654lkj/zero/go/value"
)

// ---------------------------------------------------------------------------
// 1. Push / Pop / Dup
// ---------------------------------------------------------------------------

func TestPushPop(t *testing.T) {
	c := chunk.NewChunk()
	// PUSH 42, POP, PUSH 99, HALT  →  stack ends with 99
	idx42 := c.AddConstant(value.IntValue(42))
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(idx42), 1)
	c.WriteByte(byte(opcode.OP_POP), 1)

	idx99 := c.AddConstant(value.IntValue(99))
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(idx99), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	result := vm.RunChunk(c)
	if !result.IsInt() || result.AsInt() != 99 {
		t.Fatalf("expected int(99), got %s", result.String())
	}
}

func TestDup(t *testing.T) {
	c := chunk.NewChunk()
	idx := c.AddConstant(value.IntValue(7))
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(idx), 1) // push 7
	c.WriteByte(byte(opcode.OP_DUP), 1)                          // stack: [7, 7]
	c.WriteByte(byte(opcode.OP_POP), 1)                          // stack: [7]
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	result := vm.RunChunk(c)
	if !result.IsInt() || result.AsInt() != 7 {
		t.Fatalf("expected int(7), got %s", result.String())
	}
}

// ---------------------------------------------------------------------------
// 2. Arithmetic
// ---------------------------------------------------------------------------

func TestArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int64
		op       opcode.Opcode
		expected int64
	}{
		{"add", 3, 4, opcode.OP_ADD, 7},
		{"sub", 10, 3, opcode.OP_SUB, 7},
		{"mul", 3, 4, opcode.OP_MUL, 12},
		{"div", 12, 4, opcode.OP_DIV, 3},
		{"mod", 10, 3, opcode.OP_MOD, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chunk.NewChunk()
			c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(tt.a))), 1)
			c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(tt.b))), 1)
			c.WriteByte(byte(tt.op), 1)
			c.WriteByte(byte(opcode.OP_HALT), 1)

			vm := NewVM()
			r := vm.RunChunk(c)
			if !r.IsInt() || r.AsInt() != tt.expected {
				t.Fatalf("expected int(%d), got %s", tt.expected, r.String())
			}
		})
	}
}

func TestNeg(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(5))), 1)
	c.WriteByte(byte(opcode.OP_NEG), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != -5 {
		t.Fatalf("expected int(-5), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 3. Comparisons
// ---------------------------------------------------------------------------

func TestComparisons(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int64
		op       opcode.Opcode
		expected bool
	}{
		{"eq_equal", 5, 5, opcode.OP_EQ, true},
		{"eq_not_equal", 5, 3, opcode.OP_EQ, false},
		{"neq", 5, 3, opcode.OP_NEQ, true},
		{"lt_true", 3, 5, opcode.OP_LT, true},
		{"lt_false", 5, 3, opcode.OP_LT, false},
		{"gt_true", 5, 3, opcode.OP_GT, true},
		{"gt_false", 3, 5, opcode.OP_GT, false},
		{"lte_equal", 5, 5, opcode.OP_LTE, true},
		{"lte_less", 3, 5, opcode.OP_LTE, true},
		{"lte_greater", 5, 3, opcode.OP_LTE, false},
		{"gte_equal", 5, 5, opcode.OP_GTE, true},
		{"gte_greater", 5, 3, opcode.OP_GTE, true},
		{"gte_less", 3, 5, opcode.OP_GTE, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chunk.NewChunk()
			c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(tt.a))), 1)
			c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(tt.b))), 1)
			c.WriteByte(byte(tt.op), 1)
			c.WriteByte(byte(opcode.OP_HALT), 1)

			vm := NewVM()
			r := vm.RunChunk(c)
			if !r.IsBool() || r.AsBool() != tt.expected {
				t.Fatalf("expected bool(%v), got %s", tt.expected, r.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. Jumps
// ---------------------------------------------------------------------------

func TestJmp(t *testing.T) {
	c := chunk.NewChunk()
	// JMP forward 4 bytes → skip PUSH 999 + HALT → land on PUSH 42 + HALT
	c.WriteWordWithOperand(byte(opcode.OP_JMP), 4, 1)
	// ---- skipped ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(999))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)
	// ---- target ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(42))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 42 {
		t.Fatalf("expected int(42), got %s", r.String())
	}
}

func TestJmpIf(t *testing.T) {
	c := chunk.NewChunk()
	// Push true, JMP_IF forward 4 → skip PUSH 999 + HALT
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(true))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_JMP_IF), 4, 1)
	// ---- skipped ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(999))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)
	// ---- target ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(77))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 77 {
		t.Fatalf("expected int(77), got %s", r.String())
	}
}

func TestJmpIfN(t *testing.T) {
	c := chunk.NewChunk()
	// Push nil, JMP_IFN forward 4 → jump (nil is falsy) → skip PUSH 999 + HALT
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.NilValue())), 1)
	c.WriteWordWithOperand(byte(opcode.OP_JMP_IFN), 4, 1)
	// ---- skipped ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(999))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)
	// ---- target ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(55))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 55 {
		t.Fatalf("expected int(55), got %s", r.String())
	}
}

func TestJmpIfNotTaken(t *testing.T) {
	c := chunk.NewChunk()
	// Push true, JMP_IFN forward 4 → NOT taken (true is truthy) → execute PUSH 999 + HALT
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(true))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_JMP_IFN), 4, 1)
	// ---- executed (not skipped) ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(999))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)
	// ---- never reached ----
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(11))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 999 {
		t.Fatalf("expected int(999), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 5. Global variables
// ---------------------------------------------------------------------------

func TestGlobalVariables(t *testing.T) {
	c := chunk.NewChunk()
	nameIdx := c.AddName("x")

	// DEF_GLOBAL x = 42
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(42))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_DEF_GLOBAL), uint16(nameIdx), 1)

	// STORE_GLOBAL x = 100
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(100))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_STORE_GLOBAL), uint16(nameIdx), 1)

	// LOAD_GLOBAL x
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(nameIdx), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 100 {
		t.Fatalf("expected int(100), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 6. Local variables
// ---------------------------------------------------------------------------

func TestLocalVariables(t *testing.T) {
	c := chunk.NewChunk()

	// Allocate 2 local slots (push placeholder values so slots 0 and 1 exist)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(0))), 1) // slot 0
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(0))), 1) // slot 1

	// Store 10 into slot 0: PUSH 10, STORE_0, POP (cleanup pushed value)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(10))), 1)
	c.WriteByte(byte(opcode.OP_STORE_0), 1)
	c.WriteByte(byte(opcode.OP_POP), 1)

	// Store 20 into slot 1: PUSH 20, STORE_1, POP
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(20))), 1)
	c.WriteByte(byte(opcode.OP_STORE_1), 1)
	c.WriteByte(byte(opcode.OP_POP), 1)

	// Load slot 0 + Load slot 1 + ADD
	c.WriteByte(byte(opcode.OP_LOAD_0), 1)
	c.WriteByte(byte(opcode.OP_LOAD_1), 1)
	c.WriteByte(byte(opcode.OP_ADD), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 30 {
		t.Fatalf("expected int(30), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 7. Built-in functions
// ---------------------------------------------------------------------------

func TestBuiltinLenString(t *testing.T) {
	c := chunk.NewChunk()
	nameIdx := c.AddName("len")
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(nameIdx), 1) // push fn
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.StringValue("hello"))), 1)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 1, 1) // len("hello"), 1 arg
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 5 {
		t.Fatalf("expected int(5), got %s", r.String())
	}
}

func TestBuiltinLenArray(t *testing.T) {
	c := chunk.NewChunk()
	// len( [1, 2, 3] )
	nameIdx := c.AddName("len")
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(nameIdx), 1) // fn

	// push array elements + count, then ARRAY_NEW
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(1))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(2))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(3))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(3))), 1) // count
	c.WriteByte(byte(opcode.OP_ARRAY_NEW), 1)

	c.WriteByteWithOperand(byte(opcode.OP_CALL), 1, 1) // len(array), 1 arg
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 3 {
		t.Fatalf("expected int(3), got %s", r.String())
	}
}

func TestBuiltinType(t *testing.T) {
	c := chunk.NewChunk()
	nameIdx := c.AddName("type")
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(nameIdx), 1) // fn
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(42))), 1)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 1, 1) // type(42), 1 arg
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsString() || r.AsString() != "Int" {
		t.Fatalf("expected string(\"Int\"), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 8. End-to-end: print(42)
// ---------------------------------------------------------------------------

func TestEndToEndPrint42(t *testing.T) {
	c := chunk.NewChunk()
	// load "print" → push 42 → CALL 1 → HALT
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(c.AddName("print")), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(42))), 1)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 1, 1) // print(42), 1 arg
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsNil() {
		t.Fatalf("expected nil (print returns nil), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 9. End-to-end: print(2 + 3)
// ---------------------------------------------------------------------------

func TestEndToEndPrint2Plus3(t *testing.T) {
	c := chunk.NewChunk()
	// load "print" → push 2 → push 3 → ADD → CALL 1 → HALT
	c.WriteWordWithOperand(byte(opcode.OP_LOAD_GLOBAL), uint16(c.AddName("print")), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(2))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(3))), 1)
	c.WriteByte(byte(opcode.OP_ADD), 1)
	c.WriteByteWithOperand(byte(opcode.OP_CALL), 1, 1) // print(5), 1 arg
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsNil() {
		t.Fatalf("expected nil (print returns nil), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// 10. HALT stops execution
// ---------------------------------------------------------------------------

func TestHalt(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(42))), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)
	// This PUSH should never execute.
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(99))), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 42 {
		t.Fatalf("expected int(42), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// Extra: string concatenation, float arithmetic, NOT, AND/OR
// ---------------------------------------------------------------------------

func TestStringConcat(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.StringValue("hello "))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.StringValue("world"))), 1)
	c.WriteByte(byte(opcode.OP_ADD), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsString() || r.AsString() != "hello world" {
		t.Fatalf("expected string(\"hello world\"), got %s", r.String())
	}
}

func TestFloatArithmetic(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.FloatValue(2.5))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.FloatValue(3.5))), 1)
	c.WriteByte(byte(opcode.OP_ADD), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsFloat() || r.AsFloat() != 6.0 {
		t.Fatalf("expected float(6), got %s", r.String())
	}
}

func TestIntFloatPromotion(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(2))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.FloatValue(3.5))), 1)
	c.WriteByte(byte(opcode.OP_ADD), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsFloat() || r.AsFloat() != 5.5 {
		t.Fatalf("expected float(5.5), got %s", r.String())
	}
}

func TestNot(t *testing.T) {
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(false))), 1)
	c.WriteByte(byte(opcode.OP_NOT), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsBool() || !r.AsBool() {
		t.Fatalf("expected bool(true), got %s", r.String())
	}
}

func TestAndOr(t *testing.T) {
	// (true AND false) OR true  →  true
	c := chunk.NewChunk()
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(true))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(false))), 1)
	c.WriteByte(byte(opcode.OP_AND), 1) // false
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.BoolValue(true))), 1)
	c.WriteByte(byte(opcode.OP_OR), 1) // true
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsBool() || !r.AsBool() {
		t.Fatalf("expected bool(true), got %s", r.String())
	}
}

// ---------------------------------------------------------------------------
// Extra: Array and Map operations
// ---------------------------------------------------------------------------

func TestArrayOperations(t *testing.T) {
	c := chunk.NewChunk()
	// Create [10, 20, 30], get index 1, expect 20
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(10))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(20))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(30))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(3))), 1) // count
	c.WriteByte(byte(opcode.OP_ARRAY_NEW), 1)

	// Get index 1
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(1))), 1)
	c.WriteByte(byte(opcode.OP_ARRAY_GET), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 20 {
		t.Fatalf("expected int(20), got %s", r.String())
	}
}

func TestMapOperations(t *testing.T) {
	c := chunk.NewChunk()
	// Create {"a": 1}, get "a", expect 1
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.StringValue("a"))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(1))), 1)
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.IntValue(1))), 1) // count
	c.WriteByte(byte(opcode.OP_MAP_NEW), 1)

	// Get "a"
	c.WriteWordWithOperand(byte(opcode.OP_PUSH), uint16(c.AddConstant(value.StringValue("a"))), 1)
	c.WriteByte(byte(opcode.OP_MAP_GET), 1)
	c.WriteByte(byte(opcode.OP_HALT), 1)

	vm := NewVM()
	r := vm.RunChunk(c)
	if !r.IsInt() || r.AsInt() != 1 {
		t.Fatalf("expected int(1), got %s", r.String())
	}
}
