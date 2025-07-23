# Universal HTML Generator DSL for LiveView

A universal and dynamic HTML generator specifically designed for integration with [go-echo-live-view](https://github.com/arturoeanton/go-echo-live-view), working with any Go struct type using reflection.

## 🎯 Features

- **100% Universal**: Works with any Go struct using reflection
- **LiveView Optimized**: Specifically designed for Phoenix LiveView patterns
- **Bilingual**: Supports commands in Spanish and English
- **Dynamic Components**: Dynamic HTML component generation
- **go-echo-live-view Ready**: Perfect integration with Echo LiveView
- **Dynamic Context**: Dynamic context support
- **Struct Tags**: Support for `html:"fieldname"` tags
- **LiveView Events**: Automatic Phoenix LiveView event generation

## 🚀 Usage

```bash
cd examples/liveview
go run main.go
```

## 📖 Command Syntax

### Form Generation (Spanish)
```
generar formulario para ENTITY
crear formulario para ENTITY con accion "ACTION_NAME"
```

### Form Generation (English)
```
generate form for ENTITY
create form for ENTITY with action "ACTION_NAME"
```

## 🔧 Command Examples

### Forms / Formularios
```go
// Spanish
generar formulario para user
crear formulario para producto con accion "create_product"

// English
generate form for user
create form for product with action "create_product"
```

### Tables / Tablas
```go
// Spanish
generar tabla para user
mostrar tabla de producto

// English
generate table for user
show table of product
```

### Cards / Tarjetas
```go
// Spanish
generar tarjeta para user
mostrar tarjeta de producto

// English
generate card for user
show card of product
```

### Buttons / Botones
```go
// Spanish
generar boton con texto "Guardar"
crear boton con texto "Eliminar" y accion "delete_item"

// English
generate button with text "Save"
create button with text "Delete" and action "delete_item"
```

### Modals / Modales
```go
// Spanish
generar modal con titulo "Confirmar Acción"

// English
generate modal with title "Confirm Action"
```

### Lists / Listas
```go
// Spanish
generar lista de user
mostrar lista de producto

// English
generate list of user
show list of product
```

### Pages / Páginas
```go
// Spanish
generar pagina con plantilla "layout"
crear pagina con plantilla "crud"

// English
generate page with template "layout"
create page with template "crud"
```

## 📊 Type Support

The generator works with **any Go struct**:

```go
type User struct {
    ID       int    `html:"id"`
    Name     string `html:"nombre"`
    Email    string `html:"email"`
    Age      int    `html:"edad"`
    Role     string `html:"rol"`
    Status   string `html:"estado"`
}

type Product struct {
    ID          int     `html:"id"`
    Name        string  `html:"nombre"`
    Description string  `html:"descripcion"`
    Price       float64 `html:"precio"`
    Stock       int     `html:"stock"`
    Category    string  `html:"categoria"`
}

type Order struct {
    ID         int     `html:"id"`
    UserID     int     `html:"user_id"`
    Total      float64 `html:"total"`
    Status     string  `html:"estado"`
    Date       string  `html:"fecha"`
}
```

## 🎭 Available Components

### Base Components
- **form** - Forms with LiveView validation
- **input** - Input fields with events
- **button** - Buttons with LiveView actions
- **table** - Tables with sorting and filtering
- **card** - Information cards
- **list** - Dynamic lists
- **modal** - Interactive modals

### Page Templates
- **layout** - Base HTML5 template
- **crud** - Complete CRUD template
- **dashboard** - Dashboard template

## ⚡ Generated LiveView Events

The DSL automatically generates Phoenix LiveView events:

```html
<!-- Form events -->
<form phx-submit="submit_form" phx-change="form_change">
  <input phx-blur="validate_field" phx-focus="input_focus" />
</form>

<!-- Button events -->
<button phx-click="button_action">Click Me</button>

<!-- Table events -->
<table phx-update="stream">
  <th phx-click="sort" phx-value-field="name">Name</th>
</table>

<!-- Modal events -->
<div phx-click-away="close_modal" phx-key="escape">Modal</div>
```

## 🔧 Generator API

### Create Generator
```go
liveview := universal.NewUniversalLiveViewDSL()
generator := liveview.GetGenerator()
```

### Set Context
```go
users := []*User{{1, "Ana", "ana@example.com", 28, "admin", "active"}}
liveview.SetContext("user", users)
```

### Generate Components
```go
// Generate form
command := `generate form for user`
result, err := liveview.Parse(command)

// With dynamic context
context := map[string]interface{}{
    "user": customUsers,
    "theme": "dark",
}
result, err := liveview.Use(command, context)
```

### Custom Components
```go
// Add custom component
template := &universal.ComponentTemplate{
    Name:     "custom_card",
    Template: `<div class="custom {{.classes}}">{{.content}}</div>`,
    Fields:   []string{"content", "classes"},
    Events:   []string{"click", "hover"},
}
generator.AddComponent("custom_card", template)
```

## 🌐 Integration with go-echo-live-view

### Echo Handler
```go
func (h *Handler) RenderUserForm(c echo.Context) error {
    liveview := universal.NewUniversalLiveViewDSL()
    liveview.SetContext("user", h.users)
    
    result, err := liveview.Parse("generate form for user")
    if err != nil {
        return err
    }
    
    return c.HTML(200, result.GetOutput())
}
```

### WebSocket Updates
```go
func (h *Handler) HandleUserUpdate(c echo.Context) error {
    // Update data
    h.updateUser(userID, userData)
    
    // Regenerate component
    liveview := universal.NewUniversalLiveViewDSL()
    liveview.SetContext("user", h.getUser(userID))
    
    result, err := liveview.Parse("generate card for user")
    if err != nil {
        return err
    }
    
    // Send update via WebSocket
    return h.sendLiveUpdate(c, result.GetOutput())
}
```

## 🎯 Use Cases

- **Admin Panels**: Dynamic administration interfaces
- **E-commerce**: Product catalogs, shopping carts
- **CMS**: Real-time content management
- **Dashboards**: Interactive control panels
- **Forms**: Complex forms with validation
- **Tables**: Data tables with filtering and sorting
- **Real-time Apps**: Real-time applications with WebSockets

## 🏗️ Architecture

- **UniversalLiveViewDSL**: Bilingual command parser
- **UniversalHTMLGenerator**: HTML generation engine
- **ComponentTemplate**: Reusable template system
- **Context Engine**: Dynamic context handling
- **Event System**: Automatic LiveView event generation

## ✅ Enterprise Features

- ✅ **100% Generic** - Works with any struct
- ✅ **LiveView Optimized** - Automatic Phoenix LiveView events
- ✅ **Bilingual** - Spanish and English commands
- ✅ **Reusable Components** - Modular component system
- ✅ **Dynamic Context** - Real-time dynamic data
- ✅ **WebSocket Ready** - Ready for real-time updates
- ✅ **Production Ready** - Ready for production
- ✅ **Zero Dependencies** - Only standard Go + go-dsl

## 🚀 Complete Integration

### Technology Stack
- **Go Echo**: Web framework
- **go-echo-live-view**: LiveView for Go
- **WebSockets**: Real-time communication
- **HTML/CSS/JS**: Dynamic frontend
- **go-dsl**: Universal DSL engine

### Workflow
1. **Define Entities**: Go structs with HTML tags
2. **Write Commands**: Bilingual DSL to generate components
3. **Integrate with Echo**: Handlers using the generator
4. **WebSocket Updates**: Real-time updates
5. **Dynamic UI**: Automatic reactive interface

A complete and universal HTML generator for LiveView applications in Go! 🚀