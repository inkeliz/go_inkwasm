# Golang InkWasm

**InkWasm** is faster `syscall/js` replacement, it's a package and a generator. Our goal is to be as faster and avoid unnecessary allocations. **InkWasm** initially created for [Gio](https://gioui.org/), improving the performance for [WebGL API](https://developer.mozilla.org/pt-BR/docs/Web/API/WebGL_API), in some devices and tests **InkWasm** is 2x faster than `syscall/js`, in real application it's 1.6x faster.

> ⚠️ The generator changes the file of the current packages and imported packages. It is still experimental, keep a backup file and use some versioning tool (like Git). 


## Benchmark

The performance may vary based on how you use `syscall/js` and **InkWasm**. Look into `inkwasm/tests_js_test.go` for more details.

```
Benchmark_DOM_INKWASM                       50000        11462 ns/op         52 B/op         1 allocs/op
Benchmark_DOM_RUNTIME_INKWASM               50000        14934 ns/op        100 B/op         4 allocs/op
Benchmark_DOM_BAD_RUNTIME_INKWASM           50000        18302 ns/op        100 B/op         4 allocs/op
Benchmark_DOM_JS_SYSCALL                    50000        23550 ns/op        167 B/op        11 allocs/op
Benchmark_DOM_BAD_JS_SYSCALL                50000        30540 ns/op        175 B/op        12 allocs/op

Benchmark_BytesRandom_INKWASM              100000         1026 ns/op          0 B/op         0 allocs/op
Benchmark_BytesRandom_JS_SYSCALL           100000         1808 ns/op         40 B/op         3 allocs/op

Benchmark_SetLocalStorage_INKWASM          854655         6312 ns/op          0 B/op         0 allocs/op
Benchmark_SetLocalStorage_JS_SYSCALL       486326        11796 ns/op         64 B/op         4 allocs/op

Benchmark_GetLocalStorage_INKWASM         1830148         3300 ns/op          4 B/op         1 allocs/op
Benchmark_GetLocalStorage_JS_SYSCALL       826088         7356 ns/op         48 B/op         6 allocs/op

Benchmark_GetLocationHostname_JS_SYSCALL  3038510         1979 ns/op         32 B/op         2 allocs/op
Benchmark_GetLocationHostname_INKWASM     2499522         2442 ns/op         16 B/op         1 allocs/op
```

Currently, **InkWasm** is used in one fork of Gio, **you can test it online here**.

|                              | Chrome 98 @ Ryzen 3900X | Chrome 98 @ Xiaomi Note 9 | Safari 15 @ MacBook Air |
|-------------------------------------|------------------|-----------------|----------------|
| gioui.org@bed59024                  | ~4.65ms          | ~31.91ms         | ~5.58ms        |
| github.com/inkeliz/gio@main_inkjs   | ~3.02ms          | ~20.03ms         | ~4.84ms        |

The performance improvement varies from 13% up to 37%. [You can test it online, and compare the performance on your machine](https://gio-bench.pages.dev/). Want to know how it was implemented? [Check it here](https://github.com/Inkeliz/gio/commit/baf7d943e5d50debf354d4cd5f951d442d4d9b4e).

## Usage

In order to use the generator, you must create one `func` without body and describe what is the JS function (or attribute) that imports the function.

```
func main() {
    alert("Your String")
}

//inkwasm:func globalThis.alert
func alert(s string)
```

You should run: `go run github.com/inkeliz/go_inkwasm build .`. It will create a new `wasm-build` folder, you can run `npx serve ./wasm-build` and run it on browser.

The generator is faster, but you can also use the `inkwasm` on "runtime", similar to `syscall/js`:

```
func main() {
    alert := inkwasm.Global().Get("alert")
    defer alert.Free()

    alert.Invoke("Text")
}
```

Usually, it still faster than `syscall/js` and have less allocations.

-----------

#### Calling static functions:

You can invoke functions that are "hardcoded", for instance `alert`:
```
//inkwasm:func globalThis.alert
func alert(s string)
```
```
func main() {
    alert("Text")
}
```

It will call `globalThis.alert`.

### Calling dynamic functions:

You can also call functions that belongs to a specific Object, the name must start with `.`:

```
//inkwasm:func .bufferData
func glBufferData(o inkwasm.Object, target uint, data []byte, usage uint)
```
```
func main() {
    ///... 
    glBufferData(gl, len(data), STATIC_DRAW, data)
}
```

It will call `o.bufferData` (is expected that the given `o` (`inkwasm.Object`) is a WebGL Context).

### Get attribute:

In order to get a attribute use `inkwasm:get`:

```
//inkwasm:get globalThis.location.hostname
func getHostname() string
```

#### Set attribute:

In order to set attribute use `inkwasm:set`:

```
//inkwasm:set .innerHTML
func setInnerHTML(o inkwasm.Object, v string)
```

## Roadmap

Currently, **InkWasm** is very experimental and WebAssembly, in general, is also very experimental.

- [x] Calling JS functions (`inkwasm:func`)
- [x] Get JS property (`inkwasm:get`)
- [x] Set JS property (`inkwasm:set`)

- [x] Support integers input (`uint`, `int`, `int64`, ...)
- [x] Support integers output (`uint`, `int`, `int64`, ...)
- [x] Support floats input (`float64`, `float32`)
- [x] Support floats output (`float64`, `float32`)
- [x] Support slices/array input (`string`, `[]byte`, `[]float64`, `[10]byte`, ...)
- [x] Support slices/array output (`string`, `[]byte`, `[]float64`, `[10]byte`, ...)
- [x] Support big integers input (`big.Int`)
- [ ] Support big integers output (`big.Int`)
- [ ] Support channels input (`chan string`, ...)
- [ ] Support channels output (`chan string`, ...)
- [ ] Support complex input (`complex64`, `complex128`)
- [ ] Support complex output (`complex64`, `complex128`)
- [ ] Support functions input (`func(){}`)

- [ ] Support TinyGo

- [~] Export struct (`inkwasm:export`)
- [ ] Export functions (`inkwasm:export`)
- [ ] Export alias type (`inkwasm:export`)
- [ ] Exported functions without `js.FuncOf`

- [x] Import custom scripts (`*_js.js`) into `wasm.js`.
- [ ] Use `-overlay` on `cmd/go`.

- [ ] Improve tests

## Blockers

- `CallImport` will be removed:

  Currently, **InkWasm** heavily rely on `CallImport`, it will be removed in the future and will be exclusive to `syscall/js` and `runtime` [golang#38248](https://github.com/golang/go/issues/38248). If that happens, the only solution is to replace `syscall/js`, maybe using `-overlay` on `cmd/go`.

- `go:wasmexport` isn't (yet) supported:

  Currently, the only way to export function is using `syscall/js`. However, that is quite slow. We can hack into `runtime` and `syscall/js` to be able to replace `js.FuncOf`. However, I'm waiting to see the progress of the proposal ([golang#42372](https://github.com/golang/go/issues/42372)). 


## Compatibility

The current API is supposed to be backward compatible with any future update. However, the **InkWasm** may not work on older versions of Golang and the idea is to be compatible with the latest golang version.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)