# PRP — P1.M6.T1.S1: `skills/example/SKILL.md` (exact PRD §11 content)

> **Subtask:** P1.M6.T1.S1 (plan graph: P1.M6.T12.S1) — build-order step 6.
> Ship the **one** example skill so `skpp --list`, `skpp example`, `skpp -f
> example`, and `skpp check` are demonstrable out of the box (PRD §11). This is
> the ONLY skill that ships (PRD §2 constraint 4, §17). It proves the whole
> pipeline end-to-end before packaging/docs (the rest of M6).
>
> **Scope:** CREATE exactly **one** file — `skills/example/SKILL.md` — containing
> EXACTLY the PRD §11 content. No Go code. No new packages. No test changes.
> Optional defensive `rm -f skills/.gitkeep` (it does not exist on disk today).
>
> **Why this is a small, low-risk task:** the file content is fixed verbatim in
> PRD §11, and the `discover`/`resolve`/`ui` machinery that consumes it is already
> landed and green (M2/M3). The only implementation subtlety is the Markdown
> **nested-code-fence** rendering in the PRD (see Gotcha 1). The only verification
> subtlety is that `skpp check` depends on two not-yet-landed upstream items
> (see Dependencies + the check-independent fallback).

---

## Goal

**Feature Goal**: Ship `skills/example/SKILL.md` with byte-exact PRD §11 content
so that, from a clean clone, `go build -o skpp .` then `./skpp example`,
`./skpp -f example`, `./skpp --list`, and (once M4.T10/M5.T1 land) `./skpp check`
all succeed against it — proving the manifest-free discovery → frontmatter-parse
→ tag-resolve → validate pipeline has real input.

**Deliverable**: One new file, `skills/example/SKILL.md` (569 bytes, content in
§"All Needed Context" below). Nothing else changes.

**Success Definition**: `test -f skills/example/SKILL.md`; `./skpp example`
prints `<repo>/skills/example` (absolute) exit 0; `./skpp -f example` prints
`<repo>/skills/example/SKILL.md` exit 0; `./skpp --list` shows the `example` row
exit 0; the check-independent probe prints `OK   example (example)` exit 0; and
(optionally, when M4.T10+M5.T1 are landed) `./skpp check` exits 0 with an `OK`
line. No second skill is created. The path is NOT under a pi auto-discovery
location.

## User Persona

**Target User**: A pi operator / skpp adopter cloning the repo for the first time.

**Use Case**: "I just built skpp — show me it works." → `./skpp --list` and
`./skpp example` produce a real answer immediately, with no manual setup.

**Pain Points Addressed**: an empty `skills/` dir would make `--list` print
nothing and `skpp <tag>` always exit 1, giving no way to see the pipeline
working out of the box. The example is the canonical demo and the §13 acceptance
target.

## Why

- **PRD §2 constraint 4 / §17 — exactly ONE skill ships.** This file is it. The
  repo is a loader, not a skill library; do not add a second skill.
- **PRD §11 — the example is the §13 acceptance target.** The acceptance suite
  (P1.M6.T16.S1) asserts `./skpp --list` shows `example`, `test -d "$(./skpp
  example)"`, `test -f "$(./skpp -f example)"`, and `./skpp check` reports OK.
  Those assertions are impossible without this file.
- **De-risks the whole pipeline with real input.** `--list`/resolution/search/
  check are all exercised against a frontmatter that exercises the folded-scalar
  `description: >`, the `metadata` map, and a `keywords` flow-sequence — the
  trickiest parse cases — in one tiny, self-documenting file.
- **Non-auto-discovery proof.** Living at `<repo>/skills` (sibling-of-binary,
  PRD §8 rule 2), it loads ONLY via `pi --skill "$(skpp example)"`, never via pi
  auto-discovery (PRD §2 constraint 2). The file body even documents this usage.

## What

User-visible behavior (all against the freshly-built binary at repo root):

