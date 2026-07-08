# PRP — P1.M2.T2.S1: Choose the store dir — cwd auto-detect + TTY-gated prompt + non-interactive overrides

> **Subtask:** The store-CHOOSING half of `skilldozer init` (PRD §8.2). Delivers a pure, unit-testable core `chooseStore(haveStore, cwd, isTTY, defaultStore, prompt)` implementing the 4-step resolution (non-interactive override → cwd-auto-detect default → TTY-gated prompt → non-TTY silent default), plus the three I/O primitives it needs (`stdinIsTerminal` — the `ModeCharDevice` check applied to `os.Stdin`, a different stream than the existing stdout `isTerminal`; `readPrompt` — a `bufio.NewReader.ReadString('\n')` reader where empty/EOF ⇒ default; and `resolveStore` — the thin I/O wrapper that supplies `os.Getwd`/`config.DefaultStore`/`stdinIsTerminal`/a shared bufio prompt closure and absolutizes the result). **It does NOT mkdir, seed, or write the config** — that is P1.M2.T2.S2; **it does NOT add `if c.init { … }` to `run()`** — that is P1.M2.T2.S3. After this subtask the store-choice logic exists and is fully unit-tested, but `init` still no-ops to exit 1 until S3 wires `resolveStore` into `run()`.
>
> **Scope:** Two existing files only — `main.go` (add 2 imports + append 4 functions at the end) and `main_test.go` (add 7 `chooseStore` unit tests). No new files. No `internal/*` change. Zero new dependencies (`bufio` is stdlib; `internal/config` is this module). `go.mod`/`go.sum` byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** main.go + main_test.go + internal/config + internal/skillsdir read; `config.DefaultStore()` (config.go:150) and `skillsdir.HasSkillMD()` (skillsdir.go:207) confirmed exported and present (P1.M1.T1.S2 / P1.M1.T2.S1 = Complete). The parallel sibling P1.M2.T1.S1 PRP was read as a CONTRACT — its `config.init`/`initStore` fields + `case "init"`/`--store` parsing are assumed present (it does NOT touch imports and edits main.go's MIDDLE; this subtask appends at the END — no overlap). `grep` confirms main.go has zero current `chooseStore`/`resolveStore`/`readPrompt`/`stdinIsTerminal` references, so this is purely additive.

---

## Goal

**Feature Goal**: Implement and unit-test the `skilldozer init` store-selection decision logic in isolation from the terminal, so that PRD §8.2's four behaviors — (a) `init <dir>`/`--store <dir>` uses the given dir with NO prompt; (b) cwd that already contains a `SKILL.md` becomes the default store ("detected skills in <cwd>"); (c) otherwise the `$XDG_DATA_HOME/skilldozer/skills` default applies; (d) on a TTY a "Where should skilldozer keep your skills? [<default>]" prompt lets the user override (Enter/EOF ⇒ default), while off-TTY (pipes/CI/scripts) the auto-detected default is used SILENTLY with no prompt and no hang — all hold and are provably correct via 7 unit tests that never touch a real terminal.

**Deliverable**: Additive edits to two existing files:
1. `main.go` — add `"bufio"` + `"github.com/dabstractor/skilldozer/internal/config"` to the import block; append four package-level functions after `skillPath` (the current last function): `stdinIsTerminal()`, `readPrompt(r, w, label, def)`, `chooseStore(haveStore, cwd, isTTY, defaultStore, prompt)`, and `resolveStore(haveStore)`.
2. `main_test.go` — add 7 `TestChooseStore*` unit tests covering the 5 contract OUTPUT cases + the no-prompt guard + the prompt-error path.

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `chooseStore("/tmp/x", anyCwd, false, anyDef, failIfCalled)` ⇒ `("/tmp/x", nil)` with the prompt fn never invoked; `chooseStore("", skillCwd, false, "/def", failIfCalled)` ⇒ `(skillCwd, nil)`; `chooseStore("", emptyCwd, false, "/def", failIfCalled)` ⇒ `("/def", nil)`; `chooseStore("", emptyCwd, true, "/def", fakeReturns(""))` ⇒ `("/def", nil)`; `chooseStore("", emptyCwd, true, "/def", fakeReturns("/custom"))` ⇒ `("/custom", nil)`.

---

## User Persona (if applicable)

**Target User**: A first-run `skilldozer` user at a TTY, AND the scripts/CI that drive `skilldozer init --store <dir>` or `skilldozer init < /dev/null`. The decision logic serves both without the second ever blocking on stdin.

**Use Case**: `skilldozer init` run inside an existing skills repo (adopts cwd), inside an empty dir (offers the XDG default), or non-interactively with `--store`/piped stdin (no prompt, no hang).

**User Journey**: User runs `skilldozer init` → if they passed `--store <dir>` or `init <dir>`, that dir wins silently; else skilldozer inspects cwd for any `SKILL.md` to pick a default; if stdin is a TTY it asks "Where should skilldozer keep your skills? [<default>]" (Enter keeps the default); if stdin is a pipe/file it uses the default silently. (The actual mkdir/seed/config-write/print happens in S2/S3 — this subtask only DECIDES the dir.)

**Pain Points Addressed**: a CLI that would otherwise hang inside `$(...)` or CI if it prompted unconditionally (PRD §8.2 prompt-safety, decision #13); a go-install user with no obvious store location getting a sane XDG default; a clone user running `init` inside their skills repo getting zero-typing adoption.

---

## Why

- **Implements the decision core of PRD §8.2 step 1** (auto-detect cwd → default → prompt → Enter/override) and the non-interactive forms (`init <dir>` / `--store <dir>`), which are the input to S2's create+seed+writeconfig and S3's run() dispatch.
- **Closes gap G13** (`code_prd_delta.md` §8/§10): `isTerminal` (main.go:96-112) checks *stdout*; init needs `isatty(stdin)` — a different stream, same `ModeCharDevice` technique. Delivered as `stdinIsTerminal()`.
- **Honors the load-bearing prompt-safety guarantee** (PRD §8.2 / decision #13): the prompt is gated on `isatty(stdin)`, so piped/redirected/CI invocations never block — the `pi --skill "$(skilldozer init …)"`-style command-substitution hazard is structurally impossible because the prompt fn is only CALLED when `isTTY` is true.
- **Keeps yaml.v3 the sole non-stdlib dependency** (PRD §4 / external_deps.md §3): reuses the repo's existing `ModeCharDevice` heuristic instead of adding `golang.org/x/term`.
- **Unblocks P1.M2.T2.S2** (create+seed+writeconfig) and **P1.M2.T2.S3** (run() dispatch), both of which consume this subtask's `resolveStore(c.initStore) → (absStore, error)` output.

---

## What

### Success Criteria

- [ ] `main.go` import block adds `"bufio"` (stdlib, before `"fmt"`) and `"github.com/dabstractor/skilldozer/internal/config"` (between `internal/check` and `internal/discover`).
- [ ] `stdinIsTerminal() bool` checks `os.Stdin.Stat()` `& os.ModeCharDevice` (mirrors the stdout `isTerminal` technique on a different stream); returns false on any stat error.
- [ ] `readPrompt(r *bufio.Reader, w io.Writer, label, def string) (string, error)` prints `label + " [" + def + "]: "` (or `label + ": "` if def=="") to `w`, reads one line via `r.ReadString('\n')`, returns the trimmed non-empty answer; returns `def` on empty line OR `io.EOF`; returns the error only on a genuine non-EOF read failure.
- [ ] `chooseStore(haveStore, cwd string, isTTY bool, defaultStore string, prompt func(label, def string) (string, error)) (store string, err error)` implements the 4-step resolution: (1) `haveStore != ""` ⇒ return it, prompt NEVER called; (2) `def := defaultStore; if skillsdir.HasSkillMD(cwd) { def = cwd }`; (3) `!isTTY` ⇒ return `def`, prompt NEVER called; (4) `isTTY` ⇒ `return prompt("Where should skilldozer keep your skills?", def)` (readPrompt makes empty/EOF ⇒ def).
- [ ] `resolveStore(haveStore string) (string, error)` is the thin I/O wrapper: gets cwd via `os.Getwd()` (wrapped error on fail), default via `config.DefaultStore()` (wrapped error on fail), creates ONE shared `bufio.NewReader(os.Stdin)`, builds the prompt closure `(label, def) => readPrompt(r, os.Stdout, label, def)`, calls `chooseStore(haveStore, cwd, stdinIsTerminal(), def, prompt)`, and returns `filepath.Abs(store)` (wrapped error on fail).
- [ ] `go test ./...` green, including the 7 new `TestChooseStore*` tests; existing tests unaffected (purely additive — no existing symbol renamed/moved).
- [ ] `go.mod`/`go.sum` unchanged; no new files; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every function is pinned to a verified, live symbol: the stdout `isTerminal` technique (main.go:96-112, transcribed in `research/verified_facts.md` §2) is the template for `stdinIsTerminal`; external_deps.md §4's `readPrompt` (transcribed in §3) is the template for `readPrompt` verbatim; `config.DefaultStore()` and `skillsdir.HasSkillMD()` are confirmed exported (§4). The 4-step resolution is traced case-by-case against all 5 contract OUTPUT test cases (§6), and the test design (§7) explains why `chooseStore` is the only unit-test surface (cwd/isTTY/prompt are injected params; no `t.Chdir` or terminal mocking needed). The boundary with P1.M2.T1.S1 (parallel) is fixed: it edits main.go's middle, this subtask appends at the end + adds 2 import lines P1.M2.T1.S1 does not touch. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (line anchors + the 4-step trace + the test matrix)
- file: plan/002_38acb6d28a6a/P1M2T2S1/research/verified_facts.md
  why: "§1 = what the parallel sibling P1.M2.T1.S1 delivers (the c.initStore INPUT +
        the no-collision boundary). §2 = the isTerminal technique to clone for stdin
        (and WHY stdinIsTerminal is a fn, not a var — the test seam is chooseStore's
        isTTY PARAMETER). §3 = readPrompt verbatim (external_deps.md §4) + the empty/
        EOF⇒default rule + the one-shared-reader rule. §4 = the consumed APIs
        (DefaultStore, HasSkillMD, os.Getwd, filepath.Abs) confirmed present. §5 = the
        exact 2 import lines to add. §6 = the 4-step resolution traced to the 5 OUTPUT
        cases + the VERBATIM-vs-ABS decision. §7 = the 7-test matrix. §8 = sibling
        boundaries (do NOT mkdir/seed/writeconfig; do NOT add if c.init to run())."
  critical: "§6 step (1) — haveStore != '' must short-circuit BEFORE computing HasSkillMD
             (else an empty cwd test still calls HasSkillMD, harmless but the no-prompt
             guard is the load-bearing assertion). §6 VERBATIM-vs-ABS — chooseStore
             returns the choice verbatim; ONLY resolveStore applies filepath.Abs, so the
             unit assertions match the contract literally."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as P1.M2.T1.S1 lands)
- file: main.go
  why: "THE edit target. Import block @14-25 (add bufio + internal/config). The stdout
        isTerminal var @96-112 (the TECHNIQUE to clone for stdin — read its doc comment;
        note it is a var because run() has no isTTY param, whereas stdinIsTerminal is a
        plain fn because chooseStore takes isTTY as a param). The LAST function is
        skillPath (~@685-694) — APPEND the 4 new functions AFTER it (keeps the change at
        the file tail, non-overlapping with P1.M2.T1.S1's middle-of-file edits)."
  pattern: "TTY check = os.File.Stat() & os.ModeCharDevice. I/O wrapper = assemble real
            deps (os.Getwd, config.DefaultStore, os.Stdin, os.Stdout) then delegate to a
            pure core that takes them as params (the established testability pattern:
            run()/parseArgs/exclusivityError/skillPath are all pure-of-terminal)."

# MUST READ — the test file under edit (mirror these helpers/shapes exactly)
- file: main_test.go
  why: "THE other edit target + the test-template source. withTerminal(t,bool) @18 = the
        var-override pattern for isTerminal (NOT used here — chooseStore takes isTTY as a
        param instead). unsetSkillsEnv(t) @25 + writeSkillTree(t,layout) @40 + t.TempDir()
        / t.Setenv / t.Chdir conventions = the house style. The chooseStore tests need NO
        t.Chdir (cwd is a param): build a temp cwd with a nested SKILL.md via
        os.MkdirAll+os.WriteFile, pass it as the cwd arg."
  gotcha: "Do NOT write a run(['init']) or resolveStore unit test — run() dispatch is
           P1.M2.T2.S3 (today exits 1) and resolveStore touches os.Stdin/os.Getwd which
           are not cleanly unit-testable (integration coverage is S3 + P1.M4.T1.S1 §13).
           The unit surface is chooseStore ONLY."

# READ-ONLY — external facts (TTY technique + bufio reader) — transcribed into verified_facts §2/§3
- file: plan/002_38acb6d28a6a/architecture/external_deps.md
  why: "§3 prescribes stdinIsTerminal (ModeCharDevice on os.Stdin; /dev/null-is-char-device
        caveat is harmless — EOF⇒default). §4 prescribes readPrompt (ReadString('\\n');
        empty/EOF⇒default; err!=nil&&err!=io.EOF is the only hard error; ONE shared
        bufio.NewReader). §3 explicitly says do NOT add golang.org/x/term (yaml.v3 stays
        the sole non-stdlib dep)."
  section: "§3 (TTY detection), §4 (prompt reader), §5 (confirms go install needs the config
            default this subtask offers)."

# READ-ONLY — the gap analysis (G13 is THIS subtask)
- file: plan/002_38acb6d28a6a/architecture/code_prd_delta.md
  why: "§8 confirms isTerminal checks stdout and init needs a SEPARATE stdin check (G13);
        §8 confirms hasSkillMD is now exported (P1.M1.T2.S1 done — 'Note' resolved); §10
        gap index row G13 = this subtask. §8 'go.mod: no new dep needed' confirms bufio/
        config add nothing to go.mod."
  section: "§8 (cross-cutting: isTerminal/hasSkillMD/go.mod), §10 (G13 row)."

# READ-ONLY — the parallel sibling PRP (defines the INPUT contract: c.initStore)
- file: plan/002_38acb6d28a6a/P1M2T1S1/PRP.md
  why: "Defines c.init (bool) + c.initStore (string) populated by `init <dir>` / `--store
        <dir>` / `--store=<dir>`. Confirms it does NOT touch imports and edits main.go's
        middle (config struct, parseArgs, usageText, exclusivityError) — so this subtask's
        2 import lines + 4 appended functions compose without collision. c.initStore is
        the haveStore argument resolveStore receives."

# READ-ONLY — PRD (source of truth for the init store-choice contract)
- file: PRD.md
  why: "§8.2 (h3.9) = the authoritative 4-step flow + prompt text + non-interactive forms +
        prompt-safety guarantee. §8.3 (h3.10) rule 2 = the config store init writes (S2,
        not here — but this subtask's chosen dir is what gets written). §6.4 (h3.4) = the
        never-prompt-on-bare-tag contract this subtask's TTY gate upholds. decisions #13
        (TTY-gated auto-prompt) + #16 (cwd auto-detect via HasSkillMD)."
  section: "h3.9 (§8.2), h3.10 (§8.3), h3.4 (§6.4), decisions 13/16."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer && ls *.go
main.go          # EDIT: +2 imports (bufio, internal/config); APPEND stdinIsTerminal/readPrompt/chooseStore/resolveStore after skillPath
main_test.go     # EDIT: +7 TestChooseStore* unit tests
internal/        # untouched (config.DefaultStore, skillsdir.HasSkillMD CONSUMED, not modified)
# go.mod / go.sum untouched (bufio=stdlib; internal/config=this module)
$ grep -n 'chooseStore\|resolveStore\|readPrompt\|stdinIsTerminal' main.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
main.go          # ADD: bufio+config imports; 4 appended fns (stdinIsTerminal, readPrompt, chooseStore, resolveStore)
main_test.go     # ADD: TestChooseStore* (7): override/no-prompt, cwd-detect×TTY, default×TTY, custom×TTY, error path
```

**No new files.** All edits are additive to existing files.

| File | Responsibility |
|---|---|
| `main.go` | The store-CHOICE decision logic + its terminal I/O primitives. Pure core (`chooseStore`) is terminal-free and unit-testable; `resolveStore` is the thin I/O assembler run()/S3 calls. |
| `main_test.go` | Lock the 4-step resolution + the no-prompt guarantee + the error path via 7 `chooseStore` unit tests (injected isTTY + fake prompt fn). |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL — load-bearing prompt safety) — the prompt fn is ONLY called
// when isTTY==true. The two non-interactive branches (haveStore != "" and !isTTY)
// must return WITHOUT ever invoking prompt. This is what makes `echo y | skilldozer
// init` and `skilldozer init < /dev/null` and CI not hang. ENFORCE IT IN TESTS with
// a fake prompt that calls t.Fatalf("prompt must not be called") if invoked.
// Concretely chooseStore must be:
//   if haveStore != "" { return haveStore, nil }      // (1) NO prompt
//   def := defaultStore
//   if skillsdir.HasSkillMD(cwd) { def = cwd }        // (2) auto-detect default
//   if !isTTY { return def, nil }                      // (4) NO prompt
//   return prompt("Where should skilldozer keep your skills?", def)  // (3) prompt
// NOTE the ORDER: the !isTTY return (step 4) comes AFTER computing def (step 2) but
// BEFORE calling prompt (step 3). Do not re-order — computing HasSkillMD is cheap and
// has no side effects, but the prompt call MUST be the last thing.

