# Issue Analysis — skpp Bug Fix 001

Per-issue root cause, fix surface, and test impact. All file:line references
confirmed against the working tree (commit 13a687a).

---

## Issue 1 (MAJOR) — `--path` does not report which discovery rule won

**Root cause**: `main.go` `--path` branch discards the `Source` return value:
```go
dir, _, err := skillsdir.Find() // src is for reporting only; not printed
```
The `Source` enum, its `String()` method, and unit tests all exist in
`skillsdir.go:25-50` but are dead code in the binary.

**Fix**: In the `c.path` branch (`main.go:268-281`), bind `src` instead of `_`
and print the label to **stderr** after the stdout dir print:
```go
dir, src, err := skillsdir.Find()
...
fmt.Fprintln(stdout, dir)                       // stdout stays byte-exact (§13)
fmt.Fprintf(stderr, "(found via %s)\n", src)     // NEW: rule label to stderr
```
The labels already exist: `"SKPP_SKILLS_DIR"`, `"sibling of binary"`,
`"ancestor of cwd"`.

**Test impact** (CRITICAL — existing tests assert current behavior):
- `main_test.go:169 TestRunPathSuccess`: asserts `errOut.Len() != 0 → error`.
  Must be updated to expect `(found via SKPP_SKILLS_DIR)\n` on stderr.
- `main_test.go:187 TestRunPathShortFlag`: should also verify stderr label.
- New test: assert the source label matches the winning rule for each Source
  (env / sibling / walk-up). The sibling/walk-up cases are hard to test in
  `run()` (binary path / cwd dependent); the env case via `t.Setenv` is the
  deterministic one to assert.

---

## Issue 2 (MINOR) — Unicode tags/descriptions misalign the table

**Root cause**: `ui.go:132 padRight` uses `len(s)` (byte length). Multi-byte
UTF-8 runes (é=2 bytes, —=3 bytes) are under-counted for display width.
Also affected: column-width computation `ui.go:79-82` (`len(s.RelTag)`,
`len(name)`) and `wrapWords:143` (`len(cur)+1+len(word)`).
The false ASCII comment is at `ui.go:128-131`.

**Key insight**: `RelTag` and `Name` ARE charset-constrained (check.go
`validName` regex `^[a-z0-9]+(-[a-z0-9]+)*$`), so byte==display for those.
`Description` is free-form prose (only length-checked, no charset rule) —
it is the realistic misalignment vector.

**Fix**: Add a stdlib `displayWidth(s string) int` helper using
`utf8.RuneCountInString(s)` (no new dependency — respects §4/§7.3 policy).
Replace `len()` calls in `padRight`, the width computation, and `wrapWords`.
Fix the false comment. Note limitation: wide CJK runes (display width 2) are
not fully handled by rune count, but the common case (é, —, smart quotes,
emoji as 1-wide) is fixed.

**Test impact**:
- `ui_test.go:153 TestPrintListColumnsAlignedAcrossRows` (ASCII) — still passes.
- New test: multi-byte tag/description, assert column alignment via `colOf()`.

---

## Issue 3 (MINOR) — `.gitignore` deviates from PRD §16

**Current** `.gitignore` has 4 extra entries: `/build`, `.env`, `.env.*`,
`.pi-subagents/`. PRD §16 specifies exactly: `/skpp`, `/dist`, `*.test`,
`*.out`, `.DS_Store`.

**Fix**: Trim to the §16 spec set (5 entries). PRD.md cannot be modified, so
implementation must conform to the spec.

**Test impact**: None (no code).

---

## Issue 4 (MINOR) — `--search` does not match aliases/category

**Root cause**: `search.go:59 matches()` scans only RelTag, Name, Description,
Keywords. `Aliases []string` and `Category string` are on the Skill struct
(skill.go:48,47) and populated by `BuildSkill`, but ignored by `matches`.
The code comment (search.go:49-53) documents this as deliberate per §6.1.

**Spec tension**: §6.1 lists "tag, name, description, metadata.keywords" but
§10 says keywords/category/aliases "exist only to enrich `skpp --search`".
Decision: §10's intent wins — see `decisions.md`.

**Fix**: Add two checks to `matches()` after the Keywords loop:
```go
for _, a := range s.Aliases {
    if strings.Contains(strings.ToLower(a), q) { return true }
}
if strings.Contains(strings.ToLower(s.Category), q) { return true }
```

**Test impact** (CRITICAL — existing test asserts OPPOSITE behavior):
- `search_test.go:126 TestSearchDoesNotMatchCategoryOrAliases`: must be
  REPLACED with `TestSearchMatchesCategoryAndAliases` that asserts they DO
  match.

---

## Issue 5 (MINOR) — Combined short flags and `--flag=value` rejected

**Root cause**: `main.go:152` `switch a` matches exact whole tokens only.
`-vh` ≠ `-v`, `--version=x` ≠ `--version` → default branch → unknownFlag →
exit 2.

**Fix**: Normalize each token BEFORE the switch in the `parseArgs` loop:
1. Long `--flag=value`: split on first `=` → set flag, capture value (for
   `--search=`). Only `--search` takes a value; `--bool=x` treats x as ignored
   or could error (simplest: accept and ignore for bool flags).
2. Short bundle `-abc`: expand into `-a`, `-b`, `-c` and process each. A `-s`
   in a bundle consumes the remainder (`-squery`) or the next arg.

**Short flag set** (only these have short forms): `v h p l a f s`. `--relative`
and `--no-color` are long-only (no bundling).

**Test impact**: New tests for `-vh`, `--version=x`, `--search=foo`, `-afl`.

---

## Issue 6 (MINOR) — Combining listing modes silently picks one

**Root cause**: `exclusivityError` (main.go:484) rejects only tags+mode,
check+tags, check+mode. It does NOT reject mode+mode (e.g. `--list --search`).
Dispatch order (path > list > search > check > all) silently picks the first.

Also: stale comment at main.go:279-281 says dispatch order is
`check → path → list → search → all → tags` but `check` is actually step 8
(after path/list/search). Harmless today (check is guaranteed standalone by
exclusivity) but misleading.

**Fix**: Extend `exclusivityError` with a 4th family: any combination of two+
listing modes (`path`, `list`, `searchMode`, `all`) is an error (exit 2).
Fix the stale dispatch-order comment to match actual code order.

**Test impact**: New tests that `--list --search`, `--all --list`, `--path --list`
exit 2. Existing exclusivity tests unaffected (they test tags+mode etc.).

---

## Issue 7 (MINOR) — check cannot report "skill dir has no SKILL.md"

**Root cause**: `discover.Index` (index.go:46) defines a skill as "any directory
directly containing SKILL.md" — grouping dirs without SKILL.md are never
indexed, never checked. `check.go:123-129` documents this as deliberate (a
heuristic would false-positive on legitimate grouping dirs).

**Fix**: No code change required (PRD says "None required for correctness").
The existing comment at check.go:123-129 already documents the rationale.
Action is documentation-only: verify the comment accurately reflects §9 and
note the deliberate reframing ("invalid SKILL.md frontmatter" IS the reframed
§9 rule — see check.go:150-151).

**Test impact**: None. If touched, ensure §13 acceptance (`skpp check` reports
example as OK) still passes.
