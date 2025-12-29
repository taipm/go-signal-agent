// Package ollama provides a streaming client for the Ollama API.
//
// Ollama is a local LLM runtime that supports various models including
// qwen3, llama, mistral, and more.
//
// This client uses STREAMING by default for fast response display.
//
// API Documentation: https://github.com/ollama/ollama/blob/main/docs/api.md
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// CONFIGURATION
// =============================================================================

const (
	DefaultEndpoint  = "http://localhost:11434"
	DefaultModel     = "qwen3:1.7b"
	DefaultTimeout   = 120 * time.Second // Longer timeout for streaming
	MaxErrorBodySize = 4 * 1024          // 4KB max for error response body
)

// ClientConfig holds configuration for the Ollama client.
type ClientConfig struct {
	Endpoint string        // Ollama API endpoint (default: http://localhost:11434)
	Model    string        // Model to use (default: qwen3:1.7b)
	Timeout  time.Duration // Request timeout (default: 120s)
}

// DefaultConfig returns default client configuration.
func DefaultConfig() ClientConfig {
	return ClientConfig{
		Endpoint: DefaultEndpoint,
		Model:    DefaultModel,
		Timeout:  DefaultTimeout,
	}
}

// =============================================================================
// CLIENT
// =============================================================================

// Client is a streaming HTTP client for the Ollama API.
// Thread-safe: all methods can be called concurrently from multiple goroutines.
type Client struct {
	mu         sync.RWMutex // Protects config.Model for concurrent access
	config     ClientConfig
	httpClient *http.Client
}

// NewClient creates a new Ollama client with the given configuration.
func NewClient(config ClientConfig) *Client {
	if config.Endpoint == "" {
		config.Endpoint = DefaultEndpoint
	}
	if config.Model == "" {
		config.Model = DefaultModel
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NewDefaultClient creates a client with default configuration.
func NewDefaultClient() *Client {
	return NewClient(DefaultConfig())
}

// =============================================================================
// STREAMING API - PRIMARY METHODS
// =============================================================================

// StreamCallback is called for each chunk of streamed text.
// Return an error to stop streaming early.
type StreamCallback func(chunk string) error

// GenerateStream sends a prompt and streams the response.
// Each token is passed to the callback as it arrives.
// Thread-safe: uses getModel() for concurrent access to model name.
func (c *Client) GenerateStream(ctx context.Context, prompt string, callback StreamCallback) error {
	req := GenerateRequest{
		Model:  c.getModel(),
		Prompt: prompt,
		Stream: true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	return doStreamRequest(c, ctx, "/api/generate", body, callback,
		func(chunk *GenerateResponse) (string, bool) {
			return chunk.Response, chunk.Done
		})
}

// ChatStream sends a conversation and streams the response.
// Thread-safe: uses getModel() for concurrent access to model name.
func (c *Client) ChatStream(ctx context.Context, messages []Message, callback StreamCallback) error {
	req := ChatRequest{
		Model:    c.getModel(),
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	return doStreamRequest(c, ctx, "/api/chat", body, callback,
		func(chunk *ChatResponse) (string, bool) {
			return chunk.Message.Content, chunk.Done
		})
}

// =============================================================================
// INTERNAL HELPERS
// =============================================================================

// chunkExtractor extracts content and done flag from a response chunk.
type chunkExtractor[T any] func(*T) (content string, done bool)

// doStreamRequest performs a streaming HTTP request and processes chunks.
// This is the core streaming logic shared by GenerateStream and ChatStream.
// Note: Go methods cannot have type parameters, so this is a package-level function.
func doStreamRequest[T any](
	c *Client,
	ctx context.Context,
	endpoint string,
	body []byte,
	callback StreamCallback,
	extract chunkExtractor[T],
) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.config.Endpoint+endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readErrorBody(resp)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk T
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue // Skip malformed lines
		}

		content, done := extract(&chunk)
		if content != "" {
			if err := callback(content); err != nil {
				return err // Callback requested stop
			}
		}

		if done {
			break
		}
	}

	return scanner.Err()
}

// readErrorBody reads and formats an error response from the API.
// Limits read size to prevent memory exhaustion from large error responses.
func readErrorBody(resp *http.Response) error {
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
	return fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(bodyBytes))
}

// =============================================================================
// CONVENIENCE METHODS (collect full response)
// =============================================================================

// Generate sends a prompt and collects the full response.
// Uses streaming internally but returns the complete text.
// Uses strings.Builder for O(n) performance instead of O(n²) concatenation.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	var builder strings.Builder
	err := c.GenerateStream(ctx, prompt, func(chunk string) error {
		builder.WriteString(chunk)
		return nil
	})
	return builder.String(), err
}

// Chat sends a conversation and collects the full response.
// Uses strings.Builder for O(n) performance instead of O(n²) concatenation.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	var builder strings.Builder
	err := c.ChatStream(ctx, messages, func(chunk string) error {
		builder.WriteString(chunk)
		return nil
	})
	return builder.String(), err
}

// =============================================================================
// UTILITY METHODS
// =============================================================================

// IsAvailable checks if Ollama is running and accessible.
func (c *Client) IsAvailable(ctx context.Context) bool {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.Endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListModels returns the list of available models.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d)", resp.StatusCode)
	}

	var result TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Models, nil
}

// Model returns the currently configured model name.
// Thread-safe: can be called concurrently with SetModel.
func (c *Client) Model() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Model
}

// SetModel changes the model for subsequent requests.
// Thread-safe: can be called concurrently with other methods.
func (c *Client) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.Model = model
}

// getModel returns the current model with read lock (internal use).
func (c *Client) getModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.Model
}
