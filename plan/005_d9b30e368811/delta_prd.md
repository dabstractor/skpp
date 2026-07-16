# Delta PRD — Skilldozer: list every match on the first Tab (§14.7)

> **Delta from:** session 004 (`plan/004_5851dcff4371/`), whose work (subcommands → flags, skills-first completions, symlink discovery, derived-zsh emit) is **Complete and green**.
> **Scope of THIS task:** implement PRD §14.7 (new) + the §14.1 rule-2 cross-reference + Decision 22. **Nothing else in the PRD changed.** §6 CLI contract, §7 discovery, §8 store, §9 check, §13 acceptance, the flag set, and `usageText` are all untouched — this delta adds **no flags** and **no CLI behavior**, only a listing *option* set inside the emitted completion scripts.
> **Size:** small. Three touch points (1 on-disk completion file, 1 Go const, README) + tests. Do not expand scope.

---

## 1. What actually changed (diff)

A precise `diff` of the two PRD snapshots shows exactly three edits, all in the completions area:

1. **§14.1 rule 2** — one sentence appended: *"Filtering narrows the set; §14.7 additionally requires the shell to **show** that set (list every prefix match on the first Tab — never a silent halt at the common prefix)."* (cross-reference only; no behavior of its own).
2. **NEW §14.7** *"Listing behavior — show every match, never a silent halt"* (~24 lines) — the core new requirement. Selector `h3.21`.
3. **§19 Decision 22** *"Completion listing (ambiguous matches)"* — the decisions-log mirror of §14.7.

No other section changed. (§13 acceptance criteria are unchanged — there is no new flag to exercise.)

---

## 2. The new requirement (§14.7, distilled)

When the token being completed matches **two or more** candidates — skill tags *or* long-form flags — the completion must surface the **full filtered candidate set** as visible hints on the **first Tab**, never silently halt at the longest common prefix. Two halves:

1. **Offer every match** — **already true; lock it, do not change it.** skilldozer always offers the complete candidate set (tags via `skilldozer --relative --all`, flags via the frozen long-form set) and lets the shell prefix-filter (§14.1 rule 2). No code change here.
2. **Make the shell SHOW them on the first Tab (the fix this delta adds).** The **emitted `--completions` scripts** must set the shell's list-ambiguous-matches option:
   - **bash:** `bind 'set show-all-if-ambiguous on'` (default is **off** → first Tab completes the common prefix and beeps; the list appears only on the second Tab).
   - **zsh:** `setopt NO_LIST_AMBIGUOUS` (default `LIST_AMBIGUOUS` is **on** → first Tab completes the common prefix and lists only at the exact ambiguous point; `NO_LIST_AMBIGUOUS` + the default `AUTO_LIST` lists **all** prefix matches immediately).
   - **fish:** lists all matches in the pager by default → **NO action**.

These are **session-global** options (they change listing for *every* command in that shell, not just skilldozer). Therefore the emitted scripts must, per §14.7:

- **set** the option (the "just works" default the user asked for);
- **disclose** the change **in the emitted script's comments AND in the README** (§15), naming the exact option set;
- provide a one-line **opt-out** — `bind 'set show-all-if-ambiguous off'` (bash) / `setopt LIST_AMBIGUOUS` (zsh) — so a user who prefers the shell's stock behavior can restore it after `eval`.

> **Why it matters:** the store is manifest-free (§2), so the user often does **not** know a tag ahead of time — discovery-via-completion is the primary way to find a skill. An ambiguous prefix that hides its candidates defeats that.

---

## 3. Where it lands (touch points)

The change is confined to the completion *emit* layer. Nothing in `parseArgs`, `discover`, `resolve`, `skillsdir`, `ui`, or `config` is touched.

| Surface | File | Delivery path | Change |
|---|---|---|---|
| **bash emitted script** | `completions/skilldozer.bash` | **verbatim** — §14.5 `source` path == §14.6 `eval` path (identical bytes) | add active `bind` line + disclosure comment + commented opt-out |
| **zsh emitted script** | `main.go` — `zshEvalRegistration` const (line ~1260) | **derived** — eval path only (`runCompletion` → `zshEvalScript` appends this const) | add active `setopt` line + disclosure comment + commented opt-out |
| **fish** | `completions/skilldozer.fish` | verbatim | **NO action** (default lists); optional one-line clarifying comment |
| **README** | `README.md` — *Shell completions* section (line ~290) | — | disclosure paragraph + opt-out one-liners |

**Structural facts (verified in the current code — leverage, do not re-derive):**

