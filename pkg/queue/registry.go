package queue

import (
	"errors"
	"sync"
)

var (
	// registry stores the mapping between job names (display names) and handlers
	registry = make(map[string]Handler)
	mu       sync.RWMutex
)

// Register adds a handler for a given job name (usually the Laravel class name)
func Register(name string, handler Handler) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = handler
}

// GetHandler retrieves a handler by name
func GetHandler(name string) (Handler, error) {
	mu.RLock()
	defer mu.RUnlock()
	if handler, ok := registry[name]; ok {
		return handler, nil
	}
	return nil, errors.New("handler not found: " + name)
}
