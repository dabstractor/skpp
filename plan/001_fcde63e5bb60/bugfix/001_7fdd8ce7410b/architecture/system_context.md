# System Context — skpp Bug Fix 001

## Project

`skpp` (skill path printer): a Go CLI that resolves skill tags to on-disk
skill directory paths so they can be loaded into `pi` via `pi --skill`.

- **Module**: `github.com/dabstractor/skpp` (go 1.25)
- **Only third-party dep**: `gopkg.in/yaml.v3` (PRD §4/§7.3 policy)
- **Build**: `go build -o skpp .` ; tests: `go test ./...` ; vet clean
- **Status**: all §13 acceptance criteria pass, all unit tests pass

## Architecture (data flow)

```
argv
 → main.parseArgs(args) → config            [main.go:145]
 → main.run(args, stdout, stderr) int        [main.go:228]
     precedence: help > version > unknownFlag > exclusivity > dispatch
     dispatch order: path > list > search > check > all > tags
     each mode body: skillsdir.Find() → discover.Index(dir) → []Skill
```

### Package map

| Package | File | Responsibility |
|---|---|---|
| `main` | `main.go` | arg parsing (`parseArgs`), dispatch (`run`), exit codes, modifiers (`skillPath`) |
| `skillsdir` | `internal/skillsdir/skillsdir.go` | locate `skills/` dir (§8 rules); `Source` enum + `String()` |
| `discover` | `internal/discover/discover.go` | `Frontmatter` type, `ParseFrontmatter()` |
| | `internal/discover/skill.go` | `Skill` struct, `BuildSkill()` |
| | `internal/discover/index.go` | `Index()` — WalkDir, builds `[]Skill` sorted by RelTag |
| `resolve` | `internal/resolve/resolve.go` | tag → skill resolution (§7.2 precedence) |
| `search` | `internal/search/search.go` | substring filter over §6.1 fields |
| `check` | `internal/check/check.go` | §9 validation rules |
| `ui` | `internal/ui/ui.go` | `--list`/`--search` table rendering (ANSI color) |

### Key types

```go
// skillsdir.go:25
type Source int  // SourceEnv=0, SourceSibling=1, SourceWalkUp=2
// String() → "SKPP_SKILLS_DIR" | "sibling of binary" | "ancestor of cwd" | "unknown"

// skillsdir.go:232
func Find() (dir string, src Source, err error)

// discover/skill.go:42
type Skill struct {
    Dir, RelTag, Name, Description string
    Keywords, Aliases []string
    Category   string
    HasFM      bool
    SourceFile string
}

// search.go:36
func Search(query string, skills []discover.Skill) []discover.Skill
// matches() at search.go:59 scans RelTag, Name, Description, Keywords ONLY

// ui.go:55
func PrintList(w io.Writer, skills []discover.Skill, useColor bool)
// padRight (ui.go:132) + wrapWords (ui.go:143) both use len() = byte length
```

## Critical contracts that fixes must preserve

1. **§13 acceptance gate**: `test "$(./skpp --path)" = "$PWD/skills"` — stdout
   for `--path` must remain exactly `<dir>\n`. Any source reporting must go to
   **stderr**, never stdout.
2. **§6.4 stdout discipline**: failed/unknown tag resolution prints nothing to
   stdout. `$(...)` consumers depend on this.
3. **Dependency policy**: `gopkg.in/yaml.v3` is the ONLY allowed third-party
   dependency (PRD §4/§7.3). Any display-width fix must use the Go stdlib.
4. **Atomicity**: multi-tag invocations buffer all paths; flush only if ALL
   resolve (main.go:433-451).

## Test patterns (for context_scope accuracy)

- `main_test.go`: `run(args, &stdoutBuf, &stderrBuf)` returns exit code; tests
  assert on byte-exact stdout, stderr content, and exit code. Uses
  `t.Setenv("SKPP_SKILLS_DIR", dir)` to control discovery deterministically.
- `search_test.go`: `Search(query, []Skill)` returns filtered slice; helper
  `sk(tag, name, desc, keywords...)` builds skills. Tests assert match/no-match.
- `ui_test.go`: `PrintList(&buf, skills, useColor)` then assert on `buf.String()`.
  `colOf(out, substring)` returns the byte column index of a substring (for
  alignment assertions). Helper `mk(tag, name, desc, hasFM)` builds skills.
- All tests use `bytes.Buffer` (non-TTY → no color) for deterministic output.
