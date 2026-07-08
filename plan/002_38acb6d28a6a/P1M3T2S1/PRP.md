name: "P1.M3.T2.S1 — Add init subcommand + --store flag to bash/zsh/fish completions"
description: |

---

## Goal

**Feature Goal**: Close the completion drift flagged in `architecture/docs_and_assets_drift.md §4` by adding the `init` subcommand and the `--store <dir>` flag to all three shell completion files (`completions/skilldozer.bash`, `completions/_skilldozer`, `completions/skilldozer.fish`), keeping them LOCKSTEP with `main.go parseArgs()`. `init` gets the same exclusive-first-positional treatment as `check`; `--store` becomes a value-taking flag that completes directory paths.

**Deliverable**: Three edited completion files (no new files, no Go changes). After the edit: `skilldozer <TAB>` offers `init` alongside `check` (and tags) in all three shells; `skilldozer init --<TAB>` offers `--store`; `--store <TAB>` completes directory paths; tag completion and `check` behavior are byte-for-byte unchanged.

**Success Definition**: (1) `init` and `--store` are present in all three files; (2) `check` and every existing flag are still present (no regression); (3) `init` is treated as exclusive (offered only as the first positional; tags suppressed once seen); (4) `--store` completes directories; (5) all three files pass `bash -n` / `zsh -n` / `fish -n` and ShellCheck introduces no NEW warning classes; (6) the programmatic bash harness produces the intended `COMPREPLY` set (see Level 3).

---

## User Persona (if applicable)

**Target User**: Anyone who sources the shipped completions to get tab-completion for `skilldozer` in bash/zsh/fish.

**Use Case**: A new user runs `skilldozer init` (PRD §8.2 first-run setup). They type `skilldozer <TAB>` expecting to discover `init`, then `skilldozer init --<TAB>` expecting to discover `--store <dir>`.

**User Journey**: `skilldozer <TAB>` → sees `init` (and `check`, and their tags) → picks `init` → `--<TAB>` → sees `--store` → `--store <TAB>` → shell completes a directory path.

**Pain Points Addressed**: Before this fix, `init` was invisible in tab-completion (the only way to learn it existed was `--help` or the README), and `--store` could not be discovered or path-completed at all.

---

## Why

- **Completes the `init`/`--store` surface from P1.M2.T1.S1.** `main.go parseArgs()` already parses `init` (reserved positional subcommand) and `--store <dir>` (value-taking, implies `init`, no short form). The completions are the last non-core surface that does not reflect this. (drift doc §4b/§4c, verdict ❌.)
- **Satisfies PRD §14 (h2.13):** "They complete: Subcommands/flags after `skilldozer`." `init` is a first-class subcommand (PRD §6.1, h3.1); `init --store <dir>` is documented (PRD §8.2, h3.9). The completion files must offer both.
- **Honors the LOCKSTEP contract.** Each completion file carries a "LOCKSTEP to main.go parseArgs()" comment: the flag matrix must match `main.go` exactly. `--store` is currently absent from all three; adding it restores parity (drift doc §4c).
- **[Mode A docs-with-work]:** the completion files ARE the user-facing surface for tab-completion. Updating them here satisfies the doc-with-work rule. The README already documents sourcing the completions (README §"Shell completions", lines 70-103) — no separate doc subtask, no README change.

---

## What

### Success Criteria

- [ ] All three files contain the token `init` as a completable subcommand offered ONLY as the first positional (bash cands line; zsh `first` state compadd; fish `__fish_is_first_arg` directive).
- [ ] All three files treat `init` as EXCLUSIVE: once `init` is seen, tags are suppressed (bash walk-loop return 0; zsh `rest`-state guard; fish `__fish_seen_subcommand_from check init`).
- [ ] All three files list `--store` in the flag matrix (bash `compgen -W`; zsh `flags` array; fish `complete -l store`).
- [ ] `--store` completes DIRECTORY PATHS in all three (bash `compgen -d`; zsh `:_files`; fish `-r`).
- [ ] `check` is still present and still exclusive in all three (no regression).
- [ ] All pre-existing flags remain: `--version/-v --help/-h --path/-p --list/-l --all/-a --file/-f --relative --no-color --search/-s`.
- [ ] Tag source `skilldozer --relative --all` is UNCHANGED in all three (drift doc §4d: defensible deviation, out of scope).
- [ ] No `skpp` residue introduced (`grep -rn skpp completions/` stays clean).
- [ ] `bash -n`, `zsh -n`, `fish -n` all pass; `shellcheck -x` introduces no NEW warning class (SC2207/SC2148 are the pre-existing baseline — see research §6).
- [ ] `git status` shows exactly three changed files under `completions/` and nothing else.

