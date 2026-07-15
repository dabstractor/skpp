# Verified Facts — P1.M2.T2.S1 (Issue 4: POSIX `--` end-of-options separator)

Researched against the LIVE codebase (main.go parseArgs read at the cited line
ranges; main_test.go templates read; issue_analysis.md §Issue 4 + decisions.md §D6
read in full). The parallel sibling P1.M2.T1.S1 (Issue 3, missing-value handling)
PRP was read as a CONTRACT — it edits parseArgs' `--search`/`--shell` cases
(~288-326) + expandShortBundle (~444); this subtask edits the TOP of the parseArgs
loop body (~190) + a var before the loop (~185) — DISJOINT regions, no collision.

---

## §1. The bug — bare `--` is classified as an unknown flag (exit 2)

**Repro (issue_analysis.md §Issue 4, line 247):** `./skilldozer -- -x` →
`skilldozer: unknown flag '--'`, exit 2. A skill whose tag begins with `-`
(e.g. `-foo`) is therefore impossible to address.

**Root cause** — parseArgs (main.go:184-360) has NO `--` handling. A bare `--` token:
1. fails the `=`-form guard (main.go:200: no `=` in `--`).
2. fails the short-bundle guard (main.go:262: `len("--")==2`, not `> 2`).
3. reaches `switch a` → `default:` (main.go:351), where `strings.HasPrefix("--", "-")`
   (main.go:357) is true → `c.unknownFlag = "--"` (main.go:359).
4. run() prints `unknown flag '--'` to stderr, exit 2.

There is NO end-of-options semantics anywhere in parseArgs.

---

## §2. The fix — an `endOfOpts` loop flag (decisions.md §D6, ACCEPTED)

decisions.md §D6 (line 76) chose the **loop-flag** approach over early-return:

> "Use a boolean `endOfOpts` flag in the parseArgs loop... Placed before ALL token
> classification (before the `=` check, before the short-bundle check, before the
> switch). Clean, extensible, and keeps the loop structure intact."

REJECTED alternative: early-return (`if a == "--" { collect rest as tags; return c }`)
— "breaks the uniform loop structure and skips any post-loop logic."

**The exact insertion** (issue_analysis.md line 262 transcribes it verbatim):

```go
func parseArgs(args []string) config {
	var c config
	endOfOpts := false  // NEW (Issue 4)
	for i := 0; i < len(args); i++ {
		a := args[i]
		if endOfOpts {           // NEW: everything after -- is positional
			c.tags = append(c.tags, a)
			continue
		}
		if a == "--" {           // NEW: end-of-options separator
			endOfOpts = true
			continue
		}
		// ... existing =-form (line 200), short-bundle (262), and switch (351) ...
	}
	return c
}
```

**CRITICAL ORDERING:** the `if endOfOpts` guard MUST come BEFORE the `if a == "--"`
guard, so that a SECOND `--` (e.g. `skilldozer -- --`) is appended as a positional
tag named "--" (POSIX-correct: once end-of-options is set, even `--` is a positional).
If the order were reversed, the second `--` would re-trigger the separator and never
become a tag. (Traced in §4.)

---

## §3. Placement anchors (verified-current line numbers)

- `func parseArgs(args []string) config {` — main.go:184
- `var c config` — main.go:185  ← PLACE `endOfOpts := false` immediately AFTER this
- the loop comment + `for i := 0; i < len(args); i++ {` — main.go:186-189
- `a := args[i]` — main.go:190  ← PLACE the two guards immediately AFTER this
- the `=`-form check `if strings.HasPrefix(a, "--") && strings.Contains(a, "=") {`
  — main.go:200 (the contract cited 202; drift is harmless — the guards go BEFORE it)
- the short-bundle guard `len(a) > 2 && a[0] == '-' && a[1] != '-'` — main.go:262
  (contract cited 259)
- `switch a` → `default:` — main.go:351; `strings.HasPrefix(a, "-")` @357;
  `c.unknownFlag = a` @359 (the path `--` currently falls into)

The guards are the FIRST checks in the loop body, before any classification. No
existing token-classification line is edited — this is purely INSERTION (2 lines
before the loop + ~10 lines at the top of the loop body + a doc comment).

---

## §4. The 4-case trace (every following-token shape stays correct)

| input | trace | result |
|---|---|---|
| `-- -x` | i=0 a="--": endOfOpts=false→guard2 sets endOfOpts=true. i=1 a="-x": guard1 appends. | tags=["-x"], unknownFlag="" ✓ |
| `-- mytag` | i=0 "--"→sep; i=1 "mytag"→guard1 append. | tags=["mytag"] ✓ |
| `--list -- --check` | i=0 "--list"→switch case --list→c.list=true. i=1 "--"→sep. i=2 "--check"→guard1 append. | list=true, tags=["--check"] ✓ |
| `-- --` | i=0 "--"→sep (endOfOpts=true). i=1 "--": guard1 (endOfOpts true) appends. | tags=["--"] ✓ (second -- is a positional tag — POSIX) |

The guard ORDER (endOfOpts BEFORE `a=="--"`) is what makes the `-- --` edge case
correct. A plain `-x` NOT after `--` still → unknown flag (unchanged: endOfOpts is
false, it skips both guards, reaches switch→default→unknownFlag).

---

## §5. Zero breakage — grep-confirmed (no test asserts bare `--` → unknown flag)

