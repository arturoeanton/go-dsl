package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func TestDSLFmtProcessFileAndIdempotence(t *testing.T) {
	d := dslbuilder.New("FmtCLI")
	_ = d.SkipToken("COMMENT", `//[^\n]*`)
	_ = d.KeywordToken("SET", "set")
	_ = d.LiteralToken("EQ", "=")
	_ = d.Token("IDENT", `[a-z]+`)
	_ = d.Token("NUMBER", `[0-9]+`)
	d.Rule("stmt", []string{"SET", "IDENT", "EQ", "NUMBER"}, "")

	dir := t.TempDir()
	srcFile := filepath.Join(dir, "sample.dsl")
	if err := os.WriteFile(srcFile, []byte("set    alpha=42   // keep comment\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Check-only before formatting must report changed=true without modifying file.
	var buf bytes.Buffer
	changed, err := processFile(d, srcFile, false, true, &buf)
	if err != nil {
		t.Fatalf("processFile check failed: %v", err)
	}
	if !changed {
		t.Fatal("expected unformatted file to report changed=true")
	}

	// Write-back formats the file in place.
	changed, err = processFile(d, srcFile, true, false, &buf)
	if err != nil {
		t.Fatalf("processFile -w failed: %v", err)
	}
	if !changed {
		t.Fatal("expected -w to report changed=true")
	}

	after, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "set alpha = 42 // keep comment") {
		t.Fatalf("unexpected formatted file content: %q", string(after))
	}

	// Second check-only must report changed=false (idempotent!).
	changed, err = processFile(d, srcFile, false, true, &buf)
	if err != nil {
		t.Fatalf("second check failed: %v", err)
	}
	if changed {
		t.Fatal("expected second check to be idempotent (changed=false)")
	}
}
