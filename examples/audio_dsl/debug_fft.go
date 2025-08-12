package main

import (
	"fmt"
	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

func main() {
	dsl := dslbuilder.New("TestDSL")
	
	// Keywords
	dsl.KeywordToken("fft", "fft")
	dsl.KeywordToken("with", "with")
	dsl.KeywordToken("window", "window")
	dsl.KeywordToken("hamming", "hamming")
	
	// Rules
	dsl.Rule("program", []string{"fft_cmd"}, "passthrough")
	dsl.Rule("fft_cmd", []string{"fft"}, "fft_simple")
	dsl.Rule("fft_cmd", []string{"fft", "with", "window", "hamming"}, "fft_window")
	
	// Actions
	dsl.Action("passthrough", func(args []interface{}) (interface{}, error) {
		return args[0], nil
	})
	
	dsl.Action("fft_simple", func(args []interface{}) (interface{}, error) {
		return "FFT simple", nil
	})
	
	dsl.Action("fft_window", func(args []interface{}) (interface{}, error) {
		return "FFT with window", nil
	})
	
	// Test parsing
	tests := []string{
		"fft",
		"fft with window hamming",
	}
	
	for _, test := range tests {
		fmt.Printf("\nTesting: %s\n", test)
		result, err := dsl.Parse(test)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else {
			fmt.Printf("  SUCCESS: %v\n", result.Output)
		}
	}
}