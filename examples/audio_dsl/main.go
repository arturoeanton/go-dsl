package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/arturoeanton/go-dsl/examples/audio_dsl/universal"
)

func main() {
	fmt.Println("Audio DSL - Digital Signal Processing Example")
	fmt.Println("=" + strings.Repeat("=", 45))

	// Create DSL instance
	audioDSL := universal.NewAudioDSL()

	// Generate a test WAV file for demonstration
	generateTestWAV("test_input.wav")

	// Example commands to demonstrate functionality
	commands := []string{
		// Load audio file
		`load "test_input.wav"`,

		// Frequency analysis
		`fft with window hamming`,
		`spectrum`,
		`phase`,

		// Back to time domain
		`ifft`,

		// Audio editing operations
		`mark at 0.5s`,
		`copy from 0.2s to 0.8s`,
		`paste at 1.5s`,

		// Audio processing
		`normalize`,
		`filter lowpass 2000 Hz`,
		`amplify 6 dB`,

		// Trimming
		`trim from 0.1s to 2.0s`,

		// Compression
		`compress type flac ratio 4`,

		// Save processed audio
		`save "test_output.wav"`,

		// Advanced transforms
		`laplace`,
		`inverse_laplace`,

		// Resampling
		`resample rate 22050 Hz`,
		`save "test_resampled.wav"`,
	}

	fmt.Println("\nExecuting Audio DSL Commands:")
	fmt.Println("-" + strings.Repeat("-", 45))

	for _, cmd := range commands {
		fmt.Printf("\n> %s\n", cmd)
		result, err := audioDSL.Parse(cmd)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if result != nil {
			fmt.Printf("  %v\n", result)
		}
	}

	// Demonstrate programmatic usage
	fmt.Println("\n\nProgrammatic Usage Example:")
	fmt.Println("-" + strings.Repeat("-", 45))

	engine := audioDSL.GetEngine()
	
	// Generate and process a synthetic signal
	generateSyntheticSignal(engine, 440.0, 1.0) // 440 Hz for 1 second
	
	fmt.Println("\n1. Generated 440 Hz sine wave")
	
	// Apply FFT
	result, _ := engine.FFT("blackman")
	fmt.Printf("2. %s\n", result)
	
	// Get spectrum info
	result, _ = engine.GetSpectrum()
	fmt.Printf("3. %s\n", result)
	
	// Apply filter
	result, _ = engine.Filter("bandpass", 400)
	fmt.Printf("4. %s\n", result)
	
	// Normalize
	result, _ = engine.Normalize()
	fmt.Printf("5. %s\n", result)
	
	// Save
	result, _ = engine.SaveWAV("synthetic_output.wav")
	fmt.Printf("6. %s\n", result)

	fmt.Println("\n\nAdvanced Signal Processing:")
	fmt.Println("-" + strings.Repeat("-", 45))

	// Demonstrate Laplace transform for system analysis
	fmt.Println("\nSystem Analysis with Laplace Transform:")
	result, _ = engine.LaplaceTransform()
	fmt.Printf("  %s\n", result)

	// Demonstrate audio compression workflow
	fmt.Println("\nAudio Compression Workflow:")
	engine.LoadWAV("test_input.wav")
	engine.Compress("mp3", 8.0)
	fmt.Println("  Applied MP3 compression at 8:1 ratio")
	engine.SaveWAV("compressed_output.wav")

	fmt.Println("\n\nDemo completed successfully!")
	fmt.Println("\nGenerated files:")
	fmt.Println("  - test_input.wav (original)")
	fmt.Println("  - test_output.wav (processed)")
	fmt.Println("  - test_resampled.wav (22.05 kHz)")
	fmt.Println("  - synthetic_output.wav (synthesized)")
	fmt.Println("  - compressed_output.wav (compressed)")

	// Clean up test files (optional)
	// cleanupTestFiles()
}

// generateTestWAV creates a test WAV file with multiple frequency components
func generateTestWAV(filename string) {
	engine := universal.NewAudioEngine()
	
	// Generate a complex test signal
	sampleRate := 44100
	duration := 3.0 // 3 seconds
	numSamples := int(float64(sampleRate) * duration)
	
	samples := make([]float64, numSamples)
	
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		
		// Mix of frequencies
		samples[i] = 0.3 * math.Sin(2*math.Pi*440*t)    // A4 note
		samples[i] += 0.2 * math.Sin(2*math.Pi*880*t)   // A5 note
		samples[i] += 0.1 * math.Sin(2*math.Pi*220*t)   // A3 note
		samples[i] += 0.15 * math.Sin(2*math.Pi*1320*t) // E6 note
		
		// Add envelope
		envelope := 1.0
		if t < 0.1 {
			envelope = t / 0.1 // Fade in
		} else if t > duration-0.1 {
			envelope = (duration - t) / 0.1 // Fade out
		}
		samples[i] *= envelope
		
		// Add some noise
		samples[i] += 0.02 * (math.Sin(t*1000) * math.Cos(t*3000))
	}
	
	// Set engine properties
	engine.Samples = samples
	engine.SampleRate = sampleRate
	engine.Channels = 1
	
	// Save to file
	engine.SaveWAV(filename)
	fmt.Printf("Generated test file: %s\n", filename)
}

// generateSyntheticSignal creates a pure sine wave
func generateSyntheticSignal(engine *universal.AudioEngine, frequency, duration float64) {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)
	
	samples := make([]float64, numSamples)
	
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		samples[i] = math.Sin(2 * math.Pi * frequency * t)
	}
	
	engine.Samples = samples
	engine.SampleRate = sampleRate
	engine.Channels = 1
}

// cleanupTestFiles removes generated test files
func cleanupTestFiles() {
	files := []string{
		"test_input.wav",
		"test_output.wav",
		"test_resampled.wav",
		"synthetic_output.wav",
		"compressed_output.wav",
	}
	
	for _, file := range files {
		os.Remove(file)
	}
}