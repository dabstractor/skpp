# Verified Facts — P1.M2.T1.S1: `completion` token + `--shell` flag + config fields + USAGE row + exclusivity family

Plan `003_3ace946c2a5c` (the `completion` subcommand delta). Every claim below was
read directly from the live source at `/home/dustin/projects/skilldozer` (main.go
config struct / parseArgs / exclusivityError / usageText read in full at the cited
ranges; main_test.go check+init parse/exclusivity/help tests read at exact line
ranges). Module `github.com/dabstractor/skilldozer`, `go 1.25`, sole third-party
dep `gopkg.in/yaml.v3` (NOT touched by this task — pure argv parsing). Zero new
deps.

---

## §1 — Exact edit anchors (CURRENT main.go line numbers, sibling-safe)

The parallel sibling P1.M1.T1.S1 (no-args implicit-help flip) edits the usageText
DOC comment (main.go:48-51), the run() exit-code doc (417-423), the no-mode
fallthrough (695-700), and two tests (TestRunDefaultNoArgs, TestRunModifiersOnlyNoMode).
It does NOT touch any region this subtask edits (confirmed by reading its PRP).
Plan ordering lands P1.M1 before P1.M2, so its changes are present when this
subtask begins. Every anchor below is therefore stable.

```
main.go:52    const usageText = `...`            (the help block: USAGE @~56, EXAMPLES @~66, OPTIONS @~78)
main.go:128   type config struct { ... }         (ADD completion + completionShell here, ~after storeMissingValue)
main.go:188   the '='-form switch (HasPrefix("--") && Contains("=")) — ADD case "--shell" after case "--store" (@~203)
main.go:220   the main token switch              — ADD case "completion" (after case "check" @253) + case "--shell" (after case "--store" @263)
main.go:253   case "check":                      (the SIMPLEST-reserved-token PATTERN to mirror for completion)
main.go:263   case "--store":                    (the value-taking PATTERN to mirror for --shell: if i+1<len{...;i++}else{...})
main.go:290   case "init": (the positional-capture GUARD to extend — see §4)
main.go:722   func exclusivityError(...)         (ADD the completion block AFTER the init block @~761, before `return false, ""`)
main.go:761   the init block `if c.init { ... }` (the PATTERN to mirror for the completion block)
```

`grep -n 'completion\|--shell\|completionShell\|c\.completion' main.go main_test.go`
returns ZERO hits today — this change is purely additive; nothing is renamed/removed.

---

## §2 — The `--shell` flag design (mirrors `--store`: it IMPLIES completion)

PRD §14.6 / contract LOGIC (b)+(c) fix `--shell` as a **long-only** value-taking
flag that **implies completion mode** (exactly as `--store` implies `init`):

