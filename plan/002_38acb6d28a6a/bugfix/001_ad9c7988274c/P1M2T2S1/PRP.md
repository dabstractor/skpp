# PRP — P1.M2.T2.S1: Capture a duplicate reserved `init` token as a conflict in `parseArgs` + test (Issue 4)

> **Subtask:** A small, surgical bugfix. `skilldozer init init` today runs init (exit 0, config written) instead of erroring — inconsistent with `init check` (exit 2). Root cause: `parseArgs`' `case "init":` guard (`next != "check" && next != "init"`, main.go:289) refuses to capture a following `init` as the store (correct intent) but neither captures nor consumes it, so the loop re-processes the second `init` via `case "init":` → re-sets `c.init=true` (idempotent) with NO conflict flag → passes exclusivity → init dispatch runs. The fix (Option A, decisions.md §D3): when the token after `init` is literally `"init"`, **append it to `c.tags`** (and `i++`) instead of swallowing it. Then `exclusivityError`'s EXISTING init+tags branch (main.go:742-743: `if hasTags { return true, "skilldozer: 'init' cannot be combined with tag arguments" }`) fires → exit 2, empty stdout, config NOT written. No new config field, no new exclusivity family, no change to `init <dir>` / `init --store <dir>` / `init check`.
>
> **Scope:** Two existing files only — `main.go` (split the `case "init":` guard into a duplicate-`init`⇒tag branch + a positional-store branch; update the case comment) and `main_test.go` (2 new tests: 1 parseArgs-level locking the capture, 1 run-level locking exit 2 + empty stdout + config-not-written). No new files. No `exclusivityError` change. No new deps.
>
> **STATUS (verified at PRP-write time):** main.go `case "init":` (277-292) + `exclusivityError` init branch (741-746) + run() precedence (438-472: storeMissingValue → exclusivity → init dispatch) read in full; the 4 test-template functions read at exact line ranges. The contract line cites (268-284 / 277 / 281 / 711-716 / 1679) are STALE PRD-write-time numbers — this PRP anchors by TEXT (the unique guard `next != "check" && next != "init"`) and verified-current line numbers. `grep` confirms ZERO existing tests cover `init init` (purely additive) and NONE assert the exact tags-message string (zero breakage). The parallel sibling P1.M2.T1.S1 (Issue 3) edits `exclusivityError`'s tags family (~727), NOT `case "init":` or the init family (~741) — disjoint, no collision.

---

## Goal

**Feature Goal**: Make `skilldozer init init` exit 2 with an empty stdout, the message `skilldozer: 'init' cannot be combined with tag arguments`, and NO config written — exactly as `init check` already behaves — so PRD §6.3 (mutually-exclusive modes) is applied consistently for a duplicate reserved `init` token. After the fix, the only way to use a literal dir named `init` as the store is `init --store init` / `init --store=init` (unchanged).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — split the `case "init":` positional-capture guard (main.go:287-292) so `next == "init"` is appended to `c.tags` (+ `i++`) instead of swallowed; update the case comment (main.go:278-285, Mode A).
2. `main_test.go` — 2 new tests: `TestParseArgsInitInitCapturedAsTag` (parseArgs-level) and `TestRunExclusivityInitInit` (run-level, mirrors `TestRunExclusivityInitAndStrayTag` + adds a config-not-written assertion).

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `parseArgs(["init","init"])` → `{init:true, tags:["init"], initStore:""}`; `run(["init","init"])` → exit 2, empty stdout, stderr Contains `init`, and the `SKILLDOZER_CONFIG`-pointed file is NOT created; the `init init </dev/null` repro now exits 2 (was 0).

---

## User Persona (if applicable)

**Target User**: A user (or a typo / a pasted command) who runs `skilldozer init init` — and any tooling/acceptance suite that asserts `init`-vs-reserved-token conflicts are rejected uniformly.

**Use Case**: `skilldozer init init` — the user accidentally typed `init` twice (or pasted a doubled command). Today it silently runs init (exit 0, writes the config); after the fix it errors clearly like `init check`.

**User Journey**: User runs `skilldozer init init` → (today) init proceeds with auto-detect, exit 0, config written (surprising/inconsistent) → (after fix) `skilldozer: 'init' cannot be combined with tag arguments`, exit 2, nothing written — matching the `init check` precedent.

**Pain Points Addressed**: an inconsistency where `init check` is rejected but `init init` is not; a silent config-write on a clearly-conflicting command.

---

## Why

