// Command dsltest runs declarative grammar test suites, synthesizes valid
// sample inputs, and fuzzes grammars for panics or hangs.
//
// Usage:
//
//	dsltest -dsl grammar.yaml
//	dsltest -dsl grammar.yaml -fuzz 250 -samples 5 -json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type cliSummary struct {
	DSL      string                    `json:"dsl"`
	Tests    *dslbuilder.DSLTestReport `json:"tests,omitempty"`
	Samples  []string                  `json:"samples,omitempty"`
	Fuzz     *dslbuilder.FuzzReport    `json:"fuzz,omitempty"`
	Warnings []string                  `json:"warnings,omitempty"`
	Valid    bool                      `json:"valid"`
}

func main() {
	var (
		dslFile    string
		fuzzIters  int
		numSamples int
		seed       int64
		jsonOut    bool
	)
	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.IntVar(&fuzzIters, "fuzz", 0, "number of fuzzing iterations to run (0 = disabled)")
	flag.IntVar(&numSamples, "samples", 0, "number of valid sample DSL programs to synthesize")
	flag.Int64Var(&seed, "seed", 42, "deterministic RNG seed for sample generation and fuzzing")
	flag.BoolVar(&jsonOut, "json", false, "emit JSON report")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dsltest - Declarative grammar test runner, sample synthesizer & fuzzer\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -dsl grammar.yaml [-fuzz 200] [-samples 5] [-seed 42] [-json]\n\n", os.Args[0])
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

	ok := runSuite(dsl, fuzzIters, numSamples, seed, jsonOut, os.Stdout)
	if !ok {
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

func runSuite(dsl *dslbuilder.DSL, fuzzIters, numSamples int, seed int64, jsonOut bool, out io.Writer) bool {
	summary := cliSummary{
		DSL:   dsl.ToConfig().Name,
		Valid: true,
	}

	warnings, valErr := dsl.Validate()
	summary.Warnings = warnings
	if valErr != nil {
		summary.Valid = false
	}

	report := dsl.RunConfigTests()
	summary.Tests = &report
	if report.Failed > 0 {
		summary.Valid = false
	}

	if numSamples > 0 {
		summary.Samples = dsl.GenerateSamples(seed, numSamples, 4)
	}

	if fuzzIters > 0 {
		fuzzRep := dsl.FuzzGrammar(seed, fuzzIters)
		summary.Fuzz = &fuzzRep
		if len(fuzzRep.Failures) > 0 {
			summary.Valid = false
		}
	}

	if jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(summary)
		return summary.Valid
	}

	if valErr != nil {
		fmt.Fprintf(out, "✗ Grammar validation failed: %v\n", valErr)
	} else {
		fmt.Fprintf(out, "✓ Grammar %q validated\n", summary.DSL)
	}

	if report.Total > 0 {
		fmt.Fprintf(out, "\nDeclarative Tests: %d/%d passed\n", report.Passed, report.Total)
		for _, r := range report.Results {
			if r.Passed {
				fmt.Fprintf(out, "  ✓ %s\n", r.Name)
			} else {
				fmt.Fprintf(out, "  ✗ %s: %s\n", r.Name, r.Message)
			}
		}
	} else {
		fmt.Fprintf(out, "\nDeclarative Tests: 0 cases defined in grammar\n")
	}

	if len(summary.Samples) > 0 {
		fmt.Fprintf(out, "\nSynthesized Valid Samples (%d):\n", len(summary.Samples))
		for i, s := range summary.Samples {
			fmt.Fprintf(out, "  %d. %s\n", i+1, s)
		}
	}

	if summary.Fuzz != nil {
		fmt.Fprintf(out, "\nFuzzing: %d iterations (%d parsed, %d rejected cleanly, %d failures)\n",
			summary.Fuzz.Iterations, summary.Fuzz.ValidParsed, summary.Fuzz.RejectedCleanly, len(summary.Fuzz.Failures))
		for _, f := range summary.Fuzz.Failures {
			fmt.Fprintf(out, "  ✗ iter %d (%q): %s\n", f.Iteration, f.Input, f.Reason)
		}
	}

	return summary.Valid
}
