# PRP — P1.M2.T5.S1: Add a reserved-tag-names note to the README skill-tag section (Issue 7)

> **Subtask:** a single, surgical documentation edit to `README.md`'s "Where skills live"
> section (README.md:174–200). It adds a short "**Reserved tag names.**" note stating that
> `check` and `init` are reserved subcommand names that never resolve as skill tags, and
> lists the four workarounds (a skill at `skills/check/SKILL.md` is still discoverable via
> `--list`/`--all`, by a nested path like `writing/check`, by its frontmatter `name`, or by
> an alias). **No code change. No Go test. No `PRD.md` edit. No new files.** The note is the
> user-facing mirror of the two `parseArgs` code comments that already document this reservation
> as deliberate (`main.go:254–262` `case "check"`, `main.go:269–290` `case "init"`).
>
> **Why documentation-only (Mode A):** architecture/decisions.md **§D6** is the binding decision —
> "`check`/`init` reservation is deliberate and documented in code. A code change to resolve a skill
> named `check` would silently shadow the `skilldozer check` subcommand — a worse UX surprise.
> Decision: record the reservation + workarounds … and **surface it in the README's skill-tag
> section**. PRD.md is human-owned/READ-ONLY — the architect does NOT edit it." PRD §7.2 defines the
> canonical tag as the relTag with no carve-out; this note documents the one deliberate carve-out so
> the spec, the code, and the user docs all agree. The note is the canonical home for this rule; the
> later Mode B sweep (P1.M3.T1) verifies README consistency across all 7 fixes but does NOT duplicate
> it (contract DOCS step 5).
>
> **STATUS (verified at PRP-write time):** README is 298 lines; "Where skills live" = README.md:174–200,
> next section `## Adding a skill` at :201. The insertion seam (after `… all resolve.` at :199, before
> `## Adding a skill` at :201) is unique and confirmed byte-for-byte with `cat -A` (plain `\n`, no
> trailing spaces). README voice = bold lead-in callouts (`**Error contract.**` at :164 is the template)
> and **zero** `§`-PRD citations (`grep -n '§' README.md` → NONE → the note must NOT cite §8.2/§9).
> Parallel sibling P1.M2.T4.S1 edits ONLY `.gitignore`; this task edits ONLY `README.md` → **zero
> file-level overlap**, land in either order. No markdown linter, no Makefile doc target → validation
> is grep + render check + `go build/vet/test` stays green. All numbers reproducible (see
> research/verified_facts.md).

---

## Goal

**Feature Goal**: Document, in `README.md`'s "Where skills live" section, the deliberate rule that
`check` and `init` are reserved subcommand names and therefore never resolve as skill tags — so a user
who creates `skills/check/SKILL.md` (canonical tag `check`) is not surprised that `skilldozer check`
runs validation instead of resolving the skill, and knows the four ways to still reach that skill.

**Deliverable**: ONE edited file, `/home/dustin/projects/skilldozer/README.md`, with a single new
"**Reserved tag names.**" paragraph inserted at the end of the "Where skills live" section (after the
`… all resolve.` paragraph at README.md:199, before the `## Adding a skill` heading at README.md:201).
No other file is created or modified.

**Success Definition**: a user reading the README's tag section can (a) state that `check`/`init` are
reserved and why, and (b) name the four workarounds for a skill whose canonical tag collides — without
reading the code. The note matches the README's existing voice (bold lead-in callout, plain prose, no
`§`-citations, inline `` `code` `` for commands/paths). `go build/vet/test ./...` remains green
(proof of "no code change"); `git status --short` shows only `README.md` modified.

## User Persona (if applicable)

**Target User**: a skill author or end user who creates or addresses a skill and hits the (rare) case
where the skill's canonical tag is literally `check` or `init`.

**Use Case**: A user drops a skill at `skills/check/SKILL.md` and runs `skilldozer check`, expecting
the skill's path. Instead validation runs. The README note explains why and what to do instead.

**User Journey**: (today) `skilldozer check` silently runs validation (by design) and the README gives
no hint that `check` is reserved → confusion → (after) the tag section states up front that
`check`/`init` are reserved subcommand names and lists the four ways to still reach the skill.

**Pain Points Addressed**: the gap between PRD §7.2 (canonical tag = relTag, no carve-out stated) and
the implemented behavior (`check`/`init` reserved as subcommands); a user has no README-level signal
that those two names are special.

## Why

- **Closes architecture/bug_fixes_validation.md §ISSUE 7 (Minor)** and the bugfix PRD h3.6: a skill
  whose canonical tag is literally `check` or `init` cannot be resolved by that tag. The chosen
  resolution (decisions.md §D6) is documentation-only — the reservation is deliberate and a code change
  would shadow a real subcommand (a worse surprise).
