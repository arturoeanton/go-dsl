// Command dslgen generates Go source code (and optionally a ready-to-use
// VS Code / Cursor extension) from a declarative DSL grammar (YAML or JSON).
//
// Usage:
//
//	dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -o mydsl_gen.go
//	dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -typed -o mydsl_gen.go
//	dslgen -dsl grammar.yaml -vscode ./vscode-mydsl
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
	"github.com/arturoeanton/go-dsl/pkg/dslgen"
	"gopkg.in/yaml.v3"
)

// GenerateOptions aliases dslgen.GenerateOptions for CLI/test callers.
type GenerateOptions = dslgen.GenerateOptions

func main() {
	var (
		dslFile   string
		pkgName   string
		funcName  string
		outFile   string
		typedMode bool
		vscodeDir string
	)
	flag.StringVar(&dslFile, "dsl", "", "DSL configuration file (YAML or JSON)")
	flag.StringVar(&pkgName, "package", "main", "package name for the generated file")
	flag.StringVar(&funcName, "func", "NewGeneratedDSL", "name of the generated constructor function")
	flag.StringVar(&outFile, "o", "", "output file (default: stdout)")
	flag.BoolVar(&typedMode, "typed", false, "generate strongly-typed Actions interface and binder")
	flag.StringVar(&vscodeDir, "vscode", "", "generate a VS Code / Cursor extension bundle in this directory")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dslgen - Generate Go code or VS Code extensions from a declarative DSL grammar\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s -dsl grammar.yaml [-package name] [-func Name] [-typed] [-o out.go] [-vscode dir]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if dslFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	config, err := loadConfig(dslFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if vscodeDir != "" {
		if err := GenerateVSCodeExtension(config, vscodeDir); err != nil {
			fmt.Fprintf(os.Stderr, "error generating VS Code extension: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "generated VS Code extension in %s\n", vscodeDir)
		if outFile == "" && !isFlagPassed("package") && !isFlagPassed("func") && !typedMode {
			return
		}
	}

	source, err := GenerateWithOptions(config, pkgName, funcName, GenerateOptions{Typed: typedMode})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if outFile == "" {
		fmt.Print(string(source))
		return
	}
	if err := os.WriteFile(outFile, source, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "generated %s\n", outFile)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func loadConfig(filename string) (*dslbuilder.DSLConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config dslbuilder.DSLConfig
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	switch ext {
	case "yaml", "yml":
		err = yaml.Unmarshal(data, &config)
	case "json":
		err = json.Unmarshal(data, &config)
	default:
		if ext == "" {
			return nil, fmt.Errorf("missing file extension in %q (expected .yaml, .yml, or .json)", filename)
		}
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Generate delegates to pkg/dslgen.Generate.
func Generate(config *dslbuilder.DSLConfig, pkgName, funcName string) ([]byte, error) {
	return dslgen.Generate(config, pkgName, funcName)
}

// GenerateWithOptions delegates to pkg/dslgen.GenerateWithOptions.
func GenerateWithOptions(config *dslbuilder.DSLConfig, pkgName, funcName string, opts GenerateOptions) ([]byte, error) {
	return dslgen.GenerateWithOptions(config, pkgName, funcName, opts)
}

// GenerateVSCodeExtension delegates to pkg/dslgen.GenerateVSCodeExtension.
func GenerateVSCodeExtension(config *dslbuilder.DSLConfig, outDir string) error {
	return dslgen.GenerateVSCodeExtension(config, outDir)
}
