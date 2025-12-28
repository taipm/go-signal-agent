// Package signal provides a minimal, first-principles-based multi-agent framework
// built on signal-passing semantics. It leverages Go's native concurrency primitives
// (goroutines, channels) to provide efficient, type-safe agent orchestration.
package signal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// CORE TYPES: The Fundamental Building Blocks
// =============================================================================

// SignalType identifies the type of signal for routing and processing.
// Using a custom type instead of string provides type safety and enables
// compile-time checks for signal type usage.
type SignalType string

// Signal is the fundamental unit of communication between agents.
// It is designed to be immutable after creation - agents should create
// new signals rather than modifying existing ones to maintain data integrity
// and enable safe concurrent processing.
type Signal struct {
	// Identity
	ID        string     // Unique identifier for tracing and debugging
	Type      SignalType // Signal type for routing and type-safe handling
	Timestamp time.Time  // When the signal was created

	// Routing
	Source      string // Agent ID that emitted this signal
	Destination string // Suggested next agent (can be overridden by router)

	// Payload
	Payload any // The actual data - type depends on SignalType

	// Lineage (for debugging and tracing complex workflows)
	ParentID string            // ID of the signal that caused this one
	Metadata map[string]string // Additional context without struct changes
}

// NewSignal creates a new signal with a unique ID and timestamp.
// This is the primary constructor for creating signals.
func NewSignal(signalType SignalType, payload any) *Signal {
	return &Signal{
		ID:        generateID(),
		Type:      signalType,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  make(map[string]string),
	}
}

// WithDestination returns a new signal with the destination set.
// This follows the immutability principle - the original signal is unchanged.
func (s *Signal) WithDestination(dest string) *Signal {
	newSig := *s
	newSig.Destination = dest
	// Deep copy metadata map to maintain immutability
	newSig.Metadata = make(map[string]string)
	for k, v := range s.Metadata {
		newSig.Metadata[k] = v
	}
	return &newSig
}

// WithSource returns a new signal with the source set.
func (s *Signal) WithSource(source string) *Signal {
	newSig := *s
	newSig.Source = source
	newSig.Metadata = make(map[string]string)
	for k, v := range s.Metadata {
		newSig.Metadata[k] = v
	}
	return &newSig
}

// WithMetadata returns a new signal with additional metadata.
// Multiple calls can be chained: signal.WithMetadata("k1", "v1").WithMetadata("k2", "v2")
func (s *Signal) WithMetadata(key, value string) *Signal {
	newSig := *s
	newSig.Metadata = make(map[string]string)
	for k, v := range s.Metadata {
		newSig.Metadata[k] = v
	}
	newSig.Metadata[key] = value
	return &newSig
}

// Derive creates a child signal with lineage tracking.
// The new signal has its own ID but maintains a reference to its parent,
// enabling tracing of signal chains through complex workflows.
func (s *Signal) Derive(signalType SignalType, payload any) *Signal {
	child := NewSignal(signalType, payload)
	child.ParentID = s.ID
	child.Source = s.Destination // The destination of parent becomes source of child
	// Copy metadata from parent
	for k, v := range s.Metadata {
		child.Metadata[k] = v
	}
	return child
}

// String returns a human-readable representation of the signal.
func (s *Signal) String() string {
	return fmt.Sprintf("Signal{id=%s, type=%s, src=%s, dest=%s}",
		truncateID(s.ID), s.Type, s.Source, s.Destination)
}

// =============================================================================
// AGENT: The Processing Unit
// =============================================================================

// AgentResult represents the output of an agent's processing.
// An agent can produce zero or more output signals.
type AgentResult struct {
	Signals []*Signal // Output signals (zero or more)
	Error   error     // Processing error, if any
}

// OK creates a successful result with the given signals.
func OK(signals ...*Signal) AgentResult {
	return AgentResult{Signals: signals}
}

// Err creates an error result.
func Err(err error) AgentResult {
	return AgentResult{Error: err}
}

