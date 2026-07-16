# PRP — P1.M1.T2.S1: Assert bash + zsh emitted output carry the §14.7 list-ambiguous option + opt-out

> **Subtask:** P1.M1.T2.S1 — the test half of P1.M1.T2 (§14.7: list every ambiguous match on the first Tab). Adds/extends tests in `main_test.go` that lock the emitted-byte contract produced by the parallel **T1.S1** (bash, Complete) + **T1.S2** (zsh, in current code): the emitted `--completions` scripts carry the active list-ambiguous option **and** the disclosed opt-out token.
> **Scope boundary:** **Test-only.** Edits ONLY `main_test.go` — one new bash test + two extended zsh tests. Does NOT touch `main.go` (the `zshEvalRegistration` const is T1.S2's, already present), `completions/*` (T1.S1's, already present), `README.md` (the §14.7 disclosure is P1.M1.T3.S1), or any production code. The option strings already exist; this subtask just locks them with assertions.

---

## Goal

**Feature Goal**: Lock PRD §14.7's emitted-byte contract for bash + zsh with automated tests: the script emitted by `skilldozer --completions` (bash verbatim; zsh derived) MUST contain the active list-ambiguous option (`show-all-if-ambiguous on` / `setopt NO_LIST_AMBIGUOUS`) AND the disclosed opt-out token (`show-all-if-ambiguous off` / `setopt LIST_AMBIGUOUS`). This prevents a future regression that silently drops the option (which would reintroduce the §14.7 silent-halt defect).

**Deliverable**: Edits to `main_test.go` only:
1. **NEW** `TestRunCompletionBashListsAmbiguous` — runs `--completions --shell bash`, asserts stdout contains `show-all-if-ambiguous on` + `show-all-if-ambiguous off` (+ the `*i*` guard).
2. **EXTEND** `TestZshEvalScriptRegistersCompdef` (3288) — add `setopt NO_LIST_AMBIGUOUS` + `setopt LIST_AMBIGUOUS` to its want-loop.
3. **EXTEND** `TestRunCompletionZshIsEvalSafe` (3316) — assert the end-to-end zsh output contains `NO_LIST_AMBIGUOUS`.

**Success Definition**: `go test ./...` 100% green (no existing test regresses); `TestEmbeddedCompletionsMatchOnDisk` (3139) stays GREEN (untouched); the new/extended assertions provably catch a regression (removing the option line from `completions/skilldozer.bash` or `zshEvalRegistration` makes them fail); `main.go`/`completions/*`/`README.md` byte-identical; `go.mod`/`go.sum` unchanged.

---

## User Persona (if applicable)

**Not applicable directly** — this is a test-only change with no user-facing surface. The eventual beneficiary is the maintainer: if a future edit accidentally drops the §14.7 option from the emitted scripts, these tests fail loudly instead of shipping the silent-halt regression.

---

## Why

- **PRD §14.7** mandates the emitted scripts set the list-ambiguous option (so the first Tab shows every match, not a silent halt at the common prefix), **disclose** it, and provide a one-line **opt-out**. The code half (T1.S1/T1.S2) shipped the option lines; the test half (this) locks them so they can't silently regress.
- **The store is manifest-free (§2)** → completion is the primary discovery path. A regression that hides ambiguous candidates defeats the whole point. A byte-level assertion is the cheapest possible guard.
- **Touch point 5 (code_change_map.md)** explicitly scopes this exact test set; **system_context.md 'Existing test patterns'** prescribes "byte-level assertion on emitted scripts (active option + disclosed opt-out token)" mirroring the existing `TestZshEvalScript*` / `TestRunCompletion*Script` pattern.

---

## What

Three test edits in `main_test.go` (all using the existing `run()`/`completionScript()`/`zshEvalScript()` helpers — no new harness):
- **BASH** — add `TestRunCompletionBashListsAmbiguous`: `run([]string{"--completions","--shell","bash"}, &out, &errOut)` → assert code 0, errOut empty, then `out.String()` CONTAINS `show-all-if-ambiguous on` AND `show-all-if-ambiguous off` (+ the `*i*` interactivity guard).
- **ZSH unit** — extend `TestZshEvalScriptRegistersCompdef` (3288): add `"setopt NO_LIST_AMBIGUOUS"` (active) and `"setopt LIST_AMBIGUOUS"` (opt-out) to its existing `for _, want := range` loop.
- **ZSH e2e** — extend `TestRunCompletionZshIsEvalSafe` (3316): under `t.Setenv("SKILLDOZER_SHELL","zsh")`, assert `out.String()` CONTAINS `NO_LIST_AMBIGUOUS`.

