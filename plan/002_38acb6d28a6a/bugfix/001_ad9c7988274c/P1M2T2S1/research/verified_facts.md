# Verified Facts — P1.M2.T2.S1 (Issue 4: capture a duplicate `init` token as a conflict)

Researched against the LIVE codebase (main.go + main_test.go read at the cited
line ranges; architecture/decisions.md §D3 + bug_fixes_validation.md §ISSUE 4
read in full). The parallel sibling P1.M2.T1.S1 (Issue 3, tags+--path) PRP was
read as a CONTRACT — its `exclusivityError` tags-predicate edit is assumed present
(it does NOT touch `case "init":` or the init exclusivity branch, so no collision).

---

## §1. The bug — `init init` runs (exit 0) instead of erroring

**Repro (bug_fixes_validation.md §ISSUE 4, line 89):** `./skilldozer init init </dev/null`
→ exit 0, config WRITTEN. Control: `init check </dev/null` → exit 2 (correctly rejected).

**Root cause** — `parseArgs` `case "init":` (main.go:277-292). The positional-capture
guard at main.go:289-291:

```go
c.init = true                                   // line 286
if i+1 < len(args) {
    next := args[i+1]
    if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {  // line 289
        c.initStore = next                       // line 290
        i++
    }
}
```

The `&& next != "init"` clause makes the guard REFUSE to capture a following `init`
as the store (correct intent: `init` is not a dir). But because it neither captures
nor consumes the token, the for loop's own `i++` advances past the first `init`, and
the second `init` token is then processed by `case "init":` AGAIN → re-sets
`c.init=true` (idempotent), sets NO conflict flag. Result: `c.init=true, c.initStore="",
c.tags=[]` → passes exclusivity → init dispatch runs → exit 0, config written.

**Why `init check` is NOT affected** — `check` is ALSO refused by the guard
(`&& next != "check"`), so it too is neither captured nor consumed; the loop advances
and the `check` token reaches `case "check":` → `c.check=true`. Then exclusivityError's
init branch fires on `c.init && c.check` (main.go:744-746) → exit 2. So `init check`
works purely because `check` has its OWN case that sets a mode flag. `init` has no
such flag (it only re-sets `c.init`), so a duplicate `init` sets no conflict.

---

## §2. The fix — Option A (decisions.md §D3): capture the duplicate `init` into c.tags

D3 (read in full, line 21-27): "Option A (capture the duplicate reserved `init`/`check`
token into `c.tags` so the init exclusivity branch already rejects it) needs no new
field and reuses the existing `'init' cannot be combined with tag arguments` path."

**IMPORTANT SCOPE NOTE:** although D3 says "init/check", the contract LOGIC §3 and
bug_fixes_validation.md both note "`check` already works via `c.check`". So the code
change needs to handle ONLY `next == "init"` — `next == "check"` is UNCHANGED (it
already flows to `case "check":`). Splitting out the `init` case is the minimal fix.

**The change** (main.go:287-292, the `if i+1 < len(args) { … }` body). BEFORE:

```go
if i+1 < len(args) {
    next := args[i+1]
    if !strings.HasPrefix(next, "-") && next != "check" && next != "init" {
        c.initStore = next
        i++
    }
}
```

AFTER (Option A — a duplicate `init` is appended to c.tags + consumed via i++):

```go
if i+1 < len(args) {
    next := args[i+1]
    if next == "init" {
        // Issue 4: a duplicate reserved `init` token is a conflict, not a store
        // dir. Capture it as a tag so exclusivityError's init+tags branch rejects
        // `init init` with exit 2 (consistent with `init check`, where the second
        // token reaches case "check" and sets c.check). A literal store dir named
        // "init" must still use --store.
        c.tags = append(c.tags, next)
        i++
    } else if !strings.HasPrefix(next, "-") && next != "check" {
        c.initStore = next
        i++
    }
}
```

