# PRP — P1.M2.T2.S2: Create store (mkdir + seed compiled-in template) and write config.yaml

> **Subtask:** The store-CREATION half of `skilldozer init` (PRD §8.2 steps 2–4). Delivers a testable core `setupStore(store, configPath string) (seeded bool, err error)` that: (a) `os.MkdirAll(store, 0o755)` — create the store if missing (§8.2 step 2); (b) if the store dir is **empty** (`os.ReadDir` returns zero entries), seed `example/SKILL.md` under it from a **compiled-in string constant** `exampleSkillTemplate` (PRD §11 verbatim — `skilldozer`, NOT `skpp); seeded=true; (c) if the store already contains **anything**, adopt it in place and NEVER clobber or delete (§8.2/§17); seeded=false; (d) always write the config via `config.Save(configPath, config.File{Store: store})` (§8.2 step 4) — `store` is already absolute (S1's `resolveStore` absolutized it). Returns `(seeded, nil)` on success or a wrapped `fmt.Errorf` on any fs failure. **It does NOT choose the store** (that is P1.M2.T2.S1's `resolveStore`, already landed) and **it does NOT add `if c.init { … }` to `run()`** (that is P1.M2.T2.S3). After this subtask the create+seed+writeconfig logic exists and is fully unit-tested, but `init` still no-ops to exit 1 until S3 wires `resolveStore`+`setupStore` into `run()`.
>
> **Scope:** Two existing files only — `main.go` (append ONE `const` + ONE `func` after `resolveStore`, the file tail at line 876; **zero new imports** — `os`/`path/filepath`/`fmt`/`config` are all already imported) and `main_test.go` (add 4 `setupStore` unit tests + one `config` import line). No new files. No `internal/*` change. `go.mod`/`go.sum` byte-for-byte unchanged. Mirrors S1's "purely additive at the file tail" discipline.
>
> **STATUS (verified at PRP-write time):** main.go + main_test.go + internal/config/config.go read directly. `config.Save(path, File)` (config.go:69), `config.File{Store string}` (config.go:30), `config.Load(path)` (config.go:52) signatures confirmed. `resolveStore` (main.go:843–876, returns ABSOLUTE store) is present (S1 code already landed). main.go's import block already contains `internal/config` (S1 added it) — `grep -c 'internal/config' main.go` == 1 — so setupStore needs **zero** new main.go imports. main_test.go currently imports stdlib only (bytes/errors/io/os/path/filepath/strings/testing) — needs +1 import (`internal/config`) for the `config.Load` round-trip assertions. `grep` confirms main.go has no current `setupStore`/`exampleSkillTemplate` references — purely additive. The parallel sibling P1.M2.T2.S1 PRP was read as a CONTRACT; its `resolveStore` output is this subtask's `store` input.

---

## Goal

**Feature Goal**: Implement and unit-test the `skilldozer init` store-creation logic in isolation from the terminal and from store-CHOICE, so that PRD §8.2 steps 2–4 hold and are provably correct via 4 unit tests that never touch a real config-file location or a real terminal: (a) an empty store dir is created and seeded with the PRD §11 `example/SKILL.md` from a compiled-in string constant (NOT `go:embed` — PRD §17 forbids it; code_prd_delta.md G11 is this gap); (b) a non-empty store is adopted in place with zero clobber/delete; (c) the config is always written with `store: <absolute>`; (d) the whole thing is idempotent on re-run; (e) any fs failure returns a wrapped error and writes nothing.

**Deliverable**: Additive edits to two existing files:
1. `main.go` — append `const exampleSkillTemplate` (the PRD §11 body as a compiled-in string constant, spliced for backticks — see GOTCHA #1) and `func setupStore(store, configPath string) (seeded bool, err error)` immediately after `resolveStore` (file tail, line 876). **No import changes** (`os`/`path/filepath`/`fmt`/`config` already imported by S1).
2. `main_test.go` — add the `internal/config` import (new second import group) and 4 `TestSetupStore*` unit tests (empty-seeds, non-empty-adopts, idempotent, MkdirAll-failure-error-path).

**Success Definition**: `gofmt -l main.go main_test.go` empty; `go build/vet/test ./...` all pass; `go.mod`/`go.sum` unchanged; `setupStore(emptyTmp, cfg)` ⇒ `(true, nil)` with `emptyTmp/example/SKILL.md` bytes == `exampleSkillTemplate` and `config.Load(cfg).Store == emptyTmp`; `setupStore(dirWithAFile, cfg)` ⇒ `(false, nil)` with the file intact, no `example/` created, config written; `setupStore` run twice ⇒ run1 `(true,nil)`, run2 `(false,nil)`, `example/SKILL.md` byte-identical across runs; `setupStore(pathThatIsAFile, cfg)` ⇒ `(false, err)` and `cfg` not written.

---

## User Persona (if applicable)

**Target User**: A first-run `skilldozer` user whose chosen store dir (from S1's `resolveStore`) may not exist yet, may exist but be empty, or may exist and already contain skills. The create+seed+writeconfig logic serves all three without ever destroying user data.

**Use Case**: User runs `skilldozer init` (or `init --store <dir>`) → skilldozer ensures the store dir exists; if it was just created/empty it drops in a copy-paste `example/SKILL.md` so `--list`/resolution have something to show out of the box; if the dir already has skills it adopts them silently; it then records `store: <abs>` in config.yaml so future runs find the store via the §8.3 config rule.

**User Journey**: (S1 chooses the dir →) `setupStore(absStore, configPath)`: `MkdirAll(absStore)` → `ReadDir(absStore)` empty? → write `absStore/example/SKILL.md` from the compiled-in template → `config.Save(configPath, {Store: absStore})`. (S3 then prints `--path` + `check`.)

**Pain Points Addressed**: a user pointed at a non-existent dir gets it created (no manual `mkdir`); a brand-new user gets a working example skill immediately (the §11 "out of the box" goal); a user who already has a skills repo pointed `--store` at it loses nothing (never clobber/delete — §17); the config is written so the §8.3 config-file rule resolves on the very next invocation.

---

## Why

- **Implements PRD §8.2 steps 2–4** (mkdir → seed-if-empty/adopt-if-not → write config), the create+seed+writeconfig half of `init`. S1 chose the dir; this subtask makes it real on disk.
- **Closes gap G11** (`code_prd_delta.md` §10): "(absent) No compiled-in string-constant example template (PRD forbids `go:embed`)". Delivered as `const exampleSkillTemplate` (a string constant, NOT `go:embed`), honoring PRD §17 "nothing about the user's collection is compiled in."
- **Honors the §17 guardrail** ("never clobber or delete existing files"): the adopt-in-place branch does NOTHING to existing entries; only an empty store is seeded. Locked by the non-empty + idempotent tests.
- **Makes the §8.3 config-file rule resolvable immediately** by writing `store: <abs>` on every successful init (seeded OR adopted).
- **Keeps yaml.v3 the sole non-stdlib dependency** (PRD §4): pure stdlib fs ops + the existing `config` package. No new imports in main.go.
- **Unblocks P1.M2.T2.S3** (run() dispatch), which will call `config.Path()` + `resolveStore(c.initStore)` + `setupStore(absStore, configPath)` and then print `--path` + `check`.

---

## What

### Success Criteria

- [ ] `main.go` appends `const exampleSkillTemplate` whose bytes are EXACTLY the PRD §11 body (frontmatter + `# Example Skill` + the inline `` `skilldozer` `` + the ```bash fenced block), ending with a trailing newline. Implemented as raw-string-segments `+`-concatenated with double-quoted backtick runs (GOTCHA #1 — a single raw literal cannot contain the 8 backticks in §11).
- [ ] `main.go` appends `func setupStore(store, configPath string) (seeded bool, err error)` after `resolveStore` (line 876) that: `os.MkdirAll(store, 0o755)`; `os.ReadDir(store)`; if `len(entries)==0` → `os.MkdirAll(filepath.Join(store,"example"),0o755)` + `os.WriteFile(filepath.Join(store,"example","SKILL.md"), []byte(exampleSkillTemplate), 0o644)` + `seeded=true`; (non-empty → adopt, `seeded` stays false); then `config.Save(configPath, config.File{Store: store})`; return `(seeded, nil)`. Every fs failure returns `(false, fmt.Errorf("skilldozer init: <step>: %w", err))`.
- [ ] `setupStore` does NOT absolutize `store` (S1's `resolveStore` already did) and does NOT obtain `configPath` itself (run()/S3 passes `config.Path()`'s result) — both are injected strings, so the function is directly unit-testable with temp paths (no wrapper layer needed, unlike S1's chooseStore/resolveStore pair).
- [ ] `main_test.go` adds the `internal/config` import (new second group) and 4 `TestSetupStore*` tests: empty-seeds, non-empty-adopts, idempotent, MkdirAll-failure.
- [ ] `go test ./...` green including the 4 new tests; existing tests unaffected (purely additive).
- [ ] `go.mod`/`go.sum` unchanged; no new files; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every call is pinned to a verified, live symbol: `config.Save(path, File)` (config.go:69), `config.File{Store string}` (config.go:30), `config.Load(path)` (config.go:52) — signatures read from source. `resolveStore` (main.go:843–876) is confirmed present (S1 landed) and returns an ABSOLUTE store (it calls `filepath.Abs`), so setupStore's `store` param is already absolute and is written verbatim into the config. main.go's import block already contains `os`, `path/filepath`, `fmt`, and `internal/config` (S1 added config) — `grep -c` confirms — so setupStore needs ZERO new main.go imports. The empty-check semantics (`os.ReadDir` → `len==0`) are traced to the contract test "a dir containing a file leaves it intact (seeded=false)". The backtick-in-raw-string gotcha (§5 of verified_facts) is resolved with an exact spliced `const` whose bytes are locked byte-for-byte by Test #1. The adopt-in-place rule is traced to PRD §17. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (signatures + the backtick gotcha + the test matrix)
- file: plan/002_38acb6d28a6a/P1M2T2S2/research/verified_facts.md
  why: "§1 = the INPUT contract (resolveStore returns ABS store; config.Save/File/Load
        exact signatures). §2 = ZERO new main.go imports (config already imported by S1).
        §3 = main_test.go needs +1 import (config). §4 = append after resolveStore (line 876).
        §5 = THE backtick gotcha + the exact spliced const source. §6 = os.ReadDir empty
        check. §7 = adopt-in-place (never clobber). §8 = error-path (false,err) convention.
        §9 = the 4-test matrix. §10 = the two-copies note. §11 = sibling boundaries."
  critical: "§5 — a single raw string literal CANNOT hold the §11 template (8 backticks);
             the spliced const is mandatory and its bytes are locked by Test #1. §6 — empty
             means ZERO os.ReadDir entries (ANY file ⇒ adopt), NOT 'no SKILL.md'."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as S1/S3 land)
- file: main.go
  why: "THE edit target. Import block @14-27 ALREADY has os/path/filepath/fmt/internal/config
        (S1) — DO NOT add imports. resolveStore is the LAST func @843-876 — APPEND the
        exampleSkillTemplate const + setupStore func immediately after line 876 (file tail).
        const usageText @52 (the existing big-compiled-in-string-const pattern to mirror for
        exampleSkillTemplate — same home style, same const-not-var, same doc-comment care)."
  pattern: "Compiled-in string constant = package-level `const name = `...`` raw literal (see
            usageText @52). I/O logic that takes its targets as injected strings = directly
            unit-testable, no wrapper layer (cf. S1's chooseStore factoring, but simpler here
            because there is no terminal/global to factor out — both deps are strings)."

# MUST READ — the consumed config package (signatures are the contract)
- file: internal/config/config.go
  why: "config.File @30 (`Store string yaml:\"store,omitempty\"`); config.Save @69
        (path FIRST then File; MkdirAll parent 0o755; write 0o644); config.Load @52
        (lenient unknown keys; missing→fs.ErrNotExist verbatim). setupStore calls
        `config.Save(configPath, config.File{Store: store})` EXACTLY. Tests call
        `config.Load(cfg)` to round-trip-assert Store."
  gotcha: "Save's arg order is (path, File) — path FIRST. Do not transpose to (File, path).
           Save already MkdirAll's the config FILE's parent dir, but NOT the store — setupStore
           must MkdirAll the store itself (§8.2 step 2 is setupStore's job, not config's)."

# MUST READ — the test file under edit (mirror these helpers/shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. Imports @3-11 are stdlib-only —
        ADD a second group with internal/config. writeSkillTree @39 + t.TempDir() /
        os.MkdirAll/WriteFile conventions = the house style for fs fixtures. The setupStore
        tests build a temp store via t.TempDir(), pre-seed a file with os.WriteFile for the
        non-empty case, and read config back with config.Load (mirroring config_test.go's
        TestSaveLoadRoundTrip @31 assertion style: got.Store != want)."
  gotcha: "Do NOT write a run(['init']) or resolveStore test — run() dispatch is S3 (today
           exits 1) and resolveStore touches os.Stdin/os.Getwd (S1's integration territory).
           The unit surface is setupStore ONLY (both deps are injectable strings)."

# READ-ONLY — the gap analysis (G11 is THIS subtask)
- file: .pi-subagents/artifacts/outputs/76bb9bcc/plan/002_38acb6d28a6a/architecture/code_prd_delta.md
  why: "G11 (§10 row) = '(absent) No compiled-in string-constant example template (PRD forbids
        go:embed)' — this subtask closes it. §8 confirms go.mod needs no new dep (stdlib fs +
        existing config). §17 guardrail (never clobber/delete) is the adopt-in-place rule."

# READ-ONLY — the input sibling PRP (defines the store INPUT: resolveStore → absolute)
- file: plan/002_38acb6d28a6a/P1M2T2S1/PRP.md
  why: "Defines resolveStore(haveStore) → (absStore, error), which absolutizes via filepath.Abs
        and is what run()/S3 passes as setupStore's `store`. Confirms S1 appends after skillPath
        and added the config+bufio imports this subtask REUSES (zero new main.go imports here).
        c.initStore (from P1.M2.T1.S1) is the haveStore resolveStore receives — NOT setupStore's
        concern."

# READ-ONLY — PRD (source of truth for the init create/seed/writeconfig contract + the template)
- file: PRD.md
  why: "§8.2 (h3.9) = the authoritative step 2 (mkdir -p), step 3 (seed-if-empty / adopt-if-not,
        string constant NOT go:embed), step 4 (write config with absolute store). §11 (h2.10) =
        the example SKILL.md body to embed VERBATIM (skilldozer, not skpp). §17 (h2.16) =
        guardrail 'never clobber or delete existing files'. §8.1 (h3.8) = the config format
        (`store: /abs`)."
  section: "h3.9 (§8.2 steps 2-4), h2.10 (§11 the template), h2.16 (§17 guardrail), h3.8 (§8.1)."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer && ls *.go
main.go          # EDIT: APPEND const exampleSkillTemplate + func setupStore after resolveStore (line 876). ZERO new imports.
main_test.go     # EDIT: +1 import (internal/config); +4 TestSetupStore* unit tests
internal/        # untouched (config.Save/File/Load CONSUMED, not modified)
# go.mod / go.sum untouched (stdlib fs + existing config — no new deps)
$ grep -n 'setupStore\|exampleSkillTemplate' main.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
main.go          # APPEND: const exampleSkillTemplate (PRD §11 compiled-in seed) + func setupStore (mkdir/seed/writeconfig)
main_test.go     # APPEND: TestSetupStore* (4) — empty-seeds, non-empty-adopts, idempotent, MkdirAll-failure
```

**No new files.** All edits are additive to existing files.

| File | Responsibility |
|---|---|
| `main.go` | The store-CREATION logic (mkdir → seed-if-empty/adopt-if-not → write config) + the compiled-in example-skill seed template. `setupStore` takes both targets as injected strings, so it is directly unit-testable AND directly callable by run()/S3 (no separate wrapper, unlike S1's chooseStore/resolveStore). |
| `main_test.go` | Lock the 4 contract behaviors (seed/adopt/idempotent/error) via temp-dir fs fixtures + `config.Load` round-trip assertions. |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL — load-bearing for the template) — Go raw string literals
// (delimited by backticks) CANNOT contain a backtick character. The PRD §11 example
// SKILL.md body contains EIGHT backticks: 2 inline (`skilldozer`) + 6 fence (```bash
// open + ``` close). So the template CANNOT be one raw literal. The idiomatic,
// no-go:embed workaround is to SPLICE double-quoted backtick runs between raw segments
// via `+` concatenation. A `const` string expression of `lit + lit + …` IS valid Go
// (all operands are constant string literals), so this stays a `const`. See verified_facts
// §5 for the EXACT spliced source. Do NOT use //go:embed (PRD §17 / G11 forbid it).
//
// GOTCHA #2 — the seed template is DUPLICATED on purpose. There are two copies of the
// PRD §11 body after this plan: (a) the on-disk skills/example/SKILL.md (P1.M3.T1.S1's
// repo asset) and (b) this task's compiled-in `exampleSkillTemplate` const. They are
// intentionally separate (the const does NOT read the on-disk file — that would be
// go:embed in disguise). Both MUST equal PRD §11 verbatim. The const's doc comment cites
// PRD §11 + P1.M3.T1.S1 so a future editor keeps them in sync.

// GOTCHA #3 — "empty store" means ZERO os.ReadDir entries of ANY kind (file, dir,
// dotfile) — NOT "no SKILL.md at any depth" (that is skillsdir.HasSkillMD, S1's
// cwd-auto-detect concern, a different question). A store with a single pre-existing
// file (even a non-skill file or a .gitkeep) ⇒ ADOPT in place, do NOT seed. This is
// what the contract test "a dir containing a file leaves it intact (seeded=false)"
// locks. Use:
//   entries, err := os.ReadDir(store)
//   if len(entries) == 0 { /* seed */ }

