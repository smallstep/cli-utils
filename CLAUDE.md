# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Overview

`cli-utils` (module `github.com/smallstep/cli-utils`) is a small Go library of shared building blocks for Smallstep's `urfave/cli`-based command-line tools. Its main consumers are [`step`](https://github.com/smallstep/cli) and [`step-ca`](https://github.com/smallstep/certificates), which pin it at a tagged version. It is a public, Apache-2.0 library but is not a stable API: the README warns that other projects should not depend on it and that the API can change at any time. There is no binary here, only packages.

## Commands

```bash
make bootstrap          # install golangci-lint, govulncheck, gotestsum
make test               # unit tests via gotestsum (-short, coverage); this is what `make ci` runs
make race               # unit tests with the race detector
make lint               # golangci-lint (config fetched from smallstep/workflows) + govulncheck
make fmt                # goimports -l -w on all .go files
make                    # lint + test
```

Plain `go` works too and needs no special environment (no private modules, no Docker, no `go generate`):

```bash
go build ./...
go test -short ./...
go test -run TestParse ./token/          # single test
```

`make test` and `make lint` require the tools from `make bootstrap` on `$PATH`. `make lint` needs network access to download the shared golangci config. CI (`.github/workflows/ci.yml`) calls the shared `goCI` reusable workflow with `run-build: false`, so tests and lint are what gate a PR; CodeQL runs with `go build ./...`.

## Architecture

```
cli-utils/
├── command/            # Global command registry for urfave/cli apps
│   ├── command.go      #   Register/Retrieve commands; ActionFunc captures the ctx; IsForce()
│   └── version/        #   `version` command, registered in init()
├── errs/               # Error constructors with user-facing messages for flag/argument misuse
├── fileutil/           # File writes with overwrite prompts (WriteFile, WriteSnippet, AppendNewLine, ...)
├── step/               # $STEPPATH layout, contexts (profile + authority), defaults.json flag loading
│   ├── config.go       #   Path(), Home(), BasePath(), Version(), file-location helpers
│   └── context.go      #   Context/CtxState, contexts.json, SetEnvVar, getConfigVars
├── token/              # JWT claim builders (token.Options) and parsing for step provisioning tokens
│   └── provision/      #   provision.Token: builds a signed JWT from token.Options
├── ui/                 # promptui-based interactive prompts, validators, colored output to stderr
├── usage/              # Custom help templates and renderer (markdown-ish help text, HTML export)
└── pkg/blackfriday/    # Vendored fork of russross/blackfriday v2 used by usage/ (own LICENSE.txt)
```

### How the pieces fit

- A CLI registers each `cli.Command` with `command.Register`. That calls `step.SetEnvVar`, which gives every flag an `EnvVar` of `STEP_<FLAG_NAME>` (uppercased, `-` to `_`) unless one is already set, and installs `getConfigVars` as the command's `Before` hook so unset flags are filled from the active context's `defaults.json`. Set a flag's `EnvVar` to `step.IgnoreEnvVar` to opt it out of both.
- `step` resolves the config root from `STEPPATH` (default `$HOME/.step`) once, via `sync.Once`. With contexts enabled, per-authority config lives under `authorities/<name>/` and profiles under `profiles/<name>/`; `contexts.json` and `current-context.json` sit at the root. Call `step.Init()` before using these helpers.
- `usage` overrides urfave/cli's `help` command and templates. Command `Description`/`UsageText` strings use a lightweight markdown dialect (`**bold**`, `'''` fenced blocks, `## SECTIONS`) that `usage.Render` turns into terminal output via `pkg/blackfriday`; `step help --html <dir>` exports the same content as HTML.
- `fileutil.WriteFile` and friends consult `command.IsForce()`; without `--force` they prompt through `ui` before overwriting.
- `ui` prints prompts and messages to stderr (never stdout) so command output stays pipeable; `ui_windows.go`/`ui_other.go` are build-tagged for console-mode handling.

## Conventions

**CLI framework**: `urfave/cli` v1 (`github.com/urfave/cli`), not v2. Errors returned to users go through `errs` constructors so messages are consistent across `step` commands.

**Error wrapping**: `github.com/pkg/errors` throughout (`errors.Errorf`, `errors.Wrapf`); `errs.Wrap` normalizes causes for display. Do not introduce `fmt.Errorf("%w")` in a file that otherwise uses `pkg/errors`.

**Logging**: none. Output goes to `ui.Print*` (stderr) or `fmt.Print*` (stdout) as appropriate.

**Testing**: `testify` (`assert`/`require`) for new tests; a few older tests still use `github.com/smallstep/assert`. Tests that touch `$STEPPATH` use `t.TempDir()` plus `t.Setenv(step.HomeEnv, ...)` to stay hermetic. Fixtures live in `token/testdata/` (certificates and keys) and `pkg/blackfriday/testdata/` (markdown/HTML pairs). `command/`, `fileutil/`, and `usage/` have no tests.

**Vendored code**: `pkg/blackfriday/` is a copy of an upstream library with its own license and README. Keep changes there minimal and clearly motivated; the rest of the repo is where Smallstep-specific behavior belongs.

**Compatibility**: `step` and `step-ca` are the callers. Renaming or changing the signature of an exported symbol breaks them on their next dependency bump, so prefer additive changes and check both consumers before removing anything.

## Environment Variables

- `STEPPATH` — root of the step configuration directory (default `$HOME/.step`)
- `HOME` — used to derive the default `STEPPATH`; falls back to `os/user`
- `STEP_<FLAG>` — auto-derived per-flag overrides for any command registered via `command.Register`
- `STEP_IGNORE_ENV_VAR` — sentinel value, not a variable to set: assign it to a flag's `EnvVar` to disable env and defaults.json lookup for that flag

## Releases

Versions are git tags (`v0.12.x`). After tagging, bump the dependency in `step` and `step-ca`; there is no release workflow in this repo.
