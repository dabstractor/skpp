# PRP — P1.M6.T12.S1: The one shipped example skill (`skills/example/SKILL.md`, PRD §11)

> **Subtask:** P1.M6.T12.S1 — ship the single example skill so `--list`, tag
> resolution, `check`, and the `pi --skill "$(skpp example)"` end-to-end flow are
> demonstrable out of the box. Previously tracked as P1.M6.T1.S1 (same deliverable).
>
> **Scope:** CREATE exactly **one** file — `skills/example/SKILL.md` — with the
> **byte-exact** PRD §11 content. No other file is created or modified. This task
> touches **zero `.go` files** and adds **zero** dependencies. It is the last
> missing piece (PRD §11) before `install.sh`/README/completions (P1.M6.T13–T15)
> and the §13 acceptance sweep (P1.M6.T16).
>
> **Status of upstream gates (all LANDED + GREEN — verified today):** P1.M1
> (`skillsdir.Find` → `--path`), P1.M2 (`discover.ParseFrontmatter`/`Skill`/`Index`
> + `ui --list`), P1.M3 (`resolve.Resolve` + `skpp <tag>`/`-f`/`--all`/`--relative`),
> P1.M4 (`--search` + `internal/check`/`skpp check`), P1.M5 (full CLI matrix +
> exit-code/error semantics). `go build -o skpp .` → exit 0; `go test ./...` →
> 212 `Test*` funcs green. The example skill is the only thing the tooling has had
> nothing to resolve against.
>
> **Predecessor research (READ FIRST):**
> `plan/001_fcde63e5bb60/P1M6T1S1/research/verified_facts.md` — the exhaustive
> byte-level + parse-level verification of this exact deliverable (written before
> `check` landed; updated in this task's own
> `research/verified_facts.md` §3 which confirms `skpp check` is now wired).

---

## Goal

**Feature Goal:** A single, spec-compliant example skill exists on disk at
`skills/example/SKILL.md` so that every already-landed `skpp` command has a real
target to resolve, list, search, and validate against — proving the full
manifest-free pipeline end to end.

