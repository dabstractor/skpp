# Verified Facts — P1.M2.T2.S2 (setupStore + seed template + write config)

> All signatures/behaviors below were read DIRECTLY from source at PRP-write time
> (not from memory or a subagent summary). Line numbers are current; they shift as
> sibling subtasks land — locate symbols by NAME, not line.

## 1. INPUT contract (what exists when this subtask starts)

This subtask consumes TWO already-present things (verified):

### 1a. `resolveStore(haveStore string) (string, error)` — main.go:843–876 (P1.M2.T2.S1)
- Returns the **absolute** chosen store dir (it calls `filepath.Abs` on chooseStore's
  verbatim choice — main.go:871). **This is the `store` argument setupStore receives.**
- `store` is therefore ALREADY absolute when setupStore runs. setupStore must NOT
  absolutize again (resolveStore owns that). setupStore writes `config.Store = store`
  verbatim.
- Status: S1 is "Ready" in plan_status but its CODE IS ALREADY LANDED in main.go
  (stdinIsTerminal/readPrompt/chooseStore/resolveStore all present — verified by
  `grep -n '^func ' main.go`). Treat the S1 PRP as a contract; the code matches it.

### 1b. The config package — internal/config/config.go (P1.M1.T1.S1, Complete)
EXACT signatures (read from source — these are what setupStore calls):

```go
// config.go:30-32
type File struct {
    Store string `yaml:"store,omitempty"`
}

// config.go:69 — creates parent dirs (MkdirAll 0o755), writes file 0o644
func Save(path string, f File) error          // NOTE: (path, File) order — path FIRST

// config.go:52
func Load(path string) (File, error)          // lenient: unknown keys ignored; missing→fs.ErrNotExist (VERBATIM, unwrapped)

// config.go:131
func Path() (string, error)                    // $SKILLDOZER_CONFIG literal, else $XDG_CONFIG_HOME/skilldozer/config.yaml
```

- setupStore writes config via: `config.Save(configPath, config.File{Store: store})`
  — matches the item_description contract call EXACTLY (path first, File second).
