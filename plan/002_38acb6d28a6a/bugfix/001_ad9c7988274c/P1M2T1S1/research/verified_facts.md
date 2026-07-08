# Verified Facts — P1.M2.T1.S1: Add `c.path` to the tags-conflict predicate in `exclusivityError` (Issue 3)

Bugfix round `plan/002_38acb6d28a6a/bugfix/001_ad9c7988274c`. Every claim below was
read directly from the live source at `/home/dustin/projects/skilldozer` (main.go
read around `exclusivityError`; main_test.go exclusivity tests read at exact line
ranges). Module `github.com/dabstractor/skilldozer`, `go 1.25`. This is a 1-point,
one-line predicate change + message string + doc comment + tests. Zero new deps.

---

## §1 — The exact buggy line and the fix (CURRENT main.go)

`exclusivityError` (func at main.go:708; doc comment 688-707). The buggy predicate
is at **main.go:724-725** (the contract cites 702/708 and "686" — those are the
PRD-write-time numbers; the CURRENT line is 724 because the bugfix round's M1 work
shifted things; anchor by TEXT, not line number):

```go
hasTags := len(c.tags) > 0
if hasTags && (c.list || c.searchMode || c.all) {   // ← c.path MISSING (the bug)
    return true, "skilldozer: tags cannot be combined with --list/--search/--all"
}
```

**The asymmetry (the whole bug):** `c.path` IS a first-class member of the
mutually-exclusive "inspection mode" set in BOTH sibling predicates:
- the mode+mode count set (main.go:715): `for _, b := range []bool{c.path, c.list, c.searchMode, c.all}`
- the check+mode set      (main.go:730): `if c.check && (c.path || c.list || c.searchMode || c.all)`

…but it is OMITTED only from the tags predicate at 724. So `tags + --list/search/all`
exits 2, but `tags + --path` silently runs `--path` and drops the tag — even an
UNKNOWN tag (`NONEXISTENTTAG --path` → exit 0, the repro in bug_fixes_validation.md
§ISSUE 3). A user typing `skilldozer myskill --path` expecting the skill's path
instead receives the STORE path with no warning.

**The fix (one line — predicate + message):**
```go
if hasTags && (c.path || c.list || c.searchMode || c.all) {
    return true, "skilldozer: tags cannot be combined with --path/--list/--search/--all"
}
```

No exit-code change (still 2). No ordering change (still the 2nd family, checked
after the mode+mode count and before check+tags). No new family (the existing
"tags" family's predicate is expanded; the doc comment's "four families" count is
unchanged).

---

## §2 — The repro and the N1 precedent (check+path was already fixed the same way)

bug_fixes_validation.md §ISSUE 3 confirmed:
```
./skilldozer NONEXISTENTTAG --path; echo "exit=$?"   # 0  (tag silently dropped) ← BUG
./skilldozer NONEXISTENTTAG --list; echo "exit=$?"   # 2  (as expected)
```

The SAME asymmetry was ALREADY fixed for the `check` subcommand — see the doc
comment at main.go:698-701 and the test `TestRunExclusivityCheckAndPath`
(main_test.go:1786), whose comment says: "N1: previously fell through to dispatch
and ran --path, silently ignoring `check`. `--path` is now in the check+mode set …
closing the exclusivity asymmetry." Issue 3 closes the IDENTICAL asymmetry for the
tags case. `--path` is now treated consistently in all three predicates.

---

## §3 — Zero existing-test breakage from the message-string change (grep-verified)

The message changes from `"--list/--search/--all"` to `"--path/--list/--search/--all"`.
Grepped `main_test.go` + `main.go` for the exact tags-message substrings:

```
$ grep -rn "tags cannot be combined\|--list/--search/--all\|--path/--list" main_test.go main.go
# main.go hits only (the message definitions): 698 (doc), 721 (listing-modes msg),
# 725 (the tags msg we are editing), 731 (check+mode msg), 743 (init+mode msg).
# main_test.go: ZERO hits on any exact message string.
```

The ONLY test asserting a tags-exclusivity message is `TestRunExclusivityTagsAndList`
(main_test.go:1726): `if !strings.Contains(errOut.String(), "cannot be combined")`.
It uses `Contains("cannot be combined")` — NOT the mode list. So prepending
`--path/` to the message is invisible to it. `TestRunExclusivityTagsAndSearch`
(@1732) and `TestRunExclusivityTagsAndAll` (@1744) do not assert the message text
at all. `TestRunExclusivityCheckAndPath` (@1786) asserts the CHECK message (different
family) and is unaffected. **Net: the message change breaks zero tests.**

---

## §4 — Test plan: mirror TestRunExclusivityTagsAndList + a direct unit test

