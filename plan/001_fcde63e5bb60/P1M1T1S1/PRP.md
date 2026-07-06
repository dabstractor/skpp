# PRP — P1.M1.T1.S1: `go.mod`, `go.sum`, `.gitignore`, `LICENSE`

> **Subtask:** P1.M1.T1.S1 (Repo & Go module scaffolding) — the FIRST subtask of the skpp one-shot build.
> **Scope boundary:** Creates the compilable Go module skeleton ONLY. Does NOT create `main.go`, `internal/`, `skills/`, README, or `install.sh` — those are later subtasks (P1.M1.T3, P1.M2, P1.M6).

---

## Goal

**Feature Goal**: Stand up a valid, compilable Go module skeleton for `github.com/dabstractor/skpp` with the single declared third-party dependency (`gopkg.in/yaml.v3`) and the repo-hygiene files (`go.mod`, `go.sum`, `.gitignore`, `LICENSE`) exactly as the PRD specifies, so every downstream subtask can build on it.

**Deliverable**: Four files at the repo root (`/home/dustin/projects/skpp/`):
1. `go.mod` — `module github.com/dabstractor/skpp`, `go 1.25` directive, `require gopkg.in/yaml.v3 v3.0.1 // indirect`
2. `go.sum` — checksums for `yaml.v3 v3.0.1` (and its test dep `check.v1`)
3. `.gitignore` — EXACTLY the five lines from PRD §16
4. `LICENSE` — standard MIT license text

**Success Definition**: `go build ./...` exits 0 in the repo root (the "matched no packages" warning is acceptable and expected). `go vet ./...` and `go mod verify` pass. Every file's content matches its spec exactly (verified by the validation commands below).

---

## Why

- This is the dependency-free foundation: **every** subsequent subtask (`internal/skillsdir`, `internal/discover`, `main.go`, etc.) imports the module path declared here. No downstream work can compile until `go.mod` exists.
- `yaml.v3` is the ONLY third-party dependency the entire project will ever pull in (PRD §7.3, §19 decision 8). Pinning it now means later subtasks just `import "gopkg.in/yaml.v3"` with no network/module-graph surprises.
- `.gitignore` (PRD §16) ignores the locally-built `skpp` binary (`/skpp`) so it is never committed — critical because `install.sh` (later) builds `skpp` in the repo root.
- `LICENSE` (MIT, PRD §19 decision 11) is required repo hygiene; the reference repo `mcpeepants` omits one, but the PRD explicitly mandates MIT for skpp.

---

## What

Four root-level files. No source code. No `main.go`. No packages. The module is intentionally empty-but-valid: it compiles, it declares its module path, and it records its one dependency. Concrete file contents are specified in the **Implementation Blueprint** below.

### Success Criteria

- [ ] `go.mod` exists at repo root with `module github.com/dabstractor/skpp`
- [ ] `go.mod` `go` directive is exactly `go 1.25` (NOT `1.26.4`, NOT `1.26`)
- [ ] `go.sum` exists and contains the `yaml.v3 v3.0.1` hash line
- [ ] `.gitignore` contains EXACTLY the 5 PRD §16 lines and nothing else
- [ ] `LICENSE` is a complete MIT license (copyright line, permission text, no-warranty clause)
- [ ] `go build ./...` exits 0 in repo root
- [ ] `go mod verify` reports `all modules verified`
- [ ] NO `main.go` or `internal/` package created in this subtask

---

## All Needed Context

### Context Completeness Check

_Pass: If someone knew nothing about this codebase, the four file specs + the three verified gotchas below give them everything to complete this subtask in one pass. No source-code knowledge is required — this is pure scaffolding._

### Documentation & References

