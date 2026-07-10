name: "P1.M3.T1.S1 — Lockstep `completion` into the three shell completion files (first-arg exclusive subcommand)"
description: |

---

## Goal

**Feature Goal**: Make all three shell completion files (`completions/skilldozer.bash`, `completions/_skilldozer`, `completions/skilldozer.fish`) offer the new `completion` subcommand as a **completable first-arg exclusive subcommand**, gated identically to the existing `check`/`init` subcommands — fulfilling PRD §14.4 (lockstep with `main.go parseArgs()`) and §14.6 now that `completion` is a reserved exclusive token (P1.M2.T1.S1). Tab-completing `skilldozer <TAB>` now offers `completion` alongside `check`/`init`; once any of the three is typed, tag completion is suppressed (mirrors §6.3 exclusivity).

**Deliverable**: Six small shell-script edits across three existing files (2 per file) — NO Go code, NO new files:
1. `completions/skilldozer.bash` — add `completion` to (a) the suppression-walk `[[ ]]` test and (b) the first-positional `cands` offer.
2. `completions/_skilldozer` (zsh) — add `completion` to (a) the `first)`-state `compadd` and (b) the `rest)`-state suppression `if (( ))` condition.
3. `completions/skilldozer.fish` — (a) add a new `__fish_is_first_arg` directive for `completion` after the `init` directive, and (b) add `completion` to the `__fish_seen_subcommand_from` suppression predicate.

**Success Definition**: each file offers `completion` only as the first positional arg and suppresses tag completion once it (or `check`/`init`) is seen; all three files pass their shell syntax check (`bash -n` / `zsh -n` / `fish --no-execute`); the `//go:embed` byte-identity stays in sync (`go test ./...` green, incl. `TestEmbeddedCompletionsMatchOnDisk`); `--shell` is NOT added to any flag matrix; the tag-completion probe (`skilldozer --relative --all 2>/dev/null`) and the `--search`/`--store` routing are byte-for-byte unchanged; a rebuilt binary's `skilldozer completion --shell bash` output now contains the `completion` token (the §14.6 embedded-bytes delivery path reflects the edit).

---

## User Persona (if applicable)

**Target User**: A `skilldozer` user typing at a bash/zsh/fish prompt who wants to install completions via the new `eval "$(skilldozer completion)"` idiom (PRD §14.6) — and now discovers it by tab-completing the subcommand name rather than reading docs.

**Use Case**: `skilldozer <TAB>` offers `check`, `init`, and `completion` (plus the dynamic tag list); the user picks `completion`, then runs `eval "$(skilldozer completion)"`. Once `completion` is typed, tag completion disappears (it would be a guaranteed §6.3 exit-2 error).

**User Journey**: before this task, `skilldozer <TAB>` offered `check`/`init` but NOT `completion` (the new subcommand was invisible to tab-completion); after, all three exclusive subcommands are discoverable and consistently gated.

