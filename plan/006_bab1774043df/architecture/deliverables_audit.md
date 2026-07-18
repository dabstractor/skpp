# Deliverables Audit — skilldozer

Run context: task 006 (delta PRD), verifying the shipped deliverables against
PRD §11–§16 and observing actual binary behavior. All evidence collected by
reading files and executing the binary/tests in this repo at
`/home/dustin/projects/skilldozer`.

Verdict legend: **COMPLETE** (meets spec), **PARTIAL** (works but with a specific
gap), **MISSING** (not present / non-functional).

---

## 1. `completions/` (bash + zsh + fish) — **PARTIAL**

Files audited:
- `completions/_skilldozer` (lines 1-66) — zsh autoload
- `completions/skilldozer.bash` (lines 1-79) — bash
- `completions/skilldozer.fish` (lines 1-69) — fish

### Requirements check

| Requirement (PRD) | Status | Evidence |
|---|---|---|
| Dynamic tag completion via `skilldozer --relative --all` | COMPLETE | All three shells invoke it: zsh `:64` `tags=(${(f)"$(skilldozer --relative --all 2>/dev/null)"})`; bash `:60` `tags=$(skilldozer --relative --all 2>/dev/null)`; fish `:69` `-a '(skilldozer --relative --all 2>/dev/null)'` |
| Errors swallowed (`2>/dev/null`) — §14.3 | COMPLETE | All three pipe `2>/dev/null` |
| Skills-first (bare `<tab>` = tags, never help/commands) — §14.1 r4 | COMPLETE | Positionals unconditionally offer tags; no subcommand namespace |
| Long-form-only flags — §14.1 r3 | COMPLETE | Each flag list omits short aliases (`-a`, `-l`, …); zsh `:22-46`, bash `:49-51`, fish `:14-56` |
| `--init`/`--store`/`--link` → directory completion | COMPLETE | zsh `:directory:_files` on `--store`/`--init`/`--link` (`:34`,`:40`,`:42`); bash `compgen -d` (`:43`); fish `-r` (`:34`,`:39`,`:56`) |
| `--shell` → closed enum `bash zsh fish` — §14.2 | COMPLETE | zsh `:shell:(bash zsh fish)` (`:46`); bash `compgen -W "bash zsh fish"` (`:44`); fish `-x -a "bash zsh fish"` (`:62`) |
| `--search` → nothing (free-text) — §14.2 | COMPLETE | zsh routes `:query:` away from state; bash `return 0` (`:42`); fish no `-r` + global `-f` |
| §14.7 list-ambiguous on first Tab (emitted eval path) | COMPLETE | bash emitted: `{ [[ $- == *i* ]] && bind 'set show-all-if-ambiguous on'; } || true`; zsh emitted: `setopt NO_LIST_AMBIGUOUS`; fish: no-op (default lists) |
| Embedded scripts byte-identical to on-disk (bash/fish) — §14.6 | COMPLETE | `diff <(sd-test --completions --shell bash) completions/skilldozer.bash` empty; same for fish. zsh derived (self-call stripped, `compdef`+`compinit` appended) as PRD specifies |
| zsh emitted wrapper strips self-call + adds `compdef`/`compinit` — §14.6 | COMPLETE | on-disk has 1 self-call, emitted has 0; emitted tail has `autoload -Uz compinit` / `compdef _skilldozer skilldozer` |
| `TestEmbeddedCompletionsMatchOnDisk` lockstep test | COMPLETE | Passes (`go test -run TestEmbeddedCompletionsMatchOnDisk` ok) |

### Gaps (PARTIAL)

1. **§14.5 manual zsh copy path lacks the §14.7 enhancement.** The on-disk
   `completions/_skilldozer` contains **zero** `NO_LIST_AMBIGUOUS` references
   (verified by `grep -c`), because the `setopt` is injected only by
   `zshEvalScript` (main.go:1277) at the `--completions` eval path. A user who
   follows the documented manual path (`cp completions/_skilldozer
   ~/.zsh/completions/_skilldozer` + `compinit`) therefore does **not** get
   first-Tab listing on ambiguous prefixes. The PRD §14.7 text focuses on the
   "emitted `--completions` scripts", so this is arguably in-spec, but it is an
   asymmetry vs bash (whose on-disk file is verbatim and *does* carry the
   `bind` line). Severity: low (cosmetic UX asymmetry; eval path is the
   recommended one).

