package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"gopkg.in/yaml.v3"
)

type ASTNode struct {
	Type     string     `json:"type" yaml:"type"`
	Value    string     `json:"value,omitempty" yaml:"value,omitempty"`
	Action   string     `json:"action,omitempty" yaml:"action,omitempty"`
	Children []ASTNode  `json:"children,omitempty" yaml:"children,omitempty"`
	Token    *TokenInfo `json:"token,omitempty" yaml:"token,omitempty"`
	Rule     *RuleInfo  `json:"rule,omitempty" yaml:"rule,omitempty"`
}

type TokenInfo struct {
	Type  string `json:"type" yaml:"type"`
	Value string `json:"value" yaml:"value"`
	Line  int    `json:"line" yaml:"line"`
	Col   int    `json:"col" yaml:"col"`
}

type RuleInfo struct {
	Name    string   `json:"name" yaml:"name"`
	Pattern []string `json:"pattern" yaml:"pattern"`
}

type ASTViewer struct {
	dsl     *dslbuilder.DSL
	format  string
	indent  bool
	verbose bool
}

func main() {
	var (
		dslFile   string
		input     string
		format    string
		indent    bool
		verbose   bool
		inputFile string
	)

	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.StringVar(&input, "input", "", "Input string to parse")
	flag.StringVar(&inputFile, "file", "", "Input file to parse")
	flag.StringVar(&format, "format", "json", "Output format: json, yaml, or tree")
	flag.BoolVar(&indent, "indent", true, "Indent output (for json/yaml)")
	flag.BoolVar(&verbose, "verbose", false, "Show detailed token and rule information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "AST Viewer - Visualize the Abstract Syntax Tree of your DSL\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if dslFile == "" || (input == "" && inputFile == "") {
		flag.Usage()
		os.Exit(1)
	}

	dsl, err := loadDSL(dslFile)
	if err != nil {
		log.Fatalf("Error loading DSL: %v", err)
	}

	if inputFile != "" {
		content, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatalf("Error reading input file: %v", err)
		}
		input = string(content)
	}

	viewer := &ASTViewer{
		dsl:     dsl,
		format:  format,
		indent:  indent,
		verbose: verbose,
	}

	if err := viewer.visualize(input); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func loadDSL(filename string) (*dslbuilder.DSL, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")

	switch ext {
	case "yaml", "yml":
		return dslbuilder.LoadFromYAMLFile(filename)
	case "json":
		return dslbuilder.LoadFromJSONFile(filename)
	default:
		if ext == "" {
			return nil, fmt.Errorf("missing file extension in %q (expected .yaml, .yml, or .json)", filename)
		}
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (v *ASTViewer) visualize(input string) error {
	// ParseAST builds the exact lossless syntax tree without requiring Go semantic actions.
	node, err := v.dsl.ParseAST(input)
	if err != nil {
		return fmt.Errorf("parsing error: %v", dslbuilder.GetDetailedError(err))
	}

	ast := v.buildFromNode(node, input)

	switch v.format {
	case "json":
		return v.outputJSON(ast)
	case "yaml":
		return v.outputYAML(ast)
	case "tree":
		return v.outputTree(ast, 0)
	default:
		return fmt.Errorf("unsupported format: %s", v.format)
	}
}

func lineAndColAt(source string, byteOffset int) (int, int) {
	if byteOffset < 0 {
		byteOffset = 0
	}
	if byteOffset > len(source) {
		byteOffset = len(source)
	}
	line := strings.Count(source[:byteOffset], "\n") + 1
	lineStart := strings.LastIndexByte(source[:byteOffset], '\n') + 1
	col := utf8.RuneCountInString(source[lineStart:byteOffset]) + 1
	return line, col
}

// buildFromNode converts a real parse tree (*dslbuilder.Node) into the
// viewer's display structure with full Token, Rule, Action, Line, and Col metadata.
func (v *ASTViewer) buildFromNode(node *dslbuilder.Node, source string) ASTNode {
	if node == nil {
		return ASTNode{Type: "empty"}
	}
	if node.IsToken() {
		line, col := lineAndColAt(source, node.Token.Start)
		out := ASTNode{
			Type:  node.Token.TokenType,
			Value: node.Token.Value,
		}
		if v.verbose {
			out.Token = &TokenInfo{
				Type:  node.Token.TokenType,
				Value: node.Token.Value,
				Line:  line,
				Col:   col,
			}
		}
		return out
	}

	out := ASTNode{
		Type:   node.Rule,
		Action: node.Action,
	}
	var pattern []string
	for _, child := range node.Children {
		if child.IsToken() {
			pattern = append(pattern, child.Token.TokenType)
		} else {
			pattern = append(pattern, child.Rule)
		}
		out.Children = append(out.Children, v.buildFromNode(child, source))
	}
	if v.verbose {
		out.Rule = &RuleInfo{
			Name:    node.Rule,
			Pattern: pattern,
		}
	}
	return out
}

func (v *ASTViewer) outputJSON(ast ASTNode) error {
	var data []byte
	var err error
	if v.indent {
		data, err = json.MarshalIndent(ast, "", "  ")
	} else {
		data, err = json.Marshal(ast)
	}
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (v *ASTViewer) outputYAML(ast ASTNode) error {
	data, err := yaml.Marshal(ast)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func (v *ASTViewer) outputTree(ast ASTNode, depth int) error {
	v.outputTreeNode(ast, "", depth == 0)
	return nil
}

func (av *ASTViewer) outputTreeNode(node ASTNode, prefix string, isRoot bool) {
	symbol := "○"
	if len(node.Children) > 0 {
		symbol = "◆"
	} else if node.Value != "" {
		symbol = "●"
	}

	if isRoot {
		fmt.Printf("%s %s", symbol, node.Type)
	} else {
		fmt.Printf("%s%s %s", prefix, symbol, node.Type)
	}

	if node.Value != "" {
		fmt.Printf(": %q", node.Value)
	}
	if node.Action != "" {
		fmt.Printf(" \033[35m→ %s\033[0m", node.Action)
	}
	fmt.Println()

	if av.verbose {
		if node.Token != nil {
			fmt.Printf("%s  ├─ token: %s = %q @ %d:%d\n",
				prefix, node.Token.Type, node.Token.Value, node.Token.Line, node.Token.Col)
		}
		if node.Rule != nil {
			fmt.Printf("%s  ├─ rule: %s → %v\n",
				prefix, node.Rule.Name, node.Rule.Pattern)
		}
	}

	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		var childPrefix string
		if isRoot {
			if isLast {
				fmt.Print("└─ ")
				childPrefix = "   "
			} else {
				fmt.Print("├─ ")
				childPrefix = "│  "
			}
		} else {
			if isLast {
				fmt.Printf("%s└─ ", prefix)
				childPrefix = prefix + "   "
			} else {
				fmt.Printf("%s├─ ", prefix)
				childPrefix = prefix + "│  "
			}
		}
		av.outputTreeNode(child, childPrefix, false)
	}
}
