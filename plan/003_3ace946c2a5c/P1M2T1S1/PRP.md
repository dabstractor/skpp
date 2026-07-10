# PRP — P1.M2.T1.S1: `completion` token + `--shell` flag + config fields + USAGE row + exclusivity family

> **Subtask:** The parsing/exclusivity/help-scaffolding half of the `completion` subcommand (PRD §6.1 / §6.3 / §14.6). Adds the `completion`/`completionShell` config fields, a `case "completion":` in the token switch (mirroring `case "check":`), the `--shell` flag in both the `=`-form and long-form switches (mirroring `--store`, which implies completion), a `completion` exclusivity family in `exclusivityError` (mirroring the `init` family), the `completion` rows in the in-code `usageText` help block, and a one-condition extension to the `init` positional-capture guard so `init completion` exits 2 (consistency with `init check`). **It does NOT implement the completion success flow** — `run()`'s `if c.completion { return runCompletion(...) }` dispatch, the `//go:embed` declarations, `completionScript`, `detectShell`, and `runCompletion` are P1.M2.T2.S1/S2.
>
> **Scope:** Two existing files only — `main.go` and `main_test.go`. No new files. No `internal/*` change. Zero new dependencies (pure stdlib `strings`, already imported). go.mod/go.sum byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** read main.go (config struct / parseArgs switches / exclusivityError / usageText) + main_test.go (check+init parse/exclusivity/help tests) at exact line ranges. The parallel sibling P1.M1.T1.S1 (no-args implicit-help flip) edits DISJOINT regions (doc comment 48-51, run() docs 417-423, fallthrough 695-700, two flipped tests) — no collision. `grep` confirms main.go/main_test.go have ZERO current `completion`/`--shell` references, so this is purely additive.

---

## Goal

**Feature Goal**: Wire the `completion` subcommand (PRD §14.6) and its `--shell` flag through the argument-parsing + exclusivity + help layers of `main.go`, so that PRD §6.3's "`completion` is a reserved subcommand (like `check`/`init`), mutually exclusive with tags and other modes" is enforced, and `--help` advertises it. This unblocks P1.M2.T2.S2 (run() reads `c.completion`/`c.completionShell` and emits the embedded script).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — (a) two new fields on `config` (`completion bool`, `completionShell string`); (b) `case "--shell":` in the `=`-form switch + `case "completion":` and `case "--shell":` in the main token switch; (c) extend the `init` positional-capture guard to exclude `completion`; (d) a `completion` exclusivity family in `exclusivityError`; (e) `completion [--shell <name>]` USAGE line + `eval "$(skilldozer completion)"` EXAMPLE + two OPTIONS lines in `usageText`.
2. `main_test.go` — 4 parseArgs-level tests + 4 run-level exclusivity tests + 1 USAGE assertion.

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `parseArgs(["completion"])` → `{completion:true, tags:[]}`; `parseArgs(["completion","--shell","bash"])` and `["completion","--shell=bash"]` → `{completion:true, completionShell:"bash"}`; `run(["completion","--list"])`, `run(["completion","example"])`, `run(["check","completion"])`, `run(["init","completion"])` → exit 2 with empty stdout; `skilldozer --help` stdout contains `skilldozer completion` and `--shell`.

---

## User Persona (if applicable)

**Target User**: A user who wants to load shell completions via `eval "$(skilldozer completion)"` (PRD §14.6), and who needs the binary to *recognize* the `completion` command and its `--shell` flag. (The actual script emission is P1.M2.T2; this subtask only makes the parser understand `completion`/`--shell`.)

**Use Case**: `skilldozer completion --shell bash` (or `skilldozer completion` for auto-detect), or `skilldozer --help` to discover the command.

**User Journey**: User runs `skilldozer --help` → sees the `completion [--shell <name>]` row → runs `eval "$(skilldozer completion)"` → (dispatch is P1.M2.T2; today it no-ops to implicit help, correct scaffolding). If they typo `skilldozer completion --list`, they get a clear `completion cannot be combined with …` error and exit 2.

**Pain Points Addressed**: today `skilldozer completion` is captured as a *tag* by the default branch and fails as an unknown tag (G7-equivalent); `--shell` is an unknown flag; `completion` is invisible in `--help`; `completion`+other-modes is silently mishandled.

---

## Why

- **Implements PRD §6.1** (the `completion [--shell <name>]` CLI row) and **§6.3** ("`completion` is a reserved subcommand … mutually exclusive with tags and other modes").
- **Implements PRD §14.6** parsing contract: `--shell <bash|zsh|fish>` is the explicit override; bare `completion` auto-detects (the detection + emission is P1.M2.T2; this subtask supplies the flag plumbing).
- **Closes the gap that the `check`/`init` precedent establishes**: every reserved subcommand needs (a) a config field, (b) a parse case, (c) an exclusivity family, (d) a USAGE row. `completion` currently has none.
- **Unblocks P1.M2.T2.S2** (run() completion dispatch + embed + detection), which reads `c.completion`/`c.completionShell`.
- **Does NOT** touch the `completions/*` files (P1.M3.T1.S1 lockstep), the README (P1.M3.T1.S2 Mode B), or the embed/dispatch (P1.M2.T2).

---

## What

### Success Criteria

