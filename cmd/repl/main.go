package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type REPL struct {
	dsl         *dslbuilder.DSL
	history     []HistoryEntry
	context     map[string]interface{}
	multiline   bool
	buffer      []string
	showAST     bool
	showTime    bool
	historyFile string
	historyIdx  int
	tokens      []string // For autocomplete
	rules       []string // For autocomplete
}

type HistoryEntry struct {
	Index     int
	Input     string
	Output    interface{}
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

func main() {
	var (
		dslFile     string
		historyFile string
		contextFile string
		showAST     bool
		showTime    bool
		multiline   bool
		commands    []string
	)

	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.StringVar(&historyFile, "history", "", "History file to save/load commands")
	flag.StringVar(&contextFile, "context", "", "Context file (JSON) to preload")
	flag.BoolVar(&showAST, "ast", false, "Show AST representation of parsed input")
	flag.BoolVar(&showTime, "time", false, "Show execution time for each command")
	flag.BoolVar(&multiline, "multiline", false, "Enable multiline input mode")
	flag.Func("exec", "Execute commands (can be used multiple times)", func(s string) error {
		commands = append(commands, s)
		return nil
	})

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "DSL REPL - Interactive Read-Eval-Print Loop for your DSL\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  .help        Show available REPL commands\n")
		fmt.Fprintf(os.Stderr, "  .exit        Exit the REPL\n")
		fmt.Fprintf(os.Stderr, "  .history     Show command history\n")
		fmt.Fprintf(os.Stderr, "  .clear       Clear the screen\n")
		fmt.Fprintf(os.Stderr, "  .context     Show current context\n")
		fmt.Fprintf(os.Stderr, "  .set <k> <v> Set context variable\n")
		fmt.Fprintf(os.Stderr, "  .load <file> Load and execute commands from file\n")
		fmt.Fprintf(os.Stderr, "  .save <file> Save history to file\n")
		fmt.Fprintf(os.Stderr, "  .ast on/off  Toggle AST display\n")
		fmt.Fprintf(os.Stderr, "  .time on/off Toggle execution time display\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -dsl calculator.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl query.json -context data.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dsl accounting.yaml -exec \"venta de 1000\" -exec \"venta de 2000\"\n", os.Args[0])
	}

	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Load DSL
	dsl, err := loadDSL(dslFile)
	if err != nil {
		log.Fatalf("Error loading DSL: %v", err)
	}

	// Create REPL instance
	repl := &REPL{
		dsl:         dsl,
		history:     []HistoryEntry{},
		context:     make(map[string]interface{}),
		multiline:   multiline,
		buffer:      []string{},
		showAST:     showAST,
		showTime:    showTime,
		historyFile: historyFile,
	}

	// Load context if provided
	if contextFile != "" {
		if err := repl.loadContext(contextFile); err != nil {
			log.Printf("Warning: Failed to load context: %v", err)
		}
	}

	// Load history if file specified
	if historyFile != "" {
		repl.loadHistory()
	}

	// Execute commands if provided
	for _, cmd := range commands {
		repl.execute(cmd)
	}

	// Start interactive mode if no commands or after executing commands
	if len(commands) == 0 || isInteractive() {
		repl.run()
	}

	// Save history on exit
	if historyFile != "" {
		repl.saveHistory()
	}
}

func loadDSL(filename string) (*dslbuilder.DSL, error) {
	// First try to load from config file
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])

	var dsl *dslbuilder.DSL
	var err error

	switch ext {
	case "yaml", "yml":
		dsl, err = dslbuilder.LoadFromYAMLFile(filename)
	case "json":
		dsl, err = dslbuilder.LoadFromJSONFile(filename)
	case "go":
		// For .go files, we would need to compile and load
		// This is a placeholder for future implementation
		return nil, fmt.Errorf("loading from Go files not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	// Register example actions (in real use, these would be provided by the DSL definition)
	registerExampleActions(dsl)

	return dsl, nil
}

func registerExampleActions(dsl *dslbuilder.DSL) {
	// Register common actions to prevent errors
	// In a real implementation, actions would be defined in the DSL file or separately

	// Math operations
	dsl.Action("add", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toNumber(args[0])
			right, _ := toNumber(args[2])
			return left + right, nil
		}
		return nil, fmt.Errorf("invalid arguments for add")
	})

	dsl.Action("subtract", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toNumber(args[0])
			right, _ := toNumber(args[2])
			return left - right, nil
		}
		return nil, fmt.Errorf("invalid arguments for subtract")
	})

	dsl.Action("multiply", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toNumber(args[0])
			right, _ := toNumber(args[2])
			return left * right, nil
		}
		return nil, fmt.Errorf("invalid arguments for multiply")
	})

	dsl.Action("divide", func(args []interface{}) (interface{}, error) {
		if len(args) >= 3 {
			left, _ := toNumber(args[0])
			right, _ := toNumber(args[2])
			if right == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
		return nil, fmt.Errorf("invalid arguments for divide")
	})

	// Generic passthrough
	dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return args, nil
	})

	// Number parsing
	dsl.Action("number", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return toNumber(args[0])
		}
		return nil, fmt.Errorf("no number provided")
	})
}

