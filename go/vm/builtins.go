// Package vm implements the Zero language bytecode virtual machine.
package vm

import (
	"fmt"
	"os"
	"strings"

	"github.com/123654lkj/zero/go/value"
)

// BuiltinFunc is the signature for a native (built-in) function.
type BuiltinFunc func(args []value.Value) value.Value

// ---------------------------------------------------------------------------
// Built-in functions
// ---------------------------------------------------------------------------

// builtinPrint prints all arguments space-separated to stdout.
func builtinPrint(args []value.Value) value.Value {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = formatValue(a)
	}
	fmt.Println(strings.Join(parts, " "))
	return value.NilValue()
}

// builtinLen returns the length of a string, array, or map.
func builtinLen(args []value.Value) value.Value {
	if len(args) != 1 {
		panic("len: expected exactly 1 argument")
	}
	a := args[0]
	switch {
	case a.IsString():
		return value.IntValue(int64(len(a.AsString())))
	case a.IsArray():
		return value.IntValue(int64(len(a.AsArray())))
	case a.IsMap():
		return value.IntValue(int64(len(a.AsMap())))
	default:
		panic("len: argument must be string, array, or map")
	}
}

// builtinType returns the type name of its argument as a string.
func builtinType(args []value.Value) value.Value {
	if len(args) != 1 {
		panic("type: expected exactly 1 argument")
	}
	return value.StringValue(args[0].ValueType().String())
}

// builtinCharAt returns the character at index i in string s.
func builtinCharAt(args []value.Value) value.Value {
	if len(args) != 2 {
		panic("char_at: expected exactly 2 arguments")
	}
	s := args[0]
	idx := args[1]
	if !s.IsString() || !idx.IsInt() {
		panic("char_at: need (string, int)")
	}
	str := s.AsString()
	i := int(idx.AsInt())
	if i < 0 || i >= len(str) {
		panic(fmt.Sprintf("char_at: index out of bounds: %d", i))
	}
	return value.StringValue(string(str[i]))
}

// builtinReadFile reads a file and returns its content as a string.
func builtinReadFile(args []value.Value) value.Value {
	if len(args) != 1 {
		panic("read_file: expected exactly 1 argument")
	}
	path := args[0]
	if !path.IsString() {
		panic("read_file: argument must be a string")
	}
	data, err := os.ReadFile(path.AsString())
	if err != nil {
		panic(fmt.Sprintf("read_file: %v", err))
	}
	return value.StringValue(string(data))
}

// builtinWriteFile writes content to a file.
func builtinWriteFile(args []value.Value) value.Value {
	if len(args) != 2 {
		panic("write_file: expected exactly 2 arguments")
	}
	path := args[0]
	content := args[1]
	if !path.IsString() || !content.IsString() {
		panic("write_file: need (string, string)")
	}
	err := os.WriteFile(path.AsString(), []byte(content.AsString()), 0644)
	if err != nil {
		panic(fmt.Sprintf("write_file: %v", err))
	}
	return value.NilValue()
}

// ---------------------------------------------------------------------------
// Formatting helpers
// ---------------------------------------------------------------------------

// formatValue produces a human-readable representation of a Value for printing.
func formatValue(v value.Value) string {
	switch {
	case v.IsNil():
		return "nil"
	case v.IsBool():
		if v.AsBool() {
			return "true"
		}
		return "false"
	case v.IsInt():
		return fmt.Sprintf("%d", v.AsInt())
	case v.IsFloat():
		return fmt.Sprintf("%g", v.AsFloat())
	case v.IsString():
		return v.AsString()
	case v.IsArray():
		arr := v.AsArray()
		parts := make([]string, len(arr))
		for i, elem := range arr {
			parts[i] = formatValue(elem)
		}
		return "[" + strings.Join(parts, " ") + "]"
	case v.IsMap():
		m := v.AsMap()
		parts := make([]string, 0, len(m))
		for k, val := range m {
			parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(val)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return v.String()
	}
}
