// Package dslbuilder - Incremental documents.
//
// Document keeps the parse state of one text buffer across edits and reuses
// subtrees the edit did not touch. The granularity is the TOP-LEVEL PARSE
// UNIT — one match of the grammar's start rule (typically "a statement"):
// units entirely inside the unchanged prefix are reused as-is, units
// entirely inside the unchanged suffix are reused with their spans shifted,
// and the edited unit(s) are re-parsed IN FULL. This is deliberate,
// bounded incrementality — not arbitrary sub-tree magic: a grammar whose
// whole document is a single start-rule match re-parses that unit on every
// edit. Lexing is always a full linear pass (it is cheap); the savings are
// in parsing and tree construction.
//
// This is the incremental engine behind cmd/lsp: hover needs a tree for the
// whole document (NodeAt) and diagnostics need all errors, on every
// keystroke.
package dslbuilder

import "sort"

// docStatement is one parsed statement of a Document.
type docStatement struct {
	node *Node // parse tree (nil when the statement failed to parse)
	err  *ParseError
	span Span // byte range of the statement (or of the error region)
}

// DocumentStats reports what the last Update did — useful to verify that
// incrementality is actually working.
type DocumentStats struct {
	Reused   int // statements reused from the previous parse
	Reparsed int // statements parsed from scratch
}

// Document is an incrementally re-parsed text buffer bound to a DSL.
// It is not safe for concurrent use; guard it externally (cmd/lsp
// serializes on its own mutex).
type Document struct {
	dsl   *DSL
	text  string
	stmts []docStatement
	stats DocumentStats
}

// NewDocument creates an empty incremental document for this DSL.
func (d *DSL) NewDocument() *Document {
	return &Document{dsl: d}
}

// Text returns the current document text.
func (doc *Document) Text() string { return doc.text }

// Stats returns the reuse statistics of the last Update call.
func (doc *Document) Stats() DocumentStats { return doc.stats }

// Diagnostics returns the errors of the current text (computed by Update).
func (doc *Document) Diagnostics() []*ParseError {
	var diags []*ParseError
	for _, s := range doc.stmts {
		if s.err != nil {
			diags = append(diags, s.err)
		}
	}
	return diags
}

// Statements returns the parse trees of the successfully parsed statements,
// in document order.
func (doc *Document) Statements() []*Node {
	var nodes []*Node
	for _, s := range doc.stmts {
		if s.node != nil {
			nodes = append(nodes, s.node)
		}
	}
	return nodes
}

// NodeAt returns the innermost AST node whose span contains the byte
// offset, or nil if the offset is outside every parsed statement.
func (doc *Document) NodeAt(offset int) *Node {
	for _, s := range doc.stmts {
		if s.node == nil || offset < s.node.Span.Start || offset >= s.node.Span.End {
			continue
		}
		return descendAt(s.node, offset)
	}
	return nil
}

func descendAt(n *Node, offset int) *Node {
	for _, child := range n.Children {
		if offset >= child.Span.Start && offset < child.Span.End {
			return descendAt(child, offset)
		}
	}
	return n
}

