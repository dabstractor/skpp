# PRP — P1.M4.T10.S1: `internal/check` validation + `skpp check` subcommand (PRD §9)

> **Subtask:** P1.M4.T10.S1 — the `skpp check` subcommand (PRD §6.1 + §9). It is
> the ONLY subtask of T10 ("`skpp check` — validation"). It walks the manifest-free
> store and reports every skill's problems (missing/invalid frontmatter fields,
> duplicate names, over-length descriptions), printing a `OK`/`WARN`/`ERROR`
> report and exiting `1` if any `ERROR` is found.
>
> **Scope:** four files — CREATE `internal/check/check.go` +
> `internal/check/check_test.go`; MODIFY `main.go` (add `check bool` to `config`,
> add `case "check":` to `parseArgs`, add the `check` branch to `run`, add the
> `internal/check` import) and `main_test.go` (12 new tests + 1 existing test
> MODIFIED: `TestParseArgsUnknownTolerated` used the now-reserved token `check` as a
> positional tag, so it must switch to a non-reserved tag; the other 68 stay byte-identical).
>
> **DEPENDENCIES (CONTRACT):** P1.M2.T4 (`discover.ParseFrontmatter` + `Frontmatter`),
> P1.M2.T4.S2 (`discover.Skill` + `BuildSkill`), P1.M2.T5.S1 (`discover.Index`),
> P1.M1.T2 (`skillsdir.Find`), and P1.M1.T3 (`main.run`/`config`/`parseArgs`) are
> LANDED and GREEN — all consumed verbatim. `discover.Index(dir)` returns the
> pre-sorted `[]Skill`; `discover.ParseFrontmatter(path)` is re-run per skill to
> recover the malformed-YAML-vs-no-block distinction the index drops.
>
> **SCOPE DECISION (authoritative — see research/verified_facts.md §2,§3,§7):**
> The validation logic lives in a new **`internal/check`** package that mirrors
> `internal/search` (a function over `[]discover.Skill`, own package, own
> `_test.go`); `main.run` stays a thin dispatcher + renderer. This subtask owns
> `check` + the `internal/check` package ONLY. It does NOT add `--help` (M5), the
> §6.3 mutual-exclusivity / exit-2 logic (M5), `install.sh`/README/completions
> (M6), or `skills/example/` (M6). It does NOT touch `internal/discover/*`,
> `internal/skillsdir/*`, `internal/resolve/*`, `internal/ui/*`,
> `internal/search/*`, `go.mod`, `go.sum`, or `PRD.md`, and adds NO third-party
> dependency (stdlib `fmt`/`regexp`/`sort`/`strings` only).
>
> **TWO PRD §9 BULLETS REFRAMED (documented, not dropped silently):**
> - *"ERROR: skill dir has no SKILL.md"* → reframed to **"unusable SKILL.md"**:
>   the re-parse returning `err != nil` (malformed YAML or unreadable file).
>   Directory-based discovery cannot surface dirs that never had a SKILL.md without
>   a fragile heuristic that false-positives on grouping dirs (§2). OUT OF SCOPE.
> - *"WARN: a skill dir is empty besides SKILL.md"* (PRD: "optional") → **SKIPPED**.
>   The shipped `example` skill IS only `SKILL.md`; enabling it would break the
>   §13 acceptance ("reports the example as OK") (§3).

---

## Goal

**Feature Goal**: Ship `skpp check` so a user can validate every skill in their
manifest-free store in one command — catching missing/invalid frontmatter
fields, duplicate `name`s, and over-length `description`s — and get a readable
`OK`/`WARN`/`ERROR` report with a summary, exiting non-zero iff any `ERROR`
exists. `check` re-parses each `SKILL.md` (via the existing `discover.ParseFrontmatter`)
to recover the malformed-YAML-vs-no-block distinction that `discover.Index`
intentionally hides, so the report is strictly richer than what `--list` shows.

**Deliverable**: Four files (two NEW, two MODIFIED; no other files touched):
1. `internal/check/check.go` — `package check`; `func Check(skills []discover.Skill) Report`
   + the `Severity`/`Finding`/`SkillReport`/`Report` types + the unexported
   `validateName`/`dup` helpers. Imports only `fmt`/`regexp`/`sort`/`strings` +
   `internal/discover`. ~140 lines.
2. `internal/check/check_test.go` — `package check` (white-box); 18 tests
   covering valid-skill, missing-block, malformed-YAML, missing/invalid name
   (all 5 name rules + 64-char boundary), missing/empty/too-long description
   (1024-char boundary), duplicate names (incl. cross-tag message + the
   "missing name is not a dup" guard), `HasErrors`, and nil input.
3. `main.go` — MODIFY (3 localized edits + 1 import): `config` gains `check bool`;
   `parseArgs` gains `case "check":`; `run` gains the `if c.check {…}` branch
   (after `--search`, before `--all`); the import block gains `internal/check`.
4. `main_test.go` — MODIFY (append 12 + fix 1 existing): 12 new tests (3 `parseArgs`
   + 9 `run`) using the existing `sampleStore`/`writeSkillTree`/`unsetSkillsEnv`
   helpers, PLUS a one-line fix to the existing `TestParseArgsUnknownTolerated`
   (it asserted the now-reserved token `check` is captured as a tag; swap to a
   non-reserved tag). The other 68 existing tests are byte-identical.

**Success Definition**: `gofmt -l internal/check/*.go main.go main_test.go` is
silent; `go vet ./...` is clean; `go build ./...` and `go test ./...` pass
(**196 tests total**: 166 baseline + 18 new check-package + 12 new main).
`go mod tidy` is a **no-op** (`go.mod`/`go.sum` unchanged — stdlib only).
`./skpp check` on a clean store prints `OK` lines + a `0 errors` summary, exit 0;
on a store with a bad skill prints `ERROR` lines + a summary with `errors > 0`,
exit 1; WARNs never change the exit code. No touch to
`internal/{discover,skillsdir,resolve,ui,search}/*`, `go.mod`/`go.sum`, `PRD.md`;
no `--help`, exit-2, `skills/example/`, or completions.

## User Persona

**Target User**: A pi operator who keeps skills in a centralized, non-auto-discovered
store and loads them on demand via `pi --skill "$(skpp <tag>)"`.

**Use Case**: "I just dropped three new skills into `skills/` by hand and tweaked
their frontmatter. Before I rely on them, I run `skpp check` to confirm none have a
missing `name`, a duplicate `name`, an invalid name, or an over-long `description`
— and that pi will actually load them."

**User Journey**: edit `skills/foo/SKILL.md` → `skpp check` → read the report →
fix any `ERROR`/`WARN` line (it names the tag + the problem) → re-run until clean.

**Pain Points Addressed**: pi silently refuses to load a skill with no
`description` and warns on duplicate `name`s (keeping the first). With a growing
store, those failures are invisible until a `--skill` load misbehaves. `check`
surfaces them on demand, manifest-free, against the same on-disk source of truth
`skpp` already walks.

## Why

- **Catches the failures pi hides until load time.** pi does not load a skill with
  an empty `description` (PRD §3) and warns + keeps the first on duplicate `name`s.
  `check` surfaces both (plus invalid `name` charset/length, which pi would also
  reject) before they bite a `pi --skill "$(skpp …)"` invocation.
- **Strictly richer than `--list`, same data source.** `--list` shows
  `(missing)`/`(none)` placeholders but cannot tell *why* (malformed YAML vs no
  block vs empty field) because `discover.Index` drops the parse error. `check`
  re-runs `discover.ParseFrontmatter` per skill — the exact re-parse the
  `index.go` doc comment already documents for M4 — so it recovers that distinction.
- **Cheap to build, zero new surface risk.** The validation is a pure function
  over `[]discover.Skill` plus an idempotent re-parse; `main` renders the structured
  result. It reads the same store, so it cannot regress `--list`/`--search`/
  `--all`/`<tag>` — it only reports on them.
- **Mirrors `internal/search`.** Establishes that the report/matching concerns
  over `[]discover.Skill` are each a self-tested `internal/*` package, keeping
  `main` a thin dispatcher + renderer.

## What

User-visible behavior (PRD §6.1 + §9):

- `skpp check`: validates every skill in the resolved store and prints a report:
  - One `OK   <relTag> (<name>)` line per clean skill.
  - One `ERROR <relTag> (<name>): <reason>` line per ERROR (a skill with several
    problems emits several ERROR/WARN lines).
  - One `WARN  <relTag> (<name>): <reason>` line per WARN.
  - A final summary: `<N> skills, <M> errors, <K> warnings`.
- Exit `0` if there are no ERRORs; exit `1` if there is any ERROR. WARNs never
  change the exit code (so `if skpp check; then …` works as a pass/fail gate).
- An empty store → `0 skills, 0 errors, 0 warnings`, exit `0` (clean).
- The skills dir cannot be located → one-line fix to stderr, exit `1`, empty stdout
  (same as `--list`/`--search`/`--all`/`<tag>`).
- `check` is a REPORT: it prints its **full findings to stdout** regardless of
  pass/fail (pipeable). It is NOT subject to §6.4's "nothing on stdout on failure"
  (that is the path-emitter contract for `$(...)`; `check` never participates in
  command substitution).
- `--file`/`--relative`/`--no-color` do NOT apply (status report, not paths/table).

### Validation rules implemented (PRD §9)

For each skill (from `discover.Index`), re-parse `s.SourceFile` via
`discover.ParseFrontmatter`, then:

| Outcome of re-parse | Finding |
|---|---|
| `err != nil` (malformed YAML OR unreadable file) | **ERROR** `invalid SKILL.md frontmatter: <err>` |
| `fm.HasFM == false` (no `---` block) | **ERROR** `missing frontmatter block (no '---' delimiters)` |
| `HasFM` true, `name == ""` | **ERROR** `frontmatter 'name' is missing` |
| `HasFM` true, `name` invalid charset/structure | **ERROR** `frontmatter 'name' must be lowercase a-z0-9 with single hyphens (no leading/trailing/consecutive hyphens)` |
| `HasFM` true, `len(name) > 64` | **ERROR** `frontmatter 'name' is N chars (max 64)` |
| `HasFM` true, `description` trims to `""` | **ERROR** `frontmatter 'description' is missing or empty` |
| `HasFM` true, `len(trimmed description) > 1024` | **WARN** `description is N chars (max 1024)` |
| (global) `name` shared by >1 skill | **ERROR** per owner: `duplicate frontmatter 'name' "<name>" (also in: <other relTags>)` |

