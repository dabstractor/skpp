# Verified Facts — P1.M1.T1.S1 (parseArgs: subcommands → flags)

Confirmed against the live source (`main.go` @ HEAD `55ada20`; 1288 lines; architecture
map verified against `f30d5c5` — line numbers match, since `55ada20` only added plan
docs). This is decision 19 / PRD §6.3: there are NO bare-word subcommands —
`check`/`init`/`completion` become `--check`/`--init`/`--completions`, and bare
`check`/`init`/`completion`/`completions` tokens fall through to `c.tags` (they are now
skill tags). Baseline: all tests green today (bare subcommands still work).

## 1. parseArgs structure (main.go:179-369) — the THREE bare cases to DELETE

The main `switch a {` (starts ~line 252) currently has bare-word cases. Verified exact
sites (live grep):

- **`case "check":` — main.go:284-293.** Body is just `c.check = true` behind a long
  "RESERVED positional token" comment. DELETE the case + comment; replace with
  `case "--check":` → `c.check = true`.
- **`case "completion":` — main.go:294-299.** Body `c.completion = true` + "RESERVED"
  comment. DELETE; replace with `case "--completions":` → `c.completion = true`.
  (NOTE: PLURAL `--completions` per decision 19; the old subcommand was singular
  `completion`.)
- **`case "init":` — main.go:324-351.** The ENTIRE block, INCLUDING all Issue-4
  duplicate-init logic (`next == "init"` → append to tags) and the
  `next != "check" && next != "completion"` special-casing. DELETE it all; replace with
  `case "--init":` → `c.init = true` + a simple non-dashed-next-token capture into
  `c.initStore` (no Issue-4 logic — see §4).

## 2. The new flag cases (main switch) — exact target code

In-place conversion (replace each deleted bare case with its flag form at the same spot):

```go
case "--check":
    // `skilldozer --check` flag (PRD §9). A bare `check` is now a skill TAG
    // (decision 19 / §6.3: no bare-word subcommands), not this flag.
    c.check = true
case "--completions":
    // `skilldozer --completions [--shell <name>]` flag (PRD §14.6). PLURAL (decision 19).
    // A bare `completion`/`completions` is a skill tag, not this flag.
    c.completion = true
case "--init":
    // `skilldozer --init [<dir>]` first-run setup (PRD §8.2). --init is the sole mode
    // that accepts a positional <dir> (the store to adopt, §6.3); a following NON-dashed
    // token is captured as c.initStore. A dashed follower (--init --store …) is left for
    // its own case. A bare `init` is a skill tag (decision 19).
    c.init = true
    if i+1 < len(args) {
        next := args[i+1]
        if !strings.HasPrefix(next, "-") {
            c.initStore = next
            i++
        }
    }
```

Switch cases are order-independent (non-overlapping), so exact placement among the
existing cases is free; replacing each bare case in spot = minimal diff.

## 3. The `=`-form switch (main.go:196-234) — add `--init=` (and, for consistency, --check/--completions)

The contract (c) mandates adding `--init=` after `case "--store":` (lines 218-227):
```go
case "--init":
    // `--init=<dir>`: non-interactive store path (mirrors --store=). Empty value
    // (--init=) records storeMissingValue so run() rejects with exit 2 (Issue 2).
    c.init = true
    c.initStore = val
    if val == "" {
        c.storeMissingValue = true
    }
```
CONSISTENCY GOTCHA: every existing BOOL flag is ALSO in this `=`-form switch (it ignores
the value: `--version=x` == `--version`). See lines 197-205: --version/--help/--path/
--list/--all/--file/--relative/--no-color are all there. For consistency, ALSO add
`--check` and `--completions` here (bool → ignore value), so `--check=foo` behaves like
`--check` instead of falling to the `default:` → unknownFlag → exit 2. This matches the
established convention and prevents a subtle inconsistency. (--init is the only one of
the three that takes a value, like --store.)

## 4. Why the Issue-4 duplicate-init logic is GONE (and that's correct)

The old `case "init":` had special handling so `init init` (a duplicate) was captured
into `c.tags` so exclusivity could reject `init init` (Issue 4). With flags this is no
longer needed:
- `--init --init` → both are dashed; the second sets `c.init=true` again (idempotent). No
  tag captured. exclusivity does not flag it (same flag twice is harmless, like `--path --path`).
- `--init init` → `init` is non-dashed → captured as `c.initStore` (a store dir literally
  named "init"). That's the documented `--init <dir>` form (§6.3: --init accepts a
  positional <dir>). Correct.
- `--init check` / `--init completion` → `check`/`completion` are non-dashed → captured
  as `c.initStore`. (A store named "check"/"completion"; if the user meant a tag, that's
  the §6.3 tradeoff — --init owns the next positional.)

