package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/memory"
	"github.com/taipm/go-signal-agent/ollama"
	"github.com/taipm/go-signal-agent/signal"
)

// =============================================================================
// SENTINEL ERRORS
// =============================================================================

var (
	// ErrInvalidPayload indicates the signal payload is not of the expected type
	ErrInvalidPayload = errors.New("invalid payload type")
	// ErrNoWorkers indicates no workers are configured or available
	ErrNoWorkers = errors.New("no workers available")
	// ErrLLMUnavailable indicates the LLM service is not responding
	ErrLLMUnavailable = errors.New("LLM service unavailable")
	// ErrRoutingFailed indicates the routing decision could not be parsed
	ErrRoutingFailed = errors.New("routing decision failed")
)

// =============================================================================
// INTERFACES
// =============================================================================

// LLMClient defines the interface for LLM interactions.
// This allows mocking the ollama client in tests.
type LLMClient interface {
	Chat(ctx context.Context, messages []ollama.Message) (string, error)
	SetModel(model string)
}

// =============================================================================
// COORDINATOR AGENT
// =============================================================================

// CoordinatorAgent routes requests to appropriate workers using LLM
type CoordinatorAgent struct {
	id           string
	ollamaClient LLMClient
	config       *config.CoordinatorConfig
}

// NewCoordinatorAgent creates a new coordinator agent
func NewCoordinatorAgent(cfg *config.CoordinatorConfig, client LLMClient) *CoordinatorAgent {
	return &CoordinatorAgent{
		id:           cfg.ID,
		ollamaClient: client,
		config:       cfg,
	}
}

// ID returns the agent's identifier
func (c *CoordinatorAgent) ID() string {
	return c.id
}

// Process analyzes the user request and routes to appropriate workers
func (c *CoordinatorAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	req, ok := sig.Payload.(*UserRequest)
	if !ok {
		return signal.Err(fmt.Errorf("%w: expected *UserRequest", ErrInvalidPayload))
	}

	// Validate workers configuration
	if len(c.config.AvailableWorkers) == 0 {
		return signal.Err(ErrNoWorkers)
	}

	// Build routing prompt
	messages := []ollama.Message{
		ollama.SystemMessage(c.config.SystemPrompt),
		ollama.UserMessage(fmt.Sprintf("User request: %s\nLanguage: %s", req.Message, req.Language)),
	}

	// Get routing decision from LLM
	c.ollamaClient.SetModel(c.config.Model)
	response, err := c.ollamaClient.Chat(ctx, messages)
	if err != nil {
		// Fallback: use all workers on error
		log.Printf("Coordinator LLM error, using fallback: %v", err)
		return c.createTaskAssignment(sig, req, c.config.AvailableWorkers, "fallback due to LLM error")
	}

	// Parse JSON response
	var decision RoutingDecision
	if err := json.Unmarshal([]byte(extractJSON(response)), &decision); err != nil {
		// Fallback: use first worker
		log.Printf("Failed to parse routing decision: %v", err)
		return c.createTaskAssignment(sig, req, []string{c.config.AvailableWorkers[0]}, "fallback due to parse error")
	}

	// Validate workers
	validWorkers := c.validateWorkers(decision.Workers)
	if len(validWorkers) == 0 {
		validWorkers = []string{c.config.AvailableWorkers[0]}
	}

	// Limit to max workers
	if len(validWorkers) > c.config.MaxWorkers {
		validWorkers = validWorkers[:c.config.MaxWorkers]
	}

	return c.createTaskAssignment(sig, req, validWorkers, decision.Reason)
}

func (c *CoordinatorAgent) createTaskAssignment(sig *signal.Signal, req *UserRequest, workers []string, context string) signal.AgentResult {
	taskID := uuid.New().String()[:8]

	assignment := &TaskAssignment{
		TaskID:          taskID,
		OriginalRequest: req,
		SelectedWorkers: workers,
		Context:         context,
	}

	// Create signals for each selected worker
	signals := make([]*signal.Signal, len(workers))
	for i, workerID := range workers {
		signals[i] = sig.Derive(SignalTaskAssignment, assignment).
			WithDestination(workerID).
			WithMetadata("task_id", taskID).
			WithMetadata("worker_index", fmt.Sprintf("%d", i))
	}

	return signal.OK(signals...)
}

