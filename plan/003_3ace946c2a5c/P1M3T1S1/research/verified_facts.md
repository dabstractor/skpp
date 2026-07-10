# Verified Facts — P1.M3.T1.S1 (Add `completion` as a completable first-arg subcommand to all 3 completion files)

Shell-script lockstep edit (PRD §14.4 + §14.6). All facts read directly from source at PRP-write time.
Repo: `/home/dustin/projects/skilldozer`. Plan: `003_3ace946c2a5c`.

## §0 — What this task IS

Add `completion` (the new reserved subcommand from P1.M2.T1.S1) as a **completable first-arg
exclusive subcommand** to all three shell completion files, mirroring the EXISTING `check`/`init`
pattern exactly (three-part: suppression walk + first-positional-only offer + tag suppression once
seen). **Shell-script edits only** — NO Go code. The authoritative per-file edit list is
`architecture/completions_change_map.md`; every anchor below was cross-checked against the LIVE files.

## §1 — `completion` is exclusive like check/init (justifies the suppression-walk treatment)

Verified in main.go (P1.M2.T1.S1 = Complete):
- main.go:166 `completion bool // ... exclusive like check/init`.
- main.go:297 "run()'s exclusivity rejects completion+tags / completion+mode with exit 2."
- main.go:534 "completion dispatch (PRD §14.6). completion is an exclusive mode (like ...)".

So `completion` behaves EXACTLY like check/init for completion purposes: offered ONLY as the first
positional token; once seen, tag completion is suppressed (a tag after it would be a guaranteed exit 2
per §6.3). This is WHY the change map mirrors the check/init pattern verbatim — there is nothing
special about `completion` relative to check/init from the completion-file perspective.

## §2 — The exact current state of the three touchpoints in each file (the check/init pattern to mirror)

### bash (`completions/skilldozer.bash`, 69 lines) — 2 edits
- **Suppression walk** (the `for ((i=1; i<cword; i++))` loop): `[[ "${words[i]}" == "check" || "${words[i]}" == "init" ]] && return 0`
- **First-pos offer**: `(( have_pos == 0 )) && cands="$cands check init"`

### zsh (`completions/_skilldozer`, 61 lines) — 2 edits
- **First-pos compadd** (the `first)` state): `compadd -- "$tags[@]" check init`
- **Suppression condition** (the `rest)` state): `if (( ${words[(I)check]} || ${words[(I)init]} )); then`

### fish (`completions/skilldozer.fish`, 51 lines) — 2 edits
- **First-arg directive block**: the two `complete -c skilldozer -n '__fish_is_first_arg' -a 'check'|'init' ...` lines
- **Tag-suppression predicate**: `complete -c skilldozer -n 'not __fish_seen_subcommand_from check init; and not __fish_prev_arg_in --search -s' \`

(All four anchors above were read verbatim from the live files and match completions_change_map.md.)

## §3 — The exact per-file edits (from completions_change_map.md, cross-checked)

### bash
1. Suppression walk: `[[ ... == "check" || ... == "init" ]]` → append `|| "${words[i]}" == "completion"` (inside the `[[ ]]`).
2. First-pos offer: `cands="$cands check init"` → `cands="$cands check init completion"`.

### zsh
1. First-pos compadd: `compadd -- "$tags[@]" check init` → `compadd -- "$tags[@]" check init completion`.
2. Suppression: `if (( ${words[(I)check]} || ${words[(I)init]} )); then` → add `|| ${words[(I)completion]}`.

### fish
1. ADD a new directive after the `init` directive:
   `complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'`
2. Tag-suppression predicate: `__fish_seen_subcommand_from check init` → `__fish_seen_subcommand_from check init completion`.

## §4 — The fish `-d` description matches the binary's USAGE row (consistency)

The contract fixes the fish description as `'Emit the shell completion script for eval'`. Verified this
is byte-consistent with the binary's own USAGE row (P1.M2.T1.S1, main.go:107):
`completion [--shell <name>]   Emit the shell completion script for eval (§14.6)`.
So the fish `-d` text == the USAGE description (minus the trailing `(§14.6)` PRD cite, which the
completion files never carry). Consistency holds.

## §5 — `--shell` is a REAL flag but CONTEXT-ONLY — do NOT add to the flag matrix (CRITICAL scope boundary)

