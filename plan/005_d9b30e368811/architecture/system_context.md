# System Context — Skilldozer session 005 (§14.7 listing behavior)

## What this session is

A **delta** from session 004 (`plan/004_5851dcff4371/`), whose work (subcommands →
flags, skills-first completions, symlink discovery, derived-zsh emit) is
**Complete and green**. This session implements **one new PRD section: §14.7**
*"Listing behavior — show every match, never a silent halt"*, plus its
decisions-log mirror (Decision 22) and a one-sentence cross-reference in §14.1
rule 2.

**Scope is deliberately small and confined to the completion-emit layer.**
Nothing in `parseArgs`, `discover`, `resolve`, `skillsdir`, `ui`, `config`,
`runInit`, `runCheck`, or any tag-resolution/store code is touched. No flags are
added/removed/renamed — the §14.4 lockstep and `usageText` are NOT triggered.

## Verified current state (build + test baseline)

- `go build ./...` → **exit 0** (clean).
- `go test ./...` → **all packages green** (`ok` for main + all 8 internal pkgs).
- `grep` for `show-all-if-ambiguous`, `NO_LIST_AMBIGUOUS`, `LIST_AMBIGUOUS` across
  `README.md`, `completions/*`, `main.go`, `main_test.go` → **zero matches**.
  The §14.7 option-setting does not exist yet anywhere in the tree.

## The completion-emit machinery this delta builds on (verified line numbers)

All line numbers verified against the current `main.go` (73 KB) /
`main_test.go` (139 KB) as of this session.

| Symbol / file | Location | Role in this delta |
|---|---|---|
| `//go:embed completions/skilldozer.bash` / `var bashCompletion` | main.go:54-55 | bash bytes; **verbatim** for both manual-`source` and `eval` paths |
| `//go:embed completions/_skilldozer` / `var zshCompletion` | main.go:57-58 | zsh autoload bytes; **verbatim** for the fpath manual path; **derived** for eval |
| `//go:embed completions/skilldozer.fish` / `var fishCompletion` | main.go:60-61 | fish bytes; verbatim; **NO change needed** (lists by default) |
| `completionScript(shell)` | main.go:1215 | returns embedded var **verbatim** for the given shell; `(script, ok)` |
| `zshEvalScript(raw)` | main.go:1244 | strips trailing `_skilldozer "$@"` self-call; appends `zshEvalRegistration` |
| `zshEvalRegistration` const | main.go:1260 | the **eval-time append** — this is where the zsh §14.7 fix lands |
| `runCompletion(c, stdout, stderr)` | main.go:1499 | bash/fish → emit verbatim; zsh → emit `zshEvalScript(...)` derived wrapper |
| `completions/skilldozer.bash` | file (69 lines) | ends with `complete -F _skilldozer_completion skilldozer`; this is where the bash §14.7 fix lands (covers both paths because bytes are identical) |
| `completions/_skilldozer` | file (autoload) | the fpath manual path; **leave alone** for the required fix (eval-path optional note below) |
| `completions/skilldozer.fish` | file | lists by default → optional one-line comment only |
| `README.md` *Shell completions* section | lines 290-366 | where the Mode B disclosure + opt-out lands |

### Key invariant: bash on-disk == bash emitted (verbatim)

`completionScript("bash")` returns the `//go:embed` var, which points at
`completions/skilldozer.bash`. `runCompletion` emits bash **verbatim**. Therefore
**one edit to `completions/skilldozer.bash` covers both delivery paths**
(§14.5 manual `source` AND §14.6 `eval`). This is the simplest touch point.

### Key invariant: zsh is derived (eval-path only)

`runCompletion` emits zsh via `zshEvalScript(completionScript("zsh"))`, which
appends `zshEvalRegistration`. `completionScript("zsh")` itself stays
byte-identical to `completions/_skilldozer` (the §14.6 byte-identity lock).
Therefore the zsh §14.7 fix lands in **`zshEvalRegistration`** (main.go:1260),
NOT in `completionScript` or the on-disk autoload file. The fpath manual path is
**optionally** addressed (see delta_prd §5 Note) — simplest is to leave
`completions/_skilldozer` alone.

## Existing test patterns to mirror (verified)

| Test | Line | Pattern |
|---|---|---|
| `TestEmbeddedCompletionsMatchOnDisk` | main_test.go:3139 | byte-identity lock (on-disk == `completionScript` return); MUST stay green |
| `TestRunCompletionBashScript` | main_test.go:3163 | `run(["--completions","--shell","bash"], …)` → assert `out` contains marker + stderr empty |
| `TestRunCompletionFishScript` | main_test.go:3179 | same shape for fish |
| `TestZshEvalScriptStripsSelfCall` | main_test.go:3266 | `zshEvalScript(raw)` must not contain self-call |
| `TestZshEvalScriptRegistersCompdef` | main_test.go:3288 | `zshEvalScript(raw)` must contain compdef lines — **extend this** for NO_LIST_AMBIGUOUS |
| `TestRunCompletionZshIsEvalSafe` | main_test.go:3316 | end-to-end `run(["--completions"], …)` under `SKILLDOZER_SHELL=zsh` → assert compdef present + self-call absent — **extend this** for NO_LIST_AMBIGUOUS |

The §14.7 test strategy is **byte-level assertion on emitted scripts** (does the
output contain the active option + the disclosed opt-out token?), mirroring the
existing `TestZshEvalScript*` / `TestRunCompletion*Script` pattern. There is NO
in-process way to drive a live shell's first-Tab behavior from the binary, so
§13 acceptance gains no new in-process assertions.

## Out of scope (hard guardrails)

- ❌ No flag add/remove/rename (§14.4 lockstep + `usageText` not triggered).
- ❌ Do not touch `parseArgs`, `discover`, `resolve`, `skillsdir`, `ui`, `config`, `runInit`, `runCheck`.
- ❌ Do not alter `completionScript`'s verbatim return or weaken `TestEmbeddedCompletionsMatchOnDisk`.
- ❌ No in-process §13 assertions trying to drive live-shell first-Tab behavior.
- ⚠️ zsh manual (fpath) path parity is OPTIONAL; simplest is to leave `completions/_skilldozer` alone.
- ❌ fish needs no option change (lists by default); at most one clarifying comment.
