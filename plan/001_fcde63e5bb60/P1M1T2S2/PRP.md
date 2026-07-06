# PRP — P1.M1.T2.S2: `internal/skillsdir` — sibling-of-binary resolution (§8 rule 2, symlink-aware)

> **Subtask:** P1.M1.T2.S2 — the SECOND of three subtasks that build
> `internal/skillsdir`. **Scope:** add `findSibling()` (rule 2) to the package
> created by S1. Does **NOT** add `findWalkUp` (rule 3 — S3) or the public
> `Find()` combiner (S3). Pure stdlib; no new deps; `go.mod`/`go.sum` untouched.
>
> **Status of S1:** ALREADY IMPLEMENTED on disk (`internal/skillsdir/skillsdir.go`
> + `skillsdir_test.go` exist and pass). This PRP **MODIFIES** those two files;
> it does not recreate them.

---

## Goal

**Feature Goal**: Add `findSibling()` — the §8 rule-2 helper — to
`internal/skillsdir`, implementing the symlink-aware sibling-of-binary
resolution: `os.Executable()` → `filepath.EvalSymlinks()` (with a verbatim
fallback to `exe` on error) → `filepath.Dir()` → `Join(…, "skills")` →
`os.Stat` existing dir ⇒ return `(candidate, SourceSibling, true)`. This is the
rule that makes a **symlink install** work (`~/.local/bin/skpp →
~/projects/skpp/skpp` resolves back to the repo's `skills/`).

**Deliverable**: Two existing files are **modified** (no new files):
1. `internal/skillsdir/skillsdir.go` — append `findSibling()` (the rule-2 entry,
   `(dir, src, found)` shape locked by S1) and a private `resolveSiblingFromExe()`
   (the unit-testable core; see Implementation Blueprint for why it is extracted).
2. `internal/skillsdir/skillsdir_test.go` — append white-box tests covering the
   symlink-install scenario, the EvalSymlinks-error fallback, the direct-binary
   case, and the two miss cases.

**Success Definition**: `go build ./...` exits 0; `go vet ./internal/skillsdir/`
clean; `gofmt -l internal/skillsdir/` silent; `go test ./internal/skillsdir/ -v`
passes all cases — including the contract test that a symlink to a binary in a
**different** temp dir resolves back to the real binary's `skills/` dir (mirrors
`architecture/verified_symlink_resolution.md`). `go.mod`/`go.sum`/`PRD.md`
unchanged; no `Find()`/`findWalkUp` present (those are S3).

---

## Why

