package history

import (
	"strconv"
	"strings"
)

const zshExtendedPrefix = ": "

// ParseLine parses one zsh history line. Line numbers are 1-based when called
// from ParseContent, but callers may provide any source line number they track.
func ParseLine(lineNo int, line string) Entry {
	entry := Entry{
		Raw:    line,
		LineNo: lineNo,
	}

	if strings.HasPrefix(line, zshExtendedPrefix) {
		return parseExtendedLine(entry, line)
	}

	entry.Command = line
	entry.Format = FormatPlain
	return entry
}

func parseExtendedLine(entry Entry, line string) Entry {
	rest := strings.TrimPrefix(line, zshExtendedPrefix)
	firstColon := strings.IndexByte(rest, ':')
	semicolon := strings.IndexByte(rest, ';')
	if firstColon <= 0 || semicolon < 0 || semicolon < firstColon {
		return malformedEntry(entry)
	}

	timestamp, err := strconv.ParseInt(rest[:firstColon], 10, 64)
	if err != nil {
		return malformedEntry(entry)
	}
	duration, err := strconv.Atoi(rest[firstColon+1 : semicolon])
	if err != nil {
		return malformedEntry(entry)
	}

	entry.Command = rest[semicolon+1:]
	entry.Timestamp = &timestamp
	entry.Duration = &duration
	entry.Format = FormatZshExtended
	return entry
}

func malformedEntry(entry Entry) Entry {
	entry.Command = entry.Raw
	if semicolon := strings.IndexByte(entry.Raw, ';'); semicolon >= 0 {
		entry.Command = entry.Raw[semicolon+1:]
	}
	entry.Format = FormatMalformed
	return entry
}

// ParsedHistory is a full parsed history file, including file-level newline
// metadata that individual entries cannot represent.
type ParsedHistory struct {
	Entries         []Entry
	TrailingNewline bool
}

// ParseHistory parses full history file contents and records whether the
// original file ended with a newline.
func ParseHistory(content string) ParsedHistory {
	return ParsedHistory{
		Entries:         ParseContent(content),
		TrailingNewline: strings.HasSuffix(content, "\n"),
	}
}

// Serialize serializes the parsed entries using the original file's trailing
// newline convention.
func (h ParsedHistory) Serialize() string {
	return h.SerializeEntries(h.Entries)
}

// SerializeEntries serializes kept entries using the original file's trailing
// newline convention. This is useful after pruning a subset of h.Entries.
func (h ParsedHistory) SerializeEntries(entries []Entry) string {
	content := SerializeEntries(entries)
	if len(entries) > 0 && h.TrailingNewline {
		content += "\n"
	}
	return content
}

// ParseContent parses full history file contents into entries.
func ParseContent(content string) []Entry {
	if content == "" {
		return nil
	}

	trimmed := strings.TrimSuffix(content, "\n")
	lines := strings.Split(trimmed, "\n")
	entries := make([]Entry, 0, len(lines))
	for i, line := range lines {
		entries = append(entries, ParseLine(i+1, line))
	}
	return entries
}

// SerializeEntries serializes entries by joining line-level serialized values.
// File-level trailing newline semantics are handled by ParsedHistory.
func SerializeEntries(entries []Entry) string {
	if len(entries) == 0 {
		return ""
	}

	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, entry.Serialize())
	}
	return strings.Join(lines, "\n")
}
