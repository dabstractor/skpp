# PRP — P1.M1.T2.S1: `internal/skillsdir` — Source type + env-var resolution (§8 rule 1)

> **Subtask:** P1.M1.T2.S1 — the FIRST of three subtasks that build `internal/skillsdir`
> (the package that locates the on-disk `skills/` directory per PRD §8).
> **Scope boundary:** Creates the `internal/skillsdir/` package, the `Source` type
> (shared by ALL three rules and by `skpp --path`), and **rule 1 only** (`SKPP_SKILLS_DIR`).
> Does **NOT** implement rule 2 (sibling-of-binary — that is S2), rule 3
> (walk-up — that is S3), or the `Find()` combiner (S3). Pure standard library;
> no new dependencies; `go.mod`/`go.sum` untouched.

---

## Goal

**Feature Goal**: Create the `internal/skillsdir` Go package with (a) the `Source`
type + its three constants + a `String()` method that labels the winning §8 rule
for `skpp --path` reporting, and (b) an internal `findEnv()` helper that
implements §8 rule 1 — resolving `SKPP_SKILLS_DIR` to an absolute directory path
without ever calling `filepath.EvalSymlinks` on it — so that the `Find()`
combiner (built in S3) can call it as the first of three fall-through rules.

**Deliverable**: Two new files under the repo root:
1. `internal/skillsdir/skillsdir.go` — `package skillsdir`, the `Source` type +
   constants + `String()`, and the unexported `findEnv()` rule-1 helper.
2. `internal/skillsdir/skillsdir_test.go` — white-box table-driven tests
   (`package skillsdir`) covering `Source.String()` and every rule-1 branch.

**Success Definition**: `go build ./...` exits 0; `go vet ./internal/skillsdir/`
is clean; `gofmt -l internal/skillsdir/` prints nothing; and `go test
./internal/skillsdir/ -v` passes all cases — including the contract test that a
symlinked `SKPP_SKILLS_DIR` is returned **as the symlink** (not resolved to its
target). No `Find()`, `findSibling`, or `findWalkUp` exists yet.

---

## Why

