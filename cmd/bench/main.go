// bench is a YCSB-inspired benchmark tool for mkvs.
//
// Usage:
//
//	go run ./cmd/bench [flags]
//
// Example — run all workloads with 8 threads:
//
//	go run ./cmd/bench -all -threads 8
//
// Example — run workload A with Zipfian key distribution:
//
//	go run ./cmd/bench -workload A -dist zipfian -ops 500000
package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/migeru111/mkvs/benchmark"
	"github.com/migeru111/mkvs/kvs"
)

func main() {
	var (
		records   = flag.Int("records", 100_000, "records to pre-load")
		ops       = flag.Int("ops", 500_000, "operations to run per workload")
		threads   = flag.Int("threads", 4, "concurrent goroutines")
		valueSize = flag.Int("valuesize", 100, "value size in bytes")
		workload  = flag.String("workload", "A", "workload A/B/C/D/F")
		dist      = flag.String("dist", "zipfian", "key distribution: zipfian or uniform")
		all       = flag.Bool("all", false, "run all workloads sequentially")
	)
	flag.Parse()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  mkvs — YCSB-like Benchmark")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  Store       : HashMap (sync.Mutex)\n")
	fmt.Printf("  Distribution: %s\n", *dist)
	fmt.Printf("  Records     : %d\n", *records)
	fmt.Printf("  Operations  : %d per workload\n", *ops)
	fmt.Printf("  Threads     : %d\n", *threads)
	fmt.Printf("  Value size  : %d bytes\n", *valueSize)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	cfg := benchmark.Config{
		RecordCount:    *records,
		OperationCount: *ops,
		Threads:        *threads,
		ValueSize:      *valueSize,
		Distribution:   benchmark.Distribution(*dist),
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

		fmt.Printf("--- Workload %s ---\n", benchmark.Describe(wl))
		res := benchmark.Run(store, cfg)
		_ = store.Close()

		results = append(results, res)
		printResult(res)
		fmt.Println()
	}

	if len(results) > 1 {
		printSummary(results)
	}
}

func printResult(res benchmark.Result) {
	fmt.Printf("Elapsed    : %v\n", res.Elapsed.Round(time.Millisecond))
	fmt.Printf("Throughput : %.0f ops/sec\n", res.Throughput)
	fmt.Println()
	fmt.Println("Latency breakdown:")
	if !res.Read.Empty() {
		fmt.Printf("  Read  : %s\n", res.Read)
	}
	if !res.Update.Empty() {
		fmt.Printf("  Update: %s\n", res.Update)
	}
	if !res.Insert.Empty() {
		fmt.Printf("  Insert: %s\n", res.Insert)
	}
	if !res.RMW.Empty() {
		fmt.Printf("  RMW   : %s\n", res.RMW)
	}
}

func printSummary(results []benchmark.Result) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  %-12s  %12s  %10s  %10s\n", "Workload", "Throughput", "Avg Lat", "P99 Lat")
	fmt.Println(strings.Repeat("-", 60))
	for _, r := range results {
		combined := combinedAvg(r)
		combined99 := combinedP99(r)
		fmt.Printf("  %-12s  %9.0f ops/s  %10v  %10v\n",
			benchmark.Describe(r.Workload),
			r.Throughput,
			combined.Round(time.Microsecond),
			combined99.Round(time.Microsecond),
		)
	}
	fmt.Println(strings.Repeat("=", 60))
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
