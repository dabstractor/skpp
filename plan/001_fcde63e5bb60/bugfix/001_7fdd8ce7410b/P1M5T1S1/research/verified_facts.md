# Verified facts — P1.M5.T1.S1 (bugfix): Trim `.gitignore` to PRD §16 spec

> All facts read directly from the working tree at `/home/dustin/projects/skpp`
> on 2026-07-07. This is a spec-alignment change to a single non-code file.

## 1. PRD §16 — the authoritative target (PRD.md line 376)

PRD §16 specifies the `.gitignore` contents EXACTLY as a 5-line code block, with
no comments and no section headers:

```
/skpp
/dist
*.test
*.out
.DS_Store
```

Accompanying prose (verbatim): "(`/skpp` ignores the locally-built binary;
everything else is committed, including `skills/example/`.)"

PRD.md is **read-only / human-owned** (FORBIDDEN list + decision D3). So the
implementation conforms to §16; it does NOT edit the PRD.

## 2. Current `.gitignore` (exact bytes, `cat -A` verified)

```
# Build output
/skpp
/dist
/build

# Test / coverage artifacts
*.test
*.out

# Environment files
.env
.env.*

# OS files
.DS_Store

# Tool scratch (pi-subagent run artifacts; regenerated per run)
.pi-subagents/
```

= **9 entries** across **4 commented sections** with blank-line separators.

## 3. The diff (what the implementer does)

KEEP (5): `/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store`.
REMOVE (4): `/build`, `.env`, `.env.*`, `.pi-subagents/`.
ALSO REMOVE: every `# …` comment line and every blank separator line.
RESULT: exactly the 5 lines from §16, each terminated by `\n`, file ends after
the `.DS_Store\n` line (single trailing newline — matches the §16 block shape
and the original file's trailing-newline convention).

## 4. Decision D3 (architecture/decisions.md line 37) — verbatim rationale

> **Decision**: Remove `/build`, `.env`, `.env.*`, `.pi-subagents/`. Restore the
> exact §16 set.
> **Rationale**: PRD.md is read-only (human-owned). The spec is explicit about
> the 5-entry set. The extras are "reasonable hygiene" but represent undocumented
> deviation. Bringing the file into spec compliance is the only action that
> resolves the discrepancy without modifying PRD.md. If maintainers want the
> extras, they update §16 themselves.

## 5. Issue 3 (architecture/issue_analysis.md) — impact

> **Test impact**: None (no code).

So no Go test changes; `go test ./...` is a *collateral-sanity* gate only, not a
behavioral one. The real verification is byte-diffing `.gitignore` against §16.

## 6. The two gotchas (the only things an implementer can get wrong)

1. **Do NOT keep the comments / blank lines.** The natural temptation (and the
   "reasonable hygiene" framing in the issue) is to keep the helpful section
   comments and just delete the 4 extra entries. That would STILL deviate from §16
   (§16 has zero comments). The whole file must become the bare 5-line block.
2. **`.pi-subagents/` artifacts will appear as untracked in `git status` after the
   change.** This is EXPECTED and CORRECT — the item and D3 both say the spec does
   not ignore them. Do NOT "fix" this by re-adding the entry, and do NOT delete
   the artifacts (they are live pi-subagent run outputs).

## 7. Parallel-context check (no conflict)

P1.M4.T2.S1 (in flight) extends `exclusivityError` in `main.go` for mode+mode
combos. It does **not** touch `.gitignore`. The two files are disjoint → zero
merge/race risk. `grep gitignore plan/…/P1M4T2S1/PRP.md` = no hits.

## 8. Verification commands (all verified present in this tree)

```bash
cd /home/dustin/projects/skpp
cat .gitignore                       # must show exactly the 5 §16 lines
git diff --stat .gitignore           # 9 entries → 5 entries
go test ./...                        # collateral sanity (no code changed) → still green
```
No `go.mod`/`go.sum` change. No dependency involved. This is a one-file edit.
