package universal

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// UniversalLiveViewDSL creates a universal LiveView HTML DSL that works with any struct type
type UniversalLiveViewDSL struct {
	dsl       *dslbuilder.DSL
	generator *UniversalHTMLGenerator
}

// NewUniversalLiveViewDSL creates a new universal LiveView DSL
func NewUniversalLiveViewDSL() *UniversalLiveViewDSL {
	dsl := dslbuilder.New("UniversalLiveViewDSL")
	generator := NewHTMLGenerator()

	ul := &UniversalLiveViewDSL{
		dsl:       dsl,
		generator: generator,
	}

	ul.setupTokens()
	ul.setupRules()
	ul.setupActions()

	return ul
}

// setupTokens defines all the tokens for the LiveView DSL
func (ul *UniversalLiveViewDSL) setupTokens() {
	// HTML Generation commands (Spanish)
	ul.dsl.KeywordToken("GENERAR", "generar")
	ul.dsl.KeywordToken("CREAR", "crear")
	ul.dsl.KeywordToken("MOSTRAR", "mostrar")
	ul.dsl.KeywordToken("LISTAR", "listar")
	ul.dsl.KeywordToken("FORMULARIO", "formulario")
	ul.dsl.KeywordToken("TABLA", "tabla")
	ul.dsl.KeywordToken("TARJETA", "tarjeta")
	ul.dsl.KeywordToken("BOTON", "boton")
	ul.dsl.KeywordToken("MODAL", "modal")
	ul.dsl.KeywordToken("PAGINA", "pagina")

	// HTML Generation commands (English)
	ul.dsl.KeywordToken("GENERATE", "generate")
	ul.dsl.KeywordToken("CREATE", "create")
	ul.dsl.KeywordToken("SHOW", "show")
	ul.dsl.KeywordToken("LIST", "list")
	ul.dsl.KeywordToken("FORM", "form")
	ul.dsl.KeywordToken("TABLE", "table")
	ul.dsl.KeywordToken("CARD", "card")
	ul.dsl.KeywordToken("BUTTON", "button")
	ul.dsl.KeywordToken("MODAL", "modal")
	ul.dsl.KeywordToken("PAGE", "page")

	// Component types
	ul.dsl.KeywordToken("INPUT", "input")
	ul.dsl.KeywordToken("SELECT", "select")
	ul.dsl.KeywordToken("TEXTAREA", "textarea")
	ul.dsl.KeywordToken("CHECKBOX", "checkbox")
	ul.dsl.KeywordToken("RADIO", "radio")

	// Actions and events (Spanish)
	ul.dsl.KeywordToken("CON", "con")           // with
	ul.dsl.KeywordToken("PARA", "para")         // for
	ul.dsl.KeywordToken("USANDO", "usando")     // using
	ul.dsl.KeywordToken("ACCION", "accion")     // action
	ul.dsl.KeywordToken("EVENTO", "evento")     // event
	ul.dsl.KeywordToken("CLASE", "clase")       // class
	ul.dsl.KeywordToken("ESTILO", "estilo")     // style
	ul.dsl.KeywordToken("PLANTILLA", "plantilla") // template

	// Actions and events (English)
	ul.dsl.KeywordToken("WITH", "with")
	ul.dsl.KeywordToken("FOR", "for")
	ul.dsl.KeywordToken("USING", "using")
	ul.dsl.KeywordToken("ACTION", "action")
	ul.dsl.KeywordToken("EVENT", "event")
	ul.dsl.KeywordToken("CLASS", "class")
	ul.dsl.KeywordToken("STYLE", "style")
	ul.dsl.KeywordToken("TEMPLATE", "template")

	// LiveView specific events
	ul.dsl.KeywordToken("PHX_CLICK", "phx-click")
	ul.dsl.KeywordToken("PHX_SUBMIT", "phx-submit")
	ul.dsl.KeywordToken("PHX_CHANGE", "phx-change")
	ul.dsl.KeywordToken("PHX_BLUR", "phx-blur")
	ul.dsl.KeywordToken("PHX_FOCUS", "phx-focus")
	ul.dsl.KeywordToken("PHX_KEYUP", "phx-keyup")
	ul.dsl.KeywordToken("PHX_UPDATE", "phx-update")

	// Value tokens
	ul.dsl.Token("STRING", `"[^"]*"`)
	ul.dsl.Token("NUMBER", `[0-9]+\.?[0-9]*`)
	ul.dsl.Token("WORD", `[a-zA-Z][a-zA-Z0-9_]*`)
}

