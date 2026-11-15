package main

// StudioPreset represents a curated starter DSL grammar + sample program for DSL Studio.
type StudioPreset struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Badge       string `json:"badge"`
	GrammarYAML string `json:"grammar_yaml"`
	SampleCode  string `json:"sample_code"`
}

func builtinPresets() []StudioPreset {
	return []StudioPreset{
		{
			ID:       "fintech-pricing",
			Title:    "Fintech Pricing & Risk",
			Subtitle: "Expresiones Pratt con ternario (? :), comparaciones, comentarios y reglas de negocio",
			Badge:    "Pratt + Ternary",
			GrammarYAML: `name: FintechPricing
start_rule: rule_stmt
token_defs:
  - name: COMMENT
    kind: skip
    pattern: "//[^\\r\\n]*"
  - name: WHEN
    kind: keyword
    pattern: when
  - name: APPLY
    kind: keyword
    pattern: apply
  - name: GTE
    kind: literal
    pattern: ">="
  - name: LTE
    kind: literal
    pattern: "<="
  - name: EQ
    kind: literal
    pattern: "=="
  - name: GT
    kind: literal
    pattern: ">"
  - name: LT
    kind: literal
    pattern: "<"
  - name: QMARK
    kind: literal
    pattern: "?"
  - name: COLON
    kind: literal
    pattern: ":"
  - name: PLUS
    kind: literal
    pattern: "+"
  - name: MINUS
    kind: literal
    pattern: "-"
  - name: STAR
    kind: literal
    pattern: "*"
  - name: SLASH
    kind: literal
    pattern: "/"
  - name: LPAREN
    kind: literal
    pattern: "("
  - name: RPAREN
    kind: literal
    pattern: ")"
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+(\\.[0-9]+)?"
  - name: IDENT
    kind: regex
    pattern: "[a-zA-Z_][a-zA-Z0-9_]*"
rules:
  - name: rule_stmt
    pattern: [WHEN, expr, APPLY, expr]
    action: pricing_rule
    error_hint: "Usa el formato: when <condicion> apply <formula>"
expressions:
  - name: expr
    atoms:
      - token: NUMBER
        action: number
      - token: IDENT
        action: ident
    groups:
      - open: LPAREN
        inner: expr
        close: RPAREN
        action: paren
    prefix:
      - token: MINUS
        power: 35
        action: neg
    infix:
      - token: STAR
        power: 30
        assoc: left
        action: mul
      - token: SLASH
        power: 30
        assoc: left
        action: div
      - token: PLUS
        power: 20
        assoc: left
        action: add
      - token: MINUS
        power: 20
        assoc: left
        action: sub
      - token: GTE
        power: 10
        assoc: nonassoc
        action: gte
      - token: LTE
        power: 10
        assoc: nonassoc
        action: lte
      - token: GT
        power: 10
        assoc: nonassoc
        action: gt
      - token: LT
        power: 10
        assoc: nonassoc
        action: lt
      - token: EQ
        power: 10
        assoc: nonassoc
        action: eq
    ternary:
      question_token: QMARK
      colon_token: COLON
      power: 5
      action: ternary
tests:
  - name: regla vip con ternario
    input: "when 1200 >= 1000 apply (1200 >= 1000 ? 1200 * 0.85 : 1200)"
    should_parse: true
  - name: impide encadenar comparadores no asociativos
    input: "when 10 < 20 < 30 apply 100"
    should_parse: false
`,
			SampleCode: `// Regla dinámica de pricing con evaluación determinista
when 1200 >= 1000 apply (1200 >= 1000 ? 1200 * 0.85 : 1200)`,
		},
		{
			ID:       "smart-query",
			Title:    "Smart Filter & Query",
			Subtitle: "Consultas declarativas tipo SQL/SCIM con precedencia booleana (AND/OR)",
			Badge:    "Query Engine",
			GrammarYAML: `name: SmartQuery
start_rule: query
token_defs:
  - name: COMMENT
    kind: skip
    pattern: "#[^\\r\\n]*"
  - name: WHERE
    kind: keyword
    pattern: where
  - name: AND
    kind: keyword
    pattern: and
  - name: OR
    kind: keyword
    pattern: or
  - name: EQ
    kind: literal
    pattern: "=="
  - name: NEQ
    kind: literal
    pattern: "!="
  - name: GT
    kind: literal
    pattern: ">"
  - name: LT
    kind: literal
    pattern: "<"
  - name: LPAREN
    kind: literal
    pattern: "("
  - name: RPAREN
    kind: literal
    pattern: ")"
  - name: STRING
    kind: regex
    pattern: "\"[^\"]*\""
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+(\\.[0-9]+)?"
  - name: IDENT
    kind: regex
    pattern: "[a-zA-Z_][a-zA-Z0-9_]*"
rules:
  - name: query
    pattern: [WHERE, filter_expr]
    action: run_query
expressions:
  - name: filter_expr
    atoms:
      - token: NUMBER
        action: number
      - token: STRING
        action: string
      - token: IDENT
        action: ident
    groups:
      - open: LPAREN
        inner: filter_expr
        close: RPAREN
        action: paren
    infix:
      - token: EQ
        power: 20
        assoc: nonassoc
        action: eq
      - token: NEQ
        power: 20
        assoc: nonassoc
        action: neq
      - token: GT
        power: 20
        assoc: nonassoc
        action: gt
      - token: LT
        power: 20
        assoc: nonassoc
        action: lt
      - token: AND
        power: 10
        assoc: left
        action: and
      - token: OR
        power: 5
        assoc: left
        action: or
tests:
  - name: filtro compuesto valido
    input: 'where ("active" == "active" and 450 > 100) or 0 == 1'
    should_parse: true
  - name: sintaxis incompleta
    input: "where status =="
    should_parse: false
`,
			SampleCode: `# Consulta evaluada con precedencia formal AND > OR
where ("active" == "active" and 450 > 100) or 0 == 1`,
		},
		{
			ID:       "workflow-ops",
			Title:    "Automation & Calls DSL",
			Subtitle: "Acceso a propiedades (obj.prop), invocación de funciones fn(a, b) y ruteo",
			Badge:    "Call + Member",
			GrammarYAML: `name: WorkflowOps
start_rule: trigger
token_defs:
  - name: COMMENT
    kind: skip
    pattern: "//[^\\r\\n]*"
  - name: ON
    kind: keyword
    pattern: on
  - name: DO
    kind: keyword
    pattern: do
  - name: DOT
    kind: literal
    pattern: "."
  - name: COMMA
    kind: literal
    pattern: ","
  - name: LPAREN
    kind: literal
    pattern: "("
  - name: RPAREN
    kind: literal
    pattern: ")"
  - name: GT
    kind: literal
    pattern: ">"
  - name: STRING
    kind: regex
    pattern: "\"[^\"]*\""
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+"
  - name: IDENT
    kind: regex
    pattern: "[a-zA-Z_][a-zA-Z0-9_]*"
rules:
  - name: trigger
    pattern: [ON, expr, DO, expr]
    action: emit_workflow
expressions:
  - name: expr
    atoms:
      - token: IDENT
        action: ident
      - token: NUMBER
        action: number
      - token: STRING
        action: string
    groups:
      - open: LPAREN
        inner: expr
        close: RPAREN
        action: paren
    infix:
      - token: GT
        power: 10
        assoc: nonassoc
        action: gt
    calls:
      - open: LPAREN
        close: RPAREN
        sep: COMMA
        power: 40
        action: call_fn
    members:
      - dot_token: DOT
        field_token: IDENT
        power: 50
        action: get_member
tests:
  - name: alerta antifraude
    input: 'on order.risk_score > 80 do notify("secops", order.id)'
    should_parse: true
`,
			SampleCode: `// Dispara acciones sobre objetos anidados e invocaciones tipadas
on order.risk_score > 80 do notify("secops", order.id)`,
		},
		{
			ID:       "calc-pratt",
			Title:    "Scientific Pratt Lab",
			Subtitle: "Potencia asociativa derecha (^), factorial postfijo (!), unario (-) y paréntesis",
			Badge:    "Math Core",
			GrammarYAML: `name: ScientificCalc
start_rule: expr
token_defs:
  - name: COMMENT
    kind: skip
    pattern: "//[^\\r\\n]*"
  - name: PLUS
    kind: literal
    pattern: "+"
  - name: MINUS
    kind: literal
    pattern: "-"
  - name: STAR
    kind: literal
    pattern: "*"
  - name: SLASH
    kind: literal
    pattern: "/"
  - name: CARET
    kind: literal
    pattern: "^"
  - name: BANG
    kind: literal
    pattern: "!"
  - name: LPAREN
    kind: literal
    pattern: "("
  - name: RPAREN
    kind: literal
    pattern: ")"
  - name: NUMBER
    kind: regex
    pattern: "[0-9]+(\\.[0-9]+)?"
expressions:
  - name: expr
    atoms:
      - token: NUMBER
        action: number
    groups:
      - open: LPAREN
        inner: expr
        close: RPAREN
        action: paren
    prefix:
      - token: MINUS
        power: 35
        action: neg
    postfix:
      - token: BANG
        power: 45
        action: fact
    infix:
      - token: PLUS
        power: 10
        assoc: left
        action: add
      - token: MINUS
        power: 10
        assoc: left
        action: sub
      - token: STAR
        power: 20
        assoc: left
        action: mul
      - token: SLASH
        power: 20
        assoc: left
        action: div
      - token: CARET
        power: 30
        assoc: right
        action: pow
tests:
  - name: factorial y potencia derecha
    input: "5! + 2 ^ 3 ^ 2"
    expect_output: 632
  - name: division entre cero controlada
    input: "10 + +"
    should_parse: false
`,
			SampleCode: `// 5! (120) + 2^(3^2) (512) = 632
5! + 2 ^ 3 ^ 2 // prueba cambiar valores en vivo`,
		},
	}
}
