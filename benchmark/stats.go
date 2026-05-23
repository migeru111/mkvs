package benchmark

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Recorder collects latency samples from multiple goroutines concurrently.
type Recorder struct {
	mu   sync.Mutex
	data []int64 // nanoseconds
}

// NewRecorder pre-allocates a Recorder with the given capacity hint.
func NewRecorder(cap int) *Recorder {
	return &Recorder{data: make([]int64, 0, cap)}
}

// Add records a single operation latency.
func (rec *Recorder) Add(d time.Duration) {
	ns := d.Nanoseconds()
	rec.mu.Lock()
	rec.data = append(rec.data, ns)
	rec.mu.Unlock()
}

// Stats computes latency statistics over all recorded samples.
// This sorts a copy of the data, so it can be called after the benchmark run.
func (rec *Recorder) Stats() Stats {
	rec.mu.Lock()
	data := make([]int64, len(rec.data))
	copy(data, rec.data)
	rec.mu.Unlock()

	n := len(data)
	if n == 0 {
		return Stats{}
	}
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })

	var total int64
	for _, v := range data {
		total += v
	}

	pct := func(p float64) time.Duration {
		idx := int(float64(n-1) * p)
		return time.Duration(data[idx])
	}

	return Stats{
		Count: n,
		Avg:   time.Duration(total / int64(n)),
		P50:   pct(0.50),
		P95:   pct(0.95),
		P99:   pct(0.99),
		P999:  pct(0.999),
		Max:   time.Duration(data[n-1]),
	}
}

// Stats holds the computed latency percentiles for one operation type.
type Stats struct {
	Count int
	Avg   time.Duration
	P50   time.Duration
	P95   time.Duration
	P99   time.Duration
	P999  time.Duration
	Max   time.Duration
}

// Empty reports whether there are no samples.
func (s Stats) Empty() bool { return s.Count == 0 }

func (s Stats) String() string {
	if s.Count == 0 {
		return "n/a"
	}
	return fmt.Sprintf(
		"n=%-8d  avg=%-10v  p50=%-10v  p95=%-10v  p99=%-10v  p999=%-10v  max=%v",
		s.Count,
		round(s.Avg), round(s.P50), round(s.P95), round(s.P99), round(s.P999), round(s.Max),
	)
}

func round(d time.Duration) time.Duration {
	if d < time.Microsecond {
		return d
	}
	return d.Round(time.Microsecond)
}
