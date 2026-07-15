# PRP — P1.M2.T1.S1: Add missing-value exit-2 detection for `--search` and `--shell` (Issue 3)

> **Subtask:** P1.M2.T1.S1 — the sole subtask of P1.M2.T1 (Issue 3). Makes `--search`/`-s` and `--shell` presented WITHOUT their value exit 2 with a precise error message, mirroring the existing `storeMissingValue` pattern EXACTLY (bool field in `config`, set in `parseArgs` no-value branches, checked in `run()` before exclusivity dispatch). Today they silently fall through to implicit help (stdout usage, exit 0) — and the code comment at main.go:293 even mislabels that path as "(exit 1)" when it is actually exit 0.
>
> **Scope:** Two existing files only — `main.go` (2 config fields + 3 parseArgs branches + 2 run() checks + 3 comment fixes) and `main_test.go` (1 rename+flip, 1 update, 4 new tests). No new files. No `internal/*` change. Zero new deps (`fmt.Fprintln` already imported). go.mod/go.sum byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** read `main.go` (config struct / parseArgs `--search`+`--shell` cases / `expandShortBundle` -s default / run() `storeMissingValue` check) + `main_test.go` (store + search/shell tests) at exact line ranges; `decisions.md` §D4-D5 + `parseargs_research.md` read in full. The parallel sibling P1.M1.T2.S1 edits `internal/skillsdir/*` (Issue 2) and does NOT touch `main.go`/`main_test.go` — no collision. grep-confirmed: NO run-level test locks the old exit-0 behavior, so the only test churn is the documented rename+flip+update+adds.

---

## Goal

**Feature Goal**: Make missing-value handling symmetrical across all value-taking flags (decisions D4). `--store` (no value) already exits 2; after this fix `--search`/`-s` (no value) → exit 2 with `skilldozer: --search requires a query`, and `--shell` (no value) → exit 2 with `skilldozer: --shell requires a value (bash|zsh|fish)`. This makes `$(skilldozer --search)` fail loudly (empty stdout, exit 2) instead of silently capturing the help text — predictable for `$(...)` use (PRD §6.4).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — (a) `searchMissingValue bool` + `shellMissingValue bool` fields after `storeMissingValue`; (b) `else { c.searchMissingValue = true }` in the `--search`/`-s` case + the `expandShortBundle` -s default; (c) `else { c.shellMissingValue = true }` in the `--shell` case; (d) the two peer `if c.XxxMissingValue { … return 2 }` checks in `run()` after `storeMissingValue`; (e) fix the 3 misleading comments (the `(exit 1)` → exit-2 description, the `--shell` "silent" comment, the `storeMissingValue` precedent comment).
2. `main_test.go` — rename+flip `TestParseArgsSearchNoValueStaysInactive` → `TestParseArgsSearchMissingValue`; update `TestParseArgsShortBundleSearchNoValue` to assert the new signal; add `TestRunSearchNoValueExits2`, `TestRunSearchShortNoValueExits2`, `TestRunShellNoValueExits2`, `TestParseArgsShellMissingValue`.

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `run(["--search"])`, `run(["-s"])`, `run(["--shell"])` → exit 2 with empty stdout + the exact stderr message; `--search foo` / `--completions --shell bash` (value present) unchanged; the `=`-forms `--search=`/`--shell=` unchanged (decisions D5).

---

## User Persona (if applicable)

**Target User**: a user/script typing `skilldozer --search` and forgetting the query, or `skilldozer --shell` without the shell name; and the `$(skilldozer --search …)` command-substitution consumer.

**Use Case**: `Q="$(skilldozer --search foo)"` — today a typo'd `skilldozer --search` (no value) silently captures the help text into `$Q`; after the fix it fails loudly (exit 2, empty `$Q`).

**User Journey**: User types `skilldozer --search` `<enter>` → (today) help printed, exit 0, confusing → (after fix) clear `--search requires a query` on stderr, exit 2, nothing on stdout.