- This is the foundation of the "hardest part" of skpp (PRD §18 build-order
  step 1, risk area #1 in `architecture/codebase_state.md`). Every downstream
  feature (`discover.Index`, `resolve.Resolve`, the whole CLI) calls
  `skillsdir.Find()` first; if location resolution is wrong, nothing works.
- The **`Source` type is shared infrastructure**: S2 (rule 2) and S3 (rule 3 +
  `Find()`) both depend on it existing with the exact constants and labels
  defined here. S3's `Find()` will chain `findEnv` → `findSibling` →
  `findWalkUp`. Establishing the type + the per-rule helper signature now lets
  S2/S3 slot in without rework.
- Rule 1 is pure stdlib (`os`, `path/filepath`) and fully unit-testable with
  `t.TempDir()` + `t.Setenv` — the lowest-risk place to lock the package's
  conventions (white-box tests, no mocking, no interfaces) before the harder
  symlink/walk-up logic arrives.
- The env-var rule is what makes `SKPP_SKILLS_DIR="$PWD/skills" ./skpp example`
  in PRD §13 acceptance pass, and what lets a user point at multiple stores by
  re-invoking with a different env (PRD §8.1).

---

## What

A single new internal package `internal/skillsdir` containing a type and one
rule. Concretely:

1. `type Source int` with constants `SourceEnv`, `SourceSibling`, `SourceWalkUp`
   (in that `iota` order — matches `architecture/go_architecture.md`).
2. `func (s Source) String() string` returning the labels `"SKPP_SKILLS_DIR"`,
   `"sibling of binary"`, `"ancestor of cwd"` (and `"unknown"` for any
   out-of-range value), so `Source` satisfies `fmt.Stringer` for `--path`.
3. `func findEnv() (dir string, src Source, found bool)` implementing §8 rule 1:
   - Read `SKPP_SKILLS_DIR` via `os.LookupEnv`.
   - Unset or empty → `found=false` (fall through; **no error**).
   - Set but `os.Stat` fails or is not a directory → `found=false` (fall
     through; **no error** — a bad env value never hard-errors, per contract).
   - Set and an existing directory → `filepath.Abs(val)` (make absolute + clean,
     **NOT** `EvalSymlinks`), return `(abs, SourceEnv, true)`.
4. White-box tests in `package skillsdir` proving every branch + the no-resolve-
   symlinks contract.

No exported `Find()` exists yet. No rule 2 / rule 3 helpers exist yet.

### Success Criteria

- [ ] `internal/skillsdir/skillsdir.go` exists with `package skillsdir`
- [ ] `type Source int` + constants `SourceEnv`/`SourceSibling`/`SourceWalkUp` defined (iota order)
- [ ] `Source.String()` returns the four exact labels above (incl. `"unknown"` default)
- [ ] `findEnv()` returns `(absDir, SourceEnv, true)` for a valid existing-dir env value
- [ ] `findEnv()` does **not** call `filepath.EvalSymlinks` on the env path (verified by symlink test)
- [ ] `findEnv()` returns `found=false` for: unset, empty, non-existent path, regular file
- [ ] `go build ./...` exits 0
- [ ] `go vet ./internal/skillsdir/` is clean
- [ ] `gofmt -l internal/skillsdir/` prints nothing
- [ ] `go test ./internal/skillsdir/ -v` — all cases pass
- [ ] NO `Find()`, `findSibling`, or `findWalkUp` present (those are S2/S3)
- [ ] `go.mod` / `go.sum` / `PRD.md` unchanged

---

## All Needed Context

### Context Completeness Check

_Pass: the exact source for both files is given verbatim in the Implementation
Blueprint, every stdlib behavior it relies on was empirically verified (see
`research/verified_facts.md`), and the package is pure stdlib with no existing
codebase patterns to discover (the repo is greenfield except for `go.mod` from
T1.S1). An implementer who knows Go but nothing about this repo can complete
this in one pass from this document alone._

### Documentation & References

```yaml
# MUST READ - the authoritative type shape + data flow
- file: plan/001_fcde63e5bb60/architecture/go_architecture.md
  why: "Defines the exact `internal/skillsdir` contract: Source type, the three
        constants, and the Find() signature (dir, src, err) that the rule
        helpers feed. Also the relTag/error/exit-code conventions S2/S3 must honor."
  critical: "The `Source int` + iota order (SourceEnv, SourceSibling, SourceWalkUp)
             and the Find() return signature are FIXED here. Match them exactly;
             S2/S3 import and extend this type."
  section: "Core types (contract between packages) → internal/skillsdir"

- file: plan/001_fcde63e5bb60/architecture/verified_symlink_resolution.md
  why: "Empirically proves WHY EvalSymlinks stays in the design (macOS needs it).
        Rules the contrast for rule 1: rule 2 uses os.Executable+EvalSymlinks;
        rule 1 must NOT EvalSymlinks the env path."
  critical: "Do NOT 'simplify' rule 1 by dropping or adding EvalSymlinks. Rule 1
             keeps the env path verbatim (filepath.Abs only); rule 2 keeps
             EvalSymlinks. The asymmetry is intentional and cross-platform."

- file: plan/001_fcde63e5bb60/architecture/external_deps.md
  why: "§4 lists the verified stdlib API table: os.Stat, filepath.Abs,
        filepath.EvalSymlinks, os.LookupEnv. Confirms these are the right tools
        in Go 1.26 and that there are NO runtime deps beyond stdlib + yaml.v3."
  section: "4. Go standard library APIs to use (verified available in Go 1.26)"

- file: plan/001_fcde63e5bb60/P1M1T2S1/research/verified_facts.md
  why: "Verbatim output of the empirical test proving: os.Stat follows symlinks
        (IsDir=true on a symlink-to-dir); filepath.Abs does NOT resolve symlinks
        (returns the link path); EvalSymlinks DOES resolve (contrast);
        LookupEnv unset vs empty; and the design rationale for the
        (dir, src, found) helper signature."
  critical: "The exact behaviors rule 1 depends on. Read before writing findEnv."

- file: plan/001_fcde63e5bb60/P1M1T1S1/PRP.md
  why: "The CONTRACT for the input: defines the go.mod this subtask builds on
        (module github.com/dabstractor/skpp, go 1.25). Internal packages import
        as github.com/dabstractor/skpp/internal/<pkg>."
  section: "Goal / What (the module path + the 'no internal/ yet' scope boundary)"

- file: PRD.md
  why: "§8 (locating the skills dir, priority order) is the authoritative spec
        for rule 1; §13 acceptance includes `SKPP_SKILLS_DIR=\"$PWD/skills\" ./skpp example`
        which this rule must make pass; §17 guardrails (no stdout-on-failure etc.)
        bound behavior. PRD.md is READ-ONLY — do not modify."
  critical: "§8 rule 1 wording: 'if set and points to an existing dir, use it ...
             Do NOT EvalSymlinks the env path (user points exactly where they want).'"
  section: "8. Locating the skills directory (priority order) — rule 1"

- url: https://pkg.go.dev/os#LookupEnv
  why: "os.LookupEnv returns (value, ok) — ok==false means unset. The idiomatic
        'is the var set?' check (vs os.Getenv which cannot distinguish unset from empty)."
- url: https://pkg.go.dev/os#Stat
  why: "os.Stat follows symlinks and returns FileInfo.IsDir(). The existence+dir check."
- url: https://pkg.go.dev/path/filepath#Abs
  why: "filepath.Abs makes a path absolute (joining cwd if relative) and runs
        filepath.Clean. It does NOT resolve symlinks — exactly what rule 1 needs."
- url: https://pkg.go.dev/testing#T.Setenv
  why: "t.Setenv sets an env var for the duration of a test and restores it after.
        NOTE: it cannot unset; see the unsetEnvVar helper in the test file."
```

### Current Codebase tree (after P1.M1.T1.S1; before this subtask)

```bash
$ cd /home/dustin/projects/skpp && ls -A
.git/
.gitignore      # PRD §16 content (from T1.S1)
LICENSE         # MIT (from T1.S1)
PRD.md          # READ-ONLY
go.mod          # module github.com/dabstractor/skpp, go 1.25, yaml.v3 // indirect (from T1.S1)
go.sum          # yaml.v3 checksums (from T1.S1)
plan/           # planning artifacts (untracked)
.pi-subagents/
```

There is **no `internal/` directory and no source code yet.** This subtask
creates `internal/skillsdir/`.

### Desired Codebase tree with files to be added

```bash
skpp/
├── ... (go.mod, go.sum, .gitignore, LICENSE, PRD.md — unchanged)
└── internal/
    └── skillsdir/
        ├── skillsdir.go        # CREATE — package decl, Source type + String(), findEnv() [rule 1]
        └── skillsdir_test.go   # CREATE — white-box tests for Source.String() + findEnv()
```

| File | Responsibility | Consumed by |
|---|---|---|
| `internal/skillsdir/skillsdir.go` | Defines the `Source` type/labels + rule-1 env resolver | S2 (adds `findSibling`), S3 (adds `findWalkUp` + `Find()`), T3 (`--path` prints `src`) |
| `internal/skillsdir/skillsdir_test.go` | Proves rule 1 + `Source.String()` | `go test` (validation gate) |

### Known Gotchas of our codebase & Go stdlib

```go
// GOTCHA #1 — Rule 1 must NOT call filepath.EvalSymlinks on the env path.
// VERIFIED (research/verified_facts.md §2): os.Stat follows symlinks (so a
// symlinked SKPP_SKILLS_DIR is detected as a dir WITHOUT EvalSymlinks), and
// filepath.Abs does NOT resolve symlinks (returns the link path verbatim).
// EvalSymlinks is rule 2's tool (sibling-of-binary, macOS needs it) — NEVER
// rule 1's. The asymmetry is the contract ("user points exactly where they
// want"). Enforced by TestFindEnvDoesNotResolveSymlinks.
//
//   RIGHT:  abs, _ := filepath.Abs(val)      // link path preserved
//   WRONG:  abs, _ := filepath.EvalSymlinks(val)  // resolves to target — breaks contract

// GOTCHA #2 — A bad/empty/unset env value must NOT error; it must fall through.
// The contract is explicit: "do not hard-error on a bad env value — let later
// rules try." So findEnv returns (..., found=false), NEVER an error. Only when
// ALL THREE rules miss does Find() (S3) return an error. Do not add an error
// return to findEnv "for safety" — it would break the fall-through design and
// diverge from S2/S3's (dir, src, found) shape.

// GOTCHA #3 — os.Stat follows symlinks; that is DESIRED here. os.Lstat would
// NOT follow (it'd report the symlink itself, IsDir==false) and would wrongly
// reject a symlinked store. Use os.Stat, not os.Lstat, in rule 1.

// GOTCHA #4 — Use os.LookupEnv, not os.Getenv, to express "is it set?".
// os.Getenv returns "" for both unset and explicitly-empty; os.LookupEnv
// returns (val, ok) where ok distinguishes them. Behavior is equivalent for
// rule 1 (empty→Stat("")→fall through), but LookupEnv documents intent and
// makes the unset test (TestFindEnvUnset) meaningful.

// GOTCHA #5 — t.Setenv cannot UNSET a var. The "var unset" test case needs a
// manual os.Unsetenv + tb.Cleanup restore (the unsetEnvVar helper in the test
// file). And because every rule-1 test mutates process env, NONE of these
// tests may call t.Parallel() (t.Setenv enforces this; document it inline).

// GOTCHA #6 — findEnv will appear "unused" until S3 wires it into Find().
// Go's `go build` and `go vet` do NOT flag unused package-level functions, so
// the package compiles cleanly now. Do NOT delete findEnv or stub a Find()
// to "use" it — Find() is S3's deliverable and must not be created here.

// GOTCHA #7 — This package is under internal/, so it is importable only from
// within github.com/dabstractor/skpp. main.go (T3) imports skillsdir via
// "github.com/dabstractor/skpp/internal/skillsdir". That import is added in
// T3, not here — this subtask just makes the package exist and compile.

// GOTCHA #8 — No new imports beyond "os" and "path/filepath". Do NOT import
// "fmt" just for String() — a switch returning string literals needs no fmt.
// An unused import is a compile error in Go; keep the import block minimal.
```

---

## Implementation Blueprint

### Data model — the `Source` type

The only data model in this subtask. It is the shared "which rule won?" tag
reported by `skpp --path` and threaded through every resolver.

```go
// Source identifies which §8 rule located the skills directory.
type Source int

const (
    SourceEnv     Source = iota // SKPP_SKILLS_DIR env var (rule 1 — this subtask)
    SourceSibling               // sibling of the running binary (rule 2 — S2)
    SourceWalkUp                // ancestor of cwd (rule 3 — S3)
)
```

The per-rule helper signature (this subtask's design choice, documented in
`research/verified_facts.md §4`; S2/S3 are expected to mirror it):

```go
// findEnv implements §8 rule 1. Returns found=true only when SKPP_SKILLS_DIR is
// set and names an existing directory; the dir is made absolute via filepath.Abs
// (symlinks intentionally NOT resolved). src is only meaningful when found==true.
func findEnv() (dir string, src Source, found bool)
```

### File 1 — `internal/skillsdir/skillsdir.go` (exact contents)

```go
// Package skillsdir locates the on-disk skills/ directory for skpp.
//
// It implements the PRD §8 priority order:
//
//  1. SKPP_SKILLS_DIR env var — if set and an existing dir, use it as-is.
//  2. Sibling of the running binary (symlink-aware via os.Executable + EvalSymlinks).
//  3. Walk up from the current working directory.
//
// The public entry point is Find() (added in P1.M1.T2.S3), which calls the
// per-rule helpers in order and returns the first hit. Each per-rule helper
// returns (dir string, src Source, found bool) where found is true only when
// that rule produced a usable absolute directory; on found==false the src value
// is meaningless and the caller falls through to the next rule.
package skillsdir

import (
	"os"
	"path/filepath"
)

// Source identifies which §8 rule located the skills directory. It is reported
// by `skpp --path` so users can tell how the dir was found.
type Source int

const (
	// SourceEnv means SKPP_SKILLS_DIR was set and pointed at an existing dir.
	SourceEnv Source = iota
	// SourceSibling means the skills dir was found next to the running binary.
	SourceSibling
	// SourceWalkUp means the skills dir was found by walking up from cwd.
	SourceWalkUp
)

// String returns a human-readable label for the rule that won, used by
// `skpp --path` reporting. Satisfies fmt.Stringer.
func (s Source) String() string {
	switch s {
	case SourceEnv:
		return "SKPP_SKILLS_DIR"
	case SourceSibling:
		return "sibling of binary"
	case SourceWalkUp:
		return "ancestor of cwd"
	default:
		return "unknown"
	}
}

// envVar is the environment variable consulted by rule 1. It is a package
// constant (not a parameter): the contract is "mock/replace nothing" — tests
// drive it via t.Setenv / os.Unsetenv, never via injection.
const envVar = "SKPP_SKILLS_DIR"

// findEnv implements PRD §8 rule 1.
//
// It reads SKPP_SKILLS_DIR; if the value names an existing directory it returns
// that directory as an absolute path with src=SourceEnv and found=true. The env
// path is NOT passed through filepath.EvalSymlinks: the user points exactly
// where they want (a symlink is preserved verbatim, only made absolute/clean
// via filepath.Abs). If the variable is unset, empty, or does not name an
// existing directory, it returns found=false with src's zero value so Find()
// can fall through to rule 2 — a bad env value never hard-errors.
func findEnv() (dir string, src Source, found bool) {
	val, ok := os.LookupEnv(envVar)
	if !ok || val == "" {
		return "", 0, false
	}
	info, err := os.Stat(val)
	if err != nil || !info.IsDir() {
		return "", 0, false // not an existing dir -> let the next rule try
	}
	abs, err := filepath.Abs(val)
	if err != nil {
		return "", 0, false // cwd unresolvable -> let the next rule try
	}
	return abs, SourceEnv, true
}
```

### File 2 — `internal/skillsdir/skillsdir_test.go` (exact contents)

White-box test (`package skillsdir`) so the unexported `findEnv` is callable.

```go
package skillsdir

import (
	"os"
	"path/filepath"
	"testing"
)

// unsetEnvVar removes envVar for the duration of the test and restores the
// prior state on cleanup. Needed because t.Setenv can only set, never unset.
// Do NOT call t.Parallel() in any test that uses this or t.Setenv.
func unsetEnvVar(tb testing.TB) {
	tb.Helper()
	prev, had := os.LookupEnv(envVar)
	if err := os.Unsetenv(envVar); err != nil {
		tb.Fatalf("unsetenv %s: %v", envVar, err)
	}
	tb.Cleanup(func() {
		if had {
			_ = os.Setenv(envVar, prev)
		} else {
			_ = os.Unsetenv(envVar)
		}
	})
}

func TestSourceString(t *testing.T) {
	cases := []struct {
		src  Source
		want string
	}{
		{SourceEnv, "SKPP_SKILLS_DIR"},
		{SourceSibling, "sibling of binary"},
		{SourceWalkUp, "ancestor of cwd"},
		{Source(-1), "unknown"}, // out-of-range -> default
		{Source(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.src.String(); got != c.want {
			t.Errorf("Source(%d).String() = %q, want %q", c.src, got, c.want)
		}
	}
}

// Rule 1: env unset -> not found (fall through, no error).
func TestFindEnvUnset(t *testing.T) {
	unsetEnvVar(t)
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() env unset: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to "" -> not found (treated as no usable dir).
func TestFindEnvEmpty(t *testing.T) {
	t.Setenv(envVar, "")
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() env empty: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to an existing directory (absolute temp dir) -> found, abs path, SourceEnv.
func TestFindEnvExistingDir(t *testing.T) {
	dir := t.TempDir() // already absolute + clean
	t.Setenv(envVar, dir)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() existing dir: found=false; want true")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() existing dir: src=%v; want SourceEnv", src)
	}
	if want := filepath.Clean(dir); got != want {
		t.Errorf("findEnv() existing dir: dir=%q; want %q", got, want)
	}
}

// Rule 1: env set to a path that does not exist -> not found (fall through).
func TestFindEnvNonexistent(t *testing.T) {
	t.Setenv(envVar, filepath.Join(t.TempDir(), "does-not-exist"))
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() nonexistent: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to a regular file (not a directory) -> not found (fall through).
func TestFindEnvRegularFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envVar, f)
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() regular file: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to a RELATIVE existing dir (".") -> found, absolutized via filepath.Abs.
// Proves the filepath.Abs (relative->absolute) path without chdir or cwd pollution.
func TestFindEnvRelativePathAbsolutized(t *testing.T) {
	t.Setenv(envVar, ".") // "." always exists and is a dir (the test's cwd)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() relative '.': found=false; want true")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() relative '.': src=%v; want SourceEnv", src)
	}
	want, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findEnv() relative '.': dir=%q; want absolutized %q", got, want)
	}
}

// Rule 1 CONTRACT: the env path must NOT be passed through EvalSymlinks.
// If SKPP_SKILLS_DIR points at a symlink-to-a-dir, findEnv must return the
// symlink path (made absolute/clean), NOT the resolved target. The user points
// exactly where they want.
func TestFindEnvDoesNotResolveSymlinks(t *testing.T) {
	realDir := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "link-to-skills")
	if err := os.Symlink(realDir, link); err != nil {
		t.Skipf("symlinks not supported on this platform: %v", err)
	}
	t.Setenv(envVar, link)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() symlink-to-dir: found=false; want true (os.Stat follows the symlink)")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() symlink-to-dir: src=%v; want SourceEnv", src)
	}
	if got == realDir {
		t.Errorf("findEnv() symlink-to-dir: dir=%q == resolved target; must NOT EvalSymlinks the env path", got)
	}
	want, err := filepath.Abs(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findEnv() symlink-to-dir: dir=%q; want symlink path (absolutized) %q", got, want)
	}
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: PRECONDITION — confirm the T1.S1 module exists
  - COMMAND: cd /home/dustin/projects/skpp && test -f go.mod && head -1 go.mod
  - EXPECT: prints "module github.com/dabstractor/skpp"
  - WHY: this subtask imports as github.com/dabstractor/skpp/internal/skillsdir.
         If go.mod is missing, P1.M1.T1.S1 has not run yet — stop and let it land.

