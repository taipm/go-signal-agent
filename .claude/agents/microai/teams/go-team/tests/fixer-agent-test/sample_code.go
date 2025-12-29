// Package sample contains code with various issues for testing Fixer Agent
package sample

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// ═══════════════════════════════════════════════════════════════════════════
// FIXED BY FIXER AGENT
// ═══════════════════════════════════════════════════════════════════════════

// GetUser retrieves a user by their unique identifier.
// Returns the user and nil error on success, or nil and error on failure.
func GetUser(id string) (*User, error) {
	return nil, nil
}

// FetchData retrieves data from the specified URL.
// Returns the response body bytes or an error with context.
func FetchData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch data from %s: %w", url, err) // ✅ FIXED: wrapped error
	}
	defer resp.Body.Close()
	return nil, nil
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	return email != "" && strings.Contains(email, "@") && strings.Contains(email, ".")
}

// UpdateEmail handles email update requests.
// Validates email format before processing.
func UpdateEmail(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	// ✅ FIXED: Added input validation
	if !isValidEmail(email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Email updated to: %s", email)
}

// Counter provides a thread-safe counter implementation.
type Counter struct {
	mu    sync.Mutex // ✅ FIXED: Added mutex for synchronization
	value int
}

// Increment safely adds 1 to the counter value.
func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++ // ✅ FIXED: Now race-free
}

// Get safely returns the current counter value.
func (c *Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// ═══════════════════════════════════════════════════════════════════════════
// ESCALATED TO CODER AGENT (Complex fixes > 20 lines)
// ═══════════════════════════════════════════════════════════════════════════

// ⚠️ ESCALATED Issue 5: SQL injection vulnerability
// Reason: Requires parameterized query redesign (~15 lines, security-critical)
// Route: Coder Agent
func GetUserByName(db *sql.DB, name string) (*User, error) {
	// TODO: CODER AGENT - Fix SQL injection with parameterized query
	// CURRENT (VULNERABLE):
	query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return nil, nil
}

// ⚠️ ESCALATED Issue 6: O(n²) algorithm needs rewrite
// Reason: Requires algorithm redesign with hash map (~25 lines)
// Route: Coder Agent
func ProcessData(data []Item) []Result {
	// TODO: CODER AGENT - Rewrite with O(n) hash map approach
	// CURRENT (INEFFICIENT O(n²)):
	var results []Result
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data); j++ {
			if i != j && data[i].Key == data[j].Key {
				results = append(results, Result{
					Key:   data[i].Key,
					Count: 2,
				})
			}
		}
	}
	return results
}

// Types
type User struct {
	ID    string
	Name  string
	Email string
}

type Item struct {
	Key   string
	Value interface{}
}

type Result struct {
	Key   string
	Count int
}
