# Guía de Instalación - go-dsl

Esta guía te ayudará a instalar y configurar go-dsl en tu entorno de desarrollo.

## 📋 Requisitos del Sistema

### Requisitos Mínimos
- **Go**: Versión 1.18 o superior
- **Sistema Operativo**: Linux, macOS, Windows
- **Memoria RAM**: 512 MB disponible
- **Espacio en Disco**: 50 MB

### Requisitos Recomendados
- **Go**: Versión 1.21 o superior (para mejor rendimiento)
- **Sistema Operativo**: Linux/macOS (desarrollo nativo)
- **Memoria RAM**: 2 GB disponible (para proyectos grandes)
- **Espacio en Disco**: 200 MB (incluye ejemplos y documentación)

## 🔧 Verificación de Prerrequisitos

### 1. Verificar Instalación de Go

```bash
go version
```

**Salida esperada:**
```
go version go1.21.0 linux/amd64
```

Si Go no está instalado, descárgalo desde [golang.org/dl](https://golang.org/dl/).

### 2. Verificar Variables de Entorno

```bash
echo $GOPATH
echo $GOROOT
```

### 3. Verificar Conectividad a Módulos Go

```bash
go env GOPROXY
```

## 📦 Métodos de Instalación

### Método 1: Instalación Directa (Recomendado)

```bash
# Instalar la librería
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder

# Verificar instalación
go list -m github.com/arturoeanton/go-dsl/pkg/dslbuilder
```

### Método 2: Instalación en Proyecto Existente

```bash
# Navegar a tu proyecto
cd mi-proyecto

# Inicializar módulo Go (si no existe)
go mod init mi-proyecto

# Agregar dependencia
go get github.com/arturoeanton/go-dsl/pkg/dslbuilder

# Actualizar go.mod
go mod tidy
```

### Método 3: Clonación para Desarrollo

```bash
# Clonar repositorio completo
git clone https://github.com/arturoeanton/go-dsl.git
cd go-dsl

# Instalar dependencias
go mod download

# Ejecutar tests para verificar instalación
go test ./pkg/dslbuilder/...
```

## 🚀 Verificación de Instalación

### Test Básico de Funcionalidad

Crea un archivo `test_instalacion.go`:

```go
package main

import (
    "fmt"
    "log"
    "github.com/arturoeliasanton/go-dsl/pkg/dslbuilder"
)

func main() {
    // Crear DSL simple
    dsl := dslbuilder.New("TestInstalacion")
    
    // Definir token
    dsl.KeywordToken("HOLA", "hola")
    
    // Definir regla
    dsl.Rule("saludo", []string{"HOLA"}, "procesar")
    
    // Definir acción
    dsl.Action("procesar", func(args []interface{}) (interface{}, error) {
        return "¡Instalación exitosa!", nil
    })
    
    // Probar parsing
    result, err := dsl.Parse("hola")
    if err != nil {
        log.Fatal("Error en instalación:", err)
    }
    
    fmt.Println("✅", result.GetOutput())
    fmt.Println("🎉 go-dsl instalado correctamente")
}
```

**Ejecutar test:**
```bash
go run test_instalacion.go
```

**Salida esperada:**
```
✅ ¡Instalación exitosa!
🎉 go-dsl instalado correctamente
```

### Ejecutar Ejemplos Oficiales

```bash
# Si clonaste el repositorio
cd go-dsl

# Ejemplo básico
go run examples/simple_context/main.go

# Sistema contable empresarial
go run examples/contabilidad/main.go

# Sistema multi-país
go run examples/accounting/main.go
```

## 🛠️ Configuración del Entorno de Desarrollo

### IDE/Editor Recomendado

#### Visual Studio Code
```bash
# Instalar extensión Go
code --install-extension golang.Go
```

**Configuración recomendada (`.vscode/settings.json`):**
```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.formatTool": "gofmt",
    "go.testFlags": ["-v"],
    "go.coverageDecorator": {
        "type": "gutter"
    }
}
```

#### GoLand/IntelliJ IDEA
- Plugin Go habilitado
- Indexación de módulos activada

### Herramientas de Desarrollo Adicionales

```bash
# Linter (opcional pero recomendado)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Documentación local
go install golang.org/x/tools/cmd/godoc@latest

# Herramientas de profiling
go install github.com/google/pprof@latest
```

## 📊 Verificación Completa

### Script de Verificación Automática

Crea `verificar_instalacion.sh`:

```bash
#!/bin/bash

echo "🔍 Verificando instalación de go-dsl..."

# Verificar Go
if ! command -v go &> /dev/null; then
    echo "❌ Go no está instalado"
    exit 1
fi

echo "✅ Go $(go version | awk '{print $3}') detectado"

# Verificar módulo
if go list -m github.com/arturoeanton/go-dsl/pkg/dslbuilder &> /dev/null; then
    echo "✅ Módulo go-dsl disponible"
else
    echo "❌ Módulo go-dsl no encontrado"
    echo "🔧 Ejecutando: go get github.com/arturoeliasanton/go-dsl/pkg/dslbuilder"
    go get github.com/arturoeanton/go-dsl/pkg/dslbuilder
fi

# Test funcional
cat > temp_test.go << 'EOF'
package main
import (
    "fmt"
    "github.com/arturoeliasanton/go-dsl/pkg/dslbuilder"
)
func main() {
    dsl := dslbuilder.New("Test")
    dsl.KeywordToken("OK", "ok")
    dsl.Rule("test", []string{"OK"}, "success")
    dsl.Action("success", func(args []interface{}) (interface{}, error) {
        return "SUCCESS", nil
    })
    result, err := dsl.Parse("ok")
    if err != nil || result.GetOutput() != "SUCCESS" {
        fmt.Println("FAIL")
    } else {
        fmt.Println("PASS")
    }
}
EOF

if [ "$(go run temp_test.go)" = "PASS" ]; then
    echo "✅ Test funcional: PASS"
else
    echo "❌ Test funcional: FAIL"
fi

rm temp_test.go

echo ""
echo "🎉 Instalación verificada correctamente"
echo "📚 Lee la documentación en docs/es/"
echo "🚀 Prueba los ejemplos en examples/"
```

**Ejecutar:**
```bash
chmod +x verificar_instalacion.sh
./verificar_instalacion.sh
```

## 🔧 Solución de Problemas Comunes

### Error: "module not found"

```bash
# Limpiar caché de módulos
go clean -modcache

# Reinstalar
go get github.com/arturoeliasanton/go-dsl/pkg/dslbuilder

# Verificar proxy
go env GOPROXY
```

### Error: "cannot find package"

```bash
# Verificar GOPATH
echo $GOPATH

# Regenerar go.mod
rm go.mod go.sum
go mod init tu-proyecto
go get github.com/arturoeliasanton/go-dsl/pkg/dslbuilder
```

### Error: versión Go incompatible

```bash
# Verificar versión mínima
go version

# Actualizar Go si es necesario
# Descargar desde https://golang.org/dl/
```

### Problemas de Red/Proxy

```bash
# Configurar proxy si es necesario
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org

# O usar proxy corporativo
go env -w GOPROXY=http://tu-proxy-corporativo
```

## 📝 Próximos Pasos

Una vez completada la instalación:

1. **📖 Lee la documentación:**
   - [Guía Rápida](guia_rapida.md)
   - [Manual de Uso](manual_de_uso.md)
   - [Guía de Implementación](implementacion.md)

2. **🔬 Explora los ejemplos:**
   - Sistema contable empresarial
   - Consultas LINQ en español
   - Calculadoras especializadas

3. **🏗️ Crea tu primer DSL:**
   - Sigue los ejemplos básicos
   - Experimenta con KeywordToken
   - Prueba gramáticas recursivas

## 📞 Soporte

- **Documentación**: `docs/es/`
- **Ejemplos**: `examples/`
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/go-dsl/issues)
- **Tests**: `pkg/dslbuilder/dsl_test.go`

---

**¡Listo para crear DSLs empresariales!** 🚀