package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupTimeLayout = "20060102T150405"
const backupMarker = ".histprune-backup-"

// CreateBackup copies the current history file to a timestamped backup path in the same directory.
func CreateBackup(historyPath string) (string, error) {
	content, err := os.ReadFile(historyPath)
	if err != nil {
		return "", err
	}
	backupPath := backupPathFor(historyPath, time.Now().UTC())
	if _, err := os.Stat(backupPath); err == nil {
		backupPath = backupPath + fmt.Sprintf(".%09d", time.Now().UnixNano()%1_000_000_000)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := AtomicWriteFile(backupPath, content); err != nil {
		return "", err
	}
	return backupPath, nil
}

// ListBackups returns valid histprune backups for a history file sorted newest first.
func ListBackups(historyPath string) ([]string, error) {
	dir := filepath.Dir(historyPath)
	base := filepath.Base(historyPath)
	prefix := base + backupMarker
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	backups := make([]backupFile, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		timestamp, ok := parseBackupTimestamp(name, prefix)
		if !ok {
			continue
		}
		backups = append(backups, backupFile{path: filepath.Join(dir, name), name: name, timestamp: timestamp})
	}
	sort.Slice(backups, func(i, j int) bool {
		if !backups[i].timestamp.Equal(backups[j].timestamp) {
			return backups[i].timestamp.After(backups[j].timestamp)
		}
		return backups[i].name > backups[j].name
	})
	paths := make([]string, 0, len(backups))
	for _, backup := range backups {
		paths = append(paths, backup.path)
	}
	return paths, nil
}

// RestoreLatest restores the newest backup and first backs up the current history file.
func RestoreLatest(historyPath string) (string, error) {
	backups, err := ListBackups(historyPath)
	if err != nil {
		return "", err
	}
	if len(backups) == 0 {
		return "", fmt.Errorf("no backups found for %s", historyPath)
	}
	return RestoreBackup(historyPath, backups[0])
}

// RestoreBackup restores an explicit backup path and first backs up the current history file.
func RestoreBackup(historyPath, backupPath string) (string, error) {
	if err := validateBackupPathForHistory(historyPath, backupPath); err != nil {
		return "", err
	}
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		return "", err
	}
	currentBackup, err := CreateBackup(historyPath)
	if err != nil {
		return "", err
	}
	if err := AtomicWriteFile(historyPath, backupContent); err != nil {
		return "", err
	}
	return currentBackup, nil
}

type backupFile struct {
	path      string
	name      string
	timestamp time.Time
}

func backupPathFor(historyPath string, ts time.Time) string {
	return filepath.Join(filepath.Dir(historyPath), filepath.Base(historyPath)+backupMarker+ts.Format(backupTimeLayout))
}

func validateBackupPathForHistory(historyPath, backupPath string) error {
	historyDir, err := filepath.Abs(filepath.Dir(historyPath))
	if err != nil {
		return err
	}
	backupDir, err := filepath.Abs(filepath.Dir(backupPath))
	if err != nil {
		return err
	}
	if backupDir != historyDir {
		return fmt.Errorf("backup path %s is not in history directory %s", backupPath, filepath.Dir(historyPath))
	}

	prefix := filepath.Base(historyPath) + backupMarker
	backupName := filepath.Base(backupPath)
	if !strings.HasPrefix(backupName, prefix) {
		return fmt.Errorf("backup path %s does not match history file %s", backupPath, historyPath)
	}
	if _, ok := parseBackupTimestamp(backupName, prefix); !ok {
		return fmt.Errorf("backup path %s is not a valid histprune backup", backupPath)
	}
	return nil
}

func parseBackupTimestamp(name, prefix string) (time.Time, bool) {
	rest := strings.TrimPrefix(name, prefix)
	if len(rest) < len(backupTimeLayout) {
		return time.Time{}, false
	}
	timestampText := rest[:len(backupTimeLayout)]
	suffix := rest[len(backupTimeLayout):]
	if suffix != "" {
		if len(suffix) != 10 || suffix[0] != '.' || !allDigits(suffix[1:]) {
			return time.Time{}, false
		}
	}
	timestamp, err := time.Parse(backupTimeLayout, timestampText)
	if err != nil || timestamp.Format(backupTimeLayout) != timestampText {
		return time.Time{}, false
	}
	return timestamp, true
}

func allDigits(text string) bool {
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return text != ""
}