- **decisions.md §D6 is explicit and binding:** "record the reservation + workarounds … and surface it
  in the README's skill-tag section. PRD.md is human-owned/READ-ONLY." So the README note is the one
  action this subtask takes; it is not optional and it is not a code change.
- **Makes spec, code, and docs agree.** PRD §7.2 (PRD.md:161, canonical tag = relTag) says nothing about
  reserved names; the code reserves them (main.go:254 `case "check"`, main.go:269 `case "init"`, both
  matched before the default tag-capture branch at main.go:281+). The note is the user-facing statement
  of that carve-out so the three artifacts no longer appear to contradict.
- **Cheap and isolated:** one README paragraph, no code, no test, no deps. Zero regression risk to the
  Go build/test surface. Disjoint from the parallel sibling (P1.M2.T4.S1 → `.gitignore` only).

## What

A single new paragraph (a bold-lead-in callout) appended to the "Where skills live" section of
`README.md`, immediately after the existing `… all resolve.` paragraph and immediately before the
`## Adding a skill` heading. The note:

1. Names the rule: `check` and `init` are subcommand names, so they never resolve as skill tags —
   `skilldozer check` runs validation, `skilldozer init` runs first-run setup.
2. Frames it as the standard CLI subcommand-reservation rule (a subcommand name takes precedence over a
   positional argument) — echoing the code comment "subcommand names are reserved, as in any CLI"
   (main.go:257) and the contract's "standard CLI subcommand-reservation rule".
3. Lists the four workarounds for a colliding skill (`skills/check/SKILL.md`, tag `check`): it still
   appears in `--list` / `--all`; and still resolves by a nested path (`writing/check`), by its
   frontmatter `name`, or by a declared alias.
4. (Optional, recommended) one short closing sentence: to point `init` at a store *directory* literally
   named `check`/`init`, pass it with `--store`, not as the positional `init <dir>` (the code's own
   GOTCHA, main.go:281–290; surfaced by §D6). Marked optional — drop if it feels off-topic.

No code change. No Go test. No `PRD.md` edit. No `tasks.json` edit. No new files.

### Success Criteria

- [ ] `README.md` "Where skills live" section contains a `**Reserved tag names.**` callout (bold lead-in
      + period, matching the `**Error contract.**` pattern at README.md:164).
- [ ] The note states `check` and `init` are reserved subcommand names that never resolve as tags, and
      that `skilldozer check` runs validation / `skilldozer init` runs setup.
- [ ] The note frames the rule as the standard CLI subcommand-reservation rule.
- [ ] The note lists all four workarounds: `--list`/`--all`, nested path (`writing/check`), frontmatter
      `name`, alias.
- [ ] The note sits INSIDE the "Where skills live" section (after the `… all resolve.` paragraph at
      README.md:199, before `## Adding a skill` at :201) — NOT in another section.
- [ ] The note uses NO `§`-PRD citations (voice rule: `grep -c '§' README.md` must remain 0).
- [ ] No other file is touched (no Go code, no test, no `.gitignore`, no `PRD.md`, no `tasks.json`).
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` remain green (unchanged — not code).
- [ ] `git status --short` shows only `README.md` modified (plus pre-existing plan/ churn).

## All Needed Context

### Context Completeness Check

**Pass.** The deliverable is one surgical paragraph inserted at a verified-unique seam. The exact
anchor text is quoted byte-for-byte (with `cat -A` line-ends), the section boundaries are pinned by
line number (`grep -n '^## ' README.md`), and the README's callout convention is pinned by an existing
example (`**Error contract.**` at :164). The voice rules are explicit (bold lead-in; plain prose;
inline `` `code` ``; **zero** `§`-citations — `grep`-confirmed). The four required workarounds are
enumerated and each is traced to its tag-resolution step (README.md:193–195 / PRD §7.2) or to a code
comment (nested path, main.go:261). The binding decision (§D6) and the forbidden actions (no code, no
PRD, no test) are stated. An implementer who has never seen this repo can write the note in one pass.

### Documentation & References

```yaml
# MUST READ — the file under edit (the ONLY deliverable)
- file: README.md
  why: "THE edit target. The 'Where skills live' section is README.md:174–200 (confirmed by
        `grep -n '^## ' README.md`). The canonical tag is defined at :179 ('the skill directory's
        path relative to skills/'); the tag-resolution precedence list is :190–196; the closing
        'So skilldozer example ... all resolve.' paragraph is :198–199. INSERT the new note
        immediately AFTER :199 and BEFORE the '## Adding a skill' heading at :201."
  section: "## Where skills live (README.md:174–200). Insert at the END of the section."
  pattern: "README callout = a line starting with a **bold lead-in phrase + period**, then prose.
            Existing template: README.md:164 '**Error contract.** An unknown tag prints ...'.
            Use '**Reserved tag names.**' as the lead-in. Inline `code` for commands/paths/tags."
  gotcha: "VOICE RULE: `grep -c '§' README.md` is 0. The user-facing README NEVER cites PRD section
           numbers. Do NOT write '§8.2' or '§9' in the note — say 'runs validation' / 'runs first-run
           setup' in plain words. (The contract's '(§8.2, §9)' tells the PRP writer where check/init
           are defined; it is NOT a README citation instruction.)"

