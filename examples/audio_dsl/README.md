# Audio DSL - Digital Signal Processing Language

A domain-specific language for audio processing, signal analysis, and wave manipulation. This DSL provides comprehensive support for Fourier and Laplace transforms, WAV file handling, and audio editing operations.

## Features

### Transform Operations
- **Fast Fourier Transform (FFT)**: Convert audio from time domain to frequency domain with multiple window functions
- **Inverse FFT (IFFT)**: Reconstruct time-domain signals from frequency data
- **Laplace Transform**: Analyze system characteristics and stability
- **Inverse Laplace Transform**: Convert from s-domain back to time domain

### Audio Processing
- **WAV File Support**: Load and save WAV files (16, 24, and 32-bit)
- **Compression**: Support for MP3, FLAC, and Opus compression algorithms
- **Filtering**: Apply lowpass, highpass, bandpass, and notch filters
- **Resampling**: Change sample rates with linear interpolation

### Wave Editing
- **Mark**: Set markers at specific time points
- **Cut**: Remove sections of audio (with clipboard support)
- **Copy**: Copy sections to clipboard
- **Paste**: Insert clipboard content at any position
- **Trim**: Keep only specified time ranges
- **Amplify**: Adjust gain with dB support
- **Normalize**: Optimize audio levels

### Analysis Tools
- **Spectrum Analysis**: Extract frequency spectrum information
- **Phase Analysis**: Compute phase relationships
- **Window Functions**: Hamming, Hanning, Blackman, and Rectangular windows

## Installation

```bash
# Clone or copy the audio_dsl directory
cp -r examples/audio_dsl /your/project/

# The DSL uses only the core go-dsl package
go get github.com/arturoeanton/go-dsl
```

## Usage

### DSL Commands

The DSL supports natural language-like commands for audio processing:

```dsl
# Load a WAV file
load "input.wav"

# Apply FFT with Hamming window
fft with window hamming

# Get spectrum information
spectrum

# Return to time domain
ifft

# Edit operations
mark at 1.5s
copy from 0.5s to 1.0s
paste at 2.0s
trim from 0s to 3s

# Process audio
normalize
filter lowpass 2000 Hz
amplify 6 dB

# Resample
resample rate 22050 Hz

# Compress
compress type flac ratio 4

# Save result
save "output.wav"
```

### Programmatic API

```go
package main

import (
    "github.com/arturoeanton/go-dsl/examples/audio_dsl/universal"
)

func main() {
    // Create DSL instance
    audioDSL := universal.NewAudioDSL()
    
    // Parse DSL commands
    result, err := audioDSL.Parse(`load "audio.wav"`)
    
    // Or use the engine directly
    engine := audioDSL.GetEngine()
    engine.LoadWAV("audio.wav")
    engine.FFT("hamming")
    engine.Filter("lowpass", 1000)
    engine.Normalize()
    engine.SaveWAV("processed.wav")
}
```

## Command Reference

### File Operations

| Command | Description | Example |
|---------|-------------|---------|
| `load "file"` | Load WAV file | `load "music.wav"` |
| `save "file"` | Save WAV file | `save "output.wav"` |

### Transform Operations

| Command | Description | Example |
|---------|-------------|---------|
| `fft [with window TYPE]` | Apply FFT | `fft with window hamming` |
| `ifft` | Apply inverse FFT | `ifft` |
| `laplace` | Laplace transform | `laplace` |
| `inverse_laplace` | Inverse Laplace | `inverse_laplace` |

### Editing Operations

| Command | Description | Example |
|---------|-------------|---------|
| `mark at TIME` | Add marker | `mark at 1.5s` |
| `cut from TIME to TIME` | Cut section | `cut from 1s to 2s` |
| `copy from TIME to TIME` | Copy section | `copy from 0s to 1s` |
| `paste at TIME` | Paste clipboard | `paste at 3s` |
| `trim from TIME to TIME` | Keep range | `trim from 0.5s to 2.5s` |

### Processing Operations

| Command | Description | Example |
|---------|-------------|---------|
| `amplify GAIN` | Adjust volume | `amplify 3 dB` |
| `normalize` | Normalize levels | `normalize` |
| `filter TYPE FREQ` | Apply filter | `filter lowpass 1000 Hz` |
| `resample rate RATE` | Change sample rate | `resample rate 48000 Hz` |
| `compress type TYPE ratio RATIO` | Compress audio | `compress type mp3 ratio 8` |
| `decompress` | Decompress audio | `decompress` |

