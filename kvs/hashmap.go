package kvs

type HashMap struct {
	data map[string]string
}

func NewHashMap() *HashMap {
	return &HashMap{data: make(map[string]string)}
}

func (h *HashMap) Get(key string) (string, bool) {
	v, ok := h.data[key]
	return v, ok
}

func (h *HashMap) Set(key, value string) error {
	h.data[key] = value
	return nil
}

func (h *HashMap) Delete(key string) bool {
	_, ok := h.data[key]
	if ok {
		delete(h.data, key)
	}
	return ok
}

func (h *HashMap) Len() int {
	return len(h.data)
}

func (h *HashMap) Close() error { return nil }
