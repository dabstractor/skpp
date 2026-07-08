# Verified Facts — P1.M3.T1.S1 (Rewrite skills/example/SKILL.md → PRD §11)

Single-file Markdown edit. All facts below were read directly from source at PRP-write time.
Repo: `/home/dustin/projects/skilldozer`.

## §1 — What is stale vs what is already correct

| Artifact | Location | Says | Status |
|---|---|---|---|
| **Repo asset** (THE edit target) | `skills/example/SKILL.md` | `skpp` (7 places) | ❌ STALE — this subtask fixes it |
| **Compiled-in seed constant** | `main.go:896` `const exampleSkillTemplate` | `skilldozer` (already) | ✅ CORRECT — do NOT touch |
| PRD §11 (canonical) | `PRD.md` h2.10 | `skilldozer` | ✅ source of truth |

**CRITICAL:** the `exampleSkillTemplate` constant in `main.go` is ALREADY PRD §11-compliant
(P1.M2.T2.S2 wrote it with the `skilldozer` wording). The contract's conditional — "If
P1.M2.T2.S2's compiled-in template was written before this fix, re-sync the constant" — is
FALSE here. **main.go is NOT edited by this subtask.** The two copies are currently OUT OF
SYNC (constant=skilldozer, asset=skpp); this subtask re-syncs them by editing ONLY the asset.

`grep -rn "skpp" --include="*.go" --include="*.md"` (whole repo) → the ONLY non-plan hit is
`skills/example/SKILL.md` (7 lines). main.go has ZERO `skpp`. Confirms the constant is clean.

## §2 — The 7 line-level changes (from architecture/docs_and_assets_drift.md §1 table)

Current `skills/example/SKILL.md` is 569 bytes, 20 lines. The 7 `skpp` → `skilldozer` swaps:

| Line | Current | Target (PRD §11 == rendered constant) |
|---|---|---|
| 4 | `  Reference example skill for skpp. Demonstrates the required frontmatter and` | `  Reference example skill for skilldozer. Demonstrates the required frontmatter and` |
| 5 | `  how skpp resolves a tag to an absolute path. Safe to delete once you add real skills.` | `  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.` |
| 7 | `  keywords: [example, demo, skpp]` | `  keywords: [example, demo, skilldozer]` |
| 13 | `` This skill exists only so `skpp` has something to resolve. `` | `` This skill exists only so `skilldozer` has something to resolve. `` |
| 18 | `` skpp example                       # prints this directory's absolute path `` | `` skilldozer example                       # prints this directory's absolute path `` |
| 19 | `` skpp -f example                    # prints .../skills/example/SKILL.md `` | `` skilldozer -f example                    # prints .../skills/example/SKILL.md `` |
| 20 | `` pi --skill "$(skpp example)"       # loads this skill into pi `` | `` pi --skill "$(skilldozer example)"       # loads this skill into pi `` |

Unchanged (already match §11): line 2 `name: example`, line 8 `category: meta`. The closing
fence (line 21) and trailing newline are also unchanged.

## §3 — Exact target content (byte-for-byte == rendered exampleSkillTemplate == PRD §11)

The constant splices 3 backtick runs (`+ "`skilldozer`" +`, `+ "```bash" +`, `+ "```" +`)
between 4 raw segments. Its RENDERED text (and thus the target file content) is:

```
---
name: example
description: >
  Reference example skill for skilldozer. Demonstrates the required frontmatter and
  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skilldozer]
  category: meta
---

# Example Skill

This skill exists only so `skilldozer` has something to resolve.

Try:

```bash
skilldozer example                       # prints this directory's absolute path
skilldozer -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skilldozer example)"       # loads this skill into pi
```
```

Ends with the closing ``` fence + a single trailing `\n` (verified: current file's last 4
bytes are `60 60 60 0a` = ```` ```\n ````; target preserves this). The file ends WITHOUT a
blank line after the fence (just `fence\n`).

NOTE on alignment inside the bash block: the three command lines keep their ORIGINAL column
alignment from §11 (the `#` comments are column-aligned across all three lines). The constant
preserves this exact spacing; transcribe verbatim — do not re-flow or re-align.

## §4 — Consumers (why these edits matter / what stays unaffected)

