package inkwasm

import (
	"fmt"
	"math/big"
	"unsafe"
)

var (
	decodeObject        = newBasicDecoder("InkwasmObject")
	decodeString        = newBasicDecoder("String")
	decodeFloat32       = newBasicDecoder("Float32")
	decodeFloat64       = newBasicDecoder("Float64")
	decodeBool          = newBasicDecoder("Bool")
	decodeUintptr       = newBasicDecoder("UintPtr")
	decodeInt           = newBasicDecoder("Int")
	decodeUint          = newBasicDecoder("Uint")
	decodeUint8         = newBasicDecoder("Uint8")
	decodeInt8          = newBasicDecoder("Int8")
	decodeUint16        = newBasicDecoder("Uint16")
	decodeInt16         = newBasicDecoder("Int16")
	decodeUint32        = newBasicDecoder("Uint32")
	decodeInt32         = newBasicDecoder("Int32")
	decodeUint64        = newBasicDecoder("Uint64")
	decodeInt64         = newBasicDecoder("Int64")
	decodeBigInt        = newBasicDecoder("BigInt")
	decodeUnsafePointer = newBasicDecoder("UnsafePointer")
	decodeSliceUint8    = newSliceDecoder("ArrayUint8")
	decodeSliceUint16   = newSliceDecoder("ArrayUint16")
	decodeSliceUint32   = newSliceDecoder("ArrayUint32")
	decodeSliceUint64   = newSliceDecoder("ArrayUint64")
	decodeSliceInt8     = newSliceDecoder("ArrayInt8")
	decodeSliceInt16    = newSliceDecoder("ArrayInt16")
	decodeSliceInt32    = newSliceDecoder("ArrayInt32")
	decodeSliceInt64    = newSliceDecoder("ArrayInt64")
)

func newSliceDecoder(s string) uint64 {
	v := getSliceDecoder(getBasicDecoder(s))
	return *(*uint64)(unsafe.Pointer(&v))
}

func newBasicDecoder(s string) uint64 {
	v := getBasicDecoder(s).value
	return *(*uint64)(unsafe.Pointer(&v))
}

//inkwasm:get globalThis.inkwasm.Load
func getBasicDecoder(s string) Object

//inkwasm:func globalThis.inkwasm.Load.SliceOf
func getSliceDecoder(f Object) Object

// createArgs changes the given interface{}. It is two-word size. To
// avoid allocs, we replace the "type-pointer" to decodeObject.
func createArgs(args []interface{}) ([]int32, error) {
	for i, a := range args {
		var argDecoder uint64
		switch a.(type) {
		case Object:
			argDecoder = decodeObject
		case string:
			argDecoder = decodeString
		case float32:
			argDecoder = decodeFloat32
		case float64:
			argDecoder = decodeFloat64
		case uintptr:
			argDecoder = decodeUintptr
		case bool:
			argDecoder = decodeBool
		case int:
			argDecoder = decodeInt
		case uint:
			argDecoder = decodeUint
		case uint8:
			argDecoder = decodeUint8
		case int8:
			argDecoder = decodeInt8
		case uint16:
			argDecoder = decodeUint16
		case int16:
			argDecoder = decodeInt16
		case uint32:
			argDecoder = decodeUint32
		case int32:
			argDecoder = decodeInt32
		case uint64:
			argDecoder = decodeUint64
		case int64:
			argDecoder = decodeInt64
		case big.Int:
			argDecoder = decodeBigInt
		case unsafe.Pointer:
			argDecoder = decodeUnsafePointer
		case []uint8:
			argDecoder = decodeSliceUint8
		case []int8:
			argDecoder = decodeSliceInt8
		case []uint16:
			argDecoder = decodeSliceUint16
		case []int16:
			argDecoder = decodeSliceInt16
		case []uint32:
			argDecoder = decodeSliceUint32
		case []int32:
			argDecoder = decodeSliceInt32
		case []uint64:
			argDecoder = decodeSliceUint64
		case []int64:
			argDecoder = decodeSliceInt64
		default:
			return nil, fmt.Errorf("unsupported argument type of %T", a)
		}

		// Replace the "type information" from interface{}
		// to the Object which decodes current argument.
		*(*uint64)(unsafe.Pointer(&args[i])) = argDecoder
	}

	v := *(*[]int32)(unsafe.Pointer(&args))
	(*[3]uint64)(unsafe.Pointer(&v))[1] *= 4
	return v, nil
}