# MUST READ — the binding decision (documentation-only; surface in README skill-tag section; no PRD edit)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/decisions.md
  why: "§D6 is the decision of record: 'check/init reservation is deliberate and documented in code.
        A code change to resolve a skill named check would silently shadow the skilldozer check
        subcommand — a worse UX surprise. Decision: record the reservation + workarounds ... and
        SURFACE IT IN THE README's SKILL-TAG SECTION. PRD.md is human-owned/READ-ONLY — the architect
        does NOT edit it.' This is WHY the task is doc-only and WHY PRD.md is off-limits."
  section: "D6 (Issue 7)."

# MUST READ — the authoritative bug writeup (the rule + the four workarounds)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/bug_fixes_validation.md
  why: "§ISSUE 7 states the confirmed behavior (check/init reserved BEFORE the default tag-capture
        branch), the deliberate reservation (code comments), and the workarounds (resolve by
        frontmatter name, alias, basename, or nest it e.g. writing/check; still discoverable by
        --list/--all). Also: 'No new code test needed; the decision record is the deliverable' ->
        no Go test to add."
  section: "ISSUE 7 (Minor)."

# MUST READ — the code the note mirrors (already implemented; do NOT change)
- file: main.go
  why: "parseArgs `case \"check\"` (main.go:254) and `case \"init\"` (main.go:269) are matched BEFORE
        the default tag-capture branch (main.go:281+), which is WHY check/init never become tags. The
        comments are explicit: main.go:257 'subcommand names are reserved, as in any CLI'; main.go:261
        'A nested skill writing/check still resolves: this case matches only the EXACT token check';
        main.go:281-290 'GOTCHA: a store literally named check/init must be passed via --store'. The
        README note is the plain-English mirror of these two comments."
  section: "parseArgs, case \"check\" (main.go:254-262) and case \"init\" (main.go:269-290)."
  gotcha: "READ-ONLY here. Do NOT edit main.go. The reservation is already implemented and already
           commented as deliberate. `go build/vet/test` must be unchanged."

# READ-ONLY — the spec the note reconciles with the code (canonical tag = relTag; the note is the carve-out)
- file: PRD.md
  why: "§7.2 (PRD.md:161) defines the canonical tag as the skill dir's relTag with NO carve-out for
        reserved names; §9 defines the `check` subcommand; §8.2 defines `init`. The README note is the
        user-facing statement of the one deliberate carve-out so the spec and behavior agree. The
        PRD 'suggests' a §7.2 note — that is a HUMAN action item (recorded in decisions.md §D6), NOT
        executed here."
  section: "§7.2 (PRD.md:161), §9 (check), §8.2 (init)."
  gotcha: "READ-ONLY. NEVER edit PRD.md (human-owned). Do NOT add a §7.2 note to PRD.md."

# READ-ONLY — the parallel sibling boundary (disjoint paths; land in either order)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/P1M2T4S1/PRP.md
  why: "P1.M2.T4.S1 (Issue 6) edits ONLY .gitignore (its 'Desired Codebase tree' lists exactly one
        file). This subtask edits ONLY README.md. DISJOINT file sets -> no merge conflict; land in
        either order."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/tasks.json
  why: "P1.M2.T5.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes
        it; tasks.json wins on any conflict. OUTPUT: 'README documents the reserved-tag rule and
        workarounds. Consumed by: users reading the README and by the final Mode B consistency sweep
        (P1.M3.T1).' DOCS: '[Mode A] ... the final Mode B sweep (P1.M3.T1) verifies README consistency
        across all 7 fixes but does not duplicate this note.'"
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && grep -n '^## ' README.md   # section boundaries
5:## Why
19:## Install
80:## Shell completions
116:## Usage
174:## Where skills live          ← EDIT TARGET SECTION (body = :174–:200)
201:## Adding a skill             ← next heading (insert the note BEFORE this)
244:## How `skilldozer` finds the store
277:## Constraints

