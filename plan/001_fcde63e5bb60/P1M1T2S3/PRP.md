# PRP — P1.M1.T2.S3: `internal/skillsdir` — walk-up-from-cwd (rule 3) + `Find()` combiner (§8 rules 3-4)

> **Subtask:** P1.M1.T2.S3 — the THIRD and FINAL subtask that builds
> `internal/skillsdir`. **Scope:** add `findWalkUp()` (rule 3), its testable core
> `findWalkUpAncestor()`, the `hasSkillMD()` helper, the exported `ErrNotFound`
> sentinel, and the public `Find()` entrypoint that chains rule 1 → rule 2 →
> rule 3 → error. This subtask **completes** the skills-dir-location feature
> (PRD §8) and produces the public API consumed by `main.go` (P1.M1.T3) and
> `discover.Index` (P1.M2.T5).
>
> **Status of S1+S2:** S1 (`Source` type + `findEnv`) is COMPLETE on disk and
> green. S2 (`findSibling` + `resolveSiblingFromExe`) is being implemented in
> parallel and is treated as a **CONTRACT** — assume it lands exactly as its PRP
> specifies. This PRP **MODIFIES** the two existing files; it appends code after
> S2's code and grows the import block. It creates no new files.
>
> **IMPLEMENTATION ORDER NOTE:** S3 is implemented AFTER S2 lands. Task 0 below
> asserts S2's symbols are present; if they are not, S3 must wait for S2 (do not
> recreate S2's code).

---

## Goal

**Feature Goal**: Add PRD §8 rule 3 (walk-up-from-cwd) and the public `Find()`
combiner to `internal/skillsdir`. Rule 3 ascends from the current working
directory, checking each ancestor (including cwd itself) for a `skills/` subdir
that contains **at least one `SKILL.md`** at any depth; the first such ancestor
wins. `Find()` chains the three rules in priority order and, on total miss,
returns an exported `ErrNotFound` whose message is the user-facing one-line fix
(`set $SKPP_SKILLS_DIR`, `cd` into the repo, or reinstall). This closes out PRD
§8 and unblocks every downstream consumer (`main --path`, `discover.Index`).

**Deliverable**: Two existing files are **modified** (no new files):
1. `internal/skillsdir/skillsdir.go` — grow the import block by `"errors"` +
   `"io/fs"`; append `errSkillMDFound` sentinel, `hasSkillMD()`,
   `findWalkUpAncestor()` (testable core), `findWalkUp()` (rule-3 entry, matching
   the S1/S2 `(dir, src, found)` shape), `ErrNotFound`, and `Find()`.
2. `internal/skillsdir/skillsdir_test.go` — grow the import block by `"errors"` +
   `"strings"`; append a `makeSkill` helper and tests covering `hasSkillMD`,
   `findWalkUpAncestor` (every branch incl. the skip-empty-and-continue
   contract), `findWalkUp` (via `t.Chdir`), `Find()` (rule-1 win, rule-3 win,
   all-miss → `ErrNotFound`), and the `ErrNotFound` message fix.

**Success Definition**: `go build ./...` exits 0; `go vet ./internal/skillsdir/`
clean; `gofmt -l internal/skillsdir/` silent; `go test ./internal/skillsdir/ -v`
passes all S1 + S2 + S3 cases — including the contract tests that (a) an
ancestor with an empty `skills/` dir is skipped and ascent continues, (b)
`Find()` reaches rule 3 and returns `SourceWalkUp` when env is unset and cwd has
an ancestor skills tree, and (c) `Find()` returns `ErrNotFound` (with the fix
phrase) when all three rules miss. `go.mod`/`go.sum`/`PRD.md` unchanged. No
new packages created (no `main.go`, no `internal/discover`, etc.).

---

## Why

