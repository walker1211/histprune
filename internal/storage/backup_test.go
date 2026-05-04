package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultHistoryPathUsesHistfileThenHomeFallback(t *testing.T) {
	t.Setenv("HISTFILE", "/tmp/custom-history")
	path, err := DefaultHistoryPath()
	if err != nil {
		t.Fatalf("DefaultHistoryPath returned error: %v", err)
	}
	if path != "/tmp/custom-history" {
		t.Fatalf("path = %q, want HISTFILE", path)
	}

	home := t.TempDir()
	t.Setenv("HISTFILE", "")
	t.Setenv("HOME", home)
	path, err = DefaultHistoryPath()
	if err != nil {
		t.Fatalf("DefaultHistoryPath fallback returned error: %v", err)
	}
	if path != filepath.Join(home, ".zsh_history") {
		t.Fatalf("fallback path = %q", path)
	}
}

func TestReadFileAndAtomicWriteContentAndPermissions(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	if err := os.WriteFile(historyPath, []byte("old\n"), 0o640); err != nil {
		t.Fatal(err)
	}

	content, err := ReadFile(historyPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "old\n" {
		t.Fatalf("content = %q", content)
	}

	if err := AtomicWriteFile(historyPath, []byte("new\n")); err != nil {
		t.Fatalf("AtomicWriteFile returned error: %v", err)
	}
	written, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "new\n" {
		t.Fatalf("written content = %q", written)
	}
	info, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("permissions = %o, want 640", got)
	}
}

func TestCreateBackupWritesContentWithTimestampedName(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	if err := os.WriteFile(historyPath, []byte("history\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	backupPath, err := CreateBackup(historyPath)
	if err != nil {
		t.Fatalf("CreateBackup returned error: %v", err)
	}
	if filepath.Dir(backupPath) != dir {
		t.Fatalf("backup dir = %q, want %q", filepath.Dir(backupPath), dir)
	}
	base := filepath.Base(backupPath)
	if !strings.HasPrefix(base, ".zsh_history.histprune-backup-") {
		t.Fatalf("backup name = %q", base)
	}
	if len(strings.TrimPrefix(base, ".zsh_history.histprune-backup-")) != len("20060102T150405") {
		t.Fatalf("backup timestamp has unexpected shape: %q", base)
	}
	content, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "history\n" {
		t.Fatalf("backup content = %q", content)
	}
}

func TestListBackupsNewestFirst(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	older := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000")
	newer := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T130000")
	tieBreaker := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T130000.000000001")
	for _, path := range []string{older, newer, tieBreaker, filepath.Join(dir, ".other.histprune-backup-20260503T140000")} {
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	backups, err := ListBackups(historyPath)
	if err != nil {
		t.Fatalf("ListBackups returned error: %v", err)
	}
	if len(backups) != 3 {
		t.Fatalf("backups = %#v", backups)
	}
	if backups[0] != tieBreaker || backups[1] != newer || backups[2] != older {
		t.Fatalf("backups not newest first with deterministic tie-breaker: %#v", backups)
	}
}

func TestListBackupsIgnoresInvalidPrefixMatches(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	valid := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000")
	invalidNames := []string{
		".zsh_history.histprune-backup-zz",
		".zsh_history.histprune-backup-20260503T120000.extra",
		".zsh_history.histprune-backup-20260503T120000.1",
		".zsh_history.histprune-backup-20260503T120000.abcdefghi",
		".zsh_history.histprune-backup-20261303T120000",
	}
	if err := os.WriteFile(valid, []byte("valid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range invalidNames {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("invalid\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	backups, err := ListBackups(historyPath)
	if err != nil {
		t.Fatalf("ListBackups returned error: %v", err)
	}
	if len(backups) != 1 || backups[0] != valid {
		t.Fatalf("ListBackups included invalid names: %#v", backups)
	}
}

func TestRestoreLatestCreatesCurrentBackupAndReplacesHistory(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	if err := os.WriteFile(historyPath, []byte("current\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	older := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000")
	newer := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T130000")
	invalidLexicographicallyLarger := filepath.Join(dir, ".zsh_history.histprune-backup-zz")
	if err := os.WriteFile(older, []byte("older\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newer, []byte("newer\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(invalidLexicographicallyLarger, []byte("invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	currentBackup, err := RestoreLatest(historyPath)
	if err != nil {
		t.Fatalf("RestoreLatest returned error: %v", err)
	}
	restored, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "newer\n" {
		t.Fatalf("restored content = %q", restored)
	}
	currentContent, err := os.ReadFile(currentBackup)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentContent) != "current\n" {
		t.Fatalf("current backup content = %q", currentContent)
	}
}

func TestRestoreBackupRejectsArbitraryExplicitBackupPath(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	backupPath := filepath.Join(dir, "explicit.backup")
	if err := os.WriteFile(historyPath, []byte("current\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath, []byte("explicit\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := RestoreBackup(historyPath, backupPath); err == nil {
		t.Fatalf("RestoreBackup accepted arbitrary backup path")
	}
	restored, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "current\n" {
		t.Fatalf("history was modified after rejected restore: %q", restored)
	}
}

func TestRestoreBackupRejectsBackupPathOutsideHistoryDirectory(t *testing.T) {
	dir := t.TempDir()
	otherDir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	backupPath := filepath.Join(otherDir, ".zsh_history.histprune-backup-20260503T120000")
	if err := os.WriteFile(historyPath, []byte("current\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath, []byte("explicit\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := RestoreBackup(historyPath, backupPath); err == nil {
		t.Fatalf("RestoreBackup accepted backup outside history directory")
	}
	restored, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "current\n" {
		t.Fatalf("history was modified after rejected restore: %q", restored)
	}
}

func TestRestoreBackupUsesValidExplicitBackupPath(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, ".zsh_history")
	backupPath := filepath.Join(dir, ".zsh_history.histprune-backup-20260503T120000.000000001")
	if err := os.WriteFile(historyPath, []byte("current\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backupPath, []byte("explicit\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	currentBackup, err := RestoreBackup(historyPath, backupPath)
	if err != nil {
		t.Fatalf("RestoreBackup returned error: %v", err)
	}
	if currentBackup == "" {
		t.Fatalf("expected backup path for current history")
	}
	restored, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "explicit\n" {
		t.Fatalf("restored content = %q", restored)
	}
}