$ sed -n '190,201p' README.md      # the tail of the section (the insertion seam)
Tag resolution tries, in order:

1. the exact canonical tag (`writing/reddit`)
2. the final path segment / basename (`reddit`)
3. the frontmatter `name`
4. a declared alias (see `metadata.aliases`)
5. else: unknown

So `skilldozer example`, `skilldozer writing/reddit`, `skilldozer reddit` (if unique), and
`skilldozer foo-helper` (matching a frontmatter `name`) all resolve.

## Adding a skill

$ grep -c '§' README.md            # voice invariant: must stay 0
0
```

### Desired Codebase tree with files to be changed

```bash
README.md    # EDIT — insert ONE new "**Reserved tag names.**" paragraph at the end of the
             #          "Where skills live" section (after the "...all resolve." paragraph at
             #          README.md:199, before the "## Adding a skill" heading at :201).
             #          This is the ONLY file changed.
# main.go / main_test.go / internal/ / install.sh / .gitignore / go.mod / go.sum — UNCHANGED.
# PRD.md / tasks.json / prd_snapshot.md — READ-ONLY (never touched).
```

| File | Change | Owner |
|---|---|---|
| `README.md` | Insert one `**Reserved tag names.**` callout paragraph at the end of the "Where skills live" section (after `… all resolve.`, before `## Adding a skill`). | Issue 7 contract + decisions.md §D6 |

### Known Gotchas of our codebase & Library Quirks

```bash
# GOTCHA #1 (VOICE — the #1 one-pass slip) — the user-facing README NEVER cites PRD section numbers.
# `grep -c '§' README.md` is 0. The contract's "(§8.2, §9)" tells the PRP writer where check/init are
# defined in the PRD; it is NOT a README-citation instruction. Do NOT write "§8.2", "§9", or "§7.2" in
# the note. Say "runs validation" / "runs first-run setup" in plain words. (verified_facts.md §2.)

# GOTCHA #2 (PLACEMENT) — the note MUST sit INSIDE "Where skills live" (README.md:174–200), at the END
# of that section (after the "...all resolve." paragraph at :199, before the "## Adding a skill"
# heading at :201). Putting it in "Usage" or "Adding a skill" or "Constraints" fails the contract
# ("Add a ... note to the README skill-tag section"). The seam is unique: the only "## Adding a skill"
# heading, and "all resolve." appears once.

# GOTCHA #3 (LEAD-IN FORMAT) — match the house callout style: a line starting with **Bold Phrase.**
# (bold + period), then prose. The template is README.md:164 "**Error contract.** An unknown tag
# prints ...". Use "**Reserved tag names.**" exactly (capital R, capital T, period inside the bold).
# Do NOT use a blockquote (>), a heading (###), or a table — the README uses none of those for callouts.

# GOTCHA #4 (NO CODE / NO TEST) — this is documentation-only (decisions.md §D6). Do NOT edit main.go
# (the reservation is already implemented and commented at main.go:254 / main.go:269). Do NOT add a Go
# test (TestParseArgsCheckSubcommand @main_test.go:1120 already asserts the reservation). `go
# build/vet/test ./...` must be byte-for-byte unchanged in behavior — prove it stays green.

# GOTCHA #5 (NO PRD EDIT) — PRD.md is human-owned / READ-ONLY. The PRD "suggests" a §7.2 note; that is
# a HUMAN action item recorded in decisions.md §D6, NOT executed by this subtask. Do NOT touch PRD.md,
# tasks.json, or prd_snapshot.md.

# GOTCHA #6 (ALL FOUR WORKAROUNDS) — the contract requires the note to list all four: --list/--all,
# nested path (writing/check), frontmatter name, alias. Missing any one fails the contract. Each maps
# to a tag-resolution fallback (README.md:193 name, :194 alias) or a code comment (nested path,
# main.go:261) or discovery-walk independence (--list/--all). The optional 5th (--store for a store
# dir named check/init) is recommended but not required.

# GOTCHA #7 (MARKDOWN HYGIENE) — keep the blank-line discipline: one blank line before the note and
# one after it (so it is a distinct paragraph), and preserve the existing blank line before
# "## Adding a skill". No broken code fences (the note adds none). Verify with `cat` render after edit.

