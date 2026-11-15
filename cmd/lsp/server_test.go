package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func testDSL(t *testing.T) *dslbuilder.DSL {
	t.Helper()
	d := dslbuilder.New("lsptest")
	if err := d.KeywordToken("SET", "set"); err != nil {
		t.Fatal(err)
	}
	if err := d.Token("IDENT", `[a-z]+`); err != nil {
		t.Fatal(err)
	}
	if err := d.Token("NUMBER", `[0-9]+`); err != nil {
		t.Fatal(err)
	}
	d.Rule("stmt", []string{"SET", "IDENT", "NUMBER"}, "")
	return d
}

func frame(payload string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload)
}

// decodeFrames splits the server's output stream into JSON payloads.
func decodeFrames(t *testing.T, out string) []map[string]interface{} {
	t.Helper()
	var messages []map[string]interface{}
	rest := out
	for len(rest) > 0 {
		var length int
		idx := strings.Index(rest, "\r\n\r\n")
		if idx < 0 {
			break
		}
		header := rest[:idx]
		if _, err := fmt.Sscanf(header, "Content-Length: %d", &length); err != nil {
			t.Fatalf("bad header %q: %v", header, err)
		}
		body := rest[idx+4 : idx+4+length]
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(body), &msg); err != nil {
			t.Fatalf("bad payload %q: %v", body, err)
		}
		messages = append(messages, msg)
		rest = rest[idx+4+length:]
	}
	return messages
}

func TestLSPLifecycleAndDiagnostics(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	// Document with one bad statement ("set y" is missing its number).
	didOpen := `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///test.dsl","text":"set x 1\nset y\nset z 3"}}}`
	input.WriteString(frame(didOpen))
	// Fixed document: diagnostics must clear.
	didChange := `{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///test.dsl"},"contentChanges":[{"text":"set x 1\nset y 2\nset z 3"}]}}`
	input.WriteString(frame(didChange))
	input.WriteString(frame(`{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`))
	input.WriteString(frame(`{"jsonrpc":"2.0","method":"exit"}`))

	var output bytes.Buffer
	srv := newServer(testDSL(t), &input, &output)
	if err := srv.run(); err != nil {
		t.Fatalf("server error: %v", err)
	}

	messages := decodeFrames(t, output.String())
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages (init resp, 2 diagnostics, shutdown resp), got %d", len(messages))
	}

	// 1. initialize response advertises full text sync.
	initResult := messages[0]["result"].(map[string]interface{})
	caps := initResult["capabilities"].(map[string]interface{})
	if caps["textDocumentSync"].(float64) != 1 {
		t.Errorf("expected full textDocumentSync, got %v", caps["textDocumentSync"])
	}

	// 2. didOpen publishes one error diagnostic on line 2 (0-based: 1).
	diag1 := messages[1]
	if diag1["method"] != "textDocument/publishDiagnostics" {
		t.Fatalf("expected publishDiagnostics, got %v", diag1["method"])
	}
	params := diag1["params"].(map[string]interface{})
	diags := params["diagnostics"].([]interface{})
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d: %v", len(diags), diags)
	}
	first := diags[0].(map[string]interface{})
	if !strings.Contains(first["message"].(string), "expected") {
		t.Errorf("diagnostic message should mention expectation: %v", first["message"])
	}
	startLine := first["range"].(map[string]interface{})["start"].(map[string]interface{})["line"].(float64)
	if startLine < 1 {
		t.Errorf("diagnostic should be at line >= 1 (0-based), got %v", startLine)
	}

	// 3. didChange with the fix publishes an empty diagnostics list.
	diag2 := messages[2]
	params2 := diag2["params"].(map[string]interface{})
	diags2 := params2["diagnostics"].([]interface{})
	if len(diags2) != 0 {
		t.Errorf("expected diagnostics to clear after fix, got %v", diags2)
	}

	// 4. shutdown gets a response.
	if messages[3]["id"].(float64) != 2 {
		t.Errorf("expected shutdown response with id 2, got %v", messages[3])
	}
}

func TestLSPCompletionAndHover(t *testing.T) {
	var input bytes.Buffer
	input.WriteString(frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	// Document: "set x 1\nset y " — cursor at end of line 2 expects NUMBER.
	didOpen := `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///t.dsl","text":"set x 1\nset y "}}}`
	input.WriteString(frame(didOpen))
	completion := `{"jsonrpc":"2.0","id":2,"method":"textDocument/completion","params":{"textDocument":{"uri":"file:///t.dsl"},"position":{"line":1,"character":6}}}`
	input.WriteString(frame(completion))
	// Hover over "x" (line 0, char 4).
	hover := `{"jsonrpc":"2.0","id":3,"method":"textDocument/hover","params":{"textDocument":{"uri":"file:///t.dsl"},"position":{"line":0,"character":4}}}`
	input.WriteString(frame(hover))
	input.WriteString(frame(`{"jsonrpc":"2.0","method":"exit"}`))

	var output bytes.Buffer
	srv := newServer(testDSL(t), &input, &output)
	if err := srv.run(); err != nil {
		t.Fatalf("server error: %v", err)
	}

	messages := decodeFrames(t, output.String())
	// init resp, diagnostics (1 error: incomplete stmt), completion resp, hover resp
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(messages))
	}

	// Capabilities advertise completion and hover.
	caps := messages[0]["result"].(map[string]interface{})["capabilities"].(map[string]interface{})
	if caps["hoverProvider"] != true {
		t.Errorf("hoverProvider not advertised")
	}
	if _, ok := caps["completionProvider"]; !ok {
		t.Errorf("completionProvider not advertised")
	}

	// Completion result contains a NUMBER suggestion as a snippet.
	items := messages[2]["result"].([]interface{})
	var hasNumber bool
	for _, it := range items {
		item := it.(map[string]interface{})
		if item["label"] == "NUMBER" {
			hasNumber = true
			if item["insertText"] != "${1:NUMBER}" {
				t.Errorf("free-form token should insert as snippet placeholder, got %v", item["insertText"])
			}
			if item["insertTextFormat"].(float64) != 2 {
				t.Errorf("free-form token should use snippet format, got %v", item["insertTextFormat"])
			}
		}
	}
	if !hasNumber {
		t.Errorf("completion should suggest NUMBER, got %v", items)
	}

	// Hover result describes the IDENT token "x".
	hoverResult := messages[3]["result"].(map[string]interface{})
	value := hoverResult["contents"].(map[string]interface{})["value"].(string)
	if !strings.Contains(value, "IDENT") || !strings.Contains(value, "x") {
		t.Errorf("hover should describe IDENT token x, got %q", value)
	}
}

