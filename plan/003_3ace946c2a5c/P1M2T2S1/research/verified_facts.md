# Verified Facts — P1.M2.T2.S1 (embed the three completion scripts + completionScript)

Researched against the LIVE codebase (main.go import block / var version / runInit
anchors read; the three completion files verified present + line-counted + header-grepped;
architecture/external_deps.md embed section read in full). The parallel sibling P1.M2.T1.S1
(completion parsing/config/USAGE/exclusivity) PRP was read as a CONTRACT — it adds
`completion`/`completionShell` config fields + `case "completion"`/`--shell` parsing +
USAGE + exclusivity, and explicitly does NOT touch `//go:embed`/`completionScript` (this
subtask's scope). Disjoint regions; no collision.

---

## §1. The deliverable (exactly three things, per contract + external_deps.md)

1. `import _ "embed"` — add to the stdlib import group (main.go:15-26). `embed` is NOT
   currently imported (grep-confirmed). The blank import `_ "embed"` is REQUIRED for
   `//go:embed` directives that target `string`/`[]byte` vars (no embed.FS symbol referenced).
   external_deps.md verified this with a live test (§below).
2. Three package-scope `string` vars, each with a `//go:embed` directive, placed right after
   `var version = "dev"` (main.go:43) — keeps all package-scope string vars together.
3. `completionScript(shell string) (string, bool)` — a trivial switch placed before the
   runInit doc comment (main.go:1046; the contract's "1057" is stale). Maps
   bash/zsh/fish → the matching embedded var + true; anything else → "", false.

That is the ENTIRE subtask. No run() dispatch, no detectShell, no runCompletion — those are
P1.M2.T2.S2 (per the sibling PRP's GOTCHA G and the contract OUTPUT "consumed by P1.M2.T2.S2").

---

## §2. The CRITICAL embed subtlety — `_skilldozer` embeds WITHOUT `all:` (verified)

external_deps.md §"Embedding _skilldozer" (read in full, lines 5-13):

> **VERIFIED by live test + Go source analysis:** `//go:embed completions/_skilldozer` works
> WITHOUT `all:` prefix.

Go's embed package excludes `_`/`.`-prefixed names ONLY during DIRECTORY-WALK patterns, NOT
for explicit single-file paths. From `src/cmd/go/internal/load/pkg.go` (`resolveEmbed`):
- Regular-file match → checks `isBadEmbedName()`, which passes `_skilldozer` (it is not
  `.git`/`.hg`/`.svn`/etc.).
- Directory match → walks the tree, skipping `_`/`.`-prefixed children (unless `all:`).

So all THREE directives use the plain explicit-file form:
```go
//go:embed completions/skilldozer.bash
//go:embed completions/_skilldozer      // NO all: prefix — verified
//go:embed completions/skilldozer.fish
```
If the build ERRORS with `pattern completions/_skilldozer: no matching files found`, that
would indicate the file is missing or the name is wrong — but it will NOT (verified + the
file exists, 61 lines). `go vet ./... && go build ./...` is the contract's verification.

---

## §3. The recommended pattern (verbatim from external_deps.md — already live-tested)

external_deps.md §"Recommended pattern" gives the EXACT code, confirmed to compile + run:

```go
import _ "embed"

//go:embed completions/skilldozer.bash
var bashCompletion string

//go:embed completions/_skilldozer
var zshCompletion string

//go:embed completions/skilldozer.fish
var fishCompletion string

func completionScript(shell string) (string, bool) {
	switch shell {
	case "bash": return bashCompletion, true
	case "zsh":  return zshCompletion, true
	case "fish": return fishCompletion, true
	}
	return "", false
}
```

Why three `string` vars (NOT embed.FS): external_deps.md §"Why NOT embed.FS" — an
`embed.FS` over `all:completions` would also work but adds runtime lookups + error handling
for no benefit with 3 known static files. Three `string` vars is simpler, type-safe, and
compile-time-checked. The contract LOGIC §1 fixes this: "Three `string` vars is the
idiomatic pattern for 3 known static files."

GOTCHA: the `//go:embed` directive MUST immediately precede its `var` (only blank lines /
comments allowed between). So each directive sits directly above its var, NOT above a group.
A shared intro comment can precede the first directive.

---

## §4. Placement anchors (verified-current line numbers)

- Import block: main.go:15-26. stdlib group = `bufio`, `fmt`, `io`, `os`, `path/filepath`,
  `strings`. ADD `_ "embed"` in the stdlib group, ALPHABETICALLY after `bufio` and before
  `fmt` (bufio < embed < fmt). The internal group starts after a blank line at line 24
  (`internal/check`, `configpkg` (ALIASED), `discover`, `resolve`, …). Do NOT touch internals.
- `var version = "dev"`: main.go:43 (doc comment 40-42). PLACE the three embed vars
  immediately AFTER line 43 and BEFORE the `// usageText is …` comment (line 45).
- `runInit`: doc comment main.go:1046-1058, `func runInit` at 1059. PLACE `completionScript`
  immediately BEFORE the runInit doc comment (line 1046) — it sits among the bottom helpers
  (skillPath @785, resolveStore @929, runInit @1059). The contract's "before runInit at
  1057" is stale; anchor by "the runInit doc comment".

NOTE: `configpkg` is the ALIAS for `internal/config` (avoids colliding with the `config`
struct in main.go). Not relevant to this subtask, but do not be confused if you see it.

---

## §5. The three source files (verified present, single source of truth per §14.6)

```
completions/skilldozer.bash   69 lines   (contract said 69 ✓)
completions/_skilldozer       61 lines   (contract said 59; current is 61 — drift, harmless)
completions/skilldozer.fish   51 lines   (contract said 51 ✓)
```

PRD §14.6: "The on-disk `completions/` files remain the **single source of truth**;
§14.5's manual source/copy path and this subcommand emit identical bytes (both read the
same files)." So the embed MUST be byte-identical to the on-disk files — lock this with a
verbatim test (§7). These files are EDITED by P1.M3.T1.S1 (adds `completion` as a
completable subcommand); that happens AFTER this subtask and updates both the embed and the
on-disk file together, so a verbatim test stays green.

Stable, shell-unique header substrings (for the mapping-correctness tests; survive
P1.M3.T1.S1's flag additions):
- bash → `# Bash completion for skilldozer.` (line 1; also `complete -F` is bash-specific).
- zsh  → `#compdef skilldozer` (line 1; the compdef directive — extremely stable).
- fish → `# Fish completion for skilldozer.` (line 1; also `complete -c skilldozer`).

---

## §6. Boundary with the parallel sibling P1.M2.T1.S1 (parsing) — no collision

P1.M2.T1.S1 (read its PRP in full) edits:
- config struct (~128): +`completion bool`, +`completionShell string`.
- parseArgs switches (~188-312): +`case "completion":`, +`case "--shell":` (×2 switches).
- init positional-capture guard (~290): +`&& next != "completion"`.
- exclusivityError (~722): +the completion family.
- usageText (~52): +the completion USAGE/EXAMPLES/OPTIONS rows.
- main_test.go: +9 parse/exclusivity/help tests.

It does NOT touch:
- the import block (it uses existing `strings`; adds ZERO imports).
- `var version` (main.go:43).
- the runInit region (1046+).
- any embed / completionScript / runCompletion.

So this subtask's edits (import block line 15-26 + after var version line 43 + before
runInit line 1046 + new completionScript/embed tests) are DISJOINT from the sibling's
regions. The import block is shared, but the sibling adds 0 import lines and this subtask
adds exactly 1 (`_ "embed"`) — no text-level merge conflict. Land in either order.

P1.M2.T2.S2 (the NEXT subtask) consumes this subtask's output: `run()` adds
`if c.completion { return runCompletion(c, stdout, stderr) }`, and runCompletion calls
`completionScript(detectedShell)` + `detectShell(...)` to emit the bytes. So this subtask
must EXPOSE `completionScript` (package scope) for S2 to call. Do NOT add runCompletion /
detectShell / the dispatch here (S2's scope).

---

## §7. Test design — 3 functions (mapping + unsupported + verbatim lockstep)

The contract OUTPUT lists no explicit tests, but a PRP must validate the deliverable. Three
test functions in main_test.go (package main):

1. `TestCompletionScriptMapping` (table-driven) — for each of bash/zsh/fish,
   `completionScript(shell)` returns a NON-EMPTY string + true, AND the string Contains the
   shell-unique header (§5). This locks the mapping correctness AND that the right file
   loaded (catches a swapped directive).
2. `TestCompletionScriptUnsupportedShell` — `completionScript("powershell")` → ("", false).
   (The `bool` is what runCompletion will use to emit the §6.4 "unsupported shell" exit-2 path.)
3. `TestEmbeddedCompletionsMatchOnDisk` (the §14.6 lockstep guard) — for each shell/file,
   `bashCompletion == string(os.ReadFile("completions/skilldozer.bash"))` etc. This directly
   encodes PRD §14.6's "emit identical bytes (both read the same files)". `go test` runs from
   the repo root (package main's dir), so the relative `completions/...` path resolves. This
   is the strongest guard against a swapped directive OR post-processing drift.

GOTCHA (test CWD): `os.ReadFile("completions/...")` relies on the test CWD = repo root,
which is `go test`'s default (the existing suite relies on it too — t.Chdir tests
EXPLICITLY change away from it). Do NOT t.Chdir in this test. If a future change relocates
the test runner, this is the one assumption to revisit.

No store fixture / SKILLDOZER_SKILLS_DIR / t.Setenv needed — these are pure-function +
embedded-data tests (completionScript is a pure switch over package-scope string vars).
