# OpenAI Go SDK Patterns

> "Talk is cheap. Show me the code." — Linus Torvalds

---

## TL;DR - Quick Start

```go
import "github.com/openai/openai-go/v3"

func main() {
    client := openai.NewClient()  // Uses OPENAI_API_KEY env var

    completion, err := client.Chat.Completions.New(context.Background(),
        openai.ChatCompletionNewParams{
            Model: openai.ChatModelGPT4o,
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.UserMessage("Hello!"),
            },
        })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(completion.Choices[0].Message.Content)
}
```

---

## Installation

```bash
# Requires Go 1.22+
go get -u 'github.com/openai/openai-go/v3'
```

```go
import "github.com/openai/openai-go/v3"
```

---

## Client Configuration

### ❌ BROKEN: Hardcoded API Key

```go
// 🔴 THẢM HỌA — API key in source code = SECURITY NIGHTMARE
client := openai.NewClient(
    option.WithAPIKey("sk-1234567890abcdef"),  // NEVER DO THIS
)
```

### ✅ ĐÚNG: Environment Variable

```go
// API key from OPENAI_API_KEY environment variable
client := openai.NewClient()

// Or explicit with custom env var
apiKey := os.Getenv("MY_OPENAI_KEY")
if apiKey == "" {
    log.Fatal("MY_OPENAI_KEY environment variable required")
}
client := openai.NewClient(option.WithAPIKey(apiKey))
```

### Custom Configuration

```go
client := openai.NewClient(
    option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
    option.WithBaseURL("https://custom-endpoint.com"),
    option.WithMaxRetries(3),
    option.WithRequestTimeout(30 * time.Second),
)
```

### Azure OpenAI

```go
import "github.com/openai/openai-go/v3/azure"

client := openai.NewClient(
    azure.WithEndpoint("https://your-resource.openai.azure.com", "2024-06-01"),
    azure.WithAPIKey(os.Getenv("AZURE_OPENAI_API_KEY")),
)
```

---

## Chat Completions

### Basic Request

```go
func chat(ctx context.Context, prompt string) (string, error) {
    completion, err := client.Chat.Completions.New(ctx,
        openai.ChatCompletionNewParams{
            Model: openai.ChatModelGPT4o,
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.SystemMessage("You are a helpful assistant."),
                openai.UserMessage(prompt),
            },
        })
    if err != nil {
        return "", fmt.Errorf("chat completion: %w", err)
    }

    if len(completion.Choices) == 0 {
        return "", errors.New("no choices in response")
    }

    return completion.Choices[0].Message.Content, nil
}
```

### With Parameters

```go
completion, err := client.Chat.Completions.New(ctx,
    openai.ChatCompletionNewParams{
        Model: openai.ChatModelGPT4o,
        Messages: messages,
        MaxTokens:   openai.Int(1000),
        Temperature: openai.Float(0.7),
        TopP:        openai.Float(0.9),
        N:           openai.Int(1),
        Stop:        []string{"\n\n"},
    })
```

---

## Multi-Turn Conversations

### ❌ BROKEN: Losing Context

```go
// 🔴 BROKEN — Each call loses conversation history
func chat(prompt string) string {
    completion, _ := client.Chat.Completions.New(ctx,
        openai.ChatCompletionNewParams{
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.UserMessage(prompt),  // No history!
            },
        })
    return completion.Choices[0].Message.Content
}
```

### ✅ ĐÚNG: Maintaining History

```go
type Conversation struct {
    messages []openai.ChatCompletionMessageParamUnion
    client   openai.Client
}

func NewConversation(client openai.Client, systemPrompt string) *Conversation {
    return &Conversation{
        client: client,
        messages: []openai.ChatCompletionMessageParamUnion{
            openai.SystemMessage(systemPrompt),
        },
    }
}

func (c *Conversation) Send(ctx context.Context, userMessage string) (string, error) {
    // Add user message
    c.messages = append(c.messages, openai.UserMessage(userMessage))

    // Get completion
    completion, err := c.client.Chat.Completions.New(ctx,
        openai.ChatCompletionNewParams{
            Model:    openai.ChatModelGPT4o,
            Messages: c.messages,
        })
    if err != nil {
        return "", err
    }

    // Add assistant response to history
    assistantMsg := completion.Choices[0].Message
    c.messages = append(c.messages, assistantMsg.ToParam())

    return assistantMsg.Content, nil
}
```

---

## Streaming Responses

### ❌ BROKEN: Blocking Until Complete