func TestBug13UTF16PositionsAndLineClamping(t *testing.T) {
	// Line 0 has a surrogate-pair emoji ("🚀" = 4 UTF-8 bytes, 2 UTF-16 code units).
	text := `say "🚀" 42` + "\n" + `say "ok" 7`

	// Clamping: character 999 on line 0 must stay at the end of line 0, never bleeding into line 1.
	clamped := offsetForPosition(text, lspPosition{Line: 0, Character: 999})
	line0Len := len(`say "🚀" 42`)
	if clamped != line0Len {
		t.Fatalf("expected clamped offset %d at end of line 0, got %d", line0Len, clamped)
	}

	// Offset of "42" on line 0: `say "` (5 UTF-16) + `🚀` (2 UTF-16) + `" ` (2 UTF-16) = 9 UTF-16 units,
	// whereas in UTF-8 bytes it is 5 + 4 + 2 = 11 bytes.
	pos42 := positionForOffset(text, 11)
	if pos42.Line != 0 || pos42.Character != 9 {
		t.Fatalf("expected UTF-16 position (0, 9), got (%d, %d)", pos42.Line, pos42.Character)
	}
	backOffset := offsetForPosition(text, pos42)
	if backOffset != 11 {
		t.Fatalf("expected UTF-16 round-trip byte offset 11, got %d", backOffset)
	}
}

func TestLSPFormattingAndSemanticTokens(t *testing.T) {
	d := dslbuilder.New("fmtlsp")
	if err := d.SkipToken("COMMENT", `//[^\n]*`); err != nil {
		t.Fatal(err)
	}
	if err := d.KeywordToken("SET", "set"); err != nil {
		t.Fatal(err)
	}
	if err := d.LiteralToken("EQ", "="); err != nil {
		t.Fatal(err)
	}
	if err := d.Token("IDENT", `[a-z]+`); err != nil {
		t.Fatal(err)
	}
	if err := d.Token("NUMBER", `[0-9]+`); err != nil {
		t.Fatal(err)
	}
	d.Rule("stmt", []string{"SET", "IDENT", "EQ", "NUMBER"}, "")

	var input bytes.Buffer
	input.WriteString(frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	input.WriteString(frame(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///f.dsl","text":"set   x=42 // comment"}}}`))
	input.WriteString(frame(`{"jsonrpc":"2.0","id":2,"method":"textDocument/formatting","params":{"textDocument":{"uri":"file:///f.dsl"}}}`))
	input.WriteString(frame(`{"jsonrpc":"2.0","id":3,"method":"textDocument/semanticTokens/full","params":{"textDocument":{"uri":"file:///f.dsl"}}}`))
	input.WriteString(frame(`{"jsonrpc":"2.0","method":"exit"}`))

	var output bytes.Buffer
	srv := newServer(d, &input, &output)
	if err := srv.run(); err != nil {
		t.Fatalf("server error: %v", err)
	}

	messages := decodeFrames(t, output.String())
	if len(messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(messages))
	}

	caps := messages[0]["result"].(map[string]interface{})["capabilities"].(map[string]interface{})
	if caps["documentFormattingProvider"] != true {
		t.Errorf("documentFormattingProvider not advertised")
	}
	if _, ok := caps["semanticTokensProvider"]; !ok {
		t.Errorf("semanticTokensProvider not advertised")
	}

	edits := messages[2]["result"].([]interface{})
	if len(edits) != 1 {
		t.Fatalf("expected 1 formatting TextEdit, got %v", edits)
	}
	newText := edits[0].(map[string]interface{})["newText"].(string)
	if !strings.Contains(newText, "set x = 42 // comment") {
		t.Errorf("unexpected formatted text: %q", newText)
	}

	sem := messages[3]["result"].(map[string]interface{})["data"].([]interface{})
	// 5 tokens ("set", "x", "=", "42", "// comment") * 5 ints each = 25 ints.
	if len(sem) != 25 {
		t.Fatalf("expected 25 semantic token ints, got %d: %v", len(sem), sem)
	}
}
