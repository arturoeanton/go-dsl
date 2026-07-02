// Command apiflow is the product CLI for the HTTP DSL: run, check, and
// inspect .http automation scripts.
//
//	apiflow run script.http [args...]     execute a script
//	apiflow check script.http             validate without executing
//	apiflow version                       print version
//
// apiflow is the productized face of examples/http_dsl (see the module
// README): the DSL engine is go-dsl, the language and runtime live in the
// universal package, and this CLI is the entry point users script against.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arturoeanton/go-dsl/examples/http_dsl/universal"
)

const version = "0.1.0 (HTTP DSL v3.1.1)"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: apiflow run <script.http> [args...]")
			os.Exit(1)
		}
		if err := runScript(os.Args[2], os.Args[3:], false); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}

	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: apiflow check <script.http>")
			os.Exit(1)
		}
		if err := runScript(os.Args[2], nil, true); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ script is valid")

	case "version":
		fmt.Println("apiflow", version)

	case "help", "-h", "--help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

// runScript executes (or, with checkOnly, just validates) a .http script.
func runScript(filename string, args []string, checkOnly bool) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filename, err)
	}
	script := string(content)

	dsl := universal.NewHTTPDSLv3()

	// Script arguments become $ARG1..$ARGn / $ARGC, matching the runner.
	for i, arg := range args {
		dsl.SetVariable(fmt.Sprintf("ARG%d", i+1), arg)
	}
	dsl.SetVariable("ARGC", len(args))

	if checkOnly {
		return validateScript(dsl, script)
	}

	start := time.Now()
	result, err := dsl.ParseWithBlockSupport(script)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Print statement outputs (print results are plain strings).
	if results, ok := result.([]interface{}); ok {
		for _, res := range results {
			if str, ok := res.(string); ok && str != "" {
				if !strings.HasPrefix(str, "HTTP ") &&
					!strings.HasPrefix(str, "Variable set:") &&
					!strings.HasPrefix(str, "Condition evaluated") {
					fmt.Println(str)
				}
			}
		}
	}

	fmt.Printf("✓ completed in %v\n", time.Since(start).Round(time.Millisecond))
	return nil
}

// validateScript checks the script without executing anything:
//
//  1. Block structure: if/endif and loop/endloop nesting must balance, and
//     else/break/continue must appear inside the right construct.
//  2. Statement syntax: every non-structural line is validated through the
//     AST phase only (ValidateStatement), which by design runs no actions —
//     so validation can never fire an HTTP request or mutate variables.
//
// Limitation: statements inside blocks are validated individually; cross-
// statement semantics (e.g. a variable used before it is set at run time)
// are not checked.
func validateScript(dsl *universal.HTTPDSLv3, script string) error {
	var errs []string
	var blocks []string // stack of open constructs: "if" or "loop"

	openLoop := func() bool {
		for _, b := range blocks {
			if b == "loop" {
				return true
			}
		}
		return false
	}

	for i, line := range strings.Split(script, "\n") {
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "if ") && strings.HasSuffix(trimmed, "then"):
			blocks = append(blocks, "if")
		case trimmed == "else":
			if len(blocks) == 0 || blocks[len(blocks)-1] != "if" {
				errs = append(errs, fmt.Sprintf("line %d: 'else' without an open 'if'", lineNo))
			}
		case trimmed == "endif":
			if len(blocks) == 0 || blocks[len(blocks)-1] != "if" {
				errs = append(errs, fmt.Sprintf("line %d: 'endif' without an open 'if'", lineNo))
			} else {
				blocks = blocks[:len(blocks)-1]
			}
		case (strings.HasPrefix(trimmed, "while ") || strings.HasPrefix(trimmed, "foreach ") ||
			strings.HasPrefix(trimmed, "repeat ")) && strings.HasSuffix(trimmed, "do"):
			blocks = append(blocks, "loop")
		case trimmed == "endloop":
			if len(blocks) == 0 || blocks[len(blocks)-1] != "loop" {
				errs = append(errs, fmt.Sprintf("line %d: 'endloop' without an open loop", lineNo))
			} else {
				blocks = blocks[:len(blocks)-1]
			}
		case trimmed == "break" || trimmed == "continue":
			if !openLoop() {
				errs = append(errs, fmt.Sprintf("line %d: '%s' outside of a loop", lineNo, trimmed))
			}
		default:
			// Single-line block statements (if ... then <stmt> [else <stmt>],
			// while ... do ... endloop on one line) parse as one unit.
			if err := dsl.ValidateStatement(trimmed); err != nil {
				errs = append(errs, fmt.Sprintf("line %d: %v", lineNo, err))
			}
		}
	}

	for range blocks {
		errs = append(errs, "unclosed block at end of script (missing 'endif' or 'endloop')")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid script:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func usage() {
	fmt.Println(`apiflow - HTTP automation scripting (powered by go-dsl)

Usage:
  apiflow run <script.http> [args...]   execute a script
  apiflow check <script.http>           validate a script without executing
  apiflow version                       print version

Script arguments are available inside the script as $ARG1..$ARGn and $ARGC.`)
}
