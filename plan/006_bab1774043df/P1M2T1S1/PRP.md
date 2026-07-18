# PRP — P1.M2.T1.S1: Multi-position directory completion after `--link` (bash, zsh, fish)

> **Subtask:** P1.M2.T1.S1 — the sole subtask of P1.M2.T1. Updates all three shell completion scripts so that once `--link` appears anywhere before the cursor, EVERY following **positional** completes to a DIRECTORY (not a skill tag) — implementing PRD §14.1 rule 5 + §14.2 + §8.4 (`--link` takes many dirs). Today each shell completes dirs ONLY for the first position after `--link`; after `--link d1`, the next `<tab>` offers skill tags.
>
> **Scope:** THREE existing completion-script files only — `completions/skilldozer.bash`, `completions/_skilldozer`, `completions/skilldozer.fish`. No `.go` source change (the Go multi-target `--link` is P1.M1.T1.S1, parallel, which EXPLICITLY excludes the completion files). No new files. Zero Go/deps changes — `//go:embed` re-reads the edited files on rebuild. The flag list already includes `--link` in all three files (no §14.4 flag-set change); the ONLY change is the multi-position directory behavior.
>
> **STATUS (verified at PRP-write time):** read all three completion files (full), `codebase_state.md` §"Completion Scripts" (full), and `TestEmbeddedCompletionsMatchOnDisk` (full). The parallel sibling P1.M1.T1.S1 edits `main.go`/`main_test.go`/`README.md` and does NOT touch completions (its GOTCHA #10 defers multi-link dir completion to this task) — disjoint files, no collision. The flag list already has `--link`; no flag-set change.

---

## Goal

**Feature Goal**: Make `skilldozer --link d1 <tab>` (and `--link d1 d2 <tab>`, etc.) offer DIRECTORIES at every positional after `--link` — across bash, zsh, and fish — instead of falling through to skill-tag completion. PRD §14.1 rule 5: "when `--link` is present, **every** following positional completes to directories (it links many, §8.4)." All other completion behavior (skills-first bare `<tab>`, long-form-only flags, `--init`/`--store` dir completion, `--search` nothing, `--shell` enum, §14.7 list-ambiguous) is unchanged.

**Deliverable**: Edits to three existing files:
1. `completions/skilldozer.bash` — a scan guard before the `case "$prev"` block: if `--link` is in `words[1..cword-1]` and the current token is a positional (cur not `-*`), offer dirs via `compgen -d` and return.
2. `completions/_skilldozer` (zsh) — branch the `args` state: if `--link` precedes the cursor, call `_files` (dirs) instead of `compadd` (tags).
3. `completions/skilldozer.fish` — add a dir directive (`-n` fires when `--link` has been seen) and tighten the tag directive's `-n` to suppress when `--link` has been seen (mutual exclusion).

**Success Definition**: `bash -n` / `zsh -n` / fish-syntax clean; `go build` succeeds (re-embeds); `go test ./...` green (incl. `TestEmbeddedCompletionsMatchOnDisk`, auto-pass on rebuild); `--link d1 <tab>` → dirs in all three shells; `--link <tab>` → dirs (unchanged); bare `<tab>` → skills (unchanged); `-<tab>` → flags (unchanged).

---

## User Persona (if applicable)

**Target User**: a developer batch-linking several external skill repos via `skilldozer --link ~/proj/a ~/proj/b ~/proj/c` and tab-completing each directory in turn.

**Use Case**: `skilldozer --link ~/proj/<tab>` (first dir) then `<tab>` again for the next dir — today the second `<tab>` offers skill tags, forcing the user to type the path.

**User Journey**: User types `skilldozer --link ` → `<tab>` shows dirs (first position, already works) → picks `d1` → `<tab>` again → (today) tags, confusing → (after fix) dirs, picks `d2`.

**Pain Points Addressed**: the jarring switch from dirs→tags after the first `--link` target; having to type subsequent link paths by hand.

---

## Why

- **PRD §14.1 rule 5 + §14.2 + §8.4** are authoritative: `--link <dir> [<dir>...]` takes one or more dirs; "offer file/dir completion at **every position after `--link`**". Today only the first position completes dirs; the rest fall through to tags — a spec deviation.
- **The Go side already batches** (P1.M1.T1.S1): `--link a b c` links all three. The completion must match — a user who tab-completes the 2nd/3rd target should see dirs, not tags.
- **§14.4 lockstep**: the completion behavior is frozen to `main.go parseArgs()`. Since `--link` now collects trailing positionals as dirs (the Go change), the completion must offer dirs at every trailing positional.
- **`--link d1 <tab>` is the natural extension** of the already-documented `--link <tab>` case (PRD §14.6 behavior-contract table lists both rows).

---

## What

Three surgical, per-shell edits, each preserving all existing behavior (the ONLY change is multi-position dir completion after `--link`):

- **bash**: a scan guard (loop over `words`) before the `case "$prev"` block, gated on the current token being a positional.
- **zsh**: a branch in the `args` state of the `_arguments` `case` (preserves `_arguments`' flag/value routing).
- **fish**: a new dir directive + a tightened tag-directive condition (mutual exclusion via `commandline -opc`).

### Success Criteria

- [ ] bash: `--link d1 <tab>` and `--link d1 d2 <tab>` → dirs; `<tab>` → skills; `-<tab>` → flags; `--store`/`--init`/`--search`/`--shell` routing unchanged.
- [ ] zsh: `--link d1 <tab>` → dirs (`_files`); `<tab>` → skills; flag/value routing unchanged.
- [ ] fish: `--link d1 <tab>` → dirs (`__fish_complete_directories`); `<tab>` → skills; `--store`/`--init`/`--search`/`--shell` unchanged.
- [ ] `--link <tab>` (first position) still → dirs in all three (unchanged).
- [ ] `bash -n` / `zsh -n` / fish syntax clean; `go build` succeeds; `go test ./...` green (TestEmbeddedCompletionsMatchOnDisk auto-passes on rebuild).
- [ ] Only the three completion files change; no `.go`/README/deps change.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit is pinned to exact current lines with before/after (all three files read in full; `codebase_state.md` §"Completion Scripts" transcribed). The per-shell mechanics that make each fix correct (bash `_init_completion` sets `words`/`cword`/`cur`; zsh `_arguments -C ... '*: :->args' && return 0` reaches the `case` ONLY for positionals; fish aggregates ALL matching directives → needs mutual exclusion) are documented and traced against every OUTPUT case. The `cur != -*` gate (bash) and the mutual-exclusion (fish) refinements over the contract's literal wording are justified by PRD §14.1 rule 5 ("POSITIONAL"). The embed-identity/rebuild constraint is confirmed. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (per-shell current+fix, the mechanics, the preservation matrix)
- file: plan/006_bab1774043df/P1M2T1S1/research/verified_facts.md
  why: "§1 the bug + boundary + embed/rebuild. §2 bash (the scan guard; the cur!=-* gate refinement; the trace).
        §3 zsh (the args-case branch; why _arguments reaches the case ONLY for positionals; the ${words[(i)--link]}
        idiom). §4 fish (the mutual-exclusion: new dir directive + tightened tag directive; the commandline -opc
        condition). §5 the full preservation matrix (8 behaviors × 3 shells, all unchanged). §6 disjoint from sibling."
  critical: "§3 (zsh: fix the args CASE, don't add a top-level guard — preserves flag routing) and §4 (fish: MUST
             tighten the tag directive's -n too, or dirs+tags both fire after --link) are the two non-obvious points."

# MUST READ — the authoritative current-state map (exact per-shell code + line numbers)
- file: plan/006_bab1774043df/architecture/codebase_state.md
  why: "§'Completion Scripts' gives the exact current --link handling per shell (bash case line 43, zsh spec line 51,
        fish -r line 39) + the 'Needs fix' note for each. The line numbers match the live files."

# MUST READ — the three files under edit (read each in full before editing)
- file: completions/skilldozer.bash
  why: "THE bash edit target. _init_completion block (sets words/cword/cur in both branches). case \"$prev\" block @41-45.
        Place the scan guard AFTER _init_completion and BEFORE the case. The compgen -W flag list @50 (already has --link;
        NO change). The §14.7 bind block @76-89 (NO change)."
- file: completions/_skilldozer
  why: "THE zsh edit target. _arguments line @62 ('*: :->args'). case \"$state\" @64-67. Edit the args arm to branch on
        ${words[(i)--link]} < CURRENT. The flags array @28-60 (already has --link; NO change). The _skilldozer \"$@\"
        self-call @71 (autoload idiom; NO change)."
- file: completions/skilldozer.fish
  why: "THE fish edit target. The --link -r line @39 (NO change — owns the first value slot). The tag directive @65-69
        (TIGHTEN its -n). Add the new dir directive near @39 or @65. The -f global @17, the flag directives, the
        --shell -x @62 (NO change)."
  pattern: "fish completion = a stack of `complete -c skilldozer` rules; fish aggregates ALL whose -n is true. So a
            NEW positional directive needs the EXISTING one's -n tightened to exclude it, or both fire (dirs+tags)."

# MUST READ — the embed-identity test (rebuild is required)
- file: main_test.go
  why: "TestEmbeddedCompletionsMatchOnDisk @3140 compares completionScript(shell) (//go:embed var) to os.ReadFile(path).
        After editing the on-disk files, `go build`/`go test` re-embeds → embedded==on-disk → PASS. Do NOT edit the test
        or main.go's //go:embed. bash/fish/zsh are all embedded VERBATIM (the zsh zshEvalScript derivation is at emit
        time in runCompletion, not in the embed/test)."
  gotcha: "Run `go build` (or `go test`) AFTER editing so the embed var holds the new bytes — else the test compares
           stale embedded bytes to the new file and FAILS."

