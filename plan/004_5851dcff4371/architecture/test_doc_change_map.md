# Test Change Map — Delta 004 (main_test.go + README.md)

Exact line numbers verified against HEAD `f30d5c5`.

---

## main_test.go (3089 lines)

### Scope: ~45-46 test functions pass bare tokens; ~99 references total

The flip is **mechanical**: every test that passes a bare `"check"` token to
`parseArgs()` or `run()` expecting `c.check=true` (or a check-mode dispatch)
must instead pass `"--check"`. Same for `"init"` → `"--init"`, `"completion"` →
`"--completions"`.

### parseArgs-level tests to flip (~15 functions)

| Function | Line | Token | New token |
|----------|------|-------|-----------|
| `TestParseArgsCheckSubcommand` | 1224 | `"check"` | `"--check"` |
| `TestParseArgsCheckAfterFlag` | 1235 | `"check"` | `"--check"` |
| `TestParseArgsCheckAndTagBothCaptured` | 1248 | `"check"` | `"--check"` |
| `TestParseArgsInitSubcommand` | 1262 | `"init"` | `"--init"` |
| `TestParseArgsInitPositionalDir` | 1279 | `"init"`, `"/tmp/x"` | `"--init"`, `"/tmp/x"` |
| `TestParseArgsInitStoreLongForm` | 1293 | `"init"`, `"--store"` | `"--init"`, `"--store"` |
| `TestParseArgsInitStoreEqualsForm` | 1310 | `"init"`, `"--store=..."` | `"--init"`, `"--store=..."` |
| `TestParseArgsInitStoreLongFormNoValueSetsSignal` | 1343 | `"init"`, `"--store"` | `"--init"`, `"--store"` |
| `TestParseArgsCompletionSubcommand` | 1389 | `"completion"` | `"--completions"` |
| `TestParseArgsCompletionShellLongForm` | 1404 | `"completion"`, `"--shell"` | `"--completions"`, `"--shell"` |
| `TestParseArgsCompletionShellEqualsForm` | 1418 | `"completion"`, `"--shell=..."` | `"--completions"`, `"--shell=..."` |
| `TestParseArgsInitDirNotCapturedAsTag` | 1445 | `"init"`, `"/tmp/x"` | `"--init"`, `"/tmp/x"` |
| `TestParseArgsInitInitCapturedAsTag` | 1457 | `"init"`, `"init"` | `"--init"`, `"init"` (2nd is now a tag) |
| `TestRunInitStoreNoValueExits2` | 321 | `"init"`, `"--store"` | `"--init"`, `"--store"` |
| `TestRunInitStoreNoValueDoesNotWriteConfig` | 373 | `"init"`, `"--store"` | `"--init"`, `"--store"` |

**RENAME**: `TestParseArgsCheckSubcommand` → `TestParseArgsCheckFlag` (etc.) for
accuracy.

### Exclusivity tests to flip (~19 functions)

All `TestRunExclusivity*` functions (lines 1795–2115) that pass bare tokens must
flip to `--flags`. Error message assertions must be updated:

