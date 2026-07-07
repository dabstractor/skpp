# Verified Facts — P1.M2.T5.S1: `internal/discover` — `Index()` walk + relTag normalization + sorting

Every load-bearing claim below was **executed** (not reasoned about) in a throwaway
module `/tmp/skpp_index_verify` against the **real landed** `internal/discover`
code: S1's `discover.go` (`Frontmatter` + `ParseFrontmatter`) and S2's `skill.go`
(`Skill` + `BuildSkill` + `toStringSlice`) were copied **verbatim** (only the
package line flipped to `main` for the throwaway run). The proposed `Index()`
algorithm was compiled and run; `index_test.go`'s 12 tests all pass against it.
Nothing in the repo was touched.

## Repo state at authoring time (read directly)

```
go version go1.26.4-X:nodwarf5 linux/amd64
go.mod     module github.com/dabstractor/skpp ; go 1.25 ; require gopkg.in/yaml.v3 v3.0.1
go.sum     yaml.v3 v3.0.1 present
internal/discover/discover.go       S1 landed: Frontmatter(8 fields) + ParseFrontmatter + utf8BOM
internal/discover/discover_test.go  S1 landed: 12 white-box tests + writeSkill helper (package discover)
internal/discover/skill.go          S2 landed: Skill(9 fields) + BuildSkill + toStringSlice
internal/discover/skill_test.go     S2 landed: toStringSlice + BuildSkill tests (+ strEq helper)
internal/skillsdir/skillsdir.go     M1.T2: Find() + findEnv/findSibling/findWalkUp (returns ABSOLUTE dir)
main.go / main_test.go               M1.T3 landed (builds + tests green)
baseline: `go build ./...` OK ; `go test ./...` OK (skillsdir + discover + main all green)
NO internal/discover/index.go yet (THIS subtask). NO resolve/ ui/ skills/ yet.
```

## Architecture contract (read from `plan/001_fcde63e5bb60/architecture/go_architecture.md`)

```go
// Index walks absSkillsDir and returns every skill (dir containing SKILL.md),
// sorted by RelTag for deterministic output.
func Index(absSkillsDir string) ([]Skill, error)
```
- Data flow (line 23): `discover.Index(absSkillsDir) ──▶ []Skill, error`.
- relTag normalization (lines 114-118): `filepath.Rel(skillsDir, skillDir)` then
  `filepath.ToSlash(...)`; canonical tag comparisons are on the slash form.
- `discover` is a LEAF library; `resolve` imports it (`[]discover.Skill`).
- S2's PRP already pre-specified the T5 call site: `fm, _, err :=
  ParseFrontmatter(filepath.Join(dir, "SKILL.md")); s := BuildSkill(dir, relTag,
  fm)` and that **"T5 owns: the walk, relTag computation, sorting by RelTag, and
  the err policy (malformed YAML → surface to `check` M4)."**

## Run 1 — the naive `Index` (Stat guard MISSING): the missing-root BUG

First attempt used the obvious callback shape, mirroring `skillsdir.hasSkillMD`:

```go
filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if err != nil { return nil }   // <-- swallows the ROOT error too
    ...
})
```

Fixture: `skills/example`, `skills/writing/reddit` (nested), `skills/writing/empty`
(no `---` block), `skills/bad` (malformed YAML), plus stray `README.md` and a
`writing/notes/draft.txt`.

```text
=== sorted Index(root) ===
err=<nil> len=4
  RelTag="bad"            Dir-absolute=true Name=""      HasFM=false   (malformed YAML, included)
  RelTag="example"        Dir-absolute=true Name="example" HasFM=true
  RelTag="writing/empty"  Dir-absolute=true Name=""      HasFM=false   (no frontmatter, included)
  RelTag="writing/reddit" Dir-absolute=true Name="reddit" HasFM=true
nested relTag="writing/reddit"   OK (separator normalized to /)
malformed YAML is INCLUDED (HasFM=false), not aborted   OK

=== missing root returns error ===
  missing-root err != nil: false        <-- *** BUG: a missing root returns (nil, nil) ***
