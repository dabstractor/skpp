# PRP — P1.M2.T1.S1: Add `c.path` to the tags-conflict predicate in `exclusivityError` + tests (Issue 3)

> **Subtask:** A one-line bugfix. `exclusivityError` (main.go) treats `--path` as a first-class mutually-exclusive "inspection mode" in BOTH the mode+mode count set (main.go:715) and the check+mode set (main.go:730) — but OMITS it from the **tags** predicate (main.go:724). So `skilldozer <tag> --list/search/all` correctly exits 2, while `skilldozer <tag> --path` (or `--path <tag>`) silently runs `--path` and drops the tag — even an UNKNOWN tag (`NONEXISTENTTAG --path` → exit 0). This closes the asymmetry by adding `c.path` to the tags predicate and its message, and ships the missing tests. This is the same class of fix that N1 already applied to the `check` case (`TestRunExclusivityCheckAndPath`).
>
> **Scope:** Two existing files only — `main.go` (one predicate + one message string + one doc-comment bullet) and `main_test.go` (3 new tests). No new files. No parseArgs/config/run() change. Zero new deps.
>
> **STATUS (verified at PRP-write time):** read `exclusivityError` + its doc comment + every sibling exclusivity test at exact line ranges. The buggy line is main.go:724-725 (the contract cites "702/708/686" — PRD-write-time numbers; anchor by the predicate TEXT `hasTags && (c.list || c.searchMode || c.all)`, which is unique in the file). The parallel sibling P1.M1.T2.S2 edits run() + run()-level `--store` tests — disjoint regions, no collision (see §6 of verified_facts). Zero existing tests assert the exact tags-message string (grep-confirmed), so the message edit breaks nothing.

---

## Goal

**Feature Goal**: Make `exclusivityError` reject `tags + --path` exactly as it already rejects `tags + --list/search/all`, so PRD §6.3 ("Mixing `<tag>` with `--list`/`--search`/`--all` is an error (exit 2)") is applied consistently for `--path`. After the fix, `skilldozer <tag> --path` and `skilldozer --path <tag>` exit 2 with an empty stdout and a stderr message naming the conflict — instead of silently running `--path` and dropping the tag (which today masks user typos, e.g. `skilldozer myskill --path` printing the STORE path with no warning).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — (a) add `c.path` to the tags predicate (main.go:724) and prepend `--path/` to its message (main.go:725); (b) update the `exclusivityError` doc-comment bullet (main.go:694) to note `--path` is now part of the tags conflict.
2. `main_test.go` — 3 new tests: `TestRunExclusivityTagsAndPath` (run-level, mirrors `TestRunExclusivityTagsAndList`), `TestRunExclusivityPathAndTag` (reversed order), and `TestExclusivityErrorTagsAndPath` (direct `exclusivityError` unit test locking the exact buggy predicate line).

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `run(["foo","--path"])` and `run(["--path","foo"])` → exit 2 + empty stdout + stderr containing `cannot be combined`; the `NONEXISTENTTAG --path` repro now exits 2 (was 0).

---

## User Persona (if applicable)

**Target User**: A user who runs `skilldozer <tag> --path` (or `--path <tag>`) expecting either an error or the tag's path — and a script author relying on `$(skilldozer <tag> --path)` to fail loudly on a typo'd tag.

**Use Case**: `skilldozer myskill --path` — the user intends to inspect where `myskill` lives; today they silently get the STORE path with no warning that `myskill` was dropped (and even `NONEXISTENTTAG --path` exits 0).

**User Journey**: User runs `skilldozer myskill --path` → (today) store path printed, exit 0, tag silently dropped → (after fix) clear `tags cannot be combined with --path/...` error, exit 2, empty stdout — so a typo is caught instead of masked.

**Pain Points Addressed**: silent tag drop that masks typos; internal inconsistency (`--path` is exclusive vs other modes but not vs tags).

---

## Why

