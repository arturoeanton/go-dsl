package universal

import (
	"fmt"
	"reflect"
	"strings"
)

// UniversalHTMLGenerator handles generic HTML generation for any struct type using reflection
type UniversalHTMLGenerator struct {
	components map[string]*ComponentTemplate
	templates  map[string]string
}

// ComponentTemplate represents a reusable HTML component template
type ComponentTemplate struct {
	Name       string
	Template   string
	Fields     []string
	Attributes map[string]string
	Children   []string
	Events     []string
}

// HTMLElement represents a generic HTML element
type HTMLElement struct {
	Tag        string                 `html:"tag"`
	Content    string                 `html:"content"`
	Attributes map[string]interface{} `html:"attributes"`
	Classes    []string               `html:"classes"`
	Events     map[string]string      `html:"events"`
	Children   []*HTMLElement         `html:"children"`
}

// NewHTMLGenerator creates a new universal HTML generator
func NewHTMLGenerator() *UniversalHTMLGenerator {
	hg := &UniversalHTMLGenerator{
		components: make(map[string]*ComponentTemplate),
		templates:  make(map[string]string),
	}
	
	hg.setupDefaultComponents()
	hg.setupDefaultTemplates()
	
	return hg
}

// setupDefaultComponents sets up default HTML components for LiveView
func (hg *UniversalHTMLGenerator) setupDefaultComponents() {
	// Form component
	hg.components["form"] = &ComponentTemplate{
		Name:     "form",
		Template: `<form class="{{.classes}}" {{.attributes}} {{.events}}>{{.content}}</form>`,
		Fields:   []string{"action", "method", "classes"},
		Attributes: map[string]string{
			"phx-submit": "submit_form",
			"phx-change": "form_change",
		},
		Events: []string{"submit", "change"},
	}
	
	// Input component
	hg.components["input"] = &ComponentTemplate{
		Name:     "input",
		Template: `<input type="{{.type}}" name="{{.name}}" value="{{.value}}" class="{{.classes}}" {{.attributes}} {{.events}} />`,
		Fields:   []string{"type", "name", "value", "placeholder"},
		Attributes: map[string]string{
			"phx-blur":   "input_blur",
			"phx-focus":  "input_focus",
			"phx-change": "input_change",
		},
		Events: []string{"blur", "focus", "change", "keyup"},
	}
	
	// Button component
	hg.components["button"] = &ComponentTemplate{
		Name:     "button",
		Template: `<button type="{{.type}}" class="{{.classes}}" {{.attributes}} {{.events}}>{{.content}}</button>`,
		Fields:   []string{"type", "content", "classes"},
		Attributes: map[string]string{
			"phx-click": "button_click",
		},
		Events: []string{"click"},
	}
	
	// Table component
	hg.components["table"] = &ComponentTemplate{
		Name:     "table",
		Template: `<table class="{{.classes}}" {{.attributes}}>{{.content}}</table>`,
		Fields:   []string{"classes"},
		Attributes: map[string]string{
			"phx-update": "stream",
		},
		Events: []string{},
	}
	
	// Card component
	hg.components["card"] = &ComponentTemplate{
		Name:     "card",
		Template: `<div class="card {{.classes}}" {{.attributes}}>{{.content}}</div>`,
		Fields:   []string{"title", "content", "classes"},
		Attributes: map[string]string{
			"phx-click": "card_click",
		},
		Events: []string{"click", "hover"},
	}
	
	// List component
	hg.components["list"] = &ComponentTemplate{
		Name:     "list",
		Template: `<ul class="{{.classes}}" {{.attributes}} {{.events}}>{{.content}}</ul>`,
		Fields:   []string{"classes"},
		Attributes: map[string]string{
			"phx-update": "stream",
		},
		Events: []string{},
	}
	
	// Modal component
	hg.components["modal"] = &ComponentTemplate{
		Name:     "modal",
		Template: `<div class="modal {{.classes}}" {{.attributes}} {{.events}}>{{.content}}</div>`,
		Fields:   []string{"title", "content", "classes"},
		Attributes: map[string]string{
			"phx-click-away": "close_modal",
			"phx-key":        "escape",
		},
		Events: []string{"close", "open"},
	}
}

// setupDefaultTemplates sets up default page templates
func (hg *UniversalHTMLGenerator) setupDefaultTemplates() {
	// Main layout template
	hg.templates["layout"] = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.title}}</title>
    <script defer phx-track-static type="text/javascript" src="/assets/app.js"></script>
    <link phx-track-static rel="stylesheet" href="/assets/app.css"/>
    <meta name="csrf-token" content="{{.csrf_token}}"/>