empty skills dir -> len=0 err=<nil>     OK
root-level SKILL.md -> relTag="."       OK (filepath.Rel(root,root) edge case)
relative input -> Dir absolute          OK
```

**THE BUG (critical):** `filepath.WalkDir` on a non-existent root calls the
callback ONCE with `(root, nil, lstatErr)`. The `if err != nil { return nil }`
line — correct for *per-entry* errors — **swallows the root error**, so WalkDir
returns nil and `Index` returns `(nil, nil)`. A missing/unreadable skills dir
would silently look like "no skills", which breaks the PRD §6.4/§8.4 "skills dir
cannot be located ⇒ stderr error + exit 1" contract. **Fix: a `Stat`-guard BEFORE
the walk.**

## Run 2 — the corrected `Index` (Stat-guard + the resolved behaviors)

```go
root, err := filepath.Abs(skillsDir)   // guarantee absolute (protects §6.1/§13)
if err != nil { return nil, err }
info, err := os.Stat(root)             // *** the fix: root must exist & be a dir ***
if err != nil { return nil, err }
if !info.IsDir() { return nil, errors.New(root + ": not a directory") }
filepath.WalkDir(root, func(path, d, err) error {
    if err != nil { return nil }       // now ONLY per-entry errors are skipped
    if d.IsDir() || d.Name() != "SKILL.md" { return nil }
    skillDir := filepath.Dir(path)
    rel, _ := filepath.Rel(root, skillDir)
    relTag := filepath.ToSlash(rel)
    fm, _, _ := ParseFrontmatter(path) // lenient: malformed YAML -> Frontmatter{} -> HasFM=false
    skills = append(skills, BuildSkill(skillDir, relTag, fm))
    return nil
})
sort.Slice(skills, func(i,j int) bool { return skills[i].RelTag < skills[j].RelTag })
```

```text
=== missing root now returns error ===
  err != nil: true   (err=stat .../does-not-exist: no such file or directory)
=== root is a FILE -> error ===
  err != nil: true   (err=.../notadir...: not a directory)
=== per-entry error does NOT abort (unreadable subdir skipped) ===
  partial err=<nil> len=1 (chmod-000 'locked' subdir skipped; 'ok' kept)
=== symlinked skill dir NOT followed by WalkDir (default) ===
  symlinked tree found 1 skill(s)   (only 'real', NOT the 'linked' symlink -> dir)
    RelTag="real"
=== sort order is by RelTag (lexicographic) ===
    bad, example, writing/reddit
