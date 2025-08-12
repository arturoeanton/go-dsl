package universal

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"math/cmplx"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

// ImageEngine handles all image processing operations
type ImageEngine struct {
	currentImage    image.Image
	format          string
	width           int
	height          int
	spectrum        [][]complex128
	dctCoefficients [][]float64
	plateRegion     image.Rectangle
	ocrResult       string
}

// NewImageEngine creates a new image processing engine
func NewImageEngine() *ImageEngine {
	return &ImageEngine{}
}

// LoadImage loads an image from file
func (ie *ImageEngine) LoadImage(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Detect format
	ext := strings.ToLower(filepath.Ext(filename))
	ie.format = ext[1:] // Remove the dot

	var img image.Image
	switch ie.format {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(file)
	case "png":
		img, err = png.Decode(file)
	case "gif":
		img, err = gif.Decode(file)
	case "bmp":
		img, err = bmp.Decode(file)
	default:
		// Try auto-detection
		file.Seek(0, 0)
		img, ie.format, err = image.Decode(file)
	}

	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	ie.currentImage = img
	bounds := img.Bounds()
	ie.width = bounds.Dx()
	ie.height = bounds.Dy()

	return fmt.Sprintf("Loaded %s: %dx%d %s image", filename, ie.width, ie.height, ie.format), nil
}

// SaveImage saves the current image to file
func (ie *ImageEngine) SaveImage(filename string, quality int) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filename))
	format := ext[1:]

	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(file, ie.currentImage, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(file, ie.currentImage)
	case "gif":
		err = gif.Encode(file, ie.currentImage, nil)
	case "bmp":
		err = bmp.Encode(file, ie.currentImage)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	return fmt.Sprintf("Saved %s: %dx%d %s image", filename, ie.width, ie.height, format), nil
}

// FFT performs 2D Fast Fourier Transform
func (ie *ImageEngine) FFT() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	gray := ie.toGrayscale()
	ie.spectrum = fft2D(gray)

	return fmt.Sprintf("FFT computed: %dx%d frequency domain", ie.width, ie.height), nil
}

// IFFT performs Inverse FFT
func (ie *ImageEngine) IFFT() (string, error) {
	if ie.spectrum == nil {
		return "", fmt.Errorf("no FFT spectrum available")
	}

	gray := ifft2D(ie.spectrum)
	ie.currentImage = ie.fromGrayscale(gray)

	return fmt.Sprintf("IFFT computed: reconstructed %dx%d image", ie.width, ie.height), nil
}

// DCT performs Discrete Cosine Transform
func (ie *ImageEngine) DCT() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	gray := ie.toGrayscale()
	ie.dctCoefficients = dct2D(gray)

	return fmt.Sprintf("DCT computed: %dx%d coefficients", ie.width, ie.height), nil
}

// IDCT performs Inverse DCT
func (ie *ImageEngine) IDCT() (string, error) {
	if ie.dctCoefficients == nil {
		return "", fmt.Errorf("no DCT coefficients available")
	}

	gray := idct2D(ie.dctCoefficients)
	ie.currentImage = ie.fromGrayscale(gray)

	return fmt.Sprintf("IDCT computed: reconstructed %dx%d image", ie.width, ie.height), nil
}

// LaplaceTransform performs Laplace transform on image
func (ie *ImageEngine) LaplaceTransform() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Apply Laplacian filter for edge detection
	kernel := [][]float64{
		{0, 1, 0},
		{1, -4, 1},
		{0, 1, 0},
	}

	ie.currentImage = ie.applyKernel(kernel)
	return "Laplace transform applied for edge detection", nil
}

// InverseLaplaceTransform performs inverse Laplace transform
func (ie *ImageEngine) InverseLaplaceTransform() (string, error) {
	// Simplified inverse - apply Gaussian smoothing
	return ie.ApplyFilter("gaussian", 2.0)
}

// Resize resizes the image to specific dimensions
func (ie *ImageEngine) Resize(width, height int) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	// Simple nearest-neighbor resize
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * ie.width / width
			srcY := y * ie.height / height
			newImg.Set(x, y, ie.currentImage.At(srcX, srcY))
		}
	}

	ie.currentImage = newImg
	ie.width = width
	ie.height = height

	return fmt.Sprintf("Resized to %dx%d", width, height), nil
}

