// Command studio launches DSL Studio — a local-first, organic visual IDE and
// Language Engineering Lab for go-dsl with native Ollama & LM Studio AI
// integration (GBNF constrained decoding + deterministic closed-loop repair).
//
// Usage:
//
//	go run ./cmd/studio [-port 8420] [-dsl grammar.yaml] [-open]
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	var (
		host          string
		port          int
		dslFile       string
		openBrowser   bool
		allowRemoteAI bool
	)
	flag.StringVar(&host, "host", "127.0.0.1", "bind address for DSL Studio")
	flag.IntVar(&port, "port", 8420, "HTTP port for DSL Studio")
	flag.StringVar(&dslFile, "dsl", "", "optional YAML/JSON grammar file to open and edit in place")
	flag.BoolVar(&openBrowser, "open", false, "automatically open DSL Studio in the default browser")
	flag.BoolVar(&allowRemoteAI, "allow-remote-ai", false, "allow connecting AI Copilot to non-localhost Ollama / LM Studio URLs")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSL Studio — Local-First Visual Language Lab for go-dsl (with Ollama & LM Studio)\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [-port 8420] [-dsl grammar.yaml] [-open]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	srv, err := newStudioServer(dslFile, allowRemoteAI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing DSL Studio: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	url := fmt.Sprintf("http://%s", addr)

	fmt.Printf("\n  ✦ go-dsl Studio v1.5 — Organic Language Lab\n")
	fmt.Printf("  ➜ Local IDE:     %s\n", url)
	if dslFile != "" {
		fmt.Printf("  ➜ Editing file:  %s\n", dslFile)
	}
	fmt.Printf("  ➜ Local AI Hub:  Ollama (127.0.0.1:11434) · LM Studio (127.0.0.1:1234)\n\n")

	if openBrowser {
		go func() {
			time.Sleep(250 * time.Millisecond)
			launchBrowser(url)
		}()
	}

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "DSL Studio server error: %v\n", err)
		os.Exit(1)
	}
}

func launchBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
