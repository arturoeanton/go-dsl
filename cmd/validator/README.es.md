# Validador de DSL

Herramienta integral para validar definiciones de gramática DSL y detectar problemas potenciales antes del tiempo de ejecución.

## Descripción General

El Validador de DSL analiza tus archivos de configuración DSL para identificar errores de sintaxis, problemas semánticos y posibles problemas que podrían causar fallos en tiempo de ejecución. Proporciona retroalimentación detallada para ayudarte a crear DSLs robustos y eficientes.

## Instalación

```bash
go install github.com/arturoeanton/go-dsl/cmd/validator@latest
```

O compilar desde el código fuente:

```bash
cd cmd/validator
go build -o validator
```

## Uso

```bash
validator -dsl <archivo-dsl> [opciones]
```

### Opciones

- `-dsl` - Archivo de configuración DSL para validar (YAML o JSON) **[requerido]**
- `-verbose` - Mostrar información detallada de validación
- `-format` - Formato de salida: `text`, `json`, o `yaml` (por defecto: text)
- `-test` - Cadena de entrada de prueba para validar contra el DSL
- `-info` - Mostrar resumen de información del DSL
- `-strict` - Habilitar modo de validación estricto

### Ejemplos

**Validación básica:**
```bash
validator -dsl calculadora.yaml
```

**Validación detallada con información:**
```bash
validator -dsl consultas.json -verbose -info
```

**Prueba con entrada de muestra:**
```bash
validator -dsl contabilidad.yaml -test "venta de 1000" -strict
```

**Salida JSON para integración CI/CD:**
```bash
validator -dsl midsl.yaml -format json
```

## Verificaciones de Validación

### Validación de Tokens
- ✓ Patrones regex válidos
- ✓ Análisis de complejidad de patrones
- ✓ Detección de patrones duplicados
- ✓ Advertencias de patrones demasiado amplios
- ✓ Detección de caracteres especiales sin escapar

### Validación de Reglas
- ✓ Verificación de referencias token/regla
- ✓ Detección de patrones vacíos
- ✓ Advertencias de reglas duplicadas
- ✓ Detección de recursión por la izquierda
- ✓ Seguimiento de referencias de acciones

### Análisis de Gramática
- ✓ Identificación de regla de inicio
- ✓ Detección de reglas inalcanzables
- ✓ Patrones de gramática ambiguos
- ✓ Conflictos de prioridad de tokens

### Mejores Prácticas
- ✓ Verificación de convenciones de nombres
- ✓ Advertencias de complejidad
- ✓ Análisis de impacto en rendimiento

## Ejemplos de Salida

### Formato Texto (Por defecto)
```
✓ Validación DSL exitosa

Información del DSL:
  Nombre: Calculadora
  Tokens: 6
  Reglas: 8

Advertencias (2):
  ⚠ [RecursiónIzquierda] La regla expr tiene recursión por la izquierda
    Detalles: La recursión por la izquierda está soportada pero puede impactar el rendimiento
  ⚠ [AcciónNoImplementada] La acción calcular está referenciada pero no implementada
    Detalles: Asegúrate de implementar todas las acciones referenciadas en las reglas
```

### Formato JSON
```json
{
  "valid": true,
  "errors": [],
  "warnings": [
    {
      "type": "LeftRecursion",
      "message": "La regla expr tiene recursión por la izquierda",
      "details": "La recursión por la izquierda está soportada pero puede impactar el rendimiento"
    }
  ],
  "info": {
    "name": "Calculadora",
    "tokenCount": 6,
    "ruleCount": 8
  }
}
```

## Reglas de Validación

### Errores (Fallan la Validación)
- Patrones regex inválidos en tokens
- Referencias a tokens/reglas no definidos
- Patrones de reglas vacíos
- Fallo en la instanciación del DSL

### Advertencias (Pasan con Precauciones)
- Reglas recursivas por la izquierda
- Nombres de reglas duplicados
- Patrones de tokens demasiado amplios
- Implementaciones de acciones faltantes
- Sin regla de inicio clara

## Casos de Uso

1. **Validación Pre-despliegue**: Verificar archivos DSL antes de producción
2. **Integración CI/CD**: Validación automática de gramática en pipelines
3. **Ayuda al Desarrollo**: Detectar problemas durante el desarrollo del DSL
4. **Documentación**: Generar documentación de estructura DSL
5. **Validación de Migración**: Verificar archivos DSL después de actualizaciones

## Códigos de Salida

- `0` - Validación exitosa
- `1` - Validación fallida o error ocurrido

## Ejemplos de Integración

### Hook Pre-commit de Git
```bash
#!/bin/sh
validator -dsl midsl.yaml -strict || exit 1
```

### GitHub Actions
```yaml
- name: Validar DSL
  run: |
    go install github.com/arturoeanton/go-dsl/cmd/validator@latest
    validator -dsl config/midsl.yaml -format json
```

### Makefile
```makefile
validar:
    @validator -dsl $(ARCHIVO_DSL) -verbose -info
```