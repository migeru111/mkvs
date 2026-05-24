package kvs

import "sync"

// HashMapMutex uses sync.Mutex — all operations are exclusive.
type HashMapMutex struct {
	mu   sync.Mutex
	data map[string]string
}

func NewHashMapMutex() *HashMapMutex {
	return &HashMapMutex{data: make(map[string]string)}
}

func (h *HashMapMutex) Get(key string) (string, bool) {
	h.mu.Lock()
	v, ok := h.data[key]
	h.mu.Unlock()
	return v, ok
}

func (h *HashMapMutex) Set(key, value string) error {
	h.mu.Lock()
	h.data[key] = value
	h.mu.Unlock()
	return nil
}

func (h *HashMapMutex) Delete(key string) bool {
	h.mu.Lock()
	_, ok := h.data[key]
	if ok {
		delete(h.data, key)
	}
	h.mu.Unlock()
	return ok
}

func (h *HashMapMutex) Len() int {
	h.mu.Lock()
	n := len(h.data)
	h.mu.Unlock()
	return n
}

func (h *HashMapMutex) Close() error { return nil }

// HashMapRWMutex uses sync.RWMutex — concurrent reads are allowed.
type HashMapRWMutex struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewHashMapRWMutex() *HashMapRWMutex {
	return &HashMapRWMutex{data: make(map[string]string)}
}

func (h *HashMapRWMutex) Get(key string) (string, bool) {
	h.mu.RLock()
	v, ok := h.data[key]
	h.mu.RUnlock()
	return v, ok
}

func (h *HashMapRWMutex) Set(key, value string) error {
	h.mu.Lock()
	h.data[key] = value
	h.mu.Unlock()
	return nil
}

func (h *HashMapRWMutex) Delete(key string) bool {
	h.mu.Lock()
	_, ok := h.data[key]
	if ok {
		delete(h.data, key)
	}
	h.mu.Unlock()
	return ok
}

func (h *HashMapRWMutex) Len() int {
	h.mu.RLock()
	n := len(h.data)
	h.mu.RUnlock()
	return n
}

func (h *HashMapRWMutex) Close() error { return nil }
