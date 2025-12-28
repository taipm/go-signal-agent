package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/memory"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/testutil"
	"github.com/taipm/go-signal-agent/signal"
)

// =============================================================================
// COORDINATOR AGENT TESTS
// =============================================================================

func TestCoordinatorAgent_ID(t *testing.T) {
	cfg := &config.CoordinatorConfig{ID: "test-coordinator"}
	mock := testutil.NewMockOllamaClient()
	agent := NewCoordinatorAgent(cfg, mock)

	if got := agent.ID(); got != "test-coordinator" {
		t.Errorf("ID() = %q, want %q", got, "test-coordinator")
	}
}

func TestCoordinatorAgent_Process(t *testing.T) {
	tests := []struct {
		name            string
		llmResponse     string
		llmError        error
		request         *UserRequest
		expectedWorkers []string
		wantErr         bool
	}{
		{
			name:            "valid routing to writing worker",
			llmResponse:     `{"workers": ["writing"], "reason": "content creation"}`,
			request:         &UserRequest{Message: "Write an email", Language: "en"},
			expectedWorkers: []string{"writing"},
		},
		{
			name:            "multi-worker routing",
			llmResponse:     `{"workers": ["summary", "translation"], "reason": "needs both"}`,
			request:         &UserRequest{Message: "Summarize and translate", Language: "en"},
			expectedWorkers: []string{"summary", "translation"},
		},
		{
			name:            "LLM error fallback to all workers",
			llmError:        errors.New("connection refused"),
			request:         &UserRequest{Message: "Test", Language: "en"},
			expectedWorkers: []string{"writing", "translation", "summary"},
		},
		{
			name:            "invalid JSON fallback to first worker",
			llmResponse:     "not valid json",
			request:         &UserRequest{Message: "Test", Language: "en"},
			expectedWorkers: []string{"writing"},
		},
		{
			name:            "empty workers in response uses fallback",
			llmResponse:     `{"workers": [], "reason": "empty"}`,
			request:         &UserRequest{Message: "Test", Language: "en"},
			expectedWorkers: []string{"writing"},
		},
		{
			name:            "unknown worker filtered out",
			llmResponse:     `{"workers": ["unknown_worker"], "reason": "test"}`,
			request:         &UserRequest{Message: "Test", Language: "en"},
			expectedWorkers: []string{"writing"}, // fallback
		},
		{
			name:            "max workers limit applied",
			llmResponse:     `{"workers": ["writing", "translation", "summary"], "reason": "too many"}`,
			request:         &UserRequest{Message: "Test", Language: "en"},
			expectedWorkers: []string{"writing", "translation"}, // limited to 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.CoordinatorConfig{
				ID:               "coordinator",
				Model:            "test-model",
				MaxWorkers:       2, // Limit for testing
				AvailableWorkers: []string{"writing", "translation", "summary"},
				SystemPrompt:     "You are a router",
			}

			mock := testutil.NewMockOllamaClient()
			mock.ChatResponse = tt.llmResponse
			mock.ChatError = tt.llmError

			agent := NewCoordinatorAgent(cfg, mock)

			sig := signal.NewSignal(SignalUserRequest, tt.request)
			result := agent.Process(context.Background(), sig)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("Process() expected error, got nil")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("Process() unexpected error: %v", result.Error)
				return
			}

			// Check number of output signals matches expected workers
			if len(result.Signals) != len(tt.expectedWorkers) {
				t.Errorf("Process() returned %d signals, want %d", len(result.Signals), len(tt.expectedWorkers))
				return
			}

			// Verify each signal destination
			for i, sig := range result.Signals {
				if sig.Destination != tt.expectedWorkers[i] {
					t.Errorf("Signal[%d] destination = %q, want %q", i, sig.Destination, tt.expectedWorkers[i])
				}
			}
		})
	}
}

