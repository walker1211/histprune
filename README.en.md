# histprune

[Landing Page](./README.md) | [中文](./README.zh-CN.md)

`histprune` is a shell history cleaner written in Go. The first version focuses on zsh history and aims to make cleanup safe, explainable, and recoverable.

## Features

- First-class zsh history support, including extended history lines and zsh-metafied command decoding
- Dry-run by default
- Reasoned removals for every dropped entry
- Duplicate removal while keeping the latest occurrence
- Literal, regex, line-number, and timestamp-range pruning
- Automatic backups before writes
- Backup listing and restore support
- Text and JSON reports

## Installation

Download a release archive from GitHub Releases and unpack it:

```bash
tar -xzf histprune_<tag>_<os>_<arch>.tar.gz
./histprune --help
```

Or install from source with Go:

```bash
go install github.com/walker1211/histprune/cmd/histprune@latest
histprune --help
```

For local development, you can build a binary with the helper script:

```bash
bash ./build.sh
./histprune --help
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

## Optional zsh History Configuration

`histprune` does not require changing your zsh configuration, but these settings make the history file easier to clean later:

```zsh
# zsh history
HISTFILE="$HOME/.zsh_history"
HISTSIZE=50000
SAVEHIST=50000

setopt EXTENDED_HISTORY        # Store timestamps
setopt APPEND_HISTORY          # Append to the history file
setopt INC_APPEND_HISTORY      # Write each command immediately after it runs
```

`EXTENDED_HISTORY` stores timestamps, making `--before` / `--between` range rules more useful. `APPEND_HISTORY` and `INC_APPEND_HISTORY` reduce the chance of multiple terminals overwriting each other's history and keep the file closer to the current shell state.

If you want zsh to reduce future duplicate history entries on its own, you can also enable:

```zsh
setopt HIST_IGNORE_DUPS        # Do not record consecutive duplicate commands
setopt HIST_IGNORE_ALL_DUPS    # Remove older duplicates when a new duplicate is added
setopt HIST_SAVE_NO_DUPS       # Do not write duplicate entries when saving history
setopt HIST_EXPIRE_DUPS_FIRST  # Expire duplicate entries first when history is full
setopt HIST_FIND_NO_DUPS       # Hide duplicates when searching history
setopt HIST_REDUCE_BLANKS      # Compress extra whitespace
setopt HIST_IGNORE_SPACE       # Do not record commands that start with a space
```

These options do not guarantee that `~/.zsh_history` will never contain duplicates. Existing entries, history written before the configuration took effect, and concurrent writes from multiple terminals may still be reported by `histprune prune --dedupe`. Treat the zsh configuration as a way to reduce new duplicates, while `histprune` provides previewable, backed-up, and recoverable cleanup for the history file on disk.

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
bash ./scripts/ci-local.sh
```

## Roadmap

- bash / fish history support
- config-file rules
- backup retention
- secret redaction
- typo / failed-command analysis
- Homebrew packaging

## License

Licensed under the MIT License. See [LICENSE](./LICENSE).
