# Verified Facts — P1.M2.T2.S3: `run()` init dispatch

> Source of truth: read directly from the LIVE `main.go`, `main_test.go`,
> `internal/config/config.go`, `internal/skillsdir/skillsdir.go`,
> `internal/check/check.go`, and the three sibling PRPs (P1.M2.T1.S1, P1.M2.T2.S1,
> P1.M2.T2.S2). Every line number below is the CURRENT value in the working tree
> at PRP-write time (S1 + S2 are already LANDED — see §1 — so these anchors are
> stable; S3's own edits shift them only AFTER it lands).
>
> Mission of this file: pin every symbol/signature/anchor S3 consumes so an
> implementer who has never seen the repo can wire `run()`'s `if c.init` dispatch
> in one pass with zero guessing.

---

## §1. The siblings S3 consumes are ALREADY LANDED (verified in the live tree)

```
main.go:854   func resolveStore(haveStore string) (string, error)        // S1 (P1.M2.T2.S1) — Complete
main.go:886   const exampleSkillTemplate = `...`                         // S2 (P1.M2.T2.S2)
main.go:933   func setupStore(store, configPath string) (seeded bool, err error)  // S2 — landed
main.go:141   initStore   string   // (config struct field)             // T1.S1 (P1.M2.T1.S1) — Complete
main.go:706   if c.init { ... }     // (inside exclusivityError)         // T1.S1 — Complete
```

So S3 is the LAST of the four init subtasks: T1.S1 (parse) ✅, S1 (choose) ✅, S2 (create) ✅
are all in the tree. S3 only ADDS the dispatch + the orchestrating `runInit`. Nothing else is
pending. (S2's status field says "Implementing" in the plan, but `setupStore`/`exampleSkillTemplate`
are present in main.go — treat them as the contract.)

---

## §2. The INPUT contract (what S3 calls — exact signatures, verified live)

```go
// main.go:854 (S1) — the I/O wrapper run()'s init dispatch calls. Returns the
// store ABSOLUTIZED (filepath.Abs). Supplies os.Getwd/config.DefaultStore/
// stdinIsTerminal/a shared bufio prompt reader internally; haveStore != ""
// short-circuits BEFORE touching the prompt/stdin (never blocks for --store/init <dir>).
func resolveStore(haveStore string) (string, error)

// main.go:933 (S2) — create store (MkdirAll) + seed example/SKILL.md if empty +
// ALWAYS config.Save(configPath, File{Store: store}). (seeded=true only when it
// wrote the template; false when it adopted an existing non-empty store). store
// is already absolute (resolveStore absolutized it). Returns (false, err) on any
// fs failure (caller checks err FIRST).
func setupStore(store, configPath string) (seeded bool, err error)

// internal/config/config.go — config.Path is the config-file LOCATION (pure env fn:
// $SKILLDOZER_CONFIG literal, else $XDG_CONFIG_HOME/skilldozer/config.yaml).
// IMPORTED IN main.go AS THE ALIAS `configpkg` (NOT bare `config`):
//   configpkg "github.com/dabstractor/skilldozer/internal/config"     // main.go import block
// So the call is `configpkg.Path()` (mirrors resolveStore's `configpkg.DefaultStore()`
// at main.go:859 and setupStore's `configpkg.Save`/`configpkg.File` at main.go:955).
func config.Path() (string, error)                  // == configpkg.Path()
type config.File struct{ Store string `yaml:"store,omitempty"` }   // == configpkg.File

// internal/skillsdir/skillsdir.go — Find() runs the 5-rule §8.3 ladder; returns
// ErrNotFound (verbatim message "skilldozer is not configured; run `skilldozer init`")
// when EVERY rule misses. (dir, src, nil) on a hit.
func skillsdir.Find() (dir string, src Source, err error)
//   src.String() labels: "SKILLDOZER_SKILLS_DIR" | "config file" | "sibling of binary" | "ancestor of cwd"
//   After setupStore writes the config, the config rule (priority #2) resolves UNLESS
//   SKILLDOZER_SKILLS_DIR is also set (env is priority #1 and beats config — §8.3).

// internal/discover/discover.go — walk the store, return []Skill sorted by RelTag.
func discover.Index(dir string) ([]discover.Skill, error)

// internal/check/check.go — validate the skills; Report is the full output.
func check.Check(skills []discover.Skill) check.Report
//   check.Report{ BySkill []SkillReport, Errors int, Warnings int; HasErrors() bool }
//   check.SkillReport{ Skill discover.Skill; Findings []Finding }    // Findings empty => OK
//   check.Finding{ Level Severity; Message string }                   // Level is a fmt.Stringer (%-5s)
//   discover.Skill has .Name (frontmatter) and .RelTag (canonical tag) — both used by the check render.
```

**CRITICAL (item-description clarification):** the item's LOGIC writes "store := chooseStore(...)",
but `chooseStore` is S1's PURE 5-arg core (`chooseStore(haveStore, cwd, isTTY, defaultStore,
prompt)`) — calling it from runInit would re-implement the I/O assembly S1 already factored into
`resolveStore`. S1's PRP (Integration Points) explicitly fixes the run() call site as
`resolveStore(c.initStore)`. **So runInit calls `resolveStore(c.initStore)`** (NOT chooseStore).
The item's "chooseStore(...)" is shorthand for "the store-resolution step"; resolveStore is the
function. (Verified: resolveStore at main.go:854 is documented as "the thin I/O wrapper that
run()'s init dispatch (P1.M2.T2.S3) calls.")

---

## §3. The run() precedence + the EXACT insertion anchor (verified, line-stable)

run() dispatch order (main.go:408–668): `help → version → unknownFlag → exclusivity → path →
list → search → check → all → tags → no-args-usage`. init is an EXCLUSIVE mode (exclusivityError
at main.go:706 already rejects init+tags / init+any-mode → exit 2), so by the time control passes
exclusivity, `c.init` true implies every other mode flag is false. Therefore the init dispatch
can sit IMMEDIATELY AFTER exclusivity (before `if c.path`) — it short-circuits and returns before
the mode ladder is ever consulted. This matches the item: "Insert the init branch in the right
precedence slot (after exclusivity, alongside the other single modes like check; init is exclusive
so it won't collide)."

**EXACT anchor** (the text to match in the `edit` tool — main.go:438-448, copied verbatim):

```
	if bad, msg := exclusivityError(c); bad {
		fmt.Fprintln(stderr, msg)
		return 2
	}

	// 5) Normal mode dispatch (order: path → list → search → check → all →
	//    tags). Each branch body is byte-identical to pre-M5 (any mode that
	//    reaches here is guaranteed standalone: exclusivityError caught
	//    mode+mode/check+tags/check+mode above).

	if c.path {
```

INSERT between the exclusivity block's closing `}` (line 441) and the `// 5)` comment (line 443):

