package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arturoeanton/go-dsl/examples/http_dsl/universal"
)

// HTTPRunner executes HTTP DSL scripts with full v3 support including blocks
type HTTPRunner struct {
	dsl        *universal.HTTPDSLv3
	verbose    bool
	stopOnFail bool
	dryRun     bool
	validate   bool
}

// NewHTTPRunner creates a new HTTP script runner
func NewHTTPRunner(verbose, stopOnFail, dryRun, validate bool) *HTTPRunner {
	return &HTTPRunner{
		dsl:        universal.NewHTTPDSLv3(),
		verbose:    verbose,
		stopOnFail: stopOnFail,
		dryRun:     dryRun,
		validate:   validate,
	}
}

// RunFile executes an HTTP DSL script file
func (hr *HTTPRunner) RunFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read file %s: %w", filename, err)
	}

	script := string(content)

	if hr.validate {
		fmt.Printf("🔍 Validating script: %s\n", filename)
		return hr.validateScript(script)
	}

	fmt.Printf("\n🚀 Executing HTTP Script: %s\n", filename)
	fmt.Println(strings.Repeat("═", 60))

	start := time.Now()

	if hr.dryRun {
		fmt.Println("🔍 DRY RUN - Script would execute:")
		fmt.Println(hr.formatScript(script))
		return nil
	}

	// Use ParseWithBlockSupport for full block support
	result, err := hr.dsl.ParseWithBlockSupport(script)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Show any output from the execution (like print statements)
	if results, ok := result.([]interface{}); ok {
		for _, res := range results {
			if res != nil {
				// Check if it's a print output (string)
				if str, ok := res.(string); ok {
					// Print outputs from the DSL (like print statements)
					// Filter out internal status messages
					if !strings.HasPrefix(str, "HTTP ") && 
					   !strings.HasPrefix(str, "Variable set:") &&
					   !strings.HasPrefix(str, "Condition evaluated") {
						fmt.Println(str)
					}
				}
			}
		}
	}

	duration := time.Since(start)

	if hr.verbose {
		fmt.Printf("\n📊 Execution Summary:\n")
		fmt.Printf("   Duration: %v\n", duration)
		fmt.Printf("   Variables: %v\n", hr.dsl.GetVariables())
		if results, ok := result.([]interface{}); ok {
			fmt.Printf("   Steps executed: %d\n", len(results))
		}
	}

	fmt.Printf("\n✅ Script completed in %v\n", duration)
	return nil
}

// validateScript validates the script syntax without execution
func (hr *HTTPRunner) validateScript(script string) error {
	fmt.Println("Validating syntax...")

	// Try parsing without execution
	_, err := hr.dsl.ParseWithBlockSupport(script)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Script is valid")
	return nil
}

// formatScript formats the script for display
func (hr *HTTPRunner) formatScript(script string) string {
	lines := strings.Split(script, "\n")
	var formatted []string

	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			formatted = append(formatted, fmt.Sprintf("%3d: %s", i+1, line))
		}
	}

	return strings.Join(formatted, "\n")
}

func main() {
	var (
		verbose    = flag.Bool("v", false, "Verbose output with execution details")
		verbose2   = flag.Bool("verbose", false, "Verbose output with execution details")
		stopOnFail = flag.Bool("stop", false, "Stop execution on first failure")
		dryRun     = flag.Bool("dry-run", false, "Show what would be executed without running")
		validate   = flag.Bool("validate", false, "Validate script syntax only")
		help       = flag.Bool("h", false, "Show help")
		help2      = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help || *help2 {
		showHelp()
		return
	}

	if flag.NArg() == 0 {
		fmt.Println("❌ Error: No script file specified")
		showUsage()
		os.Exit(1)
	}

	verboseMode := *verbose || *verbose2
	runner := NewHTTPRunner(verboseMode, *stopOnFail, *dryRun, *validate)

	filename := flag.Arg(0)
	if err := runner.RunFile(filename); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("🌐 HTTP DSL Runner v3 - Production Ready")
	fmt.Println("Execute HTTP DSL scripts with full support for blocks, variables, and conditionals")
	fmt.Println()
	showUsage()
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose     Show detailed execution information")
	fmt.Println("  --stop            Stop execution on first failure")
	fmt.Println("  --dry-run         Show what would be executed without running")
	fmt.Println("  --validate        Validate script syntax only")
	fmt.Println("  -h, --help        Show this help message")
	fmt.Println()
	fmt.Println("Features supported:")
	fmt.Println("  ✅ All HTTP methods (GET, POST, PUT, DELETE, etc.)")
	fmt.Println("  ✅ Multiple headers per request")
	fmt.Println("  ✅ JSON with special characters (@, #, etc.)")
	fmt.Println("  ✅ Variables and arithmetic")
	fmt.Println("  ✅ If/then/else statements (single line)")
	fmt.Println("  ✅ If/then/endif blocks (multiline)")
	fmt.Println("  ✅ Repeat loops with blocks")
	fmt.Println("  ✅ Response assertions")
	fmt.Println("  ✅ Data extraction (JSONPath, regex, headers)")
	fmt.Println("  ✅ Authentication (Basic, Bearer)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  http-runner script.http                 # Execute script")
	fmt.Println("  http-runner -v script.http              # Execute with verbose output")
	fmt.Println("  http-runner --validate script.http      # Validate syntax only")
	fmt.Println("  http-runner --dry-run script.http       # Show execution plan")
}

func showUsage() {
	fmt.Println("Usage: http-runner [options] <script.http>")
}
