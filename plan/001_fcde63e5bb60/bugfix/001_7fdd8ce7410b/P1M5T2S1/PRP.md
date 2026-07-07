# PRP — P1.M5.T2.S1 (bugfix): Confirm `check.go` "no SKILL.md" omission rationale

> **Subtask:** P1.M5.T2.S1 — Spec Alignment & Documentation (bugfix milestone M5,
> Issue 7). A **verification gate with a one-paragraph clarity refinement**: confirm
> that the `Check` doc-comment in `internal/check/check.go` accurately documents
> why the PRD §9 rule "ERROR: skill dir has no SKILL.md" is not implemented, and
> close the one gap found (contract point c — the reframe is not stated in the
> 123-129 block).
>
> **Scope:** ONE file — `internal/check/check.go`, comment lines ~123-129 only.
> **Comment-only. No code. No behavioral change. No new tests. No other files.**
> `PRD.md` is read-only/human-owned and is NOT modified (decision D7). Mode A:
> the comment IS the documentation for this decision.
>
> **DEPENDENCY (PARALLEL CONTEXT):** P1.M5.T1.S1 (in flight) overwrites
> `.gitignore` only — it does **not** touch `check.go` (verified:
> `grep -c check.go …/P1M5T1S1/PRP.md` = 0). No other in-flight/ prior item
> touches `internal/check/check.go` (P1.M4.T2.S1→`main.go`, P1.M4.T1.S1→`main.go`,
> P1.M3.T1.S1→`ui.go`, P1.M2.T1.S1→`search.go`, P1.M1.T1.S1→`main.go`/`skillsdir.go`).
> ⇒ **Zero conflict. This PRP lands independently, in any order.**

---

## Goal

**Feature Goal**: The `Check` doc-comment in `internal/check/check.go`
(lines 123-129) is **self-contained and accurate** on all three contract points
about the omitted PRD §9 "skill dir has no SKILL.md" rule: (a) *why it can never
fire* (`discover.Index` only emits dirs that contain a `SKILL.md`), (b) *why no
heuristic* (false-positives on legitimate grouping dirs), and (c) *that it is
reframed* below as the `"invalid SKILL.md frontmatter"` ERROR.

**Deliverable**: A single edited doc-comment in `internal/check/check.go`
(the block immediately preceding `func Check`). Points (a) and (b) are already
accurate (no change); point (c) is added by a one-paragraph clarity refinement so
the block is complete on all three. No executable code changes.

**Success Definition**:
- The 123-129 comment states (a), (b), AND (c) (the reframe cross-reference).
- `gofmt` and `go vet` are clean on the package.
- `go test ./internal/check/` is green (incl. `TestCheckMalformedYAML`, which
  already pins the reframed ERROR string).
- The §13 acceptance gate passes: `go build -o skpp .` then `./skpp check` prints
  `OK    example (example)` and exits 0.
- No file other than `internal/check/check.go` is modified; `PRD.md` is untouched.

## User Persona

**Target User**: Future maintainers reading `check.go` (and the QA process that
filed Issue 7) who need to understand, from the comment alone, why §9's
"skill dir has no SKILL.md" ERROR is absent — and what `check` does instead.

**Use Case**: A maintainer lands on the `Check` doc-comment and reads the complete
rationale (omit → why → why-no-heuristic → the reframe) without having to hunt
through the switch branches to discover that the rule was reframed.

**Pain Points Addressed**: Today the reframe is documented only inside the
`perr != nil` branch (lines 146-149); the 123-129 rationale block omits it, so
the comment looks like "we just dropped the rule" rather than "we dropped the
unreachable form and reframed the reachable one."

## Why

- **Decision D7 is "documentation-only."** The bug report says "None required for
  correctness"; a heuristic would false-positive on legitimate grouping dirs.
  The only action D7 allows is confirming the comment is accurate.
- **Point (c) is the real gap.** Points (a) and (b) are already accurate; the
  reframe (c) exists but is disconnected from the rationale block the contract
  points at (123-129). A one-sentence-per-line refinement ties them together.
- **Zero risk.** Comment-only. No behavioral change. Tests + §13 gate are the
  safety net that proves nothing moved.

## What

Edit **only** the doc-comment block in `internal/check/check.go` that currently
spans lines 124-129 (the lines after the blank `//` separator at 123 and before
`func Check` at 130). Concretely:

