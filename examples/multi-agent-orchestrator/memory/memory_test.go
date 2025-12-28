package memory

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
)

// =============================================================================
// STORE TESTS
// =============================================================================

func TestNewStore(t *testing.T) {
	store := NewStore("test-agent", "conversation", 100, time.Hour, "")

	if store.agentName != "test-agent" {
		t.Errorf("agentName = %q, want %q", store.agentName, "test-agent")
	}
	if store.memType != "conversation" {
		t.Errorf("memType = %q, want %q", store.memType, "conversation")
	}
	if store.maxSize != 100 {
		t.Errorf("maxSize = %d, want %d", store.maxSize, 100)
	}
	if store.ttl != time.Hour {
		t.Errorf("ttl = %v, want %v", store.ttl, time.Hour)
	}
	if store.storePath != "" {
		t.Errorf("storePath = %q, want empty", store.storePath)
	}
}

func TestNewStore_WithStorageDir(t *testing.T) {
	store := NewStore("test-agent", "conversation", 100, time.Hour, "/tmp/test")

	expectedPath := "/tmp/test/test-agent_conversation.json"
	if store.storePath != expectedPath {
		t.Errorf("storePath = %q, want %q", store.storePath, expectedPath)
	}
}

func TestStore_Add(t *testing.T) {
	store := NewStore("test", "conv", 10, time.Hour, "")

	entry := Entry{Role: "user", Content: "Hello"}
	store.Add(entry)

	stats := store.Stats()
	if stats["entries"] != 1 {
		t.Errorf("entries = %d, want %d", stats["entries"], 1)
	}
}

func TestStore_Add_AutoTimestamp(t *testing.T) {
	store := NewStore("test", "conv", 10, time.Hour, "")

	before := time.Now()
	entry := Entry{Role: "user", Content: "Hello"}
	store.Add(entry)
	after := time.Now()

	entries := store.GetAll()
	if len(entries) != 1 {
		t.Fatal("Expected 1 entry")
	}

	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", ts, before, after)
	}
}

func TestStore_Add_MaxSize_FIFO(t *testing.T) {
	store := NewStore("test", "conv", 3, time.Hour, "")

	// Add 5 entries to a store with maxSize 3
	for i := 1; i <= 5; i++ {
		store.Add(Entry{Role: "user", Content: string(rune('A' - 1 + i))}) // A, B, C, D, E
	}

	entries := store.GetAll()
	if len(entries) != 3 {
		t.Fatalf("entries = %d, want 3", len(entries))
	}

	// Should have C, D, E (first two evicted)
	expected := []string{"C", "D", "E"}
	for i, e := range entries {
		if e.Content != expected[i] {
			t.Errorf("entries[%d].Content = %q, want %q", i, e.Content, expected[i])
		}
	}
}

func TestStore_GetRecent(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "")

	// Add 10 entries
	for i := 0; i < 10; i++ {
		store.Add(Entry{Role: "user", Content: string(rune('0' + i))})
	}

	tests := []struct {
		name     string
		n        int
		wantLen  int
		wantLast string
	}{
		{"get 3", 3, 3, "9"},
		{"get 5", 5, 5, "9"},
		{"get 0", 0, 10, "9"},      // 0 means all
		{"get -1", -1, 10, "9"},    // negative means all
		{"get 100", 100, 10, "9"},  // more than available
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := store.GetRecent(tt.n)
			if len(entries) != tt.wantLen {
				t.Errorf("GetRecent(%d) len = %d, want %d", tt.n, len(entries), tt.wantLen)
			}
			if len(entries) > 0 && entries[len(entries)-1].Content != tt.wantLast {
				t.Errorf("GetRecent(%d) last = %q, want %q", tt.n, entries[len(entries)-1].Content, tt.wantLast)
			}
		})
	}
}

func TestStore_GetRecent_TTLExpiration(t *testing.T) {
	// Very short TTL for testing
	store := NewStore("test", "conv", 100, 10*time.Millisecond, "")

	store.Add(Entry{Role: "user", Content: "old"})

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	entries := store.GetRecent(10)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after TTL expiration, got %d", len(entries))
	}
}

func TestStore_GetRecent_ZeroTTL_NoExpiry(t *testing.T) {
	store := NewStore("test", "conv", 100, 0, "") // TTL = 0 means no expiry

	store.Add(Entry{Role: "user", Content: "permanent"})

	// Even after some time, entry should still be there
	time.Sleep(10 * time.Millisecond)

	entries := store.GetRecent(10)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry with TTL=0, got %d", len(entries))
	}
}

func TestStore_GetAll(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "")

	for i := 0; i < 5; i++ {
		store.Add(Entry{Role: "user", Content: "test"})
	}

	entries := store.GetAll()
	if len(entries) != 5 {
		t.Errorf("GetAll() len = %d, want 5", len(entries))
	}
}