// GOTCHA #2 — haveStore short-circuits BEFORE HasSkillMD(cwd). If you compute
// HasSkillMD first (step 2 before step 1), an empty/non-skill cwd still triggers a
// needless tree walk; harmless but the contract's "no prompt attempted" for the
// --store case is cleaner if haveStore is checked first. Either order passes the
// OUTPUT #1 test (prompt never called), but haveStore-first is the contract LOGIC
// order (1)→(2)→(3)/(4) and avoids the walk. Follow the LOGIC order exactly.

// GOTCHA #3 — chooseStore returns the choice VERBATIM; ONLY resolveStore applies
// filepath.Abs. If chooseStore itself calls filepath.Abs, the unit test
// chooseStore(..., prompt returns "/custom") would need to know the test's cwd to
// predict filepath.Abs("/custom") — breaking the clean "in ⇒ out" assertion. Keep
// chooseStore a pure decision fn; put Abs in the wrapper (resolveStore is the
// "absolute chosen store dir" deliverable, chooseStore is the decision core).

// GOTCHA #4 — readPrompt's empty-line and EOF BOTH mean "accept default" (return def,
// nil). `bufio.Reader.ReadString('\n')` returns `(line, io.EOF)` when there is no
// newline before end of input. Treat `err == io.EOF` with empty/whitespace text as
// default, NOT an error. The ONLY error path is `err != nil && err != io.EOF`
// (a genuine read failure). This is what makes `init < /dev/null` behave like Enter.

