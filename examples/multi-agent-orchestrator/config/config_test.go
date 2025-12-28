package config

import (
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func createTestYAML(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func setEnv(t *testing.T, key, value string) func() {
	t.Helper()
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() {
		if old == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, old)
		}
	}
}

// =============================================================================
// LOAD CONFIG TESTS
// =============================================================================

func TestLoadConfig_ValidYAML(t *testing.T) {
	content := `
ollama:
  host: http://localhost:11434
  timeout: 120

memory:
  storage_dir: /tmp/memory
  default_max_size: 100
  default_ttl: 1h

coordinator:
  id: coordinator
  model: qwen3:1.7b
  max_workers: 3
  available_workers: [writing, translation]
  system_prompt: "You are a router"

workers:
  writing:
    id: writing
    model: qwen3:1.7b
    system_prompt: "Write content"

output:
  id: output
  merge_strategy: llm
`
	path := createTestYAML(t, content)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// Check Ollama config
	if cfg.Ollama.Host != "http://localhost:11434" {
		t.Errorf("Ollama.Host = %q, want %q", cfg.Ollama.Host, "http://localhost:11434")
	}
	if cfg.Ollama.Timeout != 120 {
		t.Errorf("Ollama.Timeout = %d, want %d", cfg.Ollama.Timeout, 120)
	}

	// Check Memory config
	if cfg.Memory.StorageDir != "/tmp/memory" {
		t.Errorf("Memory.StorageDir = %q, want %q", cfg.Memory.StorageDir, "/tmp/memory")
	}
	if cfg.Memory.DefaultMaxSize != 100 {
		t.Errorf("Memory.DefaultMaxSize = %d, want %d", cfg.Memory.DefaultMaxSize, 100)
	}

	// Check Coordinator config
	if cfg.Coordinator.ID != "coordinator" {
		t.Errorf("Coordinator.ID = %q, want %q", cfg.Coordinator.ID, "coordinator")
	}
	if len(cfg.Coordinator.AvailableWorkers) != 2 {
		t.Errorf("AvailableWorkers len = %d, want 2", len(cfg.Coordinator.AvailableWorkers))
	}

	// Check Workers config
	if _, ok := cfg.Workers["writing"]; !ok {
		t.Error("Workers should contain 'writing'")
	}

	// Check Output config
	if cfg.Output.MergeStrategy != "llm" {
		t.Errorf("Output.MergeStrategy = %q, want %q", cfg.Output.MergeStrategy, "llm")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := `
invalid yaml content
  - this is not valid: yaml: syntax
`
	path := createTestYAML(t, content)

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("LoadConfig() should error on invalid YAML")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadConfig() should error on nonexistent file")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Minimal config with no values set
	content := `
workers:
  test:
    id: test
`
	path := createTestYAML(t, content)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// Check defaults applied
	if cfg.Ollama.Host != FallbackOllamaHost {
		t.Errorf("Ollama.Host = %q, want %q", cfg.Ollama.Host, FallbackOllamaHost)
	}
	if cfg.Ollama.Timeout != FallbackOllamaTimeout {
		t.Errorf("Ollama.Timeout = %d, want %d", cfg.Ollama.Timeout, FallbackOllamaTimeout)
	}
	if cfg.Memory.DefaultMaxSize != FallbackMemoryMaxSize {
		t.Errorf("Memory.DefaultMaxSize = %d, want %d", cfg.Memory.DefaultMaxSize, FallbackMemoryMaxSize)
	}
	if cfg.Memory.DefaultTTL != FallbackMemoryTTL {
		t.Errorf("Memory.DefaultTTL = %q, want %q", cfg.Memory.DefaultTTL, FallbackMemoryTTL)
	}
	if cfg.Coordinator.ID != "coordinator" {
		t.Errorf("Coordinator.ID = %q, want %q", cfg.Coordinator.ID, "coordinator")
	}
	if cfg.Coordinator.MaxWorkers != 3 {
		t.Errorf("Coordinator.MaxWorkers = %d, want %d", cfg.Coordinator.MaxWorkers, 3)
	}
	if cfg.Output.ID != "output" {
		t.Errorf("Output.ID = %q, want %q", cfg.Output.ID, "output")
	}
	if cfg.Output.MergeStrategy != "llm" {
		t.Errorf("Output.MergeStrategy = %q, want %q", cfg.Output.MergeStrategy, "llm")
	}
	if cfg.Output.ResponseTimeout != "60s" {
		t.Errorf("Output.ResponseTimeout = %q, want %q", cfg.Output.ResponseTimeout, "60s")
	}
}

func TestLoadConfig_EnvOverride_OllamaHost(t *testing.T) {
	content := `
ollama:
  host: http://yaml-host:11434
`
	path := createTestYAML(t, content)

	// Set env var
	cleanup := setEnv(t, EnvOllamaHost, "http://env-host:11434")
	defer cleanup()

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// Env should override YAML
	if cfg.Ollama.Host != "http://env-host:11434" {
		t.Errorf("Ollama.Host = %q, want %q (env override)", cfg.Ollama.Host, "http://env-host:11434")
	}
}

func TestLoadConfig_EnvOverride_OllamaTimeout(t *testing.T) {
	content := `
ollama:
  timeout: 60
`
	path := createTestYAML(t, content)

	cleanup := setEnv(t, EnvOllamaTimeout, "180")
	defer cleanup()

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Ollama.Timeout != 180 {
		t.Errorf("Ollama.Timeout = %d, want %d (env override)", cfg.Ollama.Timeout, 180)
	}
}

func TestLoadConfig_EnvOverride_MemoryStorageDir(t *testing.T) {
	content := `
memory:
  storage_dir: /yaml/storage
`
	path := createTestYAML(t, content)

	cleanup := setEnv(t, EnvMemoryStorageDir, "/env/storage")
	defer cleanup()

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Memory.StorageDir != "/env/storage" {
		t.Errorf("Memory.StorageDir = %q, want %q (env override)", cfg.Memory.StorageDir, "/env/storage")
	}
}

func TestLoadConfig_EnvFallback_NoYAMLValue(t *testing.T) {
	content := `
workers:
  test:
    id: test
`
	path := createTestYAML(t, content)

	cleanup := setEnv(t, EnvOllamaHost, "http://custom:11434")
	defer cleanup()

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// When YAML is empty, env should be used
	if cfg.Ollama.Host != "http://custom:11434" {
		t.Errorf("Ollama.Host = %q, want %q (env when yaml empty)", cfg.Ollama.Host, "http://custom:11434")
	}
}

// =============================================================================
// GETENV TESTS
// =============================================================================

func TestGetEnv(t *testing.T) {
	cleanup := setEnv(t, "TEST_VAR", "test_value")
	defer cleanup()

	if got := getEnv("TEST_VAR", "fallback"); got != "test_value" {
		t.Errorf("getEnv() = %q, want %q", got, "test_value")
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_VAR")

	if got := getEnv("NONEXISTENT_VAR", "fallback"); got != "fallback" {
		t.Errorf("getEnv() = %q, want %q", got, "fallback")
	}
}

func TestGetEnvInt(t *testing.T) {
	cleanup := setEnv(t, "TEST_INT", "42")
	defer cleanup()

	if got := getEnvInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvInt() = %d, want %d", got, 42)
	}
}

func TestGetEnvInt_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_INT")

	if got := getEnvInt("NONEXISTENT_INT", 99); got != 99 {
		t.Errorf("getEnvInt() = %d, want %d", got, 99)
	}
}

