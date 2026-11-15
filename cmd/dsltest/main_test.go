package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func TestDSLTestRunnerAndFuzzerCLI(t *testing.T) {
	yamlSpec := []byte(`
name: OrderFilter
token_defs:
  - name: WHERE
    kind: keyword
    pattern: where
  - name: EQ
    kind: literal
    pattern: "=="
  - name: IDENT
    kind: regex
    pattern: "[a-z]+"
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+"
rules:
  - name: query
    pattern: [WHERE, IDENT, EQ, NUMBER]
    action: filter_eq
tests:
  - name: valid filter
    input: "where total == 100"
    expect_ast: "query"
  - name: invalid missing rhs
    input: "where total =="
    should_parse: false
`)
	dsl, err := dslbuilder.LoadFromYAML(yamlSpec)
	if err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}

	var out bytes.Buffer
	ok := runSuite(dsl, 50, 3, 42, true, &out)
	if !ok {
		t.Fatalf("expected suite + fuzzer to pass, output:\n%s", out.String())
	}

	var rep cliSummary
	if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
		t.Fatalf("invalid JSON summary: %v", err)
	}
	if rep.Tests == nil || rep.Tests.Passed != 2 || rep.Tests.Failed != 0 {
		t.Fatalf("unexpected test report: %+v", rep.Tests)
	}
	if len(rep.Samples) == 0 {
		t.Fatal("expected synthesized samples")
	}
	if rep.Fuzz == nil || len(rep.Fuzz.Failures) != 0 {
		t.Fatalf("unexpected fuzz report: %+v", rep.Fuzz)
	}
}
