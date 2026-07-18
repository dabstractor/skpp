name: "P1.M3.T1.S1 — Mode B sweep: README overview coherence + verify §13 acceptance for --link batch"
description: |

---

## Goal

**Feature Goal**: Make `README.md`'s cross-cutting/overview sections fully coherent with the `--link` multi-target batch behavior (PRD §8.4), and confirm the entire §13 acceptance suite passes against the built binary. The main `--link` section was already updated Mode A by P1.M1.T1.S1; this sweep closes the two remaining cross-cutting drift spots (the Error-contract/mode-flags discussion and the completions value-completion enumeration) and runs the §13 gate.

**Deliverable**: (1) Two small, voice-consistent edits to `/home/dustin/projects/skilldozer/README.md` — the Error-contract paragraph (~line 152) and the completions value-completion bullet (~line 340). (2) A documented run of the full PRD §13 acceptance suite (every line passes; the `--link` slice is the changeset-relevant contribution). No code, no completion files.

**Success Definition**: the README's Error contract distinguishes `--link` (collects one-or-more directory positionals) from the single-value flags (`--store`/`--search`/`--shell`); the completions section notes `--link` offers directories at **every** following positional (`--link d1 <tab>` …); no stale single-target `--link` wording remains anywhere in the repo; the §13 suite is green (incl. multi-link, multi-link-partial, link-non-skill-refused, link-missing-value); `go build/vet/test ./...` stays green (doc-only).

---

## User Persona (if applicable)

**Target User**: A `skilldozer` user reading the README overview to understand `--link`, and a reviewer/CI running the §13 acceptance gate after the multi-target changeset lands.

**Use Case**: (1) A user reads the Error contract to learn what exits 2 — and needs `--link` described as a multi-positional collector, not "a flag that takes a value." (2) A user tab-completes `skilldozer --link d1 <tab>` and the README tells them directories keep being offered for every positional. (3) A reviewer runs §13 to confirm the batch changeset didn't regress anything.

**User Journey**: before the sweep, the README's Error contract lumps `--link` with single-value flags (misleading), and the completions section implies `--link` takes a single directory; after, both spots accurately describe the batch/multi-position behavior, and §13 is documented green.

**Pain Points Addressed**: cross-cutting doc drift that the per-subtask Mode A edits (which updated the main `--link` section + code) did not carry into the README's overview paragraphs; an unverified §13 gate.

---

## Why

- **Closes the Mode B documentation gap** (contract DOCS §5). P1.M1.T1.S1 (Mode A) updated the main `--link` section + quick-start + code; P1.M2.T1.S1 (parallel) updates the completion FILES. Neither touched the README's Error-contract paragraph or the completions value-completion bullet — those still describe `--link` as a single-value flag. This sweep is the one task that reconciles the overview.
- **Accuracy for the two drifted spots.** The Error contract says "flags that take a value — `--store`, `--search`, `--shell`, and `--link`" — but `--link` collects one-or-more positionals, not "a value." The completions bullet groups `--link` with `--init`/`--store` (single directory) — but after P1.M2.T1.S1, `--link` offers directories at every positional. Both contradict PRD §8.4 / §6.1.
- **Verifies the §13 acceptance gate.** The changeset added explicit §13 cases (multi-link, mixed-batch partial, single-bad-dir, missing-value). Running the full suite confirms the changeset is end-to-end sound and nothing regressed.
- **Consumed by**: end users reading the README and the §13/§16 acceptance review.

---

## What

[Mode B] Two README edits + one §13 run. `--link` is described consistently with batch behavior across all sections.

**(Edit 1 — Error contract, ~line 152).** Pull `--link` out of the "flags that take a value" group: state it is an exclusive mode that collects **one or more** directory positionals (so `--link` with nothing after it exits 2 rather than linking nothing), and separately note the single-value flags (`--store`/`--search`/`--shell`) exit 2 when given as the last token.

**(Edit 2 — completions value-completion, ~line 340).** Pull `--link` out of the `--init`/`--store` "single directory" group: state `--link` offers directories too, and keeps offering them at **every** following positional (`--link d1 <tab>`, `--link d1 d2 <tab>`, …), because `--link` batches one or more directories.

**(§13 run).** Execute the full PRD §13 acceptance script against `go build -o skilldozer .`; document that every line passes (the `--link` slice is verified; the `pi` line is skipped/noted if `pi` is absent).

### Success Criteria

