# Verified Facts — P1.M1.T1.S1 (Flip the no-mode fallthrough to implicit help)

Confirmed by reading the live source (`main.go`/`main_test.go` @ HEAD `5efd3d9`) and
the plan/003 architecture docs. This is decision-17 (PRD §6.3): bare `skilldozer`
(no args / modifiers-only) ⇒ usage to **stdout**, exit **0** (implicit `--help`),
overriding the old stderr/exit-1 "parity with get-server-config.sh" behavior.

## 0. Ground-truth state (resolve the confusing git log)

`git log` shows commit `bbd4e74 Add completion subcommand and flip bare-inv to implicit help`,
but the **current** `main.go:699` is STILL `fmt.Fprint(stderr, usage())` + `return 1`.
So the no-mode flip is genuinely **NOT yet applied** — this task is real and needed.
(Completion is also absent from current main.go; that is P1.M2 and irrelevant to this
task.) The contract's line numbers match the live file exactly. Verified landmarks:
`main.go:98 func usage() string { return usageText }`; `main.go:52 const usageText`;
`main.go:428 func run(...)`; the no-mode fallthrough at `main.go:695-700`.

## 1. THE FLIP — main.go:695-700 (the only behavior change)

Current:
```go
	// No recognized mode → usage to STDERR, exit 1 (PRD §6.3: parity with
	// get-server-config.sh). Covers both truly-no-args and modifiers-only (e.g.
	// `skilldozer --no-color`): if skilldozer was asked to DO nothing, show usage. stdout stays
	// empty so $(...) never sees garbage.
	fmt.Fprint(stderr, usage())
	return 1
```
Target: swap `stderr`→`stdout` AND `1`→`0`, rewrite the comment to cite §6.3 / decision 17
(implicit --help; `skilldozer | grep …` must see the help; genuine failures stay on stderr).
This is the FINAL fallthrough in `run()` — it sits AFTER every genuine-failure path, so
flipping it cannot touch any error semantics.

## 2. Genuine-failure paths that MUST stay on stderr / non-zero (untouched)

Verified present and correct in current `run()` — do NOT touch any of these:
- unknown flag → stderr, exit 2 (`main.go:~451`, `if c.unknownFlag != ""`)
- `--store` missing value → stderr, exit 2 (`main.go:~468`, `if c.storeMissingValue`)
- exclusivity → stderr, exit 2 (`main.go:~481`, `exclusivityError`)
- unresolved/ambiguous tag → stderr, exit 1 (the `c.tags` branch, `main.go:~662`)
- unconfigured / skills dir unresolvable → stderr, exit 1

Decision 17 reclassifies ONLY the no-args/modifiers-only case (error→help). §6.4's
`$(...)` contract for genuine failures is unchanged. `test_patterns.md` lists the
regression guard: `--help`/`-h` → stdout/exit0 (wins over everything); `./skilldozer nope`
→ empty stdout/exit1; all exclusivity tests unchanged.

## 3. The three doc comments to correct (Mode A — these ARE the dev-facing exit-code docs)

**a) usageText doc comment — main.go:48-51.** Current ends: "The SAME text is printed to
stdout for --help (exit 0) and to stderr for the no-args default (exit 1) — only the
destination differs." → Now both --help AND no-args go to stdout/exit 0; genuine failures
(not this text) go to stderr.

**b) run() exit-code doc list — main.go:417-423.** Current exit-1 bullet ends with "no
recognized mode (usage to stderr)". → Remove that clause from exit-1; ADD "no-args/
modifiers-only printed usage to stdout (implicit --help, §6.3)" to the exit-0 bullet.

**c) Fallthrough comment — main.go:695-698.** (Covered by the flip in §1 — same site.)
Remove "parity with get-server-config.sh"; cite §6.3 / decision 17.

## 4. The two tests to flip (the only tests asserting the old behavior)

Collateral grep confirmed EXACTLY two tests assert no-args/modifiers-only behavior; no
others break. (`main_test.go:883 run([]string{"-a"})` is `--all`, a real mode — unaffected.)

**a) TestRunDefaultNoArgs — main_test.go:277-291.** Current asserts code==1, stdout
EMPTY, stderr contains "USAGE". → Flip to code==0, stdout contains "USAGE", stderr EMPTY
(`errOut.Len()==0`). Update the doc string (currently "to STDERR, exit 1 … parity with
get-server-config.sh") to cite §6.3 / decision 17.

**b) TestRunModifiersOnlyNoMode — main_test.go:1668-1684.** Current asserts
`run([]string{"--no-color"})` → code==1, stdout empty, stderr has "USAGE". → Flip to
code==0, stdout has "USAGE", stderr empty. Update the doc string. Mirrors (a).

Marker note: `usageText` contains the literal `USAGE:` header (`main.go:56`), so
`strings.Contains(out.String(), "USAGE")` matches (USAGE is a substring of USAGE:). The
§13 gate greps case-insensitively for `USAGE`. Use `"USAGE"` for consistency with the
original test (both forms work).

## 5. The §13 "Grepability contract" (the acceptance gate this satisfies)

From PRD §13 (prd_snapshot.md:370-373):
```bash
out=$(./skilldozer 2>/dev/null); rc=$?
[ "$rc" = "0" ] && printf '%s' "$out" | grep -qi 'USAGE' && echo "no-args-help-on-stdout OK"
test -z "$(./skilldozer 2>&1 >/dev/null)"   # no-args writes NOTHING to stderr
```
The flip makes both pass: `run(nil)` writes usage to stdout (rc 0, grep finds USAGE) and
nothing to stderr. `run(["--no-color"])` behaves identically.

## 6. Deps / build invariant

`go.mod`: module `github.com/dabstractor/skilldozer`, `go 1.25`, sole dep
`gopkg.in/yaml.v3 v3.0.1`. This task changes no imports (stdout/stderr/usage() all already
in scope), adds no deps. `go.mod`/`go.sum` stay byte-for-byte identical.

## 7. Scope boundary (what this subtask is NOT)

- NOT the `completion` subcommand (P1.M2 — separate task; absent from current code).
- NOT the README sync (Mode B is P1.M3.T1; §15 README outline does not mention no-args
  behavior, so no per-subtask README edit is needed — contract OUTPUT §5).
- NOT touching any genuine-failure path (§2).