// GOTCHA #5 — ONE shared bufio.NewReader(os.Stdin) in resolveStore. Do NOT create a
// fresh reader per prompt call (external_deps.md §4: a fresh reader can swallow
// buffered bytes from the previous read). Today there is one prompt, but the shared
// reader is created once in resolveStore and captured by the prompt closure so a
// future second prompt (if added) reuses it.

// GOTCHA #6 — stdinIsTerminal is a PLAIN FUNCTION, not a package var. The existing
// isTerminal (stdout) IS a var because run() has no isTTY parameter and tests override
// it via withTerminal(t,bool). For init the contract FACTORING chose PARAMETER
// injection: chooseStore takes isTTY as a bool arg, so tests pass it directly and do
// NOT need to mutate package state. Do NOT make stdinIsTerminal a var; do NOT add a
// withStdinTerminal helper. (If a future test wants to exercise resolveStore end-to-
// end, that is S3/acceptance territory with a real pipe, not a unit test here.)

// GOTCHA #7 — /dev/null is a char device, so stdinIsTerminal() reports TRUE for
// `init < /dev/null`. This is HARMLESS: the prompt is shown, ReadString reads EOF
// immediately, readPrompt returns the default. Do NOT special-case /dev/null. Adding
// golang.org/x/term for a precise isatty would VIOLATE the yaml.v3-only constraint
// (external_deps.md §3 explicitly forbids it).

