// Package benchmark provides a YCSB-inspired benchmark for KVS implementations.
//
// YCSB (Yahoo Cloud Serving Benchmark) is the de-facto standard for measuring
// key-value store performance. This package implements a simplified version with
// the five core workloads.
package benchmark

// Workload represents a YCSB workload preset.
type Workload string

const (
	// WorkloadA — Update heavy: 50% Read, 50% Update.
	// Models a session store application.
	WorkloadA Workload = "A"
	// WorkloadB — Read mostly: 95% Read, 5% Update.
	// Models a photo tagging application.
	WorkloadB Workload = "B"
	// WorkloadC — Read only: 100% Read.
	// Models user profile cache.
	WorkloadC Workload = "C"
	// WorkloadD — Read latest: 95% Read, 5% Insert.
	// New records are inserted and reads are biased toward recent data.
	WorkloadD Workload = "D"
	// WorkloadF — Read-modify-write: 50% Read, 50% RMW.
	// Each update is a read followed by a write to the same key.
	WorkloadF Workload = "F"
)

type opMix struct {
	read, update, insert, rmw float64
	desc                      string
}

var workloads = map[Workload]opMix{
	WorkloadA: {read: 0.50, update: 0.50, desc: "50% Read / 50% Update (update-heavy)"},
	WorkloadB: {read: 0.95, update: 0.05, desc: "95% Read / 5%  Update (read-mostly)"},
	WorkloadC: {read: 1.00, desc: "100% Read (read-only)"},
	WorkloadD: {read: 0.95, insert: 0.05, desc: "95% Read / 5%  Insert (read-latest)"},
	WorkloadF: {read: 0.50, rmw: 0.50, desc: "50% Read / 50% RMW (read-modify-write)"},
}

// Describe returns a human-readable description of the workload.
func Describe(w Workload) string {
	if m, ok := workloads[w]; ok {
		return string(w) + ": " + m.desc
	}
	return "unknown workload"
}

// AllWorkloads lists all available workloads in canonical order.
var AllWorkloads = []Workload{WorkloadA, WorkloadB, WorkloadC, WorkloadD, WorkloadF}

// Distribution controls how keys are selected during the run phase.
type Distribution string

const (
	Zipfian Distribution = "zipfian" // Hot-key skew: few keys get most traffic
	Uniform  Distribution = "uniform" // Every key equally likely
)