- **Closes the inconsistency documented in bug_fixes_validation.md §ISSUE 4.** `init check` exits 2 because `check` flows to `case "check":` (sets `c.check`, caught by init+mode); a duplicate `init` has no such flag, so it slipped through.
- **Restores PRD §6.3 uniformity** (mutually-exclusive modes; `init` is its own exclusive mode, peer of `check`).
- **Reuses the already-tested init+tags exclusivity path** (the same path `init foo bar` uses via `TestRunExclusivityInitAndStrayTag`). Option A needs no new field and no new exclusivity family (decisions.md §D3) — it converts the duplicate token into a stray tag that the existing predicate already rejects.
- **Prevents a silent config write.** The contract OUTPUT requires the config is NOT written; because exclusivity fires at run() step 5 (before the init dispatch at step 6 / `runInit` / `config.Save`), exit 2 structurally guarantees it (verified in `research/verified_facts.md` §4).

---

## What

A 3-line guard split (one new `if next == "init"` branch + drop a redundant clause from the existing `else if`), a comment update, and 2 tests. No exit-code change (still 2), no ordering change, no new family, no new field.

### Success Criteria

- [ ] `case "init":` (main.go:277-292): when `args[i+1] == "init"`, append it to `c.tags` and `i++` (consume it); the positional-store branch keeps capturing a non-flag, non-`check`, non-`init` token into `c.initStore`.
- [ ] The 3 unchanged behaviors are byte-identical: `init check` (left for `case "check":` → `c.check` → exit 2), `init --store <dir>` (left for `case "--store":`), `init <dir>` (captured into `c.initStore`).
- [ ] `parseArgs(["init","init"])` → `c.init==true`, `c.tags==["init"]`, `c.initStore==""`.
- [ ] `run(["init","init"])` → exit 2, empty stdout, stderr Contains `init`.
- [ ] The config file (`SKILLDOZER_CONFIG` → a temp path) is NOT created after `run(["init","init"])`.
- [ ] `go test ./...` green; existing tests unaffected (purely additive — `grep`-confirmed no test covers `init init`).
- [ ] `go.mod`/`go.sum` unchanged; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** The single edit site is pinned by the unique guard text `next != "check" && next != "init"` (the only such clause in the file). The fix is fixed verbatim by the contract LOGIC §3 + decisions.md §D3 + bug_fixes_validation.md §ISSUE 4, and traced case-by-case against all 4 following-token cases (`init`/`check`/`--store`/`<dir>`) in `research/verified_facts.md` §2. The mechanism (why appending to `c.tags` makes exclusivity fire) is traced to the exact `exclusivityError` init branch (main.go:741-746, transcribed §3), and the config-not-written guarantee is traced to the run() precedence (storeMissingValue → exclusivity → init dispatch, §4). Zero breakage is grep-proven (no test covers `init init`; no test asserts the exact tags-message string). The 2 test templates (`TestRunExclusivityInitAndStrayTag` @1877, `TestParseArgsInitPositionalDir` @1278) are read at exact line ranges. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (root cause, the 4-case trace, precedence, test plan)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/P1M2T2S1/research/verified_facts.md
  why: "§1 = the root cause (why init init slips through but init check doesn't — the
        second init re-enters case \"init\": idempotently, no flag). §2 = the EXACT fix
        (the before/after guard) + a 4-case trace proving check/--store/<dir> are
        unchanged + WHY the redundant && next != \"init\" is dropped from the else-if.
        §3 = how exclusivity catches it (no new family; reuses the init+tags path at
        main.go:742-743). §4 = run() precedence (exclusivity @step5 before runInit/
        config.Save @step6 ⇒ config NOT written). §5 = grep PROOF of zero breakage.
        §6 = the 2-test plan + templates. §7 = disjointness from P1.M2.T1.S1. §8 = the
        Mode-A comment edit."
  critical: "§2's 4-case trace is the one-pass-stall guard: the naive fix (just remove
             `&& next != \"init\"` from the guard) would capture a duplicate `init` as
             c.initStore (WRONG — it'd be used as the store dir, not a conflict). The
             fix MUST append to c.tags in a DEDICATED branch, not relax the store-capture
             predicate. §4's precedence is why the config-not-written assertion always holds."

# MUST READ — the authoritative bug writeup + repro + prescribed tests
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/bug_fixes_validation.md
  why: "§ISSUE 4 (line 89) is the authoritative repro (init init </dev/null → exit 0,
        config written; init check → exit 2), pins the site (case \"init\": guard), and
        prescribes Option A + the 2 tests (TestParseArgsInitInitDoesNotSwallow +
        TestRunExclusivityInitInit mirroring TestRunExclusivityInitAndCheck). Its line
        cites (268-284/277/281/711-716/1679) are stale — anchor by text per verified_facts."
  section: "ISSUE 4 (Minor)."

# MUST READ — the design decision (Option A over Option B)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/decisions.md
  why: "§D3 (line 21): Option A (capture the duplicate reserved token into c.tags, reusing
        the existing init+tags path) chosen over Option B (a counter field) for minimal
        config-struct churn. Notes a literal store dir named init/check must still use
        --store (already documented in the init case comment)."
  section: "D3."

