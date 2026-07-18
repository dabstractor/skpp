# main.go CLI Audit — IMPLEMENTED vs PRD

Scout pass over `/home/dustin/projects/skilldozer/main.go` (1577 lines, read in full) and
`/home/dustin/projects/skilldozer/PRD.md` (§6 CLI contract, §8 locating dir, §14 completions).
For each of the 10 requested items: status (IMPLEMENTED / PARTIAL / MISSING) with exact line
numbers.

Status legend: ✅ IMPLEMENTED · 🟡 PARTIAL · ❌ MISSING.

---

## 1. Flags §6.1 — ✅ IMPLEMENTED (all 10 modes)

Every §6.1 mode flag is parsed in `parseArgs` and dispatched in `run`. Long forms are handled
in the exact-match `switch a` (main.go:318–405); short aliases (`-a -l -s -f -p -h -v`) are
handled in `expandShortBundle` (main.go:425–504). `=`-forms land in the prefix switch
(main.go:226–294).

| Flag | Parse site | Dispatch | Status |
|---|---|---|---|
| `--all` / `-a` | main.go:246-247, expandShortBundle `case 'a'` | main.go:790-811 | ✅ |
| `--list` / `-l` | main.go:243-244, expandShortBundle `case 'l'` | main.go:707-738 | ✅ |
| `--search <q>` / `-s` | main.go:250-258 (next-token); main.go:264-266 (`=`-form); expandShortBundle s | main.go:739-767 | ✅ |
| `--check` | main.go:286-287 (`=`), main.go:390-392 (long) | main.go:768-789 | ✅ |
| `--init [<dir>]` | main.go:283-290 (`=`), main.go:395-404 (long) | main.go:630-635 → `runInit` (1310-1396) | ✅ |
| `--link <dir>` | main.go:266-274 (`=`), main.go:403-429 (long) | main.go:651-655 → `runLink` (1411-1490) | 🟡 single only — see §6 |
| `--completions` | main.go:276 (`=`), main.go:376 (long) | main.go:640-645 → `runCompletion` (1550-1577) | ✅ |
| `--path` / `-p` | main.go:239-240, expandShortBundle `case 'p'` | main.go:679-706 | ✅ |
| `--help` / `-h` | main.go:236-238, expandShortBundle `case 'h'` | main.go:562-565 (precedence 1) | ✅ |
| `--version` / `-v` | main.go:233-234, expandShortBundle `case 'v'` | main.go:569-571 (precedence 2) | ✅ |

Help/completions advertise **long forms only** (usageText main.go:106-115 matches PRD §6.1
sidebar); shorts remain valid but unadvertised. ✅ matches PRD.

---

## 2. Modifiers §6.2 — ✅ IMPLEMENTED (all 3)

| Modifier | Parse site | Effect site | Status |
|---|---|---|---|
| `--file` / `-f` | main.go:246, expandShortBundle `case 'f'` | `skillPath` main.go:966 (`p = s.SourceFile`) | ✅ |
| `--no-color` | main.go:248 | gated into `ui.PrintList` calls: main.go:736, 762 | ✅ |
| `--relative` | main.go:246 (`=`), main.go:244 (long) | `skillPath` main.go:969-972 (`filepath.Rel`) | ✅ |

