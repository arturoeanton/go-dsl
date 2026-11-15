package wasm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWASMEngineEndToEnd(t *testing.T) {
	eng := NewEngine()

	// Calling before loading returns clean JSON error.
	var pre Response
	require.NoError(t, json.Unmarshal([]byte(eng.Parse("1 + 2", 100)), &pre))
	assert.False(t, pre.OK)

	yamlSpec := `
name: calc
token_defs:
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+"
  - name: PLUS
    kind: literal
    pattern: "+"
  - name: STAR
    kind: literal
    pattern: "*"
  - name: COMMENT
    kind: skip
    pattern: "//[^\\r\\n]*"
expressions:
  - name: expr
    atoms:
      - token: NUMBER
    infix:
      - token: PLUS
        power: 10
        assoc: left
      - token: STAR
        power: 20
        assoc: left
tests:
  - name: precedence
    input: "1 + 2 * 3"
    should_parse: true
  - name: bad syntax
    input: "1 + * 3"
    should_parse: false
`
	var loadResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.LoadGrammar(yamlSpec, "yaml")), &loadResp))
	assert.True(t, loadResp.OK, loadResp.Error)

	// Parse + AST + Trace
	var parseResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.Parse("1 + 2 * 3 // comment", 500)), &parseResp))
	assert.True(t, parseResp.OK, parseResp.Error)
	require.NotNil(t, parseResp.AST)
	require.NotNil(t, parseResp.Trace)
	assert.Greater(t, parseResp.Trace.Steps, 0)

	// Diagnostics on broken input
	var diagResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.Diagnostics("1 + @ 2")), &diagResp))
	assert.True(t, diagResp.OK)
	assert.NotEmpty(t, diagResp.Diagnostics)

	// Completions
	var compResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.Completions("1 + ", 4)), &compResp))
	assert.True(t, compResp.OK)
	assert.NotEmpty(t, compResp.Completions)

	// Format preserving comment
	var fmtResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.Format("1+2*3 // inline")), &fmtResp))
	assert.True(t, fmtResp.OK, fmtResp.Error)
	assert.Equal(t, "1 + 2 * 3 // inline", fmtResp.Formatted)

	// GBNF & AI Prompt Spec
	var gbnfResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.ToGBNF()), &gbnfResp))
	assert.True(t, gbnfResp.OK)
	assert.Contains(t, gbnfResp.GBNF, "root ::=")

	var promptResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.AIPromptSpec()), &promptResp))
	assert.True(t, promptResp.OK)
	assert.Contains(t, promptResp.PromptSpec, "DSL Specification")

	// Samples & Embedded Tests
	var sampleResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.GenerateSamples(1, 3)), &sampleResp))
	assert.True(t, sampleResp.OK)
	assert.NotEmpty(t, sampleResp.Samples)

	var testResp Response
	require.NoError(t, json.Unmarshal([]byte(eng.RunTests()), &testResp))
	assert.True(t, testResp.OK)
	require.NotNil(t, testResp.TestReport)
	assert.Equal(t, 2, testResp.TestReport.Passed)
}
