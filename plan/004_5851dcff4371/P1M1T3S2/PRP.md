# PRP — P1.M1.T3.S2: Flip run-level dispatch + completion tests + help-text assertions

> **Subtask:** P1.M1.T3.S2 — the run-level dispatch + completion + help-text + unconfigured half of P1.M1.T3 (update the test suite to the `--check`/`--init`/`--completions` flag contract). Sibling **S1** (parseArgs-level + exclusivity-level) is **already committed** — it landed during this research session.
> **Scope boundary:** Edits `main_test.go` + ONE assertion in `internal/skillsdir/skillsdir_test.go`. **23 tests** are currently RED; this task makes ALL of them green so `go test ./...` is 100% green. Does NOT touch any `.go` source (already done by S1+S2+T2.S1) or the completion files (those are P1.M2.T1).
> **THE critical finding (proven by a live `go test` run):** the architecture change map (`test_doc_change_map.md`) **MISSED 6 tests**. It lists 20 functions, but 6 MORE tests are RED purely because T2.S1 changed the `ErrNotFound` message from `skilldozer init` → `skilldozer --init`. Those 6 block the "100% green" contract and MUST be flipped too (assertion-string flips, not token flips). See GOTCHA #1 + §1 of the verified-facts file. Without them, `go test ./...` stays RED.

---

## Goal

**Feature Goal**: Make the entire test suite assert the `--check`/`--init`/`--completions` flag contract that the source already implements (S1 `parseArgs` 594be07, S2 `exclusivityError` 1e2fe53, T2.S1 `usageText`+`ErrNotFound` working-tree), so that `go test ./...` is **100% green** and **no bare-subcommand reference remains in any test**.

**Deliverable**: Test-only edits across 2 files (no source, no new files):
1. **16 run()-token flips** in `main_test.go`: every `run([]string{...})` dispatch call passing a bare `"check"`/`"init"`/`"completion"` token passes `"--check"`/`"--init"`/`"--completions"` instead. (8 check + 2 init + 6 completion.)
2. **2 help-text assertion flips**: `TestRunHelpShowsInitRow` asserts `--init` (not `init`); `TestRunHelpShowsCompletionRow` asserts `--completions` (not `completion`).
3. **8 `"skilldozer init"` → `"skilldozer --init"` assertion-string flips** across the unconfigured/ErrNotFound message assertions (`TestRunCheckSkillsDirUnresolvable`, `TestRunBareTagUnconfiguredNeverPrompts`, + the 5 change-map-omitted SkillsDirUnresolvable tests + `internal/skillsdir/skillsdir_test.go::TestErrNotFoundMessageHasFix`).
4. **Leave the 2 no-flip tests alone**: `TestRunTagStillResolvesAlongsideCheck` (passes `"example"`, a tag) and the run() token of `TestRunBareTagUnconfiguredNeverPrompts` (`"someTag"`).

**Success Definition**: `go build ./...` + `go vet ./...` pass; `gofmt -l main_test.go internal/skillsdir/skillsdir_test.go` clean; **`go test ./...` is 100% GREEN (0 failures)**; no source file (`main.go`, `internal/*`, `go.mod`, `go.sum`) is modified; no completion file is modified.

---

## User Persona (if applicable)

Not applicable at runtime — this is test maintenance. The end-user-visible contract it locks in (PRD §6.3 / decision 19): `skilldozer --check` runs validation, `skilldozer --init` runs first-run setup, `skilldozer --completions` emits the completion script — while `skilldozer check` / `init` / `completions` resolve as ordinary skill tags. The entire positional namespace is reserved for skill tags.

---

## Why

