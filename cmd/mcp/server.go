package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type mcpRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *mcpError        `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type mcpToolResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpServer struct {
	dsl    *dslbuilder.DSL
	reader *bufio.Scanner
	writer io.Writer
	mu     sync.Mutex
}

func newMCPServer(dsl *dslbuilder.DSL, r io.Reader, w io.Writer) *mcpServer {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16<<20)
	return &mcpServer{
		dsl:    dsl,
		reader: sc,
		writer: w,
	}
}

func (s *mcpServer) run() error {
	for s.reader.Scan() {
		line := strings.TrimSpace(s.reader.Text())
		if line == "" {
			continue
		}
		var req mcpRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}
		s.handle(&req)
	}
	return s.reader.Err()
}

func (s *mcpServer) handle(req *mcpRequest) {
	// Notifications (no ID) require no response.
	if req.ID == nil {
		return
	}

	switch req.Method {
	case "initialize":
		s.reply(req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "go-dsl-mcp",
				"version": "1.2.0",
			},
		})

	case "ping":
		s.reply(req.ID, map[string]interface{}{})

	case "tools/list":
		s.reply(req.ID, map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "dsl_grammar_info",
					"description": "Returns the complete LLM prompt specification, rules, tokens, valid examples, and GBNF grammar for the loaded DSL.",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
				{
					"name":        "dsl_validate_code",
					"description": "Validates DSL source code against the formal grammar and returns structured diagnostics plus a self-repair prompt if invalid.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type":        "string",
								"description": "DSL source code to validate",
							},
						},
						"required": []string{"code"},
					},
				},
				{
					"name":        "dsl_complete",
					"description": "Returns valid next tokens/keywords expected by the parser at a byte offset in the DSL code.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type":        "string",
								"description": "DSL source code",
							},
							"offset": map[string]interface{}{
								"type":        "integer",
								"description": "Byte offset of cursor (defaults to end of code)",
							},
						},
						"required": []string{"code"},
					},
				},
				{
					"name":        "dsl_format",
					"description": "Formats DSL source code canonically while preserving comments.",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type":        "string",
								"description": "DSL source code to format",
							},
						},
						"required": []string{"code"},
					},
				},
				{
					"name":        "dsl_gbnf",
					"description": "Exports the formal GBNF grammar for constrained decoding in Ollama, LM Studio, or llama.cpp.",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		})

	case "tools/call":
		var params mcpToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.replyError(req.ID, -32602, "invalid tool call parameters")
			return
		}
		s.callTool(req.ID, params)

	default:
		s.replyError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *mcpServer) callTool(id *json.RawMessage, p mcpToolCallParams) {
	strArg := func(k string) string {
		if v, ok := p.Arguments[k].(string); ok {
			return v
		}
		return ""
	}

	switch p.Name {
	case "dsl_grammar_info":
		s.reply(id, mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: s.dsl.AIPromptSpec()}},
		})

	case "dsl_gbnf":
		s.reply(id, mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: s.dsl.ToGBNF()}},
		})

	case "dsl_validate_code":
		code := strArg("code")
		repair, clean := s.dsl.RepairPrompt(code)
		if clean {
			ast, err := s.dsl.ParseAST(code)
			if err == nil {
				s.reply(id, mcpToolResult{
					Content: []mcpContent{{Type: "text", Text: "VALID DSL CODE\n\nAST:\n" + ast.Pretty()}},
				})
				return
			}
			repair = dslbuilder.GetDetailedError(err)
		}
		s.reply(id, mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: repair}},
			IsError: true,
		})

	case "dsl_complete":
		code := strArg("code")
		offset := len(code)
		if rawOff, ok := p.Arguments["offset"].(float64); ok {
			offset = int(rawOff)
		}
		comps := s.dsl.Completions(code, offset)
		data, _ := json.MarshalIndent(comps, "", "  ")
		s.reply(id, mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: string(data)}},
		})

	case "dsl_format":
		code := strArg("code")
		formatted, err := s.dsl.Format(code)
		if err != nil {
			s.reply(id, mcpToolResult{
				Content: []mcpContent{{Type: "text", Text: dslbuilder.GetDetailedError(err)}},
				IsError: true,
			})
			return
		}
		s.reply(id, mcpToolResult{
			Content: []mcpContent{{Type: "text", Text: formatted}},
		})

	default:
		s.replyError(id, -32602, fmt.Sprintf("unknown tool: %s", p.Name))
	}
}

func (s *mcpServer) reply(id *json.RawMessage, result interface{}) {
	s.send(mcpResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *mcpServer) replyError(id *json.RawMessage, code int, msg string) {
	s.send(mcpResponse{JSONRPC: "2.0", ID: id, Error: &mcpError{Code: code, Message: msg}})
}

func (s *mcpServer) send(v interface{}) {
	payload, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(s.writer, "%s\n", payload)
}
