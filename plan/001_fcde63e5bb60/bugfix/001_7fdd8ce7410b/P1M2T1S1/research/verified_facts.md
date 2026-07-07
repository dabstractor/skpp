# Verified Facts — P1.M2.T1.S1: Extend `--search` to match aliases + category

Bug fix for QA **Issue 4** (Minor): `--search` matches `metadata.keywords` but not
`metadata.aliases` or `metadata.category`, even though §10 says all three "exist
only to enrich `skpp --search`". Decision `decisions.md §D4`: **§10 wins over
§6.1** — extend `matches()` to scan Aliases and Category.

Every claim below was verified by reading the working tree (`internal/search/`,
`internal/discover/skill.go`, `main.go`) AND by executing the proposed `matches()`
logic in a throwaway Go 1.25 module against 8 fixtures (all PASS). The change is
surgical: one code block + doc-comment rewrites + one test inversion + usageText.

---

## §1 — Current `matches()` scans exactly FOUR fields (the bug)

`internal/search/search.go:59-77` (read in full):

```go
func matches(q string, s discover.Skill) bool {
	if strings.Contains(strings.ToLower(s.RelTag), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Description), q) {
		return true
	}
	for _, kw := range s.Keywords {
		if strings.Contains(strings.ToLower(kw), q) {
			return true
		}
	}
	return false
}
```

`RelTag`, `Name`, `Description`, and each element of `Keywords`. That's it. The
`Skill.Aliases []string` and `Skill.Category string` fields exist on the struct
but are NOT consulted. The doc comment at `search.go:51-56` documents this as
DELIBERATE per §6.1 — that interpretation is now overridden by `decisions.md §D4`.

---

## §2 — `Skill.Aliases` and `Skill.Category` ARE populated and available

`internal/discover/skill.go:42-43` (the struct):

```go
type Skill struct {
	...
	Category    string   // line 42
	Aliases     []string // line 43
	...
}
```

`skill.go:105-121` `BuildSkill` populates them from `Frontmatter.Metadata`:

```go
category, _ := fm.Metadata["category"].(string) // nil-map read is safe; comma-ok
return Skill{
	...
	Category:    category,                              // line 113
	Aliases:     toStringSlice(fm.Metadata["aliases"]), // line 114
	...
}
```

So by the time `Search(query, []discover.Skill)` runs (main supplies the index
from `discover.Index`), every `Skill` already carries its `Aliases` and `Category`
ready to be matched. **No struct change, no discover change, no Index() change.**
This subtask is purely the `search.matches()` consumer.

---

## §3 — Proposed `matches()` addition: VERIFIED (8/8 fixtures PASS)

Executed in a throwaway module replicating the exact struct + the proposed
`matches()` (current 4 fields + Aliases loop + Category check). Results:

| # | Fixture / query | Expected | Result |
|---|---|---|---|
| 1 | skill with `Aliases: ["secret-alias"]`, query `"secret-alias"` | match | ✅ PASS |
| 2 | skill with `Category: "secret-cat"`, query `"secret-cat"` | match | ✅ PASS |
| 3 | `"SECRET-ALIAS"` / `"SECRET-CAT"` (case-insensitive) | match | ✅ PASS |
| 4 | plain skill (no aliases/category), query `"secret"` | NO match | ✅ PASS (no false positive) |
| 5 | skill with `Aliases: nil`, `Category: ""` | no panic, no false match | ✅ PASS |
| 6 | keyword-boundary query `"wriocial"` vs `["writing","social"]` | NO match | ✅ PASS (unchanged keyword logic) |
| 7 | alias-boundary query `"aliasbar"` vs `["foo-alias","bar-alias"]` | NO match | ✅ PASS (aliases matched individually) |
| 8 | category substring `"secret"` vs `"secret-cat"` | match | ✅ PASS |

The added block (mirrors the existing Keywords loop):

```go
	// Aliases — matched INDIVIDUALLY (same boundary-safety as Keywords).
	for _, a := range s.Aliases {
		if strings.Contains(strings.ToLower(a), q) {
			return true
		}
	}
	// Category — a single scalar field.
	if strings.Contains(strings.ToLower(s.Category), q) {
		return true
	}
```

