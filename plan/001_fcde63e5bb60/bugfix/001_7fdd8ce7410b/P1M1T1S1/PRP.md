# PRP — P1.M1.T1.S1: Wire `Source` into `--path` stderr reporting and update tests

> **Subtask:** P1.M1.T1.S1 — fixes QA Issue 1 ("`--path` does not report which discovery rule won").
> **Scope boundary:** Surgical change to ONE branch in `main.go` + update/add tests in `main_test.go`. No new types, no `skillsdir.go` edits, no docs (Mode A).
> **Severity:** Major (documented diagnostic feature computed but never surfaced).

---

## Goal

**Feature Goal**: Make `skpp --path` tell the user which of the three PRD §8 discovery rules (env var / sibling-of-binary / walk-up) produced the resolved directory, by printing the `skillsdir.Source` label to **stderr** — without changing the byte-exact **stdout** output that the §13 acceptance gate locks.

**Deliverable**: A one-line wiring change in `main.go`'s `c.path` branch (bind the `src` return value that is currently discarded to `_`, and emit `(found via <Source.String()>)\n` to stderr), plus two updated tests and one new feature test in `main_test.go`.

**Success Definition**:
- `./skpp --path` prints `<dir>\n` to stdout AND `(found via <rule>)\n` to stderr.
- The §13 gate `test "$(./skpp --path)" = "$PWD/skills"` STILL passes (stdout byte-exact — `$()` ignores stderr).
- `go test ./...` is green; `go vet ./...` and `gofmt -l .` are clean.
- The `Source.String()` labels are no longer dead-code-eliminated from the binary.

---

## Why

- **PRD §8 explicitly requires it**: "`skpp --path` reports which rule won. This is the single most failure-prone area." The `Source` enum, its `String()` method, and its unit tests all exist *specifically* for this — but `main.go` discards the value (`dir, _, err := ...`), so the whole subsystem is dead code in the shipped binary.
- **Fixes a real footgun**: a user who typos `SKPP_SKILLS_DIR=/typo/not/real` gets silent fall-through to the sibling/walk-up rule and sees a valid-looking path with no indication the env var was ignored. The stderr label makes that fall-through visible on a terminal while staying invisible to `$(...)` capture.
- **Zero new surface**: no new flag, no help-text change, no docs (Mode A). The labels already exist and are unit-tested in `skillsdir_test.go:TestSourceString`. This subtask just wires them through.

---

## What

`skpp --path` success path gains a single stderr line. Everything else is unchanged.

### Behavior change (success path only)

| | stdout | stderr | exit |
|---|---|---|---|
| **Before** | `<dir>\n` | *(empty)* | 0 |
| **After**  | `<dir>\n` | `(found via <Source.String()>)\n` | 0 |

### Behavior UNCHANGED

| Case | stdout | stderr | exit |
|---|---|---|---|
| `--path` failure (ErrNotFound) | *(empty)* | one-line fix message (SKPP_SKILLS_DIR / cd / reinstall) | 1 |
| All other modes (`<tag>`, `--list`, `--search`, `--all`, `check`, `--version`, `--help`) | unchanged | unchanged | unchanged |

The exact stderr label depends on which rule won:
- `SourceEnv` → `(found via SKPP_SKILLS_DIR)`
- `SourceSibling` → `(found via sibling of binary)`
- `SourceWalkUp` → `(found via ancestor of cwd)`

### Success Criteria

- [ ] `main.go` `c.path` branch binds `src` (not `_`) and calls `fmt.Fprintf(stderr, "(found via %s)\n", src)` after the stdout print
- [ ] `fmt.Fprintln(stdout, dir)` line is byte-identical (still emits `<dir>\n` only)
- [ ] `--path` failure path is untouched (src not referenced; stdout empty on error)
- [ ] `internal/skillsdir/skillsdir.go` is NOT modified
- [ ] `usageText` const (`--path, -p  Print the resolved skills directory`) is NOT modified
- [ ] `TestRunPathSuccess` and `TestRunPathShortFlag` updated to expect the stderr label
- [ ] One new test added asserting stdout byte-exact + stderr label for the env case
- [ ] `go test ./...`, `go vet ./...`, `gofmt -l .` all clean
- [ ] §13 gate still passes: `test "$(./skpp --path)" = "$PWD/skills"`

