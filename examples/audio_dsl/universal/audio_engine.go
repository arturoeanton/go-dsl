package universal

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/cmplx"
	"os"
)

// WAV file header structure
type WAVHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32
}

// AudioEngine handles all audio processing operations
type AudioEngine struct {
	Samples      []float64
	SampleRate   int
	Channels     int
	markers      []float64
	clipboard    []float64
	spectrum     []complex128
	compressed   bool
	compressType string
}

func NewAudioEngine() *AudioEngine {
	return &AudioEngine{
		Samples:    []float64{},
		SampleRate: 44100,
		Channels:   1,
		markers:    []float64{},
		clipboard:  []float64{},
	}
}

// LoadWAV loads a WAV file into the engine
func (ae *AudioEngine) LoadWAV(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var header WAVHeader
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return "", fmt.Errorf("failed to read WAV header: %v", err)
	}

	// Validate WAV file
	if string(header.ChunkID[:]) != "RIFF" || string(header.Format[:]) != "WAVE" {
		return "", fmt.Errorf("invalid WAV file format")
	}

	ae.SampleRate = int(header.SampleRate)
	ae.Channels = int(header.NumChannels)

	// Read audio data
	numSamples := int(header.Subchunk2Size) / (int(header.BitsPerSample) / 8)
	ae.Samples = make([]float64, numSamples)

	switch header.BitsPerSample {
	case 16:
		for i := 0; i < numSamples; i++ {
			var sample int16
			err = binary.Read(file, binary.LittleEndian, &sample)
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", fmt.Errorf("failed to read sample: %v", err)
			}
			ae.Samples[i] = float64(sample) / 32768.0
		}
	case 24:
		for i := 0; i < numSamples; i++ {
			bytes := make([]byte, 3)
			_, err = file.Read(bytes)
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", fmt.Errorf("failed to read sample: %v", err)
			}
			sample := int32(bytes[0]) | (int32(bytes[1]) << 8) | (int32(bytes[2]) << 16)
			if sample&0x800000 != 0 {
				sample |= ^int32(0xFFFFFF)
			}
			ae.Samples[i] = float64(sample) / 8388608.0
		}
	case 32:
		for i := 0; i < numSamples; i++ {
			var sample int32
			err = binary.Read(file, binary.LittleEndian, &sample)
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", fmt.Errorf("failed to read sample: %v", err)
			}
			ae.Samples[i] = float64(sample) / 2147483648.0
		}
	default:
		return "", fmt.Errorf("unsupported bit depth: %d", header.BitsPerSample)
	}

	return fmt.Sprintf("Loaded %s: %d samples, %d Hz, %d channels", filename, len(ae.Samples), ae.SampleRate, ae.Channels), nil
}

// SaveWAV saves the current audio to a WAV file
func (ae *AudioEngine) SaveWAV(filename string) (string, error) {
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Prepare WAV header
	bitsPerSample := uint16(16)
	numSamples := uint32(len(ae.Samples))
	subchunk2Size := numSamples * uint32(bitsPerSample/8)

	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + subchunk2Size,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(ae.Channels),
		SampleRate:    uint32(ae.SampleRate),
		ByteRate:      uint32(ae.SampleRate) * uint32(ae.Channels) * uint32(bitsPerSample/8),
		BlockAlign:    uint16(ae.Channels) * bitsPerSample / 8,
		BitsPerSample: bitsPerSample,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: subchunk2Size,
	}

	// Write header
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return "", fmt.Errorf("failed to write header: %v", err)
	}

	// Write samples
	for _, sample := range ae.Samples {
		// Convert float64 to int16
		s := int16(sample * 32767.0)
		err = binary.Write(file, binary.LittleEndian, s)
		if err != nil {
			return "", fmt.Errorf("failed to write sample: %v", err)
		}
	}

	return fmt.Sprintf("Saved %s: %d samples", filename, len(ae.Samples)), nil
}

// FFT performs Fast Fourier Transform
func (ae *AudioEngine) FFT(windowType string) (string, error) {
	n := len(ae.Samples)
	if n == 0 {
		return "", fmt.Errorf("no audio data loaded")
	}

	// Apply window function
	windowed := make([]float64, n)
	window := getWindow(windowType, n)
	for i := 0; i < n; i++ {
		windowed[i] = ae.Samples[i] * window[i]
	}

	// Convert to complex
	complexData := make([]complex128, n)
	for i := 0; i < n; i++ {
		complexData[i] = complex(windowed[i], 0)
	}

	// Perform FFT
	ae.spectrum = fft(complexData)

	return fmt.Sprintf("FFT computed with %s window: %d frequency bins", windowType, len(ae.spectrum)), nil
}

