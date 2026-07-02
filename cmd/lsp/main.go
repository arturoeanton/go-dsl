// Command lsp is a minimal Language Server Protocol server for DSLs built
// with go-dsl. Point it at a declarative grammar and connect it to any
// LSP-capable editor:
//
//   - Diagnostics: every open document is re-parsed incrementally
//     (dslbuilder.Document — only edited statements re-parse) and ALL
//     syntax/lexical errors are published, not just the first.
//   - Completion: textDocument/completion suggests what the parser itself
//     expects at the cursor (keywords as text, free-form tokens as
//     placeholder snippets).
//   - Hover: textDocument/hover shows the AST node under the cursor
//     (token, or rule + action + source snippet).
//
// Usage:
//
//	lsp -dsl grammar.yaml
//
// Wire it as a stdio language server in your editor. Implemented methods:
// initialize, shutdown/exit, textDocument/didOpen, didChange, didClose,
// textDocument/completion, textDocument/hover.
// No external dependencies — the JSON-RPC framing is implemented inline.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	var dslFile string
	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "lsp - Minimal LSP server for go-dsl grammars\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -dsl grammar.yaml\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	dsl, err := loadDSL(dslFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading grammar: %v\n", err)
		os.Exit(1)
	}

	server := newServer(dsl, os.Stdin, os.Stdout)
	if err := server.run(); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func loadDSL(filename string) (*dslbuilder.DSL, error) {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	switch ext {
	case "yaml", "yml":
		return dslbuilder.LoadFromYAMLFile(filename)
	case "json":
		return dslbuilder.LoadFromJSONFile(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// --- LSP wire types (the minimal subset we implement) ---------------------

type rpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

type rpcNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"` // 1 = Error
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

type publishDiagnosticsParams struct {
	URI         string          `json:"uri"`
	Diagnostics []lspDiagnostic `json:"diagnostics"`
}

type textDocumentParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
	Position lspPosition `json:"position"`
}

type lspCompletionItem struct {
	Label            string `json:"label"`
	Kind             int    `json:"kind"` // 14 = Keyword, 1 = Text
	Detail           string `json:"detail,omitempty"`
	InsertText       string `json:"insertText,omitempty"`
	InsertTextFormat int    `json:"insertTextFormat,omitempty"` // 1 = plain, 2 = snippet
}

type lspHover struct {
	Contents lspMarkup `json:"contents"`
	Range    *lspRange `json:"range,omitempty"`
}

type lspMarkup struct {
	Kind  string `json:"kind"` // "markdown"
	Value string `json:"value"`
}
