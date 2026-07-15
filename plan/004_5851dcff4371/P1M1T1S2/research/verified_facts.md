# Verified Facts — P1.M1.T1.S2 (exclusivityError: messages → --flags + mode-set consistency)

Confirmed against the **live** `main.go`. CRITICAL STATE CHANGE during research:
the parallel sibling **S1 (P1.M1.T1.S1) was COMMITTED as HEAD `594be07`
"Replace bare subcommands with flags in parseArgs"** while this research ran.
So S2 now implements against the **post-S1** tree (1275 lines), not the pre-S1
tree the original contract/change-map line numbers were measured against.

---

## 0. Line numbers are STALE — locate by CONTENT, never by line number

Three different line-number sets exist for the SAME six messages. ALL are wrong
for S2 except the post-S1 current set:

| Source | exclusivityError `func` | check msgs | init msgs | completion msgs | era |
|---|---|---|---|---|---|
| contract (tasks.json) / change-map (f30d5c5) | 782 | 804/807 | 815/818 | 825/829 | pre-S1, vs old commit |
| my first grep (55ada20, pre-S1-commit) | 757 | 777/780 | 789/792 | 800/803 | pre-S1 |
| **CURRENT post-S1 (594be07) — AUTHORITATIVE** | **769** | **789/792** | **801/804** | **812/815** | **S2's input** |

**S2 MUST locate `func exclusivityError(c config) (bad bool, msg string)` and edit
by matching the exact current text (given verbatim in §1 below), NOT by line
number.** The function closes with `return false, ""` + `}` (currently line 818).

---

## 1. The CURRENT (post-S1, authoritative) exclusivityError body — exact text

Doc comment (750-768) + function (769-818). The six `cannot be combined` messages
and the three mode-set conditions, transcribed verbatim from HEAD `594be07`:

### Listing-modes block (UNCHANGED by S2 — no bare check/init/completion words)
```go
	n := 0
	for _, b := range []bool{c.path, c.list, c.searchMode, c.all} {
		if b { n++ }
	}
	if n >= 2 {
		return true, "skilldozer: listing modes --path/--list/--search/--all are mutually exclusive"
	}
	hasTags := len(c.tags) > 0
	if hasTags && (c.path || c.list || c.searchMode || c.all) {
		return true, "skilldozer: tags cannot be combined with --path/--list/--search/--all"
	}
```
↑ These two returns contain NO bare `check`/`init`/`completion` — S2 does NOT
touch them. (The `tags cannot be combined` message is asserted by
`TestExclusivityErrorTagsAndPath`, which is a direct-call test that PASSES today
and must keep passing.)

### check block (TWO edits: message + add c.completion to set+message)
```go
	if c.check && hasTags {
		return true, "skilldozer: 'check' cannot be combined with tag arguments"          // [789] → '--check'
	}
	if c.check && (c.path || c.list || c.searchMode || c.all) {                            // [791] add c.completion
		return true, "skilldozer: 'check' cannot be combined with --path/--list/--search/--all"  // [792] → '--check' + add --completions
	}
```

### init block (comment flag-language + message + add c.completion to set+message)
```go
	// init is its own exclusive mode (PRD §6.3 / §8.2: like `check`). It rejects the     // [794] → --init/--check
	// listing/inspection modes AND stray tags. A single positional <dir> after `init`    // [795] → --init
	// is consumed as the store (c.initStore) by parseArgs, so it never reaches           // [796]
	// c.tags; a SECOND positional, or any positional after `init --store`, lands in       // [797] → --init --store
	// c.tags and is rejected here as a stray.
	if c.init {
		if hasTags {
			return true, "skilldozer: 'init' cannot be combined with tag arguments"        // [801] → '--init'
		}
		if c.check || c.list || c.searchMode || c.all || c.path {                           // [803] add c.completion
			return true, "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"  // [804] → '--init' ... --check/--completions/...
		}
	}
```

### completion block (comment flag-language + message; set ALREADY complete)
```go
	// completion is its own exclusive mode (PRD §6.3 / §14.6: like check/init). It rejects the other  // [806] → --completions/--check/--init
	// modes/subcommands AND stray tags. `completion` does no positional capture (mirrors check), so    // [807] → mode flags; --completions; --check
	// any positional after it lands in c.tags and is rejected here as a stray.
	if c.completion {
		if hasTags {
			return true, "skilldozer: 'completion' cannot be combined with tag arguments"   // [812] → '--completions'
		}
		if c.check || c.init || c.list || c.searchMode || c.all || c.path {                  // [814] ALREADY complete — NO change
			return true, "skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all"  // [815] → '--completions' ... --check/--init/...
		}
	}
	return false, ""                                                                        // [818]
```

---

## 2. The six exact message rewrites (old → NEW)

| # | Block | OLD (current) | NEW (target — copy verbatim) |
|---|---|---|---|
| 1 | check+tags | `skilldozer: 'check' cannot be combined with tag arguments` | `skilldozer: '--check' cannot be combined with tag arguments` |
| 2 | check+modes | `skilldozer: 'check' cannot be combined with --path/--list/--search/--all` | `skilldozer: '--check' cannot be combined with --completions/--path/--list/--search/--all` |
| 3 | init+tags | `skilldozer: 'init' cannot be combined with tag arguments` | `skilldozer: '--init' cannot be combined with tag arguments` |
| 4 | init+modes | `skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check` | `skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path` |
| 5 | completion+tags | `skilldozer: 'completion' cannot be combined with tag arguments` | `skilldozer: '--completions' cannot be combined with tag arguments` |
| 6 | completion+modes | `skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all` | `skilldozer: '--completions' cannot be combined with --check/--init/--path/--list/--search/--all` |

