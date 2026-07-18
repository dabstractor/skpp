# System Context — Plan 006

## Project

**Skilldozer** — a Go CLI that resolves human-friendly skill tags to absolute
filesystem paths for loading into `pi` via `--skill <path>`.

Repo: `github.com/dabstractor/skilldozer` at `~/projects/skilldozer`.
Language: Go 1.25. Single third-party dependency: `gopkg.in/yaml.v3`.

## Current State (pre-006)

The project is **fully implemented** per the prior PRD (plan/001). All core
features work: skills dir resolution (§8.3), frontmatter parsing (§7.3), tag
resolution (§7.2), `--list`/`--search`/`--check`/`--all`/`--path`/`--init`/
`--completions`, error semantics (§6.4), shell completions (§14), install.sh,
README, and the example skill. The full test suite (3,984 lines across 10 test
files) passes.

## The Delta (what 006 adds)

**`--link` multi-target batch linking** (PRD §8.4, §6.1, §6.3, §6.4, decisions
21+23). The PRD was updated (commit `983672e`) to extend `--link` from
single-target (`--link <dir>`) to multi-target batch
(`--link <dir> [<dir>...]`), but **no code was changed** — the commit touched
only `PRD.md`. The current binary exits 2 with `"'--link' cannot be combined
with tag arguments"` when given multiple directories.

### What changes

| Area | Current (single-target) | Target (batch) |
|---|---|---|
| `config` struct | `linkTarget string` | `linkTargets []string` |
| `parseArgs` | Captures ONE token after `--link` | Collects ALL following non-flag positionals |
| `exclusivityError` | Rejects trailing positionals as "tags" | Link-collected positionals bypass tag exclusivity |
| `runLink` | Processes one target | Loops over targets; partial success; exit 0 iff all link |
| Error message | `--link requires a path to a skill directory` | `--link requires at least one path to a skill directory` |
| Usage text | `--link <dir>` | `--link <dir> [<dir>...]` |
| Completions | Dir completion only at `--link <tab>` | Dir completion at every position after `--link` |

### What does NOT change

All other packages (`skillsdir`, `discover`, `resolve`, `search`, `check`,
`config`, `ui`) are unaffected — discovery already follows symlinks (§7.1), so
a linked skill resolves by its link name with no new discovery code. The
`install.sh`, `go.mod`, `.gitignore`, `LICENSE`, and `skills/example/SKILL.md`
are untouched.

## Key Files Touched

1. **`main.go`** (1577 lines) — config struct, parseArgs, exclusivityError,
   runLink, usageText
2. **`main_test.go`** (3984 lines) — existing link tests updated + new batch tests
3. **`completions/skilldozer.bash`** (79 lines) — multi-link dir completion
4. **`completions/_skilldozer`** (66 lines) — zsh multi-value handling
5. **`completions/skilldozer.fish`** (69 lines) — multi-link dir completion
6. **`README.md`** — `--link` section updated for batch behavior
