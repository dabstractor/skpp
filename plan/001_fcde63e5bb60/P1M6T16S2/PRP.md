---
name: "P1.M6.T16.S2 — Sync changeset-level documentation (Mode B final sweep)"
description: >
  Final cross-cutting documentation sync. Depends on all 20 implementing
  subtasks and runs last. Verifies README.md is consistent with the final
  shipped binary and patches the verified drift: undocumented shell
  completions, two missing modifier flags in Usage, and write-tech-docs
  linter hits (8 em dashes). Edits README.md in place. Does NOT create new
  doc files or touch source code.
---

## Goal

**Feature Goal**: `README.md` is fully consistent with the final shipped
`skpp` binary (post §13 acceptance): every flag/subcommand in the binary is
documented, the shipped shell completions are explained, and the prose passes
the `write-tech-docs` linter (exit 0). No stale overview docs ship with the
changeset.

**Deliverable**: An updated `README.md` (in place) that (1) documents the
three shipped completion files with correct per-shell sourcing instructions,
(2) explicitly names the `--relative` and `--no-color` modifier flags in
Usage, and (3) passes `bash scripts/lint.sh README.md` with zero hits.

**Success Definition**: the binary↔README flag diff is empty, the completion
files are documented, the linter exits 0, and `go test ./...` + §13 acceptance
remain green (docs edits did not regress anything).

## Why

- **Mode B contract**: this is the SOW §5 mandatory final docs task. It depends
  on every implementing subtask specifically so it runs last and catches drift
  introduced across M1–M6.
- **Prevents a coherent delta shipping with stale docs** (the SOW's stated
  purpose). The completions were deliberately deferred here by `P1.M6.T15.S1`;
  leaving them undocumented would ship a feature with no user-facing entry
  point in the README.
- **Deterministic quality gate**: the `write-tech-docs` linter turns "is the
  README clean?" from a judgment call into a pass/fail command.

## What

User-visible behavior: the README gains a `## Shell completions` section with
bash/zsh/fish sourcing instructions and a one-line note that completion is
dynamic (`skpp --relative --all`), names `--relative` and `--no-color` in
Usage, and drops all em dashes in favor of colons/periods. No CLI behavior,
exit codes, or file layout change.

### Success Criteria

- [ ] `./skpp --help` flag set ⊆ flags named in README (no binary flag missing
      from the README), and every flag the README names exists in `--help`.
- [ ] README contains a `## Shell completions` section covering bash, zsh, and
      fish with sourcing instructions that match the `completions/*` file
      headers.
- [ ] README Usage names both `--relative` and `--no-color`.
- [ ] `bash /home/dustin/.pi/agent/skills/write-tech-docs/scripts/lint.sh README.md`
      exits 0 (zero hits).
- [ ] `grep -c "—" README.md` ⇒ `0`.
- [ ] No source code or test files modified; `go test ./...` and `go vet ./...`
      stay clean; `./skpp check` still exits 0.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ Yes: the drift is enumerated with exact line
numbers and linter output, the authoritative flag list is one command away
(`./skpp --help`), and the completion sourcing instructions are quoted verbatim
from the file headers (the single source of truth).

### Documentation & References

