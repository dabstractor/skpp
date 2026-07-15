# Verified Facts — P1.M1.T3.S2 (run-level dispatch + completion + help-text + unconfigured test flips)

All facts verified by direct file reads + a live `go test ./...` run + an
independent `scout` subagent recon. Line numbers are **POST-S1** (S1
`P1.M1.T3.S1` committed during this research session — it added 5 namespace
safety tests, renamed parseArgs tests, and rewrote the 2 Issue-4 tests; that
shifted main_test.go line numbers ~65 lines BELOW the values in
`test_doc_change_map.md`). **Locate every edit by FUNCTION NAME, not line
number.** Line numbers below are the post-S1 values for orientation only.

---

## §0 — Current repo state (verified)

- `git status --short main_test.go` → **clean** (S1 COMMITTED its changes; HEAD moved).
- `git status --short` shows `main.go` + `internal/skillsdir/skillsdir.go` modified
  (T2.S1 working-tree: usageText + ErrNotFound already flipped to `--flags`).
- `go build ./...` OK. `go vet ./...` OK.
- **`go test ./...` RED count = 23** (down from 34 before S1 landed). S1 owned
  parseArgs-level + exclusivity-level; those are now GREEN. **The remaining 23
  RED tests are 100% S2's scope** (this task). Two packages are RED:
  `github.com/dabstractor/skilldozer` (22) and
  `github.com/dabstractor/skilldozer/internal/skillsdir` (1).

## §1 — CRITICAL: the change map (`test_doc_change_map.md`) MISSED 6 tests

`test_doc_change_map.md` lists ~11 dispatch + ~6 completion + 2 help-text + 1
unconfigured = 20 functions, ALL in main_test.go. But it does **NOT** list the
**6 tests** that are RED purely because T2.S1 changed the `ErrNotFound`
one-line-fix message from `skilldozer init` → `skilldozer --init`. These block
the contract OUTPUT "go test ./... is fully green" and "All bare-subcommand
references in tests are eliminated." They are a DIFFERENT edit shape
(assertion-string flip, NOT a run() token flip):

| Test | File | Line | Assertion (exact) |
|---|---|---|---|
| `TestRunPathFailureErrNotFound` | main_test.go | ~247 | `for _, want := range []string{"run", "skilldozer init"}` |
| `TestRunListSkillsDirUnresolvableExit1` | main_test.go | ~490 | `strings.Contains(errOut.String(), "skilldozer init")` |
| `TestRunTagSkillsDirUnresolvable` | main_test.go | ~704 | `strings.Contains(errOut.String(), "skilldozer init")` |
| `TestRunAllSkillsDirUnresolvable` | main_test.go | ~962 | `strings.Contains(errOut.String(), "skilldozer init")` |
| `TestRunSearchSkillsDirUnresolvable` | main_test.go | ~1202 | `strings.Contains(errOut.String(), "skilldozer init")` |
| `TestErrNotFoundMessageHasFix` | internal/skillsdir/skillsdir_test.go | ~531 | `for _, want := range []string{"run", "skilldozer init"}` |

Fix = flip the assertion string `"skilldozer init"` → `"skilldozer --init"`.
NO run() token changes for these 6 (their run() tokens are `--path`/`--list`/
`example`/`--all`/`--search`/ErrNotFound-direct — all already correct).

Root cause: `internal/skillsdir/skillsdir.go:275` —
`var ErrNotFound = errors.New("skilldozer is not configured; run \`skilldozer --init\`")`
(already `--init`). `run()` prints `err.Error()` verbatim to stderr on the
unconfigured path (main.go ~543, ~641, ~660, ~709, etc.). So every test asserting
the printed fix hint must match `--init`.

## §2 — The 20 change-map functions (run() TOKEN flips + help-text/unconfigured ASSERTION flips)

### §2a Check dispatch — flip run() token `"check"` → `"--check"` (8 funcs)
All: `run([]string{"check"}, …)` → `run([]string{"--check"}, …)`. Assertions
unchanged EXCEPT `TestRunCheckSkillsDirUnresolvable` (see §2a-note).

