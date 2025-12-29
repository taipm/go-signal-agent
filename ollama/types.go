package ollama

// =============================================================================
// GENERATE API TYPES
// =============================================================================

// GenerateRequest is the request body for /api/generate.
type GenerateRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	System  string        `json:"system,omitempty"`
	Stream  bool          `json:"stream"`
	Options *ModelOptions `json:"options,omitempty"`
	Format  string        `json:"format,omitempty"` // "json" for JSON output
}

// GenerateResponse is the response from /api/generate.
type GenerateResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// =============================================================================
// CHAT API TYPES
// =============================================================================

// Role represents the role in a chat message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single message in a chat conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request body for /api/chat.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  *ModelOptions `json:"options,omitempty"`
	Format   string        `json:"format,omitempty"`
}

// ChatResponse is the response from /api/chat.
type ChatResponse struct {
	Model              string  `json:"model"`
	Message            Message `json:"message"`
	Done               bool    `json:"done"`
	TotalDuration      int64   `json:"total_duration,omitempty"`
	LoadDuration       int64   `json:"load_duration,omitempty"`
	PromptEvalCount    int     `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64   `json:"prompt_eval_duration,omitempty"`
	EvalCount          int     `json:"eval_count,omitempty"`
	EvalDuration       int64   `json:"eval_duration,omitempty"`
}

// =============================================================================
// MODEL OPTIONS
// =============================================================================

// ModelOptions contains parameters that control model behavior.
type ModelOptions struct {
	// Temperature controls randomness (0.0 = deterministic, 1.0 = creative)
	Temperature float64 `json:"temperature,omitempty"`

	// TopP is nucleus sampling parameter
	TopP float64 `json:"top_p,omitempty"`

	// TopK limits vocabulary for sampling
	TopK int `json:"top_k,omitempty"`

	// NumPredict is max tokens to generate
	NumPredict int `json:"num_predict,omitempty"`

	// Stop sequences that will stop generation
	Stop []string `json:"stop,omitempty"`

	// RepeatPenalty penalizes repetition
	RepeatPenalty float64 `json:"repeat_penalty,omitempty"`

	// Seed for reproducibility (-1 = random)
	Seed int `json:"seed,omitempty"`
}

// =============================================================================
// TAGS API (LIST MODELS)
// =============================================================================

// TagsResponse is the response from /api/tags.
type TagsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo contains information about an available model.
type ModelInfo struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at"`
	Size       int64  `json:"size"`
	Digest     string `json:"digest"`
}

// =============================================================================
// HELPER CONSTRUCTORS
// =============================================================================

// UserMessage creates a user message.
func UserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// AssistantMessage creates an assistant message.
func AssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// SystemMessage creates a system message.
func SystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}
