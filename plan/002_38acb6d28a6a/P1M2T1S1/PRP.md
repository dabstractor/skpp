# PRP — P1.M2.T1.S1: Add `init` config fields, `case "init"`, `--store` flag, USAGE row, init exclusivity family

> **Subtask:** The parsing/dispatch-scaffolding half of `skilldozer init` (PRD §8.2). Adds the `init`/`initStore` config fields, a `case "init":` in the token switch (with `init <dir>` positional capture), the `--store` flag (both `=`-form and long-form, mirroring `--search`), the `init` row in the in-code `usageText` help block, and an `init` exclusivity family in `exclusivityError`. **It does NOT implement the init success flow** — `run()`'s `if c.init { … }` dispatch is the next subtask (P1.M2.T2.S3). After this subtask, `skilldozer init` parses correctly (sets `c.init`) but falls through dispatch to the no-mode default (exit 1); the exclusivity conflicts (`init --list`, `init --path`, `init check`, …) already exit 2.
>
> **Scope:** Two existing files only — `main.go` and `main_test.go`. No new files. No `internal/*` change. Zero new dependencies (pure stdlib `strings`, already imported). go.mod/go.sum byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** read main.go + main_test.go in full. The parallel sibling P1.M1.T2.S2 does NOT touch `main.go` (confirmed by its "main.go UNCHANGED" invariant) and does NOT touch the parseArgs/exclusivity/USAGE/init regions of main_test.go — so every line number below is stable. `grep` confirms main.go has zero current `init`/`--store`/`initStore` references, so this is purely additive.

---

## Goal

**Feature Goal**: Wire the `init` subcommand and its `--store` flag through the argument-parsing + exclusivity + help layers of `main.go`, so that PRD §8.2's four invocation forms — `skilldozer init`, `skilldozer init <dir>`, `skilldozer init --store <dir>`, and `skilldozer --store <dir>` — all populate `config{init:true, initStore:<dir>}` without `init` being mis-captured as a tag, the help block advertises `init`, and every init-vs-other-mode conflict exits 2 via `exclusivityError`. This unblocks P1.M2.T2.S3 (run() reads `c.init`/`c.initStore`).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — (a) two new fields on `config` (`init bool`, `initStore string`); (b) `case "--store":` in the `=`-form switch + `case "--store":` and `case "init":` in the main token switch; (c) `init [<dir>]` USAGE line + `init --store <dir>` EXAMPLE + two OPTIONS lines in `usageText`; (d) an `init` exclusivity family in `exclusivityError`.
2. `main_test.go` — new tests: 6 parseArgs-level (the four forms + `--store`-without-`init` + dir-not-a-tag) and 7 run-level exclusivity (`init --list`, `init --path`, `init check`, `init --search q`, `init --all`, `init foo bar` stray-tag) + 1 USAGE assertion (`--help` shows `init` + `--store`).

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `parseArgs(["init","/tmp/x"])` → `{init:true, initStore:"/tmp/x", tags:[]}`; `run(["init","--list"])` → exit 2 with empty stdout; `skilldozer --help` stdout contains `skilldozer init` and `--store <dir>`.

---

## User Persona (if applicable)

**Target User**: A first-run `skilldozer` user (and the scripts/CI that drive `init --store <dir>` non-interactively) who needs the binary to *recognize* the `init` command. (The interactive prompt, store creation, and config writing land in P1.M2.T2.S1–S3; this subtask only makes the parser understand `init`.)

**Use Case**: `skilldozer init` / `skilldozer init --store /path` typed at the shell, or `--help` to discover the command exists.

**User Journey**: User runs `skilldozer --help` → sees the `init [<dir>]` row + `--store <dir>` option → runs `skilldozer init` → (dispatch is P1.M2.T2.S3; today it no-ops to exit 1, which is correct scaffolding). If they typo `skilldozer init --list`, they get a clear `init cannot be combined with …` error and exit 2.

**Pain Points Addressed**: today `skilldozer init` is captured as a *tag* by the default branch and fails as an unknown tag (G7); `--store` is an unknown flag (G7); `init` is invisible in `--help` (G10); `init`+other-modes is silently mishandled (G9).

---

## Why

- **Closes gap G6** (`code_prd_delta.md` §3): the `config` struct has no `init`/`initStore` fields, so there is nowhere for the parser to record that `init` was requested.
- **Closes gap G7** (`code_prd_delta.md` §3): `parseArgs` has no `case "init"` (so `init` is captured as a tag and fails in resolve) and no `--store` handling (so `--store` is an unknown flag → exit 2 with the wrong message).
- **Closes gap G9** (`code_prd_delta.md` §3): `exclusivityError` has no `init` family, so PRD §6.3/§8.2's "init is its own exclusive mode" rule is unenforced.
- **Closes gap G10** (`code_prd_delta.md` §3): the in-code `usageText` help block (the user-facing help surface for `init` — Mode A per the contract) has no `init` row, so `--help` does not advertise the command.
- **Unblocks P1.M2.T2.S3** (run() init dispatch), which reads `c.init`/`c.initStore` and orchestrates the prompt/mkdir/seed/config-write/print flow. That subtask cannot begin until parsing produces those fields correctly.
- **Does NOT** touch README (Mode B, P1.M4.T2.S1) or completions (P1.M3.T2.S1) or the example skill (P1.M3.T1.S1) — those are sibling subtasks with their own PRPs.