func TestCoordinatorAgent_Process_InvalidPayload(t *testing.T) {
	cfg := &config.CoordinatorConfig{
		ID:               "coordinator",
		AvailableWorkers: []string{"writing"},
	}
	mock := testutil.NewMockOllamaClient()
	agent := NewCoordinatorAgent(cfg, mock)

	// Wrong payload type
	sig := signal.NewSignal(SignalUserRequest, "not a UserRequest")
	result := agent.Process(context.Background(), sig)

	if result.Error == nil {
		t.Error("Process() expected error for invalid payload")
	}
	if !errors.Is(result.Error, ErrInvalidPayload) {
		t.Errorf("Process() error = %v, want ErrInvalidPayload", result.Error)
	}
}

func TestCoordinatorAgent_Process_NoWorkers(t *testing.T) {
	cfg := &config.CoordinatorConfig{
		ID:               "coordinator",
		AvailableWorkers: []string{}, // Empty!
	}
	mock := testutil.NewMockOllamaClient()
	agent := NewCoordinatorAgent(cfg, mock)

	sig := signal.NewSignal(SignalUserRequest, &UserRequest{Message: "test"})
	result := agent.Process(context.Background(), sig)

	if result.Error == nil {
		t.Error("Process() expected error for no workers")
	}
	if !errors.Is(result.Error, ErrNoWorkers) {
		t.Errorf("Process() error = %v, want ErrNoWorkers", result.Error)
	}
}

// =============================================================================
// WORKER AGENT TESTS
// =============================================================================

func TestWorkerAgent_ID(t *testing.T) {
	cfg := &config.WorkerConfig{ID: "test-worker"}
	agent := NewWorkerAgent(cfg, nil, testutil.NewMockOllamaClient())

	if got := agent.ID(); got != "test-worker" {
		t.Errorf("ID() = %q, want %q", got, "test-worker")
	}
}

func TestWorkerAgent_Process(t *testing.T) {
	tests := []struct {
		name        string
		llmResponse string
		llmError    error
		wantErr     bool
	}{
		{
			name:        "successful processing",
			llmResponse: "Here is your content...",
			wantErr:     false,
		},
		{
			name:     "LLM error",
			llmError: errors.New("model not found"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.WorkerConfig{
				ID:           "writing",
				Model:        "test-model",
				SystemPrompt: "You are a writer",
			}

			mock := testutil.NewMockOllamaClient()
			mock.ChatResponse = tt.llmResponse
			mock.ChatError = tt.llmError

			agent := NewWorkerAgent(cfg, nil, mock)

			assignment := &TaskAssignment{
				TaskID:          "task-001",
				OriginalRequest: &UserRequest{Message: "Write email", Language: "en"},
			}
			sig := signal.NewSignal(SignalTaskAssignment, assignment)
			result := agent.Process(context.Background(), sig)

			if tt.wantErr {
				if result.Error == nil {
					t.Error("Process() expected error")
				}
				return
			}

			if result.Error != nil {
				t.Errorf("Process() unexpected error: %v", result.Error)
				return
			}

			if len(result.Signals) != 1 {
				t.Errorf("Process() returned %d signals, want 1", len(result.Signals))
				return
			}

			workerResult, ok := result.Signals[0].Payload.(*WorkerResult)
			if !ok {
				t.Error("Process() signal payload is not WorkerResult")
				return
			}

			if workerResult.Content != tt.llmResponse {
				t.Errorf("WorkerResult.Content = %q, want %q", workerResult.Content, tt.llmResponse)
			}
		})
	}
}

