# PRP — P1.M2.T1.S1: Extend `--search` to match aliases + category

> **Subtask:** P1.M2.T1.S1 — fixes QA **Issue 4** (Minor): `--search` matches
> `metadata.keywords` but not `metadata.aliases` / `metadata.category`, even
> though PRD §10 says all three "exist only to enrich `skpp --search`".
> **Decision:** `decisions.md §D4` — **§10 wins over §6.1**; extend `matches()`.
>
> **Scope boundary:** A CONSUMER-ONLY change. The `Skill.Aliases` / `Skill.Category`
> fields already exist (`skill.go:42-43`) and are already populated by `BuildSkill`
> (`skill.go:113-114`). This subtask only makes `search.matches()` consult them.
> **No struct change, no discover change, no resolve change, no Index() change.**
> Three files touched: `internal/search/search.go` (logic + docs),
> `internal/search/search_test.go` (invert one test), `main.go` (usageText + 2
> code-comment consistency updates). Mode A — README deferred to P1.M5.T3.
>
> **PARALLEL CONTEXT:** P1.M1.T1.S1 (Issue 1, `--path` source label) edits
> `main.go`'s `c.path` branch (~268-281) + `main_test.go`. This subtask edits
> `main.go` usageText (lines 71, 78), a config comment (131), and a dispatch
> comment (317-319) — **zero overlap** with the `c.path` branch, and a different
> test file (`search_test.go` vs `main_test.go`). Both apply cleanly regardless of
> landing order. P1.M1.T1.S1 does NOT touch usageText or the search dispatch.

---

## Goal

**Feature Goal**: Make `skpp --search <q>` match skills whose `metadata.aliases`
(any element) or `metadata.category` contain `<q>` as a case-insensitive
substring — closing the gap between §10 ("aliases/category/keywords exist only to
enrich `skpp --search`") and the current implementation (which scans only tag /
name / description / keywords). After the fix, `skpp --search <alias>` and
`skpp --search <category-substring>` return matching skills, consistent with the
fact that `skpp <alias>` (resolution by alias, §7.2 step 4) already works.

**Deliverable**: Surgical edits to 3 files:
1. `internal/search/search.go` — add an Aliases `for` loop + a Category check to
   `matches()` (after the Keywords loop, before `return false`); rewrite the
   `matches()` doc comment + the package/Search doc comments that say "four
   fields" / "deliberately does NOT include Category or Aliases".
2. `internal/search/search_test.go` — REPLACE `TestSearchDoesNotMatchCategoryOrAliases`
   with `TestSearchMatchesCategoryAndAliases` (asserts they DO match).
3. `main.go` — update the `--search` description in `usageText` (OPTIONS table +
   EXAMPLE comment) and 2 stale code comments to list aliases + category.

**Success Definition**: `go test ./internal/search/ -v` passes with the renamed
test asserting alias/category matches; the unchanged
`TestSearchKeywordSubstringNotJoinBoundary` still passes (keyword logic
untouched); `go test ./...` whole module green; `gofmt`/`go vet` clean; the bug
report's reproduce snippet (`SKPP_SKILLS_DIR=... ./skpp --search my-alias`) now
exits 0 with the skill listed (was exit 1 "no skills matched"). Only the 3 files
above changed (`git diff --name-only`).

---

## Why

- **Spec consistency (§10 vs §6.1).** §10 explicitly states keywords/category/
  aliases "exist only to enrich `skpp --search`". The current code honored only
  keywords, following §6.1's summary field list. `decisions.md §D4` resolved the
  tension in favor of §10's intentional prose. This subtask implements that
  decision.
- **User expectation / UX.** A user who reads §10, adds `metadata.aliases`, and
  runs `skpp --search <alias>` gets "no skills matched" — even though `skpp
  <alias>` (resolution) works. The asymmetry is surprising. Search and resolution
  should agree on what an alias is.
- **Zero data-model cost.** `Aliases` and `Category` are already on `Skill` and
  already populated. The fix is ~8 lines of matching logic + doc/test alignment.
  No struct, no discover, no resolve, no Index() change. Lowest-risk way to
  remove the inconsistency.
- **No new dependency, no behavior change to existing matches.** Adding checks
  can only INCREASE the match set (a previously-non-matching skill may now match
  via alias/category); nothing that matched before stops matching. The exit-1
  "no matches" path is unchanged for queries that match nothing.

---

## What

`search.matches()` gains two checks (Aliases loop + Category scalar), mirroring
the existing Keywords pattern. Doc comments and the `usageText` `--search` line
are updated so the code never contradicts its docs.

### Behavior change (search match set GROWS)

| Query | Before | After |
|---|---|---|
| substring of an alias (e.g. `my-alias`) | no match (exit 1) | **match** |
| substring of category (e.g. `writ` vs `writing`) | no match (exit 1) | **match** |
| substring of tag/name/description/keyword | match (unchanged) | match (unchanged) |
| query matching nothing | no match (exit 1, unchanged) | no match (exit 1, unchanged) |

Empty query still matches all (`Contains(hay, "")` is always true); a
no-frontmatter skill (`Aliases==nil`, `Category==""`) still matches only by tag.

### Success Criteria

