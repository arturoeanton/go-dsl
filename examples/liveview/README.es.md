# DSL Universal Generador de HTML para LiveView

Un generador de HTML universal y dinámico diseñado específicamente para integración con [go-echo-live-view](https://github.com/arturoeanton/go-echo-live-view), que funciona con cualquier tipo de estructura usando reflexión de Go.

## 🎯 Características

- **100% Universal**: Funciona con cualquier estructura Go usando reflexión
- **Optimizado para LiveView**: Diseñado específicamente para Phoenix LiveView pattern
- **Bilingüe**: Soporta comandos en Español e Inglés
- **Componentes Dinámicos**: Generación dinámica de componentes HTML
- **go-echo-live-view Ready**: Integración perfecta con Echo LiveView
- **Contexto Dinámico**: Soporte para contextos dinámicos
- **Tags de Estructura**: Soporte para tags `html:"fieldname"`
- **Eventos LiveView**: Generación automática de eventos Phoenix LiveView

## 🚀 Uso

```bash
cd examples/liveview
go run main.go
```

## 📖 Sintaxis de Comandos

### Generación de Formularios (Español)
```
generar formulario para ENTIDAD
crear formulario para ENTIDAD con accion "NOMBRE_ACCION"
```

### Form Generation (English)
```
generate form for ENTITY
create form for ENTITY with action "ACTION_NAME"
```

## 🔧 Ejemplos de Comandos

### Formularios / Forms
```go
// Español
generar formulario para user
crear formulario para producto con accion "create_product"

// English
generate form for user
create form for product with action "create_product"
```

### Tablas / Tables
```go
// Español
generar tabla para user
mostrar tabla de producto

// English
generate table for user
show table of product
```

### Tarjetas / Cards
```go
// Español
generar tarjeta para user
mostrar tarjeta de producto

// English
generate card for user
show card of product
```

### Botones / Buttons
```go
// Español
generar boton con texto "Guardar"
crear boton con texto "Eliminar" y accion "delete_item"

// English
generate button with text "Save"
create button with text "Delete" and action "delete_item"
```

### Modales / Modals
```go
// Español
generar modal con titulo "Confirmar Acción"

// English
generate modal with title "Confirm Action"
```

### Listas / Lists
```go
// Español
generar lista de user
mostrar lista de producto

// English
generate list of user
show list of product
```

### Páginas / Pages
```go
// Español
generar pagina con plantilla "layout"
crear pagina con plantilla "crud"

// English
generate page with template "layout"
create page with template "crud"
```

## 📊 Soporte de Tipos

El generador funciona con **cualquier estructura Go**:

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

## 🎭 Componentes Disponibles

### Componentes Base
- **form** - Formularios con validación LiveView
- **input** - Campos de entrada con eventos
- **button** - Botones con acciones LiveView
- **table** - Tablas con ordenación y filtrado
- **card** - Tarjetas de información
- **list** - Listas dinámicas
- **modal** - Modales interactivos

### Plantillas de Página
- **layout** - Plantilla base HTML5
- **crud** - Plantilla CRUD completa
- **dashboard** - Plantilla de dashboard

## ⚡ Eventos LiveView Generados

El DSL genera automáticamente eventos Phoenix LiveView:

```html
<!-- Eventos de formulario -->
<form phx-submit="submit_form" phx-change="form_change">
  <input phx-blur="validate_field" phx-focus="input_focus" />
</form>

<!-- Eventos de botón -->
<button phx-click="button_action">Click Me</button>

<!-- Eventos de tabla -->
<table phx-update="stream">
  <th phx-click="sort" phx-value-field="name">Name</th>
</table>

<!-- Eventos de modal -->
<div phx-click-away="close_modal" phx-key="escape">Modal</div>
```

## 🔧 API del Generador

### Crear Generador
```go
liveview := universal.NewUniversalLiveViewDSL()
generator := liveview.GetGenerator()
```

### Establecer Contexto
```go
users := []*User{{1, "Ana", "ana@example.com", 28, "admin", "active"}}
liveview.SetContext("user", users)
```

### Generar Componentes
```go
// Generar formulario
command := `generar formulario para user`
result, err := liveview.Parse(command)

// Con contexto dinámico
context := map[string]interface{}{
    "user": customUsers,
    "theme": "dark",
}
result, err := liveview.Use(command, context)
```

### Componentes Personalizados
```go
// Agregar componente personalizado
template := &universal.ComponentTemplate{
    Name:     "custom_card",
    Template: `<div class="custom {{.classes}}">{{.content}}</div>`,
    Fields:   []string{"content", "classes"},
    Events:   []string{"click", "hover"},
}
generator.AddComponent("custom_card", template)
```

## 🌐 Integración con go-echo-live-view

### Handler de Echo
```go
func (h *Handler) RenderUserForm(c echo.Context) error {
    liveview := universal.NewUniversalLiveViewDSL()
    liveview.SetContext("user", h.users)
    
    result, err := liveview.Parse("generar formulario para user")
    if err != nil {
        return err
    }
    
    return c.HTML(200, result.GetOutput())
}
```

### WebSocket Updates
```go
func (h *Handler) HandleUserUpdate(c echo.Context) error {
    // Actualizar datos
    h.updateUser(userID, userData)
    
    // Regenerar componente
    liveview := universal.NewUniversalLiveViewDSL()
    liveview.SetContext("user", h.getUser(userID))
    
    result, err := liveview.Parse("generar tarjeta para user")
    if err != nil {
        return err
    }
    
    // Enviar actualización via WebSocket
    return h.sendLiveUpdate(c, result.GetOutput())
}
```

## 🎯 Casos de Uso

- **Admin Panels**: Interfaces de administración dinámicas
- **E-commerce**: Catálogos de productos, carritos de compra
- **CMS**: Gestión de contenido en tiempo real
- **Dashboards**: Paneles de control interactivos
- **Forms**: Formularios complejos con validación
- **Tables**: Tablas de datos con filtrado y ordenación
- **Real-time Apps**: Aplicaciones en tiempo real con WebSockets

## 🏗️ Arquitectura

- **UniversalLiveViewDSL**: Parser de comandos bilingüe
- **UniversalHTMLGenerator**: Motor de generación HTML
- **ComponentTemplate**: Sistema de plantillas reutilizables
- **Context Engine**: Manejo de contextos dinámicos
- **Event System**: Generación automática de eventos LiveView

## ✅ Características Empresariales

- ✅ **100% Genérico** - Funciona con cualquier estructura
- ✅ **LiveView Optimizado** - Eventos Phoenix LiveView automáticos
- ✅ **Bilingüe** - Comandos en Español e Inglés
- ✅ **Componentes Reutilizables** - Sistema de componentes modular
- ✅ **Contexto Dinámico** - Datos dinámicos en tiempo real
- ✅ **WebSocket Ready** - Listo para actualizaciones en tiempo real
- ✅ **Production Ready** - Listo para producción
- ✅ **Zero Dependencies** - Solo Go estándar + go-dsl

## 🚀 Integración Completa

### Stack Tecnológico
- **Go Echo**: Framework web
- **go-echo-live-view**: LiveView para Go
- **WebSockets**: Comunicación en tiempo real
- **HTML/CSS/JS**: Frontend dinámico
- **go-dsl**: Motor DSL universal

### Flujo de Trabajo
1. **Definir Entidades**: Estructuras Go con tags HTML
2. **Escribir Comandos**: DSL bilingüe para generar componentes
3. **Integrar con Echo**: Handlers que usan el generador
4. **WebSocket Updates**: Actualizaciones en tiempo real
5. **UI Dinámica**: Interfaz reactiva automática

¡Un generador HTML completo y universal para aplicaciones LiveView en Go! 🚀