- [ ] The README Error contract distinguishes `--link` (one-or-more directory positionals) from `--store`/`--search`/`--shell` (single-value flags); both still exit 2 on a missing value.
- [ ] The README completions section states `--link` offers directories at every following positional (multi-position), citing the `--link d1 <tab>` form.
- [ ] `grep` finds NO remaining wording that calls `--link` a single-value/single-directory flag (the only `--link` mentions describe batch/multi-position).
- [ ] The full PRD §13 acceptance suite passes (document the run); incl. multi-link, multi-link-partial, link-non-skill-refused, link-missing-value.
- [ ] Only `README.md` is modified; `main.go`/`main_test.go`/`internal/*`/`completions/*`/`install.sh` unchanged.
- [ ] `go build ./... && go vet ./... && go test ./...` all green (doc-only — scope-discipline guard).
- [ ] No §-PRD citations added to the edited paragraphs (match local voice: those sections use zero `§` cites).

---

## All Needed Context

### Context Completeness Check

**Pass.** The two edit zones were read verbatim (current text quoted in research/verified_facts.md §4, with exact anchor strings). The batch contract is fixed by PRD §8.4 / §6.1 (§1) and confirmed against the built binary (§2: all 7 `--link` §13 cases PASS). The two drifted spots were identified by `grep -n -- '--link' README.md` and cross-checked against the rest of the README (the main section + quick-start + advertised flag list are already batch-correct — §3). The completions edit documents the behavior P1.M2.T1.S1 lands (multi-position dir completion) — read its PRP as a contract. The README voice rules are grep-confirmed (3 §-cites total; the edited sections use zero; bold `**…**` is the house emphasis). The §13 script is the verbatim PRD §13 block. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (drift spots, exact edits, §13 results, voice, scope, conflict)
- file: plan/006_bab1774043df/P1M3T1S1/research/verified_facts.md
  why: "§1 = the --link batch contract (PRD §8.4/§6.1). §2 = §13 --link slice VERIFIED PASSING (all 7
        cases). §3 = README sections already batch-correct (NO edit: quick-start, main section,
        advertised flag list, constraints). §4 = the TWO edit spots with EXACT current→target text.
        §5 = sweep checks (c)/(d) — no further drift. §6 = voice rules. §7 = parallel-conflict
        analysis (P1.M2.T1.S1 edits completions only; disjoint from README). §8 = scope boundary."
  critical: "§4 — the two edits are the ENTIRE README change; both target texts are fixed verbatim.
             §3 — do NOT re-edit the main --link section or quick-start (P1.M1.T1.S1's Mode A; already
             correct). §7 — Edit 2 documents P1.M2.T1.S1's landed multi-position dir completion; at
             execution time (P1.M3 after P1.M2) it is in place. Pin seams by ANCHOR TEXT, not line."

# MUST READ — the file under edit (read the Error-contract para + completions section in full first)
- file: README.md
  why: "THE edit target. (Edit 1) the **Error contract.** paragraph in ## Usage (~line 145-155): the
        last 3 lines ('...the whole store). `--link` is another such mode. The flags that take a
        value — `--store`, `--search`, `--shell`, and `--link` — all exit 2...'). (Edit 2) the
        completions value-completion bullet in ## Shell completions (~line 340). Both anchors are
        UNIQUE (grep-confirmed). Read both sections in full before editing."
  pattern: "README prose = plain declarative sentences, inline `code` for flags/commands, bold
            `**emphasis**` for the key term, em dashes. The edited sections use ZERO §-PRD cites."
  gotcha: "Do NOT re-edit the main '### Linking skills from elsewhere' section (~line 159) or the
           quick-start (~line 128) — P1.M1.T1.S1 already made them batch-correct. Do NOT touch the
           advertised flag list (~line 336, already includes --link)."

# MUST READ — the §13 acceptance script (run it verbatim; document every line)
- file: PRD.md
  why: "§13 (h2.12) is the authoritative acceptance block. Run it against `go build -o skilldozer .`.
        The --link slice (multi-link / multi-link-partial / link-non-skill-refused / link-missing-value
        / link / resolve-linked / link-refresh) is this changeset's contribution and is VERIFIED PASSING
        (research §2). The pi line needs `pi` installed — skip + note if absent."
  section: "h2.12 (§13 — the full bash block)."
  critical: "Every line that echoes an 'OK' must print OK. If any fails, STOP, document the failure +
             root cause (the contract (a) requires this). Do NOT silently skip a failing line."

# MUST READ — the batch contract (the behavior the README must describe)
- file: PRD.md
  why: "§8.4 (h3.11) 'Linking an external skill' = the batch semantics (one --link, every following
        positional is a directory; partial success; exit 0/1/2). §6.1 (h3.1) line 138: '--link accepts
        one or more positional directories … the sole mode that collects trailing positionals.' These
        are the source of truth the README overview must match."
  section: "h3.11 (§8.4), h3.1 (§6.1 --link row + the mode-flags paragraph)."

