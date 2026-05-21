# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build all checks
go build ./...

# Build a specific check
go build ./cmd/check-puppet-last-run/

# Run tests (cross-platform: also runs on Windows in CI)
go test -v ./...

# Lint (mirrors CI)
golangci-lint run

# Build release binaries (requires goreleaser)
goreleaser build --snapshot --clean
```

## Architecture

This is a collection of Sensu monitoring plugin binaries written in Go. Each binary lives under `cmd/<check-name>/main.go` and is built independently.

**Shared package:** `sensupluginspuppet/sensupluginspuppet.go` provides OS-aware default file paths (Windows vs. Unix) for Puppet state files. All checks import this package for their defaults.

**Plugin pattern:** Every check follows the same structure using `github.com/sensu/sensu-plugin-sdk/sensu`:
1. Define a `Config` struct embedding `sensu.PluginConfig`
2. Declare `options` as `[]sensu.ConfigOption` (maps CLI flags to config fields)
3. Implement `checkArgs` (validates preconditions, returns early with a status) and `executeCheck` (main logic)
4. Wire them together with `sensu.NewCheck(...).Execute()` in `main`

**Exit codes** follow Sensu/Nagios conventions via `sensu.CheckStateOK` (0), `sensu.CheckStateWarning` (1), `sensu.CheckStateCritical` (2), `sensu.CheckStateUnknown` (3).

**Checks:**
- `check-puppet-last-run` — checks age of last Puppet run and failure count from `last_run_summary.yaml`
- `check-puppet-errors` — checks for Puppet failures and whether the agent is disabled (reads lock file)
- `check-puppet-environment` — alerts if `environment` key is set in `puppet.conf` (critical if found)
- `check-puppet-last-run-report` — stub/placeholder, currently returns OK

**Release:** GoReleaser builds all four binaries for linux/windows × multiple architectures with `CGO_ENABLED=0`. Binaries land in `bin/` inside each archive to comply with Sensu Go Asset structure.