# GOTCHA #8 (DO NOT DUPLICATE) — this note is the canonical home for the reserved-tag rule. The later
# Mode B sweep P1.M3.T1 "verifies README consistency across all 7 fixes but does not duplicate this
# note" (contract DOCS step 5). Do not also add the rule to "Usage" or "Adding a skill".
```

## Implementation Blueprint

### Data models and structure

**None.** This is a prose edit to a Markdown file. No types, fields, signatures, config, code, or tests
are involved. The note text IS the deliverable.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: INSERT the "**Reserved tag names.**" note into README.md (the ONE production action)
  - FILE: README.md (repo root)
  - ACTION: insert ONE new paragraph at the END of the "Where skills live" section.
  - ANCHOR (verified unique; the exact bytes to find):
        `skilldozer foo-helper` (matching a frontmatter `name`) all resolve.

        ## Adding a skill
  - INSERT BETWEEN: the "...all resolve." line and the blank line + "## Adding a skill" heading.
    Result: "...all resolve.\n\n<NOTE>\n\n## Adding a skill\n" (preserve the surrounding blank lines).
  - NOTE TEXT (recommended; matches README voice — bold lead-in, plain prose, inline code, NO §-cites):
        **Reserved tag names.** `check` and `init` are subcommand names, so they never resolve as
        skill tags: `skilldozer check` runs validation and `skilldozer init` runs first-run setup.
        That is the standard CLI rule — a subcommand name takes precedence over a positional
        argument. A skill whose canonical tag collides (`skills/check/SKILL.md`, tag `check`) is
        still fully usable, just not via that one tag: it appears in `--list` and `--all`, and
        resolves by a nested path (`writing/check`), by its frontmatter `name`, or by a declared
        alias. To point `init` at a store directory literally named `check` or `init`, pass it with
        `--store` rather than as a positional argument.
  - REQUIRED CONTENT CHECKS (the note MUST contain all of these):
      * lead-in: "**Reserved tag names.**"
      * both names: "`check`" and "`init`"
      * the rule: never resolve as skill tags; `skilldozer check` runs validation; `skilldozer init`
        runs first-run setup (or "setup")
      * the framing: "standard CLI rule" / "subcommand name takes precedence" (echo the code comment
        "subcommand names are reserved, as in any CLI", main.go:257)
      * all four workarounds: "`--list` and `--all`"; nested path "`writing/check`"; "frontmatter
        `name`"; "alias" (or "declared alias")
  - OPTIONAL (last sentence, the --store companion): recommended because it is the same reserved-name
        rule and the code's own GOTCHA (main.go:281-290); DROP it if it reads as off-topic for the
        skill-tag section. It is NOT one of the four contract-required workarounds (those are about
        reaching a SKILL; this is about a STORE DIR) — keep it clearly distinct.
  - VOICE INVARIANTS (GOTCHA #1/#3): bold lead-in + period; NO "§8.2"/"§9"/"§7.2"; inline `code` for
        commands/paths/tags; plain declarative prose; one idea per sentence.

Task 2: VERIFY the note + isolation invariants (the acceptance loop)
  - PRESENCE + SECTION: awk '/^## Where skills live/{f=1} /^## Adding a skill/{f=0} f&&/Reserved tag names/'
        README.md   # MUST print the lead-in (proves it is INSIDE the section)
  - REQUIRED KEYWORDS (each must match; adjust regex to your final wording):
      grep -q 'Reserved tag names'      README.md   # lead-in present
      grep -q '`check`'                 README.md   # both reserved names...
      grep -q '`init`'                  README.md   # ...named
      grep -qE 'standard CLI|takes precedence|subcommand' README.md   # the framing
      grep -q -- '--list'               README.md   # workaround 1 (discovery)
      grep -q 'writing/check'           README.md   # workaround 2 (nested path)
      grep -q 'frontmatter .name'       README.md   # workaround 3 (name) [regex: adjust quotes]
      grep -qi 'alias'                  README.md   # workaround 4 (alias)
  - VOICE INVARIANT:    grep -c '§' README.md        # MUST be 0 (no PRD citations added)
  - NO CODE REGRESSION: go build ./... && go vet ./... && go test ./...   # all exit 0 (unchanged)
  - ISOLATION: git status --short README.md          # shows " M README.md" only for this path
  - RENDER CHECK: sed -n '174,205p' README.md         # eyeball: note sits after "...all resolve.",
                                                       # blank lines intact, "## Adding a skill" follows
```

### Implementation Patterns & Key Details

