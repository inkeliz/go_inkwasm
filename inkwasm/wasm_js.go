package inkwasm

import (
	"errors"
	"syscall/js"
	"unsafe"
)

// Object represents one Javascript Object
type Object struct {
	_ [0]func() // not comparable
	// value holds the value of the js-object
	// if typ == Object it's the value of the index
	value     [8]byte
	typ       ObjectType // type of js object
	protected bool       // protected prevents been released
	_         [2]uint8   // padding, reserved
	len       uint32     // len for array/string
}

// NewObject creates a new Javascript Object.
func NewObject(args ...interface{}) Object {
	v, err := createArgs(args)
	if err != nil {
		return makeObj(nil)
	}
	return makeObj(v)
}

//inkwasm:func globalThis.inkwasm.Internal.Make
func makeObj([]int32) Object

// Free deletes the reference from the array on Javascript's side.
func (o Object) Free() {
	if o.protected {
		return
	}
	switch o.typ {
	case TypeFunction, TypeSymbol, TypeBigInt, TypeObject:
		free(*(*int)(unsafe.Pointer(&o.value)))
	}
}

//inkwasm:func globalThis.inkwasm.Internal.Free
func free(ref int)

// Call calls the method from the current Object, using args as
// arguments for the method.
//
// The resulting Object must be released using Free, when
// no longer in use.
func (o Object) Call(method string, args ...interface{}) (Object, error) {
	v, err := createArgs(args)
	if err != nil {
		return Undefined(), err
	}
	r, ok := call(o, method, v)
	if !ok {
		return Undefined(), ErrExecutionJS
	}
	return r, nil
}

//inkwasm:func globalThis.inkwasm.Internal.Call
func call(o Object, k string, args []int32) (Object, bool)

// CallVoid is similar to Call, but doesn't return the resulting Object.
// Look at Call function for more details.
func (o Object) CallVoid(method string, args ...interface{}) error {
	v, err := createArgs(args)
	if err != nil {
		return err
	}
	if _, ok := callVoid(o, method, v); !ok {
		return ErrExecutionJS
	}
	return nil
}

//inkwasm:func globalThis.inkwasm.Internal.Call
func callVoid(o Object, k string, args []int32) (_, ok bool)

// Invoke invokes the current Object, calling itself with the
// provided args as arguments of the function.
//
// The Object must be a Javascript-Function, or be callable.
// The resulting Object must be released using Free, when
// no longer in use.
func (o Object) Invoke(args ...interface{}) (Object, error) {
	v, err := createArgs(args)
	if err != nil {
		return Undefined(), err
	}
	r, ok := invoke(o, v)
	if !ok {
		return Undefined(), ErrExecutionJS
	}
	return r, nil
}

//inkwasm:func globalThis.inkwasm.Internal.Invoke
func invoke(o Object, args []int32) (Object, bool)

// InvokeVoid is similar to Invoke, but doesn't return the resulting Object.
// Look at Invoke function for more details.
func (o Object) InvokeVoid(args ...interface{}) error {
	v, err := createArgs(args)
	if err != nil {
		return err
	}
	if _, ok := invokeVoid(o, v); !ok {
		return ErrExecutionJS
	}
	return nil
}

//inkwasm:func globalThis.inkwasm.Internal.Invoke
func invokeVoid(o Object, args []int32) (_, ok bool)

// New uses the "new" operator from Javascript with the current object
// as the constructor and the given arg as arguments.
func (o Object) New(args ...interface{}) (Object, error) {
	v, err := createArgs(args)
	if err != nil {
		return Undefined(), err
	}
	r, ok := newObj(o, v)
	if !ok {
		return Undefined(), ErrExecutionJS
	}
	return r, nil
}

//inkwasm:new .
func newObj(o Object, args []int32) (Object, bool)

// GetIndex returns given index of the current Object.
func (o Object) GetIndex(index int) Object {
	return getIndex(o, index)
}

//inkwasm:get .
func getIndex(o Object, i int) Object

// GetProperty returns property of the current Object.
func (o Object) GetProperty(property string) Object {
	return getProp(o, property)
}

//inkwasm:get .
func getProp(o Object, k string) Object

// SetProperty defines the given property of the current Object with
// the given value.
func (o Object) SetProperty(property string, value string) {
	setProp(o, property, value)
}

//inkwasm:func Reflect.set
func setProp(o Object, k, v string)

