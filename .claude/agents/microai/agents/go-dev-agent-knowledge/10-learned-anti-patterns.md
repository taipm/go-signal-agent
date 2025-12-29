# Learned Anti-Patterns - Go Development

> "Talk is cheap. Show me the code." -- Linus Torvalds

**Purpose:** Anti-patterns discovered through code reviews, bugs, and production issues.
**Last Updated:** 2025-12-29
**Total Anti-Patterns:** 0

---

## TL;DR

- Anti-patterns discovered from REAL mistakes, not theoretical
- Each anti-pattern caused actual problems (bugs, performance, security)
- Includes root cause analysis and correct solution
- Severity based on production impact

---

## Severity Legend

| Level | Impact | Example |
|-------|--------|---------|
| CRITICAL | Production crash, data loss, security breach | Nil pointer panic, SQL injection |
| HIGH | Performance degradation, resource leak | Goroutine leak, memory leak |
| MEDIUM | Maintainability, code quality issues | Magic numbers, unclear naming |
| LOW | Style, minor inefficiencies | Unused imports, formatting |

---

## Concurrency Anti-Patterns

### Anti-Pattern: Signal Handler Without Exit Path

**Discovered:** 2025-12-29
**Source:** Bug fix during testing
**Severity:** CRITICAL
**Occurrences:** Found via race detector
**Learning ID:** L002

**The Problem:**
```go
// BROKEN: Goroutine leaks on graceful shutdown
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh // Blocks forever if no signal!
}()
```

**Why It's Wrong:**
1. Goroutine blocks indefinitely on signal channel
2. If program exits via context cancellation (not signal), goroutine leaks
3. `signal.Stop()` never called, OS resources not released
4. Race detector may flag this in tests

**The Solution:**
```go
// CORRECT: Always have exit path via context
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    defer signal.Stop(sigCh) // Always cleanup!

    select {
    case sig := <-sigCh:
        // Handle signal
    case <-ctx.Done():
        // Exit cleanly on context cancel
        return
    }
}()
```

**Key Rules:**
1. Signal handlers MUST check `ctx.Done()` in select
2. ALWAYS call `signal.Stop(sigCh)` on cleanup (use defer)
3. Buffer signal channel (`make(chan os.Signal, 1)`)

**Detection:** Search for `signal.Notify` without corresponding `signal.Stop` or context check.

**Related:** [02-graceful-shutdown.md](./02-graceful-shutdown.md)

---

<!-- NEW ANTI-PATTERNS WILL BE ADDED BELOW THIS LINE -->

---

## Resource Management Anti-Patterns

### Anti-Pattern: External API Call Without Timeout

**Discovered:** 2025-12-29
**Source:** Code review
**Severity:** HIGH
**Occurrences:** 3+ times caught
**Learning ID:** L001

**The Problem:**
```go
// BROKEN: This can hang forever
resp, err := client.CreateChatCompletion(
    context.Background(), // No timeout!
    request,
)
```

**Why It's Wrong:**
1. External services can be slow or unresponsive
2. Goroutine blocks indefinitely waiting for response
3. Resources (connections, memory) never released
4. One slow call can cascade to system-wide failure

**The Solution:**
```go
// CORRECT: Always timeout external calls
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

resp, err := client.CreateChatCompletion(ctx, request)
```

**Recommended Timeouts:**

| Service Type | Timeout |
|--------------|---------|
| LLM API (OpenAI, Claude) | 30-60s |
| REST API | 10s |
| Database query | 5s |
| Health check | 2s |

**Detection:** Search for `context.Background()` or `context.TODO()` in API call sites.

**Related:** [04-http-patterns.md](./04-http-patterns.md), [06-concurrency.md](./06-concurrency.md)

---

## Error Handling Anti-Patterns

<!-- Anti-patterns for error handling will be added here -->

---

## Security Anti-Patterns

<!-- Anti-patterns for security will be added here -->

---

## Quick Reference

| Anti-Pattern | Category | Severity | Detection Signal |
|--------------|----------|----------|------------------|
| Signal Handler Without Exit Path | Concurrency | CRITICAL | `signal.Notify` without `ctx.Done()` check |
| External API Call Without Timeout | Resource Mgmt | HIGH | `context.Background()` in API calls |

---

## Root Cause Analysis

| Anti-Pattern | Root Cause | Prevention |
|--------------|------------|------------|
| Signal Handler Without Exit Path | Blocking on channel without exit path | Always use `select` with `ctx.Done()` |
| External API Call Without Timeout | No timeout on external call | Always use `context.WithTimeout()` |

---

## Related Knowledge

- [08-anti-patterns.md](./08-anti-patterns.md) - Pre-documented anti-patterns
- [06-concurrency.md](./06-concurrency.md) - Safe concurrency patterns
- [09-learned-patterns.md](./09-learned-patterns.md) - Correct patterns

---

## Statistics

| Metric | Value |
|--------|-------|
| Total Anti-Patterns | 2 |
| Critical | 1 |
| High | 1 |
| Medium | 0 |
| Low | 0 |
| Last Addition | 2025-12-29 (L001) |