```markdown
# The deliverable is one inserted paragraph. The house callout pattern (from README.md:164) is:
#
#   **<Bold lead-in>.** <prose sentence>. <prose sentence>. ...
#
# Applied here:
#
#   **Reserved tag names.** `check` and `init` are subcommand names, so they never resolve as
#   skill tags: `skilldozer check` runs validation and `skilldozer init` runs first-run setup.
#   That is the standard CLI rule — a subcommand name takes precedence over a positional
#   argument. A skill whose canonical tag collides (`skills/check/SKILL.md`, tag `check`) is
#   still fully usable, just not via that one tag: it appears in `--list` and `--all`, and
#   resolves by a nested path (`writing/check`), by its frontmatter `name`, or by a declared
#   alias. To point `init` at a store directory literally named `check` or `init`, pass it with
#   `--store` rather than as a positional argument.
#
# The em dash (—) is already used elsewhere in the README (e.g. README.md "the `--path` label
# is the only way to tell the env var was skipped" region and the Constraints bullets), so it is
# voice-consistent. If you prefer to avoid it, substitute a comma + "and" or a colon.
#
# Placement in context (the seam, before -> after):
#
#   BEFORE:
#     ...`skilldozer foo-helper` (matching a frontmatter `name`) all resolve.
#
#     ## Adding a skill
#
#   AFTER:
#     ...`skilldozer foo-helper` (matching a frontmatter `name`) all resolve.
#
#     **Reserved tag names.** `check` and `init` are subcommand names, ... --store` rather than
#     as a positional argument.
#
#     ## Adding a skill
#
# Keep exactly ONE blank line on each side of the note (Markdown paragraph separation).
```

Notes easy to get wrong:
- Forgetting one of the four workarounds (GOTCHA #6) — the contract requires all four.
- Adding `§8.2`/`§9` citations (GOTCHA #1) — breaks README voice (`grep -c '§'` must stay 0).
- Putting the note in the wrong section (GOTCHA #2) — must be inside "Where skills live".
- Using a blockquote/heading/table instead of the bold-lead-in callout (GOTCHA #3).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Placement: end of "Where skills live", not earlier.** The section defines the canonical tag (:179)
   and the resolution precedence (:190–196); the reserved-name carve-out is the natural closing caveat
   after the precedence list and its "all resolve" summary. Inserting mid-section would interrupt the
   precedence flow; inserting after `## Adding a skill` leaves the tag-defining section without the
   caveat. (verified_facts.md §1.)
