# PRP — P1.M1.T1.S1: parseArgs — delete bare cases, add `--check`/`--init`/`--completions` flags

> **Subtask:** P1.M1.T1.S1 — the first half of P1.M1.T1 (parseArgs flag conversion). Implements decision 19 / PRD §6.3: there are **no bare-word subcommands**; `check`/`init`/`completion` become `--check`/`--init`/`--completions`, and bare `check`/`init`/`completion`/`completions` tokens become ordinary skill tags.
> **Scope boundary:** Edits ONLY `parseArgs` (delete 3 bare cases, add 3 flag cases, add `--init=` to the `=`-form switch) + the config-struct field comments + the package doc comment. Does NOT touch `exclusivityError` (that is S2), `usageText`/error-prefix strings/completion-function doc comments (that is T2.S1), `run()` dispatch, or any test (that is T3). **Tests are EXPECTED to fail after this subtask (TDD) — the gate is `go build` + `go vet`, not `go test`.**

---

## Goal

**Feature Goal**: Convert `parseArgs` so `--check`/`--init`/`--completions` are the way to select validation/init/completion modes, and bare `check`/`init`/`completion`/`completions` tokens fall through to `c.tags` (namespace safety: the entire positional namespace is reserved for skill tags, §6.3/decision 19).

**Deliverable**: Edits to `main.go` only (no new files):
1. **Delete** the three bare-word cases: `case "check":` (284-293), `case "completion":` (294-299), `case "init":` (324-351, including all Issue-4 duplicate-init logic).
2. **Add** `case "--check":`, `case "--completions":`, `case "--init":` (with the optional positional `<dir>` capture) to the main switch.
3. **Add** `case "--init":` to the `=`-form switch (mirrors `--store=`); add `--check`/`--completions` there too for consistency with the existing bool flags.
4. **Rewrite** the config-struct field comments (162/163/164/166) and the package doc comment (line 10) from "subcommand"→"flag" language.
5. Leave `--store`/`--shell` wiring (they imply `--init`/`--completions`) and the `default:` branch untouched.

**Success Definition**: `go build ./...` succeeds; `go vet ./...` passes; `go.mod`/`go.sum` unchanged. After the change: `parseArgs([]string{"--check"}).check == true`; `parseArgs([]string{"check"}).check == false && tags==["check"]`; same for init/completion/completions. (`go test ./...` is EXPECTED RED until P1.M1.T3 — that is by design, not a failure of this subtask.)

---

## User Persona (if applicable)

Not applicable at runtime — this is an internal parser refactor. The end-user-visible consequence (documented in §6.1/§6.3): `skilldozer --check` runs validation, while `skilldozer check` resolves a skill literally tagged `check`; a bare `<tab>` always means "skills" (§14).

---

## Why

- **PRD §6 (authoritative) + decision 19**: "Every action that is not a skill tag is a `--flag` — there are *no bare-word subcommands*, so the entire positional namespace is reserved for skill tags." `check`/`init`/`completion` are now `--check`/`--init`/`--completions` precisely so a skill named `check`/`init`/`completions` is never shadowed.
- **§14 namespace-safety guarantee**: a bare `<tab>` must unambiguously mean "show me my skills". Driving every non-tag action to a `--flag` makes this trivially correct and is the foundation for the skills-first completions rewrite (P1.M2.T1).
- **`completion` → `--completions` (plural)** is a deliberate rename (decision 19) so the flag and the concept align and the bare word is free for tags.

---

## What

`parseArgs` changes how three config fields get set:

| Old (bare subcommand) | New (flag) | Field set |
|---|---|---|
| `check` | `--check` | `c.check = true` |
| `completion` | `--completions` | `c.completion = true` |
| `init` / `init <dir>` | `--init` / `--init <dir>` | `c.init = true` (+ optional `c.initStore`) |

Bare `check`/`init`/`completion`/`completions` now land in `c.tags` (via the unchanged `default:` branch). `--store <dir>` still implies `--init`; `--shell <name>` still implies `--completions` (unchanged). Dispatch (`run()` → runCheck/runInit/runCompletion on the same config fields) is unchanged.

### Success Criteria