```yaml
# MUST READ — load into context before editing
- file: README.md
  why: The ONLY file this task edits. Current state already covers Install / Usage /
        Where skills live / Adding a skill / How skpp finds the store / Constraints.
  pattern: Sectioned Markdown, example-driven, code fences for every command.
  gotcha: It already passes items (b)(c)(e) of the contract — do NOT rewrite those
           sections; this is a consistency sync, not a rewrite. Edits are surgical.

- file: plan/001_fcde63e5bb60/P1M6T16S2/research/verified_facts.md
  why: The full drift inventory with exact line numbers and verified command output.
  pattern: Re-read it; it is the authoritative list of what to change.

- file: completions/skpp.bash
  why: Header comment (lines 3-6) holds the authoritative bash sourcing instructions.
  pattern: Copy the three install options verbatim into the README completions section.
- file: completions/_skpp
  why: Header comment (lines 4-8) holds the authoritative zsh instructions + the
        `autoload -U compinit && compinit` requirement.
- file: completions/skpp.fish
  why: Header comment (line 4) holds the authoritative fish instruction.

- cmd: ./skpp --help
  why: The binary's authoritative flag list. Diff it against README Usage.
  critical: Output is stable (frozen to main.go parseArgs); re-run it, do not
            hardcode the list in the PRP. Currently lists: <tag>, --all/-a,
            --list/-l, --search/-s, check, --path/-p, --file/-f, --relative,
            --no-color, --help/-h, --version/-v.

- url: file:///home/dustin/.pi/agent/skills/write-tech-docs/SKILL.md
  why: Defines the linter's hard rules (no em dashes, no marketing tell-words,
        no >100-word paragraphs) and the lint command.
  critical: Rule #1 is "No em dashes. Not once." — currently 8 violations.

- cmd: bash /home/dustin/.pi/agent/skills/write-tech-docs/scripts/lint.sh README.md
  why: The deterministic docs gate. Currently exits 1 with 8 em-dash hits.
  critical: The linter STRIPS inline code before counting, so the line numbers
            IT prints (3, 99, 101, 102, 105, 117, 119, 122) differ from the
            REAL file lines you see with `grep -n "\u2014" README.md`
            (3, 167, 169, 170, 173, 192, 194, 197). Use grep's numbers to edit;
            use the linter's exit code to confirm you are done. Do not be
            confused if the two disagree — they index the same 8 dashes.

- file: PRD.md
  why: §14 (completions spec), §15 (README outline), §6.1/§6.2 (CLI contract).
  section: "## 14. Shell completions" and "## 15. README.md outline"
  gotcha: PRD §15 outline does NOT include a completions section — it was
           deferred to this task by P1.M6.T15.S1. Add the section anyway; it is
           in scope here.
```

### Current Codebase tree (repo root)

```bash
skpp/
├── PRD.md                 # read-only
├── README.md              # ← THIS TASK EDITS THIS FILE
├── LICENSE
├── go.mod / go.sum
├── .gitignore
├── main.go                # read-only (CLI surface; parseArgs has the flag list)
├── main_test.go           # read-only
├── install.sh             # read-only (already matches README §3)
├── internal/              # read-only (discover/resolve/skillsdir/ui/check)
├── completions/
│   ├── skpp.bash          # header = authoritative bash sourcing instructions
│   ├── _skpp              # header = authoritative zsh instructions
│   └── skpp.fish          # header = authoritative fish instruction
├── skills/example/SKILL.md
└── skpp                   # built binary
```

### Desired Codebase tree

```bash
# Only ONE file changes. No new files are created.
skpp/
└── README.md   # + ## Shell completions section; + --relative/--no-color in Usage; 0 em dashes
```

### Known Gotchas of our codebase & Library Quirks

```python
# CRITICAL: This task is DOCS-ONLY. It must not edit main.go, tests, install.sh,
#           completions/*, PRD.md, tasks.json, or prd_snapshot.md. If a "fix"
#           seems to require a code change, stop — that belongs to a different
#           subtask. The README must conform to the binary, never the reverse.

# CRITICAL: Do NOT invent completion install paths. The completion file headers
#           (completions/skpp.bash, _skpp, skpp.fish) are the single source of
#           truth. Mirror them verbatim.

# CRITICAL: The em-dash replacements must keep the markdown valid. The 8 hits
#           are all the `X — Y` lead-in pattern (e.g. `` `name` — required.``).
#           Replace the em dash with a colon: `` `name`: required.`` Do NOT
#           introduce a hyphen-minus or en dash; the linter scans for `—` (U+2014)
#           AND other dash glyphs.

# GOTCHA: `--no-color` and `--relative` are modifiers, not modes. In Usage, list
#         them as modifiers (grouped or one-liner each), consistent with how the
#         binary's --help labels them "(modifier)".

# GOTCHA: README §3 Install, "How skpp finds the store", and "Constraints" are
#         ALREADY consistent with install.sh / PRD §8 / PRD §17. Verify them, do
#         not edit them. Editing them is scope creep and risks introducing drift.

# GOTCHA: The write-tech-docs linter also flags tell-words (powerful, robust,
#         seamless, leverage, utilize, unlock, empower, supercharge, streamline,
#         elevate, delve, ...) and >100-word paragraphs. Re-run after edits — a
#         new sentence could trip one. Currently only em dashes fail.
```

## Implementation Blueprint

### Data models and structure

