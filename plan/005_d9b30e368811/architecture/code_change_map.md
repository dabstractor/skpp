# Code Change Map — Skilldozer session 005 (§14.7 listing behavior)

Exact touch points verified against the current tree. Three surfaces + tests +
one Mode B doc. Nothing else changes.

## Touch point 1 — bash: `completions/skilldozer.bash` (on-disk == emitted)

- **File:** `completions/skilldozer.bash` (currently 69 lines).
- **Last line (verbatim):** `complete -F _skilldozer_completion skilldozer`
- **Why one edit covers both paths:** the file is `//go:embed`-ded
  (`var bashCompletion`, main.go:55) and `runCompletion` emits bash **verbatim**
  (main.go:1520). So the §14.5 manual `source` path and the §14.6 `eval` path
  produce **identical bytes**.
- **Change:** AFTER the `complete -F …` line, append:
  - a disclosure comment block naming `show-all-if-ambiguous`, noting it is
    session-global, and giving the opt-out `bind 'set show-all-if-ambiguous off'`;
  - an **active** line guarded for interactivity:
    `[[ $- == *i* ]] && bind 'set show-all-if-ambiguous on'`
    (the guard silences a `bind` warning if sourced non-interactively).
- **Byte-identity impact:** `TestEmbeddedCompletionsMatchOnDisk`
  (main_test.go:3139) compares the embed var to the file on disk — they move
  together (same file), so it **stays green**. A `go build` is required for
  `skilldozer --completions` to reflect the on-disk edit (§14.6 lockstep).

## Touch point 2 — zsh: `main.go` `zshEvalRegistration` const (eval-path only)

- **Location:** `zshEvalRegistration` raw-string const, main.go:1260.
- **Current const body:**
  ```
  autoload -Uz compinit
  (( $+functions[compdef] )) || compinit
  (( $+functions[compdef] )) && compdef _skilldozer skilldozer
  ```
  (plus the disclosure comment lines already in the const.)
- **Why here, not the autoload file:** `completionScript("zsh")` returns the
  on-disk `completions/_skilldozer` **verbatim** and `TestEmbeddedCompletions
  MatchOnDisk` locks that. `runCompletion` derives the eval output via
  `zshEvalScript` which appends `zshEvalRegistration`. So the §14.7 option
  belongs in the **registration append** (eval-time), leaving the autoload file
  and the byte-identity lock untouched.
- **Change:** inside the const, add (before or after the compdef block):
  - a disclosure comment naming `NO_LIST_AMBIGUOUS`, noting it is
    session-global, and giving the opt-out `setopt LIST_AMBIGUOUS`;
  - an **active** line: `setopt NO_LIST_AMBIGUOUS`.
  - Keep the const a valid Go raw string literal — **no backticks** inside
    (the existing const already observes this).
  - Broaden the `zshEvalRegistration` / `zshEvalScript` doc comments
    (main.go:1257, ~1244) if they claim the registration is *only* about
    compdef; do not over-edit.
- **Byte-identity impact:** `completionScript("zsh")` is unchanged (still
  byte-identical to `completions/_skilldozer`); the const is appended
  **after** `completionScript` returns. The byte-identity lock holds.

### Optional zsh manual (fpath) parity

The hard requirement is the **eval path** (`zshEvalRegistration`). Full manual
(`fpath`) parity is OPTIONAL. If pursued: add `setopt NO_LIST_AMBIGUOUS` as the
first line of the `_skilldozer()` function body in `completions/_skilldozer`.
**If** you touch that file, `completionScript("zsh")` must remain byte-identical
to it — the autoload-file edit must not depend on the derivation. **Simplest: leave
`completions/_skilldozer` alone.**

## Touch point 3 — fish: NO code change required

- **File:** `completions/skilldozer.fish`.
- fish lists all matches in the pager **by default** (§14.7). No option to set.
- **Optional:** a one-line clarifying comment that fish lists by default so the
  §14.7 contract is already satisfied. Not required.

## Touch point 4 — README disclosure (Mode B, changeset-level docs)

- **File:** `README.md`, *Shell completions* section (lines 290-366).
- **Insertion point:** after the existing skills-first/long-form-only bullet list
  (ends ~line 334) and before "Prefer to copy the file instead?" (~line 336).
- **Change:** a disclosure paragraph + opt-out block stating:
  - the emitted script sets a **session-global** option so ambiguous matches list
    on the FIRST Tab instead of halting at the common prefix;
  - name the option per shell: bash `show-all-if-ambiguous`, zsh
    `NO_LIST_AMBIGUOUS` (fish lists by default — no option set);
  - note it affects listing for **every command** in that shell, not just
    skilldozer;
  - give the opt-out one-liners:
    `bind 'set show-all-if-ambiguous off'` (bash);
    `setopt LIST_AMBIGUOUS` (zsh).
- **Verify:** `grep -q 'show-all-if-ambiguous' README.md && grep -q 'NO_LIST_AMBIGUOUS' README.md && grep -q 'LIST_AMBIGUOUS' README.md`

## Touch point 5 — Tests (byte-level emitted-script assertions)

- **File:** `main_test.go`.
- **bash:** add/extend a test (e.g. `TestRunCompletionBashListsAmbiguous` or
  extend `TestRunCompletionBashScript` at main_test.go:3163): `run(["--completions","--shell","bash"], …)`
  exit 0 → assert `out` contains `show-all-if-ambiguous on` AND the opt-out token
  `show-all-if-ambiguous off`; optionally assert the `*i*` guard is present.
- **zsh:** extend `TestZshEvalScriptRegistersCompdef` (main_test.go:3288): assert
  `zshEvalScript(completionScript("zsh"))` contains `setopt NO_LIST_AMBIGUOUS`
  (active) AND `setopt LIST_AMBIGUOUS` (opt-out token). Extend
  `TestRunCompletionZshIsEvalSafe` (main_test.go:3316) to assert the end-to-end
  emitted script (via `run(["--completions"], …)` under `SKILLDOZER_SHELL=zsh`)
  contains `NO_LIST_AMBIGUOUS`.
- **Invariant:** `TestEmbeddedCompletionsMatchOnDisk` (main_test.go:3139) MUST
  remain green (do not weaken it).

## Build verification (must all pass after the change)

```bash
go build ./...                                  # clean
go test ./...                                   # 100% green, no regression
grep -q 'show-all-if-ambiguous' completions/skilldozer.bash
grep -q 'show-all-if-ambiguous off' completions/skilldozer.bash   # opt-out token disclosed
./skilldozer --completions --shell bash 2>/dev/null | grep -q 'show-all-if-ambiguous on'
./skilldozer --completions --shell zsh  2>/dev/null | grep -q 'NO_LIST_AMBIGUOUS'
./skilldozer --completions --shell zsh  2>/dev/null | grep -q 'setopt LIST_AMBIGUOUS'   # opt-out
grep -q 'show-all-if-ambiguous' README.md && grep -q 'NO_LIST_AMBIGUOUS' README.md && grep -q 'LIST_AMBIGUOUS' README.md
```