```go
// 🔴 BROKEN — User waits for entire response
completion, _ := client.Chat.Completions.New(ctx, params)
fmt.Println(completion.Choices[0].Message.Content)
// Long wait, then entire response at once
```

### ✅ ĐÚNG: Real-Time Streaming

```go
func streamChat(ctx context.Context, prompt string) error {
    stream := client.Chat.Completions.NewStreaming(ctx,
        openai.ChatCompletionNewParams{
            Model: openai.ChatModelGPT4o,
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.UserMessage(prompt),
            },
        })

    for stream.Next() {
        chunk := stream.Current()
        if len(chunk.Choices) > 0 {
            delta := chunk.Choices[0].Delta
            if delta.Content != "" {
                fmt.Print(delta.Content)  // Print as it arrives
            }
        }
    }
    fmt.Println()

    if err := stream.Err(); err != nil {
        return fmt.Errorf("stream error: %w", err)
    }

    return nil
}
```

### With Accumulator (Recommended)

```go
func streamWithAccumulator(ctx context.Context, prompt string) (*openai.ChatCompletion, error) {
    stream := client.Chat.Completions.NewStreaming(ctx,
        openai.ChatCompletionNewParams{
            Model: openai.ChatModelGPT4o,
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.UserMessage(prompt),
            },
        })

    acc := openai.ChatCompletionAccumulator{}

    for stream.Next() {
        chunk := stream.Current()
        acc.AddChunk(chunk)

        // Check for completed content blocks
        if content, ok := acc.JustFinishedContent(); ok {
            fmt.Printf("Content complete: %s\n", content)
        }

        // Check for completed tool calls
        if toolCall, ok := acc.JustFinishedToolCall(); ok {
            fmt.Printf("Tool call complete: %s(%s)\n",
                toolCall.Function.Name,
                toolCall.Function.Arguments)
        }
    }

    if err := stream.Err(); err != nil {
        return nil, err
    }

    // Get final accumulated response
    return acc.ChatCompletion, nil
}
```

---

## Tool/Function Calling

### Define Tools

```go
type WeatherParams struct {
    Location string `json:"location" jsonschema:"description=City name"`
    Unit     string `json:"unit,omitempty" jsonschema:"enum=celsius,enum=fahrenheit"`
}

func getWeatherTool() openai.ChatCompletionToolParam {
    return openai.ChatCompletionToolParam{
        Type: openai.ChatCompletionToolTypeFunction,
        Function: openai.FunctionDefinitionParam{
            Name:        "get_weather",
            Description: openai.String("Get current weather for a location"),
            Parameters: openai.FunctionParameters{
                "type": "object",
                "properties": map[string]any{
                    "location": map[string]any{
                        "type":        "string",
                        "description": "City name, e.g., San Francisco",
                    },
                    "unit": map[string]any{
                        "type": "string",
                        "enum": []string{"celsius", "fahrenheit"},
                    },
                },
                "required": []string{"location"},
            },
        },
    }
}
```

### Complete Tool Calling Flow

```go
func chatWithTools(ctx context.Context, userMessage string) (string, error) {
    messages := []openai.ChatCompletionMessageParamUnion{
        openai.UserMessage(userMessage),
    }

    // First call with tools
    completion, err := client.Chat.Completions.New(ctx,
        openai.ChatCompletionNewParams{
            Model:    openai.ChatModelGPT4o,
            Messages: messages,
            Tools:    []openai.ChatCompletionToolParam{getWeatherTool()},
        })
    if err != nil {
        return "", err
    }

    choice := completion.Choices[0]

    // Check if model wants to use tools
    if choice.FinishReason == openai.ChatCompletionChoicesFinishReasonToolCalls {
        // Add assistant's response to history
        messages = append(messages, choice.Message.ToParam())

        // Process each tool call
        for _, toolCall := range choice.Message.ToolCalls {
            result := executeToolCall(toolCall)

            // Add tool result to messages
            messages = append(messages,
                openai.ToolMessage(toolCall.ID, result))
        }

        // Get final response with tool results
        completion, err = client.Chat.Completions.New(ctx,
            openai.ChatCompletionNewParams{
                Model:    openai.ChatModelGPT4o,
                Messages: messages,
            })
        if err != nil {
            return "", err
        }
    }

    return completion.Choices[0].Message.Content, nil
}

func executeToolCall(toolCall openai.ChatCompletionMessageToolCall) string {
    switch toolCall.Function.Name {
    case "get_weather":
        var params WeatherParams
        json.Unmarshal([]byte(toolCall.Function.Arguments), &params)
        // Call actual weather API
        return fmt.Sprintf(`{"temperature": 22, "unit": "%s", "location": "%s"}`,
            params.Unit, params.Location)
    default:
        return `{"error": "unknown function"}`
    }
}
```

