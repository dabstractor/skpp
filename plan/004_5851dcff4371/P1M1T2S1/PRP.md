# PRP — P1.M1.T2.S1: Rewrite `usageText` + update all error prefix strings + doc comments (subcommands → flags, user-facing surfaces)

> **Subtask:** P1.M1.T2.S1 — the user-facing-surface half of P1.M1.T1's `check`/`init`/`completion` → `--check`/`--init`/`--completions` flag conversion. S1 (committed `594be07`) converted `parseArgs`; the parallel **S2** rewrites `exclusivityError`; **this subtask** sweeps every remaining surface: the `usageText` help text, the `skillsdir.ErrNotFound` message, the 8 `skilldozer init:` error-prefix strings, the prefix-quote comments, and the function doc comments.
> **Scope boundary:** String-literal + doc-comment edits ONLY in `main.go` (everywhere EXCEPT `parseArgs` [S1-done] and `exclusivityError` [S2, parallel]) plus the ONE `ErrNotFound` line in `internal/skillsdir/skillsdir.go`. Zero logic change, zero signature change, zero import change. Does NOT touch `main_test.go`/`skillsdir_test.go` (those are P1.M1.T3 — `go test` is EXPECTED RED).

---

## Goal

**Feature Goal**: Make every user-facing and developer-facing surface in `main.go` (and the one `ErrNotFound` message in `skillsdir`) reference the new `--check`/`--init`/`--completions` **flags** instead of the deleted bare subcommands, so the `--help` output, the error messages, and the doc comments all match what S1's `parseArgs` actually accepts (PRD §6 "no bare-word subcommands", §6.1 "Help & completions advertise long forms only", §8.2/§8.3/§6.4 which all say `--init`).

**Deliverable**: Edits to two existing files:
1. `main.go` — rewrite `usageText` (USAGE/EXAMPLES/OPTIONS blocks + add the long-form-only note); change the 8 `"skilldozer init: "` error-prefix strings → `"skilldozer --init: "`; update the 3 prefix-quote comments; update the `runCheck`/`detectShell`/`runCompletion` doc comments (the `completionScript` doc is a verified no-op); sweep the remaining `skilldozer init`/`skilldozer check` command-name prose in function doc comments.
2. `internal/skillsdir/skillsdir.go` — the single `ErrNotFound` message (`skilldozer init` → `skilldozer --init`).