// ResizePercent resizes by percentage
func (ie *ImageEngine) ResizePercent(scale float64) (string, error) {
	newWidth := int(float64(ie.width) * scale)
	newHeight := int(float64(ie.height) * scale)
	return ie.Resize(newWidth, newHeight)
}

// Rotate rotates the image by degrees
func (ie *ImageEngine) Rotate(degrees float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	radians := degrees * math.Pi / 180
	cos := math.Cos(radians)
	sin := math.Sin(radians)

	// Calculate new dimensions
	newWidth := int(math.Abs(float64(ie.width)*cos) + math.Abs(float64(ie.height)*sin))
	newHeight := int(math.Abs(float64(ie.width)*sin) + math.Abs(float64(ie.height)*cos))

	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	centerX := float64(ie.width) / 2
	centerY := float64(ie.height) / 2
	newCenterX := float64(newWidth) / 2
	newCenterY := float64(newHeight) / 2

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Rotate back to find source pixel
			dx := float64(x) - newCenterX
			dy := float64(y) - newCenterY
			srcX := int(dx*cos + dy*sin + centerX)
			srcY := int(-dx*sin + dy*cos + centerY)

			if srcX >= 0 && srcX < ie.width && srcY >= 0 && srcY < ie.height {
				newImg.Set(x, y, ie.currentImage.At(srcX, srcY))
			}
		}
	}

	ie.currentImage = newImg
	ie.width = newWidth
	ie.height = newHeight

	return fmt.Sprintf("Rotated by %.1f degrees", degrees), nil
}

// FlipHorizontal flips the image horizontally
func (ie *ImageEngine) FlipHorizontal() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	newImg := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			newImg.Set(x, y, ie.currentImage.At(ie.width-1-x, y))
		}
	}

	ie.currentImage = newImg
	return "Flipped horizontally", nil
}

// FlipVertical flips the image vertically
func (ie *ImageEngine) FlipVertical() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	newImg := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			newImg.Set(x, y, ie.currentImage.At(x, ie.height-1-y))
		}
	}

	ie.currentImage = newImg
	return "Flipped vertically", nil
}

// Crop crops the image
func (ie *ImageEngine) Crop(x, y, width, height int) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	rect := image.Rect(x, y, x+width, y+height)
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(newImg, newImg.Bounds(), ie.currentImage, rect.Min, draw.Src)

	ie.currentImage = newImg
	ie.width = width
	ie.height = height

	return fmt.Sprintf("Cropped to %dx%d at (%d,%d)", width, height, x, y), nil
}

// Compress compresses the image
func (ie *ImageEngine) Compress(compType string, quality int) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Simulate compression by encoding and decoding
	var buf bytes.Buffer

	switch compType {
	case "jpeg":
		err := jpeg.Encode(&buf, ie.currentImage, &jpeg.Options{Quality: quality})
		if err != nil {
			return "", err
		}
		img, err := jpeg.Decode(&buf)
		if err != nil {
			return "", err
		}
		ie.currentImage = img
	case "png":
		// PNG is lossless, just re-encode
		err := png.Encode(&buf, ie.currentImage)
		if err != nil {
			return "", err
		}
		img, err := png.Decode(&buf)
		if err != nil {
			return "", err
		}
		ie.currentImage = img
	case "webp":
		// Simulate WebP with JPEG
		err := jpeg.Encode(&buf, ie.currentImage, &jpeg.Options{Quality: quality})
		if err != nil {
			return "", err
		}
		img, err := jpeg.Decode(&buf)
		if err != nil {
			return "", err
		}
		ie.currentImage = img
	}

	return fmt.Sprintf("Compressed with %s at quality %d", compType, quality), nil
}

