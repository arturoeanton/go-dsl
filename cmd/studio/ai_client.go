package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"gopkg.in/yaml.v3"
)

const (
	defaultOllamaURL   = "http://127.0.0.1:11434"
	defaultLMStudioURL = "http://127.0.0.1:1234"
)

// AIProviderInfo reports connection status and discovered models for a local AI runtime.
type AIProviderInfo struct {
	Provider       string    `json:"provider"` // "ollama" | "lmstudio"
	BaseURL        string    `json:"base_url"`
	Online         bool      `json:"online"`
	LatencyMs      int64     `json:"latency_ms"`
	Models         []AIModel `json:"models"`
	Recommended    string    `json:"recommended,omitempty"`
	Error          string    `json:"error,omitempty"`
	SupportsGBNF   bool      `json:"supports_gbnf"`
	SetupHint      string    `json:"setup_hint,omitempty"`
}

// AIModel describes one model exposed by Ollama or LM Studio.
type AIModel struct {
	ID            string `json:"id"`
	ParameterSize string `json:"parameter_size,omitempty"`
	Quantization  string `json:"quantization,omitempty"`
	Recommended   bool   `json:"recommended,omitempty"`
}

// AIStatusResponse aggregates live telemetry for both Ollama and LM Studio.
type AIStatusResponse struct {
	Ollama   AIProviderInfo `json:"ollama"`
	LMStudio AIProviderInfo `json:"lmstudio"`
}

// AIRunRequest represents a Studio AI Copilot request.
type AIRunRequest struct {
	Provider    string  `json:"provider"`     // "ollama" | "lmstudio"
	BaseURL     string  `json:"base_url"`     // Optional custom localhost URL
	Model       string  `json:"model"`        // Selected model ID
	Mode        string  `json:"mode"`         // "generate_code" | "repair_code" | "generate_grammar" | "generate_tests"
	Prompt      string  `json:"prompt"`       // Natural language instruction
	Code        string  `json:"code"`         // Current DSL code (for repair_code)
	GrammarYAML string  `json:"grammar_yaml"` // Current grammar YAML
	UseGBNF     bool    `json:"use_gbnf"`     // Request GBNF grammar constrained decoding
	Temperature float64 `json:"temperature"`
}

