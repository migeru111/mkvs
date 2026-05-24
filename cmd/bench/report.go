package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/migeru111/mkvs/benchmark"
)

func gitRef() string {
	id, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	ref := strings.TrimSpace(string(id))

	dirty, _ := exec.Command("git", "status", "--porcelain").Output()
	if len(strings.TrimSpace(string(dirty))) > 0 {
		ref += "-dirty"
	}
	return ref
}

func writeReport(cfg benchmark.Config, results []benchmark.Result) (string, error) {
	ref := gitRef() // ファイル生成前に1度だけ取得

	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", err
	}

	ts := time.Now().Format("20060102_150405")
	name := filepath.Join("reports", fmt.Sprintf("bench_%s_%s.txt", ts, ref))
	f, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := io.MultiWriter(os.Stdout, f)
	renderReport(w, cfg, ref, results)

	return name, nil
}

func renderReport(w io.Writer, cfg benchmark.Config, ref string, results []benchmark.Result) {
	sep := strings.Repeat("=", 60)
	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, "  mkvs Benchmark Report")
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  Date        : %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "  Commit      : %s\n", ref)
	fmt.Fprintf(w, "  Distribution: %s\n", cfg.Distribution)
	fmt.Fprintf(w, "  Records     : %d\n", cfg.RecordCount)
	fmt.Fprintf(w, "  Operations  : %d per workload\n", cfg.OperationCount)
	fmt.Fprintf(w, "  Value size  : %d bytes\n", cfg.ValueSize)
	fmt.Fprintf(w, "  Workers     : %d\n", cfg.Concurrency)
	fmt.Fprintln(w, sep)

	for _, res := range results {
		fmt.Fprintf(w, "\n--- Workload %s ---\n", benchmark.Describe(res.Workload))
		fmt.Fprintf(w, "Elapsed    : %v\n", res.Elapsed.Round(time.Millisecond))
		fmt.Fprintf(w, "Throughput : %.0f ops/sec\n", res.Throughput)
		fmt.Fprintln(w, "\nLatency breakdown:")
		if !res.Read.Empty() {
			fmt.Fprintf(w, "  Read  : %s\n", res.Read)
		}
		if !res.Update.Empty() {
			fmt.Fprintf(w, "  Update: %s\n", res.Update)
		}
		if !res.Insert.Empty() {
			fmt.Fprintf(w, "  Insert: %s\n", res.Insert)
		}
		if !res.RMW.Empty() {
			fmt.Fprintf(w, "  RMW   : %s\n", res.RMW)
		}
	}

	if len(results) > 1 {
		fmt.Fprintln(w, "\n"+sep)
		fmt.Fprintln(w, "  Summary")
		fmt.Fprintln(w, sep)
		fmt.Fprintf(w, "  %-12s  %12s  %10s  %10s\n", "Workload", "Throughput", "Avg Lat", "P99 Lat")
		fmt.Fprintln(w, strings.Repeat("-", 60))
		for _, r := range results {
			fmt.Fprintf(w, "  %-12s  %9.0f ops/s  %10v  %10v\n",
				benchmark.Describe(r.Workload),
				r.Throughput,
				fmtDuration(combinedAvg(r)),
				fmtDuration(combinedP99(r)),
			)
		}
		fmt.Fprintln(w, sep)
	}
}
