# PRP — P1.M1.T1.S2: zsh — add `NO_LIST_AMBIGUOUS` to the derived eval registration

> **Subtask:** P1.M1.T1.S2 — the zsh half of P1.M1.T1 (§14.7 / decision 22: list every ambiguous match on the first Tab; never a silent halt at the common prefix). Mirrors the already-scoped S1 (bash) for zsh, but zsh's option lives in a **main.go const** (not the autoload file) because zsh is a **derived** emit path.
> **Scope boundary:** Edits ONLY `main.go`: the `zshEvalRegistration` raw-string const (body + its doc comment), plus an optional one-clause accuracy tweak to `zshEvalScript`'s doc. Does NOT touch `completionScript`, `completions/_skilldozer` (the autoload file — simplest path leaves it byte-identical), `completions/skilldozer.bash` (S1), `completions/skilldozer.fish`, `runCompletion` logic, any test (the byte-level assertion is P1.M1.T2.S1), or the README (P1.M1.T3).

---

## Goal

**Feature Goal**: Make the zsh **derived eval wrapper** (what `eval "$(skilldozer --completions)"` runs) set `setopt NO_LIST_AMBIGUOUS`, so an ambiguous prefix lists **all** matches on the first Tab instead of completing the common prefix and halting silently — fulfilling PRD §14.7 for zsh, with the change disclosed in comments and a one-line opt-out.

**Deliverable**: Edits to `main.go` only:
1. **Const body** (`zshEvalRegistration`, main.go:1260): append a disclosure comment block + an active `setopt NO_LIST_AMBIGUOUS` line + a commented `setopt LIST_AMBIGUOUS` opt-out (after the compdef block, before the closing raw-string backtick).
2. **Doc comment** on `zshEvalRegistration` (main.go:1257): broaden to mention it now also sets the §14.7 listing option (not just compdef).
3. **(Optional, minimal)** one clause in `zshEvalScript`'s doc (main.go:1240) for accuracy.

**Success Definition**: `go build ./...` succeeds; `completionScript("zsh")` stays byte-identical to `completions/_skilldozer` (`TestEmbeddedCompletionsMatchOnDisk` green); the 3 other zsh tests stay green; after rebuild, `./skilldozer --completions --shell zsh` output contains `setopt NO_LIST_AMBIGUOUS` and the opt-out token `setopt LIST_AMBIGUOUS`; `go.mod`/`go.sum` unchanged.

---

## User Persona (if applicable)

**Target User**: zsh users who tab-complete `skilldozer` via `eval "$(skilldozer --completions)"` (the primary discovery path for a manifest-free store).

**Use Case**: A user types `skilldozer a<Tab>` (with tags `agent-browser`, `agent-builder`) and immediately sees both candidates — not a frozen `agent-b` with nothing shown.

**Pain Points Addressed**: zsh defaults to `LIST_AMBIGUOUS` ON, so the first Tab completes the common prefix and lists only once you've typed to the exact ambiguous point — a silent halt that hides the very tags the user is trying to discover.

---

## Why

- **PRD §14.7**: "A completion that completes the longest common prefix and then stops with nothing shown is a defect." The store is manifest-free (§2), so the user often doesn't know a tag ahead of time — discovery-via-completion is primary. An ambiguous prefix that hides candidates defeats that.
- **§14.7 zsh half**: `LIST_AMBIGUOUS` is **on by default** in zsh; `setopt NO_LIST_AMBIGUOUS` (with the default `AUTO_LIST`) makes the first Tab list all prefix matches immediately. There is **no per-command scope** (a scoped `zstyle … menu select` does NOT list on the first Tab — verified in the PRD), so the global option is the only lever.
- **Why the const, not the autoload file**: zsh is a **derived** emit path. `completionScript("zsh")` returns the on-disk `completions/_skilldozer` verbatim (locked by `TestEmbeddedCompletionsMatchOnDisk`); `runCompletion` derives the eval wrapper via `zshEvalScript`, which appends `zshEvalRegistration`. The §14.7 option belongs in that eval-time append — leaving the autoload file and the byte-identity lock structurally untouched. (The fpath/manual parity is OPTIONAL and skipped — simplest path.)
- **Disclosure + opt-out required**: the option is **session-global** (it changes listing for *every* command in the shell, not just skilldozer). PRD §14.7 mandates the emitted script (a) set it, (b) disclose it in comments, and (c) provide a one-line opt-out (`setopt LIST_AMBIGUOUS`). This subtask delivers all three for zsh.
- **Decision 22**: "First-Tab list-all-matches; never a silent halt at the common prefix" — disclosed and opt-out-able.

---

## What