func TestWorkerAgent_Process_WithMemory(t *testing.T) {
	cfg := &config.WorkerConfig{
		ID:           "writing",
		Model:        "test-model",
		SystemPrompt: "You are a writer",
	}

	mock := testutil.NewMockOllamaClient()
	mock.ChatResponse = "Response with context"

	// Create memory store
	store := memory.NewStore("test", "conversation", 100, time.Hour, "")
	store.Add(memory.Entry{Role: "user", Content: "Previous question"})
	store.Add(memory.Entry{Role: "assistant", Content: "Previous answer"})

	agent := NewWorkerAgent(cfg, store, mock)

	assignment := &TaskAssignment{
		TaskID:          "task-001",
		OriginalRequest: &UserRequest{Message: "Follow up question", Language: "en"},
	}
	sig := signal.NewSignal(SignalTaskAssignment, assignment)
	result := agent.Process(context.Background(), sig)

	if result.Error != nil {
		t.Fatalf("Process() unexpected error: %v", result.Error)
	}

	// Check that memory was used in prompt
	lastMessages := mock.GetLastMessages()
	if len(lastMessages) == 0 {
		t.Fatal("No messages sent to LLM")
	}

	systemPrompt := lastMessages[0].Content
	if !containsSubstring(systemPrompt, "MEMORY CONTEXT") {
		t.Error("System prompt should contain MEMORY CONTEXT")
	}

	// Check that new entries were added to memory
	stats := agent.GetMemoryStats()
	if stats["entries"] != 4 { // 2 original + 2 new
		t.Errorf("Memory entries = %d, want 4", stats["entries"])
	}
}

func TestWorkerAgent_Process_InvalidPayload(t *testing.T) {
	cfg := &config.WorkerConfig{ID: "writing"}
	agent := NewWorkerAgent(cfg, nil, testutil.NewMockOllamaClient())

	sig := signal.NewSignal(SignalTaskAssignment, "not a TaskAssignment")
	result := agent.Process(context.Background(), sig)

	if result.Error == nil {
		t.Error("Process() expected error for invalid payload")
	}
	if !errors.Is(result.Error, ErrInvalidPayload) {
		t.Errorf("Process() error = %v, want ErrInvalidPayload", result.Error)
	}
}

func TestWorkerAgent_ClearMemory(t *testing.T) {
	cfg := &config.WorkerConfig{ID: "writing"}
	store := memory.NewStore("test", "conversation", 100, time.Hour, "")
	store.Add(memory.Entry{Role: "user", Content: "Test"})

	agent := NewWorkerAgent(cfg, store, testutil.NewMockOllamaClient())

	// Verify initial state
	if stats := agent.GetMemoryStats(); stats["entries"] != 1 {
		t.Errorf("Initial entries = %d, want 1", stats["entries"])
	}

	agent.ClearMemory()

	if stats := agent.GetMemoryStats(); stats["entries"] != 0 {
		t.Errorf("After clear entries = %d, want 0", stats["entries"])
	}
}

// =============================================================================
// OUTPUT AGENT TESTS
// =============================================================================

func TestOutputAgent_ID(t *testing.T) {
	cfg := &config.OutputConfig{ID: "test-output"}
	agent := NewOutputAgent(cfg, testutil.NewMockOllamaClient(), nil)

	if got := agent.ID(); got != "test-output" {
		t.Errorf("ID() = %q, want %q", got, "test-output")
	}
}