- [ ] `search.go` `matches()` has an Aliases `for` loop AND a Category `if` after the Keywords loop, before `return false`
- [ ] The Aliases loop matches each alias INDIVIDUALLY (boundary-safe, like Keywords) — NOT `strings.Join`'d
- [ ] The Category check is a single `strings.Contains(strings.ToLower(s.Category), q)`
- [ ] `search.go` `matches()` doc comment rewritten: no longer says "deliberately does NOT include Category or Aliases"; states they ARE included per §10
- [ ] `search.go` package doc + `Search()` doc updated to say six fields (no lingering "four fields")
- [ ] `search_test.go` `TestSearchDoesNotMatchCategoryOrAliases` REPLACED by `TestSearchMatchesCategoryAndAliases` asserting `Search("secret-alias", withAliases)` and `Search("secret-cat", withCategory)` each return 1 skill
- [ ] `TestSearchKeywordSubstringNotJoinBoundary` is UNCHANGED and still passes
- [ ] `main.go` `usageText` OPTIONS line for `--search` reads `... / keywords / aliases / category`
- [ ] `main.go` `usageText` EXAMPLE comment for `--search` reads `.../keywords/aliases/category`
- [ ] `main.go` stale code comments (config field line 131, dispatch comment line 317) updated to list aliases/category
- [ ] `go test ./internal/search/ -v` passes; `go test ./...` green; `gofmt -l`/`go vet ./...` clean
- [ ] `git diff --name-only` == exactly `internal/search/search.go`, `internal/search/search_test.go`, `main.go`

---

## All Needed Context

### Context Completeness Check

_Pass: the exact current code of `matches()`, the exact target code block, the
exact old+new text for every stale doc reference (verified by reading the working
tree), and the exact test-to-invert are all in the Implementation Blueprint. The
proposed `matches()` logic was EXECUTED in a throwaway Go 1.25 module against 8
fixtures — all PASS (research §3). An implementer who knows Go but nothing about
this repo can apply the edits verbatim._

### Documentation & References

```yaml
# MUST READ — this subtask's verified facts (every load-bearing decision)
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/P1M2T1S1/research/verified_facts.md
  why: "§1 current matches() (4 fields). §2 Aliases/Category already on Skill +
        populated by BuildSkill (no struct/discover change needed). §3 the
        proposed block EXECUTED against 8 fixtures (all PASS). §4 range over nil
        Aliases is safe (no panic, no guard needed). §5 aliases matched
        INDIVIDUALLY (boundary-safe, mirrors Keywords). §6 the exact test to
        invert + the contract for its replacement. §7 COMPLETE inventory of all
        7 stale 'four fields' doc refs (search.go x3, main.go x4) with exact
        current text. §8 usageText column alignment is NOT broken by the longer
        line. §9 scope boundaries (skill.go/resolve.go/index.go/README NOT
        touched). §10 validation gates."
  critical: "Do NOT add a nil guard before the Aliases range (range over nil is
             safe — verified §4). Do NOT Join the aliases (boundary safety — §5).
             Do NOT touch skill.go/resolve.go (data already exists — §2/§9)."

# CONTRACT — the decision this implements
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/decisions.md
  why: "§D4: 'Extend search to match Aliases AND Category (§10 wins over §6.1)'.
        Rationale: §10's prose is the more specific, intentional claim; a user
        who adds aliases (which already drive resolution) reasonably expects
        search to find them."
  section: "D4"

# CONTRACT — the issue root cause + exact fix snippet + test impact
- file: plan/001_fcde63e5bb60/bugfix/001_7fdd8ce7410b/architecture/issue_analysis.md
  why: "Issue 4: root cause (matches scans 4 fields only), the exact two-check
        fix snippet, and the CRITICAL test-impact note that
        TestSearchDoesNotMatchCategoryOrAliases asserts the OPPOSITE and must be
        replaced."
  section: "Issue 4 (MINOR)"

# THE FILE BEING EDITED — current matches() + the 3 stale doc refs in it
- file: internal/search/search.go
  why: "matches() at :59-77 (add Aliases loop + Category check before return
        false). Doc comments at :1-3 (package), :20-22 (Search), :49-56 (matches)
        all say 'four fields' / 'deliberately does NOT include Category or
        Aliases' and must be rewritten."
  pattern: "Mirror the existing Keywords loop EXACTLY for Aliases (range +
            Contains(ToLower(...), q) -> return true); Category is a single
            scalar Contains like Name/Description."
  gotcha: "range over a nil s.Aliases is a safe no-op (verified) — do NOT add a
           guard; it would diverge from the Keywords loop's style for no benefit."

# THE DATA SOURCE — proves Aliases/Category already exist + are populated
- file: internal/discover/skill.go
  why: "Skill struct :42-43 (Category string, Aliases []string). BuildSkill :113-114
        populates them from Frontmatter.Metadata via toStringSlice / comma-ok. So
        the data is ALREADY present at search time — this subtask changes NOTHING
        here (READ-ONLY reference)."
  pattern: "Aliases is []string (already normalized from yaml.v3's []any by
            toStringSlice); Category is string. Callers test with len(), not nil."

# THE TEST FILE — the one test to invert
- file: internal/search/search_test.go
  why: "TestSearchDoesNotMatchCategoryOrAliases :126 asserts len(out)!=0 is a
        FAILURE — it encodes the WRONG behavior and will FAIL against the new
        code. Replace with TestSearchMatchesCategoryAndAliases. The helper sk()
        does NOT set Aliases/Category, so the new test uses literal Skill{} values
        (like the old test did). TestSearchKeywordSubstringNotJoinBoundary is
        UNCHANGED."
  pattern: "White-box package search; literal Skill{} values for the alias/category
            fixtures; plain t.Errorf; no testify; no t.Parallel()."

# USAGE TEXT — the --search description to extend
- file: main.go
  why: "usageText const :50-90: line 71 (EXAMPLE '# substring search over
        tag/name/description/keywords') and line 78 (OPTIONS 'Substring search
        over tag / name / description / keywords') — both within usageText, both
        describe the search field set, both must append aliases/category. Plus 2
        stale code comments: :131 (config field) and :317-319 (dispatch comment)."
  critical: "Line 78 is in the OPTIONS table but the description is the LAST
             column (variable-length, no wrap) — appending ' / aliases / category'
             does NOT misalign other rows (verified research §8)."

# REFERENCE — the matching primitive used throughout
- url: https://pkg.go.dev/strings#Contains
  why: "strings.Contains(haystack, needle) — case-insensitive via ToLower(hay)
        and the already-lowercased query q. This is the EXACT primitive the
        existing RelTag/Name/Description/Keywords checks use; the new Aliases/
        Category checks reuse it verbatim."
- url: https://go.dev/ref/spec#For_statements
  why: "'for _, a := range s.Aliases' over a nil slice executes zero iterations
        (no panic). This is why no nil guard is needed — same as the existing
        Keywords loop."
```

