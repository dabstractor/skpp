# PRP — P1.M1.T1.S1: Flip the no-mode fallthrough to implicit help (stdout, exit 0)

> **Subtask:** P1.M1.T1.S1 — the sole subtask of P1.M1.T1 (decision 17 / PRD §6.3: bare `skilldozer` ⇒ usage to stdout, exit 0).
> **Scope boundary:** Three lines of behavior change in `run()`'s final fallthrough (stderr→stdout, 1→0), three doc-comment corrections, and two test flips. Does NOT touch any genuine-failure path (unknown flag / exclusivity / unresolved tag / unconfigured all stay stderr + non-zero), does NOT add the `completion` subcommand (P1.M2), does NOT edit the README (P1.M3.T1), adds no deps.

---

## Goal

**Feature Goal**: Make a bare `skilldozer` invocation (no args, or modifiers-only like `--no-color` with no mode) print usage to **stdout** and exit **0** — i.e. behave as implicit `--help` (PRD §6.3 / §19 decision 17) — so `skilldozer | grep …` sees the help on the piped stream.

**Deliverable**: Five surgical edits, no new files:
1. `main.go:699` — `fmt.Fprint(stderr, usage())` → `fmt.Fprint(stdout, usage())`.
2. `main.go:700` — `return 1` → `return 0`.
3. `main.go:695-698` (fallthrough comment), `main.go:48-51` (usageText doc), `main.go:417-423` (run() exit-code doc) — correct all three to state no-args ⇒ stdout/exit 0 (§6.3/decision 17).
4. `main_test.go:277-291` `TestRunDefaultNoArgs` — flip assertions to code==0, USAGE on stdout, stderr empty.
5. `main_test.go:1668-1684` `TestRunModifiersOnlyNoMode` — same flip for `run([]string{"--no-color"})`.

**Success Definition**: `go test ./...` green; `run(nil, …)` returns 0 with usage on stdout and empty stderr; `run(["--no-color"], …)` same; the §13 Grepability contract passes (`out=$(./skilldozer 2>/dev/null)` ⇒ rc 0, grep finds USAGE, and `./skilldozer 2>&1 >/dev/null` is empty); every genuine-failure test still asserts stderr + non-zero unchanged; `go.mod`/`go.sum` unchanged.

---

## User Persona (if applicable)

**Target User**: anyone piping skilldozer's output — `skilldozer | grep …`, scripts, `skilldozer` typed at a shell to recall usage.

**Use Case**: A user runs bare `skilldozer` and pipes to `grep` to find a flag.

**Pain Points Addressed**: Today bare `skilldozer` writes usage to **stderr** (exit 1), so `skilldozer | grep …` sees nothing (grep reads stdout). The flip puts help on the piped stream while keeping genuine failures on stderr (§6.4 `$(...)` contract intact).

---

## Why

- **PRD §6.3 (decision 17) is authoritative**: "No arguments and no flag ⇒ print usage to **stdout**, exit `0`. Bare invocation is **implicit `--help`** … chosen so `skilldozer | grep …` works — the help must land on the piped stream to be grep-friendly."
- **Decision 17 explicitly overrides the old behavior**: "Previously stderr/exit-1 'parity with `get-server-config.sh`'; overridden." Only no-args is reclassified (error→help); genuine failures stay on stderr, non-zero, preserving the §6.4 `$(...)` contract.
- **§13 acceptance requires it**: the "Grepability contract" gate asserts no-args help is on stdout, exit 0, and that no-args writes NOTHING to stderr. The current code (stderr/exit-1) fails this gate.
- **The in-code docs currently lie**: three comments still describe the old stderr/exit-1 behavior. Fixing the behavior without fixing the docs leaves the exit-code contract self-contradictory.

---

## What

`run()`'s **final fallthrough** (the "no recognized mode" tail) changes from `stderr`/`1` to `stdout`/`0`. It is reached only when no mode and no tags were given (truly no-args, or modifiers-only like `--no-color`/`--relative`/`--file` with no tag/mode). Everything else in `run()` is untouched:

- `--help`/`--version` already print to stdout/exit 0 (precedence 1 & 2) — unchanged.
- unknown flag → stderr/exit 2 — unchanged.
- `--store` missing value → stderr/exit 2 — unchanged.
- exclusivity → stderr/exit 2 — unchanged.
- unresolved/ambiguous tag → stderr/exit 1 — unchanged.
- unconfigured → stderr/exit 1 — unchanged.