---

## All Needed Context

### Context Completeness Check

**Pass.** All three target files were read in full (65 / 55 / 42 lines). The LOCKSTEP source (`main.go parseArgs()` + `usageText`) was read in full and the exact flag set, `init` exclusivity rules, and the canonical `init`/`--store` description strings were extracted. The drift audit (§4) was read. The exact per-line edit anchors were captured (research §4). A PROGRAMMATIC bash-completion harness was executed to PROVE the current (drift) and intended behavior (research §3). The shell quirks (`compgen -d`, zsh `:_files`, fish `-r` inverse-of-`--search` rationale) were verified against the installed bash 5.3 / zsh 5.9 / fish 4.7 and the existing file comments. An implementer who has never seen this repo can complete it in one pass: it is three small shell files, a fixed external source of truth (main.go), and a deterministic validation harness.

### Documentation & References

```yaml
# MUST READ — the verified facts (edit anchors, LOCKSTEP set, proven behavior, quirks)
- file: plan/002_38acb6d28a6a/P1M3T2S1/research/verified_facts.md
  why: "§2 = the exact main.go parseArgs() flag set (the LOCKSTEP contract) + init exclusivity
        + the canonical init/--store description strings (main.go usageText lines 84-85).
        §3 = current-vs-intended COMPREPLY, PROVEN by the programmatic bash harness.
        §4 = the EXACT per-line edit anchors for all three files (current -> target).
        §5 = per-shell gotchas (bash compgen -d / SC2207; zsh :_files + shellcheck false-positives;
        fish -r inverse-of-search rationale). §6 = available validators (bash/zsh/fish -n, shellcheck).
        §7 = scope boundary (do NOT touch main.go, SKILL.md, README, or the tag source)."
  critical: "§4 is the authoritative edit table — the line numbers and current/target strings.
             §2's description strings MUST be used verbatim as the completion -d/[...] text
             (LOCKSTEP to main.go usageText). §5 fish gotcha: --store uses -r (the OPPOSITE of
             --search's deliberate no-r); a comment must explain why, or a future reader will
             'fix' it and break dir completion."

# MUST READ — the three files under edit (the ONLY edit targets)
- file: completions/skilldozer.bash
  why: "bash completion. 65 lines. 4 edits (research §4 table B1-B4): (B1) prev-case add a
        --store arm with `compgen -d`; (B2) append --store to the compgen -W word list (line 37);
        (B3) walk-loop condition add `|| == init` + comment; (B4) cands string add `init` (line 59)."
  pattern: "function _skilldozer_completion: (1) `case $prev` to special-case value-taking flags;
            (2) `if [[ $cur == -* ]]` flag-matrix via compgen -W; (3) walk-loop to detect exclusive
            subcommands + track have_pos; (4) tags via `skilldozer --relative --all`; (5) cands
            adds check (and now init) only when have_pos==0. Registered: `complete -F _skilldozer_completion skilldozer`."
  gotcha: "The flag branch (2) runs BEFORE the walk loop (3), so `init --<TAB>` correctly hits the
           flag matrix regardless of the walk-loop suppression. The walk-loop `return 0` only
           suppresses POSITIONAL candidates, not flags. SC2207 (word-splitting) is pre-existing on
           lines 36 & 62 and is acknowledged in-code; the new compgen -d line may reuse the same style."

- file: completions/_skilldozer
  why: "zsh completion. 55 lines. 3 edits (research §4 table Z1-Z3): (Z1) flags array add
        `'--store[Non-interactive store path for init]:directory:_files'`; (Z2) `first` state
        compadd add `init`; (Z3) `rest`-state guard add `|| ${words[(I)init]}` + update _message."
  pattern: "`#compdef skilldozer`; builds a `flags` array (long+short, with `[desc]` and `:value-fn:`
            markers); `_arguments -C \"$flags[@]\" '1: :->first' '*: :->rest'`; then a `case $state`
            that compadds tags+check in `first` and suppresses tags in `rest` when check is seen."
  gotcha: "`:query:` (used by --search) is a bare placeholder = NO completion for the value (correct:
           free-text). `:directory:_files` (for --store) routes the value slot to file/path completion.
           ShellCheck is bash-only and FALSE-POSITIVES on this zsh file (SC2296/SC1087/SC2128/SC2154/
           SC2206) — do NOT 'fix' them; the real check is `zsh -n` (passes)."

