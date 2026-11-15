package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

var semanticTokenLegend = []string{
	"keyword",  // 0
	"number",   // 1
	"string",   // 2
	"operator", // 3
	"comment",  // 4
	"variable", // 5
}

// server is a stdio LSP server: framing, lifecycle, diagnostics on open/change
// (incrementally re-parsed via dslbuilder.Document), completion, hover,
// formatting, and semantic tokens.
type server struct {
	dsl    *dslbuilder.DSL
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex // serializes writes to the client
	exited bool

	docs       map[string]*dslbuilder.Document // open documents by URI
	tokenTypes map[string]int                  // token name -> semanticTokenLegend index
}

func newServer(dsl *dslbuilder.DSL, r io.Reader, w io.Writer) *server {
	return &server{
		dsl:        dsl,
		reader:     bufio.NewReader(r),
		writer:     w,
		docs:       make(map[string]*dslbuilder.Document),
		tokenTypes: classifyGrammarTokens(dsl),
	}
}

func classifyGrammarTokens(dsl *dslbuilder.DSL) map[string]int {
	out := make(map[string]int)
	if dsl == nil {
		return out
	}
	cfg := dsl.ToConfig()
	for _, td := range cfg.TokenDefs {
		kind := strings.ToLower(td.Kind)
		nameUpper := strings.ToUpper(td.Name)
		pat := td.Pattern
		switch {
		case kind == "skip":
			out[td.Name] = 4 // comment / trivia
		case kind == "keyword":
			out[td.Name] = 0 // keyword
		case kind == "literal":
			if isAlphaWord(pat) {
				out[td.Name] = 0 // keyword
			} else {
				out[td.Name] = 3 // operator / punctuation
			}
		case strings.Contains(nameUpper, "NUM") || strings.Contains(nameUpper, "INT") || strings.Contains(nameUpper, "FLOAT") || strings.HasPrefix(pat, "[0-9]"):
			out[td.Name] = 1 // number
		case strings.Contains(nameUpper, "STR") || strings.HasPrefix(pat, `"`) || strings.HasPrefix(pat, `'`):
			out[td.Name] = 2 // string
		case strings.Contains(pat, `(?i)\b`):
			out[td.Name] = 0 // keyword
		default:
			out[td.Name] = 5 // variable / identifier
		}
	}
	return out
}

func isAlphaWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
			return false
		}
	}
	return true
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
				"textDocumentSync":           1,
				"completionProvider":         map[string]interface{}{},
				"hoverProvider":              true,
				"documentFormattingProvider": true,
				"semanticTokensProvider": map[string]interface{}{
					"legend": map[string]interface{}{
						"tokenTypes":     semanticTokenLegend,
						"tokenModifiers": []string{},
					},
					"full": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "go-dsl-lsp",
				"version": "1.2",
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

	case "textDocument/formatting":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.reply(req.ID, []lspTextEdit{})
			return
		}
		s.reply(req.ID, s.formatDocument(params.TextDocument.URI))

	case "textDocument/semanticTokens/full":
		var params textDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.reply(req.ID, lspSemanticTokens{Data: []int{}})
			return
		}
		s.reply(req.ID, s.semanticTokens(params.TextDocument.URI))

	default:
		// Requests we don't implement get an empty result so clients
		// don't hang; notifications are simply ignored.
		if req.ID != nil {
			s.reply(req.ID, nil)
		}
	}
}