| Function | Line | Old assertion | New assertion |
|----------|------|---------------|--------------|
| `TestRunExclusivityCheckAndTags` | 1867 | `'check' cannot be combined` | `'--check' cannot be combined` |
| `TestRunExclusivityCheckAndList` | 1882 | `'check' cannot...` | `'--check' cannot...` |
| `TestRunExclusivityCheckAndPath` | 1897 | `'check' cannot...` | `'--check' cannot...` |
| `TestRunExclusivityInitAndList` | 1917 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitAndPath` | 1932 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitAndCheck` | 1948 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitAndSearch` | 1963 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitAndAll` | 1975 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitAndStrayTag` | 1988 | `'init' cannot...` | `'--init' cannot...` |
| `TestRunExclusivityInitInit` | 2006 | `"init"`, `"init"` | `"--init"`, `"--init"` (now init+tags conflict) |
| `TestRunExclusivityCompletionAndTag` | 2050 | `'completion' cannot...` | `'--completions' cannot...` |
| `TestRunExclusivityCompletionAndList` | 2065 | `'completion' cannot...` | `'--completions' cannot...` |
| `TestRunExclusivityCheckAndCompletion` | 2081 | bare check+completion | `--check` + `--completions` |
| `TestRunExclusivityInitAndCompletion` | 2099 | bare init+completion | `--init` + `--completions` |
| `TestExclusivityErrorListingModes` | 2350 | (table-driven, includes mode messages) | update any check/init/completion refs |

### Dispatch tests to flip (~11 functions)

| Function | Line | Token flip |
|----------|------|------------|
| `TestRunCheckCleanStore` | 1474 | `"check"` → `"--check"` |
| `TestRunCheckReportsMissingNameExit1` | 1502 | same |
| `TestRunCheckReportsDuplicateNames` | 1523 | same |
| `TestRunCheckWarnOnlyExitsZero` | 1546 | same |
| `TestRunCheckEmptyStoreExit0` | 1567 | same |
| `TestRunCheckSkillsDirUnresolvable` | 1581 | same |
| `TestRunCheckStatusColumnAligned` | 1598 | same |
| `TestRunVersionPrecedenceOverCheck` | 1622 | `"check"` → `"--check"` |
| `TestRunTagStillResolvesAlongsideCheck` | 1637 | `"check"` → `"--check"` |
| `TestRunInitStoreWritesConfig...` | 2760 | `"init"` → `"--init"` |
| `TestRunInitStoreTildeExpandsHome` | 2820 | `"init"` → `"--init"` |
| `TestRunCompletionBashScript` | 2953 | `"completion"` → `"--completions"` |
| `TestRunCompletionFishScript` | 2969 | same |
| `TestRunCompletionUnsupportedShell` | 2986 | same |
| `TestRunCompletionUndetectableShell` | 3004 | same |
| `TestRunCompletionEnvShellDetected` | 3022 | same |
| `TestRunCompletionLoginShellDetected` | 3037 | same |

### Help text tests to update

| Function | Line | Change |
|----------|------|--------|
| `TestRunHelpShowsInitRow` | 2029 | Assert `--init` (not bare `init`) in help output |
| `TestRunHelpShowsCompletionRow` | 2116 | Assert `--completions` (not bare `completion`) |

### Unconfigured test

| Function | Line | Change |
|----------|------|--------|
| `TestRunBareTagUnconfiguredNeverPrompts` | 2867 | Assert stderr contains `run \`skilldozer --init\`` (not bare `init`) |

### NEW tests to add (namespace safety)

Add tests verifying the namespace-safety guarantee:
- `TestParseArgsBareCheckNowTag` — bare `"check"` → `c.check=false`, `c.tags=["check"]`
- `TestParseArgsBareInitNowTag` — bare `"init"` → `c.init=false`, `c.tags=["init"]`
- `TestParseArgsBareCompletionsNowTag` — bare `"completions"` → `c.completion=false`, `c.tags=["completions"]`
- `TestParseArgsInitFlagWithDir` — `"--init"`, `"/tmp/x"` → `c.init=true`, `c.initStore="/tmp/x"`, tags empty
- `TestParseArgsInitEqualsDir` — `"--init=/tmp/x"` → `c.init=true`, `c.initStore="/tmp/x"`

---

## README.md (345 lines)

### Bare subcommand references to update (~16 lines):

| Line | Old | New |
|------|-----|-----|
| 43 | `skilldozer init` | `skilldozer --init` |
| 63 | `skilldozer init` | `skilldozer --init` |
| 66 | `skilldozer init` | `skilldozer --init` |
| 76 | `skilldozer init` | `skilldozer --init` |
| 77 | `skilldozer init` | `skilldozer --init` |
| 89 | `skilldozer init` | `skilldozer --init` |
| 107 | `eval "$(skilldozer completion)"` | `eval "$(skilldozer --completions)"` |
| 113 | `skilldozer completion --shell fish \| source` | `skilldozer --completions --shell fish \| source` |
| 117 | (detection note references `completion`) | update to `--completions` |
| 183 | `skilldozer check` | `skilldozer --check` |
| 240 | `skilldozer check` / `skilldozer init` | `skilldozer --check` / `skilldozer --init` |
| 281 | `skilldozer check` | `skilldozer --check` |
| 297 | `skilldozer init` | `skilldozer --init` |
| 315 | `skilldozer init` | `skilldozer --init` |
| 330 | `skilldozer init` | `skilldozer --init` |

### "Reserved tag names" paragraph (lines 239-247) — DELETE entirely

This section is now factually wrong: there are NO reserved names. Replace with
a one-line note: "Bare words are always skill tags; actions are `--flags` (§6.1)."

### Completions section (lines 94-150) — update

- Eval/source commands now use `--completions`
- Describe skills-first + long-form-only behavior (decision 20)
- Note that bare `<tab>` lists skills; `skilldozer -<tab>` lists long-form flags

### Verification

```bash
grep -q 'skilldozer --completions' README.md    # new commands present
! grep -q 'Reserved tag names' README.md        # old section removed
! grep -q 'skilldozer init\b' README.md          # no bare init (word boundary)
! grep -q 'skilldozer check\b' README.md         # no bare check
! grep -q 'skilldozer completion\b' README.md    # no bare completion
```