- file: completions/skilldozer.fish
  why: "fish completion. 42 lines. 3 edits (research §4 table F1-F3): (F1) add `complete -c skilldozer
        -l store -d 'Non-interactive store path for init' -r`; (F2) add `complete -c skilldozer -n
        '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'`;
        (F3) tags-suppression condition `__fish_seen_subcommand_from check` -> `check init`."
  pattern: "`complete -c skilldozer -f` (global no_files); per-flag `complete -c skilldozer -s X -l long -d 'desc'`;
            check offered via `__fish_is_first_arg`; tags via ONE dynamic directive with command
            substitution `(skilldozer --relative --all 2>/dev/null)`, suppressed via
            `__fish_seen_subcommand_from` + `__fish_prev_arg_in`."
  gotcha: "fish 4.x `-r` switches into 'complete the option's value' mode and BYPASSES the global `-f`
           (no_files), offering file/dir paths. The file DELIBERATELY omits `-r` on --search (free-text
           -> offer nothing). For --store we WANT path completion -> `-r` is correct (the inverse
           decision). A comment MUST explain this so a future reader doesn't 'reconcile' the two."

# READ-ONLY — the LOCKSTEP source of truth (the flag set the completions must MATCH)
- file: main.go
  why: "parseArgs() (lines ~155-300) is the authoritative flag set; exclusivityError() (~line 470)
        defines init as exclusive (rejects init+tags, init+check, init+list/search/all/path); usageText
        (lines 60-89) holds the canonical init/--store description strings to reuse as completion -d text."
  critical: "main.go ALREADY supports init/--store (P1.M2 complete). DO NOT EDIT main.go — the
             completions must come INTO MATCH with it, not vice-versa. Re-read parseArgs() if unsure
             whether a flag is value-taking or has a short form (e.g. --store has NO short form)."

# READ-ONLY — the drift audit (the contract for this work item)
- file: plan/002_38acb6d28a6a/architecture/docs_and_assets_drift.md
  why: "§4 = the per-file drift verdict + the exact contract for adding init (alongside check, same
        gating) and --store (value-taking, dir completion). §4d documents the --relative--all vs --all
        judgment call (leave unchanged). Cross-cutting note #3: keep the three files LOCKSTEP."
  section: "§4b (init missing), §4c (--store missing), §4d (tag source: defensible, out of scope)."

# READ-ONLY — the PRD sections selected for this item
- file: PRD.md
  why: "§14 (h2.13) = completions must complete subcommands/flags. §6.1 (h3.1) = the init subcommand
        row. §8.2 (h3.9) = `init`/`init <dir>`/`init --store <dir>` first-run setup. §6.3 (h3.3) =
        mutual exclusivity (init is exclusive, like check)."
  section: "h2.13 (§14), h3.1 (§6.1 init row), h3.2 (§6.2 modifiers), h3.9 (§8.2 init)."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer
$ ls completions/
_skilldozer          # zsh autoload (#compdef skilldozer)        — EDIT (3 changes)
skilldozer.bash      # bash completion (complete -F)             — EDIT (4 changes)
skilldozer.fish      # fish completion (sourced)                 — EDIT (3 changes)
$ grep -c skpp completions/*    # → all 0 (clean; rename already landed here)
$ bash -n completions/skilldozer.bash && echo OK   # → OK (baseline syntax valid)
$ zsh  -n completions/_skilldozer      && echo OK   # → OK
$ fish -n completions/skilldozer.fish  && echo OK   # → OK
# main.go parseArgs() ALREADY supports init + --store (P1.M2 complete) — do NOT touch.
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
completions/skilldozer.bash   # ADD init subcommand + --store flag (dir completion) to the bash matrix
completions/_skilldozer       # ADD init subcommand + --store flag (path completion) to the zsh matrix
completions/skilldozer.fish   # ADD init subcommand + --store flag (-r value) to the fish matrix
```

**No new files. No Go files touched.** Three existing files edited in place.

| File | Responsibility |
|---|---|
| `completions/skilldozer.bash` | bash tab-completion for `skilldozer`. After edit: offers `init`+`check` as first-positionals; `--store` in the flag matrix with `compgen -d` dir completion; tags suppressed once `init`/`check` seen. |
| `completions/_skilldozer` | zsh tab-completion (`#compdef skilldozer`). After edit: `init` in the `first` compadd; `--store` in the `flags` array with `:_files` path completion; tags suppressed once `init`/`check` seen. |
| `completions/skilldozer.fish` | fish tab-completion. After edit: `init` via `__fish_is_first_arg`; `--store ... -r` (value-taking, path completion); tags suppressed once `init`/`check` seen. |