### Current Codebase tree (relevant slice)

```bash
$ cd /home/dustin/projects/skpp && ls internal/search/ internal/discover/skill.go main.go
internal/search/
├── search.go          # matches() :59-77 (4 fields) — EDIT (add Aliases+Category + docs)
└── search_test.go     # TestSearchDoesNotMatchCategoryOrAliases :126 — REPLACE
internal/discover/skill.go   # Skill.Aliases :43, Skill.Category :42 — READ-ONLY (already populated)
main.go                      # usageText :50-90 (lines 71, 78), comments :131, :317-319 — EDIT
```

### Desired Codebase tree (files touched)

```bash
skpp/
├── internal/search/
│   ├── search.go          # MODIFY — matches() + 3 doc comments
│   └── search_test.go     # MODIFY — invert 1 test (rename + flip assertion)
└── main.go                # MODIFY — usageText (2 lines) + 2 code comments
# (internal/discover/*, internal/resolve/*, internal/ui/*, README.md — UNCHANGED)
```

| File | Change | Lines (approx) |
|---|---|---|
| `internal/search/search.go` | add Aliases loop + Category check; rewrite matches()/Search()/package docs | body 69-76; docs 1-3, 20-22, 49-56 |
| `internal/search/search_test.go` | replace `TestSearchDoesNotMatchCategoryOrAliases` → `TestSearchMatchesCategoryAndAliases` | 126-134 |
| `main.go` | extend `--search` text in usageText (EXAMPLE + OPTIONS); update 2 stale comments | 71, 78, 131, 317-319 |

### Known Gotchas of our codebase & the matching logic

```go
// GOTCHA #1 — range over a nil s.Aliases is SAFE (no panic, no guard needed).
// Go's `for _, a := range nilSlice` iterates zero times. Verified (research §4).
// The existing Keywords loop already relies on this (a no-frontmatter skill has
// Keywords==nil). Do NOT add `if s.Aliases != nil` — it would diverge from the
// Keywords style for zero benefit and imply a falsehood.
//   RIGHT: for _, a := range s.Aliases { ... }
//   WRONG: if s.Aliases != nil { for _, a := range s.Aliases { ... } }

// GOTCHA #2 — Aliases MUST be matched INDIVIDUALLY, not strings.Join'd. A query
// spanning a boundary between two aliases (e.g. "aliasbar" across
// ["foo-alias","bar-alias"]) must NOT match. Same rationale as the existing
// Keywords loop (TestSearchKeywordSubstringNotJoinBoundary). Verified §5.
//   RIGHT: for _, a := range s.Aliases { if strings.Contains(strings.ToLower(a), q) { return true } }
//   WRONG: if strings.Contains(strings.ToLower(strings.Join(s.Aliases, " ")), q) { return true }

// GOTCHA #3 — the match set only GROWS; it never shrinks. Adding checks cannot
// make a previously-matching skill stop matching. So no existing positive test
// breaks; only the ONE test that asserted aliases/category do NOT match
// (TestSearchDoesNotMatchCategoryOrAliases) breaks and must be inverted. Do not
// "fix" other passing tests.

// GOTCHA #4 — the test to invert uses query "secret" (a substring of BOTH
// "secret-cat" and "secret-alias") against a SINGLE skill holding both fields.
// The REPLACEMENT should split into TWO skills (one aliases-only, one
// category-only) and query the FULL value each, so each path is asserted
// independently. This matches the item's skillWithAliases/skillWithCategory
// phrasing and is clearer than the old single-skill form.

// GOTCHA #5 — usageText line 78 is in the OPTIONS table, but the description is
// the LAST column with variable-length rows (e.g. the `check` row is longer).
// Appending " / aliases / category" extends only this row's trailing text; it
// does NOT shift any other column. Verified §8. Do NOT re-align other rows.

// GOTCHA #6 — the sk() test helper does NOT set Aliases/Category, so the new
// test must use literal discover.Skill{} values (as the old
// TestSearchDoesNotMatchCategoryOrAliases already did). Do not try to extend sk().

// GOTCHA #7 — there are SEVEN stale "four fields"/"tag/name/description/keywords"
// references total (research §7): 3 in search.go (package doc, Search() doc,
// matches() doc) and 4 in main.go (usageText EXAMPLE :71, usageText OPTIONS :78,
// config comment :131, dispatch comment :317). Update ALL of them or the docs
// contradict the code. The item explicitly names the matches() doc and the
// usageText OPTIONS line; the other 5 are required for consistency.

// GOTCHA #8 — README.md is OUT OF SCOPE (deferred to P1.M5.T3.S1 Mode B doc
// sync). Do NOT touch it. This subtask is Mode A: code + in-source docs + tests.

// GOTCHA #9 — this subtask runs IN PARALLEL with P1.M1.T1.S1, which edits
// main.go's c.path branch (~268-281) and main_test.go. Your main.go edits
// (usageText 50-90, comment 131, comment 317-319) do NOT overlap the c.path
// branch, and you edit search_test.go (not main_test.go). No merge conflict.

// GOTCHA #10 — do NOT change resolve.go. Alias RESOLUTION (skpp <alias>,
// §7.2 step 4) already works; this fix makes SEARCH consistent with it. The two
// packages (resolve, search) stay isolated.
```

