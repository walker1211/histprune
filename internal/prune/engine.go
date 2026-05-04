package prune

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/walker1211/histprune/internal/history"
)

const (
	dateLayout = "2006-01-02"

	ReasonTypeContains = "contains"
	ReasonTypeRegex    = "regex"
	ReasonTypeLine     = "line"
	ReasonTypeBefore   = "before"
	ReasonTypeBetween  = "between"
	ReasonTypeDedupe   = "dedupe"
)

// Engine applies compiled prune rules to history entries.
type Engine struct {
	contains []string
	regex    []compiledRegex
	lines    map[int]struct{}
	before   *dateRule
	between  *betweenRule
	dedupe   bool
}

type compiledRegex struct {
	pattern string
	re      *regexp.Regexp
}

type dateRule struct {
	text  string
	start time.Time
}

type betweenRule struct {
	startText string
	endText   string
	start     time.Time
	end       time.Time
}

// BuildRules validates options and returns an Engine ready to decide entries.
func BuildRules(options Options) (*Engine, error) {
	engine := &Engine{
		contains: make([]string, 0, len(options.Contains)),
		lines:    make(map[int]struct{}, len(options.Lines)),
		dedupe:   options.Dedupe,
	}

	for _, literal := range options.Contains {
		if literal == "" {
			return nil, fmt.Errorf("contains literal must not be empty")
		}
		engine.contains = append(engine.contains, literal)
	}

	for _, pattern := range options.Regex {
		if pattern == "" {
			return nil, fmt.Errorf("regex pattern must not be empty")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compile regex %q: %w", pattern, err)
		}
		engine.regex = append(engine.regex, compiledRegex{pattern: pattern, re: re})
	}

	for _, line := range options.Lines {
		if line <= 0 {
			return nil, fmt.Errorf("line number must be positive: %d", line)
		}
		engine.lines[line] = struct{}{}
	}

	if options.Before != "" {
		start, err := parseUTCDate(options.Before)
		if err != nil {
			return nil, fmt.Errorf("parse before date %q: %w", options.Before, err)
		}
		engine.before = &dateRule{text: options.Before, start: start}
	}

	if options.Between.Start != "" || options.Between.End != "" {
		if options.Between.Start == "" || options.Between.End == "" {
			return nil, fmt.Errorf("between requires both start and end dates")
		}
		start, err := parseUTCDate(options.Between.Start)
		if err != nil {
			return nil, fmt.Errorf("parse between start date %q: %w", options.Between.Start, err)
		}
		end, err := parseUTCDate(options.Between.End)
		if err != nil {
			return nil, fmt.Errorf("parse between end date %q: %w", options.Between.End, err)
		}
		if end.Before(start) {
			return nil, fmt.Errorf("between end date %q is before start date %q", options.Between.End, options.Between.Start)
		}
		engine.between = &betweenRule{startText: options.Between.Start, endText: options.Between.End, start: start, end: end}
	}

	return engine, nil
}

// Decide returns one explainable decision per input entry.
func (e *Engine) Decide(entries []history.Entry) []Decision {
	decisions := make([]Decision, 0, len(entries))
	lastByCommand := map[string]int(nil)
	if e.dedupe {
		lastByCommand = make(map[string]int, len(entries))
		for i, entry := range entries {
			lastByCommand[entry.Command] = i
		}
	}

	for i, entry := range entries {
		reasons := e.reasonsFor(entry)
		if e.dedupe && lastByCommand[entry.Command] != i {
			reasons = append(reasons, Reason{Type: ReasonTypeDedupe, Detail: fmt.Sprintf("duplicate command %q; keeping last occurrence", entry.Command)})
		}

		decisions = append(decisions, Decision{
			Entry:   entry,
			Remove:  len(reasons) > 0,
			Reasons: reasons,
		})
	}

	return decisions
}

func (e *Engine) reasonsFor(entry history.Entry) []Reason {
	var reasons []Reason
	for _, literal := range e.contains {
		if strings.Contains(entry.Command, literal) {
			reasons = append(reasons, Reason{Type: ReasonTypeContains, Detail: "command contains " + literal})
		}
	}

	for _, candidate := range e.regex {
		if candidate.re.MatchString(entry.Command) {
			reasons = append(reasons, Reason{Type: ReasonTypeRegex, Detail: "command matches regex " + candidate.pattern})
		}
	}

	if _, ok := e.lines[entry.LineNo]; ok {
		reasons = append(reasons, Reason{Type: ReasonTypeLine, Detail: fmt.Sprintf("line %d matched", entry.LineNo)})
	}

	if entry.Timestamp != nil {
		entryTime := time.Unix(*entry.Timestamp, 0).UTC()
		if e.before != nil && entryTime.Before(e.before.start) {
			reasons = append(reasons, Reason{Type: ReasonTypeBefore, Detail: fmt.Sprintf("timestamp is before %s UTC", e.before.text)})
		}
		if e.between != nil && isDateWithinInclusiveRange(entryTime, e.between.start, e.between.end) {
			reasons = append(reasons, Reason{Type: ReasonTypeBetween, Detail: fmt.Sprintf("timestamp date is between %s and %s UTC inclusive", e.between.startText, e.between.endText)})
		}
	}

	return reasons
}

func parseUTCDate(text string) (time.Time, error) {
	parsed, err := time.ParseInLocation(dateLayout, text, time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.Format(dateLayout) != text {
		return time.Time{}, fmt.Errorf("invalid date %q", text)
	}
	return parsed, nil
}

func isDateWithinInclusiveRange(entryTime, start, end time.Time) bool {
	entryDate := time.Date(entryTime.UTC().Year(), entryTime.UTC().Month(), entryTime.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return !entryDate.Before(start) && !entryDate.After(end)
}

// FilterRemoved returns entries that are not marked for removal by decisions.
func FilterRemoved(decisions []Decision) []history.Entry {
	kept := make([]history.Entry, 0, len(decisions))
	for _, decision := range decisions {
		if !decision.Remove {
			kept = append(kept, decision.Entry)
		}
	}
	return kept
}

// LineSet returns a copy of the configured line rule set, primarily for callers
// that need to inspect normalized options.
func (e *Engine) LineSet() map[int]struct{} {
	lines := make(map[int]struct{}, len(e.lines))
	for line := range e.lines {
		lines[line] = struct{}{}
	}
	return lines
}

// String renders a reason in a compact human-readable form.
func (r Reason) String() string {
	if r.Detail == "" {
		return r.Type
	}
	return r.Type + ": " + r.Detail
}
