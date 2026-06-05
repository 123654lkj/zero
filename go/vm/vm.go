// Package vm implements the Zero language bytecode virtual machine.
//
// The VM is a stack-based interpreter that executes bytecode stored in Chunk
// objects. It supports global and local variables, arithmetic, comparisons,
// control flow, function calls (including built-in functions), and basic
// data structures (arrays and maps).
package vm

import (
	"fmt"

	"github.com/123654lkj/zero/go/chunk"
	"github.com/123654lkj/zero/go/opcode"
	"github.com/123654lkj/zero/go/value"
)

// ---------------------------------------------------------------------------
// Frame (call frame)
// ---------------------------------------------------------------------------

// Frame represents a single call frame on the VM's call stack.
type Frame struct {
	Chunk *chunk.Chunk // bytecode being executed
	Ip    int          // instruction pointer into Chunk.Code
	Base  int          // base pointer into VM stack (first slot for locals)
}

// ---------------------------------------------------------------------------
// VM
// ---------------------------------------------------------------------------

const (
	stackSize  = 4096
	frameLimit = 256
)

// VM is the Zero bytecode virtual machine.
type VM struct {
	stack      [stackSize]value.Value
	stackTop   int
	frames     [frameLimit]Frame
	frameCount int
	globals    map[string]value.Value
	natives    map[string]BuiltinFunc
}

// NewVM creates a new VM with built-in functions registered as globals.
func NewVM() *VM {
	vm := &VM{
		globals: make(map[string]value.Value),
		natives: make(map[string]BuiltinFunc),
	}
	vm.registerBuiltin("print", builtinPrint)
	vm.registerBuiltin("len", builtinLen)
	vm.registerBuiltin("type", builtinType)
	return vm
}

// registerBuiltin registers a native function so it can be called from
// bytecode.  The global's value is a string matching the name, which the
// CALL opcode uses to dispatch to the native implementation.
func (vm *VM) registerBuiltin(name string, fn BuiltinFunc) {
	vm.natives[name] = fn
	vm.globals[name] = value.StringValue(name)
}

// ---------------------------------------------------------------------------
// Stack helpers
// ---------------------------------------------------------------------------

// Push pushes a value onto the stack.
func (vm *VM) Push(v value.Value) {
	if vm.stackTop >= stackSize {
		panic("stack overflow")
	}
	vm.stack[vm.stackTop] = v
	vm.stackTop++
}

// Pop pops and returns the top value from the stack.
func (vm *VM) Pop() value.Value {
	if vm.stackTop == 0 {
		panic("stack underflow")
	}
	vm.stackTop--
	return vm.stack[vm.stackTop]
}

// Peek returns the top value without removing it.
func (vm *VM) Peek() value.Value {
	if vm.stackTop == 0 {
		panic("peek on empty stack")
	}
	return vm.stack[vm.stackTop-1]
}

// ---------------------------------------------------------------------------
// Execution
// ---------------------------------------------------------------------------

// RunChunk pushes a new frame for c, executes the bytecode, and returns
// the value left on top of the stack (or nil if the stack is empty).
func (vm *VM) RunChunk(c *chunk.Chunk) value.Value {
	vm.frameCount = 0
	vm.stackTop = 0
	vm.pushFrame(c, 0, 0)
	vm.Run()
	if vm.stackTop > 0 {
		return vm.stack[vm.stackTop-1]
	}
	return value.NilValue()
}

