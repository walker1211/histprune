# histprune

Safe, explainable shell history cleanup in Go.

[中文](./README.zh-CN.md) | [English](./README.en.md)

## Features

- First-class zsh history support, including zsh-metafied command decoding
- Dry-run by default for safe previews
- Explain every removal with rule-based reasons
- Remove duplicates while keeping the latest occurrence
- Drop entries by literal text, regex, line number, or timestamp range
- Create backups before writes and support restore
- Text and JSON reports

## Quick Start

Install from GitHub Releases:

```bash
tar -xzf histprune_<tag>_<os>_<arch>.tar.gz
./histprune --help
```

Or build from source:

```bash
go install github.com/walker1211/histprune/cmd/histprune@latest
histprune --help
```

Analyze your default zsh history file:

```bash
histprune analyze
```

Preview cleanup first:

```bash
histprune prune --dedupe
```

Apply the cleanup after reviewing the preview:

```bash
histprune prune --dedupe --write
```

By default, `histprune` uses `$HISTFILE`, then falls back to `~/.zsh_history`. Use `--file PATH` only when you want to target another history file.

After writing changes to an active zsh history file, reload it in your current shell:

```bash
fc -R ~/.zsh_history
```

Show all commands and flags:

```bash
histprune --help
```

## License

Licensed under the MIT License. See [LICENSE](./LICENSE).