### Success Criteria

- [ ] `TestRunCompletionBashListsAmbiguous` asserts both `show-all-if-ambiguous on` and `show-all-if-ambiguous off` (and `*i*`) in the emitted bash script
- [ ] `TestZshEvalScriptRegistersCompdef` (3288) asserts `setopt NO_LIST_AMBIGUOUS` + `setopt LIST_AMBIGUOUS` in `zshEvalScript(completionScript("zsh"))`
- [ ] `TestRunCompletionZshIsEvalSafe` (3316) asserts `NO_LIST_AMBIGUOUS` in the end-to-end emitted zsh script
- [ ] `TestEmbeddedCompletionsMatchOnDisk` (3139) stays GREEN (untouched — not weakened)
- [ ] `go test ./...` 100% green; `main.go`/`completions/*`/`README.md` byte-identical; `go.mod`/`go.sum` unchanged
- [ ] The new assertions are load-bearing: removing the option line makes them fail (verify in Level 2)

---

## All Needed Context

### Context Completeness Check

**Pass.** Every test extension is pinned to its exact current body with before→after text in `research/verified_facts.md` §3; the token precision (§1 — non-overlapping assertions), the harness (§2), the byte-identity invariant (§3), and the scope boundary (§6) are all documented. The option strings are already present in the code (T1.S1 done + T1.S2 in current tree — §0), so the tests pass against the current code. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified test inventory (current bodies + before/after + token precision)
- file: plan/005_d9b30e368811/P1M1T2S1/research/verified_facts.md
  why: "THE source of truth. §0 proves the option strings are ALREADY present (test-only subtask). §1 proves the assertion tokens are precise/non-overlapping (setopt NO_LIST_AMBIGUOUS does NOT contain setopt LIST_AMBIGUOUS; bash lowercase on/off don't match the uppercase disclosure comments). §3 gives the exact current bodies of the 3 tests + the extension plan. §4 is the authoritative Touch-point-5 spec."
  critical: "§1 (token precision) and §3 (the byte-identity invariant) prevent the two most likely errors: an assertion that false-matches the wrong line, and accidentally weakening TestEmbeddedCompletionsMatchOnDisk."

# MUST READ — the authoritative test spec (mirrors the contract)
- file: plan/005_d9b30e368811/architecture/code_change_map.md
  why: "Touch point 5 pins the exact test edits: bash add/extend for show-all-if-ambiguous on/off (+optional *i*); zsh extend TestZshEvalScriptRegistersCompdef for setopt NO_LIST_AMBIGUOUS + setopt LIST_AMBIGUOUS; zsh extend TestRunCompletionZshIsEvalSafe for NO_LIST_AMBIGUOUS; TestEmbeddedCompletionsMatchOnDisk stays green."
  section: "Touch point 5 — Tests (byte-level emitted-script assertions)."

# MUST READ — the test-pattern map (which tests to extend + why byte-level)
- file: plan/005_d9b30e368811/architecture/system_context.md
  why: "'Existing test patterns' table maps the 6 completion tests + explicitly says EXTEND TestZshEvalScriptRegistersCompdef and TestRunCompletionZshIsEvalSafe for NO_LIST_AMBIGUOUS. States the strategy: byte-level assertion on emitted scripts (active option + disclosed opt-out), mirroring the existing TestZshEvalScript*/TestRunCompletion*Script pattern. Notes there is NO in-process way to drive a live shell's first-Tab behavior."
  section: "'Existing test patterns to mirror' + 'Out of scope'."

# MUST READ — the sibling PRP (T1.S2, zsh const) — confirms the exact target tokens
- file: plan/005_d9b30e368811/P1M1T1S2/PRP.md
  why: "T1.S2 (zsh) emits setopt NO_LIST_AMBIGUOUS (active) + #   setopt LIST_AMBIGUOUS (opt-out) in the zshEvalRegistration const. Its GOTCHA #8 says: 'Keep the opt-out token literally setopt LIST_AMBIGUOUS (commented) and the active line literally setopt NO_LIST_AMBIGUOUS. The P1.M1.T2.S1 test greps for both substrings.' This subtask IS that test. The current code (HEAD 5cf81d4) already matches T1.S2's contract."
  pattern: "T2.S1 locks exactly the tokens T1.S1 (bash) + T1.S2 (zsh) emit. Treat their exact strings as the contract: show-all-if-ambiguous on/off (bash), setopt NO_LIST_AMBIGUOUS / setopt LIST_AMBIGUOUS (zsh)."