---

## Implementation Blueprint

### Data models and structure

**No new data models.** This subtask reuses the existing `discover.Skill` fields
verbatim — `Aliases []string` (`skill.go:43`) and `Category string`
(`skill.go:42`), both already populated by `BuildSkill` (`skill.go:113-114`).

```go
// internal/discover/skill.go:42-43 (EXISTING — do not modify)
type Skill struct {
	...
	Category    string   // metadata.category ("" if absent)
	Aliases     []string // metadata.aliases (nil if absent/non-list)
	...
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT internal/search/search.go — add Aliases loop + Category check
  - FILE: internal/search/search.go
  - LOCATE: func matches() body; the Keywords `for` loop (ends ~line 73) right
            before `return false` (~line 74).
  - EDIT (exact oldText → newText): see "File 1 edit A" below. Insert the Aliases
            `for` loop and the Category `if` BETWEEN the Keywords loop's closing
            `}` and the `return false`.
  - PATTERN: mirror the existing Keywords loop EXACTLY for Aliases
            (range + Contains(ToLower(a), q) -> return true); Category is a single
            scalar Contains like the Name/Description checks.
  - GOTCHA: no nil guard on s.Aliases (range over nil is safe — §4). Do NOT Join.

Task 2: EDIT internal/search/search.go — rewrite the matches() doc comment
  - FILE: internal/search/search.go
  - LOCATE: the matches() doc comment block (~lines 49-56) that says "Field scope
            is EXACTLY the four PRD §6.1 fields. It deliberately does NOT include
            Category or Aliases ...".
  - EDIT: see "File 1 edit B". Rewrite to state six fields ARE scanned, citing
            PRD §10 (keywords/category/aliases "exist only to enrich skpp
            --search") and the consistency with resolve (§7.2 step 4 resolves by
            alias). The item EXPLICITLY requires this rewrite.

Task 3: EDIT internal/search/search.go — update package + Search() doc comments
  - FILE: internal/search/search.go
  - LOCATE: package doc (~lines 1-3, "over the four fields PRD §6.1 names") and
            Search() doc (~lines 20-22, "ANY of the four PRD §6.1 fields").
  - EDIT: see "File 1 edits C and D". Change "four fields" -> six fields and list
            aliases + category. (Consistency — leaving these stale contradicts
            the code.)

Task 4: EDIT internal/search/search_test.go — INVERT the test
  - FILE: internal/search/search_test.go
  - LOCATE: func TestSearchDoesNotMatchCategoryOrAliases (~line 126).
  - EDIT: see "File 2 edit". REPLACE the entire function with
            TestSearchMatchesCategoryAndAliases: two skills (one aliases-only,
            one category-only), assert Search("secret-alias",...) and
            Search("secret-cat",...) each return 1 skill.
  - DO NOT TOUCH: TestSearchKeywordSubstringNotJoinBoundary (unchanged; aliases
            get the same boundary semantics automatically).

Task 5: EDIT main.go — extend usageText --search description (2 lines)
  - FILE: main.go
  - LOCATE: const usageText; line 71 (EXAMPLE comment) and line 78 (OPTIONS table).
  - EDIT: see "File 3 edits A and B". Append aliases + category to both lines.
            The item EXPLICITLY requires the OPTIONS line (78).
  - GOTCHA: line 78's description is the last column — extending it does not
            misalign other rows (§8).

Task 6: EDIT main.go — update 2 stale code comments (consistency)
  - FILE: main.go
  - LOCATE: line 131 (config struct searchMode field comment) and lines 317-319
            (--search dispatch comment). Both say "tag/name/description/keywords"
            / "tag, frontmatter name, description, or any metadata keyword".
  - EDIT: see "File 3 edits C and D". Append aliases + category; change the
            (§6.1) reference on line 131 to (§10) per decisions.md §D4.
  - DO NOT TOUCH: main.go:149 (about flag VALUE arity, not fields), :481/:485
            (exclusivity). The c.path branch (~268-281) — owned by P1.M1.T1.S1.

Task 7: VALIDATE (all gates green)
  - gofmt -w internal/search/search.go internal/search/search_test.go main.go
  - test -z "$(gofmt -l .)"   # whole tree gofmt-clean
  - go vet ./...              # clean
  - go build ./...            # compiles
  - go test ./internal/search/ -v   # all search tests pass (incl. renamed + unchanged boundary test)
  - go test ./...             # whole module green (regression guard)
  - Level 3 smoke test (bug-report reproduce snippet now matches)
  - Level 4 scope check (git diff --name-only == exactly the 3 files)
```