Notes: a malformed/no-block skill gets ONE root-cause ERROR (its name/description
are definitionally absent — not double-reported). `name` validity uses the regex
`^[a-z0-9]+(-[a-z0-9]+)*$` (charset + structure) plus an explicit `len > 64` cap
(the regex can't express the max). Description length is measured on the
**trimmed** description (matches `ui.go`'s display length; a folded-scalar
trailing newline does not count). The duplicate scan is global, case-sensitive on
**non-empty** names, run in a second pass; the "other" relTags are sorted.

### Success Criteria

- [ ] `skpp check` (subcommand) is recognized; the token `check` is NOT captured
      as a tag.
- [ ] A clean store → one `OK` line per skill + `N skills, 0 errors, 0 warnings`,
      exit `0`.
- [ ] A store with a bad skill → `ERROR` line(s) + summary with `errors > 0`,
      exit `1`.
- [ ] Malformed YAML and a missing frontmatter block are reported as DISTINCT
      root-cause ERRORs (re-parse distinguishes them).
- [ ] All five `name` rules enforced: missing, leading/trailing/consecutive
      hyphen, uppercase/other charset; plus the 64-char max (ERROR) and the
      64-char boundary (OK).
- [ ] Missing/empty `description` → ERROR; `> 1024` chars → WARN (1024 boundary OK).
- [ ] Duplicate `name` across skills → one ERROR per owner, naming the other tag(s).
- [ ] WARNs never change the exit code (a store with only WARNs exits 0).
- [ ] Empty store → `0 skills, 0 errors, 0 warnings`, exit 0.
- [ ] Full report prints to **stdout**; exit code is the pass/fail signal.
- [ ] Skills dir unresolvable → exit 1, empty stdout, one-line fix on stderr.
- [ ] `--version` still precedes `check` (PRD §6.3).
- [ ] `go test ./...` green (196 tests); `gofmt`/`go vet` clean; `go.mod` unchanged.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed to
implement this successfully?_ **Yes.** The four files are specified verbatim below
(two full new files, three exact edit snippets for `main.go`, twelve named test
functions for `main_test.go`). The re-parse contract, the `Skill`/`Frontmatter`
field semantics, the dispatch order, the output format, and every validation rule
are pinned in the research notes and the Blueprint. No external library is involved.

### Documentation & References

```yaml
# MUST READ — the authoritative spec for this command
- file: PRD.md
  section: "§9 (Validation — skpp check) and §6.1 (the `check` row)"
  why: "Defines every validation rule (ERROR/WARN), the output format
        (OK/ERROR/WARN lines + summary), and the exit code (1 if any ERROR)."
  critical: "§9 lists 'skill dir has no SKILL.md' (ERROR) and 'empty besides
             SKILL.md' (WARN, optional). Both are REFRAMED in this PRP (see
             Scope decisions + research §2/§3): the former -> unusable SKILL.md
             via re-parse; the latter -> skipped (would break §13 acceptance)."

# MUST READ — the parser check RE-RUNS (the malformed-YAML distinction lives here)
- file: internal/discover/discover.go
  why: "ParseFrontmatter(path) returns (Frontmatter, body, err). err != nil means
        malformed YAML between fences (or an unreadable file); fm.HasFM==false with
        nil err means no '---' block. check re-runs this PER SKILL because
        discover.Index drops the err."
  pattern: "fm, _, err := discover.ParseFrontmatter(s.SourceFile)"
  gotcha: "Index IGNORES err and builds a HasFM=false Skill for both cases, so the
           index alone CANNOT tell them apart — the re-parse is the ONLY way."

# MUST READ — the data type + the field semantics check depends on
- file: internal/discover/skill.go
  why: "Skill.SourceFile (== Dir+'/SKILL.md') is the path to re-parse; Skill.RelTag
        is the display tag; Skill.Name is the display name (== re-parsed fm.Name
        when the parse is clean)."
  pattern: "discover.Skill{ RelTag, Name, Description string; SourceFile string; HasFM bool; ... }"
  gotcha: "Description is copied VERBATIM incl. a folded-scalar trailing newline —
           check measures len(strings.TrimSpace(fm.Description)) for the 1024 WARN."

# MUST READ — the pre-sorted source of skills
- file: internal/discover/index.go
  why: "Index(dir) returns []Skill SORTED by RelTag. check iterates it in order, so
        the report is already sorted by tag (deterministic output)."
  pattern: "func Index(skillsDir string) ([]Skill, error)  // sorted by RelTag"
  gotcha: "Lines 38-44 of index.go EXPLICITLY say check re-runs ParseFrontmatter to
           distinguish malformed YAML from no block — this task implements that note."

# MUST READ — the direct precedent: a function over []discover.Skill in its own pkg
- file: internal/search/search.go
  why: "check mirrors this exactly: a function over []discover.Skill in its own
        internal/ package with its own _test.go, called by main. check returns a
        structured Report (data only); main renders it (search let main render via
        ui.PrintList — same split: data vs rendering)."
  pattern: "package search; func Search(query, skills) []discover.Skill  // over the index"

# MUST READ — the Agent Skills name rule (PRD §3) check enforces
- file: PRD.md
  section: "§3 (name rules: 1-64 chars, lowercase a-z0-9-, no leading/trailing/consecutive hyphens)"
  why: "Drives the name-validity ERROR. Verified live: regex
        ^[a-z0-9]+(-[a-z0-9]+)*$ enforces charset+structure; the 64-char max is a
        separate len check (regex can't express it). Names are ASCII -> len == runes."

# MUST READ — the three exact insertion points in the dispatcher
- file: main.go
  why: "config struct (add field), parseArgs (new case 'check'), run dispatch (insert
        check branch after --search, before --all), import block."
  pattern: "run() order: version -> path -> list -> search -> [CHECK] -> all -> tags -> default"
  gotcha: "importing internal/check as `check` does NOT collide with c.check (struct
           field) or case \"check\": (string literal) — different Go namespaces."

# Reference (stdlib; stable since Go 1.0 — no version concern)
- url: https://pkg.go.dev/regexp#MatchString
  why: "regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(name) validates the
        Agent Skills name charset+structure. Compile ONCE at package scope."
```

### Current Codebase tree (relevant slice)

```bash
skpp/
├── go.mod                      # module github.com/dabstractor/skpp; go 1.25; dep yaml.v3 (UNCHANGED)
├── main.go                     # MODIFY: config + parseArgs + run + import
├── main_test.go                # MODIFY: append 12 tests + fix 1 existing (TestParseArgsUnknownTolerated)
└── internal/
    ├── discover/{skill.go,index.go,discover.go,*_test.go}   # READ-ONLY (provides Skill, Index, ParseFrontmatter)
    ├── resolve/{resolve.go,resolve_test.go}                 # READ-ONLY
    ├── search/{search.go,search_test.go}                    # READ-ONLY (the pattern to mirror)
    ├── skillsdir/{skillsdir.go,*_test.go}                   # READ-ONLY (provides Find)
    ├── ui/{ui.go,ui_test.go}                                # READ-ONLY
    └── check/                                              # NEW PACKAGE (this subtask)
        ├── check.go                                        # CREATE
        └── check_test.go                                   # CREATE
```

### Desired Codebase tree with files to be added/modified

```bash
internal/check/check.go        # NEW — package check; Check() + Report/Severity/Finding; stdlib-only
internal/check/check_test.go   # NEW — 18 tests (white-box; temp SKILL.md files via a local helper)
main.go                        # MODIFY — 3 localized edits (config / parseArgs / run) + 1 import
main_test.go                   # MODIFY — append 12 tests (3 parseArgs + 9 run)
```

### Known Gotchas of our codebase & Library Quirks

```go
// CRITICAL: discover.Index DROPS the per-skill parse error (malformed YAML builds
// a HasFM=false Skill and is still resolvable by dir). check MUST re-run
// discover.ParseFrontmatter(s.SourceFile) to recover err!=nil (malformed) vs
// HasFM==false && err==nil (no block). The index alone CANNOT distinguish them.

// CRITICAL: `check` is a POSITIONAL subcommand token (no dash). parseArgs currently
// captures every non-dash token into c.tags. Add `case "check": c.check = true`
// BEFORE the default tag-capture, so `check` is NOT appended to tags. `check` is a
// RESERVED name: a skill literally tagged `check` cannot be resolved this way.

// CRITICAL: check is a REPORT — print the full findings to STDOUT regardless of
// pass/fail; signal pass/fail via the exit code (1 iff any ERROR). Do NOT apply
// §6.4's "nothing on stdout on failure" (that is the path-emitter / $(...) contract;
// check never participates in command substitution).

// GOTCHA: description length is measured on strings.TrimSpace(fm.Description), NOT
// the raw fm.Description (which may carry a folded-scalar trailing newline). A
// whitespace-only description trims to "" -> ERROR "missing or empty", not a WARN.

// GOTCHA: the name regex ^[a-z0-9]+(-[a-z0-9]+)*$ enforces charset + structure
// (non-empty, no leading/trailing/consecutive hyphens) but CANNOT express the
// 64-char max. Check len(name) > 64 FIRST (before the regex) and emit a distinct
// "N chars (max 64)" ERROR. Names are ASCII -> len(name) == rune count.

// GOTCHA: a malformed-YAML or no-block skill gets ONE root-cause ERROR only (its
// name/description are definitionally absent — do NOT also emit missing-name/
// missing-description ERRORs for it; that is noise). Field checks run ONLY when
// fm.HasFM is true and err is nil.

// GOTCHA: the duplicate-name scan is GLOBAL and runs in a SECOND pass (after the
// per-skill field checks), collecting non-empty names -> []relTag. Each owner of a
// duplicated name gets its own ERROR naming the OTHER relTag(s), sorted. A skill
// with a missing name is excluded from the scan (it already has its own ERROR).

// GOTCHA: WARNs never change the exit code. A store with only WARNs (e.g. one
// over-long description) exits 0. Empty store -> 0 skills/0 errors/0 warnings, exit 0.

// GOTCHA: an empty store is CLEAN -> exit 0 (unlike --list which exits 1 on empty).
// check is validation: no skills == nothing wrong.

// GOTCHA: --file/--relative/--no-color do NOT apply to check (it prints a status
// report, not paths or a colorized table). M5.T11 owns the §6.3 exclusivity error
// (`check` mixed with tags/--list -> exit 2); until then check wins silently in
// run() dispatch (mirrors how searchMode currently wins over tags).

// NO new third-party dependency: stdlib fmt/regexp/sort/strings only. `go mod tidy`
// is a no-op.
```

## Implementation Blueprint

### Data models and structure

The feature introduces a small, self-contained data model in `internal/check`.
It consumes the existing `discover.Skill` (READ-ONLY) and re-uses
`discover.ParseFrontmatter` (READ-ONLY). No change to any `discover` type.

### File 1 — CREATE `internal/check/check.go` (full content)

```go
// Package check validates every skill in a manifest-free store against the PRD §9
// rules and PRD §3 Agent Skills name rules. It is a FUNCTION over
// []discover.Skill (the pre-sorted catalog from discover.Index) that returns a
// structured Report; main.run (P1.M4.T10.S1) renders the report to stdout and
// maps Report.HasErrors() to the exit code.
//
// It mirrors internal/search (a function over []discover.Skill in its own
// internal/ package with its own _test.go): the validation concern stays
// isolated, independently unit-testable, and out of the thin main dispatcher.
//
// The non-obvious part: discover.Index DROPS the per-skill frontmatter parse
// error (a malformed-YAML SKILL.md still builds a HasFM=false Skill so it stays
// resolvable by directory). check therefore RE-RUNS discover.ParseFrontmatter on
// each s.SourceFile to recover the malformed-YAML-vs-no-block distinction — the
// exact re-parse internal/discover/index.go's doc comment already documents for
// "check (M4/T10)". The double parse is cheap (small files, small store) and
// idempotent (no rework in discover).
package check

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dabstractor/skpp/internal/discover"
)

// Severity ranks a finding. OK < WARN < ERROR. OK is the implicit value for a
// skill with no findings (main prints an "OK" line); it is never carried by a
// Finding. Exported so main can switch on it if needed (it renders via String()).
type Severity int

const (
	LevelOK Severity = iota
	LevelWarn
	LevelError
)

// String renders a Severity as the 3-5 char status word main left-pads to width 5
// (`OK   `, `WARN `, `ERROR`). Mirrors resolve.MatchKind.String() / Source.String().
// An out-of-range value renders as "OK" (the zero value), which is safe.
func (s Severity) String() string {
	switch s {
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "OK"
	}
}

// Finding is one validation result line for a single skill. A skill with zero
// findings is OK (main emits one "OK" line); a skill with N findings emits N
// ERROR/WARN lines. Message is empty for OK (OK findings are never created).
type Finding struct {
	Level   Severity
	Message string
}

// SkillReport binds a skill to its findings. BySkill is in the input order
// (discover.Index sorts by RelTag), so the report is deterministic.
type SkillReport struct {
	Skill    discover.Skill
	Findings []Finding // empty => the skill is OK
}

// Report is the full check output. BySkill is in input order; Errors/Warnings are
// the totals across all findings (drive the summary line + exit code).
type Report struct {
	BySkill  []SkillReport
	Errors   int
	Warnings int
}

// HasErrors reports whether any ERROR finding exists. main maps this to the exit
// code (PRD §9: exit 1 if any ERROR). WARNs never affect it.
func (r Report) HasErrors() bool { return r.Errors > 0 }

// nameOwner pairs a skill's canonical tag with its frontmatter name for the
// duplicate-name scan. Only skills whose frontmatter parsed (HasFM, no err) AND
// whose name is non-empty are collected — the only names that can duplicate.
// Defined at PACKAGE scope (not inside Check) so it can be passed to
// appendDupFindings without an anonymous-struct assignability mismatch (a named
// type and an anonymous struct with identical fields are NOT the same type in Go).
type nameOwner struct {
	relTag string
	name   string
}

// validName enforces the PRD §3 Agent Skills name charset + structure: lowercase
// a-z0-9 with single hyphens, no leading/trailing/consecutive hyphens. It CANNOT
// express the 64-char max (checked separately) nor emptiness (a missing name is
// its own ERROR, handled before this regex is consulted). Verified live: accepts
// example/foo-helper/a/123/a-b-c; rejects -foo/foo-/foo--bar/Foo/foo_bar.
var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// PRD §3 / §9 limits. nameLenMax is the Agent Skills name ceiling; descLenMax is
// the description ceiling (PRD §3: "description max 1024 chars"). Both measured on
// ASCII / trimmed content respectively (see Check).
const (
	nameLenMax = 64
	descLenMax = 1024
)

// Check validates every skill in skills against the PRD §9 rules and returns a
// structured Report. It is the P1.M4.T10.S1 deliverable; main.run renders it.
//
// Algorithm (three passes):
//
//  1. PER-SKILL local checks: re-parse s.SourceFile (recover malformed-YAML vs
//     no-block), then — only when the frontmatter parsed — check name presence,
//     name validity (charset/structure/length), description presence, and
//     description length. Collect each non-empty name for the dup scan.
//  2. GLOBAL duplicate-name scan: any non-empty name owned by >1 skill yields one
//     ERROR per owner, naming the other relTag(s) (sorted).
//  3. TALLY errors/warnings across all findings.
//
// A skill whose re-parse failed (malformed YAML) or had no '---' block gets ONE
// root-cause ERROR (its name/description are definitionally absent; field checks
// are skipped to avoid noise). Field checks run only when fm.HasFM && err == nil.
//
// check does NOT scan for "directories that lack SKILL.md but look like skills":
// discover.Index only emits dirs that CONTAIN a SKILL.md, and a heuristic for the
// gap would false-positive on legitimate grouping dirs (research §2). The §9
// "empty besides SKILL.md" WARN is intentionally NOT implemented (research §3):
// the shipped example skill IS only SKILL.md, and enabling it would break the
// §13 acceptance ("reports the example as OK").
func Check(skills []discover.Skill) Report {
	var rep Report
	rep.BySkill = make([]SkillReport, len(skills))

	// owners collects (relTag, name) ONLY for skills whose frontmatter parsed
	// (HasFM, no err) and whose name is non-empty — the only names that can dup.
	// nameOwner is package-scoped (see above) so it is passable to appendDupFindings.
	var owners []nameOwner

	// Pass 1: per-skill local checks.
	for i := range skills {
		s := skills[i]
		var findings []Finding
		fm, _, perr := discover.ParseFrontmatter(s.SourceFile)
		switch {
		case perr != nil:
			// Malformed YAML between fences, OR the file vanished between Index and
			// check (race) -> ParseFrontmatter returns the os/yaml error. This is the
			// reframed §9 "skill dir has no SKILL.md": an UNUSABLE SKILL.md.
			findings = append(findings, Finding{LevelError, "invalid SKILL.md frontmatter: " + perr.Error()})
		case !fm.HasFM:
			// No '---' block at all. Root cause: name+description are absent. One ERROR.
			findings = append(findings, Finding{LevelError, "missing frontmatter block (no '---' delimiters)"})
		default:
			findings = append(findings, checkFields(fm)...)
			if fm.Name != "" {
				owners = append(owners, nameOwner{s.RelTag, fm.Name})
			}
		}
		rep.BySkill[i] = SkillReport{Skill: s, Findings: findings}
	}

	// Pass 2: global duplicate-name scan (case-sensitive on non-empty names).
	appendDupFindings(&rep, owners)

	// Pass 3: tally.
	for i := range rep.BySkill {
		for _, f := range rep.BySkill[i].Findings {
			switch f.Level {
			case LevelError:
				rep.Errors++
			case LevelWarn:
				rep.Warnings++
			}
		}
	}
	return rep
}

// checkFields runs the per-field ERROR/WARN checks for a skill whose frontmatter
// parsed (HasFM, no err). Order: name presence -> name length -> name charset ->
// description presence -> description length. Each failure appends its own Finding
// (a skill can accumulate several, e.g. invalid name + over-long description).
//
// Description length is measured on the TRIMMED value (strings.TrimSpace), matching
// ui.go's display length: a folded-scalar trailing newline does not count, and a
// whitespace-only description trims to "" -> "missing or empty" ERROR (not a WARN).
func checkFields(fm discover.Frontmatter) []Finding {
	var f []Finding

	// name presence + validity.
	if fm.Name == "" {
		f = append(f, Finding{LevelError, "frontmatter 'name' is missing"})
	} else if len(fm.Name) > nameLenMax {
		f = append(f, Finding{LevelError, fmt.Sprintf("frontmatter 'name' is %d chars (max %d)", len(fm.Name), nameLenMax)})
	} else if !validName.MatchString(fm.Name) {
		f = append(f, Finding{LevelError, "frontmatter 'name' must be lowercase a-z0-9 with single hyphens (no leading/trailing/consecutive hyphens)"})
	}

	// description presence + length.
	desc := strings.TrimSpace(fm.Description)
	if desc == "" {
		f = append(f, Finding{LevelError, "frontmatter 'description' is missing or empty"})
	} else if len(desc) > descLenMax {
		f = append(f, Finding{LevelWarn, fmt.Sprintf("description is %d chars (max %d)", len(desc), descLenMax)})
	}

	return f
}

// appendDupFindings adds a duplicate-name ERROR to every skill that shares a
// non-empty frontmatter name with at least one other skill. owners is the
// (relTag, name) list collected in pass 1. The "also in" list excludes the skill
// itself and is sorted for deterministic output (PRD §6.4-style stable reports).
//
// It mutates rep.BySkill in place by matching sr.Skill.RelTag (RelTag is unique
// per directory, so the match is unambiguous). A skill whose name is invalid (bad
// charset) but duplicated still counts — the literal name string matches.
func appendDupFindings(rep *Report, owners []nameOwner) {
	byName := map[string][]string{}
	for _, o := range owners {
		byName[o.name] = append(byName[o.name], o.relTag)
	}
	for name, tags := range byName {
		if len(tags) < 2 {
			continue
		}
		sort.Strings(tags)
		for i := range rep.BySkill {
			sr := &rep.BySkill[i]
			if sr.Skill.Name != name {
				continue
			}
			others := make([]string, 0, len(tags)-1)
			for _, t := range tags {
				if t != sr.Skill.RelTag {
					others = append(others, t)
				}
			}
			sr.Findings = append(sr.Findings, Finding{
				Level:   LevelError,
				Message: fmt.Sprintf("duplicate frontmatter 'name' %q (also in: %s)", name, strings.Join(others, ", ")),
			})
		}
	}
}
```

### File 2 — CREATE `internal/check/check_test.go` (full content)

```go
package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dabstractor/skpp/internal/discover"
)

// mkSkill writes content to a temp skills/<relTag>/SKILL.md and returns the
// discover.Skill the way discover.Index would: parse (ignoring err, like Index),
// then discover.BuildSkill. Each skill gets its own temp root so SourceFile is
// unique — check re-parses each SourceFile independently, so isolation is correct.
// relTag uses '/' separators (cross-platform via filepath.FromSlash).
//
// This mirrors main_test.go's writeSkillTree but produces a single Skill, so
// check tests can build an arbitrary []discover.Skill (incl. dups) without a
// shared root.
func mkSkill(t *testing.T, relTag, content string) discover.Skill {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(relTag))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	fm, _, _ := discover.ParseFrontmatter(path) // Index ignores err; check re-parses
	return discover.BuildSkill(dir, relTag, fm)
}

// skill returns a Skill with a valid block, for the cases that only vary name/desc.
func skill(t *testing.T, relTag, name, desc string) discover.Skill {
	t.Helper()
	return mkSkill(t, relTag, "---\nname: "+name+"\ndescription: "+desc+"\n---\n# body\n")
}

// repeat returns a string of n copies of s (for boundary-length name/desc tests).
func repeat(s string, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(s)
	}
	return b.String()
}

func TestCheckValidSkillIsClean(t *testing.T) {
	rep := Check([]discover.Skill{skill(t, "example", "example", "A demo skill.")})
	if len(rep.BySkill) != 1 || len(rep.BySkill[0].Findings) != 0 {
		t.Errorf("clean skill should have zero findings; got %+v", rep.BySkill[0].Findings)
	}
	if rep.Errors != 0 || rep.Warnings != 0 {
		t.Errorf("clean skill: Errors=%d Warnings=%d; want 0,0", rep.Errors, rep.Warnings)
	}
	if rep.HasErrors() {
		t.Errorf("clean skill: HasErrors=true; want false")
	}
}

func TestCheckMissingFrontmatterBlock(t *testing.T) {
	// No '---' fences at all -> HasFM false, err nil -> ONE root-cause ERROR.
	s := mkSkill(t, "bare", "# just a heading\nno frontmatter here\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || fs[0].Level != LevelError || !strings.Contains(fs[0].Message, "missing frontmatter block") {
		t.Errorf("no-block skill should be one 'missing frontmatter block' ERROR; got %+v", fs)
	}
	if rep.Errors != 1 {
		t.Errorf("Errors=%d; want 1", rep.Errors)
	}
}

func TestCheckMalformedYAML(t *testing.T) {
	// Broken YAML between valid fences -> ParseFrontmatter returns err.
	s := mkSkill(t, "broken", "---\nname: [unclosed\n---\n# body\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || fs[0].Level != LevelError || !strings.Contains(fs[0].Message, "invalid SKILL.md frontmatter") {
		t.Errorf("malformed YAML should be one 'invalid SKILL.md frontmatter' ERROR; got %+v", fs)
	}
}

func TestCheckMissingName(t *testing.T) {
	s := mkSkill(t, "a", "---\ndescription: has desc but no name\n---\nx\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || !strings.Contains(fs[0].Message, "'name' is missing") {
		t.Errorf("missing name -> one 'name is missing' ERROR; got %+v", fs)
	}
}

func TestCheckMissingDescription(t *testing.T) {
	s := mkSkill(t, "a", "---\nname: a\n---\nx\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || !strings.Contains(fs[0].Message, "'description' is missing or empty") {
		t.Errorf("missing description -> one ERROR; got %+v", fs)
	}
}

func TestCheckEmptyDescription(t *testing.T) {
	// Whitespace-only description trims to "" -> ERROR (not a WARN).
	s := mkSkill(t, "a", "---\nname: a\ndescription: \"   \"\n---\nx\n")
	rep := Check([]discover.Skill{s})
	if rep.Errors != 1 || !strings.Contains(rep.BySkill[0].Findings[0].Message, "description") {
		t.Errorf("whitespace-only description -> one description ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameLeadingHyphen(t *testing.T) {
	s := skill(t, "a", "-foo", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 || !strings.Contains(rep.BySkill[0].Findings[0].Message, "lowercase a-z0-9") {
		t.Errorf("leading hyphen -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameTrailingHyphen(t *testing.T) {
	s := skill(t, "a", "foo-", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("trailing hyphen -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameConsecutiveHyphens(t *testing.T) {
	s := skill(t, "a", "foo--bar", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("consecutive hyphens -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameUppercase(t *testing.T) {
	s := skill(t, "a", "Foo", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("uppercase name -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameTooLong(t *testing.T) {
	long := repeat("a", 65) // 65 chars, otherwise valid
	s := skill(t, "a", long, "d")
	rep := Check([]discover.Skill{s})
	if rep.Errors != 1 {
		t.Errorf("65-char name -> 1 ERROR; got %d: %+v", rep.Errors, rep.BySkill[0].Findings)
	}
	if !strings.Contains(rep.BySkill[0].Findings[0].Message, "65 chars (max 64)") {
		t.Errorf("too-long message should name the length; got %q", rep.BySkill[0].Findings[0].Message)
	}
}

func TestCheckNameAtLimitOK(t *testing.T) {
	at := repeat("a", 64) // exactly 64, valid -> OK
	s := skill(t, "a", at, "d")
	if rep := Check([]discover.Skill{s}); rep.Errors != 0 {
		t.Errorf("64-char valid name should be OK; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckDescriptionTooLongWarns(t *testing.T) {
	long := repeat("x", 1025)
	s := skill(t, "a", "a", long)
	rep := Check([]discover.Skill{s})
	if rep.Errors != 0 {
		t.Errorf("over-long description is a WARN, not an ERROR; Errors=%d", rep.Errors)
	}
	if rep.Warnings != 1 || rep.HasErrors() {
		t.Errorf("expected 1 warning, no errors; got Warnings=%d HasErrors=%v", rep.Warnings, rep.HasErrors())
	}
	if !strings.Contains(rep.BySkill[0].Findings[0].Message, "1025 chars (max 1024)") {
		t.Errorf("WARN should name the length; got %q", rep.BySkill[0].Findings[0].Message)
	}
}

func TestCheckDescriptionAtLimitOK(t *testing.T) {
	at := repeat("x", 1024) // exactly 1024 -> no WARN
	s := skill(t, "a", "a", at)
	if rep := Check([]discover.Skill{s}); rep.Warnings != 0 || rep.Errors != 0 {
		t.Errorf("1024-char description should be clean; got W=%d E=%d", rep.Warnings, rep.Errors)
	}
}

func TestCheckDuplicateNames(t *testing.T) {
	a := skill(t, "alpha", "shared", "d")
	b := skill(t, "beta", "shared", "d")
	rep := Check([]discover.Skill{a, b})
	if rep.Errors != 2 {
		t.Errorf("two skills sharing a name -> 2 ERRORs; got %d", rep.Errors)
	}
}

func TestCheckDupMessageNamesOtherTag(t *testing.T) {
	a := skill(t, "alpha", "shared", "d")
	b := skill(t, "beta", "shared", "d")
	rep := Check([]discover.Skill{a, b})
	// alpha's dup ERROR must name beta (sorted "also in" list), and vice versa.
	alphaMsg := rep.BySkill[0].Findings[0].Message
	betaMsg := rep.BySkill[1].Findings[0].Message
	if !strings.Contains(alphaMsg, "beta") || !strings.Contains(alphaMsg, "duplicate") {
		t.Errorf("alpha dup message should name beta: %q", alphaMsg)
	}
	if !strings.Contains(betaMsg, "alpha") {
		t.Errorf("beta dup message should name alpha: %q", betaMsg)
	}
}

func TestCheckMissingNameNotCountedAsDup(t *testing.T) {
	// A skill with NO name must NOT participate in the dup scan (it has its own
	// missing-name ERROR). Here one skill has name "x", another has no name:
	// no dup ERROR should appear, only the single missing-name ERROR.
	withName := skill(t, "alpha", "x", "d")
	noName := mkSkill(t, "beta", "---\ndescription: d\n---\nx\n")
	rep := Check([]discover.Skill{withName, noName})
	for _, f := range rep.BySkill[0].Findings {
		if strings.Contains(f.Message, "duplicate") {
			t.Errorf("alpha should have NO dup ERROR (no other 'x'); got %q", f.Message)
		}
	}
	if rep.Errors != 1 { // only the missing-name ERROR on beta
		t.Errorf("expected exactly 1 ERROR (missing name on beta); got %d", rep.Errors)
	}
}

func TestCheckEmptyInputNoPanic(t *testing.T) {
	rep := Check(nil)
	if rep.Errors != 0 || rep.Warnings != 0 || len(rep.BySkill) != 0 {
		t.Errorf("Check(nil) should be empty clean report; got %+v", rep)
	}
	if rep.HasErrors() {
		t.Errorf("Check(nil).HasErrors()=true; want false")
	}
}
```

### File 3 — MODIFY `main.go` (three localized edits + one import)

The file is large and otherwise unchanged; apply these EXACT edits. Run
`gofmt -w main.go` after (it will realign the `config` struct comments and place the
new import first — both expected and correct).

**Edit 3a — imports** (add `internal/check`; gofmt places it FIRST, before `discover`):

```go
// OLD:
	"github.com/dabstractor/skpp/internal/discover"
	"github.com/dabstractor/skpp/internal/resolve"
	"github.com/dabstractor/skpp/internal/search"
	"github.com/dabstractor/skpp/internal/skillsdir"
	"github.com/dabstractor/skpp/internal/ui"

// NEW:
	"github.com/dabstractor/skpp/internal/check"
	"github.com/dabstractor/skpp/internal/discover"
	"github.com/dabstractor/skpp/internal/resolve"
	"github.com/dabstractor/skpp/internal/search"
	"github.com/dabstractor/skpp/internal/skillsdir"
	"github.com/dabstractor/skpp/internal/ui"
```

**Edit 3b — `config` struct** (add the `check` field; drop `check bool` from the
"future" comment):

```go
// OLD (the tail of the config struct):
	searchMode bool     // --search <q>/-s : substring search over tag/name/description/keywords (§6.1) [NEW]
	searchQ    string   // the --search query value (consumed from the token after --search/-s) [NEW]
	// Future (M5), do NOT add yet:
	//   check bool; help bool
}

// NEW (gofmt realigns the whole struct block; the field set is what matters):
	searchMode bool     // --search <q>/-s : substring search over tag/name/description/keywords (§6.1)
	searchQ    string   // the --search query value (consumed from the token after --search/-s)
	check      bool     // `skpp check` subcommand: validate every skill in the store (§9) [NEW]
	// Future (M5), do NOT add yet:
	//   help bool
}
```

**Edit 3c — `parseArgs`** (add the `check` subcommand case BEFORE the default
tag-capture branch; every other `case` is unchanged):

```go
// OLD (the default branch at the end of the switch):
		case "--search", "-s":
			// ... existing --search case unchanged ...
		default:
			// Positional <tag> ... (the existing comment, unchanged)
			if !strings.HasPrefix(a, "-") {
				c.tags = append(c.tags, a)
			}
		}

// NEW (insert a new `case "check":` immediately before `default:`):
		case "--search", "-s":
			// ... existing --search case unchanged ...
		case "check":
			// `skpp check` subcommand (PRD §9). `check` is a RESERVED positional
			// token: it selects validation mode and is NOT captured as a tag. A
			// skill literally tagged `check` cannot be resolved via `skpp check`
			// (subcommand names are reserved, as in any CLI). P1.M5.T11 turns
			// `check` mixed with tags/--list/--search/--all into a §6.3 exit-2
			// error; for now check wins silently in run() dispatch (mirrors how
			// searchMode currently wins over tags).
			c.check = true
		default:
			// Positional <tag> ... (unchanged comment)
			if !strings.HasPrefix(a, "-") {
				c.tags = append(c.tags, a)
			}
		}
```

> The existing `--search`/`-s` case and its body stay byte-identical; only the new
> `case "check":` is inserted between it and `default:`. Keep the existing
> `default:`-branch comment verbatim.

**Edit 3d — `run` dispatch** (insert the check branch AFTER the `if c.searchMode {…}`
block and BEFORE the `// --all mode:` block):

```go
// INSERT THIS BLOCK between the --search block and the --all block:

	// `skpp check` subcommand (PRD §9). Validates every skill in the store and
	// prints a report: one line per problem (prefixed ERROR/WARN) plus one OK line
	// per clean skill, ending with a "N skills, M errors, K warnings" summary. Exit
	// 0 if there are no ERRORs, 1 if there are any (WARNs never change the exit
	// code, so `if skpp check; then …` works as a gate). An empty store is clean
	// (0 skills, 0 errors, 0 warnings) -> exit 0 (check is validation: no skills ==
	// nothing wrong, unlike --list which exits 1 on empty).
	//
	// check is a REPORT, not a path emitter: it always prints its full findings to
	// STDOUT (pipeable to less/grep, like eslint/ruff/govet) and signals pass/fail
	// via the exit code. It is NOT subject to §6.4's "nothing on stdout on failure"
	// — that contract is for tag/path emitters used inside $(...); check never
	// participates in command substitution.
	//
	// internal/check.Check re-runs discover.ParseFrontmatter per skill to recover
	// the malformed-YAML-vs-no-frontmatter-block distinction that discover.Index
	// intentionally drops (index.go doc comment). --file/--relative/--no-color do
	// NOT apply (status report, not paths/table).
	if c.check {
		dir, _, err := skillsdir.Find()
		if err != nil {
			fmt.Fprintln(stderr, err) // one-line fix (PRD §6.4/§8); stdout stays empty
			return 1
		}
		skills, err := discover.Index(dir)
		if err != nil {
			fmt.Fprintln(stderr, err) // e.g. skills dir vanished between Find and Index
			return 1
		}
		rep := check.Check(skills)
		// Render: status word left-padded to width 5 (OK/WARN/ERROR align); OK
		// skills get one line, problem skills get one line per finding.
		for _, sr := range rep.BySkill {
			name := sr.Skill.Name
			if name == "" {
				name = "(none)"
			}
			if len(sr.Findings) == 0 {
				fmt.Fprintf(stdout, "%-5s %s (%s)\n", "OK", sr.Skill.RelTag, name)
				continue
			}
			for _, f := range sr.Findings {
				fmt.Fprintf(stdout, "%-5s %s (%s): %s\n", f.Level, sr.Skill.RelTag, name, f.Message)
			}
		}
		fmt.Fprintf(stdout, "%d skills, %d errors, %d warnings\n", len(skills), rep.Errors, rep.Warnings)
		if rep.HasErrors() {
			return 1
		}
		return 0
	}
```

### File 4 — MODIFY `main_test.go` (append these 12 tests + fix 1 existing test;
the other 68 are byte-identical)

