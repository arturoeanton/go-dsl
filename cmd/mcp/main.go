// Command mcp is a Model Context Protocol (MCP) server over stdio JSON-RPC 2.0
// for any go-dsl grammar. Connect it to Claude Code, Cursor, Windsurf, or
// custom AI agents so the LLM can inspect the grammar, validate generated DSL
// code with self-repair prompts, compute completions, format code, and export
// GBNF grammar constraints.
//
// Usage:
//
//	mcp -dsl grammar.yaml
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	var dslFile string
	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "mcp - Model Context Protocol (MCP) server for go-dsl grammars\n\n")
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

	srv := newMCPServer(dsl, os.Stdin, os.Stdout)
	if err := srv.run(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

func loadDSL(filename string) (*dslbuilder.DSL, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch ext {
	case "yaml", "yml":
		return dslbuilder.LoadFromYAMLFile(filename)
	case "json":
		return dslbuilder.LoadFromJSONFile(filename)
	default:
		if ext == "" {
			return nil, fmt.Errorf("missing file extension in %q (expected .yaml, .yml, or .json)", filename)
		}
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}
