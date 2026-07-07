# Verified facts — P1.M4.T10.S1 (`skpp check`, PRD §9)

Research notes for the PRP. Every claim below was checked against the on-disk
code at `~/projects/skpp` (post P1.M4.T9.S1) or against live Go output.

## 1. What exists and is GREEN (the contract this task builds on)

- `discover.Index(dir string) ([]Skill, error)` — `internal/discover/index.go`.
  Walks the tree; a "skill" = any dir directly containing `SKILL.md`; returns
  `[]Skill` **sorted by RelTag**. Returns `(nil, err)` if the root is
  missing/not-a-dir; **swallows** per-entry errors and **per-skill malformed YAML**.
- `discover.Skill` — `internal/discover/skill.go`. Fields consumed by check:
  `RelTag string`, `Name string`, `Description string`, `SourceFile string`
  (== Dir+"/SKILL.md"), `HasFM bool`. `Name`/`Description` are "" and `HasFM`
  is false for both "no frontmatter block" and "malformed YAML" (see §2).
- `discover.ParseFrontmatter(path) (fm Frontmatter, body string, err error)` —
  `internal/discover/discover.go`. **This is what check re-runs.** Returns:
  - no `---` block / opening fence with no close → `Frontmatter{HasFM:false}`,
    body=whole file, **nil err** (lenient).
  - syntactically broken YAML between valid fences → `Frontmatter{HasFM:false}`,
    body=post-fence text, **non-nil err** (the yaml.v3 error). This is the ONLY
    way to tell "malformed YAML" from "no block" — and `Index` drops this err.
- `index.go` doc comment (lines ~38-44) EXPLICITLY says: *"check (M4/T10) can
  re-run `ParseFrontmatter(s.SourceFile)` to distinguish 'malformed YAML' from
  'no frontmatter block' (idempotent; no rework)."* → this task's design is the
  one the codebase already documents.
- `skillsdir.Find() (dir, src, err)` — resolves the store (§8); `err` is already
  a one-line user-facing fix (print verbatim to stderr). `main.run` calls it.
- Baseline tests GREEN: `go test ./...` → 166 top-level `func Test`
  (main=69, discover=31, resolve=10, search=16, skillsdir=29, ui=11).

## 2. The §9 "skill dir has no SKILL.md" bullet — reframed (DECISION)

PRD §9 lists "ERROR: skill dir has no SKILL.md". Under directory-based discovery
(`Index` only emits dirs that CONTAIN a SKILL.md) this literal case **can never
arise from the index**. Detecting "dirs that lack SKILL.md but look like skills"
would need a fragile heuristic (does it contain `scripts/`/`assets/`? is it a
leaf?) and would false-positive on legitimate grouping dirs like `skills/writing/`.
DECISION: reframe the bullet to **"unusable SKILL.md"** via the re-parse —
`ParseFrontmatter` returning `err != nil` (malformed YAML OR unreadable file,
e.g. a vanishing-file race) → ERROR. This is the actionable analog and is exactly
what `index.go`'s own comment directs. Clearly documented in the PRP; the
literal "dir without SKILL.md" scan is OUT OF SCOPE.

## 3. The §9 "empty besides SKILL.md" WARN — SKIPPED (DECISION)

