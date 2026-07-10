# PRP — P1.M2.T2.S1: Embed the three completion scripts + `completionScript()` function

> **Subtask:** The data-embedding half of the `completion` subcommand (PRD §14.6 "Embedding (self-sufficient binary)"). Adds `import _ "embed"` + three package-scope `string` vars (`bashCompletion`/`zshCompletion`/`fishCompletion`) each with a `//go:embed completions/...` directive, plus a pure `completionScript(shell string) (string, bool)` switch that maps `bash`/`zsh`/`fish` → the matching embedded bytes (+true) and anything else → `("", false)`. The on-disk `completions/*` files are the **single source of truth** (PRD §14.6: "both read the same files"), so the embed is byte-identical to them. **It does NOT add run() dispatch, shell detection, or script emission** — those are P1.M2.T2.S2 (`runCompletion`/`detectShell`), which consumes this subtask's `completionScript`.
>
> **Scope:** Two existing files only — `main.go` (1 import + 3 embed vars + 1 function) and `main_test.go` (3 tests). No new files. No `internal/*` change. **No new dependency** (`embed` is stdlib, stable since Go 1.16; the module is on go 1.25). go.mod/go.sum byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** main.go import block (15-26), `var version` (43), and the runInit region (1046-1059) read at exact line numbers; the three completion files verified present + header-grepped (bash 69L, zsh `_skilldozer` 61L, fish 51L); `grep` confirms main.go has ZERO current `//go:embed` directives (the two `embed` string matches are unrelated comments). architecture/external_deps.md §embed read in full — it VERIFIED by live test + Go source analysis that `//go:embed completions/_skilldozer` works WITHOUT an `all:` prefix. The parallel sibling P1.M2.T1.S1 (parsing/config/USAGE/exclusivity) PRP was read as a CONTRACT — it adds `completion`/`completionShell` config fields + parse cases + USAGE + exclusivity, adds ZERO imports, and does NOT touch the import block / `var version` / runInit / embed / `completionScript` — disjoint regions, no collision.

---

## Goal

**Feature Goal**: Compile the three on-disk completion scripts (`completions/skilldozer.bash`, `completions/_skilldozer`, `completions/skilldozer.fish`) into the binary as package-scope `string` vars via `//go:embed`, and expose a pure `completionScript(shell) (string, bool)` accessor, so that P1.M2.T2.S2's `runCompletion` can emit the correct shell's script with zero runtime file I/O — making the binary self-sufficient for `go install` users with no repo clone (PRD §14.6 / §12.2 decision 9).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — add `_ "embed"` to the stdlib import group; add three `//go:embed` + three `string` vars after `var version` (line 43); add `completionScript(shell string) (string, bool)` before the `runInit` doc comment (line 1046).
2. `main_test.go` — 3 tests: `TestCompletionScriptMapping` (table-driven per-shell + shell-unique header), `TestCompletionScriptUnsupportedShell` (`("", false)`), `TestEmbeddedCompletionsMatchOnDisk` (the §14.6 byte-identity lockstep guard).

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `completionScript("bash"|"zsh"|"fish")` each return a non-empty string whose first line is the matching file's header + `true`; `completionScript("powershell")` → `("", false)`; and each embedded var is byte-identical to its on-disk `completions/*` file (the §14.6 guarantee). The build does NOT error on the `_skilldozer` embed (verified by external_deps.md).

---

## User Persona (if applicable)

**Target User**: A `go install` user (no repo clone) who runs `eval "$(skilldozer completion)"`, and P1.M2.T2.S2 (`runCompletion`) which calls `completionScript`. This subtask is the internal data layer; there is no direct user-facing surface.

**Use Case**: `runCompletion` (S2) resolves the shell, calls `completionScript(shell)`, and writes the returned bytes to stdout for the parent shell to `eval`/`source`. This subtask supplies that byte blob.

**User Journey**: (Internal) S2 dispatch → detectShell → `completionScript(shell)` → embedded bytes → stdout. This subtask delivers the embed + the accessor.

**Pain Points Addressed**: a non-self-sufficient binary (completions only work for clone users via §14.5 file copy); runtime file I/O / missing-file errors at completion time.

---

## Why

