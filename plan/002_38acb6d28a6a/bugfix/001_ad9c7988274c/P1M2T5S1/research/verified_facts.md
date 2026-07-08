# Verified facts — P1.M2.T5.S1 (README reserved-tag-names note, Issue 7)

All facts verified directly against the working tree @ 2026-07-07. This is a
**documentation-only** (Mode A) task: the entire deliverable is an edit to
`README.md`'s "Where skills live" section. No Go code, no test, no PRD.md edit.

---

## 1. The edit target — README.md "Where skills live" section (lines 174–200)

`grep -n '^## ' README.md` places the section boundaries precisely:

```
174:## Where skills live          ← SECTION START
...
201:## Adding a skill             ← NEXT SECTION (the insertion sits BEFORE this)
```

The full section body (verified with `sed -n '174,200p' README.md`):

```
## Where skills live

Skills live in the `skills/` directory at the repo root. A skill is any
directory that directly contains a `SKILL.md`.

The canonical **tag** is the skill directory's path **relative to `skills/`**,
with `/` separators. It is **not** the frontmatter `name`.

```text
skills/example/SKILL.md              → tag example
skills/writing/reddit/SKILL.md       → tag writing/reddit
```

Nested skills count, so `skills/writing/reddit/SKILL.md` is addressed as
`writing/reddit`, not `reddit`.

Tag resolution tries, in order:

1. the exact canonical tag (`writing/reddit`)
2. the final path segment / basename (`reddit`)
3. the frontmatter `name`
4. a declared alias (see `metadata.aliases`)
5. else: unknown

So `skilldozer example`, `skilldozer writing/reddit`, `skilldozer reddit` (if unique), and
`skilldozer foo-helper` (matching a frontmatter `name`) all resolve.

## Adding a skill
```

**INSERTION POINT (the natural seam):** immediately AFTER the final paragraph of
the section — the line ending `… all resolve.` (README.md:199) — and BEFORE the
blank line + `## Adding a skill` (README.md:200–201). This reads as "here is the
one carve-out to the tag-resolution rule", which is exactly the framing the
contract asks for. Putting it earlier (e.g. right after the canonical-tag
definition) would interrupt the resolution-precedence flow; putting it after
`## Adding a skill` leaves the section that defines tags without the caveat.

The exact anchor text for the edit (verified byte-for-byte with `cat -A`, so the
line-endings are known — plain `\n`, no trailing spaces):

```
`skilldozer foo-helper` (matching a frontmatter `name`) all resolve.

## Adding a skill
```

This block is unique in README.md (the only `## Adding a skill` heading; the
phrase `all resolve.` appears once), so it is a safe surgical-edit anchor.

---

## 2. README voice — the note MUST match the existing register

- `grep -n '^\*\*[A-Z]' README.md` → the README's established "callout" pattern
  is a **bold lead-in phrase followed by a period**, then prose. Confirmed
  existing instance: README.md:164 `**Error contract.** An unknown tag prints …`.
  So the new note's lead-in MUST be `**Reserved tag names.**` (bold + period) to
  match the house style. (README.md:23/36/46 use the same `**A./B./C.**` lead-in
  for install paths.)
- **VOICE RULE (critical):** `grep -n '§' README.md` → **NONE.** The user-facing
  README NEVER cites PRD section numbers (no `§7.2`, `§8.2`, `§9`). The contract's
  "(§8.2, §9)" is the contract author telling the PRP writer where check/init are
  defined in the PRD — it is NOT an instruction to put §-citations in the README.
  Adding `§8.2`/`§9` to the note would BREAK README voice consistency. The note
  must say "runs validation" / "runs first-run setup" in plain language instead.
- Tone: declarative, terse, one-sentence-per-idea, inline `` `code` `` for
  commands/paths/tags (e.g. `` `skilldozer check` ``, `` `writing/check` ``).
  Matches the surrounding "So `skilldozer example` … all resolve." register.

---

## 3. The binding decision — decisions.md §D6 (documentation-only, NOT a code change, NOT a PRD edit)

`plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c/architecture/decisions.md` §D6,
verbatim core:

> `check`/`init` reservation is deliberate and documented in code. A code change
> to resolve a skill named `check` would silently shadow the `skilldozer check`
> subcommand — a worse UX surprise. Decision: record the reservation + workarounds
> in this decisions.md (D6) and **surface it in the README's skill-tag section**.
> PRD.md is human-owned/READ-ONLY — the architect does NOT edit it. The §7.2 note
> suggested by the PRD is a human action item recorded here, not executed.

**Implication for THIS subtask:** the deliverable is a README edit ONLY. Do NOT
touch `main.go`, do NOT touch `PRD.md`, do NOT add a Go test. The reservation is
already implemented in code and already commented as deliberate (see §4 below);
this task only documents it for users.

---

## 4. The code reality the note must describe (already implemented; do NOT change)

`grep -n 'reserved\|Reserved\|subcommand' main.go` confirms the reservation is in
`parseArgs` and is commented as deliberate:

- `case "check":` (main.go:254) — comment (main.go:254–262):
  > "`skilldozer check` subcommand (PRD §9). `check` is a RESERVED positional
  > token: it selects validation mode and is NOT captured as a tag. A skill
  > literally tagged `check` cannot be resolved via `skilldozer check` (**subcommand
  > names are reserved, as in any CLI**). … A nested skill `writing/check` still
  > resolves: this case matches only the EXACT token `check`."

