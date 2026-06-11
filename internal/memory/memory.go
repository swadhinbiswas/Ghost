package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const memoryDir = ".ghost"
const memoryFile = "memory.json"

// MemoryStore represents the persistent storage for project-specific knowledge.
// It allows the agent to remember facts, style guides, and user preferences across sessions.
type MemoryStore struct {
	mu       sync.RWMutex
	data     map[string]string
	basePath string
}

// NewMemoryStore initializes a new memory store in the given workspace root
func NewMemoryStore(workspaceRoot string) (*MemoryStore, error) {
	store := &MemoryStore{
		data:     make(map[string]string),
		basePath: workspaceRoot,
	}

	if err := store.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// getFilePath returns the absolute path to the memory configuration file
func (s *MemoryStore) getFilePath() string {
	return filepath.Join(s.basePath, memoryDir, memoryFile)
}

// Load reads the memory state from disk
func (s *MemoryStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.data)
}

// Save writes the current memory state to disk
func (s *MemoryStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.basePath, memoryDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.getFilePath(), data, 0644)
}

// Set stores a piece of information and saves it to disk
func (s *MemoryStore) Set(key, value string) error {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	return s.Save()
}

// Get retrieves a piece of information from memory
func (s *MemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Delete removes a key from memory
func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
	return s.Save()
}

// GetAll returns a copy of all memorized facts (useful for injecting into prompts)
func (s *MemoryStore) GetAll() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copyState := make(map[string]string, len(s.data))
	for k, v := range s.data {
		copyState[k] = v
	}

	return copyState
}