# MUST READ — the file under edit (read case \"init\": + exclusivityError init branch in full)
- file: main.go
  why: "THE edit target. case \"init\": @277-292 (comment @278-285; c.init=true @286; the
        guard `if i+1 < len(args) { next := args[i+1]; if !strings.HasPrefix(next, \"-\")
        && next != \"check\" && next != \"init\" { c.initStore = next; i++ } }` @287-292).
        exclusivityError init branch @741-746 (hasTags check @742, tags message @743) —
        the CONSUMER, unchanged. run() precedence: storeMissingValue @~450 → exclusivity
        @~460 → `if c.init { return runInit(...) }` @~472 (confirms exit 2 ⇒ config not
        written)."
  pattern: "Reserved-subcommand guard = a `case \"<name>\":` that sets one bool and refuses
            to let the next token be mis-captured. Value capture = `if <ok> { c.field =
            args[i+1]; i++ }`. The fix adds a THIRD shape: append a reserved duplicate to
            c.tags (+ i++) so an existing exclusivity family catches it."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestRunExclusivityInitAndStrayTag
        @1877 is the EXACT run-level template (`init foo bar` → exit 2 via init+tags, the
        SAME path my fix uses; assertions: code==2, empty stdout, Contains(\"tag\")).
        TestRunExclusivityInitAndCheck @1837 is the CONTROL (init+check → exit 2).
        TestParseArgsInitSubcommand @1261 + TestParseArgsInitPositionalDir @1278 are the
        parseArgs-level templates (assert c.init/c.tags/c.initStore directly).
        TestParseArgsInitDirNotCapturedAsTag @1384 is the regression guard for `init <dir>`."
  gotcha: "Exclusivity runs at run() step 5, BEFORE skillsdir.Find() — so the run-level test
           needs NO store fixture / SKILLDOZER_SKILLS_DIR / t.Chdir / unsetSkillsEnv for the
           exit-2 assertion. The config-not-written check adds ONLY a SKILLDOZER_CONFIG
           t.Setenv + an os.Stat(IsNotExist)."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint regions, no collision)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/P1M2T1S1/PRP.md
  why: "Confirms P1.M2.T1.S1 (Issue 3) edits exclusivityError's TAGS family (~main.go:727
        predicate + 728 message + doc bullet) + 3 tags+--path tests. It does NOT touch
        case \"init\": (277-292) or the init exclusivity family (741-746) or any InitAnd*
        test. Disjoint regions in both files; land in either order."

# READ-ONLY — PRD (the authority for mutual exclusivity)
- file: PRD.md
  why: "READ-ONLY. §6.3 (mutually-exclusive modes; init is its own exclusive mode, peer of
        check). §8.2 (init forms). The bugfix PRD h3.3 Issue 4 is the repro."
  section: "§6.3, §8.2 (and the bugfix PRD h3.3 Issue 4)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/tasks.json
  why: "P1.M2.T2.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP
        transcribes it; tasks.json wins on any conflict."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep)
$ grep -n 'next != "check" && next != "init"' main.go   # the buggy guard — unique in the file
# main.go:289 (anchor by TEXT, not number — the contract's 268-284/277/281 are stale)
$ grep -n '"init", "init"\|InitInit' main_test.go   # (empty — no test covers init init today)
```

### Desired Codebase tree with files to be changed

```bash
main.go        # MODIFY parseArgs case "init": guard (split: duplicate init⇒c.tags; else positional store) + update the case comment
main_test.go   # ADD 2 tests: TestParseArgsInitInitCapturedAsTag (parseArgs-level) + TestRunExclusivityInitInit (run-level, +config-not-written)
# go.mod / go.sum — UNCHANGED (no new deps; guard split + c.tags append + Contains/stat use only existing constructs)
```

| File | Change | Owner |
|---|---|---|
| `main.go` | `case "init":` guard: +`if next == "init" { c.tags = append(...); i++ }` before the positional-store branch; drop the redundant `&& next != "init"` from the else-if; update the case comment | Issue 4 contract + decisions.md §D3 |
| `main_test.go` | +2 tests (1 parseArgs-level capture lock + 1 run-level exit-2/empty-stdout/config-not-written) | QA Issue 4 |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL — the #1 one-pass stall) — the fix is NOT "just remove
// `&& next != \"init\"` from the guard". That naive edit would make `init init`
// capture "init" into c.initStore (WRONG: it'd be USED AS THE STORE DIR, not a
// conflict → init proceeds with store="init", config written to store: .../init).
// The fix MUST append the duplicate "init" to c.tags in a DEDICATED `if next == "init"`
// branch (so hasTags becomes true → exclusivity exit 2), and KEEP refusing "init" as a
// store dir. The 4-case trace in research/verified_facts.md §2 is the guard against this.

