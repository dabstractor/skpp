# Verified Facts — P1.M1.T2.S3 (walk-up rule 3 + `Find()` combiner)

All facts below were **empirically verified** in the target environment
(`go version go1.26.4-X:nodwarf5 linux/amd64`, module `github.com/dabstractor/skpp`,
go 1.25) on 2026-07-06. Every code snippet in the PRP was pre-run.

---

## 1. CRITICAL — `filepath.Glob` does NOT support `**` (PRD §8.3 suggestion is wrong for Go)

The PRD item description suggests *"use filepath.WalkDir to check, or
filepath.Glob `skills/**/SKILL.md`"*. The Glob alternative is **non-functional**
in Go's stdlib: `path/filepath.Match`/`Glob` treat `**` the same as `*`
(single-level only). Verified:

```
tree:  skills/foo/bar/SKILL.md   (nested 2 deep)
filepath.Glob("<root>/skills/**/SKILL.md")  -> []   (0 matches)  ← BROKEN
filepath.Glob("<root>/skills/*/SKILL.md")   -> []   (0 matches, nested 2 deep)
filepath.Glob("<root>/skills/*/*/SKILL.md") -> [1 match]
```

**Decision: use `filepath.WalkDir` with an early-exit sentinel.** It recurses to
arbitrary depth and matches `discover.Index` (which also uses WalkDir). Do NOT
use Glob with `**`. This is the single most important correctness gotcha.

## 2. `filepath.WalkDir` early-exit via sentinel error — works

Returning any non-nil error from a `WalkDir` callback **stops the walk**. Pattern
to find "at least one SKILL.md anywhere":

```go
var errSkillMDFound = errors.New("SKILL.md found")
func hasSkillMD(dir string) bool {
    found := false
    _ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return nil }                 // skip unreadable, keep walking
        if !d.IsDir() && d.Name() == "SKILL.md" {
            found = true
            return errSkillMDFound                   // stop the walk
        }
        return nil
    })
    return found
}
```

Verified: `hasSkillMD` on a tree with `skills/a/b/SKILL.md` → true; on an empty
`skills/` dir → false; on a `skills/` with only `README.md` → false (name must
be exactly `SKILL.md`). The sentinel error is swallowed by the `_=`.
**Requires `import "io/fs"`** for the `fs.DirEntry` callback param type.

## 3. Walk-up loop termination — `filepath.Dir(root) == root`

`filepath.Dir("/")` returns `"/"` (not `""`). So the ascent loop terminates when
`parent == cur`. Verified:
```
filepath.Dir("/")      == "/"
filepath.Dir("/a")     == "/"
filepath.Dir("/a/b")   == "/a"
```
Loop body:
```go
cur := filepath.Clean(start)
for {
    candidate := filepath.Join(cur, "skills")
    if info, err := os.Stat(candidate); err == nil && info.IsDir() {
        if hasSkillMD(candidate) { return candidate, true }
        // skills exists but has no SKILL.md -> fall through, keep ascending
    }
    parent := filepath.Dir(cur)
    if parent == cur { return "", false }   // reached root
    cur = parent
}
```
Verified: from `/a/b/c` the loop runs 4 steps (c→b→a→/→stop).

## 4. §8.3 "skip empty skills dir, keep ascending" — correct & verified

PRD §8.3 qualifies the match: *"the first ancestor containing a skills/ subdir
**with at least one SKILL.md** wins."* So an ancestor that has a `skills/` dir
with NO `SKILL.md` does **not** win — ascent must continue to higher ancestors.

Verified with this layout:
```
root/a/sub          <- start
root/a/skills       <- EMPTY (no SKILL.md)  [lower ancestor]
root/skills/foo/SKILL.md                   <- REAL  [higher ancestor]
```
`findWalkUpAncestor(root/a/sub)` → returns `root/skills` (the higher one), NOT
`root/a/skills`. The inner `if hasSkillMD(...)` falling through (no return) is
what makes this work — do NOT add an early `return "", false` when a `skills/`
dir exists but is empty.

## 5. Testability — `t.Chdir` (Go 1.24+) controls `os.Getwd()` for `findWalkUp`

`os.Getwd()` cannot be parameterized, but `findWalkUp` calls it, so the rule-3
entry is only testable by changing cwd. **`testing.T.Chdir(dir)`** (added Go
1.24) changes cwd for the test and auto-restores on cleanup — the cwd analog of
`t.Setenv`. go.mod is `go 1.25`, so it is available. Verified it compiles + runs:

```
package main_test ... t.Chdir(sub)   // compiles, runs, restores cwd
```

