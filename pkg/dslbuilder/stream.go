// Package dslbuilder - Streaming statement-by-statement parsing.
//
// ParseStream processes line-oriented DSL scripts from an io.Reader without
// loading the whole input into memory: each non-empty, non-comment line is
// parsed and evaluated as one statement and handed to a callback. This is
// the streaming counterpart of ParseMultiline, suitable for large script
// files, pipes, and network streams.
package dslbuilder

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// maxStreamLine bounds a single statement line (1 MB): protects against
// unbounded memory use on malformed streams.
const maxStreamLine = 1 << 20

// StreamHandler receives the result of each parsed statement.
// line is the 1-based line number in the stream. Returning an error stops
// the stream and propagates the error to the ParseStream caller.
type StreamHandler func(line int, result *Result) error

// ParseStream reads DSL statements line by line from r, parsing and
// evaluating each one, and invokes handler with every result.
//
// Behavior (matches ParseMultiline):
//   - Empty lines and comments (# or //) are skipped.
//   - Each remaining line is parsed as one statement.
//   - The first parse error stops the stream; the error carries the line
//     number. (Use Diagnostics on a full string when you want all errors.)
//
// Example:
//
//	f, _ := os.Open("script.dsl")
//	defer f.Close()
//	err := dsl.ParseStream(f, func(line int, result *dslbuilder.Result) error {
//	    fmt.Printf("line %d => %v\n", line, result.GetOutput())
//	    return nil
//	})
func (d *DSL) ParseStream(r io.Reader, handler StreamHandler) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxStreamLine)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		result, err := d.Parse(line)
		if err != nil {
			if parseErr, ok := err.(*ParseError); ok {
				parseErr.Line = lineNum
				return parseErr
			}
			return fmt.Errorf("line %d: %w", lineNum, err)
		}

		if handler != nil {
			if err := handler(lineNum, result); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stream: %w", err)
	}
	return nil
}
