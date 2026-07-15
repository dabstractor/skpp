# PRP — P1.M1.T3.S1: Flip parseArgs-level + exclusivity-level tests to `--flag` contract

> **Subtask:** P1.M1.T3.S1 — the parseArgs-level + exclusivity-level half of P1.M1.T3 (update `main_test.go` to the `--check`/`--init`/`--completions` flag contract after S1's `parseArgs` + S2's `exclusivityError` conversion). Sibling **S2** later flips run-level dispatch + completion + help-text tests.
> **Scope boundary:** Edits `main_test.go` ONLY. Flips ~15 parseArgs-level + 14 exclusivity-level test functions that pass bare `check`/`init`/`completion` tokens to pass `--flags` instead; renames 5 `*Subcommand*` parseArgs tests to `*Flag*`; adds 5 new namespace-safety tests; and specially handles the 2 obsolete Issue-4 tests. Does NOT touch help-text tests, run-level dispatch tests, the unconfigured test, or any `.go` source (those are S2 / already done).
> **Two CRITICAL correctness traps** (proven by probe — see GOTCHA #1/#2): the item-description's flip table has **two WRONG entries** for the Issue-4-derived `init init` tests. `--init` OWNS the next positional (§6.3), so `--init init`→initStore (not a tag) and `--init sometag`→initStore (not a tag, no conflict). Naive flips FAIL. The PRP gives the verified rewrites.

---

## Goal

**Feature Goal**: Make `main_test.go`'s parseArgs-level and exclusivity-level tests assert the `--check`/`--init`/`--completions` flag contract that S1 (`parseArgs`, committed `594be07`) and S2 (`exclusivityError`, committed `1e2fe53`) already implement, plus add 5 new namespace-safety tests proving bare `check`/`init`/`completions` are now skill tags.

**Deliverable**: Edits to `main_test.go` only (no new files, no source changes):
1. Flip 13 in-range parseArgs-level tests (lines 1224–1457) + 2 out-of-range green init-tests (321, 373) = **15** (matches contract "~15"): every `"check"`→`"--check"`, `"init"`→`"--init"`, `"completion"`→`"--completions"` token.
2. Rename 5 parseArgs tests from `*Subcommand*` → `*Flag*` (GOTCHA #4).
3. Specially handle the 2 Issue-4 tests: `TestRunExclusivityInitInit` (2006) and `TestParseArgsInitInitCapturedAsTag` (1457) — their naive flips FAIL (GOTCHA #1/#2); use the verified rewrites.
4. Flip 14 exclusivity-level tests (lines 1867–2099) that pass bare tokens; tighten their loose `Contains(...)` message assertions to the `--flag` form (GOTCHA #3).
5. Add 5 new namespace-safety tests (`TestParseArgsBareCheckNowTag`, `BareInitNowTag`, `BareCompletionsNowTag`, `InitFlagWithDir`, `InitEqualsDir`).

**Success Definition**: `go build ./...` + `go vet ./...` pass; `gofmt -l main_test.go` clean; the targeted test selector (parseArgs-level + exclusivity-level + new tests) is **GREEN**; no source file (`main.go`, `internal/*`, `go.mod`) is modified. Other reds (help-text, dispatch, unconfigured) are EXPECTED until S2 (GOTCHA #5).

---

## User Persona (if applicable)

Not applicable at runtime — this is test maintenance. The end-user-visible contract it locks in: `skilldozer --check` runs validation while `skilldozer check` resolves a skill literally tagged `check` (PRD §6.3 / decision 19 — "no bare-word subcommands; the entire positional namespace is reserved for skill tags").

---

## Why

- **PRD §6 + decision 19**: S1/S2 already drove `check`/`init`/`completion` → `--check`/`--init`/`--completions` in the source. The tests still pass the OLD bare tokens and are RED. T3.S1 makes the parseArgs-level + exclusivity-level tests assert the new contract so `go test` reflects reality.
- **Namespace-safety guarantee (§6.3/§14)**: 5 NEW tests prove bare `check`/`init`/`completions` land in `c.tags` (not mode flags) — the foundation for the skills-first completions rewrite (P1.M2.T1). Without these, the namespace-safety contract is untested.
- **Closing the test half of Change Group (test_doc_change_map.md)**: the parseArgs/exclusivity slice is this subtask's bounded scope; help-text/dispatch/unconfigured slices are S2.

---

## What

A mechanical token-flip across two test regions, PLUS two non-mechanical Issue-4 rewrites, PLUS 5 new tests. The flip is `check`→`--check`, `init`→`--init`, `completion`→`--completions` (PLURAL — GOTCHA #6) in every `parseArgs([]string{...})` / `run([]string{...})` call of the in-scope functions. Message assertions get tightened from bare-word `Contains` to flag-form `Contains` (GOTCHA #3).

The two Issue-4 tests need rewrites because `--init` captures its following positional as `initStore` (§6.3), so the old "duplicate init becomes a tag" logic no longer exists:

| Test | Naive flip | Problem | Verified rewrite |
|---|---|---|---|
| `TestRunExclusivityInitInit` (2006) | `["--init","--init"]` | idempotent, NO tags, NO conflict → NOT exit 2 | `["--init","store1","straytag"]` → init+tags → exit 2 |
| `TestParseArgsInitInitCapturedAsTag` (1457) | `["--init","init"]` | initStore="init", NOT a tag | rewrite to assert initStore="init", tags empty; OR delete |

### Success Criteria

- [ ] All 13 in-range parseArgs tests + 2 out-of-range (321/373) pass `--flag` tokens (bare `check`/`init`/`completion` gone from their args)
- [ ] 5 parseArgs tests renamed `*Subcommand*`→`*Flag*` (1224/1235/1262/1389/1404/1418 — see GOTCHA #4 for the exact set)
- [ ] `TestRunExclusivityInitInit` (2006) rewritten to `["--init","store1","straytag"]`, exit 2, config-not-written assertion preserved
- [ ] `TestParseArgsInitInitCapturedAsTag` (1457) rewritten (initStore="init", no tags) or deleted
- [ ] All 14 exclusivity tests flip tokens + tighten `Contains` assertions to `--flag` form
- [ ] 5 new namespace-safety tests added and GREEN
- [ ] `go build ./...` + `go vet ./...` pass; `gofmt -l main_test.go` clean; targeted test selector GREEN
- [ ] `main.go`, `internal/*`, `go.mod`, `go.sum` UNCHANGED (test-only); help-text/dispatch/unconfigured tests UNTOUCHED (S2)

---

## All Needed Context

### Context Completeness Check

**Pass.** Every flip is pinned to a verified live line number with exact old→new args in `research/verified_facts.md` §2; the two WRONG flip-table entries (§1) are proven by probe with the corrected rewrites; the message-assertion tightening (§3) and scope boundary (§4) are documented; the 5 new tests have verified expected behavior (§5). An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — THE source of truth (verified line numbers + exact old/new args + the two trap rewrites)
- file: plan/004_5851dcff4371/P1M1T3S1/research/verified_facts.md
  why: "THE inventory. §0 = current repo state (S1+S2 committed; T2.S1 uncommitted-but-present). §1 = the two
        WRONG flip-table entries (InitInit run + InitInitCapturedAsTag parseArgs) with probe proof + verified
        rewrites. §2 = complete flip inventory (parseArgs 15, exclusivity 14, new 5) with exact args. §3 =
        message-assertion tightening. §4 = scope boundary (help-text/dispatch/unconfigured = S2; do NOT touch).
        §5 = new-test specs with verified behavior."
  critical: "§1 (the two traps) is the single most important section — naive flips there produce FAILING tests.
             §4 prevents scope creep into S2 and into already-done source files."

# MUST READ — the change map (its line numbers ARE accurate for main_test.go; but TWO of its flip entries are wrong)
- file: plan/004_5851dcff4371/architecture/test_doc_change_map.md
  why: "The detailed per-function flip tables. Its line numbers match the live main_test.go EXACTLY (verified).
        BUT its entries for TestRunExclusivityInitInit (2006) and TestParseArgsInitInitCapturedAsTag (1457) are
        WRONG (verified_facts §1) — follow the rewrites there, not the map's naive flips. Also: it lists help-text,
        dispatch, and unconfigured tests — those are S2's scope, NOT this task's (verified_facts §4)."
  section: "parseArgs-level tests to flip; Exclusivity tests to flip; NEW tests to add. (SKIP: Dispatch, Help text,
            Unconfigured — those are S2.)"

# MUST READ — what parseArgs/exclusivityError actually do now (already implemented)
- file: main.go
  why: "READ-ONLY here. parseArgs --init case @327-339 (OWNS next positional — the root cause of the two traps);
        --check @296, --completions @299. exclusivityError @753-818 emits single-quoted flag messages
        ('--check'/'--init'/'--completions' cannot be combined ...). These define what the flipped tests must assert."
  pattern: "exclusivityError message family @789/792/801/804/812/815 — the new Contains assertions target these."

# READ-ONLY — the parallel sibling (defines the T2.S1 boundary this task assumes)
- file: plan/004_5851dcff4371/P1M1T2S1/PRP.md
  why: "T2.S1 owns usageText + error-prefix strings + ErrNotFound + doc comments. Its working-tree changes are
        ALREADY PRESENT (verified_facts §0) — which is why help-text tests are red now. This task assumes T2.S1's
        output (usageText in --flag form) but does NOT touch it. The contract INPUT says 'P1.M1.T1 + P1.M1.T2'."

# READ-ONLY — the PRD authority
- file: PRD.md
  why: "§6.3 ('no bare-word subcommands; the entire positional namespace is reserved for skill tags; --init is the
        sole mode accepting a positional <dir>'), §6.1 (--check/--init/--completions flags), decision 19
        (completion→--completions PLURAL). Do NOT edit PRD.md."
  section: "h2.5 (§6), h3.1 (§6.1), h3.4 (§6.4), h2.18 (decision 19)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD && git status --short
1e2fe53                                  # S2 (exclusivityError) committed; parent 594be07 = S1 (parseArgs)
M  internal/skillsdir/skillsdir.go        # T2.S1 working-tree (uncommitted) — usageText/ErrNotFound already flipped
M  main.go                                # T2.S1 working-tree (uncommitted) — present but NOT this task's concern
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK
BUILD_OK / VET_OK                         # green; go test RED (parseArgs/exclusivity/help/dispatch — T3.S1 + T3.S2)
# main_test.go: 3089 lines. parseArgs-level @1224-1472; exclusivity-level @1795-2115 (19 funcs, 14 need flip);
#               TestRunInitStoreNoValue* @321/373 (green, flip for consistency); new tests slot in @~1470.
```

### Desired Codebase tree with files to be changed

```bash
main_test.go    # MODIFY — parseArgs flips (15) + renames (5) + exclusivity flips (14) + 2 Issue-4 rewrites + 5 new tests
# main.go / internal/* / go.mod / go.sum — UNCHANGED (test-only; source already done by S1+S2+T2.S1)
```

**File responsibilities:**
| Region | Change | Count |
|---|---|---|
| parseArgs-level (1224–1457) | token flip + rename `*Subcommand*`→`*Flag*` | 13 (+1 special rewrite) |
| parseArgs-level (321, 373) | token flip (consistency) | 2 |
| exclusivity-level (1867–2099) | token flip + tighten `Contains` | 14 (+1 special rewrite) |
| new namespace-safety tests | ADD | 5 |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — CRITICAL: --init OWNS the next positional (PRD §6.3). A following NON-dashed token becomes
// c.initStore, NEVER a tag. The test_doc_change_map / item-description flip tables have TWO WRONG entries
// because they assume the old Issue-4 "duplicate init → tag" behavior still holds. It does NOT (S1 deleted it).
// PROVEN BY PROBE (research/verified_facts.md §1):
//   ["--init","--init"]   => init=true, initStore="",  tags=[]        | exclusivity bad=FALSE (idempotent)
//   ["--init","sometag"]  => init=true, initStore="sometag", tags=[]   | exclusivity bad=FALSE (sometag is the STORE)
//   ["--init","init"]     => init=true, initStore="init",    tags=[]   | exclusivity bad=FALSE (init is the STORE)
//   ["--init","store1","straytag"] => init=true, initStore="store1", tags=[straytag] | exclusivity bad=TRUE ✓
// → TestRunExclusivityInitInit (2006): do NOT use ["--init","--init"] or ["--init","sometag"]; use
//   ["--init","store1","straytag"] (two positionals → init+tags conflict → exit 2, config not written).

// GOTCHA #2 — CRITICAL: TestParseArgsInitInitCapturedAsTag (1457) tested Issue-4 (duplicate bare init → tag).
// That behavior is GONE. ["--init","init"] → initStore="init", tags=[] (the flip-table note "(2nd is now a tag)"
// is WRONG). REWRITE it to assert the real behavior (initStore="init", no tags) and rename to
// TestParseArgsInitFlagLiteralInitStore, OR DELETE it (the new TestParseArgsInitFlagWithDir covers --init <dir>).

// GOTCHA #3 — exclusivity message assertions are LOOSE Contains("check"/"init"/"completion") that PASS via
// substring for BOTH old and new messages (the new flag messages contain those words). The RED cause is the
// bare INPUT token (no longer sets the mode flag → wrong exit code). So: token flip = ESSENTIAL fix;
// tightening Contains("X")→Contains("--X") = contract LOGIC(c) directive (assert the flag form). Both needed.

// GOTCHA #4 — RENAME only tests whose name literally contains "Subcommand": TestParseArgsCheckSubcommand (1224)
// → TestParseArgsCheckFlag; TestParseArgsInitSubcommand (1262) → TestParseArgsInitFlag;
// TestParseArgsCompletionSubcommand (1389) → TestParseArgsCompletionsFlag. The *AfterFlag / *AndTag* /
// *PositionalDir* / *StoreLongForm* / *ShellLongForm* / *ShellEqualsForm* / *DirNotCapturedAsTag* names are
// FINE (rename the two Completion*Shell* ones to Completions*Shell* for the plural, optional but consistent).
// Do NOT rename exclusivity tests (no "Subcommand" in their names).

// GOTCHA #5 — SCOPE: do NOT touch these (other tasks own them):
//   - Help-text tests: TestRunHelpShowsInitRow (2029), TestRunHelpShowsCompletionRow (2116) → P1.M1.T3.S2.
//     (They're RED now because T2.S1's usageText rewrite is in the working tree; flipping them is S2's job.)
//   - Dispatch tests: TestRunCheck* (1474-1637), TestRunInitStoreWritesConfig* (2760), TestRunInitStoreTildeExpandsHome
//     (2820), TestRunCompletion* (2953-3037) → P1.M1.T3.S2.
//   - Unconfigured test: TestRunBareTagUnconfiguredNeverPrompts (2867) → P1.M1.T3.S2.
//   - Any .go source (main.go, internal/skillsdir/*) — already done by S1+S2+T2.S1. T3.S1 = main_test.go ONLY.
//   - The 5 non-bare exclusivity tests (TagsAndList 1795, TagsAndSearch 1810, TagsAndAll 1822, TagsAndPath 1837,
//     PathAndTag 1852) and TestExclusivityErrorListingModes (2350) and the Issue-6 listing-mode tests (2390-2454)
//     need NO change (no bare check/init/completion tokens).

// GOTCHA #6 — --completions is PLURAL (decision 19). The old bare subcommand was singular "completion"; the new
// flag is "--completions". Every flip of a completion token → "--completions" (never "--completion"). The bare-tag
// namespace-safety test uses "completions" (plural bare; a bare "completion" also becomes a tag).

// GOTCHA #7 — TestRunInitStoreNoValueExits2 (321) and TestRunInitStoreNoValueDoesNotWriteConfig (373) are
// currently GREEN (the bare "init"→tag change doesn't alter their outcome: --store last-token sets
// storeMissingValue → exit 2 fires before init+tags exclusivity). The flip (init→--init) is CONSISTENCY-only.
// They're the contract's "~15" out-of-range pair. S2 may also touch them; the edit is idempotent (no real conflict).

// GOTCHA #8 — Place the 5 new tests in the parseArgs section, AFTER TestParseArgsInitDirNotCapturedAsTag / the
// rewritten InitInit test (around line 1470), BEFORE the `// --- run: skilldozer check` divider. Use the
// standalone one-function-per-case style (NOT table-driven) — matches the existing parseArgs-capture tests.

// GOTCHA #9 — No deps/import change. main_test.go already imports bytes/strings/os/filepath/testing. Adding
// tests changes nothing in go.mod/go.sum. The gate is gofmt + go vet + go build + targeted go test.
```

---

## Implementation Blueprint

### Data models and structure

**None.** No structs, no signatures, no imports change. This is a test-only edit. `config` fields (`c.check`/`c.init`/`c.completion`/`c.initStore`/`c.tags`) are unchanged.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: FLIP the 13 in-range parseArgs-level tests (main_test.go:1224-1472)
  - For EACH function in research/verified_facts.md §2 "IN-range" table: change the bare token(s) in the
    parseArgs([]string{...}) call. Exact old→new:
      1224 TestParseArgsCheckSubcommand:        ["check"]                      → ["--check"]
      1235 TestParseArgsCheckAfterFlag:         ["--no-color","check"]         → ["--no-color","--check"]
      1248 TestParseArgsCheckAndTagBothCaptured:["check","sometag"]            → ["--check","sometag"]
      1262 TestParseArgsInitSubcommand:         ["init"]                       → ["--init"]
      1279 TestParseArgsInitPositionalDir:      ["init","/tmp/x"]              → ["--init","/tmp/x"]
      1293 TestParseArgsInitStoreLongForm:      ["init","--store","/tmp/x"]    → ["--init","--store","/tmp/x"]
      1310 TestParseArgsInitStoreEqualsForm:    ["init","--store=/tmp/x"]      → ["--init","--store=/tmp/x"]
      1343 ...NoValueSetsSignal:                ["init","--store"]             → ["--init","--store"]
      1389 TestParseArgsCompletionSubcommand:   ["completion"]                 → ["--completions"]
      1404 ...CompletionShellLongForm:          ["completion","--shell","bash"]→ ["--completions","--shell","bash"]
      1418 ...CompletionShellEqualsForm:        ["completion","--shell=bash"]  → ["--completions","--shell=bash"]
      1445 TestParseArgsInitDirNotCapturedAsTag:["init","/tmp/x"]              → ["--init","/tmp/x"]
      1457 TestParseArgsInitInitCapturedAsTag:  SPECIAL — see Task 4
  - These tests' ASSERTIONS (c.check==true, c.tags empty, etc.) already match the flag behavior, so no assertion
    edits beyond the token flip. (The flipped tests now PASS because the flags set the fields.)
  - GOTCHA #6: --completions plural. GOTCHA #1 does NOT apply to these 12 (none are the init+init duplicate).

Task 2: RENAME 5 parseArgs tests Subcommand→Flag (GOTCHA #4)
  - 1224: TestParseArgsCheckSubcommand         → TestParseArgsCheckFlag
  - 1262: TestParseArgsInitSubcommand          → TestParseArgsInitFlag
  - 1389: TestParseArgsCompletionSubcommand    → TestParseArgsCompletionsFlag
  - 1404: TestParseArgsCompletionShellLongForm → TestParseArgsCompletionsShellLongForm   (plural, optional-consistent)
  - 1418: TestParseArgsCompletionShellEqualsForm → TestParseArgsCompletionsShellEqualsForm (plural, optional-consistent)
  - Also update the 3 function doc comments that say "RESERVED subcommand" / "subcommand" → "flag":
      1224-ish comment "The bare token 'check' selects the check subcommand" → "'--check' selects check mode"
      1262-ish comment "'init' alone is a RESERVED subcommand" → "'--init' sets the init mode"
      1389-ish comment "'completion' alone is a RESERVED subcommand" → "'--completions' sets completion mode"
  - (Use your editor's rename; a Go test function rename is a plain text change — no callers to update.)

Task 3: FLIP + TIGHTEN the 14 exclusivity-level tests (main_test.go:1867-2099)
  - For EACH function in research/verified_facts.md §2 "exclusivity" table: (a) flip the bare token(s) in
    run([]string{...}); (b) tighten the Contains(...) message assertion per §3. Exact:
      1867 CheckAndTags:        ["check","foo"]        → ["--check","foo"];        Contains("check")→Contains("--check")
      1882 CheckAndList:        ["check","--list"]     → ["--check","--list"]
      1897 CheckAndPath:        ["check","--path"]     → ["--check","--path"];     (already checks --path; add/tighten --check)
      1917 InitAndList:         ["init","--list"]      → ["--init","--list"];      Contains("init")→Contains("--init")
      1932 InitAndPath:         ["init","--path"]      → ["--init","--path"]
      1948 InitAndCheck:        ["init","check"]       → ["--init","--check"]
      1963 InitAndSearch:       ["init","--search","q"]→ ["--init","--search","q"]
      1975 InitAndAll:          ["init","--all"]       → ["--init","--all"]
      1988 InitAndStrayTag:     ["init","foo","bar"]   → ["--init","foo","bar"];   (Contains("tag") stays)
      2006 InitInit:            SPECIAL — see Task 5
      2050 CompletionAndTag:    ["completion","example"]→ ["--completions","example"]; Contains("completion")→Contains("--completions")
      2065 CompletionAndList:   ["completion","--list"] → ["--completions","--list"]
      2081 CheckAndCompletion:  ["check","completion"]  → ["--check","--completions"]
      2099 InitAndCompletion:   ["init","completion"]   → ["--init","--completions"]
  - GOTCHA #3: the exit-code + empty-stdout assertions need NO change (the flip makes them pass). Only the token
    + the one Contains assertion change. For tests with NO Contains("check"/"init"/"completion") assertion
    (e.g. 1882/1963/1975 assert only exit code + empty stdout), just flip the token.
  - GOTCHA: also update each test's doc comment if it says "check"/"init"/"completion" as a bare token in prose
    (e.g. "check + tag" → "--check + tag"). Light prose touch for honesty; not all comments need it.

Task 4: REWRITE TestParseArgsInitInitCapturedAsTag (main_test.go:1457) — GOTCHA #2
  - The Issue-4 behavior is GONE. ["--init","init"] → initStore="init", tags=[] (NOT a tag).
  - PRIMARY (rewrite + rename): make it assert the real new behavior and rename:
      func TestParseArgsInitFlagLiteralInitStore(t *testing.T) {
          // Namespace safety (decision 19 / §6.3): --init owns its following positional as the store dir, so a
          // store literally named "init" is accepted (--init init ⇒ initStore="init", NOT a tag, NOT special-cased).
          c := parseArgs([]string{"--init", "init"})
          if !c.init { t.Errorf("init not set: %+v", c) }
          if c.initStore != "init" { t.Errorf("initStore=%q; want \"init\"", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty", c.tags) }
      }
  - ALTERNATIVE (delete): remove the function + its preceding `// Issue 4 ...` comment block entirely. (The new
    TestParseArgsInitFlagWithDir covers --init <dir>.) Pick ONE; rewrite is preferred (keeps a guard at that spot).

Task 5: REWRITE TestRunExclusivityInitInit (main_test.go:2006) — GOTCHA #1
  - OLD ["init","init"] exited 2 via Issue-4 (duplicate init → tag → init+tags). GONE.
  - ["--init","--init"] is idempotent → NO conflict → NOT exit 2 (the exit-2 + config-not-written assertions FAIL).
  - REWRITE to trigger init+tags via a SECOND positional (the only way --init + a tag conflict can arise):
      func TestRunExclusivityInitInitStrayTagNoConfigWrite(t *testing.T) {   // (or keep TestRunExclusivityInitInit)
          // Decision 19 / §6.3: --init owns ONE positional (the store). A SECOND positional is a stray tag →
          // init+tags conflict → exit 2, and exclusivity fires BEFORE init dispatch so the config is NOT written.
          // (Supersedes the old Issue-4 `init init` regression, which had no flag-world equivalent.)
          cfg := filepath.Join(t.TempDir(), "must-not-exist.yaml")
          t.Setenv("SKILLDOZER_CONFIG", cfg)
          var out, errOut bytes.Buffer
          code := run([]string{"--init", "store1", "straytag"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(--init store1 straytag): code=%d; want 2 (init + stray tag)", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "--init") { t.Errorf("stderr=%q; want a message mentioning --init", errOut.String()) }
          if _, err := os.Stat(cfg); !os.IsNotExist(err) {
              t.Errorf("config %s was written; exclusivity must fire before init dispatch (got err=%v)", cfg, err)
          }
      }
  - Keep the name TestRunExclusivityInitInit OR rename to TestRunExclusivityInitStrayTagNoConfigWrite for accuracy.
    (Distinct from TestRunExclusivityInitAndStrayTag by its config-not-written assertion; use distinct tokens store1/straytag.)

Task 6: FLIP the 2 out-of-range green init tests (main_test.go:321, 373) — GOTCHA #7
  - 321  TestRunInitStoreNoValueExits2:             ["init","--store"]  → ["--init","--store"]
  - 373  TestRunInitStoreNoValueDoesNotWriteConfig: ["init","--store"]  → ["--init","--store"]
  - These are currently GREEN; the flip is consistency-only (no assertion change). Leave their doc comments as-is
    unless they say bare "init" in the run() arg description.

Task 7: ADD 5 new namespace-safety tests (main_test.go, slot ~line 1470) — GOTCHA #8
  - Place AFTER the rewritten TestParseArgsInitFlagLiteralInitStore (1457) and BEFORE the
    `// --- run: skilldozer check (P1.M4.T10.S1) ---` divider (~1473). Standalone one-per-case style. Verified behavior:
      func TestParseArgsBareCheckNowTag(t *testing.T) {
          // Namespace safety (decision 19 / §6.3): a bare "check" is a skill TAG, never the check mode.
          c := parseArgs([]string{"check"})
          if c.check { t.Errorf("bare check: check=true; want false (it is a tag)") }
          if len(c.tags) != 1 || c.tags[0] != "check" { t.Errorf("bare check: tags=%v; want [check]", c.tags) }
      }
      func TestParseArgsBareInitNowTag(t *testing.T) {
          c := parseArgs([]string{"init"})
          if c.init { t.Errorf("bare init: init=true; want false (it is a tag)") }
          if len(c.tags) != 1 || c.tags[0] != "init" { t.Errorf("bare init: tags=%v; want [init]", c.tags) }
      }
      func TestParseArgsBareCompletionsNowTag(t *testing.T) {
          c := parseArgs([]string{"completions"})
          if c.completion { t.Errorf("bare completions: completion=true; want false (it is a tag)") }
          if len(c.tags) != 1 || c.tags[0] != "completions" { t.Errorf("bare completions: tags=%v; want [completions]", c.tags) }
      }
      func TestParseArgsInitFlagWithDir(t *testing.T) {
          // --init owns its following positional as the store (§6.3): not captured as a tag.
          c := parseArgs([]string{"--init", "/tmp/x"})
          if !c.init { t.Errorf("--init /tmp/x: init=false; want true") }
          if c.initStore != "/tmp/x" { t.Errorf("--init /tmp/x: initStore=%q; want /tmp/x", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("--init /tmp/x: tags=%v; want empty", c.tags) }
      }
      func TestParseArgsInitEqualsDir(t *testing.T) {
          // --init=<dir> '='-form: sets init + initStore (mirrors --store=).
          c := parseArgs([]string{"--init=/tmp/x"})
          if !c.init { t.Errorf("--init=/tmp/x: init=false; want true") }
          if c.initStore != "/tmp/x" { t.Errorf("--init=/tmp/x: initStore=%q; want /tmp/x", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("--init=/tmp/x: tags=%v; want empty", c.tags) }
      }

Task 8: VERIFY (the gate — GOTCHA #9)
  - COMMAND: gofmt -l main_test.go                       (must print NOTHING)
  - COMMAND: go vet ./...                                (exit 0)
  - COMMAND: go build ./...                              (exit 0)
  - COMMAND: git diff --stat                            (expect ONLY main_test.go changed)
  - COMMAND: go test -run '<SELECTOR>' -v ./...          (the parseArgs + exclusivity + new families — see Validation)
  - INVARIANT: grep -n '"check"\|"init"\|"completion"' on the in-scope test bodies → bare tokens gone (except the
               NEW bare-tag tests and any bare "foo"/"sometag"/"example" tag tokens, which are correct).
```

### Implementation Patterns & Key Details

```go
// The mechanical flip is a one-token swap inside the existing []string{...} literal. Example (1224):
//   BEFORE:  c := parseArgs([]string{"check"})
//   AFTER:   c := parseArgs([]string{"--check"})
// (the assertions c.check==true / len(c.tags)==0 already hold for the flag — no assertion edit needed.)

// The exclusivity flip + tighten (1867):
//   BEFORE:  code := run([]string{"check", "foo"}, &out, &errOut)
//            ...
//            if !strings.Contains(errOut.String(), "check") { ... }
//   AFTER:   code := run([]string{"--check", "foo"}, &out, &errOut)
//            ...
//            if !strings.Contains(errOut.String(), "--check") { ... }
// (Contains("--check") is MORE precise and still matches "'--check' cannot be combined ...".)

// The two NON-mechanical rewrites are the hazard. They are NOT token swaps:
//   TestRunExclusivityInitInit (2006):   run([]string{"--init", "store1", "straytag"})  // 3 tokens, exit 2
//   TestParseArgsInitInitCapturedAsTag:  parseArgs([]string{"--init", "init"})         // assert initStore=="init"
// Do NOT follow test_doc_change_map's naive ["--init","--init"] / ["--init","init"] "(2nd is a tag)" there.
```

Notes easy to get wrong:
- The `*Subcommand*`→`*Flag*` rename changes the Go test function name; `go test -run` matches by name, so a stale name means the test silently doesn't run. Rename consistently (function + any doc-comment reference).
- `TestRunExclusivityInitInit` rewritten to `--init store1 straytag` is functionally adjacent to `TestRunExclusivityInitAndStrayTag` (`--init foo bar`); keep both (the former asserts config-not-written), but use distinct tokens to avoid an exact duplicate.
- The 2 out-of-range init tests (321/373) are GREEN — do not be alarmed if they were already passing; the flip is for consistency and does not change their outcome.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **TestRunExclusivityInitInit: `--init --init` vs `--init sometag` vs `--init store1 straytag`? → the third.** Proven by probe (§1A): the first two produce NO exclusivity error (idempotent / sometag→initStore). Only a SECOND positional lands in `c.tags`, triggering init+tags. This preserves the test's exit-2 + config-not-written invariant. The item-description's "alternatively `--init sometag`" is factually wrong.
2. **TestParseArgsInitInitCapturedAsTag: rewrite vs delete? → rewrite (primary), delete (alternative).** Issue-4 is gone; `--init init`→initStore="init". Rewrite keeps a namespace-safety guard (store literally named "init"); delete is acceptable since TestParseArgsInitFlagWithDir covers `--init <dir>`.
3. **Tighten Contains("check")→Contains("--check")? → YES.** Contract LOGIC(c) says "update ALL error message assertions." The loose assertions already pass (substring), but tightening asserts the flag form specifically and matches the directive. Safe (new messages contain the flag substring).
4. **Include the 2 out-of-range green init tests (321/373)? → YES.** They're the contract's "~15" pair (13 in-range + 2 = 15). Currently green; flip is consistency-only. S2 may also touch them; idempotent, no conflict.
5. **Rename the two Completion*Shell* tests to Completions*Shell*? → optional but recommended.** The "Subcommand"→"Flag" rename is mandatory for the 3 named "*Subcommand*"; the plural rename of the two Shell tests is consistency polish.
6. **Touch test doc comments? → light prose only where they describe a bare token as the mode selector.** Not all comments need it; keep the diff focused on the contract.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports. (GOTCHA #9)

OWNERSHIP (no conflicts):
  - S1 (done, 594be07): parseArgs flag cases.
  - S2 (done, 1e2fe73): exclusivityError flag messages.
  - T2.S1 (working-tree, uncommitted): usageText + error prefixes + ErrNotFound. PRESENT; this task assumes it.
  - T3.S1 (this): main_test.go parseArgs-level + exclusivity-level flips + 5 new tests.
  - T3.S2 (later): main_test.go help-text + dispatch + unconfigured tests.

NO SOURCE / NO ROUTES / NO DATABASE / NO COMPLETIONS:
  - T3.S1 edits main_test.go ONLY. The completion FILES (completions/*) are P1.M2.T1; source is done.
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (hard gates)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main_test.go        # must print NOTHING (gofmt -w if it lists the file)
go vet ./...                 # expect exit 0
go build ./...               # expect exit 0
git diff --stat              # expect ONLY main_test.go changed (no main.go / internal/* / go.mod)
```

### Level 2: Targeted test run (the families this task owns — must be GREEN)

```bash
cd /home/dustin/projects/skilldozer

# parseArgs-level (flipped + renamed + new) + exclusivity-level (flipped). Use a -run regex covering all of them.
go test -v -run 'TestParseArgs(Check|Init|Completions|Bare|Store|Shell)|TestRunExclusivity(Check|Init|Completion|TagsAnd|PathAnd|ListAnd|AllAnd|ListingMode)|TestExclusivityError' ./...
# Expected: all PASS. Specifically these (renamed) names should appear as PASS:
#   TestParseArgsCheckFlag, TestParseArgsInitFlag, TestParseArgsCompletionsFlag,
#   TestParseArgsBareCheckNowTag, TestParseArgsBareInitNowTag, TestParseArgsBareCompletionsNowTag,
#   TestParseArgsInitFlagWithDir, TestParseArgsInitEqualsDir, TestParseArgsInitFlagLiteralInitStore (or deleted),
#   TestRunExclusivityCheckAndTags, ...InitAndList, ...InitInit (rewritten), ...CompletionAndTag, etc.
# If any FAIL, READ the assertion vs the actual parseArgs/exclusivityError behavior and fix.
```

### Level 3: Scope invariants (prove T3.S1 stayed in its lane)

```bash
cd /home/dustin/projects/skilldozer

# Only main_test.go changed:
git diff --name-only                  # Expected: main_test.go (ONLY)
git diff --quiet go.mod go.sum && echo "deps unchanged"  # Expected: deps unchanged

# No bare check/init/completion tokens remain in the FLIPPED test bodies (the NEW bare-tag tests are the exception):
# (Manual eyeball: the only `[]string{"check"}` / `{"init"}` / `{"completions"}` literals should be in the
#  3 NEW BareNowTag tests; everything else in scope uses --check/--init/--completions.)

# The remaining RED tests (if any) are EXACTLY S2's scope (help-text + dispatch + unconfigured) — NOT compile errors:
go test ./... 2>&1 | grep -E '^--- FAIL' | sort
# Expected RED set (S2's scope, NOT this task's): TestRunHelpShowsInitRow, TestRunHelpShowsCompletionRow,
#   TestRunCheck*, TestRunInitStoreWritesConfig*, TestRunInitStoreTildeExpandsHome, TestRunCompletion*,
#   TestRunBareTagUnconfiguredNeverPrompts. (And possibly skillsdir_test ErrNotFound-message.)
# CRITICAL: there must be NO TestParseArgs* or TestRunExclusivity* in the RED set (those are this task's and must be GREEN).
```

### Level 4: Namespace-safety confidence (the point of decision 19)

```bash
cd /home/dustin/projects/skilldozer

# The 3 new BareNowTag tests + 2 new InitFlag tests prove the namespace-safety guarantee end-to-end at the unit level:
go test -v -run 'TestParseArgsBare(Check|Init|Completions)NowTag|TestParseArgsInit(FlagWithDir|EqualsDir)' ./...
# Expected: all PASS. (Bare check/init/completions → tags; --init <dir>/--init=<dir> → initStore, no tags.)

# Plus a binary-level smoke (optional, proves the user-visible contract):
go build -o /tmp/sdz . && /tmp/sdz --check >/dev/null 2>&1; echo "--check exit=$?"   # 0 (example skill clean)
/tmp/sdz check >/dev/null 2>&1; echo "bare check exit=$?"                            # 1 (unknown tag)
rm -f /tmp/sdz
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go vet ./...` exit 0, `go build ./...` exit 0; `git diff --stat` shows ONLY main_test.go
- [ ] Level 2 PASS — targeted `-run` selector all GREEN (parseArgs-level + exclusivity-level + new tests)
- [ ] Level 3 PASS — no `TestParseArgs*`/`TestRunExclusivity*` in the remaining RED set; remaining reds are ONLY help-text/dispatch/unconfigured (S2's scope); `git diff go.mod go.sum` → "deps unchanged"
- [ ] Level 4 PASS — the 5 new namespace-safety tests GREEN; binary smoke (`--check` exit 0, bare `check` exit 1)

### Feature Validation
- [ ] All 13 in-range parseArgs tests + 2 out-of-range (321/373) pass `--flag` tokens
- [ ] 5 parseArgs tests renamed `*Subcommand*`→`*Flag*` (1224/1262/1389 + optional 1404/1418)
- [ ] `TestRunExclusivityInitInit` (2006) rewritten to `["--init","store1","straytag"]`, exit 2, config-not-written preserved
- [ ] `TestParseArgsInitInitCapturedAsTag` (1457) rewritten (initStore="init") or deleted — NOT a naive `["--init","init"]` flip
- [ ] All 14 exclusivity tests flip tokens + tighten `Contains` to `--flag` form
- [ ] 5 new namespace-safety tests added and GREEN

### Code Quality / Convention Validation
- [ ] New tests follow the existing standalone one-per-case parseArgs-capture style (not table-driven)
- [ ] `--completions` is PLURAL everywhere (never `--completion`)
- [ ] Minimal diff (token swaps + the 2 rewrites + 5 new tests); no churn in untouched tests
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical

### Scope Discipline (the S2 + source boundaries)
- [ ] Did NOT touch help-text tests (TestRunHelpShowsInitRow 2029, TestRunHelpShowsCompletionRow 2116) — S2
- [ ] Did NOT touch dispatch tests (TestRunCheck*, TestRunInitStoreWritesConfig*, TestRunCompletion*) — S2
- [ ] Did NOT touch TestRunBareTagUnconfiguredNeverPrompts (2867) — S2
- [ ] Did NOT touch the 5 non-bare exclusivity tests, TestExclusivityErrorListingModes, or Issue-6 listing-mode tests
- [ ] Did NOT modify `main.go`, `internal/*`, `go.mod`, `go.sum`, or any source file (already done)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't naively flip `TestRunExclusivityInitInit` to `["--init","--init"]`.** It's idempotent → NO conflict → NOT exit 2 (GOTCHA #1). And don't use the item-description's `["--init","sometag"]` either (sometag→initStore, no conflict). Use `["--init","store1","straytag"]` (a SECOND positional is the only way to get init+tags).
- ❌ **Don't naively flip `TestParseArgsInitInitCapturedAsTag` to `["--init","init"]` expecting `tags==["init"]`.** It yields `initStore="init"`, `tags=[]` (GOTCHA #2). Rewrite to assert initStore="init", or delete.
- ❌ **Don't treat the remaining RED tests as your problem.** Help-text + dispatch + unconfigured tests are S2's scope (GOTCHA #5). Your gate is the parseArgs + exclusivity families; fix only those.
- ❌ **Don't edit any `.go` source.** parseArgs/exclusivityError/usageText/ErrNotFound are already done (S1+S2+T2.S1). T3.S1 = main_test.go ONLY.
- ❌ **Don't skip the message-assertion tightening.** Contract LOGIC(c) says update ALL error message assertions. Flip the token AND tighten `Contains("X")`→`Contains("--X")` (GOTCHA #3).
- ❌ **Don't write `--completion` (singular).** The flag is `--completions` (PLURAL, decision 19). The bare subcommand was singular; the flag is not (GOTCHA #6).
- ❌ **Don't rename tests whose name lacks "Subcommand".** Only the 3 `*Subcommand*` parseArgs tests MUST rename; the 2 `Completion*Shell*`→`Completions*Shell*` is optional polish (GOTCHA #4).
- ❌ **Don't touch the 5 non-bare exclusivity tests or TestExclusivityErrorListingModes.** They pass no bare check/init/completion tokens and need no change (GOTCHA #5).
- ❌ **Don't add deps or imports.** main_test.go already imports everything the new tests use (testing/strings). go.mod/go.sum byte-for-byte identical (GOTCHA #9).

---

## Confidence Score

**9.5/10** — Every flip is pinned to a verified live line number with exact old→new args in `research/verified_facts.md` §2; the two WRONG flip-table entries (the single biggest failure risk) are proven wrong by probe (§1) with verified rewrites; the message-assertion tightening and scope boundary (no help-text/dispatch/source) are explicit; and the source contract (parseArgs/exclusivityError behavior) is already implemented and probed. The 0.5 reservation is the `TestParseArgsInitInitCapturedAsTag` rewrite-vs-delete choice (§1B) — both are valid; the PRP recommends rewrite but an orchestrator preferring zero-redundancy may delete, and that call is theirs.