### Analysis Operations

| Command | Description | Example |
|---------|-------------|---------|
| `spectrum` | Get frequency spectrum | `spectrum` |
| `phase` | Get phase information | `phase` |

## Time and Frequency Units

- **Time**: `NUMBER`, `NUMBERs`, `NUMBERms` (e.g., `1.5`, `1.5s`, `500ms`)
- **Frequency**: `NUMBER Hz`, `NUMBER kHz` (e.g., `440 Hz`, `1.5 kHz`)
- **Gain**: `NUMBER`, `NUMBER dB` (e.g., `2`, `6 dB`)

## Window Functions

Available window types for FFT:
- `hamming` - Hamming window (default)
- `hanning` - Hanning window
- `blackman` - Blackman window  
- `rectangular` - Rectangular window

## Filter Types

- `lowpass` - Removes high frequencies
- `highpass` - Removes low frequencies
- `bandpass` - Keeps a frequency band
- `notch` - Removes a specific frequency

## Compression Types

- `mp3` - MPEG Layer 3 compression
- `flac` - Free Lossless Audio Codec
- `opus` - Opus audio codec

## Technical Details

### Fourier Transform
The FFT implementation uses the Cooley-Tukey algorithm for efficient computation. Input signals are automatically padded to the nearest power of 2 for optimal performance.

### Laplace Transform
Numerical Laplace transform is computed at specific frequency points for system analysis. This is particularly useful for:
- System stability analysis
- Transfer function computation
- Control system design

### WAV Format Support
- Sample rates: Any (commonly 44.1kHz, 48kHz)
- Bit depths: 16, 24, 32-bit
- Channels: Mono and stereo

### Filtering
Implements simplified Butterworth filters with configurable cutoff frequencies. The filters use recursive implementations for efficiency.

## Examples

### Basic Audio Processing
```dsl
load "voice.wav"
normalize
filter highpass 200 Hz
filter lowpass 3000 Hz
save "voice_cleaned.wav"
```

### Frequency Analysis
```dsl
load "music.wav"
fft with window blackman
spectrum
phase
```

### Audio Editing
```dsl
load "podcast.wav"
cut from 0s to 5s        # Remove intro
cut from 3600s to 3650s  # Remove outro
normalize
save "podcast_edited.wav"
```

### Compression Workflow
```dsl
load "master.wav"
normalize
compress type flac ratio 4
save "master_compressed.wav"
```

## Architecture

The audio DSL consists of three main components:

1. **DSL Parser** (`audio_dsl.go`): Defines grammar and parsing rules
2. **Audio Engine** (`audio_engine.go`): Implements audio processing algorithms
3. **Main Application** (`main.go`): Example usage and demonstrations

The system uses the go-dsl framework for parsing and can be easily extended with new commands and operations.

## Extending the DSL

To add new operations:

1. Add keywords in `setupGrammar()`
2. Define grammar rules
3. Implement processing in `AudioEngine`
4. Add action handlers

Example:
```go
// Add keyword
ad.dsl.KeywordToken("reverb", "reverb")

// Add rule
ad.dsl.Rule("reverb_cmd", "reverb NUMBER", func(ctx map[string]interface{}, args ...interface{}) (interface{}, error) {
    amount, _ := strconv.ParseFloat(args[1].(string), 64)
    return ad.engine.ApplyReverb(amount)
})

// Implement in engine
func (ae *AudioEngine) ApplyReverb(amount float64) (string, error) {
    // Implementation here
    return fmt.Sprintf("Applied reverb: %.1f%%", amount*100), nil
}
```

## Performance Considerations

- FFT operations are O(n log n) complexity
- Large files are processed in memory
- Resampling uses linear interpolation for speed
- Filters use recursive implementations

## Limitations

- Currently supports mono and stereo only
- Compression simulation (not actual codec implementation)
- Basic filter implementations
- Memory-based processing (not streaming)

## Future Enhancements

Potential additions:
- Real-time streaming support
- More window functions
- Advanced filter designs
- Actual codec implementations
- Multi-channel support
- Convolution reverb
- Pitch shifting
- Time stretching

## License

This example is part of the go-dsl project and follows the same license terms.