- **Implements PRD §14.6 "Embedding (self-sufficient binary)".** The three scripts are compiled in with `//go:embed` (stdlib, **no new dependency**), so `completion` works for `go install` users with no repo clone — consistent with the "binary is self-sufficient" decision (§12.2 / decision 9).
- **Holds the §14.6 single-source-of-truth invariant.** The on-disk `completions/` files remain authoritative; the embed reads them verbatim, so §14.5's manual source/copy path and this subcommand emit **identical bytes**. A verbatim test locks this.
- **No new dependency.** `embed` is stdlib (stable since Go 1.16); the module is go 1.25. yaml.v3 stays the sole non-stdlib module.
- **Unblocks P1.M2.T2.S2** (run() dispatch + detection + emission), which calls `completionScript(detectedShell)`. After this subtask the bytes + accessor exist; S2 only adds `if c.completion { return runCompletion(...) }` + `detectShell`/`runCompletion`.
- **Does NOT** touch `completions/*` (P1.M3.T1.S1 lockstep adds `completion` as a completable subcommand), the README (P1.M3.T1.S2 Mode B), or the run() dispatch/detection/emission (P1.M2.T2.S2).

---

## What

A 1-line import, three `//go:embed`+`var` pairs, one ~10-line pure function, and three tests. No behavior change (nothing calls `completionScript` yet — S2 wires it). No new files.

### Success Criteria

- [ ] `main.go` import block has `_ "embed"` in the stdlib group (after `bufio`, before `fmt`).
- [ ] Three package-scope `string` vars (`bashCompletion`, `zshCompletion`, `fishCompletion`), each preceded immediately by its `//go:embed completions/...` directive, placed after `var version` (main.go:43).
- [ ] The `_skilldozer` directive is the plain form `//go:embed completions/_skilldozer` (NO `all:` prefix — verified to work).
- [ ] `completionScript(shell string) (string, bool)` (before the runInit doc comment, ~line 1046) returns the matching embedded var + true for `bash`/`zsh`/`fish`; `("", false)` otherwise.
- [ ] `go test ./...` green, including the 3 new tests; existing tests unaffected (purely additive; `completionScript` is uncalled until S2 — Go permits unused package-level functions).
- [ ] `go.mod`/`go.sum` unchanged; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** The exact code is fixed verbatim by external_deps.md §"Recommended pattern" (which VERIFIED it compiles + runs by live test). The one subtlety — that `//go:embed completions/_skilldozer` works without `all:` — is proven by external_deps.md (live test + Go source analysis, transcribed in `research/verified_facts.md` §2). Placement anchors (import block 15-26, `var version` 43, runInit doc 1046) are verified-current. The three source files are confirmed present with line counts + shell-unique header substrings (§5). The `import _ "embed"` blank-import requirement (for `string`/`[]byte` embed targets) is the standard Go idiom, live-tested. The boundary with the parallel sibling (parsing) is fixed: disjoint regions, the sibling adds 0 imports. An implementer who has never seen this repo can do it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (the embed subtlety, the exact pattern, anchors, test plan)
- file: plan/003_3ace946c2a5c/P1M2T2S1/research/verified_facts.md
  why: "§1 the exact 3-part deliverable + scope boundary (NOT runCompletion/detectShell = S2).
        §2 THE critical subtlety: //go:embed completions/_skilldozer works WITHOUT all: (verified
        by live test + Go source analysis — the _/. exclusion is directory-walk-only). §3 the
        exact recommended pattern (verbatim from external_deps.md, already live-tested). §4 the
        verified placement anchors (import 15-26, var version 43, runInit doc 1046; the contract's
        '1057' is stale). §5 the three source files + shell-unique header substrings for tests.
        §6 disjointness from P1.M2.T1.S1 (parsing). §7 the 3-test plan + the §14.6 verbatim lockstep
        guard + the test-CWD gotcha."
  critical: "§2 (do NOT prefix _skilldozer with all: — it works plain, verified) and §3 (import _
             \"embed\" is REQUIRED for string-var embeds; three string vars, NOT embed.FS) are the
             two things most likely to be mishandled. §6 (do NOT add runCompletion/detectShell/dispatch
             — S2's scope) bounds the work."

# MUST READ — the authoritative embed research (verified by live test + Go source analysis)
- file: plan/003_3ace946c2a5c/architecture/external_deps.md
  why: "§'Embedding _skilldozer' proves the plain //go:embed works on the underscore file. §'Recommended
        pattern' gives the EXACT import + 3 vars + completionScript code (live-tested). §'Why NOT
        embed.FS' justifies three string vars over embed.FS (simpler, type-safe, no runtime lookups
        for 3 known static files). §'Existing dependencies' confirms embed is stdlib (Go 1.16+;
        module is go 1.25) and adds nothing to go.mod."
  section: "Go embed package (whole section)."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as the sibling lands)
