# Verified Facts — P1.M6.T16.S2 (Mode B documentation sync)

Source of truth for this PRP. Every claim below was re-verified against the
actual repo state on 2026-07-07, post `P1.M6.T16.S1` acceptance (binary built,
§13 suite green, example skill shipped, completions shipped).

## 1. What this task is (Mode B)

`Mode B` = the mandatory final cross-cutting documentation-sync task defined in
the plan's SOW §5. It depends on **all 20 implementing subtasks** and is
guaranteed to run last. Its job: prevent the changeset from shipping a README
that has drifted from the final implemented binary. It edits `README.md`
**in place**. It does NOT create new doc files unless an overview gap is found.

## 2. Drift inventory (README.md vs final binary) — VERIFIED

### GAP 1 — PRIMARY: shell completions are undocumented in README

- Binary ships three completion files (PRD §14):
  - `completions/skpp.bash`
  - `completions/_skpp` (zsh)
  - `completions/skpp.fish`
- `grep -ni "complet\|_skpp" README.md` ⇒ **zero matches**. The README never
  mentions completions at all.
- This gap was **deliberately deferred** to this task by the completions
  subtask: `P1.M6.T15.S1` context_scope says *"Add a short note in README (via
  P1.M6.T16.S2) on how to source/install them."* This task owns that note.
- Sourcing instructions exist verbatim in each file's header comment (the
  authoritative source — do not invent new paths):

  ```text
  # bash (one of):
    source /path/to/skpp/completions/skpp.bash
    cp completions/skpp.bash ~/.local/share/bash-completion/completions/skpp
    cp completions/skpp.bash /etc/bash_completion.d/skpp

  # zsh (one of):
    cp completions/_skpp ~/.zsh/completions/_skpp
    cp completions/_skpp /usr/local/share/zsh/site-functions/_skpp
    then ensure: autoload -U compinit && compinit   (in .zshrc)

  # fish:
    cp completions/skpp.fish ~/.config/fish/completions/skpp.fish
  ```

- Completions are dynamic (manifest-free): they call `skpp --relative --all`
  for tag completion. README should state this so users do not expect a static
  list.

### GAP 2 — MINOR: two modifier flags absent from README Usage block

- `./skpp --help` lists these flags (the binary's authoritative surface):
  `<tag>`, `--all/-a`, `--list/-l`, `--search/-s`, `check`, `--path/-p`,
  `--file/-f`, `--relative`, `--no-color`, `--help/-h`, `--version/-v`.
- README §4 "Usage" shows every flag **except** `--relative` and `--no-color`.
  It closes with *"skpp --help lists every flag."*
- Subtask contract item (a): *"every flag/subcommand in the binary must appear
  in the README, and vice versa."* The deferral line covers them loosely, but
  the strict contract is satisfied only by naming both modifiers. Low effort:
  add a one-liner each (or a grouped "Modifiers" line) to the Usage block.

### GAP 3 — OPTIONAL (item f): write-tech-docs linter fails with 8 em-dash hits

- The `write-tech-docs` skill ships `scripts/lint.sh`. Hard rule #1:
  *"No em dashes. Not once."*
- Ran on the current README: `bash scripts/lint.sh README.md` ⇒ **exit 1,
  8 hits**, all the `X — Y` lead-in pattern.
- IMPORTANT line-number discrepancy: the linter strips inline code before
  counting, so the line numbers IT prints (3, 99, 101, 102, 105, 117, 119, 122)
  differ from the REAL file lines (3, 167, 169, 170, 173, 192, 194, 197) you see
  with `grep -n "\u2014" README.md`. Both index the same 8 dashes; use grep's
  numbers to edit, the linter's exit code to confirm done.
- Each is a single em dash before a description, trivially replaceable with a
  colon (e.g. `` `name` — required.`` → `` `name`: required.``) or period.
- TENSION (record it, then resolve in favor of the linter): the PRD and the
  mcpeepants reference style use em dashes. But the repo is expected to be
  write-tech-docs-linted, and the subtask explicitly invites the linter.
  Decision for the PRP: **apply the linter** and replace all 8. It is cheap
  and makes the README pass the gate deterministically.

## 3. Already-consistent sections (verify only — do NOT edit unless broken)

These matched on inspection and should NOT be changed by this task (changing
them is scope creep):

- **README §3 "Install"** ↔ `install.sh`:
  - install.sh target order = `$SKPP_INSTALL_BIN` → `$HOME/.local/bin` →
    `/usr/local/bin` (only if writable) → fail with hint. README §3 matches.
  - install.sh **symlinks** with `ln -sfn` (not copy). README §3 explains the
    symlink rationale (resolves back to repo).
  - README §3 **B. `go install`** documents the `SKPP_SKILLS_DIR` caveat
    prominently (a `go install`'d binary has no adjacent `skills/`). Matches
    PRD §12.2. ✓
  - README §3 **C. From source** shows manual `go build` + `ln -sfn`. ✓
- **README "How `skpp` finds the store"** ↔ PRD §8 priority order:
  env → sibling-of-binary (`os.Executable()`+`EvalSymlinks()`) → walk-up-from-cwd
  → one-line fix. README matches all four, including the symlink-vs-copy warning.
- **README "Constraints"** ↔ PRD §17: manifest-free; never auto-discovered by pi
  (with the full forbidden-location list: `~/.pi/agent/skills`,
  `~/.agents/skills`, `.pi/skills`, `.agents/skills`, `node_modules`,
  `pi.skills` in package.json); loaded only via `--skill`; only prints paths;
  zero runtime deps. ✓

## 4. Command inventory (verified working in this repo)

```bash
./skpp --help                                   # the authoritative flag list (stdout, exit 0)
bash /home/dustin/.pi/agent/skills/write-tech-docs/scripts/lint.sh README.md
                                                # linter; currently exit 1, 8 em-dash hits
grep -c "—" README.md                           # quick em-dash count (should reach 0)
go test ./...                                   # all unit tests still green after doc edits
go vet ./...                                    # clean (docs edits must not touch code)
./skpp check                                    # exit 0; example skill OK
```

## 5. Scope guardrails (do NOT do these in this task)

- Do NOT create new doc files (no `docs/`, no `CHANGELOG.md`) unless a real
  overview gap is found — none was found.
- Do NOT edit `PRD.md`, `tasks.json`, `prd_snapshot.md`, or any source code.
  This task touches `README.md` only.
- Do NOT weaken or alter the §13 acceptance suite or unit tests.
- Do NOT "improve" prose beyond the linter hits + the two required content
  additions (completions + modifiers). Mode B is a consistency sync, not a
  rewrite.