// GOTCHA #8 — resolveStore is defined here but left UNCALLED (S3 wires it into run()'s
// `if c.init`). Go ALLOWS unused package-level functions (only unused imports and
// unused LOCAL variables are compile errors), so `go build`/`go vet` are green with
// resolveStore uncalled. Do NOT add a throwaway call to "use" it; do NOT add the
// `if c.init` dispatch (that is S3's scope and would conflict with it).

// GOTCHA #9 — resolveStore needs the `internal/config` import (for DefaultStore) and
// `bufio` (for NewReader). main.go does NOT currently import either. Add BOTH.
// `io`, `fmt`, `os`, `path/filepath`, `strings` are already imported (reused). The 2
// new imports are already-resolved modules (bufio=stdlib; config=this module), so
// `go mod tidy` is a NO-OP and go.mod/go.sum stay byte-for-byte unchanged (verify:
// `git diff --quiet go.mod go.sum`).

// GOTCHA #10 — no merge collision with the parallel sibling P1.M2.T1.S1. That sibling
// edits main.go's MIDDLE (config struct @~122, parseArgs switches, usageText @~50,
// exclusivityError @~635) and adds ZERO import lines. This subtask adds 2 import
// lines and APPENDS 4 functions after skillPath (the file tail). The import-block edit
// (inserting 2 lines) does not textually collide with P1.M2.T1.S1's 0 import-line
// change. So the two changesets compose cleanly regardless of merge order.
```

---

## Implementation Blueprint

### Data models and structure

**No new types.** The only new "model" is the prompt function type, expressed inline as a parameter signature (no named type needed, matching the contract's literal signature):

```go
// chooseStore's prompt parameter (inline signature — no named type):
//   prompt func(label, def string) (string, error)
// label  = the question text (e.g. "Where should skilldozer keep your skills?")
// def    = the default store shown in [brackets] and returned on empty/EOF
// returns = the user's typed path (already trimmed), OR def, OR a genuine read error
```

No struct changes (the `config.init`/`initStore` fields are added by P1.M2.T1.S1, consumed here as `c.initStore` via `resolveStore`).

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add the 2 imports
  - FILE: main.go (import block, ~lines 14-25)
  - ADD `"bufio"` as the FIRST stdlib import (alphabetical: bufio < fmt).
  - ADD `"github.com/dabstractor/skilldozer/internal/config"` BETWEEN the existing
    `"internal/check"` and `"internal/discover"` lines (alphabetical: check < config < discover).
  - DO NOT touch any other import line. DO NOT run `go mod tidy` (both are already
    resolved; GOTCHA #9).
  - VERIFY after: `go build ./...` compiles (the imports are referenced by Task 4's fns).

Task 2: APPEND main.go — stdinIsTerminal (the stdin TTY check)
  - FILE: main.go (APPEND after skillPath, the current last function, ~line 694)
  - ADD (mirror main.go:96-112's ModeCharDevice technique on os.Stdin directly;
    GOTCHA #6 — a plain fn, NOT a var):
      // stdinIsTerminal reports whether os.Stdin is an interactive terminal. It is
      // the stdin counterpart of the stdout isTerminal check (PRD §6.2 color gating,
      // main.go:96-112): the SAME ModeCharDevice technique applied to a DIFFERENT
      // stream (os.Stdin, not an io.Writer). init uses it to gate the interactive
      // prompt so piped/redirected invocations never block (PRD §8.2 prompt safety).
      //
      // It is a plain function (not a package var) because init's test seam is
      // chooseStore's isTTY PARAMETER, not a global override — see chooseStore.
      // Caveat (harmless): /dev/null is also a char device, so this reports true for
      // `init < /dev/null`; the immediate EOF there makes readPrompt return the
      // default, so it never hangs. No golang.org/x/term (yaml.v3 stays the sole
      // non-stdlib dep — external_deps.md §3).
      func stdinIsTerminal() bool {
          fi, err := os.Stdin.Stat()
          if err != nil {
              return false
          }
          return fi.Mode()&os.ModeCharDevice != 0
      }

Task 3: APPEND main.go — readPrompt (the bufio line reader; external_deps.md §4)
  - FILE: main.go (APPEND right after stdinIsTerminal)
  - ADD (verbatim from external_deps.md §4; GOTCHA #4 — empty/EOF ⇒ default):
      // readPrompt prints the prompt (label, with [def] in brackets) to w, reads one
      // line from r, and returns the trimmed answer — or def when the user just presses
      // Enter (empty line) or sends EOF on an otherwise-empty line. A genuine read
      // error (non-EOF) is returned. Used by init's interactive prompt (PRD §8.2).
      // (external_deps.md §4 prescribes bufio.Reader.ReadString('\n') over Scanner.)
      func readPrompt(r *bufio.Reader, w io.Writer, label, def string) (string, error) {
          if def != "" {
              fmt.Fprintf(w, "%s [%s]: ", label, def)
          } else {
              fmt.Fprintf(w, "%s: ", label)
          }
          line, err := r.ReadString('\n') // includes the trailing '\n'
          if err != nil && err != io.EOF {
              return "", err
          }
          if s := strings.TrimSpace(line); s != "" {
              return s, nil
          }
          return def, nil // empty Enter OR EOF-with-no-text ⇒ accept default
      }

Task 4: APPEND main.go — chooseStore (the PURE testable core; the 4-step resolution)
  - FILE: main.go (APPEND right after readPrompt)
  - ADD (GOTCHA #1/#2/#3 — the order and the verbatim return are load-bearing):
      // chooseStore resolves the store directory for `skilldozer init` (PRD §8.2) via a
      // 4-step decision that is fully independent of os.Stdin/os.Stdout/os.Getwd: the
      // caller injects cwd, isTTY, the default store, and a prompt function, so the
      // logic is unit-testable without a real terminal (the contract FACTORING).
      //
      // Resolution order (first applicable wins):
      //  1. haveStore != "" — the non-interactive override from `init <dir>` or
      //     `--store <dir>`. Returned VERBATIM; the prompt is NEVER called (scripts/CI).
      //  2. auto-detect the default: if cwd already looks like a store (it contains at
      //     least one SKILL.md at any depth — skillsdir.HasSkillMD, PRD §8.2 "detected
      //     skills in <cwd>"), default = cwd; else default = defaultStore (the
      //     $XDG_DATA_HOME/skilldozer/skills value from config.DefaultStore).
      //  3. isTTY — prompt "Where should skilldozer keep your skills? [<default>]".
      //     readPrompt makes empty line / EOF ⇒ default; a typed path ⇒ override.
      //  4. !isTTY and no explicit haveStore — return the auto-detected default with NO
      //     prompt (scripts / CI / pipes). The prompt is NEVER called.
      //
      // The chosen string is returned VERBATIM (it may be relative if the user typed a
      // relative path); resolveStore absolutizes it via filepath.Abs. A non-nil error is
      // returned ONLY on a genuine prompt read failure (a non-EOF error from the prompt
      // fn); empty/EOF is "accept default", never an error.
      func chooseStore(haveStore, cwd string, isTTY bool, defaultStore string, prompt func(label, def string) (string, error)) (string, error) {
          // (1) Non-interactive override: `init <dir>` / `--store <dir>`. No prompt.
          if haveStore != "" {
              return haveStore, nil
          }
          // (2) Auto-detect the default from cwd (PRD §8.2 "detected skills in <cwd>").
          def := defaultStore
          if skillsdir.HasSkillMD(cwd) {
              def = cwd
          }
          // (4) Off-TTY (pipe/file/CI): use the default, NO prompt (never blocks).
          if !isTTY {
              return def, nil
          }
          // (3) Interactive: prompt. Empty/EOF ⇒ def (readPrompt); typed ⇒ override.
          return prompt("Where should skilldozer keep your skills?", def)
      }

Task 5: APPEND main.go — resolveStore (the thin I/O wrapper run()/S3 calls)
  - FILE: main.go (APPEND right after chooseStore — this is the file's new tail)
  - ADD (GOTCHA #5/#8/#9 — one shared reader; left uncalled until S3; needs config+bufio):
      // resolveStore is the I/O-bearing wrapper around chooseStore that run()'s init
      // dispatch (P1.M2.T2.S3) calls. It supplies the real dependencies — os.Getwd(),
      // config.DefaultStore(), the os.Stdin TTY check (stdinIsTerminal), and a bufio
      // prompt reader over os.Stdin/os.Stdout (readPrompt) — and returns chooseStore's
      // choice ABSOLUTIZED via filepath.Abs (PRD §8.2 "absolute store path"). The ONE
      // shared bufio.NewReader is created here and captured by the prompt closure so a
      // future second prompt would reuse it (external_deps.md §4: a fresh reader per
      // prompt can swallow buffered bytes).
      //
      // The os.Stdin / os.Stdout / os.Getwd access is confined to THIS function so the
      // pure decision logic in chooseStore stays terminal-free and unit-testable. A
      // genuine cwd/default/absolutize/prompt error is returned wrapped; an empty or
      // EOF prompt answer is NOT an error (readPrompt ⇒ default).
      func resolveStore(haveStore string) (string, error) {
          cwd, err := os.Getwd()
          if err != nil {
              return "", fmt.Errorf("skilldozer init: resolve cwd: %w", err)
          }
          def, err := config.DefaultStore()
          if err != nil {
              return "", fmt.Errorf("skilldozer init: resolve default store: %w", err)
          }
          r := bufio.NewReader(os.Stdin)
          prompt := func(label, def string) (string, error) {
              return readPrompt(r, os.Stdout, label, def)
          }
          store, err := chooseStore(haveStore, cwd, stdinIsTerminal(), def, prompt)
          if err != nil {
              return "", err
          }
          abs, err := filepath.Abs(store)
          if err != nil {
              return "", fmt.Errorf("skilldozer init: absolutize store: %w", err)
          }
          return abs, nil
      }

Task 6: EDIT main_test.go — add the 7 chooseStore unit tests
  - FILE: main_test.go (APPEND a new block; mirror the house style: t.Helper where apt,
    t.TempDir(), filepath.Join, os.MkdirAll/WriteFile for the cwd-with-SKILL.md fixture)
  - GROUP near the end of the file (or near the other parseArgs/run init tests P1.M2.T1.S1
    adds — different function names, no collision). Package is `main` (white-box).
  - (6a) A helper to build a temp dir containing a NESTED SKILL.md (cwd-looks-like-a-store):
      func mkdirWithSkillMD(t *testing.T) string {
          t.Helper()
          dir := t.TempDir()
          sub := filepath.Join(dir, "writing", "reddit")
          if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("MkdirAll: %v", err) }
          if err := os.WriteFile(filepath.Join(sub, "SKILL.md"), []byte("# skill\n"), 0o644); err != nil {
              t.Fatalf("WriteFile: %v", err)
          }
          return dir
      }
  - (6b) A fake prompt that FAILS if called (the no-prompt guard; GOTCHA #1):
      failIfCalled := func(t *testing.T) func(string, string) (string, error) {
          t.Helper()
          return func(label, def string) (string, error) {
              t.Errorf("chooseStore: prompt must not be called (label=%q)", label)
              return "", nil
          }
      }
  - (6c) The 7 tests (assertions match the contract OUTPUT list + the guard + the error):
      func TestChooseStoreExplicitOverrideNoPrompt(t *testing.T) {
          // OUTPUT #1: `init --store /tmp/x` ⇒ /tmp/x; prompt NEVER called.
          got, err := chooseStore("/tmp/x", "/any/cwd", true, "/def", failIfCalled(t))
          if err != nil || got != "/tmp/x" {
              t.Errorf("chooseStore(/tmp/x,...): got (%q,%v); want (/tmp/x,nil)", got, err)
          }
      }
      func TestChooseStoreCwdDetectNonTTY(t *testing.T) {
          // OUTPUT #2: cwd-with-SKILL.md + non-TTY ⇒ cwd; prompt NEVER called.
          cwd := mkdirWithSkillMD(t)
          got, err := chooseStore("", cwd, false, "/def", failIfCalled(t))
          if err != nil || got != cwd {
              t.Errorf("chooseStore(cwd-with-skill,non-TTY): got (%q,%v); want (%q,nil)", got, err, cwd)
          }
      }
      func TestChooseStoreNoSkillNonTTYUsesDefault(t *testing.T) {
          // OUTPUT #3: cwd-without + non-TTY ⇒ defaultStore; prompt NEVER called.
          got, err := chooseStore("", t.TempDir(), false, "/def", failIfCalled(t))
          if err != nil || got != "/def" {
              t.Errorf("chooseStore(empty-cwd,non-TTY): got (%q,%v); want (/def,nil)", got, err)
          }
      }
      func TestChooseStoreTTYEmptyPromptAcceptsDefault(t *testing.T) {
          // OUTPUT #4: isTTY + prompt "" ⇒ default (cwd-without so default=defaultStore).
          prompt := func(label, def string) (string, error) { return "", nil }
          got, err := chooseStore("", t.TempDir(), true, "/def", prompt)
          if err != nil || got != "/def" {
              t.Errorf("chooseStore(TTY,empty-prompt): got (%q,%v); want (/def,nil)", got, err)
          }
      }
      func TestChooseStoreTTYTypedPathOverrides(t *testing.T) {
          // OUTPUT #5: isTTY + prompt "/custom" ⇒ /custom (VERBATIM — GOTCHA #3).
          prompt := func(label, def string) (string, error) { return "/custom", nil }
          got, err := chooseStore("", t.TempDir(), true, "/def", prompt)
          if err != nil || got != "/custom" {
              t.Errorf("chooseStore(TTY,typed-/custom): got (%q,%v); want (/custom,nil)", got, err)
          }
      }
      func TestChooseStoreCwdDetectIsAlsoTheTTYDefault(t *testing.T) {
          // The cwd-auto-detect DEFAULT is cwd even on a TTY; an empty prompt answer
          // accepts that cwd default (not defaultStore). Guards against a bug where
          // HasSkillMD is only consulted on the !isTTY branch.
          cwd := mkdirWithSkillMD(t)
          prompt := func(label, def string) (string, error) {
              if def != cwd { t.Errorf("prompt default=%q; want cwd %q (auto-detect)", def, cwd) }
              return "", nil // Enter ⇒ accept the cwd default
          }
          got, err := chooseStore("", cwd, true, "/def", prompt)
          if err != nil || got != cwd {
              t.Errorf("chooseStore(cwd-with-skill,TTY,empty): got (%q,%v); want (%q,nil)", got, err, cwd)
          }
      }
      func TestChooseStorePropagatesPromptError(t *testing.T) {
          // A genuine (non-EOF) prompt read error is returned, not swallowed.
          wantErr := errors.New("simulated read failure")
          prompt := func(label, def string) (string, error) { return "", wantErr }
          got, err := chooseStore("", t.TempDir(), true, "/def", prompt)
          if err == nil || !errors.Is(err, wantErr) {
              t.Errorf("chooseStore(prompt-error): got (%q,%v); want error wrapping %v", got, err, wantErr)
          }
      }
  - GOTCHA: TestChooseStorePropagatesPromptError needs `errors` — ADD it to main_test.go's
    imports if not already present (it currently imports bytes, io, os, path/filepath,
    strings, testing — check and add "errors"). This is the ONLY test-side import change.

Task 7: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go test ./...                     # ALL pass incl. the 7 new tests; existing unaffected
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #9
  - manual: go test -run TestChooseStore -v ./...           # the 7 tests named + green
```