### Success Criteria

- [ ] `main.go:699` writes to `stdout`; `main.go:700` returns `0`
- [ ] The fallthrough comment (`main.go:695-698`), usageText doc (`main.go:48-51`), and run() exit-code doc (`main.go:417-423`) all describe no-args ⇒ stdout/exit 0 (§6.3/decision 17)
- [ ] `TestRunDefaultNoArgs` asserts code==0, `Contains(out, "USAGE")`, `errOut.Len()==0`
- [ ] `TestRunModifiersOnlyNoMode` asserts the same for `run([]string{"--no-color"})`
- [ ] No genuine-failure path or its tests change (unknown flag/exclusivity/unresolved tag/unconfigured stay stderr + non-zero)
- [ ] `go test ./...` green; `go vet ./...` clean; `go.mod`/`go.sum` unchanged
- [ ] §13 Grepability contract passes: `out=$(./skilldozer 2>/dev/null)` ⇒ rc 0 + grep finds USAGE; `./skilldozer 2>&1 >/dev/null` ⇒ empty

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact current text of all five edit sites (with before/after), the line numbers verified against the live file, the two tests' full current bodies, the marker (`usageText` contains `USAGE:` at main.go:56), the collateral check (only two tests assert this behavior), and the §13 gate text are all specified below. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
- file: main.go
  why: "THE edit site. The no-mode fallthrough at :695-700 (stderr/return 1 -> stdout/return 0). usage() at :98 returns the usageText const (:52). The three doc comments to correct: usageText doc :48-51, run() exit-code doc :417-423, fallthrough comment :695-698. The fallthrough is the FINAL tail of run() (:428) — after every genuine-failure path."
  pattern: "Two-token behavior change (stderr->stdout, 1->0) + comment rewrites. The genuine-failure paths (~451 unknown flag, ~468 storeMissingValue, ~481 exclusivity, ~662 unresolved tag) are ABOVE the fallthrough and must NOT change."
  gotcha: "The fallthrough covers BOTH truly-no-args AND modifiers-only (the comment says so). Do not narrow it. Do not touch the --help branch (:435, already stdout/exit 0)."

- file: main_test.go
  why: "THE two tests to flip. TestRunDefaultNoArgs at :277-291 (run(nil)); TestRunModifiersOnlyNoMode at :1668-1684 (run([--no-color])). Both currently assert code==1 / stdout empty / stderr has USAGE — invert each to code==0 / stdout has USAGE / stderr empty. Harness: run([]string{...}, &out, &errOut) returns int; out.String()=stdout, errOut.String()=stderr."
  pattern: "Keep the existing assertion style (t.Errorf with %q, strings.Contains). Use marker \"USAGE\" (usageText contains USAGE: at main.go:56, so Contains matches)."
  gotcha: "Collateral grep confirmed ONLY these two tests assert no-args/modifiers-only behavior. run([]string{\"-a\"}) at :883 is --all (a real mode) — unaffected. Do not add new tests; flip the existing two."

- file: plan/003_3ace946c2a5c/architecture/code_change_map.md
  why: "§'Change A' pins the exact line numbers and the before/after for the flip + the three comment sites + the two tests. Verified against HEAD bbd4e74 (line numbers still match live HEAD 5efd3d9)."
  section: "Change A: No-args implicit help flip (Sites 1-3)"

- file: plan/003_3ace946c2a5c/architecture/test_patterns.md
  why: "Gives the TARGET assertions for both tests (code==0, stdout contains USAGE, stderr empty) and the regression-guard list (--help/-h stdout/exit0; ./skilldozer nope empty-stdout/exit1; all exclusivity tests unchanged)."
  section: "Tests that BREAK and must be rewritten; Regression guard"

- file: plan/003_3ace946c2a5c/P1M1T1S1/research/verified_facts.md
  why: "Direct-from-source proof: resolves the confusing git log (the flip is genuinely not applied yet), gives the exact current text of every site, the collateral check, and the §13 gate text."

- url: (PRD §6.3 / §6.4 / §19 decision 17 — in PRD.md, READ-ONLY)
  why: "§6.3: no-args => stdout/exit 0 (implicit --help). §6.4: bare no-args is NOT an error; the stderr/non-zero contract applies to genuine failures only. §19 decision 17: overrides the old stderr/exit-1 parity; only no-args is reclassified. Do NOT edit PRD.md."