- file: main.go
  why: "THE edit target. Import block @15-26 (stdlib group: bufio, fmt, io, os, path/filepath, strings —
        add _ \"embed\" after bufio, before fmt). var version = \"dev\" @43 (doc 40-42) — PLACE the 3
        embed vars immediately AFTER line 43, before the // usageText comment @45. runInit doc comment
        @1046-1058, func runInit @1059 — PLACE completionScript immediately BEFORE the doc comment @1046
        (it joins the bottom helpers: skillPath @785, resolveStore @929, runInit @1059). NOTE configpkg
        is the ALIAS for internal/config — irrelevant here, don't be confused by it."
  pattern: "Package-scope string var + directive: `//go:embed path\nvar name string` (directive
            IMMEDIATELY above the var; only blank lines/comments between). Pure accessor:
            `func f(x string)(string,bool){ switch x { case ...: return v,true } return \"\",false }`."

# MUST READ — the test file under edit (mirror these test shapes; pure-function + embedded-data tests)
- file: main_test.go
  why: "THE other edit target. The existing suite uses package main (white-box), bytes.Buffer,
        strings.Contains, t.TempDir, t.Setenv. These 3 new tests are PURE (no run(), no store, no env):
        completionScript is a pure switch over package-scope string vars. Mirror any table-driven
        test for the mapping test. The verbatim test uses os.ReadFile (already imported in main_test.go)."
  gotcha: "Do NOT call run(['completion']) — dispatch is S2 (today it falls to no-mode). These tests
           call completionScript() directly + read the embedded vars directly. No store/env/t.Chdir."

# READ-ONLY — the parallel sibling PRP (boundary: disjoint regions, no collision)
- file: plan/003_3ace946c2a5c/P1M2T1S1/PRP.md
  why: "Confirms P1.M2.T1.S1 (parsing) adds completion/completionShell config fields + case 'completion'/
        '--shell' + init-guard extension + exclusivity family + USAGE rows + 9 tests. It adds ZERO
        imports and does NOT touch the import block / var version / runInit / embed / completionScript.
        Disjoint; land in either order. It EXPOSES c.completion/c.completionShell which S2 (not this
        subtask) reads — but confirms the parser half is the sibling's job."

# READ-ONLY — PRD (the authority for the embed + single-source-of-truth contract)
- file: PRD.md
  why: "READ-ONLY. §14.6 (h3.19) 'Embedding (self-sufficient binary)': //go:embed (stdlib, no new dep);
        on-disk completions/ are the single source of truth; §14.5 path and this subcommand 'emit
        identical bytes (both read the same files)'. §14.4 lockstep (the verbatim test encodes it).
        §12.2 / decision 9 (binary is self-sufficient). §17 (no new runtime deps)."
  section: "h3.19 (§14.6 Embedding + Lockstep), h2.13 (§14), h2.16 (§17)."

# READ-ONLY — the contract (the orchestrator owns it)
- file: plan/003_3ace946c2a5c/tasks.json
  why: "P1.M2.T2.S1's CONTRACT block (INPUT/LOGIC/OUTPUT/DOCS) is authoritative. This PRP transcribes
        it; tasks.json wins on any conflict. NOTE: the contract DOCS '[Mode A] none beyond code' — the
        embed is an internal implementation detail, so NO README/doc surface here (README is P1.M3.T1.S2)."
```

### Current Codebase tree (the relevant slice)

```bash
$ cd /home/dustin/projects/skilldozer && ls main.go main_test.go go.mod completions/
main.go        main_test.go   go.mod
completions/:  skilldozer.bash   _skilldozer   skilldozer.fish
# go.mod: module github.com/dabstractor/skilldozer, go 1.25, require gopkg.in/yaml.v3 v3.0.1 (sole non-stdlib dep)
$ grep -n 'go:embed\|^var.*Completion\|completionScript' main.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be changed

```bash
main.go        # ADD: `_ "embed"` import; 3 //go:embed + 3 string vars (after var version); completionScript() (before runInit doc)
main_test.go   # ADD: 3 tests (mapping table + unsupported + verbatim lockstep)
# go.mod / go.sum — UNCHANGED (embed is stdlib; no new module)
```

| File | Change | Owner |
|---|---|---|
| `main.go` | Embed the 3 completion files as string vars + the pure `completionScript` accessor | Issue §14.6 contract + external_deps.md |
| `main_test.go` | Lock the mapping + unsupported-shell + §14.6 byte-identity | QA §14.6 |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL) — `//go:embed completions/_skilldozer` works WITHOUT an `all:` prefix.
// Go's embed package excludes _/.-prefixed names ONLY during DIRECTORY-WALK patterns, NOT for
// explicit single-file paths (src/cmd/go/internal/load/pkg.go resolveEmbed: a regular-file match
// only checks isBadEmbedName(), which passes _skilldozer). VERIFIED by external_deps.md live test +
// Go source analysis. Do NOT write `//go:embed all:completions/_skilldozer` or switch to
// `//go:embed all:completions` + embed.FS — the plain per-file form is correct and simplest.

