# P1.M1.T2.S1 — Verified Facts & Site Inventory (subcommands → flags, user-facing surfaces)

Every line number below is the **CURRENT** main.go (post-S1 HEAD `594be07`, 1275 lines) or
internal/skillsdir/skillsdir.go, verified by grep + read on 2026-07-15. The contract
(§1 RESEARCH NOTE) and code_change_map.md (Change Groups 4-6) cite **PRE-S1** line numbers
(HEAD `f30d5c5`, 1288 lines) — they are ALL STALE by ~13 lines. **LOCATE BY CONTENT.**

## 0. The two HEADs that explain the line-number drift

- `f30d5c5` (1288 lines): the snapshot the contract/change-map was written against (BEFORE S1).
- `594be07` (1275 lines): S1 committed ("Replace bare subcommands with flags in parseArgs").
  S1 NET-DELETED 13 lines (the bare `case "check"/"init"/"completion":` blocks were larger
  than the flag cases that replaced them) → every line after parseArgs shifted UP ~13.
- The sibling S2 PRP (P1.M1.T1S2) already documented this drift for exclusivityError
  (GOTCHA #1: "contract 782-835 / change-map 804-829 are pre-S1; function is now 769-818").
  T2.S1 inherits the SAME lesson for ALL its sites.

## 1. The S1/S2/T2.S1 ownership boundary (no conflicts — verified)

| Region | Current lines | Owner | Status |
|---|---|---|---|
| parseArgs bare→flag cases + config struct doc | 179-369, 148-168 | **S1** | DONE (committed 594be07) |
| exclusivityError body + doc + inline comments | 750-818 | **S2** (parallel, in-progress) | DO NOT TOUCH |
| usageText constant | 71-117 | **T2.S1 (this)** | — |
| skillsdir.ErrNotFound | skillsdir.go:275 | **T2.S1** | — |
| runCheck doc comment | 618-630 | **T2.S1** | — |
| chooseStore / expandHome / resolveStore / exampleSkillTemplate / setupStore docs + error prefixes | 895-1097 | **T2.S1** | — |
| completionScript / runInit / detectShell / runCompletion docs | 1095-1247 | **T2.S1** | — |
| main_test.go / skillsdir_test.go | — | **T3** (P1.M1.T3) | DO NOT TOUCH |

S2's GOTCHA #11 EXPLICITLY defers to T2.S1: "usageText (71-117) / error-prefix strings /
runInit 'skilldozer init:' prefixes / runCheck doc (631) / completion-function docs
(1112-1271) → P1.M1.T2.S1." → No overlap with S2 (exclusivityError is the ONLY thing S2 edits).

## 2. usageText constant — EXACT current text (main.go:71-117), a raw string literal

gofmt does NOT reformat inside a raw string literal (backtick), so manual alignment of the
USAGE/EXAMPLES/OPTIONS columns is required after the flag-name swaps (the description column
shifts because `--check`/`--init [<dir>]`/`--completions [--shell <name>]` are wider than
`check`/`init [<dir>]`/`completion [--shell <name>]`).

USAGE block (lines 80-82) — 3 bare lines:
```
  skilldozer check
  skilldozer init [<dir>]
  skilldozer completion [--shell <name>]
```
→ `--check` / `--init [<dir>]` / `--completions [--shell <name>]`

EXAMPLES block (lines 95-97) — 3 lines:
```
  skilldozer check                   # validate every skill on disk
  skilldozer init --store <dir>     # non-interactive first-run setup
  eval "$(skilldozer completion)"     # load completions into your shell
```
→ `skilldozer --check …` / `skilldozer --init --store <dir> …` / `eval "$(skilldozer --completions)" …`

OPTIONS block (lines 107-111) — 3 bare lines (the `--store <dir>` line at ~109 is already a flag, UNCHANGED):
```
  check              Validate every skill on disk (report OK / WARN / ERROR)
  init [<dir>]      First-run setup: pick/create the skills store and write the config
  --store <dir>     Non-interactive store path for init
  completion [--shell <name>]   Emit the shell completion script for eval (§14.6)
```
→ `--check` / `--init [<dir>]` / `--completions [--shell <name>]` (the `--store <dir>` line stays).

Long-form-only note (contract LOGIC a): add a line before OPTIONS (or after USAGE). Text:
`Help and --completions advertise long forms only; short aliases (-a, -l, -s, -f, -p, -h, -v)
remain valid for typing but are not advertised (§6.1).`

## 3. skillsdir.ErrNotFound — internal/skillsdir/skillsdir.go:275 (the ONLY internal/ change)

```go
275: var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer init`")
```
→ `run \`skilldozer --init\`` (PRD §8.2/§8.3/§6.4 all mandate `--init`).

## 4. Error-prefix strings — 8 fmt.Errorf sites (all `"skilldozer init: "` → `"skilldozer --init: "`)

| Current line | Function | Exact current string |
|---|---|---|
| 988 | resolveStore | `fmt.Errorf("skilldozer init: resolve cwd: %w", err)` |
| 992 | resolveStore | `fmt.Errorf("skilldozer init: resolve default store: %w", err)` |
| 1014 | resolveStore | `fmt.Errorf("skilldozer init: absolutize store: %w", err)` |
| 1077 | setupStore | `fmt.Errorf("skilldozer init: create store dir %q: %w", store, err)` |
| 1082 | setupStore | `fmt.Errorf("skilldozer init: read store dir %q: %w", store, err)` |
| 1087 | setupStore | `fmt.Errorf("skilldozer init: create example dir: %w", err)` |
| 1090 | setupStore | `fmt.Errorf("skilldozer init: seed example SKILL.md: %w", err)` |
| 1097 | setupStore | `fmt.Errorf("skilldozer init: write config %q: %w", configPath, err)` |

(Contract cited 1001/1005/1027/1090/1095/1100/1103/1110 — all shifted up ~13 by S1. The 8
COUNT and the resolveStore×3 / setupStore×5 split match.)

## 5. Error-prefix-quote comment sites — 3 (contract LOGIC d)

These quote the `"skilldozer init: …"` wrap format in backticks:
- 1073 (setupStore doc, last line): `Errors are wrapped with a "skilldozer init: <step>: %w" prefix.`
- 1137 (runInit, resolveStore err branch): `// one-line (resolveStore wraps with "skilldozer init: …")`
- 1149 (runInit, setupStore err branch): `// setupStore wraps with "skilldozer init: …"`
All three: `"skilldozer init: …"` → `"skilldozer --init: …"`.

## 6. Function doc comments (contract LOGIC e/f + change-map 6a/6b/6c)

### 6a. completionScript (main.go:1095-1101) — VERIFIED NO-OP
The contract/change-map (6a) says "update `skilldozer completion` → `skilldozer --completions`"
in completionScript's doc. VERIFIED by grep: completionScript's doc (1095-1101) contains
ZERO `skilldozer completion` references (it talks about embedded scripts / //go:embed, not
the command). → **completionScript doc needs NO change.** Do not search for a phantom edit.
Only `detectShell` and `runCompletion` have the reference (below).

### 6b. detectShell (main.go:1221)
`// detectShell resolves the target shell for \`skilldozer completion\` (PRD §14.6` → `--completions`

### 6c. runCompletion (main.go:1244 + 1247)
- 1244: `// runCompletion is the \`skilldozer completion\` handler (PRD §14.6 / §6.4).` → `--completions`
- 1247: `// the matching embedded script to stdout for \`eval "$(skilldozer completion)"\`` → `--completions`

### 6f. runCheck (main.go:618-622)
- 618: `// \`skilldozer check\` subcommand (PRD §9). Validates every skill in the store and` → `\`skilldozer --check\` flag`
- 622: `// code, so \`if skilldozer check; then …\` works as a gate).` → `\`if skilldozer --check; then …\``

## 7. Mode-A consistency sweep (OUTPUT §4 "all doc comments use flag language")

OUTPUT §4 mandates the WHOLE main.go use flag language. These full-command-name prose refs
in doc comments are stale and belong to functions T2.S1 is already editing (no sibling conflict):

- 895 (chooseStore doc): `// chooseStore resolves the store directory for \`skilldozer init\` (PRD §8.2) via a` → `--init`
- 998 (resolveStore comment): `so a caller doing store="$(skilldozer init)" must not capture it.` → `skilldozer --init`
- 1021 (exampleSkillTemplate doc): `// in"; code_prd_delta.md G11). skilldozer init writes this verbatim into an EMPTY store's` → `skilldozer --init`
- 1051 (setupStore doc): `// the create+seed+writeconfig half of \`skilldozer init\` (PRD §8.2 steps 2-4); the` → `--init`
- 1120 (runInit doc): `// runInit is the \`skilldozer init\` orchestrator (PRD §8.2). run()'s dispatch calls it` → `--init`
- 1176 (runInit check-report comment): `// (6) \`skilldozer check\` report on the effective store (PRD §8.2 step 5). init renders` → `--check`
- 1178 (runInit check-report comment): `//     \`check\` subcommand keeps its report on stdout (its report IS its stdout product),` → `\`--check\` flag`

EXCLUDED (positional-form shorthand, not full command names — lowest priority, leave to avoid
churn): `init ~/x` / `init ~/myskills` shorthand in expandHome (978-985) and resolveStore
(986-989) docs; the `the \`check\` report to stderr` shorthand at runInit 1125. These refer to
the flag's positional/argument form in passing, not the command surface. (If the implementer is
already in those doc blocks and wants full consistency, `--init ~/x` / `--init <dir>` are the
correct replacements — but they are NOT contract-pinned.)

