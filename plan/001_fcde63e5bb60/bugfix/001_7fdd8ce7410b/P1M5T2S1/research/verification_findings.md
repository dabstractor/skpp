# Verification Findings — P1.M5.T2.S1

Verify that `internal/check/check.go` lines **123-129** (the `Check` doc-comment
"omission rationale" block) accurately document why the PRD §9 rule
"ERROR: skill dir has no SKILL.md" is not implemented, against the three points
the contract requires. All line numbers from `nl -ba internal/check/check.go`
(commit at research time; working tree).

---

## 1. The three contract points to verify

| # | Contract point | Evidence required |
|---|----------------|-------------------|
| a | WHY the rule can't fire: `discover.Index` only emits dirs WITH a `SKILL.md` | index.go discovery semantics |
| b | WHY a heuristic is rejected: false-positives on legitimate grouping dirs | issue_analysis §Issue 7, decisions §D7 |
| c | The rule is REFRAMED as `invalid SKILL.md frontmatter` (contract cites check.go:150) | the `perr != nil` branch comment + ERROR |

---

## 2. The current 123-129 comment (verbatim)

```go
//  (3 lines of §118-122 preamble above, then:)
//
// check does NOT scan for "directories that lack SKILL.md but look like skills":
// discover.Index only emits dirs that CONTAIN a SKILL.md, and a heuristic for the
// gap would false-positive on legitimate grouping dirs (research §2). The §9
// "empty besides SKILL.md" WARN is intentionally NOT implemented (research §3):
// the shipped example skill IS only SKILL.md, and enabling it would break the
// §13 acceptance ("reports the example as OK").
func Check(skills []discover.Skill) Report {          // line 130
```

---

## 3. Point-by-point verdict

### (a) WHY the rule can't fire — VERIFIED ACCURATE ✅

- Comment: "discover.Index only emits dirs that CONTAIN a SKILL.md".
- index.go:46 doc: *"A 'skill' is any directory that directly contains a SKILL.md
  file"*. The `WalkDir` callback (`index.go` body) returns early for every entry
  that is not the literal file `SKILL.md` (`if d.IsDir() || d.Name() != "SKILL.md"
  { return nil }`). It therefore emits ONLY dirs that contain `SKILL.md`.
- PRD §9 line 148: *"A skill = any directory that directly contains a SKILL.md."*
- ⇒ A grouping dir without `SKILL.md` is never indexed, never passed to `Check`.
  The §9 "skill dir has no SKILL.md" ERROR has no input it could ever fire on.
  Comment is accurate. **No change needed for (a).**

### (b) WHY a heuristic is rejected — VERIFIED ACCURATE ✅

- Comment: "a heuristic for the gap would false-positive on legitimate grouping
  dirs (research §2)".
- issue_analysis.md §Issue 7: *"a heuristic would false-positive on legitimate
  grouping dirs"*.
- decisions.md §D7: *"A heuristic (e.g. flagging dirs with scripts/ but no
  SKILL.md) would false-positive on legitimate grouping directories."*
- PRD §7.1 line 39: *"Arbitrary sibling assets (scripts/, references/, assets/)
  are allowed"* — confirms a `scripts/`-bearing dir is a legit grouping dir.
- ⇒ Comment is accurate. **No change needed for (b).**

### (c) Reframed as "invalid SKILL.md frontmatter" — GAP ⚠️

- The 123-129 block does NOT state the reframing. It documents the OMISSION
  (why no scan, why no heuristic) and a SEPARATE §9 omission (the "empty besides
  SKILL.md" WARN), but never connects the §9 rule to its actionable reframe.
- The reframing IS documented — but at a DIFFERENT location, inside the switch:

  ```go
  case perr != nil:                                        // line 145
      // Malformed YAML between fences, OR the file vanished between Index and
      // check (race) -> ParseFrontmatter returns the os/yaml error. This is the
      // reframed §9 "skill dir has no SKILL.md": an UNUSABLE SKILL.md.
      findings = append(findings, Finding{LevelError,      // line 149
          "invalid SKILL.md frontmatter: " + perr.Error()})
  case !fm.HasFM:                                          // line 150  <- NEXT branch
  ```

- So a reader landing on the 123-129 rationale block learns (a) and (b) but NOT
  (c). The comment is **incomplete** on contract point (c).
- **LINE-NUMBER DRIFT (flag for the implementer):** the contract cites
  "check.go:150" for the reframed ERROR. The actual ERROR `Finding` is at
  **line 149**; **line 150 is the NEXT branch** (`case !fm.HasFM:`). The
  reframing prose ("...reframed §9 'skill dir has no SKILL.md'...") is on
  **lines 147-148**. Read 145-149, not 150.

  ⇒ **Refine the 123-129 block (clarity only) to add the reframing so it is
  self-contained on (c).** No code, no behavioral change. The 146-149 branch is
  already correct AND tested (`check_test.go:77 TestCheckMalformedYAML` asserts
  the "invalid SKILL.md frontmatter" ERROR) — do NOT touch it.

---

## 4. Corroboration sources (all read; line numbers confirmed)

- `PRD.md`
  - line 148 (§9): skill = dir directly containing SKILL.md (supports a).
  - line 202 (§9): "ERROR: skill dir has no `SKILL.md`." (the rule being reframed).
  - line 207 (§9): "WARN: a skill dir is empty besides SKILL.md ... optional."
  - line 334 (§13): "./skpp check  # exits 0, reports the example as OK" (gate).
  - line 317 (§13): "go build -o skpp . && echo OK" (build command).
- `internal/discover/index.go` line 46 + WalkDir body (supports a).
- `internal/check/check.go` lines 123-129 (the comment), 145-150 (the reframe).
- `internal/check/check_test.go:77 TestCheckMalformedYAML` (the reframe is tested).
- `architecture/issue_analysis.md` §Issue 7 (supports a,b).
- `architecture/decisions.md` §D7 (supports a,b,c — "documentation-only").

---

## 5. Validation commands (run during research; all PASS in current tree)

```
go test ./internal/check/                         # ok  (cached)
go build -o skpp .                                # build OK
./skpp check                                      # "OK    example (example)" ; exit 0
gofmt -l internal/check/check.go                  # (empty = clean)
go vet ./internal/check/                          # clean
```

Note: there is NO `./skpp` binary committed (it is gitignored, §16 line 1
`/skpp`). `go build -o skpp .` is required before the `./skpp check` gate.

---

## 6. Conflict check with parallel/in-flight items

- P1.M5.T1.S1 edits **`.gitignore`** only → disjoint from `check.go`. No conflict.
- P1.M4.T2.S1 (in flight) edits **`main.go`** `exclusivityError` → disjoint.
- Earlier-done items touched `main.go`, `ui.go`, `search.go`, `skillsdir.go`.
- **Nothing else touches `internal/check/check.go`.** Zero conflict for this item.

---

## 7. Bottom line

(a) accurate, (b) accurate, (c) is the single gap. One comment-only refinement to
the 123-129 block (cross-referencing the reframe) closes it. No behavioral change;
all tests + the §13 `skpp check` gate continue to pass.