- url: https://pkg.go.dev/fmt#Fprint
  why: "Confirms fmt.Fprint(w io.Writer, ...) takes the writer first — swapping stderr->stdout is a one-token change."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# main.go: 1126 lines. usageText const @ :52 (contains "USAGE:" header @ :56); usage() @ :98;
#          run() @ :428; --help branch @ :435 (already stdout/exit 0); no-mode fallthrough @ :695-700 (stderr/exit 1).
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# No new files. This subtask edits main.go and main_test.go only.
```

### Desired Codebase tree with files to be changed

```bash
main.go           # MODIFY — 2 behavior tokens (:699, :700) + 3 doc-comment rewrites (:48-51, :417-423, :695-698)
main_test.go      # MODIFY — flip 2 tests (TestRunDefaultNoArgs :277-291, TestRunModifiersOnlyNoMode :1668-1684)
# go.mod / go.sum — UNCHANGED (no new deps; stdout/stderr/usage() already in scope)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main.go` | Flip the no-mode fallthrough to stdout/exit 0; correct the 3 doc comments | PRD §6.3 / decision 17 |
| `main_test.go` | Flip the 2 no-args/modifiers-only tests to assert stdout/exit 0 | §13 Grepability contract |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — Flip BOTH tokens at :699-700, not just the writer. The fix is
// fmt.Fprint(stderr, usage()) -> fmt.Fprint(stdout, usage()) AND return 1 -> return 0.
// Flipping only the stream (and leaving return 1) breaks `$(...)` rc checks; flipping
// only the rc (leaving stderr) breaks the grep-pipe. They are one logical change.
//
// GOTCHA #2 — Do NOT touch the --help branch (:435). It is ALREADY fmt.Fprint(stdout,
// usage()); return 0. After this fix, --help and no-args converge on identical behavior
// (both stdout/exit 0) — that is exactly decision 17's intent ("implicit --help").
//
// GOTCHA #3 — The fallthrough covers BOTH truly-no-args AND modifiers-only (the comment
// says so; both reach it because no mode/tag is set). Do NOT add a separate branch for
// `--no-color`. TestRunModifiersOnlyNoMode proves the modifiers-only case is the SAME
// fallthrough.
//
// GOTCHA #4 — THREE doc comments describe the old behavior and ALL must be corrected,
// or the exit-code contract self-contradicts: the fallthrough comment (:695-698), the
// usageText doc (:48-51 — "to stderr for the no-args default (exit 1)"), and the run()
// exit-code list (:417-423 — exit-1 bullet ends "no recognized mode (usage to stderr)").
// Remove the stale exit-1 clause; add the no-args case to the exit-0 bullet.
//
// GOTCHA #5 — The two tests' CURRENT assertions (code==1, stdout EMPTY, stderr has USAGE)
// are the INVERSE of the target. This is a full flip, not an augmentation: invert every
// assertion (code==0, stdout has USAGE, stderr EMPTY). Leaving any old assertion in place
// will fail.
//
// GOTCHA #6 — Assert on "USAGE" (not "USAGE:") for consistency with the original test and
// the §13 case-insensitive grep. usageText contains "USAGE:" (main.go:56), so
// strings.Contains(out.String(), "USAGE") matches. Either form works; pick "USAGE".
//
// GOTCHA #7 — Keep every genuine-failure path + its test UNCHANGED. The no-mode flip is
// the ONLY behavior change. Verified paths: unknown flag (~:451 stderr/2), storeMissingValue
// (~:468 stderr/2), exclusivity (~:481 stderr/2), unresolved tag (~:662 stderr/1),
// unconfigured (stderr/1). Their tests (TestRunDefaultUnknownFlag, TestRunExclusivity*,
// TestRunBareTagUnconfiguredNeverPrompts, etc.) MUST stay green as-is.
//
// GOTCHA #8 — No deps change. fmt.Fprint/Fprintf, io.Writer, bytes.Buffer, strings are all
// stdlib and already imported. go.mod/go.sum byte-for-byte identical after this subtask.
//
// GOTCHA #9 — The git log is misleading: commit bbd4e74 says "flip bare-inv to implicit
// help" but the flip is NOT applied (main.go:699 is still stderr). Trust the LIVE code +
// the task tree (P1.M1.T1 is "Researching"), not the commit message. Verify with
// `grep -n 'fmt.Fprint(stderr, usage())' main.go` — it must return NOTHING after the fix.
```

---

## Implementation Blueprint

### Data models and structure

None. No structs, no signatures change. `run(args []string, stdout, stderr io.Writer) int` keeps its signature; the fix uses the `stdout` argument already in scope.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: FLIP the no-mode fallthrough (main.go:695-700)
  - EDIT main.go:699  fmt.Fprint(stderr, usage())  ->  fmt.Fprint(stdout, usage())
  - EDIT main.go:700  return 1                     ->  return 0
  - DO NOT touch main.go:435 (--help branch, already stdout/exit 0) (GOTCHA #2)
  - DO NOT touch any genuine-failure path above the fallthrough (GOTCHA #7)
  - NAMING/PLACEMENT: no renames/moves — same lines

Task 2: CORRECT the three doc comments (Mode A inline dev docs)
  - EDIT main.go:695-698 (fallthrough comment). Remove "usage to STDERR, exit 1 (PRD §6.3:
    parity with get-server-config.sh)" and the "stdout stays empty so $(...) never sees
    garbage" line. Rewrite to: no-mode => usage to STDOUT, exit 0 (PRD §6.3 / §19 decision
    17: bare invocation is implicit --help); covers truly-no-args AND modifiers-only; the
    help lands on stdout so `skilldozer | grep …` works; genuine failures stay on stderr (§6.4).
  - EDIT main.go:48-51 (usageText doc). Current: "The SAME text is printed to stdout for
    --help (exit 0) and to stderr for the no-args default (exit 1) — only the destination
    differs." -> "The SAME text is printed to stdout for --help (exit 0) and for the
    no-args/modifiers-only default (exit 0 — implicit --help, PRD §6.3 / §19 decision 17);
    genuine failures go to stderr (§6.4), never this text."
  - EDIT main.go:417-423 (run() exit-code doc). (a) Remove "no recognized mode (usage to
    stderr)" from the exit-1 bullet. (b) Add to the exit-0 bullet: "no-args/modifiers-only
    printed usage to stdout (implicit --help, §6.3)". Keep exit-2 bullet unchanged.
  - KEEP all three accurate to the new behavior; cite PRD §6.3 + decision 17.

Task 3: FLIP TestRunDefaultNoArgs (main_test.go:277-291)
  - EDIT the doc string (currently "to STDERR, exit 1 … parity with get-server-config.sh")
    to cite §6.3 / decision 17 (implicit --help, stdout, exit 0; §13 Grepability contract).
  - FLIP the three assertions:
      code:  `if code != 0` (was != 1); message "want 0 (no-args -> stdout usage, implicit --help)"
      stdout: `if !strings.Contains(out.String(), "USAGE")` (was out.Len()!=0); msg "want USAGE on stdout (§6.3)"
      stderr: `if errOut.Len() != 0` (was !Contains(errOut,"USAGE")); msg "want EMPTY (no-args writes nothing to stderr)"
  - KEEP the run(nil, &out, &errOut) call shape.

Task 4: FLIP TestRunModifiersOnlyNoMode (main_test.go:1668-1684)
  - EDIT the doc string to cite §6.3 / decision 17 (same as no-args; implicit --help).
  - FLIP the three assertions identically to Task 3, for run([]string{"--no-color"}):
      code==0; strings.Contains(out.String(), "USAGE"); errOut.Len()==0
  - KEEP the run([]string{"--no-color"}, ...) call shape.

Task 5: VERIFY in isolation + whole module + §13 gate
  - COMMAND: go build ./...                  (exit 0)
  - COMMAND: go vet ./...                    (clean)
  - COMMAND: go test ./...                   (all green — incl. every genuine-failure test)
  - COMMAND: go test -run 'TestRunDefaultNoArgs|TestRunModifiersOnlyNoMode' -v ./...
  - COMMAND: grep -n 'fmt.Fprint(stderr, usage())' main.go   (MUST print NOTHING post-fix — GOTCHA #9)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"
  - MANUAL: build ./skilldozer; run the §13 Grepability contract (see Level 4)
```