- `case "init":` (main.go:269) — comment (main.go:269–290):
  > "`skilldozer init [<dir>]` first-run setup (PRD §8.2). `init` is a RESERVED
  > positional token (like `check`) … **GOTCHA: a store literally named
  > `check`/`init` must be passed via `--store`.**"

These two `case` arms are matched BEFORE the default tag-capture branch
(main.go:281+, the `default:` that appends a non-flag token to `c.tags`). That
ordering is WHY `check`/`init` never become tags. The README note is the
user-facing mirror of these two code comments — same rule, plain words.

**The exact phrasing to echo:** the code says "subcommand names are reserved, as
in any CLI". The contract says "Frame it as the standard CLI subcommand-reservation
rule." → the note should say "This is the standard CLI rule — a subcommand name
takes precedence over a positional argument" (or equivalent), so the README and
the code comment agree in spirit.

---

## 5. The four workarounds the contract requires (and the optional 5th)

CONTRACT LOGIC step 3 enumerates the workarounds the note MUST list for a skill
at `skills/check/SKILL.md` (canonical tag `check`):

1. **`--list` / `--all`** — discovery. `check` is still a valid skill on disk; it
   shows up in the catalog listings (the `case "check"` arm only blocks TAG
   *resolution*, not *discovery* — `--list`/`--all` walk the tree independently).
2. **A nested path** — e.g. `writing/check` (move/keep the skill one level deeper
   so its canonical tag is no longer the bare reserved token). The code comment
   explicitly notes "A nested skill `writing/check` still resolves".
3. **The frontmatter `name`** — the tag-resolution fallback #3 (README.md:194,
   PRD §7.2 step 3). A skill whose `name:` differs from its dir name resolves by
   `name` even if the canonical tag is shadowed.
4. **A declared alias** — `metadata.aliases` (README.md:195, PRD §7.2 step 4).

**Optional 5th (include for completeness — it is the same reserved-name rule,
documented in the code's own GOTCHA at main.go:281–290 and surfaced by §D6):**
   - to point `init` at a *store directory* literally named `check`/`init`, pass
     it with `--store` (e.g. `skilldozer init --store ./check`), not as the
     positional `init <dir>`. This is a different scenario (store dir, not skill
     tag) but is the identical reserved-token rule and prevents the exact
     confusion a user hitting this will have. RECOMMENDED to include as one short
     closing sentence; it is consistent with the code comment and §D6's "record
     the workarounds". The PRP marks it optional so the implementer can drop it
     if it feels off-topic for the skill-tag section.

---

## 6. Scope isolation — zero overlap with the parallel sibling (P1.M2.T4.S1)

- This subtask edits **ONLY** `README.md`.
- The parallel sibling **P1.M2.T4.S1** (Issue 6) edits **ONLY** `.gitignore`
  (its PRP's "Desired Codebase tree" lists exactly one file: `.gitignore`).
- DISJOINT file sets → no merge conflict; land in either order. Verified: T4's
  PRP (`P1M2T4S1/PRP.md`) scope-discipline explicitly excludes README, and this
  task explicitly excludes `.gitignore`.
- The later Mode B sweep **P1.M3.T1.S1** ("Sweep README.md init/error-contract
  sections for the changeset delta") is a *consistency verification* across all 7
  fixes; per the contract DOCS step 5, it "verifies README consistency across all
  7 fixes but does not duplicate this note." So this note is the canonical home
  for the reserved-tag rule; P1.M3.T1 must not re-state it.

---

## 7. Validation approach — there is NO markdown linter and NO Go test (grep + render checks)

- `ls -a | grep -iE 'markdownlint|mdlint|prettier|vale|\.github'` → **NONE.** No
  markdown linter is configured. No `.github/` workflow, no `Makefile`, no doc
  test target (`grep -in 'readme\|lint\|doc' Makefile` → no Makefile).
- Therefore the validation gates are: (a) grep that the note + each required
  keyword is present and inside the "Where skills live" section; (b) `cat`/render
  check that the markdown structure is intact (heading order, code fence balance,
  blank-line discipline); (c) `go build/vet/test ./...` stays green (proof of
  "no code change"); (d) `git status --short` shows ONLY `README.md` modified.
- Precedent for doc-only validation: the sibling `P1M2T4S1/PRP.md` (Issue 6,
  `.gitignore`) uses exactly this pattern — content-is-the-test, grep checks,
  no code regression proof. This PRP follows the same precedent.

---

## 8. Forbidden actions (re-stated for the implementer)

- ❌ Do NOT edit `PRD.md` (human-owned / READ-ONLY — §D6, and the global FORBIDDEN
  OPERATIONS). The §7.2 note the PRD "suggests" is a human action item, recorded
  in decisions.md §D6, NOT executed here.
- ❌ Do NOT edit `tasks.json` or `prd_snapshot.md` (orchestrator-owned).
- ❌ Do NOT change `main.go` or any Go file (documentation-only — §D6). The
  reservation is already implemented and commented; `go build/vet/test` must be
  unchanged.
- ❌ Do NOT add a Go test (there is nothing to test — the code already behaves as
  documented; `TestParseArgsCheckSubcommand` @main_test.go:1120 already asserts
  the reservation).
- ❌ Do NOT add `§8.2`/`§9` PRD citations to the README (voice rule, §2).
- ❌ Do NOT duplicate this note elsewhere (P1.M3.T1 verifies but does not
  re-state it — contract DOCS step 5).
