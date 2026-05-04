package prune

import "github.com/walker1211/histprune/internal/history"

// Reason explains why a history entry should be removed.
type Reason struct {
	Type   string
	Detail string
}

// Decision records the pruning decision for one history entry.
type Decision struct {
	Entry   history.Entry
	Remove  bool
	Reasons []Reason
}

// DateRange configures an inclusive YYYY-MM-DD date range.
type DateRange struct {
	Start string
	End   string
}

// Options configures the prune rules used by an Engine.
type Options struct {
	Contains []string
	Regex    []string
	Lines    []int
	Before   string
	Between  DateRange
	Dedupe   bool
}
