# Ollama Local LLM Client Patterns

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## TL;DR - Quick Start

```go
// Create client (connects to localhost:11434)
client := ollama.NewDefaultClient()

// Check if Ollama is running
if !client.IsAvailable(ctx) {
    log.Fatal("Ollama is not running. Start with: ollama serve")
}

// Simple generation
response, err := client.Generate(ctx, "Hello, how are you?")

// Streaming (recommended for UX)
err := client.GenerateStream(ctx, "Tell me a story", func(chunk string) error {
    fmt.Print(chunk)  // Print as tokens arrive
    return nil
})
```

---

## What is Ollama?

Ollama là local LLM runtime cho phép chạy các model AI trên máy local:

| Feature | Benefit |
|---------|---------|
| **Local execution** | Không cần API key, không mất phí |
| **Privacy** | Data không rời khỏi máy |
| **Speed** | Không latency mạng |
| **Models** | Qwen, Llama, Mistral, CodeLlama, Phi, etc. |

**API Documentation:** https://github.com/ollama/ollama/blob/main/docs/api.md

---

## Installation & Setup

### 1. Install Ollama

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows
# Download from https://ollama.com/download
```

### 2. Start Ollama Server

```bash
# Start server (runs on localhost:11434)
ollama serve

# Or run in background
ollama serve &
```

### 3. Pull Models

```bash
# Pull a model
ollama pull qwen3:1.7b
ollama pull llama3.2:3b
ollama pull codellama:7b
ollama pull mistral:7b

# List available models
ollama list
```

---

## Client Configuration

### Default Client

```go
// Uses localhost:11434 and qwen3:1.7b
client := ollama.NewDefaultClient()
```

### Custom Configuration

```go
config := ollama.ClientConfig{
    Endpoint: "http://localhost:11434",  // Ollama server URL
    Model:    "llama3.2:3b",             // Model to use
    Timeout:  120 * time.Second,         // Request timeout
}
client := ollama.NewClient(config)
```

### Remote Ollama Server

```go
// Connect to remote Ollama instance
config := ollama.ClientConfig{
    Endpoint: "http://192.168.1.100:11434",
    Model:    "qwen3:1.7b",
}
client := ollama.NewClient(config)
```

---

## Core Types

### Message Types

```go
// Role constants
const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
)

// Message structure
type Message struct {
    Role    Role   `json:"role"`
    Content string `json:"content"`
}

// Helper constructors
msg := ollama.UserMessage("Hello!")
msg := ollama.AssistantMessage("Hi there!")
msg := ollama.SystemMessage("You are a helpful assistant.")
```

### Model Options

```go
type ModelOptions struct {
    Temperature   float64  // 0.0 = deterministic, 1.0 = creative
    TopP          float64  // Nucleus sampling parameter
    TopK          int      // Vocabulary limit for sampling
    NumPredict    int      // Max tokens to generate
    Stop          []string // Stop sequences
    RepeatPenalty float64  // Penalize repetition
    Seed          int      // For reproducibility (-1 = random)
}
```

---

## API Patterns

### Pattern 1: Simple Generation

```go
func generate(ctx context.Context, prompt string) (string, error) {
    client := ollama.NewDefaultClient()

    response, err := client.Generate(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("generate failed: %w", err)
    }

    return response, nil
}
```

### Pattern 2: Streaming Generation (Recommended)

```go
func generateStream(ctx context.Context, prompt string) error {
    client := ollama.NewDefaultClient()

    err := client.GenerateStream(ctx, prompt, func(chunk string) error {
        fmt.Print(chunk)  // Print each token as it arrives
        return nil        // Return error to stop early
    })

    if err != nil {
        return fmt.Errorf("stream failed: %w", err)
    }

    fmt.Println()  // Newline after complete response
    return nil
}
```

### Pattern 3: Chat with History

```go
type Conversation struct {
    client   *ollama.Client
    messages []ollama.Message
}

func NewConversation(systemPrompt string) *Conversation {
    return &Conversation{
        client: ollama.NewDefaultClient(),
        messages: []ollama.Message{
            ollama.SystemMessage(systemPrompt),
        },
    }
}