// Agent is the interface all agents must implement.
// The interface is deliberately minimal to maximize implementation flexibility.
type Agent interface {
	// ID returns the unique identifier for this agent.
	// This ID is used for routing and must be unique within a Router.
	ID() string

	// Process handles an incoming signal and returns results.
	// Context should be used for cancellation and timeout.
	// The agent should NOT block indefinitely - respect context deadlines.
	Process(ctx context.Context, signal *Signal) AgentResult
}

// AgentFunc is a functional adapter for simple agents.
// It allows using a function as an Agent without creating a struct.
type AgentFunc struct {
	id      string
	process func(ctx context.Context, signal *Signal) AgentResult
}

// NewAgentFunc creates an agent from a function.
// Useful for simple agents that don't need to maintain state.
func NewAgentFunc(id string, fn func(ctx context.Context, signal *Signal) AgentResult) *AgentFunc {
	return &AgentFunc{id: id, process: fn}
}

// ID returns the agent's unique identifier.
func (a *AgentFunc) ID() string {
	return a.id
}

// Process delegates to the wrapped function.
func (a *AgentFunc) Process(ctx context.Context, signal *Signal) AgentResult {
	return a.process(ctx, signal)
}

// =============================================================================
// ROUTER: The Decision Point
// =============================================================================

// RoutingRule is a function that determines where a signal should go.
// It returns a list of destination agent IDs, or nil if the rule doesn't apply.
// Multiple agents can be returned for fanout patterns.
type RoutingRule func(signal *Signal) []string

// Router manages agent registration and signal routing.
// It implements a priority-based routing strategy:
// 1. Explicit destination (signal.Destination)
// 2. Rules evaluated in order
// This separation of routing from agents enables loose coupling.
type Router struct {
	mu     sync.RWMutex
	agents map[string]Agent
	rules  []RoutingRule
}

// NewRouter creates a new router with empty agent registry.
func NewRouter() *Router {
	return &Router{
		agents: make(map[string]Agent),
		rules:  make([]RoutingRule, 0),
	}
}

// Register adds an agent to the router.
// If an agent with the same ID exists, it will be replaced.
func (r *Router) Register(agent Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.ID()] = agent
}

// Unregister removes an agent from the router by ID.
func (r *Router) Unregister(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.agents, agentID)
}

// AddRule adds a routing rule. Rules are evaluated in the order they are added.
// The first rule that returns a non-empty destination list wins.
func (r *Router) AddRule(rule RoutingRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
}

// Route determines where a signal should go based on:
// 1. Explicit destination in signal.Destination
// 2. Routing rules evaluated in order
// Returns nil if no valid destination is found.
func (r *Router) Route(signal *Signal) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Priority 1: Explicit destination
	if signal.Destination != "" {
		if _, exists := r.agents[signal.Destination]; exists {
			return []string{signal.Destination}
		}
	}

	// Priority 2: Apply routing rules in order
	for _, rule := range r.rules {
		destinations := rule(signal)
		if len(destinations) > 0 {
			// Validate that destinations exist
			valid := make([]string, 0, len(destinations))
			for _, dest := range destinations {
				if _, exists := r.agents[dest]; exists {
					valid = append(valid, dest)
				}
			}
			if len(valid) > 0 {
				return valid
			}
		}
	}

	return nil
}

// GetAgent returns an agent by ID.
func (r *Router) GetAgent(id string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, exists := r.agents[id]
	return agent, exists
}

// ListAgents returns all registered agent IDs.
func (r *Router) ListAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.agents))
	for id := range r.agents {
		ids = append(ids, id)
	}
	return ids
}

// AgentCount returns the number of registered agents.
func (r *Router) AgentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

var (
	idCounter uint64
	idMu      sync.Mutex
)

// generateID creates a unique signal ID.
// Format: sig-{timestamp_nano}-{counter}
func generateID() string {
	idMu.Lock()
	defer idMu.Unlock()
	idCounter++
	return fmt.Sprintf("sig-%d-%d", time.Now().UnixNano(), idCounter)
}

// truncateID shortens an ID for display purposes.
func truncateID(id string) string {
	if len(id) <= 20 {
		return id
	}
	return id[:17] + "..."
}
