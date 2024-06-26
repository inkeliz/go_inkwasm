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
	defer obj.Free()

	objInvoked, err := obj.Invoke(args...)
	if err != nil {
		t.Fatal(err)
	}
	defer objInvoked.Free()

	r, err := objInvoked.Bool()
	if err != nil {
		t.Fatal(err)
	}
	if r != true {
		t.Fatal("not true")
	}
}

//inkwasm:export
type TestExportStruct struct {
	_  uint64
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

//inkwasm:export
type TestExportStruct2 struct {
	_         uint64
	SomeField uint8            `js:"someField"`
	Inside    TestExportStruct `js:"nested"`
}

//inkwasm:func globalThis.TestExportedNested
func gen_TestExportedNested(TestExportStruct2, int32) int32

func TestExportNested(t *testing.T) {
	x := TestExportStruct2{Inside: TestExportStruct{ID: rand.Int31()}}
	if xx := gen_TestExportedNested(x, int32(x.Inside.ID)); xx != x.Inside.ID {
		t.Errorf("exported struct fail, expect %v receives %v", x.Inside.ID, xx)
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

//inkwasm:func nonExistentFunction
func gen_nonExistentFunction(s string) (Object, bool)

func TestNonExistentFunction(t *testing.T) {
	_, ok := gen_nonExistentFunction("Hello, 世界")
	if ok {
		t.Error("non existent function should return false")
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

//inkwasm:func globalThis.TestFloat_Echo
func gen_TestFloat_Echo(o float64) float64

//inkwasm:get globalThis.TestFloat_Echo
func runtime_TestFloat_Echo() Object

func TestFloat_Echo(t *testing.T) {
	if gen_TestFloat_Echo(52.0) != 52.0 {
		t.Error("float error, generator")
	}
	obj, err := runtime_TestFloat_Echo().Invoke(float64(52))
	if err != nil {
		t.Error(err)
	}
	defer obj.Free()

	v, err := obj.Float()
	if err != nil {
		t.Error(err)
	}
	if v != 52 {
		t.Error("uint error, runtime")
	}
}

//inkwasm:func globalThis.TestUint_Echo
func gen_TestUint_Echo(o uint64) uint64

//inkwasm:get globalThis.TestUint_Echo
func runtime_TestUint_Echo() Object

func TestUint_Echo(t *testing.T) {
	if gen_TestUint_Echo(18446744073709551615) != 18446744073709551615 {
		t.Error("uint error, generator")
	}
	obj, err := runtime_TestUint_Echo().Invoke(int(52))
	if err != nil {
		t.Error(err)
	}
	defer obj.Free()

	v, err := obj.Int()
	if err != nil {
		t.Error(err)
	}
	if v != 52 {
		t.Error("uint error, runtime")
	}
}

//inkwasm:func globalThis.TestUint64_Static
func gen_TestUint64_Static() uint64

func TestUint64_Static(t *testing.T) {
	if gen_TestUint64_Static() != 18446744073709551615 {
		t.Error("uint error, generator")
	}
}

//inkwasm:func globalThis.TestInt64_Static
func gen_TestInt64_Static() uint64

func TestInt64_Static(t *testing.T) {
	if gen_TestInt64_Static() != 9223372036854775807 {
		t.Error("uint error, generator")
	}
}

//inkwasm:func globalThis.TestSumFromArray
func gen_TestSumFromArray([]uint32) uint32

func TestArgsReuse(t *testing.T) {
	args := []interface{}{
		[]uint32{1, 2, 3, 4, 5},
	}

	if gen_TestSumFromArray(args[0].([]uint32)) != 15 {
		t.Error("args reuse error")
	}

	obj := Global().GetProperty("TestSumFromArray")
	res, err := obj.Invoke(args...)
	if err != nil {
		t.Error(err)
	}

	v, err := res.Int()
	if err != nil {
		t.Error(err)
	}

	if v != 15 {
		t.Error("args reuse error")
	}

	resReused, err := obj.Invoke(args...)
	if err != nil {
		t.Error(err)
	}

	v, err = resReused.Int()
	if err != nil {
		t.Error(err)
	}

	if v != 15 {
		t.Error("args reuse error")
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
func gen_GetLocationHostname() Object

func Benchmark_GetLocationHostname_INKWASM(b *testing.B) {
	b.ReportAllocs()

	fn := gen_GetLocationHostname()
	for i := 0; i < b.N; i++ {
		if r, err := fn.String(); r == "" || err != nil {
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

//inkwasm:func .appendChild
func gen_appendChild(o Object, v Object)

//inkwasm:get globalThis.document.body
func gen_getBody() Object

func Benchmark_DOM_INKWASM(b *testing.B) {
	b.ReportAllocs()

	body := gen_getBody()

	for i := 0; i < b.N; i++ {
		div := gen_createElement("div")
		gen_setInnerHTML(div, "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		gen_appendChild(body, div)
		div.Free()

		if i%100 == 0 {
			gen_setInnerHTML(body, "")
		}
	}
}

func Benchmark_InsertAdjacentHTML_INKWASM(b *testing.B) {
	hijackInsertAdjacentHTML()

	b.ReportAllocs()
	b.ResetTimer()

	pattern, _ := Global().Get("String").New("beforebegin")
	defer pattern.Free()

	for i := 0; i < b.N; i++ {
		gen_InsertAdjacentHTML(pattern, i)
	}
}

//inkwasm:func globalThis.document.body.insertAdjacentHTML
func gen_InsertAdjacentHTML(Object, int)

func hijackInsertAdjacentHTML() {
	fn := js.Global().Get("Function").New()
	js.Global().Get("HTMLElement").Get("prototype").Set("insertAdjacentHTML", fn)
}

func Benchmark_InsertAdjacentHTML_INKWASM_RUNTIME(b *testing.B) {
	hijackInsertAdjacentHTML()

	b.ReportAllocs()
	b.ResetTimer()

	pattern, _ := Global().Get("String").New("beforebegin")
	defer pattern.Free()

	body := Global().Get("document").Get("body")
	defer body.Free()

	fn, _ := Global().Get("HTMLElement").Get("prototype").Get("insertAdjacentHTML").Call("bind", body)
	defer fn.Free()

	for i := 0; i < b.N; i++ {
		fn.Invoke(pattern, i)
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

		if i%100 == 0 {
			body.SetProperty("innerHTML", "")
		}
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

		if i%100 == 0 {
			body.Set("innerHTML", "")
		}
	}
}

func Benchmark_DOM_BAD_JS_SYSCALL(b *testing.B) {
	b.ReportAllocs()

	doc := js.Global().Get("document")

	for i := 0; i < b.N; i++ {
		div := doc.Call("createElement", "div")
		div.Set("innerHTML", "foo <strong>bar</strong> baz "+strconv.Itoa(i))
		doc.Get("body").Call("appendChild", div)

		if i%100 == 0 {
			doc.Get("body").Set("innerHTML", "")
		}
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

		if i%100 == 0 {
			doc.GetProperty("body").SetProperty("innerHTML", "")
		}
	}
}
