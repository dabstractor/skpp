# PRP — P1.M2.T2.S3: `run()` init dispatch — orchestrate, print --path + check, exit codes, never-prompt guarantee

> **Subtask:** The DISPATCH half of `skilldozer init` (PRD §8.2). The LAST of four init subtasks: T1.S1 (parse `init`/`--store`) ✅, S1 (`chooseStore`/`resolveStore`) ✅, S2 (`setupStore`/`exampleSkillTemplate`) ✅ are all LANDED in the tree; S3 only ADDS (a) one `if c.init { return runInit(c, stdout, stderr) }` branch in `run()` immediately after the `exclusivityError` gate, and (b) one package-level `runInit(c config, stdout, stderr io.Writer) int` that orchestrates `resolveStore` → `config.Path()` → `setupStore` → then prints the `--path` rendering (dir + "found via") + the `check` report (mirroring the existing branches verbatim), then `return 0`. The bare-tag path is UNCHANGED and re-asserted by a test to never prompt (PRD §6.4 / §8.2 prompt-safety).
>
> **Scope:** Two existing files only — `main.go` (one MID-FILE insert in `run()` ~line 441 + one APPEND `runInit` at the file tail after `setupStore` @933) and `main_test.go` (2 new `run`-level tests: init-success + bare-tag-never-prompts). No new files. No `internal/*` change. **Zero new imports** (everything `runInit` calls is already imported — `configpkg` alias, `skillsdir`, `discover`, `check`, `fmt`, `io`, `path/filepath`). `go.mod`/`go.sum` byte-for-byte unchanged.
>
> **STATUS (verified at PRP-write time):** main.go + main_test.go + internal/{config,skillsdir,check,discover} read directly. `resolveStore(haveStore) (string, error)` @854 (S1), `setupStore(store, configPath) (seeded, err)` @933 (S2), `const exampleSkillTemplate` @886 (S2), `c.init`/`c.initStore` @141 + exclusivity init-family @706 (T1.S1) all CONFIRMED PRESENT. The `configpkg` alias for `internal/config` confirmed (configpkg.DefaultStore @859, configpkg.Save/File @955) — so `config.Path()` is `configpkg.Path()`. The exact `exclusivityError`→`if c.path` anchor transcribed verbatim (main.go:438-448). The sibling PRPs (T1.S1, S1, S2) read as CONTRACTS; S1's Integration Points explicitly fixes the run() call site as `resolveStore(c.initStore)`.

---

## Goal

**Feature Goal**: Wire `skilldozer init` end to end so PRD §6.1 (init row) + §8.2 hold: `skilldozer init` / `init <dir>` / `init --store <dir>` resolve the store, create+seed it, write the config, and report (the configured store path to stdout per §6.1; the `--path` "found via" annotation + the `check` report per §8.2 step 5); exit 0 on setup success, 1 on setup failure. The bare `skilldozer <tag>` path stays untouched and is re-asserted to never prompt / never block / write nothing to stdout / exit 1 when unconfigured (PRD §6.4 / §8.2 prompt-safety).

**Deliverable**: Additive edits to two existing files:
1. `main.go` — insert `if c.init { return runInit(c, stdout, stderr) }` in `run()` right after the `exclusivityError` block (before `if c.path`); append `func runInit(c config, stdout, stderr io.Writer) int` at the file tail (after `setupStore`).
2. `main_test.go` — two `run`-level tests: `TestRunInitStoreWritesConfigCreatesStorePrintsPathExit0` and `TestRunBareTagUnconfiguredNeverPrompts`.

**Success Definition**: `go build/vet/test ./...` all pass; `gofmt -l main.go main_test.go` empty; `go.mod`/`go.sum` unchanged; `run(["init","--store",<tmp>])` exits 0 with the store created, config written (`configpkg.Load(cfg).Store == store`), and stdout containing the store path; `run(["someTag"])` under a clean env (unsetSkillsEnv + t.Chdir to a temp dir) exits 1 with stderr containing the `run \`skilldozer init\`` hint, empty stdout, and no stdin block; existing tests unaffected.

---

## User Persona (if applicable)

**Target User**: A first-run `skilldozer` user (and the scripts/CI that drive `init --store <dir>`). After this subtask, `skilldozer init` actually DOES something instead of no-op-ing to exit 1.

**Use Case**: `skilldozer init --store /path` (CI/scripts), `skilldozer init` inside an existing skills repo (adopts cwd), or `skilldozer init` on a fresh TTY (prompts, then sets up). Each lands a working store + config + a validation report.

**User Journey**: `skilldozer init --store ~/.local/share/skilldozer/skills` → resolveStore returns the absolute store → setupStore creates it + seeds the example template + writes `store:` to config.yaml → runInit prints the store path to stdout, `(found via config file)` to stderr, and the check report to stdout → exit 0. Next `skilldozer --list` / `skilldozer example` resolve via the config rule.

