# Image DSL - Lenguaje Completo de Procesamiento de Imágenes

Un poderoso lenguaje de dominio específico para procesamiento de imágenes, con transformadas de Fourier/Laplace, filtrado avanzado, marcas de agua, capacidades OCR y detección de matrículas.

## Características

### 🔬 Transformadas Matemáticas
- **Transformada Rápida de Fourier (FFT/IFFT)**: Análisis y reconstrucción en el dominio de frecuencia
- **Transformada de Coseno Discreta (DCT/IDCT)**: Transformadas de compresión estilo JPEG
- **Transformada de Laplace**: Detección de bordes y análisis de imágenes
- **Transformada Inversa de Laplace**: Reconstrucción de imágenes

### 🖼️ Soporte de Formatos de Imagen
- **BMP**: Formato Windows Bitmap
- **JPEG/JPG**: Formato Joint Photographic Experts Group con control de calidad
- **PNG**: Portable Network Graphics con compresión sin pérdida
- **GIF**: Graphics Interchange Format para animaciones simples

### 🎨 Operaciones de Edición de Imágenes
- **Redimensionar**: Por dimensiones o porcentaje
- **Rotar**: Rotación en ángulo arbitrario
- **Voltear**: Volteo horizontal y vertical
- **Recortar**: Extraer regiones de interés
- **Compresión**: Formatos JPEG, PNG, WebP con control de calidad

### 🔧 Filtros Avanzados
- **Filtros de Desenfoque**: Gaussiano, desenfoque de caja, desenfoque de movimiento
- **Detección de Bordes**: Sobel, Prewitt, Canny, Laplaciano
- **Mejora**: Enfocar, relieve
- **Reducción de Ruido**: Mediana, filtrado bilateral
- **Núcleos Personalizados**: Aplicar matrices de convolución personalizadas

### 💧 Marcas de Agua
- **Marcas de Agua de Texto**: Agregar texto en varias posiciones con control de opacidad
- **Marcas de Agua de Imagen**: Superponer imágenes con transparencia
- **Posiciones**: top_left, top_right, bottom_left, bottom_right, center, tiled, diagonal
- **Eliminación de Marcas de Agua**: Múltiples métodos (inpaint, mediana, dominio de frecuencia)

### 🚗 Procesamiento de Matrículas
- **Detección**: Detección automática de región de matrícula
- **Extracción**: Recortar región de matrícula detectada
- **OCR**: Reconocimiento Óptico de Caracteres para texto de matrícula
- **Pre-procesamiento**: Mejora de bordes y reducción de ruido

### 🔐 Procesamiento de CAPTCHA
- **Pre-procesamiento**: Eliminación de ruido y umbralización para mejor reconocimiento
- **OCR**: Extraer texto de imágenes CAPTCHA
- **Mejora**: Mejorar legibilidad mediante filtrado

### 📊 Análisis de Imágenes
- **Histograma**: Calcular distribución de intensidad
- **Ecualización de Histograma**: Mejorar contraste
- **Detección de Bordes**: Encontrar límites de objetos
- **Estadísticas**: Propiedades y métricas de imagen

## Instalación

```bash
# Clonar el repositorio
git clone https://github.com/arturoeanton/go-dsl.git

# Navegar al ejemplo image_dsl
cd go-dsl/examples/image_dsl

# Instalar dependencias
go mod tidy

# Ejecutar el ejemplo
go run main.go
```

## Referencia de Comandos DSL

### Operaciones de Archivo

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `load "archivo"` | Cargar imagen desde archivo | `load "foto.jpg"` |
| `save "archivo"` | Guardar imagen en archivo | `save "salida.png"` |
| `save "archivo" quality N` | Guardar con calidad | `save "salida.jpg" quality 85` |

### Operaciones de Transformada

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `fft` | Aplicar Transformada Rápida de Fourier | `fft` |
| `ifft` | Aplicar FFT Inversa | `ifft` |
| `dct` | Aplicar Transformada de Coseno Discreta | `dct` |
| `idct` | Aplicar DCT Inversa | `idct` |
| `laplace` | Aplicar transformada de Laplace | `laplace` |
| `inverse_laplace` | Aplicar Laplace inversa | `inverse_laplace` |