2. **Bold-lead-in callout, not a blockquote/heading.** Matches the only existing README callout
   (`**Error contract.**` at :164). A `> ` blockquote or `### ` subheading would introduce a format the
   README never uses for inline notes. (GOTCHA #3, verified_facts.md §2.)
3. **No `§`-PRD citations.** `grep -c '§' README.md` is 0; the user-facing README never cites PRD
   sections. The contract's "(§8.2, §9)" is a pointer for the PRP writer, not a README instruction.
   Plain words ("runs validation", "runs first-run setup") instead. (GOTCHA #1, verified_facts.md §2.)
4. **Include the `--store` companion sentence (optional, recommended).** It is the same reserved-name
   rule, documented in the code's own GOTCHA (main.go:281–290) and surfaced by §D6. It prevents the
   exact confusion a user hitting this would have. It is NOT one of the four contract-required
   workarounds (which are about reaching a SKILL), so it is framed as a distinct closing sentence and
   marked optional — drop it if it feels off-topic. (verified_facts.md §5.)
5. **Documentation-only; no code, no test.** decisions.md §D6 is binding. The reservation is already
   implemented (main.go:254/269) and already asserted by TestParseArgsCheckSubcommand (main_test.go:1120).
   Adding a code change or a test would violate §D6 and the contract. (GOTCHA #4, verified_facts.md §3/§8.)
6. **Echo the code's framing.** The code says "subcommand names are reserved, as in any CLI"
   (main.go:257); the contract says "Frame it as the standard CLI subcommand-reservation rule." The note
   uses "the standard CLI rule — a subcommand name takes precedence over a positional argument" so the
   README and the code comment agree.

### Integration Points

```yaml
DOCUMENTATION (Mode A — the deliverable IS the doc edit):
  - file: README.md
  - section: "## Where skills live" (README.md:174-200)
  - effect: "The tag-defining section now states the check/init reservation + the four workarounds,
             so PRD §7.2 (canonical tag = relTag) and the implemented behavior (check/init reserved as
             subcommands) no longer appear to contradict for the reader."
  - downstream: "Consumed by users reading the README and by the final Mode B consistency sweep
                 (P1.M3.T1), which VERIFIES but does NOT duplicate this note (contract DOCS step 5)."

CODE: NONE.
  - main.go / main_test.go / internal/ UNCHANGED. No Go file is read or written.
  - `go build/vet/test ./...` unaffected (proof of 'no code change').
  - "No new code test needed; the decision record is the deliverable." (bug_fixes_validation.md §ISSUE 7.)

PRD.md / tasks.json / prd_snapshot.md: READ-ONLY (never touched — §D6 + global FORBIDDEN OPERATIONS).

PARALLEL SIBLING (no conflict):
  - P1.M2.T4.S1 edits ONLY .gitignore. This subtask edits ONLY README.md. DISJOINT paths; land in
    either order with no merge conflict.

NO DATABASE / NO ROUTES / NO CONFIG-FORMAT CHANGE / NO PARSEARGS CHANGE / NO NEW FILES.
```

## Validation Loop

### Level 1: Syntax & Style (immediate, after the edit)

```bash
cd /home/dustin/projects/skilldozer

# The note is INSIDE the "Where skills live" section (must print the lead-in):
awk '/^## Where skills live/{f=1} /^## Adding a skill/{f=0} f && /Reserved tag names/' README.md
# Expected: prints the "**Reserved tag names.**" line (proves placement is between :174 and :201).

# Voice invariant — NO PRD section citations were added:
grep -c '§' README.md          # MUST be 0
# Expected: 0

# Render check — eyeball the seam (note after "...all resolve.", blank lines intact, heading follows):
sed -n '196,205p' README.md
# Expected: the "...all resolve." paragraph, a blank line, the new note, a blank line, "## Adding a skill".
```

### Level 2: Required-content checks (the contract's OUTPUT)

```bash
cd /home/dustin/projects/skilldozer

# Lead-in + both reserved names:
grep -q 'Reserved tag names' README.md && echo "lead-in OK"
grep -q '`check`'            README.md && echo "check named"
grep -q '`init`'             README.md && echo "init named"

# The rule + framing:
grep -qE 'never resolve as skill tags|never.*as skill tags' README.md && echo "rule OK"
grep -qE 'standard CLI|takes precedence|subcommand'        README.md && echo "framing OK"

# All four workarounds:
grep -q -- '--list'         README.md && echo "workaround 1 (--list/--all) OK"
grep -q 'writing/check'     README.md && echo "workaround 2 (nested path) OK"
grep -qE 'frontmatter .name' README.md && echo "workaround 3 (name) OK"
grep -qi 'alias'            README.md && echo "workaround 4 (alias) OK"
# Expected: all eight checks print OK. (Adjust the 'name' regex to your exact quoting.)
```

### Level 3: Isolation / no-regression validation

```bash
cd /home/dustin/projects/skilldozer

# No Go regression (nothing changed in Go land, but prove it):
go build ./... ; echo "build exit $?"    # 0
go vet  ./...  ; echo "vet exit $?"      # 0
go test ./...  ; echo "test exit $?"     # 0

# Isolation: only README.md changed by THIS subtask (besides pre-existing plan/ churn):
git status --short README.md             # expect " M README.md"
git diff --name-only                     # README.md should be the only non-plan/ path listed
# Expected: build/vet/test all exit 0; README.md shows as modified; no Go/PRD/tasks file changed.
```

### Level 4: Behavioral / Domain-Specific Validation (manual read-through)

```bash
cd /home/dustin/projects/skilldozer

# Read the whole edited section as a user would, and confirm the four user-facing claims hold:
sed -n '174,205p' README.md
#   (a) A user can state check/init are reserved subcommand names -> YES if the rule sentence is present.
#   (b) A user knows why (standard CLI rule) -> YES if the framing sentence is present.
#   (c) A user can name the four workarounds for skills/check/SKILL.md -> YES if all four are listed.
#   (d) The note matches README voice (bold lead-in, plain prose, no §-cites) -> YES per Level 1.

# Cross-check the note against the code it mirrors (they must agree):
grep -n 'subcommand names are reserved' main.go    # main.go:257 — the code's framing
grep -n 'writing/check still resolves'  main.go    # main.go:261 — the nested-path workaround
# Expected: both code comments present (READ-ONLY — do not edit); the README note echoes them in prose.

# Confirm the reservation is STILL implemented (no accidental code change weakened it):
grep -n 'case "check"' main.go    # main.go:254 — still present
grep -n 'case "init"'  main.go    # main.go:269 — still present
# Expected: both cases still present and unchanged (this subtask touched neither).
```

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — the `awk` placement check prints the lead-in (note is inside "Where skills live"); `grep -c '§' README.md` == 0; `sed` render shows clean markdown at the seam
- [ ] Level 2 PASS — all eight required-content greps print OK (lead-in, both names, rule, framing, four workarounds)
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (no Go regression); `git status --short README.md` shows `M README.md`; no non-plan/ file other than README.md changed
- [ ] Level 4 PASS — manual read-through confirms the four user-facing claims; the note echoes the code comments (main.go:257/261); the reservation cases (main.go:254/269) are unchanged

### Feature Validation
- [ ] The note states `check` and `init` are reserved subcommand names that never resolve as skill tags
- [ ] The note says `skilldozer check` runs validation and `skilldozer init` runs first-run setup
- [ ] The note frames the rule as the standard CLI subcommand-reservation rule
- [ ] The note lists all four workarounds (`--list`/`--all`, nested path `writing/check`, frontmatter `name`, alias)
- [ ] The note sits at the end of the "Where skills live" section (after `… all resolve.`, before `## Adding a skill`)
- [ ] Only `README.md` changed; no code, no test, no `.gitignore`, no `PRD.md`, no `tasks.json`

### Code Quality / Convention Validation
- [ ] Follows the README's existing callout convention (`**Bold lead-in.**` + prose, per `**Error contract.**` at :164)
- [ ] Matches README voice: plain prose, inline `` `code` `` for commands/paths/tags, NO `§`-PRD citations
- [ ] Markdown structure intact (one blank line each side of the note; `## Adding a skill` heading preserved; no broken fences)
- [ ] Does not duplicate the rule elsewhere (P1.M3.T1 verifies but does not re-state it)

### Documentation & Deployment
- [ ] Mode A: this subtask's deliverable IS the doc edit (README skill-tag section) — contract DOCS step 5
- [ ] No `PRD.md` edit (the §7.2 note is a human action item recorded in decisions.md §D6, not executed here)
- [ ] No new environment variables or configuration introduced

---

## Anti-Patterns to Avoid

- ❌ **Don't add `§8.2`/`§9`/`§7.2` citations to the README.** `grep -c '§' README.md` is 0; the
  user-facing README never cites PRD sections. The contract's "(§8.2, §9)" points the PRP writer at the
  PRD, not the README reader. Use plain words ("runs validation", "runs first-run setup"). (GOTCHA #1.)
- ❌ **Don't put the note in the wrong section.** It must be inside "Where skills live" (README.md:174–200),
  at the end (after `… all resolve.`, before `## Adding a skill`). Not in Usage, Adding a skill, or
  Constraints. (GOTCHA #2.)
- ❌ **Don't use a blockquote, heading, or table for the callout.** The README's house style is a
  `**Bold lead-in.**` + prose paragraph (see `**Error contract.**` at :164). (GOTCHA #3.)
- ❌ **Don't edit code or add a test.** This is documentation-only (decisions.md §D6). The reservation is
  already implemented (main.go:254/269) and asserted by TestParseArgsCheckSubcommand (main_test.go:1120).
  `go build/vet/test` must be unchanged. (GOTCHA #4.)
- ❌ **Don't edit `PRD.md`, `tasks.json`, or `prd_snapshot.md`.** READ-ONLY / human-owned / orchestrator-
  owned. The PRD "suggests" a §7.2 note — that is a human action item (§D6), not this subtask's job. (GOTCHA #5.)
- ❌ **Don't omit any of the four workarounds.** The contract requires all four: `--list`/`--all`, nested
  path (`writing/check`), frontmatter `name`, alias. (GOTCHA #6.)
- ❌ **Don't duplicate the rule elsewhere.** This note is canonical; P1.M3.T1 verifies but does not
  re-state it (contract DOCS step 5). (GOTCHA #8.)

---

## Confidence Score

**9.5/10** — This is a single-paragraph surgical insert at a verified-unique seam (the only
`## Adding a skill` heading; `all resolve.` appears once), whose anchor bytes are pinned with `cat -A`.
The README's callout convention is pinned by an existing example (`**Error contract.**` at :164) and the
voice rules are grep-confirmed (zero `§`-citations). All four required workarounds are enumerated and
each is traced to a tag-resolution step (README.md:193–194 / PRD §7.2) or code comment (nested path,
main.go:261). The binding decision (§D6: documentation-only; surface in README skill-tag section; no PRD
edit) and the forbidden actions (no code, no test, no PRD) are explicit and traced to the contract. It
is grep-confirmed that the change is fully isolated to `README.md` (the parallel sibling P1.M2.T4.S1
edits only `.gitignore`), so `go build/vet/test ./...` cannot regress and there is no merge conflict.
The 0.5 reservation is for the two most-likely one-pass slips the PRP cannot fully mechanize away: (a) a
voice slip (an implementer adding a `§8.2` citation or a blockquote despite the GOTCHAs) and (b) a
markdown-hygiene slip (collapsing the blank line before `## Adding a skill`, or breaking paragraph
separation) — both of which the Level 1 `awk`/`grep -c '§'`/`sed` checks catch immediately.