- `./skpp example` → prints one absolute path ending in `/skills/example`, exit 0.
- `./skpp -f example` → prints one absolute path ending in `/skills/example/SKILL.md`, exit 0.
- `./skpp --list` → prints a `TAG NAME DESCRIPTION` table whose first (and only)
  row is `example  example  Reference example skill for skpp…`, exit 0.
- `./skpp check` (once M4.T10/M5.T1 land) → prints an `OK` line for `example`,
  summary `1 skills, 0 errors, 0 warnings`, exit 0.
- `pi --skill "$(./skpp example)"` → pi loads the skill (the body tells pi what
  it is); never auto-discovered.

### Success Criteria

- [ ] `test -f skills/example/SKILL.md` and the file is byte-identical to PRD §11
      (outer four-backtick fence stripped; inner ```bash fence kept; 569 bytes).
- [ ] `./skpp example` → absolute path, exit 0 (no trailing slash, no stdout junk).
- [ ] `./skpp -f example` → `…/skills/example/SKILL.md`, exit 0.
- [ ] `./skpp --list` → shows the `example` skill exactly once, exit 0.
- [ ] `test "$(./skpp --path)" = "$PWD/skills"` (sibling-of-binary resolution).
- [ ] Probe (check-independent): `go run …/validate_example_probe.go
      skills/example/SKILL.md` prints `OK   example (example)` exit 0.
- [ ] `$PWD/skills` is NOT under `$HOME/.pi` (not a pi auto-discovery location).
- [ ] (If M4.T10+M5.T1 landed) `./skpp check` exit 0, an `OK   example` line.
- [ ] `git status` shows only `skills/example/SKILL.md` as a new file (plus the
      optional `.gitkeep` removal if it somehow reappeared) — no Go changes.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ **Yes.** The exact file content is given verbatim
below (including the single non-obvious step: stripping the PRD's outer
four-backtick `````markdown` fence). The verification commands are exact and were
all run against a temp copy of this exact content during research (see
`research/verified_facts.md`). No Go knowledge is required to implement — the
task is "write one markdown file, then run five commands."

### Documentation & References

```yaml
# MUST READ — the authoritative, byte-exact spec for this file (PRD §11)
- file: PRD.md
  section: "§11 ('The one shipped example skill') — the fenced code block whose
            inner content IS skills/example/SKILL.md."
  why: "This section is the single source of truth for the file's bytes. The
        content is rendered inside a FOUR-backtick `````markdown` fence so the
        inner THREE-backtick ```bash fence survives Markdown. The four-backtick
        wrapper is NOT part of the file."
  critical: "STRIP the outer `````markdown` line and the matching closing
             ````` line. KEEP the inner ```bash … ``` block verbatim — it is real
             file content pi renders on load. Do NOT copy the PRD sentence
             'No other skills ship in this repo.' into the file (it is PRD prose)."
  gotcha: "Reproducing the inner ```bash fence while authoring this PRP itself
           required a four-backtick outer fence here too — same nesting rule. The
           implementer just needs the FINAL file to contain exactly one top-level
           YAML frontmatter block (---…---) then Markdown with one ```bash block."

# MUST READ — every empirical fact behind this PRP (verified against the live repo)
- file: plan/001_fcde63e5bb60/P1M6T1S1/research/verified_facts.md
  why: "§1 exact 569-byte content + the fence-stripping rule; §2 the parsed
        frontmatter table (HasFM/name/description len 162/keywords/category) that
        proves §9 passes; §3 which verify cmds work TODAY vs. pending; §4 the
        check-independent probe; §5 .gitkeep is absent (no-op); §6 the non-pi-location
        guardrail; §7 'No other skills' is prose; §8 verified validation cmds."
  critical: "skpp check is NOT wired on disk today (internal/check from M4.T10 is
             absent; main.go has no runCheck — that's M5.T1 in parallel flight).
             Use the §4 probe as the check-independent gate. skpp check becomes a
             real gate only once BOTH M4.T10 and M5.T1 land (build-order steps 4+5,
             before this step 6)."

# CONTRACT — the parser that consumes this file (already landed + green)
- file: internal/discover/discover.go
  why: "ParseFrontmatter(): strips a UTF-8 BOM, finds the first two '---' lines,
        unmarshals the YAML block with gopkg.in/yaml.v3 (lenient: unknown keys
        ignored). Returns Frontmatter{HasFM:true} for this file. Verified on the
        exact §11 content: parses cleanly, no error."
  pattern: "folded scalar 'description: >' → a single string with newlines folded
            to spaces (162 chars here, well under the §9 1024 cap)."
- file: internal/discover/skill.go
  why: "BuildSkill(): extracts metadata.keywords (flow-seq [example,demo,skpp] →
        []string via toStringSlice), metadata.category ('meta' via comma-ok
        assertion), metadata.aliases (absent → nil). No code change needed."

# CONTRACT — the consumers (already landed) whose output the verify cmds assert on
- file: internal/skillsdir/skillsdir.go
  why: "Find(): §8 resolver. Rule 2 (sibling-of-binary) makes ./skpp resolve
        <repo>/skills. `./skpp --path` prints that dir; SKPP_SKILLS_DIR overrides."
- file: internal/resolve/resolve.go
  why: "Resolve(): `skpp example` → tag 'example' matches RelTag 'example' → prints
        Dir (absolute). `skpp -f example` → prints SourceFile (Dir/SKILL.md)."
- file: internal/ui/ui.go
  why: "PrintList(): renders the TAG/NAME/DESCRIPTION table for `--list`."

# REFERENCE — the spec the frontmatter conforms to (skpp check validates against it)
- file: plan/001_fcde63e5bb60/architecture/external_deps.md
  section: "§1 (Agent Skills spec): name regex ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$,
            len 1–64; description max 1024; metadata is a spec'd OPTIONAL field;
            lists under metadata are spec-compliant. Confirms this frontmatter is OK."
  why: "This is the source `skpp check` (M4.T10) validates against. The example
        passes every rule (verified in research §2)."

# TOOL — the check-independent validation probe (lives in research/, //go:build ignore)
- file: plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go
  why: "Imports the LANDED internal/discover package and asserts every §9 rule
        directly against the file — works whether or not skpp check is wired.
        Run: go run plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go skills/example/SKILL.md"

# URL — pi's --skill contract (the usage the example body documents)
- url: https://agentskills.io/specification
  why: "Confirms SKILL.md = YAML frontmatter + Markdown body; `metadata` optional;
        `name`/`description` required. (external_deps.md §1 already verified this.)"
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── go.mod                      # module github.com/dabstractor/skpp; go 1.25; yaml.v3
├── main.go                     # CLI (has example/-f/--list; check NOT yet wired)
├── main_test.go
├── internal/
│   ├── discover/{discover,skill,index,*_test.go}   # LANDED — parses this file
│   ├── resolve/{resolve,*_test.go}                  # LANDED — resolves the tag
│   ├── skillsdir/{skillsdir,*_test.go}              # LANDED — locates <repo>/skills
│   ├── ui/{ui,*_test.go}                            # LANDED — renders --list
│   ├── search/{search,*_test.go}                    # LANDED (M4.T9)
│   └── check/                                       # ABSENT (M4.T10, planned)
└── (no skills/ dir yet)        # <-- created fresh by this task
```

### Desired Codebase tree (file added)

```bash
skills/
└── example/
    └── SKILL.md                # NEW — byte-exact PRD §11 (569 bytes)
```
No other files. No Go changes. No package changes. No dependency changes.

### The exact file to create — `skills/example/SKILL.md`

Create this file with EXACTLY this content (the outer four-backtick fence shown
here is a rendering necessity of THIS document, **not** part of the file — the
file begins at `---` and ends at the final ``` line):

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

### Known Gotchas of our codebase & Library Quirks

```python
# CRITICAL — NESTED CODE FENCE. PRD §11 renders the file inside a FOUR-backtick
# `````markdown fence so the inner THREE-backtick ```bash fence survives Markdown
# rendering. The four-backtick wrapper is NOT part of the file. The implementer
# must emit: the `---` frontmatter, the Markdown body, and ONE inner ```bash …
# ``` block. Verified: the correct file is 569 bytes (research §1).

# CRITICAL — 'No other skills ship in this repo.' is PRD PROSE, not a file line.
# It appears in the PRD after the closing ````` of the rendered block. Do not
# write it into SKILL.md. (It IS a hard scope rule for this task: create EXACTLY
# ONE skill — no second skill.)

# CRITICAL — `skpp check` is NOT wired on disk today. It depends on BOTH
# internal/check (P1.M4.T10.S1, status Planned) AND the check-subcommand dispatch
# in main.go (P1.M5.T1.S1, parallel flight; main.go currently has NO runCheck —
# `grep runCheck main.go` is empty). Today `./skpp check` returns
# 'unknown skill tag "check"' exit 1 (it treats check as a tag).
# PER PRD §18 BUILD ORDER this task (step 6) runs AFTER check (step 4) and the
# dispatch (step 5), so if the orchestrator honors order, check WILL be wired by
# the time this runs. Either way, use the check-independent probe as the PRIMARY
# §9 gate (it imports the LANDED discover package — research §4), and treat
# `./skpp check` exit-0-with-OK as a BONUS assertion that simply passes once the
# upstream items land.

# GOTCHA — skills/.gitkeep does NOT exist. git ls-files shows no skills/ dir and
# no .gitkeep anywhere. The contract's "remove the throwaway skills/.gitkeep from
# P1.M1.T3.S1 if present" is therefore a NO-OP today. Keep it as a defensive
# `rm -f skills/.gitkeep` (harmless if absent); do NOT fail if it's missing.

# GOTCHA — the skills dir is located by §8 rule 2 (sibling-of-binary): the example
# MUST live at <repo>/skills/example/SKILL.md so `./skpp --path` == "$PWD/skills".
# Do NOT place it under ~/.pi/agent/skills, .pi/skills, or node_modules — those are
# pi auto-discovery locations (PRD §2 constraint 2). $PWD is /home/dustin/projects/skpp.

# GOTCHA — folded scalar `description: >` yields a trailing newline in the parsed
# string (discover copies it VERBATIM; ui wraps it for --list; check's 1024-char
# test counts 162 here). This is expected and correct — do NOT "fix" the
# description to a single-line scalar; PRD §11 uses `>` deliberately.
```

## Implementation Blueprint

### Data models and structure

None. This task adds a single Markdown file; it defines no Go types, no schemas,
no migrations. The data model it exercises (`discover.Frontmatter` /
`discover.Skill`) already exists and is consumed read-only.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: CREATE skills/example/SKILL.md
  - ACTION: write the file with EXACTLY the PRD §11 content shown in
            "The exact file to create" above.
  - RULE: strip the PRD's outer four-backtick `````markdown` fence and its
          matching closer; KEEP the inner ```bash … ``` block verbatim.
  - NAMING/PLACEMENT: directory MUST be skills/example/ (dir name == the tag ==
          the frontmatter `name`, per PRD §10 "name should match the directory
          where practical" — here it matches exactly).
  - CHECK: `wc -c skills/example/SKILL.md` → 569 (sanity; trailing newline
           included). `head -1 skills/example/SKILL.md` → `---`.
           `grep -c '^metadata:' skills/example/SKILL.md` → 1.
  - GOTCHA: do NOT add a trailing blank line beyond the file's natural end;
            do NOT reflow the description (the folded-scalar indentation matters).

Task 2: REMOVE the throwaway skills/.gitkeep (defensive, likely a no-op)
  - COMMAND: rm -f skills/.gitkeep
  - EXPECT: no error whether or not it existed (git ls-files shows it does NOT
            exist today, so this is expected to be a no-op).
  - WHY: PRD §11 replaces the placeholder; the contract requires it gone.
  - GOTCHA: do NOT remove any OTHER file; do NOT touch skills/example/SKILL.md.

Task 3: BUILD the binary
  - COMMAND: cd /home/dustin/projects/skpp && go build -o skpp . && echo OK
  - EXPECT: exit 0, a ./skpp binary appears. (This task adds no Go, so a build
            failure means an UPSTREAM item is broken — report, do not fix.)

Task 4: VERIFY — resolution + listing (all LANDED commands; PRIMARY gates)
  - COMMAND: test "$(./skpp --path)" = "$PWD/skills" && echo "PATH OK"
  - COMMAND: ./skpp example   ; test -d "$(./skpp example)" && echo "RESOLVE OK"
  - COMMAND: ./skpp -f example; test -f "$(./skpp -f example)" && echo "FILE OK"
  - COMMAND: ./skpp --list | grep -q '^example' && echo "LIST OK"
  - COMMAND: case "$(./skpp example)" in /*) echo "ABSOLUTE OK";; *) echo FAIL; exit 1;; esac
  - EXPECT: all five echo OK; ./skpp example prints an absolute path ending
            /skills/example with NO trailing slash; -f prints …/SKILL.md.

Task 5: VERIFY — frontmatter correctness (check-INDEPENDENT; PRIMARY §9 gate)
  - COMMAND: go run plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go skills/example/SKILL.md
  - EXPECT: exit 0; prints `OK   example (example)` then a line with
            HasFM=true keywords=[example demo skpp] category="meta" and
            description chars=162. (Imports the LANDED internal/discover; works
            regardless of whether skpp check is wired.)
  - NOTE: the probe carries `//go:build ignore` so it never affects `go build ./...`.

Task 6: VERIFY — skpp check (BONUS gate; passes once M4.T10+M5.T1 land)
  - COMMAND: ./skpp check; rc=$?
  - IF rc == 0 AND output contains 'OK': skpp check is wired and clean → DONE.
  - IF output is 'unknown skill tag "check"' (rc 1): check is NOT wired yet
            (M4.T10 absent / M5.T1 dispatch pending). This is NOT a failure of
            THIS task — Task 5 already proved §9 correctness. Record the
            'check pending' state and move on; the acceptance suite (P1.M6.T16.S1)
            re-runs `./skpp check` once upstream lands.
  - GOTCHA: do NOT implement internal/check or a runCheck stub here — those are
            owned by M4.T10 and M5.T1 respectively (PRD §9, §6.1).

Task 7: VERIFY — not a pi auto-discovery location + scope
  - COMMAND: case "$(./skpp --path)" in "$HOME"/.pi/*|*/.pi/skills/*|*/node_modules/*) echo FAIL; exit 1;; *) echo "LOCATION OK";; esac
  - COMMAND: test "$(find skills -name SKILL.md | wc -l)" -eq 1 && echo "ONE-SKILL OK"
  - EXPECT: LOCATION OK (path is <repo>/skills, sibling-of-binary); exactly ONE
            SKILL.md ships (PRD §2 constraint 4 / §17).

Task 8: FINAL — build/test sanity (this task touches no Go)
  - COMMAND: go build ./... && go test ./... -count=1
  - EXPECT: build exit 0; whole module green (this task added no .go files, so
            counts are unchanged). gofmt/go vet unaffected (no .go touched).
```

### Implementation Patterns & Key Details

```python
# PATTERN: the file is authored from PRD §11 verbatim, not "designed". The single
# mechanical step is stripping the outer `````markdown fence. Reproduce the bytes,
# do not paraphrase the description or reformat the keywords list.
#   WRITE: skills/example/SKILL.md  (569 bytes; see "The exact file to create")

# PATTERN: verify with the SHIPPED commands first (example/-f/--list), then the
# check-independent probe, and treat `skpp check` as a bonus that passes later.
# This ordering means this task's acceptance NEVER blocks on an upstream item.

# PATTERN: the probe is the bridge to `skpp check`. It asserts the SAME §9 rules
# (name regex, description present + ≤1024, HasFM, metadata extraction) using the
# LANDED discover package, so its `OK` line is equivalent to what `skpp check`
# will emit once M4.T10 lands.
```

### Integration Points

```yaml
DISK (the only integration surface):
  - add file: skills/example/SKILL.md   # PRD §11, byte-exact; 569 bytes
  - remove (if present): skills/.gitkeep   # absent today; defensive rm -f

CLI (consumed, not changed):
  - ./skpp example      → internal/resolve matches RelTag 'example' → Dir
  - ./skpp -f example   → same match → SourceFile (Dir/SKILL.md)
  - ./skpp --list       → internal/ui renders the Skill row
  - ./skpp check        → (pending M4.T10+M5.T1) internal/check over discover.Index

NO CHANGES TO:
  - any .go file ; go.mod / go.sum ; PRD.md ; tasks.json
  - internal/{discover,resolve,search,skillsdir,ui,check}/*
  - README.md ; install.sh ; completions/  (later M6 tasks)
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
# This task writes Markdown, not Go — gofmt/vet are sanity-only (no .go touched).
test -f skills/example/SKILL.md && echo "FILE EXISTS"
head -1 skills/example/SKILL.md            # MUST print: ---
grep -c '^metadata:' skills/example/SKILL.md   # MUST print: 1
grep -q '^```bash' skills/example/SKILL.md && echo "INNER FENCE PRESENT"
# Trailing fence present and balanced:
tail -1 skills/example/SKILL.md            # MUST print: ```
wc -c skills/example/SKILL.md             # ~569 (sanity)
go build ./...                             # MUST exit 0 (probe is //go:build ignore)
# Expected: file exists, frontmatter present, inner bash fence intact, build clean.
```

### Level 2: Frontmatter correctness (check-independent; PRIMARY §9 gate)

```bash
# Asserts every PRD §9 rule via the LANDED discover package — does NOT need skpp check.
go run plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go skills/example/SKILL.md
# Expected: exit 0; prints:
#   OK   example (example)
#     HasFM=true keywords=[example demo skpp] category="meta" aliases=[]
#     description chars=162 body_present=true
```

### Level 3: Integration Testing (the shipped CLI against the real file)

```bash
cd /home/dustin/projects/skpp
go build -o skpp . && echo "BUILD OK"

# Skills-dir resolution (sibling-of-binary, PRD §8 rule 2)
test "$(./skpp --path)" = "$PWD/skills" && echo "PATH OK"

# Tag resolution + -f modifier (PRD §7.2, §6.2)
test -d "$(./skpp example)"  && echo "RESOLVE-DIR OK"     # resolves to a real dir
test -f "$(./skpp -f example)" && echo "RESOLVE-FILE OK"   # -f prints SKILL.md path

# Absolute-path contract (PRD §2 constraint 3 / §13)
case "$(./skpp example)" in /*) echo "ABSOLUTE OK";; *) echo "FAIL"; exit 1;; esac

# Listing (PRD §6.1 --list)
./skpp --list | grep -q '^example' && echo "LIST OK"

# Location guardrail (NOT a pi auto-discovery location; PRD §2 constraint 2)
case "$(./skpp --path)" in "$HOME"/.pi/*|*/.pi/skills/*|*/node_modules/*) echo FAIL; exit 1;; *) echo "LOCATION OK";; esac

# Scope guardrail (exactly ONE skill ships; PRD §2 constraint 4 / §17)
test "$(find skills -name SKILL.md | wc -l)" -eq 1 && echo "ONE-SKILL OK"

# BONUS — skpp check (passes once M4.T10 + M5.T1 land; else 'unknown tag' is expected)
./skpp check; rc=$?
[ "$rc" = "0" ] && echo "CHECK OK (M4.T10+M5.T1 landed)" || echo "CHECK PENDING (use Level 2 probe; not a failure of this task)"
# Expected: PATH/RESOLVE-DIR/RESOLVE-FILE/ABSOLUTE/LIST/LOCATION/ONE-SKILL all OK.
```

### Level 4: End-to-end with pi (PRD §13 acceptance line)

```bash
cd /home/dustin/projects/skpp
# The example loads ONLY via --skill, never auto-discovered (PRD §2 constraint 2).
pi --no-skills --skill "$(./skpp example)" -p "briefly confirm the example skill is loaded" 2>&1 | head
# Expected: pi's output references the example skill / does not error. --no-skills
# proves we rely solely on the explicit --skill path. (If pi is unavailable in the
# run env, skip this level — it is informational, not a build gate.)
```

## Final Validation Checklist

### Technical Validation

- [ ] Level 1 passed: file exists, frontmatter present, inner ```bash fence intact, `go build ./...` clean.
- [ ] Level 2 passed: the probe prints `OK   example (example)` exit 0 (check-INDEPENDENT §9 gate).
- [ ] Level 3 passed: PATH / RESOLVE-DIR / RESOLVE-FILE / ABSOLUTE / LIST / LOCATION / ONE-SKILL all OK.
- [ ] `go build -o skpp .` exit 0; `go test ./...` green (no .go touched by this task).
- [ ] (Bonus) `./skpp check` exit 0 + OK line IF M4.T10+M5.T1 have landed; otherwise recorded as pending.

### Feature Validation

- [ ] `skills/example/SKILL.md` is byte-exact PRD §11 (outer four-backtick fence stripped).
- [ ] `./skpp example` / `./skpp -f example` / `./skpp --list` all exit 0 with correct output.
- [ ] Exactly ONE skill ships; the path is NOT under a pi auto-discovery location.
- [ ] No second skill, no Go changes, no go.mod/go.sum/PRD.md/tasks.json changes.

### Code Quality Validation

- [ ] File placement matches the desired tree (`skills/example/SKILL.md`).
- [ ] Description kept as a folded scalar `>` (not reflowed); keywords kept as a flow sequence.
- [ ] No trailing stray lines; no paraphrasing of PRD §11 prose into the file.

### Documentation & Deployment

- [ ] The example body is self-documenting (PRD §11 DOCS clause: "none — the example
      skill body is self-documenting"). README usage references it (Mode B, P1.M6.T14).
- [ ] No new env vars, no config changes.

---

## Anti-Patterns to Avoid

- ❌ Don't add a SECOND skill (PRD §2 constraint 4 / §17 / §11: "No other skills
  ship in this repo."). This task creates exactly `skills/example/SKILL.md`.
- ❌ Don't implement `internal/check` or a `runCheck` dispatch to make `./skpp
  check` pass — those are owned by M4.T10 and M5.T1. Use the probe for §9 validation.
- ❌ Don't paraphrase the PRD §11 description or reformat the keywords. Reproduce
  the bytes; the folded-scalar indentation is load-bearing for YAML.
- ❌ Don't copy the PRD sentence "No other skills ship in this repo." into the file
  (it is PRD prose, not file content).
- ❌ Don't place the skill under `~/.pi/agent/skills`, `.pi/skills`, or
  `node_modules` (pi auto-discovery locations, PRD §2 constraint 2). It lives at
  `<repo>/skills/example/`.
- ❌ Don't "fix" the folded-scalar trailing newline in the parsed description —
  discover copies it verbatim by contract; `--list` wraps it; the 1024-cap test
  counts 162 here.
- ❌ Don't fail the task if `skills/.gitkeep` is absent (it is absent today); the
  `rm -f` is defensive.