// Run executes bytecode until HALT or the chunk is exhausted.
func (vm *VM) Run() {
	for vm.frameCount > 0 {
		frame := &vm.frames[vm.frameCount-1]
		c := frame.Chunk

		if frame.Ip >= len(c.Code) {
			break
		}

		op := opcode.Opcode(c.Code[frame.Ip])
		frame.Ip++

		switch op {

		// ── Stack ──────────────────────────────────────────────────────

		case opcode.OP_NOP:
			// nothing

		case opcode.OP_PUSH:
			idx := vm.readWord16(frame)
			vm.Push(c.Constants[idx])

		case opcode.OP_POP:
			vm.Pop()

		case opcode.OP_DUP:
			vm.Push(vm.Peek())

		// ── Local variables ───────────────────────────────────────────

		case opcode.OP_LOAD_0:
			vm.Push(vm.stack[frame.Base+0])
		case opcode.OP_LOAD_1:
			vm.Push(vm.stack[frame.Base+1])
		case opcode.OP_LOAD_2:
			vm.Push(vm.stack[frame.Base+2])
		case opcode.OP_LOAD_3:
			vm.Push(vm.stack[frame.Base+3])

		case opcode.OP_STORE_0:
			vm.stack[frame.Base+0] = vm.Peek()
		case opcode.OP_STORE_1:
			vm.stack[frame.Base+1] = vm.Peek()
		case opcode.OP_STORE_2:
			vm.stack[frame.Base+2] = vm.Peek()
		case opcode.OP_STORE_3:
			vm.stack[frame.Base+3] = vm.Peek()

		// ── Global variables ──────────────────────────────────────────

		case opcode.OP_LOAD_GLOBAL:
			idx := vm.readWord16(frame)
			name := c.Names[idx]
			v, ok := vm.globals[name]
			if !ok {
				panic(fmt.Sprintf("undefined global: %s", name))
			}
			vm.Push(v)

		case opcode.OP_STORE_GLOBAL:
			idx := vm.readWord16(frame)
			name := c.Names[idx]
			vm.globals[name] = vm.Peek()

		case opcode.OP_DEF_GLOBAL:
			idx := vm.readWord16(frame)
			name := c.Names[idx]
			vm.globals[name] = vm.Pop()

		// ── Arithmetic ────────────────────────────────────────────────

		case opcode.OP_ADD:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(binaryAdd(a, b))
		case opcode.OP_SUB:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(binarySub(a, b))
		case opcode.OP_MUL:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(binaryMul(a, b))
		case opcode.OP_DIV:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(binaryDiv(a, b))
		case opcode.OP_MOD:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(binaryMod(a, b))
		case opcode.OP_NEG:
			vm.Push(unaryNeg(vm.Pop()))

		// ── Logic / comparisons ───────────────────────────────────────

		case opcode.OP_NOT:
			vm.Push(value.BoolValue(!vm.isTruthy(vm.Pop())))
		case opcode.OP_EQ:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(valuesEqual(a, b)))
		case opcode.OP_NEQ:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(!valuesEqual(a, b)))
		case opcode.OP_LT:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(compareValues(a, b) < 0))
		case opcode.OP_GT:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(compareValues(a, b) > 0))
		case opcode.OP_LTE:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(compareValues(a, b) <= 0))
		case opcode.OP_GTE:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(compareValues(a, b) >= 0))
		case opcode.OP_AND:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(vm.isTruthy(a) && vm.isTruthy(b)))
		case opcode.OP_OR:
			b, a := vm.Pop(), vm.Pop()
			vm.Push(value.BoolValue(vm.isTruthy(a) || vm.isTruthy(b)))

		// ── Control flow ──────────────────────────────────────────────

		case opcode.OP_JMP:
			offset := vm.readSignedWord16(frame)
			frame.Ip += int(offset)

		case opcode.OP_JMP_IF:
			offset := vm.readSignedWord16(frame)
			if vm.isTruthy(vm.Peek()) {
				frame.Ip += int(offset)
			}

		case opcode.OP_JMP_IFN:
			offset := vm.readSignedWord16(frame)
			if !vm.isTruthy(vm.Peek()) {
				frame.Ip += int(offset)
			}

		// ── Functions ─────────────────────────────────────────────────

		case opcode.OP_CALL:
			argCount := int(vm.readByte(frame))
			// Function value sits below the arguments on the stack.
			fnSlot := vm.stackTop - 1 - argCount
			fnVal := vm.stack[fnSlot]

			if fnVal.IsString() {
				name := fnVal.AsString()
				if fn, ok := vm.natives[name]; ok {
					// Collect arguments.
					args := make([]value.Value, argCount)
					copy(args, vm.stack[vm.stackTop-argCount:vm.stackTop])
					// Remove function + arguments from stack.
					vm.stackTop -= argCount + 1
					result := fn(args)
					vm.Push(result)
				} else if meta, ok := vm.findFunction(frame.Chunk, name); ok {
					// User-defined function: shift args down to overwrite function value.
					// Stack before: [..., F, A0, A1, ..., A_{n-1}]
					// Stack after:  [..., A0, A1, ..., A_{n-1}]
					// Base points to A0 so LOAD_0 reads first arg.
					for i := 0; i < argCount; i++ {
						vm.stack[fnSlot+i] = vm.stack[fnSlot+1+i]
					}
					vm.stackTop = fnSlot + argCount
					if vm.frameCount >= 256 {
						panic("stack overflow: too many call frames")
					}
					vm.frames[vm.frameCount] = Frame{
						Chunk: frame.Chunk,
						Ip:    meta.Start,
						Base:  fnSlot,
					}
					vm.frameCount++
					frame = &vm.frames[vm.frameCount-1]
					continue
				} else {
					panic(fmt.Sprintf("undefined function: %s", name))
				}
			} else {
				panic("cannot call non-function value")
			}

		case opcode.OP_RET:
			result := vm.Pop()
			vm.frameCount--
			vm.stackTop = vm.frames[vm.frameCount].Base
			vm.Push(result)

		// ── Special ───────────────────────────────────────────────────

		case opcode.OP_PRINT:
			v := vm.Pop()
			fmt.Println(formatValue(v))

		case opcode.OP_HALT:
			return

		// ── Data structures ──────────────────────────────────────────

		case opcode.OP_ARRAY_NEW:
			countVal := vm.Pop()
			if !countVal.IsInt() {
				panic("array_new: count must be an int")
			}
			count := int(countVal.AsInt())
			arr := make([]value.Value, count)
			for i := count - 1; i >= 0; i-- {
				arr[i] = vm.Pop()
			}
			vm.Push(value.ArrayValue(arr))

		case opcode.OP_ARRAY_GET:
			idx, arr := vm.Pop(), vm.Pop()
			if !arr.IsArray() || !idx.IsInt() {
				panic("array_get: need (array, int)")
			}
			a := arr.AsArray()
			i := int(idx.AsInt())
			if i < 0 || i >= len(a) {
				panic(fmt.Sprintf("array index out of bounds: %d", i))
			}
			vm.Push(a[i])

		case opcode.OP_ARRAY_SET:
			val, idx, arr := vm.Pop(), vm.Pop(), vm.Pop()
			if !arr.IsArray() || !idx.IsInt() {
				panic("array_set: need (array, int, value)")
			}
			a := arr.AsArray()
			i := int(idx.AsInt())
			if i < 0 || i >= len(a) {
				panic(fmt.Sprintf("array index out of bounds: %d", i))
			}
			a[i] = val
			vm.Push(val)

		case opcode.OP_MAP_NEW:
			countVal := vm.Pop()
			if !countVal.IsInt() {
				panic("map_new: count must be an int")
			}
			count := int(countVal.AsInt())
			m := make(map[string]value.Value, count)
			for i := 0; i < count; i++ {
				v := vm.Pop()
				k := vm.Pop()
				if !k.IsString() {
					panic("map_new: key must be a string")
				}
				m[k.AsString()] = v
			}
			vm.Push(value.MapValue(m))

		case opcode.OP_MAP_GET:
			key, m := vm.Pop(), vm.Pop()
			if !m.IsMap() || !key.IsString() {
				panic("map_get: need (map, string)")
			}
			v, ok := m.AsMap()[key.AsString()]
			if !ok {
				vm.Push(value.NilValue())
			} else {
				vm.Push(v)
			}

		case opcode.OP_MAP_SET:
			val, key, m := vm.Pop(), vm.Pop(), vm.Pop()
			if !m.IsMap() || !key.IsString() {
				panic("map_set: need (map, string, value)")
			}
			m.AsMap()[key.AsString()] = val
			vm.Push(val)

		default:
			panic(fmt.Sprintf("unknown opcode: 0x%02X", byte(op)))
		}
	}
}