`--file` + `--relative` COMBINE correctly (skillPath precedence: file selects SourceFile, then
relative rewrites it) — main.go:965-973. Modifiers are excluded from `exclusivityError`
(main.go:901 doc comment: "--file/--relative/--no-color are MODIFIERS and never trigger
exclusivity"). ✅

`--no-color` is correctly gated on `isTerminal(stdout) && !c.noColor` (main.go:736, 762). ✅

---

## 3. Mutual exclusivity of mode flags (§6.3) — ✅ IMPLEMENTED

`exclusivityError(c config)` main.go:898-954 enforces four families, each returning a one-line
stderr message + exit 2 (checked at main.go:618):

1. **2+ listing modes among {path,list,searchMode,all}** — main.go:913-920 (`n >= 2`).
2. **tags + an inspection mode** — main.go:922-924.
3. **`--check` + tags** — main.go:925-926.
4. **`--check` + a listing mode** — main.go:927-929.
5. **`--init` + tags / + other modes** — main.go:940-946.
6. **`--completions` + tags / + other modes** — main.go:948-954... actually main.go:955-961.
7. **`--link` + tags / + other modes** — main.go:964-971.

All eight §6.3 mode flags (`--check --init --link --list --search --all --completions --path`)
are covered as mutually exclusive. ✅

**Caveat (ties to §6 below):** the `--link` exclusivity branch at main.go:966-968 *rejects*
trailing positionals as `"'--link' cannot be combined with tag arguments"`. This is correct
for current single-target behavior but **directly contradicts** PRD §8.4/§6.1 which require
`--link` to *collect* every trailing positional as a link directory.

---

## 4. No-args = implicit help to stdout exit 0 (§6.3) — ✅ IMPLEMENTED

End of `run` (main.go:864-869):
```go
// No recognized mode → usage to STDOUT, exit 0 (PRD §6.3 / §19 decision 17)
fmt.Fprint(stdout, usage())
return 0
```
Covers both truly-no-args AND modifiers-only (e.g. `skilldozer --no-color`). Help is emitted
PLAIN to **stdout** (not stderr), exit 0 — matches §6.3 and the grepability contract
(§13: `skilldozer | grep …` must see it). ✅

`--help` itself is precedence-1 (main.go:562-565), also stdout exit 0. ✅

---

## 5. Error semantics §6.4: unknown tag → nothing on stdout, exit 1 — ✅ IMPLEMENTED

Tag-resolution block (main.go:829-859) is correctly **atomic**:
- Paths buffered in `paths := make([]string, 0, len(c.tags))` (main.go:842).
- On any per-tag `rerr != nil`: error printed to **stderr** (main.go:846), `hadErr = true`,
  `continue` (main.go:847-848).
- If `hadErr`: `return 1` **before** the flush loop (main.go:855-857) → stdout stays EMPTY.

```go
paths := make([]string, 0, len(c.tags)) // buffered; flushed only if all resolve
hadErr := false
for _, tag := range c.tags {
    res, rerr := resolve.Resolve(tag, skills)
    if rerr != nil {
        fmt.Fprintln(stderr, rerr) // one error line per problem tag (verbatim)
        hadErr = true
        continue
    }
    paths = append(paths, skillPath(res.Skill, dir, c))
}
if hadErr {
    return 1 // paths buffered but never written → stdout empty (§6.4)
}
```

This satisfies the critical `pi --skill "$(skilldozer badtag)"` fails-loudly contract. ✅

---

## 6. `--link` multi-target batch linking (§8.4) — ❌ MISSING (single-target only)

**This is the headline gap.** PRD §8.4 and §6.1 both specify `--link <dir> [<dir>...]` as a
**batched** operation: pass `--link` once, then every following positional is a directory to
link. The implementation handles exactly ONE directory.

### Evidence the batch is NOT implemented

1. **`config.linkTarget` is a `string`, not a `[]string`** — main.go:177:
   ```go
   linkTarget string // `--link <dir>` value (the skill dir to link); ...
   ```
2. **`parseArgs --link` (long form) captures exactly ONE token** — main.go:403-429:
   ```go
   case "--link":
       if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
           c.link = true
           c.linkTarget = args[i+1]   // single token
           i++
       } else {
           c.linkMissingValue = true
       }
   ```
   A second positional (e.g. `--link dirA dirB`) falls through to the default branch and is
   collected as a **tag** in `c.tags` (main.go:436-437).
3. **`--link=` (`=`-form) also captures a single value** — main.go:266-274.
4. **`runLink` processes a single `c.linkTarget`** — main.go:1411-1490. No loop over targets.
5. **`exclusivityError` REJECTS the second positional** — main.go:964-971:
   ```go
   if c.link {
       if hasTags {
           return true, "skilldozer: '--link' cannot be combined with tag arguments"
       }
       ...
   }
   ```
   So `skilldozer --link dirA dirB` exits **2** with a "tags cannot be combined" error — the
   opposite of the §8.4 batch contract. Compare PRD §13 acceptance (main.md:467-478):
   `# multi-link: one --link, then several directories (§8.4)` and
   `# mixed batch: two valid + one invalid → valid ones link, exit 1, the bad dir named on stderr`.

### What single-target does correctly (so it is 🟡 not totally ❌)
The single-directory path is fully implemented and §8.4-compliant: store resolution
(main.go:1419-1422), `~`/`~/` expansion before `filepath.Abs` (main.go:1429-1437), the three
validations — existing dir / not store-or-inside-it / `HasSkillMD` (main.go:1439-1462),
conflict handling (create / refresh existing symlink / refuse non-symlink) (main.go:1465-1482),
stdout = link path + stderr confirmation (main.go:1488-1490). The exit-2 missing-value case
(main.go:608-610) also works for the single-target form.

### To close the gap (for the implementing agent)
- Convert `linkTarget string` → `linkTargets []string` (config struct, main.go:177).
- In `parseArgs` `--link` / `--link=`: set a "collecting link targets" mode and append **every**
  subsequent non-flag positional to `linkTargets` until end of argv (mirroring how a `--`
  end-of-options works, but scoped to after `--link`). PRD: "once `--link` is parsed, every
  following non-flag token is a directory to link (never a tag)". A dashed follower stays a
  missing-value → exit 2 (or per §8.4 a mode-flag follower is its own error).
- In `exclusivityError`: REMOVE the `--link` "cannot be combined with tag arguments" branch —
  after `--link`, trailing positionals are link targets, not tags, so `hasTags` must be
  re-evaluated to ignore positionals consumed by `--link`.
- In `runLink`: loop over `linkTargets` in input order, per-dir validate+link, collecting
  stdout link paths and stderr errors; exit 0 iff all succeed, 1 iff any fail (successful
  links remain — idempotent). Mixed output: success paths to stdout, per-failure line to
  stderr. (PRD §6.4 / §8.4 step 4.)
- Empty `linkTargets` after `--link` ⇒ exit 2; the existing `linkMissingValue` signal already
  covers the no-value case — extend it to "zero following directories".

---

## 7. `--init` with cwd auto-detect, `--init --store` — ✅ IMPLEMENTED

### cwd auto-detect
`chooseStore` (main.go:1015-1075) implements the 4-step PRD §8.2 decision:
- (1) `haveStore != ""` → non-interactive override, no prompt (main.go:1031-1033).
- (2) Auto-detect default: `skillsdir.HasSkillMD(cwd)` → default = cwd ("detected skills in
  <cwd>"), else `configpkg.DefaultStore()` ($XDG_DATA_HOME/skilldozer/skills)
  (main.go:1035-1039).
- (3) TTY → prompt "Where should skilldozer keep your skills? [<default>]" (main.go:1043-1062);
  readPrompt makes empty/EOF ⇒ default (main.go:1092-1108).
- (4) Non-TTY → default, no prompt (main.go:1041-1043).

`resolveStore` (main.go:1117-1158) wires the real I/O (os.Getwd, DefaultStore,
stdinIsTerminal, bufio prompt) and applies `expandHome` + `filepath.Abs`.

### `--init --store <dir>`
`parseArgs --store` **implies init**: sets `c.init = true` + `c.initStore = <dir>`
(main.go:335-346 for `--store=`, main.go:329-340 for `--store <dir>`). So
`skilldozer --init --store <dir>` lands in `runInit` with `c.initStore` populated →
`resolveStore` returns it verbatim (haveStore != "", no prompt). ✅

`--init <dir>` (positional store) is also handled (main.go:395-404). ✅
Missing-value guards for `--store` (exit 2) before init dispatch (main.go:580-583). ✅

### Minor nits
- `--store` is NOT advertised in `usageText` §8.2 row but IS listed in OPTIONS (main.go:111) —
  acceptable.
- PRD §8.2 step 5 says "Print the output of `skilldozer --path` (which rule won) and
  `skilldozer --check`." `runInit` prints path "found via" to **stderr** and the check report
  to **stderr** (main.go:1376-1394), keeping stdout = store path only (§6.1 stdout contract).
  This is a deliberate, documented divergence (main.go:1386-1389) — acceptable.

---

## 8. `--completions` with `--shell`, embedded scripts — ✅ IMPLEMENTED

### Embedded scripts (PRD §14.6 / §17 "self-sufficient binary")
Three `//go:embed` vars (main.go:51-59):
```go
//go:embed completions/skilldozer.bash   var bashCompletion string
//go:embed completions/_skilldozer       var zshCompletion string
//go:embed completions/skilldozer.fish   var fishCompletion string
```
`completionScript(shell)` (main.go:1235-1248) is a pure switch returning the embedded bytes +
`true`, or `("", false)` for unsupported.

### `--shell` flag
`parseArgs --shell` **implies completion**: sets `c.completion = true` + `c.completionShell`
(main.go:351-358 for `--shell=`, main.go:368-376 for `--shell <name>`). Missing-value guard
→ exit 2 (main.go:587-589). ✅

### Dispatch
`runCompletion` (main.go:1550-1577):
1. `detectShell(explicit, envShell, loginShell)` precedence: `--shell` →
   `$SKILLDOZER_SHELL` → `basename($SHELL)` (main.go:1520-1540). ✅ matches §14.6.
2. Undetectable → stderr "could not detect shell; pass --shell {bash|zsh|fish}", exit **1**,
   nothing on stdout (main.go:1555-1558). ✅
3. Unsupported value → stderr "unsupported shell '%s'", exit **2**, nothing on stdout
   (main.go:1559-1562). ✅
4. bash/fish → emit verbatim embedded bytes via `Fprint` (main.go:1569-1570). ✅
5. zsh → emit DERIVED eval-safe wrapper `zshEvalScript` (strips the `_skilldozer "$@"`
   self-call + appends compdef registration + `setopt NO_LIST_AMBIGUOUS` per §14.7)
   (main.go:1564-1567, 1251-1308). ✅

`--completions` is an exclusive mode (exclusivityError main.go:955-961). ✅

---

## 9. Config file support (§8.1) — ✅ IMPLEMENTED (in `internal/config`)

Config file logic lives in `internal/config/config.go` (referenced from main.go via
`configpkg`), not in main.go directly. main.go consumes it at two sites:
- `configpkg.Path()` — config-file location (runInit main.go:1325-1329).
- `configpkg.Save(configPath, configpkg.File{Store: store})` — write config (setupStore
  main.go:1206-1208).

`internal/config/config.go` provides (verified via grep):
- `Path() (string, error)` (config.go:123) — `$XDG_CONFIG_HOME/skilldozer/config.yaml`,
  overridden by literal `$SKILLDOZER_CONFIG` (const `configEnv = "SKILLDOZER_CONFIG"`,
  config.go:100). ✅ matches §8.1.
- `Load(path) (File, error)` (config.go:60), `Save(path, f)` (config.go:83) — YAML round-trip.
- `File{ Store string }` minimal valid file `store: /home/dustin/skills`.
- Unknown keys ignored (test `TestLoadIgnoresUnknownKeys`, config_test.go:49). ✅ matches §8.1.
- Missing/unreadable config ⇒ "not yet configured" fall-through (§8.3 rules 3-5), never a hard
  error — handled by `skillsdir.Find()`, not config itself. ✅
- `DefaultStore()` (config.go:150) — `$XDG_DATA_HOME/skilldozer/skills`. ✅

---

## 10. `--store` flag handling in arg parser — ✅ IMPLEMENTED

`--store` is parsed in BOTH forms and **implies `--init`** (PRD §8.2 non-interactive store):

### `--store <dir>` (next-token form) — main.go:329-346
```go
case "--store":
    if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
        c.init = true
        c.initStore = args[i+1]
        i++
    } else {
        c.storeMissingValue = true
    }
```
- **Dashed-follower guard**: a following `-`-prefixed token is NOT consumed as the value
  (prevents `--store --check` silently mutating state). Sets `storeMissingValue` → exit 2
  before init dispatch (main.go:580-583). Non-destructive (no config written). ✅
- **No short form** (matches §8.2: --store is not in the advertised short set). ✅

### `--store=<dir>` (`=`-form) — main.go:309-321
```go
case "--store":
    c.init = true
    c.initStore = val
    if val == "" { c.storeMissingValue = true }   // --store= → exit 2
```
Empty `--store=` records `storeMissingValue` → exit 2. ✅

### Missing-value enforcement — main.go:580-583
```go
if c.storeMissingValue {
    fmt.Fprintln(stderr, "skilldozer: --store requires a value")
    return 2
}
```
Runs BEFORE init dispatch, so `runInit`/`setupStore`/`configpkg.Save` are never called → a
pre-existing config.yaml's `store:` is preserved (Issue 2 non-destructive contract). ✅

Symmetric missing-value handling exists for `--search` (main.go:584-586), `--shell`
(main.go:587-589), and `--link` (main.go:608-610).

---

## Summary table

| # | Item | Status | Key line refs |
|---|---|---|---|
| 1 | Flags §6.1 (all 10) | ✅ IMPLEMENTED | parseArgs main.go:226-437; run dispatch 679-869 |
| 2 | Modifiers §6.2 (all 3) | ✅ IMPLEMENTED | parseArgs 244-248; skillPath 965-973 |
| 3 | Mutual exclusivity §6.3 | ✅ IMPLEMENTED | exclusivityError main.go:898-971 |
| 4 | No-args = implicit help, stdout, exit 0 | ✅ IMPLEMENTED | run main.go:864-869 |
| 5 | Unknown tag → nothing on stdout, exit 1 | ✅ IMPLEMENTED | run main.go:842-857 (atomic buffer) |
| 6 | `--link` multi-target batch (§8.4) | ❌ MISSING (single only) | config.linkTarget string 177; parseArgs 403-429; runLink 1411-1490; exclusivity rejects trailing pos 964-971 |
| 7 | `--init` cwd auto-detect + `--init --store` | ✅ IMPLEMENTED | chooseStore 1015-1075; resolveStore 1117-1158; --store implies init 329-346 |
| 8 | `--completions` + `--shell` + embedded scripts | ✅ IMPLEMENTED | go:embed 51-59; completionScript 1235-1248; runCompletion 1550-1577 |
| 9 | Config file support §8.1 | ✅ IMPLEMENTED | internal/config/config.go (Path/Save/Load/DefaultStore); setupStore writes 1206-1208 |
| 10 | `--store` flag handling in arg parser | ✅ IMPLEMENTED | parseArgs 329-346 + 309-321; missing-value guard 580-583 |

**Net: 9 of 10 items fully IMPLEMENTED. Item 6 (`--link` batch) is the single MATERIAL GAP.**

---

## Residual risks / open questions

1. **`--link` batch (§8.4/§6.1) — the only blocker.** Single-target works; multi-target exits
   2 instead of linking. The §13 acceptance suite has explicit `# multi-link` and
   `# mixed batch` cases (PRD §13, lines 467-478) that will FAIL against the current binary.
   Implementing the batch requires parser + exclusivity + runLink changes (see §6 "To close
   the gap"). Severity: **HIGH** (acceptance-critical, user-facing headline feature).
2. **`--link` missing-value message wording** (minor): current `"skilldozer: --link requires a
   path to a skill directory"` (main.go:609) is singular; PRD §6.4 specifies the exact text
   `skilldozer: --link requires at least one path to a skill directory`. Severity: **LOW**
   (string-match tests may care).
3. **`--search` field set is a SUPERSET of PRD §6.1** (also searches `aliases` and `category`,
   not just tag/name/description/keywords). This is a benign superset — matches §7.1 captured
   fields — but if a test asserts the EXACT §6.1 field list it could surface as a deviation.
   Severity: **LOW** (broader than spec, not narrower).
4. **`--init --check` report goes to stderr** by design (§6.1 stdout contract); a test asserting
   the check report on stdout (per §8.2 step 5 literal wording) would need to read stderr.
   Documented divergence (main.go:1386-1389). Severity: **LOW** (intentional).
5. **`--store` / `--init` dashed-follower semantics** treat `--store --check` as missing-value
   exit 2 rather than as two flags. This is a deliberate non-destructive choice (main.go:329
   comment) and is consistent, but differs from `--search`'s greedy dashed-value consume.
   Severity: **INFO** (documented decision, not a bug).
