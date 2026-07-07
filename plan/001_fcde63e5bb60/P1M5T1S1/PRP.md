# PRP — P1.M5.T1.S1: `--help` text + no-args/unknown-flag/mode-exclusivity exit codes

> **Subtask:** P1.M5.T1.S1 (plan graph: P1.M5.T11.S1) — the CLI-surface *gate*
> (build-order step 5). It finalizes `main.go` arg parsing + dispatch so the
> binary matches PRD §6.1–§6.4 exactly: a full `--help`/`-h` (mirrors the
> mcpeepants USAGE/EXAMPLES/OPTIONS structure), the §6.3 no-args →
> usage-to-stderr + exit 1, the §6-header unknown-flag → `skpp: unknown flag
> '<x>'` + exit 2, the §6.3 mode-mutual-exclusivity → exit 2, and the `check`
> subcommand dispatch (positional `check` as first arg, mutually exclusive with
> tags). Every exit code (0/1/2) and every stdout/stderr discipline is now
> pinned. This is the last functional gate before packaging/docs (P1.M6).
>
> **Scope:** MODIFY two files — `main.go` (config + parseArgs + run reorder +
> `runCheck` + `usage`/`usageText` + `exclusivityError`) and `main_test.go`
> (update 3 existing tests; append ~18 new). **No new files, no new packages.**
> Consumes `internal/check` if/when P1.M4.T10.S1 has landed (see Dependencies).
>
> **DEPENDENCIES (CONTRACT):**
> - **P1.M4.T9.S1 (`--search`) — LANDED & GREEN** (parallel sibling; treat its
>   PRP as the contract). By the time this runs, `config` has `searchMode bool` +
>   `searchQ string`, `parseArgs` is an **index loop** (`for i := 0; …`) with a
>   `case "--search","-s":` that consumes `args[i+1]` and `i++`, and `run()` has
>   a `if c.searchMode {…}` branch after `--list` and before `--all`. main_test.go
>   = **69 tests** (53 + 16 search). This PRP is written against that 69-test
>   baseline; confirm it on disk (Task 0).
> - **P1.M4.T10.S1 (`check` validation, §9) — REQUIRED prerequisite** for the
>   *dispatch* of `check`, but NOT for the rest of this task. It owns the §9
>   rules + OK/WARN/ERROR format + exit code. This task owns only the DISPATCH
>   (parseArgs recognition, exclusivity, `run()` routing). If T10.S1 has NOT
>   landed, implement everything EXCEPT the one delegate call inside `runCheck`
>   (see Task 4 / Contingency) and flag it.
>
> **PARALLEL CONTEXT:** This PRP is authored while P1.M4.T9.S1 is in flight. The
> on-disk `main.go` (read in full) is post-M3.T8.S2 (range-loop parseArgs, no
> `--search`). The 69-test post-T9 baseline is assumed; the only T9 surface this
> task depends on is the index-loop `parseArgs` and the `searchMode`/`searchQ`
> fields — both stable per the T9 PRP contract.

---

## Goal

**Feature Goal**: Ship the complete `skpp` CLI surface (PRD §6.1–§6.4). `skpp
--help`/`-h` prints a structured USAGE/EXAMPLES/OPTIONS block to **stdout** and
exits 0, taking precedence over `--version` (help wins the tie, per the item).
`skpp` with no args prints usage to **stderr** and exits 1 (parity with
mcpeepants `get-server-config.sh`). Any unknown dashed flag
(`--bogus`/`-x`) prints `skpp: unknown flag '<x>'` to stderr and exits 2. Mixing
positional `<tag>`s with `--list`/`--search`/`--all`, or `check` with tags/other
modes, prints an error to stderr and exits 2. `check` is recognized as a
subcommand (first token) and dispatched to the validation engine (T10.S1).

**Deliverable**: Two MODIFIED files (no new files/packages):
1. `main.go` — `config` gains `help bool`, `check bool`, `unknownFlag string`;
   `parseArgs` gains `--help`/`-h`, `check`-as-first-arg detection, and
   unknown-flag capture; `run()` is reordered (help→version→unknown→exclusivity→
   dispatch→no-args-usage); new `usageText`/`usage()`, `exclusivityError()`, and
   `runCheck()` helpers. ~+90 lines.
2. `main_test.go` — UPDATE 3 existing tests (`TestRunDefaultUnknownFlag`,
   `TestParseArgsUnknownTolerated`, `TestRunDefaultNoArgs`) and APPEND ~18 new
   tests (help short/long, help-beats-version, no-args-stderr, unknown-flag x2,
   version/help-beats-unknown precedence, exclusivity x4, check dispatch +
   unresolvable).

**Success Definition**: `gofmt -l main.go main_test.go` silent; `go vet ./...`
clean; `go build ./...` exit 0; `go test ./...` green (**main grows 69 → ~87**;
whole module ~**195**). Empirically: `./skpp --help`→stdout usage+exit 0;
`./skpp`→stderr usage+exit 1; `./skpp --bogus`→stderr `skpp: unknown flag
'--bogus'`+exit 2; `./skpp foo --list`→stderr+exit 2; `./skpp check`→dispatches
to check (exit 0 on a clean store). `go mod tidy` is a no-op (stdlib only). No
touch to `internal/{discover,resolve,search,skillsdir,ui,check}/*`,
`go.mod`/`go.sum`, `PRD.md`, `tasks.json`.

## User Persona

**Target User**: A pi operator who drives skills via
`pi --skill "$(skpp <tag>)"` and occasionally introspects the CLI by hand.

**Use Case**: "I forgot the exact flag for searching" → `skpp --help` shows the
full matrix in one screen. "I mistyped `--serch`" → `skpp: unknown flag '--serch'`
(exit 2) tells me immediately, instead of silent wrong behavior.

**Pain Points Addressed**: silent wrong answers (unknown flag tolerated as no-op
→ exit 1 with no message); no discoverable help; ambiguous combos (`foo --list`)
that silently did something surprising.

## Why

- **Closes the §6 contract.** §6.3 (no-args, precedence) and the §6 header
  (unknown ⇒ exit 2) are the only §6 behaviors still "tolerated" (silent exit 1)
  after M1–M4. This task flips them to their final, tested shape — the gate
  before packaging/docs (M6) and the README (which documents these exit codes).
- **Makes `$(...)` and scripts safe and debuggable.** Unknown flags now fail
  loudly (exit 2) instead of silently exiting 1; help is one flag away. This is
  the UX parity the PRD demands with `get-server-config.sh`.
- **Locks the precedence + exclusivity rules in code + tests** before the
  acceptance sweep (P1.M6.T16.S1) and the shell completions (P1.M6.T15.S1, which
  list these exact flags) lean on them.
- **Wires the `check` dispatch** so T10.S1's validation engine has a real CLI
  entry point the moment it lands, and §9/§13's `./skpp check` acceptance gate is
  reachable.

## What

User-visible behavior (PRD §6.1–§6.4; item points a–e):

- **(a) `--help`/`-h`** → print the full usage block (USAGE / EXAMPLES / OPTIONS
  + a one-line exit-code reference) to **stdout**, exit 0. Takes precedence over
  EVERYTHING, including `--version` (help wins the tie) and unknown flags.
- **(b) No args AND no mode** → print the SAME usage block to **stderr**, exit 1.
- **(c) Unknown dashed flag** (any token starting with `-` that is not in the
  known set: `--help/-h --version/-v --path/-p --list/-l --search/-s --all/-a
  --file/-f --relative --no-color`) → print `skpp: unknown flag '<x>'` to stderr
  (the FIRST such token wins), exit 2. Positional tags are unaffected.
- **(d) Mode mutual exclusivity** → `len(tags)>0 && (--list|--search|--all)` ⇒
  stderr error, exit 2.
