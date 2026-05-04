package history

import "strconv"

// Format describes the recognized history line format.
type Format int

const (
	FormatPlain Format = iota
	FormatZshExtended
	FormatMalformed
)

// Entry is a parsed history entry with enough metadata for later filtering.
type Entry struct {
	Raw       string
	Command   string
	Timestamp *int64
	Duration  *int
	Format    Format
	LineNo    int
}

// Serialize returns the original text whenever it is available, preserving
// history files byte-for-byte for unchanged entries.
func (e Entry) Serialize() string {
	if e.Raw != "" {
		return e.Raw
	}

	switch e.Format {
	case FormatZshExtended:
		if e.Timestamp != nil && e.Duration != nil {
			return ": " + strconv.FormatInt(*e.Timestamp, 10) + ":" + strconv.Itoa(*e.Duration) + ";" + e.Command
		}
		return e.Command
	case FormatPlain, FormatMalformed:
		return e.Command
	default:
		return e.Command
	}
}