// GOTCHA #4 (CRITICAL — §17 guardrail) — on a non-empty store do NOTHING to existing
// files: no delete, no overwrite, no re-seed. Just fall through to config.Save. The
// idempotency test proves example/SKILL.md is byte-identical across re-runs (a re-run
// sees a non-empty store → adopt → never touches the seeded file).

// GOTCHA #5 — setupStore must NOT absolutize `store`. S1's resolveStore already called
// filepath.Abs (main.go:871) and returns the absolute path; setupStore receives it and
// writes config.Store = store VERBATIM. Do not call filepath.Abs in setupStore (it is
// resolveStore's job; duplicating it is harmless but wrong-by-factoring and would mask
// a caller bug). Likewise setupStore must NOT call config.Path() itself — configPath is
// an injected param (run()/S3 obtains it) so the function is unit-testable with a temp
// config path. BOTH deps are strings ⇒ directly testable, no wrapper layer.

// GOTCHA #6 — `seeded` is a SUCCESS-PATH signal (it tells run()/S3 which message to
// print: "seeded example skill" vs "adopted existing store"). On ANY fs error, return
// (false, err) — the conventional Go "zero value on error" — because the caller checks
// `err` FIRST and exits 1, never reading `seeded`. This means even if config.Save fails
// AFTER a successful seed, return (false, err), NOT (true, err). Document this in the
// doc comment so a reader isn't confused by "seeded then returned false".

