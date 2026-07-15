# External Dependencies & Technical Verification — Delta 004

## No new dependencies

This delta introduces **zero new dependencies**. All changes are to existing Go source files and shell completion scripts.

## Existing dependencies (unchanged)
- `gopkg.in/yaml.v3 v3.0.1` — the ONLY non-stdlib dependency (frontmatter parsing). No change needed.
- `embed` (stdlib) — `//go:embed` directives already exist from session 003; this delta does NOT touch them (the embedded files are rewritten in place; a rebuild is required for `--completions` to reflect edits, per §14.6 lockstep).
- Go 1.25 (`go.mod` `go` directive) — no version change.

## Reusable prior research (from session 003)
The following research documents from `plan/003_3ace946c2a5c/architecture/` remain valid and are reused:
- `external_deps.md` — embed verified, `//go:embed completions/_skilldozer` works without `all:` prefix. Shell detection idiom confirmed.
- `code_change_map.md` — line-number conventions for main.go (NOTE: line numbers have shifted since 003; use the verified line numbers from `main_go_analysis.md` in THIS session's architecture directory).

## Shell completion mechanics (unchanged from 003)
- The `--completions` flag emits the `//go:embed`-ded script to stdout for `eval`.
- `completionScript(shell)` is a pure switch over package-scope string vars — no filesystem access.
- On-disk `completions/*` files are the single source of truth; `go build` bakes them in.
- `TestEmbeddedCompletionsMatchOnDisk` (main_test.go:2929) asserts byte-identity between embedded vars and on-disk files.

## Naming convention (CRITICAL)
- **Old:** `completion` (singular bare-word subcommand), `c.completion` field, `runCompletion()` function, `--shell` implies `c.completion=true`
- **New user-facing flag:** `--completions` (PLURAL per PRD §6.1 decision 19)
- **Internal names:** `c.completion` field, `runCompletion()`, `completionScript()`, `detectShell()` — **STAY AS-IS** (internal; renaming is cosmetic churn and adds risk). Only their **doc comments** move to reference `--completions`.
- The `--shell` flag still implies `c.completion=true` — unchanged.