```yaml
# MUST READ - authoritative specs for each file
- file: PRD.md
  why: "§5 (Target layout) + §16 (.gitignore) + §19 decision 11 (MIT license) are the source of truth"
  critical: "PRD.md is READ-ONLY. Do not modify it. Read §16 and §19 verbatim before writing .gitignore and LICENSE."

- file: plan/001_fcde63e5bb60/architecture/codebase_state.md
  why: "Confirms repo is greenfield (only PRD.md + .git); module path github.com/dabstractor/skpp; toolchain Go 1.26.4; single dep yaml.v3"
  pattern: "States the exact module path, Go-directive policy (latest two stable), and build command"

- file: plan/001_fcde63e5bb60/P1M1T1S1/research/verified_facts.md
  why: "Direct-execution proof of every gotcha in this PRP (go mod init output, // indirect behavior, go.sum contents, pre-existing .gitignore content)"
  critical: "Documents the THREE failure modes this subtask must avoid (see Known Gotchas)"

- url: https://go.dev/ref/mod#go-mod-file
  why: "Authoritative Go modules reference — go.mod file format, the `go` directive semantics, require directives"
  section: "The go directive (minimum required Go version); require directive syntax"

- url: https://go.dev/doc/modules/gomod-ref
  why: "go.mod file reference — confirms `go` directive is major.minor (no patch) and is a minimum-version floor, not a pin"

- url: https://choosealicense.com/licenses/mit/
  why: "Canonical MIT license full text (SPDX header + copyright line + permission grant + NO WARRANTY)"
  pattern: "Copy the full 'MIT License' text verbatim; fill the <year> and <copyright holder> placeholders"

- url: https://pkg.go.dev/gopkg.in/yaml.v3
  why: "Confirms v3.0.1 is the latest stable release (verified via `go list -m -versions`)"
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skpp && ls -A
.git/
.gitignore      # <-- PRE-EXISTING, WRONG CONTENT (generic template, see gotcha #3)
.pi-subagents/
PRD.md
plan/
```

There is **no `go.mod`, no `go.sum`, no `LICENSE`**. The `.gitignore` that exists was created by planning tooling and has the WRONG content for a Go project.

### Desired Codebase tree with files to be added

```bash
skpp/
├── PRD.md                  # exists — DO NOT TOUCH (read-only)
├── .gitignore              # OVERWRITE with PRD §16 content (see gotcha #3)
├── go.mod                  # CREATE — module github.com/dabstractor/skpp, go 1.25
├── go.sum                  # CREATE — yaml.v3 v3.0.1 checksums (auto from `go get`)
└── LICENSE                 # CREATE — MIT license
```

**File responsibilities:**
| File | Responsibility | Owner of contents |
|---|---|---|
| `go.mod` | Declares module path, Go version floor, the one dependency | PRD §5 (path) + contract (go 1.25) |
| `go.sum` | Cryptographic checksums for reproducible builds | Generated by `go get` — do NOT hand-edit |
| `.gitignore` | Keeps built binary + coverage artifacts out of git | PRD §16 — verbatim |
| `LICENSE` | Grants MIT rights to all skpp source | PRD §19 decision 11 — standard MIT text |

### Known Gotchas of our codebase & Go toolchain

```bash
# GOTCHA #1 — `go mod init` writes the FULL toolchain version, not 1.25.
# Verified: on Go 1.26.4, `go mod init github.com/dabstractor/skpp` produces:
#     go 1.26.4
# The contract requires `go 1.25`. You MUST edit the directive to `go 1.25`
# (major.minor, no patch) AFTER init. The `go` directive is a MINIMUM-version
# floor; 1.25 means "needs Go >= 1.25", which the 1.26.4 toolchain satisfies.

# GOTCHA #2 — yaml.v3 will be recorded as `// indirect`. This is CORRECT.
# `go get gopkg.in/yaml.v3` adds:
#     require gopkg.in/yaml.v3 v3.0.1 // indirect
# The `// indirect` comment is expected: no .go file imports yaml.v3 yet.
# DO NOT run `go mod tidy` to "clean it up" — tidy would DELETE the require
# line and its go.sum hash (nothing imports it). Leave `// indirect` as-is.
# It auto-promotes to a direct require once internal/discover imports it (P1.M2.T4).

# GOTCHA #3 — a PRE-EXISTING .gitignore has the WRONG content. OVERWRITE it.
# The repo root already has an untracked .gitignore (created by planning tooling)
# containing node_modules/, .env, venv/, .pi-subagents/, Thumbs.db, etc.
# This is WRONG for skpp. PRD §16 (authoritative) requires EXACTLY these 5 lines:
#     /skpp
#     /dist
#     *.test
#     *.out
#     .DS_Store
# Overwrite the whole file. A naive "does .gitignore exist?" check PASSES on the
# bad file — you MUST verify the CONTENT matches PRD §16 exactly.

# GOTCHA #4 — `go build ./...` on an empty module prints a warning but exits 0.
#     $ go build ./...
#     go: warning: "./..." matched no packages
#     $ echo $?
#     0
# The warning is HARMLESS and expected (no .go files yet). Do NOT add a dummy
# .go file or a placeholder package to silence it — that is explicitly out of
# scope (main.go and internal/ are later subtasks). Exit 0 == success.

# GOTCHA #5 — Do NOT run `go mod tidy` at the end of this subtask.
# Because no code imports yaml.v3 yet, `go mod tidy` would STRIP the require line
# and the go.sum entry, undoing the "add the dependency" requirement. Only
# `go mod init`, `go get`, and a manual directive edit are used here.
```

---

## Implementation Blueprint

### File specifications (exact contents)

#### `go.mod`

After running the commands below, `go.mod` MUST contain exactly:

```
module github.com/dabstractor/skpp