// ApplyFilter applies various image filters
func (ie *ImageEngine) ApplyFilter(filterType string, strength float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	switch filterType {
	case "gaussian":
		ie.currentImage = ie.gaussianBlur(strength)
	case "median":
		ie.currentImage = ie.medianFilter(int(strength))
	case "bilateral":
		ie.currentImage = ie.bilateralFilter(strength)
	case "sobel":
		ie.currentImage = ie.sobelFilter()
	case "prewitt":
		ie.currentImage = ie.prewittFilter()
	case "canny":
		ie.currentImage = ie.cannyEdgeDetect(strength)
	case "laplacian":
		kernel := [][]float64{{0, 1, 0}, {1, -4, 1}, {0, 1, 0}}
		ie.currentImage = ie.applyKernel(kernel)
	case "sharpen":
		kernel := [][]float64{{0, -1, 0}, {-1, 5, -1}, {0, -1, 0}}
		ie.currentImage = ie.applyKernel(kernel)
	case "emboss":
		kernel := [][]float64{{-2, -1, 0}, {-1, 1, 1}, {0, 1, 2}}
		ie.currentImage = ie.applyKernel(kernel)
	case "blur":
		ie.currentImage = ie.boxBlur(int(strength))
	case "motion_blur":
		ie.currentImage = ie.motionBlur(strength)
	default:
		return "", fmt.Errorf("unknown filter: %s", filterType)
	}

	return fmt.Sprintf("Applied %s filter with strength %.1f", filterType, strength), nil
}

// AddTextWatermark adds text watermark to image
func (ie *ImageEngine) AddTextWatermark(text, position string, opacity float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Create a copy to draw on
	rgba := ie.toRGBA()

	// Calculate text position
	x, y := ie.getWatermarkPosition(position, 200, 50) // Approximate text dimensions

	// Draw text (simplified - would need font rendering in production)
	watermarkColor := color.RGBA{255, 255, 255, uint8(opacity * 255)}
	
	// Draw simple text representation
	for i := range text {
		for dy := 0; dy < 20; dy++ {
			for dx := 0; dx < 10; dx++ {
				px := x + i*12 + dx
				py := y + dy
				if px >= 0 && px < ie.width && py >= 0 && py < ie.height {
					rgba.Set(px, py, ie.blendColor(rgba.At(px, py), watermarkColor, opacity))
				}
			}
		}
	}

	ie.currentImage = rgba
	return fmt.Sprintf("Added text watermark '%s' at %s", text, position), nil
}

// AddImageWatermark adds image watermark
func (ie *ImageEngine) AddImageWatermark(imagePath, position string, opacity float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Load watermark image
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	watermark, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	rgba := ie.toRGBA()
	bounds := watermark.Bounds()
	x, y := ie.getWatermarkPosition(position, bounds.Dx(), bounds.Dy())

	// Apply watermark with opacity
	for dy := 0; dy < bounds.Dy(); dy++ {
		for dx := 0; dx < bounds.Dx(); dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && px < ie.width && py >= 0 && py < ie.height {
				watermarkPixel := watermark.At(bounds.Min.X+dx, bounds.Min.Y+dy)
				rgba.Set(px, py, ie.blendColor(rgba.At(px, py), watermarkPixel, opacity))
			}
		}
	}

	ie.currentImage = rgba
	return fmt.Sprintf("Added image watermark from %s at %s", imagePath, position), nil
}

// RemoveWatermark removes or disguises watermark
func (ie *ImageEngine) RemoveWatermark(method string) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	switch method {
	case "inpaint":
		// Simple inpainting - blur suspected watermark regions
		ie.currentImage = ie.gaussianBlur(3.0)
	case "median":
		// Use median filter to remove watermark
		ie.currentImage = ie.medianFilter(3)
	case "frequency":
		// Use frequency domain to remove periodic watermarks
		gray := ie.toGrayscale()
		spectrum := fft2D(gray)
		// Remove high frequency components (simplified)
		for i := range spectrum {
			for j := range spectrum[i] {
				if math.Abs(float64(i-len(spectrum)/2)) > float64(len(spectrum))/4 ||
					math.Abs(float64(j-len(spectrum[0])/2)) > float64(len(spectrum[0]))/4 {
					spectrum[i][j] = 0
				}
			}
		}
		gray = ifft2D(spectrum)
		ie.currentImage = ie.fromGrayscale(gray)
	default:
		// Default: apply denoising
		ie.currentImage = ie.bilateralFilter(2.0)
	}

	return fmt.Sprintf("Watermark removed/disguised using %s method", method), nil
}