### Operaciones de Edición

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `resize width W height H` | Redimensionar a dimensiones | `resize width 640 height 480` |
| `resize N%` | Redimensionar por porcentaje | `resize 50%` |
| `rotate N degrees` | Rotar imagen | `rotate 45 degrees` |
| `flip horizontal` | Voltear horizontalmente | `flip horizontal` |
| `flip vertical` | Voltear verticalmente | `flip vertical` |
| `crop x X y Y width W height H` | Recortar región | `crop x 10 y 10 width 100 height 100` |

### Operaciones de Filtro

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `filter TIPO` | Aplicar filtro | `filter gaussian` |
| `filter TIPO strength N` | Aplicar con intensidad | `filter gaussian strength 2.5` |

Filtros disponibles:
- `gaussian` - Desenfoque gaussiano
- `median` - Filtro de mediana para eliminación de ruido
- `bilateral` - Suavizado preservando bordes
- `sobel` - Detección de bordes Sobel
- `prewitt` - Detección de bordes Prewitt
- `canny` - Detección de bordes Canny
- `laplacian` - Detección de bordes Laplaciana
- `sharpen` - Enfoque de imagen
- `emboss` - Efecto de relieve
- `blur` - Desenfoque de caja
- `motion_blur` - Efecto de desenfoque de movimiento

### Operaciones de Marca de Agua

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `watermark text "TEXTO" position POS` | Agregar marca de agua de texto | `watermark text "©2024" position top_right` |
| `watermark text "TEXTO" position POS opacity N` | Texto con opacidad | `watermark text "BORRADOR" position center opacity 0.3` |
| `watermark image "ARCHIVO" position POS` | Agregar marca de agua de imagen | `watermark image "logo.png" position bottom_right` |
| `watermark image "ARCHIVO" position POS opacity N` | Imagen con opacidad | `watermark image "logo.png" position center opacity 0.5` |
| `remove_watermark` | Eliminar marca de agua | `remove_watermark` |
| `remove_watermark method "MÉTODO"` | Eliminar con método | `remove_watermark method "frequency"` |

Posiciones: `top_left`, `top_right`, `bottom_left`, `bottom_right`, `center`, `tiled`, `diagonal`

### Operaciones OCR

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `ocr_plate` | Extraer texto de matrícula | `ocr_plate` |
| `ocr_captcha` | Extraer texto de CAPTCHA | `ocr_captcha` |
| `detect_plate` | Detectar región de matrícula | `detect_plate` |
| `extract_plate` | Extraer región de matrícula | `extract_plate` |

### Operaciones de Ajuste

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `brightness N` | Ajustar brillo | `brightness 0.2` |
| `contrast N` | Ajustar contraste | `contrast 1.5` |
| `saturation N` | Ajustar saturación | `saturation 1.2` |
| `denoise` | Eliminar ruido | `denoise` |

### Operaciones de Análisis

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `histogram` | Calcular histograma | `histogram` |
| `equalize` | Ecualizar histograma | `equalize` |
| `edge_detect` | Detectar bordes | `edge_detect` |

### Operaciones de Compresión

| Comando | Descripción | Ejemplo |
|---------|-------------|---------|
| `compress type TIPO quality N` | Comprimir imagen | `compress type jpeg quality 85` |

Tipos: `jpeg`, `png`, `webp`

## Ejemplos de Uso

### Procesamiento Básico de Imágenes
```dsl
load "entrada.jpg"
resize 50%
filter gaussian strength 2
brightness 0.1
save "salida.jpg" quality 90
```

### Reconocimiento de Matrículas
```dsl
load "foto_auto.jpg"
detect_plate
extract_plate
filter sharpen
ocr_plate
save "matricula.png"
```

### Adición y Eliminación de Marcas de Agua
```dsl
load "foto.jpg"
watermark text "COPYRIGHT 2024" position bottom_right opacity 0.4
save "con_marca.jpg"

load "con_marca.jpg"
remove_watermark method "frequency"
save "limpia.jpg"
```

### Procesamiento de CAPTCHA
```dsl
load "captcha.png"
denoise
filter median strength 2
contrast 1.5
ocr_captcha
```

### Análisis Avanzado de Imágenes
```dsl
load "imagen.png"
fft
# Procesar en dominio de frecuencia
ifft
edge_detect
histogram
equalize
save "analizada.png"
```

