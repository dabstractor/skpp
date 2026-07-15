# Verified Facts — P1.M1.T1.S2 (zsh completion: add `--shell` value entry + advertisement)

Confirmed against the **live** tree at HEAD `6fb3f7e` + the working tree. The parallel
sibling **S1 (bash) is IMPLEMENTED but UNCOMMITTED** (`git status` shows
`M completions/skilldozer.bash` with all four S1 edits present). The zsh file
(`completions/_skilldozer`) is **unchanged** — S2 implements against the current
committed zsh file (read in full; 62 lines).

---

## 0. Lockstep state at S2 start

| File | State | --shell present? |
|---|---|---|
| `completions/skilldozer.bash` | **MODIFIED** (working tree, uncommitted — S1) | YES (case @43, flag-list @49, header @18, doc @38) |
| `completions/_skilldozer` (zsh) | **clean** (HEAD) | NO — **this is S2's input** |
| `completions/skilldozer.fish` | clean (HEAD) | NO — S3's input |

There is **no cross-file lockstep test** (TestEmbeddedCompletionsMatchOnDisk compares
each embedded var to its OWN on-disk file, not to each other). So editing only the zsh
file breaks no test; the §14.4 "all three identical" lockstep is restored when S3 lands.
`go build` / `go vet` PASS today.

---

## 1. The CURRENT (HEAD `6fb3f7e`) zsh file — exact text of the edit sites

File: `completions/_skilldozer` (62 lines). Read in full.

### The flags array (lines 27-48) — the `--shell` entry goes right before the closing `)`
```zsh
    local -a flags=(
        '--version[Print the skilldozer version]'
        '--help[Show this help message]'
        '--path[Print the resolved skills directory]'
        '--list[Human-readable catalog (TAG, NAME, DESCRIPTION)]'
        '--all[Print every skill directory path, sorted by tag]'
        '--file[Print the SKILL.md path instead of the directory]'
        '--relative[Print paths relative to the skills directory]'
        '--no-color[Disable ANSI color]'
        # `:query:` marks --search as value-taking ... (free-text, no completion offered)
        '--search[Substring search over tag/name/description/keywords]:query:'
        # `:directory:_files` routes --store's and --init's value slots to file/path completion
        '--store[Non-interactive store path for init]:directory:_files'
        # Decision 19: check/init/completions promoted ... decision 20: no short aliases advertised.
        '--check[Validate every skill on disk]'
        '--init[First-run setup: pick/create the skills store]:directory:_files'
        '--completions[Emit the shell completion script for eval]'          # [47] <-- --shell goes AFTER this
    )                                                                       # [48] <-- closing paren
```
**The `--completions` entry (line 47) is the LAST array element.** The contract pins
placement: add `--shell` immediately after `--completions` (between line 47 and the `)`).

### The _arguments call + catch-all + trailing self-call (lines 50, 52, 62)
```zsh
    _arguments -C "$flags[@]" '*: :->args' && return 0                      # [50] routing (NO edit)
    case "$state" in                                                        # [52]
        args) compadd -- "$tags[@]" ;;                                      # tags catch-all
    esac
}
_skilldozer "$@"                                                           # [62] self-call (stripped by zshEvalScript)
```

### The LOCKSTEP header (lines 11-16) — append a --shell note (mirror S1's bash file)
```zsh
# LOCKSTEP: the flag list below is frozen to `main.go parseArgs()`. If a future
# task adds/renames a flag there, update this list — and the bash/fish files —
# identically. There is no shared source of truth the shells can import.
# Flags are long-form-only (decision 20): short aliases stay valid at runtime
# but are not advertised. Updated for --check/--init/--completions (decision 19):
# these were promoted from bare subcommands so the bare positional namespace
# belongs to skill tags — a bare <tab> shows skills, never commands.
```
S1's bash file appended (lines 17-18): `# --shell's value completes to the bash/zsh/fish
enum (§14.2); --shell is advertised (D7) since it is a real, documented flag in usageText
OPTIONS.` **Mirror this verbatim in the zsh file** (after line 16) for lockstep consistency.

---

## 2. The exact edit (old → NEW)

### Edit A — the flags array entry (the value routing; contract LOGIC 3a)
OLD (lines 47-48):
```zsh
        '--completions[Emit the shell completion script for eval]'
    )
```
NEW:
```zsh
        '--completions[Emit the shell completion script for eval]'
        # `:shell:(bash zsh fish)` routes --shell's value slot to a FIXED enum —
        # offer exactly the three supported shells, nothing else (PRD §14.2). The
        # third value-routing pattern: --search = free-text (nothing), --store/--init
        # = path (_files), --shell = closed enum. --shell is advertised (decision D7).
        '--shell[Force a shell for completion]:shell:(bash zsh fish)'
    )
