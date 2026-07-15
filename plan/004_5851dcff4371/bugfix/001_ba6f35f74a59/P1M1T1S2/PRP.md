# PRP — P1.M1.T1.S2: Add `--shell` value entry + flag advertisement to the zsh completion file

> **Subtask:** P1.M1.T1.S2 — the zsh third of P1.M1.T1 (Issue 1: `--shell` value completion offers skill tags instead of `bash zsh fish`). Mirrors the already-implemented S1 (bash) for the zsh file.
> **Scope boundary:** Edits ONLY `completions/_skilldozer` (the zsh autoload completion script, embedded via `//go:embed main.go:57`). Does NOT touch the bash file (S1, already done in the working tree) or the fish file (S3); does NOT touch any `.go` file (the `//go:embed` picks up the edit automatically); does NOT change `usageText` (it already documents `--shell`); does NOT edit the README (Mode B sweep is P1.M3.T1).

---

## Goal

**Feature Goal**: Make `skilldozer --shell <TAB>` offer exactly `bash zsh fish` (PRD §14.2's fixed enum, nothing else) instead of skill tags, and make `--shell` discoverable via `skilldozer --<TAB>` — in the **zsh** completion script. This is the zsh-specific slice of Issue 1's three-file lockstep fix.

**Deliverable**: Edits to `completions/_skilldozer` only (no new files):
1. **Value entry** (~line 47, in the `flags=( ... )` array): add `'--shell[Force a shell for completion]:shell:(bash zsh fish)'` immediately after the `--completions` entry.
2. **Inline comment** above the entry: document the `:shell:(bash zsh fish)` enum-routing syntax (mirrors the existing `:query:` / `:directory:_files` explanatory comments).
3. **LOCKSTEP header note** (~line 16): append the `--shell` note (mirroring S1's bash header lines 17-18 verbatim).

**Success Definition**: After `go build`, `TestEmbeddedCompletionsMatchOnDisk` passes (embedded zsh bytes == on-disk file); no existing test regresses; in real zsh, `skilldozer --shell <TAB>` offers `bash zsh fish` (not skill tags) and `skilldozer --<TAB>` offers `--shell`; `go.mod`/`go.sum` unchanged; no `.go` file edited.

---

## User Persona (if applicable)

**Target User**: zsh users who tab-complete `skilldozer` invocations (especially the `skilldozer --completions --shell <shell> | source` / `eval "$(skilldozer --completions)"` install idiom).

**Use Case**: A user types `skilldozer --shell ` and tabs to pick the shell for which to emit completions.

**Pain Points Addressed**: Today tab after `--shell` offers skill tags (the opposite of §14.2's "nothing else"), and `--shell` is not offered after `--`, so users must know the flag exists.

---

## Why

- **PRD §14.2**: "`--shell` takes a fixed enum (`bash`/`zsh`/`fish`); offer those three words, nothing else." Today the zsh file has no `--shell` entry in its `_arguments` array, so the value slot falls through to the `*: :->args` catch-all → `compadd $tags` → skill tags. A spec deviation.
- **PRD §14.4 lockstep**: completion files are frozen to `main.go parseArgs()`, which accepts `--shell`. The zsh file is missing it.
- **Decision D7** (decisions.md): `--shell` is a real, documented flag (in `usageText` OPTIONS, used in the canonical install idiom). Add it to the advertised flag list in all three files for discoverability. (PRD §14.6's 13-flag table omits `--shell` — a noted tension; D7 resolves it in favor of consistency.)
- **Lockstep with S1**: the bash file already has the `--shell` handling (S1, working tree). This PRP brings the zsh file into the same state so the two match; the §14.4 "all three identical" invariant is fully restored when S3 (fish) lands.

---

## What

`completions/_skilldozer` gains `--shell` handling via the data-driven `_arguments` array:

- **Value entry**: a new array element `'--shell[Force a shell for completion]:shell:(bash zsh fish)'`. The `:shell:(bash zsh fish)` value-action tells `_arguments` the value slot offers the closed enum — and bypasses the `*: :->args` positional catch-all (so no skill tags leak in).
- **Advertisement**: because `--shell` is now in the `flags=( ... )` array, `_arguments` automatically offers it for `skilldozer --<TAB>` (D7). No separate advertisement code needed — the array IS the flag list.

No behavior change for any other flag; the tag-completion default (`*: :->args` → `compadd $tags`), the `--search`/`--store`/`--init` routing, the long-form-only policy, and the trailing `_skilldozer "$@"` autoload self-call are all untouched.

### Success Criteria

- [ ] the `flags=( ... )` array contains `'--shell[Force a shell for completion]:shell:(bash zsh fish)'` immediately after the `--completions` entry
- [ ] an inline comment above the entry documents the `:shell:(bash zsh fish)` enum-routing (mirrors the `:query:` / `:directory:_files` comment style)
- [ ] the LOCKSTEP header (lines 11-16) ends with the `--shell` note (verbatim mirror of S1's bash header lines 17-18)
- [ ] `TestEmbeddedCompletionsMatchOnDisk` passes (embedded zsh == on-disk file)
- [ ] in real zsh, `--shell <TAB>` offers `bash zsh fish` (not skill tags); `--<TAB>` offers `--shell`
- [ ] no existing test regresses; `go.mod`/`go.sum` unchanged; no `.go` file edited

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact current text of the edit site (the flags array end + the LOCKSTEP header), the exact target additions, the `_arguments` routing mechanic (why the array entry alone suffices), the eval-safe `zshEvalScript` interaction (why my edit region is safe), the embed/rebuild mechanics, the automated gate (TestEmbeddedCompletionsMatchOnDisk), the manual repro, and the scope boundary (zsh only) are all specified with line numbers and exact before/after text. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the parallel sibling PRP (S1, bash) — the pattern to mirror for zsh
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M1T1S1/PRP.md
  why: "S1 (bash) is ALREADY IMPLEMENTED in the working tree (M completions/skilldozer.bash). Its header note (lines 17-18) and value-routing doc are the LOCKSTEP template for the zsh file. S2 mirrors S1's --shell note verbatim and follows the same embed/rebuild/test stance."
  pattern: "S1 added: (1) the value routing, (2) the flag advertisement, (3) doc-comment accuracy, (4) a LOCKSTEP header note. S2 does the zsh equivalents — but zsh is DATA-DRIVEN: a single array entry does BOTH routing AND advertisement (no separate case block / flag list)."

# MUST READ — the authoritative current text + exact old→new strings (verified against live HEAD 6fb3f7e)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M1T1S2/research/verified_facts.md
  why: "§0 the lockstep state (S1 bash done-uncommitted; zsh/fish clean). §1 the exact current text of the flags array (47-48), the _arguments/catch-all/self-call (50/52/62), and the LOCKSTEP header (11-16). §2 the exact edits (array entry + header note). §3 WHY one array entry suffices (data-driven routing). §4 the eval-safe zshEvalScript interaction (my edit is safe). §5/§6 the embed gate + no-test-breaks."
  critical: "§3 (the _arguments value-action mechanic) and §4 (zshEvalScript only strips the trailing self-call) are the two facts that prevent the most likely implementation errors (over-editing, or fearing the eval-safe wrapper)."

# MUST READ — the issue writeup + the decision
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/issue_analysis.md
  why: "§Issue 1 is the authoritative bug writeup: the repro (zsh no --shell entry → skill tags), confirmation --shell is MISSING from all three files, and the per-file fix prescription (the zsh _arguments entry: '--shell[Force a shell for completion]:shell:(bash zsh fish)')."
  section: "Issue 1 (--shell value completion offers skill tags)"
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/decisions.md
  why: "§D7 is the decision to ADD --shell to the advertised flag list (not just value routing) — resolves the PRD §14.6 (13-flag table) vs §14.2 (value enum) tension in favor of consistency/discoverability."
  section: "D7 (Issue 1 — --shell Added to Flag Advertisement List)"

# MUST READ — the edit site (the ONLY file S2 touches)
- file: completions/_skilldozer
  why: "THE edit site. flags=( ... ) array ends at line 48 ()); --completions is the LAST entry (line 47) — --shell goes right after it. _arguments call @ :50 (NO edit). case/state/compadd @ :52-56 (NO edit). trailing _skilldozer \"\\$@\" @ :62 (NO edit — stripped by zshEvalScript on emit). LOCKSTEP header @ :11-16 (append the --shell note)."
  pattern: "Mirror the existing value-taking entries' shape: '--search[desc]:query:' (free-text), '--store[desc]:directory:_files' (path). --shell is the THIRD pattern: '--shell[desc]:shell:(bash zsh fish)' (closed enum). Match the existing inline-comment style (a # block explaining the :syntax:)."
  gotcha: "The entry is single-quoted; parens/spaces inside are literal (no shell expansion) — exactly like the existing quoted entries. Do NOT double-quote it."

- file: main.go
  why: "The //go:embed wiring (NO edit — confirms it picks up the file change). :57 //go:embed completions/_skilldozer; :58 var zshCompletion string; :1118 completionScript(\"zsh\") returns zshCompletion. zshEvalScript (~1126-1167) STRIPS the trailing _skilldozer \"\\$@\" (line 62) on emit for eval-safe sourcing — my edit is in the array body, untouched by the strip."
  pattern: "Do NOT touch main.go. The embed is the mechanism, not an edit site."
  gotcha: "A PRE-BUILT binary holds the OLD embedded bytes. Always rebuild (go build / go test) before behavioral testing."

- file: main_test.go
  why: "The automated gate: TestEmbeddedCompletionsMatchOnDisk @ :2995 reads the on-disk file via os.ReadFile and asserts completionScript(\"zsh\") == on-disk (byte identity, PRD §14.6). go test re-embeds at compile, so it passes automatically after the edit. No test asserts the flags-array content, so no test breaks. The eval-safe strip test (~3117) asserts zshEvalScript strips line 62 — my edit doesn't touch line 62, so it passes."
  pattern: "The embed-match test is the automated regression gate; the zsh repro (Level 3) is the behavioral gate (requires real zsh)."

- url: https://zsh.sourceforge.io/Doc/Release/Completion-Functions.html#index-_005farguments
  why: "Documents the _arguments value-action syntax: an optspec of the form '--opt[desc]:msg:action' where action '(w1 w2 w3)' is a closed word-list group offered for the value slot. This is the syntax behind :shell:(bash zsh fish)."
- url: (PRD §14.2 / §14.4 — in PRD.md, READ-ONLY)
  why: "§14.2: --shell takes the fixed enum bash/zsh/fish, offer those three words nothing else. §14.4: completion files frozen to parseArgs (lockstep). Do NOT edit PRD.md."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD
6fb3f7e
$ git status --short completions/
 M completions/skilldozer.bash        # S1 (bash) done in working tree, UNCOMMITTED
                                       # completions/_skilldozer (zsh) is CLEAN — S2's input
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK
BUILD_OK / VET_OK
# completions/_skilldozer: 62 lines. flags=( ... ) array @ :27-48 (--completions LAST @ :47).
#   _arguments -C "$flags[@]" '*: :->args' @ :50 (data-driven; NO edit). self-call @ :62.
#   Embedded via //go:embed @ main.go:57 → var zshCompletion (main.go:58).
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# No new files. This subtask edits completions/_skilldozer ONLY (no .go file).
```

### Desired Codebase tree with files to be changed

```bash
completions/_skilldozer   # MODIFY — add --shell array entry (+ inline comment) + LOCKSTEP header note
# main.go / main_test.go / go.mod / go.sum — UNCHANGED (//go:embed picks up the file edit on rebuild)
# completions/skilldozer.bash — UNCHANGED here (S1 owns it; already done)
# completions/skilldozer.fish — UNCHANGED here (S3)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `completions/_skilldozer` | `--shell` array entry (routing + advertisement) + inline comment + LOCKSTEP note | PRD §14.2 + decision D7 |

### Known Gotchas of our codebase & Library Quirks

```bash
# GOTCHA #1 — A pre-built binary holds STALE embedded bytes. //go:embed reads the file
# at COMPILE time. If you edit completions/_skilldozer and run an already-built
# ./skilldozer, --completions emits the OLD script (no --shell). Always rebuild
# (go build, or just go test which compiles) before behavioral testing.

# GOTCHA #2 — The automated gate is byte-identity, not behavior. TestEmbeddedCompletionsMatchOnDisk
# asserts completionScript("zsh") == the on-disk file. It passes as long as you rebuild
# after editing. It does NOT assert the --shell entry exists — so the manual zsh repro
# (Level 3) is the behavioral gate that proves the routing actually works.

# GOTCHA #3 — Rebuild for the embed, but DO NOT edit any .go file. The //go:embed directive
# at main.go:57 is the mechanism; it picks up the file change automatically. Adding --shell
# to the zsh file requires ZERO Go changes. (parseArgs already accepts --shell — that's
# why §14.4 says completions are frozen to parseArgs.)

# GOTCHA #4 — One array entry does BOTH value routing AND advertisement. Unlike bash
# (which needs a separate case block AND a flag-list token), zsh's _arguments reads the
# flags=( ... ) array as the single source of truth: an entry with a :msg:action value-spec
# routes the value slot, AND its presence makes the flag advertised on --<TAB>. So S2's
# ONLY functional edit is the one array line (plus comments). Do NOT add a separate
# "advertisement" mechanism — there is none in zsh; the array IS it.

# GOTCHA #5 — zshEvalScript STRIPS the trailing `_skilldozer "$@"` self-call (line 62) when
# emitting via --completions (the eval-safe fix, committed 9682042). My edit is in the
# flags array (~line 47-48), in the function BODY — untouched by the strip. So --shell is
# present in BOTH the autoload form (cp/autoload users) and the eval-safe form
# (--completions | eval users). Do NOT touch line 62 and do NOT worry about the wrapper.

# GOTCHA #6 — The entry is SINGLE-QUOTED: '--shell[Force a shell for completion]:shell:(bash zsh fish)'
# The parens and spaces inside single quotes are LITERAL (no zsh expansion at array-build
# time; _arguments parses the value-action itself). This mirrors the existing quoted
# entries (--search, --store). Do NOT use double quotes (would risk expansion) and do NOT
# drop the quotes (the [desc] and (words) would break array parsing).

# GOTCHA #7 — The enum order is "bash zsh fish" (PRD §14.2 + the contract + S1's bash file).
# Consistent order across all three shells avoids user surprise. Do NOT reorder.

# GOTCHA #8 — Placement is AFTER --completions (contract LOGIC 3a; --completions is the
# current last array element). This also reads naturally: --completions --shell <name>
# are a paired concept. Do NOT group --shell with --store/--init (the contract is explicit
# about placement after --completions; lockstep with the issue_analysis prescription).

# GOTCHA #9 — SCOPE: edit ONLY the zsh file. bash (completions/skilldozer.bash) is S1
# (done in the working tree); fish (completions/skilldozer.fish) is S3. After S2 alone
# the three files temporarily diverge (bash+zsh have --shell, fish doesn't). That's
# EXPECTED — §14.4 lockstep is restored when S3 lands. There is NO cross-file lockstep
# test (TestEmbeddedCompletionsMatchOnDisk compares each embed to its OWN file), so editing
# only zsh breaks no test. Do NOT edit bash/fish here, and do NOT commit S1's bash change
# (that's S1's job).

# GOTCHA #10 — Do NOT change usageText. It already lists --shell in OPTIONS (D7: "--shell is
# a real, documented flag in usageText OPTIONS"). The gap is ONLY the completion file.

# GOTCHA #11 — No deps change. No .go file is edited, so go.mod/go.sum are byte-for-byte
# identical. The sole edited file is a shell data asset.
```

---

## Implementation Blueprint

### Data models and structure

None. This subtask edits a zsh completion script (a data asset embedded into the Go binary). No Go types, no signatures change.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: ADD the --shell array entry + inline comment (completions/_skilldozer, ~line 47-48)
  - EDIT the flags=( ... ) array: AFTER the '--completions[...]' entry and BEFORE the closing ')',
    ADD (entry first, then the comment can precede it — see exact block below):
        # `:shell:(bash zsh fish)` routes --shell's value slot to a FIXED enum — offer
        # exactly the three supported shells, nothing else (PRD §14.2). The third
        # value-routing pattern: --search = free-text (nothing), --store/--init = path
        # (_files), --shell = closed enum. --shell is advertised (decision D7).
        '--shell[Force a shell for completion]:shell:(bash zsh fish)'
  - WHY: one array entry does BOTH value routing (the :shell:(bash zsh fish) value-action)
    AND advertisement (the flag is now in the array → offered on --<TAB>). (GOTCHA #4)
  - EXACT oldText/newText in Implementation Patterns below. Entry is SINGLE-QUOTED (GOTCHA #6).

Task 2: APPEND the --shell note to the LOCKSTEP header (~line 16)
  - EDIT the header: AFTER the "belongs to skill tags — a bare <tab> shows skills, never
    commands." line, APPEND (verbatim mirror of S1's bash header lines 17-18):
        # --shell's value completes to the bash/zsh/fish enum (§14.2); --shell is
        # advertised (D7) since it is a real, documented flag in usageText OPTIONS.
  - WHY: contract LOGIC 3b (keep LOCKSTEP header intact) + DOCS §5 (verify it mentions
    --shell). Mirroring S1 verbatim keeps the three files' headers in lockstep.

Task 3: VERIFY — embed-match test + behavioral repro + no regression
  - COMMAND: go build ./...                                  (exit 0; re-embeds the edited file)
  - COMMAND: go test -run TestEmbeddedCompletionsMatchOnDisk -v ./...   (PASS — embedded zsh == on-disk)
  - COMMAND: go test ./...                                   (no regression — zsh content tests still pass)
  - MANUAL (requires real zsh): the repro (Level 3) → --shell <TAB> offers bash zsh fish
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"   (GOTCHA #11)
  - COMMAND: git diff --stat -- '*.go'                       (MUST be empty — no .go file edited)
  - COMMAND: git diff --name-only                            (MUST list ONLY completions/_skilldozer;
                                                              do NOT also touch bash/fish — GOTCHA #9)
```

### Implementation Patterns & Key Details

```zsh
# Task 1 — the array entry (exact oldText → newText). The entry is single-quoted; the
# inline comment explains the third value-routing pattern (enum) vs the existing two
# (free-text, path). Placement: immediately AFTER --completions (the current last element).

#   OLD (lines 47-48):
        '--completions[Emit the shell completion script for eval]'
    )
#   NEW:
        '--completions[Emit the shell completion script for eval]'
        # `:shell:(bash zsh fish)` routes --shell's value slot to a FIXED enum —
        # offer exactly the three supported shells, nothing else (PRD §14.2). The
        # third value-routing pattern: --search = free-text (nothing), --store/--init
        # = path (_files), --shell = closed enum. --shell is advertised (decision D7).
        '--shell[Force a shell for completion]:shell:(bash zsh fish)'
    )

# Task 2 — the LOCKSTEP header note (exact oldText → newText). Append the verbatim S1
# mirror after the decision-19 paragraph.

#   OLD (the last header line, line 16):
# belongs to skill tags — a bare <tab> shows skills, never commands.
#   NEW (append two lines):
# belongs to skill tags — a bare <tab> shows skills, never commands.
# --shell's value completes to the bash/zsh/fish enum (§14.2); --shell is
# advertised (D7) since it is a real, documented flag in usageText OPTIONS.
```

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **One array entry does both routing + advertisement (no separate mechanism).** Unlike bash (which needs a `case "$prev"` branch AND a `compgen -W` flag-list token), zsh `_arguments` treats the `flags=( ... )` array as the single source of truth: the entry's `:shell:(bash zsh fish)` value-action routes the value slot, and the entry's mere presence advertises the flag on `--<TAB>`. So S2's ONLY functional edit is the one array line — there is no separate "flag list" to update. This is the zsh-specific simplification of S1's two-edit bash approach.

2. **Placement after `--completions` (contract-pinned).** The contract LOGIC 3a says place after the `--completions` entry. This also reads naturally (`--completions --shell <name>` are a paired concept). Do not regroup `--shell` with `--store`/`--init` even though all three are value-taking — the contract is explicit, and lockstep with the issue_analysis prescription matters.

3. **Header note mirrors S1 verbatim.** S1's bash header (lines 17-18) says exactly *"--shell's value completes to the bash/zsh/fish enum (§14.2); --shell is advertised (D7) since it is a real, documented flag in usageText OPTIONS."* Copy it verbatim into the zsh header so the three files' headers stay in lockstep (wording divergence would be pointless review noise).

### Integration Points

```yaml
EMBED (the mechanism — NO edit):
  - main.go:57 //go:embed completions/_skilldozer  →  var zshCompletion string (main.go:58)
  - completionScript("zsh") returns zshCompletion (main.go:1118); runCompletion/zshEvalScript emit it.
  - Editing the on-disk file + go build/go test re-embeds the new bytes automatically.

EVAL-SAFE WRAPPER (NO edit — interaction verified safe):
  - zshEvalScript (~main.go:1126-1167) STRIPS the trailing `_skilldozer "$@"` (line 62) for
    eval-safe sourcing. My edit is in the flags array (~47-48), untouched by the strip.
  - So --shell is present in BOTH autoload form (cp/autoload) and eval-safe form (--completions|eval).

TESTS (unchanged; they gate the change):
  - TestEmbeddedCompletionsMatchOnDisk (main_test.go:2995): embedded zsh == on-disk. PASS after rebuild.
  - TestCompletionScript (2958) / TestRunCompletionZshScript (~3089) / TestRunCompletionShellFromEnv
    (~3105): assert the `#compdef skilldozer` header — line 1 untouched. Still pass.
  - eval-safe strip test (~3117): asserts zshEvalScript strips line 62 — line 62 untouched. Passes.

NO DATABASE / NO CONFIG / NO ROUTES / NO GO SOURCE:
  - This subtask edits exactly one shell data asset. No parseArgs, no run(), no usageText, no main.go.
```

---

## Validation Loop

### Level 1: Shell sanity (immediate, after the edits)

```bash
cd /home/dustin/projects/skilldozer

# zsh syntax-check the edited file (catches a broken array/quoting/parens):
if command -v zsh >/dev/null 2>&1; then
  zsh -n completions/_skilldozer && echo "zsh -n OK" || echo "FAIL: zsh syntax error"
else
  echo "zsh not installed — skipping zsh -n (the embed-match test still gates byte-identity)"
fi
# Expected: "zsh -n OK" (or the skip message if zsh is absent).

# Confirm the two edits are present:
grep -n -- "--shell\[Force a shell" completions/_skilldozer          # Expected: 1 hit (the new array entry)
grep -c -- '--shell' completions/_skilldozer                         # Expected: >=3 (entry + comment + header note)
grep -n -- 'advertised (D7)' completions/_skilldozer                 # Expected: 1 hit (header note)
# Expected: zsh -n clean (or skipped); the greps find the additions.
```

### Level 2: The embed-match gate (the automated regression check)

```bash
cd /home/dustin/projects/skilldozer

go build ./...     ; echo "build exit $?"    # Expected: 0 (re-embeds the edited file)
go test -run TestEmbeddedCompletionsMatchOnDisk -v ./...
# Expected: PASS — completionScript("zsh") == on-disk completions/_skilldozer.

# Whole module: the zsh content tests + everything else still green:
go test ./... ; echo "test exit $?"          # Expected: 0 (no regression)
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
git diff --stat -- '*.go'                    # Expected: EMPTY (no .go file edited)
git diff --name-only                         # Expected: completions/_skilldozer (and S1's bash file, already modified — do NOT touch fish)
# Expected: build/test exit 0; deps unchanged; no .go diff; only the zsh file newly changed by S2.
```

### Level 3: The behavioral repro (the actual contract — Issue 1 fixed; requires real zsh)

```bash
cd /home/dustin/projects/skilldozer
command -v zsh >/dev/null 2>&1 || { echo "SKIP: zsh not installed (the embed-match test is the fallback gate)"; exit 0; }
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# Set up a fpath dir with the autoload function pointing at the BUILT script (so the
# embedded --shell entry is exercised, not a stale on-disk copy):
ZFP=$(mktemp -d)
/tmp/sdz --completions --shell zsh > "$ZFP/_skilldozer"

# Issue 1 repro: after --shell, offer exactly bash zsh fish (NOT skill tags).
# `compadd -O` would need a real completer; instead drive _arguments directly via zpty
# or use the simpler `compadd` inspection. The most robust in-process check: source the
# function and inspect the generated _arguments spec parses (zsh -n already did); for a
# true behavioral check, drive the completion state machine:
zsh -c "
  fpath=($ZFP \$fpath); autoload -Uz _skilldozer; compinit -u 2>/dev/null
  # Use zpty-free introspection: ask zsh's completion for the candidates after --shell
  zmodload zsh/zselect 2>/dev/null
  # The _arguments spec's value-action for --shell must be the closed enum:
  typeset -a spec; _skilldozer 2>/dev/null  # loads the function
  # Direct check: the array literal must contain the --shell entry with (bash zsh fish)
"
grep -q -- "--shell\[Force a shell for completion\]:shell:(bash zsh fish)" "$ZFP/_skilldozer" \
  && echo "--shell enum entry present in emitted script OK" || echo "FAIL: --shell entry missing"

# Advertisement: the emitted script must list --shell in the flags array (offered on --<TAB>).
grep -q -- "'--shell\[" "$ZFP/_skilldozer" && echo "--shell advertised OK" || echo "FAIL: --shell not advertised"

# Full behavioral check (interactive state machine) — most reliable via a real zsh session:
#   run:  skilldozer --shell <press TAB>  → expect bash zsh fish
zsh -ic 'echo "fpath+=('$ZFP'); autoload -Uz compinit; compinit" >/dev/null' 2>/dev/null
# (If a headless interactive TAB check is needed, use `zsh/zpty`; the grep checks above are
#  the deterministic, CI-friendly proof that the enum entry + advertisement are in the bytes
#  the binary actually emits.)
rm -rf "$ZFP" /tmp/sdz
# Expected: "--shell enum entry present" + "--shell advertised OK".
```

> **Note on Level 3:** the deterministic proof is the grep against the EMITTED script
> (`/tmp/sdz --completions --shell zsh` output) — it confirms the rebuilt binary embeds
> the `--shell` array entry with the closed enum. A full interactive `zpty`-driven TAB
> simulation is the strongest behavioral check but is fragile in CI; the grep on the
> emitted bytes is the reliable gate. (zsh `-n` in Level 1 confirms the file parses.)

### Level 4: Lockstep-awareness check (scope discipline)

```bash
cd /home/dustin/projects/skilldozer

# S2 edits ONLY the zsh file. Confirm S2 did NOT touch fish (S3) and did NOT edit any .go file.
git diff --name-only | grep -vE 'skilldozer\.bash'   # (bash is S1's working-tree change, not S2's)
# Expected: ONLY completions/_skilldozer from S2. (If fish appears, you over-reached into S3.)
git diff --stat -- '*.go'                             # Expected: EMPTY

# The embed-match test passes for all three shells (zsh now matches the edited file;
# bash matches S1's edit; fish unchanged):
go test -run TestEmbeddedCompletionsMatchOnDisk -v ./... 2>&1 | grep -E 'zsh|fish|bash|PASS|FAIL'
# Expected: PASS (no per-shell breakdown in output, but the test loop covers all three).
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `zsh -n` clean (or skipped if zsh absent); greps confirm the `--shell` array entry, the inline comment, and the header note
- [ ] Level 2 PASS — `go build` exit 0; `TestEmbeddedCompletionsMatchOnDisk` PASS; `go test ./...` exit 0; `git diff go.mod go.sum` → "deps unchanged"; no `.go` diff
- [ ] Level 3 PASS — the emitted `--completions` script contains `'--shell[Force a shell for completion]:shell:(bash zsh fish)'` and the `--shell[` advertisement (deterministic grep on rebuilt binary output)
- [ ] Level 4 PASS — only `completions/_skilldozer` newly changed by S2; no fish touch; no `.go` touch; embed-match passes for all three shells

### Feature Validation
- [ ] the `flags=( ... )` array contains `'--shell[Force a shell for completion]:shell:(bash zsh fish)'` immediately after `--completions`
- [ ] an inline comment documents the `:shell:(bash zsh fish)` enum-routing (third value pattern)
- [ ] the LOCKSTEP header ends with the verbatim S1 mirror `--shell` note
- [ ] the enum is exactly `bash zsh fish` (order per §14.2)
- [ ] `--shell` is advertised (present in the flags array → offered on `--<TAB>`)

### Code Quality / Convention Validation
- [ ] the `--shell` entry mirrors the existing value-taking entries' shape (`'flag[desc]:msg:action'`, single-quoted)
- [ ] the inline comment mirrors the existing `:query:` / `:directory:_files` explanatory style
- [ ] no Go file edited (the `//go:embed` picks up the change); `go.mod`/`go.sum` byte-for-byte identical
- [ ] minimal diff (one array entry + one comment block + one header note)

### Scope Discipline
- [ ] Did NOT touch `completions/skilldozer.bash` (S1 — already done in working tree) or `completions/skilldozer.fish` (S3)
- [ ] Did NOT edit any `.go` file (main.go //go:embed, parseArgs, run(), usageText, zshEvalScript all unchanged)
- [ ] Did NOT touch the trailing `_skilldozer "$@"` self-call (line 62) or the `_arguments` call (line 50)
- [ ] Did NOT edit the README (Mode B sweep is P1.M3.T1)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't skip the rebuild.** A pre-built binary holds stale embedded bytes. `go build` (or `go test`) re-embeds; an already-built `./skilldozer` does not.
- ❌ **Don't add a separate "advertisement" mechanism.** zsh `_arguments` is data-driven: the array IS the flag list. One entry does BOTH value routing AND advertisement. There is no separate `compgen -W` flag-list to update (that's a bash thing).
- ❌ **Don't edit any `.go` file.** The `//go:embed` is the mechanism; it picks up the file change automatically. Adding `--shell` to the zsh file requires zero Go changes (parseArgs already accepts `--shell`).
- ❌ **Don't touch the `_arguments -C "$flags[@]"` call (line 50), the `case "$state"` block, or the trailing `_skilldozer "$@"` (line 62).** The routing is automatic once the entry is in the array; the self-call is stripped by `zshEvalScript` and must stay for autoload users.
- ❌ **Don't use double quotes or drop quotes on the entry.** It MUST be single-quoted: the `[desc]` and `(bash zsh fish)` parens/spaces are literal inside single quotes. Double quotes risk expansion; no quotes break array parsing.
- ❌ **Don't reorder the enum.** Use `bash zsh fish` (PRD §14.2 + S1's bash file). A different order isn't wrong functionally but diverges from the spec and the other shells.
- ❌ **Don't place `--shell` anywhere but after `--completions`.** The contract pins placement; lockstep with the issue_analysis prescription matters. (It also reads naturally: `--completions --shell <name>` are paired.)
- ❌ **Don't edit fish or bash here.** bash is S1 (done); fish is S3. The temporary cross-file divergence is expected and resolved when S3 lands. There is NO cross-file lockstep test, so editing only zsh breaks nothing.
- ❌ **Don't change usageText.** It already documents `--shell` (D7). The gap is only the completion file.
- ❌ **Don't add deps.** No `.go` file is edited; the sole change is a shell data asset.
- ❌ **Don't fear the eval-safe wrapper.** `zshEvalScript` strips only the trailing self-call (line 62); your edit in the array body (~47-48) is untouched, so `--shell` works in both autoload and eval-safe forms.

---

## Confidence Score

**9.5/10** — Every edit is pinned to the exact current (HEAD `6fb3f7e`) text with before/after blocks: the array entry (mirroring the existing value-taking entries' `'flag[desc]:msg:action'` shape), the inline comment (mirroring the `:query:`/`:directory:_files` explanatory style), and the LOCKSTEP header note (verbatim mirror of S1's bash header). The two subtleties that matter most are proven in `research/verified_facts.md`: (§3) one array entry does both routing and advertisement because zsh `_arguments` is data-driven, and (§4) `zshEvalScript` strips only the trailing self-call (line 62) so my edit region is safe. The embed/rebuild mechanic is confirmed (TestEmbeddedCompletionsMatchOnDisk gates byte-identity), and no existing test asserts the array content so nothing regresses. The 0.5 reservation is the Level 3 behavioral proof: a fully deterministic interactive `zpty`-driven TAB simulation is fragile in CI, so the PRP's primary behavioral gate is the deterministic grep on the rebuilt binary's emitted script (confirming the enum entry + advertisement are in the bytes the binary actually ships), with `zsh -n` as the parse gate. That combination is reliable and sufficient to prove the fix.