Task 1: CREATE internal/skillsdir/skillsdir.go
  - WRITE: the exact file contents above (package doc comment, package decl,
           imports "os" + "path/filepath", Source type + 3 constants + String(),
           envVar const, findEnv())
  - NAMING: file = skillsdir.go (mirrors the package); type Source; constants
            SourceEnv/SourceSibling/SourceWalkUp; helper findEnv (unexported)
  - PLACEMENT: internal/skillsdir/ (internal/ created implicitly by the write)
  - GOTCHA: do NOT import "fmt" (String() uses a switch, no Sprintf needed);
            do NOT add Find()/findSibling/findWalkUp (S2/S3); do NOT call
            EvalSymlinks anywhere in this file

Task 2: CREATE internal/skillsdir/skillsdir_test.go
  - WRITE: the exact file contents above (package skillsdir white-box)
  - COVERAGE: Source.String() (5 cases incl. out-of-range) + findEnv() branches:
             unset, empty, existing dir, nonexistent, regular file, relative
             absolutization, symlink-not-resolved (the contract test)
  - NAMING: TestSourceString, TestFindEnv{Unset,Empty,ExistingDir,Nonexistent,
            RegularFile,RelativePathAbsolutized,DoesNotResolveSymlinks}
  - GOTCHA: use the unsetEnvVar helper for the unset case; no t.Parallel() in
            any env-mutating test

