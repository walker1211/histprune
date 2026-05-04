package prune

import (
	"strings"
	"testing"

	"github.com/walker1211/histprune/internal/history"
)

func TestDecideContainsAndRegexAccumulateReasons(t *testing.T) {
	engine, err := BuildRules(Options{
		Contains: []string{"secret"},
		Regex:    []string{`token=\w+`},
	})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	entries := []history.Entry{{Command: "deploy --secret token=abc123", LineNo: 1}}
	decisions := engine.Decide(entries)

	if len(decisions) != 1 {
		t.Fatalf("got %d decisions, want 1", len(decisions))
	}
	if !decisions[0].Remove {
		t.Fatal("decision should remove matching entry")
	}
	assertReason(t, decisions[0], "contains", "secret")
	assertReason(t, decisions[0], "regex", `token=\w+`)
}

func TestBuildRulesInvalidRegexReturnsError(t *testing.T) {
	_, err := BuildRules(Options{Regex: []string{"[unterminated"}})
	if err == nil {
		t.Fatal("BuildRules should return error for invalid regex")
	}
}

func TestBuildRulesEmptyContainsReturnsError(t *testing.T) {
	_, err := BuildRules(Options{Contains: []string{""}})
	if err == nil {
		t.Fatal("BuildRules should return error for empty contains literal")
	}
	if !strings.Contains(err.Error(), "contains") {
		t.Fatalf("error should mention contains, got %v", err)
	}
}

func TestBuildRulesEmptyRegexReturnsError(t *testing.T) {
	_, err := BuildRules(Options{Regex: []string{""}})
	if err == nil {
		t.Fatal("BuildRules should return error for empty regex pattern")
	}
	if !strings.Contains(err.Error(), "regex") {
		t.Fatalf("error should mention regex, got %v", err)
	}
}

func TestBuildRulesNonPositiveLinesReturnErrors(t *testing.T) {
	for _, line := range []int{0, -1} {
		_, err := BuildRules(Options{Lines: []int{line}})
		if err == nil {
			t.Fatalf("BuildRules should return error for line %d", line)
		}
		if !strings.Contains(err.Error(), "line") {
			t.Fatalf("error should mention line, got %v", err)
		}
	}
}

func TestDecideLineMatchingRemovesRequestedLine(t *testing.T) {
	engine, err := BuildRules(Options{Lines: []int{2}})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	decisions := engine.Decide([]history.Entry{
		{Command: "keep", LineNo: 1},
		{Command: "remove", LineNo: 2},
	})

	if decisions[0].Remove {
		t.Fatal("line 1 should be kept")
	}
	if !decisions[1].Remove {
		t.Fatal("line 2 should be removed")
	}
	assertReason(t, decisions[1], "line", "2")
}

func TestDecideBeforeRemovesOnlyTimestampedEntriesBeforeDate(t *testing.T) {
	jan1 := int64(1704067200) // 2024-01-01T00:00:00Z
	jan2 := int64(1704153600) // 2024-01-02T00:00:00Z

	engine, err := BuildRules(Options{Before: "2024-01-02"})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	decisions := engine.Decide([]history.Entry{
		{Command: "old", Timestamp: &jan1, LineNo: 1},
		{Command: "boundary", Timestamp: &jan2, LineNo: 2},
		{Command: "plain", LineNo: 3},
	})

	if !decisions[0].Remove {
		t.Fatal("timestamp before 2024-01-02 should be removed")
	}
	assertReason(t, decisions[0], "before", "2024-01-02")
	if decisions[1].Remove {
		t.Fatal("timestamp at start of 2024-01-02 should be kept")
	}
	if decisions[2].Remove {
		t.Fatal("non-timestamped entry should not match before rule")
	}
}

func TestDecideBetweenRemovesTimestampedEntriesInclusively(t *testing.T) {
	dec31 := int64(1703980800)    // 2023-12-31T00:00:00Z
	jan1Noon := int64(1704110400) // 2024-01-01T12:00:00Z
	jan3End := int64(1704326399)  // 2024-01-03T23:59:59Z
	jan4 := int64(1704326400)     // 2024-01-04T00:00:00Z

	engine, err := BuildRules(Options{Between: DateRange{Start: "2024-01-01", End: "2024-01-03"}})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	decisions := engine.Decide([]history.Entry{
		{Command: "before", Timestamp: &dec31, LineNo: 1},
		{Command: "start", Timestamp: &jan1Noon, LineNo: 2},
		{Command: "end", Timestamp: &jan3End, LineNo: 3},
		{Command: "after", Timestamp: &jan4, LineNo: 4},
		{Command: "plain", LineNo: 5},
	})

	if decisions[0].Remove {
		t.Fatal("entry before range should be kept")
	}
	if !decisions[1].Remove {
		t.Fatal("entry on start date should be removed")
	}
	assertReason(t, decisions[1], "between", "2024-01-01")
	if !decisions[2].Remove {
		t.Fatal("entry on end date should be removed")
	}
	if decisions[3].Remove {
		t.Fatal("entry after range should be kept")
	}
	if decisions[4].Remove {
		t.Fatal("non-timestamped entry should not match between rule")
	}
}

func TestDecideDedupeKeepsLastOccurrence(t *testing.T) {
	engine, err := BuildRules(Options{Dedupe: true})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	decisions := engine.Decide([]history.Entry{
		{Command: "go test", LineNo: 1},
		{Command: "ls", LineNo: 2},
		{Command: "go test", LineNo: 3},
	})

	if !decisions[0].Remove {
		t.Fatal("earlier duplicate should be removed")
	}
	assertReason(t, decisions[0], "dedupe", "go test")
	if decisions[1].Remove {
		t.Fatal("unique command should be kept")
	}
	if decisions[2].Remove {
		t.Fatal("last duplicate occurrence should be kept")
	}
}

func TestDecideMalformedEntryWithCommandCanBeRemovedByContains(t *testing.T) {
	engine, err := BuildRules(Options{Contains: []string{"rm -rf"}})
	if err != nil {
		t.Fatalf("BuildRules returned error: %v", err)
	}

	decisions := engine.Decide([]history.Entry{
		{Raw: ": malformed;rm -rf /tmp/cache", Command: "rm -rf /tmp/cache", Format: history.FormatMalformed, LineNo: 1},
	})

	if !decisions[0].Remove {
		t.Fatal("malformed entry command should be matched by contains")
	}
	assertReason(t, decisions[0], "contains", "rm -rf")
}

func assertReason(t *testing.T, decision Decision, wantType, wantDetailPart string) {
	t.Helper()
	for _, reason := range decision.Reasons {
		if reason.Type == wantType && strings.Contains(reason.Detail, wantDetailPart) {
			return
		}
	}
	t.Fatalf("missing reason type %q detail containing %q in %#v", wantType, wantDetailPart, decision.Reasons)
}