// setupRules defines grammar rules for LiveView HTML generation
func (ul *UniversalLiveViewDSL) setupRules() {
	// Form generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "FORMULARIO", "PARA", "WORD"}, "generateForm")
	ul.dsl.Rule("html_gen", []string{"CREAR", "FORMULARIO", "PARA", "WORD"}, "generateForm")
	ul.dsl.Rule("html_gen", []string{"GENERAR", "FORMULARIO", "PARA", "WORD", "CON", "ACCION", "STRING"}, "generateFormWithAction")

	// Form generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "FORM", "FOR", "WORD"}, "generateForm")
	ul.dsl.Rule("html_gen", []string{"CREATE", "FORM", "FOR", "WORD"}, "generateForm")
	ul.dsl.Rule("html_gen", []string{"GENERATE", "FORM", "FOR", "WORD", "WITH", "ACTION", "STRING"}, "generateFormWithAction")

	// Table generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "TABLA", "PARA", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"CREAR", "TABLA", "PARA", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"MOSTRAR", "TABLA", "DE", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"LISTAR", "TABLA", "DE", "WORD"}, "generateTable")

	// Table generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "TABLE", "FOR", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"CREATE", "TABLE", "FOR", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"SHOW", "TABLE", "OF", "WORD"}, "generateTable")
	ul.dsl.Rule("html_gen", []string{"LIST", "TABLE", "OF", "WORD"}, "generateTable")

	// Card generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "TARJETA", "PARA", "WORD"}, "generateCard")
	ul.dsl.Rule("html_gen", []string{"CREAR", "TARJETA", "PARA", "WORD"}, "generateCard")
	ul.dsl.Rule("html_gen", []string{"MOSTRAR", "TARJETA", "DE", "WORD"}, "generateCard")

	// Card generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "CARD", "FOR", "WORD"}, "generateCard")
	ul.dsl.Rule("html_gen", []string{"CREATE", "CARD", "FOR", "WORD"}, "generateCard")
	ul.dsl.Rule("html_gen", []string{"SHOW", "CARD", "OF", "WORD"}, "generateCard")

	// Button generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "BOTON", "CON", "TEXTO", "STRING"}, "generateButton")
	ul.dsl.Rule("html_gen", []string{"CREAR", "BOTON", "CON", "TEXTO", "STRING"}, "generateButton")
	ul.dsl.Rule("html_gen", []string{"GENERAR", "BOTON", "CON", "TEXTO", "STRING", "Y", "ACCION", "STRING"}, "generateButtonWithAction")

	// Button generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "BUTTON", "WITH", "TEXT", "STRING"}, "generateButton")
	ul.dsl.Rule("html_gen", []string{"CREATE", "BUTTON", "WITH", "TEXT", "STRING"}, "generateButton")
	ul.dsl.Rule("html_gen", []string{"GENERATE", "BUTTON", "WITH", "TEXT", "STRING", "AND", "ACTION", "STRING"}, "generateButtonWithAction")

	// Modal generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "MODAL", "CON", "TITULO", "STRING"}, "generateModal")
	ul.dsl.Rule("html_gen", []string{"CREAR", "MODAL", "CON", "TITULO", "STRING"}, "generateModal")

	// Modal generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "MODAL", "WITH", "TITLE", "STRING"}, "generateModal")
	ul.dsl.Rule("html_gen", []string{"CREATE", "MODAL", "WITH", "TITLE", "STRING"}, "generateModal")

	// Page generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "PAGINA", "CON", "PLANTILLA", "STRING"}, "generatePage")
	ul.dsl.Rule("html_gen", []string{"CREAR", "PAGINA", "CON", "PLANTILLA", "STRING"}, "generatePage")

	// Page generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "PAGE", "WITH", "TEMPLATE", "STRING"}, "generatePage")
	ul.dsl.Rule("html_gen", []string{"CREATE", "PAGE", "WITH", "TEMPLATE", "STRING"}, "generatePage")

	// List generation (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "LISTA", "DE", "WORD"}, "generateList")
	ul.dsl.Rule("html_gen", []string{"CREAR", "LISTA", "DE", "WORD"}, "generateList")
	ul.dsl.Rule("html_gen", []string{"MOSTRAR", "LISTA", "DE", "WORD"}, "generateList")

	// List generation (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "LIST", "OF", "WORD"}, "generateList")
	ul.dsl.Rule("html_gen", []string{"CREATE", "LIST", "OF", "WORD"}, "generateList")
	ul.dsl.Rule("html_gen", []string{"SHOW", "LIST", "OF", "WORD"}, "generateList")

	// Component with classes (Spanish)
	ul.dsl.Rule("html_gen", []string{"GENERAR", "WORD", "CON", "CLASSE", "STRING"}, "generateComponentWithClass")
	ul.dsl.Rule("html_gen", []string{"CREAR", "WORD", "CON", "CLASSE", "STRING"}, "generateComponentWithClass")

	// Component with classes (English)
	ul.dsl.Rule("html_gen", []string{"GENERATE", "WORD", "WITH", "CLASS", "STRING"}, "generateComponentWithClass")
	ul.dsl.Rule("html_gen", []string{"CREATE", "WORD", "WITH", "CLASS", "STRING"}, "generateComponentWithClass")

	// Add missing tokens
	ul.dsl.KeywordToken("DE", "de")
	ul.dsl.KeywordToken("OF", "of")
	ul.dsl.KeywordToken("TEXTO", "texto")
	ul.dsl.KeywordToken("TEXT", "text")
	ul.dsl.KeywordToken("Y", "y")
	ul.dsl.KeywordToken("AND", "and")
	ul.dsl.KeywordToken("TITULO", "titulo")
	ul.dsl.KeywordToken("TITLE", "title")
	ul.dsl.KeywordToken("LISTA", "lista")
	ul.dsl.KeywordToken("CLASSE", "classe")
}