These reuse the existing `sampleStore`, `writeSkillTree`, and `unsetSkillsEnv`
helpers (already defined in `main_test.go`). The FIRST edit (4a) fixes one
existing test that collides with the now-reserved `check` subcommand; edits 4b
appends the 12 new tests at the end of the file.

**Edit 4a — MODIFY the existing `TestParseArgsUnknownTolerated` (collision fix).**
The current test feeds `["--frobnicate", "sometag", "check"]` and asserts
c.tags == `[sometag check]`. Under PRD §6.1 `check` is a RESERVED subcommand, so
this task's `case "check":` (Edit 3c) means `check` is NO LONGER captured as a
tag. The test's INTENT ("non-dashed positionals are captured as tags; dashed
unknowns are excluded") is preserved by swapping `check` for a non-reserved tag.
Find the existing function and apply these three replacements verbatim:

```go
// OLD (the doc comment's second sentence):
// tokens are now captured as <tag>s (so "sometag"/"check" land in c.tags rather
// NEW:
// tokens are now captured as <tag>s ("sometag"/"othertag" here). `check` is now a
// RESERVED subcommand (P1.M4.T10.S1) and is NOT captured as a tag, so it is
// deliberately excluded from this positional-capture test.

// OLD (the parseArgs call):
	c := parseArgs([]string{"--frobnicate", "sometag", "check"})
// NEW:
	c := parseArgs([]string{"--frobnicate", "sometag", "othertag"})

// OLD (the assertion):
	if len(c.tags) != 2 || c.tags[0] != "sometag" || c.tags[1] != "check" {
		t.Errorf("parseArgs tags=%v; want [sometag check] (positionals captured)", c.tags)
	}
// NEW:
	if len(c.tags) != 2 || c.tags[0] != "sometag" || c.tags[1] != "othertag" {
		t.Errorf("parseArgs tags=%v; want [sometag othertag] (positionals captured)", c.tags)
	}
```

