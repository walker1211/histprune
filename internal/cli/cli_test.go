package cli

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestAnalyzeParsesFileAndJSON(t *testing.T) {
	var got AnalyzeOptions
	r := Runners{
		Analyze: func(ctx context.Context, opts AnalyzeOptions) error {
			got = opts
			return nil
		},
	}

	code := Run(context.Background(), []string{"analyze", "--file", "/tmp/.zsh_history", "--json"}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	want := AnalyzeOptions{File: "/tmp/.zsh_history", JSON: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AnalyzeOptions = %#v, want %#v", got, want)
	}
}

func TestPruneParsesSelectionFlags(t *testing.T) {
	var got PruneOptions
	r := Runners{
		Prune: func(ctx context.Context, opts PruneOptions) error {
			got = opts
			return nil
		},
	}

	code := Run(context.Background(), []string{
		"prune",
		"--file", "/tmp/hist",
		"--dedupe",
		"--contains", "token",
		"--contains", "secret",
		"--regex", "aws_[A-Z]+",
		"--line", "12",
		"--line", "20",
		"--before", "2026-01-02",
		"--between", "2026-01-01", "2026-01-31",
		"--write",
		"--json",
	}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	want := PruneOptions{
		File:     "/tmp/hist",
		Dedupe:   true,
		Contains: []string{"token", "secret"},
		Regex:    []string{"aws_[A-Z]+"},
		Lines:    []int{12, 20},
		Before:   "2026-01-02",
		Between:  [2]string{"2026-01-01", "2026-01-31"},
		Write:    true,
		JSON:     true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PruneOptions = %#v, want %#v", got, want)
	}
}

func TestBackupsParsesFileAndJSON(t *testing.T) {
	var got BackupsOptions
	r := Runners{
		Backups: func(ctx context.Context, opts BackupsOptions) error {
			got = opts
			return nil
		},
	}

	code := Run(context.Background(), []string{"backups", "--file", "/tmp/hist", "--json"}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	want := BackupsOptions{File: "/tmp/hist", JSON: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BackupsOptions = %#v, want %#v", got, want)
	}
}

func TestRestoreParsesLatestAndFile(t *testing.T) {
	var got RestoreOptions
	r := Runners{
		Restore: func(ctx context.Context, opts RestoreOptions) error {
			got = opts
			return nil
		},
	}

	code := Run(context.Background(), []string{"restore", "latest", "--file", "/tmp/hist"}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	want := RestoreOptions{Target: "latest", File: "/tmp/hist"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RestoreOptions = %#v, want %#v", got, want)
	}
}

func TestRestoreParsesBackupPath(t *testing.T) {
	var got RestoreOptions
	r := Runners{
		Restore: func(ctx context.Context, opts RestoreOptions) error {
			got = opts
			return nil
		},
	}

	code := Run(context.Background(), []string{"restore", "/tmp/backup.hist"}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	want := RestoreOptions{Target: "/tmp/backup.hist"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RestoreOptions = %#v, want %#v", got, want)
	}
}

func TestVersionCallsRunner(t *testing.T) {
	called := false
	r := Runners{
		Version: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	code := Run(context.Background(), []string{"version"}, r)

	if code != 0 {
		t.Fatalf("Run returned code %d, want 0", code)
	}
	if !called {
		t.Fatal("version runner was not called")
	}
}

func TestHelpAliasesCallRunner(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			called := false
			r := Runners{
				Help: func(ctx context.Context) error {
					called = true
					return nil
				},
			}

			code := Run(context.Background(), args, r)

			if code != 0 {
				t.Fatalf("Run returned code %d, want 0", code)
			}
			if !called {
				t.Fatal("help runner was not called")
			}
		})
	}
}

func TestUnknownSubcommandReturnsUsageError(t *testing.T) {
	code := Run(context.Background(), []string{"unknown"}, Runners{})

	if code != 2 {
		t.Fatalf("Run returned code %d, want 2", code)
	}
}