### File 1 edit A — `search.go` matches() body (Task 1)

Exact `oldText` → `newText` (the Keywords loop + `return false`):

```
OLD:
	for _, kw := range s.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			return true
		}
	}
	return false

NEW:
	for _, kw := range s.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			return true
		}
	}
	// Aliases (metadata.aliases) — matched INDIVIDUALLY, same boundary-safety
	// as Keywords: a query spanning two aliases must not match. PRD §10 says
	// aliases "exist only to enrich skpp --search"; this also makes --search
	// consistent with resolve, which resolves by alias (§7.2 step 4).
	for _, a := range s.Aliases {
		if strings.Contains(strings.ToLower(a), q) {
			return true
		}
	}
	// Category (metadata.category) — a single scalar field (PRD §10).
	if strings.Contains(strings.ToLower(s.Category), q) {
		return true
	}
	return false
```

### File 1 edit B — `search.go` matches() doc comment (Task 2)

```
OLD:
// Field scope is EXACTLY the four PRD §6.1 fields. It deliberately does NOT
// include Category or Aliases — both exist on discover.Skill and would be a
// tempting (but spec-violating) addition. PRD §6.1: "tag, frontmatter name,
// description, and metadata.keywords".

NEW:
// Field scope is SIX fields: RelTag, Name, Description, each Keyword, each
// Alias, and Category. PRD §10 states keywords/category/aliases "exist only to
// enrich skpp --search" — so aliases and category ARE searched (decisions.md
// §D4: §10 wins over §6.1's summary field list). This makes --search consistent
// with resolve, which resolves by alias (§7.2 step 4). Aliases are matched
// INDIVIDUALLY (see the Keywords note below) for the same boundary-safety reason.
```

### File 1 edit C — `search.go` package doc (Task 3)

```
OLD:
// query over the four fields PRD §6.1 names for `skpp --search`: the tag, the
// frontmatter name, the description, and each metadata keyword. It is a PURE

NEW:
// query over the six fields PRD §10 enriches for `skpp --search`: the tag, the
// frontmatter name, the description, each metadata keyword, each metadata alias,
// and the metadata category. It is a PURE
```

### File 1 edit D — `search.go` Search() doc (Task 3)

```
OLD:
// substring of ANY of the four PRD §6.1 fields: RelTag (the tag), Name
// (frontmatter name), Description, or any element of Keywords. Input order is

NEW:
// substring of ANY of six fields: RelTag (the tag), Name (frontmatter name),
// Description, any Keyword, any Alias, or Category (PRD §10: keywords/category/
// aliases "exist only to enrich skpp --search"). Input order is
```

### File 2 edit — `search_test.go` test inversion (Task 4)

Replace the ENTIRE `TestSearchDoesNotMatchCategoryOrAliases` function:

```
OLD:
func TestSearchDoesNotMatchCategoryOrAliases(t *testing.T) {
	// PRD §6.1 scopes search to tag/name/description/keywords ONLY. Category and
	// Aliases are on the struct but must NOT be searched.
	in := []discover.Skill{
		{RelTag: "x", Name: "n", Description: "d", Category: "secret-cat", Aliases: []string{"secret-alias"}, HasFM: true},
	}
	if out := Search("secret", in); len(out) != 0 {
		t.Errorf("search must NOT match category/aliases (PRD §6.1 scope); got %+v", out)
	}
}

NEW:
func TestSearchMatchesCategoryAndAliases(t *testing.T) {
	// PRD §10 states metadata.aliases/category "exist only to enrich skpp
	// --search" — so aliases and category ARE searched (decisions.md §D4: §10
	// wins over §6.1). This makes --search consistent with resolve, which
	// resolves by alias (§7.2 step 4). Issue 4 fix: inverts the old
	// TestSearchDoesNotMatchCategoryOrAliases that encoded the wrong behavior.
	withAliases := []discover.Skill{
		{RelTag: "x", Name: "n", Description: "d", Aliases: []string{"secret-alias"}, HasFM: true},
	}
	if out := Search("secret-alias", withAliases); len(out) != 1 {
		t.Errorf("search must match metadata.aliases: query %q got %+v", "secret-alias", out)
	}
	withCategory := []discover.Skill{
		{RelTag: "x", Name: "n", Description: "d", Category: "secret-cat", HasFM: true},
	}
	if out := Search("secret-cat", withCategory); len(out) != 1 {
		t.Errorf("search must match metadata.category: query %q got %+v", "secret-cat", out)
	}
}
```

### File 3 edit A — `main.go` usageText EXAMPLE (Task 5)

```
OLD:
  skpp --search reddit         # substring search over tag/name/description/keywords

NEW:
  skpp --search reddit         # substring search over tag/name/description/keywords/aliases/category
```

### File 3 edit B — `main.go` usageText OPTIONS (Task 5)

```
OLD:
  --search <q>, -s   Substring search over tag / name / description / keywords

NEW:
  --search <q>, -s   Substring search over tag / name / description / keywords / aliases / category
```

### File 3 edit C — `main.go` config struct comment (Task 6)

