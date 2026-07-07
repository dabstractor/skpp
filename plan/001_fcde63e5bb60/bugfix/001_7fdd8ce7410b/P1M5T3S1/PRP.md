# PRP — P1.M5.T3.S1 (bugfix): Sync README.md to the shipped changeset

> **Subtask:** P1.M5.T3.S1 — Spec Alignment & Documentation (bugfix milestone M5,
> **Mode B changeset-level documentation sweep** per SOW §5). README.md mirrors the
> mcpeepants README structure (PRD §15 outline). The changeset shipped five
> behavioral fixes; **only README.md is stale** — `main.go`'s `usageText` is
> already synced (it lists all 6 `--search` fields). This PRP brings README.md in
> line with the binary's actual behavior.
>
> **Scope:** ONE file — `README.md`. **Documentation only. No code. No tests. No
> new files.** `PRD.md` is READ-ONLY/human-owned (decision D7) and is NOT modified.
>
> **DEPENDENCY (PARALLEL CONTEXT):** All implementing subtasks are COMPLETE
> (P1.M1.T1.S1 `--path` source, P1.M2.T1.S1 `--search` aliases/category,
> P1.M3.T1.S1 unicode tables, P1.M4.T1.S1 combined shorts/`=value`,
> P1.M4.T2.S1 listing-mode exclusivity). P1.M5.T2.S1 (parallel) edits
> `internal/check/check.go` only; P1.M5.T1.S1 (done) edited `.gitignore`. ⇒
> **Zero file conflict. README.md is touched by no other in-flight item.**

---

## Goal

**Feature Goal**: `README.md` makes **no stale claim** about `--path`,
`--search` fields, flag syntax, or listing-mode behavior. Every README statement
about those four topics matches what `./skpp` actually does in the shipped
binary, verified by spot-checking `--path`, `--search`, and `--help`.

**Deliverable**: An edited `README.md` (5 targeted edits: 3 required, 2
optional). No other file changes. README keeps its existing tone and structure
(PRD §15 outline) — edits are surgical, not a rewrite.

