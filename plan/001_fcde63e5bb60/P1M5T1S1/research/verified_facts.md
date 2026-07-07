# Verified facts — P1.M5.T1.S1 (`--help` + no-args/unknown-flag/mode-exclusivity exit codes)

> All facts below were read directly from the working tree at
> `/home/dustin/projects/skpp` on 2026-07-06. Every quoted code fragment is the
> CURRENT on-disk state (post M3.T8.S2; `main.go` line counts grow once
> P1.M4.T9.S1 lands — see §6). Go toolchain: `go1.26.4-X:nodwarf5 linux/amd64`.

## 1. What this subtask owns (contract from the item + PRD §6)

The CLI-surface *gate* (build-order step 5). Four behaviors + the `check`
subcommand dispatch, each with a pinned stdout/stderr/exit-code discipline:

| Input                                  | stdout        | stderr                  | exit | PRD        |
| -------------------------------------- | ------------- | ----------------------- | ---- | ---------- |
| `skpp --help` / `-h`                   | full usage    | (empty)                 | 0    | §6.1, §6.3 |
| `skpp` (no args, no flag)              | (empty)       | usage                   | 1    | §6.3       |
| `skpp --bogus` (unknown dashed token)  | (empty)       | `skpp: unknown flag '<x>'` | 2 | §6 header |
| `skpp foo --list` (tags + a mode)      | (empty)       | error line              | 2    | §6.3       |
| `skpp check` (subcommand, first arg)   | check report  | (empty unless error)    | 0/1  | §6.1, §9   |

**Precedence decision (item: "both → help wins"):** `--help` beats `--version`
beats unknown-flag beats exclusivity beats normal dispatch beats no-args.
PRD §6.3 ("--help / --version take precedence over everything else") + the
item's explicit help-wins tiebreak → check `help` BEFORE `version` in `run()`.

## 2. mcpeepants `get-server-config.sh` help-text STRUCTURE (the thing to mirror)

Read in full from `~/projects/mcpeepants/get-server-config.sh`. The `--help`
branch prints, in order:

1. A bold-cyan title line + a one-line description.
2. **USAGE:** section — one `skpp <form>` line per invocation shape.
3. **EXAMPLES:** section — including the canonical `claude --mcp-config "$($0 ...)"`
   one-liner plus `--list` / `--search` examples.
4. **OPTIONS:** section — two aligned columns: `flag` … `one-line description`.
5. (mcpeepants also has a **FILES:** section; skpp has no manifest, so it is
   omitted or replaced by a one-line exit-code reference.)

skpp mirrors the STRUCTURE (USAGE / EXAMPLES / OPTIONS, aligned OPTIONS
columns) but swaps mcpeepants' `claude --mcp-config "$($0 ...)"` canonical
example for skpp's: `pi --skill "$(skpp <tag>)"`.

**Color decision (authoritative):** the help text is emitted PLAIN (no ANSI),
unconditionally. Reasons: (a) `skpp --help | grep` must work; (b) PRD §6.1's
`--help` row says only "Help text (to stdout)" with no color requirement; (c)
deterministic test output (tests use `*bytes.Buffer`, non-TTY); (d) §13 does
not assert on help color. `--no-color` still exists for `--list`/`--search`.
This is a deliberate, documented simplification — NOT a regression.

## 3. The current `main.go` `config` + `parseArgs` + `run` (what we edit)

`config` (current, post-M3.T8.S2; P1.M4.T9.S1 will have ADDED `searchMode`/
`searchQ` and converted `parseArgs` to an index loop by the time this runs):

```go
type config struct {
	version  bool     // --version / -v
	path     bool     // --path / -p
	list     bool     // --list / -l
	all      bool     // --all / -a
	file     bool     // --file / -f
	relative bool     // --relative
	noColor  bool     // --no-color
	tags     []string // positional <tag>
	// Future (M4/M5), do NOT add yet:  <-- T9 removes "search"; THIS task removes the rest
	//   search string; check bool; help bool
}
```

`parseArgs` default branch (current):
```go
default:
    if !strings.HasPrefix(a, "-") { c.tags = append(c.tags, a) }
    // dashed unknowns: tolerated (no-op)  <-- THIS task turns this into capture-for-exit-2
```

`run()` precedence (current): `version` → `path` → `list` → (T9: `search`) →
`all` → `tags` → default `return 1` (silent). **THIS task inserts `help` ABOVE
`version`, adds unknown-flag→exit2 + exclusivity→exit2 checks between version
and the mode dispatch, and turns the silent default into usage→stderr+exit1.**

## 4. The `check` subcommand — dependency on P1.M4.T10.S1 (CRITICAL)

`grep -rn "check" --include="*.go"` shows **NO check package and NO `c.check`
field exist yet** — only comments in `discover/*.go` and `main.go` naming T10/M4
as the future owner ("T10's length check trims…", "check (M4)"). P1.M4.T10.S1
("Validation rules (§9) + output format + exit codes") is **Planned** (not yet
in flight; no PRP on disk).

**Division of labor (authoritative):**
- **THIS task (M5.T1.S1)** owns the `check` DISPATCH only: `parseArgs`
  recognizes `check` as the FIRST arg → `c.check`; `check` is mutually exclusive
  with tags / `--list`/`--search`/`--all` (exit 2); `run()` routes the check
  branch. **It does NOT implement the §9 validation rules, the OK/WARN/ERROR
  output format, or the per-finding exit-code mapping** — those are T10.S1.
- **P1.M4.T10.S1** owns the validation engine (mirrors `internal/resolve`/
  `internal/search` as a pure function over `[]discover.Skill`, very likely an
  `internal/check` package) plus the report format + exit code.