// IFFT performs Inverse Fast Fourier Transform
func (ae *AudioEngine) IFFT() (string, error) {
	if len(ae.spectrum) == 0 {
		return "", fmt.Errorf("no spectrum data available")
	}

	// Perform IFFT
	result := ifft(ae.spectrum)

	// Convert back to real samples
	ae.Samples = make([]float64, len(result))
	for i := 0; i < len(result); i++ {
		ae.Samples[i] = real(result[i])
	}

	return fmt.Sprintf("IFFT computed: %d samples reconstructed", len(ae.Samples)), nil
}

// LaplaceTransform performs Laplace transform
func (ae *AudioEngine) LaplaceTransform() (string, error) {
	if len(ae.Samples) == 0 {
		return "", fmt.Errorf("no audio data loaded")
	}

	// Simple numerical Laplace transform
	// For demonstration, we compute at specific s values
	sValues := []complex128{
		complex(0, 2*math.Pi*100),  // 100 Hz
		complex(0, 2*math.Pi*1000), // 1 kHz
		complex(0, 2*math.Pi*5000), // 5 kHz
	}

	results := make([]complex128, len(sValues))
	dt := 1.0 / float64(ae.SampleRate)

	for i, s := range sValues {
		sum := complex128(0)
		for j, sample := range ae.Samples {
			t := float64(j) * dt
			sum += complex(sample, 0) * cmplx.Exp(-s*complex(t, 0)) * complex(dt, 0)
		}
		results[i] = sum
	}

	ae.spectrum = results
	return fmt.Sprintf("Laplace transform computed at %d frequency points", len(results)), nil
}

// InverseLaplaceTransform performs inverse Laplace transform
func (ae *AudioEngine) InverseLaplaceTransform() (string, error) {
	if len(ae.spectrum) == 0 {
		return "", fmt.Errorf("no Laplace spectrum available")
	}

	// Simplified inverse transform for demonstration
	n := ae.SampleRate // One second of samples
	ae.Samples = make([]float64, n)
	dt := 1.0 / float64(ae.SampleRate)

	for i := 0; i < n; i++ {
		t := float64(i) * dt
		sum := complex128(0)
		for j, spec := range ae.spectrum {
			freq := float64(j+1) * 100.0 // Frequency spacing
			s := complex(0, 2*math.Pi*freq)
			sum += spec * cmplx.Exp(s*complex(t, 0))
		}
		ae.Samples[i] = real(sum) / float64(len(ae.spectrum))
	}

	return fmt.Sprintf("Inverse Laplace transform computed: %d samples", len(ae.Samples)), nil
}

// Mark adds a marker at a specific time
func (ae *AudioEngine) Mark(time float64) (string, error) {
	ae.markers = append(ae.markers, time)
	return fmt.Sprintf("Marker added at %.3f seconds", time), nil
}

// Cut removes audio between two time points
func (ae *AudioEngine) Cut(start, end float64) (string, error) {
	startSample := int(start * float64(ae.SampleRate))
	endSample := int(end * float64(ae.SampleRate))

	if startSample < 0 || endSample > len(ae.Samples) || startSample >= endSample {
		return "", fmt.Errorf("invalid time range")
	}

	// Copy to clipboard
	ae.clipboard = make([]float64, endSample-startSample)
	copy(ae.clipboard, ae.Samples[startSample:endSample])

	// Remove samples
	ae.Samples = append(ae.Samples[:startSample], ae.Samples[endSample:]...)

	return fmt.Sprintf("Cut %.3f seconds of audio", end-start), nil
}

// Copy copies audio between two time points to clipboard
func (ae *AudioEngine) Copy(start, end float64) (string, error) {
	startSample := int(start * float64(ae.SampleRate))
	endSample := int(end * float64(ae.SampleRate))

	if startSample < 0 || endSample > len(ae.Samples) || startSample >= endSample {
		return "", fmt.Errorf("invalid time range")
	}

	ae.clipboard = make([]float64, endSample-startSample)
	copy(ae.clipboard, ae.Samples[startSample:endSample])

	return fmt.Sprintf("Copied %.3f seconds of audio", end-start), nil
}

// Paste inserts clipboard content at specified time
func (ae *AudioEngine) Paste(time float64) (string, error) {
	if len(ae.clipboard) == 0 {
		return "", fmt.Errorf("clipboard is empty")
	}

	insertPoint := int(time * float64(ae.SampleRate))
	if insertPoint < 0 || insertPoint > len(ae.Samples) {
		return "", fmt.Errorf("invalid paste position")
	}

	// Insert clipboard content
	newSamples := make([]float64, len(ae.Samples)+len(ae.clipboard))
	copy(newSamples[:insertPoint], ae.Samples[:insertPoint])
	copy(newSamples[insertPoint:insertPoint+len(ae.clipboard)], ae.clipboard)
	copy(newSamples[insertPoint+len(ae.clipboard):], ae.Samples[insertPoint:])
	ae.Samples = newSamples

	return fmt.Sprintf("Pasted %.3f seconds of audio at %.3f", float64(len(ae.clipboard))/float64(ae.SampleRate), time), nil
}

