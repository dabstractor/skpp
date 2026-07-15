# PRP — P1.M2.T2.S1: Recognize `--` as end-of-options separator in `parseArgs` (Issue 4)

> **Subtask:** P1.M2.T2.S1 — the sole subtask of P1.M2.T2 (Issue 4). Makes a bare `--` token end option parsing in `parseArgs`: every token after `--` (even dashed ones) becomes a positional **skill tag**, so a skill whose tag begins with `-` (e.g. `-foo`) can be addressed via `skilldozer -- -foo`. Today `--` itself is classified as an unknown dashed flag (`skilldozer: unknown flag '--'`, exit 2), making such tags impossible to resolve. The fix is a boolean `endOfOpts` loop flag (decisions.md §D6, the ACCEPTED design) placed **before all token classification**; two guards at the top of the loop body intercept `--` (set the flag) and route post-`--` tokens to `c.tags`. No new config field, no `run()`/exclusivity change.
>
> **Scope:** Two existing files only — `main.go` (1 loop-local var + 2 guards + a doc comment in `parseArgs`) and `main_test.go` (5 new tests). No new files. No `internal/*`, no completions, no config-struct change. Zero new deps (pure stdlib: `==`, a `bool`, `append`). go.mod/go.sum byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** main.go `parseArgs` (184-360) read at the cited line ranges; `default:`/`unknownFlag` path (351/357/359) + the `=`-form check (200) + the short-bundle guard (262) confirmed. `issue_analysis.md` §Issue 4 + `decisions.md` §D6 read in full (the exact insertion is transcribed verbatim). `grep`-confirmed ZERO existing tests assert bare `--` → unknown flag (every `unknownFlag` test uses `--frobnicate`/`-z`/`--bogus`/`--more` — none use `--`), so this is purely additive. The parallel sibling P1.M2.T1.S1 (Issue 3, missing-value) edits the `--search`/`--shell` cases (~288-326) + `expandShortBundle` (~444) + `config` struct + `run()`; this subtask edits the TOP of the loop body (~190) + a var before the loop (~185) — DISJOINT regions, no collision.

---

## Goal

**Feature Goal**: Make `parseArgs` honor the POSIX `--` end-of-options separator (issue_analysis.md §Issue 4): `skilldozer -- <tag>` treats `<tag>` as a literal positional, so a tag beginning with `-` can be resolved. After the fix, `skilldozer -- -x` sets `c.tags=["-x"]` (not `c.unknownFlag="--"`); `skilldozer --list -- --check` sets `list=true` + `tags=["--check"]`; and `skilldozer -- --` makes the second `--` a positional tag named `--` (POSIX-correct once end-of-options is set).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — add `endOfOpts := false` after `var c config` (parseArgs, ~line 185); add two guards as the FIRST checks in the loop body (right after `a := args[i]`, ~line 190, before the `=`-form check at ~200): `if endOfOpts { c.tags = append(c.tags, a); continue }` then `if a == "--" { endOfOpts = true; continue }`; add a doc comment citing the POSIX convention + Issue 4/§D6.
2. `main_test.go` — 5 new tests: `TestParseArgsDashDashSeparator`, `TestParseArgsDashDashBeforeTag`, `TestParseArgsDashDashWithFlags`, `TestRunDashDashUnknownFlagStillWorks`, `TestParseArgsDashDashSecondDashDashIsTag`.

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `parseArgs(["--","-x"])` → `{tags:["-x"], unknownFlag:""}`; `parseArgs(["--list","--","--check"])` → `{list:true, tags:["--check"]}`; `parseArgs(["--","--"])` → `{tags:["--"]}`; `run(["--","--bogus"])` (with a store) → exit 1 unknown-tag (NOT exit 2 unknown-flag), empty stdout; existing unknown-flag tests (`--frobnicate`, `-z`) stay green.

---

## User Persona (if applicable)

**Target User**: A user with a skill whose canonical tag begins with `-` (pathological but possible), and any POSIX-literate user who expects `--` to work as it does in every standard CLI.

**Use Case**: `skilldozer -- -foo` to resolve a skill literally tagged `-foo`; or defensive `skilldozer -- <weird-tag>` to guarantee the tag is taken literally.

**User Journey**: User has `skills/-foo/SKILL.md` → `skilldozer -foo` is parsed as an unknown short-flag bundle (exit 2); `skilldozer -- -foo` → (today) `unknown flag '--'` exit 2 → (after fix) resolves the `-foo` tag.

**Pain Points Addressed**: a class of tags that is literally impossible to address; a missing POSIX convention that surprises shell users.

---

## Why