> This is the ONLY change to an existing test. It is mandatory: without it,
> `go test .` fails on `TestParseArgsUnknownTolerated` (it would still expect
> `check` in c.tags). Verified: with the swap, the full suite is green.

**Edit 4b — APPEND the 12 new tests** (existing 68 byte-identical, 1 fixed in 4a).
Append at end of file.

```go
// --- parseArgs: `check` subcommand (P1.M4.T10.S1) ---

// The bare token "check" selects the check subcommand and is NOT captured as a tag.
func TestParseArgsCheckSubcommand(t *testing.T) {
	c := parseArgs([]string{"check"})
	if !c.check {
		t.Errorf("parseArgs(check): check=false; want true")
	}
	if len(c.tags) != 0 {
		t.Errorf("parseArgs(check): tags=%v; want empty ('check' is a subcommand, not a tag)", c.tags)
	}
}

// `check` is recognized even when it follows a flag (--no-color check).
func TestParseArgsCheckAfterFlag(t *testing.T) {
	c := parseArgs([]string{"--no-color", "check"})
	if !c.check {
		t.Errorf("parseArgs(--no-color check): check=false; want true")
	}
	if !c.noColor {
		t.Errorf("parseArgs(--no-color check): noColor=false; want true (flag still parsed)")
	}
}

// `check` + a later positional: check wins in dispatch (pre-M5; M5 makes this exit 2).
// Here we only assert both are captured as set; run() ordering is tested below.
func TestParseArgsCheckAndTagBothCaptured(t *testing.T) {
	c := parseArgs([]string{"check", "sometag"})
	if !c.check {
		t.Errorf("check not set: %+v", c)
	}
	if len(c.tags) != 1 || c.tags[0] != "sometag" {
		t.Errorf("tags=%v; want [sometag]", c.tags)
	}
}

// --- run: `skpp check` (P1.M4.T10.S1) ---

// Clean store -> one OK line per skill + summary, exit 0, no ANSI, empty stderr.
// sampleStore has example + writing/reddit, both valid.
func TestRunCheckCleanStore(t *testing.T) {
	dir := sampleStore(t)
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(check) clean: code=%d; want 0", code)
	}
	got := out.String()
	if !strings.Contains(got, "OK") {
		t.Errorf("clean store should have OK lines:\n%s", got)
	}
	if !strings.Contains(got, "example") || !strings.Contains(got, "writing/reddit") {
		t.Errorf("both skills should appear:\n%s", got)
	}
	if !strings.Contains(got, "2 skills, 0 errors, 0 warnings") {
		t.Errorf("summary line missing/wrong:\n%s", got)
	}
	if strings.Contains(got, "ERROR") || strings.Contains(got, "WARN") {
		t.Errorf("clean store should have no ERROR/WARN lines:\n%s", got)
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr=%q; want empty", errOut.String())
	}
}

// A store with a missing-name skill -> ERROR line on STDOUT + exit 1. Full report
// still prints (check is a report: pass/fail is the exit code, NOT stdout emptiness).
func TestRunCheckReportsMissingNameExit1(t *testing.T) {
	dir := writeSkillTree(t, map[string]string{
		"example": "---\nname: example\ndescription: d\n---\nx\n",
		"bad":     "---\ndescription: no name here\n---\nx\n",
	})
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("run(check) with a bad skill: code=%d; want 1", code)
	}
	got := out.String()
	if !strings.Contains(got, "ERROR") || !strings.Contains(got, "'name' is missing") {
		t.Errorf("stdout should report the missing-name ERROR:\n%s", got)
	}
	if !strings.Contains(got, "1 errors") {
		t.Errorf("summary should count 1 error:\n%s", got)
	}
}

// Duplicate names across skills -> two ERROR lines (one per owner) + exit 1.
func TestRunCheckReportsDuplicateNames(t *testing.T) {
	dir := writeSkillTree(t, map[string]string{
		"alpha": "---\nname: dup\nmetadata:\n  category: x\n---\nx\n",
		"beta":  "---\nname: dup\nmetadata:\n  category: x\n---\nx\n",
	})
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("run(check) dups: code=%d; want 1", code)
	}
	got := out.String()
	if !strings.Contains(got, "duplicate") {
		t.Errorf("stdout should report duplicate-name ERRORs:\n%s", got)
	}
	// Both skills lack a description -> that's 2 more ERRORs; total >= 2 errors.
	if !strings.Contains(got, "errors") {
		t.Errorf("summary line missing:\n%s", got)
	}
}

// A WARN-only problem (over-long description) -> WARN line but exit 0 (WARNs never
// fail). Proves the exit code is driven by ERRORs only.
func TestRunCheckWarnOnlyExitsZero(t *testing.T) {
	long := strings.Repeat("x", 1025)
	dir := writeSkillTree(t, map[string]string{
		"big": "---\nname: big\ndescription: " + long + "\n---\nx\n",
	})
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(check) warn-only: code=%d; want 0 (WARNs never fail)", code)
	}
	got := out.String()
	if !strings.Contains(got, "WARN") || !strings.Contains(got, "1025 chars") {
		t.Errorf("stdout should have the over-long WARN:\n%s", got)
	}
	if !strings.Contains(got, "0 errors, 1 warnings") {
		t.Errorf("summary should be 0 errors / 1 warning:\n%s", got)
	}
}

// Empty store -> 0 skills / 0 errors / 0 warnings, exit 0 (clean, unlike --list).
func TestRunCheckEmptyStoreExit0(t *testing.T) {
	dir := writeSkillTree(t, map[string]string{}) // empty skills tree
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(check) empty store: code=%d; want 0", code)
	}
	if got := out.String(); !strings.Contains(got, "0 skills, 0 errors, 0 warnings") {
		t.Errorf("empty store summary wrong:\n%s", got)
	}
}

// Skills dir unresolvable -> exit 1, EMPTY stdout, one-line fix on stderr.
func TestRunCheckSkillsDirUnresolvable(t *testing.T) {
	unsetSkillsEnv(t)
	t.Chdir(t.TempDir()) // all three §8 rules miss
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("run(check) unresolvable: code=%d; want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty (no store -> no report)", out.String())
	}
	if !strings.Contains(errOut.String(), "SKPP_SKILLS_DIR") {
		t.Errorf("stderr=%q; want the one-line fix", errOut.String())
	}
}

// Status column alignment: OK/ERROR/WARN all pad to width 5.
func TestRunCheckStatusColumnAligned(t *testing.T) {
	dir := writeSkillTree(t, map[string]string{
		"good": "---\nname: good\ndescription: d\n---\nx\n",
		"bad":  "---\ndescription: missing name\n---\nx\n",
	})
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	run([]string{"check"}, &out, &errOut)
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "0 ") || strings.Contains(line, " skills,") {
			continue // summary line
		}
		// Every status line starts with a 5-wide status word + a single space.
		switch {
		case strings.HasPrefix(line, "OK    "):
		case strings.HasPrefix(line, "ERROR "):
		case strings.HasPrefix(line, "WARN  "):
		default:
			t.Errorf("status line not 5-wide aligned: %q", line)
		}
	}
}

// --version precedes `check` (PRD §6.3).
func TestRunVersionPrecedenceOverCheck(t *testing.T) {
	dir := sampleStore(t)
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check", "--version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(check --version): code=%d; want 0 (version precedence)", code)
	}
	if got := out.String(); got != "skpp "+version+"\n" {
		t.Errorf("stdout=%q; want the version line (precedence over check)", got)
	}
}

// A pre-existing tag-resolution test guard: `check` is reserved, so a real skill
// tagged `example` still resolves (the subcommand only steals the literal "check").
func TestRunTagStillResolvesAlongsideCheck(t *testing.T) {
	dir := sampleStore(t)
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"example"}, &out, &errOut) // NOT "check" -> tag resolution
	if code != 0 {
		t.Fatalf("run(example): code=%d; want 0 (tag resolution unaffected)", code)
	}
	if !strings.HasSuffix(out.String(), "/example\n") {
		t.Errorf("run(example) stdout=%q; want .../example dir", out.String())
	}
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: PRECONDITION — confirm dependencies are on disk and green
  - COMMAND: cd /home/dustin/projects/skpp
  - COMMAND: grep -q 'func ParseFrontmatter(path string)' internal/discover/discover.go
  - COMMAND: grep -q 'func Index(skillsDir string)' internal/discover/index.go
  - COMMAND: grep -q 'type Skill struct' internal/discover/skill.go
  - COMMAND: grep -q 'SourceFile' internal/discover/skill.go
  - COMMAND: go test ./internal/discover/ ./internal/search/ ./internal/resolve/ ./internal/skillsdir/ ./internal/ui/ >/dev/null && echo "deps green"
  - EXPECT: all symbols exist AND those packages pass. If ParseFrontmatter/Index/Skill
            are MISSING, M2 has NOT landed — STOP and let it land first.
  - COMMAND: go test . -count=1 >/dev/null && echo "main green (69 tests)"
  - EXPECT: baseline main green so the test delta (69 -> 81) is attributable to THIS task.
  - COMMAND: grep -n '"check"' main_test.go   # find existing tests using "check"
  - EXPECT: exactly ONE pre-existing hit — `TestParseArgsUnknownTolerated` (it feeds
            `"check"` as a positional tag and asserts it lands in c.tags). This is the
            KNOWN collision with the now-reserved `check` subcommand; Edit 4a (File 4)
            fixes it by swapping `check` -> `othertag`. Apply Edit 4a; do NOT skip it.

Task 1: CREATE internal/check/check.go
  - WRITE: the exact content from Blueprint File 1.
  - CHECK: `package check`; imports ONLY fmt/regexp/sort/strings + internal/discover;
           func Check + types Report/SkillReport/Finding/Severity + checkFields +
           appendDupFindings + the validName regex + nameLenMax/descLenMax consts.
  - GOTCHA: re-parse via discover.ParseFrontmatter(s.SourceFile) to split malformed-YAML
            (err!=nil) from no-block (HasFM false, err nil). Field checks ONLY when
            HasFM && err==nil. Missing-block/malformed => ONE root-cause ERROR each.
  - GOTCHA: description length on strings.TrimSpace(fm.Description); name length on
            len(fm.Name) (ASCII). len>64 ERROR BEFORE the regex (regex can't express max).
  - GOTCHA: dup scan is a 2nd pass over non-empty names; each owner gets its own ERROR.

Task 2: CREATE internal/check/check_test.go
  - WRITE: the exact content from Blueprint File 2.
  - CHECK: `package check`; imports os/path/filepath/strings/testing + internal/discover;
           helper mkSkill/skill/repeat; 18 tests incl. both boundaries (64/1024), the
           5 name rules, malformed-vs-no-block distinction, dup cross-tag message, and
           the missing-name-not-a-dup guard.
  - GOTCHA: NO testify; NO t.Parallel(); mkSkill writes real temp SKILL.md files (check
            re-parses SourceFile, so the files must exist). mkSkill mirrors Index: parse,
            ignore err, BuildSkill.

Task 3: MODIFY main.go (apply Edits 3a-3d)
  - EDIT 3a: add "github.com/dabstractor/skpp/internal/check" to the import group (gofmt
             puts it FIRST, before discover).
  - EDIT 3b: add `check bool` to config; change the "Future" comment to "help bool" only.
  - EDIT 3c: add `case "check": c.check = true` in parseArgs, BEFORE the default tag-capture.
  - EDIT 3d: insert the `if c.check {…}` branch AFTER the --search block, BEFORE --all.
  - CHECK: version precedence still first; search/list/all/tags/default blocks unchanged.
  - GOTCHA: importing `check` does not collide with c.check or case "check":. Run
            `gofmt -w main.go` — it realigns config comments and orders the import.

Task 4: MODIFY main_test.go (Edit 4a collision fix + Edit 4b append)
  - EDIT 4a: MODIFY the existing `TestParseArgsUnknownTolerated` per File 4 Edit 4a —
             swap the `"check"` positional for `"othertag"` (input + assertion + the
             doc-comment sentence naming "check"). The other 68 existing tests are
             byte-identical. MANDATORY: without it, `go test .` fails.
  - EDIT 4b: APPEND the 12 named tests at end of file (3 parseArgs + 9 run).
  - CHECK: reuses sampleStore/writeSkillTree/unsetSkillsEnv; existing 68 tests
           unchanged; 1 existing test (TestParseArgsUnknownTolerated) MODIFIED
           (it used the now-reserved `check` token as a positional tag — see Edit 4a).
           Net main count: 69 -> 81 (+12 new, 1 modified in place).
  - GOTCHA: do NOT duplicate helper definitions (they already exist). `bytes`/`strings`/
            `filepath` are already imported in main_test.go — do NOT re-import.

Task 5: FORMAT + VET + TIDY + BUILD + TEST (validation gates — run in order)
  - COMMAND: gofmt -w internal/check/*.go main.go main_test.go
  - COMMAND: gofmt -l internal/check/*.go main.go main_test.go   # MUST print nothing
  - COMMAND: go vet ./...                                         # MUST be clean
  - COMMAND: go mod tidy   # EXPECTED: a NO-OP (stdlib only; no new module)
  - COMMAND: go build ./...                                       # exit 0
  - COMMAND: go test ./internal/check/ -v                        # 18 NEW tests PASS
  - COMMAND: go test . -v                                         # 81 main tests (69 old + 12 new)
  - COMMAND: go test ./...                                        # whole module green (196 total)
  - EXPECT: zero errors, zero vet findings, gofmt silent, go.mod/go.sum unchanged.

Task 6: SMOKE + SCOPE CHECK — Levels 3 + 4 in the Validation Loop.
  - COMMAND: the Level 3 block (build, check over a clean/empty/bad/dup/warn store,
            unresolvable path, exit codes).
  - COMMAND: the Level 4 block (scope boundaries + go.mod unchanged + the five
            READ-ONLY packages untouched).
```

