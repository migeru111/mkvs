package kvs

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
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
}

func (h *HashMapMutexTCP) Get(key string) (string, bool) {
	// 1. サーバー（localhost:8080）に接続
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("サーバーへの接続に失敗しました: %v", err)
	}
	defer conn.Close()
	fmt.Println("サーバーに接続しました。メッセージを入力してください（'quit' で終了）:")

	// キーボード入力用のリーダー
	inputReader := bufio.NewReader(os.Stdin)
	// サーバーからの受信用のリーダー
	serverReader := bufio.NewReader(conn)

	_, err = conn.Write([]byte(key))
	if err != nil {
		log.Fatalf("サーバーへの送信に失敗しました: %v", err)
	}
	slog.Debug(key)

	// 4. サーバーからの返信を受信して表示
	response, err := serverReader.ReadString('\n')
	if err != nil {
		log.Fatalf("サーバーからの受信に失敗しました: %v", err)
	}

	slog.Debug(response)
}
