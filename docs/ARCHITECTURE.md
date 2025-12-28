# Architecture: First Principles Analysis

## Problem Statement

Build a multi-agent framework operating on signal mechanisms (SignalBase), where agents:
- Receive input signals
- Emit output signals
- Designate the next agent to execute

## Problem Decomposition

### What is the fundamental need?

Using the 5 Whys technique:

| Why | Question | Answer |
|-----|----------|--------|
| 1 | Why a signal-based multi-agent framework? | To coordinate multiple agents working together |
| 2 | Why multiple agents instead of monolithic? | Different capabilities/specializations, task decomposition |
| 3 | Why signals instead of direct function calls? | Loose coupling - agents shouldn't know implementation details |
| 4 | Why loose coupling? | Agents need to be added/removed/modified independently |
| 5 | Why independent modification? | Problem domain is exploratory, optimal topology unknown |

**Root Need**: Dynamic, loosely-coupled orchestration of specialized processors with runtime-configurable workflows.

### Theoretical Minimum

The absolute minimum implementation is ~30 lines:

```go
type Signal struct { Data any }
type Agent interface { Process(Signal) (Signal, string) }
type Router struct { agents map[string]Agent }

func (r *Router) Run(id string, s Signal) {
    for id != "" {
        s, id = r.agents[id].Process(s)
    }
}
```

Everything beyond this must justify its existence.

## Component Justification

### Signal Type

| Field | Justification |
|-------|---------------|
| ID | Tracing, debugging, idempotency |
| Type | Routing decisions, type safety |
| Timestamp | Ordering, debugging, metrics |
| Source | Tracing, debugging |
| Destination | Explicit routing hint |
| Payload | The actual data (required) |
| ParentID | Lineage tracking for complex flows |
| Metadata | Extensibility without struct changes |

### Agent Interface

| Method | Justification |
|--------|---------------|
| ID() | Required for routing and identification |
| Process() | The core transformation (required) |

Minimal interface = maximum flexibility in implementation.

### Router

| Feature | Justification |
|---------|---------------|
| Agent registry | Need to look up agents by ID |
| Rule-based routing | Decouples agents from topology |
| Priority rules | Explicit destination > type-based > default |

### Engine

| Feature | Justification |
|---------|---------------|
| Worker pool | Concurrent processing without goroutine explosion |
| Buffered channel | Backpressure handling |
| Timeout | Prevent stuck agents from blocking system |
| Hooks | Observability without code modification |

## What Was Intentionally Excluded

| Feature | Why Excluded |
|---------|--------------|
| Persistence | Can be added as agent behavior |
| Distributed | Start local, scale later |
| Retry logic | Agent responsibility or middleware |
| Circuit breaker | Can be added as wrapper agent |
| Metrics | Hooks provide integration point |
| Serialization | Only needed for distribution |

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         ENGINE                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Inbox Channel                            ││
│  │              (buffered, backpressure)                       ││
│  └─────────────────────────────────────────────────────────────┘│
│                          │                                      │
│            ┌─────────────┼─────────────┐                       │
│            ▼             ▼             ▼                       │
│       [Worker1]     [Worker2]     [Worker3]                    │
│            │             │             │                       │
│            └─────────────┼─────────────┘                       │
│                          ▼                                      │
│                     [ROUTER]                                    │
│                          │                                      │
└──────────────────────────┼──────────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
    [Agent A]         [Agent B]         [Agent C]
         │                 │                 │
         └─────────────────┼─────────────────┘
                           ▼
                   Output signals
                (re-enter Engine)
```

## Signal Flow

1. **Submit**: External code submits signal to Engine inbox
2. **Route**: Worker pulls signal, Router determines destination(s)
3. **Process**: Agent processes signal, returns AgentResult
4. **Emit**: Output signals re-enter Engine inbox
5. **Repeat**: Until no more signals or terminal agent reached

## Comparison with Alternatives

### vs. Actor Model (Akka-style)

| Aspect | Actor Model | Signal Framework |
|--------|-------------|------------------|
| Message passing | Untyped mailbox | Typed signals |
| Routing | Address-based | Rule-based + explicit |
| State | Actor-internal | Agent-optional |
| Complexity | High (supervision, etc.) | Low (Go primitives) |

### vs. Message Queue (Kafka, RabbitMQ)

| Aspect | Message Queue | Signal Framework |
|--------|---------------|------------------|
| Latency | Higher (network) | Lower (in-process) |
| Persistence | Built-in | Not included |
| Scaling | Horizontal | Vertical (single process) |
| Complexity | Infrastructure needed | No dependencies |

### vs. Direct Function Calls

| Aspect | Function Calls | Signal Framework |
|--------|----------------|------------------|
| Coupling | Tight | Loose |
| Concurrency | Manual | Built-in |
| Tracing | Manual | Automatic (lineage) |
| Reconfiguration | Code change | Runtime |

## Extension Points

The framework is designed for extension without modification:

1. **Custom Agents**: Implement `Agent` interface
2. **Middleware**: Wrap agents with decorator pattern
3. **Routing Rules**: Add functions to Router
4. **Observability**: Use Engine hooks
5. **Persistence**: Create a PersistenceAgent that logs signals
6. **Distribution**: Create network agents that bridge processes

## Design Principles

### Signals are Immutable

Agents receive signals and create new ones. This:
- Prevents race conditions
- Enables safe parallel processing
- Makes debugging easier (signals don't change)

### Routing is External

Agents don't know about other agents. This:
- Enables runtime reconfiguration
- Allows topology changes without code changes
- Makes testing easier (agents are isolated)

### Go Primitives First

Built on channels, goroutines, and context. This:
- Leverages Go's strengths
- Reduces complexity
- Makes behavior predictable to Go developers

### Minimal Interface

The Agent interface has only 2 methods. This:
- Maximizes implementation flexibility
- Reduces learning curve
- Enables functional adapters