- [ ] `config` struct (main.go:128) has `completion bool` and `completionShell string` fields with doc comments citing §14.6.
- [ ] `=`-form switch (main.go:188-204) has `case "--shell":` setting `c.completion=true; c.completionShell=val` (mirrors `--store`; NO empty-value guard — mirrors `--search`).
- [ ] main token switch (main.go:220-312) has `case "completion":` (sets `c.completion=true`, mirrors `case "check":`) and `case "--shell":` (next-token capture mirroring `--store`, silent no-value mirroring `--search`).
- [ ] the `init` positional-capture guard (main.go:290+) is extended to `&& next != "completion"` so `init completion` exits 2 (consistency with `init check`).
- [ ] `exclusivityError` (main.go:722-770) has a `completion` family (after the `init` block): `c.completion` + (`hasTags` OR any of `c.check`/`c.init`/`c.list`/`c.searchMode`/`c.all`/`c.path`) ⇒ exit 2.
- [ ] `usageText` (main.go:52) USAGE block has `skilldozer completion [--shell <name>]`; EXAMPLES has `eval "$(skilldozer completion)"`; OPTIONS has `completion [--shell <name>]` and `--shell <bash|zsh|fish>` lines.
- [ ] `go test ./...` green, including the new tests; existing tests unaffected (purely additive).
- [ ] `go.mod`/`go.sum` unchanged; no new files; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit is pinned to a symbol located in the live `main.go` (read in full at the cited ranges). The two non-obvious failure modes — (A) the `init` positional-capture guard swallowing `completion` as a store dir (must be extended so `init completion` exits 2), and (B) the contract's mistaken claim that `completion completion` is caught as a duplicate (it is idempotent like `check check`) — are traced in `research/verified_facts.md` §4 and §5. The `--store` value-capture pattern (template for `--shell`) and the `check`/`init` subcommand + exclusivity patterns (templates for `completion`) are read in full. The run() precedence (exclusivity before dispatch/no-mode) is confirmed, so the exclusivity tests need no store fixture. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (line anchors + the two cross-cutting gotchas traced case-by-case)
- file: plan/003_3ace946c2a5c/P1M2T1S1/research/verified_facts.md
  why: "§1 the exact current main.go anchors (sibling-safe). §2 the --shell design (mirrors --store:
        implies completion; =form + longform; NO missing-value guard, mirror --search). §3 case
        'completion' mirrors case 'check' (simplest, no positional capture). §4 GOTCHA A: extend the
        init positional-capture guard (`&& next != 'completion'`) so init completion exits 2 (REQUIRED
        for §6.3; traces init check/init init/init <dir>). §5 GOTCHA B: completion completion is
        idempotent like check check — the contract's 'duplicate→tag' claim is WRONG; do NOT add
        duplicate handling or a test asserting exit 2. §6 the exclusivity block. §7 exclusivity runs
        before dispatch/no-mode. §8 usageText edits + column-overflow reality. §9 disjoint from the
        sibling. §10 scope discipline."
  critical: "§4 (init-guard extension — without it `init completion` silently uses 'completion' as a
             store dir, violating §6.3) and §5 (do NOT make completion completion an error — it would
             diverge from check) are the two things most likely to be mishandled."

# MUST READ — the authoritative change map (exact line numbers + mirror patterns)
- file: plan/003_3ace946c2a5c/architecture/code_change_map.md
  why: "Change B (B1-B4, B8) pins THIS subtask's exact sites: B1 config struct fields, B2 the =-form
        --shell case, B3 the main-switch completion + --shell cases, B4 the exclusivity block, B8 the
        usageText rows. NOTE the map's B2/B3 'waffle' resolves to: --shell sets BOTH c.completion=true
        AND c.completionShell (mirrors --store implies init) — match the contract LOGIC (b)/(c), which
        is authoritative. B5/B6/B7 (run dispatch / embed / new funcs) are P1.M2.T2 — NOT this subtask."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as you edit)
- file: main.go
  why: "THE edit target. config struct @128 (add completion+completionShell ~after storeMissingValue).
        '='-form switch @188-204 (add case '--shell' after case '--store' @203). main token switch
        @220-312: add case 'completion' after case 'check' @253; add case '--shell' after case '--store'
        @263. init case @290 (EXTEND its positional-capture guard — §4). exclusivityError @722-770 (add
        the completion block after the init block @~761). usageText const @52 (USAGE/EXAMPLES/OPTIONS)."
  pattern: "Simplest reserved token = case 'check' @253 (`c.check = true`, no capture). Value-taking
            flag = case '--store' @263 (`if i+1<len(args){ c.init=true; c.initStore=args[i+1]; i++ } else{...}`)
            and its '='-form @203. Exclusivity family = the init block @~752-761 (`if c.init { if hasTags{...}
            if <modes>{...} }`)."

# MUST READ — the test file under edit (mirror these test shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. TestParseArgsCheckSubcommand @1224 = the
        parse template for completion. TestParseArgsInitStoreLongForm @1293 / EqualsForm @1310 = the
        value-flag parse templates for --shell. TestRunExclusivityCheckAndPath @1837 + InitAndCheck
        @1888 + InitInit @1946 = the exclusivity templates (run-level, NO store fixture; the InitInit
        test also proves the Issue-4 duplicate-init handling this PRP's §4 mirrors). TestRunHelpShowsInitRow
        @1969 = the USAGE-substring test template."
  gotcha: "The completion SUCCESS tests are parseArgs-level ONLY (assert c.completion/c.completionShell/tags).
           Do NOT write run(['completion'])==0 — dispatch is P1.M2.T2, so run(['completion']) today falls to
           no-mode (implicit help). The completion EXCLUSIVITY tests are run-level and need NO store fixture
           (exclusivity runs before Find())."

# READ-ONLY — external deps (embed/detection context for the --shell values; the embed itself is T2.S1)
- file: plan/003_3ace946c2a5c/architecture/external_deps.md
  why: "Confirms --shell's value set is {bash,zsh,fish} (used in the USAGE `--shell <bash|zsh|fish>`
        wording) and the detection order (explicit --shell → $SKILLDOZER_SHELL → basename($SHELL) → none).
        The //go:embed + completionScript/detectShell/runCompletion it documents are P1.M2.T2 — NOT this
        subtask; do NOT add them here."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint regions, no collision)