- **PRD §6 + decision 19**: the source already drove `check`/`init`/`completion` → `--check`/`--init`/`--completions`. The run-level dispatch, completion, help-text, and unconfigured tests still pass the OLD bare tokens / assert the OLD bare messages and are RED. T3.S2 flips them so `go test` reflects reality.
- **The "100% green" contract (item description OUTPUT)**: `go test ./...` must be fully green. S1 already green'd parseArgs + exclusivity. This task green's the entire remainder — including the 6 tests the change map forgot (GOTCHA #1).
- **"All bare-subcommand references in tests are eliminated" (item description OUTPUT)**: every `"skilldozer init"`/`"skilldozer completion"` assertion string and every bare dispatch token must go.

---

## What

A **mechanical** edit, in three shapes:

**Shape A — run() token flip** (16 sites): swap one bare token inside an existing `run([]string{...})` literal. Assertions (exit codes, stdout/stderr content) are UNCHANGED — only the input arg changes, because the flipped flag drives the same dispatch the old bare subcommand did.

**Shape B — assertion-string flip** (8 sites): the unconfigured/`ErrNotFound` one-line-fix message changed `skilldozer init` → `skilldozer --init` (T2.S1). 8 tests assert that message; flip the asserted substring. NO run() token change.

**Shape C — help-text assertion flip** (2 sites): `usageText` now advertises `--init`/`--completions`; flip the 2 help tests' asserted substrings.

The flip is uniform: `check`→`--check`, `init`→`--init`, `completion`→`--completions` (**PLURAL** — GOTCHA #6).

### Success Criteria

- [ ] 8 check dispatch tests pass `"--check"` tokens (`TestRunCheck*`, incl. the 2 vacuous-green ones — GOTCHA #2)
- [ ] 2 init dispatch tests pass `"--init"` tokens (`TestRunInitStore*`)
- [ ] 6 completion dispatch tests pass `"--completions"` tokens (`TestRunCompletion*`)
- [ ] 2 help-text tests assert `--init` / `--completions` (not bare)
- [ ] `TestRunCheckSkillsDirUnresolvable` + `TestRunBareTagUnconfiguredNeverPrompts` assertion strings flipped to `"skilldozer --init"`
- [ ] **6 change-map-omitted tests flipped** (GOTCHA #1): the 5 SkillsDirUnresolvable tests + `TestErrNotFoundMessageHasFix` — all `"skilldozer init"` → `"skilldozer --init"`
- [ ] `TestRunTagStillResolvesAlongsideCheck` UNCHANGED (passes `"example"`; no flip — GOTCHA #3)
- [ ] `go build ./...` + `go vet ./...` pass; `gofmt -l` clean on both test files
- [ ] **`go test ./...` is 100% GREEN**
- [ ] `main.go`, `internal/skillsdir/skillsdir.go`, `go.mod`, `go.sum`, `completions/*` UNCHANGED

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit is pinned to a function NAME with exact old→new strings in `research/verified_facts.md` §2 (line numbers are POST-S1 and may drift — locate by name). The 6 change-map omissions (the biggest failure risk) are enumerated with exact assertions. The 2 no-flip cases and 2 vacuous-green cases are called out. The source contract (dispatch/usageText/ErrNotFound) is already implemented and verified. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — THE inventory (post-S1 line numbers, exact old→new, the 6 omissions)
- file: plan/004_5851dcff4371/P1M1T3S2/research/verified_facts.md
  why: "§1 = the 6 change-map-omitted tests (the #1 risk — without them go test stays RED). §2 = the 20
        change-map functions with exact token/assertion flips, grouped by shape. §3 = OUT-OF-SCOPE (S1 done;
        leave alone — incl. the 5 NEW namespace tests that intentionally use bare tokens). §4 = the 2
        vacuous-green check tests (flip anyway). §5 = scope summary. §6 = grep locators robust to line drift."
  critical: "§1 + §3 are the two sections that prevent the two failure modes: missing the 6 omissions (RED
             suite) vs. wrongly flipping S1's new bare-tag namespace tests (breaks green)."

# MUST READ — the change map (line numbers are PRE-S1 and ~65 STALE; its FUNCTION list missed 6 tests)
- file: plan/004_5851dcff4371/architecture/test_doc_change_map.md
  why: "The per-function intent. BUT: (a) its line numbers are ~65 too LOW (S1 added tests after it was
        written) — locate by NAME; (b) it MISSES the 6 tests in verified_facts §1; (c) its row for
        TestRunTagStillResolvesAlongsideCheck ('check→--check') is WRONG — that test passes 'example' (no flip)."
  section: "Dispatch tests to flip; Help text tests to update; Unconfigured test. (IGNORE its line numbers;
            cross-check every entry against verified_facts §2.)"

# MUST READ — what run()/parseArgs/usageText/ErrNotFound actually do now (READ-ONLY; already implemented)
- file: main.go
  why: "run() dispatch: c.init→runInit (~522), c.completion→runCompletion (~528), c.check→check report (~634),
        c.path/list/search/all/tags ladder (~547+). On the unconfigured path run() prints err.Error() verbatim
        to stderr (the --init message). usageText const @71-119 advertises --check/--init/--completions.
        defines what the flipped tests must assert."
- file: internal/skillsdir/skillsdir.go
  why: "ErrNotFound var @275 = 'skilldozer is not configured; run `skilldozer --init`' (the message the 8
        assertion-string flips target)."

# READ-ONLY — the sibling PRP (S1, already committed) — defines the boundary this task assumes
- file: plan/004_5851dcff4371/P1M1T3S1/PRP.md
  why: "S1 owned parseArgs-level + exclusivity-level flips + 5 NEW namespace-safety tests. It is COMMITTED.
        This task assumes its output (the 5 BareNowTag tests exist and intentionally pass bare tokens — do NOT
        'fix' them). Confirms the two tasks partition main_test.go with no overlap."

# READ-ONLY — the PRD authority
- file: PRD.md
  why: "§6.1 (--check/--init/--completions flags), §6.3 (no bare-word subcommands; --init is the sole mode
        accepting a positional <dir>), §6.4 (unconfigured ⇒ stderr 'run `skilldozer --init`', exit 1),
        decision 19 (completion→--completions PLURAL). Do NOT edit PRD.md."
  section: "h2.5 (§6), h3.1 (§6.1), h3.4 (§6.4), h2.18 (decision 19)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer
