package memory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
)

// File permission constants for secure storage
const (
	FilePermission     = 0600
	DirPermission      = 0700
	DefaultTTLFallback = 24 * time.Hour
)

// Entry represents a single memory entry
type Entry struct {
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Store manages memory entries for a single agent
type Store struct {
	mu        sync.RWMutex
	entries   []Entry
	maxSize   int
	ttl       time.Duration
	storePath string
	agentName string
	memType   string
}

// NewStore creates a new memory store
func NewStore(agentName, memType string, maxSize int, ttl time.Duration, storageDir string) *Store {
	path := ""
	if storageDir != "" {
		path = filepath.Join(storageDir, fmt.Sprintf("%s_%s.json", agentName, memType))
	}

	return &Store{
		entries:   make([]Entry, 0),
		maxSize:   maxSize,
		ttl:       ttl,
		storePath: path,
		agentName: agentName,
		memType:   memType,
	}
}

// Add adds an entry to the store
func (s *Store) Add(entry Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	s.entries = append(s.entries, entry)

	// Trim to max size
	if len(s.entries) > s.maxSize {
		s.entries = s.entries[len(s.entries)-s.maxSize:]
	}
}

// GetRecent returns the most recent entries
func (s *Store) GetRecent(n int) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Remove expired entries
	now := time.Now()
	valid := make([]Entry, 0)
	for _, e := range s.entries {
		if s.ttl == 0 || now.Sub(e.Timestamp) < s.ttl {
			valid = append(valid, e)
		}
	}

	if n <= 0 || n > len(valid) {
		n = len(valid)
	}

	if n == 0 {
		return []Entry{}
	}
	return valid[len(valid)-n:]
}

// GetAll returns all valid entries
func (s *Store) GetAll() []Entry {
	return s.GetRecent(-1)
}

// Clear removes all entries
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make([]Entry, 0)
}

// Stats returns memory statistics
func (s *Store) Stats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := 0
	for _, e := range s.entries {
		size += len(e.Content) + len(e.Role)
	}

	return map[string]int{
		"entries": len(s.entries),
		"size":    size,
	}
}

// Save persists memory to disk
func (s *Store) Save() error {
	if s.storePath == "" {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}

	if err := os.WriteFile(s.storePath, data, FilePermission); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}

	return nil
}

// Load loads memory from disk
func (s *Store) Load() error {
	if s.storePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.storePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read memory: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, &s.entries); err != nil {
		return fmt.Errorf("unmarshal memory: %w", err)
	}

	return nil
}

// Manager manages memory stores for multiple agents
type Manager struct {
	mu         sync.RWMutex
	stores     map[string]*Store
	config     config.MemoryConfig
	storageDir string
}

// NewManager creates a new memory manager
func NewManager(cfg config.MemoryConfig) *Manager {
	return &Manager{
		stores:     make(map[string]*Store),
		config:     cfg,
		storageDir: cfg.StorageDir,
	}
}

// Initialize sets up the memory storage directory
func (m *Manager) Initialize() error {
	if m.storageDir != "" {
		if err := os.MkdirAll(m.storageDir, DirPermission); err != nil {
			return fmt.Errorf("create storage dir: %w", err)
		}
	}
	return nil
}

// GetStore returns or creates a store for an agent
func (m *Manager) GetStore(agentName string) *Store {
	return m.GetStoreByType(agentName, "conversation")
}

// GetStoreByType returns or creates a store for an agent and memory type
func (m *Manager) GetStoreByType(agentName, memType string) *Store {
	key := fmt.Sprintf("%s_%s", agentName, memType)

	m.mu.RLock()
	if store, ok := m.stores[key]; ok {
		m.mu.RUnlock()
		return store
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check after acquiring write lock
	if store, ok := m.stores[key]; ok {
		return store
	}

	// Get config for this memory type
	maxSize := m.config.DefaultMaxSize
	ttl := parseDuration(m.config.DefaultTTL)

	if typeConfig, ok := m.config.Types[memType]; ok {
		if typeConfig.MaxSize > 0 {
			maxSize = typeConfig.MaxSize
		}
		if typeConfig.TTL != "" {
			ttl = parseDuration(typeConfig.TTL)
		}
	}

	store := NewStore(agentName, memType, maxSize, ttl, m.storageDir)
	if err := store.Load(); err != nil {
		log.Printf("ERROR: Failed to load memory for %s: %v", key, err)
	}

	m.stores[key] = store
	return store
}

// SaveAll saves all stores to disk
func (m *Manager) SaveAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for key, store := range m.stores {
		if err := store.Save(); err != nil {
			return fmt.Errorf("save %s: %w", key, err)
		}
	}
	return nil
}

// Shutdown saves all stores and cleans up
func (m *Manager) Shutdown() error {
	return m.SaveAll()
}

func parseDuration(s string) time.Duration {
	if s == "" {
		return DefaultTTLFallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("WARNING: Invalid duration %q, using default %v: %v", s, DefaultTTLFallback, err)
		return DefaultTTLFallback
	}
	return d
}