// GOTCHA #2 — `init check` is UNCHANGED and must keep exiting 2. It works via a
// DIFFERENT path than the fix: `check` is left for case "check": → c.check → caught by
// exclusivityError's init+MODE branch (main.go:744). Do NOT route `check` into c.tags
// (that would also work, but it's unnecessary and changes a tested path). Only `init` is
// split out. The contract LOGIC §3 and bug_fixes_validation.md both note check already works.

// GOTCHA #3 — APPEND to c.tags AND i++. Both are required: append makes hasTags true (so
// exclusivity fires); i++ consumes the second "init" so the for loop does NOT re-process
// it via case "init": (which would re-set c.init idempotently and add nothing to tags,
// silently reverting the fix). If you append WITHOUT i++, the loop's own i++ advances
// past the first init but the second init is then processed by case "init": again →
// c.init re-set, NO tag appended (the `if i+1 < len(args)` is false on the second pass
// since there's no third token) → tags stays empty → exits 0. i++ is load-bearing.

// GOTCHA #4 — anchor the edit by the guard TEXT `next != "check" && next != "init"`
// (unique in main.go), NOT a line number. The contract cites 268-284/277/281; the CURRENT
// lines are 277-292/286/289 (the bugfix round's M1 work shifted things). Same for the
// exclusivityError init branch (contract 711-716; current 741-746) and the control test
// (contract 1679; current 1837).

// GOTCHA #5 — drop the redundant `&& next != "init"` from the positional-store else-if.
// After adding `if next == "init" { … }`, the else-if only runs when next != "init", so
// the clause is logically dead. Dropping it is cleaner and equivalent. (Keeping it is also
// correct — defensive against a future revert of the first branch — but the cleaner form is
// chosen; the test TestParseArgsInitInitCapturedAsTag locks the behavior either way. See
// design decision #3.)

// GOTCHA #6 — the config-not-written check needs SKILLDOZER_CONFIG (NOT
// SKILLDOZER_SKILLS_DIR). runInit writes via config.Path(), which honors SKILLDOZER_CONFIG;
// pointing it at a temp path and asserting os.IsNotExist after run() directly verifies the
// contract OUTPUT. Do NOT need a store tree / SKILLDOZER_SKILLS_DIR / t.Chdir — exclusivity
// runs before skillsdir.Find().

// GOTCHA #7 — exclusivity runs at run() step 5, BEFORE the init dispatch (step 6, runInit).
// So exit 2 structurally guarantees config-not-written — the os.Stat assertion is a
// belt-and-suspenders lock on a structural guarantee, not the only thing preventing a write.

// GOTCHA #8 — No deps/imports change. The guard split + c.tags append + strings.Contains +
// os.Stat use only already-imported constructs (strings + os are imported in both files).
// go.mod/go.sum must be byte-for-byte identical. Verify `git diff --quiet go.mod go.sum`.

