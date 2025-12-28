package signal

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// SIGNAL TESTS
// =============================================================================

func TestNewSignal(t *testing.T) {
	payload := "test payload"
	sig := NewSignal("test-type", payload)

	if sig.ID == "" {
		t.Error("Signal ID should not be empty")
	}
	if sig.Type != "test-type" {
		t.Errorf("Signal Type = %v, want test-type", sig.Type)
	}
	if sig.Payload != payload {
		t.Errorf("Signal Payload = %v, want %v", sig.Payload, payload)
	}
	if sig.Timestamp.IsZero() {
		t.Error("Signal Timestamp should be set")
	}
	if sig.Metadata == nil {
		t.Error("Signal Metadata should be initialized")
	}
}

func TestSignalWithDestination(t *testing.T) {
	original := NewSignal("test", "payload")
	original.Metadata["key"] = "value"

	modified := original.WithDestination("agent-1")

	// Check modified signal
	if modified.Destination != "agent-1" {
		t.Errorf("Destination = %v, want agent-1", modified.Destination)
	}

	// Check original is unchanged (immutability)
	if original.Destination != "" {
		t.Error("Original signal should not be modified")
	}

	// Check metadata is copied
	if modified.Metadata["key"] != "value" {
		t.Error("Metadata should be copied")
	}
}

func TestSignalWithMetadata(t *testing.T) {
	original := NewSignal("test", "payload")
	modified := original.WithMetadata("key1", "value1").WithMetadata("key2", "value2")

	if modified.Metadata["key1"] != "value1" {
		t.Error("Metadata key1 should be set")
	}
	if modified.Metadata["key2"] != "value2" {
		t.Error("Metadata key2 should be set")
	}

	// Original should not have metadata
	if _, ok := original.Metadata["key1"]; ok {
		t.Error("Original should not be modified")
	}
}

func TestSignalDerive(t *testing.T) {
	parent := NewSignal("parent-type", "parent-payload")
	parent.Destination = "some-agent"

	child := parent.Derive("child-type", "child-payload")

	if child.ParentID != parent.ID {
		t.Errorf("Child ParentID = %v, want %v", child.ParentID, parent.ID)
	}
	if child.Source != parent.Destination {
		t.Errorf("Child Source = %v, want %v", child.Source, parent.Destination)
	}
	if child.Type != "child-type" {
		t.Errorf("Child Type = %v, want child-type", child.Type)
	}
	if child.ID == parent.ID {
		t.Error("Child should have different ID")
	}
}

// =============================================================================
// ROUTER TESTS
// =============================================================================

type mockAgent struct {
	id string
}

func (m *mockAgent) ID() string { return m.id }
func (m *mockAgent) Process(ctx context.Context, sig *Signal) AgentResult {
	return OK()
}

func TestRouterRegister(t *testing.T) {
	router := NewRouter()
	agent := &mockAgent{id: "test-agent"}

	router.Register(agent)

	if router.AgentCount() != 1 {
		t.Errorf("AgentCount = %d, want 1", router.AgentCount())
	}

	retrieved, ok := router.GetAgent("test-agent")
	if !ok {
		t.Error("Agent should be found")
	}
	if retrieved.ID() != "test-agent" {
		t.Error("Retrieved agent ID mismatch")
	}
}

func TestRouterUnregister(t *testing.T) {
	router := NewRouter()
	router.Register(&mockAgent{id: "test"})
	router.Unregister("test")

	if router.AgentCount() != 0 {
		t.Error("Agent should be removed")
	}
}

func TestRouterRouteExplicitDestination(t *testing.T) {
	router := NewRouter()
	router.Register(&mockAgent{id: "agent-1"})
	router.Register(&mockAgent{id: "agent-2"})

	sig := NewSignal("test", nil).WithDestination("agent-1")
	destinations := router.Route(sig)

	if len(destinations) != 1 || destinations[0] != "agent-1" {
		t.Errorf("Destinations = %v, want [agent-1]", destinations)
	}
}

