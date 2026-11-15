package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/arturoeanton/go-dsl/pkg/dslgen"
	"github.com/arturoeanton/go-dsl/pkg/wasm"
)

//go:embed web/*
var embeddedWebFS embed.FS

type studioServer struct {
	engine      *wasm.Engine
	ai          *aiBridge
	dslFilePath string
	mu          sync.RWMutex
	lastSpec    string
}

func newStudioServer(dslFilePath string, allowRemoteAI bool) (*studioServer, error) {
	s := &studioServer{
		engine:      wasm.NewEngine(),
		ai:          newAIBridge(allowRemoteAI),
		dslFilePath: dslFilePath,
	}

	if dslFilePath != "" {
		raw, err := os.ReadFile(dslFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", dslFilePath, err)
		}
		s.lastSpec = string(raw)
		_ = s.engine.LoadGrammar(s.lastSpec, "auto")
	} else {
		presets := builtinPresets()
		if len(presets) > 0 {
			s.lastSpec = presets[0].GrammarYAML
			_ = s.engine.LoadGrammar(s.lastSpec, "yaml")
		}
	}

	return s, nil
}

func (s *studioServer) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/studio/init", s.handleInit)
	mux.HandleFunc("/api/studio/load", s.handleLoad)
	mux.HandleFunc("/api/studio/eval", s.handleEval)
	mux.HandleFunc("/api/studio/complete", s.handleComplete)
	mux.HandleFunc("/api/studio/format", s.handleFormat)
	mux.HandleFunc("/api/studio/tests", s.handleTests)
	mux.HandleFunc("/api/studio/codegen", s.handleCodegen)
	mux.HandleFunc("/api/studio/vscode.zip", s.handleVSCodeZip)
	mux.HandleFunc("/api/studio/save", s.handleSave)

	mux.HandleFunc("/api/ai/status", s.handleAIStatus)
	mux.HandleFunc("/api/ai/run", s.handleAIRun)

	subFS, _ := fs.Sub(embeddedWebFS, "web")
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	return mux
}

func (s *studioServer) handleInit(w http.ResponseWriter, r *http.Request) {
	presets := builtinPresets()
	s.mu.RLock()
	spec := s.lastSpec
	filePath := s.dslFilePath
	s.mu.RUnlock()

	sampleCode := ""
	if filePath != "" {
		if d := s.engine.DSL(); d != nil {
			if samples := d.GenerateSamples(42, 1, 4); len(samples) > 0 {
				sampleCode = samples[0]
			}
		}
	} else if len(presets) > 0 {
		sampleCode = presets[0].SampleCode
	}

	aiStatus := s.ai.ProbeAll(r.Context(), "", "")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"presets":      presets,
		"grammar_yaml": spec,
		"sample_code":  sampleCode,
		"file_path":    filePath,
		"ai_status":    aiStatus,
	})
}

func (s *studioServer) handleLoad(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Grammar string `json:"grammar"`
		Format  string `json:"format"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	rawResp := s.engine.LoadGrammar(req.Grammar, req.Format)
	var parsed wasm.Response
	_ = json.Unmarshal([]byte(rawResp), &parsed)
	if d := s.engine.DSL(); d != nil {
		parsed.GBNF = d.ToGBNF()
		parsed.PromptSpec = d.AIPromptSpec()
	}

	s.mu.Lock()
	s.lastSpec = req.Grammar
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, parsed)
}

func (s *studioServer) handleEval(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Grammar  string `json:"grammar"`
		Code     string `json:"code"`
		MaxSteps int    `json:"max_steps"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Grammar) != "" {
		s.mu.RLock()
		same := req.Grammar == s.lastSpec
		s.mu.RUnlock()
		if !same {
			loadRaw := s.engine.LoadGrammar(req.Grammar, "auto")
			var lr wasm.Response
			_ = json.Unmarshal([]byte(loadRaw), &lr)
			if !lr.OK && s.engine.DSL() == nil {
				writeJSON(w, http.StatusOK, lr)
				return
			}
			s.mu.Lock()
			s.lastSpec = req.Grammar
			s.mu.Unlock()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, s.engine.Parse(req.Code, req.MaxSteps))
}