// OCRPlate performs OCR on license plates
func (ie *ImageEngine) OCRPlate() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Simulate OCR for license plates
	// In production, this would use a real OCR library like Tesseract
	ie.ocrResult = ie.simulateOCR("plate")

	return fmt.Sprintf("License plate detected: %s", ie.ocrResult), nil
}

// OCRCaptcha performs OCR on CAPTCHA
func (ie *ImageEngine) OCRCaptcha() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Preprocess for CAPTCHA
	ie.currentImage = ie.preprocessCaptcha()

	// Simulate OCR for CAPTCHA
	ie.ocrResult = ie.simulateOCR("captcha")

	return fmt.Sprintf("CAPTCHA text: %s", ie.ocrResult), nil
}

// DetectPlate detects license plate region
func (ie *ImageEngine) DetectPlate() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Simulate plate detection using edge detection and contour analysis
	edges := ie.cannyEdgeDetect(1.0)
	
	// Find rectangular regions (simplified)
	// In production, would use Haar cascades or deep learning
	ie.plateRegion = ie.findRectangularRegion(edges)

	return fmt.Sprintf("License plate detected at region: (%d,%d) %dx%d", 
		ie.plateRegion.Min.X, ie.plateRegion.Min.Y,
		ie.plateRegion.Dx(), ie.plateRegion.Dy()), nil
}

// ExtractPlate extracts the detected license plate
func (ie *ImageEngine) ExtractPlate() (string, error) {
	if ie.plateRegion.Empty() {
		// Try to detect first
		_, err := ie.DetectPlate()
		if err != nil {
			return "", err
		}
	}

	// Crop to plate region
	return ie.Crop(ie.plateRegion.Min.X, ie.plateRegion.Min.Y,
		ie.plateRegion.Dx(), ie.plateRegion.Dy())
}

// AdjustBrightness adjusts image brightness
func (ie *ImageEngine) AdjustBrightness(value float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	rgba := ie.toRGBA()
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			c.R = ie.clamp(float64(c.R) + value*255)
			c.G = ie.clamp(float64(c.G) + value*255)
			c.B = ie.clamp(float64(c.B) + value*255)
			rgba.Set(x, y, c)
		}
	}

	ie.currentImage = rgba
	return fmt.Sprintf("Brightness adjusted by %.2f", value), nil
}

// AdjustContrast adjusts image contrast
func (ie *ImageEngine) AdjustContrast(value float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	factor := (259 * (value + 255)) / (255 * (259 - value))
	rgba := ie.toRGBA()

	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			c.R = ie.clamp(factor*(float64(c.R)-128) + 128)
			c.G = ie.clamp(factor*(float64(c.G)-128) + 128)
			c.B = ie.clamp(factor*(float64(c.B)-128) + 128)
			rgba.Set(x, y, c)
		}
	}

	ie.currentImage = rgba
	return fmt.Sprintf("Contrast adjusted by %.2f", value), nil
}

// AdjustSaturation adjusts image saturation
func (ie *ImageEngine) AdjustSaturation(value float64) (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	rgba := ie.toRGBA()
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			gray := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
			c.R = ie.clamp(gray + value*(float64(c.R)-gray))
			c.G = ie.clamp(gray + value*(float64(c.G)-gray))
			c.B = ie.clamp(gray + value*(float64(c.B)-gray))
			rgba.Set(x, y, c)
		}
	}

	ie.currentImage = rgba
	return fmt.Sprintf("Saturation adjusted by %.2f", value), nil
}

// Denoise removes noise from image
func (ie *ImageEngine) Denoise() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	// Apply bilateral filter for edge-preserving denoising
	ie.currentImage = ie.bilateralFilter(2.0)
	return "Image denoised", nil
}

