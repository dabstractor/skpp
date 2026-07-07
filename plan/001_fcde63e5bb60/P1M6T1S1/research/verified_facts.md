# Verified Facts — P1.M6.T1.S1 (plan graph: P1.M6.T12.S1)

> The one shipped example skill: `skills/example/SKILL.md` with EXACTLY the PRD
> §11 content. Every fact below was verified empirically against the live codebase
> at commit time (repo: `/home/dustin/projects/skpp`, module
> `github.com/dabstractor/skpp`, go 1.25, yaml.v3).

## §1. The exact deliverable content (PRD §11, byte-verified via `cat -A`)

The PRD renders the file inside a **four-backtick** `````markdown` fence so the
inner **three-backtick** ```bash fence survives Markdown rendering. **The outer
four-backtick wrapper is NOT part of the file.** The actual `SKILL.md` content
starts at `---` and ends at the final ``` line (the closing of the inner bash
fence). Exact file content (569 bytes when written):

```markdown
---
name: example
description: >
  Reference example skill for skpp. Demonstrates the required frontmatter and
  how skpp resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skpp]
  category: meta
---

# Example Skill

This skill exists only so `skpp` has something to resolve.

Try:

```bash
skpp example                       # prints this directory's absolute path
skpp -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skpp example)"       # loads this skill into pi
```
```

The body's inner ```bash … ``` block is real file content (pi renders it on load).
The PRD's trailing sentence "No other skills ship in this repo." is PRD prose,
NOT part of the file.

## §2. The frontmatter is spec-compliant and parses correctly (verified)

Ran the actual `discover` package against a temp copy of the exact §11 content
(via the probe in this dir, `validate_example_probe.go`). Results:

| field          | parsed value                                                            | §9 verdict |
|----------------|-------------------------------------------------------------------------|------------|
| HasFM          | `true`                                                                  | OK         |
| name           | `example` — matches `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, len 7 ≤ 64       | OK         |
| description    | folded scalar `>` → 162 chars (≤ 1024); non-empty                        | OK (no WARN) |
| keywords       | `[example demo skpp]` (flow-sequence list; `toStringSlice` → []string)  | OK         |
| category       | `meta`                                                                  | OK         |
| aliases        | absent → `[]` (nil)                                                     | OK         |
| body present   | `true` (the `# Example Skill` heading + Try: block)                     | OK         |

`discover.ParseFrontmatter` (internal/discover/discover.go) uses yaml.v3 which
handles the `description: >` folded scalar and the `metadata:` map with list
values. `discover.BuildSkill` (skill.go) extracts keywords/category/aliases via
`toStringSlice` ([]any → []string) and a comma-ok `category` assertion. **No
changes to the discover package are required** — it already accepts this file.

external_deps.md §1 independently confirms: `name` lowercase-hyphen regex, 1024
description cap, `metadata` is a spec'd optional field, lists under metadata are
spec-compliant.

## §3. Verification commands — what works TODAY vs. what is pending

Built `/tmp/skpp-probe` from the current `main.go` and ran against a temp skills
dir (`SKPP_SKILLS_DIR=/tmp/skpp-verify/skills`):

| command (current main.go)                | result                                              | status |
|------------------------------------------|-----------------------------------------------------|--------|
| `skpp example`                           | `/tmp/skpp-verify/skills/example`, exit 0           | ✅ works (M3.T8.S1 landed) |
| `skpp -f example`                        | `/tmp/skpp-verify/skills/example/SKILL.md`, exit 0  | ✅ works (M3.T8.S2 landed) |
| `skpp --list`                            | TAG/NAME/DESCRIPTION row, exit 0                    | ✅ works (M2.T6.S1 landed) |
| `skpp check`                             | `unknown skill tag "check"`, exit 1                 | ❌ NOT wired yet |

**Why `skpp check` is not wired:** it depends on TWO not-yet-landed upstream
items: `internal/check` (P1.M4.T10.S1, status Planned) AND the `check`
subcommand dispatch in `main.go` (P1.M5.T1.S1, in parallel flight — currently
`main.go` has NO `runCheck`/`t10CheckDelegate`; `grep runCheck main.go` is empty,
and the P1.M5.T1.S1 PRP adds the dispatch with a placeholder that returns
exit 1 "not yet implemented" until T10 lands).

Per PRD §18 build order, M6.T12 (this task) is step 6; M4.T10 (check) is step 4
and M5.T1 is step 5 — so in proper milestone order BOTH land BEFORE this task.
**But the on-disk reality today is they have not.** The PRP therefore makes
`skpp check` a CONDITIONAL gate with a check-independent fallback (§4).

## §4. The check-independent fallback probe (verified working)

The probe `validate_example_probe.go` (this dir) imports the LANDED
`internal/discover` package and asserts every §9 rule directly against the file.
Guarded by `//go:build ignore` so it never compiles into the binary or breaks
`go build ./...` (verified: `go build ./...` is clean with the probe present).
Run from repo root:

```
go run plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go skills/example/SKILL.md
```

Expected output on the real file:
```
OK   example (example)
  HasFM=true keywords=[example demo skpp] category="meta" aliases=[]
  description chars=162 body_present=true
```
exit 0. This is equivalent to the `OK   example (example)` line `skpp check`
(§9) would emit, and works regardless of whether M4.T10/M5.T1 have landed.

## §5. `skills/.gitkeep` — NOT present (no-op removal)

The contract says "Remove the throwaway `skills/.gitkeep` from P1.M1.T3.S1 if
present." `git ls-files` shows **no `skills/` directory at all** and **no
`.gitkeep` anywhere**. So the removal is a **no-op** — the `skills/` tree is
created fresh by writing `skills/example/SKILL.md`. The PRP keeps the removal as
a defensive `rm -f` (harmless if absent).

## §6. Location & the "not a pi auto-discovery location" guardrail (PRD §2.2)

- The skills dir resolves (PRD §8 rule 2, sibling-of-binary) to `<repo>/skills`.
  `$PWD` = `/home/dustin/projects/skpp`, so the example lives at
  `/home/dustin/projects/skpp/skills/example/SKILL.md`.
- `~/.pi/agent/skills` EXISTS and is a real pi auto-discovery location (it
  contains `agent-browser`, `mdsel`, `write-pull-request`, …). Our `<repo>/skills`
  is a DIFFERENT, sibling-of-binary path — satisfies PRD §2 constraint 2
  (skills load ONLY via `pi --skill "$(skpp <tag>)"`, never auto-discovered).
- Assertable: `test "$(./skpp --path)" = "$PWD/skills"` and the printed path is
  NOT under `$HOME/.pi`.

## §7. PRD §11 sentence "No other skills ship in this repo" = prose, not a file

It appears in the PRD AFTER the closing ````` of the rendered block. It is the
author restating PRD §2 constraint 4 / §17 — it is NOT a line inside the example
SKILL.md. Do not write it into the file. (It is, however, a hard scope guardrail:
this task creates EXACTLY ONE skill.)

## §8. Validation commands verified working in this repo

- `go build -o skpp .` → exit 0 (verified).
- `go build ./...` → exit 0 with the probe present (verified).
- `go test ./...` → whole module green (verified baseline; this task adds NO Go
  code, so no test changes needed).
- `gofmt`/`go vet` are unaffected (no .go files touched by this task).