Not applicable — this task edits prose in a single Markdown file. There is no
data model, schema, or type to define.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: VERIFY the already-consistent sections (no edit unless broken)
  - RE-READ README.md §3 "Install", "How skpp finds the store", "Constraints".
  - RUN: ./skpp --help  (capture the full flag list)
  - RUN: grep -n "ln -sfn\|SKPP_INSTALL_BIN\|\.local/bin\|usr/local/bin" install.sh
  - CHECK: README §3 matches install.sh target order + symlink + go-install caveat.
  - CHECK: README "finds the store" matches §8 (env → sibling-of-binary → walk-up → fail).
  - CHECK: README "Constraints" lists manifest-free, never-auto-discovered (with the
           full forbidden-location list), loaded-only-via---skill, zero runtime deps.
  - DECISION: if all match (they do today), make NO edit. If any drift, fix only that.
  - WHY FIRST: establishes the baseline so later edits do not accidentally regress it.

Task 2: ADD "## Shell completions" section to README.md  (GAP 1 — PRIMARY)
  - CREATE a new top-level section "## Shell completions".
  - PLACE: immediately AFTER the "## Install" section and BEFORE "## Usage"
           (completions are install-adjacent setup; keeps setup content together).
  - CONTENT — three subsections (bash / zsh / fish), sourcing copied VERBATIM from
    the completion file headers (the single source of truth):
      bash:  source /path/to/skpp/completions/skpp.bash
             cp completions/skpp.bash ~/.local/share/bash-completion/completions/skpp
             cp completions/skpp.bash /etc/bash_completion.d/skpp
      zsh:   cp completions/_skpp ~/.zsh/completions/_skpp
             cp completions/_skpp /usr/local/share/zsh/site-functions/_skpp
             then: autoload -U compinit && compinit   (in .zshrc)
      fish:  cp completions/skpp.fish ~/.config/fish/completions/skpp.fish
  - ADD one sentence: tag completion is DYNAMIC (manifest-free) — the shell calls
    `skpp --relative --all` at completion time, so it never goes stale. This is the
    core difference from a hardcoded list and matters to users.
  - NAMING: section header exactly `## Shell completions`.
  - FOLLOW tone: the rest of the README (short, example-driven, code-fenced).
  - GOTCHA: do NOT reference a completion installer in install.sh — install.sh does
            NOT currently install completions (PRD §14 allows deferral). Document
            manual sourcing only.