func toNumber(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case string:
		var num float64
		_, err := fmt.Sscanf(n, "%f", &num)
		return num, err
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func (r *REPL) run() {
	fmt.Printf("DSL REPL\n")
	fmt.Println("Type '.help' for help, '.exit' to quit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Show prompt
		if r.multiline && len(r.buffer) > 0 {
			fmt.Printf("... ")
		} else {
			fmt.Printf("DSL> ")
		}

		// Read input
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		// Handle multiline mode
		if r.multiline {
			if input == "" && len(r.buffer) > 0 {
				// Empty line ends multiline input
				fullInput := strings.Join(r.buffer, "\n")
				r.buffer = []string{}
				r.execute(fullInput)
			} else if input != "" {
				r.buffer = append(r.buffer, input)
			}
			continue
		}

		// Process single line input
		r.processInput(input)
	}

	fmt.Println("\nGoodbye!")
}

func (r *REPL) processInput(input string) {
	// Handle REPL commands
	if strings.HasPrefix(input, ".") {
		r.handleCommand(input)
		return
	}

	// Execute DSL input
	r.execute(input)
}

func (r *REPL) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case ".help":
		r.showHelp()
	case ".exit", ".quit":
		os.Exit(0)
	case ".history":
		r.showHistory()
	case ".clear":
		fmt.Print("\033[H\033[2J")
	case ".context":
		r.showContext()
	case ".set":
		if len(parts) >= 3 {
			r.setContext(parts[1], strings.Join(parts[2:], " "))
		} else {
			fmt.Println("Usage: .set <key> <value>")
		}
	case ".load":
		if len(parts) >= 2 {
			r.loadFile(parts[1])
		} else {
			fmt.Println("Usage: .load <filename>")
		}
	case ".save":
		if len(parts) >= 2 {
			r.saveHistoryToFile(parts[1])
		} else {
			fmt.Println("Usage: .save <filename>")
		}
	case ".ast":
		if len(parts) >= 2 {
			r.showAST = parts[1] == "on"
			fmt.Printf("AST display: %v\n", r.showAST)
		} else {
			fmt.Printf("AST display: %v\n", r.showAST)
		}
	case ".time":
		if len(parts) >= 2 {
			r.showTime = parts[1] == "on"
			fmt.Printf("Time display: %v\n", r.showTime)
		} else {
			fmt.Printf("Time display: %v\n", r.showTime)
		}
	case ".multiline":
		r.multiline = !r.multiline
		fmt.Printf("Multiline mode: %v\n", r.multiline)
		if r.multiline {
			fmt.Println("Enter empty line to execute")
		}
	case ".tokens":
		r.showTokens()
	case ".rules":
		r.showRules()
	case ".reset":
		r.context = make(map[string]interface{})
		r.buffer = []string{}
		fmt.Println("Context and buffer reset")
	case ".last":
		if len(r.history) > 0 {
			last := r.history[len(r.history)-1]
			fmt.Printf("Last command: %s\n", last.Input)
			if last.Error == nil && last.Output != nil {
				fmt.Printf("Result: %v\n", last.Output)
			}
		} else {
			fmt.Println("No history available")
		}
	default:
		fmt.Printf("Unknown command: %s\n", parts[0])
		fmt.Println("Type .help for available commands")
	}
}

func (r *REPL) execute(input string) {
	if strings.TrimSpace(input) == "" {
		return
	}

	start := time.Now()

	// Parse with context
	result, err := r.dsl.Use(input, r.context)

	duration := time.Since(start)

	// Record in history
	entry := HistoryEntry{
		Index:     len(r.history) + 1,
		Input:     input,
		Output:    nil,
		Error:     err,
		Timestamp: start,
		Duration:  duration,
	}

	if err != nil {
		fmt.Printf("\033[31mError: %v\033[0m\n", err) // Red color for errors
		
		// Try to provide helpful suggestions
		if strings.Contains(err.Error(), "unexpected token") {
			r.suggestTokens(input)
		} else if strings.Contains(err.Error(), "no matching rule") {
			fmt.Println("\033[33mHint: Check available rules with .rules command\033[0m")
		}
	} else {
		output := result.GetOutput()
		entry.Output = output

		// Display result with better formatting
		r.displayOutput(output)

		// Show AST if enabled
		if r.showAST {
			r.displayAST(result)
		}
	}

	// Show execution time if enabled
	if r.showTime {
		fmt.Printf("⏱  %v\n", duration)
	}

	r.history = append(r.history, entry)
}

func (r *REPL) displayAST(result interface{}) {
	fmt.Println("\n--- AST ---")
	// Simplified AST display
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	fmt.Println("--- End AST ---\n")
}


func (r *REPL) showHistory() {
	if len(r.history) == 0 {
		fmt.Println("No history")
		return
	}

	for _, entry := range r.history {
		fmt.Printf("[%d] %s", entry.Index, entry.Input)
		if entry.Error != nil {
			fmt.Printf(" => Error: %v", entry.Error)
		} else if entry.Output != nil {
			fmt.Printf(" => %v", entry.Output)
		}
		fmt.Println()
	}
}