- **(e) `check` subcommand** → `check` as the FIRST token is the subcommand
  (sets `c.check`); mutually exclusive with tags AND with `--list`/`--search`/
  `--all` (any combo ⇒ exit 2). `run()` routes it to `runCheck` (T10.S1's
  engine). `check` in any non-first position is an ordinary tag.

### Success Criteria

- [ ] `skpp --help` and `skpp -h` print USAGE/EXAMPLES/OPTIONS to **stdout**, exit 0, stderr empty.
- [ ] The help block contains the canonical `pi --skill "$(skpp <tag>)"` example.
- [ ] `--help` beats `--version` (`skpp --help --version` → help, exit 0; NOT the version line).
- [ ] `--version` still beats unknown flag + every mode (`skpp --version --bogus` → version, exit 0).
- [ ] `--help` beats unknown flag (`skpp --help --bogus` → help, exit 0).
- [ ] `skpp` (no args) → usage to **stderr**, exit 1, stdout empty.
- [ ] `skpp --bogus` → `skpp: unknown flag '--bogus'` on stderr, stdout empty, exit 2.
- [ ] `skpp -x` (short unknown) → same, exit 2.
- [ ] First unknown wins: `skpp --bogus --more` → message names `--bogus`.
- [ ] `skpp foo --list` / `foo --search q` / `foo --all` → stderr error, stdout empty, exit 2.
- [ ] `skpp check foo` (check + tag) → stderr error, exit 2.
- [ ] `skpp check --list` (check + mode) → stderr error, exit 2.
- [ ] `skpp foo check` → `check` is a TAG (not subcommand); resolves/errors normally.
- [ ] `skpp check` → dispatches to the validation engine; exit 0 on a clean store.
- [ ] Existing behavior preserved: version/path/list/search/all/tags precedence + §6.4 atomicity unchanged.
- [ ] `go test ./...` green; `gofmt`/`go vet` clean; `go.mod` unchanged.

## All Needed Context

### Context Completeness Check

_If someone knew nothing about this codebase, would they have everything needed
to implement this successfully?_ **Yes.** The exact `config`/`parseArgs`/`run`
edits are given verbatim below (full new `parseArgs`, full new `run` skeleton,
full `usageText`, full `exclusivityError`). The 3 test updates and ~18 new test
functions are named and specified. The one unknown — T10.S1's `internal/check`
signature — is isolated to a single delegate call inside `runCheck`, with a
precondition task that reads the real signature off disk. Every load-bearing
decision (precedence order, check-as-first-arg, exclusivity families, plain-text
help, exit-code mapping) is pinned in `research/verified_facts.md`.

### Documentation & References

```yaml
# MUST READ — this subtask's own empirical verification (every load-bearing decision)
- file: plan/001_fcde63e5bb60/P1M5T1S1/research/verified_facts.md
  why: "§1 the 4-behavior + check-dispatch contract table; §2 the mcpeepants help
        STRUCTURE to mirror (USAGE/EXAMPLES/OPTIONS, aligned columns) + the plain-text
        color decision; §3 the current config/parseArgs/run; §4 the check↔T10.S1
        division of labor + the integration seam; §5 the 3 existing tests that MUST
        change (exit-1→exit-2 etc.); §6 the check-as-first-arg rule; §7 the exact 3
        exclusivity families; §8 verified validation commands; §9 scope boundary."
  critical: "Precedence is help → version → unknown-flag → exclusivity → dispatch →
             no-args. NOT version-first (current). Inserting help above version is the
             one reorder that could surprise; the existing version-precedence tests
             use --version alone (no --help) so they stay green."

# MUST READ — the authoritative §6 spec (the whole point of this task)
- file: PRD.md
  section: "§6 header (unknown ⇒ exit 2), §6.1 (--help/--version/--path rows),
            §6.3 (no-args stderr+exit1; --help/--version precedence; mutual-exclusivity
            exit 2), §6.4 (stdout-empty-on-failure discipline)."
  why: "These four sub-sections ARE the contract this task implements. §6.3 line 3
        ('--help / --version take precedence over everything else') + the item's
        'help wins' tiebreak ⇒ help checked before version. READ-ONLY."
  critical: "§6.4 applies to the NEW exit-2 paths too: NOTHING on stdout when we
             exit 2 (unknown flag / exclusivity). Keep stdout empty on every
             non-success path so pi --skill \"$(skpp …)\" never sees garbage."

# MUST READ — the parallel sibling whose output we build on (the post-T9 main.go)
- file: plan/001_fcde63e5bb60/P1M4T9S1/PRP.md
  why: "Defines the 69-test baseline this task edits: config has searchMode/searchQ;
        parseArgs is an index loop with case \"--search\",\"-s\"; run has a search
        branch between --list and --all. This task's parseArgs/run edits are layered
        ON TOP of that shape."
  pattern: "parseArgs post-T9: `for i := 0; i < len(args); i++ { a := args[i]; switch a {…} }`"
  gotcha: "Do NOT revert the index loop or touch the --search case. This task only
           ADDS: --help case, check-first-arg pre-check, unknown-flag capture in the
           default branch, and the run() reorder + new helpers."

# CONTRACT — the validation engine this task's check branch delegates to (T10.S1)
- file: internal/check/   # (created by P1.M4.T10.S1 — may not exist yet at impl time)
  why: "runCheck calls T10's entry point to produce the §9 report + exit code. This
        task owns the dispatch (Find+Index preamble + routing), NOT the §9 rules."
  pattern: "Likely (mirrors resolve/search as a pure fn over the index):
            package check
            func Report(w io.Writer, skills []discover.Skill) int   // 0 clean, 1 if any ERROR"
  gotcha: "T10's exact signature is UNKNOWN at PRP time. Task 0 precondition: read
           internal/check/*.go on disk and adapt the single delegate call. If
           internal/check does NOT exist, T10.S1 has not landed — implement the rest
           and flag runCheck's delegate as the one unfinished seam (Contingency)."

# MUST READ — the consumed packages (the data flow runCheck/main rely on)
- file: internal/skillsdir/skillsdir.go
  why: "skillsdir.Find() (dir, src, err) — the §8 resolver. runCheck (and every mode)
        calls it first; on err, print err verbatim to stderr + exit 1 (§6.4)."
- file: internal/discover/index.go
  why: "discover.Index(dir) ([]Skill, error) — the pre-sorted catalog. runCheck
        passes it to T10's engine."

# REFERENCE — the help-text STRUCTURE to mirror (read in full; see research §2)
- file: ~/projects/mcpeepants/get-server-config.sh   # the --help branch
  why: "USAGE / EXAMPLES / OPTIONS sections, aligned OPTIONS columns, a canonical
        one-liner example. skpp mirrors the STRUCTURE (not the bash/color) and swaps
        the canonical example to pi --skill \"$(skpp <tag>)\"."
  pattern: "title line → USAGE: → EXAMPLES: (canonical one-liner + a few) → OPTIONS:
            (two aligned columns, flag … one-liner)."
  gotcha: "mcpeepants colorizes help with ANSI unconditionally; skpp emits help PLAIN
           (no ANSI) unconditionally — see research §2 for the 4 reasons. Do NOT add
           color to the help block."

# URLS — the stdlib mechanisms used (stable since Go 1.0; no version concern)
- url: https://pkg.go.dev/fmt#Fprint
  why: "fmt.Fprint(stdout, usageText) prints the help block verbatim (no extra
        newline — usageText already ends in \\n). fmt.Fprintf(stderr, \"skpp: unknown
        flag '%s'\\n\", flag) for the exit-2 message."
- url: https://pkg.go.dev/strings#HasPrefix
  why: "strings.HasPrefix(a, \"-\") distinguishes dashed tokens (unknown-flag
        candidates) from positionals (tags) in the parseArgs default branch."
```

