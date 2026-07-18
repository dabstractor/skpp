# Verified Facts — P1.M2.T1.S1: Multi-position directory completion after `--link` (bash, zsh, fish)

Plan `006_bab1774043df` (--link multi-target batch). Every claim below was read
directly from the live source at `/home/dustin/projects/skilldozer` (all three
completion files read in full; `codebase_state.md` §"Completion Scripts" read;
`TestEmbeddedCompletionsMatchOnDisk` read). This is a 3-file completion-script
edit (bash + zsh + fish). Zero Go/deps changes.

---

## §1 — The bug + the boundary

**Bug:** today each shell completes directories ONLY for the first positional
after `--link`. After `--link d1`, the next `<tab>` offers skill TAGS, not dirs.
PRD §14.1 rule 5 + §14.2 + §8.4 require: once `--link` appears, EVERY following
**positional** completes to a directory (`--link` takes many dirs).

**Boundary (disjoint from the parallel sibling):** P1.M1.T1.S1 (Implementing)
implements the GO side of multi-target `--link` (struct/parser/runLink/tests/
README) and EXPLICITLY does NOT touch the completion files (its GOTCHA #10:
"this subtask does NOT touch the completion scripts. Multi-link DIRECTORY
completion is P1.M2.T1.S1"). This task edits ONLY the three completion files.
Plan ordering lands P1.M1 before P1.M2, so the Go side is done when this runs —
but this task does not depend on it: the flag list already includes `--link` in
all three files (no §14.4 flag-set change); the ONLY change is the multi-position
directory behavior.

**Embed identity (§14.6 / TestEmbeddedCompletionsMatchOnDisk):** the test compares
`completionScript(shell)` (the `//go:embed` var) to `os.ReadFile(path)`. After
editing the on-disk files, `go build` (or `go test`, which builds) re-embeds →
embedded == on-disk → test PASSES. (bash/fish are embedded verbatim; zsh is
embedded verbatim too — the `zshEvalScript` derivation happens at EMIT time in
`runCompletion`, not in the embed/test. So the test compares the verbatim on-disk
autoload file to the verbatim embed for all three.) **Rebuild is required** before
`--completions` reflects the edits.

---

## §2 — BASH (`completions/skilldozer.bash`, 89 lines)

**Current** (lines 41-45): the `case "$prev"` completes dirs only when `$prev` is
exactly `--store|--init|--link`. After `--link d1`, prev=d1 → no match → falls to
tag completion (the bug).

**Fix — a scan guard BEFORE the `case "$prev"` block** (contract LOGIC (a); the
contract says "before the existing `$prev` case dispatch"). After `_init_completion`
sets `words`/`cword`/`cur`, scan `words[1..cword-1]` for `--link`; if found AND the
current token is a positional (cur does not start with `-`) → dirs + return 0:

```bash
# Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` appears
# anywhere before the cursor, EVERY following POSITIONAL completes to a directory
# (--link takes many dirs). Dashed tokens are still flags, so only fire when the
# current token is a positional (cur does not start with '-'). The first position
# after --link is ALSO caught here (and by the case below) — redundant but harmless.
if [[ "$cur" != -* ]]; then
    local i
    for ((i=1; i<cword; i++)); do
        [[ "${words[i]}" == "--link" ]] && { COMPREPLY=($(compgen -d -- "$cur")); return 0; }
    done
fi
```

**Placement:** AFTER the `_init_completion` block (so words/cword/cur are set, in
BOTH the package-present and fallback branches) and BEFORE the `case "$prev"` block.

**GATE on `cur != -*` (deliberate refinement of the contract's "bypass entirely"):**
the contract LOGIC (a) literally says "bypassing tag/flag completion entirely",
but PRD §14.1 rule 5 qualifies "every following POSITIONAL". A dashed token is a
FLAG, not a positional — so `--link d1 -<tab>` should still offer flags (matches
the Go parser, where dashed tokens after --link are flags, not link targets).
Gating on `cur != -*` satisfies ALL contract OUTPUT cases AND keeps flag
completion working after --link. (Without the gate, `--link d1 -<tab>` → dirs
filtered by `-` → empty/silent — harmless but less correct.) This is the
PRD-consistent choice.

**Preserved (no change to the `case "$prev"` line):** `--link` STAYS in
`--store|--init|--link)` — it's now redundant (the guard catches `--link <tab>`
too, since --link is in words) but harmless (defense in depth). Minimal diff =
leave it. Update the value-routing comment (lines 35-40) to note --link now
completes dirs at EVERY position.

**Trace (all OUTPUT cases):**
- `--link <tab>` (prev=--link, cur=""): guard — cur not -*, words[1]="--link" < cword=2 → dirs. ✓
- `--link d1 <tab>` (prev=d1, cur=""): guard — words[1]="--link" < cword=3 → dirs. ✓ (NEW)
- `--link d1 d2 <tab>` (prev=d2, cur=""): guard — --link in words < cword=4 → dirs. ✓ (NEW)
- `<tab>` (cur=""): no --link in words → guard skips → tags. ✓ (unchanged)
- `-<tab>` (cur="-"): no --link → guard skips (cur is -* anyway) → flag block → flags. ✓ (unchanged)
- `--store <tab>` (prev=--store): no --link → guard skips → case "$prev" `--store)` → dirs. ✓ (unchanged)

---

## §3 — ZSH (`completions/_skilldozer`, 72 lines)

**Current** (lines 51, 62-67): `'--link[...]:directory:_files'` (single-value —
the `:directory:_files` handles ONLY the value slot right after `--link`). The
`_arguments -C "$flags[@]" '*: :->args'` catch-all routes ALL positionals to
`$state=args`, and the `case "$state" args) compadd -- "$tags[@]"` offers TAGS for
every positional (the bug: `--link d1 <tab>` → tags).

**KEY zsh mechanic:** `_arguments -C ... '*: :->args' && return 0` — when the
current word is a FLAG (dashed) or a flag VALUE (the `:directory:_files` slots),
`_arguments` handles it INLINE and returns 0 → `&& return 0` fires → the `case`
is SKIPPED. The `case "$state"` runs ONLY for plain POSITIONALS. So fixing the
`args` case preserves all flag/value routing automatically.

**Fix — branch the `args` case on whether `--link` precedes the cursor** (cleaner
than the contract's "top-level guard" or `'*::link-dirs:_files'` restructure
options; equivalent behavior, preserves `_arguments` flag routing):

```zsh
case "$state" in
    args)
        # Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` has
        # been typed before the cursor, EVERY following positional is a directory
        # to link (not a tag). `${words[(i)--link]}` = first index of --link (or
        # past-end if absent); `< CURRENT` ⇒ --link is a completed word before the
        # cursor. Flag/value routing is unaffected (dashed tokens + flag value slots
        # never reach this `args` state — _arguments handles them inline).
        if (( ${words[(i)--link]} < CURRENT )); then
            _files
        else
            # Positionals are ALWAYS skills (decision 19); offer them on every positional.
            compadd -- "$tags[@]"
        fi
        ;;