---

## All Needed Context

### Context Completeness Check

_Pass: the exact current code block, the exact target code block, the exact label strings, the exact tests to change, and the §13-preservation argument are all below. An implementer who has never seen this repo can apply the edits verbatim._

### Documentation & References

```yaml
# MUST READ — the change spec and rationale
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/decisions.md
  why: "§D1 decides the label goes to stderr (not stdout, not a new flag) precisely to preserve the §13 gate"
  critical: "stderr is invisible to $() capture — this is WHY the fix is safe"

- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/issue_analysis.md
  why: "Issue 1 root cause (src discarded to _), the exact fix snippet, and the test-impact list"
  section: "Issue 1 (MAJOR)"

- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/P1M1T1S1/research/verified_facts.md
  why: "Line-exact current/target code blocks, the label table, and the exact test edits (old+new)"
  critical: "Contains the verbatim oldText/newText for both the code edit and each test edit"

- file: internal/skillsdir/skillsdir.go
  why: "The Source type + String() method (lines 24-50) and Find() signature (line 232: returns dir string, src Source, err error)"
  pattern: "type Source int with String() satisfying fmt.Stringer — that's why %s works on a Source value"

- file: internal/skillsdir/skillsdir_test.go
  why: "TestSourceString (lines 30-46) proves the label strings: SKPP_SKILLS_DIR / sibling of binary / ancestor of cwd / unknown"
  pattern: "table-driven; the env case (SourceEnv) is the only one deterministic enough to assert through run()"

- file: PRD.md
  why: "§8 (discovery rules + 'skpp --path reports which rule won'); §13 acceptance gate; §6.1 --path row"
  critical: "PRD.md is READ-ONLY. Do not edit it. The §13 gate is the hard stdout contract."

- url: https://pkg.go.dev/fmt
  why: "fmt.Stringer: a value passed to %s calls its String() method. Source implements Stringer."
  section: "Printing — the %s verb for types implementing fmt.Stringer"
```

### Current Codebase tree (relevant slice)

```bash
$ cd /home/dustin/projects/skpp && ls main.go main_test.go internal/skillsdir/
main.go                              # 529 lines — c.path branch is lines 268-281
main_test.go                         # 1510 lines — TestRunPathSuccess ~169, TestRunPathShortFlag ~187
internal/skillsdir/
├── skillsdir.go                     # Source enum + String() + Find() — NOT MODIFIED
└── skillsdir_test.go                # TestSourceString already covers the labels
```

### Desired Codebase tree (files touched)

```bash
skpp/
├── main.go              # MODIFY — c.path branch: bind src, add fmt.Fprintf(stderr,...)
└── main_test.go         # MODIFY — update 2 tests; ADD TestRunPathReportsSourceLabel
# (internal/skillsdir/* unchanged)
```

| File | Change | Lines |
|---|---|---|
| `main.go` | edit `c.path` branch | ~268-281 |
| `main_test.go` | update `TestRunPathSuccess`, `TestRunPathShortFlag`; add `TestRunPathReportsSourceLabel` | ~169, ~187, new |

### Known Gotchas of our codebase & Go toolchain

```go
// GOTCHA #1 — stdout is locked by the §13 acceptance gate. Do NOT change
// fmt.Fprintln(stdout, dir). The new line MUST go to stderr. Confirmed safe:
// `test "$(./skpp --path)" = "$PWD/skills"` captures stdout only ($() ignores
// stderr), so an extra stderr line is invisible to the gate.

// GOTCHA #2 — Source satisfies fmt.Stringer, so use %s (not %d, not manual switch).
//   fmt.Fprintf(stderr, "(found via %s)\n", src)   // src is skillsdir.Source
// %s calls src.String() automatically. The labels come from skillsdir.go:38-49.

// GOTCHA #3 — the existing comment `// src is for reporting only; not printed`
// on the `dir, _, err :=` line becomes FALSE after the edit. Either drop it or
// rewrite it; do not leave a comment contradicting the new code.

// GOTCHA #4 — do NOT add the stderr print to the FAILURE path. On ErrNotFound,
// Find() still returns src (a zero-value Source), but the contract is "nothing
// on stdout, the one-line-fix message on stderr, exit 1". The new
// fmt.Fprintf(stderr, ...) must go AFTER the `if err != nil { return 1 }` block,
// on the success path only. (Putting it before the error check would print
// "(found via unknown)" before the real error and corrupt the failure contract.)