// GOTCHA #7 — main.go needs ZERO new imports. os, path/filepath, fmt, and
// internal/config are ALL already imported (S1 added config+bufio; os/filepath/fmt were
// pre-existing). `grep -c 'internal/config' main.go` == 1. Do NOT touch the import block.
// Only main_test.go needs +1 import (internal/config) for the config.Load assertions.

// GOTCHA #8 — APPEND after resolveStore (the file tail, line 876). This is purely
// additive at the END — non-overlapping with S3's mid-file run() dispatch edit (~line
// 408+) and with the import block. Mirror S1's "append at the tail" discipline so the
// changeset composes cleanly with S3 regardless of merge order.

// GOTCHA #9 — setupStore is defined here but left UNCALLED (S3 wires it into run()'s
// `if c.init`). Go ALLOWS unused package-level functions (only unused imports and unused
// LOCAL variables are compile errors), so `go build`/`go vet`/`go test` are green with
// setupStore uncalled. Do NOT add a throwaway call to "use" it; do NOT add the
// `if c.init` dispatch (that is S3's scope and would collide with it). Exactly as S1's
// resolveStore was defined-and-uncalled until S3.

// GOTCHA #10 — the example/ subdir is created with os.MkdirAll(filepath.Join(store,
// "example"), 0o755) BEFORE os.WriteFile of the SKILL.md. Two steps: mkdir the example
// level (the store level was already MkdirAll'd at the top), then write the file. Do not
// WriteFile into a dir you have not MkdirAll'd (would fail with fs.ErrNotExist). os.WriteFile
// does not create parent dirs (unlike config.Save, which MkdirAll's ITS parent — but that is
// the config FILE's parent, a different path).