- [ ] `case "check":`, `case "completion":`, `case "init":` are GONE from the main switch (including all Issue-4 logic in the init case)
- [ ] `case "--check":` sets `c.check = true`; `case "--completions":` sets `c.completion = true`
- [ ] `case "--init":` sets `c.init = true` and captures a following NON-dashed token as `c.initStore` (i++); a dashed follower is left for its own case
- [ ] `=`-form switch has `case "--init":` (mirrors `--store=`, sets `storeMissingValue` on empty); `--check`/`--completions` also present for bool consistency
- [ ] `--store`/`--shell` wiring (implies init/completion) UNCHANGED; `default:` branch UNCHANGED
- [ ] config-struct comments (162/163/164/166) + package doc (line 10) rewritten to flag language
- [ ] `go build ./...` succeeds; `go vet ./...` passes; `go.mod`/`go.sum` unchanged
- [ ] S1 does NOT touch exclusivityError (821), runCheck doc (631), runInit comment (1191), or any test (those are S2/T2/T3)

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact line ranges to delete, the exact target code to add (main switch + `=`-form switch), the exact comment lines to rewrite, the "leave for siblings" boundary, the no-test-gate expectation, and the reasoning for why the Issue-4 logic disappears are all specified with file:line references. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
- file: main.go
  why: "THE edit site. parseArgs at :179-369. Bare cases to DELETE: case \"check\": (284-293), case \"completion\": (294-299), case \"init\": (324-351, incl. Issue-4 logic). =-form switch :196-234 (add --init= after --store= at 218-227). default branch :353-368 (NO CHANGE). --store (218-227,300-313) and --shell (228-233,314-323) wiring UNCHANGED. config struct :148-168 (rewrite field comments 162/163/164/166). package doc :1-14 (rewrite line 10)."
  pattern: "In-place conversion: each bare case → its flag form at the same spot (minimal diff). The new --init case is SIMPLER than the old init case (no Issue-4 branches)."
  gotcha: "Switch cases are order-independent; replace in spot. The =-form switch already lists EVERY bool flag (--version/--help/--path/--list/--all/--file/--relative/--no-color) — add --check/--completions there too for consistency, else --check=foo becomes an unknown flag."

- file: plan/004_5851dcff4371/architecture/code_change_map.md
  why: "Change Group 1 (1a-1e) pins the exact delete/add sites and line numbers for parseArgs; Change Group 2 lists the config-struct comment rewrites; 6d/6e list the comment sites. Verified against HEAD f30d5c5 (line numbers match live 55ada20)."
  section: "Change Group 1 (parseArgs flag conversion); Change Group 2 (config struct doc comments); 6d/6e (comment sites)"

- file: plan/004_5851dcff4371/P1M1T1S1/research/verified_facts.md
  why: "Direct-from-source proof: the exact current text of every site, the new flag-case target code, why Issue-4 logic disappears, the consistency gotcha for the =-form switch, the scope boundary (which 'reserved'/'subcommand' sites belong to siblings S2/T2), and the no-test-gate expectation."

- url: (PRD §6.1/§6.3 + decision 19 — in PRD.md, READ-ONLY)
  why: "§6.1: --check/--init/--completions are flags. §6.3: 'There are no bare-word subcommands'; --init is the sole mode accepting a positional <dir>; every other positional is a tag. Decision 19: completion→--completions (plural); namespace reserved for tags. Do NOT edit PRD.md."

- url: https://pkg.go.dev/strings#HasPrefix
  why: "Confirms strings.HasPrefix(next, \"-\") is the dashed-follower guard for the --init positional capture (already used elsewhere in parseArgs)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go go.mod