### Known Gotchas of our codebase & Library Quirks

```bash
# CRITICAL (LOCKSTEP): the flag matrix in EACH file is hand-frozen to main.go parseArgs().
# There is no shared source of truth. When adding --store, add it to ALL THREE files in the
# SAME edit session so they cannot drift apart. Each file's header already says so.

# bash: the flag-completion branch `if [[ "$cur" == -* ]]` runs BEFORE the walk loop, so
# `skilldozer init --<TAB>` correctly offers the flag matrix (incl. --store) even though the
# walk loop later returns 0 on seeing `init`. The walk-loop suppression only kills POSITIONAL
# candidates (tags / re-offering subcommands), never flags.

# bash: `compgen -d` word-splits into COMPREPLY (reuses the file's existing SC2207-acceptable
# style). Directory names CAN contain spaces; the robust form is
#   mapfile -t COMPREPLY < <(compgen -d -- "$cur")
# Either is acceptable; the simple form matches the file's prevailing convention. ShellCheck
# baseline already shows 2x SC2207 (lines 36, 62) — do not introduce a NEW warning CLASS.

# zsh: `:query:` (used by --search) is a bare placeholder = no completion for the value.
# `:directory:_files` (for --store) routes the value slot to file/path completion. Use `_files`
# (files+dirs); `_path_files -/` is the dirs-only alternative.

# zsh: ShellCheck is BASH-ONLY and false-positives on this zsh file (SC2296/SC1087/SC2128/
# SC2154/SC2206). Do NOT 'fix' them — they would break the zsh code. The real check is `zsh -n`.

# fish: `-r` makes the option require an argument AND switches fish 4.x into "complete the
# option's value" mode, bypassing the global `complete -c skilldozer -f` (no_files) and offering
# file/dir paths. --search DELIBERATELY omits -r (free-text -> offer nothing). --store WANTS
# path completion -> -r is correct (the OPPOSITE decision). A comment MUST say why, or a future
# reader will "reconcile" the two and break one of them.

# SCOPE: do NOT add bare-positional `init <dir>` directory completion. The contract scopes
# dir-completion to --store only; `init` is treated like `check` (suppress further positionals).
# The discoverable route for directory completion is `init --store <dir>`.
```

---

## Implementation Blueprint

### Data models and structure

None. These are shell completion scripts (bash function + zsh `#compdef` + fish `complete` directives). No Go types, no schemas, no migrations. The only "model" is the flag/subcommand set, which is defined externally by `main.go parseArgs()` (research §2).

### Implementation Tasks (ordered by dependencies)

The three files are independent — edit them in any order. Within each file, apply ALL its edits together (they are tiny and adjacent). The exact current→target text per line is in `research/verified_facts.md §4`; the strings below are the authoritative targets.

```yaml
Task 1: EDIT completions/skilldozer.bash  (4 changes — research §4 table B1-B4)
  - B1 (prev-case, ~lines 30-32): AFTER the existing `--search|-s) return 0 ;;` arm, add:
        --store) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
    and extend the preceding comment to note --store WANTS path completion (unlike --search).
  - B2 (flag matrix, ~line 37): append ` --store` to the compgen -W word list, i.e.
        "--version -v --help -h --path -p --list -l --all -a --file -f --relative --no-color --search -s --store"
  - B3 (walk loop, ~lines 42-49): change the condition
        [[ "${words[i]}" == "check" ]] && return 0
    to
        [[ "${words[i]}" == "check" || "${words[i]}" == "init" ]] && return 0
    and update the comment to say `check` AND `init` are exclusive (PRD §6.3).
  - B4 (cands, ~line 59):
        (( have_pos == 0 )) && cands="$cands check init"
  - PRESERVE: the `_init_completion` fallback block; the `--search|-s` arm; the tags fetch;
    the SC2207 comment; the `complete -F _skilldozer_completion skilldozer` registration.
  - NAMING/PLACEMENT: edits in place; no new functions, no new files.
  - DO NOT touch: main.go, the other two completion files, README, SKILL.md.

Task 2: EDIT completions/_skilldozer  (3 changes — research §4 table Z1-Z3)
  - Z1 (flags array, after the `-s[...]` entry ~line 33): add the entry
        '--store[Non-interactive store path for init]:directory:_files'
    (with a one-line comment that :_files routes the value slot to path completion; no short form).
  - Z2 (first state, ~line 41):
        compadd -- "$tags[@]" check init
  - Z3 (rest state, ~lines 44-48): change
        if (( ${words[(I)check]} )); then
            _message 'check takes no further arguments'
    to
        if (( ${words[(I)check]} || ${words[(I)init]} )); then
            _message 'subcommand takes no tag arguments'
    and update the comment to name check AND init.
  - PRESERVE: the `#compdef skilldozer` line; the tags fetch `${(f)"$(...)"}`; the
    `_arguments -C "$flags[@]" '1: :->first' '*: :->rest'` line; the `:query:` markers on
    --search/-s.
  - DO NOT "fix" ShellCheck output on this file (bash-only linter false-positives on zsh).