**Deliverable:** The file `skills/example/SKILL.md` containing EXACTLY the PRD §11
content (569 bytes): valid Agent-Skills frontmatter (`name: example`, a folded
`description: >`, `metadata.keywords`, `metadata.category`) followed by a short
markdown body whose inner ``` ```bash ``` fence demonstrates `skpp example`,
`skpp -f example`, and `pi --skill "$(skpp example)"`. The `skills/` and
`skills/example/` directories are created implicitly by writing the file.

**Success Definition:** (1) the file exists with byte-exact PRD §11 content;
(2) `./skpp example`, `./skpp -f example`, `./skpp --list`, `./skpp --path`,
`./skpp check`, and `./skpp --search skpp` all succeed against it; (3) `go build`
and `go test ./...` remain green; (4) `pi --no-skills --skill "$(./skpp example)"`
loads the skill (PRD §13 acceptance); (5) no other skill ships.

## Why

- **PRD §2 constraint 4 / §11:** "No development of skills beyond one example.
  Ship exactly one example skill to prove the pipeline." Without it the whole CLI
  resolves/list/validates against an empty store, and the §13 acceptance suite and
  the downstream `install.sh`/README (P1.M6.T13–T14) have nothing concrete to
  point at.
- **PRD §13 acceptance:** the end-to-end `pi --skill "$(skpp example)"` line is
  the proof that skills load ONLY via the explicit `--skill` path (PRD §2
  constraint 2), never via pi auto-discovery. That line is un-runnable until this
  file exists.
- **Integration with existing features:** `discover.Index` (P1.M2.T5),
  `resolve.Resolve` (P1.M3.T7), `ui.PrintList` (P1.M2.T6), `search`
  (P1.M4.T9), and `check.Validate` (P1.M4.T10) are all exercised for the first
  time against a real, on-disk skill — a regression catch on the entire discovery
  stack.
- **Cohesion with future work items:** `install.sh` (T13) ends with
  `skpp example` as its verification step; README (T14) documents the canonical
  `pi --skill "$(skpp example)"` example; completions (T15) call `skpp --all`.
  All three need this skill present.

## What

A user-visible behavior change with **zero** code change: after this task,
`./skpp example` prints an absolute path instead of erroring on an empty store,
`./skpp --list` shows one row, and `./skpp check` reports `OK`. The file is the
canonical "drop a `<tag>/SKILL.md` under `skills/`" pattern made concrete, so the
README's "Adding a skill" section (PRD §15.6) has a worked example on disk.

### Success Criteria

- [ ] `skills/example/SKILL.md` exists and is byte-identical to the PRD §11 block
      (569 bytes; content given verbatim in "All Needed Context" below).
- [ ] `git ls-files skills/` lists EXACTLY `skills/example/SKILL.md` — one file,
      one skill (PRD §2 constraint 4, §17).
- [ ] `test "$(./skpp example)" = "$PWD/skills/example"` passes.
- [ ] `test "$(./skpp -f example)" = "$PWD/skills/example/SKILL.md"` passes.
- [ ] `test "$(./skpp --path)" = "$PWD/skills"` passes.
- [ ] `./skpp --list` prints a row whose TAG is `example` and exits 0.
- [ ] `./skpp --search skpp` prints the `example` row (matches `metadata.keywords`
      entry `skpp`) and exits 0.
- [ ] `./skpp check` prints `OK   example (example)` and exits 0.
- [ ] `go build ./...` and `go test ./...` are green (no .go touched).

## All Needed Context

### Context Completeness Check

_Pass:_ A stranger who knows nothing about this repo can implement this by (a)
reading the exact 569-byte file content below, (b) the single `write` command to
`skills/example/SKILL.md`, and (c) the validation commands. The frontmatter is
already proven to parse cleanly through the landed `discover` package; no code
reasoning is required.

### Documentation & References

```yaml
# MUST READ — primary spec for this exact file
- docfile: PRD.md
  why: §11 "The one shipped example skill" gives the AUTHORITATIVE, byte-exact file
        content (inside a four-backtick fence whose wrapper is NOT part of the file).
  section: "## 11. The one shipped example skill"
  critical: >
    The outer `````markdown` fence is rendering-only — do NOT write it into the
    file. The inner ``` ```bash ``` ... ``` ``` block IS real file content (pi
    renders it on load). The trailing PRD sentence "No other skills ship in this
    repo." is PROSE after the closing fence, NOT a line in the file.

# MUST READ — exhaustive byte/parse verification of this deliverable
- docfile: plan/001_fcde63e5bb60/P1M6T1S1/research/verified_facts.md
  why: §1 gives the byte-verified 569-byte content; §2 proves it parses to the
        exact typed Skill fields; §4 shows the discover package already accepts it.
  critical: >
    Confirms description folded-scalar `>` parses to 162 chars (≤ 1024, no WARN),
    name `example` matches `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, and aliases are
        absent (nil) — all OK under PRD §9.

# Updated current-state note (supersedes only the "check pending" part of the above)
- docfile: plan/001_fcde63e5bb60/P1M6T12S1/research/verified_facts.md
  why: §3 confirms `skpp check` is NOW wired (main.go:184) and build/test are
        green today; §6 is the verified-working command table.

# Frontmatter parser that consumes this file (no changes needed — reference only)
- file: internal/discover/discover.go
  why: ParseFrontmatter + Frontmatter define what the file must satisfy. yaml.v3
        handles `description: >` folded scalars and `metadata:` map with list values.
  pattern: BOM-stripped, CRLF-tolerant "---" fence detection; lenient on unknown keys.
  gotcha: Opening "---" present but no closing "---" → treated as NO frontmatter.
          The example has both fences, so this is not a concern.

# Skill struct built from the frontmatter (reference only — no change)
- file: internal/discover/skill.go
  why: BuildSkill extracts keywords/category/aliases via toStringSlice ([]any→[]string).
        Confirms `metadata.keywords: [example, demo, skpp]` → []string{...}.

