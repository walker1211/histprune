package cli

import (
	"context"
	"errors"
	"flag"
	"io"
	"strconv"
)

const Version = "dev"

const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

type AnalyzeOptions struct {
	File string
	JSON bool
}

type PruneOptions struct {
	File     string
	Dedupe   bool
	Contains []string
	Regex    []string
	Lines    []int
	Before   string
	Between  [2]string
	Write    bool
	JSON     bool
}

type BackupsOptions struct {
	File string
	JSON bool
}

type RestoreOptions struct {
	Target string
	File   string
}

type Runners struct {
	Analyze func(context.Context, AnalyzeOptions) error
	Prune   func(context.Context, PruneOptions) error
	Backups func(context.Context, BackupsOptions) error
	Restore func(context.Context, RestoreOptions) error
	Version func(context.Context) error
}

func Run(ctx context.Context, args []string, runners Runners) int {
	if len(args) == 0 {
		return ExitUsage
	}

	var err error
	switch args[0] {
	case "analyze":
		err = runAnalyze(ctx, args[1:], runners.Analyze)
	case "prune":
		err = runPrune(ctx, args[1:], runners.Prune)
	case "backups":
		err = runBackups(ctx, args[1:], runners.Backups)
	case "restore":
		err = runRestore(ctx, args[1:], runners.Restore)
	case "version":
		err = callVersion(ctx, runners.Version)
	default:
		return ExitUsage
	}

	if err == nil {
		return ExitOK
	}
	if errors.Is(err, errUsage) {
		return ExitUsage
	}
	return ExitError
}

var errUsage = errors.New("usage error")

func runAnalyze(ctx context.Context, args []string, runner func(context.Context, AnalyzeOptions) error) error {
	fs := newFlagSet("analyze")
	var opts AnalyzeOptions
	fs.StringVar(&opts.File, "file", "", "history file")
	fs.BoolVar(&opts.JSON, "json", false, "emit JSON")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return errUsage
	}
	if runner == nil {
		return nil
	}
	return runner(ctx, opts)
}

func runPrune(ctx context.Context, args []string, runner func(context.Context, PruneOptions) error) error {
	var opts PruneOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.File = value
		case "--dedupe":
			opts.Dedupe = true
		case "--contains":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.Contains = append(opts.Contains, value)
		case "--regex":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.Regex = append(opts.Regex, value)
		case "--line":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			line, err := strconv.Atoi(value)
			if err != nil {
				return errUsage
			}
			opts.Lines = append(opts.Lines, line)
		case "--before":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.Before = value
		case "--between":
			start, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			end, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.Between = [2]string{start, end}
		case "--write":
			opts.Write = true
		case "--json":
			opts.JSON = true
		default:
			return errUsage
		}
	}
	if runner == nil {
		return nil
	}
	return runner(ctx, opts)
}

func runBackups(ctx context.Context, args []string, runner func(context.Context, BackupsOptions) error) error {
	fs := newFlagSet("backups")
	var opts BackupsOptions
	fs.StringVar(&opts.File, "file", "", "history file")
	fs.BoolVar(&opts.JSON, "json", false, "emit JSON")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return errUsage
	}
	if runner == nil {
		return nil
	}
	return runner(ctx, opts)
}

func runRestore(ctx context.Context, args []string, runner func(context.Context, RestoreOptions) error) error {
	var opts RestoreOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			value, ok := nextArg(args, &i)
			if !ok {
				return errUsage
			}
			opts.File = value
		default:
			if opts.Target != "" {
				return errUsage
			}
			opts.Target = args[i]
		}
	}
	if opts.Target == "" {
		return errUsage
	}
	if runner == nil {
		return nil
	}
	return runner(ctx, opts)
}

func callVersion(ctx context.Context, runner func(context.Context) error) error {
	if runner == nil {
		return nil
	}
	return runner(ctx)
}

func nextArg(args []string, i *int) (string, bool) {
	if *i+1 >= len(args) {
		return "", false
	}
	(*i)++
	return args[*i], true
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