func (c *CoordinatorAgent) validateWorkers(workers []string) []string {
	valid := make([]string, 0)
	for _, w := range workers {
		for _, available := range c.config.AvailableWorkers {
			if w == available {
				valid = append(valid, w)
				break
			}
		}
	}
	return valid
}

// extractJSON extracts JSON object from a string that may contain other text
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return "{}"
	}
	return s[start : end+1]
}

// =============================================================================
// WORKER AGENT (Base for all workers)
// =============================================================================

// WorkerAgent is a specialized processor with memory
type WorkerAgent struct {
	id           string
	workerType   string
	config       *config.WorkerConfig
	memoryStore  *memory.Store
	ollamaClient LLMClient
}

// NewWorkerAgent creates a new worker agent
func NewWorkerAgent(cfg *config.WorkerConfig, memStore *memory.Store, client LLMClient) *WorkerAgent {
	return &WorkerAgent{
		id:           cfg.ID,
		workerType:   cfg.ID,
		config:       cfg,
		memoryStore:  memStore,
		ollamaClient: client,
	}
}

// ID returns the agent's identifier
func (w *WorkerAgent) ID() string {
	return w.id
}

// Process handles the task assignment and produces a result
func (w *WorkerAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	assignment, ok := sig.Payload.(*TaskAssignment)
	if !ok {
		return signal.Err(fmt.Errorf("%w: expected *TaskAssignment", ErrInvalidPayload))
	}

	// Build messages with memory context
	messages := w.buildMessagesWithMemory(assignment.OriginalRequest)

	// Call LLM
	w.ollamaClient.SetModel(w.config.Model)
	response, err := w.ollamaClient.Chat(ctx, messages)
	if err != nil {
		return signal.Err(fmt.Errorf("worker %s LLM error: %w", w.id, err))
	}

	// Store in memory
	if w.memoryStore != nil {
		w.memoryStore.Add(memory.Entry{
			Role:    "user",
			Content: assignment.OriginalRequest.Message,
		})
		w.memoryStore.Add(memory.Entry{
			Role:    "assistant",
			Content: response,
		})
	}

	// Create result signal
	result := &WorkerResult{
		TaskID:     assignment.TaskID,
		WorkerID:   w.id,
		Content:    response,
		Confidence: 0.8, // Default confidence
	}

	resultSig := sig.Derive(SignalWorkerResult, result).
		WithDestination("output").
		WithMetadata("task_id", assignment.TaskID).
		WithMetadata("worker_id", w.id)

	return signal.OK(resultSig)
}

func (w *WorkerAgent) buildMessagesWithMemory(req *UserRequest) []ollama.Message {
	messages := make([]ollama.Message, 0)

	// System prompt with memory context
	systemPrompt := w.config.SystemPrompt
	if w.memoryStore != nil {
		recentMemory := w.memoryStore.GetRecent(5)
		if len(recentMemory) > 0 {
			var memoryContext strings.Builder
			memoryContext.WriteString("\n\n--- MEMORY CONTEXT ---\n")
			for _, entry := range recentMemory {
				memoryContext.WriteString(fmt.Sprintf("[%s]: %s\n", entry.Role, truncate(entry.Content, 200)))
			}
			systemPrompt += memoryContext.String()
		}
	}

	messages = append(messages, ollama.SystemMessage(systemPrompt))
	messages = append(messages, ollama.UserMessage(req.Message))

	return messages
}

// GetMemoryStats returns memory statistics
func (w *WorkerAgent) GetMemoryStats() map[string]int {
	if w.memoryStore == nil {
		return map[string]int{"entries": 0, "size": 0}
	}
	return w.memoryStore.Stats()
}

// ClearMemory clears the worker's memory
func (w *WorkerAgent) ClearMemory() {
	if w.memoryStore != nil {
		w.memoryStore.Clear()
	}
}

// =============================================================================
// OUTPUT AGENT
// =============================================================================

// TaskCollector tracks results for a single task
type TaskCollector struct {
	TaskID        string
	ExpectedCount int
	Results       []*WorkerResult
	CreatedAt     time.Time
	mu            sync.Mutex
}

// OutputAgent collects and consolidates worker results
type OutputAgent struct {
	id           string
	config       *config.OutputConfig
	ollamaClient LLMClient
	pendingTasks map[string]*TaskCollector
	mu           sync.RWMutex
	resultChan   chan *signal.Signal // Channel to send final responses
}