### Ejemplo de Flujo de Trabajo Completo
```dsl
# Cargar y preparar imagen
load "foto_cruda.jpg"
resize width 1920 height 1080
denoise

# Aplicar mejoras
brightness 0.1
contrast 1.2
saturation 1.1

# Aplicar filtro artístico
filter gaussian strength 1.5
filter sharpen

# Agregar marca de agua
watermark text "PROCESADO" position bottom_left opacity 0.3

# Guardar resultado final
compress type jpeg quality 95
save "salida_final.jpg"
```

## Uso Programático

```go
package main

import (
    "fmt"
    "github.com/arturoeanton/go-dsl/examples/image_dsl/universal"
)

func main() {
    // Crear instancia DSL
    imageDSL := universal.NewImageDSL()
    
    // Ejecutar comandos
    comandos := []string{
        `load "entrada.jpg"`,
        `filter gaussian strength 2`,
        `watermark text "MUESTRA" position center opacity 0.5`,
        `save "salida.jpg" quality 90`,
    }
    
    for _, cmd := range comandos {
        resultado, err := imageDSL.Parse(cmd)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        } else {
            fmt.Printf("Resultado: %v\n", resultado)
        }
    }
    
    // Acceso directo al motor
    motor := imageDSL.GetEngine()
    motor.LoadImage("foto.jpg")
    motor.ApplyFilter("gaussian", 2.0)
    motor.AddTextWatermark("COPYRIGHT", "bottom_right", 0.5)
    motor.SaveImage("procesada.jpg", 95)
}
```

## Pruebas

Ejecutar el conjunto completo de pruebas:

```bash
# Ejecutar todas las pruebas
go test ./universal/...

# Ejecutar pruebas con cobertura
go test -cover ./universal/...

# Ejecutar prueba específica
go test -run TestFilterCommands ./universal/

# Ejecutar benchmarks
go test -bench=. ./universal/
```

## Arquitectura

El Image DSL consta de tres componentes principales:

### 1. Analizador DSL (`image_dsl.go`)
- Define gramática y sintaxis
- Tokeniza y analiza comandos
- Mapea comandos a funciones del motor

### 2. Motor de Imagen (`image_engine.go`)
- Implementa todos los algoritmos de procesamiento de imágenes
- Maneja E/S de archivos para múltiples formatos
- Proporciona implementaciones de transformadas y filtros

### 3. Patrón Universal
- Módulo autocontenido
- Puede ser copiado y usado independientemente
- Sin dependencias externas más allá de las bibliotecas estándar de Go

## Consideraciones de Rendimiento

- **FFT/DCT**: Complejidad O(n log n) para transformadas
- **Filtros**: El tamaño del núcleo afecta el rendimiento (típicamente O(n*k²))
- **Memoria**: Imágenes procesadas en memoria, adecuado para imágenes < 100MP
- **Procesamiento Paralelo**: Algunas operaciones pueden paralelizarse para mejor rendimiento

## Limitaciones

- La funcionalidad OCR está simulada (producción requeriría Tesseract o similar)
- La compresión WebP está simulada usando JPEG
- Algunos filtros avanzados usan implementaciones simplificadas
- La detección de matrículas usa detección básica de bordes (producción usaría modelos ML)

## Mejoras Futuras

Adiciones potenciales:
- Integración real de OCR (Tesseract)
- Detección de matrículas basada en aprendizaje automático
- Transferencia de estilo con redes neuronales
- Procesamiento de imágenes HDR
- Soporte de formato RAW
- Procesamiento de cuadros de video
- Soporte de procesamiento por lotes
- Aceleración GPU

## Contribuciones

¡Las contribuciones son bienvenidas! Por favor asegúrese de:
1. Todas las pruebas pasen
2. Las nuevas características incluyan pruebas
3. La documentación esté actualizada
4. El código siga las mejores prácticas de Go

## Licencia

Este ejemplo es parte del proyecto go-dsl y sigue los mismos términos de licencia.

## Agradecimientos

- Usa las bibliotecas de imágenes estándar de Go
- Soporte BMP de golang.org/x/image
- Inspirado en las interfaces de comandos de ImageMagick y OpenCV