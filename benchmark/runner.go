package benchmark

import (
	"fmt"
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
	readRec := NewRecorder(cfg.OperationCount)
	updateRec := NewRecorder(cfg.OperationCount)
	insertRec := NewRecorder(cfg.OperationCount)
	rmwRec := NewRecorder(cfg.OperationCount)

	g := NewGenerator(cfg.Distribution, int64(cfg.RecordCount))
	insertSeq := cfg.RecordCount
	totalOps := 0

	start := time.Now()

	for i := 0; i < cfg.OperationCount; i++ {
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
			key := KeyForIndex(insertSeq)
			insertSeq++
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

		totalOps++
	}

	elapsed := time.Since(start)

	return Result{
		Workload:   cfg.Workload,
		Elapsed:    elapsed,
		TotalOps:   int64(totalOps),
		Throughput: float64(totalOps) / elapsed.Seconds(),
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