### Current Codebase tree (relevant slice; post-T9 assumed)

```bash
skpp/
├── go.mod                      # module github.com/dabstractor/skpp; go 1.25; yaml.v3 (UNCHANGED)
├── main.go                     # MODIFY: config + parseArgs + run + new helpers + usageText
├── main_test.go                # MODIFY: update 3 tests + append ~18 (existing 69 unchanged otherwise)
└── internal/
    ├── discover/{skill,index,discover,*_test.go}  # READ-ONLY (Index, Skill)
    ├── resolve/{resolve,*_test.go}                # READ-ONLY
    ├── search/{search,*_test.go}                  # READ-ONLY (P1.M4.T9.S1)
    ├── skillsdir/{skillsdir,*_test.go}            # READ-ONLY (Find)
    ├── ui/{ui,*_test.go}                          # READ-ONLY
    └── check/{*.go}                               # T10.S1-owned; CONSUMED by runCheck only
```

### Desired Codebase tree (files added/modified)

```bash
main.go          # MODIFY — config(+3 fields) / parseArgs(+help, +check-first, +unknown capture)
                           run(reorder + runCheck/usage/exclusivityError) / usageText const
main_test.go     # MODIFY — 3 updated + ~18 appended (no helper changes; reuse sampleStore etc.)
```
No new files. No new packages. No new dependencies.

### Known Gotchas of our codebase & Library Quirks

```go
// CRITICAL — PRECEDENCE REORDER. Current run() checks version FIRST. This task
// inserts help ABOVE version (item: "both → help wins"). Every existing
// version-precedence test uses --version WITHOUT --help, so they stay green —
// but double-check TestRunVersionPrecedenceOver{Path,Tag,All,Search} after the
// edit. Order in run(): help → version → unknownFlag → exclusivity → dispatch
// (check→path→list→search→all→tags) → no-args-usage.

// CRITICAL — §6.4 stdout discipline applies to the NEW exit-2 paths too. On
// unknown-flag and exclusivity, print ONLY to stderr; stdout must be EMPTY.
// (Tests assert out.Len()==0.) This keeps `pi --skill "$(skpp --bogus)"` from
// passing a garbage path — the whole point of §6.4.

// CRITICAL — `check` is a subcommand iff args[0] == "check". Detect it BEFORE
// the token loop, set c.check, and start the loop at index 1 (so the `check`
// token is not also captured as a tag). A `case "check":` is NOT needed: `check`
// in any non-first slot falls through `default` → not dashed → captured as a tag.
// `skpp foo check` therefore treats `check` as a tag, NOT the subcommand.

// CRITICAL — unknown-flag capture stores the FIRST offender only
// (`if c.unknownFlag == "" { c.unknownFlag = a }`). run() reports that one. Do
// NOT collect a slice; one loud error is the §6 contract.

// GOTCHA — `--search <q>` (T9) consumes args[i+1] verbatim, INCLUDING a value
// that starts with '-' (e.g. `--search -x` → query "-x"). So a dashed token
// grabbed as a search value is NOT an unknown flag. Do NOT re-scan consumed
// search values. Leave the T9 case byte-identical.

// GOTCHA — help text is PLAIN (no ANSI), unconditionally. Not gated on
// isTerminal/--no-color. `skpp --help | grep` must work; tests use non-TTY
// buffers. (See research §2 for the full rationale.)

// GOTCHA — the same usageText is printed to stdout (--help) AND stderr (no-args).
// Use fmt.Fprint (NOT Fprintln) so there is no double trailing newline; the
// constant already ends in exactly one '\n'.

// GOTCHA — `runCheck`'s delegate call to internal/check is the ONE line whose
// exact form depends on T10.S1's signature (unknown at PRP time). Read
// internal/check/*.go on disk (Task 0) and adapt. Do NOT invent a stub that
// would later conflict with T10's real package.

// GOTCHA — gofmt will realign the config struct comments when fields are added;
// run `gofmt -w main.go` after editing (expected, correct).
```

## Implementation Blueprint

### Data models — `config` additions

```go
// config holds the parsed CLI flags. Grown milestone by milestone. THIS subtask
// (P1.M5.T1.S1) adds the final three: help, check, unknownFlag — completing the
// §6.1–§6.4 matrix. (searchMode/searchQ arrived in P1.M4.T9.S1.)
type config struct {
	version    bool     // --version / -v
	help       bool     // --help / -h       : print usage to STDOUT, exit 0 [NEW]
	path       bool     // --path / -p
	list       bool     // --list / -l
	all        bool     // --all / -a
	file       bool     // --file / -f
	relative   bool     // --relative
	noColor    bool     // --no-color
	searchMode bool     // --search <q> / -s   (P1.M4.T9.S1)
	searchQ    string   // the --search value  (P1.M4.T9.S1)
	check      bool     // `check` subcommand (first arg) [NEW]
	tags       []string // positional <tag> args
	unknownFlag string  // first unknown dashed token, "" if none [NEW]
}
```

### File 1 — MODIFY `main.go`

Apply these edits to the post-T9 `main.go`. Run `gofmt -w main.go` after.

**Edit 1a — `usageText` constant + `usage()` (add near the top, after the
`version` var).** Plain text, mirrors mcpeepants USAGE/EXAMPLES/OPTIONS. The
`%[1]s` is NOT used (binary name is fixed as `skpp`); keep it literal.

```go
// usageText is the full --help / no-args usage block (PRD §6.1, §6.3). It mirrors
// the STRUCTURE of mcpeepants get-server-config.sh (USAGE / EXAMPLES / OPTIONS,
// aligned columns) but lists the full skpp §6 flag matrix and the canonical
// pi --skill "$(skpp <tag>)" one-liner. It is emitted PLAIN (no ANSI):
// `skpp --help | grep` must work, §13 does not assert on help color, and tests
// use non-TTY buffers. The SAME text is printed to stdout for --help (exit 0) and
// to stderr for the no-args default (exit 1) — only the destination differs.
const usageText = `skpp — skill path printer

Resolve skill tags to on-disk skill directory paths (manifest-free).

USAGE:
  skpp <tag> [<tag>...]
  skpp --all
  skpp --list
  skpp --search <query>
  skpp check
  skpp --path
  skpp --help
  skpp --version

EXAMPLES:
  pi --skill "$(skpp example)"
  pi --skill "$(skpp writing/reddit)"
  skpp example reddit          # one absolute path per line, input order
  skpp -f example              # print the SKILL.md path
  skpp --relative --all        # every skill path, relative to the skills dir
  skpp --list                  # human-readable catalog
  skpp --search reddit         # substring search over tag/name/description/keywords
  skpp check                   # validate every skill on disk

OPTIONS:
  <tag> [<tag>...]   Resolve tags to skill directory paths (one absolute path per line)
  --all, -a          Print every skill's directory path, sorted by tag
  --list, -l         Human-readable catalog (TAG, NAME, DESCRIPTION)
  --search <q>, -s   Substring search over tag / name / description / keywords
  check              Validate every skill on disk (report OK / WARN / ERROR)
  --path, -p         Print the resolved skills directory
  --file, -f         Print the SKILL.md path instead of the directory (modifier)
  --relative         Print paths relative to the skills directory (modifier)
  --no-color         Disable ANSI color even on a TTY (modifier)
  --help, -h         Show this help message
  --version, -v      Print the skpp version

Exit codes: 0 success | 1 unresolved/no skills/unresolvable dir | 2 unknown flag / mutually-exclusive modes
`

// usage returns the help block. A tiny indirection so tests can assert on the
// constant without importing it, and so future localization could swap it.
func usage() string { return usageText }
```