1. **Keep** the (a) and (b) content (it is accurate) — reworded only enough to
   thread in the (c) sentence.
2. **Add** one statement that the §9 rule's actionable form is REFRAMED below as
   the `"invalid SKILL.md frontmatter"` ERROR in the `perr != nil` branch (so a
   reader of the block learns the reframe without leaving it).
3. **Preserve** the existing trailing sentence about the §9 "empty besides
   SKILL.md" WARN (that is a separate, correct §9-omission note — leave it).

Do NOT touch: the `perr != nil` branch (lines 145-149), the `!fm.HasFM` branch
(line 150+), `checkFields`, `appendDupFindings`, any ERROR string, any test, any
other file, or `PRD.md`.

### Success Criteria

- [ ] The 123-129 comment names the reframe: that "invalid SKILL.md frontmatter"
      (the `perr != nil` branch) is the reframed §9 rule.
- [ ] The comment still explains (a) Index-only-emits-SKILL.md-dirs and (b)
      heuristic-false-positives.
- [ ] The "empty besides SKILL.md" WARN note is retained.
- [ ] `gofmt -l internal/check/check.go` prints nothing.
- [ ] `go vet ./internal/check/` is clean.
- [ ] `go test ./internal/check/` passes (incl. `TestCheckMalformedYAML`).
- [ ] `./skpp check` reports `OK    example (example)`, exit 0.
- [ ] Only `internal/check/check.go` changed; `PRD.md` byte-identical.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ **Yes.** The exact current comment (lines
123-129) is quoted verbatim in `research/verification_findings.md` §2; the exact
proposed replacement is given in the Implementation Blueprint below; the three
contract points are verified point-by-point in §3 of the research note with
line-anchored evidence (index.go, PRD §9, D7, issue_analysis §Issue 7). The task
is a single comment edit plus four read-only validation commands. The only
non-obvious hazard — the contract's "check.go:150" citation actually points at
line 149 for the ERROR and line 150 is the *next* branch — is documented in §3(c).

### Documentation & References

```yaml
# MUST READ — the file under edit (comment block lines 123-129, reframe at 145-149)
- file: internal/check/check.go
  why: "The ONLY file this task edits. 123-129 = the rationale comment (the
        verification target); 145-149 = the perr branch that holds the reframe
        (DO NOT EDIT — already correct and tested)."
  pattern: "Go doc-comment preceding func Check; one // per line; §-citations
            like (research §2) refer to the check package's own research notes
            from when check.go was authored — leave them as citations."
  gotcha: "Contract cites 'check.go:150' for the reframe, but line 150 is
           `case !fm.HasFM:` (the NEXT branch). The reframe prose is lines
           147-148; the ERROR Finding is line 149. Read 145-149, not 150."

# MUST READ — the decision that pins this as documentation-only
- docfile: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/decisions.md
  section: "D7 — Issue 7: No code change; documentation-only"
  why: "Locks the approach: do NOT implement a 'no SKILL.md' heuristic; confirm
        the comment is accurate. Bounds this task to comment-only."
  critical: "D7 explicitly forbids adding a heuristic. The ONLY allowed change
             is clarifying the existing rationale comment. No code."

# MUST READ — the issue being closed (impact = none on code/tests)
- docfile: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/issue_analysis.md
  section: "Issue 7 (MINOR) — check cannot report 'skill dir has no SKILL.md'"
  why: "States root cause (index.go:46), the deliberate-omission rationale, and
        that test impact is 'None'. Confirms the reframe lives at check.go:150
        (sic — see line-drift gotcha above; actual = 149)."

# MUST READ — the source of truth for the §9 rule + §13 gate (READ-ONLY)
- file: PRD.md
  section: "§9 line 202 (the rule), §9 line 207 (the WARN), §9 line 148 (skill
            def), §13 line 334 (`./skpp check` reports example as OK), §13 line
            317 (`go build -o skpp .`)."
  why: "§9 line 202 is the rule being reframed; line 148 corroborates (a);
        line 334 is the acceptance gate this task must not break. Do NOT edit
        PRD.md."

# MUST READ — this task's own line-anchored verification (a/b/c verdict)
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/P1M5T2S1/research/verification_findings.md
  why: "§3 the point-by-point verdict (a✅ b✅ c⚠️gap); §2 the verbatim current
        comment; §3(c) the line-drift gotcha; §5 the validation commands (all
        PASS in current tree); §6 the conflict check (none)."
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── internal/check/
│   ├── check.go        # EDIT doc-comment lines ~123-129 ONLY (perr branch 145-149 untouched)
│   └── check_test.go   # UNCHANGED (TestCheckMalformedYAML already pins the reframe ERROR)
├── PRD.md              # READ-ONLY (§9/§13 are the spec; do NOT edit)
└── … (main.go, internal/*, .gitignore, etc. — UNCHANGED; no conflict)
```

