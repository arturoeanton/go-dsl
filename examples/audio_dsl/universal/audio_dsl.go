package universal

import (
	"math"
	"strconv"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

type AudioDSL struct {
	dsl    *dslbuilder.DSL
	engine *AudioEngine
}

func NewAudioDSL() *AudioDSL {
	ad := &AudioDSL{
		dsl:    dslbuilder.New("AudioDSL"),
		engine: NewAudioEngine(),
	}
	ad.setupGrammar()
	return ad
}

func (ad *AudioDSL) setupGrammar() {
	// Keywords for operations
	ad.dsl.KeywordToken("load", "load")
	ad.dsl.KeywordToken("save", "save")
	ad.dsl.KeywordToken("fft", "fft")
	ad.dsl.KeywordToken("ifft", "ifft")
	ad.dsl.KeywordToken("laplace", "laplace")
	ad.dsl.KeywordToken("inverse_laplace", "inverse_laplace")
	ad.dsl.KeywordToken("compress", "compress")
	ad.dsl.KeywordToken("decompress", "decompress")
	ad.dsl.KeywordToken("mark", "mark")
	ad.dsl.KeywordToken("cut", "cut")
	ad.dsl.KeywordToken("copy", "copy")
	ad.dsl.KeywordToken("paste", "paste")
	ad.dsl.KeywordToken("trim", "trim")
	ad.dsl.KeywordToken("amplify", "amplify")
	ad.dsl.KeywordToken("normalize", "normalize")
	ad.dsl.KeywordToken("filter", "filter")
	ad.dsl.KeywordToken("resample", "resample")
	ad.dsl.KeywordToken("spectrum", "spectrum")
	ad.dsl.KeywordToken("phase", "phase")
	ad.dsl.KeywordToken("window", "window")
	ad.dsl.KeywordToken("from", "from")
	ad.dsl.KeywordToken("to", "to")
	ad.dsl.KeywordToken("at", "at")
	ad.dsl.KeywordToken("with", "with")
	ad.dsl.KeywordToken("type", "type")
	ad.dsl.KeywordToken("rate", "rate")
	ad.dsl.KeywordToken("ratio", "ratio")
	ad.dsl.KeywordToken("gain", "gain")
	ad.dsl.KeywordToken("dB", "dB")
	ad.dsl.KeywordToken("Hz", "Hz")
	ad.dsl.KeywordToken("kHz", "kHz")
	ad.dsl.KeywordToken("ms", "ms")
	ad.dsl.KeywordToken("s", "s")

	// Window types
	ad.dsl.KeywordToken("hamming", "hamming")
	ad.dsl.KeywordToken("hanning", "hanning")
	ad.dsl.KeywordToken("blackman", "blackman")
	ad.dsl.KeywordToken("rectangular", "rectangular")

	// Filter types
	ad.dsl.KeywordToken("lowpass", "lowpass")
	ad.dsl.KeywordToken("highpass", "highpass")
	ad.dsl.KeywordToken("bandpass", "bandpass")
	ad.dsl.KeywordToken("notch", "notch")

	// Compression types
	ad.dsl.KeywordToken("mp3", "mp3")
	ad.dsl.KeywordToken("flac", "flac")
	ad.dsl.KeywordToken("opus", "opus")

	// Tokens for values
	ad.dsl.Token("STRING", `"[^"]*"`)
	ad.dsl.Token("NUMBER", `[0-9]+(\.[0-9]+)?`)
	ad.dsl.Token("ID", `[a-zA-Z_][a-zA-Z0-9_]*`)

	// Grammar rules
	ad.dsl.Rule("program", []string{"statement"}, "executeStatement")
	ad.dsl.Action("executeStatement", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})

	// Statement alternatives
	ad.dsl.Rule("statement", []string{"load_cmd"}, "passthrough")
	ad.dsl.Rule("statement", []string{"save_cmd"}, "passthrough")
	ad.dsl.Rule("statement", []string{"transform_cmd"}, "passthrough")
	ad.dsl.Rule("statement", []string{"edit_cmd"}, "passthrough")
	ad.dsl.Rule("statement", []string{"process_cmd"}, "passthrough")
	ad.dsl.Rule("statement", []string{"analysis_cmd"}, "passthrough")
	ad.dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return nil, nil
	})

	// Load/Save operations
	ad.dsl.Rule("load_cmd", []string{"load", "STRING"}, "loadFile")
	ad.dsl.Action("loadFile", func(args []interface{}) (interface{}, error) {
		filename := strings.Trim(args[1].(string), "\"")
		return ad.engine.LoadWAV(filename)
	})

	ad.dsl.Rule("save_cmd", []string{"save", "STRING"}, "saveFile")
	ad.dsl.Action("saveFile", func(args []interface{}) (interface{}, error) {
		filename := strings.Trim(args[1].(string), "\"")
		return ad.engine.SaveWAV(filename)
	})

	// Transform operations
	ad.dsl.Rule("transform_cmd", []string{"fft_cmd"}, "passthrough")
	ad.dsl.Rule("transform_cmd", []string{"ifft_cmd"}, "passthrough")
	ad.dsl.Rule("transform_cmd", []string{"laplace_cmd"}, "passthrough")
	ad.dsl.Rule("transform_cmd", []string{"inverse_laplace_cmd"}, "passthrough")

	// FFT commands - longer pattern first for priority
	ad.dsl.Rule("fft_cmd", []string{"fft", "with", "window", "window_type"}, "fftWindow")
	ad.dsl.Rule("fft_cmd", []string{"fft"}, "fftDefault")
	
	ad.dsl.Action("fftDefault", func(args []interface{}) (interface{}, error) {
		return ad.engine.FFT("hamming")
	})
	
	ad.dsl.Action("fftWindow", func(args []interface{}) (interface{}, error) {
		window := args[3].(string)
		return ad.engine.FFT(window)
	})

	ad.dsl.Rule("window_type", []string{"hamming"}, "windowType")
	ad.dsl.Rule("window_type", []string{"hanning"}, "windowType")
	ad.dsl.Rule("window_type", []string{"blackman"}, "windowType")
	ad.dsl.Rule("window_type", []string{"rectangular"}, "windowType")
	ad.dsl.Action("windowType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	ad.dsl.Rule("ifft_cmd", []string{"ifft"}, "ifftCmd")
	ad.dsl.Action("ifftCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.IFFT()
	})

	ad.dsl.Rule("laplace_cmd", []string{"laplace"}, "laplaceCmd")
	ad.dsl.Action("laplaceCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.LaplaceTransform()
	})

	ad.dsl.Rule("inverse_laplace_cmd", []string{"inverse_laplace"}, "inverseLaplaceCmd")
	ad.dsl.Action("inverseLaplaceCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.InverseLaplaceTransform()
	})

	// Edit operations
	ad.dsl.Rule("edit_cmd", []string{"mark_cmd"}, "passthrough")
	ad.dsl.Rule("edit_cmd", []string{"cut_cmd"}, "passthrough")
	ad.dsl.Rule("edit_cmd", []string{"copy_cmd"}, "passthrough")
	ad.dsl.Rule("edit_cmd", []string{"paste_cmd"}, "passthrough")
	ad.dsl.Rule("edit_cmd", []string{"trim_cmd"}, "passthrough")

	ad.dsl.Rule("mark_cmd", []string{"mark", "at", "time_value"}, "markCmd")
	ad.dsl.Action("markCmd", func(args []interface{}) (interface{}, error) {
		time := args[2].(float64)
		return ad.engine.Mark(time)
	})

	ad.dsl.Rule("cut_cmd", []string{"cut", "from", "time_value", "to", "time_value"}, "cutCmd")
	ad.dsl.Action("cutCmd", func(args []interface{}) (interface{}, error) {
		start := args[2].(float64)
		end := args[4].(float64)
		return ad.engine.Cut(start, end)
	})

	ad.dsl.Rule("copy_cmd", []string{"copy", "from", "time_value", "to", "time_value"}, "copyCmd")
	ad.dsl.Action("copyCmd", func(args []interface{}) (interface{}, error) {
		start := args[2].(float64)
		end := args[4].(float64)
		return ad.engine.Copy(start, end)
	})

	ad.dsl.Rule("paste_cmd", []string{"paste", "at", "time_value"}, "pasteCmd")
	ad.dsl.Action("pasteCmd", func(args []interface{}) (interface{}, error) {
		time := args[2].(float64)
		return ad.engine.Paste(time)
	})

	ad.dsl.Rule("trim_cmd", []string{"trim", "from", "time_value", "to", "time_value"}, "trimCmd")
	ad.dsl.Action("trimCmd", func(args []interface{}) (interface{}, error) {
		start := args[2].(float64)
		end := args[4].(float64)
		return ad.engine.Trim(start, end)
	})

	// Process operations
	ad.dsl.Rule("process_cmd", []string{"amplify_cmd"}, "passthrough")
	ad.dsl.Rule("process_cmd", []string{"normalize_cmd"}, "passthrough")
	ad.dsl.Rule("process_cmd", []string{"filter_cmd"}, "passthrough")
	ad.dsl.Rule("process_cmd", []string{"resample_cmd"}, "passthrough")
	ad.dsl.Rule("process_cmd", []string{"compress_cmd"}, "passthrough")
	ad.dsl.Rule("process_cmd", []string{"decompress_cmd"}, "passthrough")

	ad.dsl.Rule("amplify_cmd", []string{"amplify", "gain_value"}, "amplifyCmd")
	ad.dsl.Action("amplifyCmd", func(args []interface{}) (interface{}, error) {
		gain := args[1].(float64)
		return ad.engine.Amplify(gain)
	})

	ad.dsl.Rule("gain_value", []string{"NUMBER", "dB"}, "gainDB")
	ad.dsl.Rule("gain_value", []string{"NUMBER"}, "gainLinear")
	
	ad.dsl.Action("gainDB", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return dBToLinear(value), nil
	})
	
	ad.dsl.Action("gainLinear", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value, nil
	})

	ad.dsl.Rule("normalize_cmd", []string{"normalize"}, "normalizeCmd")
	ad.dsl.Action("normalizeCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.Normalize()
	})

	ad.dsl.Rule("filter_cmd", []string{"filter", "filter_type", "freq_value"}, "filterCmd")
	ad.dsl.Action("filterCmd", func(args []interface{}) (interface{}, error) {
		filterType := args[1].(string)
		freq := args[2].(float64)
		return ad.engine.Filter(filterType, freq)
	})

	ad.dsl.Rule("filter_type", []string{"lowpass"}, "filterType")
	ad.dsl.Rule("filter_type", []string{"highpass"}, "filterType")
	ad.dsl.Rule("filter_type", []string{"bandpass"}, "filterType")
	ad.dsl.Rule("filter_type", []string{"notch"}, "filterType")
	ad.dsl.Action("filterType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	ad.dsl.Rule("resample_cmd", []string{"resample", "rate", "NUMBER", "Hz"}, "resampleCmd")
	ad.dsl.Action("resampleCmd", func(args []interface{}) (interface{}, error) {
		rate, _ := strconv.ParseFloat(args[2].(string), 64)
		return ad.engine.Resample(int(rate))
	})

	ad.dsl.Rule("compress_cmd", []string{"compress", "type", "compression_type", "ratio", "NUMBER"}, "compressCmd")
	ad.dsl.Action("compressCmd", func(args []interface{}) (interface{}, error) {
		compType := args[2].(string)
		ratio, _ := strconv.ParseFloat(args[4].(string), 64)
		return ad.engine.Compress(compType, ratio)
	})

	ad.dsl.Rule("compression_type", []string{"mp3"}, "compressionType")
	ad.dsl.Rule("compression_type", []string{"flac"}, "compressionType")
	ad.dsl.Rule("compression_type", []string{"opus"}, "compressionType")
	ad.dsl.Action("compressionType", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})

	ad.dsl.Rule("decompress_cmd", []string{"decompress"}, "decompressCmd")
	ad.dsl.Action("decompressCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.Decompress()
	})

	// Analysis operations
	ad.dsl.Rule("analysis_cmd", []string{"spectrum_cmd"}, "passthrough")
	ad.dsl.Rule("analysis_cmd", []string{"phase_cmd"}, "passthrough")

	ad.dsl.Rule("spectrum_cmd", []string{"spectrum"}, "spectrumCmd")
	ad.dsl.Action("spectrumCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.GetSpectrum()
	})

	ad.dsl.Rule("phase_cmd", []string{"phase"}, "phaseCmd")
	ad.dsl.Action("phaseCmd", func(args []interface{}) (interface{}, error) {
		return ad.engine.GetPhase()
	})

	// Time and frequency values
	ad.dsl.Rule("time_value", []string{"NUMBER", "s"}, "timeSeconds")
	ad.dsl.Rule("time_value", []string{"NUMBER", "ms"}, "timeMillis")
	ad.dsl.Rule("time_value", []string{"NUMBER"}, "timeRaw")
	
	ad.dsl.Action("timeSeconds", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value, nil
	})
	
	ad.dsl.Action("timeMillis", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value / 1000.0, nil
	})
	
	ad.dsl.Action("timeRaw", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value, nil
	})

	ad.dsl.Rule("freq_value", []string{"NUMBER", "Hz"}, "freqHz")
	ad.dsl.Rule("freq_value", []string{"NUMBER", "kHz"}, "freqKHz")
	ad.dsl.Rule("freq_value", []string{"NUMBER"}, "freqRaw")
	
	ad.dsl.Action("freqHz", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value, nil
	})
	
	ad.dsl.Action("freqKHz", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value * 1000.0, nil
	})
	
	ad.dsl.Action("freqRaw", func(args []interface{}) (interface{}, error) {
		value, _ := strconv.ParseFloat(args[0].(string), 64)
		return value, nil
	})
}

func (ad *AudioDSL) Parse(input string) (interface{}, error) {
	result, err := ad.dsl.Parse(input)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

func (ad *AudioDSL) GetEngine() *AudioEngine {
	return ad.engine
}

// Helper function
func dBToLinear(dB float64) float64 {
	return math.Pow(10, dB/20.0)
}