// GOTCHA #5 — TestRunPathSuccess currently asserts `errOut.Len() != 0 → FAIL`
// (i.e. stderr must be EMPTY). This assertion WILL FAIL after the fix because
// stderr now holds the label. You MUST update it, not delete it. Same idea for
// the short-flag test (add the assertion there; it currently has none).

// GOTCHA #6 — only the env case (SourceEnv → "SKPP_SKILLS_DIR") is deterministic
// to test through run() via t.Setenv. Do NOT try to unit-test the sibling/walk-up
// labels through run() — they depend on the binary path / cwd and are flaky.
// Those Source values are already covered by skillsdir.TestSourceString.
```

---

## Implementation Blueprint

### Data models and structure

**No new data models.** This subtask reuses the existing `skillsdir.Source` type and its `String()` method verbatim:

```go
// internal/skillsdir/skillsdir.go:24-50 (EXISTING — do not modify)
type Source int
const (
    SourceEnv Source = iota
    SourceSibling
    SourceWalkUp
)
func (s Source) String() string {
    switch s {
    case SourceEnv:     return "SKPP_SKILLS_DIR"
    case SourceSibling: return "sibling of binary"
    case SourceWalkUp:  return "ancestor of cwd"
    default:            return "unknown"
    }
}

// Find() returns (dir string, src Source, err error) — EXISTING signature, unchanged.
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — bind src and emit the stderr label
  - FILE: main.go
  - LOCATE: the `if c.path {` branch (lines ~268-281), inside func run(args, stdout, stderr) int
  - EDIT A (exact oldText → newText):
      OLD: `dir, _, err := skillsdir.Find() // src is for reporting only; not printed`
      NEW: `dir, src, err := skillsdir.Find()`
      (drop the now-false trailing comment; src is now used on the next line)
  - EDIT B (insert one line after the stdout print, before return 0):
      After:  `fmt.Fprintln(stdout, dir)`
      Insert: `fmt.Fprintf(stderr, "(found via %s)\n", src)`
      KEEP:   `return 0` immediately after
  - WHY %s: Source implements fmt.Stringer (skillsdir.go:31-49), so %s calls String().
  - DO NOT TOUCH: the `if err != nil { ... return 1 }` failure block above it
  - DO NOT TOUCH: const usageText (the --path help line stays "Print the resolved skills directory")
  - DO NOT TOUCH: internal/skillsdir/skillsdir.go
  - VERIFY: `go build ./...` compiles; `go vet ./...` clean

Task 2: UPDATE main_test.go — TestRunPathSuccess (line ~169)
  - FILE: main_test.go
  - LOCATE: func TestRunPathSuccess; the stderr assertion block:
        `if errOut.Len() != 0 {
             t.Errorf("run(--path) success stderr=%q; want empty", errOut.String())
         }`
  - REPLACE WITH:
        `if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
             t.Errorf("run(--path) success stderr=%q; want %q (Issue 1 source label)", got, want)
         }`
  - WHY: this test sets SKPP_SKILLS_DIR via t.Setenv (rule 1 wins → SourceEnv),
         so the expected stderr label is exactly "(found via SKPP_SKILLS_DIR)\n".
         The OLD assertion demanded empty stderr, which the fix breaks.

Task 3: UPDATE main_test.go — TestRunPathShortFlag (line ~187)
  - FILE: main_test.go
  - LOCATE: func TestRunPathShortFlag; after the existing stdout assertion, ADD:
        `if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
             t.Errorf("run(-p) stderr=%q; want %q (Issue 1 source label)", got, want)
         }`
  - WHY: same env setup → same label; this test previously asserted nothing on
         stderr and silently would not have caught the feature. Now it does.

Task 4: ADD main_test.go — TestRunPathReportsSourceLabel (new feature test)
  - FILE: main_test.go
  - PLACEMENT: immediately after TestRunPathShortFlag (keep the --path tests grouped)
  - NAMING: TestRunPathReportsSourceLabel
  - CONTENT: see Implementation Patterns below — sets SKPP_SKILLS_DIR via t.Setenv,
             asserts stdout == filepath.Clean(dir)+"\n" (byte-exact, §13 preserved)
             AND stderr == "(found via SKPP_SKILLS_DIR)\n".
  - WHY: a distinct, intent-named test for Issue 1 that survives refactors of the
         two updated tests and documents the contract explicitly.

