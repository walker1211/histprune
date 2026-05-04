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

#### Option 1: Run from source

```bash
go run ./cmd/histprune version
```

#### Option 2: Install locally

```bash
go install github.com/walker1211/histprune/cmd/histprune@latest
histprune version
```

For local development, you can also build a binary:

```bash
go build -o histprune ./cmd/histprune
./histprune version
```

## Quick Start

Analyze zsh history:

```bash
histprune analyze --file ~/.zsh_history
```

Preview duplicate cleanup without writing:

```bash
histprune prune --file ~/.zsh_history --dedupe
```

Apply after reviewing the preview:

```bash
histprune prune --file ~/.zsh_history --dedupe --write
```

Drop entries containing literal text:

```bash
histprune prune --file ~/.zsh_history --contains 'gti status'
histprune prune --file ~/.zsh_history --contains 'gti status' --write
```

Drop entries matching a regex:

```bash
histprune prune --file ~/.zsh_history --regex 'token=[^ ]+'
```

Drop by line number:

```bash
histprune prune --file ~/.zsh_history --line 1287
```

Drop by date:

```bash
histprune prune --file ~/.zsh_history --before 2024-01-01
histprune prune --file ~/.zsh_history --between 2024-01-01 2024-06-30
```

Emit a JSON report:

```bash
histprune prune --file ~/.zsh_history --dedupe --json
```

## Safety Model

`prune` is a dry-run unless `--write` is explicitly passed:

```bash
histprune prune --file ~/.zsh_history --dedupe --write
```

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
histprune backups --file ~/.zsh_history
```

Restore the latest backup:

```bash
histprune restore latest --file ~/.zsh_history
```

Restore an explicit backup:

```bash
histprune restore /path/to/.zsh_history.histprune-backup-20260504T120000 --file ~/.zsh_history
```

Restoring creates a backup of the current history file first.

## Configuration

`histprune` does not read a config file yet. The current version is configured entirely through CLI flags.

This repository intentionally does not include `configs/config.example.yaml` yet, because a template config would imply config-file support that is not implemented. If structured rule configuration is added later, the layout will be:

- `configs/config.example.yaml`: committed template config
- `configs/config.yaml`: local ignored config

Sensitive data should not be stored in shell history or project config. If future features require secrets, the project will use `.example.env` plus a local ignored `.env`.

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
