# Verified Facts — P1.M2.T1.S1: Add missing-value exit-2 detection for `--search` and `--shell`

Bugfix round `plan/004_5851dcff4371/bugfix/001_ba6f35f74a59` (Issue 3). Every claim
below was read directly from the live source at `/home/dustin/projects/skilldozer`
(`main.go` config struct / parseArgs / expandShortBundle / run read at exact line
ranges; `main_test.go` store + search/shell tests read at exact line ranges;
`decisions.md` §D4-D5 + `parseargs_research.md` read in full). Module
`github.com/dabstractor/skilldozer`, `go 1.25`, sole third-party dep
`gopkg.in/yaml.v3` (untouched). Zero new deps.

---

## §1 — The bug + the fix in one sentence

`--store` (no value) exits 2 (the `storeMissingValue` pattern), but `--search`/`-s`
and `--shell` (no value) silently fall through to the implicit-help default
(stdout usage, exit 0) — and the code comment at main.go:293 even MISLABELS that
path as "(exit 1)" when it is actually exit 0. Fix (decisions D4): make
`--search`/`-s`/`--shell` no-value exit 2 with an error, mirroring `storeMissingValue`
EXACTLY — a bool field in the config struct, set in parseArgs (no-value branches),
checked in run() before exclusivity dispatch.

---

## §2 — The existing `storeMissingValue` pattern (THE template)

**Field** (main.go:169, in `config`):
```go
storeMissingValue bool     // --store / --store= seen with NO value ...
```

**Set in parseArgs** in 3 places (the `=`-form `--store=`/`--init=` empty-value
branches at main.go:229/237, and the bare `--store` last-token `else` at main.go:318):
```go
} else {
    c.storeMissingValue = true
}
```

**Read once in run()** at main.go:499-502, as precedence step 3.5 (AFTER
help/version/unknownFlag, BEFORE exclusivityError / dispatch):
```go
if c.storeMissingValue {
    fmt.Fprintln(stderr, "skilldozer: --store requires a value")
    return 2
}
```

run() precedence ladder (main.go:466): `help(471) → version(477) → unknownFlag(484)
→ storeMissingValue(499) → exclusivity(508) → init dispatch(520) → completion(528)
→ normal modes → no-args-usage(749)`.

The two new fields/checks are PEERS of this exact pattern — same shape, same
placement, same message style.

---

## §3 — Exact edit anchors in main.go (CURRENT line numbers, sibling-safe)

The parallel sibling P1.M1.T2.S1 edits `internal/skillsdir/*` (Issue 2:
vanished-store). It does NOT touch `main.go` or `main_test.go` — confirmed by
reading its PRP (scope: "Does NOT touch main.go"). So every line number below is
stable. (parseargs_research.md verified all of these empirically.)

```
main.go:153-174  type config struct { ... }         (ADD searchMissingValue + shellMissingValue AFTER storeMissingValue @169)
main.go:288-298  case "--search", "-s":              (ADD `else { c.searchMissingValue = true }`; FIX the misleading comment @293)
main.go:320-326  case "--shell":                     (ADD `else { c.shellMissingValue = true }`; FIX the comment @324)
main.go:444      expandShortBundle -s default:       (ADD `c.searchMissingValue = true`; update the comment)
main.go:493-502  run() storeMissingValue check       (ADD the two peer checks AFTER line 502; UPDATE the precedent comment @493-498)
```

---

## §4 — The three no-value forms for `--search`/`-s` (and which get the fix)

parseargs_research.md §1 enumerates three forms; decisions D5 fixes ONLY the bare
no-token case:

**(1a) `=`-form `--search=`** (main.go:218-220): `c.searchMode = true; c.searchQ = val`
(UNCONDITIONAL — no empty-value guard). **UNCHANGED (D5).** `--search=` is a valid
(if useless) empty query, distinct from `--search` with no following token. Do NOT
add a missing-value guard here. The existing test `TestParseArgsLongEqualsSearchEmpty`
(main_test.go:2294) locks this (searchMode=true, searchQ="") and stays GREEN.

**(1b) main switch `case "--search", "-s":`** (main.go:288-298): the bare no-token
case. CURRENT:
```go
if i+1 < len(args) {
    c.searchMode = true
    c.searchQ = args[i+1]
    i++
}
```
FIX — add the `else`:
```go
if i+1 < len(args) {
    c.searchMode = true
    c.searchQ = args[i+1]
    i++
} else {
    c.searchMissingValue = true
}
```
This covers BOTH `--search` (last token) and bare `-s` (last token, len==2 → main
switch, NOT expandShortBundle — parseargs_research §1c note).

**(1c) short-bundle `-s` no-value** (expandShortBundle default, main.go:444):
e.g. `-vs` / `-ls` as the last token. CURRENT:
```go
default:
    // 's' seen but no value anywhere: mirror the bare "-s"-no-value rule
    // (searchMode stays false). The bool flags before it remain set.
```
FIX — set the signal:
```go
default:
    // 's' seen but no value anywhere: mirror the bare "-s"-no-value rule (now
    // records searchMissingValue so run() exits 2 with "--search requires a
    // query"). The bool flags before it remain set.
    c.searchMissingValue = true
```
(c is `*config` — `c.searchMissingValue = true` is a valid assignment.)