**Trace (the 4 cases that must stay correct):**
| input after `init` | next | branch taken | result |
|---|---|---|---|
| `init` | "init" | NEW first `if` | c.tags=["init"], i++ → hasTags → exit 2 ✓ (Issue 4 fix) |
| `check` | "check" | else-if: `!= "check"` FALSE → skip | left for case "check" → c.check → exit 2 ✓ (unchanged) |
| `--store` | "--store" | else-if: HasPrefix("-") TRUE → skip | left for case "--store" ✓ (unchanged) |
| `/tmp/x` | "/tmp/x" | else-if: both TRUE → capture | c.initStore="/tmp/x", i++ ✓ (unchanged) |

**The redundant `&& next != "init"` is DROPPED from the else-if** because the first
`if next == "init"` exclusively handles that value (the else-if only runs when the
first is false, so next can never be "init" there). Logically equivalent, cleaner.
(Alternative: keep the redundant clause for defensive self-containment — both are
correct; dropping is chosen for clarity. See PRP design decision #3.)

---

## §3. How exclusivity catches it (no new family, no new field)

`exclusivityError` init branch (main.go:741-746, read in full):

```go
if c.init {
    if hasTags {                                                              // line 742
        return true, "skilldozer: 'init' cannot be combined with tag arguments"  // line 743
    }
    if c.check || c.list || c.searchMode || c.all || c.path {                 // line 744
        return true, "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"
    }
}
```

`hasTags := len(c.tags) > 0` (main.go:726). With Option A, `init init` produces
`c.tags=["init"]`, so `hasTags` is true → the EXISTING line 743 fires → exit 2.
**No change to exclusivityError is needed** — the fix is entirely in `case "init":`.
This is why Option A was chosen over Option B (a counter field): it reuses the
already-tested init+tags path (the same path `init foo bar` uses — see §4).

---

## §4. run() precedence — exclusivity fires BEFORE config write (config NOT written)

run() dispatch order (main.go:438-472, read in full):
1. `if c.help { … }`
2. `if c.version { … }`
3. `if c.unknownFlag != "" { … exit 2 }`
4. `if c.storeMissingValue { … exit 2 }` (P1.M1.T2.S2, line ~450)
5. `if bad, msg := exclusivityError(c); bad { … exit 2 }` (line ~460) ← fires here
6. `if c.init { return runInit(c, stdout, stderr) }` (line ~472) ← never reached

`runInit` (main.go:1014) is the ONLY place config is written (resolveStore →
config.Path → config.Save, main.go:953/966). So exit 2 at step 5 structurally
GUARANTEES the config is NOT written. This is the contract OUTPUT requirement
"the config is NOT written (exclusivityError fires before init dispatch)".

For `init init`: storeMissingValue is NOT set (no `--store` token), so it skips
step 4 and hits exclusivity at step 5 → exit 2. runInit never runs. ✓

---

## §5. Zero breakage — grep-confirmed, no existing test asserts init-init

`grep -n '"init", "init"\|init init\|InitInit\|DuplicateInit\|SecondInit' main_test.go`
→ **(none)**. The fix is purely additive: today NO test covers `init init`, and the
change ONLY affects the `next == "init"` sub-case (the other 3 cases — check, --store,
positional dir — are byte-identical in behavior, traced in §2). Regression coverage
for the unchanged cases is ALREADY present:
- `init <dir>`: TestParseArgsInitPositionalDir (1278) + TestParseArgsInitDirNotCapturedAsTag (1384).
- `init check`: TestRunExclusivityInitAndCheck (1837) — the CONTROL test.
- `init --store <dir>`: TestParseArgsInitStoreLongForm (1292) + EqualsForm (1309).

The exact tags-message string `'init' cannot be combined with tag arguments` is
asserted by NO test today (grep-confirmed: the InitAnd* tests use
`Contains("init")`, not the exact message), so the existing tests stay green and the
new test can assert the exact wording or `Contains("init")` freely.

---

## §6. Test design — 1 parseArgs-level + 1 run-level

bug_fixes_validation.md §ISSUE 4 (line 112) prescribes: "Add `TestParseArgsInitInitDoesNotSwallow`
and a run-level `TestRunExclusivityInitInit` asserting exit 2 + empty stdout + config
NOT written (mirror `TestRunExclusivityInitAndCheck`)."

**Template sources (read at exact line ranges):**
- `TestRunExclusivityInitAndStrayTag` (main_test.go:1877) — the EXACT run-level template:
  `init foo bar` → foo=initStore, bar=tags → exit 2 via init+tags (the SAME path my fix
  uses). Its assertions: code==2, empty stdout, `Contains("tag")`.
- `TestRunExclusivityInitAndCheck` (1837) — the control (init+check → exit 2).
- `TestParseArgsInitSubcommand` (1261) + `TestParseArgsInitPositionalDir` (1278) — the
  parseArgs-level template (assert c.init, c.tags, c.initStore fields directly).

**The 2 new tests:**
1. `TestParseArgsInitInitCapturedAsTag` (parseArgs-level) — locks the core fix:
   `parseArgs(["init","init"])` → c.init=true, c.tags==["init"], c.initStore=="".
2. `TestRunExclusivityInitInit` (run-level) — locks the contract OUTPUT:
   `run(["init","init"])` → exit 2, empty stdout, stderr `Contains("init")`, AND the
   config file (SKILLDOZER_CONFIG → temp path) is NOT written. The config-not-written
   check uses `t.Setenv("SKILLDOZER_CONFIG", <temp>)` + `os.Stat` asserting IsNotExist
   (exclusivity runs before runInit/config.Save, so this always holds — the assertion
   makes the contract guarantee explicit and regression-proof).

NO store fixture / SKILLDOZER_SKILLS_DIR / t.Chdir / unsetSkillsEnv needed for the
exclusivity assertion (exclusivity runs before skillsdir.Find()). The
config-not-written check adds ONLY a SKILLDOZER_CONFIG t.Setenv + an os.Stat — no
store tree.

---

## §7. Boundary with the parallel sibling P1.M2.T1.S1 (Issue 3, tags+--path) — no collision

P1.M2.T1.S1 (read its PRP in full) edits:
- `exclusivityError` tags predicate (main.go:727) + its message (728) + a doc-comment bullet.
- 3 NEW tests: TestRunExclusivityTagsAndPath, TestRunExclusivityPathAndTag, TestExclusivityErrorTagsAndPath.

It does NOT touch:
- `parseArgs` `case "init":` (main.go:277-292) — THIS subtask's edit site.
- `exclusivityError` init branch (main.go:741-746) — THIS subtask's consumer (unchanged).
- Any InitAnd* test or parseArgs Init test.

So the two changesets are DISJOINT in both files (different regions, different test
functions). They land in either order. The only shared line region is the
`exclusivityError` func, but Issue 3 edits the TAGS family (~727) while Issue 4 reads
(consumes, does not edit) the INIT family (~741) — no text overlap.

---

## §8. DOCS (Mode A) — update the `case "init":` comment only

The contract DOCS: "[Mode A] Update the `case "init":` comment (main.go:278-285) to
note a duplicate init token is now treated as a conflict. No README surface change —
covered by the final Mode B sweep (P1.M3.T1)."

The current comment (main.go:278-285) says "A following flag (`init --store …`) or
subcommand (`init check`) is left for its own case so exclusivity can flag the conflict.
GOTCHA: a store literally named `check`/`init` must be passed via `--store`."

UPDATE: add that a duplicate `init` is now captured as a tag (a conflict), while
`init check` remains left for case "check". The GOTCHA about a literal `init`-named
store dir still holds (must use --store). No README change (Mode B = P1.M3.T1).
