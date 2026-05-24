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
	RecordCount    int          // number of records inserted in the load phase
	OperationCount int          // number of operations in the run phase
	ValueSize      int          // size of each value in bytes
	Workload       Workload     // YCSB workload preset
	Distribution   Distribution // key selection distribution
	Concurrency    int          // number of parallel worker goroutines (0 or 1 = sequential)
}

// Result holds the benchmark outcome.
type Result struct {
	Workload   Workload
	Workers    int
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
	readRec := NewRecorder(cfg.OperationCount)
	updateRec := NewRecorder(cfg.OperationCount)
	insertRec := NewRecorder(cfg.OperationCount)
	rmwRec := NewRecorder(cfg.OperationCount)

	workers := cfg.Concurrency
	if workers <= 1 {
		workers = 1
	}

	opsPerWorker := cfg.OperationCount / workers
	remainder := cfg.OperationCount % workers
	insertSeq := int64(cfg.RecordCount)

	var wg sync.WaitGroup
	start := time.Now()

	for w := 0; w < workers; w++ {
		myOps := opsPerWorker
		if w < remainder {
			myOps++
		}
		wg.Add(1)
		go func(ops int) {
			defer wg.Done()
			g := NewGenerator(cfg.Distribution, int64(cfg.RecordCount))
			for i := 0; i < ops; i++ {
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
					idx := atomic.AddInt64(&insertSeq, 1) - 1
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
			}
		}(myOps)
	}

	wg.Wait()
	elapsed := time.Since(start)

	return Result{
		Workload:   cfg.Workload,
		Workers:    workers,
		Elapsed:    elapsed,
		TotalOps:   int64(cfg.OperationCount),
		Throughput: float64(cfg.OperationCount) / elapsed.Seconds(),
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