Pattern (mirrors S2's `resolveSiblingFromExe` extraction for `findSibling`):
- **Testable core:** `findWalkUpAncestor(start string) (dir, found)` — pure, no
  cwd dependency. Drive with `t.TempDir()` trees.
- **Thin entry:** `findWalkUp() (dir, src, found)` = `os.Getwd()` → core.
  Test it by `t.Chdir`-ing into a temp subdir whose ancestor has skills/.

Both `t.Chdir` and `t.Setenv` forbid `t.Parallel()` — none of these tests may be
parallel (same constraint as S1's env tests).

## 6. No spurious ancestor `skills/` on this host → smoke tests are deterministic

Checked every ancestor of the repo (`~/projects/skpp`) up to `/`: **none** has a
`skills/` dir. The repo itself has no `skills/` yet (created in P1.M6.T12). So:
- `findWalkUp()` called from the repo cwd with no test setup → `found=false`
  (deterministic, today). **BUT** once P1.M6.T12 adds `skills/example/`, a bare
  "findWalkUp from repo" assertion would flip to true and break.
- **Therefore: ALL findWalkUp/Find-rule-3 tests `t.Chdir` into their OWN temp
  tree** (hermetic + forward-compatible). Do NOT assert findWalkUp behavior
  against the real repo cwd.

## 7. `findSibling` deterministically MISSES inside `go test` (so `Find()` can reach rule 3)

In a test process `os.Executable()` is the test binary at
`/tmp/go-buildXXXXX/b001/exe/<pkg>.test`; that dir has no sibling `skills/`, so
`findSibling()` returns `found=false`. This means `Find()` (with env unset +
`t.Chdir` into a temp tree with an ancestor skills) reaches rule 3 and returns
`SourceWalkUp` — **the `Find()` rule-3-win test is deterministic.** (Symmetric to
S2: rule 2's resolution is tested via `resolveSiblingFromExe`, not via `Find`.)

## 8. Imports required by S3 (a MODIFY, not an append)

`internal/skillsdir/skillsdir.go` currently imports (after S1+S2; S2 adds none):
```go
import (
    "os"
    "path/filepath"
)
```
S3 must ADD `"errors"` (for `errors.New` on `ErrNotFound`) and `"io/fs"` (for the
`fs.DirEntry` WalkDir callback param). Result:
```go
import (
    "errors"
    "io/fs"
    "os"
    "path/filepath"
)
```
(Already gofmt-sorted alphabetically; `errors` < `io/fs` < `os` < `path/filepath`.)

`internal/skillsdir/skillsdir_test.go` currently imports `os`, `path/filepath`,
`testing`. S3 must ADD `"errors"` (for `errors.Is`) and `"strings"` (for
`strings.Contains` on the `ErrNotFound` message).

## 9. Exact user-facing error wording (PRD §6.4 / §8.4)

PRD §6.4 (line 139): *"Skills dir cannot be located ⇒ stderr: concise reason +
the fix (`set $SKPP_SKILLS_DIR`, or `cd` into the repo, or reinstall), exit `1`."*
PRD §8.4 (line 192): *"None found ⇒ stderr error + exit `1`, with a one-line fix."*

The fix phrase is: `set $SKPP_SKILLS_DIR, or cd into the repo, or reinstall`.
The error message (printed verbatim to stderr by main in P1.M1.T3) is:

```
could not locate the skills directory: set $SKPP_SKILLS_DIR, cd into the skpp repo, or reinstall skpp
```

Defined as an exported sentinel `ErrNotFound` so tests can `errors.Is` it and
main can print `err.Error()` directly (no wrapping).

## 10. `findWalkUp` returns `(dir, src, found bool)` — NOT `(dir, src, nil)`

The work-item contract loosely says rule 3 should "return (that skills dir,
SourceWalkUp, nil)". The `nil` describes the error-free success case. The
**locked per-rule helper signature** (S1/S2) is `func findX() (dir string, src
Source, found bool)` — no error return; `found` is the win signal; `src`'s zero
value is ignored on miss so `Find()` falls through. `findWalkUp` MUST match this
shape so `Find()` can chain `findEnv → findSibling → findWalkUp` uniformly. Only
`Find()` returns an error, and only when all three miss. (Verified against S1's
`research/verified_facts.md §4` and S2's PRP.)

## 11. `Find()` signature is locked by `architecture/go_architecture.md`

```go
// Find locates the skills dir per §8 priority. Returns absolute path + which
// rule won. Returns error if none found (caller prints fix hint + exit 1).
func Find() (dir string, src Source, err error)
```
Downstream consumers: `main.go` (P1.M1.T3 `--path`) and `discover.Index`
(P1.M2.T5). This subtask's `Find()` is the public entrypoint they both call.
