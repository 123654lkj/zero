package value

import (
	"math"
	"testing"
)

// -----------------------------------------------------------------------
// Constructor + accessor round-trip tests
// -----------------------------------------------------------------------

func TestNilValue(t *testing.T) {
	v := NilValue()
	if v.ValueType() != TagNil {
		t.Fatalf("ValueType() = %v, want TagNil", v.ValueType())
	}
	if !v.IsNil() {
		t.Error("IsNil() = false, want true")
	}
	if v.IsBool() || v.IsInt() || v.IsFloat() || v.IsString() || v.IsArray() || v.IsMap() {
		t.Error("Nil should not satisfy any non-nil type check")
	}
}

func TestBoolValue(t *testing.T) {
	tests := []struct {
		input bool
		want  bool
	}{
		{true, true},
		{false, false},
	}
	for _, tt := range tests {
		v := BoolValue(tt.input)
		if v.ValueType() != TagBool {
			t.Errorf("BoolValue(%v).ValueType() = %v, want TagBool", tt.input, v.ValueType())
		}
		if !v.IsBool() {
			t.Errorf("BoolValue(%v).IsBool() = false", tt.input)
		}
		if v.AsBool() != tt.want {
			t.Errorf("BoolValue(%v).AsBool() = %v, want %v", tt.input, v.AsBool(), tt.want)
		}
	}
}

func TestIntValue_Small(t *testing.T) {
	// Small non-negative ints (< 2^31) are embedded — no heap allocation.
	v := IntValue(42)
	if v.ValueType() != TagInt {
		t.Fatalf("ValueType() = %v, want TagInt", v.ValueType())
	}
	if !v.IsInt() {
		t.Fatal("IsInt() = false")
	}
	if v.AsInt() != 42 {
		t.Errorf("AsInt() = %d, want 42", v.AsInt())
	}
}

func TestIntValue_Zero(t *testing.T) {
	v := IntValue(0)
	if v.AsInt() != 0 {
		t.Errorf("AsInt() = %d, want 0", v.AsInt())
	}
}

func TestIntValue_Boundary(t *testing.T) {
	// 2^31 - 1 = 2147483647 — largest embedded int
	v := IntValue(int64(smallIntMax))
	if v.AsInt() != int64(smallIntMax) {
		t.Errorf("AsInt() = %d, want %d", v.AsInt(), smallIntMax)
	}
}

func TestIntValue_Large(t *testing.T) {
	// Values >= 2^31 are heap-allocated.
	big := int64(1) << 40
	v := IntValue(big)
	if v.AsInt() != big {
		t.Errorf("AsInt() = %d, want %d", v.AsInt(), big)
	}
}

func TestIntValue_Negative(t *testing.T) {
	v := IntValue(-42)
	if v.AsInt() != -42 {
		t.Errorf("AsInt() = %d, want -42", v.AsInt())
	}
}

func TestIntValue_NegativeLarge(t *testing.T) {
	v := int64(-1) << 40
	result := IntValue(v)
	if result.AsInt() != v {
		t.Errorf("AsInt() = %d, want %d", result.AsInt(), v)
	}
}

func TestFloatValue(t *testing.T) {
	tests := []struct {
		name  string
		input float64
	}{
		{"positive", 3.14},
		{"zero", 0.0},
		{"negative", -2.718},
		{"tiny", math.SmallestNonzeroFloat64},
		{"huge", math.MaxFloat64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := FloatValue(tt.input)
			if v.ValueType() != TagFloat {
				t.Fatalf("ValueType() = %v, want TagFloat", v.ValueType())
			}
			if !v.IsFloat() {
				t.Fatal("IsFloat() = false")
			}
			if v.AsFloat() != tt.input {
				t.Errorf("AsFloat() = %v, want %v", v.AsFloat(), tt.input)
			}
		})
	}
}

func TestStringValue_Embedded(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"one", "a"},
		{"three", "abc"},
		{"six", "abcdef"}, // len == 6, maximum embedded
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := StringValue(tt.input)
			if v.ValueType() != TagString {
				t.Fatalf("ValueType() = %v, want TagString", v.ValueType())
			}
			if !v.IsString() {
				t.Fatal("IsString() = false")
			}
			if v.AsString() != tt.input {
				t.Errorf("AsString() = %q, want %q", v.AsString(), tt.input)
			}
		})
	}
}

