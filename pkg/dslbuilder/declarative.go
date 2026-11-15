// Package dslbuilder provides declarative configuration support for DSLs
// (v1 backward-compatible + v2 lossless schema supporting TokenDefs,
// Pratt Expressions, StartRule, SignificantNewlines, ErrorHint, Layouts, and Tests).
package dslbuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// DSLConfig represents the declarative configuration for a DSL.
// It supports both legacy v1 maps (Tokens + Rules) and lossless v2 fields
// (TokenDefs, Expressions, StartRule, SignificantNewlines, Layouts, Tests).
type DSLConfig struct {
	Name                string                 `yaml:"name" json:"name"`                                                     // DSL identifier
	StartRule           string                 `yaml:"start_rule,omitempty" json:"start_rule,omitempty"`                     // Optional explicit start rule
	SignificantNewlines bool                   `yaml:"significant_newlines,omitempty" json:"significant_newlines,omitempty"` // Emit NEWLINE tokens
	Tokens              map[string]string      `yaml:"tokens,omitempty" json:"tokens,omitempty"`                             // Legacy token map (preserved for v1 compatibility)
	TokenDefs           []TokenConfig          `yaml:"token_defs,omitempty" json:"token_defs,omitempty"`                     // Lossless ordered v2 token definitions
	Rules               []RuleConfig           `yaml:"rules,omitempty" json:"rules,omitempty"`                               // Grammar rules
	Expressions         []ExpressionConfig     `yaml:"expressions,omitempty" json:"expressions,omitempty"`                   // Pratt expression rules
	Layouts             map[string]RuleLayout  `yaml:"layouts,omitempty" json:"layouts,omitempty"`                           // Optional AST formatting layouts
	Context             map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`                           // Runtime context defaults
	Tests               []DSLTestCase          `yaml:"tests,omitempty" json:"tests,omitempty"`                               // Declarative grammar test suite
}

// TokenConfig represents a lossless ordered token definition in v2 configs.
type TokenConfig struct {
	Name       string `yaml:"name" json:"name"`
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"` // regex (default), keyword, literal, skip, lookaround
	Pattern    string `yaml:"pattern" json:"pattern"`               // Regex pattern or keyword/literal raw text
	Lookahead  string `yaml:"lookahead,omitempty" json:"lookahead,omitempty"`
	Lookbehind string `yaml:"lookbehind,omitempty" json:"lookbehind,omitempty"`
}

// RuleConfig represents a rule alternative in the declarative configuration.
type RuleConfig struct {
	Name          string   `yaml:"name" json:"name"`                                       // Rule identifier
	Pattern       []string `yaml:"pattern" json:"pattern"`                                 // Symbol sequence
	Action        string   `yaml:"action" json:"action"`                                   // Action name
	ErrorHint     string   `yaml:"error_hint,omitempty" json:"error_hint,omitempty"`       // Custom RuleWithError hint
	Precedence    int      `yaml:"precedence,omitempty" json:"precedence,omitempty"`       // Legacy precedence metadata
	Associativity string   `yaml:"associativity,omitempty" json:"associativity,omitempty"` // Legacy associativity metadata
}