- `completionScript(shell)` (main.go:1215) returns the `//go:embed` vars **verbatim**. `runCompletion` (main.go:1499) emits **bash/fish verbatim** and **derives zsh** via `zshEvalScript` (main.go:1244), which strips the trailing `_skilldozer "$@"` self-call and appends `zshEvalRegistration` (main.go:1260).
  - ⇒ **bash:** editing the on-disk file covers **both** the manual `source` path and the `eval` path (the bytes are identical). One edit, two paths.
  - ⇒ **zsh:** the option belongs in `zshEvalRegistration` (the **eval-time** append), **not** the autoload function body. See the Note below on the zsh manual path.
- `TestEmbeddedCompletionsMatchOnDisk` (main_test.go:3139) locks on-disk == embedded bytes. Editing the on-disk bash file keeps this **green** (the `//go:embed` points at the same file, so both update together). A **rebuild** is required for `skilldozer --completions` to reflect on-disk edits (§14.6 lockstep).
- The **flag set is unchanged** (no flag added/removed/renamed), so the §14.4 lockstep (flag set frozen to `parseArgs`) is **not triggered**, and `usageText`/`--help` need no change.
- **§13 acceptance has no new in-process assertions.** First-Tab listing is a *live-shell-session* behavior that cannot be asserted by running the `skilldozer` binary. The test strategy here is therefore **byte-level assertions on the emitted scripts** (does the bash/zsh output *contain* the option + opt-out?), mirroring the existing `TestZshEvalScript*` / `TestRunCompletion*Script` pattern.

### Note on the zsh manual (fpath) path

§14.7's hard requirement is scoped to *"the emitted `--completions` scripts"* — i.e. the **eval path** (§14.6), which is the headline, "easiest way to load completions" path documented first in the README. The zsh §14.5 manual path (copy `completions/_skilldozer` onto `fpath` + `compinit`) loads the file as an **autoload function** whose body runs **per completion**, so placing a `setopt` there would re-run on every Tab. This delta **requires the eval-path fix** (in `zshEvalRegistration`); **full manual-path parity is OPTIONAL** and may be skipped to keep scope tight. If pursued, place `setopt NO_LIST_AMBIGUOUS` as the **first line of the `_skilldozer()` function body** in `completions/_skilldozer` (harmless to re-run; persists to the session). **Bash needs no such caveat** — its on-disk file *is* its emitted script, so the manual `source` path gets the fix for free.

---

## 4. Reference to completed work (do not redo)

Session 004 already delivered the entire completion machinery this delta builds on:
- The `//go:embed` model + `completionScript` verbatim return + `TestEmbeddedCompletionsMatchOnDisk` byte lock (P2.M2.T1).
- The zsh **derived** emit path (`zshEvalScript` + `zshEvalRegistration` + the eval-safety tests) (P2.M2.T1).
- The skills-first / long-form-only completion files (P1.M2.T1).

This delta **adds one option-setting line + disclosure to two of those existing surfaces** (bash on-disk, zsh-derived registration) and documents it. It does not restructure any of that machinery. The §14.4 lockstep is preserved (flag set unchanged).

---

## 5. Breakdown structure

### Phase P1 — List every match on the first Tab (§14.7)

**Single phase, single milestone.** The change is cohesive (one behavior, the completion emit layer) and small. No sequencing risk beyond the standard "code → test → changeset docs."

#### Milestone P1.M1 — Emitted scripts set the list-ambiguous-matches option (disclosed + opt-out)

##### Task P1.M1.T1 — bash + zsh emitted scripts set the option