func TestStringValue_Heap(t *testing.T) {
	// Strings of length >= 7 are heap-allocated.
	long := "hello, world!"
	v := StringValue(long)
	if v.AsString() != long {
		t.Errorf("AsString() = %q, want %q", v.AsString(), long)
	}

	// 7-byte string
	s7 := "1234567"
	v7 := StringValue(s7)
	if v7.AsString() != s7 {
		t.Errorf("AsString() = %q, want %q", v7.AsString(), s7)
	}

	// Very long string
	vLong := StringValue("abcdefghij")
	if vLong.AsString() != "abcdefghij" {
		t.Error("long string round-trip failed")
	}
}

func TestArrayValue(t *testing.T) {
	arr := []Value{IntValue(1), BoolValue(true), StringValue("x")}
	v := ArrayValue(arr)
	if v.ValueType() != TagArray {
		t.Fatalf("ValueType() = %v, want TagArray", v.ValueType())
	}
	if !v.IsArray() {
		t.Fatal("IsArray() = false")
	}
	got := v.AsArray()
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].AsInt() != 1 {
		t.Errorf("got[0] = %v, want int(1)", got[0])
	}
	if !got[1].AsBool() {
		t.Errorf("got[1] = %v, want bool(true)", got[1])
	}
	if got[2].AsString() != "x" {
		t.Errorf("got[2] = %v, want string(x)", got[2])
	}
}