// NewOutputAgent creates a new output agent
func NewOutputAgent(cfg *config.OutputConfig, client LLMClient, resultChan chan *signal.Signal) *OutputAgent {
	return &OutputAgent{
		id:           cfg.ID,
		config:       cfg,
		ollamaClient: client,
		pendingTasks: make(map[string]*TaskCollector),
		resultChan:   resultChan,
	}
}

// ID returns the agent's identifier
func (o *OutputAgent) ID() string {
	return o.id
}

// RegisterTask registers a new task with expected worker count
func (o *OutputAgent) RegisterTask(taskID string, expectedCount int) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.pendingTasks[taskID] = &TaskCollector{
		TaskID:        taskID,
		ExpectedCount: expectedCount,
		Results:       make([]*WorkerResult, 0),
		CreatedAt:     time.Now(),
	}
}

// Process receives worker results and consolidates when complete.
// Thread-safe: properly handles concurrent result submissions.
func (o *OutputAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
	result, ok := sig.Payload.(*WorkerResult)
	if !ok {
		return signal.Err(fmt.Errorf("%w: expected *WorkerResult", ErrInvalidPayload))
	}

	// Lock outer mutex and keep it while adding result to prevent race
	// between getting collector and modifying it
	o.mu.Lock()
	collector, exists := o.pendingTasks[result.TaskID]
	if !exists {
		// Task not registered, create a single-result collector
		collector = &TaskCollector{
			TaskID:        result.TaskID,
			ExpectedCount: 1,
			Results:       make([]*WorkerResult, 0, 1),
			CreatedAt:     time.Now(),
		}
		o.pendingTasks[result.TaskID] = collector
	}

	// Add result while still holding outer lock to prevent race with consolidateResults
	collector.Results = append(collector.Results, result)
	receivedCount := len(collector.Results)
	expectedCount := collector.ExpectedCount

	// If all results received, remove from pending before releasing lock
	// This prevents other goroutines from adding more results
	shouldConsolidate := receivedCount >= expectedCount
	if shouldConsolidate {
		delete(o.pendingTasks, collector.TaskID)
	}
	o.mu.Unlock()

	// Check if all results received
	if shouldConsolidate {
		return o.consolidateResults(ctx, sig, collector)
	}

	// Not all results yet, return empty (waiting)
	return signal.OK()
}

func (o *OutputAgent) consolidateResults(ctx context.Context, sig *signal.Signal, collector *TaskCollector) signal.AgentResult {
	// Note: collector has already been removed from pendingTasks in Process()
	// No need to lock collector.mu as we're the only owner now
	results := collector.Results

	var finalContent string
	contributors := make([]string, 0, len(results))

	for _, r := range results {
		contributors = append(contributors, r.WorkerID)
	}

	if len(results) == 1 {
		// Single result, use directly
		finalContent = results[0].Content
	} else {
		// Multiple results, merge
		finalContent = o.mergeResults(ctx, results)
	}

	response := &FinalResponse{
		TaskID:       collector.TaskID,
		Content:      finalContent,
		Contributors: contributors,
	}

	finalSig := sig.Derive(SignalFinalResponse, response).
		WithMetadata("task_id", collector.TaskID).
		WithMetadata("contributors", strings.Join(contributors, ","))

	// Send to result channel for CLI display
	if o.resultChan != nil {
		select {
		case o.resultChan <- finalSig:
		default:
			log.Printf("Warning: result channel full, dropping signal")
		}
	}

	return signal.OK(finalSig)
}

func (o *OutputAgent) mergeResults(ctx context.Context, results []*WorkerResult) string {
	if o.config.MergeStrategy == "template" {
		return o.templateMerge(results)
	}

	// LLM-based merge
	var resultsText strings.Builder
	for _, r := range results {
		resultsText.WriteString(fmt.Sprintf("\n--- From %s ---\n%s\n", r.WorkerID, r.Content))
	}

	messages := []ollama.Message{
		ollama.SystemMessage(o.config.SystemPrompt),
		ollama.UserMessage(fmt.Sprintf("Please consolidate these responses into a coherent answer:\n%s", resultsText.String())),
	}

	o.ollamaClient.SetModel(o.config.Model)
	response, err := o.ollamaClient.Chat(ctx, messages)
	if err != nil {
		log.Printf("LLM merge failed, using template: %v", err)
		return o.templateMerge(results)
	}

	return response
}

func (o *OutputAgent) templateMerge(results []*WorkerResult) string {
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("\n[%s]\n%s\n", r.WorkerID, r.Content))
	}
	return sb.String()
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