# MUST READ — the edit target (read the 3 tests + helpers before editing)
- file: main_test.go
  why: "THE edit target. TestEmbeddedCompletionsMatchOnDisk @3139 (INVARIANT — don't touch). TestRunCompletionBashScript @3163 (the §13 marker test; add a DEDICATED sibling test rather than mixing). TestZshEvalScriptRegistersCompdef @3288 (EXTEND the want-loop). TestRunCompletionZshIsEvalSafe @3316 (EXTEND with one Contains). Helpers: run() in main.go:524, completionScript() main.go:1215, zshEvalScript() main.go:1245 — all called directly (package main)."
  pattern: "Assertion idiom: `strings.Contains(out.String(), want)` + `t.Errorf(\"...:\\n%s\", out.String())`. The want-loop pattern in TestZshEvalScriptRegistersCompdef is the model for the zsh-unit extension."
  gotcha: "strings.Contains is CASE-SENSITIVE. The bash disclosure comments use uppercase OFF/ON (lines 73/76) which do NOT match the lowercase assertion tokens 'on'/'off' — so the assertions are precise to the active (83) / opt-out (85) lines."

# READ-ONLY — the PRD authority
- file: PRD.md
  why: "READ-ONLY. §14.7 mandates: emitted scripts SET the option (show-all-if-ambiguous on / setopt NO_LIST_AMBIGUOUS), DISCLOSE it, and provide a one-line OPT-OUT (bind 'set show-all-if-ambiguous off' / setopt LIST_AMBIGUOUS). §14.6 (bash emitted verbatim; zsh derived via zshEvalScript). These define what the tests lock. Do NOT edit PRD.md."
  section: "h3.21 (§14.7), h2.13 (§14), h2.12 (§13 acceptance)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/005_d9b30e368811/tasks.json
  why: "P1.M1.T2.S1's CONTRACT block is authoritative INPUT/LOGIC/OUTPUT. This PRP transcribes it; tasks.json wins on conflict."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD && wc -l main_test.go