# READ-ONLY — the parallel sibling (defines the multi-position dir completion Edit 2 documents)
- file: plan/006_bab1774043df/P1M2T1S1/PRP.md
  why: "P1.M2.T1.S1 makes `--link d1 <tab>` offer DIRECTORIES at every positional after --link (bash
        scan guard / zsh args-case branch / fish dir directive). Edit 2 documents exactly this: '--link
        offers directories at every following positional.' P1.M2.T1.S1 edits the three completion FILES
        only — DISJOINT from README (no conflict). At execution time (P1.M3 after P1.M2) it is landed."
  critical: "Edit 2's wording ('keeps offering them at every following positional … because --link "
             "batches one or more directories') must match P1.M2.T1.S1's landed behavior — read its
             PRP to confirm the multi-position dir contract before writing the README line."

# READ-ONLY — the Go impl (already Complete; confirms the batch semantics)
- file: plan/006_bab1774043df/P1M1T1S1/PRP.md
  why: "P1.M1.T1.S1 implemented the Go multi-target --link (struct/parser/runLink/tests) AND updated
        README's MAIN --link section + quick-start (Mode A). Confirms the main section is already
        batch-correct (do NOT re-edit it) and the quick-start already shows the batch form."

# READ-ONLY — confirms PRD §17/§6.1 describe --link batch-correctly (no PRD drift to mirror)
- file: PRD.md
  why: "§17 (h2.16) line 120 + §6.1 line 138 + §8.4 line 149 all describe --link as batch/multi-target.
        So the PRD is consistent; only the README OVERVIEW drifted. Do NOT edit PRD.md (read-only)."
  section: "h2.16 (§17 --link row), h3.1 (§6.1), h3.11 (§8.4)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && grep -n -- '--link' README.md   # the --link mentions
128:skilldozer --link ~/projects/agent-browser   # quick-start (single) — already batch-correct
129:skilldozer --link ~/projects/a ~/projects/b  # quick-start (batch)  — already batch-correct
153:... `--link` is another such mode. The flags that take a value —          ← EDIT 1 (Error contract)
154:`--store`, `--search`, `--shell`, and `--link` — all exit 2 ...           ← EDIT 1
336:  `--completions`, ..., `--init`, `--link`, `--list`, ...                 ← advertised list (has --link; NO edit)
340:- `skilldozer --init <tab>`, `skilldozer --link <tab>`, and ...           ← EDIT 2 (completions enumeration)
# (line 159-190 = main --link section — P1.M1.T1.S1 Mode A, already correct, NO edit)
# main.go / main_test.go / internal/* / completions/* — UNCHANGED.
```

### Desired Codebase tree with files to be changed

```bash
README.md    # EDIT 1: Error-contract paragraph (~line 152) — pull --link out of single-value group.
             # EDIT 2: completions value-completion bullet (~line 340) — note multi-position dir completion.
# every other file UNCHANGED (no Go, no completions, no PRD, no .gitignore).
```

| File | Change | Why |
|---|---|---|
| `README.md` | (1) Error-contract para: `--link` = multi-positional collector, not a single-value flag. (2) Completions bullet: `--link` offers dirs at every positional. | Match PRD §8.4/§6.1 batch semantics + P1.M2.T1.S1's landed completion behavior. |

### Known Gotchas of our codebase & Library Quirks

```markdown
<!-- GOTCHA #1 (CRITICAL — --link is NOT a single-value flag) — the Error contract currently groups --link
     with --store/--search/--shell ("flags that take a value"). That is WRONG for --link: it COLLECTS one
     or more directory positionals (a batch), per PRD §8.4/§6.1. Pull it out and describe the multi-
     positional collection. The exit-2-on-missing fact stays TRUE for --link (zero dirs ⇒ exit 2) — keep
     it, just don't phrase it as "guessing a value." (research §4 Edit 1.) -->

