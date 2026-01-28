package storage

import "sync"

// Store is a minimal interface to record whether an account has attended today.
type Store interface {
	HasAttended(key string) (bool, error)
	MarkAttended(key string) error
}

// MemoryStore is an in-memory implementation suitable for single run executions
// such as Docker one-shot containers, QingLong tasks, or a single cloud function
// invocation.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]struct{}
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]struct{}),
	}
}

// HasAttended checks whether the given key has been marked as attended.
func (s *MemoryStore) HasAttended(key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok, nil
}

// MarkAttended marks the given key as attended.
func (s *MemoryStore) MarkAttended(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = struct{}{}
	return nil
}

