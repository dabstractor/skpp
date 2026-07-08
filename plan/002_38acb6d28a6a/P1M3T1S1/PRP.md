name: "P1.M3.T1.S1 — Rewrite skills/example/SKILL.md to match PRD §11 (skpp → skilldozer)"
description: >

---

## Goal

**Feature Goal**: Eliminate the last literal `skpp` residue in a non-core repo asset by rewriting `skills/example/SKILL.md` to be **byte-for-byte identical** to PRD §11's canonical example-skill block (== the rendered `exampleSkillTemplate` constant already compiled into `main.go`). This restores the intended `--search` behavior (matching `skilldozer`, not `skpp`) and re-syncs the two copies of the example-skill text that the contract requires to stay identical.

**Deliverable**: A single edited file — `skills/example/SKILL.md` — whose 7 `skpp` occurrences become `skilldozer` (description ×2, `metadata.keywords` last element ×1, body ×1, three bash try-command lines ×3), with frontmatter `name: example` and `category: meta` preserved unchanged and the exact column alignment + trailing-newline structure of §11 preserved.

**Success Definition**: `grep -c skpp skills/example/SKILL.md` returns `0`; `./skilldozer --search skilldozer` lists the `example` skill; `./skilldozer --search skpp` returns no match; `./skilldozer check` still prints `OK example (example)`; `./skilldozer example` still resolves; and the file content equals PRD §11 exactly (verifiable by diff against the seeded constant once `init` is live, or by grep assertions).

---

## User Persona (if applicable)

**Target User**: Anyone running `skilldozer --search` (or reading the shipped example to learn the frontmatter shape). Also the `skilldozer init` user: the seeded store's `example/SKILL.md` (from the constant) and the repo's shipped `skills/example/SKILL.md` are meant to be the SAME text.

**Use Case**: `skilldozer --search skilldozer` to find the example skill, or `skilldozer --search demo`; copying `skills/example/SKILL.md` as the starting template for a new skill.

**User Journey**: After this fix, searching for the product's own name returns the example skill (it didn't before), and the shipped example matches what `skilldozer init` would seed into a fresh store.

