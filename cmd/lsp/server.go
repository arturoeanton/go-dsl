package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// server is a minimal stdio LSP server. It is deliberately tiny: framing,
// four lifecycle methods, and diagnostics on open/change.
type server struct {
	dsl    *dslbuilder.DSL
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex // serializes writes to the client
	exited bool
}

func newServer(dsl *dslbuilder.DSL, r io.Reader, w io.Writer) *server {
	return &server{
		dsl:    dsl,
		reader: bufio.NewReader(r),
		writer: w,
	}
}

// run processes messages until exit or EOF.
func (s *server) run() error {
	for !s.exited {
		payload, err := s.readMessage()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		var req rpcRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			continue // ignore malformed frames
		}
		s.handle(&req)
	}
	return nil
}

// readMessage reads one Content-Length framed JSON-RPC payload.
func (s *server) readMessage() ([]byte, error) {
	contentLength := 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end of headers
		}
		if v, ok := strings.CutPrefix(line, "Content-Length:"); ok {
			n, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
			contentLength = n
		}
	}

	if contentLength <= 0 || contentLength > 16<<20 {
		return nil, fmt.Errorf("invalid Content-Length: %d", contentLength)
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(s.reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (s *server) handle(req *rpcRequest) {
	switch req.Method {
	case "initialize":
		s.reply(req.ID, map[string]interface{}{
			"capabilities": map[string]interface{}{
				// 1 = full document sync: the client sends the whole text.
				"textDocumentSync": 1,
			},
			"serverInfo": map[string]interface{}{
				"name":    "go-dsl-lsp",
				"version": "1.0",
			},
		})

	case "shutdown":
		s.reply(req.ID, nil)

	case "exit":
		s.exited = true

	case "textDocument/didOpen", "textDocument/didChange":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return
		}
		text := params.TextDocument.Text
		if len(params.ContentChanges) > 0 {
			// Full sync: the last change carries the complete text.
			text = params.ContentChanges[len(params.ContentChanges)-1].Text
		}
		s.publishDiagnostics(params.TextDocument.URI, text)

	case "textDocument/didClose":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return
		}
		s.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []lspDiagnostic{},
		})

	default:
		// Requests we don't implement get an empty result so clients
		// don't hang; notifications are simply ignored.
		if req.ID != nil {
			s.reply(req.ID, nil)
		}
	}
}

// publishDiagnostics validates the document with the multi-error engine
// and pushes the results to the client.
func (s *server) publishDiagnostics(uri, text string) {
	diags := s.dsl.Diagnostics(text)

	lspDiags := make([]lspDiagnostic, 0, len(diags))
	for _, d := range diags {
		// ParseError positions are 1-based; LSP is 0-based.
		line := d.Line - 1
		if line < 0 {
			line = 0
		}
		col := d.Column - 1
		if col < 0 {
			col = 0
		}
		endCol := col + len(d.Token)
		if endCol == col {
			endCol = col + 1
		}
		lspDiags = append(lspDiags, lspDiagnostic{
			Range: lspRange{
				Start: lspPosition{Line: line, Character: col},
				End:   lspPosition{Line: line, Character: endCol},
			},
			Severity: 1, // Error
			Source:   "go-dsl",
			Message:  d.Message,
		})
	}

	s.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

func (s *server) reply(id *json.RawMessage, result interface{}) {
	s.send(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *server) notify(method string, params interface{}) {
	s.send(rpcNotification{JSONRPC: "2.0", Method: method, Params: params})
}

func (s *server) send(v interface{}) {
	payload, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(s.writer, "Content-Length: %d\r\n\r\n%s", len(payload), payload)
}
