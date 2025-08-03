# SCIM Filter DSL Example

Este ejemplo implementa un DSL para el lenguaje de filtrado SCIM (System for Cross-domain Identity Management) versión 2.0.

## ¿Qué es SCIM?

SCIM (RFC 7644) es un estándar para automatizar el intercambio de información de identidad de usuario entre sistemas. El filtrado SCIM permite consultas complejas sobre recursos de usuario utilizando una sintaxis específica.

## Características Implementadas

### Operadores de Comparación
- `eq` (equal) - Igualdad
- `ne` (not equal) - Desigualdad  
- `co` (contains) - Contiene
- `sw` (starts with) - Comienza con
- `ew` (ends with) - Termina con
- `gt` (greater than) - Mayor que
- `ge` (greater than or equal) - Mayor o igual
- `lt` (less than) - Menor que
- `le` (less than or equal) - Menor o igual
- `pr` (present) - Presente (not null)

### Operadores Lógicos
- `and` - Y lógico
- `or` - O lógico
- `not` - Negación

### Agrupación
- Paréntesis para agrupar expresiones

## Ejemplos de Filtros SCIM

```bash
# Usuarios con nombre "John"
userName eq "john"

# Usuarios que contienen "example" en su email
emails[type eq "work" and value co "example.com"]

# Usuarios activos con departamento específico
active eq true and department eq "Engineering"

# Usuarios creados después de cierta fecha
meta.created gt "2023-01-01T00:00:00Z"

# Filtros complejos con agrupación
(userName sw "test" or email co "test") and active eq true
```

## Arquitectura

El ejemplo utiliza la misma arquitectura que otros ejemplos del proyecto:

- **main.go**: Punto de entrada con datos de ejemplo y casos de prueba
- **universal/**: Contiene la implementación reutilizable
  - **scim_dsl.go**: Definición del DSL SCIM
  - **scim_engine.go**: Motor de filtrado con funciones inyectables

## Funciones Inyectables

El motor permite inyectar funciones personalizadas para diferentes implementaciones:

```go
type DataProvider interface {
    // Obtener todos los usuarios
    GetUsers() ([]interface{}, error)
    
    // Filtrar usuarios por criterio específico
    FilterUsers(attribute, operator, value string) ([]interface{}, error)
    
    // Aplicar operadores lógicos
    ApplyLogicalOperator(operator string, left, right []interface{}) ([]interface{}, error)
}
```

## Uso

```bash
cd examples/scim
go run main.go
```

## Casos de Uso Reales

- **Sincronización de Directorios**: Integración entre Active Directory y sistemas cloud
- **Provisioning Automático**: Creación/actualización automática de usuarios
- **Sistemas de Identity Management**: Okta, Azure AD, Auth0
- **APIs de Usuario**: Filtrado avanzado en endpoints REST
- **Auditoría y Compliance**: Consultas complejas sobre datos de usuario

## Extensiones Futuras

- Soporte para filtros de grupos y roles
- Validación de esquemas SCIM
- Cache inteligente de consultas
- Optimización de índices automática
- Integración con bases de datos reales