### Implementation Patterns & Key Details

```go
// The WHOLE feature is: re-parse each skill, classify, collect, render.

// internal/check/check.go — the core (pure-ish: re-reads SKILL.md, returns data):
func Check(skills []discover.Skill) Report {
    var rep Report
    rep.BySkill = make([]SkillReport, len(skills))
    var owners []nameOwner
    for i := range skills {                              // PASS 1: per-skill local checks
        s := skills[i]
        fm, _, perr := discover.ParseFrontmatter(s.SourceFile) // recover the err Index drops
        switch {
        case perr != nil:
            findings = []Finding{{LevelError, "invalid SKILL.md frontmatter: " + perr.Error()}}
        case !fm.HasFM:
            findings = []Finding{{LevelError, "missing frontmatter block (no '---' delimiters)"}}
        default:
            findings = checkFields(fm)                  // name + description rules
            if fm.Name != "" { owners = append(owners, nameOwner{s.RelTag, fm.Name}) }
        }
        rep.BySkill[i] = SkillReport{Skill: s, Findings: findings}
    }
    appendDupFindings(&rep, owners)                     // PASS 2: global dup-name scan
    for /* each finding */ { /* PASS 3: tally Errors/Warnings */ }
    return rep
}

// checkFields — name THEN description (each failure appends its own Finding):
func checkFields(fm discover.Frontmatter) []Finding {
    var f []Finding
    switch {
    case fm.Name == "":                  f = append(f, Finding{LevelError, "...name is missing"})
    case len(fm.Name) > 64:              f = append(f, Finding{LevelError, "...N chars (max 64)"})
    case !validName.MatchString(fm.Name):f = append(f, Finding{LevelError, "...lowercase a-z0-9..."})
    }
    desc := strings.TrimSpace(fm.Description)
    switch {
    case desc == "":           f = append(f, Finding{LevelError, "...description is missing or empty"})
    case len(desc) > 1024:     f = append(f, Finding{LevelWarn, "...N chars (max 1024)"})
    }
    return f
}

// main.go run() — the dispatch branch (renders the structured Report):
if c.check {
    dir, _, err := skillsdir.Find()
    if err != nil { fmt.Fprintln(stderr, err); return 1 }     // one-line fix; empty stdout
    skills, err := discover.Index(dir)
    if err != nil { fmt.Fprintln(stderr, err); return 1 }
    rep := check.Check(skills)                                // re-parse + classify
    for _, sr := range rep.BySkill {
        name := sr.Skill.Name; if name == "" { name = "(none)" }
        if len(sr.Findings) == 0 {
            fmt.Fprintf(stdout, "%-5s %s (%s)\n", "OK", sr.Skill.RelTag, name)   // one OK line
            continue
        }
        for _, f := range sr.Findings {
            fmt.Fprintf(stdout, "%-5s %s (%s): %s\n", f.Level, sr.Skill.RelTag, name, f.Message)
        }
    }
    fmt.Fprintf(stdout, "%d skills, %d errors, %d warnings\n", len(skills), rep.Errors, rep.Warnings)
    if rep.HasErrors() { return 1 }                           // ERRORs drive exit; WARNs don't
    return 0
}

// parseArgs — the new subcommand case (before the default tag-capture):
case "check":
    c.check = true   // RESERVED positional token; NOT appended to c.tags
```