Task 3: FORMAT + VET + BUILD + TEST (the validation gates — run in order)
  - COMMAND: gofmt -w internal/skillsdir/        # format in place
  - COMMAND: gofmt -l internal/skillsdir/        # MUST print nothing
  - COMMAND: go vet ./internal/skillsdir/        # MUST be clean
  - COMMAND: go build ./...                       # exit 0 (package now non-empty; no "matched no packages" warning expected)
  - COMMAND: go test ./internal/skillsdir/ -v     # all 8 test fns PASS
  - EXPECT: zero errors, zero vet findings, gofmt silent

Task 4: SCOPE BOUNDARY CHECK
  - COMMAND: the Level 3 block in Validation Loop below
  - EXPECT: no Find()/findSibling/findWalkUp; PRD.md/go.mod/go.sum unchanged
```

### Implementation Patterns & Key Details

```go
// PATTERN: the per-rule helper shape (locked here; S2/S3 mirror it).
//   func findX() (dir string, src Source, found bool)
// - found==true  -> dir is absolute, src names the winning Source; Find() returns it.
// - found==false -> this rule did not win; src's zero value is ignored; Find()
//                   falls through to the next rule. NO error return — only Find()
//                   (S3) produces an error, and only when all three rules miss.

// PATTERN: rule 1's absolutize-without-resolve step.
//   abs, err := filepath.Abs(val)   // make absolute + lexical clean; symlinks PRESERVED
// NEVER:    filepath.EvalSymlinks(val)   // resolves symlinks — that is rule 2's job