```go
	// init dispatch (PRD §8.2). init is an exclusive mode: exclusivityError above
	// guarantees no other mode is set when c.init is true, so this self-contained
	// branch returns before the path/list/search/check/all/tags ladder below.
	if c.init {
		return runInit(c, stdout, stderr)
	}
```

(Then renumber the subsequent `// 5)` comment to `// 6)` — cosmetic, optional.)

---

## §4. The renderings to MIRROR (item: "reuse the existing rendering")

S3 does NOT refactor the existing branches; it MIRRORS their output in runInit (purely additive,
no change to passing code). Copy these Fprintf/Fprintln calls verbatim.

**--path rendering** (main.go:448-459 — the `if c.path` branch body):
```go
	dir, src, err := skillsdir.Find()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1            // <-- runInit does NOT `return 1` here; see §5 (best-effort report)
	}
	fmt.Fprintln(stdout, dir)
	fmt.Fprintf(stderr, "(found via %s)\n", src)
```

**check rendering** (main.go:547-590 — the `if c.check` branch body):
```go
	skills, err := discover.Index(dir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1            // <-- runInit does NOT `return 1`; best-effort (see §5)
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
```

DIFFERENCE from the standalone branches: runInit's exit code reflects SETUP success (create+config),
NOT the --path/check sub-outcomes. So where the standalone branches `return 1` on Find/Index
failure or `return 1` when `rep.HasErrors()`, runInit keeps going / returns 0 (see §5 + §7).

---

## §5. runInit output design (the load-bearing decisions)

PRD §6.1 init row: stdout = "The configured store path"; exit 0 on success / 1 on error/cancel.
PRD §8.2 step 5: "Print the output of `skilldozer --path` (which rule won) and `skilldozer check`."