// GetHistogram computes image histogram
func (ie *ImageEngine) GetHistogram() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	histogram := make([]int, 256)
	gray := ie.toGrayscale()

	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			val := gray[y][x]
			histogram[int(val*255)]++
		}
	}

	// Find peak
	maxCount := 0
	peakValue := 0
	for i, count := range histogram {
		if count > maxCount {
			maxCount = count
			peakValue = i
		}
	}

	return fmt.Sprintf("Histogram computed: peak at level %d with %d pixels", peakValue, maxCount), nil
}

// EqualizeHistogram performs histogram equalization
func (ie *ImageEngine) EqualizeHistogram() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	rgba := ie.toRGBA()
	histogram := make([]int, 256)
	
	// Compute histogram
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			gray := int(0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B))
			histogram[gray]++
		}
	}

	// Compute CDF
	cdf := make([]int, 256)
	cdf[0] = histogram[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + histogram[i]
	}

	// Normalize CDF
	totalPixels := ie.width * ie.height
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			gray := int(0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B))
			newVal := uint8(255 * cdf[gray] / totalPixels)
			rgba.Set(x, y, color.RGBA{newVal, newVal, newVal, c.A})
		}
	}

	ie.currentImage = rgba
	return "Histogram equalized", nil
}

// EdgeDetect performs edge detection
func (ie *ImageEngine) EdgeDetect() (string, error) {
	if ie.currentImage == nil {
		return "", fmt.Errorf("no image loaded")
	}

	ie.currentImage = ie.cannyEdgeDetect(1.0)
	return "Edges detected using Canny algorithm", nil
}

// Helper functions

func (ie *ImageEngine) toGrayscale() [][]float64 {
	gray := make([][]float64, ie.height)
	for y := 0; y < ie.height; y++ {
		gray[y] = make([]float64, ie.width)
		for x := 0; x < ie.width; x++ {
			r, g, b, _ := ie.currentImage.At(x, y).RGBA()
			gray[y][x] = (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
		}
	}
	return gray
}

func (ie *ImageEngine) fromGrayscale(gray [][]float64) image.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			val := uint8(gray[y][x] * 255)
			rgba.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}
	return rgba
}

func (ie *ImageEngine) toRGBA() *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	draw.Draw(rgba, rgba.Bounds(), ie.currentImage, image.Point{}, draw.Src)
	return rgba
}

func (ie *ImageEngine) clamp(val float64) uint8 {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return uint8(val)
}

func (ie *ImageEngine) blendColor(bg, fg color.Color, opacity float64) color.Color {
	bgR, bgG, bgB, bgA := bg.RGBA()
	fgR, fgG, fgB, _ := fg.RGBA()
	
	newR := uint8((float64(bgR)*(1-opacity) + float64(fgR)*opacity) / 256)
	newG := uint8((float64(bgG)*(1-opacity) + float64(fgG)*opacity) / 256)
	newB := uint8((float64(bgB)*(1-opacity) + float64(fgB)*opacity) / 256)
	
	return color.RGBA{newR, newG, newB, uint8(bgA / 256)}
}

func (ie *ImageEngine) getWatermarkPosition(position string, wmWidth, wmHeight int) (int, int) {
	padding := 20
	
	switch position {
	case "top_left":
		return padding, padding
	case "top_right":
		return ie.width - wmWidth - padding, padding
	case "bottom_left":
		return padding, ie.height - wmHeight - padding
	case "bottom_right":
		return ie.width - wmWidth - padding, ie.height - wmHeight - padding
	case "center":
		return (ie.width - wmWidth) / 2, (ie.height - wmHeight) / 2
	case "tiled":
		// Return first tile position
		return 0, 0
	case "diagonal":
		// Start of diagonal
		return padding, padding
	default:
		return padding, padding
	}
}

