# PRP — P1.M2.T5.S1: `internal/discover` — `Index()` walk + relTag normalization + sorting

> **Subtask:** P1.M2.T5.S1 — the Index walk (PRD §7.1). It is the FIRST subtask of
> T5 ("internal/discover — Index() walk"). It builds directly on **T4.S1**
> (`Frontmatter` + `ParseFrontmatter`, in `discover.go`) and **T4.S2** (`Skill` +
> `BuildSkill` + `toStringSlice`, in `skill.go`) — both landed & green — and feeds
> **T6** (`--list`), **T7** (`resolve`), **T9** (`--search`), and **T10** (`check`).
>
> **Scope:** create `internal/discover/index.go` (`package discover`) and its
> white-box test `internal/discover/index_test.go`. Implement `Index(skillsDir
> string) ([]Skill, error)`: `WalkDir` the skills dir, identify every directory
> that directly contains a `SKILL.md`, compute each skill's canonical tag
> (`RelTag`, OS separators normalized to `/`), parse frontmatter, build a `Skill`
> via `BuildSkill`, and return the slice sorted by `RelTag`.
>
> **SCOPE DECISION (authoritative — see verified_facts.md §8, §14, §15):** T4.S2's
> PRP pre-specified the T5 call site and ownership:
> *"T5 owns: the walk, relTag computation, sorting by RelTag, and the err policy
> (malformed YAML → surface to `check` M4). BuildSkill gives T5 a one-line call;
> T5 never touches frontmatter fields directly."* This subtask owns EXACTLY
> `Index()`. It does **NOT** touch `discover.go`/`discover_test.go` (S1-owned) or
> `skill.go`/`skill_test.go` (S2-owned), does **NOT** add `resolve`/`ui`/`main`
> code (later milestones), and does **NOT** create `skills/example/` (P1.M6.T12).
>
> **DEPENDENCY:** depends on S1's `Frontmatter`/`ParseFrontmatter` and S2's
> `Skill`/`BuildSkill` (same `package discover`, already landed). `internal/discover`
> stays a LEAF library — `index.go` imports ONLY stdlib (`errors`, `io/fs`, `os`,
> `path/filepath`, `sort`). It does NOT import `yaml.v3` (S1 does, in `discover.go`;
> the package shares it), so T5 adds **NO** module dependency → `go.mod`/`go.sum`
> are UNCHANGED.
>
> **NOTE (main.go):** M1.T3 (`main.go`/`main_test.go`) is landed and irrelevant
> here — `discover` is a leaf library. The wiring `skillsdir.Find() →
> discover.Index(dir)` lands in a LATER milestone (the `--list`/`<tag>` modes), NOT
> here. Do NOT touch `main.go`/`main_test.go`.

---

## Goal

**Feature Goal**: Produce the `[]Skill` index that every skpp read mode operates
on, by walking the on-disk skills tree and turning each `SKILL.md` into a typed
`Skill`. After T5, `discover.Index(dir)` is the single data source for `--list`
(T6), `resolve` (T7), `--search` (T9), and `check` (T10). The walk is manifest-free
(PRD §2.1): everything is inferred from disk — a skill is "any directory that
directly contains a `SKILL.md`" (PRD §7.1.2) — and the index is rebuilt every
invocation. `RelTag` is computed as `filepath.ToSlash(filepath.Rel(skillsDir,
skillDir))` so canonical tags are `writing/reddit` on every platform, and the
returned slice is sorted by `RelTag` for deterministic `--all`/`--list` output.

**Deliverable**: Two NEW files (no other files touched):
1. `internal/discover/index.go` — `package discover`; `func Index(skillsDir string)
   ([]Skill, error)`. Imports only `errors`/`io/fs`/`os`/`path/filepath`/`sort`.
2. `internal/discover/index_test.go` — `package discover` (white-box); 12 tests
   covering single/nested skills, sort order, no-frontmatter + malformed-YAML
   leniency, non-`SKILL.md` filtering, empty dir, missing root, root-is-file,
   nested-both-levels, the `relTag=="."` edge case, and relative-input→absolute-Dir.

**Success Definition**: `gofmt -l internal/discover/*.go` is silent; `go vet
./internal/discover/` is clean; `go build ./...` and `go test ./...` pass (S1's 12
+ S2's + T5's 12 new tests, all green); `go mod tidy` is a **no-op** (go.mod/go.sum
unchanged). `go doc ./internal/discover Index` shows the exported, documented
function. No touch to `discover.go`/`discover_test.go`/`skill.go`/`skill_test.go`/
`main*`/`skillsdir*`; no `resolve`/`ui`/`skills/`.

---

## Why

- This subtask **makes the index real.** S1 parses frontmatter; S2 defines the
  record; **T5 is the loop that produces `[]Skill` from the filesystem.** Without
  it there is nothing for `--list`/`resolve`/`--search`/`check` to read.
- It **locks the manifest-free contract.** PRD §2.1 forbids any index file; the
  catalog is rebuilt from disk every call. `Index()` is exactly that rebuild — a
  directory walk of a (small) tree, so latency is a non-issue.
- It **locks the two load-bearing contracts downstream code depends on.** (1)
  `RelTag` is the canonical tag (PRD §7.2 step 1) and must be `/`-normalized on
  every OS; (2) `Skill.Dir` is always absolute (PRD §6.1, §13). T5 computes both,
  in one place, so T7/T6/T9 never re-derive them.
