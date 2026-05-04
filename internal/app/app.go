package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/walker1211/histprune/internal/cli"
	"github.com/walker1211/histprune/internal/history"
	"github.com/walker1211/histprune/internal/prune"
	"github.com/walker1211/histprune/internal/render"
	"github.com/walker1211/histprune/internal/report"
	"github.com/walker1211/histprune/internal/storage"
)

const (
	topCommandLimit     = 10
	noPruneRulesMessage = "no prune rules specified; pass --dedupe, --contains, --regex, --line, --before, or --between"
)

// NewRunners returns cli runners wired to the real application implementation.
func NewRunners(stdout, stderr io.Writer) cli.Runners {
	app := App{stdout: stdout, stderr: stderr}
	if app.stdout == nil {
		app.stdout = io.Discard
	}
	if app.stderr == nil {
		app.stderr = io.Discard
	}
	return cli.Runners{
		Analyze: app.analyze,
		Prune:   app.prune,
		Backups: app.backups,
		Restore: app.restore,
		Version: app.version,
	}
}

// App coordinates CLI options with history parsing, pruning, reporting, and storage.
type App struct {
	stdout io.Writer
	stderr io.Writer
}

func (a App) analyze(ctx context.Context, opts cli.AnalyzeOptions) error {
	return a.run(func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := resolveHistoryPath(opts.File)
		if err != nil {
			return err
		}
		parsed, err := readParsedHistory(path)
		if err != nil {
			return err
		}
		summary := analyzeHistory(parsed.Entries)
		if opts.JSON {
			return writeJSON(a.stdout, summary)
		}
		return writeAnalyzeText(a.stdout, summary)
	})
}

func (a App) prune(ctx context.Context, opts cli.PruneOptions) error {
	return a.run(func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := resolveHistoryPath(opts.File)
		if err != nil {
			return err
		}
		if !hasPruneRules(opts) {
			return fmt.Errorf(noPruneRulesMessage)
		}
		parsed, err := readParsedHistory(path)
		if err != nil {
			return err
		}
		engine, err := prune.BuildRules(pruneOptionsFromCLI(opts))
		if err != nil {
			return err
		}
		decisions := engine.Decide(parsed.Entries)
		backupPath := ""
		if opts.Write {
			backupPath, err = storage.CreateBackup(path)
			if err != nil {
				return err
			}
			kept := prune.FilterRemoved(decisions)
			if err := storage.AtomicWriteFile(path, []byte(parsed.SerializeEntries(kept))); err != nil {
				return err
			}
		}
		summary := report.FromDecisions(decisions, !opts.Write, backupPath)
		if opts.JSON {
			text, err := report.JSON(summary)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(a.stdout, text)
			return err
		}
		return report.WriteText(a.stdout, summary)
	})
}

func (a App) backups(ctx context.Context, opts cli.BackupsOptions) error {
	return a.run(func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := resolveHistoryPath(opts.File)
		if err != nil {
			return err
		}
		backups, err := storage.ListBackups(path)
		if err != nil {
			return err
		}
		if opts.JSON {
			return writeJSON(a.stdout, backupsPayload{Backups: backups})
		}
		if len(backups) == 0 {
			_, err = io.WriteString(a.stdout, "No backups found.\n")
			return err
		}
		for _, backup := range backups {
			if _, err := fmt.Fprintln(a.stdout, backup); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a App) restore(ctx context.Context, opts cli.RestoreOptions) error {
	return a.run(func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		path, err := resolveHistoryPath(opts.File)
		if err != nil {
			return err
		}
		target := opts.Target
		if target == "" {
			return fmt.Errorf("restore target is required")
		}
		if target == "latest" {
			backups, err := storage.ListBackups(path)
			if err != nil {
				return err
			}
			if len(backups) == 0 {
				return fmt.Errorf("no backups found for %s", path)
			}
			target = backups[0]
		}
		currentBackup, err := storage.RestoreBackup(path, target)
		if err != nil {
			return err
		}
		return writeRestoreText(a.stdout, target, currentBackup)
	})
}

func (a App) version(ctx context.Context) error {
	return a.run(func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := fmt.Fprintln(a.stdout, cli.Version)
		return err
	})
}

