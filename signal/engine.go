package signal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// ENGINE CONFIGURATION
// =============================================================================

// EngineConfig configures the signal engine behavior.
type EngineConfig struct {
	// BufferSize is the capacity of the signal inbox channel.
	// A larger buffer can absorb bursts but uses more memory.
	// 0 means unbuffered (synchronous submission).
	BufferSize int

	// WorkerCount is the number of goroutines processing signals concurrently.
	// More workers increase throughput but also resource usage.
	WorkerCount int

	// ProcessTimeout is the maximum time allowed for a single agent.Process call.
	// Prevents stuck agents from blocking the system indefinitely.
	ProcessTimeout time.Duration
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() EngineConfig {
	return EngineConfig{
		BufferSize:     100,
		WorkerCount:    4,
		ProcessTimeout: 30 * time.Second,
	}
}

// =============================================================================
// ENGINE HOOKS
// =============================================================================

// SignalHook is called when a signal is received by the engine.
type SignalHook func(signal *Signal)

// ProcessedHook is called after an agent processes a signal.
type ProcessedHook func(signal *Signal, result AgentResult)

// ErrorHook is called when an error occurs during signal processing.
type ErrorHook func(signal *Signal, err error)

// =============================================================================
// ENGINE: The Orchestrator
// =============================================================================

// Engine orchestrates concurrent signal processing across agents.
// It manages a pool of worker goroutines that pull signals from a channel
// and route them to the appropriate agents.
type Engine struct {
	config  EngineConfig
	router  *Router
	inbox   chan *Signal
	done    chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.Mutex

	// Hooks for extensibility and observability
	onSignalReceived  SignalHook
	onSignalProcessed ProcessedHook
	onError           ErrorHook
}

// NewEngine creates a new signal engine with the given configuration and router.
func NewEngine(config EngineConfig, router *Router) *Engine {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 1
	}
	if config.ProcessTimeout <= 0 {
		config.ProcessTimeout = 30 * time.Second
	}

	return &Engine{
		config: config,
		router: router,
		inbox:  make(chan *Signal, config.BufferSize),
		done:   make(chan struct{}),
	}
}

// =============================================================================
// HOOK SETTERS
// =============================================================================

// OnSignalReceived sets a hook called when a signal enters the engine.
// Useful for logging, metrics, or tracing.
func (e *Engine) OnSignalReceived(hook SignalHook) {
	e.onSignalReceived = hook
}

// OnSignalProcessed sets a hook called after processing completes.
// Called even if the processing resulted in an error.
func (e *Engine) OnSignalProcessed(hook ProcessedHook) {
	e.onSignalProcessed = hook
}

// OnError sets a hook called when any error occurs.
// This includes routing errors, processing errors, and submission errors.
func (e *Engine) OnError(hook ErrorHook) {
	e.onError = hook
}

// =============================================================================
// LIFECYCLE METHODS
// =============================================================================

// Start begins processing signals with the configured number of workers.
// Calling Start on an already running engine is a no-op.
func (e *Engine) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	// Spin up worker goroutines
	for i := 0; i < e.config.WorkerCount; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
}

// Stop gracefully stops the engine, waiting for all workers to finish.
// Any signals in the inbox will be processed before stopping.
// Calling Stop on a stopped engine is a no-op.
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	e.mu.Unlock()

	close(e.done)
	e.wg.Wait()
}

// IsRunning returns whether the engine is currently running.
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

// =============================================================================
// SIGNAL SUBMISSION
// =============================================================================

// Submit sends a signal into the engine for processing.
// Returns an error if the engine is not running.
// This method blocks if the inbox buffer is full.
func (e *Engine) Submit(signal *Signal) error {
	e.mu.Lock()
	running := e.running
	e.mu.Unlock()

	if !running {
		return fmt.Errorf("engine not running")
	}

	select {
	case e.inbox <- signal:
		return nil
	case <-e.done:
		return fmt.Errorf("engine stopped")
	}
}

