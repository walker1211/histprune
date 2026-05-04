package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/walker1211/histprune/internal/prune"
)

// Summary is the report model shared by text, JSON, and app-level integration.
type Summary struct {
	Scanned        int
	Removed        int
	Kept           int
	DryRun         bool
	BackupPath     string
	RuleCounts     map[string]int
	RemovedEntries []RemovedEntry
}

// RemovedEntry is the report-safe representation of a removed history entry.
type RemovedEntry struct {
	LineNo    int
	Command   string
	Timestamp *int64
	Reasons   []prune.Reason
}

// FromDecisions builds a report summary from pruning decisions.
func FromDecisions(decisions []prune.Decision, dryRun bool, backupPath string) Summary {
	summary := Summary{
		Scanned:    len(decisions),
		DryRun:     dryRun,
		BackupPath: backupPath,
		RuleCounts: make(map[string]int),
	}

	for _, decision := range decisions {
		if !decision.Remove {
			summary.Kept++
			continue
		}

		summary.Removed++
		for _, reason := range decision.Reasons {
			summary.RuleCounts[reason.Type]++
		}
		summary.RemovedEntries = append(summary.RemovedEntries, RemovedEntry{
			LineNo:    decision.Entry.LineNo,
			Command:   decision.Entry.Command,
			Timestamp: decision.Entry.Timestamp,
			Reasons:   append([]prune.Reason(nil), decision.Reasons...),
		})
	}

	return summary
}

// Text renders a compact human-readable report.
func Text(summary Summary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Scanned: %d\n", summary.Scanned)
	fmt.Fprintf(&b, "Removed: %d\n", summary.Removed)
	fmt.Fprintf(&b, "Kept: %d\n", summary.Kept)
	if summary.DryRun {
		b.WriteString("Mode: dry-run\n")
	} else {
		b.WriteString("Mode: write\n")
	}
	if summary.BackupPath != "" {
		fmt.Fprintf(&b, "Backup: %s\n", summary.BackupPath)
	}
	if len(summary.RuleCounts) > 0 {
		b.WriteString("Rules:\n")
		keys := make([]string, 0, len(summary.RuleCounts))
		for key := range summary.RuleCounts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&b, "  %s: %d\n", key, summary.RuleCounts[key])
		}
	}
	if len(summary.RemovedEntries) > 0 {
		b.WriteString("Removed entries:\n")
		for _, entry := range summary.RemovedEntries {
			fmt.Fprintf(&b, "  Line %d: %s\n", entry.LineNo, entry.Command)
			for _, reason := range entry.Reasons {
				if reason.Detail == "" {
					fmt.Fprintf(&b, "    - %s\n", reason.Type)
					continue
				}
				fmt.Fprintf(&b, "    - %s: %s\n", reason.Type, reason.Detail)
			}
		}
	}
	return b.String()
}
