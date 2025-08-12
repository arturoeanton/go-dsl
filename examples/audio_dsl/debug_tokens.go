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
	
	// Debug tokenization
	input := "fft with window hamming"
	fmt.Printf("Input: %s\n", input)
	
	tokens, err := dsl.DebugTokens(input)
	if err != nil {
		fmt.Printf("Tokenization error: %v\n", err)
		return
	}
	
	fmt.Println("\nTokens:")
	for i, tok := range tokens {
		fmt.Printf("  %d: Type=%s, Value=%s\n", i, tok.TokenType, tok.Value)
	}
}