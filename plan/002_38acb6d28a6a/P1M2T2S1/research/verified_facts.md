# Verified Facts — P1.M2.T2.S1 (chooseStore core + TTY-gated prompt)

Researched against the LIVE codebase (main.go read in full, internal/config +
internal/skillsdir exports confirmed, main_test.go conventions captured). The
parallel sibling P1.M2.T1.S1 PRP was read as a CONTRACT — its config-field +
parseArgs additions are assumed present when this subtask runs.

---

## §1. What P1.M2.T1.S1 (parallel sibling) delivers — the INPUTS to this subtask

P1.M2.T1.S1 (read its PRP in full) adds, to `main.go` ONLY:

- `config` struct (~line 122): two new fields
  - `init bool` — set by the literal `init` token OR by `--store <dir>` (which implies init)
  - `initStore string` — the non-interactive store dir from `init <dir>` / `--store <dir>` / `--store=<dir>`; empty ⇒ auto-detect
- `parseArgs`: `case "--store":` (both `=`-form @~163 and long-form @~222) and `case "init":` (@~234, after `case "check":`)
- `usageText` (@~50): `skilldozer init [<dir>]` USAGE + EXAMPLE + two OPTIONS lines
- `exclusivityError` (@~635): an `init` family (init+tags, init+mode ⇒ exit 2)

**P1.M2.T1.S1 does NOT touch the import block** (it uses existing `strings`).
**P1.M2.T1.S1 does NOT add `if c.init { … }` to run()** — that is P1.M2.T2.S3.
So after P1.M2.T1.S1: `c.init` / `c.initStore` are populated correctly; run() still
falls through init to the no-mode default (exit 1). This subtask (S1) consumes
`c.initStore` (via the `resolveStore` wrapper that run()/S3 calls) and produces the
chosen store dir.

**Boundary (no collision):** P1.M2.T1.S1 edits the config struct, parseArgs switches,
usageText, exclusivityError — all in the MIDDLE of main.go. This subtask ADDS two
import lines + appends 4 new functions at the END of main.go (after `skillPath`,
the current last function, ~line 694). The regions do not overlap. The import
block is shared but P1.M2.T1.S1 adds 0 import lines; this subtask adds 2
(`bufio`, `internal/config`) — no text-level merge conflict.

---

## §2. The stdout TTY technique to reuse (for os.Stdin) — `isTerminal`

main.go:96-112 (read in full) defines the established repo pattern:

```go
var isTerminal = func(w io.Writer) bool {
    f, ok := w.(*os.File)
    if !ok { return false }
    fi, err := f.Stat()
    if err != nil { return false }
    return fi.Mode()&os.ModeCharDevice != 0
}
```

This checks an arbitrary `io.Writer` typed to `*os.File`. For `init`'s stdin gate
the SAME `ModeCharDevice` bit applies, but to `os.Stdin` DIRECTLY (a different
stream — no `io.Writer` indirection needed; os.Stdin is already `*os.File`):

```go
func stdinIsTerminal() bool {
    fi, err := os.Stdin.Stat()
    if err != nil { return false }
    return fi.Mode()&os.ModeCharDevice != 0
}
```

(external_deps.md §3 + code_prd_delta.md G13 both prescribe exactly this.) Known
harmless caveat: `/dev/null` is also a char device ⇒ `stdinIsTerminal()` reports
true for `init < /dev/null`, BUT a read there yields immediate EOF ⇒ readPrompt
returns the default (never blocks). No `golang.org/x/term` (yaml.v3 stays the
sole non-stdlib dep — code_prd_delta.md §8 / external_deps.md §3).

DESIGN NOTE: `isTerminal` is a package VAR (overridable via `withTerminal`) because
run() does NOT take an isTTY parameter. `stdinIsTerminal` is a plain FUNCTION
(NOT a var) because the contract's test seam is `chooseStore`'s `isTTY` PARAMETER,
not a global override. Do NOT make stdinIsTerminal a var — chooseStore(haveStore,
cwd, isTTY, …) is injected with the bool directly. This is the contract FACTORING:
"keep os.Stdin access in a thin wrapper; pass the prompt fn + isTTY in for tests."

---

## §3. The prompt reader — bufio.NewReader(os.Stdin).ReadString('\n')

external_deps.md §4 (read in full) prescribes `bufio.Reader.ReadString('\n')`:

```go
func readPrompt(r *bufio.Reader, w io.Writer, label, def string) (string, error) {
    if def != "" {
        fmt.Fprintf(w, "%s [%s]: ", label, def)
    } else {
        fmt.Fprintf(w, "%s: ", label)
    }
    line, err := r.ReadString('\n')   // includes trailing '\n'
    if err != nil && err != io.EOF {
        return "", err                // genuine read error (non-EOF)
    }
    if s := strings.TrimSpace(line); s != "" {
        return s, nil
    }
    return def, nil                   // empty Enter OR EOF-no-text ⇒ accept default
}
```

- `ReadString('\n')` returns `(line, error)` where error == `io.EOF` if no newline
  before end of input. A bare `io.EOF` with empty text = "accept default" (NOT a
  hard error) — that is what makes `init < /dev/null` and `echo | skilldozer init`
  behave like "press Enter".
- `err != nil && err != io.EOF` is the ONLY genuine error path (propagate up).
- ONE shared `bufio.NewReader(os.Stdin)` for all prompts (a fresh reader per prompt
  can swallow buffered bytes — external_deps.md §4). The wrapper creates it once.

The prompt fn injected into chooseStore is a closure over the shared reader:
`func(label, def string) (string, error) { return readPrompt(r, os.Stdout, label, def) }`

PRD §8.2 prompt LABEL text (the runtime string): `"Where should skilldozer keep your skills?"`
shown as `Where should skilldozer keep your skills? [<default>]: ` by readPrompt's
`%s [%s]: ` format. Enter accepts default; typing a path overrides.

---

## §4. Consumed APIs (verified present in the live codebase)

- `config.DefaultStore() (string, error)` — internal/config/config.go:150. Pure fn
  of env: `$XDG_DATA_HOME` (if set AND absolute) ⇒ `$XDG_DATA_HOME/skilldozer/skills`;
  else `~/.local/share/skilldozer/skills` via os.UserHomeDir(). Returns its error
  verbatim ($HOME unset). P1.M1.T1.S2 = Complete, so it exists.
- `skillsdir.HasSkillMD(dir string) bool` — internal/skillsdir/skillsdir.go:207.
  Walks `dir` (filepath.WalkDir) for ANY `SKILL.md` at ANY depth; returns true on
  the first hit (stops the walk via a sentinel error), false if none. This is the
  PRD §8.2 "looks like a store — contains at least one SKILL.md at any depth"
  predicate. P1.M1.T2.S1 ("export HasSkillMD") = Complete, so it is EXPORTED.
- `os.Getwd() (string, error)` — stdlib. Used already at skillsdir.go:257 (the
  walk-up rule). On error, return wrapped error (cwd unresolvable is a hard fail
  for init, unlike the walk-up rule which silently misses).