- file: plan/003_3ace946c2a5c/P1M1T1S1/PRP.md
  why: "Confirms P1.M1.T1.S1 edits the usageText DOC comment (48-51) · run() exit-code doc (417-423) ·
        no-mode fallthrough (695-700) · TestRunDefaultNoArgs · TestRunModifiersOnlyNoMode. It does NOT
        touch the config struct, parseArgs switches, exclusivityError, the usageText CONST body, or the
        init case. Disjoint from this subtask's regions; land in either order (P1.M1 before P1.M2 by
        plan ordering). NONE of this subtask's tests depend on the no-mode flip (exclusivity exits at
        step 5, before no-mode; parseArgs tests don't call run())."

# READ-ONLY — PRD (the authority for the completion CLI contract)
- file: PRD.md
  why: "READ-ONLY. §6.1 (h3.1) the `completion [--shell <name>]` row (stdout = the script; exit 0 / 1
        undetectable / 2 unsupported --shell). §6.3 (h3.3) completion is a reserved subcommand, mutually
        exclusive with tags and other modes. §6.4 (h3.4) the completion shell-detection-failure +
        unsupported-shell exit codes (1 / 2). §14.6 (h3.19) the --shell detection order + the eval idiom.
        §13 (h2.12) the acceptance greps for `completion --shell bash|fish` (those depend on T2 dispatch,
        but the PARSER must recognize the tokens first)."
  section: "h3.1 (§6.1 row), h3.3 (§6.3 reserved/exclusive), h3.4 (§6.4 exit codes), h3.19 (§14.6), h2.12 (§13)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/003_3ace946c2a5c/tasks.json
  why: "P1.M2.T1.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes
        it; tasks.json wins on any conflict — EXCEPT the two corrections in verified_facts §4/§5 (the
        init-guard extension, required for §6.3; and the completion-completion idempotency, where the
        contract's parenthetical reasoning is mistaken)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod
main.go        main_test.go   go.mod
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep)
$ grep -n 'completion\|--shell\|completionShell\|c\.completion' main.go main_test.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be changed

```bash
main.go        # ADD: completion/completionShell config fields; case "completion" + case "--shell" (×2 switches);
               #     init-guard `&& next != "completion"` extension; exclusivity completion family; USAGE completion lines
main_test.go   # ADD: 4 parseArgs completion tests + 4 run exclusivity completion tests + 1 USAGE test
# go.mod / go.sum — UNCHANGED (zero new deps; stdlib strings already imported)
```

| File | Responsibility |
|---|---|
| `main.go` | Make the parser recognize `completion` + `--shell`; advertise `completion` in `--help`; reject completion+other-mode via exclusivity; keep `init completion` consistent (exit 2). |
| `main_test.go` | Lock the parse forms + the exclusivity matrix + the USAGE row. |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA A (CRITICAL — REQUIRED for §6.3) — the init positional-capture guard SWALLOWS "completion"
// as a store dir unless extended. The current init case (main.go:290+) captures the next positional
// into c.initStore unless it is a dashed flag, "init", or "check":
//     } else if !strings.HasPrefix(next, "-") && next != "check" {
//         c.initStore = next; i++
//     }
// `init completion` → "completion" passes that condition → initStore="completion", the completion
// token never reaches its case → c.completion stays false → no exclusivity fires → init runs with a
// store literally named "completion". That violates §6.3 ("completion is mutually exclusive with
// other modes") and is inconsistent with `init check` (exit 2) and `init init` (exit 2). FIX: add
// `&& next != "completion"` so "completion" reaches its own case → c.completion=true → the completion
// family catches c.completion && c.init → exit 2. (verified_facts §4; mirrors the existing check
// exclusion exactly. No Issue-4 tag-capture trick needed — completion reaching its case sets a DISTINCT
// flag the completion family checks against c.init.)

// GOTCHA B — `completion completion` is IDEMPOTENT like `check check`; do NOT make it an error. The
// contract LOGIC (d) claims it is caught as "duplicate→tag→hasTags" — that is WRONG: because completion
// has its OWN case (mirroring check), the second token hits case "completion" again (idempotent), NOT
// the default→tag branch. hasTags stays false; the completion family does not fire. This matches the
// existing `check check` (which dispatches, not exit 2). Only `init init` exits 2 (Issue-4 tag-capture).
// Do NOT add duplicate-handling for completion and do NOT write a test asserting `completion completion`
// exits 2 (it would fail). (verified_facts §5.) [The OUTPUT test list is authoritative and omits it.]