**Success Definition**:
- The `--path` docs say it prints the dir to stdout AND the discovery rule
  (env var / sibling of binary / ancestor of cwd) to stderr, and explain why the
  stderr label matters (typo'd `SKPP_SKILLS_DIR` falls through silently).
- The `--search` docs list all six searchable fields: tag / name / description /
  keywords / aliases / category.
- No README sentence asserts the pre-changeset behavior (e.g. "`--path` reports
  which directory won" with no source rule; `--search` with no field list).
- `go build -o skpp .` succeeds and the §13 acceptance commands in Validation
  Loop Level 3 reproduce the README's claims verbatim.
- Only `README.md` is modified; `PRD.md` is byte-identical.

## User Persona

**Target User**: A pi user who reads README.md to learn `skpp`, then runs the
commands it shows. Two failure modes this task prevents: (1) a user typos
`SKPP_SKILLS_DIR`, sees a valid-looking `--path` result, and cannot tell the env
var was ignored; (2) a user adds `metadata.aliases`/`category`, runs
`--search <alias>`, and (per the stale README) does not expect a match.

**Use Case**: User runs `./skpp --path` and `./skpp --search`, compares the
output to README, and the two agree.

**Pain Points Addressed**: README currently under-documents `--path` (no source
label) and `--search` (no field list), so the binary appears to do "more" than
the docs claim — eroding trust and hiding the typo-fall-through diagnostic.

## Why

- **README is the contract users trust.** The code and `--help` are already
  correct; README is the last stale surface for this changeset (Mode B = sweep
  the human-facing doc so it reflects the shipped, coherent state).
- **QA Issue 1 (`--path` source) and Issue 4 (`--search` fields) were the two
  spec/impl gaps the QA pass flagged as user-visible.** Both ship a user-facing
  behavior; both deserve a one-line README note so the feature is discoverable.
- **Zero risk.** Markdown edits only. The build + `--help`/`--path`/`--search`
  spot-checks are the safety net that proves README ↔ binary agreement.

## What

Edit **only** `README.md`. Five edits, applied to the sections the contract
names (item LOGIC a–d). Each is a small, surgical change; **do not restructure
the document** (PRD §15 outline is the structure; preserve headings, order, tone).

- **Edit A (REQUIRED — contract b):** In the *Usage* commented block, add the
  six searchable fields to the `--search` example.
- **Edit B (REQUIRED — contract a):** In the *Usage* commented block, note that
  `--path` also prints the discovery rule to stderr.
- **Edit E (REQUIRED — contract a):** In *How `skpp` finds the store*, rewrite
  the one stale sentence "`skpp --path` reports which directory won." to cover
  stdout+stderr, the three labels, and the typo-fall-through rationale.
- **Edit C (OPTIONAL — contract c):** In *Usage*, add a one-line note that short
  flags combine and `--flag=value` is accepted (with valid example combos only).
- **Edit D (OPTIONAL — contract d):** In *Usage*'s **Error contract.** paragraph,
  note listing modes are mutually exclusive. (README does not currently document
  mode exclusivity; this is additive polish, not a stale-claim fix.)

### Do NOT

- Do NOT edit `PRD.md` (READ-ONLY; §6.1 still lists the OLD 4-field search list;
  the QA Issue 4 resolution = §10 wins and code matches §10 — README reflects the
  **shipped** 6-field behavior, NOT §6.1's stale summary; do not "fix" §6.1).
- Do NOT add a README claim about unicode table width (Issue 2). README makes no
  ASCII/byte-width claim today, so there is nothing stale to fix; the change is
  invisible to users. Adding a "tables handle unicode" note is out of scope and
  would be marketing-y. Leave README silent on it (see research §3).
- Do NOT touch any `.go` file, any test, `.gitignore`, `install.sh`,
  `completions/*`, or the example skill.
- Do NOT renumber/restructure README headings (PRD §15 outline is the structure).

### Success Criteria

- [ ] README's `--search` text names all six fields (incl. aliases + category).
- [ ] README's `--path` text names the stderr source label + the three rule names.
- [ ] README's `--path` text explains the typo'd-`SKPP_SKILLS_DIR` fall-through.
- [ ] (If Edit C/D applied) flag-syntax and/or mode-exclusivity notes are present.
- [ ] `./skpp --path`, `./skpp --search <alias>`, and `./skpp --help` reproduce
      the README's claims (manual spot-check, item TESTS step).
- [ ] `go build -o skpp .` succeeds; only `README.md` changed; `PRD.md` untouched.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ **Yes.** Every edit below gives exact
`oldText` (verbatim from the current README) and exact `newText`. The shipped
binary output that each edit must match is quoted in `research/verified_facts.md`
(§1 the `(found via …)` stderr line + the three `Source.String()` labels; §2 the
six search fields; §5 the exclusivity error string). The implementer needs no
prior skpp knowledge — apply the five edits, run the four validation commands,
confirm README ↔ binary agreement.

### Documentation & References

```yaml
# MUST READ — the ONLY file under edit
- file: README.md
  why: "The deliverable. Lines 129-131 (Usage --search), 139-140 (Usage --path),
        ~150-154 (Error contract paragraph), 158 (`--help` sentence), 200-212
        (Adding a skill frontmatter ex), 243 (How skpp finds the store)."
  pattern: "Markdown; fenced ```bash blocks hold commented examples where each
            line is `command   # comment`. Preserve the `→` and `…` glyphs and
            existing comment-alignment where present."
  gotcha: "Line 140 uses the Unicode ellipsis `…` and arrow `→` — copy them
           verbatim into oldText or the match will fail. The Usage block is a
           fenced code block; do not re-flow it."

# MUST READ — the binary's actual --path behavior (Edit B + E source of truth)
- file: main.go
  section: "the `if c.path {` branch — `fmt.Fprintln(stdout, dir)` then
            `fmt.Fprintf(stderr, \"(found via %s)\\n\", src)`."
  why: "Proves stdout is byte-identical to before (§13 gate holds) and the source
        rides on STDERR. Edit E must state stdout+stderr, not conflate them."
  critical: "stdout carries ONLY the dir+newline (the §13 `$(...)` contract).
             The source label is stderr-only. README must not imply the label
             appears in `$(skpp --path)`."

# MUST READ — the three source labels (verbatim strings for Edit E)
- file: internal/skillsdir/skillsdir.go
  section: "Source.String() switch."
  why: "Edit E must name the EXACT labels: `SKPP_SKILLS_DIR`, `sibling of
        binary`, `ancestor of cwd` (capitalization/wording as-is)."
  gotcha: "The env-var label is the variable NAME `SKPP_SKILLS_DIR`, not a prose
           phrase. The walk-up label is `ancestor of cwd`, not `walk-up`."

# MUST READ — the six --search fields (verbatim field list for Edit A)
- file: internal/search/search.go
  section: "matches() — six OR'd Contains checks: RelTag, Name, Description,
            each Keyword, each Alias, Category."
  why: "Edit A's field list must be exactly: tag / name / description / keywords
        / aliases / category."
  critical: "Aliases + category are NEW (Issue 4 fix). README must list them. The
             `--help` text already lists all six (`Substring search over tag /
             name / description / keywords / aliases / category`) — README is the
             stale surface."

# MUST READ — the exclusivity error string (for optional Edit D)
- file: main.go
  section: "exclusivityError() — `n >= 2` over {path,list,searchMode,all}."
  why: "If Edit D is applied, the README note must match the actual message:
        `listing modes --path/--list/--search/--all are mutually exclusive` (exit 2)."

# MUST READ — the README structure this task must NOT alter
- file: PRD.md
  section: "§15 README.md outline (8 sections: Title, Why, Install, Usage,
            Where skills live, Adding a skill, How skpp finds the store,
            Constraints)."
  why: "README must stay in this section order/tone. Edits are in-section only.
        PRD.md is READ-ONLY — do not edit (notably §6.1's stale 4-field list)."

# MUST READ — this task's own verified binary output + line map
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/P1M5T3S1/research/verified_facts.md
  why: "§1 the `(found via sibling of binary)` stderr output + the 3 labels;
        §2 the 6 fields + proof `--help` already lists them; §4 valid vs
        conflict-prone flag-syntax examples (use `--search=reddit` / `-af`;
        AVOID `-pl`/`-la`/`-vh`); §5 the exclusivity exit-2 output; §6 the
        README line-by-line map of every stale/accurate claim; §8 conflict check."
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── README.md          # EDIT — 5 surgical edits (A,B required; C,D optional; E required)
├── main.go            # READ-ONLY reference (usageText ALREADY lists 6 fields; --path stderr wire-up)
├── internal/skillsdir/skillsdir.go   # READ-ONLY reference (Source.String() labels)
├── internal/search/search.go         # READ-ONLY reference (matches() = 6 fields)
├── PRD.md             # READ-ONLY (human-owned; §15 outline = README structure; do NOT edit §6.1)
└── … (.gitignore, internal/*, install.sh, completions/, skills/ — UNCHANGED; no conflict)
```

### Desired Codebase tree (file responsibility)

```bash
README.md   # In-sync with shipped behavior: --path source rule, --search 6 fields, flag syntax, mode exclusivity.
```
No new files. No code. No tests. No other docs. Mode B = the README IS the deliverable.

### Known Gotchas of our codebase & Library Quirks

```markdown
# CRITICAL — README is the ONLY deliverable. Do not edit any .go file, any test,
# PRD.md, .gitignore, install.sh, completions/, or the example skill.

# CRITICAL — README ≠ PRD §6.1 on search fields, and that is CORRECT. PRD §6.1
# still lists the OLD 4 fields (tag/name/description/keywords); QA Issue 4's
# resolution (decisions.md §D4) made §10 win → the SHIPPED code searches 6 fields.
# README must describe the SHIPPED 6-field behavior. Do NOT edit PRD §6.1 to make
# the docs "agree" — PRD is human-owned/read-only (D7).

# GOTCHA — Unicode glyphs in the Usage block. Line 140 is:
#   skpp --path                        # → /…/skills
# The `→` (U+2192) and `…` (U+2026) are real characters. Copy them verbatim into
# oldText; an ASCII `->`/`...` substitute will fail to match.

# GOTCHA — flag-syntax examples must be VALID combos (no false examples).
# VALID (single listing mode + modifier, or = form): --search=reddit, -af, --all --file.
# CONFLICT-PRONE (would exit 2 or short-circuit): -pl (path+list → exit 2),
# -la (list+all → exit 2), -vh (help wins → prints help, surprising).
# Edit C must use only VALID examples.

# GOTCHA — do not claim anything about unicode table alignment in README. The
# code handles it (Issue 2 fix), but README never asserted otherwise, so there is
# no stale claim to fix. Adding a "unicode-safe tables" line is out of scope and
# reads as marketing. Leave README silent on it (research §3).

# NO build/config involvement. go.mod/go.sum untouched. `go build -o skpp .` is
# only run to produce the binary for the Level 3 spot-check (./skpp is gitignored).
```

## Implementation Blueprint

### Data models and structure

None. Markdown documentation edit — no models, no types, no code, no config.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: RE-READ the exact current README text (do not trust memory)
  - COMMAND: sed -n '125,160p' README.md ; sed -n '240,246p' README.md
  - WHY: confirm the verbatim oldText for each edit BEFORE applying (whitespace,
         the `→`/`…` glyphs, and comment alignment must match exactly). If the
         line content differs from the oldText quoted below (e.g. a prior edit
         landed), STOP and reconcile against the actual file, not this PRP.

Task 1 (REQUIRED, contract b): EDIT A — Usage: --search field list
  - FILE: README.md, the *Usage* fenced ```bash block.
  - FIND: the three lines
      # Human-readable catalog and substring search
      skpp --list
      skpp --search reddit
  - REPLACE with the "newText" in Exact Edit A below (adds a `# matches …`
    comment listing all six fields). This is the searchable-fields list update.
  - DO NOT: change the `--list` line, surrounding examples, or alignment beyond
            the added comment.

Task 2 (REQUIRED, contract a): EDIT B — Usage: --path stderr label
  - FILE: README.md, same *Usage* block, two lines down from Edit A.
  - FIND:
      # Where is the resolved skills directory?
      skpp --path                        # → /…/skills
  - REPLACE with the "newText" in Exact Edit B below (notes the stderr source
    label). Preserves the `→` and `…` glyphs.
  - DO NOT: alter the stdout path (it is the §13 contract) or imply the label
            appears on stdout.

Task 3 (REQUIRED, contract a): EDIT E — How skpp finds the store
  - FILE: README.md, *How `skpp` finds the store* section, the line just before
          *Constraints* (currently: `` `skpp --path` reports which directory won.``).
  - FIND: that single line.
  - REPLACE with the "newText" in Exact Edit E below (covers stdout+stderr, the
    three exact labels, and the typo-fall-through rationale).
  - DO NOT: reword the §8 1–4 priority list above it (it is accurate and in-scope-
            adjacent; only the trailing `--path` sentence is stale).

Task 4 (OPTIONAL, contract c): EDIT C — Usage: flag-syntax note
  - FILE: README.md, *Usage* block, immediately after the `skpp --version` line
          (last example), before the blank line + `**Error contract.**`.
  - FIND:
      # Version is the git-describe value (dynamic, not a fixed string)
      skpp --version
  - REPLACE by appending the "newText" in Exact Edit C below (one commented
    example line + a short prose sentence). Uses ONLY valid combos
    (`--search=reddit`, `-af`).
  - SKIP IF: you judge it bloats the Usage block. It is optional per the contract.

Task 5 (OPTIONAL, contract d): EDIT D — Error contract: mode exclusivity
  - FILE: README.md, the `**Error contract.**` paragraph just after the Usage
          fenced block.
  - FIND: the paragraph ending "…so `pi` never sees a partial result."
  - REPLACE by appending one sentence (Exact Edit D below) noting the four
    listing modes are mutually exclusive (exit 2).
  - SKIP IF: README is intentionally silent on modes. It is additive polish.

Task 6: VALIDATE (Mode B = manual review + build gate)
  - BUILD: go build -o skpp .
  - SPOT-CHECK (must match README claims):
      ./skpp --path                      # stdout: <dir>; stderr: (found via …)
      ./skpp --path 2>&1 >/dev/null      # shows ONLY the (found via …) line
      ./skpp --help | grep -i search     # shows the 6-field list
      ./skpp --search=reddit >/dev/null  # =value form works (exit 0 or 1, not 2)
  - DIFF GUARD: git diff --name-only  → only README.md
                 git diff --quiet PRD.md && echo "PRD.md untouched OK"
```

### Implementation Patterns & Key Details

The five edits are given verbatim below. `oldText` is copied from the current
README; apply each with the `edit` tool (or an equivalent exact replacement). If
`Task 0` shows a line has already drifted, reconcile to the file, not this text.

```markdown
=== EXACT EDIT A (REQUIRED) — Usage: --search field list ===
oldText:
# Human-readable catalog and substring search
skpp --list
skpp --search reddit
newText:
# Human-readable catalog and substring search
skpp --list
skpp --search reddit            # matches tag / name / description / keywords / aliases / category

=== EXACT EDIT B (REQUIRED) — Usage: --path stderr label ===
oldText:
# Where is the resolved skills directory?
skpp --path                        # → /…/skills
newText:
# Where is the resolved skills directory? (its discovery rule prints to stderr)
skpp --path                        # → /…/skills (stderr: found via sibling of binary)

=== EXACT EDIT C (OPTIONAL) — Usage: flag-syntax note ===
oldText:
# Version is the git-describe value (dynamic, not a fixed string)
skpp --version
newText:
# Version is the git-describe value (dynamic, not a fixed string)
skpp --version

# Short flags combine (-af) and long flags accept --flag=value (--search=reddit)
newText (prose form, if you prefer appending to the `--help` sentence instead):
  Append to the line `` `skpp --help` lists every flag.`` :
  `` `skpp --help` lists every flag. Short flags combine (`-af`) and long flags
  accept `--flag=value` (`--search=reddit`); the canonical forms above are what
  `--help` documents.``
  (Pick ONE of the two placements; do not add both.)

=== EXACT EDIT D (OPTIONAL) — Error contract: mode exclusivity ===
oldText:
**Error contract.** An unknown tag prints **nothing** to stdout and exits 1
(the error goes to stderr only). That is why
`pi --skill "$(skpp badtag)"` fails loudly instead of loading nothing. When
multiple tags are given, any unresolved tag causes nothing to be printed and
exit 1, so `pi` never sees a partial result.
newText:
**Error contract.** An unknown tag prints **nothing** to stdout and exits 1
(the error goes to stderr only). That is why
`pi --skill "$(skpp badtag)"` fails loudly instead of loading nothing. When
multiple tags are given, any unresolved tag causes nothing to be printed and
exit 1, so `pi` never sees a partial result. The listing modes `--path`,
`--list`, `--search`, and `--all` are mutually exclusive — combining any two
exits 2.

=== EXACT EDIT E (REQUIRED) — How skpp finds the store ===
oldText:
`skpp --path` reports which directory won.
newText:
`skpp --path` reports the winning directory on stdout and the matching rule on
stderr — one of `SKPP_SKILLS_DIR`, `sibling of binary`, or `ancestor of cwd`.
The stderr label matters when `SKPP_SKILLS_DIR` is typo'd: a bad value is
silently ignored and discovery falls through to the sibling / walk-up rule, so
the `(found via …)` line is the only way to tell the env var was skipped.
```

### Integration Points

```yaml
DOCUMENTATION:
  - file: README.md
  - edits: A, B (Usage block); E (How skpp finds the store); C, D (Usage, optional)
  - side-effects: none (markdown only; no behavior, no API, no build, no config)

NO OTHER INTEGRATION POINTS:
  - no database, no config, no routes, no dependency, no test, no help text.
  - main.go usageText is ALREADY correct — do not touch it.
  - PRD.md is read-only (D7); do NOT edit §6.1 or §15.
  - no conflict with P1.M5.T2.S1 (check.go) or P1.M5.T1.S1 (.gitignore).
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
# Markdown has no compiler. The only "syntax" check is: did the fenced code
# blocks survive intact (no broken ``` fences, no mangled indentation)?
# Re-render sanity: confirm the Usage block is still a single fenced ```bash block.
awk '/^```bash$/{f=1;c++} f{print} /^```$/{if(f){f=0}} END{print "fenced blocks:", c}' README.md
# Expected: "fenced blocks: <N>" where N matches the pre-edit count (edits are
# in-place inside existing blocks; you should NOT have added or removed a fence).
# Also confirm no stray ``` leaked mid-paragraph:
grep -c '^```' README.md   # Expected: an EVEN number (every open fence is closed).
```

### Level 2: Stale-Claim Grep (the real "unit test" for a doc task)

```bash
cd /home/dustin/projects/skpp
# Each grep must print NOTHING (the stale claim is gone). If any prints a line,
# that edit did not land — redo it.
grep -n -- '`skpp --path` reports which directory won' README.md   # Expected: empty
grep -n -- '# → /…/skills$' README.md                               # Expected: empty (now has stderr note)
# Each grep below must print AT LEAST one line (the new claim is present):
grep -n 'aliases / category' README.md                              # Edit A landed
grep -n 'found via' README.md                                       # Edit B landed
grep -n 'ancestor of cwd' README.md                                 # Edit E landed
grep -n 'SKPP_SKILLS_DIR`' README.md | grep -iq found && echo "labels OK" || true
```

### Level 3: README ↔ Binary Agreement (Integration / item TESTS step)

```bash
cd /home/dustin/projects/skpp
# Build the binary the README describes (./skpp is gitignored; build is expected).
go build -o skpp . && echo "build OK"

# (a) --path: README says dir on stdout, "(found via …)" on stderr. Verify BOTH.
./skpp --path                       # Expected stdout: an absolute …/skills path
./skpp --path 2>&1 >/dev/null       # Expected: a line "(found via <one of the 3 labels>)"
# Confirm the label is one of the three README names:
./skpp --path 2>&1 >/dev/null | grep -Eq '\(found via (SKPP_SKILLS_DIR|sibling of binary|ancestor of cwd)\)' \
  && echo "--path label OK" || echo "--path label MISMATCH"

# Typo'd-env fall-through (the rationale Edit E explains):
SKPP_SKILLS_DIR=/typo/not/real ./skpp --path 2>&1 >/dev/null   # Expected: NOT "(found via SKPP_SKILLS_DIR)"
# (It falls through to sibling/walk-up — exactly the silent-fall-through Edit E warns about.)

# (b) --search: README says it matches aliases + category. Verify with the shipped example.
#     The example skill (skills/example) has metadata.category: meta and no alias; add a
#     throwaway alias in a TEMP copy to prove alias matching without touching the repo skill:
T=$(mktemp -d); mkdir -p "$T/skills/aliased"
printf -- '---\nname: aliased\ndescription: x\nmetadata:\n  aliases: [my-alias]\n  category: tmpcat\n---\n\n# x\n' > "$T/skills/aliased/SKILL.md"
SKPP_SKILLS_DIR="$T/skills" ./skpp --search my-alias | grep -q aliased && echo "--search alias OK"
SKPP_SKILLS_DIR="$T/skills" ./skpp --search tmpcat  | grep -q aliased && echo "--search category OK"
rm -rf "$T"

# (c) --help: README points users here; confirm it lists all six fields (already true).
./skpp --help | grep -qi 'tag / name / description / keywords / aliases / category' && echo "--help fields OK"

# (d) optional edits, if applied: confirm the behaviors the README now mentions.
./skpp --search=reddit >/dev/null 2>&1; echo "search=value exit=$? (0 or 1, NOT 2)"
./skpp --list --search foo 2>&1 | grep -q 'mutually exclusive' && echo "mode exclusivity message OK"
```

### Level 4: Creative & Domain-Specific Validation

```bash
cd /home/dustin/projects/skpp
# Doc task: the only "domain" check is that README no longer contradicts the binary.
# Render the README and skim the three edited sections for tone/structure:
sed -n '/^## Usage$/,/^## Where skills live$/p' README.md | head -60
sed -n '/^## How .*finds the store$/,/^## Constraints$/p' README.md
# Expected: edits read as native prose (no em-dash spam, no "now supports!" marketing
# tone), the §15 section order is unchanged, and fenced ```bash blocks are intact.

# Collateral sanity — nothing else moved:
git diff --name-only                          # Expected: README.md (only)
git diff --quiet PRD.md && echo "PRD.md untouched OK"
git diff --quiet main.go internal/ .gitignore install.sh && echo "code untouched OK"
```

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — fenced ```bash block count unchanged; even number of ``` lines.
- [ ] Level 2 PASS — stale-claim greps empty; new-claim greps non-empty.
- [ ] Level 3 PASS — `go build -o skpp .` OK; `--path` shows dir(stdout)+label(stderr);
      `--search <alias>` and `--search <category>` match; `--help` lists 6 fields.
- [ ] Level 4 PASS — edited sections read natively; §15 order intact.

### Feature Validation
- [ ] README `--search` text lists all six fields (incl. aliases + category). [Edit A]
- [ ] README `--path` text (Usage) notes the stderr source label. [Edit B]
- [ ] README `--path` text (How it finds the store) names the three labels + the
      typo-fall-through rationale. [Edit E]
- [ ] (If applied) README notes combined shorts / `--flag=value`. [Edit C]
- [ ] (If applied) README notes listing-mode mutual exclusivity. [Edit D]
- [ ] Manual spot-check (`./skpp --path`, `./skpp --search`, `./skpp --help`)
      reproduces every README claim (item TESTS step).

### Code Quality Validation
- [ ] Only `README.md` modified; `PRD.md` byte-identical.
- [ ] No `.go` file, test, `.gitignore`, `install.sh`, completion, or skill touched.
- [ ] README tone/structure (PRD §15) preserved — edits are surgical, in-section.
- [ ] No marketing tell-words ("now supports!", "powerful", em-dashes) introduced.

### Documentation & Deployment
- [ ] Mode B: README.md IS the deliverable — no separate doc file written.
- [ ] No claim added about unicode table width (out of scope; nothing stale to fix).

---

## Anti-Patterns to Avoid

- ❌ Don't edit `PRD.md` (READ-ONLY; human-owned; D7). Notably do NOT "fix" §6.1's
  stale 4-field search list — README reflects the SHIPPED 6-field behavior, and §6.1
  divergence is a known, accepted spec/impl state (QA Issue 4 / decisions.md §D4).
- ❌ Don't touch any `.go` file or `usageText` — `--help` is ALREADY synced; only
  README is stale. Editing code is out of scope and would create merge conflict
  risk with the (complete) implementing subtasks.
- ❌ Don't add a unicode-table-width claim to README. Issue 2 was fixed in code, but
  README never asserted ASCII-only, so there is no stale claim. Adding one is scope
  creep + marketing tone. Leave README silent (research §3).
- ❌ Don't give a conflict-prone flag-syntax example in Edit C. `-pl`, `-la`, `-vh`
  exit 2 or short-circuit; use only `--search=reddit` and `-af` (research §4).
- ❌ Don't conflate `--path` stdout and stderr in Edit E — the §13 contract
  (`$(skpp --path)` = the dir) depends on the label staying on STDERR. README must
  not imply the `(found via …)` line is captured by `$(...)`.
- ❌ Don't restructure README (renumber/rename sections). PRD §15 outline is the
  structure; edits are surgical, in-section only.
- ❌ Don't skip the build + spot-check (Level 3). A README that compiles in your
  head but contradicts the binary fails the item's TESTS step.