$ git status --short
 M internal/skillsdir/skillsdir.go   # T2.S1 working-tree (ErrNotFound → --init) — NOT this task's concern
 M main.go                           # T2.S1 working-tree (usageText → --flags) — NOT this task's concern
# main_test.go: CLEAN in git (S1 committed). 23 tests RED (this task's entire scope).
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK   # both OK
$ go test ./... 2>&1 | grep -c '^--- FAIL'                        # → 23
```

### Desired Codebase tree with files to be changed

```bash
main_test.go                                  # MODIFY — 22 of the 23 RED tests (token + assertion flips)
internal/skillsdir/skillsdir_test.go          # MODIFY — 1 assertion flip (TestErrNotFoundMessageHasFix)
# main.go / internal/skillsdir/skillsdir.go / go.mod / go.sum / completions/* — UNCHANGED (test-only)
```

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — CRITICAL: the change map MISSED 6 tests. go test ./... currently has 23 RED tests, but
// test_doc_change_map.md only lists ~20 functions. The 6 it missed are RED purely because T2.S1 changed the
// ErrNotFound message 'skilldozer init' → 'skilldozer --init' (skillsdir.go:275). They are a DIFFERENT edit
// shape (assertion-string flip, NOT a run() token flip). If you flip only the change-map functions, go test
// stays RED and the contract OUTPUT '100% green' FAILS. See verified_facts §1 for the exact 6:
//   main_test.go: TestRunPathFailureErrNotFound, TestRunListSkillsDirUnresolvableExit1,
//                 TestRunTagSkillsDirUnresolvable, TestRunAllSkillsDirUnresolvable,
//                 TestRunSearchSkillsDirUnresolvable (all: Contains(...,"skilldozer init") → "--init")
//   internal/skillsdir/skillsdir_test.go: TestErrNotFoundMessageHasFix (Contains "skilldozer init" → "--init")

// GOTCHA #2 — Two check tests are GREEN-but-vacuous/stale and STILL need a token flip. The contract says
// 'All bare-subcommand references in tests are eliminated', and both currently pass a bare "check":
//   TestRunCheckStatusColumnAligned (~1670): run(["check"]) is GREEN only because bare "check"→tag→not found→
//     empty stdout→the alignment loop continues immediately (vacuous). Flip to ["--check"] so the check report
//     actually runs (store has good+bad → OK/ERROR aligned lines → assertions hold → STAYS GREEN).
//   TestRunVersionPrecedenceOverCheck (~1691): run(["check","--version"]) is GREEN because --version wins;
//     flip to ["--check","--version"] (version still wins → STAYS GREEN).
// Flip both (consistency + the 'no bare refs' contract). Do NOT skip them because they're already green.

// GOTCHA #3 — TestRunTagStillResolvesAlongsideCheck passes "example" (a TAG), NOT "check". NO token flip.
//   (The change-map row '1637: check→--check' for it is WRONG — confirmed by reading the run() call + its
//   '// NOT "check" -> tag resolution' comment.) Optional: update its now-stale 'check is reserved' doc
//   comment/name prose (decision 19 removed reserved names). Light touch only.

// GOTCHA #4 — DO NOT touch S1's 5 NEW namespace-safety tests. They intentionally pass bare "check"/"init"/
//   "completions" as TAGS to parseArgs() to prove namespace safety (decision 19): TestParseArgsBareCheckNowTag,
//   BareInitNowTag, BareCompletionsNowTag (~1473-1500). A blanket grep for bare tokens will HIT them — scope
//   your grep to run() calls (they call parseArgs, not run) and SKIP them by name. verified_facts §3.

// GOTCHA #5 — TestRunBareTagUnconfiguredNeverPrompts needs an ASSERTION flip ONLY. Its run() token is
//   "someTag" (a tag) — correct, no token flip. Flip its assertion ["run","skilldozer init"] →
//   ["run","skilldozer --init"] (~2946). Same for TestRunCheckSkillsDirUnresolvable: it needs BOTH a token
//   flip ("check"→"--check") AND an assertion flip ("skilldozer init"→"skilldozer --init") at ~1657.

// GOTCHA #6 — --completions is PLURAL (decision 19). The old bare subcommand was singular "completion"; the
//   new flag is "--completions". Every completion token flip → "--completions" (never "--completion"). The
//   help-text assertion flip is "skilldozer completion" → "skilldozer --completions".

// GOTCHA #7 — TestEmbeddedCompletionsMatchOnDisk (~2990 post-S1) is GREEN and must STAY green. It reads the
//   on-disk completions/* files and compares to the embedded vars (both UNCHANGED — the completion FILES are
//   rewritten in P1.M2.T1, a LATER milestone). Do NOT touch it. If it were RED it would mean an embed bug,
//   not a test-contract issue — but it is GREEN.

// GOTCHA #8 — No deps/import change. The 2 test files already import bytes/strings/os/filepath/testing/etc.
// The gate is gofmt + go vet + go build + FULL go test ./... 100% green. go.mod/go.sum byte-for-byte identical.
```