**Success Definition**: `go build ./...` succeeds and `go vet ./...` passes; `go.mod`/`go.sum` byte-for-byte unchanged; `usageText` shows only `--flag` forms and includes the long-form-only note; no `"skilldozer init:"` (bare) error-prefix string remains; no `skilldozer check`/`skilldozer init`/`skilldozer completion` full-command-name reference remains in any doc comment; `go test` is EXPECTED RED (T3's scope — do not "fix" it here).

---

## User Persona (if applicable)

**Target User**: any user reading `skilldozer --help` or hitting an `init`/config error, and any developer reading the `main.go` doc comments. The conversion S1 already did means `skilldozer --check` works but `skilldozer check` is now a (resolved-by-tag-or-error) bare word — so the help text, the `run \`skilldozer --init\`` unconfigured hint, and the `skilldozer --init: <step>: …` error prefixes must stop advertising the old bare forms.

**Pain Points Addressed**: today `--help` still lists `skilldozer check` / `skilldozer init` / `skilldozer completion` (commands that no longer exist as subcommands), the unconfigured hint says `run \`skilldozer init\``, and the init error prefixes say `skilldozer init:` — all contradicted by S1's `parseArgs`. This subtask makes the surfaces honest.

---

## Why

- **PRD §6 + §17 + decision 19**: "There are no bare-word subcommands. Every non-tag action is a `--flag`." S1 drove this in `parseArgs`; the surfaces (`--help`, errors, docs) must follow or the binary lies about its own contract.
- **PRD §8.2/§8.3/§6.4** all literally say `run \`skilldozer --init\`` and `skilldozer --init`; the `ErrNotFound` message and the help text must match (the §13 acceptance greps for these).
- **Closing Change Groups 4, 5, 6** (code_change_map.md): usageText (4), error-prefix strings incl. `ErrNotFound` (5), function doc comments (6) — the exact bounded slice the map assigns to this subtask.
- **Mode-A honesty**: same rule the sibling S2 PRP applied (its GOTCHA: don't leave comments that lie). OUTPUT §4 explicitly says "all doc comments use flag language."

---

## What

A mechanical, region-by-region string + comment sweep. **No logic changes anywhere.** Concretely (all line numbers = current post-S1 `594be07`):

- **usageText** (main.go:71-117): USAGE block lines 80-82, EXAMPLES block lines 95-97, OPTIONS block lines 107-111 (the 3 bare `check`/`init [<dir>]`/`completion` lines) → `--check`/`--init [<dir>]`/`--completions [--shell <name>]`; add the long-form-only note before OPTIONS.
- **ErrNotFound** (skillsdir.go:275): `skilldozer init` → `skilldozer --init`.
- **8 error-prefix `fmt.Errorf` strings** (main.go:988, 992, 1014, 1077, 1082, 1087, 1090, 1097): `"skilldozer init: "` → `"skilldozer --init: "`.
- **3 prefix-quote comments** (main.go:1073, 1137, 1149): `"skilldozer init: …"` → `"skilldozer --init: …"`.
- **Function doc comments**: runCheck (618/622), detectShell (1221), runCompletion (1244/1247); `completionScript` (1095-1101) is a **verified no-op** (no `skilldozer completion` ref present).
- **Mode-A prose sweep** (OUTPUT §4): command-name references in chooseStore (895), resolveStore (998), exampleSkillTemplate (1021), setupStore (1051), runInit (1120), and the runInit check-report comments (1176/1178).

### Success Criteria

- [ ] `usageText` USAGE/EXAMPLES/OPTIONS show only `--check`/`--init [<dir>]`/`--completions [--shell <name>]`; the long-form-only note is present
- [ ] `skillsdir.ErrNotFound` message says `run \`skilldozer --init\``
- [ ] All 8 `"skilldozer init: "` error-prefix strings say `"skilldozer --init: "`
- [ ] The 3 prefix-quote comments say `"skilldozer --init: …"`
- [ ] `detectShell` (1221) and `runCompletion` (1244/1247) doc comments say `skilldozer --completions`
- [ ] `runCheck` doc (618/622) says `skilldozer --check` / `--check` flag
- [ ] No `skilldozer check`/`skilldozer init`/`skilldozer completion` full-command-name reference remains in any doc comment (the §7 sweep)
- [ ] `completionScript`'s doc is NOT pointlessly edited (no `skilldozer completion` ref exists there)
- [ ] `parseArgs` (179-369), the `config` struct (148-168), and `exclusivityError` (750-818) are UNTOUCHED
- [ ] `go build ./...` succeeds; `go vet ./...` passes; `go.mod`/`go.sum` unchanged; `main_test.go`/`skillsdir_test.go` UNTOUCHED

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit site is pinned to its CURRENT line number with the exact before→after text transcribed in `research/verified_facts.md` §2-§7; the stale-line-number hazard (§0), the `completionScript` no-op (§6a), the ownership boundary with S2 (§1), and the EXPECTED-RED test list (§8) are all documented. An implementer who has never seen this repo can do it in one pass by matching the given old→new strings (locate by content, not by the contract's stale line numbers).

### Documentation & References

```yaml
# MUST READ — the verified site inventory (current line numbers + exact before/after)
- file: plan/004_5851dcff4371/P1M1T2S1/research/verified_facts.md
  why: "THE source of truth. §0 explains why the contract/change-map line numbers are stale (pre-S1). §2-§7 list every site with its CURRENT line + exact old→new text. §6a proves completionScript's doc is a NO-OP (no 'skilldozer completion' ref). §1 pins the S1/S2/T2.S1 boundary (no conflicts). §8 lists the tests that go EXPECTED RED (T3's scope)."
  critical: "§0 (line drift) and §6a (completionScript no-op) are the two facts that prevent the most likely errors: editing wrong lines, and hunting for a phantom edit."

# MUST READ — the parallel sibling PRP (S2) — defines the boundary + the gate stance
- file: plan/004_5851dcff4371/P1M1T1S2/PRP.md
  why: "S2 owns ONLY exclusivityError (body 769-818 + doc 750-768); its GOTCHA #11 EXPLICITLY defers usageText/error-prefixes/runCheck-doc/completion-docs to T2.S1 (this subtask). S2's GOTCHA #8 establishes the gate stance: go build + go vet ONLY; go test EXPECTED RED (S1 drift); fixing tests is P1.M1.T3. T2.S1 inherits both the boundary and the gate."

# MUST READ — the change map (Change Groups 4/5/6 are this subtask's spec)
- file: plan/004_5851dcff4371/architecture/code_change_map.md
  why: "Change Group 4 (usageText), 5 (error prefixes incl. ErrNotFound), 6 (function doc comments) pin the old→new strings. NOTE: its line numbers (71-117 / 1001-1110 / 1112-1271) are pre-S1 and shifted up ~13 — use its STRING TEXT, not its line numbers (verified_facts §0). 6a claims completionScript needs a change — that's STALE/WRONG (verified_facts §6a: no-op)."
  section: "Change Groups 4, 5, 6."

# MUST READ — the edit target (the two files)
- file: main.go
  why: "THE primary edit target. usageText @71-117; runCheck doc @618-630; chooseStore doc @895; resolveStore err prefixes @988/992/1014 + comment @998; exampleSkillTemplate doc @1021; setupStore doc @1051 + prefix-quote @1073 + err prefixes @1077-1097; completionScript doc @1095-1101 (NO-OP); runInit doc @1120 + comments @1137/1149 + check-report comments @1176/1178; detectShell doc @1221; runCompletion doc @1244/1247. DO NOT touch parseArgs (179-369), config struct (148-168), or exclusivityError (750-818)."
  pattern: "Existing doc-comment style: backtick-quoted command names, PRD §-citations, `//` prose explaining WHY. Match it. usageText is a raw string literal — gofmt does NOT reformat inside it, so manually re-align the OPTIONS description column after the flag-name swaps."
  gotcha: "Raw-string-literal edits need manual column alignment (gofmt won't help). The flag names are WIDER than the bare words, so the description column shifts right."

- file: internal/skillsdir/skillsdir.go
  why: "The ONE internal/ change: ErrNotFound @275. Change the message, nothing else in this file."
  gotcha: "skillsdir_test.go:526-530 asserts this EXACT message — it goes RED after this change (T3 fixes it). Do NOT touch skillsdir_test.go."

# READ-ONLY — the PRD authority
- file: PRD.md
  why: "READ-ONLY. §6 header + §6.1 (flags table: --check/--init/--completions; 'Help & completions advertise long forms only'); §6.4 ('run \`skilldozer --init\`'); §8.2/§8.3 ('skilldozer --init', 'run \`skilldozer --init\`'); §17 ('check/init/completions are --check/--init/--completions'). These are the authority the new strings must match. Do NOT edit PRD.md."
  section: "h2.5 (§6), h3.1 (§6.1), h3.4 (§6.4), h2.7/h3.9/h3.10 (§8.2/§8.3), h2.16 (§17)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/004_5851dcff4371/tasks.json
  why: "P1.M1.T2.S1's CONTRACT block is authoritative INPUT/LOGIC/OUTPUT. This PRP transcribes it; tasks.json wins on conflict (and this PRP documents where the contract's line numbers/no-op claims are stale vs the current code)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD && wc -l main.go
594be07   # S1 committed ("Replace bare subcommands with flags in parseArgs")
1275 main.go
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK
BUILD_OK / VET_OK   # green; go test is RED from S1's bare→flag drift (T3's scope)
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# exclusivityError @ 769-818 (doc 750-768) → S2's territory, DO NOT TOUCH.
# parseArgs flag cases present: --check @296, --completions @299, --init @327 (S1 done).
```

### Desired Codebase tree with files to be changed

```bash
main.go                            # MODIFY — usageText + 8 err prefixes + comments + function docs (everywhere except parseArgs/config/exclusivityError)
internal/skillsdir/skillsdir.go    # MODIFY — the ONE ErrNotFound message line (275)
# go.mod / go.sum — UNCHANGED (string-literal + doc-comment edits add no imports)
# main_test.go / skillsdir_test.go — UNCHANGED (P1.M1.T3; go test EXPECTED RED)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main.go` (usageText) | USAGE/EXAMPLES/OPTIONS → flags + long-form-only note | Contract LOGIC a |
| `main.go` (resolveStore/setupStore) | 8 `"skilldozer init: "` → `"skilldozer --init: "` | Contract LOGIC c |
| `main.go` (comments) | 3 prefix-quote comments + function doc comments + Mode-A prose sweep | Contract LOGIC d/e/f + OUTPUT §4 |
| `internal/skillsdir/skillsdir.go` | `ErrNotFound` message `--init` | Contract LOGIC b |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — LINE NUMBERS ARE STALE. The contract (71-117 / 1001-1110 / 1112-1271) and
// code_change_map.md cite PRE-S1 numbers (HEAD f30d5c5, 1288 lines); S1 (committed 594be07)
// net-deleted ~13 lines, shifting everything after parseArgs UP. LOCATE EVERY EDIT BY CONTENT
// (match the exact old strings in research/verified_facts.md §2-§7). Do NOT edit by the
// contract's line numbers or you will hit the wrong lines. (verified_facts §0.)

// GOTCHA #2 — usageText is a RAW STRING LITERAL (backticks). gofmt does NOT reformat inside
// it. After swapping check→--check, init [<dir>]→--init [<dir>], completion→--completions
// [--shell <name>], the flag names are WIDER, so the description column in USAGE/EXAMPLES/
// OPTIONS shifts. Manually re-align the descriptions to keep it readable (match the existing
// visual style). Do NOT expect gofmt to fix the alignment — it can't see inside the string.

// GOTCHA #3 — completionScript's doc (1095-1101) is a NO-OP. The contract LOGIC(e)/change-map
// 6a says "update skilldozer completion → skilldozer --completions" there, but VERIFIED by
// grep: completionScript's doc has ZERO "skilldozer completion" references (it talks about
// //go:embed / byte-identity, not the command). Do NOT search for a phantom edit. Only
// detectShell (1221) and runCompletion (1244/1247) have the reference. (verified_facts §6a.)

// GOTCHA #4 — The gate is go build + go vet ONLY. go test is EXPECTED RED: S1's bare→flag
// parseArgs drift already reddened the run-level tests, and THIS subtask's string changes
// (usageText, the 8 error prefixes, ErrNotFound) redden ~10 more (verified_facts §8 lists
// them). The contract OUTPUT §4 says "go build ./... must succeed" — NOT "go test passes".
// P1.M1.T3 is the dedicated test-flip milestone. Do NOT touch main_test.go or
// skillsdir_test.go to make them green. (Inherited from sibling S2's GOTCHA #8.)

// GOTCHA #5 — DO NOT touch exclusivityError (750-818) or parseArgs (179-369) / config struct
// (148-168). exclusivityError is S2's ONLY territory (parallel, in-progress); parseArgs + the
// config struct doc comments are S1's (done, committed 594be07). S2's GOTCHA #11 explicitly
// defers usageText/error-prefixes/runCheck-doc/completion-docs to THIS subtask — the boundary
// is clean, but a careless "replace all skilldozer init → skilldozer --init across main.go"
// would land inside exclusivityError's comments and collide with S2. Scope each edit to the
// exact sites in research/verified_facts.md §2-§7.

// GOTCHA #6 — --completions is PLURAL (decision 19 / S1). Write "--completions" everywhere,
// never "--completion". (The old bare subcommand was singular "completion"; the new flag is
// plural.) This applies to usageText (line 82, 97, 110), detectShell (1221), runCompletion
// (1244/1247). Mirrors S2's GOTCHA #6.

// GOTCHA #7 — The long-form-only note (contract LOGIC a) goes EITHER "after USAGE or before
// OPTIONS". Before OPTIONS is the most natural spot (the note qualifies how OPTIONS are
// advertised). Either is acceptable per the contract. Keep it to ~2 lines, citing §6.1, and
// listing the exact short-alias set (-a -l -s -f -p -h -v) verbatim from the PRD.

// GOTCHA #8 — The init error PREFIX is "skilldozer init: " (note the trailing space + colon).
// Replace the WHOLE token "skilldozer init:" → "skilldozer --init:" — do not partially edit
// (e.g. "skilldozer init" → "skilldozer --init" leaving the ": " dangling, or vice versa).
// Match each of the 8 exact strings in verified_facts §4. The 3 COMMENT sites (1073/1137/1149)
// quote the prefix inside backticks/double-quotes — update those quoted forms identically.

// GOTCHA #9 — ErrNotFound lives in internal/skillsdir, NOT main.go. It is the ONE internal/
// change. Its exact-message test (skillsdir_test.go:526-530) will go RED — that is expected
// (T3 / test-flip scope). Do not "preserve" the old message to keep the test green; the PRD
// (§8.2/§8.3/§6.4) mandates `run \`skilldozer --init\``.

// GOTCHA #10 — No deps/import change. Every edit is a string literal or a // comment. go.mod
// and go.sum must be byte-for-byte identical. (Carries over from S1/S2.)

// GOTCHA #11 — gofmt -l main.go internal/skillsdir/skillsdir.go must print NOTHING after the
// edits (the doc-comment edits are normal // lines gofmt reformats; usageText's raw-string
// interior is untouched by gofmt). Run gofmt -w if it lists a file.
```

---

## Implementation Blueprint

### Data models and structure

**None.** No structs, no signatures, no imports change. Pure string-literal + comment edits. `config` fields (`c.check`/`c.init`/`c.completion`) are unchanged (S1 left them); `runInit`/`setupStore`/`resolveStore`/`runCheck`/`detectShell`/`runCompletion`/`completionScript` keep their signatures.

### Implementation Tasks (ordered by dependencies — region by region)

```yaml
Task 1: REWRITE usageText (main.go:71-117)
  - FILE: main.go, the `const usageText = \`...\`` raw string literal.
  - EDIT USAGE block (lines 80-82): the 3 bare lines →
        skilldozer --check
        skilldozer --init [<dir>]
        skilldozer --completions [--shell <name>]
  - EDIT EXAMPLES block (lines 95-97): the 3 lines →
        skilldozer --check                   # validate every skill on disk
        skilldozer --init --store <dir>      # non-interactive first-run setup
        eval "$(skilldozer --completions)"   # load completions into your shell
  - EDIT OPTIONS block (lines 107-111): the 3 bare lines →
        --check            Validate every skill on disk (report OK / WARN / ERROR)
        --init [<dir>]     First-run setup: pick/create the skills store and write the config
        --completions [--shell <name>]   Emit the shell completion script for eval (§14.6)
    (the `--store <dir>` line and every other OPTIONS line are UNCHANGED.)
  - ADD the long-form-only note BEFORE the OPTIONS: heading (after the EXAMPLES block's blank line):
        Help and --completions advertise long forms only; short aliases (-a, -l, -s, -f, -p, -h, -v)
        remain valid for typing but are not advertised (§6.1).
  - GOTCHA #2: manually re-align the OPTIONS description column (the flag names are wider).
    GOTCHA #6: --completions is PLURAL.

Task 2: EDIT internal/skillsdir/skillsdir.go — the ErrNotFound message (line 275)
  - FILE: internal/skillsdir/skillsdir.go (the ONLY internal/ change)
  - CHANGE: `skilldozer is not configured; run \`skilldozer init\`` →
            `skilldozer is not configured; run \`skilldozer --init\``
  - GOTCHA #9: skillsdir_test.go:526-530 goes RED — expected (T3). Do NOT touch the test.

Task 3: EDIT main.go — the 8 error-prefix fmt.Errorf strings (LOGIC c)
  - FILE: main.go. Replace `"skilldozer init: "` → `"skilldozer --init: "` in EXACTLY these 8:
      988  resolveStore:  fmt.Errorf("skilldozer --init: resolve cwd: %w", err)
      992  resolveStore:  fmt.Errorf("skilldozer --init: resolve default store: %w", err)
      1014 resolveStore:  fmt.Errorf("skilldozer --init: absolutize store: %w", err)
      1077 setupStore:    fmt.Errorf("skilldozer --init: create store dir %q: %w", store, err)
      1082 setupStore:    fmt.Errorf("skilldozer --init: read store dir %q: %w", store, err)
      1087 setupStore:    fmt.Errorf("skilldozer --init: create example dir: %w", err)
      1090 setupStore:    fmt.Errorf("skilldozer --init: seed example SKILL.md: %w", err)
      1097 setupStore:    fmt.Errorf("skilldozer --init: write config %q: %w", configPath, err)
  - GOTCHA #8: replace the whole "skilldozer init:" token. Do NOT touch the format verbs (%w/%q).
    GOTCHA #5: these are in resolveStore (3) + setupStore (5) — NOT exclusivityError.

Task 4: EDIT main.go — the 3 prefix-quote comments (LOGIC d)
  - 1073 (setupStore doc tail): `...wrapped with a "skilldozer init: <step>: %w" prefix.`
        → `...wrapped with a "skilldozer --init: <step>: %w" prefix.`
  - 1137 (runInit, resolveStore err branch): `// one-line (resolveStore wraps with "skilldozer init: …")`
        → `// one-line (resolveStore wraps with "skilldozer --init: …")`
  - 1149 (runInit, setupStore err branch): `// setupStore wraps with "skilldozer init: …"`
        → `// setupStore wraps with "skilldozer --init: …"`

Task 5: EDIT main.go — function doc comments (LOGIC e + f)
  - runCheck (618): `// \`skilldozer check\` subcommand (PRD §9)...` → `// \`skilldozer --check\` flag (PRD §9)...`
  - runCheck (622): `// so \`if skilldozer check; then …\` works as a gate).` → `\`if skilldozer --check; then …\``
  - detectShell (1221): `// detectShell resolves the target shell for \`skilldozer completion\` (PRD §14.6` → `--completions`
  - runCompletion (1244): `// runCompletion is the \`skilldozer completion\` handler (PRD §14.6 / §6.4).` → `--completions`
  - runCompletion (1247): `// ...for \`eval "$(skilldozer completion)"\`` → `\`eval "$(skilldozer --completions)"\``
  - completionScript (1095-1101): NO-OP (GOTCHA #3 — verified no `skilldozer completion` ref). Do nothing.

Task 6: EDIT main.go — Mode-A prose sweep (OUTPUT §4 "all doc comments use flag language")
  - Sweep every remaining `skilldozer init` / `skilldozer check` full-command-name ref in doc
    comments → the --flag form. Exact sites (verified_facts §7):
      895  (chooseStore doc):        `for \`skilldozer init\`` → `for \`skilldozer --init\``
      998  (resolveStore comment):   `store="$(skilldozer init)"` → `store="$(skilldozer --init)"`
      1021 (exampleSkillTemplate):   `skilldozer init writes this verbatim` → `skilldozer --init writes this verbatim`
      1051 (setupStore doc):         `half of \`skilldozer init\`` → `half of \`skilldozer --init\``
      1120 (runInit doc):            `runInit is the \`skilldozer init\` orchestrator` → `\`skilldozer --init\` orchestrator`
      1176 (runInit comment):        `(6) \`skilldozer check\` report on the effective store` → `\`skilldozer --check\` report`
      1178 (runInit comment):        `the standalone \`check\` subcommand keeps its report` → `the standalone \`--check\` flag keeps its report`
  - EXCLUDED (positional-form shorthand, leave to avoid churn): `init ~/x` / `init ~/myskills`
    in expandHome (978-985) + resolveStore (986-989); `the \`check\` report to stderr` at 1125.
    (If you want full consistency while in those blocks: `init ~/x`→`--init ~/x`, `init <dir>`→
    `--init <dir>`. Not contract-pinned.)
  - GOTCHA #5: NONE of these sites are in exclusivityError (750-818) or parseArgs (179-369).

Task 7: VERIFY build + vet (the ONLY hard gate — GOTCHA #4)
  - COMMAND: gofmt -l main.go internal/skillsdir/skillsdir.go   (must print NOTHING)
  - COMMAND: go build ./...   (exit 0)
  - COMMAND: go vet ./...     (exit 0)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"
  - INVARIANT: grep -c '"skilldozer init:' main.go           (expect 0 — all 8 prefixes flipped)
  - INVARIANT: grep -c 'skilldozer init:' main.go            (expect 0 — incl. the 3 comments)
  - INVARIANT: grep -c 'skilldozer check\|skilldozer completion' main.go  (expect 0 outside the
               excluded shorthand — verify the remaining hits are ONLY the positional `init ~/x`
               shorthand you deliberately left, if any)
  - INVARIANT: git diff main.go internal/skillsdir/skillsdir.go | grep -c '^[+-].*exclusivityError\|^[+-].*case "--check"\|^[+-].*case "--init"\|^[+-].*case "--completions"'  (expect 0 — did NOT touch parseArgs or exclusivityError)
```

### Implementation Patterns & Key Details

```go
// The usageText swaps are inside a RAW STRING LITERAL — gofmt cannot reflow the interior, so
// the description column must be re-aligned by hand. Example (OPTIONS block, before → after):
//   BEFORE:   check              Validate every skill on disk (report OK / WARN / ERROR)
//   AFTER:    --check            Validate every skill on disk (report OK / WARN / ERROR)
// (the "Validate..." text starts a few columns further right because "--check" is wider than
//  "check"; pad so the descriptions line up with the neighbouring --store/--path rows.)

// The error-prefix swaps are one-token replacements (GOTCHA #8 — whole "skilldozer init:" token):
//   BEFORE:  return "", fmt.Errorf("skilldozer init: resolve cwd: %w", err)
//   AFTER:   return "", fmt.Errorf("skilldozer --init: resolve cwd: %w", err)
// (the %w / %q verbs and the arg list are UNCHANGED — only the prefix string changes.)

// The ErrNotFound swap (the one internal/ change):
//   BEFORE:  var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer init`")
//   AFTER:   var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer --init`")
```

Notes easy to get wrong:
- `usageText` is a `const ... = \`...\`` — a multi-line raw string. Edits inside it must preserve the backtick fencing and the trailing newline before the closing backtick. gofmt won't catch a misaligned OPTIONS column; eyeball it.
- The 8 error prefixes are split 3 (resolveStore) + 5 (setupStore). Don't miss the 3 in resolveStore (988/992/1014) — they're ~25 lines apart from the setupStore cluster.
- `completionScript`'s doc genuinely needs nothing (GOTCHA #3) — resist the urge to "find" a `skilldozer completion` reference there; it isn't present.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Gate = build + vet, NOT test? → YES.** The contract OUTPUT §4 says "go build must succeed" only. S1's parseArgs drift already reddened the run-level tests; this subtask's string changes redden ~10 more. P1.M1.T3 is the dedicated test-flip milestone. Inherited cleanly from sibling S2's GOTCHA #8. Editing tests here = scope creep into T3.
2. **completionScript doc (contract 6a) — edit anyway? → NO.** Verified by grep: zero `skilldozer completion` references in its doc. The contract/change-map claim is stale. Documenting it as a no-op prevents a phantom-edit search.
3. **Mode-A prose sweep (OUTPUT §4) — include? → YES, the full-command-name refs; EXCLUDE positional shorthand.** OUTPUT §4 says "all doc comments use flag language." The 7 full-command-name sites (§7) are stale and in functions T2.S1 already edits — no sibling conflict. The `init ~/x`/`init <dir>` positional shorthand and `check report` shorthand are lower-value (they reference the flag's arg form, not the command surface) — left to avoid churn, noted as optional.
4. **Long-form-only note placement? → before OPTIONS.** The note qualifies how OPTIONS are advertised; "before OPTIONS" is the contract's stated option and the most natural spot. Either placement is contract-acceptable.
5. **`--completions` plural everywhere? → YES.** Decision 19 / S1. Never `--completion`. (Mirrors S2's GOTCHA #6.)

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports. (GOTCHA #10)

OWNERSHIP (no conflicts):
  - S1 (done, 594be07): parseArgs (179-369) + config struct doc (148-168).
  - S2 (parallel): exclusivityError ONLY (body 769-818 + doc 750-768). Its GOTCHA #11 defers
    everything T2.S1 touches to T2.S1.
  - T2.S1 (this): usageText, ErrNotFound, the 8 error prefixes, all doc comments outside
    parseArgs/config/exclusivityError.
  - T3 (later): main_test.go + skillsdir_test.go — flips the EXPECTED-RED tests.

CALLERS/CONSUMERS (unchanged):
  - run() dispatch (c.check→runCheck, c.init→runInit, c.completion→runCompletion) is UNCHANGED —
    the config fields still drive the same functions; only the strings/docs change.
  - main.go prints ErrNotFound verbatim (fmt.Fprintln(stderr, err)) — the new --init message
    reaches the user with no main.go plumbing change.

NO ROUTES / NO DATABASE / NO CONFIG SCHEMA / NO COMPLETIONS:
  - T2.S1 changes strings + comments only. The completion FILES (completions/*.bash/_skilldozer/
    *.fish) are P1.M2.T1's scope — NOT touched here.
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (the ONLY hard gate — GOTCHA #4)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go internal/skillsdir/skillsdir.go   # must print NOTHING (gofmt -w if it lists a file)
go build ./...   # expect exit 0
go vet ./...     # expect exit 0

# No bare "skilldozer init:" error-prefix string remains (all 8 flipped):
grep -c '"skilldozer init:' main.go              # Expected: 0
# No "skilldozer init:" anywhere (incl. the 3 comments):
grep -c 'skilldozer init:' main.go               # Expected: 0
# No bare subcommand full-command-name in doc comments (only positional shorthand may remain):
grep -n 'skilldozer check\|skilldozer init\|skilldozer completion' main.go
# Expected: each remaining hit is a DELIBERATELY-left positional shorthand (`init ~/x` etc.)
# OR an already-correct `--init`/`--check`/`--completions` substring — NOT a bare command.
# The new flag forms ARE present:
grep -c 'skilldozer --init:\|skilldozer --check\|skilldozer --completions' main.go  # Expected: >0
```

### Level 2: usageText content spot-check (the user-facing surface)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# --help now shows only --flag forms + the long-form-only note:
/tmp/sdz --help > /tmp/h.txt 2>&1
grep -q -- '--check'                /tmp/h.txt && echo "OK: --check in USAGE/OPTIONS"     || echo "FAIL: no --check"
grep -q -- '--init \[<dir>\]'       /tmp/h.txt && echo "OK: --init [<dir>]"               || echo "FAIL: no --init"
grep -q -- '--completions \[--shell' /tmp/h.txt && echo "OK: --completions [--shell ...]" || echo "FAIL: no --completions"
grep -qi 'long forms only'          /tmp/h.txt && echo "OK: long-form-only note present"  || echo "FAIL: no long-form note"
# NO bare forms remain in --help:
if grep -qE '^  skilldozer (check|init|completion)( |$)' /tmp/h.txt; then
  echo "FAIL: bare subcommand still in --help"; grep -nE '^  skilldozer (check|init|completion)' /tmp/h.txt
else echo "OK: no bare subcommand in --help"; fi
rm -f /tmp/sdz /tmp/h.txt
```

### Level 3: Error-prefix + ErrNotFound message spot-check (the user-facing errors)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# ErrNotFound (unconfigured) now says --init (run from an unconfigured cwd):
TMP=$(mktemp -d); cd "$TMP"
err=$(env -u SKILLDOZER_SKILLS_DIR -u SKILLDOZER_CONFIG /tmp/sdz nonexistenttag 2>&1 1>/dev/null)
echo "$err" | grep -q 'skilldozer --init' && echo "OK: ErrNotFound says --init" || echo "FAIL: $err"

# An init error prefix now says "skilldozer --init:" (force a setupStore failure: unwritable
# config path). Point SKILLDOZER_CONFIG at a path whose parent is a FILE -> Save fails:
cd /tmp; touch /tmp/blocker_file
err=$(SKILLDOZER_CONFIG=/tmp/blocker_file/cfg.yaml env -u SKILLDOZER_SKILLS_DIR \
      /tmp/sdz --init --store /tmp/sdz_dummy_store </dev/null 2>&1 1>/dev/null)
echo "$err" | grep -q 'skilldozer --init:' && echo "OK: error prefix says --init:" || echo "FAIL: $err"
rm -f /tmp/blocker_file; rm -rf /tmp/sdz_dummy_store "$TMP" /tmp/sdz
# Expected: both print "OK: ...". (The exit codes themselves are unchanged by T2.S1 — only the strings.)
```

### Level 4: Whole-module build + dependency + scope invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # Expected: 0
go vet  ./...  ; echo "vet exit $?"     # Expected: 0

# go test is EXPECTED RED (S1 drift + this subtask's string changes). Confirm the RED set is
# the EXPECTED one (usageText/error-prefix/ErrNotFound assertions), NOT a compile error:
go test ./... 2>&1 | grep -E '^--- FAIL' | head -20
# Expected: TestRunInit* / TestExclusivity* / TestParseArgs* / the ErrNotFound-message test —
# all substring/exact-message assertions on the OLD strings. NONE should be a COMPILE error.

git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
# Expected: "deps unchanged".

# Scope invariants (prove T2.S1 stayed in its lane):
git diff main.go internal/skillsdir/skillsdir.go | grep -E '^\+' | grep -cE 'exclusivityError|func parseArgs|case "--check"|case "--init"|case "--completions"'
# Expected: 0 (did NOT add/touch parseArgs flag cases or exclusivityError).
git diff --stat   # Expected: only main.go + internal/skillsdir/skillsdir.go changed.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go build ./...` exit 0, `go vet ./...` exit 0; `grep '"skilldozer init:'` = 0; `grep 'skilldozer init:'` = 0
- [ ] Level 2 PASS — `--help` shows `--check`/`--init [<dir>]`/`--completions [--shell ...]` + the long-form-only note; no bare subcommand lines
- [ ] Level 3 PASS — ErrNotFound says `skilldozer --init`; an init error prefix says `skilldozer --init:`
- [ ] Level 4 PASS — build+vet exit 0; `go test` failures are the EXPECTED substring/message set (no compile errors); `git diff go.mod go.sum` → "deps unchanged"; scope invariants hold (no parseArgs/exclusivityError additions)

### Feature Validation
- [ ] `usageText` USAGE/EXAMPLES/OPTIONS show only `--flag` forms; long-form-only note present
- [ ] `skillsdir.ErrNotFound` message says `run \`skilldozer --init\``
- [ ] All 8 `"skilldozer init: "` error-prefix strings say `"skilldozer --init: "`
- [ ] The 3 prefix-quote comments (1073/1137/1149) say `"skilldozer --init: …"`
- [ ] `detectShell` (1221) + `runCompletion` (1244/1247) docs say `skilldozer --completions`
- [ ] `runCheck` doc (618/622) says `skilldozer --check` / `--check` flag
- [ ] `completionScript`'s doc is NOT pointlessly edited (verified no-op)
- [ ] The §7 Mode-A prose sweep done (no `skilldozer init`/`skilldozer check` full-command-name refs in doc comments)

### Code Quality / Convention Validation
- [ ] Matches the existing doc-comment style (backtick-quoted names, PRD §-citations)
- [ ] `--completions` is PLURAL everywhere (never `--completion`)
- [ ] usageText OPTIONS column re-aligned by hand (gofmt can't see inside the raw string)
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] Minimal, mechanical diff (string-literal + comment swaps only)

### Scope Discipline (the S1/S2/T3 boundaries)
- [ ] Did NOT touch `parseArgs` (179-369) or the `config` struct (148-168) — S1's territory
- [ ] Did NOT touch `exclusivityError` (750-818) — S2's parallel territory
- [ ] Did NOT touch `main_test.go` or `skillsdir_test.go` — P1.M1.T3's scope (go test EXPECTED RED)
- [ ] Did NOT touch the completion FILES (`completions/*`) — P1.M2.T1's scope
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't edit by the contract's/change-map's line numbers.** They're pre-S1 (HEAD `f30d5c5`, 1288 lines); S1 shifted everything up ~13. Locate every edit by content (match the exact old strings in `research/verified_facts.md`). (GOTCHA #1)
- ❌ **Don't search for a `skilldozer completion` edit in completionScript's doc.** Verified: none exists (GOTCHA #3). Only detectShell + runCompletion have the reference.
- ❌ **Don't run `go test` as a gate or "fix" the red tests.** The gate is `go build` + `go vet`. The tests are EXPECTED RED (S1 drift + this subtask's strings); P1.M1.T3 flips them. (GOTCHA #4)
- ❌ **Don't "replace all `skilldozer init` → `skilldozer --init` across main.go" blindly.** That would land inside `exclusivityError`'s comments (750-818, S2's territory) and collide. Scope each edit to the exact sites in verified_facts §2-§7. (GOTCHA #5)
- ❌ **Don't write `--completion` (singular).** The flag is `--completions` (PLURAL, decision 19). (GOTCHA #6)
- ❌ **Don't partially edit the `"skilldozer init:"` prefix.** Replace the whole token `"skilldozer init:"` → `"skilldozer --init:"`; keep the `%w`/`%q` verbs and args intact. (GOTCHA #8)
- ❌ **Don't preserve the old ErrNotFound message to keep its test green.** The PRD mandates `run \`skilldozer --init\``; the exact-message test going RED is expected (T3). (GOTCHA #9)
- ❌ **Don't expect gofmt to align usageText's OPTIONS column.** It's a raw string literal; gofmt can't see inside. Re-align the descriptions by hand. (GOTCHA #2)
- ❌ **Don't add deps or imports.** Every edit is a string literal or a comment. go.mod/go.sum byte-for-byte identical. (GOTCHA #10)
- ❌ **Don't touch the completion files or main_test.go.** Those are P1.M2.T1 and P1.M1.T3 respectively.

---

## Confidence Score

**9.5/10** — Every edit site is pinned to its CURRENT (post-S1, HEAD `594be07`) line number with the exact before→after string transcribed in `research/verified_facts.md` §2-§7; the two hazards that would otherwise cause failure (stale line numbers §0; the completionScript no-op §6a) are explicitly documented; the S2 ownership boundary is clean (S2's GOTCHA #11 defers everything T2.S1 touches); and the build+vet gate + EXPECTED-RED test stance are inherited cleanly from the already-committed S1 / parallel S2. The work is mechanical string + comment swaps with zero logic change and zero import change. The 0.5 reservation is the Mode-A prose sweep (Task 6): OUTPUT §4 ("all doc comments use flag language") supports it, but a strict reader could limit scope to the contract's explicitly-named (a-f) sites and defer the §7 prose — the PRP includes the sweep (it's low-risk, no sibling conflict, and makes the code honest) but flags it as separable.