func (a App) run(fn func() error) error {
	err := fn()
	if err != nil {
		content := render.NewContent(a.stderr)
		content.Writef("error: %v\n", err)
	}
	return err
}

func resolveHistoryPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return storage.DefaultHistoryPath()
}

func readParsedHistory(path string) (history.ParsedHistory, error) {
	content, err := storage.ReadFile(path)
	if err != nil {
		return history.ParsedHistory{}, err
	}
	return history.ParseHistory(string(content)), nil
}

func hasPruneRules(opts cli.PruneOptions) bool {
	return opts.Dedupe ||
		len(opts.Contains) > 0 ||
		len(opts.Regex) > 0 ||
		len(opts.Lines) > 0 ||
		opts.Before != "" ||
		opts.Between[0] != "" ||
		opts.Between[1] != ""
}

func pruneOptionsFromCLI(opts cli.PruneOptions) prune.Options {
	return prune.Options{
		Contains: append([]string(nil), opts.Contains...),
		Regex:    append([]string(nil), opts.Regex...),
		Lines:    append([]int(nil), opts.Lines...),
		Before:   opts.Before,
		Between: prune.DateRange{
			Start: opts.Between[0],
			End:   opts.Between[1],
		},
		Dedupe: opts.Dedupe,
	}
}

type analyzeSummary struct {
	Scanned           int            `json:"scanned"`
	Formats           map[string]int `json:"formats"`
	DuplicateCommands int            `json:"duplicate_commands"`
	TopCommands       []commandCount `json:"top_commands"`
}

type commandCount struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
}

type backupsPayload struct {
	Backups []string `json:"backups"`
}

func analyzeHistory(entries []history.Entry) analyzeSummary {
	formats := map[string]int{
		"zsh_extended": 0,
		"plain":        0,
		"malformed":    0,
	}
	commands := make(map[string]int, len(entries))
	duplicates := 0
	for _, entry := range entries {
		formats[formatName(entry.Format)]++
		commands[entry.Command]++
		if commands[entry.Command] > 1 {
			duplicates++
		}
	}
	counts := make([]commandCount, 0, len(commands))
	for command, count := range commands {
		counts = append(counts, commandCount{Command: command, Count: count})
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count != counts[j].Count {
			return counts[i].Count > counts[j].Count
		}
		return counts[i].Command < counts[j].Command
	})
	if len(counts) > topCommandLimit {
		counts = counts[:topCommandLimit]
	}
	return analyzeSummary{
		Scanned:           len(entries),
		Formats:           formats,
		DuplicateCommands: duplicates,
		TopCommands:       counts,
	}
}

func formatName(format history.Format) string {
	switch format {
	case history.FormatZshExtended:
		return "zsh_extended"
	case history.FormatMalformed:
		return "malformed"
	default:
		return "plain"
	}
}

func writeAnalyzeText(w io.Writer, summary analyzeSummary) error {
	content := render.NewContent(w)
	content.Writef("Scanned: %d\n", summary.Scanned)
	content.WriteString("Formats:\n")
	for _, name := range []string{"zsh_extended", "plain", "malformed"} {
		content.Writef("  %s: %d\n", name, summary.Formats[name])
	}
	content.Writef("Duplicate commands: %d\n", summary.DuplicateCommands)
	content.WriteString("Top commands:\n")
	if len(summary.TopCommands) == 0 {
		content.WriteString("  (none)\n")
		return content.Err()
	}
	for _, top := range summary.TopCommands {
		content.Writef("  %s: %d\n", top.Command, top.Count)
	}
	return content.Err()
}

func writeRestoreText(w io.Writer, restoredPath, currentBackupPath string) error {
	content := render.NewContent(w)
	content.Writef("Restored: %s\n", restoredPath)
	content.Writef("Current backup: %s\n", currentBackupPath)
	return content.Err()
}

func writeJSON(w io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}