func (c *Conversation) Send(ctx context.Context, userMessage string) (string, error) {
    // Add user message
    c.messages = append(c.messages, ollama.UserMessage(userMessage))

    // Get response
    response, err := c.client.Chat(ctx, c.messages)
    if err != nil {
        return "", err
    }

    // Add assistant response to history
    c.messages = append(c.messages, ollama.AssistantMessage(response))

    return response, nil
}
```

### Pattern 4: Streaming Chat

```go
func (c *Conversation) SendStream(ctx context.Context, userMessage string, callback func(string) error) error {
    c.messages = append(c.messages, ollama.UserMessage(userMessage))

    var response strings.Builder

    err := c.client.ChatStream(ctx, c.messages, func(chunk string) error {
        response.WriteString(chunk)
        return callback(chunk)  // Pass to caller
    })

    if err != nil {
        return err
    }

    // Add complete response to history
    c.messages = append(c.messages, ollama.AssistantMessage(response.String()))
    return nil
}
```

---

## Streaming Implementation Details

### Why Streaming?

| Non-Streaming | Streaming |
|---------------|-----------|
| Wait 10+ seconds for full response | See first token in ~100ms |
| Bad UX for long responses | Feels responsive and fast |
| All-or-nothing | Can cancel mid-generation |

### StreamCallback Pattern

```go
// Callback signature
type StreamCallback func(chunk string) error

// Usage patterns
client.GenerateStream(ctx, prompt, func(chunk string) error {
    // 1. Print to stdout
    fmt.Print(chunk)
    return nil
})

client.GenerateStream(ctx, prompt, func(chunk string) error {
    // 2. Send to channel
    select {
    case outputChan <- chunk:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
})

client.GenerateStream(ctx, prompt, func(chunk string) error {
    // 3. Stop on keyword
    if strings.Contains(chunk, "STOP") {
        return errors.New("stop requested")
    }
    fmt.Print(chunk)
    return nil
})
```

### Collecting Full Response from Stream

```go
// strings.Builder for O(n) performance
func generateCollected(ctx context.Context, prompt string) (string, error) {
    client := ollama.NewDefaultClient()

    var builder strings.Builder

    err := client.GenerateStream(ctx, prompt, func(chunk string) error {
        builder.WriteString(chunk)
        return nil
    })

    return builder.String(), err
}
```

---

## Context & Cancellation

### ❌ BROKEN: No Cancellation

```go
// 🔴 BROKEN — Cannot stop if user presses Ctrl+C
response, _ := client.Generate(context.Background(), prompt)
```

### ✅ ĐÚNG: With Context

```go
func generate(prompt string) (string, error) {
    // Create cancellable context
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    return client.Generate(ctx, prompt)
}

// Or with signal handling
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // This will cancel if user presses Ctrl+C
    client.GenerateStream(ctx, prompt, callback)
}
```

---

## Error Handling

### Check Availability First

```go
func ensureOllamaRunning(ctx context.Context) error {
    client := ollama.NewDefaultClient()

    if !client.IsAvailable(ctx) {
        return errors.New("Ollama is not running. Start with: ollama serve")
    }

    return nil
}
```

### Handle API Errors

```go
func generate(ctx context.Context, prompt string) (string, error) {
    response, err := client.Generate(ctx, prompt)
    if err != nil {
        // Check for context cancellation
        if errors.Is(err, context.Canceled) {
            return "", fmt.Errorf("generation cancelled: %w", err)
        }
        if errors.Is(err, context.DeadlineExceeded) {
            return "", fmt.Errorf("generation timed out: %w", err)
        }

        // Check for connection errors
        if strings.Contains(err.Error(), "connection refused") {
            return "", errors.New("Ollama server not running")
        }

        return "", fmt.Errorf("generation failed: %w", err)
    }

    return response, nil
}
```

---

## Thread Safety

The client is **thread-safe** for concurrent use:

```go
client := ollama.NewDefaultClient()

// Safe to call from multiple goroutines
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        response, _ := client.Generate(ctx, fmt.Sprintf("Prompt %d", id))
        fmt.Printf("Response %d: %s\n", id, response)
    }(i)
}
wg.Wait()

// Safe to change model while requests in flight
client.SetModel("llama3.2:3b")
```

---

## Model Management

### List Available Models

```go
func listModels(ctx context.Context) error {
    client := ollama.NewDefaultClient()

    models, err := client.ListModels(ctx)
    if err != nil {
        return err
    }

    for _, model := range models {
        fmt.Printf("%s (%.2f GB)\n", model.Name, float64(model.Size)/(1024*1024*1024))
    }

    return nil
}
```

### Switch Models Dynamically

```go
func chatWithModel(ctx context.Context, model, prompt string) (string, error) {
    client := ollama.NewDefaultClient()
    client.SetModel(model)

    return client.Generate(ctx, prompt)
}