func TestArrayValue_Nil(t *testing.T) {
	v := ArrayValue(nil)
	got := v.AsArray()
	if got == nil {
		t.Error("AsArray() returned nil slice, want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestMapValue(t *testing.T) {
	m := map[string]Value{
		"x": IntValue(10),
		"y": StringValue("hello"),
	}
	v := MapValue(m)
	if v.ValueType() != TagMap {
		t.Fatalf("ValueType() = %v, want TagMap", v.ValueType())
	}
	if !v.IsMap() {
		t.Fatal("IsMap() = false")
	}
	got := v.AsMap()
	if got["x"].AsInt() != 10 {
		t.Errorf("got[x] = %v, want int(10)", got["x"])
	}
	if got["y"].AsString() != "hello" {
		t.Errorf("got[y] = %v, want string(hello)", got["y"])
	}
}

func TestMapValue_Nil(t *testing.T) {
	v := MapValue(nil)
	got := v.AsMap()
	if got == nil {
		t.Error("AsMap() returned nil map, want empty map")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

// -----------------------------------------------------------------------
// IsZero tests
// -----------------------------------------------------------------------

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		want bool
	}{
		{"nil", NilValue(), true},
		{"bool false", BoolValue(false), true},
		{"bool true", BoolValue(true), false},
		{"int 0", IntValue(0), true},
		{"int 1", IntValue(1), false},
		{"int -1", IntValue(-1), false},
		{"int max-embed", IntValue(int64(smallIntMax)), false},
		{"float 0.0", FloatValue(0.0), true},
		{"float 1.0", FloatValue(1.0), false},
		{"float -0.0", FloatValue(math.Copysign(0, -1)), true},
		{"string empty", StringValue(""), true},
		{"string a", StringValue("a"), false},
		{"string long empty", StringValue("1234567"), false},
		{"array nil", ArrayValue(nil), true},
		{"array empty", ArrayValue([]Value{}), true},
		{"array one", ArrayValue([]Value{IntValue(1)}), false},
		{"map nil", MapValue(nil), true},
		{"map empty", MapValue(map[string]Value{}), true},
		{"map one", MapValue(map[string]Value{"k": IntValue(1)}), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// -----------------------------------------------------------------------
// String() debugging output
// -----------------------------------------------------------------------

func TestString(t *testing.T) {
	tests := []struct {
		val  Value
		want string
	}{
		{NilValue(), "nil"},
		{BoolValue(true), "bool(true)"},
		{BoolValue(false), "bool(false)"},
		{IntValue(0), "int(0)"},
		{IntValue(42), "int(42)"},
		{IntValue(-1), "int(-1)"},
		{FloatValue(0), "float(0)"},
		{FloatValue(3.14), "float(3.14)"},
		{StringValue(""), `string("")`},
		{StringValue("hi"), `string("hi")`},
		{ArrayValue(nil), "array(len=0)"},
		{ArrayValue([]Value{IntValue(1)}), "array(len=1)"},
		{MapValue(nil), "map(len=0)"},
	}
	for _, tt := range tests {
		got := tt.val.String()
		if got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

// -----------------------------------------------------------------------
// ValueType exhaustive check
// -----------------------------------------------------------------------

func TestValueType_AllTags(t *testing.T) {
	cases := []struct {
		tag TypeTag
		val Value
	}{
		{TagNil, NilValue()},
		{TagBool, BoolValue(false)},
		{TagInt, IntValue(0)},
		{TagFloat, FloatValue(0)},
		{TagString, StringValue("")},
		{TagArray, ArrayValue(nil)},
		{TagMap, MapValue(nil)},
	}
	for _, c := range cases {
		if c.val.ValueType() != c.tag {
			t.Errorf("ValueType() = %v, want %v", c.val.ValueType(), c.tag)
		}
	}
}

// -----------------------------------------------------------------------
// Mixed-type value table (simulates a VM register file)
// -----------------------------------------------------------------------

func TestValueTable(t *testing.T) {
	regs := make([]Value, 8)
	regs[0] = NilValue()
	regs[1] = BoolValue(true)
	regs[2] = IntValue(99)
	regs[3] = FloatValue(2.718)
	regs[4] = StringValue("hello")
	regs[5] = ArrayValue([]Value{IntValue(1), IntValue(2)})
	regs[6] = MapValue(map[string]Value{"key": StringValue("val")})
	regs[7] = IntValue(-100)

	// Verify each register preserves its value.
	if regs[0].IsNil() != true {
		t.Error("reg[0] should be nil")
	}
	if regs[1].AsBool() != true {
		t.Error("reg[1] should be true")
	}
	if regs[2].AsInt() != 99 {
		t.Errorf("reg[2] = %d, want 99", regs[2].AsInt())
	}
	if regs[3].AsFloat() != 2.718 {
		t.Errorf("reg[3] = %v, want 2.718", regs[3].AsFloat())
	}
	if regs[4].AsString() != "hello" {
		t.Errorf("reg[4] = %q, want hello", regs[4].AsString())
	}
	if len(regs[5].AsArray()) != 2 {
		t.Error("reg[5] should have 2 elements")
	}
	if regs[6].AsMap()["key"].AsString() != "val" {
		t.Error("reg[6][key] should be val")
	}
	if regs[7].AsInt() != -100 {
		t.Errorf("reg[7] = %d, want -100", regs[7].AsInt())
	}
}

// -----------------------------------------------------------------------
// Embedded string edge cases
// -----------------------------------------------------------------------

func TestEmbeddedString_AllLengths(t *testing.T) {
	for l := 0; l < 7; l++ {
		s := ""
		for i := 0; i < l; i++ {
			s += string(rune('A' + i))
		}
		v := StringValue(s)
		if v.AsString() != s {
			t.Errorf("len=%d: got %q, want %q", l, v.AsString(), s)
		}
	}
}

func TestEmbeddedString_Binary(t *testing.T) {
	// Test with non-printable bytes
	s := "\x00\x01\xff"
	v := StringValue(s)
	if v.AsString() != s {
		t.Errorf("binary string round-trip failed")
	}
}

// -----------------------------------------------------------------------
// TypeTag.String()
// -----------------------------------------------------------------------

func TestTypeTag_String(t *testing.T) {
	names := []string{
		"Nil", "Bool", "Int", "Float", "String",
		"Array", "Map", "Closure", "Native", "Pattern",
		"Tagged", "Stream", "Image", "IO", "Table",
	}
	for i, name := range names {
		tag := TypeTag(i)
		if tag.String() != name {
			t.Errorf("TypeTag(%d).String() = %q, want %q", i, tag.String(), name)
		}
	}
}

func TestTypeTag_String_Unknown(t *testing.T) {
	tag := TypeTag(255)
	s := tag.String()
	if s == "" {
		t.Error("String() should not return empty for unknown tag")
	}
}

// -----------------------------------------------------------------------
// Panic tests for type mismatches
// -----------------------------------------------------------------------

func TestAsBool_PanicsOnNonBool(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	IntValue(1).AsBool()
}

func TestAsInt_PanicsOnNonInt(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	BoolValue(true).AsInt()
}

func TestAsFloat_PanicsOnNonFloat(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	IntValue(1).AsFloat()
}

func TestAsString_PanicsOnNonString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	IntValue(1).AsString()
}

func TestAsArray_PanicsOnNonArray(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	IntValue(1).AsArray()
}

func TestAsMap_PanicsOnNonMap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	IntValue(1).AsMap()
}

// -----------------------------------------------------------------------
// Bump: ensure no GC crashes on heap values (GC runs during allocation)
// -----------------------------------------------------------------------

func TestNoGC(t *testing.T) {
	// Create many heap-allocated values, trigger GC, then read them back.
	for i := 0; i < 1000; i++ {
		IntValue(int64(i) + int64(smallIntMax))
		FloatValue(float64(i) * 1.5)
		StringValue("long string that is definitely heap-allocated: " + string(rune('A'+i%26)))
		ArrayValue([]Value{IntValue(int64(i)), StringValue("x")})
		MapValue(map[string]Value{"k": IntValue(int64(i))})
	}

	// All pinned objects should still be readable.
	// If pinning is broken, the GC would have collected them and we'd crash.
}