### Integration Points

```yaml
NEW PACKAGE:
  - path: internal/check/
  - exports: func Check(skills []discover.Skill) Report; type Report/SkillReport/Finding/Severity
  - deps:    fmt, regexp, sort, strings (stdlib) + github.com/dabstractor/skpp/internal/discover (read-only)
  - tests:   internal/check/check_test.go (white-box, 18 tests, temp SKILL.md files)

MAIN DISPATCH:
  - file: main.go
  - insert: the `if c.check {…}` branch AFTER the `--search` block, BEFORE `--all`
  - precedence: version -> path -> list -> search -> CHECK -> all -> tags -> default (§6.3 exclusivity is M5)

CLI SURFACE:
  - subcommand: `skpp check`  (positional token "check"; reserved)
  - modifiers that apply: (none — status report, not paths/table)
  - modifiers that do NOT apply: "--file", "--relative", "--no-color"
  - exit codes: 0 if no ERROR; 1 if any ERROR (WARNs never fail); 1 if store unresolvable

CONFIG:
  - struct main.config gains: check bool

DEPENDENCIES (go.mod): UNCHANGED. No new module; stdlib only. `go mod tidy` is a no-op.
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
gofmt -w internal/check/*.go main.go main_test.go
gofmt -l internal/check/*.go main.go main_test.go   # MUST print nothing
go vet ./...                                          # MUST be clean
# go mod tidy   # OPTIONAL sanity check — EXPECTED to be a no-op (diff go.mod before/after)
go build ./...                                        # exit 0
# Expected: zero output from gofmt -l; zero vet findings; build succeeds.
```

