# Verified Facts — P1.M1.T2.S1 (skillsdir: Source type + rule 1 env resolution)

All facts below were **empirically verified** with Go 1.26.4 on linux/amd64
(matching the build environment in `architecture/codebase_state.md`) by running
a throwaway Go program. Rule-1 resolution is pure standard library; no third-party
code is involved.

## 1. Input dependency (from P1.M1.T1.S1)

This subtask consumes the Go module produced by P1.M1.T1.S1:

- `go.mod` → `module github.com/dabstractor/skpp`, `go 1.25`, `require gopkg.in/yaml.v3 v3.0.1 // indirect`
- Internal packages import as `github.com/dabstractor/skpp/internal/<pkg>`.

This subtask adds **no** new dependency (pure stdlib: `os`, `path/filepath`).
`go.mod` / `go.sum` are NOT modified by this subtask.

## 2. Verified stdlib behavior — the crux of rule 1

Test program created a `realDir`, a `link` symlink → `realDir`, a `nope`
non-existent path, and a regular file `afile`. Output (verbatim):

```
os.Stat(symlink)   err=<nil>  IsDir=true
filepath.Abs(link) = "/tmp/par1145748910/link-to-skills"  err=<nil>
filepath.Abs(link) == realDir? false  (want FALSE: env path must stay as the link)
filepath.Abs(link) == link (abs form)? true
EvalSymlinks(link) = "/tmp/real-skills1598985500"  == realDir? true
os.Stat(nonexistent) err=stat .../nope: no such file or directory  IsNotExist=true
os.Stat(regularFile).IsDir=false (want FALSE -> rule falls through)
LookupEnv(unset)   val="" ok=false
LookupEnv(empty)   val="" ok=true
os.Stat("")        err=stat : no such file or directory
```

### Conclusions (all VERIFIED)

1. **`os.Stat` follows symlinks.** `os.Stat(symlinkToDir).IsDir()` returns `true`
   (reports the *target's* dir-ness). So a symlinked `SKPP_SKILLS_DIR` is
   correctly detected as a valid directory WITHOUT needing EvalSymlinks.
2. **`filepath.Abs` does NOT resolve symlinks.** It only makes the path absolute
   (joining with cwd if relative) and lexical-cleans it (`filepath.Clean`). The
   symlink path is preserved verbatim. This is exactly the contract:
   *"Do NOT EvalSymlinks the env path (user points exactly where they want)."*
   → Use `filepath.Abs(val)`, never `filepath.EvalSymlinks(val)` in rule 1.
3. **`filepath.EvalSymlinks` DOES resolve** (returns the real target). This is
   what rule 2 (sibling-of-binary) uses and what rule 1 must NOT. The contrast is
   intentional and is enforced by `TestFindEnvDoesNotResolveSymlinks`.
4. **Non-dir / non-existent / empty inputs all fall through** with no error:
   - `os.Stat(nonexistent)` → `os.IsNotExist` true
   - `os.Stat(regularFile).IsDir()` → false
   - `os.Stat("")` → error (so an empty `SKPP_SKILLS_DIR` behaves like unset)
5. **`os.LookupEnv` distinguishes unset from empty**: unset → `("",
   false)`; empty → `("", true)`. Both are treated identically by rule 1
   (fall through), so the choice of `os.LookupEnv` vs `os.Getenv` is
   behaviorally equivalent here; `LookupEnv` is used because it expresses
   "is it set?" idiomatically.

## 3. `t.Setenv` cannot unset — testing implication

`t.Setenv("X", v)` always **sets**; there is no `t.Unsetenv` in the stdlib
`testing` package. To test the "var unset" case deterministically (the parent
shell may have `SKPP_SKILLS_DIR` exported), the test uses a small
`unsetEnvVar(tb)` helper that does `os.Unsetenv` + `tb.Cleanup` to restore the
prior set/unset state. Because every rule-1 test mutates process env, **none**
of these tests may call `t.Parallel()` (documented inline).

## 4. Design decision — per-rule helper signature

The architecture doc fixes only the public combiner signature
(`func Find() (dir string, src Source, err error)`, built in P1.M1.T2.S3). The
per-rule helper signature is this subtask's design choice. Chosen shape:

```go
func findEnv() (dir string, src Source, found bool)
```

Rationale:
- **No `error` return.** Rule 1 explicitly never hard-errors (contract: "do not
  hard-error on a bad env value — let later rules try"). A bad/unset/empty value
  is expressed as `found=false`, not as an error. Rules 2 and 3 are equally
  best-effort, so `(dir, src, found)` is the consistent shape for all three.
- **`found bool`** is the unambiguous "did this rule win?" signal the `Find()`
  combiner needs to decide whether to fall through.
- **`src Source`** is returned so each rule self-describes which `Source` it won
  as, keeping `Find()` dumb (no rule→source mapping table). `src` is only
  meaningful when `found == true`.

Intended sibling helpers (added by later subtasks in the SAME file):
- P1.M1.T2.S2 → `func findSibling() (dir string, src Source, found bool)`
- P1.M1.T2.S3 → `func findWalkUp() (dir string, src Source, found bool)` +
  the public `func Find() (dir string, src Source, err error)` that chains:
  `findEnv` → `findSibling` → `findWalkUp` → `ErrNotFound`.

## 5. "Unused symbol" note

`findEnv` (unexported) is not called until S3 wires it into `Find()`. Go's
`go build` and `go vet` do NOT flag unused package-level functions, so the
package compiles cleanly in this subtask with `findEnv` present-but-uncalled.
Do NOT delete it to "fix" a linter — there is no linter configured for this
repo (validation gates are `go build` / `go vet` / `go test` / `gofmt`).

## 6. Source.String() labels (locked from the work-item contract)

| Source        | String() label        | Used by            |
|---------------|-----------------------|--------------------|
| `SourceEnv`   | `"SKPP_SKILLS_DIR"`   | `skpp --path` (T3) |
| `SourceSibling` | `"sibling of binary"` | `skpp --path` (T3) |
| `SourceWalkUp`  | `"ancestor of cwd"`   | `skpp --path` (T3) |
| out-of-range   | `"unknown"`           | defensive default  |

These exact strings are part of the contract; `--path` reporting (P1.M1.T3) and
README examples depend on them. Make `Source` satisfy `fmt.Stringer` implicitly.