func TestOutputAgent_Process_SingleResult(t *testing.T) {
	cfg := &config.OutputConfig{
		ID:            "output",
		MergeStrategy: "template",
	}

	resultChan := make(chan *signal.Signal, 10)
	agent := NewOutputAgent(cfg, testutil.NewMockOllamaClient(), resultChan)

	workerResult := &WorkerResult{
		TaskID:     "task-001",
		WorkerID:   "writing",
		Content:    "Written content",
		Confidence: 0.8,
	}

	sig := signal.NewSignal(SignalWorkerResult, workerResult)
	result := agent.Process(context.Background(), sig)

	if result.Error != nil {
		t.Fatalf("Process() unexpected error: %v", result.Error)
	}

	// Check result was sent to channel
	select {
	case finalSig := <-resultChan:
		finalResp, ok := finalSig.Payload.(*FinalResponse)
		if !ok {
			t.Error("Result payload is not FinalResponse")
			return
		}
		if finalResp.Content != "Written content" {
			t.Errorf("FinalResponse.Content = %q, want %q", finalResp.Content, "Written content")
		}
		if len(finalResp.Contributors) != 1 || finalResp.Contributors[0] != "writing" {
			t.Errorf("FinalResponse.Contributors = %v, want [writing]", finalResp.Contributors)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for result on channel")
	}
}

func TestOutputAgent_Process_MultipleResults(t *testing.T) {
	cfg := &config.OutputConfig{
		ID:            "output",
		MergeStrategy: "template",
	}

	resultChan := make(chan *signal.Signal, 10)
	agent := NewOutputAgent(cfg, testutil.NewMockOllamaClient(), resultChan)

	// Register task expecting 2 results
	agent.RegisterTask("task-002", 2)

	// First result
	result1 := &WorkerResult{
		TaskID:   "task-002",
		WorkerID: "writing",
		Content:  "Writing content",
	}
	sig1 := signal.NewSignal(SignalWorkerResult, result1)
	agent.Process(context.Background(), sig1)

	// No result on channel yet
	select {
	case <-resultChan:
		t.Error("Should not receive result before all workers complete")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Second result
	result2 := &WorkerResult{
		TaskID:   "task-002",
		WorkerID: "translation",
		Content:  "Translation content",
	}
	sig2 := signal.NewSignal(SignalWorkerResult, result2)
	agent.Process(context.Background(), sig2)

	// Now should receive merged result
	select {
	case finalSig := <-resultChan:
		finalResp := finalSig.Payload.(*FinalResponse)
		if len(finalResp.Contributors) != 2 {
			t.Errorf("FinalResponse.Contributors = %v, want 2 contributors", finalResp.Contributors)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for merged result")
	}
}

func TestOutputAgent_Process_ConcurrentResults(t *testing.T) {
	cfg := &config.OutputConfig{
		ID:            "output",
		MergeStrategy: "template",
	}

	resultChan := make(chan *signal.Signal, 10)
	agent := NewOutputAgent(cfg, testutil.NewMockOllamaClient(), resultChan)

	// Register task expecting 3 results
	agent.RegisterTask("task-003", 3)

	var wg sync.WaitGroup
	workers := []string{"writing", "translation", "summary"}

	// Send results concurrently
	for _, workerID := range workers {
		wg.Add(1)
		go func(wID string) {
			defer wg.Done()
			result := &WorkerResult{
				TaskID:   "task-003",
				WorkerID: wID,
				Content:  wID + " content",
			}
			sig := signal.NewSignal(SignalWorkerResult, result)
			agent.Process(context.Background(), sig)
		}(workerID)
	}

	wg.Wait()

	// Should receive exactly one merged result
	select {
	case finalSig := <-resultChan:
		finalResp := finalSig.Payload.(*FinalResponse)
		if len(finalResp.Contributors) != 3 {
			t.Errorf("FinalResponse.Contributors = %d, want 3", len(finalResp.Contributors))
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for result")
	}

	// Channel should be empty now
	select {
	case <-resultChan:
		t.Error("Should not receive multiple results")
	default:
		// Expected
	}
}

func TestOutputAgent_Process_InvalidPayload(t *testing.T) {
	cfg := &config.OutputConfig{ID: "output"}
	agent := NewOutputAgent(cfg, testutil.NewMockOllamaClient(), nil)

	sig := signal.NewSignal(SignalWorkerResult, "not a WorkerResult")
	result := agent.Process(context.Background(), sig)

	if result.Error == nil {
		t.Error("Process() expected error for invalid payload")
	}
	if !errors.Is(result.Error, ErrInvalidPayload) {
		t.Errorf("Process() error = %v, want ErrInvalidPayload", result.Error)
	}
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`prefix {"key": "value"} suffix`, `{"key": "value"}`},
		{`{"nested": {"inner": "data"}}`, `{"nested": {"inner": "data"}}`},
		{`no json here`, `{}`},
		{`{incomplete`, `{}`},
		{`only closing}`, `{}`},
		{`multiple {first} and {second}`, `{first} and {second}`},
		{``, `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSON(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a longer string", 10, "this is a ..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