5cf81d4
~3700 main_test.go
$ go test ./...   # green (cached); the option strings are present so the new tests will pass
# main.go: zshEvalRegistration const carries setopt NO_LIST_AMBIGUOUS (1292) + #   setopt LIST_AMBIGUOUS (1291).
# completions/skilldozer.bash: bind 'set show-all-if-ambiguous on' (83) + #   bind '... off' (85).
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# Helpers: run() main.go:524, completionScript() main.go:1215, zshEvalScript() main.go:1245.
# T2.S1 edits ONLY main_test.go (no new files).
```

### Desired Codebase tree with files to be changed

```bash
main_test.go   # MODIFY — +1 new bash test (TestRunCompletionBashListsAmbiguous); +2 want-loop entries (zsh unit); +1 Contains (zsh e2e)
# main.go / completions/* / README.md — UNCHANGED (test-only subtask)
# go.mod / go.sum — UNCHANGED (no new imports; bytes/strings/os already imported)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main_test.go` (new bash test) | `TestRunCompletionBashListsAmbiguous` — §14.7 active + opt-out + guard | Contract LOGIC a |
| `main_test.go` (zsh unit) | extend `TestZshEvalScriptRegistersCompdef` want-loop | Contract LOGIC b |
| `main_test.go` (zsh e2e) | extend `TestRunCompletionZshIsEvalSafe` | Contract LOGIC c |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — This is TEST-ONLY. The option strings (show-all-if-ambiguous on/off in
// completions/skilldozer.bash; setopt NO_LIST_AMBIGUOUS + #   setopt LIST_AMBIGUOUS in main.go's
// zshEvalRegistration) are ALREADY PRESENT (T1.S1 done, T1.S2 in current code). T2.S1 ONLY adds
// assertions to main_test.go. Do NOT edit main.go, completions/*, or README.md — that collides
// with T1.S1/T1.S2 (done) and P1.M1.T3.S1 (README disclosure). (research §0, §6.)

// GOTCHA #2 — The zsh assertions are PRECISE and non-overlapping. `setopt NO_LIST_AMBIGUOUS` does
// NOT contain the substring `setopt LIST_AMBIGUOUS` (verified: grep finds no match). So asserting
// BOTH works: the active line (1292) matches `setopt NO_LIST_AMBIGUOUS` only; the opt-out comment
// (1291) matches `setopt LIST_AMBIGUOUS` only. Do NOT "improve" the opt-out assertion to the full
// comment line — the bare token is what the contract specifies and what's robust to comment-indent
// tweaks. (research §1.)

// GOTCHA #3 — strings.Contains is CASE-SENSITIVE. The bash disclosure comments (lines 73/76) use
// UPPERCASE "OFF"/"ON"; the assertion tokens are lowercase "on"/"off". So `show-all-if-ambiguous on`
// matches ONLY line 83 (the active bind) and `show-all-if-ambiguous off` ONLY line 85 (the opt-out).
// Do NOT uppercase the assertion tokens or they'll false-match the disclosure comments (and miss a
// regression that drops the active/opt-out lines while keeping the prose).

// GOTCHA #4 — For the zsh e2e test, assert the SHORTER token `NO_LIST_AMBIGUOUS` (per contract LOGIC c),
// not the full `setopt NO_LIST_AMBIGUOUS`. The shorter token is present in the active line and is
// what the contract prescribes for the end-to-end assertion. (The unit test uses the full tokens.)

// GOTCHA #5 — TestEmbeddedCompletionsMatchOnDisk (3139) MUST stay GREEN and UNTOUCHED. It compares
// completionScript(shell) to the on-disk file; for zsh that's the autoload (the const is appended
// AFTER completionScript returns, in zshEvalScript), so it's unaffected by the §14.7 option. For
// bash the embed == on-disk file. T2.S1 does not touch production code, so this test cannot regress.
// Do NOT weaken it (e.g. removing the zsh case) to "make room" for the new assertions — it is an
// independent invariant. (research §3; Touch point 5 'Invariant'.)

// GOTCHA #6 — Prefer a DEDICATED bash test (TestRunCompletionBashListsAmbiguous) over mixing §14.7
// into TestRunCompletionBashScript (the §13 marker test). The dedicated test isolates the §14.7
// contract and is self-documenting. (Extending TestRunCompletionBashScript is an acceptable
// alternative per Touch point 5, but a dedicated test is cleaner; the contract lists it first.)

// GOTCHA #7 — Use the EXISTING helpers, do not invent new ones. run(args, &out, &errOut) (main.go:524)
// returns the exit code; out/errOut are bytes.Buffer. completionScript("zsh") (main.go:1215) returns
// the autoload verbatim; zshEvalScript(raw) (main.go:1245) returns the DERIVED wrapper. Mirror the
// existing assertion idiom (strings.Contains + t.Errorf with the full output for debuggability).

// GOTCHA #8 — Prove the assertions are load-bearing (Level 2). After adding them, temporarily
// comment out the active `bind 'set show-all-if-ambiguous on'` line in completions/skilldozer.bash
// (or the setopt in zshEvalRegistration) and re-run — the new test(s) MUST fail. Then restore. This
// confirms the test actually catches the regression it exists to catch. (A vacuous test that passes
// regardless is worse than no test.)

// GOTCHA #9 — No deps change. main_test.go already imports bytes/fmt/os/strings/testing. The new
// test + extensions add no imports. go.mod/go.sum byte-for-byte identical.

// GOTCHA #10 — Rebuild is automatic for `go test` (it compiles main.go + main_test.go together).
// The option strings are in the compiled const/on-disk file; `go test` reads the current source.
// No separate `go build` needed for the tests (though Level 1 runs it as a sanity gate).
```

---

## Implementation Blueprint

### Data models and structure

**None.** Test-only; no types, no production code, no signatures change.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: ADD TestRunCompletionBashListsAmbiguous (main_test.go, near TestRunCompletionBashScript @3163)
  - PLACEMENT: insert right after TestRunCompletionBashScript (3163-3177) or alongside the other
    TestRunCompletion* tests. (Dedicated test — GOTCHA #6; do NOT mix into the §13 marker test.)
  - BODY (mirror the existing run()+strings.Contains idiom):
        // TestRunCompletionBashListsAmbiguous locks PRD §14.7 for bash: the emitted script sets
        // show-all-if-ambiguous ON (first Tab lists every prefix match, not a silent halt at the
        // common prefix), discloses the change, and provides the one-line opt-out. bash is emitted
        // verbatim, so this also covers the on-disk completions/skilldozer.bash.
        func TestRunCompletionBashListsAmbiguous(t *testing.T) {
        	var out, errOut bytes.Buffer
        	code := run([]string{"--completions", "--shell", "bash"}, &out, &errOut)
        	if code != 0 {
        		t.Fatalf("run(completion --shell bash): code=%d; want 0", code)
        	}
        	if errOut.Len() != 0 {
        		t.Fatalf("stderr=%q; want empty on success", errOut.String())
        	}
        	script := out.String()
        	// §14.7 active option: list all ambiguous matches on the first Tab.
        	if !strings.Contains(script, "show-all-if-ambiguous on") {
        		t.Errorf("bash completion missing the §14.7 active 'show-all-if-ambiguous on':\n%s", script)
        	}
        	// §14.7 disclosed opt-out: a user can restore stock bash behavior after eval.
        	if !strings.Contains(script, "show-all-if-ambiguous off") {
        		t.Errorf("bash completion missing the §14.7 opt-out 'show-all-if-ambiguous off':\n%s", script)
        	}
        	// The interactivity guard (bash's `bind` warns when sourced non-interactively).
        	if !strings.Contains(script, "*i*") {
        		t.Errorf("bash completion missing the '*i*' interactivity guard for bind:\n%s", script)
        	}
        }
  - GOTCHA #3: lowercase "on"/"off" tokens (case-sensitive Contains — precise to lines 83/85).

Task 2: EXTEND TestZshEvalScriptRegistersCompdef (main_test.go:3288) — add 2 want-entries
  - EDIT the want-slice in the existing `for _, want := range []string{...}` loop. CURRENT:
        for _, want := range []string{
        	"autoload -Uz compinit",
        	"(( $+functions[compdef] )) || compinit", // bootstrap only if compdef absent
        	"compdef _skilldozer skilldozer",         // explicit registration
        } {
  - ADD two entries (§14.7 active + opt-out):
        for _, want := range []string{
        	"autoload -Uz compinit",
        	"(( $+functions[compdef] )) || compinit", // bootstrap only if compdef absent
        	"compdef _skilldozer skilldozer",         // explicit registration
        	"setopt NO_LIST_AMBIGUOUS",               // §14.7 active: list all ambiguous matches on first Tab
        	"setopt LIST_AMBIGUOUS",                  // §14.7 opt-out (commented in the script; substring present)
        } {
  - GOTCHA #2: the two tokens are non-overlapping (setopt NO_LIST_AMBIGUOUS does NOT contain
    setopt LIST_AMBIGUOUS). The existing #compdef/_arguments Contains checks below the loop stay.

Task 3: EXTEND TestRunCompletionZshIsEvalSafe (main_test.go:3316) — add one Contains
  - EDIT the test body. After the existing compdef-registration assertion:
        if !strings.Contains(script, "compdef _skilldozer skilldozer") {
        	t.Errorf("zsh eval output missing compdef registration:\n%s", script)
        }
  - ADD (after it, before the on-disk-inequality check):
        // §14.7: the derived wrapper sets NO_LIST_AMBIGUOUS so the first Tab lists every prefix
        // match instead of halting silently at the common prefix.
        if !strings.Contains(script, "NO_LIST_AMBIGUOUS") {
        	t.Errorf("zsh eval output missing the §14.7 NO_LIST_AMBIGUOUS listing option:\n%s", script)
        }
  - GOTCHA #4: assert the shorter token `NO_LIST_AMBIGUOUS` (present in the active line
    `setopt NO_LIST_AMBIGUOUS`), per contract LOGIC c.

Task 4: VERIFY — build, the full suite, and the load-bearing proof
  - COMMAND: go build ./...   (exit 0 — sanity; main.go unchanged)
  - COMMAND: go test ./...    (100% green — no regression; the new/extended tests pass)
  - COMMAND: go test -run 'TestRunCompletionBashListsAmbiguous|TestZshEvalScriptRegistersCompdef|TestRunCompletionZshIsEvalSafe|TestEmbeddedCompletionsMatchOnDisk' -v ./...
                                (the 4 relevant tests PASS)
  - LOAD-BEARING PROOF (GOTCHA #8): temporarily comment out the active `bind 'set show-all-if-ambiguous on'`
    line in completions/skilldozer.bash → re-run TestRunCompletionBashListsAmbiguous → it MUST fail
    (missing active token). Restore the line. (Optionally repeat for the zsh setopt in
    zshEvalRegistration.) This proves the assertions catch a real regression.
  - COMMAND: git diff --quiet main.go completions/ README.md && echo "production unchanged"
    (MUST print "production unchanged" — GOTCHA #1)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"
  - COMMAND: git diff --name-only   (MUST be ONLY main_test.go)
```

### Implementation Patterns & Key Details

```go
// Task 2 — the want-loop extension (the idiomatic pattern in this file):
for _, want := range []string{
	"autoload -Uz compinit",
	"(( $+functions[compdef] )) || compinit",
	"compdef _skilldozer skilldozer",
	"setopt NO_LIST_AMBIGUOUS", // §14.7 active
	"setopt LIST_AMBIGUOUS",    // §14.7 opt-out (commented; substring present)
} {
	if !strings.Contains(got, want) {
		t.Errorf("zshEvalScript: missing %q in eval output:\n%s", want, got)
	}
}
// The existing per-test #compdef / _arguments Contains checks (below the loop) are UNCHANGED.

// Task 3 — the e2e assertion (shorter token, per contract):
if !strings.Contains(script, "NO_LIST_AMBIGUOUS") {
	t.Errorf("zsh eval output missing the §14.7 NO_LIST_AMBIGUOUS listing option:\n%s", script)
}
```

Notes easy to get wrong:
- The bash test is DEDICATED (not mixed into the §13 marker test) — keeps each contract isolated and debuggable.
- The zsh-unit uses the FULL tokens (`setopt NO_LIST_AMBIGUOUS` / `setopt LIST_AMBIGUOUS`); the zsh-e2e uses the SHORTER token (`NO_LIST_AMBIGUOUS`). This asymmetry is intentional — the contract specifies it (unit locks the exact active/opt-out pair; e2e is a coarser smoke that the derived output carries the option).
- `*i*` is asserted in the bash test as the interactivity guard (bash's `bind` warns non-interactively, so the script gates it with `[[ $- == *i* ]]`). zsh's `setopt` needs no guard (silent non-interactively — T1.S2's decision), so no zsh equivalent.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Dedicated bash test vs extend TestRunCompletionBashScript? → DEDICATED (TestRunCompletionBashListsAmbiguous).** Touch point 5 allows either; the contract lists the dedicated test first. Isolating §14.7 from §13 keeps each test's failure mode clear and the file self-documenting. (Extending is acceptable but mixes concerns.)
2. **Unit tokens vs e2e token? → unit uses full `setopt NO_LIST_AMBIGUOUS`/`setopt LIST_AMBIGUOUS`; e2e uses shorter `NO_LIST_AMBIGUOUS`.** The contract specifies both forms; the unit test locks the exact active/opt-out pair, the e2e is a coarser end-to-end smoke.
3. **Assert the `*i*` guard in bash? → YES (optional per contract, included).** It documents the bash-specific interactivity gating and catches a regression that drops the guard (which would make `eval` warn non-interactively). No zsh equivalent (setopt is silent).

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports (bytes/strings/os already in main_test.go). (GOTCHA #9)

TEST SUITE:
  - The 3 edits are ADDITIVE (1 new test + 2 extensions). No existing assertion is removed or weakened.
  - TestEmbeddedCompletionsMatchOnDisk (3139) is UNTOUCHED and stays GREEN (the byte-identity lock
    is independent of the §14.7 option — see GOTCHA #5).

NO PRODUCTION CODE / NO COMPLETIONS / NO README:
  - T2.S1 edits ONLY main_test.go. The option strings (T1.S1 bash + T1.S2 zsh) are already present;
    this subtask only locks them. (GOTCHA #1)
```

---

## Validation Loop

### Level 1: Build + edit presence (immediate)

```bash
cd /home/dustin/projects/skilldozer

go build ./...   # expect exit 0 (sanity — main.go is unchanged)
# Confirm the option strings the tests assert against are present in the current code:
grep -q 'show-all-if-ambiguous on'  completions/skilldozer.bash && echo "bash active present"  || echo "FAIL"
grep -q 'show-all-if-ambiguous off' completions/skilldozer.bash && echo "bash opt-out present" || echo "FAIL"
grep -q 'setopt NO_LIST_AMBIGUOUS'  main.go && echo "zsh active present"  || echo "FAIL"
grep -q 'setopt LIST_AMBIGUOUS'     main.go && echo "zsh opt-out present" || echo "FAIL"
# Expected: all four "present". (If any is absent, the code half T1.S1/T1.S2 isn't landed yet —
# this subtask assumes it is; the tests would then fail until it lands.)
```

### Level 2: The full test suite + the load-bearing proof (the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test ./... ; echo "test exit $?"   # Expected: 0 (100% green; the new/extended tests pass)

# The 4 relevant tests pass:
go test -run 'TestRunCompletionBashListsAmbiguous|TestZshEvalScriptRegistersCompdef|TestRunCompletionZshIsEvalSafe|TestEmbeddedCompletionsMatchOnDisk' -v ./...
# Expected: all 4 PASS. Critically TestEmbeddedCompletionsMatchOnDisk stays GREEN (GOTCHA #5).

# LOAD-BEARING PROOF (GOTCHA #8) — prove the bash assertion catches a regression:
cp completions/skilldozer.bash /tmp/skb.bak
sed -i "/bind 'set show-all-if-ambiguous on'/d" completions/skilldozer.bash
go test -run TestRunCompletionBashListsAmbiguous -v ./...   # Expected: FAIL ("missing active")
cp /tmp/skb.bak completions/skilldozer.bash && rm /tmp/skb.bak
go test -run TestRunCompletionBashListsAmbiguous -v ./...   # Expected: PASS (restored)
# (Optionally repeat for zsh: comment the `setopt NO_LIST_AMBIGUOUS` line in main.go's
#  zshEvalRegistration → TestZshEvalScriptRegistersCompdef + TestRunCompletionZshIsEvalSafe FAIL → restore.)
# Expected: the test FAILS when the active line is removed, PASSES when restored. This proves the
# assertion is load-bearing (not vacuous).
```

### Level 3: Whole-module regression + scope invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # Expected: 0
go vet  ./...  ; echo "vet exit $?"     # Expected: 0
go test ./...  ; echo "test exit $?"    # Expected: 0

# GOTCHA #1/#9 invariants (production untouched, deps unchanged):
git diff --quiet main.go completions/ README.md && echo "production unchanged" || echo "FAIL: production changed"
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
git diff --name-only   # Expected: ONLY main_test.go
# Expected: "production unchanged" + "deps unchanged" + only main_test.go in the name list.
```

### Level 4: End-to-end behavioral smoke (the actual emitted bytes)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# bash emit carries the active option + opt-out:
bash_out=$(/tmp/sdz --completions --shell bash 2>/dev/null)
echo "$bash_out" | grep -q 'show-all-if-ambiguous on'  && echo "bash emit active OK"  || echo "FAIL"
echo "$bash_out" | grep -q 'show-all-if-ambiguous off' && echo "bash emit opt-out OK" || echo "FAIL"

# zsh emit (derived wrapper) carries the active option + opt-out:
zsh_out=$(/tmp/sdz --completions --shell zsh 2>/dev/null)
echo "$zsh_out" | grep -q 'setopt NO_LIST_AMBIGUOUS' && echo "zsh emit active OK"  || echo "FAIL"
echo "$zsh_out" | grep -q 'setopt LIST_AMBIGUOUS'    && echo "zsh emit opt-out OK" || echo "FAIL"

# The autoload file is UNCHANGED (the option is ONLY in the derived wrapper, not completions/_skilldozer):
grep -q 'NO_LIST_AMBIGUOUS' completions/_skilldozer && echo "FAIL: autoload changed" || echo "autoload unchanged OK"
rm -f /tmp/sdz
# Expected: every line "...OK". (These mirror exactly what the Go tests now assert programmatically.)
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `go build` exit 0; all 4 option strings present in the current code
- [ ] Level 2 PASS — `go test ./...` 100% green; the 4 relevant tests PASS; the load-bearing proof (removing the active line makes the test FAIL) holds
- [ ] Level 3 PASS — build+vet+test exit 0; `git diff main.go completions/ README.md` → "production unchanged"; `git diff go.mod go.sum` → "deps unchanged"; only `main_test.go` changed
- [ ] Level 4 PASS — emitted bash + zsh output carry active + opt-out; autoload file unchanged

### Feature Validation
- [ ] `TestRunCompletionBashListsAmbiguous` asserts `show-all-if-ambiguous on` + `show-all-if-ambiguous off` + `*i*`
- [ ] `TestZshEvalScriptRegistersCompdef` (3288) asserts `setopt NO_LIST_AMBIGUOUS` + `setopt LIST_AMBIGUOUS`
- [ ] `TestRunCompletionZshIsEvalSafe` (3316) asserts `NO_LIST_AMBIGUOUS`
- [ ] `TestEmbeddedCompletionsMatchOnDisk` (3139) stays GREEN and untouched

### Code Quality / Convention Validation
- [ ] Uses the existing `run()`/`completionScript()`/`zshEvalScript()` helpers + the `strings.Contains` + `t.Errorf("...:\n%s", ...)` idiom
- [ ] Tokens match the contract exactly (lowercase bash on/off; full zsh setopt tokens in unit, shorter in e2e)
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] Minimal diff (1 new test + 2 small extensions)

### Scope Discipline (the test-only boundary)
- [ ] Did NOT touch `main.go` (the zshEvalRegistration const — T1.S2's, present)
- [ ] Did NOT touch `completions/*` (T1.S1's bash file / autoload / fish — present)
- [ ] Did NOT touch `README.md` (the §14.7 disclosure is P1.M1.T3.S1, Mode B)
- [ ] Did NOT weaken `TestEmbeddedCompletionsMatchOnDisk`
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't edit production code.** This is test-only. The option strings are already shipped (T1.S1/T1.S2). Editing main.go/completions/README collides with those siblings. (GOTCHA #1)
- ❌ **Don't weaken `TestEmbeddedCompletionsMatchOnDisk`.** It's an independent byte-identity invariant; the §14.7 option lives in the derived zsh wrapper (appended after completionScript returns) and the bash on-disk file. It stays GREEN untouched. (GOTCHA #5)
- ❌ **Don't uppercase the bash assertion tokens.** `strings.Contains` is case-sensitive; the disclosure comments (uppercase OFF/ON) would false-match while the active/opt-out lines (lowercase on/off) are missed. Use lowercase `on`/`off`. (GOTCHA #3)
- ❌ **Don't conflate `setopt NO_LIST_AMBIGUOUS` with `setopt LIST_AMBIGUOUS`.** They are non-overlapping (verified); assert BOTH — one catches a dropped active line, the other catches a dropped opt-out. (GOTCHA #2)
- ❌ **Don't skip the load-bearing proof.** A test that passes regardless of the option's presence is vacuous. Temporarily remove the active line, confirm the test fails, restore. (GOTCHA #8)
- ❌ **Don't mix §14.7 into the §13 marker test.** Add a dedicated `TestRunCompletionBashListsAmbiguous` (cleaner; contract lists it first). (GOTCHA #6)
- ❌ **Don't invent new helpers.** Use the existing `run()`/`completionScript()`/`zshEvalScript()` + `strings.Contains` idiom. (GOTCHA #7)
- ❌ **Don't add deps.** main_test.go already imports bytes/strings/os/testing. go.mod/go.sum byte-for-byte identical. (GOTCHA #9)

---

## Confidence Score

**9.5/10** — This is the smallest possible subtask: 3 additive test edits in one file, with every assertion pinned to a verified-present option string and proven non-overlapping/precise (`research/verified_facts.md` §1). The test harness (`run`/`completionScript`/`zshEvalScript`) and the exact current bodies of the 3 tests are read directly from source (§2/§3). The option strings are already present (T1.S1 done + T1.S2 in current HEAD `5cf81d4`), so the tests pass against the current code, and the load-bearing proof (Level 2) confirms they catch a real regression. The byte-identity invariant (`TestEmbeddedCompletionsMatchOnDisk`) is independent and stays GREEN by construction. The 0.5 reservation is the single judgment call: dedicated bash test (chosen) vs extending `TestRunCompletionBashScript` (the contract allows both; the PRP picks the dedicated test for cleaner concern separation, but a reviewer could prefer the extension to minimize the test count).
