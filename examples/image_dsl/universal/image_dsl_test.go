package universal

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
)

// TestImageDSLCreation tests DSL creation
func TestImageDSLCreation(t *testing.T) {
	dsl := NewImageDSL()
	if dsl == nil {
		t.Fatal("Failed to create ImageDSL")
	}
	if dsl.engine == nil {
		t.Fatal("Engine not initialized")
	}
}

// TestLoadCommand tests load command parsing
func TestLoadCommand(t *testing.T) {
	dsl := NewImageDSL()
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Load image", `load "test.jpg"`, false},
		{"Load with spaces", `load "my image.png"`, false},
		{"Invalid load", `load`, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestSaveCommand tests save command parsing
func TestSaveCommand(t *testing.T) {
	dsl := NewImageDSL()
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Save simple", `save "output.jpg"`, false},
		{"Save with quality", `save "output.jpg" quality 85`, false},
		{"Invalid save", `save`, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestTransformCommands tests transform commands
func TestTransformCommands(t *testing.T) {
	dsl := NewImageDSL()
	
	// Create a test image
	testImg := createTestImage()
	dsl.engine.SetCurrentImage(testImg)
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"FFT", "fft", false},
		{"IFFT", "ifft", false},
		{"DCT", "dct", false},
		{"IDCT", "idct", false},
		{"Laplace", "laplace", false},
		{"Inverse Laplace", "inverse_laplace", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
		})
	}
}

// TestEditCommands tests edit commands
func TestEditCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Resize", "resize width 100 height 100", false},
		{"Resize percent", "resize 50%", false},
		{"Rotate", "rotate 45 degrees", false},
		{"Flip horizontal", "flip horizontal", false},
		{"Flip vertical", "flip vertical", false},
		{"Crop", "crop x 10 y 10 width 50 height 50", false},
		{"Compress JPEG", "compress type jpeg quality 85", false},
		{"Compress PNG", "compress type png quality 100", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
		})
	}
}

// TestFilterCommands tests filter commands
func TestFilterCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	filters := []string{
		"gaussian", "median", "bilateral", "sobel", "prewitt",
		"canny", "laplacian", "sharpen", "emboss", "blur", "motion_blur",
	}
	
	for _, filter := range filters {
		t.Run("Filter_"+filter, func(t *testing.T) {
			input := "filter " + filter
			result, err := dsl.Parse(input)
			if err != nil {
				t.Errorf("Parse(%q) error = %v", input, err)
			}
			if result == nil {
				t.Errorf("Parse(%q) returned nil result", input)
			}
		})
		
		t.Run("Filter_"+filter+"_strength", func(t *testing.T) {
			input := "filter " + filter + " strength 2.5"
			result, err := dsl.Parse(input)
			if err != nil {
				t.Errorf("Parse(%q) error = %v", input, err)
			}
			if result == nil {
				t.Errorf("Parse(%q) returned nil result", input)
			}
		})
	}
}

// TestWatermarkCommands tests watermark commands
func TestWatermarkCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	positions := []string{
		"top_left", "top_right", "bottom_left", "bottom_right",
		"center", "tiled", "diagonal",
	}
	
	for _, pos := range positions {
		t.Run("Watermark_text_"+pos, func(t *testing.T) {
			input := `watermark text "WATERMARK" position ` + pos
			result, err := dsl.Parse(input)
			if err != nil {
				t.Errorf("Parse(%q) error = %v", input, err)
			}
			if result == nil {
				t.Errorf("Parse(%q) returned nil result", input)
			}
		})
		
		t.Run("Watermark_text_opacity_"+pos, func(t *testing.T) {
			input := `watermark text "WATERMARK" position ` + pos + ` opacity 0.5`
			result, err := dsl.Parse(input)
			if err != nil {
				t.Errorf("Parse(%q) error = %v", input, err)
			}
			if result == nil {
				t.Errorf("Parse(%q) returned nil result", input)
			}
		})
	}
	
	// Test watermark removal
	t.Run("Remove_watermark", func(t *testing.T) {
		input := "remove_watermark"
		result, err := dsl.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
		if result == nil {
			t.Errorf("Parse(%q) returned nil result", input)
		}
	})
	
	t.Run("Remove_watermark_method", func(t *testing.T) {
		input := `remove_watermark method "inpaint"`
		result, err := dsl.Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
		if result == nil {
			t.Errorf("Parse(%q) returned nil result", input)
		}
	})
}

// TestOCRCommands tests OCR commands
func TestOCRCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"OCR plate", "ocr_plate", false},
		{"OCR CAPTCHA", "ocr_captcha", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
			
			// Check OCR result
			if !tt.wantErr {
				ocrResult := dsl.engine.GetOCRResult()
				if ocrResult == "" {
					t.Error("OCR returned empty result")
				}
			}
		})
	}
}

// TestDetectionCommands tests detection commands
func TestDetectionCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Detect plate", "detect_plate", false},
		{"Extract plate", "extract_plate", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
		})
	}
}

// TestAdjustmentCommands tests adjustment commands
func TestAdjustmentCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Brightness up", "brightness 0.5", false},
		{"Brightness down", "brightness -0.3", false},
		{"Contrast increase", "contrast 1.5", false},
		{"Saturation increase", "saturation 1.2", false},
		{"Denoise", "denoise", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
		})
	}
}