Task 5: VALIDATE (all gates green)
  - gofmt -w main.go main_test.go     (then `gofmt -l .` must print nothing)
  - go test ./...                     (all pass)
  - go test ./ -run 'TestRunPath' -v  (just the --path tests, verbose)
  - go vet ./...                      (clean)
  - go build -o skpp . && ./acceptance_check.sh   (§13 gate; see Validation Loop)
```

### Implementation Patterns & Key Details

```go
// === main.go: the target c.path branch (after Task 1) ===
// NOTE: only TWO substantive changes vs current — `dir, src, err` and the
// fmt.Fprintf line. Comments updated to stay accurate.

if c.path {
    dir, src, err := skillsdir.Find()
    if err != nil {
        // Find() returns skillsdir.ErrNotFound whose message is the
        // user-facing one-line fix (PRD §8.4/§6.4). Print it verbatim to
        // stderr (NOT stdout) so $(...) stays empty on failure.
        fmt.Fprintln(stderr, err)
        return 1
    }
    // Byte-exact: ONLY the dir + newline on stdout. The §13 acceptance gate
    // `test "$(./skpp --path)" = "$PWD/skills"` depends on this — $() captures
    // stdout only, so the stderr label below does NOT break it.
    fmt.Fprintln(stdout, dir)
    // Issue 1 (QA): report which §8 discovery rule won, to stderr. A typo'd
    // SKPP_SKILLS_DIR silently falls through to sibling/walk-up; this label
    // makes that visible without polluting stdout. Labels from Source.String().
    fmt.Fprintf(stderr, "(found via %s)\n", src)
    return 0
}
```

```go
// === main_test.go: the new feature test (Task 4) ===
// Place after TestRunPathShortFlag. Imports needed are already present in
// main_test.go (bytes, filepath, testing) — verify, do not assume.

// Issue 1 (QA): --path must report which §8 rule won to stderr, while stdout
// stays byte-exact so the §13 `test "$(./skpp --path)" = "$PWD/skills"` gate
// still passes. The env case is deterministic; sibling/walk-up are covered by
// skillsdir.TestSourceString.
func TestRunPathReportsSourceLabel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SKPP_SKILLS_DIR", dir) // rule 1 wins -> SourceEnv
	var out, errOut bytes.Buffer
	if code := run([]string{"--path"}, &out, &errOut); code != 0 {
		t.Fatalf("run(--path): code=%d; want 0", code)
	}
	// stdout: byte-exact dir + newline (§13 contract preserved).
	if got, want := out.String(), filepath.Clean(dir)+"\n"; got != want {
		t.Errorf("--path stdout=%q; want %q", got, want)
	}
	// stderr: the SourceEnv label, exactly, nothing else.
	if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
		t.Errorf("--path stderr=%q; want %q", got, want)
	}
}
```

### Integration Points

```yaml
NO NEW INTEGRATION POINTS:
  - No new types, no new imports (fmt already imported in main.go; bytes/filepath
    already imported in main_test.go).
  - No DB, no config, no routes, no API surface.
  - skillsdir.Find()'s signature already returns (dir, src, err) — we simply
    stop discarding src.
  - The stderr label is a diagnostic; it changes NO flag, NO help text, NO exit
    code, NO stdout bytes. Downstream subtasks and the pi integration are
    unaffected (pi consumes stdout via $(...), which ignores stderr).
```

---

## Validation Loop

### Level 1: Build, format, vet (immediate)

```bash
cd /home/dustin/projects/skpp

# Format the two touched files, then assert the whole tree is gofmt-clean.
gofmt -w main.go main_test.go
test -z "$(gofmt -l .)" || { echo "FAIL: gofmt reports unformatted files: $(gofmt -l .)"; exit 1; }

# Compile (catches the `src` declared-and-used wiring).
go build ./... || { echo "FAIL: go build"; exit 1; }