- **Closes the asymmetry documented in bug_fixes_validation.md §ISSUE 3.** `--path` is already a member of the mode+mode count set (main.go:715) and the check+mode set (main.go:730); omitting it only from the tags predicate (main.go:724) is an internal inconsistency, not a deliberate exemption.
- **Restores PRD §6.3 consistency.** "Mixing `<tag>` with `--list`/`--search`/`--all` is an error (exit 2)" — `--path` is the same class of inspection mode; the fix makes the rule uniform.
- **Matches the precedent N1 already set for `check`.** `check + --path` was fixed the same way (doc comment main.go:698-701 + `TestRunExclusivityCheckAndPath`); this applies the identical reasoning to tags.
- **Prevents a real footgun.** `NONEXISTENTTAG --path` exiting 0 means a typo silently produces the store path — exactly the kind of quiet failure PRD §6.4 is written to prevent for `$(...)` use.

---

## What

A one-line predicate change, a one-line message change, a one-line doc-comment edit, and 3 tests. No exit-code change, no ordering change, no new family.

### Success Criteria

- [ ] `exclusivityError`'s tags predicate (main.go:724) is `hasTags && (c.path || c.list || c.searchMode || c.all)`.
- [ ] Its message (main.go:725) is `"skilldozer: tags cannot be combined with --path/--list/--search/--all"`.
- [ ] The doc-comment bullet (main.go:694) reflects `--path` inclusion and cites Issue 3.
- [ ] `run(["foo","--path"])` → exit 2, empty stdout, stderr Contains(`cannot be combined`).
- [ ] `run(["--path","foo"])` → exit 2, empty stdout, stderr Contains(`cannot be combined`).
- [ ] `exclusivityError(config{tags:[]string{"foo"}, path:true})` → bad=true, msg Contains(`tags cannot be combined`) AND Contains(`--path`).
- [ ] `go test ./...` green; existing tests unaffected (the message edit breaks nothing — verified_facts §3).
- [ ] `go.mod`/`go.sum` unchanged; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** The single edit site is pinned by the unique predicate text `hasTags && (c.list || c.searchMode || c.all)` (the only such line in the file). The asymmetry is proven by reading the two sibling predicates that DO include `c.path` (main.go:715 count set, main.go:730 check+mode set). The fix (predicate + message) is fixed verbatim by the contract LOGIC §3 + bug_fixes_validation.md §ISSUE 3. The message edit's zero-breakage is grep-confirmed (no test asserts the exact tags-message string). The test templates (`TestRunExclusivityTagsAndList`, `TestRunExclusivityCheckAndPath`) are read at exact line ranges. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (exact line, asymmetry proof, zero-breakage, test plan)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/P1M2T1S1/research/verified_facts.md
  why: "§1 pins the buggy line by TEXT (main.go:724-725; the contract's 702/708/686 are
        stale PRD-write-time numbers) and gives the exact one-line fix. §2 the repro +
        the N1 precedent (check+path already fixed). §3 grep PROOF that the message edit
        breaks zero tests (the only assertion is Contains('cannot be combined')). §4 the
        test plan + WHY the bug-doc's 'add to TestExclusivityErrorListingModes' suggestion
        is table-inappropriate (use a dedicated direct-unit test). §5 the doc-comment edit.
        §6 disjointness from P1.M1.T2.S2. §7 scope discipline."
  critical: "§3 (zero breakage) and §4(c) (do NOT add a tags case to the listing-modes
             table — its assertion is Contains('mutually exclusive'), which a tags case
             would fail) are the two things most likely to be missed."

# MUST READ — the authoritative bug writeup + repro
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/bug_fixes_validation.md
  why: "§ISSUE 3 is the authoritative repro (NONEXISTENTTAG --path → exit 0; --list → 2),
        pins the site (exclusivityError predicate), and gives the exact fix (add c.path +
        update message). NOTE: its 'add a case to TestExclusivityErrorListingModes'
        suggestion is shape-correct but table-inappropriate — verified_facts §4(c)
        corrects this to a dedicated direct-unit test."
  section: "ISSUE 3 (Minor)."

