# Verified Facts — P1.M5.T3.S1 (README sync)

Scope: README.md only. Mode B (changeset-level documentation sweep). The code
changes shipped in P1.M1.T1.S1–P1.M4.T2.S1 are DONE; `main.go`'s `usageText`
is ALREADY synced (it lists all 6 search fields). Only README.md is stale.

All facts below were verified by reading source + running the binary in
`~/projects/skpp` (commit at time of writing). No source file is edited by this
task — README.md is the ONLY deliverable.

## 1. `--path` discovery-source reporting (P1.M1.T1.S1) — SHIPPED

**Wire-up** — `main.go`, the `if c.path {` branch:
```go
dir, src, err := skillsdir.Find()
...
fmt.Fprintln(stdout, dir)                         // stdout: ONLY dir + newline
fmt.Fprintf(stderr, "(found via %s)\n", src)      // stderr: source label
```
So stdout is byte-identical to pre-change (§13 gate `test "$(./skpp --path)" =
"$PWD/skills"` still holds); the source rides on **stderr**.

**Source labels** — `internal/skillsdir/skillsdir.go` `Source.String()`:
- `SourceEnv`      → `"SKPP_SKILLS_DIR"`
- `SourceSibling`  → `"sibling of binary"`
- `SourceWalkUp`   → `"ancestor of cwd"`

**Verified binary output** (`cd ~/projects/skpp`):
```
$ ./skpp --path              # stdout
/home/dustin/projects/skpp/skills
$ ./skpp --path 2>&1 >/dev/null   # stderr only
(found via sibling of binary)
```

Why it matters (PRD §8 / QA Issue 1): a typo'd `SKPP_SKILLS_DIR` silently falls
through to sibling/walk-up; the stderr label is the only way to tell the env var
was ignored. README currently (line 243) says only "`skpp --path` reports which
directory won." — STALE: it omits the stderr source label entirely.

## 2. `--search` field expansion (P1.M2.T1.S1) — SHIPPED

**Match scope** — `internal/search/search.go` `matches()`: six fields, ALL OR'd:
`RelTag`, `Name`, `Description`, each `Keyword`, each `Alias`, `Category`.
Aliases/keywords are matched INDIVIDUALLY (boundary-safe). Empty query matches
all (== `--list`).

**main.go usageText is ALREADY synced** (verified via `./skpp --help`):
```
--search <q>, -s   Substring search over tag / name / description / keywords / aliases / category
```
and the EXAMPLES line `--search reddit  # substring search over tag/name/description/keywords/aliases/category`.

So `--help` is correct; README is the stale surface. README Usage (line 131)
shows only `skpp --search reddit` with no field list → add the six fields.

## 3. Unicode table width (P1.M3.T1.S1) — SHIPPED (no README change required)

`internal/ui/ui.go`: `displayWidth(s)` = `utf8.RuneCountInString(s)`; `padRight`
and the column/wrap math use it. Multi-byte runes (é, —) now pad correctly. Known
limitation (documented in code): wide CJK runes counted as 1 cell (dependency-free
trade-off per PRD §4/§7.3). 

README does NOT make any ASCII/byte-width claim about tables, so NO edit is
required for this change. Listed here for completeness only — do NOT touch README
for unicode.

## 4. Combined short flags + `--flag=value` (P1.M4.T1.S1) — SHIPPED

`main.go` `parseArgs` + `expandShortBundle`:
- `--flag=value`: split on first `=`; bool flags ignore the value, `--search`
  takes it as the query. Verified: `--search=reddit` works.
- Short bundles `-xyz`: e.g. `-af` (all+file), `-sfoo` (search "foo"). Unknown
  char rejects the WHOLE bundle (two-phase validate-then-commit).

README Usage shows only canonical long/short forms. Contract says "optionally
mention" these. Safe VALID example combos (do not conflict with mode
exclusivity): `--search=reddit` (= form); `-af` / `--all --file` (modifier +
single listing mode). AVOID conflict-prone examples like `-pl` (path+list → exit 2),
`-la` (list+all → exit 2), `-vh` (help wins, prints help).

## 5. Conflicting listing modes (P1.M4.T2.S1) — SHIPPED

`main.go` `exclusivityError`: any 2+ of `{path, list, searchMode, all}` → exit 2.
Verified binary output:
```
$ ./skpp --list --search foo
skpp: listing modes --path/--list/--search/--all are mutually exclusive
exit=2
```
(check+tags and check+mode and tags+mode remain exit 2 per §6.3.)

README does NOT currently document mode exclusivity; the "Error contract."
paragraph (Usage, lines ~150-154) is the natural home. Contract: "if README
documents mode exclusivity, note listing modes are mutually exclusive." Optional
but adds a one-line note that prevents surprise.

## 6. README section map (what touches --path / --search / flag syntax / modes)

| README line(s) | Section | Current text (abridged) | Stale? | Action |
|---|---|---|---|---|
| 129–131 | Usage | `# Human-readable catalog and substring search` / `skpp --search reddit` | YES (no field list) | add 6 fields |
| 139–140 | Usage | `# Where is the resolved skills directory?` / `skpp --path # → /…/skills` | YES (no stderr label) | note stderr source |
| ~150–154 | Usage | `**Error contract.** ...unknown tag...nothing to stdout...` | PARTIAL (no mode rule) | add listing-mode exclusivity (optional) |
| 158 | Usage | `` `skpp --help` lists every flag.`` | fine | optional: append flag-syntax note |
| 200–212 | Adding a skill | frontmatter ex w/ `category`/`aliases`; "optional" bullet | accurate | optional: note aliases/category are searchable |
| 243 | How skpp finds the store | `` `skpp --path` reports which directory won.`` | YES (omits stderr source) | rewrite to cover stdout+stderr+labels+typo fall-through |

## 7. Validation approach (Mode B = manual review)

No test snapshots README.md or `usageText` (grep of main_test.go + internal/*:
only unrelated `README.md` fixture mentions in skillsdir/discover tests). The
item's TESTS step is **manual**: spot-check `./skpp --path`, `./skpp --search`,
`./skpp --help` and confirm README claims match. Build gate (`go build -o skpp .`)
ensures the binary the README describes actually runs. No `go test` change.

## 8. Conflict / parallel-context check

- This task edits **README.md only**. P1.M5.T2.S1 (parallel) edits
  `internal/check/check.go` only; P1.M5.T1.S1 (done) edited `.gitignore`. Zero
  overlap. README.md is not touched by any other in-flight item.
- `PRD.md` is READ-ONLY (human-owned) — do NOT edit. §6.1 still lists the OLD
  4-field search list (tag/name/description/keywords); §10 says aliases/category
  "enrich --search". The QA Issue 4 resolution (decisions.md §D4) made §10 win and
  the code matches §10. README must reflect the **shipped** (6-field) behavior,
  NOT §6.1's stale summary. Do not "fix" PRD §6.1.
