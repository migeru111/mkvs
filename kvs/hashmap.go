package kvs

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
)

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

type HashMapMutexTCP struct {
	mu   sync.Mutex
	data map[string]string
	conn net.Conn
}

func NewHashMapMutexTCP() *HashMapMutexTCP {
	return &HashMapMutexTCP{
		data: make(map[string]string),
		conn: MakeConnectionForSending(),
	}
}

func MakeConnectionForSending() net.Conn {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		slog.Error("サーバーへの接続に失敗しました", err)
	}
	return conn
}

func (h *HashMapMutexTCP) Get(key string) (string, bool) {

	message := fmt.Sprintf("GET(%s)\n", key)

	// GET: %s,%s\n
	res := request(message, h)

	return strings.Trim(res, "\n"), true

}

func (h *HashMapMutexTCP) Set(key, value string) error {

	message := fmt.Sprintf("SET(%s,%s)\n", key, value)
	_ = request(message, h)
	return nil
}

func (h *HashMapMutexTCP) Delete(key string) bool {
	return true
}

func (h *HashMapMutexTCP) Len() int {
	return len(h.data)
}

func (h *HashMapMutexTCP) Close() error {
	return nil
}

func request(message string, kvs *HashMapMutexTCP) string {
	// 1. サーバー（localhost:8080）に接続

	//defer kvs.conn.Close()
	slog.Debug("サーバーに接続しました。")

	// サーバーからの受信用のリーダー
	serverReader := bufio.NewReader(kvs.conn)

	_, err := kvs.conn.Write([]byte(message))
	if err != nil {
		slog.Error("サーバーへの送信に失敗しました", err)
	}
	slog.Debug(message)

	// 4. サーバーからの返信を受信して表示
	response, err := serverReader.ReadString('\n')
	if err != nil {
		slog.Error("サーバーからの受信に失敗しました", err)
	}

	slog.Debug(response)
	return response
}