# MUST READ — the file under edit (read exclusivityError + its doc comment in full)
- file: main.go
  why: "THE edit target. exclusivityError func @708; doc comment @688-707. Buggy predicate
        @724: `if hasTags && (c.list || c.searchMode || c.all) {`. Message @725. Doc bullet
        @694: `//   - tags + a listing mode (--list/--search/--all) — PRD §6.3 explicit`.
        The two sibling predicates that ALREADY include c.path: the count set @715
        `[]bool{c.path, c.list, c.searchMode, c.all}` and the check+mode set @730
        `(c.path || c.list || c.searchMode || c.all)`. The N1 precedent paragraph @698-701."
  pattern: "Exclusivity family = `if <condition> { return true, \"skilldozer: ...\" }`.
            Messages use the `skilldozer: <subject> cannot be combined with <modes>` convention
            (@725 tags, @728 check+tags, @731 check+mode, @743 init+mode)."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestRunExclusivityTagsAndList
        @1717-1729 is the EXACT run-level template (code==2, empty stdout, Contains
        'cannot be combined'). TestRunExclusivityTagsAndSearch @1732 + TagsAndAll @1744
        are sibling variants. TestRunExclusivityCheckAndPath @1786 is the N1 precedent
        (the same asymmetry, already fixed for check — its comment is the model wording).
        TestExclusivityErrorListingModes @2128 is the direct exclusivityError unit test,
        BUT its assertion is Contains('mutually exclusive') so a tags case does NOT fit
        there (verified_facts §4c) — add a dedicated TestExclusivityErrorTagsAndPath."
  gotcha: "Exclusivity runs at run() step 4, BEFORE skillsdir.Find() — so NONE of these
           tests need a store fixture, SKILLDOZER_SKILLS_DIR, t.Chdir, or unsetSkillsEnv.
           Pure argv → exit-2 checks."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint regions, no collision)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/P1M1T2S2/PRP.md
  why: "Confirms P1.M1.T2.S2 edits run() (~main.go:438, the storeMissingValue guard) + the
        run() precedence comment + 4 run()-level --store tests. It does NOT touch
        exclusivityError's tags predicate, its message, or its doc comment, and does NOT
        touch the tags+mode exclusivity tests. Disjoint regions; land in either order."

# READ-ONLY — PRD (the authority for the mutual-exclusivity rule)
- file: PRD.md
  why: "READ-ONLY. §6.3 ('Mixing <tag> with --list/--search/--all is an error (exit 2)':
        these are mutually exclusive modes') — the rule this fix makes uniform for --path.
        §6.4 (nothing on stdout on failure, for $(...) safety). The bugfix PRD §3.2 Issue 3
        is the repro."
  section: "§6.3, §6.4 (and the bugfix PRD h3.2 Issue 3)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/tasks.json
  why: "P1.M2.T1.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP
        transcribes it; tasks.json wins on any conflict."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep)