Task 3: ADD --relative and --no-color to README Usage  (GAP 2)
  - FIND: the "## Usage" section's commented example block (ends with the line
          "skpp --version" and the "skpp --help lists every flag." sentence).
  - ADD two concise one-liners consistent with the existing comment style, e.g.:
        # Print paths relative to the skills directory (instead of absolute)
        skpp --relative example
        # Disable ANSI color even on a TTY (for --list / --search tables)
        skpp --no-color --list
  - NAMING: keep the `--long, -short` form where a short form exists; --relative and
            --no-color have no short form (mirror --help's "(modifier)" labeling).
  - PRESERVE: every existing usage line and the closing "skpp --help lists every
              flag." sentence (it stays true and useful).

Task 4: RUN the write-tech-docs linter and fix every hit  (GAP 3 — item f)
  - RUN: bash /home/dustin/.pi/agent/skills/write-tech-docs/scripts/lint.sh README.md
  - CURRENT OUTPUT (verified): exit 1, 8 em-dash hits, all the `X — Y` lead-in
    pattern. NOTE the linter prints its OWN line numbers (3, 99, 101, 102, 105,
    117, 119, 122) because it strips inline code first; the REAL file lines are
    3, 167, 169, 170, 173, 192, 194, 197 (see `grep -n "\u2014" README.md`).
    Edit at the grep line numbers; trust the linter's exit code for completion.
  - FIX each: replace the em dash (U+2014) with a colon. Examples:
        "Standalone skill loader for pi — resolves..." -> use a period: two sentences.
        "`name` — required."               -> "`name`: required."
        "`description` — required (...)."  -> "`description`: required (...)."
        "`metadata...` — optional."        -> "`metadata...`: optional."
        "is a copy-pasteable template — start from it." -> "...template; start from it."
        "**`SKPP_SKILLS_DIR` env var** — wins..."      -> "**`SKPP_SKILLS_DIR` env var**: wins..."
        "**Sibling of the binary** — ..."              -> "**Sibling of the binary**: ..."
        "**Walk up from the current directory** — ..." -> "**Walk up ...**: ..."
  - RE-RUN the linter until exit 0. It also flags tell-words and >100-word
    paragraphs; if a Task 2/3 addition trips one, fix that sentence too.
  - VERIFY: grep -c "—" README.md  -> 0
  - GOTCHA: do NOT replace an em dash with a hyphen-minus or en dash; the linter
            scans dash glyphs generally. Use a colon, period, semicolon, or comma.

Task 5: FINAL binary↔README diff + regression re-run
  - RUN: ./skpp --help
  - DIFF: confirm every flag/subcommand in --help appears in README, and every flag
          the README names appears in --help. Empty symmetric diff = pass.
  - RUN (regression — must stay green; this task touched no code):
        go test ./...
        go vet ./...
        ./skpp check                  # exit 0, example OK
        ./skpp example >/dev/null && echo "resolve OK"
  - RUN a spot of the §13 acceptance that proves docs did not break behavior:
        test "$(./skpp --path)" = "$PWD/skills" && echo "path OK"
  - EXPECTED: all pass. If any fail, the cause is NOT this task's prose edits —
              investigate, but do not weaken the test.
```

### Implementation Patterns & Key Details

```markdown
# The single README edit pattern: surgical, evidence-driven.

# 1. Completion section — mirror the file headers exactly (verbatim copy),
#    never paraphrase a path:
## Shell completions
skpp ships dynamic completions (bash, zsh, fish). Tags are completed at runtime
from `skpp --relative --all`, so the list never goes stale.
<verbatim per-shell sourcing from completions/* headers>

# 2. Modifier flags — match the binary's "(modifier)" grouping, comment style
#    identical to the surrounding Usage block.

# 3. Em dashes — colon substitution preserves markdown list/bold rendering:
`name` — required.        # BEFORE
`name`: required.         # AFTER  (colon, not hyphen)
```

### Integration Points

```yaml
README.md (the ONLY integration surface):
  - section add:   "## Shell completions" (after "## Install", before "## Usage")
  - section edit:  "## Usage"  (add --relative, --no-color lines)
  - prose edit:    8 em-dash replacements across §1 one-liner, "Adding a skill",
                   "How skpp finds the store"

NO OTHER FILES:
  - code:        none (main.go, internal/*, tests untouched)
  - build:       none (go.mod, install.sh untouched)
  - config:      none
  - completions: untouched (their headers are the source the README copies)
  - PRD/plan:    none (PRD.md, tasks.json, prd_snapshot.md are read-only)
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
# Docs gate — must exit 0 before anything else is considered done.
bash /home/dustin/.pi/agent/skills/write-tech-docs/scripts/lint.sh README.md
# Expected: "lint: 0 hit(s)" and exit 0.

# Em-dash sanity check (the linter enforces this, but a fast grep confirms).
grep -c "—" README.md        # Expected: 0
# Also scan for accidental look-alikes the linter may flag:
grep -nP "[—–]" README.md    # Expected: no output (no em/en dashes)

# Markdown still well-formed (code fences balanced). Quick fence parity check:
[ "$(grep -c '```' README.md)" -eq $(( $(grep -c '```' README.md) / 2 * 2 )) ] && echo "fences OK"
```

### Level 2: Binary↔README Consistency (the core contract)

```bash
# Capture the binary's authoritative flag list.
./skpp --help > /tmp/skpp_help.txt

# Every long flag in --help must appear in the README.
for f in --all --list --search --path --file --relative --no-color --help --version; do
  grep -q -- "$f" README.md || echo "MISSING in README: $f"
done
# Expected: no MISSING lines. (check is a subcommand, also verify:)
grep -q '`check`\|^check\| skpp check' README.md && echo "check documented OK"

# Symmetric: nothing the README advertises should be absent from the binary.
# (Manual eyeball against /tmp/skpp_help.txt OPTIONS block — small surface.)

# Completion section present and covers all three shells.
grep -q "## Shell completions" README.md   # section header exists
grep -qi 'bash' README.md && grep -qi 'zsh' README.md && grep -qi 'fish' README.md
grep -q 'skpp --relative --all\|skpp --all' README.md   # dynamic-completion note
# Expected: all true.
```

### Level 3: Regression (docs edits must not break behavior)

```bash
# This task touches no code; these must remain green from P1.M6.T16.S1.
go test ./...        # Expected: ok / PASS, all packages
go vet ./...         # Expected: clean (no output)
./skpp check         # Expected: exit 0, "OK    example (example)" + 0 errors
test -d "$(./skpp example)" && echo "resolve OK"
test "$(./skpp --path)" = "$PWD/skills" && echo "path OK"
```

### Level 4: Domain-Specific Validation

```bash
# Render check — eyeball the new completions section + Usage additions in a
# terminal markdown viewer (or `mdsel README.md` for a TOC) to confirm section
# ordering and that no code fence is broken.
mdsel README.md 2>/dev/null | sed -n '1,40p'   # if mdsel available

# Prove the documented completion sourcing path is real (the file exists where
# the README says to copy it from):
test -f completions/skpp.bash && test -f completions/_skpp && test -f completions/skpp.fish \
  && echo "completion source files present"
```

## Final Validation Checklist

### Technical Validation

- [ ] Level 1: `bash .../write-tech-docs/scripts/lint.sh README.md` exits 0.
- [ ] Level 1: `grep -c "—" README.md` ⇒ `0`.
- [ ] Level 2: every `--help` flag appears in README (no MISSING lines).
- [ ] Level 2: `## Shell completions` section present; bash/zsh/fish all covered.
- [ ] Level 2: README Usage names `--relative` and `--no-color`.
- [ ] Level 3: `go test ./...` passes (untouched code still green).
- [ ] Level 3: `go vet ./...` clean.
- [ ] Level 3: `./skpp check` exits 0; `./skpp --path` == `$PWD/skills`.

### Feature Validation

- [ ] All success criteria from "What" section met.
- [ ] Completion sourcing instructions in README match `completions/*` headers
      verbatim (no invented paths).
- [ ] Dynamic-completion behavior (`skpp --relative --all`) stated in the README.
- [ ] README §3 / "finds the store" / "Constraints" verified consistent (unedited
      unless drift found).
- [ ] Em-dash replacements kept markdown valid (lists, bold, code spans render).

### Code Quality Validation

- [ ] Only `README.md` modified (`git status` shows exactly one changed file).
- [ ] No source code, tests, install.sh, completions, PRD.md, tasks.json, or
      prd_snapshot.md touched.
- [ ] No new doc files created (no overview gap was found).
- [ ] Prose stays in the existing README tone (short, example-driven).

### Documentation & Deployment

- [ ] README is internally consistent (flag names match binary spelling exactly:
      `--relative`, `--no-color`, `--file`/`-f`, etc.).
- [ ] No environment variables newly documented or removed (`SKPP_SKILLS_DIR` and
      `SKPP_INSTALL_BIN` coverage unchanged).

---

## Anti-Patterns to Avoid

- ❌ Don't edit any file other than `README.md`. This is a docs sync, not a code
  task. If `--help` and the README disagree, the README is wrong, not the binary.
- ❌ Don't invent completion install paths or a completion step in `install.sh`.
  install.sh does not install completions; copy sourcing from the file headers.
- ❌ Don't rewrite the already-consistent sections (Install, finds-the-store,
  Constraints) to "improve" them. Verify, don't refactor. Scope creep reintroduces
  drift.
- ❌ Don't replace an em dash with a hyphen or en dash — the linter scans dash
  glyphs. Use a colon, period, semicolon, or comma.
- ❌ Don't hardcode the `--help` flag list in the PRP or README as the source of
  truth; re-run `./skpp --help` and diff.
- ❌ Don't skip the regression run (`go test ./...`, `./skpp check`) just because
  "I only edited Markdown" — confirm no accidental code change slipped in.
- ❌ Don't weaken or alter the §13 acceptance suite or any test.

---

## Confidence Score

**9/10.** The drift is fully enumerated with verified line numbers and linter
output, the edits are surgical (one file, three small changes), the completion
sourcing instructions are quoted verbatim from the file headers, and the
validation gates are deterministic commands already confirmed to run in this
repo. The only residual risk is an editor introducing a new tell-word or long
paragraph while rewriting an em-dash line, which the Level 1 linter re-run
catches immediately.
