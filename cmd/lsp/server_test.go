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