- **Implements the POSIX `--` convention** (issue_analysis.md §Issue 4). Every standard CLI honors it; skilldozer's absence is a minor-but-real spec gap.
- **Closes a class of unresolvable tags.** A skill at `skills/-foo/SKILL.md` (tag `-foo`) is unreachable today (`skilldozer -foo` → unknown short bundle; `skilldozer -- -foo` → unknown flag `--`). The fix makes it addressable.
- **Cheapest possible support** (issue_analysis.md: "Low priority — such tag names are pathological — but cheap to support"). A 2-line flag + 2 guards; no new config field, no run()/exclusivity/dispatch change.
- **Purely additive, zero breakage.** grep-confirmed no test asserts bare `--` → unknown flag; every existing `unknownFlag` test uses a non-`--` token and stays green.

---

## What

A loop-local boolean + two early `continue` guards at the top of the `parseArgs` loop body, before any token classification. No exit-code change anywhere except the bare-`--` path (was exit 2; now consumed as a separator). The unknown-flag detection for genuinely-unknown flags (`--frobnicate`, `-z`) is UNCHANGED.

### Success Criteria

- [ ] `parseArgs` has `endOfOpts := false` declared after `var c config` (before the loop).
- [ ] The loop body's FIRST two checks (right after `a := args[i]`, before the `=`-form check) are: `if endOfOpts { c.tags = append(c.tags, a); continue }` then `if a == "--" { endOfOpts = true; continue }` — in that ORDER (endOfOpts guard before the `--` guard).
- [ ] A doc comment cites the POSIX `--` convention + Issue 4/§D6 + the `-- -foo` use case.
- [ ] `parseArgs(["--","-x"])` → `tags==["-x"]`, `unknownFlag==""`.
- [ ] `parseArgs(["--","mytag"])` → `tags==["mytag"]`.
- [ ] `parseArgs(["--list","--","--check"])` → `list==true`, `tags==["--check"]`.
- [ ] `parseArgs(["--","--"])` → `tags==["--"]` (second `--` is a positional tag).
- [ ] `run(["--","--bogus"])` with a store → exit 1 (unknown TAG, not unknown flag), empty stdout, stderr names `--bogus`, stderr does NOT contain `unknown flag`.
- [ ] Existing unknown-flag tests (`--frobnicate`, `-z`, `--bogus`) stay GREEN (the fix only adds the bare-`--` path).
- [ ] `go test ./...` green; `go.mod`/`go.sum` unchanged; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact code insertion is transcribed verbatim from `issue_analysis.md` §Issue 4 (line 262) and fixed by `decisions.md` §D6 (the loop-flag approach). Placement anchors are verified-current (parseArgs @184; `var c config` @185; `a := args[i]` @190; `=`-form check @200). The one non-obvious point — the guard ORDER (`endOfOpts` before `a=="--"`) that makes `-- --` treat the second `--` as a positional — is traced case-by-case in `research/verified_facts.md` §4. Zero breakage is grep-proven (no test asserts bare `--` → unknown flag). The 4-case trace (§4) proves the three unchanged classifications (`--list` flag, `-x` unknown, plain tag) are unaffected. The boundary with the parallel Issue-3 sibling is disjoint (this subtask touches the top of the loop body + the var-before-loop; the sibling touches the deep cases + config + run()). An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (root cause, the exact insertion, the 4-case trace, guard order)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M2T2S1/research/verified_facts.md
  why: "§1 the root cause (the 5-step fall to unknownFlag). §2 the exact insertion (verbatim from
        issue_analysis) + decisions §D6 (loop-flag chosen over early-return). §3 verified placement
        anchors (parseArgs @184, var c @185, a:=args[i] @190, =-form @200, default @351). §4 THE 4-case
        trace + the guard-ORDER rule (endOfOpts BEFORE a==\"--\" so the second -- becomes a tag).
        §5 grep PROOF of zero breakage. §6 disjointness from Issue 3. §7 the 5-test plan + templates.
        §8 Mode-A comment only (README is P1.M3.T1.S1). §9 scope (main.go + main_test.go)."
  critical: "§4 (the guard order is load-bearing for the `-- --` edge case) and §2 (place the guards
             BEFORE all classification — the =-form check, the short-bundle check, AND the switch) are
             the two things most likely to be mishandled."

# MUST READ — the authoritative bug writeup + the exact insertion code
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/issue_analysis.md
  why: "§Issue 4 (line 247) is the authoritative repro (-- -x → unknown flag '--', exit 2), root cause
        (the 5-step fall), and the exact Fix Surface (the endOfOpts insertion, transcribed verbatim).
        Its 'Test Impact' lists TestParseArgsDashDashSeparator/BeforeTag/RunDashDashResolvesTag + confirms
        TestParseArgsDashedUnknownNotATag (--frobnicate) still passes."
  section: "Issue 4 (MINOR)."

# MUST READ — the design decision (loop-flag over early-return)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/decisions.md
  why: "§D6 (line 76): loop-flag ACCEPTED (clean, extensible, keeps the loop intact); early-return
        REJECTED (breaks uniform structure, skips post-loop logic). Pins the approach."
  section: "D6."