2. **`--link d1 <tab>` (multi-link, §14.1 r5) does not complete directories at
   every subsequent position.** PRD §14.1 rule 5 + the §14.1 matrix explicitly
   require "every position after `--link` completes dirs". Current behavior:
   - fish: only guard is `__fish_prev_arg_in --search` (line 68), so after
     `--link d1` the previous arg is `d1` (not `--search`), and skill tags are
     offered instead of directories.
   - bash: `case "$prev"` checks `--store|--init|--link` (line 43); after
     `--link d1` `prev=d1` falls through to tag completion.
   - zsh: `_arguments` treats `--link` as value-taking for one slot; subsequent
     positionals get tags.
   The single-arg case (`--link <dir>`) works correctly on all three shells;
   only the *additional* dirs in a multi-link batch mis-complete. Severity: low
   (multi-link is a power-user path; the single-arg case is the common one).

---

## 2. `install.sh` — **COMPLETE**

File: `install.sh` (lines 1-87). Audited against PRD §12.1 steps 1-7.

| §12.1 step | Status | Evidence |
|---|---|---|
| 1. `cd` to script's own dir | COMPLETE | `:21-22` `SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"; cd "$SCRIPT_DIR"` |
| 2. Verify `go` on PATH; else print install instructions + exit 1 | COMPLETE | `:26-32` prints go.dev URL + `exit 1` BEFORE building |
| 3. `go build -trimpath -ldflags "-s -w -X main.version=$(git describe --tags --always 2>/dev/null \|\| echo dev)"` | COMPLETE | `:36-37` byte-exact match to PRD |
| 4. Pick target bin dir: `$SKILLDOZER_INSTALL_BIN` → `~/.local/bin` → `/usr/local/bin` | COMPLETE | `:40-58`; order matches; no silent sudo (prints exact sudo command on failure) |
| 5. Symlink (not copy) `<target>/skilldozer` → `<repo>/skilldozer`; refresh existing | COMPLETE | `:60-66` `ln -sfn "$SCRIPT_DIR/skilldozer" "$TARGET/skilldozer"` (absolute target, `-n` guard) |
| 6. Ensure target on PATH; else print export line for detected shell | COMPLETE | `:68-83` detects bash/zsh/fish/other via `basename "$SHELL"`; prints only (no auto-edit) |
| 7. Verify command | COMPLETE | `:85-87` runs `"$TARGET/skilldozer" --version` via absolute symlink path (works pre-PATH-reload) |

Bonus: explicitly declines to install completions (`:8-11`, `:87`) per §14.5.
`set -euo pipefail` at top. All 7 steps match spec.

---

## 3. `README.md` — **COMPLETE**

File: `README.md`. Audited against PRD §15 outline (9 sections) + the three
required flag mentions.

| §15 section | Present | Evidence |
|---|---|---|
| 1. Title + one-liner | ✓ | "Skilldozer / Standalone skill loader for pi. Resolves a skill tag to an absolute path for `pi --skill`." |
| 2. Why | ✓ | "Why" section (not in any pi discovery location, on-demand only) |
| 3. Install (3 paths) + First run | ✓ | A. `./install.sh` (symlink), B. `go install`, C. from source; "First run" subsection with `--init` |
| 4. Usage | ✓ | canonical `pi --skill "$(skilldozer example)"`, multi-skill, `-f`, `--list`, `--search`, `--all`, `--check`, `--path`, `--link`, `--relative`, `--no-color`, `--version` |
| 5. Where skills live | ✓ | skills/ dir, tag = relative dir path, 5-step resolution order |
| 6. Adding a skill | ✓ | frontmatter template, field rules, `skilldozer --check` |
| 7. How skilldozer finds the store | ✓ | 5-rule priority (`SKILLDOZER_SKILLS_DIR` → config → sibling → walk-up → none) |
| 8. Shell completions | ✓ | `eval "$(skilldozer --completions)"`, fish `| source`, skills-first / long-form-only behavior, §14.7 disclosure |
| 9. Constraints | ✓ | no catalog index, never auto-discovered, loaded only via `--skill`, zero runtime deps |

Required flag mentions:
- **`--completions` eval**: COMPLETE — "Shell completions" section documents
  `eval "$(skilldozer --completions)"` and `skilldozer --completions --shell fish | source`.
- **`--link`**: COMPLETE — dedicated "Linking skills from elsewhere (`--link`)" section.
- **`--init`**: COMPLETE — "First run" section covers prompt + non-interactive
  `--store` forms, and documents `--store` implies `--init`.

§14.7 disclosure: COMPLETE — README names `show-all-if-ambiguous` (bash),
`NO_LIST_AMBIGUOUS` (zsh), and notes fish needs no option; provides opt-out
one-liners for bash and zsh.

---

## 4. `skills/example/SKILL.md` — **COMPLETE**

