package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/arturoeanton/go-dsl/examples/image_dsl/universal"
)

func main() {
	fmt.Println("Image DSL - Comprehensive Image Processing")
	fmt.Println("=" + strings.Repeat("=", 45))

	// Create DSL instance
	imageDSL := universal.NewImageDSL()

	// Generate test images
	generateTestImage("test_input.png")
	generateLicensePlateImage("test_plate.jpg")
	generateCaptchaImage("test_captcha.png")
	generateWatermarkImage("watermark.png")

	// Example commands demonstrating all features
	commands := []string{
		// Basic operations
		`load "test_input.png"`,
		`save "test_output.png" quality 95`,

		// Transforms
		`fft`,
		`ifft`,
		`dct`,
		`idct`,
		`laplace`,
		`inverse_laplace`,

		// Image editing
		`resize width 640 height 480`,
		`resize 75%`,
		`rotate 45 degrees`,
		`flip horizontal`,
		`flip vertical`,
		`crop x 50 y 50 width 200 height 200`,

		// Filters
		`filter gaussian strength 2.5`,
		`filter median strength 3`,
		`filter bilateral strength 2`,
		`filter sobel`,
		`filter prewitt`,
		`filter canny strength 1.5`,
		`filter laplacian`,
		`filter sharpen`,
		`filter emboss`,
		`filter blur strength 3`,
		`filter motion_blur strength 5`,

		// Watermarking
		`watermark text "COPYRIGHT 2024" position top_right opacity 0.5`,
		`watermark text "WATERMARK" position center opacity 0.3`,
		`watermark text "DIAGONAL" position diagonal opacity 0.4`,
		`watermark image "watermark.png" position bottom_right opacity 0.6`,
		`remove_watermark`,
		`remove_watermark method "inpaint"`,
		`remove_watermark method "frequency"`,

		// License plate operations
		`load "test_plate.jpg"`,
		`detect_plate`,
		`extract_plate`,
		`ocr_plate`,

		// CAPTCHA operations
		`load "test_captcha.png"`,
		`ocr_captcha`,

		// Image adjustments
		`brightness 0.2`,
		`brightness -0.1`,
		`contrast 1.5`,
		`saturation 1.2`,
		`denoise`,

		// Analysis
		`histogram`,
		`equalize`,
		`edge_detect`,

		// Compression
		`compress type jpeg quality 85`,
		`compress type png quality 100`,
		`compress type webp quality 90`,
	}

	fmt.Println("\nExecuting Image DSL Commands:")
	fmt.Println("-" + strings.Repeat("-", 45))

	successCount := 0
	failCount := 0

	for _, cmd := range commands {
		fmt.Printf("\n> %s\n", cmd)
		result, err := imageDSL.Parse(cmd)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			failCount++
		} else if result != nil {
			fmt.Printf("  SUCCESS: %v\n", result)
			successCount++
		}
	}

	fmt.Printf("\n\nTest Summary:")
	fmt.Printf("\n  Total commands: %d", len(commands))
	fmt.Printf("\n  Successful: %d", successCount)
	fmt.Printf("\n  Failed: %d", failCount)
	fmt.Printf("\n  Success rate: %.1f%%\n", float64(successCount)/float64(len(commands))*100)

	// Demonstrate complex workflow
	fmt.Println("\n\nComplex Workflow Example:")
	fmt.Println("-" + strings.Repeat("-", 45))

	workflow := []struct {
		description string
		command     string
	}{
		{"Load high-resolution image", `load "test_input.png"`},
		{"Apply Gaussian blur for smoothing", `filter gaussian strength 1.5`},
		{"Enhance edges", `filter sharpen`},
		{"Adjust brightness", `brightness 0.1`},
		{"Increase contrast", `contrast 1.3`},
		{"Add watermark", `watermark text "PROCESSED" position bottom_left opacity 0.4`},
		{"Save processed image", `save "final_output.jpg" quality 90`},
	}

	for _, step := range workflow {
		fmt.Printf("\n%s:\n  > %s\n", step.description, step.command)
		result, err := imageDSL.Parse(step.command)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if result != nil {
			fmt.Printf("  RESULT: %v\n", result)
		}
	}

	// OCR Demonstration
	fmt.Println("\n\nOCR Demonstration:")
	fmt.Println("-" + strings.Repeat("-", 45))

	ocrCommands := []struct {
		description string
		command     string
	}{
		{"Process license plate", `load "test_plate.jpg"`},
		{"Detect plate region", `detect_plate`},
		{"Extract plate", `extract_plate`},
		{"Read plate text", `ocr_plate`},
		{"Process CAPTCHA", `load "test_captcha.png"`},
		{"Read CAPTCHA text", `ocr_captcha`},
	}

	for _, step := range ocrCommands {
		fmt.Printf("\n%s:\n  > %s\n", step.description, step.command)
		result, err := imageDSL.Parse(step.command)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if result != nil {
			fmt.Printf("  RESULT: %v\n", result)
		}
	}

	fmt.Println("\n\nDemo completed successfully!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - test_input.png (test image)")
	fmt.Println("  - test_plate.jpg (license plate)")
	fmt.Println("  - test_captcha.png (CAPTCHA)")
	fmt.Println("  - watermark.png (watermark)")
	fmt.Println("  - test_output.png (processed)")
	fmt.Println("  - final_output.jpg (workflow result)")

	// Clean up test files (optional)
	// cleanupTestFiles()
}