// GOTCHA #2 — `import _ "embed"` (blank import) is REQUIRED. //go:embed targeting a `string` (or
// []byte) var needs the embed package present, but no embed symbol is referenced, so it must be a
// BLANK import. Without it the build fails. (If you used embed.FS you'd import "embed" non-blank —
// but this subtask uses three string vars per external_deps.md §'Why NOT embed.FS', so blank import.)

// GOTCHA #3 — the //go:embed directive MUST immediately precede its `var` (only blank lines or
// comments allowed between). So each directive sits DIRECTLY above its var; do NOT put one directive
// above a group of three vars. A shared intro comment may precede the first directive.

// GOTCHA #4 — three `string` vars, NOT embed.FS. external_deps.md §'Why NOT embed.FS': an embed.FS
// over all:completions adds runtime ReadFile lookups + error handling for zero benefit with 3 known
// static files. Three string vars are simpler, type-safe, compile-time-checked, and let
// completionScript be a trivial switch. The contract LOGIC §1 fixes this.

// GOTCHA #5 — `completionScript` is defined but UNCALLED until P1.M2.T2.S2 wires runCompletion. Go
// ALLOWS unused package-level functions (only unused imports + unused LOCAL vars are errors), so
// `go build`/`go vet`/`go test` are green with completionScript uncalled. Do NOT add a throwaway call
// or the run() dispatch (S2's scope).

// GOTCHA #6 — anchor placement by the runInit DOC COMMENT (main.go:1046), not the stale "1057" the
// contract cites. The func itself is @1059. PLACE completionScript immediately before the doc comment.
// (Line numbers will shift as the parallel sibling P1.M2.T1.S1 lands its mid-file edits; anchor by the
// "runInit doc comment" symbol, then re-locate if needed.)

// GOTCHA #7 — the embed must be byte-identical to the on-disk files (PRD §14.6 "both read the same
// files"). //go:embed reads the EXACT file bytes (incl. trailing newline), so it is identical by
// construction — but LOCK it with TestEmbeddedCompletionsMatchOnDisk (embed == os.ReadFile). This is
// the §14.6 lockstep guard and the strongest catch for a swapped directive or future post-processing.

// GOTCHA #8 — the verbatim test reads `os.ReadFile("completions/...")` with a RELATIVE path, relying
// on the test CWD = repo root (package main's dir), which is `go test`'s default (the existing suite
// relies on it; t.Chdir tests explicitly change away). Do NOT t.Chdir in that test. os is already
// imported in main_test.go.

// GOTCHA #9 — go.mod/go.sum UNCHANGED. `embed` is stdlib (no go.mod entry; stdlib modules are never
// in go.mod). Do NOT run `go mod tidy` (it would be a no-op for stdlib but could touch go.sum
// needlessly). Verify with `git diff --quiet go.mod go.sum`.

