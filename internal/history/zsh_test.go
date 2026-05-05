package history

import (
	"reflect"
	"testing"
)

func TestParseLineExtendedHistory(t *testing.T) {
	entry := ParseLine(7, ": 1714752000:0;git status")

	wantTimestamp := int64(1714752000)
	wantDuration := 0
	want := Entry{
		Raw:       ": 1714752000:0;git status",
		Command:   "git status",
		Timestamp: &wantTimestamp,
		Duration:  &wantDuration,
		Format:    FormatZshExtended,
		LineNo:    7,
	}
	if !reflect.DeepEqual(entry, want) {
		t.Fatalf("ParseLine() = %#v, want %#v", entry, want)
	}
	if got := entry.Serialize(); got != want.Raw {
		t.Fatalf("Serialize() = %q, want %q", got, want.Raw)
	}
}

func TestParseLinePlainHistory(t *testing.T) {
	entry := ParseLine(3, "git status")

	want := Entry{
		Raw:     "git status",
		Command: "git status",
		Format:  FormatPlain,
		LineNo:  3,
	}
	if !reflect.DeepEqual(entry, want) {
		t.Fatalf("ParseLine() = %#v, want %#v", entry, want)
	}
	if got := entry.Serialize(); got != want.Raw {
		t.Fatalf("Serialize() = %q, want %q", got, want.Raw)
	}
}

func TestParseLineUnmetafiesZshCommandBytes(t *testing.T) {
	metafiedCommand := "echo 中" + string([]byte{0xe6, zshMetaByte, 0xb6, zshMetaByte, 0xa7})
	timestamp := int64(1714752000)
	duration := 0
	tests := []struct {
		name string
		line string
		want Entry
	}{
		{
			name: "extended",
			line: ": 1714752000:0;" + metafiedCommand,
			want: Entry{
				Raw:       ": 1714752000:0;" + metafiedCommand,
				Command:   "echo 中文",
				Timestamp: &timestamp,
				Duration:  &duration,
				Format:    FormatZshExtended,
				LineNo:    4,
			},
		},
		{
			name: "plain",
			line: metafiedCommand,
			want: Entry{
				Raw:     metafiedCommand,
				Command: "echo 中文",
				Format:  FormatPlain,
				LineNo:  4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := ParseLine(4, tt.line)
			if !reflect.DeepEqual(entry, tt.want) {
				t.Fatalf("ParseLine() = %#v, want %#v", entry, tt.want)
			}
			if got := entry.Serialize(); got != tt.line {
				t.Fatalf("Serialize() = %q, want %q", got, tt.line)
			}
		})
	}
}

func TestParseLineKeepsValidUTF8ContainingMetaByte(t *testing.T) {
	entry := ParseLine(5, "echo ăx")

	want := Entry{
		Raw:     "echo ăx",
		Command: "echo ăx",
		Format:  FormatPlain,
		LineNo:  5,
	}
	if !reflect.DeepEqual(entry, want) {
		t.Fatalf("ParseLine() = %#v, want %#v", entry, want)
	}
	if got := entry.Serialize(); got != want.Raw {
		t.Fatalf("Serialize() = %q, want %q", got, want.Raw)
	}
}

func TestParseLineMalformedZshLookingHistoryWithSemicolon(t *testing.T) {
	entry := ParseLine(11, ": not-a-timestamp:0;aws_secret_access_key=secret")

	want := Entry{
		Raw:     ": not-a-timestamp:0;aws_secret_access_key=secret",
		Command: "aws_secret_access_key=secret",
		Format:  FormatMalformed,
		LineNo:  11,
	}
	if !reflect.DeepEqual(entry, want) {
		t.Fatalf("ParseLine() = %#v, want %#v", entry, want)
	}
	if got := entry.Serialize(); got != want.Raw {
		t.Fatalf("Serialize() = %q, want %q", got, want.Raw)
	}
}