Every `unknownFlag` test uses a NON-bare-dashdash token: `--frobnicate` (594, 111,
599, 301, 306), `-z` (1875), `--bogus`/`--more` (1783). NONE use a bare `--`. So the
fix is purely additive — `TestParseArgsDashedUnknownNotATag` (asserts `--frobnicate`)
and `TestRunUnknownFlagExits2` (asserts `-z`) stay GREEN unchanged. issue_analysis.md
confirms this explicitly. Adding `--` handling changes ONLY the bare-`--` path.

---

## §6. Boundary with the parallel sibling P1.M2.T1.S1 (Issue 3) — no collision

P1.M2.T1.S1 (read its PRP) edits, in main.go parseArgs:
- `config` struct: +`searchMissingValue`/`shellMissingValue` after `storeMissingValue` (~169).
- the `--search`/`-s` main-switch case (~288-298): +`else { c.searchMissingValue = true }`.
- the `--shell` main-switch case (~320-326): +`else { c.shellMissingValue = true }`.
- expandShortBundle -s default (~444): +`c.searchMissingValue = true`.
- run(): +2 peer checks after storeMissingValue (~499).
- main_test.go: rename/flip/add search/shell tests.

This subtask (Issue 4) edits, in main.go parseArgs:
- `endOfOpts := false` AFTER `var c config` (~185).
- the two guards AFTER `a := args[i]` (~190), BEFORE the =-form check (~200).
- main_test.go: +5 `--` tests.

**The regions are DISJOINT** in both files: this subtask touches the TOP of the loop
body + the var-before-loop (lines ~185-191); the sibling touches the cases deep in
the switch (~288-326) + expandShortBundle (~444) + run() + the config struct. No
text-level overlap; the changesets compose in either order. Both are purely additive
insertions into different parts of parseArgs.

**Behavioral interaction (correct, no conflict):** if a `--search` token appears
AFTER `--`, this subtask's guard1 routes it to `c.tags` (as a positional) BEFORE it
reaches the `--search` case — correct POSIX semantics (`--` means everything after
is positional). The sibling's missing-value detection only fires for `--search` NOT
after `--`. No contradiction.

---

## §7. Test design — 5 tests (4 contract OUTPUT + 1 POSIX edge case)

**Templates (read at exact lines):**
- parseArgs-level: `TestParseArgsDashedUnknownNotATag` @594 (assert c.tags/c.unknownFlag directly).
- run-level: `TestRunTagAtomicityUnknownPrintsNothing` @667 (sampleStore + SKILLDOZER_SKILLS_DIR +
  run returns int; assert code, empty stdout, stderr Contains the tag).

**Helpers:** `sampleStore(t)` (a store with example + writing/reddit skills, used by run tag tests);
`unsetSkillsEnv(t)` @28 (neutralizes env for determinism). `bytes.Buffer`, `strings.Contains`,
`strings`/`os`/`path/filepath` all imported in main_test.go.

The 5 tests (matching the contract OUTPUT list + the §4 `-- --` edge case):
1. `TestParseArgsDashDashSeparator` — `parseArgs(["--","-x"])` → tags==["-x"], unknownFlag=="".
2. `TestParseArgsDashDashBeforeTag` — `parseArgs(["--","mytag"])` → tags==["mytag"].
3. `TestParseArgsDashDashWithFlags` — `parseArgs(["--list","--","--check"])` → list==true, tags==["--check"].
4. `TestRunDashDashUnknownFlagStillWorks` — `run(["--","--bogus"])` with sampleStore → exit 1
   (unknown TAG "--bogus", NOT unknown flag), empty stdout, stderr Contains "--bogus",
   stderr does NOT Contain "unknown flag". (Proves the dashed token reached tag resolution.)
5. `TestParseArgsDashDashSecondDashDashIsTag` — `parseArgs(["--","--"])` → tags==["--"] (the §4
   POSIX edge case the contract explicitly highlights).

Test 4 needs a store fixture (sampleStore) so Find() succeeds and "--bogus" reaches tag
resolution (exit 1 unknown-tag). Without a store it'd be exit 1 "not configured" — still
not-exit-2, but the sampleStore version proves the full "dashed token → tag resolution"
flow, matching the contract's "resolves --bogus as a tag". The assertion that stderr does
NOT contain "unknown flag" is the load-bearing one (it distinguishes the tag path from the
unknown-flag path).

---

## §8. DOCS (Mode A) — inline comment only; README deferred

The contract DOCS: "[Mode A] Add a brief inline comment in parseArgs documenting the POSIX
`--` convention. Optionally mention in README.md Usage section... (include in the final doc
sweep task P1.M3.T1.S1)." So THIS subtask adds the inline comment (in the insertion at
§2); the README mention is P1.M3.T1.S1's Mode B sweep, NOT here. No README/help-text change.

## §9. Scope discipline (README-only is wrong — this is main.go + main_test.go)

This subtask edits main.go (parseArgs: the `endOfOpts` flag + 2 guards + comment) and
main_test.go (5 tests). It does NOT touch: config struct (no new field needed — `endOfOpts`
is a parse-local var, not a config field), run() (no change — the tags flow through
unchanged), exclusivityError, completions/*, internal/*, go.mod/go.sum, PRD.md, tasks.json.
go.mod/go.sum byte-for-byte unchanged (pure stdlib: a `==` compare + a bool + append).