// AIRunResponse returns the verified artifact + full guardrail telemetry.
type AIRunResponse struct {
	OK           bool     `json:"ok"`
	Error        string   `json:"error,omitempty"`
	Mode         string   `json:"mode"`
	ResultCode   string   `json:"result_code,omitempty"`
	ResultYAML   string   `json:"result_yaml,omitempty"`
	SampleCode   string   `json:"sample_code,omitempty"`
	Verified     bool     `json:"verified"`
	Attempts     int      `json:"attempts"`
	GBNFUsed     bool     `json:"gbnf_used"`
	GBNFFallback bool     `json:"gbnf_fallback"`
	DurationMs   int64    `json:"duration_ms"`
	Log          []string `json:"log,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiBridge struct {
	allowRemote bool
	httpClient  *http.Client
}

func newAIBridge(allowRemote bool) *aiBridge {
	return &aiBridge{
		allowRemote: allowRemote,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// validateEndpoint ensures Studio only proxies to loopback addresses unless -allow-remote-ai is passed.
func (b *aiBridge) validateEndpoint(rawURL, fallback string) (string, error) {
	uStr := strings.TrimSpace(rawURL)
	if uStr == "" {
		uStr = fallback
	}
	uStr = strings.TrimRight(uStr, "/")
	parsed, err := url.Parse(uStr)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("invalid AI endpoint URL %q", uStr)
	}
	if !b.allowRemote {
		host := parsed.Hostname()
		if host != "localhost" {
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() {
				return "", fmt.Errorf("for security, Studio only connects to localhost/127.0.0.1 AI servers by default (pass -allow-remote-ai to allow %q)", host)
			}
		}
	}
	return uStr, nil
}

// ProbeAll checks Ollama and LM Studio concurrently with a short timeout.
func (b *aiBridge) ProbeAll(ctx context.Context, ollamaURL, lmStudioURL string) AIStatusResponse {
	probeCtx, cancel := context.WithTimeout(ctx, 1800*time.Millisecond)
	defer cancel()

	var (
		res AIStatusResponse
		wg  sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		res.Ollama = b.probeOllama(probeCtx, ollamaURL)
	}()
	go func() {
		defer wg.Done()
		res.LMStudio = b.probeLMStudio(probeCtx, lmStudioURL)
	}()
	wg.Wait()
	return res
}

func (b *aiBridge) probeOllama(ctx context.Context, rawURL string) AIProviderInfo {
	base, err := b.validateEndpoint(rawURL, defaultOllamaURL)
	info := AIProviderInfo{
		Provider:     "ollama",
		BaseURL:      base,
		SupportsGBNF: true,
		SetupHint:    "Ejecuta `ollama serve` y descarga un modelo como `ollama pull qwen2.5-coder:7b`",
	}
	if err != nil {
		info.Error = err.Error()
		return info
	}

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/tags", nil)
	resp, err := b.httpClient.Do(req)
	if err != nil {
		info.Error = "No detectado en " + base
		return info
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		info.Error = fmt.Sprintf("HTTP %d desde Ollama", resp.StatusCode)
		return info
	}

	var payload struct {
		Models []struct {
			Name    string `json:"name"`
			Details struct {
				ParameterSize     string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err := json.Unmarshal(body, &payload); err != nil {
		info.Error = "Respuesta inválida de /api/tags"
		return info
	}

	info.Online = true
	info.LatencyMs = time.Since(start).Milliseconds()
	for _, m := range payload.Models {
		rec := isRecommendedCoderModel(m.Name)
		info.Models = append(info.Models, AIModel{
			ID:            m.Name,
			ParameterSize: m.Details.ParameterSize,
			Quantization:  m.Details.QuantizationLevel,
			Recommended:   rec,
		})
		if rec && info.Recommended == "" {
			info.Recommended = m.Name
		}
	}
	if info.Recommended == "" && len(info.Models) > 0 {
		info.Recommended = info.Models[0].ID
	}
	return info
}

func (b *aiBridge) probeLMStudio(ctx context.Context, rawURL string) AIProviderInfo {
	base, err := b.validateEndpoint(rawURL, defaultLMStudioURL)
	info := AIProviderInfo{
		Provider:     "lmstudio",
		BaseURL:      base,
		SupportsGBNF: true,
		SetupHint:    "Abre LM Studio → Developer → Start Server (puerto 1234) y carga un modelo GGUF/MLX",
	}
	if err != nil {
		info.Error = err.Error()
		return info
	}

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/models", nil)
	resp, err := b.httpClient.Do(req)
	if err != nil {
		info.Error = "No detectado en " + base
		return info
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		info.Error = fmt.Sprintf("HTTP %d desde LM Studio", resp.StatusCode)
		return info
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err := json.Unmarshal(body, &payload); err != nil {
		info.Error = "Respuesta inválida de /v1/models"
		return info
	}

	info.Online = true
	info.LatencyMs = time.Since(start).Milliseconds()
	for _, m := range payload.Data {
		if strings.Contains(strings.ToLower(m.ID), "embed") {
			continue
		}
		rec := isRecommendedCoderModel(m.ID)
		info.Models = append(info.Models, AIModel{
			ID:          m.ID,
			Recommended: rec,
		})
		if rec && info.Recommended == "" {
			info.Recommended = m.ID
		}
	}
	if info.Recommended == "" && len(info.Models) > 0 {
		info.Recommended = info.Models[0].ID
	}
	return info
}

func isRecommendedCoderModel(name string) bool {
	lower := strings.ToLower(name)
	for _, kw := range []string{"qwen2.5-coder", "qwen3-coder", "qwen-coder", "deepseek-coder", "codestral", "llama3.1", "llama-3.1", "llama3.3", "llama-3.3"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// chat invokes either Ollama or LM Studio, applying GBNF grammar constraints when requested
// and automatically falling back if the loaded model backend (e.g. MLX) rejects raw GBNF.
func (b *aiBridge) chat(ctx context.Context, provider, rawURL, model string, msgs []chatMessage, gbnf string, temp float64) (reply string, gbnfUsed bool, gbnfFallback bool, err error) {
	if strings.TrimSpace(model) == "" {
		return "", false, false, fmt.Errorf("selecciona un modelo activo en %s", provider)
	}
	if temp <= 0 {
		temp = 0.15
	}

	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "ollama":
		base, err := b.validateEndpoint(rawURL, defaultOllamaURL)
		if err != nil {
			return "", false, false, err
		}
		return b.chatOpenAICompatOrOllama(ctx, base, model, msgs, gbnf, temp, true)

	case "lmstudio":
		base, err := b.validateEndpoint(rawURL, defaultLMStudioURL)
		if err != nil {
			return "", false, false, err
		}
		return b.chatOpenAICompatOrOllama(ctx, base, model, msgs, gbnf, temp, false)

	default:
		return "", false, false, fmt.Errorf("proveedor IA desconocido %q (usa 'ollama' o 'lmstudio')", provider)
	}
}

func (b *aiBridge) chatOpenAICompatOrOllama(ctx context.Context, base, model string, msgs []chatMessage, gbnf string, temp float64, isOllama bool) (string, bool, bool, error) {
	// Try OpenAI-compatible /v1/chat/completions first when GBNF is requested (supported by llama.cpp / LM Studio / modern Ollama v1),
	// with automatic fallback to unconstrained /api/chat (for Ollama) or unconstrained /v1/chat/completions (for LM Studio MLX).
	if gbnf != "" {
		reply, status, err := b.postOpenAIChat(ctx, base+"/v1/chat/completions", model, msgs, gbnf, temp)
		if err == nil && strings.TrimSpace(reply) != "" {
			return reply, true, false, nil
		}
		// Graceful fallback if the backend does not accept raw GBNF (e.g. HTTP 400/404/422 on MLX or older Ollama).
		if status >= 400 || err != nil {
			if isOllama {
				rep2, err2 := b.postOllamaNativeChat(ctx, base+"/api/chat", model, msgs, temp)
				return rep2, false, true, err2
			}
			rep2, _, err2 := b.postOpenAIChat(ctx, base+"/v1/chat/completions", model, msgs, "", temp)
			return rep2, false, true, err2
		}
	}

	if isOllama {
		rep, err := b.postOllamaNativeChat(ctx, base+"/api/chat", model, msgs, temp)
		return rep, false, false, err
	}
	rep, _, err := b.postOpenAIChat(ctx, base+"/v1/chat/completions", model, msgs, "", temp)
	return rep, false, false, err
}

func (b *aiBridge) postOpenAIChat(ctx context.Context, endpoint, model string, msgs []chatMessage, gbnf string, temp float64) (string, int, error) {
	bodyMap := map[string]interface{}{
		"model":       model,
		"messages":    msgs,
		"temperature": temp,
		"stream":      false,
	}
	if gbnf != "" {
		bodyMap["grammar"] = gbnf
	}
	payload, _ := json.Marshal(bodyMap)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode >= 400 {
		return "", resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", resp.StatusCode, err
	}
	if len(parsed.Choices) == 0 {
		return "", resp.StatusCode, fmt.Errorf("respuesta vacía del modelo")
	}
	return stripThinkTags(parsed.Choices[0].Message.Content), resp.StatusCode, nil
}

func (b *aiBridge) postOllamaNativeChat(ctx context.Context, endpoint, model string, msgs []chatMessage, temp float64) (string, error) {
	bodyMap := map[string]interface{}{
		"model":    model,
		"messages": msgs,
		"stream":   false,
		"options": map[string]interface{}{
			"temperature": temp,
		},
	}
	payload, _ := json.Marshal(bodyMap)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("Ollama HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}

	var parsed struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", err
	}
	return stripThinkTags(parsed.Message.Content), nil
}

// Run executes one of the 4 closed-loop guardrailed Studio AI workflows.
func (b *aiBridge) Run(ctx context.Context, dsl *dslbuilder.DSL, req AIRunRequest) AIRunResponse {
	start := time.Now()
	out := AIRunResponse{Mode: req.Mode}

	switch req.Mode {
	case "generate_code":
		if dsl == nil {
			out.Error = "Carga primero una gramática válida antes de generar código DSL."
			return out
		}
		return b.runGenerateOrRepairCode(ctx, dsl, req, false, start)

	case "repair_code":
		if dsl == nil {
			out.Error = "Carga primero una gramática válida antes de reparar código DSL."
			return out
		}
		return b.runGenerateOrRepairCode(ctx, dsl, req, true, start)

	case "generate_grammar":
		return b.runGenerateGrammar(ctx, req, start)

	case "generate_tests":
		if dsl == nil {
			out.Error = "Carga primero una gramática válida antes de generar tests."
			return out
		}
		return b.runGenerateTests(ctx, dsl, req, start)

	default:
		out.Error = fmt.Sprintf("modo IA desconocido: %q", req.Mode)
		return out
	}
}

func (b *aiBridge) runGenerateOrRepairCode(ctx context.Context, dsl *dslbuilder.DSL, req AIRunRequest, isRepair bool, start time.Time) AIRunResponse {
	out := AIRunResponse{Mode: req.Mode}
	sysPrompt := "You are an expert DSL engineer. Output ONLY valid DSL code conforming strictly to the grammar specification below. Do NOT wrap in explanations.\n\n" + dsl.AIPromptSpec()

	var userMsg string
	if isRepair {
		repairPrompt, clean := dsl.RepairPrompt(req.Code)
		if clean {
			formatted, err := dsl.Format(req.Code)
			if err != nil {
				formatted = req.Code
			}
			out.OK = true
			out.Verified = true
			out.Attempts = 0
			out.ResultCode = formatted
			out.DurationMs = time.Since(start).Milliseconds()
			out.Log = []string{"El código ya es 100% válido según el parser determinista."}
			return out
		}
		userMsg = repairPrompt
		if strings.TrimSpace(req.Prompt) != "" {
			userMsg += "\nAdditional user context: " + req.Prompt
		}
	} else {
		userMsg = fmt.Sprintf("Write valid `%s` DSL code for the following request:\n%s\n\nOutput ONLY the raw DSL code.", dsl.ToConfig().Name, req.Prompt)
	}

	msgs := []chatMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userMsg},
	}

	gbnf := ""
	if req.UseGBNF {
		gbnf = dsl.ToGBNF()
	}

	const maxAttempts = 3
	var lastCandidate string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		out.Attempts = attempt
		reply, usedGBNF, fellBack, err := b.chat(ctx, req.Provider, req.BaseURL, req.Model, msgs, gbnf, req.Temperature)
		if err != nil {
			out.Error = err.Error()
			out.DurationMs = time.Since(start).Milliseconds()
			return out
		}
		out.GBNFUsed = out.GBNFUsed || usedGBNF
		out.GBNFFallback = out.GBNFFallback || fellBack

		candidate := extractCodeBlock(reply)
		lastCandidate = candidate

		repairPrompt, clean := dsl.RepairPrompt(candidate)
		if clean {
			if _, parseErr := dsl.ParseAST(candidate); parseErr == nil {
				if formatted, fmtErr := dsl.Format(candidate); fmtErr == nil {
					candidate = formatted
				}
				out.OK = true
				out.Verified = true
				out.ResultCode = candidate
				out.DurationMs = time.Since(start).Milliseconds()
				out.Log = append(out.Log, fmt.Sprintf("Intento %d: Verificado sin errores por go-dsl ParseAST + Format.", attempt))
				return out
			}
		}

		out.Log = append(out.Log, fmt.Sprintf("Intento %d: Detectado error sintáctico; aplicando bucle cerrado RepairPrompt.", attempt))
		msgs = append(msgs,
			chatMessage{Role: "assistant", Content: candidate},
			chatMessage{Role: "user", Content: repairPrompt},
		)
	}

	out.OK = true
	out.Verified = false
	out.ResultCode = lastCandidate
	out.DurationMs = time.Since(start).Milliseconds()
	out.Log = append(out.Log, "Se alcanzó el máximo de 3 intentos; revisa los diagnósticos en el editor.")
	return out
}

func (b *aiBridge) runGenerateGrammar(ctx context.Context, req AIRunRequest, start time.Time) AIRunResponse {
	out := AIRunResponse{Mode: req.Mode}
	sysPrompt := `You are a master Programming Language Designer specialized in go-dsl v2 YAML grammars.
Output ONLY raw valid YAML (no markdown prose) matching this exact go-dsl v2 schema:

name: MyDSL
start_rule: stmt
token_defs:
  - name: COMMENT
    kind: skip
    pattern: "//[^\\r\\n]*"
  - name: WHEN
    kind: keyword
    pattern: when
  - name: EQ
    kind: literal
    pattern: "=="
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+(\\.[0-9]+)?"
  - name: IDENT
    kind: regex
    pattern: "[a-zA-Z_][a-zA-Z0-9_]*"
rules:
  - name: stmt
    pattern: [WHEN, expr]
    action: on_when
expressions:
  - name: expr
    atoms:
      - token: NUMBER
        action: number
      - token: IDENT
        action: ident
    infix:
      - token: EQ
        power: 10
        assoc: nonassoc
        action: eq
tests:
  - name: valid sample
    input: "when score == 100"
    should_parse: true

CRITICAL RULES:
1. Declare COMMENT (kind: skip) first, then keywords (kind: keyword), then operators/punctuation (kind: literal), then regex tokens (NUMBER, STRING, IDENT) last!
2. Every symbol referenced in rules[].pattern or expressions[] MUST exist in token_defs, rules, or expressions.
3. No token regex may match the empty string (use + instead of * at top level).
4. Include 2 valid tests in tests: whose input strictly matches start_rule!`

	msgs := []chatMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: "Design a complete go-dsl v2 YAML grammar for:\n" + req.Prompt},
	}

	var lastYAML string
	for attempt := 1; attempt <= 3; attempt++ {
		out.Attempts = attempt
		reply, _, _, err := b.chat(ctx, req.Provider, req.BaseURL, req.Model, msgs, "", 0.2)
		if err != nil {
			out.Error = err.Error()
			out.DurationMs = time.Since(start).Milliseconds()
			return out
		}

		candidateYAML := extractCodeBlock(reply)
		lastYAML = candidateYAML

		d, loadErr := dslbuilder.LoadFromYAML([]byte(candidateYAML))
		if loadErr != nil {
			out.Log = append(out.Log, fmt.Sprintf("Intento %d: Error cargando YAML (%v); auto-reparando.", attempt, loadErr))
			msgs = append(msgs,
				chatMessage{Role: "assistant", Content: candidateYAML},
				chatMessage{Role: "user", Content: fmt.Sprintf("LoadFromYAML failed with error:\n%v\n\nReturn ONLY the fixed YAML.", loadErr)},
			)
			continue
		}

		_, valErr := d.Validate()
		if valErr != nil {
			out.Log = append(out.Log, fmt.Sprintf("Intento %d: Error estructural Validate() (%v); auto-reparando.", attempt, valErr))
			msgs = append(msgs,
				chatMessage{Role: "assistant", Content: candidateYAML},
				chatMessage{Role: "user", Content: fmt.Sprintf("Grammar Validate() failed with structural errors:\n%v\n\nReturn ONLY the fixed YAML.", valErr)},
			)
			continue
		}

		sample := ""
		if tests := d.Tests(); len(tests) > 0 {
			for _, tc := range tests {
				if (tc.ShouldParse == nil || *tc.ShouldParse) && tc.ExpectError == "" {
					if _, err := d.ParseAST(tc.Input); err == nil {
						sample = tc.Input
						break
					}
				}
			}
		}
		if sample == "" {
			if samples := d.GenerateSamples(42, 1, 4); len(samples) > 0 {
				sample = samples[0]
			}
		}

		out.OK = true
		out.Verified = true
		out.ResultYAML = candidateYAML
		out.SampleCode = sample
		out.DurationMs = time.Since(start).Milliseconds()
		out.Log = append(out.Log, fmt.Sprintf("Intento %d: Gramática v2 validada sin errores estructurales por go-dsl.", attempt))
		return out
	}

	out.OK = true
	out.Verified = false
	out.ResultYAML = lastYAML
	out.DurationMs = time.Since(start).Milliseconds()
	return out
}

func (b *aiBridge) runGenerateTests(ctx context.Context, dsl *dslbuilder.DSL, req AIRunRequest, start time.Time) AIRunResponse {
	out := AIRunResponse{Mode: req.Mode}
	sysPrompt := "You generate declarative YAML test cases for a formal go-dsl grammar.\nOutput ONLY a YAML list of test objects with fields: name, input, should_parse (true or false).\n\n" + dsl.AIPromptSpec()
	userPrompt := "Generate 4 realistic domain test cases (3 valid with should_parse: true, and 1 invalid syntax edge case with should_parse: false)."
	if strings.TrimSpace(req.Prompt) != "" {
		userPrompt += "\nDomain focus: " + req.Prompt
	}

	msgs := []chatMessage{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: userPrompt},
	}

	out.Attempts = 1
	reply, _, _, err := b.chat(ctx, req.Provider, req.BaseURL, req.Model, msgs, "", 0.25)
	if err != nil {
		out.Error = err.Error()
		out.DurationMs = time.Since(start).Milliseconds()
		return out
	}

	rawBlock := extractCodeBlock(reply)
	var cases []dslbuilder.DSLTestCase
	// Support either raw list or `{tests: [...]}` wrapper.
	if err := yaml.Unmarshal([]byte(rawBlock), &cases); err != nil || len(cases) == 0 {
		var wrapper struct {
			Tests []dslbuilder.DSLTestCase `yaml:"tests"`
		}
		_ = yaml.Unmarshal([]byte(rawBlock), &wrapper)
		cases = wrapper.Tests
	}

	var verified []dslbuilder.DSLTestCase
	for _, tc := range cases {
		if strings.TrimSpace(tc.Input) == "" {
			continue
		}
		_, parseErr := dsl.ParseAST(tc.Input)
		actualValid := parseErr == nil
		tcCopy := tc
		tcCopy.ShouldParse = &actualValid
		verified = append(verified, tcCopy)
	}

	if len(verified) == 0 {
		// Deterministic fallback synthesis if LLM returned malformed YAML
		for i, s := range dsl.GenerateSamples(time.Now().UnixNano(), 3, 4) {
			tVal := true
			verified = append(verified, dslbuilder.DSLTestCase{
				Name:        fmt.Sprintf("synthesized_valid_%d", i+1),
				Input:       s,
				ShouldParse: &tVal,
			})
		}
	}

	cfg := dsl.ToConfig()
	if strings.TrimSpace(req.GrammarYAML) != "" {
		var userCfg dslbuilder.DSLConfig
		if yaml.Unmarshal([]byte(req.GrammarYAML), &userCfg) == nil && userCfg.Name != "" {
			cfg = userCfg
		}
	}
	cfg.Tests = append(cfg.Tests, verified...)
	updatedYAML, _ := yaml.Marshal(cfg)

	out.OK = true
	out.Verified = true
	out.ResultYAML = string(updatedYAML)
	out.DurationMs = time.Since(start).Milliseconds()
	out.Log = []string{fmt.Sprintf("Generados y verificados determinísticamente %d casos de prueba contra ParseAST.", len(verified))}
	return out
}

func stripThinkTags(s string) string {
	for {
		start := strings.Index(s, "<think>")
		end := strings.Index(s, "</think>")
		if start >= 0 && end > start {
			s = s[:start] + s[end+len("</think>"):]
		} else {
			break
		}
	}
	return strings.TrimSpace(s)
}

func extractCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "```") {
		return s
	}
	first := strings.Index(s, "```")
	rest := s[first+3:]
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[nl+1:]
	}
	if end := strings.Index(rest, "```"); end >= 0 {
		return strings.TrimSpace(rest[:end])
	}
	return strings.TrimSpace(rest)
}
