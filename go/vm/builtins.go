// Package vm implements the Zero language bytecode virtual machine.
package vm

import (
	"fmt"
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