### Implementation Patterns & Key Details

```go
// Task 1 — the flip (before/after). ONLY the first arg + the return value change:

// BEFORE (main.go:699-700):
//	fmt.Fprint(stderr, usage())
//	return 1
// AFTER:
	fmt.Fprint(stdout, usage())
	return 0

// Task 3 — the test flip (before/after). Full inversion of every assertion:
// BEFORE (TestRunDefaultNoArgs):
//	if code != 1 { ... want 1 ... }
//	if out.Len() != 0 { ... want EMPTY (usage goes to stderr) ... }
//	if !strings.Contains(errOut.String(), "USAGE") { ... want USAGE block ... }
// AFTER:
	if code != 0 {
		t.Errorf("run(nil): code=%d; want 0 (no-args → stdout usage, implicit --help)", code)
	}
	if !strings.Contains(out.String(), "USAGE") {
		t.Errorf("run(nil) stdout=%q; want the USAGE block on stdout (§6.3)", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("run(nil) stderr=%q; want EMPTY (no-args writes nothing to stderr)", errOut.String())
	}
// (TestRunModifiersOnlyNoMode is the same inversion for run([]string{"--no-color"}).)
```

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports; stdout/stderr/usage()/strings all in scope. (GOTCHA #8)

DISPATCH (unchanged shape):
  - run() precedence is unchanged: help -> version -> unknownFlag -> storeMissingValue ->
    exclusivity -> init -> (path/list/search/check/all/tags) -> NO-MODE FALLTHROUGH.
    Only the fallthrough's stream/exit flips; nothing above it moves.

CONSUMERS (behavioral):
  - `skilldozer | grep …` now sees usage on stdout (the whole point of decision 17).
  - `pi --skill "$(skilldozer badtag)"` is UNAFFECTED — that is the unresolved-tag path
    (stderr/exit 1), not the no-mode fallthrough. §6.4 $(...) contract intact.

NO ROUTES / NO DATABASE / NO CONFIG:
  - Output-stream + exit-code change only. No discovery, no config, no exclusivity change.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after the edits)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go     # must print NOTHING
