package inkwasm

import (
	"syscall/js"
	"unsafe"
)

// NewObjectFromSyscall creates a new Object using the given
// js.Value. It will copy the information and calls Javascript
// functions.
//
// You must avoid calling that function. You also must release
// the resulting Object using Free, when no longer in use.
//
// It's useful for js.FuncOf, since Golang doesn't have option
// to export function directly.
func NewObjectFromSyscall(o js.Value) Object {
	return newObjectFromSyscall(*(*uint32)(unsafe.Pointer(&o)))
}

//inkwasm:get go._values
func newObjectFromSyscall(i uint32) Object
