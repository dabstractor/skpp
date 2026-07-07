# Verified facts — P1.M1.T1.S1 (Wire Source into --path stderr reporting)

Bug fix for QA Issue 1. All code locations, signatures, and labels verified by
direct inspection of the built repo on 2026-07-07. No new types; no
skillsdir.go changes; no docs changes (Mode A).

## 1. The exact code to change — main.go `c.path` branch (lines 268-281)

CURRENT (src discarded):

```go
if c.path {
    dir, _, err := skillsdir.Find() // src is for reporting only; not printed
    if err != nil {
        // Find() returns skillsdir.ErrNotFound whose message is the
        // user-facing one-line fix (PRD §8.4/§6.4). Print it verbatim to
        // stderr (NOT stdout) so $(...) stays empty on failure.
        fmt.Fprintln(stderr, err)
        return 1
    }
    // Byte-exact: ONLY the dir + newline. The §13 acceptance gate
    // `test "$(./skpp --path)" = "$PWD/skills"` depends on this.
    fmt.Fprintln(stdout, dir)
    return 0
}
```

TARGET (src wired to stderr):

```go
if c.path {
    dir, src, err := skillsdir.Find()
    if err != nil {
        // Find() returns skillsdir.ErrNotFound whose message is the
        // user-facing one-line fix (PRD §8.4/§6.4). Print it verbatim to
        // stderr (NOT stdout) so $(...) stays empty on failure.
        fmt.Fprintln(stderr, err)
        return 1
    }
    // Byte-exact: ONLY the dir + newline on stdout. The §13 acceptance gate
    // `test "$(./skpp --path)" = "$PWD/skills"` depends on this — $() captures
    // stdout only, so the stderr source label below does NOT break it.
    fmt.Fprintln(stdout, dir)
    // Issue 1 (QA): report which §8 discovery rule won, to stderr. A typo'd
    // SKPP_SKILLS_DIR silently falls through to sibling/walk-up; this label
    // makes that visible without polluting stdout. Labels come from
    // skillsdir.Source.String().
    fmt.Fprintf(stderr, "(found via %s)\n", src)
    return 0
}
```

Two real edits:
1. `dir, _, err := skillsdir.Find() // src is for reporting only; not printed`
   → `dir, src, err := skillsdir.Find()` (drop the now-false trailing comment).
2. Insert `fmt.Fprintf(stderr, "(found via %s)\n", src)` immediately after
   `fmt.Fprintln(stdout, dir)` and BEFORE `return 0`.
3. (Optional but recommended) update the two comments noted above to reflect
   that src IS now printed and stdout is still byte-exact because $() ignores
   stderr.

## 2. Why `%s` works on a `Source` value

`skillsdir.Source` is `type Source int` with a `func (s Source) String() string`
method (skillsdir.go:24-50). That satisfies `fmt.Stringer`, so `fmt.Fprintf(w,
"%s", src)` calls `src.String()` automatically. No manual conversion needed.

## 3. The exact label strings (from Source.String(), already unit-tested)

Verified in `internal/skillsdir/skillsdir_test.go:30-46 TestSourceString`:

| Source value   | .String() label      | full stderr line                  |
|----------------|----------------------|-----------------------------------|
| SourceEnv      | `SKPP_SKILLS_DIR`    | `(found via SKPP_SKILLS_DIR)\n`   |
| SourceSibling  | `sibling of binary`  | `(found via sibling of binary)\n` |
| SourceWalkUp   | `ancestor of cwd`    | `(found via ancestor of cwd)\n`   |
| out-of-range   | `unknown`            | `(found via unknown)\n`           |

The env case (SourceEnv → `SKPP_SKILLS_DIR`) is the only deterministic one to
assert in `run()` tests via `t.Setenv`. Sibling/walk-up depend on the binary
path / cwd and are NOT exercised through run() in this subtask (covered
indirectly by `TestSourceString` in the skillsdir package).

