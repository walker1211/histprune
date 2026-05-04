# histprune

[Landing Page](./README.md) | [中文](./README.zh-CN.md)

`histprune` is a shell history cleaner written in Go. The first version focuses on zsh history and aims to make cleanup safe, explainable, and recoverable.

## Features

- First-class zsh history support, including extended history lines
- Dry-run by default
- Reasoned removals for every dropped entry
- Duplicate removal while keeping the latest occurrence
- Literal, regex, line-number, and timestamp-range pruning
- Automatic backups before writes
- Backup listing and restore support
- Text and JSON reports

## Installation

Install the CLI:

```bash
go install github.com/walker1211/histprune/cmd/histprune@latest
histprune version
```

For local development, you can build a binary with the helper script:

```bash
./build.sh
./histprune version
```

## Quick Start

Analyze your default zsh history file:

```bash
histprune analyze
```

Preview cleanup first:

```bash
histprune prune --dedupe
```

Apply the same cleanup after reviewing the preview:

```bash
histprune prune --dedupe --write
```

By default, `histprune` uses `$HISTFILE`, then falls back to `~/.zsh_history`. Use `--file PATH` only when targeting another history file:

```bash
histprune analyze --file /path/to/history
```

Show all commands and flags:

```bash
histprune --help
```

## Prune Rules

All `prune` commands are previews unless `--write` is explicitly passed. Review the report first, then add `--write` to the same command to modify the file.

Remove duplicates while keeping the latest occurrence:

```bash
histprune prune --dedupe
```

Drop entries containing literal text:

```bash
histprune prune --contains 'gti status'
```

Drop entries matching a regex:

```bash
histprune prune --regex 'token=[^ ]+'
```

Drop by line number:

```bash
histprune prune --line 1287
```

Drop by date:

```bash
histprune prune --before 2024-01-01
histprune prune --between 2024-01-01 2024-06-30
```

Emit a JSON report:

```bash
histprune prune --dedupe --json
```

## Safety Model

`prune` is a dry-run unless `--write` is explicitly passed.

On write, `histprune` will:

1. read and parse the history file
2. build removal decisions with reasons
3. create a `*.histprune-backup-*` backup first
4. replace the target through a temporary file and rename

After changing the history file used by your current zsh session, reload it manually:

```bash
fc -R ~/.zsh_history
```

If multiple zsh terminals are open, consider writing the latest in-memory history before cleanup:

```bash
fc -W ~/.zsh_history
```

## Backups and Restore

List backups:

```bash
histprune backups
```

Restore the latest backup:

```bash
histprune restore latest
```

Restore an explicit backup:

```bash
histprune restore /path/to/.zsh_history.histprune-backup-20260504T120000
```

Restoring creates a backup of the current history file first.

## Development / Testing

```bash
go test ./...
go vet ./...
```

## Roadmap

- bash / fish history support
- config-file rules
- backup retention
- secret redaction
- typo / failed-command analysis
- Homebrew / release automation

## License

Licensed under the MIT License. See [LICENSE](./LICENSE).