func (s *studioServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code   string `json:"code"`
		Offset int    `json:"offset"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, s.engine.Completions(req.Code, req.Offset))
}

func (s *studioServer) handleFormat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, s.engine.Format(req.Code))
}

func (s *studioServer) handleTests(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FuzzIterations int   `json:"fuzz_iterations"`
		SampleCount    int   `json:"sample_count"`
		Seed           int64 `json:"seed"`
	}
	_ = decodeJSONBody(r, &req)
	if req.Seed == 0 {
		req.Seed = 42
	}

	d := s.engine.DSL()
	if d == nil {
		writeJSON(w, http.StatusOK, wasm.Response{OK: false, Error: "no grammar loaded"})
		return
	}

	testRep := d.RunConfigTests()
	resp := wasm.Response{
		OK:         testRep.Failed == 0,
		TestReport: &testRep,
	}
	if req.FuzzIterations > 0 {
		fuzzRep := d.FuzzGrammar(req.Seed, req.FuzzIterations)
		resp.FuzzReport = &fuzzRep
		if len(fuzzRep.Failures) > 0 {
			resp.OK = false
		}
	}
	if req.SampleCount > 0 {
		resp.Samples = d.GenerateSamples(req.Seed, req.SampleCount, 4)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *studioServer) handleCodegen(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PkgName  string `json:"package_name"`
		FuncName string `json:"func_name"`
		Typed    bool   `json:"typed"`
	}
	_ = decodeJSONBody(r, &req)
	if req.PkgName == "" {
		req.PkgName = "mydsl"
	}
	if req.FuncName == "" {
		req.FuncName = "NewMyDSL"
	}

	d := s.engine.DSL()
	if d == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": "no grammar loaded"})
		return
	}

	cfg := d.ToConfig()
	src, err := dslgen.GenerateWithOptions(&cfg, req.PkgName, req.FuncName, dslgen.GenerateOptions{Typed: req.Typed})
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"go_source": string(src),
		"gbnf":      d.ToGBNF(),
		"prompt":    d.AIPromptSpec(),
	})
}

func (s *studioServer) handleVSCodeZip(w http.ResponseWriter, r *http.Request) {
	d := s.engine.DSL()
	if d == nil {
		http.Error(w, "no grammar loaded", http.StatusBadRequest)
		return
	}
	cfg := d.ToConfig()
	zipBytes, err := dslgen.GenerateVSCodeExtensionZip(&cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slug := dslgen.SlugifyDSLName(cfg.Name)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", slug+"-vscode.zip"))
	_, _ = w.Write(zipBytes)
}

func (s *studioServer) handleSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Grammar string `json:"grammar"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	s.mu.RLock()
	target := s.dslFilePath
	s.mu.RUnlock()

	if target == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    false,
			"error": "Studio no se inició con -dsl <archivo.yaml>. Usa Exportar para descargar tu gramática.",
		})
		return
	}
	if err := os.WriteFile(target, []byte(req.Grammar), 0644); err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "saved_to": target})
}

func (s *studioServer) handleAIStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OllamaURL   string `json:"ollama_url"`
		LMStudioURL string `json:"lmstudio_url"`
	}
	if r.Method == http.MethodPost {
		_ = decodeJSONBody(r, &req)
	} else {
		req.OllamaURL = r.URL.Query().Get("ollama_url")
		req.LMStudioURL = r.URL.Query().Get("lmstudio_url")
	}
	status := s.ai.ProbeAll(r.Context(), req.OllamaURL, req.LMStudioURL)
	writeJSON(w, http.StatusOK, status)
}

func (s *studioServer) handleAIRun(w http.ResponseWriter, r *http.Request) {
	var req AIRunRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, AIRunResponse{OK: false, Error: err.Error()})
		return
	}
	resp := s.ai.Run(r.Context(), s.engine.DSL(), req)
	if resp.OK && resp.ResultYAML != "" {
		_ = s.engine.LoadGrammar(resp.ResultYAML, "yaml")
		s.mu.Lock()
		s.lastSpec = resp.ResultYAML
		s.mu.Unlock()
	}
	writeJSON(w, http.StatusOK, resp)
}

func decodeJSONBody(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
