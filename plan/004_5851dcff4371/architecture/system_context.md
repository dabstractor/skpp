# System Context — Delta 004

## What this is

A **delta task** against an existing, fully-working Go CLI (`skilldozer`). The codebase is NOT greenfield — it is a complete implementation built across sessions 001–003. All tests pass at HEAD `f30d5c5`.

## The delta (decisions 19 + 20)

Two coupled changes, NOT a feature explosion:

### Decision 19 — Subcommand → flag (namespace safety)
- **OLD (sessions 001–003):** `check`, `init`, `completion` are **bare-word reserved positional subcommands**. A skill literally tagged `check` is shadowed by the validation subcommand.
- **NEW:** **No bare-word subcommands.** Every non-tag action is a `--flag`. The bare positional namespace is reserved entirely for skill tags:
  - `skilldozer --check` runs validation; `skilldozer check` resolves the **tag** `check`.
  - `skilldozer --init [<dir>]` runs first-run setup; `skilldozer init` resolves the **tag** `init`.
  - `skilldozer --completions [--shell <name>]` emits the completion script (singular→plural rename); `skilldozer completions` resolves the **tag** `completions`.

### Decision 20 — Completions are skills-first & long-form-only
- **OLD:** completions offered `check`/`init`/`completion` as exclusive first-positional subcommands, and offered both `--long` and `-short` flag forms.
- **NEW:**
  - A bare `skilldozer <tab>` shows **skills** — never the help menu, never a command list.
  - `skilldozer -<tab>` / `--<tab>` offers **long-form flags only** (`--check`, `--completions`, …); short aliases (`-a`, `-l`, `-s`, `-f`, `-p`, `-h`, `-v`) stay valid at runtime but are **not advertised**.
  - Prefix filtering (`a<tab>`, `--c<tab>`) is the shell's native job.

## What is NOT in scope (do not touch)
- `internal/discover`, `internal/resolve`, `internal/check`, `internal/config`, `internal/search`, `internal/ui`, `internal/skillsdir` **logic** — unchanged. (Only `skillsdir.ErrNotFound` *message string* changes: `init` → `--init`.)
- The `//go:embed` mechanism, `completionScript()` switch body, `detectShell()` logic — unchanged.
- Short-form flags (`-v`/`-h`/`-p`/`-l`/`-a`/`-f`/`-s`) — **still valid at runtime**; only the completion menu stops advertising them.
- The example skill `skills/example/SKILL.md` — no change needed.

## Repository facts (verified at HEAD `f30d5c5`)
- **Module:** `github.com/dabstractor/skilldozer` (Go 1.25, `go.mod`)
- **Sole dependency:** `gopkg.in/yaml.v3 v3.0.1` (stdlib for everything else)
- **Structure:** `main.go` (1288 lines) + `main_test.go` (3089 lines) + `internal/{check,config,discover,resolve,search,skillsdir,ui}/` + `completions/{skilldozer.bash,_skilldozer,skilldozer.fish}` + `skills/example/SKILL.md`
- **Build:** `go build -o skilldozer .` succeeds
- **Tests:** `go test ./...` passes (green)
- **Naming:** the existing subcommand is `completion` (singular); the new flag is `--completions` (PLURAL) per PRD §6.1 decision 19.

## Verified ground truth (binary probes)

| Assertion | Current (HEAD `f30d5c5`) | Required (post-delta) | Status |
|---|---|---|---|
| `skilldozer --check` runs validation | NO (bare `check` runs it) | YES | ❌ |
| `skilldozer --init [<dir>]` runs init | NO (bare `init` runs it) | YES | ❌ |
| `skilldozer --completions` emits script | NO (bare `completion`) | YES (plural flag) | ❌ |
| `skilldozer check` resolves tag | NO (reserved subcommand) | YES (or unknown tag exit 1) | ❌ |
| `skilldozer init` resolves tag | NO (reserved subcommand) | YES (or unknown tag exit 1) | ❌ |
| ErrNotFound mentions `--init` | NO ("run \`skilldozer init\`") | YES | ❌ |
| Error prefixes say `--init:` | NO ("skilldozer init:") | YES | ❌ |
| Completions offer bare subcommands | YES | NO (skills only) | ❌ |
| Completions offer short-form flags | YES (-v -h -p -l -a -f -s) | NO (long-form only) | ❌ |
| Completions offer `--check`/`--init`/`--completions` | NO | YES | ❌ |
| `completion` → `--completions` (plural rename) | NO (singular) | YES | ❌ |
| All existing green tests pass | YES | YES (after flipping) | ✅ baseline |
