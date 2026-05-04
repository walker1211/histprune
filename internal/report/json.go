package report

import "encoding/json"

type jsonSummary struct {
	Scanned    int            `json:"scanned"`
	Removed    int            `json:"removed"`
	Kept       int            `json:"kept"`
	DryRun     bool           `json:"dry_run"`
	BackupPath string         `json:"backup_path,omitempty"`
	RuleCounts map[string]int `json:"rule_counts"`
	Entries    []jsonRemoved  `json:"removed_entries"`
}

type jsonRemoved struct {
	LineNo    int          `json:"line_no"`
	Command   string       `json:"command"`
	Timestamp *int64       `json:"timestamp,omitempty"`
	Reasons   []jsonReason `json:"reasons"`
}

type jsonReason struct {
	Type   string `json:"type"`
	Detail string `json:"detail,omitempty"`
}

// JSON renders a machine-readable report.
func JSON(summary Summary) (string, error) {
	payload := jsonSummary{
		Scanned:    summary.Scanned,
		Removed:    summary.Removed,
		Kept:       summary.Kept,
		DryRun:     summary.DryRun,
		BackupPath: summary.BackupPath,
		RuleCounts: summary.RuleCounts,
		Entries:    make([]jsonRemoved, 0, len(summary.RemovedEntries)),
	}
	if payload.RuleCounts == nil {
		payload.RuleCounts = map[string]int{}
	}
	for _, entry := range summary.RemovedEntries {
		removed := jsonRemoved{
			LineNo:    entry.LineNo,
			Command:   entry.Command,
			Timestamp: entry.Timestamp,
			Reasons:   make([]jsonReason, 0, len(entry.Reasons)),
		}
		for _, reason := range entry.Reasons {
			removed.Reasons = append(removed.Reasons, jsonReason{Type: reason.Type, Detail: reason.Detail})
		}
		payload.Entries = append(payload.Entries, removed)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