esac
```

**`${words[(i)--link]}` (zsh subscript `(i)` = first-match ascending index):**
returns the index of the first `--link`, or an index past the end if absent. `< CURRENT`
⇒ `--link` exists and was completed before the cursor. `$words`/`$CURRENT` are
standard compsys variables available throughout the `#compdef` function (not
clobbered by `_arguments -C`, which only sets `$curcontext`). `_files` completes
the current positional as a directory (zsh prefix-filters automatically).

**Trace:** `--link <tab>` → the `--link` spec's `:directory:_files` handles it inline
(doesn't reach args) → dirs. ✓ (unchanged first position). `--link d1 <tab>` →
positional → args state → `--link` at index 2 < CURRENT=4 → `_files` → dirs. ✓ (NEW).
`<tab>` → args → no --link → tags. ✓. `-<tab>` → `_arguments` handles the flag inline
→ `&& return 0` → never reaches args. ✓.

---

## §4 — FISH (`completions/skilldozer.fish`, 69 lines)

**Current** (line 39, 65-69): `complete -c skilldozer -l link ... -r` (`-r` makes
ONLY the next arg a file arg — the first position after `--link`). The positional
tag directive (lines 65-69) `complete -c skilldozer -n 'not __fish_prev_arg_in --search' -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'` then offers TAGS for
every subsequent positional (the bug: `--link d1 <tab>` → tags).

**KEY fish mechanic:** fish aggregates ALL `complete` rules whose `-n` condition is
true for the current cursor position. So the fix is a MUTUAL EXCLUSION of two
positional directives: (a) a NEW dir directive that fires only when `--link` has
been seen, and (b) the EXISTING tag directive with its `-n` tightened to ALSO
suppress when `--link` has been seen.

**`--link`-seen condition:** `string match -q -- "--link" (commandline -opc)`.
`commandline -opc` = the command-line tokens up to (NOT including) the current
token being completed. `string match -q -- "--link" <tokens>` is true iff `--link`
is among those completed tokens. (While typing `--link` itself, the partial token
is the current token and excluded, so the condition is false — correct, flag
completion applies then via the `-l link` directive.)

**Fix — add the dir directive + tighten the tag directive's `-n`:**

ADD (after the existing `--link` `-r` line, ~line 39, or grouped with the tag
directive near line 65):
```fish
# Multi-link directory completion (§8.4 / §14.1 rule 5): once `--link` has been
# typed, EVERY following positional is a directory to link (not a tag). Fires at
# every position after --link (the first position is handled by -r above).
complete -c skilldozer -n 'string match -q -- "--link" (commandline -opc)' \
    -a '(__fish_complete_directories)' -d 'skill directory to link'
```

MODIFY the existing tag directive (lines 65-69) — tighten `-n` to also suppress
when `--link` has been seen:
```fish
# Dynamic tags ... Suppressed when the previous arg is --search (free-text query)
# AND when --link has been seen (post-link positionals are dirs, not tags — §8.4).
complete -c skilldozer -n 'not __fish_prev_arg_in --search; and not string match -q -- "--link" (commandline -opc)' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
```

**Mutual exclusion:** after `--link` → only the dir directive fires (tag directive's
`-n` is false). Before/no `--link` → only the tag directive fires (dir directive's
`-n` is false). The `-r`/`-x` value-slot machinery (`--store`/`--init`/`--link`/
`--shell`) still owns the flag-value slots (the positional directives don't fire
there). `__fish_complete_directories` is a standard fish helper (dirs only).

**Trace:** `--link <tab>` → `-r` value slot → dirs. ✓ (unchanged). `--link d1 <tab>`
→ positional → dir directive fires (--link seen), tag directive suppressed → dirs. ✓
(NEW). `<tab>` → no --link → tag directive fires → tags. ✓. `--store <tab>` → `-r`
value slot → dirs (no --link). ✓ (unchanged).

---

## §5 — Preservation matrix (what MUST stay unchanged)

The contract OUTPUT is emphatic: "KEEP all existing behavior intact ... The ONLY
change is the multi-position directory behavior after --link." Verified per shell:

| Behavior | bash | zsh | fish | Status |
|---|---|---|---|---|
| bare `<tab>` → skills | guard skips (no --link) → tags | args, no --link → compadd tags | tag directive fires (no --link) | ✓ unchanged |
| `-<tab>` → long-form flags | guard skips (cur is -*) / no --link → flag block | `_arguments` inline → flags | `-l *` directives | ✓ unchanged |
| `--store`/`--init <tab>` → dirs | case "$prev" (no --link) | `:directory:_files` inline | `-r` value slot | ✓ unchanged |
| `--search <tab>` → nothing | case "$prev" `--search)` | `:query:` inline (no offer) | tag directive suppressed (`__fish_prev_arg_in --search`) | ✓ unchanged |
| `--shell <tab>` → bash/zsh/fish enum | case "$prev" `--shell)` | `:shell:(bash zsh fish)` inline | `-x -a "bash zsh fish"` | ✓ unchanged |
| §14.7 list-ambiguous options | the `bind show-all-if-ambiguous` block (lines 76-89) | (setopt in the derived wrapper, not this file) | (fish lists by default) | ✓ unchanged |
| flag list / LOCKSTEP | the compgen -W line (line 50, already has --link) | the flags array (already has --link) | the `-l *` directives (already has --link) | ✓ unchanged (no flag-set change) |
| tag probe byte-identical | `tags=$(skilldozer --relative --all 2>/dev/null)` | `tags=(${(f)"$(skilldozer --relative --all 2>/dev/null)"})` | `(skilldozer --relative --all 2>/dev/null)` | ✓ unchanged |

---

## §6 — No conflict with the parallel sibling (disjoint files)

P1.M1.T1.S1 edits `main.go` + `main_test.go` + `README.md` (Go side). This task
edits `completions/{skilldozer.bash,_skilldozer,skilldozer.fish}` ONLY. **Disjoint
files; no merge collision; land in either order.** The one shared touchpoint is
`TestEmbeddedCompletionsMatchOnDisk` — the sibling does NOT touch it (it's a
main_test.go test that reads completions/*, which the sibling leaves unchanged),
and this task preserves it via the rebuild (§1). The sibling's `--link` flag is
ALREADY in all three completion files (no flag-set change needed here).

---

## §7 — Scope discipline + validation

- Edit ONLY the three completion files. Do NOT touch `main.go`, `main_test.go`,
  `README.md` (P1.M1.T1.S1 / P1.M3.T1.S1), or `PRD.md`/`tasks.json`/etc.
- No Go/deps changes. The gate is `bash -n` + `zsh -n` (if available) + fish syntax
  + the §14.6 embed identity (rebuild) + `go test ./...`.
- Validation commands:
```bash
bash -n completions/skilldozer.bash                      # syntax gate
zsh -n completions/_skilldozer 2>/dev/null && echo OK || echo "(zsh absent — skip; compsys syntax is standard)"
go build -o skilldozer .                                  # REBUILD so //go:embed re-reads the files
go test ./...                                             # incl. TestEmbeddedCompletionsMatchOnDisk (auto-pass on rebuild)
git diff --name-only | grep -E '\.go$|README' && echo "FAIL: touched Go/README" || echo "scope OK (completions only)"
# content spot-check: --link d1 <tab> behavior (drive the completion fn directly)
go build -o /tmp/sdz . && bash -c '
  source completions/skilldozer.bash
  COMP_WORDS=(skilldozer --link d1 ""); COMP_CWORD=3; _skilldozer_completion
  echo "bash --link d1 <tab> offers (dirs, not tags): ${COMPREPLY[*]:0:3} …"
'; rm -f /tmp/sdz
```