- Rule 2 is the **most failure-prone** part of skills-dir location (PRD §8 calls
  it out; `architecture/codebase_state.md` ranks it risk area #1). It is what the
  PRD §13 acceptance suite exercises twice: `test "$(./skpp --path)" = "$PWD/skills"`
  (sibling-of-binary rule wins for a binary built at the repo root) and the
  explicit `/tmp/skpp-bin/skpp example` cross-dir symlink test.
- It is the rule that makes **install.sh's symlink-not-copy decision** (PRD §12.1)
  correct: a symlink install keeps one source of truth and lets `os.Executable()`
  resolve back to the repo. Without rule 2, copying the binary to `~/.local/bin`
  would silently break discovery.
- The rule's correctness is **cross-platform load-bearing**: on Linux
  `os.Executable()` already returns the real path via `/proc/self/exe` (so
  `EvalSymlinks` is redundant-but-harmless), but on macOS `EvalSymlinks` is
  REQUIRED. Empirically verified in
  `architecture/verified_symlink_resolution.md`. Implementing BOTH calls exactly
  (with the documented fallback) is the contract; "simplifying" by dropping
  `EvalSymlinks` breaks macOS.

---

## What

Append two functions to `internal/skillsdir/skillsdir.go` (S1's file), matching
the per-rule helper shape S1 locked (`func findX() (dir string, src Source, found
bool)`):

1. `func findSibling() (dir string, src Source, found bool)` — the rule-2 entry:
   - `exe, err := os.Executable()`; if `err != nil`, return `("", 0, false)` (skip rule).
   - Delegate to `resolveSiblingFromExe(exe)`.
   - On `found==true` return `(candidate, SourceSibling, true)`; else `("", 0, false)`.
2. `func resolveSiblingFromExe(exe string) (dir string, found bool)` — the
   symlink-aware core (extracted so it is testable with arbitrary exe paths;
   `os.Executable()` cannot be controlled in a test — see Known Gotchas #1):
   - `real, err := filepath.EvalSymlinks(exe)`; if `err != nil`, `real = exe` (fallback).
   - `repoDir := filepath.Dir(real)`.
   - `candidate := filepath.Join(repoDir, "skills")`.
   - `info, err := os.Stat(candidate)`; if `err != nil` or `!info.IsDir()`, return `("", false)`.
   - Else return `(candidate, true)`.

Append tests to `internal/skillsdir/skillsdir_test.go` proving: the symlink-cross-dir
win, the direct-binary win, the EvalSymlinks-error fallback, no-skills-dir miss,
skills-is-a-file miss, and a smoke test of the real `findSibling()` (returns
`found=false` because the test binary runs from a temp build dir with no sibling
`skills/`).

### Success Criteria

- [ ] `findSibling()` exists with the exact `(dir string, src Source, found bool)` signature
- [ ] `resolveSiblingFromExe(exe string) (dir string, found bool)` exists (private)
- [ ] `findSibling()` skips the rule (returns `found=false`) when `os.Executable()` errors
- [ ] `EvalSymlinks` is called and its error falls back to `exe` (not to skipping the rule)
- [ ] Returns `(candidate, SourceSibling, true)` when `<binaryDir>/skills` is an existing dir
- [ ] Returns `found=false` when `skills` sibling is absent OR is a regular file
- [ ] Cross-dir symlink test passes (symlink in dir B → resolves to binary's dir A `skills/`)
- [ ] `go build ./...` exits 0; `go vet ./internal/skillsdir/` clean; `gofmt -l` silent
- [ ] `go test ./internal/skillsdir/ -v` — all new + existing tests pass
- [ ] NO `Find()` or `findWalkUp` present (S3); `go.mod`/`go.sum`/`PRD.md` unchanged

---

## All Needed Context

### Context Completeness Check

_Pass: the exact source for both appended functions is given verbatim in the
Implementation Blueprint; every stdlib behavior they rely on (`os.Executable()`
path in tests, `EvalSymlinks` symlink resolution + error fallback, `os.Stat`
dir-check, `filepath.Dir`/`Join`) was **empirically verified** in the target Go
1.26.4 environment (see `research/verified_facts.md`). S1 is already on disk and
read in full — the exact insertion points are known. An implementer who knows Go
but nothing about this repo can complete this in one pass from this document._

### Documentation & References

```yaml
# MUST READ — the verified symlink behavior (the crux of rule 2)
- file: plan/001_fcde63e5bb60/architecture/verified_symlink_resolution.md
  why: "Empirically proves WHY EvalSymlinks stays in the design (macOS needs it;
        redundant-but-harmless on Linux). States the exact two-call sequence and
        the install scenario it enables."
  critical: "On Linux os.Executable() already returns the REAL path via
             /proc/self/exe; EvalSymlinks is required on macOS. Implement BOTH
             calls exactly; do NOT drop EvalSymlinks."

# MUST READ — this subtask's own empirical verification (esp. the testing strategy)
- file: plan/001_fcde63e5bb60/P1M1T2S2/research/verified_facts.md
  why: "Proves: (a) os.Executable() in go test returns a /tmp/go-buildXXX dir with
        NO sibling skills/ — so findSibling() must delegate to a testable core;
        (b) the symlink-cross-dir win; (c) the EvalSymlinks-error fallback; (d) a
        regular file suffices as a fake binary. Every test case in this PRP was
        pre-run against this exact logic."
  critical: "The reason for extracting resolveSiblingFromExe(exe). Without it,
             the symlink behavior (the whole point of rule 2) is untestable."

# CONTRACT — the package this subtask modifies (read first, do not recreate)
- file: internal/skillsdir/skillsdir.go
  why: "S1's delivered file. Contains the Source type, SourceSibling constant,
        String() label 'sibling of binary', and findEnv(). This subtask APPENDS
        findSibling + resolveSiblingFromExe to it and adds no new imports."
  pattern: "findEnv() is the template: (dir string, src Source, found bool); returns
            ('', 0, false) to fall through; ('<abs>', SourceX, true) to win."
  gotcha: "Imports already declared: 'os' + 'path/filepath' — both are exactly what
           rule 2 needs, so NO import block change is required. Do not add 'fmt'."

- file: internal/skillsdir/skillsdir_test.go
  why: "S1's white-box test file (package skillsdir). Contains unsetEnvVar(tb) and
        8 findEnv/Source tests. This subtask APPENDS findSibling tests; do not touch
        the existing tests."
  pattern: "t.TempDir() for scratch dirs; os.Symlink wrapped in t.Skipf on error;
            filepath.Join for paths; plain t.Errorf/t.Fatalf assertions (no testify)."

# CONTRACT — the locked helper signature + Source type S2 must conform to
- file: plan/001_fcde63e5bb60/P1M1T2S1/research/verified_facts.md
  why: "§4 locks the per-rule helper signature func findX() (dir string, src
        Source, found bool) — no error return; found is the 'did this rule win?'
        signal; src self-describes the winner so Find() (S3) stays dumb."
  section: "4. Design decision — per-rule helper signature"

# ARCHITECTURE — the skillsdir contract and Find() data flow
- file: plan/001_fcde63e5bb60/architecture/go_architecture.md
  why: "Defines internal/skillsdir's Source type + the Find() signature
        (dir, src, err) that findSibling feeds, and the §8 rule-2 wording
        (exe -> EvalSymlinks -> Dir -> /skills)."
  section: "internal/skillsdir (Core types)" and "Key implementation notes #2"

- file: plan/001_fcde63e5bb60/architecture/external_deps.md
  why: "§4 lists the verified stdlib API table: os.Executable() +
        filepath.EvalSymlinks() for binary/symlink resolution; path/filepath for
        Dir/Join. Confirms no runtime deps beyond stdlib + yaml.v3."
  section: "4. Go standard library APIs to use (verified available in Go 1.26)"

- file: PRD.md
  why: "§8.2 is the authoritative rule-2 spec; §12.1 explains why symlink-not-copy
        (this rule is why it works); §13 acceptance exercises the cross-dir symlink
        test. READ-ONLY — do not modify."
  critical: "§8.2 exact sequence: 'compute exe, _ := os.Executable(), then real, _ :=
             filepath.EvalSymlinks(exe); let repoDir = filepath.Dir(real); if
             repoDir/skills exists, use it.'"

- url: https://pkg.go.dev/os#Executable
  why: "os.Executable() returns the path of the running binary; on Linux via
        /proc/self/exe (already symlink-resolved). Returns an error only in
        exotic cases (deleted binary, unsupported platform)."
- url: https://pkg.go.dev/path/filepath#EvalSymlinks
  why: "filepath.EvalSymlinks resolves ALL symlinks in a path to the final target.
        Returns an error if any component does not exist. This is the call rule 2
        MUST make (macOS) and that rule 1 must NOT make (env path)."
- url: https://pkg.go.dev/path/filepath#Dir
  why: "filepath.Dir returns the directory portion of a path (drops the last elem)."
- url: https://pkg.go.dev/testing#T.TempDir
  why: "t.TempDir() creates + auto-cleans a per-test scratch dir; returns an
        absolute, cleaned path. The test scaffolding for rule-2 tests."
```

### Current Codebase tree (S1 on disk; before this subtask)

```bash
$ cd /home/dustin/projects/skpp && find . -name '*.go' -not -path './.pi-subagents/*'
internal/skillsdir/skillsdir.go        # S1: Source type + String() + findEnv() [rule 1]
internal/skillsdir/skillsdir_test.go   # S1: unsetEnvVar + 8 tests (findEnv + Source)

$ ls -A
.git/  .gitignore  LICENSE  PRD.md  go.mod  go.sum  internal/  plan/  .pi-subagents/
# go.mod: module github.com/dabstractor/skpp, go 1.25, yaml.v3 // indirect
```

### Desired Codebase tree with files to be added/modified

```bash
skpp/
├── ... (go.mod, go.sum, .gitignore, LICENSE, PRD.md — UNCHANGED)
└── internal/
    └── skillsdir/
        ├── skillsdir.go        # MODIFY — APPEND findSibling() + resolveSiblingFromExe()
        └── skillsdir_test.go   # MODIFY — APPEND findSibling/resolveSiblingFromExe tests
```

| File (modified) | Change | Consumed by |
|---|---|---|
| `internal/skillsdir/skillsdir.go` | Append rule-2 entry + testable core | S3 (`Find()` chains findEnv→findSibling→findWalkUp), T3 (`--path` prints `SourceSibling`) |
| `internal/skillsdir/skillsdir_test.go` | Append rule-2 tests | `go test` (validation gate) |

**No new files. No new directories. No import changes.**

### Known Gotchas of our codebase & Go stdlib

```go
// GOTCHA #1 — os.Executable() cannot be controlled in a test. In `go test` it
// returns a path like /tmp/go-buildXXXXX/b001/exe/<pkg>.test (a temp build dir
// with NO sibling skills/). VERIFIED (research/verified_facts.md §2). Therefore
// the REAL findSibling() can only be exercised as the found=false smoke case.
// The symlink-resolution behavior — the whole point of rule 2 — is tested via
// resolveSiblingFromExe(exe), which takes the exe path as a parameter so tests
// can pass a symlink in a different temp dir. Do NOT inline everything into
// findSibling or the symlink path goes untested.
//
//   RIGHT: findSibling() -> os.Executable() -> resolveSiblingFromExe(exe)
//   WRONG: findSibling() does EvalSymlinks+Dir inline (untestable symlink path)

// GOTCHA #2 — Keep EvalSymlinks. On Linux os.Executable() ALREADY returns the
// real path (/proc/self/exe), so EvalSymlinks looks redundant. It is NOT: macOS
// needs it (os.Executable may return the symlink path there). VERIFIED
// (architecture/verified_symlink_resolution.md). Implement BOTH calls exactly.
// (Contrast: rule 1 in findEnv must NOT EvalSymlinks the env path — the
// asymmetry is the contract; do not "unify" them.)

// GOTCHA #3 — EvalSymlinks ERROR must fall back to exe, NOT skip the rule.
// Contract: 'real, err := filepath.EvalSymlinks(exe); if err, use exe as fallback.'
// VERIFIED (research/verified_facts.md §4): a non-existent exe whose Dir has a
// sibling skills/ still wins via the fallback. Implement as:
//   real, err := filepath.EvalSymlinks(exe)
//   if err != nil { real = exe }
// NOT:  if err != nil { return "", false }   // WRONG: drops the fallback

// GOTCHA #4 — os.Executable() ERROR skips the rule (return found=false). This is
// different from #3: EvalSymlinks errors are recoverable (fallback to exe);
// os.Executable() errors are not (no exe path at all). The contract is explicit:
// 'exe, err := os.Executable(); if err, skip rule.'
//   exe, err := os.Executable()
//   if err != nil { return "", 0, false }

// GOTCHA #5 — A fake "binary" is just a regular file. EvalSymlinks and
// os.Stat(Join(dir,"skills")) do not require a real ELF executable. VERIFIED
// (research/verified_facts.md §5). Tests use os.WriteFile(exe, []byte("x"), 0o644);
// do NOT `go build` a helper binary in tests.

// GOTCHA #6 — No NEW imports. skillsdir.go already imports "os" and
// "path/filepath" (S1), which are exactly what rule 2 needs. Reusing them means
// the import block is UNTOUCHED. Do not add "fmt" or any other package.

// GOTCHA #7 — The (dir, src, found) shape is locked by S1 (no error return).
// findSibling matches it: src is only meaningful when found==true; on found==false
// return ("", 0, false) so S3's Find() falls through to findWalkUp. Do NOT add an
// error return to findSibling "for symmetry with Find()" — Find() (S3) is the ONLY
// rule that produces an error, and only when all three rules miss.

// GOTCHA #8 — findSibling (and resolveSiblingFromExe) appear unused until S3
// wires them into Find(). go build / go vet do NOT flag unused package-level
// funcs; no linter is configured. Do NOT delete them or stub a Find() to "use"
// them. (Same situation S1's findEnv already has.)

// GOTCHA #9 — Symlink tests must t.Skipf on os.Symlink error (platforms without
// symlink support), matching S1's TestFindEnvDoesNotResolveSymlinks convention.
// On linux/amd64 (CI/host) the symlink tests RUN, not skip.
```

---

## Implementation Blueprint

### The helper pair (the only data/logic added)

```go
// findSibling implements PRD §8 rule 2: locate <repoDir>/skills next to the
// running binary, symlink-aware. It is a thin entry that asks the OS for the
// running binary path and delegates the symlink/dir logic to
// resolveSiblingFromExe (which is unit-testable with arbitrary exe paths —
// os.Executable() itself cannot be controlled in a test).
//
// Returns (candidate, SourceSibling, true) when the binary's repo dir contains
// an existing skills/ subdir; otherwise ("", 0, false) so Find() (S3) falls
// through to rule 3. Never returns an error (matches the locked per-rule shape).
func findSibling() (dir string, src Source, found bool)

// resolveSiblingFromExe is the symlink-aware sibling-resolution core, factored
// out so it can be tested with arbitrary exe paths (t.TempDir + os.Symlink).
//
// Sequence (PRD §8.2):
//   real, err := filepath.EvalSymlinks(exe)   // REQUIRED on macOS, harmless on Linux
//   if err != nil { real = exe }              // fall back to the raw exe path
//   repoDir  := filepath.Dir(real)
//   candidate := filepath.Join(repoDir, "skills")
//   win iff os.Stat(candidate) is an existing dir
func resolveSiblingFromExe(exe string) (dir string, found bool)
```

### File 1 — `internal/skillsdir/skillsdir.go` (APPEND this code)

The file already has `package skillsdir`, the `import ("os"; "path/filepath")`
block, the `Source` type/constants/`String()`, and `findEnv()`. **Append** the
two functions below at the end of the file. **Do not touch the existing code or
imports** (the needed imports are already present).

```go
// findSibling implements PRD §8 rule 2 — locate <repoDir>/skills next to the
// running binary, symlink-aware. This is the rule that makes a symlink install
// work: ~/.local/bin/skpp -> ~/projects/skpp/skpp resolves back to the repo.
//
// It is a thin entry that asks the OS for the running binary (os.Executable)
// and delegates the symlink/dir logic to resolveSiblingFromExe. os.Executable
// cannot be controlled in a test (it returns the test binary's own path), so
// the testable core lives in resolveSiblingFromExe.
//
// Returns (candidate, SourceSibling, true) on a hit; ("", 0, false) otherwise so
// Find() (S3) can fall through to rule 3. Never errors (locked per-rule shape).
func findSibling() (dir string, src Source, found bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", 0, false // no binary path at all -> skip rule
	}
	d, ok := resolveSiblingFromExe(exe)
	if !ok {
		return "", 0, false
	}
	return d, SourceSibling, true
}

// resolveSiblingFromExe is the symlink-aware sibling-resolution core for rule 2,
// factored out so it can be unit-tested with arbitrary exe paths.
//
// PRD §8.2 sequence:
//   real, err := filepath.EvalSymlinks(exe)  // REQUIRED on macOS (redundant but
//                                            //   harmless on Linux via /proc/self/exe)
//   if err != nil { real = exe }             // fall back to raw exe on EvalSymlinks error
//   repoDir := filepath.Dir(real)
//   candidate := filepath.Join(repoDir, "skills")
//   win iff os.Stat(candidate) reports an existing directory
//
// See architecture/verified_symlink_resolution.md for why EvalSymlinks must stay.
func resolveSiblingFromExe(exe string) (dir string, found bool) {
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		real = exe // EvalSymlinks could not resolve -> use exe verbatim
	}
	repoDir := filepath.Dir(real)
	candidate := filepath.Join(repoDir, "skills")
	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		return "", false // no existing skills/ sibling -> rule misses
	}
	return candidate, true
}
```

### File 2 — `internal/skillsdir/skillsdir_test.go` (APPEND these tests)

The file is white-box (`package skillsdir`) and already imports `os`,
`path/filepath`, `testing` (S1). **Append** the helpers + tests below. **Do not
touch the existing tests.** A small `makeFakeBinary` helper keeps the cases DRY.

```go
// makeFakeBinary creates a regular file at dir/name to stand in for a compiled
// binary. EvalSymlinks + os.Stat(Join(dir,"skills")) do not require a real ELF,
// so a 1-byte file is sufficient (research/verified_facts.md §5).
func makeFakeBinary(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatalf("write fake binary %s: %v", p, err)
	}
	return p
}

// --- resolveSiblingFromExe (the testable core) ---

// Rule 2 CONTRACT: a symlink to the binary in a DIFFERENT dir must resolve back
// to the REAL binary's repo dir's skills/. Mirrors architecture/verified_symlink_resolution.md.
func TestResolveSiblingFromExeSymlinkCrossDir(t *testing.T) {
	// tempA holds the REAL binary + its sibling skills/
	tempA := t.TempDir()
	binary := makeFakeBinary(t, tempA, "skpp")
	skillsA := filepath.Join(tempA, "skills")
	if err := os.Mkdir(skillsA, 0o755); err != nil {
		t.Fatal(err)
	}
	// tempB holds a symlink to the binary (different dir, like ~/.local/bin)
	tempB := t.TempDir()
	link := filepath.Join(tempB, "skpp")
	if err := os.Symlink(binary, link); err != nil {
		t.Skipf("symlinks not supported on this platform: %v", err)
	}

	got, found := resolveSiblingFromExe(link)
	if !found {
		t.Fatalf("resolveSiblingFromExe(symlink): found=false; want true")
	}
	if got != skillsA {
		t.Errorf("resolveSiblingFromExe(symlink): dir=%q; want the REAL binary's skills %q", got, skillsA)
	}
	if filepath.Dir(got) != tempA {
		t.Errorf("resolveSiblingFromExe(symlink): resolved to %q, not the real binary's dir %q", filepath.Dir(got), tempA)
	}
}

// Rule 2: direct (non-symlinked) binary with a sibling skills/ also wins.
func TestResolveSiblingFromExeDirect(t *testing.T) {
	tempA := t.TempDir()
	binary := makeFakeBinary(t, tempA, "skpp")
	skillsA := filepath.Join(tempA, "skills")
	if err := os.Mkdir(skillsA, 0o755); err != nil {
		t.Fatal(err)
	}
	got, found := resolveSiblingFromExe(binary)
	if !found {
		t.Fatalf("resolveSiblingFromExe(direct): found=false; want true")
	}
	if got != skillsA {
		t.Errorf("resolveSiblingFromExe(direct): dir=%q; want %q", got, skillsA)
	}
}

// Rule 2: EvalSymlinks-error fallback. A non-existent exe whose parent dir DOES
// have a sibling skills/ must still win via real=exe. (Contract: 'if err, use
// exe as fallback.')
func TestResolveSiblingFromExeEvalSymlinksFallback(t *testing.T) {
	tempC := t.TempDir()
	skillsC := filepath.Join(tempC, "skills")
	if err := os.Mkdir(skillsC, 0o755); err != nil {
		t.Fatal(err)
	}
	// 'ghost' binary does not exist -> EvalSymlinks errors -> fall back to exe.
	ghost := filepath.Join(tempC, "does-not-exist-binary")
	got, found := resolveSiblingFromExe(ghost)
	if !found {
		t.Fatalf("resolveSiblingFromExe(ghost): found=false; want true (EvalSymlinks fallback to exe)")
	}
	if got != skillsC {
		t.Errorf("resolveSiblingFromExe(ghost): dir=%q; want %q (Dir(exe)/skills)", got, skillsC)
	}
}

// Rule 2: binary exists but NO sibling skills/ dir -> miss.
func TestResolveSiblingFromExeNoSkillsDir(t *testing.T) {
	tempA := t.TempDir()
	binary := makeFakeBinary(t, tempA, "skpp")
	// deliberately create no skills/ sibling
	if _, found := resolveSiblingFromExe(binary); found {
		t.Errorf("resolveSiblingFromExe(no skills): got found=true; want false")
	}
}

// Rule 2: sibling path 'skills' is a regular FILE, not a dir -> miss (IsDir guard).
func TestResolveSiblingFromExeSkillsIsFile(t *testing.T) {
	tempA := t.TempDir()
	binary := makeFakeBinary(t, tempA, "skpp")
	if err := os.WriteFile(filepath.Join(tempA, "skills"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, found := resolveSiblingFromExe(binary); found {
		t.Errorf("resolveSiblingFromExe(skills is file): got found=true; want false (IsDir guard)")
	}
}

// --- findSibling (the rule-2 entry; os.Executable exercised) ---

// Smoke test: the REAL test binary runs from a temp build dir (go-buildXXX)
// that has NO sibling skills/, so findSibling must return found=false without
// panicking. This is the only deterministic assertion possible for findSibling
// (os.Executable cannot be controlled); the symlink logic is covered by the
// resolveSiblingFromExe tests above.
func TestFindSiblingNoSkillsNextToTestBinary(t *testing.T) {
	dir, src, found := findSibling()
	if found {
		t.Errorf("findSibling(): got found=true dir=%q src=%v; want false (test binary's dir has no sibling skills/)", dir, src)
	}
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: PRECONDITION — confirm S1 is on disk and green
  - COMMAND: cd /home/dustin/projects/skpp && test -f internal/skillsdir/skillsdir.go && test -f internal/skillsdir/skillsdir_test.go
  - COMMAND: go test ./internal/skillsdir/ >/dev/null 2>&1 && echo "S1 green" || echo "S1 not green"
  - EXPECT: file exists AND S1 tests pass. If S1 is missing/red, P1.M1.T2.S1 has
            not landed — stop and let it land first.

Task 1: APPEND rule-2 helpers to internal/skillsdir/skillsdir.go
  - EDIT: append the findSibling() + resolveSiblingFromExe() code from the
          Blueprint at the END of the file (after findEnv()).
  - NAMING: findSibling (unexported, matches the (dir,src,found) shape);
            resolveSiblingFromExe (unexported, (dir,found) shape)
  - DO NOT: change the import block ("os"+"path/filepath" already present);
            do NOT add fmt; do NOT touch Source/findEnv; do NOT add Find()/findWalkUp
  - GOTCHA: EvalSymlinks error -> real=exe (fallback), NOT skip; os.Executable
            error -> skip rule (return "",0,false)

Task 2: APPEND rule-2 tests to internal/skillsdir/skillsdir_test.go
  - EDIT: append makeFakeBinary helper + the 6 test functions from the Blueprint
          at the END of the file.
  - COVERAGE: resolveSiblingFromExe (symlink-cross-dir win, direct win, EvalSymlinks
             fallback, no-skills miss, skills-is-file miss) + findSibling smoke test
  - NAMING: TestResolveSiblingFromExe{SymlinkCrossDir,Direct,EvalSymlinksFallback,
            NoSkillsDir,SkillsIsFile}, TestFindSiblingNoSkillsNextToTestBinary
  - GOTCHA: symlink test uses t.Skipf on os.Symlink error (platform convention);
            no t.Parallel() needed (these tests do not mutate process env, but keep
            them serial-safe by not opting into parallel)

Task 3: FORMAT + VET + BUILD + TEST (validation gates — run in order)
  - COMMAND: gofmt -w internal/skillsdir/        # format in place
  - COMMAND: gofmt -l internal/skillsdir/        # MUST print nothing
  - COMMAND: go vet ./internal/skillsdir/        # MUST be clean
  - COMMAND: go build ./...                       # exit 0
  - COMMAND: go test ./internal/skillsdir/ -v     # all S1 + S2 tests PASS
  - EXPECT: zero errors, zero vet findings, gofmt silent, symlink test RUNS (not skipped) on linux

Task 4: SCOPE BOUNDARY CHECK
  - COMMAND: the Level 3 block in Validation Loop below
  - EXPECT: no Find()/findWalkUp; go.mod/go.sum/PRD.md unchanged; imports unchanged
```

### Implementation Patterns & Key Details

```go
// PATTERN: the rule-2 two-function split (thin entry + testable core).
//   func findSibling() (dir string, src Source, found bool)   // os.Executable -> core
//   func resolveSiblingFromExe(exe string) (dir string, found bool)  // symlink/dir logic
// WHY: os.Executable() returns the running test binary's path and CANNOT be set,
//      so the symlink-resolution logic must live in a parameterized helper to be
//      testable. findSibling matches S1's locked (dir,src,found) shape; the core
//      drops src (the entry always wins as SourceSibling).

// PATTERN: the two-call sequence with asymmetric error handling (the contract).
//   exe, err := os.Executable()
//   if err != nil { return "", 0, false }              // os.Executable error: SKIP rule
//   ...
//   real, err := filepath.EvalSymlinks(exe)
//   if err != nil { real = exe }                       // EvalSymlinks error: FALL BACK, keep going
// WHY: os.Executable errors mean "no binary path at all" (unrecoverable); EvalSymlinks
//      errors are recoverable (use the raw path). Do not conflate them.

// PATTERN: the "fall through on any non-dir" check (os.Stat, NOT os.Lstat) — same as rule 1.
//   info, err := os.Stat(candidate)
//   if err != nil || !info.IsDir() { return "", false }
// os.Stat follows symlinks; os.Lstat would wrongly reject a symlinked skills dir.

// PATTERN: fake-binary test scaffolding (no go build in tests).
//   func makeFakeBinary(t, dir, name) string {
//       p := filepath.Join(dir, name); os.WriteFile(p, []byte("x"), 0o644); return p
//   }
// EvalSymlinks + Stat(Join(dir,"skills")) do not need a real ELF; a 1-byte file suffices.
```

### Integration Points

```yaml
PACKAGE BOUNDARIES (unchanged by S2):
  - import path: "github.com/dabstractor/skpp/internal/skillsdir"
  - S2 adds NO new exported symbols (findSibling + resolveSiblingFromExe are unexported).
  - reused from S1: Source, SourceSibling, (Source).String (already on disk).

DOWNSTREAM CONSUMERS (what relies on this subtask's output):
  - P1.M1.T2.S3: adds findWalkUp() + the public Find() that chains
    findEnv -> findSibling -> findWalkUp -> ErrNotFound, returning (dir, src, err).
    S3 calls findSibling() (this subtask) and reads its (dir, src, found) return.
  - P1.M1.T3 (main.go --path): prints src.String() -> "sibling of binary" when
    rule 2 wins (the most common case in a normal install).

NO CHANGES TO:
  - go.mod / go.sum (no new deps; pure stdlib)
  - PRD.md (read-only)
  - the Source type / findEnv / Source.String() (S1-owned; do not modify)
  - any other package (discover/resolve/ui/main are later subtasks)
```

---

## Validation Loop

### Level 1: Format, vet, build (immediate, per file)

```bash
cd /home/dustin/projects/skpp

# Format in place, then confirm nothing left unformatted (silent == pass)
gofmt -w internal/skillsdir/
test -z "$(gofmt -l internal/skillsdir/)" || { echo "FAIL: gofmt found unformatted files"; gofmt -d internal/skillsdir/; exit 1; }
echo "gofmt OK"

# Vet the package
go vet ./internal/skillsdir/ || { echo "FAIL: go vet"; exit 1; }
echo "go vet OK"

# Build the whole module
go build ./... || { echo "FAIL: go build ./..."; exit 1; }
echo "go build ./... OK"
```

### Level 2: Unit tests (component validation)

```bash
cd /home/dustin/projects/skpp

# Run the skillsdir tests verbosely — S1 + S2 functions must all PASS
go test ./internal/skillsdir/ -v

# Explicit assertions the run must satisfy:
go test ./internal/skillsdir/ -run 'TestResolveSiblingFromExe|TestFindSibling' -v \
  || { echo "FAIL: rule-2 tests"; exit 1; }

# The symlink-cross-dir contract test must RUN (not SKIP) on linux/amd64 — it is
# the whole point of rule 2.
go test ./internal/skillsdir/ -run TestResolveSiblingFromExeSymlinkCrossDir -v \
  | grep -E '--- (PASS|SKIP):.*TestResolveSiblingFromExeSymlinkCrossDir' \
  || { echo "FAIL: symlink contract test did not run"; exit 1; }

# Full module still green (only skillsdir has tests so far)
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Scope-boundary & contract check

```bash
cd /home/dustin/projects/skpp

# findSibling + resolveSiblingFromExe now EXIST
grep -qE 'func findSibling\(\) \(dir string, src Source, found bool\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: findSibling missing or wrong signature"; exit 1; }
grep -qE 'func resolveSiblingFromExe\(exe string\) \(dir string, found bool\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: resolveSiblingFromExe missing or wrong signature"; exit 1; }

# EvalSymlinks is PRESENT (do not drop it — macOS needs it)
grep -q 'filepath.EvalSymlinks(exe)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: EvalSymlinks missing — required for macOS"; exit 1; }

# EvalSymlinks error falls back to exe (NOT a skip)
grep -qE 'real = exe' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: EvalSymlinks fallback to exe missing"; exit 1; }

# os.Executable error skips the rule (return found=false), distinct from the fallback
grep -qE 'os.Executable\(\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: os.Executable missing"; exit 1; }

# MUST NOT have created Find() / findWalkUp yet (S3 deliverables)
! grep -nE 'func (Find|findWalkUp)\b' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: found S3 symbols — out of scope"; exit 1; }

# MUST NOT have touched go.mod / go.sum / PRD.md
git diff --quiet go.mod   || { echo "FAIL: go.mod changed (should be untouched)"; exit 1; }
git diff --quiet go.sum   || { echo "FAIL: go.sum changed (should be untouched)"; exit 1; }
git diff --quiet PRD.md   || { echo "FAIL: PRD.md changed (read-only)"; exit 1; }

# imports unchanged: still exactly "os" + "path/filepath" (no fmt added)
grep -A3 '^import (' internal/skillsdir/skillsdir.go | grep -q '"os"'       || { echo "FAIL: os import missing"; exit 1; }
grep -A3 '^import (' internal/skillsdir/skillsdir.go | grep -q '"path/filepath"' || { echo "FAIL: path/filepath import missing"; exit 1; }
! grep -A4 '^import (' internal/skillsdir/skillsdir.go | grep -q '"fmt"' \
  || { echo "FAIL: fmt import added (not needed)"; exit 1; }

echo "Level 3 PASS (scope + contract respected)"
```

### Level 4: Downstream-readiness smoke test

Prove rule 2 is ready for S3's `Find()` and the PRD §13 symlink scenario —
without depending on S3/main existing yet.

```bash
cd /home/dustin/projects/skpp

# (a) findSibling is callable in-package and produces the contracted (dir,src,found)
#     shape — covered by go test above; here we confirm the Source label downstream.
mkdir -p /tmp/skpp-sib-check && cat > /tmp/skpp-sib-check/main.go <<'EOF'
package main
import (
	"fmt"
	"github.com/dabstractor/skpp/internal/skillsdir"
)
func main() {
	// SourceSibling label (findSibling wins as this) is what --path will print.
	fmt.Println(skillsdir.SourceSibling.String()) // sibling of binary
}
EOF
cat > /tmp/skpp-sib-check/go.mod <<EOF
module skpp-sib-check
go 1.25
require github.com/dabstractor/skpp v0.0.0
replace github.com/dabstractor/skpp => $(pwd)
EOF
( cd /tmp/skpp-sib-check && go run . ) | grep -qx 'sibling of binary' \
  || { echo "FAIL: SourceSibling label not usable downstream"; rm -rf /tmp/skpp-sib-check; exit 1; }
rm -rf /tmp/skpp-sib-check

# (b) Live cross-dir symlink proof (the PRD §13 acceptance scenario for rule 2),
#     run via the in-package test binary by setting SKPP_SKILLS_DIR-less env.
#     The resolveSiblingFromExe test already proves this; here we just confirm
#     the package compiles + the Source for rule 2 is wired for S3.
echo "Level 4 PASS (rule-2 output ready for S3's Find() and T3's --path)"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` silent, `go vet ./internal/skillsdir/` clean, `go build ./...` exit 0
- [ ] Level 2 PASS — `go test ./internal/skillsdir/ -v`: all S1 + S2 tests pass; symlink contract test RAN (not skipped) on linux/amd64
- [ ] Level 3 PASS — helpers present with correct signatures; EvalSymlinks present + fallback present; no S3 symbols; imports unchanged; go.mod/go.sum/PRD.md unchanged
- [ ] Level 4 PASS — `SourceSibling.String()` usable downstream

### Feature Validation
- [ ] `findSibling()` returns `(candidate, SourceSibling, true)` when the binary's repo dir has an existing `skills/`
- [ ] `findSibling()` returns `found=false` when `os.Executable()` errors (skip rule)
- [ ] `resolveSiblingFromExe` resolves a cross-dir symlink back to the real binary's `skills/` (contract test)
- [ ] `resolveSiblingFromExe` honors the EvalSymlinks-error fallback (falls back to `exe`, still wins if `Dir(exe)/skills` exists)
- [ ] `resolveSiblingFromExe` returns `found=false` when there is no `skills/` sibling OR it is a regular file

### Code Quality / Convention Validation
- [ ] Both helpers appended to the EXISTING file (no new files/dirs)
- [ ] White-box tests appended to the EXISTING test file; existing S1 tests untouched
- [ ] Imports unchanged (`os` + `path/filepath`; no `fmt` or new deps)
- [ ] Helper names match S1's precedent (`findSibling` ← `findEnv`; `resolveSiblingFromExe` is the testable core)
- [ ] Symlink tests use `t.Skipf` on `os.Symlink` error (platform convention)

### Scope Discipline
- [ ] Did NOT create `Find()` or `findWalkUp` (S3)
- [ ] Did NOT modify the `Source` type, `findEnv`, or `Source.String()` (S1-owned)
- [ ] Did NOT create `main.go`, `internal/discover`, `internal/resolve`, `internal/ui` (later milestones)
- [ ] Did NOT modify `go.mod` / `go.sum` (no new deps)
- [ ] Did NOT modify `PRD.md` (read-only) or any `tasks.json` (orchestrator-owned)

---

## Anti-Patterns to Avoid

- ❌ **Don't inline the symlink/dir logic into `findSibling`.** `os.Executable()`
  cannot be controlled in a test, so the cross-dir symlink behavior — the entire
  point of rule 2 — would go untested. Extract `resolveSiblingFromExe(exe)` and
  test it with `t.TempDir()` + `os.Symlink`. (Verified: `research/verified_facts.md §2`.)
- ❌ **Don't drop `filepath.EvalSymlinks`.** It looks redundant on Linux
  (`os.Executable()` already returns the real path via `/proc/self/exe`) but is
  REQUIRED on macOS. Implement BOTH calls. The contrast with rule 1 (which must
  NOT EvalSymlinks the env path) is intentional — do not "unify" them.
  (Verified: `architecture/verified_symlink_resolution.md`.)
- ❌ **Don't conflate the two error paths.** `os.Executable()` error ⇒ skip the
  rule (`return "", 0, false`). `EvalSymlinks` error ⇒ fall back to `exe` and
  KEEP GOING (still may win via `Dir(exe)/skills`). The fallback is load-bearing
  and verified (`research/verified_facts.md §4`).
- ❌ **Don't return an error from `findSibling`.** It matches S1's locked
  `(dir, src, found)` shape; `src`'s zero value is ignored on `found==false` so
  S3's `Find()` falls through to rule 3. Only `Find()` (S3) errors, and only when
  all three rules miss.
- ❌ **Don't use `os.Lstat`.** It does not follow symlinks; a symlinked `skills/`
  dir would be wrongly rejected. Use `os.Stat` (same as rule 1).
- ❌ **Don't `go build` a helper binary in tests.** `EvalSymlinks` and
  `os.Stat(Join(dir,"skills"))` do not require a real ELF; a 1-byte `os.WriteFile`
  fake is sufficient and far simpler. (Verified: `research/verified_facts.md §5`.)
- ❌ **Don't modify the import block or add `fmt`.** `os` + `path/filepath` are
  already imported (S1) and are exactly what rule 2 needs. An unused import is a
  compile error; keep the block as-is.
- ❌ **Don't touch S1's code.** Append only. The `Source` type, `findEnv`, and
  `Source.String()` are S1-owned and already pass their tests.
- ❌ **Don't stub `Find()` to "use" `findSibling`.** `Find()` is S3's deliverable;
  `go build`/`go vet` do not flag unused package-level funcs, and no linter is
  configured.

---

## Confidence Score

**10/10** — A small, pure-stdlib addition to an existing, already-green package.
The exact source for both appended functions is given verbatim, and every stdlib
behavior they rely on (`os.Executable()` path in tests, `EvalSymlinks` symlink
resolution + error fallback, `os.Stat` dir-check) was **empirically verified** in
the target Go 1.26.4 environment, including a pre-run of the symlink-cross-dir
test and the EvalSymlinks-fallback test (`research/verified_facts.md`). The
testing crux — `os.Executable()` is untestable, so extract a parameterized core —
is the single non-obvious design decision, and it is documented and justified.
No new imports, no new deps, no concurrency. The only residual (non-)risk — an
unused-symbol linter flagging `findSibling`/`resolveSiblingFromExe` — is
explicitly excluded by the validation gates (`go build`/`go vet`/`go test`/`gofmt`,
no staticcheck) and by an inline anti-pattern.