# Validator that `skpp check` runs over the file (reference only — no change)
- file: internal/check/check.go
  why: Defines the OK/WARN/ERROR report. The example must hit the OK path.
  pattern: name regex `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` len≤64; description≤1024;
        non-empty name+description; duplicate-name detection across the store.

# Agent Skills spec (frontmatter contract — required: name, description; optional: metadata)
- url: https://agentskills.io/specification
  why: `metadata` is a spec'd OPTIONAL field; keywords/category/aliases under it are
        standard-compliant, so the example is not a skpp-specific extension.

# pi's --skill behavior (factual grounding for the §13 pi line)
- docfile: /home/dustin/.local/lib/node_modules/@earendil-works/pi-coding-agent/docs/skills.md
  why: --skill <dir> accepts a skill directory (additive, works with --no-skills).
  critical: >
    A skill with NO description is NOT loaded by pi. The example's folded
        description is non-empty after folding, so it loads. name need NOT match
        the directory (pi relaxes the Agent Skills standard for shared dirs).
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── PRD.md                 # spec; §11 is this task's source of truth
├── go.mod                 # module github.com/dabstractor/skpp; go 1.25; yaml.v3
├── main.go                # CLI entrypoint; case "check": at line 184 (WIRED)
├── main_test.go           # 97 tests
├── internal/
│   ├── discover/          # ParseFrontmatter, Skill, BuildSkill, Index  (LANDED)
│   ├── skillsdir/         # Find() §8 priority                         (LANDED)
│   ├── resolve/           # Resolve() §7.2 precedence                  (LANDED)
│   ├── ui/                # PrintList table + ANSI                     (LANDED)
│   ├── search/            # substring search                           (LANDED)
│   └── check/             # Validate() §9                              (LANDED)
└── plan/001_fcde63e5bb60/
    ├── P1M6T1S1/research/{verified_facts.md, validate_example_probe.go}  # OLD id, SAME task
    └── P1M6T12S1/research/verified_facts.md                              # THIS task (updated note)
# NOTE: no skills/ directory exists yet — created by this task.
```

### Desired Codebase tree with files to be added

```bash
skpp/
└── skills/
    └── example/                       # the ONE shipped skill dir (PRD §11)
        └── SKILL.md                   # ← THE ONLY file this task creates (569 bytes)
# No scripts/, references/, assets/ siblings — the PRD §11 example has none.
# No manifest/index file (PRD §2 constraint 1, §17).
# No second skill (PRD §2 constraint 4, §17).
```

### Known Gotchas of our codebase & Library Quirks

```python
# CRITICAL (PRD §11 rendering): the PRD wraps the example in a FOUR-backtick
# `````markdown` fence so the inner THREE-backtick ``` ```bash ``` fence survives
# Markdown. ONLY the inner content (from `---` to the closing ``` of the bash
# fence) is the file. Do NOT write four backticks, do NOT write the wrapper.

# CRITICAL (PRD prose vs file): "No other skills ship in this repo." appears in
# the PRD AFTER the closing fence. It is the author restating §2 constraint 4.
# It is NOT a line in SKILL.md.

# GOTCHA (yaml folded scalar `>`): `description: >` joins the following indented
# lines with SPACES and appends ONE trailing newline → 162 chars, ≤ 1024, so no
# WARN. Do not "normalize" the indentation or reflow the lines — byte-exactness
# is the contract. The discover package copies the description VERBATIM
# (skill.go); check's 1024-char test trims if it wants visible length.

# GOTCHA (flow-sequence lists): `keywords: [example, demo, skpp]` is a YAML
# flow sequence. yaml.v3 delivers it as []interface{} ([]any); discover's
# toStringSlice ([]any→[]string, skill.go) normalizes it. A bare scalar or a
# block sequence would also work, but use the PRD's exact flow-sequence form.

# GOTCHA (no skills/ dir yet): `git ls-files` shows no skills/ and no .gitkeep.
# Writing skills/example/SKILL.md (write tool auto-creates parents) is the whole
# creation step. There is nothing to delete.