**Message ordering note (inherited from contract; do NOT "normalize"):** the three
mode-list messages intentionally differ in flag order (#2 lists `--completions`
first; #4 is contract-pinned as `--check/--completions/--list/--search/--all/--path`;
#6 is `--check/--init/--path/--list/--search/--all`). These are hints tested only
by substring, and the contract pins #4 verbatim — copy each NEW string exactly.

---

## 3. The mode-set condition changes

| Block | OLD condition | NEW condition | Why |
|---|---|---|---|
| check+modes | `c.check && (c.path || c.list || c.searchMode || c.all)` | `c.check && (c.completion || c.path || c.list || c.searchMode || c.all)` | contract 3b: add c.completion so check+completion is caught HERE (not only by the completion block) |
| init+modes | `c.check || c.list || c.searchMode || c.all || c.path` | `c.check || c.completion || c.list || c.searchMode || c.all || c.path` | contract 3b/3c: add c.completion (init-block was the one the contract named explicitly) |
| completion+modes | `c.check || c.init || c.list || c.searchMode || c.all || c.path` | **UNCHANGED** | already complete (catches completion + every other mode) |

**Symmetry result after the two additions:** each of the three exclusive modes
(check/init/completion) now catches its combination with `completion` in its OWN
block, so the fired message names the actual pair the user typed. Before, only the
completion block did, so `--check --completions` printed the *completion* message;
now it prints the *check* message (still exit 2 — only the wording's source
changes). This is the "consistency" the contract requests.

**IMPORTANT — no double-handling bug:** adding c.completion to the check/init
blocks does NOT create a gap or double-exit. check+completion and init+completion
are caught EARLIER (check block < init block < completion block in source order),
so they now fire the check/init message instead of the completion message. Every
pair still returns `(true, msg)` exactly once. Verified by reading source order.

---

## 4. Comment sites needing flag-language (contract 3d)

S2 owns the **entire** exclusivityError function INCLUDING its doc comment and the
two inline comments (the change map's 6d lists the exclusivityError comment site as
S2's territory — confirmed by the S1 PRP GOTCHA #8 "main.go:821 exclusivityError
comment → P1.M1.T1.S2"). T2.S1 owns the OTHER comment sites (631, 1191, etc.).

Sites (confirmed via `awk` against the current 750-818 range):
- Doc comment: "`check` ignores tags" / "`check` is NOT in the listing-mode set" /
  "check+mode", "check+path", "check+list/check+search/check+all" → bare `check`
  becomes `--check`. (The word "modes" STAYS — PRD §6.3 legitimately says "mode
  flags"; the contract only forbids "subcommands".)
- init inline comment: "init is its own exclusive mode ... like `check`" / "after
  `init`" / "after `init --store`" → `--init` / `--check`.
- completion inline comment: "**modes/subcommands**" → "**mode flags**" (the ONE
  "subcommand" instance in the function); "like check/init" → "like --check/--init";
  "`completion` ... mirrors check" → "`--completions` ... mirrors --check".

---

## 5. Test impact — S2's gate is `go build` + `go vet` (NOT `go test`)

- The DIRECT-call exclusivity tests (`TestExclusivityErrorListingModes` ~2350,
  `TestExclusivityErrorTagsAndPath` ~2390) call `exclusivityError` with hand-built
  `config` structs and assert substrings. S2 does NOT change the listing-modes
  logic or the `tags cannot be combined` message, so these KEEP PASSING. (Neither
  is in the current FAIL list.)
- The RUN-level exclusivity tests (`TestRunExclusivityCheckAndTags`, etc.) pass bare
  words (`run([]string{"check", "foo"})`) through `parseArgs`. After S1 (committed),
  bare `check` is a TAG, so these fail at the exit-code assertion — that is S1's
  TDD drift, fixed by **P1.M1.T3**, NOT S2. S2 must NOT touch main_test.go.
- S2's string changes do NOT break the run-level substring assertions even if they
  ran: new messages still contain `check`/`init`/`completion` substrings
  (`--check` ⊃ "check", `--completions` ⊃ "completion", `--init` ⊃ "init"). So S2
  introduces zero NEW test red beyond S1's existing drift.
- **Gate:** `go build ./...` + `go vet ./...` must pass. `go test ./...` is
  EXPECTED RED (S1's drift; T3's scope). go.mod/go.sum byte-for-byte unchanged
  (string-literal + bool-OR edits add no imports).

## 6. S2's contract is build-only ("go build ./... must succeed" — OUTPUT §4)

The contract's OUTPUT §4 literally says "go build ./... must succeed." It does NOT
list `go test`. This matches S1's TDD stance. S2 verifies its own contract with a
throwaway direct-call probe (see PRP Level 2) that is deleted after — it does NOT
modify main_test.go.

## 7. Consumers / dispatch unchanged

`run()` calls `exclusivityError(c)` at main.go:494 (pre-dispatch) and dispatches
c.check→runCheck / c.init→runInit / c.completion→runCompletion. S2 changes only the
STRINGS exclusivityError returns and two boolean conditions (adding c.completion).
The `(bad bool, msg string)` signature, the return COUNT, the families' ORDER, and
the exit-2 semantics are all UNCHANGED. No caller is affected.