func TestParseLineMalformedZshLookingHistoryWithoutSemicolon(t *testing.T) {
	entry := ParseLine(12, ": not-a-timestamp:0 aws_secret_access_key=secret")

	want := Entry{
		Raw:     ": not-a-timestamp:0 aws_secret_access_key=secret",
		Command: ": not-a-timestamp:0 aws_secret_access_key=secret",
		Format:  FormatMalformed,
		LineNo:  12,
	}
	if !reflect.DeepEqual(entry, want) {
		t.Fatalf("ParseLine() = %#v, want %#v", entry, want)
	}
	if got := entry.Serialize(); got != want.Raw {
		t.Fatalf("Serialize() = %q, want %q", got, want.Raw)
	}
}

func TestParseContentAndSerializeEntries(t *testing.T) {
	content := ": 1714752000:0;git status\nplain command\n: nope:0;kept raw"

	entries := ParseContent(content)

	if len(entries) != 3 {
		t.Fatalf("ParseContent() returned %d entries, want 3", len(entries))
	}
	if got, want := entries[0].LineNo, 1; got != want {
		t.Fatalf("entries[0].LineNo = %d, want %d", got, want)
	}
	if got, want := entries[1].Command, "plain command"; got != want {
		t.Fatalf("entries[1].Command = %q, want %q", got, want)
	}
	if got, want := entries[2].Format, FormatMalformed; got != want {
		t.Fatalf("entries[2].Format = %v, want %v", got, want)
	}
	if got := SerializeEntries(entries); got != content {
		t.Fatalf("SerializeEntries() = %q, want %q", got, content)
	}
}

func TestParseHistorySerializePreservesTrailingNewline(t *testing.T) {
	content := ": 1714752000:0;git status\nplain command\n"

	parsed := ParseHistory(content)

	if len(parsed.Entries) != 2 {
		t.Fatalf("ParseHistory() returned %d entries, want 2", len(parsed.Entries))
	}
	if got := parsed.Serialize(); got != content {
		t.Fatalf("Serialize() = %q, want %q", got, content)
	}
}

func TestParseHistorySerializePreservesNoTrailingNewline(t *testing.T) {
	content := "git status"

	parsed := ParseHistory(content)

	if len(parsed.Entries) != 1 {
		t.Fatalf("ParseHistory() returned %d entries, want 1", len(parsed.Entries))
	}
	if got := parsed.Serialize(); got != content {
		t.Fatalf("Serialize() = %q, want %q", got, content)
	}
}

func TestParsedHistorySerializeEntriesUsesOriginalTrailingNewline(t *testing.T) {
	parsed := ParseHistory("drop me\nkeep me\n")
	kept := parsed.Entries[1:]

	if got, want := parsed.SerializeEntries(kept), "keep me\n"; got != want {
		t.Fatalf("SerializeEntries() = %q, want %q", got, want)
	}
}

func TestParseHistorySerializeEmptyContent(t *testing.T) {
	parsed := ParseHistory("")

	if got := len(parsed.Entries); got != 0 {
		t.Fatalf("ParseHistory() returned %d entries, want 0", got)
	}
	if got, want := parsed.Serialize(), ""; got != want {
		t.Fatalf("Serialize() = %q, want %q", got, want)
	}
}

func TestParseHistorySerializeSingleBlankLine(t *testing.T) {
	parsed := ParseHistory("\n")

	if got := len(parsed.Entries); got != 1 {
		t.Fatalf("ParseHistory() returned %d entries, want 1", got)
	}
	if got, want := parsed.Entries[0].LineNo, 1; got != want {
		t.Fatalf("blank entry LineNo = %d, want %d", got, want)
	}
	if got, want := parsed.Entries[0].Command, ""; got != want {
		t.Fatalf("blank entry Command = %q, want %q", got, want)
	}
	if got, want := parsed.Serialize(), "\n"; got != want {
		t.Fatalf("Serialize() = %q, want %q", got, want)
	}
}