// GOTCHA #11 — the MkdirAll-failure error test uses a regular FILE at the store path
// (deterministic, portable, no root-skip): os.MkdirAll("/tmp/x/notadir") where `notadir`
// is an existing regular file returns a *PathError ("not a directory"). Assert err != nil
// AND that no config.yaml was written (the failure precedes config.Save). Do NOT use a
// 0500-permission parent for the error test — that needs an os.Geteuid()==0 skip on CI.
```

---

## Implementation Blueprint

### Data models and structure

**No new types.** The only new "model" is the seed template, expressed as a package-level `const` string. No struct changes (the `config.init`/`initStore` fields are P1.M2.T1.S1's; the `config.File{Store}` value is constructed inline at the `config.Save` call site).

```go
// exampleSkillTemplate is the PRD §11 example skill body, compiled into the binary as a
// STRING CONSTANT (NOT go:embed — PRD §17 "nothing about the user's collection is compiled
// in"; code_prd_delta.md G11). skilldozer init writes this verbatim into an EMPTY store's
// example/SKILL.md (PRD §8.2 step 3). NOTE: there is a SECOND copy of this exact text on
// disk at skills/example/SKILL.md (P1.M3.T1.S1's repo asset); both MUST equal PRD §11.
//
// Raw literals can't hold backticks; the §11 body has 8 (2 inline `skilldozer` + the
// ```bash fence). Splice double-quoted backtick runs between raw segments via `+`.
const exampleSkillTemplate = `---
name: example
description: >
  Reference example skill for skilldozer. Demonstrates the required frontmatter and
  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skilldozer]
  category: meta
---

# Example Skill

This skill exists only so ` + "`skilldozer`" + ` has something to resolve.

Try:

` + "```bash" + `
skilldozer example                       # prints this directory's absolute path
skilldozer -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skilldozer example)"       # loads this skill into pi
` + "```" + `
`
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main_test.go — add the internal/config import (needed by the Task 5 tests)
  - FILE: main_test.go (import block, lines 3-11)
  - CURRENTLY a single stdlib group: bytes, errors, io, os, path/filepath, strings, testing.
  - ADD a SECOND group (blank line + the internal import), mirroring main.go's grouping:
      "testing"

      "github.com/dabstractor/skilldozer/internal/config"
    )
  - DO NOT touch any existing import line. This is the ONLY import change in the whole subtask.
  - VERIFY after Task 5: `go build ./...` compiles (config.Load is referenced by the tests).

Task 2: APPEND main.go — const exampleSkillTemplate (the PRD §11 compiled-in seed)
  - FILE: main.go (APPEND immediately after resolveStore's closing brace, line 876)
  - ADD the const from the "Data models" block above VERBATIM. It is a `const` (not var) —
    matches usageText @52's pattern (the repo's existing big-compiled-in-string-const home).
  - GOTCHA #1: the backtick splices (`+ "`skilldozer`" +`, `+ "```bash" +`, `+ "```" +`) are
    MANDATORY — a single raw literal cannot hold the 8 backticks. Do NOT simplify to one
    raw string (it will not compile / will drop the fences).
  - GOTCHA #2: the doc comment cites PRD §11 + §17 + G11 + P1.M3.T1.S1 (the on-disk twin).
  - The constant ends with the closing fence + a trailing newline (raw segment `\n` after
    the last splice). This is a standard text file shape.