- **`skilldozer --search <q>`** (`internal/search/search.go`): `matches()` does a
  case-insensitive substring test over 6 fields incl. each `metadata.keywords` entry
  (matched INDIVIDUALLY, not joined). Current keyword `skpp` ⇒ `--search skilldozer` returns
  NO match while `--search skpp` matches — the INVERSE of intended (PRD §6.1/§10). Swapping
  the keyword to `skilldozer` FLIPS this: `--search skilldozer` matches; `--search skpp` no
  longer matches. The `description` field (lines 4-5) is ALSO searched, so the `skpp`→`skilldozer`
  swap there matters for search relevance too.
- **`skilldozer check`** (`internal/check/check.go`): validates frontmatter. `name: example`
  is UNCHANGED → check still reports `OK example (example)`. (Name drives the `(name)` column;
  tag `example` drives the `OK    example` column — both unaffected.)
- **`skilldozer example`** (resolve, `internal/resolve/`): resolves the tag to the directory by
  the DIRECTORY NAME `example` (RelTag), NOT by any frontmatter token. Dir name unchanged →
  resolution unaffected (`/abs/skills/example` still prints).
- **Frontmatter parsing** (`internal/discover/skill.go` `BuildSkill`): `Keywords =
  toStringSlice(fm.Metadata["keywords"])`. The YAML inline list `[example, demo, skilldozer]`
  parses to `[]any` → normalized to `[]string{"example","demo","skilldozer"}`. Identical
  structure to the current `[example, demo, skpp]`; only the last element's value changes.

## §5 — No test reads the repo asset (isolated change)

`grep -rn "Reference example skill\|has something to resolve\|demo, skilldozer\|demo, skpp"`
across `main_test.go` + `internal/` → only TWO hits, both SYNTHETIC unit-test fixtures that do
NOT read `skills/example/SKILL.md`:
- `internal/discover/skill_test.go:171` — a hand-written frontmatter string (its own skill).
- `internal/ui/ui_test.go:128` — a description literal passed to a ui helper.

Neither touches the repo asset. So `go test ./...` is UNAFFECTED by this edit. No test asserts
asset==constant today (there is no such regression guard; adding one would touch
`main_test.go`, which the parallel S3 PRP edits — AVOID to prevent conflict).

## §6 — Baseline CLI behavior (observed at PRP-write time, env `SKILLDOZER_SKILLS_DIR=$PWD/skills`)

| Command | Current output | Current exit |
|---|---|---|
| `./skilldozer check` | `OK    example (example)` / `1 skills, 0 errors, 0 warnings` | 0 |
| `./skilldozer --search skilldozer` | `no skills matched skilldozer` | 0 |
| `./skilldozer --search skpp` | table: `example  example  Reference example skill for skpp. ...` | 0 |
| `./skilldozer example` | `/home/dustin/projects/skilldozer/skills/example` | 0 |

EXIT-CODE NUANCE (observed, do NOT assert on it): `--search <nomatch>` currently exits 0 with
the "no skills matched" line (PRD §6.1 says exit 1 on no matches, but the live binary returns
0 in the "store found, query matched nothing" case — that is a SEPARATE concern, out of scope
here). Validation gates must assert on STDOUT CONTENT (does the example row appear?), NOT on
the exit code, to be robust.

## §7 — Scope boundary (what NOT to touch)

- `main.go` — the `exampleSkillTemplate` constant is already §11-correct. Do NOT edit.
- `main_test.go` — owned by the PARALLEL P1.M2.T2.S3 PRP (adds run-level init tests). Do NOT
  edit here (merge-conflict risk). No regression test added by this subtask.
- `PRD.md`, `README.md`, `completions/*`, `install.sh` — out of scope (README = P1.M4.T2.S1;
  completions = P1.M3.T2.S1). This subtask edits EXACTLY ONE file: `skills/example/SKILL.md`.
- The `plan/` archive still contains `skpp` (historical, PRD §19 #15) — do NOT touch.

## §8 — Relationship to the parallel P1.M2.T2.S3 (run() init dispatch)

S3's `runInit` calls `setupStore` → which seeds `example/SKILL.md` in a freshly-created store
from the `exampleSkillTemplate` CONSTANT (not the repo asset). Because the constant is already
correct, S3's seeded stores are ALREADY §11-compliant regardless of this subtask. This
subtask fixes the separate on-disk REPO asset. The two are independent copies; S3's tests are
unaffected by this edit. (S3 is "currently being implemented"; treat its PRP as a contract.)
