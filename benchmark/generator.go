package benchmark

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// Generator provides per-goroutine random key generation.
// Each goroutine must own its own Generator to avoid lock contention.
type Generator struct {
	r    *rand.Rand
	zipf *rand.Zipf
	n    int64
	dist Distribution
}

// unique seed counter so concurrent goroutines don't share the same seed.
var seedCounter uint64

// NewGenerator creates a Generator for the given distribution and key space size.
// Must be called once per goroutine — not goroutine-safe.
func NewGenerator(dist Distribution, recordCount int64) *Generator {
	seed := time.Now().UnixNano() ^ int64(atomic.AddUint64(&seedCounter, 1)*0x9e3779b97f4a7c15)
	r := rand.New(rand.NewSource(seed))
	g := &Generator{r: r, n: recordCount, dist: dist}
	if dist == Zipfian {
		// s=1.2, v=1.0 are the standard YCSB Zipfian parameters.
		// They produce a realistic hot-key distribution where the top ~20% of
		// keys receive ~80% of the traffic.
		g.zipf = rand.NewZipf(r, 1.2, 1.0, uint64(recordCount-1))
	}
	return g
}

// Next returns the next key according to the configured distribution.
func (g *Generator) Next() string {
	var n int64
	if g.dist == Zipfian {
		n = int64(g.zipf.Uint64())
	} else {
		n = g.r.Int63n(g.n)
	}
	return fmt.Sprintf("key%010d", n)
}

// Float64 returns a random float in [0.0, 1.0) for operation selection.
func (g *Generator) Float64() float64 { return g.r.Float64() }

// KeyForIndex returns the deterministic key string for a given index.
func KeyForIndex(i int) string { return fmt.Sprintf("key%010d", i) }