---

## §4 — `range` over a nil `Aliases` slice is SAFE (no panic, no special case)

Go's `range` over a nil slice iterates zero times — it is NOT a panic. Verified
(fixture #5: `Aliases: nil`, no panic, `len(Search("anything", nilAlias))==0`).
So the new `for _, a := range s.Aliases` needs NO nil guard. This matches the
existing Keywords loop, which also ranges over a possibly-nil `s.Keywords`
without guarding (a no-frontmatter skill has `Keywords == nil`).

`strings.Contains(strings.ToLower(s.Category), q)` with `Category == ""` is also
safe: `ToLower("") == ""`, `Contains("", q)` is true only when `q == ""` (the
empty-query-matches-all case, already the documented behavior). No special case.

---

## §5 — Aliases matched INDIVIDUALLY (boundary-safe), mirroring Keywords

The existing Keywords loop tests each keyword separately (NOT `strings.Join`'d) so
a query spanning a boundary between two keywords cannot match. The Aliases loop
uses the IDENTICAL pattern for the same reason: a query spanning two aliases
(e.g. `"aliasbar"` across `["foo-alias","bar-alias"]`) must NOT match. Verified
(fixture #7). The existing `TestSearchKeywordSubstringNotJoinBoundary` test is
UNCHANGED by this fix; aliases get the same boundary semantics automatically.

---

## §6 — The existing test encodes the WRONG behavior and MUST be inverted

`internal/search/search_test.go:126` `TestSearchDoesNotMatchCategoryOrAliases`:

```go
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
```

This asserts `len(out) != 0` is a FAILURE — i.e. it asserts aliases/category do
NOT match. After the fix, `Search("secret", in)` WILL match (both "secret-cat"
and "secret-alias" contain "secret"), so this test FAILS against the new code.
**It must be REPLACED** with `TestSearchMatchesCategoryAndAliases` that asserts
the OPPOSITE. The item description's contract for the replacement test:

- `Search("secret-alias", skillWithAliases)` returns the skill (len 1)
- `Search("secret-cat", skillWithCategory)` returns the skill (len 1)

Note: the OLD test used a SINGLE skill with BOTH fields set and queried `"secret"`
(a substring of both). The NEW test should split into two skills (one with only
aliases, one with only category) and query the full value each, so each
assertion is unambiguous (proves the alias path and the category path
independently). This is cleaner than the old single-skill form and matches the
item's `skillWithAliases` / `skillWithCategory` phrasing.

---

## §7 — Complete inventory of STALE "four fields" doc references

After the code change, every doc/comment that says search covers "four fields" /
"tag/name/description/keywords" / "deliberately does NOT include Category or
Aliases" becomes FALSE and must be updated for consistency. Full inventory
(verified by grep + reading):

### `internal/search/search.go` (3 references)
1. **Lines 1-3** (package doc): `// query over the four fields PRD §6.1 names for`
   `skpp --search`: the tag, the frontmatter name, the description, and each
   metadata keyword.` — STALE.
2. **Lines 20-22** (Search() doc): `// substring of ANY of the four PRD §6.1
   fields: RelTag (the tag), Name (frontmatter name), Description, or any element
   of Keywords.` — STALE.
3. **Lines 49-56** (matches() doc): `// Field scope is EXACTLY the four PRD §6.1
   fields. It deliberately does NOT include Category or Aliases ...` — STALE.
   **(The item EXPLICITLY requires this one be rewritten to state they ARE now
   included per PRD §10.)**

### `main.go` (4 references)
4. **Line 71** (usageText EXAMPLE comment): `  skpp --search reddit         #
   substring search over tag/name/description/keywords` — STALE.
5. **Line 78** (usageText OPTIONS table): `  --search <q>, -s   Substring
   search over tag / name / description / keywords` — STALE.
   **(The item EXPLICITLY requires this one: append " / aliases / category".)**
6. **Line 131** (config struct field comment): `searchMode  bool     // --search
   <q>/-s : substring search over tag/name/description/keywords (§6.1)` — STALE
   (code comment; update for consistency + change §6.1→§10 per D4).
7. **Lines 317-319** (--search dispatch comment): `// ... Filters the index to
   skills where <q> is a case-insensitive substring of the tag, frontmatter name,
   description, or any metadata keyword (internal/search) ...` — STALE.

### NOT stale (do NOT touch)
- `main.go:149` `// PRD §6.1/§6.2: --search/-s take exactly one value` — about
  flag VALUE arity, not search FIELDS. Leave as-is.
- `main.go:481,485` — exclusivity comments, not fields. Leave as-is.

All 7 stale references are updated by this subtask so the docs never contradict
the code. The 3 in search.go and the 2 in usageText are the load-bearing ones;
the 2 code comments (lines 131, 317) are hygiene but leaving them would mislead
the next maintainer.

---

## §8 — usageText column alignment is NOT broken by the longer --search line

`main.go:78` is a row in a fixed-ish-width OPTIONS table:

```
  --search <q>, -s   Substring search over tag / name / description / keywords
```

The description column is the LAST column (no column follows it), and rows have
variable-length descriptions already (e.g. `check` row: "Validate every skill on
disk (report OK / WARN / ERROR)" is longer). Appending ` / aliases / category`
extends ONLY this row's trailing text — it does not shift any other column or
row. Verified by eye; no wrap/truncate logic exists. The example comment
(line 71) is free-form trailing text after `#`, also safe to extend. **No
alignment regression.**

---

## §9 — Scope boundaries (what NOT to touch)

This is a CONSUMER-ONLY change to `search.matches()`. The data is already there.

- **DO NOT modify** `internal/discover/skill.go` — `Skill`, `Aliases`, `Category`,
  and `BuildSkill` already exist and populate the fields. No struct change.
- **DO NOT modify** `internal/resolve/resolve.go` — alias RESOLUTION (§7.2 step 4)
  already works (`skpp <alias>` resolves). This fix only makes SEARCH consistent;
  resolution is untouched.
- **DO NOT modify** `internal/discover/index.go` — the walk is unchanged.
- **DO NOT modify** `README.md` — deferred to **P1.M5.T3.S1** ("Update README.md
  to reflect --path source reporting and search field expansion", Mode B doc
  sync). This subtask is Mode A (code + in-source docs + tests).
- **DO NOT modify** `PRD.md` (read-only) / `tasks.json` / `prd_snapshot.md`.

### Parallel-safety with P1.M1.T1.S1 (running concurrently)
P1.M1.T1.S1 edits `main.go`'s `c.path` branch (~lines 268-281) + `main_test.go`.
This subtask edits `main.go` usageText (50-90), config comment (131), and search
dispatch comment (317-319) + `search.go` + `search_test.go`. **Zero region
overlap** with the `c.path` branch, and a DIFFERENT test file (`search_test.go`
vs `main_test.go`). Both edits apply cleanly regardless of which lands first.
P1.M1.T1.S1 explicitly does NOT touch usageText or the search dispatch comment.

---

## §10 — Validation gates (verified executable in this repo)

- `gofmt -l internal/search/search.go internal/search/search_test.go main.go` →
  must print nothing after edits.
- `go vet ./...` → clean.
- `go build ./...` → compiles.
- `go test ./internal/search/ -v` → all search tests pass, INCLUDING the renamed
  `TestSearchMatchesCategoryAndAliases` and the unchanged
  `TestSearchKeywordSubstringNotJoinBoundary`.
- `go test ./...` → whole module green (regression guard).
- Level 3: the bug report's exact reproduce steps now MATCH (was exit 1 "no
  skills matched"; now exit 0 with the skill listed). Plus a category smoke test.
- Level 4: `git diff --name-only` shows ONLY `internal/search/search.go`,
  `internal/search/search_test.go`, `main.go`. No skill.go/resolve.go/index.go/
  README.md/PRD.md changes.