go 1.25

require gopkg.in/yaml.v3 v3.0.1 // indirect
```

(Trailing newline at EOF. Blank line between each block. No `toolchain` directive.)

#### `go.sum`

Generated by `go get` — do NOT hand-write. Expected contents:

```
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
```

#### `.gitignore` — EXACTLY PRD §16 (overwrite existing)

```
/skpp
/dist
*.test
*.out
.DS_Store
```

(Trailing newline at EOF. `/skpp` ignores the locally-built binary at repo root; `/dist` ignores release artifacts; `*.test`/`*.out` ignore Go coverage/profile outputs; `.DS_Store` ignores macOS noise.)

#### `LICENSE` — standard MIT (PRD §19 decision 11)

Use the canonical MIT text. Copyright holder and year are derived from the repo:
- Year: `2026`
- Holder: `Dustin Schultz` (from `git config user.name`; repo owner `dabstractor`)

```
MIT License

Copyright (c) 2026 Dustin Schultz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: INITIALIZE the Go module
  - COMMAND: cd /home/dustin/projects/skpp && go mod init github.com/dabstractor/skpp
  - RESULT: go.mod created with `module github.com/dabstractor/skpp` and `go 1.26.4`
  - GOTCHA: init writes the full version 1.26.4, NOT 1.25 (see gotcha #1)

Task 2: EDIT the go directive to 1.25
  - MODIFY: go.mod
  - CHANGE: the line `go 1.26.4` → `go 1.25`
  - WHY: contract mandates 1.25 (latest-two-stable policy); the directive is a
         minimum-version floor, and 1.25 + 1.26 are the two latest stable lines
  - VERIFY: grep -E '^go 1.25$' go.mod  (must print the line)

Task 3: ADD the yaml.v3 dependency
  - COMMAND: go get gopkg.in/yaml.v3
  - RESULT: go.sum created; go.mod gains `require gopkg.in/yaml.v3 v3.0.1 // indirect`
  - GOTCHA: the `// indirect` suffix is CORRECT and expected — do NOT run
            `go mod tidy`, which would strip it (see gotcha #2, #5)
  - VERSION: v3.0.1 confirmed latest via `go list -m -versions gopkg.in/yaml.v3`

Task 4: OVERWRITE .gitignore with PRD §16 content
  - WRITE: the 5 lines `/skpp`, `/dist`, `*.test`, `*.out`, `.DS_Store`
  - GOTCHA: a .gitignore ALREADY EXISTS with wrong generic content (node_modules,
            .env, .pi-subagents/, etc.) — OVERWRITE the entire file (gotcha #3)
  - VERIFY: `cat .gitignore` shows exactly 5 lines, nothing else
  - SOURCE: PRD §16 (verbatim)

Task 5: CREATE LICENSE (MIT)
  - WRITE: the standard MIT license text above
  - FILL: year = 2026, holder = "Dustin Schultz"
  - VERIFY: head -1 LICENSE prints "MIT License"; grep -c "Copyright (c) 2026" LICENSE = 1
  - SOURCE: PRD §19 decision 11 (MIT); canonical text from choosealicense.com

Task 6: VERIFY the skeleton compiles and is consistent
  - COMMAND: go build ./...     (expect exit 0 + harmless "matched no packages" warning)
  - COMMAND: go mod verify      (expect "all modules verified")
  - COMMAND: go vet ./...       (expect exit 0 + same harmless warning)
  - EXPECT: no errors. The warning is NOT an error (gotcha #4).
```

### Implementation Patterns & Key Details

```bash
# The complete command sequence for Tasks 1-3 (the only shell-driven part):
cd /home/dustin/projects/skpp
go mod init github.com/dabstractor/skpp   # writes go.mod with `go 1.26.4`

# CRITICAL edit (Task 2): downgrade the directive to major.minor 1.25.
# Replace the single line `go 1.26.4` with `go 1.25` in go.mod.

go get gopkg.in/yaml.v3                     # adds require + creates go.sum
# Do NOT run `go mod tidy` afterward.
```

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod declares: module github.com/dabstractor/skpp
  - go.mod requires: gopkg.in/yaml.v3 v3.0.1 (indirect now; direct once internal/discover imports it)
  - Downstream subtasks import internal packages as github.com/dabstractor/skpp/internal/<pkg>

GIT:
  - .gitignore now ignores /skpp (built binary) — install.sh (P1.M6.T13) builds `skpp` at repo root
  - plan/ and .pi-subagents/ are NOT in PRD §16 .gitignore (they are orchestration artifacts; leave git tracking as-is)

NO CONFIG / NO ROUTES / NO DATABASE:
  - This subtask has no config file, no API surface, no database. It is pure scaffolding.
```

---

## Validation Loop

### Level 1: File existence & exact content (immediate, per file)

```bash
cd /home/dustin/projects/skpp

# go.mod — exact content check (the three blocks, nothing else)
test -f go.mod || { echo "FAIL: go.mod missing"; exit 1; }
grep -qx 'module github.com/dabstractor/skpp' go.mod || { echo "FAIL: module path"; exit 1; }
grep -qx 'go 1.25'                go.mod || { echo "FAIL: go directive must be 'go 1.25' (got $(grep -E '^go ' go.mod))"; exit 1; }
grep -q 'require gopkg.in/yaml.v3 v3.0.1' go.mod || { echo "FAIL: yaml.v3 require missing"; exit 1; }
! grep -q 'go 1.26' go.mod || { echo "FAIL: go directive still 1.26.x"; exit 1; }

# go.sum — yaml.v3 hash present
test -f go.sum || { echo "FAIL: go.sum missing"; exit 1; }
grep -q 'gopkg.in/yaml.v3 v3.0.1 h1:' go.sum || { echo "FAIL: yaml.v3 hash missing from go.sum"; exit 1; }

# .gitignore — EXACTLY PRD §16 (5 lines, correct content)
test -f .gitignore || { echo "FAIL: .gitignore missing"; exit 1; }
diff <(printf '/skpp\n/dist\n*.test\n*.out\n.DS_Store\n') .gitignore \
  || { echo "FAIL: .gitignore does not match PRD §16 exactly"; exit 1; }

# LICENSE — MIT + copyright
test -f LICENSE || { echo "FAIL: LICENSE missing"; exit 1; }
head -1 LICENSE | grep -qx 'MIT License' || { echo "FAIL: LICENSE first line"; exit 1; }
grep -q 'Copyright (c) 2026' LICENSE || { echo "FAIL: LICENSE copyright line"; exit 1; }
grep -qi 'Permission is hereby granted' LICENSE || { echo "FAIL: LICENSE permission grant"; exit 1; }
grep -qi 'WITHOUT WARRANTY OF ANY KIND' LICENSE || { echo "FAIL: LICENSE no-warranty clause"; exit 1; }
echo "Level 1 PASS"
```

### Level 2: Module consistency (build + verify)

```bash
cd /home/dustin/projects/skpp

# Build the whole module (empty == success; warning is expected, NOT an error)
go build ./... ; BUILD_EXIT=$?
if [ $BUILD_EXIT -ne 0 ]; then echo "FAIL: go build ./... exited $BUILD_EXIT"; exit 1; fi
echo "go build ./... OK (warning 'matched no packages' is expected)"

# Module checksums verify
go mod verify || { echo "FAIL: go mod verify"; exit 1; }
echo "go mod verify OK"

# Vet (empty module == clean; warning expected)
go vet ./... ; VET_EXIT=$?
if [ $VET_EXIT -ne 0 ]; then echo "FAIL: go vet ./... exited $VET_EXIT"; exit 1; fi
echo "go vet ./... OK"

# Dependency list is exactly one third-party module (yaml.v3)
DEPS=$(go list -m all | grep -v '^github.com/dabstractor/skpp$' | grep -v '^gopkg.in/check.v1' || true)
test -z "$DEPS" || { echo "FAIL: unexpected modules: $DEPS"; exit 1; }
test -n "$(go list -m all | grep '^gopkg.in/yaml.v3 v3.0.1$')" || { echo "FAIL: yaml.v3 not in graph"; exit 1; }
echo "Level 2 PASS"
```

### Level 3: Scope-boundary check (do NOT exceed this subtask)

```bash
cd /home/dustin/projects/skpp

# MUST NOT have created main.go or any internal/ package (those are later subtasks)
test ! -e main.go     || { echo "FAIL: main.go must not exist in this subtask"; exit 1; }
test ! -d internal    || { echo "FAIL: internal/ must not exist in this subtask"; exit 1; }
test ! -d skills      || { echo "FAIL: skills/ must not exist in this subtask"; exit 1; }
test ! -e README.md   || { echo "FAIL: README.md must not exist in this subtask"; exit 1; }
test ! -e install.sh  || { echo "FAIL: install.sh must not exist in this subtask"; exit 1; }

# PRD.md must be unchanged (read-only)
git diff --quiet PRD.md || { echo "FAIL: PRD.md was modified (read-only)"; exit 1; }
echo "Level 3 PASS (scope boundary respected)"
```

### Level 4: Downstream-readiness smoke test

```bash
cd /home/dustin/projects/skpp

# Prove the module path is importable by a throwaway program (then remove it),
# confirming downstream subtasks can write `import "github.com/dabstractor/skpp/..."`.
mkdir -p /tmp/skpp-importcheck && cat > /tmp/skpp-importcheck/main.go <<'EOF'
package main
import (
	"fmt"
	_ "gopkg.in/yaml.v3"
)
func main() { fmt.Println("module importable, yaml.v3 resolves") }
EOF
cat > /tmp/skpp-importcheck/go.mod <<EOF
module skpp-importcheck
go 1.25
require github.com/dabstractor/skpp v0.0.0
replace github.com/dabstractor/skpp => $(pwd)
EOF
( cd /tmp/skpp-importcheck && go mod tidy 2>&1 | grep -v '^go: ' || true; go run . ) \
  || { echo "FAIL: module not importable by downstream code"; rm -rf /tmp/skpp-importcheck; exit 1; }
rm -rf /tmp/skpp-importcheck
echo "Level 4 PASS (downstream-ready)"
```

---

## Final Validation Checklist

### Technical Validation
- [ ] Level 1 PASS — all 4 files exist with exact specified content
- [ ] Level 2 PASS — `go build ./...` exits 0, `go mod verify` clean, `go vet` clean
- [ ] Level 3 PASS — no out-of-scope files created; PRD.md unchanged
- [ ] Level 4 PASS — module is importable by downstream Go code

### Feature Validation
- [ ] `go.mod` module path is exactly `github.com/dabstractor/skpp`
- [ ] `go.mod` `go` directive is exactly `go 1.25` (verified: `grep -qx 'go 1.25' go.mod`)
- [ ] `go.sum` contains the `yaml.v3 v3.0.1` checksum line
- [ ] `.gitignore` matches PRD §16 byte-for-byte (`diff` against the 5-line spec)
- [ ] `LICENSE` is complete MIT text (header + copyright + permission + warranty disclaimer)
- [ ] `go build ./...` succeeds (exit 0; "matched no packages" warning is acceptable)

### Code Quality / Convention Validation
- [ ] `go.mod` has no `toolchain` directive added (not needed; the `go 1.25` floor is sufficient)
- [ ] No `go mod tidy` was run (would have stripped the `// indirect` yaml.v3 require)
- [ ] The pre-existing wrong `.gitignore` was fully overwritten (not appended to)
- [ ] LICENSE copyright holder matches the git author (Dustin Schultz / dabstractor)

### Scope Discipline
- [ ] Did NOT create `main.go` (that is P1.M1.T3)
- [ ] Did NOT create `internal/` packages (those start at P1.M1.T2)
- [ ] Did NOT create `skills/example/` (that is P1.M6.T12)
- [ ] Did NOT modify `PRD.md` (read-only) or any `tasks.json` (orchestrator-owned)

---

## Anti-Patterns to Avoid

- ❌ **Don't run `go mod tidy`.** With no importing code, tidy deletes the yaml.v3 `require` and its go.sum hash, undoing Task 3. Leave `// indirect` alone.
- ❌ **Don't leave the `go` directive as `1.26.4`.** The contract requires `1.25` (latest-two-stable floor). `go mod init` writes the full version; you must edit it down.
- ❌ **Don't trust the pre-existing `.gitignore`.** It has generic template content (node_modules, .env, venv). Overwrite it entirely with PRD §16's 5 lines. Verify with `diff`, not "does it exist".
- ❌ **Don't silence the "matched no packages" build warning** by adding a placeholder `.go` file or empty package. `main.go` and `internal/` are explicitly out of scope. Exit 0 is the success signal.
- ❌ **Don't hand-write `go.sum`.** Let `go get` generate it. Hand-written hashes will fail `go mod verify`.
- ❌ **Don't add a `toolchain` directive.** It is unnecessary for this subtask and the PRD does not ask for one.
- ❌ **Don't create `main.go`, `internal/`, `skills/`, `README.md`, or `install.sh`.** Each is a separate later subtask.

---

## Confidence Score

**9/10** — This is deterministic scaffolding with all commands and exact file contents verified by direct execution in the target environment (see `research/verified_facts.md`). The 1-point reservation is only for the `.gitignore` overwrite decision: the pre-existing file was created by orchestration tooling, and while PRD §16 is unambiguously authoritative, a future orchestrator step might regenerate the generic template. The `diff`-based Level 1 check makes any such regression immediately visible.
