// Package main demonstrates a text processing pipeline using the signal framework.
//
// Pipeline Flow:
//
//	RawText -> Tokenizer -> Analyzer -> Output
//
// This example shows:
// - Multiple agents working in sequence
// - Signal derivation with lineage tracking
// - Type-based routing rules
// - Observability hooks
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/taipm/go-signal-agent/signal"
)

// =============================================================================
// SIGNAL TYPES
// =============================================================================

const (
	SignalRawText   signal.SignalType = "raw_text"
	SignalTokenized signal.SignalType = "tokenized"
	SignalAnalyzed  signal.SignalType = "analyzed"
)

// =============================================================================
// PAYLOAD TYPES
// =============================================================================

// RawTextPayload is the input payload containing raw text.
type RawTextPayload struct {
	Text string
}

// TokenizedPayload contains the result of tokenization.
type TokenizedPayload struct {
	OriginalText string
	Tokens       []string
}

// AnalyzedPayload contains text analysis results.
type AnalyzedPayload struct {
	Tokens    []string
	WordCount int
	AvgLength float64
	Uppercase int
}

// =============================================================================
// AGENTS
// =============================================================================

// TokenizerAgent breaks raw text into words (tokens).
type TokenizerAgent struct{}

func (a *TokenizerAgent) ID() string { return "tokenizer" }

func (a *TokenizerAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	// Type-safe payload extraction
	payload, ok := sig.Payload.(*RawTextPayload)
	if !ok {
		return signal.Err(fmt.Errorf("expected *RawTextPayload, got %T", sig.Payload))
	}

	// Simple tokenization by whitespace
	tokens := strings.Fields(payload.Text)

	// Derive a new signal with the tokenized result
	// This maintains lineage tracking (ParentID = original signal ID)
	outSignal := sig.Derive(SignalTokenized, &TokenizedPayload{
		OriginalText: payload.Text,
		Tokens:       tokens,
	})

	// Explicitly set next destination (could also rely on routing rules)
	outSignal.Destination = "analyzer"

	log.Printf("[Tokenizer] Processed: %d tokens", len(tokens))

	return signal.OK(outSignal)
}

// AnalyzerAgent performs text analysis on tokenized input.
type AnalyzerAgent struct{}

func (a *AnalyzerAgent) ID() string { return "analyzer" }

func (a *AnalyzerAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	payload, ok := sig.Payload.(*TokenizedPayload)
	if !ok {
		return signal.Err(fmt.Errorf("expected *TokenizedPayload, got %T", sig.Payload))
	}

	// Compute statistics
	var totalLen int
	var uppercase int
	for _, token := range payload.Tokens {
		totalLen += len(token)
		// Check if token is all uppercase (and has letters)
		if isUppercase(token) {
			uppercase++
		}
	}

	avgLen := 0.0
	if len(payload.Tokens) > 0 {
		avgLen = float64(totalLen) / float64(len(payload.Tokens))
	}

	outSignal := sig.Derive(SignalAnalyzed, &AnalyzedPayload{
		Tokens:    payload.Tokens,
		WordCount: len(payload.Tokens),
		AvgLength: avgLen,
		Uppercase: uppercase,
	})
	outSignal.Destination = "output"

	log.Printf("[Analyzer] Analysis: %d words, avg length %.2f, %d uppercase",
		len(payload.Tokens), avgLen, uppercase)

	return signal.OK(outSignal)
}

// isUppercase checks if a token is all uppercase letters.
func isUppercase(token string) bool {
	hasLetter := false
	for _, r := range token {
		if r >= 'a' && r <= 'z' {
			return false // lowercase letter found
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	return hasLetter
}

// OutputAgent is a terminal agent that outputs results.
type OutputAgent struct {
	results chan<- string
}

func NewOutputAgent(results chan<- string) *OutputAgent {
	return &OutputAgent{results: results}
}

func (a *OutputAgent) ID() string { return "output" }

func (a *OutputAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	payload, ok := sig.Payload.(*AnalyzedPayload)
	if !ok {
		return signal.Err(fmt.Errorf("expected *AnalyzedPayload, got %T", sig.Payload))
	}

	// Format the analysis summary
	summary := fmt.Sprintf(`
========================================
ANALYSIS COMPLETE
========================================
Word Count:        %d
Avg Word Length:   %.2f characters
Uppercase Words:   %d
Signal Chain:      %s -> %s
========================================`,
		payload.WordCount,
		payload.AvgLength,
		payload.Uppercase,
		shortenID(sig.ParentID),
		shortenID(sig.ID),
	)

	log.Printf("[Output] Generated summary")

	// Send result to output channel
	select {
	case a.results <- summary:
	case <-ctx.Done():
		return signal.Err(ctx.Err())
	}

	// Terminal agent - no output signals
	return signal.OK()
}

// shortenID returns a shortened version of a signal ID.
func shortenID(id string) string {
	if len(id) <= 15 {
		return id
	}
	return id[:12] + "..."
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	fmt.Println()
	fmt.Println("=== Signal-Based Multi-Agent Framework Demo ===")
	fmt.Println()

	// Channel for receiving final results
	results := make(chan string, 10)

	// Create and configure router
	router := signal.NewRouter()

	// Register agents
	router.Register(&TokenizerAgent{})
	router.Register(&AnalyzerAgent{})
	router.Register(NewOutputAgent(results))

	// Add type-based routing rules
	// This is an alternative to explicit signal.Destination
	router.AddRule(func(sig *signal.Signal) []string {
		switch sig.Type {
		case SignalRawText:
			return []string{"tokenizer"}
		case SignalTokenized:
			return []string{"analyzer"}
		case SignalAnalyzed:
			return []string{"output"}
		default:
			return nil
		}
	})

	// Create engine with custom configuration
	config := signal.DefaultConfig()
	config.WorkerCount = 2 // 2 workers for this simple demo
	config.BufferSize = 50

	engine := signal.NewEngine(config, router)

	// Set up observability hooks
	engine.OnSignalReceived(func(sig *signal.Signal) {
		log.Printf("[Engine] Received: type=%s dest=%s", sig.Type, sig.Destination)
	})

	engine.OnError(func(sig *signal.Signal, err error) {
		log.Printf("[Engine] ERROR: %v (signal=%s)", err, shortenID(sig.ID))
	})

	// Start processing
	engine.Start()
	defer engine.Stop()

	// Show engine stats
	stats := engine.Stats()
	log.Printf("[Engine] Started with %d workers, buffer size %d",
		stats.WorkerCount, stats.BufferSize)

	// Submit test signals
	testTexts := []string{
		"Hello World this is a TEST of the signal framework",
		"GO is a GREAT language for building CONCURRENT systems",
		"The QUICK brown FOX jumps over the lazy DOG",
	}

	for i, text := range testTexts {
		sig := signal.NewSignal(SignalRawText, &RawTextPayload{Text: text})
		if err := engine.Submit(sig); err != nil {
			log.Printf("Failed to submit signal: %v", err)
			continue
		}
		log.Printf("[Main] Submitted text #%d: %q", i+1, text)
	}

	// Wait for results
	fmt.Println()
	log.Println("Waiting for results...")

	timeout := time.After(5 * time.Second)
	received := 0
	expected := len(testTexts)

	for received < expected {
		select {
		case result := <-results:
			fmt.Println(result)
			received++
		case <-timeout:
			log.Printf("Timeout! Received %d/%d results", received, expected)
			return
		}
	}

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}