// ---------------------------------------------------------------------------
// Frame helpers
// ---------------------------------------------------------------------------

func (vm *VM) pushFrame(c *chunk.Chunk, ip, base int) {
	if vm.frameCount >= frameLimit {
		panic("call stack overflow")
	}
	vm.frames[vm.frameCount] = Frame{Chunk: c, Ip: ip, Base: base}
	vm.frameCount++
}

// ---------------------------------------------------------------------------
// Bytecode readers
// ---------------------------------------------------------------------------

func (vm *VM) readByte(frame *Frame) byte {
	b := frame.Chunk.Code[frame.Ip]
	frame.Ip++
	return b
}

// findFunction looks up a user-defined function by name in the chunk's Functions list.
func (vm *VM) findFunction(c *chunk.Chunk, name string) (chunk.FunctionMeta, bool) {
	for _, meta := range c.Functions {
		if meta.Name == name {
			return meta, true
		}
	}
	return chunk.FunctionMeta{}, false
}

func (vm *VM) readWord16(frame *Frame) uint16 {
	hi := frame.Chunk.Code[frame.Ip]
	lo := frame.Chunk.Code[frame.Ip+1]
	frame.Ip += 2
	return uint16(hi)<<8 | uint16(lo)
}

func (vm *VM) readSignedWord16(frame *Frame) int16 {
	return int16(vm.readWord16(frame))
}

// ---------------------------------------------------------------------------
// Truthiness
// ---------------------------------------------------------------------------

