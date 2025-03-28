package vigilant

import (
	"fmt"
	"strconv"
	"time"
)

// FieldType represents the type of a field
type FieldType int8

const (
	TypeString FieldType = iota
	TypeInt
	TypeBool
	TypeTime
	TypeFloat32
	TypeFloat64
	TypeComplex64
	TypeComplex128
	TypeByte
	TypeRune
	TypeUint
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeError
	TypeArray
	TypeSlice
	TypeMap
	TypeAny
)

// Field represents a field in an observability event
type Field struct {
	Type  FieldType `json:"type"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

// String returns the string representation of a Field
func String(key string, val string) Field {
	return Field{
		Type:  TypeString,
		Key:   key,
		Value: val,
	}
}

// Int returns the int representation of a Field
func Int(key string, val int) Field {
	return Field{
		Type:  TypeInt,
		Key:   key,
		Value: strconv.Itoa(val),
	}
}

// Bool returns the bool representation of a Field
func Bool(key string, val bool) Field {
	return Field{
		Type:  TypeBool,
		Key:   key,
		Value: strconv.FormatBool(val),
	}
}

// Time returns the time representation of a Field
func Time(key string, val time.Time) Field {
	return Field{
		Type:  TypeTime,
		Key:   key,
		Value: val.Format(time.RFC3339),
	}
}

// Float32 returns the float32 representation of a Field
func Float32(key string, val float32) Field {
	return Field{
		Type:  TypeFloat32,
		Key:   key,
		Value: strconv.FormatFloat(float64(val), 'f', -1, 32),
	}
}

// Float64 returns the float64 representation of a Field
func Float64(key string, val float64) Field {
	return Field{
		Type:  TypeFloat64,
		Key:   key,
		Value: strconv.FormatFloat(val, 'f', -1, 64),
	}
}

// Complex64 returns the complex64 representation of a Field
func Complex64(key string, val complex64) Field {
	return Field{
		Type:  TypeComplex64,
		Key:   key,
		Value: fmt.Sprintf("%g", val),
	}
}

// Complex128 returns the complex128 representation of a Field
func Complex128(key string, val complex128) Field {
	return Field{
		Type:  TypeComplex128,
		Key:   key,
		Value: fmt.Sprintf("%g", val),
	}
}

// Byte returns the byte representation of a Field
func Byte(key string, val byte) Field {
	return Field{
		Type:  TypeByte,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Rune returns the rune representation of a Field
func Rune(key string, val rune) Field {
	return Field{
		Type:  TypeRune,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Uint returns the uint representation of a Field
func Uint(key string, val uint) Field {
	return Field{
		Type:  TypeUint,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Uint8 returns the uint8 representation of a Field
func Uint8(key string, val uint8) Field {
	return Field{
		Type:  TypeUint8,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Uint16 returns the uint16 representation of a Field
func Uint16(key string, val uint16) Field {
	return Field{
		Type:  TypeUint16,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Uint32 returns the uint32 representation of a Field
func Uint32(key string, val uint32) Field {
	return Field{
		Type:  TypeUint32,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Uint64 returns the uint64 representation of a Field
func Uint64(key string, val uint64) Field {
	return Field{
		Type:  TypeUint64,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Int8 returns the int8 representation of a Field
func Int8(key string, val int8) Field {
	return Field{
		Type:  TypeInt8,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Int16 returns the int16 representation of a Field
func Int16(key string, val int16) Field {
	return Field{
		Type:  TypeInt16,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Int32 returns the int32 representation of a Field
func Int32(key string, val int32) Field {
	return Field{
		Type:  TypeInt32,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Int64 returns the int64 representation of a Field
func Int64(key string, val int64) Field {
	return Field{
		Type:  TypeInt64,
		Key:   key,
		Value: fmt.Sprintf("%d", val),
	}
}

// Error returns the error representation of a Field
func Error(key string, val error) Field {
	return Field{
		Type:  TypeError,
		Key:   key,
		Value: val.Error(),
	}
}

// Array returns the array representation of a Field
func Array(key string, val []any) Field {
	return Field{
		Type:  TypeArray,
		Key:   key,
		Value: fmt.Sprintf("%#v", val),
	}
}

// Slice returns the slice representation of a Field
func Slice(key string, val []any) Field {
	return Field{
		Type:  TypeSlice,
		Key:   key,
		Value: fmt.Sprintf("%#v", val),
	}
}

// Map returns the map representation of a Field
func Map(key string, val map[string]any) Field {
	return Field{
		Type:  TypeMap,
		Key:   key,
		Value: fmt.Sprintf("%#v", val),
	}
}

// Any returns the any representation of a Field
func Any(key string, val any) Field {
	return Field{
		Type:  TypeAny,
		Key:   key,
		Value: fmt.Sprintf("%#v", val),
	}
}
