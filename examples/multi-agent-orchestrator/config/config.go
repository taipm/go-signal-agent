package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Environment variable names
const (
	EnvOllamaHost           = "OLLAMA_HOST"
	EnvOllamaTimeout        = "OLLAMA_TIMEOUT"
	EnvMemoryStorageDir     = "MEMORY_STORAGE_DIR"
	EnvMemoryDefaultMaxSize = "MEMORY_DEFAULT_MAX_SIZE"
	EnvMemoryDefaultTTL     = "MEMORY_DEFAULT_TTL"
)

// Fallback defaults (used when env vars are not set)
const (
	FallbackOllamaHost    = "http://localhost:11434"
	FallbackOllamaTimeout = 120
	FallbackMemoryMaxSize = 100
	FallbackMemoryTTL     = "24h"
)

func init() {
	// Load .env file if it exists (silently ignore if not found)
	_ = godotenv.Load()
}

// getEnv returns the value of an environment variable or a fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt returns the value of an environment variable as int or a fallback
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

// Config represents the full configuration
type Config struct {
	Ollama      OllamaConfig           `yaml:"ollama"`
	Memory      MemoryConfig           `yaml:"memory"`
	Coordinator CoordinatorConfig      `yaml:"coordinator"`
	Workers     map[string]WorkerConfig `yaml:"workers"`
	Output      OutputConfig           `yaml:"output"`
}

// OllamaConfig holds Ollama connection settings
type OllamaConfig struct {
	Host    string `yaml:"host"`
	Timeout int    `yaml:"timeout"`
}

// MemoryConfig holds global memory settings
type MemoryConfig struct {
	StorageDir     string                `yaml:"storage_dir"`
	DefaultMaxSize int                   `yaml:"default_max_size"`
	DefaultTTL     string                `yaml:"default_ttl"`
	Types          map[string]TypeConfig `yaml:"types"`
}

// TypeConfig holds settings for a memory type
type TypeConfig struct {
	MaxSize int    `yaml:"max_size"`
	TTL     string `yaml:"ttl"`
}

// CoordinatorConfig holds coordinator agent settings
type CoordinatorConfig struct {
	ID               string   `yaml:"id"`
	Model            string   `yaml:"model"`
	Temperature      float64  `yaml:"temperature"`
	MaxWorkers       int      `yaml:"max_workers"`
	AvailableWorkers []string `yaml:"available_workers"`
	SystemPrompt     string   `yaml:"system_prompt"`
}

// WorkerConfig holds per-worker configuration
type WorkerConfig struct {
	ID           string       `yaml:"id"`
	Description  string       `yaml:"description"`
	Model        string       `yaml:"model"`
	SystemPrompt string       `yaml:"system_prompt"`
	Temperature  float64      `yaml:"temperature"`
	Capabilities []string     `yaml:"capabilities"`
	Memory       WorkerMemory `yaml:"memory"`
}

// WorkerMemory configures which memory types a worker uses
type WorkerMemory struct {
	Types   []string `yaml:"types"`
	MaxSize int      `yaml:"max_size,omitempty"`
	TTL     string   `yaml:"ttl,omitempty"`
}

// OutputConfig holds output agent settings
type OutputConfig struct {
	ID              string `yaml:"id"`
	Model           string `yaml:"model"`
	Temperature     float64 `yaml:"temperature"`
	MergeStrategy   string `yaml:"merge_strategy"`
	ResponseTimeout string `yaml:"response_timeout"`
	StreamOutput    bool   `yaml:"stream_output"`
	SystemPrompt    string `yaml:"system_prompt"`
}

// ResolvedWorkerMemory holds fully resolved memory config for a worker
type ResolvedWorkerMemory struct {
	Types   []ResolvedMemoryType
	Enabled bool
}

// ResolvedMemoryType holds resolved settings for a memory type
type ResolvedMemoryType struct {
	Name    string
	MaxSize int
	TTL     string
}

// LoadConfig loads configuration from a YAML file.
// Environment variables take precedence over YAML values.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Apply defaults with environment variable precedence
	if cfg.Ollama.Host == "" {
		cfg.Ollama.Host = getEnv(EnvOllamaHost, FallbackOllamaHost)
	} else if envHost := os.Getenv(EnvOllamaHost); envHost != "" {
		cfg.Ollama.Host = envHost
	}

	if cfg.Ollama.Timeout == 0 {
		cfg.Ollama.Timeout = getEnvInt(EnvOllamaTimeout, FallbackOllamaTimeout)
	} else if envTimeout := os.Getenv(EnvOllamaTimeout); envTimeout != "" {
		if t, err := strconv.Atoi(envTimeout); err == nil {
			cfg.Ollama.Timeout = t
		}
	}

	if cfg.Memory.StorageDir == "" {
		cfg.Memory.StorageDir = getEnv(EnvMemoryStorageDir, "")
	} else if envDir := os.Getenv(EnvMemoryStorageDir); envDir != "" {
		cfg.Memory.StorageDir = envDir
	}

	if cfg.Memory.DefaultMaxSize == 0 {
		cfg.Memory.DefaultMaxSize = getEnvInt(EnvMemoryDefaultMaxSize, FallbackMemoryMaxSize)
	}

	if cfg.Memory.DefaultTTL == "" {
		cfg.Memory.DefaultTTL = getEnv(EnvMemoryDefaultTTL, FallbackMemoryTTL)
	}

	// Set default coordinator values
	if cfg.Coordinator.ID == "" {
		cfg.Coordinator.ID = "coordinator"
	}
	if cfg.Coordinator.MaxWorkers == 0 {
		cfg.Coordinator.MaxWorkers = 3
	}

	// Set default output values
	if cfg.Output.ID == "" {
		cfg.Output.ID = "output"
	}
	if cfg.Output.MergeStrategy == "" {
		cfg.Output.MergeStrategy = "llm"
	}
	if cfg.Output.ResponseTimeout == "" {
		cfg.Output.ResponseTimeout = "60s"
	}

	return &cfg, nil
}

// ResolveWorkerMemory resolves memory configuration for a worker
func (c *Config) ResolveWorkerMemory(workerName string) ResolvedWorkerMemory {
	worker, ok := c.Workers[workerName]
	if !ok || len(worker.Memory.Types) == 0 {
		return ResolvedWorkerMemory{Enabled: false}
	}

	resolved := ResolvedWorkerMemory{
		Enabled: true,
		Types:   make([]ResolvedMemoryType, 0, len(worker.Memory.Types)),
	}

	for _, typeName := range worker.Memory.Types {
		rt := ResolvedMemoryType{
			Name:    typeName,
			MaxSize: c.Memory.DefaultMaxSize,
			TTL:     c.Memory.DefaultTTL,
		}

		// Override with type-specific settings
		if typeConfig, ok := c.Memory.Types[typeName]; ok {
			if typeConfig.MaxSize > 0 {
				rt.MaxSize = typeConfig.MaxSize
			}
			if typeConfig.TTL != "" {
				rt.TTL = typeConfig.TTL
			}
		}

		// Override with worker-specific settings
		if worker.Memory.MaxSize > 0 {
			rt.MaxSize = worker.Memory.MaxSize
		}
		if worker.Memory.TTL != "" {
			rt.TTL = worker.Memory.TTL
		}

		resolved.Types = append(resolved.Types, rt)
	}

	return resolved
}

// GetWorkerIDs returns all configured worker IDs
func (c *Config) GetWorkerIDs() []string {
	ids := make([]string, 0, len(c.Workers))
	for id := range c.Workers {
		ids = append(ids, id)
	}
	return ids
}