<!-- GOTCHA #2 (CRITICAL — --link completion is MULTI-POSITION) — the completions bullet groups --link
     with --init/--store ("a single directory"). After P1.M2.T1.S1, --link offers directories at EVERY
     following positional (--link d1 <tab> → dirs, --link d1 d2 <tab> → dirs, …). The README must convey
     this multi-position behavior, citing the --link d1 <tab> form. (research §4 Edit 2.) -->

<!-- GOTCHA #3 (do NOT re-edit the already-correct sections) — the main '### Linking skills from
     elsewhere' section (~line 159), the quick-start (~line 128), and the advertised flag list
     (~line 336, already includes --link) are ALL already batch-correct (P1.M1.T1.S1 Mode A). This task
     edits ONLY the Error-contract paragraph and the completions value-completion bullet. -->

<!-- GOTCHA #4 (VOICE — no § citations in the edited paragraphs) — `grep -no '§[0-9.]*' README.md` shows
     3 cites total, NONE in the Error-contract or completions-enumeration sections. Keep new prose
     citation-free (plain words). Bold **one or more** / **every** is the house emphasis style. -->

<!-- GOTCHA #5 (§13 — document EVERY line; the pi line is conditional) — run the full PRD §13 block.
     Every line that echoes an 'OK' must print OK. The `pi --no-skills --skill ...` line needs `pi`
     installed — if absent, skip it and NOTE that in the run output (the contract (a) wants the run
     documented; a missing `pi` is an environment limitation, not a changeset failure). The --link slice
     is the changeset's contribution and is verified passing (research §2). -->