**Edit 1b — `config` struct** (replace the post-T9 struct with the version in
"Data models" above; gofmt realigns the comments). Drop the `// Future …`
comment entirely — the matrix is now complete.

**Edit 1c — `parseArgs`** (layer on top of the post-T9 index loop; add the
`--help` case, the `check`-as-first-arg pre-check, and unknown-flag capture in
the default branch; leave the `--search` case byte-identical):

```go
func parseArgs(args []string) config {
	var c config
	// `check` subcommand: recognized iff it is the FIRST token (PRD §6.1; item
	// P1.M5.T1.S1.e). Start the token loop at index 1 so the `check` token is not
	// also captured as a tag. Any positional AFTER a leading `check` is a tag,
	// which exclusivity (run) flags as exit 2 (check mutually exclusive with tags).
	start := 0
	if len(args) > 0 && args[0] == "check" {
		c.check = true
		start = 1
	}
	for i := start; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--version", "-v":
			c.version = true
		case "--help", "-h": // [NEW] precedence over everything except itself
			c.help = true
		case "--path", "-p":
			c.path = true
		case "--list", "-l":
			c.list = true
		case "--all", "-a":
			c.all = true
		case "--file", "-f":
			c.file = true
		case "--relative":
			c.relative = true
		case "--no-color":
			c.noColor = true
		case "--search", "-s":
			// Value-taking flag (P1.M4.T9.S1): consume the NEXT token verbatim as
			// the query (even if it starts with '-'); i++ skips it. If --search is
			// the LAST token, searchMode stays false → falls to the no-mode default.
			if i+1 < len(args) {
				c.searchMode = true
				c.searchQ = args[i+1]
				i++
			}
		default:
			// [NEW] unknown dashed flag → capture the FIRST offender for run() to
			// report as exit 2 (PRD §6 header). Non-dashed tokens are positional
			// <tag>s (captured in input order). Do NOT collect a slice of unknowns;
			// one loud error is the §6 contract.
			if strings.HasPrefix(a, "-") {
				if c.unknownFlag == "" {
					c.unknownFlag = a
				}
			} else {
				c.tags = append(c.tags, a)
			}
		}
	}
	return c
}
```

**Edit 1d — `run()` reorder** (replace the post-T9 `run` precedence preamble;
keep every mode branch body byte-identical; change the silent default to
usage→stderr+exit1; insert the unknown-flag + exclusivity checks; add the
`check` dispatch). The full new `run` skeleton (mode bodies elided with `…` —
they are UNCHANGED from the post-T9 main.go; only their ordering context + the
new preamble/coda change):

```go
func run(args []string, stdout, stderr io.Writer) int {
	c := parseArgs(args)

	// 1) --help takes precedence over EVERYTHING, including --version (item:
	//    "both → help wins") and unknown flags (PRD §6.3). Usage to STDOUT, exit 0.
	if c.help {
		fmt.Fprint(stdout, usage())
		return 0
	}
	// 2) --version next (PRD §6.3: --version takes precedence over everything
	//    except --help). Keeps Find() uncalled and a broken skills dir invisible.
	if c.version {
		fmt.Fprintf(stdout, "skpp %s\n", version)
		return 0
	}
	// 3) Unknown dashed flag → exit 2 (PRD §6 header; item c). stdout stays EMPTY
	//    (§6.4 discipline: pi --skill "$(skpp --bogus)" must fail loudly).
	if c.unknownFlag != "" {
		fmt.Fprintf(stderr, "skpp: unknown flag '%s'\n", c.unknownFlag)
		return 2
	}
	// 4) Mode mutual exclusivity → exit 2 (PRD §6.3; item d/e). Checked AFTER
	//    unknown-flag so a combo like `--bogus foo --list` reports the unknown
	//    flag first (both are exit 2; unknown is the more fundamental error).
	if bad, msg := exclusivityError(c); bad {
		fmt.Fprintln(stderr, msg)
		return 2
	}

	// 5) Normal mode dispatch (order unchanged from post-T9, with check added
	//    first — it is guaranteed standalone here because exclusivity caught any
	//    check+mode/tags combo above).
	if c.check {
		return runCheck(stdout, stderr)
	}
	if c.path {
		// … UNCHANGED post-T9 body: Find → Fprintln(stdout, dir) / err→stderr+exit1 …
	}
	if c.list {
		// … UNCHANGED post-T9 body …
	}
	if c.searchMode {
		// … UNCHANGED post-T9 body …
	}
	if c.all {
		// … UNCHANGED post-T9 body …
	}
	if len(c.tags) > 0 {
		// … UNCHANGED post-T9 body (§6.4 atomic resolve loop) …
	}

	// 6) No recognized mode → usage to STDERR, exit 1 (PRD §6.3; item b). Parity
	//    with get-server-config.sh. Covers both truly-no-args and modifiers-only
	//    (e.g. `skpp --no-color`): if skpp was asked to DO nothing, show usage.
	fmt.Fprint(stderr, usage())
	return 1
}
```

**Edit 1e — `exclusivityError` helper** (new; the exactly-three specified
families from research §7):

```go
// exclusivityError reports whether c combines modes that PRD §6.3 / the item
// forbid, returning a one-line stderr message. It implements EXACTLY three
// families (research §7): tags + a listing mode; check + tags; check + a listing
// mode. Unspecified combos (e.g. --list --search, no tags) are left to the
// deterministic dispatch order, not flagged. --file/--relative/--no-color are
// MODIFIERS and never trigger exclusivity.
func exclusivityError(c config) (bad bool, msg string) {
	hasTags := len(c.tags) > 0
	if hasTags && (c.list || c.searchMode || c.all) {
		return true, "skpp: tags cannot be combined with --list/--search/--all"
	}
	if c.check && hasTags {
		return true, "skpp: 'check' cannot be combined with tag arguments"
	}
	if c.check && (c.list || c.searchMode || c.all) {
		return true, "skpp: 'check' cannot be combined with --list/--search/--all"
	}
	return false, ""
}
```