</head>
<body>
    <div id="app">
        {{.content}}
    </div>
</body>
</html>`

	// CRUD page template
	hg.templates["crud"] = `
<div class="container mx-auto p-4">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold">{{.title}}</h1>
        <button phx-click="new_{{.entity}}" class="btn btn-primary">
            New {{.entity_display}}
        </button>
    </div>
    
    <div class="grid gap-4">
        {{.search_form}}
        {{.data_table}}
        {{.pagination}}
    </div>
    
    {{.modal_form}}
</div>`

	// Dashboard template
	hg.templates["dashboard"] = `
<div class="dashboard-container">
    <div class="stats-grid">
        {{.stats_cards}}
    </div>
    
    <div class="charts-section">
        {{.charts}}
    </div>
    
    <div class="recent-activity">
        {{.activity_list}}
    </div>
</div>`
}

// GenerateComponent generates HTML for a component from any struct
func (hg *UniversalHTMLGenerator) GenerateComponent(componentType string, data interface{}) (string, error) {
	template, exists := hg.components[componentType]
	if !exists {
		return "", fmt.Errorf("component type '%s' not found", componentType)
	}
	
	return hg.processTemplate(template.Template, data)
}

// GenerateForm generates a complete form from any struct type
func (hg *UniversalHTMLGenerator) GenerateForm(entity interface{}, action string) (string, error) {
	v := reflect.ValueOf(entity)
	t := reflect.TypeOf(entity)
	
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	var formFields []string
	
	// Generate form fields based on struct fields
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		fieldName := hg.getFieldName(field)
		fieldType := hg.getInputType(field.Type.Kind())
		
		// Generate input field
		input := fmt.Sprintf(`
			<div class="form-group">
				<label for="%s" class="form-label">%s</label>
				<input type="%s" name="%s" value="%v" class="form-input" 
				       phx-blur="validate_field" phx-value-field="%s" />
			</div>`, 
			fieldName, strings.Title(fieldName), fieldType, fieldName, value, fieldName)
		
		formFields = append(formFields, input)
	}
	
	formHTML := fmt.Sprintf(`
		<form phx-submit="%s" phx-change="form_change" class="space-y-4">
			%s
			<div class="form-actions">
				<button type="submit" class="btn btn-primary">Submit</button>
				<button type="button" phx-click="cancel" class="btn btn-secondary">Cancel</button>
			</div>
		</form>`, action, strings.Join(formFields, "\n"))
	
	return formHTML, nil
}

// GenerateTable generates a data table from a slice of any struct type
func (hg *UniversalHTMLGenerator) GenerateTable(data interface{}, options map[string]interface{}) (string, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return "", fmt.Errorf("data must be a slice")
	}
	
	if v.Len() == 0 {
		return "<p class='empty-state'>No data available</p>", nil
	}
	
	// Get the first element to determine structure
	firstElement := v.Index(0).Interface()
	headers := hg.getTableHeaders(firstElement)
	
	// Generate table header
	headerHTML := "<thead><tr>"
	for _, header := range headers {
		headerHTML += fmt.Sprintf("<th class='table-header sortable' phx-click='sort' phx-value-field='%s'>%s</th>", 
			strings.ToLower(header), strings.Title(header))
	}
	headerHTML += "<th class='table-header'>Actions</th></tr></thead>"
	
	// Generate table rows
	var rows []string
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		rowHTML := hg.generateTableRow(item, headers, i)
		rows = append(rows, rowHTML)
	}
	
	tableHTML := fmt.Sprintf(`
		<table class="data-table" phx-update="stream" id="data-table">
			%s
			<tbody>
				%s
			</tbody>
		</table>`, headerHTML, strings.Join(rows, "\n"))
	
	return tableHTML, nil
}

// GenerateCard generates a card component from any struct
func (hg *UniversalHTMLGenerator) GenerateCard(data interface{}, template string) (string, error) {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)
	
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	// Extract card data
	cardData := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldName := hg.getFieldName(field)
			cardData[fieldName] = v.Field(i).Interface()
		}
	}
	
	// Use default card template if none provided
	if template == "" {
		template = `
		<div class="card" phx-click="select_item" phx-value-id="{{.id}}">
			<div class="card-header">
				<h3 class="card-title">{{.title}}</h3>
			</div>
			<div class="card-body">
				{{.content}}
			</div>
			<div class="card-footer">
				<button phx-click="edit_item" phx-value-id="{{.id}}" class="btn btn-sm">Edit</button>
				<button phx-click="delete_item" phx-value-id="{{.id}}" class="btn btn-sm btn-danger">Delete</button>
			</div>
		</div>`
	}
	
	return hg.processTemplate(template, cardData)
}

// GenerateList generates a list component from a slice
func (hg *UniversalHTMLGenerator) GenerateList(data interface{}, itemTemplate string) (string, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return "", fmt.Errorf("data must be a slice")
	}
	
	var items []string
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		itemHTML, err := hg.GenerateCard(item, itemTemplate)
		if err != nil {
			return "", err
		}
		
		listItem := fmt.Sprintf(`<li class="list-item" id="item-%d">%s</li>`, i, itemHTML)
		items = append(items, listItem)
	}
	
	listHTML := fmt.Sprintf(`
		<ul class="data-list" phx-update="stream" id="data-list">
			%s
		</ul>`, strings.Join(items, "\n"))
	
	return listHTML, nil
}

// GeneratePage generates a complete page using templates
func (hg *UniversalHTMLGenerator) GeneratePage(template string, data interface{}) (string, error) {
	templateStr, exists := hg.templates[template]
	if !exists {
		return "", fmt.Errorf("template '%s' not found", template)
	}
	
	return hg.processTemplate(templateStr, data)
}

// AddComponent adds a custom component template
func (hg *UniversalHTMLGenerator) AddComponent(name string, template *ComponentTemplate) {
	hg.components[name] = template
}

// AddTemplate adds a custom page template
func (hg *UniversalHTMLGenerator) AddTemplate(name string, template string) {
	hg.templates[name] = template
}

// Helper methods

func (hg *UniversalHTMLGenerator) getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("html"); tag != "" {
		return tag
	}
	return strings.ToLower(field.Name)
}

func (hg *UniversalHTMLGenerator) getInputType(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "text"
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "checkbox"
	default:
		return "text"
	}
}

func (hg *UniversalHTMLGenerator) getTableHeaders(item interface{}) []string {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	var headers []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			headers = append(headers, hg.getFieldName(field))
		}
	}
	
	return headers
}

func (hg *UniversalHTMLGenerator) generateTableRow(item interface{}, headers []string, index int) string {
	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)
	
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	var cells []string
	
	// Generate data cells
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			value := v.Field(i).Interface()
			cells = append(cells, fmt.Sprintf("<td class='table-cell'>%v</td>", value))
		}
	}
	
	// Add actions cell
	actionsCell := fmt.Sprintf(`
		<td class='table-cell actions'>
			<button phx-click='edit_item' phx-value-id='%d' class='btn btn-sm'>Edit</button>
			<button phx-click='delete_item' phx-value-id='%d' class='btn btn-sm btn-danger'>Delete</button>
		</td>`, index, index)
	
	cells = append(cells, actionsCell)
	
	return fmt.Sprintf("<tr class='table-row' id='row-%d'>%s</tr>", index, strings.Join(cells, ""))
}

func (hg *UniversalHTMLGenerator) processTemplate(template string, data interface{}) (string, error) {
	// Simple template processing - replace {{.field}} with values
	result := template
	
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)
	
	// Handle map[string]interface{}
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			placeholder := fmt.Sprintf("{{.%s}}", key.String())
			value := fmt.Sprintf("%v", v.MapIndex(key).Interface())
			result = strings.ReplaceAll(result, placeholder, value)
		}
		return result, nil
	}
	
	// Handle structs
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				fieldName := hg.getFieldName(field)
				placeholder := fmt.Sprintf("{{.%s}}", fieldName)
				value := fmt.Sprintf("%v", v.Field(i).Interface())
				result = strings.ReplaceAll(result, placeholder, value)
			}
		}
	}
	
	return result, nil
}

// GetAvailableComponents returns list of available component types
func (hg *UniversalHTMLGenerator) GetAvailableComponents() []string {
	var components []string
	for name := range hg.components {
		components = append(components, name)
	}
	return components
}

// GetAvailableTemplates returns list of available page templates
func (hg *UniversalHTMLGenerator) GetAvailableTemplates() []string {
	var templates []string
	for name := range hg.templates {
		templates = append(templates, name)
	}
	return templates
}