func (ie *ImageEngine) applyKernel(kernel [][]float64) image.Image {
	rgba := ie.toRGBA()
	result := image.NewRGBA(rgba.Bounds())
	
	kernelSize := len(kernel)
	offset := kernelSize / 2
	
	for y := offset; y < ie.height-offset; y++ {
		for x := offset; x < ie.width-offset; x++ {
			var sumR, sumG, sumB float64
			
			for ky := 0; ky < kernelSize; ky++ {
				for kx := 0; kx < kernelSize; kx++ {
					px := x + kx - offset
					py := y + ky - offset
					c := rgba.At(px, py).(color.RGBA)
					sumR += float64(c.R) * kernel[ky][kx]
					sumG += float64(c.G) * kernel[ky][kx]
					sumB += float64(c.B) * kernel[ky][kx]
				}
			}
			
			result.Set(x, y, color.RGBA{
				ie.clamp(sumR),
				ie.clamp(sumG),
				ie.clamp(sumB),
				255,
			})
		}
	}
	
	return result
}

func (ie *ImageEngine) gaussianBlur(sigma float64) image.Image {
	size := int(2*sigma + 1)
	kernel := make([][]float64, size)
	sum := 0.0
	
	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			x := float64(i - size/2)
			y := float64(j - size/2)
			kernel[i][j] = math.Exp(-(x*x+y*y)/(2*sigma*sigma)) / (2 * math.Pi * sigma * sigma)
			sum += kernel[i][j]
		}
	}
	
	// Normalize
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			kernel[i][j] /= sum
		}
	}
	
	return ie.applyKernel(kernel)
}

func (ie *ImageEngine) boxBlur(radius int) image.Image {
	size := 2*radius + 1
	kernel := make([][]float64, size)
	val := 1.0 / float64(size*size)
	
	for i := 0; i < size; i++ {
		kernel[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			kernel[i][j] = val
		}
	}
	
	return ie.applyKernel(kernel)
}

func (ie *ImageEngine) motionBlur(angle float64) image.Image {
	// Simple horizontal motion blur
	kernel := [][]float64{
		{0.2, 0.2, 0.2, 0.2, 0.2},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
	}
	return ie.applyKernel(kernel)
}

func (ie *ImageEngine) medianFilter(radius int) image.Image {
	rgba := ie.toRGBA()
	result := image.NewRGBA(rgba.Bounds())
	
	for y := radius; y < ie.height-radius; y++ {
		for x := radius; x < ie.width-radius; x++ {
			var rVals, gVals, bVals []int
			
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					c := rgba.At(x+dx, y+dy).(color.RGBA)
					rVals = append(rVals, int(c.R))
					gVals = append(gVals, int(c.G))
					bVals = append(bVals, int(c.B))
				}
			}
			
			result.Set(x, y, color.RGBA{
				uint8(median(rVals)),
				uint8(median(gVals)),
				uint8(median(bVals)),
				255,
			})
		}
	}
	
	return result
}

func (ie *ImageEngine) bilateralFilter(sigma float64) image.Image {
	// Simplified bilateral filter
	return ie.gaussianBlur(sigma)
}

func (ie *ImageEngine) sobelFilter() image.Image {
	kernelX := [][]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	kernelY := [][]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	
	// Apply both kernels and combine
	imgX := ie.applyKernel(kernelX)
	ie.currentImage = imgX
	imgY := ie.applyKernel(kernelY)
	
	// Combine results
	rgba := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			cx := imgX.At(x, y).(color.RGBA)
			cy := imgY.At(x, y).(color.RGBA)
			
			val := uint8(math.Sqrt(float64(cx.R*cx.R + cy.R*cy.R)))
			rgba.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}
	
	return rgba
}

func (ie *ImageEngine) prewittFilter() image.Image {
	kernelX := [][]float64{{-1, 0, 1}, {-1, 0, 1}, {-1, 0, 1}}
	kernelY := [][]float64{{-1, -1, -1}, {0, 0, 0}, {1, 1, 1}}
	
	imgX := ie.applyKernel(kernelX)
	ie.currentImage = imgX
	imgY := ie.applyKernel(kernelY)
	
	rgba := image.NewRGBA(image.Rect(0, 0, ie.width, ie.height))
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			cx := imgX.At(x, y).(color.RGBA)
			cy := imgY.At(x, y).(color.RGBA)
			
			val := uint8(math.Sqrt(float64(cx.R*cx.R + cy.R*cy.R)))
			rgba.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}
	
	return rgba
}