func TestGetEnvInt_InvalidNumber(t *testing.T) {
	cleanup := setEnv(t, "INVALID_INT", "not_a_number")
	defer cleanup()

	if got := getEnvInt("INVALID_INT", 99); got != 99 {
		t.Errorf("getEnvInt() = %d, want %d (fallback for invalid)", got, 99)
	}
}

// =============================================================================
// RESOLVE WORKER MEMORY TESTS
// =============================================================================

func TestConfig_ResolveWorkerMemory(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			DefaultMaxSize: 100,
			DefaultTTL:     "1h",
			Types: map[string]TypeConfig{
				"glossary": {MaxSize: 500, TTL: "720h"},
			},
		},
		Workers: map[string]WorkerConfig{
			"translation": {
				ID: "translation",
				Memory: WorkerMemory{
					Types: []string{"conversation", "glossary"},
				},
			},
		},
	}

	resolved := cfg.ResolveWorkerMemory("translation")

	if !resolved.Enabled {
		t.Error("ResolveWorkerMemory should be enabled")
	}
	if len(resolved.Types) != 2 {
		t.Fatalf("ResolveWorkerMemory types = %d, want 2", len(resolved.Types))
	}

	// conversation should use defaults
	conv := resolved.Types[0]
	if conv.Name != "conversation" {
		t.Errorf("Types[0].Name = %q, want %q", conv.Name, "conversation")
	}
	if conv.MaxSize != 100 {
		t.Errorf("conversation MaxSize = %d, want 100 (default)", conv.MaxSize)
	}
	if conv.TTL != "1h" {
		t.Errorf("conversation TTL = %q, want 1h (default)", conv.TTL)
	}

	// glossary should use type-specific config
	glossary := resolved.Types[1]
	if glossary.Name != "glossary" {
		t.Errorf("Types[1].Name = %q, want %q", glossary.Name, "glossary")
	}
	if glossary.MaxSize != 500 {
		t.Errorf("glossary MaxSize = %d, want 500 (type override)", glossary.MaxSize)
	}
	if glossary.TTL != "720h" {
		t.Errorf("glossary TTL = %q, want 720h (type override)", glossary.TTL)
	}
}

