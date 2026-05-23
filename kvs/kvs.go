// Package kvs defines the KVS interface and its implementations.
package kvs

// KVS is the interface every key-value store implementation must satisfy.
type KVS interface {
	Get(key string) (string, bool)
	Set(key string, value string) error
	Delete(key string) bool
	Len() int
	Close() error
}
