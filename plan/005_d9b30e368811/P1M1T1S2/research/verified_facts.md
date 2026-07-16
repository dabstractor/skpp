# Verified Facts — P1.M1.T1.S2 (zsh: add NO_LIST_AMBIGUOUS to the derived eval registration)

Confirmed against the **live** tree at HEAD `2dc7deb` (main.go = 1522 lines; line numbers in
the contract + code_change_map are ACCURATE this session). The working tree is CLEAN — the
parallel sibling S1 (bash, `completions/skilldozer.bash`) has NOT landed yet, but it edits a
DIFFERENT file, so there is no conflict with S2's `main.go` const edit.

---

## 0. The eval-derivation flow (why the option goes in a main.go const, NOT the autoload file)

```
completionScript("zsh")      → returns zshCompletion VERBATIM (== on-disk completions/_skilldozer)
                               [main.go:1215-1226; UNCHANGED by S2]
        │
        ▼  (only for the eval path, in runCompletion)
zshEvalScript(raw)           → strips trailing `_skilldozer "$@"` self-call, APPENDS zshEvalRegistration
                               [main.go:1244-1255]
        │
        ▼
runCompletion emits          → bash/fish: verbatim;  zsh: the DERIVED wrapper  [main.go:1499-1522, esp. 1518]
```

**The §14.7 listing option belongs in `zshEvalRegistration` (eval-time append), NOT in the
autoload file.** Three reasons (all from the contract + code_change_map Touch point 2):

1. `completionScript("zsh")` returns the on-disk `completions/_skilldozer` **verbatim**, and
   `TestEmbeddedCompletionsMatchOnDisk` locks that byte-identity. Editing the autoload file
   would move WITH the embed (test stays green), but the contract's SIMPLEST path is to leave
   the autoload file alone so the eval-path fix does NOT depend on the autoload file changing.
2. The autoload (fpath) path is a SEPARATE delivery path (§14.5: copy onto fpath + compinit).
   Manual parity there is OPTIONAL (code_change_map "Optional zsh manual parity") — skip it.
3. The const is appended **AFTER** `completionScript` returns, so the byte-identity lock is
   structurally untouched: `completionScript("zsh")` still == on-disk file; only the DERIVED
   wrapper (what `--completions` emits for zsh) gains the option.

---

## 1. The CURRENT (HEAD `2dc7deb`) `zshEvalRegistration` const — exact text (main.go:1257-1270)

```go
// zshEvalRegistration is appended to the stripped autoload body to make it eval-safe. Its own
// const so a test can lock the exact registration contract (compdef binding + the no-op-when-
// loaded compinit bootstrap). No backticks inside: it is a Go raw string literal.
const zshEvalRegistration = `
# Register the completion for eval. The #compdef header above only binds this as an
# autoload file on fpath; under eval it is inert, so bind the function explicitly.
# compsys (_arguments/_files/compadd) is bootstrapped only if not already loaded —
# oh-my-zsh / prezto / a manual compinit all define compdef, making the compinit
# below a no-op. The autoload file's trailing self-call is intentionally omitted:
# it would fire the function at eval time, before _arguments is guaranteed to exist.
autoload -Uz compinit
(( $+functions[compdef] )) || compinit
(( $+functions[compdef] )) && compdef _skilldozer skilldozer
`
```

- Line 1260: `const zshEvalRegistration = \``  ← opening raw-string backtick
- Lines 1261-1266: the disclosure comment (eval-safety rationale)
- Line 1267: `autoload -Uz compinit`
- Line 1268: `(( $+functions[compdef] )) || compinit`
- Line 1269: `(( $+functions[compdef] )) && compdef _skilldozer skilldozer`
- Line 1270: `` ` ``  ← closing raw-string backtick

**The const body (1261-1269) contains NO backticks** (verified: the only backticks on lines
1260/1270 are the Go raw-string DELIMITERS). S2's added disclosure comment + active line must
ALSO contain no backticks — else the raw string literal breaks compilation. (Constraint stated
in the contract LOGIC §3 and code_change_map Touch point 2.)

---

## 2. The exact edit (old → NEW)

### Edit A — append the §14.7 block inside the const body (after the compdef line, before the closing backtick)

OLD (the const's last content line + closing delimiter):
```
(( $+functions[compdef] )) && compdef _skilldozer skilldozer
`
```
NEW:
```
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
```

**Hard requirements satisfied** (contract LOGIC 3a/3b + OUTPUT 4):
- active line `setopt NO_LIST_AMBIGUOUS` (present verbatim) ✓
- opt-out token `setopt LIST_AMBIGUOUS` (commented, so it does NOT cancel the active line) ✓
- disclosure names `NO_LIST_AMBIGUOUS`, notes SESSION-GLOBAL, gives the opt-out ✓
- **NO backticks** anywhere in the added block (raw-string-safe) ✓

### Edit B — broaden the `zshEvalRegistration` doc comment (main.go:1257-1259; contract-required)