### Implementation Patterns & Key Details

```go
// stdinIsTerminal — the ModeCharDevice check on os.Stdin (sibling of the stdout isTerminal).
func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// readPrompt — one line; empty/EOF ⇒ default; non-EOF error ⇒ propagate (external_deps.md §4).
func readPrompt(r *bufio.Reader, w io.Writer, label, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(w, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(w, "%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	if s := strings.TrimSpace(line); s != "" {
		return s, nil
	}
	return def, nil
}

// chooseStore — the pure 4-step core. ORDER MATTERS (GOTCHA #1/#2): override → detect
// default → off-TTY silent → on-TTY prompt. Verbatim return (GOTCHA #3).
func chooseStore(haveStore, cwd string, isTTY bool, defaultStore string, prompt func(label, def string) (string, error)) (string, error) {
	if haveStore != "" {
		return haveStore, nil
	}
	def := defaultStore
	if skillsdir.HasSkillMD(cwd) {
		def = cwd
	}
	if !isTTY {
		return def, nil
	}
	return prompt("Where should skilldozer keep your skills?", def)
}

// resolveStore — the thin I/O assembler. One shared bufio reader (GOTCHA #5); left
// uncalled until S3 (GOTCHA #8); absolutizes the verbatim choice (GOTCHA #3).
func resolveStore(haveStore string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("skilldozer init: resolve cwd: %w", err)
	}
	def, err := config.DefaultStore()
	if err != nil {
		return "", fmt.Errorf("skilldozer init: resolve default store: %w", err)
	}
	r := bufio.NewReader(os.Stdin)
	prompt := func(label, def string) (string, error) {
		return readPrompt(r, os.Stdout, label, def)
	}
	store, err := chooseStore(haveStore, cwd, stdinIsTerminal(), def, prompt)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(store)
	if err != nil {
		return "", fmt.Errorf("skilldozer init: absolutize store: %w", err)
	}
	return abs, nil
}
```