Verified `--shell` is parsed (main.go:228 `case "--shell"`, sets `c.completion=true` +
`c.completionShell`). BUT the contract + §14.1/§14.2 are explicit: the completion files' flag matrix
is the §6.1/§6.2 GLOBAL flag set ONLY; `--shell` is `completion`-context-only and must NOT appear in
any flag matrix. Rationale: `--shell` only makes sense after `completion`; offering it as a global flag
(`skilldozer --<TAB>`) would be wrong, and a context-gated offer is out of scope for this 1-point
lockstep task. Consequence: `skilldozer completion --<TAB>` will NOT offer `--shell` (the user types it
manually); `skilldozer completion --shell <TAB>` offers nothing for the value (it falls through to the
`completion`-seen suppression). This is the contract's deliberate decision — HONOR it; do NOT "help by
adding --shell".

## §6 — Cross-cutting "do NOT" constraints (from change map + contract + §14)

1. Do NOT add `--shell` to any flag matrix (§5 above).
2. Do NOT change tag completion — `skilldozer --relative --all 2>/dev/null` stays byte-for-byte
   identical in all three files (bash `tags=$(...)`, zsh `tags=(${(f)"$(...)"})`, fish
   `-a '(skilldozer --relative --all 2>/dev/null)'`).
3. Do NOT change `--search`/`--store` value routing (the `case "$prev"` in bash, the `:query:`/`:directory:
   _files` specs in zsh, the `-r`/no-`-r` directives in fish).
4. Do NOT change the global no-file-completion rule (bash `complete -F`, zsh `#compdef`, fish `complete -f`).

## §7 — Shell syntax-checkers are available (validation gates)

All three shells are installed and the CURRENT files pass their syntax checks (baseline confirmed):
- `bash -n completions/skilldozer.bash` → exit 0 (bash 5.3.15)
- `zsh -n completions/_skilldozer` → exit 0 (zsh 5.9.1)
- `fish --no-execute completions/skilldozer.fish` → exit 0 (fish 4.7.1)

These are the project-specific validation gates proving the edits stay syntactically valid (a stray
quote/paren in a shell file would fail). Re-run after EACH file's edit.

## §8 — Embed sync (§14.6): the on-disk files are //go:embed'd; go test rebuilds

The three files are compiled into the binary via `//go:embed` (main.go:54/57/60) and surfaced by
`completionScript(shell)` (main.go:1121). `TestEmbeddedCompletionsMatchOnDisk` (main_test.go:2929)
asserts `completionScript(shell)` bytes == `os.ReadFile("completions/...")` bytes (the §14.6 byte-
identity lock). Because `go test` REBUILDS the package (re-running //go:embed), editing the on-disk
files keeps this test GREEN automatically (the re-embed picks up the new bytes). So:
- `go test ./...` is the gate that proves embed == on-disk after the edit (rebuild happens implicitly).
- A STANDALONE pre-built binary would NOT reflect the edits until `go build` is re-run — so the
  validation must REBUILD and confirm `./skilldozer completion --shell bash` OUTPUT now contains
  `completion` (the contract DOCS note: "a rebuild is required for `skilldozer completion` to emit the
  updated bytes"). The parallel P1.M2.T2.S2 wires `run()`→`runCompletion`→`completionScript`, so by
  execution time `skilldozer completion` emits the embedded bytes end-to-end.

## §9 — Scope boundary / what NOT to touch

- `main.go` / `main_test.go` / `internal/*` — UNCHANGED (all completion machinery is P1.M2's scope,
  Complete/Implementing; this task is the shell-file lockstep only).
- `PRD.md`, `tasks.json`, `prd_snapshot.md` — READ-ONLY. (PRD §14.1 still lists "check and init" as the
  exclusive subcommands; updating it to mention `completion` would be a PRD edit — human-owned, out of
  scope. The completion FILES + README (P1.M3.T1.S2) are the user-facing surfaces that matter.)
- README.md — P1.M3.T1.S2 (sibling) documents the `completion` subcommand; do NOT touch it here.
- The `check`/`init` handling in each file — mirror it, do NOT rewrite it.

## §10 — Parallel-execution consideration

P1.M2.T2.S2 (the `run()` completion dispatch) is "Implementing" in parallel. It edits ONLY main.go +
main_test.go. This task edits ONLY the three `completions/*` files. DISJOINT file sets → no merge
conflict; land in either order. At THIS task's execution time, `skilldozer completion` works end-to-end
(S2 landed), so the "rebuild + emit check" validation gate is runnable. Pin all edit seams by the
anchor TEXT in §2 (not line numbers — the files are stable but text-anchoring is clearer and safer).
