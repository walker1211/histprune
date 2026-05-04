package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/walker1211/histprune/internal/cli"
)

func TestVersionPrintsCLIVersion(t *testing.T) {
	var out bytes.Buffer
	runners := NewRunners(&out, nil)

	if err := runners.Version(context.Background()); err != nil {
		t.Fatalf("Version returned error: %v", err)
	}
	if got := out.String(); got != cli.Version+"\n" {
		t.Fatalf("version output = %q, want %q", got, cli.Version+"\n")
	}
}

func TestAnalyzeTextIncludesCountsAndTopCommand(t *testing.T) {
	historyPath := copySampleHistory(t)
	var out, errOut bytes.Buffer
	runners := NewRunners(&out, &errOut)

	if err := runners.Analyze(context.Background(), cli.AnalyzeOptions{File: historyPath}); err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"Scanned: 5",
		"zsh_extended: 3",
		"plain: 1",
		"malformed: 1",
		"Duplicate commands: 1",
		"Top commands:",
		"git status: 2",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("analyze output missing %q:\n%s", want, text)
		}
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
}

func TestPruneDryRunDoesNotModifyFileAndReportsRemovedEntries(t *testing.T) {
	historyPath := copySampleHistory(t)
	before := readFileString(t, historyPath)
	var out bytes.Buffer
	runners := NewRunners(&out, nil)

	err := runners.Prune(context.Background(), cli.PruneOptions{File: historyPath, Contains: []string{"git status"}})
	if err != nil {
		t.Fatalf("Prune returned error: %v", err)
	}
	if after := readFileString(t, historyPath); after != before {
		t.Fatalf("dry-run modified file: got %q want %q", after, before)
	}
	text := out.String()
	for _, want := range []string{"Scanned: 5", "Removed: 2", "Kept: 3", "Mode: dry-run", "contains: 2"} {
		if !strings.Contains(text, want) {
			t.Fatalf("prune dry-run output missing %q:\n%s", want, text)
		}
	}
}

func TestPruneWriteCreatesBackupAndUpdatesFileContent(t *testing.T) {
	historyPath := copySampleHistory(t)
	before := readFileString(t, historyPath)
	var out bytes.Buffer
	runners := NewRunners(&out, nil)

	err := runners.Prune(context.Background(), cli.PruneOptions{File: historyPath, Contains: []string{"git status"}, Write: true})
	if err != nil {
		t.Fatalf("Prune --write returned error: %v", err)
	}
	after := readFileString(t, historyPath)
	if after == before {
		t.Fatalf("write did not update file")
	}
	if strings.Contains(after, "git status") {
		t.Fatalf("write kept pruned command in file: %q", after)
	}
	if !strings.HasSuffix(after, "\n") {
		t.Fatalf("write did not preserve trailing newline: %q", after)
	}
	backups := validBackups(t, historyPath)
	if len(backups) != 1 {
		t.Fatalf("backups = %#v, want one", backups)
	}
	backupContent := readFileString(t, backups[0])
	if backupContent != before {
		t.Fatalf("backup content = %q, want original %q", backupContent, before)
	}
	if !strings.Contains(out.String(), "Backup: "+backups[0]) {
		t.Fatalf("write output missing backup path %q:\n%s", backups[0], out.String())
	}
}

func TestPruneWriteWithNoRulesReturnsErrorWithoutBackupOrModification(t *testing.T) {
	historyPath := copySampleHistory(t)
	before := readFileString(t, historyPath)
	var out, errOut bytes.Buffer
	runners := NewRunners(&out, &errOut)

	err := runners.Prune(context.Background(), cli.PruneOptions{File: historyPath, Write: true})
	if err == nil {
		t.Fatal("Prune --write with no rules returned nil error")
	}
	if got := err.Error(); got != noPruneRulesMessage {
		t.Fatalf("error = %q, want %q", got, noPruneRulesMessage)
	}
	if after := readFileString(t, historyPath); after != before {
		t.Fatalf("no-rule write modified file: got %q want %q", after, before)
	}
	if backups := validBackups(t, historyPath); len(backups) != 0 {
		t.Fatalf("no-rule write created backups: %#v", backups)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", out.String())
	}
	if !strings.Contains(errOut.String(), noPruneRulesMessage) {
		t.Fatalf("stderr = %q, want message %q", errOut.String(), noPruneRulesMessage)
	}
}

