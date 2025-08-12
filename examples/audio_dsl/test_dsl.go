package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/arturoeanton/go-dsl/examples/audio_dsl/universal"
)

func main() {
	// Create DSL instance
	audioDSL := universal.NewAudioDSL()
	
	// Generate test WAV file
	generateTestWAV("test_input.wav", audioDSL.GetEngine())
	
	// Read test commands
	file, err := os.Open("test_commands.txt")
	if err != nil {
		fmt.Printf("Error opening test file: %v\n", err)
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}
		
		fmt.Printf("\n> %s\n", cmd)
		result, err := audioDSL.Parse(cmd)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if result != nil {
			fmt.Printf("  SUCCESS: %v\n", result)
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func generateTestWAV(filename string, engine *universal.AudioEngine) {
	// Generate a simple test signal
	sampleRate := 44100
	duration := 3.0
	numSamples := int(float64(sampleRate) * duration)
	
	samples := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		// Mix of frequencies
		samples[i] = 0.3 * math.Sin(2*math.Pi*440*t)    // A4
		samples[i] += 0.2 * math.Sin(2*math.Pi*880*t)   // A5
		samples[i] += 0.1 * math.Sin(2*math.Pi*220*t)   // A3
	}
	
	engine.Samples = samples
	engine.SampleRate = sampleRate
	engine.Channels = 1
	
	engine.SaveWAV(filename)
	fmt.Printf("Generated test file: %s\n", filename)
}