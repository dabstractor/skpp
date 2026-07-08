# P1.M3.T2.S1 — Verified Facts (research notes)

**Item:** Add `init` subcommand + `--store <dir>` flag to bash/zsh/fish completions.
**Method:** read all three files in full; read `main.go parseArgs()` (the LOCKSTEP
source of truth); read `plan/002_38acb6d28a6a/architecture/docs_and_assets_drift.md §4`;
ran `bash -n`/`zsh -n`/`fish -n`/`shellcheck` for baselines; ran a PROGRAMMATIC
bash-completion harness (no real TAB key) to confirm current vs intended behavior.
All commands below were EXECUTED on 2026-07-07 in `/home/dustin/projects/skilldozer`.

---

## §1. The three files under edit (current, byte-exact)

| File | Lines | Bytes | Role |
|---|---|---|---|
| `completions/skilldozer.bash` | 65 | 2504 | bash completion (registered via `complete -F _skilldozer_completion skilldozer`) |
| `completions/_skilldozer` | 55 | — | zsh autoload function (`#compdef skilldozer`) |
| `completions/skilldozer.fish` | 42 | 2611 | fish completion (sourced from `~/.config/fish/completions/`) |

All three carry a **"LOCKSTEP to main.go parseArgs()"** header comment. There is NO
shared source of truth the shells can import — edits must be hand-kept in sync across
all three (drift doc §4 cross-cutting note #3).

No `skpp` residue in any completion file (`grep -rn skpp completions/` → clean).
The rename landed cleanly here; only the `init`/`--store` surface is missing.

---

## §2. main.go parseArgs() — the LOCKSTEP flag set (authoritative)

The completion flag matrices MUST equal this set. Read from `main.go` (parseArgs
switch + usageText). Each long form and its short alias:

| Flag | Short | Value? | Notes (from main.go) |
|---|---|---|---|
| `--version` | `-v` | no | |
| `--help` | `-h` | no | help wins tiebreak |
| `--path` | `-p` | no | |
| `--list` | `-l` | no | |
| `--all` | `-a` | no | |
| `--file` | `-f` | no | modifier |
| `--relative` | — | no | no short form |
| `--no-color` | — | no | no short form |
| `--search <q>` | `-s <q>` | yes (free-text) | value-taking |
| `--store <dir>` | — | yes (DIRECTORY) | **NEW**; no short form; **implies init** (`c.init=true`) |
| `init` | — | positional subcommand | **NEW**; reserved token (like `check`) |
| `check` | — | positional subcommand | already completed |

### init exclusivity (main.go `exclusivityError`, ~line 470)

`init` is an **exclusive** mode, identical in spirit to `check`:
- `init` + tags → exit 2 ("'init' cannot be combined with tag arguments")
- `init` + check/list/search/all/path → exit 2

Therefore completion must treat `init` EXACTLY like `check`: offer it ONLY as the
first positional token; once seen, suppress tags (completion must never invite a
guaranteed exit-2 invocation).

### --store behavior (main.go parseArgs lines ~256-266, 258-266)

`--store <dir>` is value-taking, captures the NEXT argv token (or `--store=<dir>`),
sets `c.init=true` + `c.initStore`. It works at TOP LEVEL too
(`skilldozer --store <dir>` == `skilldozer init --store <dir>`), so it is valid to
offer `--store` ungated in the flag matrix.

### usageText descriptions (the canonical completion `-d` strings, main.go lines 84-85)

```
init [<dir>]      First-run setup: pick/create the skills store and write the config
--store <dir>     Non-interactive store path for init
```
Use these EXACT descriptions in the fish `-d` / zsh `[...]` completion text so the
LOCKSTEP invariant holds (main.go USAGE == completion descriptions).

---

## §3. Current vs intended behavior (PROVEN by programmatic harness)

I sourced `completions/skilldozer.bash`, forced the `_init_completion` fallback
(`unset -f _init_completion`), set `COMP_WORDS`/`COMP_CWORD`, called
`_skilldozer_completion`, and printed `COMPREPLY`:

```
skilldozer <TAB>   → [example] [check]          # init MISSING (drift)
skilldozer init <TAB> → [example]               # TAGS LEAK after init (drift — init is exclusive)
skilldozer check <TAB> → []                     # correct (check suppresses)
```

**Intended AFTER this edit:**
```
skilldozer <TAB>        → [example] [check] [init]   # init added
skilldozer init <TAB>   → []                          # tags suppressed (init exclusive)
skilldozer init --<TAB> → [--store ...]               # flag matrix, --store present
skilldozer init --store <TAB> → <directories>         # dir completion
skilldozer check <TAB>  → []                          # unchanged (no regression)
```

This harness is the strongest deterministic validation available (no interactive
shell needed). It is captured in the PRP Validation Loop Level 3.

---

## §4. Exact edit anchors (line numbers from current files)

### bash (`completions/skilldozer.bash`)

| # | Line | Current | Target |
|---|---|---|---|
| B1 | 30-32 | `case "$prev" in  --search\|-s) return 0 ;;  esac` (+1 comment line) | add arm `--store) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;` |
| B2 | 36-38 | `compgen -W "... --search -s" -- "$cur"` | append ` --store` to the `-W` word list |
| B3 | 42-49 | comment + loop `[[ "${words[i]}" == "check" ]] && return 0` | comment mentions check AND init; condition `== "check" \|\| == "init"` |
| B4 | 59 | `(( have_pos == 0 )) && cands="$cands check"` | `... check init"` |

### zsh (`completions/_skilldozer`)

| # | Line | Current | Target |
|---|---|---|---|
| Z1 | 32-34 | flags array ends with `'-s[...]:query:'  )` | add `'--store[Non-interactive store path for init]:directory:_files'` |
| Z2 | 40-41 | `first) compadd -- "$tags[@]" check` | `compadd -- "$tags[@]" check init` |
| Z3 | 44-48 | `rest) if (( ${words[(I)check]} )); then _message 'check takes no further arguments'` | `if (( ${words[(I)check]} \|\| ${words[(I)init]} )); then _message 'subcommand takes no tag arguments'` |

### fish (`completions/skilldozer.fish`)

| # | Line | Current | Target |
|---|---|---|---|
| F1 | after 32 | (no --store line) | add `complete -c skilldozer -l store -d 'Non-interactive store path for init' -r` |
| F2 | after 35 | (no init line) | add `complete -c skilldozer -n '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'` |
| F3 | 39-41 | `__fish_seen_subcommand_from check; and not __fish_prev_arg_in --search -s` | `__fish_seen_subcommand_from check init; and not __fish_prev_arg_in --search -s` |

---

## §5. Shell quirks / gotchas (per shell)

### bash
- `compgen -d -- "$cur"` completes directory names (standard dir-completion idiom).
  It word-splits into `COMPREPLY=( $(...) )`, reusing the file's EXISTING
  SC2207-acceptable style (tags/flags never contain spaces). Directory names CAN
  contain spaces, so the robust form is `mapfile -t COMPREPLY < <(compgen -d -- "$cur")`.
  Either is acceptable; the simple form matches the file's prevailing convention.
  ShellCheck baseline already shows 2× SC2207 (lines 36, 62) — do not ADD new
  warning classes; SC2207 on the new line is consistent with the existing two.
- The flag-completion branch (`if [[ "$cur" == -* ]]`) runs BEFORE the walk loop, so
  `skilldozer init --<TAB>` correctly offers the flag matrix (incl. `--store`)
  regardless of the walk-loop suppression. The walk-loop `return 0` only affects
  POSITIONAL candidates (tags / re-offering subcommands) — flag completion is unaffected.

### zsh
- `_arguments` value-taking marker: `'--store[desc]:directory:_files'`. The third
  colon-field (`_files`) is the completion function for the value slot. `_files`
  completes files + dirs; `_path_files -/` restricts to dirs only (acceptable
  alternative). `:query:` (used by --search) is a bare placeholder with NO function
  → no completion offered for the search value, which is correct (free-text).
- ShellCheck is bash-only and FALSE-POSITIVES on zsh syntax (`SC2296`, `SC1087`,
  `SC2128`, `SC2154`, `SC2206` on the current `_skilldozer`). The REAL zsh check is
  `zsh -n completions/_skilldozer` (currently: OK). Do NOT "fix" shellcheck output on
  the zsh file — it would break the zsh code.

### fish
- `-r` / `--require-parameter` makes the option take a mandatory argument AND
  switches fish 4.x into "complete the option's value" mode, which BYPASSES the
  global `complete -c skilldozer -f` (no_files) and offers file/dir paths for the
  value. The existing file DELIBERATELY omits `-r` on `--search/-s` (a free-text
  query should offer nothing). For `--store`, we WANT path completion for the value,
  so `-r` is exactly right — the inverse rationale. A comment must explain WHY
  `--store` uses `-r` so a future reader is not confused by the two opposite
  `-r` decisions in the same file.
- `__fish_is_first_arg` restricts `init`/`check` to the first positional (matches the
  contract's "offered only as the first positional"). `__fish_seen_subcommand_from
  check init` suppresses tags once EITHER exclusive subcommand is seen.

---

## §6. Validation tooling available (verified present)

```
bash       5.3.15(1)-release
zsh        5.9.1
fish       4.7.1
shellcheck (installed)
```
- `bash -n completions/skilldozer.bash` → currently OK
- `zsh  -n completions/_skilldozer`      → currently OK
- `fish -n completions/skilldozer.fish`  → currently OK
- `shellcheck -x completions/skilldozer.bash` → 2× SC2207 (pre-existing, acknowledged)
- `shellcheck -x completions/skilldozer.fish` → SC2148 (shebang tip; harmless for a sourced completion file)

There is NO Go test, Makefile, or CI that exercises the completion files
(`grep -rln "compgen\|compadd\|_skilldozer\|__fish" --include="*.go"` → none).
Validation is therefore: syntax checks + structural grep + the programmatic bash
harness (Level 3). The contract explicitly permits "a structural grep asserting
init + --store present and check still present" as the smoke check.

---

## §7. Scope boundary (what NOT to touch)

- DO NOT edit `main.go` (parseArgs already supports `init`/`--store`; it is the
  LOCKSTEP source the completions must MATCH). main.go is owned by P1.M2 (complete).
- DO NOT edit `skills/example/SKILL.md` (owned by the parallel P1.M3.T1.S1 PRP).
- DO NOT edit `README.md` (owned by P1.M4.T2.S1; README already documents sourcing
  the completions — Mode A says no separate doc subtask).
- DO NOT change the tag source `skilldozer --relative --all` to literal `--all`
  (drift doc §4d: `--relative` is a defensible/better deviation; out of scope).
- DO NOT add `init <dir>` bare-positional directory completion — the contract scopes
  dir-completion to `--store` only; `init` is treated like `check` (suppress further
  positionals). The discoverable path for directory completion is `init --store <dir>`.

---

## §8. Relationship to the parallel P1.M3.T1.S1 PRP

P1.M3.T1.S1 rewrites `skills/example/SKILL.md` (skpp → skilldozer). It touches a
DIFFERENT file set (one Markdown asset). There is ZERO file overlap with this
subtask (the three completion files). No merge-conflict risk. This PRP's behavioral
test `skilldozer <TAB>` → `[example]` relies on the `example` skill existing (it
does regardless of T1.S1's rename, since the DIRECTORY name `example` is unchanged).