---

## What

### Success Criteria

- [ ] `config` struct (main.go:122) has `init bool` and `initStore string` fields with doc comments matching the existing field-comment style.
- [ ] `=`-form switch (main.go:159-189) has `case "--store":` setting `c.init = true; c.initStore = val` (mirrors the `--search` case at :181).
- [ ] main token switch (main.go:196-243) has `case "--store":` (long-form, consumes next token like `--search` at :222) AND `case "init":` (sets `c.init = true`, captures the following *positional* token into `c.initStore` unless it is a dashed flag or a reserved subcommand `check`/`init`).
- [ ] `usageText` (main.go:50) USAGE block has `skilldozer init [<dir>]`; EXAMPLES block has `skilldozer init --store <dir>`; OPTIONS block has `init [<dir>]  …` and `--store <dir>  …` lines.
- [ ] `exclusivityError` (main.go:635) has an init family: `c.init` + (`len(c.tags)>0` OR any of `c.check`/`c.list`/`c.searchMode`/`c.all`/`c.path`) ⇒ exit 2 with a message naming the conflict.
- [ ] `go test ./...` green, including the new init parse tests + init exclusivity tests + USAGE test; existing tests unaffected (purely additive).
- [ ] `go.mod`/`go.sum` unchanged; no new files; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit is pinned to a symbol located in the live `main.go` (read in full). The two non-obvious failure modes — (a) `init`'s positional capture swallowing a reserved subcommand (`init check` must NOT set `initStore="check"`), and (b) `--store` implying `init` so `skilldozer --store <dir>` parses as init — are traced against all six contract test cases in `research/verified_facts.md` §4. The `--search` value-capture pattern (the template for `--store`) and the `check` subcommand pattern (the template for `init`) are both read in full from main.go + their tests. The run() precedence (exclusivity before dispatch/Find) is confirmed, so the exclusivity tests need no store fixture. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (line anchors + the init-capture gotcha traced case-by-case)
- file: plan/002_38acb6d28a6a/P1M2T1S1/research/verified_facts.md
  why: "§1 gives the EXACT current main.go line anchors (sibling-safe: P1.M1.T2.S2 does
        not touch main.go). §3 transcribes the --search value-capture pattern to mirror
        for --store (=form @181, longform @222). §2 transcribes the check subcommand
        pattern to mirror for init. §4 is THE gotcha: case 'init' must NOT swallow a
        reserved subcommand (init check must set c.check, not initStore='check') — full
        case-by-case trace. §5 places the init exclusivity family. §6 confirms exclusivity
        runs before Find() (exclusivity tests need no store). §7 splits tests parseArgs vs
        run. §9 gives the exact USAGE text edits."
  critical: "§4 (the init-capture guard) is the #1 one-pass stall if skipped: the naive
             --search mirror makes `init check` capture 'check' as the store dir and
             silently pass exclusivity instead of exiting 2."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as you edit)
- file: main.go
  why: "THE edit target. config struct @122 (add init+initStore). '='-form switch @159-189
        (add case '--store' mirroring --search @181). main token switch @196-243 (add
        case '--store' after --search @222 mirroring its value capture; add case 'init'
        after check @234 with the §4 positional guard). usageText const @50 (USAGE/EXAMPLES/
        OPTIONS edits). exclusivityError @635 (add init family after the check families,
        before return false,''; reuse hasTags @650)."
  pattern: "Value-taking flag = the --search case: `if i+1 < len(args) { c.field = args[i+1];
            i++ }`. Subcommand = the check case: a `case \"<name>\":` that sets one bool and
            is NOT captured as a tag (the default branch would otherwise grab it)."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestParseArgsCheckSubcommand
        @1117 / TestParseArgsCheckAfterFlag @1128 / TestParseArgsCheckAndTagBothCaptured
        @1141 = the parse-test template for init. TestParseArgsSearchLong @877 /
        TestParseArgsSearchConsumesOneValue @907 = the value-flag parse-test template for
        --store. TestRunExclusivityCheckAndTags @1514 / CheckAndList @1529 / CheckAndPath
        @1544 = the exclusivity-test template (assert code==2, empty stdout, stderr Contains
        the mode name). TestRunHelpToStdoutExit0 @1361 = the USAGE-substring test template."
  gotcha: "The init SUCCESS tests are parseArgs-level ONLY (assert c.init/c.initStore/tags).
           Do NOT write run(['init'])==0 — dispatch is P1.M2.T2.S3, so run(['init']) today
           exits 1 (no-mode). The init EXCLUSIVITY tests are run-level and need NO store
           fixture (exclusivity runs before Find())."