**Edit 1f — `runCheck` helper** (new; the check DISPATCH. The Find+Index
preamble is fully specified — identical to every other mode. The ONE delegate
call to T10.S1's engine is marked ADAPT):

```go
// runCheck dispatches the `check` subcommand (PRD §6.1, §9). It owns the same
// skillsdir.Find() → discover.Index() preamble every mode uses (§8 + §7.1), then
// delegates the §9 validation report + exit code to P1.M4.T10.S1's engine. This
// task owns the DISPATCH only; the OK/WARN/ERROR rules, output format, and the
// per-finding exit-code mapping live in T10.S1's package (very likely
// internal/check). Exit 1 on Find/Index failure (§6.4: one-line fix to stderr,
// stdout empty); otherwise whatever T10's engine returns (0 clean, 1 if any ERROR).
func runCheck(stdout, stderr io.Writer) int {
	dir, _, err := skillsdir.Find()
	if err != nil {
		fmt.Fprintln(stderr, err) // verbatim one-line fix (§6.4/§8); stdout stays empty
		return 1
	}
	skills, err := discover.Index(dir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	// >>> ADAPT TO P1.M4.T10.S1'S ACTUAL SIGNATURE (read internal/check/*.go). <<<
	// Assumed shape (mirrors resolve/search as a pure fn over the index):
	//   return check.Report(stdout, skills)   // writes the §9 report, returns 0/1
	// If T10 instead returns findings, map severity here:
	//   findings := check.Check(skills)
	//   // print findings + summary, then return 0 if no ERROR else 1 (§9).
	return t10CheckDelegate(stdout, skills) // see Task 4 — replace with the real call
}

// t10CheckDelegate is a PLACEHOLDER so main.go compiles even if internal/check is
// absent (T10.S1 not yet landed). REPLACE its body with the real delegate call
// once internal/check exists (Task 4). It currently reports "not implemented" so
// `skpp check` fails loudly instead of pretending success.
func t10CheckDelegate(stdout io.Writer, skills []discover.Skill) int {
	fmt.Fprintln(stdout, "skpp: check is not yet implemented (P1.M4.T10.S1 pending)")
	return 1
}
```

> **The placeholder is intentional and safe:** if T10.S1 has NOT landed, `skpp
> check` exits 1 with a clear message (never silently succeeds), and the
> `TestRunCheckDispatchesToCheck` test is SKIPPED/gated (see File 2). Once T10
> lands, Task 4 replaces `t10CheckDelegate` with the real `check.Report(...)`
> call and un-gates the test. This keeps main.go compiling + green in BOTH
> worlds and avoids inventing a §9 engine that would conflict with T10.

### File 2 — MODIFY `main_test.go`

**Update these 3 existing tests** (their names/locations are unchanged; only
assertions + the outdated "tolerated" framing change):

```go
// === TestRunDefaultNoArgs — UPDATE: now usage to STDERR + exit 1 (was silent) ===
func TestRunDefaultNoArgs(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(nil, &out, &errOut)
	if code != 1 {
		t.Errorf("run(nil): code=%d; want 1 (no-args → stderr usage, exit 1)", code)
	}
	if out.Len() != 0 {
		t.Errorf("run(nil) stdout=%q; want EMPTY (usage goes to stderr, not stdout)", out.String())
	}
	if !strings.Contains(errOut.String(), "USAGE") {
		t.Errorf("run(nil) stderr=%q; want the USAGE block (no-args prints usage to stderr)", errOut.String())
	}
}

// === TestRunDefaultUnknownFlag — UPDATE: unknown flag now exits 2 (was 1) ===
func TestRunDefaultUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--frobnicate"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(--frobnicate): code=%d; want 2 (unknown flag, PRD §6)", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want EMPTY (§6.4: nothing on stdout on exit-2)", out.String())
	}
	want := "skpp: unknown flag '--frobnicate'\n"
	if got := errOut.String(); got != want {
		t.Errorf("stderr=%q; want %q", got, want)
	}
}

// === TestParseArgsUnknownTolerated → RENAME TestParseArgsUnknownFlagCaptured ===
// (and update: unknown is now CAPTURED for exit-2, not tolerated as a no-op)
func TestParseArgsUnknownFlagCaptured(t *testing.T) {
	c := parseArgs([]string{"--frobnicate", "sometag", "check"})
	if c.version || c.path {
		t.Errorf("version=%v path=%v; want both false", c.version, c.path)
	}
	if c.unknownFlag != "--frobnicate" {
		t.Errorf("unknownFlag=%q; want --frobnicate (first unknown captured)", c.unknownFlag)
	}
	// `check` here is the 2nd positional (sometag is first) → an ordinary TAG,
	// NOT the subcommand. So both positionals are captured as tags in order.
	if len(c.tags) != 2 || c.tags[0] != "sometag" || c.tags[1] != "check" {
		t.Errorf("tags=%v; want [sometag check] (positionals captured; check-not-first is a tag)", c.tags)
	}
	if c.check {
		t.Errorf("c.check=true; want false (check was not the first arg)")
	}
}
```

**Append these ~18 new tests** (reuse the existing `sampleStore`/`writeSkillTree`/
`withTerminal`/`unsetSkillsEnv` helpers; existing 66 other tests unchanged):

```go
// --- parseArgs: --help/-h, check subcommand, unknown capture (P1.M5.T1.S1) ---

func TestParseArgsHelpLong(t *testing.T) {
	c := parseArgs([]string{"--help"})
	if !c.help {
		t.Errorf("parseArgs(--help): help=false; want true")
	}
}

func TestParseArgsHelpShort(t *testing.T) {
	c := parseArgs([]string{"-h"})
	if !c.help {
		t.Errorf("parseArgs(-h): help=false; want true")
	}
}

// `check` as the FIRST arg → subcommand; c.check true; the token is NOT a tag.
func TestParseArgsCheckSubcommand(t *testing.T) {
	c := parseArgs([]string{"check"})
	if !c.check {
		t.Fatalf("parseArgs(check): check=false; want true")
	}
	if len(c.tags) != 0 {
		t.Errorf("tags=%v; want empty (check token not captured as a tag)", c.tags)
	}
}

// `check` as a NON-first positional → an ordinary TAG, not the subcommand.
func TestParseArgsCheckNotFirstIsTag(t *testing.T) {
	c := parseArgs([]string{"foo", "check"})
	if c.check {
		t.Errorf("c.check=true; want false (check was not first)")
	}
	if len(c.tags) != 2 || c.tags[0] != "foo" || c.tags[1] != "check" {
		t.Errorf("tags=%v; want [foo check]", c.tags)
	}
}

// Only the FIRST unknown dashed token is captured.
func TestParseArgsFirstUnknownWins(t *testing.T) {
	c := parseArgs([]string{"--bogus", "--more"})
	if c.unknownFlag != "--bogus" {
		t.Errorf("unknownFlag=%q; want --bogus (first unknown wins)", c.unknownFlag)
	}
}

// A short unknown flag is captured too.
func TestParseArgsShortUnknownCaptured(t *testing.T) {
	c := parseArgs([]string{"-x"})
	if c.unknownFlag != "-x" {
		t.Errorf("unknownFlag=%q; want -x", c.unknownFlag)
	}
}

// --- run: --help / -h (P1.M5.T1.S1) ---

func TestRunHelpToStdoutExit0(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--help"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(--help): code=%d; want 0", code)
	}
	got := out.String()
	for _, want := range []string{"USAGE:", "EXAMPLES:", "OPTIONS:", `pi --skill "$(skpp example)"`} {
		if !strings.Contains(got, want) {
			t.Errorf("run(--help) stdout missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("help must be PLAIN (no ANSI):\n%s", got)
	}
	if errOut.Len() != 0 {
		t.Errorf("run(--help) stderr=%q; want empty", errOut.String())
	}
}

func TestRunHelpShortFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-h"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(-h): code=%d; want 0", code)
	}
	if !strings.Contains(out.String(), "USAGE:") {
		t.Errorf("run(-h) stdout missing USAGE block:\n%s", out.String())
	}
}

// help beats version (item: "both → help wins").
func TestRunHelpBeatsVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--help", "--version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(--help --version): code=%d; want 0", code)
	}
	if strings.Contains(out.String(), "skpp "+version) {
		t.Errorf("help must beat version; got the version line:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "USAGE:") {
		t.Errorf("stdout should be the help block, not the version:\n%s", out.String())
	}
}

// --- run: no-args (P1.M5.T1.S1) ---

// (TestRunDefaultNoArgs updated above covers run(nil).)

// Modifiers-only (a flag, but no mode) → usage to stderr, exit 1 (treated as
// "asked to do nothing"). stdout empty.
func TestRunModifiersOnlyNoMode(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--no-color"}, &out, &errOut)
	if code != 1 {
		t.Errorf("run(--no-color): code=%d; want 1 (no mode → stderr usage)", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "USAGE") {
		t.Errorf("stderr=%q; want usage block", errOut.String())
	}
}

// --- run: unknown flag → exit 2 (P1.M5.T1.S1) ---

// (TestRunDefaultUnknownFlag updated above covers run(--frobnicate).)

func TestRunUnknownShortFlagExit2(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-z"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(-z): code=%d; want 2", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
	if got := errOut.String(); got != "skpp: unknown flag '-z'\n" {
		t.Errorf("stderr=%q; want the exact unknown-flag line", got)
	}
}

// version still beats an unknown flag (PRD §6.3: version precedes everything but help).
func TestRunVersionBeatsUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--version", "--bogus"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(--version --bogus): code=%d; want 0 (version precedence)", code)
	}
	if got := out.String(); got != "skpp "+version+"\n" {
		t.Errorf("stdout=%q; want the version line (version beats unknown flag)", got)
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr=%q; want empty (version won; unknown flag not reported)", errOut.String())
	}
}

// help beats an unknown flag (help is highest precedence).
func TestRunHelpBeatsUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--help", "--bogus"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("run(--help --bogus): code=%d; want 0 (help precedence)", code)
	}
	if !strings.Contains(out.String(), "USAGE:") {
		t.Errorf("stdout should be help, not an unknown-flag error:\n%s", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr=%q; want empty (help won)", errOut.String())
	}
}

// --- run: mode mutual exclusivity → exit 2 (P1.M5.T1.S1) ---

func TestRunExclusivityTagsAndList(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"foo", "--list"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(foo --list): code=%d; want 2", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "cannot be combined") {
		t.Errorf("stderr=%q; want an exclusivity message", errOut.String())
	}
}

func TestRunExclusivityTagsAndSearch(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"foo", "--search", "q"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(foo --search q): code=%d; want 2", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
}

func TestRunExclusivityTagsAndAll(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"foo", "--all"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(foo --all): code=%d; want 2", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
}

func TestRunExclusivityCheckAndTags(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"check", "foo"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(check foo): code=%d; want 2 (check + tag)", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "check") {
		t.Errorf("stderr=%q; want a message mentioning check", errOut.String())
	}
}

func TestRunExclusivityCheckAndList(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"check", "--list"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("run(check --list): code=%d; want 2 (check + mode)", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
}

// --- run: `check` subcommand dispatch (P1.M5.T1.S1; depends on P1.M4.T10.S1) ---
// These verify the DISPATCH. The exact exit code / output of a clean store is
// owned by T10.S1; if T10 has NOT landed, t10CheckDelegate returns exit 1 with a
// "not yet implemented" message, so the dispatch test asserts THAT shape instead
// (see the t.Skip note). Once T10 lands, un-gate and assert exit 0 + OK output.

func TestRunCheckDispatchesToCheck(t *testing.T) {
	dir := sampleStore(t) // a clean store (example + writing/reddit)
	t.Setenv("SKPP_SKILLS_DIR", dir)
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	// If internal/check exists (T10 landed), expect exit 0 + an OK line.
	// Otherwise the placeholder exits 1 with a clear message.
	if _, err := os.Stat("../internal/check"); err == nil {
		// T10 present: clean store → exit 0, an OK line on stdout.
		if code != 0 {
			t.Fatalf("run(check) clean store, T10 present: code=%d; want 0", code)
		}
		if !strings.Contains(out.String(), "OK") {
			t.Errorf("stdout=%q; want an OK report line", out.String())
		}
	} else {
		// T10 absent: placeholder exits 1 with a clear not-implemented message.
		if code != 1 {
			t.Fatalf("run(check) with T10 pending: code=%d; want 1 (placeholder)", code)
		}
		if !strings.Contains(out.String(), "not yet implemented") {
			t.Errorf("stdout=%q; want the placeholder not-implemented message", out.String())
		}
	}
}

// `check` with an unresolvable skills dir → exit 1, empty stdout, the one-line
// fix (shared with every other mode's Find() failure). This is independent of T10.
func TestRunCheckSkillsDirUnresolvable(t *testing.T) {
	unsetSkillsEnv(t)
	t.Chdir(t.TempDir()) // all three §8 rules miss
	var out, errOut bytes.Buffer
	code := run([]string{"check"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("run(check) unresolvable: code=%d; want 1", code)
	}
	if out.Len() != 0 {
		t.Errorf("stdout=%q; want empty", out.String())
	}
	if !strings.Contains(errOut.String(), "SKPP_SKILLS_DIR") {
		t.Errorf("stderr=%q; want the one-line fix", errOut.String())
	}
}
```

> `TestRunCheckDispatchesToCheck` uses `os.Stat("../internal/check")` to detect
> T10. If that path check is awkward in the test cwd, replace with a build-tag
> or simply assert the placeholder shape and flip when T10 lands — the point is
> the DISPATCH is exercised either way.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: PRECONDITION — confirm the post-T9 baseline + read T10's check signature
  - COMMAND: cd /home/dustin/projects/skpp
  - COMMAND: go test . -count=1 >/dev/null && echo "main green"   # expect 69 tests
  - COMMAND: grep -q 'searchMode' main.go && grep -q 'for i := 0; i < len(args)' main.go && echo "T9 landed"
  - EXPECT: T9 (--search) LANDED (index-loop parseArgs + searchMode/searchQ + run search branch).
            If parseArgs is STILL a range loop / searchMode is absent, T9 has NOT landed — STOP.
  - COMMAND: ls internal/check/ 2>/dev/null && echo "T10 check package present" || echo "T10 ABSENT"
  - IF PRESENT: read internal/check/*.go; note the exact exported function signature; plan
                to replace t10CheckDelegate's body with the real call in Task 4.
  - IF ABSENT: keep the t10CheckDelegate placeholder; TestRunCheckDispatchesToCheck asserts the
               placeholder shape; flag T10 as the one unfinished seam in the final report.
  - COMMAND: go test ./internal/{discover,resolve,search,skillsdir,ui}/ >/dev/null && echo "deps green"

Task 1: MODIFY main.go — usageText + usage() (Edit 1a)
  - WRITE: the usageText const + usage() func verbatim (Blueprint Edit 1a).
  - CHECK: raw string literal ends in exactly one '\n'; contains USAGE/EXAMPLES/OPTIONS
           sections + the canonical pi --skill "$(skpp example)" example; NO ANSI.
  - GOTCHA: PLAIN text only. fmt.Fprint (not Fprintln) where it is printed.

Task 2: MODIFY main.go — config struct (Edit 1b)
  - REPLACE the post-T9 config struct with the 3-field-added version (help/check/unknownFlag);
    drop the "// Future …" comment entirely (matrix complete).
  - GOTCHA: gofmt will realign comments — run `gofmt -w main.go` after.

Task 3: MODIFY main.go — parseArgs (Edit 1c)
  - ADD: `start`/check-first-arg pre-check; `case "--help","-h":`; unknown-flag capture
         in the default branch. Leave the --search case + every other case byte-identical.
  - CHECK: `skpp check` → c.check true, no tags; `skpp foo check` → tags=[foo check];
           `skpp --bogus` → c.unknownFlag="--bogus"; `skpp --search -x` → searchQ="-x" (unchanged).

Task 4: MODIFY main.go — run() reorder + exclusivityError + runCheck (Edits 1d/1e/1f)
  - REORDER run(): help → version → unknownFlag → exclusivity → dispatch(check…tags) → no-args-usage.
  - KEEP every mode branch BODY byte-identical (path/list/search/all/tags); only the
    preamble/coda + the inserted check dispatch change.
  - ADD exclusivityError (3 families exactly) + runCheck (Find+Index preamble + delegate).
  - IF Task 0 found internal/check: REPLACE t10CheckDelegate's body with the real
    `check.Report(stdout, skills)` (or adapt to T10's actual signature); DELETE the
    placeholder. ELSE: keep the placeholder.
  - GOTCHA: the no-args default is now `fmt.Fprint(stderr, usage()); return 1` (was silent return 1).

Task 5: MODIFY main_test.go — update the 3 existing tests + append ~18 new (File 2)
  - UPDATE: TestRunDefaultNoArgs (stderr usage), TestRunDefaultUnknownFlag (exit 2),
            TestParseArgsUnknownTolerated → TestParseArgsUnknownFlagCaptured.
  - APPEND: the ~18 tests from File 2 (help, no-args, unknown, exclusivity, check dispatch).
  - CHECK: reuses sampleStore/writeSkillTree/withTerminal/unsetSkillsEnv; existing 66
           other tests UNCHANGED.
  - GOTCHA: NO t.Parallel() on env/cwd tests (repo convention). TestRunCheckDispatchesToCheck
            adapts its assertion to whether internal/check exists.

Task 6: FORMAT + VET + BUILD + TEST (validation gates — run in order)
  - COMMAND: gofmt -w main.go main_test.go
  - COMMAND: gofmt -l main.go main_test.go   # MUST print nothing
  - COMMAND: go vet ./...                    # MUST be clean
  - COMMAND: go mod tidy   # EXPECTED: a NO-OP (stdlib only; no new module)
  - COMMAND: go build ./...                  # exit 0
  - COMMAND: go test . -v                    # ~87 main tests (69 + ~18)
  - COMMAND: go test ./...                   # whole module green (~195)
  - EXPECT: zero errors, zero vet findings, gofmt silent, go.mod/go.sum unchanged.

Task 7: SMOKE + SCOPE CHECK — Levels 3 + 4 in the Validation Loop.
```

### Implementation Patterns & Key Details

```go
// PATTERN: precedence tier (help highest; item "help wins" tiebreak).
//   if c.help       { fmt.Fprint(stdout, usage()); return 0 }   // STDOUT, exit 0
//   if c.version    { fmt.Fprintf(stdout, "skpp %s\n", version); return 0 }
//   if c.unknownFlag != "" { fmt.Fprintf(stderr, "skpp: unknown flag '%s'\n", …); return 2 }
//   if bad, _ := exclusivityError(c); bad { fmt.Fprintln(stderr, msg); return 2 }
//   …dispatch… ; fmt.Fprint(stderr, usage()); return 1   // no-args → STDERR, exit 1
// WHY: PRD §6.3 ("--help/--version take precedence") + item ("help wins"). The
//      unknown-flag + exclusivity checks sit BELOW help/version (so --help/--version
//      still win) but ABOVE dispatch (so a bad combo never partially runs).

// PATTERN: same usage text, two destinations (--help→stdout exit0; no-args→stderr exit1).
//   const usageText = `…`                  // plain; ends in one '\n'
//   fmt.Fprint(stdout, usage())   // --help
//   fmt.Fprint(stderr, usage())   // no-args / modifiers-only
// WHY: PRD §6.3 parity with get-server-config.sh. fmt.Fprint (no extra '\n').

// PATTERN: check subcommand = first-token detection (start the loop at index 1).
//   if len(args) > 0 && args[0] == "check" { c.check = true; start = 1 }
// WHY: item (e) "positional check as first arg". `check` elsewhere is an ordinary
//      tag (falls through default → not dashed → tag). Avoids a fragile scan.

// PATTERN: unknown-flag capture = FIRST offender only, in the default branch.
//   default:
//     if strings.HasPrefix(a, "-") {
//         if c.unknownFlag == "" { c.unknownFlag = a }   // first wins
//     } else { c.tags = append(c.tags, a) }
// WHY: one loud error (PRD §6). --search's consumed value never reaches default (i++ skip).
```

### Integration Points

```yaml
CLI SURFACE (final, §6.1–§6.4):
  - flags: --help/-h [NEW], --version/-v, --path/-p, --list/-l, --search/-s,
           --all/-a, --file/-f, --relative, --no-color
  - subcommand: check (first token)
  - exit codes: 0 success/help/version | 1 no-args/no-skills/unresolved/unresolvable
                | 2 unknown flag / mutually-exclusive modes
  - stdout discipline (§6.4): EMPTY on every non-success path (incl. the new exit-2s)

CONFIG:
  - struct main.config gains: help bool; check bool; unknownFlag string

MAIN DISPATCH:
  - file: main.go
  - order: help → version → unknownFlag → exclusivity → (check→path→list→search→all→tags) → no-args-usage
  - new helpers: usage()/usageText, exclusivityError(), runCheck() (+ t10CheckDelegate placeholder)

CHECK DISPATCH (consumes P1.M4.T10.S1):
  - runCheck: skillsdir.Find() → discover.Index() → T10's engine (adapt signature)
  - if T10 absent: t10CheckDelegate returns exit 1 + "not yet implemented" (never silent success)

DEPENDENCIES (go.mod): UNCHANGED. stdlib only (fmt/io/os/strings/path/filepath). go mod tidy is a no-op.

NO CHANGES TO:
  - internal/{discover,resolve,search,skillsdir,ui}/* ; internal/check/* (T10-owned, consume only)
  - go.mod / go.sum / PRD.md / tasks.json ; skills/ ; install.sh ; README.md ; completions
```

## Validation Loop

### Level 1: Syntax & Style (Immediate Feedback)

```bash
cd /home/dustin/projects/skpp
gofmt -w main.go main_test.go
gofmt -l main.go main_test.go   # MUST print nothing
go vet ./...                    # MUST be clean
# go mod tidy   # OPTIONAL sanity — EXPECTED no-op (diff go.mod before/after)
go build ./...                  # exit 0
# Expected: zero gofmt output; zero vet findings; build succeeds.
```

### Level 2: Unit Tests (Component Validation)

```bash
# The new parseArgs + run branches — fastest feedback.
go test . -run 'Help|NoArgs|ModifiersOnly|UnknownFlag|UnknownShortFlag|VersionBeatsUnknown|HelpBeatsUnknown|Exclusivity|CheckDispatchesToCheck|CheckSkillsDirUnresolvable|DefaultNoArgs|DefaultUnknownFlag' -v
# Expected: all the new + updated tests PASS.

# Full main suite — confirms the 3 updated tests + no regression in the other 66.
go test . -v
# Expected: ~87 main tests PASS (69 baseline + ~18 new; 3 updated in place).

# Whole module — confirms the read-only packages are untouched.
go test ./... -count=1
# Expected: ALL PASS. discover=39, resolve=13, skillsdir=29, ui=11, search=16, main=~87.
```

### Level 3: Integration Testing (System Validation)

```bash
cd /home/dustin/projects/skpp
go build -o skpp . && echo OK

# 1) --help → STDOUT usage, exit 0, stderr empty, NO ANSI.
out=$(./skpp --help 2>err.txt); rc=$?
[ "$rc" = "0" ] && echo "$out" | grep -q 'USAGE:' && echo "$out" | grep -q 'EXAMPLES:' \
  && echo "$out" | grep -q 'OPTIONS:' && echo "$out" | grep -qF 'pi --skill "$(skpp example)"' \
  && [ ! -s err.txt ] && ! printf '%s' "$out" | grep -q $'\x1b' && echo "HELP OK"

# 2) -h short form works identically.
./skpp -h | grep -q 'USAGE:' && echo "HELP-SHORT OK"

# 3) No args → STDERR usage, exit 1, stdout empty.
out=$(./skpp 2>err.txt); rc=$?
[ "$rc" = "1" ] && [ -z "$out" ] && grep -q 'USAGE:' err.txt && echo "NO-ARGS OK"

# 4) Unknown flag → stderr exact line, exit 2, stdout empty.
out=$(./skpp --bogus 2>err.txt); rc=$?
[ "$rc" = "2" ] && [ -z "$out" ] && [ "$(cat err.txt)" = "skpp: unknown flag '--bogus'" ] && echo "UNKNOWN-FLAG OK"

# 5) Short unknown flag.
out=$(./skpp -z 2>err.txt); rc=$?
[ "$rc" = "2" ] && [ "$(cat err.txt)" = "skpp: unknown flag '-z'" ] && echo "UNKNOWN-SHORT OK"

# 6) Mode exclusivity: tag + --list → exit 2, stdout empty.
out=$(./skpp foo --list 2>err.txt); rc=$?
[ "$rc" = "2" ] && [ -z "$out" ] && grep -qi 'cannot be combined' err.txt && echo "EXCLUSIVITY OK"

# 7) check + tag → exit 2.
out=$(./skpp check foo 2>err.txt); rc=$?; [ "$rc" = "2" ] && [ -z "$out" ] && echo "CHECK+TAG OK"

# 8) help beats version.
[ "$(./skpp --help --version 2>/dev/null | head -1)" = "skpp — skill path printer" ] && echo "HELP>VERSION OK"

# 9) version beats unknown flag.
[ "$(./skpp --version --bogus 2>/dev/null)" = "skpp $(./skpp --version | awk '{print $2}')" ] && echo "VERSION>UNKNOWN OK" || true
#   (simpler: ./skpp --version --bogus should print exactly the version line, exit 0)
./skpp --version --bogus >/dev/null 2>&1; [ $? = 0 ] && echo "VERSION>UNKNOWN exit-0 OK"

# 10) `check` dispatches (clean store → exit 0 if T10 landed; else placeholder exit 1).
TMPROOT=$(mktemp -d); mkdir -p "$TMPROOT/example"
printf -- '---\nname: example\ndescription: A demo skill.\n---\nx\n' > "$TMPROOT/example/SKILL.md"
out=$(SKPP_SKILLS_DIR="$TMPROOT" ./skpp check 2>err.txt); rc=$?
if [ -d internal/check ]; then [ "$rc" = "0" ] && echo "$out" | grep -q OK && echo "CHECK-DISPATCH(T10) OK"
else [ "$rc" = "1" ] && echo "$out" | grep -q 'not yet implemented' && echo "CHECK-DISPATCH(placeholder) OK"; fi
rm -rf "$TMPROOT" err.txt skpp
# Expected: every "… OK" line prints; the no-args/unknown/exclusivity branches empty stdout.
```

### Level 4: Creative & Domain-Specific Validation (Scope & Contract)

```bash
cd /home/dustin/projects/skpp

# SCOPE: go.mod / go.sum UNCHANGED (stdlib only).
git diff --quiet go.mod go.sum && echo "go.mod/go.sum unchanged OK" || echo "FAIL: go.mod/go.sum changed"

# SCOPE: the read-only packages (and T10's check pkg) untouched by THIS task.
git diff --name-only internal/discover internal/resolve internal/search internal/skillsdir internal/ui | (! read) \
  && echo "read-only packages untouched OK"

# CONTRACT: §6.4 holds on the new exit-2 paths (stdout empty on unknown + exclusivity).
./skpp --bogus 2>/dev/null | (! read) && echo "unknown-flag stdout-empty OK"
./skpp foo --list 2>/dev/null | (! read) && echo "exclusivity stdout-empty OK"

# CONTRACT: help precedence is help > version > everything.
./skpp --help --version --bogus 2>/dev/null | grep -q 'USAGE:' && echo "help-highest-precedence OK"

# CONTRACT: `check` is a subcommand only as the first token.
[ "$(./skpp --help 2>/dev/null | grep -c 'check')" -ge 1 ] && echo "help lists check OK"

# Expected: all scope/contract guards pass; no collateral changes; §6.4 discipline verified.
```

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l main.go main_test.go` silent; `go vet ./...` clean; `go build ./...` ok.
- [ ] Level 2 PASS — `go test ./... -count=1` green (~195 tests; main ~87).
- [ ] Level 3 PASS — every `… OK` smoke line printed (help/no-args/unknown/exclusivity/check).
- [ ] Level 4 PASS — scope guards confirm no collateral changes; §6.4 stdout-empty verified.
- [ ] `go mod tidy` was a no-op (`go.mod`/`go.sum` byte-identical).

### Feature Validation
- [ ] `--help`/`-h` → STDOUT usage (USAGE/EXAMPLES/OPTIONS + canonical `pi --skill` example), exit 0, no ANSI, stderr empty.
- [ ] `--help` beats `--version` and beats unknown flags.
- [ ] `--version` still beats unknown flags and every mode.
- [ ] No args (and modifiers-only) → STDERR usage, exit 1, stdout empty.
- [ ] Unknown flag (long + short) → `skpp: unknown flag '<x>'` on stderr, stdout empty, exit 2; first offender reported.
- [ ] tags + (--list/--search/--all) → stderr error, stdout empty, exit 2.
- [ ] `check` + tags, and `check` + (--list/--search/--all) → stderr error, exit 2.
- [ ] `check` as first token → subcommand dispatch; `check` elsewhere → ordinary tag.
- [ ] `check` with unresolvable dir → exit 1, empty stdout, one-line fix.
- [ ] Existing version/path/list/search/all/tags behavior + §6.4 atomicity unchanged.

### Code Quality Validation
- [ ] `run()` precedence reordered cleanly; mode branch bodies byte-identical (only preamble/coda changed).
- [ ] `usageText` is plain, ends in one newline, lists the full §6 matrix.
- [ ] `exclusivityError` implements exactly the 3 specified families (no over-engineering).
- [ ] `runCheck` preamble is correct; the T10 delegate is a single, clearly-marked seam (or real call).
- [ ] The 3 updated tests + ~18 new tests reuse existing helpers; existing 66 tests untouched.
- [ ] Comments explain the *why* (precedence order, plain help, check-as-first-arg, stdout discipline).

### Documentation & Deployment
- [ ] The help text is self-consistent with the README to be written in P1.M6.T16.S2 (same flag names/semantics).
- [ ] No new env vars; exit codes documented in the help block's footer line.

---

## Anti-Patterns to Avoid

- ❌ Don't check `version` before `help` — the item explicitly makes help win the tie.
- ❌ Don't print ANYTHING to stdout on the exit-2 paths (unknown flag / exclusivity) — §6.4.
- ❌ Don't collect a *list* of unknown flags — one loud error (the first) is the §6 contract.
- ❌ Don't treat `check` anywhere-but-first as the subcommand — only `args[0]=="check"`.
- ❌ Don't add ANSI color to the help block — plain text, unconditionally (research §2).
- ❌ Don't implement the §9 validation rules here — that's P1.M4.T10.S1; this task owns DISPATCH only.
- ❌ Don't invent an `internal/check` stub that produces fake OK/ERROR output — if T10 is absent, use the explicit "not yet implemented" placeholder (Task 4), never silent success.
- ❌ Don't touch the `--search` case or revert parseArgs to a range loop — layer on top of T9.
- ❌ Don't add unspecified exclusivity families (e.g. `--list --search`) — the dispatch order handles those; only the 3 specified families exit 2.

---

## Contingency — if P1.M4.T10.S1 (`check`) has NOT landed

This task's CORE deliverable (help + no-args + unknown-flag + mode-exclusivity)
is **fully independent of T10** and MUST ship green regardless. Only the
`check`-dispatch *delegate* depends on T10:

1. Implement Edits 1a–1e + the `runCheck` preamble (Find+Index) verbatim.
2. Keep the `t10CheckDelegate` placeholder (Edit 1f) — it prints
   `skpp: check is not yet implemented (P1.M4.T10.S1 pending)` to stdout and
   returns 1. This makes `skpp check` fail loudly (never silently succeed).
3. `TestRunCheckDispatchesToCheck` asserts the placeholder shape (exit 1 +
   "not yet implemented") when `internal/check` is absent.
4. When T10.S1 lands, a follow-up replaces `t10CheckDelegate`'s body with the
   real `check.Report(stdout, skills)` call (adapt to T10's signature), deletes
   the placeholder, and flips the test to assert exit 0 + an OK line.

This keeps main.go **compiling + green in both worlds** and avoids a §9 engine
that would conflict with T10's real package.