- `filepath.Abs(string) (string, error)` — stdlib, imported (`path/filepath`).
  Absolutizes the chosen store (PRD §8.2 "absolute store path"). On a relative
  typed path like "myskills" ⇒ cwd/myskills (correct: the dir gets mkdir -p'd).

---

## §5. main.go imports — what to ADD

Current (main.go:14-25, read in full):
```go
import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"

    "github.com/dabstractor/skilldozer/internal/check"
    "github.com/dabstractor/skilldozer/internal/discover"
    "github.com/dabstractor/skilldozer/internal/resolve"
    "github.com/dabstractor/skilldozer/internal/search"
    "github.com/dabstractor/skilldozer/internal/skillsdir"
)
```

ADD (2 lines):
- `"bufio"` — stdlib, ALPHABETICAL before `"fmt"` (bufio < fmt).
- `"github.com/dabstractor/skilldozer/internal/config"` — alphabetical between
  `internal/check` and `internal/discover` (check < config < discover).

`io`, `fmt`, `os`, `path/filepath`, `strings` are ALREADY imported (reused by
readPrompt/chooseStore/resolveStore). go.mod/go.sum UNCHANGED (both additions are
already-resolved modules: bufio is stdlib; internal/config is part of THIS module).

---

## §6. The 4-step resolution (contract LOGIC, traced to the 5 OUTPUT test cases)

`chooseStore(haveStore, cwd string, isTTY bool, defaultStore string, prompt func(label, def string)(string,error)) (store string, err error)`:

| Step | Condition | Action | OUTPUT test case it satisfies |
|------|-----------|--------|------------------------------|
| 1 | `haveStore != ""` | return haveStore, nil (prompt NEVER called) | #1: `init --store /tmp/x` ⇒ /tmp/x (no prompt) |
| 2 | (haveStore == "") auto-detect default | `def := defaultStore; if HasSkillMD(cwd) { def = cwd }` | feeds #2/#3/#4 |
| 3 | `isTTY` | `choice, err := prompt(label, def); return choice, err` (empty/EOF⇒def via readPrompt) | #4: prompt "" ⇒ default; #5: prompt "/custom" ⇒ /custom |
| 4 | `!isTTY` (and no haveStore) | return def, nil (prompt NEVER called) | #2: cwd-with-SKILL.md + non-TTY ⇒ cwd; #3: cwd-without + non-TTY ⇒ DefaultStore |

ERROR: a non-nil error is returned ONLY when the prompt fn returns a genuine
(non-EOF) error (readPrompt propagates `err != nil && err != io.EOF`). Empty/EOF
is "accept default" — never an error.

VERBATIM vs ABS: `chooseStore` returns the chosen string VERBATIM (cwd as-passed,
defaultStore as-passed, or the user's typed string). The I/O wrapper `resolveStore`
applies `filepath.Abs` before returning (so the unit tests on chooseStore match the
contract assertions exactly: chooseStore(..., prompt returns "/custom") ⇒ "/custom",
not filepath.Abs("/custom")). This keeps chooseStore a pure decision fn.

---

## §7. Test design — chooseStore is the unit-test surface; resolveStore is integration-only

The contract's 5 OUTPUT test cases are ALL against `chooseStore` (fake prompt fn +
injected isTTY + cwd/defaultStore as params). There is NO unit test for
`resolveStore` (the os.Getwd/os.Stdin wrapper) — it is hard to unit-test cleanly
(os.Stdin is a process-global *os.File; the TTY bit is not a var). resolveStore is
exercised via S3's run dispatch + the P1.M4.T1.S1 §13 acceptance suite
(`SKILLDOZER_CONFIG=… ./skilldozer init --store /tmp/…`).

chooseStore tests use these helpers (NO t.Chdir needed — cwd is a param):
- "cwd-with-SKILL.md": `tmp := t.TempDir(); os.MkdirAll(filepath.Join(tmp,"sub"),0755); os.WriteFile(filepath.Join(tmp,"sub","SKILL.md"),[]byte("# x"),0644)`; pass `cwd=tmp`.
- "cwd-without": pass `cwd=t.TempDir()` (empty).
- fake prompt that FAILS if called (sentinel for the no-prompt guarantee):
  `prompt := func(label, def string)(string,error){ t.Errorf("prompt must not be called"); return "", nil }`
- fake prompt returning "" (accept default) / "/custom" (override) / an error.

TEST MATRIX (7 cases — the 5 OUTPUT cases + prompt-not-called guards + error path):
1. haveStore set ⇒ returns haveStore, prompt NOT called.        (OUTPUT #1)
2. cwd-with-SKILL.md + !isTTY ⇒ returns cwd, prompt NOT called. (OUTPUT #2)
3. cwd-without + !isTTY ⇒ returns defaultStore, prompt NOT called. (OUTPUT #3)
4. isTTY + prompt "" ⇒ returns defaultStore (cwd-without so default=defaultStore). (OUTPUT #4)
5. isTTY + prompt "/custom" ⇒ returns "/custom".                (OUTPUT #5)
6. isTTY + prompt returns an error ⇒ chooseStore returns ("", err). (error propagation)
7. cwd-with-SKILL.md + isTTY + prompt "" ⇒ returns cwd (the cwd-auto-detect DEFAULT is cwd even on TTY; the prompt's empty answer accepts that cwd default). (confirms #2's default is also the TTY default)

Convention: place these near the existing parseArgs/run tests; mirror the
`t.Helper()` + `t.TempDir()` + `filepath.Join` style. Package is `main` (white-box,
like all of main_test.go).

---

## §8. Boundary with sibling subtasks (do NOT cross)

- P1.M2.T2.S2 (create+seed+writeconfig): takes the chosen store STRING (this
  subtask's output) and does mkdir -p + seed example/SKILL.md + config.Save.
  This subtask does NOT mkdir, seed, or write config — it ONLY chooses the dir.
- P1.M2.T2.S3 (run() dispatch): adds `if c.init { … }` to run() that calls
  resolveStore(c.initStore) then S2's create then prints --path + check. This
  subtask does NOT add the `if c.init` branch to run() — resolveStore is defined
  here but left UNCALLED (Go allows unused package-level fns; `go build` is fine).
- P1.M3.T1/S2 (example skill + completions), P1.M4.T2.S1 (README) — untouched.
- NO new files (everything in main.go + main_test.go), matching the repo's
  single-root-file layout (`ls *.go` ⇒ main.go, main_test.go only).