Task 3: EDIT completions/skilldozer.fish  (3 changes — research §4 table F1-F3)
  - F1 (flag matrix, AFTER the existing -s/-l search directive ~line 32): add
        complete -c skilldozer -l store -d 'Non-interactive store path for init' -r
    preceded by a comment explaining that -r HERE is intentional (the inverse of --search's
    deliberate no-r): --store's value is a directory, so we WANT fish to complete paths for it.
  - F2 (subcommands, AFTER the existing check directive ~line 35): add
        complete -c skilldozer -n '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'
    (with a one-line comment that init is an exclusive first-positional, like check).
  - F3 (tags suppression, ~lines 39-41): change
        -n 'not __fish_seen_subcommand_from check; and not __fish_prev_arg_in --search -s'
    to
        -n 'not __fish_seen_subcommand_from check init; and not __fish_prev_arg_in --search -s'
    and update the comment to name check OR init.
  - PRESERVE: the global `complete -c skilldozer -f`; all existing flag directives; the check
    directive; the dynamic tags directive and its command substitution.
  - DO NOT add -r to --search (it would break free-text behavior).

Task 4: VERIFY  (run after Tasks 1-3 — see Validation Loop)
  - Structural grep (Level 2): assert init + --store present, check still present, no skpp,
    flag matrix complete, in ALL THREE files.
  - Syntax (Level 1): bash -n / zsh -n / fish -n pass; shellcheck introduces no NEW warning class.
  - Behavior (Level 3): the programmatic bash harness produces the intended COMPREPLY set.
```

### Implementation Patterns & Key Details

```bash
# ---- bash: the prev-case special-casing value-taking flags ----
# --search/-s  -> free-text query  -> offer NOTHING (return 0 with empty COMPREPLY)
# --store      -> directory value  -> complete DIRECTORIES via compgen -d
case "$prev" in
    --search|-s) return 0 ;;
    --store) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
esac

# ---- bash: the walk loop detects EXCLUSIVE subcommands (PRD §6.3) ----
# check AND init: once seen as a non-flag token, offer no further POSITIONAL candidate.
# (Flag completion above still fires for `init --<TAB>` — it is gated on `cur == -*`,
#  which runs before this loop.)
local i have_pos=0
for ((i=1; i<cword; i++)); do
    [[ "${words[i]}" == "check" || "${words[i]}" == "init" ]] && return 0
    [[ "${words[i]}" == -* ]] && continue
    have_pos=1
done
# ...
(( have_pos == 0 )) && cands="$cands check init"

# ---- zsh: the value-slot marker decides whether the value gets completion ----
# `:query:`     -> bare placeholder, NO completion (free-text)   [--search/-s]
# `:directory:_files` -> routes value to file/path completion    [--store, NEW]
local -a flags=(
    # ... existing entries ...
    '--search[Substring search over tag/name/description/keywords]:query:'
    '-s[Substring search over tag/name/description/keywords]:query:'
    '--store[Non-interactive store path for init]:directory:_files'
)

# ---- zsh: first/rest states ----
first) compadd -- "$tags[@]" check init ;;        # offer init + check only as first positional
rest)  if (( ${words[(I)check]} || ${words[(I)init]} )); then
           _message 'subcommand takes no tag arguments'   # exclusive: suppress tags
       else
           compadd -- "$tags[@]"
       fi ;;

