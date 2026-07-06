# Verified Facts — P1.M1.T2.S2 (skillsdir: sibling-of-binary resolution, rule 2)

All facts below were **empirically verified** with Go 1.26.4-X:nodwarf5 on
linux/amd64 (the build environment) by running a throwaway Go program. Rule 2 is
pure standard library (`os`, `path/filepath`); no third-party code is involved.

## 1. Input dependency (from P1.M1.T2.S1 — ALREADY IMPLEMENTED)

This subtask **modifies** the package produced by P1.M1.T2.S1, which now exists
at `internal/skillsdir/skillsdir.go`. It provides (verified by reading the file):

- `package skillsdir` with imports `"os"` + `"path/filepath"`.
- `type Source int` + constants `SourceEnv`, `SourceSibling`, `SourceWalkUp`.
- `func (s Source) String() string` → `"SKPP_SKILLS_DIR"` / `"sibling of binary"`
  / `"ancestor of cwd"` / `"unknown"`.
- `const envVar = "SKPP_SKILLS_DIR"` and `func findEnv() (dir string, src Source, found bool)`.
- `internal/skillsdir/skillsdir_test.go` (white-box, `package skillsdir`) with the
  `unsetEnvVar(tb)` helper + 8 test functions.

The per-rule helper signature `func findX() (dir string, src Source, found bool)`
is **locked by S1** (see `P1.M1.T2S1/research/verified_facts.md §4`). S2 must add
`findSibling()` matching that exact shape.

S2 adds **no** new dependency (pure stdlib). `go.mod` / `go.sum` untouched.

## 2. The crux: how to test `os.Executable()`

`os.Executable()` returns the path of the **currently running** binary. In a
`go test`/`go run` process it is NOT controllable. Empirically verified:

```
os.Executable()     = "/tmp/go-build819269557/b001/exe/sibverify"
EvalSymlinks(exe)   = "/tmp/go-build819269557/b001/exe/sibverify"  (same — no symlink)
Dir(real)           = "/tmp/go-build819269557/b001/exe"            (temp build dir)
```

The build-temp dir (`/tmp/go-buildXXXXX/b001/exe/`) has **no** sibling `skills/`
subdirectory. Therefore the REAL `findSibling()` deterministically returns
`found=false` inside the test process. **Consequence:** the symlink-resolution
behavior — the entire point of rule 2 — cannot be exercised through `findSibling()`
alone. The fix is to **extract the testable core** into a helper that takes the
exe path as a parameter:

```go
// findSibling is the rule-2 entry: it asks the OS for the running binary, then
// delegates the symlink/dir logic to resolveSiblingFromExe (the unit-testable core).
func findSibling() (dir string, src Source, found bool) {
    exe, err := os.Executable()
    if err != nil { return "", 0, false }
    d, ok := resolveSiblingFromExe(exe)
    if !ok { return "", 0, false }
    return d, SourceSibling, true
}

// resolveSiblingFromExe is the symlink-aware sibling-resolution logic, factored out
// so it can be tested with arbitrary exe paths (t.TempDir + os.Symlink).
func resolveSiblingFromExe(exe string) (dir string, found bool)
```

This mirrors the S1 design precedent (thin rule entry + testable core) and keeps
`findSibling` matching the locked `(dir, src, found)` shape.

## 3. Verified: the symlink-install scenario (mirrors verified_symlink_resolution.md)

Setup: a fake "binary" (regular file — see §5) in `tempA` with a sibling
`skills/`; a symlink to that binary placed in a DIFFERENT dir `tempB`. Invoke
`resolveSiblingFromExe` with the **symlink path** (`tempB/skpp`):

```
link = "/tmp/linkB.../skpp"  ->  /tmp/realA.../skpp
resolveSiblingFromExe(link)
  got="/tmp/realA.../skills"  found=true
  == tempA/skills?  true        <-- resolves back to the REAL binary's dir
```

This is exactly what makes `~/.local/bin/skpp → ~/projects/skpp/skpp` resolve
back to `~/projects/skpp/skills` (PRD §8.2, §12.1). The mechanism:
`EvalSymlinks(link)` → real binary path → `Dir()` → real binary's dir →
`Join(dir,"skills")` → `Stat` → existing dir → win.

## 4. Verified: the EvalSymlinks-error fallback

The contract requires: "`real, err := filepath.EvalSymlinks(exe)`; if err, use
`exe` as fallback." Verified with a "ghost" exe that does not exist but whose
parent dir + sibling `skills/` DO exist:

```
ghost = "/tmp/fallbackC.../does-not-exist-binary"   (Stat -> no such file)
resolveSiblingFromExe(ghost)
  EvalSymlinks errors -> real := exe (fallback)
  Dir(exe)        = "/tmp/fallbackC..."
  candidate      = "/tmp/fallbackC.../skills"  (exists, IsDir)
  got="/tmp/fallbackC.../skills"  found=true
```

So the fallback is real and must be preserved (it lets rule 2 still work when
`EvalSymlinks` cannot resolve for any reason — e.g. a broken intermediate path
on some platforms). Implemented as `if err != nil { real = exe }`.

## 5. Verified: a regular file suffices as a fake "binary"

`filepath.EvalSymlinks` and `os.Stat(filepath.Join(dir,"skills"))` do NOT require
the exe to be a real ELF executable. Creating the fake binary with
`os.WriteFile(path, []byte("x"), 0o644)` is enough — `EvalSymlinks` resolves the
symlink to that file, `Dir()` gives its parent. So tests need NOT `go build` a
helper binary; `t.TempDir()` + `os.WriteFile` + `os.Symlink` + `os.Mkdir` is the
full toolkit (matching the contract's "Use `t.TempDir()` + `os.Symlink`").

Also verified: direct (non-symlink) exe with a sibling `skills/` resolves
correctly (`found=true`), which covers the non-symlinked-binary install case too.

## 6. Verified: "no sibling skills/" and "skills is a file" both miss

- Exe exists, parent dir exists, but NO `skills/` sibling → `found=false`.
- Sibling path `skills` is a regular FILE (not a dir) → `os.Stat(...).IsDir()==false`
  → `found=false`. (Same `IsDir` guard rule 1 uses.)

## 7. Cross-platform note (from architecture/verified_symlink_resolution.md)

On **Linux**, `os.Executable()` already returns the REAL path via `/proc/self/exe`,
so `EvalSymlinks` is redundant-but-harmless. On **macOS**, `os.Executable()` may
return the symlink path, so `EvalSymlinks` is REQUIRED. **Do NOT "simplify" by
dropping `EvalSymlinks`** — it breaks macOS. Implement BOTH calls exactly as the
contract specifies. (This asymmetry vs rule 1 — which must NOT EvalSymlinks the
env path — is intentional and documented in S1's verified_facts.md §2.)

## 8. Platform-skip convention for symlink tests (from S1)

`os.Symlink` can fail on platforms that do not support symlinks. S1's
`TestFindEnvDoesNotResolveSymlinks` uses `t.Skipf("symlinks not supported...")`
on error. S2's symlink tests follow the same convention so the suite stays green
on all platforms while exercising the symlink path on linux/macOS. On linux/amd64
(the CI/host environment) the symlink tests RUN (not skipped).

## 9. "Unused symbol" note (carried from S1)

`findSibling` (and `resolveSiblingFromExe`) are not called until S3 wires them
into `Find()`. Go's `go build` / `go vet` do NOT flag unused package-level funcs,
and there is no linter configured for this repo (gates are `go build` / `go vet`
/ `go test` / `gofmt`). Do NOT delete the helpers or stub a `Find()` to "use"
them — `Find()` is S3's deliverable.
