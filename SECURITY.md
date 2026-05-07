# Security Policy

## Reporting a vulnerability

Please do not publish vulnerability details, exploit steps, tokens, credentials, private configuration, private shell history, or generated local data in a public issue, pull request, discussion, or commit.

This project does not currently publish a dedicated security email address. If you need to report a security issue, use GitHub's private vulnerability reporting for this repository if it is available from the repository Security tab.

If private vulnerability reporting is unavailable, open a minimal public GitHub issue so a maintainer can arrange private follow-up. Keep the public issue limited to:

* The affected area at a high level.
* A statement that you can share details privately with a maintainer.
* No secrets, private shell history, `.env` values, `configs/config.yaml`, local repositories, or generated local data.
* No step-by-step exploit instructions or weaponized proof-of-concept details.

## Supported scope

Security fixes are generally handled for the current `main` branch and the latest released version when releases are available. Older unreleased snapshots or local forks may not receive separate fixes.

## Project security boundaries

`histprune` is a local-first Go CLI for inspecting, pruning, backing up, and restoring shell history files. It is designed to run on files you explicitly target, or on the default zsh history path discovered from `$HISTFILE` and `~/.zsh_history`.

Important boundaries and assumptions:

* Treat shell history as potentially sensitive. It may contain tokens, passwords, hostnames, private repository paths, or internal commands.
* Review reports, logs, screenshots, test fixtures, and command examples before publishing them.
* Do not commit `.env`, `configs/config.yaml`, private shell history, generated local repositories, backups, logs, or local artifacts.
* Keep examples minimal and synthetic. Avoid copying real history entries into issues, tests, or documentation.
* Rotate or revoke any exposed credential immediately; deleting it from a later commit is not enough once it has been published.

## Secret handling

Before contributing, run the local checks and secret scanner when possible:

```bash
bash ./scripts/ci-local.sh
bash ./scripts/secret-scan.sh --current --history
```

If you accidentally commit or publish a secret, rotate or revoke it immediately. Removing it from a later commit is not enough once it has been exposed.