### Level 2: Unit Tests (Component Validation)

```bash
# The new validation package — re-parses temp SKILL.md files; fastest deep feedback.
go test ./internal/check/ -v
# Expected: 18 tests PASS (TestCheck*). If any fail, the rule logic is wrong;
#   read the failure (it names the rule/scenario) and fix check.go. Pay attention
#   to the two boundary tests (name 64, description 1024) and the malformed-vs-no-block
#   distinction (TestCheckMalformedYAML vs TestCheckMissingFrontmatterBlock).

# The dispatcher + parser — uses temp skill trees via the existing helpers.
go test . -run 'Check|VersionPrecedenceOverCheck|TagStillResolves' -v
# Expected: the 12 new main tests PASS (3 parseArgs + 9 run).

# Full suite — confirms NO regression in the five read-only packages.
go test ./... -count=1
# Expected: ALL PASS. Totals: . = 81 (69+12), discover = 31, resolve = 10, search = 16,
#   skillsdir = 29, ui = 11, check = 18. Grand total = 196.
```

### Level 3: Integration Testing (System Validation)

```bash
cd /home/dustin/projects/skpp
go build -o skpp . && echo OK
./skpp --version                       # prints: skpp <something>

# Build a throwaway store with one clean + several broken skills (§8 rule 1 via env).
TMPROOT=$(mktemp -d)
mkdir -p "$TMPROOT"
cat > "$TMPROOT/good/SKILL.md" <<'EOF'
---
name: good
description: A valid skill.
---
# Good
EOF
cat > "$TMPROOT/noname/SKILL.md" <<'EOF'
---
description: Missing the name field.
---
# Bad
EOF
cat > "$TMPROOT/badname/SKILL.md" <<'EOF'
---
name: Bad_Name
description: Invalid name charset.
---
# Bad
EOF
cat > "$TMPROOT/noblock/SKILL.md" <<'EOF'
# No frontmatter here at all
EOF
printf -- '---\nname: big\ndescription: %s\n---\nx\n' "$(head -c 1025 < /dev/zero | tr '\0' 'x')" > "$TMPROOT/big/SKILL.md" 2>/dev/null || mkdir -p "$TMPROOT/big" && printf -- '---\nname: big\ndescription: %s\n---\nx\n' "$(python3 -c "print('x'*1025)")" > "$TMPROOT/big/SKILL.md"

# 1) check exits 1 (several ERRORs), prints the full report to STDOUT.
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check; rc=$?
[ "$rc" = "1" ] && echo "exit 1 on errors OK"

# 2) each problem is reported with its tag + reason.
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q "noname.*name.*missing" && echo "missing-name reported OK"
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q "badname.*lowercase a-z0-9" && echo "invalid-name reported OK"
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q "noblock.*missing frontmatter block" && echo "no-block reported OK"
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q "big.*WARN.*1025 chars" && echo "over-long WARN reported OK"

# 3) the clean skill shows as OK.
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q "^OK.*good (good)" && echo "clean skill OK line OK"

# 4) summary line present and counts errors>0.
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -qE '[0-9]+ skills, [0-9]+ errors, [0-9]+ warnings' && echo "summary line OK"

# 5) WARNs never change the exit code: a WARN-only store exits 0.
WROOT=$(mktemp -d)
mkdir -p "$WROOT/big"
printf -- '---\nname: big\ndescription: %s\n---\nx\n' "$(python3 -c "print('x'*1025)")" > "$WROOT/big/SKILL.md"
SKPP_SKILLS_DIR="$WROOT" ./skpp check; rc=$?
[ "$rc" = "0" ] && echo "warn-only exits 0 OK"

# 6) clean store exits 0.
CROOT=$(mktemp -d)
mkdir -p "$CROOT/example"
cat > "$CROOT/example/SKILL.md" <<'EOF'
---
name: example
description: Clean.
---
EOF
SKPP_SKILLS_DIR="$CROOT" ./skpp check >/dev/null && echo "clean store exit 0 OK"

# 7) empty store -> 0 skills, exit 0.
EROOT=$(mktemp -d)
SKPP_SKILLS_DIR="$EROOT" ./skpp check | grep -q "0 skills, 0 errors, 0 warnings" && echo "empty store OK"

# 8) unresolvable store -> exit 1, EMPTY stdout, fix on stderr.
out=$(SKPP_SKILLS_DIR=/does/not/exist ./skpp check 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "1" ] && echo "unresolvable exit 1 / empty stdout OK"

# 9) `if skpp check; then` gating works (clean -> enters the branch).
if SKPP_SKILLS_DIR="$CROOT" ./skpp check >/dev/null; then echo "if-gate clean OK"; fi

# 10) duplicate names -> two ERROR lines.
DROOT=$(mktemp -d)
mkdir -p "$DROOT/alpha" "$DROOT/beta"
printf -- '---\nname: shared\ndescription: d\n---\nx\n' > "$DROOT/alpha/SKILL.md"
printf -- '---\nname: shared\ndescription: d\n---\nx\n' > "$DROOT/beta/SKILL.md"
n=$(SKPP_SKILLS_DIR="$DROOT" ./skpp check | grep -c 'duplicate'); [ "$n" = "2" ] && echo "dup names reported x2 OK"

rm -rf "$TMPROOT" "$WROOT" "$CROOT" "$EROOT" "$DROOT"
# Expected: every "… OK" line prints; the error store exits 1 with a full report; the
#   warn-only and clean stores exit 0; the unresolvable store empties stdout and exits 1.
```