// GOTCHA C — --shell IMPLIES completion (sets c.completion=true), exactly as --store implies init. The
// contract LOGIC (b)/(c) fixes this. So both --shell handlers set BOTH c.completion=true AND
// c.completionShell=<val>. `--shell bash` with NO `completion` token is therefore a valid completion
// invocation (OUTPUT: "`--shell <dir>` parse without being treated as tags" — same shape --store uses).
// Do NOT add a separate "shell-without-completion" branch; the completion exclusivity family covers
// conflicts (--shell + --list ⇒ completion+list ⇒ exit 2).

// GOTCHA D — NO missing-value guard for --shell (do NOT mirror --store's storeMissingValue). PRD §6.4
// specifies NO missing-value exit code for --shell. Mirror --search's no-value behavior instead:
// long-form `--shell` (last token, no value) → silent no-op (completion stays false → no-mode). The
// =-form `--shell=` is unconditional (completion=true, completionShell=""), mirroring `--search=`.
// (verified_facts §2.)

// GOTCHA E — --shell has NO short form. PRD §6.2/§14.6 define none. Add `case "--shell"` ONLY. Do NOT
// touch expandShortBundle (its char set stays v h p l a f s — 's' is search, NOT shell).

// GOTCHA F — completion is NOT a "listing mode". exclusivityError's family-1 count
// (`[]bool{c.path,c.list,c.searchMode,c.all}`) must stay exactly those four. Adding c.completion there
// would mask `completion <single-mode>`. The completion family (peer of init) catches it. A 2+-listing-
// mode combo WITH completion is still caught by family 1 first (correct, exit 2). (verified_facts §6.)

// GOTCHA G — run() dispatch is P1.M2.T2.S2, NOT this subtask. After this subtask, `skilldozer completion`
// (no conflict) sets c.completion=true, passes exclusivity, then falls through dispatch to no-mode
// (implicit help, stdout/exit0 per P1.M1.T1.S1). That is EXPECTED scaffolding, not a bug. So: do NOT
// add `if c.completion { return runCompletion(...) }`; do NOT write a run-level completion SUCCESS test.
// Completion success tests are parseArgs-level.

// GOTCHA H — exclusivity runs BEFORE dispatch and BEFORE skillsdir.Find() (run() step 5). So the
// completion EXCLUSIVITY tests need NO store fixture, NO SKILLDOZER_SKILLS_DIR, NO t.Chdir, NO
// unsetSkillsEnv. Pure argv → exit-2 checks. (verified_facts §7.)

// GOTCHA I — No conflict with the parallel sibling P1.M1.T1.S1 (disjoint regions — verified_facts §9).
// It edits the usageText DOC comment (48-51) + run() docs + no-mode fallthrough + two flipped tests;
// this subtask edits the config struct + parseArgs switches + exclusivityError + the usageText CONST
// body + new tests. No text-level overlap; the changesets compose.

// GOTCHA J — the USAGE OPTIONS column overflows for completion. `completion [--shell <name>]` (26 chars)
// is longer than the ~16-char column the existing rows use. Do NOT re-align the whole table; let the
// long entry's description follow with 1-2 spaces. gofmt does not reformat raw-string consts. Tests
// assert only substring presence (Contains "skilldozer completion", "--shell"), never exact columns.

// GOTCHA K — No deps/imports change. strings is already imported. go.mod/go.sum byte-for-byte
// identical. Verify with `git diff --quiet go.mod go.sum`.
```

---

## Implementation Blueprint

### Data models and structure

**Two new fields on the existing `config` struct** (main.go:128). No new types/structs.

```go
type config struct {
	// … existing fields (check, init, initStore, storeMissingValue) …
	completion      bool   // `skilldozer completion` subcommand (PRD §14.6); exclusive like check/init.
	                       // Also set by `--shell <name>` (which implies completion, like --store implies init).
	completionShell string // `--shell <bash|zsh|fish>` value (PRD §14.6); "" ⇒ detect from $SKILLDOZER_SHELL/$SHELL (P1.M2.T2.S2).
	tags            []string // …
	unknownFlag     string   // …
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add completion + completionShell fields to the config struct
  - FILE: main.go (type config struct, ~line 128-151)
  - PLACE the two new fields AFTER `storeMissingValue bool` and BEFORE `tags []string`
    (group the subcommand-ish fields with check/init).
  - ADD (doc comments cite §14.6, matching the existing field style):
      completion      bool   // `skilldozer completion` subcommand (PRD §14.6); exclusive like check/init; also set by `--shell <name>` (which implies completion)
      completionShell string // `--shell <bash|zsh|fish>` value (PRD §14.6); "" ⇒ detect from $SKILLDOZER_SHELL/$SHELL (P1.M2.T2.S2)
  - gofmt -w will fix column alignment.

Task 2: EDIT main.go — add case "--shell" to the '='-form switch (--shell=<name>)
  - FILE: main.go (the HasPrefix("--") && Contains("=") switch, ~line 188-204)
  - PLACE the new case right AFTER `case "--store":` (sibling value flag).
  - ADD (GOTCHA C — implies completion; GOTCHA D — NO empty-value guard, mirror --search):
      case "--shell":
          // `--shell=<name>`: force a shell for completion (PRD §14.6). Mirrors --store's '='-form;
          // implies completion mode (c.completion=true). No short form. NO empty-value guard (PRD §6.4
          // specifies no missing-value exit code for --shell — `--shell=` is completion=true, shell="").
          c.completion = true
          c.completionShell = val

Task 3: EDIT main.go — add case "completion" and case "--shell" to the main token switch
  - FILE: main.go (the main `switch a { … }`, ~line 220-312)
  - (3a) ADD `case "completion":` right AFTER `case "check":` (sibling reserved subcommand). Mirror
         check EXACTLY — simplest reserved token, NO positional capture (GOTCHA B):
      case "completion":
          // `skilldozer completion [--shell <name>]` (PRD §14.6). `completion` is a RESERVED positional
          // token (like `check`): it selects completion mode and is NOT captured as a tag. Captured
          // ANYWHERE in argv; run()'s exclusivity rejects completion+tags / completion+mode with exit 2.
          // A nested skill `writing/completion` still resolves: this case matches only the EXACT token.
          c.completion = true
  - (3b) ADD `case "--shell":` right AFTER `case "--store":` (sibling value flag). Mirror --store's
         next-token capture, but with NO else-guard (mirror --search no-value; GOTCHA D):
      case "--shell":
          // `--shell <name>`: force a shell for completion (PRD §14.6). Mirrors --store's next-token
          // capture; implies completion mode (c.completion=true) when a value follows. No short form.
          // If --shell is the LAST token (no value), completion stays false — mirrors --search's
          // no-value silent behavior (PRD §6.4 specifies no missing-value exit code for --shell).
          if i+1 < len(args) {
              c.completion = true
              c.completionShell = args[i+1]
              i++
          }

Task 4: EDIT main.go — extend the init positional-capture guard (GOTCHA A — REQUIRED for §6.3)
  - FILE: main.go (the `case "init":` positional-capture guard, ~line 290-310)
  - FIND the guard's else-if (it currently excludes dashed flags, "init", and "check"):
        } else if !strings.HasPrefix(next, "-") && next != "check" {
            c.initStore = next
            i++
        }
  - REPLACE with (add `&& next != "completion"`):
        } else if !strings.HasPrefix(next, "-") && next != "check" && next != "completion" {
            c.initStore = next
            i++
        }
  - WHY: without this, `init completion` swallows "completion" as initStore (a store literally named
    "completion") and never sets c.completion, so the completion exclusivity family never fires —
    violating §6.3 ("completion is mutually exclusive with other modes") and diverging from `init check`
    (exit 2). With it, "completion" reaches its own case → c.completion=true → the completion family
    catches c.completion && c.init → exit 2. (verified_facts §4; mirrors the existing `check` exclusion.)
  - Update the init-case doc comment to list completion among the reserved tokens it defers to
    ("a dashed flag (`init --store …`), `init check`, or `init completion` → left for its own case").