<!-- GOTCHA #6 (NO CODE / NO COMPLETIONS / NO PRD) — this is doc-only (Mode B). Do NOT edit main.go,
     main_test.go, internal/*, completions/* (P1.M2.T1.S1 owns those), install.sh, PRD.md, tasks.json,
     prd_snapshot.md, .gitignore. `go build/vet/test ./...` must stay green (the scope-discipline guard). -->

<!-- GOTCHA #7 (REBUILD not needed for README) — README is NOT //go:embed'd (only the completion files
     are). A README edit needs NO rebuild. (Do NOT confuse this with the completions embed lockstep.)
     `go test ./...` is the scope guard, not an embed check. -->

<!-- GOTCHA #8 (anchor by TEXT, not line number) — the README may shift as siblings land (P1.M1.T1.S1
     already touched ~line 128/159). Locate Edit 1 by the 'The flags that take a value' sentence and
     Edit 2 by the 'skilldozer --init <tab>, skilldozer --link <tab>, and skilldozer --store <tab>'
     bullet. Both are grep-unique. -->
```

---

## Implementation Blueprint

### Data models and structure

**None.** Documentation-only. No code, no types, no config. The two edited paragraphs + the §13 run ARE the deliverable.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: READ README + run the §13 acceptance suite (the verify half of the deliverable)
  - READ README.md end-to-end; confirm the two edit zones (Error contract ~line 152; completions
    bullet ~line 340) and that the main --link section / quick-start / advertised list are already
    batch-correct (research §3).
  - RUN the full PRD §13 block against `go build -o skilldozer .`. Document every line. The --link
    slice (multi-link / multi-link-partial / link-non-skill-refused / link-missing-value / link /
    resolve-linked / link-refresh) MUST all echo OK (verified passing — research §2). If `pi` is
    absent, skip + note the pi line. If ANY other line fails, STOP and document the failure + root
    cause (contract (a) requires this).

Task 1: EDIT README — the Error-contract paragraph (Edit 1)
  - FILE: README.md (the **Error contract.** paragraph in ## Usage)
  - ANCHOR (locate by this text — the last 3 lines of the paragraph; grep-unique):
        the whole store). `--link` is another such mode. The flags that take a value —
        `--store`, `--search`, `--shell`, and `--link` — all exit 2 when given as the
        last token with nothing after them, rather than guessing a value.
  - REPLACE with (pull --link out; describe multi-positional collection; keep exit-2 for all):
        the whole store). `--link` is another exclusive mode: it collects **one or more**
        directory positionals, so `--link` with nothing after it exits 2 rather than
        linking nothing. The single-value flags — `--store`, `--search`, and `--shell` —
        likewise exit 2 when given as the last token with nothing after them, rather than
        guessing a value.
  - PRESERVE the paragraph's opening sentences (unknown tag → stdout empty + exit 1; multi-tag
    atomicity; the --path/--list/--search/--all mutual-exclusivity + tag+mode sentence) UNCHANGED.
    Keep the `**Error contract.**` bold lead-in.

Task 2: EDIT README — the completions value-completion bullet (Edit 2)
  - FILE: README.md (the bullet in ## Shell completions that lists --init/--link/--store/--search/--shell)
  - ANCHOR (locate by this text — grep-unique):
        - `skilldozer --init <tab>`, `skilldozer --link <tab>`, and `skilldozer --store <tab>`
          offer directories (a path value); `skilldozer --search <tab>` offers nothing
          (free-text); `skilldozer --shell <tab>` offers the three supported shells —
          `bash`, `zsh`, and `fish`.
  - REPLACE with (pull --link out; note multi-position dir completion — mirrors P1.M2.T1.S1):
        - `skilldozer --init <tab>` and `skilldozer --store <tab>` offer a single directory
          (a path value); `skilldozer --search <tab>` offers nothing (free-text);
          `skilldozer --shell <tab>` offers the three supported shells — `bash`, `zsh`, and
          `fish`. `skilldozer --link <tab>` offers directories too, and keeps offering them
          at **every** following positional (`--link d1 <tab>`, `--link d1 d2 <tab>`, …),
          because `--link` batches one or more directories.
  - PRESERVE the surrounding bullets (skills-first bare <tab>; long-form-only -<tab> flag list)
    UNCHANGED. The advertised flag list (~line 336) already includes --link — do NOT touch it.

Task 3: VERIFY — edits + §13 + scope/discipline
  - grep checks (see Validation Loop Level 2): the 2 edits present; NO stale single-value --link wording.
  - the full §13 suite green (documented in Task 0).
  - go build/vet/test ./... green (doc-only scope guard); git diff --name-only = ONLY README.md.
```

### Implementation Patterns & Key Details

```markdown
# Both edits follow the same shape: PULL --link out of a group that implies a single value, and give
# it its own clause that states the multi-positional/batch behavior. Keep the exit-2-on-missing fact
# (it is TRUE for --link: zero dirs ⇒ exit 2) — just don't phrase --link as "taking a value."

# Edit 1 in context (Error contract, before -> after):
#   BEFORE (tail of the paragraph):
#     ... combining a tag with any of them (a tag resolves one path; those modes inspect
#     the whole store). `--link` is another such mode. The flags that take a value —
#     `--store`, `--search`, `--shell`, and `--link` — all exit 2 when given as the
#     last token with nothing after them, rather than guessing a value.
#
#   AFTER (tail of the paragraph):
#     ... combining a tag with any of them (a tag resolves one path; those modes inspect
#     the whole store). `--link` is another exclusive mode: it collects **one or more**
#     directory positionals, so `--link` with nothing after it exits 2 rather than
#     linking nothing. The single-value flags — `--store`, `--search`, and `--shell` —
#     likewise exit 2 when given as the last token with nothing after them, rather than
#     guessing a value.

# Edit 2 in context (completions bullet, before -> after):
#   BEFORE:
#     - `skilldozer --init <tab>`, `skilldozer --link <tab>`, and `skilldozer --store <tab>`
#       offer directories (a path value); `skilldozer --search <tab>` offers nothing
#       (free-text); `skilldozer --shell <tab>` offers the three supported shells —
#       `bash`, `zsh`, and `fish`.
#
#   AFTER:
#     - `skilldozer --init <tab>` and `skilldozer --store <tab>` offer a single directory
#       (a path value); `skilldozer --search <tab>` offers nothing (free-text);
#       `skilldozer --shell <tab>` offers the three supported shells — `bash`, `zsh`, and
#       `fish`. `skilldozer --link <tab>` offers directories too, and keeps offering them
#       at **every** following positional (`--link d1 <tab>`, `--link d1 d2 <tab>`, …),
#       because `--link` batches one or more directories.
```

Notes easy to get wrong:
- Phrasing `--link` as "a flag that takes a value" (GOTCHA #1) — it collects multiple positionals.
- Implying `--link` completes a single directory (GOTCHA #2) — it completes dirs at every positional.
- Re-editing the already-correct main section / quick-start / advertised list (GOTCHA #3).
- Adding a `§` citation to the edited paragraphs (GOTCHA #4) — those sections use none.
- Skipping a failing §13 line silently (GOTCHA #5) — document failures + root cause.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Two edits, not more.** `grep -n -- '--link' README.md` shows every `--link` reference is batch-aware EXCEPT the Error-contract paragraph and the completions bullet. The main section, quick-start, and advertised flag list are already correct (P1.M1.T1.S1 Mode A). Editing only the two drifted spots is the minimal, warranted change. (research §3/§5.)
2. **Pull `--link` OUT of the single-value groups (don't append a footnote).** Both drifted spots LUMP `--link` with single-value flags; the cleanest fix is to separate it into its own clause that states the multi-positional/batch behavior. A footnote would leave the misleading grouping in place. (GOTCHA #1/#2.)
3. **Edit 2 documents P1.M2.T1.S1's landed behavior.** The completions bullet must match what the completion files now DO (`--link d1 <tab>` → dirs at every positional). P1.M3 runs after P1.M2, so the behavior is in place; read P1.M2.T1.S1's PRP to confirm the multi-position contract before wording the line. (research §7.)
4. **Keep the exit-2-on-missing fact for `--link`.** It is TRUE (zero dirs ⇒ exit 2, PRD §8.4). The drift is the "takes a value / guessing a value" phrasing, not the exit code. The fix restates exit 2 accurately ("linking nothing"). (GOTCHA #1.)
5. **§13 is run + documented, not just asserted.** The contract (a) requires running it. The `--link` slice is verified passing; the implementer runs the FULL suite and records the output (the pi line is conditional on `pi` being installed). (GOTCHA #5.)
6. **No § citations in the edited paragraphs.** The README tolerates § cites (3 total) but the edited sections use none; match the local voice. Bold `**…**` is the house emphasis. (GOTCHA #4.)

### Integration Points

```yaml
DOCUMENTATION (Mode B — the deliverable IS the README edits + the §13 run):
  - file: README.md (Error-contract paragraph + completions value-completion bullet)
  - effect: "README overview now describes --link as a multi-positional batch collector consistently
            with PRD §8.4/§6.1 and the landed completion behavior (P1.M2.T1.S1)."

ACCEPTANCE (§13 — run + document):
  - the full PRD §13 block against `go build -o skilldozer .`; every 'OK' line prints OK; the --link
    slice (this changeset's contribution) is the focus. Document the run output in the task result.

CODE: NONE.
  - main.go / main_test.go / internal/* / completions/* / install.sh UNCHANGED. README is NOT
    //go:embed'd (only completion files are) → no rebuild needed. `go build/vet/test ./...` is the
    scope-discipline guard, not an embed check. (GOTCHA #6/#7.)

PRD.md / tasks.json / prd_snapshot.md / .gitignore: READ-ONLY.

PARALLEL SIBLING (no conflict):
  - P1.M2.T1.S1 edits the three completion FILES only; this task edits README.md ONLY. DISJOINT.
  - P1.M1.T1.S1 (Complete) edited README's main --link section + quick-start; this task's edits are
    in DIFFERENT regions (Error contract, completions bullet). No overlap.

NO DATABASE / NO ROUTES / NO CONFIG / NO GO CODE / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Edit presence + voice (immediate, after each edit)

```bash
cd /home/dustin/projects/skilldozer

# Edit 1 — Error contract no longer groups --link with single-value flags:
grep -q -- '--link is another exclusive mode' README.md && echo "Edit1: --link separated OK"
grep -q 'collects \*\*one or more\*\*' README.md && echo "Edit1: multi-positional stated OK"
# (the OLD grouping "flags that take a value — \`--store\`, \`--search\`, \`--shell\`, and \`--link\`" is GONE:)
! grep -q 'and `--link` — all exit 2 when given as the' README.md && echo "Edit1: old grouping removed OK"

# Edit 2 — completions bullet notes multi-position dir completion:
grep -q 'keeps offering them at \*\*every\*\* following positional' README.md && echo "Edit2: multi-position stated OK"
grep -q -- '--link d1 <tab>' README.md && echo "Edit2: --link d1 <tab> form cited OK"
# (the OLD grouping "skilldozer --init <tab>, skilldozer --link <tab>, and skilldozer --store <tab>
#  offer directories (a path value)" is GONE:)
! grep -q 'skilldozer --link <tab>`, and `skilldozer --store <tab>`' README.md && echo "Edit2: old grouping removed OK"

# Voice — no § citations in the edited paragraphs (the 3 § cites are elsewhere: §8.4 ×2, §6.1 ×1):
grep -c '§' README.md   # expect 3 (unchanged — no new cites added)

# Render check — eyeball both edited zones:
sed -n '/^\*\*Error contract\.\*\*/,/^`skilldozer --help`/p' README.md
sed -n '/--init <tab>/p' README.md
# Expected: Edit 1 separates --link; Edit 2 pulls --link out + cites --link d1 <tab>; markdown intact.
```

### Level 2: Sweep checks (c) — no stale single-target --link wording anywhere

```bash
cd /home/dustin/projects/skilldozer