// setupActions defines all the action handlers
func (ul *UniversalLiveViewDSL) setupActions() {
	// Generate form action
	ul.dsl.Action("generateForm", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for form generation")
		}

		entityName := args[3].(string)
		return ul.executeFormGeneration(entityName, "submit_form")
	})

	// Generate form with action
	ul.dsl.Action("generateFormWithAction", func(args []interface{}) (interface{}, error) {
		if len(args) < 7 {
			return nil, fmt.Errorf("insufficient arguments for form generation with action")
		}

		entityName := args[3].(string)
		action := strings.Trim(args[6].(string), `"`)
		return ul.executeFormGeneration(entityName, action)
	})

	// Generate table action
	ul.dsl.Action("generateTable", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for table generation")
		}

		entityName := args[3].(string)
		return ul.executeTableGeneration(entityName)
	})

	// Generate card action
	ul.dsl.Action("generateCard", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for card generation")
		}

		entityName := args[3].(string)
		return ul.executeCardGeneration(entityName)
	})

	// Generate button action
	ul.dsl.Action("generateButton", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for button generation")
		}

		text := strings.Trim(args[4].(string), `"`)
		return ul.executeButtonGeneration(text, "button_click")
	})

	// Generate button with action
	ul.dsl.Action("generateButtonWithAction", func(args []interface{}) (interface{}, error) {
		if len(args) < 8 {
			return nil, fmt.Errorf("insufficient arguments for button generation with action")
		}

		text := strings.Trim(args[4].(string), `"`)
		action := strings.Trim(args[7].(string), `"`)
		return ul.executeButtonGeneration(text, action)
	})

	// Generate modal action
	ul.dsl.Action("generateModal", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for modal generation")
		}

		title := strings.Trim(args[4].(string), `"`)
		return ul.executeModalGeneration(title)
	})

	// Generate page action
	ul.dsl.Action("generatePage", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for page generation")
		}

		template := strings.Trim(args[4].(string), `"`)
		return ul.executePageGeneration(template)
	})

	// Generate list action
	ul.dsl.Action("generateList", func(args []interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("insufficient arguments for list generation")
		}

		entityName := args[3].(string)
		return ul.executeListGeneration(entityName)
	})

	// Generate component with class
	ul.dsl.Action("generateComponentWithClass", func(args []interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("insufficient arguments for component generation with class")
		}

		componentType := args[1].(string)
		classes := strings.Trim(args[4].(string), `"`)
		return ul.executeComponentGeneration(componentType, classes)
	})
}

// Execute generation methods

func (ul *UniversalLiveViewDSL) executeFormGeneration(entityName, action string) (interface{}, error) {
	// Get entity data from context
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		// Create a mock entity for demonstration
		data = ul.createMockEntity(entityName)
	} else {
		// If data is a slice, take the first element for form generation
		if ul.isSlice(data) {
			sliceData := ul.convertToInterfaceSlice(data)
			if len(sliceData) > 0 {
				data = sliceData[0]
			} else {
				data = ul.createMockEntity(entityName)
			}
		}
	}

	html, err := ul.generator.GenerateForm(data, action)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated form for %s:\n%s", entityName, html), nil
}

func (ul *UniversalLiveViewDSL) executeTableGeneration(entityName string) (interface{}, error) {
	// Get entity data from context
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		// Create mock data for demonstration
		data = ul.createMockEntitySlice(entityName)
	}

	options := map[string]interface{}{
		"sortable":   true,
		"filterable": true,
		"paginated":  true,
	}

	html, err := ul.generator.GenerateTable(data, options)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated table for %s:\n%s", entityName, html), nil
}

func (ul *UniversalLiveViewDSL) executeCardGeneration(entityName string) (interface{}, error) {
	// Get entity data from context
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		// Create a mock entity for demonstration
		data = ul.createMockEntity(entityName)
	} else {
		// If data is a slice, take the first element for card generation
		if ul.isSlice(data) {
			sliceData := ul.convertToInterfaceSlice(data)
			if len(sliceData) > 0 {
				data = sliceData[0]
			} else {
				data = ul.createMockEntity(entityName)
			}
		}
	}

	html, err := ul.generator.GenerateCard(data, "")
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated card for %s:\n%s", entityName, html), nil
}