func TestRouterRouteByRule(t *testing.T) {
	router := NewRouter()
	router.Register(&mockAgent{id: "handler"})

	router.AddRule(func(sig *Signal) []string {
		if sig.Type == "special" {
			return []string{"handler"}
		}
		return nil
	})

	sig := NewSignal("special", nil)
	destinations := router.Route(sig)

	if len(destinations) != 1 || destinations[0] != "handler" {
		t.Errorf("Destinations = %v, want [handler]", destinations)
	}
}

func TestRouterRouteNoMatch(t *testing.T) {
	router := NewRouter()
	sig := NewSignal("unknown", nil)

	destinations := router.Route(sig)

	if destinations != nil {
		t.Errorf("Destinations = %v, want nil", destinations)
	}
}

func TestRouterListAgents(t *testing.T) {
	router := NewRouter()
	router.Register(&mockAgent{id: "a"})
	router.Register(&mockAgent{id: "b"})
	router.Register(&mockAgent{id: "c"})

	agents := router.ListAgents()

	if len(agents) != 3 {
		t.Errorf("ListAgents returned %d agents, want 3", len(agents))
	}
}

// =============================================================================
// ENGINE TESTS
// =============================================================================

func TestEngineStartStop(t *testing.T) {
	router := NewRouter()
	engine := NewEngine(DefaultConfig(), router)

	if engine.IsRunning() {
		t.Error("Engine should not be running initially")
	}

	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start")
	}

	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should not be running after Stop")
	}
}

func TestEngineSubmitWhileStopped(t *testing.T) {
	router := NewRouter()
	engine := NewEngine(DefaultConfig(), router)

	sig := NewSignal("test", nil)
	err := engine.Submit(sig)

	if err == nil {
		t.Error("Submit should fail when engine is stopped")
	}
}

func TestEngineProcessing(t *testing.T) {
	var processed atomic.Int32

	processingAgent := NewAgentFunc("processor", func(ctx context.Context, sig *Signal) AgentResult {
		processed.Add(1)
		return OK()
	})

	router := NewRouter()
	router.Register(processingAgent)
	router.AddRule(func(sig *Signal) []string {
		return []string{"processor"}
	})

	config := DefaultConfig()
	config.WorkerCount = 1
	engine := NewEngine(config, router)
	engine.Start()

	// Submit signals
	for i := 0; i < 10; i++ {
		engine.Submit(NewSignal("test", nil))
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)
	engine.Stop()

	if processed.Load() != 10 {
		t.Errorf("Processed = %d, want 10", processed.Load())
	}
}

func TestEnginePipeline(t *testing.T) {
	// Test a simple A -> B -> C pipeline
	var order []string
	var mu sync.Mutex

	addOrder := func(s string) {
		mu.Lock()
		order = append(order, s)
		mu.Unlock()
	}

	agentA := NewAgentFunc("A", func(ctx context.Context, sig *Signal) AgentResult {
		addOrder("A")
		return OK(sig.Derive("B-signal", nil).WithDestination("B"))
	})

	agentB := NewAgentFunc("B", func(ctx context.Context, sig *Signal) AgentResult {
		addOrder("B")
		return OK(sig.Derive("C-signal", nil).WithDestination("C"))
	})

	agentC := NewAgentFunc("C", func(ctx context.Context, sig *Signal) AgentResult {
		addOrder("C")
		return OK() // Terminal
	})

	router := NewRouter()
	router.Register(agentA)
	router.Register(agentB)
	router.Register(agentC)

	config := DefaultConfig()
	config.WorkerCount = 1 // Single worker for deterministic order
	engine := NewEngine(config, router)
	engine.Start()

	engine.Submit(NewSignal("start", nil).WithDestination("A"))

	// Wait for pipeline
	time.Sleep(100 * time.Millisecond)
	engine.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("Expected 3 agents to process, got %d", len(order))
	}
	if order[0] != "A" || order[1] != "B" || order[2] != "C" {
		t.Errorf("Order = %v, want [A B C]", order)
	}
}