Task 5: EDIT main.go — add the completion family to exclusivityError
  - FILE: main.go (func exclusivityError, ~line 722-770)
  - PLACE the new family AFTER the `if c.init { … }` block and BEFORE `return false, ""`. It reuses
    `hasTags` (defined ~line 733, in scope).
  - ADD (GOTCHA F — completion is NOT in the family-1 listing count; it is a peer of init):
      // completion is its own exclusive mode (PRD §6.3 / §14.6: like check/init). It rejects the other
      // modes/subcommands AND stray tags. `completion` does no positional capture (mirrors check), so
      // any positional after it lands in c.tags and is rejected here as a stray.
      if c.completion {
          if hasTags {
              return true, "skilldozer: 'completion' cannot be combined with tag arguments"
          }
          if c.check || c.init || c.list || c.searchMode || c.all || c.path {
              return true, "skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all"
          }
      }
  - Message wording mirrors the `skilldozer: '<cmd>' cannot be combined with …` convention.
  - GOTCHA B reminder: do NOT add a `completion completion` duplicate check here — it is idempotent
    like `check check` (the second token hits case "completion", not default→tag).

Task 6: EDIT main.go — update the usageText help block (USAGE / EXAMPLES / OPTIONS)
  - FILE: main.go (const usageText, ~line 52-100)
  - (6a) USAGE block: add `skilldozer completion [--shell <name>]` on its own line immediately AFTER
         the `skilldozer init [<dir>]` line and BEFORE `skilldozer --path` (PRD §6.1 table order).
  - (6b) EXAMPLES block: add one line after the `skilldozer init --store <dir> …` line:
         `  eval "$(skilldozer completion)"     # load completions into your shell`
  - (6c) OPTIONS block: add two lines after the `--store <dir>  Non-interactive store path for init` line
         (GOTCHA J — let the long `completion [--shell <name>]` entry overflow the column; eyeball it):
         `  completion [--shell <name>]   Emit the shell completion script for eval (§14.6)`
         `  --shell <bash|zsh|fish>      Force a shell for completion (else auto-detect)`
  - These additions do NOT remove any substring TestRunHelpToStdoutExit0 / TestRunHelpShowsInitRow
    assert, so they stay green. A NEW test asserts the completion row + --shell line.

