# Verified Facts — P1.M2.T1.S1: init parsing, --store flag, USAGE row, exclusivity family

Every claim below was read directly from the live source at
`/home/dustin/projects/skilldozer` (main.go + main_test.go read in full; line
numbers are the CURRENT state — the parallel sibling P1.M1.T2.S2 does NOT touch
main.go, confirmed by reading its PRP's "main.go UNCHANGED" invariant). Module
path `github.com/dabstractor/skilldozer`, `go 1.25`, single third-party dep
`gopkg.in/yaml.v3` (NOT used by this task — pure argv parsing, zero new deps).

---

## §1 — The exact edit anchors in main.go (CURRENT line numbers, sibling-safe)

The parallel sibling P1.M1.T2.S2 edits `internal/skillsdir/*` + the 6 unresolvable
main_test.go tests + flips `ErrNotFound`. It does NOT edit `main.go` and does NOT
edit the parseArgs/exclusivity/USAGE regions of main_test.go. So every line
number below is stable when this subtask's implementation begins.

```
main.go:50    const usageText = `skilldozer — skill path printer`   (the help block; USAGE @~54, EXAMPLES @~64, OPTIONS @~74)
main.go:122   type config struct { ... }                              (12 fields; ADD init + initStore here)
main.go:145   func parseArgs(args []string) config {
main.go:159   (a) the '='-form switch (HasPrefix("--") && Contains("=")) — ADD case "--store" here (mirror --search @181)
main.go:196   the main token switch — ADD case "--store" (after --search @222) + case "init" (after check @234)
main.go:222   case "--search", "-s":  (the value-capture PATTERN to mirror for --store)
main.go:234   case "check":           (the subcommand PATTERN to mirror for init)
main.go:244   default:                (tags/unknownFlag branch — init currently falls here as a tag; the new case "init" steals it)
main.go:367   func run(...)           (precedence: help→version→unknownFlag→exclusivity→dispatch; NO init dispatch here — that is P1.M2.T2.S3)
main.go:397   if bad, msg := exclusivityError(c); bad { ... return 2 }   (exclusivity runs BEFORE dispatch + BEFORE Find())
main.go:635   func exclusivityError(c config) (bad bool, msg string)     (ADD the init family here; hasTags defined @650)
main.go:683   func skillPath(...)     (UNTOUCHED)
```

`main.go` has ZERO current references to `init` / `--store` / `initStore` / `c.init`
(verified by `grep -n 'init\|--store\|initStore\|c\.init\b' main.go` → empty).
So this is a purely additive change; nothing is renamed or removed.

---

## §2 — The `check` subcommand is the exact pattern to mirror for `init`

`case "check":` (main.go:234-243) is a RESERVED positional token: it sets `c.check =
true` and is NOT captured as a tag (the default branch is what would otherwise
grab it). The tests that lock this (main_test.go:1117-1155) are the template for
the init parse tests:

- `TestParseArgsCheckSubcommand`        → `parseArgs(["check"])` sets check, tags empty.
- `TestParseArgsCheckAfterFlag`         → `["--no-color","check"]` sets both.
- `TestParseArgsCheckAndTagBothCaptured`→ `["check","sometag"]` sets check AND tags=[sometag]
  (run() later rejects via exclusivity; parseArgs captures both).

`init` must behave the SAME way for the bare token (`init` alone → c.init, no
tags). The ONLY difference: `init` optionally swallows ONE following positional
as the store dir (`init <dir>`), which `check` does not. See §4.

---

## §3 — The `--search` value-capture is the exact pattern to mirror for `--store`

`--store` must take a value two ways, exactly like `--search`:

(a) `=`-form (`--store=/tmp/x`), handled in the '='-form switch (main.go:159-189).
    The existing `--search` case there (main.go:181-183):
        case "--search":
            c.searchMode = true
            c.searchQ = val
    → mirror with:
        case "--store":
            c.init = true
            c.initStore = val

(b) long-form (`--store /tmp/x`), handled in the main token switch. The existing
    `--search`/`-s` case (main.go:222-232):
        case "--search", "-s":
            if i+1 < len(args) {
                c.searchMode = true
                c.searchQ = args[i+1]
                i++
            }
    → mirror with (NOTE: --store has NO short form; PRD §6.2 defines none):
        case "--store":
            if i+1 < len(args) {
                c.init = true
                c.initStore = args[i+1]
                i++
            }

KEY (decided): `--store` SETS `c.init = true` unconditionally (when a value is
present). This makes `skilldozer --store <dir>` equivalent to `skilldozer init
--store <dir>` — which is exactly what the contract OUTPUT §4 demands ("`skilldozer
--store <dir>` parse without being treated as tags"). The contract LOGIC (c)
phrase "if seen without init it is an unknown/incompatible flag (exit 2)" is
SATISFIED by the init exclusivity family: --store implies init, so any OTHER mode
combined with --store trips init+mode → exit 2. No separate "store-without-init"
branch is needed. (OUTPUT is authoritative over the ambiguous LOGIC sentence.)

MIRROR DETAIL (no-value case): `--search` with no following value leaves
searchMode=false (main_test.go:897 `TestParseArgsSearchNoValueStaysInactive`).
`--store` with no value mirrors this: leaves c.init=false, initStore="" → falls to
the no-mode default (exit 1). Do NOT invent an exit-2 "needs argument" here; the
codebase deliberately defers that (see the `--search` no-value comment at
main.go:227-231).

---

## §4 — THE GOTCHA: `case "init":` positional capture must NOT swallow reserved subcommands

This is the single most likely one-pass failure. The naive mirror of --search
(`if i+1 < len(args) { c.initStore = args[i+1]; i++ }`) is WRONG for init, because
`init <dir>` must capture a DIRECTORY but must NOT swallow another subcommand:

  `init check`  — MUST reach `case "check":` so c.check=true, then exclusivityError
                  flags init+check → exit 2 (PRD §6.3 / contract LOGIC (e)).
  `init --store x` — MUST NOT swallow "--store"; the --store case fills initStore.
  `init /tmp/x` — MUST swallow "/tmp/x" as initStore (contract test case).
  `init foo bar` — foo → initStore; bar → tags (default branch); exclusivity
                   init+tags → exit 2 (contract LOGIC (e): "stray tag after init").

RESOLUTION (verified against all contract test cases): in `case "init":`, peek
args[i+1] and capture it as initStore ONLY if it is NEITHER a dashed flag NOR a
reserved subcommand token. The reserved set is exactly {"check", "init"} (the only
positional subcommands). Concretely:

    case "init":
        c.init = true
        if i+1 < len(args) {
            next := args[i+1]
            if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
                c.initStore = next
                i++
            }
        }

Trace (all contract test cases pass):
  ["init"]            → i+1 not < len → no peek. c.init=true, initStore="".            ✓ (OUTPUT: "init sets c.init and no tags")
  ["init","/tmp/x"]   → next="/tmp/x" (not flag, not reserved) → initStore="/tmp/x".     ✓ (OUTPUT: "init /tmp/x sets initStore")
  ["init","--store","/tmp/x"] → next="--store" (flag) → skip; --store case fills initStore="/tmp/x". ✓
  ["init","--list"]   → next="--list" (flag) → skip; --list sets c.list; exclusivity init+list → exit 2. ✓
  ["init","--path"]   → next="--path" (flag) → skip; --path sets c.path; exclusivity init+path → exit 2. ✓
  ["init","check"]    → next="check" (reserved) → skip; check sets c.check; exclusivity init+check → exit 2. ✓ (LOGIC (e))
  ["init","foo","bar"]→ foo (positional) → initStore="foo", i++; bar → tags=["bar"]; exclusivity init+tags → exit 2. ✓ (LOGIC (e) stray tag)

EDGE CASE (documented, accepted): a store directory literally named `check` or
`init` cannot be passed via the `init <dir>` positional form (use `--store
./check` or an absolute path). PRD §8.2 does not contemplate this; pathological.

GOTCHA for future maintainers: adding a NEW positional subcommand requires adding
it to this reserved guard, or it would be swallowed as a store dir after `init`.

---

## §5 — exclusivityError: where the init family goes, and why it cannot collide

Current exclusivityError (main.go:635-661) has 4 families, checked in order:
  1. ≥2 listing modes among {path,list,searchMode,all} (the `n >= 2` count)
  2. tags + a listing mode (list/search/all)
  3. check + tags
  4. check + a listing mode (path/list/search/all)

The init family is a PEER of the check families (PRD §6.3 / §8.2: "init is its own
exclusive mode like check"). `init` is NOT in the listing-mode count (family 1) —
that count is exactly {path,list,search,all} and must stay that way, so
`init <single-mode>` is caught by the init family, NOT masked by family 1.

`hasTags` is defined at main.go:650 (`hasTags := len(c.tags) > 0`) BEFORE the
check families, so the init family (placed after family 4, before `return
false,""`) can reuse it. Placement AFTER family 4 is correct: a 2+-listing-mode
combo still gets the precise "listing modes" message (family 1), and check's
families keep their messages; init only needs to catch init+{one mode or tags}.

No collision: init+2-listing-modes (e.g. `init --path --list`) is caught by family
1 first (exit 2, correct message). init+single-mode is caught by the init family.
init+check is caught by the init family (c.check is in its mode set). init+tags is
caught by the init family (stray tag). All return exit 2.

Message wording (mirror the existing `skilldozer: '<cmd>' cannot be combined with`
convention at main.go:655, 658):
  init+tags  → "skilldozer: 'init' cannot be combined with tag arguments"
  init+mode  → "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"

---

## §6 — run() precedence: exclusivity runs BEFORE dispatch and BEFORE Find()

run() (main.go:367-661) order:
  1. help   (return 0)
  2. version (return 0)
  3. unknownFlag (return 2)
  4. exclusivityError (return 2)   ← init --list / init --path / init check exit HERE
  5. dispatch: path → list → search → check → all → tags → no-mode-default

CONSEQUENCE (load-bearing for the tests): the init EXCLUSIVITY tests
(`run(["init","--list"])` etc.) return 2 from step 4 and NEVER call
skillsdir.Find() or discover.Index(). They therefore need NO store fixture, NO
SKILLDOZER_SKILLS_DIR, NO t.Chdir, NO unsetSkillsEnv. They are pure argv → exit-2
checks. (Contrast: the init SUCCESS-flow tests are parseArgs-level only — see §7.)

This subtask does NOT add an `if c.init { … }` dispatch branch. That is explicitly
P1.M2.T2.S3 ("run() reads c.init/c.initStore"). So after this subtask,
`skilldozer init` (no conflict) parses correctly (c.init=true) but falls through
dispatch to the no-mode default (usage to stderr, exit 1). That is EXPECTED and
NOT a failure of this subtask — the success dispatch lands in the next subtask.

---

## §7 — Test split: parseArgs-level for success, run-level for exclusivity

Because dispatch is deferred (§6), the init tests split cleanly:

SUCCESS cases → parseArgs-level (assert fields, no run()):
  parseArgs(["init"])                → c.init=true, len(c.tags)==0, initStore==""
  parseArgs(["init","/tmp/x"])       → c.init=true, initStore=="/tmp/x", len(c.tags)==0
  parseArgs(["init","--store","/tmp/x"]) → c.init=true, initStore=="/tmp/x"
  parseArgs(["init","--store=/tmp/x"])   → c.init=true, initStore=="/tmp/x"
  parseArgs(["--store","/tmp/x"])    → c.init=true, initStore=="/tmp/x" (no init token; --store implies init)
  parseArgs(["init","/tmp/x"]) extra → c.tags EMPTY (the dir is NOT a tag)

EXCLUSIVITY cases → run-level (assert exit 2 + empty stdout + stderr msg):
  run(["init","--list"])    → 2
  run(["init","--path"])    → 2
  run(["init","check"])     → 2
  run(["init","--search","q"]) → 2
  run(["init","--all"])     → 2
  run(["init","foo","bar"]) → 2   (stray tag: foo→initStore, bar→tags)

USAGE case → run-level --help (assert substrings):
  run(["--help"]) stdout contains "skilldozer init" AND "--store"

DO NOT write a run-level init SUCCESS test (e.g. `run(["init"]) == 0`). It would
FAIL today (no dispatch → exit 1) and is owned by P1.M2.T2.S3.

---

## §8 — No existing test breaks; no skillsdir/main.go rename churn

grep confirms main_test.go references "init" ONLY in the ErrNotFound message
assertions (main_test.go:244,383,597,855,1095,1273 — all the parallel sibling's
"skilldozer init" substrings). No existing test passes "init" or "--store" as
argv. No existing test asserts a USAGE substring that this subtask removes. So
this change is purely ADDITIVE to the test suite (new tests) + additive to
main.go (new cases/fields/USAGE lines/family).

main.go and main_test.go are the ONLY files touched. internal/* is untouched.
go.mod/go.sum untouched (zero new deps — pure stdlib `strings` which is already
imported).

---

## §9 — USAGE block edits (PRD §6.1 / contract LOGIC (d)) — exact text

The current usageText (main.go:50-98) has three blocks. Edits (mirror the column
alignment of the existing rows; the OPTIONS column is aligned to the longest
existing entry — keep `init [<dir>]` and `--store <dir>` aligned with neighbors):

USAGE block — add after the `skilldozer check` line (PRD §6.1 table order is
  tag, --all, --list, --search, check, init, --path, --help, — so init sits
  between check and --path):
    skilldozer init [<dir>]

EXAMPLES block — add one line (e.g. after the `skilldozer check` example):
    skilldozer init --store <dir>     # non-interactive first-run setup

OPTIONS block — add two lines (after the `check` option line):
    init [<dir>]      First-run setup: pick/create the skills store and write the config
    --store <dir>     Non-interactive store path for init

These additions do NOT remove any token that TestRunHelpToStdoutExit0
(main_test.go:1361-1378) asserts ({USAGE:, EXAMPLES:, OPTIONS:,
`pi --skill "$(skilldozer example)"`}), so that test stays green. A NEW test
asserts the init row + --store line are present.

---

## §10 — Dependency on the parallel sibling (P1.M1.T2.S2): NONE for main.go

The sibling edits internal/skillsdir/skillsdir.go (+skillsdir_test.go) and the 6
unresolvable main_test.go tests + hardens unsetSkillsEnv. It does NOT edit:
  - main.go (verified: sibling PRP states "main.go UNCHANGED")
  - the parseArgs/exclusivity/USAGE regions
  - any check-tag/check-mode/init test

So when this subtask begins implementation, main.go is in the state read here and
the sibling's main_test.go edits (the 6 "skilldozer init" message flips +
hardened unsetSkillsEnv) have LANDED (plan ordering: P1.M1 completes before
P1.M2). The init tests added here are NEW functions; they do not overlap the
sibling's edited functions, so there is no text-level merge collision.

unsetSkillsEnv (main_test.go:25) is hardened by the sibling to neutralize the
config rule. The init EXCLUSIVITY tests do NOT need it (they exit before Find()),
and the init parseArgs tests do NOT need it (parseArgs never touches the
filesystem). So whether or not the sibling has landed, the init tests are
hermetic as written. Belt-and-suspenders: the exclusivity tests can call
unsetSkillsEnv anyway (harmless), but it is not required.
