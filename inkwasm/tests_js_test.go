package inkwasm

import (
	"bytes"
	"math"
	"math/rand"
	"strconv"
	"syscall/js"
	"testing"
	_ "unsafe"
)

func testInvoke(t *testing.T, obj Object, args ...interface{}) {
	t.Helper()
	defer obj.Free()

	objInvoked, err := obj.Invoke(args...)
	if err != nil {
		t.Error(err)
	}
	defer objInvoked.Free()

	r, err := objInvoked.Bool()
	if err != nil {
		t.Error(err)
	}
	if r != true {
		t.Fail()
	}
}

//inkwasm:export
type TestExportStruct struct {
	ID int32 `js:"id"`
	X  int8  `js:"x"`
}

//inkwasm:func globalThis.TestExported
func gen_TestExported(TestExportStruct, int32) int32

func TestExport(t *testing.T) {
	x := TestExportStruct{ID: rand.Int31()}
	if xx := gen_TestExported(x, int32(x.ID)); xx != x.ID {
		t.Errorf("exported struct fail, expect %v receives %v", x.ID, xx)
	}
}

//inkwasm:func globalThis.TestAlignment
func gen_TestAlignment(bool, int16) int16

func TestAlignment(t *testing.T) {
	if gen_TestAlignment(true, math.MaxInt16) != math.MaxInt16 {
		t.Error("alignment wrong")
	}
	if gen_TestAlignment(false, math.MaxInt16) != 0 {
		t.Error("alignment wrong")
	}
}

//inkwasm:func globalThis.TestAlignment2
func gen_TestAlignment2([3]float64) int16

func TestAlignment2(t *testing.T) {
	x := int16(rand.Int31())
	xf := float64(x)
	if gen_TestAlignment2([3]float64{xf, xf, xf}) != x {
		t.Error("alignment wrong")
	}
	if gen_TestAlignment2([3]float64{xf, xf, xf}) != x {
		t.Error("alignment wrong")
	}
}

func TestObject_Bool(t *testing.T) {
	if gen_TestObjectType_Bool(true) != true {
		t.Error("bool error, generator")
	}
	testInvoke(t, runtime_TestObjectType_Bool(), true)
}

//inkwasm:func globalThis.TestObjectType_Bool
func gen_TestObjectType_Bool(b bool) bool

//inkwasm:get globalThis.TestObjectType_Bool
func runtime_TestObjectType_Bool() Object

func TestObjectType_String(t *testing.T) {
	if gen_TestObjectType_String("Hello, 世界") != true {
		t.Error("string error, generator")
	}
	testInvoke(t, runtime_TestObjectType_String(), "Hello, 世界")

	txt := `Lorem ipsum dolor sit amet, consectetur adipiscing elit.`
	b := gen_TestEcho(txt)
	if b != txt {
		t.Errorf("string echo error, generator, receving: %d expecting %d", len(b), len(txt))
	}

	obj := runtime_TestEcho()
	defer obj.Free()

	objInvoked, err := obj.Invoke(txt)
	if err != nil {
		t.Error(err)
	}

	b, err = objInvoked.String()
	if err != nil {
		t.Error(err)
	}
	if b != txt {
		t.Error("string echo error, runtime", b, txt)
	}
}

//inkwasm:func globalThis.TestObjectType_String
func gen_TestObjectType_String(s string) bool

//inkwasm:get globalThis.TestObjectType_String
func runtime_TestObjectType_String() Object

//inkwasm:func globalThis.TestEcho
func gen_TestEcho(s string) string

//inkwasm:get globalThis.TestEcho
func runtime_TestEcho() Object

func TestObjectType_Object(t *testing.T) {
	if gen_TestObjectType_Object(runtime_TestObjectType_String()) != true {
		t.Fail()
	}
	testInvoke(t, runtime_TestObjectType_Object(), runtime_TestObjectType_String())
}

//inkwasm:func globalThis.TestObjectType_Object
func gen_TestObjectType_Object(o Object) bool

//inkwasm:get globalThis.TestObjectType_Object
func runtime_TestObjectType_Object() Object

func TestObject_Bytes(t *testing.T) {
	g := gen_TestObject_Bytes()
	if len(g) != 4 || cap(g) != 4 {
		t.Error("invalid bytes len")
	}
	if !bytes.Equal(g, []byte{0x00, 0x01, 0x02, 0x03}) {
		t.Error("invalid bytes")
	}

	o, err := runtime_TestObject_Bytes().Invoke()
	if err != nil {
		t.Error(err)
	}
	r, err := o.Bytes(nil)
	if err != nil {
		t.Error(err)
	}
	if len(r) != 4 || cap(r) != 4 {
		t.Error("invalid bytes len")
	}
	if !bytes.Equal(r, []byte{0x00, 0x01, 0x02, 0x03}) {
		t.Error("invalid bytes")
	}

	inputs := make([][]byte, 0, 100)
	outputs := make([][]byte, 0, 100)
	for i := 0; i < 100; i++ {
		in := make([]byte, 1<<11)
		rand.Read(in)
		inputs = append(inputs, in)
		outputs = append(outputs, gen_TestEchoByte(in))
	}

	for i, expected := range inputs {
		if !bytes.Equal(outputs[i], expected) {
			t.Error("error bytes not equal")
		}
	}

}