go vet ./...                       # expect exit 0
go build ./...                     # expect exit 0

# GOTCHA #9 proof: the stale stderr call must be GONE after the fix.
grep -n 'fmt.Fprint(stderr, usage())' main.go   # Expected: NO output (empty)
# Expected: zero gofmt/vet/build errors; the grep prints nothing.
```

### Level 2: The two flipped tests (the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test -run 'TestRunDefaultNoArgs|TestRunModifiersOnlyNoMode' -v ./...
# Expected: PASS. TestRunDefaultNoArgs: run(nil) -> code 0, USAGE on stdout, stderr empty.
#           TestRunModifiersOnlyNoMode: run([--no-color]) -> same.

# Prove the flipped assertions are load-bearing: temporarily revert main.go:699-700 to
# stderr/return 1, re-run -> BOTH tests MUST fail (code-mismatch / USAGE-on-stderr /
# stderr-not-empty). Restore the flip. (Confirms the tests actually guard the behavior.)
```

### Level 3: Whole-module regression (genuine-failure paths unchanged)

```bash
cd /home/dustin/projects/skilldozer

go test ./...   ; echo "test exit $?"
# Expected: exit 0. CRITICAL regression set that MUST stay green (they assert stderr + non-zero):
#   TestRunDefaultUnknownFlag       (unknown flag -> stderr/exit 2)
#   TestRunExclusivity*             (mutually-exclusive -> stderr/exit 2)
#   TestRunBareTagUnconfiguredNeverPrompts (unconfigured -> stderr/exit 1, stdout empty)
#   any unresolved-tag test         (badtag -> stderr/exit 1, stdout empty)
# If any of these flip green->red, you accidentally touched a genuine-failure path (GOTCHA #7).

# Dependency invariant:
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
# Expected: "deps unchanged".
```

### Level 4: The §13 "Grepability contract" (end-to-end, user-facing)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# §6.3 / §13: no-args help is on stdout, exit 0 — pipes MUST see it.
out=$(/tmp/sdz 2>/dev/null); rc=$?
[ "$rc" = "0" ]                                    && echo "rc OK"   || echo "FAIL: rc=$rc"
printf '%s' "$out" | grep -qi 'USAGE'              && echo "grep OK" || echo "FAIL: no USAGE on stdout"
# no-args writes NOTHING to stderr:
test -z "$(/tmp/sdz 2>&1 >/dev/null)"              && echo "stderr-empty OK" || echo "FAIL: no-args wrote to stderr"