$ grep -n 'hasTags && (c.list' main.go   # the buggy predicate — unique in the file
# main.go:724 (anchor by TEXT, not number — the contract's 702/708/686 are stale)
```

### Desired Codebase tree with files to be changed

```bash
main.go        # MODIFY exclusivityError: tags predicate + message (1 line) + doc-comment bullet (1 line)
main_test.go   # ADD 3 tests: TestRunExclusivityTagsAndPath, TestRunExclusivityPathAndTag, TestExclusivityErrorTagsAndPath
# go.mod / go.sum — UNCHANGED (no new deps; predicate + string + Contains use only existing constructs)
```

| File | Change | Owner |
|---|---|---|
| `main.go` | `exclusivityError`: +`c.path` to tags predicate + prepend `--path/` to message; doc-comment bullet updated | Issue 3 contract + bug_fixes_validation.md §ISSUE 3 |
| `main_test.go` | +3 tests (2 run-level both orderings + 1 direct exclusivityError unit test) | QA Issue 3 |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — ANCHOR BY TEXT, not line number. The contract cites the predicate at "main.go:702"
// and the func at "686/708"; those are PRD-write-time numbers. The CURRENT line is 724 (the bugfix
// round's M1 work shifted things). The predicate text `hasTags && (c.list || c.searchMode || c.all)`
// is UNIQUE in the file — match on it, not a line number.

// GOTCHA #2 — Edit BOTH the predicate AND the message string. The bug is that c.path is missing
// from the predicate, but the message also lists only "--list/--search/--all". The fix adds c.path
// to the predicate AND prepends "--path/" to the message so the message matches the predicate's
// mode set (consistency: every exclusivity message names exactly the modes its predicate checks).

// GOTCHA #3 — Zero breakage is grep-proven, not assumed. The message edit changes
// "--list/--search/--all" → "--path/--list/--search/--all". The ONLY test asserting a tags-
// exclusivity message is TestRunExclusivityTagsAndList, which uses Contains("cannot be combined")
// (NOT the mode list). Verified by grep: no test asserts the exact tags-message string. Still
// re-run `go test ./...` to confirm. (verified_facts §3.)

// GOTCHA #4 — Do NOT add a tags case to TestExclusivityErrorListingModes (main_test.go:2128). That
// table asserts every `bad` case contains "mutually exclusive" (the listing-mode family's wording).
// A tags+path case returns "tags cannot be combined with …" which does NOT contain "mutually
// exclusive" → it would FAIL the table assertion. bug_fixes_validation.md §ISSUE 3 suggested it,
// but the suggestion is table-inappropriate; use a dedicated TestExclusivityErrorTagsAndPath
// (direct exclusivityError call) instead. (verified_facts §4c.)

// GOTCHA #5 — No store fixture / env for these tests. Exclusivity runs at run() step 4, BEFORE
// skillsdir.Find(). So run(["foo","--path"]) exits 2 without touching the filesystem — NO
// SKILLDOZER_SKILLS_DIR, NO t.Chdir, NO unsetSkillsEnv, NO sampleStore. (Contrast: the
// tag-RESOLUTION tests DO need a store, but these are exclusivity tests, not resolution tests.)

// GOTCHA #6 — Both orderings must be tested. The contract OUTPUT requires "`<tag> --path` (or
// `--path <tag>`)". parseArgs captures flags and tags in any order, so both reach exclusivityError
// identically — but test BOTH to lock the contract literally (a future parser refactor could
// reorder/drop one).

// GOTCHA #7 — Do NOT touch the mode+mode count set (main.go:715) or the check+mode set (730).
// Those ALREADY include c.path correctly. Only the tags predicate (724) is buggy. Editing the
// other two is out of scope and would be a no-op (they're already correct).

// GOTCHA #8 — No deps/imports change. The predicate edit + message string + Contains-based tests
// use only already-imported constructs (strings is imported in both main.go and main_test.go).
// go.mod/go.sum must be byte-for-byte identical. Verify with `git diff --quiet go.mod go.sum`.
```

---

## Implementation Blueprint

### Data models and structure