OLD:
```go
// zshEvalRegistration is appended to the stripped autoload body to make it eval-safe. Its own
// const so a test can lock the exact registration contract (compdef binding + the no-op-when-
// loaded compinit bootstrap). No backticks inside: it is a Go raw string literal.
```
NEW:
```go
// zshEvalRegistration is appended to the stripped autoload body to make it eval-safe. Its own
// const so a test can lock the exact registration contract: the compdef binding, the no-op-when-
// loaded compinit bootstrap, AND the §14.7 listing option (setopt NO_LIST_AMBIGUOUS + the
// commented setopt LIST_AMBIGUOUS opt-out). No backticks inside: it is a Go raw string literal.
```

### Edit C (OPTIONAL, minimal — DESIGN DECISION) — one clause in `zshEvalScript`'s doc for accuracy
The contract says broaden zshEvalScript/runCompletion docs ONLY if they claim the registration
is "solely compdef" — "do not over-edit." zshEvalScript's doc (main.go:1240) currently says the
append is "an explicit compdef registration plus a compinit bootstrap" — which becomes incomplete
after Edit A. The minimal accuracy fix: "an explicit compdef registration, a compinit bootstrap,
and the §14.7 NO_LIST_AMBIGUOUS listing option". runCompletion's doc describes the derivation
REASON (unchanged) → leave it. (See PRP DESIGN DECISION 2.)

---

## 3. Why NO interactivity guard (the key zsh-vs-bash difference)

The bash sibling S1 used `[[ $- == *i* ]] && bind 'set show-all-if-ambiguous on'` because
bash's `bind` PRINTS A WARNING when sourced non-interactively (eval test harnesses). **zsh's
`setopt` does NOT** — it is a builtin, silent in any context. Verified empirically:

```
$ zsh -c 'setopt NO_LIST_AMBIGUOUS; echo "silent OK"; setopt | grep -i listambiguous'
silent OK
nolistambiguous           ← option set, NO warning printed
```

So `setopt NO_LIST_AMBIGUOUS` needs **NO guard**. Adding one (e.g. `[[ $- == *i* ]] &&`) would
be cargo-culted from bash and is unnecessary — and in zsh `$-`/`[[ ]]` interactivity detection
is less clean than bash's. Leave the line unguarded. (This is the single most important
shell-specific fact: do NOT mirror S1's guard.)

---

## 4. The opt-out must be COMMENTED (not active)

`setopt NO_LIST_AMBIGUOUS` immediately followed by an ACTIVE `setopt LIST_AMBIGUOUS` would
CANCEL the option (the last setopt wins). So the opt-out MUST be a comment the user copies
manually: `#   setopt LIST_AMBIGUOUS`. The substring `setopt LIST_AMBIGUOUS` is still present
for the P1.M1.T2.S1 test to grep (the test greps the emitted output for the token, comment or
not). Mirrors bash S1's commented opt-out (`#   bind 'set show-all-if-ambiguous off'`).

---

## 5. Test impact — S2 breaks NOTHING; all 4 zsh tests stay GREEN

| Test | What it asserts | S2 effect |
|---|---|---|
| `TestEmbeddedCompletionsMatchOnDisk` (3139) | `completionScript("zsh")` == on-disk `completions/_skilldozer` | PASSES — the const is appended AFTER completionScript returns; the autoload file + embed var are UNCHANGED |
| `TestZshEvalScriptStripsSelfCall` (3266) | `zshEvalScript(raw)` has NO self-call; raw embed DOES end with self-call | PASSES — S2 doesn't touch stripping or the autoload file |
| `TestZshEvalScriptRegistersCompdef` (3288) | `zshEvalScript(raw)` CONTAINS the 3 registration lines + header + `_arguments -C` | PASSES — S2 only ADDS to the const; all asserted substrings remain present |
| `TestRunCompletionZshIsEvalSafe` (3316) | `run(["--completions"], zsh)` exit 0, no self-call, has compdef, output != on-disk file | PASSES — all four hold (the derived wrapper is still longer than the autoload file; adding content keeps them unequal) |

No test asserts the const's EXACT content or that it ENDS with the compdef line — all are
substring / inequality checks. S2 ADDS content, so every assertion still holds. **S2 does NOT
add a test** (the byte-level assertion is P1.M1.T2.S1's scope; the contract OUTPUT §4 gate is
build + the CLI grep).

## 6. S2's contract gate (OUTPUT §4)
- `go build ./...` succeeds.
- `completionScript("zsh")` is still byte-identical to `completions/_skilldozer` (TestEmbeddedCompletionsMatchOnDisk green) — the const is appended AFTER completionScript returns.
- After rebuild, `./skilldozer --completions --shell zsh` output contains `setopt NO_LIST_AMBIGUOUS` (active) AND `setopt LIST_AMBIGUOUS` (opt-out token).
- go.mod/go.sum unchanged (no new imports; `setopt`/comments are inside an existing raw-string const).

## 7. Scope discipline
S2 edits ONLY: the `zshEvalRegistration` const body (Edit A) + its doc comment (Edit B) + an
optional one-clause zshEvalScript doc tweak (Edit C). It does NOT touch: `completionScript`,
`completions/_skilldozer`, `completions/skilldozer.bash` (S1), `completions/skilldozer.fish`,
`runCompletion` logic, any test, the README (P1.M3.T3), or PRD.md.