```

## Run 3 — the full `index_test.go` suite (12 tests)

`index_test.go` (white-box `package discover`; reuses `writeSkill` from
`discover_test.go` and `strEq` from `skill_test.go`; defines ONE new helper
`makeTree`; uses `t.Chdir` from Go 1.24+ for the relative-input test; go.mod is
`go 1.25` so `t.Chdir` is available):

```text
=== RUN   TestIndexSingle ....................... PASS
=== RUN   TestIndexNestedRelTag ................. PASS   (relTag == "writing/reddit")
=== RUN   TestIndexSortedByRelTag ............... PASS   (apple < mango/beta < mango/fig < zebra)
=== RUN   TestIndexNoFrontmatterIncluded ........ PASS   (HasFM=false, still indexed)
=== RUN   TestIndexMalformedYAMLNotAborted ...... PASS   (len==2; "bad" included HasFM=false)
=== RUN   TestIndexIgnoresNonSkillMD ............ PASS   (README.md / draft.txt / SKILL.md.bak ignored)
=== RUN   TestIndexEmptyDir ..................... PASS   (len==0, nil err)
=== RUN   TestIndexMissingRoot .................. PASS   (err != nil, the Stat-guard)
=== RUN   TestIndexRootIsFile ................... PASS   (err != nil)
=== RUN   TestIndexNestedBothLevels ............. PASS   (writing AND writing/reddit both indexed)
=== RUN   TestIndexRootLevelSkillMD ............. PASS   (relTag == "." edge case)
=== RUN   TestIndexRelativeInputDirStillAbsolute  PASS   (Dir absolute even for relative input)
PASS
ok  	skpp_index_verify	0.003s
```
`gofmt -l` silent; `go vet ./...` clean.

## Decisions locked (each traceable to a run above)

1. **Signature is `func Index(skillsDir string) ([]Skill, error)`** — exactly the
   architecture contract (go_architecture.md line 53). `[]Skill` (value slice,
   matching S2's value-semantics `BuildSkill`); `error` is for walk-level
   failure. ✓ Run 2.
2. **A `Stat`-guard BEFORE `WalkDir` is MANDATORY.** Without it a missing root
   returns `(nil, nil)` (Run 1 BUG) because WalkDir feeds the root's `lstat`
   error to the callback, and the per-entry `if err != nil { return nil }` swallows
   it. Fix: `os.Stat(root)` → propagate `err`; then `!info.IsDir()` → `errors.New`.
   `os.Stat` (not `Lstat`) follows a symlinked root, matching `skillsdir.findEnv`.
   ✓ Run 2.
3. **`filepath.Abs(skillsDir)` first → `Skill.Dir` is ALWAYS absolute.** This
   protects the PRD §6.1 ("absolute path") and §13 (`case "$(./skpp example)" in
   /*)`) contracts even if a caller passes a relative path (Run 1/3 relative-input
   test). On the canonical absolute input (from `skillsdir.Find`) it is a no-op
   (`filepath.Abs` just `Clean`s an already-absolute path). ✓ Run 3.
4. **relTag = `filepath.ToSlash(filepath.Rel(root, skillDir))`.** Matches
   go_architecture.md lines 114-118 verbatim. Nested skills → `writing/reddit`.
   No backslash on any platform (`ToSlash`). ✓ Run 1/3.
5. **Sort by `RelTag`, lexicographic (string `<`).** Deterministic output for
   `--all`/`--list`. `sort.Slice`. Verified order: `apple < mango/beta < mango/fig
   < zebra`. ✓ Run 3.
6. **A skill = any directory that DIRECTLY contains a `SKILL.md`.** The callback
   acts only on non-dir entries named exactly `SKILL.md`; `skillDir =
   filepath.Dir(path)`. Stray `README.md`/`draft.txt`/`SKILL.md.bak` and empty
   subdirs are ignored. ✓ Run 3 (`TestIndexIgnoresNonSkillMD`).
7. **Nested skills at multiple levels are ALL indexed.** `writing/SKILL.md` AND
   `writing/reddit/SKILL.md` both exist → both returned (`writing`, then
   `writing/reddit`). PRD §7.1 explicitly allows nested skills. ✓ Run 3
   (`TestIndexNestedBothLevels`).
8. **Malformed YAML does NOT abort the walk and is NOT propagated.** `ParseFrontmatter`
   returns `(Frontmatter{}, body, err)` for broken YAML; `Index` ignores `err`
   and calls `BuildSkill(dir, relTag, Frontmatter{})` → a `HasFM=false` Skill that
   is STILL resolvable by directory/basename (PRD §7.1). This is the "err policy"
   S2's PRP assigned to T5. **M4/`check` can re-run `ParseFrontmatter(s.SourceFile)`
   if it needs to distinguish "malformed YAML" from "no frontmatter block"** —
   documented forward-compat, NOT rework (the re-parse is idempotent & cheap).
   ✓ Run 1/3.
9. **Per-entry errors are SKIPPED, not fatal.** An unreadable subtree
   (`chmod 000`) is skipped; the rest of the walk continues. Matches
   `skillsdir.hasSkillMD`'s `if err != nil { return nil }`. ✓ Run 2.
10. **Empty skills dir → `(nil, nil)`** (nil slice, nil error). Callers (`--list`
    exits 1, `--all` prints nothing) test with `len()`. ✓ Run 1/3.
11. **`WalkDir` does NOT follow symlinked directories (stdlib default).** A symlink
    to a skill dir is NOT descended into (Run 2 found only `real`, not `linked`).
    PRD §7.1 does not require following symlinks; the default avoids cycles.
    Documented behavior, no action needed.
12. **Root-level `SKILL.md` → `relTag == "."`.** `filepath.Rel(root, root) == "."`.
    Included for spec-compliance (the dir DOES contain `SKILL.md`); unusual in
    practice (canonical layout nests under `<tag>/`). Documented edge case.
    ✓ Run 1/3 (`TestIndexRootLevelSkillMD`).
13. **`index.go` imports ONLY stdlib:** `errors`, `io/fs`, `os`, `path/filepath`,
    `sort`. No `fmt` (the verifier's `fmt` was only in its throwaway `main()`).
    No new module dep → `go.mod`/`go.sum` UNCHANGED. ✓ Run 3 (`go vet` clean).
14. **`index.go` is a NEW file, NOT appended to `discover.go`.** S1's `discover.go`
    package doc explicitly says "The Index() walk (T5)... are LATER subtasks — do
    not add them here." S2's anti-pattern list reserved `Index()` for T5 in its own
    file. Tests live in `index_test.go` (white-box `package discover`). ✓
15. **`index_test.go` reuses cross-file helpers, defines ONE new helper.** Reuses
    `writeSkill` (from `discover_test.go`) and `strEq` (from `skill_test.go`) —
    same `package discover`, so redefining them is a compile error. Defines only
    `makeTree(t, layout map[string]string) string` (a SHARED skills tree, unlike
    `writeSkill`'s isolated single file). `t.Chdir` (Go 1.24+) scopes the cwd
    change in the relative-input test. NO testify, NO `t.Parallel()`. ✓ Run 3.

## Reproducing these runs

```bash
cd /home/dustin/projects/skpp
go test ./internal/discover/ -run TestIndex -v   # the 12 NEW tests
go test ./internal/discover/                     # + S1's 12 + S2's tests, all green
go doc ./internal/discover Index                 # exported Index + its godoc
```

The throwaway verifier lived at `/tmp/skpp_index_verify` (built during authoring;
copies of the real `discover.go`/`skill.go` with the package line flipped to
`main`; not part of the repo).