### Desired Codebase tree (file responsibility)

```bash
internal/check/check.go   # Doc-comment block is self-contained on contract points a, b, AND c.
```
No new files. No code. No tests. No other docs (Mode A — the comment IS the doc).

### Known Gotchas of our codebase & Library Quirks

```go
// CRITICAL — comment-only. Do NOT change any executable line, any ERROR string,
// any test, or any other file. D7 bounds this to documentation.

// CRITICAL — line drift in the contract. "check.go:150" (cited in the item +
// issue_analysis) is the NEXT branch (case !fm.HasFM:). The reframe is at
// 145-149 (prose on 147-148, the "invalid SKILL.md frontmatter" Finding on 149).
// Anchor the refinement to "the perr != nil branch", not to a literal line 150.

// GOTCHA — the reframe branch is already tested. check_test.go:77
// TestCheckMalformedYAML asserts a malformed-YAML skill yields exactly one
// "invalid SKILL.md frontmatter" ERROR. If you (mistakenly) changed that ERROR
// string, this test would fail. Leaving the branch alone keeps it green.

// GOTCHA — §-citations like "(research §2)" / "(research §3)" in the comment
// refer to the check package's OWN research notes (written when check.go was
// authored in P1.M2.T10). They are not this task's research. Preserve them.

// NO dependency / build-config involvement. go.mod/go.sum are untouched.
// `gofmt`/`go vet` are the only "compilers" that touch this change.
```

## Implementation Blueprint

### Data models and structure

None. Comment-only edit — no models, no types, no code.

### Verification (do this FIRST — it is the gate)

Read `internal/check/check.go` lines 118-160 and confirm, against the evidence in
`research/verification_findings.md` §3:

| Point | Expected in 123-129 | Status |
|-------|---------------------|--------|
| (a) Index only emits dirs with SKILL.md → rule can't fire | present, accurate | ✅ keep |
| (b) Heuristic would false-positive on grouping dirs | present, accurate | ✅ keep |
| (c) Rule reframed as "invalid SKILL.md frontmatter" (perr branch) | **MISSING** | ⚠️ add |

If your read disagrees with this table (e.g. you find (c) IS already in 123-129),
STOP and add nothing — the gate is "confirm accurate," and an already-complete
comment requires no edit. Otherwise proceed to Task 1.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: REFINEn the Check doc-comment in internal/check/check.go (lines ~123-129)
  - GOAL: make the block self-contained on contract points a, b, AND c.
  - FIND: the block starting "// check does NOT scan for ..." (lines 124-129),
          immediately before `func Check` (line 130).
  - REPLACE the whole block with the Proposed text below. It (1) keeps the (a)
          "Index only emits dirs that CONTAIN a SKILL.md" claim and makes explicit
          "so the §9 ... rule can never fire"; (2) keeps the (b) heuristic-
          false-positive claim; (3) ADDS (c) the reframe cross-reference to the
          "invalid SKILL.md frontmatter" ERROR (perr != nil branch); (4) preserves
          the trailing "empty besides SKILL.md" WARN sentence.
  - DO NOT: touch lines 145-149 (the perr branch), 150+ (!fm.HasFM), checkFields,
          appendDupFindings, any ERROR string, or any test.
  - NAMING/STYLE: one "// " per line; keep the §-citations ((research §2)/(§3));
          keep ASCII where the original is ASCII (the § and "" are already
          non-ASCII in the original — preserve them as-is; gofmt handles UTF-8).

