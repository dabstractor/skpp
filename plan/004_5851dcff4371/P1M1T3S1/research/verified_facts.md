# Verified Facts — P1.M1.T3.S1 (flip parseArgs-level + exclusivity-level tests to --flag contract)

All facts verified directly against the working tree on this date.
Probes were run via throwaway `probe_test.go` (deleted after); see §1 for the probe output.

---

## §0 — Current repo state (what T3.S1 executes against)

- HEAD `1e2fe53 "Rewrite mode exclusivity messages for flags"` = **P1.M1.T1.S2 (exclusivityError) committed**.
- Its parent `594be07 "Replace bare subcommands with flags in parseArgs"` = **P1.M1.T1.S1 (parseArgs) committed**.
- **T2.S1 (usageText + error-prefix strings + ErrNotFound) is UNCOMMITTED but PRESENT in the working tree**
  (`git status` shows `M main.go`, `M internal/skillsdir/skillsdir.go`). Consequence:
  - `usageText` already shows `--check`/`--init [<dir>]`/`--completions [--shell <name>]` + the long-form-only note.
  - The 8 `"skilldozer init: "` error prefixes are now `"skilldozer --init: "`; `ErrNotFound` says `run \`skilldozer --init\``.
  - → `TestRunHelpShowsInitRow` / `TestRunHelpShowsCompletionRow` are RED (they assert bare `skilldozer init`/`completion`).
    **Those help-text tests are P1.M1.T3.S2's scope, NOT this task's.** (See §4.)
- `parseArgs` flag cases present: `case "--check":` @296, `case "--completions":` @299, `case "--init":` @327
  (positional <dir> capture via `!strings.HasPrefix(next, "-")`); `=`-form switch has `--init` @228, `--check` @236, `--completions` @238.
- `exclusivityError` (now committed) emits single-quoted flag messages:
  - `'--check' cannot be combined with tag arguments` (789)
  - `'--check' cannot be combined with --completions/--path/--list/--search/--all` (792)
  - `'--init' cannot be combined with tag arguments` (801)
  - `'--init' cannot be combined with --check/--completions/--list/--search/--all/--path` (804)
  - `'--completions' cannot be combined with tag arguments` (812)
  - `'--completions' cannot be combined with --check/--init/--path/--list/--search/--all` (815)
- `main_test.go` last modified Jul 9 (pre-delta); **line numbers in `test_doc_change_map.md` match the live file EXACTLY** (verified by grep).

---

## §1 — CRITICAL: two flip-table entries in `test_doc_change_map.md` / item-description are WRONG

The `--init` case OWNS the next positional (§6.3 tradeoff): a following NON-dashed token becomes
`c.initStore`, NEVER a tag. This makes two Issue-4-derived tests' naive flips FAIL. Proven by probe:

```
ARGS=[--init --init]      => init=true initStore=""   tags=[]       | exclusivity bad=false msg=""
ARGS=[--init sometag]     => init=true initStore="sometag" tags=[]   | exclusivity bad=false msg=""
ARGS=[--init init]        => init=true initStore="init"    tags=[]   | exclusivity bad=false msg=""
ARGS=[--init store1 straytag] => init=true initStore="store1" tags=[straytag] | exclusivity bad=true msg="'--init' cannot be combined with tag arguments"
```

### Finding 1A — `TestRunExclusivityInitInit` (main_test.go:2006)
- OLD: `run([]string{"init", "init"})` → Issue-4 captured 2nd `init` as a tag → init+tags → exit 2.
- NAIVE flip `["--init", "--init"]` → idempotent (`c.init` set twice), NO tags, NO conflict → **falls through to init dispatch, NOT exit 2**. The exit-2 + config-not-written assertions FAIL.
- **Item-description "alternatively" `["--init", "sometag"]` is WRONG**: `sometag` → `initStore`, NOT a tag → NO conflict → NOT exit 2.
- **CORRECT rewrite: `["--init", "store1", "straytag"]`** → `initStore="store1"`, `tags=["straytag"]` → init+tags conflict → exit 2, config NOT written. Preserves the test's purpose (exclusivity fires before init dispatch). Rename comment from "Issue 4" to the new rationale.
- (Note: this becomes functionally adjacent to `TestRunExclusivityInitAndStrayTag`; the distinct value is the config-not-written invariant. Use distinct tokens `store1`/`straytag` to avoid exact duplication.)