Task 3: APPEND main.go — func setupStore (the testable core; mkdir/seed/writeconfig)
  - FILE: main.go (APPEND immediately after the exampleSkillTemplate const — the file's new tail)
  - ADD (GOTCHA #3/#4/#5/#6/#10 — empty=ReadDir len 0; never clobber; no Abs; (false,err)
    on error; MkdirAll example before WriteFile):
      // setupStore creates the skills store, seeds it if empty, and writes the config. It is
      // the create+seed+writeconfig half of `skilldozer init` (PRD §8.2 steps 2-4); the
      // store-CHOICE half is resolveStore (P1.M2.T2.S1), and run()'s `if c.init` dispatch
      // (P1.M2.T2.S3) calls both. Both targets are INJECTED as strings (store is already
      // absolute — resolveStore absolutized it; configPath is config.Path()'s result from
      // run()), so this function is directly unit-testable with temp paths and needs no
      // separate wrapper layer.
      //
      // Steps:
      //  (a) os.MkdirAll(store, 0o755) — create the store dir if missing (PRD §8.2 step 2).
      //  (b) os.ReadDir(store): if the dir is EMPTY (zero entries of any kind), seed
      //      example/SKILL.md from the compiled-in exampleSkillTemplate (PRD §8.2 step 3,
      //      §11) — MkdirAll(store/example) then WriteFile; seeded=true. "Empty" means no
      //      entries at all, NOT "no SKILL.md" (a single pre-existing file ⇒ adopt).
      //  (c) If the store already contains ANYTHING, adopt it in place: NEVER clobber or
      //      delete existing files (PRD §17 guardrail); seeded stays false.
      //  (d) config.Save(configPath, config.File{Store: store}) — write the config with the
      //      absolute store path (PRD §8.2 step 4). ALWAYS runs, whether seeded or adopted.
      //
      // Returns (seeded, nil) on success, or (false, err) on any fs failure — `seeded` is a
      // SUCCESS-PATH signal (run()/S3 prints "seeded" vs "adopted"); callers MUST check err
      // before reading seeded, so a config.Save failure after a successful seed still returns
      // (false, err). Errors are wrapped with a "skilldozer init: <step>: %w" prefix.
      func setupStore(store, configPath string) (seeded bool, err error) {
          // (a) Ensure the store dir exists (idempotent — no-op if present).
          if err := os.MkdirAll(store, 0o755); err != nil {
              return false, fmt.Errorf("skilldozer init: create store dir %q: %w", store, err)
          }
          // (b) Seed only if the store is EMPTY (zero entries of any kind).
          entries, err := os.ReadDir(store)
          if err != nil {
              return false, fmt.Errorf("skilldozer init: read store dir %q: %w", store, err)
          }
          if len(entries) == 0 {
              exampleDir := filepath.Join(store, "example")
              if err := os.MkdirAll(exampleDir, 0o755); err != nil {
                  return false, fmt.Errorf("skilldozer init: create example dir: %w", err)
              }
              if err := os.WriteFile(filepath.Join(exampleDir, "SKILL.md"), []byte(exampleSkillTemplate), 0o644); err != nil {
                  return false, fmt.Errorf("skilldozer init: seed example SKILL.md: %w", err)
              }
              seeded = true
          }
          // (c) Non-empty: adopt in place. Do NOTHING to existing files (PRD §17). seeded stays false.
          // (d) Always write the config with the (already-absolute) store path.
          if err := config.Save(configPath, config.File{Store: store}); err != nil {
              return false, fmt.Errorf("skilldozer init: write config %q: %w", configPath, err)
          }
          return seeded, nil
      }

Task 4: VERIFY (isolated, then whole-module + invariants) — run AFTER Task 5 too
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0 (setupStore uncalled is fine — pkg-level fn)
  - go build ./...                    # exit 0
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #7

Task 5: EDIT main_test.go — add the 4 setupStore unit tests
  - FILE: main_test.go (APPEND a new block; mirror house style: t.Helper where apt,
    t.TempDir(), filepath.Join, os.MkdirAll/WriteFile for fixtures, config.Load for the
    round-trip assertion — cf. config_test.go TestSaveLoadRoundTrip @31).
  - GROUP near the end (or near S1's TestChooseStore* block — different function names,
    no collision). Package is `main` (white-box), so setupStore + exampleSkillTemplate are
    directly accessible.
  - (5a) Test #1 — empty dir seeds + writes config (locks the template bytes byte-for-byte):
      func TestSetupStoreEmptyDirSeedsExampleAndWritesConfig(t *testing.T) {
          store := t.TempDir() // empty
          cfg := filepath.Join(t.TempDir(), "config.yaml")
          seeded, err := setupStore(store, cfg)
          if err != nil {
              t.Fatalf("setupStore(empty): %v; want nil", err)
          }
          if !seeded {
              t.Errorf("setupStore(empty): seeded=false; want true")
          }
          // example/SKILL.md exists with the template bytes EXACTLY (catches a botched splice).
          got, err := os.ReadFile(filepath.Join(store, "example", "SKILL.md"))
          if err != nil {
              t.Fatalf("read seeded example/SKILL.md: %v", err)
          }
          if string(got) != exampleSkillTemplate {
              t.Errorf("seeded example/SKILL.md != exampleSkillTemplate:\ngot:\n%s\nwant:\n%s", got, exampleSkillTemplate)
          }
          // config written with store=<abs> verbatim (round-trip via config.Load).
          f, err := config.Load(cfg)
          if err != nil {
              t.Fatalf("config.Load: %v", err)
          }
          if f.Store != store {
              t.Errorf("config.Store=%q; want %q", f.Store, store)
          }
      }
  - (5b) Test #2 — non-empty dir adopts in place + still writes config (locks §17 never-clobber):
      func TestSetupStoreNonEmptyDirAdoptsInPlaceAndWritesConfig(t *testing.T) {
          store := t.TempDir()
          preExisting := filepath.Join(store, "mynotes.md") // a non-skill file
          if err := os.WriteFile(preExisting, []byte("# my stuff\n"), 0o644); err != nil {
              t.Fatalf("seed fixture: %v", err)
          }
          cfg := filepath.Join(t.TempDir(), "config.yaml")
          seeded, err := setupStore(store, cfg)
          if err != nil {
              t.Fatalf("setupStore(non-empty): %v; want nil", err)
          }
          if seeded {
              t.Errorf("setupStore(non-empty): seeded=true; want false (adopt in place)")
          }
          // the pre-existing file is byte-intact (never clobbered).
          got, err := os.ReadFile(preExisting)
          if err != nil {
              t.Fatalf("read pre-existing: %v", err)
          }
          if string(got) != "# my stuff\n" {
              t.Errorf("pre-existing file changed: %q; want %q", got, "# my stuff\n")
          }
          // NO example/ dir was created.
          if _, err := os.Stat(filepath.Join(store, "example")); !os.IsNotExist(err) {
              t.Errorf("example/ must NOT be created in a non-empty store; stat err=%v", err)
          }
          // config still written.
          f, err := config.Load(cfg)
          if err != nil {
              t.Fatalf("config.Load: %v", err)
          }
          if f.Store != store {
              t.Errorf("config.Store=%q; want %q", f.Store, store)
          }
      }
  - (5c) Test #3 — idempotent re-run (locks adopt-on-second-run + no clobber of the seed):
      func TestSetupStoreIdempotent(t *testing.T) {
          store := t.TempDir()
          cfg := filepath.Join(t.TempDir(), "config.yaml")
          // first run: empty -> seed.
          seeded1, err := setupStore(store, cfg)
          if err != nil || !seeded1 {
              t.Fatalf("first run: (%v,%v); want (true,nil)", seeded1, err)
          }
          first, err := os.ReadFile(filepath.Join(store, "example", "SKILL.md"))
          if err != nil {
              t.Fatalf("read after first run: %v", err)
          }
          // second run: store now non-empty -> adopt, no clobber, rewrite config (idempotent).
          seeded2, err := setupStore(store, cfg)
          if err != nil {
              t.Fatalf("second run: %v; want nil", err)
          }
          if seeded2 {
              t.Errorf("second run: seeded=true; want false (store already has content)")
          }
          second, err := os.ReadFile(filepath.Join(store, "example", "SKILL.md"))
          if err != nil {
              t.Fatalf("read after second run: %v", err)
          }
          if string(first) != string(second) {
              t.Errorf("idempotent re-run changed example/SKILL.md:\nfirst:\n%s\nsecond:\n%s", first, second)
          }
          // config still valid after the rewrite.
          f, err := config.Load(cfg)
          if err != nil {
              t.Fatalf("config.Load after re-run: %v", err)
          }
          if f.Store != store {
              t.Errorf("config.Store=%q; want %q", f.Store, store)
          }
      }
  - (5d) Test #4 — MkdirAll failure returns a wrapped error and writes NO config (GOTCHA #11):
      func TestSetupStoreMkdirAllFailureReturnsWrappedError(t *testing.T) {
          // Make the store path a regular FILE: os.MkdirAll on an existing non-dir fails.
          parent := t.TempDir()
          store := filepath.Join(parent, "notadir")
          if err := os.WriteFile(store, []byte("x"), 0o644); err != nil {
              t.Fatalf("fixture: %v", err)
          }
          cfg := filepath.Join(t.TempDir(), "config.yaml")
          seeded, err := setupStore(store, cfg)
          if err == nil {
              t.Fatalf("expected MkdirAll error; got (%v,nil)", seeded)
          }
          if seeded {
              t.Errorf("on error: seeded=true; want false")
          }
          // the failure precedes config.Save, so no config.yaml must exist.
          if _, err := os.Stat(cfg); !os.IsNotExist(err) {
              t.Errorf("config must NOT be written on MkdirAll failure; stat err=%v", err)
          }
      }
  - GOTCHA: all 4 tests reference setupStore/exampleSkillTemplate (main package, white-box)
    and config.Load (the Task 1 import). No other symbols. No t.Chdir, no terminal mocking
    (both deps are injected strings).

Task 6: VERIFY (whole-module + invariants — re-run after Task 5)
  - gofmt -l main.go main_test.go     # nothing
  - go vet ./...                      # exit 0
  - go test ./...                     # ALL pass incl. the 4 new tests; existing unaffected
  - git diff --quiet go.mod go.sum && echo deps unchanged
  - manual: go test -run TestSetupStore -v ./...   # the 4 tests named + green
```

### Implementation Patterns & Key Details

```go
// exampleSkillTemplate — compiled-in PRD §11 seed (NOT go:embed). Backticks spliced.
const exampleSkillTemplate = `---
name: example
description: >
  Reference example skill for skilldozer. Demonstrates the required frontmatter and
  how skilldozer resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skilldozer]
  category: meta
---

# Example Skill

This skill exists only so ` + "`skilldozer`" + ` has something to resolve.

Try:

` + "```bash" + `
skilldozer example                       # prints this directory's absolute path
skilldozer -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skilldozer example)"       # loads this skill into pi
` + "```" + `
`

// setupStore — mkdir → seed-if-empty/adopt-if-not → write config. Both deps injected
// (store already absolute; configPath from run()). (false, err) on any fs failure.
func setupStore(store, configPath string) (seeded bool, err error) {
	if err := os.MkdirAll(store, 0o755); err != nil {
		return false, fmt.Errorf("skilldozer init: create store dir %q: %w", store, err)
	}
	entries, err := os.ReadDir(store)
	if err != nil {
		return false, fmt.Errorf("skilldozer init: read store dir %q: %w", store, err)
	}
	if len(entries) == 0 {
		exampleDir := filepath.Join(store, "example")
		if err := os.MkdirAll(exampleDir, 0o755); err != nil {
			return false, fmt.Errorf("skilldozer init: create example dir: %w", err)
		}
		if err := os.WriteFile(filepath.Join(exampleDir, "SKILL.md"), []byte(exampleSkillTemplate), 0o644); err != nil {
			return false, fmt.Errorf("skilldozer init: seed example SKILL.md: %w", err)
		}
		seeded = true
	}
	// Non-empty: adopt in place — never clobber/delete (PRD §17). seeded stays false.
	if err := config.Save(configPath, config.File{Store: store}); err != nil {
		return false, fmt.Errorf("skilldozer init: write config %q: %w", configPath, err)
	}
	return seeded, nil
}
```

Notes easy to get wrong:
- The `const` MUST use the `+`-spliced form. A single raw literal `const x = `…`` will either fail to compile (if you try to escape) or silently drop the code fences. Test #1's `string(got) != exampleSkillTemplate` byte-equality check is what catches a botched template.
- `len(entries) == 0` is the empty check — NOT "no SKILL.md". A store with one stray file adopts. Test #2 locks this.
- `setupStore` does not absolutize and does not call `config.Path()` — both are the caller's job. This is what makes it a clean, injectable unit.
- On the `config.Save` error path the function returns `(false, err)` even though the seed may have succeeded on disk; the caller checks `err` first. Documented in the doc comment.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **`setupStore` takes `store` + `configPath` as injected strings (no wrapper layer).** Unlike S1's chooseStore/resolveStore pair (which needed factoring because the prompt is terminal I/O), setupStore's only I/O is filesystem and both target paths are strings — so the same function is both the testable core AND the thing run() calls directly. No `resolveStore`-style indirection.
2. **The seed template is a `const` in main.go (not `var`, not `internal/config`).** Matches `usageText` @52 (the repo's existing big-compiled-in-string-const pattern). `const` is valid because the `+`-spliced expression is all constant string literals. main.go is the home (co-located with setupStore and the other init code); `internal/config` is rejected because config is the settings sidecar, not a skill-asset home. `go:embed` is forbidden (PRD §17 / G11).
3. **Empty = `os.ReadDir` returns zero entries (any kind).** Not "no SKILL.md". The contract test "a dir containing a file leaves it intact" demands that a single pre-existing file ⇒ adopt. `len(entries)==0` captures exactly that and is the conservative, never-clobber reading of §8.2 step 3.
4. **Adopt-in-place does literally nothing to existing files.** No delete, no overwrite, no re-seed. Only config.Save runs. This is the §17 guardrail; the idempotency test proves the seeded file is untouched on re-run.
5. **`(false, err)` on every error path.** `seeded` is a success-only signal; the zero-value-on-error convention keeps it unambiguous. The doc comment states callers must check `err` first.
6. **No absolutization in setupStore.** resolveStore (S1) already absolutized; duplicating `filepath.Abs` here would mask a caller bug and muddy the factoring.
7. **The error test uses a regular file at the store path, not a 0500-permission parent.** Deterministic and portable (no `os.Geteuid()==0` skip needed on root-CI). `os.MkdirAll` on an existing non-directory returns a `*PathError`.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod UNCHANGED. setupStore uses only already-imported symbols:
      os (MkdirAll/ReadDir/WriteFile), path/filepath (Join), fmt (Errorf), and
      internal/config (Save/File) — ALL already in main.go's import block (S1 added config).
    main_test.go adds ONE import (internal/config) for the config.Load assertions.
    No `go get`, no `go mod tidy`. git diff --quiet go.mod go.sum ⇒ "deps unchanged".

CONSUMERS (NOT built in this subtask — listed to fix the interface):
  - run() init dispatch (P1.M2.T2.S3): `if c.init { cfg, err := config.Path(); if err != nil
    {…exit 1}; abs, err := resolveStore(c.initStore); if err != nil {…exit 1}; seeded, err :=
    setupStore(abs, cfg); if err != nil {…exit 1}; … print --path + check (seeded-aware msg) }`.
    setupStore(abs, cfg) is the call site; abs comes from S1's resolveStore; cfg from config.Path.

CONSUMED (already present — verified):
  - config.Save(path string, f File) error — internal/config/config.go:69 (P1.M1.T1.S1).
  - config.File{Store string} — internal/config/config.go:30.
  - config.Load(path string) (File, error) — internal/config/config.go:52 (test-side).
  - resolveStore(haveStore string) (string, error) — main.go:843-876 (P1.M2.T2.S1, returns ABS).

NO ROUTES / NO DATABASE / NO CONFIG-FIELD-ADDITIONS / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after editing main.go)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0 (setupStore uncalled is fine — pkg-level fn)
go build ./...                  # expect exit 0
# Expected: zero output / exit 0.
```

### Level 2: Unit Tests (component validation — the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test ./... -run 'TestSetupStore' -v
# Expected: ALL 4 pass. The load-bearing assertions:
#   TestSetupStoreEmptyDirSeedsExampleAndWritesConfig
#       -> seeded=true; example/SKILL.md bytes == exampleSkillTemplate (locks the splice, GOTCHA #1);
#          config.Load(cfg).Store == store.
#   TestSetupStoreNonEmptyDirAdoptsInPlaceAndWritesConfig
#       -> seeded=false; pre-existing file byte-intact; NO example/ dir; config written (§17 never-clobber).
#   TestSetupStoreIdempotent
#       -> run1 (true,nil); run2 (false,nil); example/SKILL.md byte-identical; config valid.
#   TestSetupStoreMkdirAllFailureReturnsWrappedError
#       -> err != nil; seeded=false; NO config.yaml written (failure precedes config.Save).

# Regression: the existing suite (parseArgs/run/path/list/search/check/all/help + S1's
# TestChooseStore*) stays green.
go test ./...   # expect exit 0 (purely additive — no symbol renamed/moved)
```

### Level 3: Whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0

# GOTCHA #7 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Manual: the 2 new symbols exist and compile (setupStore is uncalled — that is correct
# scaffolding until P1.M2.T2.S3 wires it into run()).
go doc . exampleSkillTemplate setupStore 2>/dev/null | head -40
```

### Level 4: Creative & Domain-Specific Validation

```bash
# N/A for THIS subtask — there is no end-to-end init flow to exercise yet (run() dispatch is
# P1.M2.T2.S3). The 4 unit tests ARE the validation surface: they prove the create/seed/
# writeconfig contract + the §17 never-clobber guardrail + idempotency + the error path
# without a real config-file location or terminal. The integration proof (a real
# `skilldozer init --store /tmp/x` writing config.yaml and seeding example/SKILL.md, and
# `skilldozer init` adopting a cwd skills repo) lands in P1.M4.T1.S1's §13 acceptance run,
# which calls setupStore through run()'s S3 dispatch.

# Optional sanity check (non-blocking): confirm the seeded template parses as the on-disk
# example skill format the discover package expects — but Test #1's byte-equality to the
# §11 body already guarantees this, so it is redundant.
```

---

## Final Validation Checklist

### Technical Validation

- [ ] All validation levels completed successfully
- [ ] All tests pass: `go test ./...`
- [ ] No vet errors: `go vet ./...`
- [ ] No formatting issues: `gofmt -l main.go main_test.go` (empty)
- [ ] go.mod/go.sum unchanged: `git diff --quiet go.mod go.sum`

### Feature Validation

- [ ] `setupStore(emptyTmp, cfg)` ⇒ `(true, nil)`; `emptyTmp/example/SKILL.md` bytes == `exampleSkillTemplate`; `config.Load(cfg).Store == emptyTmp`
- [ ] `setupStore(dirWithAFile, cfg)` ⇒ `(false, nil)`; file intact; no `example/` dir; config written
- [ ] `setupStore` run twice ⇒ run1 `(true,nil)`, run2 `(false,nil)`; `example/SKILL.md` byte-identical; config valid
- [ ] `setupStore(pathThatIsAFile, cfg)` ⇒ `(false, err)`; `cfg` not written
- [ ] `exampleSkillTemplate` equals PRD §11 verbatim (skilldozer, not skpp; backtick fences intact)
- [ ] `setupStore` defined, compiles, and is uncalled (wired by P1.M2.T2.S3)

### Code Quality Validation

- [ ] `exampleSkillTemplate` mirrors `usageText`'s compiled-in-const pattern (const, raw-string + backtick splices)
- [ ] `setupStore` takes both targets as injected strings (no `os.Getwd`/`config.Path`/`filepath.Abs` inside it)
- [ ] Empty check is `len(os.ReadDir(store)) == 0` (ANY entry ⇒ adopt), not "no SKILL.md"
- [ ] Adopt-in-place does nothing to existing files (§17 never-clobber)
- [ ] Every fs error returns `(false, fmt.Errorf("skilldozer init: <step>: %w", err))`
- [ ] Doc comments cite PRD §8.2 steps 2-4, §11, §17, and G11
- [ ] Anti-patterns avoided (see below)
- [ ] No new dependencies; stdlib fs + existing `config` only

### Documentation & Deployment

- [ ] No doc files (the seed template is a runtime string synced in P1.M3.T1.S1; README init UX is P1.M4.T2.S1)
- [ ] No new environment variables

---

## Anti-Patterns to Avoid

- ❌ Don't use `//go:embed` for the template — PRD §17 / G11 forbid it; it must be a compiled-in string `const`. Use the `+`-spliced raw-string form (GOTCHA #1).
- ❌ Don't write the template as a single raw string literal — it cannot hold the 8 backticks in §11; the fences would be dropped. The splices are mandatory and Test #1 locks the bytes.
- ❌ Don't define "empty" as "no SKILL.md" — it is `len(os.ReadDir(store)) == 0`. A single stray file ⇒ adopt (Test #2). Confusing this with `skillsdir.HasSkillMD` (S1's cwd-auto-detect) is the classic mix-up.
- ❌ Don't clobber, delete, or re-seed an existing store — adopt in place (§17). The idempotency test enforces this (Test #3).
- ❌ Don't absolutize `store` or call `config.Path()` inside setupStore — both are the caller's job (resolveStore / run()). setupStore takes both as injected strings (GOTCHA #5).
- ❌ Don't return `(true, err)` on the config.Save-after-seed path — return `(false, err)`; `seeded` is success-only and the caller checks `err` first (GOTCHA #6).
- ❌ Don't `WriteFile` into `store/example/` without first `os.MkdirAll(store/example)` — `WriteFile` does not create parent dirs (GOTCHA #10). (config.Save MkdirAll's the CONFIG file's parent, a different path.)
- ❌ Don't add imports to main.go — `os`/`path/filepath`/`fmt`/`internal/config` are all already imported (S1). Only main_test.go needs +`internal/config` (GOTCHA #7).
- ❌ Don't add `if c.init { … }` to `run()` — that is P1.M2.T2.S3 (GOTCHA #9). `setupStore` is defined and uncalled here, exactly as S1's `resolveStore` was.
- ❌ Don't edit `skills/example/SKILL.md` (on-disk) or `internal/config` — those are sibling subtasks (P1.M3.T1.S1 / config is consumed not modified).
- ❌ Don't run `go mod tidy` — there are no new deps; it would be a no-op but could touch go.sum needlessly (GOTCHA #7).

---

## Confidence Score

**9/10** — one-pass implementation success likelihood. The change is purely additive to two files and consumes three already-verified APIs (`config.Save(path, File)`, `config.File{Store}`, `config.Load(path)`) whose exact signatures were read from source, plus S1's already-landed `resolveStore` (which supplies the absolute `store`). main.go needs ZERO new imports (S1 already added `internal/config`). The logic is small (MkdirAll → ReadDir → seed-if-empty → config.Save) and mirrors the repo's "injected deps ⇒ directly testable" convention. The single non-obvious risk — the backtick-in-raw-string gotcha — is resolved with an exact spliced `const` whose bytes are locked byte-for-byte by Test #1 (`string(got) != exampleSkillTemplate`). The §17 never-clobber rule and the empty=`ReadDir len 0` semantics are locked by Tests #2 and #3. The one residual uncertainty is whether an implementer might confuse "empty store" with "no SKILL.md" (skillsdir.HasSkillMD, S1's concern) — the GOTCHA #3 callout and Test #2's "a dir containing a [non-skill] file adopts" assertion close that gap.
