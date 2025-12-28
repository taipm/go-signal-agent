// Package testutil provides testing utilities and mocks for the multi-agent-orchestrator.
package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/taipm/go-signal-agent/ollama"
)

// =============================================================================
// MOCK OLLAMA CLIENT
// =============================================================================

// MockOllamaClient is a configurable mock for testing LLM interactions.
// Thread-safe: can be used from multiple goroutines.
type MockOllamaClient struct {
	mu sync.Mutex

	// Default response configuration
	ChatResponse string
	ChatError    error

	// Call tracking
	CallCount    int
	LastMessages []ollama.Message

	// Sequential responses (for multi-call scenarios)
	Responses   []string
	responseIdx int

	// Response delay for timeout testing
	Delay time.Duration

	// Custom handler for complex test scenarios
	Handler func(ctx context.Context, messages []ollama.Message) (string, error)

	// Model tracking
	currentModel string
}

// NewMockOllamaClient creates a new mock client with default settings.
func NewMockOllamaClient() *MockOllamaClient {
	return &MockOllamaClient{
		currentModel: "mock-model",
	}
}

// Chat implements the ollama.Client Chat interface.
func (m *MockOllamaClient) Chat(ctx context.Context, messages []ollama.Message) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallCount++
	m.LastMessages = messages

	// Custom handler takes precedence
	if m.Handler != nil {
		return m.Handler(ctx, messages)
	}

	// Simulate delay
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// Sequential responses
	if len(m.Responses) > 0 && m.responseIdx < len(m.Responses) {
		resp := m.Responses[m.responseIdx]
		m.responseIdx++
		return resp, nil
	}

	return m.ChatResponse, m.ChatError
}

// SetModel implements the ollama.Client SetModel interface.
func (m *MockOllamaClient) SetModel(model string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentModel = model
}

// Model returns the currently set model.
func (m *MockOllamaClient) Model() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentModel
}

// Reset clears all call tracking and resets sequential responses.
func (m *MockOllamaClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CallCount = 0
	m.LastMessages = nil
	m.responseIdx = 0
}

// GetCallCount returns the number of times Chat was called.
func (m *MockOllamaClient) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CallCount
}

// GetLastMessages returns the messages from the last Chat call.
func (m *MockOllamaClient) GetLastMessages() []ollama.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to prevent race conditions
	if m.LastMessages == nil {
		return nil
	}
	result := make([]ollama.Message, len(m.LastMessages))
	copy(result, m.LastMessages)
	return result
}

// WithResponse sets a single response and returns the mock for chaining.
func (m *MockOllamaClient) WithResponse(response string) *MockOllamaClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ChatResponse = response
	return m
}

// WithError sets an error response and returns the mock for chaining.
func (m *MockOllamaClient) WithError(err error) *MockOllamaClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ChatError = err
	return m
}

// WithResponses sets sequential responses and returns the mock for chaining.
func (m *MockOllamaClient) WithResponses(responses ...string) *MockOllamaClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses = responses
	m.responseIdx = 0
	return m
}

// WithDelay sets a response delay and returns the mock for chaining.
func (m *MockOllamaClient) WithDelay(delay time.Duration) *MockOllamaClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Delay = delay
	return m
}

// WithHandler sets a custom handler and returns the mock for chaining.
func (m *MockOllamaClient) WithHandler(handler func(ctx context.Context, messages []ollama.Message) (string, error)) *MockOllamaClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Handler = handler
	return m
}