# Modifiers-only behaves the same (no mode => implicit help):
out2=$(/tmp/sdz --no-color 2>/dev/null); rc2=$?
[ "$rc2" = "0" ] && printf '%s' "$out2" | grep -qi 'USAGE' && echo "modifiers-only OK" || echo "FAIL"

# CONTROL — genuine failures STILL stay on stderr / non-zero (§6.4 contract intact):
/tmp/sdz --frobnicate >/dev/null 2>&1; [ "$?" = "2" ] && echo "unknown-flag still exit 2" || echo "FAIL"
badout=$(/tmp/sdz nope 2>/dev/null); [ -z "$badout" ] && echo "badtag stdout still empty" || echo "FAIL"

rm -f /tmp/sdz
# Expected: every line prints "... OK" / "still ..."; no FAIL.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go vet ./...` exit 0, `go build ./...` exit 0; `grep 'fmt.Fprint(stderr, usage())' main.go` empty
- [ ] Level 2 PASS — both flipped tests pass; reverting the flip makes them fail (load-bearing)
- [ ] Level 3 PASS — `go test ./...` exit 0 (every genuine-failure test still green); `git diff go.mod go.sum` → "deps unchanged"
- [ ] Level 4 PASS — §13 Grepability contract: no-args rc 0 + USAGE on stdout + stderr empty; modifiers-only same; genuine failures unchanged

### Feature Validation
- [ ] main.go:699 writes `stdout`; main.go:700 returns `0`
- [ ] Three doc comments corrected (fallthrough :695-698, usageText :48-51, exit-code list :417-423) to no-args ⇒ stdout/exit 0 (§6.3/decision 17)
- [ ] `TestRunDefaultNoArgs` + `TestRunModifiersOnlyNoMode` flipped (code 0, USAGE on stdout, stderr empty)
- [ ] No genuine-failure path or its tests changed

### Code Quality / Convention Validation
- [ ] Two-token behavior change only (stream + return); no signature/logic restructuring
- [ ] Flipped tests keep the existing assertion style (`t.Errorf` + `%q`, `strings.Contains`, marker `"USAGE"`)
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical

### Scope Discipline
- [ ] Did NOT add the `completion` subcommand (P1.M2; absent from current code)
- [ ] Did NOT edit the README (Mode B is P1.M3.T1; §15 outline doesn't mention no-args behavior)
- [ ] Did NOT touch unknown-flag / exclusivity / unresolved-tag / unconfigured paths
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't flip only one token.** Both the writer (`stderr`→`stdout`) AND the return (`1`→`0`) must change together — they are one logical change (decision 17).
- ❌ **Don't touch the `--help` branch** (:435). It's already stdout/exit 0; after this fix it and no-args converge — that's the intent.
- ❌ **Don't add a separate modifiers-only branch.** The fallthrough already covers `--no-color`-with-no-mode (the comment says so; `TestRunModifiersOnlyNoMode` proves it).
- ❌ **Don't leave any stale doc comment.** Three comments describe stderr/exit-1; fix all three or the exit-code contract contradicts itself.
- ❌ **Don't augment the tests — invert them.** The old assertions (code==1, stdout empty, stderr has USAGE) are the exact inverse of the target. Keep one old assertion and the test fails.
- ❌ **Don't touch any genuine-failure path.** Only the no-mode fallthrough flips. If `TestRunDefaultUnknownFlag`, an exclusivity test, or the unconfigured test goes red, you overreached.
- ❌ **Don't trust the git log.** Commit bbd4e74 claims the flip is done, but `main.go:699` is still `stderr`. Trust the live code + the task tree; prove the flip with `grep`.
- ❌ **Don't add deps or imports.** Everything needed (stdout, stderr, usage(), strings) is already in scope.

---

## Confidence Score

**9/10** — Every edit site is pinned to a verified live line number with exact before/after text; the fix is a two-token stream/exit swap plus a full inversion of two tests, all confirmed against the live source and the §13 gate. The flipped tests are provably load-bearing (they currently assert the exact inverse and pass on the un-fixed code; after the flip they assert the new behavior). The 1-point reservation is the misleading git log (commit bbd4e74's message vs. the actual un-fixed code) — resolved by verifying `main.go:699` directly and adding a `grep` proof to Level 1 so the implementer cannot be fooled by it.
