package benchmark

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/migeru111/mkvs/kvs"
)

// Config holds all benchmark parameters.
type Config struct {
	RecordCount    int        // number of records inserted in the load phase
	OperationCount int        // number of operations in the run phase
	Threads        int        // number of concurrent goroutines
	ValueSize      int        // size of each value in bytes
	Workload       Workload   // YCSB workload preset
	Distribution   Distribution // key selection distribution
}

// Result holds the benchmark outcome.
type Result struct {
	Workload   Workload
	Elapsed    time.Duration
	TotalOps   int64
	Throughput float64 // ops/sec
	Read       Stats
	Update     Stats
	Insert     Stats
	RMW        Stats
}

// Run executes the load phase followed by the run phase and returns metrics.
func Run(store kvs.KVS, cfg Config) Result {
	wl := workloads[cfg.Workload]
	value := makeValue(cfg.ValueSize)

	// --- Load phase ---
	fmt.Printf("[Load] inserting %d records...\n", cfg.RecordCount)
	t0 := time.Now()
	for i := 0; i < cfg.RecordCount; i++ {
		_ = store.Set(KeyForIndex(i), value)
	}
	fmt.Printf("[Load] done in %v\n\n", time.Since(t0).Round(time.Millisecond))

	// --- Run phase ---
	opsPerThread := cfg.OperationCount / cfg.Threads
	totalExpected := opsPerThread * cfg.Threads

	readRec := NewRecorder(totalExpected)
	updateRec := NewRecorder(totalExpected)
	insertRec := NewRecorder(totalExpected / 10)
	rmwRec := NewRecorder(totalExpected)

	var totalOps int64
	// Monotonically increasing counter for insert keys (beyond initial load).
	var insertSeq uint64 = uint64(cfg.RecordCount)

	var wg sync.WaitGroup
	start := time.Now()

	for t := 0; t < cfg.Threads; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := NewGenerator(cfg.Distribution, int64(cfg.RecordCount))

			for i := 0; i < opsPerThread; i++ {
				op := g.Float64()
				cumR := wl.read
				cumU := cumR + wl.update
				cumI := cumU + wl.insert

				switch {
				case op < cumR:
					key := g.Next()
					t0 := time.Now()
					_, _ = store.Get(key)
					readRec.Add(time.Since(t0))

				case op < cumU:
					key := g.Next()
					t0 := time.Now()
					_ = store.Set(key, value)
					updateRec.Add(time.Since(t0))

				case op < cumI:
					// Atomic increment ensures each goroutine inserts a unique key.
					idx := atomic.AddUint64(&insertSeq, 1) - 1
					key := KeyForIndex(int(idx))
					t0 := time.Now()
					_ = store.Set(key, value)
					insertRec.Add(time.Since(t0))

				default: // RMW: read then write back
					key := g.Next()
					t0 := time.Now()
					v, ok := store.Get(key)
					if ok {
						_ = store.Set(key, v)
					}
					rmwRec.Add(time.Since(t0))
				}

				atomic.AddInt64(&totalOps, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	ops := atomic.LoadInt64(&totalOps)
	return Result{
		Workload:   cfg.Workload,
		Elapsed:    elapsed,
		TotalOps:   ops,
		Throughput: float64(ops) / elapsed.Seconds(),
		Read:       readRec.Stats(),
		Update:     updateRec.Stats(),
		Insert:     insertRec.Stats(),
		RMW:        rmwRec.Stats(),
	}
}

func makeValue(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}