# MUST READ — the file under edit (read parseArgs' loop head + default branch before editing)
- file: main.go
  why: "THE edit target. parseArgs @184. var c config @185 (PLACE endOfOpts := false right after). the
        loop `for i := 0...` @189. `a := args[i]` @190 (PLACE the 2 guards right after). the =-form check
        @200 (the guards go BEFORE it). short-bundle guard @262. switch a default: @351, HasPrefix(a,\"-\")
        @357, c.unknownFlag=a @359 (the path -- currently falls into)."
  pattern: "parseArgs is an index-based for loop; each iteration classifies args[i] via the =-form check →
            short-bundle check → switch a. Value-taking flags consume the next token via i++. The fix adds
            two EARLY `continue` guards at the very top of the iteration, before any classification."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestParseArgsDashedUnknownNotATag @594 = the
        parseArgs-level template (assert c.tags + c.unknownFlag directly). TestRunTagAtomicityUnknown
        PrintsNothing @667 = the run-level template (sampleStore + SKILLDOZER_SKILLS_DIR + run returns int;
        assert code, empty stdout, stderr Contains the tag). sampleStore(t) + unsetSkillsEnv(t) @28 are
        the helpers; bytes.Buffer / strings.Contains / strings are imported."
  gotcha: "TestRunDashDashUnknownFlagStillWorks needs a STORE (sampleStore) so Find() succeeds and --bogus
           reaches TAG resolution (exit 1 unknown-tag), proving it was NOT classified as an unknown flag
           (exit 2). The load-bearing assertion is stderr does NOT contain 'unknown flag' (distinguishes the
           two paths). Do NOT use unsetSkillsEnv for it (that yields exit 1 'not configured' — still not-2,
           but the sampleStore version proves the full tag-resolution flow per the contract)."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint regions, no collision)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M2T1S1/PRP.md
  why: "Confirms P1.M2.T1.S1 (Issue 3, missing-value) edits the config struct (searchMissingValue/
        shellMissingValue ~169) + the --search/--shell cases (~288-326) + expandShortBundle (~444) + run()
        (~499). It does NOT touch the top of the parseArgs loop body (~190) or the var-before-loop (~185) —
        this subtask's regions. Disjoint; land in either order. Behavioral note (§6): a --search token
        AFTER -- is routed to c.tags by this subtask's guard BEFORE it reaches the --search case — correct
        POSIX semantics; the sibling's missing-value detection only fires for --search NOT after --."

# READ-ONLY — PRD (the authority for the §6.1 unknown-flag rule + POSIX convention)
- file: PRD.md
  why: "READ-ONLY. §6.1 (h3.1) 'Unknown flags ⇒ error + exit 2' — the rule the bare -- currently trips;
        the fix adds the POSIX -- carve-out (a separator, not a flag). The bugfix-2 PRD context's §Issue 4
        + decisions §D6 are the operative authority. Do NOT edit PRD.md."
  section: "h2.3/h3.3 (Issue 4) in the bugfix requirements doc."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/tasks.json
  why: "P1.M2.T2.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes it;
        tasks.json wins on any conflict. NOTE the contract's line cites (202/259) drift vs the live file
        (200/262) — anchor by TEXT/symbol per verified_facts §3."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep)
$ grep -n 'endOfOpts\|a == "--"' main.go main_test.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be changed

```bash
main.go        # ADD: `endOfOpts := false` (after var c config) + 2 guards at top of parseArgs loop body + doc comment
main_test.go   # ADD: 5 tests (3 parseArgs-level + 1 run-level + 1 -- -- edge case)
# go.mod / go.sum — UNCHANGED (pure stdlib: ==, bool, append)
```

| File | Change |
|---|---|
| `main.go` | `parseArgs`: recognize `--` as the POSIX end-of-options separator (loop flag + 2 early-continue guards before all classification). |
| `main_test.go` | Lock the separator behavior: dashed/plain tokens after `--` become tags; `--list` before `--` still a flag; `-- --` → tag `--`; run-level proof a post-`--` dashed token reaches tag resolution (not unknown-flag). |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL) — guard ORDER: `if endOfOpts` MUST come BEFORE `if a == "--"`.
// Once endOfOpts is set, a SECOND `--` (e.g. `skilldozer -- --`) must be appended as a
// positional tag named "--" (POSIX: after end-of-options, everything — including a bare
// `--` — is positional). If you put `if a == "--"` first, the second `--` would re-trigger
// the separator and never become a tag. The §4 trace locks this. (verified_facts §4.)

// GOTCHA #2 — the guards go BEFORE ALL classification: the =-form check (main.go:200),
// the short-bundle check (262), AND switch a (351). A bare "--" must NEVER reach the
// =-form check (no `=`, skipped anyway), the short-bundle check (len==2, skipped), or the
// switch default (HasPrefix("-") true → unknownFlag). Placing the guards after any of these
// reintroduces the bug. They are the FIRST two checks in the loop body, right after `a := args[i]`.

