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

func main() {
	var (
		records   = flag.Int("records", 100_000, "records to pre-load")
		ops       = flag.Int("ops", 500_000, "operations to run per workload")
		valueSize = flag.Int("valuesize", 100, "value size in bytes")
		workload  = flag.String("workload", "A", "workload A/B/C/D/F")
		dist      = flag.String("dist", "zipfian", "key distribution: zipfian or uniform")
		all       = flag.Bool("all", false, "run all workloads sequentially")
		workers   = flag.Int("workers", 1, "number of parallel benchmark goroutines")
	)
	flag.Parse()

	cfg := benchmark.Config{
		RecordCount:    *records,
		OperationCount: *ops,
		ValueSize:      *valueSize,
		Distribution:   benchmark.Distribution(*dist),
		Concurrency:    *workers,
	}

	var workloads []benchmark.Workload
	if *all {
		workloads = benchmark.AllWorkloads
	} else {
		workloads = []benchmark.Workload{benchmark.Workload(strings.ToUpper(*workload))}
	}

	var results []benchmark.Result
	for _, wl := range workloads {
		cfg.Workload = wl
		store := kvs.NewHashMap()
		res := benchmark.Run(store, cfg)
		_ = store.Close()
		results = append(results, res)
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
