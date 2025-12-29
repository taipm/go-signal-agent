# Learning Log: 2025-12-29

**Session:** test-session-001
**Project:** go-signal-agent
**Agent:** go-dev-agent

---

## Learnings

### 1. Context Timeout for External API Calls

**Type:** anti-pattern
**Severity:** high
**Time:** 09:00

**Context:**
Reviewing code in `examples/chatbot-memory/main.go` that calls OpenAI API without timeout.

**Discovery:**
External API calls without context timeout can hang forever, causing goroutine leaks and resource exhaustion. The OpenAI client accepts context but many developers forget to add timeout.

**Code Example:**
```go
// BEFORE (problematic) - No timeout, can hang forever
func callOpenAI(client *openai.Client, prompt string) (string, error) {
    resp, err := client.CreateChatCompletion(
        context.Background(), // WRONG: no timeout!
        openai.ChatCompletionRequest{
            Model: "gpt-4",
            Messages: []openai.ChatCompletionMessage{
                {Role: "user", Content: prompt},
            },
        },
    )
    return resp.Choices[0].Message.Content, err
}

// AFTER (correct) - Always use timeout for external calls
func callOpenAI(ctx context.Context, client *openai.Client, prompt string) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.ChatCompletionMessage{
            {Role: "user", Content: prompt},
        },
    })
    if err != nil {
        return "", fmt.Errorf("openai call failed: %w", err)
    }
    return resp.Choices[0].Message.Content, nil
}
```

**Key Takeaway:**
ALWAYS wrap external API calls with `context.WithTimeout()` - 30s for LLM calls, 10s for regular APIs.

**Should Escalate to Knowledge Base:** yes
**Reason:** High severity, common mistake, clear fix pattern

---

### 2. Defer Cleanup for Resources

**Type:** pattern
**Severity:** medium
**Time:** 09:15

**Context:**
Implementing file processing in signal package.

**Discovery:**
Using `defer` immediately after resource acquisition ensures cleanup even on panic.

**Code Example:**
```go
// CORRECT pattern
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer f.Close() // Immediately after successful open

    // ... process file
    return nil
}
```

**Key Takeaway:**
`defer` cleanup should be on the NEXT LINE after successful resource acquisition.

**Should Escalate to Knowledge Base:** no
**Reason:** Already documented in 06-concurrency.md, this is reinforcement

---

### 3. Signal Handler Goroutine Leak

**Type:** bugfix
**Severity:** critical
**Time:** 09:30

**Context:**
Found goroutine leak in signal handler during shutdown testing.

**Discovery:**
Signal handler goroutine wasn't exiting when context was cancelled because it was blocked on `signal.Notify` channel.

**Code Example:**
```go
// BEFORE (leaks goroutine)
func handleSignals(ctx context.Context) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh // Blocks forever if no signal received!
        // cleanup...
    }()
}

// AFTER (proper cleanup)
func handleSignals(ctx context.Context) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        defer signal.Stop(sigCh) // Stop receiving signals

        select {
        case <-sigCh:
            // Handle signal
        case <-ctx.Done():
            // Context cancelled, exit cleanly
            return
        }
    }()
}
```

**Key Takeaway:**
Signal handlers MUST have exit path via context cancellation AND call `signal.Stop()` on cleanup.

**Should Escalate to Knowledge Base:** yes
**Reason:** Critical severity, production impact, not in current knowledge base

---

## Session Summary

| Type | Count | Escalate |
|------|-------|----------|
| anti-pattern | 1 | yes |
| pattern | 1 | no |
| bugfix | 1 | yes |

**Total Learnings:** 3
**To Escalate:** 2
