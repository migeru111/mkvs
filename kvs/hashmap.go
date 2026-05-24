package kvs

import "sync"

type HashMap struct {
	mu   sync.Mutex
	data map[string]string
}

func NewHashMap() *HashMap {
	return &HashMap{data: make(map[string]string)}
}

func (h *HashMap) Get(key string) (string, bool) {
	h.mu.Lock()
	v, ok := h.data[key]
	h.mu.Unlock()
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
	h.mu.Lock()
	n := len(h.data)
	h.mu.Unlock()
	return n
}

func (h *HashMap) Close() error { return nil }