**Trace (all stay correct):** `--search foo` → searchMode=true, searchQ="foo"
(value path unchanged). `--search`/`-s` (last) → searchMissingValue=true (NEW).
`-vs` (last) → version=true, searchMode=false, searchMissingValue=true (NEW).
`--search=` → searchMode=true, searchQ="" (unchanged, D5).

---

## §5 — The `--shell` no-value fix (mirrors --search, form 1b only)

parseargs_research.md §2: `--shell` has two forms; only the bare no-token case
(main switch, main.go:320-326) gets the fix. The `=`-form `--shell=` (main.go:248-253)
is UNCHANGED (D5 — no empty-value guard; `--shell=` → completion=true, shell="").

CURRENT main switch `case "--shell":` (main.go:320-326):
```go
if i+1 < len(args) {
    c.completion = true
    c.completionShell = args[i+1]
    i++
}
```
FIX — add the `else`:
```go
if i+1 < len(args) {
    c.completion = true
    c.completionShell = args[i+1]
    i++
} else {
    c.shellMissingValue = true
}
```
**Trace:** `--completions --shell bash` → completion=true, shell="bash" (value
path unchanged — shellMissingValue stays false). `--shell` (last) →
shellMissingValue=true (NEW). `--completions --shell` (last) → completion=true
(from --completions) + shellMissingValue=true → run() exits 2 (the missing-value
guard fires before completion dispatch). `--shell=` → completion=true, shell=""
(unchanged, D5).

---

## §6 — The two new run() checks (peer of storeMissingValue, exact placement + messages)

Insert AFTER the `storeMissingValue` check (main.go:502, before the exclusivity
comment at 508), mirroring its `fmt.Fprintln(stderr, ...); return 2` shape:

```go
if c.searchMissingValue {
    fmt.Fprintln(stderr, "skilldozer: --search requires a query")
    return 2
}
if c.shellMissingValue {
    fmt.Fprintln(stderr, "skilldozer: --shell requires a value (bash|zsh|fish)")
    return 2
}
```

**Exact messages** (contract LOGIC (e), verbatim): `skilldozer: --search requires
a query` and `skilldozer: --shell requires a value (bash|zsh|fish)`. The
`skilldozer:` prefix + `requires a …` phrasing mirror `skilldozer: --store requires
a value` (main.go:500). Use `fmt.Fprintln` (fixed string, no args — matches the
storeMissingValue branch). Do NOT print to stdout.

**Ordering is load-bearing:** these run AFTER help/version/unknownFlag/
storeMissingValue and BEFORE exclusivity/dispatch. So `--help --search` → exit 0
(help wins, same as `--help --store`); `--search` alone → exit 2. The four
missing-value guards (store/search/shell) are all peer parse-error checks at
step 3.5-3.7.

---

## §7 — The misleading comments to fix (Mode A / DOCS)

**(a) main.go:290-293** (the `--search`/`-s` comment). CURRENT ends:
```
// ... If --search is the LAST
// token (no value follows) searchMode stays false and the call falls
// through to the no-recognized-mode default (exit 1).
```
The `(exit 1)` is WRONG (actual: exit 0 today — parseargs_research §4). After the
fix it is exit 2. REPLACE the tail to: `... records searchMissingValue so run()
exits 2 with "--search requires a query" (mirrors --store, Issue 3).`

**(b) main.go:323-324** (the `--shell` comment). CURRENT: `If --shell is the LAST
token (no value), completion stays false — mirrors --search's no-value silent
behavior (PRD §6.4 specifies no missing-value exit code for --shell).` After the
fix this is no longer "silent". REPLACE to: `If --shell is the LAST token (no
value), records shellMissingValue so run() exits 2 with "--shell requires a value
(bash|zsh|fish)" (mirrors --store/--search, Issue 3).`

**(c) main.go:493-498** (the storeMissingValue precedent comment). Add a sentence
that --search/--shell now follow the same pattern (contract DOCS): e.g. append
`The same missing-value-exit-2 pattern is applied to --search (searchMissingValue)
and --shell (shellMissingValue) for symmetry (Issue 3).`

---

## §8 — Test churn: ONE rename+flip, ONE update, FIVE new (no run-level test locks the old exit-0)

grep-confirmed: NO run-level test asserts exit-0 for `--search`/`--shell` no-value.
The only existing tests touching these no-value paths are parseArgs-level:

