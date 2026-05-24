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
	// git describe --tags: タグがあれば "hashmap-mutex"、タグより後なら "hashmap-mutex-3-gabcdef" を返す
	out, err := exec.Command("git", "describe", "--tags", "--always").Output()
	if err != nil {
		return "unknown"
	}
	ref := strings.TrimSpace(string(out))

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
	scaling := isScalingRun(results)
	if scaling {
		workers := uniqueWorkers(results)
		fmt.Fprintf(w, "  Workers     : scaling %v\n", workers)
	} else {
		fmt.Fprintf(w, "  Workers     : %d\n", cfg.Concurrency)
	}
	fmt.Fprintln(w, sep)

	for _, res := range results {
		fmt.Fprintf(w, "\n--- [%s] Workload %s  workers=%d ---\n", res.StoreName, benchmark.Describe(res.Workload), res.Workers)
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

	if scaling {
		renderScalingTable(w, results)
	} else if len(results) > 1 {
		renderSummaryTable(w, results)
	}
}

// uniqueWorkers は results に含まれる Workers 値を重複なく昇順で返す。
func uniqueWorkers(results []benchmark.Result) []int {
	seen := map[int]bool{}
	var out []int
	for _, r := range results {
		if !seen[r.Workers] {
			seen[r.Workers] = true
			out = append(out, r.Workers)
		}
	}
	return out
}

// isScalingRun は results に複数の異なる Workers 値が含まれているか判定する。
func isScalingRun(results []benchmark.Result) bool {
	if len(results) == 0 {
		return false
	}
	first := results[0].Workers
	for _, r := range results[1:] {
		if r.Workers != first {
			return true
		}
	}
	return false
}

func renderScalingTable(w io.Writer, results []benchmark.Result) {
	sep := strings.Repeat("=", 70)

	// ワークロード種別を順序を保って収集
	seen := map[benchmark.Workload]bool{}
	var wls []benchmark.Workload
	for _, r := range results {
		if !seen[r.Workload] {
			seen[r.Workload] = true
			wls = append(wls, r.Workload)
		}
	}

	// ワークロードごとにスケーリング表を出力
	for _, wl := range wls {
		fmt.Fprintln(w, "\n"+sep)
		fmt.Fprintf(w, "  Scaling: %s\n", benchmark.Describe(wl))
		fmt.Fprintln(w, sep)
		fmt.Fprintf(w, "  %-18s  %7s  %14s  %10s  %10s\n", "Store", "Workers", "Throughput", "Avg Lat", "P99 Lat")
		fmt.Fprintln(w, strings.Repeat("-", 68))
		for _, r := range results {
			if r.Workload != wl {
				continue
			}
			fmt.Fprintf(w, "  %-18s  %7d  %11.0f ops/s  %10s  %10s\n",
				r.StoreName,
				r.Workers,
				r.Throughput,
				fmtDuration(combinedAvg(r)),
				fmtDuration(combinedP99(r)),
			)
		}
	}
	fmt.Fprintln(w, sep)
}

func renderSummaryTable(w io.Writer, results []benchmark.Result) {
	sep := strings.Repeat("=", 75)
	fmt.Fprintln(w, "\n"+sep)
	fmt.Fprintln(w, "  Summary")
	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "  %-18s  %-12s  %12s  %10s  %10s\n", "Store", "Workload", "Throughput", "Avg Lat", "P99 Lat")
	fmt.Fprintln(w, strings.Repeat("-", 75))
	for _, r := range results {
		fmt.Fprintf(w, "  %-18s  %-12s  %9.0f ops/s  %10v  %10v\n",
			r.StoreName,
			benchmark.Describe(r.Workload),
			r.Throughput,
			fmtDuration(combinedAvg(r)),
			fmtDuration(combinedP99(r)),
		)
	}
	fmt.Fprintln(w, sep)
}