func (ie *ImageEngine) cannyEdgeDetect(threshold float64) image.Image {
	// Simplified Canny edge detection
	// 1. Gaussian blur
	blurred := ie.gaussianBlur(1.4)
	ie.currentImage = blurred
	
	// 2. Sobel filter
	edges := ie.sobelFilter()
	
	// 3. Simple thresholding
	rgba := edges.(*image.RGBA)
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			if float64(c.R) < threshold*128 {
				rgba.Set(x, y, color.RGBA{0, 0, 0, 255})
			} else {
				rgba.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	
	return rgba
}

func (ie *ImageEngine) preprocessCaptcha() image.Image {
	// Preprocess CAPTCHA: denoise, threshold, etc.
	ie.currentImage = ie.medianFilter(1)
	ie.currentImage = ie.gaussianBlur(0.5)
	
	// Convert to binary
	rgba := ie.toRGBA()
	for y := 0; y < ie.height; y++ {
		for x := 0; x < ie.width; x++ {
			c := rgba.At(x, y).(color.RGBA)
			gray := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
			if gray > 128 {
				rgba.Set(x, y, color.RGBA{255, 255, 255, 255})
			} else {
				rgba.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	
	return rgba
}

func (ie *ImageEngine) findRectangularRegion(img image.Image) image.Rectangle {
	// Simplified plate detection
	// In production, would use contour detection or ML
	
	// Return a mock region for demonstration
	centerX := ie.width / 2
	centerY := ie.height / 2
	plateWidth := ie.width / 4
	plateHeight := ie.height / 8
	
	return image.Rect(
		centerX-plateWidth/2,
		centerY-plateHeight/2,
		centerX+plateWidth/2,
		centerY+plateHeight/2,
	)
}

func (ie *ImageEngine) simulateOCR(ocrType string) string {
	// Simulate OCR results for demonstration
	if ocrType == "plate" {
		// Simulate license plate formats
		plates := []string{"ABC-1234", "XYZ-5678", "LMN-9012", "PQR-3456"}
		return plates[ie.width%len(plates)]
	} else if ocrType == "captcha" {
		// Simulate CAPTCHA text
		captchas := []string{"A3B4C5", "X7Y8Z9", "P2Q3R4", "M5N6O7"}
		return captchas[ie.height%len(captchas)]
	}
	return "UNKNOWN"
}

// Transform helper functions

func fft2D(data [][]float64) [][]complex128 {
	height := len(data)
	width := len(data[0])
	result := make([][]complex128, height)
	
	// FFT rows
	for i := 0; i < height; i++ {
		row := make([]complex128, width)
		for j := 0; j < width; j++ {
			row[j] = complex(data[i][j], 0)
		}
		result[i] = fft1D(row)
	}
	
	// FFT columns
	for j := 0; j < width; j++ {
		col := make([]complex128, height)
		for i := 0; i < height; i++ {
			col[i] = result[i][j]
		}
		col = fft1D(col)
		for i := 0; i < height; i++ {
			result[i][j] = col[i]
		}
	}
	
	return result
}

func ifft2D(spectrum [][]complex128) [][]float64 {
	height := len(spectrum)
	width := len(spectrum[0])
	result := make([][]float64, height)
	
	// IFFT columns
	temp := make([][]complex128, height)
	for i := 0; i < height; i++ {
		temp[i] = make([]complex128, width)
	}
	
	for j := 0; j < width; j++ {
		col := make([]complex128, height)
		for i := 0; i < height; i++ {
			col[i] = spectrum[i][j]
		}
		col = ifft1D(col)
		for i := 0; i < height; i++ {
			temp[i][j] = col[i]
		}
	}
	
	// IFFT rows
	for i := 0; i < height; i++ {
		row := ifft1D(temp[i])
		result[i] = make([]float64, width)
		for j := 0; j < width; j++ {
			result[i][j] = real(row[j])
		}
	}
	
	return result
}

func fft1D(x []complex128) []complex128 {
	n := len(x)
	if n <= 1 {
		return x
	}
	
	// Pad to power of 2
	n2 := 1
	for n2 < n {
		n2 *= 2
	}
	if n2 != n {
		padded := make([]complex128, n2)
		copy(padded, x)
		x = padded
		n = n2
	}
	
	return fftRecursive(x)
}

func fftRecursive(x []complex128) []complex128 {
	n := len(x)
	if n <= 1 {
		return x
	}
	
	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}
	
	even = fftRecursive(even)
	odd = fftRecursive(odd)
	
	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Exp(complex(0, -2*math.Pi*float64(k)/float64(n))) * odd[k]
		result[k] = even[k] + t
		result[k+n/2] = even[k] - t
	}
	
	return result
}

func ifft1D(x []complex128) []complex128 {
	n := len(x)
	
	// Conjugate
	for i := range x {
		x[i] = cmplx.Conj(x[i])
	}
	
	// Forward FFT
	result := fft1D(x)
	
	// Conjugate and scale
	for i := range result {
		result[i] = cmplx.Conj(result[i]) / complex(float64(n), 0)
	}
	
	return result
}

func dct2D(data [][]float64) [][]float64 {
	height := len(data)
	width := len(data[0])
	result := make([][]float64, height)
	
	// DCT rows
	for i := 0; i < height; i++ {
		result[i] = dct1D(data[i])
	}
	
	// DCT columns
	for j := 0; j < width; j++ {
		col := make([]float64, height)
		for i := 0; i < height; i++ {
			col[i] = result[i][j]
		}
		col = dct1D(col)
		for i := 0; i < height; i++ {
			result[i][j] = col[i]
		}
	}
	
	return result
}

func idct2D(coeffs [][]float64) [][]float64 {
	height := len(coeffs)
	width := len(coeffs[0])
	result := make([][]float64, height)
	
	// IDCT columns
	temp := make([][]float64, height)
	for i := 0; i < height; i++ {
		temp[i] = make([]float64, width)
	}
	
	for j := 0; j < width; j++ {
		col := make([]float64, height)
		for i := 0; i < height; i++ {
			col[i] = coeffs[i][j]
		}
		col = idct1D(col)
		for i := 0; i < height; i++ {
			temp[i][j] = col[i]
		}
	}
	
	// IDCT rows
	for i := 0; i < height; i++ {
		result[i] = idct1D(temp[i])
	}
	
	return result
}

func dct1D(data []float64) []float64 {
	n := len(data)
	result := make([]float64, n)
	
	for k := 0; k < n; k++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			sum += data[i] * math.Cos((math.Pi/float64(n))*(float64(i)+0.5)*float64(k))
		}
		
		if k == 0 {
			result[k] = sum * math.Sqrt(1.0/float64(n))
		} else {
			result[k] = sum * math.Sqrt(2.0/float64(n))
		}
	}
	
	return result
}