Task 7: EDIT main_test.go — add the parse tests + exclusivity tests + USAGE test
  - FILE: main_test.go
  - GROUP parseArgs completion tests near TestParseArgsCheckSubcommand (@1224)/TestParseArgsInitStore*
    (@1293-1320); group run exclusivity completion tests near TestRunExclusivityInitAndCheck (@1888)/
    InitInit (@1946). Mirror those tests' shapes exactly (verified_facts §7 — exclusivity tests need
    NO store/env). package main; var out, errOut bytes.Buffer; run returns int.
  - (7a) parseArgs SUCCESS tests (parseArgs-level; assert fields, NO run(); GOTCHA G):
      func TestParseArgsCompletionSubcommand(t *testing.T) {
          c := parseArgs([]string{"completion"})
          if !c.completion { t.Errorf("parseArgs(completion): completion=false; want true") }
          if len(c.tags) != 0 { t.Errorf("parseArgs(completion): tags=%v; want empty ('completion' is a subcommand, not a tag)", c.tags) }
          if c.completionShell != "" { t.Errorf("parseArgs(completion): completionShell=%q; want empty", c.completionShell) }
      }
      func TestParseArgsCompletionShellLongForm(t *testing.T) {
          c := parseArgs([]string{"completion", "--shell", "bash"})
          if !c.completion { t.Errorf("completion not set") }
          if c.completionShell != "bash" { t.Errorf("completionShell=%q; want bash", c.completionShell) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty", c.tags) }
      }
      func TestParseArgsCompletionShellEqualsForm(t *testing.T) {
          c := parseArgs([]string{"completion", "--shell=bash"})
          if !c.completion { t.Errorf("completion not set") }
          if c.completionShell != "bash" { t.Errorf("completionShell=%q; want bash", c.completionShell) }
      }
      func TestParseArgsShellImpliesCompletion(t *testing.T) {
          // --shell implies completion (mirrors --store implies init): `--shell bash` with no
          // `completion` token still sets c.completion=true (GOTCHA C).
          c := parseArgs([]string{"--shell", "bash"})
          if !c.completion { t.Errorf("--shell should set completion=true; got false") }
          if c.completionShell != "bash" { t.Errorf("completionShell=%q; want bash", c.completionShell) }
          if len(c.tags) != 0 { t.Errorf("tags=%v; want empty", c.tags) }
      }
  - (7b) run EXCLUSIVITY tests (run-level; assert code==2, empty stdout, stderr msg; GOTCHA H — NO store/env):
      func TestRunExclusivityCompletionAndTag(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"completion", "example"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(completion example): code=%d; want 2 (completion + tag)", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "completion") { t.Errorf("stderr=%q; want a message mentioning completion", errOut.String()) }
      }
      func TestRunExclusivityCompletionAndList(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"completion", "--list"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(completion --list): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "completion") { t.Errorf("stderr=%q; want a message mentioning completion", errOut.String()) }
      }
      func TestRunExclusivityCheckAndCompletion(t *testing.T) {
          // `check completion`: both reserved tokens reach their own cases (c.check + c.completion);
          // the completion family catches c.completion && c.check -> exit 2. (NOT via default->tag —
          // completion has its own case; the contract LOGIC (d) parenthetical is mistaken here, but the
          // exit-2 outcome is correct.)
          var out, errOut bytes.Buffer
          code := run([]string{"check", "completion"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(check completion): code=%d; want 2", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "completion") { t.Errorf("stderr=%q; want a message mentioning completion", errOut.String()) }
      }
      func TestRunExclusivityInitAndCompletion(t *testing.T) {
          // GOTCHA A proof: the init-guard extension (`&& next != "completion"`) lets "completion"
          // reach its case (c.completion) instead of being swallowed as initStore, so the completion
          // family catches c.completion && c.init -> exit 2 (consistent with `init check` / `init init`).
          var out, errOut bytes.Buffer
          code := run([]string{"init", "completion"}, &out, &errOut)
          if code != 2 { t.Fatalf("run(init completion): code=%d; want 2 (GOTCHA A: init-guard must defer completion to its case)", code) }
          if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
          if !strings.Contains(errOut.String(), "completion") { t.Errorf("stderr=%q; want a message mentioning completion", errOut.String()) }
      }
  - (7c) USAGE test (run-level --help; assert substrings — does NOT break TestRunHelpShowsInitRow):
      func TestRunHelpShowsCompletionRow(t *testing.T) {
          var out, errOut bytes.Buffer
          code := run([]string{"--help"}, &out, &errOut)
          if code != 0 { t.Fatalf("run(--help): code=%d; want 0", code) }
          got := out.String()
          for _, want := range []string{"skilldozer completion", "--shell"} {
              if !strings.Contains(got, want) { t.Errorf("run(--help) stdout missing %q:\n%s", want, got) }
          }
      }
  - GOTCHA G: do NOT add a run-level completion SUCCESS test (run(["completion"]) falls to no-mode today).

Task 8: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0
  - go test -run 'Completion|Shell' -v ./...   # the 9 new tests pass
  - go test ./...                     # whole module green; zero regressions
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA K
  - manual: go run . --help | grep -E 'skilldozer completion|--shell'
  - manual: go run . completion --list; echo "exit $?"     # exit 2
  - manual: go run . init completion; echo "exit $?"       # exit 2 (requires the Task 4 guard fix)
```

### Implementation Patterns & Key Details

```go
// case "completion": — the simplest reserved token (mirror case "check": no positional capture).
case "completion":
	c.completion = true

// case "--shell": (long form) — mirrors --store's next-token capture; implies completion; NO else-guard.
case "--shell":
	if i+1 < len(args) {
		c.completion = true
		c.completionShell = args[i+1]
		i++
	}

// (in the '='-form switch) case "--shell": — mirrors --store's =form; implies completion; unconditional.
case "--shell":
	c.completion = true
	c.completionShell = val

// the init-guard extension (Task 4) — add `&& next != "completion"`:
} else if !strings.HasPrefix(next, "-") && next != "check" && next != "completion" {
	c.initStore = next
	i++
}

// exclusivityError — the completion family (peer of init; NOT in the listing-mode count).
if c.completion {
	if hasTags {
		return true, "skilldozer: 'completion' cannot be combined with tag arguments"
	}
	if c.check || c.init || c.list || c.searchMode || c.all || c.path {
		return true, "skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all"
	}
}
```

Notes easy to get wrong:
- The init-guard extension (Task 4) is REQUIRED — without it `init completion` silently uses "completion" as a store dir (GOTCHA A). It is the one edit beyond the literal contract test list, justified by PRD §6.3.
- `completion completion` must NOT be made an error (GOTCHA B); it is idempotent like `check check`. Do not write a test asserting exit 2 for it.
- `--shell` sets `c.completion = true` (GOTCHA C); do not gate it on a `completion` token.
- The completion exclusivity tests need no store fixture (GOTCHA H); the completion success tests are parseArgs-level only (GOTCHA G).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **`completion` mirrors `check`, not `init` (no positional capture).** PRD §14.6's `completion [--shell <name>]` takes its option via the `--shell` FLAG, not a positional. So `case "completion":` is the simplest reserved token (`c.completion = true`), and any positional after it (`completion example`) lands in `c.tags` → caught by the completion+tags family. This also means `completion completion` is idempotent like `check check` (GOTCHA B).
2. **`--shell` implies `completion` (sets `c.completion=true`).** Mirrors `--store` implies `init` (decision from plan/002, reused). So `--shell bash` (no `completion` token) is a valid completion invocation. The exclusivity family covers conflicts.
3. **No missing-value guard for `--shell`.** PRD §6.4 specifies no missing-value exit code for `--shell` (unlike `--store`, whose guard protects the non-destructive config contract). Mirror `--search`'s silent no-value instead. (GOTCHA D.)
4. **Init-guard extension is REQUIRED (not optional).** PRD §6.3 makes completion "mutually exclusive with other modes"; without the guard, `init completion` swallows the token as a store dir, violating §6.3 and diverging from `init check`/`init init`. The fix mirrors the existing `check` exclusion exactly. (GOTCHA A.)
5. **No `completion completion` duplicate handling.** The contract LOGIC (d) parenthetical claiming it is caught as "duplicate→tag" is mistaken (completion has its own case). Mirroring `check` (idempotent) is correct and consistent; adding duplicate handling would diverge from `check`. (GOTCHA B.)
6. **No run() dispatch in this subtask.** The contract OUTPUT assigns dispatch to P1.M2.T2.S2. Adding it here is scope creep. This subtask's success tests are parseArgs-level. (GOTCHA G.)

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. Zero new imports (stdlib strings already imported). (GOTCHA K)

CONSUMERS (NOT built in this subtask — listed to fix the interface):
  - run() completion dispatch (P1.M2.T2.S2): reads c.completion / c.completionShell to emit the
        //go:embed-ded script for the detected shell. After THIS subtask, c.completion/c.completionShell
        are populated correctly; P1.M2.T2.S2 only adds `if c.completion { return runCompletion(...) }`.
  - §13 acceptance (P1.M4 / the delta's acceptance): `completion --shell bash|fish` greps depend on
        T2's dispatch + the embedded scripts, but the PARSER must recognize the tokens first (this subtask).

NO ROUTES / NO DATABASE / NO CONFIG-FORMAT CHANGE / NO NEW FILES.
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

### Level 2: Unit Tests (the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test -run 'Completion|Shell' -v ./...
# Expected: ALL 9 pass. The load-bearing assertions:
#   TestParseArgsCompletionSubcommand        -> c.completion=true, tags empty, completionShell="".
#   TestParseArgsCompletionShellLongForm     -> completionShell="bash" via --shell <name>.
#   TestParseArgsCompletionShellEqualsForm   -> completionShell="bash" via --shell=bash.
#   TestParseArgsShellImpliesCompletion      -> c.completion=true via --shell alone (GOTCHA C).
#   TestRunExclusivityCompletionAndTag       -> run(["completion","example"]) exit 2.
#   TestRunExclusivityCompletionAndList      -> run(["completion","--list"]) exit 2.
#   TestRunExclusivityCheckAndCompletion     -> run(["check","completion"]) exit 2 (completion family).
#   TestRunExclusivityInitAndCompletion      -> run(["init","completion"]) exit 2 (GOTCHA A guard fix).
#   TestRunHelpShowsCompletionRow            -> --help contains "skilldozer completion" + "--shell".

# Regression — the existing check/init/store/help tests stay green:
go test -run 'TestParseArgsCheck|TestParseArgsInit|TestRunExclusivityCheck|TestRunExclusivityInit|TestRunHelpShowsInitRow|TestRunHelpToStdoutExit0' -v ./...
# Expected: PASS (purely additive; nothing renamed/removed; the init-guard change is backwards-compatible).
```

### Level 3: Whole-module regression + manual behavior check

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions

# GOTCHA K invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Manual: --help advertises completion + --shell
go run . --help | grep -E 'skilldozer completion|--shell'   # both lines present

# Manual: exclusivity fires (exit 2) and stdout stays empty
for args in "completion --list" "completion example" "check completion" "init completion"; do
  out=$(go run . $args 2>/dev/null); rc=$?
  [ -z "$out" ] && [ "$rc" = "2" ] && echo "OK: '$args' -> exit 2 + empty stdout" || echo "FAIL: '$args' rc=$rc out=$out"
done
# Expected: all four "OK" (the 'init completion' line requires the Task 4 guard fix).
```

### Level 4: Behavioral spot-checks (lock the fix is scoped, no over-reach)

```bash
cd /home/dustin/projects/skilldozer

# 4a. `completion` alone (no conflict) parses but does NOT dispatch (T2's job) -> falls to no-mode
#     (implicit help, stdout/exit0 per P1.M1.T1.S1). NOT exit 2.
go build -o /tmp/sdz .
/tmp/sdz completion >/dev/null 2>&1; echo "bare completion exit=$? (want 0 — no-mode implicit help, NOT 2; dispatch is T2)"

# 4b. --shell implies completion: `--shell bash` (no completion token) parses as completion (GOTCHA C),
#     does not error as an unknown flag, and (no conflict) falls to no-mode:
/tmp/sdz --shell bash >/dev/null 2>&1; echo "shell-implies-completion exit=$? (want 0 — no-mode; NOT 2 'unknown flag')"

# 4c. CONTROL — a genuine unknown flag still exits 2 (unchanged), proving the new cases didn't widen parsing:
/tmp/sdz --frobnicate >/dev/null 2>&1; echo "unknown-flag exit=$? (want 2)"

rm -f /tmp/sdz
# Expected: bare completion exit 0; --shell bash exit 0; unknown-flag exit 2.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go vet ./...` exit 0, `go build` exit 0
- [ ] Level 2 PASS — the 9 new tests pass (4 parse + 4 exclusivity + 1 USAGE)
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (zero regressions); `git diff go.mod go.sum` → "deps unchanged"; manual exclusivity matrix all exit 2
- [ ] Level 4 PASS — bare `completion` and `--shell bash` fall to no-mode (exit 0, NOT 2); unknown flag still exit 2

### Feature Validation
- [ ] `parseArgs(["completion"])` → `c.completion=true`, `tags==[]`, `completionShell==""`
- [ ] `parseArgs(["completion","--shell","bash"])` and `["completion","--shell=bash"]` → `completionShell=="bash"`
- [ ] `parseArgs(["--shell","bash"])` → `c.completion=true`, `completionShell=="bash"` (implies completion)
- [ ] `run(["completion","--list"])`, `run(["completion","example"])`, `run(["check","completion"])`, `run(["init","completion"])` → exit 2, empty stdout
- [ ] `run(["--help"])` stdout contains `skilldozer completion` and `--shell`
- [ ] `completion` is NOT captured as a tag in any of the above

### Code Quality / Convention Validation
- [ ] Mirrors existing patterns: `case "completion":` mirrors `case "check":`; `--shell` mirrors `--store`; exclusivity block mirrors the `init` block; USAGE rows mirror `check`/`init`/`--store`
- [ ] Field/case doc comments cite PRD §14.6 and match the existing style
- [ ] Anti-patterns avoided (see below)
- [ ] No new dependencies; `strings` already imported

### Scope Discipline
- [ ] Did NOT add run() completion dispatch / `//go:embed` / `completionScript` / `detectShell` / `runCompletion` (P1.M2.T2)
- [ ] Did NOT touch `completions/*` (P1.M3.T1.S1) or the README (P1.M3.T1.S2)
- [ ] Did NOT make `completion completion` an error (GOTCHA B; idempotent like `check check`)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't skip the init-guard extension (Task 4).** Without `&& next != "completion"`, `init completion` swallows "completion" as a store dir, violating §6.3 and diverging from `init check`/`init init`. (GOTCHA A.)
- ❌ **Don't make `completion completion` an error.** It is idempotent like `check check`; the contract's "duplicate→tag" claim is wrong. Do not write a test asserting exit 2 for it. (GOTCHA B.)
- ❌ **Don't gate `--shell` on a `completion` token, and don't add a missing-value guard.** `--shell` implies completion (sets `c.completion=true`); its no-value case mirrors `--search` (silent no-op), NOT `--store` (storeMissingValue). (GOTCHA C/D.)
- ❌ **Don't add a short form for `--shell` or touch `expandShortBundle`.** PRD defines none. (GOTCHA E.)
- ❌ **Don't add `c.completion` to the exclusivityError listing-mode count.** It would mask `completion <single-mode>`. (GOTCHA F.)
- ❌ **Don't add run() completion dispatch.** That is P1.M2.T2.S2; don't write a run-level completion SUCCESS test either. (GOTCHA G.)
- ❌ **Don't add a store fixture/env to the exclusivity tests.** Exclusivity runs before `skillsdir.Find()`. (GOTCHA H.)
- ❌ **Don't re-align the whole USAGE OPTIONS table.** Let the long `completion [--shell <name>]` entry overflow; tests assert substrings only. (GOTCHA J.)
- ❌ **Don't add deps/imports or touch README/completions.** `strings` is already imported; README is Mode B (P1.M3.T1.S2); completions are P1.M3.T1.S1. (GOTCHA K.)

---

## Confidence Score

**9/10** — The change is purely additive to two files, mirrors three existing well-tested patterns (`check` subcommand, `--store` value-capture, `init` exclusivity family) that are fully transcribed, and the two non-obvious gotchas are traced case-by-case in `research/verified_facts.md` §4 (init-guard extension) and §5 (completion-completion idempotency). The init-guard extension is the one edit beyond the literal contract test list, justified by PRD §6.3 and the `init check` precedent, with a dedicated test. The 1-point reservation is for the two deliberate divergences from the contract's literal wording (the init-guard fix the contract omits, and the completion-completion behavior the contract mis-states) — both resolved and documented, but they are the places an implementer following the contract verbatim would go wrong.