(a) `TestRunExclusivityTagsAndList` (main_test.go:1717-1729) is the EXACT template
for the primary run-level test — copy its shape, swap `--list` for `--path`:
```go
func TestRunExclusivityTagsAndList(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"foo", "--list"}, &out, &errOut)
	if code != 2 { t.Fatalf("run(foo --list): code=%d; want 2", code) }
	if out.Len() != 0 { t.Errorf("stdout=%q; want empty", out.String()) }
	if !strings.Contains(errOut.String(), "cannot be combined") {
		t.Errorf("stderr=%q; want an exclusivity message", errOut.String())
	}
}
```
→ `TestRunExclusivityTagsAndPath`: `run([]string{"foo", "--path"})` → code 2, empty
stdout, stderr Contains("cannot be combined"). Also add `TestRunExclusivityPathAndTag`
(reversed order `--path foo`) since the contract OUTPUT requires BOTH orderings
("`<tag> --path` (or `--path <tag>`)").

(b) A DIRECT unit test that calls `exclusivityError` itself — this is the strongest
test because it isolates the ONE line that changed, independent of parseArgs/run:
```go
func TestExclusivityErrorTagsAndPath(t *testing.T) {
	bad, msg := exclusivityError(config{tags: []string{"foo"}, path: true})
	if !bad { t.Fatalf("exclusivityError(tags+path)=bad=false; want true (Issue 3)") }
	if !strings.Contains(msg, "tags cannot be combined") {
		t.Errorf("msg=%q; want 'tags cannot be combined'", msg)
	}
	if !strings.Contains(msg, "--path") {
		t.Errorf("msg=%q; want it to mention --path", msg)
	}
}
```

(c) DO NOT add a tags case to `TestExclusivityErrorListingModes` (main_test.go:2128)
even though bug_fixes_validation.md §ISSUE 3 suggested it. That table asserts every
`bad` case contains `"mutually exclusive"` (the listing-mode family's wording); a
tags+path case returns `"tags cannot be combined with …"` which does NOT contain
"mutually exclusive" → it would fail that table's assertion. Use the dedicated
direct-unit test (b) instead. (The bug doc's suggestion was shape-correct but
table-inappropriate; this PRP corrects the placement.)

(d) No store fixture / env needed for ANY of these tests — exclusivity runs at
run() step 4, BEFORE `skillsdir.Find()` (verified: the exclusivity branch is
`fmt.Fprintln(stderr, msg); return 2` at ~main.go:443-446, before any dispatch).
So `run([]string{"foo","--path"})` exits 2 without touching the filesystem.

---

## §5 — The doc-comment edit (Mode A, main.go:688-707)

The bullet at **main.go:694** currently reads:
```
//   - tags + a listing mode (--list/--search/--all) — PRD §6.3 explicit
```
Update to reflect `--path` is now included (and cite Issue 3). Also optionally add a
one-line note near the existing N1 paragraph (698-701) that the SAME asymmetry is now
closed for tags. Keep the "four families" framing (the count is unchanged — the tags
family's predicate is expanded, not split). The function-level exit-code/wording
conventions are unchanged.

---

## §6 — No conflict with the parallel sibling P1.M1.T2.S2 (disjoint regions)

P1.M1.T2.S2 (Implementing) edits:
- `run()`: inserts a `storeMissingValue` guard between the `unknownFlag` block and
  the `exclusivityError` block (~main.go:438), and updates the run() precedence
  COMMENT (a different comment block, ~main.go:412).
- `main_test.go`: adds 4 run()-level `--store` no-value tests.

This subtask edits:
- `exclusivityError`: the tags predicate + message (main.go:724-725) and its DOC
  COMMENT (main.go:694) — a different region and a different comment block.
- `main_test.go`: adds tags+path tests — different functions.

**Disjoint in both files; no text-level overlap; the two changesets compose.** The
only interaction is run() ordering: P1.M1.T2.S2's guard runs BEFORE exclusivity, but
it only fires for `--store`-no-value shapes (irrelevant to tags+path). tags+path
cases skip that guard and hit exclusivity exactly as before. Land in either order.

---

## §7 — Scope discipline (what NOT to touch)

- Do NOT touch `parseArgs`, the `config` struct, or the `--store` branches (P1.M1.T2).
- Do NOT touch the mode+mode count set (main.go:715) or the check+mode set
  (main.go:730) — those ALREADY include `c.path` correctly. Only the tags predicate
  (724) is buggy.
- Do NOT change any exit code or the family ordering.
- Do NOT modify README (Mode B, P1.M3.T1) or PRD.md/tasks.json/prd_snapshot/.gitignore.
- Do NOT add deps/imports (predicate change + message string + Contains test use
  only existing constructs; `strings` is already imported in both files).

---

## §8 — Validation (Go toolchain, verified commands)

```bash
gofmt -l main.go main_test.go   # must print NOTHING
go vet ./...                    # exit 0
go test -run 'TagsAndPath|PathAndTag|ExclusivityErrorTagsAndPath' -v ./...  # the 3 new tests
go test ./...                   # whole module green; zero regressions (§3)
git diff --quiet go.mod go.sum && echo deps unchanged
# Manual repro (now FIXED): NONEXISTENTTAG --path → exit 2 (was 0)
go build -o /tmp/sd . && env -u SKILLDOZER_SKILLS_DIR /tmp/sd NONEXISTENTTAG --path; echo "exit=$? (want 2)"
```
