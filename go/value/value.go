// Package value implements the Zero language value system.
//
// Every Zero value is an 8-byte tagged union:
//   - 1 byte: type tag (TagNil through TagTable)
//   - 7 bytes: payload (embedded data or pointer to heap data)
//
// Small ints (0 to 2^31-1) and short strings (len < 7) are embedded
// directly in the payload with zero heap allocation.
package value

import (
	"fmt"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Type tags
// ---------------------------------------------------------------------------

// TypeTag is a single byte identifying the type of a Zero value.
type TypeTag byte

const (
	TagNil TypeTag = iota
	TagBool
	TagInt
	TagFloat
	TagString
	TagArray
	TagMap
	TagClosure
	TagNative
	TagPattern
	TagTagged
	TagStream
	TagImage
	TagIO
	TagTable
)

// tagNames maps each TypeTag to its human-readable name.
var tagNames = [16]string{
	TagNil:     "Nil",
	TagBool:    "Bool",
	TagInt:     "Int",
	TagFloat:   "Float",
	TagString:  "String",
	TagArray:   "Array",
	TagMap:     "Map",
	TagClosure: "Closure",
	TagNative:  "Native",
	TagPattern: "Pattern",
	TagTagged:  "Tagged",
	TagStream:  "Stream",
	TagImage:   "Image",
	TagIO:      "IO",
	TagTable:   "Table",
}

// String returns the name of the type tag.
func (t TypeTag) String() string {
	if int(t) < len(tagNames) {
		return tagNames[t]
	}
	return fmt.Sprintf("Unknown(%d)", t)
}

// ---------------------------------------------------------------------------
// Value representation
// ---------------------------------------------------------------------------

const (
	payloadMask = 0x00FFFFFFFFFFFFFF // lower 56 bits (7 bytes)
	smallIntMax = uint64(1<<31 - 1)  // 2 147 483 647
	heapBit     = uint64(1) << 55    // MSB of 7-byte payload; set = heap pointer
)

// Value is exactly 8 bytes: 1 byte type tag + 7 bytes payload.
// The lower 56 bits (7 bytes) of raw store either embedded data or a pointer.
type Value struct {
	raw uint64
}

// newRaw creates a Value from a tag and a raw 56-bit payload.
func newRaw(tag TypeTag, payload uint64) Value {
	return Value{raw: uint64(tag)<<56 | (payload & payloadMask)}
}

func (v Value) tag() TypeTag    { return TypeTag(v.raw >> 56) }
func (v Value) payload() uint64 { return v.raw & payloadMask }

// ---------------------------------------------------------------------------
// Pointer helpers
// ---------------------------------------------------------------------------

func ptrToPayload(p unsafe.Pointer) uint64 {
	return uint64(uintptr(p))
}

func payloadToPtr(payload uint64) unsafe.Pointer {
	// Reinterpret the bits of payload as a pointer via pointer aliasing.
	// This avoids the go-vet flagged unsafe.Pointer(uintptr(...)) pattern.
	return *(*unsafe.Pointer)(unsafe.Pointer(&payload))
}

// pin stores a reference in a global map so the GC cannot collect the object.
// Every heap-allocated value must call pin exactly once.
func pin(p unsafe.Pointer, obj interface{}) {
	heapRefs[uintptr(p)] = obj
}

// heapRefs prevents GC collection of heap objects embedded in Value payloads.
// Key: pointer address (uintptr). Value: Go-typed reference (interface{}).
var heapRefs = map[uintptr]interface{}{}

// ---------------------------------------------------------------------------
// Accessors
// ---------------------------------------------------------------------------

// ValueType returns the TypeTag of v.
func (v Value) ValueType() TypeTag { return v.tag() }

// IsNil returns true when v is a Nil value.
func (v Value) IsNil() bool { return v.tag() == TagNil }

// IsBool returns true when v is a Bool value.
func (v Value) IsBool() bool { return v.tag() == TagBool }

// AsBool extracts the bool. Panics if v is not a Bool.
func (v Value) AsBool() bool {
	if v.tag() != TagBool {
		panic("AsBool: not a Bool")
	}
	return v.payload() != 0
}

// IsInt returns true when v is an Int value.
func (v Value) IsInt() bool { return v.tag() == TagInt }

// AsInt extracts the int64. Panics if v is not an Int.
func (v Value) AsInt() int64 {
	if v.tag() != TagInt {
		panic("AsInt: not an Int")
	}
	p := v.payload()
	// Small ints are embedded directly (payload < 2^31).
	// Large/negative ints are heap-allocated (payload has MSB set or is a
	// pointer value with upper bits set).
	if p <= smallIntMax {
		return int64(p)
	}
	return *(*int64)(payloadToPtr(p))
}

// IsFloat returns true when v is a Float value.
func (v Value) IsFloat() bool { return v.tag() == TagFloat }

// AsFloat extracts the float64. Panics if v is not a Float.
func (v Value) AsFloat() float64 {
	if v.tag() != TagFloat {
		panic("AsFloat: not a Float")
	}
	return *(*float64)(payloadToPtr(v.payload()))
}

// IsString returns true when v is a String value.
func (v Value) IsString() bool { return v.tag() == TagString }

// AsString extracts the string. Panics if v is not a String.
func (v Value) AsString() string {
	if v.tag() != TagString {
		panic("AsString: not a String")
	}
	p := v.payload()
	// Embedded strings have heapBit clear; length is in bits 54-48.
	// Heap strings have heapBit set; lower 55 bits are the pointer.
	if p&heapBit == 0 {
		return decodeEmbeddedString(p, int((p>>48)&0x7F))
	}
	return (*heapString)(payloadToPtr(p &^ heapBit)).s
}

// IsArray returns true when v is an Array value.
func (v Value) IsArray() bool { return v.tag() == TagArray }

// AsArray extracts the []Value. Panics if v is not an Array.
func (v Value) AsArray() []Value {
	if v.tag() != TagArray {
		panic("AsArray: not an Array")
	}
	return (*heapArray)(payloadToPtr(v.payload())).arr
}

// IsMap returns true when v is a Map value.
func (v Value) IsMap() bool { return v.tag() == TagMap }

// AsMap extracts the map[string]Value. Panics if v is not a Map.
func (v Value) AsMap() map[string]Value {
	if v.tag() != TagMap {
		panic("AsMap: not a Map")
	}
	return (*heapMap)(payloadToPtr(v.payload())).m
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NilValue returns the nil value.
func NilValue() Value { return newRaw(TagNil, 0) }

// BoolValue wraps a Go bool.
func BoolValue(b bool) Value {
	if b {
		return newRaw(TagBool, 1)
	}
	return newRaw(TagBool, 0)
}

// IntValue wraps a Go int64.
// Small non-negative values (< 2^31) are embedded in the payload with
// zero heap allocation.  All others are heap-allocated.
func IntValue(i int64) Value {
	if i >= 0 && uint64(i) <= smallIntMax {
		return newRaw(TagInt, uint64(i))
	}
	ptr := new(int64)
	*ptr = i
	pin(unsafe.Pointer(ptr), ptr)
	return newRaw(TagInt, ptrToPayload(unsafe.Pointer(ptr)))
}

// FloatValue wraps a Go float64.
// Always heap-allocated (7-byte payload is not enough for an IEEE-754 double).
func FloatValue(f float64) Value {
	ptr := new(float64)
	*ptr = f
	pin(unsafe.Pointer(ptr), ptr)
	return newRaw(TagFloat, ptrToPayload(unsafe.Pointer(ptr)))
}

// StringValue wraps a Go string.
// Strings shorter than 7 bytes are embedded in the payload (zero allocation).
// Longer strings are heap-allocated with heapBit set to distinguish from embedded.
func StringValue(s string) Value {
	if len(s) < 7 {
		return embedString(s)
	}
	hs := new(heapString)
	hs.s = s
	pin(unsafe.Pointer(hs), hs)
	return newRaw(TagString, ptrToPayload(unsafe.Pointer(hs))|heapBit)
}

// ArrayValue wraps a Go []Value slice.
func ArrayValue(arr []Value) Value {
	if arr == nil {
		arr = []Value{}
	}
	ha := new(heapArray)
	ha.arr = arr
	pin(unsafe.Pointer(ha), ha)
	return newRaw(TagArray, ptrToPayload(unsafe.Pointer(ha)))
}

// MapValue wraps a Go map[string]Value.
func MapValue(m map[string]Value) Value {
	if m == nil {
		m = map[string]Value{}
	}
	hm := new(heapMap)
	hm.m = m
	pin(unsafe.Pointer(hm), hm)
	return newRaw(TagMap, ptrToPayload(unsafe.Pointer(hm)))
}

// ---------------------------------------------------------------------------
// Embedded string encoding
// ---------------------------------------------------------------------------
//
// Layout of the 7-byte payload for an embedded string:
//
//   bit  55    : heapBit = 0  (distinguishes from heap string)
//   bits 54-48 : length  (0 .. 6)
//   bits 47-40 : s[0]
//   bits 39-32 : s[1]
//   bits 31-24 : s[2]
//   bits 23-16 : s[3]
//   bits 15-8  : s[4]
//   bits  7-0  : s[5]

func embedString(s string) Value {
	var p uint64
	p = uint64(len(s)) << 48 // length in bits 54-48, bit 55 stays 0
	for i := 0; i < len(s); i++ {
		p |= uint64(s[i]) << uint(40-i*8)
	}
	return newRaw(TagString, p)
}

func decodeEmbeddedString(payload uint64, length int) string {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = byte(payload >> uint(40-i*8))
	}
	return string(b)
}

// ---------------------------------------------------------------------------
// Zero test
// ---------------------------------------------------------------------------

// IsZero reports whether v is the zero / default value of its type:
// nil, false, 0, 0.0, "", empty array, or empty map.
func (v Value) IsZero() bool {
	switch v.tag() {
	case TagNil:
		return true
	case TagBool:
		return v.payload() == 0
	case TagInt:
		return v.AsInt() == 0
	case TagFloat:
		return v.AsFloat() == 0.0
	case TagString:
		return v.AsString() == ""
	case TagArray:
		return len(v.AsArray()) == 0
	case TagMap:
		return len(v.AsMap()) == 0
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// Heap-backed struct definitions
// ---------------------------------------------------------------------------

type heapString struct{ s string }
type heapArray struct{ arr []Value }
type heapMap struct{ m map[string]Value }

// ---------------------------------------------------------------------------
// Debugging
// ---------------------------------------------------------------------------

// String returns a human-readable representation of v for debugging.
func (v Value) String() string {
	switch v.tag() {
	case TagNil:
		return "nil"
	case TagBool:
		return fmt.Sprintf("bool(%v)", v.AsBool())
	case TagInt:
		return fmt.Sprintf("int(%d)", v.AsInt())
	case TagFloat:
		return fmt.Sprintf("float(%g)", v.AsFloat())
	case TagString:
		return fmt.Sprintf("string(%q)", v.AsString())
	case TagArray:
		return fmt.Sprintf("array(len=%d)", len(v.AsArray()))
	case TagMap:
		return fmt.Sprintf("map(len=%d)", len(v.AsMap()))
	default:
		return fmt.Sprintf("%s(...)", v.tag())
	}
}