// Usage
response, _ := chatWithModel(ctx, "codellama:7b", "Write a Go function")
response, _ := chatWithModel(ctx, "llama3.2:3b", "Explain this code")
```

---

## Common Models

| Model | Size | Best For |
|-------|------|----------|
| `qwen3:1.7b` | 1.7B | Fast, general purpose |
| `qwen3:8b` | 8B | Better quality, slower |
| `llama3.2:3b` | 3B | Good balance |
| `llama3.2:8b` | 8B | High quality |
| `codellama:7b` | 7B | Code generation |
| `mistral:7b` | 7B | Instruction following |
| `phi3:mini` | 3.8B | Efficient, good reasoning |

```bash
# Pull specific models
ollama pull qwen3:1.7b
ollama pull codellama:7b

# Check model info
ollama show qwen3:1.7b
```

---

## Integration with Signal Engine

```go
type OllamaAgent struct {
    client *ollama.Client
}

func NewOllamaAgent() *OllamaAgent {
    return &OllamaAgent{
        client: ollama.NewDefaultClient(),
    }
}

func (a *OllamaAgent) Name() string { return "ollama" }

func (a *OllamaAgent) Process(ctx context.Context, sig *signal.Signal) (*signal.Signal, error) {
    prompt, ok := sig.Payload.(string)
    if !ok {
        return nil, errors.New("payload must be string")
    }

    var response strings.Builder

    err := a.client.GenerateStream(ctx, prompt, func(chunk string) error {
        response.WriteString(chunk)
        return nil
    })

    if err != nil {
        return nil, err
    }

    return signal.NewSignal("llm_response", response.String()), nil
}
```

---

## Interactive CLI Pattern

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    client := ollama.NewDefaultClient()

    if !client.IsAvailable(ctx) {
        log.Fatal("Ollama not running. Start with: ollama serve")
    }

    conversation := []ollama.Message{
        ollama.SystemMessage("You are a helpful assistant."),
    }

    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("> ")

    for scanner.Scan() {
        input := strings.TrimSpace(scanner.Text())
        if input == "" {
            fmt.Print("> ")
            continue
        }
        if input == "exit" || input == "quit" {
            break
        }

        conversation = append(conversation, ollama.UserMessage(input))

        var response strings.Builder
        err := client.ChatStream(ctx, conversation, func(chunk string) error {
            fmt.Print(chunk)
            response.WriteString(chunk)
            return nil
        })

        if err != nil {
            fmt.Printf("\nError: %v\n", err)
        } else {
            conversation = append(conversation,
                ollama.AssistantMessage(response.String()))
        }

        fmt.Print("\n> ")
    }
}
```

---

## Best Practices Checklist

### Setup
- [ ] Ollama server running (`ollama serve`)
- [ ] Model pulled (`ollama pull <model>`)
- [ ] Check availability before first request

### Performance
- [ ] Use streaming for better UX
- [ ] Use `strings.Builder` for collecting response
- [ ] Set appropriate timeout (120s for long responses)

### Reliability
- [ ] Context with cancellation support
- [ ] Handle connection refused errors
- [ ] Check for empty responses

### Thread Safety
- [ ] Client safe for concurrent use
- [ ] Use `SetModel()` for dynamic switching
- [ ] Don't share message slices between goroutines

---

## Quick Reference

| Operation | Method |
|-----------|--------|
| Simple generation | `client.Generate(ctx, prompt)` |
| Streaming generation | `client.GenerateStream(ctx, prompt, callback)` |
| Chat (non-streaming) | `client.Chat(ctx, messages)` |
| Chat (streaming) | `client.ChatStream(ctx, messages, callback)` |
| List models | `client.ListModels(ctx)` |
| Check availability | `client.IsAvailable(ctx)` |
| Get current model | `client.Model()` |
| Change model | `client.SetModel(model)` |

| Message Helper | Usage |
|----------------|-------|
| `UserMessage(content)` | Create user message |
| `AssistantMessage(content)` | Create assistant message |
| `SystemMessage(content)` | Create system message |

---

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| `connection refused` | Ollama not running | `ollama serve` |
| `model not found` | Model not pulled | `ollama pull <model>` |
| `timeout` | Model loading slow | Increase timeout, wait for warmup |
| `out of memory` | Model too large | Use smaller model variant |

---

## References

- **Ollama:** https://ollama.com
- **API Docs:** https://github.com/ollama/ollama/blob/main/docs/api.md
- **Models:** https://ollama.com/library

---

**Talk is cheap. Show me the code.**