```
OLD:
	searchMode  bool     // --search <q>/-s : substring search over tag/name/description/keywords (§6.1)

NEW:
	searchMode  bool     // --search <q>/-s : substring search over tag/name/description/keywords/aliases/category (§10)
```

### File 3 edit D — `main.go` --search dispatch comment (Task 6)

```
OLD:
	// --search mode: `skpp --search <q>` / `-s <q>` (PRD §6.1). Filters the index to
	// skills where <q> is a case-insensitive substring of the tag, frontmatter name,
	// description, or any metadata keyword (internal/search), then renders the SAME

NEW:
	// --search mode: `skpp --search <q>` / `-s <q>` (PRD §10). Filters the index to
	// skills where <q> is a case-insensitive substring of the tag, frontmatter name,
	// description, any metadata keyword, any metadata alias, or the metadata category
	// (internal/search), then renders the SAME
```

### Implementation Patterns & Key Details

```go
// PATTERN: mirror the existing Keywords loop for Aliases; Category is a scalar.
//   for _, a := range s.Aliases {                       // boundary-safe, like Keywords
//       if strings.Contains(strings.ToLower(a), q) {
//           return true
//       }
//   }
//   if strings.Contains(strings.ToLower(s.Category), q) {  // scalar, like Name/Description
//       return true
//   }
// WHY: consistency with the existing field checks. The query q is already
//      lowercased once by Search(); each field is lowercased lazily inside Contains
//      (the established pattern — do not pre-lowercase the whole Skill).

// PATTERN: match set only grows — never retracts.
// WHY: adding checks can only add matches. So no positive test breaks; only the
//      one test that asserted NON-match (aliases/category) breaks and is inverted.

// PATTERN: keep the two matching packages (resolve, search) consistent in MEANING.
// WHY: resolve already resolves by alias (§7.2 step 4). After this fix, search
//      also finds aliases — so "an alias is an alias" across both code paths,
//      matching user expectation (§10).
```

### Integration Points

```yaml
NO NEW INTEGRATION POINTS:
  - No new types, no new imports (strings + discover already imported in search.go).
  - No DB, no config, no routes, no new flag, no exit-code change, no stdout/stderr
    contract change. The --search exit codes (0 on matches, 1 on no matches) are
    unchanged; the fix only changes WHICH skills match.
  - discover.Skill, BuildSkill, Index(), resolve.Resolve, ui.PrintList are all
    UNCHANGED. search.Search's signature is unchanged. main's --search dispatch
    is unchanged (it already calls Search then ui.PrintList).
  - README.md is NOT touched (deferred to P1.M5.T3.S1 Mode B doc sync).

PARALLEL-SAFETY (vs P1.M1.T1.S1, running concurrently):
  - main.go regions touched by THIS subtask: usageText (50-90), comment 131,
    comment 317-319.
  - main.go regions touched by P1.M1.T1.S1: c.path branch (268-281) + main_test.go.
  - ZERO overlap. Both subtasks' edits apply cleanly regardless of landing order.
  - Test files are disjoint: this subtask edits search_test.go; P1.M1.T1.S1 edits
    main_test.go.
```

---

## Validation Loop

### Level 1: Format, vet, build (immediate)

```bash
cd /home/dustin/projects/skpp

# Format the touched files, then assert the whole tree is gofmt-clean.
gofmt -w internal/search/search.go internal/search/search_test.go main.go
test -z "$(gofmt -l .)" || { echo "FAIL: gofmt reports unformatted files: $(gofmt -l .)"; exit 1; }

# Compile (catches any typo in the new block).
go build ./... || { echo "FAIL: go build"; exit 1; }

# Static checks.
go vet ./... || { echo "FAIL: go vet"; exit 1; }
echo "Level 1 PASS"
```

### Level 2: Unit tests (component validation)

```bash
cd /home/dustin/projects/skpp

# The search tests specifically (verbose). Must include the RENAMED
# TestSearchMatchesCategoryAndAliases AND the unchanged boundary test.
go test ./internal/search/ -v \
  -run 'TestSearchMatchesCategoryAndAliases|TestSearchKeywordSubstringNotJoinBoundary|TestSearchMatchByKeyword|TestSearchCaseInsensitive|TestSearchNoMatchReturnsEmpty' \
  || { echo "FAIL: targeted search tests"; exit 1; }

# Full search package (regression: nothing else broke).
go test ./internal/search/ -v || { echo "FAIL: go test ./internal/search/"; exit 1; }

# Whole module (regression guard across resolve/ui/discover/main).
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Integration smoke test (the bug-report reproduce snippet, now fixed)

```bash
cd /home/dustin/projects/skpp

# Build the binary.
go build -o skpp . || { echo "FAIL: build"; exit 1; }

# Reproduce the EXACT bug-report scenario (aliases) — was exit 1, now exit 0 + match.
ALIAS_DIR="$(mktemp -d)"
mkdir -p "$ALIAS_DIR/skills/aliased"
printf -- '---\nname: aliased\ndescription: x\nmetadata:\n  aliases: [my-alias]\n---\n\n# x\n' \
  > "$ALIAS_DIR/skills/aliased/SKILL.md"

OUT="$(SKPP_SKILLS_DIR="$ALIAS_DIR/skills" ./skpp --search my-alias 2>/dev/null)"
CODE=$?
test "$CODE" = 0 || { echo "FAIL: alias search exit=$CODE; want 0 (Issue 4 fix)"; rm -rf "$ALIAS_DIR" skpp; exit 1; }
echo "$OUT" | grep -q 'aliased' \
  || { echo "FAIL: alias search did not list the 'aliased' skill; got: $OUT"; rm -rf "$ALIAS_DIR" skpp; exit 1; }
