package opcode

import (
	"testing"
)

func TestOpcodeValues(t *testing.T) {
	// Verify that the opcode constants are at the expected numeric positions.
	tests := []struct {
		op   Opcode
		want byte
	}{
		// Stack
		{OP_NOP, 0x00},
		{OP_PUSH, 0x01},
		{OP_POP, 0x02},
		{OP_DUP, 0x03},

		// Variables
		{OP_LOAD_0, 0x20},
		{OP_LOAD_1, 0x21},
		{OP_LOAD_2, 0x22},
		{OP_LOAD_3, 0x23},
		{OP_STORE_0, 0x24},
		{OP_STORE_1, 0x25},
		{OP_STORE_2, 0x26},
		{OP_STORE_3, 0x27},
		{OP_LOAD_GLOBAL, 0x28},
		{OP_STORE_GLOBAL, 0x29},
		{OP_DEF_GLOBAL, 0x2A},

		// Arithmetic / logic
		{OP_ADD, 0x40},
		{OP_SUB, 0x41},
		{OP_MUL, 0x42},
		{OP_DIV, 0x43},
		{OP_MOD, 0x44},
		{OP_NEG, 0x45},
		{OP_NOT, 0x46},
		{OP_EQ, 0x47},
		{OP_NEQ, 0x48},
		{OP_LT, 0x49},
		{OP_GT, 0x4A},
		{OP_LTE, 0x4B},
		{OP_GTE, 0x4C},
		{OP_AND, 0x4D},
		{OP_OR, 0x4E},

		// Jumps
		{OP_JMP, 0x60},
		{OP_JMP_IF, 0x61},
		{OP_JMP_IFN, 0x62},

		// Functions
		{OP_CALL, 0x80},
		{OP_RET, 0x81},
		{OP_CLOSURE, 0x82},

		// Data structures
		{OP_ARRAY_NEW, 0xA0},
		{OP_ARRAY_GET, 0xA1},
		{OP_ARRAY_SET, 0xA2},
		{OP_MAP_NEW, 0xA3},
		{OP_MAP_GET, 0xA4},
		{OP_MAP_SET, 0xA5},

		// Phase 0 specials
		{OP_PRINT, 0x90},
		{OP_HALT, 0x91},
	}

	for _, tc := range tests {
		if byte(tc.op) != tc.want {
			t.Errorf("op = %d (0x%02X), want 0x%02X", tc.op, byte(tc.op), tc.want)
		}
	}
}

func TestOpcodeString(t *testing.T) {
	tests := []struct {
		op   Opcode
		want string
	}{
		{OP_ADD, "ADD [0x40]"},
		{OP_PUSH, "PUSH [0x01]"},
		{OP_JMP_IFN, "JMP_IFN [0x62]"},
		{OP_HALT, "HALT [0x91]"},
		{Opcode(0xFF), "UNKNOWN [0xFF]"},
	}

	for _, tc := range tests {
		got := tc.op.String()
		if got != tc.want {
			t.Errorf("Opcode(0x%02X).String() = %q, want %q", byte(tc.op), got, tc.want)
		}
	}
}

func TestOpcodeName(t *testing.T) {
	tests := []struct {
		op   Opcode
		want string
	}{
		{OP_ADD, "ADD"},
		{OP_STORE_3, "STORE_3"},
		{OP_RET, "RET"},
		{OP_DEF_GLOBAL, "DEF_GLOBAL"},
		{OP_ARRAY_NEW, "ARRAY_NEW"},
		{Opcode(0xFE), "UNKNOWN(0xFE)"},
	}

	for _, tc := range tests {
		got := tc.op.Name()
		if got != tc.want {
			t.Errorf("Opcode(0x%02X).Name() = %q, want %q", byte(tc.op), got, tc.want)
		}
	}
}

func TestOperandCount(t *testing.T) {
	tests := []struct {
		op   Opcode
		want int
	}{
		// Zero operands
		{OP_NOP, 0},
		{OP_POP, 0},
		{OP_DUP, 0},
		{OP_ADD, 0},
		{OP_EQ, 0},
		{OP_RET, 0},
		{OP_HALT, 0},
		{OP_PRINT, 0},

		// Two-byte operands
		{OP_PUSH, 2},
		{OP_LOAD_GLOBAL, 2},
		{OP_STORE_GLOBAL, 2},
		{OP_DEF_GLOBAL, 2},
		{OP_JMP, 2},
		{OP_JMP_IF, 2},
		{OP_JMP_IFN, 2},

		// One-byte operand
		{OP_CALL, 1},

		// Unknown opcode should default to 0.
		{Opcode(0xFF), 0},
	}

	for _, tc := range tests {
		got := tc.op.OperandCount()
		if got != tc.want {
			t.Errorf("Opcode(0x%02X).OperandCount() = %d, want %d", byte(tc.op), got, tc.want)
		}
	}
}

func TestOpcodeAlias(t *testing.T) {
	// OP_JMP_IF_NOT is an alias for OP_JMP_IFN.
	if OP_JMP_IF_NOT != OP_JMP_IFN {
		t.Errorf("OP_JMP_IF_NOT (0x%02X) != OP_JMP_IFN (0x%02X)",
			byte(OP_JMP_IF_NOT), byte(OP_JMP_IFN))
	}
}

func TestIsValid(t *testing.T) {
	if !OP_ADD.IsValid() {
		t.Error("OP_ADD should be valid")
	}
	if !OP_HALT.IsValid() {
		t.Error("OP_HALT should be valid")
	}
	if !OP_DEF_GLOBAL.IsValid() {
		t.Error("OP_DEF_GLOBAL should be valid")
	}
	if Opcode(0xFE).IsValid() {
		t.Error("Opcode(0xFE) should not be valid in Phase 0")
	}
}

func TestLoadLocal(t *testing.T) {
	cases := []struct {
		n    int
		want Opcode
	}{
		{0, OP_LOAD_0},
		{1, OP_LOAD_1},
		{2, OP_LOAD_2},
		{3, OP_LOAD_3},
	}
	for _, tc := range cases {
		got := LoadLocal(tc.n)
		if got != tc.want {
			t.Errorf("LoadLocal(%d) = 0x%02X, want 0x%02X", tc.n, byte(got), byte(tc.want))
		}
	}

	// Out-of-range should panic.
	defer func() {
		if r := recover(); r == nil {
			t.Error("LoadLocal(4) should have panicked")
		}
	}()
	LoadLocal(4)
}

func TestStoreLocal(t *testing.T) {
	cases := []struct {
		n    int
		want Opcode
	}{
		{0, OP_STORE_0},
		{1, OP_STORE_1},
		{2, OP_STORE_2},
		{3, OP_STORE_3},
	}
	for _, tc := range cases {
		got := StoreLocal(tc.n)
		if got != tc.want {
			t.Errorf("StoreLocal(%d) = 0x%02X, want 0x%02X", tc.n, byte(got), byte(tc.want))
		}
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("StoreLocal(4) should have panicked")
		}
	}()
	StoreLocal(4)
}

func TestByte(t *testing.T) {
	if OP_ADD.Byte() != 0x40 {
		t.Errorf("OP_ADD.Byte() = 0x%02X, want 0x40", OP_ADD.Byte())
	}
	if OP_HALT.Byte() != 0x91 {
		t.Errorf("OP_HALT.Byte() = 0x%02X, want 0x91", OP_HALT.Byte())
	}
}