**Pain Points Addressed**: silent help-on-error that pollutes `$(...)`; the internal inconsistency (`--store` errors but `--search`/`--shell` don't); the misleading code comment.

---

## Why

- **decisions D4 (the authoritative choice):** "Make `--search`/`-s` and `--shell` no-value exit 2 with an error message, mirroring `--store`." The alternative (document the asymmetry + fix the comment) was REJECTED — the PRD says "The symmetrical option is more predictable for `$(...)` use."
- **PRD §6.4 (error semantics):** a value-taking flag presented without its value is a parse error. `--store` already enforces this; `--search`/`--shell` should too, so `$(skilldozer --search)` fails loudly rather than capturing help.
- **The code comment lies:** main.go:293 claims the `--search` no-value path yields "(exit 1)" but it actually exits 0 (parseargs_research.md §4, empirically confirmed). The fix both changes the behavior AND corrects the comment.
- **Closes Issue 3** (the bugfix-round-2 QA finding): the value-taking flags were handled inconsistently.

---

## What

A surgical change mirroring the proven `storeMissingValue` pattern: 2 new bool fields, 3 no-value branches that set them, 2 run() checks that exit 2, 3 comment fixes. No exit-code change anywhere else; the success paths (`--search foo`, `--completions --shell bash`) and the `=`-forms (`--search=`, `--shell=`) are unchanged (decisions D5).

### Success Criteria

- [ ] `config` has `searchMissingValue bool` and `shellMissingValue bool` after `storeMissingValue`.
- [ ] `--search`/`-s` main-switch no-value → `c.searchMissingValue = true`; `-s` short-bundle no-value → `c.searchMissingValue = true`.
- [ ] `--shell` main-switch no-value → `c.shellMissingValue = true`.
- [ ] `run()` has the two peer checks after `storeMissingValue`: `--search requires a query` and `--shell requires a value (bash|zsh|fish)`, each `return 2` with empty stdout.
- [ ] The 3 misleading comments are corrected (the `(exit 1)` → exit-2 description; the `--shell` "silent" comment; the `storeMissingValue` precedent comment notes --search/--shell).
- [ ] `=`-forms UNCHANGED: `--search=` → searchMode=true, searchQ=""; `--shell=` → completion=true, shell="" (no missing-value guard — D5).
- [ ] `run(["--search"])`, `run(["-s"])`, `run(["--shell"])` → exit 2, empty stdout, exact stderr; `run(["--search","foo"])` / `run(["--completions","--shell","bash"])` unchanged.
- [ ] `go test ./...` green; `go.mod`/`go.sum` unchanged; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit is pinned to an exact line with before→after text (read from the live file + `parseargs_research.md` §1-§4). The template (`storeMissingValue`) is read in full. The two non-obvious points — (a) the `=`-forms stay UNCHANGED (D5), and (b) bare `-s` (len==2) goes through the main switch, NOT `expandShortBundle` — are confirmed. The test churn is enumerated precisely (1 rename+flip, 1 update, 4 adds; no run-level test locks the old exit-0). An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (exact lines, the 3 no-value forms, the test churn)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M2T1S1/research/verified_facts.md
  why: "§2 the storeMissingValue template (field+set+read). §3 exact main.go anchors (sibling-safe). §4 the
        three --search/-s no-value forms + which get the fix (only 1b bare + 1c bundle; NOT 1a =-form). §5 the
        --shell fix. §6 the two run() checks (exact messages + ordering). §7 the 3 comment fixes. §8 the test
        churn (1 rename+flip, 1 update, 4 adds; the =-form test stays green). §9 disjoint from sibling."
  critical: "§4 (the =-form stays unchanged — D5) and §8 (bare -s goes through the main switch, NOT the bundle)
             are the two things most likely to be mishandled."

# MUST READ — the design decisions (D4 symmetrical exit 2; D5 leave =-form as-is)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/decisions.md
  why: "§D4 chose symmetrical exit 2 (REJECTED the 'document the asymmetry' alternative). §D5: the =-form empty
        values (--search=/--shell=) keep their current behavior; ONLY the bare no-token case gets exit 2. These
        pin the exact scope — do NOT add missing-value guards to the =-form switch."

# MUST READ — the parseArgs/run analysis (exact line numbers + the misleading comment proof)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/parseargs_research.md
  why: "§1 enumerates the 3 --search/-s no-value forms (=-form @218, main switch @288, bundle @434) with exact
        code. §2 the --shell forms. §3 the storeMissingValue set/read sites + the run() precedence ladder. §4
        PROVES the (exit 1) comment is wrong (actual: exit 0). §5 the default-branch classification (out of
        scope for this task — Issue 4). The empirical matrix confirms today's --search/-s/--shell → exit 0."

# MUST READ — the authoritative bug writeup
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/issue_analysis.md
  why: "§Issue 3 is the authoritative repro (--search/--shell → exit 0; --store → exit 2; the misleading
        main.go:293 comment) + the fix prescription (symmetrical exit 2, mirroring --store)."
  section: "Issue 3 (Minor)."

# MUST READ — the file under edit (read the config struct + the 2 parseArgs cases + the bundle + run() before editing)
- file: main.go
  why: "THE edit target. config struct @153-174 (storeMissingValue @169 — add the 2 fields after it). main switch
        case '--search','-s' @288-298 (the if/else to add; the misleading comment @290-293). case '--shell' @320-326
        (the if/else; comment @323-324). expandShortBundle -s default @444 (add c.searchMissingValue=true). run()
        storeMissingValue check @499-502 (add the 2 peer checks after; precedent comment @493-498)."
  pattern: "storeMissingValue = the EXACT template: bool field @169; set in parseArgs no-value `else { c.X = true }`;
            read in run() `if c.X { fmt.Fprintln(stderr, \"skilldozer: ... requires ...\"); return 2 }` at step 3.5."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestRunInitStoreNoValueExits2 @321 / TestRunStoreBareNoValueExits2
        @350 = the run-level exit-2 template (code==2, empty stdout, EXACT stderr via `errOut.String() != want`).
        TestParseArgsSearchNoValueStaysInactive @1004 = RENAME+flip to TestParseArgsSearchMissingValue. TestParseArgsShort
        BundleSearchNoValue @2356 = UPDATE (add searchMissingValue assertion). TestParseArgsLongEqualsSearchEmpty @2294
        = UNCHANGED (=-form, D5) — leave it green."
  gotcha: "Assert the run-level stderr with EXACT equality (errOut.String() == \"skilldozer: --search requires a
           query\\n\"), mirroring the store tests — the contract fixes the message verbatim. Do NOT use Contains."