// ExpressionConfig represents a Pratt expression rule in declarative v2 configs.
type ExpressionConfig struct {
	Name    string            `yaml:"name" json:"name"`
	Atoms   []ExprAtomConfig  `yaml:"atoms,omitempty" json:"atoms,omitempty"`
	Groups  []ExprGroupConfig `yaml:"groups,omitempty" json:"groups,omitempty"`
	Prefix  []ExprOpConfig    `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Postfix []ExprOpConfig    `yaml:"postfix,omitempty" json:"postfix,omitempty"`
	Infix   []ExprOpConfig    `yaml:"infix,omitempty" json:"infix,omitempty"`
	Calls   []ExprCallConfig  `yaml:"calls,omitempty" json:"calls,omitempty"`
	Indices []ExprIndexConfig `yaml:"indices,omitempty" json:"indices,omitempty"`
	Members []ExprMemberConf  `yaml:"members,omitempty" json:"members,omitempty"`
	Ternary *ExprTernaryConf  `yaml:"ternary,omitempty" json:"ternary,omitempty"`
}

// ExprAtomConfig describes a terminal atom of a Pratt expression.
type ExprAtomConfig struct {
	Token  string `yaml:"token" json:"token"`
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprGroupConfig describes a bracketed group inside a Pratt expression.
type ExprGroupConfig struct {
	Open   string `yaml:"open" json:"open"`
	Inner  string `yaml:"inner" json:"inner"`
	Close  string `yaml:"close" json:"close"`
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprOpConfig describes a unary or binary operator inside a Pratt expression.
type ExprOpConfig struct {
	Token  string `yaml:"token" json:"token"`
	Power  int    `yaml:"power" json:"power"`
	Assoc  string `yaml:"assoc,omitempty" json:"assoc,omitempty"` // left (default), right, nonassoc
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprCallConfig describes function-call syntax inside a Pratt expression.
type ExprCallConfig struct {
	Open   string `yaml:"open" json:"open"`
	Close  string `yaml:"close" json:"close"`
	Sep    string `yaml:"sep,omitempty" json:"sep,omitempty"`
	Power  int    `yaml:"power" json:"power"`
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprIndexConfig describes subscript syntax inside a Pratt expression.
type ExprIndexConfig struct {
	Open   string `yaml:"open" json:"open"`
	Inner  string `yaml:"inner,omitempty" json:"inner,omitempty"`
	Close  string `yaml:"close" json:"close"`
	Power  int    `yaml:"power" json:"power"`
	Action string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprMemberConf describes member-access syntax inside a Pratt expression.
type ExprMemberConf struct {
	DotToken   string `yaml:"dot_token" json:"dot_token"`
	FieldToken string `yaml:"field_token" json:"field_token"`
	Power      int    `yaml:"power" json:"power"`
	Action     string `yaml:"action,omitempty" json:"action,omitempty"`
}

// ExprTernaryConf describes ternary conditional syntax inside a Pratt expression.
type ExprTernaryConf struct {
	QuestionToken string `yaml:"question_token" json:"question_token"`
	ColonToken    string `yaml:"colon_token" json:"colon_token"`
	Power         int    `yaml:"power" json:"power"`
	Action        string `yaml:"action,omitempty" json:"action,omitempty"`
}

// LoadFromYAML creates a DSL from a YAML configuration.
func LoadFromYAML(yamlData []byte) (*DSL, error) {
	var config DSLConfig
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return createDSLFromConfig(config)
}

// LoadFromYAMLFile creates a DSL from a YAML file.
func LoadFromYAMLFile(filename string) (*DSL, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	return LoadFromYAML(data)
}

// LoadFromJSON creates a DSL from a JSON configuration.
func LoadFromJSON(jsonData []byte) (*DSL, error) {
	var config DSLConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return createDSLFromConfig(config)
}

// LoadFromJSONFile creates a DSL from a JSON file.
func LoadFromJSONFile(filename string) (*DSL, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}
	return LoadFromJSON(data)
}

// CreateFromConfig builds a *DSL from a parsed DSLConfig.
func CreateFromConfig(config DSLConfig) (*DSL, error) {
	return createDSLFromConfig(config)
}

func createDSLFromConfig(config DSLConfig) (*DSL, error) {
	dsl := New(config.Name)
	dsl.SetSignificantNewlines(config.SignificantNewlines)

	if len(config.TokenDefs) > 0 {
		// v2 lossless ordered token definitions.
		for _, td := range config.TokenDefs {
			switch strings.ToLower(strings.TrimSpace(td.Kind)) {
			case "keyword":
				if err := dsl.KeywordToken(td.Name, td.Pattern); err != nil {
					return nil, fmt.Errorf("failed to add keyword token %s: %w", td.Name, err)
				}
			case "literal":
				if err := dsl.LiteralToken(td.Name, td.Pattern); err != nil {
					return nil, fmt.Errorf("failed to add literal token %s: %w", td.Name, err)
				}
			case "skip":
				if err := dsl.SkipToken(td.Name, td.Pattern); err != nil {
					return nil, fmt.Errorf("failed to add skip token %s: %w", td.Name, err)
				}
			case "lookaround":
				if err := dsl.TokenWithLookaround(td.Name, td.Pattern, td.Lookahead, td.Lookbehind); err != nil {
					return nil, fmt.Errorf("failed to add lookaround token %s: %w", td.Name, err)
				}
			default:
				if err := dsl.Token(td.Name, td.Pattern); err != nil {
					return nil, fmt.Errorf("failed to add token %s: %w", td.Name, err)
				}
			}
		}
	} else {
		// Legacy v1 map: sort keys so tie-breaking is deterministic across runs.
		names := make([]string, 0, len(config.Tokens))
		for name := range config.Tokens {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			pattern := config.Tokens[name]
			if isKeywordTokenPattern(pattern) {
				keyword := extractKeywordFromPattern(pattern)
				if err := dsl.KeywordToken(name, keyword); err != nil {
					return nil, fmt.Errorf("failed to add keyword token %s: %w", name, err)
				}
			} else if isKeywordToken(pattern) {
				if err := dsl.KeywordToken(name, pattern); err != nil {
					return nil, fmt.Errorf("failed to add keyword token %s: %w", name, err)
				}
			} else {
				if err := dsl.Token(name, pattern); err != nil {
					return nil, fmt.Errorf("failed to add token %s: %w", name, err)
				}
			}
		}
	}

	// Add rules in order.
	for _, rule := range config.Rules {
		if rule.ErrorHint != "" {
			dsl.RuleWithError(rule.Name, rule.Pattern, rule.Action, rule.ErrorHint)
		} else if rule.Precedence != 0 {
			dsl.RuleWithPrecedence(rule.Name, rule.Pattern, rule.Action, rule.Precedence, rule.Associativity)
		} else {
			dsl.Rule(rule.Name, rule.Pattern, rule.Action)
		}
	}

	// Add Pratt expression rules.
	for _, expr := range config.Expressions {
		eb := dsl.Expression(expr.Name)
		for _, a := range expr.Atoms {
			eb.Atom(a.Token, a.Action)
		}
		for _, g := range expr.Groups {
			if g.Action != "" {
				eb.GroupWithAction(g.Open, g.Inner, g.Close, g.Action)
			} else {
				eb.Group(g.Open, g.Inner, g.Close)
			}
		}
		for _, op := range expr.Prefix {
			eb.Prefix(op.Token, op.Power, op.Action)
		}
		for _, op := range expr.Postfix {
			eb.Postfix(op.Token, op.Power, op.Action)
		}
		for _, op := range expr.Infix {
			switch strings.ToLower(strings.TrimSpace(op.Assoc)) {
			case "right":
				eb.InfixRight(op.Token, op.Power, op.Action)
			case "nonassoc", "none":
				eb.InfixNonAssoc(op.Token, op.Power, op.Action)
			default:
				eb.InfixLeft(op.Token, op.Power, op.Action)
			}
		}
		for _, c := range expr.Calls {
			eb.Call(c.Open, c.Close, c.Sep, c.Power, c.Action)
		}
		for _, idx := range expr.Indices {
			eb.Index(idx.Open, idx.Inner, idx.Close, idx.Power, idx.Action)
		}
		for _, m := range expr.Members {
			eb.Member(m.DotToken, m.FieldToken, m.Power, m.Action)
		}
		if expr.Ternary != nil {
			eb.Ternary(expr.Ternary.QuestionToken, expr.Ternary.ColonToken, expr.Ternary.Power, expr.Ternary.Action)
		}
	}

	// Restore explicit start rule if configured.
	if config.StartRule != "" {
		dsl.SetStartRule(config.StartRule)
	}

	// Restore formatting layouts.
	for ruleName, layout := range config.Layouts {
		dsl.SetRuleLayout(ruleName, layout)
	}

	// Restore context defaults.
	for key, value := range config.Context {
		dsl.SetContext(key, value)
	}

	// Store declarative test suite if present.
	if len(config.Tests) > 0 {
		dsl.tests = append([]DSLTestCase(nil), config.Tests...)
	}

	return dsl, nil
}

// isKeywordToken checks if a legacy v1 pattern is likely a keyword (simple word without regex).
func isKeywordToken(pattern string) bool {
	for _, r := range pattern {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	hasLetter := false
	for _, r := range pattern {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
			break
		}
	}
	return hasLetter && len(pattern) > 0
}

func isKeywordTokenPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "(?i)\\b") && strings.HasSuffix(pattern, "\\b")
}

func extractKeywordFromPattern(pattern string) string {
	if isKeywordTokenPattern(pattern) {
		pattern = strings.TrimPrefix(pattern, "(?i)\\b")
		pattern = strings.TrimSuffix(pattern, "\\b")
		return pattern
	}
	return pattern
}

// SaveToYAML exports the DSL configuration to YAML format.
func (d *DSL) SaveToYAML() ([]byte, error) {
	config := d.toConfig()
	return yaml.Marshal(config)
}

// SaveToYAMLFile exports the DSL configuration to a YAML file.
func (d *DSL) SaveToYAMLFile(filename string) error {
	data, err := d.SaveToYAML()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// SaveToJSON exports the DSL configuration to indented JSON format.
func (d *DSL) SaveToJSON() ([]byte, error) {
	config := d.toConfig()
	return json.MarshalIndent(config, "", "  ")
}

// SaveToJSONFile exports the DSL configuration to a JSON file.
func (d *DSL) SaveToJSONFile(filename string) error {
	data, err := d.SaveToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ToConfig exports the DSL configuration as a serializable DSLConfig struct.
func (d *DSL) ToConfig() DSLConfig {
	return d.toConfig()
}

func (d *DSL) toConfig() DSLConfig {
	d.mu.RLock()
	ctxCopy := make(map[string]interface{}, len(d.context))
	for k, v := range d.context {
		ctxCopy[k] = v
	}
	testsCopy := append([]DSLTestCase(nil), d.tests...)
	d.mu.RUnlock()

	config := DSLConfig{
		Name:                d.name,
		StartRule:           d.grammar.startRule,
		SignificantNewlines: d.grammar.significantNewlines,
		Tokens:              make(map[string]string),
		Rules:               []RuleConfig{},
		Context:             ctxCopy,
		Tests:               testsCopy,
	}

	// Export tokens in exact declaration order for lossless v2, while keeping
	// legacy config.Tokens populated for v1 readers.
	for _, tok := range d.grammar.tokenList {
		if !tok.skip {
			config.Tokens[tok.name] = tok.pattern
		}
		kind := string(tok.kind)
		if kind == "" {
			kind = string(tokenKindRegex)
		}
		pat := tok.pattern
		if (tok.kind == tokenKindKeyword || tok.kind == tokenKindLiteral) && tok.rawText != "" {
			pat = tok.rawText
		}
		config.TokenDefs = append(config.TokenDefs, TokenConfig{
			Name:       tok.name,
			Kind:       kind,
			Pattern:    pat,
			Lookahead:  tok.lookahead,
			Lookbehind: tok.lookbehind,
		})
	}

	// Export rules in declaration order.
	for _, name := range d.grammar.ruleOrder {
		rule := d.grammar.rules[name]
		for _, alt := range rule.alternatives {
			rc := RuleConfig{
				Name:      name,
				Pattern:   alt.sequence,
				Action:    alt.action,
				ErrorHint: alt.errHint,
			}
			if alt.precedence != 0 {
				rc.Precedence = alt.precedence
				rc.Associativity = alt.associativity
			}
			config.Rules = append(config.Rules, rc)
		}
	}

	// Export Pratt expression rules deterministically.
	for _, exprName := range sortedExprRuleNames(d.grammar) {
		spec := d.grammar.exprRules[exprName]
		ec := ExpressionConfig{Name: exprName}
		for _, a := range spec.atoms {
			ec.Atoms = append(ec.Atoms, ExprAtomConfig{Token: a.token, Action: a.action})
		}
		for _, g := range spec.groups {
			ec.Groups = append(ec.Groups, ExprGroupConfig{Open: g.open, Inner: g.inner, Close: g.close, Action: g.action})
		}
		for _, tok := range sortedMapKeys(spec.prefix) {
			op := spec.prefix[tok]
			ec.Prefix = append(ec.Prefix, ExprOpConfig{Token: op.token, Power: op.power, Action: op.action})
		}
		for _, tok := range sortedMapKeys(spec.postfix) {
			op := spec.postfix[tok]
			ec.Postfix = append(ec.Postfix, ExprOpConfig{Token: op.token, Power: op.power, Action: op.action})
		}
		for _, tok := range sortedMapKeys(spec.infix) {
			op := spec.infix[tok]
			assoc := "left"
			if op.rightAssoc {
				assoc = "right"
			} else if op.nonAssoc {
				assoc = "nonassoc"
			}
			ec.Infix = append(ec.Infix, ExprOpConfig{Token: op.token, Power: op.power, Assoc: assoc, Action: op.action})
		}
		for _, tok := range sortedCallKeys(spec.calls) {
			c := spec.calls[tok]
			ec.Calls = append(ec.Calls, ExprCallConfig{Open: c.open, Close: c.close, Sep: c.sep, Power: c.power, Action: c.action})
		}
		for _, tok := range sortedIndexKeys(spec.indices) {
			idx := spec.indices[tok]
			ec.Indices = append(ec.Indices, ExprIndexConfig{Open: idx.open, Inner: idx.inner, Close: idx.close, Power: idx.power, Action: idx.action})
		}
		for _, tok := range sortedMemberKeys(spec.members) {
			m := spec.members[tok]
			ec.Members = append(ec.Members, ExprMemberConf{DotToken: m.dotToken, FieldToken: m.fieldToken, Power: m.power, Action: m.action})
		}
		if spec.ternary != nil {
			ec.Ternary = &ExprTernaryConf{
				QuestionToken: spec.ternary.questionToken,
				ColonToken:    spec.ternary.colonToken,
				Power:         spec.ternary.power,
				Action:        spec.ternary.action,
			}
		}
		config.Expressions = append(config.Expressions, ec)
	}

	if len(d.grammar.layouts) > 0 {
		config.Layouts = make(map[string]RuleLayout, len(d.grammar.layouts))
		for k, v := range d.grammar.layouts {
			config.Layouts[k] = v
		}
	}

	return config
}