// TrySubmit attempts to submit a signal without blocking.
// Returns false if the inbox is full or the engine is stopped.
func (e *Engine) TrySubmit(signal *Signal) bool {
	e.mu.Lock()
	running := e.running
	e.mu.Unlock()

	if !running {
		return false
	}

	select {
	case e.inbox <- signal:
		return true
	default:
		return false
	}
}

// SubmitWithTimeout submits a signal with a timeout.
// Returns an error if submission doesn't complete within the timeout.
func (e *Engine) SubmitWithTimeout(signal *Signal, timeout time.Duration) error {
	e.mu.Lock()
	running := e.running
	e.mu.Unlock()

	if !running {
		return fmt.Errorf("engine not running")
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case e.inbox <- signal:
		return nil
	case <-timer.C:
		return fmt.Errorf("submission timeout after %v", timeout)
	case <-e.done:
		return fmt.Errorf("engine stopped")
	}
}

// =============================================================================
// WORKER IMPLEMENTATION
// =============================================================================

// worker is the main processing loop for each worker goroutine.
func (e *Engine) worker(id int) {
	defer e.wg.Done()

	for {
		select {
		case signal := <-e.inbox:
			e.processSignal(signal)
		case <-e.done:
			// Drain remaining signals in inbox before exiting
			e.drainInbox()
			return
		}
	}
}

// drainInbox processes any remaining signals in the inbox.
func (e *Engine) drainInbox() {
	for {
		select {
		case signal := <-e.inbox:
			e.processSignal(signal)
		default:
			return
		}
	}
}

// processSignal handles routing and processing of a single signal.
func (e *Engine) processSignal(signal *Signal) {
	// Call receive hook
	if e.onSignalReceived != nil {
		e.onSignalReceived(signal)
	}

	// Route the signal to destination(s)
	destinations := e.router.Route(signal)
	if len(destinations) == 0 {
		if e.onError != nil {
			e.onError(signal, fmt.Errorf("no destination for signal type '%s' (id=%s)",
				signal.Type, truncateID(signal.ID)))
		}
		return
	}

	// Process in each destination agent (supports fanout)
	for _, destID := range destinations {
		e.processInAgent(signal, destID)
	}
}

// processInAgent sends a signal to a specific agent for processing.
func (e *Engine) processInAgent(signal *Signal, destID string) {
	agent, exists := e.router.GetAgent(destID)
	if !exists {
		if e.onError != nil {
			e.onError(signal, fmt.Errorf("agent '%s' not found", destID))
		}
		return
	}

	// Create processing context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.ProcessTimeout)
	defer cancel()

	// Update signal destination for this processing
	processingSignal := signal.WithDestination(destID)

	// Execute agent processing
	result := agent.Process(ctx, processingSignal)

	// Call processed hook
	if e.onSignalProcessed != nil {
		e.onSignalProcessed(processingSignal, result)
	}

	// Handle processing error
	if result.Error != nil {
		if e.onError != nil {
			e.onError(processingSignal, result.Error)
		}
		return
	}

	// Submit output signals back to the engine
	for _, outSignal := range result.Signals {
		// Set source to the agent that produced this signal
		outSignal = outSignal.WithSource(destID)
		if err := e.Submit(outSignal); err != nil {
			if e.onError != nil {
				e.onError(outSignal, fmt.Errorf("failed to submit output signal: %w", err))
			}
		}
	}
}

// =============================================================================
// ENGINE STATISTICS
// =============================================================================

// EngineStats contains runtime statistics about the engine.
type EngineStats struct {
	Running     bool          // Whether the engine is running
	WorkerCount int           // Number of worker goroutines
	BufferSize  int           // Configured inbox buffer size
	BufferUsed  int           // Current number of signals in inbox
	Timeout     time.Duration // Processing timeout per agent
}

// Stats returns current engine statistics.
func (e *Engine) Stats() EngineStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return EngineStats{
		Running:     e.running,
		WorkerCount: e.config.WorkerCount,
		BufferSize:  e.config.BufferSize,
		BufferUsed:  len(e.inbox),
		Timeout:     e.config.ProcessTimeout,
	}
}

// Router returns the engine's router for agent management.
func (e *Engine) Router() *Router {
	return e.router
}