## 8. The gate: go build + go vet ONLY (go test is EXPECTED RED — T3's scope)

Inherited from sibling S2's stance (its GOTCHA #8): S1's bare→flag parseArgs conversion made
the run-level tests RED (they pass bare words through parseArgs). T2.S1's OWN string changes
make MORE tests red. The contract OUTPUT §4 says only "go build ./... must succeed" — NOT
"go test passes". P1.M1.T3 (S1+S2) is the dedicated test-flip milestone. **Do NOT touch
main_test.go or skillsdir_test.go.**

### Tests T2.S1's changes will turn red (so T3 / reviewers know the impact):
- `skillsdir_test.go:526-530` — asserts ErrNotFound message contains `"run", "skilldozer init"`
  (exact-message test; breaks on `--init`).
- `main_test.go:247, 490, 704, 962, 1202, 1592` — `Contains(errOut, "skilldozer init")` (the
  error-prefix strings now say `skilldozer --init`).
- `main_test.go:2036` — usageText assertion wants `"skilldozer init"` (now `--init`).
- `main_test.go:2123` — usageText assertion wants `"skilldozer completion"` (now `--completions`).
- `main_test.go:2880` — wants `"run", "skilldozer init"` (the ErrNotFound/help path).

All of these are EXACTLY what P1.M1.T3.S1 (parseArgs/exclusivity-level) and S2
(run-level dispatch + completion + help-text) flip. T2.S1 just feeds them the new strings.

## 9. Baseline (verified green for build+vet, the gate)

- `go build ./...` → BUILD_OK
- `go vet ./...` → VET_OK
- go.mod/go.sum UNCHANGED (pure string-literal + doc-comment edits; no imports).

## 10. Discrepancies between the contract and the current code (resolved)

1. **All line numbers stale (pre-S1).** Resolved: §0 + every site listed by CURRENT line.
2. **completionScript doc (contract 6a) is a no-op.** Resolved: §6a — verified no
   `skilldozer completion` reference exists there; only detectShell + runCompletion change.
3. **Contract LOGIC(d) "3 comment sites"** = the error-PREFIX-quote comments (1073/1137/1149),
   NOT the broader `skilldozer init` prose (which OUTPUT §4 sweeps — §7).
4. **Contract LOGIC(a) "Add a note after USAGE or before OPTIONS"** — placement is a judgment
   call; the note qualifies how OPTIONS are advertised, so BEFORE OPTIONS is the most natural
   spot. (Either is acceptable per the contract's "after USAGE or before OPTIONS".)