| Function | run() line | Token | Also needs assertion flip? |
|---|---|---|---|
| `TestRunCheckCleanStore` | ~1543 | `"check"`→`"--check"` | no |
| `TestRunCheckReportsMissingNameExit1` | ~1574 | `"check"`→`"--check"` | no |
| `TestRunCheckReportsDuplicateNames` | ~1595 | `"check"`→`"--check"` | no |
| `TestRunCheckWarnOnlyExitsZero` | ~1618 | `"check"`→`"--check"` | no |
| `TestRunCheckEmptyStoreExit0` | ~1636 | `"check"`→`"--check"` | no |
| `TestRunCheckSkillsDirUnresolvable` | ~1650 | `"check"`→`"--check"` | **YES** — ~1657 `Contains(..., "skilldozer init")`→`"skilldozer --init"` |
| `TestRunCheckStatusColumnAligned` | ~1670 | `"check"`→`"--check"` | no (see §4) |
| `TestRunVersionPrecedenceOverCheck` | ~1691 | `["check","--version"]`→`["--check","--version"]` | no (see §4) |

§2a-note: `TestRunCheckSkillsDirUnresolvable` is RED for TWO reasons — (1) the
bare `"check"` token (now a tag → goes to tag resolution → wrong path) AND (2)
the stderr assertion `Contains("skilldozer init")` (message is now
`skilldozer --init`). Fix BOTH.

### §2b Tag-resolution guard — NO flip (1 func)
`TestRunTagStillResolvesAlongsideCheck` (~1706): `run([]string{"example"}, …)`.
The token is **`"example"`** (a real skill tag), NOT `"check"`. **NO token
flip.** (The change-map row saying `1637: "check"→"--check"` is WRONG — confirmed
by reading the actual run() call + its `// NOT "check" -> tag resolution` comment.)
OPTIONAL: the test name + doc comment reference "check is reserved" framing that
is now obsolete (decision 19: no reserved names). Light prose touch only.

### §2c Help-text — flip ASSERTION strings (2 funcs); run() token is already `--help`
| Function | Assertion line | Old | New |
|---|---|---|---|
| `TestRunHelpShowsInitRow` | ~2103 | `[]string{"skilldozer init", "--store <dir>"}` | `[]string{"skilldozer --init", "--store <dir>"}` |
| `TestRunHelpShowsCompletionRow` | ~2189 | `[]string{"skilldozer completion", "--shell"}` | `[]string{"skilldozer --completions", "--shell"}` |

Verified usageText (main.go:71-119) contains `skilldozer --init [<dir>]` (L81)
and `skilldozer --completions [--shell <name>]` (L82) + `--store <dir>` (L109) +
`--shell <bash|zsh|fish>` (L111). So the flipped assertions WILL pass.

### §2d Init dispatch — flip run() token `"init"` → `"--init"` (2 funcs)
| Function | run() line | Old token slice | New token slice |
|---|---|---|---|
| `TestRunInitStoreWritesConfigCreatesStorePrintsPathExit0` | ~2836 | `[]string{"init", "--store", store}` | `[]string{"--init", "--store", store}` |
| `TestRunInitStoreTildeExpandsHome` | ~2896 | `[]string{"init", "--store", "~/sub"}` | `[]string{"--init", "--store", "~/sub"}` |

### §2e Unconfigured — flip ASSERTION string only (1 func); run() token is a tag
`TestRunBareTagUnconfiguredNeverPrompts` (~2938): `run([]string{"someTag"}, …)`.
Token `"someTag"` is correct (a tag) — **NO token flip**. Assertion ~2946:
`[]string{"run", "skilldozer init"}` → `[]string{"run", "skilldozer --init"}`.

### §2f Completion dispatch — flip run() token `"completion"` → `"--completions"` (6 funcs)
**PLURAL** (`--completions`, decision 19 — NEVER `--completion`).
| Function | run() line | Old token slice | New token slice |
|---|---|---|---|
| `TestRunCompletionBashScript` | ~3021 | `[]string{"completion","--shell","bash"}` | `[]string{"--completions","--shell","bash"}` |
| `TestRunCompletionFishScript` | ~3037 | `[]string{"completion","--shell","fish"}` | `[]string{"--completions","--shell","fish"}` |
| `TestRunCompletionUnsupportedShell` | ~3054 | `[]string{"completion","--shell","tcsh"}` | `[]string{"--completions","--shell","tcsh"}` |
| `TestRunCompletionUndetectableShell` | ~3074 | `[]string{"completion"}` | `[]string{"--completions"}` |
| `TestRunCompletionEnvShellDetected` | ~3091 | `[]string{"completion"}` | `[]string{"--completions"}` |
| `TestRunCompletionLoginShellDetected` | ~3107 | `[]string{"completion"}` | `[]string{"--completions"}` |

