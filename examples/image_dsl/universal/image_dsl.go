package universal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// ImageDSL represents the image processing DSL
type ImageDSL struct {
	dsl    *dslbuilder.DSL
	engine *ImageEngine
}

// NewImageDSL creates a new image processing DSL instance
func NewImageDSL() *ImageDSL {
	id := &ImageDSL{
		dsl:    dslbuilder.New("ImageDSL"),
		engine: NewImageEngine(),
	}
	id.setupGrammar()
	return id
}

func (id *ImageDSL) setupGrammar() {
	// Keywords for operations
	id.dsl.KeywordToken("load", "load")
	id.dsl.KeywordToken("save", "save")
	id.dsl.KeywordToken("fft", "fft")
	id.dsl.KeywordToken("ifft", "ifft")
	id.dsl.KeywordToken("dct", "dct")
	id.dsl.KeywordToken("idct", "idct")
	id.dsl.KeywordToken("laplace", "laplace")
	id.dsl.KeywordToken("inverse_laplace", "inverse_laplace")
	id.dsl.KeywordToken("compress", "compress")
	id.dsl.KeywordToken("decompress", "decompress")
	id.dsl.KeywordToken("resize", "resize")
	id.dsl.KeywordToken("rotate", "rotate")
	id.dsl.KeywordToken("flip", "flip")
	id.dsl.KeywordToken("crop", "crop")
	id.dsl.KeywordToken("filter", "filter")
	id.dsl.KeywordToken("watermark", "watermark")
	id.dsl.KeywordToken("remove_watermark", "remove_watermark")
	id.dsl.KeywordToken("detect_plate", "detect_plate")
	id.dsl.KeywordToken("ocr_plate", "ocr_plate")
	id.dsl.KeywordToken("ocr_captcha", "ocr_captcha")
	id.dsl.KeywordToken("extract_plate", "extract_plate")
	id.dsl.KeywordToken("brightness", "brightness")
	id.dsl.KeywordToken("contrast", "contrast")
	id.dsl.KeywordToken("saturation", "saturation")
	id.dsl.KeywordToken("histogram", "histogram")
	id.dsl.KeywordToken("equalize", "equalize")
	id.dsl.KeywordToken("denoise", "denoise")
	id.dsl.KeywordToken("edge_detect", "edge_detect")
	
	// Keywords for parameters
	id.dsl.KeywordToken("to", "to")
	id.dsl.KeywordToken("at", "at")
	id.dsl.KeywordToken("with", "with")
	id.dsl.KeywordToken("from", "from")
	id.dsl.KeywordToken("type", "type")
	id.dsl.KeywordToken("quality", "quality")
	id.dsl.KeywordToken("position", "position")
	id.dsl.KeywordToken("opacity", "opacity")
	id.dsl.KeywordToken("text", "text")
	id.dsl.KeywordToken("image", "image")
	id.dsl.KeywordToken("degrees", "degrees")
	id.dsl.KeywordToken("width", "width")
	id.dsl.KeywordToken("height", "height")
	id.dsl.KeywordToken("x", "x")
	id.dsl.KeywordToken("y", "y")
	id.dsl.KeywordToken("horizontal", "horizontal")
	id.dsl.KeywordToken("vertical", "vertical")
	id.dsl.KeywordToken("method", "method")
	id.dsl.KeywordToken("strength", "strength")
	
	// Filter types
	id.dsl.KeywordToken("gaussian", "gaussian")
	id.dsl.KeywordToken("median", "median")
	id.dsl.KeywordToken("bilateral", "bilateral")
	id.dsl.KeywordToken("sobel", "sobel")
	id.dsl.KeywordToken("prewitt", "prewitt")
	id.dsl.KeywordToken("canny", "canny")
	id.dsl.KeywordToken("laplacian", "laplacian")
	id.dsl.KeywordToken("sharpen", "sharpen")
	id.dsl.KeywordToken("emboss", "emboss")
	id.dsl.KeywordToken("blur", "blur")
	id.dsl.KeywordToken("motion_blur", "motion_blur")
	
	// Watermark positions
	id.dsl.KeywordToken("top_left", "top_left")
	id.dsl.KeywordToken("top_right", "top_right")
	id.dsl.KeywordToken("bottom_left", "bottom_left")
	id.dsl.KeywordToken("bottom_right", "bottom_right")
	id.dsl.KeywordToken("center", "center")
	id.dsl.KeywordToken("tiled", "tiled")
	id.dsl.KeywordToken("diagonal", "diagonal")
	
	// Compression types
	id.dsl.KeywordToken("jpeg", "jpeg")
	id.dsl.KeywordToken("png", "png")
	id.dsl.KeywordToken("webp", "webp")
	id.dsl.KeywordToken("lossless", "lossless")
	id.dsl.KeywordToken("lossy", "lossy")
	
	// Tokens for values
	id.dsl.Token("STRING", `"[^"]*"`)
	id.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	id.dsl.Token("PERCENT", `[0-9]+%`)
	id.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)
	
	// Grammar rules
	id.dsl.Rule("program", []string{"statement"}, "executeStatement")
	id.dsl.Action("executeStatement", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// Statement alternatives
	id.dsl.Rule("statement", []string{"load_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"save_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"transform_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"edit_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"filter_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"watermark_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"ocr_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"detect_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"adjust_cmd"}, "passthrough")
	id.dsl.Rule("statement", []string{"analyze_cmd"}, "passthrough")
	
	id.dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})
	
	// Load/Save operations
	id.dsl.Rule("load_cmd", []string{"load", "STRING"}, "loadFile")
	id.dsl.Action("loadFile", func(args []interface{}) (interface{}, error) {
		filename := strings.Trim(args[1].(string), "\"")
		return id.engine.LoadImage(filename)
	})
	
	id.dsl.Rule("save_cmd", []string{"save", "STRING"}, "saveFile")
	id.dsl.Rule("save_cmd", []string{"save", "STRING", "quality", "NUMBER"}, "saveFileQuality")
	
	id.dsl.Action("saveFile", func(args []interface{}) (interface{}, error) {
		filename := strings.Trim(args[1].(string), "\"")
		return id.engine.SaveImage(filename, 95)
	})
	
	id.dsl.Action("saveFileQuality", func(args []interface{}) (interface{}, error) {
		filename := strings.Trim(args[1].(string), "\"")
		quality, _ := strconv.Atoi(args[3].(string))
		return id.engine.SaveImage(filename, quality)
	})
	
	// Transform operations
	id.dsl.Rule("transform_cmd", []string{"fft_cmd"}, "passthrough")
	id.dsl.Rule("transform_cmd", []string{"ifft_cmd"}, "passthrough")
	id.dsl.Rule("transform_cmd", []string{"dct_cmd"}, "passthrough")
	id.dsl.Rule("transform_cmd", []string{"idct_cmd"}, "passthrough")
	id.dsl.Rule("transform_cmd", []string{"laplace_cmd"}, "passthrough")
	id.dsl.Rule("transform_cmd", []string{"inverse_laplace_cmd"}, "passthrough")
	
	// FFT commands
	id.dsl.Rule("fft_cmd", []string{"fft"}, "fftCmd")
	id.dsl.Action("fftCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.FFT()
	})
	
	id.dsl.Rule("ifft_cmd", []string{"ifft"}, "ifftCmd")
	id.dsl.Action("ifftCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.IFFT()
	})
	
	// DCT commands
	id.dsl.Rule("dct_cmd", []string{"dct"}, "dctCmd")
	id.dsl.Action("dctCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.DCT()
	})
	
	id.dsl.Rule("idct_cmd", []string{"idct"}, "idctCmd")
	id.dsl.Action("idctCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.IDCT()
	})
	
	// Laplace commands
	id.dsl.Rule("laplace_cmd", []string{"laplace"}, "laplaceCmd")
	id.dsl.Action("laplaceCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.LaplaceTransform()
	})
	
	id.dsl.Rule("inverse_laplace_cmd", []string{"inverse_laplace"}, "inverseLaplaceCmd")
	id.dsl.Action("inverseLaplaceCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.InverseLaplaceTransform()
	})
	
	// Edit operations
	id.dsl.Rule("edit_cmd", []string{"resize_cmd"}, "passthrough")
	id.dsl.Rule("edit_cmd", []string{"rotate_cmd"}, "passthrough")
	id.dsl.Rule("edit_cmd", []string{"flip_cmd"}, "passthrough")
	id.dsl.Rule("edit_cmd", []string{"crop_cmd"}, "passthrough")
	id.dsl.Rule("edit_cmd", []string{"compress_cmd"}, "passthrough")
	
	// Resize command
	id.dsl.Rule("resize_cmd", []string{"resize", "width", "NUMBER", "height", "NUMBER"}, "resizeCmd")
	id.dsl.Rule("resize_cmd", []string{"resize", "PERCENT"}, "resizePercentCmd")
	
	id.dsl.Action("resizeCmd", func(args []interface{}) (interface{}, error) {
		width, _ := strconv.Atoi(args[2].(string))
		height, _ := strconv.Atoi(args[4].(string))
		return id.engine.Resize(width, height)
	})
	
	id.dsl.Action("resizePercentCmd", func(args []interface{}) (interface{}, error) {
		percentStr := args[1].(string)
		percent := strings.TrimSuffix(percentStr, "%")
		scale, _ := strconv.ParseFloat(percent, 64)
		return id.engine.ResizePercent(scale / 100.0)
	})
	
	// Rotate command
	id.dsl.Rule("rotate_cmd", []string{"rotate", "NUMBER", "degrees"}, "rotateCmd")
	id.dsl.Action("rotateCmd", func(args []interface{}) (interface{}, error) {
		degrees, _ := strconv.ParseFloat(args[1].(string), 64)
		return id.engine.Rotate(degrees)
	})
	
	// Flip command
	id.dsl.Rule("flip_cmd", []string{"flip", "horizontal"}, "flipHorizontal")
	id.dsl.Rule("flip_cmd", []string{"flip", "vertical"}, "flipVertical")
	
	id.dsl.Action("flipHorizontal", func(args []interface{}) (interface{}, error) {
		return id.engine.FlipHorizontal()
	})
	
	id.dsl.Action("flipVertical", func(args []interface{}) (interface{}, error) {
		return id.engine.FlipVertical()
	})
	
	// Crop command
	id.dsl.Rule("crop_cmd", []string{"crop", "x", "NUMBER", "y", "NUMBER", "width", "NUMBER", "height", "NUMBER"}, "cropCmd")
	id.dsl.Action("cropCmd", func(args []interface{}) (interface{}, error) {
		x, _ := strconv.Atoi(args[2].(string))
		y, _ := strconv.Atoi(args[4].(string))
		width, _ := strconv.Atoi(args[6].(string))
		height, _ := strconv.Atoi(args[8].(string))
		return id.engine.Crop(x, y, width, height)
	})
	
	// Compress command
	id.dsl.Rule("compress_cmd", []string{"compress", "type", "compress_type", "quality", "NUMBER"}, "compressCmd")
	id.dsl.Rule("compress_type", []string{"jpeg"}, "compressType")
	id.dsl.Rule("compress_type", []string{"png"}, "compressType")
	id.dsl.Rule("compress_type", []string{"webp"}, "compressType")
	
	id.dsl.Action("compressType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	id.dsl.Action("compressCmd", func(args []interface{}) (interface{}, error) {
		compType := args[2].(string)
		quality, _ := strconv.Atoi(args[4].(string))
		return id.engine.Compress(compType, quality)
	})
	
	// Filter operations
	id.dsl.Rule("filter_cmd", []string{"filter", "filter_type"}, "filterSimple")
	id.dsl.Rule("filter_cmd", []string{"filter", "filter_type", "strength", "NUMBER"}, "filterStrength")
	
	id.dsl.Rule("filter_type", []string{"gaussian"}, "filterType")
	id.dsl.Rule("filter_type", []string{"median"}, "filterType")
	id.dsl.Rule("filter_type", []string{"bilateral"}, "filterType")
	id.dsl.Rule("filter_type", []string{"sobel"}, "filterType")
	id.dsl.Rule("filter_type", []string{"prewitt"}, "filterType")
	id.dsl.Rule("filter_type", []string{"canny"}, "filterType")
	id.dsl.Rule("filter_type", []string{"laplacian"}, "filterType")
	id.dsl.Rule("filter_type", []string{"sharpen"}, "filterType")
	id.dsl.Rule("filter_type", []string{"emboss"}, "filterType")
	id.dsl.Rule("filter_type", []string{"blur"}, "filterType")
	id.dsl.Rule("filter_type", []string{"motion_blur"}, "filterType")
	
	id.dsl.Action("filterType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	id.dsl.Action("filterSimple", func(args []interface{}) (interface{}, error) {
		filterType := args[1].(string)
		return id.engine.ApplyFilter(filterType, 1.0)
	})
	
	id.dsl.Action("filterStrength", func(args []interface{}) (interface{}, error) {
		filterType := args[1].(string)
		strength, _ := strconv.ParseFloat(args[3].(string), 64)
		return id.engine.ApplyFilter(filterType, strength)
	})
	
	// Watermark operations
	id.dsl.Rule("watermark_cmd", []string{"watermark_add"}, "passthrough")
	id.dsl.Rule("watermark_cmd", []string{"watermark_remove"}, "passthrough")
	
	// Add watermark
	id.dsl.Rule("watermark_add", []string{"watermark", "text", "STRING", "position", "watermark_position"}, "watermarkText")
	id.dsl.Rule("watermark_add", []string{"watermark", "text", "STRING", "position", "watermark_position", "opacity", "NUMBER"}, "watermarkTextOpacity")
	id.dsl.Rule("watermark_add", []string{"watermark", "image", "STRING", "position", "watermark_position"}, "watermarkImage")
	id.dsl.Rule("watermark_add", []string{"watermark", "image", "STRING", "position", "watermark_position", "opacity", "NUMBER"}, "watermarkImageOpacity")
	
	id.dsl.Rule("watermark_position", []string{"top_left"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"top_right"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"bottom_left"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"bottom_right"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"center"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"tiled"}, "watermarkPos")
	id.dsl.Rule("watermark_position", []string{"diagonal"}, "watermarkPos")
	
	id.dsl.Action("watermarkPos", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	id.dsl.Action("watermarkText", func(args []interface{}) (interface{}, error) {
		text := strings.Trim(args[2].(string), "\"")
		position := args[4].(string)
		return id.engine.AddTextWatermark(text, position, 0.5)
	})
	
	id.dsl.Action("watermarkTextOpacity", func(args []interface{}) (interface{}, error) {
		text := strings.Trim(args[2].(string), "\"")
		position := args[4].(string)
		opacity, _ := strconv.ParseFloat(args[6].(string), 64)
		return id.engine.AddTextWatermark(text, position, opacity)
	})
	
	id.dsl.Action("watermarkImage", func(args []interface{}) (interface{}, error) {
		imagePath := strings.Trim(args[2].(string), "\"")
		position := args[4].(string)
		return id.engine.AddImageWatermark(imagePath, position, 0.5)
	})
	
	id.dsl.Action("watermarkImageOpacity", func(args []interface{}) (interface{}, error) {
		imagePath := strings.Trim(args[2].(string), "\"")
		position := args[4].(string)
		opacity, _ := strconv.ParseFloat(args[6].(string), 64)
		return id.engine.AddImageWatermark(imagePath, position, opacity)
	})
	
	// Remove watermark
	id.dsl.Rule("watermark_remove", []string{"remove_watermark"}, "removeWatermark")
	id.dsl.Rule("watermark_remove", []string{"remove_watermark", "method", "STRING"}, "removeWatermarkMethod")
	
	id.dsl.Action("removeWatermark", func(args []interface{}) (interface{}, error) {
		return id.engine.RemoveWatermark("inpaint")
	})
	
	id.dsl.Action("removeWatermarkMethod", func(args []interface{}) (interface{}, error) {
		method := strings.Trim(args[2].(string), "\"")
		return id.engine.RemoveWatermark(method)
	})
	
	// OCR operations
	id.dsl.Rule("ocr_cmd", []string{"ocr_plate_cmd"}, "passthrough")
	id.dsl.Rule("ocr_cmd", []string{"ocr_captcha_cmd"}, "passthrough")
	
	id.dsl.Rule("ocr_plate_cmd", []string{"ocr_plate"}, "ocrPlate")
	id.dsl.Action("ocrPlate", func(args []interface{}) (interface{}, error) {
		return id.engine.OCRPlate()
	})
	
	id.dsl.Rule("ocr_captcha_cmd", []string{"ocr_captcha"}, "ocrCaptcha")
	id.dsl.Action("ocrCaptcha", func(args []interface{}) (interface{}, error) {
		return id.engine.OCRCaptcha()
	})
	
	// Detection operations
	id.dsl.Rule("detect_cmd", []string{"detect_plate_cmd"}, "passthrough")
	id.dsl.Rule("detect_cmd", []string{"extract_plate_cmd"}, "passthrough")
	
	id.dsl.Rule("detect_plate_cmd", []string{"detect_plate"}, "detectPlate")
	id.dsl.Action("detectPlate", func(args []interface{}) (interface{}, error) {
		return id.engine.DetectPlate()
	})
	
	id.dsl.Rule("extract_plate_cmd", []string{"extract_plate"}, "extractPlate")
	id.dsl.Action("extractPlate", func(args []interface{}) (interface{}, error) {
		return id.engine.ExtractPlate()
	})
	
	// Adjustment operations
	id.dsl.Rule("adjust_cmd", []string{"brightness_cmd"}, "passthrough")
	id.dsl.Rule("adjust_cmd", []string{"contrast_cmd"}, "passthrough")
	id.dsl.Rule("adjust_cmd", []string{"saturation_cmd"}, "passthrough")
	id.dsl.Rule("adjust_cmd", []string{"denoise_cmd"}, "passthrough")
	
	id.dsl.Rule("brightness_cmd", []string{"brightness", "NUMBER"}, "brightnessCmd")
	id.dsl.Action("brightnessCmd", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		return id.engine.AdjustBrightness(value)
	})
	
	id.dsl.Rule("contrast_cmd", []string{"contrast", "NUMBER"}, "contrastCmd")
	id.dsl.Action("contrastCmd", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		return id.engine.AdjustContrast(value)
	})
	
	id.dsl.Rule("saturation_cmd", []string{"saturation", "NUMBER"}, "saturationCmd")
	id.dsl.Action("saturationCmd", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[1].(string), 64)
		return id.engine.AdjustSaturation(value)
	})
	
	id.dsl.Rule("denoise_cmd", []string{"denoise"}, "denoiseCmd")
	id.dsl.Action("denoiseCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.Denoise()
	})
	
	// Analysis operations
	id.dsl.Rule("analyze_cmd", []string{"histogram_cmd"}, "passthrough")
	id.dsl.Rule("analyze_cmd", []string{"equalize_cmd"}, "passthrough")
	id.dsl.Rule("analyze_cmd", []string{"edge_detect_cmd"}, "passthrough")
	
	id.dsl.Rule("histogram_cmd", []string{"histogram"}, "histogramCmd")
	id.dsl.Action("histogramCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.GetHistogram()
	})
	
	id.dsl.Rule("equalize_cmd", []string{"equalize"}, "equalizeCmd")
	id.dsl.Action("equalizeCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.EqualizeHistogram()
	})
	
	id.dsl.Rule("edge_detect_cmd", []string{"edge_detect"}, "edgeDetectCmd")
	id.dsl.Action("edgeDetectCmd", func(args []interface{}) (interface{}, error) {
		return id.engine.EdgeDetect()
	})
}

// Parse processes DSL input and returns the result
func (id *ImageDSL) Parse(input string) (interface{}, error) {
	result, err := id.dsl.Parse(input)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// GetEngine returns the image processing engine
func (id *ImageDSL) GetEngine() *ImageEngine {
	return id.engine
}

// ParseError checks if an error is a parse error
func (id *ImageDSL) ParseError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if parseErr, ok := err.(*dslbuilder.ParseError); ok {
		return fmt.Sprintf("Parse error at line %d, column %d: %s", 
			parseErr.Line, parseErr.Column, parseErr.Message), true
	}
	return err.Error(), false
}