So the new --init case is SIMPLER (no `next == "init"` / `next != "check"` branches).
The contract explicitly says delete all that. exclusivity (init+tags) is enforced later
by exclusivityError — that's S2, unchanged logic.

## 5. The `default:` branch (main.go:353-368) — NO CHANGE

```go
default:
    if strings.HasPrefix(a, "-") {
        if c.unknownFlag == "" { c.unknownFlag = a }   // dashed unknown → exit 2
    } else {
        c.tags = append(c.tags, a)                      // bare word → tag
    }
```
After deleting the bare cases, `check`/`init`/`completion`/`completions` tokens land here
→ `c.tags`. That IS the namespace-safety guarantee (decision 19). No edit needed.

## 6. DO NOT touch --store / --shell wiring (they imply init / completion)

- `=`-form `case "--store":` (218-227) and token `case "--store":` (300-313): set
  `c.init=true` + `c.initStore`. UNCHANGED.
- `=`-form `case "--shell":` (228-233) and token `case "--shell":` (314-323): set
  `c.completion=true` + `c.completionShell`. UNCHANGED.
`--store <dir>` still implies `--init`; `--shell <name>` still implies `--completions`.
The contract (e) forbids changing this.

## 7. Doc-comment rewrites in S1's territory (config struct + package doc)

Exact current lines (live):
- **main.go:10** (package doc): "subcommands like `check`, positional <tag> args, …" →
  "flags like `--check`, positional <tag> args, …".
- **main.go:162** `check`: "`skilldozer check` subcommand: validate every skill in the
  store (§9)" → "`skilldozer --check` flag: validate every skill in the store (§9)".
- **main.go:163** `init`: "`skilldozer init [<dir>]` first-run setup …" → "`skilldozer
  --init [<dir>]` first-run setup …".
- **main.go:164** `initStore`: "`init <dir>` positional or `--store <dir>` …" → "`--init
  <dir>` flag or `--store <dir>` …".
- **main.go:166** `completion`: "`skilldozer completion` subcommand (§14.6)…" →
  "`skilldozer --completions` flag (§14.6)…".
The deleted bare-case comments (285/288, 295, 325/328/330/340) vanish with the cases.

## 8. SCOPE BOUNDARY — "reserved"/"subcommand" sites that belong to SIBLINGS (do NOT edit)

The grep found "reserved"/"subcommand" comments outside parseArgs/config/package. S1
MUST LEAVE these to avoid colliding with sibling tasks that own those functions:
- **main.go:631** — runCheck doc comment ("`skilldozer check` subcommand …"). Owned by
  **P1.M1.T2.S1** (doc-comments sweep).
- **main.go:821** — exclusivityError comment ("modes/subcommands AND stray tags …").
  Owned by **P1.M1.T1.S2** (exclusivityError rewrite — messages + comments together).
- **main.go:1191** — runInit comment ("`check` subcommand keeps its report on stdout …").
  Owned by **P1.M1.T2.S1** (doc-comments sweep).
- **main.go:506** — run() storeMissingValue comment (not actually reserved/subcommand;
  grep artifact) — run() dispatch, out of scope.
S1's (h) "remove all reserved/subcommand comments" is satisfied WITHIN S1's territory
(parseArgs + config + package). S2 cleans exclusivityError; T2 cleans the function doc
comments. Editing 821 during S1 would collide with S2.

## 9. Dispatch is UNCHANGED — config fields still drive the same functions

`run()` dispatches on `c.check`→runCheck, `c.init`→runInit, `c.completion`→runCompletion.
S1 only changes HOW those fields get SET (bare tokens → flags), not the dispatch. No
run() edit in this subtask.

## 10. Tests WILL FAIL after S1 — that is EXPECTED (TDD), NOT a gate

The contract OUTPUT §4: "go build ./... must succeed. go vet ./... must pass. Tests will
FAIL until P1.M1.T3 — that is expected (TDD)." Tests like TestParseArgsCheckSubcommand
assert `parseArgs([]string{"check"}).check == true`; after S1 that's `check`→tags,
`c.check==false` → test fails. Flipping those tests is **P1.M1.T3**'s job. So S1's
validation gate is **build + vet only**; `go test` is EXPECTED red. Do NOT "fix" tests
in S1 (that is T3's scope and would be a scope violation).

## 11. Deps / build

`go.mod`: module `github.com/dabstractor/skilldozer`, go 1.25, sole dep
`gopkg.in/yaml.v3 v3.0.1`. S1 changes no imports (strings.HasPrefix, args indexing all
already in parseArgs). go.mod/go.sum byte-for-byte unchanged.