func (r *REPL) showContext() {
	if len(r.context) == 0 {
		fmt.Println("Context is empty")
		return
	}

	fmt.Println("Current context:")
	for k, v := range r.context {
		fmt.Printf("  %s: %v\n", k, v)
	}
}

func (r *REPL) setContext(key, value string) {
	// Try to parse value as JSON first
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err == nil {
		r.context[key] = parsed
	} else {
		// Otherwise store as string
		r.context[key] = value
	}
	fmt.Printf("Set %s = %v\n", key, r.context[key])
}

func (r *REPL) loadContext(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.context)
}

func (r *REPL) loadFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fmt.Printf("[%s:%d] %s\n", filename, lineNum, line)
		r.execute(line)
	}
}

func (r *REPL) saveHistoryToFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error saving history: %v\n", err)
		return
	}
	defer file.Close()

	for _, entry := range r.history {
		fmt.Fprintf(file, "# %s\n", entry.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(file, "%s\n", entry.Input)
		if entry.Error != nil {
			fmt.Fprintf(file, "# Error: %v\n", entry.Error)
		} else if entry.Output != nil {
			fmt.Fprintf(file, "# => %v\n", entry.Output)
		}
		fmt.Fprintln(file)
	}

	fmt.Printf("History saved to %s\n", filename)
}

func (r *REPL) loadHistory() {
	// Implementation would load history from file
	// For now, this is a placeholder
}

func (r *REPL) saveHistory() {
	if r.historyFile != "" {
		r.saveHistoryToFile(r.historyFile)
	}
}

func isInteractive() bool {
	fi, _ := os.Stdin.Stat()
	return fi.Mode()&os.ModeCharDevice != 0
}

func (r *REPL) displayOutput(output interface{}) {
	switch v := output.(type) {
	case string:
		fmt.Println(v)
	case int, int64, float64:
		fmt.Printf("\033[36m%v\033[0m\n", v) // Cyan for numbers
	case bool:
		fmt.Printf("\033[35m%v\033[0m\n", v) // Magenta for booleans
	case []interface{}:
		if len(v) == 0 {
			fmt.Println("[]")
		} else {
			fmt.Println("[")
			for i, item := range v {
				fmt.Printf("  [%d] %v\n", i, item)
			}
			fmt.Println("]")
		}
	case map[string]interface{}:
		data, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(data))
	case nil:
		fmt.Println("\033[90mnil\033[0m") // Gray for nil
	default:
		// Try JSON for complex types
		if data, err := json.MarshalIndent(output, "", "  "); err == nil {
			fmt.Println(string(data))
		} else {
			fmt.Printf("%v\n", output)
		}
	}
}

func (r *REPL) suggestTokens(input string) {
	// This would need access to DSL tokens to provide suggestions
	fmt.Println("\033[33mHint: Use .tokens to see available tokens\033[0m")
}

func (r *REPL) showTokens() {
	fmt.Println("\033[1mAvailable Tokens:\033[0m")
	// In a real implementation, we would introspect the DSL
	// For now, show a message
	fmt.Println("Token information not available in current implementation")
	fmt.Println("Check your DSL configuration file for token definitions")
}

func (r *REPL) showRules() {
	fmt.Println("\033[1mAvailable Rules:\033[0m")
	// In a real implementation, we would introspect the DSL
	// For now, show a message
	fmt.Println("Rule information not available in current implementation")
	fmt.Println("Check your DSL configuration file for rule definitions")
}

func (r *REPL) showHelp() {
	fmt.Println("\033[1mREPL Commands:\033[0m")
	fmt.Println("  \033[32m.help\033[0m         Show this help message")
	fmt.Println("  \033[32m.exit\033[0m         Exit the REPL")
	fmt.Println("  \033[32m.history\033[0m      Show command history")
	fmt.Println("  \033[32m.clear\033[0m        Clear the screen")
	fmt.Println("  \033[32m.context\033[0m      Show current context variables")
	fmt.Println("  \033[32m.set k v\033[0m      Set context variable k to value v")
	fmt.Println("  \033[32m.load file\033[0m    Load and execute commands from file")
	fmt.Println("  \033[32m.save file\033[0m    Save history to file")
	fmt.Println("  \033[32m.ast on/off\033[0m   Toggle AST display")
	fmt.Println("  \033[32m.time on/off\033[0m  Toggle execution time display")
	fmt.Println("  \033[32m.multiline\033[0m    Toggle multiline input mode")
	fmt.Println("  \033[32m.tokens\033[0m       Show available tokens")
	fmt.Println("  \033[32m.rules\033[0m        Show available rules")
	fmt.Println("  \033[32m.reset\033[0m        Reset context and buffer")
	fmt.Println("  \033[32m.last\033[0m         Show last command and result")
	fmt.Println()
	fmt.Println("\033[1mDSL Syntax:\033[0m")
	fmt.Println("  Enter DSL commands directly")
	fmt.Println("  Use context variables with .set command")
	fmt.Println("  Check your DSL configuration for available syntax")
}
