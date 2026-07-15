# Completions Change Map — Delta 004

Exact line numbers verified against HEAD `f30d5c5`. All three files are the
single source of truth, compiled in via `//go:embed` — a **rebuild** is required
for `--completions` to emit the new bytes (§14.6 lockstep).

---

## General principles (apply to ALL three files)

1. **Remove bare subcommand offers:** `check`, `init`, `completion` no longer
   appear as first-positional completion candidates. A bare `<tab>` yields
   skills only.
2. **Add three new long flags:** `--check`, `--init`, `--completions` to each
   file's flag set. No short aliases for these.
3. **Remove short-form flag offers:** Drop `-v`, `-h`, `-p`, `-l`, `-a`, `-f`,
   `-s` from what is OFFERED (they remain valid at runtime — just not advertised).
4. **Keep tag completion byte-identical:** The `skilldozer --relative --all
   2>/dev/null` probe and the mechanism that feeds its output to the shell's
   candidate list must be unchanged.
5. **Route `--init <dir>` to directory completion** (like `--store`).
6. **Route `--completions --shell <name>` to the enum** bash/zsh/fish.
7. **Update LOCKSTEP comments** to cite decisions 19/20 and the long-form-only rule.

---

## 1. `completions/skilldozer.bash` (69 lines)

### Changes:

**Line 41 (flag list — `compgen -W`):**
```bash
# OLD:
"--version -v --help -h --path -p --list -l --all -a --file -f --relative --no-color --search -s --store"
# NEW (long-form only + add --check/--init/--completions):
"--version --help --path --list --all --file --relative --no-color --search --store --check --init --completions"
```

**Line 34 (value-routing):**
```bash
# OLD:
--search|-s) return 0 ;;
--store) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
# NEW (drop -s from search; add --init to dir routing):
--search) return 0 ;;
--store|--init) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
```

**Lines 46-55 (walk-guard):** Remove the `check`/`init`/`completion` bare-word
guard from the loop condition (line 52). These are no longer special tokens.

**Line 63 (subcommand offers):**
```bash
# OLD:
cands="$cands check init completion"
# NEW:
# (delete this line entirely; cands = tags only)
```

**Lines 50-55 (have_pos tracking):** Simplify — without subcommand exclusivity
guards, the walk just needs to know if a positional has been seen (it always
has tags to offer). The entire `have_pos` / subcommand-walk logic can be
simplified to: always offer tags as positionals (skills are never exclusive with
each other — you can pass multiple tags to `skilldozer <tag> [<tag>...]`).

---

## 2. `completions/_skilldozer` (zsh, 61 lines)

### Changes:

**Lines 22-27 (flag array — remove short forms):**
```zsh
# OLD (each line has both --long and -short):
'--version[...]'   '-v[...]'
'--help[...]'      '-h[...]'
'--path[...]'      '-p[...]'
'--list[...]'      '-l[...]'
'--all[...]'       '-a[...]'
'--file[...]'      '-f[...]'
# NEW (long-form only — delete the -x entries):
'--version[Print the skilldozer version]'
'--help[Show this help message]'
'--path[Print the resolved skills directory]'
'--list[Human-readable catalog (TAG, NAME, DESCRIPTION)]'
'--all[Print every skill directory path, sorted by tag]'
'--file[Print the SKILL.md path instead of the directory]'
```

**Line 34 (remove -s short form for search):**
```zsh
# OLD:
'-s[Substring search over tag/name/description/keywords]:query:'
# DELETE THIS LINE (keep line 33: '--search[...]:query:')
```

**Add new flags (before the closing `)` of the flags array, ~line 38):**
```zsh
'--check[Validate every skill on disk]'
'--init[First-run setup: pick/create the skills store]:directory:_files'
'--completions[Emit the shell completion script for eval]'
```

**Lines 43-58 (subcommand offers + state handling):**
```zsh
# OLD: first/rest split offers check/init/completion as first positional
# NEW: simplify — positionals are ALWAYS skills (no subcommand exclusivity)
# Replace the entire case "$state" block with:
'*: :->args'
case "$state" in
    args)
        compadd -- "$tags[@]"
        ;;
esac
```

**Line 19 (tag probe):** KEEP BYTE-IDENTICAL:
```zsh
tags=(${(f)"$(skilldozer --relative --all 2>/dev/null)"})
```

---

## 3. `completions/skilldozer.fish` (52 lines)

### Changes:

**Lines 17-22 (flag matrix — remove `-s` short opts):**
```fish
# OLD:
complete -c skilldozer -s v -l version  ...
complete -c skilldozer -s h -l help     ...
complete -c skilldozer -s p -l path     ...
complete -c skilldozer -s l -l list     ...
complete -c skilldozer -s a -l all      ...
complete -c skilldozer -s f -l file     ...
# NEW (long-form only — drop -s tokens):
complete -c skilldozer -l version  -d 'Print the skilldozer version'
complete -c skilldozer -l help     -d 'Show this help message'
complete -c skilldozer -l path     -d 'Print the resolved skills directory'
complete -c skilldozer -l list     -d 'Human-readable catalog (TAG, NAME, DESCRIPTION)'
complete -c skilldozer -l all      -d 'Print every skill directory path, sorted by tag'
complete -c skilldozer -l file     -d 'Print the SKILL.md path instead of the directory'
```

**Line 32 (remove -s short form for search):**
```fish
# OLD:
complete -c skilldozer -s s -l search -d '...'
# NEW:
complete -c skilldozer -l search -d 'Substring search over tag/name/description/keywords'
```

**Add new flags (after line 24):**
```fish
complete -c skilldozer -l check       -d 'Validate every skill on disk'
complete -c skilldozer -l init        -d 'First-run setup: pick/create the skills store' -r
complete -c skilldozer -l completions -d 'Emit the shell completion script for eval'
```

**Lines 39 (store routing) — add --init to dir completion:**
```fish
complete -c skilldozer -l store -d '...' -r
complete -c skilldozer -l init  -d '...' -r   # (or merge into the -r list)
```

**Lines 43-45 (delete bare subcommand offers):**
```fish
# DELETE ALL THREE:
complete -c skilldozer -n '__fish_is_first_arg' -a 'check' ...
complete -c skilldozer -n '__fish_is_first_arg' -a 'init' ...
complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' ...
```

**Lines 51-52 (tag directive guard):**
```fish
# OLD guard references check/init/completion AND --search -s:
complete -c skilldozer -n 'not __fish_seen_subcommand_from check init completion; and not __fish_prev_arg_in --search -s' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
# NEW (no subcommands to guard against; drop -s from --search ref):
complete -c skilldozer -n 'not __fish_prev_arg_in --search' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
```
NOTE: The `-a '(skilldozer --relative --all 2>/dev/null)'` probe MUST stay byte-identical.
Only the `-n` guard condition changes.

---

## Post-change verification

1. `go build -o skilldozer .` — must succeed (embeds new completion files)
2. `./skilldozer --completions --shell bash 2>/dev/null | grep -q '\-\-completions'` — new flag present
3. `./skilldozer --completions --shell bash 2>/dev/null | grep -q '\-\-check'` — new flag present
4. `! ./skilldozer --completions --shell bash 2>/dev/null | grep -Eq '\-\-version[ ]+-v'` — long-form only
5. `go test ./...` — all tests pass (including `TestEmbeddedCompletionsMatchOnDisk`)
