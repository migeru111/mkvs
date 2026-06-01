package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/migeru111/mkvs/benchmark"
	"github.com/migeru111/mkvs/kvs"
)

type storeEntry struct {
	name string
	new  func() kvs.KVS
}

var allStores = []storeEntry{
	{"HashMap/Mutex", func() kvs.KVS { return kvs.NewHashMapMutex() }},
	{"HashMap/RWMutex", func() kvs.KVS { return kvs.NewHashMapRWMutex() }},
	{"HashMap/RWMutexTCP", func() kvs.KVS { return kvs.NewHashMapMutexTCP() }},
}

func selectStores(flag string) []storeEntry {
	switch strings.ToLower(flag) {
	case "mutex":
		return allStores[:1]
	case "rwmutex":
		return allStores[1:]
	default:
		return allStores
	}
}

func main() {
	var (
		records    = flag.Int("records", 100_000, "records to pre-load")
		ops        = flag.Int("ops", 500_000, "operations to run per workload")
		valueSize  = flag.Int("valuesize", 100, "value size in bytes")
		workload   = flag.String("workload", "A", "workload A/B/C/D/F")
		dist       = flag.String("dist", "zipfian", "key distribution: zipfian or uniform")
		all        = flag.Bool("all", false, "run all workloads sequentially")
		workers    = flag.Int("workers", 0, "goroutine数 (0=スケーリング計測)")
		maxWorkers = flag.Int("maxworkers", 8, "スケーリング計測の最大ワーカー数 (2の累乗まで)")
		store      = flag.String("store", "all", "KVS実装 (all/mutex/rwmutex)")
	)
	flag.Parse()

	var scalingSteps []int
	if *workers != 0 {
		scalingSteps = []int{*workers}
	} else {
		for w := 1; w <= *maxWorkers; w *= 2 {
			scalingSteps = append(scalingSteps, w)
		}
	}

	cfg := benchmark.Config{
		RecordCount:    *records,
		OperationCount: *ops,
		ValueSize:      *valueSize,
		Distribution:   benchmark.Distribution(*dist),
	}

	var workloads []benchmark.Workload
	if *all {
		workloads = benchmark.AllWorkloads
	} else {
		workloads = []benchmark.Workload{benchmark.Workload(strings.ToUpper(*workload))}
	}

	stores := selectStores(*store)

	var results []benchmark.Result
	for _, s := range stores {
		cfg.StoreName = s.name
		for _, w := range scalingSteps {
			cfg.Concurrency = w
			for _, wl := range workloads {
				cfg.Workload = wl
				store := s.new()
				res := benchmark.Run(store, cfg)
				_ = store.Close()
				results = append(results, res)
			}
		}
	}

	path, err := writeReport(cfg, results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "report: %v\n", err)
		return
	}
	fmt.Printf("\nReport saved: %s\n", path)
}

// combinedAvg returns a weighted average latency across all operation types.
func combinedAvg(r benchmark.Result) time.Duration {
	total, count := int64(0), int64(0)
	for _, s := range []benchmark.Stats{r.Read, r.Update, r.Insert, r.RMW} {
		if !s.Empty() {
			total += s.Avg.Nanoseconds() * int64(s.Count)
			count += int64(s.Count)
		}
	}
	if count == 0 {
		return 0
	}
	return time.Duration(total / count)
}

func fmtDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// combinedP99 returns the max P99 across all operation types.
func combinedP99(r benchmark.Result) time.Duration {
	var max time.Duration
	for _, s := range []benchmark.Stats{r.Read, r.Update, r.Insert, r.RMW} {
		if !s.Empty() && s.P99 > max {
			max = s.P99
		}
	}
	return max
}