main.go   go.mod
# main.go: 1288 lines. parseArgs @ :179-369. config struct @ :148-168. package doc @ :1-14.
#          Bare cases: "check" @ :284-293, "completion" @ :294-299, "init" @ :324-351.
#          =-form switch @ :196-234 (--store= @ 218-227, --shell= @ 228-233).
#          Token-form --store @ :300-313, --shell @ :314-323. default branch @ :353-368.
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole dep).
# No new files. This subtask edits main.go only.
```

### Desired Codebase tree with files to be changed

```bash
main.go           # MODIFY — parseArgs: 3 case deletions + 3 flag-case additions + --init= in =-form switch; config-struct + package-doc comment rewrites
# go.mod / go.sum — UNCHANGED (no new imports; strings.HasPrefix + index loop already in parseArgs)
# main_test.go — UNCHANGED in THIS subtask (T3 flips the tests; go test is EXPECTED RED here)
```

**File responsibilities:**
| File | Change | Owner |
|---|---|---|
| `main.go` (parseArgs) | Bare cases → flag cases; `--init=` in `=`-form switch | Decision 19 / PRD §6.3 |
| `main.go` (config struct + package doc) | "subcommand"→"flag" comment language | This contract (f, g) |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — Tests WILL FAIL after this subtask (by design). go test ./... is NOT a
// gate here. Tests like TestParseArgsCheckSubcommand assert parseArgs([]string{"check"}).check==true;
// after S1, "check" → c.tags, c.check==false → red. Flipping tests is P1.M1.T3's job.
// S1's gate is go build ./... + go vet ./... ONLY. Do NOT edit tests to make them green.

// GOTCHA #2 — The new --init case is SIMPLER than the old init case. DELETE all the
// Issue-4 duplicate-init logic (next == "init" → append to tags) and the
// next != "check" && next != "completion" special-casing. With flags, --init --init is
// idempotent (c.init set twice), and --init <dir> captures <dir> as initStore. No
// duplicate-token handling is needed in parseArgs; exclusivity (init+tags) is enforced
// later by exclusivityError (S2, unchanged logic).

// GOTCHA #3 — --completions is PLURAL (decision 19). The old bare subcommand was
// singular "completion"; the new flag is "--completions". Do not write "--completion".
// A bare "completion" or "completions" token is now a skill tag (lands in c.tags).

// GOTCHA #4 — --init's positional capture uses strings.HasPrefix(next, "-") to decide
// whether the next token is a <dir> (capture) or a flag (leave). A dashed follower
// (--init --store …) is left for its own case. A NON-dashed follower is ALWAYS captured
// as initStore — including "check"/"init"/"completion" (--init check ⇒ initStore="check",
// a store dir so named). That is the §6.3 tradeoff (--init owns the next positional);
// do NOT re-introduce bare-word special-casing.

// GOTCHA #5 — =-form switch consistency. EVERY existing bool flag is in the =-form
// switch (it ignores the =value: --version=x == --version). For consistency, add
// --check and --completions there too (bool → ignore value), so --check=foo behaves
// like --check instead of falling to default → unknownFlag → exit 2. --init is the only
// one of the three that takes a value (mirror --store=, incl. storeMissingValue on empty).

// GOTCHA #6 — DO NOT touch --store / --shell wiring. --store <dir> still implies --init
// (sets c.init + c.initStore); --shell <name> still implies --completions (sets
// c.completion + c.completionShell). Both forms (=-form and token-form) stay as-is.

// GOTCHA #7 — DO NOT touch the default: branch (353-368). Bare check/init/completion/
// completions now flow there → c.tags. That IS the namespace-safety guarantee. No edit.

// GOTCHA #8 — SCOPE: leave these "reserved"/"subcommand" comment sites for siblings
// (editing them collides with tasks that own those functions):
//   main.go:631  (runCheck doc comment)        → P1.M1.T2.S1 (doc-comments sweep)
//   main.go:821  (exclusivityError comment)    → P1.M1.T1.S2 (exclusivityError rewrite)
//   main.go:1191 (runInit comment)             → P1.M1.T2.S1 (doc-comments sweep)
// S1's comment cleanup is ONLY: package doc (10), config struct (162/163/164/166), and
// the comments inside the deleted bare cases (which vanish with the cases).

// GOTCHA #9 — No deps change. parseArgs already imports strings and uses index-based
// iteration; the new cases add no imports. go.mod/go.sum byte-for-byte identical.

// GOTCHA #10 — Dispatch is UNCHANGED. run() still dispatches c.check→runCheck,
// c.init→runInit, c.completion→runCompletion. S1 only changes how those fields get SET,
// not the dispatch. Do not edit run() in this subtask.
```

---

## Implementation Blueprint

### Data models and structure