// TestAnalysisCommands tests analysis commands
func TestAnalysisCommands(t *testing.T) {
	dsl := NewImageDSL()
	dsl.engine.SetCurrentImage(createTestImage())
	
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Histogram", "histogram", false},
		{"Equalize", "equalize", false},
		{"Edge detect", "edge_detect", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsl.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Parse(%q) returned nil result", tt.input)
			}
		})
	}
}

// TestComplexWorkflow tests a complex workflow
func TestComplexWorkflow(t *testing.T) {
	dsl := NewImageDSL()
	
	// Create test image file
	testFile := "test_image.png"
	img := createTestImage()
	saveTestImage(t, img, testFile)
	defer os.Remove(testFile)
	
	// Test workflow
	workflow := []string{
		`load "test_image.png"`,
		"resize 50%",
		"filter gaussian strength 2",
		"brightness 0.2",
		"contrast 1.2",
		`watermark text "TEST" position center opacity 0.3`,
		"detect_plate",
		"edge_detect",
		`save "output.png" quality 90`,
	}
	
	for i, cmd := range workflow {
		t.Run(strings.ReplaceAll(cmd, " ", "_"), func(t *testing.T) {
			result, err := dsl.Parse(cmd)
			if err != nil {
				t.Errorf("Step %d: Parse(%q) error = %v", i, cmd, err)
			}
			if result == nil && !strings.HasPrefix(cmd, "save") {
				t.Errorf("Step %d: Parse(%q) returned nil result", i, cmd)
			}
		})
	}
	
	// Clean up output file
	os.Remove("output.png")
}

// TestImageEngineCore tests core engine functionality
func TestImageEngineCore(t *testing.T) {
	engine := NewImageEngine()
	
	// Test with no image loaded
	_, err := engine.FFT()
	if err == nil {
		t.Error("FFT should fail with no image loaded")
	}
	
	// Set test image
	img := createTestImage()
	engine.SetCurrentImage(img)
	
	// Test FFT
	result, err := engine.FFT()
	if err != nil {
		t.Errorf("FFT failed: %v", err)
	}
	if !strings.Contains(result, "FFT computed") {
		t.Errorf("Unexpected FFT result: %s", result)
	}
	
	// Test IFFT
	result, err = engine.IFFT()
	if err != nil {
		t.Errorf("IFFT failed: %v", err)
	}
	if !strings.Contains(result, "IFFT computed") {
		t.Errorf("Unexpected IFFT result: %s", result)
	}
	
	// Test DCT
	result, err = engine.DCT()
	if err != nil {
		t.Errorf("DCT failed: %v", err)
	}
	if !strings.Contains(result, "DCT computed") {
		t.Errorf("Unexpected DCT result: %s", result)
	}
	
	// Test IDCT
	result, err = engine.IDCT()
	if err != nil {
		t.Errorf("IDCT failed: %v", err)
	}
	if !strings.Contains(result, "IDCT computed") {
		t.Errorf("Unexpected IDCT result: %s", result)
	}
}

// TestImageFormats tests different image format support
func TestImageFormats(t *testing.T) {
	engine := NewImageEngine()
	img := createTestImage()
	engine.SetCurrentImage(img)
	
	formats := []struct {
		ext     string
		quality int
	}{
		{".jpg", 85},
		{".jpeg", 90},
		{".png", 100},
		{".gif", 100},
		{".bmp", 100},
	}
	
	for _, format := range formats {
		filename := "test" + format.ext
		t.Run("Format_"+format.ext, func(t *testing.T) {
			result, err := engine.SaveImage(filename, format.quality)
			if err != nil {
				t.Errorf("SaveImage failed for %s: %v", format.ext, err)
			}
			if !strings.Contains(result, "Saved") {
				t.Errorf("Unexpected save result for %s: %s", format.ext, result)
			}
			
			// Clean up
			os.Remove(filename)
		})
	}
}

// TestParseError tests parse error handling
func TestParseError(t *testing.T) {
	dsl := NewImageDSL()
	
	invalidCommands := []string{
		"invalid command",
		"load",
		"save",
		"filter unknown_filter",
		"resize invalid",
	}
	
	for _, cmd := range invalidCommands {
		t.Run("Invalid_"+strings.ReplaceAll(cmd, " ", "_"), func(t *testing.T) {
			_, err := dsl.Parse(cmd)
			if err == nil {
				t.Errorf("Expected error for invalid command: %s", cmd)
			}
			
			// Test ParseError method
			msg, isParseErr := dsl.ParseError(err)
			if isParseErr && msg == "" {
				t.Error("ParseError returned empty message")
			}
		})
	}
}

// Helper functions

func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	
	// Create a pattern
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			r := uint8((x + y) % 256)
			g := uint8((x * 2) % 256)
			b := uint8((y * 2) % 256)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	
	// Add some rectangles to simulate license plate
	for y := 80; y < 120; y++ {
		for x := 50; x < 150; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	
	return img
}

func saveTestImage(t *testing.T, img image.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer file.Close()
	
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
}

// BenchmarkFFT benchmarks FFT performance
func BenchmarkFFT(b *testing.B) {
	engine := NewImageEngine()
	engine.SetCurrentImage(createTestImage())
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.FFT()
	}
}

// BenchmarkFilters benchmarks filter performance
func BenchmarkFilters(b *testing.B) {
	engine := NewImageEngine()
	testImg := createTestImage()
	
	filters := []string{"gaussian", "median", "sobel", "canny"}
	
	for _, filter := range filters {
		b.Run(filter, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.SetCurrentImage(testImg)
				engine.ApplyFilter(filter, 1.0)
			}
		})
	}
}