- **=-form** (`--shell=bash`), in the `=`-form switch (main.go:188-204), added right
  after `case "--store":`. Per contract LOGIC (b), set BOTH fields unconditionally:
  ```go
  case "--shell":
      c.completion = true
      c.completionShell = val
  ```
  NO empty-value guard (do NOT mirror --store's `if val == "" { c.storeMissingValue = true }`).
  Mirror --search's =-form instead (unconditional `c.searchMode = true; c.searchQ = val`).
  So `--shell=` (empty) → completion=true, completionShell="" (a T2.S2 detection concern,
  not this subtask's). Rationale: PRD §6.4 specifies NO missing-value exit code for --shell
  (unlike --store, which has one for the non-destructive contract).

- **long-form** (`--shell bash`), in the main token switch, added right after
  `case "--store":`. Mirror --store's next-token capture (main.go:263-274) — but with
  NO else-branch guard (mirror --search's silent no-value):
  ```go
  case "--shell":
      if i+1 < len(args) {
          c.completion = true
          c.completionShell = args[i+1]
          i++
      }
      // else: --shell with no value -> silent no-op (completion stays false),
      //   mirrors --search no-value; falls to no-mode (implicit help). PRD §6.4
      //   specifies no missing-value exit code for --shell.
  ```

- **no short form.** PRD §6.2/§14.6 define no short alias for --shell. Do NOT touch
  `expandShortBundle` (its validated char set stays `v h p l a f s` — `s` is search).

Trace (all contract test cases):
  `completion --shell bash`   → completion case sets c.completion; --shell consumes "bash" → completionShell="bash". ✓
  `completion --shell=bash`   → completion case; --shell=-form → completionShell="bash". ✓
  `--shell bash` (no completion token) → --shell sets completion=true, completionShell="bash" (implies completion). ✓
  `--shell` (last token, no value) → silent no-op → no-mode (implicit help). ✓ (mirrors --search)

---

## §3 — `case "completion":` mirrors `case "check":` (SIMPLEST reserved token)

Contract LOGIC (c): "add `case "completion": c.completion = true` (mirror
`case "check":` at :253 — simplest reserved token)." So completion is a plain bool
setter with NO positional capture (unlike `init`, which captures a following
`<dir>`). It is captured ANYWHERE in argv (so `--no-color completion` works), and
a skill literally tagged `completion` cannot be resolved via the bare token
(same reservation as check/init — PRD §6.3).

Place `case "completion":` right AFTER `case "check":` (group the reserved-token
subcommands). Place `case "--shell":` right AFTER `case "--store":` (group the
value-taking flags). Exact-match switch order does not affect behavior; grouping is
for readability.

---

## §4 — GOTCHA A (REQUIRED for §6.3): extend the `init` positional-capture guard so `init completion` exits 2

This is the #1 cross-cutting trap. The CURRENT `init` case (main.go:290+) captures a
following positional into `c.initStore` unless it is a dashed flag, the literal
`init`, or the literal `check`:

```go
case "init":
    c.init = true
    if i+1 < len(args) {
        next := args[i+1]
        if next == "init" {
            c.tags = append(c.tags, next); i++   // Issue 4: duplicate init -> tag -> init+tags exit 2
        } else if !strings.HasPrefix(next, "-") && next != "check" {
            c.initStore = next; i++               // the `init <dir>` positional form
        }
        // else: dashed flag or "check" -> left for its own case
    }
```

**The bug this subtask would introduce if untreated:** `init completion` — the guard's
`else if` condition (`!HasPrefix("-") && next != "check"`) is TRUE for `"completion"`,
so `init` SWALLOWS `"completion"` as `initStore="completion"` and advances past it.
The `completion` token NEVER reaches `case "completion":`, so `c.completion` stays
FALSE → the completion exclusivity family never fires → `init completion` silently
runs `init` with a store literally named "completion". That VIOLATES PRD §6.3
("completion is mutually exclusive with tags and other modes") and is inconsistent
with `init check` (which exits 2 because "check" is excluded → reaches its case →
`c.check` → the init family's `c.check || ...` catches it) and `init init` (exit 2
via the Issue-4 tag-capture).

**THE FIX (one-condition extension, mirrors the existing `check` exclusion):** add
`&& next != "completion"` to the `else if`:
```go
} else if !strings.HasPrefix(next, "-") && next != "check" && next != "completion" {
    c.initStore = next; i++
}
```
Then `init completion` → init does NOT swallow "completion" → next iteration hits
`case "completion":` → `c.completion=true` → the completion family's
`if c.check || c.init || ...` catches `c.completion && c.init` → exit 2. ✓ Consistent
with `init check` and `init init`. (No need for the Issue-4 tag-capture trick here:
unlike a duplicate `init` (which is idempotent and would otherwise dispatch),
`completion` reaching its own case sets a DISTINCT flag `c.completion`, which the
completion family checks against `c.init`.)

Trace after the fix:
  `init completion` → c.init=true; next="completion" excluded → falls to case "completion" → c.completion=true → completion family (c.completion && c.init) → exit 2. ✓
  `init check`      → unchanged (next="check" excluded → case "check" → c.check → init family → exit 2). ✓
  `init init`       → unchanged (next=="init" → tag → init+tags → exit 2). ✓
  `init /tmp/x`     → unchanged (positional → initStore). ✓
  `init --store x`  → unchanged (dashed → --store case). ✓

This makes the reserved-token set fully symmetric: init + {check, init, completion}
all exit 2. It is REQUIRED for PRD §6.3 compliance, not optional, even though the
contract's literal test list (OUTPUT §4) does not name `init completion`. Add a
test `TestRunExclusivityInitAndCompletion` → exit 2.

**Reverse direction (`completion init`) needs no change:** `case "completion":` does
NOT do positional capture (it mirrors `check`), so `completion init` → c.completion
(true), then `init` → c.init (true) → completion family catches c.completion && c.init
→ exit 2. ✓

---

## §5 — GOTCHA B: `completion completion` is idempotent like `check check` (the contract's "duplicate→tag" claim is WRONG)

The contract LOGIC (d) parenthetical says the completion block catches
"`completion completion` (duplicate→tag→hasTags)". That reasoning is **incorrect**:
because `case "completion":` is a recognized case (mirroring `case "check":`), the
SECOND `completion` token hits `case "completion":` again — NOT the default branch —
so it is NOT captured as a tag. `c.completion` is simply set true twice (idempotent),
`hasTags` stays false, and the completion family does NOT fire.

This matches the EXISTING precedent: `check check` does NOT exit 2 today
(check is idempotent; the check family needs `hasTags` or a second MODE). Only
`init init` exits 2 (via the special Issue-4 tag-capture), because a duplicate `init`
would otherwise be a silent idempotent dispatch with a config write.

**Guidance:** do NOT add duplicate-handling for `completion` (that would diverge from
`check`'s behavior and is out of scope). `completion completion` dispatches normally
(in T2.S2) — same as `check check`. Do NOT write a test asserting `completion completion`
exits 2 (it would fail). The contract's parenthetical is mistaken reasoning; the
OUTPUT test list (which does NOT include `completion completion`) is authoritative.

Likewise, `check completion` (a contract test case) exits 2 NOT because "the second
'completion' hits default→tag" (the contract's words) but because `completion` reaches
its own case → `c.completion=true` → the completion family's `c.check || ...` catches
`c.completion && c.check`. Same OUTCOME (exit 2), correct mechanism.

---

## §6 — exclusivityError: the completion block (mirror the init block)

Current exclusivityError (main.go:722-770) families, in order: (1) ≥2 listing modes
{path,list,search,all}; (2) tags + a listing mode; (3) check+tags; (4) check+mode;
(5) init+tags / init+mode (the init block @~752-761). `hasTags` is defined at
main.go:733, in scope for a later block.

Add the completion block AFTER the init block, BEFORE `return false, ""`. It mirrors
the init block but its mode set includes `c.init` (so init+completion is caught here,
and check+completion is caught here too):

```go
// completion is its own exclusive mode (PRD §6.3 / §14.6: like check/init). It
// rejects the other modes/subcommands AND stray tags. `completion` does no
// positional capture (mirrors check), so any positional after it lands in c.tags
// and is rejected here as a stray.
if c.completion {
    if hasTags {
        return true, "skilldozer: 'completion' cannot be combined with tag arguments"
    }
    if c.check || c.init || c.list || c.searchMode || c.all || c.path {
        return true, "skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all"
    }
}
```

No collision with family (1): completion is NOT in the listing-mode count (that set
stays exactly {path,list,search,all}). `completion <single-mode>` (e.g.
`completion --list`) is caught by THIS block, not masked by family (1). A
2+-listing-mode combo WITH completion (e.g. `completion --path --list`) is caught by
family (1) first (exit 2, correct). `completion + check` and `completion + init` are
caught by THIS block (c.check / c.init are in its set).

Message wording mirrors the existing `skilldozer: '<cmd>' cannot be combined with …`
convention (main.go:741, 744, 748, 757, 760).

---

## §7 — run() precedence: exclusivity fires BEFORE dispatch and BEFORE no-mode

run() ladder (main.go, after P1.M1.T1.S1 lands): help → version → unknownFlag →
storeMissingValue → **exclusivity** → init-dispatch → (path/list/search/check/all/tags)
→ no-mode-fallthrough (now stdout/exit0).

CONSEQUENCE (load-bearing for the tests): the completion EXCLUSIVITY tests
(`run(["completion","--list"])` etc.) exit at the exclusivity step and NEVER call
skillsdir.Find(), discover.Index(), or runCompletion. They need NO store fixture,
NO SKILLDOZER_SKILLS_DIR, NO t.Chdir, NO unsetSkillsEnv. Pure argv → exit-2 checks.

This subtask does NOT add `if c.completion { return runCompletion(...) }` dispatch
(that is P1.M2.T2.S2). So after this subtask, `skilldozer completion` (no conflict)
parses correctly (c.completion=true) but falls through dispatch to the no-mode
default (implicit help: stdout usage, exit 0 — per P1.M1.T1.S1). That is EXPECTED
scaffolding, NOT a bug. So: do NOT write a run()-level completion SUCCESS test
(it would depend on T2.S2's dispatch). Completion SUCCESS tests are parseArgs-level
only.

---

## §8 — usageText edits (PRD §6.1 / contract LOGIC (e)) — exact placement

Current usageText (main.go:52-100) already has the `init [<dir>]` row (plan/002).
Edits (gofmt does NOT reformat raw-string const contents, so alignment is manual):

- **USAGE block:** add `skilldozer completion [--shell <name>]` on its own line
  immediately AFTER the `skilldozer init [<dir>]` line and BEFORE `skilldozer --path`
  (PRD §6.1 table order: … init, completion, --path, …).
- **EXAMPLES block:** add one line after the `skilldozer init --store <dir> …` line:
  `  eval "$(skilldozer completion)"     # load completions into your shell`.
- **OPTIONS block:** add two lines after the `--store <dir>  Non-interactive store path for init` line:
  `  completion [--shell <name>]   Emit the shell completion script for eval (§14.6)`
  `  --shell <bash|zsh|fish>      Force a shell for completion (else auto-detect)`

ALIGNMENT REALITY: the OPTIONS description column is ~col 20 today (the longest
existing entry is `--search <q>, -s` at ~16 chars). `completion [--shell <name>]` is
26 chars — it OVERFLOWS the column. Do not re-align the whole table; let the long
entry's description start right after it (1-2 spaces). Tests assert only SUBSTRING
presence (e.g. Contains("skilldozer completion"), Contains("--shell")), never exact
columns, so the overflow is cosmetically imperfect but functionally correct.
Eyeball it against the `init [<dir>]` / `--store <dir>` rows.

These additions do NOT remove any substring TestRunHelpToStdoutExit0 (main_test.go:1620)
or TestRunHelpShowsInitRow (main_test.go:1969) assert, so they stay green. A NEW test
asserts the completion row + --shell line are present.

---

## §9 — No conflict with the parallel sibling P1.M1.T1.S1 (disjoint regions)

P1.M1.T1.S1 edits: usageText DOC comment (48-51) · run() exit-code doc (417-423) ·
no-mode fallthrough (695-700) · TestRunDefaultNoArgs (277-291) · TestRunModifiersOnlyNoMode
(1668-1684).

This subtask edits: config struct (128-151) · =-form switch (188-204) · main token
switch + init-case guard (220-312) · exclusivityError (722-770) · usageText CONST body
(52-100) · NEW main_test.go functions.

DISJOINT in both files (doc-comment lines 48-51 vs const body 52-100; fallthrough 695-700
vs exclusivityError 722-770; the two flipped tests vs new completion tests). No
text-level overlap; the two changesets compose. P1.M1 lands before P1.M2 by plan
ordering, so the no-mode fallthrough is stdout/exit0 when this subtask begins — but
NONE of this subtask's tests depend on that (exclusivity tests exit at step 5, before
no-mode; parseArgs tests don't call run()).

---

## §10 — Scope discipline + zero deps

- Do NOT add the run() completion dispatch (`if c.completion { return runCompletion }`),
  the `//go:embed` declarations, `completionScript`, `detectShell`, or `runCompletion` —
  those are P1.M2.T2.S1 (embed) and P1.M2.T2.S2 (dispatch+detection). This subtask is
  PARSING + EXCLUSIVITY + USAGE only.
- Do NOT touch the three `completions/*` files (P1.M3.T1.S1 lockstep) or the README
  (P1.M3.T1.S2 Mode B).
- Do NOT add deps/imports. `strings` is already imported (the init-guard fix and the
  new cases use only strings.HasPrefix + bool/string fields). go.mod/go.sum byte-for-byte
  unchanged. Verify with `git diff --quiet go.mod go.sum`.
- Do NOT modify PRD.md (read-only), tasks.json, prd_snapshot.md, or .gitignore.

---

## §11 — Validation (Go toolchain, verified commands)

```bash
gofmt -l main.go main_test.go   # must print NOTHING
go vet ./...                    # exit 0
go build ./...                  # exit 0
go test -run 'Completion|Shell' -v ./...                 # the new parse/exclusivity/help tests
go test ./...                                              # whole module green; zero regressions
git diff --quiet go.mod go.sum && echo deps unchanged
# Manual: --help advertises completion + --shell
go run . --help | grep -E 'skilldozer completion|--shell'
# Manual: exclusivity fires (exit 2) and stdout stays empty
go run . completion --list >/dev/null 2>&1; echo "exit=$? (want 2)"
go run . completion example >/dev/null 2>&1; echo "exit=$? (want 2)"
go run . check completion >/dev/null 2>&1; echo "exit=$? (want 2)"
go run . init completion >/dev/null 2>&1; echo "exit=$? (want 2 — requires the §4 init-guard fix)"
```