### Finding 1B — `TestParseArgsInitInitCapturedAsTag` (main_test.go:1457)
- OLD: `parseArgs([]string{"init", "init"})` → Issue-4: 2nd `init` → tag → asserts `tags==["init"]`, `initStore==""`.
- **Flip-table note "(2nd is now a tag)" is WRONG**: `["--init", "init"]` → `initStore="init"`, `tags=[]`. The `initStore==""` and `tags==["init"]` assertions BOTH FAIL.
- The Issue-4 behavior is GONE (S1 deleted it). Two valid resolutions:
  - **REWRITE (primary)**: assert the real new behavior — `["--init", "init"]` → `c.init==true`, `c.initStore=="init"`, `len(c.tags)==0` (a store dir literally named "init" is accepted as the positional). Rename → `TestParseArgsInitFlagLiteralInitStore`. Keeps a namespace-safety guard at that location.
  - **DELETE (alternative)**: behavior now covered by the new `TestParseArgsInitFlagWithDir`. If the orchestrator prefers no redundancy, delete the function + its preceding comment block.

---

## §2 — Complete flip inventory (verified line numbers)

### parseArgs-level — 13 IN-range (1224–1457) + 2 OUT-of-range = 15 (matches contract "~15")

IN-range (flip bare token; RENAME `Subcommand`→`Flag` where the name says "Subcommand"):

| Function | Line | Old args | New args | Rename |
|---|---|---|---|---|
| TestParseArgsCheckSubcommand | 1224 | `["check"]` | `["--check"]` | → TestParseArgsCheckFlag |
| TestParseArgsCheckAfterFlag | 1235 | `["--no-color","check"]` | `["--no-color","--check"]` | → TestParseArgsCheckAfterFlag (name fine) |
| TestParseArgsCheckAndTagBothCaptured | 1248 | `["check","sometag"]` | `["--check","sometag"]` | (name fine) |
| TestParseArgsInitSubcommand | 1262 | `["init"]` | `["--init"]` | → TestParseArgsInitFlag |
| TestParseArgsInitPositionalDir | 1279 | `["init","/tmp/x"]` | `["--init","/tmp/x"]` | → TestParseArgsInitPositionalDir (name fine) |
| TestParseArgsInitStoreLongForm | 1293 | `["init","--store","/tmp/x"]` | `["--init","--store","/tmp/x"]` | (name fine) |
| TestParseArgsInitStoreEqualsForm | 1310 | `["init","--store=/tmp/x"]` | `["--init","--store=/tmp/x"]` | (name fine) |
| TestParseArgsInitStoreLongFormNoValueSetsSignal | 1343 | `["init","--store"]` | `["--init","--store"]` | (name fine) |
| TestParseArgsCompletionSubcommand | 1389 | `["completion"]` | `["--completions"]` | → TestParseArgsCompletionsFlag |
| TestParseArgsCompletionShellLongForm | 1404 | `["completion","--shell","bash"]` | `["--completions","--shell","bash"]` | → TestParseArgsCompletionsShellLongForm |
| TestParseArgsCompletionShellEqualsForm | 1418 | `["completion","--shell=bash"]` | `["--completions","--shell=bash"]` | → TestParseArgsCompletionsShellEqualsForm |
| TestParseArgsInitDirNotCapturedAsTag | 1445 | `["init","/tmp/x"]` | `["--init","/tmp/x"]` | (name fine) |
| TestParseArgsInitInitCapturedAsTag | 1457 | `["init","init"]` | **SPECIAL — see §1B** | (rename per rewrite) |