None added/removed. The `config` struct (main.go:150-168) is unchanged in fields; only its field doc comments change. No new types.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: DELETE the three bare-word cases in the main switch
  - DELETE main.go:284-293  case "check":  (case + comment + c.check = true)
  - DELETE main.go:294-299  case "completion":  (case + comment + c.completion = true)
  - DELETE main.go:324-351  case "init":  (ENTIRE block incl. Issue-4 logic + comments)
  - RESULT: bare check/init/completion now reach default → c.tags (GOTCHA #7)

Task 2: ADD the three flag cases (in the same switch, in-place where the bare cases were)
  - ADD case "--check":  → c.check = true  (with a one-line comment: bare check is a tag, decision 19)
  - ADD case "--completions":  → c.completion = true  (PLURAL — GOTCHA #3; comment notes bare completion/completions is a tag)
  - ADD case "--init":  → c.init = true; then if i+1 < len(args) and !strings.HasPrefix(args[i+1],"-"):
      c.initStore = args[i+1]; i++  (comment: --init owns the next positional; dashed follower left for its own case)
  - NO Issue-4 logic (GOTCHA #2). Exact target code is in research/verified_facts.md §2.

Task 3: ADD --init= to the =-form switch (and --check/--completions for consistency)
  - ADD (after case "--store": at 218-227):
      case "--init": c.init = true; c.initStore = val; if val == "" { c.storeMissingValue = true }
    (mirrors --store=; empty value ⇒ storeMissingValue, Issue 2)
  - ADD for consistency (GOTCHA #5), near the other bool cases:
      case "--check": c.check = true
      case "--completions": c.completion = true
    (bool flags ignore the =value, matching --version/--help/--path/etc.)

Task 4: REWRITE config-struct field comments (main.go:162/163/164/166) — logic unchanged
  - :162 check:     "`skilldozer check` subcommand: …"  → "`skilldozer --check` flag: …"
  - :163 init:      "`skilldozer init [<dir>]` …"        → "`skilldozer --init [<dir>]` …"
  - :164 initStore: "`init <dir>` positional or `--store <dir>` …" → "`--init <dir>` flag or `--store <dir>` …"
  - :166 completion: "`skilldozer completion` subcommand (§14.6)…" → "`skilldozer --completions` flag (§14.6)…"

Task 5: REWRITE package doc comment (main.go:10)
  - "subcommands like `check`, positional <tag> args, …" → "flags like `--check`, positional <tag> args, …"
  - (Keep the rest of the package doc as-is.)

Task 6: VERIFY build + vet (NOT test — GOTCHA #1)
  - COMMAND: go build ./...     (exit 0)
  - COMMAND: go vet ./...       (clean)
  - COMMAND: go test ./... 2>&1 | head   (EXPECTED RED — confirm failures are the bare→flag
    test drift, e.g. TestParseArgsCheckSubcommand; do NOT fix them — that is T3)
  - COMMAND: git diff --quiet go.mod go.sum && echo "deps unchanged"
```

### Implementation Patterns & Key Details

```go
// Task 2 — the three new flag cases (exact target). Switch cases are order-independent;
// replace each deleted bare case in spot for a minimal diff.

case "--check":
	// `skilldozer --check` flag (PRD §9). A bare `check` is a skill TAG (decision 19).
	c.check = true
case "--completions":
	// `skilldozer --completions [--shell <name>]` flag (PRD §14.6). PLURAL (decision 19).
	// A bare `completion`/`completions` is a skill tag, not this flag.
	c.completion = true
case "--init":
	// `skilldozer --init [<dir>]` first-run setup (PRD §8.2). --init is the sole mode
	// accepting a positional <dir> (the store to adopt, §6.3): a following NON-dashed
	// token is c.initStore. A dashed follower (--init --store …) is left for its own
	// case. A bare `init` is a skill tag (decision 19).
	c.init = true
	if i+1 < len(args) {
		next := args[i+1]
		if !strings.HasPrefix(next, "-") {
			c.initStore = next
			i++
		}
	}

// Task 3 — the =-form additions (after case "--store":). --init takes a value (mirror --store=);
// --check/--completions are bools (ignore the value, matching the 8 existing bool cases).

case "--init":
	// `--init=<dir>`: non-interactive store path (mirrors --store=). Empty value (--init=)
	// records storeMissingValue so run() rejects with exit 2 (Issue 2).
	c.init = true
	c.initStore = val
	if val == "" {
		c.storeMissingValue = true
	}
// (and, among the bool cases:)
case "--check":
	c.check = true
case "--completions":
	c.completion = true
```

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. No new imports. (GOTCHA #9)

DISPATCH (unchanged):
  - run() still dispatches c.check→runCheck, c.init→runInit, c.completion→runCompletion.
    S1 changes only how those fields get SET. No run() edit. (GOTCHA #10)

DOWNSTREAM (sibling tasks — do NOT start them here):
  - P1.M1.T1.S2 rewrites exclusivityError messages to --flags (+ mode-set consistency).
    Its territory includes main.go:821 — S1 leaves it. (GOTCHA #8)
  - P1.M1.T2.S1 rewrites usageText + error-prefix strings + completion-function doc
    comments (631, 1191, 1112-1271). S1 leaves those.
  - P1.M1.T3 flips main_test.go to the --flag contract. S1 leaves all tests alone.

NAMESPACE SAFETY (the point of decision 19):
  - After S1: skilldozer check → resolves TAG "check"; skilldozer --check → validation.
    skilldozer init → TAG "init"; skilldozer --init → first-run. skilldozer completion →
    TAG "completion"; skilldozer --completions → completion script. (§6.3)
```

---

## Validation Loop

### Level 1: Syntax & Style + build/vet (the ONLY hard gates for this subtask)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go          # must print NOTHING
go vet ./...              # expect exit 0
go build ./...            # expect exit 0

# No stale bare cases remain in the main switch:
grep -n 'case "check":\|case "completion":\|case "init":' main.go   # Expected: NO output
# The new flag cases ARE present:
grep -n 'case "--check":\|case "--completions":\|case "--init":' main.go   # Expected: 3 (main switch) + the =-form --init
# Expected: gofmt/vet/build clean; the first grep empty; the second shows the new cases.
```

### Level 2: Parser behavior spot-check (compile a throwaway probe — NOT a committed test)

```bash
cd /home/dustin/projects/skilldozer
cat > /tmp/parse_probe_test.go <<'EOF'
package main
import "testing"
func TestS1ProbeFlagContract(t *testing.T) {
	// flags set their fields
	if c := parseArgs([]string{"--check"}); !c.check || len(c.tags) != 0 { t.Fatalf("--check: %+v", c) }
	if c := parseArgs([]string{"--completions"}); !c.completion { t.Fatalf("--completions: %+v", c) }
	if c := parseArgs([]string{"--init", "/tmp/x"}); !c.init || c.initStore != "/tmp/x" { t.Fatalf("--init <dir>: %+v", c) }
	// bare words are now TAGS
	if c := parseArgs([]string{"check"}); c.check || len(c.tags) != 1 || c.tags[0] != "check" { t.Fatalf("bare check: %+v", c) }
	if c := parseArgs([]string{"init"}); c.init || len(c.tags) != 1 || c.tags[0] != "init" { t.Fatalf("bare init: %+v", c) }
	if c := parseArgs([]string{"completion"}); c.completion || c.tags[0] != "completion" { t.Fatalf("bare completion: %+v", c) }
	// --store/--shell still imply init/completion (unchanged wiring)
	if c := parseArgs([]string{"--store", "/s"}); !c.init || c.initStore != "/s" { t.Fatalf("--store: %+v", c) }
	if c := parseArgs([]string{"--shell", "bash"}); !c.completion || c.completionShell != "bash" { t.Fatalf("--shell: %+v", c) }
}
EOF
cp /tmp/parse_probe_test.go parse_probe_test.go
go test -run TestS1ProbeFlagContract -v ./
rm parse_probe_test.go   # throwaway; the REAL tests are flipped in P1.M1.T3
# Expected: PASS. (This probe confirms S1's contract without touching main_test.go.)
```

### Level 3: Whole-module build + dependency invariant (NOT a test gate)

```bash
cd /home/dustin/projects/skilldozer

go build ./...   ; echo "build exit $?"   # Expected: 0
go vet  ./...    ; echo "vet exit $?"     # Expected: 0

# go test ./... is EXPECTED RED here (bare→flag test drift). Confirm the failures are the
# expected ones, then STOP — do not fix them (that is P1.M1.T3):
go test ./... 2>&1 | grep -E "^(--- FAIL|FAIL|ok)" | head
# Expected: failures in TestParseArgsCheckSubcommand / TestParseArgsInit* / TestParseArgsCompletion*
# and run-level dispatch tests that pass "check"/"init"/"completion". These are T3's scope.

git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"
# Expected: "deps unchanged".
```

### Level 4: Namespace-safety smoke test (end-to-end, decision 19)

```bash
cd /home/dustin/projects/skilldozer
go build -o /tmp/sdz . || { echo "FAIL: build"; exit 1; }

# --check runs validation (mode); bare `check` would resolve a TAG (none here → exit 1, empty stdout)
/tmp/sdz --check >/dev/null 2>&1; echo "--check exit=$?"      # Expected: 0 (example skill is clean)
/tmp/sdz check >/dev/null 2>&1; echo "bare check exit=$?"     # Expected: 1 (unknown tag, nothing on stdout)

# --completions emits a script; bare `completion` is a tag
/tmp/sdz --completions --shell bash 2>/dev/null | grep -q '_skilldozer_completion' && echo "--completions OK" || echo "FAIL"

# --init <dir> captures the positional (non-interactive); use a throwaway config
TMP=$(mktemp -d); CFG="$TMP/cfg.yaml"; STORE="$TMP/store"
SKILLDOZER_CONFIG="$CFG" env -u SKILLDOZER_SKILLS_DIR /tmp/sdz --init "$STORE" </dev/null >/dev/null 2>&1; echo "--init exit=$?"  # Expected: 0
test -d "$STORE" && echo "--init created store OK" || echo "FAIL"
rm -rf /tmp/sdz "$TMP"
# Expected: every line prints "...OK" / the expected exit; no FAIL.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean, `go vet ./...` exit 0, `go build ./...` exit 0; no stale `case "check"/"init"/"completion":`; the three new flag cases present
- [ ] Level 2 PASS — the throwaway probe confirms --check/--init/--completions set fields and bare check/init/completion go to tags (probe removed after)
- [ ] Level 3 PASS — build+vet exit 0; `go test` failures are the EXPECTED bare→flag drift (left for T3); `git diff go.mod go.sum` → "deps unchanged"
- [ ] Level 4 PASS — `--check` runs validation (exit 0); bare `check` is a tag (exit 1, empty stdout); `--completions` emits a script; `--init <dir>` creates the store

### Feature Validation
- [ ] `case "check":`/`"completion":`/`"init":` deleted (init's Issue-4 logic gone)
- [ ] `case "--check":`/`"--completions":`/`"--init":` added (init captures a non-dashed positional)
- [ ] `=`-form switch has `--init=` (+ `--check`/`--completions` for bool consistency)
- [ ] `--store`/`--shell` wiring unchanged; `default:` branch unchanged
- [ ] config-struct comments (162/163/164/166) + package doc (line 10) in flag language

### Code Quality / Convention Validation
- [ ] New flag cases follow the existing case style (comment + single statement); --init's positional guard uses `strings.HasPrefix` like the rest of parseArgs
- [ ] No new imports; no new deps; go.mod/go.sum byte-for-byte identical
- [ ] Minimal diff (in-place conversion where possible)

### Scope Discipline
- [ ] Did NOT touch `exclusivityError` (main.go:821) — that is S2
- [ ] Did NOT touch `usageText` / error-prefix strings / completion-function doc comments (631, 1191, 1112-1271) — that is T2.S1
- [ ] Did NOT touch `run()` dispatch
- [ ] Did NOT edit any test — that is T3 (go test is EXPECTED RED here)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't treat `go test` as a gate.** It is EXPECTED RED after this subtask (TDD). The gate is `go build` + `go vet`. Trying to make tests green here means editing tests — that is T3's scope and a scope violation.
- ❌ **Don't keep any Issue-4 duplicate-init logic.** The old `next == "init"` → append-to-tags branch is deleted with the bare `init` case. Flags make it unnecessary (`--init --init` is idempotent).
- ❌ **Don't write `--completion` (singular).** The flag is `--completions` (PLURAL, decision 19). The old bare subcommand was singular; the new flag is not.
- ❌ **Don't special-case bare words in the `--init` capture.** `--init check` legitimately means initStore="check" (a store so named). The §6.3 tradeoff is that --init owns the next positional. Re-introducing `next != "check"` logic re-creates the namespace conflict decision 19 removed.
- ❋ **Don't omit `--check`/`--completions` from the `=`-form switch.** Every existing bool flag is there (ignoring the =value). Adding them keeps `--check=foo` == `--check`; omitting them makes it an unknown flag (inconsistent).
- ❌ **Don't touch `--store`/`--shell` wiring.** They still imply `--init`/`--completions`. Both `=`-form and token-form stay as-is.
- ❌ **Don't edit exclusivityError (821), runCheck doc (631), runInit comment (1191), or completion-function docs.** Those are siblings' scope (S2/T2). Editing 821 collides with S2's wholesale exclusivityError rewrite.
- ❌ **Don't add deps or imports.** `strings.HasPrefix` and the index loop are already in parseArgs.
- ❌ **Don't edit `run()` dispatch.** The config fields still drive the same functions; only how they get SET changes.

---

## Confidence Score

**9/10** — Every delete/add site is pinned to a verified live line number with exact target code; the new `--init` case is strictly simpler than the old `init` case (Issue-4 logic deleted by design); the consistency gotcha (`=-form` switch) is called out; and the no-test-gate / leave-for-siblings boundaries are explicit so the implementer neither over-reaches into T3/S2/T2 nor under-delivers. The 1-point reservation is the `--check`/`--completions` addition to the `=`-form switch: the contract literally mandates only `--init=`, but omitting the two bools creates a subtle inconsistency with the 8 existing bool flags — the PRP recommends adding them for consistency and flags the judgment call.
