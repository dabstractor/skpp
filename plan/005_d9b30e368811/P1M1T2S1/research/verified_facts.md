# P1.M1.T2.S1 — Verified Facts & Test Inventory (§14.7 emitted-byte contract locks)

Source of truth for the PRP. Every line number below is the CURRENT `main_test.go` /
`main.go` / `completions/skilldozer.bash`, verified on 2026-07-15 against HEAD `5cf81d4`.
The contract's cited test line numbers are ACCURATE (no drift).

## 0. The option strings are ALREADY PRESENT (T1.S1 done + T1.S2 in current code)

This subtask is **test-only**. The code it locks is already shipped:
- **bash (T1.S1, Complete)** — `completions/skilldozer.bash` (== the embedded bytes; bash is
  emitted VERBATIM, so `completionScript("bash")` == on-disk file):
  - line 83: `[[ $- == *i* ]] && bind 'set show-all-if-ambiguous on'` (ACTIVE)
  - line 85: `#   bind 'set show-all-if-ambiguous off'` (OPT-OUT, commented)
  - lines 73/76: disclosure comments (uppercase OFF/ON — do NOT match lowercase tokens below)
- **zsh (T1.S2)** — `zshEvalRegistration` raw-string const in main.go (the DERIVED eval wrapper;
  appended to the stripped autoload body by `zshEvalScript`, emitted for zsh by `runCompletion`):
  - main.go:1292: `setopt NO_LIST_AMBIGUOUS` (ACTIVE, unguarded — zsh setopt is silent non-interactively)
  - main.go:1291: `#   setopt LIST_AMBIGUOUS` (OPT-OUT, commented)
  - The autoload file `completions/_skilldozer` is UNCHANGED (byte-identity holds).

The sibling PRP (P1.M1.T1S2) confirms these exact tokens; the current code matches its contract.

## 1. Token PRECISION (verified non-overlapping — the assertions are exact)

- zsh: `setopt NO_LIST_AMBIGUOUS` does **NOT** contain `setopt LIST_AMBIGUOUS` (verified:
  `echo "setopt NO_LIST_AMBIGUOUS" | grep -q "setopt LIST_AMBIGUOUS"` → no match). So asserting
  BOTH substrings is precise: `setopt NO_LIST_AMBIGUOUS` matches the active line (1292) only;
  `setopt LIST_AMBIGUOUS` matches the opt-out comment (1291) only.
- bash: `strings.Contains` is CASE-SENSITIVE. `show-all-if-ambiguous on` (lowercase "on") appears
  ONLY in line 83 (the active bind); `show-all-if-ambiguous off` (lowercase "off") ONLY in line 85
  (the opt-out). The disclosure comments (73 "OFF", 76 "ON") use UPPERCASE and do NOT match. So
  both assertions are precise.
- `*i*` (the bash interactivity guard) appears in line 83 (`[[ $- == *i* ]]`).

## 2. The test harness (confirmed — package main internal tests)

- `run(args []string, stdout, stderr io.Writer) int` — main.go:524 (the exported test seam). Tests
  call `run([]string{...}, &out, &errOut)` where out/errOut are `bytes.Buffer`; `code := run(...)`.
- `completionScript(shell string) (string, bool)` — main.go:1215. Returns the embedded bytes
  verbatim (bash/zsh/fish). For zsh this is the AUTOLOAD file, NOT the derived wrapper.
- `zshEvalScript(raw string) string` — main.go:1245. Strips the `_skilldozer "$@"` self-call from
  the raw autoload body, then APPENDS the `zshEvalRegistration` const (main.go:1262+). This is the
  DERIVED eval wrapper that carries the §14.7 setopt.
- Tests are `package main` (internal) — they call run/completionScript/zshEvalScript directly.
- Assertion idiom: `strings.Contains(out.String(), want)` + `t.Errorf("...:\n%s", out.String())`.

## 3. The 4 cited tests — current bodies (main_test.go) + the extension plan

### TestEmbeddedCompletionsMatchOnDisk (main_test.go:3139) — INVARIANT, do NOT weaken
Compares `completionScript(shell)` to the on-disk file for bash/zsh/fish. For zsh it compares the
embed var (autoload) to `completions/_skilldozer` — UNCHANGED by T1.S1/T1.S2 (the const is appended
AFTER completionScript returns, in zshEvalScript). For bash the embed == on-disk file, both move
together. T2.S1 is test-only → this test is UNTOUCHED and stays GREEN by construction.