- It **resolves the parse-error policy S2 deferred to T5.** `ParseFrontmatter`
  returns an error for malformed YAML. T5's decision (verified): **do not abort,
  do not propagate** — build a `HasFM=false` Skill so the skill is still resolvable
  by directory (PRD §7.1), and let `check` (M4) re-parse if it needs the exact YAML
  error. This keeps `Index` robust (one bad file can't break the whole catalog) and
  avoids rework (the re-parse is idempotent and cheap).
- It **keeps `go.mod` clean.** `index.go` imports only stdlib → `go.mod`/`go.sum`
  unchanged (unlike S1's indirect→direct yaml.v3 flip).

---

## What

One new function in the existing `package discover` (`internal/discover/`):

**`func Index(skillsDir string) ([]Skill, error)`** — walks `skillsDir` and returns
every skill in it, sorted by `RelTag`. Behavior (every clause verified — see
`research/verified_facts.md`):

1. **Abs first.** `root, err := filepath.Abs(skillsDir)` → guarantees every
   `Skill.Dir` is absolute (PRD §6.1/§13). No-op `Clean` on the canonical absolute
   input from `skillsdir.Find`.
2. **Stat-guard the root.** `os.Stat(root)`; propagate `err` (missing/unreadable);
   `!info.IsDir()` → `errors.New(root + ": not a directory")`. **This guard is
   mandatory** — without it a missing root is silently swallowed (see Gotcha #1).
3. **WalkDir, identify skills.** `filepath.WalkDir(root, fn)`. In `fn`: skip
   per-entry errors (`return nil`); act only on non-dir entries named exactly
   `SKILL.md`; `skillDir = filepath.Dir(path)`. Stray `README.md`/`SKILL.md.bak`
   and empty subdirs are ignored.
4. **relTag.** `relTag = filepath.ToSlash(filepath.Rel(root, skillDir))`. Nested
   skill → `writing/reddit`.
5. **Build.** `fm, _, _ := ParseFrontmatter(path)` (lenient — error ignored);
   `skills = append(skills, BuildSkill(skillDir, relTag, fm))`.
6. **Sort.** `sort.Slice` by `RelTag` (`<`), lexicographic, ascending.
7. **Return.** `nil` slice + `nil` error when the tree has no skills (callers use
   `len()`). Propagate `walkErr` if the walk itself failed (defensive; unreachable
   after the Stat guard).

### Success Criteria

- [ ] `internal/discover/index.go` is `package discover` with EXACTLY:
      `func Index(skillsDir string) ([]Skill, error)` (and its imports). No other
      exported symbol added.
- [ ] `index.go` imports ONLY `errors`, `io/fs`, `os`, `path/filepath`, `sort`
      (no `fmt`, no `yaml.v3`, no `strings`).
- [ ] `Index` on a tree with `skills/example` returns 1 Skill: `RelTag=="example"`,
      `Name=="example"`, `Dir` is absolute, `SourceFile == filepath.Join(Dir,
      "SKILL.md")` and EXISTS, `Keywords==[a b]`/`Category=="meta"`/`Aliases==[ex]`
      (end-to-end `[]any`→`[]string`).
- [ ] `Index` on `skills/writing/reddit/SKILL.md` yields `RelTag=="writing/reddit"`
      (separator normalized to `/`; no backslash).
- [ ] `Index` returns skills **sorted by `RelTag`** lexicographically
      (`apple < mango/beta < mango/fig < zebra`).
- [ ] `Index` on a no-`---`-block `SKILL.md` includes it: `HasFM==false`,
      `Name==""`, `RelTag` set (resolvable by directory, PRD §7.1).
- [ ] `Index` on a **malformed-YAML** `SKILL.md` does NOT abort and does NOT
      error: the skill is included with `HasFM==false`; a sibling good skill is
      also present (len == 2).
- [ ] `Index` ignores `README.md`, `SKILL.md.bak`, and files in subdirs without a
      `SKILL.md` (exactly the real skills are returned).
- [ ] `Index` on an empty dir returns `len==0`, `nil` error.
- [ ] `Index` on a **missing** path returns a non-nil error (the Stat guard).
- [ ] `Index` on a path that is a **regular file** returns a non-nil error.
- [ ] `Index` finds BOTH `writing/SKILL.md` and `writing/reddit/SKILL.md` when both
      exist (nested skills, PRD §7.1).
- [ ] A root-level `SKILL.md` (directly in the skills dir) yields `RelTag=="."`
      (the `filepath.Rel(root,root)` edge case) — documented, included.
- [ ] A **relative** `skillsDir` still yields absolute `Skill.Dir` (the `filepath.Abs`).
- [ ] `index_test.go` is white-box `package discover`; reuses `writeSkill`
      (`discover_test.go`) and `strEq` (`skill_test.go`) WITHOUT redefining them;
      defines ONE new helper `makeTree`; uses `t.Chdir` (Go 1.24+, go.mod is 1.25);
      NO testify, NO `t.Parallel()`.
- [ ] `gofmt -l` silent; `go vet ./internal/discover/` clean; `go build ./...` +
      `go test ./...` pass (S1's 12 + S2's + T5's 12 new); `go mod tidy` no-op;
      `discover.go`/`discover_test.go`/`skill.go`/`skill_test.go`/`main*`/
      `skillsdir*` untouched.

---

## All Needed Context

### Context Completeness Check

_Pass: the EXACT source for `index.go` and `index_test.go` is given verbatim in the
Implementation Blueprint (gofmt-clean, compiles & all 12 tests pass as-is — the
algorithm was compiled and run against a verbatim copy of the real `discover.go` +
`skill.go` in a throwaway `/tmp/skpp_index_verify` module during research). Every
load-bearing behavior was empirically verified (`research/verified_facts.md`): the
missing-root Stat-guard bug (Run 1) and fix (Run 2), per-entry skip, the
malformed-YAML leniency, relTag normalization, sorting, the `relTag=="."` edge
case, the relative-input→absolute-Dir guarantee, and symlink non-following. The
consumed contracts are S1's `discover.go` (`ParseFrontmatter`) and S2's `skill.go`
(`Skill`/`BuildSkill`), both read directly and landed green. An implementer who
knows Go but nothing about this repo can complete this in one pass from this
document._

### Documentation & References

```yaml
# MUST READ — this subtask's own empirical verification (every load-bearing decision)
- file: plan/001_fcde63e5bb60/P1M2T5S1/research/verified_facts.md
  why: "Proves (against a verbatim copy of the real discover.go + skill.go on
        go1.26.4, /tmp/skpp_index_verify): (1) signature Index(skillsDir)([]Skill,error).
        (2) *** A Stat-guard BEFORE WalkDir is MANDATORY *** — without it a missing
        root returns (nil,nil) because WalkDir feeds the root lstat error to the
        callback and `if err != nil {return nil}` swallows it. (3) filepath.Abs
        first -> Skill.Dir is ALWAYS absolute (PRD §6.1/§13). (4) relTag =
        ToSlash(Rel(root,skillDir)). (5) sort by RelTag lexicographic. (6) a skill
        = a dir directly containing SKILL.md. (7) nested skills at multiple levels
        all indexed. (8) malformed YAML does NOT abort and is NOT propagated ->
        HasFM=false Skill, still resolvable by dir; M4/check can re-parse
        SourceFile. (9) per-entry errors skipped. (10) empty dir -> (nil,nil).
        (11) WalkDir does NOT follow symlinked dirs (default). (12) root-level
        SKILL.md -> relTag='.'. (13) index.go imports ONLY stdlib -> no go.mod
        change. (14) index.go is a NEW file (S1 reserved Index for T5 in its own
        file). (15) index_test.go reuses writeSkill+strEq, defines makeTree only."
  critical: "Do NOT skip the os.Stat guard (Run 1 shows the missing-root bug).
             Do NOT propagate ParseFrontmatter's malformed-YAML error (aborting
             breaks the whole catalog on one bad file). Do NOT put Index in
             discover.go (S1's doc forbids it). Do NOT redefine writeSkill/strEq
             (same package -> compile error)."

# CONTRACT — the discover package design (the exact Index signature + data flow)
- file: plan/001_fcde63e5bb60/architecture/go_architecture.md
  why: "Locks `func Index(absSkillsDir string) ([]Skill, error)` (line 53) and the
        data flow `discover.Index(absSkillsDir) -> []Skill, error` (line 23) that
        resolve/ui consume. The 'relTag normalization' note (lines 114-118):
        filepath.Rel(skillsDir, skillDir) then filepath.ToSlash(...) so tags are
        always writing/reddit; canonical comparisons on the slash form. Also the
        nested-skills note and that discover is a LEAF library (resolve imports it
        as []discover.Skill)."
  section: "internal/discover", "Data flow", "relTag normalization"

# PREDECESSOR — S2's landed skill.go (Skill + BuildSkill, the constructor Index calls)
- file: internal/discover/skill.go
  why: "The Skill struct (9 fields, NO yaml tags) and BuildSkill(dir, relTag string,
        fm Frontmatter) Skill — the ONE call Index makes per skill. S2's doc spells
        out the T5 call site verbatim: `fm, _, err := ParseFrontmatter(...);
        // T5 decides how to surface err; s := BuildSkill(dir, relTag, fm)`. S2 also
        documents that BuildSkill is TOTAL (never errors/panics, even on
        Frontmatter{}), so Index can call it unconditionally after a failed parse.
        READ-ONLY — do not modify."
  pattern: "Index computes (skillDir, relTag) from the walk and delegates ALL field
            extraction to BuildSkill — it never reads frontmatter fields itself."
  gotcha: "BuildSkill expects an ABSOLUTE dir (its doc: 'Dir: absolute path'). Index
           abs-ifies the root first so every appended Skill.Dir is absolute."

# PREDECESSOR — S1's landed discover.go (ParseFrontmatter, the parser Index calls)
- file: internal/discover/discover.go
  why: "ParseFrontmatter(path) (fm Frontmatter, body string, err error). On a
        no-`---`-block file it returns (Frontmatter{HasFM:false}, wholeFile, nil);
        on malformed YAML it returns (Frontmatter{}, body, err). Index relies on
        the Frontmatter{}-on-error shape to build a HasFM=false Skill and keep
        walking. S1's package doc says Index is 'a LATER subtask — do not add it
        here' (=> new file index.go). READ-ONLY."
  gotcha: "ParseFrontmatter's malformed-YAML error MUST be swallowed by Index (per
           the policy), not returned — else one bad file aborts the whole catalog."

# PREDECESSOR RESEARCH — the S1/S2/T5 scope split (confirms T5 owns Index only)
- file: plan/001_fcde63e5bb60/P1M2T4S2/research/verified_facts.md
  why: "§12: S2 owns Skill+toStringSlice+BuildSkill ONLY; Index() is T5. §11:
        BuildSkill is the S1<->T5 seam and owns NO error policy — T5 handles
        ParseFrontmatter errors. Cross-checks the leniency model T5 inherits."

# PREDECESSOR PRP — the exact T5 contract S2 pre-specified (downstream extension points)
- file: plan/001_fcde63e5bb60/P1M2T4S2/PRP.md
  why: "S2's 'DOWNSTREAM EXTENSION POINTS' locked the T5 call site: WalkDir; each
        dir containing SKILL.md -> ParseFrontmatter(Join(dir,'SKILL.md')) ->
        BuildSkill(dir, relTag, fm); relTag = ToSlash(Rel(skillsDir, skillDir));
        sort by RelTag; err policy (malformed YAML -> check M4). This PRP
        implements exactly that."
  section: "Integration Points > DOWNSTREAM EXTENSION POINTS (P1.M2.T5.S1)"

# CONTRACT — how Index is consumed downstream (do not break these shapes)
- file: plan/001_fcde63e5bb60/architecture/go_architecture.md
  why: "resolve.Resolve(tag string, skills []discover.Skill) (Result, error) (line
        93) consumes Index's output; ui.Print* renders it. Index must return a
        plain []Skill (value slice), sorted, with absolute Dir + /-normalized
        RelTag — the shapes T7/T6/T9 match against."
  section: "internal/resolve"

# CONTRACT — the PRD sections this implements
- file: PRD.md
  why: "§7.1 discovery (walk recursively; a skill = a dir directly containing
        SKILL.md; nested skills count; capture dir/relTag/name/description/
        keywords/category/aliases). §7.2 RelTag is the canonical tag (step 1).
        §2.1 manifest-free (infer from disk; no index file). §6.1 --all 'sorted by
        tag'. §6.1 --list '1 if no skills found' (=> Index empty -> caller exits 1).
        §6.4/§8.4 skills-dir-unresolvable -> stderr + exit 1 (=> Index's root error
        path). READ-ONLY."
  critical: "§2.1: NO index/manifest file — Index rebuilds the catalog from disk
             every call. Do not cache or write any sidecar."

# REFERENCE — the repo's test convention (white-box, same-package, shared helpers)
- file: internal/discover/discover_test.go
  why: "Defines writeSkill(t, content) string — the fixture helper (isolated single
        file). index_test.go is package discover, so it shares scope; it REUSES
        writeSkill and strEq (from skill_test.go) and must NOT redefine either
        (redeclaration is a build error). Convention: t.TempDir()+os.WriteFile,
        plain t.Errorf/t.Fatalf, NO testify, NO t.Parallel(). READ-ONLY."
  pattern: "White-box test file; reuse cross-file helpers in the same package. For a
            MULTI-skill tree (which writeSkill can't express), define a new helper
            makeTree (see Blueprint)."

# URLS — the stdlib functions this subtask is built from
- url: https://pkg.go.dev/path/filepath#WalkDir
  why: "filepath.WalkDir(root, fn) — the recursive directory walk. Key behaviors
        verified: it does NOT follow symlinked dirs; on a missing root it calls fn
        once with (root,nil,lstatErr) (why the Stat guard is needed); it visits
        entries in lexical order (final sort still applied for safety)."
- url: https://pkg.go.dev/path/filepath#Rel
  why: "filepath.Rel(base, target) — the skill dir path relative to the skills dir.
        Rel(root,root) == '.' (the root-level edge case)."
- url: https://pkg.go.dev/path/filepath#ToSlash
  why: "filepath.ToSlash — normalizes OS separators to '/' so RelTag is
        'writing/reddit' on Windows too (no-op on Linux/macOS)."
- url: https://pkg.go.dev/sort#Slice
  why: "sort.Slice(skills, less) — deterministic ordering by RelTag."
- url: https://pkg.go.dev/testing#T.Chdir
  why: "t.Chdir (Go 1.24+) — safely scopes a cwd change to one test, used by the
        relative-input test. go.mod is 'go 1.25', so it is available."
```

### Current Codebase tree (S1 + S2 landed; before this subtask)

```bash
$ cd /home/dustin/projects/skpp && find . -name '*.go' -not -path './.pi-subagents/*'
internal/discover/discover.go        # S1: Frontmatter(8 fields) + ParseFrontmatter + utf8BOM
internal/discover/discover_test.go   # S1: 12 white-box tests + writeSkill helper (package discover)
internal/discover/skill.go           # S2: Skill(9 fields) + BuildSkill + toStringSlice
internal/discover/skill_test.go      # S2: toStringSlice + BuildSkill tests + strEq helper
internal/skillsdir/skillsdir.go      # M1.T2: Find() (+ findEnv/findSibling/findWalkUp) -> ABSOLUTE dir
internal/skillsdir/skillsdir_test.go # M1.T2 tests (white-box, package skillsdir)
main.go / main_test.go               # M1.T3: arg-parse + --path + --version (landed, green)

$ ls -A internal/discover/
discover.go  discover_test.go  skill.go  skill_test.go
# go.mod: module github.com/dabstractor/skpp, go 1.25, require gopkg.in/yaml.v3 v3.0.1
#         (yaml.v3 is DIRECT; T5 adds NO module dep -> go.mod/go.sum UNCHANGED)
# baseline: `go build ./...` OK ; `go test ./...` OK (skillsdir + discover + main all green)
# NO internal/discover/index.go yet (THIS subtask). NO resolve/, ui/. NO skills/ (P1.M6.T12).
```

### Desired Codebase tree with files to be added

```bash
skpp/
├── ... (go.mod, go.sum, .gitignore, LICENSE, PRD.md, internal/skillsdir/*, main.go,
│        main_test.go — ALL UNCHANGED; internal/discover/discover.go,
│        discover_test.go, skill.go, skill_test.go — UNCHANGED [S1/S2-owned, Level 4 gate])
└── internal/
    └── discover/
        ├── discover.go       # UNCHANGED (S1)
        ├── discover_test.go  # UNCHANGED (S1; provides writeSkill)
        ├── skill.go          # UNCHANGED (S2)
        ├── skill_test.go     # UNCHANGED (S2; provides strEq)
        ├── index.go          # CREATE — Index(skillsDir) ([]Skill, error): WalkDir walk
        └── index_test.go     # CREATE — white-box tests (12) + makeTree helper
```

| File (created) | Responsibility | Imports |
|---|---|---|
| `internal/discover/index.go` | Walk the skills dir; identify dirs containing `SKILL.md`; compute `RelTag` (`ToSlash(Rel)`); parse frontmatter (lenient); build `Skill` via `BuildSkill`; sort by `RelTag`; return `[]Skill` | `errors`, `io/fs`, `os`, `path/filepath`, `sort` |
| `internal/discover/index_test.go` | White-box tests: single/nested/sort/no-frontmatter/malformed-YAML/non-SKILL.md-filter/empty/missing-root/root-is-file/nested-both/root-level/relative-input | `os`, `path/filepath`, `strings`, `testing` |

**Two new files in the existing `internal/discover/`. Zero changes to `go.mod`,
`go.sum`, `discover.go`, `discover_test.go`, `skill.go`, `skill_test.go`, `main.go`,
`main_test.go`, or any other file. No `resolve`/`ui`/`skills/`.**

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 — A Stat-guard BEFORE WalkDir is MANDATORY (the missing-root bug).
// filepath.WalkDir on a NON-EXISTENT root calls the callback ONCE with
// (root, nil, lstatErr). The per-entry `if err != nil { return nil }` line —
// correct for unreadable SUBTREES — ALSO swallows that root error, so WalkDir
// returns nil and Index returns (nil, nil): a missing skills dir silently looks
// like "no skills", breaking the PRD §6.4/§8.4 "skills dir unresolvable -> exit 1"
// contract. VERIFIED (research Run 1 = the bug, Run 2 = the fix).
//   RIGHT: info, err := os.Stat(root); if err != nil { return nil, err }
//          if !info.IsDir() { return nil, errors.New(root + ": not a directory") }
//          filepath.WalkDir(root, func(...))   // guard FIRST, walk SECOND
//   WRONG: filepath.WalkDir(root, func(path,d,err) error {
//              if err != nil { return nil }   // <-- swallows the ROOT error
//              ...
//          })                                  // missing root -> (nil,nil)

// GOTCHA #2 — filepath.Abs the root FIRST so Skill.Dir is ALWAYS absolute.
// PRD §6.1 says skpp prints an ABSOLUTE path; §13 gates it
// (`case "$(./skpp example)" in /*)`). BuildSkill copies its `dir` arg into
// Skill.Dir verbatim. If a caller passes a relative skillsDir, WalkDir yields
// relative paths and Skill.Dir would be relative -> the absolute contract breaks.
// Abs-ing the root once up front (a no-op Clean on the canonical absolute input
// from skillsdir.Find) makes Dir absolute regardless of caller. VERIFIED.
//   RIGHT: root, err := filepath.Abs(skillsDir)
//   WRONG: using skillsDir as-is (relative input -> relative Dir -> pi gets a rel path)

// GOTCHA #3 — Swallow ParseFrontmatter's malformed-YAML error; do NOT propagate it.
// ParseFrontmatter returns (Frontmatter{}, body, err) for broken YAML. If Index
// returned that err, ONE bad SKILL.md would abort the whole catalog (--list would
// error instead of showing the good skills). The PRD §7.1 leniency is "a skill
// with no frontmatter block resolves by directory". So: ignore the parse err, call
// BuildSkill with the returned Frontmatter{} -> a HasFM=false Skill that is STILL
// resolvable by directory/basename. check (M4/T10) can re-run
// ParseFrontmatter(s.SourceFile) to distinguish "malformed YAML" from "no block".
// VERIFIED (research Run 1/3, TestIndexMalformedYAMLNotAborted).
//   RIGHT: fm, _, _ := ParseFrontmatter(path); skills = append(skills, BuildSkill(skillDir, relTag, fm))
//   WRONG: fm, _, err := ParseFrontmatter(path); if err != nil { return skills, err }

// GOTCHA #4 — relTag MUST be ToSlash'd. filepath.Rel returns OS-native separators
// (already '/' on Linux/macOS, but '\\' on Windows). Canonical tags are compared
// on the '/' form (PRD §7.2; go_architecture.md "relTag normalization"). A
// backslash in RelTag would silently break --list/resolve on Windows. VERIFIED on
// Linux (no-op) + asserted in TestIndexNestedRelTag.
//   RIGHT: relTag := filepath.ToSlash(filepath.Rel(root, skillDir))
//   WRONG: relTag := filepath.Rel(root, skillDir)   // '\\' on Windows

// GOTCHA #5 — WalkDir does NOT follow symlinked directories (stdlib default).
// A symlink that POINTS AT a skill dir is NOT descended into (research Run 2 found
// only 'real', not the 'linked' symlink). PRD §7.1 does not require following
// symlinks, and the default avoids cycles. No action needed — documented behavior.
// (If symlink-following ever becomes a requirement, it needs an explicit
// os.ReadDir + EvalSymlinks recursion, NOT a WalkDir flag — there is none.)

// GOTCHA #6 — Root-level SKILL.md yields relTag == ".". filepath.Rel(root, root)
// returns ".". It IS a skill (the dir contains SKILL.md), so it is INCLUDED for
// spec-compliance; the "." tag is unusual but harmless. Canonical layout nests
// under <tag>/, so this is an edge case. VERIFIED (TestIndexRootLevelSkillMD).
// Do NOT special-case-skip it (no PRD basis to); just document it.

// GOTCHA #7 — index.go is a NEW FILE, NOT appended to discover.go.
// S1's discover.go package doc explicitly says: "The Index() walk (T5)... are
// LATER subtasks — do not add them here." S2's anti-pattern list reserved Index()
// for T5. Putting Index in discover.go would (a) bloat S1's file, (b) risk
// conflating the parser and the walker, (c) drift from the architecture's
// one-concept-per-file layout. Create internal/discover/index.go. VERIFIED (§14).
//   RIGHT: new file internal/discover/index.go  (package discover)
//   WRONG: appending Index() into discover.go

// GOTCHA #8 — Reuse writeSkill (discover_test.go) and strEq (skill_test.go); do
// NOT redefine them. index_test.go is package discover (white-box), so it shares
// scope with both sibling test files. writeSkill is defined in discover_test.go;
// strEq is defined in skill_test.go. Redefining either is a COMPILE ERROR
// ("redeclared in this block"). index_test.go defines ONE new helper, makeTree
// (a multi-skill tree, which writeSkill's isolated-single-file shape can't express).
// VERIFIED (§15).
//   RIGHT: func makeTree(...) {...}   // the only new helper
//   WRONG: func writeSkill(...) {...} // redeclared -> build fails

// GOTCHA #9 — index.go imports ONLY stdlib; NO fmt, NO yaml.v3, NO strings.
// Index needs errors (errors.New), io/fs (fs.DirEntry), os (os.Stat),
// path/filepath (Abs/Stat/WalkDir/Dir/Rel/ToSlash/Join via BuildSkill — but Join
// is in skill.go, not here), sort (sort.Slice). It references the in-package
// Frontmatter/Skill/BuildSkill/ParseFrontmatter but imports neither yaml.v3 (S1
// already does, in discover.go) nor fmt (a dead import fails vet/build). So NO
// module dep is added -> go.mod/go.sum UNCHANGED. VERIFIED (§13).
//   RIGHT: import ("errors"; "io/fs"; "os"; "path/filepath"; "sort")
//   WRONG: import "fmt"   // unused -> `go vet`/`go build` fails

// GOTCHA #10 — An empty skills dir returns (nil, nil), NOT an error.
// "No skills found" is a normal outcome (--list exits 1, --all prints nothing);
// it is distinct from "skills dir missing/unreadable" (which IS an error). Index
// returns a nil slice + nil err for an empty-but-existing tree; callers test with
// len(). VERIFIED (TestIndexEmptyDir). The skillsdir walkup rule already ensures
// Find() only returns a dir with >=1 SKILL.md, but the env/sibling rules do not,
// so Index must handle empty gracefully.

// GOTCHA #11 — Sort happens AFTER the walk, by RelTag (string `<`).
// WalkDir visits entries lexically, but that is per-directory and does NOT give a
// globally-sorted []Skill (a top-level 'apple' vs a nested 'mango/beta' interleave
// by visit order, not by final RelTag). PRD §6.1 --all is "sorted by tag", so sort
// the collected slice by RelTag with sort.Slice. Use byte-wise string comparison
// (fine for ASCII tags). VERIFIED (TestIndexSortedByRelTag).
```

---

## Implementation Blueprint

### Data model — no new types

T5 adds **no** new data model. It consumes S2's `Skill` (value struct, 9 fields, no
yaml tags) and S1's `Frontmatter` (the unmarshal target), both already landed. The
only new symbol is the function `Index`. `[]Skill` is the value slice
`resolve.Resolve(tag, skills []discover.Skill)` (T7) will consume.

### File 1 — `internal/discover/index.go` (CREATE)

Create the file with EXACTLY this content (gofmt-clean; the algorithm was compiled
and all 12 tests pass against a verbatim copy of the real `discover.go` + `skill.go`
in `/tmp/skpp_index_verify` during research):

```go
// index.go implements discover.Index — the on-disk skills/ walk that ties S1's
// ParseFrontmatter (discover.go) and S2's BuildSkill (skill.go) into the []Skill
// the rest of skpp consumes (PRD §7.1). This is the P1.M2.T5.S1 deliverable.
// discover.go (S1) owns the frontmatter model/parser; skill.go (S2) owns the Skill
// type + metadata extraction; index.go (T5) owns the WalkDir scan, the relTag
// normalization, the sort, and the parse-error policy. It is the data source for
// T6 (--list), T7 (resolve), T9 (--search), and T10 (check).
package discover

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Index walks the skills directory at skillsDir and returns every skill it
// contains, as a []Skill sorted by canonical tag (RelTag) for deterministic
// output. It implements PRD §7.1 discovery (manifest-free: the catalog is rebuilt
// from disk on every call — there is no index file).
//
// A "skill" is any directory that directly contains a SKILL.md file; nested skills
// count (skills/writing/reddit/SKILL.md is a skill whose RelTag is
// "writing/reddit"). relTag is the skill dir path relative to skillsDir, with OS
// separators normalized to '/' via filepath.ToSlash — so tags are "writing/reddit"
// on every platform (PRD §7.2 step 1; go_architecture.md "relTag normalization").
//
// skillsDir is made absolute first (filepath.Abs), so every Skill.Dir is an
// absolute path — the contract behind PRD §6.1 ("absolute path") and the §13
// acceptance gate (`case "$(./skpp example)" in /*)`). On the canonical absolute
// input (from skillsdir.Find) Abs is a no-op Clean.
//
// Error policy (the decision S2's PRP assigned to T5; see research/verified_facts.md §8):
//   - skillsDir missing, unreadable, or not a directory -> returned as the error.
//     (The caller, main, prints it to stderr and exits 1, PRD §6.4/§8.4.)
//   - A per-entry error (an unreadable subtree) is SKIPPED; the walk continues.
//   - Malformed YAML inside a SKILL.md does NOT abort the walk and is NOT
//     propagated: ParseFrontmatter returns (Frontmatter{}, body, err); Index
//     ignores err and builds a HasFM=false Skill via BuildSkill so the skill is
//     still resolvable by directory/basename (PRD §7.1). check (M4/T10) can
//     re-run ParseFrontmatter(s.SourceFile) to distinguish "malformed YAML" from
//     "no frontmatter block" (idempotent; no rework).
//
// filepath.WalkDir does NOT follow symlinked directories (stdlib default); a
// symlink to a skill dir is therefore not discovered. PRD §7.1 does not require
// following symlinks, and the default avoids cycles.
//
// An empty skills dir (no SKILL.md anywhere) yields a nil slice and a nil error;
// callers test with len() (e.g. --list exits 1 "if no skills found").
func Index(skillsDir string) ([]Skill, error) {
	root, err := filepath.Abs(skillsDir)
	if err != nil {
		return nil, err
	}
	// Stat-guard BEFORE WalkDir: a missing root is otherwise SWALLOWED. WalkDir
	// feeds the root's lstat error to the callback, and the per-entry
	// `if err != nil { return nil }` below would hide it -> (nil, nil). See
	// research/verified_facts.md Run 1 (the bug) vs Run 2 (the fix).
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New(root + ": not a directory")
	}

	var skills []Skill
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // per-entry unreadable (e.g. a chmod-000 subdir) -> skip, keep walking
		}
		// A skill is identified by the FILE entry "SKILL.md"; directories and any
		// other filename are walked past. d.IsDir() guards against a directory
		// literally named "SKILL.md".
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		skillDir := filepath.Dir(path)
		rel, rerr := filepath.Rel(root, skillDir)
		if rerr != nil {
			return nil // skillDir is always under root (found by walking it), so this is unreachable
		}
		relTag := filepath.ToSlash(rel)
		// Lenient parse: a malformed or frontmatter-less SKILL.md still yields a
		// resolvable Skill (HasFM=false). err is intentionally ignored here; see
		// the doc comment above for the policy + the M4 re-parse note.
		fm, _, _ := ParseFrontmatter(path)
		skills = append(skills, BuildSkill(skillDir, relTag, fm))
		return nil
	})
	if walkErr != nil {
		return skills, walkErr
	}
	// Deterministic output: sort by canonical tag (PRD §6.1 --all "sorted by tag").
	sort.Slice(skills, func(i, j int) bool { return skills[i].RelTag < skills[j].RelTag })
	return skills, nil
}
```

### File 2 — `internal/discover/index_test.go` (CREATE, `package discover` white-box)

Create the file with EXACTLY this content. It mirrors the repo convention (white-box
same-package, `t.TempDir()`/`os.WriteFile`, plain `t.Errorf`/`t.Fatalf`, no testify,
no `t.Parallel()`). It REUSES `writeSkill` (`discover_test.go`) and `strEq`
(`skill_test.go`) — same package — and defines ONE new helper, `makeTree` (a
multi-skill tree). It was compiled and all 12 tests pass against the real package
in `/tmp/skpp_index_verify`.

```go
package discover

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NOTE: this file is white-box `package discover`, so it shares scope with
// discover_test.go (writeSkill) and skill_test.go (strEq). Reuse them; do NOT
// redefine either (redeclaration is a build error). Index() is EXPORTED, so a
// black-box `package discover_test` would also work — we stay white-box to match
// discover_test.go / skill_test.go.

// makeTree builds a temp skills/ tree from a map[relTag]SKILL.md-content and
// returns its root. relTag uses '/' separators (cross-platform via FromSlash).
// A "" key writes SKILL.md directly in the root (the relTag="." edge case).
func makeTree(t *testing.T, layout map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for relTag, content := range layout {
		dir := filepath.Join(root, filepath.FromSlash(relTag))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", dir, err)
		}
	}
	return root
}

// Single top-level skill: full field round-trip + absolute Dir + existing SourceFile.
func TestIndexSingle(t *testing.T) {
	root := makeTree(t, map[string]string{
		"example": "---\nname: example\ndescription: demo\nmetadata:\n  keywords: [a, b]\n  category: meta\n  aliases: [ex]\n---\n# body\n",
	})
	got, err := Index(root)
	if err != nil {
		t.Fatalf("err=%v; want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d; want 1 (%v)", len(got), got)
	}
	s := got[0]
	if s.RelTag != "example" {
		t.Errorf("RelTag=%q; want example", s.RelTag)
	}
	if s.Name != "example" {
		t.Errorf("Name=%q; want example", s.Name)
	}
	if !filepath.IsAbs(s.Dir) {
		t.Errorf("Dir=%q is not absolute (PRD §6.1/§13 absolute contract)", s.Dir)
	}
	if s.SourceFile != filepath.Join(s.Dir, "SKILL.md") {
		t.Errorf("SourceFile=%q; want %q", s.SourceFile, filepath.Join(s.Dir, "SKILL.md"))
	}
	if _, err := os.Stat(s.SourceFile); err != nil {
		t.Errorf("SourceFile does not exist: %v", err)
	}
	if !strEq(s.Keywords, []string{"a", "b"}) {
		t.Errorf("Keywords=%v; want [a b] (end-to-end []any -> []string)", s.Keywords)
	}
	if s.Category != "meta" {
		t.Errorf("Category=%q; want meta", s.Category)
	}
	if !strEq(s.Aliases, []string{"ex"}) {
		t.Errorf("Aliases=%v; want [ex]", s.Aliases)
	}
}

// Nested skill: relTag uses '/' separators (filepath.ToSlash), no backslash.
func TestIndexNestedRelTag(t *testing.T) {
	root := makeTree(t, map[string]string{
		"writing/reddit": "---\nname: reddit\ndescription: d\n---\nx\n",
	})
	got, _ := Index(root)
	if len(got) != 1 || got[0].RelTag != "writing/reddit" {
		t.Fatalf("got=%v; want one skill RelTag=writing/reddit (separator normalized to /)", got)
	}
	if strings.Contains(got[0].RelTag, "\\") {
		t.Errorf("RelTag contains a backslash: %q (must be /-normalized)", got[0].RelTag)
	}
}

// Returned slice is sorted by RelTag (lexicographic), not by walk visit order.
func TestIndexSortedByRelTag(t *testing.T) {
	root := makeTree(t, map[string]string{
		"zebra":      "---\nname: z\ndescription: d\n---\n",
		"apple":      "---\nname: a\ndescription: d\n---\n",
		"mango/fig":  "---\nname: f\ndescription: d\n---\n",
		"mango/beta": "---\nname: b\ndescription: d\n---\n",
	})
	got, _ := Index(root)
	var tags []string
	for _, s := range got {
		tags = append(tags, s.RelTag)
	}
	// Lexicographic by canonical tag; "mango/beta" < "mango/fig" < "zebra".
	want := []string{"apple", "mango/beta", "mango/fig", "zebra"}
	if !strEq(tags, want) {
		t.Errorf("order=%v; want %v (lexicographic by RelTag)", tags, want)
	}
}

// No-frontmatter SKILL.md is still resolved by directory (PRD §7.1): HasFM=false.
func TestIndexNoFrontmatterIncluded(t *testing.T) {
	root := makeTree(t, map[string]string{
		"plain": "# just markdown, no --- block\n",
	})
	got, _ := Index(root)
	if len(got) != 1 {
		t.Fatalf("len=%d; want 1 (no-frontmatter skill still resolved by dir, PRD §7.1)", len(got))
	}
	s := got[0]
	if s.HasFM {
		t.Error("HasFM=true; want false (no --- block)")
	}
	if s.Name != "" || s.Description != "" {
		t.Errorf("Name=%q Description=%q; want empty", s.Name, s.Description)
	}
	if s.RelTag != "plain" {
		t.Errorf("RelTag=%q; want plain", s.RelTag)
	}
}

// Malformed YAML does NOT abort the walk and is NOT propagated: the bad skill is
// included (HasFM=false) and the good sibling is kept. (verified_facts §8.)
func TestIndexMalformedYAMLNotAborted(t *testing.T) {
	root := makeTree(t, map[string]string{
		"good": "---\nname: good\ndescription: d\n---\n",
		"bad":  "---\nname: bad\nmetadata: [unbalanced\n---\nbody\n",
	})
	got, err := Index(root)
	if err != nil {
		t.Fatalf("err=%v; want nil (malformed YAML must NOT abort the walk)", err)
	}
	if len(got) != 2 {
		t.Fatalf("len=%d; want 2 (malformed skill still included, lenient)", len(got))
	}
	if got[0].RelTag != "bad" || got[1].RelTag != "good" {
		t.Errorf("order=%v; want [bad good]", got)
	}
	if got[0].HasFM {
		t.Error("bad: HasFM=true; want false (malformed YAML -> Frontmatter{} -> HasFM=false)")
	}
	if got[1].HasFM != true || got[1].Name != "good" {
		t.Errorf("good: HasFM=%v Name=%q; want true/good", got[1].HasFM, got[1].Name)
	}
}

// Stray files (README.md, *.bak) and subdirs without a SKILL.md are ignored.
func TestIndexIgnoresNonSkillMD(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "real"), 0o755)
	os.WriteFile(filepath.Join(root, "real/SKILL.md"), []byte("---\nname: real\ndescription: d\n---\n"), 0o644)
	// Distractions that must NOT be treated as skills.
	os.WriteFile(filepath.Join(root, "README.md"), []byte("# hi"), 0o644)
	os.MkdirAll(filepath.Join(root, "notes"), 0o755)
	os.WriteFile(filepath.Join(root, "notes/draft.txt"), []byte("draft"), 0o644)
	os.WriteFile(filepath.Join(root, "SKILL.md.bak"), []byte("bak"), 0o644)

	got, _ := Index(root)
	if len(got) != 1 || got[0].RelTag != "real" {
		t.Fatalf("got=%v; want exactly one skill 'real' (stray files/subdirs ignored)", got)
	}
}

// Empty skills dir (exists, no SKILL.md) -> nil/empty slice, nil error.
func TestIndexEmptyDir(t *testing.T) {
	root := t.TempDir() // exists, empty
	got, err := Index(root)
	if err != nil {
		t.Fatalf("err=%v; want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("len=%d; want 0 (empty tree -> no skills)", len(got))
	}
}

// Missing root -> error (the Stat guard; without it this returns (nil,nil)).
func TestIndexMissingRoot(t *testing.T) {
	_, err := Index(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("err=nil; want an error (missing root must propagate after the Stat guard)")
	}
}

// Root that is a regular file -> error ("not a directory").
func TestIndexRootIsFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "notadir")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := Index(f.Name()); err == nil {
		t.Fatal("err=nil; want an error (root must be a directory)")
	}
}

// Nested skills at multiple levels: writing AND writing/reddit are BOTH indexed.
func TestIndexNestedBothLevels(t *testing.T) {
	root := makeTree(t, map[string]string{
		"writing":        "---\nname: writing\ndescription: d\n---\n",
		"writing/reddit": "---\nname: reddit\ndescription: d\n---\n",
	})
	got, _ := Index(root)
	if len(got) != 2 {
		t.Fatalf("len=%d; want 2 (parent is a skill AND has a nested subskill)", len(got))
	}
	if got[0].RelTag != "writing" || got[1].RelTag != "writing/reddit" {
		t.Errorf("got=%v; want [writing writing/reddit]", got)
	}
}

// Edge case: a SKILL.md directly in the skills root -> relTag == "."
// (filepath.Rel(root, root)). Included for spec-compliance; unusual in practice.
func TestIndexRootLevelSkillMD(t *testing.T) {
	root := makeTree(t, map[string]string{"": "---\nname: root\ndescription: d\n---\n"})
	got, err := Index(root)
	if err != nil {
		t.Fatalf("err=%v; want nil", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d; want 1 (root-level SKILL.md is still a skill)", len(got))
	}
	if got[0].RelTag != "." {
		t.Errorf("RelTag=%q; want '.' (filepath.Rel(root,root) edge case)", got[0].RelTag)
	}
}

// Defensive: a RELATIVE skillsDir still yields ABSOLUTE Skill.Dir (filepath.Abs).
// Protects the PRD §6.1/§13 absolute-output contract. t.Chdir (Go 1.24+) scopes cwd.
func TestIndexRelativeInputDirStillAbsolute(t *testing.T) {
	absRoot := makeTree(t, map[string]string{
		"example": "---\nname: example\ndescription: d\n---\n",
	})
	parent := filepath.Dir(absRoot)
	t.Chdir(parent)
	rel, err := filepath.Rel(parent, absRoot)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Index(rel)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d; want 1", len(got))
	}
	if !filepath.IsAbs(got[0].Dir) {
		t.Errorf("Dir=%q is RELATIVE; want absolute (relative input must still abs-ify)", got[0].Dir)
	}
	if got[0].RelTag != "example" {
		t.Errorf("RelTag=%q; want example", got[0].RelTag)
	}
}
```

> **Copy-paste correctness:** both blueprint files are gofmt-clean and compile
> against the real `discover.go` (S1) + `skill.go` (S2). `index.go` imports exactly
> `errors`/`io/fs`/`os`/`path/filepath`/`sort` (all used). `index_test.go` imports
> `os`/`path/filepath`/`strings`/`testing` and reuses `writeSkill`/`strEq` from the
> sibling test files (same package — must NOT be redefined). The algorithm + tests
> were compiled and run in `/tmp/skpp_index_verify` (a verbatim copy of the real
> discover package with the package line flipped to `main`); every asserted value
> traces to recorded output in `research/verified_facts.md`.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: CREATE internal/discover/index.go
  - WRITE: the exact content from the Blueprint (File 1).
  - CHECK: `package discover`; imports ONLY errors/io/fs/os/path/filepath/sort;
           `func Index(skillsDir string) ([]Skill, error)`; the Stat-guard BEFORE
           WalkDir; filepath.Abs first; relTag via ToSlash(Rel); lenient
           ParseFrontmatter (err ignored); sort.Slice by RelTag.
  - GOTCHA: the Stat-guard is MANDATORY (Gotcha #1 — missing root else swallowed).
            Do NOT propagate the parse err (Gotcha #3). Do NOT use fmt (Gotcha #9).
            Do NOT put Index in discover.go (Gotcha #7).

Task 2: CREATE internal/discover/index_test.go
  - WRITE: the exact content from the Blueprint (File 2).
  - CHECK: `package discover` (white-box); imports os/path/filepath/strings/testing.
           REUSES writeSkill (discover_test.go) + strEq (skill_test.go); defines ONE
           new helper makeTree. 12 tests (see success criteria).
  - GOTCHA: do NOT redefine writeSkill/strEq (Gotcha #8 — redeclaration build error).
            NO testify; NO t.Parallel(). Compare slices with strEq (len+elements),
            NOT reflect.DeepEqual. t.Chdir (Go 1.24+) is fine (go.mod is go 1.25).

Task 3: FORMAT + VET + TIDY + BUILD + TEST (validation gates — run in order)
  - COMMAND: gofmt -w internal/discover/index.go internal/discover/index_test.go
  - COMMAND: gofmt -l internal/discover/*.go   # MUST print nothing
  - COMMAND: go vet ./internal/discover/       # MUST be clean
  - COMMAND: go mod tidy   # EXPECTED: a NO-OP (T5 adds no module dep; all imports stdlib)
  - COMMAND: go build ./...                    # exit 0 (whole module compiles)
  - COMMAND: go test ./internal/discover/ -v   # T5's 12 NEW tests + S1's 12 + S2's, ALL PASS
  - COMMAND: go test ./...                     # whole module green (skillsdir + discover + main)
  - EXPECT: zero errors, zero vet findings, gofmt silent, all tests pass, go.mod/go.sum unchanged.

Task 4: EXPORTED-API SMOKE TEST (Level 3 in Validation Loop)
  - COMMAND: go doc ./internal/discover Index
  - EXPECT: prints Index's godoc (signature + the error-policy / relTag / sort notes).

Task 5: SCOPE BOUNDARY CHECK — Level 4 in Validation Loop
  - COMMAND: the Level 4 block below.
  - EXPECT: index.go has Index only; imports exactly the 5 stdlib packages;
            discover.go/discover_test.go/skill.go/skill_test.go UNCHANGED;
            go.mod/go.sum UNCHANGED; no main.go/skillsdir/resolve/ui touch; no skills/.
```

### Implementation Patterns & Key Details

```go
// PATTERN: Stat-guard the walk root, THEN WalkDir; distinguish root vs per-entry errors.
//   info, err := os.Stat(root)
//   if err != nil { return nil, err }              // root missing/unreadable -> fatal
//   if !info.IsDir() { return nil, errors.New(...) } // root must be a dir
//   filepath.WalkDir(root, func(path, d, err) error {
//       if err != nil { return nil }               // per-entry -> skip, keep walking
//       ...
//   })
// WHY: WalkDir feeds the ROOT's lstat error to the callback, so the per-entry
//      `if err != nil { return nil }` would swallow a missing root (-> (nil,nil)).
//      Stat-ing the root first makes root-level failures fatal while per-entry
//      failures stay non-fatal. os.Stat (not Lstat) follows a symlinked root,
//      matching skillsdir.findEnv. VERIFIED (research Run 1 vs Run 2).
//   WRONG: WalkDir with no pre-check (missing root silently returns nil).

// PATTERN: make the root absolute ONCE, up front (filepath.Abs).
//   root, err := filepath.Abs(skillsDir)
// WHY: BuildSkill copies its `dir` arg verbatim into Skill.Dir; WalkDir yields
//      paths joined to the root AS GIVEN, so a relative root -> relative Dir ->
//      breaks the PRD §6.1/§13 absolute contract. Abs-ing once makes Dir absolute
//      regardless of caller, and is a no-op Clean on the canonical absolute input.
//      VERIFIED (TestIndexRelativeInputDirStillAbsolute).

// PATTERN: identify a skill by the SKILL.md FILE entry; act in the file branch.
//   if d.IsDir() || d.Name() != "SKILL.md" { return nil }
//   skillDir := filepath.Dir(path)
// WHY: "a skill = any directory that directly contains a SKILL.md" (PRD §7.1).
//      Walking files (not dirs) and taking the parent dir is the natural read.
//      d.IsDir() guards against a directory literally named "SKILL.md".
//      Stray README.md / *.bak / files-in-subdirs-without-SKILL.md are ignored.

// PATTERN: delegate ALL field extraction to BuildSkill; never read frontmatter directly.
//   fm, _, _ := ParseFrontmatter(path)
//   skills = append(skills, BuildSkill(skillDir, relTag, fm))
// WHY: S2 owns the []any->[]string normalization and the nil-metadata safety.
//      Index computes ONLY (skillDir, relTag) from the walk and hands them to
//      BuildSkill. This is the seam S2's PRP pre-specified; touching frontmatter
//      fields here would duplicate S2 and risk the []any panic.

// PATTERN: lenient parse — swallow the malformed-YAML error, build anyway.
//   fm, _, _ := ParseFrontmatter(path)   // err intentionally discarded
// WHY: ParseFrontmatter returns (Frontmatter{}, body, err) on broken YAML.
//      Propagating err would let ONE bad file abort the whole catalog. BuildSkill
//      on Frontmatter{} yields a HasFM=false Skill (resolvable by dir, PRD §7.1).
//      check (M4) can re-parse s.SourceFile if it needs the exact YAML error.
//      VERIFIED (TestIndexMalformedYAMLNotAborted).

// PATTERN: sort the collected slice by RelTag AFTER the walk.
//   sort.Slice(skills, func(i,j int) bool { return skills[i].RelTag < skills[j].RelTag })
// WHY: WalkDir's per-directory lexical order is not a global sort. PRD §6.1 --all
//      is "sorted by tag"; sorting the slice gives deterministic output for
//      --all/--list. Byte-wise string `<` is correct for ASCII tags. VERIFIED.

// PATTERN: white-box test file reuses cross-file helpers, adds one tree builder.
//   // index_test.go (package discover) reuses writeSkill + strEq; adds makeTree
//   func makeTree(t *testing.T, layout map[string]string) string { ... }
// WHY: Go compiles all *_test.go in a package together, so helpers are shared.
//      writeSkill (discover_test.go) builds an isolated single file; a multi-skill
//      tree needs a new helper (makeTree). Redefining writeSkill/strEq is a build
//      error. VERIFIED (research §15).
```

### Integration Points

```yaml
PACKAGE BOUNDARIES (after this subtask):
  - internal/discover/ contains discover.go (S1, UNCHANGED) + skill.go (S2,
    UNCHANGED) + index.go (T5, NEW) + their test files. All `package discover`
    (internal -> unimportable outside the module; correct for a CLI's innards).
  - index.go imports ONLY stdlib (errors/io/fs/os/path/filepath/sort). It
    references the in-package Frontmatter/Skill/BuildSkill/ParseFrontmatter but
    imports neither yaml.v3 (S1 does, in discover.go) nor fmt.
  - exposes (exported): func Index. (Plus S1's Frontmatter/ParseFrontmatter and
    S2's Skill/BuildSkill.)

go.mod / go.sum (UNCHANGED — verified_facts.md §13):
  - before/after: require gopkg.in/yaml.v3 v3.0.1   (yaml.v3 already DIRECT from S1)
  - go.sum: UNCHANGED. `go mod tidy` is a NO-OP (T5 adds no module; all imports stdlib).

DOWNSTREAM EXTENSION POINTS (what later subtasks plug into):
  - P1.M2.T6.S1 (--list): ui reads the sorted []Skill; shows Description as
    "(missing)" when !HasFM or Description==""; exits 1 when len==0.
  - P1.M3.T7.S1 (resolve): resolve.Resolve(tag, skills []discover.Skill) matches
    §7.2 precedence over Skill.RelTag (exact), the basename of RelTag, Skill.Name,
    and Skill.Aliases. Index's deterministic sort + absolute Dir + /-normalized
    RelTag are the shapes it relies on.
  - P1.M3.T8.S1 (skpp <tag>): prints Skill.Dir (absolute) or, with -f, Skill.SourceFile.
  - P1.M4.T9.S1 (--search): substring over Skill.RelTag/Name/Description AND
    Skill.Keywords (the []string S2 normalized).
  - P1.M4.T10.S1 (check): derives most rules from []Skill (HasFM/Name/Description,
    duplicate Name, length); can re-run ParseFrontmatter(s.SourceFile) to flag
    malformed YAML distinctly from a missing block.
  - main wiring (a later milestone): dir,_,_ := skillsdir.Find(); skills,err :=
    discover.Index(dir); if err != nil { stderr; exit 1 }. (NOT added in this subtask.)

NO CHANGES TO:
  - go.mod / go.sum (T5 adds no module)
  - PRD.md (read-only) / any tasks.json (orchestrator-owned) / prd_snapshot.md
  - internal/discover/discover.go + discover_test.go (S1-owned — Level 4 gate)
  - internal/discover/skill.go + skill_test.go (S2-owned — Level 4 gate)
  - internal/skillsdir/* (M1-owned)
  - main.go / main_test.go (M1.T3-owned; Index wiring is a LATER milestone)
  - any other package or file (resolve/ui are later milestones; skills/ is P1.M6.T12)
```

---

## Validation Loop

### Level 1: Format, vet, tidy, build (immediate, per file)

```bash
cd /home/dustin/projects/skpp

# Format in place, then confirm nothing is left unformatted (silent == pass).
gofmt -w internal/discover/index.go internal/discover/index_test.go
test -z "$(gofmt -l internal/discover/*.go)" \
  || { echo "FAIL: gofmt found unformatted files"; gofmt -d internal/discover/; exit 1; }
echo "gofmt OK"

# Vet the package (index.go + discover.go + skill.go together).
go vet ./internal/discover/ || { echo "FAIL: go vet ./internal/discover/"; exit 1; }
echo "go vet OK"

# Tidy: EXPECTED no-op for T5 (all imports are stdlib; yaml.v3 already direct).
go mod tidy
git diff --quiet go.mod go.sum 2>/dev/null \
  && echo "go.mod/go.sum unchanged OK" \
  || { echo "FAIL: go.mod/go.sum changed (T5 must not touch them)"; git diff go.mod go.sum; exit 1; }

# Build the whole module (compile check across packages).
go build ./... || { echo "FAIL: go build ./..."; exit 1; }
echo "go build OK"
```

### Level 2: Unit tests (component validation)

```bash
cd /home/dustin/projects/skpp

# Run discover tests verbosely — T5's 12 NEW tests + S1's 12 + S2's together.
go test ./internal/discover/ -v || { echo "FAIL: go test ./internal/discover/ -v"; exit 1; }

# Targeted: T5's own tests only.
go test ./internal/discover/ -run TestIndex -v \
  || { echo "FAIL: T5 Index tests"; exit 1; }

# The load-bearing assertions (the Stat guard, malformed-YAML leniency, nested relTag).
go test ./internal/discover/ -run \
  'TestIndexMissingRoot|TestIndexMalformedYAMLNotAborted|TestIndexNestedRelTag|TestIndexSortedByRelTag|TestIndexRelativeInputDirStillAbsolute' -v \
  || { echo "FAIL: load-bearing T5 tests"; exit 1; }

# Whole module still green (skillsdir + discover + main).
go test ./... || { echo "FAIL: go test ./..."; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Exported-API smoke test (library subtask — no main to run)

T5 is a leaf library, so the "acceptance smoke" is confirming the exported API is
present, documented, and behaves via the tests (Level 2). The full CLI end-to-end
(`skpp <tag>`, `skpp --path`) is gated on a LATER milestone that wires `Find() →
Index()`; T5 is not reachable from `main` yet, by design.

```bash
cd /home/dustin/projects/skpp

# Index is exported with its behavior godoc.
go doc ./internal/discover Index | grep -qE 'func Index\(skillsDir string\) \(\[\]Skill, error\)' \
  || { echo "FAIL: Index signature wrong/undocumented"; exit 1; }
go doc ./internal/discover Index | grep -q 'sorted by canonical tag' \
  || { echo "FAIL: Index godoc missing the sort/relTag contract"; exit 1; }

# End-to-end behavior (Index -> Skill fields) is covered by TestIndexSingle (Level 2).
go test ./internal/discover/ -run 'TestIndexSingle|TestIndexNestedRelTag' -v
```

### Level 4: Scope boundary check (do not regress S1/S2 or the module)

```bash
cd /home/dustin/projects/skpp

echo "--- index.go owns ONLY Index (no S1/S2 symbols redefined here) ---"
grep -nE '^func |^type ' internal/discover/index.go
test "$(grep -cE '^func Index\(' internal/discover/index.go)" = "1" \
  || { echo "FAIL: index.go must define exactly one func, Index"; exit 1; }
# index.go must NOT redefine Skill/BuildSkill/ParseFrontmatter/Frontmatter/toStringSlice.
! grep -qE '^type (Skill|Frontmatter) struct' internal/discover/index.go \
  || { echo "FAIL: index.go redefines an S1/S2 type"; exit 1; }
! grep -qE '^func (BuildSkill|ParseFrontmatter|toStringSlice)' internal/discover/index.go \
  || { echo "FAIL: index.go redefines an S1/S2 function"; exit 1; }

echo "--- index.go imports ONLY stdlib (no fmt/yaml/strings) ---"
grep -A8 '^import (' internal/discover/index.go | grep -qE '"errors"|"io/fs"|"os"|"path/filepath"|"sort"'
! grep -A8 '^import (' internal/discover/index.go | grep -qE '"fmt"|"strings"|yaml \
  || { echo "FAIL: index.go must not import fmt/strings/yaml"; exit 1; }

echo "--- S1 & S2 files unchanged (no Index / no edits) ---"
git diff --quiet internal/discover/discover.go internal/discover/discover_test.go \
  internal/discover/skill.go internal/discover/skill_test.go 2>/dev/null \
  && echo "S1/S2 files unchanged OK" \
  || { echo "FAIL: S1/S2 files were modified (T5 must not touch them)"; git diff internal/discover/discover.go internal/discover/skill.go; exit 1; }
! grep -q 'func Index' internal/discover/discover.go internal/discover/skill.go \
  || { echo "FAIL: Index must live ONLY in index.go"; exit 1; }

echo "--- go.mod / go.sum unchanged ---"
git diff --quiet go.mod go.sum 2>/dev/null \
  && echo "go.mod/go.sum unchanged OK" \
  || { echo "FAIL: go.mod/go.sum changed"; exit 1; }

echo "--- no out-of-scope files touched ---"
git status --porcelain | grep -vE 'internal/discover/index\.go|internal/discover/index_test\.go' \
  && { echo "FAIL: unexpected files changed (see above)"; exit 1; } \
  || echo "scope OK (only index.go + index_test.go changed)"

echo "Level 4 PASS"
```

### Level 5: Creative & domain-specific validation (manifest-free invariant)

```bash
cd /home/dustin/projects/skpp

# Manifest-free (PRD §2.1): there must be NO index/manifest file anywhere. Index
# rebuilds the catalog from disk on every call; it writes nothing.
! find . -path ./.git -prune -o -type f \( -name 'skills.json' -o -name 'skills.yaml' \
  -o -name 'skills.idx' -o -name '.skpp-index' \) -print | grep -q . \
  && echo "manifest-free OK" \
  || { echo "FAIL: a manifest/index file exists (forbidden by PRD §2.1)"; exit 1; }

# Behavior repro of research Run 2 (a tiny throwaway check is already covered by
# the unit tests; this just re-states the key invariant for the reviewer):
#   - missing root      -> err  (TestIndexMissingRoot)
#   - malformed YAML    -> included, HasFM=false, walk NOT aborted (TestIndexMalformedYAMLNotAborted)
#   - nested relTag     -> "writing/reddit", '/'-normalized        (TestIndexNestedRelTag)
echo "Level 5 PASS (invariants asserted by Level 2 tests)"
```

## Final Validation Checklist

### Technical Validation

- [ ] All 5 validation levels completed successfully.
- [ ] All tests pass: `go test ./...` (S1's 12 + S2's + T5's 12 new, all green).
- [ ] No vet findings: `go vet ./internal/discover/` clean.
- [ ] No formatting issues: `gofmt -l internal/discover/*.go` silent.
- [ ] `go mod tidy` is a no-op (go.mod/go.sum unchanged — T5 adds no module).

### Feature Validation

- [ ] `Index(root)` returns `[]Skill` sorted by `RelTag` (TestIndexSortedByRelTag).
- [ ] Nested skill `writing/reddit` → `RelTag == "writing/reddit"` (TestIndexNestedRelTag).
- [ ] Every `Skill.Dir` is absolute; `Skill.SourceFile` exists (TestIndexSingle).
- [ ] No-frontmatter and malformed-YAML skills are INCLUDED (`HasFM=false`), and
      malformed YAML does NOT abort the walk (TestIndexNoFrontmatterIncluded,
      TestIndexMalformedYAMLNotAborted).
- [ ] Stray `README.md`/`*.bak`/subdir-without-`SKILL.md` are ignored
      (TestIndexIgnoresNonSkillMD).
- [ ] Missing root and root-is-file both return an error (the Stat guard)
      (TestIndexMissingRoot, TestIndexRootIsFile).
- [ ] Empty dir → `len==0`, nil error (TestIndexEmptyDir).

### Code Quality Validation

- [ ] Follows the repo's one-concept-per-file layout (Index in `index.go`, not
      appended to `discover.go`).
- [ ] Reuses `writeSkill`/`strEq` (no redefine); adds only `makeTree`.
- [ ] No testify, no `t.Parallel()` (repo convention).
- [ ] Imports limited to stdlib (`errors`/`io/fs`/`os`/`path/filepath`/`sort`).
- [ ] Anti-patterns avoided (see below).

### Documentation & Scope

- [ ] `Index` carries a godoc comment covering the walk, relTag, sort, and the
      parse-error policy (Level 3 greps for it).
- [ ] S1/S2 files, `main*`, `skillsdir*`, `go.mod`/`go.sum` all UNCHANGED (Level 4).
- [ ] No manifest/index file written (manifest-free invariant, PRD §2.1) (Level 5).

---

## Anti-Patterns to Avoid

- ❌ **Do NOT skip the `os.Stat` root guard.** Without it a missing skills dir
  returns `(nil, nil)` — the WalkDir callback swallows the root's lstat error.
  (Gotcha #1; research Run 1 = the bug.)
- ❌ **Do NOT propagate `ParseFrontmatter`'s malformed-YAML error.** One bad file
  would abort the whole catalog and break `--list`. Swallow it; build a
  `HasFM=false` Skill; let `check` (M4) re-parse if it needs the exact error.
  (Gotcha #3.)
- ❌ **Do NOT append `Index` to `discover.go`.** S1's package doc reserves
  `discover.go` for the frontmatter model/parser; S2 reserved `Index` for T5 in its
  own file. Use `internal/discover/index.go`. (Gotcha #7.)
- ❌ **Do NOT skip `filepath.ToSlash`** on relTag. On Windows `filepath.Rel`
  returns `\` separators; canonical tags must be `/`-normalized. (Gotcha #4.)
- ❌ **Do NOT use `skillsDir` raw (skip `filepath.Abs`).** A relative input would
  produce relative `Skill.Dir`, breaking the PRD §6.1/§13 absolute-path contract.
  (Gotcha #2.)
- ❌ **Do NOT redefine `writeSkill`/`strEq` in `index_test.go`.** They already live
  in `discover_test.go`/`skill_test.go` (same `package discover`); redefinition is a
  build error. (Gotcha #8.)
- ❌ **Do NOT import `fmt` (or `strings`/`yaml.v3`) in `index.go`.** They are
  unused; a dead import fails `go vet`/`go build`. Use `errors.New` (string concat)
  for the not-a-directory message. (Gotcha #9.)
- ❌ **Do NOT add a manifest/index/cache file.** PRD §2.1 is manifest-free; `Index`
  rebuilds from disk every call and writes nothing. (Level 5.)
- ❌ **Do NOT wire `Index` into `main.go` here.** The `skillsdir.Find() →
  discover.Index(dir)` wiring is a LATER milestone (the `--list`/`<tag>` modes).
  T5 is a library-only deliverable. (Scope; see NOTE in the header.)
- ❌ **Do NOT touch S1's or S2's files.** `discover.go`/`discover_test.go`/
  `skill.go`/`skill_test.go` are landed and green; modifying them risks the S1/S2
  contracts and is forbidden (Level 4 gate asserts unchanged).