// GOTCHA #3 — `endOfOpts` is a PARSE-LOCAL var, NOT a config field. It does not need to
// survive past parseArgs (it only governs the current parse). Do NOT add it to the `config`
// struct. This is the key difference from the Issue-3 sibling (which adds config fields) —
// this subtask adds no config field at all.

// GOTCHA #4 — anchor placement by TEXT/symbol, not the contract's line numbers (202/259).
// The live lines are: var c config @185, a := args[i] @190, =-form check @200, short-bundle
// guard @262, switch default @351. The contract's numbers drifted; the placement LOGIC
// (guards = first two checks in the loop body) is what matters.

// GOTCHA #5 — NO run() change. After parseArgs, post-`--` tokens are already in c.tags,
// which flow through run()'s existing tag-resolution branch UNCHANGED. Do NOT add any
// endOfOpts handling to run() or exclusivityError — tags are tags. The only observable
// run() difference: `run(["--","--bogus"])` now reaches tag resolution (exit 1 unknown-tag)
// instead of the unknown-flag guard (exit 2).

// GOTCHA #6 — TestRunDashDashUnknownFlagStillWorks needs a STORE (sampleStore) so --bogus
// reaches TAG resolution. With sampleStore (no skill named "--bogus") → exit 1 unknown-tag,
// empty stdout, stderr Contains "--bogus", stderr does NOT Contain "unknown flag". Do NOT
// use unsetSkillsEnv (that gives exit 1 "not configured" — still not-2 but doesn't prove the
// tag-resolution flow). The load-bearing assertion is "stderr does NOT contain 'unknown flag'".

// GOTCHA #7 — the unknown-flag detection for GENUINELY-unknown flags is UNCHANGED. A bare
// `-x` NOT after `--` still → unknownFlag="-x" → exit 2 (endOfOpts is false, it skips both
// guards, reaches switch→default). TestParseArgsDashedUnknownNotATag (--frobnicate) and
// TestRunUnknownFlagExits2 (-z) stay GREEN. The fix is purely additive (only the bare-`--`
// path changes).

// GOTCHA #8 — no merge collision with the parallel sibling P1.M2.T1.S1 (Issue 3). It edits
// the config struct (~169), the --search/--shell cases (~288-326), expandShortBundle (~444),
// and run() (~499) — all DEEP in parseArgs/run. This subtask edits the TOP of the loop body
// (~190) + a var before the loop (~185). No text-level overlap; both are additive insertions
// into different parts of parseArgs. Compose cleanly in either order.