**Pain Points Addressed**: the binary now has a working first-run path (was a no-op exit 1 after T1.S1); a go-install user gets a seeded store + config in one command; `$(skilldozer badtag)` still fails loudly (the bare-tag path never prompts, so command substitution can't hang).

---

## Why

- **Closes gap G8** (`architecture/code_prd_delta.md` §3/§10): "run() has no `if c.init {…}` branch; entire init flow absent." This subtask IS that branch + the orchestrating `runInit`. T1.S1/S1/S2 delivered the building blocks (parse, choose, create); S3 assembles them into the live command.
- **Implements PRD §8.2 step 5** ("Print the output of `skilldozer --path` and `skilldozer check`") by mirroring the existing `--path` and `check` renderings inside `runInit`, so the user sees which §8.3 rule won and the validation result immediately after setup.
- **Satisfies PRD §6.1 init row** (stdout = the configured store path; exit 0/1): the store path is the stdout headline; exit 0 once create+config succeed.
- **Re-asserts the load-bearing prompt-safety guarantee** (PRD §8.2 / §6.4 / decision #13): the bare `<tag>` path never prompts. The guarantee is STRUCTURAL (stdin access is confined to `resolveStore`, only called inside `if c.init`); S3 locks it with a regression test.
- **Unblocks the §13 acceptance gate** (P1.M4.T1.S1): the gate's `init --store …` + `grep 'store:'` + unconfigured-hint block all depend on this dispatch existing.
- **Keeps yaml.v3 the sole non-stdlib dependency** (PRD §4): runInit uses only already-imported symbols; zero new imports.

---

## What

### Success Criteria

- [ ] `run()` inserts `if c.init { return runInit(c, stdout, stderr) }` immediately AFTER the `exclusivityError` block (main.go:438-441) and BEFORE the `// 5) Normal mode dispatch` comment / `if c.path` (main.go:443-448). init is exclusive, so this placement is collision-free.
- [ ] `runInit(c config, stdout, stderr io.Writer) int` is appended at the file tail (after `setupStore` @933) and: (1) `store, err := resolveStore(c.initStore)` → on err, `Fprintln(stderr, err); return 1`; (2) `cfgPath, err := configpkg.Path()` → on err, `Fprintln(stderr, err); return 1`; (3) `seeded, err := setupStore(store, cfgPath)` → on err, `Fprintln(stderr, err); return 1`; (4) uses `seeded` for a one-line stderr note ("Seeded example skill at …" / "Adopted existing store at …"); (5) calls `skillsdir.Find()` and prints `dir` to stdout + `(found via <src>)` to stderr (mirrors `--path`; falls back to `store` if Find errs); (6) renders the `check` report on `dir` to stdout (mirrors the `check` branch); (7) `return 0`.
- [ ] runInit uses ZERO new imports (`configpkg`/`skillsdir`/`discover`/`check`/`fmt`/`io`/`filepath` all imported).
- [ ] The bare-tag branch (`len(c.tags) > 0`) is UNCHANGED — no `resolveStore`/stdin call leaks into it.
- [ ] `main_test.go` adds 2 tests: init-success (creates store, writes config, stdout has store path, exit 0) and bare-tag-never-prompts (clean env → hint on stderr, empty stdout, exit 1, no hang).
- [ ] `go test ./...` green incl. the 2 new tests; existing tests unaffected (purely additive).
- [ ] `go.mod`/`go.sum` unchanged; no new files; `main.go` + `main_test.go` only.

---

## All Needed Context

### Context Completeness Check

**Pass.** Every call is pinned to a verified, live symbol: `resolveStore(haveStore) (string, error)` @854, `setupStore(store, configPath) (seeded, err)` @933, `configpkg.Path() (string, error)` (alias confirmed via configpkg.DefaultStore @859 / configpkg.Save @955), `skillsdir.Find() (dir, src, err)` + `Source.String()` labels, `discover.Index(dir) ([]Skill, error)`, `check.Check(skills) Report` with `BySkill`/`Errors`/`Warnings`/`HasErrors()` + `SkillReport{Skill, Findings}` + `Finding{Level, Message}` — all read from source and confirmed by the live `if c.path`/`if c.check` branches that already use them. The exact insertion anchor (main.go:438-448) is transcribed verbatim. The two renderings to mirror are copied line-for-line from the live branches. The never-prompt guarantee is traced to the structural fact that stdin access lives only in `resolveStore` (called only inside `if c.init`). The test patterns (unsetSkillsEnv + t.Chdir(t.TempDir()) for all-miss; the unconfigured-hint substrings; the init-success fixture) are mirrored from existing tests. An implementer who has never seen this repo can complete it in one pass.

### Documentation & References

```yaml
# MUST READ — the verified facts (signatures + anchor + renderings + the never-prompt trace)
- file: plan/002_38acb6d28a6a/P1M2T2S3/research/verified_facts.md
  why: "§1 = the siblings are LANDED (resolveStore @854, setupStore @933, c.init/initStore @141).
        §2 = the INPUT contract (exact signatures; CRITICAL: runInit calls resolveStore(c.initStore),
        NOT chooseStore — the item's 'chooseStore(...)' is shorthand). §3 = run() precedence +
        the EXACT insertion anchor (main.go:438-448, verbatim). §4 = the --path + check renderings
        to MIRROR (copied line-for-line). §5 = the runInit output design + the env-edge-case
        decision + exit-code semantics. §6 = the never-prompt guarantee (structural + test).
        §7 = zero new imports. §8 = test patterns. §9 = sibling boundaries. §10 = §13 gate."
  critical: "§2 — runInit calls resolveStore(c.initStore), not chooseStore (chooseStore is S1's pure
             5-arg core; resolveStore is the I/O wrapper purpose-built for this call site). §3 — the
             init dispatch sits AFTER exclusivity, BEFORE if c.path. §5 — init returns 0 once
             create+config succeed (check is a best-effort report, NOT a gate; differs from standalone
             check). §6 — the never-prompt test needs unsetSkillsEnv + t.Chdir(t.TempDir()) to escape
             the repo's walk-up rule (repo cwd HAS skills/example/SKILL.md)."

# MUST READ — the file under edit (locate symbols by NAME; line numbers shift as you edit)
- file: main.go
  why: "THE edit target. run() @408-668; the exclusivity→dispatch anchor @438-448 (insert the init
        branch between them). The if c.path branch @448-459 = the --path rendering to MIRROR. The
        if c.check branch @547-590 = the check rendering to MIRROR. The bare-tag branch @619
        = UNCHANGED (never-prompt guarantee). resolveStore @854, setupStore @933 = the appended
        runInit's placement target (append AFTER setupStore, the file tail). The import block has
        configpkg (alias), skillsdir, discover, check, fmt, io, path/filepath — ALL runInit needs."
  pattern: "run()-level mode handler = a self-contained `if c.<mode> { …; return <code> }` block that
            takes (stdout, stderr io.Writer) and returns int (so main() calls os.Exit(run(...)) and
            tests capture output via *bytes.Buffer). runInit is the init analogue of the check/path
            branches, but it ORCHESTRATES the S1/S2 helpers instead of just Find/Index."

# MUST READ — the consumed config package (configpkg alias; Path is the config-file LOCATION)
- file: internal/config/config.go
  why: "config.Path() (config.go:~140) = $SKILLDOZER_CONFIG literal, else $XDG_CONFIG_HOME/skilldozer/
        config.yaml (pure env fn, reads no fs). config.File{Store string} @30; config.Save(path, File)
        @69 (path FIRST). runInit calls configpkg.Path() to get cfgPath for setupStore. NOTE the alias:
        main.go imports it as configpkg, so the call is configpkg.Path() (NOT config.Path())."
  gotcha: "config.Path() can error (relative $XDG_CONFIG_HOME, or neither $XDG_CONFIG_HOME nor $HOME).
           runInit treats that as a setup failure: Fprintln(stderr, err); return 1 (can't determine
           where to write the config). It does NOT fall through silently."

# MUST READ — the consumed skillsdir package (Find + Source labels + ErrNotFound)
- file: internal/skillsdir/skillsdir.go
  why: "Find() @~275 = the 5-rule §8.3 ladder; returns (dir, src, ErrNotFound) when all miss.
        ErrNotFound.Error() = 'skilldozer is not configured; run `skilldozer init`' (verbatim, literal
        backticks). Source.String() labels: 'SKILLDOZER_SKILLS_DIR' | 'config file' | 'sibling of
        binary' | 'ancestor of cwd'. After setupStore writes the config, Find() resolves via the
        config rule (src='config file') UNLESS SKILLDOZER_SKILLS_DIR is set (env is priority #1)."
  gotcha: "runInit calls Find() AFTER setupStore (not before) so the just-written config is visible.
           In the common case dir==store (config rule wins). If SKILLDOZER_SKILLS_DIR is set, dir==env
           value (env beats config) — runInit honestly reports that (§5 env-edge-case, not §13-tested)."

# MUST READ — the test file under edit + the test-template source
- file: main_test.go
  why: "THE other edit target + the test-template source. unsetSkillsEnv @25 (sets
        SKILLDOZER_SKILLS_DIR='' AND SKILLDOZER_CONFIG=<temp non-existent>). writeSkillTree @40. The
        all-rules-miss idiom: unsetSkillsEnv(t) + t.Chdir(t.TempDir()) — used @237,377,591,849,1089,1348.
        TestRunPathFailureErrNotFound @235 = the EXACT template for the never-prompt test (same setup;
        tag arg instead of --path). imports configpkg (aliased) for config.Load round-trips."
  gotcha: "t.Chdir(t.TempDir()) is MANDATORY for the never-prompt test: the repo cwd
           (~/projects/skilldozer) HAS skills/example/SKILL.md, so without t.Chdir the walk-up rule
           HITS and the bare tag goes to resolve (UnknownError) instead of the unconfigured hint.
           Go 1.24+ t.Chdir; go.mod is `go 1.25`. For the init-success test do NOT use unsetSkillsEnv
           (it points SKILLDOZER_CONFIG at a non-existent path — you WANT config to be writable):
           set SKILLDOZER_CONFIG to a temp file + SKILLDOZER_SKILLS_DIR='' + t.Chdir(t.TempDir())."

# READ-ONLY — the gap analysis (G8 is THIS subtask)
- file: plan/002_38acb6d28a6a/architecture/code_prd_delta.md
  why: "§2 + §3 G8: 'run() has no if c.init {…} branch; entire init flow absent.' §0 #2 + the §8.2
        prompt-safety sub-gap (line ~178): once init lands, the bare-tag never-prompt guarantee must
        be RE-ASSERTED (init's prompt must be TTY-gated and must NOT leak into tag resolution). §10
        G8 row = this subtask. Confirms run() is @367-617 and the init dispatch is the missing piece."

# READ-ONLY — the consumed check package (Report/SkillReport/Finding field names to mirror)
- file: internal/check/check.go
  why: "Check(skills) Report @135; Report{BySkill []SkillReport, Errors, Warnings; HasErrors()} @71-79;
        SkillReport{Skill discover.Skill; Findings []Finding} @62-66; Finding{Level Severity; Message
        string} @56-59 (Level is a fmt.Stringer — %-5s). Skill.Name (frontmatter) + Skill.RelTag
        (canonical tag). runInit mirrors the live if c.check render verbatim, so these are confirmed."

# READ-ONLY — the sibling PRPs (the contracts S3 consumes + assembles)
- file: plan/002_38acb6d28a6a/P1M2T2S1/PRP.md
  why: "Defines resolveStore(haveStore) → (absStore, error) — the I/O wrapper run()'s init dispatch
        calls (S1's Integration Points LITERALLY writes the run() call site as resolveStore(c.initStore)).
        Confirms haveStore != '' short-circuits BEFORE the prompt (never blocks for --store/init <dir>),
        and that stdin access is confined to resolveStore (the never-prompt guarantee's structural root)."
- file: plan/002_38acb6d28a6a/P1M2T2S2/PRP.md
  why: "Defines setupStore(store, configPath) (seeded, err) — mkdir + seed-if-empty + config.Save.
        Confirms seeded is 'a SUCCESS-PATH signal (it tells run()/S3 which message to print)' — so
        runInit USES seeded for the Seeded/Adopted stderr note (and to avoid Go's declared-not-used
        compile error on the `seeded, err :=` binding). store is already absolute (resolveStore
        absolutized); configPath comes from config.Path()."
- file: plan/002_38acb6d28a6a/P1M2T1S1/PRP.md
  why: "Defines c.init / c.initStore + case 'init' / --store + the exclusivity init-family. Confirms
        init is an EXCLUSIVE mode (exclusivityError rejects init+tags / init+any-mode → exit 2), so by
        the time run() passes exclusivity with c.init true, no other mode flag is set — the init
        dispatch can sit anywhere among the single modes (S3 places it first, right after exclusivity)."

# READ-ONLY — PRD (source of truth for the init dispatch contract + the never-prompt guarantee)
- file: PRD.md
  why: "§6.1 (h3.1) init row: stdout = the configured store path; exit 0/1. §8.2 (h3.9) step 5: print
        --path output + check; prompt-safety (load-bearing): bare <tag> never prompts. §6.4 (h3.4):
        unconfigured ⇒ stderr one-line fix, nothing on stdout, exit 1. §13 (h2.12): the acceptance
        block (init --store + grep config + unconfigured hint)."
  section: "h3.1 (§6.1 init row), h3.9 (§8.2 step 5 + prompt-safety), h3.4 (§6.4), h2.12 (§13)."
```

### Current Codebase tree

```bash
$ cd /home/dustin/projects/skilldozer && ls *.go
main.go          # EDIT: INSERT `if c.init { return runInit(c, stdout, stderr) }` in run() @~441; APPEND func runInit after setupStore @933
main_test.go     # EDIT: +2 run-level tests (init-success, bare-tag-never-prompts)
internal/        # untouched (config/skillsdir/discover/check CONSUMED, not modified)
# go.mod / go.sum untouched (zero new imports)
$ grep -n 'runInit\|return runInit' main.go   # (empty today — purely additive)
```

### Desired Codebase tree with files to be added and responsibility of file

```bash
main.go          # ADD: the `if c.init` dispatch (mid-file) + func runInit (file tail)
main_test.go     # ADD: TestRunInitStore* (init success) + TestRunBareTagUnconfiguredNeverPrompts
```

**No new files.** All edits are additive to existing files.

| File | Responsibility |
|---|---|
| `main.go` | The init DISPATCH + orchestration: `run()` routes `c.init` to `runInit`, which calls resolveStore→config.Path→setupStore then prints the `--path` rendering + the `check` report. Mirrors (does not refactor) the existing `--path`/`check` renderings. |
| `main_test.go` | Lock the two contract behaviors via `run`-level tests: init-success (store created, config written, stdout has the store path, exit 0) and the never-prompt guarantee (bare tag, clean env → hint, empty stdout, exit 1, no hang). |

### Known Gotchas of our codebase & Library Quirks

```go
// GOTCHA #1 (CRITICAL — item-description clarification) — runInit calls resolveStore(c.initStore),
// NOT chooseStore. The item's LOGIC writes "store := chooseStore(...)", but chooseStore is S1's PURE
// 5-arg core (chooseStore(haveStore, cwd, isTTY, defaultStore, prompt)) — calling it from runInit
// would re-implement the I/O assembly (os.Getwd/config.DefaultStore/stdinIsTerminal/bufio prompt)
// that S1 already factored into resolveStore. S1's PRP Integration Points LITERALLY fixes the run()
// call site as resolveStore(c.initStore). So: store, err := resolveStore(c.initStore). Verified:
// resolveStore @854 is documented as "the thin I/O wrapper that run()'s init dispatch (P1.M2.T2.S3)
// calls" and returns the store ABSOLUTIZED.

// GOTCHA #2 (CRITICAL — placement) — insert `if c.init { return runInit(c, stdout, stderr) }` in
// run() AFTER the exclusivityError block (main.go:438-441) and BEFORE `if c.path` (main.go:448).
// init is EXCLUSIVE (exclusivityError @706 rejects init+anything → exit 2), so once c.init is true
// and exclusivity passed, NO other mode flag is set — the dispatch short-circuits cleanly. Do NOT
// put it inside the path/list/search/check/all/tags ladder (init is a peer, not a sub-case).

// GOTCHA #3 — runInit is DEFINED at the FILE TAIL (append after setupStore @933) but CALLED from
// run() mid-file (~441). Go permits package-level functions to be defined after their call site
// (no forward-reference error). The mid-file insert (run's body) and the tail append (runInit's
// definition) are NON-overlapping edits, so they compose regardless of edit order.

// GOTCHA #4 — `config.Path()` is `configpkg.Path()` in this file. main.go imports the config
// package with the ALIAS `configpkg` (configpkg "github.com/dabstractor/skilldozer/internal/config").
// resolveStore uses configpkg.DefaultStore() @859; setupStore uses configpkg.Save/configpkg.File @955.
// Match that: cfgPath, err := configpkg.Path(). Do NOT write `config.Path()` (bare `config` is not
// imported — it would be a compile error).

// GOTCHA #5 — `seeded` MUST be used. `seeded, err := setupStore(store, cfgPath)` with seeded never
// read is a Go compile error ("seeded declared and not used"). S2's PRP explicitly says seeded is "a
// SUCCESS-PATH signal (it tells run()/S3 which message to print: 'seeded example skill' vs 'adopted
// existing store')". So runInit prints a one-line STDERR note based on seeded:
//   if seeded { fmt.Fprintf(stderr, "Seeded example skill at %s\n", filepath.Join(store, "example", "SKILL.md")) }
//   else      { fmt.Fprintf(stderr, "Adopted existing store at %s\n", store) }
// STDERR (not stdout) so §6.1's "stdout = the configured store path" headline stays clean. If you
// prefer zero extra prose, discard with `_, err := setupStore(...)` — but the item writes `seeded,`
// and S2 designed seeded for this message, so USING it is the intended design.

// GOTCHA #6 — runInit calls skillsdir.Find() AFTER setupStore (not before), so the just-written
// config is visible to Find()'s config rule (priority #2). In the common case Find() returns the
// configured store with src="config file" (the item's "src now possibly 'config file'"). Do NOT
// print `store` directly with a hardcoded "config file" label — that would be WRONG if
// SKILLDOZER_SKILLS_DIR is set (env beats config, §8.3 rule 1; src would be "SKILLDOZER_SKILLS_DIR").
// Find() reports the truthful EFFECTIVE resolution.

// GOTCHA #7 (env-edge-case, documented, NOT §13-tested) — `SKILLDOZER_SKILLS_DIR=/A skilldozer init
// --store /B` creates+configures /B, but Find() reports /A (env wins). runInit prints /A to stdout +
// "(found via SKILLDOZER_SKILLS_DIR)" + checks /A. This HONESTLY tells the user their env overrides
// the just-written config. §13 runs init WITHOUT env (store == Find().dir == config store), so this
// edge is invisible to the acceptance gate. Do NOT special-case it.

// GOTCHA #8 (exit-code semantics — differs from standalone check) — init returns 0 once create+config
// succeed. The check report is a best-effort REPORT, NOT a gate: even if check finds ERRORs in an
// adopted store, init exits 0 (setup succeeded; the user sees the report and fixes their skills).
// This DIFFERS from standalone `skilldozer check` (which exits 1 on ERRORs). A Find()/discover.Index()
// failure AFTER setupStore is non-fatal: stderr note + return 0 (the store+config are correct). ONLY
// resolveStore / config.Path() / setupStore failures return 1.

// GOTCHA #9 (CRITICAL — the never-prompt test needs t.Chdir) — the repo cwd
// (~/projects/skilldozer) HAS skills/example/SKILL.md, so skillsdir.Find()'s walk-up rule (priority
// #4) HITS when run from the repo. To test the UNCONFIGURED path (ErrNotFound → hint), you MUST
// `t.Chdir(t.TempDir())` (Go 1.24+; go.mod is `go 1.25`) to a temp dir with no skills/ ancestor.
// The established idiom is `unsetSkillsEnv(t)` (neutralizes env + config rules) + `t.Chdir(t.TempDir())`
// (escapes walk-up). Without t.Chdir the bare-tag test would resolve via walk-up and fail differently.

// GOTCHA #10 — the never-prompt guarantee is STRUCTURAL: the bare-tag branch (main.go:619)
// calls skillsdir.Find() and on ErrNotFound prints the hint + return 1; it NEVER calls resolveStore
// (the only place os.Stdin is read). So `skilldozer someTag` with stdin=/dev/null cannot hang. S3's
// regression test RE-ASSERTS this; a hang would prove someone leaked resolveStore into the tag branch.
// Do NOT modify the tag branch — it is already correct.

// GOTCHA #11 — runInit MIRRORS the existing --path and check renderings (copy their Fprintf/Fprintln
// calls verbatim); it does NOT refactor them into shared helpers. Refactoring would touch passing
// code (risk) and is out of the safe one-pass scope. The ~12 duplicated lines (the check render) are
// acceptable; the item says "reuse the existing rendering ... mirroring". (Factoring a shared
// printCheckReport helper is an OPTIONAL cleanup a reviewer may suggest, but do not do it here.)

// GOTCHA #12 — runInit's `dir` for the check report is Find()'s dir (the EFFECTIVE store), NOT the
// `store` variable, EXCEPT in the Find()-failed fallback (dir = store). Keep them consistent: print
// `dir` to stdout AND run check on `dir`. (In the common case dir == store; in the env-edge-case
// they differ and check correctly validates the effective store.)

// GOTCHA #13 — the init dispatch is reached ONLY when c.init is true. Because exclusivityError
// (@706) runs BEFORE the dispatch and rejects init combined with any mode/tags, a bare
// `skilldozer someTag` sets c.init=false → never reaches runInit → never calls resolveStore. This is
// the architectural seam the never-prompt guarantee rests on. Do NOT weaken exclusivityError.
```

---

## Implementation Blueprint

### Data models and structure

**No new types.** runInit takes the existing `config` struct (for `c.initStore`) and the injected `(stdout, stderr io.Writer)`. It constructs no new structs — `configpkg.File` is written by `setupStore` (S2), `check.Report` is returned by `check.Check`.

### Implementation Tasks (ordered by dependencies)

```yaml
Task 1: EDIT main.go — insert the init dispatch in run() (mid-file)
  - FILE: main.go (func run, the exclusivity→dispatch boundary @438-448)
  - ANCHOR (match this text exactly with the edit tool):
        if bad, msg := exclusivityError(c); bad {
            fmt.Fprintln(stderr, msg)
            return 2
        }

        // 5) Normal mode dispatch (order: path → list → search → check → all →
        //    tags). Each branch body is byte-identical to pre-M5 (any mode that
        //    reaches here is guaranteed standalone: exclusivityError caught
        //    mode+mode/check+tags/check+mode above).

        if c.path {
  - INSERT, between the exclusivity block's closing `}` and the `// 5)` comment, the init dispatch:
        // init dispatch (PRD §8.2). init is an exclusive mode: exclusivityError above guarantees
        // no other mode is set when c.init is true, so this self-contained branch returns before
        // the path/list/search/check/all/tags ladder below. runInit orchestrates resolveStore →
        // config.Path → setupStore, then prints the --path rendering + the check report (§8.2
        // step 5). The bare-tag path (c.tags) is untouched and never prompts (§6.4).
        if c.init {
            return runInit(c, stdout, stderr)
        }
  - GOTCHA #2: AFTER exclusivity, BEFORE if c.path. GOTCHA #3: runInit is defined at the tail (Task 2).
  - (Optional cosmetic: renumber the subsequent `// 5)` comment to `// 6)`. Not required for correctness.)

Task 2: APPEND main.go — func runInit (the orchestrator; the file's new tail)
  - FILE: main.go (APPEND immediately after setupStore's closing brace @933 — the file tail)
  - ADD (GOTCHA #1/#4/#5/#6/#8/#11/#12 — resolveStore not chooseStore; configpkg.Path; use seeded;
    Find AFTER setupStore; return 0 on setup success; mirror the renderings; dir for both print+check):
      // runInit is the `skilldozer init` orchestrator (PRD §8.2). run()'s dispatch calls it when
      // c.init is true (init is exclusive, so no other mode is active). It assembles the three
      // already-landed helpers — resolveStore (P1.M2.T2.S1: choose+absolutize the store),
      // configpkg.Path (the config-file location), setupStore (P1.M2.T2.S2: mkdir+seed+writeconfig)
      // — and then reports: the configured store path to stdout (PRD §6.1), the `--path` "found via"
      // annotation to stderr, and the `check` report to stdout (PRD §8.2 step 5). Exit 0 once
      // create+config succeed; the check report is best-effort (NOT a gate — see GOTCHA #8).
      //
      // The bare `skilldozer <tag>` path NEVER reaches here (c.init is false for tags), so tag
      // resolution never prompts (PRD §6.4/§8.2 prompt-safety): stdin access is confined to
      // resolveStore, which only init calls.
      func runInit(c config, stdout, stderr io.Writer) int {
          // (1) Choose the store (haveStore != "" never blocks; resolveStore absolutizes).
          store, err := resolveStore(c.initStore)
          if err != nil {
              fmt.Fprintln(stderr, err) // one-line (resolveStore wraps with "skilldozer init: …")
              return 1
          }
          // (2) Resolve the config-file location (pure env fn; $SKILLDOZER_CONFIG or XDG default).
          cfgPath, err := configpkg.Path()
          if err != nil {
              fmt.Fprintln(stderr, err)
              return 1
          }
          // (3) Create the store, seed it if empty, write the config (PRD §8.2 steps 2-4).
          seeded, err := setupStore(store, cfgPath)
          if err != nil {
              fmt.Fprintln(stderr, err) // setupStore wraps with "skilldozer init: …"
              return 1
          }
          // (4) Report what happened. Uses `seeded` (S2's success-path signal). STDERR so §6.1's
          //     stdout headline (the store path) stays clean.
          if seeded {
              fmt.Fprintf(stderr, "Seeded example skill at %s\n", filepath.Join(store, "example", "SKILL.md"))
          } else {
              fmt.Fprintf(stderr, "Adopted existing store at %s\n", store)
          }
          // (5) Show the EFFECTIVE store + which §8.3 rule won (mirrors `skilldozer --path`, PRD
          //     §8.2 step 5). Find() runs AFTER setupStore so the just-written config is visible.
          //     In the common case dir == store and src == "config file"; if SKILLDOZER_SKILLS_DIR
          //     is set, env beats config and dir/src reflect that (GOTCHA #6/#7).
          dir, src, ferr := skillsdir.Find()
          if ferr != nil {
              // Should not happen (setupStore just wrote a valid config + created the store). Fall
              // back to the configured store so §6.1 (stdout = store path) still holds.
              fmt.Fprintln(stderr, ferr)
              dir = store
          }
          // §6.1: stdout = the configured store path (== dir, the effective resolved store).
          fmt.Fprintln(stdout, dir)
          if ferr == nil {
              // Mirror `skilldozer --path`: which rule won.
              fmt.Fprintf(stderr, "(found via %s)\n", src)
          }
          // (6) `skilldozer check` report on the effective store (PRD §8.2 step 5). Mirrors the
          //     `if c.check` branch render VERBATIM (GOTCHA #11 — do not refactor; mirror). Best-
          //     effort: a discover.Index failure is non-fatal (setup succeeded).
          skills, ierr := discover.Index(dir)
          if ierr != nil {
              fmt.Fprintln(stderr, ierr)
              return 0 // setup OK; the report is best-effort
          }
          rep := check.Check(skills)
          for _, sr := range rep.BySkill {
              name := sr.Skill.Name
              if name == "" {
                  name = "(none)"
              }
              if len(sr.Findings) == 0 {
                  fmt.Fprintf(stdout, "%-5s %s (%s)\n", "OK", sr.Skill.RelTag, name)
                  continue
              }
              for _, f := range sr.Findings {
                  fmt.Fprintf(stdout, "%-5s %s (%s): %s\n", f.Level, sr.Skill.RelTag, name, f.Message)
              }
          }
          fmt.Fprintf(stdout, "%d skills, %d errors, %d warnings\n", len(skills), rep.Errors, rep.Warnings)
          return 0 // setup succeeded; check findings do not change init's exit code (GOTCHA #8)
      }

Task 3: VERIFY (isolated — run after Task 2)
  - gofmt -l main.go          # MUST print nothing (run gofmt -w if it lists the file)
  - go vet ./...              # exit 0 (runInit called from run(); seeded used; no unused imports)
  - go build ./...            # exit 0

Task 4: EDIT main_test.go — add the 2 run-level tests
  - FILE: main_test.go (APPEND a new block; package main, white-box. configpkg already imported.)
  - (4a) Test #1 — init --store <tmp> writes config + creates store + prints store path + exit 0:
      func TestRunInitStoreWritesConfigCreatesStorePrintsPathExit0(t *testing.T) {
          // A store path that does NOT exist yet (under a temp parent) -> assert setupStore CREATES it.
          parent := t.TempDir()
          store := filepath.Join(parent, "newstore")
          cfg := filepath.Join(t.TempDir(), "config.yaml")
          t.Setenv("SKILLDOZER_CONFIG", cfg)   // redirect the config write to a temp file (NOT unsetSkillsEnv)
          t.Setenv("SKILLDOZER_SKILLS_DIR", "") // ensure the config rule wins (env unset)
          t.Chdir(t.TempDir())                  // escape the repo's walk-up rule (deterministic)
          var out, errOut bytes.Buffer
          code := run([]string{"init", "--store", store}, &out, &errOut)
          if code != 0 {
              t.Fatalf("run(init --store): code=%d; want 0; stderr=%q", code, errOut.String())
          }
          // store created (setupStore MkdirAll).
          info, err := os.Stat(store)
          if err != nil || !info.IsDir() {
              t.Errorf("store %q not created: stat err=%v", store, err)
          }
          // config written with store=<abs> (store is already absolute; resolveStore's Abs is idempotent).
          f, err := configpkg.Load(cfg)
          if err != nil {
              t.Fatalf("config.Load: %v", err)
          }
          if f.Store != store {
              t.Errorf("config.Store=%q; want %q", f.Store, store)
          }
          // §6.1: stdout contains the configured store path.
          if !strings.Contains(out.String(), store) {
              t.Errorf("init stdout=%q; want it to contain the store path %q", out.String(), store)
          }
      }
  - (4b) Test #2 — bare someTag under a clean env: hint on stderr, empty stdout, exit 1, no hang
         (the never-prompt guarantee; PRD §6.4/§8.2). Mirrors TestRunPathFailureErrNotFound @235:
      func TestRunBareTagUnconfiguredNeverPrompts(t *testing.T) {
          unsetSkillsEnv(t)        // neutralize env (SKILLDOZER_SKILLS_DIR) + config (SKILLDOZER_CONFIG) rules
          t.Chdir(t.TempDir())     // escape the repo walk-up rule (repo cwd HAS skills/example/SKILL.md)
          var out, errOut bytes.Buffer
          code := run([]string{"someTag"}, &out, &errOut)
          if code != 1 {
              t.Fatalf("run(someTag) unconfigured: code=%d; want 1", code)
          }
          if out.Len() != 0 {
              t.Errorf("run(someTag) stdout=%q; want EMPTY (§6.4: print nothing on failure)", out.String())
          }
          msg := errOut.String()
          for _, want := range []string{"run", "skilldozer init"} {
              if !strings.Contains(msg, want) {
                  t.Errorf("run(someTag) stderr=%q; missing substring %q (unconfigured hint)", msg, want)
              }
          }
          // The never-prompt guarantee is structural: the tag branch never calls resolveStore (the
          // only stdin reader). If this test HANGS, someone leaked resolveStore into the tag branch.
      }
  - GOTCHA #9: t.Chdir is MANDATORY for Test #2 (else walk-up hits the repo's ./skills).
  - GOTCHA: Test #1 does NOT use unsetSkillsEnv (it sets SKILLDOZER_CONFIG to a WRITABLE temp path).

Task 5: VERIFY (whole-module + invariants)
  - gofmt -l main.go main_test.go     # nothing
  - go vet ./...                      # exit 0
  - go test ./...                     # ALL pass incl. the 2 new tests; existing unaffected
  - git diff --quiet go.mod go.sum && echo deps unchanged   # GOTCHA (§7)
  - manual: go test -run 'TestRunInit|TestRunBareTag' -v ./...   # the 2 tests named + green
```

### Implementation Patterns & Key Details

```go
// The init dispatch in run() — AFTER exclusivity, BEFORE if c.path. init is exclusive.
if c.init {
	return runInit(c, stdout, stderr)
}

// runInit — orchestrate S1+S2 helpers, then mirror --path + check renderings. Exit 0 on setup
// success (check is best-effort). Uses `seeded` (required: S2's success-path signal).
func runInit(c config, stdout, stderr io.Writer) int {
	store, err := resolveStore(c.initStore) // GOTCHA #1: resolveStore, NOT chooseStore
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cfgPath, err := configpkg.Path() // GOTCHA #4: configpkg alias
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	seeded, err := setupStore(store, cfgPath) // GOTCHA #5: seeded MUST be used
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if seeded { // uses seeded (else Go "declared and not used")
		fmt.Fprintf(stderr, "Seeded example skill at %s\n", filepath.Join(store, "example", "SKILL.md"))
	} else {
		fmt.Fprintf(stderr, "Adopted existing store at %s\n", store)
	}
	dir, src, ferr := skillsdir.Find() // GOTCHA #6: AFTER setupStore; truthful effective resolution
	if ferr != nil {
		fmt.Fprintln(stderr, ferr)
		dir = store
	}
	fmt.Fprintln(stdout, dir) // §6.1: stdout = the store path
	if ferr == nil {
		fmt.Fprintf(stderr, "(found via %s)\n", src) // mirror --path
	}
	skills, ierr := discover.Index(dir)
	if ierr != nil {
		fmt.Fprintln(stderr, ierr)
		return 0 // setup OK; report is best-effort (GOTCHA #8)
	}
	rep := check.Check(skills)
	for _, sr := range rep.BySkill { // mirror the `if c.check` render VERBATIM (GOTCHA #11)
		name := sr.Skill.Name
		if name == "" {
			name = "(none)"
		}
		if len(sr.Findings) == 0 {
			fmt.Fprintf(stdout, "%-5s %s (%s)\n", "OK", sr.Skill.RelTag, name)
			continue
		}
		for _, f := range sr.Findings {
			fmt.Fprintf(stdout, "%-5s %s (%s): %s\n", f.Level, sr.Skill.RelTag, name, f.Message)
		}
	}
	fmt.Fprintf(stdout, "%d skills, %d errors, %d warnings\n", len(skills), rep.Errors, rep.Warnings)
	return 0 // setup succeeded; check findings do not change init's exit code
}
```

Notes easy to get wrong:
- The dispatch goes AFTER exclusivity. init is exclusive, so placement among the single modes is behaviorally free, but "first, right after exclusivity" is cleanest and matches the item.
- `resolveStore(c.initStore)` — not `chooseStore(...)`. The item's "chooseStore" is shorthand; resolveStore is the I/O wrapper S1 built for this call site (GOTCHA #1).
- `seeded` must be used (compile error otherwise) — the Seeded/Adopted stderr note is the designed use (S2's PRP).
- `Find()` is called AFTER `setupStore` (so the config is visible) and its `(dir, src)` is used for BOTH the stdout print AND the check target (coherence: the path shown == the store checked).
- init returns 0 on setup success even if check finds ERRORs — check is a report, not a gate (GOTCHA #8). Only resolveStore/config.Path/setupStore failures return 1.
- The never-prompt test REQUIRES `t.Chdir(t.TempDir())` — without it the repo's walk-up rule resolves and the bare tag never reaches the unconfigured hint (GOTCHA #9).

### DESIGN DECISIONS (the judgment calls this PRP resolves)

1. **runInit calls `resolveStore(c.initStore)`, not `chooseStore`.** resolveStore is the I/O wrapper S1 purpose-built for run()'s call site (it supplies cwd/defaultStore/TTY/prompt and absolutizes). Calling chooseStore directly would duplicate that assembly. The item's "chooseStore(...)" is shorthand for "the store-resolution step." (S1's Integration Points literally fixes the call as `resolveStore(c.initStore)`.)
2. **The dispatch sits right after exclusivity (before `if c.path`).** init is exclusive (exclusivityError rejects init+anything), so once `c.init` is true and exclusivity passed, no other mode is active — the dispatch can short-circuit immediately. This is the cleanest, earliest slot and matches the item ("after exclusivity, alongside the other single modes").
3. **runInit calls `skillsdir.Find()` (not `store` + hardcoded label) for the --path rendering.** The src label ("config file" vs "SKILLDOZER_SKILLS_DIR") REQUIRES Find(). Printing `store` + "config file" would lie if env is set (env beats config, §8.3). Find() reports the truthful EFFECTIVE resolution; in the common case dir == store.
4. **init returns 0 on setup success; check is a best-effort report, not a gate.** This differs from standalone `check` (exit 1 on ERRORs). Rationale: init's job is setup; a user who `init`s an existing repo with broken skills should still see exit 0 (setup done) + the check report (so they can fix their skills), not exit 1 (which would imply setup failed). §13 doesn't gate init's exit on check findings.
5. **`seeded` is used for a one-line stderr note ("Seeded …" / "Adopted …").** Required (Go compile error if unused) AND the designed use (S2's PRP: seeded tells S3 which message to print). STDERR so §6.1's stdout headline stays clean.
6. **runInit MIRRORS the existing --path/check renderings (no refactor).** Purely additive: ~12 lines of check-render duplication is preferable to touching passing code (the item says "reuse … mirroring"). A shared `printCheckReport` helper is an optional future cleanup, explicitly out of scope here.
7. **Find()/Index() failure after setupStore is non-fatal (stderr note, return 0).** The store + config WERE written (persistent); the reporting failure doesn't undo setup. resolveStore/config.Path/setupStore failures (the setup steps) DO return 1.

### Integration Points

```yaml
MODULE GRAPH:
  - go.mod UNCHANGED. runInit uses ONLY already-imported symbols:
      resolveStore/setupStore (same pkg), configpkg.Path() (main.go imports configpkg),
      skillsdir.Find/Source, discover.Index, check.Check/Report/SkillReport/Finding,
      filepath.Join, fmt, io. ZERO new imports in main.go or main_test.go.
    No `go get`, no `go mod tidy`. git diff --quiet go.mod go.sum ⇒ "deps unchanged".

CONSUMERS (verified present — this subtask assembles them):
  - resolveStore(haveStore) (string, error)  — main.go:854 (S1, returns ABS store).
  - setupStore(store, configPath) (seeded, err) — main.go:933 (S2, mkdir+seed+config.Save).
  - configpkg.Path() (string, error)         — internal/config/config.go (~140), aliased in main.go.
  - skillsdir.Find() (dir, src, err)         — internal/skillsdir/skillsdir.go (~275).
  - discover.Index(dir) ([]Skill, error)     — internal/discover/discover.go.
  - check.Check(skills) Report               — internal/check/check.go:135.

DOWNSTREAM (NOT built in this subtask):
  - §13 acceptance (P1.M4.T1.S1): runs `init --store …` + greps the config + the unconfigured-hint
    block — all depend on THIS dispatch existing. After S3, that gate is runnable.

NO ROUTES / NO DATABASE / NO CONFIG-FIELD-ADDITIONS / NO NEW FILES.
```

---

## Validation Loop

### Level 1: Syntax & Style (immediate, after editing main.go)

```bash
cd /home/dustin/projects/skilldozer

gofmt -l main.go main_test.go   # must print NOTHING (run gofmt -w if it lists a file)
go vet ./...                    # expect exit 0 (runInit called from run(); seeded used; no unused imports)
go build ./...                  # expect exit 0
# Expected: zero output / exit 0.
```

### Level 2: Unit Tests (component validation — the core gate)

```bash
cd /home/dustin/projects/skilldozer

go test ./... -run 'TestRunInit|TestRunBareTag' -v
# Expected: BOTH pass. The load-bearing assertions:
#   TestRunInitStoreWritesConfigCreatesStorePrintsPathExit0
#       -> exit 0; store dir created; configpkg.Load(cfg).Store == store; stdout Contains store;
#          (the seeded example/SKILL.md + "(found via config file)" + check report are also produced).
#   TestRunBareTagUnconfiguredNeverPrompts
#       -> exit 1; stderr Contains "run" + "skilldozer init"; stdout EMPTY; no hang (structural: tag
#          branch never calls resolveStore). If it HANGS, resolveStore leaked into the tag branch.

# Regression: the existing suite stays green (purely additive — no symbol renamed/moved). Critically,
# TestRunPathFailureErrNotFound @235 + the exclusivity init tests (T1.S1) + the S1/S2 unit tests.
go test ./...   # expect exit 0
```

### Level 3: Integration / whole-module regression + invariants

```bash
cd /home/dustin/projects/skilldozer

go build ./... ; echo "build exit $?"   # 0
go vet  ./...  ; echo "vet exit $?"     # 0
go test ./...  ; echo "test exit $?"    # 0

# go.mod / go.sum byte-for-byte unchanged (§7)
git diff --quiet go.mod go.sum && echo "deps unchanged" || echo "FAIL: deps changed"

# Manual end-to-end (mirrors §13's init block, isolated to a temp HOME/config so nothing pollutes):
iso=$(mktemp -d)
go build -o "$iso/skilldozer" .
store="$iso/store"
SKILLDOZER_CONFIG="$iso/cfg.yaml" "$iso/skilldozer" init --store "$store"
echo "init exit $?"
test -d "$store"                                                              && echo "store created OK"
grep -q "store: $store" "$iso/cfg.yaml"                                       && echo "config written OK"
SKILLDOZER_CONFIG="$iso/cfg.yaml" "$iso/skilldozer" --path | grep -q "$store" && echo "config rule wins OK"
SKILLDOZER_CONFIG="$iso/cfg.yaml" "$iso/skilldozer" check; echo "check exit $? (0: example skill is OK)"
# unconfigured hint (clean env, bare tag)
env -u SKILLDOZER_SKILLS_DIR HOME="$iso/home" XDG_CONFIG_HOME="$iso/home/.config" \
  "$iso/skilldozer" x 2>"$iso/err"; rc=$?
[ "$rc" = 1 ] && grep -q 'run `skilldozer init`' "$iso/err" && echo "unconfigured-hint OK"
rm -rf "$iso"
# Expected: all six echoes print OK; init exit 0; check exit 0.
```

### Level 4: Creative & Domain-Specific Validation

```bash
# The never-prompt guarantee under a REAL pipe (integration proof, beyond the unit test). A bare tag
# with stdin=/dev/null must print the hint + exit 1 + NOT hang (PRD §8.2 prompt-safety / §6.4):
iso=$(mktemp -d); go build -o "$iso/skilldozer" .
timeout 5 env -u SKILLDOZER_SKILLS_DIR HOME="$iso/home" XDG_CONFIG_HOME="$iso/home/.config" \
  "$iso/skilldozer" someTag </dev/null >"$iso/out" 2>"$iso/err"; rc=$?
[ "$rc" = 1 ] && [ ! -s "$iso/out" ] && grep -q 'run `skilldozer init`' "$iso/err" \
  && echo "never-prompt (stdin=/dev/null) OK" || echo "FAIL: rc=$rc out=$(wc -c <"$iso/out")"
rm -rf "$iso"
# `timeout 5` guards against a hang (would exit 124). Expected: never-prompt OK.

# Optional: confirm init's stdout is parseable as the store path (§6.1 headline), ignoring the
# trailing check report. (Not asserted by §13; informational.)
iso=$(mktemp -d); go build -o "$iso/skilldozer" .
head -1 <(SKILLDOZER_CONFIG="$iso/cfg.yaml" "$iso/skilldozer" init --store "$iso/store") | grep -q "^$iso/store$" \
  && echo "init stdout headline == store path OK"
rm -rf "$iso"
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

- [ ] `run(["init","--store",<tmp>])` exits 0; store dir created; `configpkg.Load(cfg).Store == store`; stdout contains the store path
- [ ] `run(["someTag"])` under `unsetSkillsEnv(t)` + `t.Chdir(t.TempDir())` exits 1, stderr contains "run" + "skilldozer init", stdout empty, no hang
- [ ] init dispatch sits AFTER exclusivity, BEFORE `if c.path` (collision-free: init is exclusive)
- [ ] init returns 0 on setup success even if check finds ERRORs (check is a report, not a gate)
- [ ] runInit calls `resolveStore(c.initStore)` (NOT chooseStore), `configpkg.Path()`, `setupStore`, then mirrors `--path` + `check` renderings

### Code Quality Validation

- [ ] runInit uses `seeded` (Seeded/Adopted stderr note) — no "declared and not used"
- [ ] runInit calls `skillsdir.Find()` AFTER setupStore and uses its `(dir, src)` for both the stdout print and the check target (coherent)
- [ ] The `--path` and `check` renderings are MIRRORED verbatim (no refactor of passing code)
- [ ] Zero new imports (`configpkg`/`skillsdir`/`discover`/`check`/`fmt`/`io`/`filepath` already imported)
- [ ] Doc comments cite PRD §6.1, §8.2 (step 5 + prompt-safety), §6.4
- [ ] Anti-patterns avoided (see below)

### Documentation & Deployment

- [ ] No doc files (the seeded/adopted note is a runtime string consistent with USAGE, already updated in T1.S1; README init UX is P1.M4.T2.S1 — Mode B, NOT touched here)
- [ ] No new environment variables

---

## Anti-Patterns to Avoid

- ❌ Don't call `chooseStore(...)` from runInit — call `resolveStore(c.initStore)` (GOTCHA #1). chooseStore is S1's pure 5-arg core; resolveStore is the I/O wrapper purpose-built for run()'s call site.
- ❌ Don't insert the init dispatch INSIDE the path/list/search/check/all/tags ladder — put it AFTER exclusivity, BEFORE `if c.path` (GOTCHA #2). init is exclusive; it short-circuits.
- ❌ Don't write `config.Path()` (bare) — the alias is `configpkg`; write `configpkg.Path()` (GOTCHA #4). Bare `config` is not imported.
- ❌ Don't leave `seeded` unused — it's a compile error. Use it for the Seeded/Adopted stderr note (GOTCHA #5), or discard with `_` (but using it is the designed intent).
- ❌ Don't print `store` + a hardcoded "config file" label for the --path rendering — call `skillsdir.Find()` to get the truthful src (env beats config; GOTCHA #6/#7).
- ❌ Don't make init exit 1 when check finds ERRORs — check is a best-effort report; init exits 0 once setup succeeds (GOTCHA #8). Only resolveStore/config.Path/setupStore failures exit 1.
- ❌ Don't omit `t.Chdir(t.TempDir())` from the never-prompt test — the repo cwd has `skills/example/SKILL.md`, so the walk-up rule would resolve and the bare tag would never reach the unconfigured hint (GOTCHA #9).
- ❌ Don't modify the bare-tag branch (`len(c.tags) > 0`) — it is already correct and never prompts (GOTCHA #10). S3 only RE-ASSERTS it with a test.
- ❌ Don't refactor the existing `--path`/`check` branches into shared helpers — mirror their renderings in runInit (purely additive; GOTCHA #11). Factoring touches passing code and is out of scope.
- ❌ Don't run `go mod tidy` — there are no new deps/imports; it would be a no-op but could touch go.sum needlessly (§7).
- ❌ Don't touch README, completions, the example skill, `internal/*`, or the S1/S2 helpers (resolveStore/setupStore/exampleSkillTemplate/chooseStore/readPrompt/stdinIsTerminal) — those are sibling/already-landed territory (§9).

---

## Confidence Score

**9/10** — one-pass implementation success likelihood. The change is purely additive to two files and consumes six already-verified, already-LANDED APIs (`resolveStore`, `setupStore`, `configpkg.Path`, `skillsdir.Find`, `discover.Index`, `check.Check`) whose exact signatures were read from source. runInit needs ZERO new imports. The dispatch placement is unambiguous (after exclusivity, before `if c.path` — the exact anchor is transcribed verbatim). The renderings to mirror are copied line-for-line from the live `--path`/`check` branches. The single semantic subtlety — that runInit calls `resolveStore(c.initStore)` (not `chooseStore`) — is called out explicitly with the S1 contract citation. The never-prompt guarantee is structural and locked by a regression test that mirrors the existing `TestRunPathFailureErrNotFound`. The one residual risk is whether an implementer confuses init's exit semantics with standalone `check`'s (init exits 0 on setup success even with check ERRORs) — the GOTCHA #8 callout and the test (which uses a seeded, clean store so check is OK anyway) close that gap. The env-edge-case (env beats config during init) is documented and invisible to §13.