# READ-ONLY — the parallel sibling PRP (boundary: edits internal/skillsdir/* only, NOT main.go/main_test.go)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M1T2S1/PRP.md
  why: "Confirms P1.M1.T2.S1 (Issue 2, vanished-store) edits internal/skillsdir/skillsdir.go + skillsdir_test.go.
        Its PRP states 'Does NOT touch main.go'. Disjoint from this task's files (main.go + main_test.go); land
        in either order (P1.M1 before P1.M2 by plan ordering, but even concurrent the files don't overlap)."

# READ-ONLY — PRD (the authority for the §6.4 error semantics)
- file: PRD.md
  why: "READ-ONLY. §6.1 (flag matrix), §6.4 (error semantics: a missing value is a parse error). The bugfix-2 PRD
        context's §Issue 3 + decisions D4/D5 are the operative authority. Do NOT edit PRD.md."
  section: "h2.3/h3.2 (Issue 3) in the bugfix requirements doc."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/tasks.json
  why: "P1.M2.T1.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes it; tasks.json
        wins on any conflict."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep)
$ grep -n 'searchMissingValue\|shellMissingValue' main.go main_test.go   # (empty today — purely additive fields)
# main.go: 1335 lines. config @153; parseArgs @182-358; expandShortBundle @361-450; run @466-750.
```

### Desired Codebase tree with files to be changed

```bash
main.go        # ADD: 2 config fields; 3 no-value setters (search main-switch, search bundle, shell main-switch);
               #     2 run() peer checks; 3 comment fixes
main_test.go   # ADD/FLIP: rename+flip SearchNoValueStaysInactive→SearchMissingValue; update ShortBundleSearchNoValue;
               #         +4 new tests (3 run-level exit-2 + 1 parseArgs shell)
# go.mod / go.sum — UNCHANGED (zero new deps; fmt.Fprintln + bool fields + *config already in scope)
```

| File | Responsibility |
|---|---|
| `main.go` | Symmetrical missing-value exit-2 for `--search`/`-s`/`--shell`, mirroring `storeMissingValue`; correct the misleading comments. |
| `main_test.go` | Lock the new exit-2 behavior (run-level) + the new signals (parseArgs-level); flip the stale assertions. |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — Do NOT touch the =-form switch (decisions D5). --search= (main.go:218) and --shell= (main.go:248)
// keep their UNCONDITIONAL behavior (searchMode=true/completion=true + empty value). The missing-value fix is
// ONLY for the bare no-token case (main switch `else` + bundle `default`). The =-form test
// TestParseArgsLongEqualsSearchEmpty (@2294) locks searchMode=true,searchQ="" and must stay GREEN unchanged.
// Adding a guard to the =-form is scope creep D5 explicitly rejects.

// GOTCHA #2 — Bare `-s` (len==2) goes through the MAIN SWITCH case "--search","-s" (main.go:288), NOT
// expandShortBundle (which requires len(a) > 2 at main.go:259). So `run([]string{"-s"})` hits the main-switch
// `else` → searchMissingValue=true → exit 2. The bundle case (`-vs`,`-ls` last token) is SEPARATE (GOTCHA #3).
// parseargs_research.md §1c note confirms this.

// GOTCHA #3 — The expandShortBundle -s default (main.go:444) is the THIRD setter. A bundle like `-vs` (last
// token, s present, no value) sets searchMissingValue=true there (c is *config — direct assignment is valid).
// Update TestParseArgsShortBundleSearchNoValue (@2356) to assert it (it currently only checks version/searchMode
// and stays green either way, but the assertion locks the new bundle behavior). Do NOT add a run-level -vs test
// (the contract OUTPUT covers bare -s via the main switch, not the bundle).

// GOTCHA #4 — EXACT stderr messages (contract LOGIC (e), verbatim): "skilldozer: --search requires a query"
// and "skilldozer: --shell requires a value (bash|zsh|fish)". The `skilldozer:` prefix + `requires a …` phrasing
// mirror "skilldozer: --store requires a value" (main.go:500). Use fmt.Fprintln (fixed string, no args). Assert
// with EXACT equality (errOut.String() == want+"\n") in the run-level tests — do NOT use Contains.

// GOTCHA #5 — The two run() checks go AFTER storeMissingValue (main.go:502), BEFORE exclusivity (508). They are
// peer parse-error guards (steps 3.6/3.7). So --help/--version still WIN (--help --search → exit 0 help, same as
// --help --store). Do NOT move them before help/version/unknownFlag.

// GOTCHA #6 — searchMode/completion stay FALSE on the no-value path. The `else` only sets the missing-value
// signal; it does NOT set searchMode/completion (no value was consumed). TestParseArgsSearchMissingValue asserts
// BOTH searchMissingValue=true AND searchMode=false (regression guard). Same for shell: shellMissingValue=true,
// completion=false.

// GOTCHA #7 — The misleading comment at main.go:293 says "(exit 1)" but the actual fall-through (main.go:749) is
// exit 0 (parseargs_research.md §4, empirically confirmed). After the fix it is exit 2. Fix the comment to the
// exit-2 description (do NOT just change "(exit 1)" to "(exit 0)" — the behavior is changing to exit 2). Same for
// the --shell "silent behavior" comment @324.

// GOTCHA #8 — No conflict with the parallel sibling P1.M1.T2.S1 (internal/skillsdir/* — Issue 2). Disjoint files;
// this task is main.go + main_test.go only. Land in either order.

// GOTCHA #9 — No deps/imports change. fmt.Fprintln is already imported in main.go; the new fields are bools;
// expandShortBundle already takes c *config. go.mod/go.sum byte-for-byte identical. Verify with
// `git diff --quiet go.mod go.sum`.

// GOTCHA #10 — The success paths are UNCHANGED. --search foo → search dispatch (exit 0); --completions --shell
// bash → completion dispatch; --store /tmp/x → init. Do NOT alter the value-present branches. Only the no-value
// `else`/`default` arms + the run() guards are new.
```

---

## Implementation Blueprint

### Data models and structure

**Two new bool fields on the existing `config` struct.** No new types/structs/signatures.

```go
// (after storeMissingValue, main.go:169)
searchMissingValue bool     // --search/-s seen with NO following value (Issue 3); run() exits 2 with "--search requires a query"
shellMissingValue  bool     // --shell seen with NO following value (Issue 3); run() exits 2 with "--shell requires a value (bash|zsh|fish)"
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add the two config fields
  - FILE: main.go (type config struct, ~line 169)
  - PLACE the two new fields IMMEDIATELY AFTER `storeMissingValue bool` (line 169) and BEFORE `completion bool`.
  - ADD (doc comments cite Issue 3, matching the storeMissingValue style):
      searchMissingValue bool     // --search/-s seen with NO following value (Issue 3); run() exits 2 with "--search requires a query" before dispatch. NOT set by --search= (D5: empty =-form value is a valid empty query).
      shellMissingValue  bool     // --shell seen with NO following value (Issue 3); run() exits 2 with "--shell requires a value (bash|zsh|fish)". NOT set by --shell= (D5).
  - gofmt -w fixes column alignment.

Task 2: EDIT main.go — the --search/-s main-switch no-value else + comment fix
  - FILE: main.go (case "--search", "-s": ~line 288-298)
  - (2a) FIND the if-block (no else today):
        if i+1 < len(args) {
            c.searchMode = true
            c.searchQ = args[i+1]
            i++
        }
    REPLACE with (add the else — GOTCHA #6: only sets the signal, NOT searchMode):
        if i+1 < len(args) {
            c.searchMode = true
            c.searchQ = args[i+1]
            i++
        } else {
            c.searchMissingValue = true
        }
  - (2b) FIX the misleading comment (GOTCHA #7). FIND the tail (lines ~291-293):
        // token (no value follows) searchMode stays false and the call falls
        // through to the no-recognized-mode default (exit 1).
    REPLACE with:
        // token (no value follows) searchMode stays false and the call records
        // searchMissingValue so run() exits 2 with "--search requires a query"
        // (mirrors --store, Issue 3; decisions D4).

Task 3: EDIT main.go — the --shell main-switch no-value else + comment fix
  - FILE: main.go (case "--shell": ~line 320-326)
  - (3a) FIND the if-block (no else today):
        if i+1 < len(args) {
            c.completion = true
            c.completionShell = args[i+1]
            i++
        }
    REPLACE with:
        if i+1 < len(args) {
            c.completion = true
            c.completionShell = args[i+1]
            i++
        } else {
            c.shellMissingValue = true
        }
  - (3b) FIX the comment (~line 323-324). FIND:
        // If --shell is the LAST token (no value), completion stays false — mirrors --search's
        // no-value silent behavior (PRD §6.4 specifies no missing-value exit code for --shell).
    REPLACE with:
        // If --shell is the LAST token (no value), records shellMissingValue so run() exits 2
        // with "--shell requires a value (bash|zsh|fish)" (mirrors --store/--search, Issue 3;
        // decisions D4). completion stays false (no value consumed).

Task 4: EDIT main.go — the expandShortBundle -s no-value default (GOTCHA #3)
  - FILE: main.go (expandShortBundle, the `default:` arm of the s-handling switch, ~line 444)
  - FIND:
        default:
            // 's' seen but no value anywhere: mirror the bare "-s"-no-value rule
            // (searchMode stays false). The bool flags before it remain set.
    REPLACE with (add the signal; c is *config):
        default:
            // 's' seen but no value anywhere: mirror the bare "-s"-no-value rule (now
            // records searchMissingValue so run() exits 2 with "--search requires a query",
            // Issue 3). The bool flags before it remain set.
            c.searchMissingValue = true

Task 5: EDIT main.go — the two run() peer checks + the precedent comment (GOTCHA #5)
  - FILE: main.go (run(), the storeMissingValue block ~line 499-502)
  - (5a) FIND:
        if c.storeMissingValue {
            fmt.Fprintln(stderr, "skilldozer: --store requires a value")
            return 2
        }
    REPLACE with (append the two peer checks — exact messages, GOTCHA #4):
        if c.storeMissingValue {
            fmt.Fprintln(stderr, "skilldozer: --store requires a value")
            return 2
        }
        if c.searchMissingValue {
            fmt.Fprintln(stderr, "skilldozer: --search requires a query")
            return 2
        }
        if c.shellMissingValue {
            fmt.Fprintln(stderr, "skilldozer: --shell requires a value (bash|zsh|fish)")
            return 2
        }
  - (5b) UPDATE the storeMissingValue precedent comment (~line 493-498). Append one sentence after the existing
    "it is NOT set by bare `init` …" line:
        // The same missing-value-exit-2 pattern is applied to --search (searchMissingValue) and
        // --shell (shellMissingValue) for symmetry across all value-taking flags (Issue 3, D4).

Task 6: EDIT main_test.go — RENAME + flip TestParseArgsSearchNoValueStaysInactive → TestParseArgsSearchMissingValue
  - FILE: main_test.go (~line 1002-1009)
  - FIND:
        // --search with NO following value (last token) -> searchMode stays false; falls
        // to the default exit-1 path. Proper exit-2 "needs an argument" is P1.M5.T11.
        func TestParseArgsSearchNoValueStaysInactive(t *testing.T) {
            c := parseArgs([]string{"--search"})
            if c.searchMode {
                t.Errorf("parseArgs(--search) with no value: searchMode=true; want false (no value consumed)")
            }
        }
    REPLACE with (rename; assert the new signal + keep searchMode=false as a regression guard — GOTCHA #6):
        // Issue 3: `--search` (last token, no value) records searchMissingValue so run() exits 2 with
        // "--search requires a query" (mirrors --store, D4). searchMode stays false (no value consumed).
        func TestParseArgsSearchMissingValue(t *testing.T) {
            c := parseArgs([]string{"--search"})
            if !c.searchMissingValue {
                t.Errorf("parseArgs(--search) no value: searchMissingValue=false; want true (Issue 3)")
            }
            if c.searchMode {
                t.Errorf("parseArgs(--search) no value: searchMode=true; want false (no value consumed)")
            }
        }

Task 7: EDIT main_test.go — UPDATE TestParseArgsShortBundleSearchNoValue (add the new assertion)
  - FILE: main_test.go (~line 2354-2363)
  - FIND the body asserting version=true + searchMode=false. ADD (after the searchMode assertion):
        if !c.searchMissingValue {
            t.Errorf("-vs: searchMissingValue=false; want true (s had no value -> Issue 3 signal)")
        }
  - Keep the existing version=true + searchMode=false assertions (both still hold). Update the doc comment from
    "searchMode stays false (mirrors the bare -s-no-value rule)" to note it now also records searchMissingValue.

Task 8: EDIT main_test.go — ADD the 3 run-level exit-2 tests (mirror TestRunInitStoreNoValueExits2 @321)
  - FILE: main_test.go (place near the store no-value tests, ~line 370, or grouped with the --search tests).
    package main; var out, errOut bytes.Buffer; run returns int; assert EXACT stderr (GOTCHA #4).
  - ADD:
        // Issue 3 (P1.M2.T1.S1): `--search` (no value) -> exit 2, empty stdout, exact stderr (mirrors --store).
        func TestRunSearchNoValueExits2(t *testing.T) {
            var out, errOut bytes.Buffer
            code := run([]string{"--search"}, &out, &errOut)
            if code != 2 {
                t.Fatalf("run(--search): code=%d; want 2 (missing --search value, Issue 3)", code)
            }
            if out.Len() != 0 {
                t.Errorf("stdout=%q; want EMPTY (§6.4: nothing on stdout on exit-2)", out.String())
            }
            want := "skilldozer: --search requires a query\n"
            if got := errOut.String(); got != want {
                t.Errorf("stderr=%q; want %q", got, want)
            }
        }
        // Issue 3: bare `-s` (no value) -> exit 2 (same path as --search via the main switch).
        func TestRunSearchShortNoValueExits2(t *testing.T) {
            var out, errOut bytes.Buffer
            code := run([]string{"-s"}, &out, &errOut)
            if code != 2 {
                t.Fatalf("run(-s): code=%d; want 2 (missing -s value, Issue 3)", code)
            }
            if out.Len() != 0 {
                t.Errorf("stdout=%q; want EMPTY", out.String())
            }
            want := "skilldozer: --search requires a query\n"
            if got := errOut.String(); got != want {
                t.Errorf("stderr=%q; want %q", got, want)
            }
        }
        // Issue 3: `--shell` (no value) -> exit 2, empty stdout, exact stderr.
        func TestRunShellNoValueExits2(t *testing.T) {
            var out, errOut bytes.Buffer
            code := run([]string{"--shell"}, &out, &errOut)
            if code != 2 {
                t.Fatalf("run(--shell): code=%d; want 2 (missing --shell value, Issue 3)", code)
            }
            if out.Len() != 0 {
                t.Errorf("stdout=%q; want EMPTY", out.String())
            }
            want := "skilldozer: --shell requires a value (bash|zsh|fish)\n"
            if got := errOut.String(); got != want {
                t.Errorf("stderr=%q; want %q", got, want)
            }
        }

Task 9: EDIT main_test.go — ADD TestParseArgsShellMissingValue
  - FILE: main_test.go (place near TestParseArgsSearchMissingValue / the --shell parse tests).
  - ADD:
        // Issue 3: `--shell` (no value) records shellMissingValue so run() exits 2 (mirrors --search/--store).
        // completion stays false (no value consumed).
        func TestParseArgsShellMissingValue(t *testing.T) {
            c := parseArgs([]string{"--shell"})
            if !c.shellMissingValue {
                t.Errorf("parseArgs(--shell) no value: shellMissingValue=false; want true (Issue 3)")
            }
            if c.completion {
                t.Errorf("parseArgs(--shell) no value: completion=true; want false (no value consumed)")
            }
        }

Task 10: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0
  - go test -run 'SearchMissingValue|ShellMissingValue|SearchNoValueExits2|SearchShortNoValueExits2|ShellNoValueExits2|ShortBundleSearchNoValue' -v ./...
  - go test ./...                     # whole module green; the =-form + success tests unchanged
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #9
  - manual: go build -o /tmp/sdz . && for f in --search -s --shell; do /tmp/sdz $f >/dev/null 2>&1; echo "$f exit=$? (want 2)"; done; rm -f /tmp/sdz
```

### Implementation Patterns & Key Details

```go
// The config fields (Task 1) — peers of storeMissingValue:
searchMissingValue bool
shellMissingValue  bool

// The no-value setters (Tasks 2-4) — the `else`/`default` arms mirror storeMissingValue's:
//   main switch --search/-s:  } else { c.searchMissingValue = true }
//   main switch --shell:      } else { c.shellMissingValue = true }
//   bundle -s default:        c.searchMissingValue = true   (c is *config)

// The run() checks (Task 5) — peers of the storeMissingValue block, exact messages:
if c.searchMissingValue {
	fmt.Fprintln(stderr, "skilldozer: --search requires a query")
	return 2
}
if c.shellMissingValue {
	fmt.Fprintln(stderr, "skilldozer: --shell requires a value (bash|zsh|fish)")
	return 2
}
```

Notes easy to get wrong:
- **Do NOT add guards to the `=`-form switch** (`--search=`/`--shell=`) — D5 leaves them unconditional. Only the bare no-token `else`/`default` arms get the fix.
- **Bare `-s` uses the main switch, not the bundle** — `run([]string{"-s"})` is covered by the main-switch `else`, not `expandShortBundle`. The bundle (`-vs`) is a separate setter (Task 4).
- **The `else` sets ONLY the missing-value signal** — `searchMode`/`completion` stay false (no value consumed). Assert both in the parseArgs tests.
- **Exact stderr messages** — assert with `errOut.String() == want`, not `Contains`.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Symmetrical exit 2 vs document the asymmetry? → exit 2 (decisions D4).** The PRD says the symmetrical option is "more predictable for `$(...)` use." Mirroring `storeMissingValue` exactly is the lowest-risk implementation.
2. **Fix the `=`-form too? → NO (decisions D5).** `--search=`/`--shell=` are distinct syntactic forms the PRD doesn't mention; changing them is scope creep. Only the bare no-token case exits 2. `TestParseArgsLongEqualsSearchEmpty` stays green unchanged.
3. **Three setters (main switch `--search`, bundle `-s`, main switch `--shell`) vs fewer? → all three.** Each no-value path must set its signal or that path still falls through to exit 0. The bundle is the third path for `-s`-bearing bundles like `-vs`.
4. **Update `TestParseArgsShortBundleSearchNoValue`? → YES (add the assertion).** It stays green either way (it doesn't check `searchMissingValue` today), but adding the assertion locks the new bundle behavior and keeps the test honest. Not strictly required by the contract OUTPUT, but it's the cross-cutting completeness that prevents a coverage gap.
5. **Run() ordering? → after `storeMissingValue`, before exclusivity (GOTCHA #5).** Peer parse-error guards; `--help`/`--version` still win.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. fmt.Fprintln already imported; bool fields; expandShortBundle takes *config. (GOTCHA #9)

DISPATCH (the run() ladder gains two peer steps):
  before: help → version → unknownFlag → storeMissingValue → exclusivity → init → completion → modes → no-args-usage
  after:  help → version → unknownFlag → storeMissingValue → SEARCH-MISSING → SHELL-MISSING → exclusivity → ...
  The two new checks are inert when the flags are false (the success paths + =-forms).

CONSUMERS:
  - `$(skilldozer --search)` now fails loudly (exit 2, empty stdout) instead of capturing help.
  - `$(skilldozer --shell)` likewise.

PARALLEL SIBLING (no conflict):
  - P1.M1.T2.S1 (Issue 2) edits internal/skillsdir/*. This task edits main.go + main_test.go. Disjoint.

NO ROUTES / NO DATABASE / NO CONFIG SCHEMA / NO COMPLETIONS / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (immediate)

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

go test -run 'SearchMissingValue|ShellMissingValue|SearchNoValueExits2|SearchShortNoValueExits2|ShellNoValueExits2|ShortBundleSearchNoValue' -v ./...
# Expected: ALL pass. Load-bearing assertions:
#   TestParseArgsSearchMissingValue (renamed) -> searchMissingValue=true AND searchMode=false.
#   TestParseArgsShellMissingValue            -> shellMissingValue=true AND completion=false.
#   TestParseArgsShortBundleSearchNoValue     -> version=true, searchMode=false, searchMissingValue=true (updated).
#   TestRunSearchNoValueExits2                -> run(["--search"]) code 2, empty stdout, exact stderr.
#   TestRunSearchShortNoValueExits2           -> run(["-s"]) code 2, empty stdout, exact stderr.
#   TestRunShellNoValueExits2                 -> run(["--shell"]) code 2, empty stdout, exact stderr.

# Regression — the value-present + =-form paths are UNCHANGED (must stay green):
go test -run 'TestParseArgsSearchConsumesOneValue|TestParseArgsLongEqualsSearchEmpty|TestRunSearchMatchByTag|TestParseArgsShortBundleSConsumesRestAsQuery|TestRunCompletion|TestRunInitStore' -v ./...
# Expected: PASS (the success paths + the =-form empty-value test are untouched — D5).
```

### Level 3: Whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions (the =-form + success tests unchanged)

# GOTCHA #9 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Scope invariants:
grep -c 'searchMissingValue\|shellMissingValue' main.go           # expect >= 8 (2 fields + 3 setters + ... + run reads + docs)
grep -c 'requires a query\|requires a value (bash' main.go        # expect 2 (the two run() messages)
git diff --name-only                                               # expect ONLY main.go + main_test.go
```

### Level 4: Behavioral spot-checks (the §6.4 contract, end-to-end)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz .

# (a) The three no-value flags now exit 2 + empty stdout + the exact message:
for f in --search -s --shell; do
  out=$(/tmp/sdz $f 2>/tmp/e); rc=$?
  [ "$rc" = 2 ] && [ -z "$out" ] && echo "OK: $f -> exit 2, empty stdout, stderr=$(cat /tmp/e)" || echo "FAIL: $f rc=$rc out=$out"
done

# (b) CONTROL — value-present paths UNCHANGED (exit 0, real output):
store=$(mktemp -d); mkdir -p "$store/example"; printf -- '---\nname: example\ndescription: d\n---\nx\n' > "$store/example/SKILL.md"
out=$(SKILLDOZER_SKILLS_DIR="$store" /tmp/sdz --search example 2>/dev/null); rc=$?
[ "$rc" = 0 ] && printf '%s' "$out" | grep -q example && echo "OK: --search example still works (exit 0)" || echo "FAIL: rc=$rc out=$out"
out=$(/tmp/sdz --completions --shell bash 2>/dev/null); rc=$?
[ "$rc" = 0 ] && printf '%s' "$out" | grep -q '_skilldozer_completion' && echo "OK: --completions --shell bash still works (exit 0)" || echo "FAIL: rc=$rc"

# (c) CONTROL — the =-form empty value is UNCHANGED (NOT exit 2; runs an empty search -> exit 0 with all skills):
out=$(SKILLDOZER_SKILLS_DIR="$store" /tmp/sdz --search= 2>/dev/null); rc=$?
[ "$rc" = 0 ] && echo "OK: --search= unchanged (exit 0, D5)" || echo "FAIL: --search= rc=$rc (D5 violated)"

# (d) CONTROL — --help still wins over the missing-value guard:
/tmp/sdz --help --search >/dev/null 2>&1; echo "--help --search exit=$? (want 0 — help precedence)"

rm -rf "$store" /tmp/sdz /tmp/e
# Expected: (a) all three "OK"; (b) both "OK"; (c) "OK"; (d) exit 0.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean; `go vet ./...` exit 0; `go build` exit 0
- [ ] Level 2 PASS — the 6 new/renamed/updated tests pass; the value-present + =-form regression tests stay green
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0; `git diff go.mod go.sum` → "deps unchanged"; `git diff --name-only` = ONLY main.go + main_test.go
- [ ] Level 4 PASS — `--search`/`-s`/`--shell` no-value → exit 2 + empty stdout + exact message; value-present + =-form unchanged; `--help` wins

### Feature Validation
- [ ] `config` has `searchMissingValue` + `shellMissingValue` after `storeMissingValue`
- [ ] `--search`/`-s` main-switch + `-s` bundle no-value set `searchMissingValue`; `--shell` main-switch no-value sets `shellMissingValue`
- [ ] `run(["--search"])`, `run(["-s"])`, `run(["--shell"])` → exit 2, empty stdout, exact stderr
- [ ] `run(["--search","foo"])` / `run(["--completions","--shell","bash"])` unchanged (exit 0)
- [ ] `=`-forms (`--search=`, `--shell=`) unchanged (D5)
- [ ] The 3 misleading comments corrected

### Code Quality / Convention Validation
- [ ] Mirrors the `storeMissingValue` pattern exactly (field + setter + run check + message style)
- [ ] Exact-equality stderr assertions (not Contains)
- [ ] Doc comments cite Issue 3 + decisions D4/D5
- [ ] No new deps; go.mod/go.sum byte-for-byte identical

### Scope Discipline
- [ ] Did NOT touch the `=`-form switch (D5 — `--search=`/`--shell=`/`--store=`/`--init=` unchanged)
- [ ] Did NOT touch `internal/skillsdir/*` (P1.M1.T2.S1, parallel, disjoint)
- [ ] Did NOT touch `completions/*` or `README.md`
- [ ] Did NOT change any exit code other than the new missing-value → exit-2 paths
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't add guards to the `=`-form switch.** `--search=`/`--shell=` stay unconditional (D5); only the bare no-token `else`/`default` arms get the fix. `TestParseArgsLongEqualsSearchEmpty` must stay green. (GOTCHA #1.)
- ❌ **Don't forget the `expandShortBundle` -s default.** It's the third setter (for `-vs`/`-ls` last-token bundles). c is `*config`; `c.searchMissingValue = true` is valid. (GOTCHA #3.)
- ❌ **Don't set `searchMode`/`completion` in the no-value `else`.** Only the missing-value signal is set; the mode bools stay false (no value consumed). Assert both in the parseArgs tests. (GOTCHA #6.)
- ❌ **Don't use `Contains` for the run-level stderr.** The messages are fixed verbatim by the contract; assert exact equality (`errOut.String() == want+"\n"`), mirroring the store tests. (GOTCHA #4.)
- ❌ **Don't move the run() checks before help/version/unknownFlag.** They go after `storeMissingValue` (step 3.6/3.7), so `--help --search` still exits 0. (GOTCHA #5.)
- ❌ **Don't just change the comment's "(exit 1)" to "(exit 0)".** The behavior is changing to exit 2 — fix the comment to the exit-2 description. (GOTCHA #7.)
- ❌ **Don't touch `internal/skillsdir/*` or the success paths.** Issue 2 is the sibling (disjoint files); `--search foo` / `--completions --shell bash` are unchanged. (GOTCHA #8/#10.)
- ❌ **Don't add deps or imports.** `fmt.Fprintln` is already imported; bool fields + `*config` are in scope. go.mod/go.sum byte-for-byte identical. (GOTCHA #9.)

---

## Confidence Score

**9.5/10** — The change mirrors a proven, fully-read template (`storeMissingValue`) with every edit pinned to an exact line + before/after (read from the live file and `parseargs_research.md` §1-§4). The design is fixed by decisions D4 (symmetrical exit 2) and D5 (leave `=`-form as-is). The test churn is precisely enumerated (1 rename+flip, 1 update, 4 adds), and grep confirms NO run-level test locks the old exit-0 behavior (so nothing unexpected breaks). The two non-obvious points — the `=`-form staying unchanged (D5) and bare `-s` using the main switch not the bundle — are both documented and traced. The parallel sibling (Issue 2) is in disjoint files. The 0.5 reservation is for the optional `TestParseArgsShortBundleSearchNoValue` update (it stays green without it, but the PRP adds the assertion for coverage honesty — a strict reader could skip it, though doing so leaves a coverage gap).