// GOTCHA #9 — no deps/imports change. `==`, `bool`, and `append` are pure stdlib (strings/
// the for-loop are already there). go.mod/go.sum byte-for-byte identical. Verify with
// `git diff --quiet go.mod go.sum`.
```

---

## Implementation Blueprint

### Data models and structure

**None.** No new types, no config-field change. `endOfOpts` is a single parse-local `bool`. (This is the key contrast with the Issue-3 sibling, which adds config fields — GOTCHA #3.)

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add `endOfOpts := false` after `var c config`
  - FILE: main.go (parseArgs, ~line 185, immediately after `var c config`)
  - FIND:
        func parseArgs(args []string) config {
            var c config
            // Index-based loop (not range) ...
  - REPLACE with (insert the var + its doc comment right after `var c config`, before the loop comment):
        func parseArgs(args []string) config {
            var c config
            // POSIX `--` end-of-options separator (Issue 4, decisions §D6): a bare "--" token
            // ends option parsing; every subsequent token (even dashed) is a positional skill
            // tag. This lets a skill whose tag begins with "-" be addressed (`skilldozer -- -foo`).
            // endOfOpts is parse-local (not a config field): it only governs the current parse.
            endOfOpts := false
            // Index-based loop (not range) ...

Task 2: EDIT main.go — add the two guards at the top of the loop body (BEFORE the =-form check)
  - FILE: main.go (parseArgs, immediately after `a := args[i]`, ~line 190, before the `=`-form
    comment block + check at ~192-200; GOTCHA #1 guard order; GOTCHA #2 before all classification)
  - FIND:
        for i := 0; i < len(args); i++ {
            a := args[i]

            // Issue 5 (decisions.md §D5): normalize combined / '='-bearing tokens
  - REPLACE with (insert the two guards between `a := args[i]` and the Issue-5 comment; endOfOpts
    guard FIRST, then the `--` guard):
        for i := 0; i < len(args); i++ {
            a := args[i]

            // POSIX `--`: once end-of-options is set, EVERY token is a positional skill tag.
            // This guard comes BEFORE the `--` separator guard below so a SECOND `--`
            // (`skilldozer -- --`) becomes a positional tag named "--" (POSIX-correct).
            if endOfOpts {
                c.tags = append(c.tags, a)
                continue
            }
            // A bare "--" token is the end-of-options separator: consume it (do NOT add to
            // tags) and set endOfOpts so all later tokens are treated as positionals above.
            if a == "--" {
                endOfOpts = true
                continue
            }

            // Issue 5 (decisions.md §D5): normalize combined / '='-bearing tokens
  - That is the ENTIRE logic fix. No other parseArgs line changes. (GOTCHA #2: these are the
    FIRST two checks in the loop body; the =-form check, short-bundle check, and switch all
    follow unchanged.)

Task 3: EDIT main_test.go — add the 3 parseArgs-level tests (mirror TestParseArgsDashedUnknownNotATag @594)
  - FILE: main_test.go (place near TestParseArgsDashedUnknownNotATag @594 or the other parseArgs tests;
    package main; assert c.tags + c.unknownFlag directly)
  - ADD:
        // Issue 4 (P1.M2.T2.S1): a bare "--" ends option parsing; subsequent tokens (even dashed)
        // become positional skill tags. `skilldozer -- -x` now sets tags=["-x"] (was unknownFlag="--").
        func TestParseArgsDashDashSeparator(t *testing.T) {
            c := parseArgs([]string{"--", "-x"})
            if len(c.tags) != 1 || c.tags[0] != "-x" {
                t.Errorf("tags=%v; want [-x] (-x is a positional after --)", c.tags)
            }
            if c.unknownFlag != "" {
                t.Errorf("unknownFlag=%q; want empty (-- is the separator, not an unknown flag)", c.unknownFlag)
            }
        }
        // Issue 4: a normal tag after "--" is still a positional tag.
        func TestParseArgsDashDashBeforeTag(t *testing.T) {
            c := parseArgs([]string{"--", "mytag"})
            if len(c.tags) != 1 || c.tags[0] != "mytag" {
                t.Errorf("tags=%v; want [mytag]", c.tags)
            }
        }
        // Issue 4: flags BEFORE "--" still parse as flags; tokens AFTER "--" are tags. So
        // `--list -- --check` => list=true, tags=["--check"] (--check is NOT the action here).
        func TestParseArgsDashDashWithFlags(t *testing.T) {
            c := parseArgs([]string{"--list", "--", "--check"})
            if !c.list {
                t.Errorf("list=false; want true (--list before -- is still a flag)")
            }
            if !c.check {
                t.Errorf("check=false; want true? NO — --check is AFTER --, so it must be a TAG, not the action")
            }
            if len(c.tags) != 1 || c.tags[0] != "--check" {
                t.Errorf("tags=%v; want [--check] (--check after -- is a positional tag)", c.tags)
            }
        }
  - NOTE on TestParseArgsDashDashWithFlags: the `if !c.check { ... }` block above is WRONG as
    written — remove it. --check AFTER -- is a TAG (c.check must be FALSE). Correct body:
        func TestParseArgsDashDashWithFlags(t *testing.T) {
            c := parseArgs([]string{"--list", "--", "--check"})
            if !c.list {
                t.Errorf("list=false; want true (--list before -- is still a flag)")
            }
            if c.check {
                t.Errorf("check=true; want false (--check is AFTER --, so it is a TAG, not the action)")
            }
            if len(c.tags) != 1 || c.tags[0] != "--check" {
                t.Errorf("tags=%v; want [--check] (--check after -- is a positional tag)", c.tags)
            }
        }
    (Use THIS corrected body — the §4 trace confirms c.check stays false.)

Task 4: EDIT main_test.go — add the run-level test (mirror TestRunTagAtomicityUnknownPrintsNothing @667)
  - FILE: main_test.go (place near the run tag tests, ~line 685; GOTCHA #6 — needs sampleStore)
  - ADD:
        // Issue 4 (P1.M2.T2.S1): after "--", a dashed token reaches TAG resolution, NOT the
        // unknown-flag path. With a store (no skill named "--bogus") this is exit 1 unknown-TAG
        // (empty stdout, stderr names the tag), proving the -- separator worked. Without the fix
        // this would be exit 2 "unknown flag '--bogus'".
        func TestRunDashDashUnknownFlagStillWorks(t *testing.T) {
            dir := sampleStore(t)
            t.Setenv("SKILLDOZER_SKILLS_DIR", dir)
            var out, errOut bytes.Buffer
            code := run([]string{"--", "--bogus"}, &out, &errOut)
            if code == 2 {
                t.Fatalf("run(-- --bogus): code=2 (treated as unknown flag); want 1 (treated as a tag after --). stderr=%q", errOut.String())
            }
            if code != 1 {
                t.Fatalf("run(-- --bogus): code=%d; want 1 (unknown tag --bogus after --)", code)
            }
            if out.Len() != 0 {
                t.Errorf("stdout=%q; want EMPTY (§6.4: nothing on stdout on failure)", out.String())
            }
            if !strings.Contains(errOut.String(), "--bogus") {
                t.Errorf("stderr=%q; want an error naming the tag '--bogus'", errOut.String())
            }
            if strings.Contains(errOut.String(), "unknown flag") {
                t.Errorf("stderr=%q; must NOT say 'unknown flag' (--bogus is a tag after --)", errOut.String())
            }
        }

Task 5: EDIT main_test.go — add the `-- --` POSIX edge-case test (§4 trace; GOTCHA #1 guard order)
  - FILE: main_test.go (place with the other -- tests)
  - ADD:
        // Issue 4 edge case: once end-of-options is set, a SECOND "--" is a positional tag named
        // "--" (POSIX). This locks the guard ORDER (endOfOpts check before the a=="--" check).
        func TestParseArgsDashDashSecondDashDashIsTag(t *testing.T) {
            c := parseArgs([]string{"--", "--"})
            if len(c.tags) != 1 || c.tags[0] != "--" {
                t.Errorf("tags=%v; want [--] (the second -- is a positional tag once endOfOpts is set)", c.tags)
            }
            if c.unknownFlag != "" {
                t.Errorf("unknownFlag=%q; want empty", c.unknownFlag)
            }
        }

Task 6: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0
  - go test -run 'DashDash' -v ./...  # the 5 new tests pass
  - go test ./...                     # whole module green; the unknown-flag tests (--frobnicate/-z) stay green
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #9
  - manual: go build -o /tmp/sdz . && /tmp/sdz -- -x; echo "exit=$? (NOT 2 'unknown flag --'; it's a tag-resolution result)"
  - manual regression: /tmp/sdz --frobnicate; echo "exit=$? (want 2 — genuinely-unknown flag still detected)"
```

