package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func doRequest(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

type handlerRoundTripper struct {
	routes map[string]http.Handler
}

func (rt *handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.URL.Scheme + "://" + req.URL.Host
	h, ok := rt.routes[key]
	if !ok {
		return nil, fmt.Errorf("no mock handler registered for %s", key)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result(), nil
}

func TestPresetsAllLoadEvaluateAndPassEmbeddedTests(t *testing.T) {
	srv, err := newStudioServer("", false)
	if err != nil {
		t.Fatalf("newStudioServer failed: %v", err)
	}
	h := srv.routes()

	// 1. Embedded UI assets serve cleanly.
	for _, path := range []string{"/", "/styles.css", "/app.js"} {
		rec := doRequest(t, h, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK || rec.Body.Len() == 0 {
			t.Fatalf("GET %s returned status %d, len %d", path, rec.Code, rec.Body.Len())
		}
	}

	// 2. Every builtin preset loads, evaluates its sample code, and passes embedded tests + fuzzer.
	for _, p := range builtinPresets() {
		t.Run(p.ID, func(t *testing.T) {
			evalPayload, _ := json.Marshal(map[string]interface{}{
				"grammar":   p.GrammarYAML,
				"code":      p.SampleCode,
				"max_steps": 5000,
			})
			rec := doRequest(t, h, http.MethodPost, "/api/studio/eval", evalPayload)
			var evalRes map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &evalRes); err != nil {
				t.Fatalf("invalid eval JSON: %v", err)
			}

			if evalRes["ok"] != true {
				t.Fatalf("preset %s sample failed to evaluate: %v", p.ID, evalRes["error"])
			}
			if evalRes["ast"] == nil || evalRes["trace"] == nil {
				t.Fatalf("preset %s missing AST or Trace", p.ID)
			}

			// Run embedded tests + mini fuzzing
			testPayload, _ := json.Marshal(map[string]interface{}{
				"fuzz_iterations": 40,
				"sample_count":    2,
				"seed":            42,
			})
			tRec := doRequest(t, h, http.MethodPost, "/api/studio/tests", testPayload)
			var testRes map[string]interface{}
			if err := json.Unmarshal(tRec.Body.Bytes(), &testRes); err != nil {
				t.Fatalf("invalid tests JSON: %v", err)
			}

			if testRes["ok"] != true {
				t.Fatalf("preset %s embedded tests/fuzz failed: %+v", p.ID, testRes)
			}
		})
	}

	// 3. Verify typed Go codegen & VS Code .zip archive download.
	cgPayload, _ := json.Marshal(map[string]interface{}{
		"package_name": "studiogen",
		"func_name":    "NewStudioDSL",
		"typed":        true,
	})
	cgRec := doRequest(t, h, http.MethodPost, "/api/studio/codegen", cgPayload)
	var cgData map[string]interface{}
	_ = json.Unmarshal(cgRec.Body.Bytes(), &cgData)
	if cgData["ok"] != true || !strings.Contains(cgData["go_source"].(string), "type StudioDSLActions interface") {
		t.Fatalf("unexpected codegen response: %+v", cgData)
	}

	zipRec := doRequest(t, h, http.MethodPost, "/api/studio/vscode.zip", nil)
	zipBytes := zipRec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil || len(zr.File) < 4 {
		t.Fatalf("invalid VS Code zip bundle: err=%v, files=%d", err, len(zr.File))
	}
}

func TestStudioLocalAICopilotOllamaAndLMStudioClosedLoop(t *testing.T) {
	// Mock Ollama handler (http://127.0.0.1:11434):
	// - Rejects raw "grammar" on /v1/chat/completions (simulating GBNF fallback)
	// - First /api/chat attempt returns broken DSL syntax ("when 500 >=")
	// - Second /api/chat attempt (after receiving RepairPrompt!) returns valid DSL ("when 500 >= 100 apply 450")
	var chatCalls int32
	mockOllama := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"models": []map[string]interface{}{
					{
						"name": "qwen2.5-coder:7b",
						"details": map[string]string{
							"parameter_size":     "7B",
							"quantization_level": "Q4_K_M",
						},
					},
				},
			})
		case "/v1/chat/completions":
			// Simulate backend rejecting raw GBNF so fallback to /api/chat triggers cleanly
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"grammar not supported on this backend"}`))
		case "/api/chat":
			n := atomic.AddInt32(&chatCalls, 1)
			reply := "when 500 >=" // broken on attempt 1!
			if n >= 2 {
				reply = "```dsl\nwhen 500 >= 100   apply 450\n```" // valid on attempt 2!
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"message": map[string]string{"content": reply},
			})
		}
	})

	// Mock LM Studio handler (http://127.0.0.1:1234) supporting GBNF directly on /v1/chat/completions
	mockLMStudio := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]string{
					{"id": "lmstudio-community/Qwen2.5-Coder-14B-Instruct-GGUF"},
				},
			})
		case "/v1/chat/completions":
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["grammar"] == nil || body["grammar"] == "" {
				t.Errorf("expected GBNF grammar to be sent to LM Studio")
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]string{"content": "when 2000 >= 1000 apply 1700"}},
				},
			})
		}
	})

	srv, err := newStudioServer("", false)
	if err != nil {
		t.Fatal(err)
	}
	srv.ai.httpClient = &http.Client{
		Transport: &handlerRoundTripper{
			routes: map[string]http.Handler{
				"http://127.0.0.1:11434": mockOllama,
				"http://127.0.0.1:1234":  mockLMStudio,
			},
		},
	}
	h := srv.routes()

	// 1. Probe both local AI servers
	statusPayload, _ := json.Marshal(map[string]string{
		"ollama_url":   "http://127.0.0.1:11434",
		"lmstudio_url": "http://127.0.0.1:1234",
	})
	stRec := doRequest(t, h, http.MethodPost, "/api/ai/status", statusPayload)
	var st AIStatusResponse
	_ = json.Unmarshal(stRec.Body.Bytes(), &st)

	if !st.Ollama.Online || st.Ollama.Recommended != "qwen2.5-coder:7b" {
		t.Fatalf("unexpected Ollama probe status: %+v", st.Ollama)
	}
	if !st.LMStudio.Online || len(st.LMStudio.Models) != 1 {
		t.Fatalf("unexpected LM Studio probe status: %+v", st.LMStudio)
	}

	// 2. Closed-loop code generation via Ollama (GBNF fallback + 2-attempt RepairPrompt self-healing + Format!)
	genPayload, _ := json.Marshal(AIRunRequest{
		Provider: "ollama",
		BaseURL:  "http://127.0.0.1:11434",
		Model:    "qwen2.5-coder:7b",
		Mode:     "generate_code",
		Prompt:   "aplicar 450 cuando supere 100",
		UseGBNF:  true,
	})
	genRec := doRequest(t, h, http.MethodPost, "/api/ai/run", genPayload)
	var aiOut AIRunResponse
	_ = json.Unmarshal(genRec.Body.Bytes(), &aiOut)

	if !aiOut.OK || !aiOut.Verified || aiOut.Attempts != 2 || !aiOut.GBNFFallback {
		t.Fatalf("expected verified 2-attempt self-healed result, got: %+v", aiOut)
	}
	if aiOut.ResultCode != "when 500 >= 100 apply 450" {
		t.Fatalf("expected canonically formatted code, got %q", aiOut.ResultCode)
	}

	// 3. Direct GBNF constrained decoding via LM Studio (1st attempt verified!)
	lmPayload, _ := json.Marshal(AIRunRequest{
		Provider: "lmstudio",
		BaseURL:  "http://127.0.0.1:1234",
		Model:    "lmstudio-community/Qwen2.5-Coder-14B-Instruct-GGUF",
		Mode:     "generate_code",
		Prompt:   "descuento vip",
		UseGBNF:  true,
	})
	lmRec := doRequest(t, h, http.MethodPost, "/api/ai/run", lmPayload)
	var lmOut AIRunResponse
	_ = json.Unmarshal(lmRec.Body.Bytes(), &lmOut)

	if !lmOut.OK || !lmOut.Verified || lmOut.Attempts != 1 || !lmOut.GBNFUsed {
		t.Fatalf("expected 1st-attempt GBNF-constrained LM Studio result, got: %+v", lmOut)
	}

	// 4. Security check: remote non-loopback AI URL is rejected by default!
	evilPayload, _ := json.Marshal(AIRunRequest{
		Provider: "ollama",
		BaseURL:  "http://169.254.169.254:11434",
		Model:    "qwen2.5-coder:7b",
		Mode:     "generate_code",
		Prompt:   "test",
	})
	evilRec := doRequest(t, h, http.MethodPost, "/api/ai/run", evilPayload)
	var evilOut AIRunResponse
	_ = json.Unmarshal(evilRec.Body.Bytes(), &evilOut)
	if evilOut.OK || !strings.Contains(evilOut.Error, "localhost") {
		t.Fatalf("expected SSRF loopback guard to reject remote AI URL, got: %+v", evilOut)
	}

	// 5. File save endpoint works when initialized with -dsl path.
	tmpGrammar := filepath.Join(t.TempDir(), "edit_in_place.yaml")
	if err := os.WriteFile(tmpGrammar, []byte(builtinPresets()[0].GrammarYAML), 0644); err != nil {
		t.Fatal(err)
	}
	fileSrv, err := newStudioServer(tmpGrammar, false)
	if err != nil {
		t.Fatal(err)
	}
	fh := fileSrv.routes()

	savePayload, _ := json.Marshal(map[string]string{"grammar": builtinPresets()[1].GrammarYAML})
	_ = doRequest(t, fh, http.MethodPost, "/api/studio/save", savePayload)
	savedBytes, _ := os.ReadFile(tmpGrammar)
	if !strings.Contains(string(savedBytes), "SmartQuery") {
		t.Fatalf("expected file to be updated on disk")
	}
}