The `zshEvalRegistration` const (appended to the stripped autoload body by `zshEvalScript`, emitted for zsh only by `runCompletion`) gains, after the compdef block: a disclosure comment block + one active `setopt NO_LIST_AMBIGUOUS` + a commented `setopt LIST_AMBIGUOUS` opt-out.

Because this const is what `eval "$(skilldozer --completions)"` runs, the option is set at eval time in the user's shell. The autoload file (`completions/_skilldozer`) and `completionScript("zsh")` are unchanged — the §14.5 manual fpath path is unaffected, and the byte-identity lock holds.

No behavior change to the completion function itself (`_skilldozer` is untouched); the tag/flag candidate sets it offers are complete (§14.7 half #1, already true). This adds half #2 — making zsh *show* them on the first Tab.

### Success Criteria

- [ ] the `zshEvalRegistration` const body contains `setopt NO_LIST_AMBIGUOUS` (active line) after the compdef block
- [ ] the const body contains the opt-out token `setopt LIST_AMBIGUOUS` as a COMMENT (so it does not cancel the active line)
- [ ] the disclosure comment names `NO_LIST_AMBIGUOUS`, notes it is session-global, notes the zsh default (LIST_AMBIGUOUS ON), and gives the opt-out
- [ ] the const body contains NO backticks (stays a valid Go raw string literal)
- [ ] the `zshEvalRegistration` doc comment mentions the §14.7 listing option (not just compdef)
- [ ] `go build ./...` succeeds; `TestEmbeddedCompletionsMatchOnDisk` + the 3 zsh tests stay green
- [ ] after rebuild, `./skilldozer --completions --shell zsh` output contains both `setopt NO_LIST_AMBIGUOUS` and `setopt LIST_AMBIGUOUS`
- [ ] `completionScript("zsh")` is still byte-identical to `completions/_skilldozer`; `go.mod`/`go.sum` unchanged

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact current text of the const (with delimiter lines), the exact append text (comment + active line + opt-out), the no-guard proof (zsh `setopt` is silent non-interactively — the key shell difference from bash S1), the no-backtick constraint, the eval-derivation flow (why the const, not the autoload file), the byte-identity lock reasoning, the 4 zsh tests and why each stays green, and the scope boundary (main.go const only) are all specified with line numbers and before/after blocks. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the parallel sibling PRP (S1, bash) — the cross-shell reference
- file: plan/005_d9b30e368811/P1M1T1S1/PRP.md
  why: "S1 (bash) is the sibling: it appended a disclosure + guarded `bind` + opt-out to the ON-DISK bash file (bash is emitted verbatim). S2 is the zsh analog, but zsh is DERIVED, so the option goes in the main.go zshEvalRegistration const instead. CRITICAL DIFFERENCE: bash's `bind` needs the `[[ $- == *i* ]] &&` guard (it warns non-interactively); zsh's `setopt` does NOT (verified silent non-interactively). Do NOT cargo-cult S1's guard into zsh."
  pattern: "Both deliver (a) disclosure comment, (b) active option line, (c) commented opt-out. The disclosure prose + opt-out structure mirror S1; the active-line guard does NOT."

# MUST READ — the authoritative current text + exact old→new strings (verified against live HEAD 2dc7deb)
- file: plan/005_d9b30e368811/P1M1T1S2/research/verified_facts.md
  why: "§0 the eval-derivation flow (why the const). §1 the exact current const text (1260-1270) + the no-backtick constraint. §2 the exact edits (const append + doc broaden). §3 WHY no guard (empirical proof zsh setopt is silent non-interactively). §4 why the opt-out must be commented. §5 the 4 zsh tests and why each stays green. §6 the contract gate."
  critical: "§3 (no guard — the key zsh-vs-bash difference) and §0 (the const is appended AFTER completionScript returns, so byte-identity holds) are the two facts that prevent the most likely implementation errors (cargo-culting S1's guard; editing the autoload file)."

# MUST READ — the change map (Touch point 2 is THIS task)
- file: plan/005_d9b30e368811/architecture/code_change_map.md
  why: "Touch point 2 pins the const location (main.go:1260), the current const body, the 'why here not the autoload file' rationale, the no-backtick constraint, the doc-comment broadening guidance, the byte-identity impact, and the OPTIONAL (skip) fpath parity. Touch point 5 is the TEST scope (P1.M1.T2.S1, NOT S2)."
  section: "Touch point 2 (zsh: main.go zshEvalRegistration const) + 'Optional zsh manual parity' (skip) + Touch point 5 (tests = P1.M1.T2.S1)"

# MUST READ — the edit site (the ONLY file S2 touches)
- file: main.go
  why: "THE edit site. zshEvalRegistration const @ :1260-1270 (body 1261-1269; backtick delimiters on 1260 & 1270). zshEvalScript @ :1244-1255 (appends the const; NO logic edit). completionScript @ :1215-1226 (returns zshCompletion verbatim; UNCHANGED). runCompletion @ :1499-1522 (emits the derived wrapper for zsh only; UNCHANGED)."
  pattern: "Append the §14.7 block INSIDE the const body (after the compdef line, before the closing backtick). Mirror the existing disclosure-comment style (plain # prose, no backticks)."
  gotcha: "The const is a Go raw string literal — the ONLY backticks allowed are the two DELIMITERS (lines 1260 & 1270). The body (1261-1269) is backtick-free; the added block must stay backtick-free or compilation breaks."

- file: main_test.go
  why: "The 4 zsh tests that must stay GREEN (NO new test in S2). TestEmbeddedCompletionsMatchOnDisk @ :3139 (completionScript(zsh)==on-disk; holds — const appended after). TestZshEvalScriptStripsSelfCall @ :3266 (self-call stripped; holds). TestZshEvalScriptRegistersCompdef @ :3288 (substring checks for the 3 registration lines; holds — S2 only adds). TestRunCompletionZshIsEvalSafe @ :3316 (no self-call, has compdef, output != on-disk; holds). The byte-level NO_LIST_AMBIGUOUS assertion is P1.M1.T2.S1's scope."
  pattern: "S2's automated gate is 'existing tests stay green' + the manual CLI grep. Do NOT add the asserting test here."

- url: (PRD §14.7 + decision 22 — in PRD.md, READ-ONLY)
  why: "§14.7: list every match on the first Tab; zsh LIST_AMBIGUOUS is ON by default; setopt NO_LIST_AMBIGUOUS (+ default AUTO_LIST) lists all prefix matches immediately; SESSION-GLOBAL (no per-command scope; a scoped menu-select zstyle does NOT work — verified); MUST disclose + provide opt-out (setopt LIST_AMBIGUOUS). Decision 22 locks first-Tab-list-all. Do NOT edit PRD.md."
- url: https://zsh.sourceforge.io/Doc/Release/Options.html#index-LIST_005fAMBIGUOUS
  why: "Documents LIST_AMBIGUOUS (default ON) and NO_LIST_AMBIGUOUS. Confirms it is a global shell option (setopt), not a completion-function-scoped setting — hence it lives in the eval registration, not the _skilldozer function body."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD
2dc7deb
$ git status --short completions/ main.go          # CLEAN working tree (S1/bash not yet landed; S2 is first to edit)
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK
BUILD_OK / VET_OK
# main.go: 1522 lines. zshEvalRegistration const @ :1260-1270 (body 1261-1269; backticks on 1260/1270).
#   zshEvalScript @ :1244-1255 appends it; completionScript @ :1215-1226 returns zshCompletion verbatim;
#   runCompletion @ :1499-1522 emits the derived wrapper for zsh only.
# completions/_skilldozer: clean (NO setopt lines — S2 leaves it byte-identical, simplest path).
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# No new files. This subtask edits main.go ONLY (the const body + doc comments).
```

### Desired Codebase tree with files to be changed

```bash
main.go   # MODIFY — zshEvalRegistration const body (append §14.7 block) + its doc comment (+ optional 1-clause zshEvalScript doc tweak)
# completions/_skilldozer / completions/skilldozer.bash / completions/skilldozer.fish — UNCHANGED
# main_test.go / go.mod / go.sum — UNCHANGED (no new test in S2; no new imports)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main.go` (`zshEvalRegistration` const) | Append §14.7 disclosure + active `setopt NO_LIST_AMBIGUOUS` + commented opt-out | PRD §14.7 / decision 22 |
| `main.go` (`zshEvalRegistration` + `zshEvalScript` doc comments) | Broaden to mention the §14.7 option | contract LOGIC §3 (doc-comment broadening) |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — The option goes in the main.go CONST, NOT the autoload file. zsh is a DERIVED
// emit path: completionScript("zsh") returns completions/_skilldozer VERBATIM (locked by
// TestEmbeddedCompletionsMatchOnDisk), and runCompletion derives the eval wrapper via
// zshEvalScript which APPENDS zshEvalRegistration. The §14.7 option belongs in that append —
// leaving the autoload file and the byte-identity lock structurally untouched. (The fpath
// manual parity is OPTIONAL — skip it. Simplest path = leave completions/_skilldozer alone.)

// GOTCHA #2 — NO interactivity guard (the key zsh-vs-bash difference). bash S1 used
// `[[ $- == *i* ]] && bind ...` because bash's `bind` WARNS when sourced non-interactively.
// zsh's `setopt` is a builtin that is SILENT in any context (verified empirically: `zsh -c
// 'setopt NO_LIST_AMBIGUOUS'` prints nothing, sets the option). So `setopt NO_LIST_AMBIGUOUS`
// needs NO guard. Do NOT cargo-cult S1's `[[ $- == *i* ]] &&` into zsh.

// GOTCHA #3 — The opt-out MUST be COMMENTED. `setopt NO_LIST_AMBIGUOUS` immediately followed
// by an ACTIVE `setopt LIST_AMBIGUOUS` would CANCEL the option (last setopt wins). So the
// opt-out is a comment the user copies: `#   setopt LIST_AMBIGUOUS`. The substring
// `setopt LIST_AMBIGUOUS` is still present for the P1.M1.T2.S1 test to grep.

// GOTCHA #4 — NO BACKTICKS in the const body. zshEvalRegistration is a Go raw string literal;
// the ONLY backticks are the two DELIMITERS (main.go:1260 & 1270). The body (1261-1269) is
// backtick-free, and the added disclosure comment must stay backtick-free or the file won't
// compile. Write shell tokens plainly (e.g. write `zstyle ':completion:*:*:_skilldozer:*' menu
// select` WITHOUT surrounding backticks; write `ag<tab>` not `ag<tab>`-with-backticks).

// GOTCHA #5 — Rebuild before behavioral testing. The const is compiled into the binary. An
// already-built ./skilldozer holds the OLD const. go build (or go test, which compiles)
// re-reads it. `--completions --shell zsh` only reflects the edit after a rebuild.

// GOTCHA #6 — The byte-identity lock holds BY CONSTRUCTION. completionScript("zsh") returns
// zshCompletion (the verbatim embed) UNCHANGED; zshEvalScript appends the const AFTER
// completionScript returns. So TestEmbeddedCompletionsMatchOnDisk (completionScript(zsh) ==
// on-disk file) is untouched — you are NOT editing the embed var or the autoload file.

// GOTCHA #7 — Do NOT add the asserting test. The test that greps the emitted output for
// `setopt NO_LIST_AMBIGUOUS` + `setopt LIST_AMBIGUOUS` is P1.M1.T2.S1 (code_change_map Touch
// point 5: extend TestZshEvalScriptRegistersCompdef @ :3288 + TestRunCompletionZshIsEvalSafe
// @ :3316). S2's automated gate is the EXISTING tests staying green + the manual CLI grep.

// GOTCHA #8 — Keep the opt-out token literally `setopt LIST_AMBIGUOUS` (commented) and the
// active line literally `setopt NO_LIST_AMBIGUOUS`. The P1.M1.T2.S1 test greps for both
// substrings in the emitted output. Paraphrasing (e.g. `unsetopt LIST_AMBIGUOUS`) would fail
// that future test; the PRD §14.7 opt-out name is `setopt LIST_AMBIGUOUS` anyway.

// GOTCHA #9 — SCOPE: edit ONLY main.go (the const body + doc comments). Do NOT touch
// completionScript, completions/_skilldozer, completions/skilldozer.bash (S1),
// completions/skilldozer.fish, runCompletion logic, any test, or the README (P1.M3.T3).

// GOTCHA #10 — No deps change. The edit is inside an existing raw-string const (+ doc-comment
// prose). No new imports; go.mod/go.sum byte-for-byte identical.
```

---

## Implementation Blueprint

### Data models and structure

None. This subtask edits the body of an existing Go raw-string const + doc-comment prose. No types or signatures change; `zshEvalScript`/`completionScript`/`runCompletion` logic is untouched.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: APPEND the §14.7 block inside the zshEvalRegistration const (main.go, after the compdef line ~1269, before the closing backtick ~1270)
  - ANCHOR the edit on the const's last content line: `(( $+functions[compdef] )) && compdef _skilldozer skilldozer`
    followed by the closing backtick. (The const body is main.go:1261-1269; delimiters on 1260/1270.)
  - APPEND (exact text in Implementation Patterns):
      (a) a blank line + the disclosure comment block (§14.7 intent; zsh default LIST_AMBIGUOUS ON;
          NO_LIST_AMBIGUOUS + AUTO_LIST lists all; SESSION-GLOBAL; no per-command scope; NO guard
          needed — setopt is silent non-interactively);
      (b) the active line: `setopt NO_LIST_AMBIGUOUS`;
      (c) the commented opt-out: `#   setopt LIST_AMBIGUOUS`.
  - CONSTRAINT: NO BACKTICKS anywhere in the added block (GOTCHA #4). Verify with a grep after.
  - DO NOT modify the existing disclosure comment (1261-1266), the compinit/compdef lines (1267-1269),
    zshEvalScript's logic, or the closing backtick's position (the new block goes BEFORE it).

Task 2: BROADEN the zshEvalRegistration doc comment (main.go:1257-1259)
  - OLD says the const locks "compdef binding + the no-op-when-loaded compinit bootstrap".
  - NEW: mention it ALSO sets the §14.7 listing option (setopt NO_LIST_AMBIGUOUS + the commented
    setopt LIST_AMBIGUOUS opt-out). (Exact old→new in Implementation Patterns; contract-required.)

Task 3 (OPTIONAL, minimal — DESIGN DECISION 2): one clause in zshEvalScript's doc (main.go:1240) for accuracy
  - The doc says the append is "an explicit compdef registration plus a compinit bootstrap" — now
    incomplete. Minimal accuracy fix: "an explicit compdef registration, a compinit bootstrap, and
    the §14.7 NO_LIST_AMBIGUOUS listing option". Leave runCompletion's doc (describes the derivation
    REASON, unchanged). Skip entirely if you prefer the strictest "do not over-edit" reading.

Task 4: VERIFY — build, byte-identity, existing tests, manual CLI grep
  - COMMAND: go build ./...                                           (exit 0)
  - COMMAND: go test -run 'TestEmbeddedCompletionsMatchOnDisk|TestZshEvalScriptStripsSelfCall|TestZshEvalScriptRegistersCompdef|TestRunCompletionZshIsEvalSafe' -v ./...
                                                                      (all 4 GREEN)
  - COMMAND: go test ./...                                            (no regression)
  - MANUAL: ./skilldozer --completions --shell zsh 2>/dev/null | grep -q 'setopt NO_LIST_AMBIGUOUS'
            ./skilldozer --completions --shell zsh 2>/dev/null | grep -q 'setopt LIST_AMBIGUOUS'
  - COMMAND: grep -c '`' main.go  >/dev/null   # (sanity — the file still compiles, so raw strings are balanced; build already proved this)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"   (GOTCHA #10)
  - COMMAND: git diff --name-only                                     (MUST be ONLY main.go)
```

### Implementation Patterns & Key Details

```go
// Task 1 — the const append (exact oldText → newText). Anchor on the compdef line + closing
// backtick; insert the §14.7 block between them. NO BACKTICKS in the new block (GOTCHA #4).

//   OLD (the const's last content line + closing delimiter):
	(( $+functions[compdef] )) && compdef _skilldozer skilldozer
`

//   NEW (compdef line + the appended §14.7 block + closing delimiter):
	(( $+functions[compdef] )) && compdef _skilldozer skilldozer

	# --- §14.7 listing behavior (decision 22) ------------------------------------
	# skilldozer wants every ambiguous match listed on the FIRST Tab. A manifest-free
	# store (PRD §2) makes completion the primary discovery path, so candidates hidden
	# behind a silent common-prefix halt are a UX defect. zsh defaults to LIST_AMBIGUOUS
	# ON: the first Tab completes the common prefix and lists only once you have typed
	# to the exact ambiguous point (e.g. ag<tab> -> agent-b, nothing shown).
	#
	# setopt NO_LIST_AMBIGUOUS (with the default AUTO_LIST) makes the first Tab list ALL
	# prefix matches immediately (verified empirically: it flips ag<tab> from no-list to
	# showing both agent-browser and agent-builder). This is a SESSION-GLOBAL zsh option:
	# it changes listing for EVERY command in this shell, not just skilldozer (there is
	# no per-command scope — a scoped zstyle ':completion:*:*:_skilldozer:*' menu select
	# does NOT list on the first Tab; only the global NO_LIST_AMBIGUOUS does).
	#
	# Unlike bash's bind (which warns when sourced non-interactively), zsh setopt is a
	# builtin that is silent in any context, so this line needs NO interactivity guard.
	#
	# Opt-out — restore zsh's stock (exact-ambiguous-point) listing:
	#   setopt LIST_AMBIGUOUS
	setopt NO_LIST_AMBIGUOUS
`

// Task 2 — the zshEvalRegistration doc comment (exact oldText → newText).

//   OLD:
// zshEvalRegistration is appended to the stripped autoload body to make it eval-safe. Its own
// const so a test can lock the exact registration contract (compdef binding + the no-op-when-
// loaded compinit bootstrap). No backticks inside: it is a Go raw string literal.

//   NEW:
// zshEvalRegistration is appended to the stripped autoload body to make it eval-safe. Its own
// const so a test can lock the exact registration contract: the compdef binding, the no-op-when-
// loaded compinit bootstrap, AND the §14.7 listing option (setopt NO_LIST_AMBIGUOUS + the
// commented setopt LIST_AMBIGUOUS opt-out). No backticks inside: it is a Go raw string literal.
```

Notes:
- The disclosure comment's prose is flexible; the hard requirements are (1) it names
  `NO_LIST_AMBIGUOUS`, (2) notes session-global, (3) notes the zsh default (LIST_AMBIGUOUS ON),
  (4) gives the opt-out. The exact block above satisfies all four and is the recommended form.
- The active line is UNGUARDED (GOTCHA #2 — zsh `setopt` is silent non-interactively; verified).
- The opt-out is COMMENTED (`#   setopt LIST_AMBIGUOUS`) so it does not cancel the active line
  (GOTCHA #3), while still providing the substring the P1.M1.T2.S1 test greps for (GOTCHA #8).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **No interactivity guard (unlike bash S1).** bash's `bind 'set show-all-if-ambiguous on'` warns when sourced non-interactively, so S1 guarded it with `[[ $- == *i* ]] &&`. zsh's `setopt` is a builtin that is silent in any context (empirically verified: `zsh -c 'setopt NO_LIST_AMBIGUOUS'` prints nothing and sets the option). So `setopt NO_LIST_AMBIGUOUS` is unguarded. Cargo-culting S1's guard would be unnecessary and (zsh's `$-`/`[[ ]]` interactivity detection being less clean than bash's) slightly awkward. This is the single most important shell-specific decision.

2. **zshEvalScript doc: minimal one-clause accuracy tweak (optional).** The contract says broaden zshEvalScript/runCompletion docs ONLY if they claim the registration is "solely compdef" — "do not over-edit." zshEvalScript's doc (main.go:1240) says the append is "an explicit compdef registration plus a compinit bootstrap" — now incomplete. The PRP recommends a single-clause fix ("…, and the §14.7 NO_LIST_AMBIGUOUS listing option") to keep the doc truthful, and leaves runCompletion's doc (which describes the derivation REASON, unchanged) alone. If you prefer the strictest "do not over-edit" reading, drop Task 3 entirely — the zshEvalRegistration doc (Task 2, contract-required) is the authoritative description of the const's contents.

3. **Leave completions/_skilldozer alone (simplest path).** The hard requirement is the EVAL path (the const). Full fpath/manual parity (adding `setopt NO_LIST_AMBIGUOUS` to the autoload function body) is OPTIONAL per the code_change_map and would require keeping `completionScript("zsh")` byte-identical to the edited file. Skipping it keeps the byte-identity lock trivially intact and the diff to one file. (A future task can add fpath parity if desired.)

### Integration Points

```yaml
DERIVATION FLOW (NO logic edit — the const is data zshEvalScript appends):
  - completionScript("zsh") (main.go:1215) → zshCompletion verbatim (UNCHANGED).
  - zshEvalScript (main.go:1244-1255) → strips self-call, appends zshEvalRegistration (the edited const).
  - runCompletion (main.go:1499-1522, esp. :1518) → emits the derived wrapper for zsh only.
  - So `eval "$(skilldozer --completions)"` (zsh) now runs the const's setopt at eval time.

BYTE-IDENTITY LOCK (holds by construction — GOTCHA #6):
  - completionScript("zsh") returns the embed var UNCHANGED; the const is appended AFTER it returns.
  - TestEmbeddedCompletionsMatchOnDisk (main_test.go:3139) compares the embed var to the on-disk
    autoload file — both UNCHANGED → stays GREEN.

TESTS (unchanged; they gate the change — NO new test in S2):
  - TestZshEvalScriptStripsSelfCall (3266), TestZshEvalScriptRegistersCompdef (3288),
    TestRunCompletionZshIsEvalSafe (3316): substring/inequality checks; S2 only ADDS content → GREEN.
  - The byte-level NO_LIST_AMBIGUOUS assertion is P1.M1.T2.S1 (code_change_map Touch point 5).

NO DATABASE / NO CONFIG / NO ROUTES / NO COMPLETION FILE / NO NEW DEP:
  - This subtask edits one Go raw-string const (+ doc prose). No autoload file, no embed var, no
    parseArgs, no run(), no usageText, no new import.
```

---

## Validation Loop

### Level 1: Build + edit presence (immediate, after the const edit)

```bash
cd /home/dustin/projects/skilldozer

go build ./... && echo "build OK" || echo "FAIL: build (likely a stray backtick in the raw string — GOTCHA #4)"
# Expected: "build OK". (A raw-string break = compile error; build is the first gate.)

# Confirm the three elements are present in the const (no backticks leaked into the body):
grep -q -- 'setopt NO_LIST_AMBIGUOUS' main.go && echo "active setopt present" || echo "FAIL"
grep -q -- 'setopt LIST_AMBIGUOUS' main.go     && echo "opt-out token present" || echo "FAIL"
# The const body has NO backticks beyond the two delimiters (build already proved the raw string is balanced):
awk '/const zshEvalRegistration = `/,/^`$/' main.go | grep -c '`'   # Expected: 2 (the delimiters only)
# Expected: build OK; both greps succeed; the awk count is exactly 2 (open + close delimiters).
```

### Level 2: The byte-identity gate + existing zsh tests (automated, post-rebuild)

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"     # Expected: 0
go test -run 'TestEmbeddedCompletionsMatchOnDisk|TestZshEvalScriptStripsSelfCall|TestZshEvalScriptRegistersCompdef|TestRunCompletionZshIsEvalSafe' -v ./...
# Expected: all 4 PASS.
#   - TestEmbeddedCompletionsMatchOnDisk: completionScript(zsh) == on-disk (UNCHANGED) → PASS.
#   - TestZshEvalScriptStripsSelfCall: self-call still stripped → PASS.
#   - TestZshEvalScriptRegistersCompdef: the 3 registration substrings still present → PASS.
#   - TestRunCompletionZshIsEvalSafe: no self-call, has compdef, output != on-disk → PASS.

go test ./... ; echo "test exit $?"       # Expected: 0 (no regression)
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
git diff --name-only                      # Expected: ONLY main.go
# Expected: build/test exit 0; deps unchanged; only main.go changed.
```

### Level 3: The behavioral proof (the actual §14.7 contract — manual CLI grep)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# The emitted zsh eval-path output carries the active option + the opt-out token:
/tmp/sdz --completions --shell zsh 2>/dev/null | grep -q 'setopt NO_LIST_AMBIGUOUS' && echo "emit active OK"  || echo "FAIL"
/tmp/sdz --completions --shell zsh 2>/dev/null | grep -q 'setopt LIST_AMBIGUOUS'    && echo "emit opt-out OK" || echo "FAIL"

# The autoload file is UNCHANGED (byte-identity lock holds — the option is ONLY in the derived wrapper):
grep -q 'NO_LIST_AMBIGUOUS' completions/_skilldozer && echo "FAIL: autoload file changed" || echo "autoload file unchanged OK"

# completionScript('zsh') (the embed) is still byte-identical to the on-disk autoload file:
diff <(/tmp/sdz --completions --shell zsh 2>/dev/null) completions/_skilldozer >/dev/null && echo "FAIL: emit == autoload (derivation lost)" || echo "derived != autoload OK"

# Control: the derivation's core is intact (compdef registration present, self-call stripped):
/tmp/sdz --completions --shell zsh 2>/dev/null | grep -q 'compdef _skilldozer skilldozer' && echo "compdef intact OK" || echo "FAIL"
/tmp/sdz --completions --shell zsh 2>/dev/null | grep -q '_skilldozer "$@"' && echo "FAIL: self-call present" || echo "self-call stripped OK"

rm -f /tmp/sdz
# Expected: every line "...OK"; the autoload file grep finds nothing; emit != autoload (derivation holds).
```

### Level 4: Scope-discipline check (only main.go changed)

```bash
cd /home/dustin/projects/skilldozer

git diff --name-only
# Expected: ONLY main.go. (If completions/_skilldozer, completions/skilldozer.bash, completions/skilldozer.fish,
# README.md, main_test.go, or any other file appears, you over-reached into S1 / P1.M1.T2 / P1.M1.T3 / autoload-parity scope.)

# zsh/fish/bash embed-match still holds (autoload files unchanged):
go test -run TestEmbeddedCompletionsMatchOnDisk -v ./... 2>&1 | grep -E 'bash|zsh|fish|PASS|FAIL'
# Expected: PASS (all three shells; the autoload files are untouched).
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `go build` exit 0 (raw string balanced); greps confirm active `setopt NO_LIST_AMBIGUOUS` + opt-out `setopt LIST_AMBIGUOUS`; awk backtick count in the const is exactly 2
- [ ] Level 2 PASS — `go build` exit 0; the 4 zsh tests (`TestEmbeddedCompletionsMatchOnDisk`, `TestZshEvalScriptStripsSelfCall`, `TestZshEvalScriptRegistersCompdef`, `TestRunCompletionZshIsEvalSafe`) all PASS; `go test ./...` exit 0; deps unchanged; only `main.go` changed
- [ ] Level 3 PASS — emitted zsh output contains `setopt NO_LIST_AMBIGUOUS` and `setopt LIST_AMBIGUOUS`; autoload file unchanged; emit != autoload (derivation holds); compdef intact; self-call stripped
- [ ] Level 4 PASS — only `main.go` changed; embed-match passes for all three shells

### Feature Validation
- [ ] the `zshEvalRegistration` const body contains the active line `setopt NO_LIST_AMBIGUOUS` (unguarded) after the compdef block
- [ ] the const body contains the commented opt-out `#   setopt LIST_AMBIGUOUS`
- [ ] the disclosure comment names `NO_LIST_AMBIGUOUS`, notes session-global, notes the zsh default (LIST_AMBIGUOUS ON), gives the opt-out
- [ ] the const body has NO backticks beyond the two delimiters
- [ ] the `zshEvalRegistration` doc comment mentions the §14.7 listing option

### Code Quality / Convention Validation
- [ ] the active line is UNGUARDED (zsh `setopt` is silent non-interactively — verified; NOT cargo-culted from bash S1)
- [ ] the opt-out is COMMENTED (does not cancel the active line) and uses the exact token `setopt LIST_AMBIGUOUS`
- [ ] the disclosure comment uses plain `#` prose with NO backticks (raw-string-safe), mirroring the existing const comment style
- [ ] no autoload file, no embed var, no `completionScript`/`zshEvalScript`/`runCompletion` logic edited; `go.mod`/`go.sum` byte-for-byte identical

### Scope Discipline
- [ ] Did NOT touch `completions/_skilldozer` (autoload file — simplest path leaves it byte-identical), `completions/skilldozer.bash` (S1), or `completions/skilldozer.fish`
- [ ] Did NOT touch `completionScript`, `zshEvalScript` logic, or `runCompletion` logic (only the const data + doc comments)
- [ ] Did NOT add a Go test (the byte-level assertion is P1.M1.T2.S1)
- [ ] Did NOT edit the README (the §14.7 disclosure is P1.M1.T3.S1, Mode B)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't edit the autoload file (`completions/_skilldozer`).** zsh is a DERIVED path; the option belongs in the `zshEvalRegistration` const (eval-time append). Editing the autoload file is the OPTIONAL fpath-parity path and risks the byte-identity lock for no eval-path benefit. Simplest path = leave it alone.
- ❌ **Don't cargo-cult bash S1's `[[ $- == *i* ]] &&` guard.** bash's `bind` warns non-interactively; zsh's `setopt` does not (verified silent). The zsh active line is UNGUARDED.
- ❌ **Don't make the opt-out an ACTIVE line.** `setopt NO_LIST_AMBIGUOUS` followed by active `setopt LIST_AMBIGUOUS` cancels the option. The opt-out is a COMMENT (`#   setopt LIST_AMBIGUOUS`) the user copies.
- ❌ **Don't put backticks in the const body.** `zshEvalRegistration` is a Go raw string literal; only the two delimiter backticks (main.go:1260 & 1270) are allowed. Shell tokens in the disclosure (e.g. the `zstyle …` example) must be plain text, not backticked.
- ❌ **Don't paraphrase the option tokens.** Use the exact strings `setopt NO_LIST_AMBIGUOUS` (active) and `setopt LIST_AMBIGUOUS` (opt-out). The P1.M1.T2.S1 test greps for both; PRD §14.7 names them.
- ❌ **Don't add the asserting Go test here.** The test that locks `setopt NO_LIST_AMBIGUOUS` in the emitted output is P1.M1.T2.S1 (code_change_map Touch point 5). S2's gate is existing tests staying green + the manual grep.
- ❌ **Don't skip the rebuild.** The const is compiled in; an already-built `./skilldozer` holds the OLD bytes. `go build` (or `go test`) re-reads it.
- ❋ **Don't fear the byte-identity lock.** `completionScript("zsh")` returns the embed var UNCHANGED; the const is appended AFTER it returns (in `zshEvalScript`). `TestEmbeddedCompletionsMatchOnDisk` is untouched by construction.
- ❌ **Don't over-edit docs.** Broaden `zshEvalRegistration`'s doc (required). For `zshEvalScript`/`runCompletion`, the contract says broaden ONLY if they claim "solely compdef" — the PRP recommends at most a one-clause accuracy tweak to `zshEvalScript` and leaving `runCompletion` alone.
- ❌ **Don't touch bash/fish.** bash is S1 (a different file); fish lists by default (no option). zsh is the only derived shell needing this const edit.
- ❌ **Don't add deps.** The edit is inside an existing raw-string const (+ doc prose). No new imports.

---

## Confidence Score

**9.5/10** — Every edit is pinned to the exact current (HEAD `2dc7deb`) const text with before/after blocks; the eval-derivation flow is confirmed in source (the const is appended after `completionScript` returns, so the byte-identity lock holds by construction); the two subtleties that matter most are empirically proven in `research/verified_facts.md` — (§3) zsh `setopt` is silent non-interactively so NO guard is needed (the key difference from bash S1), and (§5) all 4 zsh tests are substring/inequality checks that stay green because S2 only ADDS content. The exact tokens the next subtask's test will grep for (`setopt NO_LIST_AMBIGUOUS` / `setopt LIST_AMBIGUOUS`) are specified verbatim, and the no-backtick raw-string constraint is called out with a build-level gate. The 0.5 reservation is the optional `zshEvalScript` doc-clause tweak (DESIGN DECISION 2) — the contract's "do not over-edit" guard leaves it to judgment; the PRP recommends the minimal accuracy fix but a reviewer may prefer to drop it, which is a one-clause difference with no behavioral impact.
