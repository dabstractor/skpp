# PRP — P1.M3.T1.S1: Sweep README / help text / completion headers for changeset consistency (Mode B)

> **Subtask:** The one **changeset-level documentation sync** (decisions.md §D8) for bugfix round 2.
> It runs **AFTER** every implementing subtask (P1.M1.T1.* / P1.M1.T2.S1 / P1.M2.T1.S1 /
> P1.M2.T2.S1 / P1.M2.T3.S1) has landed. Its charter is to catch **cross-cutting doc drift**
> that the per-file Mode A edits did not carry into the README's overview prose. The contract
> (tasks.json CONTRACT LOGIC §3) lists 4 check categories (a–d); §D8 additionally charters
> catching README drift caused by the issues even when the individual code fixes landed cleanly.
>
> **Scope (S1 ONLY):** documentation-only edits to **`README.md`** in DISJOINT regions. No code,
> no completions, no help text, no install.sh, no tests. Categories (b), (c), (d) are
> **verify-only (no edits)** — the sweep *confirms* their consistency (the Mode A edits already
> landed the correct text). Category (a) Error-contract + §D8 README drift in the "Shell
> completions" and "How skilldozer finds the store" sections are the **actual edits**.
>
> **STATUS (verified at PRP-write time against the live tree):** All 5 issues LANDED. Confirmed:
>   - Issue 1 (`--shell` completion): all 3 completion files route `--shell`→`bash zsh fish` AND
>     advertise `--shell` (D7); their LOCKSTEP headers all carry the new comment line
>     (`--shell's value completes to the bash/zsh/fish enum (§14.2); --shell is advertised (D7)`).
>   - Issue 2 (vanished store): `ErrConfiguredStoreMissing` + the findConfig→Find error path land.
>   - Issue 3 (missing-value symmetry): `--store`/`--search`/`--shell` ALL exit 2 on missing value.
>   - Issue 4 (POSIX `--`): `parseArgs` `endOfOpts` lands; `skilldozer -- <tag>` resolves the tag.
>   - Issue 5 (version doc): README.md:136 already reads the NEW wording (P1.M2.T3.S1 LANDED).
>   - `go build ./...` + `go vet ./...` → clean (exit 0); `go test ./...` → green.

---

## Goal

**Feature Goal**: Perform the final Mode B consistency sweep over the changeset's user-facing
documentation surfaces (README.md prose, `main.go usageText`, completion-file LOCKSTEP headers)
and apply the small README doc fixes the sweep uncovers — specifically the **cross-cutting drift**
the per-issue Mode A edits did not propagate into the README's narrative sections.

**Deliverable**: A set of small, surgical edits to **`README.md`** only:
  - **(a, REQUIRED)** Generalize the Usage "Error contract" paragraph: it only mentions `--store`
    for the missing-value exit-2 contract; after Issue 3, `--search` and `--shell` ALSO exit 2.
  - **(D8, REQUIRED)** Update the README "Shell completions" section's **advertised flag list** to
    include `--shell` (it now advertises 14 flags, not 13) and its **value-completion enumeration**
    to mention `--shell <tab>` → `bash`/`zsh`/`fish`.
  - **(D8, RECOMMENDED)** Clarify the README "How skilldozer finds the store" discovery rule 2:
    "never a hard error" is now subtly incomplete (a *present* config whose `store:` dir vanished is
    a hard error after Issue 2, not a fall-through).
  - **(a, OPTIONAL)** A one-clause POSIX `--` mention in the "no reserved tag names" paragraph —
    **lean OMIT** (dash-leading tags are pathological; the clause clutters the sentence).
  - Categories **(b)** `main.go usageText`, **(c)** completion-file LOCKSTEP headers, and
    **(d)** README Install/version line are **VERIFY-ONLY — no edits** (the Mode A edits already
    landed the correct text; the sweep confirms consistency, it does not duplicate Mode A work).

**Success Definition**: The sweep is performed; the README prose is consistent with the landed
behavior (Issues 1–5); `git diff --name-only` shows ONLY `README.md`; the version line (136) and
all verify-only surfaces are confirmed correct and UNCHANGED by this task; `go build/vet/test ./...`
stays green (proving no code was touched).

---

## User Persona (if applicable)

**Target User**: A reader of the README (a skilldozer user / pi integrator) and any future
maintainer. The sweep's value is that the README's *overview narrative* — not just the code —
stays truthful after a multi-issue changeset.

**Use Case**: A user reads the README to learn the exit-2 contract, the completion behavior, or the
store-discovery rules, and the prose must match what the binary actually does after round 2.

**User Journey**: User reads "Shell completions" → tries `skilldozer --shell <tab>` → sees
`bash zsh fish` → the README (after fix) lists `--shell` among advertised flags and describes the
enum, so observed behavior matches docs. User reads "Error contract" → learns ALL three
value-taking flags exit 2 on a missing value (not just `--store`).

**Pain Points Addressed**: README prose that lags behind landed behavior (a stale 13-flag list, a
`--store`-only exit-2 clause, an over-broad "never a hard error" claim) — exactly the cross-cutting
drift a per-file Mode A edit cannot see.

---

## Why

- **Closes the Mode B documentation task** (decisions.md §D8): "A final 'Sync changeset-level
  documentation' task sweeps README.md and help text for cross-cutting consistency … The README's
  overall feature/coexistence story may need a consistency check after all issues land."
