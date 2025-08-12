# Image DSL - Comprehensive Image Processing Language

A powerful domain-specific language for image processing, featuring Fourier/Laplace transforms, advanced filtering, watermarking, OCR capabilities, and license plate detection.

## Features

### 🔬 Mathematical Transforms
- **Fast Fourier Transform (FFT/IFFT)**: Frequency domain analysis and reconstruction
- **Discrete Cosine Transform (DCT/IDCT)**: JPEG-style compression transforms
- **Laplace Transform**: Edge detection and image analysis
- **Inverse Laplace Transform**: Image reconstruction

### 🖼️ Image Format Support
- **BMP**: Windows Bitmap format
- **JPEG/JPG**: Joint Photographic Experts Group format with quality control
- **PNG**: Portable Network Graphics with lossless compression
- **GIF**: Graphics Interchange Format for simple animations

### 🎨 Image Editing Operations
- **Resize**: By dimensions or percentage
- **Rotate**: Arbitrary angle rotation
- **Flip**: Horizontal and vertical flipping
- **Crop**: Extract regions of interest
- **Compression**: JPEG, PNG, WebP formats with quality control

### 🔧 Advanced Filters
- **Blur Filters**: Gaussian, Box blur, Motion blur
- **Edge Detection**: Sobel, Prewitt, Canny, Laplacian
- **Enhancement**: Sharpen, Emboss
- **Noise Reduction**: Median, Bilateral filtering
- **Custom Kernels**: Apply custom convolution matrices

### 💧 Watermarking
- **Text Watermarks**: Add text at various positions with opacity control
- **Image Watermarks**: Overlay images with transparency
- **Positions**: top_left, top_right, bottom_left, bottom_right, center, tiled, diagonal
- **Watermark Removal**: Multiple methods (inpaint, median, frequency domain)

### 🚗 License Plate Processing
- **Detection**: Automatic license plate region detection
- **Extraction**: Crop detected plate region
- **OCR**: Optical Character Recognition for plate text
- **Pre-processing**: Edge enhancement and noise reduction

### 🔐 CAPTCHA Processing
- **Pre-processing**: Denoise and threshold for better recognition
- **OCR**: Extract text from CAPTCHA images
- **Enhancement**: Improve readability through filtering

### 📊 Image Analysis
- **Histogram**: Compute intensity distribution
- **Histogram Equalization**: Improve contrast
- **Edge Detection**: Find object boundaries
- **Statistics**: Image properties and metrics

## Installation

```bash
# Clone the repository
git clone https://github.com/arturoeanton/go-dsl.git

# Navigate to image_dsl example
cd go-dsl/examples/image_dsl

# Install dependencies
go mod tidy

# Run the example
go run main.go
```

## DSL Command Reference

### File Operations

| Command | Description | Example |
|---------|-------------|---------|
| `load "file"` | Load image from file | `load "photo.jpg"` |
| `save "file"` | Save image to file | `save "output.png"` |
| `save "file" quality N` | Save with quality | `save "output.jpg" quality 85` |

### Transform Operations

| Command | Description | Example |
|---------|-------------|---------|
| `fft` | Apply Fast Fourier Transform | `fft` |
| `ifft` | Apply Inverse FFT | `ifft` |
| `dct` | Apply Discrete Cosine Transform | `dct` |
| `idct` | Apply Inverse DCT | `idct` |
| `laplace` | Apply Laplace transform | `laplace` |
| `inverse_laplace` | Apply inverse Laplace | `inverse_laplace` |

### Editing Operations

| Command | Description | Example |
|---------|-------------|---------|
| `resize width W height H` | Resize to dimensions | `resize width 640 height 480` |
| `resize N%` | Resize by percentage | `resize 50%` |
| `rotate N degrees` | Rotate image | `rotate 45 degrees` |
| `flip horizontal` | Flip horizontally | `flip horizontal` |
| `flip vertical` | Flip vertically | `flip vertical` |
| `crop x X y Y width W height H` | Crop region | `crop x 10 y 10 width 100 height 100` |

### Filter Operations

| Command | Description | Example |
|---------|-------------|---------|
| `filter TYPE` | Apply filter | `filter gaussian` |
| `filter TYPE strength N` | Apply with strength | `filter gaussian strength 2.5` |

Available filters:
- `gaussian` - Gaussian blur
- `median` - Median filter for noise removal
- `bilateral` - Edge-preserving smoothing
- `sobel` - Sobel edge detection
- `prewitt` - Prewitt edge detection
- `canny` - Canny edge detection
- `laplacian` - Laplacian edge detection
- `sharpen` - Image sharpening
- `emboss` - Emboss effect
- `blur` - Box blur
- `motion_blur` - Motion blur effect

### Watermark Operations

| Command | Description | Example |
|---------|-------------|---------|
| `watermark text "TEXT" position POS` | Add text watermark | `watermark text "©2024" position top_right` |
| `watermark text "TEXT" position POS opacity N` | Text with opacity | `watermark text "DRAFT" position center opacity 0.3` |
| `watermark image "FILE" position POS` | Add image watermark | `watermark image "logo.png" position bottom_right` |
| `watermark image "FILE" position POS opacity N` | Image with opacity | `watermark image "logo.png" position center opacity 0.5` |
| `remove_watermark` | Remove watermark | `remove_watermark` |
| `remove_watermark method "METHOD"` | Remove with method | `remove_watermark method "frequency"` |

