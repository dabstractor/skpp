# PRP — P1.M1.T1.S2: `exclusivityError` — rewrite messages to `--flags` + mode-set consistency

> **Subtask:** P1.M1.T1.S2 — the second half of P1.M1.T1 (the `exclusivityError` follow-up to S1's parseArgs flag conversion). Implements PRD §6.3 / §17 / decision 19: since `check`/`init`/`completion` are now `--check`/`--init`/`--completions` **flags** (S1, committed `594be07`), the exclusivity error messages must say so, and the three exclusive-mode blocks must be symmetric in which combinations they catch.
> **Scope boundary:** Edits ONLY `exclusivityError` (the function body + its doc comment + the two inline comments) in `main.go`. Does NOT touch `parseArgs`, `usageText`, `run()`, `skillsdir`, the completion functions, or `main_test.go`. The bool-driven **gating logic is unchanged** (same families, same order, same exit-2 semantics); only six message strings, two boolean conditions (each gains `c.completion`), and comment wording change.
> **Gate:** `go build ./...` + `go vet ./...` (the contract's OUTPUT §4 says "go build must succeed"; `go test` is EXPECTED RED from S1's bare→flag drift, fixed by P1.M1.T3 — do NOT touch tests).

---

## Goal

**Feature Goal**: Make `exclusivityError`'s six `cannot be combined` messages reference the `--check`/`--init`/`--completions` **flags** (not the now-deleted bare subcommands), and make the three exclusive-mode blocks (check / init / completion) **symmetric** by adding `c.completion` to the check-block and init-block mode sets so each block catches its own `+ completion` combination.

**Deliverable**: Edits to `main.go` only (no new files):
1. **Six message rewrites** — `'check'`→`'--check'`, `'init'`→`'--init'`, `'completion'`→`'--completions'`, and the init/completion mode-lists rewritten to name the excluded flags consistently.
2. **Two condition additions** — add `c.completion` to the check-block mode set and the init-block mode set (the completion-block set is already complete — unchanged).
3. **Comment flag-language** — the `exclusivityError` doc comment + the init/completion inline comments: bare `check`/`init`/`completion` → `--check`/`--init`/`--completions`, and the one `modes/subcommands` phrase → `mode flags`.

**Success Definition**: `go build ./...` succeeds; `go vet ./...` passes; `go.mod`/`go.sum` unchanged. After the change: `exclusivityError(config{check:true, completion:true})` returns `(true, msg)` whose message contains `--check`; same for init+completion (contains `--init`); every mode pair is still caught exactly once. (The direct-call unit tests `TestExclusivityErrorListingModes`/`TestExclusivityErrorTagsAndPath` keep passing; the run-level exclusivity tests stay RED from S1's drift — T3's scope.)

---

## User Persona (if applicable)

Not applicable at runtime — this is an internal error-message + consistency refactor. The user-visible consequence (PRD §6.4): when a user types `skilldozer --check --list`, stderr now says `'--check' cannot be combined with --completions/--path/--list/--search/--all` (flag language matching the flag they typed), not `'check' cannot be combined ...` (which references a bare subcommand that no longer exists). The messages ARE the user-facing doc for the exit-2 contract (Mode A — no separate README change; that's P1.M2.T2.S1).

---

## Why

- **PRD §6.3 + §17 + decision 19**: "There are no bare-word subcommands. Every non-tag action is a `--flag`." S1 (committed `594be07`) already drove `check`/`init`/`completion` → `--check`/`--init`/`--completions` in `parseArgs`. `exclusivityError`'s messages still quote the deleted bare words (`'check'`, `'init'`, `'completion'`) — a stale reference that would confuse a user who just typed a flag and is told the conflict is with a "subcommand" that does not exist. This is the message-side cleanup that completes S1's conversion.
- **Mode-set consistency (contract 3b)**: today the **completion** block is the only one that catches `+ completion` pairs (check+completion and init+completion both fall through to the completion block). The contract asks the check and init blocks to catch their own `+ completion` combination, so the fired message names the mode the user actually led with. Symmetric, predictable wording.
- **Closing Change Group 3** (code_change_map.md): this is the exact, bounded slice the architecture map assigns to S2 — six strings, two conditions, comments. No logic, no signature, no dispatch change.

---

## What

`exclusivityError(c config) (bad bool, msg string)` keeps its signature, its four families, their check order, and its `(true, msg)` / `(false, "")` return contract. Only the contents change:

| Change | Where | What |
|---|---|---|
| 6 message strings | check/init/completion tag-args + mode-list returns | bare words → `--flags`; mode-lists rewritten |
| 2 conditions | check-block mode set, init-block mode set | each gains `|| c.completion` |
| doc comment + 2 inline comments | exclusivityError's own comments | `check`/`init`/`completion`→`--flags`; `modes/subcommands`→`mode flags` |

The two listing-mode returns at the top (`listing modes ... mutually exclusive` and `tags cannot be combined with ...`) contain **no** bare check/init/completion words and are **NOT** touched.

### Success Criteria

- [ ] `func exclusivityError` returns six messages that say `'--check'`/`'--init'`/`'--completions'` (no bare `'check'`/`'init'`/`'completion'`)
- [ ] init-block message is EXACTLY `skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path` (contract 3c, verbatim)
- [ ] check-block condition includes `c.completion`; its message lists `--completions`
- [ ] init-block condition includes `c.completion`; its message lists `--completions` (+ `--check`)
- [ ] completion-block condition UNCHANGED (already complete); message rewritten to `--completions`/`--check`/`--init` flag form
- [ ] the two listing-mode returns (top of function) UNCHANGED
- [ ] exclusivityError doc comment + init/completion inline comments use flag language; the `modes/subcommands` phrase is gone
- [ ] `go build ./...` succeeds; `go vet ./...` passes; `go.mod`/`go.sum` unchanged
- [ ] S2 does NOT touch `parseArgs`, `usageText`, `run()`, `skillsdir`, completion functions, or `main_test.go`

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact current (post-S1, HEAD `594be07`) text of every edited line is transcribed verbatim in `research/verified_facts.md` §1, with the exact NEW target strings in §2/§3 and the comment sites in §4. An implementer who has never seen this repo can do it in one pass by matching the given old→new text blocks (no line-number guessing needed).

### Documentation & References

```yaml
# MUST READ — the parallel sibling PRP (S1, ALREADY COMMITTED as 594be07) — the input tree S2 builds on
- file: plan/004_5851dcff4371/P1M1T1S1/PRP.md
  why: "S1 converted parseArgs bare cases → --check/--init/--completions flags and is COMMITTED (HEAD 594be07; main.go 1275 lines). S2 is the message-side completion of that conversion. S1's GOTCHA #8 explicitly reserves the exclusivityError comment for S2. Match S1's gate stance (build+vet, NOT go test — T3 flips the tests)."
  pattern: "S1 established: gate = go build + go vet; go test is EXPECTED RED (bare→flag drift); verify with a throwaway direct-call probe deleted after. S2 follows the same stance."

# MUST READ — the authoritative current text + exact old→new strings (verified against live HEAD 594be07)
- file: plan/004_5851dcff4371/P1M1T1S2/research/verified_facts.md
  why: "§0 proves the contract/change-map line numbers (782-835 / 804-829) are STALE (pre-S1); the function is now at 769-818 — LOCATE BY CONTENT. §1 gives the exact current body. §2 the six exact message rewrites. §3 the two condition changes (and why the completion-block needs NONE). §4 the comment sites. §5/§6 the test-impact + build-only gate."
  critical: "§0 (line-number drift) and §3 (no double-handling: adding c.completion to check/init blocks only changes WHICH message fires, never whether a pair is caught) are the two facts that prevent the most likely implementation errors."

# MUST READ — the change map (Change Group 3 is this subtask's spec)
- file: plan/004_5851dcff4371/architecture/code_change_map.md
  why: "Change Group 3 pins the six old→new messages and the init-block mode-set consistency fix. NOTE: its line numbers (804-829) are pre-S1 and WRONG now (see verified_facts §0) — use its MESSAGE TEXT, not its line numbers."
  section: "Change Group 3 (exclusivityError messages + mode-set consistency)"

# MUST READ — the edit site (the ONLY file S2 touches)
- file: main.go
  why: "exclusivityError @ 769-818 (doc comment 750-768). run() calls it @ 494 (pre-dispatch). Edit the function body + its comments ONLY."
  pattern: "Match the surrounding comment style (PRD-section citations, `//` prose explaining WHY a family exists). The function already documents its four families and the Issue-6/Issue-3/N1 history — preserve that, just fix the bare-word references."
  gotcha: "Do NOT renumber/reorder the families or change return counts. Do NOT touch the two top listing-mode returns. Do NOT touch run()'s call site."

# READ-ONLY — the PRD sections selected as relevant (in PRD.md)
- file: PRD.md
  why: "READ-ONLY. §6.3 ('mode flags (--check, --init, --list, --search, --all, --completions, --path) are mutually exclusive ... mixing a <tag> with any of them is an error (exit 2)'; 'There are no bare-word subcommands'). §6.4 (error semantics — the messages ARE the user-facing contract). §17 guardrail ('check/init/completions are --check/--init/--completions precisely so a skill named ... is never shadowed')."
  section: "h2.5 (§6 CLI contract), h3.1 (§6.1 flags & modes), h3.4 (§6.4 error semantics), h2.16 (§17 guardrails)."

- url: (no external research needed — this is a pure string-literal + boolean-OR edit in stdlib Go; no libraries involved)
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && git rev-parse --short HEAD
594be07   # "Replace bare subcommands with flags in parseArgs" — S1 is DONE & COMMITTED
$ wc -l main.go
1275 main.go
$ go build ./... && echo BUILD_OK ; go vet ./... && echo VET_OK
BUILD_OK / VET_OK   # green; go test is RED from S1 drift (T3's scope)
# exclusivityError @ main.go:769-818 (doc comment 750-768). run() calls it @ 494.
# Bare cases are GONE (S1 deleted them); flag cases present (--check @296, --completions @299, --init @327).
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
```

### Desired Codebase tree with files to be changed

```bash
main.go           # MODIFY — exclusivityError body (6 strings + 2 conditions) + its doc/inline comments ONLY
# go.mod / go.sum — UNCHANGED (string-literal + bool-OR edits add no imports)
# main_test.go    — UNCHANGED in THIS subtask (T3 flips the run-level exclusivity tests; go test is EXPECTED RED)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main.go` (`exclusivityError` body) | 6 message rewrites + 2 condition additions (c.completion) | Contract LOGIC 3a/3b/3c |
| `main.go` (`exclusivityError` doc + inline comments) | flag language; remove `modes/subcommands` | Contract LOGIC 3d |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — LINE NUMBERS ARE WRONG. The contract (782-835) and change map (804-829)
// cite PRE-S1 line numbers; S1 (committed 594be07) shifted the function to 769-818.
// LOCATE BY CONTENT: find `func exclusivityError(c config) (bad bool, msg string)` and
// match the exact current text in research/verified_facts.md §1. Do NOT edit by line
// number or you will hit the wrong lines. (verified_facts §0 documents the three
// conflicting line-number sets.)

// GOTCHA #2 — The gating LOGIC is unchanged. Same four families, same order, same
// (bad, msg) signature, same exit-2 semantics at run():494. S2 changes only STRINGS
// and adds `|| c.completion` to two boolean conditions. Do NOT reorder families, add
// new families, change return counts, or wrap errors. exclusivityError must still
// return EXACTLY one (true, msg) for a bad config and (false, "") for a good one.

// GOTCHA #3 — Adding c.completion to the check/init blocks does NOT create a gap or a
// double-exit. check+completion and init+completion are currently caught by the
// completion block (which comes LAST). Adding them to the EARLIER check/init blocks
// means they now fire the check/init message instead — still (true, msg), still once,
// still exit 2. Only the WORDING's source changes. (verified_facts §3.)

// GOTCHA #4 — The completion-block condition ALREADY includes every other mode
// (c.check || c.init || c.list || c.searchMode || c.all || c.path). Do NOT add
// c.completion to it (it's c.completion's OWN block — a self-combination is
// impossible and would be nonsense). Only the check and init blocks get the addition.

// GOTCHA #5 — Do NOT touch the two TOP listing-mode returns:
//   "listing modes --path/--list/--search/--all are mutually exclusive"
//   "tags cannot be combined with --path/--list/--search/--all"
// Neither contains a bare check/init/completion word, so they are OUT OF SCOPE. The
// second is asserted by TestExclusivityErrorTagsAndPath (a direct-call test that
// PASSES today and must keep passing). Editing it breaks that test for no benefit.

// GOTCHA #6 — --completions is PLURAL (decision 19 / S1). Write "--completions" in
// every completion message, never "--completion". (The old bare subcommand was
// singular "completion"; the new flag is plural.)

// GOTCHA #7 — The init-block message is CONTRACT-PINNED verbatim (contract 3c):
//   "skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path"
// Copy it EXACTLY, including the flag order. The three mode-list messages
// intentionally differ in flag order (inherited from the contract/change map); they
// are hints tested only by substring — do NOT "normalize" the ordering or you diverge
// from the pinned string. (verified_facts §2 note.)

// GOTCHA #8 — go test ./... is EXPECTED RED (S1's bare→flag drift; the run-level
// exclusivity tests pass bare words through parseArgs). S2's gate is go build + go vet
// ONLY. Do NOT edit main_test.go to make it green — that is P1.M1.T3's scope. S2's
// own contract is verified with a throwaway direct-call probe (Level 2) deleted after.

// GOTCHA #9 — S2's string changes do NOT break the (currently RED) run-level substring
// assertions even if those tests ran: "--check" ⊃ "check", "--init" ⊃ "init",
// "--completions" ⊃ "completion". So S2 adds zero NEW test red beyond S1's drift.

// GOTCHA #10 — No deps/import change. String literals and `|| c.completion` add no
// imports. go.mod/go.sum byte-for-byte identical.

// GOTCHA #11 — SCOPE: leave these for siblings (editing them collides):
//   parseArgs (179-369)           → DONE by S1; do not touch
//   usageText (71-117) / error-prefix strings / runInit "skilldozer init:" prefixes → P1.M1.T2.S1
//   runCheck doc (631), runInit comment (1191), completion-function docs (1112-1271) → P1.M1.T2.S1
//   main_test.go exclusivity tests → P1.M1.T3
// S2's territory is ONLY exclusivityError's body (769-818) + its doc comment (750-768)
// + the init/completion inline comments inside it.
```

---

## Implementation Blueprint

### Data models and structure

None. `config` struct fields are unchanged (S1 left them; exclusivityError reads `c.check`/`c.init`/`c.completion`/`c.path`/`c.list`/`c.searchMode`/`c.all`/`c.tags`). No types added/removed.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: REWRITE the check block (message + condition + message)
  - LOCATE (by content, not line — GOTCHA #1):
        if c.check && hasTags {
            return true, "skilldozer: 'check' cannot be combined with tag arguments"
        }
        if c.check && (c.path || c.list || c.searchMode || c.all) {
            return true, "skilldozer: 'check' cannot be combined with --path/--list/--search/--all"
        }
  - REPLACE WITH:
        if c.check && hasTags {
            return true, "skilldozer: '--check' cannot be combined with tag arguments"
        }
        if c.check && (c.completion || c.path || c.list || c.searchMode || c.all) {
            return true, "skilldozer: '--check' cannot be combined with --completions/--path/--list/--search/--all"
        }
  - NOTE: adds c.completion to the condition (GOTCHA #3/#4) and --completions to the message.

Task 2: REWRITE the init block (comment flag-language + message + condition + message)
  - LOCATE (by content):
        // init is its own exclusive mode (PRD §6.3 / §8.2: like `check`). It rejects the
        // listing/inspection modes AND stray tags. A single positional <dir> after `init`
        // is consumed as the store (c.initStore) by parseArgs, so it never reaches
        // c.tags; a SECOND positional, or any positional after `init --store`, lands in
        // c.tags and is rejected here as a stray.
        if c.init {
            if hasTags {
                return true, "skilldozer: 'init' cannot be combined with tag arguments"
            }
            if c.check || c.list || c.searchMode || c.all || c.path {
                return true, "skilldozer: 'init' cannot be combined with --list/--search/--all/--path/check"
            }
        }
  - REPLACE WITH:
        // --init is its own exclusive mode (PRD §6.3 / §8.2: like `--check`). It rejects the
        // listing/inspection modes AND stray tags. A single positional <dir> after `--init`
        // is consumed as the store (c.initStore) by parseArgs, so it never reaches
        // c.tags; a SECOND positional, or any positional after `--init --store`, lands in
        // c.tags and is rejected here as a stray.
        if c.init {
            if hasTags {
                return true, "skilldozer: '--init' cannot be combined with tag arguments"
            }
            if c.check || c.completion || c.list || c.searchMode || c.all || c.path {
                return true, "skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path"
            }
        }
  - NOTE: the mode-list message is CONTRACT-PINNED verbatim (GOTCHA #7); copy exactly.

Task 3: REWRITE the completion block (comment flag-language + message; condition UNCHANGED)
  - LOCATE (by content):
        // completion is its own exclusive mode (PRD §6.3 / §14.6: like check/init). It rejects the other
        // modes/subcommands AND stray tags. `completion` does no positional capture (mirrors check), so
        // any positional after it lands in c.tags and is rejected here as a stray.
        if c.completion {
            if hasTags {
                return true, "skilldozer: 'completion' cannot be combined with tag arguments"
            }
            if c.check || c.init || c.list || c.searchMode || c.all || c.path {
                return true, "skilldozer: 'completion' cannot be combined with check/init/--path/--list/--search/--all"
            }
        }
  - REPLACE WITH:
        // --completions is its own exclusive mode (PRD §6.3 / §14.6: like --check/--init). It rejects the other
        // mode flags AND stray tags. `--completions` does no positional capture (mirrors --check), so
        // any positional after it lands in c.tags and is rejected here as a stray.
        if c.completion {
            if hasTags {
                return true, "skilldozer: '--completions' cannot be combined with tag arguments"
            }
            if c.check || c.init || c.list || c.searchMode || c.all || c.path {
                return true, "skilldozer: '--completions' cannot be combined with --check/--init/--path/--list/--search/--all"
            }
        }
  - NOTE: the CONDITION (c.check || c.init || ...) is UNCHANGED (already complete — GOTCHA #4).
    Only the two messages + the comment change.

Task 4: REWRITE the exclusivityError doc comment's bare-`check` references (flag language)
  - LOCATE (by content) the doc comment lines that say:
        //   - check + tags — `check` ignores tags, so the combo is meaningless
        //   - check + a listing mode — modes are mutually exclusive
        ... and ...
        // `check` is NOT in the listing-mode set: check+mode is caught by the families
        // below (and check+path, too — it used to silently resolve by dispatch order
        // with path winning, which was inconsistent with check+list/check+search/
        // check+all all exiting 2; N1 closed that asymmetry).
        ... and (in the Issue-3 bullet) ...
        //     check+mode and mode+mode sets)
  - REPLACE the bare `check` tokens with `--check` (e.g. "`check` ignores tags"→"`--check`
    ignores tags"; "check + tags"→"--check + tags"; "check+mode"/"check+path"/"check+list"/
    "check+search"/"check+all"→"--check+mode"/... etc.).
  - KEEP the word "modes" (PRD §6.3 legitimately says "mode flags"; the contract only
    forbids "subcommands", which does not appear in this doc comment). Do NOT rewrite the
    whole comment — only the bare-`check` references.

Task 5: VERIFY build + vet (the gate — GOTCHA #8)
  - COMMAND: go build ./...     (exit 0)
  - COMMAND: go vet ./...       (clean)
  - COMMAND: go test ./... 2>&1 | grep -c '^--- FAIL'   (EXPECTED > 0 — S1 drift, T3's scope; do NOT fix)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"
  - PROBE (Level 2, optional but recommended): a throwaway direct-call test confirming
    the new --flag substrings, deleted after.
```

### Implementation Patterns & Key Details

The edits are pure text replacement — no new patterns. The only non-obvious detail is the **mode-set symmetry** (why adding `c.completion` to check/init is safe and what it changes):

```go
// BEFORE (asymmetric): check+completion and init+completion are caught ONLY by the
// completion block (last), so they print the COMPLETION message.
//   skilldozer --check --completions  →  "...'completion' cannot be combined with check/..."  (completion msg)
//
// AFTER (symmetric): check/init blocks catch their own +completion pair, so they print
// the message for the mode the user LED with.
//   skilldozer --check --completions  →  "...'--check' cannot be combined with --completions/..."  (check msg)
// Both still return (true, msg) → run() exits 2. Only the wording's source changes.

// The added condition is a single `|| c.completion` — mirrors the existing `||` chain:
if c.check && (c.completion || c.path || c.list || c.searchMode || c.all) {   // + c.completion
    return true, "skilldozer: '--check' cannot be combined with --completions/--path/--list/--search/--all"
}
```

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **check-block message: add `--completions` to the list? → YES.** Contract 3b says add `c.completion` to the check-block *set*; the change-map table shows the check-block *message* without `--completions`. Adding `c.completion` to the condition but NOT the message would make `--check --completions` print `'--check' cannot be combined with --path/--list/--search/--all` — a message that does not mention `--completions`, the very flag the user combined. That is misleading UX. The PRP therefore adds `--completions` to BOTH the condition and the message (fully symmetric, message names exactly the modes the condition excludes). This supersedes the change-map table's literal check-message for consistency, which is the contract's stated goal.

2. **Message flag ordering: normalize? → NO.** The three mode-list targets differ in flag order (check: `--completions/--path/--list/--search/--all`; init: `--check/--completions/--list/--search/--all/--path` [contract-pinned]; completion: `--check/--init/--path/--list/--search/--all`). These are inherited from the contract/change-map and are hints tested only by substring. "Normalizing" them would diverge from the contract-pinned init string. Copy each NEW string exactly as given.

3. **Doc comment: rewrite "modes"→"mode flags"? → NO, only bare `check`→`--check`.** The contract 3d targets "modes/**subcommands**". "subcommands" appears only in the completion-block inline comment (→ "mode flags"). The doc comment uses "modes", which PRD §6.3 legitimates ("mode flags"). Over-churning every "modes"→"mode flags" is not asked for and bloats the diff. Only bare `check`/`init`/`completion` tokens in the comments become `--flags`.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports. (GOTCHA #10)

CALLER (unchanged):
  - run() @ main.go:494: `if bad, msg := exclusivityError(c); bad { fmt.Fprintln(stderr, "skilldozer: "+msg); return 2 }`
    (the exact prefix-wrapping is run()'s concern; S2 only changes the msg STRING exclusivityError
    returns. Verify run()'s prefix behavior is unchanged — it prepends "skilldozer: " ONLY if the
    message doesn't already start with it; our messages already start with "skilldozer: ", so no
    double-prefix. Do NOT edit run().)

NO ROUTES / NO DATABASE / NO CONFIG / NO COMPLETIONS:
  - S2 changes exclusivityError strings + 2 conditions + comments. Nothing else.
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (the ONLY hard gates — GOTCHA #8)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go          # must print NOTHING
go vet ./...              # expect exit 0
go build ./...            # expect exit 0

# No bare-word messages remain in exclusivityError:
awk '/func exclusivityError/,/^}/' main.go | grep -n "'check'\|'init'\|'completion'" || echo "OK: no bare-word quotes in exclusivityError"
# Expected: "OK: no bare-word quotes ..." (the six messages now say '--check'/'--init'/'--completions').

# The new flag messages ARE present:
awk '/func exclusivityError/,/^}/' main.go | grep -c "'--check'\|'--init'\|'--completions'"
# Expected: 6 (two per block: tag-args + mode-list).
```

### Level 2: Behavior spot-check (throwaway direct-call probe — NOT a committed test)

```bash
cd /home/dustin/projects/skilldozer
cat > /tmp/excl_probe_test.go <<'EOF'
package main
import ("strings"; "testing")
func TestS2ProbeExclusivityMessages(t *testing.T) {
	cases := []struct{ name string; c config; want string }{
		{"check+tags", config{check: true, tags: []string{"x"}}, "--check"},
		{"check+completion", config{check: true, completion: true}, "--check"},   // now caught by check block (GOTCHA #3)
		{"check+list", config{check: true, list: true}, "--check"},
		{"init+tags", config{init: true, tags: []string{"x"}}, "--init"},
		{"init+completion", config{init: true, completion: true}, "--init"},      // now caught by init block
		{"init+check", config{init: true, check: true}, "--init"},
		{"completion+tags", config{completion: true, tags: []string{"x"}}, "--completions"},
		{"completion+check", config{completion: true, check: true}, "--completions"},
		{"completion+init", config{completion: true, init: true}, "--completions"},
	}
	for _, tc := range cases {
		bad, msg := exclusivityError(tc.c)
		if !bad { t.Errorf("%s: bad=false; want true", tc.name); continue }
		if !strings.Contains(msg, tc.want) { t.Errorf("%s: msg=%q; want substring %q", tc.name, msg, tc.want) }
	}
	// good configs still return (false, "")
	for _, c := range []config{ {}, {check: true}, {init: true}, {completion: true}, {list: true}, {tags: []string{"x"}} } {
		if bad, _ := exclusivityError(c); bad { t.Errorf("good config %+v: bad=true; want false", c) }
	}
	// the two untouched listing-mode messages are intact
	if bad, msg := exclusivityError(config{path: true, list: true}); !bad || !strings.Contains(msg, "mutually exclusive") {
		t.Errorf("listing-modes msg changed: bad=%v msg=%q", bad, msg)
	}
	if bad, msg := exclusivityError(config{tags: []string{"x"}, path: true}); !bad || !strings.Contains(msg, "tags cannot be combined") {
		t.Errorf("tags+path msg changed: bad=%v msg=%q", bad, msg)
	}
}
EOF
cp /tmp/excl_probe_test.go excl_probe_test.go
go test -run TestS2ProbeExclusivityMessages -v ./
rm excl_probe_test.go   # throwaway; the REAL run-level tests are flipped in P1.M1.T3
# Expected: PASS. (Confirms the six --flag messages, the two new c.completion catches, good-config
# false-returns, and the two untouched listing-mode messages — all without touching main_test.go.)
```

### Level 3: Whole-module build + dependency invariant (NOT a test gate)

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # Expected: 0
go vet  ./...  ; echo "vet exit $?"     # Expected: 0

# go test is EXPECTED RED (S1's bare→flag drift). Confirm failures are the SAME set S1 left
# (run-level exclusivity tests passing bare words) — S2 must add NO new red:
go test ./... 2>&1 | grep -E '^--- FAIL' | wc -l    # note the count
go test ./... 2>&1 | grep -E '^--- FAIL' | grep -iE 'exclusivity' | head
# Expected: the same TestRunExclusivity* / TestParseArgs* failures S1 produced; nothing new.

git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
# Expected: "deps unchanged".
```

### Level 4: End-to-end message smoke (the real exit-2 path via run())

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# Each exclusive pair exits 2 and prints the flag-form message on stderr (PRD §6.4):
for pair in "--check --list" "--check --completions" "--init --list" "--init --completions" "--completions --list" "--completions --check"; do
  err=$(/tmp/sdz $pair 2>&1 1>/dev/null); code=$?
  if [ "$code" != "2" ] || ! echo "$err" | grep -qE "'--(check|init|completions)'"; then
    echo "FAIL [$pair]: code=$code err=$err"; else echo "OK [$pair]: $err"; fi
done
# A stray tag after a mode also exits 2 with flag language:
err=$(/tmp/sdz --check foo 2>&1 1>/dev/null); code=$?
[ "$code" = 2 ] && echo "$err" | grep -q "'--check'" && echo "OK [--check foo]" || echo "FAIL [--check foo]: code=$code err=$err"
rm -f /tmp/sdz
# Expected: every line prints "OK ...".
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go vet ./...` exit 0, `go build ./...` exit 0; no bare `'check'`/`'init'`/`'completion'` quotes in exclusivityError; six `'--...'` flag messages present
- [ ] Level 2 PASS — throwaway probe confirms the six `--flag` substrings, the two new `c.completion` catches (check+completion→`--check` msg; init+completion→`--init` msg), good-config false-returns, and the two untouched listing-mode messages (probe removed after)
- [ ] Level 3 PASS — build+vet exit 0; `go test` failures are the SAME S1-drift set (no new red from S2); `git diff go.mod go.sum` → "deps unchanged"
- [ ] Level 4 PASS — every exclusive pair exits 2 and prints a `'--flag' cannot be combined ...` message on stderr

### Feature Validation
- [ ] Six messages say `'--check'`/`'--init'`/`'--completions'` (no bare words)
- [ ] init-block message is EXACTLY `skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path` (contract 3c, verbatim)
- [ ] check-block condition includes `c.completion`; message lists `--completions`
- [ ] init-block condition includes `c.completion`; message lists `--completions` + `--check`
- [ ] completion-block condition UNCHANGED (already complete); message in `--completions`/`--check`/`--init` form
- [ ] Two top listing-mode returns UNCHANGED (no bare words there; `TestExclusivityErrorTagsAndPath` still passes)
- [ ] exclusivityError doc comment + init/completion inline comments use flag language; `modes/subcommands` → `mode flags`

### Code Quality / Convention Validation
- [ ] Gating logic unchanged (same families, order, signature, return contract, exit-2 semantics)
- [ ] New conditions reuse the existing `||`-chain style; comments reuse the existing PRD-§-citation style
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] Minimal diff (six strings + two `|| c.completion` + comment word swaps)

### Scope Discipline
- [ ] Did NOT touch `parseArgs` (S1's territory, done)
- [ ] Did NOT touch `usageText` / error-prefix strings / runInit `skilldozer init:` prefixes / completion-function doc comments (P1.M1.T2.S1)
- [ ] Did NOT touch `run()` dispatch or its call site at 494
- [ ] Did NOT touch `main_test.go` (P1.M1.T3 — go test is EXPECTED RED)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't edit by line number.** The contract (782-835) and change map (804-829) cite PRE-S1 numbers; S1 shifted the function to 769-818. Locate `func exclusivityError` and match the exact current text in `research/verified_facts.md` §1. (GOTCHA #1)
- ❌ **Don't change the gating logic.** Same four families, same order, same `(bad, msg)` contract, same exit-2. S2 swaps strings and adds `|| c.completion` to two conditions — nothing more. (GOTCHA #2)
- ❌ **Don't add `c.completion` to the completion-block condition.** That block IS completion's own; a self-combination is impossible. Only check and init blocks get the addition. (GOTCHA #4)
- ❌ **Don't touch the two top listing-mode returns.** They have no bare check/init/completion words; `tags cannot be combined` is asserted by a passing direct-call test. (GOTCHA #5)
- ❌ **Don't write `--completion` (singular).** The flag is `--completions` (PLURAL, decision 19 / S1). (GOTCHA #6)
- ❌ **Don't "normalize" the message flag ordering.** The init-block string is contract-pinned verbatim; the three messages intentionally differ in order and are substring-tested. Copy each NEW string exactly. (GOTCHA #7)
- ❌ **Don't treat `go test` as a gate.** It is EXPECTED RED from S1's bare→flag drift (the run-level exclusivity tests pass bare words through parseArgs). The gate is `go build` + `go vet`. Fixing tests here is T3's scope. (GOTCHA #8)
- ❌ **Don't omit `--completions` from the check-block message.** Adding `c.completion` to the condition but not the message makes `--check --completions` print a message that doesn't mention `--completions` — misleading. (DESIGN DECISION 1)
- ❌ **Don't edit `parseArgs`, `usageText`, `run()`, completion functions, or `main_test.go`.** Those are S1 (done) / T2.S1 / T3 territory. S2's only file is `main.go`, only function is `exclusivityError` (body + comments). (GOTCHA #11)
- ❌ **Don't add deps or imports.** String literals and a boolean-OR add none. (GOTCHA #10)

---

## Confidence Score

**9.5/10** — Every edit is pinned to the exact current (post-S1, HEAD `594be07`) text transcribed verbatim in `research/verified_facts.md` §1, with the exact NEW target strings in §2/§3. The single subtlety — that adding `c.completion` to the check/init blocks only changes *which* message fires (not *whether* the pair is caught) — is proven in §3, so there is no risk of a regression gap or double-exit. The build-only gate and the "don't touch tests" boundary are inherited cleanly from the already-committed S1. The 0.5 reservation is the check-block message wording: the contract mandates adding `c.completion` to the *set* but its change-map table omits `--completions` from the *message*; the PRP resolves this toward full symmetry (message names every mode the condition excludes) and documents the decision — a reviewer could prefer the change-map's literal table, but the symmetric choice is strictly better UX and matches the contract's stated "consistency" goal.