- **Truth-in-docs.** Issue 3 made `--search`/`--shell` symmetrical with `--store` (exit 2 on a
  missing value), but the README's Error-contract paragraph still names only `--store`. A user who
  relies on the README would not know `--search`/`--shell` fail loudly too.
- **Completion discoverability.** Issue 1's Mode A updated the completion *files* (+ their LOCKSTEP
  headers) to advertise `--shell`, but the README's user-facing "Shell completions" section still
  describes the *old* 13-flag list and omits the `--shell` value completion. The README is the
  surface users actually read; the files are the embedded artifact.
- **Discovery-rule accuracy.** Issue 2 turned "present config but vanished store dir" into a HARD
  error (exit 1, named path, no silent fall-through). The README's discovery rule 2 says a bad
  config "never" produces a hard error — true for *missing/unreadable* configs, false for *vanished
  store dir* configs. A clarifying clause removes the ambiguity.
- **Cheapest possible consistency.** Doc-only; no code, no tests, no behavior change. The sweep
  itself (the verify-only categories) IS part of the deliverable — confirming consistency is a
  positive result even where no edit is made.

---

## What

Documentation-only edits to `README.md` in disjoint regions, plus a verify-only sweep of
`main.go usageText`, the three completion-file LOCKSTEP headers, and the README version line.
**No code, no completions, no help text, no install.sh, no tests, no new files.**

### Success Criteria

- [ ] **(a) Error contract (REQUIRED):** the README Error-contract paragraph no longer names *only*
      `--store` for the missing-value exit-2 contract; it now states all three value-taking flags
      (`--store`, `--search`, `--shell`) exit 2 when given as the last token with nothing after them.
- [ ] **(D8) Completions advertised flag list (REQUIRED):** `--shell` appears in the README's
      long-form flag enumeration (alphabetically between `--search` and `--store`), making 14 flags.
- [ ] **(D8) Completions value-completion (REQUIRED):** the README value-completion bullet mentions
      `skilldozer --shell <tab>` offers `bash`/`zsh`/`fish`.
- [ ] **(D8) Store discovery rule 2 (RECOMMENDED):** the rule-2 sentence is clarified so a *present*
      config whose `store:` dir vanished is described as a hard error (exit 1, named path), not a
      silent fall-through.
- [ ] **(a) POSIX `--` (OPTIONAL):** DECISION RECORDED — omit unless it reads cleanly (pathological;
      lean OMIT). If added, one clause in the "no reserved tag names" paragraph. Documented either way.
- [ ] **(b) `main.go usageText` (VERIFY-ONLY):** `--shell` is in OPTIONS and the `--completions`
      line mentions `[--shell <name>]` — confirmed, **no edit**.
- [ ] **(c) Completion LOCKSTEP headers (VERIFY-ONLY):** all three files' headers mention `--shell`
      (advertised + value-completes-to-enum) — confirmed, **no edit**.