// GOTCHA #9 — the `init init` token is appended to c.tags as the literal string "init".
// It is NEVER tag-resolved (exclusivity exits first), so the fact that "init" is a reserved
// subcommand name is irrelevant here — it's just a non-empty tag that trips hasTags. Do not
// special-case the tag value or worry about it shadowing a skill.
```

---

## Implementation Blueprint

### Data models and structure

**No data-model changes.** No new fields, types, methods, or signatures. The fix reuses the existing `c.tags` slice and the existing `exclusivityError` init+tags branch.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — split the case "init": guard (the core fix)
  - FILE: main.go (parseArgs, case "init":, the `if i+1 < len(args) { … }` block ~287-292;
    anchor by the unique guard text per GOTCHA #4)
  - FIND (the current block):
        c.init = true
        if i+1 < len(args) {
            next := args[i+1]
            if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
                c.initStore = next
                i++
            }
        }
  - REPLACE with (GOTCHA #1 — DEDICATED init branch that appends to c.tags; GOTCHA #3 —
    append AND i++; GOTCHA #5 — drop the redundant `&& next != "init"` from the else-if):
        c.init = true
        if i+1 < len(args) {
            next := args[i+1]
            if next == "init" {
                // Issue 4: a duplicate reserved `init` token is a conflict, not a store
                // dir. Capture it as a tag so exclusivityError's init+tags branch rejects
                // `init init` with exit 2 (consistent with `init check`, where the second
                // token reaches case "check" and sets c.check). A literal store dir named
                // "init" must still be passed via --store.
                c.tags = append(c.tags, next)
                i++
            } else if !strings.HasPrefix(next, "-") && next != "check" {
                c.initStore = next
                i++
            }
            // else: a dashed flag (`init --store …`) or `init check` → left for its own case.
        }
  - This is the ENTIRE logic fix. exclusivityError is UNCHANGED (its init+tags branch at
    main.go:742-743 already rejects hasTags). Trace the 4 cases in verified_facts §2.

Task 2: EDIT main.go — update the case "init": comment (Mode A, per contract DOCS)
  - FILE: main.go (case "init": comment, ~278-285)
  - FIND the sentence:
        // flag (`init --store …`) or subcommand (`init check`) is left for its
        // own case so exclusivity can flag the conflict. GOTCHA: a store
        // literally named `check`/`init` must be passed via `--store`.
  - REPLACE with (note the duplicate-init conflict; keep the --store GOTCHA):
        // flag (`init --store …`) or subcommand (`init check`) is left for its
        // own case so exclusivity can flag the conflict. A DUPLICATE `init`
        // (`init init`) is captured into c.tags below so the init+tags exclusivity
        // branch rejects it with exit 2 (Issue 4) — consistent with `init check`.
        // GOTCHA: a store literally named `check`/`init` must be passed via `--store`.

Task 3: EDIT main_test.go — add the parseArgs-level test (locks the capture)
  - FILE: main_test.go (insert right after TestParseArgsInitDirNotCapturedAsTag, ~1395,
    grouped with the other parseArgs Init tests; package main)
  - ADD:
    // Issue 4 (P1.M2.T2.S1): a duplicate reserved `init` token is captured into c.tags
    // (not swallowed, not used as the store) so exclusivityError's init+tags branch
    // rejects `init init` with exit 2. A literal store dir named "init" must use --store.
    func TestParseArgsInitInitCapturedAsTag(t *testing.T) {
        c := parseArgs([]string{"init", "init"})
        if !c.init {
            t.Errorf("parseArgs(init init): init=false; want true")
        }
        if len(c.tags) != 1 || c.tags[0] != "init" {
            t.Errorf("parseArgs(init init): tags=%v; want [init] (duplicate init captured as a tag)", c.tags)
        }
        if c.initStore != "" {
            t.Errorf("parseArgs(init init): initStore=%q; want empty (init is NOT a store dir)", c.initStore)
        }
    }

Task 4: EDIT main_test.go — add the run-level test (locks exit 2 + empty stdout + config NOT written)
  - FILE: main_test.go (insert right after TestRunExclusivityInitAndStrayTag, ~1893,
    grouped with the other InitAnd* exclusivity tests; package main)
  - ADD (mirrors TestRunExclusivityInitAndStrayTag @1877 + the config-not-written check
    per GOTCHA #6/#7; NO store fixture / SKILLDOZER_SKILLS_DIR / t.Chdir needed):
    // Issue 4 (P1.M2.T2.S1): `init init` now exits 2 (was 0, config written). The
    // duplicate `init` is captured as a tag, so exclusivityError's init+tags branch
    // fires (the same path `init foo bar` uses). exclusivity runs at run() step 5,
    // BEFORE runInit/config.Save (step 6), so exit 2 guarantees the config is NOT written.
    func TestRunExclusivityInitInit(t *testing.T) {
        cfg := filepath.Join(t.TempDir(), "must-not-exist.yaml")
        t.Setenv("SKILLDOZER_CONFIG", cfg)
        var out, errOut bytes.Buffer
        code := run([]string{"init", "init"}, &out, &errOut)
        if code != 2 {
            t.Fatalf("run(init init): code=%d; want 2 (Issue 4: duplicate init is a conflict)", code)
        }
        if out.Len() != 0 {
            t.Errorf("stdout=%q; want empty", out.String())
        }
        if !strings.Contains(errOut.String(), "init") {
            t.Errorf("stderr=%q; want a message mentioning init", errOut.String())
        }
        // Contract OUTPUT: the config is NOT written (exclusivity fires before init dispatch).
        if _, err := os.Stat(cfg); !os.IsNotExist(err) {
            t.Errorf("config %s was written; exclusivity must fire before init dispatch (got err=%v)", cfg, err)
        }
    }

Task 5: VERIFY in isolation + whole module + invariants
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0
  - go test -run 'TestParseArgsInitInitCapturedAsTag|TestRunExclusivityInitInit' -v ./...  # the 2 new tests pass
  - go test ./...                     # whole module green; zero regressions (GOTCHA zero-breakage)
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #8
  - REGRESSION (unchanged cases still correct):
      go test -run 'TestParseArgsInitPositionalDir|TestParseArgsInitDirNotCapturedAsTag|TestRunExclusivityInitAndCheck|TestParseArgsInitStoreLongForm' -v ./...
  - END-TO-END (the §ISSUE 4 repro, now fixed): build, run `init init </dev/null` → exit 2 (was 0).
```