# READ-ONLY — the parallel sibling PRP (boundary: edits main.go/main_test.go/README.md ONLY, NOT completions)
- file: plan/006_bab1774043df/P1M1T1S1/PRP.md
  why: "Confirms P1.M1.T1.S1 implements the Go multi-target --link (struct/parser/runLink/tests/README) and its GOTCHA #10
        explicitly excludes the completion files ('this subtask does NOT touch the completion scripts. Multi-link DIRECTORY
        completion is P1.M2.T1.S1'). Disjoint files; land in either order. The --link flag is ALREADY in all three
        completion files (its GOTCHA #10 notes the bash file's --link case), so this task needs NO flag-set change."

# READ-ONLY — PRD (the authority for the multi-position dir contract)
- file: PRD.md
  why: "READ-ONLY. §14.1 rule 5 ('when --link is present, EVERY following positional completes to directories'); §14.2
        (--link <dir> [<dir>...] — offer file/dir completion at every position after --link); §8.4 (batch linking); the
        §14.6 behavior-contract table rows for --link <tab> AND --link d1 <tab>. §14.4 lockstep; §14.7 list-ambiguous
        (unchanged). Do NOT edit PRD.md."
  section: "h3.15 (§14.1 rule 5), h3.16 (§14.2), h3.11 (§8.4), h3.20 (§14.6 table), h3.21 (§14.7)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/006_bab1774043df/tasks.json
  why: "P1.M2.T1.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes it; tasks.json wins
        on any conflict — EXCEPT the two refinements (bash cur!=-* gate; fish mutual-exclusion) which are PRD-faithful
        improvements over the contract's literal wording, documented in verified_facts §2/§4."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && wc -l completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish
   89 completions/skilldozer.bash
   72 completions/_skilldozer
   69 completions/skilldozer.fish
$ bash -n completions/skilldozer.bash && echo "baseline bash -n OK"
$ grep -c -- '--link' completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish   # --link already present in all 3
# main.go parseArgs has --link (lockstep); P1.M1.T1.S1 makes it multi-target. No flag-set change needed here.
```

### Desired Codebase tree with files to be changed

```bash
completions/skilldozer.bash    # ADD the scan guard before case "$prev" (+ comment update)
completions/_skilldozer        # BRANCH the args state on --link-seen (+ comment)
completions/skilldozer.fish    # ADD dir directive + TIGHTEN tag directive's -n (+ comments)
# main.go / main_test.go / README.md / go.mod / go.sum — UNCHANGED
```

| File | Responsibility |
|---|---|
| `completions/skilldozer.bash` | Offer dirs at every positional after `--link` (scan guard), preserving flag/value routing. |
| `completions/_skilldozer` | Offer `_files` (dirs) in the `args` state when `--link` precedes the cursor, preserving `_arguments` flag routing. |
| `completions/skilldozer.fish` | Offer `__fish_complete_directories` when `--link` seen; suppress the tag directive then (mutual exclusion). |

### Known Gotchas of our codebase & Library Quirks

```bash
# GOTCHA #1 (bash) — GATE the scan guard on `cur != -*`, not "bypass entirely" (contract literal). PRD §14.1 rule 5
# says "every following POSITIONAL"; a dashed token is a FLAG, not a positional. So `--link d1 -<tab>` should still
# offer flags (matches the Go parser: dashed tokens after --link are flags, not link targets). Gating satisfies ALL
# contract OUTPUT cases AND keeps flag completion working after --link. (verified_facts §2.)

# GOTCHA #2 (bash) — Place the guard AFTER _init_completion (so words/cword/cur are set in BOTH the package-present
# and fallback branches) and BEFORE case "$prev". The fallback branch (manual COMP_WORDS) is critical on minimal
# Linux / macOS default bash — if the guard ran before _init_completion, words/cword would be unset there.

# GOTCHA #3 (bash) — LEAVE `--link` in the `--store|--init|--link)` case line. It's now redundant (the guard catches
# `--link <tab>` too) but harmless (defense in depth). Minimal diff = don't remove it. Update only the value-routing
# comment to note --link completes dirs at EVERY position.

# GOTCHA #4 (zsh) — Fix the args CASE, do NOT add a top-level guard. `_arguments -C ... '*: :->args' && return 0`
# handles flags + flag-VALUE slots INLINE (returns 0 → `&& return 0` → case skipped); the case runs ONLY for plain
# positionals. A top-level `if --link in words: _files; return` would bypass `_arguments` and break flag completion
# after --link. Branching the args case preserves all routing. (verified_facts §3.)

# GOTCHA #5 (zsh) — `${words[(i)--link]}` (subscript `(i)` = first-match ascending index) returns the first index of
# --link, or past-end if absent. `< CURRENT` ⇒ --link is a completed word before the cursor. `$words`/`$CURRENT` are
# standard compsys vars, available throughout the #compdef function, NOT clobbered by `_arguments -C` (which only
# sets $curcontext). Use `(( ${words[(i)--link]} < CURRENT ))` (arithmetic).

# GOTCHA #6 (fish) — MUTUAL EXCLUSION is mandatory. fish aggregates ALL `complete` rules whose -n is true. If you
# ADD a dir directive without TIGHTENING the tag directive's -n, then after --link BOTH fire → dirs AND tags offered
# (wrong). The tag directive's -n MUST gain `; and not string match -q -- "--link" (commandline -opc)`. (verified_facts §4.)

# GOTCHA #7 (fish) — `commandline -opc` = tokens up to (NOT including) the current token. So while typing `--link`
# itself, the partial token is the current token (excluded) → the --link-seen condition is false → flag completion
# applies (via the -l link directive). Correct. The condition is true only once --link is a COMPLETED word.

# GOTCHA #8 (fish) — The `-r` on the `--link` line (line 39) owns the FIRST value slot (--link <tab> → dirs via -r).
# The NEW dir directive owns SUBSEQUENT positions (--link d1 <tab>). They don't conflict (different cursor positions).
# Do NOT remove -r. `__fish_complete_directories` is a standard fish helper (dirs only).

# GOTCHA #9 — REBUILD before the gate. //go:embed reads completions/* at BUILD time. After editing, `go build` (or
# `go test`, which builds) re-embeds → TestEmbeddedCompletionsMatchOnDisk (embedded==on-disk) passes. Editing without
# rebuilding → the test compares stale embedded bytes to the new file → FAILS.

# GOTCHA #10 — No conflict with the parallel sibling P1.M1.T1.S1 (main.go/main_test.go/README.md — disjoint files).
# The shared touchpoint TestEmbeddedCompletionsMatchOnDisk is untouched by the sibling and preserved by this task
# (rebuild). The --link flag is already in all three completion files (no flag-set change).

# GOTCHA #11 — Only the three completion files change. Do NOT touch main.go/main_test.go/README.md (P1.M1.T1.S1 /
# P1.M3.T1.S1), or PRD.md/tasks.json/prd_snapshot.md/.gitignore. No Go/deps changes; go.mod/go.sum byte-for-byte
# identical (not even referenced — these are .bash/.zsh/.fish files).
```

---

## Implementation Blueprint

### Data models and structure

**None.** This is a shell-script edit across three files. No Go types, no structs.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT completions/skilldozer.bash — add the scan guard before case "$prev"
  - FILE: completions/skilldozer.bash
  - PLACE the guard AFTER the _init_completion block (which sets words/cword/cur) and BEFORE the `case "$prev" in`
    block (GOTCHA #2). FIND the `case "$prev" in` line (~line 41) and insert the guard immediately above it:
        # Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` appears anywhere before the cursor,
        # EVERY following POSITIONAL completes to a directory (--link takes many dirs). Dashed tokens are still
        # flags, so only fire when the current token is a positional (cur does not start with '-'). The first
        # position after --link is also caught here (and by the case below) — redundant but harmless.
        if [[ "$cur" != -* ]]; then
            local i
            for ((i=1; i<cword; i++)); do
                [[ "${words[i]}" == "--link" ]] && { COMPREPLY=($(compgen -d -- "$cur")); return 0; }
            done
        fi
  - GOTCHA #1: the `cur != -*` gate (PRD "positional"); GOTCHA #3: leave `--link` in the case line below.
  - UPDATE the value-routing comment (~lines 35-40): change the `--link -> directory value -> complete DIRECTORIES
    via compgen -d (§8.4)` line to note "at EVERY position after --link (§8.4 multi-link)".

Task 2: EDIT completions/_skilldozer — branch the args state on --link-seen
  - FILE: completions/_skilldozer (zsh)
  - FIND the args case (~lines 64-67):
        case "$state" in
            args)
                # Positionals are ALWAYS skills (decision 19: no bare subcommands),
                # and skills are never mutually exclusive — offer them on every
                # positional <tab>, first or later.
                compadd -- "$tags[@]"
                ;;
        esac
  - REPLACE the compadd line with a branch (GOTCHA #4 — fix the case, NOT a top-level guard; GOTCHA #5 — the idiom):
        case "$state" in
            args)
                # Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` has been typed before the
                # cursor, EVERY following positional is a directory to link (not a tag). `${words[(i)--link]}` =
                # first index of --link (or past-end if absent); `< CURRENT` ⇒ --link is a completed word before
                # the cursor. Flag/value routing is unaffected (dashed tokens + flag value slots never reach this
                # args state — _arguments handles them inline).
                if (( ${words[(i)--link]} < CURRENT )); then
                    _files
                else
                    # Positionals are ALWAYS skills (decision 19); offer them on every positional.
                    compadd -- "$tags[@]"
                fi
                ;;
        esac
  - PRESERVE the `_arguments -C "$flags[@]" '*: :->args' && return 0` line and the `_skilldozer "$@"` self-call.

Task 3: EDIT completions/skilldozer.fish — add the dir directive + tighten the tag directive
  - FILE: completions/skilldozer.fish
  - (3a) ADD the dir directive. Place it near the --link -r line (~line 39) OR grouped with the tag directive
         (~line 65). FIND a stable anchor (e.g. after the `--link ... -r` line) and ADD:
            # Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` has been typed, EVERY following
            # positional is a directory to link (not a tag). Fires at every position after --link (the first
            # position is handled by -r above).
            complete -c skilldozer -n 'string match -q -- "--link" (commandline -opc)' \
                -a '(__fish_complete_directories)' -d 'skill directory to link'
  - (3b) TIGHTEN the tag directive's -n (GOTCHA #6 — mandatory mutual exclusion). FIND (~lines 65-69):
            complete -c skilldozer -n 'not __fish_prev_arg_in --search' \
                -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
    REPLACE the -n with (add the --link-seen suppression):
            complete -c skilldozer -n 'not __fish_prev_arg_in --search; and not string match -q -- "--link" (commandline -opc)' \
                -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
    (Update the directive's leading comment to note the --link suppression: "Suppressed when the previous arg is
    --search (free-text query) AND when --link has been seen (post-link positionals are dirs, not tags — §8.4).")
  - GOTCHA #7 (commandline -opc excludes the current token); GOTCHA #8 (leave -r; __fish_complete_directories).

Task 4: VERIFY (syntax + embed identity + behavior)
  - bash -n completions/skilldozer.bash                      # exit 0
  - zsh -n completions/_skilldozer 2>/dev/null && echo OK || echo "(zsh absent — skip; compsys syntax is standard)"
  - go build -o skilldozer .                                  # REBUILD so //go:embed re-reads the files (GOTCHA #9)
  - go test ./...                                             # incl. TestEmbeddedCompletionsMatchOnDisk (auto-pass on rebuild)
  - git diff --name-only | grep -E '\.go$|README' && echo "FAIL: scope" || echo "scope OK (completions only)"
  - behavior spot-check (drive the bash fn directly — see Level 4)
```

### Implementation Patterns & Key Details

```bash
# bash (Task 1) — the scan guard, placed after _init_completion, before case "$prev":
if [[ "$cur" != -* ]]; then
	local i
	for ((i=1; i<cword; i++)); do
		[[ "${words[i]}" == "--link" ]] && { COMPREPLY=($(compgen -d -- "$cur")); return 0; }
	done
fi
```
```zsh
# zsh (Task 2) — branch the args case (NOT a top-level guard):
case "$state" in
	args)
		if (( ${words[(i)--link]} < CURRENT )); then
			_files
		else
			compadd -- "$tags[@]"
		fi
		;;
esac
```
```fish
# fish (Task 3) — mutual exclusion: new dir directive + tightened tag directive.
complete -c skilldozer -n 'string match -q -- "--link" (commandline -opc)' \
    -a '(__fish_complete_directories)' -d 'skill directory to link'
complete -c skilldozer -n 'not __fish_prev_arg_in --search; and not string match -q -- "--link" (commandline -opc)' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
```

Notes easy to get wrong:
- **bash:** gate on `cur != -*` (flags are still completable after --link); place AFTER `_init_completion`.
- **zsh:** fix the `args` CASE, not a top-level guard (preserves `_arguments` flag routing); use `${words[(i)--link]} < CURRENT`.
- **fish:** you MUST tighten the tag directive's `-n` too, or dirs+tags both fire after --link.
- **all three:** REBUILD (`go build`/`go test`) so `//go:embed` re-reads the files, or `TestEmbeddedCompletionsMatchOnDisk` fails.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **bash: gate the guard on `cur != -*` vs "bypass entirely" (contract literal)? → GATE.** PRD §14.1 rule 5 says "every following POSITIONAL"; a dashed token is a flag, not a positional. Gating keeps flag completion working after `--link` (`--link d1 -<tab>` → flags) and satisfies all OUTPUT cases. (GOTCHA #1.)
2. **zsh: top-level guard vs args-case branch? → ARGS-CASE BRANCH.** A top-level `if --link: _files; return` bypasses `_arguments` and breaks flag/value routing after --link. Branching the `args` case preserves all routing (flags + flag-value slots are handled inline by `_arguments` and never reach `args`). Equivalent behavior, cleaner. (GOTCHA #4.)
3. **fish: one directive vs mutual exclusion? → MUTUAL EXCLUSION.** fish aggregates all matching directives; a new dir directive alone would leave the tag directive firing too (dirs+tags after --link). Tighten the tag directive's `-n` to suppress when `--link` seen. (GOTCHA #6.)
4. **Leave `--link` in the bash `case "$prev"` line? → YES (minimal diff).** It's redundant (the guard subsumes it) but harmless; removing it would require a comment rewrite for no behavioral gain. (GOTCHA #3.)
5. **Update LOCKSTEP comments? → only the inline --link comments.** The LOCKSTEP comment is about the FLAG SET matching parseArgs; the flag set is unchanged (--link already present). The inline --link comments DO change (they now describe multi-position behavior). No LOCKSTEP flag-set edit.

### Integration Points

```yaml
EMBED (the load-bearing integration):
  - main.go's //go:embed completions/{skilldozer.bash,_skilldozer,skilldozer.fish} (unchanged) reads these files at
    BUILD time. Editing + `go build` → the embed vars hold the new bytes → `skilldozer --completions --shell <x>`
    emits them. TestEmbeddedCompletionsMatchOnDisk (embedded==on-disk) auto-passes. (GOTCHA #9.)

LOCKSTEP (§14.4):
  - The flag set is unchanged (--link already in all three). The behavior change (multi-position dirs) is the
    completion-side mirror of the Go-side multi-target --link (P1.M1.T1.S1). Both land together (P1.M1 before P1.M2).

NO GO SOURCE / NO ROUTES / NO DATABASE / NO DEPS / NO NEW FILES:
  - Three .bash/.zsh/.fish files only. go.mod/go.sum/main.go/main_test.go/README.md unchanged.
```

---

## Validation Loop

### Level 1: Syntax (immediate, after each edit)

```bash
cd /home/dustin/projects/skilldozer

bash -n completions/skilldozer.bash                                                # MUST exit 0
zsh  -n completions/_skilldozer 2>/dev/null && echo "zsh -n OK" || echo "(zsh absent — skip; syntax is standard compsys)"
# fish has no standalone -n linter widely installed; rely on go test + a live fish smoke if available:
command -v fish >/dev/null && fish -n completions/skilldozer.fish && echo "fish -n OK" || echo "(fish absent — rely on go test + syntax review)"
# Expected: bash -n exit 0; zsh/fish either OK or "absent" (their syntax is standard; go test is the real gate).
```

### Level 2: Embed identity (the core Go gate) — REBUILD first

```bash
cd /home/dustin/projects/skilldozer

go build ./...                    # GOTCHA #9: re-embed the edited files
go test ./... ; echo "test exit $?"   # MUST be 0 — incl. TestEmbeddedCompletionsMatchOnDisk (embedded==on-disk)
go test -run 'TestEmbeddedCompletionsMatchOnDisk' -v ./...   # the load-bearing assertion
# Expected: 0 failures; the embed-identity test passes (the on-disk files == the embedded bytes post-rebuild).
```

### Level 3: Scope invariants (stayed in the completions lane)

```bash
cd /home/dustin/projects/skilldozer

git diff --name-only              # Expected: ONLY the three completion files
git diff --name-only | grep -E '\.go$|README|^go\.' && echo "FAIL: scope" || echo "scope OK (completions only)"
# No Go/deps change:
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
```

### Level 4: Behavior spot-checks (drive the completion fns directly + the §14.6 emit)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz .

# (a) BASH — drive _skilldozer_completion via COMP_WORDS (no shell interaction needed):
bash -c '
  source completions/skilldozer.bash
  # --link d1 <tab> → DIRS (the NEW behavior):
  COMP_WORDS=(skilldozer --link d1 ""); COMP_CWORD=3; _skilldozer_completion
  echo "bash --link d1 <tab>: ${COMPREPLY[*]:0:2} (want dir paths)"
  # bare <tab> → skills (unchanged):
  COMP_WORDS=(skilldozer ""); COMP_CWORD=1; _skilldozer_completion
  echo "bash <tab>: ${COMPREPLY[*]:0:2} (want skill tags / store skills)"
  # -<tab> → flags (unchanged):
  COMP_WORDS=(skilldozer -); COMP_CWORD=1; _skilldozer_completion
  echo "bash -<tab>: ${COMPREPLY[*]:0:2} (want --long flags)"
'

# (b) ZSH/FISH — drive directly if the shell is installed (else rely on §13 acceptance + the code review):
command -v zsh >/dev/null && zsh -c '
  autoload -U compinit && compinit -u 2>/dev/null
  fpath=(completions $fpath); autoload _skilldozer
  # a full programmatic drive of _arguments is involved; the §13 acceptance gate (below) is the canonical check
' || echo "(zsh absent — rely on §13 acceptance emit-check)"
command -v fish >/dev/null && echo "(fish: rely on §13 acceptance + the complete-rule review)" || echo "(fish absent)"

# (c) §14.6 emit — the rebuilt binary's --completions carries the new bytes:
/tmp/sdz --completions --shell bash 2>/dev/null | grep -q -- '--link' && echo "bash emit has --link OK"
/tmp/sdz --completions --shell bash 2>/dev/null | grep -q 'words\[i\] == "--link"' && echo "bash emit has the scan guard OK"
/tmp/sdz --completions --shell zsh 2>/dev/null | grep -q 'words\[(i)--link\]' && echo "zsh emit has the args branch OK" \
  || echo "(zsh emit is DERIVED via zshEvalScript — grep the autoload file instead: grep -q 'words\[(i)--link\]' completions/_skilldozer)"
/tmp/sdz --completions --shell fish 2>/dev/null | grep -q '__fish_complete_directories' && echo "fish emit has the dir directive OK"

rm -f /tmp/sdz
# Expected: bash COMP_WORDS drive shows dirs after --link d1, skills on bare <tab>, flags on -<tab>; the emit greps confirm the new bytes.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `bash -n` exit 0; `zsh -n`/`fish -n` OK or shell-absent (standard syntax)
- [ ] Level 2 PASS — `go build` + `go test ./...` exit 0; `TestEmbeddedCompletionsMatchOnDisk` passes (rebuild done)
- [ ] Level 3 PASS — `git diff --name-only` = ONLY the three completion files; `go.mod`/`go.sum` unchanged
- [ ] Level 4 PASS — bash COMP_WORDS drive: `--link d1 <tab>` → dirs, `<tab>` → skills, `-<tab>` → flags; emit greps confirm new bytes

### Feature Validation
- [ ] bash: scan guard before `case "$prev"` (gated on `cur != -*`); `--link`/`--link d1`/`--link d1 d2` `<tab>` → dirs
- [ ] zsh: `args` case branches on `${words[(i)--link]} < CURRENT`; `--link d1 <tab>` → `_files` (dirs)
- [ ] fish: new dir directive + tightened tag directive `-n`; `--link d1 <tab>` → `__fish_complete_directories`
- [ ] `--link <tab>` (first position) still → dirs in all three; bare `<tab>` → skills; `-<tab>` → flags
- [ ] `--store`/`--init`/`--search`/`--shell` routing unchanged in all three; §14.7 list-ambiguous unchanged

### Code Quality / Convention Validation
- [ ] bash guard mirrors the existing `compgen -d` idiom; zsh uses standard `${words[(i)...]}`/`_files`; fish uses standard `commandline -opc`/`__fish_complete_directories`
- [ ] Comments cite §8.4 / §14.1 rule 5; no LOCKSTEP flag-set change (--link already present)
- [ ] No Go/deps changes; go.mod/go.sum byte-for-byte identical (not even referenced)

### Scope Discipline
- [ ] Did NOT touch `main.go`, `main_test.go`, `README.md` (P1.M1.T1.S1 / P1.M3.T1.S1)
- [ ] Did NOT change the flag list / LOCKSTEP comment (flag set unchanged — --link already present)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't add a top-level zsh guard.** It bypasses `_arguments` and breaks flag/value routing after --link. Fix the `args` case instead (GOTCHA #4).
- ❌ **Don't add the fish dir directive without tightening the tag directive's `-n`.** fish aggregates all matching rules → dirs AND tags both fire after --link (GOTCHA #6).
- ❌ **Don't bypass flag completion "entirely" in bash (contract literal).** Gate on `cur != -*` so `--link d1 -<tab>` still offers flags (PRD "positional"; GOTCHA #1).
- ❌ **Don't forget to REBUILD.** `//go:embed` reads the files at build time; without `go build`/`go test`, `TestEmbeddedCompletionsMatchOnDisk` compares stale embedded bytes to the new files and FAILS (GOTCHA #9).
- ❌ **Don't remove `--link` from the bash `case "$prev"` line or the `--link -r` fish line.** They're redundant/orthogonal but harmless; removing them is needless churn (GOTCHA #3/#8).
- ❌ **Don't touch the flag list or LOCKSTEP comment.** The flag set is unchanged (--link already in all three); this is a behavior change, not a flag-set change.
- ❌ **Don't edit main.go/main_test.go/README.md.** Those are the parallel sibling (P1.M1.T1.S1) and the doc sweep (P1.M3.T1.S1) — disjoint (GOTCHA #10/#11).
- ❌ **Don't add Go/deps.** Three shell-script files only; go.mod/go.sum untouched.

---

## Confidence Score

**9/10** — All three files read in full with exact current code + before/after; the per-shell mechanics that make each fix correct (bash `_init_completion`/`words`; zsh `_arguments ... '*: :->args' && return 0` reaches the case only for positionals; fish aggregates all matching directives → needs mutual exclusion) are documented and traced against every OUTPUT case. The two refinements over the contract's literal wording (bash `cur != -*` gate; zsh args-case vs top-level guard; fish mutual-exclusion) are PRD-faithful improvements with explicit justification. The embed-identity/rebuild constraint and the disjoint boundary with P1.M1.T1.S1 are confirmed. The 1-point reservation is for the live-shell drives (Level 4): the bash COMP_WORDS probe is deterministic, but zsh/fish programmatic drives are involved and the canonical check there is the §13 acceptance emit-grep + code review (zsh emit is derived via `zshEvalScript`, so the grep targets the on-disk autoload file, not the emitted wrapper) — the syntax is standard compsys/fish idiom, but a live-shell interaction is the final proof.
