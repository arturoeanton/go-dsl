# DSL REPL

Un Bucle Interactivo de Lectura-Evaluación-Impresión (REPL) para probar y explorar tus DSLs.

## Descripción General

El DSL REPL proporciona un entorno interactivo para probar comandos de tu DSL, explorar reglas gramaticales y depurar problemas de parsing en tiempo real. Soporta variables de contexto, historial de comandos y múltiples formatos de salida.

## Instalación

```bash
go install github.com/arturoeanton/go-dsl/cmd/repl@latest
```

O compilar desde el código fuente:

```bash
cd cmd/repl
go build -o repl
```

## Uso

```bash
repl -dsl <archivo-dsl> [opciones]
```

### Opciones

- `-dsl` - Archivo de configuración DSL (YAML o JSON) **[requerido]**
- `-history` - Archivo de historial para guardar/cargar comandos
- `-context` - Archivo de contexto (JSON) para precargar
- `-ast` - Mostrar representación AST de la entrada parseada
- `-time` - Mostrar tiempo de ejecución para cada comando
- `-multiline` - Habilitar modo de entrada multilínea
- `-exec` - Ejecutar comandos (puede usarse múltiples veces)

### Ejemplos

**Sesión interactiva básica:**
```bash
repl -dsl calculadora.yaml
```

**Con contexto e historial:**
```bash
repl -dsl consultas.json -context datos.json -history historial_consultas.txt
```

**Ejecutar comandos y salir:**
```bash
repl -dsl contabilidad.yaml -exec "venta de 1000" -exec "venta de 2000"
```

**Modo depuración con AST y tiempo:**
```bash
repl -dsl midsl.yaml -ast -time
```

## Comandos del REPL

| Comando | Descripción |
|---------|-------------|
| `.help` | Mostrar comandos disponibles del REPL |
| `.exit` | Salir del REPL |
| `.history` | Mostrar historial de comandos |
| `.clear` | Limpiar la pantalla |
| `.context` | Mostrar variables de contexto actuales |
| `.set <clave> <valor>` | Establecer variable de contexto |
| `.load <archivo>` | Cargar y ejecutar comandos desde archivo |
| `.save <archivo>` | Guardar historial en archivo |
| `.ast on/off` | Alternar visualización de AST |
| `.time on/off` | Alternar visualización de tiempo de ejecución |
| `.multiline` | Alternar modo de entrada multilínea |

## Características

### Modo Interactivo
```
Calculadora> 10 + 20
30
Calculadora> 5 * (3 + 2)
25
Calculadora> .time on
Visualización de tiempo: true
Calculadora> 100 / 4
25
⏱  125µs
```

### Variables de Contexto
```
Consultas> .set usuarios ["Juan", "María", "Pedro"]
Establecido usuarios = ["Juan", "María", "Pedro"]
Consultas> select nombre from usuarios
["Juan", "María", "Pedro"]
```

### Entrada Multilínea
```
DSL> .multiline
Modo multilínea: true
DSL> crear función calcular
... con parámetros x, y
... retornar x + y
... 
Función creada exitosamente
```

### Historial de Comandos
```
DSL> .history
[1] 10 + 20 => 30
[2] 5 * 3 => 15
[3] .set x 100 
[4] x / 2 => 50
```

### Carga de Scripts
Crear un archivo de script `comandos.txt`:
```
# Comandos de calculadora
10 + 20
30 * 2
# Establecer variables
.set iva 0.16
1000 * iva
```

Cargar y ejecutar:
```
Calculadora> .load comandos.txt
[comandos.txt:2] 10 + 20
30
[comandos.txt:3] 30 * 2
60
[comandos.txt:5] .set iva 0.16
Establecido iva = 0.16
[comandos.txt:6] 1000 * iva
160
```

## Formatos de Salida

### Salida Estándar
Visualización directa de resultados:
```
Calculadora> 42
42
```

### Salida de Arreglos
Visualización indexada para arreglos:
```
Consultas> select todo from datos
[0] {nombre: "Juan", edad: 30}
[1] {nombre: "María", edad: 25}
[2] {nombre: "Pedro", edad: 35}
```

### Salida JSON
Formato legible para objetos complejos:
```
DSL> procesar datos
{
  "estado": "exitoso",
  "cantidad": 3,
  "resultados": [...]
}
```

## Casos de Uso

1. **Pruebas de DSL**: Probar rápidamente reglas gramaticales y acciones
2. **Desarrollo Interactivo**: Desarrollar características del DSL interactivamente
3. **Depuración**: Depurar problemas de parsing con visualización AST
4. **Demostraciones**: Mostrar capacidades del DSL a usuarios
5. **Aprendizaje**: Explorar sintaxis y características del DSL
6. **Procesamiento por Lotes**: Ejecutar scripts DSL desde archivos

## Consejos y Trucos

### Pruebas Rápidas
```bash
# Probar un solo comando
echo "10 + 20" | repl -dsl calculadora.yaml

# Probar múltiples comandos
repl -dsl calc.yaml -exec "10+20" -exec "30*2" -exec "15/3"
```

### Depurar Errores de Parsing
```
DSL> .ast on
Visualización AST: true
DSL> sintaxis inválida aquí
Error: Token inesperado 'sintaxis' en posición 0
```

### Pruebas de Rendimiento
```
DSL> .time on
Visualización de tiempo: true
DSL> cálculo complejo aquí
Resultado: 42
⏱  2.5ms
```

### Grabación de Sesión
```bash
# Iniciar con grabación de historial
repl -dsl midsl.yaml -history sesion.log

# Reproducir la sesión más tarde
repl -dsl midsl.yaml -exec "$(cat sesion.log | grep -v '^#')"
```