---

## Structured Outputs

### Generate JSON Schema from Go Struct

```go
import "github.com/invopop/jsonschema"

type Person struct {
    Name  string   `json:"name" jsonschema:"description=Person's full name"`
    Age   int      `json:"age" jsonschema:"minimum=0,maximum=150"`
    Email string   `json:"email" jsonschema:"format=email"`
    Tags  []string `json:"tags,omitempty"`
}

func getPersonSchema() openai.ResponseFormatJSONSchemaJSONSchemaParam {
    reflector := jsonschema.Reflector{
        AllowAdditionalProperties: false,
        DoNotReference:            true,
    }
    schema := reflector.Reflect(&Person{})

    return openai.ResponseFormatJSONSchemaJSONSchemaParam{
        Name:        "person",
        Description: openai.String("A person's information"),
        Schema:      schema,
        Strict:      openai.Bool(true),
    }
}
```

### Request with Structured Output

```go
func extractPerson(ctx context.Context, text string) (*Person, error) {
    completion, err := client.Chat.Completions.New(ctx,
        openai.ChatCompletionNewParams{
            Model: openai.ChatModelGPT4o,
            Messages: []openai.ChatCompletionMessageParamUnion{
                openai.SystemMessage("Extract person information from text as JSON."),
                openai.UserMessage(text),
            },
            ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
                OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
                    Type:       openai.ResponseFormatJSONSchemaTypeJSONSchema,
                    JSONSchema: getPersonSchema(),
                },
            },
        })
    if err != nil {
        return nil, err
    }

    var person Person
    if err := json.Unmarshal(
        []byte(completion.Choices[0].Message.Content),
        &person); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    return &person, nil
}
```

---

## Error Handling

### ❌ BROKEN: Ignoring Errors

```go
// 🔴 BROKEN — Silent failures
completion, _ := client.Chat.Completions.New(ctx, params)
fmt.Println(completion.Choices[0].Message.Content)  // PANIC if error
```

### ✅ ĐÚNG: Proper Error Handling

```go
func chat(ctx context.Context, prompt string) (string, error) {
    completion, err := client.Chat.Completions.New(ctx, params)
    if err != nil {
        var apierr *openai.Error
        if errors.As(err, &apierr) {
            // API-specific error
            log.Printf("API Error: %s (status %d)",
                apierr.Message, apierr.StatusCode)
            log.Printf("Request: %s", apierr.DumpRequest(true))

            // Handle specific error codes
            switch apierr.StatusCode {
            case 429:
                return "", fmt.Errorf("rate limited: %w", err)
            case 401:
                return "", fmt.Errorf("invalid API key: %w", err)
            case 500, 502, 503:
                return "", fmt.Errorf("OpenAI service error: %w", err)
            }
        }
        return "", fmt.Errorf("chat completion failed: %w", err)
    }

    if len(completion.Choices) == 0 {
        return "", errors.New("no choices in response")
    }

    // Check for content filter
    if completion.Choices[0].FinishReason == openai.ChatCompletionChoicesFinishReasonContentFilter {
        return "", errors.New("content filtered by safety system")
    }

    return completion.Choices[0].Message.Content, nil
}
```

---

## Retries & Timeouts

### Configure Retries

```go
client := openai.NewClient(
    option.WithMaxRetries(3),  // Default is 2
)

// Automatic retries for:
// - Connection errors
// - 408 Request Timeout
// - 409 Conflict
// - 429 Rate Limited
// - 5xx Server Errors
```

### Context Timeout

```go
func chatWithTimeout(prompt string) (string, error) {
    // Total timeout for entire operation
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    completion, err := client.Chat.Completions.New(ctx, params)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return "", fmt.Errorf("request timed out: %w", err)
        }
        return "", err
    }

    return completion.Choices[0].Message.Content, nil
}
```

### Per-Retry Timeout

```go
// Timeout per retry attempt (not total)
completion, err := client.Chat.Completions.New(ctx, params,
    option.WithRequestTimeout(30*time.Second),
)
```

---

## Embeddings