# READ-ONLY — the gap analysis (G6/G7/G9/G10 are THIS subtask)
- file: plan/002_38acb6d28a6a/architecture/code_prd_delta.md
  why: "§3 (gaps G6/G7/G9/G10) quotes the exact config struct, parseArgs switch, usageText,
        and exclusivityError regions to edit and ties each to PRD §8.2. §10 gap index gives
        the one-line summary. Confirms main.go has zero current init/--store refs (purely
        additive) and that no internal/* change is needed."
  section: "§3 (G6 config struct, G7 parseArgs, G9 exclusivity, G10 USAGE), §10 gap index."

# READ-ONLY — the parallel sibling PRP (defines the boundary: it does NOT touch main.go)
- file: plan/002_38acb6d28a6a/P1M1T2S2/PRP.md
  why: "Confirms P1.M1.T2.S2 edits internal/skillsdir/* + the 6 unresolvable main_test.go
        tests + flips ErrNotFound, and explicitly leaves main.go UNCHANGED and does not
        touch the parseArgs/exclusivity/USAGE regions. Fixes the boundary so this subtask
        does not collide with the sibling. (Plan ordering: P1.M1 lands before P1.M2, so the
        sibling's main_test.go edits are present when this subtask begins; they are in
        different functions than the init tests, so no text-level merge collision.)"

# READ-ONLY — PRD (the source of truth for the init CLI contract)
- file: PRD.md
  why: "§6.1 (h3.1) gives the `init` row (stdout = the configured store path; exit 0/1).
        §6.3 (h3.3) makes init a mutually-exclusive mode. §8.2 (h3.9) defines the four
        invocation forms (init / init <dir> / init --store <dir> / --store <dir>) and the
        prompt-safety guarantee. §13 (h2.12) acceptance greps `init --store` + the written
        config (consumed by P1.M2.T2, not this subtask, but the parser must recognize it)."
  section: "h3.1 (§6.1 init row), h3.3 (§6.3 exclusivity), h3.9 (§8.2 init forms), h2.12 (§13)."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer && tree -L 2 --noreport | grep -v '_test.go\|\.go$' ; echo "--- go files touched ---"
main.go          # EDIT: +config fields, +case "init"/case "--store" (2 switches), +USAGE lines, +exclusivity init family
main_test.go     # EDIT: +6 parseArgs init tests, +7 run exclusivity init tests, +1 USAGE test
internal/        # untouched (config/, skillsdir/, discover/, resolve/, search/, check/, ui/)
# go.mod / go.sum untouched (zero new deps; stdlib strings already imported)
$ grep -n 'init\|--store\|initStore\|c\.init\b' main.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
main.go          # ADD: init/initStore config fields; case "init" + case "--store" (×2 switches); USAGE init lines; exclusivity init family
main_test.go     # ADD: TestParseArgsInit* (6) + TestRunExclusivityInit* (7) + TestRunHelpShowsInitRow (1)
```

**No new files.** All edits are additive to existing files.

| File | Responsibility |
|---|---|
| `main.go` | Make the parser recognize `init` + `--store`; advertise `init` in `--help`; reject init+other-mode via exclusivity. |
| `main_test.go` | Lock the four parse forms + the exclusivity matrix + the USAGE row. |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL — #1 one-pass stall) — case "init": must NOT swallow a reserved
// subcommand as the store dir. The naive --search mirror
//     if i+1 < len(args) { c.initStore = args[i+1]; i++ }
// makes `init check` capture "check" as initStore and SILENTLY pass exclusivity (c.check
// never gets set), instead of exiting 2 per PRD §6.3. RESOLUTION (traced in
// research/verified_facts.md §4 against ALL contract test cases): capture the next token
// as initStore ONLY if it is neither a dashed flag NOR a reserved subcommand:
//     case "init":
//         c.init = true
//         if i+1 < len(args) {
//             next := args[i+1]
//             if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
//                 c.initStore = next
//                 i++
//             }
//         }
// The reserved set is exactly {"check","init"} (the only positional subcommands). Adding a
// NEW positional subcommand later requires extending this guard.

// GOTCHA #2 — --store IMPLIES init. The contract OUTPUT §4 requires `skilldozer --store
// <dir>` (with NO `init` token) to parse as init. So both --store handlers set c.init=true
// AND c.initStore=<val>. Do NOT add a separate "store-without-init is unknown" branch; the
// init exclusivity family covers conflicts (--store + --list ⇒ init+list ⇒ exit 2). OUTPUT
// is authoritative over the ambiguous LOGIC (c) sentence.

// GOTCHA #3 — --store has NO short form. PRD §6.2 defines no short alias for --store (unlike
// --search/-s). Add case "--store" ONLY (no "-s"-style short). Do NOT touch expandShortBundle
// (its validated char set stays v h p l a f s — 's' is search, NOT store).

// GOTCHA #4 --store-with-no-value mirrors --search-with-no-value: leave c.init=false,
// initStore="" (do NOT invent an exit-2 "needs argument"; the codebase defers that — see
// the --search no-value comment at main.go:227-231 + TestParseArgsSearchNoValueStaysInactive).
// Concretely: only set fields INSIDE `if i+1 < len(args) { … }`.

// GOTCHA #5 — init is NOT a "listing mode". exclusivityError's family-1 count
// (`for _, b := range []bool{c.path,c.list,c.searchMode,c.all}`) must stay exactly those
// four. Adding c.init there would make `init <single-mode>` hit family 1 only when a SECOND
// mode is also present, masking the init+single-mode case. The init family (peer of check)
// catches init+single-mode. A 2+-listing-mode combo WITH init (e.g. init --path --list) is
// still caught by family 1 first (exit 2, correct) — no collision.

// GOTCHA #6 — run() dispatch is P1.M2.T2.S3, NOT this subtask. After this subtask,
// `skilldozer init` (no conflict) sets c.init=true, passes exclusivity, then falls through
// dispatch to the no-mode default (usage to stderr, exit 1). That is EXPECTED scaffolding,
// not a bug. So: do NOT add `if c.init { … }` to run(); do NOT write a run-level init
// SUCCESS test (it would assert exit 0 and fail today). Init success tests are parseArgs-level.

// GOTCHA #7 — exclusivity runs BEFORE dispatch and BEFORE skillsdir.Find() (run() step 4).
// So the init EXCLUSIVITY tests (run(["init","--list"]) → 2 etc.) need NO store fixture, NO
// SKILLDOZER_SKILLS_DIR, NO t.Chdir, NO unsetSkillsEnv. They are pure argv → exit-2 checks.

// GOTCHA #8 — no merge collision with the parallel sibling. P1.M1.T2.S2 edits main_test.go's
// 6 unresolvable tests + hardens unsetSkillsEnv; it does NOT touch the parseArgs/exclusivity/
// USAGE/init regions. This subtask's new init tests are in different functions. main.go is
// untouched by the sibling. So the two changesets compose cleanly (and P1.M1 lands before
// P1.M2 by plan ordering).

// GOTCHA #9 — the USAGE OPTIONS column alignment. The existing options are aligned to a
// fixed left column (see `check              Validate …`). Match that visual alignment for
// the two new lines (`init [<dir>]` and `--store <dir>`). gofmt does NOT reformat raw-string
// const contents, so alignment is manual — eyeball it against the `check` / `--path, -p`
// lines. It is help text; exact column count is not asserted, only substring presence.

// GOTCHA #10 — strings is already imported (main.go uses strings.HasPrefix/IndexByte/etc.).
// No new import. No new dep. go.mod/go.sum must be byte-for-byte unchanged (verify with
// `git diff --quiet go.mod go.sum`).
```

---

## Implementation Blueprint

### Data models and structure

**Two new fields on the existing `config` struct** (main.go:122). No new types, no new structs.

```go
type config struct {
	// … existing fields …
	check       bool     // `skilldozer check` subcommand …
	init        bool     // `skilldozer init [<dir>]` first-run setup subcommand (PRD §8.2).
	                       // Set by the literal `init` token OR by `--store <dir>` (which implies init).
	initStore   string   // the non-interactive store path: captured from `init <dir>` (the
	                       // positional after `init`) or from `--store <dir>` / `--store=<dir>`.
	                       // Empty ⇒ init uses cwd-auto-detect (handled by P1.M2.T2.S3 dispatch).
	tags        []string // positional <tag> args …
	unknownFlag string   // …
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add init + initStore fields to the config struct
  - FILE: main.go (type config struct, ~line 122)
  - PLACE the two new fields AFTER `check bool` and BEFORE `tags []string` (keeps the
    subcommand-ish bools grouped; matches the doc-comment style of the neighbors).
  - ADD (exact text, with PRD §8.2 doc comments matching the existing field style):
      init      bool   // `skilldozer init [<dir>]` first-run setup (PRD §8.2); also set by `--store <dir>` (which implies init)
      initStore string // non-interactive store path: `init <dir>` positional or `--store <dir>` / `--store=<dir>`; empty ⇒ auto-detect (P1.M2.T2.S3)
    (align the `bool`/`string` column with the existing fields — gofmt -w will fix it.)
  - NOTE: keep the struct field order readable; gofmt does not enforce field order.

Task 2: EDIT main.go — add case "--store" to the '='-form switch (--store=<dir>)
  - FILE: main.go (the HasPrefix("--") && Contains("=") switch, ~line 163-185)
  - PLACE the new case right AFTER the existing `case "--search":` (sibling value flag).
  - ADD (mirror the --search case exactly, GOTCHA #2 — implies init):
      case "--store":
          c.init = true
          c.initStore = val
  - GOTCHA #2: sets c.init=true (OUTPUT requires `--store <dir>` to parse as init).
  - GOTCHA #3: no short form here.

Task 3: EDIT main.go — add case "--store" (long-form) and case "init" to the main token switch
  - FILE: main.go (the main `switch a { … }`, ~line 196-243)
  - (3a) ADD `case "--store":` right AFTER `case "--search", "-s":` (sibling value flag).
         Mirror --search's next-token capture (GOTCHA #2, #4):
      case "--store":
          // `--store <dir>`: non-interactive store path for init (PRD §8.2). Mirrors
          // --search's next-token capture; implies init mode (c.init=true). No short
          // form. If it is the LAST token (no value) init stays unset — mirrors
          // --search-no-value (no exit-2 "needs argument" here).
          if i+1 < len(args) {
              c.init = true
              c.initStore = args[i+1]
              i++
          }
  - (3b) ADD `case "init":` right AFTER `case "check":` (sibling subcommand).
         Use the GOTCHA #1 positional guard (traced in research/verified_facts.md §4):
      case "init":
          // `skilldozer init [<dir>]` first-run setup (PRD §8.2). `init` is a RESERVED
          // positional token (like `check`): it selects init mode and is NOT captured as
          // a tag. If the NEXT token is a positional <dir> (not a dashed flag AND not a
          // reserved subcommand check/init), capture it into c.initStore and skip it (i++)
          // — the `init <dir>` form. A following flag (`init --store …`) or subcommand
          // (`init check`) is left for its own case so exclusivity can flag the conflict.
          c.init = true
          if i+1 < len(args) {
              next := args[i+1]
              if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
                  c.initStore = next
                  i++
              }
          }
  - GOTCHA #1 (CRITICAL): the `next != "check" && next != "init"` guard is what makes
    `init check` reach `case "check":` (→ c.check) so exclusivity exits 2, instead of
    swallowing "check" as the store dir.
  - VERIFIED: trace in research/verified_facts.md §4 passes all contract test cases.

Task 4: EDIT main.go — add the init family to exclusivityError
  - FILE: main.go (func exclusivityError, ~line 635-661)
  - PLACE the new family AFTER the two `c.check && …` families and BEFORE `return false, ""`.
    It reuses `hasTags` (defined ~line 650, in scope).
  - ADD (GOTCHA #5 — init is NOT in the family-1 listing count; it is a peer of check):
      // init is its own exclusive mode (PRD §6.3 / §8.2: like `check`). It rejects the
      // listing/inspection modes AND stray tags. A single positional <dir> after `init`
      // is consumed as the store (c.initStore) by parseArgs, so it never reaches c.tags;
      // a SECOND positional, or any positional after `init --store`, lands in c.tags and
      // is rejected here as a stray.
      if c.init {
          if hasTags {
              return true, "skilldozer: 'init' cannot be combined with tag arguments"
          }
          if c.check || c.list || c.searchMode || c.all || c.path {
              return true, "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"
          }
      }
  - Message wording mirrors the existing `skilldozer: '<cmd>' cannot be combined with …`
    convention (main.go:655, 658).

Task 5: EDIT main.go — update the usageText help block (USAGE / EXAMPLES / OPTIONS)
  - FILE: main.go (const usageText, ~line 50-98)
  - (5a) USAGE block: add `skilldozer init [<dir>]` on its own line, immediately AFTER the
         `skilldozer check` line (PRD §6.1 table order: … check, init, --path, …).
  - (5b) EXAMPLES block: add one line after the `skilldozer check …` example:
         `  skilldozer init --store <dir>     # non-interactive first-run setup`
  - (5c) OPTIONS block: add two lines after the `check …` option line (GOTCHA #9 — align
         the description column with the existing rows; gofmt does not touch raw strings):
         `  init [<dir>]      First-run setup: pick/create the skills store and write the config`
         `  --store <dir>     Non-interactive store path for init`
  - These additions do NOT remove any substring TestRunHelpToStdoutExit0 asserts, so it
    stays green. A NEW test (Task 6) asserts the init row + --store line are present.

Task 6: EDIT main_test.go — add the init parse tests + exclusivity tests + USAGE test
  - FILE: main_test.go
  - GROUP the parseArgs init tests near TestParseArgsCheckAndTagBothCaptured (~line 1141),
    and the run exclusivity init tests near TestRunExclusivityCheckAndPath (~line 1544),
    mirroring those tests' shapes exactly (see research/verified_facts.md §7).
  - (6a) parseArgs SUCCESS tests (parseArgs-level; assert fields, NO run()):
      func TestParseArgsInitSubcommand(t *testing.T) {
          c := parseArgs([]string{"init"})
          if !c.init { t.Errorf("parseArgs(init): init=false; want true") }
          if len(c.tags) != 0 { t.Errorf("parseArgs(init): tags=%v; want empty ('init' is a subcommand, not a tag)", c.tags) }
          if c.initStore != "" { t.Errorf("parseArgs(init): initStore=%q; want empty", c.initStore) }
      }
      func TestParseArgsInitPositionalDir(t *testing.T) {
          c := parseArgs([]string{"init", "/tmp/x"})
          if !c.init { t.Errorf("init not set") }
          if c.initStore != "/tmp/x" { t.Errorf("initStore=%q; want /tmp/x", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty (dir consumed as store, not a tag)", c.tags) }
      }
      func TestParseArgsInitStoreLongForm(t *testing.T) {
          c := parseArgs([]string{"init", "--store", "/tmp/x"})
          if !c.init { t.Errorf("init not set") }
          if c.initStore != "/tmp/x" { t.Errorf("initStore=%q; want /tmp/x", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty", c.tags) }
      }
      func TestParseArgsInitStoreEqualsForm(t *testing.T) {
          c := parseArgs([]string{"init", "--store=/tmp/x"})
          if !c.init { t.Errorf("init not set") }
          if c.initStore != "/tmp/x" { t.Errorf("initStore=%q; want /tmp/x", c.initStore) }
      }
      func TestParseArgsStoreWithoutInitToken(t *testing.T) {
          // --store implies init (contract OUTPUT §4: `skilldozer --store <dir>` parses as init).
          c := parseArgs([]string{"--store", "/tmp/x"})
          if !c.init { t.Errorf("--store should set init=true; got false") }
          if c.initStore != "/tmp/x" { t.Errorf("initStore=%q; want /tmp/x", c.initStore) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty", c.tags) }
      }
      func TestParseArgsInitDirNotCapturedAsTag(t *testing.T) {
          // Regression guard: the `init <dir>` positional must NOT also appear in tags.
          c := parseArgs([]string{"init", "/tmp/x"})
          for _, tg := range c.tags {
              if tg == "/tmp/x" { t.Errorf("dir leaked into tags: %v", c.tags) }
          }
      }
  - (6b) run EXCLUSIVITY tests (run-level; assert code==2, empty stdout, stderr msg;
         GOTCHA #7 — NO store fixture / env needed):
      func TestRunExclusivityInitAndList(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"init", "--list"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init --list): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "init") { t.Errorf("stderr=%q; want a message mentioning init", errOut.String()) }
      }
      func TestRunExclusivityInitAndPath(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"init", "--path"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init --path): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "init") { t.Errorf("stderr=%q; want a message mentioning init", errOut.String()) }
      }
      func TestRunExclusivityInitAndCheck(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"init", "check"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init check): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          // GOTCHA #1 guard lets 'check' reach its case → c.check → init+check conflict.
          if !strings.Contains(errOut.String(), "init") { t.Errorf("stderr=%q; want a message mentioning init", errOut.String()) }
      }
      func TestRunExclusivityInitAndSearch(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"init", "--search", "q"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init --search q): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
      }
      func TestRunExclusivityInitAndAll(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"init", "--all"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init --all): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
      }
      func TestRunExclusivityInitAndStrayTag(t *testing.T) {
          // `init foo bar`: foo → initStore (consumed), bar → tags (stray) → init+tags exit 2.
          var out, errOut bytes.Buffer
          code := run([]string{"init", "foo", "bar"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init foo bar): code=%d; want 2 (stray tag)", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "tag") { t.Errorf("stderr=%q; want a message mentioning tag", errOut.String()) }
      }
  - (6c) USAGE test (run-level --help; assert substrings — does NOT break TestRunHelpToStdoutExit0):
      func TestRunHelpShowsInitRow(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"--help"}, &out, &errOut)
          if code != 0 { t.Fatalf("run(--help): code=%d; want 0", code) }
          got := out.String()
          for _, want := range []string{"skilldozer init", "--store <dir>"} {
              if !strings.Contains(got, want) { t.Errorf("run(--help) stdout missing %q:\n%s", want, got) }
          }
      }
  - GOTCHA #6: do NOT add a run-level init SUCCESS test (run(["init"]) exits 1 today; dispatch is P1.M2.T2.S3).

Task 7: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go test ./...                     # ALL pass incl. the new init tests; existing unaffected
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #10
  - manual: go run . --help | grep -E 'init|store'   # shows the new USAGE/OPTIONS lines
  - manual: go run . init --list; echo "exit $?"     # prints the exclusivity msg, exits 2
```

### Implementation Patterns & Key Details

```go
// case "init": — the subcommand with optional positional store capture.
// GOTCHA #1: the reserved-subcommand guard is load-bearing (see verified_facts.md §4).
case "init":
	c.init = true
	if i+1 < len(args) {
		next := args[i+1]
		if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
			c.initStore = next
			i++
		}
	}

// case "--store": (long form) — mirrors --search's next-token capture; implies init.
case "--store":
	if i+1 < len(args) {
		c.init = true
		c.initStore = args[i+1]
		i++
	}

// (in the '='-form switch) case "--store": — mirrors --search's =form; implies init.
case "--store":
	c.init = true
	c.initStore = val

// exclusivityError — the init family (peer of check; init is NOT in the listing-mode count).
if c.init {
	if hasTags {
		return true, "skilldozer: 'init' cannot be combined with tag arguments"
	}
	if c.check || c.list || c.searchMode || c.all || c.path {
		return true, "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"
	}
}
```

Notes easy to get wrong:
- The `next != "check" && next != "init"` guard is the difference between `init check` exiting 2 (correct) and silently setting `initStore="check"` (wrong). If you mirror `--search` blindly without it, the contract's `init check`/`init --path`-style guarantees break.
- `--store` sets `c.init = true`. Do not branch on "is init also present?" — `--store` alone IS init.
- The init exclusivity tests need no `t.Setenv`/`t.Chdir`/store: exclusivity runs before `skillsdir.Find()`.
- Do not add `if c.init { … }` dispatch to `run()` — that is P1.M2.T2.S3.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **`--store` implies `init` (sets `c.init=true`).** The contract OUTPUT §4 lists `skilldozer --store <dir>` as a form that "parses without being treated as tags", and LOGIC (c) says `--store` "sets c.init=true and c.initStore". So `--store` unconditionally activates init mode. The ambiguous LOGIC phrase "if seen without init it is an unknown/incompatible flag (exit 2)" is satisfied by the init exclusivity family (any other mode + `--store` ⇒ init+mode ⇒ exit 2), not by a special branch. OUTPUT is authoritative.
2. **`case "init":` captures at most ONE following positional, gated against flags + reserved subcommands.** This satisfies the `init <dir>` positional form while keeping `init check` and `init --store …` correct. A hardcoded reserved set {"check","init"} is the minimal correct guard; adding a future positional subcommand requires extending it (documented in the case's comment).
3. **Stray tags after init are caught by exclusivity, not parse-time error.** `init foo bar`: `foo`→initStore, `bar`→tags (default branch), then `exclusivityError`'s init+tags family exits 2. This reuses the existing tags channel rather than inventing a parse-time "stray" error, keeping the parser a pure recognizer and the policy in one place (`exclusivityError`).
4. **`--store`-with-no-value mirrors `--search`-with-no-value (leaves fields unset → no-mode → exit 1).** Consistent with the existing `--search` no-value behavior (main.go:227-231, TestParseArgsSearchNoValueStaysInactive). A stricter exit-2 "needs an argument" is a separate, repo-wide concern, out of scope here.
5. **init is NOT added to the listing-mode count (exclusivityError family 1).** Init is a peer of `check`, not a listing mode. Putting it in the count would mask `init <single-mode>`. A 2+-listing-mode combo with init is still caught by family 1 first (correct).
6. **No run() dispatch in this subtask.** The contract OUTPUT explicitly assigns `run()` init dispatch to P1.M2.T2.S3. Adding it here would be scope creep and would conflict with that subtask. This subtask's job is parsing + exclusivity + help; the success-flow tests are parseArgs-level.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod UNCHANGED. Zero new imports (stdlib strings already imported by main.go).
    No go get, no go mod tidy. git diff --quiet go.mod go.sum MUST report "deps unchanged".

CONSUMERS (NOT built in this subtask — listed to fix the interface):
  - run() init dispatch (P1.M2.T2.S3): reads c.init / c.initStore to run the prompt /
        mkdir / seed / config-write / print flow. After THIS subtask, c.init/c.initStore
        are populated correctly; P1.M2.T2.S3 only adds `if c.init { … }` to run().
  - §13 acceptance (P1.M4.T1.S1): `SKILLDOZER_CONFIG=… ./skilldozer init --store /tmp/…`
        must be RECOGNIZED (this subtask) and then EXECUTED (P1.M2.T2). The parser
        recognition is this subtask's contribution to that gate.

NO ROUTES / NO DATABASE / NO CONFIG-FIELD-ADDITIONS / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after editing main.go)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0
go build ./...                  # expect exit 0
# Expected: zero output / exit 0.
```

### Level 2: Unit Tests (component validation — the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test ./... -run 'TestParseArgsInit|TestParseArgsStore|TestRunExclusivityInit|TestRunHelpShowsInitRow' -v
# Expected: ALL pass. The load-bearing assertions:
#   TestParseArgsInitSubcommand          -> c.init=true, tags empty, initStore="".
#   TestParseArgsInitPositionalDir       -> initStore="/tmp/x", tags empty (dir is NOT a tag).
#   TestParseArgsInitStoreLongForm       -> initStore="/tmp/x" via --store <dir>.
#   TestParseArgsInitStoreEqualsForm     -> initStore="/tmp/x" via --store=/tmp/x.
#   TestParseArgsStoreWithoutInitToken   -> c.init=true (GOTCHA #2: --store implies init).
#   TestRunExclusivityInitAndCheck       -> exit 2 (GOTCHA #1: 'init check' did NOT swallow 'check').
#   TestRunExclusivityInitAndStrayTag    -> exit 2 (init foo bar: foo→store, bar→stray tag).

# Isolated re-run of the existing check/search/help tests (regression — must stay green):
go test ./... -run 'TestParseArgsCheck|TestParseArgsSearch|TestRunExclusivityCheck|TestRunHelpToStdoutExit0|TestRunExclusivityListingModePairs' -v
# Expected: PASS (purely additive change; nothing renamed/removed).
```

### Level 3: Whole-module regression + manual behavior check

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0

# GOTCHA #10 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Manual: --help advertises init + --store
go run . --help | grep -E 'skilldozer init|--store'   # both lines present

# Manual: exclusivity fires (exit 2) and stdout stays empty
out=$(go run . init --list 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "2" ] && echo "init --list exclusivity OK"
out=$(go run . init --path 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "2" ] && echo "init --path exclusivity OK"
out=$(go run . init check 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "2" ] && echo "init check exclusivity OK (GOTCHA #1 guard works)"

# Manual: --store implies init (parses, does not error as unknown flag)
go run . --store /tmp/x 2>/dev/null; echo "exit $? (NOT exit 2 'unknown flag'; today falls to no-mode exit 1 — dispatch is P1.M2.T2.S3)"
# Expected: exit 1 (no-mode default), NOT exit 2 with "unknown flag '--store'".
```

### Level 4: Creative & Domain-Specific Validation

```bash
# N/A — this subtask is pure argv parsing + exclusivity + help text. There is no
# filesystem, network, or interactive surface to exercise (the init flow is
# P1.M2.T2.S1–S3). The Level 3 manual checks cover the user-visible behavior.
```

---

## Final Validation Checklist

### Technical Validation

- [ ] All validation levels completed successfully
- [ ] All tests pass: `go test ./...`
- [ ] No vet errors: `go vet ./...`
- [ ] No formatting issues: `gofmt -l main.go main_test.go` (empty)
- [ ] go.mod/go.sum unchanged: `git diff --quiet go.mod go.sum`

### Feature Validation

- [ ] `parseArgs(["init"])` → `c.init=true`, `tags==[]`, `initStore==""`
- [ ] `parseArgs(["init","/tmp/x"])` → `c.init=true`, `initStore=="/tmp/x"`, `tags==[]`
- [ ] `parseArgs(["init","--store","/tmp/x"])` and `["init","--store=/tmp/x"]` → `initStore=="/tmp/x"`
- [ ] `parseArgs(["--store","/tmp/x"])` → `c.init=true`, `initStore=="/tmp/x"` (no `init` token)
- [ ] `run(["init","--list"])`, `run(["init","--path"])`, `run(["init","check"])` → exit 2, empty stdout
- [ ] `run(["init","foo","bar"])` → exit 2 (stray tag)
- [ ] `run(["--help"])` stdout contains `skilldozer init` and `--store <dir>`
- [ ] `init` is NOT captured as a tag in any of the above (regression guard)

### Code Quality Validation

- [ ] Follows existing codebase patterns: `case "init":` mirrors `case "check":`; `--store` mirrors `--search`; exclusivity message mirrors the `'<cmd>' cannot be combined with …` convention
- [ ] Field/case doc comments match the existing style and cite PRD §8.2
- [ ] Anti-patterns avoided (see below)
- [ ] No new dependencies; `strings` already imported

### Documentation & Deployment

- [ ] In-code `usageText` (the Mode A user-facing help surface) advertises `init` + `--store` (README is Mode B / P1.M4.T2.S1 — NOT touched here)
- [ ] No new environment variables

---

## Anti-Patterns to Avoid

- ❌ Don't mirror `--search`'s value capture for `case "init":` WITHOUT the reserved-subcommand guard (GOTCHA #1) — `init check` would swallow "check" as the store and silently pass exclusivity.
- ❌ Don't add `c.init` to the exclusivityError listing-mode count (family 1) — it would mask `init <single-mode>`.
- ❌ Don't add an `if c.init { … }` dispatch branch to `run()` — that is P1.M2.T2.S3.
- ❌ Don't write a run-level init SUCCESS test (e.g. `run(["init"]) == 0`) — dispatch isn't implemented yet (today exit 1).
- ❌ Don't add a short form for `--store` or touch `expandShortBundle` — PRD §6.2 defines none.
- ❌ Don't invent an exit-2 "needs argument" for `--store`-no-value — mirror `--search`-no-value (exit 1 / no-mode).
- ❌ Don't touch README, completions, the example skill, or `internal/*` — those are sibling subtasks.
- ❌ Don't add new dependencies or imports — `strings` is already imported; this is pure parsing.

---

## Confidence Score

**9/10** — one-pass implementation success likelihood. The change is purely additive to two files, mirrors two existing, well-tested patterns (`check` subcommand + `--search` value capture) that are fully transcribed in the PRP, and the single non-obvious gotcha (the `init`-capture reserved-subcommand guard) is traced case-by-case against every contract test in `research/verified_facts.md` §4. The one residual risk is manual column alignment in the raw-string `usageText` (gofmt does not reformat it), mitigated by the USAGE substring test and the "eyeball against the `check` line" note.