**Pain Points Addressed**: `--search` was functionally inverted for the canonical keyword (searching the product's name returned nothing); the shipped example contradicted both the PRD and the compiled-in seed template.

---

## Why

- **Fixes the only remaining literal `skpp` in a non-core file.** `grep -rn "skpp" --include="*.md"` (whole repo, excluding `plan/`) returns exactly one file: `skills/example/SKILL.md`, 7 lines. The `skpp`→`skilldozer` rename landed everywhere else; this asset was missed. (architecture/docs_and_assets_drift.md §1, verdict ❌.)
- **Restores PRD §6.1/§10 `--search` correctness.** `metadata.keywords` is one of the 6 fields `internal/search/search.go` `matches()` searches (each keyword tested individually). With `keywords: [example, demo, skpp]`, `--search skilldozer` returns no match while `--search skpp` matches — the inverse of intended. PRD §11 fixes the keyword set to `[example, demo, skilldozer]`. The `description` field (also searched) carries the same `skpp`→`skilldozer` swap, fixing search relevance there too.
- **Satisfies PRD §11 (h2.10)**, which gives the canonical example-skill block byte-for-byte (frontmatter style + whitespace included).
- **Re-syncs the two copies the contract mandates stay identical.** Contract §1: "This is also the source of truth for the compiled-in seed template used by P1.M2.T2.S2 — keep them identical." The constant (`main.go:896 exampleSkillTemplate`) is already §11-correct; only the on-disk asset lagged. This subtask brings the asset up to match (constant untouched).
- **[Mode A docs-with-work]**: this file IS the user-facing example asset and its body doubles as inline documentation — updating it here satisfies the doc-with-work rule. No separate doc subtask.

---

## What

### Success Criteria

- [ ] `skills/example/SKILL.md` contains ZERO occurrences of `skpp` (`grep -c skpp skills/example/SKILL.md` → `0`).
- [ ] The 7 targeted lines match PRD §11 exactly (see Implementation Task 1 table).
- [ ] Frontmatter `name: example` and `category: meta` are UNCHANGED.
- [ ] `metadata.keywords` is `[example, demo, skilldozer]`.
- [ ] `description` is the §11 folded scalar (2 wrapped lines, `skilldozer` wording).
- [ ] The bash try-command block's three lines preserve §11's exact column alignment and the inline backticks around `skilldozer`; the file ends with the closing ``` fence + a single `\n` (no trailing blank line).
- [ ] `./skilldozer check` still prints `OK example (example)` (name unchanged).
- [ ] `./skilldozer --search skilldozer` now lists `example`; `./skilldozer --search skpp` returns no match.
- [ ] `./skilldozer example` still resolves to the `skills/example` directory path.
- [ ] NO other file is modified (main.go, main_test.go, README.md, completions/*, PRD.md all untouched). `go build ./...` and `go test ./...` remain green (the asset is not read by any test).

---

## All Needed Context

### Context Completeness Check

**Pass.** The single edit target (`skills/example/SKILL.md`, 569 bytes, 20 lines) was read in full. The exact target content was derived TWO independent ways and confirmed equal: (a) PRD §11 (h2.10) canonical block, and (b) the rendered text of the compiled-in `exampleSkillTemplate` constant at `main.go:896-928` (which splices 3 backtick runs — verified segment-by-segment). The 7 line-level deltas are enumerated in the drift audit (architecture/docs_and_assets_drift.md §1 table) and re-verified against the live file. The consumers were read: `internal/search/search.go` (keywords matched individually → confirms the search-inversion fix), `internal/discover/skill.go` (frontmatter→Skill mapping; keywords via `toStringSlice`), and the baseline CLI behavior was captured empirically. An implementer who has never seen this repo can complete it in one pass: it is one Markdown file, one canonical target, one set of 7 substitutions.

### Documentation & References

```yaml
# MUST READ — the verified facts (target content, consumers, baseline, scope)
- file: plan/002_38acb6d28a6a/P1M3T1S1/research/verified_facts.md
  why: "§1 = stale-vs-correct map (ONLY the asset is stale; the constant is ALREADY skilldozer).
        §2 = the 7 line-level changes (current→target). §3 = the EXACT target content (rendered
        constant == PRD §11), incl. trailing-newline structure. §4 = why each edit matters to the
        consumers (search flips; check/resolve unaffected). §5 = no test reads the asset (isolated).
        §6 = baseline CLI output. §7 = scope boundary (do NOT touch main.go/main_test.go). §8 =
        relationship to the parallel S3 PRP (S3 seeds from the CONSTANT, already correct)."
  critical: "§1/§7 — main.go's exampleSkillTemplate constant is ALREADY PRD-§11-compliant. The
             contract's conditional ('re-sync the constant if it predates this fix') is FALSE here.
             DO NOT EDIT main.go. Edit ONLY skills/example/SKILL.md. DO NOT edit main_test.go
             (parallel S3 PRP owns it — merge-conflict risk). §3 — transcribe the target verbatim;
             preserve §11's column alignment inside the bash block and the fence+single-newline end."

# MUST READ — the file under edit (the only edit target)
- file: skills/example/SKILL.md
  why: "THE edit target. 20 lines, 569 bytes. Lines 4,5 (description), 7 (keywords), 13 (body),
        18,19,20 (bash try-commands) each contain one `skpp` to swap to `skilldozer`. Lines 2
        (name: example) and 8 (category: meta) UNCHANGED. Line 21 (closing ``` fence) + trailing
        newline UNCHANGED."
  pattern: "Markdown skill file = YAML frontmatter between `---` fences (folded-scalar description
            via `>`; metadata.keywords as a YAML flow sequence `[a, b, c]`), then a `# Heading` body
            with a fenced ```bash block. Frontmatter style MUST match PRD §10/§11 exactly (pi and
            skilldozer both parse it)."
  gotcha: "The description is a YAML FOLDED SCALAR (`>`): the two visible lines are ONE logical
           string; keep the line-wrap exactly as §11 shows (do not reflow to one line, do not add a
           third line). The inline backticks around `skilldozer` on the body line and inside the
           bash block are literal Markdown — preserve them."

# MUST READ — the canonical target (source of truth)
- file: PRD.md
  why: "§11 (h2.10) gives the example-skill block byte-for-byte. §10 (h2.9) documents the
        frontmatter conventions (name/description/metadata.keywords/category). §6.1 (h3.1) documents
        --search's field scope (incl. metadata.keywords) and behavior. §17 (h2.16) guardrails."
  section: "h2.10 (§11 — the example block), h2.9 (§10 — frontmatter), h3.1 (§6.1 --search row)."
  critical: "§11 is the authoritative target. Match it byte-for-byte INCLUDING the column alignment
             of the three bash command lines and the closing fence. Do not 'improve' wording."

# READ-ONLY — the already-correct compiled-in constant (proof the asset must come TO it, not vice-versa)
- file: main.go
  why: "main.go:896 `const exampleSkillTemplate` is the SECOND copy of this exact text, compiled
        into the binary and written verbatim by `skilldozer init` into a fresh store (setupStore
        @main.go:933 → WriteFile @958). It ALREADY says `skilldozer` (grep -rn skpp main.go → 0).
        This subtask makes the on-disk asset EQUAL this constant. Do NOT edit main.go."
  pattern: "The constant splices 3 backtick runs (` + \"`skilldozer`\" + `, ` + \"```bash\" + `,
            ` + \"```\" + `) between raw-string segments because Go raw literals can't hold
            backticks. Its RENDERED text is the target file content (see research/verified_facts.md §3)."

# READ-ONLY — the consumer that flips behavior (why the keyword swap matters)
- file: internal/search/search.go
  why: "matches() does a case-insensitive substring test over 6 fields, testing each metadata.keywords
        entry INDIVIDUALLY (not joined). With the stale keyword `skpp`, --search skilldozer returns
        nothing and --search skpp matches. Swapping the keyword to `skilldozer` inverts this to the
        intended behavior. The description field is ALSO searched, so its skpp→skilldozer swap matters."
  gotcha: "Validation must assert on STDOUT CONTENT (does the `example` row appear?), not exit code:
           the live binary returns exit 0 even on a no-match query (research/verified_facts.md §6)."

# READ-ONLY — the drift audit (the line-level delta table)
- file: plan/002_38acb6d28a6a/architecture/docs_and_assets_drift.md
  why: "§1 = the per-line current→target table for this file (7 rows) + the --search-inversion
        explanation. Confirms this is the ONLY non-core file with literal `skpp` and that name/category
        are already correct. Cross-cutting note #1 flags this as the highest-signal/lowest-effort fix."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer
$ ls skills/example/
SKILL.md     # EDIT: 7 skpp → skilldozer (the ONLY file this subtask touches)
$ grep -c skpp skills/example/SKILL.md
7
# main.go:896 `const exampleSkillTemplate` is ALREADY skilldozer (do NOT touch)
# main_test.go is owned by the parallel S3 PRP (do NOT touch)
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
skills/example/SKILL.md   # REWRITE body/frontmatter: skpp → skilldozer (match PRD §11 == rendered constant)
```

**No new files. No Go files touched.** One file edited.

| File | Responsibility |
|---|---|
| `skills/example/SKILL.md` | The single shipped example skill (PRD §11). Its frontmatter drives `--list`/`--search`; its body is inline user-facing documentation and the template new skills copy. After this edit it is byte-for-byte equal to PRD §11 and to the compiled-in `exampleSkillTemplate` seed constant. |

### Known Gotchas of our codebase & Library Quirks

```markdown
<!-- GOTCHA #1 (CRITICAL — scope) — Edit ONLY skills/example/SKILL.md. main.go's exampleSkillTemplate
     constant is ALREADY PRD §11-compliant (P1.M2.T2.S2 wrote it with `skilldozer`; grep -rn skpp
     main.go → 0). The contract's conditional "re-sync the constant if it predates this fix" is FALSE
     here. Touching main.go would risk changing a correct constant to match a transcribed (possibly
     mis-aligned) copy. LEAVE main.go ALONE. -->

<!-- GOTCHA #2 (CRITICAL — do NOT touch main_test.go) — The parallel P1.M2.T2.S3 PRP edits
     main_test.go (adds run-level init tests). Editing it here too risks a merge conflict. This
     subtask adds NO test. (The "keep identical" invariant is enforced by the grep gates + the
     init-seed diff in Level 3, not by a committed regression test — adding one is out of scope.) -->

<!-- GOTCHA #3 — The description is a YAML FOLDED SCALAR (`>`). The two indented lines are ONE
     logical string. Keep the wrap EXACTLY as §11 shows: line 4 ends "...frontmatter and", line 5 is
     "  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills."
     Do NOT collapse to one line, do NOT add a third wrap line, do NOT change the 2-space indent. -->

<!-- GOTCHA #4 — metadata.keywords is a YAML FLOW SEQUENCE `[example, demo, skilldozer]` (inline,
     comma+space separated). Keep the flow style; do NOT switch to block style (- example\n- demo...).
     Only the LAST element changes: skpp → skilldozer. -->

<!-- GOTCHA #5 — Preserve the exact column alignment inside the ```bash block. §11 aligns the `#`
     comment columns across all three command lines. The `skilldozer` token is 6 chars longer than
     `skpp`, so the ORIGINAL alignment is BROKEN if you only swap the word — you must reproduce §11's
     (and the constant's) post-rename alignment verbatim. Transcribe the three lines exactly:
       skilldozer example                       # prints this directory's absolute path
       skilldozer -f example                    # prints .../skills/example/SKILL.md
       pi --skill "$(skilldozer example)"       # loads this skill into pi
     (Copy these from research/verified_facts.md §3, which was rendered from the constant.) -->

<!-- GOTCHA #6 — Inline backticks are LITERAL Markdown, not delimiters to remove. The body line has
     `skilldozer` in backticks; the bash block has the pi --skill "$(skilldozer example)" line.
     Keep all backticks. The closing ``` fence (line 21) is the fence terminator — keep it + the
     single trailing newline (file ends `fence\n`, NO blank line after). Current file's last 4 bytes
     are 60 60 60 0a (```\n); preserve that. -->

<!-- GOTCHA #7 — `./skilldozer check` and `./skilldozer example` are UNAFFECTED by this edit:
     check reports OK by the `name: example` frontmatter field (unchanged) and validates structure
     (unchanged); `example` resolves by the DIRECTORY name `example` (unchanged). Only `--search`
     behavior changes (keyword + description swap). So "check still OK / example still resolves"
     are regression-assertions, not new behavior. -->

<!-- GOTCHA #8 — Validation must assert on STDOUT CONTENT for --search, NOT exit code. The live
     binary returns exit 0 for `--search <nomatch>` ("no skills matched …") — PRD §6.1 says exit 1
     on no matches, but that discrepancy is a SEPARATE concern, out of scope here. Assert that the
     `example` row IS/IS-NOT present in stdout (see Level 3 commands). -->
```

---

## Implementation Blueprint

### Data models and structure

None. This is a Markdown asset edit — no Go types, no schemas, no migrations. The only "model" is the YAML frontmatter of the skill, whose shape (name/description/metadata.keywords/metadata.category) is unchanged; only token values change (`skpp`→`skilldozer`).

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: REWRITE skills/example/SKILL.md  (the ONLY task; single file)
  - FILE: skills/example/SKILL.md (overwrite the whole file with the PRD §11 block)
  - TARGET CONTENT (byte-for-byte == PRD §11 == rendered exampleSkillTemplate constant):
      ---                                       # frontmatter open (unchanged)
      name: example                             # UNCHANGED (line 2)
      description: >                            # UNCHANGED folded-scalar marker (line 3)
        Reference example skill for skilldozer. Demonstrates the required frontmatter and
        how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.
      metadata:                                 # UNCHANGED (line 6)
        keywords: [example, demo, skilldozer]   # CHANGED: last elem skpp -> skilldozer (line 7)
        category: meta                          # UNCHANGED (line 8)
      ---                                       # frontmatter close (unchanged)
      (blank line)
      # Example Skill                           # UNCHANGED heading
      (blank line)
      This skill exists only so `skilldozer` has something to resolve.   # CHANGED: skpp -> skilldozer
      (blank line)
      Try:                                      # UNCHANGED
      (blank line)
      ```bash                                   # UNCHANGED fence open
      skilldozer example                       # prints this directory's absolute path   # CHANGED
      skilldozer -f example                    # prints .../skills/example/SKILL.md      # CHANGED
      pi --skill "$(skilldozer example)"       # loads this skill into pi                 # CHANGED
      ```                                       # fence close (unchanged) + single trailing \n
  - SUBSTITUTE table (7 rows — current → target), transcribe the TARGET column verbatim:
      | Line | Current                                              | Target                                                         |
      | 4    | `  Reference example skill for skpp. Demonstrates the required frontmatter and` | `  Reference example skill for skilldozer. Demonstrates the required frontmatter and` |
      | 5    | `  how skpp resolves a tag to an absolute path. Safe to delete once you add real skills.` | `  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.` |
      | 7    | `  keywords: [example, demo, skpp]`                 | `  keywords: [example, demo, skilldozer]`                      |
      | 13   | `` This skill exists only so `skpp` has something to resolve. `` | `` This skill exists only so `skilldozer` has something to resolve. `` |
      | 18   | `` skpp example                       # prints this directory's absolute path `` | `` skilldozer example                       # prints this directory's absolute path `` |
      | 19   | `` skpp -f example                    # prints .../skills/example/SKILL.md `` | `` skilldozer -f example                    # prints .../skills/example/SKILL.md `` |
      | 20   | `` pi --skill "$(skpp example)"       # loads this skill into pi `` | `` pi --skill "$(skilldozer example)"       # loads this skill into pi `` |
  - PRESERVE: line 2 `name: example`; line 8 `category: meta`; the closing ``` fence; the single
    trailing newline (file ends `fence\n`, no blank line after). Keep the exact column alignment of
    the three bash command lines (GOTCHA #5). Keep the YAML folded-scalar wrap (GOTCHA #3) and the
    flow-sequence keyword style (GOTCHA #4).
  - NAMING/PLACEMENT: no new files; overwrite in place at skills/example/SKILL.md.
  - DO NOT EDIT: main.go (constant already correct — GOTCHA #1), main_test.go (parallel S3 owns it
    — GOTCHA #2), README.md, completions/*, PRD.md.

Task 2: VERIFY (isolated — run after Task 1)
  - grep -c skpp skills/example/SKILL.md              # MUST print: 0
  - grep -q 'Reference example skill for skilldozer' skills/example/SKILL.md   # exit 0
  - grep -q 'keywords: \[example, demo, skilldozer\]' skills/example/SKILL.md  # exit 0
  - grep -q '`skilldozer` has something to resolve' skills/example/SKILL.md    # exit 0
  - See Validation Loop Level 1-3 for the full CLI gate set.
```

### Implementation Patterns & Key Details

```markdown
<!-- The whole deliverable is one file overwrite. The target is fixed and external (PRD §11 ==
     rendered exampleSkillTemplate). Lowest-risk approach: read the exact target block from
     research/verified_facts.md §3 and write it to skills/example/SKILL.md verbatim, then run
     the grep + CLI gates. Do NOT hand-align the bash columns from the stale file (its columns
     were built for the 3-char `skpp` and are wrong for the 9-char `skilldozer`); use §11's
     alignment, which is already correct for `skilldozer`. -->

# Equivalence check (optional, proves byte-equality to the compiled constant):
# If the parallel P1.M2.T2.S3 `init` dispatch has landed, seed a temp store and diff:
#   iso=$(mktemp -d)
#   SKILLDOZER_SKILLS_DIR="" SKILLDOZER_CONFIG="$iso/cfg.yaml" \
#     ./skilldozer init --store "$iso/store" >/dev/null 2>&1
#   diff -u skills/example/SKILL.md "$iso/store/example/SKILL.md" && echo "asset == seed constant"
#   rm -rf "$iso"
# (If `init` is still a no-op — S3 not yet landed — skip this; the grep + --search gates suffice.)
```

### Integration Points

```yaml
ASSET:
  - file: "skills/example/SKILL.md"
    change: "7 lines: skpp -> skilldozer (description x2, metadata.keywords last elem x1, body x1, 3 bash lines)"
    consumers:
      - "skilldozer --search (internal/search): keyword + description become searchable as 'skilldozer'"
      - "skilldozer check (internal/check): validates frontmatter; name 'example' unchanged -> still OK"
      - "skilldozer example (internal/resolve): resolves by dir name 'example' (unchanged)"

NO CODE CHANGES:
  - main.go: untouched (exampleSkillTemplate constant ALREADY skilldozer — research §1)
  - main_test.go: untouched (parallel S3 PRP owns it — GOTCHA #2)
  - internal/*: untouched (consumed, not modified)
  - go.mod/go.sum: untouched (no Go change)
  - No migrations, no routes, no config, no env vars, no new files.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after editing the file)

```bash
cd /home/dustin/projects/skilldozer

# The file must contain ZERO skpp and the key §11 tokens:
grep -c skpp skills/example/SKILL.md                                   # MUST print: 0
grep -q 'Reference example skill for skilldozer' skills/example/SKILL.md && echo "desc OK"
grep -q 'keywords: \[example, demo, skilldozer\]' skills/example/SKILL.md && echo "keywords OK"
grep -q '`skilldozer` has something to resolve' skills/example/SKILL.md && echo "body OK"
grep -q 'skilldozer example ' skills/example/SKILL.md && echo "try-cmd[0] OK"
grep -q 'skilldozer -f example' skills/example/SKILL.md && echo "try-cmd[1] OK"
grep -q 'pi --skill "$(skilldozer example)"' skills/example/SKILL.md && echo "try-cmd[2] OK"
# name/category preserved:
grep -q '^name: example$' skills/example/SKILL.md && echo "name OK"
grep -q '^  category: meta$' skills/example/SKILL.md && echo "category OK"
# Expected: every line prints its OK; grep -c prints 0. If any fails, READ the file and fix.

# Sanity: the Go module still builds/tests green (no test reads this asset, so unaffected):
go build ./...  && echo "build OK"
go vet ./...    && echo "vet OK"
# Expected: build OK, vet OK (zero changes to .go files).
```

### Level 2: Unit Tests (component validation)

```bash
cd /home/dustin/projects/skilldozer

# No test reads skills/example/SKILL.md (research §5), so the suite is a REGRESSION check that
# nothing was accidentally broken (e.g. a stray edit to a .go file). All must stay green:
go test ./...
# Expected: PASS, exit 0. (The synthetic fixtures in skill_test.go:171 / ui_test.go:128 already
# use the skilldozer wording and are independent of this repo asset — they were passing before.)
```

### Level 3: Integration Testing (the authoritative CLI gate set from the contract)

```bash
cd /home/dustin/projects/skilldozer
export SKILLDOZER_SKILLS_DIR="$PWD/skills"   # point the binary at the repo's skills dir

# (a) check still reports OK (name 'example' unchanged):
./skilldozer check
# Expected: a line containing "OK" + "example" + "(example)"; "1 skills, 0 errors, 0 warnings".

# (b) --search skilldozer now MATCHES (was: "no skills matched" before the fix):
./skilldozer --search skilldozer | grep -E '^(TAG|example)' 
# Expected: the table header (TAG/NAME/DESCRIPTION) AND an "example" row appear.
# (GOTCHA #8: assert on the row's PRESENCE, not exit code — live binary exits 0 on no-match too.)

# (c) --search skpp NO LONGER matches (was: matched before the fix):
out=$(./skilldozer --search skpp); echo "$out" | grep -q '^example' \
  && echo "FAIL: skpp still matches" || echo "OK: skpp no longer matches"
# Expected: "OK: skpp no longer matches".

# (d) the example tag still resolves (dir name 'example' unchanged):
./skilldozer example
# Expected: prints the absolute path .../skills/example , exit 0.

# (e) (optional) byte-equality to the compiled seed constant — only if the parallel S3 `init`
# dispatch has landed. Seeds a temp store from the constant and diffs the repo asset against it:
iso=$(mktemp -d)
if SKILLDOZER_SKILLS_DIR="" SKILLDOZER_CONFIG="$iso/cfg.yaml" \
     ./skilldozer init --store "$iso/store" >/dev/null 2>&1 && [ -f "$iso/store/example/SKILL.md" ]; then
  diff -u skills/example/SKILL.md "$iso/store/example/SKILL.md" && echo "asset == seed constant (PRD §11)"
else
  echo "(init not yet live / S3 pending — skipping seed-diff; grep + --search gates above are authoritative)"
fi
rm -rf "$iso"
# Expected (if init live): no diff output + "asset == seed constant (PRD §11)".
```

### Level 4: Creative & Domain-Specific Validation

```bash
cd /home/dustin/projects/skilldozer
export SKILLDOZER_SKILLS_DIR="$PWD/skills"

# Cross-field search parity: the description also carries 'skilldozer' now, so searching any
# distinctive description substring should match (proves the description swap, not just keywords):
./skilldozer --search "resolves a tag to an absolute path" | grep -q '^example' \
  && echo "description-substring search OK" || echo "FAIL: description not searchable"

# Negative: a token that ONLY appeared as the stale keyword must now be absent from search:
./skilldozer --search demo | grep -q '^example' && echo "keyword 'demo' still matches OK"

# Confirm no other non-plan file in the repo still says skpp (regression guard for the rename):
grep -rIl "skpp" --include="*.md" --include="*.bash" --include="*.fish" --include="*.sh" . \
  | grep -v '/plan/' | grep -v '/.git/' | grep -v '/.pi-subagents/' || echo "no skpp residue outside plan/ archive"
# Expected: "no skpp residue outside plan/ archive" (PRD §19 #15: plan/ archive is intentionally historical).
```

---

## Final Validation Checklist

### Technical Validation

- [ ] Level 1 grep gates all pass; `grep -c skpp skills/example/SKILL.md` → `0`
- [ ] `go build ./...` green (no .go file changed)
- [ ] `go vet ./...` clean
- [ ] `go test ./...` green (regression — no test reads the asset)

### Feature Validation

- [ ] `./skilldozer check` prints `OK example (example)` (name unchanged)
- [ ] `./skilldozer --search skilldozer` lists the `example` row (was: no match)
- [ ] `./skilldozer --search skpp` returns no `example` row (was: matched)
- [ ] `./skilldozer example` still resolves to the `skills/example` directory
- [ ] (optional) asset byte-equals the seeded constant via `init` (if S3 landed)

### Code Quality Validation

- [ ] Only `skills/example/SKILL.md` modified — `git status` shows exactly one changed file
- [ ] `main.go` untouched (`git diff --quiet main.go` → clean; constant already correct)
- [ ] `main_test.go` untouched (parallel S3 PRP owns it)
- [ ] PRD §11 column alignment inside the ```bash block preserved (not the stale `skpp`-era columns)
- [ ] YAML folded-scalar description wrap and flow-sequence keywords style preserved
- [ ] File ends with the closing ``` fence + single `\n` (no trailing blank line)

### Documentation & Deployment

- [ ] [Mode A] The file IS the user-facing example asset + inline doc — updated in this subtask (no separate doc task)
- [ ] No new env vars, no README change (README = P1.M4.T2.S1), no completions change (P1.M3.T2.S1)

---

## Anti-Patterns to Avoid

- ❌ Don't edit `main.go` to "re-sync" the `exampleSkillTemplate` constant — it is ALREADY PRD §11-compliant (grep -rn skpp main.go → 0). The contract's re-sync conditional is FALSE here. Editing it risks introducing a mis-aligned copy of a correct constant.
- ❌ Don't edit `main_test.go` to add a regression test — the parallel P1.M2.T2.S3 PRP owns that file (merge-conflict risk). The "keep identical" invariant is enforced by the grep gates + the optional init-seed diff, not a committed test.
- ❌ Don't hand-align the bash command columns from the STALE file — its columns were sized for the 3-char `skpp` and are wrong for `skilldozer`. Transcribe §11's (and the constant's) post-rename alignment verbatim.
- ❌ Don't reflow the `description` folded scalar to one line or add a third wrap line — keep §11's exact 2-line wrap.
- ❌ Don't switch `metadata.keywords` from flow style `[a, b, c]` to block style (`- a` / `- b`) — only the last element's value changes.
- ❌ Don't drop the inline backticks around `skilldozer` (body line + the pi `--skill` line) or the closing ``` fence — they are literal Markdown.
- ❌ Don't assert `--search` validation on exit codes — assert on the presence/absence of the `example` row in stdout (the live binary exits 0 on no-match; GOTCHA #8).
- ❌ Don't touch `plan/` archive `skpp` hits — PRD §19 #15 declares them intentionally historical.
- ❌ Don't "improve" the wording, add a license field, or deviate from §11 in any way — match it byte-for-byte.
