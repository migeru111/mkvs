package kvs

import "sync"

// HashMap is a thread-safe in-memory KVS backed by a Go map + RWMutex.
//
// Read-heavy workloads benefit from RLock allowing concurrent reads.
// Write operations take an exclusive lock.
type HashMap struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewHashMap returns a new, empty HashMap KVS.
func NewHashMap() *HashMap {
	return &HashMap{data: make(map[string]string)}
}

func (h *HashMap) Get(key string) (string, bool) {
	h.mu.RLock()
	v, ok := h.data[key]
	h.mu.RUnlock()
	return v, ok
}

func (h *HashMap) Set(key, value string) error {
	h.mu.Lock()
	h.data[key] = value
	h.mu.Unlock()
	return nil
}

func (h *HashMap) Delete(key string) bool {
	h.mu.Lock()
	_, ok := h.data[key]
	if ok {
		delete(h.data, key)
	}
	h.mu.Unlock()
	return ok
}

func (h *HashMap) Len() int {
	h.mu.RLock()
	n := len(h.data)
	h.mu.RUnlock()
	return n
}

func (h *HashMap) Close() error { return nil }
