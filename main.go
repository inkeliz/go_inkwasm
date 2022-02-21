// Package main
//

package main

import (
	"flag"
	"fmt"
	"github.com/inkeliz/go_inkwasm/build"
	"github.com/inkeliz/go_inkwasm/parser"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	buildConfig = &build.BuilderConfig{}
	testConfig  = &build.TesterConfig{}
	release     bool
)

func main() {
	flag.Parse()

	buildSet := flag.NewFlagSet("build", flag.ExitOnError)
	buildSet.StringVar(&buildConfig.Tags, "tags", "", "Sets -Tags for 'build'")
	buildSet.StringVar(&buildConfig.Output, "o", "", "Sets the output folder for 'build'")
	buildSet.StringVar(&buildConfig.Compiler, "compiler", "go", "Sets the compiler (default: go)")
	buildSet.StringVar(&buildConfig.GCFlags, "gcflags", "", "Set the compiler gcflags for 'build'")
	buildSet.BoolVar(&release, "release", false, "Compile as release-build")

	testSet := flag.NewFlagSet("test", flag.ExitOnError)
	testSet.StringVar(&testConfig.Port, "port", "", "Sets http port")
	testSet.StringVar(&buildConfig.Tags, "tags", "", "Sets -Tags")
	testSet.StringVar(&buildConfig.Output, "o", "", "Sets the output folder")
	testSet.StringVar(&testConfig.Count, "count", "1", "Run tests n times (default 1)")

	benchSet := flag.NewFlagSet("bench", flag.ExitOnError)
	benchSet.StringVar(&testConfig.Port, "port", "", "Sets http port")
	benchSet.StringVar(&buildConfig.Tags, "tags", "", "Sets -Tags for 'build'")
	benchSet.StringVar(&buildConfig.Output, "o", "", "Sets the output folder for 'build'")
	benchSet.StringVar(&testConfig.BenchRun, "run", "", "Run only benchmarks matching regexp")
	benchSet.StringVar(&testConfig.Count, "count", "1", "Run benchmarks n times (default 1)")
	benchSet.StringVar(&testConfig.Time, "time", "", "Run each benchmark for duration d (default 5s)")
	benchSet.StringVar(&testConfig.Shuffle, "shuffle", "off", "Run each benchmark at random order (default off)")

	fn := flag.Arg(0)
	pkg := flag.Arg(len(flag.Args()) - 1)
	buildConfig.Source = pkg

	if fn != "build" && fn != "generate" && fn != "test" && fn != "bench" {
		fmt.Println("invalid command, should be 'generate' or 'build' or 'test' or 'bench'")
		return
	}
	if pkg == "" {
		fmt.Println("specify a package")
		return
	}

	if release {
		buildConfig.Ldflags = "-w -s"
	}

	switch fn {
	case "generate":
		generate(pkg)
	case "build":
		buildSet.Parse(flag.Args()[1:])
		generate(pkg)
		create()
	case "test":
		testSet.Parse(flag.Args()[1:])
		generate(pkg)
		test()
	case "bench":
		testConfig.BenchRun = ".*"
		testConfig.Time = "2s"
		benchSet.Parse(flag.Args()[1:])
		generate(pkg)
		test()
	default:
		// impossible to hit
		return
	}

}

func create() {
	builder := build.NewBuilder(buildConfig)
	if err := builder.Build(); err != nil {
		fmt.Println(err)
	}
}

func test() {
	if buildConfig.Output == "" {
		out, err := os.MkdirTemp("", "*")
		if err != nil {
			fmt.Println(err)
			return
		}
		buildConfig.Output = out
		defer func() {
			os.Remove(out)
		}()
	}

	testConfig.BuilderConfig = buildConfig
	builder := build.NewTester(testConfig)
	if err := builder.Build(); err != nil {
		fmt.Println(err)
		return
	}

	if err := builder.Run(); err != nil {
		fmt.Println(err)
		return
	}
}

func generate(pkg string) {
	m, err := parser.NewParser().ParsePackages(pkg)
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg errgroup.Group

	for pkg, b := range m {
		pkg, b := pkg, b
		wg.Go(func() error {
			binderRelease := parser.NewBinder(parser.Release)
			if err := binderRelease.Create(pkg, b); err != nil {
				return err
			}

			binderTests := parser.NewBinder(parser.Test)
			if err := binderTests.Create(pkg, b); err != nil {
				return err
			}

			for _, v := range []struct {
				Source    func() io.Reader
				File, Dir string
			}{
				{Source: binderRelease.JS, Dir: pkg.Dir, File: "inkwasm_js.js"},
				{Source: binderRelease.ASM, Dir: pkg.Dir, File: "inkwasm_js.s"},
				{Source: binderRelease.GO, Dir: pkg.Dir, File: "inkwasm_js.go"},
				{Source: binderTests.JS, Dir: pkg.Dir, File: "inkwasm_js_test.js"},
				{Source: binderTests.ASM, Dir: pkg.Dir, File: "inkwasm_js_test.s"},
				{Source: binderTests.GO, Dir: pkg.Dir, File: "inkwasm_js_test.go"},
			} {
				f, err := os.Create(filepath.Join(v.Dir, v.File))
				if err != nil {
					return err
				}

				n, err := io.Copy(f, v.Source())
				if err != nil {
					return err
				}

				if err := f.Close(); err != nil {
					return err
				}

				if n == 0 {
					if err := os.Remove(filepath.Join(v.Dir, v.File)); err != nil {
						return err
					}
				} else {
					if strings.Contains(v.File, "go") {
						exec.Command("goimports", "-w", filepath.Join(v.Dir, v.File)).Run()
					}
				}
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		fmt.Println(err)
	}
}