func TestStore_Clear(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "")

	store.Add(Entry{Role: "user", Content: "test"})
	store.Add(Entry{Role: "assistant", Content: "response"})

	if store.Stats()["entries"] != 2 {
		t.Fatal("Expected 2 entries before clear")
	}

	store.Clear()

	if store.Stats()["entries"] != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", store.Stats()["entries"])
	}
}

func TestStore_Stats(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "")

	store.Add(Entry{Role: "user", Content: "Hello"})           // 4 + 5 = 9
	store.Add(Entry{Role: "assistant", Content: "Hi there"})   // 9 + 8 = 17

	stats := store.Stats()

	if stats["entries"] != 2 {
		t.Errorf("stats[entries] = %d, want 2", stats["entries"])
	}

	expectedSize := len("user") + len("Hello") + len("assistant") + len("Hi there")
	if stats["size"] != expectedSize {
		t.Errorf("stats[size] = %d, want %d", stats["size"], expectedSize)
	}
}

func TestStore_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and populate store
	store1 := NewStore("test", "conv", 100, time.Hour, tmpDir)
	store1.Add(Entry{Role: "user", Content: "Hello"})
	store1.Add(Entry{Role: "assistant", Content: "Hi"})

	// Save
	if err := store1.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, "test_conv.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Save() did not create file")
	}

	// Create new store and load
	store2 := NewStore("test", "conv", 100, time.Hour, tmpDir)
	if err := store2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	entries := store2.GetAll()
	if len(entries) != 2 {
		t.Errorf("Load() entries = %d, want 2", len(entries))
	}

	if entries[0].Content != "Hello" || entries[1].Content != "Hi" {
		t.Error("Load() content mismatch")
	}
}

func TestStore_Save_NoPath(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "") // No storage dir

	// Should be no-op, not error
	if err := store.Save(); err != nil {
		t.Errorf("Save() with no path should not error, got: %v", err)
	}
}

func TestStore_Load_NoPath(t *testing.T) {
	store := NewStore("test", "conv", 100, time.Hour, "") // No storage dir

	// Should be no-op, not error
	if err := store.Load(); err != nil {
		t.Errorf("Load() with no path should not error, got: %v", err)
	}
}

func TestStore_Load_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore("nonexistent", "conv", 100, time.Hour, tmpDir)

	// Should be no-op when file doesn't exist
	if err := store.Load(); err != nil {
		t.Errorf("Load() for nonexistent file should not error, got: %v", err)
	}
}

func TestStore_Load_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_conv.json")

	// Write invalid JSON
	if err := os.WriteFile(filePath, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewStore("test", "conv", 100, time.Hour, tmpDir)
	err := store.Load()

	if err == nil {
		t.Error("Load() should error on corrupted file")
	}
}

func TestStore_Concurrent(t *testing.T) {
	store := NewStore("test", "conv", 1000, time.Hour, "")

	var wg sync.WaitGroup
	numGoroutines := 10
	entriesPerGoroutine := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < entriesPerGoroutine; j++ {
				store.Add(Entry{Role: "user", Content: "test"})
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < entriesPerGoroutine; j++ {
				_ = store.GetRecent(10)
				_ = store.Stats()
			}
		}()
	}

	wg.Wait()

	// Should have maxSize entries (1000)
	stats := store.Stats()
	if stats["entries"] != 1000 {
		t.Errorf("Concurrent test: entries = %d, want 1000", stats["entries"])
	}
}

// =============================================================================
// MANAGER TESTS
// =============================================================================

func TestNewManager(t *testing.T) {
	cfg := config.MemoryConfig{
		DefaultMaxSize: 100,
		DefaultTTL:     "1h",
		StorageDir:     "/tmp/test",
	}

	mgr := NewManager(cfg)

	if mgr.storageDir != "/tmp/test" {
		t.Errorf("storageDir = %q, want %q", mgr.storageDir, "/tmp/test")
	}
	if len(mgr.stores) != 0 {
		t.Errorf("stores should be empty, got %d", len(mgr.stores))
	}
}

func TestManager_Initialize(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "memory", "test")

	cfg := config.MemoryConfig{StorageDir: subDir}
	mgr := NewManager(cfg)

	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	// Directory should exist
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("Initialize() did not create directory")
	}
}

func TestManager_Initialize_NoStorageDir(t *testing.T) {
	cfg := config.MemoryConfig{StorageDir: ""}
	mgr := NewManager(cfg)

	// Should be no-op
	if err := mgr.Initialize(); err != nil {
		t.Errorf("Initialize() with empty dir should not error, got: %v", err)
	}
}