# Static checks.
go vet ./... || { echo "FAIL: go vet"; exit 1; }
echo "Level 1 PASS"
```

### Level 2: Unit tests (component validation)

```bash
cd /home/dustin/projects/skpp

# The --path tests specifically (verbose — confirm the two updates + the new test).
go test ./ -run 'TestRunPath' -v
# Expected: TestRunPathSuccess, TestRunPathShortFlag, TestRunPathReportsSourceLabel,
#           and TestRunPathFailureErrNotFound all PASS.

# Full suite (regression guard: nothing else broke).
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: The §13 acceptance gate + manual behavior (system validation)

```bash
cd /home/dustin/projects/skpp

# Build the binary exactly as install.sh does.
go build -o skpp .

# GATE 1 (the locked stdout contract) — MUST still pass.
test "$(./skpp --path)" = "$PWD/skills" \
  || { echo "FAIL: §13 gate broken — stdout changed"; exit 1; }
echo "§13 stdout gate PASS"

# GATE 2 — stdout contains ONLY the dir (no label leaked into stdout).
STDOUT_ONLY="$(./skpp --path 2>/dev/null)"
test "$STDOUT_ONLY" = "$PWD/skills" \
  || { echo "FAIL: stdout has extra content: $STDOUT_ONLY"; exit 1; }

# GATE 3 — stderr now carries the source label (the fix).
LABEL="$(./skpp --path 2>&1 >/dev/null)"
echo "$LABEL" | grep -q '^found via ' \
  || { echo "FAIL: stderr missing 'found via' label; got: $LABEL"; exit 1; }
# From repo root the binary resolves via sibling-of-binary:
echo "$LABEL" | grep -q 'sibling of binary' \
  || { echo "FAIL: expected sibling-of-binary rule from repo root; got: $LABEL"; exit 1; }
echo "stderr label PASS: $LABEL"

# GATE 4 — the typo'd-env footgun is now VISIBLE (was previously invisible).
TYPO_OUT="$(SKPP_SKILLS_DIR=/typo/not/real ./skpp --path 2>/dev/null)"
TYPO_LABEL="$(SKPP_SKILLS_DIR=/typo/not/real ./skpp --path 2>&1 >/dev/null)"
test "$TYPO_OUT" = "$PWD/skills" \
  || { echo "FAIL: typo'd env stdout unexpected: $TYPO_OUT"; exit 1; }
echo "$TYPO_LABEL" | grep -q 'sibling of binary' \
  || { echo "FAIL: typo'd env should fall through to sibling; got: $TYPO_LABEL"; exit 1; }
echo "typo footgun now visible: $TYPO_LABEL"

# GATE 5 — env var honored: label reports SKPP_SKILLS_DIR, dir is the env dir.
ENV_DIR="$(mktemp -d)"
ENV_OUT="$(SKPP_SKILLS_DIR="$ENV_DIR" ./skpp --path 2>/dev/null)"
ENV_LABEL="$(SKPP_SKILLS_DIR="$ENV_DIR" ./skpp --path 2>&1 >/dev/null)"
test "$ENV_OUT" = "$ENV_DIR" || { echo "FAIL: env dir not honored: $ENV_OUT"; exit 1; }
test "$ENV_LABEL" = "found via SKPP_SKILLS_DIR" \
  || { echo "FAIL: env label wrong: $ENV_LABEL"; exit 1; }
echo "env rule PASS: $ENV_LABEL"
echo "Level 3 PASS"
```

### Level 4: Dead-code reversal + scope-boundary check