// Update replaces the document text, re-parsing only the statements that
// the edit touched. Statements fully inside the unchanged prefix keep their
// trees; statements fully inside the unchanged suffix keep their trees with
// spans shifted by the edit delta; the middle is re-parsed.
func (doc *Document) Update(text string) {
	oldText := doc.text
	oldStmts := doc.stmts
	doc.text = text
	doc.stats = DocumentStats{}

	// Unchanged prefix/suffix (non-overlapping).
	prefix := commonPrefix(oldText, text)
	suffix := commonSuffix(oldText, text)
	if maxOverlap := min(len(oldText), len(text)) - prefix; suffix > maxOverlap {
		suffix = maxOverlap
	}
	delta := len(text) - len(oldText)
	oldSuffixStart := len(oldText) - suffix

	// Index reusable old statements by their start offset.
	oldByStart := make(map[int]docStatement, len(oldStmts))
	for _, s := range oldStmts {
		if s.node != nil {
			oldByStart[s.span.Start] = s
		}
	}

	tokens, lexErrs := tokenizeTolerant(doc.dsl.grammar, text)

	doc.stmts = doc.stmts[:0]
	for _, e := range lexErrs {
		doc.stmts = append(doc.stmts, docStatement{err: e, span: Span{Start: e.Position, End: e.Position + 1}})
	}

	startRule := doc.dsl.grammar.startRule
	if startRule == "" {
		return
	}

	parser := newASTParser(doc.dsl.grammar)
	parser.input = text
	parser.tokens = tokens
	sync := firstTokens(doc.dsl.grammar, startRule)

	// tokenAt maps a byte offset to the index of the first token starting
	// at or after it.
	tokenAt := func(offset int) int {
		return sort.Search(len(tokens), func(i int) bool { return tokens[i].Start >= offset })
	}

	pos := 0
	errCount := 0
	for pos < len(tokens) && errCount < maxDiagnostics {
		tokStart := tokens[pos].Start

		// Reuse from the unchanged prefix: identical offsets, tree as-is.
		if old, ok := oldByStart[tokStart]; ok && old.span.End <= prefix && tokStart < prefix {
			doc.stmts = append(doc.stmts, old)
			doc.stats.Reused++
			pos = tokenAt(old.span.End)
			continue
		}

		// Reuse from the unchanged suffix: same content shifted by delta.
		if tokStart >= oldSuffixStart+delta && tokStart-delta >= 0 {
			if old, ok := oldByStart[tokStart-delta]; ok && old.span.Start >= oldSuffixStart {
				shifted := docStatement{
					node: shiftNode(old.node, delta),
					span: Span{Start: old.span.Start + delta, End: old.span.End + delta},
				}
				doc.stmts = append(doc.stmts, shifted)
				doc.stats.Reused++
				pos = tokenAt(shifted.span.End)
				continue
			}
		}

		// Parse this statement from scratch.
		parser.pos = pos
		parser.failure = parseFailure{tokenPos: -1}
		parser.ruleStack = nil
		parser.hintStack = nil

		node, err := parser.parseRule(startRule)
		if err == nil && parser.pos > pos {
			doc.stmts = append(doc.stmts, docStatement{node: node, span: node.Span})
			doc.stats.Reparsed++
			pos = parser.pos
			continue
		}
		if err == nil {
			pos++ // zero-width match: force progress
			continue
		}

		perr, _ := parser.farthestError(startRule).(*ParseError)
		if perr != nil {
			doc.stmts = append(doc.stmts, docStatement{err: perr, span: Span{Start: perr.Position, End: perr.Position + 1}})
			errCount++
		}

		next := parser.failure.tokenPos
		if next <= pos {
			next = pos + 1
		}
		for next < len(tokens) && !sync[tokens[next].TokenType] {
			next++
		}
		if next <= pos {
			next = pos + 1
		}
		pos = next
	}
}

// shiftNode returns a copy of the tree with every span moved by delta.
// Reused suffix statements need their positions rebased after an edit;
// copying keeps previously returned trees immutable.
func shiftNode(n *Node, delta int) *Node {
	if n == nil {
		return nil
	}
	if delta == 0 {
		return n
	}
	out := &Node{
		Rule:   n.Rule,
		Action: n.Action,
		Span:   Span{Start: n.Span.Start + delta, End: n.Span.End + delta},
		group:  n.group,
	}
	if n.Token != nil {
		tok := *n.Token
		tok.Start += delta
		tok.End += delta
		out.Token = &tok
	}
	if len(n.Children) > 0 {
		out.Children = make([]*Node, len(n.Children))
		for i, c := range n.Children {
			out.Children[i] = shiftNode(c, delta)
		}
	}
	return out
}

func commonPrefix(a, b string) int {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}

func commonSuffix(a, b string) int {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[len(a)-1-i] == b[len(b)-1-i] {
		i++
	}
	return i
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