func TestConfig_ResolveWorkerMemory_WorkerOverride(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			DefaultMaxSize: 100,
			DefaultTTL:     "1h",
		},
		Workers: map[string]WorkerConfig{
			"writing": {
				ID: "writing",
				Memory: WorkerMemory{
					Types:   []string{"conversation"},
					MaxSize: 200,
					TTL:     "2h",
				},
			},
		},
	}

	resolved := cfg.ResolveWorkerMemory("writing")

	if !resolved.Enabled {
		t.Fatal("ResolveWorkerMemory should be enabled")
	}

	conv := resolved.Types[0]
	if conv.MaxSize != 200 {
		t.Errorf("MaxSize = %d, want 200 (worker override)", conv.MaxSize)
	}
	if conv.TTL != "2h" {
		t.Errorf("TTL = %q, want 2h (worker override)", conv.TTL)
	}
}

func TestConfig_ResolveWorkerMemory_NonexistentWorker(t *testing.T) {
	cfg := &Config{
		Workers: map[string]WorkerConfig{},
	}

	resolved := cfg.ResolveWorkerMemory("nonexistent")

	if resolved.Enabled {
		t.Error("ResolveWorkerMemory should be disabled for nonexistent worker")
	}
}

func TestConfig_ResolveWorkerMemory_NoMemoryTypes(t *testing.T) {
	cfg := &Config{
		Workers: map[string]WorkerConfig{
			"simple": {
				ID: "simple",
				Memory: WorkerMemory{
					Types: []string{}, // Empty types
				},
			},
		},
	}

	resolved := cfg.ResolveWorkerMemory("simple")

	if resolved.Enabled {
		t.Error("ResolveWorkerMemory should be disabled for empty memory types")
	}
}

// =============================================================================
// GET WORKER IDS TESTS
// =============================================================================

func TestConfig_GetWorkerIDs(t *testing.T) {
	cfg := &Config{
		Workers: map[string]WorkerConfig{
			"writing":     {ID: "writing"},
			"translation": {ID: "translation"},
			"summary":     {ID: "summary"},
		},
	}

	ids := cfg.GetWorkerIDs()

	if len(ids) != 3 {
		t.Errorf("GetWorkerIDs() len = %d, want 3", len(ids))
	}

	// Check all workers are present (order not guaranteed with maps)
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expected := range []string{"writing", "translation", "summary"} {
		if !idMap[expected] {
			t.Errorf("GetWorkerIDs() missing %q", expected)
		}
	}
}