// GOTCHA #10 — no merge collision with the parallel sibling P1.M2.T1.S1. It edits the config struct,
// parseArgs switches, init guard, exclusivityError, usageText — all mid-file — and adds ZERO imports.
// This subtask edits the import block (line 15-26) + after var version (43) + before runInit (1046) +
// new tests. The import-block edit (inserting 1 line, `_ "embed"`) does not textually collide with the
// sibling's 0 import-line change. The changesets compose in either order.
```

---

## Implementation Blueprint

### Data models and structure

**No new types.** Three package-scope `string` vars + one pure function. No struct changes (the `completion`/`completionShell` config fields are added by the parallel sibling P1.M2.T1.S1, not here).

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — add `_ "embed"` to the import block
  - FILE: main.go (import block, ~line 15-26)
  - ADD `_ "embed"` in the STDLIB group, alphabetically AFTER `bufio` and BEFORE `fmt`
    (bufio < embed < fmt). Keep it in the same parenthesized group; do NOT touch the internal group.
  - Result (stdlib group):
      import (
          "bufio"
          _ "embed"
          "fmt"
          "io"
          ...
  - GOTCHA #2: blank import is REQUIRED (string-var embed targets reference no embed symbol).
  - GOTCHA #9: this adds nothing to go.mod (embed is stdlib).

Task 2: EDIT main.go — add the three //go:embed + string vars (after var version)
  - FILE: main.go (immediately AFTER `var version = "dev"` @line 43, BEFORE the // usageText comment @45)
  - ADD (GOTCHA #1 — plain per-file directives, NO all: prefix; GOTCHA #3 — directive immediately
    above each var; verbatim from external_deps.md §'Recommended pattern'):
      // Embedded shell-completion scripts (PRD §14.6 "Embedding (self-sufficient binary)"). The
      // on-disk completions/* files are the single source of truth; these vars embed them verbatim
      // so `completion` works for go install users with no repo clone (no runtime file I/O). embed is
      // stdlib (Go 1.16+; no new dependency). Three string vars (not embed.FS): the files are static
      // and known, so a trivial switch in completionScript is simpler than runtime ReadFile lookups.
      // NOTE: completions/_skilldozer embeds WITHOUT an `all:` prefix — Go's _/. exclusion applies
      // only to directory-walk patterns, not explicit file paths (verified).
      //go:embed completions/skilldozer.bash
      var bashCompletion string
      //go:embed completions/_skilldozer
      var zshCompletion string
      //go:embed completions/skilldozer.fish
      var fishCompletion string
  - GOTCHA #1: the `_skilldozer` line is the plain form (verified — do NOT add `all:`).

Task 3: EDIT main.go — add completionScript(shell) (before the runInit doc comment)
  - FILE: main.go (immediately BEFORE the `// runInit is the …` doc comment @line 1046; GOTCHA #6)
  - ADD (verbatim switch from external_deps.md; GOTCHA #5 — uncalled until S2, which is fine):
      // completionScript returns the embedded shell-completion script for shell ("bash"/"zsh"/
      // "fish") and true; for any other shell it returns ("", false). It is a pure switch over the
      // package-scope //go:embed vars (PRD §14.6) — no filesystem access. The bool lets runCompletion
      // (P1.M2.T2.S2) distinguish a known shell from an unsupported one (the §6.4 exit-2 path).
      // The bytes are verbatim from completions/* (the single source of truth); see
      // TestEmbeddedCompletionsMatchOnDisk for the byte-identity lock.
      func completionScript(shell string) (string, bool) {
          switch shell {
          case "bash":
              return bashCompletion, true
          case "zsh":
              return zshCompletion, true
          case "fish":
              return fishCompletion, true
          }
          return "", false
      }
  - gofmt -w will align the case bodies if needed (single-line returns are fine as written).

Task 4: EDIT main_test.go — add the 3 tests (pure-function + embedded-data; no store/env/run)
  - FILE: main_test.go (append a new block; package main. os is already imported.)
  - (4a) Mapping correctness (table-driven; shell-unique headers lock the right-file-loaded + that
         the directive wasn't swapped):
      // completionScript maps each supported shell to its embedded script + true (PRD §14.6). The
      // shell-unique header substring guards against a swapped //go:embed directive (bash header must
      // come from bashCompletion, not zshCompletion). Pure switch over package-scope vars — no store/env.
      func TestCompletionScriptMapping(t *testing.T) {
          cases := []struct{ shell, header string }{
              {"bash", "# Bash completion for skilldozer."},
              {"zsh", "#compdef skilldozer"},
              {"fish", "# Fish completion for skilldozer."},
          }
          for _, tc := range cases {
              got, ok := completionScript(tc.shell)
              if !ok {
                  t.Errorf("completionScript(%q): ok=false; want true", tc.shell)
                  continue
              }
              if got == "" {
                  t.Errorf("completionScript(%q): empty script; want the embedded bytes", tc.shell)
              }
              if !strings.Contains(got, tc.header) {
                  t.Errorf("completionScript(%q): missing header %q (possible swapped embed?)", tc.shell, tc.header)
              }
          }
      }
  - (4b) Unsupported shell (the bool runCompletion will use for the §6.4 exit-2 path):
      func TestCompletionScriptUnsupportedShell(t *testing.T) {
          got, ok := completionScript("powershell")
          if ok {
              t.Errorf("completionScript(powershell): ok=true; want false")
          }
          if got != "" {
              t.Errorf("completionScript(powershell): got %q; want empty", got)
          }
      }
  - (4c) The §14.6 byte-identity lockstep guard (GOTCHA #7 — embed must be verbatim; GOTCHA #8 —
         relative os.ReadFile relies on test CWD = repo root):
      // PRD §14.6: the on-disk completions/* files are the single source of truth and the embed must
      // emit identical bytes ("both read the same files"). This locks that invariant: each embedded
      // var is byte-identical to its source file. go test runs from the repo root (package main's dir),
      // so the relative completions/ path resolves. Catches a swapped directive or future post-processing.
      func TestEmbeddedCompletionsMatchOnDisk(t *testing.T) {
          cases := []struct{ shell, path string }{
              {"bash", "completions/skilldozer.bash"},
              {"zsh", "completions/_skilldozer"},
              {"fish", "completions/skilldozer.fish"},
          }
          for _, tc := range cases {
              embedded, ok := completionScript(tc.shell)
              if !ok {
                  t.Fatalf("completionScript(%q): ok=false; want true", tc.shell)
              }
              onDisk, err := os.ReadFile(tc.path)
              if err != nil {
                  t.Fatalf("os.ReadFile(%s): %v (test must run from the repo root)", tc.path, err)
              }
              if embedded != string(onDisk) {
                  t.Errorf("completionScript(%q) != on-disk %s: embed is %d bytes, file is %d bytes (§14.6 byte-identity violated)",
                      tc.shell, tc.path, len(embedded), len(onDisk))
              }
          }
      }

Task 5: VERIFY (isolated, then whole-module + invariants)
  - gofmt -l main.go main_test.go     # MUST print nothing (run gofmt -w if it lists a file)
  - go vet ./...                      # exit 0
  - go build ./...                    # exit 0 — CRITICAL: if the _skilldozer embed failed, build errors
                                      #   with 'pattern completions/_skilldozer: no matching files found'
                                      #   (it will NOT — GOTCHA #1, verified by external_deps.md)
  - go test -run 'CompletionScript|EmbeddedCompletions' -v ./...   # the 3 new tests pass
  - go test ./...                     # whole module green; completionScript uncalled is fine (GOTCHA #5)
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA #9
  - manual sanity: go run . completion >/dev/null 2>&1; echo "exit=$? (no dispatch yet → no-mode; NOT a build error)"
```

### Implementation Patterns & Key Details

```go
// The import (Task 1) — blank import for string-var embeds:
import (
	"bufio"
	_ "embed"
	"fmt"
	// ...
)

// The embed vars (Task 2) — directive immediately above each var; plain per-file paths (NO all:).
//go:embed completions/skilldozer.bash
var bashCompletion string
//go:embed completions/_skilldozer
var zshCompletion string
//go:embed completions/skilldozer.fish
var fishCompletion string

// The accessor (Task 3) — pure switch; uncalled until P1.M2.T2.S2.
func completionScript(shell string) (string, bool) {
	switch shell {
	case "bash":
		return bashCompletion, true
	case "zsh":
		return zshCompletion, true
	case "fish":
		return fishCompletion, true
	}
	return "", false
}
```

Notes easy to get wrong:
- `_ "embed"` is a **blank** import (string-var embeds reference no embed symbol). Don't import `"embed"` non-blank (that's for `embed.FS`).
- The `_skilldozer` directive is the **plain** form — no `all:` prefix (verified). Don't "fix" it by adding `all:` or switching to `all:completions` + `embed.FS`.
- Each `//go:embed` sits **immediately** above its `var` (a directive above a group of vars embeds only the first correctly / is invalid). Three directive+var pairs, not one directive + three vars.
- Don't add `runCompletion`/`detectShell`/the `if c.completion` dispatch — S2's scope. `completionScript` is uncalled here (Go allows it).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **Three `string` vars, not `embed.FS`.** external_deps.md §'Why NOT embed.FS': with 3 known static files, three string vars are simpler, type-safe, and compile-time-checked; `embed.FS` over `all:completions` would add runtime `ReadFile` lookups + error handling for no benefit. The contract LOGIC §1 fixes this.
2. **Plain per-file `//go:embed` (no `all:`), even for `_skilldozer`.** external_deps.md VERIFIED by live test + Go source analysis that the `_`/`.` exclusion is directory-walk-only; an explicit file path passes `isBadEmbedName()`. Using `all:` would be unnecessary and would change the pattern semantics.
3. **`completionScript(shell) (string, bool)` — bool signals support.** The `false` return is what `runCompletion` (S2) will use to emit the PRD §6.4 "unsupported shell" exit-2 path. A plain `string` return couldn't distinguish "unsupported" from "empty script". external_deps.md prescribes this exact signature.
4. **Placement: embed vars after `var version`; `completionScript` before `runInit`'s doc comment.** Keeps package-scope string vars grouped (`version` + the three completion vars) and puts the accessor among the bottom helpers (skillPath/resolveStore/runInit), matching the established layout. The contract's "1057" is stale; anchor by the runInit doc comment.
5. **A verbatim lockstep test (`TestEmbeddedCompletionsMatchOnDisk`).** PRD §14.6 explicitly mandates byte-identity ("both read the same files"). The embed is identical-by-construction, but the test encodes the requirement and catches a swapped directive or future post-processing. It reads the on-disk files via `os.ReadFile` (relative path; `go test` runs from the repo root).
6. **No README/doc change here.** The contract DOCS is "[Mode A] none beyond code" — the embed is an internal implementation detail. The README `completion` install path is P1.M3.T1.S2 (Mode B).

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod / go.sum UNCHANGED. `embed` is stdlib (no go.mod entry; never listed). (GOTCHA #9)

CONSUMERS (NOT built in this subtask — listed to fix the interface):
  - runCompletion (P1.M2.T2.S2): resolves the shell via detectShell(c.completionShell,
        os.Getenv("SKILLDOZER_SHELL"), loginShellBase()), calls completionScript(shell), and on
        (script, true) writes script to stdout + exit 0; on (_, false) emits the §6.4 unsupported-shell
        exit-2 message; on detection failure emits the §6.4 exit-1 message. run()'s dispatch is also S2:
        `if c.completion { return runCompletion(c, stdout, stderr) }`. After THIS subtask,
        completionScript + the embedded bytes exist for S2 to call.
  - P1.M3.T1.S1 edits completions/* (adds `completion` as a completable subcommand). That changes the
        on-disk files AND the embedded bytes together (rebuild), so TestEmbeddedCompletionsMatchOnDisk
        stays green. Land P1.M3.T1.S1 after this subtask (it depends on completion being a known command).

PARALLEL SIBLING (no conflict):
  - P1.M2.T1.S1 (parsing) edits config struct / parseArgs / init guard / exclusivityError / usageText +
    9 tests; adds 0 imports. This subtask edits import block + after var version + before runInit + 3
    tests. DISJOINT; the one shared region (import block) gets +1 line here vs +0 there. Compose cleanly.

NO ROUTES / NO DATABASE / NO CONFIG-FORMAT CHANGE / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after the main.go edits)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0
go build ./...                  # expect exit 0 — CRITICAL: the _skilldozer embed MUST succeed
# Expected: zero output / exit 0. If build errors with
#   'pattern completions/_skilldozer: no matching files found', recheck the directive spelling/path
#   (it will NOT fail — verified by external_deps.md GOTCHA #1).
```

### Level 2: Unit Tests (the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test -run 'CompletionScript|EmbeddedCompletions' -v ./...
# Expected: ALL 3 pass. The load-bearing assertions:
#   TestCompletionScriptMapping            -> bash/zsh/fish each (non-empty, true) + Contains its header.
#   TestCompletionScriptUnsupportedShell   -> completionScript("powershell") == ("", false).
#   TestEmbeddedCompletionsMatchOnDisk     -> each embedded var == os.ReadFile(on-disk file) (§14.6).

# Regression — the whole suite stays green (completionScript is uncalled; nothing renamed):
go test ./...   # expect exit 0 (purely additive)
```

### Level 3: Whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0  — CRITICAL: zero regressions

# GOTCHA #9 invariant: go.mod / go.sum byte-for-byte unchanged
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Scope invariants
grep -c '//go:embed' main.go            # expect 3 (the three completion directives)
grep -c 'import _ "embed"\|_ "embed"' main.go   # expect 1 (the blank import)

# Manual: the embed loaded real bytes (go run rebuilds, so the embed is fresh). completion has no
# dispatch yet (S2), so it falls to no-mode — this only confirms the BUILD + embed are healthy:
go run . completion >/dev/null 2>&1; echo "completion exit=$? (no dispatch yet → no-mode; the embed itself is validated by the unit tests)"
```

### Level 4: Behavioral spot-checks (lock the embed is scoped, no over-reach)

```bash
cd /home/dustin/projects/skilldozer

# 4a. The embedded bytes are reachable + correct via a tiny throwaway probe (then removed) — OR rely
#     on the unit tests (TestCompletionScriptMapping + TestEmbeddedCompletionsMatchOnDisk already prove
#     the bytes). The unit tests ARE the authoritative check; this is optional:
go test -run TestEmbeddedCompletionsMatchOnDisk -v ./... && echo "§14.6 byte-identity VERIFIED"

# 4b. The build did NOT silently drop the _skilldozer embed (a swapped/typo'd directive would either
#     fail the build OR fail TestCompletionScriptMapping's zsh-header check):
go test -run 'TestCompletionScriptMapping' -v ./... | grep -q 'zsh.*#compdef' \
  && echo "zsh embed loaded correctly (_skilldozer embedded without all:)" \
  || echo "check TestCompletionScriptMapping output"
# Expected: both PASS.
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — `gofmt -l` clean; `go vet ./...` exit 0; `go build ./...` exit 0 (the `_skilldozer` embed succeeds)
- [ ] Level 2 PASS — the 3 new tests pass (mapping + unsupported + verbatim lockstep)
- [ ] Level 3 PASS — `go build/vet/test ./...` all exit 0 (zero regressions); `git diff go.mod go.sum` → "deps unchanged"; `grep -c '//go:embed' main.go` → 3
- [ ] Level 4 PASS — the §14.6 byte-identity test passes; the zsh embed loads `#compdef skilldozer`

### Feature Validation
- [ ] `import _ "embed"` present in the stdlib import group
- [ ] Three `//go:embed completions/...` directives + three `string` vars (`bashCompletion`/`zshCompletion`/`fishCompletion`) after `var version`
- [ ] The `_skilldozer` directive is the plain form (no `all:`)
- [ ] `completionScript("bash"|"zsh"|"fish")` → (non-empty script whose header matches the shell, true)
- [ ] `completionScript("powershell")` → `("", false)`
- [ ] Each embedded var is byte-identical to its `completions/*` file (§14.6)

### Code Quality / Convention Validation
- [ ] Mirrors external_deps.md §'Recommended pattern' verbatim (the verified pattern)
- [ ] Doc comments cite PRD §14.6 and the single-source-of-truth invariant
- [ ] `completionScript` placed among the bottom helpers (before runInit's doc comment)
- [ ] Anti-patterns avoided (see below)
- [ ] No new dependency (`embed` is stdlib); go.mod/go.sum byte-for-byte identical

### Scope Discipline
- [ ] Did NOT add run() completion dispatch / `runCompletion` / `detectShell` / `loginShellBase` (P1.M2.T2.S2)
- [ ] Did NOT touch `completions/*` (P1.M3.T1.S1) or the README (P1.M3.T1.S2)
- [ ] Did NOT switch to `embed.FS` or add `all:` prefixes (three string vars + plain directives, per external_deps.md)
- [ ] Did NOT modify `PRD.md` (read-only), `tasks.json`, `prd_snapshot.md`, or `.gitignore`

---

## Anti-Patterns to Avoid

- ❌ **Don't prefix `_skilldozer` with `all:`** or switch to `//go:embed all:completions` + `embed.FS`. The plain per-file form works (verified) and is simplest. (GOTCHA #1, #4.)
- ❌ **Don't use a non-blank `import "embed"`.** String-var embeds reference no embed symbol; the blank `import _ "embed"` is required. Non-blank would be an unused-import error. (GOTCHA #2.)
- ❌ **Don't put one `//go:embed` above a group of three vars.** The directive must immediately precede EACH var (three directive+var pairs). (GOTCHA #3.)
- ❌ **Don't add `runCompletion`/`detectShell`/the `if c.completion` dispatch.** That is P1.M2.T2.S2; `completionScript` is uncalled here (Go allows it). (GOTCHA #5.)
- ❌ **Don't run `go mod tidy`.** `embed` is stdlib (no go.mod entry); tidy is a no-op for stdlib but could touch go.sum needlessly. (GOTCHA #9.)
- ❌ **Don't t.Chdir in the verbatim test.** It relies on the test CWD = repo root for the relative `completions/...` `os.ReadFile`. (GOTCHA #8.)
- ❌ **Don't touch `completions/*` or the README.** Completions are P1.M3.T1.S1; README is P1.M3.T1.S2 (Mode B). This subtask's only doc surface is in-code comments. (Scope discipline.)

---

## Confidence Score

**9.5/10** — This is a 1-line import + three directive/var pairs + a ~10-line pure switch + three tests, with the EXACT code fixed verbatim by external_deps.md §'Recommended pattern' (which VERIFIED it compiles + runs by live test), and the one real subtlety — that `//go:embed completions/_skilldozer` works without `all:` — proven by external_deps.md (live test + Go source analysis, transcribed in `research/verified_facts.md` §2). Placement anchors (import 15-26, `var version` 43, runInit doc 1046) are verified-current. The three source files are confirmed present with shell-unique headers for the mapping tests. The boundary with the parallel parsing sibling is fixed (disjoint regions; the sibling adds 0 imports). The 0.5 reservation is for the single placement anchor drift risk (the contract's "1057" is stale; the runInit doc comment is the stable anchor but line numbers shift as the sibling lands) and the verbatim test's reliance on the test-CWD = repo-root convention (standard for `go test`, documented in GOTCHA #8) — both mitigated by anchoring instructions.