```bash
cd /home/dustin/projects/skpp

# The labels were dead-code-eliminated before (src was _). Now String() is
# referenced, so the linker keeps the labels in the binary.
go build -o skpp .
COUNT=$(strings ./skpp | grep -c 'found via' || true)
test "$COUNT" -ge 1 || { echo "FAIL: 'found via' not in binary (dead code not reversed)"; exit 1; }
strings ./skpp | grep -q 'sibling of binary' || { echo "FAIL: label missing from binary"; exit 1; }
echo "dead-code reversal PASS (found via appears $COUNT time(s))"

# Scope boundary: skillsdir.go untouched (no new types, no signature change).
git diff --quiet internal/skillsdir/skillsdir.go \
  || { echo "FAIL: skillsdir.go was modified (out of scope)"; exit 1; }
# usageText untouched (Mode A — no help-text change).
git diff --quiet --grep-ignore || true   # (informational)
# Only main.go and main_test.go changed:
CHANGED="$(git status --porcelain -- main.go main_test.go | wc -l)"
git diff --quiet -- main.go main_test.go && CHANGED=0
echo "files changed: $(git status --porcelain | grep -E 'main\.go|main_test\.go')"
echo "Level 4 PASS"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l .` empty, `go build ./...` compiles, `go vet ./...` clean
- [ ] Level 2 PASS — `go test ./...` green; the 4 `TestRunPath*` tests pass verbosely
- [ ] Level 3 PASS — §13 stdout gate holds; stderr carries the label; typo/env cases verified
- [ ] Level 4 PASS — labels present in the binary (dead code reversed); skillsdir.go untouched

### Feature Validation
- [ ] `./skpp --path` prints `<dir>\n` to stdout AND `(found via <rule>)\n` to stderr
- [ ] stdout is byte-exact: `test "$(./skpp --path)" = "$PWD/skills"` passes
- [ ] stderr stays empty on the failure path (ErrNotFound contract unchanged)
- [ ] typo'd `SKPP_SKILLS_DIR` now visibly falls through (the footgun this fixes)
- [ ] `SKPP_SKILLS_DIR=<dir>` reports `found via SKPP_SKILLS_DIR`

### Code Quality Validation
- [ ] Used `%s` on the `Source` value (fmt.Stringer) — no manual switch, no `%d`
- [ ] The stale comment `// src is for reporting only; not printed` removed/rewritten
- [ ] The new `fmt.Fprintf` is on the success path only (after the error-return block)
- [ ] Tests assert the EXACT stderr string, not just non-empty (catches label drift)
- [ ] No new imports added (fmt/bytes/filepath already present)

### Scope Discipline (Mode A — docs deferred)
- [ ] `internal/skillsdir/skillsdir.go` NOT modified
- [ ] `const usageText` (--path help line) NOT modified
- [ ] README.md NOT modified (deferred to P1.M5.T3 Mode B doc sync)
- [ ] No new flag, no new type, no new file created
- [ ] PRD.md / tasks.json / prd_snapshot.md NOT modified (read-only / orchestrator-owned)

---

## Anti-Patterns to Avoid

- ❌ **Don't print the source to stdout.** The §13 gate locks stdout to `<dir>\n`. Use stderr — that's the whole point of decisions.md §D1.
- ❌ **Don't use `%d` or a manual `switch` on `src`.** `Source` implements `fmt.Stringer`; `%s` calls `String()` and pulls the already-tested labels. Re-implementing the switch duplicates `skillsdir.go` and will drift.
- ❌ **Don't add the `fmt.Fprintf` before the `if err != nil` block.** On failure `src` is a zero value and you'd print `(found via unknown)` before the real error, corrupting the §6.4 failure contract (nothing on stdout, one-line fix on stderr, exit 1).
- ❌ **Don't delete the failing `errOut.Len() != 0` assertion in TestRunPathSuccess.** Update it to expect the new label. Deleting it would let a regression pass silently.
- ❌ **Don't try to unit-test the sibling/walk-up labels through `run()`.** They depend on the binary path and cwd; they're flaky there. Assert only the env case via `t.Setenv`. The other labels are covered by `skillsdir.TestSourceString`.
- ❌ **Don't modify `skillsdir.go`, `usageText`, or README.** This is Mode A (code + tests only). README sync is P1.M5.T3.
- ❌ **Don't change the failure path.** `--path` with an unresolvable dir must still emit nothing on stdout and exit 1 — the source label is a success-path diagnostic only.

---

## Confidence Score

**10/10** — The change is two lines of substantive code (bind `src`; one `fmt.Fprintf` to stderr) plus mechanical test updates. Every artifact — the exact current code block, the exact target code block, the exact label strings (verified against `TestSourceString`), the exact old/new test assertions, and the §13-preservation argument — is captured verbatim in `research/verified_facts.md`. The `Source.String()` method and `Find()` signature already exist and are tested; nothing new is invented. The only residual risk (an implementer ignoring "stderr, not stdout") is eliminated by GOTCHA #1 and the Level 3 Gate 1 check.