### TestRunCompletionBashScript (main_test.go:3163) — the §13 marker test
```go
func TestRunCompletionBashScript(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--completions", "--shell", "bash"}, &out, &errOut)
	if code != 0 { t.Errorf("run(completion --shell bash): code=%d; want 0", code) }
	if !strings.Contains(out.String(), "_skilldozer_completion") {
		t.Errorf("stdout missing _skilldozer_completion (§13):\n%s", out.String())
	}
	if errOut.Len() != 0 { t.Errorf("stderr=%q; want empty on success", errOut.String()) }
}
```
→ T2.S1 ADDS a DEDICATED `TestRunCompletionBashListsAmbiguous` (primary) rather than mixing §14.7
into this §13 test. (Extending this test is an acceptable alternative per Touch point 5, but a
dedicated test isolates the §14.7 contract.) The dedicated test re-runs
`run(["--completions","--shell","bash"], ...)` and asserts the two option tokens + the `*i*` guard.

### TestZshEvalScriptRegistersCompdef (main_test.go:3288) — EXTEND the want-loop
```go
got := zshEvalScript(raw)   // raw = completionScript("zsh")
for _, want := range []string{
	"autoload -Uz compinit",
	"(( $+functions[compdef] )) || compinit",
	"compdef _skilldozer skilldozer",
} { if !strings.Contains(got, want) { t.Errorf(...) } }
```
→ ADD two entries to the want-slice:
```go
	"setopt NO_LIST_AMBIGUOUS", // §14.7 active
	"setopt LIST_AMBIGUOUS",    // §14.7 opt-out (commented; substring still present)
```
(Precise: see §1 — non-overlapping.) The existing compdef/header/_arguments assertions stay.

### TestRunCompletionZshIsEvalSafe (main_test.go:3316) — EXTEND with one assertion
```go
t.Setenv("SKILLDOZER_SHELL", "zsh")
... code := run([]string{"--completions"}, &out, &errOut) ...
script := out.String()
// existing: no self-call; has compdef; != on-disk autoload
```
→ ADD (after the compdef check):
```go
if !strings.Contains(script, "NO_LIST_AMBIGUOUS") {
	t.Errorf("zsh eval output missing the §14.7 NO_LIST_AMBIGUOUS listing option:\n%s", script)
}
```
(Per contract LOGIC c, the e2e asserts the shorter token `NO_LIST_AMBIGUOUS` — present in the
active line `setopt NO_LIST_AMBIGUOUS`.)

## 4. code_change_map.md Touch point 5 — the authoritative test spec (mirrors the contract)

- bash: add/extend → assert `show-all-if-ambiguous on` AND `show-all-if-ambiguous off`; optionally `*i*`.
- zsh unit: extend TestZshEvalScriptRegistersCompdef → `setopt NO_LIST_AMBIGUOUS` + `setopt LIST_AMBIGUOUS`.
- zsh e2e: extend TestRunCompletionZshIsEvalSafe → `NO_LIST_AMBIGUOUS`.
- Invariant: TestEmbeddedCompletionsMatchOnDisk stays GREEN.

## 5. Baseline (verified green now; T2.S1 only ADDS tests)

- `go test ./...` → green (cached). The option strings are present, so the new/extended tests WILL
  pass against the current code. T2.S1 adds no production code, so nothing can regress.
- go.mod/go.sum UNCHANGED (test-only; no new imports — bytes/strings/os already imported).

## 6. Scope boundary (test-only — verified)

T2.S1 edits ONLY `main_test.go`. It does NOT touch:
- `main.go` (the zshEvalRegistration const — T1.S2's territory, already present)
- `completions/skilldozer.bash` (T1.S1's territory, already present) / `_skilldozer` / `.fish`
- `README.md` (the §14.7 disclosure is P1.M1.T3.S1, Mode B)
- `PRD.md` / `tasks.json` / `prd_snapshot.md` / `.gitignore`

The parallel sibling T1.S2 (zsh const) is disjoint from main_test.go — no conflict; T2.S1's tests
lock exactly the strings T1.S1 (done) + T1.S2 (current code) emit.