---

## Implementation Blueprint

### Data models and structure

**None.** No structs, no signatures, no imports change. This is a test-only edit. `config` fields (`c.check`/`c.init`/`c.completion`/`c.initStore`/`c.tags`) and `run()`'s signature are unchanged.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: FLIP the 8 check-dispatch run() tokens (main_test.go) — Shape A
  - For EACH function, change the bare token in the run([]string{...}) call. Locate by NAME (lines are post-S1):
      TestRunCheckCleanStore                (~1543): ["check"]                          → ["--check"]
      TestRunCheckReportsMissingNameExit1   (~1574): ["check"]                          → ["--check"]
      TestRunCheckReportsDuplicateNames     (~1595): ["check"]                          → ["--check"]
      TestRunCheckWarnOnlyExitsZero         (~1618): ["check"]                          → ["--check"]
      TestRunCheckEmptyStoreExit0           (~1636): ["check"]                          → ["--check"]
      TestRunCheckSkillsDirUnresolvable     (~1650): ["check"]                          → ["--check"]  (+ Task 4)
      TestRunCheckStatusColumnAligned       (~1670): ["check"]                          → ["--check"]  (GOTCHA #2; was vacuous-green)
      TestRunVersionPrecedenceOverCheck     (~1691): ["check","--version"]              → ["--check","--version"]  (GOTCHA #2)
  - These tests' ASSERTIONS (exit codes, OK/ERROR/WARN content, summary line) already match the flag behavior;
    no assertion edits here EXCEPT TestRunCheckSkillsDirUnresolvable (Task 4). After the flip they exercise the
    real check report and PASS.

Task 2: FLIP the 2 init-dispatch run() tokens (main_test.go) — Shape A
      TestRunInitStoreWritesConfigCreatesStorePrintsPathExit0 (~2836): ["init","--store",store]   → ["--init","--store",store]
      TestRunInitStoreTildeExpandsHome                       (~2896): ["init","--store","~/sub"] → ["--init","--store","~/sub"]
  - Assertions (store created, config.Store==want, stdout==store+"\n", check report on stderr) unchanged.

Task 3: FLIP the 6 completion-dispatch run() tokens (main_test.go) — Shape A; GOTCHA #6 (PLURAL)
      TestRunCompletionBashScript        (~3021): ["completion","--shell","bash"] → ["--completions","--shell","bash"]
      TestRunCompletionFishScript        (~3037): ["completion","--shell","fish"] → ["--completions","--shell","fish"]
      TestRunCompletionUnsupportedShell  (~3054): ["completion","--shell","tcsh"]  → ["--completions","--shell","tcsh"]
      TestRunCompletionUndetectableShell (~3074): ["completion"]                   → ["--completions"]
      TestRunCompletionEnvShellDetected  (~3091): ["completion"]                   → ["--completions"]
      TestRunCompletionLoginShellDetected (~3107): ["completion"]                  → ["--completions"]
  - Assertions (script markers, exit codes, stderr messages) unchanged.

Task 4: FLIP the 2 help-text ASSERTION strings (main_test.go) — Shape C; run() token is already "--help"
      TestRunHelpShowsInitRow       (~2103): []string{"skilldozer init", "--store <dir>"}       → []string{"skilldozer --init", "--store <dir>"}
      TestRunHelpShowsCompletionRow (~2189): []string{"skilldozer completion", "--shell"}       → []string{"skilldozer --completions", "--shell"}
  - Verified usageText (main.go:81/82/109/111) contains both new substrings + the unchanged second substrings.

Task 5: FLIP the unconfigured/ErrNotFound ASSERTION strings — Shape B (8 sites) — GOTCHA #1 (the 6 omissions) + GOTCHA #5
  - main_test.go (7 sites):
      TestRunCheckSkillsDirUnresolvable     (~1657): Contains(errOut, "skilldozer init")       → "skilldozer --init"
      TestRunBareTagUnconfiguredNeverPrompts(~2946): []string{"run", "skilldozer init"}        → []string{"run", "skilldozer --init"}  (token "someTag" UNCHANGED)
      TestRunPathFailureErrNotFound         (~247):  []string{"run", "skilldozer init"}        → []string{"run", "skilldozer --init"}
      TestRunListSkillsDirUnresolvableExit1 (~490):  Contains(errOut, "skilldozer init")       → "skilldozer --init"
      TestRunTagSkillsDirUnresolvable       (~704):  Contains(errOut, "skilldozer init")       → "skilldozer --init"
      TestRunAllSkillsDirUnresolvable       (~962):  Contains(errOut, "skilldozer init")       → "skilldozer --init"
      TestRunSearchSkillsDirUnresolvable    (~1202): Contains(errOut, "skilldozer init")       → "skilldozer --init"
  - internal/skillsdir/skillsdir_test.go (1 site):
      TestErrNotFoundMessageHasFix          (~531):  []string{"run", "skilldozer init"}        → []string{"run", "skilldozer --init"}
  - All 8 target the SAME message: skillsdir.go:275 '...run `skilldozer --init`'. The "run" substring is unchanged.
  - Use: grep -rn '"skilldozer init"' main_test.go internal/skillsdir/skillsdir_test.go   (must return ZERO hits after).

Task 6: LEAVE the no-flip tests alone (do NOT edit) — GOTCHA #3, GOTCHA #4, GOTCHA #7
  - TestRunTagStillResolvesAlongsideCheck — passes "example"; NO token flip. (Optional: refresh its stale
    'check is reserved' doc-comment prose — decision 19 removed reserved names. Light touch, not required.)
  - TestRunBareTagUnconfiguredNeverPrompts run() token "someTag" — correct, leave it (only its assertion flips, Task 5).
  - The 5 NEW S1 namespace tests (TestParseArgsBareCheckNowTag etc.) — intentionally bare; LEAVE THEM.
  - TestEmbeddedCompletionsMatchOnDisk — GREEN; LEAVE IT.
  - All TestParseArgs* / TestRunExclusivity* (S1, committed) — GREEN; LEAVE THEM.

Task 7: VERIFY (the gate) — GOTCHA #8
  - gofmt -l main_test.go internal/skillsdir/skillsdir_test.go   (must print NOTHING)
  - go vet ./...                                                   (exit 0)
  - go build ./...                                                 (exit 0)
  - git diff --stat                                                (expect ONLY main_test.go + internal/skillsdir/skillsdir_test.go)
  - go test ./...                                                  (MUST be 100% GREEN — 0 FAIL)
  - grep -rn '"skilldozer init"\|"skilldozer completion"' main_test.go internal/skillsdir/skillsdir_test.go   (must return NOTHING)
  - grep -n 'run(\[\]string{"check"\|run(\[\]string{"init"\|run(\[\]string{"completion"' main_test.go          (must return NOTHING)
```

### Implementation Patterns & Key Details

```go
// Shape A (token flip) — a one-token swap inside the existing []string{...} literal. Example:
//   BEFORE:  code := run([]string{"check"}, &out, &errOut)
//   AFTER:   code := run([]string{"--check"}, &out, &errOut)
// (the exit-code / OK-ERROR-WARN / summary assertions already hold for the flag — no assertion edit.)

// Shape B (assertion-string flip) — one substring change in a Contains / []string literal. Example:
//   BEFORE:  if !strings.Contains(errOut.String(), "skilldozer init") { ... }
//   AFTER:   if !strings.Contains(errOut.String(), "skilldozer --init") { ... }
// (the run() token is correct already; only the expected stderr message changed.)

// Shape C (help-text) — like Shape B but the asserted usageText substring:
//   BEFORE:  for _, want := range []string{"skilldozer init", "--store <dir>"} { ... }
//   AFTER:   for _, want := range []string{"skilldozer --init", "--store <dir>"} { ... }
```

Notes easy to get wrong:
- **`--completions` is PLURAL everywhere** (GOTCHA #6). The old subcommand was singular `completion`; the flag is `--completions`. A singular flip re-REDs the test.
- **The 6 change-map omissions are not token flips** (GOTCHA #1). Their `run()` tokens are already correct (`--path`/`--list`/`example`/`--all`/`--search`/ErrNotFound-direct). Only their `"skilldozer init"` assertion string is wrong.
- **Do not "fix" bare tokens in S1's new namespace tests** (GOTCHA #4). They call `parseArgs`, not `run`; their bare `check`/`init`/`completions` are deliberate tags.
- **`TestRunCheckStatusColumnAligned` looks green already — flip it anyway** (GOTCHA #2). It's vacuously green; flipping makes it actually test the check report (and stays green).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Flip the 6 change-map-omitted tests, or leave them as a separate task? → FLIP them.** The contract OUTPUT is `go test ./... 100% green` + "All bare-subcommand references in tests are eliminated." These 6 are RED and contain bare references; leaving them violates both. They're the same trivial edit (assertion substring) and block the exit criterion.
2. **Flip the 2 vacuous-green check tests? → YES.** They currently pass for the wrong reason; the "no bare refs" contract + honest coverage require the flip, and they stay green.
3. **Touch `TestRunTagStillResolvesAlongsideCheck`? → token NO (it passes `example`); prose optionally.** The change map's token-flip row for it is wrong. Its name/comment framing is stale but cosmetic.
4. **Edit the stale comment in skillsdir_test.go:526 (`'...run "skilldozer init"'`)? → optional.** The assertion (Task 5) is what matters; the comment is doc-only. Refresh it for accuracy if you're already there.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports. (GOTCHA #8)

OWNERSHIP (no conflicts):
  - S1 (COMMITTED): parseArgs-level + exclusivity-level flips + 5 new namespace tests. GREEN.
  - S2 (exclusivityError, COMMITTED 1e2fe53): mode-exclusivity flag messages.
  - T2.S1 (working-tree, uncommitted): usageText + ErrNotFound (--init). PRESENT; this task's assertion flips
    target its output but do NOT touch it.
  - T3.S2 (this): run-level dispatch + completion + help-text + unconfigured + the 6 ErrNotFound-omission tests.
  - P1.M2.T1 (later): the completion FILES (completions/*) — NOT this task.

NO SOURCE / NO ROUTES / NO DATABASE / NO COMPLETIONS FILES:
  - T3.S2 edits main_test.go + ONE line in internal/skillsdir/skillsdir_test.go. Nothing else.
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (hard gates)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main_test.go internal/skillsdir/skillsdir_test.go   # must print NOTHING (gofmt -w if listed)
go vet ./...                                                  # expect exit 0
go build ./...                                                # expect exit 0
git diff --stat                                               # expect ONLY main_test.go + internal/skillsdir/skillsdir_test.go
git diff --quiet go.mod go.sum && echo "deps unchanged"       # expect: deps unchanged
```

### Level 2: Full test suite — THE gate (must be 100% GREEN)

```bash
cd /home/dustin/projects/skilldozer

go test ./...                              # MUST be 100% GREEN — every package ok, ZERO --- FAIL
go test ./... 2>&1 | grep -c '^--- FAIL'   # MUST print 0

# Targeted confirmation of the families this task flipped (all PASS):
go test -v -run 'TestRunCheck|TestRunInit|TestRunCompletion|TestRunHelpShows|TestRunBareTagUnconfigured|TestRunPathFailure|TestRunListSkillsDirUnresolvable|TestRunTagSkillsDirUnresolvable|TestRunAllSkillsDirUnresolvable|TestRunSearchSkillsDirUnresolvable' .
go test -v -run 'TestErrNotFoundMessageHasFix' ./internal/skillsdir/
```

### Level 3: Scope invariants (prove T3.S2 stayed in its lane)

```bash
cd /home/dustin/projects/skilldozer

git diff --name-only          # Expected: main_test.go + internal/skillsdir/skillsdir_test.go (ONLY)
# No source touched:
git diff --quiet main.go internal/skillsdir/skillsdir.go go.mod go.sum && echo "source/deps unchanged"
# (Note: main.go + internal/skillsdir/skillsdir.go show as M from T2.S1's PRE-EXISTING working-tree changes,
#  NOT from this task. Confirm with: git diff main.go | grep -c 'skilldozer init'  → should show T2.S1's
#  --init lines, NOT new edits. This task adds NO lines to source files.)

# No bare dispatch tokens remain in any run() call:
grep -n 'run(\[\]string{"check"\|run(\[\]string{"check",\|run(\[\]string{"init"\|run(\[\]string{"init",\|run(\[\]string{"completion"' main_test.go
# Expected: NOTHING (empty). (The 5 NEW parseArgs namespace tests call parseArgs(), not run(), so they're excluded.)

# No bare "skilldozer init"/"skilldozer completion" assertion strings remain:
grep -rn '"skilldozer init"\|"skilldozer completion"' main_test.go internal/skillsdir/skillsdir_test.go
# Expected: NOTHING (empty).
```

### Level 4: Namespace-safety + contract smoke (the point of decision 19)

```bash
cd /home/dustin/projects/skilldozer

# The full suite green already proves it at unit level. Optional binary-level smoke:
go build -o /tmp/sdz . && /tmp/sdz --check >/dev/null 2>&1; echo "--check exit=$?"   # 0 (example skill clean)
/tmp/sdz check >/dev/null 2>&1; echo "bare check exit=$?"                            # 1 (unknown tag — namespace safety)
/tmp/sdz --help | grep -q -- '--init' && /tmp/sdz --help | grep -q -- '--completions' && echo "help advertises --init/--completions OK"
rm -f /tmp/sdz
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean on both test files; `go vet ./...` exit 0; `go build ./...` exit 0
- [ ] Level 2 PASS — **`go test ./...` is 100% GREEN (0 FAIL)**; the flipped-family `-run` selectors all PASS
- [ ] Level 3 PASS — `git diff --name-only` = ONLY main_test.go + internal/skillsdir/skillsdir_test.go; the two bare-token/bare-string greps return NOTHING; no source/deps changed by this task

### Feature Validation
- [ ] 8 check dispatch tests pass `"--check"` tokens (incl. the 2 vacuous-green ones)
- [ ] 2 init dispatch tests pass `"--init"` tokens
- [ ] 6 completion dispatch tests pass `"--completions"` (PLURAL) tokens
- [ ] 2 help-text tests assert `--init` / `--completions`
- [ ] `TestRunCheckSkillsDirUnresolvable` + `TestRunBareTagUnconfiguredNeverPrompts` assertion strings flipped to `"skilldozer --init"`
- [ ] **6 change-map-omitted tests flipped** (5 SkillsDirUnresolvable + TestErrNotFoundMessageHasFix) — GOTCHA #1
- [ ] `TestRunTagStillResolvesAlongsideCheck` UNCHANGED (no token flip)
- [ ] `TestEmbeddedCompletionsMatchOnDisk` still GREEN (untouched)

### Code Quality / Convention Validation
- [ ] `--completions` is PLURAL everywhere (never `--completion`)
- [ ] Minimal diff (token swaps + assertion-string swaps); no churn in untouched tests
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical

### Scope Discipline (the S1 + source + completions boundaries)
- [ ] Did NOT touch any `TestParseArgs*` or `TestRunExclusivity*` (S1, committed, GREEN)
- [ ] Did NOT touch S1's 5 NEW namespace tests (intentionally bare tags)
- [ ] Did NOT modify `main.go`, `internal/skillsdir/skillsdir.go`, `go.mod`, `go.sum`, or any source file
- [ ] Did NOT modify `completions/*` (those are P1.M2.T1)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't flip only the ~20 change-map functions.** The change map MISSED 6 tests (GOTCHA #1). Without flipping them, `go test ./...` stays RED and the contract fails. Use `grep -rn '"skilldozer init"'` to find ALL assertion-string sites.
- ❌ **Don't flip `TestRunTagStillResolvesAlongsideCheck`'s token.** It passes `"example"` (a tag), not `"check"` (GOTCHA #3). The change-map row for it is wrong.
- ❌ **Don't "fix" the bare tokens in S1's new namespace tests** (`TestParseArgsBareCheckNowTag` etc.). They intentionally pass bare `check`/`init`/`completions` as tags to prove decision 19 (GOTCHA #4). They call `parseArgs`, not `run`.
- ❌ **Don't skip the 2 green-but-vacuous check tests** (`TestRunCheckStatusColumnAligned`, `TestRunVersionPrecedenceOverCheck`). They're green for the wrong reason; flip their tokens so they honestly cover the check path (GOTCHA #2). They stay green.
- ❌ **Don't write `--completion` (singular).** The flag is `--completions` (PLURAL, decision 19). A singular flip re-REDs the test (GOTCHA #6).
- ❌ **Don't confuse the two edit shapes.** Token flips change the `run([]string{...})` ARG. Assertion-string flips change the EXPECTED message substring. The 6 omitted tests + the 2 unconfigured/help tests need assertion flips; some tests (TestRunCheckSkillsDirUnresolvable) need BOTH.
- ❌ **Don't edit any `.go` source or `completions/*`.** Source is done (S1+S2+T2.S1); completion files are P1.M2.T1. T3.S2 = the 2 test files ONLY.
- ❌ **Don't touch `TestEmbeddedCompletionsMatchOnDisk`.** It's GREEN and verifies embed↔on-disk byte-identity (unchanged until P1.M2.T1) (GOTCHA #7).
- ❌ **Don't add deps or imports.** go.mod/go.sum byte-for-byte identical (GOTCHA #8).

---

## Confidence Score

**9.5/10** — The entire RED set (23 tests) is enumerated with exact old→new strings (verified_facts §1/§2), confirmed by a live `go test` run AND an independent `scout` recon. The single biggest risk — the 6 change-map omissions (GOTCHA #1) — is fully documented with the exact assertion lines and a grep that proves completion. The two no-flip traps (GOTCHA #3 `example` token; GOTCHA #4 S1's bare-tag tests) and the two vacuous-green tests (GOTCHA #2) are explicitly called out. The source contract (run dispatch / usageText / ErrNotFound) is already implemented and verified present. The 0.5 reservation is line-drift: the listed line numbers are post-S1 but may shift again if anything else lands before implementation — mitigated by "locate by function name + grep" throughout.