```go
func getEmbedding(ctx context.Context, text string) ([]float64, error) {
    response, err := client.Embeddings.New(ctx,
        openai.EmbeddingNewParams{
            Model: openai.EmbeddingModelTextEmbedding3Small,
            Input: openai.EmbeddingNewParamsInputUnion{
                OfString: openai.String(text),
            },
        })
    if err != nil {
        return nil, fmt.Errorf("create embedding: %w", err)
    }

    if len(response.Data) == 0 {
        return nil, errors.New("no embedding returned")
    }

    return response.Data[0].Embedding, nil
}

// Batch embeddings
func getBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
    response, err := client.Embeddings.New(ctx,
        openai.EmbeddingNewParams{
            Model: openai.EmbeddingModelTextEmbedding3Small,
            Input: openai.EmbeddingNewParamsInputUnion{
                OfArrayOfStrings: texts,
            },
        })
    if err != nil {
        return nil, err
    }

    embeddings := make([][]float64, len(response.Data))
    for i, data := range response.Data {
        embeddings[i] = data.Embedding
    }
    return embeddings, nil
}
```

---

## Audio: Speech & Transcription

### Text-to-Speech

```go
func textToSpeech(ctx context.Context, text string) ([]byte, error) {
    response, err := client.Audio.Speech.New(ctx,
        openai.AudioSpeechNewParams{
            Model:          openai.SpeechModelTTS1,
            Voice:          openai.AudioSpeechNewParamsVoiceAlloy,
            Input:          text,
            ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
        })
    if err != nil {
        return nil, err
    }
    defer response.Body.Close()

    return io.ReadAll(response.Body)
}
```

### Transcription

```go
func transcribe(ctx context.Context, audioFile io.Reader) (string, error) {
    response, err := client.Audio.Transcriptions.New(ctx,
        openai.AudioTranscriptionNewParams{
            Model: openai.AudioModelWhisper1,
            File:  openai.File(audioFile, "audio.mp3", "audio/mpeg"),
        })
    if err != nil {
        return "", err
    }

    return response.Text, nil
}
```

---

## Image Generation

```go
func generateImage(ctx context.Context, prompt string) (string, error) {
    response, err := client.Images.Generate(ctx,
        openai.ImageGenerateParams{
            Model:  openai.ImageModelDallE3,
            Prompt: prompt,
            N:      openai.Int(1),
            Size:   openai.ImageGenerateParamsSize1024x1024,
        })
    if err != nil {
        return "", err
    }

    if len(response.Data) == 0 {
        return "", errors.New("no image generated")
    }

    return response.Data[0].URL, nil
}
```

---

## Best Practices Checklist

### Security
- [ ] API key from environment variable, NEVER hardcoded
- [ ] Validate all user input before sending to API
- [ ] Implement rate limiting in your application
- [ ] Log errors but NEVER log full API keys

### Performance
- [ ] Use streaming for long responses
- [ ] Batch embeddings requests when possible
- [ ] Set appropriate timeouts
- [ ] Cache responses when applicable

### Reliability
- [ ] Handle all error cases explicitly
- [ ] Check for empty responses
- [ ] Configure retries appropriately
- [ ] Use context for cancellation

### Code Quality
- [ ] Maintain conversation history properly
- [ ] Use accumulator for streaming
- [ ] Define tools with proper schemas
- [ ] Parse structured outputs with error handling

---

## Quick Reference

| Operation | Method |
|-----------|--------|
| Chat | `client.Chat.Completions.New()` |
| Streaming | `client.Chat.Completions.NewStreaming()` |
| Embeddings | `client.Embeddings.New()` |
| Images | `client.Images.Generate()` |
| TTS | `client.Audio.Speech.New()` |
| Transcription | `client.Audio.Transcriptions.New()` |

| Model Constant | Value |
|----------------|-------|
| `ChatModelGPT4o` | gpt-4o |
| `ChatModelGPT4oMini` | gpt-4o-mini |
| `ChatModelGPT4Turbo` | gpt-4-turbo |
| `EmbeddingModelTextEmbedding3Small` | text-embedding-3-small |
| `EmbeddingModelTextEmbedding3Large` | text-embedding-3-large |
| `ImageModelDallE3` | dall-e-3 |
| `SpeechModelTTS1` | tts-1 |
| `AudioModelWhisper1` | whisper-1 |

---

## References

- **Repository:** https://github.com/openai/openai-go
- **Examples:** https://github.com/openai/openai-go/tree/main/examples
- **API Docs:** https://platform.openai.com/docs/api-reference

---

**Talk is cheap. Show me the code.**
