# Go Signal Agent

A minimal, first-principles-based multi-agent framework for Go, built on signal-passing semantics.

## Philosophy

This framework was designed by questioning every assumption about agent frameworks:

| Common Assumption | Our Challenge | Our Solution |
|-------------------|---------------|--------------|
| Agents need complex message protocols | What is the minimum viable signal? | Immutable `Signal` struct with payload, type, and routing |
| Routing should be inside agents | Should agents know about topology? | External `Router` with configurable rules |
| Need heavyweight orchestration | What is the minimum orchestration? | Simple `Engine` with goroutine workers |
| Framework should handle everything | What can we delegate to Go primitives? | Leverage channels, contexts, goroutines |

## Installation

```bash
go get github.com/taipm/go-signal-agent
```

## Core Concepts

### Signal
The fundamental unit of communication. Immutable after creation.

```go
// Create a signal
sig := signal.NewSignal(SignalType, payload)

// Chain modifications (returns new signal)
sig = sig.WithDestination("next-agent")
sig = sig.WithMetadata("key", "value")

// Create child signal (preserves lineage)
child := sig.Derive(NewType, newPayload)
```

### Agent
A processing unit that transforms signals.

```go
type Agent interface {
    ID() string
    Process(ctx context.Context, signal *Signal) AgentResult
}
```

### Router
Determines signal routing. Supports:
- Explicit destination (`signal.Destination`)
- Rule-based routing (by signal type, payload, etc.)
- Multi-cast (one signal to multiple agents)

### Engine
Orchestrates concurrent processing with:
- Worker pool (configurable)
- Buffered signal queue
- Timeout handling
- Observability hooks

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-signal-agent/signal"
)

const MySignalType signal.SignalType = "my-signal"

type MyPayload struct {
    Message string
}

// Define your agent
type MyAgent struct{}

func (a *MyAgent) ID() string { return "my-agent" }

func (a *MyAgent) Process(ctx context.Context, sig *signal.Signal) signal.AgentResult {
    payload := sig.Payload.(*MyPayload)
    fmt.Printf("Received: %s\n", payload.Message)

    // Terminal agent - no output signals
    return signal.OK()
}

func main() {
    router := signal.NewRouter()
    router.Register(&MyAgent{})

    // Route by signal type
    router.AddRule(func(sig *signal.Signal) []string {
        if sig.Type == MySignalType {
            return []string{"my-agent"}
        }
        return nil
    })

    engine := signal.NewEngine(signal.DefaultConfig(), router)
    engine.Start()
    defer engine.Stop()

    sig := signal.NewSignal(MySignalType, &MyPayload{Message: "Hello!"})
    engine.Submit(sig)
}
```

## Pipeline Example

See [examples/text-pipeline](./examples/text-pipeline/) for a complete multi-agent pipeline:

```
RawText -> Tokenizer -> Analyzer -> Output
```

Run it:
```bash
go run ./examples/text-pipeline/
```

## Design Principles

1. **Signals are immutable** - Create new signals, don't modify existing ones
2. **Agents are stateless (by default)** - Side effects through output signals
3. **Routing is external** - Agents don't need to know the topology
4. **Go primitives first** - Use channels and goroutines, not reinvent them
5. **Composition over inheritance** - Small interfaces, functional adapters

## API Reference

### Signal

```go
// Create signals
NewSignal(signalType SignalType, payload any) *Signal

// Modify signals (returns new signal)
(s *Signal) WithDestination(dest string) *Signal
(s *Signal) WithSource(source string) *Signal
(s *Signal) WithMetadata(key, value string) *Signal
(s *Signal) Derive(signalType SignalType, payload any) *Signal
```

### Agent

```go
// Interface
type Agent interface {
    ID() string
    Process(ctx context.Context, signal *Signal) AgentResult
}

// Functional adapter
NewAgentFunc(id string, fn func(ctx, signal) AgentResult) *AgentFunc

// Result helpers
OK(signals ...*Signal) AgentResult
Err(err error) AgentResult
```

### Router

```go
NewRouter() *Router
(r *Router) Register(agent Agent)
(r *Router) Unregister(agentID string)
(r *Router) AddRule(rule RoutingRule)
(r *Router) Route(signal *Signal) []string
(r *Router) GetAgent(id string) (Agent, bool)
(r *Router) ListAgents() []string
```

### Engine

```go
NewEngine(config EngineConfig, router *Router) *Engine
DefaultConfig() EngineConfig

// Lifecycle
(e *Engine) Start()
(e *Engine) Stop()
(e *Engine) IsRunning() bool

// Submit signals
(e *Engine) Submit(signal *Signal) error
(e *Engine) TrySubmit(signal *Signal) bool
(e *Engine) SubmitWithTimeout(signal *Signal, timeout time.Duration) error

// Hooks
(e *Engine) OnSignalReceived(hook func(*Signal))
(e *Engine) OnSignalProcessed(hook func(*Signal, AgentResult))
(e *Engine) OnError(hook func(*Signal, error))

// Stats
(e *Engine) Stats() EngineStats
```

## Configuration

```go
type EngineConfig struct {
    BufferSize     int           // Inbox channel buffer (default: 100)
    WorkerCount    int           // Worker goroutines (default: 4)
    ProcessTimeout time.Duration // Per-agent timeout (default: 30s)
}
```

## Project Structure

```
.
├── signal/
│   ├── signal.go    # Core types: Signal, Agent, Router
│   └── engine.go    # Engine: orchestration and workers
├── examples/
│   └── text-pipeline/  # Example multi-agent pipeline
├── docs/
│   └── ARCHITECTURE.md # Design rationale
└── README.md
```

## License

MIT