###### Subtask P1.M1.T1.S1 — bash: add `show-all-if-ambiguous` to the on-disk (== emitted) script
- **story_points:** 1 · **depends on:** (none)
- **context_scope:**
  1. **File:** `completions/skilldozer.bash` (69 lines). The last line is `complete -F _skilldozer_completion skilldozer`. This file is `//go:embed`-ded (`var bashCompletion`, main.go:55) and emitted **verbatim** by `runCompletion`, so a single edit covers both the §14.5 `source` path and the §14.6 `eval` path.
  2. **Logic:**
     - After the `complete -F _skilldozer_completion skilldozer` line, append a **disclosure comment block** + an **active** `bind` line + a **commented opt-out**. Concretely (exact wording is the implementer's, but must name the option and the opt-out):
       ```bash
       # §14.7 — list ambiguous matches on the FIRST Tab (show-all-if-ambiguous).
       # This is a readline SESSION-GLOBAL option: it changes listing for every
       # command in this shell, not just skilldozer. The default is OFF, so without
       # this the first Tab completes the common prefix and beeps, showing the list
       # only on the second Tab. To restore the shell's stock behavior, run:
       #     bind 'set show-all-if-ambiguous off'
       [[ $- == *i* ]] && bind 'set show-all-if-ambiguous on'
       ```
     - The `[[ $- == *i* ]] &&` guard keeps it quiet if the file is ever sourced in a non-interactive context (e.g. an `eval` test harness); completions are only meaningfully loaded in interactive shells, so the option still applies where it matters. (`bind` in a non-interactive bash prints a warning; the guard silences it.)
  3. **Output:** `completions/skilldozer.bash` ends with the disclosure block + guarded `bind`. `go build ./...` succeeds. `TestEmbeddedCompletionsMatchOnDisk` stays green (on-disk == embedded, same file). `TestRunCompletionBashScript` (asserts `_skilldozer_completion` present + empty stderr) stays green — the new line is appended, not a regression.
  4. **DOCS (Mode A, rides with this subtask):** the disclosure comment block above **IS** the §14.7 "disclose in the emitted script's comments" requirement for bash. No separate doc step.
- **prd_selectors:** `h3.21`, `h3.15`, `h3.19`, `h2.13`

###### Subtask P1.M1.T1.S2 — zsh: add `NO_LIST_AMBIGUOUS` to the derived eval registration
- **story_points:** 1 · **depends on:** (none — independent of S1)
- **context_scope:**
  1. **File:** `main.go`. The `zshEvalRegistration` raw-string const is at line ~1260; it is appended by `zshEvalScript` (main.go:1244) to form the **derived** eval output (`runCompletion`, main.go:1499, derives zsh only; bash/fish stay verbatim). `completionScript` still returns the on-disk autoload file **verbatim** (§14.6 byte-identity lock) — **do not** touch `completionScript` or `completions/_skilldozer` for the required fix.
  2. **Logic:**
     - Inside the `zshEvalRegistration` const (currently `autoload -Uz compinit` / guarded `compinit` / `compdef _skilldozer skilldozer`), add — before or after the compdef block — an **active** `setopt NO_LIST_AMBIGUOUS` line plus a **disclosure comment + commented opt-out**:
       ```zsh
       # §14.7 — list ambiguous matches on the FIRST Tab (NO_LIST_AMBIGUOUS).
       # This is a SESSION-GLOBAL zsh option (default LIST_AMBIGUOUS is on, which
       # completes the common prefix and lists only at the exact ambiguous point).
       # NO_LIST_AMBIGUOUS + the default AUTO_LIST lists all prefix matches at once.
       # It changes listing for every command in this shell, not just skilldozer.
       # To restore the shell's stock behavior, run:  setopt LIST_AMBIGUOUS
       setopt NO_LIST_AMBIGUOUS
       ```
     - Keep the const a valid Go raw string literal (no backticks inside — the existing const already observes this).
     - Update the `zshEvalRegistration` doc comment (main.go:1257) to mention it now also sets the §14.7 listing option. Update the `zshEvalScript` / `runCompletion` doc comments only if they claim the registration is *only* about compdef (broaden the wording; do not over-edit).
  3. **Output:** `eval "$(skilldozer --completions)"` (zsh) now sets `NO_LIST_AMBIGUOUS` at eval time. `go build ./...` succeeds. `completionScript("zsh")` is still byte-identical to `completions/_skilldozer` (the const is appended **after** `completionScript` returns; the byte-identity lock is untouched).
  4. **DOCS (Mode A, rides with this subtask):** the disclosure comment in the const **IS** the §14.7 "disclose in the emitted script's comments" requirement for zsh. The broadened doc comments on `zshEvalRegistration` are the developer-facing reference.
- **prd_selectors:** `h3.21`, `h3.15`, `h3.20`, `h4.2`, `h2.13`, `h2.18`

##### Task P1.M1.T2 — Tests locking the emitted-byte contract

###### Subtask P1.M1.T2.S1 — Assert bash + zsh emitted output carry the option + opt-out
- **story_points:** 1 · **depends on:** `P1.M1.T1.S1`, `P1.M1.T1.S2`
- **context_scope:**
  1. **File:** `main_test.go`. Existing pattern to follow: `TestZshEvalScriptStripsSelfCall` / `TestZshEvalScriptRegistersCompdef` / `TestRunCompletionZshIsEvalSafe` (lines 3266/3288/3316) and `TestRunCompletionBashScript` (line 3163).
  2. **Logic (add/extend — exact names are the implementer's):**
     - **bash:** add `TestRunCompletionBashListsAmbiguous` (or extend `TestRunCompletionBashScript`) — `run([]string{"--completions", "--shell", "bash"}, …)`, exit 0, then assert `out.String()` **contains** `show-all-if-ambiguous on` (the active line) **and** contains the opt-out token `show-all-if-ambiguous off` (disclosed in the comment). Optionally assert the interactivity guard `*i*` is present.
     - **zsh:** extend `TestZshEvalScriptRegistersCompdef` (or add `TestZshEvalScriptSetsNoListAmbiguous`) — `zshEvalScript(completionScript("zsh"))` must **contain** `setopt NO_LIST_AMBIGUOUS` (active) **and** the opt-out token `setopt LIST_AMBIGUOUS` (disclosed). Extend `TestRunCompletionZshIsEvalSafe` to also assert the end-to-end emitted script (via `run(["--completions"], …)` under `SKILLDOZER_SHELL=zsh`) contains `NO_LIST_AMBIGUOUS`.
     - **byte-identity still holds:** `TestEmbeddedCompletionsMatchOnDisk` must remain green (it is — no on-disk file that `completionScript` returns was altered for the zsh fix; the bash on-disk edit updates both sides consistently). Do not weaken this test.
  3. **Output:** `go test ./...` is 100% green. No existing green test regresses. The §14.7 emitted-byte contract is locked for bash + zsh.
  4. **DOCS:** none (test-only).
- **prd_selectors:** `h3.21`, `h2.12`, `h2.13`

##### Task P1.M1.T3 — Sync README disclosure (Mode B, depends on T1 + T2)

###### Subtask P1.M1.T3.S1 — README: disclose the option set + opt-out one-liners
- **story_points:** 1 · **depends on:** `P1.M1.T1.S1`, `P1.M1.T1.S2`, `P1.M1.T2.S1`
- **context_scope:**
  1. **File:** `README.md`, *Shell completions* section (starts line ~290). This is the **Mode B changeset-level documentation** for this delta — it only makes sense once both emitted scripts carry the option.
  2. **Logic:** After the existing "Once loaded, completions are skills-first and long-form-only" bullet list (which already documents `<tab>` lists skills and `--c<tab>` lists `--check`/`--completions`), add a short **disclosure paragraph** + opt-out block. Concretely (exact wording is the implementer's):
     - State that the emitted script sets a **session-global** option so **ambiguous matches list on the first Tab** instead of halting at the common prefix, and **name** the option per shell: bash `show-all-if-ambiguous`, zsh `NO_LIST_AMBIGUOUS` (fish already lists by default — no option set).
     - Note this affects listing for **every command** in that shell, not just skilldozer, and that it is set only when you load skilldozer completions.
     - Give the **opt-out** one-liners so a user can restore stock behavior:
       ```bash
       # bash — restore the default (list on second Tab):
       bind 'set show-all-if-ambiguous off'
       # zsh — restore the default (list at the exact ambiguous point):
       setopt LIST_AMBIGUOUS
       ```
  3. **Output:** README *Shell completions* section discloses the option + opt-out. Verify with: `grep -q 'show-all-if-ambiguous' README.md && grep -q 'NO_LIST_AMBIGUOUS' README.md && grep -q 'LIST_AMBIGUOUS' README.md`. The existing README content (eval one-liners, flag list, manual copy paths) is preserved.
  4. **DOCS:** **This IS the Mode B changeset-level documentation task.** It depends on the implementing subtasks and runs last. It fulfills the §14.7 / §15 "disclose in the README" half of the requirement (the other half — disclosure in the emitted script comments — rode with T1).
- **prd_selectors:** `h2.14`, `h3.21`, `h2.13`

---

## 6. Out of scope (do NOT do)

- ❌ Do **not** add, remove, or rename any flag. The flag set is unchanged; §14.4 lockstep and `usageText` are not triggered.
- ❌ Do **not** touch `parseArgs`, `discover`, `resolve`, `skillsdir`, `ui`, `config`, `runInit`, `runCheck`, or any tag-resolution / store code.
- ❌ Do **not** alter `completionScript`'s verbatim return or weaken `TestEmbeddedCompletionsMatchOnDisk` — the §14.6 byte-identity lock must hold.
- ❌ Do **not** add in-process §13 assertions that try to drive a live shell's first-Tab behavior — it cannot be asserted by running the binary; assert the **emitted bytes** instead (T2).
- ⚠️ **zsh manual (fpath) path parity is OPTIONAL.** The required fix is the eval-path registration (`zshEvalRegistration`). If you also edit `completions/_skilldozer`, you MUST keep `completionScript("zsh")` byte-identical to it — i.e. the autoload-file edit must not depend on the derivation. (Simplest: leave the autoload file alone.)
- ❌ Do **not** change fish (`completions/skilldozer.fish`) beyond an optional one-line clarifying comment — it already lists by default (§14.7).