# ---- fish: -r is the OPTION-VALUE completion switch (inverse of --search's no-r) ----
complete -c skilldozer -l store -d 'Non-interactive store path for init' -r   # -r: complete paths
complete -c skilldozer -n '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'
# tags suppressed once EITHER exclusive subcommand is seen:
complete -c skilldozer -n 'not __fish_seen_subcommand_from check init; and not __fish_prev_arg_in --search -s' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
```

### Integration Points

```yaml
NO CODE CHANGES:
  - main.go: untouched (parseArgs() ALREADY supports init/--store — the LOCKSTEP source).
  - internal/*: untouched.
  - go.mod/go.sum: untouched.
  - No migrations, no routes, no env vars, no new files.

ASSET (completion files only):
  - file: "completions/skilldozer.bash"
    change: "+init subcommand (cands + walk-loop suppression), +--store flag (matrix + compgen -d value rule)"
  - file: "completions/_skilldozer"
    change: "+init subcommand (first/rest states), +--store flag (flags array with :_files)"
  - file: "completions/skilldozer.fish"
    change: "+init subcommand (__fish_is_first_arg directive), +--store flag (-r value), tags suppression -> check init"

DOCS (Mode A — docs-with-work, no separate subtask):
  - README.md: untouched. Its "Shell completions" section (lines 70-103) already documents
    sourcing the three files; the contract says "do not duplicate." README's own init/config
    doc gap is owned by P1.M4.T2.S1, not this item.
```

---

## Validation Loop

> There is no Go test or CI for completions (`grep -rln "compgen|compadd|__fish" --include="*.go"` → none).
> Validation is: syntax checks (Level 1) + structural grep (Level 2) + a programmatic bash-completion
> harness + per-shell smoke (Level 3). All commands were proven to run on this host (bash 5.3,
> zsh 5.9, fish 4.7, shellcheck installed).

### Level 1: Syntax & Style (immediate, after editing each file)

```bash
cd /home/dustin/projects/skilldozer

# Syntax must be valid in each shell's own parser:
bash -n completions/skilldozer.bash && echo "bash -n OK"
zsh  -n completions/_skilldozer      && echo "zsh -n OK"
fish -n completions/skilldozer.fish  && echo "fish -n OK"
# Expected: all three print OK. If any fails, READ the parser error and fix BEFORE proceeding.

# ShellCheck (bash + fish). NOTE: shellcheck is bash-only and FALSE-POSITIVES on the zsh
# file (_skilldozer) — do NOT run it there; `zsh -n` above is the real zsh check.
shellcheck -x completions/skilldozer.bash
shellcheck -x completions/skilldozer.fish
# Expected for bash: 3x SC2207 (the 2 pre-existing on lines 36 & 62, +1 on the new --store
#   compgen -d line) — SC2207 is the file's acknowledged style; NO NEW warning class.
# Expected for fish: SC2148 (shebang tip — harmless for a sourced completion file) and nothing new.
# A NEW warning CLASS (e.g. SC2086, SC2046) means a real bug — fix it.

# Confirm no Go file changed (these completions are shell-only):
git diff --name-only | grep -E '\.go$' && echo "FAIL: Go file changed" || echo "OK: no .go changes"
go build ./... && echo "build OK"   # sanity (no .go touched -> still green)
```

### Level 2: Structural grep (the contract's smoke check — assert presence + no regression)

```bash
cd /home/dustin/projects/skilldozer
ok=1
check() { if grep -qE "$2" "$1"; then echo "  OK  $1 :: $3"; else echo "  FAIL $1 :: $3"; ok=0; fi; }

for f in completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish; do
  echo "== $f =="
  check "$f" '\binit\b'           "init subcommand present"
  check "$f" '\-\-store\b'        "--store flag present"
  check "$f" '\bcheck\b'          "check STILL present (no regression)"
  # every pre-existing flag still there:
  check "$f" '\-\-version'   ; check "$f" '\-v\b'
  check "$f" '\-\-help'      ; check "$f" '\-h\b'
  check "$f" '\-\-path'      ; check "$f" '\-p\b'
  check "$f" '\-\-list'      ; check "$f" '\-l\b'
  check "$f" '\-\-all'       ; check "$f" '\-a\b'
  check "$f" '\-\-file'      ; check "$f" '\-f\b'
  check "$f" '\-\-relative'  ; check "$f" '\-\-no-color'
  check "$f" '\-\-search'    ; check "$f" '\-s\b'
  # tag source unchanged (drift doc §4d: defensible deviation):
  check "$f" 'skilldozer --relative --all'
done

# bash-specific: init in cands + walk-loop; --store in matrix + value rule
grep -q 'cands="\$cands check init"'              completions/skilldozer.bash && echo "bash cands OK"      || { echo FAIL; ok=0; }
grep -q '== "check" || "\${words\[i\]}" == "init"' completions/skilldozer.bash && echo "bash suppress OK"  || { echo FAIL; ok=0; }
grep -q -- '--search -s --store'                  completions/skilldozer.bash && echo "bash matrix OK"     || { echo FAIL; ok=0; }
grep -q -- 'compgen -d -- "\$cur"'                completions/skilldozer.bash && echo "bash --store value OK" || { echo FAIL; ok=0; }
# zsh-specific: init in first/rest; --store :_files
grep -q 'compadd -- "\$tags\[@\]" check init'     completions/_skilldozer && echo "zsh first OK"      || { echo FAIL; ok=0; }
grep -q 'words\[(I)check\]} || \${words\[(I)init\]}' completions/_skilldozer && echo "zsh rest OK"   || { echo FAIL; ok=0; }
grep -q -- "--store\[Non-interactive store path for init\]:directory:_files" completions/_skilldozer && echo "zsh --store OK" || { echo FAIL; ok=0; }
# fish-specific: init directive; --store -r; tags suppression check init
grep -q "__fish_is_first_arg' -a 'init'"          completions/skilldozer.fish && echo "fish init OK"      || { echo FAIL; ok=0; }
grep -q -- "-l store -d 'Non-interactive store path for init' -r" completions/skilldozer.fish && echo "fish --store OK" || { echo FAIL; ok=0; }
grep -q '__fish_seen_subcommand_from check init'  completions/skilldozer.fish && echo "fish suppress OK" || { echo FAIL; ok=0; }

# No skpp residue introduced anywhere:
grep -rIn "skpp" completions/ && { echo "FAIL: skpp introduced"; ok=0; } || echo "no skpp (clean)"

[ "$ok" = 1 ] && echo "ALL STRUCTURAL CHECKS PASS" || echo "SOME CHECKS FAILED — review above"
# Expected: ALL STRUCTURAL CHECKS PASS.
```

### Level 3: Behavioral validation (programmatic bash harness + per-shell smoke)

```bash
cd /home/dustin/projects/skilldozer
export SKILLDOZER_SKILLS_DIR="$PWD/skills"   # so tags resolve without a configured store

# (a) PROGRAMMATIC bash-completion harness — deterministic, NO interactive TAB key.
# Forces the _init_completion fallback, sets COMP_WORDS/CWORD, calls the function,
# inspects COMPREPLY. This is the strongest automated check available.
source completions/skilldozer.bash 2>/dev/null
unset -f _init_completion 2>/dev/null            # force the manual fallback branch
_bashcomp() {                                    # $1=cword, rest=words
  local cw=$1; shift
  COMP_WORDS=("$@"); COMP_CWORD=$cw; COMPREPLY=()
  _skilldozer_completion
  printf '  -> %s\n' "${COMPREPLY[@]}"
}
echo "skilldozer <TAB>           (expect: example, check, init):"
_bashcomp 1 skilldozer ""
echo "skilldozer init <TAB>      (expect: NOTHING — init is exclusive):"
_bashcomp 2 skilldozer init ""
echo "skilldozer init --<TAB>    (expect: --store among the flags):"
_bashcomp 3 skilldozer init "--"
echo "skilldozer init --store <TAB>  (expect: directory names from cwd):"
_bashcomp 4 skilldozer init --store ""
echo "skilldozer check <TAB>     (expect: NOTHING — check exclusive; no regression):"
_bashcomp 2 skilldozer check ""
echo "skilldozer --store <TAB>   (expect: directories — --store works at top level too):"
_bashcomp 2 skilldozer --store ""
# Expected: init appears in the first set; nothing after `init`/`check`; `--store` in the
#   `init --` set; directories after `--store ` (and top-level `--store `).

# (b) zsh smoke (structural is the primary check; this confirms the autoload parses):
zsh -c 'autoload -U compinit; compinit -u 2>/dev/null; source completions/_skilldozer 2>/dev/null; echo "zsh parses OK"'

# (c) fish smoke (fish -n already validated syntax; confirm it loads):
fish -c 'source completions/skilldozer.fish; echo "fish loads OK"'

# (d) Optional REAL interactive smoke (only if you have the shells configured). Source the
# file in a live shell, type `skilldozer ` and hit TAB, and eyeball that init+check+tags
# appear; then `skilldozer init --` TAB -> --store appears. The programmatic harness (a)
# is the deterministic equivalent and is sufficient on its own.
```

### Level 4: Creative & Domain-Specific Validation

```bash
cd /home/dustin/projects/skilldozer
export SKILLDOZER_SKILLS_DIR="$PWD/skills"

# LOCKSTEP cross-check: confirm the completion flag set EQUALS main.go parseArgs() exactly.
# (No new/missing flag; --store added, no flag dropped.) This is the drift doc's core invariant.
echo "=== flags main.go knows (long forms) ==="
grep -oE '"--[a-z-]+"' main.go | grep -oE -- '--[a-z-]+' | sort -u
echo "=== --store/--search value-taking confirmed in main.go ==="
grep -nE 'case "--store"|case "--search"|c.initStore|c.searchQ' main.go | head

# Confirm the init/--store description strings in completions match main.go usageText
# (LOCKSTEP: main.go USAGE text == completion -d text):
grep -oE "First-run setup: pick/create the skills store and write the config" main.go completions/*.fish completions/_skilldozer
grep -oE "Non-interactive store path for init" main.go completions/*.fish completions/_skilldozer completions/skilldozer.bash
# Expected: each canonical string appears in main.go AND in the relevant completion files.
```

---

## Final Validation Checklist

### Technical Validation

- [ ] `bash -n completions/skilldozer.bash`, `zsh -n completions/_skilldozer`, `fish -n completions/skilldozer.fish` all pass
- [ ] `shellcheck -x` on bash + fish introduces NO new warning class (SC2207/SC2148 are the pre-existing baseline)
- [ ] `go build ./...` green (no `.go` file changed — sanity)
- [ ] `git status` shows exactly three changed files under `completions/`

### Feature Validation

- [ ] `init` present in all three files as a completable first-positional subcommand
- [ ] `--store` present in all three flag matrices and completes DIRECTORIES (`compgen -d` / `:_files` / `-r`)
- [ ] `init` treated as exclusive (tags suppressed once seen) in all three files
- [ ] `check` still present and still exclusive (no regression)
- [ ] All pre-existing flags still present (Level 2 grep block)
- [ ] Programmatic bash harness (Level 3a) shows `init` in the `<TAB>` set, nothing after `init`/`check`, `--store` after `init --`, directories after `--store `
- [ ] Tag source `skilldozer --relative --all` unchanged; no skpp introduced

### Code Quality Validation

- [ ] LOCKSTEP invariant holds: completion flag set == `main.go parseArgs()` flag set (Level 4)
- [ ] init/--store description strings in completions match `main.go usageText` verbatim (Level 4)
- [ ] Each shell's quirks respected (bash compgen -d; zsh `:_files` not `:query:`; fish `-r` on --store only)
- [ ] Comments updated to name BOTH check AND init where the logic now covers both
- [ ] Only `completions/*` modified — `main.go`, `skills/example/SKILL.md`, `README.md`, PRD.md untouched

### Documentation & Deployment

- [ ] [Mode A] The completion files ARE the user-facing tab-completion surface — updated here (no separate doc task)
- [ ] README untouched (its "Shell completions" section already documents sourcing; README's init/config gap is P1.M4.T2.S1)
- [ ] No new env vars, no install.sh change (install.sh does not install completions — README documents the copy step)

---

## Anti-Patterns to Avoid

- ❌ Don't edit `main.go` — parseArgs() ALREADY supports `init`/`--store` (P1.M2 complete). The completions must come INTO MATCH with main.go, not vice-versa. main.go is the LOCKSTEP source.
- ❌ Don't edit `skills/example/SKILL.md` (owned by the parallel P1.M3.T1.S1 PRP) or `README.md` (owned by P1.M4.T2.S1). This item touches ONLY the three completion files.
- ❌ Don't add bare-positional `init <dir>` directory completion — the contract scopes dir-completion to `--store` only. Treat `init` like `check` (suppress further positionals). The discoverable route is `init --store <dir>`.
- ❌ Don't "fix" ShellCheck output on the zsh file (`_skilldozer`) — it is a bash-only linter producing false-positives on zsh syntax (`SC2296`/`SC1087`/`SC2128`/`SC2154`/`SC2206`). The real check is `zsh -n`.
- ❌ Don't add `-r` to fish `--search` "for consistency" — `--search` deliberately omits `-r` (free-text query → offer nothing); `--store` deliberately uses `-r` (directory value → complete paths). They are intentionally opposite. A comment must say so.
- ❌ Don't change the tag source from `skilldozer --relative --all` to literal `--all` — `--relative` is a defensible/better deviation (drift doc §4d); out of scope.
- ❌ Don't let the three files drift — edit all three in one session; each carries a "LOCKSTEP to main.go parseArgs()" header for a reason.
- ❌ Don't drop the `check` handling when extending for `init` — `check` must remain present and exclusive (regression).
- ❌ Don't reuse `:query:` for `--store` in zsh — `:query:` offers NO completion; `--store` needs `:_files` (path completion). The marker decides whether the value slot gets completion.
