package build

import (
	"fmt"
	"golang.org/x/tools/go/packages"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuilderConfig struct {
	Compiler    string
	Source      string
	Output      string
	Tags        string
	Ldflags     string
	IncludeTest bool
	GCFlags     string
}

type Builder struct {
	config *BuilderConfig
}

func NewBuilder(cfg *BuilderConfig) *Builder {
	if cfg == nil {
		panic("missing source")
	}
	if cfg.Compiler == "" {
		cfg.Compiler = "go"
	}
	if cfg.Output == "" {
		cfg.Output = filepath.Join(cfg.Source, "wasm-build")
	}

	return &Builder{config: cfg}
}

func (b *Builder) Build() error {
	if err := os.MkdirAll(b.config.Output, 0700); err != nil {
		return err
	}

	cmd := exec.Command(
		b.config.Compiler,
		"build",
		"-ldflags="+b.config.Ldflags,
		"-tags="+b.config.Tags,
		"-gcflags="+b.config.GCFlags,
		"-o="+filepath.Join(b.config.Output, "main.wasm"),
		b.config.Source,
	)

	cmd.Env = append(
		os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)

	r, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running compiler: %s", string(r))
	}

	if len(r) > 1 {
		fmt.Println(string(r))
	}

	return b.BuildFiles()
}

func (b *Builder) BuildFiles() error {
	if _, err := os.Stat(filepath.Join(b.config.Output, "index.html")); err != nil {
		if err := ioutil.WriteFile(filepath.Join(b.config.Output, "index.html"), []byte(jsIndex), 0600); err != nil {
			return err
		}
	}

	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return err
	}
	wasmJS := filepath.Join(strings.TrimSpace(string(goroot)), "misc", "wasm", "wasm_exec.js")
	if _, err := os.Stat(wasmJS); err != nil {
		return fmt.Errorf("failed to find $GOROOT/misc/wasm/wasm_exec.js driver: %v", err)
	}
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
		Env:   append(os.Environ(), "GOOS=js", "GOARCH=wasm"),
		Tests: b.config.IncludeTest,
	}, b.config.Source)
	if err != nil {
		return err
	}
	if len(pkgs) > 1 && len(pkgs[0].GoFiles) == 0 {
		pkgs[0] = pkgs[1]
	}
	extraJS, err := b.findPackagesJS(pkgs[0], make(map[string]bool))
	if err != nil {
		return err
	}

	return mergeJSFiles(filepath.Join(b.config.Output, "wasm.js"), append([]string{wasmJS}, extraJS...)...)
}

func (b *Builder) findPackagesJS(p *packages.Package, visited map[string]bool) (extraJS []string, err error) {
	if len(p.GoFiles) == 0 {
		return nil, nil
	}

	files, err := filepath.Glob(filepath.Join(filepath.Dir(p.GoFiles[0]), "*_js.js"))
	if err != nil {
		return nil, err
	}
	if b.config.IncludeTest && len(visited) == 0 {
		filesTests, err := filepath.Glob(filepath.Join(filepath.Dir(p.GoFiles[0]), "*_js_test.js"))
		if err != nil {
			return nil, err
		}
		files = append(files, filesTests...)
	}
	if err != nil {
		return nil, err
	}

	extraJS = append(extraJS, files...)

	for _, imp := range p.Imports {
		if !visited[imp.ID] {
			extra, err := b.findPackagesJS(imp, visited)
			if err != nil {
				return nil, err
			}
			extraJS = append(extraJS, extra...)
			visited[imp.ID] = true
		}
	}
	return extraJS, nil
}

// mergeJSFiles will merge all files into a single `wasm.js`. It will prepend the jsSetGo
// and append the jsStartGo.
func mergeJSFiles(dst string, files ...string) (err error) {
	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(w, strings.NewReader(jsSetGo)); err != nil {
		return err
	}
	for i := range files {
		r, err := os.Open(files[i])
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		r.Close()
		if err != nil {
			return err
		}
	}
	if _, err = io.Copy(w, strings.NewReader(jsStartGo)); err != nil {
		return err
	}
	return nil
}

const (
	jsIndex = `<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, user-scalable=no">
		<meta name="mobile-web-app-capable" content="yes">
		<script defer src="wasm.js"></script>
		<style>html,body{padding:0;margin:0}</style>
	</head>
	<body></body>
</html>`

	// jsSetGo sets the `window.go` variable.
	jsSetGo = `(() => {
"use strict";

globalThis._exit_code = null;
let go = undefined;

(() => {
    go = {argv: [], env: {}, importObject: {gojs: {}}};
	const argv = new URLSearchParams(location.search).get("argv");
	if (argv) {
		go["argv"] = argv.split(" ");
	}
})();`

	// jsStartGo initializes the main.wasm.
	jsStartGo = `(() => {
	let defaultGo = new Go();
	Object.assign(defaultGo["argv"], defaultGo["argv"].concat(go["argv"]));
	Object.assign(defaultGo["env"], go["env"]);
	for (let key in go["importObject"]) {
		if (typeof defaultGo["importObject"][key] === "undefined") {
			defaultGo["importObject"][key] = {};
		}
		Object.assign(defaultGo["importObject"][key], go["importObject"][key]);
	}
	defaultGo.exit = function(code) {
		if (code !== 0) {
			console.warn("exit code:", code);
		}
		globalThis._exit_code = code + 1;
	};
	go = defaultGo;
    if (!WebAssembly.instantiateStreaming) { // polyfill
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
        go.run(result.instance);
    });
})();
})();`
)