func TestConfig_GetWorkerIDs_Empty(t *testing.T) {
	cfg := &Config{
		Workers: map[string]WorkerConfig{},
	}

	ids := cfg.GetWorkerIDs()

	if len(ids) != 0 {
		t.Errorf("GetWorkerIDs() len = %d, want 0", len(ids))
	}
}

// =============================================================================
// CONFIG STRUCTS TESTS
// =============================================================================

func TestMemoryTypeConfig_Alias(t *testing.T) {
	// Verify MemoryTypeConfig is same as TypeConfig (alias test)
	cfg := TypeConfig{
		MaxSize: 100,
		TTL:     "1h",
	}

	if cfg.MaxSize != 100 || cfg.TTL != "1h" {
		t.Error("TypeConfig fields not as expected")
	}
}

func TestWorkerMemory_Empty(t *testing.T) {
	wm := WorkerMemory{}

	if len(wm.Types) != 0 {
		t.Error("Empty WorkerMemory should have no types")
	}
	if wm.MaxSize != 0 {
		t.Error("Empty WorkerMemory MaxSize should be 0")
	}
	if wm.TTL != "" {
		t.Error("Empty WorkerMemory TTL should be empty")
	}
}

// =============================================================================
// FULL CONFIG INTEGRATION TEST
// =============================================================================

func TestLoadConfig_FullIntegration(t *testing.T) {
	content := `
ollama:
  host: http://localhost:11434
  timeout: 180

memory:
  storage_dir: ./storage
  default_max_size: 50
  default_ttl: "30m"
  types:
    conversation:
      max_size: 100
      ttl: "1h"
    glossary:
      max_size: 500
      ttl: "720h"

coordinator:
  id: coordinator
  model: qwen3:1.7b
  temperature: 0.3
  max_workers: 2
  available_workers:
    - writing
    - translation
    - summary
  system_prompt: "You are a router"

workers:
  writing:
    id: writing
    description: "Content writer"
    model: qwen3:1.7b
    temperature: 0.7
    capabilities:
      - content_creation
      - email_writing
    memory:
      types:
        - conversation
        - drafts
    system_prompt: "Write content"

  translation:
    id: translation
    model: qwen3:1.7b
    memory:
      types:
        - glossary
    system_prompt: "Translate"

output:
  id: output
  model: qwen3:1.7b
  merge_strategy: template
  response_timeout: "30s"
  stream_output: true
  system_prompt: "Merge results"
`
	path := createTestYAML(t, content)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	// Verify all sections loaded
	if cfg.Ollama.Timeout != 180 {
		t.Errorf("Ollama.Timeout = %d, want 180", cfg.Ollama.Timeout)
	}

	if len(cfg.Memory.Types) != 2 {
		t.Errorf("Memory.Types len = %d, want 2", len(cfg.Memory.Types))
	}

	if len(cfg.Coordinator.AvailableWorkers) != 3 {
		t.Errorf("AvailableWorkers len = %d, want 3", len(cfg.Coordinator.AvailableWorkers))
	}

	if len(cfg.Workers) != 2 {
		t.Errorf("Workers len = %d, want 2", len(cfg.Workers))
	}

	writing := cfg.Workers["writing"]
	if writing.Temperature != 0.7 {
		t.Errorf("writing.Temperature = %v, want 0.7", writing.Temperature)
	}
	if len(writing.Capabilities) != 2 {
		t.Errorf("writing.Capabilities len = %d, want 2", len(writing.Capabilities))
	}
	if len(writing.Memory.Types) != 2 {
		t.Errorf("writing.Memory.Types len = %d, want 2", len(writing.Memory.Types))
	}

	if cfg.Output.StreamOutput != true {
		t.Error("Output.StreamOutput should be true")
	}
	if cfg.Output.ResponseTimeout != "30s" {
		t.Errorf("Output.ResponseTimeout = %q, want 30s", cfg.Output.ResponseTimeout)
	}
}