## §3 — NOT in S2's scope (S1 already done; do NOT touch)

- All `TestParseArgs*` flips/renames + the 5 NEW namespace-safety tests
  (`TestParseArgsBareCheckNowTag`, `BareInitNowTag`, `BareCompletionsNowTag`,
  `InitFlagWithDir`, `InitEqualsDir`) — S1 COMMITTED these. They intentionally
  pass bare `check`/`init`/`completions` as TAGS. **Leave them.**
- All `TestRunExclusivity*` flips — S1 COMMITTED. GREEN now.
- `TestParseArgsInitInitCapturedAsTag` rewrite + `TestRunExclusivityInitInit`
  rewrite — S1 COMMITTED.
- `TestEmbeddedCompletionsMatchOnDisk` (~2990 post-S1) — **GREEN**, reads on-disk
  `completions/*` files vs embedded vars (both unchanged). Do NOT touch. (The
  completion FILES are rewritten in P1.M2.T1, a later milestone.)
- `TestCompletionScriptMapping`, `TestCompletionScriptUnsupportedShell`,
  `TestDetectShell`, `TestLoginShellBase` — pure-func tests, GREEN, no bare
  tokens. Leave them.
- Any `.go` source (`main.go`, `internal/*`) — already done by S1+S2+T2.S1.

## §4 — Two GREEN-but-vacuous/stale check tests: flip them anyway

Both are currently GREEN but for the wrong reason (bare `"check"` now parsed as
a tag). The contract "All bare-subcommand references in tests are eliminated"
requires flipping the token. After flipping to `"--check"` they actually
exercise the intended behavior and STAY green (verified by reasoning over the
dispatch):

- `TestRunCheckStatusColumnAligned` (~1670): currently vacuous (bare `check`→tag
  not found→empty stdout→alignment loop `continue`s immediately). Flip to
  `"--check"` so the check report runs; the store has good+bad skills → `OK`/`ERROR`
  aligned lines → assertions hold. STAYS GREEN.
- `TestRunVersionPrecedenceOverCheck` (~1691): `["check","--version"]` currently
  green because `--version` precedence wins. Flip to `["--check","--version"]`;
  version still wins → exit 0 + version line. STAYS GREEN.

## §5 — Scope summary (23 RED tests → all become GREEN)

- **Token flips** (run() arg): 8 check + 2 init + 6 completion = **16** (incl.
  the 2 vacuous-green check tests). `TestRunTagStillResolvesAlongsideCheck` +
  `TestRunBareTagUnconfiguredNeverPrompts` have NO token flip (correct tags).
- **Assertion-string flips** (`"skilldozer init"`→`"skilldozer --init"`):
  `TestRunCheckSkillsDirUnresolvable` + `TestRunBareTagUnconfiguredNeverPrompts`
  + the 5 §1 SkillsDirUnresolvable tests + `TestErrNotFoundMessageHasFix` = **8**.
- **Help-text assertion flips** (`"skilldozer init"`→`"skilldozer --init"`,
  `"skilldozer completion"`→`"skilldozer --completions"`): **2**.
- Total distinct tests touched: **23** (the entire current RED set). Plus
  optional comment/name prose on `TestRunTagStillResolvesAlongsideCheck`.

## §6 — Exact edit-safe locator commands (robust to line drift)

```bash
# Every bare "skilldozer init" assertion string (8 sites: 5 §1 + CheckSkillsDir + BareTag + skillsdir):
grep -rn '"skilldozer init"' main_test.go internal/skillsdir/skillsdir_test.go
# The lone "skilldozer completion" assertion:
grep -n '"skilldozer completion"' main_test.go
# Every run() call still passing a bare dispatch token (exclude the NEW namespace tests by name):
grep -n 'run(\[\]string{"check"\|run(\[\]string{"check",\|run(\[\]string{"init"\|run(\[\]string{"init",\|run(\[\]string{"completion"' main_test.go
```
The grep for bare tokens will ALSO hit the 5 NEW namespace-safety tests
(`TestParseArgsBareCheckNowTag` etc.) which call `parseArgs`, not `run`, and
intentionally use bare tokens — so scoping the grep to `run(` avoids them.
