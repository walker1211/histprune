# histprune

Safe, explainable shell history cleanup in Go.

[中文文档](./README.zh-CN.md) | [English Documentation](./README.en.md)

## Features

- First-class zsh history support
- Dry-run by default for safe previews
- Explain every removal with rule-based reasons
- Remove duplicates while keeping the latest occurrence
- Drop entries by literal text, regex, line number, or timestamp range
- Create backups before writes and support restore
- Text and JSON reports

## Quick Start

```bash
go run ./cmd/histprune version
go run ./cmd/histprune analyze --file ~/.zsh_history
go run ./cmd/histprune prune --file ~/.zsh_history --dedupe
go run ./cmd/histprune prune --file ~/.zsh_history --dedupe --write
```

After writing changes to an active zsh history file, reload it in your current shell:

```bash
fc -R ~/.zsh_history
```

## Configuration

`histprune` does not read a config file yet. The current version is configured entirely through CLI flags.

If structured configuration is added later, the repository will use:

- `configs/config.example.yaml` for committed template config
- `configs/config.yaml` for local ignored config

## License

Licensed under the MIT License. See [LICENSE](./LICENSE).