// generateTestImage creates a test image
func generateTestImage(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))

	// Create gradient background
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			r := uint8(x * 255 / 800)
			g := uint8(y * 255 / 600)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Add some shapes
	// Rectangle
	for y := 100; y < 200; y++ {
		for x := 100; x < 300; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	// Circle (approximate)
	centerX, centerY := 600, 300
	radius := 80
	for y := centerY - radius; y < centerY+radius; y++ {
		for x := centerX - radius; x < centerX+radius; x++ {
			dx := x - centerX
			dy := y - centerY
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, color.RGBA{0, 255, 0, 255})
			}
		}
	}

	saveImage(img, filename)
	fmt.Printf("Generated %s\n", filename)
}

// generateLicensePlateImage creates a simulated license plate image
func generateLicensePlateImage(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))

	// Background
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}

	// License plate area (white rectangle)
	for y := 100; y < 200; y++ {
		for x := 75; x < 325; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	// Simulate text on plate (black blocks)
	textPositions := []struct{ x, y, w, h int }{
		{90, 120, 30, 60},   // A
		{130, 120, 30, 60},  // B
		{170, 120, 30, 60},  // C
		{220, 130, 20, 50},  // 1
		{250, 130, 20, 50},  // 2
		{280, 130, 20, 50},  // 3
		{310, 130, 20, 50},  // 4
	}

	for _, pos := range textPositions {
		for y := pos.y; y < pos.y+pos.h; y++ {
			for x := pos.x; x < pos.x+pos.w; x++ {
				if (x-pos.x)%10 < 7 && (y-pos.y)%10 < 7 {
					img.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
		}
	}

	saveImage(img, filename)
	fmt.Printf("Generated %s\n", filename)
}

// generateCaptchaImage creates a simulated CAPTCHA image
func generateCaptchaImage(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 80))

	// Noisy background
	for y := 0; y < 80; y++ {
		for x := 0; x < 200; x++ {
			noise := uint8((x*y + x + y) % 50)
			img.Set(x, y, color.RGBA{200 + noise, 200 + noise, 200 + noise, 255})
		}
	}

	// Simulated CAPTCHA text
	textBlocks := []struct{ x, y, size int }{
		{20, 20, 30},
		{60, 25, 28},
		{95, 18, 32},
		{130, 22, 30},
		{165, 20, 29},
	}

	for _, block := range textBlocks {
		for dy := 0; dy < block.size; dy++ {
			for dx := 0; dx < block.size/2; dx++ {
				if (dx+dy)%3 == 0 {
					img.Set(block.x+dx, block.y+dy, color.RGBA{0, 0, 100, 255})
				}
			}
		}
	}

	// Add distortion lines
	for x := 0; x < 200; x++ {
		y := 40 + int(10*float64(x%30)/30)
		if y >= 0 && y < 80 {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}

	saveImage(img, filename)
	fmt.Printf("Generated %s\n", filename)
}

// generateWatermarkImage creates a watermark image
func generateWatermarkImage(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, 150, 50))

	// Transparent background
	for y := 0; y < 50; y++ {
		for x := 0; x < 150; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 0})
		}
	}

	// Simple watermark pattern
	for y := 10; y < 40; y++ {
		for x := 10; x < 140; x++ {
			if (x-10)%20 < 15 && (y-10)%20 < 15 {
				img.Set(x, y, color.RGBA{0, 0, 0, 128})
			}
		}
	}

	saveImage(img, filename)
	fmt.Printf("Generated %s\n", filename)
}

// saveImage saves an image to file
func saveImage(img image.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	if strings.HasSuffix(filename, ".png") {
		png.Encode(file, img)
	} else if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
		// For simplicity, save as PNG even for JPG extension
		png.Encode(file, img)
	}
}

// cleanupTestFiles removes generated test files
func cleanupTestFiles() {
	files := []string{
		"test_input.png",
		"test_plate.jpg",
		"test_captcha.png",
		"watermark.png",
		"test_output.png",
		"final_output.jpg",
	}

	for _, file := range files {
		os.Remove(file)
	}
}