func TestPruneDryRunWithNoRulesReturnsError(t *testing.T) {
	historyPath := copySampleHistory(t)
	before := readFileString(t, historyPath)
	var out, errOut bytes.Buffer
	runners := NewRunners(&out, &errOut)

	err := runners.Prune(context.Background(), cli.PruneOptions{File: historyPath})
	if err == nil {
		t.Fatal("Prune dry-run with no rules returned nil error")
	}
	if got := err.Error(); got != noPruneRulesMessage {
		t.Fatalf("error = %q, want %q", got, noPruneRulesMessage)
	}
	if after := readFileString(t, historyPath); after != before {
		t.Fatalf("no-rule dry-run modified file: got %q want %q", after, before)
	}
	if backups := validBackups(t, historyPath); len(backups) != 0 {
		t.Fatalf("no-rule dry-run created backups: %#v", backups)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", out.String())
	}
	if !strings.Contains(errOut.String(), noPruneRulesMessage) {
		t.Fatalf("stderr = %q, want message %q", errOut.String(), noPruneRulesMessage)
	}
}

func TestBackupsJSONListsValidBackups(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	writeFileString(t, historyPath, "current\n")
	older := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000")
	newer := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T130000")
	invalid := filepath.Join(dir, ".zsh_history.histprune-backup-zz")
	writeFileString(t, older, "older\n")
	writeFileString(t, newer, "newer\n")
	writeFileString(t, invalid, "invalid\n")
	var out bytes.Buffer
	runners := NewRunners(&out, nil)

	if err := runners.Backups(context.Background(), cli.BackupsOptions{File: historyPath, JSON: true}); err != nil {
		t.Fatalf("Backups returned error: %v", err)
	}
	var got struct {
		Backups []string `json:"backups"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", out.String(), err)
	}
	want := []string{newer, older}
	if strings.Join(got.Backups, "\n") != strings.Join(want, "\n") {
		t.Fatalf("backups = %#v, want %#v", got.Backups, want)
	}
}

func TestRestoreLatestRestoresNewestValidBackupAndBacksUpCurrentFile(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	writeFileString(t, historyPath, "current\n")
	older := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000")
	newer := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T130000")
	invalid := filepath.Join(dir, ".zsh_history.histprune-backup-zz")
	writeFileString(t, older, "older\n")
	writeFileString(t, newer, "newer\n")
	writeFileString(t, invalid, "invalid\n")
	var out bytes.Buffer
	runners := NewRunners(&out, nil)

	if err := runners.Restore(context.Background(), cli.RestoreOptions{File: historyPath, Target: "latest"}); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if got := readFileString(t, historyPath); got != "newer\n" {
		t.Fatalf("restored content = %q, want newest backup", got)
	}
	text := out.String()
	if !strings.Contains(text, "Restored: "+newer) || !strings.Contains(text, "Current backup: ") {
		t.Fatalf("restore output missing paths:\n%s", text)
	}
	currentBackup := strings.TrimSpace(strings.TrimPrefix(text[strings.Index(text, "Current backup: "):], "Current backup: "))
	if got := readFileString(t, currentBackup); got != "current\n" {
		t.Fatalf("pre-restore backup content = %q, want current", got)
	}
}

func copySampleHistory(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "zsh_history_sample"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, ".zsh_history")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func validBackups(t *testing.T, historyPath string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(historyPath))
	if err != nil {
		t.Fatal(err)
	}
	var backups []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".histprune-backup-") {
			backups = append(backups, filepath.Join(filepath.Dir(historyPath), entry.Name()))
		}
	}
	return backups
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeFileString(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