//inkwasm:func globalThis.TestObject_Bytes
func gen_TestObject_Bytes() []byte

//inkwasm:get globalThis.TestObject_Bytes
func runtime_TestObject_Bytes() Object

//inkwasm:func globalThis.TestEcho
func gen_TestEchoByte(s []byte) []byte

//inkwasm:get globalThis.TestObject_GetRandom
func gen_GetRandom([]byte) Object

func Benchmark_BytesRandom_INKWASM(b *testing.B) {
	b.ReportAllocs()

	r := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		gen_GetRandom(r)
	}
}

func Benchmark_BytesRandom_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	fn := js.Global().Get("TestObject_GetRandom")
	array := js.Global().Get("Uint8Array").New(10)
	r := make([]byte, 10)
	for i := 0; i < b.N; i++ {
		fn.Invoke(array)
		js.CopyBytesToGo(r, array)
	}
}

//inkwasm:func globalThis.localStorage.setItem
func gen_SetStorageItem(k, v string)

//inkwasm:func globalThis.localStorage.getItem
func gen_GetStorageItem(k string) string

func Benchmark_SetLocalStorage_INKWASM(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		gen_SetStorageItem("test", "test")
	}
}

func Benchmark_SetLocalStorage_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	ls := js.Global().Get("localStorage")
	fn := ls.Get("setItem").Call("bind", ls)
	for i := 0; i < b.N; i++ {
		fn.Invoke("test", "test")
	}
}

func Benchmark_GetLocalStorage_INKWASM(b *testing.B) {
	b.ReportAllocs()

	gen_SetStorageItem("test", "test")
	for i := 0; i < b.N; i++ {
		if gen_GetStorageItem("test") != "test" {
			b.Fail()
		}
	}
}

func Benchmark_GetLocalStorage_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	js.Global().Get("localStorage").Call("setItem", "test", "test")
	ls := js.Global().Get("localStorage")
	fn := ls.Get("getItem").Call("bind", ls)
	for i := 0; i < b.N; i++ {
		if fn.Invoke("test").String() != "test" {
			b.Fail()
		}
	}
}

//inkwasm:get globalThis.location.hostname
func gen_GetLocationHostname() string

func Benchmark_GetLocationHostname_INKWASM(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if gen_GetLocationHostname() == "" {
			b.Fail()
		}
	}
}

func Benchmark_GetLocationHostname_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	fn := js.Global().Get("location").Get("hostname")
	for i := 0; i < b.N; i++ {
		if fn.String() == "" {
			b.Fail()
		}
	}
}

//inkwasm:func globalThis.document.createElement
func gen_createElement(v string) Object

//inkwasm:set .innerHTML
func gen_setInnerHTML(o Object, v string)

//inkwasm:func globalThis.document.body.appendChild
func gen_appendChild(c Object)

func Benchmark_DOM_INKWASM(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		div := gen_createElement("div")
		gen_setInnerHTML(div, "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		gen_appendChild(div)
	}
}

func Benchmark_DOM_RUNTIME_INKWASM(b *testing.B) {
	b.ReportAllocs()

	doc := Global().GetProperty("document")
	defer doc.Free()
	createElement, _ := doc.GetProperty("createElement").Call("bind", doc)
	defer createElement.Free()
	body := doc.GetProperty("body")
	defer body.Free()
	appendChild, _ := body.GetProperty("appendChild").Call("bind", body)
	defer appendChild.Free()
	for i := 0; i < b.N; i++ {
		div, _ := createElement.Invoke("div")
		div.SetProperty("innerHTML", "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		appendChild.InvokeVoid(div)
		div.Free()
	}
}

func Benchmark_DOM_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	doc := js.Global().Get("document")
	createElement := doc.Get("createElement").Call("bind", doc)
	body := doc.Get("body")
	appendChild := body.Get("appendChild").Call("bind", body)
	for i := 0; i < b.N; i++ {
		div := createElement.Invoke("div")
		div.Set("innerHTML", "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		appendChild.Invoke(div)
	}
}

func Benchmark_DOM_BAD_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	doc := js.Global().Get("document")
	for i := 0; i < b.N; i++ {
		div := doc.Call("createElement", "div")
		div.Set("innerHTML", "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		doc.Get("body").Call("appendChild", div)
	}
}

func Benchmark_DOM_BAD_RUNTIME_INKWASM(b *testing.B) {
	b.ReportAllocs()

	doc := Global().GetProperty("document")
	defer doc.Free()
	for i := 0; i < b.N; i++ {
		div, _ := doc.Call("createElement", "div")
		div.SetProperty("innerHTML", "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		doc.GetProperty("body").Call("appendChild", div)
		div.Free()
	}
}
