package render

import (
	"errors"
	"testing"
)

func TestContentWritesFormattedText(t *testing.T) {
	writer := &recordingWriter{}
	content := NewContent(writer)

	content.WriteString("Scanned: ")
	content.Writef("%d\n", 3)

	if err := content.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if got, want := writer.text, "Scanned: 3\n"; got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestContentReturnsFirstWriteErrorAndStops(t *testing.T) {
	wantErr := errors.New("write failed")
	writer := &recordingWriter{err: wantErr}
	content := NewContent(writer)

	content.Writef("first %s", "write")
	content.WriteString("second write")

	if !errors.Is(content.Err(), wantErr) {
		t.Fatalf("Err() = %v, want %v", content.Err(), wantErr)
	}
	if got, want := writer.calls, 1; got != want {
		t.Fatalf("writer calls = %d, want %d", got, want)
	}
}

type recordingWriter struct {
	text  string
	err   error
	calls int
}

func (w *recordingWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.err != nil {
		return 0, w.err
	}
	w.text += string(p)
	return len(p), nil
}
