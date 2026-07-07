# PRP — P1.M5.T1.S1 (bugfix): Trim `.gitignore` to the PRD §16 spec set

> **Subtask:** P1.M5.T1.S1 — Spec Alignment & Documentation (bugfix milestone M5).
> Brings `.gitignore` into literal compliance with **PRD §16** by overwriting it
> with exactly the 5 entries the spec lists (no comments, no section headers),
> removing the 4 extra entries (`/build`, `.env`, `.env.*`, `.pi-subagents/`).
>
> **Scope:** ONE non-code file — overwrite `.gitignore`. No Go source, no tests,
> no docs (`.gitignore` is itself the spec artifact — Mode A). `PRD.md` is
> read-only/human-owned and is NOT modified (decision D3).
>
> **DEPENDENCY (PARALLEL CONTEXT):** P1.M4.T2.S1 (in flight) extends
> `exclusivityError` in `main.go` for mode+mode combos. It does **not** touch
> `.gitignore` (verified: `grep gitignore …/P1M4T2S1/PRP.md` = no hits). The two
> files are disjoint → zero conflict. This PRP can land independently, in any
> order relative to P1.M4.T2.S1.

---

## Goal

**Feature Goal**: `.gitignore` matches PRD §16 byte-for-byte: exactly the 5-entry
block the spec specifies (`/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store`), with
no comments, no section headers, and no blank separators — resolving the
"reasonable hygiene but undocumented deviation" flagged in Issue 3.

**Deliverable**: A single edited file, `.gitignore`, containing exactly:

```
/skpp
/dist
*.test
*.out
.DS_Store
```
(each line terminated by `\n`; a single trailing newline after `.DS_Store`).

**Success Definition**: `cat .gitignore` prints exactly the 5 §16 lines in order;
`diff <(awk '/^```$/{c++;next} c==1' PRD.md) .gitignore` is empty (the §16 block
matches the file). `go test ./...` remains green (collateral sanity — no code
changed). `git status` may show `.pi-subagents/` artifacts as untracked; that is
**expected and correct** (the spec does not ignore them).

## User Persona

**Target User**: Repository maintainers and contributors (and the QA process that
filed Issue 3) who expect the shipped files to match the PRD's literal spec.

**Use Case**: A reviewer auditing the repo against PRD §16 runs `cat .gitignore`
and sees exactly the spec contents — no surprise extras to reconcile.

**Pain Points Addressed**: The current 9-entry/4-section `.gitignore` deviates
from §16's explicit 5-entry block. Since PRD.md cannot be edited by the
implementation, the only way to close the discrepancy is to bring the file to
spec (decision D3: "do NOT bless extras").

## Why

- **Spec/impl alignment.** §16 is the authoritative `.gitignore` spec and it is
  literal. The 4 extras (`/build`, `.env`, `.env.*`, `.pi-subagents/`) are useful
  hygiene but undocumented; keeping them leaves a permanent "why does this differ
  from the PRD?" question for every future reader.
- **PRD.md is read-only.** The only resolution that does not require a human to
  edit the PRD is to conform the file to the spec (D3). If maintainers later want
  the extras, they update §16 themselves.
- **Zero risk.** No code, no tests, no behavior change. The only observable
  side-effect (`.pi-subagents/` showing as untracked) is explicitly accepted by
  the spec and the item.

## What

Overwrite `.gitignore` with the exact PRD §16 contents: the 5 lines
`/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store`, each on its own line, in that
order, with a single trailing newline. Remove all `# …` comment lines, all blank
separator lines, and the 4 extra entries (`/build`, `.env`, `.env.*`,
`.pi-subagents/`).

### Success Criteria

- [ ] `cat .gitignore` shows exactly 5 lines: `/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store` (in that order).
- [ ] No comment lines (`# …`), no blank lines remain.
- [ ] The file ends with a single trailing newline after `.DS_Store`.
- [ ] `.gitignore` is byte-identical to the PRD §16 code block.
- [ ] `PRD.md` is NOT modified.
- [ ] `go test ./...` still green (collateral sanity; no code touched).

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ **Yes.** The exact target content (5 lines) is
given verbatim above and in PRD §16; the exact current content (9 entries) is in
`research/verified_facts.md` §2; the diff is enumerated in §3. There is no logic,
no API, no dependency — a single file overwrite with a byte-level verification
command. The only failure modes are documented as the two gotchas (§6 of the
research notes).

### Documentation & References