**No data-model changes.** This subtask edits one boolean predicate and one string literal in `exclusivityError`, plus a doc comment. No new fields, types, methods, or signatures.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add c.path to the tags predicate + update the message
  - FILE: main.go (exclusivityError, the tags family ~line 724-725; anchor by TEXT)
  - FIND the unique line (GOTCHA #1):
        hasTags := len(c.tags) > 0
        if hasTags && (c.list || c.searchMode || c.all) {
            return true, "skilldozer: tags cannot be combined with --list/--search/--all"
        }
  - REPLACE with (GOTCHA #2 — predicate AND message):
        hasTags := len(c.tags) > 0
        if hasTags && (c.path || c.list || c.searchMode || c.all) {
            return true, "skilldozer: tags cannot be combined with --path/--list/--search/--all"
        }
  - This is the ENTIRE bugfix. No exit-code change (still 2). No ordering change (still the
    2nd family). No new family (the tags family's predicate is expanded; the doc comment's
    "four families" count is unchanged).
  - GOTCHA #7: do NOT touch the count set @715 or the check+mode set @730 — already correct.

Task 2: EDIT main.go — update the exclusivityError doc-comment bullet (Mode A)
  - FILE: main.go (exclusivityError doc comment, the bullet at ~line 694)
  - FIND:
        //   - tags + a listing mode (--list/--search/--all) — PRD §6.3 explicit
  - REPLACE with:
        //   - tags + an inspection mode (--path/--list/--search/--all) — PRD §6.3 (Issue 3:
        //     --path was omitted, silently dropping a stray tag; now uniform with the
        //     check+mode and mode+mode sets)
  - Optionally append one line to the N1 paragraph (~698-701) noting the SAME asymmetry is
    now closed for tags, e.g. after "N1 closed that asymmetry.":
        //     Issue 3 (P1.M2.T1.S1) closed the identical asymmetry for tags+--path.
    (Cosmetic/honesty; the bullet edit is the required Mode-A change per the contract DOCS.)

Task 3: EDIT main_test.go — add the 2 run-level tests (mirror TestRunExclusivityTagsAndList)
  - FILE: main_test.go (insert right after TestRunExclusivityTagsAndAll, ~line 1754, grouped
    with the other tags+mode exclusivity tests; package main)
  - ADD (GOTCHA #5 — no store/env; GOTCHA #6 — both orderings):
    // Issue 3 (P1.M2.T1.S1): tags + --path is now rejected like tags + --list/search/all.
    // Previously --path was omitted from the tags predicate, so `foo --path` silently ran
    // --path and dropped the tag (even NONEXISTENTTAG --path → exit 0). exclusivityError
    // runs before skillsdir.Find(), so no store fixture is needed.
    func TestRunExclusivityTagsAndPath(t *testing.T) {
        var out, errOut bytes.Buffer
        code := run([]string{"foo", "--path"}, &out, &errOut)
        if code != 2 {
            t.Fatalf("run(foo --path): code=%d; want 2 (Issue 3: tags + --path)", code)
        }
        if out.Len() != 0 {
            t.Errorf("stdout=%q; want empty", out.String())
        }
        if !strings.Contains(errOut.String(), "cannot be combined") {
            t.Errorf("stderr=%q; want an exclusivity message", errOut.String())
        }
    }
    // Reversed order: `--path foo` must also exit 2 (parseArgs captures flags/tags in any order).
    func TestRunExclusivityPathAndTag(t *testing.T) {
        var out, errOut bytes.Buffer
        code := run([]string{"--path", "foo"}, &out, &errOut)
        if code != 2 {
            t.Fatalf("run(--path foo): code=%d; want 2 (Issue 3, reversed order)", code)
        }
        if out.Len() != 0 {
            t.Errorf("stdout=%q; want empty", out.String())
        }
        if !strings.Contains(errOut.String(), "cannot be combined") {
            t.Errorf("stderr=%q; want an exclusivity message", errOut.String())
        }
    }

Task 4: EDIT main_test.go — add the direct exclusivityError unit test (locks the exact buggy line)
  - FILE: main_test.go (insert near TestExclusivityErrorListingModes, ~line 2128, as a SEPARATE
    function — NOT a case in that table; GOTCHA #4)
  - ADD:
    // Issue 3 (P1.M2.T1.S1): the direct predicate test. Calls exclusivityError itself so the
    // fix is locked independent of parseArgs/run. (Not a case in TestExclusivityErrorListingModes:
    // that table asserts Contains("mutually exclusive"); a tags case returns "tags cannot be
    // combined", a different family.)
    func TestExclusivityErrorTagsAndPath(t *testing.T) {
        bad, msg := exclusivityError(config{tags: []string{"foo"}, path: true})
        if !bad {
            t.Fatalf("exclusivityError(tags+path)=bad=false; want true (Issue 3: c.path was omitted)")
        }
        if !strings.Contains(msg, "tags cannot be combined") {
            t.Errorf("msg=%q; want 'tags cannot be combined'", msg)
        }
        if !strings.Contains(msg, "--path") {
            t.Errorf("msg=%q; want it to mention --path", msg)
        }
    }

Task 5: VERIFY in isolation + whole module + invariants
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0
  - go test -run 'TagsAndPath|PathAndTag|ExclusivityErrorTagsAndPath' -v ./...  # the 3 new tests pass
  - go test ./...                     # whole module green; zero regressions (GOTCHA #3)
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #8
  - INVARIANT: grep -c 'c.path || c.list || c.searchMode || c.all' main.go  # expect 2 (count set uses
    `c.path, c.list...` not `||`; the `||` form appears in the tags predicate (new) + check+mode set
    (730) — confirm the tags predicate now matches the check+mode predicate's mode set)
  - END-TO-END (the §ISSUE 3 repro, now fixed): build, run NONEXISTENTTAG --path → exit 2 (was 0).
```

### Implementation Patterns & Key Details

```go
// The fix (Task 1) — predicate AND message, anchored by the unique predicate text:
hasTags := len(c.tags) > 0
if hasTags && (c.path || c.list || c.searchMode || c.all) {   // was: (c.list || c.searchMode || c.all)
	return true, "skilldozer: tags cannot be combined with --path/--list/--search/--all" // was: --list/--search/--all
}

// The doc-comment bullet (Task 2):
//   - tags + an inspection mode (--path/--list/--search/--all) — PRD §6.3 (Issue 3: --path was omitted …)
```

Notes easy to get wrong:
- Edit **both** the predicate and the message — the message must list exactly the modes the predicate checks (every other exclusivity message does).
- Anchor the edit by the predicate **text** (`hasTags && (c.list || c.searchMode || c.all)`), not a line number — the contract's line cites are stale.
- Do **not** add a tags case to `TestExclusivityErrorListingModes` (its assertion is `Contains("mutually exclusive")`); use a dedicated direct-unit test.
- None of the new tests need a store/env fixture — exclusivity runs before `skillsdir.Find()`.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Predicate expansion, not a new family.** The fix adds `c.path` to the existing tags family's predicate; it does not create a 5th family. The doc comment's "four families" framing stays accurate. This is the minimal change that closes the asymmetry.
2. **Message lists `--path` first.** `--path/--list/--search/--all` matches the order of the existing mode+mode message (main.go:721) and the check+mode message (main.go:731), keeping the mode-list ordering consistent across all three families.
3. **Direct unit test (`TestExclusivityErrorTagsAndPath`) in addition to run-level tests.** The run-level tests prove the end-to-end behavior; the direct test isolates the exact one line that changed, so a future refactor that re-introduces the asymmetry is caught even if parseArgs/run routing shifts. The bug doc suggested adding to the listing-modes table, but that table's assertion is wording-specific (`Contains("mutually exclusive")`); a dedicated test is correct.
4. **Both orderings tested.** `<tag> --path` and `--path <tag>` — parseArgs handles both identically today, but locking both guards against a future parser change.
5. **No README change here.** The contract DOCS assigns the README error-contract sweep to the final Mode B task (P1.M3.T1). This subtask's doc edit is the in-code `exclusivityError` comment only (Mode A).

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports (predicate edit + string + Contains). (GOTCHA #8)

DISPATCH (unchanged): exclusivityError is still called at run() step 4, after unknownFlag and
  after P1.M1.T2.S2's storeMissingValue guard, before any mode dispatch / skillsdir.Find().
  tags+path cases skip the storeMissingValue guard (no --store) and hit this expanded predicate.

CONSUMERS:
  - The §ISSUE 3 repro: NONEXISTENTTAG --path now exits 2 (was 0).
  - README error-contract section: swept by the final Mode B task (P1.M3.T1) — no doc file rides
    here beyond the exclusivityError doc comment (Mode A, Task 2).

PARALLEL SIBLING (no conflict):
  - P1.M1.T2.S2 edits run() (~main.go:438) + run() precedence comment + 4 run()-level --store
    tests. This subtask edits exclusivityError (~main.go:724) + its doc comment (~694) + tags+path
    tests. DISJOINT regions in both files; no text-level overlap; land in either order.

NO ROUTES / NO DATABASE / NO CONFIG-FORMAT CHANGE / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after the main.go edit)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0
go build ./...                  # expect exit 0
# Expected: zero output / exit 0.
```

### Level 2: Unit Tests (the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test -run 'TagsAndPath|PathAndTag|ExclusivityErrorTagsAndPath' -v ./...
# Expected: ALL 3 pass. The load-bearing assertions:
#   TestRunExclusivityTagsAndPath         -> run(["foo","--path"]) code 2, empty stdout, msg Contains "cannot be combined".
#   TestRunExclusivityPathAndTag          -> run(["--path","foo"]) code 2 (reversed order).
#   TestExclusivityErrorTagsAndPath       -> exclusivityError({tags:["foo"],path:true}) bad=true, msg has "tags cannot be combined" + "--path".

# Regression — the existing tags+mode + check+path tests stay green (message edit is invisible to them):
go test -run 'TestRunExclusivityTagsAnd|TestRunExclusivityCheckAndPath|TestExclusivityErrorListingModes|TestRunExclusivityListingModePairs' -v ./...
# Expected: PASS (GOTCHA #3 — the only tags assertion is Contains("cannot be combined")).
```

### Level 3: Whole-module regression + the §ISSUE 3 repro (now fixed)

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions (GOTCHA #3)

# GOTCHA #8 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Scope invariant: the tags predicate now matches the check+mode predicate's mode set
grep -n 'hasTags && (c.path || c.list' main.go   # expect exactly 1 hit (the fixed predicate)

# END-TO-END — the §ISSUE 3 repro, now FIXED (NONEXISTENTTAG --path → exit 2, was 0):
go build -o /tmp/sd .
env -u SKILLDOZER_SKILLS_DIR /tmp/sd NONEXISTENTTAG --path >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" = 2 ] && [ ! -s /tmp/o ] && grep -q 'cannot be combined' /tmp/e \
  && echo "Issue-3 repro FIXED: NONEXISTENTTAG --path → exit 2 + empty stdout + msg" \
  || echo "FAIL: rc=$rc out=$(cat /tmp/o) err=$(cat /tmp/e)"
# Control (unchanged): NONEXISTENTTAG --list still exits 2 the same way.
env -u SKILLDOZER_SKILLS_DIR /tmp/sd NONEXISTENTTAG --list >/dev/null 2>&1; echo "list-control exit=$? (want 2)"
rm -f /tmp/sd /tmp/o /tmp/e
# Expected: "Issue-3 repro FIXED …"; list-control exit 2.
```

### Level 4: Behavioral spot-checks (lock the fix is scoped, no over-reach)

```bash
cd /home/dustin/projects/skilldozer

# 4a. --path ALONE (no tag) still works (exit 0/1 from --path dispatch, NOT exit 2): the fix
#     only triggers when hasTags is true. Bare --path is not a tags conflict.
go build -o /tmp/sd .
env -u SKILLDOZER_SKILLS_DIR /tmp/sd --path >/dev/null 2>&1; echo "bare --path exit=$? (want 0 or 1, NOT 2)"

# 4b. A VALID tag + --path now exits 2 (the fix), not the old silent store-path print:
store=$(mktemp -d) && mkdir -p "$store/example" && printf -- '---\nname: example\ndescription: d\n---\nx\n' > "$store/example/SKILL.md"
SKILLDOZER_SKILLS_DIR="$store" /tmp/sd example --path >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" = 2 ] && [ ! -s /tmp/o ] && grep -q 'cannot be combined' /tmp/e \
  && echo "valid-tag + --path → exit 2 (fixed; was silently printing the store path)" \
  || echo "FAIL: rc=$rc out=$(cat /tmp/o) err=$(cat /tmp/e)"
rm -rf "$store" /tmp/sd /tmp/o /tmp/e
# Expected: "valid-tag + --path → exit 2".
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l main.go main_test.go` empty; `go vet ./...` exit 0; `go build` exit 0
- [ ] Level 2 PASS — the 3 new tests pass (2 run-level both orderings + 1 direct exclusivityError unit test)
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (zero regressions); `git diff go.mod go.sum` → "deps unchanged"; the §ISSUE 3 repro now exits 2 + empty stdout + msg
- [ ] Level 4 PASS — bare `--path` is NOT a tags conflict (exit 0/1, not 2); a valid `<tag> --path` now exits 2

### Feature Validation
- [ ] `exclusivityError` tags predicate includes `c.path`; message lists `--path/--list/--search/--all`
- [ ] `run(["foo","--path"])` and `run(["--path","foo"])` → exit 2, empty stdout, stderr Contains `cannot be combined`
- [ ] `exclusivityError(config{tags:["foo"],path:true})` → bad=true, msg Contains `tags cannot be combined` + `--path`
- [ ] `NONEXISTENTTAG --path` now exits 2 (was 0)

### Code Quality / Convention Validation
- [ ] Predicate + message mirror the existing exclusivity-family convention (`skilldozer: … cannot be combined with …`)
- [ ] Tests mirror `TestRunExclusivityTagsAndList` (run-level) + a dedicated direct unit test
- [ ] Doc-comment bullet updated (Mode A honesty)
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] No new files; both edits to `main.go` + `main_test.go`

### Scope Discipline
- [ ] Did NOT touch the mode+mode count set (main.go:715) or the check+mode set (730) — already correct
- [ ] Did NOT touch `parseArgs`, the `config` struct, `run()`, or `--store` (P1.M1.T2 / disjoint)
- [ ] Did NOT change any exit code or family ordering
- [ ] Did NOT add a tags case to `TestExclusivityErrorListingModes` (GOTCHA #4)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`; no README change (Mode B = P1.M3.T1)

---

## Anti-Patterns to Avoid

- ❌ **Don't anchor by line number.** The contract's "702/708/686" are stale. Match the unique predicate text `hasTags && (c.list || c.searchMode || c.all)`. (GOTCHA #1.)
- ❌ **Don't edit only the predicate.** The message must also list `--path` so it matches the predicate's mode set (consistency with every other exclusivity message). (GOTCHA #2.)
- ❌ **Don't add a tags case to `TestExclusivityErrorListingModes`.** Its assertion is `Contains("mutually exclusive")`; a tags case returns "tags cannot be combined" and would fail. Use a dedicated direct-unit test. (GOTCHA #4.)
- ❌ **Don't add a store fixture / env to these tests.** Exclusivity runs before `skillsdir.Find()` — pure argv → exit-2 checks. (GOTCHA #5.)
- ❌ **Don't touch the count set or the check+mode set.** They already include `c.path` correctly; only the tags predicate is buggy. (GOTCHA #7.)
- ❌ **Don't test only one ordering.** The contract requires both `<tag> --path` and `--path <tag>`. (GOTCHA #6.)
- ❌ **Don't add deps/imports or touch README.** Predicate + string + `Contains` use only existing constructs; the README sweep is Mode B (P1.M3.T1). (GOTCHA #8.)

---

## Confidence Score

**9.5/10** — This is a one-line predicate change + one message string + one doc-comment bullet, with the edit site pinned by unique text, the exact fix fixed verbatim by the contract + bug_fixes_validation.md §ISSUE 3, the zero-breakage surface grep-proven (no test asserts the exact tags-message string), and the test templates (`TestRunExclusivityTagsAndList`, the N1 precedent `TestRunExclusivityCheckAndPath`) read at exact line ranges. The fix mirrors a precedent (N1) already applied to the `check` case. The 0.5 reservation is for the one analytical correction to the bug doc's test-placement suggestion (use a dedicated direct-unit test, not the listing-modes table) — resolved and documented in verified_facts §4c, but it is the one place this PRP diverges from the bug doc's literal wording.