# GUARDRAIL (PRD §2.2): the store resolves (§8 rule 2) to <repo>/skills, which is
# NOT ~/.pi/agent/skills. Assert: the path printed by `./skpp --path` is NOT
# under $HOME/.pi. (This is structural, not something the file content controls.)
```

## Implementation Blueprint

### Data models and structure

None. This task creates a Markdown content file; the data model (`discover.Skill`,
`discover.Frontmatter`) already exists and is unchanged. The file is *consumed* by
the landed `discover`/`resolve`/`ui`/`search`/`check` packages — not a producer.

### The exact file to write — `skills/example/SKILL.md`

This is the **only** artifact. Copy verbatim (the leading `---` is line 1; the
final line is the closing ``` of the bash fence; 569 bytes total):

````markdown
---
name: example
description: >
  Reference example skill for skpp. Demonstrates the required frontmatter and
  how skpp resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skpp]
  category: meta
---

# Example Skill

This skill exists only so `skpp` has something to resolve.

Try:

```bash
skpp example                       # prints this directory's absolute path
skpp -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skpp example)"       # loads this skill into pi
```
````

(Read this block from the top `---` to the final ``` line ONLY. The four-backtick
wrapper above is this PRP's own rendering fence, mirroring PRD §11.)

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: CREATE skills/example/SKILL.md
  - IMPLEMENT: write the byte-exact PRD §11 content shown above (569 bytes).
  - SOURCE: PRD.md §11 (authoritative) AND
            plan/001_fcde63e5bb60/P1M6T1S1/research/verified_facts.md §1
            (byte-verified copy).
  - NAMING: directory MUST be `skills/example/` so the sibling-of-binary store
            (§8.2) makes its relTag exactly `example`. `name: example` matches
            the dir (good practice, not required by pi).
  - PLACEMENT: repo root → skills/example/SKILL.md. No other path.
  - VERIFY byte-exactness after writing:
      wc -c skills/example/SKILL.md        # expect 569
      git ls-files skills/                 # expect exactly skills/example/SKILL.md

Task 2: (NO-OP, defensive) REMOVE a throwaway skills/.gitkeep if present
  - WHY: the contract history mentioned a placeholder .gitkeep from P1.M1.T3.S1.
         `git ls-files` today shows NONE, so this is a harmless `rm -f`.
  - COMMAND: rm -f skills/.gitkeep   # exits 0 whether or not it existed
  - DO NOT create skills/.gitkeep yourself.

Task 3: VALIDATE (no implementation — see Validation Loop)
  - RUN the build/test gate, then all six `./skpp` acceptance commands, then the
    §13 pi line. Fix regressions ONLY if a landed package misbehaves on this file
    (not expected: the file was pre-verified against discover.ParseFrontmatter).
```

### Implementation Patterns & Key Details

```python
# There is no code to write. The only "pattern" is byte-exactness.

# WRITE (one call, parent dirs auto-created):
#   write(path="skills/example/SKILL.md", content=<the 569-byte block above>)

# DO NOT:
#   - reflow the folded description, change `>` to `|`, or merge its two lines
#   - convert `keywords: [example, demo, skpp]` to a block list
#   - add an `aliases:` field (PRD §11 has none)
#   - add license/compatibility frontmatter (PRD §11 has none)
#   - add scripts/ references/ assets/ dirs (PRD §11 ships none)
#   - add a second example skill
#   - touch any .go, go.mod, go.sum, .gitignore, PRD.md, or tasks.json
```

### Integration Points

```yaml
DISCOVER (no change): discover.Index("<repo>/skills") will now return exactly one
  Skill{RelTag:"example", Name:"example", Dir:"<repo>/skills/example",
  Keywords:["example","demo","skpp"], Category:"meta", HasFM:true}. Verified.

RESOLVE (no change): resolve.Resolve("example", idx) hits §7.2 step 1 (exact
  relTag) immediately. `skpp -f example` appends "/SKILL.md" to the dir.

UI (no change): --list renders one row. --search "skpp" matches the keyword.

CHECK (no change): internal/check reports `OK   example (example)`, exit 0.
  No duplicate-name (only one skill), name valid, description non-empty & ≤1024.

CONFIG (no change): no env vars, no new flags, no settings.

DATABASE: none (manifest-free by PRD §2 constraint 1).

GIT: `git add skills/example/SKILL.md` is the only staged change. /skpp (built
  binary) is already gitignored (.gitignore line 1). The skills/ tree IS tracked.
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
# This task writes a .md file, so there is no Go linting of new code. But the
# module must still build and format-clean across all packages (probe-safe):
go build ./... && echo "BUILD OK"          # expect exit 0 (build-ignored probe is fine)
go vet   ./... && echo "VET OK"            # expect exit 0
gofmt -l . | grep -v '^$' || echo "FMT OK" # expect only the build-ignored probe, if anything

# Expected: all clean. The file itself is Markdown; a quick eyeball of the fence
# balance is the only style check (one opening --- and one closing --- in
# frontmatter; one ```bash and one closing ``` in the body).
```

### Level 2: Unit Tests (Component Validation)

```bash
# No new Go tests (no new Go code). Confirm the baseline is still green:
go test ./...                              # expect all PASS (212 Test* funcs)

# The example file is itself validated by the landed packages' behavior via the
# CLI gates in Level 3 — there is no unit test that points specifically at
# skills/example/ (the repo deliberately does not commit a fixture store; tests
# build temp stores). This is by design.
```

### Level 3: Integration Testing (System Validation) — the real gates

```bash
# Build the binary from repo root (PRD §13 build step).
go build -o skpp . && echo "BUILD OK"

# (a) Discovery + path resolution (sibling-of-binary rule, §8.2)
test "$(./skpp --path)" = "$PWD/skills"   && echo "PATH OK"      # store = <repo>/skills
./skpp --list                             | grep -q '^example' \
  && echo "LIST OK"                                            # one row, TAG=example

# (b) Tag resolution (§7.2 step 1 exact relTag)
test "$(./skpp example)"   = "$PWD/skills/example"        && echo "DIR OK"
test "$(./skpp -f example)" = "$PWD/skills/example/SKILL.md" && echo "FILE OK"
test -d "$(./skpp example)" && test -f "$(./skpp -f example)" && echo "EXISTS OK"

# (c) Absolute-path contract (default; PRD §6.1)
case "$(./skpp example)" in /*) echo "ABSOLUTE OK";; *) echo "FAIL"; exit 1;; esac

# (d) Search over metadata.keywords (PRD §6.1 --search)
./skpp --search skpp | grep -q example && echo "SEARCH OK"      # 'skpp' is a keyword

# (e) Validation — skpp check (PRD §9, NOW wired)
./skpp check | grep -E 'OK +example \(example\)' && echo "CHECK OK"
./skpp check > /dev/null; [ $? = 0 ]               && echo "CHECK EXIT OK"

# (f) Error contract parity: unknown tag prints nothing to stdout, exits 1 (PRD §6.4)
out=$(./skpp nope 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "1" ] && echo "UNKNOWN-TAG CONTRACT OK"

# (g) Guardrail: store is NOT a pi auto-discovery location (PRD §2.2)
case "$(./skpp --path)" in "$HOME"/.pi*) echo "FAIL: under ~/.pi"; exit 1;; *) echo "NOT-PI-LOCATION OK";; esac

# (h) Scope: EXACTLY one skill on disk
[ "$(git ls-files skills/ | wc -l)" = "1" ] && echo "ONE-SKILL OK"
git ls-files skills/                         # must print only skills/example/SKILL.md
```

### Level 4: Creative & Domain-Specific Validation — end-to-end with pi (PRD §13)

```bash
# The PRD §13 acceptance line: the skill loads ONLY via the explicit --skill
# path, never via pi auto-discovery. --no-skills proves we rely solely on it.
pi --no-skills --skill "$(./skpp example)" -p "briefly confirm the example skill is loaded" 2>&1 | head
# Expected: pi runs; its context/output references the example skill; NO error
# about a missing/empty skill. (If `pi` is unavailable in this environment,
# treat (a)-(h) above as the binding gate and note the pi line as deferred.)

# OPTIONAL belt-and-braces frontmatter probe (does NOT depend on internal/check):
go run plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go skills/example/SKILL.md
# Expected:
#   OK   example (example)
#     HasFM=true keywords=[example demo skpp] category="meta" aliases=[]
#     description chars=162 body_present=true
# exit 0. (The probe is //go:build ignore, so it never affects go build ./... .)
```

## Final Validation Checklist

### Technical Validation

- [ ] `go build ./...` → exit 0
- [ ] `go test ./...` → all pass
- [ ] `go vet ./...` → clean; `gofmt -l .` clean (modulo the build-ignored probe)
- [ ] `./skpp --path` → `$PWD/skills` (exit 0)
- [ ] `./skpp example` → `$PWD/skills/example` (exit 0)
- [ ] `./skpp -f example` → `$PWD/skills/example/SKILL.md` (exit 0)
- [ ] `./skpp --list` → one row, TAG `example` (exit 0)
- [ ] `./skpp --search skpp` → `example` row (exit 0)
- [ ] `./skpp check` → `OK   example (example)` (exit 0)
- [ ] `./skpp nope` → empty stdout, exit 1 (error contract intact)

### Feature Validation

- [ ] `skills/example/SKILL.md` is byte-identical to the PRD §11 block (569 bytes)
- [ ] `git ls-files skills/` lists EXACTLY `skills/example/SKILL.md`
- [ ] `pi --no-skills --skill "$(./skpp example)"` loads the skill (PRD §13)
- [ ] Store path is NOT under `$HOME/.pi` (PRD §2.2 guardrail)
- [ ] No second skill, no manifest, no extra sibling dirs (PRD §2, §11, §17)

### Code Quality Validation

- [ ] Zero `.go` files modified (this is a content-only task)
- [ ] No new dependencies; `go.mod`/`go.sum` unchanged
- [ ] File placement matches the desired tree (`skills/example/SKILL.md` only)
- [ ] Frontmatter follows Agent Skills spec (required `name`+`description`; optional `metadata`)

### Documentation & Deployment

- [ ] The example's body is self-documenting (its `Try:` block is the usage demo)
- [ ] No new env vars introduced
- [ ] Downstream readiness: `install.sh` (T13) can end with `skpp example`; README (T14) can cite `pi --skill "$(skpp example)"`; completions (T15) can call `skpp --all`

---

## Anti-Patterns to Avoid

- ❌ Don't change a single byte of the PRD §11 content (re-flowing the folded
  description, swapping `>` for `|`, or rewriting the prose breaks byte-exactness
  and the pre-verified parse).