// publishDiagnostics pushes the document's current diagnostics with UTF-16 character columns.
func (s *server) publishDiagnostics(uri string, doc *dslbuilder.Document) {
	diags := doc.Diagnostics()
	text := doc.Text()
	lines := strings.Split(text, "\n")

	lspDiags := make([]lspDiagnostic, 0, len(diags))
	for _, d := range diags {
		// ParseError positions are 1-based; d.Column is 1-based rune index on that line.
		line := d.Line - 1
		if line < 0 {
			line = 0
		}
		runeIdx := d.Column - 1
		if runeIdx < 0 {
			runeIdx = 0
		}
		col := runeIdx
		if line < len(lines) {
			col = utf16ColForRuneIndex(lines[line], runeIdx)
		}
		tokLen := 1
		if d.Token != "" && d.Token != "<end of input>" {
			tokLen = utf16StrLen(d.Token)
			if tokLen <= 0 {
				tokLen = 1
			}
		}
		endCol := col + tokLen
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
		item := lspCompletionItem{
			Label:  c.Label,
			Detail: c.Detail,
		}
		if c.IsKeyword {
			item.Kind = 14 // Keyword
			item.InsertText = c.Label
			item.InsertTextFormat = 1 // plain
		} else {
			item.Kind = 1 // Text
			item.InsertText = "${1:" + c.Label + "}"
			item.InsertTextFormat = 2 // snippet
		}
		items = append(items, item)
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

// formatDocument formats the open document using DSL.Format.
func (s *server) formatDocument(uri string) []lspTextEdit {
	doc, ok := s.docs[uri]
	if !ok {
		return []lspTextEdit{}
	}
	orig := doc.Text()
	formatted, err := s.dsl.Format(orig)
	if err != nil || formatted == orig {
		return []lspTextEdit{}
	}
	return []lspTextEdit{
		{
			Range: lspRange{
				Start: lspPosition{Line: 0, Character: 0},
				End:   positionForOffset(orig, len(orig)),
			},
			NewText: formatted,
		},
	}
}

// semanticTokens returns LSP delta-encoded semantic tokens for the document.
func (s *server) semanticTokens(uri string) lspSemanticTokens {
	doc, ok := s.docs[uri]
	if !ok {
		return lspSemanticTokens{Data: []int{}}
	}
	text := doc.Text()
	toks, trivia, err := s.dsl.TokenizeWithTrivia(text)
	if err != nil {
		// Fall back to line-by-line tokenization so healthy lines keep highlighting.
		toks, trivia = s.tokenizeLinesBestEffort(text)
	}

	type rawSemTok struct {
		start   int
		val     string
		tokType int
	}
	var all []rawSemTok
	for _, t := range toks {
		if t.TokenType == "NEWLINE" || t.Value == "" {
			continue
		}
		tt, exists := s.tokenTypes[t.TokenType]
		if !exists {
			tt = 5
		}
		all = append(all, rawSemTok{start: t.Start, val: t.Value, tokType: tt})
	}
	for _, tr := range trivia {
		if tr.Value == "" {
			continue
		}
		// Only highlight visible trivia (comments), not plain whitespace skips.
		if strings.TrimSpace(tr.Value) == "" {
			continue
		}
		all = append(all, rawSemTok{start: tr.Start, val: tr.Value, tokType: 4})
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].start < all[j].start
	})

	data := make([]int, 0, len(all)*5)
	prevLine := 0
	prevChar := 0
	for _, item := range all {
		// Multi-line tokens (e.g. block comments) are split per line as required by LSP.
		parts := strings.Split(item.val, "\n")
		curOffset := item.start
		for idx, part := range parts {
			partClean := strings.TrimSuffix(part, "\r")
			if partClean != "" {
				pos := positionForOffset(text, curOffset)
				deltaLine := pos.Line - prevLine
				deltaStart := pos.Character
				if deltaLine == 0 {
					deltaStart = pos.Character - prevChar
				}
				length := utf16StrLen(partClean)
				if length > 0 && deltaLine >= 0 && deltaStart >= 0 {
					data = append(data, deltaLine, deltaStart, length, item.tokType, 0)
					prevLine = pos.Line
					prevChar = pos.Character
				}
			}
			curOffset += len(part)
			if idx < len(parts)-1 {
				curOffset++ // '\n'
			}
		}
	}
	return lspSemanticTokens{Data: data}
}

func (s *server) tokenizeLinesBestEffort(text string) ([]dslbuilder.TokenMatch, []dslbuilder.TokenMatch) {
	var allToks, allTrivia []dslbuilder.TokenMatch
	offset := 0
	for _, line := range strings.Split(text, "\n") {
		toks, triv, err := s.dsl.TokenizeWithTrivia(line)
		if err == nil {
			for _, t := range toks {
				t.Start += offset
				t.End += offset
				allToks = append(allToks, t)
			}
			for _, tr := range triv {
				tr.Start += offset
				tr.End += offset
				allTrivia = append(allTrivia, tr)
			}
		}
		offset += len(line) + 1
	}
	return allToks, allTrivia
}

func utf16RuneLen(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

func utf16StrLen(s string) int {
	n := 0
	for _, r := range s {
		n += utf16RuneLen(r)
	}
	return n
}

func utf16ColForRuneIndex(line string, runeIdx int) int {
	line = strings.TrimSuffix(line, "\r")
	units := 0
	count := 0
	for _, r := range line {
		if count >= runeIdx {
			break
		}
		units += utf16RuneLen(r)
		count++
	}
	return units
}

// offsetForPosition converts an LSP position (0-based line + UTF-16 character)
// to a byte offset in text, strictly clamped to the target line so out-of-bounds
// character offsets never bleed into subsequent lines.
func offsetForPosition(text string, pos lspPosition) int {
	if pos.Line < 0 {
		return 0
	}
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

	lineEnd := len(text)
	if nl := strings.IndexByte(text[offset:], '\n'); nl >= 0 {
		lineEnd = offset + nl
	}
	if lineEnd > offset && text[lineEnd-1] == '\r' {
		lineEnd--
	}

	if pos.Character <= 0 {
		return offset
	}

	utf16Units := 0
	cur := offset
	for cur < lineEnd && utf16Units < pos.Character {
		r, size := utf8.DecodeRuneInString(text[cur:lineEnd])
		if size == 0 {
			break
		}
		nextUnits := utf16Units + utf16RuneLen(r)
		if nextUnits > pos.Character {
			break
		}
		utf16Units = nextUnits
		cur += size
	}
	return cur
}

// positionForOffset converts a byte offset in text to an LSP position (0-based
// line + UTF-16 character offset).
func positionForOffset(text string, offset int) lspPosition {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	line := strings.Count(text[:offset], "\n")
	lineStart := strings.LastIndexByte(text[:offset], '\n') + 1
	return lspPosition{
		Line:      line,
		Character: utf16StrLen(text[lineStart:offset]),
	}
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