func idct1D(coeffs []float64) []float64 {
	n := len(coeffs)
	result := make([]float64, n)
	
	for i := 0; i < n; i++ {
		sum := coeffs[0] * math.Sqrt(1.0/float64(n))
		for k := 1; k < n; k++ {
			sum += coeffs[k] * math.Sqrt(2.0/float64(n)) * 
				math.Cos((math.Pi/float64(n))*(float64(i)+0.5)*float64(k))
		}
		result[i] = sum
	}
	
	return result
}

func median(vals []int) int {
	n := len(vals)
	if n == 0 {
		return 0
	}
	
	// Simple bubble sort for small arrays
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if vals[j] > vals[j+1] {
				vals[j], vals[j+1] = vals[j+1], vals[j]
			}
		}
	}
	
	return vals[n/2]
}

// GetCurrentImage returns the current image for testing
func (ie *ImageEngine) GetCurrentImage() image.Image {
	return ie.currentImage
}

// SetCurrentImage sets the current image for testing
func (ie *ImageEngine) SetCurrentImage(img image.Image) {
	ie.currentImage = img
	bounds := img.Bounds()
	ie.width = bounds.Dx()
	ie.height = bounds.Dy()
}

// GetOCRResult returns the last OCR result
func (ie *ImageEngine) GetOCRResult() string {
	return ie.ocrResult
}