**Pain Points Addressed**: the new `completion` subcommand existed in the binary but was invisible to shell completion — a lockstep gap (§14.4) between `parseArgs()` (which knows `completion`) and the three completion files (which didn't).

---

## Why

- **Closes the §14.4 lockstep gap.** §14.4 (guardrail, also §17): "if a future task adds … a flag [or reserved token] in `main.go parseArgs()`, update all three completion files identically." P1.M2.T1.S1 added the `completion` reserved token; this task is the matching completion-file update. `completion` is exclusive like `check`/`init` (main.go:166/297/534), so it gets the identical three-part treatment.
- **Makes §14.6's `completion` subcommand discoverable.** PRD §14.6 documents `eval "$(skilldozer completion)"` as the easy install path; tab-completion is the natural way to find a subcommand. Without this lockstep, the subcommand is invisible to completion.
- **Keeps the three delivery paths consistent (§14.6 lockstep note).** The completion files are the single source of truth; both the §14.5 manual source/copy path AND the `//go:embed`-driven `skilldozer completion` emit path must offer `completion`. Editing the on-disk files (this task) + a rebuild makes both paths agree.
- **[Mode A docs-with-work]**: the completion files ARE the user-facing surface for tab-completion — updating them here satisfies doc-with-work. No separate doc subtask for the lockstep itself (README coverage of `completion` is the sibling P1.M3.T1.S2).

---

## What

Six edits mirroring the existing `check`/`init` pattern in each file. `completion` is gated IDENTICALLY to `check`/`init` in all three: offered only as the first positional; once seen, tag completion is suppressed.

**(bash)**
1. Suppression walk: `[[ "${words[i]}" == "check" || "${words[i]}" == "init" ]] && return 0` → add `|| "${words[i]}" == "completion"`.
2. First-pos offer: `(( have_pos == 0 )) && cands="$cands check init"` → `cands="$cands check init completion"`.

**(zsh)**
1. First-pos `compadd` (`first)` state): `compadd -- "$tags[@]" check init` → `compadd -- "$tags[@]" check init completion`.
2. Suppression (`rest)` state): `if (( ${words[(I)check]} || ${words[(I)init]} )); then` → add `|| ${words[(I)completion]}`.

**(fish)**
1. Add a new directive after the `init` directive: `complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'`.
2. Tag-suppression predicate: `__fish_seen_subcommand_from check init` → `__fish_seen_subcommand_from check init completion`.

### Success Criteria

- [ ] All three files offer `completion` as a first-positional subcommand (gated identically to `check`/`init`).
- [ ] All three files suppress tag completion once `completion` (or `check`/`init`) is seen.
- [ ] `bash -n`, `zsh -n`, and `fish --no-execute` all exit 0 on the edited files.
- [ ] `grep -q 'completion'` succeeds in all three files.
- [ ] `--shell` is NOT added to any flag matrix (the `compgen -W`/`flags=(...)`/`complete … -l …` flag sets are unchanged).
- [ ] The tag probe `skilldozer --relative --all 2>/dev/null` is byte-for-byte unchanged in all three files.
- [ ] The `--search`/`--store` value routing is unchanged in all three files.
- [ ] `go test ./...` is green (incl. `TestEmbeddedCompletionsMatchOnDisk` — the §14.6 byte-identity lock survives because `go test` rebuilds/re-embeds).
- [ ] Only the three `completions/*` files are modified; no Go file, README, PRD, or `.gitignore` touched.

---

## All Needed Context

### Context Completeness Check

**Pass.** The authoritative per-file edit list (`architecture/completions_change_map.md`) was read in full AND every anchor cross-checked against the live files (the exact `check`/`init` strings to extend are quoted in research/verified_facts.md §2). `completion`'s exclusive status — which justifies mirroring the suppression-walk treatment — was verified in main.go (main.go:166 "exclusive like check/init"; main.go:297/534 exclusivity + dispatch). The fish `-d` description was confirmed byte-consistent with the binary's USAGE row (main.go:107). The `--shell` scope boundary (do NOT add to the flag matrix) is pinned to the contract + §14.1/§14.2. All three shell syntax-checkers are confirmed installed with the current files passing. The embed-sync test (`TestEmbeddedCompletionsMatchOnDisk`, main_test.go:2929) was read and its rebuild-re-embed behavior confirmed (so the edit keeps it green). An implementer who has never seen this repo can complete it in one pass: six text-anchored substitutions mirroring an in-file pattern, validated by shell syntax checks + grep + go test + a rebuild-and-emit check.

### Documentation & References

```yaml
# MUST READ — the verified facts (exact anchors, edits, scope boundary, validation, embed sync)
- file: plan/003_3ace946c2a5c/P1M3T1S1/research/verified_facts.md
  why: "§1 = completion is exclusive like check/init (main.go:166/297/534) → the suppression-walk
        treatment is correct. §2 = the EXACT current check/init strings to extend in each file (the
        anchors). §3 = the six edits. §4 = fish -d description == USAGE row (main.go:107). §5 = the
        CRITICAL --shell scope boundary (do NOT add to flag matrix; why). §6 = the four 'do NOT'
        constraints. §7 = shell syntax-checkers available + current files pass. §8 = embed sync
        (go test rebuilds → TestEmbeddedCompletionsMatchOnDisk stays green; rebuild required for a
        standalone binary's emit). §9/§10 = scope + parallel-execution (disjoint from P1.M2.T2.S2)."
  critical: "§5 — do NOT add --shell to any flag matrix (--shell is completion-context-only; the
             flag matrix is §6.1/§6.2 global flags only). §6 — do NOT touch the tag probe or the
             --search/--store routing. §8 — go test rebuilds so the embed test stays green, but a
             pre-built binary needs a rebuild to emit the new bytes."

# MUST READ — the authoritative per-file edit list (cross-checked against live files)
- file: plan/003_3ace946c2a5c/architecture/completions_change_map.md
  why: "The exact CURRENT→TARGET for all six edits (2 per file) with line numbers. Every anchor there
        was verified to match the live files at PRP-write time. The 'Cross-file constraints' block
        restates the do-NOTs (--shell, tag probe, --search/--store)."
  critical: "Anchor each edit by the TEXT in this map + research §2, not the line numbers (the files
             are short/stable, but text-anchoring is unambiguous and survives any reflow)."

# MUST READ — the three files under edit (the ONLY deliverable)
- file: completions/skilldozer.bash
  why: "EDIT (2): the suppression-walk `[[ ... == \"check\" || ... == \"init\" ]] && return 0` line
        (add `|| \"${words[i]}\" == \"completion\"`) and the `(( have_pos == 0 )) && cands=\"$cands
        check init\"` line (append ` completion`). Mirror the in-file check/init pattern verbatim."
  pattern: "Exclusive-subcommand handling = (a) a suppression walk that `return 0`s (offers nothing)
            once the token is seen among earlier words, + (b) a first-positional-only offer appended to
            the candidate list. `completion` joins both exactly as `check`/`init` do."
  gotcha: "Keep the `[[ ]]` quoting style (`\"${words[i]}\"` quoted) — a missing quote breaks bash
           syntax (caught by `bash -n`). Do NOT touch the `compgen -W \"...\"` flag list (no --shell)."
- file: completions/_skilldozer
  why: "EDIT (2): the `first)`-state `compadd -- \"$tags[@]\" check init` (append ` completion`) and
        the `rest)`-state `if (( ${words[(I)check]} || ${words[(I)init]} )); then` (add `||
        ${words[(I)completion]}`)."
  pattern: "zsh exclusive-subcommand handling = `compadd` the names in the `first)` state +
            `${words[(I)<token>]}` (reverse-index membership test) in the `rest)` state to suppress."
  gotcha: "`${words[(I)completion]}` is the zsh subscript-flag idiom (the `(I)` reverse-index returns 0
           if absent, nonzero if present — the `(( ))` arith test treats nonzero as true). Mirror
           `check`/`init` EXACTLY. Do NOT touch the `_arguments` `flags=(...)` list (no --shell)."
- file: completions/skilldozer.fish
  why: "EDIT (2): ADD a new `complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d
        'Emit the shell completion script for eval'` line after the `init` directive; and change
        `__fish_seen_subcommand_from check init` → `__fish_seen_subcommand_from check init completion`
        in the tag-suppression predicate."
  pattern: "fish exclusive-subcommand handling = one `complete ... -n '__fish_is_first_arg' -a '<name>'`
            directive per subcommand + `__fish_seen_subcommand_from <names...>` in the suppression
            predicate of the tag directive."
  gotcha: "The `-d` text MUST be 'Emit the shell completion script for eval' (matches the binary's USAGE
           row, main.go:107). Place the new directive immediately AFTER the `init` directive (keeps
           check/init/completion together). Do NOT add a `--shell` flag directive (no --shell)."

# READ-ONLY — confirms completion is exclusive (justifies the suppression-walk mirror)
- file: main.go
  why: "main.go:166 `completion bool // ... exclusive like check/init`; main.go:297 exclusivity rejects
        completion+tags/completion+mode; main.go:534 'completion is an exclusive mode (like ...)'.
        main.go:107 USAGE row 'completion [--shell <name>]   Emit the shell completion script for eval'
        (the source of the fish -d text). main.go:228 `case \"--shell\"` (--shell IS parsed but is
        completion-context-only → NOT a flag-matrix entry). READ-ONLY — do NOT edit main.go."
  section: "config struct (166-167), parseArgs --shell (228-233) + completion token (297-321),
            run() dispatch (534), USAGE (107)."

# READ-ONLY — the parallel sibling (disjoint files; defines the emit path this task feeds)
- file: plan/003_3ace946c2a5c/P1M2T2S2/PRP.md
  why: "Defines run()→runCompletion→completionScript: `skilldozer completion --shell bash` emits the
        //go:embed'd bytes. After THIS task edits the files + a rebuild, that emit path includes the
        `completion` candidate. S2 edits ONLY main.go + main_test.go; this task edits ONLY
        completions/* → DISJOINT, no conflict."

# READ-ONLY — PRD (the lockstep guardrail + the completion subcommand contract)
- file: PRD.md
  why: "§14.4 (h3.17): the flag/token set is frozen to parseArgs() — update all three files
        identically (this task). §14.6 (h3.19): completion emits the embedded script; the on-disk
        files are the single source of truth; editing them needs a rebuild for the embed path. §14.1
        (h3.14): the flag matrix is §6.1/§6.2 only. §17 (h2.16): the completion-lockstep guardrail."
  section: "h3.17 (§14.4), h3.19 (§14.6), h3.14 (§14.1), h2.16 (§17)."
  gotcha: "READ-ONLY. PRD §14.1 still says exclusive subcommands are 'check and init' — updating it to
           mention completion is a PRD edit (human-owned), out of scope here. The completion FILES +
           README (P1.M3.T1.S2) are the surfaces that matter."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls completions/
_skilldozer  skilldozer.bash  skilldozer.fish
$ wc -l completions/*
  69 completions/skilldozer.bash   # EDIT: 2 lines (suppression walk + first-pos offer)
  61 completions/_skilldozer       # EDIT: 2 lines (first-state compadd + rest-state suppression)
  51 completions/skilldozer.fish   # EDIT: +1 directive (completion) + 1 predicate line (add completion)
$ grep -c completion completions/*   # today: only the function-name "_skilldozer_completion" / comments;
                                     # NOT offered as a completable subcommand in any file
# main.go / main_test.go / internal/* / README.md / PRD.md — UNCHANGED.
```

### Desired Codebase tree with files to be changed

```bash
completions/skilldozer.bash   # +completion in the suppression walk + the first-pos offer
completions/_skilldozer       # +completion in the first-state compadd + the rest-state suppression
completions/skilldozer.fish   # +1 __fish_is_first_arg directive for completion; +completion in the predicate
# every other file UNCHANGED (no Go, no README, no PRD, no .gitignore).
```

| File | Change | Why |
|---|---|---|
| `completions/skilldozer.bash` | Add `completion` to the suppression walk + first-pos offer. | §14.4 lockstep: `completion` is now an exclusive reserved token. |
| `completions/_skilldozer` | Add `completion` to the first-state compadd + rest-state suppression. | same |
| `completions/skilldozer.fish` | Add a `completion` first-arg directive + add it to the suppression predicate. | same |

### Known Gotchas of our codebase & Library Quirks

```bash
# GOTCHA #1 (CRITICAL — scope) — Do NOT add `--shell` to ANY flag matrix. The contract + §14.1/§14.2
# fix the completion files' flag matrix to the §6.1/§6.2 GLOBAL flag set ONLY; --shell is completion-
# context-only (it only makes sense after `completion`). So: do NOT add --shell to bash's `compgen -W`
# list, zsh's `_arguments flags=(...)`, or fish's `complete ... -l shell` directives. Consequence:
# `skilldozer completion --<TAB>` will NOT offer --shell (the user types it); that is the deliberate
# decision — honor it, do not "help". (research §5.)

# GOTCHA #2 (CRITICAL — exclusivity justifies the suppression walk) — `completion` is EXCLUSIVE like
# check/init (main.go:166/297/534: completion+tags / completion+mode → exit 2). That is WHY it gets the
# identical three-part treatment (offer first-pos only; suppress tags once seen). Mirror check/init
# EXACTLY — there is nothing special about completion from the completion-file perspective. (research §1.)

# GOTCHA #3 (do NOT change the tag probe) — `skilldozer --relative --all 2>/dev/null` must stay
# byte-for-byte identical in all three files (bash `tags=$(...)`, zsh `tags=(${(f)"$(...)"})`, fish
# `-a '(...)'`). This task only adds the `completion` SUBCOMMAND candidate, never touches tag sourcing.

# GOTCHA #4 (do NOT change --search/--store routing) — the value-flag routing stays unchanged: bash
# `case "$prev"`, zsh `:query:`/`:directory:_files` specs, fish the `-r` (store) / no-`-r` (search)
# directives. Only the exclusive-subcommand surface (check/init → check/init/completion) changes.

# GOTCHA #5 (fish -d text is fixed) — the completion directive's description MUST be
# 'Emit the shell completion script for eval' (matches the USAGE row main.go:107). Do not paraphrase.
# Place the directive AFTER the `init` directive (keeps check/init/completion contiguous).

# GOTCHA #6 (embed sync — §14.6) — the three files are //go:embed'd (main.go:54/57/60).
# TestEmbeddedCompletionsMatchOnDisk (main_test.go:2929) asserts embed bytes == on-disk bytes.
# `go test` REBUILDS (re-embeds), so editing the files KEEPS that test green automatically — but a
# PRE-BUILT binary will NOT reflect the edit until `go build` re-runs. So validation must rebuild and
# confirm `skilldozer completion --shell bash` output now contains `completion`. (research §8.)

# GOTCHA #7 (anchor by TEXT, not line number) — the files are short/stable, but the change-map line
# numbers are advisory. Locate each edit by the exact check/init string in research §2 (e.g. the bash
# `[[ ... == "check" || ... == "init" ]] && return 0` line). Text-anchoring is unambiguous.

# GOTCHA #8 (shell quoting/zsh subscript) — in bash keep the `\"${words[i]}\"` quoting in the `[[ ]]`;
# in zsh the `${words[(I)completion]}` subscript is the membership-test idiom (mirror check/init's
# exact form). A stray quote/paren fails `bash -n`/`zsh -n` — run the syntax check after each file.

# GOTCHA #9 (no Go/README/PRD edits) — this is the shell-file lockstep only. main.go/main_test.go are
# P1.M2's scope; README is P1.M3.T1.S2; PRD.md is read-only. PRD §14.1 still lists "check and init" —
# updating it is a human action item, out of scope.
```

---

## Implementation Blueprint

### Data models and structure

**None.** Shell-script text edits only. No types, no code, no config. The six substitutions ARE the deliverable.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT completions/skilldozer.bash (2 edits)
  - FILE: completions/skilldozer.bash
  - EDIT 1 (suppression walk): locate the line
        [[ "${words[i]}" == "check" || "${words[i]}" == "init" ]] && return 0
    and change it to
        [[ "${words[i]}" == "check" || "${words[i]}" == "init" || "${words[i]}" == "completion" ]] && return 0
    (GOTCHA #8 — keep the `"${words[i]}"` quoting.)
  - EDIT 2 (first-pos offer): locate
        (( have_pos == 0 )) && cands="$cands check init"
    and change it to
        (( have_pos == 0 )) && cands="$cands check init completion"
  - Do NOT touch the `compgen -W "..."` flag list (GOTCHA #1 — no --shell) or the tag probe (GOTCHA #3).
  - VERIFY: bash -n completions/skilldozer.bash && echo OK   # exit 0

Task 2: EDIT completions/_skilldozer (2 edits)
  - FILE: completions/_skilldozer
  - EDIT 1 (first-state compadd): locate
            compadd -- "$tags[@]" check init
    and change it to
            compadd -- "$tags[@]" check init completion
  - EDIT 2 (rest-state suppression): locate
            if (( ${words[(I)check]} || ${words[(I)init]} )); then
    and change it to
            if (( ${words[(I)check]} || ${words[(I)init]} || ${words[(I)completion]} )); then
  - Do NOT touch the `_arguments` `flags=(...)` list (GOTCHA #1) or the tag probe (GOTCHA #3).
  - VERIFY: zsh -n completions/_skilldozer && echo OK   # exit 0

Task 3: EDIT completions/skilldozer.fish (2 edits)
  - FILE: completions/skilldozer.fish
  - EDIT 1 (new directive): locate the `init` first-arg directive
        complete -c skilldozer -n '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'
    and ADD immediately AFTER it a new line
        complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'
    (GOTCHA #5 — the -d text is fixed; place after the init directive.)
  - EDIT 2 (suppression predicate): locate
        complete -c skilldozer -n 'not __fish_seen_subcommand_from check init; and not __fish_prev_arg_in --search -s' \
    and change `check init` to `check init completion`:
        complete -c skilldozer -n 'not __fish_seen_subcommand_from check init completion; and not __fish_prev_arg_in --search -s' \
    (preserve the trailing backslash line-continuation.)
  - Do NOT add a --shell flag directive (GOTCHA #1) or touch the tag `-a '(skilldozer --relative --all 2>/dev/null)'` (GOTCHA #3).
  - VERIFY: fish --no-execute completions/skilldozer.fish && echo OK   # exit 0

Task 4: VERIFY all gates (syntax + grep + embed sync + rebuild-and-emit)
  - bash -n completions/skilldozer.bash && zsh -n completions/_skilldozer && fish --no-execute completions/skilldozer.fish
  - grep -q 'completion' completions/skilldozer.bash && grep -q 'completion' completions/_skilldozer && grep -q 'completion' completions/skilldozer.fish
  - go test ./...   # green incl. TestEmbeddedCompletionsMatchOnDisk (§14.6 byte-identity; rebuild re-embeds)
  - go build -o /tmp/sd . && /tmp/sd completion --shell bash | grep -q 'completion' && echo "emit reflects edit OK"
  - see Validation Loop for the full gate set + the do-NOT regression checks.
```

### Implementation Patterns & Key Details

```bash
# The deliverable mirrors the in-file check/init pattern. Each file has THREE exclusive subcommands
# after this task: check / init / completion. The pattern per shell:
#
#   bash : (a) suppression walk `[[ ... == "check" || ... == "init" || ... == "completion" ]] && return 0`
#          (b) first-pos offer   `cands="$cands check init completion"`
#   zsh  : (a) first-state compadd `compadd -- "$tags[@]" check init completion`
#          (b) rest-state suppress `if (( ${words[(I)check]} || ${words[(I)init]} || ${words[(I)completion]} ))`
#   fish : (a) three `complete -c skilldozer -n '__fish_is_first_arg' -a '<name>' -d '...'` directives
#          (b) tag predicate `not __fish_seen_subcommand_from check init completion; and ...`
#
# Three things that MUST stay byte-identical (do NOT touch):
#   1. tag probe: `skilldozer --relative --all 2>/dev/null` (all three files)
#   2. --search/--store value routing (the case/`:spec:`/-r handling)
#   3. the flag matrix (compgen -W / flags=(...) / complete ... -l ...) — NO --shell anywhere

# fish directive placement (before -> after):
#   BEFORE:
#     complete -c skilldozer -n '__fish_is_first_arg' -a 'check'  -d 'Validate every skill on disk'
#     complete -c skilldozer -n '__fish_is_first_arg' -a 'init'   -d 'First-run setup: ...'
#     <blank>
#     # Dynamic tags: ...
#   AFTER:
#     complete -c skilldozer -n '__fish_is_first_arg' -a 'check'      -d 'Validate every skill on disk'
#     complete -c skilldozer -n '__fish_is_first_arg' -a 'init'       -d 'First-run setup: ...'
#     complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'
#     <blank>
#     # Dynamic tags: ...
```

Notes easy to get wrong:
- Adding `--shell` to a flag matrix (GOTCHA #1) — the most tempting "while I'm here" slip; forbidden.
- Paraphrasing the fish `-d` text (GOTCHA #5) — it must equal the USAGE row.
- Forgetting the rebuild for the emit check (GOTCHA #6) — a pre-built binary won't show the edit.
- Breaking shell quoting in the bash `[[ ]]` or the zsh `${words[(I)...]}` (GOTCHA #8) — `bash -n`/`zsh -n` catch it.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **`completion` mirrors `check`/`init` exactly (no special-casing).** It is exclusive like them (main.go:166/297/534), so it gets the identical three-part treatment. There is no `--shell`-aware offer — the contract forbids `--shell` in the flag matrix (decision 2 below). (research §1/§2.)
2. **`--shell` is NOT added to any flag matrix.** The contract + §14.1/§14.2 fix the matrix to the §6.1/§6.2 global flags; `--shell` is completion-context-only. Adding it globally would be wrong (`skilldozer --shell bash` is not a valid invocation outside `completion`); a context-gated offer is out of scope for this 1-point lockstep. The user types `--shell` manually. (GOTCHA #1, research §5.)
3. **The fish directive goes after `init` (contiguous check/init/completion).** Keeps the exclusive-subcommand block together and matches the change map's "add after the init directive". (research §3.)
4. **Validation = shell syntax checks + grep + go test (embed sync) + rebuild-and-emit.** The syntax checks are the strongest signal (a shell typo fails immediately); `go test` proves the embed stays byte-identical to the on-disk files; the rebuild-and-emit proves the §14.6 delivery path reflects the edit. (research §7/§8.)
5. **PRD §14.1 ("check and init") is NOT updated.** It is a PRD edit (human-owned, read-only). The completion files + README (P1.M3.T1.S2) are the user-facing surfaces that matter; the PRD lag is a recorded human action item. (research §9.)

### Integration Points

```yaml
COMPLETION FILES (the deliverable):
  - files: completions/skilldozer.bash, completions/_skilldozer, completions/skilldozer.fish
  - effect: "All three now offer `completion` as a first-arg exclusive subcommand, gated identically to
            check/init; tag completion suppressed once any of the three is seen."

EMBED PATH (§14.6 — consumed, not modified here):
  - main.go:54/57/60 //go:embed the three files; completionScript(shell) (main.go:1121) returns them.
  - runCompletion (P1.M2.T2.S2) emits them via `skilldozer completion`. After this task + a rebuild,
    `skilldozer completion --shell bash` output includes the `completion` candidate.
  - TestEmbeddedCompletionsMatchOnDisk (main_test.go:2929) locks embed==on-disk; go test rebuilds → green.

CODE: NONE.
  - main.go / main_test.go / internal/* UNCHANGED (P1.M2's scope). go.mod/go.sum UNCHANGED.

PRD.md / tasks.json / prd_snapshot.md: READ-ONLY. README.md: P1.M3.T1.S2 (sibling).

PARALLEL SIBLING (no conflict):
  - P1.M2.T2.S2 edits ONLY main.go + main_test.go. This task edits ONLY completions/*. DISJOINT.

NO DATABASE / NO ROUTES / NO CONFIG-FORMAT CHANGE / NO GO CODE / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Shell syntax (immediate, after each file's edit)

```bash
cd /home/dustin/projects/skilldozer

# All three must parse (a stray quote/paren fails the matching checker):
bash -n completions/skilldozer.bash   && echo "bash -n OK"
zsh  -n completions/_skilldozer       && echo "zsh -n OK"
fish --no-execute completions/skilldozer.fish && echo "fish --no-execute OK"
# Expected: all three print OK; exit 0. (Baseline confirmed: the pre-edit files pass these too.)
```

### Level 2: Required-content + do-NOT-regression checks (the contract's OUTPUT)

```bash
cd /home/dustin/projects/skilldozer

# (a) `completion` is offered as a completable subcommand in all three files:
grep -q 'completion' completions/skilldozer.bash && echo "bash: completion present OK"
grep -q 'completion' completions/_skilldozer     && echo "zsh: completion present OK"
grep -q 'completion' completions/skilldozer.fish && echo "fish: completion present OK"

# (b) bash: suppression walk + first-pos offer both extended:
grep -q '"completion" ]] && return 0' completions/skilldozer.bash && echo "bash: suppression OK"
grep -q 'cands="$cands check init completion"' completions/skilldozer.bash && echo "bash: offer OK"

# (c) zsh: first-state compadd + rest-state suppression both extended:
grep -q 'compadd -- "\$tags\[@\]" check init completion' completions/_skilldozer && echo "zsh: compadd OK"
grep -q '\${words\[(I)completion\]}' completions/_skilldozer && echo "zsh: suppression OK"

# (d) fish: new directive + predicate both extended:
grep -q "__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'" completions/skilldozer.fish && echo "fish: directive OK"
grep -q '__fish_seen_subcommand_from check init completion' completions/skilldozer.fish && echo "fish: predicate OK"

# (e) DO-NOT regressions (these must be UNCHANGED):
#   tag probe byte-identical:
grep -q 'skilldozer --relative --all 2>/dev/null' completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish && echo "tag probe unchanged OK"
#   --shell NOT added to any flag matrix:
! grep -q -- '--shell' completions/skilldozer.bash && echo "bash: no --shell in matrix OK"
! grep -q -- '--shell' completions/_skilldozer && echo "zsh: no --shell in matrix OK"
! grep -q -- '-l shell' completions/skilldozer.fish && echo "fish: no --shell directive OK"
#   --search/--store routing intact:
grep -q -- '--search|-s) return 0' completions/skilldozer.bash && echo "bash: --search/--store routing OK"
# Expected: every check prints OK.
```

### Level 3: Embed sync + whole-module regression (the §14.6 byte-identity lock)

```bash
cd /home/dustin/projects/skilldozer

# go test REBUILDS (re-embeds the edited files), so TestEmbeddedCompletionsMatchOnDisk stays green:
go test ./... ; echo "test exit $?"   # 0

# Specifically the embed-identity test (the §14.6 lock):
go test -run TestEmbeddedCompletionsMatchOnDisk -v ./...
# Expected: PASS (embedded bytes == on-disk bytes; go test rebuilt with the new files).

# No Go regression (nothing in Go land changed, but prove it):
go build ./... ; echo "build exit $?"   # 0
go vet ./...   ; echo "vet exit $?"     # 0
go.mod/go.sum unchanged:
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
# Expected: test/build/vet exit 0; deps unchanged.
```

### Level 4: Rebuild-and-emit + first-arg-offer behavior (the §14.6 delivery path)

```bash
cd /home/dustin/projects/skilldozer

# A PRE-BUILT binary does NOT reflect the edit — REBUILD, then confirm the emit path includes completion:
go build -o /tmp/sd-completion .
echo "=== skilldozer completion --shell bash now offers 'completion' as a candidate ==="
/tmp/sd-completion completion --shell bash 2>/dev/null | grep -q 'completion' && echo "bash emit reflects edit OK"
echo "=== skilldozer completion --shell fish now offers 'completion' ==="
/tmp/sd-completion completion --shell fish 2>/dev/null | grep -q 'completion' && echo "fish emit reflects edit OK"
echo "=== skilldozer completion --shell zsh now offers 'completion' ==="
/tmp/sd-completion completion --shell zsh 2>/dev/null | grep -q 'completion' && echo "zsh emit reflects edit OK"

# (Optional, requires a live shell) confirm the first-arg offer + suppression actually fire. These are
# interactive; skip if no tty, but they are the real user-facing behavior:
#   bash: source completions/skilldozer.bash; skilldozer <TAB>   # offers check init completion (+tags)
#         skilldozer completion <TAB>                            # offers NOTHING (suppression walk)
#   zsh:  autoload; skilldozer <TAB>                             # offers check init completion
#   fish: source; skilldozer <TAB>                               # offers check init completion
rm -f /tmp/sd-completion
# Expected: all three emit checks print OK (the edited bytes flow through skilldozer completion).
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `bash -n`, `zsh -n`, `fish --no-execute` all exit 0 on the edited files
- [ ] Level 2 PASS — all `completion`-presence + extension greps print OK; all do-NOT regressions (tag probe, no `--shell` in matrix, `--search`/`--store` routing) print OK
- [ ] Level 3 PASS — `go test ./...` green (incl. `TestEmbeddedCompletionsMatchOnDisk`); `go build/vet` exit 0; `go.mod`/`go.sum` unchanged
- [ ] Level 4 PASS — a rebuilt binary's `skilldozer completion --shell {bash,zsh,fish}` output contains `completion`

### Feature Validation
- [ ] All three files offer `completion` as a first-positional subcommand
- [ ] All three files suppress tag completion once `completion` (or `check`/`init`) is seen
- [ ] `completion` is gated identically to `check`/`init` in every file (the three-part pattern)
- [ ] The fish `-d` description is `'Emit the shell completion script for eval'` (matches USAGE main.go:107)

### Code Quality / Convention Validation
- [ ] Mirrors the in-file `check`/`init` pattern verbatim (no new completion idiom introduced)
- [ ] `--shell` NOT added to any flag matrix (the §6.1/§6.2-only boundary honored)
- [ ] Tag probe + `--search`/`--store` routing byte-for-byte unchanged
- [ ] Seams located by the anchor text in research §2 (not line numbers)

### Scope Discipline
- [ ] Only the three `completions/*` files modified; `git status --short` shows exactly those three
- [ ] main.go / main_test.go / internal/* UNCHANGED (P1.M2's scope)
- [ ] README.md NOT touched (P1.M3.T1.S2); PRD.md / tasks.json / prd_snapshot.md NOT touched (read-only)

---

## Anti-Patterns to Avoid

- ❌ **Don't add `--shell` to any flag matrix.** It is completion-context-only; the matrix is §6.1/§6.2 global flags. (`skilldozer completion --<TAB>` won't offer it — that's the deliberate decision.) (GOTCHA #1.)
- ❌ **Don't touch the tag probe or the `--search`/`--store` routing.** Only the exclusive-subcommand surface changes. `skilldozer --relative --all 2>/dev/null` stays byte-identical. (GOTCHA #3/#4.)
- ❌ **Don't paraphrase the fish `-d` text.** It must be `'Emit the shell completion script for eval'` (USAGE row main.go:107). (GOTCHA #5.)
- ❌ **Don't skip the rebuild for the emit check.** A pre-built binary won't reflect the on-disk edit (§14.6); `go build` then `skilldozer completion --shell bash | grep completion`. (GOTCHA #6.)
- ❌ **Don't edit main.go, README.md, or PRD.md.** Those are sibling/read-only scopes. (GOTCHA #9.) PRD §14.1 ("check and init") stays as-is — updating it is a human action item.
- ❌ **Don't special-case `completion`.** It is exclusive exactly like `check`/`init`; mirror their in-file handling with zero deviation. (GOTCHA #2.)
- ❌ **Don't break shell quoting.** Keep `"${words[i]}"` (bash) and `${words[(I)completion]}` (zsh) in their exact forms; `bash -n`/`zsh -n` catch a slip. (GOTCHA #8.)

---

## Confidence Score

**9.5/10** — one-pass implementation success likelihood. The change is six text-anchored substitutions mirroring an in-file pattern (check/init), with every anchor cross-checked against the live files and the authoritative `completions_change_map.md`. `completion`'s exclusive status — the fact that justifies the suppression-walk mirror — is verified in main.go (166/297/534). The fish `-d` text is fixed and consistency-checked against the USAGE row. The `--shell` scope boundary (the one tempting slip) is pinned to the contract + §14.1/§14.2 and mechanized into a Level 2 regression grep. All three shell syntax-checkers are confirmed installed with the current files passing (so a syntax slip fails loudly). The embed-sync test (`TestEmbeddedCompletionsMatchOnDisk`) was read and its rebuild-re-embed behavior confirmed (keeps `go test` green). The file set is fully disjoint from the parallel P1.M2.T2.S2 (main.go/main_test.go only). The 0.5 reservation is for the two slips the PRP cannot fully mechanize away — adding `--shell` "while I'm here" (caught by the Level 2 `! grep -- '--shell'` check) and a shell-quoting slip in the bash `[[ ]]` or zsh subscript (caught by `bash -n`/`zsh -n`) — both of which the validation gates catch immediately.