- `configPath` itself is NOT obtained inside setupStore — it is a PARAMETER (run()/S3
  obtains it via `config.Path()`). This keeps setupStore a pure(two-strings-in) function
  that is unit-testable by injecting a temp config path (mirrors S1's chooseStore factoring).

## 2. main.go import block — ZERO new imports needed (verified)

```go
// main.go:14-27 (current)
import (
    "bufio"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/dabstractor/skilldozer/internal/check"
    "github.com/dabstractor/skilldozer/internal/config"   // ← ALREADY PRESENT (S1 added it)
    "github.com/dabstractor/skilldozer/internal/discover"
    "github.com/dabstractor/skilldozer/internal/resolve"
    "github.com/dabstractor/skilldozer/internal/search"
    "github.com/dabstractor/skilldozer/internal/skillsdir"
    "github.com/dabstractor/skilldozer/internal/ui"
)
```

- setupStore needs: `os` (MkdirAll/ReadDir/WriteFile), `path/filepath` (Join),
  `fmt` (Errorf), `config` (Save/File). **ALL FOUR are already imported.** `grep -c
  'internal/config' main.go` == 1. So main.go gets **ZERO new import lines**.
- This differs from S1 (which added bufio+config). Here, S1 already did the import work.

## 3. main_test.go import block — +1 import (config) needed

```go
// main_test.go:3-11 (current — stdlib ONLY, no internal group yet)
import (
    "bytes"
    "errors"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"
)
```

- The setupStore tests assert `config.Load(cfg).Store == store` (semantic round-trip,
  same approach config_test.go uses). That requires importing config. ADD a second
  import group (blank line + the internal import), mirroring main.go's grouping:
  ```go
      "testing"

      "github.com/dabstractor/skilldozer/internal/config"
  )
  ```
- `errors` is already imported (S1 added it for TestChooseStorePropagatesPromptError);
  the setupStore tests do NOT need errors (the error-path test asserts `err != nil`,
  not `errors.Is`). So the ONLY test-side import change is +config.

## 4. APPEND location — after resolveStore (file tail, ~line 876)

- `func resolveStore` is the LAST function in main.go (starts :843, ends :876 — verified
  `awk` shows the closing `}` at line 876; `wc -l main.go` == 876). APPEND the new
  `exampleSkillTemplate` const + `setupStore` func immediately AFTER line 876.
- This is purely additive at the FILE TAIL — exactly S1's discipline. It does NOT touch
  the import block (§2) and does NOT touch run() (S3's mid-file edit at ~line 408+). So
  this changeset composes cleanly with S3 regardless of merge order.

## 5. CRITICAL GOTCHA — raw string literals CANNOT contain backticks

Go raw string literals (delimited by `` ` ``) cannot contain the backtick character.
The PRD §11 example SKILL.md content contains **8 backticks**:
- 2 inline: `` `skilldozer` `` in the line "This skill exists only so `skilldozer` has…"
- 6 fence: the opening ```` ```bash ```` (3) and the closing ```` ``` ```` (3) of the
  bash code block.

Therefore the template CANNOT be a single raw string literal. The idiomatic Go
workaround (no `go:embed` — forbidden by contract/PRD §17 "nothing about the user's
collection is compiled in") is to **splice double-quoted backtick runs between raw
string segments via `+` concatenation**. A `const` string expression of `lit + lit + …`
is valid Go (all operands are constant string literals), so this stays a `const`:

```go
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

- `+ "`skilldozer`" +` splices the inline code (one backtick each side of the word).
- `+ "```bash" +` and `+ "```" +` splice the fenced block's opening/closing fence.
- The constant ends with the closing fence + a trailing newline (raw segment `` `\n ``).
  This matches a standard text file (trailing newline) and the on-disk
  `skills/example/SKILL.md` shape.

### Why NOT go:embed (forbidden — the load-bearing constraint)

PRD §8.2 step 3 + §17: the seed is "a string constant compiled into the binary — NOT
`go:embed` of a directory; nothing about the user's collection is compiled in."
code_prd_delta.md G11 is THIS gap: "(absent) No compiled-in string-constant example
template (PRD forbids `go:embed`)". go:embed would embed a directory asset and bloat
the coupling; a const is the deliberate choice. Do NOT use `//go:embed`.

## 6. The empty-store check — os.ReadDir, NOT "no SKILL.md"

```go
entries, err := os.ReadDir(store)   // returns []os.DirEntry; hidden files included
if len(entries) == 0 { /* empty → seed */ }
```

- "Empty" = **zero directory entries of ANY kind** (file, dir, dotfile). It is NOT
  "no SKILL.md at any depth" (that is skillsdir.HasSkillMD, a different concern used
  by S1's chooseStore for cwd-auto-detect). Here we ask "is THIS store dir empty?".
- Contract test backing this: "setupStore on a dir containing a file leaves it intact
  (seeded=false)". So a single pre-existing file (even a non-skill file or dotfile)
  ⇒ adopt-in-place, do NOT seed. `len(entries)==0` captures exactly that.
- ReadDir is called AFTER `os.MkdirAll(store, 0o755)`. If MkdirAll just created the
  dir (it didn't exist), ReadDir returns empty ⇒ seed (correct first-run behavior).
  If the dir pre-existed empty, same. If it had content, adopt. One ReadDir covers all.

## 7. The adopt-in-place rule — NEVER clobber or delete (PRD §8.2/§17)

- On `len(entries) > 0`: do NOTHING to existing files. No delete, no overwrite, no
  re-seed. Just fall through to `config.Save`. `seeded` stays `false`.
- This is the §17 guardrail ("never clobber or delete existing files"). The idempotency
  test (§9) proves a re-run leaves example/SKILL.md byte-identical (no re-seed).

## 8. Error-path behavior — wrapped errors; `seeded` is a SUCCESS-PATH signal

- Every fs failure returns `(false, fmt.Errorf("skilldozer init: <step>: %w", err))`.
- `seeded` describes the SUCCESS outcome (tells run()/S3 what message to print: "seeded
  example skill" vs "adopted existing store"). On ANY error, run()/S3 checks `err`
  FIRST and exits 1 — it never reads `seeded`. So returning `false` on the error path
  is the conventional Go "zero value on error" and is unambiguous.
- Concretely: if `config.Save` fails AFTER a successful seed, the function returns
  `(false, err)` (not `(true, err)`). The caller bails on `err`; `seeded` is never
  consulted on the error path. Document this in the doc comment.

## 9. The 4 contract test cases (item_description OUTPUT #4) + 1 error-path bonus

| # | Test | Setup | Assert |
|---|------|-------|--------|
| 1 | empty dir seeds + writes config | `store := t.TempDir()` (empty) | `seeded==true`; `example/SKILL.md` exists & bytes==`exampleSkillTemplate`; `config.Load(cfg).Store == store` |
| 2 | non-empty dir adopts + writes config | pre-write `store/mynotes.md` | `seeded==false`; `mynotes.md` intact; NO `example/` dir; `config.Load(cfg).Store == store` |
| 3 | idempotent re-run | run setupStore twice on same store/cfg | run1 `seeded==true`; run2 `seeded==false`; `example/SKILL.md` byte-identical across runs; config valid |
| 4 | MkdirAll failure → wrapped error, no config written | make `store` path a regular FILE (MkdirAll on a non-dir fails) | `err != nil`; `seeded==false`; `cfg` file does NOT exist |

- Test #1's byte-equality check (`string(got) == exampleSkillTemplate`) is the test
  that catches a malformed template constant (e.g. a botched backtick splice) — it is
  load-bearing for validating §5.
- Test #4 uses a regular file at the store path (deterministic, portable, no root-skip
  needed): `os.MkdirAll("/tmp/x/notadir")` where `notadir` is a file returns a
  `*PathError` ("not a directory"). Assert `err != nil` AND that no config.yaml was
  written (the failure precedes config.Save).

## 10. Duplication note — two copies of the §11 example, intentionally

There are (after this plan) TWO copies of the PRD §11 example skill body:
1. **On-disk `skills/example/SKILL.md`** — owned by P1.M3.T1.S1 (rewrites skpp→skilldozer).
   This is the repo's own shipped example (so `--list`/resolution work for developers).
2. **Compiled-in `exampleSkillTemplate` const** — owned by THIS task (P1.M2.T2.S2).
   This is what `skilldozer init` WRITES into an end user's empty store.

Both MUST equal PRD §11 verbatim. They are intentionally separate (one is a repo asset,
one is the init seed) per the "no go:embed" rule — the compiled-in copy does not read
the on-disk file. The doc comment on `exampleSkillTemplate` must cite PRD §11 and
P1.M3.T1.S1 so a future editor knows the two stay in sync.

## 11. Sibling boundaries (do NOT cross)

- **P1.M2.T2.S1** (input): provides `resolveStore(c.initStore) → (absStore, error)`.
  setupStore CONSUMES its `absStore`. Do not modify resolveStore.
- **P1.M2.T2.S3** (output/consumer): wires `config.Path()` + `resolveStore()` +
  `setupStore()` into run()'s `if c.init { … }`, then prints `--path` + `check`. Do NOT
  add the `if c.init` dispatch here (GOTCHA: setupStore is defined and uncalled until S3,
  exactly as resolveStore was uncalled until S3 — Go allows unused package-level funcs).
- **P1.M3.T1.S1**: owns the on-disk `skills/example/SKILL.md`. This task owns only the
  compiled-in const. Do not edit the on-disk file.
- **internal/config**: CONSUMED (Save/File/Load), NOT modified. No new config fields.

## 12. Why no external/online research is needed

This subtask is Go-stdlib filesystem ops (`os.MkdirAll`, `os.ReadDir`, `os.WriteFile`,
`path/filepath.Join`) + the repo's existing `config` package — all already in the
codebase and verified from source. There is no new library, no API surface beyond
stdlib + config.Save/File. The PRD §11 content (the template) is provided verbatim in
the selected_prd_content. Spawning subagents to "research os.MkdirAll online" would be
wasteful theater; direct source reads (this file) are higher-fidelity.