Reconciliation (the §6.1 stdout headline and the §8.2 step-5 --path rendering are the SAME line
in the common case: Find()'s dir == the just-configured store):

```
runInit(c, stdout, stderr) int:
  1. store, err := resolveStore(c.initStore)        // S1; absolutized. haveStore != "" never blocks.
     if err != nil { fmt.Fprintln(stderr, err); return 1 }
  2. cfgPath, err := configpkg.Path()               // config-file location (pure env fn).
     if err != nil { fmt.Fprintln(stderr, err); return 1 }
  3. seeded, err := setupStore(store, cfgPath)      // S2; mkdir + seed-if-empty + config.Save.
     if err != nil { fmt.Fprintln(stderr, err); return 1 }
  4. (uses `seeded` — required, else "declared and not used" compile error) report to STDERR:
     if seeded { fmt.Fprintf(stderr, "Seeded example skill at %s\n", filepath.Join(store,"example","SKILL.md")) }
     else      { fmt.Fprintf(stderr, "Adopted existing store at %s\n", store) }
  5. dir, src, ferr := skillsdir.Find()             // the EFFECTIVE store + which §8.3 rule won.
     if ferr != nil { fmt.Fprintln(stderr, ferr); dir = store }   // shouldn't happen post-config-write; fall back
     fmt.Fprintln(stdout, dir)                      // §6.1: stdout = the store path (== `store` common case)
     if ferr == nil { fmt.Fprintf(stderr, "(found via %s)\n", src) }   // mirror --path; src usually "config file"
  6. skills, ierr := discover.Index(dir)            // §8.2 step 5: the check report.
     if ierr != nil { fmt.Fprintln(stderr, ierr); return 0 }   // setup OK; report is best-effort
     rep := check.Check(skills)
     ... mirror the check rendering (§4) to stdout ...
     return 0                          // init success = create+config succeeded; NOT gated on check clean
```

**Why call Find() instead of printing `store` directly?** The item says "print the SAME output
`skilldozer --path` would (dir + '(found via <src>)' — src now possibly 'config file')". The src
label REQUIRES Find() (Source only comes from Find). Printing `store` + a hardcoded "config file"
would be WRONG if SKILLDOZER_SKILLS_DIR is set (env beats config, §8.3 rule 1; src would be
"SKILLDOZER_SKILLS_DIR"). Find() reports the truthful effective resolution.

**Env-edge-case (documented, NOT tested by §13):** `SKILLDOZER_SKILLS_DIR=/A skilldozer init
--store /B` creates+configures /B but Find() reports /A (env wins). init honestly shows /A +
"(found via SKILLDOZER_SKILLS_DIR)" + checks /A — telling the user their env overrides the config.
§13 runs init WITHOUT env, so store == Find().dir == config store (no edge).

**Exit code (design decision):** init returns 0 once create+config succeed. The check report is a
best-effort REPORT (informational), NOT a gate — so even if check finds ERRORs in an adopted store,
init still exits 0 (setup succeeded; the user sees the report and fixes their skills). This differs
from standalone `check` (which exits 1 on ERRORs). A Find()/Index() failure after setup is
non-fatal (stderr note, return 0); only resolveStore/config.Path/setupStore failures return 1.

---

## §6. The NEVER-PROMPT guarantee (structural + test-reasserted) — PRD §8.2 / §6.4

The bare-tag path (`len(c.tags) > 0` branch, main.go:619) calls `skillsdir.Find()`; on
ErrNotFound it prints the hint to stderr and `return 1`. It NEVER calls resolveStore/chooseStore/
readPrompt, and NEVER reads os.Stdin. So `skilldozer someTag` with no config + stdin=/dev/null
prints the hint, writes nothing to stdout, exits 1, and CANNOT hang. The guarantee is STRUCTURAL
(stdin access is confined to resolveStore, which is only called inside `if c.init`).

runInit's ONLY stdin access is via resolveStore → chooseStore's prompt, which is gated on
`stdinIsTerminal()` (S1). For `init --store <dir>` / `init <dir>`, `haveStore != ""` short-circuits
in chooseStore BEFORE the prompt is ever consulted — so even a non-TTY or /dev/null stdin never
blocks. (For bare `init` with no --store on a non-TTY, chooseStore returns the auto-detected
default WITHOUT prompting — S1's `!isTTY` branch. On a TTY it prompts, which is the documented
interactive case.)

**Test to re-assert (required by the item):** `run(["someTag"], ...)` under a clean env
(unsetSkillsEnv + t.Chdir(t.TempDir())) → exit 1, stderr contains "run" + "skilldozer init",
stdout EMPTY. Mirrors the existing TestRunPathFailureErrNotFound (main_test.go:235) with a tag
arg instead of --path. A hang here would prove someone leaked resolveStore into the tag branch.

---

## §7. Zero new imports (verified)

runInit uses ONLY already-imported symbols: `resolveStore`/`setupStore` (same pkg),
`configpkg.Path()` (main.go imports `configpkg`), `skillsdir.Find`/`Source` (imported),
`discover.Index` (imported), `check.Check`/`Report`/`SkillReport`/`Finding` (imported),
`filepath.Join` (imported), `fmt`/`io` (imported). **No new import lines in main.go or
main_test.go.** `git diff --quiet go.mod go.sum` MUST report "deps unchanged" (yaml.v3 stays
the sole non-stdlib dep).

---

## §8. Test patterns to MIRROR (verified in main_test.go)

- **Package:** `main` (white-box) — runInit/resolveStore/setupStore are directly callable. The
  test file already imports `configpkg` (aliased) for config.Load round-trips.
- **All-rules-miss setup (for the never-prompt test):**
  ```go
  unsetSkillsEnv(t)        // sets SKILLDOZER_SKILLS_DIR="" AND SKILLDOZER_CONFIG=<temp non-existent>
  t.Chdir(t.TempDir())     // escape the repo's walk-up rule (repo cwd HAS skills/example/SKILL.md!)
  ```
  Without t.Chdir, Find()'s walk-up rule would HIT on the repo's ./skills and the bare tag would
  go to resolve (UnknownError) instead of the unconfigured hint. t.Chdir(t.TempDir()) (Go 1.24+;
  go.mod is `go 1.25`) is the established idiom — used at main_test.go:237,377,591,849,1089,1348.
- **init SUCCESS test setup (NOT unsetSkillsEnv — we WANT config to win):**
  ```go
  parent := t.TempDir()
  store := filepath.Join(parent, "newstore")   // does NOT exist yet -> assert setupStore CREATES it
  cfg := filepath.Join(t.TempDir(), "config.yaml")
  t.Setenv("SKILLDOZER_CONFIG", cfg)           // redirect the config write to a temp file
  t.Setenv("SKILLDOZER_SKILLS_DIR", "")        // ensure config rule wins (env unset)
  t.Chdir(t.TempDir())                         // escape repo walk-up (deterministic)
  var out, errOut bytes.Buffer
  code := run([]string{"init", "--store", store}, &out, &errOut)
  ```
  Then assert: code==0; `os.Stat(store)` is a dir (created); `configpkg.Load(cfg).Store == store`
  (config written); `strings.Contains(out.String(), store)` (§6.1 stdout has the store path).
- **Unconfigured message:** `skillsdir.ErrNotFound.Error()` == `skilldozer is not configured; run
  \`skilldozer init\`` (literal backticks). Assert substrings "run" + "skilldozer init"
  (matching TestRunPathFailureErrNotFound + TestErrNotFoundMessageHasFix).
- **run() call shape:** `code := run([]string{...}, &out, &errOut)` returns int. stdout/stderr are
  `*bytes.Buffer`. Exit codes asserted directly.

---

## §9. Sibling boundaries (do NOT cross)

- **Do NOT** edit resolveStore/setupStore/exampleSkillStore/chooseStore/readPrompt/stdinIsTerminal
  (S1/S2 territory — they are the contract; S3 CONSUMES them).
- **Do NOT** touch the existing `if c.path` / `if c.check` branch BODIES (mirror their rendering in
  runInit; refactoring them into a shared helper is OPTIONAL and out of the safe one-pass scope).
- **Do NOT** edit internal/* (config/skillsdir/discover/check are consumed, not modified).
- **Do NOT** touch README (Mode B, P1.M4.T2.S1), completions (P1.M3.T2.S1), or skills/example
  (P1.M3.T1.S1).
- **Do NOT** add new deps/imports (§7). go.mod/go.sum byte-for-byte unchanged.
- S3's edit to run() is MID-FILE (~line 441); runInit's definition is APPENDED at the file tail
  (after setupStore @933). The two edits are non-overlapping (mid-file insert + tail append).

---

## §10. Acceptance gate (§13) — what init must satisfy (P1.M4.T1.S1 runs it, but S3 must make it pass)

```bash
# non-interactive init creates the store + writes the config
SKILLDOZER_CONFIG=/tmp/.../cfg.yaml ./skilldozer init --store /tmp/skilldozer-store
test -d /tmp/skilldozer-store                                          # store created (setupStore MkdirAll)
grep -q 'store: /tmp/skilldozer-store' /tmp/.../cfg.yaml               # config written (setupStore Save)
# config rule wins; env still beats config (SEPARATE --path invocations, not init's own output)
SKILLDOZER_CONFIG=/tmp/.../cfg.yaml ./skilldozer --path | grep -q /tmp/skilldozer-store
SKILLDOZER_SKILLS_DIR=/tmp/skilldozer-store SKILLDOZER_CONFIG=/tmp/.../cfg.yaml ./skilldozer --path 2>&1 | grep -q SKILLDOZER_SKILLS_DIR
# unconfigured hint (bare tag, clean env)
env -u SKILLDOZER_SKILLS_DIR HOME=/tmp/.../home XDG_CONFIG_HOME=/tmp/.../home/.config ./skilldozer x 2>err; rc=$?
[ "$rc" = 1 ] && grep -q 'run `skilldozer init`' err && echo "unconfigured-hint OK"
```
NOTE: §13 does NOT capture init's OWN stdout — it only checks `test -d` (store) + `grep` (config)
+ the subsequent --path invocations. So init's exact stdout prose (store path + check report) is
NOT acceptance-gated, but the item's required unit test DOES assert stdout contains the store path.