**Integration seam (what this task wires, what it consumes):**
`runCheck` does the same `skillsdir.Find()` → `discover.Index()` preamble every
other mode does (fully specifiable here), then delegates the report+exit-code to
T10's entry point. Because T10's exact exported signature is unknown at PRP
time, the PRP specifies the preamble verbatim and marks the ONE delegate call as
"adapt to `internal/check`'s actual signature (read it on disk at impl time)".

Likely T10 contract (mirrors resolve/search — pure over the index):
```go
package check
// Report writes the §9 report to w and returns 0 if clean, 1 if any ERROR.
func Report(w io.Writer, skills []discover.Skill) int
```
The PRP's `runCheck` calls `check.Report(stdout, skills)` as the assumed shape,
with an explicit "verify on disk" precondition task. If T10 instead returns
findings, runCheck maps severity → exit code (0 if no `ERROR`, else 1 per §9).

## 5. Existing tests this task MUST change (exit-1 → exit-2, etc.)

`grep -c "^func Test" main_test.go` = **53 today** (pre-T9). After P1.M4.T9.S1
lands, main = **69** (53 + 16 search tests). THIS task's edits:

| Existing test (in main_test.go)        | Current assertion        | NEW assertion (this task)                            |
| -------------------------------------- | ------------------------ | ---------------------------------------------------- |
| `TestRunDefaultUnknownFlag`            | `run(--frobnicate)`→exit 1 | `run(--frobnicate)`→exit **2**, stderr `skpp: unknown flag '--frobnicate'`, stdout EMPTY |
| `TestParseArgsUnknownTolerated`        | unknown = no-op, tags captured | `c.unknownFlag == "--frobnicate"` captured (for exit-2); tags `[sometag check]` still captured (`check` here is 2nd positional → a tag, NOT the subcommand) |
| `TestRunDefaultNoArgs`                 | `run(nil)`→exit 1 only   | `run(nil)`→exit 1 AND stderr contains `USAGE` (usage now printed to stderr) |

Every OTHER existing test (version/path/list/tags/all/search precedence +
atomicity) is unaffected: none pass unknown flags, none mix tags+modes, none use
`--help`. `TestRunVersionPrecedenceOver*` all use `--version` (no `--help`), so
inserting `help` above `version` leaves them green.

## 6. The `check`-as-first-arg parse rule (and why "first arg")

Item (e): "`check` is a subcommand (positional `check` as first arg)". Decision:
`check` is the subcommand **iff `args[0] == "check"`** (literally the first
token). Consequences, all verified against the test matrix:
- `skpp check` → subcommand. ✓
- `skpp check foo` → subcommand + a tag `foo` → exclusivity exit 2. ✓ (item: check mutually exclusive with tag resolution)
- `skpp foo check` → `foo` is first positional, `check` is the 2nd positional →
  `check` is an ordinary TAG (NOT the subcommand). This is the only sane reading
  of "first arg" and keeps `check` resolvable as a skill tag if a user ever had
  one named `check`.

Implementation: detect `args[0]=="check"` BEFORE the token loop, set `c.check`,
and start the loop at index 1 (so the `check` token is not also captured as a
tag). A trailing `case "check":` is NOT needed — `check` in any non-first slot
falls through `default` → not dashed → captured as a tag.

## 7. Mode-exclusivity detection (exit 2)

PRD §6.3 + item (d)/(e) specify exactly three exclusivity families. The PRP
implements precisely these (no more, to avoid breaking the deterministic
dispatch order for unspecified combos like `--list --search`):

1. `len(tags) > 0 && (c.list || c.searchMode || c.all)` → exit 2  (PRD §6.3, item d)
2. `c.check && len(tags) > 0` → exit 2                            (item e: check vs tag resolution)
3. `c.check && (c.list || c.searchMode || c.all)` → exit 2        (extension: `check` is a standalone verb; only ADDS exit-2 cases, cannot break existing green tests)

`--file` / `--relative` / `--no-color` are MODIFIERS, never modes — they never
trigger exclusivity and may combine freely. `--list --search` (two modes, no
tags) is NOT specified; the existing dispatch order (list branch before search)
handles it deterministically and is left as-is.

## 8. Validation commands (verified working in THIS tree)

```bash
cd /home/dustin/projects/skpp
go version            # go1.26.4-X:nodwarf5 linux/amd64
go build ./...        # exit 0 (verified)
go test . -count=1    # ok  github.com/dabstractor/skpp  (53 tests today; 69 post-T9)
gofmt -l main.go main_test.go   # silent == clean
go vet .              # clean
```
Module: `github.com/dabstractor/skpp`; go 1.25 in go.mod; dep `gopkg.in/yaml.v3`.
This task adds NO dependency (stdlib only: `fmt`, `io`, `os`, `strings`,
`path/filepath` already imported). `go mod tidy` is a no-op.

## 9. Files this task touches (scope boundary)

- **MODIFY** `main.go` — add `help bool`, `check bool`, `unknownFlag string` to
  `config`; add `--help`/`-h` case + `check`-first-arg detection + unknown-flag
  capture to `parseArgs`; reorder `run()` (help→version→unknown→exclusivity→
  dispatch→no-args-usage); add `runCheck` + `usage()`/`usageText` +
  `exclusivityError`. Possibly import `internal/check` (T10).
- **MODIFY** `main_test.go` — update the 3 tests in §5; append ~18 new tests
  (help short/long, help-beats-version, no-args-stderr, unknown-flag exit2 x2,
  version/help-beats-unknown, exclusivity x4, check dispatch + unresolvable).
- **DO NOT TOUCH** `internal/{discover,resolve,search,skillsdir,ui}/*`,
  `go.mod`/`go.sum`, `PRD.md`, `tasks.json`, `skills/`, `install.sh`, `README.md`,
  completions. `internal/check/*` is OWNED by T10.S1 — consume only.
