package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func TestMCPServerToolsLifecycle(t *testing.T) {
	d := dslbuilder.New("BillingDSL")
	_ = d.KeywordToken("CHARGE", "charge")
	_ = d.Token("AMOUNT", `[0-9]+`)
	d.Rule("stmt", []string{"CHARGE", "AMOUNT"}, "")

	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"dsl_gbnf","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"dsl_validate_code","arguments":{"code":"charge 250"}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"dsl_validate_code","arguments":{"code":"charge "}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"dsl_format","arguments":{"code":"charge    250"}}}`,
	}

	var in, out bytes.Buffer
	in.WriteString(strings.Join(requests, "\n") + "\n")

	srv := newMCPServer(d, &in, &out)
	if err := srv.run(); err != nil {
		t.Fatalf("mcp run failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 6 { // notification produces no response
		t.Fatalf("expected 6 JSON-RPC responses, got %d:\n%s", len(lines), out.String())
	}

	// Verify dsl_gbnf response (#3 -> lines[2])
	var gbnfResp map[string]interface{}
	if err := json.Unmarshal([]byte(lines[2]), &gbnfResp); err != nil {
		t.Fatal(err)
	}
	gbnfText := gbnfResp["result"].(map[string]interface{})["content"].([]interface{})[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(gbnfText, "root ::=") {
		t.Errorf("expected GBNF root production, got:\n%s", gbnfText)
	}

	// Verify valid code (#4 -> lines[3]) vs invalid code (#5 -> lines[4])
	var validResp, invalidResp map[string]interface{}
	_ = json.Unmarshal([]byte(lines[3]), &validResp)
	_ = json.Unmarshal([]byte(lines[4]), &invalidResp)

	if validResp["result"].(map[string]interface{})["isError"] == true {
		t.Errorf("expected charge 250 to validate cleanly: %v", validResp)
	}
	if invalidResp["result"].(map[string]interface{})["isError"] != true {
		t.Errorf("expected incomplete code 'charge ' to return isError=true: %v", invalidResp)
	}
}