var (
	ErrInvalidType = errors.New("invalid type")
	ErrExecutionJS = errors.New("error while executing/calling Javascript, see console log for details")
)

// Bool gets current Object value to bool.
// It will return error if the current Object isn't TypeBoolean.
func (o Object) Bool() (bool, error) {
	switch o.typ {
	case TypeBoolean:
		return *(*bool)(unsafe.Pointer(&o.value)), nil
	case TypeNumber:
		v, _ := o.Float()
		return v != 0, nil
	default:
		return false, ErrInvalidType
	}
}

// Float gets current Object value to float64.
// It will return error if the current Object isn't TypeNumber.
func (o Object) Float() (float64, error) {
	if o.typ != TypeNumber {
		return 0, ErrInvalidType
	}
	return *(*float64)(unsafe.Pointer(&o.value)), nil
}

// Int is a wrapper from Float
// It will return error if the current Object isn't TypeNumber
// or higher than int53.
func (o Object) Int() (int, error) {
	const MaxFloat64 = (2 << 52) - 1
	const MinFloat64 = -MaxFloat64
	i, err := o.Float()
	if err != nil {
		return 0, err
	}
	if !(i > MinFloat64 && i < MaxFloat64) {
		return 0, ErrInvalidType
	}
	return int(i), nil
}

// String return the value from the current Object as
// golang's string.
func (o Object) String() (string, error) {
	if o.typ != TypeString && o.typ != TypeObject {
		return "", ErrInvalidType
	}
	src := o
	if o.typ == TypeString {
		src = encodeString(o)
		defer src.Free()
	}
	buf := make([]byte, src.len, src.len)
	copyBytes(src, buf)
	return *(*string)(unsafe.Pointer(&buf)), nil
}

// MustString is a wrapper to String, but suppress errors.
func (o Object) MustString() string {
	r, _ := o.String()
	return r
}

//inkwasm:func globalThis.inkwasm.Internal.EncodeString
func encodeString(o Object) (_ Object)

// Bytes return the value from the current Object as
// byte-slice.
//
// If buf is nil, a new byte-slice will be created and
// used instead.
func (o Object) Bytes(buf []byte) ([]byte, error) {
	if o.typ != TypeObject && o.typ != TypeString {
		return nil, ErrInvalidType
	}
	if buf == nil {
		buf = make([]byte, o.len, o.len)
	}
	copyBytes(o, buf)
	return buf, nil
}

// MustBytes is a wrapper to Bytes, but suppress errors.
func (o Object) MustBytes(buf []byte) []byte {
	r, _ := o.Bytes(buf)
	return r
}

//inkwasm:func globalThis.inkwasm.Internal.Copy
func copyBytes(o Object, buf []byte)

// Length returns the length of the current object,
// when the object is string or array.
//
// It uses uint32, use GetProperty("length") for larger
// results.
func (o Object) Length() uint32 {
	return o.len
}

func (o Object) InstanceOf(v Object) bool {
	js.Value{}.Truthy()
	return instanceOf(o, v)
}

//inkwasm:func globalThis.inkwasm.Internal.InstanceOf
func instanceOf(o, v Object) bool

// Truthy returns the JavaScript "truthiness" of the value v. In JavaScript,
// false, 0, "", null, undefined, and NaN are "falsy", and everything else is
// "truthy".
//
// See https://developer.mozilla.org/en-US/docs/Glossary/Truthy.
func (o Object) Truthy() bool {
	switch o.typ {
	case TypeUndefined, TypeNull:
		return false
	case TypeBoolean:
		v, _ := o.Bool()
		return v
	case TypeNumber:
		v, _ := o.Float()
		return v != 0
	case TypeString:
		v, _ := o.String()
		return len(v) > 0
	default:
		return true
	}
}

// Equal returns if o Object is equal (==) to v.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Equality.
func (o Object) Equal(v Object) bool {
	if !v.Truthy() && !o.Truthy() {
		return true
	}
	if o.len != v.len {
		return false
	}
	return equal(o, v)
}

//inkwasm:func globalThis.inkwasm.Internal.Equal
func equal(o, v Object) bool

// StrictEqual returns if o Object is strict-equal (===) to v.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Strict_equality.
func (o Object) StrictEqual(v Object) bool {
	return strictEqual(o, v)
}

//inkwasm:func globalThis.inkwasm.Internal.StrictEqual
func strictEqual(o, v Object) bool