func (vm *VM) isTruthy(v value.Value) bool {
	if v.IsNil() {
		return false
	}
	if v.IsBool() {
		return v.AsBool()
	}
	if v.IsInt() {
		return v.AsInt() != 0
	}
	if v.IsFloat() {
		return v.AsFloat() != 0
	}
	if v.IsString() {
		return v.AsString() != ""
	}
	// Arrays, maps, closures, etc. are always truthy.
	return true
}

// ---------------------------------------------------------------------------
// Arithmetic helpers
// ---------------------------------------------------------------------------

func numParts(a, b value.Value) (af, bf float64, ok bool) {
	if a.IsInt() && b.IsInt() {
		return 0, 0, false // both ints — special fast path
	}
	if a.IsInt() {
		af = float64(a.AsInt())
	} else if a.IsFloat() {
		af = a.AsFloat()
	} else {
		return 0, 0, false
	}
	if b.IsInt() {
		bf = float64(b.AsInt())
	} else if b.IsFloat() {
		bf = b.AsFloat()
	} else {
		return 0, 0, false
	}
	return af, bf, true
}

func binaryAdd(a, b value.Value) value.Value {
	if a.IsString() && b.IsString() {
		return value.StringValue(a.AsString() + b.AsString())
	}
	if a.IsInt() && b.IsInt() {
		return value.IntValue(a.AsInt() + b.AsInt())
	}
	af, bf, ok := numParts(a, b)
	if !ok {
		panic(fmt.Sprintf("add: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
	}
	return value.FloatValue(af + bf)
}

func binarySub(a, b value.Value) value.Value {
	if a.IsInt() && b.IsInt() {
		return value.IntValue(a.AsInt() - b.AsInt())
	}
	af, bf, ok := numParts(a, b)
	if !ok {
		panic(fmt.Sprintf("sub: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
	}
	return value.FloatValue(af - bf)
}

func binaryMul(a, b value.Value) value.Value {
	if a.IsInt() && b.IsInt() {
		return value.IntValue(a.AsInt() * b.AsInt())
	}
	af, bf, ok := numParts(a, b)
	if !ok {
		panic(fmt.Sprintf("mul: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
	}
	return value.FloatValue(af * bf)
}

func binaryDiv(a, b value.Value) value.Value {
	if a.IsInt() && b.IsInt() {
		if b.AsInt() == 0 {
			panic("division by zero")
		}
		return value.IntValue(a.AsInt() / b.AsInt())
	}
	af, bf, ok := numParts(a, b)
	if !ok {
		panic(fmt.Sprintf("div: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
	}
	return value.FloatValue(af / bf)
}

func binaryMod(a, b value.Value) value.Value {
	if a.IsInt() && b.IsInt() {
		if b.AsInt() == 0 {
			panic("modulo by zero")
		}
		return value.IntValue(a.AsInt() % b.AsInt())
	}
	panic(fmt.Sprintf("mod: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
}

func unaryNeg(v value.Value) value.Value {
	if v.IsInt() {
		return value.IntValue(-v.AsInt())
	}
	if v.IsFloat() {
		return value.FloatValue(-v.AsFloat())
	}
	panic(fmt.Sprintf("neg: unsupported operand %s", v.ValueType()))
}

// ---------------------------------------------------------------------------
// Comparison helpers
// ---------------------------------------------------------------------------

func valuesEqual(a, b value.Value) bool {
	if a.ValueType() != b.ValueType() {
		return false
	}
	switch {
	case a.IsNil():
		return true
	case a.IsBool():
		return a.AsBool() == b.AsBool()
	case a.IsInt():
		return a.AsInt() == b.AsInt()
	case a.IsFloat():
		return a.AsFloat() == b.AsFloat()
	case a.IsString():
		return a.AsString() == b.AsString()
	default:
		return false
	}
}

func compareValues(a, b value.Value) int {
	// int vs int
	if a.IsInt() && b.IsInt() {
		ai, bi := a.AsInt(), b.AsInt()
		switch {
		case ai < bi:
			return -1
		case ai > bi:
			return 1
		default:
			return 0
		}
	}
	// float vs float
	if a.IsFloat() && b.IsFloat() {
		af, bf := a.AsFloat(), b.AsFloat()
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	// int vs float (promote)
	if a.IsInt() && b.IsFloat() {
		af, bf := float64(a.AsInt()), b.AsFloat()
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	if a.IsFloat() && b.IsInt() {
		af, bf := a.AsFloat(), float64(b.AsInt())
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	// string vs string
	if a.IsString() && b.IsString() {
		sa, sb := a.AsString(), b.AsString()
		switch {
		case sa < sb:
			return -1
		case sa > sb:
			return 1
		default:
			return 0
		}
	}
	panic(fmt.Sprintf("compare: unsupported operands %s and %s", a.ValueType(), b.ValueType()))
}