OUT-of-range, currently GREEN, flip for consistency (the contract's "~15" pair; trivially `init`→`--init`):

| Function | Line | Old args | New args |
|---|---|---|---|
| TestRunInitStoreNoValueExits2 | 321 | `["init","--store"]` | `["--init","--store"]` |
| TestRunInitStoreNoValueDoesNotWriteConfig | 373 | `["init","--store"]` | `["--init","--store"]` |

(These are `run()`-level; currently GREEN because `--store` last-token sets `storeMissingValue` → exit 2
fires before the init+tags exclusivity check, so `init`→tag doesn't change the outcome. The flip is
consistency-only. **Flagged as a boundary: S2 (run-level dispatch) may also touch them; the edit is
idempotent so no real conflict.**)

parseArgs tests in range needing NO change (no bare token): TestParseArgsStoreWithoutInitToken (1325),
TestParseArgsInitStoreEqualsFormEmptyValueSetsSignal (1358), TestParseArgsStoreNoValueNoInitTokenSetsSignal (1374),
TestParseArgsShellImpliesCompletion (1431).

### exclusivity-level — 14 NEED flip (of the 19 in range 1795–2115)

The 14 passing bare `check`/`init`/`completion` (flip token; tighten message assertion per §3):

| Function | Line | Old args | New args |
|---|---|---|---|
| TestRunExclusivityCheckAndTags | 1867 | `["check","foo"]` | `["--check","foo"]` |
| TestRunExclusivityCheckAndList | 1882 | `["check","--list"]` | `["--check","--list"]` |
| TestRunExclusivityCheckAndPath | 1897 | `["check","--path"]` | `["--check","--path"]` |
| TestRunExclusivityInitAndList | 1917 | `["init","--list"]` | `["--init","--list"]` |
| TestRunExclusivityInitAndPath | 1932 | `["init","--path"]` | `["--init","--path"]` |
| TestRunExclusivityInitAndCheck | 1948 | `["init","check"]` | `["--init","--check"]` |
| TestRunExclusivityInitAndSearch | 1963 | `["init","--search","q"]` | `["--init","--search","q"]` |
| TestRunExclusivityInitAndAll | 1975 | `["init","--all"]` | `["--init","--all"]` |
| TestRunExclusivityInitAndStrayTag | 1988 | `["init","foo","bar"]` | `["--init","foo","bar"]` |
| TestRunExclusivityInitInit | 2006 | `["init","init"]` | **SPECIAL — see §1A** (`["--init","store1","straytag"]`) |
| TestRunExclusivityCompletionAndTag | 2050 | `["completion","example"]` | `["--completions","example"]` |
| TestRunExclusivityCompletionAndList | 2065 | `["completion","--list"]` | `["--completions","--list"]` |
| TestRunExclusivityCheckAndCompletion | 2081 | `["check","completion"]` | `["--check","--completions"]` |
| TestRunExclusivityInitAndCompletion | 2099 | `["init","completion"]` | `["--init","--completions"]` |

The 5 in range needing NO change (no bare token): TestRunExclusivityTagsAndList (1795),
TagsAndSearch (1810), TagsAndAll (1822), TagsAndPath (1837), PathAndTag (1852).

Out-of-range exclusivity needing NO change: TestExclusivityErrorListingModes (2350) — table-driven on
`config{}` literals; its cases use only path/list/searchMode/all/file/noColor/relative (NO check/init/
completion); asserts `Contains("mutually exclusive")`. **No edit.** Also no change: TestExclusivityErrorTagsAndPath
(2390, uses `config{tags,path}` literal), TestRunExclusivityListAndSearch/AllAndList/PathAndList/ListingModePairs
(2408+, Issue-6 listing-mode pairs, no bare tokens).

---

## §3 — exclusivity message assertions: loose `Contains`, tighten to flag form

The exclusivity tests assert LOOSE substrings that happen to pass for both old and new messages:
- `Contains("check")` passes for `'--check' cannot...` (substring).
- `Contains("init")` passes for `'--init' cannot...`
- `Contains("completion")` passes for `'--completions' cannot...` (substring of "completions").

**So the message assertions are NOT why these tests are red** — the RED cause is the bare INPUT token
(no longer setting the mode flag → wrong exit code). The token flip is the ESSENTIAL fix.

Contract LOGIC(c) says "Update ALL error message assertions: 'check'→'--check', ...". Tighten each
loose `Contains("X")` → `Contains("--X")` (quote-insensitive) so the test asserts the FLAG form, not
the bare word. Mapping:
- `Contains("check")` → `Contains("--check")`  (1867/1882 etc.)
- `Contains("init")`  → `Contains("--init")`   (1917/1932/1948 etc.)
- `Contains("completion")` → `Contains("--completions")`  (2050/2065/2081/2099)
- `Contains("tag")` stays `Contains("tag")` (1988/2006 — about stray tags, not modes).
- `Contains("cannot be combined")` / `Contains("mutually exclusive")` stay as-is.

This tightening is SAFE: the new messages contain `--check`/`--init`/`--completions` as substrings.

---

## §4 — Scope boundary (do NOT touch — other tasks own these)

- **Help-text tests** — `TestRunHelpShowsInitRow` (2029), `TestRunHelpShowsCompletionRow` (2116): assert
  `--help` contains bare `skilldozer init`/`completion`. RED now (usageText already flipped). → **P1.M1.T3.S2**.
- **Dispatch tests** — TestRunCheck* (1474–1637), TestRunInitStoreWritesConfig* (2760), TestRunInitStoreTildeExpandsHome
  (2820), TestRunCompletion* (2953–3037): RED, pass bare tokens. → **P1.M1.T3.S2**.
- **Unconfigured test** — TestRunBareTagUnconfiguredNeverPrompts (2867): asserts `run \`skilldozer --init\``.
  RED now (ErrNotFound already flipped). → **P1.M1.T3.S2**.
- **skillsdir_test.go** — the ErrNotFound exact-message test: → **P1.M1.T3.S2** (or already handled by T2.S1's tree).
- `usageText`, `exclusivityError`, `parseArgs`, `ErrNotFound` in main.go/skillsdir.go — **already done** (S1+S2+T2.S1 tree). T3.S1 edits `main_test.go` ONLY.

T3.S1's hard gate: the parseArgs-level + exclusivity-level tests it touches go GREEN. Other reds (help-text,
dispatch, unconfigured) are EXPECTED until S2 lands. (See PRP Validation §Level 1 for the targeted -run selector.)

---

## §5 — NEW namespace-safety tests to ADD (contract LOGIC(e))

Place in the parseArgs section, after `TestParseArgsInitDirNotCapturedAsTag`/the rewritten InitInit test
(around line 1470), before the `// --- run: skilldozer check` divider (~1473). Verified expected behavior:

| New function | Args | Assertions |
|---|---|---|
| TestParseArgsBareCheckNowTag | `["check"]` | `c.check==false`, `c.tags==["check"]` |
| TestParseArgsBareInitNowTag | `["init"]` | `c.init==false`, `c.tags==["init"]` |
| TestParseArgsBareCompletionsNowTag | `["completions"]` | `c.completion==false`, `c.tags==["completions"]` |
| TestParseArgsInitFlagWithDir | `["--init","/tmp/x"]` | `c.init==true`, `c.initStore=="/tmp/x"`, `len(c.tags)==0` |
| TestParseArgsInitEqualsDir | `["--init=/tmp/x"]` | `c.init==true`, `c.initStore=="/tmp/x"`, `len(c.tags)==0` |

Style: standalone one-function-per-case (matches the existing parseArgs-capture test style, NOT table-driven).
Each ~10-15 lines: a comment citing decision 19 / §6.3 namespace safety, `c := parseArgs(...)`, then
`t.Errorf` guard lines mirroring the flipped Subcommand tests' message wording.

`--completions` is PLURAL (decision 19); the bare-tag test uses `"completions"` (a bare `"completion"`
also becomes a tag — either is valid; the contract names `"completions"`).
