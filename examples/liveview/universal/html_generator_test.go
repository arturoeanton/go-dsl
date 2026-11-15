package universal

import (
	"strings"
	"testing"
)

type sampleUser struct {
	ID    int
	Title string
	Bio   string
}

func TestVuln16XSSEscapingInHTMLGenerator(t *testing.T) {
	hg := NewHTMLGenerator()
	payload := `<script>alert("xss")</script>" onfocus="evil()`

	form, err := hg.GenerateForm(sampleUser{ID: 1, Title: payload, Bio: "ok"}, `submit"><script>`)
	if err != nil {
		t.Fatalf("GenerateForm failed: %v", err)
	}
	if strings.Contains(form, "<script>") {
		t.Fatalf("XSS payload leaked unescaped in GenerateForm:\n%s", form)
	}
	if !strings.Contains(form, "&lt;script&gt;") {
		t.Fatalf("expected HTML-escaped script tag in GenerateForm:\n%s", form)
	}

	table, err := hg.GenerateTable([]sampleUser{{ID: 1, Title: payload, Bio: "b"}}, nil)
	if err != nil {
		t.Fatalf("GenerateTable failed: %v", err)
	}
	if strings.Contains(table, "<script>") {
		t.Fatalf("XSS payload leaked unescaped in GenerateTable:\n%s", table)
	}

	card, err := hg.GenerateCard(sampleUser{ID: 1, Title: payload, Bio: "b"}, "")
	if err != nil {
		t.Fatalf("GenerateCard failed: %v", err)
	}
	if strings.Contains(card, "<script>") {
		t.Fatalf("XSS payload leaked unescaped in GenerateCard:\n%s", card)
	}

	// Layout composition keeps trusted layout slot ('content') intact while escaping 'title'.
	page, err := hg.GeneratePage("layout", map[string]interface{}{
		"title":      payload,
		"csrf_token": "tok\"/><script>",
		"content":    form,
	})
	if err != nil {
		t.Fatalf("GeneratePage failed: %v", err)
	}
	if strings.Contains(page, "<script>") {
		t.Fatalf("XSS payload leaked unescaped in GeneratePage:\n%s", page)
	}
	if !strings.Contains(page, `<form phx-submit=`) {
		t.Fatalf("expected pre-rendered form slot to remain intact in GeneratePage:\n%s", page)
	}
}