### Implementation Patterns & Key Details

```go
// The fix (Task 1) — a dedicated `next == "init"` branch that appends to c.tags (+ i++),
// keeping check/--store/<dir> behavior identical. Anchored by the unique guard text.
c.init = true
if i+1 < len(args) {
	next := args[i+1]
	if next == "init" {
		// Issue 4: duplicate reserved `init` → tag → init+tags exclusivity exit 2.
		c.tags = append(c.tags, next)
		i++
	} else if !strings.HasPrefix(next, "-") && next != "check" {
		c.initStore = next // positional <dir> (`init <dir>`); `init check` left for case "check".
		i++
	}
}

// How it is caught (UNCHANGED exclusivityError, main.go:741-743):
if c.init {
	if hasTags { // len(c.tags) > 0 — true now for `init init` (c.tags == ["init"])
		return true, "skilldozer: 'init' cannot be combined with tag arguments"
	}
	...
}
```

Notes easy to get wrong:
- The fix appends to `c.tags` in a **dedicated** `if next == "init"` branch — do NOT relax the store-capture predicate by merely dropping `&& next != "init"` (that would capture "init" as the store dir, GOTCHA #1).
- `append` **and** `i++` are both required (GOTCHA #3): append makes `hasTags` true; `i++` consumes the token so the loop doesn't re-process it through `case "init":` (which would add nothing).
- Only `next == "init"` is split out; `next == "check"` is unchanged (it already flows to `case "check":` → `c.check`, GOTCHA #2).
- The run-level test needs no store fixture (exclusivity runs before `skillsdir.Find()`); the config-not-written check uses `SKILLDOZER_CONFIG` + `os.Stat`, not `SKILLDOZER_SKILLS_DIR` (GOTCHA #6).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Option A (capture into `c.tags`), not Option B (a counter field).** decisions.md §D3 chose A for minimal config-struct churn: it reuses the existing, already-tested init+tags exclusivity path (the same one `init foo bar` exercises via `TestRunExclusivityInitAndStrayTag`). No new field, no new exclusivity family, no new message.
2. **Only `next == "init"` is split out; `check` is untouched.** `init check` already exits 2 via `case "check":` → `c.check` → init+mode. Routing `check` into `c.tags` too would also work but would change a tested path for no benefit. The contract LOGIC §3 and the bug doc both scope the change to the duplicate `init`.
3. **Drop the redundant `&& next != "init"` from the positional-store else-if.** After the dedicated `if next == "init"` branch, the else-if provably never sees `next == "init"`, so the clause is dead. Dropping it is cleaner; the alternative (keeping it as a defensive guard against a future revert) is also correct. `TestParseArgsInitInitCapturedAsTag` locks the behavior regardless.
4. **Both a parseArgs-level test and a run-level test.** The parseArgs test isolates the exact capture (`c.tags == ["init"]`); the run test proves the end-to-end contract (exit 2, empty stdout, config not written). bug_fixes_validation.md §ISSUE 4 prescribed both. The test name `TestParseArgsInitInitCapturedAsTag` is clearer than the bug doc's `TestParseArgsInitInitDoesNotSwallow`; the run-level name `TestRunExclusivityInitInit` matches the bug doc and the `TestRunExclusivityInit*` family.
5. **A config-not-written assertion (os.Stat on a temp SKILLDOZER_CONFIG) in the run test.** The contract OUTPUT explicitly requires "the config is NOT written". exclusivity-before-dispatch structurally guarantees it (GOTCHA #7), but the assertion makes the guarantee explicit and regression-proof — it would fail if a future change reordered dispatch ahead of exclusivity.
6. **No README change here.** The contract DOCS assigns the README sweep to the final Mode B task (P1.M3.T1). This subtask's doc edit is the in-code `case "init":` comment only (Mode A, Task 2).

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports (guard split + c.tags append + Contains + os.Stat).
    (GOTCHA #8)

DISPATCH (unchanged): exclusivityError still runs at run() step 5, after storeMissingValue
  (step 4) and before the init dispatch (step 6, runInit). `init init` skips step 4 (no
  --store) and hits the init+tags branch at step 5 → exit 2; runInit/config.Save never runs.

CONSUMERS:
  - The §ISSUE 4 repro: `init init </dev/null` now exits 2 (was 0).
  - README init/error-contract section: swept by the final Mode B task (P1.M3.T1) — no doc
    file rides here beyond the case "init": comment (Mode A, Task 2).

PARALLEL SIBLING (no conflict):
  - P1.M2.T1.S1 (Issue 3) edits exclusivityError's TAGS family (~main.go:727 predicate +
    728 message + doc bullet) + 3 tags+--path tests. This subtask edits parseArgs case
    "init": (~287-292) + its comment (~278-285) + 2 init-init tests. DISJOINT regions in
    both files; no text-level overlap; land in either order.

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

go test -run 'TestParseArgsInitInitCapturedAsTag|TestRunExclusivityInitInit' -v ./...
# Expected: BOTH pass. The load-bearing assertions:
#   TestParseArgsInitInitCapturedAsTag -> parseArgs(["init","init"]) c.init=true, c.tags==["init"], c.initStore="".
#   TestRunExclusivityInitInit         -> run(["init","init"]) code 2, empty stdout, stderr Contains "init", config NOT written.

# Regression — the unchanged init cases stay green (check/--store/<dir> byte-identical):
go test -run 'TestParseArgsInitPositionalDir|TestParseArgsInitDirNotCapturedAsTag|TestRunExclusivityInitAndCheck|TestParseArgsInitStoreLongForm|TestParseArgsInitStoreEqualsForm|TestRunExclusivityInitAndStrayTag' -v ./...
# Expected: PASS (the fix only adds the next=="init" branch; the other 3 cases are unchanged).
```

### Level 3: Whole-module regression + the §ISSUE 4 repro (now fixed)

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions (grep-proven additive)

# GOTCHA #8 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Scope invariant: the dedicated init branch is present exactly once
grep -c 'next == "init"' main.go   # expect 1

# END-TO-END — the §ISSUE 4 repro, now FIXED (init init </dev/null → exit 2, was 0):
go build -o /tmp/sd .
cfg=$(mktemp -d)/cfg.yaml
/tmp/sd init init </dev/null >/tmp/o 2>/tmp/e; rc=$?
SKILLDOZER_CONFIG="$cfg" /tmp/sd init init </dev/null >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" = 2 ] && [ ! -s /tmp/o ] && grep -q 'init' /tmp/e && [ ! -e "$cfg" ] \
  && echo "Issue-4 repro FIXED: init init → exit 2 + empty stdout + msg + config NOT written" \
  || echo "FAIL: rc=$rc out=$(cat /tmp/o) err=$(cat /tmp/e) cfg-exists=$([ -e "$cfg" ] && echo yes)"
# Control (unchanged): init check still exits 2.
/tmp/sd init check </dev/null >/dev/null 2>&1; echo "check-control exit=$? (want 2)"
rm -f /tmp/sd /tmp/o /tmp/e; rm -rf "$(dirname "$cfg")"
# Expected: "Issue-4 repro FIXED …"; check-control exit 2.
```

### Level 4: Behavioral spot-checks (lock the fix is scoped, no over-reach)

```bash
cd /home/dustin/projects/skilldozer

# 4a. `init init init` (3x) also exits 2 (the first duplicate already trips hasTags):
go build -o /tmp/sd .
/tmp/sd init init init </dev/null >/dev/null 2>&1; echo "init init init exit=$? (want 2)"

# 4b. A literal store dir named "init" still works via --store (the GOTCHA holds):
store=$(mktemp -d)/init && mkdir -p "$store"
SKILLDOZER_CONFIG=$(mktemp -d)/c.yaml /tmp/sd init --store "$store" </dev/null >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" = 0 ] && echo "init --store <dir-named-init> → exit 0 (GOTCHA holds: --store is the escape hatch)" \
  || echo "FAIL: rc=$rc err=$(cat /tmp/e)"
rm -rf "$store" /tmp/sd /tmp/o /tmp/e

# 4c. `init <dir>` (positional, NOT named init) still captures the store, NOT a conflict:
store=$(mktemp -d)/skills && mkdir -p "$store"
SKILLDOZER_CONFIG=$(mktemp -d)/c.yaml /tmp/sd init "$store" </dev/null >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" = 0 ] && echo "init <dir> → exit 0 (positional store capture unchanged)" \
  || echo "FAIL: rc=$rc err=$(cat /tmp/e)"
rm -rf "$store" /tmp/sd /tmp/o /tmp/e
# Expected: all three spot-checks pass (the fix only affects the literal duplicate `init`).
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l main.go main_test.go` empty; `go vet ./...` exit 0; `go build` exit 0
- [ ] Level 2 PASS — the 2 new tests pass; the unchanged init-case regression tests stay green
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (zero regressions); `git diff go.mod go.sum` → "deps unchanged"; the §ISSUE 4 repro now exits 2 + empty stdout + msg + config NOT written
- [ ] Level 4 PASS — `init init init` exits 2; `init --store <dir-named-init>` exits 0; `init <dir>` exits 0

### Feature Validation
- [ ] `parseArgs(["init","init"])` → `{init:true, tags:["init"], initStore:""}`
- [ ] `run(["init","init"])` → exit 2, empty stdout, stderr Contains `init`
- [ ] The `SKILLDOZER_CONFIG`-pointed file is NOT created after `run(["init","init"])`
- [ ] `init init </dev/null` now exits 2 (was 0)
- [ ] `init check` still exits 2 (unchanged); `init <dir>` and `init --store <dir>` still work

### Code Quality / Convention Validation
- [ ] The dedicated `if next == "init"` branch mirrors the existing reserved-token-guard style (a `case`-internal sub-branch that routes a special token)
- [ ] Tests mirror `TestRunExclusivityInitAndStrayTag` (run-level) + `TestParseArgsInitPositionalDir` (parseArgs-level)
- [ ] The `case "init":` comment updated (Mode A honesty, citing Issue 4)
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] No new files; both edits to `main.go` + `main_test.go`

### Scope Discipline
- [ ] Did NOT relax the store-capture predicate (used a dedicated c.tags branch — GOTCHA #1)
- [ ] Did NOT route `check` into c.tags (it already works via case "check" — GOTCHA #2)
- [ ] Did NOT touch `exclusivityError` (the init+tags branch already rejects hasTags)
- [ ] Did NOT add a config field (Option A, not Option B — decisions.md §D3)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`; no README change (Mode B = P1.M3.T1)

---

## Anti-Patterns to Avoid

- ❌ **Don't "just remove `&& next != "init"`" from the guard.** That captures a duplicate `init` as `c.initStore` (the store DIR), not a conflict — init proceeds, config written to `store: .../init`. Use a DEDICATED `if next == "init" { c.tags = append(...); i++ }` branch. (GOTCHA #1.)
- ❌ **Don't append to c.tags without `i++`.** Without consuming the token, the loop re-processes the second `init` via `case "init":` (idempotent re-set, no tag) → tags stays empty → exit 0. Both are required. (GOTCHA #3.)
- ❌ **Don't route `check` into c.tags too.** `init check` already exits 2 via `case "check":` → `c.check`; changing it is unnecessary and mutates a tested path. Only `init` is split out. (GOTCHA #2.)
- ❌ **Don't anchor by line number.** The contract's 268-284/277/281/711-716/1679 are stale; current lines are 277-292/286/289/741-746/1837. Match the unique guard text `next != "check" && next != "init"`. (GOTCHA #4.)
- ❌ **Don't add a store fixture / SKILLDOZER_SKILLS_DIR / t.Chdir to the run test.** Exclusivity runs before `skillsdir.Find()`; only the config-not-written check needs `SKILLDOZER_CONFIG` + `os.Stat`. (GOTCHA #6.)
- ❌ **Don't touch `exclusivityError` or add a config field.** The init+tags branch already exists; Option A reuses it. (Scope discipline.)
- ❌ **Don't add deps/imports or touch README.** Guard split + `c.tags` append + `Contains` + `os.Stat` use only existing constructs; the README sweep is Mode B (P1.M3.T1). (GOTCHA #8.)

---

## Confidence Score

**9.5/10** — This is a 3-line guard split (one dedicated `if next == "init"` branch that appends to `c.tags` + `i++`, and dropping a redundant clause from the else-if) + a comment update + 2 tests. The edit site is pinned by unique text (`next != "check" && next != "init"`), the exact fix is fixed verbatim by the contract LOGIC §3 + decisions.md §D3 + bug_fixes_validation.md §ISSUE 4, and traced case-by-case against all 4 following-token cases in `research/verified_facts.md` §2. The mechanism (c.tags → hasTags → existing init+tags exclusivity exit 2) is traced to the exact `exclusivityError` branch (main.go:741-743), and the config-not-written guarantee is traced to the run() precedence. Zero breakage is grep-proven (no test covers `init init`; no test asserts the exact tags-message string). The fix reuses an already-tested exclusivity path (the `init foo bar` path via `TestRunExclusivityInitAndStrayTag`). The 0.5 reservation is for the single most-likely one-pass stall — the naive "just drop `&& next != \"init\"`" mis-fix that would capture `init` as the store dir instead of a conflict (GOTCHA #1) — which the dedicated-branch instruction + the 4-case trace + `TestParseArgsInitInitCapturedAsTag` (asserting `c.initStore==""`) jointly guard against.
