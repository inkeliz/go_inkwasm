package build

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/inspector"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

/*
That file is based on https://github.com/agnivade/wasmbrowsertest
*/

type TesterConfig struct {
	*BuilderConfig
	Port     string
	BenchRun string
	Count    string
	Time     string
	Shuffle  string
}

type Tester struct {
	builder *Builder
	config  *TesterConfig
}

func NewTester(cfg *TesterConfig) *Tester {
	if cfg == nil {
		panic("missing source")
	}
	cfg.IncludeTest = true
	return &Tester{builder: NewBuilder(cfg.BuilderConfig), config: cfg}
}

func (t *Tester) Build() error {
	cmd := exec.Command(
		t.builder.config.Compiler,
		"test",
		"-tags="+t.builder.config.Tags,
		"-c",
		"-o="+filepath.Join(t.builder.config.Output, "main.wasm"),
		t.builder.config.Source,
	)

	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	if r, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error running compiler: %s", string(r))
	}

	return t.builder.BuildFiles()
}

func (t *Tester) Run() error {
	logger := log.New(os.Stderr, "[inkwasm]: ", log.LstdFlags|log.Lshortfile)

	// NOTE: Since `os.Exit` will cause the process to exit, this defer
	// must be at the bottom of the defer stack to allow all other defer calls to
	// be called first.
	exitCode := 0
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	l, err := net.Listen("tcp", "localhost:"+t.config.Port)
	if err != nil {
		logger.Fatal(err)
	}

	// Setup web server.
	httpServer := &http.Server{
		Handler: http.FileServer(http.Dir(t.builder.config.Output)),
	}

	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts, chromedp.Flag("headless", false))

	// create chrome instance
	allocCtx, cancelAllocCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAllocCtx()
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		handleEvent(ctx, ev, logger)
	})

	done := make(chan struct{})
	go func() {
		if err = httpServer.Serve(l); err != http.ErrServerClosed {
			logger.Println(err)
		}
		done <- struct{}{}
	}()

	testOptions := ""
	if t.config.BenchRun != "" {
		testOptions += "-test.bench=" + t.config.BenchRun + " "
	}
	if t.config.Time != "" {
		testOptions += "-test.benchtime=" + t.config.Time + " "
	}
	if t.config.Shuffle != "" {
		testOptions += "-test.shuffle=" + t.config.Shuffle + " "
	}
	if t.config.Count != "" {
		testOptions += "-test.count=" + t.config.Count + " "
	}

	tasks := []chromedp.Action{
		chromedp.Navigate(`http://` + l.Addr().String() + "?argv=" + testOptions),
		chromedp.Poll("_exit_code", &exitCode, chromedp.WithPollingInterval(time.Second)),
	}

	err = chromedp.Run(ctx, tasks...)
	if err != nil {
		logger.Println(err)
	}

	exitCode = exitCode - 1
	if exitCode != 0 {
		exitCode = 1
	}

	// create a timeout
	ctx, cancelHTTPCtx := context.WithTimeout(ctx, 10*time.Second)
	defer cancelHTTPCtx()
	// Close shop
	err = httpServer.Shutdown(ctx)
	if err != nil {
		logger.Println(err)
	}

	return nil
}

func handleEvent(ctx context.Context, ev interface{}, logger *log.Logger) {
	switch ev := ev.(type) {
	case *cdpruntime.EventConsoleAPICalled:
		for _, arg := range ev.Args {
			line := string(arg.Value)
			if line == "" { // If Value is not found, look for Description.
				line = arg.Description
			}
			// Any string content is quoted with double-quotes.
			// So need to treat it specially.
			s, err := strconv.Unquote(line)
			if err != nil {
				// Probably some numeric content, print it as is.
				fmt.Printf("%s\n", line)
				continue
			}
			fmt.Printf("%s\n", s)
		}
	case *cdpruntime.EventExceptionThrown:
		if ev.ExceptionDetails != nil {
			details := ev.ExceptionDetails
			fmt.Printf("%s:%d:%d %s\n", details.URL, details.LineNumber, details.ColumnNumber, details.Text)
			if details.Exception != nil {
				fmt.Printf("%s\n", details.Exception.Description)
			}
			err := chromedp.Cancel(ctx)
			if err != nil {
				logger.Printf("error in cancelling context: %v\n", err)
			}
		}
	case *target.EventTargetCrashed:
		fmt.Printf("target crashed: status: %s, error code:%d\n", ev.Status, ev.ErrorCode)
		err := chromedp.Cancel(ctx)
		if err != nil {
			logger.Printf("error in cancelling context: %v\n", err)
		}
	case *inspector.EventDetached:
		fmt.Println("inspector detached: ", ev.Reason)
		err := chromedp.Cancel(ctx)
		if err != nil {
			logger.Printf("error in cancelling context: %v\n", err)
		}
	}
}