Task 2: FORMAT + VET (the only "compilers" for a comment change)
  - COMMAND: gofmt -w internal/check/check.go && gofmt -l internal/check/check.go
  - EXPECT: the second command prints nothing (file is gofmt-clean).
  - COMMAND: go vet ./internal/check/
  - EXPECT: clean (no output).

Task 3: UNIT TESTS (the reframe branch must stay green)
  - COMMAND: go test ./internal/check/ -v
  - EXPECT: PASS, including TestCheckMalformedYAML (pins the reframed ERROR) and
            TestCheckMissingFrontmatterBlock (pins the no-'---' ERROR).

Task 4: §13 ACCEPTANCE GATE (build + check the example skill)
  - COMMAND: go build -o skpp .
  - COMMAND: ./skpp check
  - EXPECT: a line exactly "OK    example (example)" and a summary
            "1 skills, 0 errors, 0 warnings"; exit code 0.
  - NOTE: ./skpp is gitignored (§16 line 1 `/skpp`); building it is expected and
          leaves git status clean (modulo the .gitignore outcome of P1.M5.T1.S1).

Task 5: COLLATERAL SANITY (nothing else moved)
  - COMMAND: git diff --name-only
  - EXPECT: only internal/check/check.go (plus whatever P1.M5.T1.S1 does to
            .gitignore if it has landed — that is a different file, not a conflict).
  - COMMAND: git diff --quiet PRD.md && echo "PRD.md untouched OK"
  - EXPECT: "PRD.md untouched OK".
```

### Implementation Patterns & Key Details

```go
// === EXACT EDIT (apply with the edit tool; oldText is unique in the file) ===

// oldText (current lines 124-129, verbatim — copy exactly, incl. the § and ""):
// check does NOT scan for "directories that lack SKILL.md but look like skills":
// discover.Index only emits dirs that CONTAIN a SKILL.md, and a heuristic for the
// gap would false-positive on legitimate grouping dirs (research §2). The §9
// "empty besides SKILL.md" WARN is intentionally NOT implemented (research §3):
// the shipped example skill IS only SKILL.md, and enabling it would break the
// §13 acceptance ("reports the example as OK").

// newText (covers a, b, AND c; c is the added reframe cross-reference):
// check does NOT scan for "directories that lack SKILL.md but look like skills":
// discover.Index only emits dirs that CONTAIN a SKILL.md, so the §9 "skill dir
// has no SKILL.md" rule can never fire — a grouping dir without SKILL.md is never
// indexed, never inspected. A heuristic for the gap would false-positive on
// legitimate grouping dirs (research §2). The §9 rule's actionable form is
// therefore REFRAMED below as the "invalid SKILL.md frontmatter" ERROR (the
// perr != nil branch): an unusable/malformed SKILL.md is the closest reachable
// condition to "no usable SKILL.md". The §9 "empty besides SKILL.md" WARN is
// intentionally NOT implemented (research §3): the shipped example skill IS only
// SKILL.md, and enabling it would break the §13 acceptance ("reports the
// example as OK").
```

### Integration Points

```yaml
DOCUMENTATION:
  - file: internal/check/check.go
  - location: Check doc-comment, lines ~123-129
  - change: add contract point (c) — the reframe cross-reference
  - side-effects: none (comment-only; no behavior, no API, no build config)

NO OTHER INTEGRATION POINTS:
  - no database, no config, no routes, no dependency, no test.
  - PRD.md is read-only (D7; do NOT edit §9).
  - no conflict with P1.M5.T1.S1 (.gitignore) or any prior/in-flight item.
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
# A comment change has no "syntax" beyond gofmt alignment. Run after the edit:
gofmt -w internal/check/check.go
gofmt -l internal/check/check.go     # Expected: empty (no files need formatting)
go vet ./internal/check/             # Expected: clean (no output)
# Expected: gofmt lists nothing; go vet prints nothing. Fix before proceeding.
```

### Level 2: Unit Tests (Component Validation)

```bash
cd /home/dustin/projects/skpp
# The reframe branch is pinned by TestCheckMalformedYAML; the no-'---' branch by
# TestCheckMissingFrontmatterBlock. Both must stay green.
go test ./internal/check/ -v
# Expected: PASS (all tests), specifically:
#   TestCheckMalformedYAML            — asserts "invalid SKILL.md frontmatter" ERROR
#   TestCheckMissingFrontmatterBlock  — asserts "missing frontmatter block" ERROR
#   TestCheckValidSkillIsClean        — the example-shaped skill is OK
# If anything fails, you changed code you should not have — revert to comment-only.
```

### Level 3: Integration / §13 Acceptance Gate

```bash
cd /home/dustin/projects/skpp
# Build the binary (gitignored; required for the gate — there is no committed ./skpp).
go build -o skpp . && echo "build OK"