Positions: `top_left`, `top_right`, `bottom_left`, `bottom_right`, `center`, `tiled`, `diagonal`

### OCR Operations

| Command | Description | Example |
|---------|-------------|---------|
| `ocr_plate` | Extract license plate text | `ocr_plate` |
| `ocr_captcha` | Extract CAPTCHA text | `ocr_captcha` |
| `detect_plate` | Detect plate region | `detect_plate` |
| `extract_plate` | Extract plate region | `extract_plate` |

### Adjustment Operations

| Command | Description | Example |
|---------|-------------|---------|
| `brightness N` | Adjust brightness | `brightness 0.2` |
| `contrast N` | Adjust contrast | `contrast 1.5` |
| `saturation N` | Adjust saturation | `saturation 1.2` |
| `denoise` | Remove noise | `denoise` |

### Analysis Operations

| Command | Description | Example |
|---------|-------------|---------|
| `histogram` | Compute histogram | `histogram` |
| `equalize` | Equalize histogram | `equalize` |
| `edge_detect` | Detect edges | `edge_detect` |

### Compression Operations

| Command | Description | Example |
|---------|-------------|---------|
| `compress type TYPE quality N` | Compress image | `compress type jpeg quality 85` |

Types: `jpeg`, `png`, `webp`

## Usage Examples

### Basic Image Processing
```dsl
load "input.jpg"
resize 50%
filter gaussian strength 2
brightness 0.1
save "output.jpg" quality 90
```

### License Plate Recognition
```dsl
load "car_photo.jpg"
detect_plate
extract_plate
filter sharpen
ocr_plate
save "plate.png"
```

### Watermark Addition and Removal
```dsl
load "photo.jpg"
watermark text "COPYRIGHT 2024" position bottom_right opacity 0.4
save "watermarked.jpg"

load "watermarked.jpg"
remove_watermark method "frequency"
save "clean.jpg"
```

### CAPTCHA Processing
```dsl
load "captcha.png"
denoise
filter median strength 2
contrast 1.5
ocr_captcha
```

### Advanced Image Analysis
```dsl
load "image.png"
fft
# Process in frequency domain
ifft
edge_detect
histogram
equalize
save "analyzed.png"
```

### Complete Workflow Example
```dsl
# Load and prepare image
load "raw_photo.jpg"
resize width 1920 height 1080
denoise

# Apply enhancements
brightness 0.1
contrast 1.2
saturation 1.1

# Apply artistic filter
filter gaussian strength 1.5
filter sharpen

# Add watermark
watermark text "PROCESSED" position bottom_left opacity 0.3

# Save final result
compress type jpeg quality 95
save "final_output.jpg"
```

## Programmatic Usage

```go
package main

import (
    "fmt"
    "github.com/arturoeanton/go-dsl/examples/image_dsl/universal"
)

func main() {
    // Create DSL instance
    imageDSL := universal.NewImageDSL()
    
    // Execute commands
    commands := []string{
        `load "input.jpg"`,
        `filter gaussian strength 2`,
        `watermark text "SAMPLE" position center opacity 0.5`,
        `save "output.jpg" quality 90`,
    }
    
    for _, cmd := range commands {
        result, err := imageDSL.Parse(cmd)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        } else {
            fmt.Printf("Result: %v\n", result)
        }
    }
    
    // Direct engine access
    engine := imageDSL.GetEngine()
    engine.LoadImage("photo.jpg")
    engine.ApplyFilter("gaussian", 2.0)
    engine.AddTextWatermark("COPYRIGHT", "bottom_right", 0.5)
    engine.SaveImage("processed.jpg", 95)
}
```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./universal/...

# Run tests with coverage
go test -cover ./universal/...

# Run specific test
go test -run TestFilterCommands ./universal/

# Run benchmarks
go test -bench=. ./universal/
```

## Architecture

The Image DSL consists of three main components:

### 1. DSL Parser (`image_dsl.go`)
- Defines grammar and syntax
- Tokenizes and parses commands
- Maps commands to engine functions

### 2. Image Engine (`image_engine.go`)
- Implements all image processing algorithms
- Handles file I/O for multiple formats
- Provides transform and filter implementations

### 3. Universal Pattern
- Self-contained module
- Can be copied and used independently
- No external dependencies beyond standard Go libraries

## Performance Considerations

- **FFT/DCT**: O(n log n) complexity for transforms
- **Filters**: Kernel size affects performance (typically O(n*k²))
- **Memory**: Images processed in memory, suitable for images < 100MP
- **Parallel Processing**: Some operations can be parallelized for better performance

## Limitations

- OCR functionality is simulated (production would require Tesseract or similar)
- WebP compression is simulated using JPEG
- Some advanced filters use simplified implementations
- License plate detection uses basic edge detection (production would use ML models)

## Future Enhancements

Potential additions:
- Real OCR integration (Tesseract)
- Machine learning-based plate detection
- Neural network style transfer
- HDR image processing
- RAW format support
- Video frame processing
- Batch processing support
- GPU acceleration

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass
2. New features include tests
3. Documentation is updated
4. Code follows Go best practices

## License

This example is part of the go-dsl project and follows the same license terms.

## Acknowledgments

- Uses Go's standard image libraries
- BMP support from golang.org/x/image
- Inspired by ImageMagick and OpenCV command interfaces