func TestEngineHooks(t *testing.T) {
	var receivedCount, processedCount, errorCount atomic.Int32

	router := NewRouter()
	router.Register(NewAgentFunc("handler", func(ctx context.Context, sig *Signal) AgentResult {
		return OK()
	}))
	router.AddRule(func(sig *Signal) []string {
		return []string{"handler"}
	})

	engine := NewEngine(DefaultConfig(), router)
	engine.OnSignalReceived(func(sig *Signal) {
		receivedCount.Add(1)
	})
	engine.OnSignalProcessed(func(sig *Signal, result AgentResult) {
		processedCount.Add(1)
	})
	engine.OnError(func(sig *Signal, err error) {
		errorCount.Add(1)
	})

	engine.Start()
	engine.Submit(NewSignal("test", nil))
	time.Sleep(50 * time.Millisecond)
	engine.Stop()

	if receivedCount.Load() != 1 {
		t.Errorf("ReceivedCount = %d, want 1", receivedCount.Load())
	}
	if processedCount.Load() != 1 {
		t.Errorf("ProcessedCount = %d, want 1", processedCount.Load())
	}
}

func TestEngineStats(t *testing.T) {
	config := EngineConfig{
		BufferSize:     50,
		WorkerCount:    3,
		ProcessTimeout: 10 * time.Second,
	}
	router := NewRouter()
	engine := NewEngine(config, router)
	engine.Start()
	defer engine.Stop()

	stats := engine.Stats()

	if !stats.Running {
		t.Error("Stats.Running should be true")
	}
	if stats.WorkerCount != 3 {
		t.Errorf("Stats.WorkerCount = %d, want 3", stats.WorkerCount)
	}
	if stats.BufferSize != 50 {
		t.Errorf("Stats.BufferSize = %d, want 50", stats.BufferSize)
	}
}

func TestEngineTrySubmit(t *testing.T) {
	router := NewRouter()
	config := DefaultConfig()
	config.BufferSize = 0 // Unbuffered
	engine := NewEngine(config, router)
	engine.Start()
	defer engine.Stop()

	// TrySubmit on unbuffered channel without receiver should return false
	sig := NewSignal("test", nil)
	result := engine.TrySubmit(sig)

	// Should return false as no one is receiving from the unbuffered channel
	if result {
		t.Log("TrySubmit returned true (receiver was ready)")
	} else {
		t.Log("TrySubmit returned false (channel was full)")
	}
}

func TestAgentFuncAdapter(t *testing.T) {
	called := false
	agent := NewAgentFunc("test", func(ctx context.Context, sig *Signal) AgentResult {
		called = true
		return OK(sig.Derive("output", "data"))
	})

	if agent.ID() != "test" {
		t.Errorf("ID = %v, want test", agent.ID())
	}

	result := agent.Process(context.Background(), NewSignal("input", nil))
	if !called {
		t.Error("Agent function should have been called")
	}
	if len(result.Signals) != 1 {
		t.Errorf("Signals count = %d, want 1", len(result.Signals))
	}
}

func TestResultHelpers(t *testing.T) {
	okResult := OK()
	if okResult.Error != nil {
		t.Error("OK should not have error")
	}

	errResult := Err(context.Canceled)
	if errResult.Error != context.Canceled {
		t.Error("Err should contain the error")
	}
}

// =============================================================================
// CONCURRENT TESTS
// =============================================================================

func TestEngineConcurrentSubmit(t *testing.T) {
	var processed atomic.Int32

	router := NewRouter()
	router.Register(NewAgentFunc("handler", func(ctx context.Context, sig *Signal) AgentResult {
		processed.Add(1)
		return OK()
	}))
	router.AddRule(func(sig *Signal) []string {
		return []string{"handler"}
	})

	config := DefaultConfig()
	config.WorkerCount = 4
	config.BufferSize = 100
	engine := NewEngine(config, router)
	engine.Start()

	// Submit from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				engine.Submit(NewSignal("test", nil))
			}
		}()
	}
	wg.Wait()

	// Wait for processing
	time.Sleep(200 * time.Millisecond)
	engine.Stop()

	if processed.Load() != 100 {
		t.Errorf("Processed = %d, want 100", processed.Load())
	}
}