func TestManager_GetStore(t *testing.T) {
	cfg := config.MemoryConfig{
		DefaultMaxSize: 50,
		DefaultTTL:     "30m",
	}
	mgr := NewManager(cfg)

	store1 := mgr.GetStore("agent1")
	store2 := mgr.GetStore("agent1")

	// Should return same store
	if store1 != store2 {
		t.Error("GetStore() should return same store for same agent")
	}

	// Default type should be conversation
	if store1.memType != "conversation" {
		t.Errorf("Default memType = %q, want %q", store1.memType, "conversation")
	}
}

func TestManager_GetStoreByType(t *testing.T) {
	cfg := config.MemoryConfig{
		DefaultMaxSize: 50,
		DefaultTTL:     "30m",
		Types: map[string]config.TypeConfig{
			"glossary": {
				MaxSize: 200,
				TTL:     "720h", // 30 days
			},
		},
	}
	mgr := NewManager(cfg)

	// Get conversation store (default config)
	convStore := mgr.GetStoreByType("agent1", "conversation")
	if convStore.maxSize != 50 {
		t.Errorf("conversation maxSize = %d, want 50", convStore.maxSize)
	}

	// Get glossary store (custom config)
	glossaryStore := mgr.GetStoreByType("agent1", "glossary")
	if glossaryStore.maxSize != 200 {
		t.Errorf("glossary maxSize = %d, want 200", glossaryStore.maxSize)
	}

	// Different types should be different stores
	if convStore == glossaryStore {
		t.Error("Different memory types should return different stores")
	}
}

func TestManager_GetStoreByType_DoubleCheckLocking(t *testing.T) {
	cfg := config.MemoryConfig{
		DefaultMaxSize: 50,
		DefaultTTL:     "1h",
	}
	mgr := NewManager(cfg)

	var wg sync.WaitGroup
	stores := make(chan *Store, 10)

	// Multiple goroutines requesting same store
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store := mgr.GetStoreByType("agent", "conv")
			stores <- store
		}()
	}

	wg.Wait()
	close(stores)

	// All should be the same store
	var first *Store
	for store := range stores {
		if first == nil {
			first = store
		} else if store != first {
			t.Error("Concurrent GetStoreByType returned different stores")
		}
	}
}

func TestManager_SaveAll(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.MemoryConfig{
		DefaultMaxSize: 50,
		DefaultTTL:     "1h",
		StorageDir:     tmpDir,
	}
	mgr := NewManager(cfg)
	mgr.Initialize()

	// Create stores and add data
	store1 := mgr.GetStoreByType("agent1", "conv")
	store1.Add(Entry{Role: "user", Content: "test1"})

	store2 := mgr.GetStoreByType("agent2", "conv")
	store2.Add(Entry{Role: "user", Content: "test2"})

	// Save all
	if err := mgr.SaveAll(); err != nil {
		t.Fatalf("SaveAll() error: %v", err)
	}

	// Verify files exist
	file1 := filepath.Join(tmpDir, "agent1_conv.json")
	file2 := filepath.Join(tmpDir, "agent2_conv.json")

	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("SaveAll() did not create file for agent1")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("SaveAll() did not create file for agent2")
	}
}

func TestManager_Shutdown(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.MemoryConfig{
		DefaultMaxSize: 50,
		DefaultTTL:     "1h",
		StorageDir:     tmpDir,
	}
	mgr := NewManager(cfg)
	mgr.Initialize()

	store := mgr.GetStore("agent")
	store.Add(Entry{Role: "user", Content: "test"})

	// Shutdown should save
	if err := mgr.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}

	// Verify saved
	filePath := filepath.Join(tmpDir, "agent_conversation.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Shutdown() did not save stores")
	}
}

// =============================================================================
// UTILITY FUNCTION TESTS
// =============================================================================

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"24h", 24 * time.Hour},
		{"720h", 720 * time.Hour}, // 30 days
		{"", DefaultTTLFallback},
		{"invalid", DefaultTTLFallback},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseDuration(tt.input)
			if got != tt.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// BENCHMARK TESTS
// =============================================================================

func BenchmarkStore_Add(b *testing.B) {
	store := NewStore("bench", "conv", 1000, time.Hour, "")
	entry := Entry{Role: "user", Content: "benchmark content"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Add(entry)
	}
}

func BenchmarkStore_GetRecent(b *testing.B) {
	store := NewStore("bench", "conv", 1000, time.Hour, "")
	for i := 0; i < 500; i++ {
		store.Add(Entry{Role: "user", Content: "test"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetRecent(10)
	}
}

func BenchmarkStore_Concurrent(b *testing.B) {
	store := NewStore("bench", "conv", 10000, time.Hour, "")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.Add(Entry{Role: "user", Content: "test"})
			store.GetRecent(5)
		}
	})
}