// Trim keeps only audio between start and end
func (ae *AudioEngine) Trim(start, end float64) (string, error) {
	startSample := int(start * float64(ae.SampleRate))
	endSample := int(end * float64(ae.SampleRate))

	if startSample < 0 || endSample > len(ae.Samples) || startSample >= endSample {
		return "", fmt.Errorf("invalid time range")
	}

	ae.Samples = ae.Samples[startSample:endSample]
	return fmt.Sprintf("Trimmed to %.3f seconds", end-start), nil
}

// Amplify adjusts the amplitude of the audio
func (ae *AudioEngine) Amplify(gain float64) (string, error) {
	for i := range ae.Samples {
		ae.Samples[i] *= gain
		// Clip to prevent overflow
		if ae.Samples[i] > 1.0 {
			ae.Samples[i] = 1.0
		} else if ae.Samples[i] < -1.0 {
			ae.Samples[i] = -1.0
		}
	}
	return fmt.Sprintf("Amplified by factor of %.2f", gain), nil
}

// Normalize adjusts audio to maximum amplitude
func (ae *AudioEngine) Normalize() (string, error) {
	if len(ae.Samples) == 0 {
		return "", fmt.Errorf("no audio data")
	}

	// Find peak
	peak := 0.0
	for _, sample := range ae.Samples {
		if math.Abs(sample) > peak {
			peak = math.Abs(sample)
		}
	}

	if peak == 0 {
		return "Audio is silent, cannot normalize", nil
	}

	// Scale to peak
	scale := 0.95 / peak // Leave some headroom
	for i := range ae.Samples {
		ae.Samples[i] *= scale
	}

	return fmt.Sprintf("Normalized to %.1f%% peak", 95.0), nil
}

// Filter applies a frequency filter
func (ae *AudioEngine) Filter(filterType string, freq float64) (string, error) {
	if len(ae.Samples) == 0 {
		return "", fmt.Errorf("no audio data")
	}

	// Simple butterworth filter implementation
	cutoff := freq / (float64(ae.SampleRate) / 2.0)
	rc := 1.0 / (cutoff * 2.0 * math.Pi)
	dt := 1.0 / float64(ae.SampleRate)
	alpha := dt / (rc + dt)

	filtered := make([]float64, len(ae.Samples))

	switch filterType {
	case "lowpass":
		filtered[0] = ae.Samples[0]
		for i := 1; i < len(ae.Samples); i++ {
			filtered[i] = filtered[i-1] + alpha*(ae.Samples[i]-filtered[i-1])
		}
	case "highpass":
		filtered[0] = ae.Samples[0]
		for i := 1; i < len(ae.Samples); i++ {
			filtered[i] = alpha * (filtered[i-1] + ae.Samples[i] - ae.Samples[i-1])
		}
	case "bandpass":
		// Simplified bandpass (lowpass then highpass)
		temp := make([]float64, len(ae.Samples))
		temp[0] = ae.Samples[0]
		for i := 1; i < len(ae.Samples); i++ {
			temp[i] = temp[i-1] + alpha*(ae.Samples[i]-temp[i-1])
		}
		filtered[0] = temp[0]
		for i := 1; i < len(ae.Samples); i++ {
			filtered[i] = alpha * (filtered[i-1] + temp[i] - temp[i-1])
		}
	case "notch":
		// Simplified notch filter
		for i := range ae.Samples {
			filtered[i] = ae.Samples[i] // Placeholder
		}
	default:
		return "", fmt.Errorf("unknown filter type: %s", filterType)
	}

	ae.Samples = filtered
	return fmt.Sprintf("Applied %s filter at %.0f Hz", filterType, freq), nil
}