- ❌ Don't write the four-backtick `````markdown` wrapper or the PRD's trailing
  "No other skills ship in this repo." sentence into the file — both are
  rendering/prose, not file content.
- ❌ Don't add a `scripts/`, `references/`, `assets/` dir, an `aliases:` field, or
  `license:`/`compatibility:` frontmatter — PRD §11 ships exactly the content shown.
- ❌ Don't ship a second example skill (PRD §2 constraint 4, §17 — repo is a loader).
- ❌ Don't add a `skills.json`/index/manifest (PRD §2 constraint 1, §17).
- ❌ Don't touch any `.go`, `go.mod`, `go.sum`, `.gitignore`, `PRD.md`, or
  `tasks.json` — you are writing one Markdown file.
- ❌ Don't create `skills/.gitkeep` — it doesn't exist today and isn't wanted.

---

## Confidence Score

**9/10** for one-pass success. The deliverable is a single, byte-specified
Markdown file whose frontmatter and parse behavior are already empirically
verified against the landed `discover`/`check` packages (see the predecessor
research). The entire risk surface is "did I copy the bytes exactly" — fully
covered by the Level 3 byte/scope assertions and the optional probe. The only
non-deterministic element is the §13 `pi` end-to-end line, which depends on `pi`
being on PATH in the execution environment; Levels 1–3 are self-contained and
binding regardless.