Notes easy to get wrong:
- The `if haveStore != ""` and `if !isTTY` branches BOTH must return WITHOUT calling `prompt`. The `failIfCalled` fake in 3 tests enforces this — if you accidentally call prompt in either branch, those tests fail loudly.
- `chooseStore` returns the choice **verbatim**; `resolveStore` is the only place `filepath.Abs` runs. If you move `filepath.Abs` into `chooseStore`, `TestChooseStoreTTYTypedPathOverrides` breaks (it asserts `/custom` literally).
- `readPrompt` must treat `io.EOF` with empty text as "accept default", not an error — only `err != nil && err != io.EOF` propagates. `TestChooseStorePropagatesPromptError` confirms propagation works; the EOF path is covered structurally by `readPrompt` (the `init < /dev/null` acceptance case).
- Do NOT add `if c.init { … }` to `run()` (S3) and do NOT mkdir/seed/writeconfig (S2). `resolveStore` is defined and uncalled.

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **`chooseStore` takes `isTTY` + `prompt` as PARAMETERS (not globals).** The contract FACTORING clause mandates this: the pure decision core is terminal-free and unit-testable by injecting a fake prompt fn and a bool. This diverges from the existing `isTerminal` package-var pattern (used because `run()` has no isTTY param), but is cleaner for init — no package-state mutation in tests, no `withStdinTerminal` helper needed.
2. **`chooseStore` returns verbatim; `resolveStore` absolutizes.** Keeps the unit assertions literal (`/custom` in ⇒ `/custom` out) and concentrates all filesystem/terminal I/O in the single wrapper run() calls. The contract's "absolute chosen store dir" describes `resolveStore`'s output (the deliverable S2/S3 consume), not the decision core.
3. **`stdinIsTerminal` is a plain function, not a var.** Because the test seam is `chooseStore`'s `isTTY` parameter, there is no need to override the real check in unit tests. `resolveStore` (which calls `stdinIsTerminal()`) is integration-tested in S3/acceptance with a real pipe, not unit-tested here.
4. **The 4-step order is override → detect-default → off-TTY → on-TTY.** Checking `haveStore != ""` first (step 1) avoids a needless `HasSkillMD(cwd)` walk on the `--store` path and matches the contract LOGIC numbering. Computing `def` (step 2) before the `!isTTY` return (step 4) means the off-TTY branch returns the *auto-detected* default (cwd if it has skills, else the XDG default) — which is exactly PRD §8.2's non-interactive behavior.
5. **One shared `bufio.NewReader(os.Stdin)` in `resolveStore`.** external_deps.md §4 warns a fresh reader per prompt can swallow buffered bytes. There is one prompt today, but the shared reader is created once and captured by the closure so a future second prompt is safe.
6. **No `golang.org/x/term`.** The `ModeCharDevice` heuristic (already the repo's stdout pattern) is reused for stdin. The `/dev/null`-is-char-device edge is harmless (EOF ⇒ default). Keeps yaml.v3 the sole non-stdlib dep (PRD §4 / external_deps.md §3).
7. **`resolveStore` is defined but uncalled in this subtask.** Go permits unused package-level functions. Wiring it into `run()`'s `if c.init` is S3; adding it here would be scope creep and collide with S3.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod UNCHANGED. Two import additions, both already-resolved:
      "bufio"                                          (stdlib — go 1.25 ships it)
      "github.com/dabstractor/skilldozer/internal/config" (this module — P1.M1.T1.S2)
    No `go get`, no `go mod tidy`. git diff --quiet go.mod go.sum ⇒ "deps unchanged".

CONSUMERS (NOT built in this subtask — listed to fix the interface):
  - run() init dispatch (P1.M2.T2.S3): `if c.init { store, err := resolveStore(c.initStore);
    if err != nil { … exit 1 } … (S2 create+seed+writeconfig) … print --path + check }`.
    resolveStore(c.initStore) is the call site; c.initStore comes from P1.M2.T1.S1 parsing.
  - create+seed+writeconfig (P1.M2.T2.S2): takes resolveStore's absolute store STRING, does
    mkdir -p + seed example/SKILL.md + config.Save. This subtask's output is its input.

CONSUMED (already present — verified):
  - config.DefaultStore() (string, error) — internal/config/config.go:150 (P1.M1.T1.S2).
  - skillsdir.HasSkillMD(dir string) bool — internal/skillsdir/skillsdir.go:207 (P1.M1.T2.S1).

NO ROUTES / NO DATABASE / NO CONFIG-FIELD-ADDITIONS / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after editing main.go)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0
go build ./...                  # expect exit 0 (resolveStore uncalled is fine — pkg-level fn)
# Expected: zero output / exit 0.
```

### Level 2: Unit Tests (component validation — the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test ./... -run 'TestChooseStore' -v
# Expected: ALL 7 pass. The load-bearing assertions:
#   TestChooseStoreExplicitOverrideNoPrompt   -> /tmp/x, prompt NEVER called (GOTCHA #1).
#   TestChooseStoreCwdDetectNonTTY            -> cwd, prompt NEVER called (auto-detect).
#   TestChooseStoreNoSkillNonTTYUsesDefault   -> /def, prompt NEVER called (off-TTY silent).
#   TestChooseStoreTTYEmptyPromptAcceptsDefault -> /def (readPrompt empty ⇒ default).
#   TestChooseStoreTTYTypedPathOverrides      -> /custom VERBATIM (GOTCHA #3: no Abs in core).
#   TestChooseStoreCwdDetectIsAlsoTheTTYDefault -> cwd (HasSkillMD consulted even on TTY).
#   TestChooseStorePropagatesPromptError      -> error wrapped (GOTCHA #4: non-EOF propagates).

# Regression: the existing suite (parseArgs/run/path/list/search/check/all/help) stays green.
go test ./...   # expect exit 0 (purely additive — no symbol renamed/moved)
```

