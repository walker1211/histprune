package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/walker1211/histprune/internal/history"
	"github.com/walker1211/histprune/internal/prune"
)

func TestFromDecisionsBuildsCountsAndRemovedEntries(t *testing.T) {
	ts := int64(1710000000)
	decisions := []prune.Decision{
		{Entry: history.Entry{LineNo: 1, Command: "keep"}},
		{Entry: history.Entry{LineNo: 2, Command: "git push token", Timestamp: &ts}, Remove: true, Reasons: []prune.Reason{
			{Type: prune.ReasonTypeContains, Detail: "command contains token"},
			{Type: prune.ReasonTypeRegex, Detail: "command matches regex token"},
		}},
		{Entry: history.Entry{LineNo: 3, Command: "duplicate"}, Remove: true, Reasons: []prune.Reason{
			{Type: prune.ReasonTypeDedupe, Detail: "duplicate command"},
		}},
	}

	summary := FromDecisions(decisions, true, "/tmp/history.backup")

	if summary.Scanned != 3 || summary.Removed != 2 || summary.Kept != 1 {
		t.Fatalf("unexpected totals: %#v", summary)
	}
	if !summary.DryRun || summary.BackupPath != "/tmp/history.backup" {
		t.Fatalf("unexpected execution metadata: %#v", summary)
	}
	if got := summary.RuleCounts[prune.ReasonTypeContains]; got != 1 {
		t.Fatalf("contains count = %d, want 1", got)
	}
	if got := summary.RuleCounts[prune.ReasonTypeRegex]; got != 1 {
		t.Fatalf("regex count = %d, want 1", got)
	}
	if got := summary.RuleCounts[prune.ReasonTypeDedupe]; got != 1 {
		t.Fatalf("dedupe count = %d, want 1", got)
	}
	if len(summary.RemovedEntries) != 2 {
		t.Fatalf("removed entries = %d, want 2", len(summary.RemovedEntries))
	}
	if summary.RemovedEntries[0].Timestamp == nil || *summary.RemovedEntries[0].Timestamp != ts {
		t.Fatalf("timestamp was not copied: %#v", summary.RemovedEntries[0])
	}
	if len(summary.RemovedEntries[0].Reasons) != 2 {
		t.Fatalf("multiple reasons were not preserved: %#v", summary.RemovedEntries[0])
	}
}

func TestTextIncludesSummaryModeBackupAndSortedRuleCounts(t *testing.T) {
	summary := Summary{
		Scanned:    4,
		Removed:    2,
		Kept:       2,
		DryRun:     true,
		BackupPath: "/tmp/.zsh_history.histprune-backup-20260503T120000",
		RuleCounts: map[string]int{
			prune.ReasonTypeRegex:    1,
			prune.ReasonTypeContains: 2,
		},
	}

	text := Text(summary)

	for _, want := range []string{
		"Scanned: 4",
		"Removed: 2",
		"Kept: 2",
		"Mode: dry-run",
		"Backup: /tmp/.zsh_history.histprune-backup-20260503T120000",
		"contains: 2",
		"regex: 1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text report missing %q in:\n%s", want, text)
		}
	}
	if strings.Index(text, "contains: 2") > strings.Index(text, "regex: 1") {
		t.Fatalf("rule counts are not sorted by rule name:\n%s", text)
	}
}

func TestTextIncludesRemovedEntryDetails(t *testing.T) {
	summary := Summary{
		Scanned: 2,
		Removed: 1,
		Kept:    1,
		RemovedEntries: []RemovedEntry{{
			LineNo:  7,
			Command: "curl secret",
			Reasons: []prune.Reason{
				{Type: prune.ReasonTypeContains, Detail: "command contains secret"},
				{Type: prune.ReasonTypeRegex, Detail: "command matches regex secret"},
			},
		}},
	}

	text := Text(summary)

	for _, want := range []string{
		"Removed entries:",
		"Line 7: curl secret",
		"contains: command contains secret",
		"regex: command matches regex secret",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text report missing %q in:\n%s", want, text)
		}
	}
}

func TestJSONIncludesTotalsAndRemovedEntries(t *testing.T) {
	ts := int64(1710000000)
	summary := Summary{
		Scanned: 2,
		Removed: 1,
		Kept:    1,
		DryRun:  false,
		RuleCounts: map[string]int{
			prune.ReasonTypeContains: 1,
			prune.ReasonTypeRegex:    1,
		},
		RemovedEntries: []RemovedEntry{{
			LineNo:    7,
			Command:   "curl secret",
			Timestamp: &ts,
			Reasons: []prune.Reason{
				{Type: prune.ReasonTypeContains, Detail: "command contains secret"},
				{Type: prune.ReasonTypeRegex, Detail: "command matches regex secret"},
			},
		}},
	}

	payload, err := JSON(summary)
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var decoded struct {
		Scanned int            `json:"scanned"`
		Removed int            `json:"removed"`
		Kept    int            `json:"kept"`
		DryRun  bool           `json:"dry_run"`
		Rules   map[string]int `json:"rule_counts"`
		Entries []struct {
			LineNo    int    `json:"line_no"`
			Command   string `json:"command"`
			Timestamp *int64 `json:"timestamp,omitempty"`
			Reasons   []struct {
				Type   string `json:"type"`
				Detail string `json:"detail,omitempty"`
			} `json:"reasons"`
		} `json:"removed_entries"`
	}
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, payload)
	}
	if decoded.Scanned != 2 || decoded.Removed != 1 || decoded.Kept != 1 || decoded.DryRun {
		t.Fatalf("unexpected decoded summary: %#v", decoded)
	}
	if decoded.Rules[prune.ReasonTypeContains] != 1 || decoded.Rules[prune.ReasonTypeRegex] != 1 {
		t.Fatalf("unexpected rule counts: %#v", decoded.Rules)
	}
	if len(decoded.Entries) != 1 || decoded.Entries[0].LineNo != 7 || decoded.Entries[0].Command != "curl secret" {
		t.Fatalf("unexpected removed entries: %#v", decoded.Entries)
	}
	if decoded.Entries[0].Timestamp == nil || *decoded.Entries[0].Timestamp != ts {
		t.Fatalf("missing timestamp: %#v", decoded.Entries[0])
	}
	if len(decoded.Entries[0].Reasons) != 2 {
		t.Fatalf("multiple reasons missing: %#v", decoded.Entries[0].Reasons)
	}
}