# §13 gate: check reports the shipped example skill as OK, exit 0.
./skpp check
# Expected stdout:
#   OK    example (example)
#   1 skills, 0 errors, 0 warnings
./skpp check >/dev/null 2>&1; echo "exit=$?"   # Expected: exit=0
```

### Level 4: Creative & Domain-Specific Validation

```bash
cd /home/dustin/projects/skpp
# Spec-alignment/documentation task: the only "domain" check is that the comment
# now says what the contract requires. Grep the refined block for the three points:
sed -n '/check does NOT scan/,/reports the example as OK/p' internal/check/check.go \
  | grep -q "only emits dirs that CONTAIN a SKILL.md" && echo "(a) Index-only OK"
sed -n '/check does NOT scan/,/reports the example as OK/p' internal/check/check.go \
  | grep -q "false-positive on" && echo "(b) heuristic-rejected OK"
sed -n '/check does NOT scan/,/reports the example as OK/p' internal/check/check.go \
  | grep -q 'REFRAMED below as the "invalid SKILL.md frontmatter" ERROR' && echo "(c) reframe OK"
# Expected: all three echo. If (c) is missing, the refinement did not land — redo Task 1.

# Confirm the perr-branch ERROR string is byte-identical (you must not have edited it):
grep -n 'invalid SKILL.md frontmatter: " + perr.Error()' internal/check/check.go
# Expected: exactly one match, inside the case perr != nil: branch.
```

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` empty; `go vet ./internal/check/` clean.
- [ ] Level 2 PASS — `go test ./internal/check/ -v` all green (incl. `TestCheckMalformedYAML`).
- [ ] Level 3 PASS — `go build -o skpp .` OK; `./skpp check` prints `OK    example (example)`, exit 0.
- [ ] Level 4 PASS — the refined block greps positive for (a), (b), and (c); the ERROR string is unchanged.

### Feature Validation
- [ ] The 123-129 comment explains (a) why the rule can't fire (Index emits only SKILL.md dirs).
- [ ] The 123-129 comment explains (b) why no heuristic (false-positives on grouping dirs).
- [ ] The 123-129 comment states (c) the reframe as "invalid SKILL.md frontmatter" (perr branch).
- [ ] The "empty besides SKILL.md" WARN note is retained.
- [ ] §13 acceptance gate (`./skpp check` → example OK, exit 0) still passes.

### Code Quality Validation
- [ ] Comment-only — no executable line, ERROR string, or test changed.
- [ ] Only `internal/check/check.go` modified; `PRD.md` byte-identical.
- [ ] No conflict with P1.M5.T1.S1 (`.gitignore`) or any prior/in-flight item.
- [ ] gofmt alignment and existing §-citation style preserved.

### Documentation & Deployment
- [ ] Mode A: the comment IS the documentation — no separate doc file written.
- [ ] If the refinement changed any wording, the clarification rides with this subtask (per item DOCS).

---

## Anti-Patterns to Avoid

- ❌ Don't implement a "no SKILL.md" heuristic — D7 explicitly forbids it; it would
  false-positive on legitimate grouping dirs (PRD §7.1 allows `scripts/`/`references/`/`assets/` siblings).
- ❌ Don't touch the `perr != nil` branch (lines 145-149), its ERROR string, or any
  test — the reframe there is already correct and pinned by `TestCheckMalformedYAML`.
- ❌ Don't trust the contract's literal "check.go:150" citation — line 150 is
  `case !fm.HasFM:` (the NEXT branch). The reframe is at 145-149.
- ❌ Don't edit `PRD.md` (§9 is the spec; human-owned; D7).
- ❌ Don't add a new doc file or test — Mode A + "No new tests needed" (item TESTS).
- ❌ Don't touch `.gitignore` (owned by the parallel P1.M5.T1.S1) or `main.go`
  (owned by P1.M4.T2.S1) — this task edits `internal/check/check.go` only.