File: `skills/example/SKILL.md` (17 lines). Byte-compared against PRD §11
example. Frontmatter (`name: example`, `description`, `metadata.keywords`,
`metadata.category`) and body match the PRD spec exactly. Verified only one
skill ships (`find skills/ -type f` → `skills/example/SKILL.md` only).

---

## 5. `LICENSE` — **COMPLETE**

File: `LICENSE`. License: **MIT License**, "Copyright (c) 2026 Dustin Schultz".
Full standard MIT text (permission grant, no-warranty disclaimer). Consistent
with a standalone open-source CLI.

---

## 6. `.gitignore` — **COMPLETE**

File: `.gitignore` (5 lines). Byte-exact match to PRD §16:
```
/skilldozer
/dist
*.test
*.out
.DS_Store
```
(`cat -A` confirms no trailing whitespace/CRLF). The locally-built `skilldozer`
binary (built during this audit) is correctly ignored by `/skilldozer`.

---

## 7. `go.mod` — **COMPLETE**

File: `go.mod`.
- **Module path:** `github.com/dabstractor/skilldozer` — matches §12.2
  `go install github.com/dabstractor/skilldozer@latest`.
- **Go version:** `go 1.25` (PRD does not pin a version; 1.25 is current-era).
- **Single dependency:** `gopkg.in/yaml.v3 v3.0.1` (for config parsing, §8).
  No other deps — consistent with "zero runtime dependencies" (§15 §9) at
  build-time-only scope.

---

## 8. Binary behavior (built + exercised) — **COMPLETE**

Built with `go build -o /tmp/sd-test .` (clean, no warnings). Full test suite
`go test ./...` passes (all 8 packages ok, including
`TestEmbeddedCompletionsMatchOnDisk`).

| Command | Result | Notes |
|---|---|---|
| `/tmp/sd-test --version` | `skilldozer dev` | Plain `go build` reports `dev` as documented (§13); `./install.sh` would inject `git describe` |
| `/tmp/sd-test --help` | full USAGE/EXAMPLES/OPTIONS dump | Advertises long forms only; lists short aliases as "valid for typing but not advertised" |
| `/tmp/sd-test --completions --shell bash` (head -20) | verbatim `skilldozer.bash` | `diff` vs on-disk file: empty |
| `/tmp/sd-test --completions --shell fish` | verbatim `skilldozer.fish` | `diff` vs on-disk file: empty |
| `/tmp/sd-test --completions --shell zsh` | derived wrapper | self-call stripped, `compdef`+`compinit` appended, `setopt NO_LIST_AMBIGUOUS` injected |
| `/tmp/sd-test --completions --shell tcsh` | exit **2** | Unsupported shell rejected with usage exit (§6.4) |
| `env -u SHELL -u SKILLDOZER_SHELL /tmp/sd-test --completions` | exit **1**, stderr `could not detect shell; pass --shell {bash\|zsh\|fish}`, nothing on stdout | Matches §14.6 "None ⇒ stderr + exit 1" |

All observed binary behavior conforms to the PRD acceptance contract (§13
completion sub-checks: `completion-bash OK`, `completion-fish OK`,
`completion-no-shell OK`, `completion-bad-shell OK` all reproduce).

---

## Summary table

| Deliverable | Verdict | Severity of any gap |
|---|---|---|
| `completions/` | **PARTIAL** | low (2 gaps, both power-user paths) |
| `install.sh` | **COMPLETE** | — |
| `README.md` | **COMPLETE** | — |
| `skills/example/SKILL.md` | **COMPLETE** | — |
| `LICENSE` | **COMPLETE** (MIT) | — |
| `.gitignore` | **COMPLETE** | — |
| `go.mod` | **COMPLETE** | — |
| Binary (`--version`/`--help`/`--completions`) | **COMPLETE** | — |

## Residual risks / open questions

- **completion-gap-A (low):** Manual zsh §14.5 copy path (`cp _skilldozer
  fpath`) does not set `NO_LIST_AMBIGUOUS`; users on that path miss the §14.7
  first-Tab listing. Decision needed: accept asymmetry, or add a guarded
  `setopt NO_LIST_AMBIGUOUS` to the on-disk autoload file too (note: the
  autoload file is sourced inside the completion function context, so a plain
  `setopt` there would still be session-global — feasible).
- **completion-gap-B (low):** Multi-link (`--link d1 d2 <tab>`) offers skills
  instead of directories on all three shells, contradicting §14.1 r5. Fix
  requires a "has `--link` been seen anywhere in the line" guard per shell.
- **No observed defects** in `install.sh`, README, SKILL.md, LICENSE,
  `.gitignore`, or `go.mod`. The binary passes the relevant §13 acceptance
  checks for completions.
