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

// server is a minimal stdio LSP server: framing, lifecycle, diagnostics on
// open/change (incrementally re-parsed via dslbuilder.Document), completion
// (parser expectations at the cursor), and hover (AST node under the cursor).
type server struct {
	dsl    *dslbuilder.DSL
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex // serializes writes to the client
	exited bool

	docs map[string]*dslbuilder.Document // open documents by URI
}

func newServer(dsl *dslbuilder.DSL, r io.Reader, w io.Writer) *server {
	return &server{
		dsl:    dsl,
		reader: bufio.NewReader(r),
		writer: w,
		docs:   make(map[string]*dslbuilder.Document),
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
				"textDocumentSync":   1,
				"completionProvider": map[string]interface{}{},
				"hoverProvider":      true,
			},
			"serverInfo": map[string]interface{}{
				"name":    "go-dsl-lsp",
				"version": "1.1",
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
		uri := params.TextDocument.URI
		doc, ok := s.docs[uri]
		if !ok {
			doc = s.dsl.NewDocument()
			s.docs[uri] = doc
		}
		// Incremental: only statements touched by the edit re-parse.
		doc.Update(text)
		s.publishDiagnostics(uri, doc)

	case "textDocument/didClose":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return
		}
		delete(s.docs, params.TextDocument.URI)
		s.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Diagnostics: []lspDiagnostic{},
		})

	case "textDocument/completion":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.reply(req.ID, nil)
			return
		}
		s.reply(req.ID, s.completionItems(params.TextDocument.URI, params.Position))

	case "textDocument/hover":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.reply(req.ID, nil)
			return
		}
		s.reply(req.ID, s.hover(params.TextDocument.URI, params.Position))

	default:
		// Requests we don't implement get an empty result so clients
		// don't hang; notifications are simply ignored.
		if req.ID != nil {
			s.reply(req.ID, nil)
		}
	}
}

// publishDiagnostics pushes the document's current diagnostics.
func (s *server) publishDiagnostics(uri string, doc *dslbuilder.Document) {
	diags := doc.Diagnostics()

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

// completionItems returns the parser's suggestions at a cursor position.
func (s *server) completionItems(uri string, pos lspPosition) []lspCompletionItem {
	doc, ok := s.docs[uri]
	if !ok {
		return []lspCompletionItem{}
	}
	offset := offsetForPosition(doc.Text(), pos)

	comps := s.dsl.Completions(doc.Text(), offset)
	items := make([]lspCompletionItem, 0, len(comps))
	for _, c := range comps {
		kind := 1 // Text (placeholder for free-form tokens)
		if c.IsKeyword {
			kind = 14 // Keyword
		}
		items = append(items, lspCompletionItem{
			Label:  c.Label,
			Kind:   kind,
			Detail: c.Detail,
		})
	}
	return items
}

// hover describes the AST node under the cursor.
func (s *server) hover(uri string, pos lspPosition) *lspHover {
	doc, ok := s.docs[uri]
	if !ok {
		return nil
	}
	offset := offsetForPosition(doc.Text(), pos)
	node := doc.NodeAt(offset)
	if node == nil {
		return nil
	}

	var md strings.Builder
	if node.IsToken() {
		fmt.Fprintf(&md, "**token** `%s`\n\n`%s`", node.Token.TokenType, node.Token.Value)
	} else {
		fmt.Fprintf(&md, "**rule** `%s`", node.Rule)
		if node.Action != "" {
			fmt.Fprintf(&md, " → action `%s`", node.Action)
		}
		text := doc.Text()
		if node.Span.Start >= 0 && node.Span.End <= len(text) {
			fmt.Fprintf(&md, "\n\n```\n%s\n```", text[node.Span.Start:node.Span.End])
		}
	}

	start := positionForOffset(doc.Text(), node.Span.Start)
	end := positionForOffset(doc.Text(), node.Span.End)
	return &lspHover{
		Contents: lspMarkup{Kind: "markdown", Value: md.String()},
		Range:    &lspRange{Start: start, End: end},
	}
}

// offsetForPosition converts an LSP position (0-based line/character) to a
// byte offset. Characters are counted as bytes, which is exact for ASCII
// DSLs and a close approximation otherwise.
func offsetForPosition(text string, pos lspPosition) int {
	offset := 0
	line := 0
	for line < pos.Line {
		idx := strings.IndexByte(text[offset:], '\n')
		if idx < 0 {
			return len(text)
		}
		offset += idx + 1
		line++
	}
	offset += pos.Character
	if offset > len(text) {
		offset = len(text)
	}
	return offset
}

// positionForOffset converts a byte offset to an LSP position.
func positionForOffset(text string, offset int) lspPosition {
	if offset > len(text) {
		offset = len(text)
	}
	line := strings.Count(text[:offset], "\n")
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	return lspPosition{Line: line, Character: offset - lineStart}
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