func (ul *UniversalLiveViewDSL) executeButtonGeneration(text, action string) (interface{}, error) {
	buttonData := map[string]interface{}{
		"type":    "button",
		"content": text,
		"classes": "btn btn-primary",
	}

	html, err := ul.generator.GenerateComponent("button", buttonData)
	if err != nil {
		return nil, err
	}

	// Add LiveView event
	html = strings.ReplaceAll(html, "{{.events}}", fmt.Sprintf(`phx-click="%s"`, action))

	return fmt.Sprintf("Generated button:\n%s", html), nil
}

func (ul *UniversalLiveViewDSL) executeModalGeneration(title string) (interface{}, error) {
	modalData := map[string]interface{}{
		"title":   title,
		"content": "<p>Modal content goes here</p>",
		"classes": "modal-overlay",
	}

	html, err := ul.generator.GenerateComponent("modal", modalData)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated modal:\n%s", html), nil
}

func (ul *UniversalLiveViewDSL) executePageGeneration(template string) (interface{}, error) {
	pageData := map[string]interface{}{
		"title":      "Generated Page",
		"content":    "<div>Page content goes here</div>",
		"csrf_token": "csrf-token-placeholder",
	}

	html, err := ul.generator.GeneratePage(template, pageData)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated page with template '%s':\n%s", template, html), nil
}

func (ul *UniversalLiveViewDSL) executeListGeneration(entityName string) (interface{}, error) {
	// Get entity data from context
	data := ul.dsl.GetContext(entityName)
	if data == nil {
		// Create mock data for demonstration
		data = ul.createMockEntitySlice(entityName)
	}

	itemTemplate := `
	<div class="list-item-content">
		<h4>{{.title}}</h4>
		<p>{{.content}}</p>
	</div>`

	html, err := ul.generator.GenerateList(data, itemTemplate)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated list for %s:\n%s", entityName, html), nil
}

func (ul *UniversalLiveViewDSL) executeComponentGeneration(componentType, classes string) (interface{}, error) {
	componentData := map[string]interface{}{
		"classes": classes,
		"content": fmt.Sprintf("%s content", componentType),
	}

	html, err := ul.generator.GenerateComponent(componentType, componentData)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("Generated %s component:\n%s", componentType, html), nil
}

// Helper methods

func (ul *UniversalLiveViewDSL) isSlice(data interface{}) bool {
	return reflect.ValueOf(data).Kind() == reflect.Slice
}

func (ul *UniversalLiveViewDSL) convertToInterfaceSlice(slice interface{}) []interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil
	}

	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}

	return result
}

// Helper methods for creating mock data

func (ul *UniversalLiveViewDSL) createMockEntity(entityName string) interface{} {
	switch strings.ToLower(entityName) {
	case "user", "usuario":
		return &struct {
			ID    int    `html:"id"`
			Name  string `html:"name"`
			Email string `html:"email"`
			Age   int    `html:"age"`
		}{1, "John Doe", "john@example.com", 30}
	case "product", "producto":
		return &struct {
			ID    int     `html:"id"`
			Name  string  `html:"name"`
			Price float64 `html:"price"`
			Stock int     `html:"stock"`
		}{1, "Sample Product", 29.99, 100}
	case "order", "pedido":
		return &struct {
			ID     int     `html:"id"`
			UserID int     `html:"user_id"`
			Total  float64 `html:"total"`
			Status string  `html:"status"`
		}{1, 1, 59.98, "pending"}
	default:
		return &struct {
			ID    int    `html:"id"`
			Title string `html:"title"`
			Value string `html:"content"`
		}{1, fmt.Sprintf("Sample %s", entityName), "Sample content"}
	}
}

func (ul *UniversalLiveViewDSL) createMockEntitySlice(entityName string) []interface{} {
	mockEntity := ul.createMockEntity(entityName)
	return []interface{}{mockEntity, mockEntity, mockEntity}
}

// Parse executes an HTML generation command
func (ul *UniversalLiveViewDSL) Parse(command string) (*dslbuilder.Result, error) {
	return ul.dsl.Parse(command)
}

// Use executes an HTML generation command with context
func (ul *UniversalLiveViewDSL) Use(command string, context map[string]interface{}) (*dslbuilder.Result, error) {
	return ul.dsl.Use(command, context)
}

// GetGenerator returns the underlying HTML generator
func (ul *UniversalLiveViewDSL) GetGenerator() *UniversalHTMLGenerator {
	return ul.generator
}

// SetContext sets a context value
func (ul *UniversalLiveViewDSL) SetContext(key string, value interface{}) {
	ul.dsl.SetContext(key, value)
}