PRD marks it "optional". The shipped `example` skill (§11) IS only `SKILL.md`, so
enabling this WARN would make the §13 acceptance line (`./skpp check` "reports
the example as OK") false. DECISION: do NOT implement this WARN. Documented.

## 4. Agent Skills `name` rule — regex verified live

PRD §3: "1–64 chars, lowercase a-z0-9-, no leading/trailing/consecutive hyphens."
Verified (live `go test`) that `^[a-z0-9]+(-[a-z0-9]+)*$` accepts `example`,
`foo-helper`, `a`, `123`, `a-b-c` and rejects `-foo`, `foo-`, `foo--bar`, `Foo`,
`foo_bar`, `foo.bar`, `` (empty). The regex enforces charset + structure
(non-empty, no leading/trailing/consecutive hyphens); the **only** thing it does
NOT enforce is the 64-char max, so check adds an explicit `len > 64` test FIRST.
Names are ASCII → byte length == rune count; `len(name)` is correct.

## 5. description length — measure TRIMMED (DECISION)

`skill.go`: description is copied VERBATIM, including a folded-scalar trailing
newline. `ui.go` displays `strings.TrimSpace(Description)`. DECISION: check's
1024-char WARN measures the **trimmed** length (consistent with the displayed
length; a folded-scalar trailing newline does not count). A whitespace-only
description trims to "" → ERROR "missing or empty" (not a WARN).

## 6. Output format — "one line per problem" (DECISION)

PRD §9: "one line per skill → `OK <relTag> (<name>)`; problem lines prefixed
ERROR/WARN". Read as: each OK skill emits ONE `OK` line; a skill with multiple
problems emits ONE line PER problem (each prefixed). This preserves all info
(e.g. invalid name AND too-long description → 1 ERROR + 1 WARN for one skill).
Status word is `%-5s` (padded to width 5): `OK   `, `WARN `, `ERROR`.
Summary line: `N skills, M errors, K warnings` (always plural, literal PRD form).

## 7. Stream + exit code — check is a REPORT, not a path emitter (DECISION)

check prints its **full findings to STDOUT** regardless of pass/fail (pipeable to
less/grep, like eslint/ruff/govet) and signals pass/fail via the exit code
(0 if no ERROR, 1 if any ERROR; WARNs never change exit code). This is NOT §6.4's
"nothing on stdout on failure" — that contract is for tag/path emitters used in
`$(...)`; check never participates in command substitution. Empty store →
`0 skills, 0 errors, 0 warnings`, exit 0 (clean). Contrast `--list` (exit 1 on
empty) — check is validation, empty = nothing wrong.

## 8. duplicate-name scan — global, 2nd pass, case-sensitive on non-empty names

Collect non-empty `name` (only from skills whose frontmatter parsed OK) →
`name -> []relTag`. For any name with >1 owner, append an ERROR to EACH owner:
`duplicate frontmatter 'name' "<name>" (also in: <other relTags>)`. Others are
sorted for deterministic output. A skill with a missing name is NOT in the map
(it already has its own "missing name" ERROR). Two skills sharing an INVALID name
string (e.g. both "Bad Name") still count as a dup (literal string match).

## 9. The `check` SUBCOMMAND token — reserved, recognized in parseArgs

`check` is a positional token (no dash). Currently `parseArgs` default-branch
captures any non-dash token into `c.tags`. This task adds `case "check":` →
`c.check = true` (NOT added to tags). `check` is a RESERVED subcommand name: a
user who names a skill `check` cannot resolve it via `skpp check` (collision;
matches how CLIs reserve subcommand names). M5.T11 owns the §6.3 exclusivity
error (`check` mixed with tags/`--list`/etc. → exit 2); until then `check` wins
silently in `run()` dispatch (mirrors how `searchMode` currently wins over tags).

## 10. Package + import — `internal/check`, mirrors `internal/search`

New package `internal/check` (mirrors search→`internal/search`: subcommand name →
package name). `func Check(skills []discover.Skill) Report` returns STRUCTURED
findings (no I/O for output); `main.run` renders the report to stdout. Split =
deterministic check tests (assert on Report struct, no string parsing) + main
tests assert on rendered text. Imports: stdlib only (`fmt`, `regexp`, `sort`,
`strings`) + `internal/discover`. **No new third-party dependency**; `go mod
tidy` is a no-op. Importing `internal/check` as `check` does NOT collide with
`c.check` (struct field) or `case "check":` (string literal) — different
namespaces.

## 11. Dispatch order in run() (where the check branch goes)

`run()` precedence after this task: `--version` → `--path` → `--list` →
`--search` → **`check`** → `--all` → `<tags>` → default. `check` slots with the
other report modes (list/search), before the path-emitting modes (all/tags).
`--file`/`--relative`/`--no-color` do NOT apply to check (it prints a status
report, not paths/tables).