```yaml
# MUST READ — the authoritative spec (read-only; do NOT edit)
- file: PRD.md
  section: "§16 (`.gitignore`, line 376)"
  why: "Defines the exact 5-entry block this task must reproduce byte-for-byte."
  critical: "§16 lists the 5 entries in a bare code block with NO comments and NO
             section headers. The shipped file must match that shape, not just the
             entry set — preserving the existing '# Build output' / '# Test / coverage'
             comments would STILL deviate from §16."

# MUST READ — the decision that pins the approach
- docfile: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/decisions.md
  section: "D3 — Issue 3: Trim .gitignore to PRD §16 spec (do NOT bless extras)"
  why: "Locks the decision: REMOVE the 4 extras, do NOT bless them; PRD.md is
        read-only so the file must conform to the spec."
  critical: "If maintainers want the extras, they update §16 themselves — the
             implementer must NOT keep the extras and must NOT edit PRD.md."

# MUST READ — the issue being fixed (impact = none on code/tests)
- docfile: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/issue_analysis.md
  section: "Issue 3 (MINOR) — .gitignore deviates from PRD §16"
  why: "Names the 4 extras and states 'Test impact: None (no code).' Confirms this
        is a one-file spec-alignment change."

# MUST READ — this task's own empirical verification (current bytes + gotchas)
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/P1M5T1S1/research/verified_facts.md
  why: "§2 the exact current .gitignore; §3 the keep/remove diff; §6 the two
        gotchas (no comments; .pi-subagents becomes untracked = expected); §8 the
        exact verification commands."
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── .gitignore        # MODIFY — trim 9 entries (4 sections) → 5 entries (§16 spec)
├── PRD.md            # READ-ONLY (§16 is the spec; do NOT edit)
└── … (main.go, internal/*, etc. — UNCHANGED; no conflict with P1.M4.T2.S1's main.go edit)
```

### Desired Codebase tree (file responsibility)

```bash
.gitignore   # Exactly the PRD §16 5-entry block; ignores the built binary + test/coverage/OS artifacts.
```
No new files. No code. No tests. No docs (Mode A).

### Known Gotchas of our codebase & Library Quirks

```bash
# CRITICAL — do NOT keep the comments. The current file has 4 '# …' section
# headers and blank separators. §16 has NONE. The result must be the BARE 5-line
# block. Keeping the comments "for hygiene" would still fail the §16 byte-match.

# CRITICAL — the trailing newline. Write the file as 5 lines each ending in '\n'
# (so the file is "…\n.DS_Store\n", one trailing newline). Matches the §16 block
# shape and the original file's convention. Avoid a missing final newline (some
# tools append a "No newline at end of file" diff) and avoid a double blank line.

# GOTCHA — after removing `.pi-subagents/`, `git status` will show
# .pi-subagents/ artifacts as UNTRACKED. This is EXPECTED and CORRECT (the spec
# does not ignore them; D3 + the item both say so). Do NOT re-add the entry, and
# do NOT delete the artifacts (they are live pi-subagent run outputs).

# GOTCHA — do NOT touch PRD.md. §16 is the source of truth and is human-owned.
# If you disagree with the spec, that is a human decision to update §16, not an
# implementation action.

# NO Go / dependency involvement. `go mod tidy`, `gofmt`, etc. are no-ops here.
```

## Implementation Blueprint

### Data models and structure

None. This is a static-text file overwrite — no models, no types, no code.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: OVERWRITE .gitignore with the exact PRD §16 block
  - WRITE: .gitignore with EXACTLY these 5 lines, each terminated by '\n',
           nothing else (no comments, no blank lines):
             /skpp
             /dist
             *.test
             *.out
             .DS_Store
  - SOURCE OF TRUTH: PRD.md §16 (line 376). Copy the block verbatim.
  - REMOVE (vs current): /build, .env, .env.*, .pi-subagents/ AND every '# …'
           comment line AND every blank separator line.
  - GOTCHA: result is the BARE 5-line block (§16 has no comments). Do not keep the
           '# Build output' / '# Test / coverage artifacts' / '# Environment files'
           / '# OS files' / '# Tool scratch' headers.

Task 2: VERIFY byte-for-byte against §16 (the real gate — no Go tests apply)
  - COMMAND: cat .gitignore
  - EXPECT: exactly 5 lines in order: /skpp, /dist, *.test, *.out, .DS_Store.
  - COMMAND: diff <(sed -n '/^## 16/,/^---/{//!p;}' PRD.md | sed '/^```$/d') .gitignore
  - EXPECT: empty diff (the §16 block == the file). See Validation Loop Level 2
            for the canonical extraction command.
  - COMMAND: git diff --stat .gitignore
  - EXPECT: the file shows as modified (9 entries → 5 entries).

Task 3: COLLATERAL SANITY — confirm no code/tests/PRD were disturbed
  - COMMAND: git diff --name-only | grep -vE '^\.gitignore$' | (! read)
  - EXPECT: only .gitignore changed (no other file touched).
  - COMMAND: git diff --quiet PRD.md && echo "PRD.md untouched OK"
  - COMMAND: go test ./...   # no code changed → still green
  - EXPECT: all packages pass (sanity only; this task changes no behavior).
```

### Implementation Patterns & Key Details

```bash
# The entire change is a single file overwrite. The only "pattern" is fidelity:
# reproduce §16 verbatim, preserve order, single trailing newline.

# Canonical target content (write this, exactly):
cat > .gitignore <<'EOF'
/skpp
/dist
*.test
*.out
.DS_Store
EOF
# (The heredoc above produces exactly the 5 lines + one trailing newline.)
```

### Integration Points

```yaml
VERSION CONTROL:
  - file: .gitignore
  - before: 9 entries across 4 commented sections (+ blank separators)
  - after:  exactly the PRD §16 5-entry block, bare, single trailing newline
  - side-effect: .pi-subagents/ artifacts become untracked (EXPECTED, per spec)