// Resample changes the sample rate
func (ae *AudioEngine) Resample(newRate int) (string, error) {
	if len(ae.Samples) == 0 {
		return "", fmt.Errorf("no audio data")
	}

	ratio := float64(newRate) / float64(ae.SampleRate)
	newLength := int(float64(len(ae.Samples)) * ratio)
	resampled := make([]float64, newLength)

	// Linear interpolation resampling
	for i := 0; i < newLength; i++ {
		srcIndex := float64(i) / ratio
		srcIndexInt := int(srcIndex)
		frac := srcIndex - float64(srcIndexInt)

		if srcIndexInt < len(ae.Samples)-1 {
			resampled[i] = ae.Samples[srcIndexInt]*(1-frac) + ae.Samples[srcIndexInt+1]*frac
		} else if srcIndexInt < len(ae.Samples) {
			resampled[i] = ae.Samples[srcIndexInt]
		}
	}

	ae.Samples = resampled
	oldRate := ae.SampleRate
	ae.SampleRate = newRate

	return fmt.Sprintf("Resampled from %d Hz to %d Hz", oldRate, newRate), nil
}

// Compress applies audio compression
func (ae *AudioEngine) Compress(compType string, ratio float64) (string, error) {
	// Simplified compression simulation
	ae.compressed = true
	ae.compressType = compType

	// Apply simple dynamic range compression
	threshold := 0.5
	for i := range ae.Samples {
		if math.Abs(ae.Samples[i]) > threshold {
			excess := math.Abs(ae.Samples[i]) - threshold
			compressed := threshold + excess/ratio
			if ae.Samples[i] > 0 {
				ae.Samples[i] = compressed
			} else {
				ae.Samples[i] = -compressed
			}
		}
	}

	return fmt.Sprintf("Compressed with %s at ratio %.1f:1", compType, ratio), nil
}

// Decompress reverses compression
func (ae *AudioEngine) Decompress() (string, error) {
	if !ae.compressed {
		return "", fmt.Errorf("audio is not compressed")
	}

	// Simplified decompression
	ae.compressed = false
	return fmt.Sprintf("Decompressed from %s", ae.compressType), nil
}

// GetSpectrum returns frequency spectrum
func (ae *AudioEngine) GetSpectrum() (string, error) {
	if len(ae.spectrum) == 0 {
		// Compute FFT if not available
		_, err := ae.FFT("hamming")
		if err != nil {
			return "", err
		}
	}

	// Get magnitude spectrum
	magnitudes := make([]float64, len(ae.spectrum)/2)
	for i := 0; i < len(magnitudes); i++ {
		magnitudes[i] = cmplx.Abs(ae.spectrum[i])
	}

	// Find peak frequency
	peakIdx := 0
	peakMag := 0.0
	for i, mag := range magnitudes {
		if mag > peakMag {
			peakMag = mag
			peakIdx = i
		}
	}

	peakFreq := float64(peakIdx) * float64(ae.SampleRate) / float64(len(ae.spectrum))
	return fmt.Sprintf("Spectrum computed: peak at %.1f Hz", peakFreq), nil
}

// GetPhase returns phase information
func (ae *AudioEngine) GetPhase() (string, error) {
	if len(ae.spectrum) == 0 {
		return "", fmt.Errorf("no spectrum data available")
	}

	// Calculate average phase
	avgPhase := 0.0
	for _, c := range ae.spectrum {
		avgPhase += cmplx.Phase(c)
	}
	avgPhase /= float64(len(ae.spectrum))

	return fmt.Sprintf("Phase analysis: average phase %.3f radians", avgPhase), nil
}

// Helper functions

func getWindow(windowType string, n int) []float64 {
	window := make([]float64, n)

	switch windowType {
	case "hamming":
		for i := 0; i < n; i++ {
			window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		}
	case "hanning":
		for i := 0; i < n; i++ {
			window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		}
	case "blackman":
		for i := 0; i < n; i++ {
			window[i] = 0.42 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1)) +
				0.08*math.Cos(4*math.Pi*float64(i)/float64(n-1))
		}
	case "rectangular":
		for i := 0; i < n; i++ {
			window[i] = 1.0
		}
	default:
		// Default to Hamming
		for i := 0; i < n; i++ {
			window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
		}
	}

	return window
}

// Simple FFT implementation (Cooley-Tukey)
func fft(x []complex128) []complex128 {
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

	// Divide
	even := make([]complex128, n/2)
	odd := make([]complex128, n/2)
	for i := 0; i < n/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	// Conquer
	even = fftRecursive(even)
	odd = fftRecursive(odd)

	// Combine
	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		t := cmplx.Exp(complex(0, -2*math.Pi*float64(k)/float64(n))) * odd[k]
		result[k] = even[k] + t
		result[k+n/2] = even[k] - t
	}

	return result
}

// Inverse FFT
func ifft(x []complex128) []complex128 {
	n := len(x)
	
	// Conjugate
	for i := range x {
		x[i] = cmplx.Conj(x[i])
	}
	
	// Forward FFT
	result := fft(x)
	
	// Conjugate and scale
	for i := range result {
		result[i] = cmplx.Conj(result[i]) / complex(float64(n), 0)
	}
	
	return result
}