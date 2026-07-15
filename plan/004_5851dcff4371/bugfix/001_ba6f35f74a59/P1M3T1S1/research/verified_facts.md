# Verified Facts — P1.M3.T1.S1 (Mode B sweep: README / help text / completion-header consistency)

Mode B changeset-level documentation sweep for bugfix round 2 (Issues 1-5).
All facts read directly from source at PRP-write time. Repo: `/home/dustin/projects/skilldozer`.
Plan: `004_5851dcff4371/bugfix/001_ba6f35f74a59/`.

## §0 — What this task IS (Mode B sweep) and the charter

- **IS**: the one changeset-level doc sync (decisions.md §D8): "A final 'Sync changeset-level
  documentation' task sweeps README.md and help text for cross-cutting consistency … The README's
  overall feature/coexistence story may need a consistency check after all issues land."
- The contract LOGIC lists 4 check categories (a-d). D8 additionally charters catching cross-cutting
  drift the per-issue Mode A edits (code + inline comments) did not carry into the README overview.
- **OUTPUT** (contract): "Any small doc fixes discovered during the sweep. If no inconsistencies are
  found, document that fact (the sweep is the deliverable — it confirms consistency)."

## §1 — All 5 issues LANDED (verified; the sweep documents these)

| # | Issue | Landed behavior (verified) | Code site |
|---|---|---|---|
| 1 | --shell value completion | all 3 completion files route --shell→`bash zsh fish` AND **advertise** --shell (D7) | completions/* + main.go embed |
| 2 | vanished configured store | config present + store dir gone ⇒ `configured skills store directory does not exist: configured store "…" does not exist; run \`skilldozer --init\` or recreate the directory`, exit 1, empty stdout | skillsdir.go findConfig→Find→ErrConfiguredStoreMissing |
| 3 | missing-value symmetry | `--store`/`--search`/`--shell` ALL exit 2 on missing value (built binary: all 3 → exit 2) | main.go searchMissingValue/shellMissingValue + run() |
| 4 | POSIX `--` separator | `skilldozer -- <tag>` treats `<tag>` as a literal positional | main.go parseArgs endOfOpts (186/206) |
| 5 | README version accuracy | README.md:136 = NEW wording (LANDED by parallel P1.M2.T3.S1) | README.md:136 |

## §2 — The sweep findings (per contract category a-d + D8 cross-cutting)

### (a) README Error contract (Issue 3) — **DRIFT → REQUIRED EDIT**
README.md:149 currently: "`--store` expects a value: `--init --store` with nothing after it exits 2
rather than guessing a store." After Issue 3, `--search` AND `--shell` ALSO exit 2 on missing value
(built binary confirmed: all three → exit 2). The contract (a) explicitly calls this out. →
GENERALIZE the sentence to cover all three value-taking flags.

### (a) POSIX `--` (Issue 4) — **OPTIONAL (pathological) → note the decision**
The README "Where skills live" section discusses tag addressing (the "no reserved tag names" note,
README.md:254-256). `skilldozer -- <tag>` now works (Issue 4). The contract says this is "optional
since such tags are pathological." → OPTIONAL one-clause mention in the "no reserved tag names"
paragraph (use `--` to address a dash-leading tag); lean OMIT if it clutters. Document the decision.

### (b) main.go usageText — **NO DRIFT → verify only**
main.go:110 `--completions [--shell <name>]   Emit the shell completion script for eval` and
main.go:111 `--shell <bash|zsh|fish>      Force a shell for completion`. --shell IS in OPTIONS and
--completions DOES mention `[--shell <name>]`. (Contract (b): "No change expected unless prior
subtasks introduced drift." None found.) → no edit.

### (c) completion-file LOCKSTEP headers — **NO DRIFT → verify only**
All three files' header comments now say: "--shell's value completes to the bash/zsh/fish enum
(§14.2); --shell is advertised (D7) since it is a real, documented flag in usageText OPTIONS."
(Issue 1's Mode A landed this.) → no edit.

### (d) README Install/version (Issue 5) — **ALREADY LANDED → verify only, do NOT duplicate**
README.md:136 already reads the NEW wording (parallel P1.M2.T3.S1 landed it): `# Version is the
git-describe value when built via ./install.sh; a plain 'go build' reports 'dev'`. → verify present;
do NOT re-edit (P1.M2.T3.S1's deliverable).

### D8 cross-cutting — README "Shell completions" section (Issue 1) — **DRIFT → REQUIRED EDIT**
Issue 1's Mode A updated the completion FILES (+ their headers) but NOT the README's user-facing
description of what completions show. TWO stale spots:
1. **Advertised flag list** (README.md:293-296): lists 13 long flags WITHOUT `--shell`. The completion
   files now advertise 14 (incl. `--shell`, per D7). → ADD `--shell` (alphabetical: between
   `--search` and `--store`).
2. **Value-completion enumeration** (README.md:298): lists `--init`/`--store` (directories) and
   `--search` (nothing) but NOT `--shell`. → ADD "`--shell <tab>` offers `bash`/`zsh`/`fish`".

### D8 cross-cutting — README "How skilldozer finds the store" section (Issue 2) — **DRIFT → RECOMMENDED EDIT**
README.md:244-246 rule 2: "A missing or unreadable config is treated as 'not yet configured' and
falls through to the rules below — never a hard error." After Issue 2, a PRESENT config whose `store:`
dir vanished is now a HARD ERROR (not fall-through). The current wording's "never a hard error" is
subtly incomplete. → RECOMMENDED clarifying clause (present config + vanished store ⇒ exit 1 with the
configured path named, not a silent fall-through to a different store).

## §3 — The exact edit anchors (current text, verified by grep)

```
README.md:149-150 (Error contract, Issue 3):
  the whole store). `--store` expects a value: `--init --store` with nothing after
  it exits 2 rather than guessing a store.

README.md:293-296 (Completions advertised flag list, Issue 1):
  - `skilldozer -<tab>` lists the **long-form flags only** — `--all`, `--check`,
    `--completions`, `--file`, `--help`, `--init`, `--list`, `--no-color`, `--path`,
    `--relative`, `--search`, `--store`, `--version` — narrowed by what you type
    after the dash. Short aliases (`-a`, `-l`, …) stay valid for typing but are
    deliberately not advertised.

README.md:298 (Completions value-completion enumeration, Issue 1):
  - `skilldozer --init <tab>` and `skilldozer --store <tab>` offer directories
    (the store to adopt); `skilldozer --search <tab>` offers nothing (free-text).

README.md:244-246 (Store section rule 2, Issue 2):
   A missing or unreadable config is treated as "not yet configured" and falls
   through to the rules below — never a hard error.

README.md:254-256 (Where skills live "no reserved tag names", Issue 4 — OPTIONAL):
  There are **no reserved tag names**: bare words are always skill tags, and every
  action is a `--flag` (§6.1). A skill named `check`, `init`, or `completions`
  resolves normally by its tag — use `--check`, `--init`, or `--completions` to run the action.
```

## §4 — README voice rules (verified)

- §-PRD citations are SPARSE: `grep -no '§[0-9.]*' README.md` → exactly ONE (`§6.1` at line 182).
  The edited sections (Error contract, Completions, Store) use ZERO § cites. → keep new prose
  citation-free (plain words), matching the local voice. (This README is NOT the citation-free one
  from a prior round — it tolerates a § cite, but the local sections don't use them.)
- Bold lead-ins / inline `` `code` `` / em dashes are the house style. Plain declarative prose.
- "Keep edits minimal" (Mode B). Match surrounding sentences' structure.

## §5 — Parallel-execution / file-conflict consideration

- P1.M2.T3.S1 (Issue 5, "Implementing") edits README.md line 136 (the version comment, inside the
  Usage ```bash block). It has ALREADY LANDED (grep confirms the new wording at :136).
- This task edits DIFFERENT regions (Error contract ~149, Completions ~293-298, Store ~244-246,
  Where-skills-live ~254-256) — all DISJOINT from line 136. Pin every seam by ANCHOR TEXT (§3).
- DO NOT touch line 136 (the version line) — that is P1.M2.T3.S1's deliverable; verify it present,
  do not re-edit. (D8: the sweep "confirms consistency," it does not duplicate Mode A edits.)

## §6 — Scope boundary / what NOT to touch

- main.go / main_test.go / internal/* / completions/* / install.sh — UNCHANGED (all 5 issues landed;
  this is doc-only). `go build/vet/test ./...` stays green (the doc-sanity check, contract LOGIC §3).
- PRD.md, tasks.json, prd_snapshot.md, .gitignore — READ-ONLY.
- README.md:136 (version line) — P1.M2.T3.S1's territory; verify only.

## §7 — Baseline (verified)

- `go build ./...` clean; `go test ./...` green.
- All three shell syntax-checkers pass on the (unchanged) completion files (not edited here).
- README §-cite count = 1 (line 182); version line = NEW wording at :136.