### Level 3: Whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0

# GOTCHA #9 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Manual: the 4 new symbols exist and compile (resolveStore is uncalled — that is correct
# scaffolding until P1.M2.T2.S3 wires it into run()).
go doc . stdinIsTerminal readPrompt chooseStore resolveStore 2>/dev/null | head -40
```

### Level 4: Creative & Domain-Specific Validation

```bash
# N/A for THIS subtask — there is no end-to-end init flow to exercise yet (run() dispatch is
# P1.M2.T2.S3; mkdir/seed/writeconfig is P1.M2.T2.S2). The 7 unit tests ARE the validation
# surface: they prove the 4-step decision + the no-prompt guarantee + the error path without
# a real terminal. The integration proof (a real `skilldozer init --store /tmp/x` writing the
# config and a piped `echo | skilldozer init` not hanging) lands in P1.M4.T1.S1's §13
# acceptance run, which calls resolveStore through run()'s S3 dispatch.

# If a quick sanity check is wanted (non-blocking, no TTY): confirm readPrompt's EOF⇒default
# behavior in isolation — but the unit tests already cover the empty-answer path, so this is
# optional.
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

- [ ] `chooseStore("/tmp/x", _, _, _, failIfCalled)` ⇒ `("/tmp/x", nil)`, prompt never called
- [ ] `chooseStore("", skillCwd, false, "/def", failIfCalled)` ⇒ `(skillCwd, nil)`, prompt never called
- [ ] `chooseStore("", emptyCwd, false, "/def", failIfCalled)` ⇒ `("/def", nil)`, prompt never called
- [ ] `chooseStore("", emptyCwd, true, "/def", fakeReturns(""))` ⇒ `("/def", nil)`
- [ ] `chooseStore("", emptyCwd, true, "/def", fakeReturns("/custom"))` ⇒ `("/custom", nil)` (verbatim)
- [ ] `chooseStore("", skillCwd, true, "/def", fakeReturns(""))` ⇒ `(skillCwd, nil)` (cwd is the TTY default too)
- [ ] `chooseStore("", emptyCwd, true, "/def", fakeErrors)` ⇒ `("", err)` (error propagated)
- [ ] `resolveStore` exists, compiles, and is uncalled (wired by P1.M2.T2.S3)

