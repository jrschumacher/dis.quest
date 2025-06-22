package session

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStorage implements Storage interface using in-memory storage.
// This is suitable for development, testing, and single-instance applications.
// Data is lost when the application restarts.
type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]*Data
}

// NewMemoryStorage creates a new memory-based session storage.
func NewMemoryStorage() Storage {
	return &MemoryStorage{
		data: make(map[string]*Data),
	}
}

// Store saves session data with the given key.
func (m *MemoryStorage) Store(ctx context.Context, key string, data *Data) error {
	if key == "" {
		return fmt.Errorf("session key cannot be empty")
	}
	if data == nil {
		return fmt.Errorf("session data cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a deep copy of the data to avoid external modifications
	dataCopy := *data
	if data.Metadata != nil {
		dataCopy.Metadata = make(map[string]interface{})
		for k, v := range data.Metadata {
			dataCopy.Metadata[k] = v
		}
	}

	m.data[key] = &dataCopy
	return nil
}

// Load retrieves session data by key.
func (m *MemoryStorage) Load(ctx context.Context, key string) (*Data, error) {
	if key == "" {
		return nil, fmt.Errorf("session key cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session has expired
	if time.Now().After(data.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Return a copy to avoid external modifications
	dataCopy := *data
	if data.Metadata != nil {
		dataCopy.Metadata = make(map[string]interface{})
		for k, v := range data.Metadata {
			dataCopy.Metadata[k] = v
		}
	}

	return &dataCopy, nil
}

// Delete removes session data.
func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("session key cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// Cleanup removes expired sessions.
func (m *MemoryStorage) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	var expiredKeys []string

	// Find expired sessions
	for key, data := range m.data {
		if now.After(data.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired sessions
	for _, key := range expiredKeys {
		delete(m.data, key)
	}

	return nil
}

// Close cleans up storage resources (no-op for memory storage).
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all data
	m.data = make(map[string]*Data)
	return nil
}

// GetSessionCount returns the number of active sessions (for testing/debugging).
func (m *MemoryStorage) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// GetAllSessionIDs returns all session IDs (for testing/debugging).
func (m *MemoryStorage) GetAllSessionIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}