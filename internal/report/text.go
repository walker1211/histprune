package report

import (
	"io"
	"sort"
	"strings"

	"github.com/walker1211/histprune/internal/prune"
	"github.com/walker1211/histprune/internal/render"
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
	_ = WriteText(&b, summary)
	return b.String()
}

func WriteText(w io.Writer, summary Summary) error {
	content := render.NewContent(w)
	content.Writef("Scanned: %d\n", summary.Scanned)
	content.Writef("Removed: %d\n", summary.Removed)
	content.Writef("Kept: %d\n", summary.Kept)
	if summary.DryRun {
		content.WriteString("Mode: dry-run\n")
	} else {
		content.WriteString("Mode: write\n")
	}
	if summary.BackupPath != "" {
		content.Writef("Backup: %s\n", summary.BackupPath)
	}
	if len(summary.RuleCounts) > 0 {
		content.WriteString("Rules:\n")
		keys := make([]string, 0, len(summary.RuleCounts))
		for key := range summary.RuleCounts {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			content.Writef("  %s: %d\n", key, summary.RuleCounts[key])
		}
	}
	if len(summary.RemovedEntries) > 0 {
		content.WriteString("Removed entries:\n")
		for _, entry := range summary.RemovedEntries {
			content.Writef("  Line %d: %s\n", entry.LineNo, entry.Command)
			for _, reason := range entry.Reasons {
				if reason.Detail == "" {
					content.Writef("    - %s\n", reason.Type)
					continue
				}
				content.Writef("    - %s: %s\n", reason.Type, reason.Detail)
			}
		}
	}
	return content.Err()
}