### Code Quality Validation

- [ ] `stdinIsTerminal` mirrors the stdout `isTerminal` `ModeCharDevice` technique on `os.Stdin`
- [ ] `readPrompt` matches external_deps.md §4 verbatim (empty/EOF ⇒ default; non-EOF ⇒ error)
- [ ] `chooseStore` is a pure function (no `os.*` / terminal access — all deps injected)
- [ ] `resolveStore` confines all `os.Stdin`/`os.Stdout`/`os.Getwd` access to one wrapper
- [ ] Doc comments cite PRD §8.2 and external_deps.md §3/§4
- [ ] Anti-patterns avoided (see below)
- [ ] No new dependencies; `bufio` (stdlib) + `internal/config` (this module) only

### Documentation & Deployment

- [ ] No doc files (the prompt wording is a runtime string; README init UX is P1.M4.T2.S1)
- [ ] No new environment variables (consumes existing `XDG_DATA_HOME` via config.DefaultStore)

---

## Anti-Patterns to Avoid

- ❌ Don't call `prompt` in the `haveStore != ""` or `!isTTY` branches — that breaks the prompt-safety guarantee (the `failIfCalled` tests enforce this). PRD §8.2 / decision #13.
- ❌ Don't put `filepath.Abs` inside `chooseStore` — it must return the choice verbatim so the unit assertions are literal (GOTCHA #3). Abs lives in `resolveStore`.
- ❌ Don't make `stdinIsTerminal` a package var or add a `withStdinTerminal` helper — the test seam is `chooseStore`'s `isTTY` parameter, not global mutation (GOTCHA #6).
- ❌ Don't treat `io.EOF` as a hard error in `readPrompt` — empty/EOF ⇒ default; only `err != nil && err != io.EOF` propagates (GOTCHA #4).
- ❌ Don't create a fresh `bufio.NewReader` per prompt — one shared reader in `resolveStore` (GOTCHA #5).
- ❌ Don't add `golang.org/x/term` — `ModeCharDevice` is reused; yaml.v3 stays the sole non-stdlib dep (GOTCHA #7).
- ❌ Don't add `if c.init { … }` to `run()` — that is P1.M2.T2.S3 (GOTCHA #8). `resolveStore` is defined and uncalled here.
- ❌ Don't mkdir, seed `example/SKILL.md`, or write `config.yaml` — that is P1.M2.T2.S2. This subtask only CHOOSES the dir.
- ❌ Don't run `go mod tidy` — both new imports are already resolved; it would be a no-op but could touch go.sum needlessly (GOTCHA #9).
- ❌ Don't touch README, completions, the example skill, or `internal/*` — those are sibling subtasks.

---

## Confidence Score

**9/10** — one-pass implementation success likelihood. The change is purely additive to two files and mirrors three already-verified patterns: the stdout `isTerminal` `ModeCharDevice` technique (main.go:96-112, transcribed in `research/verified_facts.md` §2) for `stdinIsTerminal`; external_deps.md §4's `readPrompt` (transcribed §3) verbatim; and the repo's "pure core + injected deps" testability convention (`run`/`parseArgs`/`exclusivityError`/`skillPath` are all terminal-free). The two consumed APIs (`config.DefaultStore`, `skillsdir.HasSkillMD`) are confirmed exported and present. The single non-obvious risk — re-ordering the 4 steps so the prompt is accidentally called on the non-interactive/off-TTY paths — is locked by the `failIfCalled` fake in 3 of the 7 tests. The one residual uncertainty is whether `main_test.go` already imports `errors` (needed only by `TestChooseStorePropagatesPromptError`); the PRP flags the conditional import add explicitly so the implementer checks and adds it if missing.