- [ ] **(d) README Install/version (VERIFY-ONLY):** README.md:136 has the NEW (P1.M2.T3.S1) wording
      — confirmed, **no edit, no duplicate** (that line is P1.M2.T3.S1's deliverable).
- [ ] `git diff --name-only` → ONLY `README.md` (no code files touched).
- [ ] `go build ./... && go vet ./... && go test ./...` → all green (proving no code was touched).
- [ ] No edits to main.go, install.sh, main_test.go, completions/*, PRD.md, tasks.json,
      prd_snapshot.md, or .gitignore. README.md:136 (version line) is verified, not edited.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every edit anchor is pinned by exact `oldText` (grep-verified unique at PRP-write time)
and read in full surrounding context. The verify-only surfaces are confirmed against the live tree.
The factual basis for each edit is the landed behavior of Issues 1–5 (verified in
`architecture/issue_analysis.md` and confirmed against the current tree: completion files advertise
`--shell`; `go build` clean; the version line is already new). The non-obvious points are all
captured as GOTCHAs: (1) the README's narrative sections use **zero § citations** (the one § cite in
the whole README is in the unrelated "tag names" paragraph at line 181) — so all NEW prose must be
**citation-free plain words**; (2) the version line (136) is P1.M2.T3.S1's territory — verify, do
not re-edit; (3) the POSIX `--` clause is optional and pathological — lean OMIT; (4) `--shell` is
inserted **alphabetically** between `--search` and `--store` in the flag list. An implementer who has
never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the authoritative issue analysis (root cause + fix surface per issue)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/issue_analysis.md
  why: "The 5 issues. §Issue 1 (--shell completion, all 3 files), §Issue 2 (vanished store),
        §Issue 3 (missing-value exit-2 symmetry: --store/--search/--shell), §Issue 4 (POSIX --),
        §Issue 5 (version doc). Defines the LANDED behavior this sweep must reconcile the README to."
  critical: "§Issue 3 is the source of the Error-contract drift (README names only --store). §Issue 1
             is the source of the completions-list drift. §Issue 2 is the source of the
             'never a hard error' drift. §Issue 4 is the OPTIONAL -- clause (pathological)."
  section: "Issues 1-5 (the whole file)."

# MUST READ — the decisions (esp. D7 + D8, which charter this exact task)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/architecture/decisions.md
  why: "D7 (--shell advertised in completion files — drives the README completions-list edit). D8
        (Mode A per-subtask + Mode B final sweep — THIS task's charter: 'sweeps README.md and help
        text for cross-cutting consistency … confirms consistency, does not duplicate Mode A edits')."
  section: "D7, D8."

# MUST READ — THE edit target (read every region before editing; anchors are grep-unique)
- file: README.md
  why: "THE file under edit. 4 regions: (1) Error-contract paragraph (~line 147-150, names only
        --store); (2) 'Shell completions' advertised flag list (~line 293-296, 13 flags, no --shell);
        (3) 'Shell completions' value-completion bullet (~line 298, no --shell); (4) 'How skilldozer
        finds the store' rule 2 (~line 244-246, 'never a hard error'). All 4 oldText strings are
        grep-UNIQUE (one match each) so a single edit call per region is unambiguous."
  pattern: "README narrative voice: inline `code`, em dashes (—), plain declarative prose, bold
            lead-ins. The narrative sections (Error contract, Completions, Store) use ZERO §
            citations — match that (see GOTCHA #1)."
  gotcha: "Line numbers DRIFT (verified_facts cited 254-256 for the tag-names para; it is actually
           181-182). Anchor EVERY edit by oldText, NEVER by line number. The README is a moving
           target across the changeset; only oldText is stable."

# MUST READ — the contract (the orchestrator owns it; LOGIC §3 a-d is authoritative)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/tasks.json
  why: "P1.M3.T1.S1's CONTRACT (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. LOGIC §3 categories
        a-d define the sweep. OUTPUT §4: 'Any small doc fixes discovered during the sweep. If no
        inconsistencies are found, document that fact (the sweep is the deliverable).' DOCS §5:
        Mode B — THIS IS the changeset-level doc task."
  section: "P1.M3.T1.S1 CONTRACT."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint README regions, no collision)
- file: plan/004_5851dcff4371/bugfix/001_ba6f35f74a59/P1M2T3.S1/PRP.md
  why: "Confirms P1.M2.T3.S1 (Issue 5) edits README.md line 136 (the version comment) and it has
        ALREADY LANDED. This task edits DISJOINT regions (~147-150, ~244-246, ~293-298). ZERO overlap
        with line 136. Pin every seam by oldText. Do NOT touch line 136 (GOTCHA #2)."

# READ-ONLY — verify-only surfaces (categories b, c) — confirm, do NOT edit
- file: main.go
  why: "Category (b) verify: usageText line ~111 `--shell <bash|zsh|fish>  Force a shell for
        completion` is in OPTIONS; line ~110 `--completions [--shell <name>]  Emit the shell
        completion script for eval` mentions --shell. CONFIRMED PRESENT → no edit."
  section: "usageText (grep `--shell` in main.go; ~lines 110-111)."
- file: completions/skilldozer.bash
  why: "Category (c) verify: header (lines ~18-19) says '--shell's value completes to the
        bash/zsh/fish enum; --shell is advertised (D7)'. Same header text in all 3 files. CONFIRMED."
  section: "header comment block (lines 1-20)."
- file: completions/_skilldozer
  why: "Category (c) verify: identical header line. CONFIRMED."
  section: "header comment block (lines 1-20)."
- file: completions/skilldozer.fish
  why: "Category (c) verify: identical header line. CONFIRMED."
  section: "header comment block (lines 1-20)."

# READ-ONLY — PRD grounding for the landed behaviors the README must now reflect
- file: PRD.md
  why: "READ-ONLY. §6.1 (--shell in the flags table), §6.4 (error semantics), §14.x (completions).
        Grounds the exit-2 contract, the --shell enum, and the vanished-store error. Do NOT edit."
  section: "§6.1, §6.4, §14.2/§14.6."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer
$ ls README.md main.go completions/
README.md  main.go  completions/:
_skilldozer  skilldozer.bash  skilldozer.fish
# Verify-only surfaces (categories b, c) — grep confirms --shell already present:
$ grep -n '\-\-shell' main.go | head -3        # usageText OPTIONS (~111) + --completions line (~110)
$ grep -n 'advertised (D7)' completions/*       # all 3 headers carry the --shell comment line
# The 4 README edit regions — each oldText is grep-UNIQUE (one match):
$ grep -c '\-\-store` expects a value' README.md          # 1  (Error contract, ~147)
$ grep -c 'deliberately not advertised' README.md         # 1  (completions flag list, ~296)
$ grep -c 'offers nothing (free-text)' README.md          # 1  (value-completion, ~298)
$ grep -c 'never a hard error' README.md                  # 1  (store rule 2, ~246)
```

### Desired Codebase tree with files to be changed

```bash
README.md      # EDIT 3-4 disjoint regions (Error contract; Completions flag list + value bullet; Store rule 2).
# main.go / install.sh / main_test.go / completions/* / internal/* — UNCHANGED (doc-only sweep).
```

| Region (anchor by oldText) | Category | Change | Priority |
|---|---|---|---|
| Error-contract paragraph (`--store` expects a value…) | (a) Issue 3 | Generalize to all 3 value-taking flags (`--store`/`--search`/`--shell`) exit 2 on missing value. | REQUIRED |
| Completions: advertised flag list (`…--search`, `--store`…) | D8 Issue 1 | Insert `--shell` alphabetically (→14 flags). | REQUIRED |
| Completions: value-completion bullet (`…--search <tab> offers nothing…`) | D8 Issue 1 | Add `--shell <tab>` → `bash`/`zsh`/`fish`. | REQUIRED |
| Store discovery rule 2 (`…never a hard error.`) | D8 Issue 2 | Clarify: a *present* config whose `store:` dir vanished is a hard error (exit 1, named path). | RECOMMENDED |
| "no reserved tag names" paragraph | (a) Issue 4 | OPTIONAL POSIX `--` clause — **lean OMIT** (pathological). Record decision. | OPTIONAL |
| `main.go usageText` (lines ~110-111) | (b) | VERIFY-ONLY — `--shell` present. No edit. | VERIFY |
| Completion LOCKSTEP headers (all 3 files) | (c) | VERIFY-ONLY — `--shell` documented. No edit. | VERIFY |
| README.md:136 (version line) | (d) | VERIFY-ONLY — NEW wording present (P1.M2.T3.S1). No edit, no duplicate. | VERIFY |

### Known Gotchas of our codebase & Library Quirks

```bash
# GOTCHA #1 — README narrative sections use ZERO § citations. The ONLY § cite in the entire README
# is §6.1 inside the "no reserved tag names" paragraph (grep -no '§[0-9.]*' README.md → exactly ONE
# match, line 181). The sections THIS task edits (Error contract, Completions, Store) use plain
# words and inline `code` — NO § cites. So every NEW sentence MUST be citation-free plain prose
# (e.g. "exits 2", NOT "exits 2 (§6.4)"). Matching local voice = reviewable, consistent prose.

# GOTCHA #2 — README.md:136 (the version line) is P1.M2.T3.S1's deliverable and has ALREADY LANDED
# the new wording ("…when built via ./install.sh; a plain 'go build' reports 'dev'"). This sweep
# VERIFIES it present; it does NOT re-edit it (D8: the sweep "confirms consistency," it does not
# duplicate Mode A edits). Do NOT touch line 136. All this task's edits are DISJOINT from it.

# GOTCHA #3 — anchor EVERY edit by oldText, NEVER by line number. Line numbers drift across the
# changeset (verified_facts cited 254-256 for the tag-names para; it is now 181-182). Only the
# exact oldText string is stable. Each of the 4 oldText strings below is grep-UNIQUE (1 match).

# GOTCHA #4 — in the completions flag list, insert `--shell` ALPHABETICALLY between `--search` and
# `--store` (the list is alphabetical). Do NOT append it at the end. The surrounding flags are:
# `--search`, `--shell`, `--store` (then `--version`). Keep the em-dash separators intact.

# GOTCHA #5 — the POSIX `--` clause (category a, Issue 4) is OPTIONAL and PATHOLOGICAL. The contract
# says so explicitly ("optional since such tags are pathological"). A dash-leading tag like `-foo`
# that you must address with `skilldozer -- -foo` is a vanishingly rare case. Lean OMIT. If you add
# it, ONE clause at the end of the "no reserved tag names" paragraph; if it clutters the sentence,
# OMIT. Either way, RECORD the decision (the sweep deliverable includes the decision, not just edits).

# GOTCHA #6 — README.md ONLY. No code, no completions, no help text, no install.sh, no tests
# (contract DOCS §5 = Mode B doc-only). `git diff --name-only` MUST list ONLY README.md.
# main.go usageText (category b) and completion headers (category c) are CONFIRMED CONSISTENT and
# must NOT be edited — editing them would duplicate Issue 1's Mode A work and risk embed/on-disk
# drift (main_test.go:TestEmbeddedCompletionsMatchOnDisk enforces byte-identity).

# GOTCHA #7 — em dashes in this README are the UTF-8 em dash (—, U+2014), NOT "--". When editing,
# preserve the existing em dash style (grep shows `—` throughout). Do not introduce ASCII "--" in
# prose. (This is a doc-style consistency point, not a correctness one.)

# GOTCHA #8 — keep edits minimal (Mode B). Match each surrounding sentence's structure and length.
# The Error-contract edit GENERALIZES one sentence (do not add a whole new paragraph). The
# completions edits are ONE flag insertion + ONE clause. The store edit is ONE clarifying sentence
# appended to the existing rule-2 sentence. Surgical, not expansive.
```

---

## Implementation Blueprint

### Data models and structure

**None.** Documentation-only. No code, no types, no config.

### Implementation Tasks (ordered by dependencies)

```yaml
# ===== PART 1: VERIFY-ONLY sweep (categories b, c, d) — confirm, do NOT edit =====

Task 1: VERIFY category (b) — main.go usageText already advertises --shell
  - RUN: grep -n '\-\-shell' main.go
  - EXPECT: usageText line ~111 `--shell <bash|zsh|fish>      Force a shell for completion`
            AND line ~110 `--completions [--shell <name>]   Emit the shell completion script for eval`.
  - ACTION: none. Confirmed present → consistent → NO edit. (If somehow absent, STOP and flag —
    that would mean a prior subtask introduced drift. It is NOT absent — verified at PRP-write time.)

Task 2: VERIFY category (c) — all 3 completion-file LOCKSTEP headers mention --shell
  - RUN: grep -l 'advertised (D7)' completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish
  - EXPECT: all 3 files listed. Each header carries:
            "--shell's value completes to the bash/zsh/fish enum (§14.2); --shell is
             advertised (D7) since it is a real, documented flag in usageText OPTIONS."
  - ACTION: none. Confirmed consistent → NO edit. (Completion files are EMBEDDED in main.go and
    guarded byte-for-byte by TestEmbeddedCompletionsMatchOnDisk — do NOT touch them here.)

Task 3: VERIFY category (d) — README.md:136 version line has the P1.M2.T3.S1 wording
  - RUN: grep -c "git-describe value when built via ./install.sh" README.md
  - EXPECT: 1 (line 136, NEW wording LANDED). AND `grep -c 'dynamic, not a fixed string' README.md` → 0 (OLD gone).
  - ACTION: none. Confirmed present → consistent → NO edit, NO duplicate. (Line 136 is
    P1.M2.T3.S1's deliverable; this task's edits are DISJOINT from it. GOTCHA #2.)

# ===== PART 2: the actual edits (README.md only) =====

Task 4: EDIT README.md — (a) Error contract: generalize --store → all 3 value-taking flags  [REQUIRED]
  - ANCHOR (grep-unique oldText; the tail of the Error-contract paragraph, ~line 147-150):
        the whole store). `--store` expects a value: `--init --store` with nothing after
        it exits 2 rather than guessing a store.
  - REPLACE WITH (citation-free plain prose, GOTCHA #1; generalize to --store/--search/--shell):
        the whole store). The flags that take a value — `--store`, `--search`, and
        `--shell` — all exit 2 when given as the last token with nothing after them,
        rather than guessing a value.
  - WHY: Issue 3 made --search/--shell symmetrical with --store (all exit 2 on missing value,
    verified against the built binary). The old text named only --store.

Task 5: EDIT README.md — (D8) Completions advertised flag list: add --shell (alphabetical)  [REQUIRED]
  - ANCHOR (grep-unique oldText; the long-form flag enumeration, ~line 293-296):
        - `skilldozer -<tab>` lists the **long-form flags only** — `--all`, `--check`,
          `--completions`, `--file`, `--help`, `--init`, `--list`, `--no-color`,
          `--path`, `--relative`, `--search`, `--store`, `--version` — narrowed by what
          you type after the dash. Short aliases (`-a`, `-l`, …) stay valid for typing
          but are deliberately not advertised.
  - REPLACE WITH (insert --shell between --search and --store, GOTCHA #4; now 14 flags):
        - `skilldozer -<tab>` lists the **long-form flags only** — `--all`, `--check`,
          `--completions`, `--file`, `--help`, `--init`, `--list`, `--no-color`,
          `--path`, `--relative`, `--search`, `--shell`, `--store`, `--version` — narrowed
          by what you type after the dash. Short aliases (`-a`, `-l`, …) stay valid for
          typing but are deliberately not advertised.
  - WHY: Issue 1 (D7) made the completion files advertise --shell; the README's user-facing
    list still showed the old 13. The README is the surface users read.

Task 6: EDIT README.md — (D8) Completions value-completion: add --shell <tab> → enum  [REQUIRED]
  - ANCHOR (grep-unique oldText; the value-completion bullet, ~line 298):
        - `skilldozer --init <tab>` and `skilldozer --store <tab>` offer directories
          (the store to adopt); `skilldozer --search <tab>` offers nothing (free-text).
  - REPLACE WITH (append the --shell enum clause):
        - `skilldozer --init <tab>` and `skilldozer --store <tab>` offer directories
          (the store to adopt); `skilldozer --search <tab>` offers nothing (free-text);
          `skilldozer --shell <tab>` offers the three supported shells — `bash`, `zsh`,
          and `fish`.
  - WHY: same as Task 5 — the value-completion enumeration omitted --shell.

Task 7: EDIT README.md — (D8) Store discovery rule 2: clarify vanished-store is now a hard error  [RECOMMENDED]
  - ANCHOR (grep-unique oldText; discovery rule 2, ~line 244-246):
           A missing or unreadable config is treated as "not yet configured" and falls
           through to the rules below — never a hard error.
  - REPLACE WITH (append ONE clarifying sentence — GOTCHA #8 surgical; citation-free):
           A missing or unreadable config is treated as "not yet configured" and falls
           through to the rules below — never a hard error. A config whose `store:` points
           at a directory that no longer exists is different: skilldozer names the missing
           path and exits 1 rather than silently falling through to a different store.
  - WHY: Issue 2 turned present-config-with-vanished-store-dir into ErrConfiguredStoreMissing
    (exit 1, named path). The old "never a hard error" was true only for missing/unreadable
    configs, not for vanished-store-dir configs.

Task 8: DECIDE the (a) POSIX `--` clause — record OMIT (or a minimal clause)  [OPTIONAL]
  - ANCHOR (context only — the "no reserved tag names" paragraph, ~line 181-183):
        There are **no reserved tag names**: bare words are always skill tags, and every
        action is a `--flag` (§6.1). A skill named `check`, `init`, or `completions`
        resolves normally by its tag — use `--check`, `--init`, or `--completions` to
        run the action.
  - DECISION: LEAN OMIT (GOTCHA #5 — pathological; a dash-leading tag addressable only via
    `skilldozer -- -foo` is vanishingly rare, and a clause would clutter the sentence). If the
    implementer judges a clause reads cleanly, append: "A dash-leading tag can be addressed with
    the POSIX `--` separator: `skilldozer -- -foo`." — otherwise OMIT. Either way the DECISION
    is part of the sweep deliverable (contract OUTPUT §4: "If no inconsistencies are found,
    document that fact").
  - ACTION: OMIT (recommended) OR append the one clause. Document the choice in the implementation
    summary. Do NOT force the clause if it reads awkwardly.

Task 9: VERIFY — the sweep is complete, scoped to README.md, and breaks nothing
  - The NEW prose is present (Task 4/5/6/7 grep checks below); the version line (136) is UNCHANGED.
  - git diff --name-only                  # expect ONLY README.md (GOTCHA #6)
  - git diff --quiet main.go install.sh main_test.go completions/* internal/* && echo "code unchanged"
  - go build ./... && go vet ./... && go test ./...   # all green (proves no code touched)
  - (sanity) read README.md ~145-152, ~242-247, ~291-300 to confirm the edits read cleanly and
    the surrounding prose / bullet structure is intact (GOTCHA #8 surgical, no fence imbalance).
```

### Implementation Patterns & Key Details

```markdown
# --- Task 4: Error contract (REQUIRED) ---
# before (tail of the Error-contract paragraph):
the whole store). `--store` expects a value: `--init --store` with nothing after
it exits 2 rather than guessing a store.

# after (generalized to all three value-taking flags; citation-free):
the whole store). The flags that take a value — `--store`, `--search`, and
`--shell` — all exit 2 when given as the last token with nothing after them,
rather than guessing a value.

# --- Task 5: Completions advertised flag list (REQUIRED) ---
# before (13 flags, no --shell):
  `--path`, `--relative`, `--search`, `--store`, `--version` — narrowed by what
# after (--shell inserted alphabetically between --search and --store; 14 flags):
  `--path`, `--relative`, `--search`, `--shell`, `--store`, `--version` — narrowed

# --- Task 6: Completions value-completion (REQUIRED) ---
# before (no --shell):
  (the store to adopt); `skilldozer --search <tab>` offers nothing (free-text).
# after (--shell enum appended):
  (the store to adopt); `skilldozer --search <tab>` offers nothing (free-text);
  `skilldozer --shell <tab>` offers the three supported shells — `bash`, `zsh`,
  and `fish`.

# --- Task 7: Store discovery rule 2 (RECOMMENDED) ---
# before (over-broad "never a hard error"):
   A missing or unreadable config is treated as "not yet configured" and falls
   through to the rules below — never a hard error.
# after (clarify vanished-store is a hard error after Issue 2):
   A missing or unreadable config is treated as "not yet configured" and falls
   through to the rules below — never a hard error. A config whose `store:` points
   at a directory that no longer exists is different: skilldozer names the missing
   path and exits 1 rather than silently falling through to a different store.
```

Notes easy to get wrong:
- All NEW prose is **citation-free** (GOTCHA #1) — no `§6.4`, `§14.2` etc. in the edited sections.
- `--shell` is inserted **alphabetically** between `--search` and `--store` (GOTCHA #4), never appended.
- Line numbers are **ignored** — anchor by oldText only (GOTCHA #3); each oldText is grep-unique.
- Do **not** touch the version line (136), `main.go`, or any completion file (GOTCHA #2, #6).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Categories (b), (c), (d) are verify-only, not edits.** The contract LOGIC §3 itself says "(b) …
   No change expected unless prior subtasks introduced drift" and "(d) … verify the version
   documentation … is consistent." All three are confirmed consistent at PRP-write time. Editing
   them would (a) duplicate the Mode A work (D8 forbids this) and (b) for completions, risk breaking
   the embed↔on-disk byte-identity test. Verify-and-record is the correct, lower-risk action.
2. **The Error-contract edit GENERALIZES rather than enumerates per-flag prose.** Instead of three
   separate sentences ("--store exits 2", "--search exits 2", "--shell exits 2"), one generalized
   sentence names all three flags and the shared rule. This matches the README's compact voice and
   survives future value-flag additions better than per-flag sentences.
3. **`--shell` inserted alphabetically.** The advertised list is alphabetical; appending at the end
   would be inconsistent with the list's own ordering and with the completion files (which list it
   in-flag-set order). Alphabetical between `--search` and `--store` is the only consistent spot.
4. **Store rule 2 is a clarifying APPEND, not a rewrite.** The original "never a hard error" is still
   true for missing/unreadable configs. The fix narrows the over-broad claim by adding the
   vanished-store exception, preserving the original sentence and its nuance (GOTCHA #8 surgical).
5. **POSIX `--` clause: OMIT (recommended).** Pathological, clutters the "no reserved tag names"
   sentence, and the contract calls it optional. The sweep's *decision* (omit) is the deliverable
   for this check, per OUTPUT §4. A clean one-clause add is permitted if it reads well, but OMIT is
   the lean default.
6. **The sweep is the deliverable even where no edit is made.** OUTPUT §4: "If no inconsistencies
   are found, document that fact." Categories (b)/(c)/(d) and the optional POSIX clause produce a
   *documented verification result*, which is itself a positive deliverable — not a no-op.

### Integration Points

```yaml
FILES TOUCHED:
  - README.md ONLY (4 disjoint regions: Error contract ~147-150; Completions ~293-298;
    Store ~244-246; optionally tag-names ~181-183). No code, no completions, no help text.

VERIFY-ONLY (confirmed consistent, NO edits — DO NOT touch):
  - main.go usageText (~lines 110-111): --shell in OPTIONS; --completions mentions [--shell <name>].
  - completions/skilldozer.bash, completions/_skilldozer, completions/skilldozer.fish: LOCKSTEP
    headers all carry the --shell comment line (advertised + value-enum). EMBEDDED + byte-guarded
    by TestEmbeddedCompletionsMatchOnDisk — editing would risk that test.
  - README.md:136 (version line): P1.M2.T3.S1's deliverable, already LANDED new wording.

DOCUMENTATION (this IS the deliverable — Mode B, contract DOCS §5):
  - The changeset-level doc sync. Confirms consistency for (b)/(c)/(d); fixes cross-cutting drift
    for (a) Error contract + (D8) Completions/Store.

PARALLEL SIBLINGS (no conflict — all LANDED, this task runs last):
  - P1.M1.T1.* (completions), P1.M1.T2.S1 (skillsdir), P1.M2.T1.S1 (main.go parseArgs/run),
    P1.M2.T2.S1 (main.go parseArgs --), P1.M2.T3.S1 (README line 136) — all COMPLETE. This task's
    README regions are DISJOINT from line 136. Zero overlap; runs safely as the final task.

NO ROUTES / NO DATABASE / NO CONFIG CODE / NO NEW FILES / NO TESTS.
```

---

## Validation Loop

### Level 1: Verify-only sweep (categories b, c, d) — expect CONSISTENT, no edits

```bash
cd /home/dustin/projects/skilldozer

# (b) main.go usageText advertises --shell (in OPTIONS + on the --completions line):
grep -n '\-\-shell' main.go          # expect ~line 110 (--completions [--shell <name>]) AND ~111 (--shell <bash|zsh|fish>)

# (c) all 3 completion-file LOCKSTEP headers mention --shell (advertised + value-enum):
grep -l 'advertised (D7)' completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish
# expect all 3 files listed.

# (d) README.md:136 version line has the P1.M2.T3.S1 (Issue 5) wording, OLD gone:
grep -c 'git-describe value when built via ./install.sh' README.md   # expect 1 (NEW present, line 136)
grep -c 'dynamic, not a fixed string' README.md                       # expect 0 (OLD gone)
# Expected: all three verify checks confirm consistency. NO edits made to these surfaces.
```

### Level 2: The actual README edits (NEW prose present, OLD prose gone)

```bash
cd /home/dustin/projects/skilldozer

# Task 4 — Error contract generalized (OLD "--store expects a value" clause gone; NEW names all 3 flags):
grep -c 'The flags that take a value' README.md          # expect 1 (NEW)
grep -c '\-\-store` expects a value' README.md           # expect 0 (OLD gone)

# Task 5 — --shell inserted alphabetically in the advertised flag list:
grep -c '`--search`, `--shell`, `--store`' README.md     # expect 1 (NEW — alphabetical triplet)

# Task 6 — --shell value-completion bullet added:
grep -c '\-\-shell <tab>` offers the three supported shells' README.md   # expect 1 (NEW)

# Task 7 — Store rule 2 clarified (vanished-store is a hard error):
grep -c 'rather than silently falling through to a different store' README.md   # expect 1 (NEW)

# Task 8 — POSIX `--` clause: record the OMIT decision (if OMITTED, this grep may be 0; that's fine):
grep -c 'POSIX \`--\` separator' README.md                # 0 if OMITTED (recommended), 1 if the clause was added
# Expected: Tasks 4-7 NEW prose present exactly once each; Task 8 OMIT (0) or clause (1) — decision recorded.
```

### Level 3: Scope discipline (NO code / completions / version-line touched)

```bash
cd /home/dustin/projects/skilldozer

git diff --name-only                                              # expect ONLY: README.md
git diff --quiet main.go && echo "main.go unchanged"              # GOTCHA #6
git diff --quiet install.sh && echo "install.sh unchanged"
git diff --quiet main_test.go && echo "main_test.go unchanged"
git diff --quiet go.mod go.sum && echo "go.mod/sum unchanged"
git diff --quiet completions/skilldozer.bash completions/_skilldozer completions/skilldozer.fish && echo "completions unchanged"
# Expected: README.md is the ONLY changed file. (The version line 136 is INSIDE README.md but must
# show NO diff — verify: git diff README.md should NOT touch line 136.)
git diff README.md | grep -n '136' || echo "version line (136) untouched ✓"
# Expected: all code/completion files unchanged; README.md:136 (version line) shows no diff.
```

### Level 4: Build/test still green + README reads cleanly (proves doc-only, no breakage)

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"      # 0
go vet  ./...  ; echo "vet exit $?"        # 0
go test ./...  ; echo "test exit $?"       # 0  — a README edit cannot affect this; green = scope held
# Expected: all exit 0. (Scope-discipline guard: green proves no code/completions were touched.)

# Readability spot-check — the 4 edited regions read cleanly with intact surrounding structure:
sed -n '145,152p' README.md     # Error-contract paragraph (generalized sentence flows)
sed -n '291,300p' README.md     # Completions section (14-flag list + --shell enum bullet)
sed -n '242,248p' README.md     # Store discovery rule 2 (clarifying sentence appended)
# Expected: each region's prose flows; bullet/fence structure intact; no § citations introduced.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — categories (b) `main.go usageText`, (c) all 3 completion headers, (d) README:136
      version line all CONFIRMED CONSISTENT with no edits made (the verify-only sweep result)
- [ ] Level 2 PASS — Tasks 4–7 NEW prose present exactly once each; OLD prose gone; Task 8 OMIT/decision recorded
- [ ] Level 3 PASS — `git diff --name-only` = ONLY `README.md`; main.go/install.sh/main_test.go/completions/* unchanged; README.md:136 shows no diff
- [ ] Level 4 PASS — `go build/vet/test ./...` all exit 0; the 4 edited README regions read cleanly

### Feature Validation
- [ ] (a) Error-contract paragraph names all three value-taking flags (`--store`/`--search`/`--shell`) for the exit-2 missing-value contract (Issue 3)
- [ ] (D8) Completions advertised flag list includes `--shell` alphabetically (14 flags, Issue 1)
- [ ] (D8) Completions value-completion bullet describes `--shell <tab>` → `bash`/`zsh`/`fish` (Issue 1)
- [ ] (D8) Store discovery rule 2 clarifies a present config with a vanished `store:` dir is a hard error (exit 1, named path), not a fall-through (Issue 2)
- [ ] (a) POSIX `--` clause DECISION recorded (recommended OMIT, or a clean clause added) (Issue 4)
- [ ] (b) `main.go usageText` advertises `--shell` (verified, not edited)
- [ ] (c) all 3 completion LOCKSTEP headers mention `--shell` (verified, not edited)
- [ ] (d) README.md:136 version line has the P1.M2.T3.S1 wording (verified, not duplicated)

### Code Quality / Convention Validation
- [ ] All NEW prose is citation-free plain words (GOTCHA #1) — no `§6.4`/`§14.x` in edited sections
- [ ] `--shell` inserted alphabetically between `--search` and `--store` (GOTCHA #4)
- [ ] Edits anchored by oldText, not line numbers (GOTCHA #3); each oldText was grep-unique
- [ ] Em-dash style (—) preserved; no ASCII `--` introduced in prose (GOTCHA #7)
- [ ] Edits are surgical (GOTCHA #8) — generalize/append, not expand into new paragraphs

### Scope Discipline
- [ ] Edited README.md ONLY (contract DOCS §5 = Mode B doc-only; OUTPUT §4)
- [ ] Did NOT touch main.go, install.sh, main_test.go, completions/*, internal/*
- [ ] Did NOT touch README.md:136 (the version line — P1.M2.T3.S1's deliverable; GOTCHA #2)
- [ ] Did NOT modify PRD.md (read-only), tasks.json, prd_snapshot.md, or .gitignore
- [ ] Did NOT edit the verify-only surfaces (b)/(c)/(d) — confirmed consistent, not duplicated

---

## Anti-Patterns to Avoid

- ❌ **Don't edit the verify-only surfaces.** Categories (b) `main.go usageText`, (c) completion
  headers, and (d) the version line are confirmed consistent — the Mode A edits already landed them.
  Editing duplicates work (D8 forbids) and, for completions, risks the embed↔on-disk byte-identity
  test (`TestEmbeddedCompletionsMatchOnDisk`). Verify-and-record is correct. (GOTCHA #6.)
- ❌ **Don't introduce § citations in the edited prose.** The README's narrative sections use plain
  words; the only § cite in the whole file is in the unrelated tag-names paragraph. New prose stays
  citation-free to match local voice. (GOTCHA #1.)
- ❌ **Don't append `--shell` at the end of the flag list.** The list is alphabetical — insert it
  between `--search` and `--store`. Appending breaks the list's own ordering. (GOTCHA #4.)
- ❌ **Don't touch README.md:136.** That's P1.M2.T3.S1's deliverable, already LANDED. This task's
  regions are disjoint from it; verify it present, do not re-edit. (GOTCHA #2.)
- ❌ **Don't anchor edits by line number.** Line numbers drift across the changeset; only oldText is
  stable. Each oldText is grep-unique so one `edit` call matches exactly once. (GOTCHA #3.)
- ❌ **Don't force the POSIX `--` clause.** It's optional and pathological; lean OMIT. The sweep's
  *decision* (omit) is a valid deliverable per OUTPUT §4. Only add the clause if it reads cleanly.
  (GOTCHA #5.)
- ❌ **Don't expand edits into new paragraphs.** Mode B is surgical: generalize one sentence (Error
  contract), insert one flag + one clause (Completions), append one clarifying sentence (Store).
  Match surrounding sentence structure. (GOTCHA #8.)
- ❌ **Don't touch any code file.** Documentation-only (contract DOCS §5). `git diff --name-only`
  must be ONLY `README.md`. (GOTCHA #6.)

---

## Confidence Score

**9.5/10** — This is a documentation-only Mode B sweep of `README.md` in 4 disjoint, grep-unique
regions. Every edit anchor's exact oldText is verified against the live tree at PRP-write time (each
is grep-unique → 1 match, so a single `edit` call per region is unambiguous), and the verify-only
surfaces (categories b/c/d) are confirmed consistent. The factual basis (Issues 1–5 landed behavior)
is verified in `architecture/issue_analysis.md` and re-confirmed against the current tree: the
completion files advertise `--shell`; `--store`/`--search`/`--shell` all exit 2 on a missing value;
the vanished-store is now a hard error; the POSIX `--` lands; the version line is already new;
`go build/vet` is clean. The README voice rules (citation-free narrative prose, alphabetical flag
list, em-dash style, surgical edits) are pinned by GOTCHAs and the §-cite-count grep. The change is
doc-only — no code, no tests, no completions, no behavior — so the only failure modes are stylistic
(adding a § cite, mis-ordering `--shell`, touching line 136, or expanding an edit), each explicitly
guarded. The 0.5 reservation is for (a) the subjective readability of the Task 7 clarifying sentence
and (b) the Task 8 POSIX-clause judgment call (which the PRP resolves as OMIT but leaves to the
implementer's read). Zero file-level overlap with the parallel siblings (all LANDED; this runs last;
its README regions are disjoint from line 136). The Level 1–4 grep + git-diff + build checks catch
every drift class immediately.