### Level 4: Creative & Domain-Specific Validation (Scope Boundaries)

```bash
cd /home/dustin/projects/skpp

# SCOPE: go.mod / go.sum UNCHANGED (no new dependency).
git diff --name-only go.mod go.sum | (! read) && echo "go.mod/go.sum unchanged OK"
# (Or: `git diff --exit-code go.mod go.sum` prints nothing and exits 0.)

# SCOPE: the five READ-ONLY packages are untouched.
git diff --name-only internal/discover internal/skillsdir internal/resolve internal/ui internal/search | (! read) && echo "read-only packages untouched OK"

# CONTRACT: check prints to STDOUT (a report), exit code is the signal.
TMPROOT=$(mktemp -d); mkdir -p "$TMPROOT/bad"
printf -- '%s\n' '---' 'description: no name' '---' 'x' > "$TMPROOT/bad/SKILL.md"
report=$(SKPP_SKILLS_DIR="$TMPROOT" ./skpp check 2>/dev/null)
[ -n "$report" ] && echo "check prints report to stdout even on failure OK"
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check >/dev/null 2>&1; [ $? = "1" ] && echo "check exit 1 on error OK"
rm -rf "$TMPROOT"

# BOUNDARY: the two reframed §9 bullets behave as decided.
#   - "no SKILL.md dir" heuristic NOT added (grouping dirs would false-positive).
#   - "empty besides SKILL.md" WARN NOT added (would WARN the shipped example).
TMPROOT=$(mktemp -d); mkdir -p "$TMPROOT/example"
printf -- '%s\n' '---' 'name: example' 'description: d' '---' 'x' > "$TMPROOT/example/SKILL.md"
SKPP_SKILLS_DIR="$TMPROOT" ./skpp check | grep -q '^OK.*example (example)' && echo "single-SKILL.md skill is OK (no empty-dir WARN) OK"
rm -rf "$TMPROOT"

# Expected: scope guards confirm no collateral changes; the contract guard confirms
#   check is a stdout-report + exit-code-signal tool (not a §6.4 path emitter).
```

## Final Validation Checklist

### Technical Validation

- [ ] Level 1 passed: `gofmt -l` silent; `go vet ./...` clean; `go build ./...` ok.
- [ ] Level 2 passed: `go test ./... -count=1` green (**195** tests).
- [ ] Level 3 passed: every `… OK` smoke line printed.
- [ ] Level 4 passed: scope guards confirm no collateral changes.
- [ ] `go mod tidy` was a no-op (`go.mod`/`go.sum` byte-identical).
- [ ] No linting errors, no type errors, no formatting drift.

### Feature Validation

- [ ] `skpp check` recognized as a subcommand; `check` NOT captured as a tag.
- [ ] Clean store → OK lines + `0 errors` summary, exit 0.
- [ ] Bad skill → ERROR line(s) + summary with `errors > 0`, exit 1.
- [ ] Malformed YAML and missing block reported as DISTINCT root-cause ERRORs.
- [ ] All five `name` rules enforced + 64-char max (ERROR) / boundary (OK).
- [ ] Missing/empty `description` → ERROR; `> 1024` → WARN; 1024 boundary OK.
- [ ] Duplicate `name` → one ERROR per owner, naming the other tag.
- [ ] WARNs never change the exit code (warn-only store exits 0).
- [ ] Empty store → `0 skills, 0 errors, 0 warnings`, exit 0.
- [ ] Full report to stdout; exit code is the pass/fail signal (not §6.4).
- [ ] Skills dir unresolvable → exit 1, empty stdout, one-line fix on stderr.
- [ ] `--version` precedes `check`; tag resolution still works alongside `check`.

### Code Quality Validation

- [ ] New `internal/check` mirrors `internal/search` (own package, own test, over `[]discover.Skill`).
- [ ] `main.run` stays a thin dispatcher + renderer; check logic is in `internal/check`.
- [ ] `parseArgs` change is additive (one new `case`); all other cases byte-identical.
- [ ] Existing 68 main tests + all read-only packages unchanged and green; the 1
      modified main test (TestParseArgsUnknownTolerated) passes after its `check` →
      `othertag` swap. (12 new main tests + 18 new check tests = 30 added; main 69→81.)
- [ ] No new third-party dependency; stdlib `fmt`/`regexp`/`sort`/`strings` only.
- [ ] Comments explain the *why* (re-parse rationale, the two reframed §9 bullets,
      stdout-vs-exit-code, missing-name-not-a-dup, deferred M5 exclusivity).

### Documentation & Deployment

- [ ] Doc comments on `Check`/`checkFields`/`appendDupFindings` state the PRD §9 rules,
      the re-parse contract, and the two reframed bullets.

---

## Anti-Patterns to Avoid

- ❌ Don't try to detect "directories without SKILL.md but look like skills" via a
  heuristic — it false-positives on legitimate grouping dirs (`skills/writing/`).
  The reframed "unusable SKILL.md" (re-parse `err != nil`) is the actionable analog.
- ❌ Don't implement the "empty besides SKILL.md" WARN — the shipped `example` skill
  IS only `SKILL.md`, so it would break the §13 acceptance ("example as OK").
- ❌ Don't apply §6.4's "nothing on stdout on failure" to `check` — it is a REPORT;
  print findings to stdout and signal pass/fail via the exit code.
- ❌ Don't trust `discover.Index` alone for validation — it drops the per-skill parse
  error. Re-run `discover.ParseFrontmatter(s.SourceFile)` to recover it.
- ❌ Don't double-report a no-block/malformed skill (one root-cause ERROR, not three).
- ❌ Don't let WARNs change the exit code (only ERRORs fail `check`).
- ❌ Don't measure description length on the raw value (folded-scalar trailing newline);
  trim first. Don't skip the `len > 64` name check (the regex can't express the max).
- ❌ Don't add `--help`, exit-2, or `skills/example/` here — those are M5/M6.
- ❌ Don't introduce a new third-party dependency; stdlib only.

---

## Decisions log (assumptions made in lieu of asking — override if you disagree)

| # | Decision | Default chosen | Rationale |
|---|---|---|---|
| 1 | Package home | **`internal/check`** | Mirrors `internal/search` (subcommand name → package name); isolated + self-tested |
| 2 | §9 "no SKILL.md" bullet | **Reframed to "unusable SKILL.md"** (re-parse `err != nil`) | Directory discovery can't surface dirs that never had a SKILL.md without a fragile heuristic; `index.go`'s own comment directs the re-parse |
| 3 | §9 "empty besides SKILL.md" WARN | **Not implemented** (PRD: "optional") | The shipped `example` is only `SKILL.md`; enabling it breaks the §13 acceptance |
| 4 | Output line granularity | **One line per problem**; OK skills one OK line | Preserves all info (e.g. invalid name + over-long desc → 1 ERROR + 1 WARN) |
| 5 | Status column | `%-5s` (`OK`/`WARN`/`ERROR`) | Aligned, readable; matches the `OK   <tag>` shape in PRD §9 |
| 6 | Summary pluralization | Always plural (`N skills, M errors, K warnings`) | Literal PRD §9 form; simpler than singular handling |
| 7 | Stream | Full report to **stdout**; exit code signals pass/fail | check is a report (pipeable); NOT a §6.4 path emitter |
| 8 | Empty store exit code | **0** (clean) | check is validation; no skills == nothing wrong (contrast `--list` exit 1) |
| 9 | description length | **Trimmed** (`strings.TrimSpace`) | Matches `ui.go` display length; folded-scalar trailing newline doesn't count |
| 10 | name length | `len(name)` (ASCII == runes) | Agent Skills names are lowercase a-z0-9-; byte length is correct |
| 11 | dup-name scan | Global, 2nd pass, case-sensitive on **non-empty** names | A missing name is its own ERROR, not a dup; each owner gets an ERROR naming the others |
| 12 | `check` token | **Reserved** subcommand; not a tag | PRD §6.1 lists `check` as a command; CLIs reserve subcommand names |
| 13 | M5 deferrals | `--help`, §6.3 exit-2 exclusivity | Owned by P1.M5.T11; until then `check` wins silently in dispatch (mirrors `searchMode`) |

---

**Confidence score: 9/10** for one-pass implementation success. The two non-obvious
decisions (reframing the "no SKILL.md" bullet; stdout-report-vs-exit-code) are
documented with rationale and verified against the codebase's own `index.go` comment
and against §13 acceptance. The full file contents, exact edit snippets, and named
tests leave no ambiguity. The one residual risk is the reframed-bullet judgment call
(a human might prefer the literal "dir without SKILL.md" scan) — it is called out in
the Scope section, the Decisions log, and the Anti-Patterns so it can be overridden
deliberately rather than silently.