# (c) No remaining wording calls --link a single-value/single-directory flag. The --link mentions
# should all describe batch/multi-position:
grep -n -- '--link' README.md
# Manually confirm each line is batch-aware. The two Edit spots now describe multi-positional; the
# main section (~line 159), quick-start (~line 128), and advertised list (~line 336) were already correct.
# Specifically assert the misleading phrases are GONE:
! grep -qi 'flags that take a value.*--link' README.md && echo "(c): no 'takes a value + --link' OK"
! grep -q 'skilldozer --link <tab>`, and `skilldozer --store <tab>`\n *offer directories (a path value)' README.md \
  && echo "(c): no '--link single directory' grouping OK"

# (d) Constraints section needs no --link mention (the never-clobber refuse behavior is in the main
# --link section); verify it is consistent (no contradiction):
grep -A2 '## Constraints' README.md | grep -i -- '--link' && echo "(d): NOTE --link in Constraints" || echo "(d): Constraints clean (no --link guardrail needed) OK"
# Expected: (c) clean; (d) Constraints clean.
```

### Level 3: §13 acceptance suite (the contract (a) gate — RUN + document)

```bash
cd /home/dustin/projects/skilldozer

# Build fresh:
go build -o skilldozer . && echo "build OK"

# Run the FULL PRD §13 block (copy it verbatim from PRD.md §13). Every line that echoes an 'OK' must
# print OK. The --link slice is this changeset's contribution (verified passing — research §2):
#   link OK | resolve-linked OK | link-refresh OK | multi-link OK | multi-link-partial OK |
#   link-non-skill-refused OK | link-missing-value OK
# The pi line (`pi --no-skills --skill "$(./skilldozer example)" ...`) needs `pi` installed:
#   - if `pi` is present: confirm its output references the example skill / does not error.
#   - if `pi` is absent: skip it and NOTE "pi not installed — pi acceptance line skipped" in the result.
# Document the full run output in the task result (contract (a): "document the run output").