## 4. The exact tests to change — main_test.go

### TestRunPathSuccess (line ~169) — currently asserts stderr EMPTY (will break)

```go
if errOut.Len() != 0 {
    t.Errorf("run(--path) success stderr=%q; want empty", errOut.String())
}
```
This test sets `SKPP_SKILLS_DIR` via `t.Setenv`, so after the fix stderr will
hold `(found via SKPP_SKILLS_DIR)\n`. REPLACE the assertion:

```go
if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
    t.Errorf("run(--path) success stderr=%q; want %q (Issue 1 source label)", got, want)
}
```

### TestRunPathShortFlag (line ~187) — currently has NO stderr assertion

Add the same assertion after the stdout check:

```go
if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
    t.Errorf("run(-p) stderr=%q; want %q (Issue 1 source label)", got, want)
}
```

### NEW test — feature test for Issue 1 (stdout byte-exact + stderr label)

Distinct from the two updates (which just stop failing). This documents the
contract explicitly and survives future refactors of the above:

```go
// Issue 1 (QA): --path must report which §8 rule won to stderr, while stdout
// stays byte-exact so the §13 `test "$(./skpp --path)" = "$PWD/skills"` gate
// still passes. The env case is deterministic; sibling/walk-up are covered by
// skillsdir.TestSourceString.
func TestRunPathReportsSourceLabel(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("SKPP_SKILLS_DIR", dir) // rule 1 wins -> SourceEnv
    var out, errOut bytes.Buffer
    if code := run([]string{"--path"}, &out, &errOut); code != 0 {
        t.Fatalf("run(--path): code=%d; want 0", code)
    }
    // stdout: byte-exact dir + newline (the §13 contract is preserved).
    if got, want := out.String(), filepath.Clean(dir)+"\n"; got != want {
        t.Errorf("--path stdout=%q; want %q", got, want)
    }
    // stderr: the SourceEnv label, nothing else.
    if got, want := errOut.String(), "(found via SKPP_SKILLS_DIR)\n"; got != want {
        t.Errorf("--path stderr=%q; want %q", got, want)
    }
}
```

## 5. What must NOT change

- `internal/skillsdir/skillsdir.go` — no edits. Source/String()/Find() are
  already correct and tested.
- `main.go` const `usageText` (line 50, the `--path, -p Print the resolved
  skills directory` line at ~80) — Mode A = no help-text change.
- The `--path` FAILURE path (the `if err != nil { ... return 1 }` block) —
  unchanged. src is irrelevant on failure; stdout stays empty.
- README.md — deferred to P1.M5.T3 (Mode B doc sync).
- PRD.md / tasks.json / prd_snapshot.md — read-only / orchestrator-owned.

## 6. The §13 gate is preserved (the key safety argument)

`test "$(./skpp --path)" = "$PWD/skills"` captures **stdout only**. The new
`fmt.Fprintf(stderr, ...)` writes to stderr, which `$()` does not capture.
Therefore stdout remains exactly `<dir>\n` and the gate still passes. This is
the entire reason the fix goes to stderr and not stdout (decisions.md §D1).

## 7. Dead-code-elimination reversal (bonus validation)

Currently `strings ./skpp | grep "sibling of binary"` returns nothing because
the labels are dead code (src assigned to `_`, String() never called). After
the fix, `fmt.Fprintf(stderr, "(found via %s)\n", src)` references String(),
so the linker keeps the labels and `strings ./skpp | grep -c "found via"` is
>= 1. Useful as a build-level smoke test.

## 8. Build/test commands (verified available in repo)

- `go build ./...` — compiles
- `go test ./...` — runs all tests (skillsdir + main + others)
- `go test ./ -run 'TestRunPath' -v` — just the --path tests
- `go vet ./...` — static checks
- `gofmt -l .` — formatting (expect empty output = clean)
- `./skpp --path` / `./skpp -p` — manual (repo root → sibling-of-binary rule)