**(A) RENAME + flip** `TestParseArgsSearchNoValueStaysInactive` (main_test.go:1004)
→ `TestParseArgsSearchMissingValue`. CURRENT asserts `searchMode` stays false (still
true after the fix — the `else` only sets searchMissingValue, NOT searchMode). The
test name ("StaysInactive") + doc comment ("falls to the default exit-1 path") are
now misleading. NEW body asserts BOTH `searchMissingValue == true` (the new
behavior) AND `searchMode == false` (regression guard — searchMode is genuinely not
set when there's no value):
```go
// Issue 3: `--search` (last token, no value) records searchMissingValue so run()
// exits 2 with "--search requires a query" (mirrors --store). searchMode stays
// false (no value consumed).
func TestParseArgsSearchMissingValue(t *testing.T) {
	c := parseArgs([]string{"--search"})
	if !c.searchMissingValue {
		t.Errorf("parseArgs(--search) no value: searchMissingValue=false; want true (Issue 3)")
	}
	if c.searchMode {
		t.Errorf("parseArgs(--search) no value: searchMode=true; want false (no value consumed)")
	}
}
```
(This IS contract OUTPUT test #4.)

**(B) UPDATE** `TestParseArgsShortBundleSearchNoValue` (main_test.go:2356). CURRENT
asserts `version=true` + `searchMode=false` (both still true after the fix — the
bundle default now ALSO sets searchMissingValue). ADD the new assertion so the test
covers the changed behavior:
```go
if !c.searchMissingValue {
	t.Errorf("-vs: searchMissingValue=false; want true (s had no value -> Issue 3 signal)")
}
```
(Keep the existing version=true + searchMode=false assertions; the test stays GREEN
either way, but adding the assertion locks the new bundle behavior.)

**(C) ADD 3 run-level exit-2 tests** (mirror `TestRunInitStoreNoValueExits2` @321 /
`TestRunStoreBareNoValueExits2` @350 — exact-stderr + empty-stdout + code==2):
- `TestRunSearchNoValueExits2`: `run([]string{"--search"})` → code 2, stdout empty,
  stderr == `"skilldozer: --search requires a query\n"`.
- `TestRunSearchShortNoValueExits2`: `run([]string{"-s"})` → code 2, stdout empty,
  stderr == `"skilldozer: --search requires a query\n"` (bare -s via main switch).
- `TestRunShellNoValueExits2`: `run([]string{"--shell"})` → code 2, stdout empty,
  stderr == `"skilldozer: --shell requires a value (bash|zsh|fish)\n"`.

**(D) ADD 1 parseArgs-level test**:
- `TestParseArgsShellMissingValue`: `parseArgs([]string{"--shell"}).shellMissingValue == true`
  (and completion stays false — regression guard).

NOTE (D5): the `=`-form tests `TestParseArgsLongEqualsSearchEmpty` (@2294) and any
`--shell=` behavior stay UNCHANGED and GREEN — the =-form gets NO missing-value
guard. Do NOT add an `=`-form no-value test.

---

## §9 — No conflict with the parallel sibling P1.M1.T2.S1 (disjoint files)

P1.M1.T2.S1 (Issue 2) edits `internal/skillsdir/skillsdir.go` +
`internal/skillsdir/skillsdir_test.go` (the vanished-store sentinel + 4-value
findConfig). It does NOT touch `main.go` or `main_test.go` (its PRP: "Does NOT
touch main.go"). This task edits ONLY `main.go` + `main_test.go`. **Disjoint
files; no merge collision; land in either order.** (Plan ordering: P1.M1 before
P1.M2, so Issue 2 lands first — but even concurrent, the files don't overlap.)

---

## §10 — Scope discipline + zero deps

- Do NOT touch the `=`-form switch (main.go:218-220 `--search=`, 248-253 `--shell=`,
  229 `--store=`, 237 `--init=`) — D5 leaves empty `=`-values as-is. Only the bare
  no-token branches (main switch + bundle) get the fix.
- Do NOT touch `internal/skillsdir/*` (P1.M1.T2.S1), `completions/*`, `README.md`
  (P1.M3.T1.S1), or `PRD.md`/`tasks.json`/`prd_snapshot.md`/`.gitignore`.
- Do NOT add deps/imports. `fmt.Fprintln` is already imported in main.go; the new
  fields are bools; `expandShortBundle` already takes `c *config`. go.mod/go.sum
  byte-for-byte unchanged. Verify with `git diff --quiet go.mod go.sum`.
- Do NOT change any exit code OTHER than the new missing-value → exit-2 paths. The
  success paths (`--search foo`, `--completions --shell bash`) are unchanged.

---

## §11 — Validation (Go toolchain, verified commands)

```bash
gofmt -l main.go main_test.go   # must print NOTHING
go vet ./...                    # exit 0
go build ./...                  # exit 0
go test -run 'SearchMissingValue|ShellMissingValue|SearchNoValueExits2|SearchShortNoValueExits2|ShellNoValueExits2|ShortBundleSearchNoValue' -v ./...
go test ./...                   # whole module green; zero regressions (the =-form tests + success tests unchanged)
git diff --quiet go.mod go.sum && echo deps unchanged
# Manual repro (now FIXED):
go build -o /tmp/sdz . && for f in --search -s --shell; do /tmp/sdz $f >/dev/null 2>&1; echo "$f exit=$? (want 2)"; done; rm -f /tmp/sdz
```