echo "alias search PASS: $OUT"

# Category scenario (the other half of the fix).
CAT_DIR="$(mktemp -d)"
mkdir -p "$CAT_DIR/skills/categorized"
printf -- '---\nname: categorized\ndescription: x\nmetadata:\n  category: writing\n---\n\n# x\n' \
  > "$CAT_DIR/skills/categorized/SKILL.md"

OUT="$(SKPP_SKILLS_DIR="$CAT_DIR/skills" ./skpp --search writ 2>/dev/null)"
CODE=$?
test "$CODE" = 0 || { echo "FAIL: category search exit=$CODE; want 0"; rm -rf "$CAT_DIR" skpp; exit 1; }
echo "$OUT" | grep -q 'categorized' \
  || { echo "FAIL: category search did not list the skill; got: $OUT"; rm -rf "$CAT_DIR" skpp; exit 1; }
echo "category search PASS: $OUT"

# Resolution-by-alias still works (unchanged — proves consistency, not a regression).
OUT="$(SKPP_SKILLS_DIR="$ALIAS_DIR/skills" ./skpp my-alias 2>/dev/null)"
test $? = 0 && echo "$OUT" | grep -q 'aliased' \
  || { echo "FAIL: alias resolution regressed"; rm -rf "$ALIAS_DIR" "$CAT_DIR" skpp; exit 1; }
echo "alias resolution still PASS"

# Regression: a query matching NOTHING still exits 1 with empty stdout (§6.4).
OUT="$(SKPP_SKILLS_DIR="$ALIAS_DIR/skills" ./skpp --search zzz-nope 2>/dev/null)"
test $? = 1 || { echo "FAIL: no-match should exit 1; got $?"; rm -rf "$ALIAS_DIR" "$CAT_DIR" skpp; exit 1; }
test -z "$OUT" || { echo "FAIL: no-match stdout should be empty; got: $OUT"; rm -rf "$ALIAS_DIR" "$CAT_DIR" skpp; exit 1; }
echo "no-match regression PASS"

rm -rf "$ALIAS_DIR" "$CAT_DIR" skpp
echo "Level 3 PASS"
```

### Level 4: Scope-boundary & contract check

```bash
cd /home/dustin/projects/skpp

# matches() now consults Aliases AND Category (the two new checks).
grep -q 'for _, a := range s.Aliases' internal/search/search.go \
  || { echo "FAIL: Aliases loop missing in matches()"; exit 1; }
grep -qE 'strings\.Contains\(strings\.ToLower\(s\.Category\), q\)' internal/search/search.go \
  || { echo "FAIL: Category check missing in matches()"; exit 1; }

# The matches() doc comment no longer says aliases/category are excluded.
! grep -q 'deliberately does NOT' internal/search/search.go \
  || { echo "FAIL: stale 'deliberately does NOT' doc still present"; exit 1; }
# No lingering 'four fields' anywhere in search.go.
! grep -q 'four fields' internal/search/search.go \
  || { echo "FAIL: stale 'four fields' in search.go"; exit 1; }

# Aliases matched INDIVIDUALLY (no Join).
! grep -qE 'strings\.Join\(s\.Aliases' internal/search/search.go \
  || { echo "FAIL: do not Join aliases (boundary safety)"; exit 1; }
# No nil guard on Aliases (range over nil is safe).
! grep -qE 's\.Aliases != nil' internal/search/search.go \
  || { echo "FAIL: unnecessary nil guard on Aliases"; exit 1; }

# Test renamed + inverted (old name gone, new name present).
! grep -q 'func TestSearchDoesNotMatchCategoryOrAliases' internal/search/search_test.go \
  || { echo "FAIL: old test name still present"; exit 1; }
grep -q 'func TestSearchMatchesCategoryAndAliases' internal/search/search_test.go \
  || { echo "FAIL: new test TestSearchMatchesCategoryAndAliases missing"; exit 1; }
# The boundary test is UNCHANGED.
grep -q 'func TestSearchKeywordSubstringNotJoinBoundary' internal/search/search_test.go \
  || { echo "FAIL: boundary test was removed/renamed (must stay)"; exit 1; }

# usageText --search description now lists aliases + category (both lines).
grep -q 'substring search over tag/name/description/keywords/aliases/category' main.go \
  || { echo "FAIL: usageText EXAMPLE line not updated"; exit 1; }
grep -q 'Substring search over tag / name / description / keywords / aliases / category' main.go \
  || { echo "FAIL: usageText OPTIONS line not updated"; exit 1; }

# EXACTLY 3 files changed — nothing else.
CHANGED="$(git diff --name-only HEAD -- internal/search/search.go internal/search/search_test.go main.go | wc -l)"
test "$CHANGED" -le 3 || { echo "FAIL: unexpected file count"; exit 1; }
# Must NOT have touched skill.go / resolve.go / index.go / README / PRD.
git diff --quiet -- internal/discover/skill.go      || { echo "FAIL: skill.go changed (out of scope)"; exit 1; }
git diff --quiet -- internal/resolve/resolve.go     || { echo "FAIL: resolve.go changed (out of scope)"; exit 1; }
git diff --quiet -- internal/discover/index.go      || { echo "FAIL: index.go changed (out of scope)"; exit 1; }
git diff --quiet -- README.md                        || { echo "FAIL: README.md changed (deferred to P1.M5.T3)"; exit 1; }
git diff --quiet -- PRD.md                           || { echo "FAIL: PRD.md changed (read-only)"; exit 1; }