### Implementation Patterns & Key Details

```go
// Task 1 — the parse-local flag (NOT a config field):
func parseArgs(args []string) config {
	var c config
	endOfOpts := false // POSIX -- (Issue 4, §D6)
	for i := 0; i < len(args); i++ {
		a := args[i]

		// Task 2 — the two guards, in this ORDER (GOTCHA #1):
		if endOfOpts {          // everything after -- is a positional tag
			c.tags = append(c.tags, a)
			continue
		}
		if a == "--" {          // the separator itself is consumed
			endOfOpts = true
			continue
		}
		// ... existing =-form (200) / short-bundle (262) / switch (351) unchanged ...
	}
	return c
}
```

Notes easy to get wrong:
- **Guard order:** `if endOfOpts` FIRST, `if a == "--"` SECOND — so a second `--` becomes a tag (GOTCHA #1).
- **Before all classification:** the guards are the first two checks in the loop body, ahead of the `=`-form check (GOTCHA #2).
- **`endOfOpts` is parse-local,** not a config field (GOTCHA #3) — no struct change.
- **No run() change** — post-`--` tokens are already in `c.tags` and flow through unchanged (GOTCHA #5).
- **The run-level test needs `sampleStore`** so the dashed token reaches tag resolution, proving it wasn't an unknown flag (GOTCHA #6).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Loop flag vs early-return? → loop flag (decisions §D6).** ACCEPTED: clean, extensible, keeps the uniform loop structure, doesn't skip post-loop logic. Early-return was rejected.
2. **Guard order: endOfOpts before `a=="--"`? → YES (GOTCHA #1).** Makes the `-- --` edge case POSIX-correct (the second `--` is a positional tag). Reversing them would re-trigger the separator and never produce the tag. A dedicated test (`TestParseArgsDashDashSecondDashDashIsTag`) locks it.
3. **`endOfOpts` as a config field vs parse-local var? → parse-local (GOTCHA #3).** It only governs the current parse and need not survive past `parseArgs`. This is the cleanest design and the key contrast with the Issue-3 sibling (which legitimately adds config fields for cross-function signaling); here no other function needs the flag.
4. **`TestRunDashDashUnknownFlagStillWorks` uses `sampleStore` (a store), not `unsetSkillsEnv`? → sampleStore (GOTCHA #6).** With a store, `--bogus` reaches TAG resolution (exit 1 unknown-tag) — the strongest proof it was not classified as an unknown flag (exit 2). `unsetSkillsEnv` would give exit 1 "not configured" (still not-2, but doesn't prove the tag-resolution flow). The load-bearing assertion is "stderr does NOT contain 'unknown flag'".
5. **Add the `-- --` edge-case test (5th test) beyond the contract's 4? → YES.** The contract OUTPUT explicitly highlights the `-- --` edge ("second -- is a positional tag named '--'"). A dedicated test locks the GOTCHA #1 guard order — without it, a future refactor that swaps the guards would silently break POSIX correctness with no test failure.
6. **No README/help-text change here? → correct (Mode A comment only).** The contract DOCS assigns the README mention to P1.M3.T1.S1's Mode B sweep. This subtask's doc surface is the inline parseArgs comment (Task 1/2).

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. Pure stdlib (==, bool, append). (GOTCHA #9)

DISPATCH (run() UNCHANGED — GOTCHA #5):
  parseArgs now routes post-`--` tokens to c.tags; run()'s existing tag-resolution branch handles
  them. The only observable run() difference: `run(["--","--bogus"])` reaches tag resolution
  (exit 1 unknown-tag) instead of the unknown-flag guard (exit 2).

CONSUMERS:
  - `skilldozer -- -foo` now resolves a skill tagged `-foo` (was: unknown flag '--', exit 2).
  - `skilldozer -- <literal-tag>` guarantees the tag is taken literally (POSIX).

PARALLEL SIBLING (no conflict — GOTCHA #8):
  - P1.M2.T1.S1 (Issue 3) edits config struct (~169) + --search/--shell cases (~288-326) +
    expandShortBundle (~444) + run() (~499). This subtask edits the top of the parseArgs loop
    body (~190) + a var before the loop (~185). Disjoint; both additive; compose in either order.

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

go test -run 'DashDash' -v ./...
# Expected: ALL 5 pass. Load-bearing assertions:
#   TestParseArgsDashDashSeparator          -> tags=["-x"], unknownFlag="".
#   TestParseArgsDashDashBeforeTag          -> tags=["mytag"].
#   TestParseArgsDashDashWithFlags          -> list=true, check=false, tags=["--check"].
#   TestRunDashDashUnknownFlagStillWorks    -> run(["--","--bogus"]) exit 1 (NOT 2), empty stdout,
#                                              stderr Contains "--bogus", stderr NOT "unknown flag".
#   TestParseArgsDashDashSecondDashDashIsTag -> tags=["--"] (the -- -- POSIX edge case; guard order).

# Regression — genuinely-unknown flags are STILL detected (GOTCHA #7; must stay green):
go test -run 'TestParseArgsDashedUnknownNotATag|TestRunUnknownFlagExits2|TestParseArgsUnknownFlagFirstOfTwoWins' -v ./...
# Expected: PASS (the fix only adds the bare-`--` path; --frobnicate/-z/--bogus-before-DD still exit 2).
```

### Level 3: Whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions (additive)

# GOTCHA #9 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Scope invariants:
grep -c 'endOfOpts' main.go            # expect >= 4 (decl + 2 guards + comment refs)
grep -c 'a == "--"' main.go            # expect 1 (the separator guard)
git diff --name-only                    # expect ONLY main.go + main_test.go
```

### Level 4: Behavioral spot-checks (the POSIX contract, end-to-end)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz .

# (a) `-- -x` is now a TAG path, NOT unknown-flag exit 2. With no store it's exit 1 (unconfigured
#     or unknown tag) — the point is it's NOT exit 2 "unknown flag '--'":
/tmp/sdz -- -x >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" != 2 ] && ! grep -q "unknown flag '--'" /tmp/e && echo "OK: -- -x -> exit $rc, not unknown-flag-2" || echo "FAIL: rc=$rc err=$(cat /tmp/e)"

# (b) CONTROL — a genuinely-unknown flag (NOT after --) STILL exits 2 (GOTCHA #7):
/tmp/sdz --frobnicate >/dev/null 2>&1; echo "--frobnicate exit=$? (want 2)"
/tmp/sdz -z            >/dev/null 2>&1; echo "-z exit=$? (want 2)"

# (c) A real tag after -- resolves (build a skill literally tagged with a leading-dash shape is
#     pathological on most filesystems; instead prove -- before a NORMAL tag still resolves):
store=$(mktemp -d); mkdir -p "$store/example"; printf -- '---\nname: example\ndescription: d\n---\nx\n' > "$store/example/SKILL.md"
out=$(SKILLDOZER_SKILLS_DIR="$store" /tmp/sdz -- example 2>/dev/null); rc=$?
[ "$rc" = 0 ] && printf '%s' "$out" | grep -q example && echo "OK: -- example still resolves (exit 0)" || echo "FAIL: rc=$rc out=$out"

# (d) `--list -- --check`: --list is a flag, --check is a tag (exits 1 unknown-tag, NOT 2 unknown-flag):
SKILLDOZER_SKILLS_DIR="$store" /tmp/sdz --list -- --check >/tmp/o 2>/tmp/e; rc=$?
[ "$rc" != 2 ] && ! grep -q "unknown flag" /tmp/e && echo "OK: --list -- --check -> exit $rc (--check is a tag)" || echo "FAIL: rc=$rc err=$(cat /tmp/e)"

rm -rf "$store" /tmp/sdz /tmp/o /tmp/e
# Expected: (a) OK; (b) both exit 2; (c) OK; (d) OK.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean; `go vet ./...` exit 0; `go build` exit 0
- [ ] Level 2 PASS — the 5 new `DashDash` tests pass; the unknown-flag regression tests stay green
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (zero regressions); `git diff go.mod go.sum` → "deps unchanged"; `git diff --name-only` = ONLY main.go + main_test.go
- [ ] Level 4 PASS — `-- -x` is not unknown-flag-2; `--frobnicate`/`-z` still exit 2; `-- example` resolves; `--list -- --check` treats --check as a tag

### Feature Validation
- [ ] `parseArgs(["--","-x"])` → `{tags:["-x"], unknownFlag:""}`
- [ ] `parseArgs(["--","mytag"])` → `{tags:["mytag"]}`
- [ ] `parseArgs(["--list","--","--check"])` → `{list:true, check:false, tags:["--check"]}`
- [ ] `parseArgs(["--","--"])` → `{tags:["--"]}` (second `--` is a positional tag)
- [ ] `run(["--","--bogus"])` (with store) → exit 1 unknown-tag, empty stdout, stderr NOT "unknown flag"
- [ ] Genuinely-unknown flags (`--frobnicate`, `-z`) still exit 2 (unchanged)

### Code Quality / Convention Validation
- [ ] Mirrors parseArgs' established early-`continue` guard style; doc comment cites Issue 4/§D6
- [ ] `endOfOpts` is parse-local (not a config field) — no struct change
- [ ] Guard order (`endOfOpts` before `a=="--"`) locked by `TestParseArgsDashDashSecondDashDashIsTag`
- [ ] No new deps; go.mod/go.sum byte-for-byte identical

### Scope Discipline
- [ ] Did NOT add a config field (`endOfOpts` is parse-local — GOTCHA #3)
- [ ] Did NOT change run() / exclusivityError / dispatch (GOTCHA #5)
- [ ] Did NOT touch the `=`-form check, short-bundle check, or switch (guards go BEFORE them — GOTCHA #2)
- [ ] Did NOT touch completions/*, internal/*, README.md (Mode B = P1.M3.T1.S1)
- [ ] Did NOT modify PRD.md (read-only), tasks.json, prd_snapshot.md, or .gitignore

---

## Anti-Patterns to Avoid

- ❌ **Don't reverse the guard order.** `if endOfOpts` MUST precede `if a == "--"` so a second `--` becomes a positional tag (`-- --` edge case). (GOTCHA #1.)
- ❌ **Don't place the guards after any classification.** They are the FIRST two checks in the loop body — before the `=`-form check (200), the short-bundle check (262), and the switch (351). (GOTCHA #2.)
- ❌ **Don't add `endOfOpts` to the `config` struct.** It is a parse-local `bool`. (GOTCHA #3.)
- ❌ **Don't anchor by the contract's line numbers (202/259).** The live lines are 200/262; place by symbol/text. (GOTCHA #4.)
- ❌ **Don't add run()/exclusivity handling.** Post-`--` tokens are already in `c.tags`; they flow through unchanged. (GOTCHA #5.)
- ❌ **Don't use `unsetSkillsEnv` for the run-level test.** Use `sampleStore` so `--bogus` reaches tag resolution (exit 1 unknown-tag), proving it wasn't an unknown flag. The load-bearing assertion is "stderr does NOT contain 'unknown flag'". (GOTCHA #6.)
- ❌ **Don't widen the fix to other separators or to tags-before-`--`.** Only the bare `--` token is the separator; genuinely-unknown flags still exit 2. (GOTCHA #7.)
- ❌ **Don't add deps/imports or touch the Issue-3 sibling's regions.** Pure stdlib; disjoint from the config-struct/cases/run() edits. (GOTCHA #8/#9.)

---

## Confidence Score

**9.5/10** — The change is a 1-line parse-local `bool` + two early-`continue` guards + a doc comment, with the exact insertion transcribed verbatim from `issue_analysis.md` §Issue 4 and the approach fixed by `decisions.md` §D6 (loop-flag, ACCEPTED). Placement anchors are verified-current; the 4-case trace in `research/verified_facts.md` §4 proves the three unchanged classifications (`--list` flag, `-x` unknown, plain tag) plus the `-- --` POSIX edge. Zero breakage is grep-proven (no test asserts bare `--` → unknown flag). The boundary with the parallel Issue-3 sibling is disjoint (top-of-loop vs deep-cases). The one genuinely non-obvious point — the guard ORDER that makes the second `--` a positional tag — is locked by a dedicated test (`TestParseArgsDashDashSecondDashDashIsTag`). The 0.5 reservation is for the single placement-anchor drift (contract 202/259 vs live 200/262) and the run-level test's reliance on `sampleStore` reaching the unknown-tag path (both mitigated by text/symbol anchoring and the explicit sampleStore instruction).