```
**The entry is contract-pinned verbatim:** `'--shell[Force a shell for completion]:shell:(bash zsh fish)'`
(zsh `_arguments` value-action syntax: `flag[desc]:message:(word1 word2 word3)` → the
parenthesized list is the closed enum offered for the value slot).

### Edit B — the LOCKSTEP header note (contract LOGIC 3b + DOCS §5)
Append (after line 16, mirroring S1's bash header lines 17-18):
```zsh
# --shell's value completes to the bash/zsh/fish enum (§14.2); --shell is
# advertised (D7) since it is a real, documented flag in usageText OPTIONS.
```

**Nothing else changes.** No edit to the `_arguments -C` call (line 50), the `case
"$state"` block (52-56), the `compadd` (54), or the trailing `_skilldozer "$@"` (62).

---

## 3. Why adding the array entry is SUFFICIENT (the _arguments routing mechanic)

zsh `_arguments` reads the `flags=( ... )` array and, for each entry, builds the
completion state machine. An entry with a `:message:action` value-spec (like
`:shell:(bash zsh fish)`) tells _arguments: "when the cursor is in THIS flag's value
slot, offer `action` (here the closed enum) — and do NOT fall through to the `*: :->args`
positional catch-all." So once `--shell` is in the array:

- `skilldozer --shell <TAB>` → _arguments offers `bash zsh fish` (the enum). ✓
- `skilldozer --shell <TAB>` does NOT offer skill tags (the `*: :->args` catch-all is
  bypassed for the value slot). ✓ (This is the bug fix — pre-fix, no `--shell` entry
  meant the value slot fell through to `->args` → `compadd $tags` → skill tags.)
- `skilldozer --<TAB>` → _arguments offers every flag in the array, now including
  `--shell` (advertisement, D7). ✓

No change to line 50 (`_arguments -C "$flags[@]" '*: :->args'`) or the case block is
needed — the routing is data-driven by the array contents. **This is the contract's
claim: "_arguments handles the routing automatically once the entry is in the array."**

---

## 4. The eval-safe wrapper interaction (zshEvalScript) — my edit is SAFE

The on-disk `completions/_skilldozer` is the AUTOLOAD form: it ends with
`_skilldozer "$@"` (line 62), the standard idiom that fires the function immediately
when sourced via `autoload`. BUT when a zsh user runs
`eval "$(skilldozer --completions)"`, `runCompletion` calls `zshEvalScript`
(main.go ~1126-1167) which **STRIPS that trailing self-call** and wraps the body for
eval-safe sourcing (the fix committed in `9682042`; without it, sourcing the autoload
form in `.zshrc` errors `_skilldozer:31: command not found: _arguments`).

**zshEvalScript only touches the trailing `_skilldozer "$@"` line (62).** My edit is in
the flags array (~lines 47-48), in the BODY of the function — untouched by the strip.
So the `--shell` entry is present in BOTH:
- the autoload form (on-disk file → `cp`/`autoload` users), and
- the eval-safe form (what `--completions` emits → `eval | source` users).

The eval-safe test (`main_test.go` ~3117-3125) asserts the strip happens; it does NOT
assert anything about the flags array, so it keeps passing.

---

## 5. The embed wiring + the automated gate (TestEmbeddedCompletionsMatchOnDisk)

```go
// main.go:57  //go:embed completions/_skilldozer
// main.go:58  var zshCompletion string
// main.go:1118  case "zsh": return zshCompletion, true   // completionScript("zsh")
```
`//go:embed` reads the on-disk file at COMPILE time. Editing the file + `go build` /
`go test` re-embeds the new bytes automatically — **zero Go edits needed.**

`TestEmbeddedCompletionsMatchOnDisk` (main_test.go:2995) reads the on-disk file via
`os.ReadFile(tc.path)` (case `{"zsh", "completions/_skilldozer"}`) and asserts
`completionScript("zsh") == string(onDisk)` (byte-identity, PRD §14.6). Both sides
change together on rebuild → the test PASSES. It does NOT assert the `--shell` entry
exists, so it is a regression gate, not a behavioral gate.

**A pre-built binary holds STALE embedded bytes** — always rebuild before behavioral
testing (the `//go:embed` is compile-time, not runtime).

---

## 6. Test impact — S2 breaks NOTHING

| Test | What it asserts | S2 effect |
|---|---|---|
| `TestEmbeddedCompletionsMatchOnDisk` (2995) | embedded zsh == on-disk file | PASSES (both change together on rebuild) |
| `TestCompletionScript` (2958) | `completionScript("zsh")` starts with `#compdef skilldozer`, non-empty | PASSES (header line 1 untouched) |
| `TestRunCompletionZshScript` (~3089) | `--completions` (SKILLDOZER_SHELL=zsh) stdout has `#compdef` | PASSES (header untouched) |
| `TestRunCompletionShellFromEnv` (~3105) | SHELL=/bin/zsh → zsh header | PASSES |
| eval-safe strip test (~3117) | `zshEvalScript` strips `_skilldozer "$@"` | PASSES (line 62 untouched; my edit is in the array) |

No test asserts the flags-array content, so S2 adds no new red. The behavioral proof
(`--shell <TAB>` → enum) requires real zsh — see PRP Level 3 repro (skipped if zsh absent).

## 7. S2's contract is build + embed-match + no-regression (OUTPUT §4)
Contract OUTPUT §4: "After `go build`, the embedded bytes match (TestEmbeddedCompletionsMatchOnDisk
passes). In zsh, `skilldozer --shell <TAB>` offers bash, zsh, fish (not skill tags)."
Hard gate = `go build` + `go vet` + `TestEmbeddedCompletionsMatchOnDisk` + `go test ./...`.
Behavioral gate = the zsh repro (requires zsh installed). go.mod/go.sum unchanged (no
.go file edited — the sole change is a shell data asset).