# Quick re-verification of the --link slice (deterministic, no pi needed):
rm -rf /tmp/sd-link && mkdir -p /tmp/sd-link/store /tmp/sd-link/src/linked
printf -- '---\nname: linked\ndescription: A linked skill.\n---\n# body\n' > /tmp/sd-link/src/linked/SKILL.md
test "$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked)" = "/tmp/sd-link/store/linked" && echo "link OK"
rm -rf /tmp/sd-link/store && mkdir -p /tmp/sd-link/store /tmp/sd-link/src/other /tmp/sd-link/notaskill
printf -- '---\nname: other\ndescription: Another linked skill.\n---\n# body\n' > /tmp/sd-link/src/other/SKILL.md
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked /tmp/sd-link/src/other)
printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked' && printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/other' && echo "multi-link OK"
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked /tmp/sd-link/src/other /tmp/sd-link/notaskill 2>/tmp/e); rc=$?
[ "$rc" = "1" ] && printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked' && grep -q notaskill /tmp/e && echo "multi-link-partial OK"
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/notaskill 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "1" ] && echo "link-non-skill-refused OK"
./skilldozer --link >/dev/null 2>&1; [ "$?" = "2" ] && echo "link-missing-value OK"
rm -rf /tmp/sd-link ./skilldozer
# Expected: all --link cases echo OK (mirrors §13). Run the remaining §13 lines (discovery/path/list/
# resolve/error-contract/completions/check/init/config) from the PRD block too; they are pre-existing
# and unaffected but the contract wants the FULL suite green.
```

### Level 4: Scope discipline + doc-sanity (proves doc-only)

```bash
cd /home/dustin/projects/skilldozer

# Only README.md changed by THIS subtask:
git status --short README.md             # expect " M README.md"
git diff --name-only | grep -vE '^plan/' # expect ONLY README.md

# No Go/completions/PRD change:
git diff --quiet main.go main_test.go && echo "main.go/main_test.go unchanged"
git diff --quiet go.mod go.sum && echo "deps unchanged"
ls completions/ # unchanged (P1.M2.T1.S1's territory, parallel)