NO OTHER INTEGRATION POINTS:
  - no database, no config, no routes, no build, no dependency.
  - PRD.md is read-only (do NOT edit §16 to bless the extras — D3).
  - no conflict with P1.M4.T2.S1 (it edits main.go; this edits .gitignore).
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
# .gitignore is plain text; the only "style" rule is §16's bare shape.
grep -nE '^#|^$' .gitignore && echo "FAIL: comments/blank lines remain" || echo "bare-block OK"
# Expected: "bare-block OK" (no comment lines, no blank lines).
```

### Level 2: Byte-for-byte spec match (THE gate — no Go tests apply)

```bash
cd /home/dustin/projects/skpp

# 1) Contents are exactly the 5 §16 lines in order.
cat .gitignore
# Expected:
#   /skpp
#   /dist
#   *.test
#   *.out
#   .DS_Store

# 2) Extract the §16 block from PRD.md and diff against the file (empty == pass).
#    §16 is the fenced code block immediately under "## 16. `.gitignore`".
spec=$(awk '/^## 16\. `.gitignore`/{f=1;next} /^```/{if(f){c++}; if(c>=2){exit}} f&&c>=1&&!/^```$/{print}' PRD.md)
diff <(printf '%s\n' "$spec") .gitignore && echo "§16 byte-match OK" || echo "FAIL: .gitignore != §16"

# 3) Exactly one trailing newline (no missing/double final newline).
[ "$(tail -c1 .gitignore | xxd -p)" = "0a" ] && echo "single-trailing-newline OK"

# 4) Git sees the trim (9 → 5).
git diff --stat .gitignore
# Expected: 1 file changed; the removed lines (comments, blanks, /build, .env, .env.*, .pi-subagents/).
```

### Level 3: Integration / repo-state validation

```bash
cd /home/dustin/projects/skpp

# Only .gitignore changed; nothing else disturbed.
git diff --name-only | grep -vE '^\.gitignore$' | (! read) && echo "only-.gitignore-touched OK"

# PRD.md is byte-identical (read-only; NOT modified).
git diff --quiet PRD.md && echo "PRD.md untouched OK"

# Collateral sanity: the Go module is unaffected (no code changed).
go test ./... && echo "go tests still green (sanity) OK"

# Expected side-effect: .pi-subagents/ now shows as untracked. Confirm it is NOT
# ignored anymore (correct per spec) — do NOT re-add it, do NOT delete artifacts.
git status --short | grep -q '.pi-subagents' && echo ".pi-subagents untracked (expected per §16) OK" || true
```

### Level 4: Creative & Domain-Specific Validation

```bash
cd /home/dustin/projects/skpp
# Spec-alignment tasks have no domain-specific validation beyond the §16 match.
# Confirm the example skill is still tracked (the one §16 invariant that matters):
git ls-files --error-unmatch skills/example/SKILL.md >/dev/null 2>&1 && echo "example skill still tracked OK"
# Expected: the shipped example skill remains committed (§16: "everything else is
# committed, including skills/example/").
```

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — no comment lines, no blank lines in `.gitignore`.
- [ ] Level 2 PASS — `cat .gitignore` shows exactly the 5 §16 lines; `diff` vs the §16 block is empty; single trailing newline.
- [ ] Level 3 PASS — only `.gitignore` changed; `PRD.md` untouched; `go test ./...` green (sanity).
- [ ] Level 4 PASS — `skills/example/SKILL.md` still tracked.

### Feature Validation
- [ ] `.gitignore` contains exactly `/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store` (in order).
- [ ] The 4 extras (`/build`, `.env`, `.env.*`, `.pi-subagents/`) are removed.
- [ ] All `# …` comments and blank separators removed (bare block, per §16).
- [ ] `.pi-subagents/` artifacts appear untracked (accepted, not "fixed").

### Code Quality Validation
- [ ] No source code, tests, go.mod/go.sum, PRD.md, or docs changed.
- [ ] No conflict with the parallel P1.M4.T2.S1 (main.go) edit.

### Documentation & Deployment
- [ ] Mode A: no docs change — `.gitignore` is itself the spec artifact.

---

## Anti-Patterns to Avoid

- ❌ Don't keep the `# Build output` / `# Test / coverage` / `# Environment files` /
  `# OS files` / `# Tool scratch` comments — §16 has none; keeping them still fails
  the byte-match.
- ❌ Don't "bless" the extras by editing PRD §16 — PRD.md is read-only (D3); conform
  the file, not the spec.
- ❌ Don't re-add `.pi-subagents/` after seeing it untracked in `git status` — that
  is the expected, spec-correct outcome.
- ❌ Don't delete the `.pi-subagents/` artifacts either — they are live run outputs.
- ❌ Don't touch any other file (main.go is owned by P1.M4.T2.S1; PRD.md is human-owned).
- ❌ Don't add a missing or double trailing newline — exactly one `\n` after `.DS_Store`.