echo "Level 4 PASS (scope + contract respected)"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l .` empty, `go build ./...` compiles, `go vet ./...` clean
- [ ] Level 2 PASS — `go test ./internal/search/ -v` green (incl. renamed test + unchanged boundary test); `go test ./...` whole module green
- [ ] Level 3 PASS — bug-report alias snippet now exits 0 + lists the skill; category snippet matches; alias resolution still works; no-match still exits 1 with empty stdout
- [ ] Level 4 PASS — Aliases loop + Category check present; no stale "four fields"/"deliberately does NOT"; no Join, no nil guard; test renamed; usageText updated; exactly 3 files changed

### Feature Validation
- [ ] `skpp --search <alias>` returns skills whose `metadata.aliases` contain the query (case-insensitive)
- [ ] `skpp --search <category-substring>` returns skills whose `metadata.category` contains the query
- [ ] Existing tag/name/description/keyword matches are unchanged (match set only grew)
- [ ] Empty query still matches all; no-frontmatter skills still match only by tag
- [ ] `--search` with no matches still exits 1 with empty stdout (§6.4 unchanged)
- [ ] Alias resolution (`skpp <alias>`) is unaffected (resolve.go untouched)

### Code Quality Validation
- [ ] Aliases loop mirrors the Keywords loop EXACTLY (range + Contains(ToLower)); Category is a scalar Contains
- [ ] No nil guard on `s.Aliases` (range over nil is safe — matches Keywords style)
- [ ] Aliases matched INDIVIDUALLY (no Join — boundary-safe)
- [ ] All 7 stale "four fields" doc references updated (search.go ×3, main.go ×4)
- [ ] `usageText` OPTIONS row still aligns (description is the last column)
- [ ] Tests use literal `discover.Skill{}` for the alias/category fixtures (sk() doesn't set them)

### Scope Discipline (Mode A)
- [ ] `internal/discover/skill.go` NOT modified (Aliases/Category already exist + populated)
- [ ] `internal/resolve/resolve.go` NOT modified (alias resolution already works)
- [ ] `internal/discover/index.go` NOT modified (walk unchanged)
- [ ] `README.md` NOT modified (deferred to P1.M5.T3.S1 Mode B doc sync)
- [ ] `PRD.md` / `tasks.json` / `prd_snapshot.md` NOT modified (read-only / orchestrator-owned)
- [ ] `git diff --name-only` == exactly `internal/search/search.go`, `internal/search/search_test.go`, `main.go`

---

## Anti-Patterns to Avoid

- ❌ **Don't add a nil guard before the Aliases range.** `range` over a nil slice
  is a safe no-op (verified §4). The existing Keywords loop relies on this. A
  guard diverges from the established style and implies a falsehood.
- ❌ **Don't `strings.Join` the aliases.** They must be matched INDIVIDUALLY so a
  query spanning two aliases can't match (boundary safety — same as Keywords,
  verified §5).
- ❌ **Don't "fix" other passing tests.** The match set only GROWS; only the ONE
  test asserting aliases/category do NOT match breaks. Invert that one; leave the
  rest (including `TestSearchKeywordSubstringNotJoinBoundary`).
- ❌ **Don't touch `skill.go` / `resolve.go` / `index.go`.** The data already
  exists and is populated; alias resolution already works. This is a consumer-only
  change to `search.matches()`.
- ❌ **Don't leave stale "four fields" docs.** There are 7 references (research §7).
  Update all of them or the docs contradict the code. The item names 2 explicitly
  (matches() doc, usageText OPTIONS); the other 5 are mandatory for consistency.
- ❌ **Don't edit `README.md`.** It's deferred to P1.M5.T3.S1 (Mode B doc sync).
  This subtask is Mode A: code + in-source docs + tests.
- ❌ **Don't change exit codes or the stdout/stderr contract.** `--search` still
  exits 0 on matches / 1 on no matches; no-match still prints nothing to stdout.
  The fix only changes WHICH skills match, not the output discipline.
- ❌ **Don't collide with P1.M1.T1.S1.** It owns main.go's `c.path` branch
  (~268-281) and `main_test.go`. Your main.go edits are in usageText (50-90) and
  comments (131, 317-319) — no overlap — and your test file is `search_test.go`.

---

## Confidence Score

**10/10** — The change is ~8 lines of matching logic (an Aliases `for` loop that
mirrors the existing Keywords loop, plus a Category scalar check) plus mechanical
doc/test alignment. The proposed `matches()` block was **executed** in a throwaway
Go 1.25 module against 8 fixtures (alias match, category match, case-insensitivity,
no-false-positive, nil-safety, keyword-boundary-preserved, alias-boundary-safety,
category-substring) — all PASS (research §3). The `Skill.Aliases` / `Skill.Category`
fields already exist and are populated (`skill.go:42-43, 113-114`), so no data-model
work. Every stale doc reference is inventoried with exact old/new text (research §7).
The match set only grows, so the sole breaking test is the one the item names for
inversion. Residual risk is limited to transcription typos, caught by the Level 4
grep-based contract checks.
