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
	return v.buildASTRecursive(result, "root", 0)
}

func (av *ASTViewer) buildASTRecursive(data interface{}, nodeType string, depth int) ASTNode {
	node := ASTNode{
		Type: nodeType,
	}

	// Handle different types of data
	switch v := data.(type) {
		
	case []interface{}:
		// Array of values (common in parsing results)
		node.Type = nodeType
		if nodeType == "root" {
			node.Type = "expression"
		}
		for i, item := range v {
			childType := "value"
			// Try to determine better types based on position
			if i%2 == 1 && len(v) > 2 {
				// Likely an operator in expression
				if str, ok := item.(string); ok && isOperator(str) {
					childType = "operator"
				}
			}
			child := av.buildASTRecursive(item, childType, depth+1)
			node.Children = append(node.Children, child)
		}
		
	case map[string]interface{}:
		// Object/map structure
		node.Type = "object"
		for key, value := range v {
			child := av.buildASTRecursive(value, key, depth+1)
			node.Children = append(node.Children, child)
		}
		
	case string:
		node.Value = v
		// Detect token types
		if isNumber(v) {
			node.Type = "number"
		} else if isOperator(v) {
			node.Type = "operator"
		} else if nodeType == "value" {
			node.Type = "identifier"
		}
		
	case int, int64, float64:
		node.Type = "number"
		node.Value = fmt.Sprintf("%v", v)
		
	case bool:
		node.Type = "boolean"
		node.Value = fmt.Sprintf("%v", v)
		
	default:
		// Generic handling
		node.Value = fmt.Sprintf("%v", v)
		if node.Value == "" {
			node.Value = fmt.Sprintf("<%T>", v)
		}
	}

	return node
}

func isOperator(s string) bool {
	operators := []string{"+", "-", "*", "/", "=", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "and", "or", "not"}
	for _, op := range operators {
		if s == op {
			return true
		}
	}
	return false
}

func isNumber(s string) bool {
	if _, err := fmt.Sscanf(s, "%f", new(float64)); err == nil {
		return true
	}
	return false
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
	// Determine node symbol based on type
	symbol := "○"
	switch node.Type {
	case "expression", "root":
		symbol = "◆"
	case "operator":
		symbol = "●"
	case "number":
		symbol = "#"
	case "identifier":
		symbol = "□"
	case "object":
		symbol = "{}"
	case "boolean":
		symbol = "?"
	}

	// Print current node
	if isRoot {
		fmt.Printf("%s %s", symbol, node.Type)
	} else {
		fmt.Printf("%s%s %s", prefix, symbol, node.Type)
	}
	
	if node.Value != "" {
		// Color-code values based on type
		switch node.Type {
		case "number":
			fmt.Printf(" \033[36m%s\033[0m", node.Value) // Cyan for numbers
		case "operator":
			fmt.Printf(" \033[33m%s\033[0m", node.Value) // Yellow for operators
		case "string", "identifier":
			fmt.Printf(" \033[32m\"%s\"\033[0m", node.Value) // Green for strings
		default:
			fmt.Printf(": %s", node.Value)
		}
	}
	
	if node.Action != "" {
		fmt.Printf(" \033[35m→ %s\033[0m", node.Action) // Magenta for actions
	}
	fmt.Println()

	// Show additional info in verbose mode
	if av.verbose {
		if node.Token != nil {
			fmt.Printf("%s  ├─ token: %s = \"%s\" @ %d:%d\n",
				prefix, node.Token.Type, node.Token.Value, node.Token.Line, node.Token.Col)
		}
		if node.Rule != nil {
			fmt.Printf("%s  ├─ rule: %s → %v\n",
				prefix, node.Rule.Name, node.Rule.Pattern)
		}
	}

	// Print children
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