# Doc-sanity: build/vet/test stay green (README is not //go:embed'd; no rebuild semantics):
go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0
# Expected: README.md only changed; build/vet/test all exit 0.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — both edits present; old groupings gone; `grep -c '§' README.md` == 3 (no new cites); markdown intact
- [ ] Level 2 PASS — (c) no stale single-target `--link` wording anywhere; (d) Constraints consistent
- [ ] Level 3 PASS — full §13 suite green (documented); --link slice (7 cases) all OK; pi line run or noted-skipped
- [ ] Level 4 PASS — `git diff --name-only` = ONLY README.md; `go build/vet/test ./...` all exit 0

### Feature Validation
- [ ] Error contract: `--link` described as collecting one-or-more directory positionals (not a single value); single-value flags (`--store`/`--search`/`--shell`) separated; all exit 2 on missing value
- [ ] Completions: `--link` offers directories at every following positional (`--link d1 <tab>` cited)
- [ ] §13 acceptance: every line passes (run output documented); the `--link` multi-target cases are the changeset focus

### Code Quality / Convention Validation
- [ ] Matches README voice: plain prose, inline `` `code` ``, bold `**…**` emphasis, NO `§` cites in the edited paragraphs
- [ ] Edits minimal (only the 2 drifted spots); the already-correct main section / quick-start / advertised list untouched
- [ ] Seams located by anchor text (robust to line shifts from sibling landings)

### Scope Discipline
- [ ] Edited README.md ONLY (contract DOCS §5 Mode B)
- [ ] Did NOT touch main.go, main_test.go, internal/*, completions/* (P1.M2.T1.S1), install.sh
- [ ] Did NOT modify PRD.md (read-only), tasks.json, prd_snapshot.md, or .gitignore
- [ ] Did NOT re-edit the main --link section or quick-start (P1.M1.T1.S1 Mode A — already correct)

---

## Anti-Patterns to Avoid

- ❌ **Don't phrase `--link` as "a flag that takes a value."** It collects one-or-more directory positionals (a batch). Pull it out of the single-value group; restate exit 2 as "linking nothing." (GOTCHA #1.)
- ❌ **Don't imply `--link` completes a single directory.** After P1.M2.T1.S1 it offers directories at every positional (`--link d1 <tab>`). Cite the multi-position form. (GOTCHA #2.)
- ❌ **Don't re-edit the already-correct sections.** The main `--link` section, quick-start, and advertised flag list are batch-correct (P1.M1.T1.S1). Edit ONLY the Error-contract paragraph + completions bullet. (GOTCHA #3.)
- ❌ **Don't add `§` citations to the edited paragraphs.** Those sections use zero; `grep -c '§'` must stay 3. (GOTCHA #4.)
- ❌ **Don't skip a failing §13 line silently.** Run the full suite; document every line; if a line fails, STOP and record the failure + root cause (contract (a)). The pi line is conditional on `pi` being installed — note-skip if absent. (GOTCHA #5.)
- ❌ **Don't edit code, completions, or PRD.** Doc-only (Mode B). `go build/vet/test` is the scope guard. README is not `//go:embed`d — no rebuild needed. (GOTCHA #6/#7.)
- ❌ **Don't locate seams by line number.** The README shifts as siblings land; anchor by the exact sentences in research §4. (GOTCHA #8.)

---

## Confidence Score

**9.5/10** — one-pass success likelihood. The two edit zones were read verbatim with exact current→target text (research §4), and the batch contract is fixed by PRD §8.4/§6.1 and confirmed against the built binary (the §13 `--link` slice's 7 cases all PASS — §2). The two drifted spots are the ONLY stale `--link` wording (`grep`-confirmed; the main section/quick-start/advertised list are already correct). Edit 2 documents P1.M2.T1.S1's landed multi-position dir completion (read its PRP as a contract; P1.M3 runs after P1.M2). The README voice rules are grep-confirmed (3 §-cites; edited sections use none). The §13 script is the verbatim PRD block. The file set is fully disjoint from the parallel sibling (P1.M2.T1.S1 edits completions only) and from P1.M1.T1.S1's README regions (main section/quick-start). The 0.5 reservation is for the two one-pass slips the PRP cannot fully mechanize away — (a) leaving `--link` in the single-value grouping (caught by the Level 1 `! grep` checks) and (b) a §13 line failing unexpectedly (the contract requires documenting it, and the `--link` slice is pre-verified) — both caught by the validation gates.
