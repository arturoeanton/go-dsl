package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -dsl calculator.yaml -input \"10 + 20\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl query.json -file queries.txt -format tree\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl accounting.yaml -input \"venta de 1000 con iva\" -format yaml -verbose\n", os.Args[0])
	}

	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	if input == "" && inputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load DSL
	dsl, err := loadDSL(dslFile)
	if err != nil {
		log.Fatalf("Error loading DSL: %v", err)
	}

	// Register dummy actions to prevent parsing errors
	registerDummyActions(dsl)

	// Get input
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

	// Parse and visualize
	if err := viewer.visualize(input); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func loadDSL(filename string) (*dslbuilder.DSL, error) {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])

	switch ext {
	case "yaml", "yml":
		return dslbuilder.LoadFromYAMLFile(filename)
	case "json":
		return dslbuilder.LoadFromJSONFile(filename)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func registerDummyActions(dsl *dslbuilder.DSL) {
	// Register dummy actions for all rules to prevent parsing errors
	// This is a simplified approach - in a real implementation,
	// we would introspect the DSL to find all action names
	actions := []string{
		"passthrough", "add", "subtract", "multiply", "divide",
		"number", "parentheses", "saleWithTax", "simpleSale",
		"processEntry", "singleMovement", "multipleMovements",
		"debitMovement", "creditMovement", "greet", "getValue",
		"findField", "selectField", "whereCondition",
	}

	for _, action := range actions {
		dsl.Action(action, func(args []interface{}) (interface{}, error) {
			return args, nil
		})
	}
}

func (v *ASTViewer) visualize(input string) error {
	// Parse the input
	result, err := v.dsl.Parse(input)
	if err != nil {
		return fmt.Errorf("parsing error: %v", err)
	}

	// Build AST representation
	ast := v.buildAST(result)

	// Output based on format
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

func (v *ASTViewer) buildAST(result interface{}) ASTNode {
	// This is a simplified AST builder
	// In a real implementation, we would need access to the parser's internal state
	// to build a proper AST with all parse tree information

	node := ASTNode{
		Type:  "root",
		Value: fmt.Sprintf("%v", result),
	}

	// If result is a slice (from parsing actions), show structure
	if slice, ok := result.([]interface{}); ok {
		node.Children = make([]ASTNode, len(slice))
		for i, item := range slice {
			child := ASTNode{
				Type:  fmt.Sprintf("arg_%d", i),
				Value: fmt.Sprintf("%v", item),
			}
			node.Children[i] = child
		}
	}

	return node
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
	indent := strings.Repeat("  ", depth)
	prefix := "├─"
	if depth == 0 {
		prefix = ""
	}

	fmt.Printf("%s%s %s", indent, prefix, ast.Type)
	if ast.Value != "" {
		fmt.Printf(": %s", ast.Value)
	}
	if ast.Action != "" {
		fmt.Printf(" [action: %s]", ast.Action)
	}
	fmt.Println()

	if v.verbose && ast.Token != nil {
		fmt.Printf("%s  └─ token: %s = \"%s\" @ %d:%d\n",
			indent, ast.Token.Type, ast.Token.Value, ast.Token.Line, ast.Token.Col)
	}
	if v.verbose && ast.Rule != nil {
		fmt.Printf("%s  └─ rule: %s -> %v\n",
			indent, ast.Rule.Name, ast.Rule.Pattern)
	}

	for i, child := range ast.Children {
		if i == len(ast.Children)-1 {
			// Last child
			fmt.Printf("%s└─", strings.Repeat("  ", depth+1))
		}
		v.outputTree(child, depth+1)
	}

	return nil
}