// PATTERN: the "fall through on any non-dir" check (os.Stat, NOT os.Lstat).
//   info, err := os.Stat(val)
//   if err != nil || !info.IsDir() { return "", 0, false }
// os.Stat follows symlinks (so a symlinked store is accepted); os.Lstat would
// wrongly reject symlinks. Use Stat.

// PATTERN: white-box testing of unexported helpers.
//   // internal/skillsdir/skillsdir_test.go
//   package skillsdir   // SAME package -> can call findEnv() directly
// Drive env state with t.Setenv (set) and the unsetEnvVar helper (unset);
// never t.Parallel() in env-mutating tests.
```

### Integration Points

```yaml
PACKAGE BOUNDARIES:
  - import path: "github.com/dabstractor/skpp/internal/skillsdir"
  - exported (for S2/S3/T3 to use): Source, SourceEnv, SourceSibling, SourceWalkUp, (Source).String
  - unexported (internal to the package): findEnv, envVar
  - added LATER (NOT this subtask): findSibling (S2), findWalkUp + Find() + ErrNotFound (S3)

DOWNSTREAM CONSUMERS (what relies on this subtask's output):
  - P1.M1.T2.S2: extends skillsdir.go with findSibling() using filepath.EvalSymlinks
    (sibling-of-binary). Reuses Source + SourceSibling (defined here).
  - P1.M1.T2.S3: adds findWalkUp() + the public Find() that chains
    findEnv -> findSibling -> findWalkUp -> ErrNotFound, returning (dir, src, err).
  - P1.M1.T3 (main.go --path): prints src.String() to report which rule won.

NO CHANGES TO:
  - go.mod / go.sum (no new deps; pure stdlib)
  - PRD.md (read-only)
  - any other package (discover/resolve/ui/main are later subtasks)
```

---

## Validation Loop

### Level 1: Format, vet, build (immediate, per file)

```bash
cd /home/dustin/projects/skpp

# Format in place, then confirm nothing is left unformatted (silent == pass)
gofmt -w internal/skillsdir/
test -z "$(gofmt -l internal/skillsdir/)" || { echo "FAIL: gofmt found unformatted files"; gofmt -d internal/skillsdir/; exit 1; }
echo "gofmt OK"

# Vet the new package
go vet ./internal/skillsdir/ || { echo "FAIL: go vet"; exit 1; }
echo "go vet OK"

# Build the whole module (the package is now non-empty, so no "matched no
# packages" warning is expected)
go build ./... || { echo "FAIL: go build ./..."; exit 1; }
echo "go build ./... OK"
```

### Level 2: Unit tests (component validation)

```bash
cd /home/dustin/projects/skpp

# Run the skillsdir tests verbosely — all 8 functions must PASS
go test ./internal/skillsdir/ -v

# Explicit assertions the run must satisfy:
go test ./internal/skillsdir/ -run 'TestSourceString|TestFindEnv' -v || { echo "FAIL: skillsdir tests"; exit 1; }

# Confirm the symlink-contract test actually ran (not skipped) on this platform.
# Its log line "TestFindEnvDoesNotResolveSymlinks" must appear with --- PASS
# (or --- SKIP only if the host cannot create symlinks; on linux/amd64 it runs).
go test ./internal/skillsdir/ -run TestFindEnvDoesNotResolveSymlinks -v | grep -E '--- (PASS|SKIP):.*TestFindEnvDoesNotResolveSymlinks' \
  || { echo "FAIL: symlink-contract test did not run"; exit 1; }

# Test the whole module too (only skillsdir has tests right now; should be green)
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Scope-boundary & integration check

```bash
cd /home/dustin/projects/skpp

# MUST NOT have created Find() / findSibling / findWalkUp yet (S2/S3 deliverables)
! grep -nE 'func (Find|findSibling|findWalkUp)\b' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: found S2/S3 symbols in skillsdir.go — out of scope"; exit 1; }

# MUST NOT have touched go.mod / go.sum / PRD.md
git diff --quiet go.mod   || { echo "FAIL: go.mod changed (should be untouched)"; exit 1; }
git diff --quiet go.sum   || { echo "FAIL: go.sum changed (should be untouched)"; exit 1; }
git diff --quiet PRD.md   || { echo "FAIL: PRD.md changed (read-only)"; exit 1; }

# MUST NOT have created main.go or other packages (later subtasks)
test ! -e main.go                 || { echo "FAIL: main.go must not exist (T3)"; exit 1; }
test ! -d internal/discover       || { echo "FAIL: discover/ must not exist (M2)"; exit 1; }
test ! -d internal/resolve        || { echo "FAIL: resolve/ must not exist (M3)"; exit 1; }
test ! -d internal/ui             || { echo "FAIL: ui/ must not exist (M2)"; exit 1; }

# The two new files exist exactly where expected
test -f internal/skillsdir/skillsdir.go      || { echo "FAIL: skillsdir.go missing"; exit 1; }
test -f internal/skillsdir/skillsdir_test.go || { echo "FAIL: skillsdir_test.go missing"; exit 1; }

# Source.String() labels are exactly the contracted ones (grep check)
grep -q 'return "SKPP_SKILLS_DIR"'  internal/skillsdir/skillsdir.go || { echo "FAIL: SourceEnv label"; exit 1; }
grep -q 'return "sibling of binary"' internal/skillsdir/skillsdir.go || { echo "FAIL: SourceSibling label"; exit 1; }
grep -q 'return "ancestor of cwd"'  internal/skillsdir/skillsdir.go || { echo "FAIL: SourceWalkUp label"; exit 1; }

# findEnv must NOT use EvalSymlinks (the rule-1 contract)
! grep -n 'EvalSymlinks' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: rule 1 must not EvalSymlinks the env path"; exit 1; }
echo "Level 3 PASS (scope + contract respected)"
```

### Level 4: Downstream-readiness smoke test

Prove the package is importable and the `Source` type behaves as S3's `Find()`
and T3's `--path` will rely on — without depending on S2/S3 existing yet.

```bash
cd /home/dustin/projects/skpp

# (a) Source implements fmt.Stringer and prints the contracted labels.
mkdir -p /tmp/skpp-src-check && cat > /tmp/skpp-src-check/main.go <<'EOF'
package main
import (
	"fmt"
	"github.com/dabstractor/skpp/internal/skillsdir"
)
func main() {
	var s skillsdir.Source = skillsdir.SourceEnv
	fmt.Println(s.String()) // SKPP_SKILLS_DIR
}
EOF
cat > /tmp/skpp-src-check/go.mod <<EOF
module skpp-src-check
go 1.25
require github.com/dabstractor/skpp v0.0.0
replace github.com/dabstractor/skpp => $(pwd)
EOF
( cd /tmp/skpp-src-check && go run . ) | grep -qx 'SKPP_SKILLS_DIR' \
  || { echo "FAIL: Source.Stringer not importable/usable downstream"; rm -rf /tmp/skpp-src-check; exit 1; }
rm -rf /tmp/skpp-src-check

# (b) Live behavior proof: run findEnv via a tiny test in-package is already
# covered by go test; here we confirm the env override end-to-end style that
# PRD §13 will eventually use (this is a sanity check of the rule, not the CLI).
echo "Level 4 PASS (Source usable by downstream code)"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` silent, `go vet ./internal/skillsdir/` clean, `go build ./...` exit 0
- [ ] Level 2 PASS — `go test ./internal/skillsdir/ -v`: all 8 functions pass; symlink-contract test ran (not skipped) on linux/amd64
- [ ] Level 3 PASS — no S2/S3 symbols; go.mod/go.sum/PRD.md unchanged; rule 1 has no EvalSymlinks
- [ ] Level 4 PASS — `Source.String()` usable from a downstream `internal/skillsdir` import

### Feature Validation
- [ ] `Source.String()` returns `"SKPP_SKILLS_DIR"` / `"sibling of binary"` / `"ancestor of cwd"` / `"unknown"`
- [ ] `findEnv()` returns `(absDir, SourceEnv, true)` when `SKPP_SKILLS_DIR` is an existing dir
- [ ] `findEnv()` returns `found=false` (no error) for unset / empty / non-existent / non-dir
- [ ] `findEnv()` preserves a symlinked env path (does NOT resolve it) — contract enforced by test
- [ ] `findEnv()` absolutizes a relative env path via `filepath.Abs`

### Code Quality / Convention Validation
- [ ] File placement matches the desired tree (`internal/skillsdir/`)
- [ ] White-box test (`package skillsdir`) so unexported `findEnv` is testable
- [ ] Imports limited to `os` + `path/filepath` (no unused `fmt`)
- [ ] No `t.Parallel()` in env-mutating tests (documented)
- [ ] Package doc comment explains the §8 order + the `(dir, src, found)` helper contract for S2/S3

### Scope Discipline
- [ ] Did NOT create `Find()`, `findSibling`, or `findWalkUp` (S2/S3)
- [ ] Did NOT create `main.go`, `internal/discover`, `internal/resolve`, `internal/ui` (later milestones)
- [ ] Did NOT modify `go.mod` / `go.sum` (no new deps)
- [ ] Did NOT modify `PRD.md` (read-only) or any `tasks.json` (orchestrator-owned)

---

## Anti-Patterns to Avoid

- ❌ **Don't call `filepath.EvalSymlinks` on the env path.** That resolves the symlink to its target and breaks the "user points exactly where they want" contract (PRD §8.1). Use `filepath.Abs` only. Rule 2 uses EvalSymlinks; rule 1 does not. (Verified: `research/verified_facts.md §2`.)
- ❌ **Don't return an error from `findEnv`.** A bad/empty/unset env value must fall through to rule 2, not abort. Only `Find()` (S3) errors, and only when all three rules miss. Adding an error return breaks the `(dir, src, found)` shape S2/S3 expect.
- ❌ **Don't use `os.Lstat`.** It does not follow symlinks (`IsDir()==false` on a symlink-to-dir), wrongly rejecting a symlinked store. Use `os.Stat`.
- ❌ **Don't create `Find()` to "use" `findEnv`.** `Find()` is S3's deliverable. `go build`/`go vet` do not flag unused package-level funcs; leave `findEnv` present-but-uncalled until S3 wires it in.
- ❌ **Don't stub `findSibling`/`findWalkUp`.** Those are S2/S3. Keep this file to rule 1 + the shared type.
- ❌ **Don't import `fmt` for `String()`.** A `switch` returning string literals needs no `fmt`; an unused import is a compile error.
- ❌ **Don't call `t.Parallel()` in the env tests.** They mutate process env via `t.Setenv`/`os.Unsetenv`; parallelism would race. `t.Setenv` itself forbids it.
- ❌ **Don't parameterize the env var name "for testability."** The contract is explicit: "mock/replace nothing." Drive tests with `t.Setenv` + the `unsetEnvVar` helper, not dependency injection.
- ❌ **Don't hand-edit `go.mod`/`go.sum`.** This subtask adds no dependencies; leave them exactly as T1.S1 produced them.

---

## Confidence Score

**10/10** — This is a small, pure-stdlib package whose exact source (both files) is
given verbatim, and every stdlib behavior it depends on (`os.Stat` follows
symlinks; `filepath.Abs` does not; `os.LookupEnv` semantics; `t.Setenv` cannot
unset) was **empirically verified** in the target Go 1.26.4 environment (see
`research/verified_facts.md`). The `Source` type + labels are taken directly
from the work-item contract and `architecture/go_architecture.md`; the
`(dir, src, found)` helper signature is documented for the S2/S3 consumers.
There are no external dependencies, no concurrency, and no I/O beyond env + stat.
The only residual (non-)risk — an unused-symbol linter flagging `findEnv` — is
explicitly excluded by the validation gates (`go build`/`go vet`/`go test`/`gofmt`,
no staticcheck) and by an inline anti-pattern.