- This subtask **finishes the hardest, most failure-prone part of skpp** (PRD §8,
  risk area #1). Rules 1 and 2 are useless without the combiner that orders and
  exits them. Until `Find()` exists, nothing downstream can run.
- Rule 3 is what makes **`go run`** (and dev iteration) work: `go run` puts the
  binary in a temp build dir (rules 1-2 both miss), but the developer's cwd is
  inside the repo, so walking up finds `./skills`. Without rule 3, `go run .`
  always fails location resolution.
- `Find()` is the **single public entrypoint** for the whole CLI:
  `main.go` (P1.M1.T3) calls it to resolve `--path`/discover, and
  `discover.Index` (P1.M2.T5) takes its `absDir` to walk. Locking the exact
  `(dir, src, err)` signature + the `ErrNotFound` message here lets both
  consumers be written against a stable contract.
- The `hasSkillMD` check guards against a subtle false-positive: an ancestor
  that happens to have an empty (or non-skill) `skills/` dir must NOT be
  reported as the skills store. PRD §8.3 explicitly qualifies the match with
  "at least one `SKILL.md`". The skip-empty-and-continue behavior is verified
  (see `research/verified_facts.md §4`) and is a contract test.

---

## What

Append rule 3 + the combiner to the existing package, conforming to the
`(dir, src, found bool)` per-rule shape S1/S2 locked. Concretely:

1. `func findWalkUp() (dir string, src Source, found bool)` — the rule-3 entry:
   `cwd, err := os.Getwd()`; on error return `("", 0, false)`; delegate to
   `findWalkUpAncestor(cwd)`; on hit return `(candidate, SourceWalkUp, true)`,
   else `("", 0, false)`.
2. `func findWalkUpAncestor(start string) (dir string, found bool)` — the
   testable ascent core (extracted because `os.Getwd` cannot be parameterized;
   same factoring rationale as S2's `resolveSiblingFromExe`). Ascends `start`
   → parent → … → root, returning the first ancestor whose `skills/` subdir
   `os.Stat`s as a directory AND contains ≥1 `SKILL.md` (via `hasSkillMD`).
   Terminates when `filepath.Dir(cur) == cur` (the root).
3. `func hasSkillMD(dir string) bool` — `filepath.WalkDir` over `dir`, returning
   true on the first `SKILL.md` (early exit via a sentinel error). **Uses
   WalkDir, NOT `filepath.Glob` with `**`** — Go's stdlib does not support `**`
   (verified: returns 0 matches for nested files).
4. `var ErrNotFound = errors.New("could not locate the skills directory: set $SKPP_SKILLS_DIR, cd into the skpp repo, or reinstall skpp")` — exported sentinel; the message is the PRD §8.4/§6.4 one-line fix.
5. `func Find() (dir string, src Source, err error)` — the public entrypoint:
   `findEnv()` → `findSibling()` → `findWalkUp()`, first hit wins; else
   `("", 0, ErrNotFound)`.

### Success Criteria

- [ ] `findWalkUp()` has the exact `(dir string, src Source, found bool)` signature (matches S1/S2)
- [ ] `findWalkUpAncestor(start string) (dir string, found bool)` exists (private, testable core)
- [ ] `hasSkillMD(dir string) bool` exists and uses `filepath.WalkDir` (NOT `filepath.Glob` with `**`)
- [ ] Walk-up checks `start` (cwd) first, then each ancestor up to filesystem root
- [ ] An ancestor with an empty `skills/` dir (no `SKILL.md`) is **skipped**; ascent continues to a higher ancestor
- [ ] `Find()` exists with signature `func Find() (dir string, src Source, err error)`
- [ ] `Find()` returns the first rule's hit as `(absDir, src, nil)`; on total miss returns `("", 0, ErrNotFound)`
- [ ] `ErrNotFound` is exported and its message contains `SKPP_SKILLS_DIR`, `cd`, and `reinstall`
- [ ] `errors.Is(err, ErrNotFound)` is true for the all-miss case
- [ ] `go build ./...` exits 0; `go vet ./internal/skillsdir/` clean; `gofmt -l` silent
- [ ] `go test ./internal/skillsdir/ -v` — all S1 + S2 + S3 tests pass
- [ ] Import block gains exactly `"errors"` + `"io/fs"` in skillsdir.go (and `"errors"` + `"strings"` in the test file); `go.mod`/`go.sum`/`PRD.md` unchanged

---

## All Needed Context

### Context Completeness Check

_Pass: the exact source for every added symbol is given verbatim in the
Implementation Blueprint, and every stdlib behavior it relies on was
**empirically verified** in the target Go 1.26.4 environment
(`research/verified_facts.md`): `filepath.Glob` does NOT support `**`; WalkDir
early-exit sentinel stops the walk; `filepath.Dir(root) == root` terminates the
loop; skip-empty-and-continue is correct; `t.Chdir` (Go 1.24+) controls
`os.Getwd`; `findSibling` deterministically misses inside `go test` so `Find()`
can reach rule 3. S1 is read in full on disk; S2 is treated as a landed contract.
An implementer who knows Go but nothing about this repo can complete this in one
pass from this document._

### Documentation & References

```yaml
# MUST READ — this subtask's own empirical verification (the crux decisions)
- file: plan/001_fcde63e5bb60/P1M1T2S3/research/verified_facts.md
  why: "Proves: (1) filepath.Glob with '**' is BROKEN in Go (returns 0 for
        nested) -> must use WalkDir; (2) WalkDir early-exit sentinel works;
        (3) filepath.Dir('/') == '/' terminates the ascent loop; (4) the
        skip-empty-skills-and-continue §8.3 semantics; (5) t.Chdir controls
        os.Getwd for findWalkUp tests; (6) findSibling deterministically misses
        in tests so Find() reaches rule 3; (7) the exact import delta; (8) the
        exact ErrNotFound message wording."
  critical: "Do NOT use filepath.Glob('skills/**/SKILL.md') — the PRD item
             description lists it as an option, but Go's stdlib does not support
             '**'. Use filepath.WalkDir with a sentinel. This is the #1 pitfall."

# CONTRACT — the package this subtask modifies (read first, do not recreate)
- file: internal/skillsdir/skillsdir.go
  why: "S1's delivered file (on disk) + S2's appended findSibling/resolveSiblingFromExe
        (assume landed). Contains Source/SourceWalkUp/Source.String(), envVar,
        findEnv(), and (post-S2) findSibling()+resolveSiblingFromExe(). This
        subtask APPENDS rule-3 + Find() AFTER S2's code and grows the import block."
  pattern: "findEnv()/findSibling() are the templates: (dir string, src Source,
            found bool); return ('', 0, false) to fall through, ('<abs>', SourceX,
            true) to win. findWalkUp matches this shape exactly."
  gotcha: "Import block currently has 'os' + 'path/filepath' only (S2 adds none).
           S3 MUST ADD 'errors' (errors.New for ErrNotFound) and 'io/fs' (fs.DirEntry
           for the WalkDir callback). Do not import 'fmt'."

- file: internal/skillsdir/skillsdir_test.go
  why: "S1's white-box test file (package skillsdir) + S2's appended rule-2 tests.
        Contains unsetEnvVar(tb) and (post-S2) makeFakeBinary(). This subtask
        APPENDS makeSkill + rule-3/Find tests; do not touch existing tests."
  pattern: "t.TempDir() for scratch trees; os.MkdirAll/WriteFile to build skills;
            t.Chdir (Go 1.24+) to control cwd for findWalkUp/Find; plain
            t.Errorf/t.Fatalf assertions (no testify)."
  gotcha: "Test imports grow by 'errors' (errors.Is) and 'strings' (strings.Contains
           for the ErrNotFound message check). t.Chdir + t.Setenv/unsetEnvVar all
           forbid t.Parallel() — do NOT add it to any of these tests."

# CONTRACT — the locked signatures this subtask must conform to / produce
- file: plan/001_fcde63e5bb60/P1M1T2S1/research/verified_facts.md
  why: "§4 locks the per-rule helper signature func findX() (dir string, src
        Source, found bool). findWalkUp MUST match it so Find() chains uniformly."
  section: "4. Design decision — per-rule helper signature"

- file: plan/001_fcde63e5bb60/P1M1T2S2/PRP.md
  why: "Defines findSibling()+resolveSiblingFromExe() (rule 2) that Find() calls
        second, and confirms S2 adds NO imports (so the import block S3 edits is
        still the S1 version). Treat as a landed contract."
  section: "Implementation Blueprint (the two functions + signatures)"

# ARCHITECTURE — the Find() contract + §8 priority order
- file: plan/001_fcde63e5bb60/architecture/go_architecture.md
  why: "Locks Find() (dir string, src Source, err error) and the data flow
        (main -> skillsdir.Find -> discover.Index). §8 note #3 (walk-up) and #4
        (None -> error + one-line fix) are the authoritative spec for this subtask."
  section: "internal/skillsdir (Core types)" and "Key implementation notes #3/#4"

- file: plan/001_fcde63e5bb60/architecture/external_deps.md
  why: "§4 lists filepath.WalkDir as the walk primitive (the same one discover.Index
        will use); confirms no runtime deps beyond stdlib + yaml.v3."
  section: "4. Go standard library APIs to use"

- file: PRD.md
  why: "§8.3 is the authoritative rule-3 spec (ascend from cwd; first ancestor with
        a skills/ subdir containing >=1 SKILL.md wins); §8.4 is the all-miss error
        + one-line fix; §6.4 is the error-semantics contract (stderr + exit 1);
        §13 acceptance exercises SKPP_SKILLS_DIR env override. READ-ONLY."
  critical: "§8.3 wording: 'ascend from the current working directory; the first
             ancestor containing a skills/ subdir with at least one SKILL.md wins.'
             The 'at least one SKILL.md' qualifier is what hasSkillMD enforces."

- url: https://pkg.go.dev/path/filepath#WalkDir
  why: "filepath.WalkDir(root, fn) traverses the tree; fn is func(path, fs.DirEntry, error).
        Returning any non-nil error from fn STOPS the walk — the early-exit idiom for
        hasSkillMD. Requires import 'io/fs' for the fs.DirEntry param type."
- url: https://pkg.go.dev/path/filepath#Dir
  why: "filepath.Dir returns the directory portion (drops last elem). filepath.Dir('/')
        returns '/' — the ascent-loop termination condition."
- url: https://pkg.go.dev/os#Getwd
  why: "os.Getwd returns the current working directory. Cannot be parameterized, so
        findWalkUp delegates to findWalkUpAncestor(start) for testability."
- url: https://pkg.go.dev/testing#T.Chdir
  why: "t.Chdir (Go 1.24+) changes cwd for the test and restores on cleanup. The cwd
        analog of t.Setenv; how findWalkUp()/Find() rule-3 tests control os.Getwd."
- url: https://pkg.go.dev/errors#Is
  why: "errors.Is(err, target) unwraps sentinel errors. Used in the Find() all-miss
        test to assert err is ErrNotFound."
```

### Current Codebase tree (S1 on disk; S2 assumed landed; before this subtask)

```bash
$ cd /home/dustin/projects/skpp && find . -name '*.go' -not -path './.pi-subagents/*'
internal/skillsdir/skillsdir.go        # S1: Source + String() + findEnv()  | S2: findSibling() + resolveSiblingFromExe()
internal/skillsdir/skillsdir_test.go   # S1: unsetEnvVar + 8 tests | S2: makeFakeBinary + 6 rule-2 tests

$ ls -A
.git/  .gitignore  LICENSE  PRD.md  go.mod  go.sum  internal/  plan/  .pi-subagents/
# go.mod: module github.com/dabstractor/skpp, go 1.25, yaml.v3 // indirect
# NO main.go, NO internal/discover|resolve|ui yet (later milestones)
```

### Desired Codebase tree with files to be modified

```bash
skpp/
├── ... (go.mod, go.sum, .gitignore, LICENSE, PRD.md — UNCHANGED)
└── internal/
    └── skillsdir/
        ├── skillsdir.go        # MODIFY — grow imports (errors, io/fs); APPEND hasSkillMD +
        │                        #         findWalkUpAncestor + findWalkUp + ErrNotFound + Find()
        └── skillsdir_test.go   # MODIFY — grow imports (errors, strings); APPEND makeSkill +
                                 #         hasSkillMD/findWalkUpAncestor/findWalkUp/Find tests
```

| File (modified) | Change | Consumed by |
|---|---|---|
| `internal/skillsdir/skillsdir.go` | +2 imports; append rule-3 helpers + `ErrNotFound` + `Find()` | `main.go` `--path` (P1.M1.T3), `discover.Index` (P1.M2.T5) |
| `internal/skillsdir/skillsdir_test.go` | +2 imports; append rule-3 + `Find()` tests | `go test` (validation gate) |

**No new files. No new directories. No `go.mod`/`go.sum` change (pure stdlib).**

### Known Gotchas of our codebase & Go stdlib

```go
// GOTCHA #1 — DO NOT use filepath.Glob with "**". Go's path/filepath does NOT
// support "**" (it behaves like "*" — single level only). VERIFIED
// (research/verified_facts.md §1): Glob("skills/**/SKILL.md") returns 0 matches
// for skills/foo/bar/SKILL.md. The PRD item description lists Glob as an option;
// it is WRONG for Go. Use filepath.WalkDir with a sentinel error (hasSkillMD).
//
//   RIGHT: filepath.WalkDir(dir, func(...) error { ... return errSkillMDFound })
//   WRONG: filepath.Glob(filepath.Join(dir, "**", "SKILL.md"))

// GOTCHA #2 — WalkDir early-exit: returning ANY non-nil error from the callback
// STOPS the walk. So hasSkillMD returns a sentinel (errSkillMDFound) the moment
// it sees a SKILL.md; the call site swallows it with _=. Do NOT walk the whole
// tree. (Verified §2.)

// GOTCHA #3 — Ascent loop terminates when filepath.Dir(cur) == cur. At the root,
// filepath.Dir("/") == "/" (NOT ""), so parent==cur breaks the loop. VERIFIED §3.
// Do NOT compare against "" or use a depth counter.

// GOTCHA #4 — An ancestor with an EMPTY skills/ dir (no SKILL.md) must NOT win;
// ascent must CONTINUE to higher ancestors. PRD §8.3 qualifies the match with
// "at least one SKILL.md". VERIFIED §4. The implementation achieves this by only
// returning when hasSkillMD(candidate) is true; an empty skills dir falls through
// the inner if and the loop ascends. Do NOT add `return "", false` when a skills
// dir exists but is empty — that would short-circuit and miss higher real stores.

// GOTCHA #5 — Check start (cwd) FIRST, not start's parent. For `go run` from the
// repo root, cwd IS the repo, so cwd/skills must be the first candidate. The loop
// body checks Join(cur,"skills") with cur==start before the first Dir(). Do NOT
// pre-advance cur to filepath.Dir(start).

// GOTCHA #6 — findWalkUp matches the LOCKED (dir, src, found) shape (no error
// return). The work-item contract loosely says rule 3 should "return (dir,
// SourceWalkUp, nil)"; the "nil" describes the error-free success case. Only
// Find() returns an error, and only on total miss. Do NOT add an error return to
// findWalkUp "for symmetry with Find()" — it would break the uniform chaining.

// GOTCHA #7 — os.Getwd() cannot be parameterized, so the testable ascent logic
// lives in findWalkUpAncestor(start). findWalkUp() is a thin entry: os.Getwd ->
// findWalkUpAncestor. Test findWalkUp via t.Chdir (Go 1.24+); test the ascent
// logic via findWalkUpAncestor with t.TempDir trees. (Verified §5.) Same
// factoring rationale as S2's resolveSiblingFromExe.

// GOTCHA #8 — findSibling() DETERMINISTICALLY MISSES inside `go test` (the test
// binary is in /tmp/go-buildXXX with no sibling skills/). VERIFIED §7. This means
// Find() (env unset + t.Chdir into a temp tree with an ancestor skills) reliably
// reaches rule 3 and returns SourceWalkUp — the Find() rule-3 test is deterministic.

// GOTCHA #9 — Two new imports in skillsdir.go: "errors" (errors.New for ErrNotFound)
// and "io/fs" (fs.DirEntry for the WalkDir callback param). The block becomes:
//   import ("errors"; "io/fs"; "os"; "path/filepath")  // gofmt-sorted
// Do NOT add "fmt" (the ErrNotFound message is a static string literal, no Sprintf).

// GOTCHA #10 — Test file gains "errors" (errors.Is) and "strings" (strings.Contains
// for the ErrNotFound message). t.Chdir and t.Setenv/unsetEnvVar all forbid
// t.Parallel — do NOT add t.Parallel to any rule-3/Find test.

// GOTCHA #11 — ALL findWalkUp/Find-rule-3 tests MUST t.Chdir into their OWN temp
// tree. Do NOT assert findWalkUp behavior against the real repo cwd: there is no
// skills/ today (so it'd be found=false), but P1.M6.T12 adds skills/example/,
// which would flip it to found=true and break the test. Hermetic temp trees are
// forward-compatible. (Verified §6.)

// GOTCHA #12 — ErrNotFound is EXPORTED (capital E) so tests can errors.Is it and
// main can print err.Error() directly. Its message is user-facing (printed to
// stderr by main in P1.M1.T3); do not wrap or prefix it before printing.
```

---

## Implementation Blueprint

### Data model — no new types

This subtask adds **no new types**. It reuses S1's `Source`/`SourceWalkUp` and
adds one exported variable (`ErrNotFound`) plus four functions. The only
"model" is the chained `(dir, src, err)` contract `Find()` exposes.

### File 1 — `internal/skillsdir/skillsdir.go`

**EDIT 1 — grow the import block.** Find the exact existing block (it is
unchanged from S1; S2 adds no imports) and replace it:

```go
// OLD (currently in the file):
import (
	"os"
	"path/filepath"
)
```
```go
// NEW:
import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)
```
(The new order is gofmt-sorted alphabetically: `errors` < `io/fs` < `os` < `path/filepath`.)

**EDIT 2 — append rule 3 + Find().** Append the following block at the END of
the file, **after** S2's `resolveSiblingFromExe` (the current last function).
Do not touch any existing code.

```go
// ---------------------------------------------------------------------------
// Rule 3 — walk up from the current working directory (PRD §8.3).
//
// Used for `go run` / dev, where the binary lives in a temp build dir and
// rules 1-2 both miss. The first ancestor (including cwd) whose skills/ subdir
// contains at least one SKILL.md (at any depth) wins.
// ---------------------------------------------------------------------------

// errSkillMDFound is a sentinel error used to short-circuit filepath.WalkDir as
// soon as the first SKILL.md is found, so hasSkillMD does not walk the entire
// tree. Returning any non-nil error from a WalkDir callback stops the walk.
var errSkillMDFound = errors.New("SKILL.md found")

// hasSkillMD reports whether dir contains at least one SKILL.md at any depth.
// It walks the tree under dir but returns true as soon as it finds one (early
// exit via the errSkillMDFound sentinel). PRD §8.3 requires "at least one
// SKILL.md" — a skills/ dir with none does not count.
//
// NOTE: filepath.Glob with a "**" pattern is intentionally NOT used: Go's
// path/filepath does not support "**" (it behaves like single-level "*"), so
// Glob("skills/**/SKILL.md") matches nothing for a nested file. WalkDir is the
// correct stdlib tool and recurses to arbitrary depth.
func hasSkillMD(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entry, keep walking
		}
		if !d.IsDir() && d.Name() == "SKILL.md" {
			found = true
			return errSkillMDFound // stop the walk
		}
		return nil
	})
	return found
}

// findWalkUpAncestor implements the ascent core of rule 3, factored out so it
// can be tested with an arbitrary start directory (os.Getwd itself cannot be
// controlled without chdir; findWalkUpAncestor takes start as a parameter).
//
// It checks start, then each ancestor, for a skills/ subdir that contains at
// least one SKILL.md. A skills/ dir that exists but has no SKILL.md is skipped
// and ascent continues (PRD §8.3 qualifies the match with "at least one
// SKILL.md"). Ascent stops at the filesystem root, where filepath.Dir(root)
// equals root.
func findWalkUpAncestor(start string) (dir string, found bool) {
	cur := filepath.Clean(start)
	for {
		candidate := filepath.Join(cur, "skills")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			if hasSkillMD(candidate) {
				return candidate, true
			}
			// skills/ exists here but has no SKILL.md -> keep ascending.
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", false // reached filesystem root, no match
		}
		cur = parent
	}
}

// findWalkUp implements PRD §8 rule 3 — ascend from the current working
// directory and return the first ancestor whose skills/ subdir contains at
// least one SKILL.md. This is the rule that makes `go run` work when the binary
// is in a temp build dir and rules 1-2 miss.
//
// Returns (candidate, SourceWalkUp, true) on a hit; ("", 0, false) otherwise so
// Find() can return ErrNotFound. Never errors (matches the locked per-rule shape).
func findWalkUp() (dir string, src Source, found bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", 0, false // cwd unresolvable -> rule misses
	}
	d, ok := findWalkUpAncestor(cwd)
	if !ok {
		return "", 0, false
	}
	return d, SourceWalkUp, true
}

// ---------------------------------------------------------------------------
// Find — the public entry point (PRD §8 priority order).
// ---------------------------------------------------------------------------

// ErrNotFound is returned by Find when all three §8 rules miss. Its message is
// the user-facing one-line fix (PRD §8.4 / §6.4): main prints it to stderr and
// exits 1. Print it verbatim (err.Error()); do not wrap or prefix it.
var ErrNotFound = errors.New("could not locate the skills directory: set $SKPP_SKILLS_DIR, cd into the skpp repo, or reinstall skpp")

// Find locates the skills directory per PRD §8 priority order:
//
//  1. SKPP_SKILLS_DIR env var (rule 1, findEnv).
//  2. Sibling of the running binary, symlink-aware (rule 2, findSibling).
//  3. Walk up from cwd (rule 3, findWalkUp).
//
// The first rule to hit wins and Find returns (absDir, src, nil). If all three
// miss it returns ("", 0, ErrNotFound); the caller (main) prints the error to
// stderr and exits 1.
func Find() (dir string, src Source, err error) {
	if d, s, ok := findEnv(); ok {
		return d, s, nil
	}
	if d, s, ok := findSibling(); ok {
		return d, s, nil
	}
	if d, s, ok := findWalkUp(); ok {
		return d, s, nil
	}
	return "", 0, ErrNotFound
}
```

### File 2 — `internal/skillsdir/skillsdir_test.go`

**EDIT 1 — grow the import block.** The current block (S1; S2 adds none):

```go
// OLD (currently in the file):
import (
	"os"
	"path/filepath"
	"testing"
)
```
```go
// NEW:
import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```
(gofmt-sorted: `errors` < `os` < `path/filepath` < `strings` < `testing`.)

**EDIT 2 — append rule-3 + Find() tests.** Append at the END of the file, after
S2's last test. Do not touch existing tests.

```go
// ---------------------------------------------------------------------------
// Rule 3 tests (walk-up) + Find() combiner tests (P1.M1.T2.S3).
// ---------------------------------------------------------------------------

// makeSkill creates <dir>/skills/<tag>/SKILL.md and returns the skills dir.
func makeSkill(t *testing.T, dir, tag string) string {
	t.Helper()
	skills := filepath.Join(dir, "skills")
	skillDir := filepath.Join(skills, tag)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", skillDir, err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# skill\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	return skills
}

// --- hasSkillMD ---

func TestHasSkillMDFoundNested(t *testing.T) {
	skills := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(filepath.Join(skills, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "a", "b", "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasSkillMD(skills) {
		t.Errorf("hasSkillMD(nested SKILL.md): got false; want true (WalkDir recurses)")
	}
}

func TestHasSkillMDFoundShallow(t *testing.T) {
	skills := makeSkill(t, t.TempDir(), "foo")
	if !hasSkillMD(skills) {
		t.Errorf("hasSkillMD(shallow SKILL.md): got false; want true")
	}
}

func TestHasSkillMDEmpty(t *testing.T) {
	skills := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if hasSkillMD(skills) {
		t.Errorf("hasSkillMD(empty skills): got true; want false")
	}
}

func TestHasSkillMDOnlyNonSkillFiles(t *testing.T) {
	skills := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasSkillMD(skills) {
		t.Errorf("hasSkillMD(only README.md): got true; want false (name must be SKILL.md)")
	}
}

// --- findWalkUpAncestor (the testable core; no cwd dependency) ---

// Rule 3: start IS the repo -> skills at start wins (cwd itself is checked first).
func TestFindWalkUpAncestorAtStart(t *testing.T) {
	root := t.TempDir()
	skills := makeSkill(t, root, "foo")
	got, found := findWalkUpAncestor(root)
	if !found {
		t.Fatalf("findWalkUpAncestor(start=repo): found=false; want true")
	}
	if got != skills {
		t.Errorf("findWalkUpAncestor: dir=%q; want %q", got, skills)
	}
}

// Rule 3: skills is several levels up -> ascent finds it.
func TestFindWalkUpAncestorDeep(t *testing.T) {
	root := t.TempDir()
	skills := makeSkill(t, root, "bar")
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	got, found := findWalkUpAncestor(deep)
	if !found {
		t.Fatalf("findWalkUpAncestor(deep): found=false; want true")
	}
	if got != skills {
		t.Errorf("findWalkUpAncestor(deep): dir=%q; want %q", got, skills)
	}
}

// Rule 3: a nested SKILL.md (skills/x/y/SKILL.md) counts (hasSkillMD recurses).
func TestFindWalkUpAncestorNestedSkillMD(t *testing.T) {
	root := t.TempDir()
	skills := filepath.Join(root, "skills")
	if err := os.MkdirAll(filepath.Join(skills, "x", "y"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "x", "y", "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, found := findWalkUpAncestor(root)
	if !found || got != skills {
		t.Errorf("findWalkUpAncestor(nested): found=%v dir=%q; want true %q", found, got, skills)
	}
}

// Rule 3 CONTRACT: a skills/ dir that exists but has NO SKILL.md is SKIPPED and
// ascent continues to a higher ancestor that DOES have one. PRD §8.3 qualifies
// the match with "at least one SKILL.md".
func TestFindWalkUpAncestorSkipsEmptyAndContinues(t *testing.T) {
	root := t.TempDir()
	// root/a/skills = EMPTY (no SKILL.md); root/skills/foo/SKILL.md = real.
	if err := os.MkdirAll(filepath.Join(root, "a", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	realSkills := makeSkill(t, root, "foo")
	start := filepath.Join(root, "a", "sub")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}
	got, found := findWalkUpAncestor(start)
	if !found {
		t.Fatalf("findWalkUpAncestor(skip-empty): found=false; want true")
	}
	if got != realSkills {
		t.Errorf("findWalkUpAncestor(skip-empty): dir=%q; want the higher real skills %q", got, realSkills)
	}
}

// Rule 3: no skills anywhere up to root -> miss.
func TestFindWalkUpAncestorNoSkills(t *testing.T) {
	if _, found := findWalkUpAncestor(t.TempDir()); found {
		t.Errorf("findWalkUpAncestor(no skills): got found=true; want false")
	}
}

// Rule 3: a 'skills' entry that is a regular FILE is skipped (IsDir guard).
func TestFindWalkUpAncestorSkillsIsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "skills"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, found := findWalkUpAncestor(root); found {
		t.Errorf("findWalkUpAncestor(skills is file): got found=true; want false")
	}
}

// --- findWalkUp (the rule-3 entry; os.Getwd exercised via t.Chdir) ---

// t.Chdir (Go 1.24+) changes cwd for the test and restores it on cleanup, so
// findWalkUp (which calls os.Getwd) is testable without global cwd pollution.

// Rule 3 via findWalkUp: chdir into a subdir of a temp repo and confirm walk-up
// resolves to the repo's skills/, returning SourceWalkUp.
func TestFindWalkUpFindsAncestor(t *testing.T) {
	root := t.TempDir()
	skills := makeSkill(t, root, "example")
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	got, src, found := findWalkUp()
	if !found {
		t.Fatalf("findWalkUp(): found=false; want true")
	}
	if src != SourceWalkUp {
		t.Errorf("findWalkUp(): src=%v; want SourceWalkUp", src)
	}
	if got != skills {
		t.Errorf("findWalkUp(): dir=%q; want %q", got, skills)
	}
}

// --- Find (the public combiner) ---

// Find: rule 1 wins when SKPP_SKILLS_DIR is set to an existing dir.
func TestFindRuleEnvWins(t *testing.T) {
	unsetEnvVar(t)
	dir := t.TempDir()
	t.Setenv(envVar, dir)
	got, src, err := Find()
	if err != nil {
		t.Fatalf("Find() env set: err=%v; want nil", err)
	}
	if src != SourceEnv {
		t.Errorf("Find() env set: src=%v; want SourceEnv", src)
	}
	if want := filepath.Clean(dir); got != want {
		t.Errorf("Find() env set: dir=%q; want %q", got, want)
	}
}

// Find: rule 3 wins when env is unset and cwd has an ancestor skills/ with a
// SKILL.md. (findSibling deterministically misses in a test because the test
// binary runs from a temp build dir with no sibling skills/.)
func TestFindRuleWalkUpWins(t *testing.T) {
	unsetEnvVar(t)
	root := t.TempDir()
	skills := makeSkill(t, root, "example")
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	got, src, err := Find()
	if err != nil {
		t.Fatalf("Find() walk-up: err=%v; want nil", err)
	}
	if src != SourceWalkUp {
		t.Errorf("Find() walk-up: src=%v; want SourceWalkUp", src)
	}
	if got != skills {
		t.Errorf("Find() walk-up: dir=%q; want %q", got, skills)
	}
}

// Find: all three rules miss -> ErrNotFound. (chdir into an empty temp dir; the
// walk ascends to /, which has no skills/ on this host — verified hermetic.)
func TestFindAllMissReturnsErrNotFound(t *testing.T) {
	unsetEnvVar(t)
	t.Chdir(t.TempDir())
	got, src, err := Find()
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find() all-miss: err=%v; want ErrNotFound", err)
	}
	if got != "" || src != 0 {
		t.Errorf("Find() all-miss: got=%q src=%v; want \"\" and 0", got, src)
	}
}

// ErrNotFound message carries the user-facing one-line fix (PRD §8.4 / §6.4).
func TestErrNotFoundMessageHasFix(t *testing.T) {
	msg := ErrNotFound.Error()
	for _, want := range []string{"SKPP_SKILLS_DIR", "cd", "reinstall"} {
		if !strings.Contains(msg, want) {
			t.Errorf("ErrNotFound message %q missing substring %q", msg, want)
		}
	}
}
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 0: PRECONDITION — confirm S1 + S2 are on disk and green
  - COMMAND: cd /home/dustin/projects/skpp
  - COMMAND: grep -qE 'func findEnv\(\) \(dir string, src Source, found bool\)' internal/skillsdir/skillsdir.go
  - COMMAND: grep -qE 'func findSibling\(\) \(dir string, src Source, found bool\)' internal/skillsdir/skillsdir.go
  - COMMAND: grep -qE 'func resolveSiblingFromExe\(exe string\) \(dir string, found bool\)' internal/skillsdir/skillsdir.go
  - COMMAND: go test ./internal/skillsdir/ >/dev/null 2>&1 && echo "S1+S2 green" || echo "NOT green"
  - EXPECT: all three symbols exist AND tests pass. If findSibling/resolveSiblingFromExe
            are MISSING, S2 has NOT landed — STOP and let S2 land first (do not
            recreate S2's code; Find() calls findSibling).

Task 1: EDIT the import block of internal/skillsdir/skillsdir.go
  - EDIT: replace the 3-line import block (os + path/filepath) with the 5-line
          block adding "errors" + "io/fs" (exact old/new text in Blueprint EDIT 1).
  - WHY: ErrNotFound needs errors.New; hasSkillMD's WalkDir callback needs fs.DirEntry.
  - DO NOT: add "fmt" (ErrNotFound is a static string, no Sprintf).

Task 2: APPEND rule-3 + Find() to internal/skillsdir/skillsdir.go
  - EDIT: append the Blueprint EDIT 2 code at the END of the file, after S2's
          resolveSiblingFromExe (the current last function).
  - NAMING: errSkillMDFound (sentinel), hasSkillMD, findWalkUpAncestor, findWalkUp
            (all unexported), ErrNotFound + Find (exported).
  - GOTCHA: use filepath.WalkDir, NOT filepath.Glob with "**"; check start FIRST;
            skip empty skills dirs and continue ascending; terminate on parent==cur.

Task 3: EDIT the import block of internal/skillsdir/skillsdir_test.go
  - EDIT: replace the import block (os, path/filepath, testing) with the version
          adding "errors" + "strings" (exact old/new text in Blueprint EDIT 1).

Task 4: APPEND rule-3 + Find() tests to internal/skillsdir/skillsdir_test.go
  - EDIT: append makeSkill helper + all tests from Blueprint EDIT 2 at the END of
          the file, after S2's last test.
  - COVERAGE: hasSkillMD (nested, shallow, empty, only-non-skill-files);
             findWalkUpAncestor (at-start, deep, nested-skillmd, skip-empty-and-
             continue, no-skills, skills-is-file); findWalkUp (via t.Chdir);
             Find (rule-1 win, rule-3 win, all-miss ErrNotFound); ErrNotFound message.
  - NAMING: TestHasSkillMD{FoundNested,FoundShallow,Empty,OnlyNonSkillFiles},
            TestFindWalkUpAncestor{AtStart,Deep,NestedSkillMD,SkipsEmptyAndContinues,
            NoSkills,SkillsIsFile}, TestFindWalkUpFindsAncestor,
            TestFind{RuleEnvWins,RuleWalkUpWins,AllMissReturnsErrNotFound},
            TestErrNotFoundMessageHasFix.
  - GOTCHA: all rule-3/Find tests t.Chdir into their OWN temp tree (hermetic,
            forward-compatible); NO t.Parallel() (t.Chdir/t.Setenv forbid it).

Task 5: FORMAT + VET + BUILD + TEST (validation gates — run in order)
  - COMMAND: gofmt -w internal/skillsdir/        # format in place
  - COMMAND: gofmt -l internal/skillsdir/        # MUST print nothing
  - COMMAND: go vet ./internal/skillsdir/        # MUST be clean
  - COMMAND: go build ./...                       # exit 0
  - COMMAND: go test ./internal/skillsdir/ -v     # all S1 + S2 + S3 tests PASS
  - EXPECT: zero errors, zero vet findings, gofmt silent.

Task 6: SCOPE BOUNDARY CHECK (the Level 3 block in Validation Loop)
  - COMMAND: the Level 3 block below.
  - EXPECT: Find/ErrNotFound/findWalkUp/hasSkillMD present; no Glob('**'); imports
            grew by exactly errors+io/fs; go.mod/go.sum/PRD.md unchanged; no new
            packages/files.
```

### Implementation Patterns & Key Details

```go
// PATTERN: the rule-3 two-function split (thin entry + testable core), mirroring S2.
//   func findWalkUp() (dir string, src Source, found bool)        // os.Getwd -> core
//   func findWalkUpAncestor(start string) (dir string, found bool) // ascent logic
// WHY: os.Getwd() cannot be parameterized, so the ascent logic must live in a
//      parameterized helper to be testable. findWalkUp matches S1/S2's locked
//      (dir,src,found) shape; the core drops src (the entry always wins as
//      SourceWalkUp). Same rationale as S2's resolveSiblingFromExe.

// PATTERN: WalkDir early-exit via sentinel error (the only correct stdlib "any
// SKILL.md at any depth" check).
//   var errSkillMDFound = errors.New("SKILL.md found")
//   _ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
//       if err != nil { return nil }
//       if !d.IsDir() && d.Name() == "SKILL.md" { found = true; return errSkillMDFound }
//       return nil
//   })
// WHY: filepath.Glob does NOT support "**" (verified). Returning any non-nil error
//      stops WalkDir, so we never walk the whole tree. Requires import "io/fs".

// PATTERN: the ascent loop with root-termination + skip-empty.
//   cur := filepath.Clean(start)
//   for {
//       candidate := filepath.Join(cur, "skills")
//       if info, err := os.Stat(candidate); err == nil && info.IsDir() {
//           if hasSkillMD(candidate) { return candidate, true }   // WIN
//           // empty skills dir -> fall through, keep ascending
//       }
//       parent := filepath.Dir(cur)
//       if parent == cur { return "", false }                     // reached root
//       cur = parent
//   }
// WHY: checks start FIRST (so `go run` from repo root finds cwd/skills); skips
//      empty skills dirs (PRD §8.3 "at least one SKILL.md"); terminates because
//      filepath.Dir("/") == "/".

// PATTERN: the dumb combiner (src self-describes the winner).
//   func Find() (dir string, src Source, err error) {
//       if d, s, ok := findEnv(); ok { return d, s, nil }
//       if d, s, ok := findSibling(); ok { return d, s, nil }
//       if d, s, ok := findWalkUp(); ok { return d, s, nil }
//       return "", 0, ErrNotFound
//   }
// WHY: each rule returns (dir, src, found); the first found==true wins and carries
//      its own Source. Find() needs no if/else on Source — it just returns what the
//      winning helper produced. Only Find() returns an error.

// PATTERN: t.Chdir for cwd-dependent tests (Go 1.24+).
//   root := t.TempDir(); makeSkill(t, root, "x"); t.Chdir(filepath.Join(root, "sub"))
//   _, src, _ := findWalkUp()   // now resolves to root/skills, SourceWalkUp
// WHY: os.Getwd is process-global; t.Chdir changes it for the test and restores on
//      cleanup (the cwd analog of t.Setenv). Forbids t.Parallel().
```

### Integration Points

```yaml
PACKAGE BOUNDARIES (after S3):
  - import path: "github.com/dabstractor/skpp/internal/skillsdir"
  - exported (for main/discover to use): Source, SourceEnv, SourceSibling,
    SourceWalkUp, (Source).String, ErrNotFound, Find
  - unexported (internal): envVar, findEnv, findSibling, resolveSiblingFromExe,
    errSkillMDFound, hasSkillMD, findWalkUpAncestor, findWalkUp

DOWNSTREAM CONSUMERS (what relies on this subtask's output):
  - P1.M1.T3 (main.go --path): dir, src, err := skillsdir.Find(); on err,
    fmt.Fprintln(os.Stderr, err); os.Exit(1); else print dir (and src via --path).
    PRD §13 acceptance: `test "$(./skpp --path)" = "$PWD/skills"` (rule 2 wins for
    a repo-root binary) and the all-miss error path.
  - P1.M2.T5 (discover.Index): takes Find()'s absDir as its walk root.

NO CHANGES TO:
  - go.mod / go.sum (pure stdlib; no new deps)
  - PRD.md (read-only)
  - the Source type / findEnv / findSibling / resolveSiblingFromExe (S1/S2-owned)
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

# Run all skillsdir tests verbosely — S1 + S2 + S3 functions must all PASS
go test ./internal/skillsdir/ -v

# Explicit assertions the run must satisfy (rule-3 + Find):
go test ./internal/skillsdir/ -run 'TestHasSkillMD|TestFindWalkUp|TestFind|TestErrNotFound' -v \
  || { echo "FAIL: rule-3 + Find tests"; exit 1; }

# The skip-empty-and-continue CONTRACT test must PASS (the crux §8.3 behavior)
go test ./internal/skillsdir/ -run TestFindWalkUpAncestorSkipsEmptyAndContinues -v \
  | grep -E '--- PASS:.*TestFindWalkUpAncestorSkipsEmptyAndContinues' \
  || { echo "FAIL: skip-empty contract test did not pass"; exit 1; }

# Find() rule-3 win + all-miss ErrNotFound
go test ./internal/skillsdir/ -run 'TestFindRuleWalkUpWins|TestFindAllMissReturnsErrNotFound' -v \
  || { echo "FAIL: Find combiner tests"; exit 1; }

# Full module still green
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Scope-boundary & contract check

```bash
cd /home/dustin/projects/skpp

# S3 symbols now EXIST with the contracted signatures
grep -qE 'func findWalkUp\(\) \(dir string, src Source, found bool\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: findWalkUp missing or wrong signature"; exit 1; }
grep -qE 'func findWalkUpAncestor\(start string\) \(dir string, found bool\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: findWalkUpAncestor missing or wrong signature"; exit 1; }
grep -qE 'func hasSkillMD\(dir string\) bool' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: hasSkillMD missing or wrong signature"; exit 1; }
grep -qE 'func Find\(\) \(dir string, src Source, err error\)' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: Find missing or wrong signature"; exit 1; }
grep -qE 'var ErrNotFound = errors\.New\(' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: ErrNotFound sentinel missing"; exit 1; }

# hasSkillMD uses WalkDir, NOT filepath.Glob with "**" (Go does not support "**")
grep -q 'filepath.WalkDir(' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: WalkDir missing"; exit 1; }
! grep -nE 'filepath\.Glob\([^)]*\*\*' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: found Glob with '**' — Go does not support it; use WalkDir"; exit 1; }

# imports grew by exactly "errors" + "io/fs" (no fmt, no extras)
grep -A6 '^import (' internal/skillsdir/skillsdir.go | grep -q '"errors"'   || { echo "FAIL: errors import missing"; exit 1; }
grep -A6 '^import (' internal/skillsdir/skillsdir.go | grep -q '"io/fs"'    || { echo "FAIL: io/fs import missing"; exit 1; }
! grep -A8 '^import (' internal/skillsdir/skillsdir.go | grep -q '"fmt"' \
  || { echo "FAIL: fmt import added (not needed)"; exit 1; }

# ErrNotFound message carries the fix phrase
grep -q 'SKPP_SKILLS_DIR' internal/skillsdir/skillsdir.go \
  || { echo "FAIL: ErrNotFound message missing fix phrase"; exit 1; }

# MUST NOT have touched go.mod / go.sum / PRD.md
git diff --quiet go.mod   || { echo "FAIL: go.mod changed (should be untouched)"; exit 1; }
git diff --quiet go.sum   || { echo "FAIL: go.sum changed (should be untouched)"; exit 1; }
git diff --quiet PRD.md   || { echo "FAIL: PRD.md changed (read-only)"; exit 1; }

# MUST NOT have created main.go or other packages (later subtasks)
test ! -e main.go           || { echo "FAIL: main.go must not exist (T3)"; exit 1; }
test ! -d internal/discover || { echo "FAIL: discover/ must not exist (M2)"; exit 1; }
test ! -d internal/resolve  || { echo "FAIL: resolve/ must not exist (M3)"; exit 1; }
test ! -d internal/ui       || { echo "FAIL: ui/ must not exist (M2)"; exit 1; }

echo "Level 3 PASS (scope + contract respected)"
```

### Level 4: Downstream-readiness smoke test

Prove `Find()` is importable and behaves as `main.go` (T3) and `discover.Index`
(M2.T5) will rely on — without depending on those existing yet.

```bash
cd /home/dustin/projects/skpp

# Find() is the public entrypoint; SourceWalkUp labels the rule-3 winner;
# ErrNotFound is the all-miss sentinel with the user-facing fix.
mkdir -p /tmp/skpp-find-check && cat > /tmp/skpp-find-check/main.go <<'EOF'
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dabstractor/skpp/internal/skillsdir"
)

func main() {
	// Simulate: no env, no sibling (this scratch binary runs from /tmp), no
	// ancestor skills -> Find must return ErrNotFound.
	dir, src, err := skillsdir.Find()
	if errors.Is(err, skillsdir.ErrNotFound) {
		fmt.Fprintln(os.Stderr, err) // the one-line fix, verbatim
		fmt.Println("SENTINEL_OK", skillsdir.SourceWalkUp.String())
		return
	}
	fmt.Println("FOUND", dir, src.String())
}
EOF
cat > /tmp/skpp-find-check/go.mod <<EOF
module skpp-find-check
go 1.25
require github.com/dabstractor/skpp v0.0.0
replace github.com/dabstractor/skpp => $(pwd)
EOF
( cd /tmp/skpp-find-check && go run . 2>&1 ) \
  | grep -E 'could not locate the skills directory.*SKPP_SKILLS_DIR|SENTINEL_OK ancestor of cwd' \
  || { echo "FAIL: Find()/ErrNotFound/SourceWalkUp not usable downstream"; rm -rf /tmp/skpp-find-check; exit 1; }
rm -rf /tmp/skpp-find-check

# Live rule-3 proof via the in-package test (TestFindRuleWalkUpWins) is already
# covered in Level 2; here we only confirm the public API surface compiles + the
# sentinel/label are exported and printable.
echo "Level 4 PASS (Find/ErrNotFound/SourceWalkUp usable by main.go + discover)"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` silent, `go vet ./internal/skillsdir/` clean, `go build ./...` exit 0
- [ ] Level 2 PASS — `go test ./internal/skillsdir/ -v`: all S1 + S2 + S3 tests pass; skip-empty contract + Find combiner tests pass
- [ ] Level 3 PASS — S3 symbols present with correct signatures; WalkDir used (no Glob `**`); imports grew by errors+io/fs; go.mod/go.sum/PRD.md unchanged; no new packages
- [ ] Level 4 PASS — `Find()`/`ErrNotFound`/`SourceWalkUp` importable and usable downstream

### Feature Validation
- [ ] `findWalkUp()` returns `(candidate, SourceWalkUp, true)` when an ancestor (incl. cwd) has a `skills/` with ≥1 `SKILL.md`
- [ ] `findWalkUp()` returns `found=false` when `os.Getwd()` errors
- [ ] `findWalkUpAncestor` checks `start` first, ascends to root, terminates on `parent==cur`
- [ ] `findWalkUpAncestor` **skips** an ancestor whose `skills/` dir has no `SKILL.md` and continues ascending (contract test)
- [ ] `hasSkillMD` finds a `SKILL.md` at any depth (WalkDir) and returns false for empty / non-skill-file dirs
- [ ] `Find()` returns rule 1's hit when env is set; rule 3's hit when env unset + cwd under an ancestor skills tree; `ErrNotFound` when all miss
- [ ] `errors.Is(err, ErrNotFound)` is true for the all-miss case
- [ ] `ErrNotFound.Error()` contains `SKPP_SKILLS_DIR`, `cd`, and `reinstall`

### Code Quality / Convention Validation
- [ ] Both helpers + `Find()` + `ErrNotFound` appended to the EXISTING file (no new files/dirs)
- [ ] Tests appended to the EXISTING test file; existing S1/S2 tests untouched
- [ ] Import block grown by exactly `errors` + `io/fs` (skillsdir.go) and `errors` + `strings` (test); no `fmt`
- [ ] Helper names match S1/S2's precedent (`findWalkUp` ← `findEnv`/`findSibling`; `findWalkUpAncestor` is the testable core, paralleling `resolveSiblingFromExe`)
- [ ] All rule-3/Find tests `t.Chdir` into hermetic temp trees (forward-compatible); no `t.Parallel()`

### Scope Discipline
- [ ] Did NOT modify the `Source` type, `findEnv`, `findSibling`, or `resolveSiblingFromExe` (S1/S2-owned)
- [ ] Did NOT create `main.go`, `internal/discover`, `internal/resolve`, `internal/ui` (later milestones)
- [ ] Did NOT modify `go.mod` / `go.sum` (no new deps; pure stdlib)
- [ ] Did NOT modify `PRD.md` (read-only) or any `tasks.json` (orchestrator-owned)

---

## Anti-Patterns to Avoid

- ❌ **Don't use `filepath.Glob` with `**`.** Go's `path/filepath` does NOT support
  `**` (it acts like single-level `*`). Verified: `Glob("skills/**/SKILL.md")`
  returns 0 matches for a nested file. The PRD item description lists Glob as an
  option — it is wrong for Go. Use `filepath.WalkDir` with an early-exit sentinel.
  (Verified: `research/verified_facts.md §1`.)
- ❌ **Don't walk the whole tree in `hasSkillMD`.** Return the `errSkillMDFound`
  sentinel from the WalkDir callback the instant a `SKILL.md` is seen; returning
  any non-nil error stops the walk. (Verified §2.)
- ❌ **Don't short-circuit on an empty `skills/` dir.** PRD §8.3 qualifies the win
  with "at least one `SKILL.md`": an ancestor with a `skills/` dir but no
  `SKILL.md` must be skipped and ascent must continue to higher ancestors. Do NOT
  add `return "", false` when `os.Stat(skills)` succeeds but `hasSkillMD` is false
  — just fall through the inner `if` and loop. (Verified §4.)
- ❌ **Don't check `start`'s parent first.** For `go run` from the repo root, cwd
  IS the repo, so `cwd/skills` must be the first candidate. Check `Join(cur,
  "skills")` with `cur == start` before the first `filepath.Dir`. (Verified §5.)
- ❌ **Don't add an error return to `findWalkUp`.** It matches the locked
  `(dir, src, found)` shape (S1/S2). The work-item contract's loose "return (dir,
  SourceWalkUp, nil)" describes the error-free success case; `found==true` is the
  win signal. Only `Find()` returns an error, and only on total miss.
- ❌ **Don't add `"fmt"` to imports.** `ErrNotFound` is a static string literal via
  `errors.New`; no `Sprintf`. An unused import is a compile error.
- ❌ **Don't assert `findWalkUp` behavior against the real repo cwd.** There is no
  `skills/` today (so it'd be `found=false`), but P1.M6.T12 adds
  `skills/example/`, which would flip it to `found=true` and break the test.
  Always `t.Chdir` into a hermetic temp tree. (Verified §6.)
- ❌ **Don't call `t.Parallel()` in rule-3/Find tests.** `t.Chdir` and
  `t.Setenv`/`unsetEnvVar` mutate process-global state (cwd / env) and forbid
  parallelism. Go enforces this; keep the tests serial.
- ❌ **Don't touch S1/S2's code.** Append only. The `Source` type, `findEnv`,
  `findSibling`, and `resolveSiblingFromExe` are S1/S2-owned and already pass
  their tests. If S2 hasn't landed, STOP (Task 0) — don't recreate it.
- ❌ **Don't wrap or prefix `ErrNotFound` before printing.** Its message IS the
  user-facing one-line fix (PRD §6.4). `main` (T3) does `fmt.Fprintln(os.Stderr,
  err)` verbatim. The exported sentinel lets tests `errors.Is` it.

---

## Confidence Score

**10/10** — A pure-stdlib completion of an already-green package. The exact
source for every added symbol (imports, 4 functions, 1 sentinel, 1 error, and the
full test suite) is given verbatim, and every stdlib behavior it relies on was
**empirically verified** in the target Go 1.26.4 environment, including the two
critical, non-obvious points: (1) `filepath.Glob` does NOT support `**` (so
WalkDir is mandatory), and (2) the skip-empty-skills-and-continue §8.3 semantics.
The testing cruxes — `os.Getwd` is untestable without `t.Chdir`, and
`findSibling` deterministically misses in tests so `Find()` reaches rule 3 — are
both documented, justified, and pre-run. No new deps, no concurrency, no I/O
beyond stat/walk/getwd. The only residual (non-)risk — a future `skills/` at the
repo root — is explicitly excluded by making every rule-3 test `t.Chdir` into a
hermetic temp tree.
