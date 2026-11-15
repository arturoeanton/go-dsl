# Code Generator & VS Code Extension Scaffolder (`cmd/dslgen`)

Generates static Go parser code or a complete VS Code extension from a YAML/JSON grammar.

## Usage

```bash
# Generate Go constructor
go run ./cmd/dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -o mydsl_gen.go

# Generate Go constructor + strongly-typed Actions interface
go run ./cmd/dslgen -dsl grammar.yaml -package mydsl -func NewMyDSL -typed -o mydsl_gen.go

# Generate a ready-to-launch VS Code extension (TextMate + LSP client)
go run ./cmd/dslgen -dsl grammar.yaml -vscode ./vscode-mydsl
```
