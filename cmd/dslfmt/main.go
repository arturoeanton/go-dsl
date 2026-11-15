// Command dslfmt formats DSL source files using a declarative go-dsl grammar.
//
// Usage:
//
//	dslfmt -dsl grammar.yaml [file.dsl ...]
//	dslfmt -dsl grammar.yaml -w file.dsl
//	dslfmt -dsl grammar.yaml -check file.dsl
//	cat file.dsl | dslfmt -dsl grammar.yaml
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	var (
		dslFile   string
		writeBack bool
		checkOnly bool
	)
	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.BoolVar(&writeBack, "w", false, "write formatted result back to source file(s)")
	flag.BoolVar(&checkOnly, "check", false, "exit with non-zero status if any file is not canonically formatted")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dslfmt - Universal AST + comment-preserving formatter for go-dsl\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -dsl grammar.yaml [-w] [-check] [file.dsl ...]\n\n", os.Args[0])
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

	files := flag.Args()
	if len(files) == 0 {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
		formatted, err := dsl.Format(string(raw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "format error: %v\n", dslbuilder.GetDetailedError(err))
			os.Exit(1)
		}
		if checkOnly {
			if formatted != string(raw) {
				os.Exit(1)
			}
			return
		}
		fmt.Print(formatted)
		return
	}

	unformatted := 0
	for _, f := range files {
		changed, err := processFile(dsl, f, writeBack, checkOnly, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			os.Exit(1)
		}
		if changed {
			unformatted++
			if checkOnly {
				fmt.Fprintln(os.Stdout, f)
			}
		}
	}
	if checkOnly && unformatted > 0 {
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

func processFile(dsl *dslbuilder.DSL, path string, writeBack, checkOnly bool, out io.Writer) (changed bool, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	formatted, err := dsl.Format(string(raw))
	if err != nil {
		return false, fmt.Errorf("%s", dslbuilder.GetDetailedError(err))
	}
	changed = formatted != string(raw)
	if checkOnly {
		return changed, nil
	}
	if writeBack {
		if changed {
			if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
				return changed, err
			}
		}
		return changed, nil
	}
	_, err = io.WriteString(out, formatted)
	return changed, err
}
